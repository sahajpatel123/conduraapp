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
}

// NewRollback returns a Rollback helper for the main database only.
func NewRollback(db *sql.DB) *Rollback {
	return &Rollback{mainDB: db}
}

// NewRollbackMulti returns a Rollback that can revert conversations
// (main DB), memory episodes/facts (memory DB), and skill rows
// (skills DB). Nil DB pointers are skipped.
func NewRollbackMulti(mainDB, memoryDB, skillsDB *sql.DB) *Rollback {
	return &Rollback{mainDB: mainDB, memoryDB: memoryDB, skillsDB: skillsDB}
}

// CreateCheckpoint records a rollback target at the current time.
func (r *Rollback) CreateCheckpoint(_ context.Context, reason string) (*Checkpoint, error) {
	now := time.Now().UTC()
	return &Checkpoint{ID: 1, CreatedAt: now, Reason: reason}, nil
}

// RevertToCheckpoint deletes Synaptic-owned rows inserted after
// the checkpoint. Returns the total rows deleted.
func (r *Rollback) RevertToCheckpoint(ctx context.Context, cp Checkpoint) (int, error) {
	cutoff := cp.CreatedAt.UTC().Format(time.RFC3339Nano)
	total := 0

	if r.mainDB != nil && hasColumn(ctx, r.mainDB, "conversation_messages", "created_at") {
		res, err := r.mainDB.ExecContext(ctx,
			`DELETE FROM conversation_messages WHERE created_at > ?`, cutoff)
		if err != nil {
			return total, fmt.Errorf("backup: rollback messages: %w", err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			total += int(n)
		}
		if hasColumn(ctx, r.mainDB, "conversations", "created_at") {
			res, err = r.mainDB.ExecContext(ctx,
				`DELETE FROM conversations WHERE created_at > ?`, cutoff)
			if err != nil {
				return total, fmt.Errorf("backup: rollback conversations: %w", err)
			}
			if n, _ := res.RowsAffected(); n > 0 {
				total += int(n)
			}
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

// RevertLastSession reverts rows inserted in the last hour as a
// practical "undo my recent session" action.
func (r *Rollback) RevertLastSession(ctx context.Context) (int, error) {
	cp := Checkpoint{
		ID:        1,
		CreatedAt: time.Now().UTC().Add(-1 * time.Hour),
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
