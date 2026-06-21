package daemon

import (
	"context"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/delegation"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
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

// TestGateAndAuditParsedActions_ReadsRequestsAndGates pins the
// Phase 17 Fix #7 (B5) behavior:
//   - ActionRequests with non-empty Kind are gated.
//   - Each decision is mapped to the right tag (allow / deny / etc).
//   - The returned slice mirrors request order.
//   - A request with empty Kind is skipped (defensive).
func TestGateAndAuditParsedActions_ReadsRequestsAndGates(t *testing.T) {
	cases := []struct {
		name         string
		gate         gatekeeper.Gatekeeper
		wantDec      string
		wantAllowed  bool
	}{
		{"allow", noopGatekeeper{}, "allow", true},
		{"deny", fakeGatekeeper{decision: gatekeeper.Deny, reason: "test deny"}, "deny", false},
		{"require_consent", fakeGatekeeper{decision: gatekeeper.RequireConsent, reason: "needs user"}, "require_consent", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			subs := &Subsystems{
				Gatekeeper: tc.gate,
				Delegation: nil, // not used; tests below will construct a real one
				Audit:      nil, // audit is best-effort; nil skips audit silently
			}
			// We can't directly call gateAndAuditParsedActions with
			// nil Delegation because ActionRequests lives on the
			// runner. Construct a minimal GatedRunner via the
			// factory and use it to parse our fixture output.
			runner := delegation.NewGatedRunner(delegation.Config{}, fakeGatekeeper{gatekeeper.Allow, "x"}, nil)
			subs.Delegation = runner

			result := &delegation.SpawnResult{
				AgentName: "test-agent",
				Output: `some preamble
{"agent_name":"test-agent","kind":"shell.exec","command":"rm -rf /"}
trailing text
{"agent_name":"test-agent","kind":"","command":""}
`,
			}
			decisions := gateAndAuditParsedActions(context.Background(), subs, result)
			if len(decisions) != 1 {
				t.Fatalf("expected 1 decision (empty Kind skipped), got %d", len(decisions))
			}
			if decisions[0].Kind != "shell.exec" {
				t.Errorf("expected kind shell.exec, got %q", decisions[0].Kind)
			}
			if decisions[0].Decision != tc.wantDec {
				t.Errorf("expected decision %q, got %q", tc.wantDec, decisions[0].Decision)
			}
			if decisions[0].Allowed != tc.wantAllowed {
				t.Errorf("expected allowed=%v, got %v", tc.wantAllowed, decisions[0].Allowed)
			}
		})
	}
}

// TestGateAndAuditParsedActions_NoRequestsReturnsEmpty ensures
// empty/missing sub-agent requests produce a nil result, not an
// empty slice. The GUI renders "no sub-agent actions" when the
// field is absent.
func TestGateAndAuditParsedActions_NoRequestsReturnsEmpty(t *testing.T) {
	subs := &Subsystems{
		Gatekeeper: noopGatekeeper{},
		Delegation: delegation.NewGatedRunner(delegation.Config{}, noopGatekeeper{}, nil),
	}
	result := &delegation.SpawnResult{
		AgentName: "test-agent",
		Output:    "no JSON here at all",
	}
	decisions := gateAndAuditParsedActions(context.Background(), subs, result)
	if decisions != nil {
		t.Errorf("expected nil for no requests, got %v", decisions)
	}
}
