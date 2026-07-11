//go:build integration

// Package integration contains end-to-end tests that exercise
// multiple Condura subsystems together. The //go:build integration
// tag keeps these tests out of the regular `go test ./...` runs
// (where they would be slow and rely on subsystem setup) and gates
// them to the CI "Integration Tests" job, which runs:
//
//	go test -race -count=1 -timeout=300s -tags=integration ./condura-app/test/...
//
// (per .github/workflows/ci.yml and condura-ops/ci/workflows/ci.yml).
//
// The tests use real storage.Open (with real AES-256-GCM encryption),
// the real reach package (Manager + Store + Cipher), and the real
// sse broker. The Telegram Bot API is faked via httptest.NewServer
// so the tests run hermetically without network access. The IPC
// transport is exercised via httptest.NewServer wrapping its
// ServeHTTP method, so the CORS layer added in commit 8309ece
// has direct end-to-end coverage.
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/ipc"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/reach"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/sse"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/storage"
)

// testMasterKey is a 32-byte base64 key, identical to the one
// used in internal/storage/db_test.go. Reusing the same key keeps
// the integration-test fixtures byte-compatible with the unit-test
// fixtures, which makes debugging easier (a DB created by either
// suite is interchangeable).
const testMasterKey = "k6Qm1xJ4pYqZ8cV2nB3wD5rT7eH9uL0sA1bC2dE3fG4="

// newTestStorage opens a real storage.DB on a temp dir with the
// shared test master key. All migrations v1..v8 are applied at
// open. The DB is closed automatically when the test ends.
func newTestStorage(t *testing.T) *storage.DB {
	t.Helper()
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{
		Path:      filepath.Join(dir, "condura.db"),
		MasterKey: testMasterKey,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// fakeBotAPI is a minimal in-process Telegram Bot API server
// suitable for the channels end-to-end test. It implements just
// the methods reach.telegramChannel calls:
//
//	GET  /bot<token>/getMe      — returns ok=true with a fixed identity
//	POST /bot<token>/sendMessage — records the call for assertion
//	GET  /bot<token>/getUpdates  — returns whatever the test staged
//
// Other methods return 404, which matches Telegram's real behavior.
type fakeBotAPI struct {
	srv *httptest.Server

	mu         sync.Mutex
	sent       []map[string]string
	getUpdates []map[string]any
}

func newFakeBotAPI() *fakeBotAPI {
	f := &fakeBotAPI{}
	f.srv = httptest.NewServer(http.HandlerFunc(f.handle))
	return f
}

func (f *fakeBotAPI) URL() string { return f.srv.URL }
func (f *fakeBotAPI) Close()      { f.srv.Close() }

func (f *fakeBotAPI) handle(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/bot")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}
	switch parts[1] {
	case "getMe":
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true,
			"result": map[string]any{"username": "fake_bot", "id": 123456},
		})
	case "sendMessage":
		// Telegram's Bot API uses application/x-www-form-urlencoded
		// bodies, not JSON. The channel's apiCall sets the right
		// Content-Type, so we parse with r.PostForm (which honors
		// the URL-encoded form). Parsing as JSON silently fails
		// (returns an empty map), which is what the previous
		// version of this fake did — the chat_id was correctly
		// sent but the recorded map was empty.
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form: "+err.Error(), http.StatusBadRequest)
			return
		}
		f.mu.Lock()
		f.sent = append(f.sent, map[string]string{
			"chat_id": r.PostFormValue("chat_id"),
			"text":    r.PostFormValue("text"),
		})
		f.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":     true,
			"result": map[string]any{"message_id": 1},
		})
	case "getUpdates":
		f.mu.Lock()
		updates := f.getUpdates
		f.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":     true,
			"result": updates,
		})
	default:
		http.Error(w, "unknown method "+parts[1], http.StatusNotFound)
	}
}

