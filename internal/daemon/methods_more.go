// Package daemon JSON-RPC method registration for config, telemetry,
// first-run, auto-update, and window/overlay/tray events. This is the
// "Phase 2 methods" surface that the GUI (Svelte/TS) calls.
package daemon

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sahajpatel123/conduraapp/internal/audit"
	"github.com/sahajpatel123/conduraapp/internal/config"
	"github.com/sahajpatel123/conduraapp/internal/ipc"
	"github.com/sahajpatel123/conduraapp/internal/overlay"
	"github.com/sahajpatel123/conduraapp/internal/updater"
)

// Permissions for the first-run marker file.
const (
	firstRunDirPerm  os.FileMode = 0o750
	firstRunFilePerm os.FileMode = 0o600
)

// registerControlMethods wires config.update + telemetry.setEnabled.
// Phase 2: config.update accepts partial patches for the telemetry,
// hotkey, and window sections only.
func registerControlMethods(srv *ipc.Server, cfg *config.Config, subs *Subsystems) {
	srv.Register("config.update", func(ctx context.Context, params json.RawMessage) (any, error) {
		var patch map[string]json.RawMessage
		if err := json.Unmarshal(params, &patch); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if telRaw, ok := patch["telemetry"]; ok {
			applyTelemetryPatch(cfg, subs, telRaw)
		}
		if hkRaw, ok := patch["hotkey"]; ok {
			applyHotkeyPatch(cfg, hkRaw)
		}
		if wRaw, ok := patch["window"]; ok {
			applyWindowPatch(cfg, wRaw)
		}
		// Persist the patched config so changes survive a daemon
		// restart. Without this, hotkey/window/telemetry changes
		// are lost on the next boot.
		if subs.Loader != nil {
			if err := subs.Loader.Save(cfg); err != nil {
				return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "persist config failed: " + err.Error()}
			}
		}
		if subs.Audit != nil {
			_ = subs.Audit.Append(ctx, audit.Event{
				Actor: actorGUI, Action: "config.update", App: appConduraG,
				Level: auditLevelInfo, Result: auditResultAllow,
				Message: "patched keys",
			})
		}
		return auditOK(), nil
	})

	srv.Register("telemetry.status", func(_ context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{
			"enabled":  cfg.Telemetry.Enabled,
			"endpoint": cfg.Telemetry.Endpoint,
		}, nil
	})

	srv.Register("telemetry.setEnabled", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Enabled bool `json:"enabled"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		cfg.Telemetry.Enabled = p.Enabled
		subs.Telemetry.SetEnabled(p.Enabled)
		if subs.Loader != nil {
			if err := subs.Loader.Save(cfg); err != nil {
				return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "persist config failed: " + err.Error()}
			}
		}
		if subs.Audit != nil {
			_ = subs.Audit.Append(ctx, audit.Event{
				Actor: actorGUI, Action: "telemetry.setEnabled", App: appConduraG,
				Level: auditLevelInfo, Result: auditResultAllow,
				Message: "enabled=" + boolStr(p.Enabled),
			})
		}
		return auditOK(), nil
	})
}

// applyTelemetryPatch updates cfg.Telemetry + the running reporter
// from a JSON object {"enabled": bool, "endpoint": string}. Unset
// fields are left alone.
func applyTelemetryPatch(cfg *config.Config, subs *Subsystems, raw json.RawMessage) {
	var t struct {
		Enabled  bool   `json:"enabled"`
		Endpoint string `json:"endpoint"`
	}
	if err := json.Unmarshal(raw, &t); err != nil {
		return
	}
	cfg.Telemetry.Enabled = t.Enabled
	if t.Endpoint != "" {
		cfg.Telemetry.Endpoint = t.Endpoint
	}
	subs.Telemetry.SetEnabled(t.Enabled)
}

// applyHotkeyPatch updates cfg.Hotkey from a JSON object. Empty
// fields are ignored so the GUI can patch a single key.
func applyHotkeyPatch(cfg *config.Config, raw json.RawMessage) {
	var h struct {
		Overlay    string `json:"overlay"`
		KillSwitch string `json:"kill_switch"`
	}
	if err := json.Unmarshal(raw, &h); err != nil {
		return
	}
	if h.Overlay != "" {
		cfg.Hotkey.Overlay = h.Overlay
	}
	if h.KillSwitch != "" {
		cfg.Hotkey.KillSwitch = h.KillSwitch
	}
}

// applyWindowPatch updates cfg.Window from a JSON object. Non-zero
// fields win; zero fields are left alone.
func applyWindowPatch(cfg *config.Config, raw json.RawMessage) {
	var w struct {
		Width              int   `json:"width"`
		Height             int   `json:"height"`
		X                  *int  `json:"x"`
		Y                  *int  `json:"y"`
		LastConversationID int64 `json:"last_conversation_id"`
	}
	if err := json.Unmarshal(raw, &w); err != nil {
		return
	}
	if w.Width > 0 {
		cfg.Window.Width = w.Width
	}
	if w.Height > 0 {
		cfg.Window.Height = w.Height
	}
	if w.X != nil {
		cfg.Window.X = *w.X
	}
	if w.Y != nil {
		cfg.Window.Y = *w.Y
	}
	if w.LastConversationID != 0 {
		cfg.Window.LastConversationID = w.LastConversationID
	}
}

// registerFirstRunMethods wires firstRun.status + firstRun.complete.
func registerFirstRunMethods(srv *ipc.Server, auditLog *audit.Log) {
	srv.Register("firstRun.status", func(_ context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{
			"complete": firstRunMarkerExists(),
		}, nil
	})
	srv.Register("firstRun.complete", func(ctx context.Context, _ json.RawMessage) (any, error) {
		if err := writeFirstRunMarker(); err != nil {
			return nil, err
		}
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorGUI, Action: "firstRun.complete", App: appConduraG,
			Level: auditLevelInfo, Result: auditResultAllow,
		})
		return auditOK(), nil
	})
}

// registerUpdateMethods wires update.check + update.apply.
func registerUpdateMethods(srv *ipc.Server, u *updater.Updater, auditLog *audit.Log) {
	srv.Register("update.check", func(ctx context.Context, _ json.RawMessage) (any, error) {
		r, err := u.Check(ctx)
		if err != nil {
			return nil, err
		}
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorDaemon, Action: "update.check", App: appCondurad,
			Level: auditLevelInfo, Result: auditResultAllow,
			Message: "available=" + boolStr(r.UpdateAvailable) + " latest=" + r.LatestVersion,
		})
		return r, nil
	})
	srv.Register("update.apply", func(ctx context.Context, params json.RawMessage) (any, error) {
		var input struct {
			Result updater.Result `json:"result"`
		}
		if err := json.Unmarshal(params, &input); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeParseError, Message: err.Error()}
		}
		r, err := u.Apply(ctx, input.Result)
		if err != nil {
			return nil, err
		}
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorDaemon, Action: "update.apply", App: appCondurad,
			Level: auditLevelWarn, Result: auditResultAllow,
			Message: "version=" + r.LatestVersion,
		})
		return r, nil
	})
}

// registerWindowMethods wires window.show / window.hide / overlay.show
// / overlay.hide / tray.update. Phase 6 (6B #9): overlay.show and
// overlay.hide now route to the real overlay controller; tray.update
// routes to the tray status path (when a tray is wired in the GUI
// host). window.show and window.hide remain stubs (they're driven
// from the Wails GUI, not the daemon).
func registerWindowMethods(srv *ipc.Server, subs *Subsystems) {
	noOp := func(ctx context.Context, params json.RawMessage) (any, error) {
		_ = params
		auditEvent(ctx, subs, "window.event", "")
		return auditOK(), nil
	}
	srv.Register("window.show", noOp)
	srv.Register("window.hide", noOp)

	srv.Register("overlay.show", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			AtCursor bool `json:"at_cursor"`
			X        int  `json:"x"`
			Y        int  `json:"y"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if err := subs.Overlay.Show(ctx, overlay.ShowOpts{
			AtCursor: p.AtCursor,
			X:        p.X,
			Y:        p.Y,
		}); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		auditEvent(ctx, subs, "overlay.show", "")
		return auditOK(), nil
	})

	srv.Register("overlay.hide", func(ctx context.Context, _ json.RawMessage) (any, error) {
		if err := subs.Overlay.Hide(); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		auditEvent(ctx, subs, "overlay.hide", "")
		return auditOK(), nil
	})

	srv.Register("tray.update", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Status string `json:"status"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		auditEvent(ctx, subs, "tray.update", "status="+p.Status)
		subs.Broker.PublishJSON("tray.status", map[string]any{
			statusKey: p.Status,
		})
		return auditOK(), nil
	})

	srv.Register("window.state.get", func(_ context.Context, _ json.RawMessage) (any, error) {
		return subs.Window.Snapshot(), nil
	})
	srv.Register("window.state.setSize", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if err := subs.Window.SetSize(ctx, p.Width, p.Height); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return auditOK(), nil
	})
	srv.Register("window.state.setPosition", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			X *int `json:"x"`
			Y *int `json:"y"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if err := subs.Window.SetPosition(ctx, p.X, p.Y); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return auditOK(), nil
	})
	srv.Register("window.state.setLastConversation", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			ID int64 `json:"id"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if err := subs.Window.SetLastConversation(ctx, p.ID); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return auditOK(), nil
	})
}

// auditEvent logs an audit event if the audit log is available.
func auditEvent(ctx context.Context, subs *Subsystems, action, msg string) {
	if subs.Audit == nil {
		return
	}
	_ = subs.Audit.Append(ctx, audit.Event{
		Actor: actorGUI, Action: action, App: appConduraG,
		Level: auditLevelInfo, Result: auditResultAllow, Message: msg,
	})
}

// decodeParams unmarshals JSON params into v. Returns nil if params
// is empty. Returns an IPC invalid-params error on unmarshal failure.
func decodeParams(params json.RawMessage, v any) error {
	if len(params) == 0 {
		return nil
	}
	if err := json.Unmarshal(params, v); err != nil {
		return &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
	}
	return nil
}

// firstRunMarkerExists reports whether ~/.condura/first-run-complete
// exists.
func firstRunMarkerExists() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(home, ".condura", "first-run-complete"))
	return err == nil
}

// writeFirstRunMarker creates ~/.condura/first-run-complete. The
// marker is a plain 2-byte file ("ok"). Presence/absence is the only
// signal; we never write anything else.
func writeFirstRunMarker() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".condura")
	if err := os.MkdirAll(dir, firstRunDirPerm); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "first-run-complete"), []byte("ok"), firstRunFilePerm)
}

// boolStr converts a bool to "true"/"false" for audit messages.
func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// auditOK returns a stable ack value for void RPC handlers. We return
// a small struct (not nil) so the wire payload is uniform and clients
// can detect success.
func auditOK() map[string]any {
	return map[string]any{"ok": true}
}
