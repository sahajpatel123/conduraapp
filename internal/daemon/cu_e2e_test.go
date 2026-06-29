package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/agent"
	"github.com/sahajpatel123/conduraapp/internal/computeruse"
)

// TestE2E_CUAction_CULoopThroughResolver verifies the full loop:
// Planner → CUResolver → Gatekeeper → Backend.
// This exposed the bug where SimplePlanner emits "chat" verbs
// that CUResolver rejects.
func TestE2E_CUAction_CULoopThroughResolver(t *testing.T) {
	backend := &computeruse.MockBackend{
		BackendName: "orax", Available: true,
		Caps: []computeruse.Capability{computeruse.CapClick, computeruse.CapType, computeruse.CapLaunch},
	}
	cu := computeruse.New(backend)
	// Allow-all gatekeeper so actions pass through.
	gate := allowAllGatekeeper{}
	gated := computeruse.NewGatedExecutor(cu, gate)
	r := NewCUResolver(cu, gated)

	planner := cuTestPlanner{
		plan: &agent.Plan{
			Goal: "click submit",
			Steps: []*agent.Step{
				{Action: &agent.Action{Type: "click", Target: "Submit button"}, Status: agent.StepPending},
			},
		},
	}

	v := agent.NewSimpleVerifier()
	halt := cuTestHalt{}
	loop := agent.NewCULoop(planner, v, r, halt)
	loop.StepTimeout = 5 * time.Second

	result, err := loop.Run(context.Background(), "click submit", nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
	if len(result.Steps) != 1 {
		t.Fatalf("got %d steps", len(result.Steps))
	}
}

// TestE2E_CUAction_SimplePlannerEmitsChat exposes the exact bug.
func TestE2E_CUAction_SimplePlannerEmitsChat(t *testing.T) {
	backend := &computeruse.MockBackend{BackendName: "orax", Available: true, Caps: []computeruse.Capability{computeruse.CapClick}}
	cu := computeruse.New(backend)
	// Allow-all gatekeeper so actions pass through.
	gate := allowAllGatekeeper{}
	gated := computeruse.NewGatedExecutor(cu, gate)
	r := NewCUResolver(cu, gated)

	p := agent.NewSimplePlanner()
	loop := agent.NewCULoop(p, agent.NewSimpleVerifier(), r, cuTestHalt{})
	loop.StepTimeout = 5 * time.Second

	_, err := loop.Run(context.Background(), "click", nil)
	if err == nil {
		t.Fatal("expected error: SimplePlanner emits 'chat' verb, CUResolver rejects it")
	}
}

// TestE2E_CUAction_HaltInterrupts verifies kill-switch mid-plan.
func TestE2E_CUAction_HaltInterrupts(t *testing.T) {
	backend := &computeruse.MockBackend{BackendName: "orax", Available: true, Caps: []computeruse.Capability{computeruse.CapClick}}
	cu := computeruse.New(backend)
	// Allow-all gatekeeper so actions pass through.
	gate := allowAllGatekeeper{}
	gated := computeruse.NewGatedExecutor(cu, gate)
	r := NewCUResolver(cu, gated)

	planner := cuTestPlanner{
		plan: &agent.Plan{
			Goal:  "do thing",
			Steps: []*agent.Step{{Action: &agent.Action{Type: "click", Target: "button"}, Status: agent.StepPending}},
		},
	}
	loop := agent.NewCULoop(planner, agent.NewSimpleVerifier(), r, cuTestHalt{halted: true})
	loop.StepTimeout = 5 * time.Second

	_, err := loop.Run(context.Background(), "click", nil)
	if err == nil {
		t.Fatal("expected halt error")
	}
}

// TestE2E_CUAction_MultiStep executes a 3-step plan.
func TestE2E_CUAction_MultiStep(t *testing.T) {
	backend := &computeruse.MockBackend{BackendName: "orax", Available: true,
		Caps: []computeruse.Capability{computeruse.CapClick, computeruse.CapType}}
	cu := computeruse.New(backend)
	// Allow-all gatekeeper so actions pass through.
	gate := allowAllGatekeeper{}
	gated := computeruse.NewGatedExecutor(cu, gate)
	r := NewCUResolver(cu, gated)

	planner := cuTestPlanner{
		plan: &agent.Plan{
			Goal: "search",
			Steps: []*agent.Step{
				{Action: &agent.Action{Type: "click", Target: "field"}, Status: agent.StepPending},
				{Action: &agent.Action{Type: "type", Value: "hello"}, Status: agent.StepPending, DependsOn: []int{0}},
				{Action: &agent.Action{Type: "click", Target: "btn"}, Status: agent.StepPending, DependsOn: []int{1}},
			},
		},
	}
	loop := agent.NewCULoop(planner, agent.NewSimpleVerifier(), r, cuTestHalt{})
	loop.StepTimeout = 5 * time.Second

	result, err := loop.Run(context.Background(), "search for hello", nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.Steps) != 3 {
		t.Fatalf("got %d steps", len(result.Steps))
	}
}

// TestE2E_CUAction_VerbAcceptance validates all accepted/rejected verbs.
func TestE2E_CUAction_VerbAcceptance(t *testing.T) {
	accept := []string{"click", "type", "scroll", "key_press", "keypress", "drag", "launch", "focus", "wait"}
	for _, v := range accept {
		if _, err := parseActionType(v); err != nil {
			t.Errorf("%q should be accepted: %v", v, err)
		}
	}
	reject := []string{"chat", "open", "navigate", "read", "write", "talk", ""}
	for _, v := range reject {
		if _, err := parseActionType(v); err == nil {
			t.Errorf("%q should be rejected", v)
		}
	}
}

type cuTestPlanner struct {
	plan *agent.Plan
	err  error
}

func (p cuTestPlanner) Decompose(_ context.Context, _ string, _ *agent.Context) (*agent.Plan, error) {
	return p.plan, p.err
}
func (p cuTestPlanner) Reprioritize(_ context.Context, plan *agent.Plan, _ *agent.Observation) (*agent.Plan, error) {
	return plan, nil
}

type cuTestHalt struct{ halted bool }

func (h cuTestHalt) IsHalted() bool { return h.halted }