// pushUpdate queues a fake update for the next getUpdates poll.
// The receive loop on the channel side will pick it up on its
// next 30-second long-poll tick.
func (f *fakeBotAPI) pushUpdate(u map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.getUpdates = append(f.getUpdates, u)
}

func (f *fakeBotAPI) sentCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sent)
}

// TestIntegration_TelegramConnectEndToEnd is the core test for
// the fix in commit 8309ece. It exercises the full
// Connect → encrypt token → save → receive loop → chat_id
// capture → SSE publish path against a fake Telegram Bot API.
// The cipher is the real storage.DB AES-256-GCM, so this test
// also catches any future regression where someone removes the
// encryption and stores the token as plaintext (the bug Bug #3
// in the channels audit).
func TestIntegration_TelegramConnectEndToEnd(t *testing.T) {
	db := newTestStorage(t)
	store, err := reach.NewStore(db.SQL())
	require.NoError(t, err)
	broker := sse.NewBroker()
	defer broker.Close()

	bot := newFakeBotAPI()
	defer bot.Close()

	mgr := reach.NewManager(store, db, broker)
	mgr.SetBotBaseURL(bot.URL())

	// Connect with a known test token. The fake Bot API returns
	// ok=true for getMe, so the channel flips to connected.
	status, err := mgr.Connect(context.Background(), "telegram", "test-token-12345")
	require.NoError(t, err)
	require.True(t, status.Connected, "channel must be connected after Connect")

	// The token must be in the store as ciphertext, NOT plaintext.
	// This is the regression guard for the "tokens stored in
	// memory only" bug fixed in commit 8309ece.
	ct, err := store.GetTokenCiphertext(context.Background(), "telegram")
	require.NoError(t, err)
	require.NotEmpty(t, ct, "token ciphertext must be written to the store")
	require.NotContains(t, ct, "test-token-12345",
		"token must be encrypted, not stored as plaintext")

	// Push a fake update. The receive loop long-polls
	// getUpdates, sees this on its next tick, and captures
	// the chat_id.
	bot.pushUpdate(map[string]any{
		"update_id": 100,
		"message": map[string]any{
			"chat": map[string]any{"id": 12345},
			"from": map[string]any{"username": "alice"},
			"text":    "hello bot",
		},
	})

	// Wait for chat_id to be captured via the receive loop.
	require.Eventually(t, func() bool {
		s, _ := store.List(context.Background())
		return len(s) > 0 && s[0].ChatID == "12345"
	}, 3*time.Second, 50*time.Millisecond,
		"chat_id should be captured via the receive loop")

	// The broker should have received a channel.message SSE event.
	sub := broker.Subscribe()
	select {
	case ev := <-sub.Events:
		require.Equal(t, "channel.message", ev.Name)
		data, ok := ev.Data.(map[string]any)
		require.True(t, ok, "event data should be a map")
		require.Equal(t, "telegram", data["channel"])
		require.Equal(t, "12345", data["chat_id"])
		require.Equal(t, "hello bot", data["text"])
	case <-time.After(2 * time.Second):
		t.Fatal("did not receive channel.message SSE event within 2s")
	}
}

// TestIntegration_TelegramRestoreAfterBoot verifies that a
// channel whose token was persisted encrypted in a previous
// daemon lifetime comes back online after Restore() without the
// user re-entering the token. This is the v0.1.0 user-experience
// promise: connect once, stay connected across restarts.
func TestIntegration_TelegramRestoreAfterBoot(t *testing.T) {
	db := newTestStorage(t)
	store, err := reach.NewStore(db.SQL())
	require.NoError(t, err)
	broker := sse.NewBroker()
	defer broker.Close()

	bot := newFakeBotAPI()
	defer bot.Close()

	// First "lifetime": connect. The encrypted token is in the
	// store when this returns.
	mgr1 := reach.NewManager(store, db, broker)
	mgr1.SetBotBaseURL(bot.URL())
	_, err = mgr1.Connect(context.Background(), "telegram", "restored-token-fake")
	require.NoError(t, err)
	firstStatus, err := mgr1.Status(context.Background(), "telegram")
	require.NoError(t, err)
	require.True(t, firstStatus.Connected)

	// Second "lifetime": fresh Manager, same store. Restore()
	// should bring the channel back online without the user doing
	// anything.
	mgr2 := reach.NewManager(store, db, broker)
	mgr2.SetBotBaseURL(bot.URL())
	require.NoError(t, mgr2.Restore(context.Background()))

	status, err := mgr2.Status(context.Background(), "telegram")
	require.NoError(t, err)
	require.True(t, status.Connected,
		"channel should be reconnected after Restore without re-entering the token")
}

