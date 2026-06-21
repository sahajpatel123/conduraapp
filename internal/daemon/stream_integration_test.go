// Integration tests for the LLM streaming pipeline. These tests
// wire up the real ipc.ServerTransport + sse.Broker + stream.Manager
// and verify that:
//
//   - llm.stream returns a request_id
//   - tokens arrive on /events as "stream.delta" events
//   - llm.cancel stops the stream and publishes "stream.canceled"
//
// The tests use the real HTTP transport, the real SSE broker, and the
// real JSON-RPC server — only the LLM provider is faked.
package daemon

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/conversation"
	"github.com/sahajpatel123/synapticapp/internal/halt"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
	"github.com/sahajpatel123/synapticapp/internal/llm"
	"github.com/sahajpatel123/synapticapp/internal/sse"
	"github.com/sahajpatel123/synapticapp/internal/storage"
	"github.com/sahajpatel123/synapticapp/internal/stream"
)

// fakeStreamingProvider is a controllable Provider that yields a
// configurable sequence of stream events.
type fakeStreamingProvider struct {
	name   string
	events []llm.StreamEvent
}

func (f *fakeStreamingProvider) Name() string { return f.name }

func (f *fakeStreamingProvider) Chat(_ context.Context, _ llm.ChatRequest) (llm.ChatResponse, error) {
	return llm.ChatResponse{}, nil
}

func (f *fakeStreamingProvider) Models() []llm.ModelInfo { return nil }

func (f *fakeStreamingProvider) DefaultModel(_ string) string { return "fake-1" }

func (f *fakeStreamingProvider) Stream(ctx context.Context, _ llm.ChatRequest) (<-chan llm.StreamEvent, func(), error) {
	out := make(chan llm.StreamEvent, len(f.events))
	canceled := make(chan struct{})
	for i := range f.events {
		out <- f.events[i]
	}
	close(out)
	cancel := func() { close(canceled) }
	_ = ctx
	_ = canceled
	return out, cancel, nil
}

// blockingProvider is a Provider whose Stream blocks until cancel.
type blockingProvider struct{ name string }

func (b *blockingProvider) Name() string { return b.name }
func (b *blockingProvider) Chat(_ context.Context, _ llm.ChatRequest) (llm.ChatResponse, error) {
	return llm.ChatResponse{}, nil
}
func (b *blockingProvider) Models() []llm.ModelInfo      { return nil }
func (b *blockingProvider) DefaultModel(_ string) string { return "block-1" }

func (b *blockingProvider) Stream(ctx context.Context, _ llm.ChatRequest) (<-chan llm.StreamEvent, func(), error) {
	out := make(chan llm.StreamEvent)
	canceled := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
		case <-canceled:
		}
		close(out)
	}()
	cancel := func() { close(canceled) }
	return out, cancel, nil
}

// bringUpPipeline wires the broker, stream manager, and a fresh
// JSON-RPC server with the streaming methods registered. It returns
// the bound HTTP address and a cleanup func.
func bringUpPipeline(t *testing.T, providers ...llm.Provider) (addr string, mgr *stream.Manager, cleanup func()) {
	t.Helper()
	broker := sse.NewBroker()
	registry := llm.NewRegistry()
	for _, p := range providers {
		registry.Register(p)
	}
	mgr = stream.NewManager(broker, registry)
	// halt flag with no DB backing — the manager's halt check
	// will simply report false for these tests.
	mgr.SetHaltChecker(func() bool { return false })

	// Use a real on-disk SQLite DB so audit.Append has a valid
	// handle. The DB is closed in cleanup.
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{
		Path: filepath.Join(dir, "test.db"),
	})
	if err != nil {
		t.Fatal(err)
	}
	auditLog := audit.New(db.SQL(), db.MasterKey())
	convStore := conversation.New(db.SQL())
	haltFlag := halt.New(db.SQL())
	_ = haltFlag.Refresh(context.Background())

	srv := ipc.NewServer()
	registerConversationMethods(srv, convStore, auditLog, haltFlag, mgr, registry, nil, nil)

	transport := &ipc.ServerTransport{
		S:   srv,
		SSE: broker,
	}
	if err := transport.Listen(context.Background(), "tcp://127.0.0.1:0"); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	addr = transport.Addr()

	cleanup = func() {
		_ = transport.Close()
		_ = db.Close()
	}
	return addr, mgr, cleanup
}

// streamEventRecorder subscribes to /events over HTTP and records
// the parsed events as they arrive.
type streamEventRecorder struct {
	cancel context.CancelFunc
	done   chan struct{}
	mu     sync.Mutex
	events []parsedStreamEvent
}

type parsedStreamEvent struct {
	name string
	data map[string]any
}

