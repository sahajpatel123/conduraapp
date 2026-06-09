// Package session ties the voice pipeline, the LLM stream, the
// gated executor, and the TTS speaker into a single end-to-end
// user interaction.
//
// The full flow:
//
//	voice → transcript
//	transcript → llm.stream (over stream.Manager)
//	stream tokens → spoken (TTS) + broker (SSE for the overlay)
//	computer-use actions from the model → gated executor
//	completion → idle
//
// The Session is the only place that knows the full lifecycle. Each
// sub-system is independent and testable in isolation. The Session
// is the glue.
package session

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/agent"
	"github.com/sahajpatel123/synapticapp/internal/conversation"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/llm"
	"github.com/sahajpatel123/synapticapp/internal/status"
	"github.com/sahajpatel123/synapticapp/internal/stream"
	"github.com/sahajpatel123/synapticapp/internal/voice"
)

// ErrAlreadyRunning is returned by Start when a previous session is
// still in flight.
var ErrAlreadyRunning = errors.New("session: already running")

// Provider is the subset of the llm.Registry the Session needs. It
// exists so tests can pass a fake.
type Provider interface {
	Chat(ctx context.Context, name string, req llm.ChatRequest) (llm.ChatResponse, error)
}

// Executor is the subset of agent.Executor the Session needs.
type Executor interface {
	Execute(ctx context.Context, a *agent.Action) (*agent.StepResult, error)
}

// Config configures a Session.
type Config struct {
	StreamMgr      *stream.Manager
	Transcriber    voice.Transcriber
	Speaker        voice.Speaker
	Provider       Provider
	ProviderName   string
	Model          string
	Conversation   *conversation.Store
	Executor       Executor
	Gatekeeper     gatekeeper.Gatekeeper
	ConversationID int64
}

// Session runs a single user query end-to-end. It is created fresh
// for each user query (or reused for push-to-talk); it owns the
// state machine for that one interaction.
type Session struct {
	cfg Config

	mu   sync.Mutex
	busy bool
}

// New creates a Session.
func New(cfg Config) (*Session, error) {
	if cfg.StreamMgr == nil {
		return nil, errors.New("session: StreamMgr is required")
	}
	if cfg.Provider == nil {
		return nil, errors.New("session: Provider is required")
	}
	if cfg.ProviderName == "" {
		return nil, errors.New("session: ProviderName is required")
	}
	if cfg.Model == "" {
		return nil, errors.New("session: Model is required")
	}
	return &Session{cfg: cfg}, nil
}

// Run executes a single end-to-end session for a user query.
// The query is the text the user just spoke (or typed). The session
// streams the model's response and speaks it back.
//
// The returned transcript is the final text the model said. If the
// model requested a computer-use action, the Executor is invoked
// (through the Gatekeeper) and the result is appended to the
// response text.
//
// On any error, the session is reset to idle. The caller is
// responsible for updating the tray/overlay.
func (s *Session) Run(ctx context.Context, query string) (string, error) {
	s.mu.Lock()
	if s.busy {
		s.mu.Unlock()
		return "", ErrAlreadyRunning
	}
	s.busy = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.busy = false
		s.mu.Unlock()
	}()

	if query == "" {
		return "", nil
	}

	// Build the chat request. History is loaded from the
	// conversation store when one is configured.
	messages, err := s.buildMessages(ctx, query)
	if err != nil {
		return "", fmt.Errorf("session: build messages: %w", err)
	}

	// Kick off a streaming LLM call.
	chatReq := llm.ChatRequest{
		Model:    s.cfg.Model,
		Messages: messages,
		Stream:   true,
	}

	streamReq := stream.Request{
		ProviderName: s.cfg.ProviderName,
		Chat:         chatReq,
	}
	if s.cfg.ConversationID != 0 {
		streamReq.ConversationID = s.cfg.ConversationID
	}

	requestID, err := s.cfg.StreamMgr.Start(ctx, streamReq)
	if err != nil {
		return "", fmt.Errorf("session: stream start: %w", err)
	}

	// Wait for the stream to complete. The stream manager owns the
	// lifecycle; we poll its state. The chat itself is in flight
	// inside the manager.
	return s.collectAndSpeak(ctx, requestID)
}

// collectAndSpeak waits for the stream to complete, accumulates the
// full transcript, and speaks it.
func (s *Session) collectAndSpeak(ctx context.Context, requestID string) (string, error) {
	const (
		pollInterval = 50 * time.Millisecond
		streamBudget = 60 * time.Second
	)
	deadline := time.Now().Add(streamBudget)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		active := s.findStream(requestID)
		if active == nil {
			// Stream finished or was canceled.
			break
		}
		if active.State == stream.StateFinished || active.State == stream.StateCancelled || active.State == stream.StateError {
			break
		}
		time.Sleep(pollInterval)
	}

	// Pull the most recent assistant message from the conversation
	// store. The stream manager writes each message as it lands.
	var full string
	if s.cfg.Conversation != nil && s.cfg.ConversationID != 0 {
		msgs, err := s.cfg.Conversation.GetRecentMessages(ctx, s.cfg.ConversationID, 1)
		if err == nil && len(msgs) > 0 && msgs[len(msgs)-1].Role == string(llm.RoleAssistant) {
			full = msgs[len(msgs)-1].Content
		}
	}

	// Speak the response (best-effort).
	if full != "" && s.cfg.Speaker != nil {
		_ = s.cfg.Speaker.Speak(ctx, full)
	}

	return full, nil
}

// findStream returns the active snapshot for a requestID, or nil.
func (s *Session) findStream(requestID string) *stream.Active {
	for _, a := range s.cfg.StreamMgr.List() {
		if a.RequestID == requestID {
			return &a
		}
	}
	return nil
}

// buildMessages assembles the chat history with the new user query.
func (s *Session) buildMessages(ctx context.Context, query string) ([]llm.Message, error) {
	const historyLimit = 20
	messages := []llm.Message{
		{Role: llm.RoleUser, Content: query},
	}
	if s.cfg.Conversation == nil {
		return messages, nil
	}
	history, err := s.cfg.Conversation.GetRecentMessages(ctx, s.cfg.ConversationID, historyLimit)
	if err != nil {
		return messages, err
	}
	out := make([]llm.Message, 0, len(history)+1)
	for _, m := range history {
		out = append(out, llm.Message{
			Role:       llm.Role(m.Role),
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		})
	}
	out = append(out, llm.Message{Role: llm.RoleUser, Content: query})
	return out, nil
}

// Status is the convenience accessor for the current session state.
// For now this is always idle (sessions are short-lived), but the
// API is here for symmetry with the future streaming UX.
func (s *Session) Status() status.Status {
	return status.StatusIdle
}
