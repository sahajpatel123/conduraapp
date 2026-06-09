package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails app struct. Methods on this struct are bound to
// the frontend and callable from TypeScript via window.go.main.App.
type App struct {
	ctx context.Context

	// overlay tracks the current mode: false = main chat window,
	// true = compact overlay (frameless, always-on-top, transparent).
	overlay atomic.Bool
}// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the Wails app starts. The context is
// saved so we can call Wails runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	diagLog("startup: Wails app started")
}

// domReady is called when the WebView has finished loading the
// frontend. We use it to signal that the GUI is ready for IPC.
func (a *App) domReady(ctx context.Context) {
	diagLog("domReady: WebView finished loading the frontend")
}

// Ping is the simplest possible bound method: returns a greeting
// with a timestamp. Used to verify the TS↔Go bridge works.
func (a *App) Ping(name string) string {
	diagLog(fmt.Sprintf("Ping called with name=%q", name))
	return fmt.Sprintf("Hello %s, Synaptic is online (ts=%d).", name, time.Now().Unix())
}

// DaemonStatus returns whether the in-process daemon is up and
// what its first listen address is. Returns an empty string for
// addr if the daemon is not yet ready.
func (a *App) DaemonStatus() DaemonStatusResult {
	diagLog("DaemonStatus called")
	select {
	case <-daemonReady:
	default:
		diagLog("DaemonStatus: daemon not yet ready")
		return DaemonStatusResult{Ready: false}
	}
	addr := ""
	if embeddedDaemon != nil {
		addr = embeddedDaemon.IPCAddr
	}
	diagLog(fmt.Sprintf("DaemonStatus: ready=true addr=%s", addr))
	return DaemonStatusResult{Ready: true, Addr: addr}
}

// LogFromFrontend receives a string from the JS side and appends
// it to the diagnostic file. The frontend calls this to surface
// any error or status it sees.
func (a *App) LogFromFrontend(msg string) {
	diagLog("frontend: " + msg)
}

// DaemonStatusResult is the JSON shape returned to the frontend.
type DaemonStatusResult struct {
	Ready bool   `json:"ready"`
	Addr  string `json:"addr"`
}

// ShowOverlay switches the main window into overlay mode: frameless,
// always-on-top, semi-transparent. The frontend (Svelte) is
// responsible for collapsing its own UI to a compact prompt bar.
//
// Safe to call from any goroutine.
func (a *App) ShowOverlay() {
	if a.ctx == nil {
		return
	}
	a.overlay.Store(true)
	wailsruntime.WindowSetAlwaysOnTop(a.ctx, true)
	// 0 = fully transparent. We want translucent, so the user can
	// still see the desktop behind. macOS treats A<255 as
	// compositing; we land on 230/255 = ~10% transparency.
	wailsruntime.WindowSetBackgroundColour(a.ctx, 18, 18, 22, 230)
}

// HideOverlay switches the main window back to normal mode: framed,
// not always-on-top, fully opaque.
func (a *App) HideOverlay() {
	if a.ctx == nil {
		return
	}
	a.overlay.Store(false)
	wailsruntime.WindowSetAlwaysOnTop(a.ctx, false)
	wailsruntime.WindowSetBackgroundColour(a.ctx, 18, 18, 22, 255)
}

// IsOverlay reports whether the window is currently in overlay
// mode. Pure read; safe from any goroutine.
func (a *App) IsOverlay() bool {
	return a.overlay.Load()
}

// ToggleOverlay flips between main-window and overlay mode.
func (a *App) ToggleOverlay() {
	if a.IsOverlay() {
		a.HideOverlay()
	} else {
		a.ShowOverlay()
	}
}

// diagLog appends a line to ~/Library/Logs/synaptic-gui-diag.log so
// we can diagnose startup problems from the Go side without seeing
// the GUI. Best-effort; errors are silently ignored so logging
// itself never breaks the app.
func diagLog(msg string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	logDir := filepath.Join(home, "Library", "Logs")
	_ = os.MkdirAll(logDir, 0o755)
	path := filepath.Join(logDir, "synaptic-gui-diag.log")
	line := fmt.Sprintf("%s %s\n", time.Now().Format(time.RFC3339), msg)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.WriteString(line)
}
