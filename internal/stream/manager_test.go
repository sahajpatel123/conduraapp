package stream

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/llm"
	"github.com/sahajpatel123/synapticapp/internal/sse"
)

// fakeProvider is a controllable Provider for tests. It returns a
// channel of events, and a cancel func that closes the channel.
type fakeProvider struct {
	name        string
	models      []llm.ModelInfo
	events      []llm.StreamEvent
	chatFn      func(ctx context.Context, req llm.ChatRequest) (llm.ChatResponse, error)
	cancelCalls atomic.Int32
}

func (f *fakeProvider) Name() string { return f.name }

func (f *fakeProvider) Chat(ctx context.Context, req llm.ChatRequest) (llm.ChatResponse, error) {
	if f.chatFn != nil {
		return f.chatFn(ctx, req)
	}
	return llm.ChatResponse{}, nil
}

func (f *fakeProvider) Models() []llm.ModelInfo { return f.models }

func (f *fakeProvider) DefaultModel(task string) string { return "" }

// Stream returns a channel that yields the configured events. The
// cancel func closes the channel and increments cancelCalls. If
// events is nil, the channel stays open until cancel.
func (f *fakeProvider) Stream(ctx context.Context, req llm.ChatRequest) (<-chan llm.StreamEvent, func(), error) {
	out := make(chan llm.StreamEvent, len(f.events)+1)
	canceled := make(chan struct{})
	for i := range f.events {
		out <- f.events[i]
	}
	if len(f.events) == 0 {
		// Block until cancel.
		go func() {
			select {
			case <-ctx.Done():
			case <-canceled:
			}
			close(out)
		}()
	} else {
		close(out)
	}
	cancel := func() {
		f.cancelCalls.Add(1)
		close(canceled)
	}
	return out, cancel, nil
}

// recorder captures all events published to the SSE broker for a
// specific request_id. It reads from a bufio.Scanner over an HTTP
// response body so it exercises the real ServeHTTP path.
type recorder struct {
	cancel context.CancelFunc
	done   chan struct{}
	mu     sync.Mutex
	events []parsedEvent
}

type parsedEvent struct {
	name string
	data map[string]any
	id   string
}

func newRecorder(broker *sse.Broker) *recorder {
	srv := httptest.NewServer(broker)
	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	//nolint:bodyclose // resp.Body is closed in the scanner goroutine below
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		cancel()
		srv.Close()
		return &recorder{done: make(chan struct{})}
	}
	r := &recorder{
		cancel: cancel,
		done:   make(chan struct{}),
	}
	go func() {
		defer close(r.done)
		defer func() {
			_ = resp.Body.Close()
			srv.Close()
		}()
		scanner := bufio.NewScanner(resp.Body)
		// SSE events can be larger than the default 64KB scanner
		// buffer; bump it.
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		var cur parsedEvent
		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.HasPrefix(line, "event: "):
				cur.name = strings.TrimPrefix(line, "event: ")
			case strings.HasPrefix(line, "data: "):
				payload := strings.TrimPrefix(line, "data: ")
				cur.data = map[string]any{}
				_ = json.Unmarshal([]byte(payload), &cur.data)
			case strings.HasPrefix(line, "id: "):
				cur.id = strings.TrimPrefix(line, "id: ")
			case line == "":
				if cur.name != "" || cur.data != nil {
					r.mu.Lock()
					r.events = append(r.events, cur)
					r.mu.Unlock()
					cur = parsedEvent{}
				}
			}
		}
	}()
	return r
}

func (r *recorder) close() {
	r.cancel()
	<-r.done
}

// drain reads events from a recorder until count stream events are
// seen (skipping the initial "connected" handshake event from the
// SSE broker) or the timeout elapses.
func drain(t *testing.T, r *recorder, count int, timeout time.Duration) []parsedEvent {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		r.mu.Lock()
		all := append([]parsedEvent(nil), r.events...)
		r.mu.Unlock()
		streamEvents := filterStreamEvents(all)
		if len(streamEvents) >= count {
			return streamEvents
		}
		time.Sleep(10 * time.Millisecond)
	}
	r.mu.Lock()
	all := append([]parsedEvent(nil), r.events...)
	r.mu.Unlock()
	return filterStreamEvents(all)
}

