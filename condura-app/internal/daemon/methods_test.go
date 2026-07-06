package daemon

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/config"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/ipc"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/version"
)

// newInfoRPCServer wires registerMethods (and therefore
// daemon.uptime / daemon.pid / daemon.info) into a fresh ipc.Server
// against an empty Subsystems. The three RPCs under test don't
// dereference any subsystem, so the empty bundle is safe; using a
// discarding logger keeps test output quiet if the methods log.
func newInfoRPCServer(t *testing.T) *ipc.Server {
	t.Helper()
	srv := ipc.NewServer()
	discardLog := slog.New(slog.NewTextHandler(io.Discard, nil))
	registerMethods(srv, discardLog, &config.Config{}, &Subsystems{}, version.Info{
		Version: "test-0.0.0",
		Commit:  "deadbeef",
	})
	return srv
}

func callInfoRPC(t *testing.T, srv *ipc.Server, method string) any {
	t.Helper()
	resp, err := srv.Handle(context.Background(), &ipc.Request{
		JSONRPC: "2.0",
		Method:  method,
		ID:      json.RawMessage("1"),
	})
	if err != nil {
		t.Fatalf("%s: dispatch error: %v", method, err)
	}
	if resp.Error != nil {
		t.Fatalf("%s: returned error: %v", method, resp.Error)
	}
	if resp.Result == nil {
		t.Fatalf("%s: returned nil result", method)
	}
	return resp.Result
}

func TestDaemonUptimeRPC_ReportsSecondsSinceMarkedReady(t *testing.T) {
	// MarkDaemonStart publishes a known sentinel; wait a measurable
	// interval; uptime must be >= the wait and < a generous upper
	// bound. This pins the contract: uptime measures "time since
	// MarkDaemonStart", not "time since binary load".
	MarkDaemonStart(time.Now())
	time.Sleep(25 * time.Millisecond)
	srv := newInfoRPCServer(t)

	m := callInfoRPC(t, srv, "daemon.uptime").(map[string]any)
	up, ok := m["uptime_seconds"].(float64)
	if !ok {
		t.Fatalf("uptime_seconds not float64: %T", m["uptime_seconds"])
	}
	if up < 0.025 {
		t.Fatalf("uptime_seconds=%v, expected >= 0.025 (after a 25ms wait)", up)
	}
	if up > 5 {
		t.Fatalf("uptime_seconds=%v, expected < 5s (no test waits that long)", up)
	}
}

func TestDaemonUptimeRPC_BeforeMark_ReturnsNonNegative(t *testing.T) {
	// When MarkDaemonStart has not been called by this process — say,
	// a test that runs before any other has touched it — the
	// atomic.Pointer sentinel is still safe to read. The value
	// will be the package-init time, but it must not panic and
	// must be non-negative.
	srv := newInfoRPCServer(t)

	m := callInfoRPC(t, srv, "daemon.uptime").(map[string]any)
	up, ok := m["uptime_seconds"].(float64)
	if !ok {
		t.Fatalf("uptime_seconds not float64: %T", m["uptime_seconds"])
	}
	if up < 0 {
		t.Fatalf("uptime_seconds negative: %v", up)
	}
}

// asInt extracts an integer-shaped JSON value (int / int64 /
// float64) into a Go int. encoding/json may return any of those
// depending on the source value's magnitude.
func asInt(t *testing.T, v any) int {
	t.Helper()
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	default:
		t.Fatalf("expected integer-shaped value, got %T (%v)", v, v)
		return 0
	}
}

func TestDaemonPidRPC_EqualsOSGetpid(t *testing.T) {
	srv := newInfoRPCServer(t)
	m := callInfoRPC(t, srv, "daemon.pid").(map[string]any)
	got := asInt(t, m["pid"])
	if got != os.Getpid() {
		t.Fatalf("pid: got %d, want %d (os.Getpid)", got, os.Getpid())
	}
}

func TestDaemonInfoRPC_IncludesUptimeAndPidAndVersion(t *testing.T) {
	MarkDaemonStart(time.Now())
	srv := newInfoRPCServer(t)

	m := callInfoRPC(t, srv, "daemon.info").(map[string]any)
	if _, ok := m["uptime_seconds"]; !ok {
		t.Error("missing uptime_seconds in daemon.info")
	}
	if _, ok := m["pid"]; !ok {
		t.Error("missing pid in daemon.info")
	}
	// version is whatever the closure captured — accept either the
	// raw version.Info struct (IPC returns the Go value, not a
	// JSON roundtrip) or a map from a JSON-deserializing client.
	switch m["version"].(type) {
	case version.Info, map[string]any:
		// ok
	default:
		t.Fatalf("version shape wrong: %T", m["version"])
	}
}
