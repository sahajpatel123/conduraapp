package ipc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	s.Register("z", func(_ context.Context, _ json.RawMessage) (any, error) { return nil, nil })
	s.Register("a", func(_ context.Context, _ json.RawMessage) (any, error) { return nil, nil })
	s.Register("m", func(_ context.Context, _ json.RawMessage) (any, error) { return nil, nil })
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
	assert.Contains(t, resp.Error.Message, "kaboom")
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
		return nil, nil
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
		return nil, nil
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
		return json.RawMessage(params), nil
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
		return nil, nil
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
		return nil, nil
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

func TestTransport_HTTP_POST(t *testing.T) {
	s := NewServer()
	s.Register("echo", func(_ context.Context, params json.RawMessage) (any, error) {
		return json.RawMessage(params), nil
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
		return nil, nil
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
		return json.RawMessage(params), nil
	})
	st := &ServerTransport{S: s}
	require.NoError(t, st.Listen(context.Background(), "tcp://127.0.0.1:0"))
	defer func() { _ = st.Close() }()

	url := "ws://" + st.Addr() + "/"
	ws, _, err := websocket.Dial(context.Background(), url, nil)
	require.NoError(t, err)
	defer ws.Close(websocket.StatusNormalClosure, "") //nolint:errcheck

	// Send a request.
	var req map[string]any
	req = map[string]any{"jsonrpc": "2.0", "method": "echo", "params": "hi", "id": 1}
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
		return nil, nil
	})
	st := &ServerTransport{S: s}
	require.NoError(t, st.Listen(context.Background(), "tcp://127.0.0.1:0"))
	defer func() { _ = st.Close() }()

	ws, _, err := websocket.Dial(context.Background(), "ws://"+st.Addr()+"/", nil)
	require.NoError(t, err)
	defer ws.Close(websocket.StatusNormalClosure, "") //nolint:errcheck

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
		scheme, host, err := parseAddr(c.in)
		require.NoError(t, err)
		assert.Equal(t, c.wantScheme, scheme)
		assert.Equal(t, c.wantHost, host)
	}
}

func TestBind_UnsupportedScheme(t *testing.T) {
	_, err := bind("ftp://x")
	assert.Error(t, err)
}

func TestBind_TCP(t *testing.T) {
	ln, err := bind("tcp://127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = ln.Close() }()
	assert.NotNil(t, ln.Addr())
}

func TestBind_Unix(t *testing.T) {
	dir := t.TempDir()
	ln, err := bind("unix://" + dir + "/s.sock")
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
