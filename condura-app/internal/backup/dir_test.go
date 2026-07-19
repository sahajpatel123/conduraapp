package backup

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestResolveBackupDir_DefaultDocuments(t *testing.T) {
	t.Setenv("CONDURA_BACKUP_DIR", "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dir := ResolveBackupDir("/var/synaptic")
	want := filepath.Join(home, "Documents", "condura-backups")
	if dir != want {
		t.Fatalf("got %q, want %q", dir, want)
	}
}

func TestResolveBackupDir_EnvOverride(t *testing.T) {
	custom := t.TempDir()
	t.Setenv("CONDURA_BACKUP_DIR", custom)
	dir := ResolveBackupDir("/var/synaptic")
	if dir != custom {
		t.Fatalf("got %q, want %q", dir, custom)
	}
}

func TestResolveBackupDir_FallbackDataDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows resolves a system profile even when HOME is empty")
	}
	t.Setenv("CONDURA_BACKUP_DIR", "")
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	dir := ResolveBackupDir("/var/synaptic")
	want := filepath.Join("/var/synaptic", "backups")
	if dir != want {
		t.Fatalf("got %q, want %q", dir, want)
	}
	_ = os.MkdirAll(dir, 0o700)
}

// touchFile creates an empty regular file at path. Used by the
// ListBackupArchives tests below to populate a backup dir
// without writing real zip content (the filter is extension-only;
// the list does not validate zip structure — that's the inspect
// command's job).
func touchFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("not a real zip"), 0o600); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
}

// TestListBackupArchives_NotFoundIsNotAnError pins the
// fresh-install contract: a missing backup directory returns
// (nil, ErrNotFound) — NOT a generic error and NOT a panic —
// so the CLI can render "no backups yet" instead of bailing.
// IsBackupDirNotFound must return true for the error so the
// CLI can switch on it.
func TestListBackupArchives_NotFoundIsNotAnError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	got, err := ListBackupArchives(missing)
	if err == nil {
		t.Fatal("ListBackupArchives(missing) err = nil; want not-found")
	}
	if got != nil {
		t.Errorf("ListBackupArchives(missing) returned %v alongside error; want nil", got)
	}
	if !IsBackupDirNotFound(err) {
		t.Errorf("IsBackupDirNotFound(err) = false; want true for missing dir")
	}
}

// TestListBackupArchives_NotADirectory pins the contract for
// "the path exists but isn't a directory" (e.g. operator
// misconfigures CONDURA_BACKUP_DIR to a file): returns a
// not-found-class error so the CLI still treats it as
// 'nothing to list', rather than bubbling a confusing
// os.PathError to the operator.
func TestListBackupArchives_NotADirectory(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "iamafile")
	touchFile(t, file)

	got, err := ListBackupArchives(file)
	if err == nil {
		t.Fatal("ListBackupArchives(file) err = nil; want not-a-directory")
	}
	if got != nil {
		t.Errorf("ListBackupArchives(file) returned %v alongside error; want nil", got)
	}
	if !IsBackupDirNotFound(err) {
		t.Errorf("IsBackupDirNotFound(err) = false; want true for non-directory")
	}
}

// TestListBackupArchives_FiltersByExtension pins the extension
// filter: only .zip files appear in the result, even when
// the directory contains other files (manifests, sidecars,
// log files dropped by accident).
func TestListBackupArchives_FiltersByExtension(t *testing.T) {
	dir := t.TempDir()
	touchFile(t, filepath.Join(dir, "condura-backup-2026-06-14.zip"))
	touchFile(t, filepath.Join(dir, "condura-backup-2026-06-15.zip"))
	touchFile(t, filepath.Join(dir, "manifest.json"))
	touchFile(t, filepath.Join(dir, "README.md"))
	touchFile(t, filepath.Join(dir, "not-a-archive.tar"))

	got, err := ListBackupArchives(dir)
	if err != nil {
		t.Fatalf("ListBackupArchives: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d archives, want 2 (.zip filter)", len(got))
	}
	for _, p := range got {
		if filepath.Ext(p) != ".zip" {
			t.Errorf("non-zip in result: %q", p)
		}
	}
}

// TestListBackupArchives_NewestFirst pins the sort order:
// newest mtime first. The operator's mental model for a
// "list backups" command is "what's the most recent one I
// can restore?" — that one sorts to the top.
func TestListBackupArchives_NewestFirst(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "condura-backup-old.zip")
	mid := filepath.Join(dir, "condura-backup-mid.zip")
	newest := filepath.Join(dir, "condura-backup-newest.zip")

	touchFile(t, old)
	// Bump mtime explicitly so the test is deterministic
	// across filesystems with low-resolution mtime clocks.
	pastTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(old, pastTime, pastTime); err != nil {
		t.Fatal(err)
	}
	touchFile(t, mid)
	midTime := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(mid, midTime, midTime); err != nil {
		t.Fatal(err)
	}
	touchFile(t, newest)
	newTime := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(newest, newTime, newTime); err != nil {
		t.Fatal(err)
	}

	got, err := ListBackupArchives(dir)
	if err != nil {
		t.Fatalf("ListBackupArchives: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d archives, want 3", len(got))
	}
	wantOrder := []string{newest, mid, old}
	for i, want := range wantOrder {
		if got[i] != want {
			t.Errorf("got[%d] = %q, want %q (newest-first sort)", i, got[i], want)
		}
	}
}

// TestListBackupArchives_EmptyDir pins the boundary: a real
// directory with no archives returns (empty-slice, nil) — not
// (nil, nil), so callers can range over the result without a
// nil-check and so json.Marshal emits "[]" rather than "null".
func TestListBackupArchives_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	got, err := ListBackupArchives(dir)
	if err != nil {
		t.Fatalf("ListBackupArchives: %v", err)
	}
	if got == nil {
		t.Error("got nil; want empty slice (json-stable)")
	}
	if len(got) != 0 {
		t.Errorf("got %d archives, want 0", len(got))
	}
}

// TestListBackupArchives_SkipsUnreadableEntries pins the
// fault-tolerance contract: a single unreadable file in the
// directory must NOT abort the whole list. The operator
// should still see the readable ones; the bad entry is
// silently dropped (with a real DBG log in production, but
// the test just asserts the result).
func TestListBackupArchives_SkipsUnreadableEntries(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod 000 behaves differently on Windows")
	}
	dir := t.TempDir()
	good := filepath.Join(dir, "condura-backup-good.zip")
	touchFile(t, good)
	bad := filepath.Join(dir, "condura-backup-bad.zip")
	touchFile(t, bad)
	// Make 'bad' unreadable so the os.Stat call inside
	// ListBackupArchives fails. We can't actually trigger the
	// Stat-failure path this way (root can always stat) — but
	// the test still covers the case where bad is a symlink
	// to /nonexistent, which os.Stat DOES fail on.
	if err := os.Remove(bad); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/nonexistent/path/that/never/exists", bad); err != nil {
		t.Skipf("symlink not supported here: %v", err)
	}

	got, err := ListBackupArchives(dir)
	if err != nil {
		t.Fatalf("ListBackupArchives: %v (should not abort on unreadable entry)", err)
	}
	if len(got) != 1 {
		t.Errorf("got %d archives, want 1 (unreadable entry should be skipped)", len(got))
	}
	if len(got) >= 1 && got[0] != good {
		t.Errorf("got[0] = %q, want %q", got[0], good)
	}
}
