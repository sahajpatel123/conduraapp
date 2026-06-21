package daemon

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/api_key"
	"github.com/sahajpatel123/synapticapp/internal/config"
	"github.com/sahajpatel123/synapticapp/internal/llm"
	"github.com/sahajpatel123/synapticapp/internal/storage"
)

// newTestLogger returns a slog.Logger that swallows output during tests.
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newTestStorage creates a temporary SQLite-backed storage.DB for tests.
// It uses an ephemeral master key (Secrets=nil) so each test is
// isolated. The returned DB is closed via t.Cleanup.
func newTestStorage(t *testing.T) *storage.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "synaptic.db")
	db, err := storage.Open(context.Background(), storage.Config{
		Path:      dbPath,
		MasterKey: "",
		Secrets:   nil, // ephemeral mode for tests
	})
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestBuildProvidersFromConfig_AutoEnablesFromStoredKey is the
// Phase 17 Fix #4 (B1) regression test. Before the fix,
// buildProvidersFromConfig only iterated cfg.LLM.Providers, so a
// provider that the user added an API key for (via apikeys.set)
// without first enabling it in YAML never made it into the LLM
// registry. After the fix, buildProvidersFromConfig scans the
// api_keys table and auto-flips Enabled=true for any provider with
// at least one stored key. This test asserts that an "anthropic"
// key in the table makes cfg.LLM.Providers["anthropic"].Enabled
// true even though the YAML map had it disabled.
func TestBuildProvidersFromConfig_AutoEnablesFromStoredKey(t *testing.T) {
	log := newTestLogger()
	db := newTestStorage(t)
	ctx := context.Background()

	akm := api_key.New(db, nil)
	if _, err := akm.Set(ctx, api_key.Key{
		Provider: "anthropic", Label: "anthropic-test", AuthKind: api_key.AuthAPIKey,
		Secret: "sk-test-1234567890",
	}); err != nil {
		t.Fatalf("api_key.Set: %v", err)
	}

	cfg := config.Default()
	// defaultProviders() seeds "anthropic" with Enabled:false.
	if cfg.LLM.Providers["anthropic"].Enabled {
		t.Fatal("precondition: default anthropic should be disabled")
	}

	registry := llm.NewRegistry()
	n := buildProvidersFromConfig(log, registry, cfg, akm, nil)
	if n < 1 {
		t.Errorf("expected at least 1 provider registered, got %d", n)
	}
	if !cfg.LLM.Providers["anthropic"].Enabled {
		t.Error("cfg.LLM.Providers[anthropic].Enabled should be auto-flipped to true")
	}
	if _, ok := registry.Get("anthropic"); !ok {
		t.Error("anthropic provider should be in registry after auto-enable")
	}
}

// TestBuildProvidersFromConfig_KeylessLocalProvider ensures the
// Ollama keyless bypass still works after Fix #4 (B1).
func TestBuildProvidersFromConfig_KeylessLocalProvider(t *testing.T) {
	log := newTestLogger()
	db := newTestStorage(t)
	akm := api_key.New(db, nil)
	cfg := config.Default()
	cfg.LLM.Providers[config.ProviderOllama] = config.ProviderConfig{
		Enabled: true, BaseURL: "http://127.0.0.1:11434/v1", DefaultModel: "llama4",
	}
	registry := llm.NewRegistry()
	n := buildProvidersFromConfig(log, registry, cfg, akm, nil)
	if n < 1 {
		t.Errorf("expected ollama to be registered (keyless), got %d", n)
	}
	if _, ok := registry.Get(config.ProviderOllama); !ok {
		t.Error("ollama provider should be in registry (keyless)")
	}
}

// TestBuildProvidersFromConfig_NoKeysNoProviders ensures that with
// no keys in the api_keys table AND a disabled map, no providers
// are registered. This is the original behavior for a fresh install
// before the user has added any keys — should be unchanged.
func TestBuildProvidersFromConfig_NoKeysNoProviders(t *testing.T) {
	log := newTestLogger()
	db := newTestStorage(t)
	akm := api_key.New(db, nil)
	cfg := config.Default() // all providers disabled
	registry := llm.NewRegistry()
	n := buildProvidersFromConfig(log, registry, cfg, akm, nil)
	if n != 0 {
		t.Errorf("fresh install should register 0 providers, got %d", n)
	}
}
