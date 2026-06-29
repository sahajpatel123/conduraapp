package agent

import (
	"context"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/blastradius"
	"github.com/sahajpatel123/conduraapp/internal/gatekeeper"
	"github.com/sahajpatel123/conduraapp/internal/llm"
	"github.com/sahajpatel123/conduraapp/internal/sse"
	"github.com/sahajpatel123/conduraapp/internal/stream"
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

type mockLLMProvider struct {
	name string
}

func (p *mockLLMProvider) Name() string { return p.name }
func (p *mockLLMProvider) Chat(_ context.Context, _ llm.ChatRequest) (llm.ChatResponse, error) {
	return llm.ChatResponse{}, nil
}
func (p *mockLLMProvider) Stream(_ context.Context, _ llm.ChatRequest) (<-chan llm.StreamEvent, func(), error) {
	ch := make(chan llm.StreamEvent, 10)
	cancel := func() {}
	go func() {
		defer close(ch)
		ch <- llm.StreamEvent{Delta: llm.Message{Content: "Hello"}}
		ch <- llm.StreamEvent{Delta: llm.Message{Content: ", "}}
		ch <- llm.StreamEvent{Delta: llm.Message{Content: "voice!"}}
		ch <- llm.StreamEvent{Done: true}
	}()
	return ch, cancel, nil
}
func (p *mockLLMProvider) Models() []llm.ModelInfo      { return nil }
func (p *mockLLMProvider) DefaultModel(_ string) string { return "" }

func TestLoop_Ask_Allowed(t *testing.T) {
	broker := sse.NewBroker()
	defer broker.Close()

	reg := llm.NewRegistry()
	reg.Register(&mockLLMProvider{name: "test"})

	mgr := stream.NewManager(broker, reg)
	defer mgr.Close()

	loop := &Loop{
		Gatekeeper:   &mockGatekeeper{},
		Stream:       mgr,
		Broker:       broker,
		ProviderName: "test",
		Model:        "test-model",
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
	if result.Reply != "Hello, voice!" {
		t.Errorf("reply = %q, want %q", result.Reply, "Hello, voice!")
	}
	if result.RequestID == "" {
		t.Error("expected non-empty stream request_id")
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

func TestLoop_Ask_MissingStream(t *testing.T) {
	loop := &Loop{
		Gatekeeper: &mockGatekeeper{},
	}

	_, err := loop.Ask(context.Background(), AskRequest{Text: "hello"})
	if err == nil {
		t.Fatal("expected error when stream manager is nil")
	}
}

func TestLoop_Cancel(t *testing.T) {
	loop := &Loop{
		Gatekeeper: &mockGatekeeper{},
	}
	// Should not panic.
	loop.Cancel("req-3")
}
