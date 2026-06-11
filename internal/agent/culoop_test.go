package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/status"
)

// fakeHalt implements HaltChecker for tests.
type fakeHalt struct{ halted bool }

func (h *fakeHalt) IsHalted() bool { return h.halted }

// fakeVerifier implements Verifier for tests.
type fakeVerifier struct {
	result *VerificationResult
	err    error
}

func (v *fakeVerifier) Verify(_ context.Context, _ *Step, _ *StepResult) (*VerificationResult, error) {
	return v.result, v.err
}

func (v *fakeVerifier) ShouldRetry(_ context.Context, _ *StepResult, _ int) bool {
	return false
}

func TestCULoop_Run_Success(t *testing.T) {
	plan := &Plan{
		Goal: "test",
		Steps: []*Step{
			{Action: &Action{Type: "click", Target: "button"}, Status: StepPending},
			{Action: &Action{Type: "type", Value: "hello"}, Status: StepPending},
		},
	}
	planner := &fakePlanner{plan: plan}
	exec := &mockExecutor{results: []*StepResult{
		{Success: true, Duration: 0.1},
		{Success: true, Duration: 0.2},
	}}
	halt := &fakeHalt{}
	verifier := &fakeVerifier{result: &VerificationResult{Valid: true}}

	loop := NewCULoop(planner, verifier, exec, halt)
	loop.StepTimeout = 5 * time.Second

	result, err := loop.Run(context.Background(), "click button, type hello", nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if len(result.Steps) != 2 {
		t.Fatalf("got %d steps", len(result.Steps))
	}
}

func TestCULoop_Run_HaltedMidPlan(t *testing.T) {
	plan := &Plan{
		Goal: "test",
		Steps: []*Step{
			{Action: &Action{Type: "click"}, Status: StepPending},
		},
	}
	planner := &fakePlanner{plan: plan}
	exec := &mockExecutor{}
	halt := &fakeHalt{halted: true}
	verifier := &fakeVerifier{}

	loop := NewCULoop(planner, verifier, exec, halt)
	_, err := loop.Run(context.Background(), "do something", nil)
	if err == nil {
		t.Fatal("expected error when halted")
	}
}

func TestCULoop_Run_PlannerError(t *testing.T) {
	planner := &fakePlanner{err: errors.New("planner failed")}
	loop := NewCULoop(planner, &fakeVerifier{}, &mockExecutor{}, &fakeHalt{})
	_, err := loop.Run(context.Background(), "test", nil)
	if err == nil {
		t.Fatal("expected planner error")
	}
}

func TestCULoop_Run_BeforeActionBlocks(t *testing.T) {
	plan := &Plan{
		Goal:  "test",
		Steps: []*Step{{Action: &Action{Type: "click"}, Status: StepPending}},
	}
	planner := &fakePlanner{plan: plan}
	exec := &mockExecutor{results: []*StepResult{{Success: true}}}
	loop := NewCULoop(planner, &fakeVerifier{}, exec, &fakeHalt{})

	// Block all actions.
	loop.BeforeAction = func(_ *Step) bool { return false }

	result, err := loop.Run(context.Background(), "click", nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Success {
		t.Error("plan should succeed, step skipped not failed")
	}
}

func TestCULoop_Run_OnStatusCalled(t *testing.T) {
	plan := &Plan{
		Goal:  "test",
		Steps: []*Step{{Action: &Action{Type: "click"}, Status: StepPending}},
	}
	planner := &fakePlanner{plan: plan}
	exec := &mockExecutor{results: []*StepResult{{Success: true}}}
	loop := NewCULoop(planner, &fakeVerifier{}, exec, &fakeHalt{})

	var states []status.Status
	loop.OnStatus = func(s status.Status) { states = append(states, s) }

	_, err := loop.Run(context.Background(), "click", nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(states) < 2 {
		t.Fatalf("expected at least 2 status transitions, got %d", len(states))
	}
	if states[0] != status.StatusThinking {
		t.Errorf("first status = %v, want thinking", states[0])
	}
	if states[len(states)-1] != status.StatusIdle {
		t.Errorf("last status = %v, want idle", states[len(states)-1])
	}
}

// fakePlanner implements Planner for CULoop tests.
type fakePlanner struct {
	plan *Plan
	err  error
}

func (p *fakePlanner) Decompose(_ context.Context, _ string, _ *Context) (*Plan, error) {
	return p.plan, p.err
}

func (p *fakePlanner) Reprioritize(_ context.Context, plan *Plan, _ *Observation) (*Plan, error) {
	return plan, nil
}
