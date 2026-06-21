// Package pending implements the persistent queue of sub-agent
// ActionRequests that are awaiting user approval.
//
// When a sub-agent (Claude Code, Codex, etc.) returns a structured
// ActionRequest, the daemon persists it here so the GUI can render a
// "pending actions" panel. Each row carries the gate verdict (so the
// GUI can show "the gate said this needs your approval because …"),
// an expiry deadline (the background sweeper auto-denies stale rows),
// and the decision outcome.
//
// Status flow:
//
//	pending -> approved -> executed (or failed)
//	         -> denied   (terminal)
//	         -> expired  (terminal, swept)
//
// Every status transition writes an audit row. The user's decision
// is logged with the actor (`user:<provider>:<email>` when signed in,
// `user:anonymous` otherwise) so we have a complete forensic trace.
package pending

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/storage"
)

// Status is the lifecycle state of a pending action.
type Status string

const (
	// StatusPending is the initial state of a newly queued action.
	StatusPending Status = "pending"
	// StatusApproved means the user has approved the action.
	StatusApproved Status = "approved"
	// StatusDenied means the user has denied the action.
	StatusDenied Status = "denied"
	// StatusExecuted means the action has been executed successfully.
	StatusExecuted Status = "executed"
	// StatusFailed means execution was attempted but failed.
	StatusFailed Status = "failed"
	// StatusExpired means the action's TTL has elapsed.
	StatusExpired Status = "expired"
	// StatusSuperseded means a newer action has replaced this one.
	StatusSuperseded Status = "superseded"
)

// Valid returns true if s is a known status value.
func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusApproved, StatusDenied, StatusExecuted,
		StatusFailed, StatusExpired, StatusSuperseded:
		return true
	}
	return false
}

// Action is the in-memory representation of a pending_actions row.
// The JSON tags double as the SSE payload shape.
type Action struct {
	ID             string     `json:"id"`
	SpawnID        string     `json:"spawn_id"`
	SessionID      string     `json:"session_id"`
	AgentName      string     `json:"agent_name"`
	Kind           string     `json:"kind"`
	Payload        Payload    `json:"payload"`
	GateDecision   string     `json:"gate_decision"`
	GateReason     string     `json:"gate_reason"`
	BlastClass     string     `json:"blast_class"`
	Status         Status     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      time.Time  `json:"expires_at"`
	DecidedAt      *time.Time `json:"decided_at,omitempty"`
	DecidedBy      string     `json:"decided_by,omitempty"`
	DecisionNote   string     `json:"decision_note,omitempty"`
	ExecutedAt     *time.Time `json:"executed_at,omitempty"`
	ExitCode       int        `json:"exit_code"`
	Result         string     `json:"result"`
	ExecutionError string     `json:"execution_error,omitempty"`
	DurationMS     int64      `json:"duration_ms"`
}

// Payload is the request body that the executor consumes. Fields are
// shared with delegation.ActionRequest but kept loose so the queue
// survives payload-shape evolution.
type Payload struct {
	Command string `json:"command,omitempty"`
	Path    string `json:"path,omitempty"`
	Body    string `json:"body,omitempty"`
	Target  string `json:"target,omitempty"` // human-readable target for computeruse.* actions
	Key     string `json:"key,omitempty"`    // for computeruse.key_press
}

// Store is the persistence layer for pending actions.
type Store struct {
	db *storage.DB

	stopCh  chan struct{}
	stopped chan struct{}
}

// New returns a Store backed by the given DB. Does NOT start the
// background sweeper; call Start(ctx) for that.
func New(db *storage.DB) *Store {
	return &Store{
		db:      db,
		stopCh:  make(chan struct{}),
		stopped: make(chan struct{}),
	}
}

// DB returns the underlying storage handle. Exposed so callers
// (e.g. the daemon's e2e tests, the backup subsystem) can run
// ad-hoc SQL against pending_actions without going through the
// typed Store API. Prefer the typed methods when possible.
func (s *Store) DB() *storage.DB { return s.db }

// Start launches the background sweeper that auto-denies stale
// pending actions. Cancel ctx (or call Stop) to terminate it.
func (s *Store) Start(ctx context.Context) {
	go s.sweepLoop(ctx)
}

// Stop terminates the background sweeper.
func (s *Store) Stop() {
	select {
	case <-s.stopCh:
		return
	default:
		close(s.stopCh)
	}
	<-s.stopped
}

