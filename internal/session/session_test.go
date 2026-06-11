package session

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/llm"
	"github.com/sahajpatel123/synapticapp/internal/sse"
	"github.com/sahajpatel123/synapticapp/internal/status"
	"github.com/sahajpatel123/synapticapp/internal/stream"
)

// fakeProvider records the call and returns the configured
// response. Streaming is mocked by emitting deltas to the
// stream manager's broker.
type fakeProvider struct {
	mu    sync.Mutex
	calls int
}

func (p *fakeProvider) Chat(_ context.Context, _ string, _ llm.ChatRequest) (llm.ChatResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	return llm.ChatResponse{}, nil
}

// TestSession_ReturnsReplyFromBrokerDeltas is the keystone test
// for 6A #1: the session must return the full text the model
// emitted, accumulated from the SSE broker's stream.delta events.
//
// This test would have caught the bug where collectAndSpeak read
// the reply back from the conversation store (which never persisted
// it). The fix is to subscribe to the broker and accumulate
// stream.delta events.
type mockLLMProvider struct {
	name string
}

func (p *mockLLMProvider) Name() string { return p.name }
func (p *mockLLMProvider) Chat(_ context.Context, _ llm.ChatRequest) (llm.ChatResponse, error) {
	return llm.ChatResponse{}, nil
}
func (p *mockLLMProvider) Stream(_ context.Context, _ llm.ChatRequest) (<-chan llm.StreamEvent, func(), error) {
	ch := make(chan llm.StreamEvent, 10)
	cancel := func() {}
	go func() {
		defer close(ch)
		ch <- llm.StreamEvent{Delta: llm.Message{Content: "Hello"}}
		ch <- llm.StreamEvent{Delta: llm.Message{Content: ", "}}
		ch <- llm.StreamEvent{Delta: llm.Message{Content: "world!"}}
		ch <- llm.StreamEvent{Done: true}
	}()
	return ch, cancel, nil
}
func (p *mockLLMProvider) Models() []llm.ModelInfo      { return nil }
func (p *mockLLMProvider) DefaultModel(_ string) string { return "" }

