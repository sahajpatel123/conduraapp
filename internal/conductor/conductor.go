// Package conductor wires the global hotkey to the overlay presence.
//
// The conductor is a thin glue layer: it owns no state of its own, it
// just translates hotkey presses into overlay lifecycle events. The
// presence orchestrator (internal/presence) owns the state machine.
//
// Typical wiring:
//
//	mgr := hotkey.New("Cmd+Shift+Space")
//	overlay := overlay.NewNoopController()
//	orchestrator := presence.NewOrchestrator(overlay, haltFlag, nil)
//	conductor := conductor.New(mgr, orchestrator, onShow, onHide)
//
//	if err := conductor.Start(); err != nil {
//	    log.Fatal(err)
//	}
//	defer conductor.Stop()
//
// The conductor is the only place that knows the user is asking for
// the overlay. Everything else (the orchestrator, the overlay, the
// hotkey) is independent and testable in isolation.
package conductor

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/sahajpatel123/synapticapp/internal/hotkey"
	"github.com/sahajpatel123/synapticapp/internal/overlay"
	"github.com/sahajpatel123/synapticapp/internal/presence"
)

// ErrNoHotkey is returned by Start when the underlying hotkey manager
// is nil.
var ErrNoHotkey = errors.New("conductor: hotkey manager is required")

// ErrNoOrchestrator is returned by Start when the orchestrator is nil.
var ErrNoOrchestrator = errors.New("conductor: orchestrator is required")

// Conductor wires a hotkey to a presence orchestrator. When the user
// presses the hotkey, the conductor calls orchestrator.Summon(); when
// the user presses again, it calls orchestrator.Dismiss().
//
// The "press again" semantics match the platform convention: a single
// hotkey toggle. The orchestrator decides whether the second press
// summons or dismisses based on its own state.
type Conductor struct {
	hotkey       *hotkey.Manager
	orchestrator *presence.Orchestrator
	onShow       func()
	onHide       func()

	started atomic.Bool
}

// New creates a Conductor. hotkey and orchestrator are required; the
// optional onShow/onHide callbacks fire after each summon/dismiss and
// are typically used to update the tray tooltip or play a sound.
func New(hk *hotkey.Manager, orch *presence.Orchestrator, onShow, onHide func()) (*Conductor, error) {
	if hk == nil {
		return nil, ErrNoHotkey
	}
	if orch == nil {
		return nil, ErrNoOrchestrator
	}
	return &Conductor{
		hotkey:       hk,
		orchestrator: orch,
		onShow:       onShow,
		onHide:       onHide,
	}, nil
}

// Start registers the hotkey and begins listening. Each press calls
// the toggle function.
func (c *Conductor) Start() error {
	if c.started.Load() {
		return errors.New("conductor: already started")
	}

	handler := c.toggle
	if err := c.hotkey.Start(handler); err != nil {
		return err
	}
	c.started.Store(true)
	return nil
}

// toggle is invoked on every hotkey press. It summons the orchestrator
// if not active, or dismisses it if active. This is the user-facing
// "press to toggle" semantics.
func (c *Conductor) toggle() {
	// Use a background context — the hotkey is global, the
	// orchestrator's lifetime is owned by the caller, and a
	// hotkey press should never block on a canceled context.
	//nolint:contextcheck // hotkey has no parent context
	if c.orchestrator.IsActive() {
		c.orchestrator.Dismiss()
		if c.onHide != nil {
			c.onHide()
		}
		return
	}
	// Best-effort summon. The orchestrator returns an error only
	// when the kill switch is engaged; in that case the hotkey
	// silently does nothing, which is the right UX.
	if err := c.orchestrator.Summon(backgroundCtx()); err == nil && c.onShow != nil {
		c.onShow()
	}
}

// Stop unregisters the hotkey. Safe to call multiple times.
func (c *Conductor) Stop() {
	if !c.started.Load() {
		return
	}
	c.hotkey.Stop()
	c.started.Store(false)
}

// State returns the current overlay state. Convenience accessor.
func (c *Conductor) State() overlay.State {
	return c.orchestrator.State()
}

// backgroundCtx returns a fresh, never-canceled context. Hotkey
// presses are not scoped to any request lifetime.
func backgroundCtx() context.Context { return context.Background() }