// filterStreamEvents removes the SSE broker's "connected" handshake
// event so tests see only the events the manager published.
func filterStreamEvents(in []parsedEvent) []parsedEvent {
	out := in[:0]
	for _, e := range in {
		if e.name == "connected" {
			continue
		}
		out = append(out, e)
	}
	return out
}

// newTestManager wires a manager with a real broker and registry.
func newTestManager(t *testing.T, providers ...llm.Provider) (*Manager, *sse.Broker) {
	t.Helper()
	broker := sse.NewBroker()
	reg := llm.NewRegistry()
	for _, p := range providers {
		reg.Register(p)
	}
	return NewManager(broker, reg), broker
}

// TestManager_StartReturnsRequestID verifies that Start returns a
// non-empty request_id and registers the stream.
func TestManager_StartReturnsRequestID(t *testing.T) {
	p := &fakeProvider{
		name: "fake",
		models: []llm.ModelInfo{
			{ID: "fake-1", ContextWindow: 4096},
		},
		events: []llm.StreamEvent{
			{Delta: llm.Message{Role: llm.RoleAssistant, Content: "hi"}},
			{Done: true, FinishReason: llm.FinishStop},
		},
	}
	m, _ := newTestManager(t, p)
	ctx := context.Background()

	id, err := m.Start(ctx, Request{
		ConversationID: 1,
		ProviderName:   "fake",
		Chat: llm.ChatRequest{
			Model:    "fake-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hello"}},
		},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if id == "" {
		t.Fatal("empty request_id")
	}
	if got := m.Count(); got != 1 {
		t.Fatalf("Count = %d, want 1", got)
	}
	// Wait for the goroutine to drain.
	if !waitFor(t, func() bool { return m.Count() == 0 }, time.Second) {
		t.Fatalf("stream did not finish: Count = %d", m.Count())
	}
}

// TestManager_StartPublishesEvents verifies that the manager
// publishes started, delta, and finished events.
func TestManager_StartPublishesEvents(t *testing.T) {
	p := &fakeProvider{
		name:   "fake",
		models: []llm.ModelInfo{{ID: "fake-1", ContextWindow: 4096}},
		events: []llm.StreamEvent{
			{Delta: llm.Message{Content: "hello"}},
			{Delta: llm.Message{Content: " world"}},
			{Done: true, FinishReason: llm.FinishStop},
		},
	}
	m, broker := newTestManager(t, p)
	r := newRecorder(broker)
	defer r.close()

	id, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model:    "fake-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Expect at least: started, 2 deltas, finished = 4 events.
	evs := drain(t, r, 4, 2*time.Second)
	if len(evs) < 4 {
		t.Fatalf("got %d events, want >=4", len(evs))
	}
	// Verify event names.
	want := []string{EventStarted, EventDelta, EventDelta, EventFinished}
	for i, ev := range evs[:4] {
		if ev.name != want[i] {
			t.Errorf("event %d: name = %q, want %q", i, ev.name, want[i])
		}
		// Every event must carry the same request_id.
		if ev.data["request_id"] != id {
			t.Errorf("event %d: request_id = %v, want %s", i, ev.data["request_id"], id)
		}
	}
}

// TestManager_CancelStopsStream verifies that Cancel calls the
// provider's cancel func and publishes a canceled event.
func TestManager_CancelStopsStream(t *testing.T) {
	p := &fakeProvider{
		name:   "fake",
		models: []llm.ModelInfo{{ID: "fake-1", ContextWindow: 4096}},
		// No events; Stream blocks until cancel.
	}
	m, broker := newTestManager(t, p)
	r := newRecorder(broker)
	defer r.close()

	id, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model:    "fake-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Wait for the started event to confirm the stream is live.
	started := drain(t, r, 1, 2*time.Second)
	if len(started) < 1 || started[0].name != EventStarted {
		t.Fatalf("expected started event, got %v", started)
	}

	if err := m.Cancel(id); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	// After Cancel, the pump exits (publishing "finished" with
	// reason "channel_closed") and then Cancel publishes
	// "canceled". We expect 3 stream events total.
	all := drain(t, r, 3, 3*time.Second)
	if len(all) < 3 {
		t.Fatalf("got %d events, want >=3", len(all))
	}
	if all[0].name != EventStarted {
		t.Errorf("event 0 name = %q, want %q", all[0].name, EventStarted)
	}
	if all[1].name != EventFinished {
		t.Errorf("event 1 name = %q, want %q", all[1].name, EventFinished)
	}
	if all[2].name != EventCancelled {
		t.Errorf("event 2 name = %q, want %q", all[2].name, EventCancelled)
	}
	if p.cancelCalls.Load() != 1 {
		t.Fatalf("cancel calls = %d, want 1", p.cancelCalls.Load())
	}
	if got := m.Count(); got != 0 {
		t.Fatalf("Count after cancel = %d, want 0", got)
	}
}