// TestIntegration_TokenSendToChatID exercises the chat_id capture
// + Send happy path: after Restore, the user messages the bot,
// the receive loop captures chat_id, and Send routes to that
// chat_id without the user supplying one. This is the end-to-end
// guarantee that the bot can actually reply to the user.
func TestIntegration_TokenSendToChatID(t *testing.T) {
	db := newTestStorage(t)
	store, err := reach.NewStore(db.SQL())
	require.NoError(t, err)
	broker := sse.NewBroker()
	defer broker.Close()

	bot := newFakeBotAPI()
	defer bot.Close()

	mgr := reach.NewManager(store, db, broker)
	mgr.SetBotBaseURL(bot.URL())

	// Connect.
	_, err = mgr.Connect(context.Background(), "telegram", "send-test-token")
	require.NoError(t, err)

	// First message from user → capture chat_id.
	bot.pushUpdate(map[string]any{
		"update_id": 1,
		"message": map[string]any{
			"chat": map[string]any{"id": 99999},
			"from": map[string]any{"username": "bob"},
			"text":    "first contact",
		},
	})
	require.Eventually(t, func() bool {
		s, _ := store.List(context.Background())
		return len(s) > 0 && s[0].ChatID == "99999"
	}, 3*time.Second, 50*time.Millisecond, "chat_id must be captured")

	// Now Send with empty chat_id → should route to the captured
	// chat_id automatically.
	require.NoError(t, mgr.Send(context.Background(), "telegram", "", "reply"))
	require.Eventually(t, func() bool {
		return bot.sentCount() >= 1
	}, 2*time.Second, 50*time.Millisecond, "Send must reach the fake Bot API")

	// And the recorded call should have the right chat_id and text.
	bot.mu.Lock()
	defer bot.mu.Unlock()
	require.Len(t, bot.sent, 1)
	require.Equal(t, "99999", bot.sent[0]["chat_id"])
	require.Equal(t, "reply", bot.sent[0]["text"])
}

// TestIntegration_IPC_CORSPreflight verifies the OPTIONS preflight
// handling added in commit 8309ece. The Wails webview (loaded at
// wails://) sends OPTIONS before any cross-origin POST with a
// non-simple Content-Type or Authorization header. The transport
// must respond 204 + CORS headers or the webview blocks the
// request and the GUI surfaces "TypeError: Load failed".
func TestIntegration_IPC_CORSPreflight(t *testing.T) {
	s := ipc.NewServer()
	// Register a no-op method so the JSON-RPC routing has something
	// to register. The preflight is short-circuited before any
	// method dispatch happens, but having one registered matches
	// the production setup.
	s.Register("test.echo", func(ctx context.Context, _ json.RawMessage) (any, error) {
		return nil, nil
	})

	tr := &ipc.ServerTransport{S: s, Token: "test-token"}
	srv := httptest.NewServer(http.HandlerFunc(tr.ServeHTTP))
	defer srv.Close()

	req, err := http.NewRequest("OPTIONS", srv.URL+"/api", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "http://wails.localhost")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type,Authorization")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusNoContent, resp.StatusCode,
		"OPTIONS preflight must return 204 No Content")
	require.Equal(t, "http://wails.localhost", resp.Header.Get("Access-Control-Allow-Origin"),
		"Origin must be echoed so the browser sends credentials")
	require.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "POST")
	require.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "Authorization")
	require.Equal(t, "true", resp.Header.Get("Access-Control-Allow-Credentials"))
}

