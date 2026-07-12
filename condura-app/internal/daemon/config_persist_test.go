package daemon

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/config"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/failover"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/ipc"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/telemetry"
)

// configPersistCallRPC invokes a method on the server and returns
// the JSON-RPC error (if any). The result is discarded — the tests
// only check side effects on disk.
func configPersistCallRPC(t *testing.T, srv *ipc.Server, method string, params json.RawMessage) error {
	t.Helper()
	resp, err := srv.Handle(context.Background(), &ipc.Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      json.RawMessage("1"),
	})
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

// TestConfigUpdate_PersistsToDisk verifies that config.update writes
// the patched config back to disk via Loader.Save(). Before the fix,
// the handler only mutated the in-memory cfg struct — hotkey,
// telemetry, and window changes were lost on daemon restart.
func TestConfigUpdate_PersistsToDisk(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	loader := config.NewLoader(cfgPath)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	srv := ipc.NewServer()
	subs := &Subsystems{
		cfg:    cfg,
		Loader: loader,
	}

	registerControlMethods(srv, cfg, subs)

	// Patch the hotkey via config.update.
	patch := map[string]json.RawMessage{
		"hotkey": json.RawMessage(`{"overlay":"Cmd+Shift+Space"}`),
	}
	params, _ := json.Marshal(patch)

	err = configPersistCallRPC(t, srv, "config.update", params)
	if err != nil {
		t.Fatalf("config.update call: %v", err)
	}

	// Verify in-memory config was patched.
	if cfg.Hotkey.Overlay != "Cmd+Shift+Space" {
		t.Errorf("in-memory hotkey overlay = %q, want %q", cfg.Hotkey.Overlay, "Cmd+Shift+Space")
	}

	// Verify the file on disk contains the patched value.
	loader2 := config.NewLoader(cfgPath)
	cfg2, err := loader2.Load()
	if err != nil {
		t.Fatalf("re-load config: %v", err)
	}
	if cfg2.Hotkey.Overlay != "Cmd+Shift+Space" {
		t.Errorf("persisted hotkey overlay = %q, want %q", cfg2.Hotkey.Overlay, "Cmd+Shift+Space")
	}
}

// TestConfigUpdate_AutonomyPersistsAndRewires verifies Meridian Settings
// can save autonomy.default_level + per_task and the live matrix updates.
func TestConfigUpdate_AutonomyPersistsAndRewires(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	loader := config.NewLoader(cfgPath)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Build a real safety layer so rewireAutonomy has an engine to bind.
	safety := buildSafetyLayer(nil, nil, nil, cfg, nil)
	srv := ipc.NewServer()
	subs := &Subsystems{
		cfg:    cfg,
		Loader: loader,
		Safety: safety,
	}
	registerControlMethods(srv, cfg, subs)

	patch := map[string]json.RawMessage{
		"autonomy": json.RawMessage(`{
			"default_level":"supervised",
			"per_task":{"coding":"autonomous","shell_commands":"block"}
		}`),
	}
	params, _ := json.Marshal(patch)
	if err := configPersistCallRPC(t, srv, "config.update", params); err != nil {
		t.Fatalf("config.update: %v", err)
	}

	if cfg.Autonomy.DefaultLevel != "supervised" {
		t.Errorf("default_level = %q, want supervised", cfg.Autonomy.DefaultLevel)
	}
	if cfg.Autonomy.PerTask["coding"] != "autonomous" {
		t.Errorf("per_task.coding = %q, want autonomous", cfg.Autonomy.PerTask["coding"])
	}
	if cfg.Autonomy.PerTask["shell_commands"] != "block" {
		t.Errorf("per_task.shell_commands = %q, want block", cfg.Autonomy.PerTask["shell_commands"])
	}

	// Live matrix: coding.* should be autonomous after rewire.
	if safety.Autonomy == nil {
		t.Fatal("Safety.Autonomy is nil after rewire")
	}
	if got := safety.Autonomy.Evaluate("coding", "any"); got.String() != "autonomous" {
		t.Errorf("live matrix coding = %s, want autonomous", got)
	}
	if got := safety.Autonomy.Evaluate("shell_commands", "any"); got.String() != "block" {
		t.Errorf("live matrix shell = %s, want block", got)
	}

	// Persist round-trip.
	loader2 := config.NewLoader(cfgPath)
	cfg2, err := loader2.Load()
	if err != nil {
		t.Fatalf("re-load: %v", err)
	}
	if cfg2.Autonomy.DefaultLevel != "supervised" {
		t.Errorf("persisted default_level = %q, want supervised", cfg2.Autonomy.DefaultLevel)
	}
}

