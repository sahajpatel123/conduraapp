// Package agent — bridge executors that translate agent.Action
// into the actual underlying subsystems (computer-use, etc.).
//
// Phase 14I: replaces the noopAgentExecutor placeholder with a real
// agent.Executor that delegates to the ComputerUse pipeline, so
// chat messages that surface an "agent action" actually do something
// on the user's machine. Gating still happens in the GatedExecutor
// wrapper above this layer; this is the leaf executor.
//
// Reference: CLAUDE.md §10.1 (blast radius), §11 (computer use).
package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/computeruse"
)

// ComputerUseExecutor is an agent.Executor that delegates to the
// real ComputerUse pipeline. It implements the leaf executor
// contract: every Execute call translates the agent.Action into
// a computeruse.Action, runs it through the backend, and reports
// the outcome as a StepResult.
//
// This executor is GATE-AGNOSTIC. Wrap it through agent.NewGatedExecutor
// (or computeruse.NewGatedExecutor) before use so the safety layer
// sees every action.
type ComputerUseExecutor struct {
	CU *computeruse.ComputerUse
}

// NewComputerUseExecutor returns a ComputerUseExecutor backed by cu.
func NewComputerUseExecutor(cu *computeruse.ComputerUse) *ComputerUseExecutor {
	return &ComputerUseExecutor{CU: cu}
}

// Execute satisfies agent.Executor. Errors are surfaced via the
// returned StepResult.Error rather than the function's error return
// so the agent loop can log the failure and decide whether to
// retry, abort, or surface to the user.
func (e *ComputerUseExecutor) Execute(ctx context.Context, a *Action) (*StepResult, error) {
	if e == nil || e.CU == nil {
		return &StepResult{Success: false, Error: fmt.Errorf("computer-use executor not initialized")}, nil
	}
	cuAction, err := translateAgentAction(a) //nolint:govet // shadowed err is intentional
	if err != nil {
		//nolint:nilerr // err is propagated via StepResult.Error, not the function return
		return &StepResult{Success: false, Error: err}, nil
	}
	res, err := e.CU.Execute(ctx, cuAction) //nolint:govet // shadowed err is intentional
	if err != nil {
		//nolint:nilerr // err is propagated via StepResult.Error, not the function return
		return &StepResult{Success: false, Error: err}, nil
	}
	if res == nil {
		return &StepResult{Success: false, Error: fmt.Errorf("computer-use returned nil result")}, nil
	}
	if !res.Success {
		errMsg := "computer-use action failed"
		if res.Error != nil {
			errMsg = res.Error.Error()
		}
		return &StepResult{Success: false, Error: fmt.Errorf("%s", errMsg)}, nil
	}
	return &StepResult{
		Success: true,
		Output:  fmt.Sprintf("action=%s duration=%s", a.Type, res.Duration),
	}, nil
}

// translateAgentAction maps the simple agent.Action shape (Type, Target, Value)
// to the richer computeruse.Action shape. Best-effort: unknown types
// become ActionWait so the pipeline never returns "unknown type" at the
// blast-radius layer.
func translateAgentAction(a *Action) (*computeruse.Action, error) {
	if a == nil {
		return nil, fmt.Errorf("nil action")
	}
	out := &computeruse.Action{
		Value:   a.Value,
		Timeout: 5 * time.Second,
	}
	switch strings.ToLower(strings.TrimSpace(a.Type)) {
	case "click":
		out.Type = computeruse.ActionClick
		out.Target = &computeruse.Target{Title: a.Target}
	case "type", "input":
		out.Type = computeruse.ActionTypeText
		out.Value = a.Value
	case "scroll":
		out.Type = computeruse.ActionScroll
	case "key", "key_press", "press":
		out.Type = computeruse.ActionKeyPress
		out.Value = a.Value
	case "drag":
		out.Type = computeruse.ActionDrag
		out.Target = &computeruse.Target{Title: a.Target}
	case "launch":
		out.Type = computeruse.ActionLaunch
		out.Value = a.Target // the app name
	case "focus":
		out.Type = computeruse.ActionFocus
		out.Target = &computeruse.Target{Title: a.Target}
	case "wait":
		out.Type = computeruse.ActionWait
	default:
		// Unknown action type — fall back to wait so the gatekeeper
		// still sees a READ-classified event and we never silently
		// drop an untrusted type.
		out.Type = computeruse.ActionWait
	}
	return out, nil
}

// ComputerUseAvailable reports whether the executor's underlying
// pipeline is wired. Used by subsystems initialization to decide
// whether to install the real executor or keep the noop.
func (e *ComputerUseExecutor) ComputerUseAvailable() bool {
	return e != nil && e.CU != nil
}
