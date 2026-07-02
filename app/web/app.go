package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/conductor"
	"github.com/sahajpatel123/conduraapp/internal/daemon"
	"github.com/sahajpatel123/conduraapp/internal/hotkey"
	"github.com/sahajpatel123/conduraapp/internal/presence"
)

// App is the Wails app struct. Methods on this struct are bound to
// the frontend and callable from TypeScript via window.go.main.App.
type App struct {
	ctx context.Context

	// overlay tracks the current mode: false = main chat window,
	// true = compact overlay (frameless, always-on-top, transparent).
	overlay atomic.Bool

	// overlayCtrl is the Wails-backed overlay controller. Set
	// in NewApp() so main.go can register it with the daemon's
	// Subsystems before wails.Run() takes over. The Wails
	// runtime context is wired into the controller in startup().
	overlayCtrl *wailsController

	// conductor is the hotkey → overlay toggle. Started once the
	// daemon is ready. Nil until then.
	conductor *conductor.Conductor

	// killSwitch is the Layer-1 hard hotkey → halt.Flag.Halt wiring
	// (CLAUDE.md §5.3). Started alongside the overlay conductor;
	// stopped in beforeClose to release the OS-level registration.
	killSwitch *conductor.KillSwitchConductor

	// presenceOrch is shared with the conductor so menu/tray/JS can
	// open the quick prompt through the same summon/dismiss path.
	presenceOrch *presence.Orchestrator

	// quitCancel triggers a clean shutdown. Wired in main.go
	// to the same context the daemon goroutine runs under.
	// Called from the tray's Quit menu item and from
	// beforeClose when the user closes the window.
	quitCancel context.CancelFunc
}

// NewApp creates a new App application struct. The overlayCtrl
// is constructed here so main.go can pass it to daemon.Run via
// subs.SetOverlay before the Wails runtime context exists.
func NewApp() *App {
	return &App{
		overlayCtrl: newWailsController(),
	}
}

// startup is called when the Wails app starts. The context is
// saved so we can call Wails runtime methods, AND it's wired
// into the overlay controller so Show/Hide actually drive the
// Wails window.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if a.overlayCtrl != nil {
		a.overlayCtrl.SetContext(ctx)
	}
	diagLog("startup: Wails app started; overlay controller wired")
}

// domReady is called when the WebView has finished loading the
// frontend. We use it to signal that the GUI is ready for IPC.
func (a *App) domReady(ctx context.Context) {
	diagLog("domReady: WebView finished loading the frontend")
}

// beforeClose is called when the user closes the window. It stops
// the conductor (hotkey listener) so the goroutine doesn't leak.
func (a *App) beforeClose(ctx context.Context) bool {
	if a.conductor != nil {
		a.conductor.Stop()
		a.conductor = nil
	}
	if a.killSwitch != nil {
		a.killSwitch.Stop()
		a.killSwitch = nil
	}
	return false // allow close
}

// Ping is the simplest possible bound method: returns a greeting
// with a timestamp. Used to verify the TS↔Go bridge works.
func (a *App) Ping(name string) string {
	diagLog(fmt.Sprintf("Ping called with name=%q", name))
	return fmt.Sprintf("Hello %s, Condura is online (ts=%d).", name, time.Now().Unix())
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

// ShowOverlay opens the quick prompt. Kept for backward-compatible
// Wails bindings and tray wiring.
func (a *App) ShowOverlay() {
	a.OpenQuickPrompt()
}

// HideOverlay closes the quick prompt.
func (a *App) HideOverlay() {
	a.CloseQuickPrompt()
}

// IsOverlay reports whether the window is currently in overlay
// mode. Pure read; safe from any goroutine.
func (a *App) IsOverlay() bool {
	return a.overlay.Load()
}

// ToggleOverlay flips between main-window and overlay mode.
func (a *App) ToggleOverlay() {
	a.ToggleQuickPrompt()
}

// startConductor wires the hotkey → conductor → overlay chain once
// the daemon is ready. The conductor's onShow/onHide callbacks route
// through the Wails window methods so the overlay is a real
// frameless/always-on-top mode, not the daemon's noop controller.
//
// This is called from the daemon goroutine after daemonReady is closed.
func (a *App) startConductor(subs *daemon.Subsystems, hkSpec string) {
	if subs == nil {
		return
	}

	// Create the hotkey manager.
	hk := hotkey.New(hkSpec)

	// Create the presence orchestrator with the daemon's overlay
	// controller (noop). The conductor's onShow/onHide callbacks
	// will route through the Wails window methods instead.
	orch := presence.NewOrchestrator(subs.Overlay, subs.Halt, nil)

	a.presenceOrch = orch

	// Create the conductor with Wails-backed callbacks. The
	// orchestrator already resized the window; callbacks only sync
	// frontend state and focus.
	c, err := conductor.New(hk, orch,
		func() { a.syncQuickPromptOpen() },
		func() { a.syncQuickPromptClosed() },
	)
	if err != nil {
		slog.Warn("conductor init failed", "err", err)
		return
	}

	if err := c.Start(); err != nil {
		slog.Warn("conductor start failed", "err", err)
		return
	}

	a.conductor = c
	slog.Info("conductor ready", "hotkey", hkSpec)
}

// startKillSwitch wires the hard-hotkey → halt wiring (CLAUDE.md §5.3
// Layer 1). This is the documented user-facing kill switch: pressing
// the combo immediately halts every subsystem, persists the reason
// ("hard_hotkey") to the audit log, and the UI overlay surfaces the
// halt state.
//
// This is called from the daemon goroutine after daemonReady is
// closed. The spec comes from cfg.Hotkey.KillSwitch (or the product
// default Cmd+Shift+Escape if unset).
//
// We do NOT use the same hotkey.Manager as the overlay conductor —
// the kill switch MUST remain independent so a misbehaving overlay
// cannot block the user from halting the agent.
func (a *App) startKillSwitch(subs *daemon.Subsystems, ksSpec string) {
	if subs == nil || subs.Halt == nil {
		slog.Warn("kill switch not wired: no halt flag available")
		return
	}
	ks, err := conductor.NewKillSwitch(ksSpec, subs.Halt, nil)
	if err != nil {
		slog.Warn("kill switch init failed", "err", err)
		return
	}
	if err := ks.Start(); err != nil {
		slog.Warn("kill switch start failed", "err", err)
		return
	}
	a.killSwitch = ks
	slog.Info("kill switch ready", "hotkey", ksSpec)
}

// diagLog appends a line to ~/Library/Logs/condura-gui-diag.log so
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
	path := filepath.Join(logDir, "condura-gui-diag.log")
	line := fmt.Sprintf("%s %s\n", time.Now().Format(time.RFC3339), msg)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.WriteString(line)
}
