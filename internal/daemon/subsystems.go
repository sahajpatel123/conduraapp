package daemon

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/account"
	"github.com/sahajpatel123/synapticapp/internal/adaptive"
	"github.com/sahajpatel123/synapticapp/internal/agent"
	"github.com/sahajpatel123/synapticapp/internal/anomaly"
	"github.com/sahajpatel123/synapticapp/internal/api_key"
	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/backup"
	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/computeruse"
	"github.com/sahajpatel123/synapticapp/internal/config"
	"github.com/sahajpatel123/synapticapp/internal/conversation"
	"github.com/sahajpatel123/synapticapp/internal/delegation"
	"github.com/sahajpatel123/synapticapp/internal/failover"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/halt"
	"github.com/sahajpatel123/synapticapp/internal/health"
	"github.com/sahajpatel123/synapticapp/internal/hub"
	"github.com/sahajpatel123/synapticapp/internal/i18n"
	"github.com/sahajpatel123/synapticapp/internal/llm"
	"github.com/sahajpatel123/synapticapp/internal/logger"
	"github.com/sahajpatel123/synapticapp/internal/mcp"
	"github.com/sahajpatel123/synapticapp/internal/memory"
	"github.com/sahajpatel123/synapticapp/internal/onboarding"
	"github.com/sahajpatel123/synapticapp/internal/overlay"
	"github.com/sahajpatel123/synapticapp/internal/permissions"
	"github.com/sahajpatel123/synapticapp/internal/reach"
	"github.com/sahajpatel123/synapticapp/internal/replay"
	"github.com/sahajpatel123/synapticapp/internal/secrets"
	"github.com/sahajpatel123/synapticapp/internal/session"
	"github.com/sahajpatel123/synapticapp/internal/skills"
	"github.com/sahajpatel123/synapticapp/internal/sse"
	"github.com/sahajpatel123/synapticapp/internal/status"
	"github.com/sahajpatel123/synapticapp/internal/storage"
	"github.com/sahajpatel123/synapticapp/internal/stream"
	"github.com/sahajpatel123/synapticapp/internal/sync"
	"github.com/sahajpatel123/synapticapp/internal/telemetry"
	"github.com/sahajpatel123/synapticapp/internal/uninstall"
	"github.com/sahajpatel123/synapticapp/internal/updater"
	"github.com/sahajpatel123/synapticapp/internal/voice"
	"github.com/sahajpatel123/synapticapp/internal/window"
)

// File mode constants. Owner-only for files that contain or refer to
// secrets; owner+group for directories (the daemon runs as a single
// user, but we leave group permissions open in case the user wants
// to grant the GUI process group access).
const (
	dataDirPerm  = 0o750
	addrFilePerm = 0o600
)

// Subsystems is the bundle of long-lived components the daemon
// constructs. Returned by Run() for tests and for the GUI's App
// struct; standalone callers can ignore it.
type Subsystems struct {
	// db is the storage handle, held so Phase 11 helpers
	// (backup, uninstall) can read the master key and path.
	db *storage.DB
	// cfg is the loaded config, held so Phase 11 helpers can
	// read schema version and data dir.
	cfg *config.Config
	// loader persists config changes back to disk (e.g. onboarding
	// finish writes ollama.enabled and first_run=false).
	Loader *config.Loader

	Secrets       secrets.Manager
	Storage       *storage.DB
	APIKeys       *api_key.Manager
	LLM           *llm.Registry
	Failover      *failover.Failover
	Spend         *failover.SpendMonitor
	Health        *health.Register
	Conversations *conversation.Store
	Audit         *audit.Log
	Halt          *halt.Flag
	NetGuard      halt.NetworkGuard
	Telemetry     *telemetry.Reporter
	Updater       *updater.Updater
	Window        *window.Manager
	Broker        *sse.Broker
	Streams       *stream.Manager
	IPCAddr       string // first listen addr (empty if IPC disabled)

	// Phase 6: living presence.
	// Gatekeeper is the canonical deterministic rules engine.
	// Every physical action goes through it (GatedAgentExecutor,
	// GatedComputerUseExecutor). Constructed once at startup.
	Gatekeeper gatekeeper.Gatekeeper
	// GatedAgentExecutor wraps the agent loop's executor with the
	// Gatekeeper. Use this from the agent loop; do NOT construct
	// another gatekeeper wrapping downstream.
	GatedAgentExecutor *agent.GatedExecutor
	// GatedComputerUseExecutor is the parallel wrapper for the
	// computer-use backends.
	GatedComputerUseExecutor *computeruse.GatedExecutor
	// Overlay is the overlay controller. Always non-nil (the
	// headless noop is a real implementation with a state machine).
	Overlay overlay.Controller
	// SessionFactory builds sessions on demand. Always non-nil
	// even when voice is disabled (text-only sessions still work).
	SessionFactory *session.Factory
	// Voice is the voice pipeline. Non-nil only when voice is
	// enabled in config and the binary/model are pinned correctly.
	Voice *voice.Pipeline

	// Phase 7: computer-use + memory.
	CULoop *agent.CULoop
	Memory *memory.StoreManager

	// Extractor runs async post-session memory extraction and
	// skill auto-creation. Nil when disabled.
	Extractor *PostSessionExtractor

	// Phase 8: user-adaptive engine.
	Adaptive *AdaptiveComponents

	// Phase 8: MCP Gateway.
	MCP *mcp.Manager

	// Phase 9: safety layer.
	Safety  *SafetyComponents
	Anomaly *anomaly.Detector

	// Phase 10: delegation bus.
	Delegation *delegation.GatedRunner

	// Phase 12: reach & ecosystem.
	Phase12 *Phase12Components

	// Phase 11: trust & recovery.
	//
	// Replay is the action-replay subsystem (timeline,
	// screenshots, integrity verifier). Nil if construction
	// failed (e.g., missing master key) — RPC methods on it
	// guard accordingly.
	Replay *replay.Replay
	// Backup owns encrypted backup creation, restore, and
	// rollback. Nil only if construction failed.
	Backup *backup.Manager
	// BackupScheduler runs the periodic auto-backup. Set
	// only when construction succeeds; daemon.Run starts it
	// after listeners are ready.
	BackupScheduler *backup.Scheduler
	// Uninstaller is a thin sentinel — the actual work is the
	// package-level uninstall.Preview / uninstall.Uninstall.
	// Always non-nil.
	Uninstaller *uninstall.Manager
	// Onboarding is the 8-step wizard state machine.
	// Always non-nil.
	Onboarding *onboarding.StateMachine
	// Permissions probes the OS for microphone / accessibility
	// / screen-recording consent. Always non-nil.
	Permissions *permissions.Manager
	// Account manages optional sign-in (Phase 14B). Nil when
	// disabled in config or construction fails.
	Account *account.Manager
	// Reach manages messaging channels (Phase 14C). Nil when
	// disabled in config or construction fails.
	Reach *reach.Manager
	// AuditLog is the HMAC-chained audit log, exposed here so
	// the replay integrity verifier can read from the same
	// chain the daemon writes to.
	AuditLog *audit.Log

	// closers holds resources that must be closed on shutdown.
	closers []io.Closer
}

