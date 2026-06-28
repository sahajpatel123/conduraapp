// Package stream owns the LLM streaming lifecycle for the Synaptic
// daemon. It bridges Provider.Stream channels to the SSE broker and
// tracks in-flight requests so the GUI can cancel them.
//
// The streaming pipeline is:
//
//	GUI → llm.stream (JSON-RPC) → StreamManager.Start
//	    → Provider.Stream (<-chan StreamEvent)
//	    → StreamManager goroutine
//	    → SSE broker.PublishJSON → /events → GUI EventSource
//	GUI → llm.cancel (JSON-RPC) → StreamManager.Cancel
//	    → provider cancel func + terminal "canceled" event
//
// Each stream is identified by a ULID request_id. The GUI uses the
// request_id to correlate SSE events with the originating call and to
// cancel a specific in-flight stream.
//
// CONTRACT — assistant message persistence (Phase 15 Run #1):
// StreamManager does NOT persist the assistant's response to the
// conversation store. The GUI subscribes to the stream.* events on
// the SSE broker, accumulates the deltas, and on EventFinished it
// calls conversations.append to write the final assistant message.
// This split lets the GUI render partial deltas live and decide
// when to commit the final text. Direct-RPC callers of llm.stream
// (no SSE consumer) will see a stream.complete but no assistant
// row in the conversation — by design. llm.chat (the synchronous
// sibling RPC) DOES return the full response but does not persist
// it either; persistence is always the caller's responsibility.
package stream

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/llm"
	"github.com/sahajpatel123/synapticapp/internal/sse"
)

// Common errors returned by the stream package.
var (
	ErrAlreadyExists = errors.New("stream: request_id already in flight")
	ErrNotFound      = errors.New("stream: request_id not found")
	ErrContextFull   = errors.New("stream: conversation exceeds model context window")
	ErrHalted        = errors.New("stream: daemon is halted")
)

// Event names published to the SSE broker. The GUI listens for these.
//
// EventCancelled uses the British spelling because it is part of the
// public wire format. Changing it would break every GUI client that
// already filters on "stream.cancelled". The misspell linter is
// disabled for this constant.
const (
	EventStarted   = "stream.started"
	EventDelta     = "stream.delta"
	EventUsage     = "stream.usage"
	EventFinished  = "stream.finished"
	EventError     = "stream.error"
	EventCancelled = "stream.cancelled" //nolint:misspell // wire format
)

// payload field names used in the JSON bodies of stream events. Kept
// as constants so linters catch typos and so the wire format is
// easy to grep.
const (
	fieldRequestID      = "request_id"
	fieldConversationID = "conversation_id"
	fieldProvider       = "provider"
	fieldModel          = "model"
	fieldStartedAt      = "started_at"
	fieldDelta          = "delta"
	fieldRole           = "role"
	fieldToolCalls      = "tool_calls"
	fieldInputTokens    = "input_tokens"
	fieldOutputTokens   = "output_tokens"
	fieldTotalTokens    = "total_tokens"
	fieldFinishReason   = "finish_reason"
	fieldError          = "error"
)

// State is the lifecycle state of a single stream.
type State string

// Stream lifecycle states.
const (
	StateRunning   State = "running"
	StateFinished  State = "finished"
	StateCancelled State = "canceled"
	StateError     State = "error"
)

// Request is the input to StreamManager.Start. It is a flattened view
// of the JSON-RPC params so the daemon layer doesn't have to know the
// llm package internals.
type Request struct {
	ConversationID int64
	ProviderName   string
	Chat           llm.ChatRequest
}

// Active is the public snapshot of an in-flight stream. Returned by
// StreamManager.List so the GUI can show a "streaming now" indicator.
type Active struct {
	RequestID      string
	ConversationID int64
	ProviderName   string
	Model          string
	StartedAt      time.Time
	State          State
}

