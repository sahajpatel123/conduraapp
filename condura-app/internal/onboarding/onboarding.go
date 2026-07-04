// Package onboarding implements the backend state machine for the
// Onboarding Wizard. The wizard has a Svelte frontend
// (OnboardingWizard.svelte) — this package provides the Go-side
// state machine and persistence so the wizard can be resumed across
// daemon restarts.
//
// Converged state machine (Phase 14A, replacing the original 8-step
// MISSION §20 flow):
//
//	EULA → Permissions → Hotkey → Complete
//
// Each step has a status: pending, in_progress, complete, or
// skipped. The wizard advances via Advance; the user can
// navigate back via Back. State is persisted in the
// onboarding_state table.
//
// Migration: legacy 8-step states are transparently remapped to the
// 4-step converged flow on load so existing users don't lose their
// wizard position.
package onboarding

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Step is one stage of the onboarding flow.
type Step string

const (
	StepEULA        Step = "eula"
	StepPermissions Step = "permissions"
	StepHotkey      Step = "hotkey"
	StepComplete    Step = "complete"
)

// AllSteps is the canonical ordered list for the converged 4-step flow.
// GUIs render the step indicator from this slice.
var AllSteps = []Step{
	StepEULA,
	StepPermissions,
	StepHotkey,
	StepComplete,
}

// Legacy step constants kept for DB migration. New code must not
// reference these; they exist only so migrateState can remap old
// persisted state.
const (
	legacyWelcome       Step = "welcome"
	legacyPowerSource   Step = "power_source"
	legacyBackendDetect Step = "backend_detect"
	legacyVoiceTest     Step = "voice_test"
)

// Status is the per-step state.
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusComplete   Status = "complete"
	StatusSkipped    Status = "skipped"
)

// State is the persisted state of the wizard.
type State struct {
	CurrentStep Step                  `json:"current_step"`
	Steps       map[Step]StepProgress `json:"steps"`
	StartedAt   time.Time             `json:"started_at"`
	CompletedAt time.Time             `json:"completed_at,omitempty"`
}

// StepProgress is one step's metadata.
type StepProgress struct {
	Status    Status    `json:"status"`
	Data      string    `json:"data,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StateMachine owns the wizard state and persists it.
type StateMachine struct {
	mu sync.Mutex
	db *sql.DB
}

// NewStateMachine returns a StateMachine backed by the
// onboarding_state table.
func NewStateMachine(db *sql.DB) (*StateMachine, error) {
	if db == nil {
		return nil, errors.New("onboarding: db is required")
	}
	sm := &StateMachine{db: db}
	if err := sm.migrate(context.Background()); err != nil {
		return nil, fmt.Errorf("onboarding: migrate: %w", err)
	}
	return sm, nil
}

func (sm *StateMachine) migrate(ctx context.Context) error {
	_, err := sm.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS onboarding_state (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    state_json TEXT NOT NULL DEFAULT '{}',
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
INSERT OR IGNORE INTO onboarding_state (id, state_json) VALUES (1, '{}');
`)
	if err != nil {
		return err
	}
	return nil
}

// State returns the current wizard state. If no state has been
// recorded yet, returns a fresh state starting at StepEULA.
// Legacy 8-step states are transparently migrated to the new
// 4-step flow on load.
func (sm *StateMachine) State(ctx context.Context) (*State, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	s, err := sm.loadLocked(ctx)
	if err != nil {
		return nil, err
	}
	// EULA version check: if the user previously accepted a
	// different EULA version, force them back to the EULA step
	// so they re-accept.
	if s.Steps[StepEULA].Status == StatusComplete {
		acceptedV := s.Steps[StepEULA].Data
		currentV := CurrentEULAVersion
		if acceptedV != "" && acceptedV != currentV {
			s.CurrentStep = StepEULA
			s.Steps[StepEULA] = StepProgress{
				Status: StatusPending, UpdatedAt: time.Now().UTC(),
			}
		}
	}
	return s, nil
}

