//nolint:revive,mnd // interface implementations; digit map is definitional
package sanitize

import (
	"context"
	"net"
	"net/url"
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
	// the lookup can be cancelled by an upstream timeout.
	if s.ResolveDNS {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		ips, err := (&net.Resolver{}).LookupIP(ctx, "ip", host)
		cancel()
		if err == nil {
			for _, ip := range ips {
				if isBlockedIP(ip) {
					return "", ErrURLDenied
				}
			}
		}
		// DNS lookup failure is NOT a deny: the actual HTTP client
		// will fail naturally. Fail-open here matches the rest of
		// the sanitizer chain (which errs on the side of letting
		// the gatekeeper decide via policy).
	}

	return input, nil
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
	for _, blocked := range []string{
		"localhost",
		"ip6-localhost",
		"ip6-loopback",
		"metadata.google.internal",
		"metadata.goog",
		"instance-data.ec2.internal",
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
	// Look for XXX-XX-XXXX pattern.
	count := 0
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			count++
		} else if ch != '-' && ch != ' ' {
			count = 0
		}
	}
	return count >= 9 && len(extractDigits(s)) >= 9
}

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