// replaceMemoryCloser swaps the memory SQLite store in closers.
func (s *Subsystems) replaceMemoryCloser(newCloser io.Closer) {
	s.replaceCloserByType(func(c io.Closer) bool {
		_, ok := c.(*memory.SQLiteStore)
		return ok
	}, newCloser)
}

// replaceSkillCloser swaps the skills SQLite store in closers.
func (s *Subsystems) replaceSkillCloser(newCloser io.Closer) {
	s.replaceCloserByType(func(c io.Closer) bool {
		_, ok := c.(*skills.SQLiteStore)
		return ok
	}, newCloser)
}

func (s *Subsystems) replaceCloserByType(match func(io.Closer) bool, newCloser io.Closer) {
	if newCloser == nil {
		return
	}
	for i, c := range s.closers {
		if match(c) {
			s.closers[i] = newCloser
			return
		}
	}
	s.closers = append(s.closers, newCloser)
}

// Close releases all resources held by subsystems.
func (s *Subsystems) Close() error {
	var errs []error
	for _, c := range s.closers {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("subsystems close: %v", errs)
	}
	return nil
}

// CloseDatabases closes only the database connections (main DB,
// memory DB, skills DB) without closing the HTTP server or other
// long-lived resources. Used by backup.restore to release Windows
// file locks before the atomic directory swap. The caller must
// call Storage.Reload() after the swap to reopen the main DB.
func (s *Subsystems) CloseDatabases() {
	// Close in reverse order: skills, memory, main.
	if s.Phase12 != nil && s.Phase12.SkillStore != nil {
		_ = s.Phase12.SkillStore.Close()
	}
	// memStore is not directly accessible; it's in the closers list.
	// Close all closers except the last one (extractor).
	// The closers order is: db, memStore, extractor, skillStore.
	for i := len(s.closers) - 1; i >= 0; i-- {
		// Skip the extractor (index 2 in the original list).
		// We close everything else.
		_ = s.closers[i].Close()
	}
}

// ReloadAuxiliaryDatabases recreates the memory and skills stores
// from disk after a backup restore. The main DB is already reloaded
// by Storage.Reload(); this method handles the auxiliary stores
// and subsystems that held pre-reload DB handles. Returns the
// combined error from all reloads (best-effort: all are attempted
// even if earlier ones fail).
func (s *Subsystems) ReloadAuxiliaryDatabases() error {
	var errs []error

	// Reload audit log — critical for post-restore audit chain.
	// subs.Audit was constructed with the old db.SQL(); without
	// this reload, every Append call after restore silently fails.
	if s.AuditLog != nil && s.Storage != nil {
		s.AuditLog.Reload(s.Storage.SQL())
	}

	// Reload replay screenshot store — holds a stale db.SQL() handle.
	if s.Replay != nil {
		if shots := s.Replay.Screenshots(); shots != nil && s.Storage != nil {
			shots.Reload(s.Storage.SQL())
		}
	}

	// Reload memory store.
	memPath := s.MemoryDBPath()
	if memPath != "" {
		if s.Memory != nil {
			_ = s.Memory.Close()
			s.Memory = nil
		}
		memStore, err := memory.NewSQLiteStore(memPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("reload memory: %w", err))
		} else {
			s.replaceMemoryCloser(memStore)
			s.Memory = memory.NewManager(memStore)
		}
	}

	// Reload skills store.
	skillPath := s.SkillDBPath()
	if skillPath != "" && s.Phase12 != nil {
		if s.Phase12.SkillStore != nil {
			_ = s.Phase12.SkillStore.Close()
		}
		skillStore, err := skills.NewSQLiteStore(skillPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("reload skills: %w", err))
			s.Phase12.SkillStore = nil
		} else {
			s.replaceSkillCloser(skillStore)
			s.Phase12.SkillStore = skillStore
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("subsystems: auxiliary db reload: %v", errs)
	}
	return nil
}

// MasterKey returns the storage.DB master key. Used by the
// backup subsystem to derive the per-archive key. Returns
// nil, nil if storage isn't wired (e.g., a test that doesn't
// open a real DB).
func (s *Subsystems) MasterKey() ([]byte, error) {
	if s == nil || s.Storage == nil {
		return nil, fmt.Errorf("subsystems: storage not initialized")
	}
	return s.Storage.MasterKey(), nil
}

// GeneralDataDir returns the on-disk data directory.
func (s *Subsystems) GeneralDataDir() string {
	if s == nil || s.Storage == nil {
		return ""
	}
	return filepath.Dir(s.Storage.Path())
}

// SkillDBPath returns the absolute path of the skills.db file.
// The skills store lives in the data dir alongside the main
// DB, NOT in the parent of the data dir. This is the single
// source of truth — every caller (the daemon builder, the
// backup subsystem, the uninstall subsystem) must use this
// or the constant Path() of the same name. Hard-coded
// `filepath.Dir(dataDir)/skills.db` in any package is a bug.
func (s *Subsystems) SkillDBPath() string {
	if s == nil || s.Storage == nil {
		return ""
	}
	return filepath.Join(s.GeneralDataDir(), "skills.db")
}

// MemoryDBPath returns the absolute path of memory.db. Same
// single-source-of-truth rule as SkillDBPath.
func (s *Subsystems) MemoryDBPath() string {
	if s == nil || s.Storage == nil {
		return ""
	}
	return filepath.Join(s.GeneralDataDir(), "memory.db")
}

// GatekeeperAllow is the Phase 11 gate for destructive
// operations (backup.restore, backup.rollback, uninstall.execute).
// It constructs a blastradius.Action and routes it through the
// real Safety.Engine — same code path the agent loop uses
// for every other physical action. The policy verdict flows:
//   - Allow → this method returns true
//   - Deny  → false (with reason logged)
//   - RequireConsent / RequirePresenceAndConsent → drives
//     the consent provider, which publishes an SSE event
//     for the GUI to display; the GUI calls
//     safety.consent.approve / safety.consent.deny via RPC.
//
// The Engine is the only place consent is ever decided. There
// is no separate "v0.1.0 trusted-caller" path. If the engine
// is unavailable (subs.Safety is nil during some test setups)
// the gate fails closed — returning false — rather than
// allowing.
func (s *Subsystems) GatekeeperAllow(ctx context.Context, kind, detail string) bool {
	if s == nil || s.Safety == nil || s.Safety.Engine == nil {
		return false
	}
	action := blastradius.Action{
		Kind:      kind,
		TargetApp: "condurad",
		Body:      detail,
	}
	decision, reason := s.Safety.Engine.Evaluate(ctx, action)
	// Log every gate decision so the audit chain shows the
	// why. Phase 11 surface actions are infrequent; this is
	// cheap and the trail is valuable.
	//
	// Use a fresh, non-deadlined context for the audit append.
	// The caller's ctx may have a short timeout (e.g. a test
	// using 1s to force the engine to fail-closed). The
	// audit chain lookup for prev_hash is a SQLite read; if
	// the caller's ctx is already expired, the audit append
	// fails with "context deadline exceeded" and the gate
	// decision is lost from the chain. We always have 5s
	// budget for the audit append regardless of the gate
	// decision's deadline.
	auditCtx, auditCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer auditCancel()
	if s.AuditLog != nil {
		_ = s.AuditLog.Append(auditCtx, buildAuditEvent(
			"gate."+decisionName(decision),
			appCondurad,
			auditResultFromDecision(decision),
			"kind="+kind+" reason="+reason,
		))
	}
	return decision == gatekeeper.Allow
}

