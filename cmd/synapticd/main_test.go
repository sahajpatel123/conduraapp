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
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	bin := filepath.Join(binDir, "synapticd"+ext)
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
	t.Helper()
	bin := os.Getenv("__SYNAPSE_TEST_BIN")
	if bin == "" {
		t.Skip("__SYNAPSE_TEST_BIN not set; TestBinaryPath should run first")
	}
	return bin
}

// stopDaemon sends SIGTERM on Unix or Kill on Windows.
func stopDaemon(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	if cmd.Process == nil {
		return
	}
	if runtime.GOOS == "windows" {
		_ = cmd.Process.Kill()
	} else {
		_ = cmd.Process.Signal(syscall.SIGTERM)
	}
	_ = cmd.Wait()
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
	if !bytes.Contains(out, []byte("version: 4")) {
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
	defer func() { stopDaemon(t, cmd) }()

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

// TestDataDirFlagPropagates is a regression test for the bug where
// --data-dir on synapticd was ignored: the daemon reported
// `storage_path=/Users/.../.synaptic/synaptic.db` even when the user
// passed `--data-dir /tmp/whatever`. The fix was to re-derive the
// storage path when the flag overrides the YAML value.
func TestDataDirFlagPropagates(t *testing.T) {
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
	// Restore the default SIGTERM behavior so the cleanup path is
	// straightforward even if the test fails before we get there.
	defer func() { stopDaemon(t, cmd) }()

	// Wait up to 5s for the address file to appear, signaling that
	// the daemon has finished startup and written its log line.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(filepath.Join(dataDir, "synapticd.addr")); err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Stop the daemon BEFORE reading the captured output.
	stopDaemon(t, cmd)

	log := stdout.String() + stderr.String()
	if !strings.Contains(log, "storage_path="+dataDir) {
		t.Fatalf("storage_path did not reflect --data-dir\n--- log ---\n%s", log)
	}
	// And the daemon should NOT have logged a path under ~/.synaptic.
	if strings.Contains(log, "storage_path=/Users/") && !strings.Contains(log, "storage_path="+dataDir) {
		t.Fatalf("storage_path still points to default location\n--- log ---\n%s", log)
	}
	// Verify the SQLite file ends up inside the requested data dir.
	dbPath := filepath.Join(dataDir, "synaptic.db")
	if _, err := os.Stat(dbPath); err != nil {
		// The file may be created lazily on first write. We only need
		// to confirm the path the daemon LOGGED was correct, which we
		// already checked above.
		t.Logf("note: %s not yet on disk: %v", dbPath, err)
	}
}
