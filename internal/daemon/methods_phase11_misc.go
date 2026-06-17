package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/ipc"
	"github.com/sahajpatel123/synapticapp/internal/onboarding"
	"github.com/sahajpatel123/synapticapp/internal/permissions"
	"github.com/sahajpatel123/synapticapp/internal/uninstall"
)

// registerUninstallMethods wires the uninstall.* RPC methods
// (Phase 11, sub-phase 11D). The methods are GATED through the
// Gatekeeper — uninstall is irreversible.
//
//   - uninstall.preview  — return the manifest of artifacts that
//     Uninstall would remove. Safe, no side effects.
//   - uninstall.execute  — actually remove the artifacts. Requires
//     a 32-hex ConfirmToken to prove the user typed the
//     "yes I really want to uninstall" phrase.
func registerUninstallMethods(srv *ipc.Server, subs *Subsystems) {
	srv.Register("uninstall.preview", func(_ context.Context, _ json.RawMessage) (any, error) {
		if subs.Uninstaller == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "uninstall subsystem not available"}
		}
		dataDir := subs.GeneralDataDir()
		preview, err := uninstall.Preview(uninstall.Options{DataDir: dataDir})
		if err != nil {
			return nil, fmt.Errorf("uninstall: preview: %w", err)
		}
		return preview, nil
	})

	srv.Register("uninstall.execute", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			ConfirmToken string `json:"confirm_token"`
			SkipBackup   bool   `json:"skip_backup"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if !subs.GatekeeperAllow(ctx, "uninstall.execute", "Uninstall Synaptic from this machine") {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: msgDeniedBySafetyPolicy}
		}
		var backupPath string
		if !p.SkipBackup && subs.Backup != nil {
			path, berr := subs.Backup.Create(ctx)
			if berr != nil {
				return nil, fmt.Errorf("uninstall: pre-backup failed: %w", berr)
			}
			backupPath = path
		}
		dataDir := subs.GeneralDataDir()
		result, err := uninstall.Uninstall(uninstall.Options{
			DataDir:      dataDir,
			ConfirmToken: p.ConfirmToken,
		})
		if err != nil {
			return nil, fmt.Errorf("uninstall: execute: %w", err)
		}
		if backupPath != "" {
			result.BackupPath = backupPath
		}
		_ = subs.Audit.Append(ctx, buildAuditEvent("uninstall.execute", appSynapticd, auditResultAllow, "files_removed="+fmt.Sprint(result.FilesRemoved)))
		return result, nil
	})
}

// registerPermissionMethods wires the permissions.* RPC
// methods (Phase 11, sub-phase 11E). All read-only — these
// surface the OS state to the GUI.
//
//   - permissions.status     — list every Kind with its current
//     grant Status.
//   - permissions.request_guide — return the per-platform
//     guide for granting a specific Kind.
func registerPermissionMethods(srv *ipc.Server, subs *Subsystems) {
	if subs.Permissions == nil {
		notAvailable := func(_ context.Context, _ json.RawMessage) (any, error) {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "permissions subsystem not available"}
		}
		srv.Register("permissions.status", notAvailable)
		srv.Register("permissions.request_guide", notAvailable)
		return
	}

	srv.Register("permissions.status", func(ctx context.Context, _ json.RawMessage) (any, error) {
		perms, err := permissions.Probe(ctx)
		if err != nil {
			return nil, fmt.Errorf("permissions: probe: %w", err)
		}
		return map[string]any{
			"platform": permissions.Platform(),
			"items":    perms,
		}, nil
	})

	srv.Register("permissions.request_guide", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Kind string `json:"kind"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		// Validate kind is a known value to prevent arbitrary
		// input from reaching the package-level API.
		known := false
		for _, k := range []permissions.Kind{
			permissions.KindAccessibility,
			permissions.KindScreenRecording,
			permissions.KindMicrophone,
			permissions.KindAutomation,
			permissions.KindNotifications,
		} {
			if strings.EqualFold(p.Kind, string(k)) {
				known = true
				p.Kind = string(k)
				break
			}
		}
		if !known {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "unknown permission kind: " + p.Kind}
		}
		guide := permissions.RequestGuide(permissions.Kind(p.Kind))
		return guide, nil
	})
}

