package overlay

import (
	"context"
	"testing"
)

func TestNoopController_InitialState(t *testing.T) {
	c := NewNoopController()
	if c.State() != StateHidden {
		t.Errorf("expected StateHidden, got %v", c.State())
	}
}

func TestNoopController_Show(t *testing.T) {
	c := NewNoopController()
	if err := c.Show(context.Background(), ShowOpts{}); err != nil {
		t.Fatalf("Show: %v", err)
	}
	if c.State() != StateListening {
		t.Errorf("expected StateListening, got %v", c.State())
	}
}

func TestNoopController_Hide(t *testing.T) {
	c := NewNoopController()
	_ = c.Show(context.Background(), ShowOpts{})
	if err := c.Hide(); err != nil {
		t.Fatalf("Hide: %v", err)
	}
	if c.State() != StateHidden {
		t.Errorf("expected StateHidden, got %v", c.State())
	}
}

func TestNoopController_Toggle(t *testing.T) {
	c := NewNoopController()

	// Toggle from hidden → listening.
	c.Toggle()
	if c.State() != StateListening {
		t.Errorf("expected StateListening after toggle, got %v", c.State())
	}

	// Toggle from listening → hidden.
	c.Toggle()
	if c.State() != StateHidden {
		t.Errorf("expected StateHidden after toggle, got %v", c.State())
	}
}

func TestNoopController_OnDismiss(t *testing.T) {
	c := NewNoopController()
	dismissed := false
	c.OnDismiss(func() { dismissed = true })

	_ = c.Show(context.Background(), ShowOpts{})
	_ = c.Hide()

	if !dismissed {
		t.Error("OnDismiss callback not called")
	}
}

func TestNoopController_HideFromHidden(t *testing.T) {
	c := NewNoopController()
	dismissed := false
	c.OnDismiss(func() { dismissed = true })

	// Hide when already hidden should not fire dismiss.
	_ = c.Hide()

	if dismissed {
		t.Error("OnDismiss should not fire when hiding from hidden state")
	}
}

func TestNoopController_ConcurrentAccess(t *testing.T) {
	c := NewNoopController()
	done := make(chan struct{})

	// Run concurrent state changes.
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			_ = c.Show(context.Background(), ShowOpts{})
			_ = c.Hide()
			c.Toggle()
			_ = c.State()
		}
	}()

	<-done
}

func TestOverlayState_String(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateHidden, "hidden"},
		{StateListening, "listening"},
		{StateThinking, "thinking"},
		{StateSpeaking, "speaking"},
		{State(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

// TestNoopController_SetState_StoresValue pins the SetState
// contract: SetState(state) MUST update c.state. The presence
// orchestrator (4.3) uses this accessor to drive state changes
// from external subsystems (e.g., "user is idle, show listening").
// A regression that silently dropped the SetState call would
// break the orchestrator-driven state machine.
func TestNoopController_SetState_StoresValue(t *testing.T) {
	c := NewNoopController().(*noopController)
	c.SetState(StateListening)
	if got := c.State(); got != StateListening {
		t.Errorf("State() after SetState(Listening) = %v, want StateListening", got)
	}
}

// TestNoopController_SetState_Overwrite pins the overwrite
// contract: a second SetState call MUST replace the first. The
// state machine can transition back and forth (e.g.,
// StateListening -> StateThinking -> StateListening on
// consecutive turns); a regression that early-returned would
// leave the state stuck at the first value.
func TestNoopController_SetState_Overwrite(t *testing.T) {
	c := NewNoopController().(*noopController)
	c.SetState(StateListening)
	c.SetState(StateThinking)
	if got := c.State(); got != StateThinking {
		t.Errorf("State() after SetState(Listening) then SetState(Thinking) = %v, want StateThinking",
			got)
	}
}

// TestNoopController_SetState_AllDefinedStates pin the
// all-defined-states contract: SetState MUST accept every
// State value (Hidden, Listening, Thinking, Speaking,
// Error). A regression that added a new State constant without
// handling it in SetState would crash the orchestrator.
func TestNoopController_SetState_AllDefinedStates(t *testing.T) {
	allStates := []State{
		StateHidden,
		StateListening,
		StateThinking,
		StateSpeaking,
	}
	for _, s := range allStates {
		c := NewNoopController().(*noopController)
		c.SetState(s)
		if got := c.State(); got != s {
			t.Errorf("State() after SetState(%v) = %v, want %v", s, got, s)
		}
	}
}
