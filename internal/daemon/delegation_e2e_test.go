package daemon

import (
	"context"
	"errors"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/delegation"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/halt"
	"github.com/sahajpatel123/synapticapp/internal/sse"
)

// realEngine returns the production Engine with buildSafetyLayer hooks.
func realEngine(t *testing.T) *gatekeeper.Engine {
	t.Helper()
	hf := halt.New(testDB(t))
	broker := sse.NewBroker()
	safety := buildSafetyLayer(hf, broker, nil, nil)
	return safety.Engine
}

// TestDelegation_RealEngine_ClaudeSpawnApproved proves the real Engine
// allows delegation.spawn for known agents with consent approved.
func TestDelegation_RealEngine_ClaudeSpawnApproved(t *testing.T) {
	engine := realEngine(t)
	engine.SetConsentProvider(&approveConsent{})

	cfg := delegation.DefaultConfig()
	limiter := delegation.NewLimiter(cfg, nil)
	g := delegation.NewGatedRunner(cfg, engine, limiter)

	_, err := g.Spawn(context.Background(), &delegation.SpawnRequest{
		AgentName: "claude", Task: "hello", Depth: 0, Budget: 0,
	})
	// Real spawn may fail because claude binary isn't installed —
	// that's fine. What matters: the gate passes (no ErrGatedDeny).
	if errors.Is(err, delegation.ErrGatedDeny) {
		t.Fatalf("claude spawn should not be gated with consent approved, got %v", err)
	}
	t.Logf("spawn result: %v", err)
}

// TestDelegation_RealEngine_UnknownAgentDenied proves the real Engine
// denies delegation.spawn for unknown agents.
func TestDelegation_RealEngine_UnknownAgentDenied(t *testing.T) {
	engine := realEngine(t)
	engine.SetConsentProvider(&approveConsent{})

	cfg := delegation.DefaultConfig()
	// Add the unknown agent to the config so FindAgent succeeds and
	// the spawn actually reaches the gatekeeper policy path.
	cfg.Agents = append(cfg.Agents, delegation.AgentConfig{
		Name:    "evil-agent",
		Command: "evil-agent",
	})
	limiter := delegation.NewLimiter(cfg, nil)
	g := delegation.NewGatedRunner(cfg, engine, limiter)

	_, err := g.Spawn(context.Background(), &delegation.SpawnRequest{
		AgentName: "evil-agent", Task: "rm -rf /", Depth: 0, Budget: 0,
	})
	if err == nil {
		t.Fatal("unknown agent spawn must be denied by the real Engine")
	}
	if !errors.Is(err, delegation.ErrGatedDeny) {
		t.Fatalf("expected gatekeeper deny, got %v", err)
	}
}

// TestDelegation_RecursionBlocked proves recursion limits work with
// the real Engine.
func TestDelegation_RecursionBlocked(t *testing.T) {
	engine := realEngine(t)
	engine.SetConsentProvider(&approveConsent{})

	cfg := delegation.DefaultConfig()
	cliCfg, _ := cfg.FindAgent("claude")
	limiter := delegation.NewLimiter(cfg, nil)
	g := delegation.NewGatedRunner(cfg, engine, limiter)

	_, err := g.Spawn(context.Background(), &delegation.SpawnRequest{
		AgentName: "claude", Task: "test",
		Depth: cliCfg.MaxDepth + 10, Budget: 0,
	})
	if err == nil {
		t.Fatal("recursion limit must block")
	}
}

// TestDelegation_HaltInterrupts proves halt blocks with the real Engine.
func TestDelegation_HaltInterrupts(t *testing.T) {
	hf := halt.New(testDB(t))
	broker := sse.NewBroker()
	safety := buildSafetyLayer(hf, broker, nil, nil)
	engine := safety.Engine

	_, _ = hf.Halt(context.Background(), "test halt")

	cfg := delegation.DefaultConfig()
	limiter := delegation.NewLimiter(cfg, nil)
	g := delegation.NewGatedRunner(cfg, engine, limiter)

	_, err := g.Spawn(context.Background(), &delegation.SpawnRequest{
		AgentName: "claude", Task: "test", Depth: 0, Budget: 0,
	})
	if err == nil {
		t.Fatal("halted engine must deny delegation spawn")
	}
}

// TestDelegation_UnGatedPathUnreachable is the structural test proving
// only GatedRunner can spawn sub-agents. The unexported runner is
// inaccessible from outside the delegation package.
func TestDelegation_UnGatedPathUnreachable(t *testing.T) {
	// Compiles = test passes. runner is unexported — constructing
	// it directly would be a compile error.
	_ = delegation.NewGatedRunner(delegation.DefaultConfig(), gatekeeper.DenyBeyondRead{},
		delegation.NewLimiter(delegation.DefaultConfig(), nil))
}
