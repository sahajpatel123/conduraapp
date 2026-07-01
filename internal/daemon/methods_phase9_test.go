package daemon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Audit 2026-07-01 (P1-4): policy.yaml is the file that decides
// what the agent is allowed to do. If any local user can rewrite
// it, the safety layer is bypassed by the local attack model. The
// hard requirement: mode <= 0600. These tests pin the behavior.

func TestReadPolicyFile_RejectsWorldReadable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	if err := os.WriteFile(path, []byte("version: \"1\"\nrules: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := readPolicyFile(path)
	if err == nil {
		t.Fatal("readPolicyFile accepted 0644; expected mode-rejection error")
	}
	if !strings.Contains(err.Error(), "too permissive") {
		t.Errorf("error %q must explain why the file was rejected", err)
	}
	if !strings.Contains(err.Error(), "0600") {
		t.Errorf("error %q must point the operator at the 0600 target", err)
	}
}

func TestReadPolicyFile_RejectsGroupWritable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	if err := os.WriteFile(path, []byte("version: \"1\"\nrules: []\n"), 0o660); err != nil {
		t.Fatal(err)
	}
	_, err := readPolicyFile(path)
	if err == nil {
		t.Fatal("readPolicyFile accepted 0660; expected mode-rejection error")
	}
}

func TestReadPolicyFile_AcceptsStrict0600(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	want := []byte("version: \"1\"\nrules: []\n")
	if err := os.WriteFile(path, want, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := readPolicyFile(path)
	if err != nil {
		t.Fatalf("readPolicyFile(0600): %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("readPolicyFile content = %q; want %q", got, want)
	}
}

func TestReadPolicyFile_AcceptsReadOnly0400(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	want := []byte("version: \"1\"\nrules: []\n")
	if err := os.WriteFile(path, want, 0o400); err != nil {
		t.Fatal(err)
	}
	got, err := readPolicyFile(path)
	if err != nil {
		t.Fatalf("readPolicyFile(0400): %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("readPolicyFile content = %q; want %q", got, want)
	}
}

func TestReadPolicyFile_MissingFileReturnsNil(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist.yaml")
	got, err := readPolicyFile(path)
	if err != nil {
		t.Errorf("missing file must NOT error (fall back to defaults), got: %v", err)
	}
	if got != nil {
		t.Errorf("missing file must return nil bytes, got %d bytes", len(got))
	}
}
