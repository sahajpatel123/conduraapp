package mcp

import (
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
)

func TestPrefixedName(t *testing.T) {
	tests := []struct{ server, tool, want string }{
		{"filesystem", "read_file", "mcp__filesystem__read_file"},
		{"github", "create_issue", "mcp__github__create_issue"},
		{"a", "b", "mcp__a__b"},
	}
	for _, tt := range tests {
		got := prefixedName(tt.server, tt.tool)
		if got != tt.want {
			t.Errorf("prefixedName(%q, %q) = %q, want %q", tt.server, tt.tool, got, tt.want)
		}
	}
}

func TestParsePrefixed(t *testing.T) {
	tests := []struct {
		input      string
		wantServer string
		wantTool   string
	}{
		{"mcp__filesystem__read_file", "filesystem", "read_file"},
		{"mcp__github__create_issue", "github", "create_issue"},
		{"mcp__a__b", "a", "b"},
		{"invalid", "", ""},
		{"mcp__", "", ""},
		{"mcp__a", "", ""},
	}
	for _, tt := range tests {
		server, tool := parsePrefixed(tt.input)
		if server != tt.wantServer || tool != tt.wantTool {
			t.Errorf("parsePrefixed(%q) = (%q, %q), want (%q, %q)", tt.input, server, tool, tt.wantServer, tt.wantTool)
		}
	}
}

func TestManager_Install(t *testing.T) {
	m := NewManager(gatekeeper.DenyBeyondRead{})
	if err := m.Install(ServerConfig{Name: "test", Command: "echo", Type: TransportStdio}); err != nil {
		t.Fatal(err)
	}
	if len(m.ListServers()) != 1 {
		t.Error("expected 1 server")
	}
}

func TestManager_DuplicateInstall(t *testing.T) {
	m := NewManager(gatekeeper.DenyBeyondRead{})
	_ = m.Install(ServerConfig{Name: "test", Command: "echo", Type: TransportStdio})
	if err := m.Install(ServerConfig{Name: "test", Command: "cat", Type: TransportStdio}); err == nil {
		t.Fatal("expected error on duplicate")
	}
}

func TestManager_Remove(t *testing.T) {
	m := NewManager(gatekeeper.DenyBeyondRead{})
	_ = m.Install(ServerConfig{Name: "test", Command: "echo", Type: TransportStdio})
	if err := m.Remove("test"); err != nil {
		t.Fatal(err)
	}
	if len(m.ListServers()) != 0 {
		t.Error("expected 0 servers")
	}
}

func TestClient_Gated(t *testing.T) {
	// GatedClient creation is tested. Actual subprocess connection
	// requires integration tests with a real MCP server binary.
	cfg := ServerConfig{Name: "test", Command: "echo", Type: TransportStdio}
	gc := NewGatedClient(cfg, gatekeeper.DenyBeyondRead{})
	if gc == nil {
		t.Fatal("expected non-nil client")
	}
	_ = gc.Close()
}
