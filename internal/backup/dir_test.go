package backup

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
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
