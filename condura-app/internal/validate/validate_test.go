package validate

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// TestRun_NonexistentDataDirFails pins the "data dir must exist"
// contract: passing a path that doesn't exist returns StatusFail
// for data_dir (everything else is status_skip or fail depending
// on optional/required). The operator must create the data dir
// before validate is useful.
func TestRun_NonexistentDataDirFails(t *testing.T) {
	t.Setenv("CONDURA_BACKUP_DIR", "")
	missing := filepath.Join(t.TempDir(), "no-such-data-dir")

	r := Run(context.Background(), missing)

	if r.DataDir != missing {
		t.Errorf("DataDir = %q, want %q", r.DataDir, missing)
	}
	// data_dir check on missing dir is StatusFail (NOT Skip —
	// the data dir MUST exist).
	if r.Checks[0].Status != StatusFail {
		t.Errorf("data_dir Status = %q, want fail (data dir missing)", r.Checks[0].Status)
	}
	if r.Checks[0].Detail == "" {
		t.Error("data_dir Detail empty; want stat-error message")
	}
	// main_db on missing data dir → fail (the file can't exist
	// if the dir doesn't).
	if r.Checks[1].Status != StatusFail {
		t.Errorf("main_db Status = %q, want fail", r.Checks[1].Status)
	}
}

// TestRun_EmptyFreshInstallDirReturnsExpectedShape pins the
// "data dir exists but is otherwise empty" contract — what an
// operator sees right after `condurad --init` runs:
//
//	data_dir  ok
//	main_db   fail (no DB file yet — daemon will create on first start)
//	memory_db skip (fresh install)
//	skills_db skip (fresh install)
//	config    skip (defaults apply)
//	lock      skip (daemon not running)
//	backups   skip (no backup dir)
//
// JSON shape is stable (7 checks, in deterministic order).
func TestRun_EmptyFreshInstallDirReturnsExpectedShape(t *testing.T) {
	dir := t.TempDir() // exists, but empty
	t.Setenv("CONDURA_BACKUP_DIR", filepath.Join(dir, "no-backups"))

	r := Run(context.Background(), dir)

	if r.DataDir != dir {
		t.Errorf("DataDir = %q, want %q", r.DataDir, dir)
	}
	if r.Time == "" {
		t.Error("Time is empty; want RFC3339")
	}

	if len(r.Checks) != 7 {
		t.Errorf("got %d checks, want 7", len(r.Checks))
	}
	wantNames := []string{"data_dir", "main_db", "memory_db", "skills_db", "config", "lock", "backups"}
	for i, want := range wantNames {
		if i < len(r.Checks) && r.Checks[i].Name != want {
			t.Errorf("Checks[%d].Name = %q, want %q", i, r.Checks[i].Name, want)
		}
	}
	// Fresh-install expectations:
	if r.Checks[0].Status != StatusOK {
		t.Errorf("data_dir Status = %q, want ok (dir exists)", r.Checks[0].Status)
	}
	if r.Checks[1].Status != StatusFail {
		t.Errorf("main_db Status = %q, want fail (no DB file)", r.Checks[1].Status)
	}
	if r.Checks[2].Status != StatusSkip {
		t.Errorf("memory_db Status = %q, want skip", r.Checks[2].Status)
	}
	if r.Checks[3].Status != StatusSkip {
		t.Errorf("skills_db Status = %q, want skip", r.Checks[3].Status)
	}
	if r.Checks[4].Status != StatusSkip {
		t.Errorf("config Status = %q, want skip", r.Checks[4].Status)
	}
	if r.Checks[5].Status != StatusSkip {
		t.Errorf("lock Status = %q, want skip", r.Checks[5].Status)
	}
	if r.Checks[6].Status != StatusSkip {
		t.Errorf("backups Status = %q, want skip", r.Checks[6].Status)
	}
	// Summary: 1 ok, 0 warn, 1 fail, 5 skip.
	if r.Summary.OK != 1 {
		t.Errorf("Summary.OK = %d, want 1 (data_dir)", r.Summary.OK)
	}
	if r.Summary.Fail != 1 {
		t.Errorf("Summary.Fail = %d, want 1 (main_db)", r.Summary.Fail)
	}
	if r.Summary.Skip != 5 {
		t.Errorf("Summary.Skip = %d, want 5", r.Summary.Skip)
	}
}

