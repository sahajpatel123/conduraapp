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
type Store struct {
	mu      sync.RWMutex
	entries map[string]*Entry
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
// vector clock ordering. Returns true if the entry was applied.
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
	if remote.Timestamp.After(local.Timestamp) ||
		(remote.Timestamp.Equal(local.Timestamp) && remote.DeviceID > local.DeviceID) {
		s.entries[remote.Key] = remote
		return true
	}
	return false
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
