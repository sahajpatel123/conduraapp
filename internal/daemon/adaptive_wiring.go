package daemon

import (
	"fmt"
	"log/slog"

	"github.com/sahajpatel123/synapticapp/internal/adaptive"
	"github.com/sahajpatel123/synapticapp/internal/failover"
	"github.com/sahajpatel123/synapticapp/internal/llm"
)

// AdaptiveComponents bundles the adaptive engine subsystems.
type AdaptiveComponents struct {
	Engine      *adaptive.Engine
	Observer    *adaptive.Observer
	Dialectic   *adaptive.Dialectic
	Predictor   *adaptive.Predictor
	Visibility  *adaptive.Visibility
	Store       *adaptive.EncryptedStore
	Strength    adaptive.Strength
	Adjudicator *adaptive.Adjudicator
	cfg         adaptive.Config
}

// buildAdaptiveEngine constructs the user-adaptive engine with all
// components wired end-to-end: Observer → Dialectic → Adjudicator →
// Store. The critic model and spend monitor are wired if available.
func buildAdaptiveEngine(store *adaptive.EncryptedStore, primary llm.Provider, critic llm.Provider, criticModel string, budget adaptive.BudgetChecker, log *slog.Logger) *AdaptiveComponents {
	cfg := adaptive.DefaultConfig()

	adj := adaptive.NewAdjudicator(cfg.AutoApply, cfg.RequireConfirm, cfg.DialecticMinConfidence)
	observer := adaptive.NewObserver()
	primaryModel := primary.DefaultModel("chat")

	// Wire critic model if available; fall back to proposer-only.
	cm := criticModel
	if cm == "" {
		if critic != nil {
			cm = critic.DefaultModel("chat")
		}
	}
	dialectic := adaptive.NewDialectic(primary, primaryModel, critic, cm, adj, budget, cfg.Strength)

	strength := func() adaptive.Strength { return cfg.Strength }
	predictor := adaptive.NewPredictor(store, strength)
	visibility := adaptive.NewVisibility(store)

	engine := adaptive.NewEngine(observer, dialectic, adj, store, predictor, cfg, log)

	return &AdaptiveComponents{
		Engine:      engine,
		Observer:    observer,
		Dialectic:   dialectic,
		Predictor:   predictor,
		Visibility:  visibility,
		Store:       store,
		Strength:    cfg.Strength,
		Adjudicator: adj,
		cfg:         cfg,
	}
}

// spendBudgetChecker adapts failover.SpendMonitor → adaptive.BudgetChecker.
type spendBudgetChecker struct {
	m *failover.SpendMonitor
}

func (c *spendBudgetChecker) CheckBudget() error {
	if c.m.Allow(0) {
		return nil
	}
	return fmt.Errorf("budget exceeded")
}
