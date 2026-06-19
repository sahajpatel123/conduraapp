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
