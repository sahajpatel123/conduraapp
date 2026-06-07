// Smoke tests for the synapticd binary. These spawn the actual binary
// as a subprocess so we exercise the full startup path, but they keep
// the scope small: version flag, default-config flag, and a brief
// run that exits cleanly on SIGTERM.

package main_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestBinaryPath(t *testing.T) {
	// Build the binary into a temp dir for the suite.
	t.Helper()
	binDir := t.TempDir()
	bin := filepath.Join(binDir, "synapticd")
	// Repo root is two levels up from this test file (cmd/synapticd).
	_, thisFile, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/synapticd")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build synapticd: %v\n%s", err, out)
	}
	// Use a non-SYNAPTIC_ env var name so it doesn't get parsed as a
	// config override by the daemon's env-loader.
	t.Setenv("__SYNAPSE_TEST_BIN", bin)
}

func synapticd(t *testing.T) string {
	bin := os.Getenv("__SYNAPSE_TEST_BIN")
	if bin == "" {
		t.Skip("__SYNAPSE_TEST_BIN not set; TestBinaryPath should run first")
	}
	return bin
}

func TestVersionFlag(t *testing.T) {
	TestBinaryPath(t)
	bin := synapticd(t)
	out, err := exec.Command(bin, "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("--version: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Synaptic") {
		t.Fatalf("unexpected --version output: %q", out)
	}
}

func TestPrintDefaultConfig(t *testing.T) {
	TestBinaryPath(t)
	bin := synapticd(t)
	out, err := exec.Command(bin, "--print-default-config").CombinedOutput()
	if err != nil {
		t.Fatalf("--print-default-config: %v\n%s", err, out)
	}
	if !bytes.Contains(out, []byte("version: 1")) {
		t.Fatalf("missing top-level version key in:\n%s", out)
	}
	if !bytes.Contains(out, []byte("synapticd:")) && !bytes.Contains(out, []byte("daemon:")) {
		t.Fatalf("missing daemon section in:\n%s", out)
	}
}

func TestStartsAndStopsCleanly(t *testing.T) {
	TestBinaryPath(t)
	bin := synapticd(t)
	dataDir := t.TempDir()

	cmd := exec.Command(bin,
		"--data-dir", dataDir,
		"--listen", "tcp://127.0.0.1:0",
		"--log-level", "info",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer func() {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		_ = cmd.Wait()
	}()

	// Wait up to 5s for the address file to appear.
	deadline := time.Now().Add(5 * time.Second)
	var addr string
	for time.Now().Before(deadline) {
		b, err := os.ReadFile(filepath.Join(dataDir, "synapticd.addr"))
		if err == nil && len(b) > 0 {
			addr = strings.TrimSpace(string(b))
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if addr == "" {
		t.Fatalf("synapticd.addr never appeared\n--- stdout ---\n%s\n--- stderr ---\n%s", stdout.String(), stderr.String())
	}
	if !strings.HasPrefix(addr, "127.0.0.1:") {
		t.Fatalf("unexpected addr: %q", addr)
	}
}
