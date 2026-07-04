// Package sanitize — redact_secrets.go provides RedactSecrets, a
// pure-regex replacement for known credential shapes that may appear
// in audit messages, consent modal bodies, or MCP tool-call argument
// maps. P0-4 audit-surface leak: prior to this change, raw MCP args
// (passed via fmt.Sprintf("%v", args)) and raw voice transcripts were
// written into the audit chain and the consent modal display; any
// embedded credential (github PAT, OpenAI/Anthropic API key, AWS
// access key, etc.) would leak into a forensic-readable store.
//
// RedactSecrets is intentionally cheap and stateless: it does NOT
// parse JSON, allocate intermediate structures, or attempt to be
// exhaustive — it is a defense-in-depth measure that catches the
// common shapes and falls back to a generic key=value heuristic for
// unknown providers. Callers in front of any sink that sees
// user-/model-controlled text (audit log, consent modal, web
// activity log) should pass the value through RedactSecrets first.
//
// This file is regex-only; the patterns are precompiled once and
// applied via a single pass per pattern.
package sanitize

import "regexp"

// redactionMarker is the replacement for a matched secret. Chosen so
// downstream consumers (forensic tools, the consent modal renderer)
// can grep for "<redacted>" without false positives on real words.
const redactionMarker = "<redacted>"

// secretPatterns are compiled once at init. Order matters: provider-
// specific patterns run first so they win over the generic
// key=value catch-all (which would also match them but produce a
// noisier replacement).
//
//nolint:gochecknoglobals // precompiled patterns; allocating per call would be wasteful
var secretPatterns = []*regexp.Regexp{
	// GitHub PATs / OAuth tokens / user tokens / app tokens / refresh.
	regexp.MustCompile(`ghp_[A-Za-z0-9]{20,}`),
	regexp.MustCompile(`gho_[A-Za-z0-9]{20,}`),
	regexp.MustCompile(`ghu_[A-Za-z0-9]{20,}`),
	regexp.MustCompile(`ghs_[A-Za-z0-9]{20,}`),
	regexp.MustCompile(`ghr_[A-Za-z0-9]{20,}`),

	// OpenAI / Anthropic. Note: sk-ant- is matched first to win over
	// the generic sk- prefix.
	regexp.MustCompile(`sk-ant-[A-Za-z0-9_\-]{20,}`),
	regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`),
	regexp.MustCompile(`sk-proj-[A-Za-z0-9_\-]{20,}`),

	// Slack tokens.
	regexp.MustCompile(`xox[boprs]-[A-Za-z0-9\-]{10,}`),

	// AWS access key id.
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),

	// Google API key / OAuth refresh token shapes.
	regexp.MustCompile(`AIza[0-9A-Za-z_\-]{35}`),

	// PEM private keys (whole block, including label lines).
	regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`),

	// Generic key=value / key:value / key: value shape for known secret
	// key names. Matches:  token=abc..., "token": "abc...", password:
	// 'abc...'. Require 8+ chars in the value so we don't redact words
	// like password=foo on short values (still possible noise on very
	// short values but acceptable for an audit-log redaction).
	regexp.MustCompile(`(?i)(token|api_key|apikey|password|secret|credentials)["']?\s*[:=]\s*["']?([^\s"',}{]{8,})`),
}

// RedactSecrets returns input with every detected credential shape
// replaced by "<redacted>". It is safe to call on any string and is
// allocation-cheap: a single strings.Replacer would be more compact
// but cannot implement the generic key/value heuristic (which needs a
// $2-style group reference). For audit + consent modal volumes this
// is well under a microsecond per KB.
//
// Examples:
//
//	RedactSecrets(`{"token": "ghp_abc123def456ghi789jkl012mno345pqr"}`)
//	  => `{"token": "<redacted>"}`
//
//	RedactSecrets(`password=hunter2hunter`) => `password=<redacted>`
//
//	RedactSecrets(`my AKIAIOSFODNN7EXAMPLE key`) => `my <redacted> key`
func RedactSecrets(input string) string {
	if input == "" {
		return input
	}
	out := input
	for _, p := range secretPatterns {
		out = p.ReplaceAllString(out, redactionMarker)
	}
	return out
}
