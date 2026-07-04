// Package health provides cheap, unauthenticated probe endpoints for
// orchestrators (systemd, k8s, launchd) that need to poll the daemon's
// liveness and readiness without a bearer token.
//
// These endpoints are NOT part of the JSON-RPC surface. They are
// mounted on the same listener as the IPC server only when the
// caller opts in via ListenSpec.Health. By default they are OFF
// for non-loopback bindings (a public health endpoint that leaks
// process state to the internet is a footgun).
//
// Security: this file never reads the Authorization header, never
// logs the request body, and never writes anything from the
// readiness func into the response beyond the error message. The
// readyz func is the caller's responsibility to keep boring.
package health

import (
	"errors"
	"net/http"
)

// HTTPHandler returns a mux that serves /livez and /readyz with NO
// authentication. It is intended to be mounted as a sidecar on an
// existing HTTP listener (see internal/ipc.Transport). Both
// endpoints respond in O(1) — no DB, no I/O, no header inspection.
//
// livez: process is up. Always 200 with body "alive\n" as long as
// the goroutine that serves requests is running. The k8s liveness
// probe semantic — a 200 here means "do not restart me."
//
// readyz: process can serve. 200 if readyz returns nil; 503 with
// "not ready: <reason>" otherwise. The k8s readiness semantic — a
// 200 here means "send me traffic."
//
// The readyz func is invoked on every request. Keep it cheap
// (e.g., a sync/atomic.Bool read) and side-effect free. It is
// always called from a request goroutine and may run concurrently
// with itself.
//
// The livez func is currently ignored — /livez is a constant 200
// because the process responding to HTTP at all is the signal. The
// parameter is kept for symmetry and to allow future hooks (e.g.,
// "live iff memory < 90%").
func HTTPHandler(livez, readyz func() error) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		_ = livez // signature kept for symmetry; semantics are constant
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("alive\n"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		if err := readyz(); err != nil {
			// Cap the reason at 256 bytes so a runaway readyz func
			// cannot flood the response or leak arbitrary bytes.
			reason := err.Error()
			if len(reason) > 256 {
				reason = reason[:256]
			}
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("not ready: " + reason + "\n"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready\n"))
	})
	return mux
}

// ErrNotReady is a convenience for callers who want to construct a
// readyz func that always returns a fixed reason. The message is
// exposed to the orchestrator on every probe — keep it boring and
// do not include secrets.
var ErrNotReady = errors.New("not ready")
