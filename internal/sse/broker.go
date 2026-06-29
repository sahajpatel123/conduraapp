// Package sse implements a Server-Sent Events (SSE) broker for the
// Synaptic daemon.
//
// The daemon pushes streaming events (LLM tokens, spend warnings,
// audit events, halt state changes) to connected clients over a
// long-lived HTTP response. Clients connect via EventSource on
// /events and receive a stream of "data:" frames.
//
// Architecture:
//
//   - Broker is the central hub. Every method that emits an event
//     (e.g. a streaming LLM token) goes through Broker.Publish.
//   - Each connected client holds a channel. Broker.Publish fans
//     the event out to every channel. A slow client cannot block
//     the publisher; the channel has a buffer and we drop events
//     if the buffer fills up.
//   - Authentication is enforced at HTTP-handler time (the auth
//     token is read from the request and verified before the
//     client is added to the broker).
package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/version"
)

// heartbeatInterval is how often the broker sends a comment line to
// keep idle connections alive. SSE proxies (and most browsers) close
// idle connections after ~30s; 15s is a safe middle ground.
const heartbeatInterval = 15 * time.Second

// Event is a single SSE frame. The wire format is
//
//	event: <name>
//	data: <json>
//	id: <id>
//	<blank line>
//
// which the EventSource browser API parses back into the same
// fields.
type Event struct {
	Name string      // event: <name>
	Data interface{} // JSON-encoded as the data: payload
	ID   string      // event: id (optional, used for client resume)
}

// Broker fans events out to connected clients. It is safe for
// concurrent use by many publishers and many subscribers.
type Broker struct {
	mu         sync.RWMutex
	clients    map[*Client]struct{}
	subs       map[*Subscription]struct{}
	closed     bool
	eventCount atomic.Uint64
}

// NewBroker returns a fresh Broker. Call Close to stop accepting
// new clients (existing ones are not closed — let them drain).
func NewBroker() *Broker {
	return &Broker{
		clients: make(map[*Client]struct{}),
		subs:    make(map[*Subscription]struct{}),
	}
}

// Client represents a single SSE subscriber. It writes Events to
// the underlying http.ResponseWriter as they arrive on its channel.
type Client struct {
	broker  *Broker
	id      string
	channel chan Event
	writer  http.ResponseWriter
	flusher http.Flusher
	done    chan struct{}
}

// Publish enqueues an event on every connected client and every
// in-process subscription. Slow clients have events dropped (the
// channel has a buffer; we don't block). In-process subscribers
// also have a buffer; if it's full, the event is dropped.
func (b *Broker) Publish(ev Event) {
	b.mu.RLock()
	closed := b.closed
	count := uint64(0)
	for c := range b.clients {
		select {
		case c.channel <- ev:
			count++
		default:
			// drop on the floor; slow client
		}
	}
	for sub := range b.subs {
		select {
		case sub.Events <- ev:
			count++
		default:
			// drop on the floor; slow subscriber
		}
	}
	b.mu.RUnlock()
	if !closed {
		b.eventCount.Add(count)
	}
}

// PublishJSON is a convenience wrapper that JSON-encodes the data
// before publishing.
func (b *Broker) PublishJSON(name string, data interface{}) {
	b.Publish(Event{Name: name, Data: data})
}

// Close stops the broker. Connected clients are NOT closed (they
// will close on their own when the response writer is finished).
func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
}

// Subscription is an in-process subscription to the broker. Unlike
// HTTP clients, subscribers receive events directly from the broker
// without going through an HTTP response writer. They are intended
// for internal consumers (e.g. the session package accumulating
// stream deltas without persisting them).
//
// Call Unsubscribe to release resources when done. Events are NOT
// dropped for subscribers (their channel has a buffer and the broker
// is the only writer).
type Subscription struct {
	id     string
	broker *Broker
	Events chan Event // buffered; receives all published events
	Done   chan struct{}
}

// ID returns the subscription's identifier. Useful for logging.
func (s *Subscription) ID() string { return s.id }

// Subscribe creates an in-process subscription. Events are sent
// to the returned channel. Call Unsubscribe when done.
//
// The subscription's channel has a buffer of 256 events; slow
// subscribers that don't keep up will block the broker. Internal
// subscribers are expected to drain promptly.
func (b *Broker) Subscribe() *Subscription {
	id := b.eventCount.Add(1)
	sub := &Subscription{
		id:     fmt.Sprintf("sub-%d", id),
		broker: b,
		Events: make(chan Event, 256),
		Done:   make(chan struct{}),
	}
	b.mu.Lock()
	b.subs[sub] = struct{}{}
	b.mu.Unlock()
	return sub
}

// Unsubscribe releases a subscription. After Unsubscribe returns,
// no further events will be delivered and the Done channel is
// closed.
func (b *Broker) Unsubscribe(sub *Subscription) {
	if sub == nil {
		return
	}
	b.mu.Lock()
	if _, ok := b.subs[sub]; ok {
		delete(b.subs, sub)
		close(sub.Done)
	}
	b.mu.Unlock()
}

// ClientCount returns the number of currently connected clients.
func (b *Broker) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// ServeHTTP upgrades the request to a long-lived SSE response.
// Caller is expected to have authenticated the request first
// (auth middleware should run before this handler).
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering
	w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:7475")

	client := &Client{
		broker:  b,
		id:      r.Header.Get("X-Client-Id"),
		channel: make(chan Event, 64),
		writer:  w,
		flusher: flusher,
		done:    make(chan struct{}),
	}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	b.clients[client] = struct{}{}
	b.mu.Unlock()

	// Send a "connected" event so the client knows the stream is up.
	_ = client.writeEvent(Event{
		Name: "connected",
		Data: map[string]any{
			"ts":      time.Now().Unix(),
			"version": version.Get().Version,
		},
	})

	// Heartbeat ticker — sends a comment frame every 15s so the
	// connection stays alive through proxies.
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			b.removeClient(client)
			return
		case <-client.done:
			b.removeClient(client)
			return
		case ev := <-client.channel:
			if err := client.writeEvent(ev); err != nil {
				b.removeClient(client)
				return
			}
		case <-ticker.C:
			// Comment lines start with ':' and are ignored by
			// EventSource; they keep the connection warm.
			_, _ = fmt.Fprintf(w, ": ping %d\n\n", time.Now().Unix())
			flusher.Flush()
		}
	}
}

func (b *Broker) removeClient(c *Client) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.clients[c]; !ok {
		return
	}
	delete(b.clients, c)
	close(c.done)
}

func (c *Client) writeEvent(ev Event) error {
	if ev.Name != "" {
		if _, err := fmt.Fprintf(c.writer, "event: %s\n", ev.Name); err != nil {
			return err
		}
	}
	if ev.ID != "" {
		if _, err := fmt.Fprintf(c.writer, "id: %s\n", ev.ID); err != nil {
			return err
		}
	}
	var payload []byte
	if ev.Data != nil {
		var err error
		payload, err = json.Marshal(ev.Data)
		if err != nil {
			return err
		}
	}
	if len(payload) > 0 {
		if _, err := fmt.Fprintf(c.writer, "data: %s\n\n", payload); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(c.writer, "data: \n\n"); err != nil {
			return err
		}
	}
	c.flusher.Flush()
	return nil
}
