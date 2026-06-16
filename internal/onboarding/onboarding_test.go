package onboarding

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

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

func TestStateMachine_StartsAtEULA(t *testing.T) {
	sm := newTestStateMachine(t)
	s, err := sm.State(context.Background())
	if err != nil {
		t.Fatalf("State: %v", err)
	}
	// Fresh state starts at "" (pre-launch). First Advance
	// lands on EULA, so the GUI renders EULA on first visit.
	if s.CurrentStep != "" {
		t.Fatalf("CurrentStep: want empty (pre-launch), got %s", s.CurrentStep)
	}
}

func TestStateMachine_AdvanceWalksThrough4Steps(t *testing.T) {
	sm := newTestStateMachine(t)
	ctx := context.Background()
	for _, want := range []Step{
		StepEULA, StepPermissions, StepHotkey, StepComplete,
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
	ctx := context.Background()
	// First advance to EULA, then back should stay (no-op from first).
	_, _ = sm.Advance(ctx) // "" → eula
	s, err := sm.Back(ctx)
	if err != nil {
		t.Fatalf("Back: %v", err)
	}
	if s.CurrentStep != StepEULA {
		t.Fatalf("should stay at eula, got %s", s.CurrentStep)
	}
}

func TestStateMachine_BackFromLaterStep(t *testing.T) {
	sm := newTestStateMachine(t)
	ctx := context.Background()
	_, _ = sm.Advance(ctx) // "" → eula
	_, _ = sm.Advance(ctx) // eula → permissions
	s, err := sm.Back(ctx)
	if err != nil {
		t.Fatalf("Back: %v", err)
	}
	if s.CurrentStep != StepEULA {
		t.Fatalf("back from permissions: want eula, got %s", s.CurrentStep)
	}
}

func TestStateMachine_SetStepStatus(t *testing.T) {
	sm := newTestStateMachine(t)
	ctx := context.Background()
	s, err := sm.SetStepStatus(ctx, StepEULA, StatusComplete, "v1")
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
	if got.Data != "v1" {
		t.Fatalf("data: %s", got.Data)
	}
}

func TestStateMachine_SkipAdvancesToNext(t *testing.T) {
	sm := newTestStateMachine(t)
	ctx := context.Background()
	// Advance to permissions first, then skip it.
	_, _ = sm.Advance(ctx) // "" → eula
	_, _ = sm.Advance(ctx) // eula → permissions
	s, err := sm.Skip(ctx, StepPermissions)
	if err != nil {
		t.Fatalf("Skip: %v", err)
	}
	if s.Steps[StepPermissions].Status != StatusSkipped {
		t.Fatalf("want skipped, got %s", s.Steps[StepPermissions].Status)
	}
	if s.CurrentStep != StepHotkey {
		t.Fatalf("should advance to hotkey after skipping permissions, got %s", s.CurrentStep)
	}
}

func TestStateMachine_CompleteMarksAll(t *testing.T) {
	sm := newTestStateMachine(t)
	ctx := context.Background()
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
	for _, step := range AllSteps {
		if s.Steps[step].Status == "" {
			t.Fatalf("step %s has no status", step)
		}
	}
}

func TestStateMachine_ResetReturnsToEULA(t *testing.T) {
	sm := newTestStateMachine(t)
	ctx := context.Background()
	for i := 0; i < 4; i++ {
		_, _ = sm.Advance(ctx)
	}
	_, _ = sm.Complete(ctx)
	s, err := sm.Reset(ctx)
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}
	// After reset, State() normalizes "" → StepEULA.
	if s.CurrentStep != "" {
		t.Fatalf("reset state should be empty pre-launch, got %s", s.CurrentStep)
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
	_, _ = sm1.Advance(ctx)
	_, _ = sm1.Advance(ctx)
	_, _ = sm1.Advance(ctx) // "" → eula → permissions → hotkey
	_, _ = sm1.SetStepStatus(ctx, StepEULA, StatusComplete, "v1")
	_ = db1.Close()

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
	if s.CurrentStep != StepHotkey {
		t.Fatalf("want hotkey, got %s", s.CurrentStep)
	}
	if s.Steps[StepEULA].Status != StatusComplete {
		t.Fatalf("eula status lost: %s", s.Steps[StepEULA].Status)
	}
	if s.Steps[StepEULA].Data != "v1" {
		t.Fatalf("eula data lost: %s", s.Steps[StepEULA].Data)
	}
}

func TestStateMachine_AllStepsLength(t *testing.T) {
	if len(AllSteps) != 4 {
		t.Fatalf("AllSteps length: want 4, got %d: %v", len(AllSteps), AllSteps)
	}
	if AllSteps[len(AllSteps)-1] != StepComplete {
		t.Fatalf("AllSteps must end with StepComplete: %v", AllSteps)
	}
	if AllSteps[0] != StepEULA {
		t.Fatalf("AllSteps must start with StepEULA: %v", AllSteps)
	}
}

// --- Migration tests ---

func TestMigrateState_WelcomeMapsToEULA(t *testing.T) {
	s := &State{
		CurrentStep: legacyWelcome,
		Steps: map[Step]StepProgress{
			legacyWelcome: {Status: StatusComplete},
		},
		StartedAt: time.Now(),
	}
	migrateState(s)
	if s.CurrentStep != StepEULA {
		t.Fatalf("welcome should map to eula, got %s", s.CurrentStep)
	}
}

func TestMigrateState_PowerSourceMapsToPermissions(t *testing.T) {
	s := &State{
		CurrentStep: legacyPowerSource,
		Steps: map[Step]StepProgress{
			StepEULA:          {Status: StatusComplete},
			legacyPowerSource: {Status: StatusPending},
		},
		StartedAt: time.Now(),
	}
	migrateState(s)
	if s.CurrentStep != StepPermissions {
		t.Fatalf("power_source should map to permissions, got %s", s.CurrentStep)
	}
}

func TestMigrateState_BackendDetectMapsToHotkey(t *testing.T) {
	s := &State{
		CurrentStep: legacyBackendDetect,
		Steps: map[Step]StepProgress{
			StepEULA:            {Status: StatusComplete},
			StepPermissions:     {Status: StatusComplete},
			legacyBackendDetect: {Status: StatusPending},
		},
		StartedAt: time.Now(),
	}
	migrateState(s)
	if s.CurrentStep != StepHotkey {
		t.Fatalf("backend_detect should map to hotkey, got %s", s.CurrentStep)
	}
}

func TestMigrateState_VoiceTestMapsToComplete(t *testing.T) {
	s := &State{
		CurrentStep: legacyVoiceTest,
		Steps: map[Step]StepProgress{
			StepEULA:        {Status: StatusComplete},
			StepPermissions: {Status: StatusSkipped},
			StepHotkey:      {Status: StatusComplete},
			legacyVoiceTest: {Status: StatusPending},
		},
		StartedAt: time.Now(),
	}
	migrateState(s)
	if s.CurrentStep != StepComplete {
		t.Fatalf("voice_test should map to complete, got %s", s.CurrentStep)
	}
}

func TestMigrateState_LegacyStepProgressTransferred(t *testing.T) {
	s := &State{
		CurrentStep: legacyBackendDetect,
		Steps: map[Step]StepProgress{
			legacyWelcome:     {Status: StatusComplete, Data: "seen"},
			legacyPowerSource: {Status: StatusSkipped},
		},
		StartedAt: time.Now(),
	}
	migrateState(s)
	if _, ok := s.Steps[legacyWelcome]; ok {
		t.Fatal("legacy welcome step should be deleted after migration")
	}
	if _, ok := s.Steps[legacyPowerSource]; ok {
		t.Fatal("legacy power_source step should be deleted after migration")
	}
	if s.Steps[StepEULA].Status != StatusComplete {
		t.Fatalf("eula should inherit welcome's complete status, got %s", s.Steps[StepEULA].Status)
	}
	if s.Steps[StepEULA].Data != "seen" {
		t.Fatalf("eula should inherit welcome's data, got %s", s.Steps[StepEULA].Data)
	}
}

func TestMigrateState_FullLegacyLoad(t *testing.T) {
	// Simulate a full legacy state persisted as JSON.
	legacy := &State{
		CurrentStep: legacyBackendDetect,
		Steps: map[Step]StepProgress{
			legacyWelcome:       {Status: StatusComplete},
			StepEULA:            {Status: StatusComplete, Data: "v1"},
			legacyPowerSource:   {Status: StatusSkipped},
			legacyBackendDetect: {Status: StatusPending},
		},
		StartedAt: time.Now(),
	}
	migrateState(legacy)
	if legacy.CurrentStep != StepHotkey {
		t.Fatalf("want hotkey, got %s (steps: %v)", legacy.CurrentStep, legacy.Steps)
	}
	if _, ok := legacy.Steps[legacyWelcome]; ok {
		t.Fatal("legacy welcome should be deleted")
	}
	if legacy.Steps[StepEULA].Status != StatusComplete {
		t.Fatalf("eula should be complete")
	}
	if legacy.Steps[StepPermissions].Status != StatusSkipped {
		t.Fatalf("permissions should inherit power_source skipped, got %s",
			legacy.Steps[StepPermissions].Status)
	}
}

func TestMigrateState_AlreadyMigrated(t *testing.T) {
	// A state that was already migrated should be idempotent.
	s := &State{
		CurrentStep: StepHotkey,
		Steps: map[Step]StepProgress{
			StepEULA:        {Status: StatusComplete, Data: "v1"},
			StepPermissions: {Status: StatusSkipped},
		},
		StartedAt: time.Now(),
	}
	before := s.CurrentStep
	beforeSteps := len(s.Steps)
	migrateState(s)
	if s.CurrentStep != before {
		t.Fatalf("migration should be idempotent: step changed from %s to %s", before, s.CurrentStep)
	}
	if len(s.Steps) != beforeSteps {
		t.Fatalf("migration should be idempotent: steps changed from %d to %d", beforeSteps, len(s.Steps))
	}
}

// --- EULA version check ---

func TestStateMachine_EULAVersionBumpForcesReAccept(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "onboarding.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = db.Close() }()
	sm, err := NewStateMachine(db)
	if err != nil {
		t.Fatalf("NewStateMachine: %v", err)
	}
	ctx := context.Background()

	// Mark EULA as accepted with an old version.
	_, _ = sm.SetStepStatus(ctx, StepEULA, StatusComplete, "v0")
	_, _ = sm.Advance(ctx)
	_, _ = sm.SetStepStatus(ctx, StepPermissions, StatusSkipped, "")
	_, _ = sm.Advance(ctx)

	// Now retrieve state — should be forced back to eula.
	s, err := sm.State(ctx)
	if err != nil {
		t.Fatalf("State: %v", err)
	}
	if s.CurrentStep != StepEULA {
		t.Fatalf("EULA version bump should reset to eula, got %s", s.CurrentStep)
	}
	if s.Steps[StepEULA].Status != StatusPending {
		t.Fatalf("EULA should be pending after version bump, got %s", s.Steps[StepEULA].Status)
	}
}

func TestStateMachine_EULAVersionMatchAllowsProgress(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "onboarding.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = db.Close() }()
	sm, err := NewStateMachine(db)
	if err != nil {
		t.Fatalf("NewStateMachine: %v", err)
	}
	ctx := context.Background()

	// Advance to EULA first, then accept it.
	_, _ = sm.Advance(ctx) // "" → eula
	_, _ = sm.SetStepStatus(ctx, StepEULA, StatusComplete, CurrentEULAVersion)
	_, _ = sm.Advance(ctx) // eula → permissions
	_, _ = sm.SetStepStatus(ctx, StepPermissions, StatusSkipped, "")
	_, _ = sm.Advance(ctx) // permissions → hotkey

	s, err := sm.State(ctx)
	if err != nil {
		t.Fatalf("State: %v", err)
	}
	if s.CurrentStep != StepHotkey {
		t.Fatalf("should stay at hotkey, got %s", s.CurrentStep)
	}
}

// --- DB-level migration from legacy JSON ---

func TestStateMachine_LegacyJSONInDBMigrated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "onboarding.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Manually insert legacy 8-step state.
	legacy := State{
		CurrentStep: legacyPowerSource,
		Steps: map[Step]StepProgress{
			legacyWelcome:     {Status: StatusComplete, Data: "ok"},
			StepEULA:          {Status: StatusComplete, Data: "v1"},
			legacyPowerSource: {Status: StatusPending},
		},
		StartedAt: time.Now(),
	}
	b, _ := json.Marshal(legacy)
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS onboarding_state (
		id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
		state_json TEXT NOT NULL DEFAULT '{}',
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	_, _ = db.Exec(`INSERT OR REPLACE INTO onboarding_state (id, state_json) VALUES (1, ?)`, string(b))

	sm, err := NewStateMachine(db)
	if err != nil {
		t.Fatalf("NewStateMachine: %v", err)
	}
	s, err := sm.State(context.Background())
	if err != nil {
		t.Fatalf("State: %v", err)
	}
	if s.CurrentStep != StepPermissions {
		t.Fatalf("legacy power_source should map to permissions, got %s", s.CurrentStep)
	}
	if s.Steps[StepEULA].Status != StatusComplete {
		t.Fatalf("eula should be complete after migration")
	}
	if _, ok := s.Steps[legacyWelcome]; ok {
		t.Fatal("legacy welcome step should be removed from map")
	}
}
