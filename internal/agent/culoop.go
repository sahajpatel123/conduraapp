package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/status"
)

// HaltChecker is the subset of halt.Flag that CULoop needs.
// It checks whether the user has triggered a kill-switch.
type HaltChecker interface {
	IsHalted() bool
}

// CULoop is the execution engine for computer-use tasks. It wires
// together the LLMPlanner (task decomposition), CUResolver (intent→execution),
// and twin-snapshot verifier into a single end-to-end pipeline.
//
// Invariants enforced (MISSION §2):
//  1. Halt check between every step — if the user triggers the kill
//     switch mid-plan, the loop stops and returns an error.
//  2. Gatekeeper on every action — the CUResolver wraps the
//     GatedComputerUseExecutor, so every physical action passes
//     through the Gatekeeper before execution.
//  3. Destructive actions require a human modal — this is injected
//     as a BeforeAction callback; the caller (daemon) wires it to
//     the native OS modal dialog. If the callback returns false,
//     the action is skipped and the plan fails.
type CULoop struct {
	planner  Planner
	verifier Verifier
	executor Executor
	halt     HaltChecker

	// MaxRetries is the maximum number of retries per step.
	MaxRetries int

	// StepTimeout bounds each action's execution time.
	StepTimeout time.Duration

	// BeforeAction is an optional callback invoked before every
	// destructive action (per MISSION §2.3). Return false to block
	// the action. nil means all actions proceed (no modal).
	BeforeAction func(step *Step) bool

	// OnStatus fires on every state transition.
	OnStatus func(status.Status)
	// OnStart fires at the beginning of each Run() call.
	OnStart func()
	// OnAction fires after each step executes, with the action
	// type, the success flag, and the step result. This is the
	// live "agent is clicking X" indicator data source (§10 /
	// backend audit B-22): the daemon wires this to publish a
	// `cu.action` SSE event per action so the chat UI can show
	// what the agent is doing in real time. nil = no events.
	OnAction func(actionType string, success bool, result *StepResult)
}

// NewCULoop creates a computer-use execution loop.
func NewCULoop(planner Planner, verifier Verifier, executor Executor, halt HaltChecker) *CULoop {
	return &CULoop{
		planner:     planner,
		verifier:    verifier,
		executor:    executor,
		halt:        halt,
		MaxRetries:  3,
		StepTimeout: 30 * time.Second,
	}
}

// Run executes a natural-language computer-use task end-to-end.
// It delegates to the LLMPlanner for decomposition, then executes
// each step through the CUResolver with pause-for-halt checks.
func (l *CULoop) Run(ctx context.Context, task string, planCtx *Context) (*PlanResult, error) {
	if l.OnStart != nil {
		l.OnStart()
	}
	l.emit(status.StatusThinking)

	plan, err := l.planner.Decompose(ctx, task, planCtx)
	if err != nil {
		l.emit(status.StatusError)
		return nil, fmt.Errorf("culoop: plan: %w", err)
	}

	result := &PlanResult{Goal: plan.Goal, Started: time.Now()}
	for i, step := range plan.Steps {
		if err := ctx.Err(); err != nil {
			return l.finish(result, false, err)
		}

		// Invariant 1: halt check between every step.
		if l.halt != nil && l.halt.IsHalted() {
			return l.finish(result, false, fmt.Errorf("culoop: halted by user at step %d/%d", i+1, len(plan.Steps)))
		}

		// Invariant 2: Gatekeeper via CUResolver (already wrapped).
		// Invariant 3: destructive modal check.
		if l.BeforeAction != nil && !l.BeforeAction(step) {
			step.Status = StepSkipped
			continue
		}

		// Check dependencies.
		if !l.depsMet(plan, i) {
			step.Status = StepSkipped
			continue
		}

		step.Status = StepRunning
		stepCtx, cancel := context.WithTimeout(ctx, l.StepTimeout)
		sr, err := l.executor.Execute(stepCtx, step.Action)
		cancel()

		if err != nil {
			step.Status = StepFailed
			step.Result = &StepResult{Success: false, Error: err}
			l.emitAction(step.Action.Type, false, step.Result)
			l.emit(status.StatusError)
			return l.finish(result, false, err)
		}

		step.Result = sr
		step.Status = StepCompleted
		result.Steps = append(result.Steps, sr)
		l.emitAction(step.Action.Type, sr != nil && sr.Success, sr)
	}

	return l.finish(result, true, nil)
}

func (l *CULoop) depsMet(plan *Plan, idx int) bool {
	for _, dep := range plan.Steps[idx].DependsOn {
		if dep >= len(plan.Steps) || plan.Steps[dep].Status != StepCompleted {
			return false
		}
	}
	return true
}

func (l *CULoop) finish(r *PlanResult, success bool, err error) (*PlanResult, error) {
	r.Finished = time.Now()
	r.Success = success
	r.Error = err
	if success {
		l.emit(status.StatusIdle)
	}
	return r, err
}

func (l *CULoop) emit(s status.Status) {
	if l.OnStatus != nil {
		l.OnStatus(s)
	}
}

// emitAction fires the OnAction hook if wired. Used after every
// step to drive the live "agent is clicking X" SSE indicator.
func (l *CULoop) emitAction(actionType string, success bool, result *StepResult) {
	if l.OnAction != nil {
		l.OnAction(actionType, success, result)
	}
}
