package daemon

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestListBackupArchives_EmptyDirReturnsEmptySlice pins the
// fresh-install contract: an empty backup directory returns
// ([]backupEntry{}, nil), NOT (nil, err). The GUI relies on
// "non-nil empty slice" so it can render "no backups yet"
// without a special-case for nil.
func TestListBackupArchives_EmptyDirReturnsEmptySlice(t *testing.T) {
	got, err := listBackupArchives(t.TempDir())
	if err != nil {
		t.Fatalf("listBackupArchives(empty dir) err = %v, want nil", err)
	}
	if got == nil {
		t.Fatal("got nil; want empty (non-nil) slice")
	}
	if len(got) != 0 {
		t.Errorf("got %d entries, want 0", len(got))
	}
}

// TestListBackupArchives_NewestFirst pins the sort-order contract:
// archives returned newest-first by mtime (delegated to
// backup.ListBackupArchives). The GUI renders a list with the
// most recent backup at the top.
func TestListBackupArchives_NewestFirst(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "condura-backup-old.zip")
	mid := filepath.Join(dir, "condura-backup-mid.zip")
	newest := filepath.Join(dir, "condura-backup-newest.zip")

	for _, p := range []string{old, mid, newest} {
		if err := os.WriteFile(p, []byte("not a real zip"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	// Force distinct mtimes — filesystems with low-resolution
	// mtime clocks (e.g. some macOS configs at 1s granularity)
	// can collapse the timestamps otherwise.
	pastTime := osPastTime(2020, 1, 1)
	if err := os.Chtimes(old, pastTime, pastTime); err != nil {
		t.Fatal(err)
	}
	midTime := osPastTime(2023, 6, 15)
	if err := os.Chtimes(mid, midTime, midTime); err != nil {
		t.Fatal(err)
	}
	newTime := osPastTime(2026, 7, 1)
	if err := os.Chtimes(newest, newTime, newTime); err != nil {
		t.Fatal(err)
	}

	got, err := listBackupArchives(dir)
	if err != nil {
		t.Fatalf("listBackupArchives: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d entries, want 3", len(got))
	}
	wantOrder := []string{newest, mid, old}
	for i, want := range wantOrder {
		if got[i].Path != want {
			t.Errorf("got[%d].Path = %q, want %q (newest-first)", i, got[i].Path, want)
		}
	}
}

// TestListBackupArchives_NameAndSize pins the per-entry shape:
// each backupEntry has Name (basename), Path (full path), and
// Size (bytes). A regression that swapped Name and Path would
// silently break the GUI's "open in Finder" handler.
func TestListBackupArchives_NameAndSize(t *testing.T) {
	dir := t.TempDir()
	arc := filepath.Join(dir, "test-archive.zip")
	content := []byte("12345678") // 8 bytes
	if err := os.WriteFile(arc, content, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := listBackupArchives(dir)
	if err != nil {
		t.Fatalf("listBackupArchives: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d entries, want 1", len(got))
	}
	if got[0].Name != "test-archive.zip" {
		t.Errorf("Name = %q, want test-archive.zip (basename only)", got[0].Name)
	}
	if got[0].Path != arc {
		t.Errorf("Path = %q, want %q (full path)", got[0].Path, arc)
	}
	if got[0].Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d (file byte length)", got[0].Size, len(content))
	}
}

// TestListBackupArchives_FiltersNonZip pins the .zip filter:
// sidecar files (manifest.json, README.md, lock) must NOT appear
// in the result, even when they live in the backup dir.
func TestListBackupArchives_FiltersNonZip(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{
		"condura-backup.zip",
		"manifest.json",
		"README.md",
		"condura.db-wal", // would-be sidecar if backup dir collides with data dir
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got, err := listBackupArchives(dir)
	if err != nil {
		t.Fatalf("listBackupArchives: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d entries, want 1 (.zip filter)", len(got))
	}
	if got[0].Name != "condura-backup.zip" {
		t.Errorf("Name = %q, want condura-backup.zip", got[0].Name)
	}
}

// TestListBackupArchives_EmptyDirArg pins the nil-input contract:
// an empty dir argument returns ([]backupEntry{}, nil) — NOT
// ([]backupEntry{}, error). The function is called with the
// result of subs.GeneralDataDir() which can be "" if subsystems
// are not fully wired; the GUI must not see an error toast for
// "data dir not yet configured".
func TestListBackupArchives_EmptyDirArg(t *testing.T) {
	got, err := listBackupArchives("")
	if err != nil {
		t.Errorf("listBackupArchives(\"\") err = %v, want nil (fresh-install contract)", err)
	}
	if got == nil {
		t.Error("got nil; want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Errorf("got %d entries, want 0", len(got))
	}
}

// osPastTime is a tiny helper that returns a far-past timestamp
// at the start of the given UTC day. Used by tests that need
// deterministic mtimes (filesystem mtime resolution varies —
// some macOS configs round to 1s).
func osPastTime(year int, month time.Month, day int) (t time.Time) {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
