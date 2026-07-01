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
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/hkdf"
)

// auditInfo is the HKDF info string for deriving the audit HMAC subkey
// from a shared master key. Domain separation prevents the same master
// key from protecting both API-key encryption (AES-GCM) and the audit
// HMAC chain — a compromised master key used directly in the HMAC
// would otherwise also forge audit rows. The "v1" suffix lets us bump
// the scheme without breaking verification of existing chains.
const auditInfo = "condura-audit-hmac-v1"

// auditSubKeyLen is the derived subkey length in bytes (32 = SHA-256).
const auditSubKeyLen = 32

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
// The HMAC chain secret is injected at construction time as a raw
// master key. New() derives a domain-separated subkey via HKDF-SHA-256
// (info = "condura-audit-hmac-v1", 32 bytes) so the audit HMAC is
// never computed with the raw master key. This prevents a compromised
// master key from being usable to forge audit rows: the audit subkey
// is independent from the master keys used by the secret store (which
// uses AES-GCM with the same master) and any future HMAC consumers.
//
// The derived subkey is deterministic, so existing chains written
// before this change remain verifiable.
type Log struct {
	db     *sql.DB
	subkey []byte     // derived audit HMAC subkey (HKDF-SHA-256 output)
	mu     sync.Mutex // serializes Append so the chain is consistent
}

// New returns a Log wrapping the given database. secret is the
// caller-provided master key (the same key that protects the secret
// store and other encrypted fields). New derives a domain-separated
// audit subkey via HKDF-SHA-256 with info="condura-audit-hmac-v1" and
// uses only that subkey for HMAC computation. An empty secret is a
// programming error; callers must provide a non-empty key.
func New(db *sql.DB, secret []byte) *Log {
	if len(secret) == 0 {
		panic("audit: empty HMAC secret")
	}
	subkey := deriveAuditSubkey(secret)
	return &Log{db: db, subkey: subkey}
}

// deriveAuditSubkey returns HKDF-SHA-256(masterKey, salt=nil,
// info=auditInfo, len=auditSubKeyLen). The info string provides
// domain separation so the audit subkey cannot be reused for any
// other HMAC or AEAD construction that shares the master key. The
// output is deterministic for a given (masterKey, info) pair, so
// chains written with this scheme remain verifiable across process
// restarts and across machines with the same master key.
func deriveAuditSubkey(masterKey []byte) []byte {
	r := hkdf.New(sha256.New, masterKey, nil, []byte(auditInfo))
	out := make([]byte, auditSubKeyLen)
	if _, err := io.ReadFull(r, out); err != nil {
		// hkdf-derived readers cannot return an error before the
		// requested length is satisfied unless the underlying hash
		// fails, which sha256.New cannot do. Treat as programmer
		// panic if it ever does.
		panic(fmt.Sprintf("audit: HKDF subkey derivation failed: %v", err))
	}
	return out
}

// Reload replaces the underlying database connection. Use this after
// a backup restore (storage.Reload) so the audit log writes to the
// new DB handle instead of the closed old one.
func (l *Log) Reload(db *sql.DB) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.db = db
}

// NewWithHexSecret is a convenience for callers that store the secret
// as a hex string (e.g. in config). An empty or invalid hex secret
// returns an error.
func NewWithHexSecret(db *sql.DB, hexSecret string) (*Log, error) {
	if hexSecret == "" {
		return nil, errors.New("audit: empty HMAC secret")
	}
	b, err := hex.DecodeString(hexSecret)
	if err != nil {
		return nil, fmt.Errorf("audit: invalid hex secret: %w", err)
	}
	return New(db, b), nil
}

// genesisHash is the prev_hash value for the first row in the chain.
// It is 64 ASCII zeros, which never collides with any real hmac
// (a real hmac is a 64-character hex string but won't be all zeros
// for any non-trivial secret).
const genesisHash = "0000000000000000000000000000000000000000000000000000000000000000"

