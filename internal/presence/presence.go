// Package presence orchestrates the gesture → session lifecycle.
//
// It connects the hotkey hold gesture to overlay, voice capture, and the
// agent loop. The orchestrator owns the full lifecycle:
//
//	Hotkey down → overlay.show + voice.startCapture → Listening
//	Whisper partials → voice.partial → live transcript
//	Hotkey up → voice.stopCapture → voice.final → agent.ask → Thinking
//	Agent response → voice.speaking → Speaking
//	Done → overlay.hide → Hidden
package presence

import (
	"context"
	"errors"
	"sync"

	"github.com/sahajpatel123/synapticapp/internal/overlay"
)

// ErrHalted is returned when a presence action is blocked by the kill switch.
var ErrHalted = errors.New("presence: kill switch is active")

// HaltChecker is a minimal interface for checking the kill switch state.
// This avoids a direct dependency on the halt package's database-backed Flag.
type HaltChecker interface {
	IsHalted() bool
}

// Orchestrator manages the presence session lifecycle.
type Orchestrator struct {
	overlay overlay.Controller
	halt    HaltChecker

	mu     sync.Mutex
	active bool
	cancel context.CancelFunc
}

// NewOrchestrator creates a presence orchestrator.
func NewOrchestrator(ctrl overlay.Controller, haltChecker HaltChecker) *Orchestrator {
	return &Orchestrator{
		overlay: ctrl,
		halt:    haltChecker,
	}
}

// Summon starts a presence session (equivalent to hotkey down).
// Shows the overlay and begins capture. Returns an error if already active
// or if the kill switch is engaged.
func (o *Orchestrator) Summon(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.active {
		return nil // already active, no-op
	}

	// Check kill switch before starting.
	if o.halt != nil && o.halt.IsHalted() {
		return ErrHalted
	}

	ctx, cancel := context.WithCancel(ctx)
	o.cancel = cancel
	o.active = true

	// Show overlay.
	if err := o.overlay.Show(ctx, overlay.ShowOpts{AtCursor: true}); err != nil {
		cancel()
		o.active = false
		return err
	}

	return nil
}

// Dismiss ends the current presence session (equivalent to hotkey up or Esc).
// Hides the overlay and cancels any active capture.
func (o *Orchestrator) Dismiss() {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.active {
		return
	}

	if o.cancel != nil {
		o.cancel()
	}
	_ = o.overlay.Hide()
	o.active = false
	o.cancel = nil
}

// IsActive reports whether a presence session is currently active.
func (o *Orchestrator) IsActive() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.active
}

// State returns the current overlay state.
func (o *Orchestrator) State() overlay.State {
	return o.overlay.State()
}
