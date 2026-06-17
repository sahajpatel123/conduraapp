package daemon

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/api_key"
	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/config"
	"github.com/sahajpatel123/synapticapp/internal/failover"
	"github.com/sahajpatel123/synapticapp/internal/halt"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
	"github.com/sahajpatel123/synapticapp/internal/llm"
	"github.com/sahajpatel123/synapticapp/internal/version"
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
		return subs.LLM.List(), nil
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

	registerAPIKeyMethods(srv, subs.APIKeys)
	registerLLMMethods(srv, subs.LLM, subs.Spend, subs.Halt, subs.Audit)
	registerSpendMethods(srv, subs.Spend)
	registerConversationMethods(srv, subs.Conversations, subs.Audit, subs.Halt, subs.Streams, subs.LLM)
	registerAuditMethods(srv, subs.Audit)
	registerHaltMethods(srv, subs.Halt, subs.Audit, subs.Streams)
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
}

// registerAPIKeyMethods wires the apikeys.* method family.
func registerAPIKeyMethods(srv *ipc.Server, akm *api_key.Manager) {
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
		id, err := akm.Set(ctx, api_key.Key{
			Provider: p.Provider, Label: p.Label, AuthKind: api_key.AuthAPIKey, Secret: p.Secret,
		})
		if err != nil {
			return nil, err
		}
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
func registerLLMMethods(srv *ipc.Server, registry *llm.Registry, mon *failover.SpendMonitor, haltFlag *halt.Flag, auditLog *audit.Log) {
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
		resp, err := prov.Chat(ctx, p.Request)
		if err != nil {
			return nil, err
		}
		cost := llm.EstimateCost(p.Request.Model, resp.Usage)
		mon.Record(cost)
		_ = auditLog.Append(ctx, audit.Event{
			Actor: actorGUI, Action: "llm.chat", App: appSynapticG,
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
