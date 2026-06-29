package daemon

import (
	"context"
	"encoding/json"

	"github.com/sahajpatel123/conduraapp/internal/ipc"
	"github.com/sahajpatel123/conduraapp/internal/mcp"
)

// registerMCPMethods registers MCP Gateway RPC methods.
func registerMCPMethods(srv *ipc.Server, subs *Subsystems) {
	if subs.MCP == nil {
		return
	}

	srv.Register("mcp.list_servers", func(_ context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{"servers": subs.MCP.ListServers()}, nil
	})

	srv.Register("mcp.list_tools", func(ctx context.Context, _ json.RawMessage) (any, error) {
		tools, err := subs.MCP.ListTools(ctx)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return map[string]any{"tools": tools}, nil
	})

	srv.Register("mcp.call_tool", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Name string         `json:"name"`
			Args map[string]any `json:"args"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p.Name == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "name is required"}
		}
		result, err := subs.MCP.CallTool(ctx, p.Name, p.Args)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return result, nil
	})

	srv.Register("mcp.install_server", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p mcp.ServerConfig
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p.Name == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "server name is required"}
		}
		if err := subs.MCP.Install(p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return auditOK(), nil
	})

	srv.Register("mcp.remove_server", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Name string `json:"name"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if err := subs.MCP.Remove(p.Name); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return auditOK(), nil
	})

	srv.Register("mcp.start_server", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Name string `json:"name"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if err := subs.MCP.Start(ctx, p.Name); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return auditOK(), nil
	})

	srv.Register("mcp.stop_server", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Name string `json:"name"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if err := subs.MCP.Remove(p.Name); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return auditOK(), nil
	})
}
