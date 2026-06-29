package daemon

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/anomaly"
	"github.com/sahajpatel123/synapticapp/internal/autonomy"
	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/config"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/halt"
	"github.com/sahajpatel123/synapticapp/internal/presence"
	"github.com/sahajpatel123/synapticapp/internal/sanitize"
	"github.com/sahajpatel123/synapticapp/internal/sensitive"
	"github.com/sahajpatel123/synapticapp/internal/sse"
	"github.com/sahajpatel123/synapticapp/internal/trust"
)

// SafetyComponents bundles the real safety layer.
type SafetyComponents struct {
	Engine    *gatekeeper.Engine
	Anomaly   *anomaly.Detector
	Consent   gatekeeper.ConsentProvider
	Sanitizer []sanitize.Sanitizer
	Trust     *trust.Store // Phase 16, Rec 5: per-workspace trust
	Autonomy  *autonomy.Matrix
	Presence  *presence.Detector // N1: user-presence detector; Stop on shutdown.
}

// buildSafetyLayer constructs the real safety components and wires all
// safety hooks (Anomaly, Autonomy, Sanitize, Trust) into the
// gatekeeper engine so that every action flows through the full
// safety pipeline.
//
// The autonomy hook is driven by the user-editable config matrix
// (config.AutonomyConfig.PerApp / PerTask / DefaultLevel). When the
// config has no entries, the hook falls back to the conservative
// hardcoded default (research / image_generation / code_review are
// autonomous; everything else warns) so a fresh install still works.
func buildSafetyLayer(haltFlag *halt.Flag, broker *sse.Broker, trustStore *trust.Store, cfg *config.Config, log *slog.Logger) *SafetyComponents {
	policy := gatekeeper.DefaultPolicy()
	var consent gatekeeper.ConsentProvider = &rpcConsentProvider{log: log, publish: func(nonce string, a any) {
		broker.PublishJSON("safety.consent.request", map[string]any{"nonce": nonce, "action": a})
	}}
	// CONDURA_TEST_AUTO_CONSENT — when set (test runs only), the
	// gatekeeper approves every consent ticket without going to the
	// GUI. This is wired so E2E tests that exercise the full
	// ipc.Server → registerMethods → GatekeeperAllow pipeline do
	// not block on a missing SSE consumer. The flag is gated to
	// a non-empty env var value so a production daemon (where the
	// env var is unset) is unaffected. The env-var guard is the
	// only thing standing between this and a real GUI flow.
	if os.Getenv("CONDURA_TEST_AUTO_CONSENT") != "" {
		log.Warn("CONDURA_TEST_AUTO_CONSENT is set — gatekeeper consent will be auto-approved; do not use in production")
		consent = &autoApproveConsentProvider{log: log}
	}
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
		detector.Record(a.Kind, 0, 0, false)
	}

	// Build the autonomy matrix from config. The matrix is the
	// user-defining setting from §27 — users dial per-cell via the
	// YAML config. When the config is empty, fall back to the
	// conservative hardcoded default so a fresh install still works.
	matrix := buildAutonomyMatrix(cfg)

	// Wire the autonomy hook so every Evaluate call checks autonomy.
	engine.AutonomyHook = func(taskType, app string) int {
		return int(matrix.Evaluate(taskType, app))
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

	// Phase 16, Rec 5: per-workspace trust hook. The trust store
	// is consulted by the gatekeeper before evaluating the WRITE
	// branch: a hit short-circuits to Allow with reason
	// "workspace trust: always-allow in this folder". The
	// store may be nil if trust is disabled or failed to load.
	if trustStore != nil {
		engine.TrustHook = func(workspaceID, app string) (any, bool) {
			entry := trustStore.Lookup(workspaceID, app)
			if entry == nil {
				return nil, false
			}
			return entry, true
		}
	}

	// N1: user-presence detector. Polls the OS for input-idle time
	// (macOS ioreg HIDIdleTime / Windows GetLastInputInfo / Linux
	// fail-closed) and feeds the gatekeeper's presence gate so
	// DESTRUCTIVE and require_user_active actions are held while the
	// user is absent. *presence.Detector satisfies
	// gatekeeper.PresenceChecker (IsPresent). Started here; the daemon
	// stops it on shutdown via SafetyComponents.Presence.
	presenceDetector := presence.NewDetector(5 * time.Second)
	presenceDetector.Start()
	engine.SetPresenceChecker(presenceDetector)

	return &SafetyComponents{
		Engine:    engine,
		Anomaly:   detector,
		Consent:   consent,
		Sanitizer: sanitize.DefaultChain(),
		Trust:     trustStore,
		Autonomy:  matrix,
		Presence:  presenceDetector,
	}
}

