package adaptive

import "context"

// Predictor provides next-action suggestions based on the user model.
// Injected into session.buildMessages via a PredictorStore interface
// (following the same pattern as MemoryStore).
type Predictor struct {
	store Store
}

// PredictorStore is the interface injected into the session.
// Mirrors the MemoryStore pattern in session.go.
type PredictorStore interface {
	Predict(ctx context.Context, query string) (string, error)
}

// Prediction is a structured suggestion for the next action.
type Prediction struct {
	Suggestion string  `json:"suggestion"`
	Confidence float64 `json:"confidence"`
	Category   string  `json:"category"`
}

// NewPredictor creates a predictor backed by the user model store.
func NewPredictor(store Store) *Predictor {
	return &Predictor{store: store}
}

// Predict returns a context string to prepend to the LLM prompt.
// Implements PredictorStore.
func (p *Predictor) Predict(ctx context.Context, query string) (string, error) {
	model, err := p.store.Load()
	if err != nil {
		return "", err
	}
	_ = ctx
	_ = query
	_ = model

	// Predictions are applied only at balanced/aggressive strength.
	// The caller (session via Factory) gates this.
	return "", nil
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

// Reset clears all user model data and regenerates defaults.
func (v *Visibility) Reset(ctx context.Context) error {
	return v.store.Reset()
}
