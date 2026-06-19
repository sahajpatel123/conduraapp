// Package halt — Layer 3 of the kill switch: network isolation.
//
// Spec: "A separate OS process owns a `pf` (mac) / `netsh` (win) rule
// blocking all egress from the user's UID except the LLM provider IPs.
// The agent process cannot stop it."
//
// This file ships the in-process SKELETON of that design. It is
// wired into the daemon's HTTP transport so that when the kill
// switch fires, all outbound HTTP from the agent is denied except
// to allow-listed provider hosts. The interface (NetworkGuard) is
// designed so the in-process guard can be swapped for a real
// pf/netsh process without changing call sites.
//
// In v0.1.0 this is "soft" Layer 3: the guard runs in the same
// process as the agent. A misbehaving agent that bypasses the
// guard could still reach the network. Hard Layer 3 (a real
// separate process with pf/netsh) is in the v0.2.0 roadmap.
//
// Reference: CLAUDE.md §5.3, decision #26.
package halt

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// NetworkGuard is the kill-switch's network-isolation component.
// It enforces two policies:
//   - When halted: deny all outbound connections.
//   - When running: deny any host not in AllowList.
//
// Implementations must be safe for concurrent use.
type NetworkGuard interface {
	// Allow reports whether a connection to host is permitted.
	// host is the bare hostname (no port, no scheme).
	Allow(host string) bool
	// WrapTransport returns an http.RoundTripper that enforces
	// the policy. nil means "no transport wrapping" (the caller
	// can fall back to http.DefaultTransport).
	WrapTransport(rt http.RoundTripper) http.RoundTripper
	// Halt denies all connections until Resume is called.
	Halt(reason string) error
	// Resume re-enables connections to allow-listed hosts.
	Resume() error
	// State returns the current guard state.
	State() GuardState
}

// GuardState describes the network guard.
type GuardState struct {
	Halted    bool      `json:"halted"`
	Since     time.Time `json:"since,omitempty"`
	AllowList []string  `json:"allow_list"`
}

// DefaultProviderAllowList returns the canonical allow-list for
// known LLM provider base URLs. Users can extend it via config.
func DefaultProviderAllowList() []string {
	return []string{
		"api.anthropic.com",
		"api.openai.com",
		"generativelanguage.googleapis.com",
		"api.x.ai",
		"api.mistral.ai",
		"api.deepseek.com",
		"openrouter.ai",
		"api.groq.com",
		"api.together.xyz",
		"api.fireworks.ai",
		"huggingface.co",   // for openWakeWord model downloads
		"hub.condura.app",  // Skills Hub (locked decision #18)
		"hub.synaptic.app", // Skills Hub legacy URL (transitional)
		"127.0.0.1",        // local providers (Ollama, LocalAI, LM Studio, vLLM)
		"localhost",
	}
}

// InProcessGuard is the in-process implementation of NetworkGuard.
// It is the v0.1.0 default. It is correct in the sense that the
// daemon's HTTP transport is wrapped by WrapTransport, so all
// well-behaved code paths (every LLM client) honor the policy.
// It is NOT a hard guarantee because a determined misbehaving
// agent can skip the transport.
type InProcessGuard struct {
	mu        sync.RWMutex
	halted    bool
	since     time.Time
	reason    string
	allowList map[string]bool
}

// NewInProcessGuard returns a guard seeded with the default
// provider allow-list. Callers may extend via Allow() / Deny().
func NewInProcessGuard() *InProcessGuard {
	g := &InProcessGuard{
		allowList: map[string]bool{},
	}
	for _, h := range DefaultProviderAllowList() {
		g.allowList[h] = true
	}
	return g
}

// Allow reports whether host is permitted.
func (g *InProcessGuard) Allow(host string) bool {
	if host == "" {
		return false
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.halted {
		return false
	}
	if g.allowList[host] {
		return true
	}
	// Subdomain match: api.openai.com is OK if openai.com is
	// allow-listed, etc.
	for allowed := range g.allowList {
		if host == allowed {
			return true
		}
		if isSubdomain(host, allowed) {
			return true
		}
	}
	return false
}

// Halt denies all outbound connections.
func (g *InProcessGuard) Halt(reason string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.halted = true
	g.since = time.Now()
	g.reason = reason
	return nil
}

// Resume re-enables the allow-list.
func (g *InProcessGuard) Resume() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.halted = false
	g.since = time.Time{}
	g.reason = ""
	return nil
}