// Manager owns the set of in-flight LLM streams. It is safe for
// concurrent use; the mutex protects the map. The provider registry
// is consulted on Start, not stored, so the registry can be swapped
// without touching the manager.
type Manager struct {
	mu       sync.Mutex
	active   map[string]*activeStream
	broker   *sse.Broker
	registry *llm.Registry
	now      func() time.Time // injectable for tests
	// rootCtx is the manager's own context. Every stream's
	// provider context is derived from this, NOT from the caller's
	// context. This is critical: the caller's context is typically
	// an HTTP request context that is canceled as soon as the
	// response is sent. If we tied the stream to that, the stream
	// would die before the GUI has a chance to cancel it.
	rootCtx    context.Context
	rootCancel context.CancelFunc
	// haltFunc reports whether new streams should be refused.
	// Set via SetHaltChecker; if nil, halt is not checked.
	haltFunc func() bool
	// breakerCheck is called on Start to check if the provider's
	// circuit breaker is open. Returns false if the breaker is
	// open (call should be refused). Set via SetBreakerCheck.
	breakerCheck func(provider string) bool
	// breakerResult is called after the stream completes to
	// record success or failure on the provider's breaker.
	// Set via SetBreakerResult.
	breakerResult func(provider string, success bool)
}

type activeStream struct {
	requestID      string
	conversationID int64
	providerName   string
	model          string
	startedAt      time.Time
	cancel         func()
	state          State
	done           chan struct{} // closed when goroutine exits
	// breakerResult is called after the stream completes to
	// record success/failure on the provider's circuit breaker.
	// May be nil if no breaker is configured.
	breakerResult func(provider string, success bool)
}

// NewManager returns a fresh Manager. The broker and registry must be
// non-nil; they are not owned by the manager (the daemon owns their
// lifecycle). Call Close to cancel every in-flight stream and
// release the manager's resources.
func NewManager(broker *sse.Broker, registry *llm.Registry) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		active:     make(map[string]*activeStream),
		broker:     broker,
		registry:   registry,
		now:        time.Now,
		rootCtx:    ctx,
		rootCancel: cancel,
	}
}

// Close cancels every in-flight stream and releases the manager's
// internal context. The broker is not closed (the daemon owns it).
func (m *Manager) Close() {
	m.rootCancel()
}

// SetHaltChecker registers a function the manager will call on Start
// and Cancel. If the function returns true, Start refuses with
// ErrHalted. The function is also used by Cancel to decide whether to
// short-circuit (no streams to cancel when halted).
func (m *Manager) SetHaltChecker(fn func() bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.haltFunc = fn
}

// SetBreakerCheck registers a function called on Start to check
// whether the provider's circuit breaker allows a new call. If the
// function returns false, Start refuses with an error. This wires
// the failover package's CircuitBreaker into the streaming path
// without the stream package needing to import failover directly.
func (m *Manager) SetBreakerCheck(fn func(provider string) bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.breakerCheck = fn
}

// SetBreakerResult registers a function called after a stream
// completes (success or error) to update the provider's circuit
// breaker. This wires RecordSuccess/RecordFailure into the
// streaming path.
func (m *Manager) SetBreakerResult(fn func(provider string, success bool)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.breakerResult = fn
}

// Start kicks off a new stream. It returns immediately with a
// request_id; the actual tokens arrive on the SSE broker as
// stream.delta events tagged with this request_id.
//
// The returned cancel function is for use by Manager.Cancel only —
// callers should not invoke it directly. The stream goroutine will
// exit on its own when the provider's channel closes, when the
// provider's cancel func is called, or when the context passed to
// the manager's parent is canceled.
func (m *Manager) Start(ctx context.Context, req Request) (string, error) {
	if req.ProviderName == "" {
		return "", fmt.Errorf("%w: provider_name is required", llm.ErrNoProvider)
	}
	if req.Chat.Model == "" {
		return "", llm.ErrNoModel
	}
	if len(req.Chat.Messages) == 0 {
		return "", llm.ErrNoMessages
	}

	if m.isHalted() {
		return "", ErrHalted
	}

	// Circuit breaker: fail fast if the provider is open.
	m.mu.Lock()
	checkFn := m.breakerCheck
	resultFn := m.breakerResult
	m.mu.Unlock()
	if checkFn != nil && !checkFn(req.ProviderName) {
		return "", fmt.Errorf("circuit breaker open for provider: %s", req.ProviderName)
	}

	// Context-window check: refuse up front rather than fail
	// mid-stream. We compare against the model's declared window.
	if err := m.checkContextWindow(req); err != nil {
		return "", err
	}

	// Look up the provider up front so we fail fast on a bad name
	// rather than after the goroutine has been started.
	provider, ok := m.registry.Get(req.ProviderName)
	if !ok {
		return "", fmt.Errorf("%w: %q", llm.ErrNoProvider, req.ProviderName)
	}

	requestID := newRequestID()
	s := &activeStream{
		requestID:      requestID,
		conversationID: req.ConversationID,
		providerName:   req.ProviderName,
		model:          req.Chat.Model,
		startedAt:      m.now(),
		state:          StateRunning,
		done:           make(chan struct{}),
		breakerResult:  resultFn,
	}

	// Acquire the provider's stream channel and cancel func before
	// registering the stream, so a registration failure cannot leak
	// a goroutine. The stream context is derived from the
	// manager's root context, NOT from the caller's ctx: the
	// caller's ctx is typically an HTTP request context that is
	// canceled as soon as the response is sent, and we need the
	// stream to live long enough for the GUI to cancel it.
	streamCtx, cancelStream := context.WithCancel(m.rootCtx)
	events, providerCancel, err := provider.Stream(streamCtx, req.Chat)
	if err != nil {
		cancelStream()
		return "", err
	}
	s.cancel = func() {
		providerCancel()
		cancelStream()
	}

	m.mu.Lock()
	if _, exists := m.active[requestID]; exists {
		m.mu.Unlock()
		s.cancel()
		return "", ErrAlreadyExists
	}
	m.active[requestID] = s
	m.mu.Unlock()

	// Publish the "started" event so the GUI knows the stream is
	// live before any tokens arrive.
	m.broker.PublishJSON(EventStarted, map[string]any{
		fieldRequestID:      requestID,
		fieldConversationID: req.ConversationID,
		fieldProvider:       req.ProviderName,
		fieldModel:          req.Chat.Model,
		fieldStartedAt:      s.startedAt.UTC().Format(time.RFC3339Nano),
	})

	go m.pump(requestID, events, s)

	return requestID, nil
}

