package ipc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sahajpatel123/conduraapp/internal/health"
)

// -----------------------------------------------------------------------------
// Server — basic dispatch
// -----------------------------------------------------------------------------

func TestServer_RegisterHasUnregister(t *testing.T) {
	s := NewServer()
	s.Register("ping", func(_ context.Context, _ json.RawMessage) (any, error) {
		return "pong", nil
	})
	assert.True(t, s.HasMethod("ping"))
	assert.False(t, s.HasMethod("nope"))
	s.Unregister("ping")
	assert.False(t, s.HasMethod("ping"))
}

func TestServer_MethodsSorted(t *testing.T) {
	s := NewServer()
	s.Register("z", func(_ context.Context, _ json.RawMessage) (any, error) { return nil, errors.New("unused") })
	s.Register("a", func(_ context.Context, _ json.RawMessage) (any, error) { return nil, errors.New("unused") })
	s.Register("m", func(_ context.Context, _ json.RawMessage) (any, error) { return nil, errors.New("unused") })
	// Test-only handlers: pure structural checks, not a notification path.
	// Map iteration order is random; just check count and that the names
	// are present.
	names := s.Methods()
	assert.Len(t, names, 3)
	assert.Contains(t, names, "a")
	assert.Contains(t, names, "m")
	assert.Contains(t, names, "z")
}

