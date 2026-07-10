package reach

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeTelegramHandlers controls the canned responses of a fake
// Telegram Bot API server. Tests construct one and pass it to
// newFakeTelegramServer; the server dispatches the configured
// responses to getMe / sendMessage / getUpdates requests.
type fakeTelegramHandlers struct {
	getMeOK     bool
	sendOK      bool
	updates     []fakeUpdate
	sendCapture *fakeSendCapture
	callsMu     sync.Mutex
}

type fakeUpdate struct {
	UpdateID int64
	ChatID   int64
	Sender   string
	Text     string
}

type fakeSendCapture struct {
	mu    sync.Mutex
	calls []fakeSendCall
}

type fakeSendCall struct {
	ChatID string
	Text   string
}

// newFakeTelegramServer spins up an httptest.Server that pretends
// to be api.telegram.org. The returned *http.Client routes to the
// test server; pass it to telegramChannel.setHTTPClient so the
// channel's apiCall hits the fake.
//
// handlers is taken by pointer because fakeTelegramHandlers carries
// a sync.Mutex — Go's vet warns (copylocks) when a struct with a
// mutex is passed by value, since that copies the lock state.
func newFakeTelegramServer(t *testing.T, h *fakeTelegramHandlers) *http.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/bot")
		// path is "TOKEN/method"
		parts := strings.SplitN(path, "/", 2)
		if len(parts) != 2 {
			http.Error(w, "bad path", http.StatusBadRequest)
			return
		}
		method := parts[1]
		switch method {
		case "getMe":
			h.callsMu.Lock()
			defer h.callsMu.Unlock()
			if !h.getMeOK {
				writeTelegramJSON(w, false, nil)
				return
			}
			writeTelegramJSON(w, true, map[string]any{
				"username": "test_bot",
			})
		case "sendMessage":
			h.callsMu.Lock()
			capture := h.sendCapture
			h.callsMu.Unlock()
			if capture != nil {
				if err := r.ParseForm(); err == nil {
					capture.mu.Lock()
					capture.calls = append(capture.calls, fakeSendCall{
						ChatID: r.PostFormValue("chat_id"),
						Text:   r.PostFormValue("text"),
					})
					capture.mu.Unlock()
				}
			}
			if !h.sendOK {
				writeTelegramJSON(w, false, nil)
				return
			}
			writeTelegramJSON(w, true, map[string]any{
				"message_id": 1,
			})
		case "getUpdates":
			h.callsMu.Lock()
			ups := h.updates
			h.callsMu.Unlock()
			writeTelegramUpdates(w, ups)
		default:
			http.Error(w, "unknown method "+method, http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	// Return a client whose Transport rewrites the host to the
	// test server. The Telegram channel always calls
	// https://api.telegram.org/bot<TOKEN>/<METHOD>; the RoundTripper
	// here swaps that for the test server URL while preserving
	// the path.
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &rewriteTransport{
			target: srv.URL,
			base:   http.DefaultTransport,
		},
	}
}

// rewriteTransport forwards every request to a target URL while
// preserving the request path. Used so a real Telegram-URL-formatted
// request can be routed to httptest.NewServer.
type rewriteTransport struct {
	target string
	base   http.RoundTripper
}

func (r *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite scheme + host, keep path and query.
	rewritten := req.Clone(req.Context())
	targetURL := strings.TrimRight(r.target, "/") + req.URL.Path
	if req.URL.RawQuery != "" {
		targetURL += "?" + req.URL.RawQuery
	}
	parsed, err := req.URL.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	rewritten.URL = parsed
	return r.base.RoundTrip(rewritten)
}

func writeTelegramJSON(w http.ResponseWriter, ok bool, result any) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{
		"ok": ok,
	}
	if result != nil {
		resp["result"] = result
	} else if !ok {
		resp["description"] = "fake error"
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// writeTelegramUpdates serializes fakeUpdate values in the exact
// JSON shape the long-poll decoder expects (lowercase tags:
// update_id, message.chat.id, message.from.username, message.text).
// Using the production wire shape here is what makes the test
// meaningful — it would catch a regression in the decoder struct
// tags.
func writeTelegramUpdates(w http.ResponseWriter, ups []fakeUpdate) {
	w.Header().Set("Content-Type", "application/json")
	type wireUpdate struct {
		UpdateID int64 `json:"update_id"`
		Message  struct {
			Chat struct {
				ID int64 `json:"id"`
			} `json:"chat"`
			From struct {
				Username string `json:"username"`
			} `json:"from"`
			Text string `json:"text"`
		} `json:"message"`
	}
	wire := make([]wireUpdate, len(ups))
	for i, u := range ups {
		wire[i].UpdateID = u.UpdateID
		wire[i].Message.Chat.ID = u.ChatID
		wire[i].Message.From.Username = u.Sender
		wire[i].Message.Text = u.Text
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":     true,
		"result": wire,
	})
}

