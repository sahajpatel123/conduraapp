package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/delegation"
	"github.com/sahajpatel123/synapticapp/internal/failover"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
	"github.com/sahajpatel123/synapticapp/internal/pending"
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
		// Phase 17, Fix #7 (B5) + Phase 18 (v0.2.0): parse
		// structured ActionRequests from the sub-agent's output,
		// gate each one, persist to pending_actions, and publish
		// SSE so the GUI can show a live queue. Without
		// persistence the GUI would have to be running in the
		// exact window between Spawn returning and the user
		// clicking Approve — a real product would lose rows on
		// daemon restart.
		actionRows := gateAndPersistParsedActions(ctx, subs, result)
		// Surface the persisted IDs so the GUI can poll them
		// or subscribe to the SSE stream for live updates.
		pendingIDs := make([]string, len(actionRows))
		for i, r := range actionRows {
			pendingIDs[i] = r.ID
		}
		return map[string]any{
			"agent_name":       result.AgentName,
			"output":           result.Output,
			"exit_code":        result.ExitCode,
			"duration_ms":      result.Duration.Milliseconds(),
			"token_count":      result.TokenCount,
			"spawn_id":         result.SpawnID,
			"pending_actions":  actionRows,
			"pending_action_ids": pendingIDs,
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

	// Pending-actions queue RPCs (Phase 18 / v0.2.0).
	registerPendingActionMethods(srv, subs)
}

