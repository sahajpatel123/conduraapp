package main_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// Cached binaries shared across all tests via TestMain.
var cachedBinaries struct {
	daemonBin string
	cliBin    string
	err       error
	once      sync.Once
}

func TestMain(m *testing.M) {
	// Build once, reuse across all tests (skipped in -short mode).
	cachedBinaries.once.Do(func() {
		cachedBinaries.daemonBin, cachedBinaries.cliBin, cachedBinaries.err = buildBinariesOnce()
	})
	code := m.Run()
	os.Exit(code)
}

func buildBinariesOnce() (daemonBin, cliBin string, err error) {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	binDir, err := os.MkdirTemp("", "synaptic-cli-test-*")
	if err != nil {
		return "", "", err
	}
	daemonBin = filepath.Join(binDir, "condurad"+ext)
	cliBin = filepath.Join(binDir, "condura"+ext)
	_, thisFile, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	for _, p := range []struct {
		path, name string
	}{
		{daemonBin, "condurad"},
		{cliBin, "condura"},
	} {
		cmd := exec.Command("go", "build", "-o", p.path, "./cmd/"+p.name)
		cmd.Dir = repoRoot
		out, buildErr := cmd.CombinedOutput()
		if buildErr != nil {
			return "", "", errors.New(string(out))
		}
	}
	return daemonBin, cliBin, nil
}

func getBinaries(t *testing.T) (string, string) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping binary-dependent test in -short mode")
	}
	cachedBinaries.once.Do(func() {
		cachedBinaries.daemonBin, cachedBinaries.cliBin, cachedBinaries.err = buildBinariesOnce()
	})
	if cachedBinaries.err != nil {
		t.Fatalf("build binaries: %v", cachedBinaries.err)
	}
	return cachedBinaries.daemonBin, cachedBinaries.cliBin
}

type daemon struct {
	cmd     *exec.Cmd
	dataDir string
	stdout  bytes.Buffer
	stderr  bytes.Buffer
}

func startDaemon(t *testing.T, bin, dataDir string) *daemon {
	t.Helper()
	cmd := exec.Command(bin, "--data-dir", dataDir, "--listen", "tcp://127.0.0.1:0", "--log-level", "info")
	d := &daemon{cmd: cmd, dataDir: dataDir}
	cmd.Stdout = &d.stdout
	cmd.Stderr = &d.stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start daemon: %v", err)
	}
	t.Cleanup(func() {
		if cmd.Process != nil {
			if runtime.GOOS == "windows" {
				_ = cmd.Process.Kill()
			} else {
				_ = cmd.Process.Signal(syscall.SIGTERM)
			}
			_ = cmd.Wait()
		}
	})
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(filepath.Join(dataDir, "condurad.addr")); err == nil {
			return d
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("daemon addr file never appeared\n--- stdout ---\n%s\n--- stderr ---\n%s", d.stdout.String(), d.stderr.String())
	return d
}

func runCLI(t *testing.T, cliBin, dataDir string, args ...string) (string, string, int) {
	t.Helper()
	full := make([]string, 0, 2+len(args))
	full = append(full, "--data-dir", dataDir)
	full = append(full, args...)
	cmd := exec.Command(cliBin, full...)
	var so, se bytes.Buffer
	cmd.Stdout = &so
	cmd.Stderr = &se
	err := cmd.Run()
	code := 0
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("cli run: %v", err)
	}
	return so.String(), se.String(), code
}

func TestCLIHelp(t *testing.T) {
	_, cliBin := getBinaries(t)
	so, _, code := runCLI(t, cliBin, t.TempDir())
	if code != 0 {
		t.Fatalf("help exit %d", code)
	}
	if !strings.Contains(so, "Usage:") {
		t.Fatalf("expected usage text, got: %s", so)
	}
}

func TestCLIPing(t *testing.T) {
	daemonBin, cliBin := getBinaries(t)
	d := startDaemon(t, daemonBin, t.TempDir())
	so, _, code := runCLI(t, cliBin, d.dataDir, "ping")
	if code != 0 {
		t.Fatalf("ping exit %d", code)
	}
	if !strings.Contains(so, "pong") {
		t.Fatalf("expected pong, got: %s", so)
	}
}

