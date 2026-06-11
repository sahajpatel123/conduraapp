package daemon

import (
	"github.com/sahajpatel123/synapticapp/internal/adaptive"
	"github.com/sahajpatel123/synapticapp/internal/llm"
)

// AdaptiveComponents bundles the adaptive engine subsystems.
type AdaptiveComponents struct {
	Observer    *adaptive.Observer
	Dialectic   *adaptive.Dialectic
	Predictor   *adaptive.Predictor
	Visibility  *adaptive.Visibility
	Store       *adaptive.EncryptedStore
	Strength    adaptive.Strength
	Adjudicator *adaptive.Adjudicator
}

// buildAdaptiveEngine constructs the user-adaptive engine if
// a storage.DB with encryption and an LLM provider are available.
func buildAdaptiveEngine(store *adaptive.EncryptedStore, primary llm.Provider, cfg adaptive.Config) *AdaptiveComponents {
	adj := adaptive.NewAdjudicator(cfg.AutoApply, cfg.RequireConfirm, cfg.DialecticMinConfidence)
	observer := adaptive.NewObserver()
	primaryModel := primary.DefaultModel("chat")
	dialectic := adaptive.NewDialectic(primary, primaryModel, nil, "", adj, nil, cfg.Strength)
	predictor := adaptive.NewPredictor(store)
	visibility := adaptive.NewVisibility(store)

	return &AdaptiveComponents{
		Observer:    observer,
		Dialectic:   dialectic,
		Predictor:   predictor,
		Visibility:  visibility,
		Store:       store,
		Strength:    cfg.Strength,
		Adjudicator: adj,
	}
}
