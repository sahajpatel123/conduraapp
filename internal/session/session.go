// Package session ties the voice pipeline, the LLM stream, the
// gated executor, and the TTS speaker into a single end-to-end
// user interaction.
//
// The full flow:
//
//	voice → transcript
//	transcript → llm.stream (over stream.Manager)
//	stream tokens → spoken (TTS) + broker (SSE for the overlay)
//	computer-use actions from the model → gated executor (6B)
//	completion → idle
//
// The Session is the only place that knows the full lifecycle. Each
// sub-system is independent and testable in isolation. The Session
// is the glue.
package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/conversation"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/llm"
	"github.com/sahajpatel123/synapticapp/internal/sse"
	"github.com/sahajpatel123/synapticapp/internal/status"
	"github.com/sahajpatel123/synapticapp/internal/stream"
	"github.com/sahajpatel123/synapticapp/internal/voice"
)

// ErrAlreadyRunning is returned by Run when a previous session is
// still in flight.
var ErrAlreadyRunning = errors.New("session: already running")

// Provider is the subset of the llm.Registry the Session needs. It
// exists so tests can pass a fake.
type Provider interface {
	Chat(ctx context.Context, name string, req llm.ChatRequest) (llm.ChatResponse, error)
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
	Broker         *sse.Broker
	ConversationID int64
	Gatekeeper     gatekeeper.Gatekeeper
	Audit          *audit.Log
	Memory         MemoryStore
	Predictor      PredictorStore
}

// MemoryStore is the subset of the memory package used by sessions.
type MemoryStore interface {
	Recall(ctx context.Context, query string, limit int) ([]string, error)
}

// PredictorStore is the subset of the adaptive package used by sessions.
type PredictorStore interface {
	Predict(ctx context.Context, query string) (string, error)
}

// Session runs a single user query end-to-end. It is created fresh
// for each user query (or reused for push-to-talk); it owns the
// state machine for that one interaction.
type Session struct {
	cfg Config

	mu     sync.Mutex
	busy   bool
	cur    atomic.Int32 // status.Status
	cancel context.CancelFunc

	// OnStatus fires on every state transition. The daemon wires
	// it to the SSE broker so the GUI's tray can react.
	OnStatus func(status.Status)
}

// OnStatus is set by the daemon to receive state transitions. The
// voice pipeline, the tray host, and the overlay subscribe via
// this callback (the daemon fans it out on the SSE broker so the
// GUI process can react).
//
// The OnStatus callback is set on each Session at construction
// time by the daemon. It fires on every state transition.

// Factory builds sessions from a long-lived set of dependencies.
// Construct once at daemon startup; call New for each user query.
type Factory struct {
	streamMgr    *stream.Manager
	provider     Provider
	providerName string
	model        string
	convStore    *conversation.Store
	broker       *sse.Broker
	speaker      voice.Speaker
	onStatus     func(status.Status)
	gate         gatekeeper.Gatekeeper
	audit        *audit.Log
	memory       MemoryStore
	predictor    PredictorStore

	mu         sync.Mutex
	activeSess *Session
}

// NewFactory creates a session factory. An empty providerName
// is allowed: the factory is constructed but session.Run will
// fail with an error until a provider is added (this is the
// state at first launch when the user has not yet configured
// an LLM).
func NewFactory(
	streamMgr *stream.Manager,
	provider Provider,
	providerName, model string,
	convStore *conversation.Store,
	broker *sse.Broker,
) (*Factory, error) {
	if streamMgr == nil {
		return nil, errors.New("session: streamMgr is required")
	}
	if provider == nil {
		return nil, errors.New("session: provider is required")
	}
	if broker == nil {
		return nil, errors.New("session: broker is required")
	}
	return &Factory{
		streamMgr:    streamMgr,
		provider:     provider,
		providerName: providerName,
		model:        model,
		convStore:    convStore,
		broker:       broker,
	}, nil
}

// SetSpeaker injects a TTS speaker. May be called after NewFactory
// when the speaker is constructed later (e.g. after the voice
// pipeline is initialized).
func (f *Factory) SetSpeaker(s voice.Speaker) {
	f.speaker = s
}

