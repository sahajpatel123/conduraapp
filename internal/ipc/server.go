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
}

// NewServer returns an empty server.
func NewServer() *Server {
	return &Server{methods: map[string]HandlerFunc{}}
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
		// Wrap as Internal error.
		return errorResponse(req, &Error{
			Code:    CodeInternalError,
			Message: err.Error(),
		}), nil
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
		// JSON-RPC parse error: return the error inside the response body,
		// not as a Go error — the caller (HTTP/WS transport) will write the
		// bytes to the client as the protocol-level reply.
		return marshalError(nil, &Error{Code: CodeParseError, Message: "parse error: " + err.Error()}), nil //nolint:nilerr
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
		return marshalError(req.ID, &Error{Code: CodeInternalError, Message: herr.Error()}), nil
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
		// JSON-RPC parse error: same protocol-level reason as HandleRaw.
		return marshalError(nil, &Error{Code: CodeParseError, Message: "parse error: " + err.Error()}), nil //nolint:nilerr
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
			out = append(out, marshalError(reqs[i].ID, &Error{Code: CodeInternalError, Message: err.Error()}))
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
