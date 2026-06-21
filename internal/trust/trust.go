// Package trust implements per-workspace trust for the agent.
//
// Phase 16, Rec 5: when a project has been "always trusted" by the
// user, WRITE actions targeting that workspace skip the consent
// dialog. Without this, the default-deny posture would block every
// WRITE in VS Code on first run — killing the developer experience.
//
// Design:
//   - Workspace identifier: walk up from the action target's path to
//     find a .git/ directory; if found, the absolute path to that
//     directory is the workspace ID. Otherwise use the absolute path
//     of the action's target (most conservative fallback).
//   - Trust store: a YAML file at <data-dir>/trusted_workspaces.yaml
//     keyed by workspace ID. Entries are added by the consent
//     handler when the user picks "Always allow in this folder".
//   - Lookup is a single map read; safe for concurrent use.
//   - The trust store is consulted ONLY for WRITE actions in
//     non-DESTRUCTIVE classes. DESTRUCTIVE always requires fresh
//     consent per Survival Rule §2.
package trust

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AppScope is the per-app autonomy override granted by the user.
// Empty means "all apps in this workspace". A specific bundle ID
// (e.g. "com.microsoft.VSCode") means "only this app".
type AppScope string

// DefaultScope means "any app in this workspace gets the trust".
const DefaultScope AppScope = ""

// Entry is one trusted workspace.
type Entry struct {
	// WorkspaceID is the absolute path used as the lookup key.
	WorkspaceID string `yaml:"workspace_id"`
	// Label is the human-readable name shown in the GUI's Settings.
	// Defaults to the workspace ID's base name.
	Label string `yaml:"label"`
	// AlwaysAllow means the user picked "Always allow in this folder".
	AlwaysAllow bool `yaml:"always_allow"`
	// CreatedAt is when the trust was granted.
	CreatedAt time.Time `yaml:"created_at"`
	// LastUsedAt is updated on each lookup. Useful for the
	// "trust these workspaces" Settings view that sorts by recency.
	LastUsedAt time.Time `yaml:"last_used_at"`
	// AppScope restricts the trust to a specific app bundle ID.
	// Empty = all apps.
	AppScope AppScope `yaml:"app_scope,omitempty"`
}

// Store is the in-memory trust store backed by a YAML file.
type Store struct {
	mu   sync.RWMutex
	path string
	data map[string]*Entry // keyed by WorkspaceID
}

// NewStore loads (or creates) the trust store at path. The path
// should live in the daemon's data dir.
func NewStore(path string) (*Store, error) {
	s := &Store{path: path, data: make(map[string]*Entry)}
	if err := s.load(); err != nil {
		return nil, fmt.Errorf("trust: load %s: %w", path, err)
	}
	return s, nil
}

// load reads the YAML file into memory. Missing file is fine
// (returns an empty store).
func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	// Manual YAML decode to avoid a yaml.v3 dependency in this
	// package. The format is simple enough that a hand-rolled
	// parser is faster and dependency-free.
	entries := parseTrustYAML(string(data))
	for _, e := range entries {
		s.data[e.WorkspaceID] = e
	}
	return nil
}

// Save writes the current in-memory state to disk. Called whenever
// a new entry is added or an existing one is updated.
//
// MUST NOT be called while holding the write lock — it re-acquires
// the read lock to serialize the on-disk snapshot. The package
// methods that call Save (Grant, Revoke) drop the write lock
// before calling.
func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var b strings.Builder
	b.WriteString("# Trusted workspaces (Phase 16, Rec 5).\n")
	b.WriteString("# Entries are added when the user picks \"Always allow in this folder\"\n")
	b.WriteString("# in a consent dialog. Remove an entry to revoke trust.\n")
	for _, e := range s.data {
		fmt.Fprintf(&b, "- workspace_id: %q\n", e.WorkspaceID)
		fmt.Fprintf(&b, "  label: %q\n", e.Label)
		fmt.Fprintf(&b, "  always_allow: %t\n", e.AlwaysAllow)
		fmt.Fprintf(&b, "  created_at: %s\n", e.CreatedAt.UTC().Format(time.RFC3339))
		fmt.Fprintf(&b, "  last_used_at: %s\n", e.LastUsedAt.UTC().Format(time.RFC3339))
		if e.AppScope != DefaultScope {
			fmt.Fprintf(&b, "  app_scope: %q\n", string(e.AppScope))
		}
	}
	if err := os.WriteFile(s.path, []byte(b.String()), 0o600); err != nil {
		return fmt.Errorf("trust: write %s: %w", s.path, err)
	}
	return nil
}

