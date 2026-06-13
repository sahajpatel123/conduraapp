package daemon

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/sahajpatel123/synapticapp/internal/delegation"
	"github.com/sahajpatel123/synapticapp/internal/failover"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
)

func buildDelegationBus(engine gatekeeper.Gatekeeper, sp *failover.SpendMonitor) *delegation.GatedRunner {
	cfg := delegation.DefaultConfig()
	limiter := delegation.NewLimiter(cfg, sp)
	runner := delegation.NewGatedRunner(cfg, engine, limiter)
	// MISSION S13.4: per-agent limit 4, global limit from config.
	runner.SetSemaphoreManager(delegation.NewSemaphoreManager(0, cfg.GlobalLimit))
	return runner
}

// Error codes for delegation RPC responses.
const (
	codeGatekeeperDeny = -32001
	codeCancelled      = -32002
)

func mapSpawnError(err error) *ipc.Error {
	switch {
	case errors.Is(err, delegation.ErrAgentNotFound), errors.Is(err, delegation.ErrRecursionLimit), errors.Is(err, delegation.ErrBudgetExceeded):
		return &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
	case errors.Is(err, delegation.ErrGatedDeny):
		return &ipc.Error{Code: codeGatekeeperDeny, Message: err.Error()}
	case errors.Is(err, context.Canceled):
		return &ipc.Error{Code: codeCancelled, Message: err.Error()}
	default:
		return &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
	}
}

// registerDelegationMethods registers delegation RPC methods.
func registerDelegationMethods(srv *ipc.Server, subs *Subsystems) {
	if subs.Delegation == nil {
		return
	}

	srv.Register("delegate.spawn", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			AgentName string  `json:"agent_name"`
			Task      string  `json:"task"`
			Model     string  `json:"model,omitempty"`
			Depth     int     `json:"depth"`
			Budget    float64 `json:"budget"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p.AgentName == "" || p.Task == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "agent_name and task are required"}
		}
		req := &delegation.SpawnRequest{
			AgentName: p.AgentName,
			Task:      p.Task,
			Model:     p.Model,
			Depth:     p.Depth,
			Budget:    p.Budget,
		}
		result, err := subs.Delegation.Spawn(ctx, req)
		if err != nil {
			return nil, mapSpawnError(err)
		}
		return result, nil
	})

	srv.Register("delegate.list_agents", func(_ context.Context, _ json.RawMessage) (any, error) {
		cfg := subs.Delegation.Config()
		agents := make([]map[string]any, len(cfg.Agents))
		for i := range cfg.Agents { //nolint:gocritic
			a := cfg.Agents[i]
			agents[i] = map[string]any{
				"name":        a.Name,
				"description": a.Description,
				"binary":      a.BinaryProbe,
			}
		}
		return map[string]any{"agents": agents}, nil
	})

	srv.Register("delegate.cancel", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			SpawnID string `json:"spawn_id"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p.SpawnID == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "spawn_id is required"}
		}
		if !subs.Delegation.Cancel(p.SpawnID) {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "unknown or already finished spawn_id"}
		}
		return auditOK(), nil
	})
}
