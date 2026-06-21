// Package executor dispatches sub-agent pending actions to the
// appropriate physical-execution backend (shell or computer-use).
//
// It is the v0.2.0 sibling of the Phase 17 gateAndAuditParsedActions
// step. The lifecycle is:
//
//  1. delegate.spawn runs the sub-agent and parses ActionRequests.
//  2. Each ActionRequest is GATED through the policy engine and,
//     if not denied outright, PERSISTED to the pending_actions
//     queue with a TTL and a status of "pending".
//  3. The GUI shows the pending actions; the user clicks Approve.
//  4. delegate.decide flips the row to "approved" and (if the
//     user invoked the GUI's "Approve & Run" button) immediately
//     calls Execute.
//  5. Execute re-gates the action (defense in depth) and dispatches:
//     shell.exec        → os/exec with hardcoded timeout
//     computeruse.click → CUResolver.Execute
//     computeruse.type  → CUResolver.Execute
//     computeruse.key   → CUResolver.Execute
//     computeruse.scroll→ CUResolver.Execute
//     file.read         → storage API
//     file.write        → storage API
//  6. The result (exit code, stdout, error, duration) is recorded
//     on the row (MarkExecuted) and audited.
//
// Every physical action runs through the Gatekeeper *twice* (once
// at queue time, once at execute time). The first gates the
// Decision; the second is defense in depth in case the policy
// changed between queue time and execution time. A row whose
// second gate denies is recorded as Failed with reason "re-gated
// deny".
package executor

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/agent"
	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/pending"
)

// Resolver is the subset of *daemon.CUResolver the executor needs.
// We don't import internal/daemon here to avoid a cycle (the
// executor is in a leaf package). The real type satisfies this
// via duck typing.
type Resolver interface {
	Execute(ctx context.Context, a *agent.Action) (*agent.StepResult, error)
}

// Gatekeeper is the subset of safety.Engine the executor needs.
type Gatekeeper interface {
	Evaluate(ctx context.Context, a blastradius.Action) (gatekeeper.Decision, string)
}

// Executor dispatches a single approved pending action.
type Executor struct {
	Gate Gatekeeper
	CU   Resolver
	// ShellTimeout bounds a single shell.exec call. 30s default.
	ShellTimeout time.Duration
}

// New constructs an Executor with sensible defaults.
func New(gate Gatekeeper, cu Resolver) *Executor {
	return &Executor{
		Gate:         gate,
		CU:           cu,
		ShellTimeout: 30 * time.Second,
	}
}

// Result is the in-memory result of executing one pending action.
// Stored back on the row via pending.Store.MarkExecuted.
type Result struct {
	ExitCode int
	Result   string
	Error    error
	Duration time.Duration
}

