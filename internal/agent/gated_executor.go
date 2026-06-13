package agent

import (
	"context"
	"fmt"

	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
)

// GatedExecutor wraps any Executor and routes every action through the
// Gatekeeper. This is the structural enforcement of invariant #2 from
// MISSION §2.1: "The Gatekeeper is the only path to physical action."
//
// If the Gatekeeper denies an action, the action is NOT executed. The
// StepResult reflects the denial so the Verifier can decide whether to
// abort the plan. The denial is also written to the audit log so
// forensics can reconstruct what was blocked and why.
type GatedExecutor struct {
	inner Executor
	gate  gatekeeper.Gatekeeper
	audit *audit.Log
}

// NewGatedExecutor wraps inner with the gatekeeper. auditLog may be
// nil; when non-nil, every gatekeeper decision is recorded.
func NewGatedExecutor(inner Executor, gate gatekeeper.Gatekeeper, auditLog *audit.Log) *GatedExecutor {
	return &GatedExecutor{
		inner: inner,
		gate:  gate,
		audit: auditLog,
	}
}

// Execute evaluates the action through the gatekeeper. On Allow, the
// inner executor is invoked. On Deny, the action is blocked and a
// StepResult with Error populated is returned.
func (g *GatedExecutor) Execute(ctx context.Context, action *Action) (*StepResult, error) {
	ba := action.ToBlastRadius()
	decision, reason := g.gate.Evaluate(ctx, ba)
	class := blastradius.Classify(ba)

	g.recordDecision(ctx, action, class, decision, reason)

	if decision != gatekeeper.Allow {
		err := fmt.Errorf("gatekeeper denied %s action: %s", class, reason)
		return &StepResult{
			Success: false,
			Error:   err,
		}, err
	}

	return g.inner.Execute(ctx, action)
}

// recordDecision writes the gatekeeper decision to the audit log when
// one is configured. Audit failures must never block execution, so we
// swallow any error from the audit logger.
func (g *GatedExecutor) recordDecision(ctx context.Context, action *Action, class blastradius.Class, decision gatekeeper.Decision, reason string) {
	if g.audit == nil {
		return
	}
	level := "info"
	result := "allow"
	if decision != gatekeeper.Allow {
		level = "warn"
		result = "deny"
	}
	_ = g.audit.Append(ctx, audit.Event{
		Actor:   "gatekeeper",
		Action:  action.Type,
		App:     action.Target,
		Level:   level,
		Result:  result,
		Message: fmt.Sprintf("%s [%s]: %s", class, decision, reason),
	})
}
