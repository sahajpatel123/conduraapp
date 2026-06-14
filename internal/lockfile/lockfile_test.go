package lockfile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestTryAcquire_Fresh(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "synapticd.lock")
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
	path := filepath.Join(dir, "synapticd.lock")
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
	path := filepath.Join(dir, "synapticd.lock")
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
	expected := filepath.Join(dir, ".synaptic", "installed")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}
}
