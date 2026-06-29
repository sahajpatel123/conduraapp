package daemon

import (
	"context"
	"net/http"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/halt"
)

// fakeGuard is a minimal halt.NetworkGuard that records Halt/Resume
// calls. Used to assert that every halt path (RPC + watchdog via
// guardAwareHaltFlag) toggles the network guard — the N3 fix.
type fakeGuard struct {
	halted      bool
	haltCalls   int
	resumeCalls int
	lastReason  string
}

func (f *fakeGuard) Allow(string) bool                                    { return !f.halted }
func (f *fakeGuard) WrapTransport(rt http.RoundTripper) http.RoundTripper { return rt }
func (f *fakeGuard) Halt(reason string) error {
	f.halted = true
	f.haltCalls++
	f.lastReason = reason
	return nil
}
func (f *fakeGuard) Resume() error {
	f.halted = false
	f.resumeCalls++
	return nil
}
func (f *fakeGuard) State() halt.GuardState { return halt.GuardState{Halted: f.halted} }

// TestGuardAwareHaltFlag_HaltTogglesGuard verifies the N3-completeness
// fix: the watchdog's auto-halt (which calls HaltFlag.Halt directly,
// bypassing the daemon.halt RPC handler) still isolates the network
// because the adapter forwards Halt to the guard as well as the flag.
func TestGuardAwareHaltFlag_HaltTogglesGuard(t *testing.T) {
	hf := halt.New(testDB(t))
	g := &fakeGuard{}
	adapter := guardAwareHaltFlag{flag: hf, guard: g}

	st, err := adapter.Halt(context.Background(), "watchdog fired")
	if err != nil {
		t.Fatalf("Halt returned error: %v", err)
	}
	// st is the PREVIOUS state (halt.State is "state before this call"),
	// so st.Halted is false here — that's expected. The meaningful
	// assertion is that the flag is halted NOW and the guard was called.
	if st.Halted {
		t.Fatal("returned State should be the prior (non-halted) state")
	}
	if !hf.IsHalted() {
		t.Fatal("underlying flag should be halted")
	}
	if g.haltCalls != 1 {
		t.Fatalf("guard.Halt called %d times, want 1", g.haltCalls)
	}
	if g.lastReason != "watchdog fired" {
		t.Fatalf("guard reason = %q, want %q", g.lastReason, "watchdog fired")
	}
	if !g.halted {
		t.Fatal("guard should be halted (network isolated)")
	}
}

// TestGuardAwareHaltFlag_NilGuardSafe verifies a nil guard (tests /
// headless without a guard) does not panic and still halts the flag.
func TestGuardAwareHaltFlag_NilGuardSafe(t *testing.T) {
	hf := halt.New(testDB(t))
	adapter := guardAwareHaltFlag{flag: hf, guard: nil}
	if _, err := adapter.Halt(context.Background(), "x"); err != nil {
		t.Fatalf("Halt with nil guard errored: %v", err)
	}
	if !hf.IsHalted() {
		t.Fatal("flag should halt even when guard is nil")
	}
}

// TestGuardAwareHaltFlag_IsHaltedForwards verifies IsHalted forwards
// to the underlying flag (the watchdog polls IsHalted).
func TestGuardAwareHaltFlag_IsHaltedForwards(t *testing.T) {
	hf := halt.New(testDB(t))
	adapter := guardAwareHaltFlag{flag: hf, guard: &fakeGuard{}}
	if adapter.IsHalted() {
		t.Fatal("should not be halted before Halt")
	}
	if _, err := adapter.Halt(context.Background(), "x"); err != nil {
		t.Fatal(err)
	}
	if !adapter.IsHalted() {
		t.Fatal("should be halted after Halt")
	}
}
