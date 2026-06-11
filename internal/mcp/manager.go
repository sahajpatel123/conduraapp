package mcp

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
)

// Manager discovers, starts, and manages MCP servers. Never auto-starts
// — every server requires explicit user enable. Discovery scans
// config only; never scans $PATH.
type Manager struct {
	mu      sync.Mutex
	servers map[string]*GatedClient
	gate    gatekeeper.Gatekeeper
}

// NewManager creates an MCP server manager.
func NewManager(gate gatekeeper.Gatekeeper) *Manager {
	return &Manager{
		gate:    gate,
		servers: make(map[string]*GatedClient),
	}
}

// Install registers a server configuration. Does NOT start the server.
func (m *Manager) Install(cfg ServerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.servers[cfg.Name]; exists {
		return fmt.Errorf("mcp: server %q already installed", cfg.Name)
	}
	m.servers[cfg.Name] = NewGatedClient(cfg, m.gate)
	return nil
}

// Remove unregisters and stops a server.
func (m *Manager) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.servers[name]
	if !ok {
		return fmt.Errorf("mcp: server %q not found", name)
	}
	delete(m.servers, name)
	return c.Close()
}

// Start connects to an installed server. Requires the server to
// already be installed (explicit enable) and enabled in its config.
func (m *Manager) Start(ctx context.Context, name string) error {
	m.mu.Lock()
	c, ok := m.servers[name]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("mcp: server %q not installed", name)
	}
	return c.Connect(ctx)
}

// ListServers returns the names of installed servers.
func (m *Manager) ListServers() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	names := make([]string, 0, len(m.servers))
	for n := range m.servers {
		names = append(names, n)
	}
	return names
}

// ListTools returns all tools from all started servers.
// Each tool is prefixed: mcp__<server>__<tool>.
func (m *Manager) ListTools(ctx context.Context) ([]Tool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var all []Tool
	for name, c := range m.servers {
		tools, err := c.ListTools(ctx)
		if err != nil {
			continue
		}
		for _, t := range tools {
			t.ServerName = name
			t.Name = prefixedName(name, t.Name)
			all = append(all, t)
		}
	}
	return all, nil
}

// CallTool executes a prefixed tool (mcp__<server>__<tool>) through
// the Gatekeeper. Parses the prefix to route to the correct server.
func (m *Manager) CallTool(ctx context.Context, prefixedName string, args map[string]any) (*ToolCallResult, error) {
	server, tool := parsePrefixed(prefixedName)
	if server == "" {
		return nil, fmt.Errorf("mcp: invalid prefixed name %q", prefixedName)
	}
	m.mu.Lock()
	c, ok := m.servers[server]
	m.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("mcp: server %q not installed", server)
	}
	return c.CallTool(ctx, tool, args)
}

// Close stops all servers.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.servers {
		_ = c.Close()
	}
	return nil
}

const prefixSep = "__"

// prefixedName creates the collision-free prefixed name.
func prefixedName(server, tool string) string {
	return "mcp" + prefixSep + server + prefixSep + tool
}

// parsePrefixed splits a prefixed name into server and tool.
func parsePrefixed(name string) (server, tool string) {
	prefix := "mcp" + prefixSep
	if !strings.HasPrefix(name, prefix) {
		return "", ""
	}
	rest := name[len(prefix):]
	idx := strings.Index(rest, prefixSep)
	if idx < 0 {
		return "", ""
	}
	return rest[:idx], rest[idx+len(prefixSep):]
}