// SetOnStatus injects a status callback. Every session built
// from this factory will have OnStatus set to the given function.
// The daemon uses this to fan session status out to the SSE
// broker.
func (f *Factory) SetOnStatus(fn func(status.Status)) {
	f.onStatus = fn
}

// SetGatekeeper injects the safety gatekeeper and audit log.
func (f *Factory) SetGatekeeper(gate gatekeeper.Gatekeeper, auditLog *audit.Log) {
	f.gate = gate
	f.audit = auditLog
}

// SetMemory injects the memory store for context recall.
func (f *Factory) SetMemory(m MemoryStore) {
	f.memory = m
}

// SetPredictor injects the adaptive engine predictor.
func (f *Factory) SetPredictor(p PredictorStore) {
	f.predictor = p
}

// New builds a Session for a specific conversation. The
// session's lifetime is the lifetime of one Run call.
func (f *Factory) New(conversationID int64) *Session {
	f.mu.Lock()
	defer f.mu.Unlock()

	s := &Session{
		cfg: Config{
			StreamMgr:      f.streamMgr,
			Provider:       f.provider,
			ProviderName:   f.providerName,
			Model:          f.model,
			Conversation:   f.convStore,
			Broker:         f.broker,
			Speaker:        f.speaker,
			ConversationID: conversationID,
			Gatekeeper:     f.gate,
			Audit:          f.audit,
			Memory:         f.memory,
			Predictor:      f.predictor,
		},
		OnStatus: f.onStatus,
	}
	f.activeSess = s
	return s
}

// Status returns the status of the active session, or StatusIdle if no
// session is currently running.
func (f *Factory) Status() status.Status {
	f.mu.Lock()
	s := f.activeSess
	f.mu.Unlock()
	if s == nil {
		return status.StatusIdle
	}
	return s.Status()
}

// New creates a Session. ProviderName and Model may be empty;
// Run will return an error in that case (the user has not yet
// configured an LLM).
func New(cfg Config) (*Session, error) {
	if cfg.StreamMgr == nil {
		return nil, errors.New("session: StreamMgr is required")
	}
	if cfg.Provider == nil {
		return nil, errors.New("session: Provider is required")
	}
	if cfg.Broker == nil {
		return nil, errors.New("session: Broker is required")
	}
	return &Session{cfg: cfg}, nil
}

// Run executes a single end-to-end session for a user query.
// The query is the text the user just spoke (or typed). The session
// streams the model's response, accumulates it from SSE delta
// events, and speaks it back.
//
// The returned string is the full text the model said. It is
// accumulated client-side from the SSE broker so the session does
// not depend on the stream manager persisting the reply.
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

	// Detach the session's lifetime from the caller's context.
	// The stream manager owns the stream lifetime; the session
	// only needs a context to bound how long it waits for the
	// stream to finish. The cancel here is used by Cancel() to
	// abort the wait if the user requests it.
	runCtx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.mu.Unlock()
	defer func() {
		cancel()
		s.mu.Lock()
		s.busy = false
		s.cancel = nil
		s.mu.Unlock()
	}()

	if query == "" {
		s.setStatus(status.StatusIdle)
		return "", nil
	}

	// Evaluate the utterance via the gatekeeper (Kind: "chat").
	if s.cfg.Gatekeeper != nil {
		action := blastradius.Action{Kind: "chat"}
		decision, reason := s.cfg.Gatekeeper.Evaluate(runCtx, action)

		// Audit the gatekeeper decision.
		if s.cfg.Audit != nil {
			class := blastradius.Classify(action)
			level := "info"
			result := "allow"
			if decision != gatekeeper.Allow {
				level = "warn"
				result = "deny"
			}
			_ = s.cfg.Audit.Append(runCtx, audit.Event{
				Actor:   "gatekeeper",
				Action:  "utterance",
				App:     "session",
				Level:   level,
				Result:  result,
				Message: fmt.Sprintf("%s [%s]: text=%q reason=%q", class, decision, query, reason),
			})
		}

		if decision != gatekeeper.Allow {
			s.setStatus(status.StatusError)
			// Publish a stream.error event to the broker so the frontend knows it was blocked
			if s.cfg.Broker != nil {
				s.cfg.Broker.PublishJSON("stream", map[string]any{
					"conversation_id": s.cfg.ConversationID,
					"err":             "gatekeeper blocked utterance: " + reason,
					"done":            true,
				})
			}
			return "", fmt.Errorf("gatekeeper denied utterance: %s", reason)
		}
	}

	// Persist the user message first so the next turn's history
	// includes this query, even if the stream fails.
	if err := s.persistUserMessage(runCtx, query); err != nil {
		// Persistence failure is not fatal — the stream can still
		// proceed in-memory. Log and continue.
		slog.Warn("session: persist user message failed", "err", err)
	}

	s.setStatus(status.StatusThinking)

	// Build the chat request. History is loaded from the
	// conversation store so the model has full context.
	messages, err := s.buildMessages(runCtx, query)
	if err != nil {
		s.setStatus(status.StatusError)
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

	// Subscribe to the broker BEFORE starting the stream, so we don't miss
	// any deltas. The stream manager publishes each token as a
	// stream.delta event; we accumulate those and stop when we
	// see stream.finished / stream.error / stream.cancelled.
	sub := s.cfg.Broker.Subscribe()
	defer s.cfg.Broker.Unsubscribe(sub)

	requestID, err := s.cfg.StreamMgr.Start(runCtx, streamReq)
	if err != nil {
		s.setStatus(status.StatusError)
		return "", fmt.Errorf("session: stream start: %w", err)
	}

	full, err := s.collectAndSpeak(runCtx, requestID, sub)
	if err != nil {
		s.setStatus(status.StatusError)
		return full, err
	}

	// Persist the assistant response so the next turn's history
	// includes this reply. Without this, buildMessages would only
	// see the user message and the LLM would lose multi-turn context.
	if full != "" {
		if persistErr := s.persistAssistantMessage(runCtx, full); persistErr != nil {
			slog.Warn("session: persist assistant message failed", "err", persistErr)
		}
	}

	s.setStatus(status.StatusIdle)
	return full, nil
}

