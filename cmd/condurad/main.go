// Command condurad is the Condura daemon.
//
// It owns the on-disk database, the OS keyring, and the LLM registry.
// Clients (the Wails GUI, the CLI, the future Skills Hub) talk to it
// over JSON-RPC 2.0 on a local TCP or Unix socket.
//
// Typical lifecycle:
//
//	synapticd --config ~/.condura/config.yaml
//	synapticd --print-default-config > ~/.condura/config.yaml
//
// Phase 2+: this binary is the standalone CLI daemon. The GUI
// (cmd/synaptic-gui) embeds the same daemon logic via
// internal/daemon.Run() — see that package for the actual entry
// point.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"gopkg.in/yaml.v3"

	"github.com/sahajpatel123/conduraapp/internal/config"
	"github.com/sahajpatel123/conduraapp/internal/crash"
	"github.com/sahajpatel123/conduraapp/internal/daemon"
	"github.com/sahajpatel123/conduraapp/internal/version"
)

const addrFilePerm = 0o600

func main() {
	defer crash.Recover() //nolint:gocritic // Recover catches panics; os.Exit is the normal error path.
	flags, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "condurad: %v\n", err)
		os.Exit(1) //nolint:gocritic // intentional: pre-daemon exit; no resources to clean up
	}
	if flags.quit {
		return
	}

	cfg, err := buildConfig(flags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "condurad: %v\n", err)
		os.Exit(1)
	}
	loader, err := buildLoader(flags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "condurad: %v\n", err)
		os.Exit(1)
	}

	// Translate SIGINT/SIGTERM to ctx cancellation so daemon.Run
	// can shut down gracefully. We call cancel() explicitly before
	// os.Exit because deferred functions don't run after os.Exit.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	opts := daemon.Options{
		Config: cfg,
		Loader: loader,
		Listen: daemon.ListenSpec{
			Addr:      flags.listen,
			AuthToken: cfg.APIServer.AuthToken,
			Disable:   flags.noIPC,
		},
	}
	if _, err := daemon.Run(ctx, opts); err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "condurad: %v\n", err)
		os.Exit(1)
	}
	cancel()
	_ = addrFilePerm // kept for symmetry with the GUI binary
}

// runFlags is the parsed CLI flag set. quit is true when an early-exit
// flag (--version, --print-default-config) was handled and main()
// should return without doing any work.
type runFlags struct {
	cfgPath  string
	dataDir  string
	logLevel string
	listen   string
	noIPC    bool
	quit     bool
}

func parseFlags() (runFlags, error) {
	var (
		cfgPath       = flag.String("config", "", "path to config.yaml (default: ~/.condura/config.yaml)")
		dataDir       = flag.String("data-dir", "", "data directory (overrides config)")
		logLevel      = flag.String("log-level", "", "debug | info | warn | error (default: config value)")
		listen        = flag.String("listen", "tcp://127.0.0.1:0", "IPC listen address")
		noIPC         = flag.Bool("no-ipc", false, "disable IPC server (debugging)")
		printDefaults = flag.Bool("print-default-config", false, "print default config YAML to stdout and exit")
		printVersion  = flag.Bool("version", false, "print version and exit")
	)
	flag.Parse()

	if *printVersion {
		fmt.Println(version.String())
		return runFlags{quit: true}, nil
	}
	if *printDefaults {
		out, err := yaml.Marshal(config.Default())
		if err != nil {
			return runFlags{}, fmt.Errorf("marshal default config: %w", err)
		}
		fmt.Print(string(out))
		return runFlags{quit: true}, nil
	}
	return runFlags{
		cfgPath:  *cfgPath,
		dataDir:  *dataDir,
		logLevel: *logLevel,
		listen:   *listen,
		noIPC:    *noIPC,
	}, nil
}

// buildLoader returns a config.Loader for the given flags.
func buildLoader(flags runFlags) (*config.Loader, error) {
	cfgPath := flags.cfgPath
	if cfgPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("locate home dir: %w", err)
		}
		cfgPath = filepath.Join(home, ".condura", "config.yaml")
	}
	return config.NewLoader(cfgPath), nil
}

// buildConfig resolves the config path, loads the YAML, applies any
// CLI overrides (--data-dir), and returns a fully-validated config.
func buildConfig(flags runFlags) (*config.Config, error) {
	cfgPath := flags.cfgPath
	if cfgPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("locate home dir: %w", err)
		}
		cfgPath = filepath.Join(home, ".condura", "config.yaml")
	}
	loader := config.NewLoader(cfgPath)
	cfg, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if flags.dataDir != "" {
		cfg.OverrideDataDir(flags.dataDir)
	}
	if sp, err := cfg.ResolveStoragePath(); err == nil {
		cfg.Storage.Path = sp
	}
	// daemon.Run() will also call Validate(), but doing it here too
	// gives a faster failure on bad config and a clear error.
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}
