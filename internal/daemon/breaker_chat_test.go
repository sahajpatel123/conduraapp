package daemon

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sahajpatel123/conduraapp/internal/audit"
	"github.com/sahajpatel123/conduraapp/internal/failover"
	"github.com/sahajpatel123/conduraapp/internal/halt"
	"github.com/sahajpatel123/conduraapp/internal/ipc"
	"github.com/sahajpatel123/conduraapp/internal/llm"
	"github.com/sahajpatel123/conduraapp/internal/storage"
)

// TestLLMChat_CircuitBreaker_OpenBlocksCall verifies that when a
// provider's circuit breaker is open, llm.chat fails fast instead
// of making the call. Before the fix, the breaker was never checked
// on the chat path.
func TestLLMChat_CircuitBreaker_OpenBlocksCall(t *testing.T) {
	registry := llm.NewRegistry()
	registry.Register(&fakeChatProvider{})

	breakers := failover.NewBreakerRegistry(3, 30*time.Second)
	mon := failover.NewSpendMonitor(failover.SpendCap{USDPerDay: 100})

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := storage.Open(context.Background(), storage.Config{Path: dbPath, MasterKey: ""})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	haltFlag := halt.New(db.SQL())
	auditLog := audit.New(db.SQL(), db.MasterKey())

	// Open the breaker by recording 3 failures.
	b := breakers.For("test-provider")
	for i := 0; i < 3; i++ {
		b.RecordFailure()
	}
	require.Equal(t, failover.CircuitOpen, b.State())

	srv := ipc.NewServer()
	registerLLMMethods(srv, registry, mon, breakers, haltFlag, auditLog)

	params, _ := json.Marshal(map[string]any{
		"provider": "test-provider",
		"model":    "test-model",
		"request":  map[string]any{"messages": []map[string]any{{"role": "user", "content": "hi"}}},
	})

	resp, err := srv.Handle(context.Background(), &ipc.Request{
		JSONRPC: "2.0",
		Method:  "llm.chat",
		Params:  params,
		ID:      json.RawMessage("1"),
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Error, "expected error when breaker is open")
	assert.Contains(t, resp.Error.Message, "circuit breaker open")
}

// fakeChatProvider is a minimal llm.Provider for testing the chat
// path without making real API calls.
type fakeChatProvider struct{}

func (f *fakeChatProvider) Name() string { return "test-provider" }
func (f *fakeChatProvider) Chat(_ context.Context, req llm.ChatRequest) (llm.ChatResponse, error) {
	return llm.ChatResponse{
		Model:   req.Model,
		Message: llm.Message{Role: "assistant", Content: "ok"},
		Usage:   llm.Usage{InputTokens: 10, OutputTokens: 5, TotalTokens: 15},
	}, nil
}
func (f *fakeChatProvider) Stream(_ context.Context, _ llm.ChatRequest) (<-chan llm.StreamEvent, func(), error) {
	ch := make(chan llm.StreamEvent)
	close(ch)
	return ch, func() {}, nil
}
func (f *fakeChatProvider) DefaultModel(_ string) string { return "test-model" }
func (f *fakeChatProvider) Models() []llm.ModelInfo {
	return []llm.ModelInfo{{ID: "test-model", ContextWindow: 4096}}
}
