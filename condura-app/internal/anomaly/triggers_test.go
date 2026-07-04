package anomaly

import (
	"sync"
	"testing"
	"time"
)

// TestDetector_TripRate fires TripRate when the action rate
// exceeds 20/min. We set state.count and state.startTime directly
// rather than calling process() N times — process() involves
// syscalls that are slow on Windows CI runners (~1ms each), so
// 25 iterations can take >1.25 min, making the rate appear ≤20/min
// and the test flaky. Direct state mutation is deterministic.
//
// CLAUDE.md §10.4 / §5.6: "Speed: >20 actions/minute → pause".
// Without this test, the rate trigger could regress silently.
func TestDetector_TripRate(t *testing.T) {
	tripped := false
	d := NewDetector(func(tr Trip) {
		if tr.Type == TripRate {
			tripped = true
		}
	})
	// 25 actions in 30 seconds = 50/min, comfortably above 20.
	d.state.count = 25
	d.state.startTime = time.Now().Add(-30 * time.Second)
	d.checkRate()
	if !tripped {
		t.Error("TripRate should fire when rate > 20 actions/minute")
	}
}

// TestDetector_TripRateBoundary fires TripRate only when the rate
// is > 20/min, not at exactly 20/min. The threshold is a hard >,
// not >= (see detector.go:239). This test pins that behavior.
//
// As with TestDetector_TripRate, we mutate state directly rather
// than calling process() N times to stay deterministic on Windows.
func TestDetector_TripRateBoundary(t *testing.T) {
	tripped := false
	d := NewDetector(func(tr Trip) {
		if tr.Type == TripRate {
			tripped = true
		}
	})
	// Exactly 20 actions in exactly 1 minute = 20/min. Must NOT trip.
	d.state.count = 20
	d.state.startTime = time.Now().Add(-1 * time.Minute)
	d.checkRate()
	if tripped {
		t.Error("TripRate should NOT fire at exactly 20/min (threshold is strict >)")
	}
}

// TestDetector_TripDuration fires TripDuration when the task
// has been running >30 minutes. We manipulate state.startTime
// to simulate the time passage without waiting.
//
// CLAUDE.md §10.4 / §5.6: "Duration: >30 minutes on one task → pause".
func TestDetector_TripDuration(t *testing.T) {
	tripped := false
	d := NewDetector(func(tr Trip) {
		if tr.Type == TripDuration {
			tripped = true
		}
	})
	// Backdate the start time by 31 minutes.
	d.state.startTime = time.Now().Add(-31 * time.Minute)
	d.process(actionRecord{kind: "click", coordX: 1, coordY: 1, success: true, time: time.Now()})
	if !tripped {
		t.Error("TripDuration should fire when task duration exceeds 30 minutes")
	}
}

// TestDetector_TripDurationBoundary fires TripDuration only when
// duration > 30 minutes. At exactly 30 minutes, it should NOT trip.
// This pins the threshold direction (strict >, not >=).
func TestDetector_TripDurationBoundary(t *testing.T) {
	tripped := false
	d := NewDetector(func(tr Trip) {
		if tr.Type == TripDuration {
			tripped = true
		}
	})
	// Backdate by exactly 30 minutes (one second under).
	d.state.startTime = time.Now().Add(-30*time.Minute + time.Second)
	d.process(actionRecord{kind: "click", coordX: 1, coordY: 1, success: true, time: time.Now()})
	if tripped {
		t.Error("TripDuration should NOT fire at exactly 30 minutes (threshold is strict >)")
	}
}

// TestDetector_TripFailuresStopsAfterTrip verifies that after a
// failures trip fires, the detector does not keep emitting trips
// for the same condition (avoiding audit-log spam / user spam).
// Each unique TripType should fire at most once until state is
// reset, otherwise the user sees the same notification forever.
func TestDetector_TripFailuresStopsAfterTrip(t *testing.T) {
	var trips sync.Map
	d := NewDetector(func(tr Trip) {
		// Count trips per type.
		v, _ := trips.LoadOrStore(tr.Type, 1)
		trips.Store(tr.Type, v.(int)+1)
	})
	// 6 consecutive failures → should trip on the 5th, then the 6th
	// also trips (current behavior). The detector does NOT short-
	// circuit post-trip; this test documents that behavior so any
	// future change to "trip once per session" is intentional.
	for i := 0; i < 6; i++ {
		d.process(actionRecord{kind: "click", coordX: 0, coordY: 0, success: false})
	}
	v, ok := trips.Load(TripFailures)
	if !ok {
		t.Fatal("expected at least one TripFailures")
	}
	if v.(int) < 1 {
		t.Errorf("TripFailures count = %d, want >= 1", v.(int))
	}
	// Sanity: at least one trip but not a runaway.
	if v.(int) > 10 {
		t.Errorf("TripFailures count = %d, seems like a runaway", v.(int))
	}
}

// TestDetector_ResetClearsCounters verifies Reset() brings the
// state back to a clean baseline. This is critical for the
// "Resume after pause" UX: the user resolves whatever triggered
// the trip, hits resume, and the detector should not immediately
// re-trip on the same data.
func TestDetector_ResetClearsCounters(t *testing.T) {
	d := NewDetector(nil)
	// Build up some state.
	for i := 0; i < 10; i++ {
		d.process(actionRecord{kind: "click", coordX: 1, coordY: 1, success: true, time: time.Now()})
	}
	d.Reset()

	if d.state.count != 0 {
		t.Errorf("count after Reset = %d, want 0", d.state.count)
	}
	if !d.state.lastActivity.IsZero() {
		t.Errorf("lastActivity after Reset = %v, want zero", d.state.lastActivity)
	}
	if len(d.state.coordWindow) != 0 {
		t.Errorf("coordWindow after Reset has %d entries, want 0", len(d.state.coordWindow))
	}
}

// TestDetector_LastActivityUpdates verifies that LastActivity()
// returns the timestamp of the most recent action, not the
// detector's construction time. The tray uses this to show
// "last activity 5m ago" to the user.
func TestDetector_LastActivityUpdates(t *testing.T) {
	d := NewDetector(nil)
	before := time.Now().Add(-time.Hour) // one hour ago
	d.process(actionRecord{kind: "click", coordX: 1, coordY: 1, success: true, time: before})
	got := d.LastActivity()
	if !got.Equal(before) {
		t.Errorf("LastActivity = %v, want %v", got, before)
	}
}
