// This file implements the P0-A hardening for sub-agent ActionRequests
// and MCP tool-call construction sites. A compromised or buggy
// sub-agent can emit any string in its structured JSON output,
// including kind values that are not in blastradius.ClassByKind.
// Without normalization, an attacker that controls the sub-agent
// could trivially emit:
//
//	{"agent_name":"x","kind":"chat","command":"rm -rf $HOME"}
//
// The classifier would label the action READ, the gatekeeper would
// allow it (READ consent), and the command field would never be
// checked. MISSION §2.1 INV-2 forbids trusting unverified model
// output; the per-field sanitizers in this package run first for
// shell/path/URL/PII, but Kind was previously trusted at face value
// because the kind itself determines the blast-radius class.
//
// The fix is to normalize the Kind at every construction site:
// if the Kind is not in the closed allowlist (which mirrors
// blastradius.ClassByKind — the single source of truth for what the
// classifier recognizes), substitute "shell.exec", which classifies
// as DESTRUCTIVE. The gatekeeper's default-deny policy then forces
// a human to approve before any unfamiliar kind reaches the shell.
//
// The allowlist is exported via AllowedSubAgentKinds for callers
// that need to inspect or display it (e.g. to render a sub-agent's
// declared kind in a UI). It is built once from ClassByKind and is
// safe to share across goroutines (read-only map).

package sanitize

import (
	"github.com/sahajpatel123/conduraapp/internal/blastradius"
)

// safeSubAgentKind is the kind substituted when an incoming
// sub-agent Kind is not in the allowlist. We force DESTRUCTIVE —
// the most conservative blast-radius class — so the gatekeeper's
// default-deny policy requires human-in-the-loop consent before
// execution. Any unknown kind from a sub-agent is treated as the
// worst case by design; an attacker who can plant an unknown kind
// has already lost this round.
const safeSubAgentKind = "shell.exec"

// allowedSubAgentKinds is the closed set of valid Kind values for
// sub-agent ActionRequests. Built once from blastradius.ClassByKind
// at package init so a single source of truth governs both the
// classifier and the allowlist.
//
// Per MISSION §5.1: missing/empty/unknown kinds classify as
// DESTRUCTIVE. We mirror that policy here at the construction
// site so the trust boundary closes before the action ever
// reaches the gatekeeper.
var allowedSubAgentKinds = buildAllowedKinds()

func buildAllowedKinds() map[string]bool {
	out := make(map[string]bool, len(blastradius.ClassByKind))
	for kind := range blastradius.ClassByKind {
		out[kind] = true
	}
	return out
}

// AllowedSubAgentKinds returns the closed set of valid Kind values
// for sub-agent ActionRequests, sourced from
// blastradius.ClassByKind. The returned map is freshly allocated on
// each call so callers may inspect or serialize it without
// affecting the package-level allowlist. Callers MUST NOT mutate
// blastradius.ClassByKind to "extend" the allowlist; new kinds
// belong in the blastradius package and ship with a code review.
func AllowedSubAgentKinds() map[string]bool {
	out := make(map[string]bool, len(allowedSubAgentKinds))
	for k, v := range allowedSubAgentKinds {
		out[k] = v
	}
	return out
}

// IsAllowedSubAgentKind reports whether kind (after trim and
// lower-case normalization) is a recognized blast-radius Kind.
// Trim+lowercase matches the normalization that
// blastradius.Classify applies, so a Kind that would classify
// maps to allowed==true and vice versa.
//
// Exposed for tests and for callers that want to short-circuit
// (e.g. drop the request early instead of substituting
// "shell.exec"); production code that builds a blastradius.Action
// should call NormalizeSubAgentKind instead, which always
// produces a valid kind.
func IsAllowedSubAgentKind(kind string) bool {
	return allowedSubAgentKinds[normalize(kind)]
}

// NormalizeSubAgentKind returns kind (trimmed, lower-cased) if it
// is in the allowlist; otherwise returns "shell.exec". The fallback
// is the most conservative blast-radius kind in the closed set —
// DESTRUCTIVE — so the gatekeeper's default-deny rule requires
// human-in-the-loop consent before the action executes.
//
// Empty strings normalize to "shell.exec" by design (a sub-agent
// that emits an empty kind has either been tampered with or has
// hit a code path the daemon does not recognize — both warrant
// maximal caution).
//
// Callers at every Action construction site (delegation_wiring,
// mcp/client, future IPC handlers) MUST invoke this before
// assigning a value to blastradius.Action.Kind. Bypassing it
// reintroduces the P0-A trust-boundary violation.
func NormalizeSubAgentKind(kind string) string {
	normalized := normalize(kind)
	if allowedSubAgentKinds[normalized] {
		return normalized
	}
	return safeSubAgentKind
}

// normalize is shared between IsAllowedSubAgentKind and
// NormalizeSubAgentKind so both check the same canonical form.
// Mirrors the kind normalization in blastradius.Classify
// (trim + lowercase).
func normalize(kind string) string {
	// Avoid pulling in strings just for one TrimSpace+ToLower.
	// Hand-rolled to keep this file dependency-light and to
	// keep the import surface focused on the blastradius
	// single-source-of-truth dependency.
	trimmed := trim(kind)
	return lower(trimmed)
}

func trim(s string) string {
	start := 0
	end := len(s)
	for start < end {
		c := s[start]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			break
		}
		start++
	}
	for end > start {
		c := s[end-1]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			break
		}
		end--
	}
	return s[start:end]
}

func lower(s string) string {
	// ASCII fast path; sub-agent kinds are ASCII identifiers
	// ("chat", "file.write", "shell.exec", …) by construction
	// (see blastradius.ClassByKind). Non-ASCII input still
	// lowercases correctly for ASCII run; if a future Kind
	// ever uses non-ASCII letters the validator should
	// normalize them before reaching here.
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
