package gatekeeper

import (
	"errors"
	"strings"
	"testing"
)

// P0-B (CLAUDE.md §2.1 invariant #3): LoadPolicy must reject any rule
// that downgrades a DESTRUCTIVE-class action to allow. This is the
// hard guarantee the Survival Rule depends on; without it, a user
// who edits policy.yaml could turn "every click is dangerous" into
// "every click is fine." See LOGBOOK entry "P0-B: reject
// destructive→allow in user policy".

func TestLoadPolicy_RejectsDestructiveAllow(t *testing.T) {
	yaml := `version: "1"
rules:
  - match: { class: destructive }
    decide: allow
`
	_, err := LoadPolicy([]byte(yaml))
	if err == nil {
		t.Fatal("LoadPolicy accepted destructive→allow; expected *PolicySchemaError")
	}
	var schemaErr *PolicySchemaError
	if !errors.As(err, &schemaErr) {
		t.Fatalf("error type = %T (%v); want *PolicySchemaError", err, err)
	}
	if schemaErr.Reason != "destructive_downgraded_to_allow" {
		t.Errorf("Reason = %q; want destructive_downgraded_to_allow", schemaErr.Reason)
	}
	if schemaErr.RuleIndex != 1 {
		t.Errorf("RuleIndex = %d; want 1", schemaErr.RuleIndex)
	}
	if !strings.Contains(schemaErr.Detail, "§2.1") {
		t.Errorf("Detail %q must reference spec §2.1 invariant #3", schemaErr.Detail)
	}
}

func TestLoadPolicy_AcceptsDestructiveWithPresenceAndConsent(t *testing.T) {
	// The legitimate way to allow DESTRUCTIVE: require an active human at
	// the keyboard + a confirming native modal. This must NOT be
	// rejected.
	yaml := `version: "1"
rules:
  - match: { class: destructive }
    decide: require_presence_and_consent
    consent:
      type: native_dialog
      timeout_seconds: 300
      on_timeout: deny
      require_user_active: true
      on_user_absent: queue
`
	p, err := LoadPolicy([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadPolicy rejected legitimate destructive consent rule: %v", err)
	}
	if p == nil || len(p.rules) != 1 {
		t.Fatalf("p = %+v; want 1 rule loaded", p)
	}
}

func TestLoadPolicy_AcceptsNonDestructiveAllow(t *testing.T) {
	// WRITE→allow is a perfectly reasonable user choice. The validator
	// must NOT reject downgrades of non-DESTRUCTIVE classes.
	yaml := `version: "1"
rules:
  - match: { class: write }
    decide: allow
`
	p, err := LoadPolicy([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadPolicy rejected write→allow (non-DESTRUCTIVE downgrade is user choice): %v", err)
	}
	if len(p.rules) != 1 {
		t.Fatalf("p.rules = %d entries; want 1", len(p.rules))
	}
}

func TestLoadPolicy_RejectsDestructiveQueue(t *testing.T) {
	// Silently queueing a DESTRUCTIVE action is a downgrade — the user
	// never sees a native modal and the action eventually executes
	// unattended. Only "require_consent" / "require_presence_and_consent"
	// / "deny" are legal for DESTRUCTIVE matches.
	yaml := `version: "1"
rules:
  - match: { class: destructive }
    decide: queue
`
	_, err := LoadPolicy([]byte(yaml))
	if err == nil {
		t.Fatal("LoadPolicy accepted destructive→queue; expected *PolicySchemaError")
	}
	var schemaErr *PolicySchemaError
	if !errors.As(err, &schemaErr) {
		t.Fatalf("error type = %T (%v); want *PolicySchemaError", err, err)
	}
	if schemaErr.Reason != "destructive_silently_queued" {
		t.Errorf("Reason = %q; want destructive_silently_queued", schemaErr.Reason)
	}
}

// A "any" wildcard must be treated as covering DESTRUCTIVE for the
// invariant — a permissive rule on any class is just as bad as one
// explicitly targeting DESTRUCTIVE.
func TestLoadPolicy_RejectsAnyClassAllow(t *testing.T) {
	yaml := `version: "1"
rules:
  - match: { class: any }
    decide: allow
`
	_, err := LoadPolicy([]byte(yaml))
	if err == nil {
		t.Fatal("LoadPolicy accepted any→allow; expected *PolicySchemaError")
	}
	var schemaErr *PolicySchemaError
	if !errors.As(err, &schemaErr) {
		t.Fatalf("error type = %T (%v); want *PolicySchemaError", err, err)
	}
}

// Comma-separated class list — "read,write,destructive" with decide
// allow must still be flagged on the destructive token.
func TestLoadPolicy_RejectsClassListWithDestructiveAllow(t *testing.T) {
	yaml := `version: "1"
rules:
  - match: { class: "read, write, destructive" }
    decide: allow
`
	_, err := LoadPolicy([]byte(yaml))
	if err == nil {
		t.Fatal("LoadPolicy accepted class-list containing destructive→allow; expected *PolicySchemaError")
	}
	var schemaErr *PolicySchemaError
	if !errors.As(err, &schemaErr) {
		t.Fatalf("error type = %T (%v); want *PolicySchemaError", err, err)
	}
}

// Pure destructive deny must be allowed (overriding to deny more is safe).
func TestLoadPolicy_AcceptsDestructiveDeny(t *testing.T) {
	yaml := `version: "1"
rules:
  - match: { class: destructive }
    decide: deny
`
	if _, err := LoadPolicy([]byte(yaml)); err != nil {
		t.Fatalf("LoadPolicy rejected destructive→deny: %v", err)
	}
}

// errors.Is must compare on Reason so wrapped schema errors match.
func TestPolicySchemaError_Is(t *testing.T) {
	a := &PolicySchemaError{Reason: "destructive_downgraded_to_allow", RuleIndex: 1, Detail: "x"}
	b := &PolicySchemaError{Reason: "destructive_downgraded_to_allow", RuleIndex: 99, Detail: "y"}
	if !errors.Is(a, b) {
		t.Error("errors.Is should match on Reason")
	}
	c := &PolicySchemaError{Reason: "destructive_silently_queued"}
	if errors.Is(a, c) {
		t.Error("errors.Is must NOT match across different Reasons")
	}
}
