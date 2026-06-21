package trust

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStore_NewStore_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(filepath.Join(dir, "trust.yaml"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if got := len(s.List()); got != 0 {
		t.Fatalf("empty store should have 0 entries, got %d", got)
	}
}

func TestStore_GrantLookupRevoke(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(filepath.Join(dir, "trust.yaml"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	// Grant
	e, err := s.Grant("/path/to/repo", "My Repo", DefaultScope)
	if err != nil {
		t.Fatalf("Grant: %v", err)
	}
	if e.WorkspaceID != "/path/to/repo" {
		t.Errorf("WorkspaceID: got %q, want %q", e.WorkspaceID, "/path/to/repo")
	}
	if !e.AlwaysAllow {
		t.Error("AlwaysAllow should be true after Grant")
	}

	// Lookup hit
	got := s.Lookup("/path/to/repo", "com.microsoft.VSCode")
	if got == nil {
		t.Fatal("Lookup should hit after Grant")
	}
	if got.Label != "My Repo" {
		t.Errorf("Label: got %q, want %q", got.Label, "My Repo")
	}

	// Lookup miss — different workspace
	if got := s.Lookup("/other/path", "com.microsoft.VSCode"); got != nil {
		t.Error("Lookup should miss on unknown workspace")
	}
}

func TestStore_AppScopeFiltering(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(filepath.Join(dir, "trust.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Grant("/repo", "Repo", AppScope("com.microsoft.VSCode")); err != nil {
		t.Fatal(err)
	}
	// VS Code is allowed.
	if got := s.Lookup("/repo", "com.microsoft.VSCode"); got == nil {
		t.Error("VS Code should be allowed by app_scope")
	}
	// Terminal is not.
	if got := s.Lookup("/repo", "com.apple.Terminal"); got != nil {
		t.Error("Terminal should NOT be allowed by app_scope")
	}
	// DefaultScope (empty) is a wildcard.
	if _, err := s.Grant("/repo2", "Repo2", DefaultScope); err != nil {
		t.Fatal(err)
	}
	if got := s.Lookup("/repo2", "com.apple.Terminal"); got == nil {
		t.Error("DefaultScope should be a wildcard")
	}
}

func TestStore_Revoke(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(filepath.Join(dir, "trust.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Grant("/repo", "Repo", DefaultScope); err != nil {
		t.Fatal(err)
	}
	if err := s.Revoke("/repo"); err != nil {
		t.Fatal(err)
	}
	if got := s.Lookup("/repo", ""); got != nil {
		t.Error("Revoked workspace should not match")
	}
}

func TestStore_Persistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "trust.yaml")
	s1, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s1.Grant("/repo", "Repo", DefaultScope); err != nil {
		t.Fatal(err)
	}

	// Reload from disk
	s2, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := s2.Lookup("/repo", ""); got == nil {
		t.Error("entry should survive a reload")
	}
}

func TestStore_GrantEmptyID(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(filepath.Join(dir, "trust.yaml"))
	if _, err := s.Grant("", "Label", DefaultScope); err == nil {
		t.Error("Grant with empty workspace ID should fail")
	}
}

func TestStore_LastUsedAtUpdated(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(filepath.Join(dir, "trust.yaml"))
	if _, err := s.Grant("/repo", "Repo", DefaultScope); err != nil {
		t.Fatal(err)
	}
	// Lookup mutates the underlying struct in place. We capture
	// the timestamp as a value before the second lookup.
	first := s.Lookup("/repo", "")
	if first == nil {
		t.Fatal("lookup should hit")
	}
	firstTime := first.LastUsedAt
	time.Sleep(10 * time.Millisecond)
	second := s.Lookup("/repo", "")
	if !second.LastUsedAt.After(firstTime) {
		t.Errorf("LastUsedAt should advance: first=%v second=%v", firstTime, second.LastUsedAt)
	}
}

func TestStore_ListSortedByRecency(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(filepath.Join(dir, "trust.yaml"))
	if _, err := s.Grant("/old", "Old", DefaultScope); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond)
	if _, err := s.Grant("/new", "New", DefaultScope); err != nil {
		t.Fatal(err)
	}
	list := s.List()
	if len(list) != 2 {
		t.Fatalf("List: got %d, want 2", len(list))
	}
	if list[0].WorkspaceID != "/new" {
		t.Errorf("expected /new first (more recent), got %q", list[0].WorkspaceID)
	}
}

func TestStore_RevokeNonexistent(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(filepath.Join(dir, "trust.yaml"))
	if err := s.Revoke("/nonexistent"); err != nil {
		t.Errorf("Revoke on nonexistent should be a no-op, got %v", err)
	}
}

func TestWorkspaceIDFor_GitRoot(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "myrepo")
	nested := filepath.Join(repoDir, "src", "lib")
	for _, p := range []string{repoDir, nested} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	gitPath := filepath.Join(repoDir, ".git")
	if err := os.MkdirAll(gitPath, 0o755); err != nil {
		t.Fatal(err)
	}
	got := WorkspaceIDFor(nested)
	if got != repoDir {
		t.Errorf("WorkspaceIDFor(%q) = %q, want %q", nested, got, repoDir)
	}
}

func TestWorkspaceIDFor_NoGitRoot(t *testing.T) {
	dir := t.TempDir()
	got := WorkspaceIDFor(dir)
	if got != dir {
		t.Errorf("WorkspaceIDFor (no .git): got %q, want %q", got, dir)
	}
}

func TestWorkspaceIDFor_EmptyInput(t *testing.T) {
	if got := WorkspaceIDFor(""); got != "" {
		t.Errorf("WorkspaceIDFor empty: got %q, want \"\"", got)
	}
}

func TestWorkspaceIDFor_FilesystemRoot(t *testing.T) {
	// Walk up past / to / — must not infinite-loop.
	if got := WorkspaceIDFor("/"); got == "" {
		t.Error("WorkspaceIDFor(/) should return something, not empty")
	}
}

func TestParseTrustYAML_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(filepath.Join(dir, "trust.yaml"))
	if _, err := s.Grant("/repo1", "Repo One", DefaultScope); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Grant("/repo2", "Repo Two", AppScope("com.microsoft.VSCode")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "trust.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	entries := parseTrustYAML(string(data))
	if len(entries) != 2 {
		t.Fatalf("parseTrustYAML: got %d entries, want 2", len(entries))
	}
	// Map iteration order is non-deterministic; index by ID.
	byID := make(map[string]*Entry, len(entries))
	for _, e := range entries {
		byID[e.WorkspaceID] = e
	}
	r1, ok := byID["/repo1"]
	if !ok {
		t.Fatalf("missing /repo1 entry: %+v", byID)
	}
	if r1.Label != "Repo One" || r1.AppScope != DefaultScope {
		t.Errorf("/repo1: label=%q app_scope=%q", r1.Label, r1.AppScope)
	}
	r2, ok := byID["/repo2"]
	if !ok {
		t.Fatalf("missing /repo2 entry: %+v", byID)
	}
	if r2.Label != "Repo Two" || r2.AppScope != AppScope("com.microsoft.VSCode") {
		t.Errorf("/repo2: label=%q app_scope=%q", r2.Label, r2.AppScope)
	}
}

func TestParseTrustYAML_IgnoresComments(t *testing.T) {
	yaml := `# Top comment
# another comment
- workspace_id: "/r1"
  label: "r1"
  always_allow: true
  created_at: 2026-01-01T00:00:00Z
  last_used_at: 2026-01-02T00:00:00Z
`
	entries := parseTrustYAML(yaml)
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if !entries[0].AlwaysAllow {
		t.Error("AlwaysAllow should parse as true")
	}
}

func TestParseTrustYAML_EmptyString(t *testing.T) {
	if entries := parseTrustYAML(""); len(entries) != 0 {
		t.Errorf("empty input should yield 0 entries, got %d", len(entries))
	}
	if entries := parseTrustYAML("# just a comment\n\n"); len(entries) != 0 {
		t.Errorf("comment-only should yield 0 entries, got %d", len(entries))
	}
}

// Anti-regression: the parser must handle the actual file we
// produce. Catches escaping / quoting drift between writer and
// reader.
func TestParseTrustYAML_RoundTrip_Honest(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(filepath.Join(dir, "trust.yaml"))
	if _, err := s.Grant("/path/with spaces and \"quotes\"", "Label With \"Quotes\"", AppScope("com.app.with spaces")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "trust.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	entries := parseTrustYAML(string(data))
	if len(entries) != 1 {
		t.Fatalf("got %d, want 1", len(entries))
	}
	e := entries[0]
	if e.WorkspaceID != "/path/with spaces and \"quotes\"" {
		t.Errorf("WorkspaceID roundtrip: got %q", e.WorkspaceID)
	}
	if e.Label != "Label With \"Quotes\"" {
		t.Errorf("Label roundtrip: got %q", e.Label)
	}
	if e.AppScope != AppScope("com.app.with spaces") {
		t.Errorf("AppScope roundtrip: got %q", e.AppScope)
	}
}

// Sanity: when the trust store Save() returns no error, the on-disk
// file should have the same number of entries as the in-memory map.
// Catches half-saved writes.
func TestStore_SaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "trust.yaml")
	s, _ := NewStore(path)
	for i := 0; i < 5; i++ {
		if _, err := s.Grant(filepath.Join("/repo", string(rune('a'+i))), "Label", DefaultScope); err != nil {
			t.Fatal(err)
		}
	}
	// Read raw file and count "- workspace_id" markers.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	count := strings.Count(string(data), "- workspace_id:")
	if count != 5 {
		t.Errorf("on-disk markers: got %d, want 5", count)
	}
}