// TestConfigUpdate_SecuritySpendCapLive verifies spend_limit patches
// update the in-process SpendMonitor.
func TestConfigUpdate_SecuritySpendCapLive(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	loader := config.NewLoader(cfgPath)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	mon := failover.NewSpendMonitor(failover.SpendCap{USDPerDay: 5.0})
	srv := ipc.NewServer()
	subs := &Subsystems{cfg: cfg, Loader: loader, Spend: mon}
	registerControlMethods(srv, cfg, subs)

	params, _ := json.Marshal(map[string]json.RawMessage{
		"security": json.RawMessage(`{"spend_limit_usd_per_day":12.5}`),
	})
	if err := configPersistCallRPC(t, srv, "config.update", params); err != nil {
		t.Fatalf("config.update: %v", err)
	}
	if cfg.Security.SpendLimitUSDPerDay != 12.5 {
		t.Errorf("cfg spend = %v, want 12.5", cfg.Security.SpendLimitUSDPerDay)
	}
	if mon.Cap().USDPerDay != 12.5 {
		t.Errorf("live spend cap = %v, want 12.5", mon.Cap().USDPerDay)
	}
}

// TestConfigGet_SnakeCaseAutonomy ensures config.get exposes YAML keys
// the Meridian GUI expects (not Go PascalCase field names).
func TestConfigGet_SnakeCaseAutonomy(t *testing.T) {
	cfg := config.Default()
	cfg.Autonomy.DefaultLevel = "supervised"
	cfg.Autonomy.PerTask = map[string]string{"coding": "autonomous"}
	cfg.LLM.Providers = map[string]config.ProviderConfig{
		"openai": {Enabled: true, APIKey: "sk-secret-should-not-leak", DefaultModel: "gpt-4o"},
	}

	view, err := publicConfigView(cfg)
	if err != nil {
		t.Fatalf("publicConfigView: %v", err)
	}
	aut, ok := view["autonomy"].(map[string]any)
	if !ok {
		t.Fatalf("autonomy missing or wrong type: %#v", view["autonomy"])
	}
	if aut["default_level"] != "supervised" {
		t.Errorf("default_level = %v", aut["default_level"])
	}
	perTask, _ := aut["per_task"].(map[string]any)
	if perTask["coding"] != "autonomous" {
		t.Errorf("per_task.coding = %v", perTask["coding"])
	}
	llm, _ := view["llm"].(map[string]any)
	provs, _ := llm["providers"].(map[string]any)
	openai, _ := provs["openai"].(map[string]any)
	if openai["api_key"] != "" && openai["api_key"] != nil {
		t.Errorf("api_key leaked: %v", openai["api_key"])
	}
	if openai["has_api_key"] != true {
		t.Errorf("has_api_key = %v, want true", openai["has_api_key"])
	}
}

// TestConfigUpdate_RejectsInvalidAutonomy ensures bad levels do not persist.
func TestConfigUpdate_RejectsInvalidAutonomy(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	loader := config.NewLoader(cfgPath)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	before := cfg.Autonomy.DefaultLevel
	srv := ipc.NewServer()
	subs := &Subsystems{cfg: cfg, Loader: loader}
	registerControlMethods(srv, cfg, subs)

	params, _ := json.Marshal(map[string]json.RawMessage{
		"autonomy": json.RawMessage(`{"default_level":"yolo"}`),
	})
	err = configPersistCallRPC(t, srv, "config.update", params)
	if err == nil {
		t.Fatal("expected error for invalid autonomy level")
	}
	if cfg.Autonomy.DefaultLevel != before {
		t.Errorf("default_level mutated to %q on reject", cfg.Autonomy.DefaultLevel)
	}
}

// TestTelemetrySetEnabled_PersistsToDisk verifies that
// telemetry.setEnabled writes the change to disk.
func TestTelemetrySetEnabled_PersistsToDisk(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	loader := config.NewLoader(cfgPath)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Default is false; flip to true.
	cfg.Telemetry.Enabled = false

	srv := ipc.NewServer()
	subs := &Subsystems{
		cfg:       cfg,
		Loader:    loader,
		Telemetry: telemetry.New(nil, ""),
	}

	registerControlMethods(srv, cfg, subs)

	params, _ := json.Marshal(map[string]any{"enabled": true})
	err = configPersistCallRPC(t, srv, "telemetry.setEnabled", params)
	if err != nil {
		t.Fatalf("telemetry.setEnabled call: %v", err)
	}

	if !cfg.Telemetry.Enabled {
		t.Errorf("in-memory telemetry.Enabled = false, want true")
	}

	// Verify persisted to disk.
	loader2 := config.NewLoader(cfgPath)
	cfg2, err := loader2.Load()
	if err != nil {
		t.Fatalf("re-load config: %v", err)
	}
	if !cfg2.Telemetry.Enabled {
		t.Errorf("persisted telemetry.Enabled = false, want true")
	}
}
