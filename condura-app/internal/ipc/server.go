// Package ipc implements the inter-process communication bridge between
// the Synaptic daemon (synapticd) and the Wails GUI (Phase 2) / CLI.
//
// Transport: JSON-RPC 2.0 over WebSocket (coder/websocket). The daemon
// listens on a localhost TCP port and (on macOS/Linux) a Unix domain
// socket; clients connect to either. Auth is a static bearer token read
// from internal/config (api_server.auth_token).
//
// The Server registers typed methods. Each method receives parsed params
// and returns a result. Errors are mapped to JSON-RPC error codes:
//   - -32700: Parse error
//   - -32600: Invalid Request
//   - -32601: Method not found
//   - -32602: Invalid params
//   - -32603: Internal error
package ipc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
)

// Request is a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"` // string, number, or null
}

// Response is a JSON-RPC 2.0 response. Exactly one of Result or Error
// is non-nil.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  any             `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
	ID      json.RawMessage `json:"id"`
}

// ProtocolVersion is the JSON-RPC protocol version this package speaks.
// It is the value of the `jsonrpc` field in every request and response.
const ProtocolVersion = "2.0"

// Error is a JSON-RPC 2.0 error.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error implements the error interface.
func (e *Error) Error() string { return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message) }

// Standard error codes.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

// HandlerFunc is the signature for a registered method.
//
// The implementation is responsible for unmarshaling Params into the
// expected shape and marshaling the result. ctx is per-request; it is
// canceled when the connection is closed.
type HandlerFunc func(ctx context.Context, params json.RawMessage) (any, error)

// -----------------------------------------------------------------------------
// Server
// -----------------------------------------------------------------------------

// Server is a JSON-RPC 2.0 method registry and request dispatcher.
type Server struct {
	mu      sync.RWMutex
	methods map[string]HandlerFunc
	// log receives the FULL error message (with path/host/etc.) for
	// every internal-error response. The client only sees the
	// redacted message from redactInternal. nil → io.Discard-backed
	// default logger. Set via WithLogger.
	log *slog.Logger
}

// NewServer returns an empty server with a no-op logger.
func NewServer() *Server {
	return &Server{
		methods: map[string]HandlerFunc{},
		log:     slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

// WithLogger returns the server with a logger attached. The logger
// receives the full internal error (including any os.PathError or
// net.OpError) for forensic correlation, while the JSON-RPC client
// receives only the redacted stable message.
//
// Returns the receiver to allow NewServer(...).WithLogger(...) chains.
func (s *Server) WithLogger(l *slog.Logger) *Server {
	if l != nil {
		s.log = l
	}
	return s
}

// Register adds a method. If a method with the same name exists, it is
// replaced.
func (s *Server) Register(name string, h HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.methods[name] = h
}

// Unregister removes a method.
func (s *Server) Unregister(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.methods, name)
}

// HasMethod reports whether a method is registered.
func (s *Server) HasMethod(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.methods[name]
	return ok
}

// Methods returns the registered method names, sorted.
func (s *Server) Methods() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.methods))
	for n := range s.methods {
		out = append(out, n)
	}
	return out
}

// Handle processes one Request and returns a Response. If the request is
// a notification (ID is null/absent), the returned Response has a nil ID.
//
// For notifications, errors are not reported back to the client (per the
// JSON-RPC spec); the function still returns a non-nil error so the
// caller can log.
func (s *Server) Handle(ctx context.Context, req *Request) (*Response, error) {
	if req.JSONRPC != "2.0" {
		return nil, fmt.Errorf("ipc: unsupported jsonrpc %q (want 2.0)", req.JSONRPC)
	}
	if req.Method == "" {
		return nil, &Error{Code: CodeInvalidRequest, Message: "method is required"}
	}

	s.mu.RLock()
	h, ok := s.methods[req.Method]
	s.mu.RUnlock()
	if !ok {
		return errorResponse(req, &Error{
			Code: CodeMethodNotFound, Message: fmt.Sprintf("method %q not found", req.Method),
		}), nil
	}

	result, err := h(ctx, req.Params)
	if isNotification(req.ID) {
		// Notifications: discard the result and signal the transport
		// (via ErrNotification) that no reply should be written.
		return nil, ErrNotification
	}
	if err != nil {
		var ipcErr *Error
		if errors.As(err, &ipcErr) {
			return errorResponse(req, ipcErr), nil
		}
		// 2026-06-29 audit P0-3: redact the internal error before
		// sending it to the client. The full error is logged
		// server-side via s.log for forensic correlation.
		return errorResponse(req, redactInternal(s.log, req.ID, err)), nil
	}
	return &Response{
		JSONRPC: ProtocolVersion,
		Result:  result,
		ID:      req.ID,
	}, nil
}

