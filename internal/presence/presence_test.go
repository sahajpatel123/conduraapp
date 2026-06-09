package presence

import (
	"context"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/overlay"
)

func TestOrchestrator_SummonAndDismiss(t *testing.T) {
	ctrl := overlay.NewNoopController()
	o := NewOrchestrator(ctrl, nil)

	if err := o.Summon(context.Background()); err != nil {
		t.Fatalf("Summon: %v", err)
	}
	if !o.IsActive() {
		t.Error("expected active after Summon")
	}
	if o.State() != overlay.StateListening {
		t.Errorf("expected StateListening, got %v", o.State())
	}

	o.Dismiss()
	if o.IsActive() {
		t.Error("expected inactive after Dismiss")
	}
	if o.State() != overlay.StateHidden {
		t.Errorf("expected StateHidden, got %v", o.State())
	}
}

func TestOrchestrator_SummonIdempotent(t *testing.T) {
	ctrl := overlay.NewNoopController()
	o := NewOrchestrator(ctrl, nil)

	_ = o.Summon(context.Background())
	// Second summon should be a no-op.
	if err := o.Summon(context.Background()); err != nil {
		t.Fatalf("second Summon: %v", err)
	}
	if !o.IsActive() {
		t.Error("expected still active")
	}
}

func TestOrchestrator_DismissWhenInactive(t *testing.T) {
	ctrl := overlay.NewNoopController()
	o := NewOrchestrator(ctrl, nil)

	// Dismiss when not active should be a no-op.
	o.Dismiss()
	if o.IsActive() {
		t.Error("expected inactive")
	}
}

func TestOrchestrator_State(t *testing.T) {
	ctrl := overlay.NewNoopController()
	o := NewOrchestrator(ctrl, nil)

	if o.State() != overlay.StateHidden {
		t.Errorf("expected StateHidden initially, got %v", o.State())
	}

	_ = o.Summon(context.Background())
	if o.State() != overlay.StateListening {
		t.Errorf("expected StateListening after Summon, got %v", o.State())
	}
}
