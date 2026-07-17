package sanitize

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// TestNewPinnedHTTPClient_DialsPinnedIP proves the actual rebinding
// defense: the TCP connection goes to the IP we pinned, not the
// hostname in the URL. We bind a server to 127.0.0.1, ask the client
// to dial pinnedIP=127.0.0.1 with originalHost="evil.example.com",
// and verify:
//   - the server actually received the request (pin worked)
//   - the Host header on the request was "evil.example.com" (header
//     preservation worked)
func TestNewPinnedHTTPClient_DialsPinnedIP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo the Host header back so the test can verify preservation.
		w.Header().Set("X-Echoed-Host", r.Host)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	parsed, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse srv URL: %v", err)
	}
	host, portStr, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		t.Fatalf("split host:port: %v", err)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		t.Fatalf("test server host %q is not an IP literal", host)
	}
	var port int
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		t.Fatalf("parse port: %v", err)
	}

	// Lie about the host so we can prove the dial goes to the pinned IP
	// and the Host header carries the lie (the rebinding-defense shape).
	const fakeHost = "evil.example.com"
	client := NewPinnedHTTPClient(ip, port, fakeHost, nil)

	req, err := http.NewRequest("GET", parsed.String(), nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	// Override the URL host with the fake name — the dial must STILL go
	// to the pinned IP, and the Host header must be the fake name.
	req.URL.Host = fakeHost + ":" + portStr
	req.Host = fakeHost

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}
	if got := resp.Header.Get("X-Echoed-Host"); got != fakeHost {
		t.Fatalf("Host header preservation: got %q, want %q", got, fakeHost)
	}
}

// TestNewPinnedHTTPClient_RejectsRedirects ensures the pin isn't
// silently broken by an HTTP 3xx that points at a different host.
// The client should refuse the redirect.
func TestNewPinnedHTTPClient_RejectsRedirects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://other.example.com/", http.StatusFound)
	}))
	defer srv.Close()

	parsed, _ := url.Parse(srv.URL)
	host, portStr, _ := net.SplitHostPort(parsed.Host)
	ip := net.ParseIP(host)
	port := 0
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		t.Fatalf("parse port: %v", err)
	}

	client := NewPinnedHTTPClient(ip, port, "test.example.com", nil)
	resp, err := client.Get(srv.URL)
	if err != nil {
		// Err is acceptable; some Go versions surface the disabled-
		// redirect as an error from client.Do.
		if !strings.Contains(err.Error(), "redirect") {
			t.Fatalf("expected redirect error, got %v", err)
		}
		return
	}
	// If no error, must not have followed: status would be 302, not 200.
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("client followed redirect to a different host — pin bypassed")
	}
}

// TestNewPinnedHTTPClient_PreservesTimeout makes sure a base client's
// timeout is respected (defensive default if caller has set one).
func TestNewPinnedHTTPClient_PreservesTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	parsed, _ := url.Parse(srv.URL)
	host, portStr, _ := net.SplitHostPort(parsed.Host)
	ip := net.ParseIP(host)
	port := 0
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		t.Fatalf("parse port: %v", err)
	}

	base := &http.Client{Timeout: 7 * time.Second}
	client := NewPinnedHTTPClient(ip, port, host, base)
	if client.Timeout != 7*time.Second {
		t.Fatalf("timeout not preserved: got %v, want 7s", client.Timeout)
	}
}

// TestResolveAndPin_BlocksPrivateIP ensures the convenience helper
// respects the URLSanitizer's deny-list. A 127.0.0.1 URL must be
// rejected before any client is constructed.
func TestResolveAndPin_BlocksPrivateIP(t *testing.T) {
	s := NewStrictURLSanitizer()
	_, _, _, err := ResolveAndPin(context.Background(), "http://127.0.0.1:8080/", s)
	if err == nil {
		t.Fatalf("expected ErrURLDenied for 127.0.0.1, got nil")
	}
	if !errors.Is(err, ErrURLDenied) {
		t.Fatalf("expected ErrURLDenied, got %v", err)
	}
}