// Cancel stops an in-flight stream. It calls the provider's cancel
// func, publishes a "canceled" event, and removes the stream from
// the active map. It returns ErrNotFound if the request_id is not
// active.
func (m *Manager) Cancel(requestID string) error {
	m.mu.Lock()
	s, ok := m.active[requestID]
	if !ok {
		m.mu.Unlock()
		return ErrNotFound
	}
	delete(m.active, requestID)
	conversationID := s.conversationID
	m.mu.Unlock()

	s.cancel()

	// Wait briefly for the goroutine to drain so the GUI gets a
	// single ordered "canceled" event (not interleaved with a
	// "finished" or "delta"). 2s is generous for any provider
	// whose Stream honors context.
	select {
	case <-s.done:
	case <-time.After(2 * time.Second):
		// The goroutine did not exit in time. Publish the
		// canceled event anyway; the goroutine's late events
		// will be dropped by the SSE broker or filtered by the
		// GUI on request_id.
	}

	m.broker.PublishJSON(EventCancelled, map[string]any{
		fieldRequestID:      requestID,
		fieldConversationID: conversationID,
	})

	return nil
}

// CancelAll cancels every in-flight stream and returns the number canceled.
func (m *Manager) CancelAll() int {
	m.mu.Lock()
	ids := make([]string, 0, len(m.active))
	for id := range m.active {
		ids = append(ids, id)
	}
	m.mu.Unlock()

	for _, id := range ids {
		_ = m.Cancel(id)
	}
	return len(ids)
}

// Count returns the number of in-flight streams.
func (m *Manager) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.active)
}

// List returns a snapshot of in-flight streams. The order is not
// guaranteed.
func (m *Manager) List() []Active {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Active, 0, len(m.active))
	for _, s := range m.active {
		out = append(out, Active{
			RequestID:      s.requestID,
			ConversationID: s.conversationID,
			ProviderName:   s.providerName,
			Model:          s.model,
			StartedAt:      s.startedAt,
			State:          s.state,
		})
	}
	return out
}

// CancelByConversation cancels all streams for a given conversation.
// Used when the GUI deletes a conversation mid-stream. Returns the
// number of streams canceled.
func (m *Manager) CancelByConversation(conversationID int64) int {
	m.mu.Lock()
	var toCancel []string
	for id, s := range m.active {
		if s.conversationID == conversationID {
			toCancel = append(toCancel, id)
		}
	}
	for _, id := range toCancel {
		delete(m.active, id)
	}
	m.mu.Unlock()

	for _, id := range toCancel {
		_ = m.Cancel(id)
	}
	return len(toCancel)
}

