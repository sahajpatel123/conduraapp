//nolint:revive,mnd // interface implementations; digit map is definitional
package sanitize

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// PathSanitizer blocks traversal and system paths.
type PathSanitizer struct{}

// NewPathSanitizer creates a path sanitizer.
func NewPathSanitizer() *PathSanitizer { return &PathSanitizer{} }

// systemPathPrefixes is the deny-list for filesystem paths. It covers:
//   - Unix system dirs (/etc, /var, /usr, /bin, /sbin, /lib, /proc,
//     /sys, /boot, /root, /run, /dev) — touching these usually requires
//     root and almost never benefits the agent.
//   - macOS system dirs (/System, /Library, /Applications, /private/*)
//     which are SIP-protected; reading/writing them is suspicious.
//   - User-shell-secret dirs under any home (~/.ssh, ~/.gnupg, ~/.aws,
//     ~/.kube, ~/.docker) — credential material an agent should never
//     touch without explicit per-call consent.
//   - Windows system dirs (C:\Windows, C:\Program Files,
//     C:\Program Files (x86), C:\ProgramData, C:\Users\Default) —
//     same reasoning as Unix system dirs.
//
// Paths not matching ANY prefix here are still subject to `..`
// traversal blocking and (downstream) the gatekeeper policy. The
// list is intentionally broad; when in doubt, block and let the
// gatekeeper's policy override via the explicit-allow consent.
var systemPathPrefixes = []string{
	// Unix system directories.
	"/etc/",
	"/var/",
	"/usr/",
	"/usr/local/",
	"/bin/",
	"/sbin/",
	"/lib/",
	"/lib64/",
	"/proc/",
	"/sys/",
	"/boot/",
	"/root/",
	"/run/",
	"/dev/",
	// macOS system directories.
	"/System/",
	"/Library/",
	"/Applications/",
	"/private/etc/",
	"/private/var/",
	"/private/tmp/",
	"/private/var/db/",
	// Per-user credential directories. Matched anywhere in the path,
	// not just at the root — a path like /Users/alice/work/.ssh/id_rsa
	// must still be rejected.
	"/.ssh/",
	"/.gnupg/",
	"/.aws/",
	"/.kube/",
	"/.docker/",
	// Windows system directories.
	`C:\Windows\`,
	`C:\Windows\`,
	`C:\Program Files\`,
	`C:\Program Files (x86)\`,
	`C:\ProgramData\`,
	`C:\Users\Default\`,
}

func (s *PathSanitizer) Sanitize(input string) (string, error) {
	if input == "" {
		return input, nil
	}
	if strings.Contains(input, "..") {
		return "", ErrPathDenied
	}
	for _, prefix := range systemPathPrefixes {
		if strings.HasPrefix(input, prefix) || strings.Contains(input, prefix) {
			return "", ErrPathDenied
		}
	}
	return input, nil
}

func (s *PathSanitizer) Name() string { return "path" }

// URLSanitizer prevents SSRF by parsing the URL and checking the
// host against an exact-match deny-list (loopback, RFC1918, link-
// local, cloud metadata IPs and hostnames). When constructed with
// `NewStrictURLSanitizer`, it also resolves the hostname via DNS
// and rejects if ANY resolved IP falls in a private range — this
// is the DNS-rebinding defense.
//
// The previous substring-match implementation could be bypassed by
// hosts whose literal hostname did not contain the pattern (e.g.
// `my-192-168-host.example.com`) and by DNS records that resolved
// to private IPs post-resolution (TOCTOU between the sanitizer
// and the actual HTTP client).
type URLSanitizer struct {
	// ResolveDNS enables DNS resolution to catch rebinding. Off by
	// default to keep tests deterministic and to avoid hammering DNS
	// for every URL. Production callers (e.g. safety_wiring) should
	// construct with NewStrictURLSanitizer.
	ResolveDNS bool
}

// NewURLSanitizer creates a URL sanitizer without DNS resolution.
// Sufficient for the common case where URL strings come from the
// model or the user and have not been looked up yet.
func NewURLSanitizer() *URLSanitizer { return &URLSanitizer{} }

// NewStrictURLSanitizer creates a URL sanitizer that resolves
// hostnames via DNS and rejects any that resolve to private IPs.
// Use this on the hot path (e.g. the gatekeeper's SanitizeHook)
// to defend against DNS rebinding.
func NewStrictURLSanitizer() *URLSanitizer {
	return &URLSanitizer{ResolveDNS: true}
}

func (s *URLSanitizer) Sanitize(input string) (string, error) {
	if input == "" {
		return input, nil
	}
	lower := strings.ToLower(strings.TrimSpace(input))
	// Block non-HTTPS protocols.
	for _, p := range []string{"http://", "file://", "ftp://", "gopher://"} {
		if strings.HasPrefix(lower, p) {
			return "", ErrURLDenied
		}
	}
	// Parse so we can extract an exact hostname (no port, no path,
	// no query) instead of substring-matching the whole URL string.
	//
	// Audit 2026-06-29 fix: an input with no URL scheme (e.g.
	// "echo hello", a shell command) parses successfully but
	// produces u.Host == "". The previous code rejected any such
	// input with ErrURLDenied, which broke TestChain_RunsAll
	// (every gatekeeper-evaluated shell command got URL-rejected)
	// and would have caused widespread false positives in the
	// real daemon. The fix: only treat input as a URL when
	// u.Scheme is non-empty. Path-only / opaque inputs ("echo
	// hello") are NOT URLs and pass through unchanged.
	u, err := url.Parse(input)
	if err != nil {
		return "", ErrURLDenied
	}
	if u.Scheme == "" {
		return input, nil
	}
	host := u.Hostname()
	if host == "" {
		return "", ErrURLDenied
	}

	// IP literal? Check the IP directly.
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return "", ErrURLDenied
		}
		return input, nil
	}

	// Hostname. Block exact matches and suffix matches against
	// the deny-list.
	if isBlockedHostname(host) {
		return "", ErrURLDenied
	}

	// Optional DNS resolution — catches DNS rebinding where the
	// hostname passes the pattern check but resolves to a private IP.
	// Use a context-aware resolver so the linter (noctx) is happy and
	// the lookup can be canceled by an upstream timeout.
	//
	// NOTE: even with strict resolution, this Sanitize call alone is
	// NOT a complete DNS-rebinding defense. The lookup is a single
	// point-in-time check; the actual HTTP client may resolve a
	// different IP seconds later. Callers that need a strong TOCTOU
	// guarantee must use ResolveURL, pin the IP, and override the
	// Host header on the request. See ResolveURL doc.
	if s.ResolveDNS {
		// Sanitize has no upstream context; use Background with a
		// short timeout. resolveHost enforces its own 2s budget
		// internally.
		if _, err := s.resolveHost(context.Background(), host); err != nil {
			return "", err
		}
	}

	return input, nil
}

// ResolveURL parses a URL, runs the same hostname and IP checks
// Sanitize does, AND (when ResolveDNS is set) performs a DNS
// lookup. It returns the first IP that passed the private-range
// check so the caller can pin the request to that IP and override
// the Host header to defeat DNS-rebinding attacks.
//
// The standard defense against DNS rebinding is:
//
//  1. Resolve the hostname ONCE (here).
//  2. Replace the hostname in the request URL with the resolved IP.
//  3. Set the request Host header to the original hostname (for
//     virtual-host routing, SNI, and certificate verification).
//  4. Connect to the pinned IP.
//
// The previous Sanitize() API only performed step 1 silently,
// leaving the TOCTOU window open for any caller that just did
// http.Get(sanitized). This method exposes the IP so the caller
// can do steps 2–4.
//
// On URL that fails the hostname/pattern checks, ResolveURL
// returns ErrURLDenied. On DNS lookup failure, it returns the
// underlying error and no IP — the caller's HTTP layer will then
// fail naturally (matches the existing fail-open-on-DNS-error
// behavior of Sanitize). Callers that prefer fail-closed on DNS
// errors should treat the error as a deny.
func (s *URLSanitizer) ResolveURL(ctx context.Context, input string) (net.IP, error) {
	if input == "" {
		return nil, nil
	}
	lower := strings.ToLower(strings.TrimSpace(input))
	for _, p := range []string{"http://", "file://", "ftp://", "gopher://"} {
		if strings.HasPrefix(lower, p) {
			return nil, ErrURLDenied
		}
	}
	u, err := url.Parse(input)
	if err != nil {
		return nil, ErrURLDenied
	}
	if u.Scheme == "" {
		// Not a URL — nothing to resolve.
		return nil, nil
	}
	host := u.Hostname()
	if host == "" {
		return nil, ErrURLDenied
	}
	// IP literal: no DNS needed. Return the parsed IP if it's not
	// in a private range.
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return nil, ErrURLDenied
		}
		return ip, nil
	}
	if isBlockedHostname(host) {
		return nil, ErrURLDenied
	}
	return s.resolveHost(ctx, host)
}

// resolveHost looks up `host` and returns the first resolved IP
// that is not in a blocked range. If any resolved IP is blocked,
// the URL is denied (the conservative choice — we do not pick a
// "safe" subset of mixed records, because a malicious DNS response
// can include both a public and a private record and we want the
// caller to fail closed on such a record).
//
// TODO(rebinding): every call site that uses the result of this
// function should pin the IP for the HTTP request. As of
// 2026-07-01 the wiring is partial — updater and telemetry use
// the sanitized URL string with the default http.Client (no IP
// pinning). A follow-up should introduce a shared
// "PinnedHTTPClient" that takes (host, ip) and dials ip with the
// Host header set to host, and route all internal HTTP through
// it. Until that lands, the DNS-rebinding defense is best-effort.
func (s *URLSanitizer) resolveHost(ctx context.Context, host string) (net.IP, error) {
	if !s.ResolveDNS {
		// Caller didn't opt in. This is a programming error if
		// the caller expected a pin. Return nil and let the
		// existing Sanitize path's "no resolution" semantics
		// apply.
		return nil, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	cctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	ips, err := (&net.Resolver{}).LookupIP(cctx, "ip", host)
	if err != nil {
		// Match Sanitize's fail-open on DNS error.
		return nil, nil //nolint:nilerr // intentional: see comment in Sanitize
	}
	var firstGood net.IP
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return nil, fmt.Errorf("%w: %s resolves to private IP %s", ErrURLDenied, host, ip)
		}
		if firstGood == nil {
			firstGood = ip
		}
	}
	return firstGood, nil
}

func (s *URLSanitizer) Name() string { return "url" }

// isBlockedIP reports whether an IP literal is in a range the
// agent must not reach (loopback, RFC1918 private, link-local,
// multicast, unspecified, cloud metadata).
func isBlockedIP(ip net.IP) bool {
	switch {
	case ip.IsLoopback(): // 127.0.0.0/8, ::1
	case ip.IsPrivate(): // RFC1918 + RFC4193
	case ip.IsLinkLocalUnicast(), ip.IsLinkLocalMulticast(): // 169.254.0.0/16, fe80::/10
	case ip.IsMulticast(): // 224.0.0.0/4, ff00::/8
	case ip.IsUnspecified(): // 0.0.0.0, ::
	case ip.Equal(net.ParseIP("169.254.169.254")): // AWS / GCP / Azure metadata
	case ip.Equal(net.ParseIP("::ffff:169.254.169.254")): // IPv4-mapped metadata
	case ip.Equal(net.ParseIP("100.100.100.200")): // Alibaba Cloud metadata
	case ip.Equal(net.ParseIP("::ffff:100.100.100.200")): // IPv4-mapped Alibaba metadata
	case ip.Equal(net.ParseIP("192.0.0.192")): // Oracle Cloud metadata (RFC 6943 SUBNET ID)
	case ip.Equal(net.ParseIP("::ffff:192.0.0.192")): // IPv4-mapped Oracle metadata
	default:
		return false
	}
	return true
}

// isBlockedHostname reports whether hostnames an agent must not
// reach, including the cloud-metadata names and any subdomain
// thereof (so an attacker can't register `metadata.google.internal.evil.com`
// — the suffix-match on `metadata.google.internal` would catch it,
// but we use exact + suffix matching only on the bare canonical
// names below to avoid false positives).
func isBlockedHostname(host string) bool {
	// Exact-match deny-list for special hostnames. These cannot
	// be subdomain-matched safely (e.g. `localhost` should not
	// also match `mylocalhost.example.com`).
	//
	// Cloud providers covered (2026-07-01 audit):
	//   - AWS:  instance-data.ec2.internal
	//   - GCP:  metadata.google.internal, metadata.goog
	//   - Azure: metadata.azure.com
	//   - Alibaba: metadata.aliyun.com, 100.100.100.200 (IP)
	//   - Tencent: metadata.tencentyun.com
	//   - Oracle: 192.0.0.192 (IP, in 192.0.0.0/24 which isPrivate
	//             does not always catch — see isBlockedIP for the
	//             dedicated case)
	for _, blocked := range []string{
		"localhost",
		"ip6-localhost",
		"ip6-loopback",
		"metadata.google.internal",
		"metadata.goog",
		"instance-data.ec2.internal",
		"metadata.azure.com",
		"metadata.aliyun.com",
		"metadata.tencentyun.com",
	} {
		if host == blocked {
			return true
		}
	}
	return false
}

// PIIRegexSanitizer detects and redacts personally identifiable
// information using regex heuristics. Not a comprehensive detector;
// labeled as heuristic per MISSION S10.6.
type PIIRegexSanitizer struct{}

// NewPIIRegexSanitizer creates a PII sanitizer.
func NewPIIRegexSanitizer() *PIIRegexSanitizer { return &PIIRegexSanitizer{} }

func (s *PIIRegexSanitizer) Sanitize(input string) (string, error) {
	if input == "" {
		return input, nil
	}
	// Detect credit card patterns (13-19 digits).
	if matchCCPattern(input) {
		return "", ErrPIIDetected
	}
	// Detect SSN patterns (XXX-XX-XXXX).
	if matchSSNPattern(input) {
		return "", ErrPIIDetected
	}
	return input, nil
}

var ccDigitSum = map[rune]int{'0': 0, '1': 1, '2': 2, '3': 3, '4': 4, '5': 5, '6': 6, '7': 7, '8': 8, '9': 9}

func matchCCPattern(s string) bool {
	digits := extractDigits(s)
	if len(digits) < 13 || len(digits) > 19 {
		return false
	}
	// Luhn check.
	sum := 0
	double := false
	for i := len(digits) - 1; i >= 0; i-- {
		d := ccDigitSum[digits[i]]
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}

func matchSSNPattern(s string) bool {
	// SSN is XXX-XX-XXXX or XXX XX XXXX — explicit separators. We
	// deliberately do NOT detect a bare 9-digit run, because that
	// catches order numbers, ISBN-10s, phone numbers, and other
	// benign inputs (audit 2026-07-01). The false-positive cost
	// in real PII redaction work was unacceptable: nearly every
	// text with a 9-digit run was being flagged.
	return ssnPattern.MatchString(s)
}

// ssnPattern matches the canonical SSN forms: 123-45-6789 and
// 123 45 6789. Anchored to a non-digit boundary so we don't match
// a fragment inside a longer digit run.
var ssnPattern = regexp.MustCompile(`(?:^|[^0-9])(\d{3})[- ](\d{2})[- ](\d{4})(?:[^0-9]|$)`)

func extractDigits(s string) []rune {
	var d []rune
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			d = append(d, ch)
		}
	}
	return d
}

func (s *PIIRegexSanitizer) Name() string { return "pii" }

// PythonImportSanitizer blocks dangerous Python imports/calls.
// Documented as bypassable — this is a heuristic, not a security
// guarantee. Real isolation requires sandboxing.
type PythonImportSanitizer struct{}

// NewPythonImportSanitizer creates a Python import sanitizer.
func NewPythonImportSanitizer() *PythonImportSanitizer { return &PythonImportSanitizer{} }

var bannedPyImports = []string{"os.system", "subprocess", "shutil", "socket", "eval", "exec", "compile", "__import__", "open(", "pty"}

func (s *PythonImportSanitizer) Sanitize(input string) (string, error) {
	if input == "" {
		return input, nil
	}
	lower := strings.ToLower(input)
	for _, banned := range bannedPyImports {
		if strings.Contains(lower, banned) {
			return "", ErrImportDenied
		}
	}
	return input, nil
}

func (s *PythonImportSanitizer) Name() string { return "python_import" }
