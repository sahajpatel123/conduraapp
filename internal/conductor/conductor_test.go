package conductor

import (
	"sync/atomic"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/hotkey"
	"github.com/sahajpatel123/conduraapp/internal/overlay"
	"github.com/sahajpatel123/conduraapp/internal/presence"
)

// dummyHaltChecker implements presence.HaltChecker.
type dummyHaltChecker struct{ halted bool }

func (d *dummyHaltChecker) IsHalted() bool { return d.halted }

// newTestConductor creates a Conductor that doesn't register a real
// hotkey (we test the toggle logic in isolation).
func newTestConductor(t *testing.T) (*Conductor, overlay.Controller, *dummyHaltChecker) {
	t.Helper()
	ctrl := overlay.NewNoopController()
	halt := &dummyHaltChecker{}
	orch := presence.NewOrchestrator(ctrl, halt, nil)
	// We pass nil for hotkey; the toggle function only uses
	// orchestrator and the onShow/onHide callbacks, never the
	// hotkey itself.
	cond := &Conductor{
		hotkey:       nil,
		orchestrator: orch,
	}
	return cond, ctrl, halt
}

func TestNew_RequiresHotkey(t *testing.T) {
	_, err := New(nil, presence.NewOrchestrator(overlay.NewNoopController(), &dummyHaltChecker{}, nil), nil, nil)
	if err == nil {
		t.Fatal("expected error for nil hotkey")
	}
}

func TestNew_RequiresOrchestrator(t *testing.T) {
	// Construct a real hotkey.Manager without starting it.
	hk := hotkey.New("Cmd+Shift+Space")
	_, err := New(hk, nil, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil orchestrator")
	}
}

func TestNew_OK(t *testing.T) {
	hk := hotkey.New("Cmd+Shift+Space")
	orch := presence.NewOrchestrator(overlay.NewNoopController(), &dummyHaltChecker{}, nil)
	cond, err := New(hk, orch, nil, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if cond == nil {
		t.Fatal("nil conductor")
	}
}

// TestToggle_SummonThenDismiss verifies that the conductor's toggle
// flips the orchestrator between active and inactive.
func TestToggle_SummonThenDismiss(t *testing.T) {
	cond, ctrl, _ := newTestConductor(t)

	// First press: summon.
	cond.toggle()
	if !cond.orchestrator.IsActive() {
		t.Fatal("after first toggle, orchestrator should be active")
	}
	if ctrl.State() == overlay.StateHidden {
		t.Fatal("overlay should not be hidden after summon")
	}

	// Second press: dismiss.
	cond.toggle()
	if cond.orchestrator.IsActive() {
		t.Fatal("after second toggle, orchestrator should be inactive")
	}
	if ctrl.State() != overlay.StateHidden {
		t.Fatal("overlay should be hidden after dismiss")
	}
}

// TestToggle_BlockedByHalt verifies the kill switch blocks summon.
func TestToggle_BlockedByHalt(t *testing.T) {
	cond, ctrl, halt := newTestConductor(t)
	halt.halted = true

	cond.toggle()
	if cond.orchestrator.IsActive() {
		t.Fatal("orchestrator should not activate when halted")
	}
	if ctrl.State() != overlay.StateHidden {
		t.Fatal("overlay should remain hidden when halted")
	}
}

// TestToggle_FiresCallbacks verifies onShow/onHide fire exactly once per toggle.
func TestToggle_FiresCallbacks(t *testing.T) {
	ctrl := overlay.NewNoopController()
	halt := &dummyHaltChecker{}
	orch := presence.NewOrchestrator(ctrl, halt, nil)
	var shows, hides atomic.Int64
	cond := &Conductor{
		hotkey:       nil,
		orchestrator: orch,
		onShow:       func() { shows.Add(1) },
		onHide:       func() { hides.Add(1) },
	}

	cond.toggle()
	cond.toggle()
	cond.toggle() // back to active

	if shows.Load() != 2 {
		t.Errorf("shows = %d, want 2", shows.Load())
	}
	if hides.Load() != 1 {
		t.Errorf("hides = %d, want 1", hides.Load())
	}
}

// TestState_DelegatesToOrchestrator verifies State() returns the
// overlay state through the orchestrator.
func TestState_DelegatesToOrchestrator(t *testing.T) {
	cond, _, _ := newTestConductor(t)
	if got := cond.State(); got != overlay.StateHidden {
		t.Errorf("State() = %v, want hidden", got)
	}
}

// TestStop_Idempotent verifies Stop on an unstarted conductor is safe.
func TestStop_Idempotent(t *testing.T) {
	cond, _, _ := newTestConductor(t)
	cond.Stop() // no-op
	cond.Stop() // still no-op
}

// TestSummon_AfterHaltResumesReactivated verifies that dismissing
// the halt flag re-enables summon.
func TestSummon_AfterHaltResumesReactivated(t *testing.T) {
	cond, _, halt := newTestConductor(t)
	halt.halted = true
	cond.toggle()
	if cond.orchestrator.IsActive() {
		t.Fatal("halted: should not activate")
	}
	halt.halted = false
	cond.toggle()
	if !cond.orchestrator.IsActive() {
		t.Fatal("un-halted: should activate")
	}
}
