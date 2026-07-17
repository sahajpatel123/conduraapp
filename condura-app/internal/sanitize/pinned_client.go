// Pinned HTTP client — closes the TOCTOU window between DNS
// resolution and the actual TCP dial that makes DNS-rebinding
// attacks possible.
//
// The standard rebinding attack:
//
//   1. Attacker controls `evil.example.com` (passes hostname
//      allowlist / pattern checks).
//   2. First DNS lookup (at Sanitize time) returns a public IP.
//      Allowlist passes, request proceeds.
//   3. Second DNS lookup (at http.Client.Do time, ~ms later) returns
//      a private IP (127.0.0.1, 169.254.169.254, etc.).
//   4. Default http.Client uses the rebinding target as the dial
//      address, defeating the SSRF guard.
//
// This package's defense:
//
//   1. Resolve the hostname ONCE via URLSanitizer.ResolveURL — get a
//      pinned IP that passed the private-range check.
//   2. Hand the pinned IP to NewPinnedHTTPClient, which returns an
//      http.Client whose transport dials pinnedIP:port regardless of
//      what the request URL or the current DNS says.
//   3. The original hostname is preserved in the Host header (HTTP)
//      and ServerName / TLS verification (HTTPS) so vhost routing
//      and cert validation keep working.
//
// The returned client does NOT skip certificate verification — it
// pins the dial address, not the trust anchor. TLS still validates
// against the original hostname's cert.

package sanitize

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// pinnedIdleConnTimeout is the keep-alive window for the dial addressed
// to the pinned IP. Generous enough that the common multi-request flow
// keeps the TCP connection open without holding open more than one
// generation of remote route. Named rather than magic-numbered so
// golangci-lint's `mnd` lint accepts it without an ignore-list amendment.
const pinnedIdleConnTimeout = 90 * time.Second

// NewPinnedHTTPClient returns an *http.Client whose transport dials
// pinnedIP at the given port for every connection, while keeping
// the original host in the Host header (HTTP/1.1) and the TLS SNI
// (HTTPS). Cert verification still applies against the original host.
//
// If base is non-nil, its Timeout is preserved; other fields are
// intentionally not inherited to avoid leaking the caller's cookie
// jar, redirect policy, etc. into the rebinding-defense code path.
// Pass nil for a fresh client with sensible defaults (10s timeout).
func NewPinnedHTTPClient(pinnedIP net.IP, port int, originalHost string, base *http.Client) *http.Client {
	timeout := 10 * time.Second
	if base != nil && base.Timeout > 0 {
		timeout = base.Timeout
	}
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           pinnedDialContext(dialer, pinnedIP, port),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          10,
		IdleConnTimeout:       pinnedIdleConnTimeout,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	// Make sure the request URL still uses the original host so
	// Go's http stack sets Host header / TLS ServerName correctly.
	// The DialContext above will rewrite the dial address to pinnedIP.
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
		// Don't follow redirects automatically — each redirect could
		// re-resolve the (different) Location URL and bypass the pin.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 1 {
				return errors.New("pinned client: redirects disabled")
			}
			return nil
		},
	}
}

// pinnedDialContext returns a DialContext that ignores the address
// in the request and always dials pinnedIP:port. This is the actual
// rebinding defense — even if the request URL says evil.example.com,
// the TCP connection goes to the IP we resolved at Sanitize time.
func pinnedDialContext(d *net.Dialer, pinnedIP net.IP, port int) func(ctx context.Context, _, _ string) (net.Conn, error) {
	addr := net.JoinHostPort(pinnedIP.String(), strconv.Itoa(port))
	return func(ctx context.Context, _, _ string) (net.Conn, error) {
		return d.DialContext(ctx, "tcp", addr)
	}
}

// ResolveAndPin is the convenience helper that pairs
// URLSanitizer.ResolveURL with NewPinnedHTTPClient. Callers get a
// ready-to-use http.Client whose dial address is pinned to the IP
// that passed the SSRF check.
//
// Returns:
//   - the client (ready to use)
//   - the resolved IP (for logging / debugging)
//   - the URL rewritten to use the original host with the pinned
//     IP — pass this to client.Do; the dial will go to pinnedIP
//     regardless of what the URL says
//   - any error from resolution
//
// Use the returned URL with the returned client; do NOT pass the
// original raw URL (it would re-resolve through the default
// transport's DNS).
func ResolveAndPin(ctx context.Context, rawURL string, s *URLSanitizer) (*http.Client, *url.URL, net.IP, error) {
	if s == nil {
		s = NewStrictURLSanitizer()
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, nil, nil, ErrURLDenied
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, nil, nil, ErrURLDenied
	}
	ip, err := s.ResolveURL(ctx, rawURL)
	if err != nil {
		return nil, nil, nil, err
	}
	if ip == nil {
		// ResolveURL returned nil — no resolution happened (ResolveDNS
		// disabled, or IP literal handling). Refuse to pin in this case
		// because we have nothing to pin to.
		return nil, nil, nil, errors.New("sanitize: ResolveURL returned no IP — caller must enable ResolveDNS or pass an IP literal")
	}
	host := u.Hostname()
	port := 0
	if p := u.Port(); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}
	if port == 0 {
		switch u.Scheme {
		case "http":
			port = 80
		case "https":
			port = 443
		}
	}
	client := NewPinnedHTTPClient(ip, port, host, nil)
	return client, u, ip, nil
}
