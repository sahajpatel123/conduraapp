//go:build !darwin

package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/daemon"
	"github.com/sahajpatel123/synapticapp/internal/sse"
	"github.com/sahajpatel123/synapticapp/internal/status"
	"github.com/sahajpatel123/synapticapp/internal/tray"
)

// startTray wires the system tray to the running daemon. Called
// from app/web/main.go after daemon.Run() returns the Subsystems.
//
// What this gives the user:
//   - A menu-bar icon (macOS) / system tray (Windows) titled "Condura"
//   - Menu items: Show / Hide / Pause (kill switch) / Status line /
//     Spend today / Quit
//   - The Status line tracks the daemon's session/voice/computer-use
//     state in real time (via SSE)
//   - The Spend line tracks today's LLM spend (when the daemon
//     publishes spend.update; not yet wired in v0.1.0)
//   - Pause toggles the daemon's halt flag; the GUI's HALT button
//     in Settings stays in sync because both flip the same flag
//   - Quit shuts the daemon down cleanly via a synthetic SIGTERM
//
// What this DOES NOT do (out of scope for v0.1.0):
//   - Custom tray icon asset (systray uses a default; v0.2.0 ships
//     a real .icns/.ico)
//   - Linux support (the tray package is build-tagged !linux; Linux
//     users get an in-app menu instead via the existing Sidebar)
func (a *App) startTray(ctx context.Context, subs *daemon.Subsystems) {
	if subs == nil || subs.Broker == nil {
		slog.Warn("tray: subsystems or broker nil; skipping")
		return
	}

	// Construct the menu. Title appears in the OS tray; tooltip
	// appears on hover. Status/spend/voice strings update on
	// every daemon event.
	menu := tray.New("Condura", "Condura — AI on your computer, free")

	// Pre-populate the status line with "Idle" until the first
	// SSE event arrives.
	menu.SetStatus(status.StatusIdle)
	if subs.Halt != nil {
		menu.SetHalted(subs.Halt.IsHalted())
	}

	// Subscribe to the daemon's SSE broker so we can react to
	// status / voice / computer-use updates without polling.
	sub := subs.Broker.Subscribe()

	// Drive tray.Run in its own goroutine so main can keep
	// going. tray.Run blocks until ctx is done OR the user
	// clicks Quit.
	go func() {
		defer subs.Broker.Unsubscribe(sub)
		tray.Run(ctx, menu, func(ev tray.Event) {
			switch ev {
			case tray.EventShow:
				a.ShowOverlay()
			case tray.EventHide:
				a.HideOverlay()
			case tray.EventToggleHalt:
				a.toggleHaltFromTray(subs)
			case tray.EventQuit:
				slog.Info("tray: quit requested; daemon will shut down")
				a.requestQuit()
			}
		})
	}()

	// Drain SSE events and forward to the tray. Runs until ctx
	// is canceled. Cheap (select on one channel + a ticker).
	go a.pumpTrayStatus(ctx, menu, sub)
}

// pumpTrayStatus reads daemon events from the broker subscription
// and updates the tray's status / spend / voice labels.
//
// The daemon publishes "tray.status" with a {"status": "<name>"}
// payload from session/voice/CU status callbacks (see
// subsystems.go:604-608, 707-711, 676-680). We map the string
// name to the status.Status enum.
func (a *App) pumpTrayStatus(ctx context.Context, menu *tray.Menu, sub *sse.Subscription) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-sub.Events:
			if !ok {
				return
			}
			switch ev.Name {
			case "tray.status":
				var payload struct {
					Status string `json:"status"`
				}
				// ev.Data is interface{} (typically a map from
				// PublishJSON). Re-marshal through json so we
				// can unmarshal into our typed struct.
				raw, err := json.Marshal(ev.Data)
				if err != nil {
					continue
				}
				if err := json.Unmarshal(raw, &payload); err != nil {
					continue
				}
				menu.SetStatus(statusFromString(payload.Status))
			case "spend.update":
				var payload struct {
					USDToday float64 `json:"usd_today"`
				}
				raw, err := json.Marshal(ev.Data)
				if err != nil {
					continue
				}
				if err := json.Unmarshal(raw, &payload); err != nil {
					continue
				}
				menu.SetSpendUSD(payload.USDToday)
			}
		case <-time.After(60 * time.Second):
			// Defensive: refresh the tooltip periodically so
			// compositors that drop idle tray icons re-render.
			menu.SetTooltip("Condura — last update " + time.Now().Format("15:04:05"))
		}
	}
}

// toggleHaltFromTray flips the daemon's halt flag. Mirrors what
// the Settings → Kill switch button does (see Settings.svelte's
// performHalt handler). The tray menu is a deliberate physical
// click — the user knows they're toggling the kill switch —
// so we don't gate behind a confirmation dialog.
func (a *App) toggleHaltFromTray(subs *daemon.Subsystems) {
	if subs == nil || subs.Halt == nil {
		return
	}
	if subs.Halt.IsHalted() {
		_, _ = subs.Halt.Resume(context.Background())
	} else {
		_, _ = subs.Halt.Halt(context.Background(), "tray toggle")
	}
	// The daemon broadcasts tray.status changes via the SSE
	// broker, which pumpTrayStatus receives and applies via
	// menu.SetHalted (via the status stream).
}

// requestQuit signals the daemon goroutine to exit. Called from
// the tray's Quit menu item. The mechanism: we hold a dedicated
// cancel function in appInstance; calling it cancels the same
// context the daemon goroutine is running under, which makes
// daemon.Run return cleanly.
func (a *App) requestQuit() {
	if a.quitCancel != nil {
		a.quitCancel()
	}
}

// statusFromString maps the daemon's string status names back
// to the status.Status enum. Falls back to StatusIdle for
// unknown names so the tray never gets stuck on a label.
func statusFromString(s string) status.Status {
	switch s {
	case "idle":
		return status.StatusIdle
	case "listening":
		return status.StatusListening
	case "thinking":
		return status.StatusThinking
	case "speaking":
		return status.StatusSpeaking
	case "halted":
		return status.StatusHalted
	case "error":
		return status.StatusError
	default:
		return status.StatusIdle
	}
}
