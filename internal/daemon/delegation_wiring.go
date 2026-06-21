package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
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

// errPendingNotConfigured is returned by every pending RPC when
// the Pending store is nil — happens when the daemon was
// constructed without a database (test harnesses, headless mode).
const errPendingNotConfigured = "pending store not available"

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
			"agent_name":         result.AgentName,
			"output":             result.Output,
			"exit_code":          result.ExitCode,
			"duration_ms":        result.Duration.Milliseconds(),
			"token_count":        result.TokenCount,
			"spawn_id":           result.SpawnID,
			"pending_actions":    actionRows,
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
//
// branch is a separate concern. We accept the complexity here
// rather than spreading file-3 wiring across 5 helpers.
//
//nolint:gocognit // wiring 5 RPCs in one place is intentional; each
func registerPendingActionMethods(srv *ipc.Server, subs *Subsystems) {
	srv.Register("delegate.pending.list", func(_ context.Context, params json.RawMessage) (any, error) {
		return pendingList(subs, params)
	})
	srv.Register("delegate.pending.get", func(_ context.Context, params json.RawMessage) (any, error) {
		return pendingGet(subs, params)
	})
	srv.Register("delegate.pending.decide", func(ctx context.Context, params json.RawMessage) (any, error) {
		return pendingDecide(ctx, subs, params)
	})
	srv.Register("delegate.pending.execute", func(ctx context.Context, params json.RawMessage) (any, error) {
		return pendingExecute(ctx, subs, params)
	})
	srv.Register("delegate.pending.sweep", func(ctx context.Context, _ json.RawMessage) (any, error) {
		return pendingSweep(ctx, subs)
	})
}

// pendingList handles delegate.pending.list.
func pendingList(subs *Subsystems, params json.RawMessage) (any, error) {
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
}

// pendingGet handles delegate.pending.get.
func pendingGet(subs *Subsystems, params json.RawMessage) (any, error) {
	if subs.Pending == nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errPendingNotConfigured}
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
}

// pendingDecide handles delegate.pending.decide (approve / deny).
// When AutoRun is set AND the decision is approve, the executor
// fires immediately inside this handler so the GUI's
// "Approve & Run" button is a single round-trip.
func pendingDecide(ctx context.Context, subs *Subsystems, params json.RawMessage) (any, error) {
	if subs.Pending == nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errPendingNotConfigured}
	}
	var p struct {
		ID        string `json:"id"`
		Decision  string `json:"decision"`
		DecidedBy string `json:"decided_by"`
		Note      string `json:"note"`
		AutoRun   bool   `json:"auto_run"`
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
	auditPendingDecision(ctx, subs, p.DecidedBy, row)
	publishPendingEvent(subs, row)
	if p.AutoRun && row.Status == pending.StatusApproved && subs.Executor != nil {
		return executeAndRecord(ctx, subs, row)
	}
	return row, nil
}

// pendingExecute handles delegate.pending.execute.
func pendingExecute(ctx context.Context, subs *Subsystems, params json.RawMessage) (any, error) {
	if subs.Pending == nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errPendingNotConfigured}
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
	return executeAndRecord(ctx, subs, row)
}

// pendingSweep handles delegate.pending.sweep (test helper).
func pendingSweep(ctx context.Context, subs *Subsystems) (any, error) {
	if subs.Pending == nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errPendingNotConfigured}
	}
	n, err := subs.Pending.SweepExpired(ctx, time.Now())
	if err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
	}
	return map[string]any{"swept": n}, nil
}

// auditPendingDecision writes the actor=user audit row after a
// user decision (approve / deny). Always best-effort; nil Audit
// is a no-op.
func auditPendingDecision(ctx context.Context, subs *Subsystems, decidedBy string, row *pending.Action) {
	if subs.Audit == nil {
		return
	}
	_ = subs.Audit.Append(ctx, audit.Event{
		Actor:   "user:" + decidedBy,
		Action:  "pending.decide:" + string(row.Status),
		App:     appConduraG,
		Level:   auditLevelInfo,
		Result:  string(row.Status),
		Message: "id=" + row.ID + " kind=" + row.Kind + " agent=" + row.AgentName,
	})
}

// publishPendingEvent fires the namespaced SSE event so the GUI
// updates live. No-op when the broker is missing.
func publishPendingEvent(subs *Subsystems, row *pending.Action) {
	if subs.Broker != nil {
		subs.Broker.PublishJSON("pending_action."+string(row.Status), row)
	}
}