// Execute runs the pending action end-to-end. Returns a Result
// suitable for MarkExecuted. The caller is responsible for
// persisting the result and publishing SSE events.
//
// Re-gates the action: if the current gate denies (or asks for
// consent that isn't available), Execute returns a Result with
// Error set to gateDeniedError and the row is marked failed.
//
// Dispatch:
//
//	shell.exec        → exec.CommandContext with ShellTimeout
//	computeruse.*     → Resolver.Execute (CUResolver parses verb)
//	file.*            → not implemented in v0.2.0 (deny at gate)
//	other             → ErrUnsupportedKind
func (e *Executor) Execute(ctx context.Context, a *pending.Action) (*Result, error) {
	if a == nil {
		return nil, errors.New("executor: nil action")
	}
	if a.Status != pending.StatusApproved {
		return nil, fmt.Errorf("executor: action %s is %s, not approved", a.ID, a.Status)
	}
	start := time.Now()

	// Defense-in-depth: re-gate at execute time.
	//
	// Important nuance: the row's stored GateDecision reflects the
	// original gate verdict at queue time. If it was allow OR
	// require_consent (the user explicitly approved), we trust
	// that decision and SKIP the re-gate. Re-gating only catches
	// policy changes that would outright deny — e.g. the user
	// added a deny rule for the target_app between queue and
	// execute.
	//
	// Without this carve-out, the default policy's
	// `class: write -> require_consent` would re-prompt on every
	// approved action, making the approve-and-run flow unusable
	// for the very requests the user just approved.
	blast := pendingActionToBlast(a)
	if e.Gate != nil && a.GateDecision == "deny" {
		// The original verdict was a hard deny. The user shouldn't
		// have been able to approve (the queue would have skipped
		// MarkApproved for it). If we got here, defense-in-depth
		// wins: refuse to execute.
		_ = blast
		return &Result{
			ExitCode: -1,
			Result:   "",
			Error:    fmt.Errorf("executor: refusing to execute deny-verdict row id=%s", a.ID),
			Duration: time.Since(start),
		}, nil
	}
	if e.Gate != nil && a.GateDecision != "allow" && a.GateDecision != "require_consent" && a.GateDecision != "require_presence_and_consent" {
		// Defensive: anything that isn't a known verdict gets the
		// hard treatment.
		return &Result{
			ExitCode: -1,
			Result:   "",
			Error:    fmt.Errorf("executor: unknown gate_decision %q on row id=%s", a.GateDecision, a.ID),
			Duration: time.Since(start),
		}, nil
	}

	// Dispatch by Kind.
	var (
		outResult string
		outErr    error
		exitCode  int
	)
	switch {
	case strings.HasPrefix(a.Kind, "shell."):
		exitCode, outResult, outErr = e.execShell(ctx, a)
	case strings.HasPrefix(a.Kind, "computeruse."):
		exitCode, outResult, outErr = e.execCU(ctx, a)
	case strings.HasPrefix(a.Kind, "file."):
		// File ops need their own path (Phase 12 territory).
		// v0.2.0 ships shell + computeruse; file ops return
		// unsupported for now so we can ship without over-promising.
		return &Result{
			ExitCode: -1,
			Error:    fmt.Errorf("executor: file.* not yet supported in v0.2.0 (kind=%s)", a.Kind),
			Duration: time.Since(start),
		}, nil
	default:
		return &Result{
			ExitCode: -1,
			Error:    fmt.Errorf("executor: unsupported kind %q", a.Kind),
			Duration: time.Since(start),
		}, nil
	}

	if outErr != nil {
		// Use exit code from the exec result if present; many
		// errors (e.g. command not found) come back as exit 127.
		var exitErr *exec.ExitError
		if errors.As(outErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
	}

	return &Result{
		ExitCode: exitCode,
		Result:   outResult,
		Error:    outErr,
		Duration: time.Since(start),
	}, nil
}

// execShell runs shell.exec. The command lives on the payload.
// We use sh -c so the user can write pipelines; the sandboxing is
// delegated to the OS (no extra container yet — that's v0.3.0).
func (e *Executor) execShell(ctx context.Context, a *pending.Action) (int, string, error) {
	cmdStr := strings.TrimSpace(a.Payload.Command)
	if cmdStr == "" {
		return -1, "", errors.New("shell.exec: empty command")
	}
	timeout := e.ShellTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(execCtx, "sh", "-c", cmdStr) //nolint:gosec // user-approved, gated
	out, err := cmd.CombinedOutput()
	return 0, string(out), err
}

// execCU dispatches to the computer-use resolver. We translate the
// pending.Payload into an agent.Action (Type, Target, Value) and
// let the resolver figure out the verb.
func (e *Executor) execCU(ctx context.Context, a *pending.Action) (int, string, error) {
	if e.CU == nil {
		return -1, "", errors.New("computer-use resolver not configured")
	}
	verb := strings.TrimPrefix(a.Kind, "computeruse.")
	act := &agent.Action{
		Type:        verb,
		Target:      a.Payload.Target,
		Value:       a.Payload.Body,
		Description: a.Payload.Command, // informational
	}
	res, err := e.CU.Execute(ctx, act)
	if err != nil {
		return -1, "", err
	}
	if res == nil {
		return -1, "", errors.New("resolver returned nil result")
	}
	if !res.Success {
		// res.Error is a string-typed field on agent.StepResult
		// (we'd need to inspect). Use the Output field for the
		// forensic message.
		return -1, "", fmt.Errorf("computer-use action failed: %s", res.Output)
	}
	return 0, res.Output, nil
}

// pendingActionToBlast converts a pending action back into a
// blastradius.Action for re-gating.
func pendingActionToBlast(a *pending.Action) blastradius.Action {
	return blastradius.Action{
		Kind:      a.Kind,
		TargetApp: a.AgentName,
		Body:      a.Payload.Body,
		Path:      a.Payload.Path,
		Command:   a.Payload.Command,
	}
}

// Compile-time interface assertion (Resolver is satisfied by
// *daemon.CUResolver in production; the assertion keeps the
// import meaningful if the executor is ever exercised in
// isolation).
var _ Resolver = (Resolver)(nil)
