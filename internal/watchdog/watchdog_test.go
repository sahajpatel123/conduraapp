package watchdog

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/halt"
)

// fakeHalt records Halt calls and reports IsHalted=true after the
// first call. Thread-safe for the watchdog's concurrent Run goroutine.
type fakeHalt struct {
	mu      sync.Mutex
	halted  bool
	reasons []string
	haltErr error
	hits    atomic.Int32
}

func (f *fakeHalt) Halt(_ context.Context, reason string) (halt.State, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.halted = true
	f.reasons = append(f.reasons, reason)
	f.hits.Add(1)
	return halt.State{}, f.haltErr
}

func (f *fakeHalt) IsHalted() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.halted
}

func TestWatchdog_NewSetsInitialTouch(t *testing.T) {
	w := New(time.Hour, time.Minute, nil, nil)
	if w.LastTouch().IsZero() {
		t.Fatal("New() must call lastTouch = time.Now() so daemon doesn't self-halt on startup")
	}
}

func TestWatchdog_TouchUpdatesLastTouch(t *testing.T) {
	w := New(time.Hour, time.Minute, nil, nil)
	before := w.LastTouch()
	time.Sleep(2 * time.Millisecond)
	w.Touch()
	after := w.LastTouch()
	if !after.After(before) {
		t.Fatalf("Touch should advance LastTouch: before=%v after=%v", before, after)
	}
}

func TestWatchdog_IdleDurationCountsSinceLastTouch(t *testing.T) {
	w := New(time.Hour, time.Minute, nil, nil)
	w.lastTouch = time.Now().Add(-5 * time.Minute)
	idle := w.IdleDuration()
	if idle < 4*time.Minute || idle > 6*time.Minute {
		t.Fatalf("IdleDuration = %v, want ~5m", idle)
	}
}

func TestWatchdog_Run_HaltsAfterTimeout(t *testing.T) {
	h := &fakeHalt{}
	w := New(50*time.Millisecond, 10*time.Millisecond, h, nil)
	// Pretend the user touched the watchdog 100ms ago, then went
	// idle for > 50ms (the timeout).
	w.lastTouch = time.Now().Add(-100 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watchdog did not fire within 1s of timeout expiry")
	}
	if !h.IsHalted() {
		t.Fatal("halt flag should be set after watchdog trips")
	}
	if len(h.reasons) != 1 {
		t.Fatalf("halt called %d times, want 1", len(h.reasons))
	}
	if h.reasons[0] == "" || h.reasons[0][0:8] != "watchdog" {
		t.Fatalf("halt reason should mention watchdog: %q", h.reasons[0])
	}
}

func TestWatchdog_Run_NoHaltWhenActive(t *testing.T) {
	h := &fakeHalt{}
	w := New(time.Hour, 10*time.Millisecond, h, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()
	<-done
	if h.IsHalted() {
		t.Fatal("active watchdog should not halt")
	}
	if h.hits.Load() != 0 {
		t.Fatalf("halt called %d times, want 0", h.hits.Load())
	}
}

func TestWatchdog_Run_AlreadyHaltedIsNoOp(t *testing.T) {
	h := &fakeHalt{halted: true} // pretend a prior halt already fired
	w := New(10*time.Millisecond, 5*time.Millisecond, h, nil)
	w.lastTouch = time.Now().Add(-1 * time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()
	// Even after long inactivity, an already-halted daemon should
	// not get a second halt call.
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("watchdog Run should return quickly when already halted")
	}
	if h.hits.Load() != 0 {
		t.Fatalf("halt called %d times, want 0 (already halted)", h.hits.Load())
	}
}

func TestWatchdog_Run_CtxCancelStopsLoop(t *testing.T) {
	w := New(time.Hour, 10*time.Millisecond, &fakeHalt{}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not return after ctx cancellation")
	}
}

func TestWatchdog_Defaults(t *testing.T) {
	w := New(0, 0, nil, nil)
	if w.timeout != DefaultTimeout {
		t.Errorf("timeout: got %v, want DefaultTimeout=%v", w.timeout, DefaultTimeout)
	}
	if w.interval != DefaultCheckInterval {
		t.Errorf("interval: got %v, want DefaultCheckInterval=%v", w.interval, DefaultCheckInterval)
	}
}

func TestWatchdog_OnTripOverride(t *testing.T) {
	h := &fakeHalt{}
	w := New(10*time.Millisecond, 5*time.Millisecond, h, nil)
	w.lastTouch = time.Now().Add(-1 * time.Hour)
	var tripReason string
	w.onTrip = func(reason string) {
		tripReason = reason
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("onTrip override should fire and return quickly")
	}
	if tripReason == "" {
		t.Fatal("onTrip callback did not fire")
	}
	if h.hits.Load() != 0 {
		t.Fatal("default halt should NOT be called when onTrip is set")
	}
}