// executeAndRecord runs the executor on an approved row, marks
// it executed/failed in the pending store, fires SSE + audit,
// and returns the post-execution row. Shared between
// delegate.pending.decide (auto_run=true branch) and
// delegate.pending.execute.
func executeAndRecord(ctx context.Context, subs *Subsystems, row *pending.Action) (*pending.Action, error) {
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
	updated, _ := subs.Pending.Get(ctx, row.ID)
	if updated != nil {
		publishPendingEvent(subs, updated)
		auditExecution(ctx, subs, updated)
	}
	return updated, nil
}

// auditExecution writes the actor=executor audit row after a
// row's execution completes.
func auditExecution(ctx context.Context, subs *Subsystems, row *pending.Action) {
	if subs.Audit == nil {
		return
	}
	level := auditLevelInfo
	res := "executed"
	if row.Status == pending.StatusFailed {
		level = auditLevelWarn
		res = "failed"
	}
	_ = subs.Audit.Append(ctx, audit.Event{
		Actor:  "executor",
		Action: "pending.executed",
		App:    appConduraG,
		Level:  level,
		Result: res,
		Message: fmt.Sprintf("id=%s exit=%d duration_ms=%d err=%s",
			row.ID, row.ExitCode, row.DurationMS, row.ExecutionError),
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
//
// (gate verdict -> audit level -> persist -> maybe auto-deny ->
// publish SSE). We keep it here rather than splitting across
// helpers because the helper would need 5+ arguments and the
// reader would lose the linear flow. The branches are explicit
// and tested individually in pending_e2e_test.go.
//
//nolint:gocyclo // the per-action logic is necessarily branchy
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
		if row := processOneActionRequest(ctx, subs, result.SpawnID, ar); row != nil {
			out = append(out, row)
		}
	}
	return out
}

// processOneActionRequest runs gate + audit + persist + SSE
// for a single ActionRequest. Returns nil for requests whose
// persist failed (the warning was already logged).
func processOneActionRequest(ctx context.Context, subs *Subsystems, spawnID string, ar delegation.ActionRequest) *pending.Action {
	ba := blastradius.Action{
		Kind:      ar.Kind,
		TargetApp: ar.AgentName,
		Body:      ar.Body,
		Path:      ar.Path,
		Command:   ar.Command,
	}
	decStr, reason := gateVerdictToString(subs.Gatekeeper.Evaluate(ctx, ba))

	// Audit the gate verdict (always).
	if subs.Audit != nil {
		level := auditLevelInfo
		if strings.HasSuffix(decStr, "deny") || decStr == "require_consent" || decStr == "require_presence_and_consent" {
			level = auditLevelWarn
		}
		_ = subs.Audit.Append(ctx, audit.Event{
			Actor:   "sub-agent:" + ar.AgentName,
			Action:  "subagent.action:" + ar.Kind,
			App:     appConduraG,
			Level:   level,
			Result:  decStr,
			Message: reason,
		})
	}

	if subs.Pending == nil {
		return nil
	}
	row, perr := subs.Pending.Insert(ctx, pending.InsertInput{
		SpawnID:      spawnID,
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
		return nil
	}
	// Outright deny: auto-flip so the row doesn't sit pending forever.
	if decStr == auditResultDeny {
		if _, derr := subs.Pending.Decide(ctx, pending.DecisionInput{
			ID: row.ID, Decision: "deny", DecidedBy: "gate:deny",
		}); derr != nil {
			slog.Warn("auto-deny failed", "id", row.ID, "err", derr)
		}
		if updated, gerr := subs.Pending.Get(ctx, row.ID); gerr == nil {
			row = updated
		}
	}
	if subs.Broker != nil {
		subs.Broker.PublishJSON("pending_action."+string(row.Status), row)
	}
	return row
}

// gateVerdictToString maps a gatekeeper verdict to the short
// string we store on the row + the audit row + the SSE event.
func gateVerdictToString(decision gatekeeper.Decision, reason string) (string, string) {
	switch decision {
	case gatekeeper.Allow:
		return "allow", reason
	case gatekeeper.Deny:
		return "deny", reason
	case gatekeeper.RequireConsent:
		return "require_consent", reason
	case gatekeeper.RequirePresenceAndConsent:
		return "require_presence_and_consent", reason
	default:
		return "unknown", reason
	}
}