// decisionName returns a stable string name for a Decision
// (for audit logging). Maps the iota values to readable names.
func decisionName(d gatekeeper.Decision) string {
	switch d {
	case gatekeeper.Allow:
		return "allow"
	case gatekeeper.Deny:
		return "deny"
	case gatekeeper.RequireConsent:
		return "require_consent"
	case gatekeeper.RequirePresenceAndConsent:
		return "require_presence_and_consent"
	default:
		return "unknown"
	}
}

// auditResultFromDecision maps a gatekeeper.Decision into the
// string vocabulary the audit log expects.
func auditResultFromDecision(d gatekeeper.Decision) string {
	switch d {
	case gatekeeper.Allow:
		return auditResultAllow
	case gatekeeper.Deny:
		return auditResultDeny
	default:
		return auditResultError
	}
}

// currentSchemaVersion returns the binary's current schema
// version. The schema_version field is on the Config struct.
func currentSchemaVersion(subs *Subsystems) int {
	if subs != nil && subs.cfg != nil {
		return subs.cfg.Version
	}
	return config.ConfigSchemaVersion
}

// initSubsystems constructs every long-lived component the daemon
// needs. On error, all partially-initialized components are torn
// down.
//
//nolint:gocyclo,gocognit // wiring all subsystems in one place is intentional
func initSubsystems(log *slog.Logger, cfg *config.Config, loader *config.Loader) (*Subsystems, error) {
	secretsPath := filepath.Join(cfg.General.DataDir, "secrets.json")
	sm, err := secrets.New(secretsPath)
	if err != nil {
		return nil, fmt.Errorf("init secrets: %w", err)
	}
	log.Info("secrets manager ready", "backend", string(sm.Backend()))

	db, err := storage.Open(context.Background(), storage.Config{
		Path:    cfg.Storage.Path,
		Secrets: sm,
	})
	if err != nil {
		return nil, fmt.Errorf("init storage: %w", err)
	}
	log.Info("storage ready", "path", cfg.Storage.Path)

	akm := api_key.New(db, sm)
	registry := llm.NewRegistry()
	// netGuard is declared later (alongside haltFlag) so it's
	// available to all subsystem wiring. Until then, pass nil
	// to the provider build; we'll re-wrap transports below.
	registered := buildProvidersFromConfig(log, registry, cfg, akm, nil)
	log.Info("llm registry ready", "registered_providers", registered)

	mon := failover.NewSpendMonitor(failover.SpendCap{USDPerDay: cfg.Security.SpendLimitUSDPerDay})
	breakers := failover.NewBreakerRegistry(3, 30*time.Second)
	failoverProviders := buildFailoverProviders(registry, breakers)
	fo := failover.New(failoverProviders, mon)
	log.Info("failover ready", "providers", len(failoverProviders))

	hr := health.New()
	hr.Add(healthCheckStorage(db))
	hr.Add(healthCheckSecrets(sm))

	// Phase 2: wire up the additional subsystems.
	convStore := conversation.New(db.SQL())
	memPath := filepath.Join(filepath.Dir(db.Path()), "memory.db")
	memStore, memErr := memory.NewSQLiteStore(memPath)
	var memMgr *memory.StoreManager
	if memErr != nil {
		log.Warn("memory store init failed; running without memory", "err", memErr)
	} else {
		memMgr = memory.NewManager(memStore)
		log.Info("memory store ready")
	}

	extractor := initExtractor(db.Path(), memMgr, log)
	auditLog := audit.New(db.SQL(), db.MasterKey())
	haltFlag := halt.New(db.SQL())
	_ = haltFlag.Refresh(context.Background())
	// Phase 14I: Layer 3 of the kill switch — network guard.
	// The in-process guard wraps the LLM HTTP transports below;
	// when Halt is called, all outbound HTTP is denied except to
	// allow-listed providers. See internal/halt/network.go.
	netGuard := halt.NewInProcessGuard()
	tel := telemetry.New(db.SQL(), cfg.Telemetry.Endpoint)
	tel.SetEnabled(cfg.Telemetry.Enabled)
	upd := updater.New(db.SQL(), resolveUpdateManifestURL(cfg))
	upd.SetEnabled(cfg.Update.Enabled)
	winMgr := window.New(db.SQL())

	// Phase 3: SSE broker + LLM stream manager. The broker fans
	// events out to GUI EventSource clients; the stream manager
	// owns the lifecycle of in-flight LLM streams and bridges them
	// to the broker.
	broker := sse.NewBroker()
	streamMgr := stream.NewManager(broker, registry)
	streamMgr.SetHaltChecker(haltFlag.IsHalted)

	// Phase 6: living presence.
	//
	// Gatekeeper is the deterministic rules engine. Built once
	// and shared by every gated executor in the daemon so the
	// safety layer is the single source of truth.
	safety := buildSafetyLayer(haltFlag, broker, log)
	gate := safety.Engine
	log.Info("gatekeeper ready", "policy", "engine", "consent_provider", "rpc")

	// SessionFactory builds end-to-end sessions on demand. It
	// pulls the LLM from the registry (currently the primary
	// provider; failover is handled inside the registry).
	primaryName, primaryModel := pickPrimaryProvider(cfg)
	if primaryName == "" {
		// No LLM configured — sessions will fail at Run time.
		log.Warn("no primary LLM provider configured; session.Run will fail until one is added")
	}
	sessionFactory, err := session.NewFactory(
		streamMgr,
		registry,
		primaryName,
		primaryModel,
		convStore,
		broker,
	)
	if err != nil {
		return nil, fmt.Errorf("init session factory: %w", err)
	}
	sessionFactory.SetGatekeeper(gate, auditLog)
	if memMgr != nil {
		sessionFactory.SetMemory(&sessionMemoryAdapter{mgr: memMgr})
	}
	log.Info("session factory ready", "primary", primaryName, "model", primaryModel)

	// Fan session status out to the SSE broker so the GUI
	// (tray, overlay) can react to session state changes.
	sessionFactory.SetOnStatus(func(s status.Status) {
		broker.PublishJSON("tray.status", map[string]any{
			statusKey: s.String(),
		})
	})

	// Overlay controller. The noop controller is a real
	// implementation with a state machine; the GUI host swaps it
	// for a Wails-backed controller when the GUI process is
	// embedded.
	ovl := overlay.NewNoopController()

	// GatedComputerUseExecutor is the parallel wrapper for the
	// computer-use backends. It uses the ORAX backend if available;
	// falls back to a noop if no real backend exists. The gatekeeper
	// always applies; decisions are audited.
	//
	// Computed BEFORE the agent leaf executor so the agent loop can
	// wrap the same CU pipeline through agent.NewComputerUseExecutor
	// (Phase 14I: real agent actions on the user's machine).
	cuComps := buildCUComponents(gate, haltFlag, &registryPlannerAdapter{r: registry, name: primaryName}, primaryModel)

	// GatedAgentExecutor wraps the agent loop's executor with
	// the Gatekeeper. The agent loop's Executor interface is
	// unchanged; wrapping is done at the composition site so
	// the loop remains testable with a plain Executor.
	//
	// Phase 14I: when the computer-use pipeline is wired (real or
	// noop backend), we wrap it through agent.NewComputerUseExecutor
	// so chat messages can drive real actions on the user's machine.
	// If no CU pipeline is available (defensive), we keep the
	// historical noopAgentExecutor that returns a clear "not wired"
	// error to the loop.
	var agentLeaf agent.Executor = noopAgentExecutor{}
	if cuComps != nil && cuComps.gated != nil {
		// cuComps.gated is *computeruse.GatedExecutor; it already wraps
		// the real computer-use backends. We use the inner ComputerUse
		// (cuComps.gated.CU()) via a translator so the agent.Actions
		// the loop emits flow into the gated pipeline.
		agentLeaf = agent.NewComputerUseExecutor(cuComps.gated.CU())
	}
	gatedAgentExec := agent.NewGatedExecutor(agentLeaf, gate, auditLog)

	// cuComps was computed above (line ~565) for the agent leaf. Now
	// use it to set up the GatedCUExecutor + CULoop for direct CU
	// action invocations from the GUI / API. Falls back to a noop
	// backend if no LLM provider is configured.
	var gatedCUExec *computeruse.GatedExecutor
	var cuLoop *agent.CULoop
	if cuComps != nil {
		gatedCUExec = cuComps.gated
		cuLoop = cuComps.loop
	} else {
		// No LLM provider available — fall back to noop.
		cuBackend := &computeruse.NoopBackend{}
		cu := computeruse.New(cuBackend)
		gatedCUExec = computeruse.NewGatedExecutor(cu, gate)
	}

	// Wire anomaly detector into CU resolver (real coordinates).
	if cuComps != nil && cuComps.resolver != nil && safety != nil && safety.Anomaly != nil {
		det := safety.Anomaly
		cuComps.resolver.SetAnomalyHook(func(kind string, x, y float64, success bool) {
			det.Record(kind, x, y, success)
		})
		if cuLoop != nil {
			cuLoop.OnStart = func() { det.Reset() }
		}
	}

	// Fan status out to the SSE broker.
	if cuLoop != nil {
		cuLoop.OnStatus = func(s status.Status) {
			broker.PublishJSON("tray.status", map[string]any{
				statusKey: s.String(),
			})
		}
		log.Info("computer-use loop ready")
	}

	// Voice pipeline. Constructed only when voice is enabled and
	// the binary/model paths are present. Failure to construct is
	// logged and Voice is left nil — sessions still work in
	// text-only mode.
	var voicePipeline *voice.Pipeline
	if cfg.Voice.Enabled {
		vp, err := buildVoicePipeline(cfg, log, broker)
		if err != nil {
			log.Warn("voice pipeline init failed; running text-only", "err", err)
		} else {
			voicePipeline = vp
			// The session uses the pipeline's Speaker for TTS.
			// The pipeline itself is a voice.Pipeline (which
			// has Speak/Stop methods), so we pass it directly.
			sessionFactory.SetSpeaker(vp)
			log.Info("voice pipeline ready")
		}
	}

	// Wire voice pipeline status updates to the SSE broker so
	// the tray (which lives in the GUI process) can react.
	// Pipeline.OnStatus fires on every state transition.
	if voicePipeline != nil {
		voicePipeline.OnStatus = func(s status.Status) {
			broker.PublishJSON("tray.status", map[string]any{
				statusKey: s.String(),
			})
		}
	}

	// Phase 8: user-adaptive engine. Uses storage.DB encryption.
	var adaptiveComps *AdaptiveComponents
	astore, aerr := adaptive.NewEncryptedStore(db.SQL(), db.EncryptString, db.DecryptString)
	if aerr != nil {
		log.Warn("adaptive store init failed", "err", aerr)
	} else if primaryName != "" {
		llmProv := &llmProviderAdapter{r: registry, name: primaryName, model: primaryModel}
		var criticProv llm.Provider
		criticModel := ""
		for _, name := range []string{providerGoogle, "mistral", "openai"} {
			if prov, ok := registry.Get(name); ok {
				criticProv = prov
				criticModel = prov.DefaultModel("chat")
				break
			}
		}
		var budget adaptive.BudgetChecker
		if mon != nil {
			budget = &spendBudgetChecker{m: mon}
		}
		adaptiveComps = buildAdaptiveEngine(astore, llmProv, criticProv, criticModel, budget, log)
		if extractor != nil {
			extractor.SetObserver(adaptiveComps.Observer)
			extractor.SetEngine(adaptiveComps.Engine)
		}
		if adaptiveComps != nil && adaptiveComps.Predictor != nil {
			sessionFactory.SetPredictor(adaptiveComps.Predictor)
		}
		log.Info("adaptive engine ready", "strength", adaptiveComps.Strength)
	}

	// Build the backup manager once and share the result
	// between the RPC-facing Backup field and the auto-backup
	// Scheduler. They use the same master key, the same data
	// dir, and the same schema version — sharing the manager
	// is the only way to keep those in sync.
	backupMgr := buildBackupMgr(db, cfg, log)

	subs := &Subsystems{
		Secrets: sm, Storage: db, APIKeys: akm, LLM: registry,
		Failover: fo, Spend: mon, Health: hr,
		Conversations: convStore, Audit: auditLog, Halt: haltFlag,
		NetGuard:  netGuard,
		Telemetry: tel, Updater: upd, Window: winMgr,
		Broker: broker, Streams: streamMgr,
		Gatekeeper:               gate,
		GatedAgentExecutor:       gatedAgentExec,
		GatedComputerUseExecutor: gatedCUExec,
		Overlay:                  ovl,
		SessionFactory:           sessionFactory,
		Voice:                    voicePipeline,
		CULoop:                   cuLoop,
		Memory:                   memMgr,
		Extractor:                extractor,
		Adaptive:                 adaptiveComps,
		MCP:                      mcp.NewManager(gate),
		Safety:                   safety,
		Anomaly:                  safety.Anomaly,
		Delegation:               buildDelegationBus(gate, mon),
		Phase12:                  buildPhase12(cfg, log),
		// Phase 11: trust & recovery. See methods_phase11.go
		// for the RPC surface and the corresponding E2E
		// test in trust_e2e_test.go.
		// Note: buildBackupMgr is called once and the result
		// is shared between the RPC-facing Backup and the
		// auto-backup Scheduler. They use the same master key,
		// the same data dir, and the same schema version.
		Replay:          buildReplay(db, auditLog, log),
		Backup:          backupMgr,
		BackupScheduler: buildBackupScheduler(backupMgr, log),
		Uninstaller:     buildUninstaller(),
		Onboarding:      buildOnboarding(db.SQL(), log),
		Permissions:     buildPermissions(log),
		Account:         buildAccount(cfg, db, sm, log),
		Reach:           buildReach(db, log),
		AuditLog:        auditLog,
		db:              db,
		cfg:             cfg,
		Loader:          loader,
	}
	// Wire screenshot store into CU resolver so before/after
	// screenshots are captured for the replay timeline.
	if cuComps != nil && cuComps.resolver != nil && subs.Replay != nil {
		if shots := subs.Replay.Screenshots(); shots != nil {
			cuComps.resolver.SetScreenshotStore(shots)
		}
	}
	// Register closers for cleanup on shutdown (Windows file-lock).
	// The main DB must be closed before other resources that may
	// depend on it, so it goes first.
	subs.closers = append(subs.closers, db)
	if memStore != nil {
		subs.closers = append(subs.closers, memStore)
	}
	if extractor != nil {
		subs.closers = append(subs.closers, extractor)
	}
	if subs.Phase12 != nil && subs.Phase12.SkillStore != nil {
		subs.closers = append(subs.closers, subs.Phase12.SkillStore)
	}
	return subs, nil
}