// registerPendingActionMethods wires the v0.2.0 pending-actions
// queue RPCs. Every method guards on subs.Pending != nil so the
// daemon can run without the queue (tests, headless deployments).
func registerPendingActionMethods(srv *ipc.Server, subs *Subsystems) {
	srv.Register("delegate.pending.list", func(_ context.Context, params json.RawMessage) (any, error) {
		if subs.Pending == nil {
			return map[string]any{"actions": []any{}}, nil
		}
		var p struct {
			Status string `json:"status"`
			Limit  int    `json:"limit"`
		}
		if len(params) > 0 {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
			}
		}
		rows, err := subs.Pending.List(context.Background(), pending.Status(p.Status), p.Limit)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return map[string]any{"actions": rows}, nil
	})

	srv.Register("delegate.pending.get", func(_ context.Context, params json.RawMessage) (any, error) {
		if subs.Pending == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "pending store not available"}
		}
		var p struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if p.ID == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "id is required"}
		}
		row, err := subs.Pending.Get(context.Background(), p.ID)
		if err != nil {
			if errors.Is(err, pending.ErrNotFound) {
				return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "pending action not found"}
			}
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return row, nil
	})

	srv.Register("delegate.pending.decide", func(ctx context.Context, params json.RawMessage) (any, error) {
		if subs.Pending == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "pending store not available"}
		}
		var p struct {
			ID         string `json:"id"`
			Decision   string `json:"decision"`
			DecidedBy  string `json:"decided_by"`
			Note       string `json:"note"`
			AutoRun    bool   `json:"auto_run"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if p.ID == "" || p.Decision == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "id and decision are required"}
		}
		row, err := subs.Pending.Decide(ctx, pending.DecisionInput{
			ID: p.ID, Decision: p.Decision, DecidedBy: p.DecidedBy, Note: p.Note,
		})
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		// Audit row for the decision.
		if subs.Audit != nil {
			_ = subs.Audit.Append(ctx, audit.Event{
				Actor:  "user:" + p.DecidedBy,
				Action: "pending.decide:" + p.Decision,
				App:    appConduraG,
				Level:  auditLevelInfo,
				Result: string(row.Status),
				Message: "id=" + row.ID + " kind=" + row.Kind + " agent=" + row.AgentName,
			})
		}
		// Publish SSE so other GUI windows update live.
		if subs.Broker != nil {
			subs.Broker.PublishJSON("pending_action."+string(row.Status), row)
		}
		// If auto_run was set and the row was approved, fire
		// the executor immediately. The caller can also do this
		// with a separate delegate.pending.execute call.
		if p.AutoRun && row.Status == pending.StatusApproved && subs.Executor != nil {
			execResult, execErr := subs.Executor.Execute(ctx, row)
			if execErr == nil && execResult != nil {
				_ = subs.Pending.MarkExecuted(ctx, row.ID,
					execResult.ExitCode, execResult.Result, execResult.Error, execResult.Duration)
			}
			// Re-fetch the row to capture the new status.
			if updated, gerr := subs.Pending.Get(ctx, row.ID); gerr == nil {
				if subs.Broker != nil {
					subs.Broker.PublishJSON("pending_action."+string(updated.Status), updated)
				}
				if subs.Audit != nil {
					level := auditLevelInfo
					res := "executed"
					if updated.Status == pending.StatusFailed {
						level = auditLevelWarn
						res = "failed"
					}
					_ = subs.Audit.Append(ctx, audit.Event{
						Actor:   "executor",
						Action:  "pending.executed",
						App:     appConduraG,
						Level:   level,
						Result:  res,
						Message: fmt.Sprintf("id=%s exit=%d duration_ms=%d err=%s",
							updated.ID, updated.ExitCode, updated.DurationMS, updated.ExecutionError),
					})
				}
				return updated, nil
			}
		}
		return row, nil
	})

	srv.Register("delegate.pending.execute", func(ctx context.Context, params json.RawMessage) (any, error) {
		if subs.Pending == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "pending store not available"}
		}
		if subs.Executor == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "executor not configured"}
		}
		var p struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if p.ID == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "id is required"}
		}
		row, err := subs.Pending.Get(ctx, p.ID)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "pending action not found"}
		}
		if row.Status != pending.StatusApproved {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "pending action must be approved first (got " + string(row.Status) + ")"}
		}
		result, execErr := subs.Executor.Execute(ctx, row)
		if execErr != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: execErr.Error()}
		}
		if result == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "executor returned nil result"}
		}
		if err := subs.Pending.MarkExecuted(ctx, row.ID,
			result.ExitCode, result.Result, result.Error, result.Duration); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		// Re-fetch for the post-execute status.
		updated, _ := subs.Pending.Get(ctx, row.ID)
		if subs.Broker != nil && updated != nil {
			subs.Broker.PublishJSON("pending_action."+string(updated.Status), updated)
		}
		if subs.Audit != nil && updated != nil {
			level := auditLevelInfo
			res := "executed"
			if updated.Status == pending.StatusFailed {
				level = auditLevelWarn
				res = "failed"
			}
			_ = subs.Audit.Append(ctx, audit.Event{
				Actor:   "executor",
				Action:  "pending.executed",
				App:     appConduraG,
				Level:   level,
				Result:  res,
				Message: fmt.Sprintf("id=%s exit=%d duration_ms=%d err=%s",
					updated.ID, updated.ExitCode, updated.DurationMS, updated.ExecutionError),
			})
		}
		return updated, nil
	})

	// Manual sweep trigger (test helper; not exposed in the GUI).
	srv.Register("delegate.pending.sweep", func(ctx context.Context, _ json.RawMessage) (any, error) {
		if subs.Pending == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "pending store not available"}
		}
		n, err := subs.Pending.SweepExpired(ctx, time.Now())
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return map[string]any{"swept": n}, nil
	})
}

// gateAndPersistParsedActions walks every ActionRequest the
// sub-agent emitted, runs each through the Gatekeeper, persists
// the result to pending_actions (when the store is wired), audits
// every decision, and publishes SSE so the GUI can show the new
// rows in real time. Returns the persisted rows in request order
// so the caller can include them in the delegate.spawn response.
//
// v0.2.0 design: every ActionRequest becomes a row, regardless
// of gate decision. Rows that the gate denied get marked as
// StatusDenied immediately; rows the gate wants to require_consent
// for stay pending so the GUI can prompt. Rows the gate outright
// allows become pending too — the user still has to click Approve
// because the agent's word is not sufficient.
//
// Without the pending store (tests, headless mode) we still gate
// + audit, but the result is in-memory only.
func gateAndPersistParsedActions(ctx context.Context, subs *Subsystems, result *delegation.SpawnResult) []*pending.Action {
	if subs == nil || subs.Delegation == nil || subs.Gatekeeper == nil || result == nil {
		return nil
	}
	requests := subs.Delegation.ActionRequests(result)
	if len(requests) == 0 {
		return nil
	}
	out := make([]*pending.Action, 0, len(requests))
	for _, ar := range requests {
		ba := blastradius.Action{
			Kind:      ar.Kind,
			TargetApp: ar.AgentName,
			Body:      ar.Body,
			Path:      ar.Path,
			Command:   ar.Command,
		}
		decision, reason := subs.Gatekeeper.Evaluate(ctx, ba)
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
		_ = allowed // captured in the row's gate_decision; the GUI re-reads it from there

		// Audit the gate verdict (always).
		level := auditLevelInfo
		if decision == gatekeeper.Deny || decision == gatekeeper.RequireConsent ||
			decision == gatekeeper.RequirePresenceAndConsent {
			level = auditLevelWarn
		}
		if subs.Audit != nil {
			_ = subs.Audit.Append(ctx, audit.Event{
				Actor:   "sub-agent:" + ar.AgentName,
				Action:  "subagent.action:" + ar.Kind,
				App:     appConduraG,
				Level:   level,
				Result:  decStr,
				Message: reason,
			})
		}

		// Persist to the pending store (best-effort; absent store
		// just means in-memory only).
		if subs.Pending == nil {
			continue
		}
		row, perr := subs.Pending.Insert(ctx, pending.InsertInput{
			SpawnID:      result.SpawnID,
			AgentName:    ar.AgentName,
			Kind:         ar.Kind,
			Payload:      pending.Payload{Command: ar.Command, Path: ar.Path, Body: ar.Body},
			GateDecision: decStr,
			GateReason:   reason,
			BlastClass:   blastradius.Classify(ba).String(),
			TTL:          pending.DefaultTTL,
		})
		if perr != nil {
			slog.Warn("persist pending action failed", "kind", ar.Kind, "agent", ar.AgentName, "err", perr)
			continue
		}
		// If the gate outright denied (not consent-required),
		// mark the row denied immediately so it doesn't sit in
		// the GUI's pending list forever.
		if decision == gatekeeper.Deny {
			if _, derr := subs.Pending.Decide(ctx, pending.DecisionInput{
				ID: row.ID, Decision: "deny", DecidedBy: "gate:deny",
			}); derr != nil {
				slog.Warn("auto-deny failed", "id", row.ID, "err", derr)
			}
			if updated, gerr := subs.Pending.Get(ctx, row.ID); gerr == nil {
				row = updated
			}
		}

		out = append(out, row)
		// Publish SSE for the GUI.
		if subs.Broker != nil {
			subs.Broker.PublishJSON("pending_action."+string(row.Status), row)
		}
	}
	return out
}
