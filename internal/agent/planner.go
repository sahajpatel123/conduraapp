package agent

import (
	"context"

	"github.com/sahajpatel123/conduraapp/internal/blastradius"
)

// Planner decomposes tasks into ordered steps.
type Planner interface {
	// Decompose breaks down a task into a sequence of steps.
	Decompose(ctx context.Context, task string, context *Context) (*Plan, error)

	// Reprioritize adjusts the plan based on new information.
	Reprioritize(ctx context.Context, plan *Plan, newInfo *Observation) (*Plan, error)
}

// Context provides additional context for planning.
type Context struct {
	// CurrentState describes the current screen state.
	CurrentState string

	// UserGoal is the high-level goal the user wants to achieve.
	UserGoal string

	// AvailableActions lists the actions the agent can take.
	AvailableActions []string

	// History is the conversation history.
	History []string
}

// Plan represents a sequence of steps to accomplish a task.
type Plan struct {
	// Steps is the ordered list of steps to execute.
	Steps []*Step

	// Current is the index of the current step.
	Current int

	// Goal is the high-level goal.
	Goal string

	// Context is the planning context.
	Context *Context
}

// Step represents a single step in a plan.
type Step struct {
	// Description is a human-readable description of the step.
	Description string

	// Action is the action to perform.
	Action *Action

	// DependsOn lists the indices of steps that must complete before this one.
	DependsOn []int

	// Status is the current status of the step.
	Status StepStatus

	// Result is the result of executing the step.
	Result *StepResult
}

// StepStatus represents the status of a step.
type StepStatus string

// Step status values.
const (
	StepPending   StepStatus = "pending"   // Step not yet started
	StepRunning   StepStatus = "running"   // Step in progress
	StepCompleted StepStatus = "completed" // Step finished successfully
	StepFailed    StepStatus = "failed"    // Step failed
	StepSkipped   StepStatus = "skipped"   // Step skipped
)

// Action represents a computer-use action.
type Action struct {
	// Type is the action type (click, type, scroll, etc.).
	Type string

	// Target is the target element or coordinates.
	Target string

	// Value is the value to use (text to type, key to press, etc.).
	Value string

	// Description is a human-readable description of the action.
	Description string
}

// ToBlastRadius converts an agent.Action to a blastradius.Action
// for safety classification. This bridges the three incompatible
// Action types (blastradius, computeruse, agent).
func (a *Action) ToBlastRadius() blastradius.Action {
	kind := a.Type
	if kind == "" {
		kind = "chat"
	}
	return blastradius.Action{
		Kind:      kind,
		Body:      a.Value,
		TargetApp: a.Target,
	}
}

// StepResult is the result of executing a step.
type StepResult struct {
	// Success indicates whether the step succeeded.
	Success bool

	// Error is the error if the step failed.
	Error error

	// Output is any output from the step.
	Output string

	// Duration is how long the step took.
	Duration float64

	// SSBeforeRef is the screenshot-before reference ID (from
	// replay.ScreenshotStore.Put). Empty if no screenshot was captured.
	SSBeforeRef string
	// SSAfterRef is the screenshot-after reference ID.
	SSAfterRef string
}

// Observation represents new information about the environment.
type Observation struct {
	// Type is the observation type (screen_change, error, etc.).
	Type string

	// Content is the observation content.
	Content string

	// Timestamp is when the observation was made.
	Timestamp float64
}

// SimplePlanner is a basic planner that creates linear plans.
//
// Deprecated: SimplePlanner is retained for test coverage but is NOT
// wired into the daemon. The production planner is LLMPlanner
// (llm_planner.go), which sends a real LLM request to decompose the
// task. SimplePlanner will be removed in v0.2.0 once its tests are
// migrated to LLMPlanner.
type SimplePlanner struct{}

// NewSimplePlanner creates a new simple planner.
func NewSimplePlanner() *SimplePlanner {
	return &SimplePlanner{}
}

// Decompose creates a simple linear plan from a task description.
func (p *SimplePlanner) Decompose(_ context.Context, task string, _ *Context) (*Plan, error) {
	// For now, create a single-step plan
	step := &Step{
		Description: task,
		Action: &Action{
			Type:        "chat",
			Description: task,
		},
		Status: StepPending,
	}

	return &Plan{
		Steps:   []*Step{step},
		Current: 0,
		Goal:    task,
	}, nil
}

// Reprioritize adjusts the plan based on new observations.
func (p *SimplePlanner) Reprioritize(_ context.Context, plan *Plan, _ *Observation) (*Plan, error) {
	// Simple planner doesn't reprioritize
	return plan, nil
}
