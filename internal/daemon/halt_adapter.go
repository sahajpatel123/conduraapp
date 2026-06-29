package daemon

import (
	"context"

	"github.com/sahajpatel123/conduraapp/internal/halt"
)

// guardAwareHaltFlag wraps *halt.Flag so that any Halt() call —
// including the watchdog's auto-halt, which calls HaltFlag.Halt
// directly and bypasses the daemon.halt RPC handler — also toggles
// the network guard (Layer 3 of the kill switch).
//
// N3 completeness: the daemon.halt RPC handler (methods_phase2.go)
// was wired to toggle the guard, but the watchdog halts via
// haltFlag.Halt() directly, so a watchdog auto-halt set the flag
// while leaving outbound LLM egress open. This adapter closes that
// gap: every halt path (RPC + watchdog) now isolates the network.
// In-flight LLM requests fail on their next RoundTrip because the
// guarded transport denies all hosts when halted.
//
// It satisfies watchdog.HaltFlag (Halt + IsHalted). The guard may be
// nil (tests/headless without a guard); Halt then behaves as the raw
// flag.
type guardAwareHaltFlag struct {
	flag  *halt.Flag
	guard halt.NetworkGuard
}

func (g guardAwareHaltFlag) Halt(ctx context.Context, reason string) (halt.State, error) {
	st, err := g.flag.Halt(ctx, reason)
	if g.guard != nil {
		_ = g.guard.Halt(reason)
	}
	return st, err
}

func (g guardAwareHaltFlag) IsHalted() bool {
	return g.flag.IsHalted()
}
