package perception

import (
	"slices"
	"testing"
)

// TestStrategyPreferenceFor_StateOnlyQuestionReturnsStrategyNone
// pins the cheapest-strategy contract: a question that needs no
// element identity / pixels / OCR (the question is answerable
// from agent state alone) MUST return just [StrategyNone]. The
// perception layer should NOT fire any expensive capture when
// the answer is already in agent state.
func TestStrategyPreferenceFor_StateOnlyQuestionReturnsStrategyNone(t *testing.T) {
	q := Question{}
	// All NeedsX flags false → stateOnly=true.
	dirty := DirtyState{}

	got := strategyPreferenceFor(q, dirty)
	if !slices.Equal(got, []Strategy{StrategyNone}) {
		t.Errorf("strategyPreferenceFor(stateOnly) = %v, want [StrategyNone]", got)
	}
}

// TestStrategyPreferenceFor_ElementIdentityOnlyReturnsAXOnly pins
// the element-without-pixels contract: a question that needs
// element identity but NOT pixels (and NOT OCR) MUST return
// just [StrategyAXOnly]. AX (accessibility tree) is enough; no
// need to capture screenshots.
func TestStrategyPreferenceFor_ElementIdentityOnlyReturnsAXOnly(t *testing.T) {
	q := Question{
		NeedsElementIdentity: true,
		// NeedsPixels: false (default)
		// NeedsOCR: false (default)
	}
	dirty := DirtyState{}

	got := strategyPreferenceFor(q, dirty)
	if !slices.Equal(got, []Strategy{StrategyAXOnly}) {
		t.Errorf("strategyPreferenceFor(NeedsElementIdentity only) = %v, want [StrategyAXOnly]", got)
	}
}

// TestStrategyPreferenceFor_DirtyAndVisualReturnsFullCascade pins
// the dirty + visual contract: when state has changed (dirty)
// AND the question needs pixels/OCR, the perception layer MUST
// include StrategyDifferential (cheapest fast-path) at the
// front of the cascade. Without Differential, the cascade would
// re-screenshot from scratch on every question.
func TestStrategyPreferenceFor_DirtyAndVisualReturnsFullCascade(t *testing.T) {
	q := Question{
		NeedsPixels: true, // stateOnly = false
	}
	dirty := DirtyState{Dirty: true}

	got := strategyPreferenceFor(q, dirty)
	// Expect: [StrategyAXOnly, StrategyDifferential, StrategyWindowRect, StrategyFullScreen, StrategyVisionCUA]
	// (no StrategyNone because stateOnly is false — the question NeedsPixels)
	want := []Strategy{
		StrategyAXOnly,
		StrategyDifferential,
		StrategyWindowRect,
		StrategyFullScreen,
		StrategyVisionCUA,
	}
	if !slices.Equal(got, want) {
		t.Errorf("strategyPreferenceFor(dirty+NeedsPixels) = %v, want %v", got, want)
	}
}

// TestStrategyPreferenceFor_DirtyAndStateOnlyFallsThroughToDefault
// pins a subtle but important contract: when dirty AND the question
// is state-only (no NeedsX flags), the dirty cascade does NOT fire
// (the dirty branch requires NeedsOCR || NeedsPixels to include
// Differential). The function falls through to the default branch,
// which returns [StrategyNone] for state-only. This is correct —
// state-only questions don't need any capture.
func TestStrategyPreferenceFor_DirtyAndStateOnlyFallsThroughToDefault(t *testing.T) {
	q := Question{} // stateOnly
	dirty := DirtyState{Dirty: true}

	got := strategyPreferenceFor(q, dirty)
	// Expect: just [StrategyNone] (default branch, no Differential
	// cascade because the dirty branch's `NeedsOCR || NeedsPixels`
	// guard skipped it).
	if !slices.Equal(got, []Strategy{StrategyNone}) {
		t.Errorf("strategyPreferenceFor(dirty+stateOnly) = %v, want [StrategyNone]", got)
	}
}

