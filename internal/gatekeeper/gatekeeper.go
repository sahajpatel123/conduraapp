// Package gatekeeper is the only path from a model's intent to a physical
// action. It is deterministic code, never a model (MISSION S2.1 invariant 1).
//
// Two layers per the authorization contract:
//  1. Policy.Evaluate(action) -> Verdict — pure, no I/O, unit-testable.
//  2. Engine.Evaluate(ctx, action) -> (Decision, reason) — terminal interface,
//     drives consent provider, blocks on ctx+halt, collapses to Allow/Deny.
//
// DenyBeyondRead is retained for test backward compatibility.
package gatekeeper

import (
	"context"

	"github.com/sahajpatel123/synapticapp/internal/blastradius"
)

// Decision is the terminal verdict. Engine always returns Allow or Deny.
// RequireConsent and RequirePresenceAndConsent are internal Policy verdicts;
// they never cross the gatekeeper.Gatekeeper interface boundary.
type Decision int

const (
	// Allow permits the action to proceed.
	Allow Decision = iota
	// Deny blocks the action with a reason.
	Deny
	// RequireConsent is an internal verdict requiring user consent.
	RequireConsent
	// RequirePresenceAndConsent requires user presence and consent.
	RequirePresenceAndConsent
)

func (d Decision) String() string {
	switch d {
	case Allow:
		return "Allow"
	case Deny:
		return "Deny"
	case RequireConsent:
		return "RequireConsent"
	case RequirePresenceAndConsent:
		return "RequirePresenceAndConsent"
	default:
		return "Deny"
	}
}

// Gatekeeper evaluates a proposed action. Implementations MUST be deterministic.
type Gatekeeper interface {
	Evaluate(ctx context.Context, a blastradius.Action) (Decision, string)
}

// DenyBeyondRead is the Phase 4 stub — v0 safety seam. Retained for
// test backward compatibility. Production uses Engine.
type DenyBeyondRead struct{}

// NewDenyBeyondRead returns the v0 safety stub.
func NewDenyBeyondRead() DenyBeyondRead { return DenyBeyondRead{} }

// Evaluate allows READ actions and denies everything else.
func (DenyBeyondRead) Evaluate(_ context.Context, a blastradius.Action) (Decision, string) {
	class := blastradius.Classify(a)
	if class == blastradius.READ {
		return Allow, "READ action permitted"
	}
	classStr := class.String()
	return Deny, classStr + " action blocked by v0 safety stub"
}
