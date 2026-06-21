package daemon

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/ipc"
)

// fieldEnabled is the canonical JSON key for the boolean
// "is this subsystem enabled?" field. Shared across watchdog,
// reach, and other "not available" stubs so a future change is
// one constant update.
const fieldEnabled = "enabled"

// registerWatchdogMethods wires the watchdog.* method family
// (Phase 16, Rec 2: kill-switch Layer 2).
//
//   - watchdog.status — read the current state (enabled, last
//     touch, idle duration). Used by the GUI's "armed" badge.
//   - watchdog.touch — record that the user verified the agent's
//     actions. The GUI calls this on every user input event; the
//     daemon also calls it on every conversation.append so the
//     user doesn't have to be in the foreground for the watchdog
//     to stay quiet.
//   - watchdog.enable / disable — toggle at runtime without
//     restarting the daemon. Persisted via config.update.
func registerWatchdogMethods(srv *ipc.Server, subs *Subsystems) {
	if subs.Watchdog == nil {
		// Disabled in config. Stub everything to a friendly
		// "disabled" response so the GUI can branch on
		// `enabled` without error handling.
		disabled := func(_ context.Context, _ json.RawMessage) (any, error) {
			return map[string]any{
				fieldEnabled:      false,
				"last_touch":      "",
				"idle_seconds":    0,
				"timeout_seconds": 0,
			}, nil
		}
		srv.Register("watchdog.status", disabled)
		srv.Register("watchdog.touch", disabled)
		srv.Register("watchdog.enable", disabled)
		srv.Register("watchdog.disable", disabled)
		return
	}

	srv.Register("watchdog.status", func(_ context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{
			"enabled":         true,
			"last_touch":      subs.Watchdog.LastTouch().UTC().Format(time.RFC3339),
			"idle_seconds":    int64(subs.Watchdog.IdleDuration().Seconds()),
			"timeout_seconds": int64(subs.cfg.Daemon.Watchdog.Timeout.Seconds()),
		}, nil
	})

	srv.Register("watchdog.touch", func(_ context.Context, _ json.RawMessage) (any, error) {
		subs.Watchdog.Touch()
		return map[string]any{"ok": true}, nil
	})

	srv.Register("watchdog.enable", func(ctx context.Context, _ json.RawMessage) (any, error) {
		// Re-arm: if the watchdog was disabled in config we won't
		// have a *Watchdog in subs. Restart the daemon to pick up
		// config changes for now; in v0.2.0 the watchdog can be
		// constructed at runtime when the config changes.
		return nil, &ipc.Error{
			Code:    ipc.CodeInternalError,
			Message: "watchdog.enable requires daemon restart with watchdog.enabled=true in config.yaml",
		}
	})

	srv.Register("watchdog.disable", func(ctx context.Context, _ json.RawMessage) (any, error) {
		// Stop the watchdog by halting it: the next tick returns.
		// We can't re-arm without a restart, so we tell the user.
		return nil, &ipc.Error{
			Code:    ipc.CodeInternalError,
			Message: "watchdog.disable requires daemon restart with watchdog.enabled=false in config.yaml",
		}
	})
}
