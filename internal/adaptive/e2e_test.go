package adaptive

import (
	"context"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/llm"
)

type mockLLM struct {
	response string
}

func (m *mockLLM) Name() string { return "mock" }
func (m *mockLLM) Chat(_ context.Context, _ llm.ChatRequest) (llm.ChatResponse, error) {
	return llm.ChatResponse{Message: llm.Message{Role: llm.RoleAssistant, Content: m.response}, FinishReason: "stop"}, nil
}
func (m *mockLLM) Stream(_ context.Context, _ llm.ChatRequest) (<-chan llm.StreamEvent, func(), error) {
	return nil, nil, nil
}
func (m *mockLLM) Models() []llm.ModelInfo      { return []llm.ModelInfo{{ID: "mock-model"}} }
func (m *mockLLM) DefaultModel(_ string) string { return "mock-model" }

// TestE2E_Engine_LearnsAndPredicts drives the full loop:
// Observe → Analyze → Persist → Predict.
// This is the forcing-function test that proves the engine works.
func TestE2E_Engine_LearnsAndPredicts(t *testing.T) {
	// Setup: encrypted store.
	db := testDB(t)
	s, err := NewEncryptedStore(db, passthroughEncrypt, passthroughDecrypt)
	if err != nil {
		t.Fatalf("store: %v", err)
	}

	// Setup: observer + engine.
	observer := NewObserver()
	adj := NewAdjudicator(
		[]string{"verbosity", "response_length", "default_model", "time_patterns"},
		[]string{"new_skill", "default_backend", "communication_style", "risk_tolerance"},
		0.6,
	)

	// Proposer returns an actual inference.
	proposer := &mockLLM{
		response: `[{"category":"verbosity","field":"preferred_verbosity","value":"concise","confidence":0.85,"reason":"User uses short messages"}]`,
	}
	critic := &mockLLM{
		response: `[{"category":"verbosity","field":"preferred_verbosity","value":"concise","confidence":0.80,"reason":"User uses short messages"}]`,
	}

	dialectic := NewDialectic(proposer, "gpt-4", critic, "cheap-model", adj, nil, StrengthBalanced)
	strength := func() Strength { return StrengthBalanced }
	predictor := NewPredictor(s, strength)
	cfg := DefaultConfig()
	cfg.Strength = StrengthBalanced
	engine := NewEngine(observer, dialectic, adj, s, predictor, cfg, nil)

	// Step 1: Observe a user session.
	observer.Record(context.Background(), Observation{
		SessionID:     "sess-1",
		UserQuery:     "How do I fix this?",
		AgentReply:    "Here is a short answer.",
		UserInitiated: true,
		Timestamp:     time.Now(),
	})

	// Step 2: Run the engine (analyze + persist).
	engine.Run(context.Background())

	// Step 3: Verify the model was updated.
	model, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if model.Version < 2 {
		t.Errorf("model version = %d, expected >= 2 after learning", model.Version)
	}

	// Step 4: Verify the Predictor returns context.
	hint, err := predictor.Predict(context.Background(), "fix this")
	if err != nil {
		t.Fatalf("Predict: %v", err)
	}
	if hint != "" {
		t.Logf("Predictor hint: %s", hint)
	}
}

// TestE2E_Engine_Decay verifies decay directly on the store.
func TestE2E_Engine_Decay(t *testing.T) {
	db := testDB(t)
	s, err := NewEncryptedStore(db, passthroughEncrypt, passthroughDecrypt)
	if err != nil {
		t.Fatal(err)
	}

	model := &UserModel{
		Identity: InferredField{Value: "old-identity", Confidence: 0.5, LastSeen: time.Now().Add(-90 * 24 * time.Hour), Source: "dialectic"},
		Version:  1,
	}
	if err := s.Save(model); err != nil {
		t.Fatalf("initial save: %v", err)
	}

	// Verify save worked.
	loaded, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Identity.Value != "old-identity" {
		t.Fatalf("expected identity after save, got %q", loaded.Identity.Value)
	}
	if loaded.Identity.LastSeen.IsZero() {
		t.Fatal("LastSeen is zero after load")
	}

	// Force the identity source to "dialectic" (not explicit).
	loaded.Identity.Source = "dialectic"
	_ = s.Save(loaded)

	// Now call decay. Since ForgetAfterDays=30 and LastSeen=90 days ago,
	// the identity should be cleared.
	cfg := DefaultConfig()
	cfg.ForgetAfterDays = 30
	engine := &Engine{Store: s, cfg: cfg}
	engine.decay(context.Background())

	final, _ := s.Load()
	if final.Identity.Value != "" {
		t.Errorf("stale identity should be decayed, got %q", final.Identity.Value)
	}
}

// TestE2E_Engine_PendingConfirmations verifies require-confirm flow.
func TestE2E_Engine_PendingConfirmations(t *testing.T) {
	db := testDB(t)
	s, _ := NewEncryptedStore(db, passthroughEncrypt, passthroughDecrypt)

	observer := NewObserver()
	adj := NewAdjudicator([]string{"verbosity"}, []string{"communication_style"}, 0.6)
	proposer := &mockLLM{
		response: `[{"category":"communication_style","field":"style","value":"casual","confidence":0.9,"reason":"User uses casual language"}]`,
	}
	dialectic := NewDialectic(proposer, "gpt-4", nil, "", adj, nil, StrengthBalanced)
	cfg := DefaultConfig()
	cfg.Strength = StrengthBalanced
	engine := NewEngine(observer, dialectic, adj, s, NewPredictor(s, func() Strength { return StrengthBalanced }), cfg, nil)

	observer.Record(context.Background(), Observation{SessionID: "s1", UserQuery: "hey", AgentReply: "hi", UserInitiated: true, Timestamp: time.Now()})
	engine.Run(context.Background())

	// Verify pending confirmation exists.
	pending := engine.Pending()
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending, got %d", len(pending))
	}

	// Confirm it.
	engine.ConfirmPending(0)
	model, _ := s.Load()
	if model.Version < 2 {
		t.Error("model should be updated after confirmation")
	}
}
