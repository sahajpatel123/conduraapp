package daemon

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/sahajpatel123/conduraapp/internal/config"
	"github.com/sahajpatel123/conduraapp/internal/ipc"
)

// startListeners binds the IPC transport. The "scheme" portion of
// opts.Listen.Addr picks the network:
//
//	tcp   → net.Listen("tcp", host:port)
//	unix  → net.Listen("unix", /abs/path)
//
// On macOS / Linux we always also bind a Unix socket in the data
// directory for fast local access. The TCP listener is the "primary"
// address and is what we publish via the condurad.addr sidecar.
func startListeners(ctx context.Context, ipcT *ipc.ServerTransport, log *slog.Logger, cfg *config.Config, opts ListenSpec) error {
	addr := opts.Addr
	if addr == "" {
		addr = "tcp://127.0.0.1:0"
	}
	if err := ipcT.Listen(ctx, addr); err != nil {
		return err
	}
	log.Info("ipc listening", "addr", ipcT.Addr(), "scheme", schemeOf(addr))

	// Best-effort: also bind a Unix socket in the data dir, on
	// platforms that support it. Failures are warnings, not errors.
	bindUnixSocket(ctx, ipcT, log, cfg)
	return nil
}

// bindUnixSocket binds a Unix domain socket inside the data dir, on
// macOS / Linux. The path is <data_dir>/synapticd.sock. We
// best-effort remove any stale socket file first.
func bindUnixSocket(ctx context.Context, ipcT *ipc.ServerTransport, log *slog.Logger, cfg *config.Config) {
	if isWindows {
		return
	}
	unixPath := filepath.Join(cfg.General.DataDir, "condurad.sock")
	_ = os.Remove(unixPath)
	if err := ipcT.Listen(ctx, "unix://"+unixPath); err != nil {
		log.Warn("unix socket bind failed; continuing", "err", err)
		return
	}
	log.Info("ipc unix socket ready", "path", unixPath)
}

// writeAddrFile writes the primary listen address to a sidecar file
// in the data directory. The CLI reads this to find the daemon
// without scanning a port range.
func writeAddrFile(cfg *config.Config, ipcT *ipc.ServerTransport) {
	if ipcT.Addr() == "" {
		return
	}
	path := filepath.Join(cfg.General.DataDir, "condurad.addr")
	_ = os.WriteFile(path, []byte(ipcT.Addr()+"\n"), addrFilePerm)
}

// schemeOf returns the URL scheme of an "scheme://host:port" string.
// Returns "tcp" if no scheme is present.
func schemeOf(addr string) string {
	for i := 0; i < len(addr); i++ {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return "tcp"
}
