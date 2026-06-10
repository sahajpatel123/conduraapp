// Package audit is an append-only audit log for the daemon.
// Every security-relevant action is recorded here with timestamp,
// actor, action, and outcome. Phase 2 (sub-phase 2.6) reads from
// this log to power the audit-log viewer in the GUI.
package audit

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Event is one row in the audit log.
type Event struct {
	ID      int64     `json:"id"`
	TS      time.Time `json:"ts"`
	Actor   string    `json:"actor"`
	Action  string    `json:"action"`
	App     string    `json:"app"`
	Level   string    `json:"level"`
	Result  string    `json:"result"`
	Message string    `json:"message"`
}

// Query filters for List.
type Query struct {
	Limit  int
	Offset int
	Since  time.Time
	Action string
	Level  string
}

// Log is the audit log. Construct once at startup; share across handlers.
type Log struct {
	db *sql.DB
}

// New returns a Log wrapping the given database.
func New(db *sql.DB) *Log {
	return &Log{db: db}
}

// Append records one event. The TS is set to time.Now() if zero.
func (l *Log) Append(ctx context.Context, e Event) error {
	if l == nil {
		return nil
	}
	if e.TS.IsZero() {
		e.TS = time.Now().UTC()
	}
	if e.Level == "" {
		e.Level = "info"
	}
	if e.Result == "" {
		e.Result = "allow"
	}
	_, err := l.db.ExecContext(ctx,
		`INSERT INTO audit_log (ts, actor, action, app, level, result, message)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.TS.Format(time.RFC3339Nano), e.Actor, e.Action, e.App, e.Level, e.Result, e.Message,
	)
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}

// List returns events matching q, ordered by ts desc.
func (l *Log) List(ctx context.Context, q Query) ([]Event, error) {
	if q.Limit <= 0 || q.Limit > 1000 {
		q.Limit = 100
	}
	query := `SELECT id, ts, actor, action, app, level, result, message
	          FROM audit_log WHERE 1=1`
	args := []interface{}{}
	if !q.Since.IsZero() {
		query += ` AND ts >= ?`
		args = append(args, q.Since.Format(time.RFC3339Nano))
	}
	if q.Action != "" {
		query += ` AND action LIKE ?`
		args = append(args, "%"+q.Action+"%")
	}
	if q.Level != "" {
		query += ` AND level = ?`
		args = append(args, q.Level)
	}
	query += ` ORDER BY ts DESC LIMIT ? OFFSET ?`
	args = append(args, q.Limit, q.Offset)
	rows, err := l.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit log: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []Event
	for rows.Next() {
		var e Event
		var ts string
		if err := rows.Scan(&e.ID, &ts, &e.Actor, &e.Action, &e.App, &e.Level, &e.Result, &e.Message); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		e.TS, _ = time.Parse(time.RFC3339Nano, ts)
		out = append(out, e)
	}
	return out, nil
}

// Count returns the total number of events in the log.
func (l *Log) Count(ctx context.Context) (int, error) {
	var n int
	row := l.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_log`)
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
