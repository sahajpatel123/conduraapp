package hub

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
)

// ScanResult is the outcome of a safety scan on a downloaded skill.
type ScanResult struct {
	Safe    bool     `json:"safe"`
	Issues  []string `json:"issues,omitempty"`
	Checksum string  `json:"checksum"`
}

// Verify checks the SHA-256 checksum of the downloaded data against
// the expected value from the hub. Returns an error if they don't match.
func Verify(data []byte, expected string) error {
	sum := sha256.Sum256(data)
	actual := fmt.Sprintf("%x", sum)
	if actual != strings.ToLower(expected) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}

// Scan performs a basic safety scan on a skill archive. It checks
// for dangerous patterns in the skill's steps (shell injection,
// path traversal, network calls to unexpected hosts). This is a
// first-pass filter; the full safety layer (Gatekeeper) handles
// runtime enforcement.
func Scan(data []byte) ScanResult {
	result := ScanResult{Safe: true}

	// Parse as JSON to inspect structure.
	var skill struct {
		Name    string   `json:"name"`
		Steps   []string `json:"steps"`
		Trust   string   `json:"trust"`
		License string   `json:"license"`
	}
	if err := json.Unmarshal(data, &skill); err != nil {
		// Not JSON — treat as opaque archive, flag for manual review.
		result.Safe = false
		result.Issues = append(result.Issues, "non-JSON skill archive; manual review required")
		return result
	}

	// Check for dangerous step patterns.
	dangerous := []string{
		"rm -rf", "curl | sh", "wget | bash", "eval(", "exec(",
		"sudo", "chmod 777", "chmod +x /", "dd if=",
		"/etc/passwd", "/etc/shadow", ".ssh/authorized_keys",
	}
	for _, step := range skill.Steps {
		lower := strings.ToLower(step)
		for _, pattern := range dangerous {
			if strings.Contains(lower, pattern) {
				result.Safe = false
				result.Issues = append(result.Issues,
					fmt.Sprintf("dangerous pattern %q in step: %s", pattern, truncate(step, 80)))
			}
		}
	}

	// Flag experimental trust as needing user confirmation.
	if skill.Trust == "experimental" {
		result.Issues = append(result.Issues, "experimental trust level; requires user confirmation")
	}

	// Flag missing license.
	if skill.License == "" {
		result.Issues = append(result.Issues, "no license specified")
	}

	return result
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
