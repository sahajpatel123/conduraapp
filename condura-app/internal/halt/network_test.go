package halt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestInProcessGuard_AllowByDefault(t *testing.T) {
	g := NewInProcessGuard()
	if !g.Allow("api.openai.com") {
		t.Error("api.openai.com should be allow-listed by default")
	}
	if !g.Allow("api.anthropic.com") {
		t.Error("api.anthropic.com should be allow-listed by default")
	}
}

func TestInProcessGuard_DenyUnknown(t *testing.T) {
	g := NewInProcessGuard()
	if g.Allow("evil.example.com") {
		t.Error("evil.example.com should not be allow-listed")
	}
}

func TestInProcessGuard_SubdomainMatch(t *testing.T) {
	g := NewInProcessGuard()
	if !g.Allow("api.openai.com") {
		t.Error("api.openai.com is a subdomain of openai.com — should be allowed")
	}
}

func TestInProcessGuard_HaltDeniesAll(t *testing.T) {
	g := NewInProcessGuard()
	_ = g.Halt("user pressed kill switch")
	if g.Allow("api.openai.com") {
		t.Error("after Halt, even allow-listed hosts must be denied")
	}
	if !g.State().Halted {
		t.Error("State should report Halted=true")
	}
	if g.State().Since.IsZero() {
		t.Error("State.Since should be set after Halt")
	}
}

func TestInProcessGuard_Resume(t *testing.T) {
	g := NewInProcessGuard()
	_ = g.Halt("test")
	_ = g.Resume()
	if !g.Allow("api.openai.com") {
		t.Error("after Resume, allow-list should apply again")
	}
}

func TestInProcessGuard_RuntimeAllowHost(t *testing.T) {
	g := NewInProcessGuard()
	g.AllowHost("my-proxy.local")
	if !g.Allow("my-proxy.local") {
		t.Error("AllowHost should make host allowed")
	}
	g.DenyHost("my-proxy.local")
	if g.Allow("my-proxy.local") {
		t.Error("DenyHost should make host denied")
	}
}

// TestInProcessGuard_ResumePreservesRuntimeAllowHost is the regression
// test for the 2026-07-12 audit finding #1.3: previously, Resume
// re-initialized the allow-list from DefaultProviderAllowList, silently
// dropping any runtime AllowHost() the user had configured between
// daemon startup and Halt. Now Halt snapshots the live allow-list and
// Resume restores it verbatim, so user runtime config survives the
// halt/resume cycle.
func TestInProcessGuard_ResumePreservesRuntimeAllowHost(t *testing.T) {
	g := NewInProcessGuard()

	// User adds a custom host at runtime.
	g.AllowHost("my-proxy.local")
	g.AllowHost("internal.corp.example")

	// Halt — must snapshot the runtime additions.
	if err := g.Halt("user pressed kill switch"); err != nil {
		t.Fatalf("Halt: %v", err)
	}

	// During halt, Allow returns false for everything (including
	// the user's additions). AllowHost during halt is a no-op
	// observation-wise but we still record it so the post-Resume
	// allowList reflects the latest intent.
	if g.Allow("my-proxy.local") {
		t.Error("during halt, even runtime-added hosts must be denied")
	}

	// Resume — must restore runtime additions, not revert to
	// DefaultProviderAllowList seed.
	if err := g.Resume(); err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if !g.Allow("my-proxy.local") {
		t.Error("after Resume, runtime-added host my-proxy.local should still be allowed")
	}
	if !g.Allow("internal.corp.example") {
		t.Error("after Resume, runtime-added host internal.corp.example should still be allowed")
	}
	if !g.Allow("api.openai.com") {
		t.Error("after Resume, default-allowed host api.openai.com should still be allowed")
	}
}

// TestInProcessGuard_ResumePreservesRuntimeDeny verifies the symmetric
// case: a host explicitly denied at runtime (via DenyHost) is NOT
// reintroduced by Resume.
func TestInProcessGuard_ResumePreservesRuntimeDeny(t *testing.T) {
	g := NewInProcessGuard()

	// User removes a default-allowed host at runtime.
	g.DenyHost("api.openai.com")
	if g.Allow("api.openai.com") {
		t.Fatal("setup: DenyHost should remove default-allowed host")
	}

	// Halt then Resume.
	_ = g.Halt("test")
	_ = g.Resume()

	// The runtime deny must survive the cycle.
	if g.Allow("api.openai.com") {
		t.Error("after Resume, runtime-denied host api.openai.com should still be denied")
	}
}

// TestInProcessGuard_DoubleHaltDoesNotOverwriteSnapshot verifies that
// calling Halt twice keeps the original snapshot intact. The first
// Halt captures the user's pre-halt config; if the second Halt
// (while still halted) ran AllowHost/DenyHost on the live list and
// then re-snapshotted, the user's original intent would be lost on
// the subsequent Resume.
func TestInProcessGuard_DoubleHaltDoesNotOverwriteSnapshot(t *testing.T) {
	g := NewInProcessGuard()
	g.AllowHost("user-intent-1.example")
	_ = g.Halt("first halt")

	// During halt, user (or some tool) adds another host. This
	// mutates the live allowList but the snapshot must remain the
	// pre-first-Halt state.
	g.AllowHost("during-halt.example")

	_ = g.Halt("second halt (still halted)")

	// Resume must restore the FIRST snapshot (the one with
	// user-intent-1 but without during-halt).
	_ = g.Resume()

	if !g.Allow("user-intent-1.example") {
		t.Error("user-intent-1 should survive double-halt (was in first snapshot)")
	}
	if g.Allow("during-halt.example") {
		t.Error("during-halt should NOT survive double-halt (added after first snapshot)")
	}
}

