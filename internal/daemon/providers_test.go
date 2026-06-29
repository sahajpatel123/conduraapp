package daemon

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/anomaly"
	"github.com/sahajpatel123/conduraapp/internal/api_key"
	"github.com/sahajpatel123/conduraapp/internal/config"
	"github.com/sahajpatel123/conduraapp/internal/llm"
	"github.com/sahajpatel123/conduraapp/internal/storage"
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
	n := buildProvidersFromConfig(log, registry, cfg, akm, nil, nil)
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
	n := buildProvidersFromConfig(log, registry, cfg, akm, nil, nil)
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
	n := buildProvidersFromConfig(log, registry, cfg, akm, nil, nil)
	if n != 0 {
		t.Errorf("fresh install should register 0 providers, got %d", n)
	}
}

// TestWrapProvidersWithRecorder_PinsP0_2 pins the wiring for
// CLAUDE.md §5.6's fifth "agent went insane" trigger: an outbound
// HTTP request from a registered provider's transport must record
// the host so a pivot to a new host trips the new-endpoint detector.
//
// This is the regression test for P0-2 of the 2026-06-29 audit:
// RecordNetwork was previously defined but had no production call
// sites. The test proves the daemon's wrapProvidersWithRecorder
// actually attaches a *anomaly.RecordingTransport to a registered
// provider's HTTP client.
func TestWrapProvidersWithRecorder_PinsP0_2(t *testing.T) {
	log := newTestLogger()
	db := newTestStorage(t)
	ctx := context.Background()
	akm := api_key.New(db, nil)
	// Seed an Anthropic API key so buildProvidersFromConfig actually
	// registers an Anthropic provider.
	if _, err := akm.Set(ctx, api_key.Key{
		Provider: "anthropic", Label: "anthropic-p02-test", AuthKind: api_key.AuthAPIKey,
		Secret: "sk-test-1234567890",
	}); err != nil {
		t.Fatalf("api_key.Set: %v", err)
	}
	cfg := config.Default()
	registry := llm.NewRegistry()
	buildProvidersFromConfig(log, registry, cfg, akm, nil, nil)
	if registry.Len() == 0 {
		t.Fatal("expected at least one provider registered")
	}

	det := anomaly.NewDetector(func(_ anomaly.Trip) {})
	defer det.Close()

	wrapProvidersWithRecorder(registry, det)

	// Every registered provider must now have a *anomaly.RecordingTransport
	// wrapped around its HTTP client. We assert by type-asserting the
	// transport on each provider's HTTP client.
	type clientBearer interface {
		GetHTTPClient() *http.Client
	}
	wrapped := 0
	for _, prov := range registry.List() {
		b, ok := prov.(clientBearer)
		if !ok {
			continue
		}
		hc := b.GetHTTPClient()
		if hc == nil {
			continue
		}
		if _, ok := hc.Transport.(*anomaly.RecordingTransport); ok {
			wrapped++
		}
	}
	if wrapped == 0 {
		t.Fatal("no provider got a *anomaly.RecordingTransport — P0-2 wiring is broken")
	}

	// Idempotent: a second call must not wrap twice (would record
	// the same host twice and would wrap a *RecordingTransport in
	// another *RecordingTransport, which is a no-op functionally but
	// shows intent: the wiring is safe to call repeatedly).
	wrapProvidersWithRecorder(registry, det)
	wrapped = 0
	for _, prov := range registry.List() {
		b, ok := prov.(clientBearer)
		if !ok {
			continue
		}
		hc := b.GetHTTPClient()
		if hc == nil {
			continue
		}
		if _, ok := hc.Transport.(*anomaly.RecordingTransport); ok {
			wrapped++
		}
	}
	if wrapped == 0 {
		t.Fatal("idempotency broke: a second wrapProvidersWithRecorder call removed the recorder")
	}
}
