package backends

import (
	"context"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/computeruse"
)

type fakeMCP struct {
	available  bool
	screenshot *computeruse.Screenshot
	axTree     *computeruse.AXTree
	execResult *computeruse.ActionResult
}

func (f *fakeMCP) name() string                                    { return "macos-mcp-test" }
func (f *fakeMCP) isAvailable() bool                               { return f.available }
func (f *fakeMCP) captureScreen() (*computeruse.Screenshot, error) { return f.screenshot, nil }
func (f *fakeMCP) getAXTree() (*computeruse.AXTree, error)         { return f.axTree, nil }
func (f *fakeMCP) execute(a *computeruse.Action) (*computeruse.ActionResult, error) {
	if f.execResult != nil {
		f.execResult.Action = a
		return f.execResult, nil
	}
	return &computeruse.ActionResult{Success: true}, nil
}

func TestMCP_ImplementsBackend(t *testing.T) {
	var _ computeruse.Backend = (*MacOSMCPBackend)(nil)
}

func TestMCP_Capabilities(t *testing.T) {
	b := &MacOSMCPBackend{impl: &fakeMCP{}}
	if len(b.Capabilities()) != 9 {
		t.Errorf("got %d capabilities, want 9", len(b.Capabilities()))
	}
}

func TestMCP_IsAvailable(t *testing.T) {
	b := &MacOSMCPBackend{impl: &fakeMCP{available: true}}
	if !b.IsAvailable(context.Background()) {
		t.Error("expected available")
	}
}

func TestMCP_Execute(t *testing.T) {
	b := &MacOSMCPBackend{impl: &fakeMCP{}}
	r, err := b.Execute(context.Background(), &computeruse.Action{Type: computeruse.ActionClick})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !r.Success {
		t.Error("expected success")
	}
}

func TestMCP_GetAXTree(t *testing.T) {
	want := &computeruse.AXTree{PID: 1}
	b := &MacOSMCPBackend{impl: &fakeMCP{axTree: want}}
	got, err := b.GetAXTree(context.Background())
	if err != nil {
		t.Fatalf("GetAXTree: %v", err)
	}
	if got.PID != 1 {
		t.Errorf("PID = %d", got.PID)
	}
}
