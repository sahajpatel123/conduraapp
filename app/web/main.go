// Command web is the Synaptic GUI entry point.
//
// This is a Wails v2 app: a single binary that embeds the system
// WebView and the synaptic daemon in the same process. The daemon
// is started in a goroutine via internal/daemon.Run(); the Wails
// App struct exposes IPC methods to the frontend.
package main

import (
	"context"
	"embed"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sahajpatel123/synapticapp/internal/config"
	"github.com/sahajpatel123/synapticapp/internal/daemon"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// embeddedDaemon lets the Wails App struct talk to the in-process
// daemon goroutine. Set in main() before wails.Run.
var embeddedDaemon *daemon.Subsystems

// daemonReady is closed by the daemon goroutine when IPC is bound and
// the addr-sidecar is written. The GUI waits on this to show a
// "connected" indicator.
var daemonReady = make(chan struct{})

func main() {
	cfg, err := resolveConfig()
	if err != nil {
		println("synaptic-gui: config:", err.Error())
		os.Exit(1)
	}

	// Run the daemon in a goroutine. ctx is canceled on SIGINT/SIGTERM,
	// which triggers a graceful shutdown inside daemon.Run.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		subs, err := daemon.Run(ctx, daemon.Options{
			Config: cfg,
			Listen: daemon.ListenSpec{
				Addr:      "tcp://127.0.0.1:0",
				AuthToken: cfg.APIServer.AuthToken,
			},
			Logger: slog.Default(),
		})
		if err != nil {
			println("synaptic-gui: daemon:", err.Error())
			cancel()
			return
		}
		embeddedDaemon = subs
		close(daemonReady)
	}()

	// Start the Wails app. The Wails runtime takes over the main
	// goroutine; the daemon runs in its own goroutine above.
	app := NewApp()
	err = wails.Run(&options.App{
		Title:  "Synaptic",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 18, G: 18, B: 22, A: 1},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("synaptic-gui:", err.Error())
		os.Exit(1)
	}
}

// resolveConfig loads the user's config from the same path the
// standalone daemon uses. The GUI does not parse CLI flags; flags
// for daemon-only options (--listen, --no-ipc) are ignored because
// the GUI controls its own listen address.
func resolveConfig() (*config.Config, error) {
	loader := config.NewLoader("")
	cfg, err := loader.Load()
	if err != nil {
		return nil, err
	}
	if sp, err := cfg.ResolveStoragePath(); err == nil {
		cfg.Storage.Path = sp
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}
