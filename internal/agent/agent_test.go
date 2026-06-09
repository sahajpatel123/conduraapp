package agent

import (
	"context"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
)

// mockGatekeeper is a test double that allows all actions.
type mockGatekeeper struct{}

func (m *mockGatekeeper) Evaluate(_ context.Context, _ blastradius.Action) (gatekeeper.Decision, string) {
	return gatekeeper.Allow, "allowed"
}

// denyGatekeeper is a test double that denies all actions.
type denyGatekeeper struct{}

func (m *denyGatekeeper) Evaluate(_ context.Context, _ blastradius.Action) (gatekeeper.Decision, string) {
	return gatekeeper.Deny, "denied by test"
}

func TestLoop_Ask_Allowed(t *testing.T) {
	loop := &Loop{
		Gatekeeper: &mockGatekeeper{},
	}

	result, err := loop.Ask(context.Background(), AskRequest{
		Text:      "hello",
		RequestID: "req-1",
	})
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if result.Finish != "stop" {
		t.Errorf("expected finish=stop, got %q", result.Finish)
	}
	if result.RequestID != "req-1" {
		t.Errorf("expected request_id=req-1, got %q", result.RequestID)
	}
}

func TestLoop_Ask_Denied(t *testing.T) {
	loop := &Loop{
		Gatekeeper: &denyGatekeeper{},
	}

	result, err := loop.Ask(context.Background(), AskRequest{
		Text:      "hack the planet",
		RequestID: "req-2",
	})
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if result.Finish != "blocked" {
		t.Errorf("expected finish=blocked, got %q", result.Finish)
	}
}

func TestLoop_Cancel(t *testing.T) {
	loop := &Loop{
		Gatekeeper: &mockGatekeeper{},
	}
	// Should not panic.
	loop.Cancel("req-3")
}
