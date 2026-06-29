//go:build darwin

package main

import (
	"context"

	"github.com/sahajpatel123/conduraapp/internal/daemon"
)

// startTray is a no-op on macOS when building the Wails GUI.
//
// getlantern/systray defines an AppDelegate Objective-C class that
// conflicts with Wails v2's own AppDelegate, causing a duplicate
// symbol linker error. The tray works in the standalone daemon
// build (which doesn't use Wails). For the Wails GUI, the tray
// will be reimplemented using Wails v3's native systray API when
// the project upgrades.
//
// See: https://github.com/wailsapp/wails/issues/3003
// See: https://github.com/getlantern/systray/issues/261
func (a *App) startTray(_ context.Context, _ *daemon.Subsystems) {
	diagLog("startTray: skipped on macOS (Wails GUI; getlantern/systray conflicts)")
}
