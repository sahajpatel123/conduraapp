// Package daemon is the in-process entry point for the Synaptic daemon.
// It is the same code path used by:
//
//   - cmd/synapticd      — the standalone CLI daemon (Phase 1)
//   - cmd/synaptic-gui   — the Wails-based desktop GUI (Phase 2)
//
// Run() is the single public entry point. It is safe to call from a
// goroutine; the caller is expected to cancel ctx to request a graceful
// shutdown.
//
// Concurrency: Run is NOT safe to call twice in the same process. Most
// subsystems (secrets manager, storage DB, log default) are singletons
// and will conflict.
package daemon

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sahajpatel123/synapticapp/internal/config"
	"github.com/sahajpatel123/synapticapp/internal/version"
)

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

// Run starts the daemon and blocks until ctx is canceled or a fatal
// error occurs. On return, all subsystems are torn down (storage
// closed, listeners stopped).
//
// The returned Subsystems bundle is exposed for tests and for the GUI
// (so the Wails App struct can call into it directly). Standalone
// callers can ignore it.
func Run(ctx context.Context, opts Options) (*Subsystems, error) {
	if opts.Config == nil {
		return nil, fmt.Errorf("daemon: Config is required")
	}
	if err := opts.Config.Validate(); err != nil {
		return nil, fmt.Errorf("daemon: invalid config: %w", err)
	}
	if err := mkdirDataDir(opts.Config.General.DataDir); err != nil {
		return nil, err
	}

	log := opts.Logger
	if log == nil {
		log = newLoggerFromConfig(opts.Config)
		slog.SetDefault(log)
	}
	ver := opts.VersionInfo
	if ver.Commit == "" {
		ver = version.Get()
	}
	log.Info("starting synapticd",
		"version", ver.Version,
		"commit", ver.Commit,
		"build_date", ver.BuildDate,
		"go", ver.GoVersion,
		"platform", ver.Platform,
		"data_dir", opts.Config.General.DataDir,
		"storage_path", opts.Config.Storage.Path,
	)

	subs, err := initSubsystems(log, opts.Config)
	if err != nil {
		return nil, err
	}

	ipcSrv := newIPCServer()
	registerMethods(ipcSrv, log, opts.Config, subs, ver)
	subs.Health.Add(healthCheckIPC())

	ipcT := newServerTransport(ipcSrv, opts.Listen.AuthToken)
	if !opts.Listen.Disable {
		if err := startListeners(ctx, ipcT, log, opts.Config, opts.Listen); err != nil {
			_ = subs.Storage.Close()
			return nil, err
		}
		writeAddrFile(opts.Config, ipcT)
		subs.IPCAddr = ipcT.Addr()
		log.Info("ipc listening", "addr", subs.IPCAddr)
	}

	<-ctx.Done()
	log.Info("synapticd stopped")

	// Best-effort teardown.
	_ = subs.Storage.Close()
	return subs, nil
}
