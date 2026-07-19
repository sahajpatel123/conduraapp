package daemon

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/api_key"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/audit"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/config"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/diag"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/failover"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/halt"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/ipc"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/llm"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/sanitize"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/validate"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/version"
)

// daemonStart is the wall-clock time the daemon became ready to
// serve RPCs. It is read by the daemon.uptime / daemon.info methods.
//
// The atomic.Pointer lets Run() promote the value from the package-
// init sentinel (set in init() below) to the real "Run() is up and
// registerMethods has been called" time once subsystems are ready.
// Concurrent IPC readers see a non-nil time on every call.
var daemonStart atomic.Pointer[time.Time]

func init() {
	t := time.Now()
	daemonStart.Store(&t)
}

// MarkDaemonStart records that the daemon is now ready to serve
// RPCs. Run() calls this once after subsystem init but before
// registerMethods, so subsequent daemon.uptime / daemon.info
// responses report "time since ready" rather than "time since
// binary load". The atomic store is safe for concurrent reads.
func MarkDaemonStart(t time.Time) { daemonStart.Store(&t) }

// daemonUptimeSeconds returns wall-clock seconds since the daemon
// was last marked ready. Falls back to "since binary load" if
// MarkDaemonStart was never called (e.g., in unit tests that call
// registerMethods without going through Run()).
func daemonUptimeSeconds() float64 {
	t := daemonStart.Load()
	if t == nil {
		return 0
	}
	return time.Since(*t).Seconds()
}

