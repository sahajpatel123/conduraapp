
package tui

import "testing"

// TestParseAddr_WithSchemeReturnsParsedParts pins the happy
// path: an address of the form "scheme://host:port" MUST be
// split into the scheme and host:port. A regression that
// returned the full address as the host would cause the IPC
// client to make requests to "http://scheme://host:port/".
func TestParseAddr_WithSchemeReturnsParsedParts(t *testing.T) {
	cases := []struct {
		addr      string
		wantScheme string
		wantHost   string
	}{
		{"unix:///tmp/condurad.sock", "unix", "/tmp/condurad.sock"},
		{"http://127.0.0.1:9999", "http", "127.0.0.1:9999"},
		{"https://example.com:8443", "https", "example.com:8443"},
		{"ws://localhost:1234", "ws", "localhost:1234"},
	}
	for _, c := range cases {
		t.Run(c.addr, func(t *testing.T) {
			scheme, host := parseAddr(c.addr)
			if scheme != c.wantScheme {
				t.Errorf("scheme = %q, want %q", scheme, c.wantScheme)
			}
			if host != c.wantHost {
				t.Errorf("host = %q, want %q", host, c.wantHost)
			}
		})
	}
}

// TestParseAddr_NoSchemeDefaultsToTCP pins the no-scheme
// fallback: an address without "://" MUST return ("tcp", addr)
// as the scheme. The IPC client uses "tcp" as the default
// transport (via http.Transport over TCP); a regression to
// another default would break daemon discovery.
func TestParseAddr_NoSchemeDefaultsToTCP(t *testing.T) {
	cases := []string{
		"127.0.0.1:9999",
		"localhost:8080",
		"example.com",
	}
	for _, addr := range cases {
		t.Run(addr, func(t *testing.T) {
			scheme, host := parseAddr(addr)
			if scheme != "tcp" {
				t.Errorf("scheme = %q, want \"tcp\" (default fallback)", scheme)
			}
			if host != addr {
				t.Errorf("host = %q, want %q (full address passed through)", host, addr)
			}
		})
	}
}

// TestParseAddr_EmptyStringReturnsEmptyTCP pins the empty-input
// contract: parseAddr("") MUST return ("tcp", "") without
// panicking. The IPC client receives an empty string when
// the daemon discovery file is missing or empty.
func TestParseAddr_EmptyStringReturnsEmptyTCP(t *testing.T) {
	scheme, host := parseAddr("")
	if scheme != "tcp" {
		t.Errorf("scheme = %q, want \"tcp\"", scheme)
	}
	if host != "" {
		t.Errorf("host = %q, want \"\"", host)
	}
}

// TestParseAddr_MalformedSchemeSeparatorReturnsTCP pins the
// defensive guard: an address with a ":" but no "//" (e.g.,
// IPv6 address "::1") MUST fall back to the no-scheme path.
// A regression that crashed on ":" without "//" would break
// IPv6 addresses.
func TestParseAddr_MalformedSchemeSeparatorReturnsTCP(t *testing.T) {
	// "::1" has a ":" but not "://".
	scheme, host := parseAddr("::1")
	if scheme != "tcp" {
		t.Errorf("scheme = %q, want \"tcp\" (no '://' so fall back)", scheme)
	}
	if host != "::1" {
		t.Errorf("host = %q, want \"::1\" (passed through unchanged)", host)
	}
}

// TestParseAddr_EmptySchemeAcceptedAsIs pins the current
// behavior: parseAddr("://host:9999") returns ("", "host:9999").
// This is a degenerate case (empty scheme) that the parser
// handles by simply splitting on the first "://" without
// validating the scheme. A regression that added a "scheme
// must be non-empty" guard would break callers that pass
// addresses through this path. Pin the current contract.
func TestParseAddr_EmptySchemeAcceptedAsIs(t *testing.T) {
	scheme, host := parseAddr("://host:9999")
	// The CURRENT behavior: empty scheme is allowed (degenerate
	// but accepted). The downstream code uses the scheme only
	// to switch transport ("unix" vs default HTTP).
	if scheme != "" {
		t.Errorf("scheme = %q, want \"\" (degenerate empty scheme accepted)", scheme)
	}
	if host != "host:9999" {
		t.Errorf("host = %q, want \"host:9999\"", host)
	}
}