// collectAndSpeak accumulates delta events for the given requestID from the
// active subscription, and speaks the result.
//
// Returns the accumulated text and any error from the stream
// (cancel, broker subscription failure, or context cancel).
func (s *Session) collectAndSpeak(ctx context.Context, requestID string, sub *sse.Subscription) (string, error) {
	const streamBudget = 60 * time.Second
	timer := time.NewTimer(streamBudget)
	defer timer.Stop()

	var full string
	finished := false

	for !finished {
		select {
		case <-ctx.Done():
			return full, ctx.Err()
		case <-timer.C:
			return full, fmt.Errorf("session: stream timed out after %s", streamBudget)
		case <-sub.Done:
			return full, errors.New("session: broker channel closed")
		case ev := <-sub.Events:
			// Filter by requestID and event type. The stream
			// manager publishes one event name (stream.delta,
			// stream.finished, etc.) per message.
			if !eventMatchesRequest(ev, requestID) {
				continue
			}
			switch ev.Name {
			case stream.EventDelta:
				if delta, ok := stringField(ev.Data, "delta"); ok {
					full += delta
				}
			case stream.EventFinished:
				finished = true
			case stream.EventError, stream.EventCancelled:
				if msg, ok := stringField(ev.Data, "error"); ok {
					return full, fmt.Errorf("session: stream %s: %s", ev.Name, msg)
				}
				return full, fmt.Errorf("session: stream %s", ev.Name)
			}
		}
	}

	// Speak the response (best-effort). Speaker is a separate
	// sub-system; failures here must not affect the returned text.
	if full != "" && s.cfg.Speaker != nil {
		_ = s.cfg.Speaker.Speak(ctx, full)
	}

	return full, nil
}

// eventMatchesRequest returns true if the SSE event's request_id
// (or requestID) field matches the given requestID, or if the event
// has no request_id (a broadcast event).
func eventMatchesRequest(ev sse.Event, requestID string) bool {
	if ev.Data == nil {
		return true // broadcast event; no filtering
	}
	// The Data is an opaque interface{}; try to extract a request_id
	// from it via JSON round-trip.
	data, err := json.Marshal(ev.Data)
	if err != nil {
		return true
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		return true
	}
	if id, ok := fields["request_id"].(string); ok {
		return id == requestID
	}
	return true
}