// TestStrategyPreferenceFor_TargetAppSetReturnsCascadeWithStrategyNone
// pins the target-app + state-only contract: when TargetApp is
// set AND the question is state-only, the cascade includes
// StrategyNone (prepended because stateOnly=true). The cascade
// also includes StrategyWindowRect (the single-app-specific
// strategy).
func TestStrategyPreferenceFor_TargetAppSetReturnsCascadeWithStrategyNone(t *testing.T) {
	q := Question{
		TargetApp: "Finder",
		// stateOnly = true
	}
	dirty := DirtyState{}

	got := strategyPreferenceFor(q, dirty)
	// Expect: [StrategyNone, StrategyAXOnly, StrategyWindowRect, StrategyFullScreen, StrategyVisionCUA]
	want := []Strategy{
		StrategyNone,
		StrategyAXOnly,
		StrategyWindowRect,
		StrategyFullScreen,
		StrategyVisionCUA,
	}
	if !slices.Equal(got, want) {
		t.Errorf("strategyPreferenceFor(TargetApp+stateOnly) = %v, want %v", got, want)
	}
	// Sanity: the cascade MUST contain StrategyWindowRect
	// (the single-app-specific strategy).
	if !slices.Contains(got, StrategyWindowRect) {
		t.Errorf("strategyPreferenceFor(TargetApp) = %v; want StrategyWindowRect in cascade", got)
	}
}

// TestStrategyPreferenceFor_TargetAppAndStateOnlyReturnsCascadeWithStrategyNone
// pins the target-app + state-only contract: when TargetApp is
// set AND the question is state-only, StrategyNone is prepended.
// This is the cheapest path possible — neither a capture nor a
// diff is needed; just check the agent's state for the answer.
func TestStrategyPreferenceFor_TargetAppAndStateOnlyReturnsCascadeWithStrategyNone(t *testing.T) {
	q := Question{
		TargetApp: "Finder",
		// stateOnly = true
	}
	dirty := DirtyState{}

	got := strategyPreferenceFor(q, dirty)
	// Expect: [StrategyNone, StrategyAXOnly, StrategyWindowRect, StrategyFullScreen, StrategyVisionCUA]
	want := []Strategy{
		StrategyNone,
		StrategyAXOnly,
		StrategyWindowRect,
		StrategyFullScreen,
		StrategyVisionCUA,
	}
	if !slices.Equal(got, want) {
		t.Errorf("strategyPreferenceFor(TargetApp+stateOnly) = %v, want %v", got, want)
	}
}

// TestStrategyPreferenceFor_DefaultCascadeHasStrategyDifferentialExcluded
// pins a non-obvious contract: the DEFAULT cascade does NOT
// include StrategyDifferential (Differential is only included in
// the DIRTY branch, because Differential compares against a
// cached screenshot — without a dirty event, there's nothing to
// compare against).
func TestStrategyPreferenceFor_DefaultCascadeHasStrategyDifferentialExcluded(t *testing.T) {
	q := Question{
		NeedsElementIdentity: true,
		NeedsPixels:          true,
		NeedsOCR:             true,
	}
	dirty := DirtyState{}

	got := strategyPreferenceFor(q, dirty)
	if slices.Contains(got, StrategyDifferential) {
		t.Errorf("strategyPreferenceFor(default cascade) = %v; want NO StrategyDifferential (that's dirty-only)",
			got)
	}
}

// TestStrategyPreferenceFor_DirtyBranchIncludesDifferential pins
// the inverse: the DIRTY branch MUST include StrategyDifferential.
// A regression that dropped Differential from the dirty cascade
// would force a full re-screenshot for every question on dirty
// state — defeating the purpose of dirty-tracking.
func TestStrategyPreferenceFor_DirtyBranchIncludesDifferential(t *testing.T) {
	q := Question{NeedsPixels: true}
	dirty := DirtyState{Dirty: true}

	got := strategyPreferenceFor(q, dirty)
	if !slices.Contains(got, StrategyDifferential) {
		t.Errorf("strategyPreferenceFor(dirty+NeedsPixels) = %v; want StrategyDifferential in cascade",
			got)
	}
}
