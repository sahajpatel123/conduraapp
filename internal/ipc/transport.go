package ipc

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/sahajpatel123/synapticapp/internal/sse"
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

// Message type values (mirror coder/websocket).
const (
	MessageText   MessageType = 1
	MessageBinary MessageType = 2
)

// -----------------------------------------------------------------------------
// Server transport
// -----------------------------------------------------------------------------

// ServerTransport bundles a method registry with a network listener.
//
// Listen addrs: pass "tcp://127.0.0.1:0" to bind a random TCP port, or
// "unix:///tmp/synaptic.sock" for a Unix socket.
type ServerTransport struct {
	S     *Server
	Token string // optional bearer token; if non-empty, clients must send it

	// SSE is an optional Server-Sent Events broker. If set, GET /events
	// is mounted on the same HTTP mux so the GUI can subscribe to
	// streaming events (LLM tokens, audit, spend) over the same
	// listener. Auth is enforced uniformly — the Token (if set) is
	// checked before the broker takes over the response.
	SSE *sse.Broker

	mu        sync.Mutex
	closed    bool
	listeners []net.Listener
	srv       *http.Server

	// sseTickets holds short-lived one-time tickets that let
	// EventSource connect to /events without putting the real
	// bearer token in the URL query string (where it would appear
	// in server logs, browser history, and Referer headers). Each
	// ticket is single-use and expires 30 seconds after issue.
	sseTickets   map[string]time.Time
	sseTicketsMu sync.Mutex
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
	ln, err := bind(ctx, addr)
	if err != nil {
		return err
	}
	t.mu.Lock()
	t.listeners = append(t.listeners, ln)
	t.mu.Unlock()
	go t.serveListener(ln)
	return nil
}

// serveListener runs the HTTP+WebSocket handler on one listener.
func (t *ServerTransport) serveListener(ln net.Listener) {
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
//   - POST /sse-ticket -> exchange bearer token for one-time SSE ticket
//   - GET /events  -> SSE broker (if configured), ticket or token auth
//   - GET /ws      -> WebSocket upgrade
//   - POST /       -> JSON-RPC
func (t *ServerTransport) handleHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/healthz" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return
	}
	// SSE ticket exchange: the client POSTs with Authorization:
	// Bearer <token> (header, not URL) and receives a short-lived
	// one-time ticket. The ticket is then used as ?ticket= on the
	// EventSource URL. This keeps the real token out of URLs/logs.
	if r.Method == http.MethodPost && r.URL.Path == "/sse-ticket" {
		if !t.authorize(w, r) {
			return
		}
		t.handleSSETicket(w, r)
		return
	}
	// SSE endpoint: mount the broker at /events.
	if r.Method == http.MethodGet && r.URL.Path == "/events" {
		if !t.authorizeSSE(w, r) {
			return
		}
		if t.SSE == nil {
			http.Error(w, "events not enabled", http.StatusNotImplemented)
			return
		}
		t.SSE.ServeHTTP(w, r)
		return
	}
	if !t.authorize(w, r) {
		return
	}

	if isWebsocketUpgrade(r) {
		t.serveWebSocket(w, r)
		return
	}
	t.handleJSONRPC(w, r)
}

// authorize enforces the bearer-token check, if a token is
// configured. Returns true if the request is allowed to proceed; if
// false, the response has already been written with 401.
func (t *ServerTransport) authorize(w http.ResponseWriter, r *http.Request) bool {
	if t.Token == "" {
		return true
	}
	// Check the Authorization header (constant-time comparison).
	auth := r.Header.Get("Authorization")
	expected := "Bearer " + t.Token
	if subtle.ConstantTimeCompare([]byte(auth), []byte(expected)) == 1 {
		return true
	}
	http.Error(w, "unauthorized", http.StatusUnauthorized)
	return false
}

// authorizeSSE is the SSE-specific auth check. It accepts either the
// Authorization header (for non-EventSource clients) or a one-time
// ticket from /sse-ticket. The real bearer token is never accepted
// as a query parameter — that was the old insecure path that leaked
// it into server logs and browser history.
func (t *ServerTransport) authorizeSSE(w http.ResponseWriter, r *http.Request) bool {
	if t.Token == "" {
		return true
	}
	// Header auth (for non-EventSource clients like curl, constant-time comparison).
	if auth := r.Header.Get("Authorization"); subtle.ConstantTimeCompare([]byte(auth), []byte("Bearer "+t.Token)) == 1 {
		return true
	}
	// One-time ticket from /sse-ticket.
	ticket := r.URL.Query().Get("ticket")
	if ticket != "" && t.consumeSSETicket(ticket) {
		return true
	}
	http.Error(w, "unauthorized", http.StatusUnauthorized)
	return false
}

