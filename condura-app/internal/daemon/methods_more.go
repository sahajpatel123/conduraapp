// Package daemon JSON-RPC method registration for config, telemetry,
// first-run, auto-update, and window/overlay/tray events. This is the
// "Phase 2 methods" surface that the GUI (Svelte/TS) calls.
package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/audit"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/config"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/failover"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/halt"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/ipc"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/overlay"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/updater"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/version"
)

// Permissions for the first-run marker file.
const (
	firstRunDirPerm  os.FileMode = 0o750
	firstRunFilePerm os.FileMode = 0o600
)

// registerControlMethods wires config.update + telemetry.setEnabled.
// config.update accepts partial patches for the sections the Meridian
// Settings desk actually edits: telemetry, hotkey, window, autonomy,
// security (spend/PII), llm provider toggles, and voice.wake.
//
//nolint:gocognit,gocyclo // the per-section patch handlers are flat srv.Register blocks with no nested branching; splitting them across files would scatter Meridian's edit surface and lose the audit table-of-contents this file provides.
func registerControlMethods(srv *ipc.Server, cfg *config.Config, subs *Subsystems) {
	srv.Register("config.update", func(ctx context.Context, params json.RawMessage) (any, error) {
		var patch map[string]json.RawMessage
		if err := json.Unmarshal(params, &patch); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		// Fail closed on unsupported sections — silent ignore looked like a
		// successful save (e.g. legacy delegation patches never persisted).
		supported := map[string]struct{}{
			"telemetry": {},
			"hotkey":    {},
			"window":    {},
			"autonomy":  {},
			"security":  {},
			"llm":       {},
			"voice":     {},
		}
		var unknown []string
		for k := range patch {
			if _, ok := supported[k]; !ok {
				unknown = append(unknown, k)
			}
		}
		if len(unknown) > 0 {
			sort.Strings(unknown)
			return nil, &ipc.Error{
				Code:    ipc.CodeInvalidParams,
				Message: "unsupported config.update keys: " + strings.Join(unknown, ","),
			}
		}
		keys := make([]string, 0, len(patch))
		if telRaw, ok := patch["telemetry"]; ok {
			applyTelemetryPatch(cfg, subs, telRaw)
			keys = append(keys, "telemetry")
		}
		if hkRaw, ok := patch["hotkey"]; ok {
			applyHotkeyPatch(cfg, hkRaw)
			keys = append(keys, "hotkey")
		}
		if wRaw, ok := patch["window"]; ok {
			applyWindowPatch(cfg, wRaw)
			keys = append(keys, "window")
		}
		if aRaw, ok := patch["autonomy"]; ok {
			if err := applyAutonomyPatch(cfg, aRaw); err != nil {
				return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
			}
			rewireAutonomy(subs, cfg)
			keys = append(keys, "autonomy")
		}
		if sRaw, ok := patch["security"]; ok {
			if err := applySecurityPatch(cfg, subs, sRaw); err != nil {
				return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
			}
			keys = append(keys, "security")
		}
		if lRaw, ok := patch["llm"]; ok {
			if err := applyLLMPatch(cfg, lRaw); err != nil {
				return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
			}
			// Live registry must match Settings toggles without restart.
			if subs != nil {
				subs.RebuildProviders()
			}
			keys = append(keys, "llm")
		}
		if vRaw, ok := patch["voice"]; ok {
			if err := applyVoicePatch(cfg, vRaw); err != nil {
				return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
			}
			keys = append(keys, "voice")
		}
		// Persist so Settings survive restart.
		if subs.Loader != nil {
			if err := subs.Loader.Save(cfg); err != nil {
				return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "persist config failed: " + err.Error()}
			}
		}
		if subs.Audit != nil {
			_ = subs.Audit.Append(ctx, audit.Event{
				Actor: actorGUI, Action: "config.update", App: appConduraG,
				Level: auditLevelInfo, Result: auditResultAllow,
				Message: "patched keys=" + strings.Join(keys, ","),
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

// validAutonomyLevel accepts the four config levels (+ legacy aliases).
func validAutonomyLevel(level string) bool {
	switch level {
	case config.AutonomySupervised, config.AutonomyWarn, config.AutonomyAutonomous, config.AutonomyBlock,
		"ask", syncAutoKey, "suggest":
		return true
	default:
		return false
	}
}

// normalizeAutonomyLevel maps legacy GUI aliases to canonical config values.
func normalizeAutonomyLevel(level string) string {
	switch level {
	case "ask":
		return config.AutonomySupervised
	case "auto":
		return config.AutonomyAutonomous
	case "suggest":
		return config.AutonomyWarn
	default:
		return level
	}
}

// applyAutonomyPatch updates cfg.Autonomy from Meridian Settings.
// Invalid levels fail the whole config.update (no silent drop).
func applyAutonomyPatch(cfg *config.Config, raw json.RawMessage) error {
	var a struct {
		DefaultLevel                          string            `json:"default_level"`
		PerApp                                map[string]string `json:"per_app"`
		PerTask                               map[string]string `json:"per_task"`
		ShowWarningsForRead                   *bool             `json:"show_warnings_for_read"`
		MaxConsecutiveWarnsBeforeAskingAnyway *int              `json:"max_consecutive_warns"`
	}
	if err := json.Unmarshal(raw, &a); err != nil {
		return fmt.Errorf("autonomy: %w", err)
	}
	if a.DefaultLevel != "" {
		if !validAutonomyLevel(a.DefaultLevel) {
			return fmt.Errorf("autonomy.default_level %q is invalid", a.DefaultLevel)
		}
		cfg.Autonomy.DefaultLevel = normalizeAutonomyLevel(a.DefaultLevel)
		// Keep daemon.default_autonomy aligned when present.
		cfg.Daemon.DefaultAutonomy = cfg.Autonomy.DefaultLevel
	}
	if a.PerTask != nil {
		next := make(map[string]string, len(a.PerTask))
		for task, lvl := range a.PerTask {
			if !validAutonomyLevel(lvl) {
				return fmt.Errorf("autonomy.per_task[%s] = %q is invalid", task, lvl)
			}
			next[task] = normalizeAutonomyLevel(lvl)
		}
		cfg.Autonomy.PerTask = next
	}
	if a.PerApp != nil {
		next := make(map[string]string, len(a.PerApp))
		for app, lvl := range a.PerApp {
			if !validAutonomyLevel(lvl) {
				return fmt.Errorf("autonomy.per_app[%s] = %q is invalid", app, lvl)
			}
			next[app] = normalizeAutonomyLevel(lvl)
		}
		cfg.Autonomy.PerApp = next
	}
	if a.ShowWarningsForRead != nil {
		cfg.Autonomy.ShowWarningsForRead = *a.ShowWarningsForRead
	}
	if a.MaxConsecutiveWarnsBeforeAskingAnyway != nil {
		cfg.Autonomy.MaxConsecutiveWarnsBeforeAskingAnyway = *a.MaxConsecutiveWarnsBeforeAskingAnyway
	}
	return nil
}

// applySecurityPatch updates spend cap / PII flags and rewires the live
// SpendMonitor so the cap applies to the next LLM call.
func applySecurityPatch(cfg *config.Config, subs *Subsystems, raw json.RawMessage) error {
	var s struct {
		SpendLimitUSDPerDay *float64 `json:"spend_limit_usd_per_day"`
		PIIRedaction        *bool    `json:"pii_redaction"`
		AuditRetentionDays  *int     `json:"audit_retention_days"`
	}
	if err := json.Unmarshal(raw, &s); err != nil {
		return fmt.Errorf("security: %w", err)
	}
	if s.SpendLimitUSDPerDay != nil {
		if *s.SpendLimitUSDPerDay < 0 {
			return fmt.Errorf("security.spend_limit_usd_per_day must be >= 0")
		}
		cfg.Security.SpendLimitUSDPerDay = *s.SpendLimitUSDPerDay
		if subs != nil && subs.Spend != nil {
			subs.Spend.SetCap(failover.SpendCap{USDPerDay: *s.SpendLimitUSDPerDay})
		}
	}
	if s.PIIRedaction != nil {
		cfg.Security.PIIRedaction = *s.PIIRedaction
	}
	if s.AuditRetentionDays != nil {
		if *s.AuditRetentionDays < 0 {
			return fmt.Errorf("security.audit_retention_days must be >= 0")
		}
		cfg.Security.AuditRetentionDays = *s.AuditRetentionDays
	}
	return nil
}

// applyLLMPatch merges provider toggles/default models without accepting
// api_key values (secrets stay on apikeys.set / keyring).
func applyLLMPatch(cfg *config.Config, raw json.RawMessage) error {
	var p struct {
		Providers map[string]struct {
			Enabled      *bool  `json:"enabled"`
			DefaultModel string `json:"default_model"`
			BaseURL      string `json:"base_url"`
		} `json:"providers"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("llm: %w", err)
	}
	if p.Providers == nil {
		return nil
	}
	if cfg.LLM.Providers == nil {
		cfg.LLM.Providers = map[string]config.ProviderConfig{}
	}
	for name, patch := range p.Providers {
		cur := cfg.LLM.Providers[name]
		if patch.Enabled != nil {
			cur.Enabled = *patch.Enabled
		}
		if patch.DefaultModel != "" {
			cur.DefaultModel = patch.DefaultModel
		}
		if patch.BaseURL != "" {
			cur.BaseURL = patch.BaseURL
		}
		// Never copy APIKey from a GUI patch.
		cfg.LLM.Providers[name] = cur
	}
	return nil
}

// applyVoicePatch updates voice.wake.enabled (and related flags) for Settings.
func applyVoicePatch(cfg *config.Config, raw json.RawMessage) error {
	var v struct {
		Enabled *bool `json:"enabled"`
		Wake    *struct {
			Enabled *bool `json:"enabled"`
		} `json:"wake"`
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		return err
	}
	if v.Enabled != nil {
		cfg.Voice.Enabled = *v.Enabled
	}
	if v.Wake != nil && v.Wake.Enabled != nil {
		cfg.Voice.Wake.Enabled = *v.Wake.Enabled
	}
	return nil
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

// registerCapabilitiesMethods wires daemon.capabilities. This is
// the GUI's "Trust & Safety" surface — it MUST reflect reality,
// not aspirations. The shape is consumed by app/web/frontend's
// SettingsPane and rendered as a read-only "What this build can
// and can't do" panel.
//
// See CLAUDE.md §2.1 invariant #4 (user can always stop the
// agent) and §33.5.2 row C4.14 (Layer 3 in-process limitation).
// Honest disclosure is the whole point: a kill switch that the
// agent process can disable is not a kill switch. Until the real
// pf/netsh companion ships in v0.2.0, Layer 3 in_process = true.
func registerCapabilitiesMethods(srv *ipc.Server) {
	srv.Register("daemon.capabilities", func(_ context.Context, _ json.RawMessage) (any, error) {
		killSwitch := map[string]any{
			// Layer 1 — hard hotkey. Wired into internal/hotkey +
			// kill switch overlay; works on macOS, Windows, Linux.
			"layer1_hotkey": true,
			// Layer 2 — in-process watchdog (CLAUDE.md §5.3,
			// §16 Rec 2). Auto-halt on inactivity. Defaults off
			// (user opts in via cfg.Daemon.Watchdog.Enabled); the
			// capability shape reports "available" rather than
			// "armed" because the user's choice is what matters.
			"layer2_watchdog": true,
			// Layer 3 — network isolation. Today the only
			// implementation is the in-process guard; the
			// `os_process` flag stays false until v0.2.0 swaps
			// to a real pf/netsh companion process the agent
			// cannot influence.
			"layer3_network_isolation": map[string]any{
				"in_process":  halt.IsLayer3InProcess(),
				"os_process":  false,
				"deferred_to": "v0.2.0",
				"reference":   "CLAUDE.md §33.5.2 row C4.14",
			},
		}
		computerUse := map[string]any{
			// ORAX Eye (AX-tree, MIT) is shipped as a stub because
			// the native bridge is not built for v0.1.0; the agent
			// falls back to the next non-stub backend.
			"orax":      "stub",
			"mac_cua":   "wrapper",
			"macos_mcp": "wrapper",
			// Vision CUA is intentionally disabled by default per
			// the v0.1.0 cost policy. The wrapper is present so a
			// user can opt in per-call via cfg.computer_use.
			"vision_cua": "disabled_default",
		}
		auditMap := map[string]any{
			// Redaction: secret-shaped strings in audit Message
			// fields are masked via internal/sanitize. Wired in
			// 2026-06-29.
			"redaction": true,
			// Prune tombstone: when audit pruning deletes a row,
			// a tombstone is kept so the HMAC chain's prev_hash
			// linkage stays verifiable across retention windows.
			"prune_tombstone": true,
			// HMAC subkey: the audit chain's HMAC key is derived
			// via HKDF from the master key (FIX 1, 2026-06-29) so
			// disclosure of the audit HMAC doesn't compromise the
			// master key.
			"hmac_subkey": true,
		}
		return map[string]any{
			"version":      version.Get(),
			"kill_switch":  killSwitch,
			"computer_use": computerUse,
			"audit":        auditMap,
		}, nil
	})
}

// auditOK returns a stable ack value for void RPC handlers. We return
// a small struct (not nil) so the wire payload is uniform and clients
// can detect success.
func auditOK() map[string]any {
	return map[string]any{"ok": true}
}
