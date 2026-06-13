// Package description is in sanitize.go.
//
//nolint:revive // interface implementations
package sanitize

import (
	"strings"
)

// ShellSanitizer enforces a binary allowlist and blocks shell
// metacharacters (pipes, redirects, backticks, command substitution).
type ShellSanitizer struct {
	allowed []string
}

// NewShellSanitizer creates a shell sanitizer with the given binary allowlist.
func NewShellSanitizer(allowed []string) *ShellSanitizer {
	if len(allowed) == 0 {
		allowed = []string{"git", "ls", "cat", "echo", "find", "grep", "head", "tail", "sort", "uniq", "wc"}
	}
	return &ShellSanitizer{allowed: allowed}
}

// Sanitize validates a shell command string.
func (s *ShellSanitizer) Sanitize(input string) (string, error) {
	if input == "" {
		return input, nil
	}
	tokens := strings.Fields(input)
	if len(tokens) == 0 {
		return input, nil
	}
	bin := tokens[0]
	if !s.isAllowed(bin) {
		return "", ErrShellDenied
	}
	for _, tok := range tokens {
		if isShellMetachar(tok) {
			return "", ErrShellDenied
		}
	}
	return input, nil
}

func (s *ShellSanitizer) Name() string { return "shell" }

func (s *ShellSanitizer) isAllowed(bin string) bool {
	for _, a := range s.allowed {
		if bin == a {
			return true
		}
	}
	return false
}

func isShellMetachar(s string) bool {
	dangerous := []string{"|", ">", "<", "`", ";", "&&", "||", "$(", "${", "&>"}
	for _, d := range dangerous {
		if s == d || strings.Contains(s, d) {
			return true
		}
	}
	return false
}
