package daemon

import (
	"context"
	"encoding/json"

	"github.com/sahajpatel123/synapticapp/internal/agent"
	"github.com/sahajpatel123/synapticapp/internal/computeruse"
	"github.com/sahajpatel123/synapticapp/internal/computeruse/backends"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
)

// cuComponents bundles the computer-use executor and loop
// for injection into Subsystems.
type cuComponents struct {
	gated *computeruse.GatedExecutor
	loop  *agent.CULoop
}

// buildCULoop constructs the computer-use execution pipeline:
// ORAX backend → Router → GatedExecutor → CUResolver → CULoop.
// Returns nil if no backends are available.
func buildCUComponents(gate gatekeeper.Gatekeeper, halt agent.HaltChecker) *cuComponents {
	orax := backends.NewORAX()
	oraxCU := computeruse.New(orax)
	oraxGate := computeruse.NewGatedExecutor(oraxCU, gate)

	resolver := NewCUResolver(oraxCU, oraxGate)
	verifier := agent.NewSimpleVerifier()
	planner := agent.NewSimplePlanner()

	loop := agent.NewCULoop(planner, verifier, resolver, halt)
	loop.StepTimeout = defaultActionTimeout

	return &cuComponents{
		gated: oraxGate,
		loop:  loop,
	}
}

// registerCUMethods registers computer-use RPC methods on the IPC server.
func registerCUMethods(srv *ipc.Server, subs *Subsystems) {
	if subs.CULoop == nil {
		return
	}

	srv.Register("cu.action", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Task string `json:"task"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p.Task == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "task is required"}
		}

		result, err := subs.CULoop.Run(ctx, p.Task, nil)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return map[string]any{
			"ok":       result.Success,
			"steps":    len(result.Steps),
			"goal":     result.Goal,
			"duration": result.Duration().Milliseconds(),
		}, nil
	})
}
