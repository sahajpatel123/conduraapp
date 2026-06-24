// Package daemon is the in-process entry point for the Condura daemon.
// It is the same code path used by:
//
//   - cmd/condurad    — the standalone CLI daemon
//   - cmd/condura-gui — the Wails-based desktop GUI
//
// Run() is the single public entry point. It is safe to call from a
// goroutine; the caller is expected to cancel ctx to request a graceful
// shutdown.
//
// Concurrency: Run is NOT safe to call twice in the same process. Most
// subsystems (secrets manager, storage DB, log default) are singletons
// and will conflict.
//
// Single-instance enforcement: Run acquires an advisory flock on
// <data-dir>/condurad.lock at startup. If another process holds the
// lock, Run returns ErrAlreadyRunning. The lock is released when Run
// returns.
package daemon

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/anomaly"
	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/config"
	"github.com/sahajpatel123/synapticapp/internal/crash"
	"github.com/sahajpatel123/synapticapp/internal/lockfile"
	"github.com/sahajpatel123/synapticapp/internal/updater"
	"github.com/sahajpatel123/synapticapp/internal/version"
)

// ErrAlreadyRunning is returned by Run if another synaptic instance
// is already using the same data directory. The user-visible message
// should be "Another instance of Condura is already running."
var ErrAlreadyRunning = errors.New("daemon: another instance is already running")

// DefaultLockFile is the file name used by single-instance enforcement
// inside the data dir. The full path is <data-dir>/<DefaultLockFile>.
const DefaultLockFile = "condurad.lock"

// Options configures a single Run() invocation. Build it once, reuse it
// across calls if you need to (it is read-only inside Run).
type Options struct {
	// Config is the fully-validated daemon configuration. Callers are
	// expected to have called cfg.Validate() already.
	Config *config.Config
	// Listen is the IPC binding spec. May be zero-value (Disable=false
	// with empty Addr) to get a random TCP port and a Unix socket.
	Listen ListenSpec
	// Logger is the slog logger to use. If nil, a default is created
	// from the config's logging section and installed as slog default.
	Logger *slog.Logger
	// VersionInfo overrides the build metadata (mostly for tests).
	// If zero, version.Get() is used.
	VersionInfo version.Info
	// Loader is the config loader used to persist config changes
	// back to disk (e.g. onboarding.finish writes ollama.enabled
	// and first_run=false). When nil, config persistence is
	// disabled (test/debug mode).
	Loader *config.Loader
}

// ListenSpec controls where the IPC transport binds.
//
// Addr is a "scheme://host:port" or "scheme:///abs/path" string.
// Examples:
//
//	"tcp://127.0.0.1:0"   — random TCP port on loopback
//	"tcp://127.0.0.1:7666" — fixed TCP port
//	"unix:///path/to/sock" — Unix domain socket (macOS/Linux only)
//
// AuthToken is the bearer token clients must present. Empty disables
// auth (loopback-only is enforced by the config validator).
//
// Disable, if true, suppresses IPC entirely (debugging / smoke tests).
type ListenSpec struct {
	Addr      string
	AuthToken string
	Disable   bool
}

// maybeApplyPendingUpdate completes a staged Windows binary swap on restart.
func maybeApplyPendingUpdate() {
	applied, err := updater.CompletePendingUpdate()
	if err != nil {
		slog.Warn("daemon: pending update", "err", err)
		return
	}
	if applied {
		slog.Info("daemon: applied pending update on restart")
	}
}

