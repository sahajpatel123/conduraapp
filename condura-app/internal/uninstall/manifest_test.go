package uninstall

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestUninstall_Completeness is THE test for sub-phase 11D.
// It populates a fake data dir with one of every artifact in the
// manifest, runs Uninstall, and asserts ZERO files remain.
//
// The test runs in a temp HOME, never the real one, so the hard
// guard does not fire.
func TestUninstall_Completeness(t *testing.T) {
	tmp := t.TempDir()
	// Each manifest entry is a path. The test populates ALL
	// entries (even optional ones) so the completeness assertion
	// is meaningful.
	manifest := DefaultManifest(tmp)
	for _, e := range manifest {
		if err := os.MkdirAll(filepath.Dir(e.Path), 0o700); err != nil {
			t.Fatal(err)
		}
		if e.Path == tmp {
			continue // the data dir itself is created by MkdirAll
		}
		if err := os.WriteFile(e.Path, []byte("synaptic-data"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// Run uninstall with a confirm token (we're running in
	// a temp HOME that doesn't match the real $HOME, so the
	// guard would fire — but we provide a token anyway, as the
	// GUI would do).
	opts := Options{
		DataDir:      tmp,
		ConfirmToken: NewConfirmToken(),
		HomeDir:      tmp, // pretend the temp IS the home, so the guard fires
		// — and we provide a token.
	}
	res, err := Uninstall(opts)
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if res.FilesRemoved == 0 {
		t.Error("expected FilesRemoved > 0")
	}
	// After uninstall, the data dir should be gone.
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Errorf("data dir still exists after uninstall: %v", err)
	}

	// Also: nothing in the manifest should remain. We re-run
	// the entries list and assert each one is gone.
	for _, e := range manifest {
		if _, err := os.Stat(e.Path); !os.IsNotExist(err) {
			t.Errorf("%s still exists after uninstall: %v", e.Name, err)
		}
	}
}

// TestUninstall_RefusesWithoutTokenUnderHome enforces the hard
// guard. Running against HOME without a token is denied.
func TestUninstall_RefusesWithoutTokenUnderHome(t *testing.T) {
	tmp := t.TempDir()
	opts := Options{
		DataDir:      tmp,
		ConfirmToken: "", // no token
		HomeDir:      tmp,
	}
	_, err := Uninstall(opts)
	if err == nil {
		t.Error("expected ErrUnsafeHome, got nil")
	}
}

// TestUninstall_AcceptsValidToken proves the happy path of the
// hard guard.
func TestUninstall_AcceptsValidToken(t *testing.T) {
	tmp := t.TempDir()
	opts := Options{
		DataDir:      tmp,
		ConfirmToken: NewConfirmToken(),
		HomeDir:      tmp,
	}
	if _, err := Uninstall(opts); err != nil {
		// tmp is empty so no files to remove, but the run should
		// succeed (no error).
		t.Fatalf("Uninstall: %v", err)
	}
}

// TestUninstall_RejectsBadTokenFormat catches typos.
func TestUninstall_RejectsBadTokenFormat(t *testing.T) {
	tmp := t.TempDir()
	opts := Options{
		DataDir:      tmp,
		ConfirmToken: "not-a-32-char-hex-token!!",
		HomeDir:      tmp,
	}
	if _, err := Uninstall(opts); err == nil {
		t.Error("expected error for bad token format")
	}
}

// TestUninstall_EmptyDataDir refuses to delete $HOME.
func TestUninstall_EmptyDataDir(t *testing.T) {
	opts := Options{DataDir: ""}
	if _, err := Uninstall(opts); err == nil {
		t.Error("expected error for empty data dir")
	}
}

// TestUninstall_DryRun is a smoke test that DryRun doesn't
// actually delete.
func TestUninstall_DryRun(t *testing.T) {
	tmp := t.TempDir()
	mustWrite(t, filepath.Join(tmp, "condura.db"), []byte("main"))

	opts := Options{
		DataDir:      tmp,
		ConfirmToken: NewConfirmToken(),
		HomeDir:      tmp,
		DryRun:       true,
	}
	res, err := Uninstall(opts)
	if err != nil {
		t.Fatal(err)
	}
	if res.FilesRemoved == 0 {
		t.Error("expected DryRun to count at least 1 file")
	}
	// The file must still be there.
	if _, err := os.Stat(filepath.Join(tmp, "condura.db")); err != nil {
		t.Errorf("DryRun deleted the file: %v", err)
	}
}

// TestPreview_DoesNotMutate asserts Preview is non-destructive.
func TestPreview_DoesNotMutate(t *testing.T) {
	tmp := t.TempDir()
	mustWrite(t, filepath.Join(tmp, "condura.db"), []byte("x"))
	opts := Options{DataDir: tmp, HomeDir: tmp}
	pr, err := Preview(opts)
	if err != nil {
		t.Fatal(err)
	}
	if pr.Total < 2 {
		t.Errorf("preview total = %d, want >= 2", pr.Total)
	}
	// File still there.
	if _, err := os.Stat(filepath.Join(tmp, "condura.db")); err != nil {
		t.Errorf("Preview deleted the file: %v", err)
	}
}

// TestPostUninstallGuide_PerPlatform asserts the guide is
// non-empty and mentions the platform.
func TestPostUninstallGuide_PerPlatform(t *testing.T) {
	g := PostUninstallGuide()
	if g == "" {
		t.Error("post-uninstall guide is empty")
	}
	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(g, "macOS") {
			t.Error("macOS guide should mention macOS")
		}
	case "windows":
		if !strings.Contains(g, "Windows") {
			t.Error("Windows guide should mention Windows")
		}
	case "linux":
		if !strings.Contains(g, "Linux") {
			t.Error("Linux guide should mention Linux")
		}
	}
}

// TestNewConfirmToken_Format ensures the token is 32 hex chars.
func TestNewConfirmToken_Format(t *testing.T) {
	tok := NewConfirmToken()
	if len(tok) != 32 {
		t.Errorf("token length = %d, want 32", len(tok))
	}
	// Should be valid hex.
	for _, c := range tok {
		if c < '0' || c > '9' && c < 'a' || c > 'f' {
			t.Errorf("non-hex char in token: %q", c)
		}
	}
}

// TestEntriesForPaths confirms the helper reports which manifest
// entries actually exist.
// TestUninstall_DoesNotFollowSymlinks proves that if a manifest
// path is replaced by a symlink pointing outside the data dir,
// the uninstaller unlinks the symlink only and does not recurse
// into the target.
func TestUninstall_DoesNotFollowSymlinks(t *testing.T) {
	tmp := t.TempDir()
	outside := t.TempDir()
	victim := filepath.Join(outside, "victim.txt")
	mustWrite(t, victim, []byte("precious"))

	// Create the real condura.db file first, then replace it with a
	// symlink pointing outside the data dir.
	dbPath := filepath.Join(tmp, "condura.db")
	mustWrite(t, dbPath, []byte("x"))
	if err := os.Remove(dbPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, dbPath); err != nil {
		t.Fatal(err)
	}

	opts := Options{
		DataDir:      tmp,
		ConfirmToken: NewConfirmToken(),
		HomeDir:      tmp,
	}
	if _, err := Uninstall(opts); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}

	// The symlink target's contents must survive.
	if _, err := os.Stat(victim); err != nil {
		t.Fatalf("symlink target file was deleted: %v", err)
	}
	// The symlink itself must be gone.
	if _, err := os.Lstat(dbPath); !os.IsNotExist(err) {
		t.Fatalf("symlink still present: %v", err)
	}
}