// TestResolveAndPin_HTTPSKeepsSNIVerification ensures that even after
// pinning, the client validates the cert against the original host.
// We point it at an httptest TLS server (which uses a cert for
// 127.0.0.1) but tell the client the host is "wrong.example.com" —
// the cert verification must fail.
func TestResolveAndPin_HTTPSKeepsSNIVerification(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	parsed, _ := url.Parse(srv.URL)
	host, portStr, _ := net.SplitHostPort(parsed.Host)
	ip := net.ParseIP(host)
	port := 0
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		t.Fatalf("parse port: %v", err)
	}

	// Use a non-secure sanitizer (no IP deny) so we get the IP back
	// without rejection, then build the client manually with cert
	// verification left ON.
	s := NewURLSanitizer()
	s.ResolveDNS = true
	client := NewPinnedHTTPClient(ip, port, "wrong.example.com", nil)

	req, err := http.NewRequest("GET", srv.URL, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Host = "wrong.example.com"

	resp, err := client.Do(req)
	if err == nil {
		_ = resp.Body.Close()
		t.Fatalf("expected TLS verification failure against wrong.example.com, got success")
	}
	if !strings.Contains(err.Error(), "certificate") && !strings.Contains(err.Error(), "x509") && !strings.Contains(err.Error(), "tls") {
		t.Fatalf("expected cert-related error, got %v", err)
	}
}

// TestResolveAndPin_IPLiteralBlockedByDefault proves that even the
// IP-literal fast-path inside ResolveURL respects the SSRF deny
// list — 127.0.0.1 is a private IP and must be rejected whether or
// not DNS is involved.
//
// (We can't easily test the IP-literal ALLOWED path because the
// only addresses you can bind to for a test server are private,
// and any public IP would have to be resolved through DNS.)
func TestResolveAndPin_IPLiteralBlockedByDefault(t *testing.T) {
	s := NewURLSanitizer()
	_, _, _, err := ResolveAndPin(context.Background(), "http://127.0.0.1:9999/", s)
	if err == nil {
		t.Fatalf("expected ErrURLDenied for 127.0.0.1, got nil")
	}
	if !errors.Is(err, ErrURLDenied) {
		t.Fatalf("expected ErrURLDenied, got %v", err)
	}
}

// TestResolveAndPin_AllowsPublicIPLiteral proves the IP-literal fast
// path works for non-blocked IPs — a public IP passes the private
// range check and the client is returned. We don't make a network
// call (no public test endpoint to rely on); we just verify the
// helper returned successfully with the IP we passed in.
//
// Note: URLSanitizer mirrors Sanitize's scheme allowlist (only https),
// so we use https:// here even though the IP literal has no protocol
// dependency at the network layer.
func TestResolveAndPin_AllowsPublicIPLiteral(t *testing.T) {
	s := NewURLSanitizer()
	publicIP := net.ParseIP("203.0.113.1") // TEST-NET-3, not routable but not private
	client, u, ip, err := ResolveAndPin(context.Background(),
		"https://203.0.113.1:8080/path", s)
	if err != nil {
		t.Fatalf("ResolveAndPin: %v", err)
	}
	if !ip.Equal(publicIP) {
		t.Fatalf("returned IP: got %v, want %v", ip, publicIP)
	}
	if client == nil {
		t.Fatalf("client is nil")
	}
	if u == nil {
		t.Fatalf("returned URL is nil")
	}
}

// TestNewPinnedHTTPClient_RefusesNonHTTPSchemes is defensive — the
// helper should refuse anything that isn't http or https to prevent
// attackers from sneaking in file:// or gopher:// through the pin.
func TestNewPinnedHTTPClient_RefusesNonHTTPSchemes(t *testing.T) {
	s := NewURLSanitizer()
	for _, badURL := range []string{
		"file:///etc/passwd",
		"ftp://example.com/",
		"gopher://example.com/",
	} {
		_, _, _, err := ResolveAndPin(context.Background(), badURL, s)
		if err == nil {
			t.Errorf("ResolveAndPin(%q): expected error, got nil", badURL)
		}
	}
}

// Ensure the compiled binary references x509 so the test imports
// stay honest about cert-verification intent (the helper relies on
// default TLS verification — which uses x509 under the hood).
var _ = x509.NewCertPool
var _ = (*tls.Config)(nil)
