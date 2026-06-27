package daemon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestResumeSecret_EnvOverride: env var takes priority and is not
// persisted to disk.
func TestResumeSecret_EnvOverride(t *testing.T) {
	const envVal = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	t.Setenv("CONDURA_RESUME_SECRET", envVal)
	mgr := NewResumeSecretManager(t.TempDir(), "CONDURA_RESUME_SECRET")
	got, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != envVal {
		t.Fatalf("Load = %q, want %q (env value)", got, envVal)
	}
	// Should NOT have written to disk.
	dir := t.TempDir()
	mgr2 := NewResumeSecretManager(dir, "CONDURA_RESUME_SECRET")
	// Force load without env var to ensure no auto-generation happened
	// in the previous test.
	t.Setenv("CONDURA_RESUME_SECRET", "")
	if _, err := mgr2.Load(); err != nil {
		t.Fatalf("Load (no env, fresh dir): %v", err)
	}
	// Confirm a file was written — but its value is auto-generated.
	b, err := os.ReadFile(filepath.Join(dir, "resume.secret"))
	if err != nil {
		t.Fatalf("read resume.secret: %v (env override should not have written)", err)
	}
	if string(b) == envVal {
		t.Fatalf("env value was written to disk; should not be")
	}
}

// TestResumeSecret_AutoGenerate: with no env var and no existing
// file, Load auto-generates a 64-char hex secret and persists it.
func TestResumeSecret_AutoGenerate(t *testing.T) {
	dir := t.TempDir()
	mgr := NewResumeSecretManager(dir, "CONDURA_RESUME_SECRET")
	got, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 64 {
		t.Fatalf("auto-generated secret length = %d, want 64", len(got))
	}
	if err := validateResumeSecret(got); err != nil {
		t.Fatalf("auto-generated secret failed validation: %v", err)
	}
	// Second Load returns the same secret (idempotent + persisted).
	got2, err := mgr.Load()
	if err != nil {
		t.Fatalf("second Load: %v", err)
	}
	if got2 != got {
		t.Fatalf("second Load = %q, want %q (stable across calls)", got2, got)
	}
	// And the file is mode 0600.
	info, err := os.Stat(filepath.Join(dir, "resume.secret"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("resume.secret mode = %v, want 0600", info.Mode().Perm())
	}
}

// TestResumeSecret_ExistingFile: if the file already exists with a
// valid 64-char hex secret, Load returns it.
func TestResumeSecret_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	const existing = "1111111111111111111111111111111111111111111111111111111111111111"
	if err := os.WriteFile(filepath.Join(dir, "resume.secret"), []byte(existing), 0o600); err != nil {
		t.Fatal(err)
	}
	mgr := NewResumeSecretManager(dir, "CONDURA_RESUME_SECRET")
	got, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != existing {
		t.Fatalf("Load = %q, want %q (existing file)", got, existing)
	}
}

// TestResumeSecret_BadExistingFile: a file with a non-hex / wrong-
// length content must error cleanly (not panic, not silently re-gen).
func TestResumeSecret_BadExistingFile(t *testing.T) {
	dir := t.TempDir()
	// Too short + non-hex — both validation rules fail.
	if err := os.WriteFile(filepath.Join(dir, "resume.secret"), []byte("nope"), 0o600); err != nil {
		t.Fatal(err)
	}
	mgr := NewResumeSecretManager(dir, "CONDURA_RESUME_SECRET")
	_, err := mgr.Load()
	if err == nil {
		t.Fatal("Load on bad file must error")
	}
	// Crucially: the bad file must NOT be auto-overwritten by an
	// auto-generated secret on next Load (that would silently lose
	// the user's existing config). Load should keep erroring.
	if _, err := mgr.Load(); err == nil {
		t.Fatal("Load on bad file must keep erroring (no auto-overwrite of bad data)")
	}
}

// TestValidateResumeSecret: length and hex validity checks.
func TestValidateResumeSecret(t *testing.T) {
	const good = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if err := validateResumeSecret(good); err != nil {
		t.Fatalf("validateResumeSecret(good) = %v, want nil", err)
	}
	if err := validateResumeSecret(""); err == nil {
		t.Fatal("validate empty should error")
	}
	if err := validateResumeSecret(strings.Repeat("z", 64)); err == nil {
		t.Fatal("validate non-hex should error")
	}
	if err := validateResumeSecret(strings.Repeat("0", 63)); err == nil {
		t.Fatal("validate short should error")
	}
}