// buildAutonomyMatrix translates the user-editable config.AutonomyConfig
// into an autonomy.Matrix. PerTask entries become "task.*" wildcards;
// PerApp entries become "*.app" pairs (matched via the default path when
// no task wildcard hits); DefaultLevel sets the floor. When the config
// is empty, the conservative hardcoded default from §10.9 is used so a
// fresh install still behaves sensibly.
func buildAutonomyMatrix(cfg *config.Config) *autonomy.Matrix {
	if cfg == nil {
		return autonomy.NewMatrix(autonomy.Warn, defaultAutonomyMapping())
	}
	defaultLevel := parseAutonomyLevel(cfg.Autonomy.DefaultLevel, autonomy.Warn)
	mapping := defaultAutonomyMapping()
	// PerTask → task.* wildcards. These override the hardcoded defaults.
	for task, lvlStr := range cfg.Autonomy.PerTask {
		mapping[task+".*"] = parseAutonomyLevel(lvlStr, defaultLevel)
	}
	// PerApp → *.app pairs. We register them as "<any-task>.<app>"
	// by adding an entry for each known action kind, so the
	// Evaluate(taskType, app) lookup hits regardless of the task
	// type. The known action kinds are the set the engine actually
	// passes to the hook (a.Kind).
	for app, lvlStr := range cfg.Autonomy.PerApp {
		lvl := parseAutonomyLevel(lvlStr, defaultLevel)
		for _, kind := range autonomyActionKinds {
			mapping[kind+"."+app] = lvl
		}
	}
	return autonomy.NewMatrix(defaultLevel, mapping)
}

// defaultAutonomyMapping returns the conservative hardcoded default
// from §10.9: research / image_generation / code_review are autonomous;
// everything else warns. This is the floor when the user has not
// configured the matrix.
func defaultAutonomyMapping() map[string]autonomy.Level {
	return map[string]autonomy.Level{
		"research.*":         autonomy.Autonomous,
		"image_generation.*": autonomy.Autonomous,
		"code_review.*":      autonomy.Autonomous,
	}
}

// autonomyActionKinds is the set of action kinds the engine actually
// passes to the AutonomyHook (a.Kind). Used to expand PerApp entries
// into per-kind rows so the Evaluate(taskType, app) lookup hits
// regardless of the task type.
var autonomyActionKinds = []string{
	"chat",
	"shell.exec",
	"delegation.spawn",
	"computeruse.click",
	"computeruse.type",
	"computeruse.scroll",
	"computeruse.launch",
	"computeruse.read",
	"file.read",
	"file.write",
}

// parseAutonomyLevel parses a level string ("block", "warn", "ask",
// "autonomous") into an autonomy.Level, returning the fallback on
// empty or unrecognized input.
func parseAutonomyLevel(s string, fallback autonomy.Level) autonomy.Level {
	switch s {
	case "block", "0":
		return autonomy.Block
	case "warn", "1":
		return autonomy.Warn
	case "ask", "2":
		return autonomy.Ask
	case "autonomous", "3":
		return autonomy.Autonomous
	case "":
		return fallback
	default:
		return fallback
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

func (p *rpcConsentProvider) IsAvailable() bool { return true }

// autoApproveConsentProvider approves every ticket immediately.
// Used only by E2E tests behind the CONDURA_TEST_AUTO_CONSENT env
// guard (see buildSafetyLayer). Production code must never see
// this provider — the env-var check is the only thing protecting
// the production GUI flow.
type autoApproveConsentProvider struct {
	log *slog.Logger
}

func (p *autoApproveConsentProvider) Show(ctx context.Context, ticket *gatekeeper.ConsentTicket) (bool, error) {
	if ticket.Result == nil {
		return true, nil
	}
	select {
	case ticket.Result <- true:
		return true, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

func (p *autoApproveConsentProvider) IsAvailable() bool { return true }
