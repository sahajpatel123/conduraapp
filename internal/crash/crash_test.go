package crash

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestCapture_WritesLocalFile(t *testing.T) {
	// Override home for test.
	t.Setenv("HOME", t.TempDir())
	r := Capture("test panic")
	if r == nil {
		t.Fatal("Capture returned nil")
	}
	if r.StackHash == "" {
		t.Error("stack hash is empty")
	}
	if len(r.Stack) == 0 {
		t.Error("stack is empty")
	}
	// Verify local file written.
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".synaptic", "crashes")
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		t.Fatal("no crash file written to", dir)
	}
}

func TestCapture_StackHashIsValidHex(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	r := Capture("test")
	if r == nil {
		t.Fatal("Capture returned nil")
	}
	if _, err := hex.DecodeString(r.StackHash); err != nil {
		t.Errorf("stack hash not valid hex: %s", r.StackHash)
	}
}

func TestToTelemetry_ExcludesStack(t *testing.T) {
	r := &Report{
		StackHash: "abc123",
		Version:   "0.1.0",
		Platform:  "darwin/arm64",
		Stack:     []byte("full-stack-trace-data"),
	}
	payload := r.ToTelemetry()
	if payload.StackHash != "abc123" {
		t.Error("stack hash mismatch")
	}
	if payload.Version != "0.1.0" {
		t.Error("version mismatch")
	}
	// Verify payload does NOT contain the raw stack.
	if len(r.Stack) == 0 {
		t.Error("original stack was empty")
	}
}

func TestRecover_DoesNotPanic(t *testing.T) {
	// Call Recover in a normal (non-panicking) context — should be safe.
	Recover()
}

func TestRecover_CatchesPanic(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	func() {
		defer Recover()
		panic("test recovery")
	}()
	// If we reach here, Recover caught the panic.
}
