package backends

import (
	"context"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/computeruse"
)

// fakeORAX implements oraXImpl for testing.
type fakeORAX struct {
	available     bool
	screenshot    *computeruse.Screenshot
	screenshotErr error
	axTree        *computeruse.AXTree
	axTreeErr     error
	execResult    *computeruse.ActionResult
	execErr       error
}

func (f *fakeORAX) name() string      { return "orax-test" }
func (f *fakeORAX) isAvailable() bool { return f.available }
func (f *fakeORAX) captureScreen() (*computeruse.Screenshot, error) {
	return f.screenshot, f.screenshotErr
}
func (f *fakeORAX) getAXTree() (*computeruse.AXTree, error) {
	return f.axTree, f.axTreeErr
}
func (f *fakeORAX) execute(action *computeruse.Action) (*computeruse.ActionResult, error) {
	if f.execResult != nil {
		f.execResult.Action = action
		return f.execResult, nil
	}
	return nil, f.execErr
}

func TestORAX_ImplementsBackendInterface(t *testing.T) {
	// Compile-time: ORAXBackend satisfies computeruse.Backend.
	var _ computeruse.Backend = (*ORAXBackend)(nil)

	b := NewORAX()
	if b == nil {
		t.Fatal("NewORAX returned nil")
	}
}

func TestORAX_AllCapabilities(t *testing.T) {
	b := &ORAXBackend{impl: &fakeORAX{}}
	caps := b.Capabilities()
	want := map[computeruse.Capability]bool{
		computeruse.CapScreenshot: true,
		computeruse.CapAXTree:     true,
		computeruse.CapClick:      true,
		computeruse.CapType:       true,
		computeruse.CapScroll:     true,
		computeruse.CapKeyPress:   true,
		computeruse.CapLaunch:     true,
		computeruse.CapFocus:      true,
	}
	if len(caps) != len(want) {
		t.Errorf("got %d capabilities, want %d", len(caps), len(want))
	}
	for _, c := range caps {
		if !want[c] {
			t.Errorf("unexpected capability %q", c)
		}
	}
}

func TestORAX_Name(t *testing.T) {
	b := &ORAXBackend{impl: &fakeORAX{}}
	if b.Name() != "orax-test" {
		t.Errorf("name = %q, want %q", b.Name(), "orax-test")
	}
}

func TestORAX_IsAvailable(t *testing.T) {
	t.Run("available", func(t *testing.T) {
		b := &ORAXBackend{impl: &fakeORAX{available: true}}
		if !b.IsAvailable(context.Background()) {
			t.Error("expected available")
		}
	})
	t.Run("unavailable", func(t *testing.T) {
		b := &ORAXBackend{impl: &fakeORAX{available: false}}
		if b.IsAvailable(context.Background()) {
			t.Error("expected unavailable")
		}
	})
}

func TestORAX_CaptureScreen(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		want := &computeruse.Screenshot{Width: 1920, Height: 1080}
		b := &ORAXBackend{impl: &fakeORAX{screenshot: want}}
		got, err := b.CaptureScreen(context.Background())
		if err != nil {
			t.Fatalf("CaptureScreen: %v", err)
		}
		if got.Width != want.Width || got.Height != want.Height {
			t.Errorf("screenshot = %dx%d, want %dx%d", got.Width, got.Height, want.Width, want.Height)
		}
	})
	t.Run("error", func(t *testing.T) {
		b := &ORAXBackend{impl: &fakeORAX{screenshotErr: computeruse.ErrScreenshotFailed}}
		_, err := b.CaptureScreen(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestORAX_GetAXTree(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		want := &computeruse.AXTree{
			Root: &computeruse.AXNode{Role: "AXApplication", Title: "Test"},
			PID:  12345,
		}
		b := &ORAXBackend{impl: &fakeORAX{axTree: want}}
		got, err := b.GetAXTree(context.Background())
		if err != nil {
			t.Fatalf("GetAXTree: %v", err)
		}
		if got.Root.Title != want.Root.Title {
			t.Errorf("title = %q, want %q", got.Root.Title, want.Root.Title)
		}
	})
	t.Run("error", func(t *testing.T) {
		b := &ORAXBackend{impl: &fakeORAX{axTreeErr: computeruse.ErrAXTreeFailed}}
		_, err := b.GetAXTree(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestORAX_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		want := &computeruse.ActionResult{Success: true}
		b := &ORAXBackend{impl: &fakeORAX{execResult: want}}
		action := &computeruse.Action{Type: computeruse.ActionClick, Value: "test"}
		got, err := b.Execute(context.Background(), action)
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if !got.Success {
			t.Error("expected success")
		}
	})
	t.Run("error", func(t *testing.T) {
		b := &ORAXBackend{impl: &fakeORAX{execErr: computeruse.ErrUnsupportedAction}}
		_, err := b.Execute(context.Background(), &computeruse.Action{Type: computeruse.ActionClick})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