// subscribe starts a goroutine that reads SSE events from the given
// URL until cancel() is called.
func (r *streamEventRecorder) subscribe(url string) error {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	r.done = make(chan struct{})
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		cancel()
		return err
	}
	//nolint:bodyclose // closed in goroutine
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		cancel()
		return err
	}
	go func() {
		defer close(r.done)
		defer func() { _ = resp.Body.Close() }()
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		var cur parsedStreamEvent
		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.HasPrefix(line, "event: "):
				cur.name = strings.TrimPrefix(line, "event: ")
			case strings.HasPrefix(line, "data: "):
				cur.data = map[string]any{}
				_ = json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &cur.data)
			case line == "":
				if cur.name != "" {
					r.mu.Lock()
					r.events = append(r.events, cur)
					r.mu.Unlock()
				}
				cur = parsedStreamEvent{}
			}
		}
	}()
	return nil
}

func (r *streamEventRecorder) close() {
	r.cancel()
	<-r.done
}

func (r *streamEventRecorder) snapshot() []parsedStreamEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]parsedStreamEvent, len(r.events))
	copy(out, r.events)
	return out
}

// callRPC sends a JSON-RPC 2.0 request to the daemon's HTTP
// endpoint and returns the result.
func callRPC(t *testing.T, addr, method string, params any) (map[string]any, *ipc.Error) {
	t.Helper()
	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	}
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/", strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	//nolint:bodyclose // http.Client.Do body
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var rpc struct {
		Result json.RawMessage `json:"result"`
		Error  *ipc.Error      `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpc); err != nil {
		t.Fatal(err)
	}
	out := map[string]any{}
	if rpc.Result != nil {
		_ = json.Unmarshal(rpc.Result, &out)
	}
	return out, rpc.Error
}

// TestStream_EndToEnd starts the pipeline, registers a fake
// provider that yields two tokens, calls llm.stream, and verifies
// that the tokens arrive on /events as "stream.delta" events.
func TestStream_EndToEnd(t *testing.T) {
	addr, mgr, cleanup := bringUpPipeline(t, &fakeStreamingProvider{
		name: "fake",
		events: []llm.StreamEvent{
			{Delta: llm.Message{Content: "hello"}},
			{Delta: llm.Message{Content: " world"}},
			{Done: true, FinishReason: llm.FinishStop},
		},
	})
	defer cleanup()

	rec := &streamEventRecorder{}
	if err := rec.subscribe("http://" + addr + "/events"); err != nil {
		t.Fatal(err)
	}
	defer rec.close()
	time.Sleep(100 * time.Millisecond)

	res, rpcErr := callRPC(t, addr, "llm.stream", map[string]any{
		"provider": "fake",
		"request": map[string]any{
			"model": "fake-1",
			"messages": []map[string]any{
				{"role": "user", "content": "hi"},
			},
		},
	})
	if rpcErr != nil {
		t.Fatalf("llm.stream: %v", rpcErr)
	}
	requestID, ok := res["request_id"].(string)
	if !ok || requestID == "" {
		t.Fatalf("llm.stream: missing request_id in result: %v", res)
	}

	// Wait for events: at least started, 2 deltas, finished.
	deadline := time.Now().Add(3 * time.Second)
	var events []parsedStreamEvent
	for time.Now().Before(deadline) {
		all := rec.snapshot()
		events = filterStreamEvents(all)
		if len(events) >= 4 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if len(events) < 4 {
		t.Fatalf("got %d events, want >=4: %v", len(events), events)
	}
	want := []string{"stream.started", "stream.delta", "stream.delta", "stream.finished"}
	for i, ev := range events[:4] {
		if ev.name != want[i] {
			t.Errorf("event %d: name = %q, want %q", i, ev.name, want[i])
		}
		if ev.data["request_id"] != requestID {
			t.Errorf("event %d: request_id mismatch: got %v, want %s", i, ev.data["request_id"], requestID)
		}
	}
	if events[1].data["delta"] != "hello" {
		t.Errorf("delta 1 = %v, want 'hello'", events[1].data["delta"])
	}
	if events[2].data["delta"] != " world" {
		t.Errorf("delta 2 = %v, want ' world'", events[2].data["delta"])
	}
	if got := mgr.Count(); got != 0 {
		t.Errorf("Count after stream = %d, want 0", got)
	}
}

// TestStream_CancelStopsStream verifies that llm.cancel stops an
// in-flight stream and publishes the "stream.canceled" event.
func TestStream_CancelStopsStream(t *testing.T) {
	addr, mgr, cleanup := bringUpPipeline(t, &blockingProvider{name: "blocking"})
	defer cleanup()

	rec := &streamEventRecorder{}
	if err := rec.subscribe("http://" + addr + "/events"); err != nil {
		t.Fatal(err)
	}
	defer rec.close()
	time.Sleep(100 * time.Millisecond)

	res, rpcErr := callRPC(t, addr, "llm.stream", map[string]any{
		"provider": "blocking",
		"request": map[string]any{
			"model": "block-1",
			"messages": []map[string]any{
				{"role": "user", "content": "hi"},
			},
		},
	})
	if rpcErr != nil {
		t.Fatalf("llm.stream: %v", rpcErr)
	}
	requestID := res["request_id"].(string)

	if !waitForEvents(t, rec, "stream.started", 2*time.Second) {
		t.Fatal("did not see started event")
	}
	// Small sleep to ensure the manager has registered the stream
	// in its map (Start publishes the started event before
	// returning, but the registration order is: register, then
	// publish, then start goroutine. The test reads the request_id
	// from the RPC response — by then the registration is done.
	// This sleep is defensive and can be removed once we trust the
	// ordering.)
	time.Sleep(50 * time.Millisecond)

	cancelRes, cancelErr := callRPC(t, addr, "llm.cancel", map[string]any{
		"request_id": requestID,
	})
	if cancelErr != nil {
		t.Fatalf("llm.cancel: %v", cancelErr)
	}
	if canceled, _ := cancelRes["canceled"].(bool); !canceled {
		t.Errorf("cancel response: %v", cancelRes)
	}

	deadline := time.Now().Add(3 * time.Second)
	var events []parsedStreamEvent
	for time.Now().Before(deadline) {
		events = rec.snapshot()
		hasCanceled := false
		for _, ev := range events {
			if ev.name == stream.EventCancelled {
				hasCanceled = true
				break
			}
		}
		if hasCanceled {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	hasCanceled := false
	for _, ev := range events {
		if ev.name == stream.EventCancelled {
			hasCanceled = true
		}
	}
	if !hasCanceled {
		t.Fatalf("never saw '%s' event: %v", stream.EventCancelled, events)
	}

	if got := mgr.Count(); got != 0 {
		t.Errorf("Count after cancel = %d, want 0", got)
	}
}

// TestStream_UnknownProviderReturnsError verifies error handling
// when the requested provider is not registered.
func TestStream_UnknownProviderReturnsError(t *testing.T) {
	addr, _, cleanup := bringUpPipeline(t)
	defer cleanup()

	_, rpcErr := callRPC(t, addr, "llm.stream", map[string]any{
		"provider": "nonexistent",
		"request": map[string]any{
			"model":    "x",
			"messages": []map[string]any{{"role": "user", "content": "hi"}},
		},
	})
	if rpcErr == nil {
		t.Fatal("expected error for unknown provider")
	}
}

// TestStream_CancelUnknownRequestReturnsError verifies that
// canceling a non-existent request_id is rejected cleanly.
func TestStream_CancelUnknownRequestReturnsError(t *testing.T) {
	addr, _, cleanup := bringUpPipeline(t)
	defer cleanup()

	_, rpcErr := callRPC(t, addr, "llm.cancel", map[string]any{
		"request_id": "does-not-exist",
	})
	if rpcErr == nil {
		t.Fatal("expected error for unknown request_id")
	}
}

// TestStream_ContextOverflowReturnsError verifies that the daemon
// refuses a stream whose conversation exceeds the model's context
// window. Exercises the mapStreamError(ErrContextFull) branch.
func TestStream_ContextOverflowReturnsError(t *testing.T) {
	// Register a small-context model in the global pricing registry.
	llm.RegisterModel(llm.ModelInfo{ID: "tiny", ContextWindow: 10})
	t.Cleanup(func() { llm.UnregisterModel("tiny") })

	addr, _, cleanup := bringUpPipeline(t, &fakeStreamingProvider{
		name: "fake",
		events: []llm.StreamEvent{
			{Done: true},
		},
	})
	defer cleanup()

	// A long message that overflows the 10-token window (at
	// 4 chars/token, 100 chars > 10 tokens + 1000 reserve).
	big := strings.Repeat("a", 100)
	_, rpcErr := callRPC(t, addr, "llm.stream", map[string]any{
		"provider": "fake",
		"request": map[string]any{
			"model": "tiny",
			"messages": []map[string]any{
				{"role": "user", "content": big},
			},
		},
	})
	if rpcErr == nil {
		t.Fatal("expected error for context overflow")
	}
}

func TestStream_BrokerMountedAtEvents(t *testing.T) {
	addr, _, cleanup := bringUpPipeline(t)
	defer cleanup()

	resp, err := http.Get("http://" + addr + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}
}

// waitForEvents returns true once the recorder has seen an event
// with the given name.
func waitForEvents(t *testing.T, rec *streamEventRecorder, name string, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, ev := range rec.snapshot() {
			if ev.name == name {
				return true
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

// filterStreamEvents removes the SSE broker's "connected" handshake
// event so tests see only the events the manager published.
func filterStreamEvents(in []parsedStreamEvent) []parsedStreamEvent {
	out := in[:0]
	for _, e := range in {
		if e.name == "connected" {
			continue
		}
		out = append(out, e)
	}
	return out
}
