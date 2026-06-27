package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestScheduler_RotatePrunesByAge verifies O3: Rotate prunes backups
// older than RetentionDays even when they're within KeepN. With
// KeepN=10 (no count prune of 4 backups) and RetentionDays=5, the
// 10-day and 20-day backups are pruned by age; the 0-day and 2-day
// remain. Previously RetentionDays was a shipped config knob read nowhere.
func TestScheduler_RotatePrunesByAge(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	ages := []time.Duration{0, 2 * 24 * time.Hour, 10 * 24 * time.Hour, 20 * 24 * time.Hour}
	for _, age := range ages {
		name := "condura-backup-" + now.Add(-age).Format("2006-01-02T15-04-05Z") + ".zip"
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		mtime := now.Add(-age)
		if err := os.Chtimes(path, mtime, mtime); err != nil {
			t.Fatal(err)
		}
	}
	s := NewScheduler(SchedulerConfig{KeepN: 10, RetentionDays: 5, BackupDir: dir, Now: func() time.Time { return now }}, nil, nil)
	if err := s.Rotate(); err != nil {
		t.Fatal(err)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("RetentionDays=5 should leave 2 (0d,2d), pruned 10d+20d; got %d", len(files))
	}
}

// TestScheduler_RotateCountOnlyWhenRetentionZero verifies that
// RetentionDays=0 means "forever" (no age-prune); only KeepN applies.
func TestScheduler_RotateCountOnlyWhenRetentionZero(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	// 3 recent backups (all within 1 minute, so age-prune would keep
	// them all); KeepN=1 should leave only the newest.
	for i := 0; i < 3; i++ {
		name := "condura-backup-" + now.Add(-time.Duration(i)*time.Minute).Format("2006-01-02T15-04-05Z") + ".zip"
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		mtime := now.Add(-time.Duration(i) * time.Minute)
		if err := os.Chtimes(path, mtime, mtime); err != nil {
			t.Fatal(err)
		}
	}
	s := NewScheduler(SchedulerConfig{KeepN: 1, RetentionDays: 0, BackupDir: dir, Now: func() time.Time { return now }}, nil, nil)
	if err := s.Rotate(); err != nil {
		t.Fatal(err)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("KeepN=1, RetentionDays=0 should leave 1 (newest); got %d", len(files))
	}
}
