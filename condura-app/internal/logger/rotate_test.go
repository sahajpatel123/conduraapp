package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRotatingWriter_RotatesOnSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "condura.log")

	w, err := newRotatingWriter(rotatingConfig{
		Filename:   path,
		MaxSize:    64, // tiny so we can force rotation quickly
		MaxBackups: 3,
		MaxAgeDays: 30,
	})
	if err != nil {
		t.Fatalf("newRotatingWriter: %v", err)
	}
	defer func() { _ = w.Close() }()

	// Write enough to exceed MaxSize multiple times.
	payload := strings.Repeat("x", 40)
	for i := 0; i < 6; i++ {
		if _, err := w.Write([]byte(payload + "\n")); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
	}

	// Active file must exist.
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("active log missing: %v", err)
	}
	// At least one rotated sibling should exist.
	backups, err := listBackups(path)
	if err != nil {
		t.Fatalf("listBackups: %v", err)
	}
	if len(backups) == 0 {
		t.Fatal("expected at least one rotated backup, got none")
	}
	// Cap: never more than MaxBackups.
	if len(backups) > 3 {
		t.Fatalf("kept %d backups, want ≤ 3", len(backups))
	}
}

func TestRotatingWriter_MaxBackupsPrunes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "daemon.log")

	w, err := newRotatingWriter(rotatingConfig{
		Filename:   path,
		MaxSize:    32,
		MaxBackups: 2,
		MaxAgeDays: -1, // disable age prune for this test
	})
	if err != nil {
		t.Fatalf("newRotatingWriter: %v", err)
	}
	defer func() { _ = w.Close() }()

	payload := strings.Repeat("y", 30) + "\n"
	for i := 0; i < 20; i++ {
		if _, err := w.Write([]byte(payload)); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
	}

	backups, err := listBackups(path)
	if err != nil {
		t.Fatalf("listBackups: %v", err)
	}
	if len(backups) > 2 {
		t.Fatalf("kept %d backups, want ≤ 2", len(backups))
	}
}

func TestRotatingWriter_MaxAgePrunes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "old.log")

	// Seed an "old" rotated file directly.
	oldPath := path + ".1"
	if err := os.WriteFile(oldPath, []byte("stale\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Backdate mtime to 40 days ago.
	oldTime := time.Now().AddDate(0, 0, -40)
	if err := os.Chtimes(oldPath, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	w, err := newRotatingWriter(rotatingConfig{
		Filename:   path,
		MaxSize:    32,
		MaxBackups: 5,
		MaxAgeDays: 7,
	})
	if err != nil {
		t.Fatalf("newRotatingWriter: %v", err)
	}
	defer func() { _ = w.Close() }()

	// Force a rotation so pruneByAge runs.
	payload := strings.Repeat("z", 40) + "\n"
	if _, err := w.Write([]byte(payload)); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(payload)); err != nil {
		t.Fatal(err)
	}

	// The 40-day-old file (whatever index it ended up at after shift)
	// must be gone — any remaining backups must be younger than MaxAgeDays.
	backups, err := listBackups(path)
	if err != nil {
		t.Fatal(err)
	}
	cutoff := time.Now().AddDate(0, 0, -7)
	for _, b := range backups {
		info, err := os.Stat(b.path)
		if err != nil {
			t.Fatal(err)
		}
		if info.ModTime().Before(cutoff) {
			t.Errorf("stale backup still present: %s mtime=%v", b.path, info.ModTime())
		}
	}
}

func TestRotatingWriter_OpenExistingAppends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "append.log")
	if err := os.WriteFile(path, []byte("seed\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	w, err := newRotatingWriter(rotatingConfig{
		Filename:   path,
		MaxSize:    1024 * 1024,
		MaxBackups: 3,
		MaxAgeDays: 30,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("more\n")); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()

	data, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, "seed") || !strings.Contains(got, "more") {
		t.Fatalf("expected append, got %q", got)
	}
}

func TestListBackups_IgnoresNonNumeric(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.log")
	_ = os.WriteFile(path, []byte("a"), 0o600)
	_ = os.WriteFile(path+".1", []byte("b"), 0o600)
	_ = os.WriteFile(path+".2", []byte("c"), 0o600)
	_ = os.WriteFile(path+".bak", []byte("d"), 0o600) // ignored
	_ = os.WriteFile(path+".sha256", []byte("e"), 0o600)

	backups, err := listBackups(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 2 {
		t.Fatalf("got %d backups, want 2", len(backups))
	}
	if backups[0].index != 1 || backups[1].index != 2 {
		t.Fatalf("unexpected order: %+v", backups)
	}
}

func TestParsePositiveInt(t *testing.T) {
	cases := []struct {
		in   string
		want int
		ok   bool
	}{
		{"1", 1, true},
		{"12", 12, true},
		{"0", 0, false},
		{"", 0, false},
		{"1a", 0, false},
		{"-3", 0, false},
		{"01", 1, true},
	}
	for _, tc := range cases {
		got, ok := parsePositiveInt(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Errorf("parsePositiveInt(%q) = (%d,%v), want (%d,%v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestNew_WithFileRotation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "from-new.log")
	lg := New(Config{
		Level:      LevelInfo,
		Format:     FormatJSON,
		File:       path,
		MaxSizeMB:  1,
		MaxBackups: 2,
		MaxAgeDays: 7,
		Redact:     boolPtr(true),
	})
	// Close the rotating file before TempDir cleanup — Windows holds
	// exclusive locks on open log handles and fails RemoveAll otherwise.
	t.Cleanup(func() { _ = CloseFileSink(lg) })
	lg.Info("hello", "k", "v")
	// File must have been created.
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("log file not created: %v", err)
	}
}
