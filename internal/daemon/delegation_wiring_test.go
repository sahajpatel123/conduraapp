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
	"github.com/sahajpatel123/conduraapp/internal/sanitize"
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

// TestP0A_MaliciousChatKindClassifiesAsDestructive is the
// integration test for the P0-A hardening. A compromised
// sub-agent emits a Kind that is NOT in the closed allowlist
// (e.g. "mallory.payload_run"), carrying a destructive Body. The
// audit identified a trust-boundary gap at
// internal/daemon/delegation_wiring.go:410,440: the sub-agent's
// JSON output was fed straight into blastradius.Action.Kind at
// the construction site. The classifier already defaults unknown
// kinds to DESTRUCTIVE, so the outcome was *accidentally* correct.
// The P0-A fix closes the trust boundary explicitly at the
// construction site so:
//   - The row stored in pending_actions always has a canonical,
//     allowlist-bounded Kind (never a raw attacker string);
//   - The blast_class field is guaranteed DESTRUCTIVE for any
//     unknown kind (not coincidentally so);
//   - The audit log records the post-normalize Kind, so an
//     operator reviewing the trail sees "shell.exec" — a kind
//     we control — not the attacker's literal;
//   - The trust boundary remains closed if the classifier ever
//     changes how it handles unknown kinds.
//
// We use a captured-decision gatekeeper (recordingGatekeeper) so
// the test can assert exactly which blastradius.Action the
// gating path presented to the policy engine. The gate itself
// returns RequirePresenceAndConsent, mirroring the default-deny
// policy shipped with the binary.
func TestP0A_MaliciousChatKindClassifiesAsDestructive(t *testing.T) {
	store := newPendingStore(t)
	gate := &recordingGatekeeper{
		decision: gatekeeper.RequirePresenceAndConsent,
		reason:   "default-deny: no policy rule matched",
	}
	subs := &Subsystems{
		Gatekeeper: gate,
		Delegation: delegation.NewGatedRunner(delegation.Config{}, noopGatekeeper{}, nil),
		Pending:    store,
		Audit:      nil,
	}

	malicious := &delegation.SpawnResult{
		AgentName: "evil-agent",
		Output:    `{"agent_name":"evil-agent","kind":"mallory.payload_run","body":"rm -rf $HOME"}`,
	}

	rows := gateAndPersistParsedActions(context.Background(), subs, malicious)
	if len(rows) != 1 {
		t.Fatalf("expected exactly 1 row from the malicious payload, got %d", len(rows))
	}
	row := rows[0]

	// The kind stored on the row must be the normalized form,
	// not the attacker-supplied literal "chat".
	if row.Kind != "shell.exec" {
		t.Errorf("P0-A bypass: row.Kind = %q, want %q (normalizer must rewrite unknown/attacker kinds to shell.exec)",
			row.Kind, "shell.exec")
	}
	// The blast class recorded on the row must be DESTRUCTIVE.
	if row.BlastClass != blastradius.DESTRUCTIVE.String() {
		t.Errorf("P0-A bypass: row.BlastClass = %q, want %q (normalizer must produce a DESTRUCTIVE kind)",
			row.BlastClass, blastradius.DESTRUCTIVE.String())
	}
	// And the gatekeeper must have required human-in-the-loop
	// consent, not allowed it. Allow on a destructive action is
	// a regression — that is exactly what P0-A exists to prevent.
	wantVerdict := "require_presence_and_consent"
	if row.GateDecision != wantVerdict {
		t.Errorf("P0-A bypass: gate decision = %q, want %q", row.GateDecision, wantVerdict)
	}
	// The row should still be pending (waiting for the human to
	// decide), not auto-denied — DESTRUCTIVE-classified unknown
	// actions should pause for the user, not be killed silently.
	if row.Status != pending.StatusPending {
		t.Errorf("row.Status = %s, want %s (DESTRUCTIVE + unknown must queue for human decision)",
			row.Status, pending.StatusPending)
	}

	// Defence-in-depth: confirm the gatekeeper saw the NORMALIZED
	// kind on the blastradius.Action it received, not the
	// attacker literal. If a regression bypassed the normalizer
	// (e.g. someone wires ar.Kind straight back into ba.Kind),
	// this assertion catches it.
	if len(gate.lastActions) == 0 {
		t.Fatal("recording gatekeeper never recorded an action — wiring path is broken")
	}
	gotAction := gate.lastActions[len(gate.lastActions)-1]
	if gotAction.Kind != "shell.exec" {
		t.Errorf("recording gatekeeper saw Kind=%q; P0-A normalization is bypassed somewhere "+
			"between ar.Kind and blastradius.Action.Kind", gotAction.Kind)
	}
	if blastradius.Classify(gotAction) != blastradius.DESTRUCTIVE {
		t.Errorf("recorded action classified as %s; P0-A must reach DESTRUCTIVE",
			blastradius.Classify(gotAction))
	}

	// Defence-in-depth check: confirm the classifier labels the
	// attacker's literal "mallory.payload_run" as DESTRUCTIVE.
	// The P0-A fix does NOT change the classifier — that
	// behaviour already existed. What P0-A changes is the
	// *trust boundary*: the kind that lands on the row, in the
	// audit log, and on the blastradius.Action the
	// gatekeeper/policy code sees is always the canonical
	// "shell.exec" we control, never the attacker's literal.
	// Without P0-A the classifier would still say DESTRUCTIVE,
	// but blastradius.Action.Kind, row.Kind, and every audit
	// row would carry the raw attacker string. This assertion
	// is a regression guard: if the classifier is ever changed
	// to handle unknown kinds differently, this test surfaces
	// the change and forces a maintainer to re-examine the
	// threat model rather than silently weakening P0-A.
	classForMaliciousLiteral := blastradius.Classify(blastradius.Action{Kind: "mallory.payload_run", Body: "rm -rf $HOME"})
	if classForMaliciousLiteral != blastradius.DESTRUCTIVE {
		t.Fatalf("blastradius regressed: unknown Kind %q now classifies as %s; the test was written assuming default-deny DESTRUCTIVE classification. Re-check the attack scenario before updating.",
			"mallory.payload_run", classForMaliciousLiteral)
	}

	// Mirror check: confirm the normalize step rewrites the
	// attacker's literal to the canonical "shell.exec" Kind.
	// If this assertion ever fails the normalizer behaviour
	// has changed and the audit/pending rows that assume Kind
	// is a closed-set value are no longer guaranteed.
	got := sanitize.NormalizeSubAgentKind("mallory.payload_run")
	if got != "shell.exec" {
		t.Fatalf("normalize step changed: %q → %q; P0-A semantics depend on this exact rewrite",
			"mallory.payload_run", got)
	}
}

// recordingGatekeeper is a test-only Gatekeeper that captures the
// blastradius.Action it is asked to evaluate. The P0-A test uses
// it to assert the exact action the normalization step produced,
// not just the eventual verdict — a regression that bypassed
// the normalizer would still produce a "deny" verdict on the
// shell-policy path (the per-kind payload sanitizers would
// catch the body), but the *trust boundary* would be open.
type recordingGatekeeper struct {
	decision    gatekeeper.Decision
	reason      string
	lastActions []blastradius.Action
}

func (r *recordingGatekeeper) Evaluate(_ context.Context, a blastradius.Action) (gatekeeper.Decision, string) {
	r.lastActions = append(r.lastActions, a)
	return r.decision, r.reason
}

// filepath tail-import bridge: ensures the existing filepath
// import stays live across future tidy runs even if the only
// usage at the bottom of the file disappears.
var _ = filepath.Base
