package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/ipc"
)

// IPCClient wraps the daemon's JSON-RPC connection for the TUI.
type IPCClient struct {
	conn    net.Conn
	encoder *json.Encoder
	decoder *json.Decoder
	mu      sync.Mutex
	pending map[int]chan *ipc.Response
	nextID  int
	logger  *slog.Logger
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

// NewIPCClient connects to the daemon via Unix socket or TCP.
func NewIPCClient(addr string, logger *slog.Logger) (*IPCClient, error) {
	var conn net.Conn
	var err error
	for i := 0; i < 10; i++ {
		conn, err = net.Dial("unix", addr)
		if err == nil {
			break
		}
		conn, err = net.Dial("tcp", addr)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		return nil, fmt.Errorf("connect to daemon at %s: %w", addr, err)
	}

	c := &IPCClient{
		conn:    conn,
		encoder: json.NewEncoder(conn),
		decoder: json.NewDecoder(conn),
		pending: make(map[int]chan *ipc.Response),
		logger:  logger,
	}
	go c.readLoop()
	return c, nil
}

func (c *IPCClient) readLoop() {
	for {
		var resp rpcResponse
		if err := c.decoder.Decode(&resp); err != nil {
			c.logger.Debug("IPC read error", "err", err)
			return
		}
		c.mu.Lock()
		ch, ok := c.pending[resp.ID]
		if ok {
			delete(c.pending, resp.ID)
		}
		c.mu.Unlock()
		if ok {
			res := &ipc.Response{Result: resp.Result, Error: resp.Error}
			select {
			case ch <- res:
			default:
			}
		}
	}
}

// Call sends a JSON-RPC request and waits for response.
func (c *IPCClient) Call(ctx context.Context, method string, params any) (*ipc.Response, error) {
	c.mu.Lock()
	id := c.nextID
	c.nextID++
	ch := make(chan *ipc.Response, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	paramsBytes, _ := json.Marshal(params)
	req := rpcRequest{JSONRPC: "2.0", ID: id, Method: method, Params: paramsBytes}
	if err := c.encoder.Encode(req); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, err
	}

	select {
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, ctx.Err()
	case resp := <-ch:
		return resp, nil
	case <-time.After(30 * time.Second):
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("request timeout")
	}
}

// Notify sends a notification (no response expected).
func (c *IPCClient) Notify(method string, params any) error {
	paramsBytes, _ := json.Marshal(params)
	req := rpcRequest{JSONRPC: "2.0", Method: method, Params: paramsBytes}
	return c.encoder.Encode(req)
}

func (c *IPCClient) Close() error {
	return c.conn.Close()
}