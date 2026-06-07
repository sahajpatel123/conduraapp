package daemon

import (
	"context"

	"github.com/sahajpatel123/synapticapp/internal/failover"
	"github.com/sahajpatel123/synapticapp/internal/llm"
)

// buildFailoverProviders wraps every registered llm.Provider in a
// failover adapter. The order is the registration order; future
// versions will use cfg.Router to determine priority.
func buildFailoverProviders(registry *llm.Registry, br *failover.BreakerRegistry) []failover.Provider {
	all := registry.List()
	out := make([]failover.Provider, 0, len(all))
	for _, p := range all {
		defaultModel := p.DefaultModel("chat")
		out = append(out, failover.Provider{
			Name:    p.Name(),
			Breaker: br.For(p.Name()),
			Client:  &llmAdapter{prov: p, defaultModel: defaultModel},
			Models:  modelIDs(p.Models()),
		})
	}
	return out
}

// modelIDs returns the IDs from a slice of llm.ModelInfo.
func modelIDs(models []llm.ModelInfo) []string {
	out := make([]string, 0, len(models))
	for _, m := range models {
		out = append(out, m.ID)
	}
	return out
}

// llmAdapter wraps an llm.Provider to satisfy failover.LLMClient.
// The failover layer is provider-agnostic; it just calls Chat(provider, model).
//
// Note: the "ping" message is a placeholder — the real call (with the
// full ChatRequest) goes through llm.Registry directly, not the
// failover layer. The failover layer is currently used for cost-aware
// provider selection, not request routing.
type llmAdapter struct {
	prov         llm.Provider
	defaultModel string
}

// Chat sends a minimal ping to the provider to verify it's reachable
// and to record token usage + cost. The failover layer uses this for
// health-and-cost pre-checks; full request routing goes through the
// LLM registry directly.
func (a *llmAdapter) Chat(ctx context.Context, _, model string) (failover.Usage, error) {
	modelToUse := model
	if modelToUse == "" {
		modelToUse = a.defaultModel
	}
	resp, err := a.prov.Chat(ctx, llm.ChatRequest{
		Model:    modelToUse,
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "ping"}},
	})
	if err != nil {
		return failover.Usage{}, err
	}
	cost := llm.EstimateCost(modelToUse, resp.Usage)
	return failover.Usage{
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
		TotalTokens:  resp.Usage.TotalTokens,
		CostUSD:      cost,
	}, nil
}
