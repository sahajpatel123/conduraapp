// Package mcp implements the Model Context Protocol (MCP) gateway.
// It discovers, manages, and communicates with MCP servers via
// stdio/HTTP/SSE transport. Tool calls are prefixed (mcp__<server>__<tool>)
// and every execution passes through the Gatekeeper.
//
// Discovery: scans ~/.synaptic/mcp/ or config for server manifests.
// Never auto-starts — requires explicit user enable. Never scans $PATH.
package mcp

import (
	"encoding/json"
	"time"
)

// ServerConfig describes how to connect to an MCP server.
type ServerConfig struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
	Type    TransportType     `json:"type"`
	URL     string            `json:"url,omitempty"`
	Enabled bool              `json:"enabled"`
}

// TransportType is the connection transport for an MCP server.
type TransportType string

const (
	// TransportStdio uses subprocess stdio for communication.
	TransportStdio TransportType = "stdio"
	// TransportHTTP uses HTTP to communicate with the server.
	TransportHTTP TransportType = "http"
	// TransportSSE uses Server-Sent Events for streaming.
	TransportSSE TransportType = "sse"
)

// Tool represents a single tool exposed by an MCP server.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ServerName  string `json:"server_name"`
	InputSchema any    `json:"input_schema"`
}

// ToolCallRequest is a request to execute an MCP tool.
type ToolCallRequest struct {
	PrefixedName string         `json:"prefixed_name"`
	Arguments    map[string]any `json:"arguments"`
}

// ToolCallResult is the result of executing an MCP tool.
type ToolCallResult struct {
	ServerName string         `json:"server_name"`
	ToolName   string         `json:"tool_name"`
	Content    []ContentBlock `json:"content"`
	IsError    bool           `json:"is_error"`
	Duration   time.Duration  `json:"duration_ms"`
}

// ContentBlock is a block of content returned by a tool call.
type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	MIMEType string `json:"mime_type,omitempty"`
	Data     string `json:"data,omitempty"`
}

// JSONRPCRequest is the standard MCP request envelope.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse is the standard MCP response envelope.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError is an MCP error response.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
