package agent

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type mockExecutor struct {
	results []*StepResult
	calls   int
}

func (m *mockExecutor) Execute(_ context.Context, _ *Action) (*StepResult, error) {
	if m.calls >= len(m.results) {
		return nil, fmt.Errorf("no more results")
	}
	result := m.results[m.calls]
	m.calls++
	return result, nil
}

func TestSimplePlanner(t *testing.T) {
	planner := NewSimplePlanner()
	ctx := context.Background()

	plan, err := planner.Decompose(ctx, "Click the OK button", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan == nil {
		t.Fatal("expected plan, got nil")
	}
	if len(plan.Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(plan.Steps))
	}
	if plan.Goal != "Click the OK button" {
		t.Errorf("goal = %v, want 'Click the OK button'", plan.Goal)
	}
}

func TestSimpleVerifier(t *testing.T) {
	verifier := NewSimpleVerifier()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		result := &StepResult{Success: true}
		verification, err := verifier.Verify(ctx, nil, result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !verification.Valid {
			t.Errorf("expected valid, got invalid: %s", verification.Reason)
		}
	})

	t.Run("failure", func(t *testing.T) {
		result := &StepResult{Success: false, Error: fmt.Errorf("failed")}
		verification, err := verifier.Verify(ctx, nil, result)
		if err == nil {
			t.Error("expected error")
		}
		if verification.Valid {
			t.Error("expected invalid, got valid")
		}
	})

	t.Run("nil result", func(t *testing.T) {
		verification, err := verifier.Verify(ctx, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if verification.Valid {
			t.Error("expected invalid, got valid")
		}
		if !verification.ShouldAbort {
			t.Error("expected ShouldAbort=true")
		}
	})
}

func TestSimpleVerifierShouldRetry(t *testing.T) {
	verifier := NewSimpleVerifier()
	ctx := context.Background()

	tests := []struct {
		attempt  int
		expected bool
	}{
		{0, true},
		{1, true},
		{2, true},
		{3, false},
		{4, false},
	}

	for _, tt := range tests {
		result := verifier.ShouldRetry(ctx, nil, tt.attempt)
		if result != tt.expected {
			t.Errorf("attempt %d: got %v, want %v", tt.attempt, result, tt.expected)
		}
	}
}

func TestExpandedLoop(t *testing.T) {
	executor := &mockExecutor{
		results: []*StepResult{
			{Success: true, Output: "done"},
		},
	}

	loop := NewExpandedLoop(
		NewSimplePlanner(),
		NewSimpleVerifier(),
		executor,
	)

	plan := &Plan{
		Goal: "Test plan",
		Steps: []*Step{
			{
				Description: "Test step",
				Action:      &Action{Type: "chat"},
				Status:      StepPending,
			},
		},
	}

	ctx := context.Background()
	result, err := loop.ExecutePlan(ctx, plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !result.Success {
		t.Error("expected success")
	}
	if len(result.Steps) != 1 {
		t.Errorf("expected 1 step result, got %d", len(result.Steps))
	}
	if result.Duration() < 0 {
		t.Error("expected non-negative duration")
	}
}

func TestExpandedLoopEmptyPlan(t *testing.T) {
	executor := &mockExecutor{}
	loop := NewExpandedLoop(
		NewSimplePlanner(),
		NewSimpleVerifier(),
		executor,
	)

	ctx := context.Background()
	_, err := loop.ExecutePlan(ctx, &Plan{Goal: "empty"})
	if err == nil {
		t.Error("expected error for empty plan")
	}
}

func TestExpandedLoopCancellation(t *testing.T) {
	executor := &mockExecutor{
		results: []*StepResult{
			{Success: true},
		},
	}

	loop := NewExpandedLoop(
		NewSimplePlanner(),
		NewSimpleVerifier(),
		executor,
	)

	plan := &Plan{
		Goal: "Test plan",
		Steps: []*Step{
			{
				Description: "Test step",
				Action:      &Action{Type: "chat"},
				Status:      StepPending,
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := loop.ExecutePlan(ctx, plan)
	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestExpandedLoopStepFailure(t *testing.T) {
	executor := &mockExecutor{
		results: []*StepResult{
			{Success: false, Error: fmt.Errorf("step failed")},
		},
	}

	loop := NewExpandedLoop(
		NewSimplePlanner(),
		NewSimpleVerifier(),
		executor,
	)

	plan := &Plan{
		Goal: "Test plan",
		Steps: []*Step{
			{
				Description: "Test step",
				Action:      &Action{Type: "chat"},
				Status:      StepPending,
			},
		},
	}

	ctx := context.Background()
	_, err := loop.ExecutePlan(ctx, plan)
	if err == nil {
		t.Error("expected error for failed step")
	}
}

func TestPlanResultDuration(t *testing.T) {
	started := time.Now()
	finished := started.Add(5 * time.Second)
	result := &PlanResult{
		Started:  started,
		Finished: finished,
	}

	duration := result.Duration()
	if duration < 4*time.Second || duration > 6*time.Second {
		t.Errorf("duration = %v, want ~5s", duration)
	}
}
