package anomaly

import "testing"

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
