// Package version holds build-time metadata for the Synaptic daemon and CLI.
//
// Variables are populated at build time via -ldflags (see Makefile):
//
//	-X 'github.com/sahajpatel123/synapticapp/internal/version.Version=v0.1.0'
//	-X 'github.com/sahajpatel123/synapticapp/internal/version.Commit=abc1234'
//	-X 'github.com/sahajpatel123/synapticapp/internal/version.BuildDate=2026-06-06T14:32:00Z'
//
// In dev builds (no ldflags), Version is set to "v0.0.0-dev" and Commit to "none".
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
)

// These are overridden at build time via -ldflags.
var (
	// Version is the semantic version (e.g., "v0.1.0" or "v0.1.0-3-gabc1234-dirty").
	Version = "v0.0.0-dev"
	// Commit is the full git commit SHA (40 hex chars) or "none".
	Commit = "none"
	// BuildDate is the RFC3339 build timestamp or "unknown".
	BuildDate = "unknown"
	// GoVersion is the Go toolchain version used to build the binary.
	GoVersion = runtime.Version()
	// Platform is "GOOS/GOARCH" of the build (e.g., "darwin/arm64").
	Platform = runtime.GOOS + "/" + runtime.GOARCH
)

var (
	infoOnce  sync.Once
	infoCache Info
)

// Info is a snapshot of build metadata, suitable for logging or JSON-RPC responses.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	ShortSHA  string `json:"short_sha"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
	IsDev     bool   `json:"is_dev"`
	// ModuleVersion is the version of the synaptic Go module as recorded in build info.
	ModuleVersion string `json:"module_version,omitempty"`
	// Dirty indicates an unclean working tree at build time (GoReleaser / local dev).
	Dirty bool `json:"dirty,omitempty"`
}

// Get returns the cached Info struct. Safe for concurrent use.
func Get() Info {
	infoOnce.Do(func() {
		infoCache = Info{
			Version:       Version,
			Commit:        Commit,
			ShortSHA:      shortSHA(Commit),
			BuildDate:     BuildDate,
			GoVersion:     GoVersion,
			Platform:      Platform,
			IsDev:         Version == "v0.0.0-dev" || Commit == "none",
			ModuleVersion: moduleVersion(),
			Dirty:         strings.HasSuffix(Version, "-dirty"),
		}
	})
	return infoCache
}

// String returns a one-line human-readable summary.
func String() string {
	i := Get()
	return fmt.Sprintf("Condura %s (%s, %s, %s)", i.Version, i.ShortSHA, i.GoVersion, i.Platform)
}

func shortSHA(commit string) string {
	if len(commit) >= 7 && commit != "none" {
		return commit[:7]
	}
	return commit
}

// moduleVersion returns the Go module version from debug.ReadBuildInfo, or "" if unknown.
func moduleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	// The main module's version is "" in dev; use the build settings if available.
	if info.Main.Version != "" {
		return info.Main.Version
	}
	// For installs via `go install @version`, Main.Version is populated.
	// For local dev, it's empty.
	return ""
}