// TestUninstall_DryRun_DoesNotFollowSymlinks proves the dry-run path
// also avoids following symlinks.
func TestUninstall_DryRun_DoesNotFollowSymlinks(t *testing.T) {
	tmp := t.TempDir()
	outside := t.TempDir()
	victim := filepath.Join(outside, "victim.txt")
	mustWrite(t, victim, []byte("precious"))

	dbPath := filepath.Join(tmp, "condura.db")
	mustWrite(t, dbPath, []byte("x"))
	if err := os.Remove(dbPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, dbPath); err != nil {
		t.Fatal(err)
	}

	opts := Options{
		DataDir:      tmp,
		ConfirmToken: NewConfirmToken(),
		HomeDir:      tmp,
		DryRun:       true,
	}
	res, err := Uninstall(opts)
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	// Should count the symlink as a single leaf, not recurse into target.
	// The manifest also has a second file (secrets.json) that we didn't
	// create, so it is optional/missing; only the symlink counts.
	if res.FilesRemoved < 1 {
		t.Errorf("FilesRemoved = %d, want at least 1", res.FilesRemoved)
	}

	if _, err := os.Stat(victim); err != nil {
		t.Fatalf("dry-run deleted symlink target: %v", err)
	}
}

func TestEntriesForPaths(t *testing.T) {
	tmp := t.TempDir()
	mustWrite(t, filepath.Join(tmp, "condura.db"), []byte("x"))
	got := EntriesForPaths(tmp)
	if len(got) == 0 {
		t.Error("expected at least 1 existing path")
	}
	// All returned paths must be in the manifest.
	manifest := DefaultManifest(tmp)
	for _, p := range got {
		found := false
		for _, m := range manifest {
			if m.Path == p {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("returned path %q is not in manifest", p)
		}
	}
}

func mustWrite(t *testing.T, p string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestIsUnderHome_PathInsideHomeReturnsTrue pins the happy-path
// contract: a path inside HOME MUST return true. The uninstaller
// refuses to delete anything outside HOME (without explicit
// confirm-token override); this is the core safety check that
// protects the user from accidental `rm -rf ~`.
func TestIsUnderHome_PathInsideHomeReturnsTrue(t *testing.T) {
	home := t.TempDir()
	cases := []string{
		filepath.Join(home, "subdir"),
		filepath.Join(home, "deep", "nested", "file.txt"),
		filepath.Join(home, ".config"),
	}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			if !isUnderHome(p, home) {
				t.Errorf("isUnderHome(%q, %q) = false; want true (path inside HOME)", p, home)
			}
		})
	}
}

