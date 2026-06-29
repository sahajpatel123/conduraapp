package daemon

import (
	"context"
	"encoding/json"

	"github.com/sahajpatel123/conduraapp/internal/agent"
	"github.com/sahajpatel123/conduraapp/internal/computeruse"
	"github.com/sahajpatel123/conduraapp/internal/computeruse/backends"
	"github.com/sahajpatel123/conduraapp/internal/gatekeeper"
	"github.com/sahajpatel123/conduraapp/internal/ipc"
)

// cuComponents bundles the computer-use executor and loop.
type cuComponents struct {
	gated    *computeruse.GatedExecutor
	loop     *agent.CULoop
	resolver *CUResolver
}

// buildCUComponents constructs the 4-tier computer-use pipeline:
//  1. ORAX Eye  — structured AX tree, CGEvent
//  2. mac-cua   — background, CGEventPostToPid
//  3. macOS-MCP — comprehensive, AppleScript
//  4. Vision CUA — LLM-based, last resort (disabled by default)
//
// Uses LLMPlanner for task decomposition.
func buildCUComponents(gate gatekeeper.Gatekeeper, halt agent.HaltChecker, provider agent.PlannerProvider, model string) *cuComponents {
	if provider == nil {
		return nil
	}

	// Assemble the real 4-tier router.
	orax := backends.NewORAX()
	mc := backends.NewMacCUA()
	mcp := backends.NewMacOSMCP()
	vis := backends.NewVisionCUA(backends.VisionCUAConfig{Enabled: false}) // disabled by default

	backendList := []computeruse.Backend{}
	for _, b := range []computeruse.Backend{orax, mc, mcp, vis} {
		if b != nil {
			backendList = append(backendList, b)
		}
	}
	cu := computeruse.New(backendList...)
	cuGate := computeruse.NewGatedExecutor(cu, gate)

	resolver := NewCUResolver(cu, cuGate)
	verifier := agent.NewSimpleVerifier()
	planner := agent.NewLLMPlanner(provider, model)

	loop := agent.NewCULoop(planner, verifier, resolver, halt)
	loop.StepTimeout = defaultActionTimeout

	return &cuComponents{gated: cuGate, loop: loop, resolver: resolver}
}

// registerCUMethods registers computer-use RPC methods.
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