// TestIntegration_IPC_ChannelsRPC_CORS verifies that POST /api
// (the channels.connect JSON-RPC method) carries CORS response
// headers alongside the JSON-RPC payload. The wails:// webview
// needs CORS headers on EVERY response — not just OPTIONS
// preflight — to read the response body. A missing header here
// would surface as "TypeError: Load failed" on the GUI side.
func TestIntegration_IPC_ChannelsRPC_CORS(t *testing.T) {
	s := ipc.NewServer()
	s.Register("test.echo", func(ctx context.Context, params json.RawMessage) (any, error) {
		return map[string]any{"ok": true, "echo": params}, nil
	})

	tr := &ipc.ServerTransport{S: s, Token: "test-token"}
	srv := httptest.NewServer(http.HandlerFunc(tr.ServeHTTP))
	defer srv.Close()

	body, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "test.echo",
		"id":      1,
		"params":  map[string]any{"hello": "world"},
	})
	req, _ := http.NewRequest("POST", srv.URL+"/api", bytes.NewReader(body))
	req.Header.Set("Origin", "http://wails.localhost")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// CORS headers on the actual JSON-RPC response.
	require.Equal(t, "http://wails.localhost", resp.Header.Get("Access-Control-Allow-Origin"))
	require.Equal(t, "true", resp.Header.Get("Access-Control-Allow-Credentials"))

	// JSON-RPC payload is correctly decoded alongside the CORS
	// headers — the two are independent.
	var rpc struct {
		JSONRPC string         `json:"jsonrpc"`
		ID      int            `json:"id"`
		Result  map[string]any `json:"result"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&rpc))
	require.Equal(t, "2.0", rpc.JSONRPC)
	require.Equal(t, 1, rpc.ID)
	require.NotNil(t, rpc.Result["echo"])
}

// TestIntegration_StorageTokenEncryptionRoundTrip verifies that
// the ciphertext produced by storage.DB.EncryptStringWithAAD can
// be decrypted to the original plaintext. This is the round-trip
// the channels subsystem relies on; if it ever breaks, the
// "channel restored after reboot" promise breaks with it.
func TestIntegration_StorageTokenEncryptionRoundTrip(t *testing.T) {
	db := newTestStorage(t)

	plaintext := "telegram-bot-token-with-special-chars-/:_"
	aad := []byte("reach:telegram")

	ct, err := db.EncryptStringWithAAD(plaintext, aad)
	require.NoError(t, err)
	require.NotEqual(t, plaintext, ct, "ciphertext must differ from plaintext")
	require.NotContains(t, ct, plaintext, "ciphertext must not leak plaintext")

	// Decrypt via the same envelope API. The AAD travels inside
	// the envelope, so the caller does not need to remember it.
	decrypted, err := db.DecryptStringWithAAD(ct)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

// TestIntegration_IPC_RejectsMissingAuth verifies that with a
// bearer token configured, unauthenticated POSTs return 401
// (and do NOT receive CORS headers that would let a browser read
// the 401 body). CORS is applied to every response, but 401 must
// still be the unmitigated answer to an unauthenticated caller.
func TestIntegration_IPC_RejectsMissingAuth(t *testing.T) {
	s := ipc.NewServer()
	s.Register("test.echo", func(ctx context.Context, _ json.RawMessage) (any, error) {
		return nil, nil
	})

	tr := &ipc.ServerTransport{S: s, Token: "expected-token"}
	srv := httptest.NewServer(http.HandlerFunc(tr.ServeHTTP))
	defer srv.Close()

	body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "method": "test.echo", "id": 1})
	req, _ := http.NewRequest("POST", srv.URL+"/api", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Deliberately no Authorization header.

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode,
		"missing bearer must return 401 even with CORS configured")
}