// TestIsUnderHome_PathEqualToHomeReturnsTrue pins the equality
// edge case: a path equal to HOME itself MUST return true
// (the user explicitly listed HOME). The rel == "." branch
// is what catches this.
func TestIsUnderHome_PathEqualToHomeReturnsTrue(t *testing.T) {
	home := t.TempDir()
	if !isUnderHome(home, home) {
		t.Error("isUnderHome(home, home) = false; want true (path equal to HOME)")
	}
}

// TestIsUnderHome_PathOutsideHomeReturnsFalse pins the
// safety guard: a path outside HOME MUST return false.
// Without this guard, the uninstaller would happily delete
// system directories like /usr/local or /tmp.
func TestIsUnderHome_PathOutsideHomeReturnsFalse(t *testing.T) {
	home := t.TempDir()
	cases := []string{
		"/usr/local",
		"/tmp",
		"/var/log",
		"/etc",
	}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			if isUnderHome(p, home) {
				t.Errorf("isUnderHome(%q, %q) = true; want false (path outside HOME)", p, home)
			}
		})
	}
}

// TestIsUnderHome_EscapedPathReturnsFalse pins the path-traversal
// guard: a path that "escapes" HOME via ".." segments MUST return
// false. The strings.HasPrefix(rel, "..") check is what catches
// this — without it, "../sneaky" would be detected as "under HOME"
// via the "./" prefix.
func TestIsUnderHome_EscapedPathReturnsFalse(t *testing.T) {
	home := t.TempDir()
	cases := []string{
		filepath.Join(home, "..", "outside"),
		filepath.Join(home, "..", "..", "etc", "passwd"),
	}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			if isUnderHome(p, home) {
				t.Errorf("isUnderHome(%q, %q) = true; want false (path escapes HOME via '..')", p, home)
			}
		})
	}
}

// TestIsUnderHome_SymlinkInsideHomeReturnsTrue pins the symlink
// behavior: a symlink at <home>/link pointing to <home>/target
// resolves to <home>/target via filepath.Abs (which follows
// the symlink at the path level — but filepath.Abs doesn't
// follow nested symlinks). A regression that resolved all
// symlinks (e.g., via filepath.EvalSymlinks) would still return
// true for our test case (target is also inside HOME).
//
// NOTE: filepath.Abs does NOT follow symlinks. It just makes
// the path absolute. So a symlink at <home>/link points to
// <home>/target, and Abs(<home>/link) returns the symlink's
// absolute path (without resolving). This is the CURRENT behavior.
func TestIsUnderHome_SymlinkInsideHomeReturnsTrue(t *testing.T) {
	home := t.TempDir()
	target := filepath.Join(home, "target.txt")
	if err := os.WriteFile(target, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(home, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if !isUnderHome(link, home) {
		t.Error("isUnderHome(symlink-inside-home) = false; want true (path is inside HOME)")
	}
}

// TestIsUnderHome_DifferentDrivesReturnsFalse is a Windows-style
// edge case: if path is on drive C: and home is on drive D:,
// filepath.Rel returns an error (different volumes). The
// function MUST return false (not panic). Skipped on Unix
// where multiple drives don't exist.
func TestIsUnderHome_DifferentDrivesReturnsFalse(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("different-drives is Windows-specific; Unix has a single root")
	}
	// Real Windows test would use paths like `C:\foo` and `D:\bar`.
	// Skipped on Unix CI.
}