// Run starts the daemon and blocks until ctx is canceled or a fatal
// error occurs. On return, all subsystems are torn down (storage
// closed, listeners stopped).
//
// The returned Subsystems bundle is exposed for tests and for the GUI
// (so the Wails App struct can call into it directly). Standalone
// callers can ignore it.
func Run(ctx context.Context, opts Options) (*Subsystems, error) {
	defer crash.Recover()
	if opts.Config == nil {
		return nil, fmt.Errorf("daemon: Config is required")
	}
	if err := opts.Config.Validate(); err != nil {
		return nil, fmt.Errorf("daemon: invalid config: %w", err)
	}
	if err := mkdirDataDir(opts.Config.General.DataDir); err != nil {
		return nil, err
	}
	migrateLegacyDataDir(opts.Config.General.DataDir)
	maybeApplyPendingUpdate()

	// Acquire the single-instance lock before logger/DB/IPC so a
	// second invocation fails fast with a clean error message.
	lock, err := lockfile.TryAcquire(filepath.Join(opts.Config.General.DataDir, DefaultLockFile))
	if err != nil {
		if errors.Is(err, lockfile.ErrLocked) {
			return nil, ErrAlreadyRunning
		}
		return nil, fmt.Errorf("daemon: lockfile: %w", err)
	}
	// From here on, we must release the lock on every error path.
	releaseLock := func() { _ = lock.Release() }

	log := opts.Logger
	if log == nil {
		log = newLoggerFromConfig(opts.Config)
		slog.SetDefault(log)
	}
	ver := opts.VersionInfo
	if ver.Commit == "" {
		ver = version.Get()
	}
	log.Info("starting condurad",
		"version", ver.Version,
		"commit", ver.Commit,
		"build_date", ver.BuildDate,
		"go", ver.GoVersion,
		"platform", ver.Platform,
		"data_dir", opts.Config.General.DataDir,
		"storage_path", opts.Config.Storage.Path,
	)

	subs, err := initSubsystems(log, opts.Config, opts.Loader)
	if err != nil {
		releaseLock()
		return nil, err
	}

	// Mark this machine as installed on first successful start.
	// Subsequent installer runs can detect this via lockfile.IsInstalled().
	_ = lockfile.MarkInstalled()

	ipcSrv := newIPCServer()
	registerMethods(ipcSrv, log, opts.Config, subs, ver)
	subs.Health.Add(healthCheckIPC())

	ipcT := newServerTransport(ipcSrv, opts.Listen.AuthToken)
	ipcT.SSE = subs.Broker
	if !opts.Listen.Disable {
		if err := startListeners(ctx, ipcT, log, opts.Config, opts.Listen); err != nil {
			_ = subs.Storage.Close()
			releaseLock()
			return nil, err
		}
		writeAddrFile(opts.Config, ipcT)
		subs.IPCAddr = ipcT.Addr()
		log.Info("ipc listening", "addr", subs.IPCAddr, "sse_enabled", true)
	}

	startBackgroundServices(ctx, subs, log)

	// Release the lock when ctx is canceled.
	go func() {
		<-ctx.Done()
		log.Info("releasing single-instance lock")
		_ = lock.Release()
	}()

	<-ctx.Done()
	log.Info("condurad stopped")
	shutdownDaemon(subs)
	return subs, nil
}

func startBackgroundServices(ctx context.Context, subs *Subsystems, log *slog.Logger) {
	if subs.BackupScheduler != nil {
		go subs.BackupScheduler.Run(ctx)
		log.Info("auto-backup scheduler started")
	}
	if subs.Updater != nil {
		go subs.Updater.RunPoller(ctx)
		log.Info("auto-update poller started")
	}
	// Anomaly-detector idle reset (Phase 16, Rec 6). Periodically
	// checks whether the detector has been quiet longer than the
	// idle threshold and resets its counters. Prevents stale
	// cross-session noise from triggering false positives.
	if subs.Anomaly != nil {
		go runAnomalyIdleWatcher(ctx, subs.Anomaly, 30*time.Minute, log)
	}
	// Kill-switch Layer 2 (Phase 16, Rec 2). When the watchdog is
	// armed and the user hasn't verified the agent's actions in
	// `timeout`, auto-halt the daemon. Runs in its own goroutine
	// on the daemon's main ctx; the agent cannot disable it.
	if subs.Watchdog != nil {
		go subs.Watchdog.Run(ctx)
		log.Info("kill-switch layer 2 (watchdog) started")
	}
	// Phase 18 (v0.2.0): sub-agent ActionRequest queue.
	// Background sweeper marks stale pending actions as
	// StatusExpired so they don't accumulate forever when the
	// user walks away mid-decision. Runs on the same ctx as
	// the watchdog.
	if subs.Pending != nil {
		subs.Pending.Start(ctx)
		log.Info("pending-actions sweeper started")
	}
	// Audit log retention (§10.5 / B-37). Prunes rows older than
	// the configured retention window once per day. Default 90
	// days; 0 disables pruning (keep forever). The pruner
	// re-chains the oldest surviving row so the log stays
	// tamper-evident after pruning.
	if subs.Audit != nil {
		retention := auditRetention(subs.cfg)
		if retention > 0 {
			go runAuditPruner(ctx, subs.Audit, retention, log)
			log.Info("audit log pruner started", "retention", retention.String())
		} else {
			log.Info("audit log pruner disabled (retention=0)")
		}
	}
}

