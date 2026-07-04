//go:build linux

package hotkey

import "errors"

// errLinuxUnsupported is returned by Start/StartHold/StartTap on Linux
// because no global hotkey is registered in v0.1.x (real X11 Record
// Extension / Wayland portal integration is v0.2.0). Returning an
// error (rather than silently succeeding) lets callers fall back to a
// window-level binding or the TUI honestly — the Wails app logs
// "conductor start failed" and continues without a global hotkey.
var errLinuxUnsupported = errors.New("hotkey: global hotkey not supported on Linux in v0.1.x; use condura-tui or bind a desktop-environment shortcut")

// Linux stub for the hotkey package.
//
// The real hotkey on Linux would use the X11 Record Extension or
// the Wayland portal. That integration is deferred to v0.2.0. For now
// the Manager on Linux fails Start/StartHold/StartTap with
// errLinuxUnsupported so callers fall back to a window-level key
// binding or the TUI; Stop is a no-op.
//
// This file exists so that packages which import hotkey (such as
// the conductor) compile on Linux for CI/lint/test purposes.

// Manager is a Linux no-op hotkey manager.
type Manager struct {
	spec     string
	started  bool
	callback func()
}

// New constructs a Manager. The spec is recorded but unused on
// Linux because there is no real registration.
func New(spec string) *Manager {
	return &Manager{spec: spec}
}

// Start returns errLinuxUnsupported on Linux. The callback is recorded
// for accessors but no real hotkey is registered and Started() stays
// false. Callers (the conductor) handle the error and fall back to a
// window-level binding; the Wails app logs + continues without a
// global hotkey. A future Linux implementation will use the X11 Record
// Extension or Wayland portal here.
func (m *Manager) Start(handler func()) error {
	if handler == nil {
		return errors.New("hotkey: handler is required")
	}
	m.callback = handler
	return errLinuxUnsupported
}

// StartHold returns errLinuxUnsupported on Linux, same as Start.
func (m *Manager) StartHold(onDown, onUp func(), _ int) error {
	if onDown == nil || onUp == nil {
		return errors.New("hotkey: both onDown and onUp are required")
	}
	return errLinuxUnsupported
}

// StartTap returns errLinuxUnsupported on Linux, same as Start.
func (m *Manager) StartTap(onTap func(), _ int, _ int) error {
	if onTap == nil {
		return errors.New("hotkey: onTap is required")
	}
	return errLinuxUnsupported
}

// Stop is a no-op on Linux.
func (m *Manager) Stop() {
	m.started = false
}

// Started reports whether Start/StartHold/StartTap registered a real
// hotkey. On Linux this is always false because Start errors and does
// not set started.
func (m *Manager) Started() bool {
	return m.started
}

// PressCount is a no-op on Linux — returns 0 because no real
// hotkey is registered.
func (m *Manager) PressCount() uint64 { return 0 }
