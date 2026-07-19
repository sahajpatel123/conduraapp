package lockfile

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestTryAcquire_Fresh(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "condurad.lock")
	l, err := TryAcquire(path)
	if err != nil {
		t.Fatalf("TryAcquire: %v", err)
	}
	t.Cleanup(func() { _ = l.Release() })

	if l.Path() != path {
		t.Fatalf("Path() = %q, want %q", l.Path(), path)
	}
	// File should exist on disk and contain pid=...
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("lock file missing: %v", err)
	}
}

func TestTryAcquire_Conflict(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "condurad.lock")
	l, err := TryAcquire(path)
	if err != nil {
		t.Fatalf("first TryAcquire: %v", err)
	}
	t.Cleanup(func() { _ = l.Release() })

	_, err = TryAcquire(path)
	if !errors.Is(err, ErrLocked) {
		t.Fatalf("second TryAcquire err = %v, want ErrLocked", err)
	}
}

func TestTryAcquire_AfterRelease(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "condurad.lock")
	l1, err := TryAcquire(path)
	if err != nil {
		t.Fatalf("first TryAcquire: %v", err)
	}
	if err := l1.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}

	l2, err := TryAcquire(path)
	if err != nil {
		t.Fatalf("second TryAcquire after release: %v", err)
	}
	t.Cleanup(func() { _ = l2.Release() })
}

func TestTryAcquire_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "subdir", "lock")
	l, err := TryAcquire(path)
	if err != nil {
		t.Fatalf("TryAcquire: %v", err)
	}
	t.Cleanup(func() { _ = l.Release() })
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Fatalf("parent dir not created: %v", err)
	}
}

func TestTryAcquire_IdempotentRelease(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lock")
	l, err := TryAcquire(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := l.Release(); err != nil {
		t.Fatal(err)
	}
	// Second call must not panic and must not error (resources gone).
	if err := l.Release(); err != nil {
		t.Fatalf("second Release: %v", err)
	}
}

func TestTryAcquire_NilSafe(t *testing.T) {
	var l *Lock
	if err := l.Release(); err != nil {
		t.Fatalf("nil Release: %v", err)
	}
}

func TestIsInstalled_NotInstalled(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir) // Windows
	if IsInstalled() {
		t.Fatal("should not be installed on fresh temp dir")
	}
}

func TestMarkInstalled_ThenIsInstalled(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := MarkInstalled(); err != nil {
		t.Fatalf("MarkInstalled: %v", err)
	}
	if !IsInstalled() {
		t.Fatal("should be installed after MarkInstalled")
	}
}

func TestInstalledMarkerPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir) // Windows
	path, err := InstalledMarkerPath()
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(dir, ".condura", "installed")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}
}

// TestIsInstalled_ConduraAsRegularFileStillReturnsTrue documents
// the CURRENT (potentially-surprising) behavior: IsInstalled
// uses os.Stat, which returns success for ANY path type
// (file, directory, symlink). A regular file at ~/.condura
// would make IsInstalled return true. This is the documented
// contract — a future change to "check IsDir" would update
// this test deliberately.
//
// Pinning this prevents silent regression: if a future "fix"
// changes IsInstalled to check IsDir, the test fails and the
// dev is forced to read this comment.
func TestIsInstalled_ConduraAsRegularFileStillReturnsTrue(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // Windows

	// Create ~/.condura as a regular file (not a directory).
	condura := filepath.Join(home, ".condura")
	if err := os.WriteFile(condura, []byte("stale install state"), 0o600); err != nil {
		t.Fatal(err)
	}

	if !IsInstalled() {
		t.Error("IsInstalled() with ~/.condura as regular file = false; want true (current contract: any path)")
	}
}

// TestInstalledMarkerPath_ErrorPropagatesWhenHomeLookupFails pins
// the error-propagation contract: InstalledMarkerPath MUST
// surface the os.UserHomeDir() error when the lookup fails. A
// regression that swallowed the error would silently return a
// wrong path (the join on empty string gives "<no-home>/.condura/installed"
// which the installer would then write to silently).
//
// We simulate "HOME unset / unusable" by setting HOME to empty.
// On Linux/macOS, os.UserHomeDir() returns $HOME if non-empty, else
// falls back to /etc/passwd lookup, which usually succeeds for
// the test user. To FORCE the error path, we also need to break
// the fallbacks on Linux. Skipping on Linux when we can't force
// the error is the cleanest approach.
func TestInstalledMarkerPath_ErrorPropagatesWhenHomeLookupFails(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("HOME-unset fallback to /etc/passwd usually succeeds on Linux; skip")
	}
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	_, err := InstalledMarkerPath()
	if err == nil {
		t.Fatal("InstalledMarkerPath() with empty HOME returned nil; want error")
	}
}

// TestMarkInstalled_ErrorPropagatesWhenHomeLookupFails pins the
// same error-propagation contract for MarkInstalled.
func TestMarkInstalled_ErrorPropagatesWhenHomeLookupFails(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("HOME-unset fallback to /etc/passwd usually succeeds on Linux; skip")
	}
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	err := MarkInstalled()
	if err == nil {
		t.Fatal("MarkInstalled() with empty HOME returned nil; want error")
	}
}

// TestIsInstalled_FalseWhenConduraIsSymlinkToNonexistent pins the
// symlink edge case: IsInstalled uses Stat, which follows
// symlinks. If ~/.condura is a symlink to a nonexistent target,
// Stat returns an error → IsInstalled returns false. This is the
// expected behavior (the symlink is broken; treat as not installed).
func TestIsInstalled_FalseWhenConduraIsSymlinkToNonexistent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	// Create a symlink at ~/.condura pointing to a nonexistent path.
	condura := filepath.Join(home, ".condura")
	if err := os.Symlink("/nonexistent/path/that/does/not/exist", condura); err != nil {
		t.Fatal(err)
	}

	if IsInstalled() {
		t.Error("IsInstalled() with broken symlink = true; want false")
	}
}
