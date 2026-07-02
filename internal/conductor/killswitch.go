// Package conductor — kill-switch wiring.
//
// The KillSwitchConductor is a sibling of Conductor. It binds a SECOND
// global hotkey to halt.Flag.Halt, so the kill switch (CLAUDE.md §5.3
// Layer 1: hard hotkey, default Cmd+Shift+Escape) is wired even when
// the overlay hotkey (Layer 0 / quick-prompt) is bound separately.
//
// Rationale: hotkey.Manager wraps exactly one x/hotkey registration,
// and the overlay hotkey must remain user-configurable independently
// of the kill switch. A second Manager is the cleanest isolation —
// it cannot block, conflict with, or be stopped by the overlay
// conductor.
package conductor

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/sahajpatel123/conduraapp/internal/halt"
	"github.com/sahajpatel123/conduraapp/internal/hotkey"
)

// ErrNoHaltFlag is returned by Start when the halt flag is nil.
var ErrNoHaltFlag = errors.New("killswitch: halt flag is required")

// ReasonHardHotkey is the audit-friendly reason set on the halt flag
// when the user invokes the documented kill-switch combo. Other
// subsystems (anomaly detector, watchdog) use their own reasons.
const ReasonHardHotkey = "hard_hotkey"

// KillSwitchConductor binds a global hotkey to halt.Flag.Halt.
//
// The hotkey spec comes from cfg.Hotkey.KillSwitch (default
// Cmd+Shift+Escape on macOS, Ctrl+Alt+\\ on Win/Linux per CLAUDE.md).
// On every press, the conductor calls haltFlag.Halt with
// ReasonHardHotkey so the audit log records the cause.
//
// The conductor is fail-closed: if the hotkey cannot be registered,
// Start returns the error and the calling code must surface it to
// the user. We never silently drop the kill switch.
type KillSwitchConductor struct {
	spec     string
	haltFlag *halt.Flag
	onHalt   func()
	mgr      *hotkey.Manager

	started atomic.Bool
}

// NewKillSwitch constructs a KillSwitchConductor. spec is the
// hotkey combo (e.g., "Cmd+Shift+Escape"); haltFlag is required
// and is the destination of the Halt call. The optional onHalt
// callback fires after the halt is recorded (use this to surface
// the overlay or play a sound).
func NewKillSwitch(spec string, haltFlag *halt.Flag, onHalt func()) (*KillSwitchConductor, error) {
	if haltFlag == nil {
		return nil, ErrNoHaltFlag
	}
	return &KillSwitchConductor{
		spec:     spec,
		haltFlag: haltFlag,
		onHalt:   onHalt,
		mgr:      hotkey.New(spec),
	}, nil
}

// Start registers the kill-switch hotkey and begins listening.
//
// Returns an error if registration fails. The hotkey library
// refuses to register duplicates, so calling Start twice returns
// an error. Callers must wire Stop at shutdown to release the
// registration cleanly.
func (k *KillSwitchConductor) Start() error {
	if k.started.Load() {
		return errors.New("killswitch: already started")
	}
	if err := k.mgr.Start(k.fire); err != nil {
		return err
	}
	k.started.Store(true)
	return nil
}

// fire is invoked on every kill-switch press. It calls Halt with
// ReasonHardHotkey so the audit log + UI both show "halted by hard
// hotkey". We use a fresh background context because the hotkey is
// global and not scoped to any request lifetime.
//
// We ignore the returned previous-state on purpose — the call is
// idempotent (idempotently halting an already-halted flag is fine)
// and the operator (user) does not need a return value.
func (k *KillSwitchConductor) fire() {
	ctx := context.Background()
	_, _ = k.haltFlag.Halt(ctx, ReasonHardHotkey)
	if k.onHalt != nil {
		k.onHalt()
	}
}

// Stop unregisters the kill-switch hotkey. Safe to call multiple
// times. Safe to call before Start.
func (k *KillSwitchConductor) Stop() {
	if !k.started.Load() {
		return
	}
	k.mgr.Stop()
	k.started.Store(false)
}

// Started reports whether the kill-switch hotkey is currently
// registered. Used by tests and by the tray's "kill switch active"
// indicator.
func (k *KillSwitchConductor) Started() bool {
	return k.started.Load()
}
