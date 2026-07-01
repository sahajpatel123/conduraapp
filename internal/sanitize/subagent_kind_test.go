package sanitize

import (
	"strings"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/blastradius"
)

// TestAllowedSubAgentKinds_MirrorsClassByKind pins the contract
// that the allowlist is the closed set recognized by the
// classifier. If a future change adds a kind to ClassByKind, the
// allowlist grows automatically; if a change removes one, the
// allowlist shrinks automatically. Either way the two stay in
// sync — otherwise the P0-A trust boundary re-opens.
func TestAllowedSubAgentKinds_MirrorsClassByKind(t *testing.T) {
	got := AllowedSubAgentKinds()
	if len(got) != len(blastradius.ClassByKind) {
		t.Fatalf("allowlist size %d != ClassByKind size %d; allowlist must mirror classifier allowlist",
			len(got), len(blastradius.ClassByKind))
	}
	for kind := range blastradius.ClassByKind {
		if !got[kind] {
			t.Errorf("ClassByKind has %q but allowlist missing it — P0-A trust boundary broken", kind)
		}
	}
	for kind := range got {
		if _, ok := blastradius.ClassByKind[kind]; !ok {
			t.Errorf("allowlist has %q but ClassByKind missing it — P0-A trust boundary broken", kind)
		}
	}
}

// TestAllowedSubAgentKinds_ReturnsCopy ensures the returned map is
// a fresh allocation so callers can mutate locally without
// affecting the package-level state.
func TestAllowedSubAgentKinds_ReturnsCopy(t *testing.T) {
	first := AllowedSubAgentKinds()
	first["wibble.frobnicate"] = true
	second := AllowedSubAgentKinds()
	if second["wibble.frobnicate"] {
		t.Fatal("mutating one allowlist result leaked into another — must return a fresh copy")
	}
}

// TestNormalizeSubAgentKind_KnownPassesThrough verifies that every
// kind recognized by the classifier passes through unchanged.
// This is the positive-direction contract — a legitimate
// sub-agent emitting a known kind must not be rewritten, or we
// break audit/log fidelity.
func TestNormalizeSubAgentKind_KnownPassesThrough(t *testing.T) {
	// Sample one kind from each blast-radius class to make
	// sure all four buckets work.
	known := map[blastradius.Class]string{
		blastradius.READ:        "chat",
		blastradius.WRITE:       "file.write",
		blastradius.NETWORK:     "http.request",
		blastradius.DESTRUCTIVE: "shell.exec",
	}
	for wantClass, kind := range known {
		got := NormalizeSubAgentKind(kind)
		if got != kind {
			t.Errorf("known kind %q (class %s) rewritten to %q; must pass through unchanged",
				kind, wantClass, got)
		}
		if gotClass := blastradius.Classify(blastradius.Action{Kind: got}); gotClass != wantClass {
			t.Errorf("normalize(%q) produced kind %q that classifies as %s, want %s",
				kind, got, gotClass, wantClass)
		}
	}
}

// TestNormalizeSubAgentKind_UnknownForcesShellExec is the
// core P0-A test. A sub-agent that emits a kind not in the
// allowlist must be normalized to "shell.exec" so the classifier
// labels it DESTRUCTIVE and the gatekeeper's default-deny rule
// forces human-in-the-loop consent.
func TestNormalizeSubAgentKind_UnknownForcesShellExec(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"clearly-unknown-bareword", "wibble.frobnicate"},
		{"mimic-existing-namespace", "shell.exec.cmd"}, // dot-separated prefix-bait
		{"shell-like-but-wrong", "Shell.Exec"},         // case tampered — passes through after lower? verify
		{"uppercase prefix", "Shell.exec"},
		{"unicode bait", "🦀.exec"},
		{"control chars", "shell.exec\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeSubAgentKind(tc.in)
			if got != "shell.exec" {
				t.Errorf("normalize(%q) = %q, want %q (P0-A: unknown kinds must always force DESTRUCTIVE)",
					tc.in, got, "shell.exec")
			}
			if cls := blastradius.Classify(blastradius.Action{Kind: got}); cls != blastradius.DESTRUCTIVE {
				t.Errorf("normalized kind %q classified as %s, want DESTRUCTIVE", got, cls)
			}
		})
	}
}

// TestNormalizeSubAgentKind_EmptyForcesShellExec — an empty
// kind from a sub-agent means tampered input or an unknown
// code path. Either way, default to DESTRUCTIVE.
func TestNormalizeSubAgentKind_EmptyForcesShellExec(t *testing.T) {
	got := NormalizeSubAgentKind("")
	if got != "shell.exec" {
		t.Errorf("normalize(\"\") = %q, want %q", got, "shell.exec")
	}
}