// Advance moves the wizard forward to the next step. The caller
// is responsible for setting the current step's status to complete
// first via SetStepStatus.
func (sm *StateMachine) Advance(ctx context.Context) (*State, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	s, err := sm.loadLocked(ctx)
	if err != nil {
		return nil, err
	}
	idx := stepIndex(s.CurrentStep)
	if idx < 0 {
		// Pre-launch — start at the first real step (EULA).
		s.CurrentStep = AllSteps[0]
		return s, sm.saveLocked(ctx, s)
	}
	if idx+1 >= len(AllSteps) {
		return s, nil // already at end
	}
	s.CurrentStep = AllSteps[idx+1]
	return s, sm.saveLocked(ctx, s)
}

// Back moves the wizard backward. No-op if at the first step.
func (sm *StateMachine) Back(ctx context.Context) (*State, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	s, err := sm.loadLocked(ctx)
	if err != nil {
		return nil, err
	}
	idx := stepIndex(s.CurrentStep)
	if idx <= 0 {
		return s, nil
	}
	s.CurrentStep = AllSteps[idx-1]
	return s, sm.saveLocked(ctx, s)
}

// SetStepStatus records the status (and optional data) for a step.
func (sm *StateMachine) SetStepStatus(ctx context.Context, step Step, status Status, data string) (*State, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	s, err := sm.loadLocked(ctx)
	if err != nil {
		return nil, err
	}
	if s.Steps == nil {
		s.Steps = make(map[Step]StepProgress)
	}
	s.Steps[step] = StepProgress{Status: status, Data: data, UpdatedAt: time.Now().UTC()}
	if status == StatusComplete && step == StepComplete {
		s.CompletedAt = time.Now().UTC()
	}
	return s, sm.saveLocked(ctx, s)
}

// Skip marks a step as skipped and advances PAST the skipped
// step (not from the current step). This matters when the user
// is on StepEULA and skips StepPermissions — they should land
// on StepHotkey, not StepPermissions.
func (sm *StateMachine) Skip(ctx context.Context, step Step) (*State, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	s, err := sm.loadLocked(ctx)
	if err != nil {
		return nil, err
	}
	if s.Steps == nil {
		s.Steps = make(map[Step]StepProgress)
	}
	// Mark the step as skipped.
	s.Steps[step] = StepProgress{
		Status:    StatusSkipped,
		UpdatedAt: time.Now().UTC(),
	}
	// Advance CurrentStep PAST the skipped step.
	idx := stepIndex(step)
	if idx < 0 {
		// Unknown step — fall back to "advance from current".
		return s, sm.saveLocked(ctx, s)
	}
	if idx+1 < len(AllSteps) {
		s.CurrentStep = AllSteps[idx+1]
	} else {
		s.CurrentStep = AllSteps[len(AllSteps)-1]
	}
	return s, sm.saveLocked(ctx, s)
}

// Complete marks the entire wizard done. Unfinished steps are
// auto-skipped.
func (sm *StateMachine) Complete(ctx context.Context) (*State, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	s, err := sm.loadLocked(ctx)
	if err != nil {
		return nil, err
	}
	for _, step := range AllSteps {
		if s.Steps[step].Status == "" {
			s.Steps[step] = StepProgress{Status: StatusSkipped, UpdatedAt: time.Now().UTC()}
		}
	}
	s.Steps[StepComplete] = StepProgress{Status: StatusComplete, UpdatedAt: time.Now().UTC()}
	s.CompletedAt = time.Now().UTC()
	s.CurrentStep = StepComplete
	return s, sm.saveLocked(ctx, s)
}

// Reset wipes the state and returns a fresh wizard at StepEULA.
func (sm *StateMachine) Reset(ctx context.Context) (*State, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	s := sm.freshState()
	return s, sm.saveLocked(ctx, s)
}

// IsComplete returns true if the user finished the wizard.
func (sm *StateMachine) IsComplete(ctx context.Context) (bool, error) {
	s, err := sm.State(ctx)
	if err != nil {
		return false, err
	}
	if s.Steps == nil {
		return false, nil
	}
	return s.Steps[StepComplete].Status == StatusComplete, nil
}

