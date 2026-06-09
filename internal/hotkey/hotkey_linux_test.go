//go:build linux

package hotkey

import "testing"

func TestLinuxStart_NilHandler(t *testing.T) {
	m := New("Cmd+K")
	if err := m.Start(nil); err == nil {
		t.Fatal("Start(nil) should error")
	}
}

func TestLinuxStartOK(t *testing.T) {
	m := New("Cmd+K")
	if err := m.Start(func() {}); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !m.Started() {
		t.Fatal("Started should be true after Start")
	}
	m.Stop()
	if m.Started() {
		t.Fatal("Started should be false after Stop")
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