// TestManager_CancelUnknownRequestReturnsError verifies Cancel's
// error path.
func TestManager_CancelUnknownRequestReturnsError(t *testing.T) {
	m, _ := newTestManager(t)
	if err := m.Cancel("nonexistent"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Cancel(nonexistent) = %v, want ErrNotFound", err)
	}
}

// TestManager_StartRejectsUnknownProvider verifies fail-fast on a
// bad provider name.
func TestManager_StartRejectsUnknownProvider(t *testing.T) {
	m, _ := newTestManager(t)
	_, err := m.Start(context.Background(), Request{
		ProviderName: "missing",
		Chat: llm.ChatRequest{
			Model:    "m",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if !errors.Is(err, llm.ErrNoProvider) {
		t.Fatalf("Start(unknown) = %v, want ErrNoProvider", err)
	}
}

// TestManager_StartRejectsEmptyModel verifies the empty-model guard.
func TestManager_StartRejectsEmptyModel(t *testing.T) {
	p := &fakeProvider{name: "fake", models: []llm.ModelInfo{{ID: "fake-1"}}}
	m, _ := newTestManager(t, p)
	_, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if !errors.Is(err, llm.ErrNoModel) {
		t.Fatalf("Start(no model) = %v, want ErrNoModel", err)
	}
}

// TestManager_StartRejectsEmptyMessages verifies the empty-messages
// guard.
func TestManager_StartRejectsEmptyMessages(t *testing.T) {
	p := &fakeProvider{name: "fake", models: []llm.ModelInfo{{ID: "fake-1"}}}
	m, _ := newTestManager(t, p)
	_, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model: "fake-1",
		},
	})
	if !errors.Is(err, llm.ErrNoMessages) {
		t.Fatalf("Start(no messages) = %v, want ErrNoMessages", err)
	}
}

// TestManager_ContextWindowOverflow verifies that Start refuses
// when the conversation exceeds the model's context window.
func TestManager_ContextWindowOverflow(t *testing.T) {
	p := &fakeProvider{
		name:   "fake",
		models: []llm.ModelInfo{{ID: "fake-tiny", ContextWindow: 100}},
	}
	m, _ := newTestManager(t, p)
	// Register the model in the global pricing registry so
	// checkContextWindow can find its context window.
	llm.RegisterModel(llm.ModelInfo{ID: "fake-tiny", ContextWindow: 100})
	t.Cleanup(func() { llm.UnregisterModel("fake-tiny") })
	// Build a message that, at 4 chars/token, needs > 100 tokens.
	big := strings.Repeat("a", 1000)
	_, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model:    "fake-tiny",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: big}},
		},
	})
	if !errors.Is(err, ErrContextFull) {
		t.Fatalf("Start(overflow) = %v, want ErrContextFull", err)
	}
}

// TestManager_ContextWindowUnknownAllowsStream verifies that
// unknown models with no declared context window pass the check.
func TestManager_ContextWindowUnknownAllowsStream(t *testing.T) {
	p := &fakeProvider{
		name:   "fake",
		models: []llm.ModelInfo{}, // no models registered in fake
		events: []llm.StreamEvent{{Done: true}},
	}
	m, _ := newTestManager(t, p)
	_, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model:    "unknown-model",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if err != nil {
		t.Fatalf("Start(unknown model) = %v, want nil", err)
	}
}

