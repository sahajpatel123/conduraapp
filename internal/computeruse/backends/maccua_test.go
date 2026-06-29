package backends

import (
	"context"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/computeruse"
)

type fakeMC struct {
	available  bool
	screenshot *computeruse.Screenshot
	axTree     *computeruse.AXTree
	execResult *computeruse.ActionResult
}

func (f *fakeMC) name() string                                    { return "mac-cua-test" }
func (f *fakeMC) isAvailable() bool                               { return f.available }
func (f *fakeMC) captureScreen() (*computeruse.Screenshot, error) { return f.screenshot, nil }
func (f *fakeMC) getAXTree() (*computeruse.AXTree, error)         { return f.axTree, nil }
func (f *fakeMC) execute(a *computeruse.Action) (*computeruse.ActionResult, error) {
	if f.execResult != nil {
		f.execResult.Action = a
		return f.execResult, nil
	}
	return &computeruse.ActionResult{Success: true}, nil
}

func TestMacCUA_ImplementsBackend(t *testing.T) {
	var _ computeruse.Backend = (*MacCUABackend)(nil)
}

func TestMacCUA_Capabilities(t *testing.T) {
	b := &MacCUABackend{impl: &fakeMC{}}
	caps := b.Capabilities()
	if len(caps) != 6 {
		t.Errorf("got %d capabilities, want 6", len(caps))
	}
}

func TestMacCUA_IsAvailable(t *testing.T) {
	b := &MacCUABackend{impl: &fakeMC{available: true}}
	if !b.IsAvailable(context.Background()) {
		t.Error("expected available")
	}
}

func TestMacCUA_NoScreenshot(t *testing.T) {
	b := &MacCUABackend{impl: &fakeMC{}}
	_, err := b.CaptureScreen(context.Background())
	if err != nil {
		return // Expected: fake returns nil, nil for screenshot
	}
	// Real darwin impl returns ErrUnsupportedAction, but fake returns nil.
	// Test verifies the method doesn't panic.
}

func TestMacCUA_GetAXTree(t *testing.T) {
	want := &computeruse.AXTree{PID: 42}
	b := &MacCUABackend{impl: &fakeMC{axTree: want}}
	got, err := b.GetAXTree(context.Background())
	if err != nil {
		t.Fatalf("GetAXTree: %v", err)
	}
	if got.PID != 42 {
		t.Errorf("PID = %d", got.PID)
	}
}

func TestMacCUA_Execute(t *testing.T) {
	b := &MacCUABackend{impl: &fakeMC{execResult: &computeruse.ActionResult{Success: true}}}
	r, err := b.Execute(context.Background(), &computeruse.Action{Type: computeruse.ActionClick, AppPID: 1234})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !r.Success {
		t.Error("expected success")
	}
}
