package daemon

import (
	"context"
	"time"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/halt"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/sse"
)

// guardAwareHaltFlag wraps *halt.Flag so that any Halt() call —
// including the watchdog's auto-halt, which calls HaltFlag.Halt
// directly and bypasses the daemon.halt RPC handler — also toggles
// the network guard (Layer 3 of the kill switch) and publishes an
// SSE "halt" event so the Meridian overlay updates without a 5s poll.
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
	flag   *halt.Flag
	guard  halt.NetworkGuard
	broker *sse.Broker
}

func (g guardAwareHaltFlag) Halt(ctx context.Context, reason string) (halt.State, error) {
	st, err := g.flag.Halt(ctx, reason)
	if g.guard != nil {
		_ = g.guard.Halt(reason)
	}
	if err == nil {
		publishHaltSSE(g.broker, g.flag.Halted())
	}
	return st, err
}

func (g guardAwareHaltFlag) IsHalted() bool {
	return g.flag.IsHalted()
}

// publishHaltSSE notifies connected GUIs of the kill-switch state.
// Event name "halt" matches the frontend ipc EventMap + EventSource
// named-event list. No-op when broker is nil (unit tests).
func publishHaltSSE(broker *sse.Broker, s halt.State) {
	if broker == nil {
		return
	}
	payload := map[string]any{
		syncHaltedKey: s.Halted,
		syncReasonKey: s.Reason,
	}
	if s.Halted && !s.Since.IsZero() {
		payload["since"] = s.Since.UTC().Format(time.RFC3339)
	}
	broker.PublishJSON("halt", payload)
}

// rearmNetGuardIfHalted applies Layer 3 isolation when the sticky halt
// flag was restored from disk after a crash/restart. Without this,
// NewInProcessGuard() starts open and LLM egress would work while
// haltFlag.IsHalted() is still true — a Survival Rule split-brain.
func rearmNetGuardIfHalted(haltFlag *halt.Flag, guard halt.NetworkGuard, log interface {
	Info(string, ...any)
	Warn(string, ...any)
}) {
	if haltFlag == nil || guard == nil || !haltFlag.IsHalted() {
		return
	}
	st := haltFlag.Halted()
	reason := st.Reason
	if reason == "" {
		reason = "restored halt after restart"
	}
	if err := guard.Halt(reason); err != nil {
		if log != nil {
			log.Warn("net guard re-arm failed after halt restore", "err", err)
		}
		return
	}
	if log != nil {
		log.Info("net guard re-armed (halt flag restored from disk)", "reason", reason)
	}
}
