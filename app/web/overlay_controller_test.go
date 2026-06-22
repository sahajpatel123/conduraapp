package main

import (
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/overlay"
)

// The wailsController's state machine is identical to the
// noopController's; the only thing it adds is Wails runtime
// calls guarded on c.ctx != nil. The tests below exercise
// the state machine in isolation (no Wails context wired) so
// they pass even without a running Wails app.

func TestWailsController_InitialState(t *testing.T) {
	c := newWailsController()
	if c.State() != overlay.StateHidden {
		t.Errorf("expected StateHidden, got %v", c.State())
	}
}

func TestWailsController_Show_NoCtx_StillUpdatesState(t *testing.T) {
	c := newWailsController()
	// Without a Wails ctx, Show should NOT panic and should
	// still transition the in-memory state. The conductor
	// relies on this for the "hotkey pressed before Wails is
	// ready" race.
	if err := c.Show(nil, overlay.ShowOpts{}); err != nil {
		t.Fatalf("Show: %v", err)
	}
	if c.State() != overlay.StateListening {
		t.Errorf("expected StateListening, got %v", c.State())
	}
}

func TestWailsController_Hide_FromVisible(t *testing.T) {
	c := newWailsController()
	_ = c.Show(nil, overlay.ShowOpts{})
	if err := c.Hide(); err != nil {
		t.Fatalf("Hide: %v", err)
	}
	if c.State() != overlay.StateHidden {
		t.Errorf("expected StateHidden, got %v", c.State())
	}
}

func TestWailsController_Toggle_RoundTrip(t *testing.T) {
	c := newWailsController()

	// Hidden → Listening.
	c.Toggle()
	if c.State() != overlay.StateListening {
		t.Errorf("after first toggle: expected StateListening, got %v", c.State())
	}

	// Listening → Hidden.
	c.Toggle()
	if c.State() != overlay.StateHidden {
		t.Errorf("after second toggle: expected StateHidden, got %v", c.State())
	}
}

func TestWailsController_OnDismiss_FiresOnHide(t *testing.T) {
	c := newWailsController()
	fired := false
	c.OnDismiss(func() { fired = true })
	_ = c.Show(nil, overlay.ShowOpts{})
	_ = c.Hide()
	if !fired {
		t.Errorf("OnDismiss callback should fire on Hide from non-hidden state")
	}
}

func TestWailsController_OnDismiss_DoesNotFireFromHidden(t *testing.T) {
	c := newWailsController()
	fired := false
	c.OnDismiss(func() { fired = true })
	_ = c.Hide() // already hidden
	if fired {
		t.Errorf("OnDismiss callback should NOT fire when already hidden")
	}
}

func TestWailsController_SetState(t *testing.T) {
	c := newWailsController()
	c.SetState(overlay.StateThinking)
	if c.State() != overlay.StateThinking {
		t.Errorf("expected StateThinking, got %v", c.State())
	}
}

func TestWailsController_DefaultDimensions(t *testing.T) {
	c := newWailsController()
	// The overlay dimensions are part of the UX contract;
	// changing them changes the visual size of the hotkey
	// popup. Pin them so the regression is loud.
	if c.overlayWidth != 620 || c.overlayHeight != 88 {
		t.Errorf("overlay dimensions = %dx%d, want 620x88",
			c.overlayWidth, c.overlayHeight)
	}
}
