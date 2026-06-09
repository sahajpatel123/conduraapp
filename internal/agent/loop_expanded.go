package agent

import (
	"context"
	"fmt"
	"time"
)

// ExpandedLoop is the expanded agent loop with multi-step planning.
type ExpandedLoop struct {
	// Planner decomposes tasks into steps.
	Planner Planner

	// Verifier checks step results.
	Verifier Verifier

	// Executor executes actions.
	Executor Executor

	// MaxRetries is the maximum number of retries per step.
	MaxRetries int

	// StepTimeout is the timeout for each step.
	StepTimeout time.Duration
}

// Executor executes computer-use actions.
type Executor interface {
	// Execute performs an action and returns the result.
	Execute(ctx context.Context, action *Action) (*StepResult, error)
}

// NewExpandedLoop creates a new expanded agent loop.
func NewExpandedLoop(planner Planner, verifier Verifier, executor Executor) *ExpandedLoop {
	return &ExpandedLoop{
		Planner:     planner,
		Verifier:    verifier,
		Executor:    executor,
		MaxRetries:  3,
		StepTimeout: 30 * time.Second,
	}
}

// ExecutePlan executes a plan step by step.
func (l *ExpandedLoop) ExecutePlan(ctx context.Context, plan *Plan) (*PlanResult, error) {
	if plan == nil || len(plan.Steps) == 0 {
		return nil, fmt.Errorf("empty plan")
	}

	result := &PlanResult{
		Goal:    plan.Goal,
		Started: time.Now(),
	}

	for i, step := range plan.Steps {
		// Check if context is canceled
		if ctx.Err() != nil {
			result.Finished = time.Now()
			result.Success = false
			result.Error = ctx.Err()
			return result, ctx.Err()
		}

		// Check dependencies
		if !l.dependenciesMet(plan, i) {
			step.Status = StepSkipped
			continue
		}

		// Execute the step
		step.Status = StepRunning
		stepResult, err := l.executeStep(ctx, step)
		step.Result = stepResult
		step.Status = StepCompleted

		if err != nil {
			step.Status = StepFailed
			step.Result = &StepResult{
				Success: false,
				Error:   err,
			}

			// Check if we should abort. Use step.Result (which
			// carries the error) rather than stepResult, which may
			// be nil when executeStep fails.
			verification, _ := l.Verifier.Verify(ctx, step, step.Result)
			if verification != nil && verification.ShouldAbort {
				result.Finished = time.Now()
				result.Success = false
				result.Error = err
				return result, err
			}

			// For now, don't retry - just fail
			result.Finished = time.Now()
			result.Success = false
			result.Error = err
			return result, err
		}

		result.Steps = append(result.Steps, stepResult)
	}

	result.Finished = time.Now()
	result.Success = true
	return result, nil
}

// executeStep executes a single step with timeout and verification.
func (l *ExpandedLoop) executeStep(ctx context.Context, step *Step) (*StepResult, error) {
	// Create a timeout context
	stepCtx, cancel := context.WithTimeout(ctx, l.StepTimeout)
	defer cancel()

	// Execute the action
	result, err := l.Executor.Execute(stepCtx, step.Action)
	if err != nil {
		return nil, err
	}

	// Verify the result
	verification, err := l.Verifier.Verify(stepCtx, step, result)
	if err != nil {
		return result, err
	}

	if !verification.Valid {
		return result, fmt.Errorf("verification failed: %s", verification.Reason)
	}

	return result, nil
}

// dependenciesMet checks if all dependencies for a step are satisfied.
func (l *ExpandedLoop) dependenciesMet(plan *Plan, stepIndex int) bool {
	step := plan.Steps[stepIndex]
	for _, dep := range step.DependsOn {
		if dep >= len(plan.Steps) {
			return false
		}
		if plan.Steps[dep].Status != StepCompleted {
			return false
		}
	}
	return true
}

// PlanResult is the result of executing a plan.
type PlanResult struct {
	// Goal is the high-level goal.
	Goal string

	// Success indicates whether the plan succeeded.
	Success bool

	// Steps is the results of each step.
	Steps []*StepResult

	// Started is when the plan started.
	Started time.Time

	// Finished is when the plan finished.
	Finished time.Time

	// Error is the error if the plan failed.
	Error error
}

// Duration returns how long the plan took.
func (r *PlanResult) Duration() time.Duration {
	return r.Finished.Sub(r.Started)
}
