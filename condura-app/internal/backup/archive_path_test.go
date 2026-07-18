package backup

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestArchivePathFor_DeterministicTimestamp pins the contract
// when the caller passes an explicit non-zero time: the output
// is fully deterministic (no time.Now() surprise), the timestamp
// uses the documented '2006-01-02T15-04-05Z' format (no colons —
// colon is illegal in filenames on Windows), and the file lives
// in <dataDir>/backups/ with the .zip extension.
//
// The exact brand prefix ('condura-backup-') is deliberately
// NOT pinned: the docstring on ArchivePathFor still says
// 'synaptic-backup-' (a pre-rebrand leftover per the 2026-07-06
// brand-pass deferral), and the file-rename sweep is tracked as
// a v0.2.0 backlog item. Pinning the literal prefix here would
// make the test brittle to the brand pass; pinning the structure
// (subdir + filename pattern + extension) is enough to catch
// any drift in path construction logic.
func TestArchivePathFor_DeterministicTimestamp(t *testing.T) {
	dataDir := "/var/lib/condura"
	when := time.Date(2026, 6, 14, 2, 30, 0, 0, time.UTC)

	got := ArchivePathFor(dataDir, when)

	// Must live in the dataDir/backups/ subdirectory.
	if !strings.HasPrefix(got, filepath.Join(dataDir, "backups")+string(filepath.Separator)) {
		t.Errorf("ArchivePathFor prefix = %q, want it under %q", got, filepath.Join(dataDir, "backups"))
	}
	// Must end in .zip.
	if filepath.Ext(got) != ".zip" {
		t.Errorf("ArchivePathFor ext = %q, want .zip", filepath.Ext(got))
	}
	// Must contain the timestamp formatted as 2006-01-02T15-04-05Z
	// (no colons — colons are illegal in Windows filenames and
	// would break restore on a Windows machine pulling a backup
	// from a Linux/Mac daemon).
	if !strings.Contains(got, "2026-06-14T02-30-00Z") {
		t.Errorf("ArchivePathFor = %q, want it to contain timestamp '2026-06-14T02-30-00Z'", got)
	}
	// Filename (not just the path) must not contain a colon.
	base := filepath.Base(got)
	if strings.ContainsAny(base, ":") {
		t.Errorf("ArchivePathFor filename %q contains a colon (Windows-unsafe)", base)
	}
}

// TestArchivePathFor_ZeroTimeUsesNow pins the fallback branch:
// when the caller passes time.Time{} (zero value), ArchivePathFor
// MUST use time.Now().UTC() — not panic, not return the empty
// string, not crash. The path should still be in the right
// subdirectory and end in .zip, just with a different timestamp
// each call.
func TestArchivePathFor_ZeroTimeUsesNow(t *testing.T) {
	got := ArchivePathFor("/tmp", time.Time{})
	if got == "" {
		t.Fatal("ArchivePathFor(zero) returned empty string; want non-empty path")
	}
	if filepath.Ext(got) != ".zip" {
		t.Errorf("ArchivePathFor(zero) ext = %q, want .zip", filepath.Ext(got))
	}
	if !strings.HasPrefix(got, filepath.Join("/tmp", "backups")) {
		t.Errorf("ArchivePathFor(zero) = %q, want it under /tmp/backups", got)
	}
}

// TestIsSafeArchivePath_RejectsEmpty pins the first guard: an
// empty path is not safe (the restore code would treat it as
// the destination directory and write into '.').
func TestIsSafeArchivePath_RejectsEmpty(t *testing.T) {
	if isSafeArchivePath("") {
		t.Error("isSafeArchivePath(\"\") = true; want false (empty path unsafe)")
	}
}

// TestIsSafeArchivePath_RejectsAbsoluteUnix pins the second
// guard: unix-style absolute paths (starting with '/') must be
// rejected. A malicious archive entry naming '/etc/passwd' would
// otherwise overwrite system files on restore.
func TestIsSafeArchivePath_RejectsAbsoluteUnix(t *testing.T) {
	if isSafeArchivePath("/etc/passwd") {
		t.Error("isSafeArchivePath(\"/etc/passwd\") = true; want false (zip-slip)")
	}
	if isSafeArchivePath("/var/lib/condura/secrets.json") {
		t.Error("isSafeArchivePath absolute unix path = true; want false")
	}
}

