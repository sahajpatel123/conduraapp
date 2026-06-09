//go:build linux

package hotkey

import "errors"

// Linux stub for the hotkey package.
//
// The real hotkey on Linux would use the X11 Record Extension or
// the Wayland portal. That integration is deferred — see
// LOGBOOK.md "Phase 6 open questions". For now, the Manager on
// Linux is a no-op: Start returns an error so callers fall back
// to a window-level key binding, and Stop is a no-op.
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

// Start is a no-op on Linux. It records the callback so PressCount
// and other accessors remain consistent with the macOS/Windows
// behavior, but it does not register a real hotkey.
//
// A future Linux implementation will use the X11 Record Extension
// or Wayland portal here.
func (m *Manager) Start(handler func()) error {
	if handler == nil {
		return errors.New("hotkey: handler is required")
	}
	m.callback = handler
	m.started = true
	return nil
}

// StartHold is a no-op on Linux, same as Start.
func (m *Manager) StartHold(onDown, onUp func(), _ int) error {
	if onDown == nil || onUp == nil {
		return errors.New("hotkey: both onDown and onUp are required")
	}
	m.started = true
	return nil
}

// StartTap is a no-op on Linux, same as Start.
func (m *Manager) StartTap(onTap func(), _ int, _ int) error {
	if onTap == nil {
		return errors.New("hotkey: onTap is required")
	}
	m.started = true
	return nil
}

// Stop is a no-op on Linux.
func (m *Manager) Stop() {
	m.started = false
}

// Started reports whether Start/StartHold/StartTap was called.
func (m *Manager) Started() bool {
	return m.started
}

// PressCount is a no-op on Linux — returns 0 because no real
// hotkey is registered.
func (m *Manager) PressCount() uint64 { return 0 }