// registerOnboardingMethods wires the onboarding.* RPC
// methods (Phase 11, sub-phase 11E). These let the GUI
// drive the wizard state machine.
//
//   - onboarding.state     — return the current State.
//   - onboarding.advance   — move to the next step.
//   - onboarding.back      — move to the previous step.
//   - onboarding.set_step  — record a step's status.
//   - onboarding.complete  — mark the wizard done.
//   - onboarding.reset     — start over.
//
//nolint:gocognit,gocyclo // combining all onboarding RPCs in one register function is intentional
func registerOnboardingMethods(srv *ipc.Server, subs *Subsystems) {
	if subs.Onboarding == nil {
		notAvailable := func(_ context.Context, _ json.RawMessage) (any, error) {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "onboarding subsystem not available"}
		}
		srv.Register("onboarding.state", notAvailable)
		srv.Register("onboarding.advance", notAvailable)
		srv.Register("onboarding.back", notAvailable)
		srv.Register("onboarding.set_step", notAvailable)
		srv.Register("onboarding.complete", notAvailable)
		srv.Register("onboarding.reset", notAvailable)
		return
	}

	srv.Register("onboarding.state", func(ctx context.Context, _ json.RawMessage) (any, error) {
		return subs.Onboarding.State(ctx)
	})

	srv.Register("onboarding.advance", func(ctx context.Context, _ json.RawMessage) (any, error) {
		return subs.Onboarding.Advance(ctx)
	})

	srv.Register("onboarding.back", func(ctx context.Context, _ json.RawMessage) (any, error) {
		return subs.Onboarding.Back(ctx)
	})

	srv.Register("onboarding.set_step", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Step   string `json:"step"`
			Status string `json:"status"`
			Data   string `json:"data"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		return subs.Onboarding.SetStepStatus(ctx, onboarding.Step(p.Step), onboarding.Status(p.Status), p.Data)
	})

	srv.Register("onboarding.complete", func(ctx context.Context, _ json.RawMessage) (any, error) {
		return subs.Onboarding.Complete(ctx)
	})

	srv.Register("onboarding.reset", func(ctx context.Context, _ json.RawMessage) (any, error) {
		return subs.Onboarding.Reset(ctx)
	})

	// Phase 14A — high-level onboarding wrappers for the converged
	// 4-step flow (eula → permissions → hotkey → complete).

	srv.Register("onboarding.eula", func(_ context.Context, _ json.RawMessage) (any, error) {
		dataDir := subs.GeneralDataDir()
		doc, err := onboarding.ReadEULA(dataDir)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return doc, nil
	})

	srv.Register("onboarding.probe_power", func(ctx context.Context, _ json.RawMessage) (any, error) {
		return onboarding.ProbePowerWithTimeout(ctx), nil
	})

	srv.Register("onboarding.skip", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Step string `json:"step"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if subs.Onboarding == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "onboarding subsystem not available"}
		}
		switch p.Step {
		case "permissions", "hotkey":
		default:
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "unknown step: " + p.Step}
		}
		s, err := subs.Onboarding.Skip(ctx, onboarding.Step(p.Step))
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		_ = subs.Audit.Append(ctx, buildAuditEvent("onboarding.skip", appSynapticd, auditResultAllow, "step="+p.Step))
		return s, nil
	})

	srv.Register("onboarding.finish", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Hotkey             string `json:"hotkey"`
			EULAVersion        string `json:"eula_version"`
			PermissionsSkipped bool   `json:"permissions_skipped"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if subs.Onboarding == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "onboarding subsystem not available"}
		}

		// 1. Validate EULA step is complete.
		s, err := subs.Onboarding.State(ctx)
		if err != nil {
			return nil, fmt.Errorf("onboarding: read state: %w", err)
		}
		if s.Steps[onboarding.StepEULA].Status != onboarding.StatusComplete {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "EULA must be accepted before finish"}
		}

		// 2. Validate hotkey is non-empty and parses.
		if p.Hotkey == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "hotkey is required"}
		}

		// 3. Probe power for the Ready screen response.
		power := onboarding.ProbePowerWithTimeout(ctx)

		// 4. If Ollama is reachable, enable it in config.
		// ProviderConfig doesn't carry its own name; the LLM
		// provider map is keyed by name. Look up "ollama" by key.
		if power.OllamaReachable && subs.cfg != nil {
			if prov, ok := subs.cfg.LLM.Providers["ollama"]; ok {
				prov.Enabled = true
				if power.FirstModel() != "" && prov.DefaultModel == "" {
					prov.DefaultModel = power.FirstModel()
				}
				subs.cfg.LLM.Providers["ollama"] = prov
			}
		}

		// 5. Apply hotkey to in-memory config.
		if subs.cfg != nil {
			subs.cfg.Hotkey.Overlay = p.Hotkey
		}

		// 6. Mark onboarding complete.
		if _, err := subs.Onboarding.Complete(ctx); err != nil {
			return nil, fmt.Errorf("onboarding: complete: %w", err)
		}

		// 7. Write first-run-complete marker.
		firstRunMarker := filepath.Join(subs.GeneralDataDir(), "first-run-complete")
		_ = os.WriteFile(firstRunMarker, []byte(time.Now().UTC().Format(time.RFC3339)), firstRunFilePerm) //nolint:gosec

		// 7.5. Rebuild providers so chats work immediately.
		if subs.cfg != nil {
			n := subs.RebuildProviders()
			slog.Info("onboarding.finish: providers rebuilt", "count", n)
		}

		// 8. Persist config to disk.
		if subs.Loader != nil && subs.cfg != nil {
			subs.cfg.General.FirstRun = false
			if err := subs.Loader.Save(subs.cfg); err != nil {
				slog.Error("onboarding.finish: config persist failed", "err", err)
			}
		}

		_ = subs.Audit.Append(ctx, buildAuditEvent("onboarding.finish", appSynapticd, auditResultAllow,
			fmt.Sprintf("hotkey=%s power=%s", p.Hotkey, power.Recommended)))

		return map[string]any{
			"power":        power,
			"hotkey":       p.Hotkey,
			"completed_at": time.Now().UTC().Format(time.RFC3339),
		}, nil
	})

	srv.Register("onboarding.is_complete", func(ctx context.Context, _ json.RawMessage) (any, error) {
		if subs.Onboarding == nil {
			return true, nil
		}
		done, err := subs.Onboarding.IsComplete(ctx)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return done, nil
	})

	// Phase 14H — voice readiness probe.
	srv.Register("onboarding.probe_voice", func(_ context.Context, _ json.RawMessage) (any, error) {
		return onboarding.ProbeVoice(), nil
	})
}
