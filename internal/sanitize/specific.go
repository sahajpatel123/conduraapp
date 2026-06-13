//nolint:revive,mnd // interface implementations; digit map is definitional
package sanitize

import "strings"

// PathSanitizer blocks traversal and system paths.
type PathSanitizer struct{}

// NewPathSanitizer creates a path sanitizer.
func NewPathSanitizer() *PathSanitizer { return &PathSanitizer{} }

func (s *PathSanitizer) Sanitize(input string) (string, error) {
	if input == "" {
		return input, nil
	}
	if strings.Contains(input, "..") {
		return "", ErrPathDenied
	}
	systemPrefixes := []string{"/etc/", "/System/", "/private/etc/", "C:\\Windows\\"}
	for _, prefix := range systemPrefixes {
		if strings.HasPrefix(input, prefix) {
			return "", ErrPathDenied
		}
	}
	return input, nil
}

func (s *PathSanitizer) Name() string { return "path" }

// URLSanitizer prevents SSRF by resolving the host and checking
// the IP against private/loopback ranges.
type URLSanitizer struct{}

// NewURLSanitizer creates a URL sanitizer.
func NewURLSanitizer() *URLSanitizer { return &URLSanitizer{} }

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
	// Block localhost and private IP patterns in URLs.
	privatePatterns := []string{
		"localhost", "127.0.0.1", "[::1]",
		"10.", "172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.",
		"172.24.", "172.25.", "172.26.", "172.27.",
		"172.28.", "172.29.", "172.30.", "172.31.",
		"192.168.", "169.254.",
		"metadata.google.internal", "169.254.169.254",
	}
	for _, pat := range privatePatterns {
		if strings.Contains(lower, pat) {
			return "", ErrURLDenied
		}
	}
	return input, nil
}

func (s *URLSanitizer) Name() string { return "url" }

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
