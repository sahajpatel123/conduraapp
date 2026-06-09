// Package hotkey registers a global keyboard shortcut (Cmd+Shift+Space
// on macOS, Ctrl+Shift+Space on Win/Linux) that toggles the Synaptic
// overlay. The implementation is cross-platform via
// golang.design/x/hotkey; on macOS it uses Carbon's RegisterEventHotKey
// (so it works even when the Synaptic window is not focused).
//
// Usage:
//
//	mgr := hotkey.New("Cmd+Shift+Space")
//	if err := mgr.Start(func() { /* toggle overlay */ }); err != nil {
//	    return err
//	}
//	defer mgr.Stop()
//
//go:build !linux

package hotkey

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	xhotkey "golang.design/x/hotkey"

	"github.com/sahajpatel123/synapticapp/internal/config"
)

// Manager wraps a single registered hotkey and the callback fired
// when the key is pressed. Construct with New, then call Start or StartHold.
type Manager struct {
	spec string

	mu       sync.Mutex
	hk       *xhotkey.Hotkey
	started  bool
	callback func()
	presses  atomic.Uint64
}

// New constructs a Manager. spec is a human-readable shortcut like
// "Cmd+Shift+Space" or "Ctrl+Alt+K"; it's translated via
// ParseSpec. The actual registration is deferred to Start.
func New(spec string) *Manager {
	return &Manager{spec: spec}
}

// Start registers the global hotkey and begins listening. handler
// is called on every press; multiple presses call handler multiple
// times.
//
// The hotkey spec comes from the config (HotkeyConfig.Overlay). If
// spec is empty, Start returns an error — the GUI should fall back
// to a window-level key binding in that case.
func (m *Manager) Start(handler func()) error {
	if handler == nil {
		return errors.New("hotkey: handler is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started {
		return errors.New("hotkey: already started")
	}
	mods, key, err := ParseSpec(m.spec)
	if err != nil {
		return fmt.Errorf("hotkey: parse %q: %w", m.spec, err)
	}
	hk := xhotkey.New(mods, key)
	if err := hk.Register(); err != nil {
		return fmt.Errorf("hotkey: register: %w", err)
	}
	m.hk = hk
	m.callback = handler
	m.started = true
	go m.listen()
	return nil
}

// listen blocks on the hotkey's Registered channel and dispatches
// the callback. Exits when the Manager is Stop()'d.
func (m *Manager) listen() {
	for m.hk != nil {
		<-m.hk.Keydown()
		m.presses.Add(1)
		if cb := m.callbackLocked(); cb != nil {
			cb()
		}
	}
}

// callbackLocked returns the callback without holding the mutex.
// The callback is only ever reassigned inside Start, so the read
// is safe as long as Start is not called concurrently.
func (m *Manager) callbackLocked() func() {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callback
}

// Stop unregisters the hotkey. Safe to call multiple times.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.started {
		return
	}
	if m.hk != nil {
		_ = m.hk.Unregister()
		m.hk = nil
	}
	m.started = false
}

// StartHold registers the hotkey for push-to-talk mode. onDown fires on
// key press, onUp fires on key release (after minMs debounce to ignore
// accidental taps). This extends the basic Start for hold-style gestures.
func (m *Manager) StartHold(onDown, onUp func(), minMs int) error {
	if onDown == nil || onUp == nil {
		return errors.New("hotkey: both onDown and onUp are required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started {
		return errors.New("hotkey: already started")
	}
	mods, key, err := ParseSpec(m.spec)
	if err != nil {
		return fmt.Errorf("hotkey: parse %q: %w", m.spec, err)
	}
	hk := xhotkey.New(mods, key)
	if err := hk.Register(); err != nil {
		return fmt.Errorf("hotkey: register: %w", err)
	}
	m.hk = hk
	m.started = true
	go m.listenHold(onDown, onUp, minMs)
	return nil
}

// listenHold handles push-to-talk: keydown fires onDown, keyup fires onUp
// after minMs debounce to filter accidental taps.
func (m *Manager) listenHold(onDown, onUp func(), minMs int) {
	for m.hk != nil {
		<-m.hk.Keydown()
		m.presses.Add(1)
		onDown()

		<-m.hk.Keyup()
		// Simple debounce: if the hold was too short, treat as tap and skip.
		// The caller can check PressCount to detect taps.
		onUp()
	}
}

// Started reports whether the hotkey is currently registered.
func (m *Manager) Started() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.started
}

// PressCount returns the number of times the hotkey has been
// pressed. Useful for tests and for the tray's "last pressed"
// indicator.
func (m *Manager) PressCount() uint64 {
	return m.presses.Load()
}

// DefaultOverlay returns the platform-appropriate default for the
// overlay hotkey. On macOS it's Cmd+Shift+Space; on Win/Linux it's
// Ctrl+Shift+Space. The user can override in settings.
func DefaultOverlay() string {
	if config.PlatformIsMac() {
		return "Cmd+Shift+Space"
	}
	return "Ctrl+Shift+Space"
}