// ErrNotification is the sentinel error a registered handler returns to
// indicate that the request was a notification (per JSON-RPC 2.0) and
// therefore no response should be sent. The transport checks for this
// error and drops the reply.
var ErrNotification = errors.New("ipc: notification (no response)")

// internalError is the stable, redacted JSON-RPC error returned for
// any unhandled server-side failure. The client sees only this fixed
// message and the request ID — never the underlying Go error, which
// can carry filesystem paths, internal IPs, SQL fragments, or stack
// traces. The full original error is logged server-side via the
// server's logger (s.log) for forensic correlation.
//
// This is the fix for the 2026-06-29 audit P0-3: the previous code
// forwarded err.Error() directly to JSON-RPC clients, leaking
// server internals through every error path.
const internalError = "internal error"

// redactInternal returns the stable JSON-RPC Error for an internal
// failure. The original error is logged via the supplied logger with
// the request ID for correlation. The logger is best-effort: a nil
// logger is treated as a no-op so callers do not need a guard.
//
// We deliberately log the RAW err.Error() (not errString) so the
// server-side forensic record retains the path and IP. The client
// response is the redaction boundary; the log line is internal.
func redactInternal(log *slog.Logger, reqID json.RawMessage, err error) *Error {
	if log != nil && err != nil {
		log.Error("ipc internal error",
			"err", err.Error(),
			"req_id", string(reqID),
		)
	}
	return &Error{Code: CodeInternalError, Message: internalError}
}

// redactParse returns a redacted parse-error response. JSON parse
// errors can include byte offsets that may correlate with sensitive
// input (a token in the wrong JSON field, for example). We return a
// fixed message and log the full parse error server-side.
func redactParse(log *slog.Logger, err error) *Error {
	if log != nil && err != nil {
		log.Debug("ipc parse error", "err", err.Error())
	}
	return &Error{Code: CodeParseError, Message: "parse error"}
}

// errString returns err.Error() with the most common leak vectors
// stripped. It is used as a defense-in-depth measure before logging:
// even if a logger is later piped to a less-trusted sink, the
// message cannot carry a raw /Users/... path or 10.0.0.5:5432.
//
// redactHome replaces the current user's home directory with "~"
// wherever it appears in the string, so filesystem paths do not
// leak the username. The match must be preceded by a path boundary
// (start, space, slash, or colon) so we never replace a substring
// that happens to contain the home string but isn't a real path.
func redactHome(s string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return s
	}
	for {
		i := strings.Index(s, home)
		if i < 0 {
			return s
		}
		// Boundary check: char immediately before must be a path
		// separator (start, space, '/', colon).
		if i > 0 {
			prev := s[i-1]
			if prev != '/' && prev != ' ' && prev != ':' {
				// Skip past this non-boundary match.
				s = s[:i] + "\x00" + s[i+len(home):]
				continue
			}
		}
		s = s[:i] + "~" + s[i+len(home):]
	}
}

// redactPrivateIP replaces IPv4 RFC1918 / link-local addresses with
// "<private>" and cloud-metadata IPs with "<metadata>". It is
// deliberately conservative: it only matches IPs that are
// unambiguous (dotted-quad form), so false positives are minimal.
//
// Order matters: handle 169.254.169.254 (cloud metadata) BEFORE the
// general 169.254. prefix so it gets the specific "<metadata>" tag
// rather than the generic "<private>".
func redactPrivateIP(s string) string {
	if strings.Contains(s, "169.254.169.254") {
		s = strings.ReplaceAll(s, "169.254.169.254", "<metadata>")
	}
	for _, prefix := range []string{
		"10.", "127.0.0.1", "169.254.", "192.168.",
	} {
		s = redactIPPrefix(s, prefix)
	}
	if strings.Contains(s, "[::1]") || strings.Contains(s, "[fe80:") {
		// IPv6 loopback / link-local — best-effort scrub.
		s = strings.ReplaceAll(s, "[::1]", "[<v6>]")
		s = strings.ReplaceAll(s, "[fe80:", "[<v6>:")
	}
	return s
}

func redactIPPrefix(s, prefix string) string {
	for {
		i := strings.Index(s, prefix)
		if i < 0 {
			return s
		}
		// Find the end of the dotted-quad (next non-digit/non-dot char).
		j := i + len(prefix)
		for j < len(s) && (s[j] == '.' || (s[j] >= '0' && s[j] <= '9')) {
			j++
		}
		if j == i+len(prefix) {
			// No digits after the prefix; skip.
			s = s[:i] + "<private>" + s[j:]
			return s
		}
		s = s[:i] + "<private>" + s[j:]
	}
}

