//go:build linux

package hotkey

import "testing"

func TestLinuxStart_NilHandler(t *testing.T) {
	m := New("Cmd+K")
	if err := m.Start(nil); err == nil {
		t.Fatal("Start(nil) should error")
	}
}

// TestLinuxStart_ErrorsUnsupported: O4 — Start on Linux must return
// errLinuxUnsupported (not nil) so callers fall back honestly instead
// of silently believing a global hotkey is registered. Started() must
// stay false (no real hotkey).
func TestLinuxStart_ErrorsUnsupported(t *testing.T) {
	m := New("Cmd+K")
	err := m.Start(func() {})
	if err == nil {
		t.Fatal("Start on Linux must return errLinuxUnsupported, not nil")
	}
	if err != errLinuxUnsupported {
		t.Fatalf("Start error = %v, want errLinuxUnsupported", err)
	}
	if m.Started() {
		t.Fatal("Started must be false — no real hotkey was registered")
	}
}

func TestLinuxStartHold_NilCallbacks(t *testing.T) {
	m := New("Cmd+K")
	if err := m.StartHold(nil, func() {}, 100); err == nil {
		t.Fatal("StartHold(nil, ...) should error")
	}
	if err := m.StartHold(func() {}, nil, 100); err == nil {
		t.Fatal("StartHold(..., nil, ...) should error")
	}
}

func TestLinuxStartTap_NilCallback(t *testing.T) {
	m := New("Cmd+K")
	if err := m.StartTap(nil, 2, 300); err == nil {
		t.Fatal("StartTap(nil) should error")
	}
}

func TestLinuxPressCount(t *testing.T) {
	m := New("Cmd+K")
	if m.PressCount() != 0 {
		t.Errorf("PressCount on Linux = %d, want 0", m.PressCount())
	}
}

func TestLinuxStopIdempotent(t *testing.T) {
	m := New("Cmd+K")
	m.Stop()
	m.Stop()
}
