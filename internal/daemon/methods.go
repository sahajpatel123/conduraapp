package daemon

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/api_key"
	"github.com/sahajpatel123/conduraapp/internal/audit"
	"github.com/sahajpatel123/conduraapp/internal/config"
	"github.com/sahajpatel123/conduraapp/internal/failover"
	"github.com/sahajpatel123/conduraapp/internal/halt"
	"github.com/sahajpatel123/conduraapp/internal/ipc"
	"github.com/sahajpatel123/conduraapp/internal/llm"
	"github.com/sahajpatel123/conduraapp/internal/version"
)

// registerMethods wires every JSON-RPC method the daemon exposes into
// the given server. The method list is the single source of truth for
// what the GUI and CLI can call.
func registerMethods(srv *ipc.Server, log *slog.Logger, cfg *config.Config, subs *Subsystems, ver version.Info) {
	_ = log // kept for future per-method logging

	srv.Register("ping", func(_ context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{"pong": true, "ts": time.Now().Unix()}, nil
	})
	srv.Register("version", func(_ context.Context, _ json.RawMessage) (any, error) {
		return ver, nil
	})
	srv.Register("config.get", func(_ context.Context, _ json.RawMessage) (any, error) {
		return cfg, nil
	})
	srv.Register("health.snapshot", func(ctx context.Context, _ json.RawMessage) (any, error) {
		return subs.Health.Snapshot(ctx), nil
	})
	srv.Register("providers.list", func(_ context.Context, _ json.RawMessage) (any, error) {
		list := subs.LLM.List()
		out := make([]map[string]string, 0, len(list))
		for _, p := range list {
			out = append(out, map[string]string{"name": p.Name()})
		}
		return out, nil
	})
	srv.Register("providers.models", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Provider string `json:"provider"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		prov, ok := subs.LLM.Get(p.Provider)
		if !ok {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "unknown provider: " + p.Provider}
		}
		return prov.Models(), nil
	})

	registerAPIKeyMethods(srv, subs)
	registerLLMMethods(srv, subs.LLM, subs.Spend, subs.Breakers, subs.Halt, subs.Audit)
	registerSpendMethods(srv, subs.Spend)
	registerConversationMethods(srv, subs.Conversations, subs.Audit, subs.Halt, subs.Streams, subs.LLM, subs.Anomaly, subs.Watchdog)
	registerAuditMethods(srv, subs.Audit)
	registerHaltMethods(srv, subs.Halt, subs.Audit, subs.Streams, subs.NetGuard, subs.ResumeTickets, subs.ResumeSecret)
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
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
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
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
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
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
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
		mon.Record(cost)
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorGUI, Action: "llm.chat", App: appConduraG,
			Level: auditLevelInfo, Result: auditResultAllow,
			Message: "provider=" + p.Provider + " model=" + p.Request.Model,
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