// sweepInterval is how often the sweeper checks for expired actions.
const sweepInterval = 30 * time.Second

func (s *Store) sweepLoop(ctx context.Context) {
	defer close(s.stopped)
	t := time.NewTicker(sweepInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-t.C:
			n, err := s.SweepExpired(ctx, time.Now())
			if err != nil {
				// Best-effort: log via stderr since slog isn't imported here.
				fmt.Fprintf(os.Stderr, "pending: sweep: %v\n", err)
				continue
			}
			if n > 0 {
				fmt.Fprintf(os.Stderr, "pending: swept %d expired actions\n", n)
			}
		}
	}
}

// SweepExpired marks every still-pending action whose ExpiresAt is
// in the past as StatusExpired. Returns the number of rows updated.
// Idempotent. Public so tests can drive it deterministically.
func (s *Store) SweepExpired(ctx context.Context, now time.Time) (int64, error) {
	nowISO := now.UTC().Format(time.RFC3339Nano)
	res, err := s.db.SQL().ExecContext(ctx, `
UPDATE pending_actions
SET status = 'expired', decided_at = COALESCE(decided_at, ?)
WHERE status = 'pending' AND expires_at <= ?`,
		nowISO, nowISO,
	)
	if err != nil {
		return 0, fmt.Errorf("pending: sweep expired: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// Insert creates a new pending action in StatusPending. The caller
// supplies the gate verdict + reason so the GUI can render them
// without a follow-up read. The action's ID is generated; the row's
// CreatedAt is set to now; ExpiresAt is now + ttl.
func (s *Store) Insert(ctx context.Context, in InsertInput) (*Action, error) {
	if in.AgentName == "" || in.Kind == "" {
		return nil, errors.New("pending: agent_name and kind are required")
	}
	if in.TTL == 0 {
		in.TTL = DefaultTTL
	}
	id, err := newID()
	if err != nil {
		return nil, fmt.Errorf("pending: new id: %w", err)
	}
	payloadJSON, err := json.Marshal(in.Payload)
	if err != nil {
		return nil, fmt.Errorf("pending: marshal payload: %w", err)
	}
	now := time.Now().UTC()
	row := &Action{
		ID:           id,
		SpawnID:      in.SpawnID,
		SessionID:    in.SessionID,
		AgentName:    in.AgentName,
		Kind:         in.Kind,
		Payload:      in.Payload,
		GateDecision: in.GateDecision,
		GateReason:   in.GateReason,
		BlastClass:   in.BlastClass,
		Status:       StatusPending,
		CreatedAt:    now,
		ExpiresAt:    now.Add(in.TTL),
	}
	_, err = s.db.SQL().ExecContext(ctx, `
INSERT INTO pending_actions (
    id, spawn_id, session_id, agent_name, kind, payload_json,
    gate_decision, gate_reason, blast_class, status,
    created_at, expires_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.SpawnID, row.SessionID, row.AgentName, row.Kind,
		string(payloadJSON), row.GateDecision, row.GateReason,
		row.BlastClass, string(row.Status),
		row.CreatedAt.Format(time.RFC3339Nano),
		row.ExpiresAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return nil, fmt.Errorf("pending: insert: %w", err)
	}
	return row, nil
}

// InsertInput is the constructor parameter for Insert. The store
// fills in CreatedAt/ExpiresAt/ID/Status internally.
type InsertInput struct {
	SpawnID      string
	SessionID    string
	AgentName    string
	Kind         string
	Payload      Payload
	GateDecision string
	GateReason   string
	BlastClass   string
	TTL          time.Duration
}

// DefaultTTL is the time a pending action stays alive before the
// background sweeper marks it StatusExpired. 10 minutes is long
// enough to read the prompt and decide, short enough that an
// unattended queue doesn't pile up.
const DefaultTTL = 10 * time.Minute

// Get returns one row by id. Returns ErrNotFound if the row does
// not exist.
func (s *Store) Get(ctx context.Context, id string) (*Action, error) {
	row := s.db.SQL().QueryRowContext(ctx, `
SELECT id, spawn_id, session_id, agent_name, kind, payload_json,
       gate_decision, gate_reason, blast_class, status,
       created_at, expires_at, decided_at, decided_by, decision_note,
       executed_at, execution_exit_code, execution_result, execution_error, duration_ms
FROM pending_actions WHERE id = ?`, id)
	return scanRow(row)
}

// List returns rows filtered by status. status=="" means all
// statuses. Sorted by CreatedAt DESC (newest first).
func (s *Store) List(ctx context.Context, status Status, limit int) ([]*Action, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var rows *sql.Rows
	var err error
	if status == "" {
		rows, err = s.db.SQL().QueryContext(ctx, `
SELECT id, spawn_id, session_id, agent_name, kind, payload_json,
       gate_decision, gate_reason, blast_class, status,
       created_at, expires_at, decided_at, decided_by, decision_note,
       executed_at, execution_exit_code, execution_result, execution_error, duration_ms
FROM pending_actions
ORDER BY created_at DESC
LIMIT ?`, limit)
	} else {
		rows, err = s.db.SQL().QueryContext(ctx, `
SELECT id, spawn_id, session_id, agent_name, kind, payload_json,
       gate_decision, gate_reason, blast_class, status,
       created_at, expires_at, decided_at, decided_by, decision_note,
       executed_at, execution_exit_code, execution_result, execution_error, duration_ms
FROM pending_actions
WHERE status = ?
ORDER BY created_at DESC
LIMIT ?`, string(status), limit)
	}
	if err != nil {
		return nil, fmt.Errorf("pending: list: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []*Action
	for rows.Next() {
		a, err := scanRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ListPendingBySpawn is a convenience for the GUI: returns the
// pending (still-awaiting-decision) rows for a particular spawn.
// Empty result is normal; nil error is fine.
func (s *Store) ListPendingBySpawn(ctx context.Context, spawnID string) ([]*Action, error) {
	rows, err := s.db.SQL().QueryContext(ctx, `
SELECT id, spawn_id, session_id, agent_name, kind, payload_json,
       gate_decision, gate_reason, blast_class, status,
       created_at, expires_at, decided_at, decided_by, decision_note,
       executed_at, execution_exit_code, execution_result, execution_error, duration_ms
FROM pending_actions
WHERE spawn_id = ? AND status = 'pending'
ORDER BY created_at ASC`, spawnID)
	if err != nil {
		return nil, fmt.Errorf("pending: list by spawn: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []*Action
	for rows.Next() {
		a, err := scanRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// DecisionInput is the parameter for Decide.
type DecisionInput struct {
	ID        string
	Decision  string // "approve" or "deny"
	DecidedBy string // actor identifier for audit ("user:email" or "user:anonymous")
	Note      string
	AutoRun   bool // when true and Decision==approve, mark for execution
}

// ErrNotPending is returned when Decide is called on a row that is
// no longer in StatusPending (already decided, expired, etc).
var ErrNotPending = errors.New("pending: action is no longer pending")

// ErrNotApproved is returned when Execute is called on a row that
// isn't in StatusApproved.
var ErrNotApproved = errors.New("pending: action is not approved")

// ErrNotFound is returned when an id doesn't exist.
var ErrNotFound = errors.New("pending: not found")

// Decide applies a user decision. Atomic at the SQL level (we
// filter on status='pending' so concurrent decisions don't double-
// transition). On success the updated Action is returned.
func (s *Store) Decide(ctx context.Context, in DecisionInput) (*Action, error) {
	if in.Decision != "approve" && in.Decision != "deny" {
		return nil, fmt.Errorf("pending: decision must be approve or deny, got %q", in.Decision)
	}
	if in.DecidedBy == "" {
		in.DecidedBy = "user:anonymous"
	}
	target := StatusDenied
	if in.Decision == "approve" {
		target = StatusApproved
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := s.db.SQL().ExecContext(ctx, `
UPDATE pending_actions
SET status = ?, decided_at = ?, decided_by = ?, decision_note = ?
WHERE id = ? AND status = 'pending'`,
		string(target), now, in.DecidedBy, in.Note, in.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("pending: decide: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		// Either the id is unknown or the row isn't pending.
		// Distinguish so the caller can render a useful error.
		row, gerr := s.Get(ctx, in.ID)
		if gerr != nil {
			return nil, gerr
		}
		return row, ErrNotPending
	}
	return s.Get(ctx, in.ID)
}

// MarkExecuted records the result of an Execute() call. Called by
// the executor (in delegation_wiring.go) after dispatching the
// action. Sets status to StatusExecuted or StatusFailed depending
// on the supplied error, and records exit code, result, error
// message, duration.
func (s *Store) MarkExecuted(ctx context.Context, id string, exitCode int, result string, execErr error, duration time.Duration) error {
	status := StatusExecuted
	errMsg := ""
	if execErr != nil {
		status = StatusFailed
		errMsg = execErr.Error()
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := s.db.SQL().ExecContext(ctx, `
UPDATE pending_actions
SET status = ?, executed_at = ?,
    execution_exit_code = ?, execution_result = ?, execution_error = ?,
    duration_ms = ?
WHERE id = ? AND status = 'approved'`,
		string(status), now, exitCode, result, errMsg, duration.Milliseconds(), id,
	)
	if err != nil {
		return fmt.Errorf("pending: mark executed: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotApproved
	}
	return nil
}

// -----------------------------------------------------------------------------
// Row scanning
// -----------------------------------------------------------------------------

func scanRow(s *sql.Row) (*Action, error) {
	var a Action
	var payloadJSON string
	var createdAt, expiresAt string
	var decidedAt, executedAt sql.NullString
	var decidedBy, decisionNote, executionError sql.NullString
	err := s.Scan(
		&a.ID, &a.SpawnID, &a.SessionID, &a.AgentName, &a.Kind, &payloadJSON,
		&a.GateDecision, &a.GateReason, &a.BlastClass, &a.Status,
		&createdAt, &expiresAt, &decidedAt, &decidedBy, &decisionNote,
		&executedAt, &a.ExitCode, &a.Result, &executionError, &a.DurationMS,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("pending: scan: %w", err)
	}
	if err := hydrateTimestamps(&a, createdAt, expiresAt, decidedAt, executedAt); err != nil {
		return nil, err
	}
	if decidedBy.Valid {
		a.DecidedBy = decidedBy.String
	}
	if decisionNote.Valid {
		a.DecisionNote = decisionNote.String
	}
	if executionError.Valid {
		a.ExecutionError = executionError.String
	}
	if err := json.Unmarshal([]byte(payloadJSON), &a.Payload); err != nil {
		return nil, fmt.Errorf("pending: payload decode: %w", err)
	}
	return &a, nil
}

func scanRows(s *sql.Rows) (*Action, error) {
	var a Action
	var payloadJSON string
	var createdAt, expiresAt string
	var decidedAt, executedAt sql.NullString
	var decidedBy, decisionNote, executionError sql.NullString
	if err := s.Scan(
		&a.ID, &a.SpawnID, &a.SessionID, &a.AgentName, &a.Kind, &payloadJSON,
		&a.GateDecision, &a.GateReason, &a.BlastClass, &a.Status,
		&createdAt, &expiresAt, &decidedAt, &decidedBy, &decisionNote,
		&executedAt, &a.ExitCode, &a.Result, &executionError, &a.DurationMS,
	); err != nil {
		return nil, fmt.Errorf("pending: scan: %w", err)
	}
	if err := hydrateTimestamps(&a, createdAt, expiresAt, decidedAt, executedAt); err != nil {
		return nil, err
	}
	if decidedBy.Valid {
		a.DecidedBy = decidedBy.String
	}
	if decisionNote.Valid {
		a.DecisionNote = decisionNote.String
	}
	if executionError.Valid {
		a.ExecutionError = executionError.String
	}
	if err := json.Unmarshal([]byte(payloadJSON), &a.Payload); err != nil {
		return nil, fmt.Errorf("pending: payload decode: %w", err)
	}
	return &a, nil
}

func hydrateTimestamps(a *Action, createdAt, expiresAt string, decidedAt, executedAt sql.NullString) error {
	t, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return fmt.Errorf("pending: created_at: %w", err)
	}
	a.CreatedAt = t
	t, err = time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return fmt.Errorf("pending: expires_at: %w", err)
	}
	a.ExpiresAt = t
	if decidedAt.Valid {
		t, err := time.Parse(time.RFC3339Nano, decidedAt.String)
		if err == nil {
			a.DecidedAt = &t
		}
	}
	if executedAt.Valid {
		t, err := time.Parse(time.RFC3339Nano, executedAt.String)
		if err == nil {
			a.ExecutedAt = &t
		}
	}
	return nil
}

// newID returns a 128-bit hex ID. Uses crypto/rand so IDs are
// unpredictable (otherwise an attacker could guess pending IDs and
// approve them out from under the GUI).
func newID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
