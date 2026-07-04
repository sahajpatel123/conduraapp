package sanitize

import (
	"strings"
	"testing"
)

// TestShellSanitizer_F01BypassPayloads documents the sanitizer's
// behavior on the F-01 audit findings. The sanitizer is a
// metacharacter + binary-allowlist gate. It is NOT a policy engine
// for argument-level intent; payloads like `find . -exec rm {} \;`
// have no shell metachars and the bin is allowed, so they pass
// through. The executor (internal/executor/executor.go:279) is
// responsible for rejecting these by policy — those tests live in
// the executor package.
//
// This file locks in the sanitizer's contract for the F-01 payloads
// so any future change to isShellMetachar is caught here.
func TestShellSanitizer_F01BypassPayloads(t *testing.T) {
	s := NewShellSanitizer(nil) // default allowlist: git, ls, cat, echo, find, grep, head, tail, sort, uniq, wc
	tests := []struct {
		name         string
		input        string
		expectError  bool
		whyItMatters string
	}{
		{
			// F-01: find -exec with rm payload + escaped semicolon.
			// The escaped `\;` tokenizes to `\;` which contains `;`,
			// triggering isShellMetachar. The sanitizer CORRECTLY
			// rejects this even though the bin (find) is allowed.
			// Defense in depth: the executor can still be tricked
			// by `find -exec rm {} +` (no semicolon terminator)
			// — that case is an executor-level test.
			name:         "find-exec-rm-semicolon-blocked",
			input:        "find . -exec rm {} \\;",
			expectError:  true,
			whyItMatters: "escaped ; trips isShellMetachar; sanitizer is doing its job here",
		},
		{
			// F-01: find -ok with interactive prompt + escaped semicolon.
			// Same as above.
			name:         "find-ok-confirm-semicolon-blocked",
			input:        "find . -ok rm {} \\;",
			expectError:  true,
			whyItMatters: "escaped ; trips isShellMetachar",
		},
		{
			// F-01: find -exec without semicolon (the actual bypass
			// the audit flagged — find accepts `+` as terminator).
			// The sanitizer ALLOWS this because the bin is `find`
			// and no token contains a metachar. The executor must
			// reject by policy.
			name:         "find-exec-plus-terminator-passes-sanitizer",
			input:        "find . -exec rm {} +",
			expectError:  false,
			whyItMatters: "executor is the policy boundary; sanitizer is the metachar gate",
		},
		{
			// F-01: env var EDITOR override.
			name:         "env-editor-override",
			input:        "VISUAL=vim git commit",
			expectError:  true,
			whyItMatters: "bin 'VISUAL=vim' is not in the allowlist",
		},
		{
			// Direct sh -c invocation (defense in depth).
			name:         "sh-c-blocked",
			input:        "sh -c 'rm -rf /'",
			expectError:  true,
			whyItMatters: "bin 'sh' is not in the default allowlist",
		},
		{
			// Direct bash -c invocation.
			name:         "bash-c-blocked",
			input:        "bash -c 'echo pwned'",
			expectError:  true,
			whyItMatters: "bin 'bash' is not in the default allowlist",
		},
		{
			// Git with --upload-pack= override (CVE-2017-1000117 class).
			name:         "git-upload-pack-override",
			input:        "git clone foo --upload-pack=sh",
			expectError:  false,
			whyItMatters: "bin 'git' is allowed; --upload-pack= is not a shell metachar; executor must reject",
		},
		{
			// git config core.editor injection.
			name:         "git-core-editor-injection",
			input:        "git -c core.editor=sh commit",
			expectError:  false,
			whyItMatters: "bin 'git' allowed; no metachars; executor must reject -c core.editor= overrides",
		},
		{
			// Argument that looks like an option but contains a metachar.
			name:         "option-with-metachar",
			input:        "ls -la>output",
			expectError:  true,
			whyItMatters: "the > in '-la>output' triggers isShellMetachar",
		},
		{
			// Backtick in the middle of an arg.
			name:         "arg-with-backtick",
			input:        "ls `whoami`",
			expectError:  true,
			whyItMatters: "backtick is a shell metachar, even mid-arg",
		},
		{
			// $() command substitution.
			name:         "dollar-paren-substitution",
			input:        "ls $(whoami)",
			expectError:  true,
			whyItMatters: "$( is a shell metachar, caught by isShellMetachar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := s.Sanitize(tt.input)
			gotErr := err != nil
			if gotErr != tt.expectError {
				t.Errorf("Sanitize(%q) error = %v, want %v (out=%q) — %s",
					tt.input, err, tt.expectError, out, tt.whyItMatters)
			}
			if !gotErr {
				// Sanitizer returned the input unchanged. Verify
				// the output equals the input (sanitizer doesn't
				// rewrite).
				if out != tt.input {
					t.Errorf("Sanitize(%q) returned %q, want input unchanged", tt.input, out)
				}
			}
		})
	}
}