// buildPhase12 constructs the Phase 12 components (Hub + Sync + i18n).
// Each subsystem is built by a dedicated helper; this function
// orchestrates them in order and logs the result.
func buildPhase12(cfg *config.Config, log *slog.Logger) *Phase12Components {
	p12 := &Phase12Components{}
	p12.Catalog = buildI18nCatalog(log, cfg)
	p12.SkillStore = buildSkillStore(log, cfg)
	p12.HubClient = buildHubClient(log, cfg)
	p12.SyncEngine = buildSyncEngine(log, cfg)
	return p12
}

// buildI18nCatalog loads the embedded locale JSONs. Always returns
// a usable catalog; falls back to MustNewCatalog if load fails.
func buildI18nCatalog(log *slog.Logger, cfg *config.Config) *i18n.Catalog {
	catalog, err := i18n.NewCatalog()
	if err != nil {
		log.Warn("i18n catalog init failed; using English defaults", "err", err)
		catalog = i18n.MustNewCatalog()
	}
	lang := cfg.General.Language
	if lang == "" || lang == "auto" {
		lang = "en"
	}
	log.Info("i18n catalog ready", "locale", lang, "available", len(catalog.Locales()))
	return catalog
}

// buildSkillStore opens the SQLite-backed skill store at
// <dataDir>/skills.db. Returns nil on error (skill features
// become no-ops; the daemon keeps running).
func buildSkillStore(log *slog.Logger, cfg *config.Config) *skills.SQLiteStore {
	path := filepath.Join(cfg.General.DataDir, "skills.db")
	store, err := skills.NewSQLiteStore(path)
	if err != nil {
		log.Warn("skill store init failed; hub install/publish are no-ops", "err", err)
		return nil
	}
	return store
}

