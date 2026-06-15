package backup

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupDataDir creates a temp data dir populated with a minimal
// "live" state: a main DB, a memory DB, a skills DB, secrets.json,
// and a config.yaml. Returns the data dir and the master key
// (32 bytes, base64).
//
// All artifacts live INSIDE the data dir. Previously skills.db
// was placed in the sibling dir, but the daemon (subsystems.go
// buildPhase12) and the production daemon (verified via curl
// against a real synapticd) put it in the data dir. Tests must
// match production.
func setupDataDir(t *testing.T) (dataDir, configPath string, masterKey []byte) {
	t.Helper()
	dataDir = t.TempDir()

	// main DB
	mustWrite(t, filepath.Join(dataDir, "synaptic.db"), []byte("MAIN-DB-CONTENT"))
	// memory DB
	mustWrite(t, filepath.Join(dataDir, "memory.db"), []byte("MEMORY-DB-CONTENT"))
	// skills DB lives in the data dir (matches subsystems.go
	// buildPhase12 and verified by curl smoke test).
	mustWrite(t, filepath.Join(dataDir, "skills.db"), []byte("SKILLS-DB-CONTENT"))
	mustWrite(t, filepath.Join(dataDir, "skills.db-wal"), []byte("SKILLS-WAL-CONTENT"))
	mustWrite(t, filepath.Join(dataDir, "skills.db-shm"), []byte("SKILLS-SHM-CONTENT"))
	// secrets.json
	mustWrite(t, filepath.Join(dataDir, "secrets.json"), []byte(`{"master_key":"k6Qm1xJ4pYqZ8cV2nB3wD5rT7eH9uL0sA1bC2dE3fG4="}`))
	// config.yaml
	configPath = filepath.Join(dataDir, "config.yaml")
	mustWrite(t, configPath, []byte("version: 3\n"))

	// 32-byte master key, arbitrary.
	masterKey = make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i + 1)
	}
	return dataDir, configPath, masterKey
}

func mustWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func newTestManager(t *testing.T) (dataDir, outPath string, bm *Manager, masterKey []byte) {
	t.Helper()
	var cfgPath string
	dataDir, cfgPath, masterKey = setupDataDir(t)
	outPath = filepath.Join(dataDir, "test-backup.zip")
	bmIface, err := New(Options{
		DataDir:       dataDir,
		ConfigPath:    cfgPath,
		MasterKey:     masterKey,
		SchemaVersion: 3,
		Out:           outPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	bm = bmIface
	return
}

// ---- Key derivation ----

func TestDeriveKey_StableForSameMaster(t *testing.T) {
	mk := make([]byte, 32)
	for i := range mk {
		mk[i] = byte(i + 1)
	}
	k1, err := DeriveKey(mk)
	if err != nil {
		t.Fatal(err)
	}
	k2, err := DeriveKey(mk)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(k1, k2) {
		t.Error("DeriveKey not stable")
	}
	if len(k1) != 32 {
		t.Errorf("derived key length = %d, want 32", len(k1))
	}
}

func TestDeriveKey_RejectsBadSize(t *testing.T) {
	if _, err := DeriveKey([]byte("short")); err == nil {
		t.Error("expected error for short key")
	}
}

func TestDeriveKeyBase64_RoundTrip(t *testing.T) {
	mk := make([]byte, 32)
	for i := range mk {
		mk[i] = byte(i + 1)
	}
	s, err := DeriveKeyBase64(mk)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 44 { // 32 bytes base64 = 44 chars
		t.Errorf("base64 length = %d, want 44", len(s))
	}
}

// ---- Create + ListFiles + LoadManifest ----

func TestCreate_ArchiveContainsExpectedFiles(t *testing.T) {
	dataDir, outPath, bm, _ := newTestManager(t)
	if _, err := bm.Create(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("archive not created: %v", err)
	}
	names, err := ListBackupFiles(outPath)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"config.yaml",
		"manifest.json",
		"memory.db",
		"secrets.json",
		"skills.db",
		"skills.db-shm",
		"skills.db-wal",
		"synaptic.db",
	}
	if len(names) != len(want) {
		t.Errorf("file count = %d, want %d (got %v)", len(names), len(want), names)
	}
	for i, w := range want {
		if names[i] != w {
			t.Errorf("file[%d] = %q, want %q", i, names[i], w)
		}
	}
	_ = dataDir
}

func TestCreate_ManifestSchemaAndFingerprint(t *testing.T) {
	_, outPath, bm, _ := newTestManager(t)
	if _, err := bm.Create(context.Background()); err != nil {
		t.Fatal(err)
	}
	m, err := LoadManifest(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if m.Version != ManifestVersion {
		t.Errorf("version = %d, want %d", m.Version, ManifestVersion)
	}
	if m.SchemaVersion != 3 {
		t.Errorf("schema_version = %d, want 3", m.SchemaVersion)
	}
	if m.KeyFingerprint == "" {
		t.Error("key_fingerprint empty")
	}
}

func TestCreate_ManifestChecksumsMatch(t *testing.T) {
	_, outPath, bm, _ := newTestManager(t)
	if _, err := bm.Create(context.Background()); err != nil {
		t.Fatal(err)
	}
	m, err := LoadManifest(outPath)
	if err != nil {
		t.Fatal(err)
	}
	// Each file in the manifest should have a non-empty SHA-256.
	for _, f := range m.Files {
		if f.SHA256 == "" {
			t.Errorf("file %s has no sha256", f.Path)
		}
		if !f.Encrypted {
			t.Errorf("file %s marked as unencrypted", f.Path)
		}
	}
}

// ---- Restore ----

func TestRestore_RoundTripPreservesContents(t *testing.T) {
	// 1. Back up the original.
	dataDir, outPath, bm, mk := newTestManager(t)
	if _, err := bm.Create(context.Background()); err != nil {
		t.Fatal(err)
	}
	// 2. Restore from the backup into a sibling dir (same
	// filesystem, so the atomic swap works).
	restoreDir := filepath.Join(filepath.Dir(dataDir), "restore-target")
	if err := os.MkdirAll(restoreDir, 0o700); err != nil {
		t.Fatal(err)
	}
	// Put a throwaway file in restoreDir so we can prove the
	// restore replaced it.
	mustWrite(t, filepath.Join(restoreDir, "synaptic.db"), []byte("THROWAWAY"))

	if err := Restore(context.Background(), RestoreOptions{
		ArchivePath:          outPath,
		DataDir:              restoreDir,
		MasterKey:            mk,
		CurrentSchemaVersion: 3,
		PreRestoreBackupPath: "",
	}); err != nil {
		t.Fatal(err)
	}

	// 3. Verify the originals are back.
	got, err := os.ReadFile(filepath.Join(restoreDir, "synaptic.db"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "MAIN-DB-CONTENT" {
		t.Errorf("synaptic.db = %q, want %q", got, "MAIN-DB-CONTENT")
	}
	got, _ = os.ReadFile(filepath.Join(restoreDir, "memory.db"))
	if string(got) != "MEMORY-DB-CONTENT" {
		t.Errorf("memory.db = %q, want %q", got, "MEMORY-DB-CONTENT")
	}
	// Skills DB lives in the data dir, with WAL/SHM sidecars
	// next to it. (Round-trip test exercises the same layout
	// the production daemon uses.)
	got, _ = os.ReadFile(filepath.Join(restoreDir, "skills.db"))
	if string(got) != "SKILLS-DB-CONTENT" {
		t.Errorf("skills.db = %q, want %q", got, "SKILLS-DB-CONTENT")
	}
	got, _ = os.ReadFile(filepath.Join(restoreDir, "skills.db-wal"))
	if string(got) != "SKILLS-WAL-CONTENT" {
		t.Errorf("skills.db-wal = %q, want %q", got, "SKILLS-WAL-CONTENT")
	}
	got, _ = os.ReadFile(filepath.Join(restoreDir, "skills.db-shm"))
	if string(got) != "SKILLS-SHM-CONTENT" {
		t.Errorf("skills.db-shm = %q, want %q", got, "SKILLS-SHM-CONTENT")
	}
	// Ensure WAL/SHM did NOT end up at the parent of the data
	// dir (the old broken assumption).
	parent := filepath.Dir(restoreDir)
	if _, err := os.Stat(filepath.Join(parent, "skills.db")); err == nil {
		t.Errorf("skills.db leaked into parent dir")
	}
	if _, err := os.Stat(filepath.Join(parent, "skills.db-wal")); err == nil {
		t.Errorf("skills.db-wal leaked into parent dir")
	}
	if _, err := os.Stat(filepath.Join(parent, "skills.db-shm")); err == nil {
		t.Errorf("skills.db-shm leaked into parent dir")
	}
	got, _ = os.ReadFile(filepath.Join(restoreDir, "secrets.json"))
	if !stringContains(got, "master_key") {
		t.Errorf("secrets.json = %q, missing master_key", got)
	}
}

// (copyDataDir was removed: the round-trip test uses an in-place
// restore-target sibling because the atomic swap cannot reach
// across filesystems. Re-add if a future test needs it.)

func stringContains(haystack []byte, needle string) bool {
	return len(haystack) > 0 && (stringContainsImpl(haystack, needle) || len(haystack) >= len(needle) && string(haystack) == needle)
}

func stringContainsImpl(haystack []byte, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if string(haystack[i:i+len(needle)]) == needle {
			return true
		}
	}
	return false
}

func TestRestore_RejectsNewerSchema(t *testing.T) {
	// Create a backup at schema v4 (higher than v3).
	dataDir, cfgPath, mk := setupDataDir(t)
	outPath := filepath.Join(dataDir, "test-backup.zip")
	bmIface, err := New(Options{
		DataDir:       dataDir,
		ConfigPath:    cfgPath,
		MasterKey:     mk,
		SchemaVersion: 4,
		Out:           outPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bmIface.Create(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Try to restore with CurrentSchemaVersion=3 (as if an older
	// binary were running). The restore should be rejected because
	// the backup's schema (v4) is newer than the binary's (v3).
	restoreDir := t.TempDir()
	err = Restore(context.Background(), RestoreOptions{
		ArchivePath:          outPath,
		DataDir:              restoreDir,
		MasterKey:            mk,
		CurrentSchemaVersion: 3,
	})
	if err == nil {
		t.Fatal("expected ErrSchemaIncompatible for v4 backup → v3 binary")
	}
	if !errors.Is(err, ErrSchemaIncompatible) {
		t.Fatalf("expected ErrSchemaIncompatible, got: %v", err)
	}
}

func TestRestore_RejectsUnknownKey(t *testing.T) {
	_, outPath, _, mk := newTestManager(t)
	// Create a backup with the right key.
	bmIface, _ := New(Options{
		DataDir: filepath.Dir(outPath), MasterKey: mk, SchemaVersion: 3, Out: outPath,
	})
	if _, err := bmIface.Create(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Try to restore with a wrong key.
	wrongKey := make([]byte, 32)
	for i := range wrongKey {
		wrongKey[i] = byte(i)
	}
	restoreDir := t.TempDir()
	err := Restore(context.Background(), RestoreOptions{
		ArchivePath:          outPath,
		DataDir:              restoreDir,
		MasterKey:            wrongKey,
		CurrentSchemaVersion: 3,
	})
	if err == nil {
		t.Fatal("expected ErrUnknownKey or decrypt failure")
	}
}

// TestRestore_RejectsOversizedEntry proves that a malicious manifest
// claiming a huge file size does not cause an unbounded allocation.
func TestRestore_RejectsOversizedEntry(t *testing.T) {
	_, outPath, bm, mk := newTestManager(t)
	if _, err := bm.Create(context.Background()); err != nil {
		t.Fatal(err)
	}

	// Mutate the manifest to claim synaptic.db is 2 GiB.
	m, err := LoadManifest(outPath)
	if err != nil {
		t.Fatal(err)
	}
	for i := range m.Files {
		if m.Files[i].Path == "synaptic.db" {
			m.Files[i].Size = 2 << 30
			break
		}
	}

	// Rewrite the archive with the mutated manifest. We can do this
	// by reading the original zip, replacing manifest.json, and writing
	// a new zip.
	zr, err := zip.OpenReader(outPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zr.Close() }()

	newOut := filepath.Join(t.TempDir(), "oversized.zip")
	f, err := os.Create(newOut)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	// Write mutated manifest.
	if err := writeManifestJSON(zw, *m); err != nil {
		t.Fatal(err)
	}
	// Copy original entries unchanged.
	for _, e := range zr.File {
		if e.Name == "manifest.json" {
			continue
		}
		rc, err := e.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(rc)
		if err != nil {
			t.Fatal(err)
		}
		_ = rc.Close()
		w, err := zw.CreateHeader(&e.FileHeader)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	restoreDir := t.TempDir()
	if err := os.MkdirAll(restoreDir, 0o700); err != nil {
		t.Fatal(err)
	}
	err = Restore(context.Background(), RestoreOptions{
		ArchivePath:          newOut,
		DataDir:              restoreDir,
		MasterKey:            mk,
		CurrentSchemaVersion: 3,
	})
	if err == nil {
		t.Fatal("expected error for oversized manifest entry")
	}
}

func TestCreate_ArchiveEncryptedOnDisk(t *testing.T) {
	// The bytes on disk must NOT contain the plaintext
	// "MAIN-DB-CONTENT" verbatim (the entries are encrypted).
	_, outPath, bm, _ := newTestManager(t)
	if _, err := bm.Create(context.Background()); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	plaintext := []byte("MAIN-DB-CONTENT")
	if containsBytes(raw, plaintext) {
		t.Fatal("plaintext found in archive — encryption broken")
	}
	// Also: the manifest is intentionally unencrypted, so the
	// schema version number should be findable. (Sanity check.)
	schemaBytes := []byte(`"schema_version": 3`)
	if !containsBytes(raw, schemaBytes) {
		t.Error("schema version not in archive — manifest missing")
	}
}

func containsBytes(haystack, needle []byte) bool {
	if len(needle) == 0 {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// ---- Scheduler ----

func TestScheduler_RotateKeepsN(t *testing.T) {
	dir := t.TempDir()
	// Create 5 fake backup files with different mtimes.
	for i := 0; i < 5; i++ {
		name := filepath.Join(dir, "synaptic-backup-2026-06-1"+string(rune('0'+i))+"T00-00-00Z.zip")
		if err := os.WriteFile(name, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		// Force a distinct mtime.
		pastTime := time.Now().Add(-time.Duration(i) * time.Hour)
		if err := os.Chtimes(name, pastTime, pastTime); err != nil {
			t.Fatal(err)
		}
	}
	s := NewScheduler(SchedulerConfig{
		Interval:  time.Hour,
		KeepN:     2,
		BackupDir: dir,
		Now:       time.Now,
	}, nil, nil)
	if err := s.Rotate(); err != nil {
		t.Fatal(err)
	}
	files, _ := os.ReadDir(dir)
	backups := 0
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".zip" {
			backups++
		}
	}
	if backups != 2 {
		t.Errorf("after rotate, kept %d backups, want 2", backups)
	}
}
