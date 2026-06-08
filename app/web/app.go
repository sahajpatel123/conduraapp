package main

import (
	"context"
	"fmt"
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
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the Wails app starts. The context is
// saved so we can call Wails runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// domReady is called when the WebView has finished loading the
// frontend. We use it to signal that the GUI is ready for IPC.
func (a *App) domReady(ctx context.Context) {
	// Future: notify the daemon that the GUI is online.
}

// Ping is the simplest possible bound method: returns a greeting
// with a timestamp. Used to verify the TS↔Go bridge works.
func (a *App) Ping(name string) string {
	return fmt.Sprintf("Hello %s, Synaptic is online (ts=%d).", name, time.Now().Unix())
}

// DaemonStatus returns whether the in-process daemon is up and
// what its first listen address is. Returns an empty string for
// addr if the daemon is not yet ready.
func (a *App) DaemonStatus() DaemonStatusResult {
	select {
	case <-daemonReady:
	default:
		return DaemonStatusResult{Ready: false}
	}
	addr := ""
	if embeddedDaemon != nil {
		addr = embeddedDaemon.IPCAddr
	}
	return DaemonStatusResult{Ready: true, Addr: addr}
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
