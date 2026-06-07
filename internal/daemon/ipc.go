package daemon

import (
	"runtime"

	"github.com/sahajpatel123/synapticapp/internal/ipc"
)

// isWindows is a small wrapper so the listeners code can use the
// same string check everywhere instead of duplicating runtime.GOOS.
var isWindows = runtime.GOOS == "windows"

// newIPCServer returns a fresh JSON-RPC server. Pulled out so tests
// can substitute a mock.
func newIPCServer() *ipc.Server {
	return ipc.NewServer()
}

// newServerTransport returns a transport bound to the given server
// and auth token. Wraps ipc.ServerTransport in a small abstraction
// so the daemon package doesn't leak the transport type into the
// public API.
func newServerTransport(srv *ipc.Server, token string) *ipc.ServerTransport {
	return &ipc.ServerTransport{S: srv, Token: token}
}
