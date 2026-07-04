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
