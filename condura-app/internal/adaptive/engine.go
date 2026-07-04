package adaptive

import (
	"context"
	"log/slog"
	"time"
)

const sourceExplicit = "explicit"

// Engine ties the observer, dialectic, adjudicator, and predictor
// into a single working loop. It accepts observations, runs analysis,
// persists accepted proposals, and decays stale inferences.
//
// This is the "crown jewel" — the thing that actually makes the
// User-Adaptive Engine learn and adapt.
type Engine struct {
	Observer    *Observer
	Dialectic   *Dialectic
	Adjudicator *Adjudicator
	Store       Store
	Predictor   *Predictor
	cfg         Config
	log         *slog.Logger
	pending     []Proposal // confirmations waiting for user
}

// NewEngine creates the adaptive engine with all components wired.
func NewEngine(observer *Observer, dialectic *Dialectic, adjudicator *Adjudicator, store Store, predictor *Predictor, cfg Config, log *slog.Logger) *Engine {
	return &Engine{
		Observer:    observer,
		Dialectic:   dialectic,
		Adjudicator: adjudicator,
		Store:       store,
		Predictor:   predictor,
		cfg:         cfg,
		log:         log,
	}
}

// Run triggers the full analysis pipeline.
//
//nolint:gocritic // rangeValCopy is acceptable for small Observation structs
func (e *Engine) Run(ctx context.Context) {
	_ = ctx
	if !e.cfg.Enabled || e.cfg.Strength == StrengthOff {
		return
	}

	const defaultRecentDays = 7
	observations := e.Observer.Recent(defaultRecentDays)
	if len(observations) == 0 {
		return
	}

	// Filter: reject agent-suggested-then-accepted sessions.
	filtered := observations[:0]
	for _, o := range observations {
		if !e.Observer.WasSuggested(o.SessionID) {
			filtered = append(filtered, o)
		}
	}
	observations = filtered

	// Always run decay regardless of observations.
	e.decay(ctx)

	if len(observations) == 0 {
		return
	}

	proposals, err := e.Dialectic.Analyze(ctx, observations)
	if err != nil {
		if e.log != nil {
			e.log.Warn("adaptive: analyze failed", "err", err)
		}
	} else {
		e.applyProposals(proposals)
	}
	e.decay(ctx)
}

func (e *Engine) applyProposals(proposals []Proposal) {
	auto, confirm, _ := e.Adjudicator.Filter(proposals)

	// Save auto-apply proposals.
	model, _ := e.Store.Load()
	if model == nil {
		model = &UserModel{LastUpdated: time.Now(), Version: 1}
	}

	for _, p := range auto {
		e.applyToModel(model, p)
	}
	model.Version++
	_ = e.Store.Save(model)

	// Surface confirmations for the UI.
	e.pending = append(e.pending, confirm...)
}

func (e *Engine) applyToModel(model *UserModel, p Proposal) {
	field := InferredField{
		Value:      p.Value,
		Confidence: p.Confidence,
		Source:     "dialectic",
		LastSeen:   time.Now(),
	}
	switch p.Category {
	case "verbosity", "response_length":
		model.Style = field
	case "communication_style":
		model.Communication = field
	case "risk_tolerance":
		model.RiskTolerance = field
	default:
		model.Preferences = append(model.Preferences, field)
	}
}

// Pending returns proposals awaiting user confirmation.
func (e *Engine) Pending() []Proposal {
	return e.pending
}

// ConfirmPending accepts a pending proposal.
func (e *Engine) ConfirmPending(idx int) bool {
	if idx < 0 || idx >= len(e.pending) {
		return false
	}
	p := e.pending[idx]
	e.pending = append(e.pending[:idx], e.pending[idx+1:]...)

	model, _ := e.Store.Load()
	if model != nil {
		e.applyToModel(model, p)
		model.Version++
		_ = e.Store.Save(model)
	}
	return true
}

// RejectPending dismisses a pending proposal.
func (e *Engine) RejectPending(idx int) bool {
	if idx < 0 || idx >= len(e.pending) {
		return false
	}
	e.pending = append(e.pending[:idx], e.pending[idx+1:]...)
	return true
}

// SetStrength dynamically updates the engine's strength and propagates
// it to the dialectic (P2-8: live update).
func (e *Engine) SetStrength(s Strength) {
	e.cfg.Strength = s
	if e.Dialectic != nil {
		e.Dialectic.strength = s
	}
}

// decay removes stale inferences older than forget_after_days.
func (e *Engine) decay(_ context.Context) { //nolint:unparam
	if e.cfg.ForgetAfterDays <= 0 {
		return
	}
	model, err := e.Store.Load()
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-time.Duration(e.cfg.ForgetAfterDays) * 24 * time.Hour)
	changed := false

	// Decay individual fields.
	if model.Identity.LastSeen.Before(cutoff) && model.Identity.Source != sourceExplicit {
		model.Identity = InferredField{}
		changed = true
	}
	if model.Style.LastSeen.Before(cutoff) && model.Style.Source != sourceExplicit {
		model.Style = InferredField{}
		changed = true
	}
	if model.Communication.LastSeen.Before(cutoff) && model.Communication.Source != sourceExplicit {
		model.Communication = InferredField{}
		changed = true
	}
	if model.RiskTolerance.LastSeen.Before(cutoff) && model.RiskTolerance.Source != sourceExplicit {
		model.RiskTolerance = InferredField{}
		changed = true
	}

	// Decay list fields.
	model.Preferences = pruneList(model.Preferences, cutoff)
	model.Expertise = pruneList(model.Expertise, cutoff)
	model.PetPeeves = pruneList(model.PetPeeves, cutoff)
	model.TimePatterns = prunePatterns(model.TimePatterns, cutoff)
	model.Workflows = pruneWorkflows(model.Workflows, cutoff)

	if changed {
		_ = e.Store.Save(model)
	}
}

func pruneList(items []InferredField, cutoff time.Time) []InferredField {
	out := items[:0]
	for _, item := range items {
		if !item.LastSeen.Before(cutoff) || item.Source == "explicit" {
			out = append(out, item)
		}
	}
	return out
}

func prunePatterns(items []TimePattern, _ time.Time) []TimePattern { //nolint:unparam
	out := items[:0]
	for _, item := range items {
		if item.Count > 5 {
			out = append(out, item)
		}
	}
	return out
}

func pruneWorkflows(items []WorkflowPattern, _ time.Time) []WorkflowPattern { //nolint:unparam
	out := items[:0]
	for _, item := range items {
		if item.Count > 3 {
			out = append(out, item)
		}
	}
	return out
}
