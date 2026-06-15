package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveBackupDir_DefaultDocuments(t *testing.T) {
	t.Setenv("SYNAPTIC_BACKUP_DIR", "")
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := ResolveBackupDir("/var/synaptic")
	want := filepath.Join(home, "Documents", "synaptic-backups")
	if dir != want {
		t.Fatalf("got %q, want %q", dir, want)
	}
}

func TestResolveBackupDir_EnvOverride(t *testing.T) {
	custom := t.TempDir()
	t.Setenv("SYNAPTIC_BACKUP_DIR", custom)
	dir := ResolveBackupDir("/var/synaptic")
	if dir != custom {
		t.Fatalf("got %q, want %q", dir, custom)
	}
}

func TestResolveBackupDir_FallbackDataDir(t *testing.T) {
	t.Setenv("SYNAPTIC_BACKUP_DIR", "")
	// No HOME — force fallback
	t.Setenv("HOME", "")
	dir := ResolveBackupDir("/var/synaptic")
	want := filepath.Join("/var/synaptic", "backups")
	if dir != want {
		t.Fatalf("got %q, want %q", dir, want)
	}
	_ = os.MkdirAll(dir, 0o700)
}