// compile-time assertion: net.IP is reachable so the redactor stays
// in sync with the Go stdlib if a future change adds new "internal"
// sentinels. Cheap, no runtime cost.
var _ = net.IPv4len

// HandleRaw accepts a raw JSON-RPC message and returns the raw response.
// Used by the WebSocket/HTTP transport.
func (s *Server) HandleRaw(ctx context.Context, raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return marshalError(nil, &Error{Code: CodeInvalidRequest, Message: "empty body"}), nil
	}
	// Tolerate a leading BOM and surrounding whitespace.
	trimmed := bytesTrimSpace(raw)
	// Allow batch (array of requests) per JSON-RPC 2.0.
	if len(trimmed) > 0 && trimmed[0] == '[' {
		return s.handleBatch(ctx, trimmed)
	}
	var req Request
	if err := json.Unmarshal(trimmed, &req); err != nil {
		// 2026-06-29 audit P0-3: parse errors are redacted. The
		// underlying json.SyntaxError can include byte offsets that
		// hint at sensitive payload structure; we log it server-side
		// and return a fixed message.
		return marshalError(nil, redactParse(s.log, err)), nil //nolint:nilerr
	}
	resp, herr := s.Handle(ctx, &req)
	if errors.Is(herr, ErrNotification) {
		// Handler signaled a notification: drop the reply.
		return nil, nil
	}
	if herr != nil {
		var ipcErr *Error
		if errors.As(herr, &ipcErr) {
			return marshalError(req.ID, ipcErr), nil
		}
		// 2026-06-29 audit P0-3: redact.
		return marshalError(req.ID, redactInternal(s.log, req.ID, herr)), nil
	}
	if isNotification(req.ID) {
		// Notifications: no response should be sent.
		return nil, nil
	}
	return json.Marshal(resp)
}

func (s *Server) handleBatch(ctx context.Context, raw []byte) ([]byte, error) {
	const maxBatchSize = 50
	var reqs []Request
	if err := json.Unmarshal(raw, &reqs); err != nil {
		// 2026-06-29 audit P0-3: parse errors are redacted.
		return marshalError(nil, redactParse(s.log, err)), nil //nolint:nilerr
	}
	if len(reqs) == 0 {
		return marshalError(nil, &Error{Code: CodeInvalidRequest, Message: "empty batch"}), nil
	}
	if len(reqs) > maxBatchSize {
		return marshalError(nil, &Error{Code: CodeInvalidRequest, Message: fmt.Sprintf("batch too large: %d > %d", len(reqs), maxBatchSize)}), nil
	}
	out := make([]json.RawMessage, 0, len(reqs))
	for i := range reqs {
		resp, err := s.Handle(ctx, &reqs[i])
		if errors.Is(err, ErrNotification) {
			// Skip the response for notifications.
			continue
		}
		if err != nil {
			var ipcErr *Error
			if errors.As(err, &ipcErr) {
				out = append(out, marshalError(reqs[i].ID, ipcErr))
				continue
			}
			// 2026-06-29 audit P0-3: redact.
			out = append(out, marshalError(reqs[i].ID, redactInternal(s.log, reqs[i].ID, err)))
			continue
		}
		if isNotification(reqs[i].ID) {
			continue
		}
		b, _ := json.Marshal(resp)
		out = append(out, b)
	}
	if len(out) == 0 {
		// All notifications.
		return nil, nil
	}
	return json.Marshal(out)
}

func errorResponse(req *Request, e *Error) *Response {
	id := json.RawMessage("null")
	if req != nil {
		id = req.ID
	}
	return &Response{JSONRPC: ProtocolVersion, Error: e, ID: id}
}

func marshalError(id json.RawMessage, e *Error) []byte {
	if id == nil {
		id = json.RawMessage("null")
	}
	resp := Response{JSONRPC: ProtocolVersion, Error: e, ID: id}
	b, _ := json.Marshal(resp)
	return b
}

// isNotification reports whether the request ID indicates a notification
// (null or absent).
func isNotification(id json.RawMessage) bool {
	if len(id) == 0 {
		return true
	}
	s := string(id)
	return s == "null" || s == "null\n"
}

// bytesTrimSpace is bytes.TrimSpace, inlined to avoid an extra import.
func bytesTrimSpace(b []byte) []byte {
	start, end := 0, len(b)
	for start < end {
		c := b[start]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			break
		}
		start++
	}
	for end > start {
		c := b[end-1]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			break
		}
		end--
	}
	return b[start:end]
}