// runAuditPruner prunes the audit log once per day, deleting rows
// older than retention. Stops when ctx is canceled. The first prune
// runs immediately on startup so a long-running daemon doesn't wait
// a day to clear a backlog.
func runAuditPruner(ctx context.Context, log *audit.Log, retention time.Duration, logger *slog.Logger) {
	prune := func() {
		deleted, err := log.Prune(ctx, retention)
		if err != nil {
			logger.Error("audit prune failed", "error", err)
			return
		}
		if deleted > 0 {
			logger.Info("audit pruned", "deleted", deleted, "retention", retention.String())
		}
	}
	prune()
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			prune()
		}
	}
}

// auditRetention resolves the audit retention duration from config.
// 0 → disabled (keep forever). Negative or unset → 90-day default
// per CLAUDE.md §10.5.
func auditRetention(cfg *config.Config) time.Duration {
	if cfg == nil {
		return 90 * 24 * time.Hour
	}
	if cfg.Security.AuditRetentionDays < 0 {
		return 90 * 24 * time.Hour
	}
	if cfg.Security.AuditRetentionDays == 0 {
		return 0
	}
	return time.Duration(cfg.Security.AuditRetentionDays) * 24 * time.Hour
}

// runAnomalyIdleWatcher polls the anomaly detector every minute and
// calls Reset() once it has been idle longer than `idle`. Stops when
// ctx is canceled.
func runAnomalyIdleWatcher(ctx context.Context, det *anomaly.Detector, idle time.Duration, log *slog.Logger) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if det.IdleReset(idle) {
				det.Reset()
				log.Info("anomaly detector idle-reset", "idle", idle.String())
			}
		}
	}
}

func shutdownDaemon(subs *Subsystems) {
	if subs.BackupScheduler != nil {
		subs.BackupScheduler.Stop()
	}
	if subs.Phase12 != nil && subs.Phase12.SyncEngine != nil {
		subs.Phase12.SyncEngine.Stop()
	}
	if subs.Pending != nil {
		subs.Pending.Stop()
	}
	if subs.Storage != nil {
		_, _ = subs.Storage.SQL().ExecContext(context.Background(), "PRAGMA wal_checkpoint(TRUNCATE)")
	}
	_ = subs.Close()
	_ = subs.Storage.Close()
}

// migrateLegacyDataDir copies the old ~/.condura/ data directory to
// the new ~/.condura/ location if the legacy dir exists and the new
// one is empty. This handles the Synaptic → Condura rename transparently
// so users don't lose their data, API keys, or settings.
func migrateLegacyDataDir(newDir string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	legacyDir := filepath.Join(home, ".synaptic")
	if _, err := os.Stat(legacyDir); os.IsNotExist(err) {
		return
	}
	entries, _ := os.ReadDir(newDir)
	if len(entries) > 1 {
		return
	}
	slog.Info("migrating legacy Synaptic data", "from", legacyDir, "to", newDir)
	copyDir(legacyDir, newDir)
	slog.Info("data migration complete — you can safely delete ~/.condura/")
}

func copyDir(src, dst string) {
	_ = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil //nolint:nilerr // skip unreadable files
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			_ = os.MkdirAll(target, 0o700)
			return nil
		}
		data, err := os.ReadFile(path) //nolint:gosec
		if err != nil {
			return nil //nolint:nilerr // skip unreadable files
		}
		_ = os.WriteFile(target, data, info.Mode()) //nolint:gosec // trusted internal path
		return nil
	})
}
