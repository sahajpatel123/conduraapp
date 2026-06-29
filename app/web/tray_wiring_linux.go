//go:build linux

package main

import (
	"context"
	"log/slog"

	"github.com/sahajpatel123/conduraapp/internal/daemon"
)

// startTray is a no-op on Linux. The tray package (internal/tray)
// is build-tagged !linux because the getlantern/systray dependency
// has no Linux support in this configuration. Linux users get an
// in-app menu instead via the existing Sidebar in the Wails GUI.
//
// This stub exists so app/web/main.go can call a.startTray
// unconditionally without a build-tag branch at the call site.
func (a *App) startTray(_ context.Context, _ *daemon.Subsystems) {
	slog.Info("tray: not available on Linux; using in-app menu instead")
}