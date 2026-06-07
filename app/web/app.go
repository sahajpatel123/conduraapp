package main

import (
	"context"
	"fmt"
	"time"
)

// App is the Wails app struct. Methods on this struct are bound to
// the frontend and callable from TypeScript via window.go.main.App.
type App struct {
	ctx context.Context
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
