// Package backup also implements the honest-scope Rollback
// path. "Rollback" here means: undo the Synaptic-owned state to
// a checkpoint, AND be honest about what cannot be undone
// (irreversible OS actions).
package backup

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Checkpoint is a named point in time. Used as the "rollback to
// here" target.
type Checkpoint struct {
	ID        int64
	CreatedAt time.Time
	Reason    string
}

// Rollback reverts Synaptic-owned rows inserted after a checkpoint.
// It never touches the HMAC-chained audit log.
type Rollback struct {
	mainDB   *sql.DB
	memoryDB *sql.DB
	skillsDB *sql.DB
	window   time.Duration
	// opened tracks DBs we opened (via OpenRollbackDB) so Close
	// can release them. DBs passed directly (like mainDB from the
	// caller) are NOT tracked — the caller owns their lifecycle.
	opened []*sql.DB
}

// NewRollback returns a Rollback helper for the main database only.
func NewRollback(db *sql.DB) *Rollback {
	return &Rollback{mainDB: db, window: 1 * time.Hour}
}

// NewRollbackMulti returns a Rollback that can revert conversations
// (main DB), memory episodes/facts (memory DB), and skill rows
// (skills DB). Nil DB pointers are skipped.
func NewRollbackMulti(mainDB, memoryDB, skillsDB *sql.DB) *Rollback {
	return &Rollback{mainDB: mainDB, memoryDB: memoryDB, skillsDB: skillsDB, window: 1 * time.Hour}
}

// TrackOpened records DBs opened via OpenRollbackDB so Close can
// release them. DBs passed to NewRollback/NewRollbackMulti are owned
// by the caller and NOT tracked.
func (r *Rollback) TrackOpened(dbs ...*sql.DB) {
	for _, db := range dbs {
		if db != nil {
			r.opened = append(r.opened, db)
		}
	}
}

