package diag

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTake_FreshInstallReturnsValidSnapshot pins the basic
// contract: Take() on a fresh empty data dir returns a Snapshot
// with paths populated, zero-sized file infos (because the files
// don't exist), and no backups. The JSON must marshal without
// error and round-trip cleanly.
//
// Uses a fresh empty backup dir via CONDURA_BACKUP_DIR override
// so the test is independent of whatever archives may exist in
// the developer's real ~/Documents/condura-backups.
func TestTake_FreshInstallReturnsValidSnapshot(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CONDURA_BACKUP_DIR", filepath.Join(dir, "no-such-backups"))
	s := Take(dir)

	// Paths populated from the input dir.
	if s.Paths.DataDir != dir {
		t.Errorf("Paths.DataDir = %q, want %q", s.Paths.DataDir, dir)
	}
	if s.Paths.MainDB != filepath.Join(dir, "condura.db") {
		t.Errorf("Paths.MainDB = %q, want %q", s.Paths.MainDB, filepath.Join(dir, "condura.db"))
	}

	// Files don't exist → zero-size infos (no error).
	if s.MainDB.Size != 0 {
		t.Errorf("MainDB.Size on fresh install = %d, want 0", s.MainDB.Size)
	}
	if s.Config.Size != 0 {
		t.Errorf("Config.Size on fresh install = %d, want 0", s.Config.Size)
	}
	if s.MemoryDB.Size != 0 {
		t.Errorf("MemoryDB.Size on fresh install = %d, want 0", s.MemoryDB.Size)
	}

	// No backups on fresh install.
	if len(s.Backups) != 0 {
		t.Errorf("Backups on fresh install = %d, want 0", len(s.Backups))
	}

	// Version populated (from internal/version).
	if s.Version == "" {
		t.Error("Version is empty; want the build version")
	}

	// Timestamp is RFC3339.
	if s.Timestamp == "" {
		t.Error("Timestamp is empty; want RFC3339")
	}
}

// TestTake_IncludesExistingFiles pins the file-detection
// contract: files that exist are reported with their actual size
// and mtime. Files that don't exist are reported as zero-size
// (so the snapshot is well-formed JSON regardless of install state).
func TestTake_IncludesExistingFiles(t *testing.T) {
	dir := t.TempDir()

	mainDB := filepath.Join(dir, "condura.db")
	content := []byte("main-db-content-1234567890") // 28 bytes
	if err := os.WriteFile(mainDB, content, 0o600); err != nil {
		t.Fatal(err)
	}

	memoryDB := filepath.Join(dir, "memory.db")
	if err := os.WriteFile(memoryDB, []byte("mem"), 0o600); err != nil {
		t.Fatal(err)
	}

	s := Take(dir)

	if s.MainDB.Path != mainDB {
		t.Errorf("MainDB.Path = %q, want %q", s.MainDB.Path, mainDB)
	}
	if s.MainDB.Size != int64(len(content)) {
		t.Errorf("MainDB.Size = %d, want %d", s.MainDB.Size, len(content))
	}
	if s.MainDB.MTime == "" {
		t.Error("MainDB.MTime empty; want RFC3339")
	}

	if s.MemoryDB.Size != 3 {
		t.Errorf("MemoryDB.Size = %d, want 3", s.MemoryDB.Size)
	}

	// Missing config file → zero-size.
	if s.Config.Size != 0 {
		t.Errorf("Config.Size = %d, want 0 (no file)", s.Config.Size)
	}
}

// TestTake_BackupsReflectFilesystem pins the backup-listing
// contract: snapshots include all .zip archives in the backup
// dir, newest-first, with size and mtime. The list uses the
// same path the CLI uses (backup.ListBackupArchives), so the
// snapshot matches what the operator sees.
func TestTake_BackupsReflectFilesystem(t *testing.T) {
	dir := t.TempDir()

	// Set up a fake backup dir with 2 archives.
	backupDir := filepath.Join(dir, "backups")
	if err := os.MkdirAll(backupDir, 0o700); err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(backupDir, "condura-backup-old.zip")
	newest := filepath.Join(backupDir, "condura-backup-newest.zip")
	if err := os.WriteFile(old, []byte("old-content"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newest, []byte("newest-content"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Use a CONDURA_BACKUP_DIR override to make backupDir
	// resolve to our test path (without leaking HOME lookup).
	t.Setenv("CONDURA_BACKUP_DIR", backupDir)

	s := Take(dir)

	if len(s.Backups) != 2 {
		t.Fatalf("got %d backups, want 2", len(s.Backups))
	}
	// Newest-first sort.
	if !strings.Contains(s.Backups[0].Path, "newest") {
		t.Errorf("Backups[0] = %q, want newest-first order", s.Backups[0].Path)
	}
	if !strings.Contains(s.Backups[1].Path, "old") {
		t.Errorf("Backups[1] = %q, want newest-first order", s.Backups[1].Path)
	}
	if s.Backups[0].Size != int64(len("newest-content")) {
		t.Errorf("Backups[0].Size = %d, want %d", s.Backups[0].Size, len("newest-content"))
	}
}

// TestTake_JSONIsStable pins the JSON-shape contract: the
// snapshot must marshal cleanly with no surprise nulls or
// extra fields. Support scrapers depend on this shape; changing
// it is a breaking change for the support pipeline.
//
// CONDURA_BACKUP_DIR is overridden so the test asserts the
// fresh-install "backups":[] shape regardless of the developer's
// local backup state.
func TestTake_JSONIsStable(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CONDURA_BACKUP_DIR", filepath.Join(dir, "no-such-backups"))
	s := Take(dir)

	js, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var back Snapshot
	if err := json.Unmarshal(js, &back); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	// Round-trip preserves all fields.
	if back.Version != s.Version {
		t.Errorf("Version drift: %q vs %q", back.Version, s.Version)
	}
	if back.Paths.DataDir != s.Paths.DataDir {
		t.Errorf("Paths.DataDir drift: %q vs %q", back.Paths.DataDir, s.Paths.DataDir)
	}

	// No "null" for arrays — they should marshal as [].
	if !strings.Contains(string(js), `"backups":[]`) {
		t.Errorf("backups field should marshal as [], got: %s", string(js))
	}
}

// TestTake_NoSecretsInSnapshot pins the privacy contract: the
// snapshot MUST NOT include secrets, tokens, or master key
// material. We can't easily assert "absence of arbitrary
// fields", but we CAN check that the documented fields don't
// accidentally start containing raw bytes or hex strings.
func TestTake_NoSecretsInSnapshot(t *testing.T) {
	// Set up a data dir with a config file that LOOKS like it
	// contains a secret — the snapshot must NOT echo the
	// secret bytes.
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	secret := "supersecretvalue12345"
	if err := os.WriteFile(configPath, []byte("api_key: "+secret), 0o600); err != nil {
		t.Fatal(err)
	}

	s := Take(dir)
	js, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}

	// The snapshot reports the config file's path + size + mtime,
	// NOT its content. The secret string must not appear
	// anywhere in the JSON.
	if strings.Contains(string(js), secret) {
		t.Errorf("snapshot contains secret: %s", string(js))
	}
}