// registerMethods wires every JSON-RPC method the daemon exposes into
// the given server. The method list is the single source of truth for
// what the GUI and CLI can call.
//
//nolint:gocognit,gocyclo // RPC method registration is intentionally a single inventory at startup; splitting it would scatter the daemon's RPC wire-format across files with no semantic benefit. Phase handlers are already split (methods_phase2.go, methods_phase9.go, …) where they cross the 30-line or 10-call threshold.
func registerMethods(srv *ipc.Server, log *slog.Logger, cfg *config.Config, subs *Subsystems, ver version.Info) {
	_ = log // kept for future per-method logging

	srv.Register("ping", func(_ context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{"pong": true, "ts": time.Now().Unix()}, nil
	})
	srv.Register("version", func(_ context.Context, _ json.RawMessage) (any, error) {
		return ver, nil
	})
	srv.Register("daemon.uptime", func(_ context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{"uptime_seconds": daemonUptimeSeconds()}, nil
	})
	srv.Register("daemon.pid", func(_ context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{"pid": os.Getpid()}, nil
	})
	srv.Register("daemon.info", func(_ context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{
			"uptime_seconds": daemonUptimeSeconds(),
			"pid":            os.Getpid(),
			"version":        ver,
		}, nil
	})
	srv.Register("config.get", func(_ context.Context, _ json.RawMessage) (any, error) {
		// Return a snake_case map matching YAML keys so the Meridian
		// GUI (AppConfig TS types) can read autonomy/hotkey/llm/…
		// encoding/json alone would emit PascalCase Go field names
		// because Config only carries yaml tags.
		return publicConfigView(cfg)
	})
	srv.Register("health.snapshot", func(ctx context.Context, _ json.RawMessage) (any, error) {
		return subs.Health.Snapshot(ctx), nil
	})
	// system.diag returns the local diagnostic snapshot (same
	// shape as `condura diag --json`). GUI parity with the CLI:
	// the "Support" panel in the GUI calls this to populate
	// the support-ticket form.
	srv.Register("system.diag", func(ctx context.Context, params json.RawMessage) (any, error) {
		return systemDiagHandler(ctx, subs, params)
	})
	// system.validate runs the local install health checks
	// (same shape as `condura validate`). GUI parity: the
	// "First-run diagnostics" panel runs this on startup and
	// shows the per-check pass/fail summary.
	srv.Register("system.validate", func(ctx context.Context, params json.RawMessage) (any, error) {
		return systemValidateHandler(ctx, subs, params)
	})
	// providers.list returns the full known catalog (for Settings) with
	// models + available=true when the provider is registered for live calls.
	srv.Register("providers.list", func(_ context.Context, _ json.RawMessage) (any, error) {
		registered := map[string]llm.Provider{}
		for _, p := range subs.LLM.List() {
			registered[p.Name()] = p
		}
		names := make([]string, 0, 16)
		seen := map[string]bool{}
		add := func(n string) {
			n = strings.TrimSpace(n)
			if n == "" || seen[n] {
				return
			}
			seen[n] = true
			names = append(names, n)
		}
		for _, n := range knownProviders() {
			add(n)
		}
		if cfg != nil && cfg.LLM.Providers != nil {
			for n := range cfg.LLM.Providers {
				add(n)
			}
		}
		sort.Strings(names)
		out := make([]map[string]any, 0, len(names))
		for _, name := range names {
			models := modelsForProvider(name)
			available := false
			if p, ok := registered[name]; ok {
				available = true
				if m := p.Models(); len(m) > 0 {
					models = m
				}
			}
			if models == nil {
				models = []llm.ModelInfo{}
			}
			out = append(out, map[string]any{
				"name":      name,
				"models":    models,
				"available": available,
			})
		}
		return out, nil
	})
	srv.Register("providers.models", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Provider string `json:"provider"`
		}
		if err := parseParams(params, &p); err != nil {
			return nil, err
		}
		if prov, ok := subs.LLM.Get(p.Provider); ok {
			models := prov.Models()
			if models == nil {
				models = []llm.ModelInfo{}
			}
			return models, nil
		}
		// Catalog fallback so Settings can browse models before a key is saved.
		models := modelsForProvider(p.Provider)
		if len(models) == 0 {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "unknown provider: " + p.Provider}
		}
		return models, nil
	})

	registerAPIKeyMethods(srv, subs)
	registerLLMMethods(srv, subs.LLM, subs.Spend, subs.Breakers, subs.Halt, subs.Audit)
	registerSpendMethods(srv, subs.Spend)
	registerConversationMethods(srv, subs.Conversations, subs.Audit, subs.Halt, subs.Streams, subs.LLM, subs.Anomaly, subs.Watchdog)
	registerAuditMethods(srv, subs.Audit)
	registerHaltMethods(srv, subs.Halt, subs.Audit, subs.Streams, subs.NetGuard, subs.ResumeTickets, subs.ResumeSecret, subs.Broker)
	registerControlMethods(srv, cfg, subs)
	registerFirstRunMethods(srv, subs.Audit)
	registerUpdateMethods(srv, subs.Updater, subs.Audit)
	registerWindowMethods(srv, subs)
	registerPhase6Methods(srv, subs)
	registerCUMethods(srv, subs)
	registerAdaptiveMethods(srv, subs)
	registerMCPMethods(srv, subs)
	registerSafetyMethods(srv, subs)
	registerGatekeeperMethods(srv, subs)
	registerDelegationMethods(srv, subs)
	registerPhase12Methods(srv, subs.Phase12)
	// Phase 11: trust & recovery.
	registerPhase11Methods(srv, subs)
	registerBackupMethods(srv, subs)
	registerUninstallMethods(srv, subs)
	registerPermissionMethods(srv, subs)
	registerOnboardingMethods(srv, subs)
	registerAccountMethods(srv, subs)
	registerReachMethods(srv, subs)
	registerWatchdogMethods(srv, subs)
	registerTrustMethods(srv, subs)
	registerCapabilitiesMethods(srv)
}