// Close releases any DB connections opened via OpenRollbackDB.
// It does NOT close DBs passed to the constructor (caller-owned).
func (r *Rollback) Close() error {
	var errs []error
	for _, db := range r.opened {
		if err := db.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	r.opened = nil
	if len(errs) > 0 {
		return fmt.Errorf("backup: close rollback dbs: %v", errs)
	}
	return nil
}

// SetWindow overrides the default 1-hour rollback window for
// RevertLastSession.
func (r *Rollback) SetWindow(d time.Duration) {
	if d > 0 {
		r.window = d
	}
}

// CreateCheckpoint records a rollback target at the current time.
// The checkpoint is persisted to the rollback_checkpoints table
// so it survives daemon restarts.
func (r *Rollback) CreateCheckpoint(ctx context.Context, reason string) (*Checkpoint, error) {
	if r.mainDB == nil {
		return nil, fmt.Errorf("backup: no main DB for checkpoint")
	}
	res, err := r.mainDB.ExecContext(ctx,
		`INSERT INTO rollback_checkpoints (created_at, reason) VALUES (?, ?)`,
		time.Now().UTC().Format(time.RFC3339Nano), reason)
	if err != nil {
		return nil, fmt.Errorf("backup: create checkpoint: %w", err)
	}
	id, _ := res.LastInsertId()
	return &Checkpoint{
		ID:        id,
		CreatedAt: time.Now().UTC(),
		Reason:    reason,
	}, nil
}

// LatestCheckpoint returns the most recent persisted checkpoint.
// Returns (nil, nil) if no checkpoint exists.
func (r *Rollback) LatestCheckpoint(ctx context.Context) (*Checkpoint, error) {
	if r.mainDB == nil {
		return nil, nil //nolint:nilnil // nil DB means no checkpoints possible
	}
	var cp Checkpoint
	var ts string
	err := r.mainDB.QueryRowContext(ctx,
		`SELECT id, created_at, reason FROM rollback_checkpoints ORDER BY id DESC LIMIT 1`,
	).Scan(&cp.ID, &ts, &cp.Reason)
	if err == sql.ErrNoRows {
		return nil, nil //nolint:nilnil // no checkpoint found is not an error
	}
	if err != nil {
		return nil, fmt.Errorf("backup: latest checkpoint: %w", err)
	}
	cp.CreatedAt, _ = time.Parse(time.RFC3339Nano, ts)
	return &cp, nil
}

// RevertToCheckpoint deletes Synaptic-owned rows inserted after the checkpoint.
//
//nolint:gocognit,gocyclo // multi-table rollback is intentionally sequential
func (r *Rollback) RevertToCheckpoint(ctx context.Context, cp Checkpoint) (int, error) {
	cutoff := cp.CreatedAt.UTC().Format(time.RFC3339Nano)
	total := 0

	if r.mainDB != nil && hasColumn(ctx, r.mainDB, "conversation_messages", "created_at") {
		// Check conversations column BEFORE starting the transaction to avoid SQLite deadlock.
		hasConvCol := hasColumn(ctx, r.mainDB, "conversations", "created_at")
		tx, err := r.mainDB.BeginTx(ctx, nil)
		if err != nil {
			return 0, fmt.Errorf("backup: rollback begin: %w", err)
		}
		res, err := tx.ExecContext(ctx,
			`DELETE FROM conversation_messages WHERE created_at > ?`, cutoff)
		if err != nil {
			_ = tx.Rollback()
			return total, fmt.Errorf("backup: rollback messages: %w", err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			total += int(n)
		}
		if hasConvCol {
			res, err = tx.ExecContext(ctx,
				`DELETE FROM conversations WHERE created_at > ?`, cutoff)
			if err != nil {
				_ = tx.Rollback()
				return total, fmt.Errorf("backup: rollback conversations: %w", err)
			}
			if n, _ := res.RowsAffected(); n > 0 {
				total += int(n)
			}
		}
		if err := tx.Commit(); err != nil {
			return total, fmt.Errorf("backup: rollback commit: %w", err)
		}
	}

	if r.memoryDB != nil {
		for _, q := range []struct {
			table string
			col   string
		}{
			{"episodes", "timestamp"},
			{"facts", "created_at"},
		} {
			if !hasColumn(ctx, r.memoryDB, q.table, q.col) {
				continue
			}
			res, err := r.memoryDB.ExecContext(ctx,
				fmt.Sprintf(`DELETE FROM %s WHERE %s > ?`, q.table, q.col), cutoff)
			if err != nil {
				return total, fmt.Errorf("backup: rollback memory %s: %w", q.table, err)
			}
			if n, _ := res.RowsAffected(); n > 0 {
				total += int(n)
			}
		}
	}

	if r.skillsDB != nil && hasColumn(ctx, r.skillsDB, "skills", "created_at") {
		res, err := r.skillsDB.ExecContext(ctx,
			`DELETE FROM skills WHERE created_at > ? AND source = 'auto'`, cutoff)
		if err != nil {
			return total, fmt.Errorf("backup: rollback skills: %w", err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			total += int(n)
		}
	}

	return total, nil
}

// hasColumn reports whether the given table has the given column.
func hasColumn(ctx context.Context, db *sql.DB, table, column string) bool {
	rows, err := db.QueryContext(ctx,
		`SELECT 1 FROM pragma_table_info(?) WHERE name = ? LIMIT 1`,
		table, column)
	if err != nil {
		return false
	}
	defer func() { _ = rows.Close() }()
	return rows.Next()
}

// RevertLastSession reverts rows inserted within the configured
// rollback window (default 1h, configurable via SetWindow).
func (r *Rollback) RevertLastSession(ctx context.Context) (int, error) {
	cp := Checkpoint{
		ID:        1,
		CreatedAt: time.Now().UTC().Add(-r.window),
		Reason:    "revert last session",
	}
	return r.RevertToCheckpoint(ctx, cp)
}

// HonestScope returns a human-readable statement of what Rollback
// can and cannot do.
func (r *Rollback) HonestScope() string {
	return "Rollback will delete Synaptic-owned rows inserted after the " +
		"checkpoint: conversation history, memory episodes, and " +
		"auto-created skills. It will NOT undo OS-level actions like " +
		"sent emails, deleted files, made purchases, or sent messages. " +
		"For those, see Action Replay."
}
