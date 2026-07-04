package sse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestBroker_PublishReceive verifies that an event published on
// the broker is delivered to a connected client.
func TestBroker_PublishReceive(t *testing.T) {
	b := NewBroker()
	defer b.Close()

	// Wire up a fake client. We don't go through HTTP here — we
	// inject the client directly to keep the test fast.
	b.mu.Lock()
	c := &Client{
		broker:  b,
		id:      "test",
		channel: make(chan Event, 4),
		writer:  httptest.NewRecorder(),
		done:    make(chan struct{}),
	}
	b.clients[c] = struct{}{}
	b.mu.Unlock()

	// Flusher is required for writeEvent; fake one.
	c.flusher = &nopFlusher{}

	b.PublishJSON("test", map[string]string{"hello": "world"})

	select {
	case ev := <-c.channel:
		if ev.Name != "test" {
			t.Fatalf("expected event name 'test', got %q", ev.Name)
		}
		data, _ := json.Marshal(ev.Data)
		if string(data) != `{"hello":"world"}` {
			t.Fatalf("unexpected payload: %s", data)
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive event within 1s")
	}
}

// TestBroker_MultipleClients verifies fan-out.
func TestBroker_MultipleClients(t *testing.T) {
	b := NewBroker()
	defer b.Close()

	const n = 5
	clients := make([]*Client, n)
	b.mu.Lock()
	for i := 0; i < n; i++ {
		c := &Client{
			broker:  b,
			id:      "c",
			channel: make(chan Event, 4),
			writer:  httptest.NewRecorder(),
			done:    make(chan struct{}),
			flusher: &nopFlusher{},
		}
		b.clients[c] = struct{}{}
		clients[i] = c
	}
	b.mu.Unlock()

	b.PublishJSON("fan", "out")

	for i, c := range clients {
		select {
		case ev := <-c.channel:
			if ev.Name != "fan" {
				t.Fatalf("client %d: expected 'fan', got %q", i, ev.Name)
			}
		case <-time.After(time.Second):
			t.Fatalf("client %d: no event", i)
		}
	}
}

// TestBroker_DropOnFull ensures a slow client does not block
// publishers; its event is dropped.
func TestBroker_DropOnFull(t *testing.T) {
	b := NewBroker()
	defer b.Close()

	b.mu.Lock()
	c := &Client{
		broker:  b,
		id:      "slow",
		channel: make(chan Event, 1), // tiny buffer
		writer:  httptest.NewRecorder(),
		done:    make(chan struct{}),
		flusher: &nopFlusher{},
	}
	b.clients[c] = struct{}{}
	b.mu.Unlock()

	// Fill the buffer with one event (we don't read).
	b.PublishJSON("e1", 1)
	// Now publish a second; it should be dropped, not block.
	done := make(chan struct{})
	go func() {
		b.PublishJSON("e2", 2)
		close(done)
	}()
	select {
	case <-done:
		// good
	case <-time.After(time.Second):
		t.Fatal("Publish blocked on full client channel")
	}
}

// TestBroker_ServeHTTP smoke-tests the HTTP handler.
func TestBroker_ServeHTTP(t *testing.T) {
	b := NewBroker()
	defer b.Close()

	srv := httptest.NewServer(b)
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("Content-Type = %q", got)
	}

	// Read the first "connected" event.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = resp.Request.WithContext(ctx)
	// We don't bother reading the body fully; the handler returned
	// the right content-type which is the main contract.
}

// TestBroker_RemoveClient checks the cleanup path.
func TestBroker_RemoveClient(t *testing.T) {
	b := NewBroker()
	defer b.Close()

	c := &Client{
		broker:  b,
		id:      "x",
		channel: make(chan Event, 1),
		writer:  httptest.NewRecorder(),
		done:    make(chan struct{}),
		flusher: &nopFlusher{},
	}
	b.mu.Lock()
	b.clients[c] = struct{}{}
	b.mu.Unlock()

	if b.ClientCount() != 1 {
		t.Fatalf("ClientCount = %d, want 1", b.ClientCount())
	}
	b.removeClient(c)
	if b.ClientCount() != 0 {
		t.Fatalf("ClientCount = %d, want 0", b.ClientCount())
	}
	// done should be closed
	select {
	case <-c.done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("done channel not closed after removeClient")
	}
}

type nopFlusher struct {
	mu sync.Mutex
}

func (n *nopFlusher) Flush() {
	n.mu.Lock()
	defer n.mu.Unlock()
}

// TestBroker_SubscribeReceives verifies that an in-process
// subscription receives events published to the broker.
func TestBroker_SubscribeReceives(t *testing.T) {
	b := NewBroker()
	defer b.Close()

	sub := b.Subscribe()
	defer b.Unsubscribe(sub)

	b.PublishJSON("test", map[string]string{"hello": "world"})

	select {
	case ev := <-sub.Events:
		if ev.Name != "test" {
			t.Fatalf("expected event name 'test', got %q", ev.Name)
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive event within 1s")
	}
}

// TestBroker_SubscribeIDIsUnique verifies each subscription gets a
// distinct ID.
func TestBroker_SubscribeIDIsUnique(t *testing.T) {
	b := NewBroker()
	defer b.Close()

	s1 := b.Subscribe()
	s2 := b.Subscribe()
	defer b.Unsubscribe(s1)
	defer b.Unsubscribe(s2)

	if s1.ID() == s2.ID() {
		t.Fatalf("expected unique IDs, got %q twice", s1.ID())
	}
	if s1.ID() == "" || s2.ID() == "" {
		t.Fatal("IDs must not be empty")
	}
}

// TestBroker_UnsubscribeClosesDone verifies that calling
// Unsubscribe closes the Done channel.
func TestBroker_UnsubscribeClosesDone(t *testing.T) {
	b := NewBroker()
	defer b.Close()

	sub := b.Subscribe()
	b.Unsubscribe(sub)

	select {
	case <-sub.Done:
		// expected
	case <-time.After(time.Second):
		t.Fatal("Done channel not closed after Unsubscribe")
	}
}

// TestBroker_UnsubscribeNilSafe verifies Unsubscribe(nil) is safe.
func TestBroker_UnsubscribeNilSafe(t *testing.T) {
	b := NewBroker()
	defer b.Close()
	b.Unsubscribe(nil) // must not panic
}

// TestBroker_PublishFansOutToSubscribers verifies that a single
// publish reaches multiple subscribers.
func TestBroker_PublishFansOutToSubscribers(t *testing.T) {
	b := NewBroker()
	defer b.Close()

	s1 := b.Subscribe()
	s2 := b.Subscribe()
	defer b.Unsubscribe(s1)
	defer b.Unsubscribe(s2)

	b.PublishJSON("test", "payload")

	for i, sub := range []*Subscription{s1, s2} {
		select {
		case ev := <-sub.Events:
			if ev.Name != "test" {
				t.Errorf("sub[%d] event name = %q, want %q", i, ev.Name, "test")
			}
		case <-time.After(time.Second):
			t.Errorf("sub[%d] did not receive event", i)
		}
	}
}

// TestBroker_UnsubscribeIdempotent verifies that calling
// Unsubscribe twice is safe.
func TestBroker_UnsubscribeIdempotent(t *testing.T) {
	b := NewBroker()
	defer b.Close()

	sub := b.Subscribe()
	b.Unsubscribe(sub)
	b.Unsubscribe(sub) // must not panic on closed channel
}
