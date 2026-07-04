package adaptive

import (
	"context"
	"testing"
	"time"
)

func TestObserver_RecordsOnlyUserInitiated(t *testing.T) {
	o := NewObserver()
	o.Record(context.Background(), Observation{
		SessionID: "s1", UserQuery: "user started", UserInitiated: true, Timestamp: time.Now(),
	})
	o.Record(context.Background(), Observation{
		SessionID: "s2", UserQuery: "agent suggested", UserInitiated: false, Timestamp: time.Now(),
	})
	if o.Count() != 1 {
		t.Fatalf("got %d events, want 1 (user-initiated only)", o.Count())
	}
}

func TestObserver_Recent(t *testing.T) {
	o := NewObserver()
	now := time.Now()
	o.Record(context.Background(), Observation{SessionID: "old", UserInitiated: true, Timestamp: now.Add(-48 * time.Hour)})
	o.Record(context.Background(), Observation{SessionID: "new", UserInitiated: true, Timestamp: now})

	recent := o.Recent(1) // last 1 day
	if len(recent) != 1 {
		t.Errorf("got %d recent, want 1", len(recent))
	}
}

func TestObserver_OnObserve(t *testing.T) {
	o := NewObserver()
	var got []Observation
	o.OnObserve(func(obs Observation) { got = append(got, obs) })

	o.Record(context.Background(), Observation{SessionID: "s1", UserInitiated: true, Timestamp: time.Now()})
	if len(got) != 1 {
		t.Errorf("onObserve not called")
	}

	o.Record(context.Background(), Observation{SessionID: "s2", UserInitiated: false, Timestamp: time.Now()})
	if len(got) != 1 {
		t.Error("onObserve called for non-user-initiated")
	}
}

func TestObserver_MarkSuggested(t *testing.T) {
	o := NewObserver()
	o.MarkSuggested("s1")
	if !o.WasSuggested("s1") {
		t.Error("expected s1 to be marked")
	}
	if o.WasSuggested("s2") {
		t.Error("s2 should not be marked")
	}
}
