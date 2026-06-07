// Package ipc also provides a Client that talks JSON-RPC 2.0 to a
// Server over a local TCP or Unix socket. The Client speaks plain
// HTTP POST for now; a future revision may upgrade to a long-lived
// WebSocket for streaming.

package ipc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Client is a thin JSON-RPC 2.0 client over HTTP. It is safe for
// concurrent use.
type Client struct {
	addr    string // "tcp://127.0.0.1:7666" or "unix:///tmp/synapse.sock"
	token   string // optional bearer token
	httpc   *http.Client
	scheme  string // "http" or "https"
	host    string // host:port for tcp, or path for unix
	mu      sync.Mutex
	idCtr   atomic.Int64
}

// Dial creates a Client. The addr has the form
//
//	tcp://127.0.0.1:7666
//	unix:///tmp/synapticd.sock
//
// An optional bearer token is sent in the Authorization header.
func Dial(addr, token string) (*Client, error) {
	scheme, host, err := parseAddr(addr)
	if err != nil {
		return nil, err
	}
	c := &Client{
		addr:   addr,
		token:  token,
		scheme: "http",
		host:   host,
	}
	switch scheme {
	case "tcp":
		// Plain HTTP for now (localhost). http.Client zero value uses
		// a sensible default transport; override DialContext to support
		// unix sockets if we ever dial "unix://".
		c.httpc = &http.Client{Timeout: 30 * time.Second}
	case "unix":
		c.httpc = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					var d net.Dialer
					return d.DialContext(ctx, "unix", host)
				},
			},
		}
		// For unix sockets the request URL just needs any absolute path
		// and a host header; use a placeholder.
		c.host = "localhost"
	default:
		return nil, fmt.Errorf("ipc: unsupported scheme %q (want tcp or unix)", scheme)
	}
	return c, nil
}

// Addr returns the address the client was configured with.
func (c *Client) Addr() string { return c.addr }

// Close releases any resources. The current HTTP client has no
// long-lived connections to close, but the method exists for
// API symmetry with future WebSocket-based clients.
func (c *Client) Close() error { return nil }

// Call issues a JSON-RPC 2.0 request to method, marshaling params and
// unmarshaling the result into out (or returning the response error).
func (c *Client) Call(ctx context.Context, method string, params, out any) error {
	if params == nil {
		params = struct{}{}
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}
	id := c.idCtr.Add(1)
	req := Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(fmt.Sprintf("%d", id)),
		Method:  method,
		Params:  paramsJSON,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: "/"}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpc.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNoContent {
		// Notification - no response expected.
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var rpcResp Response
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if rpcResp.Error != nil {
		return rpcResp.Error
	}
	if out != nil && rpcResp.Result != nil {
		// Result is `any` (decoded into a generic Go value). Re-marshal
		// and unmarshal into the caller's destination type.
		raw, err := json.Marshal(rpcResp.Result)
		if err != nil {
			return fmt.Errorf("re-marshal result: %w", err)
		}
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("unmarshal result: %w", err)
		}
	}
	return nil
}

// ReadAddrFile reads the listen address that synapticd writes to
// <data_dir>/synapticd.addr. Returns "" if the file does not exist.
func ReadAddrFile(dataDir string) string {
	path := dataDir + "/synapticd.addr"
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// DefaultDataDir returns ~/.synaptic, the same default the daemon uses.
func DefaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home + "/.synaptic"
}

// IsConnRefused reports whether err indicates the daemon is not
// running (or not listening on the expected address).
func IsConnRefused(err error) bool {
	if err == nil {
		return false
	}
	// net.OpError wraps the underlying syscalls; fall back to substring
	// matching so we don't have to import syscall across platforms.
	s := err.Error()
	return strings.Contains(s, "connection refused") ||
		strings.Contains(s, "no such file or directory")
}