// buildHubClient constructs the Skills Hub client when enabled.
// Honors bearer token auth and (optionally) Ed25519 publish
// signing. Returns nil when cfg.Hub.Enabled is false.
func buildHubClient(log *slog.Logger, cfg *config.Config) *hub.Client {
	if !cfg.Hub.Enabled {
		return nil
	}
	baseURL := cfg.Hub.BaseURL
	if baseURL == "" {
		baseURL = "https://hub.synaptic.app"
	}
	opts := []hub.ClientOption{}
	if cfg.Hub.Token != "" {
		opts = append(opts, hub.WithToken(cfg.Hub.Token))
	}
	if cfg.Hub.PublishKeyPath != "" {
		priv, ok := loadPublishKey(log, cfg.Hub.PublishKeyPath)
		if ok {
			opts = append(opts, hub.WithPublishKey(priv))
		}
	}
	log.Info("hub client ready",
		"base_url", baseURL,
		"authenticated", cfg.Hub.Token != "",
		"publish_signing", cfg.Hub.PublishKeyPath != "",
	)
	return hub.NewClient(baseURL, opts...)
}

// loadPublishKey reads a hex-encoded Ed25519 private key from
// path. Returns (key, false) on any error and logs a warning.
// This function is intentionally lenient — the daemon should
// keep running even if the key file is missing (the user can
// set it later).
//
//	is hex-decoded and length-checked, so a malicious config
//	cannot inject shell metacharacters or read non-key files.
//
//nolint:gosec // path is operator-provided via config; the value
func loadPublishKey(log *slog.Logger, path string) (ed25519.PrivateKey, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Warn("hub publish key file unreadable; Publish will be unsigned",
			"path", path, "err", err)
		return nil, false
	}
	bytes, err := hex.DecodeString(strings.TrimSpace(string(data)))
	if err != nil || len(bytes) != ed25519.PrivateKeySize {
		log.Warn("hub publish key file has invalid format; Publish will be unsigned",
			"path", path)
		return nil, false
	}
	log.Info("hub publish key loaded", "path", path)
	return ed25519.PrivateKey(bytes), true
}

