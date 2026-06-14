// Package onboarding implements the backend state machine for
// the Onboarding Wizard (Phase 11, sub-phase 11E). The wizard
// already has a Svelte frontend (OnboardingWizard.svelte) —
// this package provides the Go-side state machine and
// persistence so the wizard can be resumed across daemon
// restarts.
//
// State machine (per MISSION §20):
//
//	Welcome → EULA → PowerSource → Permissions → BackendDetect
//	→ Hotkey → VoiceTest → Complete
//
// Each step has a status: pending, in_progress, complete, or
// skipped. The wizard advances via Advance; the user can
// navigate back via Back. State is persisted in the
// onboarding_state table.
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
	StepWelcome       Step = "welcome"
	StepEULA          Step = "eula"
	StepPowerSource   Step = "power_source"
	StepPermissions   Step = "permissions"
	StepBackendDetect Step = "backend_detect"
	StepHotkey        Step = "hotkey"
	StepVoiceTest     Step = "voice_test"
	StepComplete      Step = "complete"
)

// AllSteps is the canonical ordered list. Use this in the
// frontend to render the step indicator.
var AllSteps = []Step{
	StepWelcome,
	StepEULA,
	StepPowerSource,
	StepPermissions,
	StepBackendDetect,
	StepHotkey,
	StepVoiceTest,
	StepComplete,
}

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
// recorded yet, returns a fresh state starting at StepWelcome.
func (sm *StateMachine) State(ctx context.Context) (*State, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
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
	// Normalize: an empty CurrentStep means the wizard is
	// pre-launch (the very first time). The fresh state should
	// point at StepWelcome so the GUI's "next" button works.
	if s.CurrentStep == "" {
		s.CurrentStep = StepWelcome
	}
	return &s, nil
}

// Advance moves the wizard forward to the next step. The
// caller is responsible for setting the current step's status
// to complete first via SetStepStatus.
func (sm *StateMachine) Advance(ctx context.Context) (*State, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	s, err := sm.loadLocked(ctx)
	if err != nil {
		return nil, err
	}
	idx := stepIndex(s.CurrentStep)
	// Empty / unknown current means the wizard is pre-launch
	// (the user has not clicked "Next" from the Welcome
	// screen yet). The first Advance lands on EULA — the
	// Welcome screen is implicitly "before step 0".
	if idx < 0 {
		s.CurrentStep = AllSteps[1] // eula
		return s, sm.saveLocked(ctx, s)
	}
	if idx+1 >= len(AllSteps) {
		// Already at the end.
		return s, nil
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
	// Before step 0 — already at the implicit start.
	if idx <= 0 {
		// Normalize: present as Welcome on the read side.
		s.CurrentStep = StepWelcome
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

// Skip marks a step as skipped (used for "I don't want to do
// voice test now, finish onboarding").
func (sm *StateMachine) Skip(ctx context.Context, step Step) (*State, error) {
	return sm.SetStepStatus(ctx, step, StatusSkipped, "")
}

// Complete marks the entire wizard done.
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

// Reset wipes the state and returns a fresh wizard.
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

// loadLocked reads the state from the DB. Caller must hold sm.mu.
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
		// Corrupt — start fresh.
		return sm.freshState(), nil
	}
	if s.Steps == nil {
		s.Steps = make(map[Step]StepProgress)
	}
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

// freshState returns an empty state at StepWelcome.
func (sm *StateMachine) freshState() *State {
	return &State{
		CurrentStep: StepWelcome,
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
