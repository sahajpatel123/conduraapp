//go:build !linux

package tray

import "testing"

// TestEventConstants verifies the Event iota is contiguous starting
// at EventNone. The systray wrapper relies on this for the Run
// loop's switch.
func TestEventConstants(t *testing.T) {
	want := []Event{EventNone, EventShow, EventHide, EventToggleHalt, EventQuit}
	for i, w := range want {
		if Event(i) != w {
			t.Fatalf("Event(%d) = %d, want %d", i, Event(i), w)
		}
	}
}

// TestNew_BuildsMenu verifies that New returns a non-nil Menu with
// a non-nil Events channel. The actual systray native code is not
// invoked here (it requires a display server).
func TestNew_BuildsMenu(t *testing.T) {
	m := New("Synaptic", "Free on-device AI agent")
	if m == nil {
		t.Fatal("New returned nil")
	}
	if m.Events() == nil {
		t.Fatal("Events() returned nil channel")
	}
	if m.title != "Synaptic" {
		t.Fatalf("title = %q", m.title)
	}
}

// TestSetHalted_LabelSwitch verifies that SetHalted flips the
// internal halted flag and the menu item title after onReady. We
// can only assert the flag here (onReady needs a display).
func TestSetHalted_Flag(t *testing.T) {
	m := New("Synaptic", "t")
	m.SetHalted(true)
	if !m.halted.Load() {
		t.Fatal("flag should be true after SetHalted(true)")
	}
	m.SetHalted(false)
	if m.halted.Load() {
		t.Fatal("flag should be false after SetHalted(false)")
	}
}

// TestSetSpendUSD verifies the cents conversion is exact.
func TestSetSpendUSD(t *testing.T) {
	m := New("Synaptic", "t")
	m.SetSpendUSD(12.34)
	if got := m.spend.Load(); got != 1234 {
		t.Fatalf("cents = %d, want 1234", got)
	}
	m.SetSpendUSD(0)
	if got := m.spend.Load(); got != 0 {
		t.Fatalf("cents after zero = %d, want 0", got)
	}
}

// TestSetSpendUSD_Fractional exercises sub-cent rounding.
func TestSetSpendUSD_Fractional(t *testing.T) {
	m := New("Synaptic", "t")
	m.SetSpendUSD(0.05)
	if got := m.spend.Load(); got != 5 {
		t.Fatalf("cents for $0.05 = %d, want 5", got)
	}
	m.SetSpendUSD(0.01)
	if got := m.spend.Load(); got != 1 {
		t.Fatalf("cents for $0.01 = %d, want 1", got)
	}
}

// TestSetTooltip_StoresValue confirms SetTooltip updates the
// struct field (the actual systray tooltip setter requires a
// display and is verified manually).
func TestSetTooltip_StoresValue(t *testing.T) {
	m := New("Synaptic", "old")
	m.SetTooltip("new")
	if m.tooltip != "new" {
		t.Fatalf("tooltip = %q, want \"new\"", m.tooltip)
	}
}

// TestSetHalted_NoMenuBeforeOnReady verifies that SetHalted is
// safe to call before onReady has run. It updates the flag but
// does not panic on the nil m.mHalt.
func TestSetHalted_NoMenuBeforeOnReady(t *testing.T) {
	m := New("Synaptic", "t")
	m.SetHalted(true)
	if !m.halted.Load() {
		t.Fatal("flag should be true even without onReady")
	}
	m.SetHalted(false)
	if m.halted.Load() {
		t.Fatal("flag should be false even without onReady")
	}
}