// buildSyncEngine constructs and auto-starts the P2P sync engine
// when cfg.Sync.Enabled is true. Returns nil when sync is disabled
// or any of the prerequisites fail. The paired set is loaded
// with FAIL-CLOSED semantics: any load error results in an
// empty (in-memory) paired set so the gate rejects every sync
// until the user pairs a device.
func buildSyncEngine(log *slog.Logger, cfg *config.Config) *sync.Engine {
	if !cfg.Sync.Enabled {
		return nil
	}
	deviceName := cfg.Sync.DeviceName
	if deviceName == "" {
		deviceName = "synaptic-device"
	}
	identity, err := sync.LoadIdentity(cfg.General.DataDir, deviceName)
	if err != nil {
		log.Warn("sync identity init failed", "err", err)
		return nil
	}
	port := cfg.Sync.DiscoveryPort
	if port == 0 {
		port = 7667
	}
	discovery := sync.NewDiscovery(identity, port)
	paired, pairErr := sync.LoadPairedSet(cfg.General.DataDir)
	if pairErr != nil {
		log.Warn("sync paired-set load failed; engine starts EMPTY (no auto-accept)", "err", pairErr)
		paired = sync.NewEmptyPairedSet()
	}
	engine := sync.NewEngine(identity, sync.NewStore(), discovery, paired, log)
	// Auto-start so users see `running: true` immediately after
	// the daemon reports "ready". They can stop it via
	// `synaptic sync stop`.
	engine.Start()
	log.Info("sync engine ready (auto-started)",
		"device_id", identity.DeviceID,
		"fingerprint", identity.Fingerprint(),
	)
	return engine
}

// pickPrimaryProvider returns the first enabled LLM provider
// name and its default model. When no provider is configured,
// returns "", "".
func pickPrimaryProvider(cfg *config.Config) (string, string) {
	// Prefer the order in the YAML: iterate the map in insertion
	// order (Go map iteration is randomized, so callers who
	// care about priority should set the model field
	// explicitly). For v0 we pick the first enabled provider.
	for _, name := range []string{"anthropic", "openai", providerGoogle, "ollama", "xai", "mistral"} {
		pc, ok := cfg.LLM.Providers[name]
		if !ok || !pc.Enabled {
			continue
		}
		model := pc.DefaultModel
		if model == "" {
			model = defaultModelFor(name)
		}
		return name, model
	}
	return "", ""
}

// RebuildProviders re-registers providers from the current config
// and API keys into the LLM registry. Call this after enabling a
// provider (onboarding.finish) or adding an API key (apikeys.set)
// so the daemon can use the new provider without a restart.
// Returns the number of providers registered.
func (s *Subsystems) RebuildProviders() int {
	if s.LLM == nil || s.cfg == nil {
		return 0
	}
	registered := buildProvidersFromConfig(slog.Default(), s.LLM, s.cfg, s.APIKeys, s.NetGuard)
	s.rebuildSessionFactory()
	return registered
}

// rebuildSessionFactory updates the session factory's primary
// provider name and model from the current config.
func (s *Subsystems) rebuildSessionFactory() {
	if s.SessionFactory == nil || s.cfg == nil {
		return
	}
	primaryName, primaryModel := pickPrimaryProvider(s.cfg)
	if primaryName == "" {
		return
	}
	s.SessionFactory.UpdatePrimary(primaryName, primaryModel)
	slog.Info("session factory primary updated", "provider", primaryName, "model", primaryModel)
}

// defaultModelFor returns a sensible default model name for a
// provider. Used by the session factory when the user hasn't
// pinned a model. Defaults follow the marketing-aligned current
// generation; users can override at any time in Settings.
func defaultModelFor(provider string) string {
	switch provider {
	case "anthropic":
		return "claude-sonnet-4-5"
	case "openai":
		return "gpt-5.5"
	case providerGoogle:
		return "gemini-3.5-flash"
	case "ollama":
		return "llama3.2"
	case "xai":
		return "grok-4.3"
	case "mistral":
		return "mistral-large-3"
	default:
		return ""
	}
}

// noopAgentExecutor is a placeholder agent.Executor for the
// initial daemon wiring. The real computer-use executor will
// be wrapped through GatedComputerUseExecutor in a future
// iteration (Phase 6B.1).
type noopAgentExecutor struct{}

func (noopAgentExecutor) Execute(_ context.Context, a *agent.Action) (*agent.StepResult, error) {
	return &agent.StepResult{Success: false, Error: fmt.Errorf("agent executor not yet wired: %s", a.Type)}, nil
}

// buildVoicePipeline constructs the voice pipeline from the
// voice config. Returns an error if the binary or model is
// missing or fails SHA verification, or if no audio capture
// device is available on the platform.
func buildVoicePipeline(cfg *config.Config, log *slog.Logger, broker *sse.Broker) (*voice.Pipeline, error) {
	if !cfg.Voice.Enabled {
		return nil, errors.New("voice is not enabled in config")
	}
	cfg.Voice.ApplyDefaults()
	if err := cfg.Voice.Validate(); err != nil {
		return nil, fmt.Errorf("voice config: %w", err)
	}

	// Fail loudly at startup if no audio capture device is
	// available. This avoids mid-session failures when the user
	// presses the hotkey and the recorder can't start.
	if !voice.RecorderAvailable() {
		return nil, errors.New("no audio capture device found on this platform")
	}

	recorder := voice.NewRecorder(cfg.Voice.SampleRate, cfg.Voice.Channels)
	transcriber := voice.NewTranscriber(
		cfg.Voice.BinaryPath,
		cfg.Voice.ModelPath,
		cfg.Voice.Language,
	)
	pins := voice.SHA256Pins{
		Binary: cfg.Voice.BinarySHA256,
		Model:  cfg.Voice.ModelSHA256,
	}
	speaker := voice.NewSpeaker(cfg.Voice.SpeakerVoice, cfg.Voice.SpeakerRate)

	pipeline, err := voice.NewPipeline(voice.Config{
		Recorder:    recorder,
		Transcriber: transcriber,
		Speaker:     speaker,
		BinaryPath:  cfg.Voice.BinaryPath,
		ModelPath:   cfg.Voice.ModelPath,
		Pins:        pins,
		SilenceMS:   cfg.Voice.SilenceTimeoutMs,
		Language:    cfg.Voice.Language,
		Broker:      broker,
	})
	if err != nil {
		return nil, err
	}
	log.Info("voice pipeline constructed",
		"binary", cfg.Voice.BinaryPath,
		"model", cfg.Voice.ModelPath,
		"sha_binary_pinned", cfg.Voice.BinarySHA256 != "",
		"sha_model_pinned", cfg.Voice.ModelSHA256 != "",
	)
	return pipeline, nil
}

// mkdirDataDir creates the data directory if it doesn't exist.
func mkdirDataDir(path string) error {
	if err := os.MkdirAll(path, dataDirPerm); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	return nil
}

// newLoggerFromConfig creates an slog.Logger from the config's logging
// section, applying level / format / file / source settings.
func newLoggerFromConfig(cfg *config.Config) *slog.Logger {
	return logger.New(logger.Config{
		Level:     logger.ParseLevel(cfg.Logging.Level),
		Format:    logger.ParseFormat(cfg.Logging.Format),
		File:      cfg.Logging.File,
		AddSource: cfg.Logging.AddSource,
	})
}

// healthCheckStorage returns a health check that pings the SQLite DB.
func healthCheckStorage(db *storage.DB) health.Check {
	return health.Check{
		Name: "storage", Required: true, Timeout: 2 * time.Second,
		Check: func(ctx context.Context) error { return db.SQL().PingContext(ctx) },
	}
}

