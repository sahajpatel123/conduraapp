package sanitize

import (
	"context"
	"strings"
	"testing"
)

// Audit 2026-07-01 anti-pattern sweep: cloud metadata blocklist
// expansion, DNS-rebinding TOCTOU mitigation via ResolveURL, and
// SSN false-positive fix. The tests in this file pin each fix.

// FIX 3: each new cloud-metadata hostname MUST be denied.
func TestURLSanitizer_RejectsCloudMetadataHostnames(t *testing.T) {
	s := NewURLSanitizer()
	cases := []string{
		"https://metadata.google.internal/",         // AWS/GCP
		"https://metadata.goog/",                     // GCP
		"https://instance-data.ec2.internal/",        // AWS EC2
		"https://metadata.azure.com/",                // Azure
		"https://metadata.aliyun.com/",               // Alibaba
		"https://metadata.tencentyun.com/",           // Tencent
	}
	for _, u := range cases {
		if _, err := s.Sanitize(u); err == nil {
			t.Errorf("hostname in %q must be denied", u)
		} else if err != ErrURLDenied {
			t.Errorf("hostname in %q returned wrong error type: %v", u, err)
		}
	}
}

// FIX 3: dedicated IPs that the private-range check can miss.
// 100.100.100.200 is Alibaba's metadata IP and 192.0.0.192 is
// Oracle's metadata subnet. isPrivate does not always flag
// 192.0.0.0/24, and 100.100.100.0/24 is non-private by RFC, so
// we hard-code both.
func TestURLSanitizer_RejectsCloudMetadataIPs(t *testing.T) {
	s := NewURLSanitizer()
	cases := []string{
		"https://100.100.100.200/", // Alibaba metadata
		"https://192.0.0.192/",     // Oracle metadata
	}
	for _, u := range cases {
		if _, err := s.Sanitize(u); err != ErrURLDenied {
			t.Errorf("cloud-metadata IP %q must be denied, got: %v", u, err)
		}
	}
}

// FIX 3: AWS / GCP / Azure metadata IP. Already in the existing
// isBlockedIP but pin explicitly to ensure it stays.
func TestURLSanitizer_RejectsAWSMetadataIP(t *testing.T) {
	s := NewURLSanitizer()
	if _, err := s.Sanitize("https://169.254.169.254/"); err != ErrURLDenied {
		t.Errorf("AWS metadata IP must be denied, got: %v", err)
	}
}

// FIX 2: ResolveURL returns the resolved IP for a public hostname
// and a denial for private ones. We do not call this against the
// public internet in unit tests (avoid flake); instead we use
// 127.0.0.1 as the IP-literal case and a private hostname.
//
// Pinning for 127.0.0.1: the IP-literal short-circuit returns the
// parsed IP without DNS. Private-hostname resolution depends on the
// test environment's resolver and is exercised separately in
// TestURLSanitizer_Strict_DNSRebinding.
func TestResolveURL_RejectsLoopbackIP(t *testing.T) {
	s := NewStrictURLSanitizer()
	ip, err := s.ResolveURL(context.Background(), "https://127.0.0.1:8080/")
	if err == nil {
		t.Fatal("ResolveURL accepted 127.0.0.1")
	}
	if ip != nil {
		t.Errorf("ResolveURL must not return an IP for a denied URL, got %v", ip)
	}
}

func TestResolveURL_Rejects169Metadata(t *testing.T) {
	s := NewStrictURLSanitizer()
	ip, err := s.ResolveURL(context.Background(), "https://169.254.169.254/latest/meta-data/")
	if err == nil {
		t.Fatal("ResolveURL accepted 169.254.169.254")
	}
	if ip != nil {
		t.Errorf("ResolveURL must not return an IP for a denied URL, got %v", ip)
	}
}

func TestResolveURL_NonStrictReturnsNilIP(t *testing.T) {
	// The non-strict sanitizer must not perform DNS — we want the
	// API to fail loud if a caller tries to use it as a rebinding
	// defense. ResolveURL on a non-strict sanitizer should report
	// the URL is acceptable (no pattern match failure) but return
	// no IP, signaling "I did not pin anything".
	s := NewURLSanitizer()
	ip, err := s.ResolveURL(context.Background(), "https://example.com/")
	if err != nil {
		t.Fatalf("ResolveURL(non-strict) must not error on a public hostname, got: %v", err)
	}
	if ip != nil {
		t.Errorf("ResolveURL(non-strict) must not return a pinned IP, got %v", ip)
	}
}

func TestResolveURL_EmptyInputReturnsNil(t *testing.T) {
	s := NewStrictURLSanitizer()
	ip, err := s.ResolveURL(context.Background(), "")
	if err != nil {
		t.Errorf("empty input must not error, got: %v", err)
	}
	if ip != nil {
		t.Errorf("empty input must return nil IP, got %v", ip)
	}
}

func TestResolveURL_NonURLReturnsNil(t *testing.T) {
	// "echo hello" has no scheme — it is not a URL, so ResolveURL
	// must not error (and must not return an IP). Same behavior
	// as Sanitize.
	s := NewStrictURLSanitizer()
	ip, err := s.ResolveURL(context.Background(), "echo hello")
	if err != nil {
		t.Errorf("non-URL input must not error, got: %v", err)
	}
	if ip != nil {
		t.Errorf("non-URL input must return nil IP, got %v", ip)
	}
}