// handleSSETicket issues a short-lived one-time ticket for SSE auth.
// The client must authenticate with the real bearer token (in the
// Authorization header) to get a ticket. The ticket is then used as
// ?ticket= on the EventSource URL.
func (t *ServerTransport) handleSSETicket(w http.ResponseWriter, _ *http.Request) {
	ticket, err := t.issueSSETicket()
	if err != nil {
		http.Error(w, "failed to generate ticket", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ticket":     ticket,
		"expires_in": 30,
	})
}

// issueSSETicket generates a random one-time ticket and stores it
// with a 30-second TTL.
func (t *ServerTransport) issueSSETicket() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	ticket := hex.EncodeToString(raw)

	t.sseTicketsMu.Lock()
	if t.sseTickets == nil {
		t.sseTickets = make(map[string]time.Time)
	}
	// Garbage-collect expired tickets on every issue.
	now := time.Now()
	for k, exp := range t.sseTickets {
		if now.After(exp) {
			delete(t.sseTickets, k)
		}
	}
	t.sseTickets[ticket] = now.Add(30 * time.Second)
	t.sseTicketsMu.Unlock()

	return ticket, nil
}

// consumeSSETicket validates and removes a one-time ticket. Returns
// false if the ticket is unknown, already used, or expired.
func (t *ServerTransport) consumeSSETicket(ticket string) bool {
	t.sseTicketsMu.Lock()
	defer t.sseTicketsMu.Unlock()
	exp, ok := t.sseTickets[ticket]
	if !ok {
		return false
	}
	delete(t.sseTickets, ticket) // single-use
	return time.Now().Before(exp)
}

// handleJSONRPC reads the body, dispatches the call, and writes the
// response.
func (t *ServerTransport) handleJSONRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer func() { _ = r.Body.Close() }()
	const maxBodySize = 10 << 20 // 10 MB
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize+1))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}
	if int64(len(body)) > maxBodySize {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return
	}
	out, err := t.S.HandleRaw(r.Context(), body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if out == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// serveWebSocket upgrades the connection and runs the read/write loop.
func (t *ServerTransport) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	// Verify the Origin header to prevent cross-site WebSocket
	// hijacking. We allow localhost origins (127.0.0.1, localhost,
	// [::1]) which is the expected production topology.
	if !t.validateWSOrigin(r) {
		http.Error(w, "forbidden: invalid origin", http.StatusForbidden)
		return
	}
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, //nolint:gosec // Origin validated above
	})
	if err == nil {
		c.SetReadLimit(10 << 20) // 10 MB max WebSocket message
	}
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

// validateWSOrigin checks that the Origin header is a localhost
// origin. This prevents cross-site WebSocket hijacking while
// allowing the local GUI to connect.
func (t *ServerTransport) validateWSOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Allow requests with no Origin (non-browser clients).
		return true
	}
	// Parse the origin URL to extract the host.
	for i := 0; i < len(origin); i++ {
		if origin[i] == ':' && i+2 < len(origin) && origin[i+1] == '/' && origin[i+2] == '/' {
			host := origin[i+3:]
			// Strip port if present.
			for j := 0; j < len(host); j++ {
				if host[j] == ':' {
					host = host[:j]
					break
				}
			}
			switch host {
			case "localhost", "127.0.0.1", "[::1]":
				return true
			}
			return false
		}
	}
	return false
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
func bind(ctx context.Context, addr string) (net.Listener, error) {
	scheme, host := parseAddr(addr)
	lc := &net.ListenConfig{}
	switch scheme {
	case "tcp":
		return lc.Listen(ctx, "tcp", host)
	case "unix":
		return lc.Listen(ctx, "unix", host)
	default:
		return nil, fmt.Errorf("ipc: unsupported scheme %q (want tcp or unix)", scheme)
	}
}

func parseAddr(addr string) (string, string) {
	for i := 0; i < len(addr); i++ {
		if addr[i] == ':' && i+1 < len(addr) && addr[i+1] == '/' && i+2 < len(addr) && addr[i+2] == '/' {
			return addr[:i], addr[i+3:]
		}
	}
	// Default to tcp.
	return "tcp", addr
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
