package adaptive

import (
	"context"
	"sync"
	"time"
)

// Observer watches session interactions and emits observations.
// Only user-initiated actions count as evidence — agent-suggested-
// then-accepted actions are filtered to prevent the Predictor →
// Observer → Dialectic loop from reinforcing its own guesses.
type Observer struct {
	mu        sync.Mutex
	events    []Observation
	onObserve func(Observation)

	// trackSuggested tracks agent-suggested actions so we can
	// exclude them from evidence when the user accepts them.
	trackSuggested map[string]bool // session_id → was_suggested
}

// NewObserver creates a session observer.
func NewObserver() *Observer {
	return &Observer{
		trackSuggested: make(map[string]bool),
	}
}

// OnObserve sets the callback for new observations.
func (o *Observer) OnObserve(fn func(Observation)) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.onObserve = fn
}

// Record stores a session event for later analysis.
func (o *Observer) Record(ctx context.Context, obs Observation) {
	_ = ctx
	o.mu.Lock()
	defer o.mu.Unlock()

	// Filter: only count user-initiated actions as evidence.
	if obs.UserInitiated {
		o.events = append(o.events, obs)
		if o.onObserve != nil {
			o.onObserve(obs)
		}
	}

	// Prune events older than 7 days.
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	filtered := o.events[:0]
	for _, e := range o.events { //nolint:gocritic
		if e.Timestamp.After(cutoff) {
			filtered = append(filtered, e)
		}
	}
	o.events = filtered
}

// Recent returns observations from the last N days.
func (o *Observer) Recent(days int) []Observation {
	o.mu.Lock()
	defer o.mu.Unlock()

	if days <= 0 {
		days = 7
	}
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	var out []Observation
	for _, e := range o.events { //nolint:gocritic
		if e.Timestamp.After(cutoff) {
			out = append(out, e)
		}
	}
	return out
}

// MarkSuggested records that a session was agent-suggested.
func (o *Observer) MarkSuggested(sessionID string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.trackSuggested[sessionID] = true
}

// WasSuggested checks if a session was agent-suggested.
func (o *Observer) WasSuggested(sessionID string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.trackSuggested[sessionID]
}

// Count returns the number of stored observations.
func (o *Observer) Count() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return len(o.events)
}