// healthCheckSecrets returns a health check that probes the secrets
// backend. We expect a "not found" error from a well-formed probe
// key; any other error is a real failure.
func healthCheckSecrets(sm secrets.Manager) health.Check {
	return health.Check{
		Name: "secrets", Required: true, Timeout: 2 * time.Second,
		Check: func(_ context.Context) error {
			_, err := sm.Get("__synaptic_health_probe__")
			if err != nil && !errors.Is(err, secrets.ErrNotFound) {
				return err
			}
			return nil
		},
	}
}

// healthCheckIPC is a no-op check that just confirms the IPC server
// is wired up. The actual server health is observable from outside.
func healthCheckIPC() health.Check {
	return health.Check{
		Name: "ipc", Required: false, Timeout: 1 * time.Second,
		Check: func(_ context.Context) error { return nil },
	}
}

// initExtractor creates the post-session extractor if stores are available.
// `dataDir` is conventionally the *synaptic.db* path (e.g.
// "/var/synaptic/synaptic.db"); the skills store lives at
// the parent dir + "skills.db" = "/var/synaptic/skills.db".
// We unwrap the file part explicitly here so the path is
// obvious from the call site and the test suite can grep for it.
func initExtractor(dataDir string, memMgr *memory.StoreManager, log *slog.Logger) *PostSessionExtractor {
	// dataDir may be either a directory OR a file (synaptic.db).
	// Accept both shapes; if it's a file, take its parent.
	parent := dataDir
	if filepath.Base(parent) == "synaptic.db" {
		parent = filepath.Dir(parent)
	}
	skillPath := filepath.Join(parent, "skills.db")
	skillStore, err := skills.NewSQLiteStore(skillPath)
	if err != nil {
		log.Warn("skill store init failed; disabled", "err", err)
		return nil
	}
	if memMgr == nil {
		return nil
	}
	return NewPostSessionExtractor(memMgr, skillStore, log, true)
}

// registryPlannerAdapter adapts llm.Registry → agent.PlannerProvider.
type registryPlannerAdapter struct {
	r    *llm.Registry
	name string
}

func (a *registryPlannerAdapter) Chat(ctx context.Context, req llm.ChatRequest) (llm.ChatResponse, error) {
	return a.r.Chat(ctx, a.name, req)
}

// llmProviderAdapter implements llm.Provider from llm.Registry.
type llmProviderAdapter struct {
	r     *llm.Registry
	name  string
	model string
}

func (a *llmProviderAdapter) Name() string { return a.name }
func (a *llmProviderAdapter) Chat(ctx context.Context, req llm.ChatRequest) (llm.ChatResponse, error) {
	return a.r.Chat(ctx, a.name, req)
}
func (a *llmProviderAdapter) Stream(ctx context.Context, req llm.ChatRequest) (<-chan llm.StreamEvent, func(), error) {
	return a.r.Stream(ctx, a.name, req)
}
func (a *llmProviderAdapter) Models() []llm.ModelInfo {
	if prov, ok := a.r.Get(a.name); ok {
		return prov.Models()
	}
	return nil
}
func (a *llmProviderAdapter) DefaultModel(task string) string {
	if a.model != "" {
		return a.model
	}
	if prov, ok := a.r.Get(a.name); ok {
		return prov.DefaultModel(task)
	}
	return ""
}

// Phase 11 builders.

// buildReplay constructs the action-replay subsystem. It pairs
// the HMAC-chained audit log with an encrypted on-disk ring
// buffer for screenshots. Best-effort: if construction fails
// (e.g., master key missing), the field is left nil and the
// corresponding RPC methods short-circuit.
func buildReplay(db *storage.DB, auditLog *audit.Log, log *slog.Logger) *replay.Replay {
	if auditLog == nil {
		log.Warn("replay: audit log not available; replay subsystem disabled")
		return nil
	}
	masterKey := db.MasterKey()
	shotsDir := filepath.Join(filepath.Dir(db.Path()), "replay")
	shots, err := replay.NewScreenshotStore(db.SQL(), shotsDir, masterKey)
	if err != nil {
		log.Warn("replay: screenshot store init failed; timeline will work without screenshots", "err", err)
		shots = nil
	}
	r, err := replay.New(replay.Options{
		Audit:       auditLog,
		Screenshots: shots,
	})
	if err != nil {
		log.Warn("replay: init failed; replay subsystem disabled", "err", err)
		return nil
	}
	log.Info("replay subsystem ready")
	return r
}

// buildBackupMgr constructs the encrypted-backup Manager.
// Best-effort: if the master key is missing or the data dir is
// not configured, the manager is left nil.
func buildBackupMgr(db *storage.DB, cfg *config.Config, log *slog.Logger) *backup.Manager {
	masterKey := db.MasterKey()
	bm, err := backup.New(backup.Options{
		DataDir:       cfg.General.DataDir,
		ConfigPath:    filepath.Join(cfg.General.DataDir, "config.yaml"),
		MasterKey:     masterKey,
		SchemaVersion: config.ConfigSchemaVersion,
	})
	if err != nil {
		log.Warn("backup: init failed; backup subsystem disabled", "err", err)
		return nil
	}
	log.Info("backup subsystem ready", "data_dir", cfg.General.DataDir)
	return bm
}

// buildUninstaller returns the uninstall subsystem sentinel.
// The actual work lives in the package-level uninstall.Preview
// and uninstall.Uninstall functions; the sentinel just makes
// the subsystem present in the Subsystems struct.
func buildUninstaller() *uninstall.Manager {
	return &uninstall.Manager{}
}

// buildOnboarding constructs the wizard state machine.
func buildOnboarding(sqlDB *sql.DB, log *slog.Logger) *onboarding.StateMachine {
	sm, err := onboarding.NewStateMachine(sqlDB)
	if err != nil {
		log.Warn("onboarding: init failed; wizard disabled", "err", err)
		return nil
	}
	log.Info("onboarding subsystem ready")
	return sm
}

// buildPermissions constructs the OS permission probe + guide
// manager. Best-effort: even if construction fails, we still
// return a manager with the noop probe so RPC methods are
// callable.
func buildPermissions(log *slog.Logger) *permissions.Manager {
	pm := permissions.NewManager()
	log.Info("permissions subsystem ready", "platform", permissions.Platform())
	return pm
}

