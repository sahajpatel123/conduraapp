package onboarding

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite" // pure-Go sqlite for tests
)

// TestReadEULAFromPath_HappyPath pins the contract: a real file
// at absPath MUST return an *EULADocument with:
// - Version == CurrentEULAVersion (always — this is the canonical
//   embedded EULA, not a user-stored one)
// - Text == the file contents verbatim
// - UpdatedAt extracted from the embedded "# Updated: YYYY-MM-DD"
//   header (if present)
func TestReadEULAFromPath_HappyPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "eula.txt")
	body := "Condura Terms\n**Last updated:** 2026-07-19\n\nDo not be evil.\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := readEULAFromPath(path)
	if err != nil {
		t.Fatalf("readEULAFromPath: %v", err)
	}
	if got.Version != CurrentEULAVersion {
		t.Errorf("Version = %q, want %q", got.Version, CurrentEULAVersion)
	}
	if got.Text != body {
		t.Errorf("Text = %q, want %q (verbatim file contents)", got.Text, body)
	}
	if got.UpdatedAt == "" {
		t.Errorf("UpdatedAt is empty; extractUpdatedAt should have parsed the '# Updated:' header")
	}
}

// TestReadEULAFromPath_MissingFileReturnsWrappedError pins the
// failure-mode contract: when absPath doesn't exist,
// readEULAFromPath MUST return an error wrapped with the "read
// EULA" prefix. Without the prefix, callers can't distinguish
// "EULA file missing" from other errors (e.g., permission
// denied, corrupted read).
func TestReadEULAFromPath_MissingFileReturnsWrappedError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "no-such-eula.txt")

	_, err := readEULAFromPath(path)
	if err == nil {
		t.Fatal("readEULAFromPath on missing file = nil; want error")
	}
	if !strings.Contains(err.Error(), "read EULA") {
		t.Errorf("error %q should mention 'read EULA' for diagnostic clarity", err.Error())
	}
}

// TestReadEULAFromPath_PermissionDeniedReturnsWrappedError pins
// the permission-denied edge case: a path the process can't
// read MUST return an error wrapped with the "read EULA" prefix.
// (Skipped on platforms where chmod 000 still allows root to read
// — e.g., root in CI Linux. We use 0o000 here, which denies
// read for non-root users; on macOS dev where the test user is
// not root, this works.)
func TestReadEULAFromPath_PermissionDeniedReturnsWrappedError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root; permission test is ineffective")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "locked-eula.txt")
	if err := os.WriteFile(path, []byte("body"), 0o000); err != nil {
		t.Fatal(err)
	}
	// Restore permissions on cleanup so t.TempDir can clean up.
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	_, err := readEULAFromPath(path)
	if err == nil {
		t.Fatal("readEULAFromPath on locked file = nil; want error")
	}
	if !strings.Contains(err.Error(), "read EULA") {
		t.Errorf("error %q should mention 'read EULA' for diagnostic clarity", err.Error())
	}
}

// TestNewStateMachine_NilDBReturnsError pins the input-validation
// guard: NewStateMachine MUST reject nil DB with a clear error
// (not panic, not return a zero-value StateMachine). A regression
// that returned &StateMachine{} with nil DB would crash later
// at first State() call with a less obvious error.
func TestNewStateMachine_NilDBReturnsError(t *testing.T) {
	sm, err := NewStateMachine(nil)
	if err == nil {
		t.Fatal("NewStateMachine(nil) = nil error; want error")
	}
	if sm != nil {
		t.Errorf("NewStateMachine(nil) returned non-nil StateMachine %p; want nil", sm)
	}
	if !strings.Contains(err.Error(), "db is required") {
		t.Errorf("error %q should mention 'db is required'", err.Error())
	}
}

// TestNewStateMachine_MigratesOnConstruction pins the
// migration-on-construct contract: NewStateMachine MUST run the
// schema migration BEFORE returning. After construction, the
// onboarding_state table exists and contains the default row
// (id=1, state_json='{}'). We verify by querying the table
// directly.
func TestNewStateMachine_MigratesOnConstruction(t *testing.T) {
	dir := t.TempDir()
	db := openOnboardingTestDB(t, dir)
	defer func() { _ = db.Close() }()

	sm, err := NewStateMachine(db)
	if err != nil {
		t.Fatalf("NewStateMachine: %v", err)
	}
	if sm == nil {
		t.Fatal("NewStateMachine returned nil StateMachine")
	}

	// Verify the migration created the table with the default row.
	var stateJSON string
	if err := db.QueryRow(`SELECT state_json FROM onboarding_state WHERE id = 1`).Scan(&stateJSON); err != nil {
		t.Fatalf("migration didn't create onboarding_state row: %v", err)
	}
	if stateJSON != "{}" {
		t.Errorf("default state_json = %q, want %q", stateJSON, "{}")
	}
}

// TestNewStateMachine_MigrationFailurePropagatesError pins the
// error-propagation contract: if the migration query fails (e.g.,
// due to a closed DB), NewStateMachine MUST return an error
// wrapped with "onboarding: migrate" so log readers can
// diagnose.
func TestNewStateMachine_MigrationFailurePropagatesError(t *testing.T) {
	dir := t.TempDir()
	db := openOnboardingTestDB(t, dir)

	// Close the DB BEFORE construction so the migration query
	// fails inside NewStateMachine.
	_ = db.Close()

	_, err := NewStateMachine(db)
	if err == nil {
		t.Fatal("NewStateMachine on closed DB returned nil error; want error")
	}
	if !strings.Contains(err.Error(), "migrate") {
		t.Errorf("error %q should mention 'migrate' for diagnostic clarity", err.Error())
	}
}

// openOnboardingTestDB opens a fresh sqlite database for tests.
func openOnboardingTestDB(t *testing.T, dir string) *sql.DB {
	t.Helper()
	path := filepath.Join(dir, "onboarding.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	return db
}