// TestManager_HaltRefusesStart verifies that the halt checker, when
// set, causes Start to refuse.
func TestManager_HaltRefusesStart(t *testing.T) {
	p := &fakeProvider{name: "fake", models: []llm.ModelInfo{{ID: "fake-1"}}}
	m, _ := newTestManager(t, p)
	m.SetHaltChecker(func() bool { return true })
	_, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model:    "fake-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if !errors.Is(err, ErrHalted) {
		t.Fatalf("Start when halted = %v, want ErrHalted", err)
	}
}

// TestManager_ErrorEventOnStreamError verifies that an error from
// the provider's stream channel is surfaced as EventError.
func TestManager_ErrorEventOnStreamError(t *testing.T) {
	p := &fakeProvider{
		name:   "fake",
		models: []llm.ModelInfo{{ID: "fake-1"}},
		events: []llm.StreamEvent{
			{Delta: llm.Message{Content: "partial"}},
			{Err: errors.New("upstream 500")},
		},
	}
	m, broker := newTestManager(t, p)
	r := newRecorder(broker)
	defer r.close()
	_, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model:    "fake-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	evs := drain(t, r, 3, 2*time.Second) // started, delta, error
	if len(evs) < 3 {
		t.Fatalf("got %d events, want >=3", len(evs))
	}
	if evs[2].name != EventError {
		t.Fatalf("event 2 name = %q, want %q", evs[2].name, EventError)
	}
	if !strings.Contains(evs[2].data["error"].(string), "upstream 500") {
		t.Fatalf("error payload = %v, want 'upstream 500'", evs[2].data["error"])
	}
}

// TestManager_RequestIDsAreUnique verifies that two concurrent
// Start calls get distinct IDs.
func TestManager_RequestIDsAreUnique(t *testing.T) {
	p := &fakeProvider{
		name:   "fake",
		models: []llm.ModelInfo{{ID: "fake-1"}},
		events: []llm.StreamEvent{{Done: true}},
	}
	m, _ := newTestManager(t, p)
	ids := make(map[string]bool)
	for i := 0; i < 20; i++ {
		id, err := m.Start(context.Background(), Request{
			ProviderName: "fake",
			Chat: llm.ChatRequest{
				Model:    "fake-1",
				Messages: []llm.Message{{Role: llm.RoleUser, Content: "x"}},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if ids[id] {
			t.Fatalf("duplicate request_id: %s", id)
		}
		ids[id] = true
	}
}

// TestManager_CancelByConversationCancelsAll verifies that all
// streams for a conversation are canceled.
func TestManager_CancelByConversationCancelsAll(t *testing.T) {
	p := &fakeProvider{
		name:   "fake",
		models: []llm.ModelInfo{{ID: "fake-1"}},
		// no events; streams block
	}
	m, _ := newTestManager(t, p)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		_, err := m.Start(ctx, Request{
			ConversationID: 42,
			ProviderName:   "fake",
			Chat: llm.ChatRequest{
				Model:    "fake-1",
				Messages: []llm.Message{{Role: llm.RoleUser, Content: "x"}},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := m.Count(); got != 3 {
		t.Fatalf("Count = %d, want 3", got)
	}
	n := m.CancelByConversation(42)
	if n != 3 {
		t.Fatalf("CancelByConversation = %d, want 3", n)
	}
	if got := m.Count(); got != 0 {
		t.Fatalf("Count after cancel = %d, want 0", got)
	}
}

// TestManager_ListReturnsActive verifies List.
func TestManager_ListReturnsActive(t *testing.T) {
	p := &fakeProvider{
		name:   "fake",
		models: []llm.ModelInfo{{ID: "fake-1"}},
		events: []llm.StreamEvent{}, // empty; blocks
	}
	m, _ := newTestManager(t, p)
	_, err := m.Start(context.Background(), Request{
		ConversationID: 7,
		ProviderName:   "fake",
		Chat: llm.ChatRequest{
			Model:    "fake-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "x"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	active := m.List()
	if len(active) != 1 {
		t.Fatalf("List len = %d, want 1", len(active))
	}
	if active[0].ConversationID != 7 {
		t.Fatalf("ConversationID = %d, want 7", active[0].ConversationID)
	}
	if active[0].State != StateRunning {
		t.Fatalf("State = %q, want %q", active[0].State, StateRunning)
	}
}

// waitFor polls fn every 10ms until it returns true or timeout.
func waitFor(t *testing.T, fn func() bool, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}