// TestRun_WithConfigAndDBsHappyPath pins the all-green contract:
// a fresh data dir with a valid config + a valid (empty) main DB
// produces mostly OK results.
func TestRun_WithConfigAndDBsHappyPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CONDURA_BACKUP_DIR", filepath.Join(dir, "no-backups"))

	// Minimal valid config.
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"),
		[]byte("api_server:\n  host: 127.0.0.1\n  port: 7666\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Valid empty SQLite DB.
	if err := createValidSQLite(t, filepath.Join(dir, "condura.db")); err != nil {
		t.Fatal(err)
	}

	r := Run(context.Background(), dir)

	// data_dir OK.
	if r.Checks[0].Status != StatusOK {
		t.Errorf("data_dir = %q, want ok", r.Checks[0].Status)
	}
	// main_db OK.
	if r.Checks[1].Status != StatusOK {
		t.Errorf("main_db = %q, want ok (valid SQLite)", r.Checks[1].Status)
	}
	// config OK.
	if r.Checks[4].Status != StatusOK {
		t.Errorf("config = %q, want ok (valid YAML)", r.Checks[4].Status)
	}
}

// TestRun_BrokenConfigFails pins the YAML-parse contract: a
// config file that doesn't parse as YAML is StatusFail (the
// daemon won't start with a broken config).
func TestRun_BrokenConfigFails(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CONDURA_BACKUP_DIR", filepath.Join(dir, "no-backups"))

	if err := os.WriteFile(filepath.Join(dir, "config.yaml"),
		[]byte("api_server:\n  host: [unclosed bracket\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	r := Run(context.Background(), dir)

	// config check must be StatusFail with parse-error detail.
	var configCheck *Check
	for i := range r.Checks {
		if r.Checks[i].Name == "config" {
			configCheck = &r.Checks[i]
			break
		}
	}
	if configCheck == nil {
		t.Fatal("config check not found")
	}
	if configCheck.Status != StatusFail {
		t.Errorf("config Status = %q, want fail (broken YAML)", configCheck.Status)
	}
	if configCheck.Detail == "" {
		t.Error("config Detail empty; want parse-error message")
	}
}

// TestRun_EmptyDataDirArg pins the empty-input contract:
// passing "" as data dir uses the homedir.Dir default. The
// resulting data dir in the Report should NOT be empty.
func TestRun_EmptyDataDirArg(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // isolated HOME
	t.Setenv("CONDURA_BACKUP_DIR", "")

	r := Run(context.Background(), "")

	if r.DataDir == "" {
		t.Error("DataDir is empty; want the homedir.Dir default")
	}
}

// TestCheckSQLiteFile_OptionalDBSkip pins the optional-DB
// contract: missing memory.db or skills.db is StatusSkip (fresh
// install), NOT StatusFail.
func TestCheckSQLiteFile_OptionalDBSkip(t *testing.T) {
	c := checkSQLiteFile(context.Background(), filepath.Join(t.TempDir(), "does-not-exist.db"), "memory_db", true)
	if c.Status != StatusSkip {
		t.Errorf("optional missing DB Status = %q, want skip", c.Status)
	}
}

// TestCheckSQLiteFile_RequiredDBFail pins the required-DB
// contract: missing main_db is StatusFail (the daemon can't
// start without it).
func TestCheckSQLiteFile_RequiredDBFail(t *testing.T) {
	c := checkSQLiteFile(context.Background(), filepath.Join(t.TempDir(), "does-not-exist.db"), "main_db", false)
	if c.Status != StatusFail {
		t.Errorf("required missing DB Status = %q, want fail", c.Status)
	}
}

// TestCheckSQLiteFile_CorruptDBFails pins the corruption contract:
// a file that exists but isn't valid SQLite is StatusFail (not
// StatusSkip — the file IS present, it's just broken).
func TestCheckSQLiteFile_CorruptDBFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "corrupt.db")
	if err := os.WriteFile(path, []byte("not a sqlite file"), 0o600); err != nil {
		t.Fatal(err)
	}
	c := checkSQLiteFile(context.Background(), path, "main_db", false)
	if c.Status != StatusFail {
		t.Errorf("corrupt DB Status = %q, want fail", c.Status)
	}
}

// createValidSQLite creates a minimal valid SQLite file using
// the same driver the production code uses. Used by the
// happy-path test.
func createValidSQLite(t *testing.T, path string) error {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec("CREATE TABLE _validate_probe (id INTEGER)"); err != nil {
		return err
	}
	return nil
}
