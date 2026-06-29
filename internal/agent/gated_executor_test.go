package agent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/blastradius"
	"github.com/sahajpatel123/conduraapp/internal/gatekeeper"
)

// fakeExecutor is a stub Executor that records every call.
type fakeExecutor struct {
	calls    int
	actions  []*Action
	response *StepResult
	err      error
}

func (f *fakeExecutor) Execute(_ context.Context, a *Action) (*StepResult, error) {
	f.calls++
	f.actions = append(f.actions, a)
	if f.err != nil {
		return nil, f.err
	}
	if f.response != nil {
		return f.response, nil
	}
	return &StepResult{Success: true}, nil
}

func TestGatedExecutor_AllowsReadAction(t *testing.T) {
	inner := &fakeExecutor{response: &StepResult{Success: true}}
	g := NewGatedExecutor(inner, gatekeeper.NewDenyBeyondRead(), nil)

	action := &Action{Type: "chat", Description: "ask a question"}
	result, err := g.Execute(context.Background(), action)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || !result.Success {
		t.Errorf("expected successful result, got %+v", result)
	}
	if inner.calls != 1 {
		t.Errorf("inner executor called %d times, want 1", inner.calls)
	}
}

func TestGatedExecutor_DeniesWriteAction(t *testing.T) {
	inner := &fakeExecutor{response: &StepResult{Success: true}}
	g := NewGatedExecutor(inner, gatekeeper.NewDenyBeyondRead(), nil)

	action := &Action{Type: "click", Description: "click a button"}
	result, err := g.Execute(context.Background(), action)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result == nil || result.Success {
		t.Errorf("expected failed result, got %+v", result)
	}
	if inner.calls != 0 {
		t.Errorf("inner executor should not be called on denial, was called %d times", inner.calls)
	}
	// The error must mention blast radius and reason.
	if !strings.Contains(result.Error.Error(), "WRITE") {
		t.Errorf("expected error to mention class WRITE, got: %v", result.Error)
	}
}

func TestGatedExecutor_DeniesDestructiveAction(t *testing.T) {
	inner := &fakeExecutor{response: &StepResult{Success: true}}
	g := NewGatedExecutor(inner, gatekeeper.NewDenyBeyondRead(), nil)

	action := &Action{Type: "shell.exec", Description: "rm -rf /"}
	_, err := g.Execute(context.Background(), action)

	if err == nil {
		t.Fatal("expected error for destructive action, got nil")
	}
	if inner.calls != 0 {
		t.Errorf("inner executor should not be called on denial")
	}
}

func TestGatedExecutor_BlastRadiusClassification(t *testing.T) {
	tests := []struct {
		actionType string
		wantClass  blastradius.Class
		allowed    bool
	}{
		{"chat", blastradius.READ, true},
		{"screenshot.read", blastradius.READ, true},
		{"ax.read", blastradius.READ, true},
		{"file.read", blastradius.READ, true},
		{"click", blastradius.WRITE, false},
		{"type", blastradius.WRITE, false},
		{"file.write", blastradius.WRITE, false},
		{"http.request", blastradius.NETWORK, false},
		{"shell.exec", blastradius.DESTRUCTIVE, false},
		{"file.delete", blastradius.DESTRUCTIVE, false},
	}

	inner := &fakeExecutor{response: &StepResult{Success: true}}
	g := NewGatedExecutor(inner, gatekeeper.NewDenyBeyondRead(), nil)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.actionType, func(t *testing.T) {
			before := inner.calls
			action := &Action{Type: tt.actionType}
			class := blastradius.Classify(action.ToBlastRadius())
			if class != tt.wantClass {
				t.Errorf("class = %v, want %v", class, tt.wantClass)
			}
			_, err := g.Execute(ctx, action)
			if tt.allowed && err != nil {
				t.Errorf("expected allow, got error: %v", err)
			}
			if !tt.allowed && err == nil {
				t.Errorf("expected deny, got nil error")
			}
			if tt.allowed && inner.calls != before+1 {
				t.Errorf("inner not called for allowed action")
			}
			if !tt.allowed && inner.calls != before {
				t.Errorf("inner called for denied action")
			}
		})
	}
}

func TestGatedExecutor_NilAuditLogIsOK(t *testing.T) {
	inner := &fakeExecutor{response: &StepResult{Success: true}}
	// nil audit log is explicitly allowed.
	g := NewGatedExecutor(inner, gatekeeper.NewDenyBeyondRead(), nil)

	action := &Action{Type: "chat"}
	_, err := g.Execute(context.Background(), action)
	if err != nil {
		t.Fatalf("unexpected error with nil audit: %v", err)
	}
}

func TestGatedExecutor_InnerErrorIsPropagated(t *testing.T) {
	wantErr := errors.New("inner failed")
	inner := &fakeExecutor{err: wantErr}
	g := NewGatedExecutor(inner, gatekeeper.NewDenyBeyondRead(), nil)

	action := &Action{Type: "chat"}
	_, err := g.Execute(context.Background(), action)
	if !errors.Is(err, wantErr) {
		t.Errorf("expected inner error, got: %v", err)
	}
}

func TestGatedExecutor_DenialAlwaysFails(t *testing.T) {
	// Even if the inner executor would succeed, a denial must NOT
	// invoke the inner executor.
	inner := &fakeExecutor{response: &StepResult{Success: true}}
	g := NewGatedExecutor(inner, gatekeeper.NewDenyBeyondRead(), nil)

	action := &Action{Type: "shell.exec"}
	result, err := g.Execute(context.Background(), action)

	if err == nil {
		t.Fatal("expected error for shell.exec")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Success {
		t.Error("denial must not produce a successful result")
	}
	if result.Error == nil {
		t.Error("denial must populate StepResult.Error")
	}
}
