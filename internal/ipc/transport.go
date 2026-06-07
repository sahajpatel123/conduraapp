package ipc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// Conn is the abstract connection interface used by ServeConn. It is
// satisfied by *websocket.Conn; we keep an interface so tests can inject
// fakes.
type Conn interface {
	Read(ctx context.Context) (Message, error)
	Write(ctx context.Context, m Message) error
	Close() error
}

// Message is one frame sent over the wire.
type Message struct {
	Type MessageType
	Data []byte
}

// MessageType mirrors websocket.MessageType without leaking the dep.
type MessageType int

const (
	MessageText   MessageType = 1
	MessageBinary MessageType = 2
)

// -----------------------------------------------------------------------------
// Server transport
// -----------------------------------------------------------------------------

// Server bundles a method registry with a network listener.
//
// Listen addrs: pass "tcp://127.0.0.1:0" to bind a random TCP port, or
// "unix:///tmp/synaptic.sock" for a Unix socket.
type ServerTransport struct {
	S     *Server
	Token string // optional bearer token; if non-empty, clients must send it

	mu        sync.Mutex
	closed    bool
	listeners []net.Listener
	srv       *http.Server
}

// Addr returns the bound address of the first listener. Useful for
// clients that connect to "tcp://127.0.0.1:0".
func (t *ServerTransport) Addr() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.listeners) == 0 {
		return ""
	}
	return t.listeners[0].Addr().String()
}

// Listen binds a single address and starts serving.
func (t *ServerTransport) Listen(ctx context.Context, addr string) error {
	ln, err := bind(addr)
	if err != nil {
		return err
	}
	t.mu.Lock()
	t.listeners = append(t.listeners, ln)
	t.mu.Unlock()
	go t.serveListener(ctx, ln)
	return nil
}

// serveListener runs the HTTP+WebSocket handler on one listener.
func (t *ServerTransport) serveListener(ctx context.Context, ln net.Listener) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", t.handleHTTP)
	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	t.mu.Lock()
	t.srv = srv
	t.mu.Unlock()
	_ = srv.Serve(ln)
}

// handleHTTP dispatches HTTP requests:
//   - GET /healthz -> 200 OK
//   - POST / with JSON-RPC body -> response
//   - Upgrade to WebSocket on /ws
func (t *ServerTransport) handleHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/healthz" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return
	}
	// Auth check.
	if t.Token != "" {
		auth := r.Header.Get("Authorization")
		want := "Bearer " + t.Token
		if auth != want {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// WebSocket upgrade.
	if isWebsocketUpgrade(r) {
		t.serveWebSocket(w, r)
		return
	}

	// Plain HTTP JSON-RPC.
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer func() { _ = r.Body.Close() }()
	body := make([]byte, 0, 1024)
	buf := make([]byte, 1024)
	for {
		n, err := r.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	ctx := r.Context()
	out, err := t.S.HandleRaw(ctx, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if out == nil {
		// Notification: 204 No Content.
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// serveWebSocket upgrades the connection and runs the read/write loop.
func (t *ServerTransport) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// We deliberately do not check Origin in dev; the IPC server is
		// bound to localhost. A production deployment should add a
		// stricter check (Phase 5: signed auth tokens).
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer c.Close(websocket.StatusNormalClosure, "bye") //nolint:errcheck
	for {
		mt, data, err := c.Read(ctx)
		if err != nil {
			return
		}
		if mt != websocket.MessageText {
			continue
		}
		out, herr := t.S.HandleRaw(ctx, data)
		if herr != nil {
			_ = c.Write(ctx, websocket.MessageText, []byte(herr.Error()))
			continue
		}
		if out == nil {
			continue // notification
		}
		_ = c.Write(ctx, websocket.MessageText, out)
	}
}

// wsConnWrapper adapts *websocket.Conn to the Conn interface for
// ServeConn (kept for tests).
type wsConnWrapper struct {
	conn *websocket.Conn
}

func (w *wsConnWrapper) Read(ctx context.Context) (Message, error) {
	mt, data, err := w.conn.Read(ctx)
	if err != nil {
		return Message{}, err
	}
	return Message{Type: MessageType(mt), Data: data}, nil
}

func (w *wsConnWrapper) Write(ctx context.Context, m Message) error {
	var mt websocket.MessageType
	switch m.Type {
	case MessageText:
		mt = websocket.MessageText
	case MessageBinary:
		mt = websocket.MessageBinary
	default:
		mt = websocket.MessageText
	}
	return w.conn.Write(ctx, mt, m.Data)
}

func (w *wsConnWrapper) Close() error {
	return w.conn.Close(websocket.StatusNormalClosure, "")
}

// ServeConn is a low-level entry point that drives a single connection.
// It reads messages, dispatches them through the Server, and writes back
// responses. Used for tests.
func (s *Server) ServeConn(ctx context.Context, c Conn) error {
	for {
		msg, err := c.Read(ctx)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		out, err := s.HandleRaw(ctx, msg.Data)
		if err != nil {
			return err
		}
		if out == nil {
			continue
		}
		if err := c.Write(ctx, Message{Type: MessageText, Data: out}); err != nil {
			return err
		}
	}
}

// Close stops all listeners.
func (t *ServerTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return nil
	}
	t.closed = true
	if t.srv != nil {
		_ = t.srv.Shutdown(context.Background())
	}
	for _, ln := range t.listeners {
		_ = ln.Close()
	}
	return nil
}

// -----------------------------------------------------------------------------
// Binding
// -----------------------------------------------------------------------------

// bind parses an "scheme://addr" string and creates a listener.
func bind(addr string) (net.Listener, error) {
	scheme, host, err := parseAddr(addr)
	if err != nil {
		return nil, err
	}
	switch scheme {
	case "tcp":
		return net.Listen("tcp", host)
	case "unix":
		return net.Listen("unix", host)
	default:
		return nil, fmt.Errorf("ipc: unsupported scheme %q (want tcp or unix)", scheme)
	}
}

func parseAddr(addr string) (string, string, error) {
	for i := 0; i < len(addr); i++ {
		if addr[i] == ':' && i+1 < len(addr) && addr[i+1] == '/' && i+2 < len(addr) && addr[i+2] == '/' {
			return addr[:i], addr[i+3:], nil
		}
	}
	// Default to tcp.
	return "tcp", addr, nil
}

// isWebsocketUpgrade reports whether the request is a WebSocket upgrade.
func isWebsocketUpgrade(r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}
	if !tokenListContainsToken(r.Header.Get("Connection"), "Upgrade") &&
		!tokenListContainsToken(r.Header.Get("Connection"), "upgrade") {
		return false
	}
	return stringsEqualFold(r.Header.Get("Upgrade"), "websocket")
}

func tokenListContainsToken(h, want string) bool {
	if h == "" {
		return false
	}
	for h != "" {
		var tok string
		if i := indexByte(h, ','); i >= 0 {
			tok = trim(h[:i])
			h = h[i+1:]
		} else {
			tok = trim(h)
			h = ""
		}
		if stringsEqualFold(tok, want) {
			return true
		}
	}
	return false
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func trim(s string) string {
	start, end := 0, len(s)
	for start < end {
		c := s[start]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			break
		}
		start++
	}
	for end > start {
		c := s[end-1]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			break
		}
		end--
	}
	return s[start:end]
}

func stringsEqualFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