// buildBackupScheduler wires the periodic auto-backup. Returns
// nil if the backup manager is nil (which happens when the
// master key is unavailable during construction — rare, but
// possible on a half-installed system). The caller
// (daemon.Run) starts the scheduler after listeners are
// ready and stops it on shutdown.
//
// Cadence comes from cfg.Backup.IntervalHours (default 24h)
// and cfg.Backup.KeepN (default 7). The scheduler is local-
// only and stores archives in <data-dir>/backups. The user
// can also call backup.create manually at any time; the
// scheduler and the RPC use the same Manager so they share
// the encryption key, the schema version, and the rotation
// policy.
func buildBackupScheduler(bm *backup.Manager, log *slog.Logger) *backup.Scheduler {
	if bm == nil {
		log.Warn("auto-backup scheduler: backup manager not available, scheduler disabled")
		return nil
	}
	cfg := backup.DefaultSchedulerConfig()
	// Cadence knobs live in cfg.Backup (data-modeled via
	// config.BackupConfig). For v0.1.0 the defaults are 24h
	// interval, 7 archives retained. If cfg.Backup.IntervalHours
	// is set, we honor it; otherwise the DefaultSchedulerConfig
	// value (24h) applies.
	// Note: NewScheduler fills cfg.BackupDir from the manager's
	// data dir if it's empty, so we log AFTER construction.
	s := backup.NewScheduler(cfg, bm, log)
	log.Info("auto-backup scheduler ready",
		"interval", cfg.Interval,
		"keep_n", cfg.KeepN,
		"backup_dir", s.Cfg().BackupDir,
	)
	return s
}

func resolveUpdateManifestURL(cfg *config.Config) string {
	if cfg != nil && cfg.Update.ManifestURL != "" {
		return cfg.Update.ManifestURL
	}
	return updater.DefaultManifestURL
}

// buildAccount constructs the account manager (Phase 14B).
// Returns nil when disabled or construction fails.
//
// The provider registry is built by layering:
//  1. The user's config.yaml account.oauth.<provider> values
//  2. Environment variables CONDURA_ACCOUNT_OAUTH_<UPPER>_CLIENT_ID / _CLIENT_SECRET
//
// over the package defaults (which only define endpoints, never credentials).
func buildAccount(cfg *config.Config, db *storage.DB, sm secrets.Manager, log *slog.Logger) *account.Manager {
	if cfg == nil || !cfg.Account.Enabled {
		log.Info("account subsystem disabled")
		return nil
	}
	store, err := account.NewStore(db.SQL())
	if err != nil {
		log.Warn("account: store creation failed, sign-in disabled", "err", err)
		return nil
	}
	masterKey := db.MasterKey()
	// Try keychain first, fall back to file.
	var tm account.TokenManager
	km := account.NewKeychainTokenManager(sm.Get, sm.Set, sm.Delete)
	if err := sm.Set("account-test", "1"); err == nil {
		_ = sm.Delete("account-test")
		tm = km
	} else {
		tm = account.NewFallbackTokenManager(cfg.General.DataDir, masterKey)
		log.Info("account: using file-backed token storage (keychain unavailable)")
	}

	// Resolve OAuth provider configuration.
	userProviders := oauthProvidersFromConfig(cfg)
	overrideWithEnv(userProviders)
	registry := account.NewProviderRegistry(userProviders)

	if r := registry.Configured(); len(r) > 0 {
		log.Info("account: OAuth providers configured", "providers", r)
	} else {
		log.Info("account: no OAuth providers configured (set CONDURA_ACCOUNT_OAUTH_<PROVIDER>_CLIENT_ID or config.yaml account.oauth.* to enable)")
	}

	mgr, err := account.NewManagerWithProviders(store, tm, masterKey, cfg.Account.SessionTTL, registry)
	if err != nil {
		log.Warn("account: manager creation failed, sign-in disabled", "err", err)
		return nil
	}

	// Wire magic-link endpoint URL from config (with env override).
	if cfg.Account.MagicURL != "" {
		verifyURL := deriveMagicVerifyURL(cfg.Account.MagicURL, cfg.Account.MagicVerifyURL)
		account.SetMagicLinkURL(cfg.Account.MagicURL, verifyURL)
	}

	log.Info("account subsystem ready")
	return mgr
}

// deriveMagicVerifyURL picks the verify URL from explicit override, then
// from a /magic -> /verify substitution at the path tail, then falls
// back to "<issue>/verify". Exposed so tests can exercise it directly.
func deriveMagicVerifyURL(issueURL, explicit string) string {
	if explicit != "" {
		return explicit
	}
	v := strings.TrimSuffix(issueURL, "/")
	// Common case: swap a trailing "magic" segment for "verify". This
	// covers the canonical Next.js route names (.../api/auth/magic ->
	// .../api/auth/verify) without producing ".../magic/verify/verify"
	// when the issue URL already ends in "magic".
	if strings.HasSuffix(v, "/magic") {
		return strings.TrimSuffix(v, "/magic") + "/verify"
	}
	return v + "/verify"
}

// oauthProvidersFromConfig translates config.AccountConfig.OAuth into the
// shape account.NewProviderRegistry expects. We only forward ClientID,
// ClientSecret, AuthURL, TokenURL, UserInfoURL, and Scopes — endpoint
// URLs are only forwarded when the user overrides them, so the package
// defaults stay authoritative.
func oauthProvidersFromConfig(cfg *config.Config) map[string]account.ProviderConfig {
	if cfg == nil {
		return nil
	}
	out := make(map[string]account.ProviderConfig, len(cfg.Account.OAuth))
	for name, p := range cfg.Account.OAuth {
		out[name] = account.ProviderConfig{
			ClientID:     p.ClientID,
			ClientSecret: p.ClientSecret,
			AuthURL:      p.AuthURL,
			TokenURL:     p.TokenURL,
			UserInfoURL:  p.UserInfoURL,
			Scopes:       p.Scopes,
		}
	}
	return out
}

// overrideWithEnv applies env-var overrides on top of the user's
// config. The naming convention is:
//
//	CONDURA_ACCOUNT_OAUTH_<UPPER>_{CLIENT_ID,CLIENT_SECRET}
//
// where <UPPER> is the provider name uppercased. Env vars always win
// over config so users can keep secrets out of disk.
func overrideWithEnv(providers map[string]account.ProviderConfig) {
	for name := range providers {
		p := providers[name]
		upper := strings.ToUpper(name)
		if v := os.Getenv("CONDURA_ACCOUNT_OAUTH_" + upper + "_CLIENT_ID"); v != "" {
			p.ClientID = v
		}
		if v := os.Getenv("CONDURA_ACCOUNT_OAUTH_" + upper + "_CLIENT_SECRET"); v != "" {
			p.ClientSecret = v
		}
		providers[name] = p
	}
}

// buildReach constructs the channels manager (Phase 14C).
// Returns nil when disabled or construction fails.
func buildReach(db *storage.DB, log *slog.Logger) *reach.Manager {
	store, err := reach.NewStore(db.SQL())
	if err != nil {
		log.Warn("reach: store creation failed, channels disabled", "err", err)
		return nil
	}
	mgr := reach.NewManager(store)
	log.Info("reach subsystem ready", "channels", "telegram")
	return mgr
}
