package presence

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/overlay"
)

func TestOrchestrator_SummonAndDismiss(t *testing.T) {
	ctrl := overlay.NewNoopController()
	o := NewOrchestrator(ctrl, nil, nil)

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
	o := NewOrchestrator(ctrl, nil, nil)

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
	o := NewOrchestrator(ctrl, nil, nil)

	// Dismiss when not active should be a no-op.
	o.Dismiss()
	if o.IsActive() {
		t.Error("expected inactive")
	}
}

func TestOrchestrator_State(t *testing.T) {
	ctrl := overlay.NewNoopController()
	o := NewOrchestrator(ctrl, nil, nil)

	if o.State() != overlay.StateHidden {
		t.Errorf("expected StateHidden initially, got %v", o.State())
	}

	_ = o.Summon(context.Background())
	if o.State() != overlay.StateListening {
		t.Errorf("expected StateListening after Summon, got %v", o.State())
	}
}

func TestOrchestrator_CaptureStartFailsRollsBack(t *testing.T) {
	ctrl := overlay.NewNoopController()
	capture := &fakeCapture{startErr: errors.New("mic dead")}
	o := NewOrchestrator(ctrl, nil, capture)

	err := o.Summon(context.Background())
	if err == nil {
		t.Fatal("expected error from Summon when capture fails")
	}
	if o.IsActive() {
		t.Error("session should not be active after capture failure")
	}
	if o.State() != overlay.StateHidden {
		t.Errorf("overlay should be hidden after capture failure, got %v", o.State())
	}
}

func TestOrchestrator_CaptureStartAndStopFires(t *testing.T) {
	ctrl := overlay.NewNoopController()
	capture := &fakeCapture{}
	o := NewOrchestrator(ctrl, nil, capture)

	if err := o.Summon(context.Background()); err != nil {
		t.Fatalf("Summon: %v", err)
	}
	if capture.starts.Load() != 1 {
		t.Errorf("capture starts = %d, want 1", capture.starts.Load())
	}

	o.Dismiss()
	if capture.stops.Load() != 1 {
		t.Errorf("capture stops = %d, want 1", capture.stops.Load())
	}
}

func TestOrchestrator_NilCaptureIsAllowed(t *testing.T) {
	ctrl := overlay.NewNoopController()
	// nil capture is explicitly allowed (overlay-only session).
	o := NewOrchestrator(ctrl, nil, nil)
	if err := o.Summon(context.Background()); err != nil {
		t.Fatalf("Summon: %v", err)
	}
	o.Dismiss() // must not panic on nil capture
}

func TestCaptureFuncs(t *testing.T) {
	var started, stopped atomic.Int32
	capture := CaptureFuncs{
		StartFn: func(_ context.Context) error { started.Add(1); return nil },
		StopFn:  func() error { stopped.Add(1); return nil },
	}
	if err := capture.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if started.Load() != 1 {
		t.Errorf("started = %d", started.Load())
	}
	if err := capture.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if stopped.Load() != 1 {
		t.Errorf("stopped = %d", stopped.Load())
	}
}

type fakeCapture struct {
	starts   atomic.Int32
	stops    atomic.Int32
	startErr error
	stopErr  error
}

func (f *fakeCapture) Start(_ context.Context) error {
	f.starts.Add(1)
	return f.startErr
}

func (f *fakeCapture) Stop() error {
	f.stops.Add(1)
	return f.stopErr
}
