package main

import (
	"context"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/presence"
)

func TestResolveOverlayHotkey(t *testing.T) {
	t.Parallel()
	if got := resolveOverlayHotkey(""); got != DefaultQuickPromptHotkey {
		t.Fatalf("empty config: got %q want %q", got, DefaultQuickPromptHotkey)
	}
	if got := resolveOverlayHotkey("  "); got != DefaultQuickPromptHotkey {
		t.Fatalf("whitespace config: got %q want %q", got, DefaultQuickPromptHotkey)
	}
	if got := resolveOverlayHotkey("Cmd+K"); got != "Cmd+K" {
		t.Fatalf("custom config: got %q want Cmd+K", got)
	}
}

func TestAppQuickPromptVisibility(t *testing.T) {
	t.Parallel()
	a := NewApp()
	if a.isQuickPromptVisible() {
		t.Fatal("new app should not show quick prompt")
	}
	a.overlay.Store(true)
	if !a.isQuickPromptVisible() {
		t.Fatal("overlay flag should make prompt visible")
	}
}

func TestPresenceOrchestratorMarksVisible(t *testing.T) {
	t.Parallel()
	a := NewApp()
	a.presenceOrch = presence.NewOrchestrator(a.overlayCtrl, nil, nil)
	_ = a.presenceOrch.Summon(context.Background())
	if !a.isQuickPromptVisible() {
		t.Fatal("active presence session should count as visible")
	}
}
