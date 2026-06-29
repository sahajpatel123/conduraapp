// Package presence orchestrates the gesture → session lifecycle.
//
// It connects the hotkey hold gesture to overlay, voice capture, and the
// agent loop. The orchestrator owns the full lifecycle:
//
//	Hotkey down → overlay.show + capture.Start → Listening
//	Whisper partials → voice.partial → live transcript
//	Hotkey up → capture.Stop → voice.final → agent.ask → Thinking
//	Agent response → voice.speaking → Speaking
//	Done → overlay.hide → Hidden
//
// The capture is injected so the orchestrator can be tested without
// a real microphone. The real voice pipeline (Phase 6B) will inject
// a capture that wraps the voice.Recorder.
package presence

import (
	"context"
	"errors"
	"sync"

	"github.com/sahajpatel123/conduraapp/internal/overlay"
)

// ErrHalted is returned when a presence action is blocked by the kill switch.
var ErrHalted = errors.New("presence: kill switch is active")

// HaltChecker is a minimal interface for checking the kill switch state.
// This avoids a direct dependency on the halt package's database-backed Flag.
type HaltChecker interface {
	IsHalted() bool
}

// Capture is the seam for microphone capture. Implementations begin
// capture on Start and return the recorded audio on Stop. The
// orchestrator calls Start in Summon and Stop in Dismiss.
//
// Returning a non-nil error from Start aborts the presence session;
// returning a non-nil error from Stop propagates it to the caller
// of Dismiss (which currently swallows errors — see Dismiss).
type Capture interface {
	Start(ctx context.Context) error
	Stop() error
}

// CaptureFuncs adapts plain functions to the Capture interface, so
// callers can wire the orchestrator without declaring a new type.
type CaptureFuncs struct {
	StartFn func(ctx context.Context) error
	StopFn  func() error
}

// Start calls StartFn.
func (c CaptureFuncs) Start(ctx context.Context) error { return c.StartFn(ctx) }

// Stop calls StopFn.
func (c CaptureFuncs) Stop() error { return c.StopFn() }

// Orchestrator manages the presence session lifecycle.
type Orchestrator struct {
	overlay overlay.Controller
	capture Capture
	halt    HaltChecker

	mu     sync.Mutex
	active bool
	cancel context.CancelFunc
}

// NewOrchestrator creates a presence orchestrator. The capture is
// optional (pass nil for an overlay-only session). When non-nil, it
// is started by Summon and stopped by Dismiss.
func NewOrchestrator(ctrl overlay.Controller, haltChecker HaltChecker, capture Capture) *Orchestrator {
	return &Orchestrator{
		overlay: ctrl,
		capture: capture,
		halt:    haltChecker,
	}
}

// Summon starts a presence session (equivalent to hotkey down).
// Shows the overlay and begins capture. Returns an error if already active
// or if the kill switch is engaged. If capture.Start fails, the overlay
// is hidden and the session is reset.
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

	runCtx, cancel := context.WithCancel(ctx)
	o.cancel = cancel
	o.active = true

	// Show overlay first — capture failure is recoverable (the
	// overlay is still useful even without a working mic).
	if err := o.overlay.Show(runCtx, overlay.ShowOpts{AtCursor: true}); err != nil {
		cancel()
		o.active = false
		o.cancel = nil
		return err
	}

	// Begin capture if a capture is wired. Capture failure rolls
	// back the overlay show.
	if o.capture != nil {
		if err := o.capture.Start(runCtx); err != nil {
			_ = o.overlay.Hide()
			cancel()
			o.active = false
			o.cancel = nil
			return err
		}
	}

	return nil
}

// Dismiss ends the current presence session (equivalent to hotkey up or Esc).
// Hides the overlay, stops any active capture, and cancels the run
// context. Stop errors are swallowed (the user's intent is to end
// the session; a capture-stop error is not actionable).
func (o *Orchestrator) Dismiss() {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.active {
		return
	}

	if o.cancel != nil {
		o.cancel()
		o.cancel = nil
	}

	if o.capture != nil {
		_ = o.capture.Stop()
	}

	_ = o.overlay.Hide()
	o.active = false
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
