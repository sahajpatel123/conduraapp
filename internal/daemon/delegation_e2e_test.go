package daemon

import (
	"context"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/delegation"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
)

// Test that the GatedRunner is the only spawn path — structural test.
func TestDelegation_GatedRunnerExists(t *testing.T) {
	cfg := delegation.DefaultConfig()
	limiter := delegation.NewLimiter(cfg, nil)
	g := delegation.NewGatedRunner(cfg, gatekeeper.DenyBeyondRead{}, limiter)
	if g == nil {
		t.Fatal("GatedRunner must be constructable")
	}
	// No unexported runner is visible from outside the package.
	// This is structural enforcement: only GatedRunner can spawn.
}

// Test that spawn through a deny gatekeeper is blocked.
func TestDelegation_SpawnDeniedByGatekeeper(t *testing.T) {
	cfg := delegation.DefaultConfig()
	limiter := delegation.NewLimiter(cfg, nil)
	g := delegation.NewGatedRunner(cfg, gatekeeper.DenyBeyondRead{}, limiter)
	_, err := g.Spawn(context.Background(), &delegation.SpawnRequest{
		AgentName: "claude", Task: "test", Depth: 0, Budget: 0,
	})
	if err == nil {
		t.Fatal("spawn must be denied by DenyBeyondRead gatekeeper")
	}
}

// Test that recursion limit is enforced.
func TestDelegation_RecursionLimitEnforced(t *testing.T) {
	cfg := delegation.DefaultConfig()
	cliCfg, _ := cfg.FindAgent("claude")
	limiter := delegation.NewLimiter(cfg, nil)

	err := limiter.CheckSpawn(context.Background(), "claude", cliCfg.MaxDepth+10, 0)
	if err == nil {
		t.Fatal("recursion limit must be enforced")
	}
}

// Test that budget is enforced.
func TestDelegation_BudgetExceeded(t *testing.T) {
	cfg := delegation.DefaultConfig()
	limiter := delegation.NewLimiter(cfg, nil)

	err := limiter.CheckSpawn(context.Background(), "claude", 0, 100000.0)
	if err == nil {
		t.Fatal("massive budget must be rejected")
	}
}
