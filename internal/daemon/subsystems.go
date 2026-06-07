package daemon

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/api_key"
	"github.com/sahajpatel123/synapticapp/internal/config"
	"github.com/sahajpatel123/synapticapp/internal/failover"
	"github.com/sahajpatel123/synapticapp/internal/health"
	"github.com/sahajpatel123/synapticapp/internal/llm"
	"github.com/sahajpatel123/synapticapp/internal/logger"
	"github.com/sahajpatel123/synapticapp/internal/secrets"
	"github.com/sahajpatel123/synapticapp/internal/storage"
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
	Secrets  secrets.Manager
	Storage  *storage.DB
	APIKeys  *api_key.Manager
	LLM      *llm.Registry
	Failover *failover.Failover
	Spend    *failover.SpendMonitor
	Health   *health.Register
	IPCAddr  string // first listen addr (empty if IPC disabled)
}

// initSubsystems constructs every long-lived component the daemon
// needs. On error, all partially-initialized components are torn
// down.
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

	return &Subsystems{
		Secrets: sm, Storage: db, APIKeys: akm, LLM: registry,
		Failover: fo, Spend: mon, Health: hr,
	}, nil
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