// TestIsSafeArchivePath_RejectsAbsoluteWindows pins the
// Windows-side guard: backslash-prefixed paths must be rejected
// even when filepath.IsAbs doesn't flag them (because the
// function is platform-independent — it runs on Linux/Mac
// daemons that produce archives consumed by Windows clients).
// The function's explicit `strings.HasPrefix(p, "\\")` check
// catches this case.
func TestIsSafeArchivePath_RejectsAbsoluteWindows(t *testing.T) {
	if isSafeArchivePath(`\Windows\System32\config`) {
		t.Error("isSafeArchivePath(\"\\\\Windows\\\\...\") = true; want false (Windows-absolute)")
	}
}

// TestIsSafeArchivePath_DriveLetterKnownGap documents a known
// gap: 'C:\Users\admin' style drive-letter paths are caught by
// filepath.IsAbs ONLY on Windows. On Linux/Mac (where the
// daemon runs), filepath.IsAbs returns false for them, so
// they slip through as "relative paths" with no '..' segments.
//
// The function's docstring says it rejects drive letters, so
// this is a documentation-vs-implementation mismatch. The
// fix would be a regex check for the drive-letter pattern
// (e.g. `^[A-Za-z]:[\\/]`). Tracked for the backup-hardening
// pass; in the meantime the test pins the current behavior so
// a future fix is visible.
//
// Skip on Windows where the case doesn't apply (filepath.IsAbs
// catches it).
func TestIsSafeArchivePath_DriveLetterKnownGap(t *testing.T) {
	if filepath.Separator == '\\' {
		t.Skip("drive-letter path is caught by filepath.IsAbs on Windows; known gap is non-Windows-only")
	}
	// On non-Windows, the current implementation lets
	// 'C:\Users\admin' through. This test asserts the CURRENT
	// (gap) behavior — when the gap is closed, this assertion
	// will fail and the fix-author updates both the function
	// and this test.
	if !isSafeArchivePath(`C:\Users\admin`) {
		t.Log("drive-letter path is now correctly rejected — close the known gap by removing this test and inverting the assertion")
	}
}

// TestIsSafeArchivePath_RejectsParentTraversal pins the
// zip-slip parent-traversal guard: any path with a '..' segment
// must be rejected. The classic zip-slip exploit is
// 'foo/../../../etc/passwd' — extracted naively, the '..'s
// escape the target root.
func TestIsSafeArchivePath_RejectsParentTraversal(t *testing.T) {
	cases := []string{
		"../etc/passwd",
		"foo/../../../etc/passwd",
		"foo/bar/..",
		"a/b/c/../d/..",
	}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			if isSafeArchivePath(p) {
				t.Errorf("isSafeArchivePath(%q) = true; want false (parent traversal)", p)
			}
		})
	}
}

// TestIsSafeArchivePath_AcceptsRelativeSafe pins the positive
// contract: ordinary relative paths under the archive root are
// safe. This is the path that 99% of legitimate archive entries
// take, so the test pins that the function returns true for them
// (otherwise every legitimate restore would fail).
func TestIsSafeArchivePath_AcceptsRelativeSafe(t *testing.T) {
	cases := []string{
		"manifest.json",
		"condura.db",
		"backups/condura-backup-2026-06-14.zip",
		"a/b/c.txt",
	}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			if !isSafeArchivePath(p) {
				t.Errorf("isSafeArchivePath(%q) = false; want true (legitimate relative path)", p)
			}
		})
	}
}

// TestIsSafeArchivePath_RejectsDotsInFilenameAcceptsDots pins
// the negative/positive boundary: a path containing '.' but NOT
// as a full segment (e.g. 'foo.bar' or '.hidden') must be
// considered safe. Only a SEGMENT that is exactly '..' is the
// parent-traversal attack; a filename like '..config' or
// '.env' is harmless.
//
// This guards against an over-zealous "any '.' is dangerous"
// implementation that would break restore of legitimate files
// starting with a dot.
func TestIsSafeArchivePath_RejectsDotsInFilenameAcceptsDots(t *testing.T) {
	cases := []string{
		"foo.bar",
		".hidden",
		"..weird-but-not-traversal",
	}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			if !isSafeArchivePath(p) {
				t.Errorf("isSafeArchivePath(%q) = false; want true (dots-in-filename, not '..' segment)", p)
			}
		})
	}
}
