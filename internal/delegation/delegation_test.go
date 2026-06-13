package delegation

import (
	"context"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
)

func TestConfig_FindAgent(t *testing.T) {
	cfg := DefaultConfig()
	a, ok := cfg.FindAgent("claude")
	if !ok {
		t.Fatal("claude not found")
	}
	if a.Name != "claude" {
		t.Errorf("name = %q", a.Name)
	}
	if a.Command != "claude" {
		t.Errorf("command = %q", a.Command)
	}
}

func TestConfig_FindAgent_NotFound(t *testing.T) {
	cfg := DefaultConfig()
	_, ok := cfg.FindAgent("nonexistent")
	if ok {
		t.Fatal("should not find nonexistent agent")
	}
}

func TestSemaphore_AcquireRelease(t *testing.T) {
	sm := NewSemaphoreManager(2, 3)
	if err := sm.Acquire(context.Background(), "claude"); err != nil {
		t.Fatal(err)
	}
	if sm.Available() != 2 {
		t.Errorf("available = %d, want 2", sm.Available())
	}
	sm.Release("claude")
	if sm.Available() != 3 {
		t.Errorf("available = %d, want 3", sm.Available())
	}
}

func TestSemaphore_GlobalLimitBlocks(t *testing.T) {
	sm := NewSemaphoreManager(5, 1)
	_ = sm.Acquire(context.Background(), "claude")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := sm.Acquire(ctx, "codex")
	if err == nil {
		t.Fatal("expected timeout on second acquire")
	}
}

func TestSemaphore_PerAgentLimit(t *testing.T) {
	sm := NewSemaphoreManager(1, 5)
	_ = sm.Acquire(context.Background(), "claude")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := sm.Acquire(ctx, "claude")
	if err == nil {
		t.Fatal("expected timeout on per-agent limit")
	}
}

func TestLimiter_RecursionLimit(t *testing.T) {
	cfg := DefaultConfig()
	l := NewLimiter(cfg, nil)
	cliCfg, _ := cfg.FindAgent("claude")
	err := l.CheckSpawn(context.Background(), "claude", cliCfg.MaxDepth+1, 0)
	if err == nil {
		t.Fatal("expected recursion limit")
	}
}

func TestLimiter_AgentNotFound(t *testing.T) {
	cfg := DefaultConfig()
	l := NewLimiter(cfg, nil)
	err := l.CheckSpawn(context.Background(), "nonexistent", 0, 0)
	if err == nil {
		t.Fatal("expected agent not found")
	}
}

// TestGatekeeper is the only path to spawn. Verifies the structural
// enforcement: GatedRunner is the sole exported spawn path.
type allowGate struct{}

func (allowGate) Evaluate(_ context.Context, _ blastradius.Action) (gatekeeper.Decision, string) {
	return gatekeeper.Allow, "allowed"
}

type denyGate struct{}

func (denyGate) Evaluate(_ context.Context, _ blastradius.Action) (gatekeeper.Decision, string) {
	return gatekeeper.Deny, "denied"
}

func TestGatedRunner_SpawnDenied(t *testing.T) {
	cfg := DefaultConfig()
	l := NewLimiter(cfg, nil)
	g := NewGatedRunner(cfg, denyGate{}, l)
	_, err := g.Spawn(context.Background(), &SpawnRequest{
		AgentName: "claude", Task: "test", Depth: 0, Budget: 0,
	})
	if err == nil {
		t.Fatal("expected deny on spawn")
	}
}

func TestGatedRunner_AgentNotFound(t *testing.T) {
	cfg := DefaultConfig()
	l := NewLimiter(cfg, nil)
	g := NewGatedRunner(cfg, allowGate{}, l)
	_, err := g.Spawn(context.Background(), &SpawnRequest{
		AgentName: "nonexistent", Task: "test",
	})
	if err == nil {
		t.Fatal("expected agent not found")
	}
}

func TestGatedRunner_RecursionBlocked(t *testing.T) {
	cfg := DefaultConfig()
	cliCfg, _ := cfg.FindAgent("claude")
	l := NewLimiter(cfg, nil)
	g := NewGatedRunner(cfg, allowGate{}, l)
	_, err := g.Spawn(context.Background(), &SpawnRequest{
		AgentName: "claude", Task: "test",
		Depth: cliCfg.MaxDepth + 1, Budget: 0,
	})
	if err == nil {
		t.Fatal("expected recursion block")
	}
}

func TestGatedRunner_ActionRequests(t *testing.T) {
	cfg := DefaultConfig()
	l := NewLimiter(cfg, nil)
	g := NewGatedRunner(cfg, allowGate{}, l)

	// Stream-JSON output with an action request.
	output := `{"kind":"shell.exec","command":"echo hello", "agent_name":"claude"}` + "\n" +
		`{"kind":"click","body":"button"}` + "\n" +
		`{"not": "an action"}` + "\n"
	result := &SpawnResult{AgentName: "claude", Output: output}
	requests := g.ActionRequests(result)
	if len(requests) != 2 {
		t.Fatalf("got %d action requests, want 2", len(requests))
	}
	if requests[0].Kind != "shell.exec" {
		t.Errorf("kind = %q", requests[0].Kind)
	}
}

func TestBudget_LowLimitBlocks(t *testing.T) {
	cfg := DefaultConfig()
	l := NewLimiter(cfg, nil)
	err := l.CheckSpawn(context.Background(), "claude", 0, 1000.0)
	if err == nil {
		t.Fatal("expected budget exceeded")
	}
}
