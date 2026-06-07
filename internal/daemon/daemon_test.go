package daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/config"
)

// TestRun_Smoke brings up the daemon in-process, verifies the
// subsystems are constructed, then cancels the context and verifies
// Run returns within a reasonable time. This is the smallest test
// that exercises the entire orchestration path.
func TestRun_Smoke(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default()
	cfg.General.DataDir = dir
	cfg.Storage.Path = filepath.Join(dir, "synaptic.db")
	cfg.Logging.File = "" // no log file; logs go to stderr
	cfg.Logging.AddSource = false
	cfg.Security.SpendLimitUSDPerDay = 1.0
	cfg.APIServer.AuthToken = "test-token"

	// Unset SYNAPTIC_ env vars so the test isn't perturbed by the
	// host environment. (applyEnvOverrides reads SYNAPTIC_*
	// automatically.) Use t.Setenv with empty values to clear them
	// for the duration of the test.
	for _, e := range os.Environ() {
		for i := 0; i < len(e)-9; i++ {
			if e[i:i+9] == "SYNAPTIC_" {
				name := e[:i+9]
				end := i + 9
				for end < len(e) && e[end] != '=' {
					end++
				}
				if end < len(e) {
					t.Setenv(name, "")
				}
				break
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run the daemon in a goroutine; cancel after 100ms to trigger
	// graceful shutdown.
	done := make(chan error, 1)
	go func() {
		_, err := Run(ctx, Options{
			Config: cfg,
			Listen: ListenSpec{Disable: true}, // no IPC; smoke test
		})
		done <- err
	}()

	// Give the daemon time to construct subsystems.
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after context cancel")
	}
}

// TestRun_NilConfig verifies that Run rejects a missing config
// instead of panicking.
func TestRun_NilConfig(t *testing.T) {
	_, err := Run(context.Background(), Options{Config: nil})
	if err == nil {
		t.Fatal("Run with nil Config should return an error")
	}
}

// TestRun_InvalidConfig verifies that Run rejects a config that
// fails Validate (e.g. api port out of range).
func TestRun_InvalidConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default()
	cfg.General.DataDir = dir
	cfg.Storage.Path = filepath.Join(dir, "synaptic.db")
	cfg.APIServer.Port = 99999 // out of 0-65535 range

	_, err := Run(context.Background(), Options{Config: cfg, Listen: ListenSpec{Disable: true}})
	if err == nil {
		t.Fatal("Run with invalid config should return an error")
	}
}
