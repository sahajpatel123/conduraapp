package onboarding

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func newTestStateMachine(t *testing.T) *StateMachine {
	t.Helper()
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "onboarding.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	sm, err := NewStateMachine(db)
	if err != nil {
		t.Fatalf("NewStateMachine: %v", err)
	}
	return sm
}

func TestStateMachine_StartsAtWelcome(t *testing.T) {
	sm := newTestStateMachine(t)
	s, err := sm.State(context.Background())
	if err != nil {
		t.Fatalf("State: %v", err)
	}
	if s.CurrentStep != StepWelcome {
		t.Fatalf("CurrentStep: want welcome, got %s", s.CurrentStep)
	}
}

func TestStateMachine_AdvanceWalksThroughAllSteps(t *testing.T) {
	sm := newTestStateMachine(t)
	ctx := context.Background()
	// Walk: welcome → eula → power_source → permissions → backend_detect
	// → hotkey → voice_test → complete
	for _, want := range []Step{
		StepEULA, StepPowerSource, StepPermissions, StepBackendDetect,
		StepHotkey, StepVoiceTest, StepComplete,
	} {
		s, err := sm.Advance(ctx)
		if err != nil {
			t.Fatalf("Advance: %v", err)
		}
		if s.CurrentStep != want {
			t.Fatalf("step: want %s, got %s", want, s.CurrentStep)
		}
	}
	// Advance past the end is a no-op.
	s, err := sm.Advance(ctx)
	if err != nil {
		t.Fatalf("Advance at end: %v", err)
	}
	if s.CurrentStep != StepComplete {
		t.Fatalf("should stay at complete, got %s", s.CurrentStep)
	}
}

func TestStateMachine_BackFromFirstStepNoOp(t *testing.T) {
	sm := newTestStateMachine(t)
	s, err := sm.Back(context.Background())
	if err != nil {
		t.Fatalf("Back: %v", err)
	}
	if s.CurrentStep != StepWelcome {
		t.Fatalf("should stay at welcome, got %s", s.CurrentStep)
	}
}

func TestStateMachine_BackFromLaterStep(t *testing.T) {
	sm := newTestStateMachine(t)
	ctx := context.Background()
	// Move forward three times: empty→eula→power_source→permissions.
	_, _ = sm.Advance(ctx)
	_, _ = sm.Advance(ctx)
	_, _ = sm.Advance(ctx)
	// Move back once: permissions→power_source.
	s, err := sm.Back(ctx)
	if err != nil {
		t.Fatalf("Back: %v", err)
	}
	if s.CurrentStep != StepPowerSource {
		t.Fatalf("want power_source, got %s", s.CurrentStep)
	}
}

func TestStateMachine_SetStepStatus(t *testing.T) {
	sm := newTestStateMachine(t)
	ctx := context.Background()
	s, err := sm.SetStepStatus(ctx, StepEULA, StatusComplete, "v1.0")
	if err != nil {
		t.Fatalf("SetStepStatus: %v", err)
	}
	got, ok := s.Steps[StepEULA]
	if !ok {
		t.Fatalf("eula not in steps")
	}
	if got.Status != StatusComplete {
		t.Fatalf("status: %s", got.Status)
	}
	if got.Data != "v1.0" {
		t.Fatalf("data: %s", got.Data)
	}
}

func TestStateMachine_Skip(t *testing.T) {
	sm := newTestStateMachine(t)
	ctx := context.Background()
	s, err := sm.Skip(ctx, StepVoiceTest)
	if err != nil {
		t.Fatalf("Skip: %v", err)
	}
	if s.Steps[StepVoiceTest].Status != StatusSkipped {
		t.Fatalf("want skipped, got %s", s.Steps[StepVoiceTest].Status)
	}
}

