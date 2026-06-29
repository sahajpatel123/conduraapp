package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/computeruse"
)

// errUnimplemented is returned by stub backend methods that aren't
// relevant to the test path. Avoids the linter's nilnil complaint
// and gives a clear signal in failure logs.
var errUnimplemented = errors.New("unimplemented in stub")

// stubBackend records the last action it received and returns a
// preset result. Used by ComputerUseExecutor tests.
type stubBackend struct {
	name     string
	lastAct  *computeruse.Action
	result   *computeruse.ActionResult
	failNext bool
}

func (s *stubBackend) Name() string { return s.name }
func (s *stubBackend) Capabilities() []computeruse.Capability {
	return []computeruse.Capability{
		computeruse.CapClick, computeruse.CapType,
		computeruse.CapScroll, computeruse.CapKeyPress,
		computeruse.CapDrag, computeruse.CapLaunch,
		computeruse.CapFocus,
	}
}
func (s *stubBackend) CaptureScreen(_ context.Context) (*computeruse.Screenshot, error) {
	return nil, errUnimplemented
}
func (s *stubBackend) GetAXTree(_ context.Context) (*computeruse.AXTree, error) {
	return nil, errUnimplemented
}
func (s *stubBackend) Execute(_ context.Context, a *computeruse.Action) (*computeruse.ActionResult, error) {
	s.lastAct = a
	if s.failNext {
		return &computeruse.ActionResult{Success: false, Error: errors.New("stub fail")}, nil
	}
	return s.result, nil
}
func (s *stubBackend) IsAvailable(_ context.Context) bool { return true }

func TestComputerUseExecutor_Nil(t *testing.T) {
	var e *ComputerUseExecutor
	res, err := e.Execute(context.Background(), &Action{Type: "click", Target: "OK"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.Success {
		t.Fatal("expected nil executor to fail")
	}
}

func TestComputerUseExecutor_NilCU(t *testing.T) {
	e := &ComputerUseExecutor{}
	res, err := e.Execute(context.Background(), &Action{Type: "click", Target: "OK"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.Success {
		t.Fatal("expected nil CU to fail")
	}
}

func TestComputerUseExecutor_Click(t *testing.T) {
	be := &stubBackend{
		name:   "stub",
		result: &computeruse.ActionResult{Success: true},
	}
	cu := computeruse.New(be)
	e := NewComputerUseExecutor(cu)
	res, err := e.Execute(context.Background(), &Action{Type: "click", Target: "Submit"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success, got %+v", res)
	}
	if be.lastAct == nil {
		t.Fatal("backend did not receive action")
	}
	if be.lastAct.Type != computeruse.ActionClick {
		t.Errorf("type=%v want click", be.lastAct.Type)
	}
	if be.lastAct.Target == nil || be.lastAct.Target.Title != "Submit" {
		t.Errorf("target title not propagated: %+v", be.lastAct.Target)
	}
}

func TestComputerUseExecutor_Type(t *testing.T) {
	be := &stubBackend{
		name:   "stub",
		result: &computeruse.ActionResult{Success: true},
	}
	cu := computeruse.New(be)
	e := NewComputerUseExecutor(cu)
	_, err := e.Execute(context.Background(), &Action{Type: "type", Value: "hello"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if be.lastAct.Type != computeruse.ActionTypeText {
		t.Errorf("type=%v want type_text", be.lastAct.Type)
	}
	if be.lastAct.Value != "hello" {
		t.Errorf("value=%q want hello", be.lastAct.Value)
	}
}

func TestComputerUseExecutor_LaunchUsesTarget(t *testing.T) {
	be := &stubBackend{
		name:   "stub",
		result: &computeruse.ActionResult{Success: true},
	}
	cu := computeruse.New(be)
	e := NewComputerUseExecutor(cu)
	_, err := e.Execute(context.Background(), &Action{Type: "launch", Target: "Safari"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if be.lastAct.Type != computeruse.ActionLaunch {
		t.Errorf("type=%v want launch", be.lastAct.Type)
	}
	if be.lastAct.Value != "Safari" {
		t.Errorf("value=%q want Safari", be.lastAct.Value)
	}
}

func TestComputerUseExecutor_UnknownTypeBecomesWait(t *testing.T) {
	be := &stubBackend{
		name:   "stub",
		result: &computeruse.ActionResult{Success: true},
	}
	cu := computeruse.New(be)
	e := NewComputerUseExecutor(cu)
	_, err := e.Execute(context.Background(), &Action{Type: "make_coffee"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if be.lastAct.Type != computeruse.ActionWait {
		t.Errorf("type=%v want wait (unknown type fallback)", be.lastAct.Type)
	}
}

func TestComputerUseExecutor_BackendFailure(t *testing.T) {
	be := &stubBackend{name: "stub", failNext: true}
	cu := computeruse.New(be)
	e := NewComputerUseExecutor(cu)
	res, err := e.Execute(context.Background(), &Action{Type: "click", Target: "X"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.Success {
		t.Fatal("expected failure when backend fails")
	}
	if res.Error == nil {
		t.Fatal("expected error in result")
	}
}

func TestComputerUseExecutor_ComputerUseAvailable(t *testing.T) {
	var e *ComputerUseExecutor
	if e.ComputerUseAvailable() {
		t.Fatal("nil executor should not be available")
	}
	e = &ComputerUseExecutor{}
	if e.ComputerUseAvailable() {
		t.Fatal("empty executor should not be available")
	}
	e = NewComputerUseExecutor(computeruse.New(&stubBackend{name: "stub"}))
	if !e.ComputerUseAvailable() {
		t.Fatal("wired executor should be available")
	}
}

func TestTranslateAgentAction_NilAction(t *testing.T) {
	_, err := translateAgentAction(nil)
	if err == nil {
		t.Fatal("expected error for nil action")
	}
}