// stringField returns the string value of a key in an opaque event
// payload. Returns ("", false) if the key is missing or not a string.
func stringField(data any, key string) (string, bool) {
	if data == nil {
		return "", false
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", false
	}
	var fields map[string]any
	if err := json.Unmarshal(bytes, &fields); err != nil {
		return "", false
	}
	v, ok := fields[key].(string)
	return v, ok
}

// persistUserMessage appends the user's turn to the conversation
// store. A nil store is allowed (caller chose to skip history).
func (s *Session) persistUserMessage(ctx context.Context, query string) error {
	if s.cfg.Conversation == nil || s.cfg.ConversationID == 0 {
		return nil
	}
	return s.cfg.Conversation.Append(ctx, s.cfg.ConversationID, conversation.Message{
		Role:    string(llm.RoleUser),
		Content: query,
	})
}

func (s *Session) persistAssistantMessage(ctx context.Context, reply string) error {
	if s.cfg.Conversation == nil || s.cfg.ConversationID == 0 {
		return nil
	}
	return s.cfg.Conversation.Append(ctx, s.cfg.ConversationID, conversation.Message{
		Role:    string(llm.RoleAssistant),
		Content: reply,
	})
}

// buildMessages assembles the chat history with the new user query.
// The new user query is NOT included — the caller is expected to
// have persisted it via persistUserMessage (so the next turn
// already sees it). We only prepend the previous history.
func (s *Session) buildMessages(ctx context.Context, query string) ([]llm.Message, error) {
	const historyLimit = 20
	if s.cfg.Conversation == nil || s.cfg.ConversationID == 0 {
		return []llm.Message{
			{Role: llm.RoleUser, Content: query},
		}, nil
	}
	history, err := s.cfg.Conversation.GetRecentMessages(ctx, s.cfg.ConversationID, historyLimit)
	if err != nil {
		return []llm.Message{{Role: llm.RoleUser, Content: query}}, err
	}
	out := make([]llm.Message, 0, len(history))
	for _, m := range history {
		out = append(out, llm.Message{
			Role:       llm.Role(m.Role),
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		})
	}
	// Note: since query was already persisted before buildMessages is called,
	// it will be the last message in history, so we don't need to append it.
	// But what if history is empty (e.g. database error or empty history)?
	// Normally history won't be empty because we just appended query.
	// To be extremely robust, if history is empty or the last message is not
	// the query, we can append it.
	if len(out) == 0 || out[len(out)-1].Content != query || out[len(out)-1].Role != llm.RoleUser {
		out = append(out, llm.Message{Role: llm.RoleUser, Content: query})
	}

	// Prepend memory recall as system context if available.
	if s.cfg.Memory != nil {
		memCtx, memErr := s.cfg.Memory.Recall(ctx, query, 3)
		if memErr == nil && len(memCtx) > 0 {
			ctxMsg := "Relevant context from past interactions:\n" + strings.Join(memCtx, "\n")
			out = append([]llm.Message{{Role: llm.RoleSystem, Content: ctxMsg}}, out...)
		}
	}

	// Prepend adaptive predictor context if available.
	if s.cfg.Predictor != nil {
		predCtx, predErr := s.cfg.Predictor.Predict(ctx, query)
		if predErr == nil && predCtx != "" {
			out = append([]llm.Message{{Role: llm.RoleSystem, Content: predCtx}}, out...)
		}
	}

	return out, nil
}

// Status returns the current session state. Reflects the actual
// phase of the most recent Run call: idle (no run), thinking (mid-
// stream), speaking (TTS playing), or error (failed).
func (s *Session) Status() status.Status {
	return status.Status(s.cur.Load())
}

// setStatus updates the current state and is safe to call from
// the goroutine that owns Run.
func (s *Session) setStatus(s2 status.Status) {
	// status.Status is a small int; the int32 conversion is safe
	// for any value the enum can hold.
	s.cur.Store(int32(s2)) //nolint:gosec // bounded by status enum
	if s.OnStatus != nil {
		s.OnStatus(s2)
	}
}

// Cancel aborts the currently-running session, if any. Safe to
// call from any goroutine.
func (s *Session) Cancel() {
	s.mu.Lock()
	cancel := s.cancel
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}
