package daemon

import (
	"context"
	"testing"
)

// TestGate_DenyReturnsIPCError pins the deny contract: when the
// Gatekeeper denies the action, gate returns an ipc.Error with
// the standard CodeInternalError + "denied by safety policy"
// message. This is the exact contract every gated IPC method
// used to inline; pinning it here means future gate call-sites
// can trust the helper.
//
// Uses a nil-Engine setup to force the deny path — the
// Gatekeeper is fail-closed, so a nil Engine means "deny",
// which is exactly what the production handlers relied on.
func TestGate_DenyReturnsIPCError(t *testing.T) {
	subs := &Subsystems{Safety: &SafetyComponents{Engine: nil}}
	err := gate(context.Background(), subs, "test.action", "test detail")
	if err == nil {
		t.Fatal("gate(deny) = nil; want ipc.Error")
	}
	msg := err.Error()
	if msg != "rpc error -32603: denied by safety policy" {
		t.Errorf("gate(deny) message = %q, want %q (matches the inline contract)", msg, "rpc error -32603: denied by safety policy")
	}
}

// TestGate_NilSubsFailsClosed pins the fail-closed contract:
// when Subsystems is nil (the gate is called before the daemon
// is fully wired), gate must return an error (deny) rather
// than panicking.
//
// The previous inline pattern (`subs.GatekeeperAllow`) had
// the same fail-closed behavior: nil receiver returns false
// from the method's nil-check, which produces the deny error.
// gate() must preserve that contract.
func TestGate_NilSubsFailsClosed(t *testing.T) {
	var subs *Subsystems
	err := gate(context.Background(), subs, "test.action", "test detail")
	if err == nil {
		t.Fatal("gate(nil subs) = nil; want deny error (fail-closed)")
	}
}

// TestGate_NilSafetyFailsClosed pins the deeper nil-guard:
// even with a non-nil Subsystems, a nil Safety subsystem
// must deny (not panic). This matches the production
// GatekeeperAllow behavior: it checks s.Safety != nil before
// calling s.Safety.Engine.
func TestGate_NilSafetyFailsClosed(t *testing.T) {
	subs := &Subsystems{Safety: nil}
	err := gate(context.Background(), subs, "test.action", "test detail")
	if err == nil {
		t.Fatal("gate(nil Safety) = nil; want deny error")
	}
}

// TestGate_NilEngineFailsClosed pins the deepest nil-guard:
// Subsystems with non-nil Safety but nil Engine must deny.
// The Safety subsystem is wired during initSubsystems; if the
// engine failed to construct, every gated action must fail
// closed rather than silently allowing.
func TestGate_NilEngineFailsClosed(t *testing.T) {
	subs := &Subsystems{
		Safety: &SafetyComponents{Engine: nil},
	}
	err := gate(context.Background(), subs, "test.action", "test detail")
	if err == nil {
		t.Fatal("gate(nil Engine) = nil; want deny error (fail-closed)")
	}
}

// TestGate_ErrorMatchesInlineContract is the regression-prevention
// test: gate must return EXACTLY the same ipc.Error that the
// old inline pattern returned. If a future refactor changes the
// error code or message, this test catches it — the GUI and
// audit log both parse this string.
func TestGate_ErrorMatchesInlineContract(t *testing.T) {
	subs := &Subsystems{Safety: &SafetyComponents{Engine: nil}}
	err := gate(context.Background(), subs, "any.action", "any detail")
	if err == nil {
		t.Fatal("gate returned nil; want error")
	}
	if got, want := err.Error(), "rpc error -32603: denied by safety policy"; got != want {
		t.Errorf("error message drift: got %q, want %q", got, want)
	}
}
