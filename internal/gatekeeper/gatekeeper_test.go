package gatekeeper

import (
	"context"
	"strings"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/blastradius"
)

// DenyBeyondRead must satisfy the Gatekeeper interface.
var _ Gatekeeper = DenyBeyondRead{}

func TestDenyBeyondRead_AllowsRead(t *testing.T) {
	g := NewDenyBeyondRead()
	ctx := context.Background()
	reads := []string{"chat", "transcribe", "speak", "screenshot.read"}
	for _, kind := range reads {
		d, reason := g.Evaluate(ctx, blastradius.Action{Kind: kind})
		if d != Allow {
			t.Errorf("Evaluate(%q) = %v (%q), want Allow", kind, d, reason)
		}
	}
}

func TestDenyBeyondRead_DeniesWrite(t *testing.T) {
	g := NewDenyBeyondRead()
	d, reason := g.Evaluate(context.Background(), blastradius.Action{Kind: "file.write"})
	if d != Deny {
		t.Fatalf("Evaluate(file.write) = %v, want Deny", d)
	}
	if reason == "" {
		t.Fatal("deny reason must not be empty")
	}
	if !strings.Contains(reason, "WRITE") {
		t.Errorf("deny reason %q should name the blocked class WRITE", reason)
	}
}

func TestDenyBeyondRead_DeniesNetwork(t *testing.T) {
	g := NewDenyBeyondRead()
	d, _ := g.Evaluate(context.Background(), blastradius.Action{Kind: "http.request"})
	if d != Deny {
		t.Errorf("Evaluate(http.request) = %v, want Deny", d)
	}
}

func TestDenyBeyondRead_DeniesDestructive(t *testing.T) {
	g := NewDenyBeyondRead()
	d, _ := g.Evaluate(context.Background(), blastradius.Action{Kind: "shell.exec"})
	if d != Deny {
		t.Errorf("Evaluate(shell.exec) = %v, want Deny", d)
	}
}

// An unrecognized action kind classifies as DESTRUCTIVE and so must be
// denied — there is no path to physical action while the real rules
// engine is unbuilt (MISSION §2 Survival Rule).
func TestDenyBeyondRead_DeniesUnknown(t *testing.T) {
	g := NewDenyBeyondRead()
	d, reason := g.Evaluate(context.Background(), blastradius.Action{Kind: "mystery.move"})
	if d != Deny {
		t.Errorf("Evaluate(unknown) = %v, want Deny", d)
	}
	if !strings.Contains(strings.ToLower(reason), "safety") {
		t.Errorf("deny reason %q should explain the safety layer is unbuilt", reason)
	}
}

func TestDecision_String(t *testing.T) {
	if Allow.String() != "Allow" {
		t.Errorf("Allow.String() = %q, want Allow", Allow.String())
	}
	if Deny.String() != "Deny" {
		t.Errorf("Deny.String() = %q, want Deny", Deny.String())
	}
}

// An out-of-range Decision must render as Deny — the safe default.
func TestDecision_String_OutOfRangeIsDeny(t *testing.T) {
	if got := Decision(99).String(); got != "Deny" {
		t.Errorf("Decision(99).String() = %q, want Deny", got)
	}
}
