package daemon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/agent"
	"github.com/sahajpatel123/synapticapp/internal/api_key"
	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/computeruse"
	"github.com/sahajpatel123/synapticapp/internal/config"
	"github.com/sahajpatel123/synapticapp/internal/conversation"
	"github.com/sahajpatel123/synapticapp/internal/failover"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/halt"
	"github.com/sahajpatel123/synapticapp/internal/health"
	"github.com/sahajpatel123/synapticapp/internal/llm"
	"github.com/sahajpatel123/synapticapp/internal/logger"
	"github.com/sahajpatel123/synapticapp/internal/memory"
	"github.com/sahajpatel123/synapticapp/internal/overlay"
	"github.com/sahajpatel123/synapticapp/internal/secrets"
	"github.com/sahajpatel123/synapticapp/internal/session"
	"github.com/sahajpatel123/synapticapp/internal/skills"
	"github.com/sahajpatel123/synapticapp/internal/sse"
	"github.com/sahajpatel123/synapticapp/internal/status"
	"github.com/sahajpatel123/synapticapp/internal/storage"
	"github.com/sahajpatel123/synapticapp/internal/stream"
	"github.com/sahajpatel123/synapticapp/internal/telemetry"
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

	// closers holds resources that must be closed on shutdown.
	closers []io.Closer
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

// initSubsystems constructs every long-lived component the daemon
// needs. On error, all partially-initialized components are torn
// down.
//
//nolint:gocyclo // wiring all subsystems in one place is intentional
func initSubsystems(log *slog.Logger, cfg *config.Config) (*Subsystems, error) {
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
	registered := buildProvidersFromConfig(log, registry, cfg, akm)
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

	// Create skill store and async extractor.
	extractor := initExtractor(db.Path(), memMgr, log)
	auditLog := audit.New(db.SQL())
	haltFlag := halt.New(db.SQL())
	_ = haltFlag.Refresh(context.Background())
	tel := telemetry.New(db.SQL(), cfg.Telemetry.Endpoint)
	tel.SetEnabled(cfg.Telemetry.Enabled)
	upd := updater.New(db.SQL(), "https://synaptic.app/updates/manifest.json")
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
	gate := gatekeeper.NewDenyBeyondRead()
	log.Info("gatekeeper ready", "policy", "deny-beyond-read")

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

	// GatedAgentExecutor wraps the agent loop's executor with
	// the Gatekeeper. The agent loop's Executor interface is
	// unchanged; wrapping is done at the composition site so
	// the loop remains testable with a plain Executor.
	gatedAgentExec := agent.NewGatedExecutor(noopAgentExecutor{}, gate, auditLog)

	// GatedComputerUseExecutor is the parallel wrapper for the
	// computer-use backends. It uses the ORAX backend if available;
	// falls back to a noop if no real backend exists. The gatekeeper
	// always applies; decisions are audited.
	cuComps := buildCUComponents(gate, haltFlag)
	gatedCUExec := cuComps.gated
	cuLoop := cuComps.loop

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

	subs := &Subsystems{
		Secrets: sm, Storage: db, APIKeys: akm, LLM: registry,
		Failover: fo, Spend: mon, Health: hr,
		Conversations: convStore, Audit: auditLog, Halt: haltFlag,
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
	}
	// Register closers for cleanup on shutdown (Windows file-lock).
	if memStore != nil {
		subs.closers = append(subs.closers, memStore)
	}
	if extractor != nil {
		subs.closers = append(subs.closers, extractor)
	}
	return subs, nil
}

// pickPrimaryProvider returns the first enabled LLM provider
// name and its default model. When no provider is configured,
// returns "", "".
func pickPrimaryProvider(cfg *config.Config) (string, string) {
	// Prefer the order in the YAML: iterate the map in insertion
	// order (Go map iteration is randomized, so callers who
	// care about priority should set the model field
	// explicitly). For v0 we pick the first enabled provider.
	for _, name := range []string{"anthropic", "openai", "google", "ollama", "xai", "mistral"} {
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

// defaultModelFor returns a sensible default model name for a
// provider. Used by the session factory when the user hasn't
// pinned a model.
func defaultModelFor(provider string) string {
	switch provider {
	case "anthropic":
		return "claude-3-5-sonnet-20241022"
	case "openai":
		return "gpt-4o-mini"
	case "google":
		return "gemini-1.5-flash"
	case "ollama":
		return "llama3.2"
	case "xai":
		return "grok-beta"
	case "mistral":
		return "mistral-large-latest"
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
func initExtractor(dataDir string, memMgr *memory.StoreManager, log *slog.Logger) *PostSessionExtractor {
	skillPath := filepath.Join(filepath.Dir(dataDir), "skills.db")
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
