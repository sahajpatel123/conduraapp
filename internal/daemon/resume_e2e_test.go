package daemon

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/halt"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
)

// fakeGuard is a no-op NetworkGuard that records Halt/Resume for the
// IPC-level halt/confirm_resume test below.
type fakeResumeGuard struct {
	haltCalls   int
	resumeCalls int
}

func (g *fakeResumeGuard) Allow(string) bool { return true }
func (g *fakeResumeGuard) WrapTransport(rt http.RoundTripper) http.RoundTripper {
	return rt
}
func (g *fakeResumeGuard) Halt(string) error { g.haltCalls++; return nil }
func (g *fakeResumeGuard) Resume() error     { g.resumeCalls++; return nil }
func (g *fakeResumeGuard) State() halt.GuardState {
	return halt.GuardState{Halted: g.haltCalls > g.resumeCalls}
}

// fakeIPCHaltBus spins up the real halt RPC surface (daemon.halt,
// daemon.resume_request, halt.confirm_resume, halt.state) on an in-
// memory IPC server + client pair and returns a wired daemon that
// can be Halted/Resumed via the IPC contract.
type fakeIPCHaltBus struct {
	srv       *ipc.Server
	client    *ipc.Client
	haltFlag  *halt.Flag
	guard     *fakeResumeGuard
	tickets   *ResumeTicketStore
	secret    *ResumeSecretManager
	secretVal string
}

