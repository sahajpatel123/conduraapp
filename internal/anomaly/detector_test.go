package anomaly

import (
	"testing"
	"time"
)

func TestDetector_LoopTrip(t *testing.T) {
	tripped := false
	d := NewDetector(func(tr Trip) { tripped = true })
	// Record synchronously enough to trigger loop with same coords.
	for i := 0; i < 4; i++ {
		d.process(actionRecord{kind: "click", coordX: 100, coordY: 200, success: true})
	}
	if !tripped {
		t.Error("expected loop trip after 4 same-coordinate actions")
	}
}

func TestDetector_FailureTrip(t *testing.T) {
	tripped := false
	d := NewDetector(func(tr Trip) {
		if tr.Type == TripFailures {
			tripped = true
		}
	})
	for i := 0; i < 6; i++ {
		d.process(actionRecord{kind: "click", coordX: 0, coordY: 0, success: false})
	}
	if !tripped {
		t.Error("expected failure trip after 5+ consecutive failures")
	}
}

func TestDetector_FailureCounterResetsOnSuccess(t *testing.T) {
	tripped := false
	d := NewDetector(func(tr Trip) {
		if tr.Type == TripFailures {
			tripped = true
		}
	})
	// 4 failures, then a success, then 4 more failures.
	for i := 0; i < 4; i++ {
		d.process(actionRecord{kind: "click", coordX: 0, coordY: 0, success: false})
	}
	d.process(actionRecord{kind: "click", coordX: 0, coordY: 0, success: true})
	for i := 0; i < 4; i++ {
		d.process(actionRecord{kind: "click", coordX: 0, coordY: 0, success: false})
	}
	if tripped {
		t.Error("failure counter should reset on success; 4 failures after a success is not 5 consecutive")
	}
}

func TestDetector_Reset(t *testing.T) {
	d := NewDetector(nil)
	d.process(actionRecord{kind: "click", coordX: 1, coordY: 1, success: true})
	d.process(actionRecord{kind: "click", coordX: 1, coordY: 1, success: true})
	d.process(actionRecord{kind: "click", coordX: 1, coordY: 1, success: true})
	d.Reset()
	tripped := false
	d.onTrip = func(tr Trip) { tripped = true }
	d.process(actionRecord{kind: "click", coordX: 1, coordY: 1, success: true})
	d.process(actionRecord{kind: "click", coordX: 1, coordY: 1, success: true})
	if tripped {
		t.Error("should not trip after reset with only 2 records")
	}
	d.process(actionRecord{kind: "click", coordX: 1, coordY: 1, success: true})
	d.process(actionRecord{kind: "click", coordX: 1, coordY: 1, success: true})
	t.Logf("tripped after 4 same-coord records (post-reset): %v", tripped)
}

// Phase 16, Rec 6 — IdleReset() returns true after the detector
// has been quiet longer than the idle threshold.
func TestDetector_IdleReset_Quiet(t *testing.T) {
	d := NewDetector(nil)
	// Manually set lastActivity to 2 hours ago.
	d.state.lastActivity = time.Now().Add(-2 * time.Hour)
	d.state.count = 5
	d.state.failures = 1
	if !d.IdleReset(30 * time.Minute) {
		t.Error("expected IdleReset to fire after 2h of quiet (threshold 30m)")
	}
}

// Phase 16, Rec 6 — IdleReset() returns false when the detector
// is active within the threshold.
func TestDetector_IdleReset_Active(t *testing.T) {
	d := NewDetector(nil)
	d.process(actionRecord{kind: "click", coordX: 1, coordY: 1, success: true, time: time.Now()})
	if d.IdleReset(30 * time.Minute) {
		t.Error("expected IdleReset to NOT fire after recent activity")
	}
}

// Phase 16, Rec 6 — IdleReset() returns false when the detector
// has never been used (zero state).
func TestDetector_IdleReset_ZeroState(t *testing.T) {
	d := NewDetector(nil)
	if d.IdleReset(30 * time.Minute) {
		t.Error("expected IdleReset to NOT fire on an unused detector")
	}
}

// Phase 16, Rec 6 — IdleReset() returns false for non-positive
// thresholds (defensive).
func TestDetector_IdleReset_ZeroThreshold(t *testing.T) {
	d := NewDetector(nil)
	d.state.lastActivity = time.Now().Add(-1 * time.Hour)
	d.state.count = 1
	if d.IdleReset(0) {
		t.Error("expected IdleReset to NOT fire when threshold is 0")
	}
	if d.IdleReset(-1 * time.Second) {
		t.Error("expected IdleReset to NOT fire when threshold is negative")
	}
}

// Phase 16, Rec 6 — LastActivity() returns zero time on an
// unused detector and the time of the most recent Record()
// after one is recorded.
func TestDetector_LastActivity(t *testing.T) {
	d := NewDetector(nil)
	if !d.LastActivity().IsZero() {
		t.Error("LastActivity on fresh detector should be zero time")
	}
	d.process(actionRecord{kind: "click", coordX: 1, coordY: 1, success: true, time: time.Now()})
	if d.LastActivity().IsZero() {
		t.Error("LastActivity should be set after a process call")
	}
}

// TestDetector_NewEndpoint_TripsOnPivot pins the 5th "agent went
// insane" trigger (§5.6): the first network endpoint in a session is
// NOT a trip, but a pivot to a new endpoint mid-session IS.
func TestDetector_NewEndpoint_TripsOnPivot(t *testing.T) {
	var trips []Trip
	d := NewDetector(func(t Trip) { trips = append(trips, t) })
	// First contact — not a trip (empty set).
	d.RecordNetwork("api.openai.com")
	if len(trips) != 0 {
		t.Errorf("first endpoint should not trip, got %d trips: %+v", len(trips), trips)
	}
	// Same host again — not a trip.
	d.RecordNetwork("api.openai.com")
	if len(trips) != 0 {
		t.Errorf("repeat endpoint should not trip, got %d trips", len(trips))
	}
	// New host mid-session — trip.
	d.RecordNetwork("evil.example.com")
	if len(trips) != 1 || trips[0].Type != TripNewEndpoint {
		t.Errorf("pivot to new endpoint should trip TripNewEndpoint, got %+v", trips)
	}
}

// TestDetector_NewEndpoint_NormalizesHost pins that the host is
// port-stripped + lowercased before comparison, so "API.OpenAI.com:443"
// matches "api.openai.com".
func TestDetector_NewEndpoint_NormalizesHost(t *testing.T) {
	var trips []Trip
	d := NewDetector(func(t Trip) { trips = append(trips, t) })
	d.RecordNetwork("API.OpenAI.com:443")
	d.RecordNetwork("api.openai.com")
	if len(trips) != 0 {
		t.Errorf("normalized repeat should not trip, got %d trips", len(trips))
	}
}

// TestDetector_NewEndpoint_EmptyHostIgnored pins that an empty
// host string is a no-op (no trip, no state change).
func TestDetector_NewEndpoint_EmptyHostIgnored(t *testing.T) {
	var trips []Trip
	d := NewDetector(func(t Trip) { trips = append(trips, t) })
	d.RecordNetwork("")
	d.RecordNetwork("   ")
	if len(trips) != 0 {
		t.Errorf("empty host should not trip, got %d trips", len(trips))
	}
}