// TestSession_ReturnsReplyFromBrokerDeltas is the keystone test
// for 6A #1: the session must return the full text the model
// emitted, accumulated from the SSE broker's stream.delta events.
func TestSession_ReturnsReplyFromBrokerDeltas(t *testing.T) {
	broker := sse.NewBroker()
	defer broker.Close()

	// Register the mock provider in a real registry
	reg := llm.NewRegistry()
	mockProv := &mockLLMProvider{name: "test"}
	reg.Register(mockProv)

	// Build a session with a real stream manager and real broker.
	mgr := stream.NewManager(broker, reg)
	defer mgr.Close()

	cfg := Config{
		StreamMgr:    mgr,
		Provider:     reg, // Registry implements session.Provider
		ProviderName: "test",
		Model:        "test-model",
		Broker:       broker,
	}
	s, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	full, err := s.Run(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if full != "Hello, world!" {
		t.Errorf("accumulated = %q, want %q", full, "Hello, world!")
	}
}

// TestEventMatchesRequest_AcceptsBroadcast verifies that events
// without a request_id are not filtered out.
func TestEventMatchesRequest_AcceptsBroadcast(t *testing.T) {
	ev := sse.Event{Name: "ping", Data: nil}
	if !eventMatchesRequest(ev, "any") {
		t.Error("event with nil Data should be treated as broadcast")
	}
}

// TestEventMatchesRequest_FiltersByID verifies that events with a
// matching request_id are accepted.
func TestEventMatchesRequest_FiltersByID(t *testing.T) {
	ev := sse.Event{
		Name: "delta",
		Data: map[string]any{"request_id": "abc"},
	}
	if !eventMatchesRequest(ev, "abc") {
		t.Error("event with matching request_id should be accepted")
	}
	if eventMatchesRequest(ev, "xyz") {
		t.Error("event with non-matching request_id should be rejected")
	}
}

// TestStringField verifies the helper for extracting string fields.
func TestStringField(t *testing.T) {
	tests := []struct {
		name   string
		data   any
		key    string
		want   string
		wantOK bool
	}{
		{"nil data", nil, "x", "", false},
		{"map with string", map[string]any{"x": "hello"}, "x", "hello", true},
		{"map with non-string", map[string]any{"x": 42}, "x", "", false},
		{"map without key", map[string]any{"y": "hello"}, "x", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := stringField(tt.data, tt.key)
			if ok != tt.wantOK || got != tt.want {
				t.Errorf("stringField(%v, %q) = (%q, %v), want (%q, %v)",
					tt.data, tt.key, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

// TestSession_New_RequiresBroker verifies New rejects a nil broker
// (the field was added for 6A #1).
func TestSession_New_RequiresBroker(t *testing.T) {
	mgr := stream.NewManager(sse.NewBroker(), nil)
	defer mgr.Close()
	_, err := New(Config{
		StreamMgr:    mgr,
		Provider:     &fakeProvider{},
		ProviderName: "x",
		Model:        "y",
		// Broker deliberately omitted.
	})
	if err == nil {
		t.Fatal("expected error for nil Broker")
	}
}

// TestSession_New_OK verifies New accepts a valid config.
func TestSession_New_OK(t *testing.T) {
	mgr := stream.NewManager(sse.NewBroker(), nil)
	defer mgr.Close()
	cfg := Config{
		StreamMgr:    mgr,
		Provider:     &fakeProvider{},
		ProviderName: "x",
		Model:        "y",
		Broker:       sse.NewBroker(),
	}
	s, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if s == nil {
		t.Fatal("nil session")
	}
	if s.Status() != status.StatusIdle {
		t.Errorf("Status = %v, want idle", s.Status())
	}
}

// TestRun_EmptyQueryReturnsImmediately verifies the empty-query
// short-circuit returns idle and no error.
func TestRun_EmptyQueryReturnsImmediately(t *testing.T) {
	mgr := stream.NewManager(sse.NewBroker(), nil)
	defer mgr.Close()
	s, err := New(Config{
		StreamMgr:    mgr,
		Provider:     &fakeProvider{},
		ProviderName: "x",
		Model:        "y",
		Broker:       sse.NewBroker(),
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := s.Run(context.Background(), "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out != "" {
		t.Errorf("expected empty output, got %q", out)
	}
	if s.Status() != status.StatusIdle {
		t.Errorf("Status = %v, want idle", s.Status())
	}
}

// TestRun_AlreadyRunning verifies the busy flag.
func TestRun_AlreadyRunning(t *testing.T) {
	mgr := stream.NewManager(sse.NewBroker(), nil)
	defer mgr.Close()
	s := &Session{cfg: Config{
		StreamMgr:    mgr,
		Provider:     &fakeProvider{},
		ProviderName: "x",
		Model:        "y",
		Broker:       sse.NewBroker(),
	}}
	s.busy = true
	defer func() { s.busy = false }()

	_, err := s.Run(context.Background(), "hello")
	if !errIs(err, ErrAlreadyRunning) {
		t.Errorf("err = %v, want ErrAlreadyRunning", err)
	}
}

// errIs is a tiny local helper to avoid the errors import
// duplication. Use errors.Is for production checks; this is a
// defensive last-resort helper.
func errIs(err, target error) bool {
	for err != nil {
		if errors.Is(err, target) {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		if u, ok := err.(unwrapper); ok {
			err = u.Unwrap()
			continue
		}
		return false
	}
	return false
}

// TestBuildMessages_NoConversation verifies the simple path
// (no history store configured).
func TestBuildMessages_NoConversation(t *testing.T) {
	s := &Session{cfg: Config{}}
	msgs, err := s.buildMessages(context.Background(), "hi")
	if err != nil {
		t.Fatalf("buildMessages: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Role != llm.RoleUser || msgs[0].Content != "hi" {
		t.Errorf("msg = %+v", msgs[0])
	}
}

// TestSession_Cancel verifies Cancel doesn't panic and doesn't
// deadlock when called on an idle session.
func TestSession_Cancel(t *testing.T) {
	mgr := stream.NewManager(sse.NewBroker(), nil)
	defer mgr.Close()
	s, err := New(Config{
		StreamMgr:    mgr,
		Provider:     &fakeProvider{},
		ProviderName: "x",
		Model:        "y",
		Broker:       sse.NewBroker(),
	})
	if err != nil {
		t.Fatal(err)
	}
	// Should be a no-op when idle.
	s.Cancel()
}
