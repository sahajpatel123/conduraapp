package gatekeeper

import (
	"context"
	"strings"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/blastradius"
)

// fakePresence is a PresenceChecker stub for testing the N1 gate.
type fakePresence struct{ present bool }

func (f *fakePresence) IsPresent() bool { return f.present }

// approveConsent always approves, so the only thing that can deny a
// consent-required action is the presence gate (or halt, which is nil here).
type approveConsent struct{}

func (approveConsent) Show(context.Context, *ConsentTicket) (bool, error) { return true, nil }
func (approveConsent) IsAvailable() bool                                  { return true }

// TestPresence_AbsentDeniesDestructive: an absent user must not be
// able to drive a DESTRUCTIVE action through to the consent modal.
// The N1 gate denies immediately (action held for safety).
func TestPresence_AbsentDeniesDestructive(t *testing.T) {
	e := NewEngine(DefaultPolicy(), approveConsent{}, nil)
	e.SetPresenceChecker(&fakePresence{present: false})
	// "shell.exec" is DESTRUCTIVE -> RequirePresenceAndConsent.
	d, reason := e.Evaluate(context.Background(), blastradius.Action{Kind: "shell.exec"})
	if d != Deny {
		t.Fatalf("absent user must deny DESTRUCTIVE, got %v (%s)", d, reason)
	}
	if !strings.Contains(reason, "absent") {
		t.Fatalf("reason should mention absent, got %q", reason)
	}
}

// TestPresence_PresentReachesConsent: a present user with an approving
// provider gets Allow (the presence gate does not block the modal path).
func TestPresence_PresentReachesConsent(t *testing.T) {
	e := NewEngine(DefaultPolicy(), approveConsent{}, nil)
	e.SetPresenceChecker(&fakePresence{present: true})
	d, _ := e.Evaluate(context.Background(), blastradius.Action{Kind: "shell.exec"})
	if d != Allow {
		t.Fatalf("present user + approved consent must allow DESTRUCTIVE, got %v", d)
	}
}

// TestPresence_NoCheckerFallsBackToModal: with no PresenceChecker
// wired, the gate must NOT deny for absence — it falls back to the
// consent modal's timeout-queue backstop. This keeps existing behavior
// (and tests that don't wire a checker) unchanged. N1 safe fallback.
func TestPresence_NoCheckerFallsBackToModal(t *testing.T) {
	e := NewEngine(DefaultPolicy(), approveConsent{}, nil)
	// no SetPresenceChecker
	d, reason := e.Evaluate(context.Background(), blastradius.Action{Kind: "shell.exec"})
	if d != Allow {
		t.Fatalf("no checker + approved consent must allow (modal backstop), got %v (%s)", d, reason)
	}
}

// TestPresence_AbsentQueuesWhenConfigured: the DESTRUCTIVE default
// carries on_user_absent: queue, so the deny reason mentions "queued".
func TestPresence_AbsentQueuesWhenConfigured(t *testing.T) {
	e := NewEngine(DefaultPolicy(), approveConsent{}, nil)
	e.SetPresenceChecker(&fakePresence{present: false})
	_, reason := e.Evaluate(context.Background(), blastradius.Action{Kind: "shell.exec"})
	if !strings.Contains(reason, "queued") {
		t.Fatalf("queue reason should mention queued, got %q", reason)
	}
}
