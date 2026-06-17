package cmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(thisFile), "..")
}

// TestBuildAllBinaries verifies the three user-facing commands compile
// on the host platform. Cross-GOOS matrix coverage lives in CI.
func TestBuildAllBinaries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping build-all in -short mode")
	}
	root := repoRoot(t)
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	outDir := t.TempDir()
	pkgs := []struct {
		pkg  string
		name string
	}{
		{"./cmd/synapticd", "condurad"},
		{"./cmd/synaptic", "condura"},
		{"./cmd/condura-tui", "condura-tui"},
	}
	for _, p := range pkgs {
		out := filepath.Join(outDir, p.name+ext)
		cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", out, p.pkg)
		cmd.Dir = root
		if combined, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build %s: %v\n%s", p.pkg, err, combined)
		}
		if _, err := os.Stat(out); err != nil {
			t.Fatalf("binary missing after build %s: %v", p.name, err)
		}
	}
}
