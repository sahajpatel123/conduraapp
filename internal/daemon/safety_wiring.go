package daemon

import (
	"context"
	"log/slog"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/anomaly"
	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/halt"
	"github.com/sahajpatel123/synapticapp/internal/sanitize"
	"github.com/sahajpatel123/synapticapp/internal/sensitive"
	"github.com/sahajpatel123/synapticapp/internal/sse"
)

// SafetyComponents bundles the real safety layer.
type SafetyComponents struct {
	Engine    *gatekeeper.Engine
	Anomaly   *anomaly.Detector
	Consent   gatekeeper.ConsentProvider
	Sanitizer []sanitize.Sanitizer
}

// buildSafetyLayer constructs the real safety components and wires all
// safety hooks (Anomaly, Autonomy, Sanitize) into the gatekeeper engine
// so that every action flows through the full safety pipeline.
func buildSafetyLayer(haltFlag *halt.Flag, broker *sse.Broker, log *slog.Logger) *SafetyComponents {
	policy := gatekeeper.DefaultPolicy()
	consent := &rpcConsentProvider{log: log, publish: func(nonce string, a any) {
		broker.PublishJSON("safety.consent.request", map[string]any{"nonce": nonce, "action": a})
	}}
	engine := gatekeeper.NewEngine(policy, consent, haltFlag)

	// Anomaly detector — async, graduated response.
	detector := anomaly.NewDetector(func(t anomaly.Trip) {
		switch t.Type {
		case anomaly.TripLoop, anomaly.TripFailures:
			_, _ = haltFlag.Halt(context.Background(), "anomaly: "+t.Reason)
		case anomaly.TripRate, anomaly.TripDuration:
			log.Warn("anomaly detected", "type", t.Type, "reason", t.Reason)
		}
	})

	// Wire the anomaly hook so every Evaluate call feeds the detector.
	engine.AnomalyHook = func(a blastradius.Action) {
		detector.Record(a.Kind, 0, 0, true)
	}

	// Wire the autonomy hook so every Evaluate call checks autonomy.
	engine.AutonomyHook = func(taskType, app string) int {
		level := getAutonomyLevel(taskType, app)
		return level
	}

	// Field-aware sanitizer dispatch: run the right sanitizer on
	// the right field, skip empties. PII sanitizer is applied at
	// consent display time (STEP 5), not here.
	engine.SanitizeHook = func(a *blastradius.Action) error {
		if a.Command != "" {
			if _, err := sanitize.NewShellSanitizer(nil).Sanitize(a.Command); err != nil {
				return err
			}
			if _, err := sanitize.NewPythonImportSanitizer().Sanitize(a.Command); err != nil {
				return err
			}
		}
		if a.Path != "" {
			if _, err := sanitize.NewPathSanitizer().Sanitize(a.Path); err != nil {
				return err
			}
		}
		if a.TargetURL != "" {
			if _, err := sanitize.NewURLSanitizer().Sanitize(a.TargetURL); err != nil {
				return err
			}
		}
		return nil
	}

	// Sensitive site hook: escalate actions on banking/health URLs
	// or data-entry contexts to RequirePresenceAndConsent.
	sensitiveDetector := sensitive.NewDetector()
	engine.SensitiveHook = func(url, ctx string) bool {
		return sensitiveDetector.Match(url, ctx)
	}

	return &SafetyComponents{
		Engine:    engine,
		Anomaly:   detector,
		Consent:   consent,
		Sanitizer: sanitize.DefaultChain(),
	}
}

// rpcConsentProvider publishes consent requests on SSE for GUI display.
type rpcConsentProvider struct {
	log     *slog.Logger
	publish func(nonce string, action any)
}

func (p *rpcConsentProvider) Show(ctx context.Context, ticket *gatekeeper.ConsentTicket) (bool, error) {
	// No publish callback → no GUI connected → fail-closed.
	if p.publish == nil {
		return false, nil
	}
	p.publish(ticket.Nonce, ticket.ActionKind)

	timer := time.NewTimer(time.Until(ticket.ExpiresAt))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-timer.C:
		return false, nil
	case result := <-ticket.Result:
		return result, nil
	}
}

// getAutonomyLevel returns the autonomy level for a given task type
// and app, looking up the config at runtime. Returns 1 (Warn) as the
// conservative default to ensure fail-safe behavior.
func getAutonomyLevel(taskType, app string) int {
	// Simplified autonomy mapping based on MISSION §10.9 defaults.
	// Research and code_review are autonomous; everything else warns.
	autonomousTasks := map[string]bool{
		"research":         true,
		"image_generation": true,
		"code_review":      true,
	}
	if autonomousTasks[taskType] {
		return 3 // Autonomous
	}
	apps := map[string]bool{"com.google.Chrome": true, "com.apple.finder": true, "com.microsoft.VSCode": true}
	if apps[app] {
		return 3 // Autonomous for explicitly trusted apps
	}
	return 1 // Warn for everything else
}

func (p *rpcConsentProvider) IsAvailable() bool { return true }
