// Package audit is an append-only, HMAC-chained audit log for the daemon.
//
// Every security-relevant action is recorded here with timestamp, actor,
// action, and outcome. Phase 2 (sub-phase 2.6) reads from this log to
// power the audit-log viewer in the GUI. Phase 11 (Trust & Recovery)
// enriches the Event with structured fields (Kind, BlastClass, Verdict,
// TargetApp/URL/Path/Command, ConsentResult, screenshot refs, SessionID)
// and a SHA-256 HMAC chain so Action Replay can detect tampering.
//
// The HMAC chain (MISSION §5.4): each row stores `prev_hash` (the
// hex SHA-256 of the previous row's hmac, or 64 zeros for the first row)
// and `hmac` (the hex SHA-256 of the canonical serialization of this
// row's payload, excluding the hmac column itself). Any modification to
// a past row invalidates every subsequent row's hmac.
package audit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ErrEventNotFound is returned by GetByID when no row exists for
// the given id. Callers compare with errors.Is.
var ErrEventNotFound = errors.New("audit: event not found")

// Event is one row in the audit log. The structured fields added in
// Phase 11 (Kind, BlastClass, Verdict, TargetApp/URL/Path/Command,
// ConsentResult, ScreenshotBeforeRef, ScreenshotAfterRef, SessionID)
// power Action Replay (11A) — they let the Replay reconstruct a
// scrubbable 24h timeline without string-parsing the Message column.
type Event struct {
	ID      int64     `json:"id"`
	TS      time.Time `json:"ts"`
	Actor   string    `json:"actor"`
	Action  string    `json:"action"`
	App     string    `json:"app"`
	Level   string    `json:"level"`
	Result  string    `json:"result"`
	Message string    `json:"message"`

	// Structured fields (Phase 11). All optional; empty means "not
	// applicable" or "not yet recorded". The audit log must remain
	// backward-compatible — old code keeps working.
	Kind          string `json:"kind,omitempty"`
	BlastClass    string `json:"blast_class,omitempty"`
	Verdict       string `json:"verdict,omitempty"`
	TargetApp     string `json:"target_app,omitempty"`
	TargetURL     string `json:"target_url,omitempty"`
	Path          string `json:"path,omitempty"`
	Command       string `json:"command,omitempty"`
	ConsentResult string `json:"consent_result,omitempty"`
	SSBeforeRef   string `json:"screenshot_before_ref,omitempty"`
	SSAfterRef    string `json:"screenshot_after_ref,omitempty"`
	SessionID     string `json:"session_id,omitempty"`

	// HMAC chain fields. Not exported via JSON for now — internal.
	prevHash string `json:"-"`
	hmac     string `json:"-"`
}

// Query filters for List.
type Query struct {
	Limit  int
	Offset int
	Since  time.Time
	Action string
	Level  string
	Kind   string
}

// Log is the audit log. Construct once at startup; share across handlers.
//
// The HMAC chain secret is injected at construction time. If the secret
// is empty, the log falls back to a deterministic zero-key — this is
// only for tests, not production. The secret is held in memory only;
// it's the master key (or a key derived from it), not persisted with
// the log itself.
type Log struct {
	db     *sql.DB
	secret []byte
	mu     sync.Mutex // serializes Append so the chain is consistent
}

// New returns a Log wrapping the given database. The secret is the
// HMAC key used to chain entries. Pass the same master key that
// protects the rest of the database (or a key derived from it).
func New(db *sql.DB, secret []byte) *Log {
	if len(secret) == 0 {
		// Tests / ephemeral mode: deterministic zero key. The chain
		// is still well-formed (the hmac values are still computed
		// correctly), it just isn't secret. This is intentionally
		// permissive — an empty secret in production is the caller's
		// bug, not a panic.
		secret = make([]byte, 32)
	}
	return &Log{db: db, secret: secret}
}

// NewWithHexSecret is a convenience for callers that store the secret
// as a hex string (e.g. in config).
func NewWithHexSecret(db *sql.DB, hexSecret string) *Log {
	if hexSecret == "" {
		return New(db, nil)
	}
	b, err := hex.DecodeString(hexSecret)
	if err != nil {
		// Fall back to zero key — log will be chained but not secret.
		return New(db, nil)
	}
	return New(db, b)
}

// genesisHash is the prev_hash value for the first row in the chain.
// It is 64 ASCII zeros, which never collides with any real hmac
// (a real hmac is a 64-character hex string but won't be all zeros
// for any non-trivial secret).
const genesisHash = "0000000000000000000000000000000000000000000000000000000000000000"