// TestTelegramChannel_ConnectAgainstFakeServer exercises the real
// Connect path against a fake api.telegram.org, asserting that
// getMe is called and the channel flips to connected.
func TestTelegramChannel_ConnectAgainstFakeServer(t *testing.T) {
	tc := newTelegramChannel()
	tc.setHTTPClient(newFakeTelegramServer(t, &fakeTelegramHandlers{getMeOK: true}))

	if err := tc.Connect(context.Background(), "fake-token"); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	status, err := tc.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !status.Connected {
		t.Fatal("channel should be connected after successful Connect")
	}
}

// TestTelegramChannel_Connect_InvalidToken covers the rejection
// path: an invalid token (getMe=ok=false) must return an error
// and leave the channel disconnected.
func TestTelegramChannel_Connect_InvalidToken(t *testing.T) {
	tc := newTelegramChannel()
	tc.setHTTPClient(newFakeTelegramServer(t, &fakeTelegramHandlers{getMeOK: false}))

	if err := tc.Connect(context.Background(), "fake-token"); err == nil {
		t.Fatal("Connect with invalid token should fail")
	}
	status, _ := tc.Status(context.Background())
	if status.Connected {
		t.Fatal("channel should NOT be connected after failed Connect")
	}
}

// TestTelegramChannel_Send_HappyPath covers the Send path
// against a fake server, asserting that the chat_id and text
// flow through to the API call.
func TestTelegramChannel_Send_HappyPath(t *testing.T) {
	tc := newTelegramChannel()
	tc.token = "fake-token"
	tc.chatID = "12345"
	tc.connected = true // bypass Connect — focus on Send
	capture := &fakeSendCapture{}
	client := newFakeTelegramServer(t, &fakeTelegramHandlers{
		sendOK:      true,
		sendCapture: capture,
	})
	tc.setHTTPClient(client)

	if err := tc.Send(context.Background(), "", "hello world"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	capture.mu.Lock()
	defer capture.mu.Unlock()
	if len(capture.calls) != 1 {
		t.Fatalf("expected 1 send call, got %d", len(capture.calls))
	}
	if capture.calls[0].ChatID != "12345" {
		t.Fatalf("chatID: got %q", capture.calls[0].ChatID)
	}
	if capture.calls[0].Text != "hello world" {
		t.Fatalf("text: got %q", capture.calls[0].Text)
	}
}

// TestTelegramChannel_Send_EmptyChatID covers the new error path
// I added: Send with no chat_id AND no stored t.chatID must fail
// with a clear error message, instead of silently calling
// Telegram with chat_id="".
func TestTelegramChannel_Send_EmptyChatID(t *testing.T) {
	tc := newTelegramChannel()
	tc.connected = true
	// t.chatID is "" by default
	if err := tc.Send(context.Background(), "", "hi"); err == nil {
		t.Fatal("Send with empty chat_id should fail")
	} else if !strings.Contains(err.Error(), "chat_id unknown") {
		t.Fatalf("error message should mention chat_id unknown, got: %v", err)
	}
}

// TestTelegramChannel_Receive_CapturesMessages covers the
// long-poll path: pre-stage one fake update and verify it flows
// out of Receive as a Message.
func TestTelegramChannel_Receive_CapturesMessages(t *testing.T) {
	tc := newTelegramChannel()
	tc.token = "fake-token"
	tc.connected = true
	tc.setHTTPClient(newFakeTelegramServer(t, &fakeTelegramHandlers{
		updates: []fakeUpdate{
			{UpdateID: 100, ChatID: 12345, Sender: "alice", Text: "hello"},
		},
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	msgCh, err := tc.Receive(ctx)
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}

	select {
	case msg := <-msgCh:
		if msg.ChatID != "12345" {
			t.Fatalf("chatID: got %q", msg.ChatID)
		}
		if msg.Sender != "alice" {
			t.Fatalf("sender: got %q", msg.Sender)
		}
		if msg.Text != "hello" {
			t.Fatalf("text: got %q", msg.Text)
		}
		if msg.Channel != "telegram" {
			t.Fatalf("channel: got %q", msg.Channel)
		}
	case <-ctx.Done():
		t.Fatal("Receive did not deliver message within 3s")
	}
}

// TestTelegramChannel_Receive_RequiresConnection covers the
// negative path: Receive on a disconnected channel returns
// an error rather than silently starting a long-poll.
func TestTelegramChannel_Receive_RequiresConnection(t *testing.T) {
	tc := newTelegramChannel()
	// not connected
	if _, err := tc.Receive(context.Background()); err == nil {
		t.Fatal("Receive on disconnected channel should fail")
	}
}

// TestTelegramChannel_Disconnect_ClearsState covers the
// Disconnect path: token, chatID, connected flag all reset.
func TestTelegramChannel_Disconnect_ClearsState(t *testing.T) {
	tc := newTelegramChannel()
	tc.token = "fake"
	tc.chatID = "12345"
	tc.connected = true

	if err := tc.Disconnect(context.Background()); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if tc.token != "" || tc.chatID != "" || tc.connected {
		t.Fatalf("Disconnect should clear state, got token=%q chatID=%q connected=%v",
			tc.token, tc.chatID, tc.connected)
	}
}

// guard against silent no-op compilation
var _ = fmt.Sprintf
