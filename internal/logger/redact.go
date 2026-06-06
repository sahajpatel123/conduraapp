package logger

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
)

// redactingHandler wraps another slog.Handler and redacts sensitive values.
//
// It redacts:
//
//  1. Attribute keys that match a known sensitive name (case-insensitive).
//     Examples: "api_key", "Authorization", "password", "secret", "token",
//     "cookie", "private_key", "client_secret", "access_token", "refresh_token",
//     "session_token", "bearer".
//
//  2. Attribute values that look like high-entropy tokens, API keys, or
//     secrets (regex-based detection for common patterns: sk-..., ghp_...,
//     xoxb-..., Bearer ..., JWT-like strings, AWS keys, etc.).
//
//  3. Attribute values whose key contains the substring "secret" or "token"
//     or "password" (even if the full key isn't in the known list).
//
// Redaction replaces the value with "[REDACTED]" by default. Use a custom
// redaction string via RedactWith() if needed.
type redactingHandler struct {
	inner        slog.Handler
	redactString string
	// sensitiveKeys: lower-case substrings that mark a key as sensitive.
	sensitiveKeys []string
	// valuePatterns: regexes for sensitive values regardless of key.
	valuePatterns []*regexp.Regexp
}

// sensitiveKeySet is a closed list of well-known sensitive keys. Keys are
// matched case-insensitively and as exact (whole-key) matches.
var sensitiveKeySet = map[string]struct{}{
	"authorization":       {},
	"proxy-authorization": {},
	"cookie":              {},
	"set-cookie":          {},
	"x-api-key":           {},
	"x-auth-token":        {},
	"api_key":             {},
	"apikey":              {},
	"api-key":             {},
	"access_token":        {},
	"refresh_token":       {},
	"id_token":            {},
	"session_token":       {},
	"bearer":              {},
	"password":            {},
	"passwd":              {},
	"secret":              {},
	"client_secret":       {},
	"private_key":         {},
	"privatekey":          {},
	"encryption_key":      {},
	"encryption-key":      {},
	"signing_key":         {},
	"signing-key":         {},
	"master_key":          {},
	"master-key":          {},
	"ssh_key":             {},
	"ssh-key":             {},
	"oauth_token":         {},
	"oauth-token":         {},
}

// substrings: any key containing one of these (case-insensitive) is sensitive.
var sensitiveKeySubstrings = []string{
	"secret",
	"token",
	"password",
	"passwd",
	"apikey",
	"api_key",
	"private",
	"credential",
	"auth",
}

// valuePatterns detects sensitive values regardless of key.
var valuePatterns = []*regexp.Regexp{
	// OpenAI / Anthropic / Google API keys
	regexp.MustCompile(`\bsk-[A-Za-z0-9_\-]{20,}\b`),
	regexp.MustCompile(`\bsk-ant-[A-Za-z0-9_\-]{20,}\b`),
	regexp.MustCompile(`\bAIza[A-Za-z0-9_\-]{30,}\b`),
	// GitHub
	regexp.MustCompile(`\bghp_[A-Za-z0-9]{30,}\b`),
	regexp.MustCompile(`\bghs_[A-Za-z0-9]{30,}\b`),
	regexp.MustCompile(`\bgho_[A-Za-z0-9]{30,}\b`),
	// Slack
	regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9\-]{10,}\b`),
	// AWS
	regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),
	// JWT
	regexp.MustCompile(`\beyJ[A-Za-z0-9_\-]+\.eyJ[A-Za-z0-9_\-]+\.[A-Za-z0-9_\-]+\b`),
	// Bearer tokens in "Authorization: Bearer xxx"
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._\-]{20,}`),
	// PEM private keys
	regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`),
}

const defaultRedactString = "[REDACTED]"

func newRedactingHandler(inner slog.Handler) *redactingHandler {
	return &redactingHandler{
		inner:         inner,
		redactString:  defaultRedactString,
		sensitiveKeys: sensitiveKeySubstrings,
	}
}

// Enabled delegates to the inner handler.
func (h *redactingHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return h.inner.Enabled(ctx, lvl)
}

// Handle redacts attributes, then delegates to the inner handler.
func (h *redactingHandler) Handle(ctx context.Context, r slog.Record) error {
	// slog.Record has no built-in "map" iteration; copy into a slice and rebuild.
	attrs := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, redactAttr(a, h.sensitiveKeys, valuePatterns, h.redactString))
		return true
	})

	// Rebuild a new record with the same level, time, message, pc, and redacted attrs.
	newRec := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	newRec.AddAttrs(attrs...)
	return h.inner.Handle(ctx, newRec)
}

// WithAttrs redacts the given attrs and returns a new handler.
func (h *redactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	redacted := make([]slog.Attr, 0, len(attrs))
	for _, a := range attrs {
		redacted = append(redacted, redactAttr(a, h.sensitiveKeys, valuePatterns, h.redactString))
	}
	return &redactingHandler{inner: h.inner.WithAttrs(redacted), redactString: h.redactString, sensitiveKeys: h.sensitiveKeys}
}

// WithGroup returns a new handler with the given group.
func (h *redactingHandler) WithGroup(name string) slog.Handler {
	return &redactingHandler{inner: h.inner.WithGroup(name), redactString: h.redactString, sensitiveKeys: h.sensitiveKeys}
}

// -----------------------------------------------------------------------------
// Attribute redaction
// -----------------------------------------------------------------------------

// redactAttr returns a redacted copy of a, recursing into groups.
func redactAttr(a slog.Attr, sensitiveSubstrings []string, valuePatterns []*regexp.Regexp, redactString string) slog.Attr {
	if a.Equal(slog.Attr{}) {
		return a
	}

	// Group: recurse into children.
	if a.Value.Kind() == slog.KindGroup {
		group := a.Value.Group()
		out := make([]slog.Attr, 0, len(group))
		for _, child := range group {
			out = append(out, redactAttr(child, sensitiveSubstrings, valuePatterns, redactString))
		}
		return slog.Group(a.Key, attrsToAny(out)...)
	}

	if isSensitiveKey(a.Key, sensitiveSubstrings) {
		return slog.String(a.Key, redactString)
	}

	// Check the value against patterns (only for string values).
	if a.Value.Kind() == slog.KindString {
		if isSensitiveValue(a.Value.String(), valuePatterns) {
			return slog.String(a.Key, redactString)
		}
	}

	return a
}

// isSensitiveKey returns true if key matches the closed set or contains a
// sensitive substring (case-insensitive).
func isSensitiveKey(key string, sensitiveSubstrings []string) bool {
	lower := strings.ToLower(key)
	if _, ok := sensitiveKeySet[lower]; ok {
		return true
	}
	for _, sub := range sensitiveSubstrings {
		if strings.Contains(lower, sub) {
			return true
		}
	}
	return false
}

// isSensitiveValue returns true if value matches any known secret pattern.
func isSensitiveValue(value string, patterns []*regexp.Regexp) bool {
	for _, p := range patterns {
		if p.MatchString(value) {
			return true
		}
	}
	return false
}

// attrsToAny converts []slog.Attr to []any for slog.Group variadic.
func attrsToAny(attrs []slog.Attr) []any {
	out := make([]any, 0, len(attrs))
	for _, a := range attrs {
		out = append(out, a)
	}
	return out
}
