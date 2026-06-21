package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/delegation"
	"github.com/sahajpatel123/synapticapp/internal/failover"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
)

func buildDelegationBus(engine gatekeeper.Gatekeeper, sp *failover.SpendMonitor) *delegation.GatedRunner {
	cfg := delegation.DefaultConfig()
	limiter := delegation.NewLimiter(cfg, sp)
	runner := delegation.NewGatedRunner(cfg, engine, limiter)
	// MISSION S13.4: per-agent limit 4, global limit from config.
	runner.SetSemaphoreManager(delegation.NewSemaphoreManager(0, cfg.GlobalLimit))
	return runner
}

// Error codes for delegation RPC responses.
const (
	codeGatekeeperDeny = -32001
	codeCancelled      = -32002
)

func mapSpawnError(err error) *ipc.Error {
	switch {
	case errors.Is(err, delegation.ErrAgentNotFound), errors.Is(err, delegation.ErrRecursionLimit), errors.Is(err, delegation.ErrBudgetExceeded):
		return &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
	case errors.Is(err, delegation.ErrGatedDeny):
		return &ipc.Error{Code: codeGatekeeperDeny, Message: err.Error()}
	case errors.Is(err, context.Canceled):
		return &ipc.Error{Code: codeCancelled, Message: err.Error()}
	default:
		return &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
	}
}

// registerDelegationMethods registers delegation RPC methods.
func registerDelegationMethods(srv *ipc.Server, subs *Subsystems) {
	if subs.Delegation == nil {
		return
	}

	srv.Register("delegate.spawn", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			AgentName string  `json:"agent_name"`
			Task      string  `json:"task"`
			Model     string  `json:"model,omitempty"`
			Depth     int     `json:"depth"`
			Budget    float64 `json:"budget"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p.AgentName == "" || p.Task == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "agent_name and task are required"}
		}
		req := &delegation.SpawnRequest{
			AgentName: p.AgentName,
			Task:      p.Task,
			Model:     p.Model,
			Depth:     p.Depth,
			Budget:    p.Budget,
		}
		result, err := subs.Delegation.Spawn(ctx, req)
		if err != nil {
			return nil, mapSpawnError(err)
		}
		// Phase 17, Fix #7 (B5): parse structured ActionRequests
		// from the sub-agent's output and gate each one. Without
		// this step the sub-agent can return "run this shell
		// command" or "type into this app" and the daemon would
		// silently trust it. Now each request goes through the
		// same Gatekeeper the user-installed CLI itself passed
		// to Spawn — same rules, same audit trail, same consent.
		//
		// v0.1.0 scope: we parse, gate, and audit. The actual
		// physical execution of approved requests is a v0.2.0
		// follow-on (the GUI surfaces them as a "pending
		// sub-agent requests" list; user approves each one
		// explicitly).
		actionDecisions := gateAndAuditParsedActions(ctx, subs, result)
		return map[string]any{
			"agent_name":       result.AgentName,
			"output":           result.Output,
			"exit_code":        result.ExitCode,
			"duration_ms":      result.Duration.Milliseconds(),
			"token_count":      result.TokenCount,
			"spawn_id":         result.SpawnID,
			"action_decisions": actionDecisions,
		}, nil
	})

	srv.Register("delegate.list_agents", func(_ context.Context, _ json.RawMessage) (any, error) {
		cfg := subs.Delegation.Config()
		agents := make([]map[string]any, len(cfg.Agents))
		for i := range cfg.Agents { //nolint:gocritic
			a := cfg.Agents[i]
			agents[i] = map[string]any{
				"name":        a.Name,
				"description": a.Description,
				"binary":      a.BinaryProbe,
			}
		}
		return map[string]any{"agents": agents}, nil
	})

	srv.Register("delegate.cancel", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			SpawnID string `json:"spawn_id"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p.SpawnID == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "spawn_id is required"}
		}
		if !subs.Delegation.Cancel(p.SpawnID) {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "unknown or already finished spawn_id"}
		}
		return auditOK(), nil
	})
}

// gatedActionDecision is the per-action gate verdict exposed in the
// delegate.spawn response. Kept as a flat struct so the GUI can show
// "sub-agent asked to <kind>; gate said <decision> because <reason>".
type gatedActionDecision struct {
	Kind      string `json:"kind"`
	AgentName string `json:"agent_name"`
	Decision  string `json:"decision"`
	Reason    string `json:"reason"`
	Allowed   bool   `json:"allowed"`
}

// gateAndAuditParsedActions walks every ActionRequest the sub-agent
// emitted and runs each through the Gatekeeper. Each decision is
// audited (actor=sub-agent, app=condurad). Returns the slice in
// request order so the caller can render it 1:1.
//
// v0.1.0 does NOT execute the approved actions — that requires the
// GUI to surface them as a confirm-then-run queue. The gate verdict
// is the contract for v0.2.0's executor.
func gateAndAuditParsedActions(ctx context.Context, subs *Subsystems, result *delegation.SpawnResult) []gatedActionDecision {
	if subs == nil || subs.Delegation == nil || subs.Gatekeeper == nil || result == nil {
		return nil
	}
	requests := subs.Delegation.ActionRequests(result)
	if len(requests) == 0 {
		return nil
	}
	out := make([]gatedActionDecision, 0, len(requests))
	for _, ar := range requests {
		ba := blastradius.Action{
			Kind:      ar.Kind,
			TargetApp: ar.AgentName,
			Body:      ar.Body,
			Path:      ar.Path,
			Command:   ar.Command,
		}
		decision, reason := subs.Gatekeeper.Evaluate(ctx, ba)
		// Map gatekeeper.Decision to a friendly string. We use a
		// short tag here so the GUI can switch on it.
		var decStr string
		var allowed bool
		switch decision {
		case gatekeeper.Allow:
			decStr, allowed = "allow", true
		case gatekeeper.Deny:
			decStr, allowed = "deny", false
		case gatekeeper.RequireConsent:
			decStr, allowed = "require_consent", false
		case gatekeeper.RequirePresenceAndConsent:
			decStr, allowed = "require_presence_and_consent", false
		default:
			decStr, allowed = "unknown", false
		}
		out = append(out, gatedActionDecision{
			Kind:      ar.Kind,
			AgentName: ar.AgentName,
			Decision:  decStr,
			Reason:    reason,
			Allowed:   allowed,
		})
		// Audit row: this is the FORENSIC trace that a sub-agent
		// asked to do X and we said Y. Even for allowed decisions
		// we log so a future audit can replay what the agent
		// would have done.
		level := auditLevelInfo
		switch decision {
		case gatekeeper.Deny:
			level = auditLevelWarn
		case gatekeeper.RequireConsent, gatekeeper.RequirePresenceAndConsent:
			level = auditLevelWarn
		}
		if subs.Audit != nil {
			if err := subs.Audit.Append(ctx, audit.Event{
				Actor:   "sub-agent:" + ar.AgentName,
				Action:  "subagent.action:" + ar.Kind,
				App:     appConduraG,
				Level:   level,
				Result:  decStr,
				Message: reason,
			}); err != nil {
				slog.Warn("audit sub-agent action failed", "kind", ar.Kind, "agent", ar.AgentName, "err", err)
			}
		}
	}
	return out
}
