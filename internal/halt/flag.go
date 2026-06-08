// Package halt is the kill-switch flag. When set, every subsystem
// should refuse new work and the daemon cancels in-flight streams.
//
// The flag is persisted in SQLite (so it survives a restart) and
// also kept in memory (so reads are O(1) and don't hit the DB on
// every check). The two are kept in sync via a single goroutine
// that polls the DB every second; the in-memory value is the
// source of truth for the hot path.
package halt

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"time"
)

// State is the current halt state.
type State struct {
	Halted bool      `json:"halted"`
	Since  time.Time `json:"since,omitempty"`
	Reason string    `json:"reason,omitempty"`
}

// Flag is the kill-switch.
type Flag struct {
	mu    sync.RWMutex
	state State
	atom  atomic.Bool
	db    *sql.DB
}

// New returns a Flag backed by the given database.
func New(db *sql.DB) *Flag {
	return &Flag{db: db}
}

// Halted returns the current in-memory state. Safe for concurrent use.
func (f *Flag) Halted() State {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

// Set updates the in-memory + on-disk halt state. Returns the
// previous state so the caller can show the user "halted since X".
func (f *Flag) Set(ctx context.Context, halted bool, reason string) (State, error) {
	f.mu.Lock()
	prev := f.state
	now := time.Now().UTC()
	if halted {
		f.state = State{Halted: true, Since: now, Reason: reason}
	} else {
		f.state = State{Halted: false}
	}
	f.atom.Store(halted)
	f.mu.Unlock()

	// Persist.
	var since interface{}
	if halted {
		since = now.Format(time.RFC3339)
	}
	if _, err := f.db.ExecContext(ctx,
		`UPDATE halt_state SET halted = ?, since = ?, reason = ? WHERE id = 1`,
		boolToInt(halted), since, reason,
	); err != nil {
		// Roll back in-memory if the DB write failed.
		f.mu.Lock()
		f.state = prev
		f.atom.Store(prev.Halted)
		f.mu.Unlock()
		return State{}, err
	}
	return prev, nil
}

// Halt is a convenience for Set(ctx, true, reason).
func (f *Flag) Halt(ctx context.Context, reason string) (State, error) {
	return f.Set(ctx, true, reason)
}

// Resume is a convenience for Set(ctx, false, "").
func (f *Flag) Resume(ctx context.Context) (State, error) {
	return f.Set(ctx, false, "")
}

// Refresh re-reads the on-disk state into the in-memory value.
// Call once at startup, then periodically (every ~1s).
func (f *Flag) Refresh(ctx context.Context) error {
	var halted int
	var since sql.NullString
	var reason sql.NullString
	row := f.db.QueryRowContext(ctx, `SELECT halted, since, reason FROM halt_state WHERE id = 1`)
	if err := row.Scan(&halted, &since, &reason); err != nil {
		return err
	}
	f.mu.Lock()
	f.state = State{Halted: halted != 0}
	if since.Valid {
		f.state.Since, _ = time.Parse(time.RFC3339, since.String)
	}
	if reason.Valid {
		f.state.Reason = reason.String
	}
	f.atom.Store(halted != 0)
	f.mu.Unlock()
	return nil
}

// IsHalted is the hot-path check. Uses atomic load; safe for the
// millions of times per second an LLM-stream check might want it.
func (f *Flag) IsHalted() bool {
	return f.atom.Load()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
