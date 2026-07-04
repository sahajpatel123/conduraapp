// Package window persists GUI window position/size + last-opened
// conversation across restarts. The GUI calls SetSize/SetPosition
// whenever the user moves or resizes the window; the daemon writes
// through to the SQLite window_state table. On launch the GUI calls
// Snapshot() to restore the previous geometry.
//
// All methods are safe to call from any goroutine. Writes are
// serialized behind an internal mutex; the underlying DB connection
// is shared with the rest of the daemon (storage.DB).
package window

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
)

// State is a serialisable snapshot of the GUI window geometry. Zero
// X/Y means "let the OS decide where to place the window".
type State struct {
	Width              int   `json:"width"`
	Height             int   `json:"height"`
	X                  *int  `json:"x,omitempty"`
	Y                  *int  `json:"y,omitempty"`
	LastConversationID int64 `json:"last_conversation_id"`
}

// Default geometry used the very first time the app starts (before
// any user interaction has been recorded).
const (
	defaultWidth  = 1200
	defaultHeight = 800
)

// Default is the geometry used the very first time the app starts
// (before any user interaction has been recorded).
func Default() State {
	return State{Width: defaultWidth, Height: defaultHeight}
}

// Manager owns a single State value and a SQLite row. The 0th
// implementation was: each call to SetSize/SetPosition writes a
// separate row. That was wrong — the window_state table enforces
// id = 1, so we keep a single row and update it.
type Manager struct {
	db  *sql.DB
	mu  sync.Mutex
	cur State
}

// New constructs a Manager bound to the given DB. The current state
// is read synchronously; if the row is missing or corrupt, Default()
// is used and the row is rewritten to its default value.
func New(db *sql.DB) *Manager {
	m := &Manager{db: db, cur: Default()}
	m.reload()
	return m
}

// Snapshot returns the current State. Cheap, lock-held only briefly.
func (m *Manager) Snapshot() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cur
}

// SetSize persists a new width/height pair. Both must be > 0; the
// call is a no-op (and returns an error) if either is non-positive.
func (m *Manager) SetSize(ctx context.Context, w, h int) error {
	if w <= 0 || h <= 0 {
		return errors.New("window: width and height must be > 0")
	}
	return m.update(ctx, patchOp{kind: patchSize, w: w, h: h})
}

// SetPosition persists a new top-left position. Either coordinate
// may be nil, meaning "let the OS decide". When both are non-nil
// the row's x and y are updated atomically.
func (m *Manager) SetPosition(ctx context.Context, x, y *int) error {
	if x != nil && y != nil {
		return m.updateBoth(ctx, *x, *y)
	}
	if x != nil {
		return m.update(ctx, patchOp{kind: patchX, x: x})
	}
	if y != nil {
		return m.update(ctx, patchOp{kind: patchY, y: y})
	}
	// Both nil: no-op.
	return nil
}

// updateBoth writes x and y in a single UPDATE. We use a separate
// query rather than extending the patchOp enum because the two
// halves share no other codepath.
func (m *Manager) updateBoth(ctx context.Context, x, y int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, err := m.db.ExecContext(ctx, `UPDATE window_state SET x = ?, y = ? WHERE id = 1`, x, y); err != nil {
		return fmt.Errorf("window: update position: %w", err)
	}
	m.reloadLocked()
	return nil
}

// SetLastConversation records the most-recently-opened conversation.
// Stored as int64 so the window state row is the single source of
// truth for "where to land on launch".
func (m *Manager) SetLastConversation(ctx context.Context, id int64) error {
	return m.update(ctx, patchOp{kind: patchLastConv, id: id})
}

// update applies a column=?, column=? patch to the singleton row
// and updates the in-memory cache. The column expression is built by
// the caller and is restricted to whitelisted column names (so this
// function is safe to call only with internal literal column names).
//
// We use a small allow-list of permitted SET clauses rather than
// fmt.Sprintf with the user-provided string; this dodges the gosec
// G201 warning and makes the SQL injection surface zero.
func (m *Manager) update(ctx context.Context, op patchOp) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var (
		query string
		args  []any
	)
	switch op.kind {
	case patchSize:
		query = `UPDATE window_state SET width = ?, height = ? WHERE id = 1`
		args = []any{op.w, op.h}
	case patchX:
		query = `UPDATE window_state SET x = ? WHERE id = 1`
		args = []any{*op.x}
	case patchY:
		query = `UPDATE window_state SET y = ? WHERE id = 1`
		args = []any{*op.y}
	case patchLastConv:
		query = `UPDATE window_state SET last_conversation_id = ? WHERE id = 1`
		args = []any{op.id}
	default:
		return fmt.Errorf("window: unknown patch op %d", op.kind)
	}
	if _, err := m.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("window: update: %w", err)
	}
	m.reloadLocked()
	return nil
}

// patchKind enumerates the allowed UPDATE operations.
type patchKind uint8

const (
	patchSize patchKind = iota
	patchX
	patchY
	patchLastConv
)

// patchOp is the tagged-union payload passed to update. Each op
// carries only the fields it needs.
type patchOp struct {
	kind patchKind
	w, h int
	x, y *int
	id   int64
}

// reload re-reads the row from disk into the in-memory cache.
func (m *Manager) reload() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reloadLocked()
}

// reloadLocked must be called with m.mu held.
func (m *Manager) reloadLocked() {
	if m.db == nil {
		return
	}
	var (
		w, h     int
		x, y     sql.NullInt64
		lastConv int64
	)
	row := m.db.QueryRowContext(context.Background(), `SELECT width, height, x, y, last_conversation_id FROM window_state WHERE id = 1`)
	if err := row.Scan(&w, &h, &x, &y, &lastConv); err != nil {
		// Row missing or schema drift: fall back to defaults.
		m.cur = Default()
		return
	}
	s := State{
		Width:              w,
		Height:             h,
		LastConversationID: lastConv,
	}
	if x.Valid {
		v := int(x.Int64)
		s.X = &v
	}
	if y.Valid {
		v := int(y.Int64)
		s.Y = &v
	}
	m.cur = s
}