// Lookup returns the trust entry for a (workspace, app) pair, or
// nil if no trust applies. Updates the entry's LastUsedAt on hit.
//
// The app filter lets a user say "only trust VS Code in this repo"
// without granting blanket trust to every terminal process.
func (s *Store) Lookup(workspaceID, app string) *Entry {
	if workspaceID == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.data[workspaceID]
	if !ok {
		return nil
	}
	if e.AppScope != DefaultScope && string(e.AppScope) != app {
		return nil
	}
	if !e.AlwaysAllow {
		return nil
	}
	e.LastUsedAt = time.Now()
	return e
}

// Grant adds or updates a workspace trust entry. Called when the
// user picks "Always allow in this folder" in the consent dialog.
//
// We hold the write lock long enough to mutate the map, then drop
// it before calling Save (which re-acquires the read lock).
func (s *Store) Grant(workspaceID, label string, appScope AppScope) (*Entry, error) {
	if workspaceID == "" {
		return nil, errors.New("trust: empty workspace ID")
	}
	now := time.Now()
	var entry *Entry
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		e, ok := s.data[workspaceID]
		if !ok {
			e = &Entry{
				WorkspaceID: workspaceID,
				Label:       label,
				AppScope:    appScope,
			}
			s.data[workspaceID] = e
		}
		e.AlwaysAllow = true
		e.CreatedAt = now
		e.LastUsedAt = now
		entry = e
	}()
	if err := s.Save(); err != nil {
		return nil, err
	}
	return entry, nil
}

// Revoke removes a workspace trust entry. Called when the user
// picks "Stop trusting this folder" in Settings, or when the
// workspace is deleted.
func (s *Store) Revoke(workspaceID string) error {
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if _, ok := s.data[workspaceID]; !ok {
			return // not found is not an error
		}
		delete(s.data, workspaceID)
	}()
	return s.Save()
}

// List returns a snapshot of all trust entries, sorted by
// LastUsedAt descending.
func (s *Store) List() []*Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Entry, 0, len(s.data))
	for _, e := range s.data {
		entryCopy := *e
		out = append(out, &entryCopy)
	}
	// Sort by LastUsedAt descending.
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].LastUsedAt.After(out[i].LastUsedAt) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// WorkspaceIDFor returns the canonical workspace ID for a path.
// Walks up from path looking for a `.git/` directory; if found,
// returns its absolute path. Otherwise returns the absolute path
// of the input (most conservative).
//
// The git-root heuristic matches what VS Code, JetBrains IDEs,
// and most editors use for "workspace" / "project" boundaries.
func WorkspaceIDFor(path string) string {
	if path == "" {
		return ""
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	cur := abs
	for {
		gitPath := filepath.Join(cur, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return cur
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			// Reached the filesystem root without finding .git.
			break
		}
		cur = parent
	}
	return abs
}

// parseTrustYAML is a hand-rolled parser for our trust store
// format. Keeps this package free of yaml.v3 (the daemon's
// config package already has it, but we don't want a chain
// dep through that).
//
// Recognized shape:
//
//   - workspace_id: "<path>"
//     label: "<name>"
//     always_allow: true
//     created_at: 2026-...
//     last_used_at: 2026-...
//     app_scope: "<bundle id>"   # optional
//
// Lines starting with `#` are comments. Unknown lines are
// ignored (forward compat: a newer entry type won't break an
// older binary).
func parseTrustYAML(content string) []*Entry {
	var out []*Entry
	var cur *Entry
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "- ") {
			if cur != nil {
				out = append(out, cur)
			}
			cur = &Entry{}
			line = strings.TrimPrefix(line, "- ")
			// Handle "- key: value" on a single line.
			if k, v, ok := splitKV(line); ok {
				applyKV(cur, k, v)
			}
			continue
		}
		if cur == nil {
			continue // pre-amble garbage; ignore
		}
		if k, v, ok := splitKV(line); ok {
			applyKV(cur, k, v)
		}
	}
	if cur != nil {
		out = append(out, cur)
	}
	return out
}

func splitKV(line string) (key, value string, ok bool) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:idx])
	value = strings.TrimSpace(line[idx+1:])
	// The writer uses Go's %q which escapes embedded " as \".
	// Strip outer quotes first, then unescape \" → ".
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
		value = strings.ReplaceAll(value, `\"`, `"`)
	}
	return key, value, true
}

func applyKV(e *Entry, key, value string) {
	switch key {
	case "workspace_id":
		e.WorkspaceID = value
	case "label":
		e.Label = value
	case "always_allow":
		e.AlwaysAllow = value == "true"
	case "created_at":
		if t, err := time.Parse(time.RFC3339, value); err == nil {
			e.CreatedAt = t
		}
	case "last_used_at":
		if t, err := time.Parse(time.RFC3339, value); err == nil {
			e.LastUsedAt = t
		}
	case "app_scope":
		e.AppScope = AppScope(value)
	}
}
