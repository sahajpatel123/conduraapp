package failover

import (
	"context"
	"errors"
	"fmt"
)

// LLMClient is the minimal LLM interface that Failover needs.
// internal/llm.Provider satisfies this.
type LLMClient interface {
	Chat(ctx context.Context, provider, model string) (Usage, error)
}

// Usage is the per-call cost record.
type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	CostUSD      float64
}

// Provider is the model of a provider for the purposes of failover.
type Provider struct {
	Name    string
	Breaker *CircuitBreaker
	Client  LLMClient
	Models  []string // candidate models in priority order
}

// Failover orchestrates a call across a prioritized list of providers.
type Failover struct {
	providers []Provider
	mon       *SpendMonitor
}

// New returns a Failover from the given providers and spend monitor.
func New(providers []Provider, mon *SpendMonitor) *Failover {
	return &Failover{providers: providers, mon: mon}
}

// EstimateCostFn is used to compute the cost of a candidate call before
// making it. If it returns an error, the call proceeds anyway (we just
// don't know the cost). If the estimated cost would exceed the spend cap,
// the candidate is skipped.
type EstimateCostFn func(provider, model string, inTokensApprox int) (float64, error)

// Chat runs the failover loop. It picks the first available candidate
// (breaker closed/half-open and spend cap OK), makes the call, and
// returns the usage + which provider/model succeeded.
//
// The breaker for the failed provider is updated on each error; the
// successful call resets its breaker.
func (f *Failover) Chat(ctx context.Context, inTokensApprox int, estimate EstimateCostFn) (Usage, Provider, Model, error) {
	var (
		zeroUsage Usage
		zeroProv  Provider
	)
	if len(f.providers) == 0 {
		return zeroUsage, zeroProv, "", errors.New("failover: no providers configured")
	}
	for _, p := range f.providers {
		if !p.Breaker.Allow() {
			continue
		}
		// Pick the first model that fits the spend cap.
		var chosenModel string
		for _, m := range p.Models {
			if f.mon == nil {
				chosenModel = m
				break
			}
			est := 0.0
			if estimate != nil {
				e, err := estimate(p.Name, m, inTokensApprox)
				if err == nil {
					est = e
				}
			}
			if f.mon.Allow(est) {
				chosenModel = m
				break
			}
		}
		if chosenModel == "" {
			continue
		}
		usage, err := p.Client.Chat(ctx, p.Name, chosenModel)
		if err != nil {
			p.Breaker.RecordFailure()
			continue
		}
		p.Breaker.RecordSuccess()
		if f.mon != nil {
			f.mon.Record(usage.CostUSD)
		}
		return usage, p, Model(chosenModel), nil
	}
	return zeroUsage, zeroProv, "", fmt.Errorf("%w: %d providers", ErrAllExhausted, len(f.providers))
}

// Model is a provider-specific model identifier.
type Model string
