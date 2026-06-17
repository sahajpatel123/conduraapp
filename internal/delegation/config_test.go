package delegation

import (
	"testing"
	"time"
)

func TestDefaultAgents_Count(t *testing.T) {
	agents := DefaultAgents()
	if len(agents) != 8 {
		t.Fatalf("DefaultAgents() len = %d, want 8", len(agents))
	}
}

func TestDefaultAgents_AllPresent(t *testing.T) {
	want := []string{
		"claude", "codex", "antigravity", "opencode",
		"kilo", "hermes", "gemini", "ollama",
	}
	seen := make(map[string]AgentConfig, len(want))
	for _, a := range DefaultAgents() {
		seen[a.Name] = a
	}
	for _, name := range want {
		a, ok := seen[name]
		if !ok {
			t.Fatalf("agent %q missing from DefaultAgents()", name)
		}
		if a.Name == "" {
			t.Fatalf("agent %q has empty Name", name)
		}
		if a.BinaryProbe == "" && a.Command != "" {
			t.Fatalf("agent %q has empty BinaryProbe", name)
		}
		if a.MaxDepth < 1 {
			t.Fatalf("agent %q MaxDepth = %d, want >= 1", name, a.MaxDepth)
		}
		if a.Timeout != 5*time.Minute {
			t.Fatalf("agent %q Timeout = %v, want 5m", name, a.Timeout)
		}
	}
}

func TestDefaultAgents_SubprocessCLIs(t *testing.T) {
	cfg := DefaultConfig()
	for _, name := range []string{"codex", "antigravity", "opencode", "kilo", "hermes", "gemini"} {
		a, ok := cfg.FindAgent(name)
		if !ok {
			t.Fatalf("FindAgent(%q) = false", name)
		}
		if a.Command == "" {
			t.Fatalf("agent %q Command is empty", name)
		}
		if a.OutputFormat == "" {
			t.Fatalf("agent %q OutputFormat is empty", name)
		}
	}
}

func TestDefaultAgents_OllamaNoSubprocess(t *testing.T) {
	a, ok := DefaultConfig().FindAgent("ollama")
	if !ok {
		t.Fatal("ollama not found")
	}
	if a.Command != "" {
		t.Fatalf("ollama Command = %q, want empty (HTTP-only)", a.Command)
	}
}
