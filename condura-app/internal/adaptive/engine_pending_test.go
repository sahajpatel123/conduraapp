package adaptive

import (
	"context"
	"testing"
	"time"
)

// helperEngineWithPending builds an Engine + populates its pending
// slice via Run, returning both. Existing e2e_test.go uses a similar
// pattern; we keep it local to avoid coupling.
func helperEngineWithPending(t *testing.T) *Engine {
	t.Helper()
	db := testDB(t)
	s, _ := NewEncryptedStore(db, passthroughEncrypt, passthroughDecrypt)

	observer := NewObserver()
	adj := NewAdjudicator([]string{"verbosity"}, []string{"communication_style"}, 0.6)
	proposer := &mockLLM{
		response: `[{"category":"communication_style","field":"style","value":"casual","confidence":0.9,"reason":"User uses casual language"}]`,
	}
	dialectic := NewDialectic(proposer, "gpt-4", nil, "", adj, nil, StrengthBalanced)
	cfg := DefaultConfig()
	cfg.Strength = StrengthBalanced
	engine := NewEngine(observer, dialectic, adj, s, NewPredictor(s, func() Strength { return StrengthBalanced }), cfg, nil)

	// Record one observation and Run to populate pending. We can
	// only get 1 pending at a time per Run; for n>1 we'd need to
	// re-seed between Runs. Use n=1 for the helper; tests that
	// need multiple pending can call Run multiple times.
	observer.Record(context.Background(), Observation{
		SessionID: "s1", UserQuery: "hey", AgentReply: "hi",
		UserInitiated: true, Timestamp: time.Now(),
	})
	engine.Run(context.Background())
	return engine
}

// TestEngine_RejectPending_OutOfRangeReturnsFalse pins the input
// validation contract: RejectPending MUST return false (no panic, no
// mutation) when idx is negative OR idx >= len(pending). A regression
// that silently dropped the last proposal on out-of-range would
// silently lose user data; a regression that panicked on out-of-range
// would crash the daemon.
func TestEngine_RejectPending_OutOfRangeReturnsFalse(t *testing.T) {
	e := helperEngineWithPending(t)
	if len(e.Pending()) != 1 {
		t.Fatalf("setup: expected 1 pending, got %d", len(e.Pending()))
	}

	if e.RejectPending(-1) {
		t.Error("RejectPending(-1) = true; want false (out-of-range guard)")
	}
	if e.RejectPending(999) {
		t.Error("RejectPending(999) = true; want false (out-of-range guard)")
	}
	// State unchanged.
	if len(e.Pending()) != 1 {
		t.Errorf("out-of-range RejectPending mutated pending slice: len now %d, want 1", len(e.Pending()))
	}
}

// TestEngine_RejectPending_RemovesProposal pins the removal
// contract: RejectPending MUST remove the proposal at idx from the
// pending slice (NOT apply it to the model).
func TestEngine_RejectPending_RemovesProposal(t *testing.T) {
	e := helperEngineWithPending(t)
	if len(e.Pending()) != 1 {
		t.Fatalf("setup: expected 1 pending, got %d", len(e.Pending()))
	}
	// Snapshot the proposal's content fingerprint before mutation.
	// Proposal has no ID field, so use Category+Field+Value.
	pre := e.Pending()[0]
	preFingerprint := pre.Category + "|" + pre.Field + "|" + pre.Value

	if !e.RejectPending(0) {
		t.Error("RejectPending(0) = false; want true")
	}
	if len(e.Pending()) != 0 {
		t.Errorf("after RejectPending(0), pending len = %d, want 0", len(e.Pending()))
	}
	// The right proposal was removed.
	for _, p := range e.Pending() {
		fp := p.Category + "|" + p.Field + "|" + p.Value
		if fp == preFingerprint {
			t.Errorf("proposal %q still in pending after rejection", preFingerprint)
		}
	}
}

// TestEngine_RejectPending_DoesNotApplyToModel pins the semantic
// difference from ConfirmPending: rejection MUST NOT apply the
// proposal to the model (model.Version stays unchanged). A regression
// that applied-on-reject would invert the user's intent ("reject"
// becomes "accept").
func TestEngine_RejectPending_DoesNotApplyToModel(t *testing.T) {
	db := testDB(t)
	s, _ := NewEncryptedStore(db, passthroughEncrypt, passthroughDecrypt)

	observer := NewObserver()
	adj := NewAdjudicator([]string{"verbosity"}, []string{"communication_style"}, 0.6)
	proposer := &mockLLM{
		response: `[{"category":"communication_style","field":"style","value":"casual","confidence":0.9,"reason":"User uses casual language"}]`,
	}
	dialectic := NewDialectic(proposer, "gpt-4", nil, "", adj, nil, StrengthBalanced)
	cfg := DefaultConfig()
	cfg.Strength = StrengthBalanced
	engine := NewEngine(observer, dialectic, adj, s, NewPredictor(s, func() Strength { return StrengthBalanced }), cfg, nil)

	observer.Record(context.Background(), Observation{
		SessionID: "s1", UserQuery: "hey", AgentReply: "hi",
		UserInitiated: true, Timestamp: time.Now(),
	})
	engine.Run(context.Background())
	if len(engine.Pending()) != 1 {
		t.Fatalf("setup: expected 1 pending, got %d", len(engine.Pending()))
	}

	// Snapshot the current model version (typically 1 from init).
	modelBefore, _ := s.Load()
	versionBefore := modelBefore.Version

	engine.RejectPending(0)

	modelAfter, _ := s.Load()
	if modelAfter.Version != versionBefore {
		t.Errorf("model.Version changed after RejectPending: %d -> %d (rejection must NOT apply)",
			versionBefore, modelAfter.Version)
	}
}

