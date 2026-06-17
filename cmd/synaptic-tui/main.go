// Command condura-tui is the terminal UI for Condura.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sahajpatel123/synapticapp/internal/tui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	addr := flag.String("addr", "", "daemon IPC address (e.g. unix:///path/to/sock, tcp://127.0.0.1:PORT)")
	flag.Parse()

	daemonAddr := *addr
	if daemonAddr == "" {
		daemonAddr = tui.FindDaemonAddr()
	}
	if daemonAddr == "" {
		return fmt.Errorf("condura-tui: cannot find daemon — is condurad running?\n  Pass --addr or start condurad first")
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	client, err := tui.NewIPCClient(daemonAddr, logger)
	if err != nil {
		return fmt.Errorf("condura-tui: connect to daemon: %w", err)
	}
	defer func() { _ = client.Close() }()

	p := tea.NewProgram(tui.InitialModel(client, logger), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("condura-tui: %w", err)
	}
	return nil
}
