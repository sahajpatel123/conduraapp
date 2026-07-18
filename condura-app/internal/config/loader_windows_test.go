//go:build windows

package config

import (
	"path/filepath"
	"testing"
)

// TestFallbackDataDir_Windows_APPSet pins the Windows branch:
// %APPDATA%/condura is the per-user, per-app data location on
// Windows (the modern equivalent of $HOME/.config on Linux).
func TestFallbackDataDir_Windows_APPSet(t *testing.T) {
	t.Setenv("APPDATA", `C:\Users\test\AppData\Roaming`)
	got := fallbackDataDir()
	want := filepath.Join(`C:\Users\test\AppData\Roaming`, "condura")
	if got != want {
		t.Errorf("fallbackDataDir() (windows, APPDATA set) = %q, want %q", got, want)
	}
}

// TestFallbackDataDir_Windows_APPEmpty_HitsLastResort pins the
// Windows failure case: if APPDATA is unset (rare but possible
// on a misconfigured system), the function falls through to the
// relative ".condura" sentinel. The same data-leak trade-off as
// on darwin — better than nothing, but never the desired path.
func TestFallbackDataDir_Windows_APPEmpty_HitsLastResort(t *testing.T) {
	t.Setenv("APPDATA", "")
	got := fallbackDataDir()
	if got != ".condura" {
		t.Errorf("fallbackDataDir() (windows, APPDATA empty) = %q, want %q", got, ".condura")
	}
}
