package daemon

import (
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/overlay"
)

// TestSubsystems_SetOverlay_NilRevertsToNoop verifies that
// SetOverlay(nil) restores the headless noop controller. This
// is the shutdown/test path: the GUI host swaps in the Wails
// controller at startup; on test reset we want the noop back.
func TestSubsystems_SetOverlay_NilRevertsToNoop(t *testing.T) {
	s := &Subsystems{}
	s.SetOverlay(nil)
	if s.Overlay == nil {
		t.Fatal("SetOverlay(nil) left Overlay nil; want noop controller")
	}
	// The noop starts in StateHidden; that's the canonical
	// post-init state.
	if s.Overlay.State() != overlay.StateHidden {
		t.Errorf("after SetOverlay(nil): State = %v, want StateHidden",
			s.Overlay.State())
	}
}

// TestSubsystems_SetOverlay_NilReceiverIsNoop verifies the
// method is safe to call on a nil *Subsystems. The contract
// says no-op when s is nil; callers in the GUI host chain
// a.b.c.SetOverlay(...) without checking each step.
func TestSubsystems_SetOverlay_NilReceiverIsNoop(t *testing.T) {
	var s *Subsystems
	// Must not panic.
	s.SetOverlay(nil)
}

// TestSubsystems_SetOverlay_SwapsController verifies that
// SetOverlay replaces the existing controller and the new
// one is the one returned via subsequent calls.
func TestSubsystems_SetOverlay_SwapsController(t *testing.T) {
	s := &Subsystems{}
	c1 := overlay.NewNoopController()
	c2 := overlay.NewNoopController()

	s.SetOverlay(c1)
	if s.Overlay != c1 {
		t.Errorf("after first SetOverlay: Overlay = %v, want c1", s.Overlay)
	}

	s.SetOverlay(c2)
	if s.Overlay != c2 {
		t.Errorf("after second SetOverlay: Overlay = %v, want c2", s.Overlay)
	}
	if s.Overlay == c1 {
		t.Errorf("second SetOverlay did not replace c1")
	}
}
