// Package sync implements device-to-device encrypted sync (Phase 12).
//
// CRDT model (Phase 16, Rec 4: documented):
//   - Each key has a VectorClock. Causally-ordered writes merge
//     automatically: remote.HappensBefore(local) → keep local;
//     local.HappensBefore(remote) → apply remote.
//   - Concurrent writes (neither clock happens-before the other)
//     fall back to Last-Writer-Wins: the entry with the higher
//     Timestamp wins; ties break by lexicographic DeviceID. This is
//     NOT a true CRDT — it's LWW with vector-clock pre-check.
//   - Conflicts are logged to the per-store Conflicts slice so
//     the user / audit system can review them.
//
// Why LWW (not OR-Set) for v0.1.0:
//   - The common case is one active device per user. LWW is
//     simple, deterministic, and gives the user a clear mental
//     model ("the most recent edit wins").
//   - OR-Set would require tombstones for deletes, complicating
//     the on-disk format. Out of scope for v0.1.0; tracked for
//     v0.2.0+ as a CRDT upgrade.
//
// Trade-off acknowledged: LWW can drop concurrent edits silently.
// The Conflict log + the audit log are the user's primary
// visibility into dropped edits.
package sync

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Entry is a single CRDT entry representing a versioned state change.
// The vector clock ensures causal ordering; on conflict, the entry
// with the higher timestamp wins (last-writer-wins per key).
type Entry struct {
	Key       string      `json:"key"`
	Value     []byte      `json:"value"`
	Version   VectorClock `json:"version"`
	DeviceID  string      `json:"device_id"`
	Timestamp time.Time   `json:"timestamp"`
	Deleted   bool        `json:"deleted,omitempty"`
}

// VectorClock is a per-device logical clock. Each device increments
// its own counter on every write; the clock captures the causal
// history observed by that device.
type VectorClock map[string]int64

// Increment bumps the counter for the given device.
func (vc VectorClock) Increment(deviceID string) {
	vc[deviceID]++
}

// Merge takes the element-wise maximum of two vector clocks.
func (vc VectorClock) Merge(other VectorClock) {
	for dev, ts := range other {
		if ts > vc[dev] {
			vc[dev] = ts
		}
	}
}

// HappensBefore returns true if vc causally precedes other.
func (vc VectorClock) HappensBefore(other VectorClock) bool {
	less := false
	// Check all devices in vc.
	for dev, ts := range vc {
		if ts > other[dev] {
			return false
		}
		if ts < other[dev] {
			less = true
		}
	}
	// Check devices only in other (vc doesn't know about them yet).
	for dev, ts := range other {
		if _, ok := vc[dev]; !ok && ts > 0 {
			less = true
		}
	}
	return less
}

// Equal returns true if both clocks have the same values.
func (vc VectorClock) Equal(other VectorClock) bool {
	if len(vc) != len(other) {
		return false
	}
	for dev, ts := range vc {
		if other[dev] != ts {
			return false
		}
	}
	return true
}

// Store is a thread-safe CRDT store of entries. It supports merge
// (for sync), get, put, and delete. Conflict resolution uses
// last-writer-wins with vector clock ordering.
// Conflict records a single LWW tie-break. Phase 16, Rec 4: every
// concurrent merge that falls back to LWW is recorded so the
// user can review dropped edits in the audit log.
type Conflict struct {
	// Key is the entry key the conflict was about.
	Key string
	// WinnerDeviceID is the device whose edit won the LWW.
	WinnerDeviceID string
	// LoserDeviceID is the device whose edit was overwritten.
	LoserDeviceID string
	// WinnerTimestamp is the timestamp of the winning edit.
	WinnerTimestamp time.Time
	// LoserTimestamp is the timestamp of the dropped edit.
	LoserTimestamp time.Time
	// ResolvedAt is when the conflict was recorded (local clock).
	ResolvedAt time.Time
}

type Store struct {
	mu      sync.RWMutex
	entries map[string]*Entry
	// conflicts is an append-only log of LWW tie-breaks. Reads
	// return a copy so callers can iterate without holding the
	// lock. Cleared on ConflictsClear (UI provides a button).
	conflicts []Conflict
}

// NewStore returns an empty CRDT store.
func NewStore() *Store {
	return &Store{entries: make(map[string]*Entry)}
}