// Append records one event. The TS is set to time.Now() if zero.
// Append serializes the chain write so the prev_hash/next hmac
// relationship is correct even under concurrent callers.
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
	if e.Action == "" {
		return errors.New("audit: Event.Action is required")
	}
	if e.Actor == "" {
		return errors.New("audit: Event.Actor is required")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Look up the prev_hash: the hmac of the most recent row, or the
	// genesis hash for the first row.
	prevHash, err := l.lastHMAC(ctx)
	if err != nil {
		return fmt.Errorf("audit: lookup prev_hash: %w", err)
	}
	if prevHash == "" {
		prevHash = genesisHash
	}
	e.prevHash = prevHash

	// Insert the row with an empty hmac first to get the id, then
	// update the hmac column. We use a transaction so the chain
	// cannot end up with a half-computed row.
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("audit: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO audit_log (
			ts, actor, action, app, level, result, message,
			kind, blast_class, verdict,
			target_app, target_url, path, command,
			consent_result,
			screenshot_before_ref, screenshot_after_ref,
			session_id,
			prev_hash, hmac
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.TS.Format(time.RFC3339Nano), e.Actor, e.Action, e.App, e.Level, e.Result, e.Message,
		e.Kind, e.BlastClass, e.Verdict,
		e.TargetApp, e.TargetURL, e.Path, e.Command,
		e.ConsentResult,
		e.SSBeforeRef, e.SSAfterRef,
		e.SessionID,
		e.prevHash, "",
	)
	if err != nil {
		return fmt.Errorf("audit: insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("audit: last insert id: %w", err)
	}
	e.ID = id

	// Compute hmac over the canonical serialization, then update.
	e.hmac = l.computeHMAC(e)
	if _, err := tx.ExecContext(ctx,
		`UPDATE audit_log SET hmac = ? WHERE id = ?`,
		e.hmac, e.ID,
	); err != nil {
		return fmt.Errorf("audit: update hmac: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("audit: commit: %w", err)
	}
	return nil
}

// List returns events matching q, ordered by ts desc.
func (l *Log) List(ctx context.Context, q Query) ([]Event, error) {
	if q.Limit <= 0 || q.Limit > 1000 {
		q.Limit = 100
	}
	query := `SELECT id, ts, actor, action, app, level, result, message,
		         kind, blast_class, verdict,
		         target_app, target_url, path, command,
		         consent_result,
		         screenshot_before_ref, screenshot_after_ref,
		         session_id,
		         prev_hash, hmac
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
	if q.Kind != "" {
		query += ` AND kind = ?`
		args = append(args, q.Kind)
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
		if err := rows.Scan(
			&e.ID, &ts, &e.Actor, &e.Action, &e.App, &e.Level, &e.Result, &e.Message,
			&e.Kind, &e.BlastClass, &e.Verdict,
			&e.TargetApp, &e.TargetURL, &e.Path, &e.Command,
			&e.ConsentResult,
			&e.SSBeforeRef, &e.SSAfterRef,
			&e.SessionID,
			&e.prevHash, &e.hmac,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		e.TS, _ = time.Parse(time.RFC3339Nano, ts)
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetByID returns one event by id, with its hmac chain fields populated.
func (l *Log) GetByID(ctx context.Context, id int64) (*Event, error) {
	row := l.db.QueryRowContext(ctx, `
		SELECT id, ts, actor, action, app, level, result, message,
		       kind, blast_class, verdict,
		       target_app, target_url, path, command,
		       consent_result,
		       screenshot_before_ref, screenshot_after_ref,
		       session_id,
		       prev_hash, hmac
		FROM audit_log WHERE id = ?`, id)
	var e Event
	var ts string
	err := row.Scan(
		&e.ID, &ts, &e.Actor, &e.Action, &e.App, &e.Level, &e.Result, &e.Message,
		&e.Kind, &e.BlastClass, &e.Verdict,
		&e.TargetApp, &e.TargetURL, &e.Path, &e.Command,
		&e.ConsentResult,
		&e.SSBeforeRef, &e.SSAfterRef,
		&e.SessionID,
		&e.prevHash, &e.hmac,
	)
	if err == sql.ErrNoRows {
		return nil, ErrEventNotFound
	}
	if err != nil {
		return nil, err
	}
	e.TS, _ = time.Parse(time.RFC3339Nano, ts)
	return &e, nil
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

// VerifyChain walks the audit log in id order and confirms that:
//   - every row's prev_hash equals the prior row's hmac (or the
//     genesis hash for the first row), and
//   - every row's hmac matches the recomputed value over its payload.
//
// Returns the id of the first row that fails verification (or 0 if all
// rows pass), and the row's stored vs. computed hmac for diagnostic
// output. The walker is bounded by limit so a caller can verify a
// recent slice first; pass limit=0 to verify the entire log.
//
// The chain is verified in chronological (id ascending) order so that
// a single tampering point invalidates exactly one verification result
// and everything after it.
func (l *Log) VerifyChain(ctx context.Context, sinceID int64, limit int) (*ChainReport, error) {
	rep := &ChainReport{}
	query := `
		SELECT id, ts, actor, action, app, level, result, message,
		       kind, blast_class, verdict,
		       target_app, target_url, path, command,
		       consent_result,
		       screenshot_before_ref, screenshot_after_ref,
		       session_id,
		       prev_hash, hmac
		FROM audit_log
		WHERE id >= ?
		ORDER BY id ASC`
	//nolint:gosec // G202: limit is a validated non-negative int.
	query += limitClause(limit)
	rows, err := l.db.QueryContext(ctx, query, sinceID)
	if err != nil {
		return nil, fmt.Errorf("verify chain: query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	expectedPrev := genesisHash
	for rows.Next() {
		var e Event
		var ts string
		if err := rows.Scan(
			&e.ID, &ts, &e.Actor, &e.Action, &e.App, &e.Level, &e.Result, &e.Message,
			&e.Kind, &e.BlastClass, &e.Verdict,
			&e.TargetApp, &e.TargetURL, &e.Path, &e.Command,
			&e.ConsentResult,
			&e.SSBeforeRef, &e.SSAfterRef,
			&e.SessionID,
			&e.prevHash, &e.hmac,
		); err != nil {
			return nil, fmt.Errorf("verify chain: scan: %w", err)
		}
		e.TS, _ = time.Parse(time.RFC3339Nano, ts)
		rep.RowsChecked++

		// Check that this row's prev_hash matches what we expect.
		if e.prevHash != expectedPrev {
			rep.FirstBreakID = e.ID
			rep.FirstBreakReason = "prev_hash mismatch (chain link broken at this row)"
			return rep, nil
		}
		// Recompute the hmac and confirm it matches.
		want := l.computeHMAC(e)
		if !hmac.Equal([]byte(e.hmac), []byte(want)) {
			rep.FirstBreakID = e.ID
			rep.FirstBreakReason = "hmac mismatch (row payload was modified after signing)"
			return rep, nil
		}
		expectedPrev = e.hmac
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("verify chain: rows: %w", err)
	}
	rep.Valid = true
	return rep, nil
}

// ChainReport is the result of VerifyChain.
type ChainReport struct {
	Valid            bool   `json:"valid"`
	RowsChecked      int    `json:"rows_checked"`
	FirstBreakID     int64  `json:"first_break_id,omitempty"`
	FirstBreakReason string `json:"first_break_reason,omitempty"`
}

// lastHMAC returns the hmac of the most recently inserted audit row.
// Returns the empty string if the table is empty.
func (l *Log) lastHMAC(ctx context.Context) (string, error) {
	var hmacValue string
	err := l.db.QueryRowContext(ctx,
		`SELECT hmac FROM audit_log ORDER BY id DESC LIMIT 1`,
	).Scan(&hmacValue)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return hmacValue, nil
}

// computeHMAC returns the hex SHA-256 HMAC of the canonical serialization
// of e, using the log's secret. The hmac column itself is excluded
// from the payload.
func (l *Log) computeHMAC(e Event) string {
	// Canonical payload: pipe-separated, stable field order. The format
	// is internal — callers must never parse this. We only need it to be
	// deterministic and to cover every field that contributes to identity.
	var sb strings.Builder
	sb.WriteString(strconv.FormatInt(e.ID, 10))
	sb.WriteByte('|')
	sb.WriteString(e.TS.UTC().Format(time.RFC3339Nano))
	sb.WriteByte('|')
	sb.WriteString(e.Actor)
	sb.WriteByte('|')
	sb.WriteString(e.Action)
	sb.WriteByte('|')
	sb.WriteString(e.App)
	sb.WriteByte('|')
	sb.WriteString(e.Level)
	sb.WriteByte('|')
	sb.WriteString(e.Result)
	sb.WriteByte('|')
	sb.WriteString(e.Message)
	sb.WriteByte('|')
	sb.WriteString(e.Kind)
	sb.WriteByte('|')
	sb.WriteString(e.BlastClass)
	sb.WriteByte('|')
	sb.WriteString(e.Verdict)
	sb.WriteByte('|')
	sb.WriteString(e.TargetApp)
	sb.WriteByte('|')
	sb.WriteString(e.TargetURL)
	sb.WriteByte('|')
	sb.WriteString(e.Path)
	sb.WriteByte('|')
	sb.WriteString(e.Command)
	sb.WriteByte('|')
	sb.WriteString(e.ConsentResult)
	sb.WriteByte('|')
	sb.WriteString(e.SSBeforeRef)
	sb.WriteByte('|')
	sb.WriteString(e.SSAfterRef)
	sb.WriteByte('|')
	sb.WriteString(e.SessionID)
	sb.WriteByte('|')
	sb.WriteString(e.prevHash)

	mac := hmac.New(sha256.New, l.secret)
	mac.Write([]byte(sb.String()))
	return hex.EncodeToString(mac.Sum(nil))
}

// limitClause appends a LIMIT clause only when limit is a positive
// integer. The value is converted with strconv.Itoa, so no SQL injection
// vector exists for integer input.
//
//nolint:gosec // G202: limit is a validated non-negative int.
func limitClause(limit int) string {
	if limit <= 0 {
		return ""
	}
	return " LIMIT " + strconv.Itoa(limit)
}