func newFakeIPCHaltBus(t *testing.T) *fakeIPCHaltBus {
	t.Helper()
	db := testDB(t)
	haltFlag := halt.New(db)
	if err := haltFlag.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	guard := &fakeResumeGuard{}
	tickets := NewResumeTicketStore()
	// Pre-load the secret (no env, no file: auto-generate in a temp dir).
	secretDir := t.TempDir()
	t.Setenv("CONDURA_RESUME_SECRET", "")
	secret := NewResumeSecretManager(secretDir, "CONDURA_RESUME_SECRET")
	secretVal, err := secret.Load()
	if err != nil {
		t.Fatalf("secret.Load: %v", err)
	}
	auditLog := audit.New(db, []byte("test-hmac-secret-32-bytes-pad")) //nolint:gosec // test secret; not security-sensitive
	srv := ipc.NewServer()
	registerHaltMethods(srv, haltFlag, auditLog, nil, guard, tickets, secret)

	// Construct the transport directly (matches internal/daemon/ipc.go's
	// newServerTransport pattern). Listen on a random localhost TCP port.
	transport := &ipc.ServerTransport{S: srv, Token: ""}
	if err := transport.Listen(context.Background(), "tcp://127.0.0.1:0"); err != nil {
		t.Fatalf("Listen: %v", err)
	}
	t.Cleanup(func() { _ = transport.Close() })

	// Wait for the listener to be bound (Addr returns "" until Listen
	// completes; the call above is synchronous but reads race with the
	// goroutine that records addr, so retry briefly).
	var addr string
	for i := 0; i < 50; i++ {
		addr = transport.Addr()
		if addr != "" {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	if addr == "" {
		t.Fatal("transport Addr empty after Listen")
	}
	client, err := ipc.Dial("tcp://"+addr, "")
	if err != nil {
		t.Fatalf("ipc.Dial: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return &fakeIPCHaltBus{
		srv:       srv,
		client:    client,
		haltFlag:  haltFlag,
		guard:     guard,
		tickets:   tickets,
		secret:    secret,
		secretVal: secretVal,
	}
}

func (b *fakeIPCHaltBus) call(ctx context.Context, method string, params any, out any) error {
	// ipc.Client.Call uses HandleRaw → uses serverTransport which is
	// tied to the srv. Call the server-side HandleRaw directly to
	// avoid the http layer (which would need a JSON-RPC envelope).
	// However the *server* and *client* types are coupled via the
	// same transport, so we just call HandleRaw.
	return b.client.Call(ctx, method, params, out)
}

// TestResumeIPC_HappyPath: daemon.halt → daemon.resume_request →
// halt.confirm_resume with the right secret un-halts and the guard
// is resumed (Layer 3).
func TestResumeIPC_HappyPath(t *testing.T) {
	bus := newFakeIPCHaltBus(t)
	ctx := context.Background()

	// 1. Halt the daemon (via the IPC contract — not directly).
	var haltRes map[string]any
	if err := bus.call(ctx, "daemon.halt", map[string]any{"reason": "test"}, &haltRes); err != nil {
		t.Fatalf("daemon.halt: %v", err)
	}
	if h, _ := haltRes["halted"].(bool); !h {
		t.Fatalf("daemon.halt halted = %v, want true", haltRes["halted"])
	}
	if !bus.haltFlag.IsHalted() {
		t.Fatal("flag should be halted after daemon.halt")
	}
	if bus.guard.haltCalls != 1 {
		t.Fatalf("guard.Halt called %d times, want 1", bus.guard.haltCalls)
	}

	// 2. Request a resume ticket.
	var reqRes map[string]any
	if err := bus.call(ctx, "daemon.resume_request", nil, &reqRes); err != nil {
		t.Fatalf("daemon.resume_request: %v", err)
	}
	ticket, _ := reqRes["ticket"].(string)
	if len(ticket) != 64 {
		t.Fatalf("ticket length = %d, want 64", len(ticket))
	}

	// 3. Confirm with the right secret.
	var confRes map[string]any
	if err := bus.call(ctx, "halt.confirm_resume", map[string]any{
		"ticket": ticket,
		"secret": bus.secretVal,
	}, &confRes); err != nil {
		t.Fatalf("halt.confirm_resume: %v", err)
	}
	if r, _ := confRes["resumed"].(bool); !r {
		t.Fatalf("halt.confirm_resume resumed = %v, want true", confRes["resumed"])
	}
	if bus.haltFlag.IsHalted() {
		t.Fatal("flag should NOT be halted after confirm_resume")
	}
	if bus.guard.resumeCalls != 1 {
		t.Fatalf("guard.Resume called %d times, want 1", bus.guard.resumeCalls)
	}
}

// TestResumeIPC_DeniesBadSecret: halt.confirm_resume with a wrong
// secret denies, ticket remains valid for retry.
func TestResumeIPC_DeniesBadSecret(t *testing.T) {
	bus := newFakeIPCHaltBus(t)
	ctx := context.Background()
	_ = bus.call(ctx, "daemon.halt", map[string]any{"reason": "test"}, nil)

	var reqRes map[string]any
	if err := bus.call(ctx, "daemon.resume_request", nil, &reqRes); err != nil {
		t.Fatalf("daemon.resume_request: %v", err)
	}
	ticket, _ := reqRes["ticket"].(string)

	// Wrong secret — must error and leave the flag halted.
	var badRes map[string]any
	err := bus.call(ctx, "halt.confirm_resume", map[string]any{
		"ticket": ticket,
		"secret": "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef", // 64 chars hex, wrong value
	}, &badRes)
	if err == nil {
		t.Fatal("halt.confirm_resume with wrong secret must error")
	}
	if !bus.haltFlag.IsHalted() {
		t.Fatal("flag must remain halted after bad confirm")
	}
	if bus.guard.resumeCalls != 0 {
		t.Fatal("guard.Resume must NOT be called after bad confirm")
	}

	// Right secret still works (ticket NOT consumed by the bad attempt).
	var confRes map[string]any
	if err := bus.call(ctx, "halt.confirm_resume", map[string]any{
		"ticket": ticket,
		"secret": bus.secretVal,
	}, &confRes); err != nil {
		t.Fatalf("good retry after bad attempt: %v", err)
	}
	if bus.haltFlag.IsHalted() {
		t.Fatal("flag should be un-halted after good retry")
	}
}

// TestResumeIPC_DeprecationShim: the old daemon.resume RPC returns a
// clear migration error.
func TestResumeIPC_DeprecationShim(t *testing.T) {
	bus := newFakeIPCHaltBus(t)
	ctx := context.Background()
	var out map[string]any
	err := bus.call(ctx, "daemon.resume", nil, &out)
	if err == nil {
		t.Fatal("daemon.resume (deprecated) must return an error pointing at the new flow")
	}
	msg := err.Error()
	if !strings.Contains(msg, "daemon.resume_request") || !strings.Contains(msg, "condura resume --confirm") {
		t.Fatalf("deprecation error must mention both new flow names; got %q", msg)
	}
	// The deprecated path must NOT flip halted state (bus starts not-halted).
	if bus.haltFlag.IsHalted() {
		t.Fatal("deprecation shim must not flip halted flag state")
	}
}

// TestResumeIPC_NotHaltedNoTicket: daemon.resume_request returns
// halted=false + ticket="" when the daemon is not halted (the GUI
// can use this to show "already running" without minting a ticket).
func TestResumeIPC_NotHaltedNoTicket(t *testing.T) {
	bus := newFakeIPCHaltBus(t)
	ctx := context.Background()
	var res map[string]any
	if err := bus.call(ctx, "daemon.resume_request", nil, &res); err != nil {
		t.Fatalf("daemon.resume_request (not halted): %v", err)
	}
	if h, _ := res["halted"].(bool); h {
		t.Fatalf("halted = true, want false (daemon not halted)")
	}
	if tk, _ := res["ticket"].(string); tk != "" {
		t.Fatalf("ticket = %q, want empty (daemon not halted)", tk)
	}
}