// AllowHost adds host to the allow-list at runtime.
func (g *InProcessGuard) AllowHost(host string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.allowList[host] = true
}

// DenyHost removes host from the allow-list.
func (g *InProcessGuard) DenyHost(host string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.allowList, host)
}

// State returns the current state.
func (g *InProcessGuard) State() GuardState {
	g.mu.RLock()
	defer g.mu.RUnlock()
	list := make([]string, 0, len(g.allowList))
	for h := range g.allowList {
		list = append(list, h)
	}
	return GuardState{
		Halted:    g.halted,
		Since:     g.since,
		AllowList: list,
	}
}

// WrapTransport returns an http.RoundTripper that rejects
// requests to non-allow-listed hosts.
func (g *InProcessGuard) WrapTransport(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}
	return &guardedTransport{guard: g, inner: rt}
}

// guardedTransport is the http.RoundTripper that enforces
// the InProcessGuard's policy.
type guardedTransport struct {
	guard *InProcessGuard
	inner http.RoundTripper
}

// RoundTrip implements http.RoundTripper. It rejects requests
// to hosts that the guard doesn't allow.
func (t *guardedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil || req.URL == nil {
		return nil, fmt.Errorf("network guard: nil request")
	}
	host := req.URL.Hostname()
	if !t.guard.Allow(host) {
		return nil, &NetworkDeniedError{Host: host, Halted: isHalted(t.guard)}
	}
	return t.inner.RoundTrip(req)
}

// NetworkDeniedError is returned by the guarded transport when
// the guard denies a request.
type NetworkDeniedError struct {
	Host   string
	Halted bool
}

func (e *NetworkDeniedError) Error() string {
	if e.Halted {
		return fmt.Sprintf("network guard: kill switch active, all egress denied (host=%s)", e.Host)
	}
	return fmt.Sprintf("network guard: host %q not in allow-list", e.Host)
}

// isHalted is a lock-free read of the halted flag.
func isHalted(g *InProcessGuard) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.halted
}

// isSubdomain reports whether sub is a subdomain of base, or equal to it.
// e.g. isSubdomain("api.openai.com", "openai.com") = true.
// e.g. isSubdomain("api.openai.com", "api.openai.com") = true.
func isSubdomain(sub, base string) bool {
	if sub == base {
		return true
	}
	if len(sub) <= len(base)+1 {
		return false
	}
	if !hasSuffix(sub, base) {
		return false
	}
	// Must be preceded by a dot.
	return sub[len(sub)-len(base)-1] == '.'
}

func hasSuffix(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}

// WireToHTTPClient attaches the guard to an http.Client. Used by
// the LLM clients in internal/llm to enforce the policy on every
// outbound request the daemon makes. The hook is opt-in; the
// daemon's llm.Registry is rebuilt when the guard's state changes
// so new clients pick up the latest policy.
func WireToHTTPClient(guard NetworkGuard, c *http.Client) {
	if c == nil {
		c = &http.Client{Timeout: 30 * time.Second}
	}
	if guard != nil {
		c.Transport = guard.WrapTransport(c.Transport)
	}
}

// Connect probes a host:port pair under the guard. Returns nil if
// the host is allowed; an error otherwise. Useful for diagnostic
// RPCs and tests.
func Connect(ctx context.Context, guard NetworkGuard, hostport string) error {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return fmt.Errorf("network guard: bad hostport %q: %w", hostport, err)
	}
	if !guard.Allow(host) {
		return &NetworkDeniedError{Host: host}
	}
	d := net.Dialer{Timeout: 5 * time.Second}
	c, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}
	_ = c.Close()
	return nil
}

// FromURL returns the host of u as a string suitable for Allow().
func FromURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	return u.Hostname()
}