// loadLocked reads the state and applies legacy migration.
// Caller must hold sm.mu.
func (sm *StateMachine) loadLocked(ctx context.Context) (*State, error) {
	var raw string
	err := sm.db.QueryRowContext(ctx,
		`SELECT state_json FROM onboarding_state WHERE id = 1`,
	).Scan(&raw)
	if err != nil {
		return nil, fmt.Errorf("onboarding: query: %w", err)
	}
	var s State
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return sm.freshState(), nil
	}
	if s.Steps == nil {
		s.Steps = make(map[Step]StepProgress)
	}
	migrateState(&s)
	return &s, nil
}

// saveLocked persists the state. Caller must hold sm.mu.
func (sm *StateMachine) saveLocked(ctx context.Context, s *State) error {
	b, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("onboarding: marshal: %w", err)
	}
	_, err = sm.db.ExecContext(ctx,
		`UPDATE onboarding_state SET state_json = ?, updated_at = datetime('now') WHERE id = 1`,
		string(b),
	)
	if err != nil {
		return fmt.Errorf("onboarding: update: %w", err)
	}
	return nil
}

// freshState returns an empty state. CurrentStep is empty so
// the first Advance lands on StepEULA (the wizard starts
// before step 0, like the old Welcome step was).
func (sm *StateMachine) freshState() *State {
	return &State{
		CurrentStep: "",
		Steps:       make(map[Step]StepProgress),
		StartedAt:   time.Now().UTC(),
	}
}

func stepIndex(s Step) int {
	for i, st := range AllSteps {
		if st == s {
			return i
		}
	}
	return -1
}

// migrateState remaps legacy 8-step positions to the converged
// 4-step flow. Called on every load so existing users don't lose
// wizard progress after the Phase 14A update.
func migrateState(s *State) {
	// Capture whether the state had an explicit legacy CurrentStep
	// so we don't override it with the "next-incomplete" heuristic
	// below. If the user was on Welcome, we want to land them on
	// EULA, not on Permissions.
	hadLegacyCurrent := false
	cs := s.CurrentStep

	// Map legacy step names to the nearest converged step.
	switch cs {
	case legacyWelcome:
		// Welcome is "before step 0" — map to EULA so the
		// user sees the first real step. The old Welcome
		// was informational only.
		s.CurrentStep = StepEULA
		hadLegacyCurrent = true
	case legacyPowerSource:
		if s.Steps[StepEULA].Status == StatusComplete ||
			s.Steps[legacyWelcome].Status == StatusComplete {
			s.CurrentStep = StepPermissions
		} else {
			s.CurrentStep = StepEULA
		}
		hadLegacyCurrent = true
	case legacyBackendDetect:
		s.CurrentStep = StepHotkey
		hadLegacyCurrent = true
	case legacyVoiceTest:
		s.CurrentStep = StepComplete
		hadLegacyCurrent = true
	}

	// Migrate per-step progress: map legacy step statuses to
	// their converged equivalents so the wizard knows what's
	// been done.
	legacyMap := map[Step]Step{
		legacyWelcome:       StepEULA,
		legacyPowerSource:   StepPermissions,
		legacyBackendDetect: StepHotkey,
		legacyVoiceTest:     StepComplete,
	}
	for legacy, modern := range legacyMap {
		if lp, ok := s.Steps[legacy]; ok {
			if _, exists := s.Steps[modern]; !exists {
				s.Steps[modern] = lp
			}
			delete(s.Steps, legacy)
		}
	}

	// Heuristic ONLY for legacy migrations: when the user's old
	// 8-step state was at a step that no longer exists, advance
	// to the next-incomplete converged step. For non-legacy
	// (4-step) state, trust the saved CurrentStep — the user
	// explicitly advanced there and we shouldn't bump them
	// back to permissions.
	if hadLegacyCurrent {
		if s.Steps[StepEULA].Status == StatusComplete &&
			s.Steps[StepHotkey].Status != StatusComplete &&
			s.Steps[StepHotkey].Status != StatusSkipped &&
			s.CurrentStep != StepEULA {
			if s.Steps[StepPermissions].Status != StatusComplete &&
				s.Steps[StepPermissions].Status != StatusSkipped {
				s.CurrentStep = StepPermissions
			} else {
				s.CurrentStep = StepHotkey
			}
		}
	}

	// If Complete is done, normalize current step.
	if s.Steps[StepComplete].Status == StatusComplete {
		s.CurrentStep = StepComplete
	}
}
