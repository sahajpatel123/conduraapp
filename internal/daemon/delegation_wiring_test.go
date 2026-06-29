package daemon

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/blastradius"
	"github.com/sahajpatel123/conduraapp/internal/delegation"
	"github.com/sahajpatel123/conduraapp/internal/gatekeeper"
	"github.com/sahajpatel123/conduraapp/internal/pending"
	"github.com/sahajpatel123/conduraapp/internal/storage"
)

// fakeGatekeeper is a deterministic gatekeeper for tests. Returns
// the configured decision + reason for every action.
type fakeGatekeeper struct {
	decision gatekeeper.Decision
	reason   string
}

func (f fakeGatekeeper) Evaluate(_ context.Context, _ blastradius.Action) (gatekeeper.Decision, string) {
	return f.decision, f.reason
}

// noopGatekeeper always allows. Used by the success-path test.
type noopGatekeeper struct{}

func (noopGatekeeper) Evaluate(_ context.Context, _ blastradius.Action) (gatekeeper.Decision, string) {
	return gatekeeper.Allow, "noop test"
}

// TestGateAndPersistParsedActions_ReadsRequestsAndGates pins the
// Phase 17 Fix #7 (B5) + Phase 18 (v0.2.0) behavior:
//   - ActionRequests with non-empty Kind are gated.
//   - Each decision is persisted to the pending_actions queue.
//   - Rows the gate denied get auto-denied in the row, not left
//     pending.
//   - The returned slice mirrors request order.
//   - A request with empty Kind is skipped (defensive).
func TestGateAndPersistParsedActions_ReadsRequestsAndGates(t *testing.T) {
	cases := []struct {
		name            string
		gate            gatekeeper.Gatekeeper
		wantDec         string
		wantFinalStatus pending.Status
	}{
		{"allow", noopGatekeeper{}, "allow", pending.StatusPending},
		{"deny", fakeGatekeeper{decision: gatekeeper.Deny, reason: "test deny"}, "deny", pending.StatusDenied},
		{"require_consent", fakeGatekeeper{decision: gatekeeper.RequireConsent, reason: "needs user"}, "require_consent", pending.StatusPending},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := newPendingStore(t)
			subs := &Subsystems{
				Gatekeeper: tc.gate,
				Delegation: delegation.NewGatedRunner(delegation.Config{}, fakeGatekeeper{gatekeeper.Allow, "x"}, nil),
				Pending:    store,
				Audit:      nil, // audit is best-effort; nil skips audit silently
			}

			result := &delegation.SpawnResult{
				AgentName: "test-agent",
				Output: `some preamble
{"agent_name":"test-agent","kind":"shell.exec","command":"rm -rf /"}
trailing text
{"agent_name":"test-agent","kind":"","command":""}
`,
			}
			rows := gateAndPersistParsedActions(context.Background(), subs, result)
			if len(rows) != 1 {
				t.Fatalf("expected 1 row (empty Kind skipped), got %d", len(rows))
			}
			if rows[0].Kind != "shell.exec" {
				t.Errorf("expected kind shell.exec, got %q", rows[0].Kind)
			}
			if rows[0].GateDecision != tc.wantDec {
				t.Errorf("expected gate_decision %q, got %q", tc.wantDec, rows[0].GateDecision)
			}
			if rows[0].Status != tc.wantFinalStatus {
				t.Errorf("expected status %s (denied rows auto-denied), got %s", tc.wantFinalStatus, rows[0].Status)
			}
		})
	}
}

// TestGateAndPersistParsedActions_NoRequestsReturnsEmpty ensures
// empty/missing sub-agent requests produce a nil result, not an
// empty slice. The GUI renders "no sub-agent actions" when the
// field is absent.
func TestGateAndPersistParsedActions_NoRequestsReturnsEmpty(t *testing.T) {
	store := newPendingStore(t)
	subs := &Subsystems{
		Gatekeeper: noopGatekeeper{},
		Delegation: delegation.NewGatedRunner(delegation.Config{}, noopGatekeeper{}, nil),
		Pending:    store,
	}
	result := &delegation.SpawnResult{
		AgentName: "test-agent",
		Output:    "no JSON here at all",
	}
	rows := gateAndPersistParsedActions(context.Background(), subs, result)
	if rows != nil {
		t.Errorf("expected nil for no requests, got %v", rows)
	}
}

// newPendingStore builds a fresh pending.Store backed by an
// ephemeral SQLite DB. Used by the wiring tests in this file.
func newPendingStore(t *testing.T) *pending.Store {
	t.Helper()
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{
		Path: filepath.Join(dir, "synaptic.db"),
	})
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return pending.New(db)
}

// Compile-time guard: ensure the tests below import strings.
var _ = strings.TrimSpace