// pump drains the provider's stream channel and republishes events
// to the SSE broker. It exits when the channel closes, the provider
// reports an error, or the stream context is canceled.
func (m *Manager) pump(requestID string, events <-chan llm.StreamEvent, s *activeStream) {
	defer close(s.done)
	defer func() {
		m.mu.Lock()
		// Only delete if this stream is still in the map; Cancel
		// may have already removed it.
		if cur, ok := m.active[requestID]; ok && cur == s {
			delete(m.active, requestID)
		}
		m.mu.Unlock()
	}()

	for ev := range events {
		if ev.Err != nil {
			m.markState(s, StateError)
			if s.breakerResult != nil {
				s.breakerResult(s.providerName, false)
			}
			m.broker.PublishJSON(EventError, map[string]any{
				fieldRequestID:      requestID,
				fieldConversationID: s.conversationID,
				fieldError:          ev.Err.Error(),
			})
			return
		}
		if ev.Done {
			m.markState(s, StateFinished)
			if s.breakerResult != nil {
				s.breakerResult(s.providerName, true)
			}
			m.broker.PublishJSON(EventFinished, map[string]any{
				fieldRequestID:      requestID,
				fieldConversationID: s.conversationID,
				fieldFinishReason:   string(ev.FinishReason),
			})
			if ev.Usage.TotalTokens > 0 || ev.Usage.InputTokens > 0 {
				m.broker.PublishJSON(EventUsage, map[string]any{
					fieldRequestID:      requestID,
					fieldConversationID: s.conversationID,
					fieldInputTokens:    ev.Usage.InputTokens,
					fieldOutputTokens:   ev.Usage.OutputTokens,
					fieldTotalTokens:    ev.Usage.TotalTokens,
				})
			}
			return
		}
		if ev.Delta.Content != "" {
			m.broker.PublishJSON(EventDelta, map[string]any{
				fieldRequestID:      requestID,
				fieldConversationID: s.conversationID,
				fieldDelta:          ev.Delta.Content,
				fieldRole:           string(ev.Delta.Role),
			})
		}
		if len(ev.Delta.ToolCalls) > 0 {
			// Tool calls are JSON-serialized as a list so the
			// GUI can apply them incrementally. Each call
			// carries id+name+args, even if args is partial.
			m.broker.PublishJSON(EventDelta, map[string]any{
				fieldRequestID:      requestID,
				fieldConversationID: s.conversationID,
				fieldToolCalls:      ev.Delta.ToolCalls,
			})
		}
		// If the final event of a non-`Done` stream carries usage,
		// surface it so the spend monitor can update.
		if ev.Usage.TotalTokens > 0 {
			m.broker.PublishJSON(EventUsage, map[string]any{
				fieldRequestID:      requestID,
				fieldConversationID: s.conversationID,
				fieldInputTokens:    ev.Usage.InputTokens,
				fieldOutputTokens:   ev.Usage.OutputTokens,
				fieldTotalTokens:    ev.Usage.TotalTokens,
			})
		}
	}

	// Channel closed without a Done event. Treat as a clean finish.
	m.markState(s, StateFinished)
	m.broker.PublishJSON(EventFinished, map[string]any{
		fieldRequestID:      requestID,
		fieldConversationID: s.conversationID,
		fieldFinishReason:   "channel_closed",
	})
}

func (m *Manager) markState(s *activeStream, st State) {
	m.mu.Lock()
	s.state = st
	m.mu.Unlock()
}

func (m *Manager) isHalted() bool {
	m.mu.Lock()
	fn := m.haltFunc
	m.mu.Unlock()
	if fn == nil {
		return false
	}
	return fn()
}

// checkContextWindow returns ErrContextFull if the requested model's
// context window cannot hold the conversation plus a reasonable
// completion budget. We use a conservative 4 chars/token estimate
// plus a 1000-token completion reserve.
func (m *Manager) checkContextWindow(req Request) error {
	info, ok := llm.LookupModel(req.Chat.Model)
	if !ok || info.ContextWindow <= 0 {
		// Unknown model or no declared window — let the provider
		// decide.
		return nil
	}
	approxTokens := 0
	for _, msg := range req.Chat.Messages {
		// 4 chars per token is a rough average for English text.
		approxTokens += len(msg.Content) / 4
	}
	reserve := 1000
	if approxTokens+reserve > info.ContextWindow {
		return fmt.Errorf("%w: model=%s window=%d need=%d",
			ErrContextFull, req.Chat.Model, info.ContextWindow, approxTokens+reserve)
	}
	return nil
}

// newRequestID returns a 16-byte hex string. Uniqueness is provided
// by the random source, not the timestamp, so concurrent Start calls
// cannot collide.
func newRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
