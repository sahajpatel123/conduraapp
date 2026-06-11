package daemon

import (
	"context"
	"encoding/json"

	"github.com/sahajpatel123/synapticapp/internal/adaptive"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
)

// registerAdaptiveMethods registers adaptive engine RPC methods.
func registerAdaptiveMethods(srv *ipc.Server, subs *Subsystems) {
	if subs.Adaptive == nil {
		return
	}

	// adaptive.profile: get the current user model.
	srv.Register("adaptive.profile", func(ctx context.Context, _ json.RawMessage) (any, error) {
		model, err := subs.Adaptive.Visibility.Profile(ctx)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return model, nil
	})

	// adaptive.forget: remove a specific inference.
	srv.Register("adaptive.forget", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Field string `json:"field"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if err := subs.Adaptive.Visibility.Forget(ctx, p.Field, p.Value); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return auditOK(), nil
	})

	// adaptive.reset: delete all inferences, start fresh.
	srv.Register("adaptive.reset", func(ctx context.Context, _ json.RawMessage) (any, error) {
		if err := subs.Adaptive.Visibility.Reset(ctx); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return auditOK(), nil
	})

	// adaptive.strength: get/set the engine strength.
	srv.Register("adaptive.strength.get", func(ctx context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{"strength": subs.Adaptive.Strength}, nil
	})
	srv.Register("adaptive.strength.set", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Strength string `json:"strength"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		switch adaptive.Strength(p.Strength) {
		case adaptive.StrengthOff, adaptive.StrengthCautious, adaptive.StrengthBalanced, adaptive.StrengthAggressive:
			subs.Adaptive.Strength = adaptive.Strength(p.Strength)
			if subs.Adaptive.Engine != nil {
				subs.Adaptive.Engine.SetStrength(adaptive.Strength(p.Strength))
			}
		default:
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "invalid strength: " + p.Strength}
		}
		return auditOK(), nil
	})
}
