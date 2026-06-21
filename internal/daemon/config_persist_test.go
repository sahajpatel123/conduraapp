package daemon

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/config"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
	"github.com/sahajpatel123/synapticapp/internal/telemetry"
)

// callRPC invokes a method on the server and returns the
// unmarshaled result or the JSON-RPC error.
func configPersistCallRPC(t *testing.T, srv *ipc.Server, method string, params json.RawMessage) (any, error) {
	t.Helper()
	resp, err := srv.Handle(context.Background(), &ipc.Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      json.RawMessage("1"),
	})
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Result, nil
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

	_, err = configPersistCallRPC(t, srv, "config.update", params)
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
		cfg:      cfg,
		Loader:   loader,
		Telemetry: telemetry.New(nil, ""),
	}

	registerControlMethods(srv, cfg, subs)

	params, _ := json.Marshal(map[string]any{"enabled": true})
	_, err = configPersistCallRPC(t, srv, "telemetry.setEnabled", params)
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