func TestStateMachine_CompleteMarksAllSkippedExceptLast(t *testing.T) {
	sm := newTestStateMachine(t)
	ctx := context.Background()
	// Skip past every step except Complete so we hit the
	// "some are still pending" branch in Complete().
	for _, step := range []Step{
		StepEULA, StepPowerSource, StepPermissions,
		StepBackendDetect, StepHotkey, StepVoiceTest,
	} {
		_, err := sm.SetStepStatus(ctx, step, StatusComplete, "")
		if err != nil {
			t.Fatalf("SetStepStatus %s: %v", step, err)
		}
	}
	s, err := sm.Complete(ctx)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if s.Steps[StepComplete].Status != StatusComplete {
		t.Fatalf("complete step: %s", s.Steps[StepComplete].Status)
	}
	if s.CompletedAt.IsZero() {
		t.Fatalf("CompletedAt not set")
	}
	// All steps should now have a status.
	for _, step := range AllSteps {
		if s.Steps[step].Status == "" {
			t.Fatalf("step %s has no status", step)
		}
	}
}

func TestStateMachine_ResetReturnsToWelcome(t *testing.T) {
	sm := newTestStateMachine(t)
	ctx := context.Background()
	// Walk to the end and complete.
	for i := 0; i < 7; i++ {
		_, _ = sm.Advance(ctx)
	}
	_, _ = sm.Complete(ctx)
	// Reset.
	s, err := sm.Reset(ctx)
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if s.CurrentStep != StepWelcome {
		t.Fatalf("want welcome, got %s", s.CurrentStep)
	}
	for step, p := range s.Steps {
		if p.Status != "" {
			t.Fatalf("step %s still has status %s after reset", step, p.Status)
		}
	}
}

func TestStateMachine_IsComplete(t *testing.T) {
	sm := newTestStateMachine(t)
	ctx := context.Background()
	ok, err := sm.IsComplete(ctx)
	if err != nil {
		t.Fatalf("IsComplete: %v", err)
	}
	if ok {
		t.Fatalf("fresh state should not be complete")
	}
	// Complete the wizard.
	_, _ = sm.Complete(ctx)
	ok, err = sm.IsComplete(ctx)
	if err != nil {
		t.Fatalf("IsComplete: %v", err)
	}
	if !ok {
		t.Fatalf("should be complete after Complete()")
	}
}

func TestStateMachine_StatePersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "onboarding.db")
	db1, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open1: %v", err)
	}
	sm1, err := NewStateMachine(db1)
	if err != nil {
		t.Fatalf("NewStateMachine1: %v", err)
	}
	ctx := context.Background()
	// Advance three times and set a step status.
	// From empty: 1→eula, 2→power_source, 3→permissions.
	_, _ = sm1.Advance(ctx)
	_, _ = sm1.Advance(ctx)
	_, _ = sm1.Advance(ctx)
	_, _ = sm1.SetStepStatus(ctx, StepEULA, StatusComplete, "v1")
	_ = db1.Close()
	// Reopen the same DB.
	db2, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open2: %v", err)
	}
	defer func() { _ = db2.Close() }()
	sm2, err := NewStateMachine(db2)
	if err != nil {
		t.Fatalf("NewStateMachine2: %v", err)
	}
	s, err := sm2.State(ctx)
	if err != nil {
		t.Fatalf("State: %v", err)
	}
	// After 3 advances from empty, we should be at permissions.
	if s.CurrentStep != StepPermissions {
		t.Fatalf("want permissions, got %s", s.CurrentStep)
	}
	if s.Steps[StepEULA].Status != StatusComplete {
		t.Fatalf("eula status lost: %s", s.Steps[StepEULA].Status)
	}
	if s.Steps[StepEULA].Data != "v1" {
		t.Fatalf("eula data lost: %s", s.Steps[StepEULA].Data)
	}
}

func TestStateMachine_AllStepsContainsComplete(t *testing.T) {
	// Sanity: AllSteps must end with StepComplete so Advance
	// naturally terminates there.
	if AllSteps[len(AllSteps)-1] != StepComplete {
		t.Fatalf("AllSteps does not end with StepComplete: %v", AllSteps)
	}
}
