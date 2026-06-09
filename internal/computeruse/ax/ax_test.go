package ax

import (
	"context"
	"errors"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/computeruse"
)

func TestBackendName(t *testing.T) {
	b := New()
	name := b.Name()
	if name == "" {
		t.Error("expected non-empty name")
	}
}

func TestBackendCapabilities(t *testing.T) {
	b := New()
	caps := b.Capabilities()
	if len(caps) == 0 {
		t.Skip("no capabilities on this platform")
	}
}

func TestIsAvailable(t *testing.T) {
	b := New()
	ctx := context.Background()

	available := b.IsAvailable(ctx)
	t.Logf("Accessibility available: %v", available)
}

func TestCaptureScreen(t *testing.T) {
	b := New()
	ctx := context.Background()

	if !b.IsAvailable(ctx) {
		t.Skip("Accessibility not available")
	}

	screenshot, err := b.CaptureScreen(ctx)
	if errors.Is(err, computeruse.ErrUnsupportedAction) {
		t.Skip("CaptureScreen not supported by this backend")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if screenshot == nil {
		t.Fatal("expected screenshot, got nil")
	}
	if screenshot.Width <= 0 || screenshot.Height <= 0 {
		t.Errorf("invalid dimensions: %dx%d", screenshot.Width, screenshot.Height)
	}
}

func TestGetAXTree(t *testing.T) {
	b := New()
	ctx := context.Background()

	if !b.IsAvailable(ctx) {
		t.Skip("Accessibility not available")
	}

	tree, err := b.GetAXTree(ctx)
	if err != nil {
		t.Skipf("GetAXTree failed (likely no focused app): %v", err)
	}
	if tree == nil {
		t.Fatal("expected tree, got nil")
	}
	if tree.Root == nil {
		t.Fatal("expected root node, got nil")
	}
	if tree.Root.Role == "" {
		t.Error("expected non-empty role")
	}
}

func TestExecuteUnsupported(t *testing.T) {
	b := New()
	ctx := context.Background()

	if !b.IsAvailable(ctx) {
		t.Skip("Accessibility not available")
	}

	action := &computeruse.Action{
		Type: computeruse.ActionClick,
	}

	result, err := b.Execute(ctx, action)
	if !errors.Is(err, computeruse.ErrUnsupportedAction) && !errors.Is(err, computeruse.ErrNoBackend) {
		t.Errorf("expected ErrUnsupportedAction or ErrNoBackend, got %v", err)
	}
	if result.Success {
		t.Error("expected failure")
	}
}