// Append records one event. The TS is set to time.Now() if zero.
// Append serializes the chain write so the prev_hash/next hmac
// relationship is correct even under concurrent callers.
//
// IMPORTANT: Append does NOT redact the Message field. Production
// call sites that pass user-derived text (utterances, paths,
// reasons, error messages) MUST wrap Message with
// sanitize.RedactSecrets before calling Append. AppendForTest
// exists for tests and DOES redact by default — use it instead
// of Append in any test that exercises user-derived content.
func (l *Log) Append(ctx context.Context, e Event) error {
	if l == nil {
		return errors.New("audit: nil Log")
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
	// Secondary sort by id keeps ordering deterministic when multiple
	// events share an identical timestamp (observed as flaky frame
	// ordering on fast clocks, e.g. Windows CI runners).
	query += ` ORDER BY ts DESC, id DESC LIMIT ? OFFSET ?` //nolint:gosec // limit/offset are validated to int; no injection surface
	args = append(args, q.Limit, q.Offset)
	rows, err := l.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit log: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := make([]Event, 0)
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
	return l.scanEvent(row)
}

// rechainFromTx rewrites prev_hash and hmac for every audit_log row
// with id > startID, in id order, using the current hmac of the
// startID row as the seed. Each row's prev_hash is set to the
// previous row's NEW hmac, and each row's hmac is recomputed over
// its canonical payload. This is what makes a post-Prune log
// verifiable by VerifyChain: the chain is genuinely re-rooted at
// the oldest surviving row, not just hand-waved via a partial
// prev_hash reset.
//
// Called from Prune inside its transaction so the rewrite is
// atomic with the delete + tombstone write.
func (l *Log) rechainFromTx(ctx context.Context, tx *sql.Tx, startID int64) error {
	// Fetch the start row's NEW hmac (already rewritten in Prune
	// before this is called). All subsequent rows will chain off
	// of this.
	var prev string
	if err := tx.QueryRowContext(ctx,
		`SELECT hmac FROM audit_log WHERE id = ?`, startID,
	).Scan(&prev); err != nil {
		return fmt.Errorf("load start hmac: %w", err)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, ts, actor, action, app, level, result, message,
		       kind, blast_class, verdict,
		       target_app, target_url, path, command,
		       consent_result,
		       screenshot_before_ref, screenshot_after_ref,
		       session_id,
		       prev_hash, hmac
		FROM audit_log WHERE id > ? ORDER BY id ASC`, startID)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	type update struct {
		id       int64
		prevHash string
		hmac     string
	}
	var pending []update
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
			return fmt.Errorf("scan: %w", err)
		}
		e.TS, _ = time.Parse(time.RFC3339Nano, ts)
		e.prevHash = prev
		pending = append(pending, update{
			id:       e.ID,
			prevHash: prev,
			hmac:     l.computeHMAC(e),
		})
		prev = pending[len(pending)-1].hmac
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows: %w", err)
	}
	// Apply updates in order. We batch into a single statement to
	// keep the tx short even on large survivors (rare in practice:
	// Prune runs once per retention window, and a 90-day retention
	// rarely has more than a few thousand rows rewritten).
	for _, u := range pending {
		if _, err := tx.ExecContext(ctx,
			`UPDATE audit_log SET prev_hash = ?, hmac = ? WHERE id = ?`,
			u.prevHash, u.hmac, u.id); err != nil {
			return fmt.Errorf("update id=%d: %w", u.id, err)
		}
	}
	return nil
}

// getByIDTx is the transaction-scoped variant of GetByID, used by
// Prune to read the oldest surviving row inside the prune tx.
func (l *Log) getByIDTx(ctx context.Context, tx *sql.Tx, id int64) (*Event, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT id, ts, actor, action, app, level, result, message,
		       kind, blast_class, verdict,
		       target_app, target_url, path, command,
		       consent_result,
		       screenshot_before_ref, screenshot_after_ref,
		       session_id,
		       prev_hash, hmac
		FROM audit_log WHERE id = ?`, id)
	return l.scanEvent(row)
}

// scanEvent reads a single row into an Event. Shared by GetByID and
// getByIDTx so the column list stays in one place.
func (l *Log) scanEvent(row *sql.Row) (*Event, error) {
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

// Prune deletes audit rows older than the given retention window
// (per CLAUDE.md §10.5: "90-day retention (configurable)"). Returns
// the number of rows deleted.
//
// The audit log is HMAC-chained, so deleting old rows would break
// the chain for the surviving rows (their prev_hash would point at
// deleted rows). Prune handles this by, after the delete, resetting
// the oldest surviving row's prev_hash to the genesis hash (64 zeros)
// and recomputing its hmac. The pruned log is therefore a valid
// standalone chain starting at the oldest surviving row.
//
// Tamper-evidence (CLAUDE.md §2.1 invariant #5, "never deleted"):
// a plain delete leaves no record that rows were ever removed.
// VerifyChain on a post-prune log would report Valid=true with no
// way to distinguish "50 rows existed" from "100 existed, 50
// pruned". To close that gap, Prune inserts a tombstone row into
// the prune_tombstone table inside the same transaction, recording
// the deleted count, the oldest surviving row's id+hmac at
// prune-time (a forensic anchor), the wall-clock timestamp, and the
// retention window in days. Forensic queries use PruneTombstones to
// reconstruct the pre-prune chain starting from oldest_surviving_id.
//
// A zero or negative retention window is a no-op (retention disabled)
// and writes no tombstone.
func (l *Log) Prune(ctx context.Context, retention time.Duration) (int64, error) {
	if retention <= 0 {
		return 0, nil
	}
	return l.pruneWithCutoff(ctx, time.Now().UTC(), time.Now().Add(-retention).Format(time.RFC3339Nano), int64(retention/(24*time.Hour)))
}

// pruneWithCutoff performs the actual prune given an explicit cutoff
// timestamp (RFC3339Nano), now time, and retention window in days.
// Extracted from Prune to keep the public function's complexity under
// the golangci-lint gocyclo threshold while preserving the same
// behavior. The two helpers share the same transaction shape; the
// only difference is where the timestamps come from.
func (l *Log) pruneWithCutoff(ctx context.Context, now time.Time, cutoff string, retentionDays int64) (int64, error) {
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("audit: prune: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx,
		`DELETE FROM audit_log WHERE ts < ?`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("audit: prune: delete: %w", err)
	}
	deleted, _ := res.RowsAffected()

	// If nothing was deleted, this was a true no-op: no tombstone
	// written, no chain rewrite needed. This is the path for
	// "retention set but no rows aged out yet" — common at
	// startup when the log is fresh.
	if deleted == 0 {
		if err := tx.Commit(); err != nil {
			return 0, fmt.Errorf("audit: prune: commit (no-op): %w", err)
		}
		return 0, nil
	}

	oldestID, preRewriteHMAC, err := l.snapshotOldestSurviving(ctx, tx)
	if err != nil {
		return 0, fmt.Errorf("audit: prune: snapshot oldest: %w", err)
	}

	if oldestID != 0 {
		if err := l.rechainRootAndRest(ctx, tx, oldestID); err != nil {
			return 0, fmt.Errorf("audit: prune: rechain: %w", err)
		}
	}

	if err := l.writeTombstone(ctx, tx, deleted, oldestID, preRewriteHMAC, now, retentionDays); err != nil {
		return 0, fmt.Errorf("audit: prune: tombstone: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("audit: prune: commit: %w", err)
	}
	return deleted, nil
}

// snapshotOldestSurviving returns the id and pre-rewrite hmac of the
// oldest row still in audit_log after a prune delete. If the table
// is now empty, returns (0, "", nil). Used by Prune to capture the
// forensic anchor for the tombstone row.
func (l *Log) snapshotOldestSurviving(ctx context.Context, tx *sql.Tx) (int64, string, error) {
	var oldestID int64
	postRow := tx.QueryRowContext(ctx,
		`SELECT id FROM audit_log ORDER BY id ASC LIMIT 1`)
	if err := postRow.Scan(&oldestID); err != nil &&
		err != sql.ErrNoRows {
		return 0, "", fmt.Errorf("audit: prune: find oldest surviving: %w", err)
	}
	if oldestID == 0 {
		return 0, "", nil
	}
	var oldestHMAC string
	hmacRow := tx.QueryRowContext(ctx,
		`SELECT hmac FROM audit_log WHERE id = ?`, oldestID)
	if err := hmacRow.Scan(&oldestHMAC); err != nil {
		return 0, "", fmt.Errorf("audit: prune: snapshot oldest hmac: %w", err)
	}
	return oldestID, oldestHMAC, nil
}

// rechainRootAndRest rewrites the oldest surviving row to point at
// the genesis hash (the chain is re-rooted here) and then walks
// every subsequent row in id order, rewriting each row's prev_hash
// to the prior row's NEW hmac. Without the second step,
// VerifyChain would correctly report a break at row 2 because its
// prev_hash would still reference the OLD hmac of the oldest
// surviving row (the one we just changed).
func (l *Log) rechainRootAndRest(ctx context.Context, tx *sql.Tx, oldestID int64) error {
	e, err := l.getByIDTx(ctx, tx, oldestID)
	if err != nil {
		return fmt.Errorf("load oldest: %w", err)
	}
	e.prevHash = genesisHash
	e.hmac = l.computeHMAC(*e)
	if _, err := tx.ExecContext(ctx,
		`UPDATE audit_log SET prev_hash = ?, hmac = ? WHERE id = ?`,
		e.prevHash, e.hmac, e.ID); err != nil {
		return fmt.Errorf("rechain oldest: %w", err)
	}
	return l.rechainFromTx(ctx, tx, oldestID)
}

// writeTombstone records the prune event. The oldest_surviving_hmac
// column carries the PRE-rewrite hmac so a forensic re-walk can
// distinguish "row was rewritten by prune" from "row is missing".
func (l *Log) writeTombstone(ctx context.Context, tx *sql.Tx, deleted, oldestID int64, preRewriteHMAC string, now time.Time, retentionDays int64) error {
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO prune_tombstone (
			pruned_count, oldest_surviving_id, oldest_surviving_hmac,
			pruned_at, retention_window_days
		) VALUES (?, ?, ?, ?, ?)`,
		deleted, oldestID, preRewriteHMAC,
		now.Format(time.RFC3339Nano), retentionDays,
	); err != nil {
		return fmt.Errorf("tombstone: %w", err)
	}
	return nil
}

// Old Prune body below retained for reference — replaced by
// pruneWithCutoff above. Kept commented so the chain rewrite logic
// stays in one place for readers.
//
//nolint:unused // reference-only comment block, intentionally dead code

// Tombstone is one prune event. The combination of
// (oldest_surviving_id, oldest_surviving_hmac) is the forensic
// anchor a reader uses to walk back further if they have a
// pre-prune backup. Tombstones are append-only — Prune never
// deletes a tombstone, and there is no public method to do so.
type Tombstone struct {
	ID                  int64     `json:"id"`
	PrunedCount         int64     `json:"pruned_count"`
	OldestSurvivingID   int64     `json:"oldest_surviving_id"`
	OldestSurvivingHMAC string    `json:"oldest_surviving_hmac"`
	PrunedAt            time.Time `json:"pruned_at"`
	RetentionWindowDays int       `json:"retention_window_days"`
}

// PruneTombstones returns all prune tombstones, ordered by
// pruned_at DESC (most recent first). Callers use this for
// forensic queries: "how many rows have ever been pruned from
// this log?" and "what was the chain root at prune-time?"
func (l *Log) PruneTombstones(ctx context.Context) ([]Tombstone, error) {
	rows, err := l.db.QueryContext(ctx, `
		SELECT id, pruned_count, oldest_surviving_id,
		       oldest_surviving_hmac, pruned_at, retention_window_days
		FROM prune_tombstone
		ORDER BY pruned_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("query prune tombstones: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]Tombstone, 0)
	for rows.Next() {
		var t Tombstone
		var ts string
		var oldestID sql.NullInt64
		if err := rows.Scan(&t.ID, &t.PrunedCount, &oldestID,
			&t.OldestSurvivingHMAC, &ts, &t.RetentionWindowDays); err != nil {
			return nil, fmt.Errorf("scan tombstone: %w", err)
		}
		t.OldestSurvivingID = oldestID.Int64
		t.PrunedAt, _ = time.Parse(time.RFC3339Nano, ts)
		out = append(out, t)
	}
	return out, rows.Err()
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

		// Legacy rows from before the HMAC-chain migration have empty
		// hmac/prev_hash; skip them without updating expectedPrev.
		if e.hmac == "" {
			continue
		}

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

// ChainHistoryReport is VerifyChain's verdict plus the full
// prune-history of the log. A Valid=true report with N tombstones
// means the surviving chain is internally consistent AND N prior
// prune events deleted rows from the log; a forensic reader can
// reconstruct the pre-prune chain by walking the tombstones in
// chronological order (oldest first) and chaining through their
// OldestSurvivingHMAC anchors.
//
// ChainHistoryReport.Valid mirrors ChainReport.Valid exactly; the
// two structs are deliberately separate so callers using the
// plain ChainReport (e.g. the GUI's live verification badge) are
// not affected by a behavior change. New code that wants the
// full forensic picture should call VerifyChainWithHistory.
type ChainHistoryReport struct {
	ChainReport
	Tombstones []Tombstone `json:"tombstones"`
}

// VerifyChainWithHistory is VerifyChain plus the prune-tombstone
// history. The chain verdict is identical to VerifyChain — the
// only difference is the additional Tombstones slice. A valid
// chain with tombstones is still valid; tombstones are
// informational, not a failure signal.
//
// Use this when the caller wants to render "X rows pruned
// historically" in the GUI, or when a forensic investigator
// needs to know the chain was rewritten.
func (l *Log) VerifyChainWithHistory(ctx context.Context, sinceID int64, limit int) (*ChainHistoryReport, error) {
	chain, err := l.VerifyChain(ctx, sinceID, limit)
	if err != nil {
		return nil, err
	}
	tombs, err := l.PruneTombstones(ctx)
	if err != nil {
		return nil, err
	}
	return &ChainHistoryReport{
		ChainReport: *chain,
		Tombstones:  tombs,
	}, nil
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
// of e, using the log's domain-separated audit subkey. The hmac column
// itself is excluded from the payload.
func (l *Log) computeHMAC(e Event) string {
	// Canonical payload: length-prefixed fields in a stable order. The
	// format is internal — callers must never parse this. Length-prefixing
	// prevents attacker-controlled '|' characters in fields from acting as
	// ambiguous delimiters, which would weaken the tamper-evidence
	// guarantee.
	var sb strings.Builder
	writeField := func(s string) {
		sb.WriteString(strconv.Itoa(len(s)))
		sb.WriteByte(':')
		sb.WriteString(s)
	}
	writeField(strconv.FormatInt(e.ID, 10))
	writeField(e.TS.UTC().Format(time.RFC3339Nano))
	writeField(e.Actor)
	writeField(e.Action)
	writeField(e.App)
	writeField(e.Level)
	writeField(e.Result)
	writeField(e.Message)
	writeField(e.Kind)
	writeField(e.BlastClass)
	writeField(e.Verdict)
	writeField(e.TargetApp)
	writeField(e.TargetURL)
	writeField(e.Path)
	writeField(e.Command)
	writeField(e.ConsentResult)
	writeField(e.SSBeforeRef)
	writeField(e.SSAfterRef)
	writeField(e.SessionID)
	writeField(e.prevHash)

	mac := hmac.New(sha256.New, l.subkey)
	if _, err := mac.Write([]byte(sb.String())); err != nil {
		panic(fmt.Sprintf("audit: hmac.Write failed: %v", err))
	}
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