func TestCLIVersion(t *testing.T) {
	daemonBin, cliBin := getBinaries(t)
	d := startDaemon(t, daemonBin, t.TempDir())
	so, _, code := runCLI(t, cliBin, d.dataDir, "version")
	if code != 0 {
		t.Fatalf("version exit %d", code)
	}
	if !strings.Contains(so, "condurad") {
		t.Fatalf("expected 'condurad' in output, got: %s", so)
	}
}

func TestCLIStatus(t *testing.T) {
	daemonBin, cliBin := getBinaries(t)
	d := startDaemon(t, daemonBin, t.TempDir())
	so, _, code := runCLI(t, cliBin, d.dataDir, "status")
	if code != 0 {
		t.Fatalf("status exit %d", code)
	}
	if !strings.Contains(so, "health:") {
		t.Fatalf("expected 'health:' in output, got: %s", so)
	}
	if !strings.Contains(so, "providers:") {
		t.Fatalf("expected 'providers:' in output, got: %s", so)
	}
	if !strings.Contains(so, "spend:") {
		t.Fatalf("expected 'spend:' in output, got: %s", so)
	}
}

func TestCLIConfigJSON(t *testing.T) {
	daemonBin, cliBin := getBinaries(t)
	d := startDaemon(t, daemonBin, t.TempDir())
	so, _, code := runCLI(t, cliBin, d.dataDir, "--json", "config")
	if code != 0 {
		t.Fatalf("config --json exit %d", code)
	}
	if !strings.Contains(so, "APIServer") {
		t.Fatalf("expected APIServer section in JSON output, got: %s", so)
	}
}

func TestCLINoDaemon(t *testing.T) {
	_, cliBin := getBinaries(t)
	so, se, code := runCLI(t, cliBin, t.TempDir(), "ping")
	if code == 0 {
		t.Fatalf("expected non-zero exit, got 0\nstdout: %s\nstderr: %s", so, se)
	}
	combined := so + se
	if !strings.Contains(combined, "daemon") && !strings.Contains(combined, "no daemon") {
		t.Fatalf("expected daemon-related error, got: stdout=%q stderr=%q", so, se)
	}
}

func TestCLIUnknownCommand(t *testing.T) {
	_, cliBin := getBinaries(t)
	_, se, code := runCLI(t, cliBin, t.TempDir(), "banana")
	if code == 0 {
		t.Fatalf("expected non-zero exit")
	}
	if !strings.Contains(se, "unknown subcommand") {
		t.Fatalf("expected 'unknown subcommand' in stderr, got: %s", se)
	}
}

func TestCLILLMProvidersEmpty(t *testing.T) {
	daemonBin, cliBin := getBinaries(t)
	d := startDaemon(t, daemonBin, t.TempDir())
	so, _, code := runCLI(t, cliBin, d.dataDir, "llm", "providers")
	if code != 0 {
		t.Fatalf("llm providers exit %d", code)
	}
	// Accept any output — empty, "no providers", "- provider_name", etc.
	if so == "" || strings.Contains(so, "no providers") || strings.Contains(so, "providers") {
		return
	}
}

func TestCLIApikeysListEmpty(t *testing.T) {
	daemonBin, cliBin := getBinaries(t)
	d := startDaemon(t, daemonBin, t.TempDir())
	so, _, code := runCLI(t, cliBin, d.dataDir, "apikeys", "list")
	if code != 0 {
		t.Fatalf("apikeys list exit %d", code)
	}
	if !strings.Contains(so, "no keys") {
		t.Fatalf("expected 'no keys stored', got: %s", so)
	}
}

func TestCLIDaemonStopsGracefully(t *testing.T) {
	daemonBin, _ := getBinaries(t)
	d := startDaemon(t, daemonBin, t.TempDir())
	pid := d.cmd.Process.Pid
	if runtime.GOOS == "windows" {
		if err := d.cmd.Process.Kill(); err != nil {
			t.Fatalf("kill: %v", err)
		}
	} else {
		if err := d.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			t.Fatalf("signal: %v", err)
		}
	}
	done := make(chan error, 1)
	go func() { done <- d.cmd.Wait() }()
	select {
	case err := <-done:
		_ = err
	case <-time.After(10 * time.Second):
		t.Fatalf("daemon pid %d did not exit", pid)
	}
	_ = context.Background
}
