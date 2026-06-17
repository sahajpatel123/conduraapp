package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
)

// Client communicates with a single MCP server via JSON-RPC.
type Client struct {
	cfg    ServerConfig
	cmd    *exec.Cmd
	stdin  *bufio.Writer
	stdout *bufio.Reader
	mu     sync.Mutex
	ioMu   sync.Mutex // serializes requests/responses over stdin/stdout
	reqID  atomic.Int64
}

// NewClient creates a client for an MCP server.
func NewClient(cfg ServerConfig) *Client {
	return &Client{cfg: cfg}
}

// Connect starts the server subprocess and performs the MCP handshake.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cfg.Type == TransportStdio {
		c.cmd = exec.CommandContext(ctx, c.cfg.Command, c.cfg.Args...) //nolint:gosec // MCP servers are user-installed, not arbitrary
		if len(c.cfg.Env) > 0 {
			c.cmd.Env = c.cmd.Environ()
			for k, v := range c.cfg.Env {
				c.cmd.Env = append(c.cmd.Env, k+"="+v)
			}
		}
		stdinPipe, err := c.cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("mcp: stdin: %w", err)
		}
		stdoutPipe, err := c.cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("mcp: stdout: %w", err)
		}
		if err := c.cmd.Start(); err != nil {
			return fmt.Errorf("mcp: start: %w", err)
		}
		c.stdin = bufio.NewWriter(stdinPipe)
		c.stdout = bufio.NewReader(stdoutPipe)

		// Initialize handshake.
		if _, err := c.call(ctx, "initialize", map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]string{
				"name":    "condura",
				"version": "0.1.0",
			},
		}); err != nil {
			_ = c.Close()
			return fmt.Errorf("mcp: initialize: %w", err)
		}
	}
	return nil
}

func (c *Client) call(_ context.Context, method string, params any) (json.RawMessage, error) {
	c.ioMu.Lock()
	defer c.ioMu.Unlock()

	id := c.reqID.Add(1)
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
	}
	if params != nil {
		b, _ := json.Marshal(params)
		req.Params = b
	}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')

	if _, err := c.stdin.Write(data); err != nil {
		return nil, fmt.Errorf("mcp: write: %w", err)
	}
	if err := c.stdin.Flush(); err != nil {
		return nil, fmt.Errorf("mcp: flush: %w", err)
	}

	line, err := c.stdout.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("mcp: read: %w", err)
	}
	var resp JSONRPCResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("mcp: parse: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("mcp: %s", resp.Error.Message)
	}
	return resp.Result, nil
}

// ListTools fetches the available tools from the server.
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	raw, err := c.call(ctx, "tools/list", nil)
	if err != nil {
		return nil, err
	}
	var result struct {
		Tools []Tool `json:"tools"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	for i := range result.Tools {
		result.Tools[i].ServerName = c.cfg.Name
	}
	return result.Tools, nil
}

// CallTool executes a tool on the server.
func (c *Client) CallTool(ctx context.Context, name string, args map[string]any) (json.RawMessage, error) {
	return c.call(ctx, "tools/call", map[string]any{
		"name":      name,
		"arguments": args,
	})
}

// Close terminates the server subprocess.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		_ = c.cmd.Wait()
	}
	return nil
}

// GatedClient wraps an MCP client with Gatekeeper enforcement.
// Every tool call passes through gatekeeper.Evaluate before execution.
type GatedClient struct {
	client *Client
	gate   gatekeeper.Gatekeeper
}

// NewGatedClient creates a Gatekeeper-wrapped MCP client.
func NewGatedClient(cfg ServerConfig, gate gatekeeper.Gatekeeper) *GatedClient {
	return &GatedClient{
		client: NewClient(cfg),
		gate:   gate,
	}
}

// Connect starts the server and initializes.
func (g *GatedClient) Connect(ctx context.Context) error {
	return g.client.Connect(ctx)
}

// ListTools returns available tools.
func (g *GatedClient) ListTools(ctx context.Context) ([]Tool, error) {
	return g.client.ListTools(ctx)
}

// CallTool executes a tool through the Gatekeeper.
// Uses blastradius to classify the action.
func (g *GatedClient) CallTool(ctx context.Context, name string, args map[string]any) (*ToolCallResult, error) {
	start := time.Now()

	// Gatekeeper: evaluate before execution.
	ba := blastradius.Action{Kind: "mcp.tool_call", TargetApp: g.client.cfg.Name, Body: fmt.Sprintf("%v", args)}
	decision, reason := g.gate.Evaluate(ctx, ba)
	if decision != gatekeeper.Allow {
		return &ToolCallResult{
			ServerName: g.client.cfg.Name,
			ToolName:   name,
			IsError:    true,
			Duration:   time.Since(start),
			Content:    []ContentBlock{{Type: "text", Text: "gatekeeper denied: " + reason}},
		}, fmt.Errorf("mcp: gatekeeper denied %s: %s", name, reason)
	}

	raw, err := g.client.CallTool(ctx, name, args)
	if err != nil {
		return &ToolCallResult{ //nolint:nilerr
			ServerName: g.client.cfg.Name,
			ToolName:   name,
			IsError:    true,
			Duration:   time.Since(start),
			Content:    []ContentBlock{{Type: "text", Text: err.Error()}},
		}, err
	}

	var result ToolCallResult
	if err := json.Unmarshal(raw, &result); err != nil {
		result.ServerName = g.client.cfg.Name
		result.ToolName = name
		result.Duration = time.Since(start)
		result.Content = []ContentBlock{{Type: "text", Text: string(raw)}}
		return &result, err
	}
	result.ServerName = g.client.cfg.Name
	result.Duration = time.Since(start)
	return &result, nil
}

// Close terminates the server.
func (g *GatedClient) Close() error { return g.client.Close() }