// registerAPIKeyMethods wires the apikeys.* method family.
func registerAPIKeyMethods(srv *ipc.Server, subs *Subsystems) {
	akm := subs.APIKeys
	srv.Register("apikeys.list", func(ctx context.Context, _ json.RawMessage) (any, error) {
		keys, err := akm.List(ctx)
		if err != nil {
			return nil, err
		}
		// Strip secrets before returning.
		type safeKey struct {
			ID       int64  `json:"id"`
			Provider string `json:"provider"`
			Label    string `json:"label"`
			AuthKind string `json:"auth_kind"`
			HasToken bool   `json:"has_token"`
		}
		out := make([]safeKey, 0, len(keys))
		for i := range keys {
			k := &keys[i]
			out = append(out, safeKey{
				ID:       k.ID,
				Provider: k.Provider,
				Label:    k.Label,
				AuthKind: string(k.AuthKind),
				HasToken: k.Secret != "",
			})
		}
		return out, nil
	})
	srv.Register("apikeys.set", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Provider string `json:"provider"`
			Label    string `json:"label"`
			Secret   string `json:"secret"`
		}
		if err := parseParams(params, &p); err != nil {
			return nil, err
		}
		// Phase 19 / audit 2026-06-28: apikeys.set is a WRITE-class
		// action (it persists a secret to disk). Per the default
		// policy in internal/gatekeeper/defaults.yaml every WRITE
		// requires gatekeeper consent. We route through
		// GatekeeperAllow so the action is denied when:
		//   - the daemon is halted (kill switch active)
		//   - the user is absent AND the policy requires presence
		//   - the engine is unavailable (fail-closed)
		// The previous bypass left the gatekeeper unenforced for
		// the only WRITE that stores a secret — a spec violation
		// vs MISSION.md §2.1 invariant #2 ("no model output flows
		// to a write without passing the Gatekeeper"). The gate
		// here is the deterministic rules engine, not a model,
		// so this does not change the GUIs existing happy-path
		// (the user explicitly typed the key).
		if !subs.GatekeeperAllow(ctx, "apikeys.set", "Store API key for "+p.Provider) {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "gatekeeper denied: " + p.Provider}
		}
		id, err := akm.Set(ctx, api_key.Key{
			Provider: p.Provider, Label: p.Label, AuthKind: api_key.AuthAPIKey, Secret: p.Secret,
		})
		if err != nil {
			return nil, err
		}
		// Phase 17, Fix #4 (B1): mark the provider as enabled in
		// the in-memory config and persist it so the choice
		// survives a daemon restart. Without this, the user adds a
		// key via the GUI but the daemon's LLM registry never
		// picks up the provider on the next boot because
		// cfg.LLM.Providers stays disabled. Rebind is best-effort;
		// we always call RebuildProviders so the current process
		// picks the new key up via buildProvidersFromConfig's
		// "auto-enable from api_keys" path.
		if subs != nil && subs.cfg != nil && subs.Loader != nil {
			entry, ok := subs.cfg.LLM.Providers[p.Provider]
			if !ok {
				entry = config.ProviderConfig{}
			}
			if !entry.Enabled {
				entry.Enabled = true
				if subs.cfg.LLM.Providers == nil {
					subs.cfg.LLM.Providers = map[string]config.ProviderConfig{}
				}
				subs.cfg.LLM.Providers[p.Provider] = entry
				if err := subs.Loader.Save(subs.cfg); err != nil {
					slog.Warn("persist provider enable failed", "provider", p.Provider, "err", err)
				}
			}
		}
		subs.RebuildProviders()
		return map[string]any{"id": id}, nil
	})
	srv.Register("apikeys.delete", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			ID int64 `json:"id"`
		}
		if err := parseParams(params, &p); err != nil {
			return nil, err
		}
		return nil, akm.Delete(ctx, p.ID)
	})
}

