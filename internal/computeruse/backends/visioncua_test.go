package backends

import (
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/computeruse"
)

func TestVisionCUA_Disabled(t *testing.T) {
	b := NewVisionCUA(VisionCUAConfig{Enabled: false})
	if b != nil {
		t.Fatal("expected nil for disabled backend")
	}
}

func TestVisionCUA_ImplementsBackend(t *testing.T) {
	b := NewVisionCUA(VisionCUAConfig{Enabled: true})
	if b == nil {
		t.Fatal("expected non-nil")
	}
	var _ computeruse.Backend = b
}

func TestVisionCUA_Capabilities(t *testing.T) {
	b := NewVisionCUA(VisionCUAConfig{Enabled: true})
	if len(b.Capabilities()) != 5 {
		t.Errorf("got %d caps", len(b.Capabilities()))
	}
}

func TestVisionCUA_MaxCallsDefault(t *testing.T) {
	b := NewVisionCUA(VisionCUAConfig{Enabled: true, MaxConsecutiveCalls: 0})
	if b.cfg.MaxConsecutiveCalls != 5 {
		t.Errorf("max calls = %d, want 5", b.cfg.MaxConsecutiveCalls)
	}
}

func TestVisionCUA_MaxCallsExceeded(t *testing.T) {
	b := NewVisionCUA(VisionCUAConfig{Enabled: true, MaxConsecutiveCalls: 2})
	b.callCount = 2
	_, err := b.doExecute(&computeruse.Action{Type: computeruse.ActionClick})
	if err == nil {
		t.Fatal("expected error for exceeded max calls")
	}
}

func TestVisionCUA_ParseResponse(t *testing.T) {
	b := NewVisionCUA(VisionCUAConfig{Enabled: true})
	result := b.parseResponse(`{"x":100,"y":200}`, &computeruse.Action{Type: computeruse.ActionClick})
	if !result.Success {
		t.Fatal("expected success")
	}
	if result.Action.Bounds.X != 100 || result.Action.Bounds.Y != 200 {
		t.Errorf("coords = (%.0f, %.0f)", result.Action.Bounds.X, result.Action.Bounds.Y)
	}
}

func TestExtractJSON_Vision(t *testing.T) {
	tests := []struct{ in, want string }{
		{`{"x":1}`, `{"x":1}`},
		{`Here: {"x":2}`, `{"x":2}`},
		{`{"y":3} extra`, `{"y":3}`},
	}
	for _, tt := range tests {
		got := extractJSON(tt.in)
		if got != tt.want {
			t.Errorf("extractJSON(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestVisionCUA_BuildPrompt(t *testing.T) {
	b := NewVisionCUA(VisionCUAConfig{Enabled: true})
	p := b.buildPrompt(&computeruse.Action{
		Type:   computeruse.ActionClick,
		Target: &computeruse.Target{Title: "Submit"},
		Value:  "",
	})
	if p == "" {
		t.Error("expected non-empty prompt")
	}
}
