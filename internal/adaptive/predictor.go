package adaptive

import (
	"context"
	"fmt"
	"strings"
)

// Predictor provides next-action suggestions based on the user model.
// Injected into session.buildMessages via a PredictorStore interface
// (following the same pattern as MemoryStore).
type Predictor struct {
	store    Store
	strength func() Strength
}

// PredictorStore is the interface injected into the session.
type PredictorStore interface {
	Predict(ctx context.Context, query string) (string, error)
}

// NewPredictor creates a predictor backed by the user model store.
// strength is a callback so the live setting is always read fresh.
func NewPredictor(store Store, strength func() Strength) *Predictor {
	return &Predictor{store: store, strength: strength}
}

// Predict returns context. Only active at balanced/aggressive strength.
//
//nolint:gocyclo
func (p *Predictor) Predict(ctx context.Context, query string) (string, error) {
	_ = ctx

	if p.strength == nil || p.strength() == StrengthOff || p.strength() == StrengthCautious {
		return "", nil
	}

	model, err := p.store.Load()
	if err != nil {
		return "", nil //nolint:nilerr // best-effort
	}

	const minConfidence = 0.5
	_ = minConfidence

	var parts []string

	// Communication style.
	if model.Communication.Value != "" && model.Communication.Confidence > 0.4 {
		parts = append(parts, fmt.Sprintf("User communication style: %s", model.Communication.Value))
	}

	// Identity.
	if model.Identity.Value != "" && model.Identity.Confidence > 0.4 {
		parts = append(parts, fmt.Sprintf("User identity: %s", model.Identity.Value))
	}

	// Risk tolerance.
	if model.RiskTolerance.Value != "" && model.RiskTolerance.Confidence > 0.4 {
		parts = append(parts, fmt.Sprintf("Risk tolerance: %s", model.RiskTolerance.Value))
	}

	// Style.
	if model.Style.Value != "" && model.Style.Confidence > 0.4 {
		parts = append(parts, model.Style.Value)
	}

	// Preferences (filter high-confidence).
	for _, pref := range model.Preferences {
		if pref.Confidence > minConfidence {
			parts = append(parts, fmt.Sprintf("Prefers: %s", pref.Value))
		}
	}

	// Expertise.
	for _, exp := range model.Expertise {
		if exp.Confidence > minConfidence {
			parts = append(parts, fmt.Sprintf("Expert in: %s", exp.Value))
		}
	}

	// Model prefs.
	if len(model.ModelPrefs) > 0 {
		var mps []string
		for k, v := range model.ModelPrefs {
			mps = append(mps, fmt.Sprintf("%s=%s", k, v))
		}
		parts = append(parts, fmt.Sprintf("Model preferences: %s", strings.Join(mps, ", ")))
	}

	if len(parts) == 0 {
		return "", nil
	}

	return "User profile:\n" + strings.Join(parts, "\n"), nil
}

// Prediction is a structured suggestion for the next action.
type Prediction struct {
	Suggestion string  `json:"suggestion"`
	Confidence float64 `json:"confidence"`
	Category   string  `json:"category"`
}

// Visibility provides the user-facing profile for the Settings UI.
type Visibility struct {
	store Store
}

// NewVisibility creates a visibility helper.
func NewVisibility(store Store) *Visibility {
	return &Visibility{store: store}
}

// Profile returns a displayable version of the user model with evidence.
func (v *Visibility) Profile(ctx context.Context) (*UserModel, error) {
	return v.store.Load()
}

// Forget removes a specific preference by field and value.
func (v *Visibility) Forget(ctx context.Context, field, value string) error {
	model, err := v.store.Load()
	if err != nil {
		return err
	}
	switch field {
	case "preferences":
		filtered := model.Preferences[:0]
		for _, p := range model.Preferences {
			if p.Value != value {
				filtered = append(filtered, p)
			}
		}
		model.Preferences = filtered
	case "expertise":
		filtered := model.Expertise[:0]
		for _, e := range model.Expertise {
			if e.Value != value {
				filtered = append(filtered, e)
			}
		}
		model.Expertise = filtered
	case "pet_peeves":
		filtered := model.PetPeeves[:0]
		for _, pp := range model.PetPeeves {
			if pp.Value != value {
				filtered = append(filtered, pp)
			}
		}
		model.PetPeeves = filtered
	}
	return v.store.Save(model)
}

// Reset clears all user model data.
func (v *Visibility) Reset(ctx context.Context) error {
	return v.store.Reset()
}
