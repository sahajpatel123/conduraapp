package tui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/ipc"
)

// IPCClient communicates with the daemon via HTTP JSON-RPC.
type IPCClient struct {
	baseURL    string
	httpClient *http.Client
	mu         sync.Mutex
	nextID     int
	logger     *slog.Logger
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ipc.Error      `json:"error,omitempty"`
}

// parseAddr splits "scheme://host:port" into scheme and host.
func parseAddr(addr string) (string, string) {
	for i := 0; i < len(addr); i++ {
		if addr[i] == ':' && i+2 < len(addr) && addr[i+1] == '/' && addr[i+2] == '/' {
			return addr[:i], addr[i+3:]
		}
	}
	return "tcp", addr
}

// NewIPCClient connects to the daemon at the given address.
func NewIPCClient(addr string, logger *slog.Logger) (*IPCClient, error) {
	scheme, host := parseAddr(addr)
	baseURL := "http://" + host

	var transport http.RoundTripper
	if scheme == "unix" {
		transport = &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", host)
			},
		}
	} else {
		transport = &http.Transport{
			DialContext: (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
		}
	}

	client := &http.Client{Transport: transport, Timeout: 30 * time.Second}

	// Verify connectivity.
	pingBody, _ := json.Marshal(rpcRequest{JSONRPC: "2.0", ID: 0, Method: "ping"})
	httpReq, err := http.NewRequest(http.MethodPost, baseURL, bytes.NewReader(pingBody))
	if err != nil {
		return nil, fmt.Errorf("connect to daemon at %s: %w", addr, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("connect to daemon at %s: %w", addr, err)
	}
	_ = resp.Body.Close()

	return &IPCClient{
		baseURL:    baseURL,
		httpClient: client,
		logger:     logger,
	}, nil
}

// Call sends a JSON-RPC request and waits for the response.
func (c *IPCClient) Call(ctx context.Context, method string, params any) (*ipc.Response, error) {
	c.mu.Lock()
	id := c.nextID
	c.nextID++
	c.mu.Unlock()

	paramsBytes, _ := json.Marshal(params)
	req := rpcRequest{JSONRPC: "2.0", ID: id, Method: method, Params: paramsBytes}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &ipc.Response{Result: rpcResp.Result, Error: rpcResp.Error}, nil
}

// Notify sends a notification (no response expected).
func (c *IPCClient) Notify(method string, params any) error {
	paramsBytes, _ := json.Marshal(params)
	req := rpcRequest{JSONRPC: "2.0", Method: method, Params: paramsBytes}
	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}

// Close is a no-op; HTTP connections are managed by the transport.
func (c *IPCClient) Close() error { return nil }

// FindDaemonAddr discovers the daemon's IPC address.
func FindDaemonAddr() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	for _, dataDir := range []string{filepath.Join(home, ".condura"), "/tmp/synaptic"} {
		addrFile := filepath.Join(dataDir, "condurad.addr")
		data, err := os.ReadFile(addrFile)
		if err == nil {
			return strings.TrimSpace(string(data))
		}
		sockPath := filepath.Join(dataDir, "condurad.sock")
		if _, err := os.Stat(sockPath); err == nil {
			return "unix://" + sockPath
		}
	}
	return ""
}