// TestEngine_SetStrength_UpdatesEngineConfig pins the config-update
// half of the contract: SetStrength MUST update the engine's
// internal cfg.Strength field so subsequent Run calls use the new
// strength.
func TestEngine_SetStrength_UpdatesEngineConfig(t *testing.T) {
	db := testDB(t)
	s, _ := NewEncryptedStore(db, passthroughEncrypt, passthroughDecrypt)
	observer := NewObserver()
	adj := NewAdjudicator([]string{"verbosity"}, []string{"communication_style"}, 0.6)
	proposer := &mockLLM{
		response: `[{"category":"communication_style","field":"style","value":"casual","confidence":0.9}]`,
	}
	dialectic := NewDialectic(proposer, "gpt-4", nil, "", adj, nil, StrengthBalanced)
	cfg := DefaultConfig()
	cfg.Strength = StrengthBalanced
	e := NewEngine(observer, dialectic, adj, s, NewPredictor(s, func() Strength { return StrengthBalanced }), cfg, nil)

	if e.cfg.Strength != StrengthBalanced {
		t.Fatalf("setup: cfg.Strength = %v, want StrengthBalanced", e.cfg.Strength)
	}

	e.SetStrength(StrengthAggressive)
	if e.cfg.Strength != StrengthAggressive {
		t.Errorf("cfg.Strength after SetStrength(Aggressive) = %v, want StrengthAggressive", e.cfg.Strength)
	}

	e.SetStrength(StrengthOff)
	if e.cfg.Strength != StrengthOff {
		t.Errorf("cfg.Strength after SetStrength(Off) = %v, want StrengthOff", e.cfg.Strength)
	}
}

// TestEngine_SetStrength_PropagatesToDialectic pins the live-update
// half of the contract (P2-8 live update): SetStrength MUST propagate
// the new strength to the dialectic so its LLM prompts use the new
// strength immediately. Without this, the user would set strength to
// Aggressive but the dialectic would keep using the old strength until
// daemon restart.
func TestEngine_SetStrength_PropagatesToDialectic(t *testing.T) {
	db := testDB(t)
	s, _ := NewEncryptedStore(db, passthroughEncrypt, passthroughDecrypt)
	observer := NewObserver()
	adj := NewAdjudicator([]string{"verbosity"}, []string{"communication_style"}, 0.6)
	proposer := &mockLLM{
		response: `[{"category":"communication_style","field":"style","value":"casual","confidence":0.9}]`,
	}
	dialectic := NewDialectic(proposer, "gpt-4", nil, "", adj, nil, StrengthBalanced)
	cfg := DefaultConfig()
	cfg.Strength = StrengthBalanced
	e := NewEngine(observer, dialectic, adj, s, NewPredictor(s, func() Strength { return StrengthBalanced }), cfg, nil)

	if e.Dialectic.strength != StrengthBalanced {
		t.Fatalf("setup: Dialectic.strength = %v, want StrengthBalanced", e.Dialectic.strength)
	}

	e.SetStrength(StrengthAggressive)
	if e.Dialectic.strength != StrengthAggressive {
		t.Errorf("Dialectic.strength after SetStrength(Aggressive) = %v, want StrengthAggressive",
			e.Dialectic.strength)
	}

	e.SetStrength(StrengthCautious)
	if e.Dialectic.strength != StrengthCautious {
		t.Errorf("Dialectic.strength after SetStrength(Conservative) = %v, want StrengthCautious",
			e.Dialectic.strength)
	}
}

// TestEngine_SetStrength_NilDialecticNoPanic pins the nil-guard: if
// the engine is constructed with a nil Dialectic (e.g., for tests or
// minimal deployments), SetStrength MUST NOT panic on the
// `e.Dialectic != nil` check.
func TestEngine_SetStrength_NilDialecticNoPanic(t *testing.T) {
	db := testDB(t)
	s, _ := NewEncryptedStore(db, passthroughEncrypt, passthroughDecrypt)
	observer := NewObserver()
	adj := NewAdjudicator([]string{"verbosity"}, []string{"communication_style"}, 0.6)
	cfg := DefaultConfig()
	cfg.Strength = StrengthBalanced
	e := NewEngine(observer, nil, adj, s, NewPredictor(s, func() Strength { return StrengthBalanced }), cfg, nil)

	// Should not panic.
	e.SetStrength(StrengthAggressive)
	if e.cfg.Strength != StrengthAggressive {
		t.Errorf("cfg.Strength = %v, want StrengthAggressive", e.cfg.Strength)
	}
}
