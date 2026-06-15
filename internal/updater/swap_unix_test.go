//go:build !windows

package updater

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSwapExecutable_Unix(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "synapticd")
	staged := filepath.Join(dir, "staged")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staged, []byte("new"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := swapExecutable(staged, target); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Fatalf("got %q", data)
	}
}
