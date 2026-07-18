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

// -----------------------------------------------------------------------------
// VectorClock.Merge + VectorClock.Equal — CRDT primitives
//
// VectorClock is the foundation of the sync engine's causal ordering.
// Merge computes the element-wise maximum of two clocks; Equal
// tests structural identity. A bug in either would let the sync
// engine either lose history (Merge taking minimum instead of
// max) or report false-positive concurrency (Equal returning
// true for different clocks).
//
// Before: Merge 0% and Equal 0% coverage. The Store-level merge
// tests cover the higher-level "did we record a conflict?"
// question but not these foundational CRDT operations.
// -----------------------------------------------------------------------------

func TestVectorClock_Merge_TakesElementwiseMax(t *testing.T) {
	vc := VectorClock{"a": 1, "b": 2, "c": 3}
	other := VectorClock{"a": 5, "b": 1, "d": 7} // "d" not in vc
	vc.Merge(other)

	want := VectorClock{"a": 5, "b": 2, "c": 3, "d": 7}
	if !vc.Equal(want) {
		t.Errorf("Merge result = %v, want %v (element-wise max)", vc, want)
	}
}

func TestVectorClock_Merge_DoesNotMutateOther(t *testing.T) {
	vc := VectorClock{"a": 1, "b": 2}
	other := VectorClock{"a": 5, "b": 1}
	// Snapshot of other before Merge.
	before := VectorClock{"a": 5, "b": 1}

	vc.Merge(other)

	// Merge modifies vc in place; it MUST NOT touch other.
	// A regression that wrote into both maps would silently
	// corrupt the caller-side state.
	if !other.Equal(before) {
		t.Errorf("Merge mutated 'other' — caller-side state should be untouched. before=%v after=%v", before, other)
	}
}

func TestVectorClock_Merge_EmptyOther_LeavesVcUnchanged(t *testing.T) {
	vc := VectorClock{"a": 1, "b": 2}
	vc.Merge(VectorClock{}) // empty merge

	want := VectorClock{"a": 1, "b": 2}
	if !vc.Equal(want) {
		t.Errorf("Merge with empty other modified vc = %v, want %v", vc, want)
	}
}

func TestVectorClock_Equal_IdenticalClocks(t *testing.T) {
	a := VectorClock{"a": 1, "b": 2}
	b := VectorClock{"a": 1, "b": 2}
	if !a.Equal(b) {
		t.Errorf("Equal(identical clocks) = false, want true")
	}
}

func TestVectorClock_Equal_DifferentLengthsReturnsFalse(t *testing.T) {
	a := VectorClock{"a": 1, "b": 2}
	b := VectorClock{"a": 1} // missing b
	if a.Equal(b) {
		t.Errorf("Equal(clocks with different lengths) = true, want false")
	}
}

func TestVectorClock_Equal_DifferentValuesReturnsFalse(t *testing.T) {
	a := VectorClock{"a": 1, "b": 2}
	b := VectorClock{"a": 1, "b": 3} // same length, b differs
	if a.Equal(b) {
		t.Errorf("Equal(clocks with different values) = true, want false")
	}
}

// -----------------------------------------------------------------------------
// Store.VectorSnapshot + Store.Hash — sync state observability
//
// VectorSnapshot returns a copy of the current vector clock state
// (element-wise max over all entries' Version maps). Used for sync
// state verification.
//
// Hash returns a SHA-256 over all entries, with keys sorted
// alphabetically to guarantee deterministic ordering. Used for
// integrity checking — a regression that lost the sort.Strings
// call would produce different hashes on Go map iteration order
// (which is randomized), breaking the integrity contract.
// -----------------------------------------------------------------------------

func TestStore_VectorSnapshot_EmptyStoreReturnsEmptyClock(t *testing.T) {
	s := NewStore()
	vc := s.VectorSnapshot()
	if len(vc) != 0 {
		t.Errorf("VectorSnapshot on empty store = %v, want empty clock", vc)
	}
}

func TestStore_VectorSnapshot_ReturnsIndependentCopy(t *testing.T) {
	s := NewStore()
	s.Put("device-a", "k1", []byte("v"))
	// Snapshot must be a copy — mutating the returned clock MUST
	// NOT affect the store's internal state.
	snap := s.VectorSnapshot()
	snap["b"] = 99 // mutate the snapshot
	vc2 := s.VectorSnapshot()
	if _, ok := vc2["b"]; ok {
		t.Errorf("mutating VectorSnapshot result leaked into store: %v", vc2)
	}
}

func TestStore_VectorSnapshot_TakesElementwiseMax(t *testing.T) {
	s := NewStore()
	// Two entries with overlapping devices; max should win per device.
	s.Put("device-a", "k1", []byte("a"))
	s.Put("device-b", "k2", []byte("b"))
	vc := s.VectorSnapshot()
	want := VectorClock{"device-a": 1, "device-b": 1}
	if !vc.Equal(want) {
		t.Errorf("VectorSnapshot = %v, want %v (element-wise max across entries)", vc, want)
	}
}

func TestStore_Hash_DeterministicForSameContent(t *testing.T) {
	s := NewStore()
	s.Put("device-a", "k1", []byte("v"))
	h1 := s.Hash()
	h2 := s.Hash()
	if h1 != h2 {
		t.Errorf("Hash not deterministic: h1=%s h2=%s", h1, h2)
	}
}

func TestStore_Hash_ChangesWhenContentChanges(t *testing.T) {
	s := NewStore()
	s.Put("device-a", "k1", []byte("v1"))
	h1 := s.Hash()
	s.Put("device-a", "k1", []byte("v2"))
	h2 := s.Hash()
	if h1 == h2 {
		t.Errorf("Hash unchanged after content change: h1=%s h2=%s", h1, h2)
	}
}

func TestStore_Hash_EmptyStoreProducesValidHex(t *testing.T) {
	s := NewStore()
	h := s.Hash()
	// SHA-256 produces 64 hex chars; the empty store should produce
	// the SHA-256 of the empty input (well-defined constant).
	if len(h) != 64 {
		t.Errorf("Hash length = %d, want 64 hex chars (SHA-256 output)", len(h))
	}
	// Hash should be reproducible on empty.
	if h2 := s.Hash(); h != h2 {
		t.Errorf("Hash of empty store not deterministic: h1=%s h2=%s", h, h2)
	}
}
