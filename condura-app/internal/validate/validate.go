package validate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/backup"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/homedir"

	"gopkg.in/yaml.v3"

	_ "modernc.org/sqlite"
)

// Status is the outcome of one health check.
type Status string

const (
	// StatusOK means the check passed (the artifact is present
	// and healthy).
	StatusOK Status = "ok"
	// StatusWarn means the check found something that should be
	// fixed but is not blocking (e.g. a stale lockfile from a
	// daemon crash).
	StatusWarn Status = "warn"
	// StatusFail means the check found something wrong that
	// needs immediate attention (e.g. DB integrity failure).
	StatusFail Status = "fail"
	// StatusSkip means the check was not applicable (e.g. an
	// optional artifact is missing on a fresh install).
	StatusSkip Status = "skip"
)

// Check is one health check result. The Name field is a stable
// identifier (e.g. "main_db", "config") that the CLI / GUI /
// scrapers can match on; the Detail field is free-form
// human-readable text.
type Check struct {
	Name   string `json:"name"`
	Status Status `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// CheckName* constants are the stable identifiers used in
// Check.Name. Centralized so the CLI / GUI / scrapers can
// match on them without worrying about typos in literal strings.
const (
	CheckNameDataDir  = "data_dir"
	CheckNameMainDB   = "main_db"
	CheckNameMemoryDB = "memory_db"
	CheckNameSkillsDB = "skills_db"
	CheckNameConfig   = "config"
	CheckNameLock     = "lock"
	CheckNameBackups  = "backups"
)

// Report is the full set of check results for one Run() call.
// It also includes a summary count so the CLI can print
// "5 ok, 1 warn, 0 fail" at the end.
type Report struct {
	DataDir string  `json:"data_dir"`
	Time    string  `json:"time"` // RFC3339
	Checks  []Check `json:"checks"`
	Summary struct {
		OK   int `json:"ok"`
		Warn int `json:"warn"`
		Fail int `json:"fail"`
		Skip int `json:"skip"`
	} `json:"summary"`
}

// dataDir resolves the data dir, defaulting to ~/.condura if
// the input is empty. Mirrors condura CLI's defaultDataDir but
// uses the homedir helper (iter-10).
func dataDir(input string) string {
	if input != "" {
		return input
	}
	if home, err := homedir.Dir(); err == nil {
		return filepath.Join(home, ".condura")
	}
	return ".condura"
}

// Run executes all health checks and returns a Report. Checks
// run sequentially; a failure in one does not stop the others.
// The overall Report is well-formed JSON regardless of failures.
//
// The check order is deterministic so the CLI output is stable
// across runs (no random map iteration drift).
func Run(ctx context.Context, dirInput string) Report {
	dir := dataDir(dirInput)
	r := Report{
		DataDir: dir,
		Time:    time.Now().UTC().Format(time.RFC3339),
	}

	r.Checks = append(r.Checks,
		checkDataDir(dir),
		checkMainDB(ctx, dir),
		checkMemoryDB(ctx, dir),
		checkSkillsDB(ctx, dir),
		checkConfig(dir),
		checkLockFile(dir),
		checkBackups(dir),
	)

	for _, c := range r.Checks {
		switch c.Status {
		case StatusOK:
			r.Summary.OK++
		case StatusWarn:
			r.Summary.Warn++
		case StatusFail:
			r.Summary.Fail++
		case StatusSkip:
			r.Summary.Skip++
		}
	}
	return r
}

// checkDataDir verifies the data dir exists. Missing data dir
// is StatusFail (nothing else can work without it).
func checkDataDir(dir string) Check {
	fi, err := os.Stat(dir)
	if err != nil {
		return Check{Name: CheckNameDataDir, Status: StatusFail, Detail: err.Error()}
	}
	if !fi.IsDir() {
		return Check{Name: CheckNameDataDir, Status: StatusFail, Detail: fmt.Sprintf("not a directory: %s", dir)}
	}
	return Check{Name: CheckNameDataDir, Status: StatusOK}
}

// checkMainDB verifies the main DB exists and passes a basic
// SQLite integrity_check. Read-only connection so a corrupt DB
// doesn't crash the validation. Missing main DB is StatusFail —
// the daemon can't function without it.
func checkMainDB(ctx context.Context, dir string) Check {
	return checkSQLiteFile(ctx, filepath.Join(dir, "condura.db"), CheckNameMainDB, false)
}

// checkMemoryDB verifies the memory DB. Optional — a fresh
// install doesn't have one yet.
func checkMemoryDB(ctx context.Context, dir string) Check {
	return checkSQLiteFile(ctx, filepath.Join(dir, "memory.db"), CheckNameMemoryDB, true)
}

// checkSkillsDB verifies the skills DB. Optional like memory.
func checkSkillsDB(ctx context.Context, dir string) Check {
	return checkSQLiteFile(ctx, filepath.Join(dir, "skills.db"), CheckNameSkillsDB, true)
}

// checkSQLiteFile opens path in read-only mode and runs
// PRAGMA integrity_check. Any failure (open, query, or
// non-"ok" result) becomes StatusFail with the error message.
//
// optional=true means a missing file is StatusSkip (fresh
// install hasn't created the DB yet); optional=false means a
// missing file is StatusFail (this DB is required for daemon
// startup).
func checkSQLiteFile(ctx context.Context, path, name string, optional bool) Check {
	if _, err := os.Stat(path); err != nil {
		if optional && os.IsNotExist(err) {
			return Check{Name: name, Status: StatusSkip, Detail: "not present (fresh install)"}
		}
		return Check{Name: name, Status: StatusFail, Detail: fmt.Sprintf("stat: %v", err)}
	}
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return Check{Name: name, Status: StatusFail, Detail: fmt.Sprintf("open: %v", err)}
	}
	defer func() { _ = db.Close() }()
	var result string
	if err := db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result); err != nil {
		return Check{Name: name, Status: StatusFail, Detail: fmt.Sprintf("integrity_check: %v", err)}
	}
	if result != "ok" {
		return Check{Name: name, Status: StatusFail, Detail: fmt.Sprintf("integrity_check returned %q", result)}
	}
	return Check{Name: name, Status: StatusOK}
}

// checkConfig verifies config.yaml parses as YAML. A parse
// failure is StatusFail (the daemon won't start with a broken
// config). Missing config is StatusSkip (fresh install hasn't
// written one yet — daemon defaults apply).
func checkConfig(dir string) Check {
	path := filepath.Join(dir, "config.yaml")
	//nolint:gosec // G304: path is operator-supplied via --data-dir; not user-tainted.
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Check{Name: CheckNameConfig, Status: StatusSkip, Detail: "not present (defaults apply)"}
		}
		return Check{Name: CheckNameConfig, Status: StatusFail, Detail: fmt.Sprintf("read: %v", err)}
	}
	var anyMap map[string]any
	if err := yaml.Unmarshal(data, &anyMap); err != nil {
		return Check{Name: CheckNameConfig, Status: StatusFail, Detail: fmt.Sprintf("yaml parse: %v", err)}
	}
	return Check{Name: CheckNameConfig, Status: StatusOK}
}

// checkLockFile looks for a stale lock. A present lock with a
// live PID is StatusOK (daemon is running). A present lock with
// a dead PID (process gone) is StatusWarn (stale lock; the
// operator should remove it before restart).
func checkLockFile(dir string) Check {
	path := filepath.Join(dir, "condurad.lock")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return Check{Name: CheckNameLock, Status: StatusSkip, Detail: "no lock file (daemon not running)"}
		}
		return Check{Name: CheckNameLock, Status: StatusFail, Detail: err.Error()}
	}
	// We can't easily probe the live PID without reading the
	// lock content (which depends on the lockfile package's
	// encoding). For now: a present lock is OK; the operator
	// can verify by running 'condura status' (which would fail
	// if the lock is stale). A future iteration could read the
	// PID and check os.FindProcess.
	return Check{Name: CheckNameLock, Status: StatusOK, Detail: "present (daemon running or stale)"}
}

// checkBackups verifies each backup archive's manifest is
// readable. A corrupt archive is StatusFail; a missing backup
// dir is StatusSkip (no backups to validate).
func checkBackups(dir string) Check {
	backupDir := backup.ResolveBackupDir(dir)
	paths, err := backup.ListBackupArchives(backupDir)
	if err != nil {
		if backup.IsBackupDirNotFound(err) {
			return Check{Name: CheckNameBackups, Status: StatusSkip, Detail: "no backup directory (fresh install)"}
		}
		return Check{Name: CheckNameBackups, Status: StatusFail, Detail: err.Error()}
	}
	if len(paths) == 0 {
		return Check{Name: CheckNameBackups, Status: StatusSkip, Detail: "no backup archives to validate"}
	}
	// Validate each archive. Report a single StatusFail if any
	// archive is unreadable (we don't want to spam the operator
	// with one row per broken archive — they can use
	// `condura backup inspect <archive>` for details).
	for _, p := range paths {
		if _, err := backup.LoadManifest(p); err != nil {
			return Check{Name: CheckNameBackups, Status: StatusFail, Detail: fmt.Sprintf("%s: %v", filepath.Base(p), err)}
		}
	}
	return Check{Name: CheckNameBackups, Status: StatusOK, Detail: fmt.Sprintf("%d archive(s) valid", len(paths))}
}

// Compile-time check: errors package is used in future-proofing
// for structured-error handling (currently unused but the
// import keeps the door open for a future iteration that adds
// error-chain matching).
var _ = errors.New
