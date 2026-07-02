package conductor

import (
	"context"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/halt"
	"github.com/sahajpatel123/conduraapp/internal/storage"
)

// setupFlag mirrors halt/flag_test.go's setupFlag so the kill-switch
// test can exercise a real Flag backed by SQLite. Kept local to this
// file (not exported) because the kill-switch test should not
// depend on internal test helpers of the halt package.
func setupFlagLocal(t *testing.T) *halt.Flag {
	t.Helper()
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{
		Path: filepath.Join(dir, "test.db"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return halt.New(db.SQL())
}

// TestNewKillSwitch_RequiresHaltFlag asserts the constructor
// rejects a nil halt.Flag. This is the safety invariant: a kill
// switch with nowhere to fire is a silent failure waiting to
// happen, and CLAUDE.md §2.1 invariant #4 says the user must
// always be able to stop the agent.
func TestNewKillSwitch_RequiresHaltFlag(t *testing.T) {
	_, err := NewKillSwitch("Cmd+Shift+Escape", nil, nil)
	if err == nil {
		t.Fatal("expected error for nil halt flag")
	}
}

// TestNewKillSwitch_OK asserts the constructor succeeds with a
// non-nil halt flag and stores the spec/manager.
func TestNewKillSwitch_OK(t *testing.T) {
	f := setupFlagLocal(t)
	ks, err := NewKillSwitch("Cmd+Shift+Escape", f, nil)
	if err != nil {
		t.Fatalf("NewKillSwitch: %v", err)
	}
	if ks == nil {
		t.Fatal("nil conductor")
	}
	if ks.spec != "Cmd+Shift+Escape" {
		t.Errorf("spec = %q, want Cmd+Shift+Escape", ks.spec)
	}
	if ks.mgr == nil {
		t.Error("manager not constructed")
	}
	if ks.Started() {
		t.Error("Started() should be false before Start()")
	}
}

// TestKillSwitch_FireHaltsWithHardHotkeyReason asserts that the
// internal fire() method calls haltFlag.Halt with ReasonHardHotkey.
// This is the core wiring test — if fire() ever drifts from this
// contract, the audit log will not show "hard_hotkey" as the
// reason, and the kill-switch UI overlay will not display the
// correct cause. We exercise fire() directly (it's unexported
// because it's bound to the hotkey callback); the hotkey.Manager
// is never started in this test because we don't have a real
// keyboard.
func TestKillSwitch_FireHaltsWithHardHotkeyReason(t *testing.T) {
	f := setupFlagLocal(t)
	ctx := context.Background()
	if err := f.Refresh(ctx); err != nil {
		t.Fatal(err)
	}

	ks, err := NewKillSwitch("Cmd+Shift+Escape", f, nil)
	if err != nil {
		t.Fatalf("NewKillSwitch: %v", err)
	}

	// Simulate a hotkey press.
	ks.fire()

	if !f.IsHalted() {
		t.Fatal("halt flag should be set after fire()")
	}
	s := f.Halted()
	if s.Reason != ReasonHardHotkey {
		t.Errorf("reason = %q, want %q", s.Reason, ReasonHardHotkey)
	}
	if s.Since.IsZero() {
		t.Error("since should be set")
	}
}

// TestKillSwitch_OnHaltCallback asserts that the optional onHalt
// callback fires exactly once after fire(). The callback is the
// UI surface for the kill-switch overlay (lib/condura/KillSwitchOverlay.svelte).
func TestKillSwitch_OnHaltCallback(t *testing.T) {
	f := setupFlagLocal(t)
	ctx := context.Background()
	_ = f.Refresh(ctx)

	var calls atomic.Int64
	ks, err := NewKillSwitch("Cmd+Shift+Escape", f, func() {
		calls.Add(1)
	})
	if err != nil {
		t.Fatal(err)
	}

	ks.fire()
	ks.fire() // second fire should still increment (callback fires every press)
	ks.fire()

	if got := calls.Load(); got != 3 {
		t.Errorf("onHalt fired %d times, want 3", got)
	}
}

// TestKillSwitch_ReasonHardHotkeyConstant asserts the public
// constant matches what the audit log expects. This is the
// pinning test: if anyone changes ReasonHardHotkey, this test
// fails and they must update both the source and the audit-log
// readers.
func TestKillSwitch_ReasonHardHotkeyConstant(t *testing.T) {
	if ReasonHardHotkey != "hard_hotkey" {
		t.Errorf("ReasonHardHotkey = %q, want hard_hotkey", ReasonHardHotkey)
	}
}

// TestKillSwitch_StopIdempotent asserts Stop is safe before Start
// and after Stop (no panic, no leaked goroutines beyond the
// goroutine started by hotkey.Manager.Start which we never
// invoke here).
func TestKillSwitch_StopIdempotent(t *testing.T) {
	f := setupFlagLocal(t)
	ks, err := NewKillSwitch("Cmd+Shift+Escape", f, nil)
	if err != nil {
		t.Fatal(err)
	}
	ks.Stop() // no-op (not started)
	ks.Stop() // still no-op
	if ks.Started() {
		t.Error("Started() should be false after Stop()")
	}
}