// registerLLMMethods wires the llm.* method family.
func registerLLMMethods(srv *ipc.Server, registry *llm.Registry, mon *failover.SpendMonitor, breakers *failover.BreakerRegistry, haltFlag *halt.Flag, auditLog *audit.Log) {
	srv.Register("llm.chat", func(ctx context.Context, params json.RawMessage) (any, error) {
		if haltFlag.IsHalted() {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "daemon is halted"}
		}
		var p struct {
			Provider string          `json:"provider"`
			Model    string          `json:"model"`
			Request  llm.ChatRequest `json:"request"`
		}
		if err := parseParams(params, &p); err != nil {
			return nil, err
		}
		prov, ok := registry.Get(p.Provider)
		if !ok {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "unknown provider: " + p.Provider}
		}
		if p.Request.Model == "" {
			p.Request.Model = p.Model
		}
		if p.Request.Model == "" {
			p.Request.Model = prov.DefaultModel("chat")
		}

		// Circuit breaker: fail fast if the provider is open.
		if breakers != nil {
			b := breakers.For(p.Provider)
			if !b.Allow() {
				return nil, &ipc.Error{
					Code:    ipc.CodeInternalError,
					Message: "circuit breaker open for provider: " + p.Provider,
				}
			}
		}

		// Spend cap: check before the call, not just after.
		if mon != nil {
			estCost := llm.EstimateCost(p.Request.Model, llm.Usage{})
			if !mon.Allow(estCost) {
				return nil, &ipc.Error{
					Code:    ipc.CodeInternalError,
					Message: "daily spend cap exceeded",
				}
			}
		}

		resp, err := prov.Chat(ctx, p.Request)
		if err != nil {
			if breakers != nil {
				breakers.For(p.Provider).RecordFailure()
			}
			return nil, err
		}
		if breakers != nil {
			breakers.For(p.Provider).RecordSuccess()
		}
		cost := llm.EstimateCost(p.Request.Model, resp.Usage)
		if mon != nil {
			mon.Record(cost)
		}
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorGUI, Action: "llm.chat", App: appConduraG,
			Level: auditLevelInfo, Result: auditResultAllow,
			// FIX B: model strings can come from user config;
			// redact defensively.
			Message: sanitize.RedactSecrets("provider=" + p.Provider + " model=" + p.Request.Model),
		})
		return map[string]any{
			"response": resp,
			"cost_usd": cost,
		}, nil
	})
}

// registerSpendMethods wires the spend.* method family.
func registerSpendMethods(srv *ipc.Server, mon *failover.SpendMonitor) {
	srv.Register("spend.today", func(_ context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{
			"spent":     mon.Spent(),
			"cap":       mon.Cap().USDPerDay,
			"remaining": mon.Remaining(),
		}, nil
	})
}

// publicConfigView returns cfg as a snake_case map suitable for the
// GUI. Uses YAML tags (Config's only tags) then strips secret fields
// so API keys never leave the daemon via config.get.
func publicConfigView(cfg *config.Config) (map[string]any, error) {
	if cfg == nil {
		return map[string]any{}, nil
	}
	// Deep-copy via YAML so we can redact without mutating runtime cfg.
	raw, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "encode config: " + err.Error()}
	}
	var out map[string]any
	if err := yaml.Unmarshal(raw, &out); err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "decode config view: " + err.Error()}
	}
	redactConfigSecrets(out)
	return out, nil
}

// redactConfigSecrets clears yaml-key secrets in the public config map.
func redactConfigSecrets(m map[string]any) {
	llmCfg, _ := m["llm"].(map[string]any)
	if llmCfg == nil {
		return
	}
	providers, _ := llmCfg["providers"].(map[string]any)
	for name, v := range providers {
		p, ok := v.(map[string]any)
		if !ok {
			continue
		}
		if s, _ := p["api_key"].(string); s != "" {
			p["api_key"] = ""
			p["has_api_key"] = true
		}
		providers[name] = p
	}
}

// systemDiagHandler is the package-private IPC handler for
// `system.diag`. Extracted from the inline closure so the
// test file can invoke it directly without spinning up an
// ipc.Server. The production registration calls this with
// the IPC params (currently nil — no input needed).
func systemDiagHandler(_ context.Context, subs *Subsystems, _ json.RawMessage) (any, error) {
	return diag.Take(subs.GeneralDataDir()), nil
}

// systemValidateHandler is the same pattern for `system.validate`.
func systemValidateHandler(ctx context.Context, subs *Subsystems, _ json.RawMessage) (any, error) {
	return validate.Run(ctx, subs.GeneralDataDir()), nil
}
