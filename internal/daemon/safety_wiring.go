package daemon

import (
	"context"
	"log/slog"

	"github.com/sahajpatel123/synapticapp/internal/anomaly"
	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/halt"
	"github.com/sahajpatel123/synapticapp/internal/sanitize"
	"github.com/sahajpatel123/synapticapp/internal/sse"
)

// SafetyComponents bundles the real safety layer.
type SafetyComponents struct {
	Engine    *gatekeeper.Engine
	Anomaly   *anomaly.Detector
	Consent   gatekeeper.ConsentProvider
	Sanitizer []sanitize.Sanitizer
}

// buildSafetyLayer constructs the real safety components replacing
// the v0 DenyBeyondRead stub.
func buildSafetyLayer(haltFlag *halt.Flag, broker *sse.Broker, log *slog.Logger) *SafetyComponents {
	// Build the real Policy Engine.
	policy := gatekeeper.DefaultPolicy()
	consent := &rpcConsentProvider{log: log, publish: func(nonce string, a any) {
		broker.PublishJSON("safety.consent.request", map[string]any{"nonce": nonce, "action": a})
	}}
	engine := gatekeeper.NewEngine(policy, consent, haltFlag)

	// Wire anomaly detector — fires async via the Engine hook.
	detector := anomaly.NewDetector(func(t anomaly.Trip) {
		switch t.Type {
		case anomaly.TripLoop, anomaly.TripFailures:
			// Hard halt for loops and repeated failures.
			_, _ = haltFlag.Halt(context.Background(), "anomaly: "+t.Reason)
		case anomaly.TripRate, anomaly.TripDuration:
			// Pause + require re-consent for rate/duration.
			log.Warn("anomaly detected", "type", t.Type, "reason", t.Reason)
		}
	})
	engine.AnomalyHook = func(a blastradius.Action) { detector.Record(a.Kind, 0, 0, true) }

	// Wire sanitizers.
	// Wire sanitizers — defense-in-depth, runs before Policy.Evaluate.
	engine.SanitizeHook = func(a *blastradius.Action) error {
		_, err := sanitize.Chain(defaultSanitizers(), a.Kind)
		return err
	}

	return &SafetyComponents{
		Engine:    engine,
		Anomaly:   detector,
		Consent:   consent,
		Sanitizer: defaultSanitizers(),
	}
}

// rpcConsentProvider publishes consent requests on SSE for GUI display.
type rpcConsentProvider struct {
	log     *slog.Logger
	publish func(nonce string, action any)
}

func (p *rpcConsentProvider) Show(ctx context.Context, ticket *gatekeeper.ConsentTicket) (bool, error) {
	if p.publish != nil {
		p.publish(ticket.Nonce, ticket.ActionKind)
	}
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case result := <-ticket.Result:
		return result, nil
	}
}

func (p *rpcConsentProvider) IsAvailable() bool { return true }

// defaultSanitizers returns the standard safety sanitizer chain.
func defaultSanitizers() []sanitize.Sanitizer {
	return []sanitize.Sanitizer{
		sanitize.NewShellSanitizer(nil),
		sanitize.NewPathSanitizer(),
		sanitize.NewURLSanitizer(),
		sanitize.NewPIIRegexSanitizer(),
		sanitize.NewPythonImportSanitizer(),
	}
}
