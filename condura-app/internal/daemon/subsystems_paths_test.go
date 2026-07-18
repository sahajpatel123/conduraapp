package daemon

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/storage"
)

// TestGeneralDataDir_NilGuards pins the nil-safety contract
// for GeneralDataDir: it MUST return "" when called on a nil
// receiver OR when Storage is nil. Callers in the backup
// and uninstall subsystems gate on this empty-string return
// to decide whether subsystems are wired up; a panic would
// crash them mid-restore.
func TestGeneralDataDir_NilGuards(t *testing.T) {
	// Nil receiver.
	var s *Subsystems
	if got := s.GeneralDataDir(); got != "" {
		t.Errorf("nil receiver: GeneralDataDir() = %q, want empty string", got)
	}

	// Non-nil receiver but Storage is nil.
	s = &Subsystems{}
	if got := s.GeneralDataDir(); got != "" {
		t.Errorf("Storage nil: GeneralDataDir() = %q, want empty string", got)
	}
}

// TestGeneralDataDir_ReturnsParentDir pins the contract:
// GeneralDataDir returns the DIRECTORY containing the main
// database file (i.e. filepath.Dir(Storage.Path())). The data
// dir is where skills.db, memory.db, secrets.json, and other
// subsystem state live — not where condura.db itself lives.
//
// A regression that returned Storage.Path() directly (the
// .db file) would silently break every downstream caller
// that joins the result with 'skills.db' / 'memory.db'.
func TestGeneralDataDir_ReturnsParentDir(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "condura.db")
	db, err := storage.Open(context.Background(), storage.Config{Path: dbPath})
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	s := &Subsystems{Storage: db}
	want := filepath.Dir(dbPath)
	if got := s.GeneralDataDir(); got != want {
		t.Errorf("GeneralDataDir() = %q, want %q (parent of %q)", got, want, dbPath)
	}
}

// TestMemoryDBPath_NilGuards pins the nil-safety contract for
// MemoryDBPath: it MUST return "" when called on a nil receiver
// OR when Storage is nil.
//
// ReloadAuxiliaryDatabases (the backup.restore path) calls
// MemoryDBPath() and gates on empty string to decide whether
// to reopen the memory store. A panic or a non-empty string
// from a nil receiver would either crash the daemon during
// restore or open a memory.db at a wrong location.
func TestMemoryDBPath_NilGuards(t *testing.T) {
	var s *Subsystems
	if got := s.MemoryDBPath(); got != "" {
		t.Errorf("nil receiver: MemoryDBPath() = %q, want empty string", got)
	}
	s = &Subsystems{}
	if got := s.MemoryDBPath(); got != "" {
		t.Errorf("Storage nil: MemoryDBPath() = %q, want empty string", got)
	}
}

// TestMemoryDBPath_LivesAlongsideMainDB pins the contract that
// memory.db lives in the SAME DIRECTORY as the main condura.db,
// NOT in a sibling directory. The comment on SkillDBPath
// explicitly calls this out as a single-source-of-truth rule —
// any caller that hard-codes 'filepath.Dir(dataDir)/memory.db'
// or 'dataDir/memory.db' (without GeneralDataDir) is a bug.
//
// A regression that moved memory.db to a sibling directory
// would break backup restore (the restore path would not
// find the file).
func TestMemoryDBPath_LivesAlongsideMainDB(t *testing.T) {
	dataDir := t.TempDir()
	dbPath := filepath.Join(dataDir, "condura.db")
	db, err := storage.Open(context.Background(), storage.Config{Path: dbPath})
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	s := &Subsystems{Storage: db}
	want := filepath.Join(dataDir, "memory.db")
	got := s.MemoryDBPath()
	if got != want {
		t.Errorf("MemoryDBPath() = %q, want %q (alongside main DB)", got, want)
	}

	// Negative contract: must NOT be in a sibling directory of
	// the data dir. The classic pre-fix bug was
	// filepath.Dir(dataDir)/memory.db — i.e. one level UP from
	// the data dir. Pin that the result starts with dataDir.
	if !filepathHasPrefix(got, dataDir) {
		t.Errorf("MemoryDBPath() = %q escapes dataDir %q (must be INSIDE, not sibling)", got, dataDir)
	}
}

// TestSkillDBPath_NilGuards pins the nil-safety contract for
// SkillDBPath: returns "" on nil receiver OR nil Storage.
func TestSkillDBPath_NilGuards(t *testing.T) {
	var s *Subsystems
	if got := s.SkillDBPath(); got != "" {
		t.Errorf("nil receiver: SkillDBPath() = %q, want empty string", got)
	}
	s = &Subsystems{}
	if got := s.SkillDBPath(); got != "" {
		t.Errorf("Storage nil: SkillDBPath() = %q, want empty string", got)
	}
}

// TestSkillDBPath_LivesAlongsideMainDB pins the same single-
// source-of-truth contract for skills.db. The docstring on
// SkillDBPath explicitly warns that hard-coded
// 'filepath.Dir(dataDir)/skills.db' in any package is a bug.
//
// A regression that moved skills.db to a sibling directory
// would silently break Phase 12 (the skills feature is
// initialized at buildPhase12 which reads from this path).
func TestSkillDBPath_LivesAlongsideMainDB(t *testing.T) {
	dataDir := t.TempDir()
	dbPath := filepath.Join(dataDir, "condura.db")
	db, err := storage.Open(context.Background(), storage.Config{Path: dbPath})
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	s := &Subsystems{Storage: db}
	want := filepath.Join(dataDir, "skills.db")
	got := s.SkillDBPath()
	if got != want {
		t.Errorf("SkillDBPath() = %q, want %q (alongside main DB)", got, want)
	}
	if !filepathHasPrefix(got, dataDir) {
		t.Errorf("SkillDBPath() = %q escapes dataDir %q (must be INSIDE, not sibling)", got, dataDir)
	}
}

// TestPathGetters_MemoryAndSkillAreDistinct pins the
// non-collision contract: MemoryDBPath and SkillDBPath MUST
// return different paths. A regression that aliased one to the
// other (e.g. "they're both in the data dir, who cares?") would
// cause the memory and skills SQLite stores to write to the
// same file and corrupt each other on first open.
func TestPathGetters_MemoryAndSkillAreDistinct(t *testing.T) {
	dataDir := t.TempDir()
	dbPath := filepath.Join(dataDir, "condura.db")
	db, err := storage.Open(context.Background(), storage.Config{Path: dbPath})
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	s := &Subsystems{Storage: db}
	memPath := s.MemoryDBPath()
	skillPath := s.SkillDBPath()

	if memPath == "" || skillPath == "" {
		t.Fatalf("both paths must be non-empty when Storage is set; got mem=%q skill=%q", memPath, skillPath)
	}
	if memPath == skillPath {
		t.Errorf("MemoryDBPath and SkillDBPath collide at %q (would corrupt one store on the other's writes)", memPath)
	}
}

// filepathHasPrefix returns true if p is path-equal to or
// nested under dir. Uses filepath.Rel to normalize before
// the check, so trailing separators and '..' segments don't
// fool the comparison.
func filepathHasPrefix(p, dir string) bool {
	rel, err := filepath.Rel(dir, p)
	if err != nil {
		return false
	}
	// If p is inside dir, rel starts without '..'. If p is
	// outside dir (sibling/parent), rel starts with '..'.
	if rel == "" {
		return true // equal
	}
	if rel[:2] == ".." {
		return false
	}
	return true
}