// TestInProcessGuard_ResumeWithoutHaltIsNoOp verifies that Resume
// without a prior Halt doesn't reset the allow-list to defaults.
// Previously this case also overwrote runtime additions — caught
// by the same audit finding.
func TestInProcessGuard_ResumeWithoutHaltIsNoOp(t *testing.T) {
	g := NewInProcessGuard()
	g.AllowHost("no-prior-halt.example")

	// Resume without Halt — should be a no-op aside from clearing
	// stale halted/since/reason (which were never set).
	_ = g.Resume()

	if !g.Allow("no-prior-halt.example") {
		t.Error("Resume without prior Halt should preserve runtime allow-list additions")
	}
}

func TestInProcessGuard_WrapTransport(t *testing.T) {
	g := NewInProcessGuard()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	// Allowed: real server.
	c := &http.Client{Transport: g.WrapTransport(nil)}
	resp, err := c.Get(srv.URL)
	if err != nil {
		t.Fatalf("allow-list request should succeed: %v", err)
	}
	_ = resp.Body.Close()

	// Denied: bad host.
	c2 := &http.Client{Transport: g.WrapTransport(nil), Timeout: 2 * time.Second}
	resp2, err := c2.Get("http://evil.example.com/")
	if err == nil {
		_ = resp2.Body.Close()
		t.Error("non-allow-listed host should be denied")
	}
	if resp2 != nil {
		_ = resp2.Body.Close()
	}
}

func TestInProcessGuard_WrapTransport_HaltDeniesEverything(t *testing.T) {
	g := NewInProcessGuard()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	_ = g.Halt("test")
	c := &http.Client{Transport: g.WrapTransport(nil), Timeout: 2 * time.Second}
	resp3, err := c.Get(srv.URL)
	if err == nil {
		_ = resp3.Body.Close()
		t.Error("after Halt, all requests should be denied")
	}
	if resp3 != nil {
		_ = resp3.Body.Close()
	}
}

func TestNetworkDeniedError_Error(t *testing.T) {
	e := &NetworkDeniedError{Host: "evil.com"}
	if e.Error() == "" {
		t.Error("error message should be non-empty")
	}
	e2 := &NetworkDeniedError{Host: "evil.com", Halted: true}
	if e2.Error() == "" {
		t.Error("halted error should be non-empty")
	}
	if e2.Error() == e.Error() {
		t.Error("halted and non-halted should produce different messages")
	}
}

func TestIsSubdomain(t *testing.T) {
	cases := []struct {
		sub, base string
		want      bool
	}{
		{"api.openai.com", "openai.com", true},
		{"api.openai.com", "api.openai.com", true},
		{"openai.com", "api.openai.com", false},
		{"myopenai.com", "openai.com", false}, // not preceded by dot
		{"x.y.z", "z", true},
	}
	for _, tc := range cases {
		if got := isSubdomain(tc.sub, tc.base); got != tc.want {
			t.Errorf("isSubdomain(%q,%q)=%v want %v", tc.sub, tc.base, got, tc.want)
		}
	}
}

func TestHasSuffix(t *testing.T) {
	if !hasSuffix("api.openai.com", "openai.com") {
		t.Error("hasSuffix should match")
	}
	if hasSuffix("openai", "openai.com") {
		t.Error("hasSuffix should not match when sub is shorter")
	}
}

func TestFromURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"https://api.openai.com/v1", "api.openai.com"},
		{"http://localhost:11434/v1", "localhost"},
		{"", ""},
	}
	for _, tc := range cases {
		var got string
		if tc.in != "" {
			u, err := url.Parse(tc.in)
			if err != nil {
				t.Fatal(err)
			}
			got = FromURL(u)
		} else {
			got = FromURL(nil)
		}
		if got != tc.want {
			t.Errorf("FromURL(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestConnect(t *testing.T) {
	g := NewInProcessGuard()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	hostport := srv.Listener.Addr().String()
	if err := Connect(ctx, g, hostport); err != nil {
		t.Errorf("connect to allow-listed test server: %v", err)
	}

	if err := Connect(ctx, g, "evil.example.com:80"); err == nil {
		t.Error("connect to non-allow-listed host should fail")
	}

	_ = g.Halt("test")
	if err := Connect(ctx, g, hostport); err == nil {
		t.Error("connect after Halt should fail")
	}
}

// TestIsLayer3InProcess pins the v0.1.0 Layer 3 status. The
// honest answer is "yes, the guard runs in-process" — which is
// precisely why the GUI surfaces this via daemon.capabilities.
// When v0.2.0 swaps to a real pf/netsh separate process, this
// test must be updated to assert false AND the test below must
// be inverted. Reference: CLAUDE.md §33.5.2 row C4.14.
func TestIsLayer3InProcess(t *testing.T) {
	if !IsLayer3InProcess() {
		t.Fatal("IsLayer3InProcess = false; v0.1.0 ships an in-process guard. " +
			"If you flipped this to false, also update the documentation " +
			"in CLAUDE.md §33.5.2 row C4.14 and the v0.2.0 milestone list.")
	}
}