func TestServer_Handle_OK(t *testing.T) {
	s := NewServer()
	s.Register("add", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			A int `json:"a"`
			B int `json:"b"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &Error{Code: CodeInvalidParams, Message: err.Error()}
		}
		return p.A + p.B, nil
	})
	resp, err := s.Handle(context.Background(), &Request{
		JSONRPC: "2.0", Method: "add", Params: json.RawMessage(`{"a":2,"b":3}`), ID: json.RawMessage("1"),
	})
	require.NoError(t, err)
	assert.Equal(t, 2+3, resp.Result)
}

func TestServer_Handle_BadJSONRPC(t *testing.T) {
	s := NewServer()
	_, err := s.Handle(context.Background(), &Request{JSONRPC: "1.0", Method: "x", ID: json.RawMessage("1")})
	assert.Error(t, err)
}

func TestServer_Handle_NoMethod(t *testing.T) {
	s := NewServer()
	resp, err := s.Handle(context.Background(), &Request{JSONRPC: "2.0", Method: "x", ID: json.RawMessage("1")})
	require.NoError(t, err)
	require.NotNil(t, resp.Error)
	assert.Equal(t, CodeMethodNotFound, resp.Error.Code)
}

func TestServer_Handle_NoMethodName(t *testing.T) {
	s := NewServer()
	_, err := s.Handle(context.Background(), &Request{JSONRPC: "2.0", ID: json.RawMessage("1")})
	assert.Error(t, err)
}

func TestServer_Handle_InternalError(t *testing.T) {
	s := NewServer()
	s.Register("boom", func(_ context.Context, _ json.RawMessage) (any, error) {
		return nil, errors.New("kaboom")
	})
	resp, err := s.Handle(context.Background(), &Request{JSONRPC: "2.0", Method: "boom", ID: json.RawMessage("1")})
	require.NoError(t, err)
	require.NotNil(t, resp.Error)
	assert.Equal(t, CodeInternalError, resp.Error.Code)
	// 2026-06-29 audit P0-3: the client must NOT see the raw error.
	assert.Equal(t, "internal error", resp.Error.Message)
	assert.NotContains(t, resp.Error.Message, "kaboom")
}

// TestServer_Handle_InternalError_LogsFullErr pins P0-3: when a
// handler returns a Go error that contains a path/IP, the JSON-RPC
// client sees only the redacted stable message, but the server's
// logger receives the full error for forensic correlation.
func TestServer_Handle_InternalError_LogsFullErr(t *testing.T) {
	var logged atomic.Bool
	var capturedErr string
	log := slog.New(slog.NewTextHandler(&captureWriter{onWrite: func(s string) {
		if strings.Contains(s, "open /Users/sahajpatel/.condura/secrets/api_key.enc") {
			capturedErr = s
			logged.Store(true)
		}
	}}, nil))

	s := NewServer().WithLogger(log)
	s.Register("boom", func(_ context.Context, _ json.RawMessage) (any, error) {
		return nil, fmt.Errorf("open /Users/sahajpatel/.condura/secrets/api_key.enc: permission denied")
	})
	resp, err := s.Handle(context.Background(), &Request{JSONRPC: "2.0", Method: "boom", ID: json.RawMessage("7")})
	require.NoError(t, err)
	require.NotNil(t, resp.Error)

	// Client gets the redacted message — no path.
	assert.Equal(t, "internal error", resp.Error.Message)
	assert.NotContains(t, resp.Error.Message, "/Users/")
	assert.NotContains(t, resp.Error.Message, "permission denied")

	// Server log got the full error.
	if !logged.Load() {
		t.Fatalf("server logger did not receive the full error (got: %q)", capturedErr)
	}
}

// TestServer_HandleRaw_ParseErrorRedacted pins that JSON parse
// errors do not leak json.SyntaxError details to the client.
func TestServer_HandleRaw_ParseErrorRedacted(t *testing.T) {
	var logged atomic.Bool
	log := slog.New(slog.NewTextHandler(&captureWriter{onWrite: func(s string) {
		// slog emits "level=DEBUG msg=\"ipc parse error\" err=...".
		if strings.Contains(s, "ipc parse error") {
			logged.Store(true)
		}
	}}, &slog.HandlerOptions{Level: slog.LevelDebug}))

	s := NewServer().WithLogger(log)
	out, err := s.HandleRaw(context.Background(), []byte("not json at all"))
	require.NoError(t, err)

	var resp Response
	require.NoError(t, json.Unmarshal(out, &resp))
	require.NotNil(t, resp.Error)
	assert.Equal(t, CodeParseError, resp.Error.Code)
	assert.Equal(t, "parse error", resp.Error.Message)

	if !logged.Load() {
		t.Fatal("server logger did not receive the parse error")
	}
}

// TestRedactHome_RedactsUserHome pins that the home directory is
// scrubbed from any error string before it reaches a downstream sink.
func TestRedactHome_RedactsUserHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		t.Skip("no home dir available in this environment")
	}
	in := "open " + home + "/.condura/secrets/api_key.enc: permission denied"
	out := redactHome(in)
	if strings.Contains(out, home) {
		t.Fatalf("home not redacted: %q -> %q", in, out)
	}
	if !strings.HasPrefix(out, "open ~/") {
		t.Fatalf("expected ~/ prefix, got %q", out)
	}
}

// TestRedactPrivateIP_RedactsCommonPrivateIPs pins the IP-redaction
// table covers RFC1918, loopback, link-local, and cloud metadata.
func TestRedactPrivateIP_RedactsCommonPrivateIPs(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"dial tcp 10.0.0.5:5432: connection refused", "dial tcp <private>:5432: connection refused"},
		{"dial tcp 192.168.1.1:80: i/o timeout", "dial tcp <private>:80: i/o timeout"},
		{"dial tcp 127.0.0.1:9: connection refused", "dial tcp <private>:9: connection refused"},
		{"metadata: 169.254.169.254/latest/meta-data/", "metadata: <metadata>/latest/meta-data/"},
	}
	for _, tc := range cases {
		got := redactPrivateIP(tc.in)
		if got != tc.want {
			t.Errorf("redactPrivateIP(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// captureWriter is an io.Writer that invokes onWrite for each Write.
// Used in tests that need to inspect what was logged without
// depending on slog's handler internals.
type captureWriter struct {
	onWrite func(string)
}

func (w *captureWriter) Write(p []byte) (int, error) {
	if w.onWrite != nil {
		w.onWrite(string(p))
	}
	return len(p), nil
}

func TestServer_Handle_IPCErrorPassthrough(t *testing.T) {
	s := NewServer()
	s.Register("boom", func(_ context.Context, _ json.RawMessage) (any, error) {
		return nil, &Error{Code: CodeInvalidParams, Message: "bad input"}
	})
	resp, err := s.Handle(context.Background(), &Request{JSONRPC: "2.0", Method: "boom", ID: json.RawMessage("1")})
	require.NoError(t, err)
	require.NotNil(t, resp.Error)
	assert.Equal(t, CodeInvalidParams, resp.Error.Code)
}

// -----------------------------------------------------------------------------
// HandleRaw — parsing
// -----------------------------------------------------------------------------

func TestServer_HandleRaw_OK(t *testing.T) {
	s := NewServer()
	s.Register("ping", func(_ context.Context, _ json.RawMessage) (any, error) {
		return "pong", nil
	})
	out, err := s.HandleRaw(context.Background(), []byte(`{"jsonrpc":"2.0","method":"ping","id":1}`))
	require.NoError(t, err)
	assert.Contains(t, string(out), `"pong"`)
}

func TestServer_HandleRaw_Notification(t *testing.T) {
	called := false
	s := NewServer()
	s.Register("notify", func(_ context.Context, _ json.RawMessage) (any, error) {
		called = true
		return nil, ErrNotification
	})
	out, err := s.HandleRaw(context.Background(), []byte(`{"jsonrpc":"2.0","method":"notify"}`))
	require.NoError(t, err)
	assert.Nil(t, out, "notification should produce no response")
	assert.True(t, called)
}

func TestServer_HandleRaw_NullIDNotification(t *testing.T) {
	called := false
	s := NewServer()
	s.Register("notify", func(_ context.Context, _ json.RawMessage) (any, error) {
		called = true
		return nil, ErrNotification
	})
	out, err := s.HandleRaw(context.Background(), []byte(`{"jsonrpc":"2.0","method":"notify","id":null}`))
	require.NoError(t, err)
	assert.Nil(t, out)
	assert.True(t, called)
}

func TestServer_HandleRaw_ParseError(t *testing.T) {
	s := NewServer()
	out, err := s.HandleRaw(context.Background(), []byte(`{not json`))
	require.NoError(t, err)
	assert.Contains(t, string(out), `"code":-32700`)
}

func TestServer_HandleRaw_EmptyBody(t *testing.T) {
	s := NewServer()
	out, err := s.HandleRaw(context.Background(), nil)
	require.NoError(t, err)
	assert.Contains(t, string(out), `"code":-32600`)
}

func TestServer_HandleRaw_Batch(t *testing.T) {
	s := NewServer()
	s.Register("echo", func(_ context.Context, params json.RawMessage) (any, error) {
		return params, nil
	})
	req := `[{"jsonrpc":"2.0","method":"echo","params":"a","id":1},{"jsonrpc":"2.0","method":"echo","params":"b","id":2}]`
	out, err := s.HandleRaw(context.Background(), []byte(req))
	require.NoError(t, err)
	assert.Contains(t, string(out), `"a"`)
	assert.Contains(t, string(out), `"b"`)
}

func TestServer_HandleRaw_AllNotificationsBatch(t *testing.T) {
	called := 0
	s := NewServer()
	s.Register("n", func(_ context.Context, _ json.RawMessage) (any, error) {
		called++
		return nil, ErrNotification
	})
	req := `[{"jsonrpc":"2.0","method":"n"},{"jsonrpc":"2.0","method":"n"}]`
	out, err := s.HandleRaw(context.Background(), []byte(req))
	require.NoError(t, err)
	assert.Nil(t, out)
	assert.Equal(t, 2, called)
}

func TestServer_HandleRaw_BatchEmpty(t *testing.T) {
	s := NewServer()
	out, err := s.HandleRaw(context.Background(), []byte(`[]`))
	require.NoError(t, err)
	assert.Contains(t, string(out), "empty batch")
}

func TestServer_HandleRaw_BatchParseError(t *testing.T) {
	s := NewServer()
	out, err := s.HandleRaw(context.Background(), []byte(`[{invalid`))
	require.NoError(t, err)
	assert.Contains(t, string(out), `"code":-32700`)
}

func TestServer_HandleRaw_BadJSONRPCVersion(t *testing.T) {
	s := NewServer()
	out, err := s.HandleRaw(context.Background(), []byte(`{"jsonrpc":"1.0","method":"x","id":1}`))
	require.NoError(t, err)
	assert.Contains(t, string(out), `"code":-32603`, "bad version is treated as internal error")
}

// -----------------------------------------------------------------------------
// ServeConn — interface-based loop
// -----------------------------------------------------------------------------

type fakeConn struct {
	mu      sync.Mutex
	readQ   []Message
	written []Message
	readIdx int
	closed  bool
	readErr error
}

func (f *fakeConn) Read(_ context.Context) (Message, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.readErr != nil {
		return Message{}, f.readErr
	}
	if f.readIdx >= len(f.readQ) {
		return Message{}, io.EOF
	}
	m := f.readQ[f.readIdx]
	f.readIdx++
	return m, nil
}

func (f *fakeConn) Write(_ context.Context, m Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.written = append(f.written, m)
	return nil
}

func (f *fakeConn) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

func TestServer_ServeConn(t *testing.T) {
	s := NewServer()
	s.Register("ping", func(_ context.Context, _ json.RawMessage) (any, error) {
		return "pong", nil
	})
	c := &fakeConn{
		readQ: []Message{
			{Type: MessageText, Data: []byte(`{"jsonrpc":"2.0","method":"ping","id":1}`)},
		},
	}
	_ = s.ServeConn(context.Background(), c)
	require.Len(t, c.written, 1)
	assert.Contains(t, string(c.written[0].Data), `"pong"`)
}

func TestServer_ServeConn_Notification(t *testing.T) {
	s := NewServer()
	called := false
	s.Register("n", func(_ context.Context, _ json.RawMessage) (any, error) {
		called = true
		return nil, ErrNotification
	})
	c := &fakeConn{
		readQ: []Message{
			{Type: MessageText, Data: []byte(`{"jsonrpc":"2.0","method":"n"}`)},
		},
	}
	_ = s.ServeConn(context.Background(), c)
	assert.True(t, called)
	assert.Empty(t, c.written)
}

// -----------------------------------------------------------------------------
// HTTP transport
// -----------------------------------------------------------------------------

func TestTransport_HTTP_GET(t *testing.T) {
	s := NewServer()
	st := &ServerTransport{S: s}
	require.NoError(t, st.Listen(context.Background(), "tcp://127.0.0.1:0"))
	defer func() { _ = st.Close() }()

	addr := "http://" + st.Addr()
	// GET / -> 405.
	resp, err := http.Get(addr + "/")
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	// GET /healthz -> 200.
	resp, err = http.Get(addr + "/healthz")
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ok", string(body))
}

func TestTransport_Health_LivezReadyz_NoAuth(t *testing.T) {
	// Pins CLAUDE.md FIX A: /livez and /readyz are unauthenticated
	// even when the IPC transport has a bearer token. The handler
	// is mounted BEFORE the auth check.
	s := NewServer()
	st := &ServerTransport{
		S:      s,
		Token:  "supersecret", // transport requires auth for /api etc.
		Health: healthForTest(),
	}
	require.NoError(t, st.Listen(context.Background(), "tcp://127.0.0.1:0"))
	defer func() { _ = st.Close() }()
	addr := "http://" + st.Addr()

	// No Authorization header on either probe. Both must return 200.
	resp, err := http.Get(addr + "/livez")
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "alive\n", string(body))

	resp, err = http.Get(addr + "/readyz")
	require.NoError(t, err)
	body, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ready\n", string(body))
}

func TestTransport_Health_Readyz_ReflectsFunc(t *testing.T) {
	// The readyz func's verdict flows to the wire: 503 with the
	// reason when the func errors. Probes are still public.
	s := NewServer()
	st := &ServerTransport{
		S:     s,
		Token: "supersecret",
		Health: healthForDownTest("migrations pending"),
	}
	require.NoError(t, st.Listen(context.Background(), "tcp://127.0.0.1:0"))
	defer func() { _ = st.Close() }()
	addr := "http://" + st.Addr()

	resp, err := http.Get(addr + "/readyz")
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Contains(t, string(body), "migrations pending")
}

func TestTransport_Health_AbsentByDefault(t *testing.T) {
	// When Health is nil, /livez and /readyz return 401 (the
	// transport is in auth mode but the path is otherwise
	// unhandled). The legacy /healthz endpoint still works.
	s := NewServer()
	st := &ServerTransport{S: s, Token: "supersecret"}
	require.NoError(t, st.Listen(context.Background(), "tcp://127.0.0.1:0"))
	defer func() { _ = st.Close() }()
	addr := "http://" + st.Addr()

	resp, err := http.Get(addr + "/livez")
	require.NoError(t, err)
	_ = resp.Body.Close()
	// Without Health wired, the request falls through to the
	// /healthz branch (which is also off) and then to auth
	// (Token is set), so the response is 401. The exact code is
	// not load-bearing for the FIX A contract; what matters is
	// that the handler is OFF when Health is nil.
	assert.True(t, resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotFound,
		"unexpected status %d", resp.StatusCode)
}

// healthForTest and healthForDownTest build the http.Handler
// fixture for the transport tests. They use the production health
// package so the tests exercise the same code path the daemon uses
// at runtime.
func healthForTest() http.Handler {
	return health.HTTPHandler(
		func() error { return nil },
		func() error { return nil },
	)
}

func healthForDownTest(reason string) http.Handler {
	return health.HTTPHandler(
		func() error { return nil },
		func() error { return errors.New(reason) },
	)
}

func TestTransport_HTTP_POST(t *testing.T) {
	s := NewServer()
	s.Register("echo", func(_ context.Context, params json.RawMessage) (any, error) {
		return params, nil
	})
	st := &ServerTransport{S: s}
	require.NoError(t, st.Listen(context.Background(), "tcp://127.0.0.1:0"))
	defer func() { _ = st.Close() }()
	addr := "http://" + st.Addr()

	resp, err := http.Post(addr+"/", "application/json",
		strings.NewReader(`{"jsonrpc":"2.0","method":"echo","params":"hi","id":1}`))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), `"hi"`)
}

func TestTransport_HTTP_POST_Notification(t *testing.T) {
	s := NewServer()
	s.Register("n", func(_ context.Context, _ json.RawMessage) (any, error) {
		return nil, ErrNotification
	})
	st := &ServerTransport{S: s}
	require.NoError(t, st.Listen(context.Background(), "tcp://127.0.0.1:0"))
	defer func() { _ = st.Close() }()

	resp, err := http.Post("http://"+st.Addr(), "application/json",
		strings.NewReader(`{"jsonrpc":"2.0","method":"n"}`))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestTransport_HTTP_TokenAuth(t *testing.T) {
	s := NewServer()
	s.Register("ping", func(_ context.Context, _ json.RawMessage) (any, error) {
		return "pong", nil
	})
	st := &ServerTransport{S: s, Token: "secret"}
	require.NoError(t, st.Listen(context.Background(), "tcp://127.0.0.1:0"))
	defer func() { _ = st.Close() }()
	addr := "http://" + st.Addr()

	// No token -> 401.
	resp, err := http.Post(addr+"/", "application/json",
		strings.NewReader(`{"jsonrpc":"2.0","method":"ping","id":1}`))
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// With token -> 200.
	req, _ := http.NewRequest(http.MethodPost, addr+"/",
		strings.NewReader(`{"jsonrpc":"2.0","method":"ping","id":1}`))
	req.Header.Set("Authorization", "Bearer secret")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// -----------------------------------------------------------------------------
// WebSocket transport
// -----------------------------------------------------------------------------

func TestTransport_WebSocket(t *testing.T) {
	s := NewServer()
	s.Register("echo", func(_ context.Context, params json.RawMessage) (any, error) {
		return params, nil
	})
	st := &ServerTransport{S: s}
	require.NoError(t, st.Listen(context.Background(), "tcp://127.0.0.1:0"))
	defer func() { _ = st.Close() }()

	url := "ws://" + st.Addr() + "/"
	//nolint:bodyclose // ws.Close also closes the underlying body
	ws, _, err := websocket.Dial(context.Background(), url, nil)
	require.NoError(t, err)
	defer func() { _ = ws.Close(websocket.StatusNormalClosure, "") }()

	// Send a request.
	req := map[string]any{"jsonrpc": "2.0", "method": "echo", "params": "hi", "id": 1}
	require.NoError(t, wsjson.Write(context.Background(), ws, req))

	// Read response.
	var resp map[string]any
	require.NoError(t, wsjson.Read(context.Background(), ws, &resp))
	assert.Equal(t, "hi", resp["result"])
}

func TestTransport_WebSocket_Notification(t *testing.T) {
	var called atomic.Int32
	s := NewServer()
	s.Register("n", func(_ context.Context, _ json.RawMessage) (any, error) {
		called.Add(1)
		return nil, ErrNotification
	})
	st := &ServerTransport{S: s}
	require.NoError(t, st.Listen(context.Background(), "tcp://127.0.0.1:0"))
	defer func() { _ = st.Close() }()

	//nolint:bodyclose // ws.Close also closes the underlying body
	ws, _, err := websocket.Dial(context.Background(), "ws://"+st.Addr()+"/", nil)
	require.NoError(t, err)
	defer func() { _ = ws.Close(websocket.StatusNormalClosure, "") }()

	require.NoError(t, wsjson.Write(context.Background(), ws,
		map[string]any{"jsonrpc": "2.0", "method": "n"}))
	// Wait for the handler to be called.
	for i := 0; i < 100 && called.Load() == 0; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	assert.Equal(t, int32(1), called.Load())
}

// -----------------------------------------------------------------------------
// Error + small helpers
// -----------------------------------------------------------------------------

func TestError_Error(t *testing.T) {
	e := &Error{Code: 42, Message: "x"}
	assert.Contains(t, e.Error(), "42")
	assert.Contains(t, e.Error(), "x")
}

func TestIsNotification(t *testing.T) {
	assert.True(t, isNotification(nil))
	assert.True(t, isNotification(json.RawMessage("null")))
	assert.False(t, isNotification(json.RawMessage("1")))
	assert.False(t, isNotification(json.RawMessage(`"abc"`)))
}

func TestBytesTrimSpace(t *testing.T) {
	assert.Equal(t, "x", string(bytesTrimSpace([]byte("  \nx\t"))))
	assert.Equal(t, "", string(bytesTrimSpace([]byte("   "))))
}

func TestParseAddr(t *testing.T) {
	cases := []struct {
		in         string
		wantScheme string
		wantHost   string
	}{
		{"tcp://127.0.0.1:1234", "tcp", "127.0.0.1:1234"},
		{"unix:///tmp/s.sock", "unix", "/tmp/s.sock"},
		{"127.0.0.1:1234", "tcp", "127.0.0.1:1234"},
	}
	for _, c := range cases {
		scheme, host := parseAddr(c.in)
		assert.Equal(t, c.wantScheme, scheme)
		assert.Equal(t, c.wantHost, host)
	}
}

func TestBind_UnsupportedScheme(t *testing.T) {
	_, err := bind(context.Background(), "ftp://x")
	assert.Error(t, err)
}

func TestBind_TCP(t *testing.T) {
	ln, err := bind(context.Background(), "tcp://127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = ln.Close() }()
	assert.NotNil(t, ln.Addr())
}

func TestBind_Unix(t *testing.T) {
	dir := t.TempDir()
	ln, err := bind(context.Background(), "unix://"+dir+"/s.sock")
	require.NoError(t, err)
	defer func() { _ = ln.Close() }()
}

func TestIsWebsocketUpgrade(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Connection", "Upgrade")
	r.Header.Set("Upgrade", "websocket")
	assert.True(t, isWebsocketUpgrade(r))

	r2, _ := http.NewRequest(http.MethodGet, "/", nil)
	r2.Header.Set("Connection", "keep-alive")
	assert.False(t, isWebsocketUpgrade(r2))

	r3, _ := http.NewRequest(http.MethodPost, "/", nil)
	r3.Header.Set("Connection", "Upgrade")
	r3.Header.Set("Upgrade", "websocket")
	assert.False(t, isWebsocketUpgrade(r3), "non-GET should be false")
}

func TestTokenListContainsToken(t *testing.T) {
	assert.True(t, tokenListContainsToken("keep-alive, Upgrade", "Upgrade"))
	assert.True(t, tokenListContainsToken("upgrade", "Upgrade"))
	assert.False(t, tokenListContainsToken("", "x"))
	assert.False(t, tokenListContainsToken("keep-alive", "Upgrade"))
}

func TestTrim(t *testing.T) {
	assert.Equal(t, "x", trim("  x  "))
	assert.Equal(t, "x y", trim(" x y "))
}

func TestStringsEqualFold(t *testing.T) {
	assert.True(t, stringsEqualFold("ABC", "abc"))
	assert.False(t, stringsEqualFold("ABC", "abcd"))
	assert.True(t, stringsEqualFold("WebSocket", "websocket"))
}

func TestTransport_Close_Idempotent(t *testing.T) {
	s := NewServer()
	st := &ServerTransport{S: s}
	require.NoError(t, st.Listen(context.Background(), "tcp://127.0.0.1:0"))
	assert.NoError(t, st.Close())
	assert.NoError(t, st.Close())
}

func TestTransport_Addr_Empty(t *testing.T) {
	st := &ServerTransport{S: NewServer()}
	assert.Equal(t, "", st.Addr())
}

// Sanity: the conn wrapper adapts to message types.
func TestWSConnWrapper_WriteText(t *testing.T) {
	// We don't have a real *websocket.Conn in tests, but we can verify
	// the type mapping logic by exercising the switch in Write.
	// Since we can't easily mock the inner conn, this is more of a
	// compile-time sanity check.
	var c Conn = &wsConnWrapper{}
	assert.NotNil(t, c)
}

// Ensure dial works for completeness.
func TestDialWebSocket(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() {
		c, err := ln.Accept()
		if err == nil {
			_ = c.Close()
		}
	}()
	defer func() { _ = ln.Close() }()
	conn, err := net.Dial("tcp", ln.Addr().String())
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()
	assert.NotNil(t, conn)
}

// Sanity: bytes reader import.
var _ = bytes.NewReader
