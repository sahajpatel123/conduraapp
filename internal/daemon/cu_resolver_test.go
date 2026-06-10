package daemon

import (
	"context"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/agent"
	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/computeruse"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
)

func resolverMocks() (*computeruse.ComputerUse, *computeruse.GatedExecutor) {
	backend := &computeruse.MockBackend{
		Available: true,
		Caps: []computeruse.Capability{
			computeruse.CapClick,
			computeruse.CapType,
			computeruse.CapScroll,
			computeruse.CapKeyPress,
			computeruse.CapLaunch,
		},
	}
	cu := computeruse.New(backend)
	gate := gatekeeper.NewDenyBeyondRead()
	gexec := computeruse.NewGatedExecutor(cu, gate)
	return cu, gexec
}

func TestCUResolver_ParseActionType(t *testing.T) {
	tests := []struct {
		verb    string
		want    computeruse.ActionType
		wantErr bool
	}{
		{"click", computeruse.ActionClick, false},
		{"CLICK", computeruse.ActionClick, false},
		{"type", computeruse.ActionTypeText, false},
		{"scroll", computeruse.ActionScroll, false},
		{"key_press", computeruse.ActionKeyPress, false},
		{"keypress", computeruse.ActionKeyPress, false},
		{"drag", computeruse.ActionDrag, false},
		{"launch", computeruse.ActionLaunch, false},
		{"focus", computeruse.ActionFocus, false},
		{"wait", computeruse.ActionWait, false},
		{"unknown", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.verb, func(t *testing.T) {
			got, err := parseActionType(tt.verb)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseActionType(%q) err=%v, wantErr=%v", tt.verb, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseActionType(%q) = %v, want %v", tt.verb, got, tt.want)
			}
		})
	}
}

func TestCUResolver_ResolveTarget(t *testing.T) {
	tests := []struct {
		desc      string
		wantRole  string
		wantTitle string
	}{
		{"Submit button", "AXButton", "submit"},
		{"password field", "AXTextField", "password"},
		{"the OK button in the dialog", "AXButton", "ok"},
		{"a text field", "AXTextField", ""},
		{"input", "AXTextField", "input"},
		{"checkbox", "AXCheckBox", "checkbox"},
		{"link", "AXLink", "link"},
		{"menu item", "AXMenuItem", ""},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			target, err := resolveTarget(tt.desc)
			if err != nil {
				t.Fatalf("resolveTarget(%q) err=%v", tt.desc, err)
			}
			if target == nil {
				t.Fatalf("expected non-nil target for %q", tt.desc)
			}
			if target.Role != tt.wantRole {
				t.Errorf("role = %q, want %q", target.Role, tt.wantRole)
			}
			if target.Title != tt.wantTitle {
				t.Errorf("title = %q, want %q", target.Title, tt.wantTitle)
			}
		})
	}
}

func TestCUResolver_ResolveFullAction(t *testing.T) {
	backend := &computeruse.MockBackend{
		Available: true,
		Caps:      []computeruse.Capability{computeruse.CapClick},
	}
	cu := computeruse.New(backend)
	gexec := computeruse.NewGatedExecutor(cu, allowAllGatekeeper{})
	r := NewCUResolver(cu, gexec)

	act := &agent.Action{
		Type:        "click",
		Target:      "Submit button",
		Value:       "",
		Description: "Click the Submit button",
	}
	result, err := r.Execute(context.Background(), act)
	if err != nil {
		t.Fatalf("Execute err=%v", err)
	}
	if !result.Success {
		t.Errorf("expected success, got Success=false")
	}
	if result.Duration <= 0 {
		t.Errorf("duration = %v, want > 0", result.Duration)
	}
	if result.Output == "" {
		t.Errorf("output should be non-empty")
	}
}

func TestCUResolver_UnknownVerb(t *testing.T) {
	cu, gexec := resolverMocks()
	r := NewCUResolver(cu, gexec)

	act := &agent.Action{Type: "frobnicate", Target: "foo"}
	result, err := r.Execute(context.Background(), act)
	if err == nil {
		t.Fatal("expected error for unknown verb, got nil")
	}
	if result == nil || result.Success {
		t.Error("expected result.Success=false")
	}
}

func TestCUResolver_EmptyTarget(t *testing.T) {
	backend := &computeruse.MockBackend{
		Available: true,
		Caps:      []computeruse.Capability{computeruse.CapLaunch},
	}
	cu := computeruse.New(backend)
	gexec := computeruse.NewGatedExecutor(cu, allowAllGatekeeper{})
	r := NewCUResolver(cu, gexec)

	act := &agent.Action{
		Type:  "launch",
		Value: "Safari",
	}
	result, err := r.Execute(context.Background(), act)
	if err != nil {
		t.Fatalf("Execute err=%v", err)
	}
	if !result.Success {
		t.Errorf("expected success for empty-target launch, got Success=false")
	}
}

func TestCUResolver_GatekeeperBlocksAction(t *testing.T) {
	cu, _ := resolverMocks()
	denyGate := denyAllGatekeeper{}
	gexec := computeruse.NewGatedExecutor(cu, denyGate)
	r := NewCUResolver(cu, gexec)

	act := &agent.Action{Type: "click", Target: "Submit button"}
	result, err := r.Execute(context.Background(), act)
	if err == nil {
		t.Fatal("expected gatekeeper denial error, got nil")
	}
	if result == nil || result.Success {
		t.Error("expected result.Success=false when gatekeeper denies")
	}
}

type denyAllGatekeeper struct{}

func (d denyAllGatekeeper) Evaluate(_ context.Context, _ blastradius.Action) (gatekeeper.Decision, string) {
	return gatekeeper.Deny, "test: all actions denied"
}

type allowAllGatekeeper struct{}

func (a allowAllGatekeeper) Evaluate(_ context.Context, _ blastradius.Action) (gatekeeper.Decision, string) {
	return gatekeeper.Allow, ""
}
