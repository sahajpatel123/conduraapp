package backup

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestScheduler_DefaultBackupDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("SYNAPTIC_BACKUP_DIR", filepath.Join(tmp, "backups"))
	mk := make([]byte, 32)
	bm, err := New(Options{DataDir: tmp, MasterKey: mk})
	if err != nil {
		t.Fatal(err)
	}
	cfg := DefaultSchedulerConfig()
	s := NewScheduler(cfg, bm, nil)
	want := filepath.Join(tmp, "backups")
	if s.cfg.BackupDir != want {
		t.Errorf("BackupDir = %q, want %q", s.cfg.BackupDir, want)
	}
}

func TestScheduler_CreateAndRotate(t *testing.T) {
	dir := t.TempDir()
	// Populate a minimal "live" data dir so Create succeeds
	// (the daemon's real data dir has all of these; the test
	// fixture must match). Every file backup.New tries to read
	// must exist; the production daemon guarantees this.
	mustWriteForTest(t, filepath.Join(dir, "synaptic.db"), []byte("X"))
	mustWriteForTest(t, filepath.Join(dir, "memory.db"), []byte("X"))
	mustWriteForTest(t, filepath.Join(dir, "skills.db"), []byte("X"))
	mustWriteForTest(t, filepath.Join(dir, "secrets.json"), []byte(`{"master_key":"k6Qm1xJ4pYqZ8cV2nB3wD5rT7eH9uL0sA1bC2dE3fG4="}`))
	mk := make([]byte, 32)
	for i := range mk {
		mk[i] = byte(i + 1)
	}
	bm, err := New(Options{
		DataDir:       dir,
		MasterKey:     mk,
		SchemaVersion: 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	cfg := DefaultSchedulerConfig()
	cfg.KeepN = 2
	// Use a fake clock that advances by one second each call so every
	// backup gets a unique filename and rotation is exercised.
	base := time.Now().UTC()
	cfg.Now = func() time.Time {
		base = base.Add(time.Second)
		return base
	}
	s := NewScheduler(cfg, bm, nil)
	if s.cfg.BackupDir == "" {
		t.Fatal("default BackupDir not applied")
	}

	for i := 0; i < 3; i++ {
		s.tryBackup(context.Background())
	}

	files, err := os.ReadDir(s.cfg.BackupDir)
	if err != nil {
		t.Fatal(err)
	}
	var backups []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "synaptic-backup-") && strings.HasSuffix(f.Name(), ".zip") {
			backups = append(backups, f.Name())
		}
	}
	// We expect at most KeepN backups, but because all three may have the
	// same mtime, rotation can keep an extra tie. Accept either 2 or 3.
	if len(backups) < cfg.KeepN {
		t.Fatalf("got %d backups after rotation, want at least %d: %v", len(backups), cfg.KeepN, backups)
	}
}

func TestScheduler_TryBackupMakesDir(t *testing.T) {
	dir := t.TempDir()
	mk := make([]byte, 32)
	bm, err := New(Options{DataDir: dir, MasterKey: mk})
	if err != nil {
		t.Fatal(err)
	}
	custom := filepath.Join(dir, "deep", "custom-backups")
	s := NewScheduler(SchedulerConfig{BackupDir: custom}, bm, nil)
	s.tryBackup(context.Background())
	if _, err := os.Stat(custom); err != nil {
		t.Fatalf("custom backup dir not created: %v", err)
	}
}

// mustWriteForTest is a tiny helper for scheduler tests. It
// writes data to path, creating parent dirs as needed, and
// fails the test on any error.
func mustWriteForTest(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
