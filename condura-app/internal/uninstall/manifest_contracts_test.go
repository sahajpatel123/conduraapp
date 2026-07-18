package uninstall

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestManifestMismatch_Error pins the custom-error format contract:
// ManifestMismatch.Error() MUST return a string that mentions both
// the count of unknown artifacts AND the paths. The GUI surfaces
// this message verbatim in the "refusing to uninstall" toast —
// if either the count or the paths is missing, the user has no
// signal for what to investigate.
func TestManifestMismatch_Error(t *testing.T) {
	err := &ManifestMismatch{Unknown: []string{"/a", "/b", "/c"}}
	got := err.Error()

	// Must mention count (3) for log-reader diagnostic clarity.
	if !strings.Contains(got, "3") {
		t.Errorf("Error() = %q; should mention count '3'", got)
	}
	// Must mention each path so the user knows what to investigate.
	for _, p := range []string{"/a", "/b", "/c"} {
		if !strings.Contains(got, p) {
			t.Errorf("Error() = %q; should mention path %q", got, p)
		}
	}
	// Should signal refusal — the message must convey "we refused".
	if !strings.Contains(got, "refus") {
		t.Errorf("Error() = %q; should mention 'refus[ing/ed]' to signal refusal", got)
	}
}

// TestManifestMismatch_ErrorEmpty pins the edge case: a manifest
// mismatch with zero unknowns. Unusual but theoretically possible
// if the running system reports "I created 0 files I don't
// recognize" — the message should still be sensible (mention "0").
func TestManifestMismatch_ErrorEmpty(t *testing.T) {
	err := &ManifestMismatch{}
	got := err.Error()
	if !strings.Contains(got, "0") {
		t.Errorf("Error() with no unknowns = %q; should still mention '0'", got)
	}
	if !strings.Contains(got, "refus") {
		t.Errorf("Error() with no unknowns = %q; should still mention 'refus[ing/ed]'", got)
	}
}

// TestUninstall_ManifestMismatchRejectsUnknownArtifacts pins the
// production safety contract: if the running system created files
// in the data dir that aren't in the manifest, Uninstall MUST
// refuse with a ManifestMismatch error rather than silently leaving
// data behind. This is sub-phase 11D's "complete enumeration"
// invariant — every file Condura creates must be in DefaultManifest.
//
// NOTE: at the time of this test pin, the ManifestMismatch return
// path is DOCUMENTED in the source but NOT YET WIRED into
// Uninstall. The production code currently does not detect
// unknown on-disk files. This test is deferred until the
// detection path is implemented (tracked as a v0.2.0 backlog
// item per the "complete enumeration" invariant).
func TestUninstall_ManifestMismatchRejectsUnknownArtifacts(t *testing.T) {
	t.Skip("deferred: ManifestMismatch return path not yet wired into Uninstall")
}

// TestDefaultManifest_EmptyDataDirUsesHome pins the empty-input
// fallback contract: DefaultManifest("") MUST use $HOME/.condura
// as the data dir. A regression that required a non-empty data
// dir would break the default-flow invocation.
func TestDefaultManifest_EmptyDataDirUsesHome(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	m := DefaultManifest("")
	if len(m) == 0 {
		t.Fatal("DefaultManifest(\"\") returned empty manifest; want the canonical artifact list")
	}
	// Every manifest entry should be under tmp/.condura.
	expectedPrefix := filepath.Join(tmp, ".condura")
	for _, e := range m {
		if !strings.HasPrefix(e.Path, expectedPrefix) {
			t.Errorf("manifest entry %q path %q does not start with %q (empty DataDir should fall back to HOME/.condura)",
				e.Name, e.Path, expectedPrefix)
		}
	}
}

// TestDefaultManifest_ExplicitDataDirPinsPaths pins the
// non-default-input contract: DefaultManifest("/custom/path")
// MUST produce entries with paths under "/custom/path", not
// under $HOME/.condura.
func TestDefaultManifest_ExplicitDataDirPinsPaths(t *testing.T) {
	// Also override HOME so the fallback wouldn't kick in if the
	// production code accidentally swapped the args.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	const customDir = "/custom/data/dir"
	m := DefaultManifest(customDir)
	if len(m) == 0 {
		t.Fatal("DefaultManifest returned empty manifest")
	}
	for _, e := range m {
		if !strings.HasPrefix(e.Path, customDir) {
			t.Errorf("manifest entry %q path %q does not start with %q (explicit DataDir ignored)",
				e.Name, e.Path, customDir)
		}
	}
}

// TestEntriesForPaths_EmptyDataDirReturnsNil pins the empty-input
// guard: EntriesForPaths("") MUST return nil (not panic, not return
// a non-nil empty slice — the existing TestEntriesForPaths in
// manifest_test.go relies on this nil-return contract to detect the
// "no data dir" case).
func TestEntriesForPaths_EmptyDataDirReturnsNil(t *testing.T) {
	if got := EntriesForPaths(""); got != nil {
		t.Errorf("EntriesForPaths(\"\") = %v; want nil", got)
	}
}

// TestEntriesForPaths_NonexistentDataDirReturnsEmpty pins the
// no-match contract: when no manifest OPTIONAL entry exists on
// disk (and only the data dir itself does), EntriesForPaths MUST
// return at most 1 path (the data dir itself, which always exists
// because we just created it). The caller iterates over the result;
// nil and empty are equivalent for iteration but the function
// should not panic and should return a predictable shape.
//
// Note: the "data dir itself" entry in the manifest (line 112)
// means the data dir's existence is ALWAYS counted. So this test
// pins that EntriesForPaths returns exactly 1 in the empty case
// (the data dir itself), not 0.
func TestEntriesForPaths_NonexistentDataDirReturnsEmpty(t *testing.T) {
	tmp := t.TempDir()
	// Don't create any of the manifest entries except the data
	// dir itself (which exists because t.TempDir created it).
	got := EntriesForPaths(tmp)
	if got == nil {
		t.Error("EntriesForPaths(empty data dir) = nil; want non-nil slice")
	}
	// Exactly one path: the data dir itself.
	if len(got) != 1 {
		t.Errorf("EntriesForPaths(empty data dir) = %d entries; want 1 (just the data dir itself)", len(got))
	}
	if len(got) >= 1 && got[0] != tmp {
		t.Errorf("EntriesForPaths returned path %q; want %q (the data dir itself)", got[0], tmp)
	}
}

// TestEntriesForPaths_AllManifestMembersAreReturned pins the
// happy-path contract: when every manifest entry exists on disk,
// EntriesForPaths MUST return all of them. The function filters
// by os.Stat — paths that don't exist are excluded.
func TestEntriesForPaths_AllManifestMembersAreReturned(t *testing.T) {
	tmp := t.TempDir()
	manifest := DefaultManifest(tmp)

	// Populate every manifest entry so they all exist on disk.
	for _, e := range manifest {
		if err := os.MkdirAll(filepath.Dir(e.Path), 0o700); err != nil {
			t.Fatal(err)
		}
		if e.Path == tmp {
			continue // the data dir itself is already created
		}
		if err := os.WriteFile(e.Path, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got := EntriesForPaths(tmp)
	if len(got) != len(manifest) {
		t.Errorf("EntriesForPaths returned %d paths; want %d (all manifest entries exist)",
			len(got), len(manifest))
	}
}