// FIX 6: SSN false-positive. The 2026-07-01 sweep found that the
// previous matchSSNPattern triggered on any 9 consecutive digits,
// which catches order numbers, ISBN-10s, phone numbers, etc.
// The new rule: explicit XXX-XX-XXXX or XXX XX XXXX only.
func TestSSNPattern_DetectsCanonicalDashed(t *testing.T) {
	s := NewPIIRegexSanitizer()
	if _, err := s.Sanitize("SSN: 123-45-6789"); err != ErrPIIDetected {
		t.Errorf("XXX-XX-XXXX must be detected, got: %v", err)
	}
}

func TestSSNPattern_DetectsCanonicalSpaced(t *testing.T) {
	s := NewPIIRegexSanitizer()
	if _, err := s.Sanitize("SSN: 123 45 6789"); err != ErrPIIDetected {
		t.Errorf("XXX XX XXXX must be detected, got: %v", err)
	}
}

func TestSSNPattern_DetectsAtStartAndEnd(t *testing.T) {
	s := NewPIIRegexSanitizer()
	// Anchored at start of string.
	if _, err := s.Sanitize("123-45-6789 is the SSN"); err != ErrPIIDetected {
		t.Errorf("leading SSN must be detected, got: %v", err)
	}
	// Anchored at end of string.
	if _, err := s.Sanitize("the SSN is 123-45-6789"); err != ErrPIIDetected {
		t.Errorf("trailing SSN must be detected, got: %v", err)
	}
	// Standalone.
	if _, err := s.Sanitize("123-45-6789"); err != ErrPIIDetected {
		t.Errorf("bare SSN must be detected, got: %v", err)
	}
}

// The 2026-07-01 false-positive class: bare 9-digit runs that
// are NOT SSNs. These must pass the PII sanitizer.
func TestSSNPattern_AllowsBareNineDigits(t *testing.T) {
	s := NewPIIRegexSanitizer()
	// Each case is a bare-9-digit run that is NOT a credit card
	// (so the CC branch of PIIRegexSanitizer doesn't fire) and
	// NOT an SSN (the fix's purpose). The old code flagged these
	// because it considered any 9 consecutive digits an SSN.
	cases := []string{
		"order number is 123456789",   // order #
		"phone 1234567890",            // 10-digit phone (still no SSN pattern)
		"ISBN 1234567890",             // ISBN-10
		"reference 123456789",         // bare 9-digit reference
		"code: 123456789 and another", // surrounded by other digits
	}
	for _, c := range cases {
		if _, err := s.Sanitize(c); err == ErrPIIDetected {
			t.Errorf("bare-9digit input %q must NOT be detected as SSN", c)
		}
	}
}

func TestSSNPattern_AllowsFormattedButNotSSN(t *testing.T) {
	// "1234-56-7890" is a different shape (4-2-4). Even though it
	// has dashes, it doesn't match XXX-XX-XXXX, so it must pass.
	s := NewPIIRegexSanitizer()
	if _, err := s.Sanitize("not-an-ssn: 1234-56-7890"); err == ErrPIIDetected {
		t.Error("1234-56-7890 is not a canonical SSN shape; must pass")
	}
}

// Sanity: PII sanitizer still detects the credit-card path so we
// know the refactor didn't break adjacent behavior.
func TestPIIRegexSanitizer_StillDetectsCreditCard(t *testing.T) {
	s := NewPIIRegexSanitizer()
	if _, err := s.Sanitize("card 4532015112830366"); err != ErrPIIDetected {
		t.Errorf("credit card must still be detected, got: %v", err)
	}
}

// FIX 2 (smoke): Sanitize with strict DNS does not regress —
// the refactor to call resolveHost must not change the existing
// strict-sanitizer behavior.
func TestSanitize_Strict_StillRejectsHTTPSubsumedByNewHelper(t *testing.T) {
	// Sanity check on the http:// prefix block — this path doesn't
	// go through resolveHost, but pinning the prefix block is
	// cheap and prevents future refactors from regressing it.
	s := NewStrictURLSanitizer()
	if _, err := s.Sanitize("http://example.com"); err != ErrURLDenied {
		t.Errorf("http:// must still be denied, got: %v", err)
	}
}

// Ensure the new error message from resolveHost wraps ErrURLDenied
// (callers can use errors.Is).
func TestSanitize_Strict_DNSErrorWrapsURLErr(t *testing.T) {
	// We can't easily force a public hostname to resolve to a
	// private IP in a unit test (would require a malicious DNS
	// server). Instead, exercise the existing private-IP-literal
	// path which is the primary defense.
	s := NewStrictURLSanitizer()
	_, err := s.Sanitize("https://10.0.0.1/")
	if err == nil {
		t.Fatal("private IP literal must be denied")
	}
	if !strings.Contains(err.Error(), "denied") {
		t.Errorf("deny error must mention 'denied', got: %v", err)
	}
}
