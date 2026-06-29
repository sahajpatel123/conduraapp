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
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/sahajpatel123/conduraapp/internal/config"
	"github.com/sahajpatel123/conduraapp/internal/daemon"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsmac "github.com/wailsapp/wails/v2/pkg/options/mac"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
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

// appInstance is the Wails App struct, set before wails.Run so the
// daemon goroutine can access it for overlay/tray wiring.
var appInstance *App

// pendingOAuthCallback stores a URL received via OnUrlOpen before
// the daemon is ready. Once the daemon is up, the goroutine processes
// any pending callbacks.
var (
	pendingOAuthMu sync.Mutex
	pendingOAuth   []string
)

func main() {
	cfg, loader, err := resolveConfig()
	if err != nil {
		println("condura-gui: config:", err.Error())
		os.Exit(1)
	}

	// Run the daemon in a goroutine. ctx is canceled on SIGINT/SIGTERM
	// OR when the tray's Quit menu item is picked (via
	// appInstance.quitCancel). Either path triggers a graceful
	// shutdown inside daemon.Run.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	appInstance = NewApp()
	appInstance.quitCancel = cancel

	go func() {
		subs, err := daemon.Run(ctx, daemon.Options{
			Config: cfg,
			Loader: loader,
			Listen: daemon.ListenSpec{
				Addr:      "tcp://127.0.0.1:0",
				AuthToken: cfg.APIServer.AuthToken,
			},
			Logger: slog.Default(),
		})
		if err != nil {
			println("condura-gui: daemon:", err.Error())
			cancel()
			return
		}
		embeddedDaemon = subs
		close(daemonReady)

		// Process any OAuth callbacks that arrived before the
		// daemon was ready.
		processPendingOAuth()

		// Swap the daemon's noop overlay controller for the
		// real Wails-backed one. The conductor's onShow/onHide
		// callbacks and any daemon-side overlay RPC (overlay.show
		// etc.) now drive the actual Wails window instead of a
		// headless state machine. The Wails runtime context is
		// wired in App.startup (it doesn't exist yet here);
		// the controller guards on a nil context until then.
		subs.SetOverlay(appInstance.overlayCtrl)

		// Start the system tray (menu-bar icon on macOS, task
		// tray on Windows). After this point the user has both
		// the GUI window and a tray icon to control the app.
		appInstance.startTray(ctx, subs)

		// Wire the conductor (hotkey → overlay toggle) once the
		// daemon is ready. The conductor's onShow/onHide callbacks
		// route through the Wails window methods so the overlay
		// is a real frameless/always-on-top mode, not a noop.
		appInstance.startConductor(subs, resolveOverlayHotkey(cfg.Hotkey.Overlay))
	}()

	// Start the Wails app. The Wails runtime takes over the main
	// goroutine; the daemon runs in its own goroutine above.
	err = wails.Run(&options.App{
		Title:     "Condura",
		Width:     1200,
		Height:    800,
		MinWidth:  800,
		MinHeight: 500,
		Menu:      buildApplicationMenu(appInstance),
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour:         &options.RGBA{R: 18, G: 18, B: 22, A: 1},
		OnStartup:                appInstance.startup,
		OnDomReady:               appInstance.domReady,
		OnBeforeClose:            appInstance.beforeClose,
		EnableDefaultContextMenu: true,
		Mac: &wailsmac.Options{
			OnUrlOpen: handleOpenURL,
		},
		Bind: []interface{}{
			appInstance,
		},
	})
	if err != nil {
		println("condura-gui:", err.Error())
		os.Exit(1)
	}
}

// handleOpenURL is called by the OS when a condura:// URL is
// opened. On macOS this fires via the OnUrlOpen callback. We queue
// the URL for processing once the daemon is ready.
func handleOpenURL(rawURL string) {
	if rawURL == "" || !strings.HasPrefix(rawURL, "condura://") {
		return
	}
	pendingOAuthMu.Lock()
	pendingOAuth = append(pendingOAuth, rawURL)
	pendingOAuthMu.Unlock()
}

// processPendingOAuth drains queued OAuth callback URLs once the
// daemon is running. Called from the daemon goroutine after the
// daemon is fully initialized.
func processPendingOAuth() {
	pendingOAuthMu.Lock()
	urls := pendingOAuth
	pendingOAuth = nil
	pendingOAuthMu.Unlock()

	for _, rawURL := range urls {
		processOAuthCallback(rawURL)
	}
}

// processOAuthCallback parses a condura://auth/callback URL and
// calls the daemon's account.oauth_callback RPC. On success, it
// emits a frontend event so the UI can refresh its signed-in state.
func processOAuthCallback(rawURL string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		slog.Warn("condura-gui: bad oauth callback url", "url", rawURL, "err", err)
		return
	}
	if !strings.HasPrefix(u.Path, "auth/callback") {
		return // not an OAuth callback
	}
	code := u.Query().Get("code")
	state := u.Query().Get("state")
	if code == "" || state == "" {
		slog.Warn("condura-gui: oauth callback missing code or state", "url", rawURL)
		return
	}
	if embeddedDaemon == nil || embeddedDaemon.Account == nil {
		slog.Warn("condura-gui: oauth callback received but daemon/account not ready")
		return
	}
	ctx := context.Background()
	sess, err := embeddedDaemon.Account.ExchangeCode(ctx, "google", code, state, "condura://auth/callback")
	if err != nil {
		// Try GitHub as fallback.
		sess, err = embeddedDaemon.Account.ExchangeCode(ctx, "github", code, state, "condura://auth/callback")
	}
	if err != nil || sess == nil {
		slog.Error("condura-gui: oauth token exchange failed", "err", err)
		return
	}
	slog.Info("condura-gui: oauth callback processed", "email", sess.Email, "provider", sess.Provider)
	// Emit event to frontend so the UI can refresh signed-in state.
	if appInstance != nil && appInstance.ctx != nil {
		wailsruntime.EventsEmit(appInstance.ctx, "condura:oauth-callback",
			map[string]interface{}{
				"signed_in": true,
				"email":     sess.Email,
				"provider":  sess.Provider,
			})
	}
}

// resolveConfig loads the user's config from the same path the
// standalone daemon uses. The GUI does not parse CLI flags; flags
// for daemon-only options (--listen, --no-ipc) are ignored because
// the GUI controls its own listen address.
func resolveConfig() (*config.Config, *config.Loader, error) {
	loader := config.NewLoader("")
	cfg, err := loader.Load()
	if err != nil {
		return nil, nil, err
	}
	if sp, err := cfg.ResolveStoragePath(); err == nil {
		cfg.Storage.Path = sp
	}
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}
	return cfg, loader, nil
}
