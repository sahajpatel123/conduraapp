package pending

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/storage"
)

// newTestStorage creates an in-memory-backed storage.DB for tests.
func newTestStorage(t *testing.T) *storage.DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "synaptic.db")
	db, err := storage.Open(context.Background(), storage.Config{
		Path:      dbPath,
		MasterKey: "",
		Secrets:   nil,
	})
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestStore_InsertAndGet covers the basic round-trip: insert returns
// an ID, Get returns the same row.
func TestStore_InsertAndGet(t *testing.T) {
	db := newTestStorage(t)
	s := New(db)
	ctx := context.Background()

	in := InsertInput{
		SpawnID:      "spawn-123",
		AgentName:    "claude",
		Kind:         "shell.exec",
		Payload:      Payload{Command: "echo hello"},
		GateDecision: "allow",
		GateReason:   "trust:workspace",
		BlastClass:   "WRITE",
		TTL:          5 * time.Minute,
	}
	got, err := s.Insert(ctx, in)
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if got.ID == "" {
		t.Error("ID should not be empty")
	}
	if got.Status != StatusPending {
		t.Errorf("status: got %q, want pending", got.Status)
	}
	if got.ExpiresAt.Sub(got.CreatedAt) != 5*time.Minute {
		t.Errorf("ttl: got %v, want 5m", got.ExpiresAt.Sub(got.CreatedAt))
	}

	// Round-trip.
	back, err := s.Get(ctx, got.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if back.AgentName != "claude" {
		t.Errorf("agent_name round-trip: got %q", back.AgentName)
	}
	if back.Payload.Command != "echo hello" {
		t.Errorf("payload round-trip: got %q", back.Payload.Command)
	}
}

// TestStore_Insert_RequiresFields pins the constructor's required
// fields so callers can't sneak empty kinds through.
func TestStore_Insert_RequiresFields(t *testing.T) {
	db := newTestStorage(t)
	s := New(db)
	cases := []struct {
		name string
		in   InsertInput
	}{
		{"empty agent", InsertInput{Kind: "shell.exec"}},
		{"empty kind", InsertInput{AgentName: "claude"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := s.Insert(context.Background(), tc.in); err == nil {
				t.Error("expected error for missing field")
			}
		})
	}
}

// TestStore_Decide_ApproveAndDeny exercises the two-terminal decision
// path. After Decide, status is one of approved/denied; further
// Decides return ErrNotPending.
func TestStore_Decide_ApproveAndDeny(t *testing.T) {
	db := newTestStorage(t)
	s := New(db)
	ctx := context.Background()

	in := InsertInput{
		SpawnID: "spawn-1", AgentName: "claude", Kind: "shell.exec",
		Payload: Payload{Command: "ls"},
	}
	a, err := s.Insert(ctx, in)
	if err != nil {
		t.Fatal(err)
	}

	approved, err := s.Decide(ctx, DecisionInput{ID: a.ID, Decision: "approve", DecidedBy: "user:alice"})
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if approved.Status != StatusApproved {
		t.Errorf("status: got %q, want approved", approved.Status)
	}
	if approved.DecidedBy != "user:alice" {
		t.Errorf("decided_by: got %q", approved.DecidedBy)
	}

	// Second decide must fail with ErrNotPending.
	_, err = s.Decide(ctx, DecisionInput{ID: a.ID, Decision: "deny"})
	if err == nil || !strings.Contains(err.Error(), "no longer pending") {
		t.Errorf("expected ErrNotPending, got %v", err)
	}

	// A fresh row can be denied.
	b, err := s.Insert(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	denied, err := s.Decide(ctx, DecisionInput{ID: b.ID, Decision: "deny", DecidedBy: "user:bob"})
	if err != nil {
		t.Fatalf("deny: %v", err)
	}
	if denied.Status != StatusDenied {
		t.Errorf("status: got %q, want denied", denied.Status)
	}
}

// TestStore_Decide_BadDecision pins the input validation on Decide.
func TestStore_Decide_BadDecision(t *testing.T) {
	db := newTestStorage(t)
	s := New(db)
	a, _ := s.Insert(context.Background(), InsertInput{
		SpawnID: "spawn", AgentName: "a", Kind: "k",
	})
	_, err := s.Decide(context.Background(), DecisionInput{ID: a.ID, Decision: "maybe"})
	if err == nil || !strings.Contains(err.Error(), "must be approve or deny") {
		t.Errorf("expected validation error, got %v", err)
	}
}

// TestStore_Decide_UnknownID pins the lookup-miss path.
func TestStore_Decide_UnknownID(t *testing.T) {
	db := newTestStorage(t)
	s := New(db)
	_, err := s.Decide(context.Background(), DecisionInput{ID: "nope", Decision: "approve"})
	if err == nil {
		t.Error("expected error for unknown id")
	}
}

// TestStore_MarkExecuted covers the success and failure paths of
// the executor's completion callback.
func TestStore_MarkExecuted(t *testing.T) {
	db := newTestStorage(t)
	s := New(db)
	ctx := context.Background()

	a, _ := s.Insert(ctx, InsertInput{SpawnID: "sp", AgentName: "a", Kind: "k"})
	_, _ = s.Decide(ctx, DecisionInput{ID: a.ID, Decision: "approve"})

	// Success path.
	if err := s.MarkExecuted(ctx, a.ID, 0, "ok", nil, 100*time.Millisecond); err != nil {
		t.Fatalf("mark executed: %v", err)
	}
	row, _ := s.Get(ctx, a.ID)
	if row.Status != StatusExecuted {
		t.Errorf("status: got %q, want executed", row.Status)
	}
	if row.ExitCode != 0 {
		t.Errorf("exit_code: got %d, want 0", row.ExitCode)
	}
	if row.Result != "ok" {
		t.Errorf("result: got %q", row.Result)
	}
	if row.DurationMS != 100 {
		t.Errorf("duration_ms: got %d, want 100", row.DurationMS)
	}
	if row.ExecutedAt == nil {
		t.Error("executed_at should be set")
	}

	// Failure path on a second action.
	b, _ := s.Insert(ctx, InsertInput{SpawnID: "sp", AgentName: "a", Kind: "k"})
	_, _ = s.Decide(ctx, DecisionInput{ID: b.ID, Decision: "approve"})
	if err := s.MarkExecuted(ctx, b.ID, 2, "", errFake("command failed"), 50*time.Millisecond); err != nil {
		t.Fatalf("mark failed: %v", err)
	}
	row2, _ := s.Get(ctx, b.ID)
	if row2.Status != StatusFailed {
		t.Errorf("status: got %q, want failed", row2.Status)
	}
	if row2.ExitCode != 2 {
		t.Errorf("exit_code: got %d", row2.ExitCode)
	}
	if row2.ExecutionError != "command failed" {
		t.Errorf("error: got %q", row2.ExecutionError)
	}

	// Marking a non-approved action must fail.
	c, _ := s.Insert(ctx, InsertInput{SpawnID: "sp", AgentName: "a", Kind: "k"})
	if err := s.MarkExecuted(ctx, c.ID, 0, "", nil, 0); err == nil {
		t.Error("expected ErrNotApproved for pending action")
	}
}

// errFake is a tiny error type used in MarkExecuted failure-path tests.
type errFake string

func (e errFake) Error() string { return string(e) }

// TestStore_SweepExpired pins the auto-deny behavior: pending actions
// past their ExpiresAt get marked StatusExpired, terminal states do not.
func TestStore_SweepExpired(t *testing.T) {
	db := newTestStorage(t)
	s := New(db)
	ctx := context.Background()

	// Insert with a TTL in the past.
	past := time.Now().Add(-1 * time.Minute)
	future := time.Now().Add(10 * time.Minute)
	a, _ := s.Insert(ctx, InsertInput{SpawnID: "sp", AgentName: "a", Kind: "k", TTL: 5 * time.Minute})
	// Manually rewind CreatedAt + ExpiresAt to the past.
	_, err := s.db.SQL().ExecContext(ctx, `
UPDATE pending_actions SET created_at = ?, expires_at = ? WHERE id = ?`,
		past.UTC().Format(time.RFC3339Nano),
		past.UTC().Format(time.RFC3339Nano),
		a.ID,
	)
	if err != nil {
		t.Fatal(err)
	}
	_ = future

	n, err := s.SweepExpired(ctx, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("swept: got %d, want 1", n)
	}
	row, _ := s.Get(ctx, a.ID)
	if row.Status != StatusExpired {
		t.Errorf("status: got %q, want expired", row.Status)
	}

	// Idempotent: second sweep should not count again.
	n, _ = s.SweepExpired(ctx, time.Now())
	if n != 0 {
		t.Errorf("idempotent sweep: got %d, want 0", n)
	}
}

// TestStore_List_FilterByStatus ensures List honors the status filter.
func TestStore_List_FilterByStatus(t *testing.T) {
	db := newTestStorage(t)
	s := New(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := s.Insert(ctx, InsertInput{SpawnID: "sp", AgentName: "a", Kind: "k"})
		if err != nil {
			t.Fatal(err)
		}
	}
	pending, err := s.List(ctx, StatusPending, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 3 {
		t.Errorf("pending count: got %d, want 3", len(pending))
	}
	denied, _ := s.List(ctx, StatusDenied, 10)
	if len(denied) != 0 {
		t.Errorf("denied count: got %d, want 0", len(denied))
	}
}

// TestStore_ListPendingBySpawn is the GUI's primary lookup. Returns
// the still-pending rows for a spawn in created-asc order.
func TestStore_ListPendingBySpawn(t *testing.T) {
	db := newTestStorage(t)
	s := New(db)
	ctx := context.Background()

	a, _ := s.Insert(ctx, InsertInput{SpawnID: "sp-1", AgentName: "a", Kind: "k"})
	b, _ := s.Insert(ctx, InsertInput{SpawnID: "sp-2", AgentName: "a", Kind: "k"})
	_, _ = s.Decide(ctx, DecisionInput{ID: b.ID, Decision: "deny"})

	rows, err := s.ListPendingBySpawn(ctx, "sp-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Errorf("sp-1 pending: got %d, want 1", len(rows))
	}
	if rows[0].ID != a.ID {
		t.Errorf("sp-1 row ID mismatch")
	}

	rows, _ = s.ListPendingBySpawn(ctx, "sp-2")
	if len(rows) != 0 {
		t.Errorf("sp-2 pending: got %d, want 0 (denied)", len(rows))
	}
}

// TestStore_Get_Unknown pins the ErrNotFound contract.
func TestStore_Get_Unknown(t *testing.T) {
	db := newTestStorage(t)
	s := New(db)
	_, err := s.Get(context.Background(), "nonexistent")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// TestStore_NewID_Unique stresses ID generation to make sure
// crypto/rand isn't producing collisions.
func TestStore_NewID_Unique(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		id, err := newID()
		if err != nil {
			t.Fatal(err)
		}
		if seen[id] {
			t.Errorf("duplicate ID: %s", id)
		}
		seen[id] = true
		if len(id) != 32 {
			t.Errorf("id length: got %d, want 32", len(id))
		}
	}
}
