// Package onboarding — EULA delivery.
//
// The EULA is shipped alongside the binary (EULA.md at the repo
// root, bundled in the app resources). This file reads it from
// disk on demand so that a text update does not require a binary
// rebuild.
package onboarding

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CurrentEULAVersion is the version string extracted from the
// EULA.md header. Must match "vN" (e.g. "v1") from the first
// H1 line. Bump this when the EULA changes.
var CurrentEULAVersion = "v1"

// EULADocument is the parsed EULA delivered to the GUI.
type EULADocument struct {
	Version   string `json:"version"`
	Text      string `json:"text"`
	UpdatedAt string `json:"updated_at"`
}

// ReadEULA reads EULA.md from the given data directory. The
// file is expected at <dataDir>/../EULA.md (sibling of the
// config directory in the app bundle). If dataDir is empty or
// the read fails, a bundled fallback message is returned so the
// wizard is never blocked.
func ReadEULA(dataDir string) (*EULADocument, error) {
	path := resolveEULAPath(dataDir)
	text, err := os.ReadFile(path) //nolint:gosec // trusted app-bundle path
	if err != nil {
		// Don't block the wizard. Return a minimal fallback
		// with a clear message that the full EULA was not
		// found.
		return &EULADocument{
			Version:   CurrentEULAVersion,
			Text:      "By using Synaptic, you agree to the Synaptic Freeware EULA v1. The full terms are available at synaptic.app/legal.",
			UpdatedAt: "2026-06-06",
		}, nil
	}
	updatedAt := extractUpdatedAt(string(text))
	return &EULADocument{
		Version:   CurrentEULAVersion,
		Text:      string(text),
		UpdatedAt: updatedAt,
	}, nil
}

// resolveEULAPath finds EULA.md relative to the data directory.
//
// In the bundled app, the layout is:
//
//	Synaptic.app/Contents/Resources/EULA.md
//	Synaptic.app/Contents/Resources/.synaptic/config.yaml
//
// So EULA.md is a sibling of the .synaptic directory, not inside it.
// We walk up from dataDir to find it.
func resolveEULAPath(dataDir string) string {
	if dataDir == "" {
		return "EULA.md" // fallback: CWD
	}
	// First try: sibling of dataDir (bundled app layout).
	sibling := filepath.Join(filepath.Dir(dataDir), "EULA.md")
	if _, err := os.Stat(sibling); err == nil {
		return sibling
	}
	// Second try: inside dataDir (development layout).
	inside := filepath.Join(dataDir, "EULA.md")
	if _, err := os.Stat(inside); err == nil {
		return inside
	}
	// Return sibling path even if missing — ReadEULA handles the
	// error gracefully.
	return sibling
}

// extractUpdatedAt parses "Last updated: YYYY-MM-DD" from the
// EULA markdown. Returns the raw date string or "".
func extractUpdatedAt(text string) string {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "**Last updated:**") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "**Last updated:**"))
		}
	}
	return ""
}

// ValidateEULAVersion returns true if the stored version matches
// the current version. An empty stored version (first-ever accept)
// is always valid.
func ValidateEULAVersion(stored string) bool {
	if stored == "" {
		return true
	}
	return stored == CurrentEULAVersion
}

// Internal helper used by tools that need to read the EULA from
// a known absolute path (primarily for the daemon RPC layer).
func readEULAFromPath(absPath string) (*EULADocument, error) {
	text, err := os.ReadFile(absPath) //nolint:gosec // caller provides trusted path
	if err != nil {
		return nil, fmt.Errorf("read EULA: %w", err)
	}
	return &EULADocument{
		Version:   CurrentEULAVersion,
		Text:      string(text),
		UpdatedAt: extractUpdatedAt(string(text)),
	}, nil
}