// TestNormalizeSubAgentKind_WhitespaceForcesShellExec — a
// sub-agent emitting whitespace-only kind has either been
// tampered with or hit a code path the daemon does not
// recognize. Force DESTRUCTIVE rather than letting it leak
// through as a Classify-default DESTRUCTIVE later — we want
// the same shape of normalized kind for every "unknown"
// outcome so audit/UI logic has one branch.
func TestNormalizeSubAgentKind_WhitespaceForcesShellExec(t *testing.T) {
	for _, in := range []string{" ", "\t", "\n", "  \r\n  ", "\t\t"} {
		got := NormalizeSubAgentKind(in)
		if got != "shell.exec" {
			t.Errorf("normalize(%q) = %q, want %q", in, got, "shell.exec")
		}
	}
}

// TestNormalizeSubAgentKind_MaliciousChatWithDestructiveBody
// pins the threat that motivated P0-A. The audit identified a
// trust-boundary gap at internal/daemon/delegation_wiring.go:410,440:
// the sub-agent's Kind flowed straight into
// blastradius.Action.Kind without verification. The classifier
// correctly denies unknown kinds by default (DESTRUCTIVE), so
// the security outcome was coincidentally correct — but the
// row-level Kind field, the audit log, and the gatekeeper
// payload would all carry the raw attacker string, leaving an
// obvious injection point if the classifier default ever
// changed or a parser regression allowed an attacker string
// through the otherwise-correct default-deny path.
//
// The test simulates the canonical attack: an unknown, name-
// collided kind (a maliciously-named dot-separated string) with
// a destructive Body. With P0-A normalization in place the kind
// is rewritten to "shell.exec" at the construction site so:
//   - The classifier labels the action DESTRUCTIVE via the
//     canonical kind, not via incidental default-deny;
//   - The row-level Kind field is the canonical one;
//   - The audit log records the canonical kind;
//   - The body is preserved for the per-field sanitizers
//     (shell sanitizer, etc.) to act on.
func TestNormalizeSubAgentKind_MaliciousChatWithDestructiveBody(t *testing.T) {
	// Attacker tries to mimic "chat" with extra fields after the
	// dot so it looks permissive at first glance but is not in
	// the closed allowlist.
	const maliciousKind = "chat.payload_run"
	const maliciousBody = "rm -rf $HOME"

	got := NormalizeSubAgentKind(maliciousKind)
	if got != "shell.exec" {
		t.Fatalf("normalize(%q) = %q; an unknown name-collided kind must NOT pass through as %q",
			maliciousKind, got, maliciousKind)
	}

	ba := blastradius.Action{Kind: got, Body: maliciousBody}
	if cls := blastradius.Classify(ba); cls != blastradius.DESTRUCTIVE {
		t.Fatalf("post-normalize classify = %s, want DESTRUCTIVE — that is the property P0-A exists to enforce",
			cls)
	}
}

// TestNormalizeSubAgentKind_NormalizationIsTrimAndLower mirrors
// blastradius.Classify's normalization so the allowlist check
// matches the classifier check. Otherwise an attacker could
// bypass the allowlist by emitting "Chat" with a capital C,
// which Classify would normalize to "chat" (allowed) while our
// pre-check might not.
func TestNormalizeSubAgentKind_NormalizationIsTrimAndLower(t *testing.T) {
	// "  CHAT  " should be normalized to "chat" before the
	// allowlist lookup. If our normalization diverges from
	// the classifier's, the two will disagree about what's
	// allowed — opening a bypass.
	got := IsAllowedSubAgentKind("  CHAT  ")
	if !got {
		t.Fatalf("kind (with case + whitespace) should normalize to a known kind and pass allowlist")
	}
	got2 := NormalizeSubAgentKind("  CHAT  ")
	if got2 != "chat" {
		t.Errorf("normalize(\"  CHAT  \") = %q, want %q (trim + lowercase must mirror blastradius.Classify)",
			got2, "chat")
	}
}

// TestNormalizeSubAgentKind_FullyExercised — sweep every kind in
// the allowlist to confirm none accidentally fall through to
// "shell.exec". A bug here would silently rewrite legitimate
// sub-agent actions to the most-conservative class.
func TestNormalizeSubAgentKind_FullyExercised(t *testing.T) {
	for kind := range AllowedSubAgentKinds() {
		got := NormalizeSubAgentKind(kind)
		if got != kind {
			t.Errorf("known kind %q was rewritten to %q; must pass through unchanged", kind, got)
		}
	}
}

// TestIsAllowedSubAgentKind — basic contract: known kinds are
// allowed, unknown kinds are not.
func TestIsAllowedSubAgentKind(t *testing.T) {
	if !IsAllowedSubAgentKind("shell.exec") {
		t.Error("shell.exec must be allowed")
	}
	if !IsAllowedSubAgentKind("file.write") {
		t.Error("file.write must be allowed")
	}
	if IsAllowedSubAgentKind("totally.bogus") {
		t.Error("totally.bogus must NOT be allowed")
	}
	if IsAllowedSubAgentKind("") {
		t.Error("empty kind must NOT be allowed")
	}
}

// compile-time guard: ensure strings import is referenced so
// future maintainers don't accidentally drop the import when
// simplifying this file (it is used by the comment that names
// the test category above).
var _ = strings.TrimSpace