// Put inserts or updates an entry. The device's vector clock is
// incremented before writing.
func (s *Store) Put(deviceID, key string, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing := s.entries[key]
	vc := VectorClock{}
	if existing != nil {
		vc = existing.Version
	}
	vc.Increment(deviceID)

	entry := &Entry{
		Key:       key,
		Value:     value,
		Version:   vc,
		DeviceID:  deviceID,
		Timestamp: time.Now().UTC(),
	}
	s.entries[key] = entry
}

// Delete marks an entry as deleted (tombstone). The entry is kept
// for causal broadcast; a full GC pass can remove old tombstones.
func (s *Store) Delete(deviceID, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing := s.entries[key]
	vc := VectorClock{}
	if existing != nil {
		vc = existing.Version
	}
	vc.Increment(deviceID)

	s.entries[key] = &Entry{
		Key:       key,
		Version:   vc,
		DeviceID:  deviceID,
		Timestamp: time.Now().UTC(),
		Deleted:   true,
	}
}

// Get retrieves an entry by key. Returns nil if not found or deleted.
func (s *Store) Get(key string) *Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e := s.entries[key]
	if e == nil || e.Deleted {
		return nil
	}
	return e
}

// Merge incorporates a remote entry using last-writer-wins with
// vector-clock pre-check (Phase 16, Rec 4: documented policy).
//
// Returns true if remote was applied (newer or won the tie-break),
// false if local is kept (newer or won the tie-break).
//
// On a tie-break, the conflict is appended to Store.conflicts so
// the user / audit system can review dropped edits.
func (s *Store) Merge(remote *Entry) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	local := s.entries[remote.Key]
	if local == nil {
		s.entries[remote.Key] = remote
		return true
	}

	// If remote version happens-after local, apply.
	if remote.Version.HappensBefore(local.Version) {
		return false // local is newer
	}
	if local.Version.HappensBefore(remote.Version) {
		s.entries[remote.Key] = remote
		return true
	}
	// Concurrent writes — tie-break by timestamp, then device ID.
	remoteWins := remote.Timestamp.After(local.Timestamp) ||
		(remote.Timestamp.Equal(local.Timestamp) && remote.DeviceID > local.DeviceID)

	// Phase 16, Rec 4: log the tie-break so dropped edits are
	// visible. Either winner is recorded.
	if remoteWins {
		s.recordConflict(local, remote)
		s.entries[remote.Key] = remote
	} else {
		s.recordConflict(remote, local)
	}
	return remoteWins
}

// recordConflict appends a Conflict entry to the in-memory log.
// Caller must hold s.mu (write lock).
func (s *Store) recordConflict(loser, winner *Entry) {
	s.conflicts = append(s.conflicts, Conflict{
		Key:             winner.Key,
		WinnerDeviceID:  winner.DeviceID,
		LoserDeviceID:   loser.DeviceID,
		WinnerTimestamp: winner.Timestamp,
		LoserTimestamp:  loser.Timestamp,
		ResolvedAt:      time.Now().UTC(),
	})
}

// Conflicts returns a snapshot of all logged conflicts. The slice
// is in append order (oldest first).
func (s *Store) Conflicts() []Conflict {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Conflict, len(s.conflicts))
	copy(out, s.conflicts)
	return out
}

// ConflictsClear empties the conflict log. The UI exposes a
// "Clear" button on the Settings → Sync page.
func (s *Store) ConflictsClear() {
	s.mu.Lock()
	s.conflicts = nil
	s.mu.Unlock()
}

// Entries returns all non-deleted entries sorted by key.
func (s *Store) Entries() []*Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Entry, 0, len(s.entries))
	for _, e := range s.entries {
		if !e.Deleted {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})
	return out
}

// VectorSnapshot returns a copy of the current vector clock state.
func (s *Store) VectorSnapshot() VectorClock {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vc := VectorClock{}
	for _, e := range s.entries {
		for dev, ts := range e.Version {
			if ts > vc[dev] {
				vc[dev] = ts
			}
		}
	}
	return vc
}

// Hash returns a SHA-256 hash of all entries for integrity checking.
func (s *Store) Hash() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h := sha256.New()
	keys := make([]string, 0, len(s.entries))
	for k := range s.entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b, _ := json.Marshal(s.entries[k])
		h.Write(b)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
