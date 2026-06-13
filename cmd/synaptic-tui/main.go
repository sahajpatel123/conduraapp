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
	addr := flag.String("addr", "", "daemon IPC address (e.g. unix:///path/to/sock, tcp://127.0.0.1:PORT)")
	flag.Parse()

	daemonAddr := *addr
	if daemonAddr == "" {
		daemonAddr = tui.FindDaemonAddr()
	}
	if daemonAddr == "" {
		fmt.Fprintln(os.Stderr, "synaptic-tui: cannot find daemon. Is synapticd running?")
		fmt.Fprintln(os.Stderr, "  Pass --addr or start synapticd first.")
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	client, err := tui.NewIPCClient(daemonAddr, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "synaptic-tui: connect to daemon: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	p := tea.NewProgram(tui.InitialModel(client, logger), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "synaptic-tui: %v\n", err)
		os.Exit(1)
	}
}