// TestIsShellMetachar_Exhaustive locks the metacharacter list so
// that adding a new dangerous pattern requires updating this test.
// This is the F-01 audit followup: the metachar list is what
// prevents shell injection, and adding/removing entries must be
// intentional.
func TestIsShellMetachar_Exhaustive(t *testing.T) {
	// These MUST be detected.
	dangerous := []string{
		"|",        // pipe
		">",        // redirect out
		"<",        // redirect in
		"`",        // backtick command substitution
		";",        // command separator
		"&&",       // and
		"||",       // or
		"$(",       // dollar-paren command substitution
		"${",       // dollar-brace variable expansion
		"&>",       // bash redirect both
		"ls|rm",    // metachar in arg
		"x>y",      // metachar in arg
		"`whoami`", // backticks in arg
	}
	for _, d := range dangerous {
		if !isShellMetachar(d) {
			t.Errorf("isShellMetachar(%q) = false, want true", d)
		}
	}

	// These MUST NOT be detected as metachars.
	// (The sanitizer relies on the absence of metachars to allow
	// these; if any become "metachars" we lose legitimate commands.)
	safe := []string{
		"find",  // command name
		"-exec", // find option
		"-ok",   // find option
		"--",    // end-of-options marker
		"{}",
		"*.go", // glob — NOT a shell metachar (the shell expands
		// it, but the sanitizer receives the literal arg
		// because it never invokes a shell)
		"-la",              // ls option
		"--upload-pack=sh", // git option with =value
		"-c",
		"core.editor=sh",
		"+", // find -exec + terminator (no metachar)
	}
	for _, s := range safe {
		if isShellMetachar(s) {
			t.Errorf("isShellMetachar(%q) = true, want false", s)
		}
	}
}

// TestShellSanitizer_DefaultAllowlist documents the default
// allowlist so any addition/removal is intentional. The F-01 audit
// noted that sh/bash are NOT in the allowlist; this test pins
// that behavior.
func TestShellSanitizer_DefaultAllowlist(t *testing.T) {
	s := NewShellSanitizer(nil)
	tests := []struct {
		bin     string
		allowed bool
	}{
		{"git", true},
		{"ls", true},
		{"cat", true},
		{"echo", true},
		{"find", true},
		{"grep", true},
		{"head", true},
		{"tail", true},
		{"sort", true},
		{"uniq", true},
		{"wc", true},
		// F-01 dangerous bins: must remain absent from default.
		{"sh", false},
		{"bash", false},
		{"zsh", false},
		{"rm", false},   // rm is destructive; not in default
		{"curl", false}, // network egress; not in default
		{"wget", false},
		{"nc", false}, // arbitrary network
		{"sudo", false},
	}
	for _, tt := range tests {
		t.Run(tt.bin, func(t *testing.T) {
			_, err := s.Sanitize(tt.bin + " foo")
			gotAllowed := err == nil
			if gotAllowed != tt.allowed {
				reason := ""
				if !tt.allowed {
					reason = " (F-01: must remain outside default allowlist)"
				}
				t.Errorf("Sanitize(%q) allowed = %v, want %v%s",
					tt.bin, gotAllowed, tt.allowed, reason)
			}
		})
	}
}

// helper to silence the unused import warning if strings is not
// otherwise referenced.
var _ = strings.Contains
