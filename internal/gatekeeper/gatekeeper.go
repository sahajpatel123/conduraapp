// Package gatekeeper is the only path from a model's intent to a physical
// action. It is deterministic code, never a model (MISSION §2.1 invariant
// 1, §5.7): the Strategist decides WHAT to do, the Gatekeeper decides
// WHETHER it is safe to do.
//
// This package ships the v0 implementation, DenyBeyondRead, which allows
// READ actions and denies everything else. It is the safety seam for
// Phase 4's "Living Presence": the agent gains voice and presence now,
// but cannot click, type, write, or exec, because the only code path to
// those actions runs through Evaluate, and Evaluate denies them until the
// real rules engine (policy.yaml, consent dialogs, queueing — MISSION
// §10.2) replaces this one behind the same interface in Phase 5.
package gatekeeper

import (
	"context"
	"fmt"

	"github.com/sahajpatel123/synapticapp/internal/blastradius"
)

// Decision is the Gatekeeper's verdict on a proposed action. v0 uses only
// Allow and Deny; the real engine will add RequireConsent and Queue.
type Decision int

const (
	// Allow permits the action to proceed.
	Allow Decision = iota
	// Deny blocks the action. The reason returned alongside it explains
	// why, for the audit log and the user-facing message.
	Deny
)

// String renders the decision for logs and audit entries.
func (d Decision) String() string {
	switch d {
	case Allow:
		return "Allow"
	case Deny:
		return "Deny"
	default:
		return "Deny"
	}
}

// Gatekeeper evaluates a proposed action and returns a Decision plus a
// human-readable reason. Implementations MUST be deterministic.
type Gatekeeper interface {
	Evaluate(ctx context.Context, a blastradius.Action) (Decision, string)
}

// DenyBeyondRead is the Phase 4 Gatekeeper: it allows READ actions and
// denies every higher class. It carries no state.
type DenyBeyondRead struct{}

// NewDenyBeyondRead returns the v0 deny-beyond-read Gatekeeper.
func NewDenyBeyondRead() DenyBeyondRead {
	return DenyBeyondRead{}
}

// Evaluate allows READ actions and denies all others, naming the blocked
// class and explaining that the real safety layer is not yet built.
func (DenyBeyondRead) Evaluate(_ context.Context, a blastradius.Action) (Decision, string) {
	class := blastradius.Classify(a)
	if class == blastradius.READ {
		return Allow, "READ action permitted"
	}
	return Deny, fmt.Sprintf(
		"%s action blocked: the deterministic safety layer is not yet "+
			"implemented (Phase 5); only READ actions are permitted",
		class,
	)
}
