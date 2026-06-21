package sync

import (
	"testing"
	"time"
)

// Phase 16, Rec 4: concurrent writes that fall through to LWW
// are recorded in the conflict log.
func TestStore_Merge_RecordsConflict(t *testing.T) {
	s := NewStore()

	// Two concurrent writes to the same key from different devices.
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	remote := &Entry{
		Key:       "skills/foo",
		Value:     []byte("from-device-b"),
		DeviceID:  "device-b",
		Version:   VectorClock{"device-b": 1},
		Timestamp: t0.Add(1 * time.Second), // later
	}
	local := &Entry{
		Key:       "skills/foo",
		Value:     []byte("from-device-a"),
		DeviceID:  "device-a",
		Version:   VectorClock{"device-a": 1},
		Timestamp: t0, // earlier
	}
	s.entries[remote.Key] = local

	applied := s.Merge(remote)
	if !applied {
		t.Fatal("remote (later timestamp) should win")
	}

	conflicts := s.Conflicts()
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	c := conflicts[0]
	if c.Key != "skills/foo" {
		t.Errorf("Key: got %q", c.Key)
	}
	if c.WinnerDeviceID != "device-b" {
		t.Errorf("WinnerDeviceID: got %q, want device-b", c.WinnerDeviceID)
	}
	if c.LoserDeviceID != "device-a" {
		t.Errorf("LoserDeviceID: got %q, want device-a", c.LoserDeviceID)
	}
}

func TestStore_Merge_NoConflictOnCausalOrder(t *testing.T) {
	s := NewStore()

	// local happens-before remote: local has seen device-b's edit
	// (so its clock is at device-b:1), and remote's new edit
	// increments device-b to 2. Causally-ordered → no conflict.
	local := &Entry{
		Key:       "k",
		Value:     []byte("a"),
		DeviceID:  "device-a",
		Version:   VectorClock{"device-a": 1, "device-b": 1},
		Timestamp: time.Now().Add(-time.Hour),
	}
	remote := &Entry{
		Key:       "k",
		Value:     []byte("b"),
		DeviceID:  "device-b",
		Version:   VectorClock{"device-a": 1, "device-b": 2},
		Timestamp: time.Now(),
	}
	s.entries[local.Key] = local

	s.Merge(remote)
	if got := len(s.Conflicts()); got != 0 {
		t.Errorf("causal merge should NOT log a conflict, got %d", got)
	}
}

func TestStore_ConflictsClear(t *testing.T) {
	s := NewStore()
	s.conflicts = append(s.conflicts, Conflict{Key: "k"})
	if len(s.Conflicts()) != 1 {
		t.Fatal("setup")
	}
	s.ConflictsClear()
	if got := len(s.Conflicts()); got != 0 {
		t.Errorf("after clear: got %d conflicts", got)
	}
}
