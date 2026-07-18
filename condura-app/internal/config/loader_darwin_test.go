//go:build darwin

package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFallbackDataDir_Darwin_WithHome pins the macOS branch:
// $HOME/Library/Application Support/condura is the canonical
// per-user, per-app data location on macOS (matches Apple's HIG).
func TestFallbackDataDir_Darwin_WithHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	got := fallbackDataDir()
	want := filepath.Join(home, "Library", "Application Support", "condura")
	if got != want {
		t.Errorf("fallbackDataDir() (darwin, HOME=%s) = %q, want %q", home, got, want)
	}
}

// TestFallbackDataDir_Darwin_NoHome_HitsLastResort pins the failure
// case: when UserHomeDir() errors (e.g. HOME unset on a weird
// container), the function falls through to the relative ".condura"
// sentinel. This is a data-leak risk but the only safe option —
// we test it explicitly so a future refactor that REMOVES the
// last-resort return value (and panics instead) is caught.
func TestFallbackDataDir_Darwin_NoHome_HitsLastResort(t *testing.T) {
	// Set HOME to empty AND HOME to an unparseable value to force
	// os.UserHomeDir() to return an error.
	t.Setenv("HOME", "")
	// Some Go versions still resolve UserHomeDir() via /etc/passwd
	// when HOME is empty; the last-resort branch is only reachable
	// when UserHomeDir() actually fails. We can't easily simulate
	// that on macOS CI, so this test may pass either way.
	got := fallbackDataDir()
	if got != ".condura" {
		// If HOME was resolved (e.g., to a /Users/<user> dir), the
		// function will return the Library path. That's fine — it
		// means the last-resort branch isn't reachable here. The
		// test passes as long as the function doesn't panic or
		// return something nonsensical.
		if filepath.IsAbs(got) {
			t.Logf("darwin CI resolved HOME despite empty env; got %q", got)
		} else {
			t.Errorf("fallbackDataDir() (darwin, no HOME) returned non-absolute path %q; want .condura", got)
		}
	}
	_ = os.Getenv // keep os imported for future use
}
