package daemon

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/agent"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/blastradius"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/computeruse"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/gatekeeper"
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

// Phase 17, Fix #6 (B4): the resolver must perform a twin-snapshot
// verification around every gated execution. When the AX tree changes
// between the pre and post snapshot in a way that indicates stale
// state (window focus changed, target node removed), the action must
// be aborted and ErrStaleState returned.
//
// This test exercises the happy path (identical pre/post trees) AND
// the abort path (pre-tree has the target button; post-tree has it
// removed because the user closed the dialog during the action).
func TestCUResolver_TwinSnapshotVerification(t *testing.T) {
	t.Run("identical_trees_no_abort", func(t *testing.T) {
		cu, _ := resolverMocks()
		gexec := computeruse.NewGatedExecutor(cu, allowAllGatekeeper{})
		r := NewCUResolver(cu, gexec)
		act := &agent.Action{Type: "click", Target: "OK button"}
		res, err := r.Execute(context.Background(), act)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.Success {
			t.Errorf("expected success with identical pre/post trees, got %+v", res)
		}
	})

	t.Run("node_removed_aborts_with_ErrStaleState", func(t *testing.T) {
		// Build a custom mock backend whose AX tree is
		// non-deterministic: returns the full tree the first
		// time, then a stripped tree thereafter (target node
		// removed = user closed the dialog mid-click).
		preTree := &computeruse.AXTree{
			Root: &computeruse.AXNode{
				Role: "AXApplication", Title: "TestApp",
				Children: []*computeruse.AXNode{{
					Role: "AXWindow", Title: "MainWindow",
					Children: []*computeruse.AXNode{{
						Role: "AXButton", Title: "OK",
						Bounds: &computeruse.Rect{X: 10, Y: 10, Width: 50, Height: 20},
					}},
				}},
			},
		}
		postTree := &computeruse.AXTree{
			Root: &computeruse.AXNode{
				Role: "AXApplication", Title: "TestApp",
				Children: []*computeruse.AXNode{{
					Role: "AXWindow", Title: "MainWindow",
					// Note: OK button is GONE.
				}},
			},
		}
		// Sequence: return preTree for the first GetAXTree call,
		// then postTree for subsequent calls.
		callCount := 0
		backend := &sequencedAXBackend{
			treeAt: func(n int) *computeruse.AXTree {
				if n == 0 {
					return preTree
				}
				return postTree
			},
			counter: &callCount,
		}
		cu := computeruse.New(backend)
		gexec := computeruse.NewGatedExecutor(cu, allowAllGatekeeper{})
		r := NewCUResolver(cu, gexec)

		act := &agent.Action{Type: "click", Target: "OK button"}
		res, err := r.Execute(context.Background(), act)
		// We expect ErrStaleState. The mock backend still records
		// the Execute call (the gate passes it; verify aborts
		// AFTER).
		if err == nil {
			t.Fatal("expected ErrStaleState, got nil")
		}
		if !errorsIs(err, computeruse.ErrStaleState) {
			t.Errorf("expected ErrStaleState, got %v", err)
		}
		if res != nil && res.Success {
			t.Error("expected Success=false when verify aborts")
		}
		if callCount < 2 {
			t.Errorf("expected at least 2 AX tree reads (pre+post), got %d", callCount)
		}
	})
}

// errorsIs is a thin alias for errors.Is so the test body reads
// like "if errorsIs(err, target)" without importing the stdlib at
// every call site.
func errorsIs(err, target error) bool { return errors.Is(err, target) }

// sequencedAXBackend returns a different AX tree on each call so
// tests can simulate state changing during an action.
type sequencedAXBackend struct {
	treeAt  func(n int) *computeruse.AXTree
	counter *int
}

func (s *sequencedAXBackend) Name() string { return "sequenced" }
func (s *sequencedAXBackend) Capabilities() []computeruse.Capability {
	return []computeruse.Capability{computeruse.CapAXTree, computeruse.CapClick}
}
func (s *sequencedAXBackend) IsAvailable(_ context.Context) bool { return true }
func (s *sequencedAXBackend) CaptureScreen(_ context.Context) (*computeruse.Screenshot, error) {
	return &computeruse.Screenshot{Width: 100, Height: 100, Image: []byte("x")}, nil
}
func (s *sequencedAXBackend) GetAXTree(_ context.Context) (*computeruse.AXTree, error) {
	n := *s.counter
	*s.counter = n + 1
	return s.treeAt(n), nil
}
func (s *sequencedAXBackend) Execute(_ context.Context, a *computeruse.Action) (*computeruse.ActionResult, error) {
	return &computeruse.ActionResult{Success: true, Action: a, Duration: time.Millisecond}, nil
}

// -----------------------------------------------------------------------------
// requireAXForNonRead — security guardrail for computer-use actions
//
// The function refuses to execute non-READ computer-use actions when the
// pre-action AX tree snapshot is nil (capture failed — no Accessibility
// permission or no AX backend). Without this guard, the agent could
// click/type/launch arbitrary targets without verification of what's
// actually on screen, turning the computer-use flow into a foot-gun.
// -----------------------------------------------------------------------------

func TestRequireAXForNonRead_AXSnapshotPresent_Passes(t *testing.T) {
	r := &CUResolver{}
	pre := &computeruse.Snapshot{} // non-nil = AX captured successfully
	action := &computeruse.Action{Type: computeruse.ActionClick}
	if err := r.requireAXForNonRead(pre, action); err != nil {
		t.Fatalf("requireAXForNonRead with non-nil pre should pass; got %v", err)
	}
}

func TestRequireAXForNonRead_NilAction_Passes(t *testing.T) {
	r := &CUResolver{}
	pre := (*computeruse.Snapshot)(nil)
	if err := r.requireAXForNonRead(pre, nil); err != nil {
		t.Fatalf("requireAXForNonRead with nil action should pass; got %v", err)
	}
}

func TestRequireAXForNonRead_EmptyActionType_Passes(t *testing.T) {
	r := &CUResolver{}
	pre := (*computeruse.Snapshot)(nil)
	action := &computeruse.Action{Type: ""}
	if err := r.requireAXForNonRead(pre, action); err != nil {
		t.Fatalf("requireAXForNonRead with empty Type should pass; got %v", err)
	}
}

func TestRequireAXForNonRead_NonReadWithoutAX_Rejects(t *testing.T) {
	r := &CUResolver{}
	pre := (*computeruse.Snapshot)(nil)
	// ActionClick is non-READ (it modifies UI state) and the AX
	// snapshot is nil. The guard MUST refuse to execute.
	action := &computeruse.Action{Type: computeruse.ActionClick}
	err := r.requireAXForNonRead(pre, action)
	if err == nil {
		t.Fatal("requireAXForNonRead must reject non-READ action without AX snapshot")
	}
	if !errors.Is(err, computeruse.ErrNoBackend) {
		t.Fatalf("expected error wrapping computeruse.ErrNoBackend; got %v", err)
	}
}

// -----------------------------------------------------------------------------
// captureScreenshot — best-effort screenshot capture
//
// captureScreenshot MUST return "" on any failure path (nil shots,
// nil cu, CaptureScreen error, empty image, ScreenshotStore.Put
// error). Screenshots are advisory metadata for the replay store;
// they must NEVER block execution.
//
// Before: 22.2% coverage. The nil-shots and nil-cu early-return
// branches were untested.
// -----------------------------------------------------------------------------

func TestCUResolver_CaptureScreenshot_NilShotsReturnsEmpty(t *testing.T) {
	cu, _ := resolverMocks()
	r := NewCUResolver(cu, nil)
	if got := r.captureScreenshot(context.Background(), "test"); got != "" {
		t.Errorf("captureScreenshot with nil shots = %q, want empty", got)
	}
}

func TestCUResolver_CaptureScreenshot_NilCuReturnsEmpty(t *testing.T) {
	r := &CUResolver{}
	if got := r.captureScreenshot(context.Background(), "test"); got != "" {
		t.Errorf("captureScreenshot with nil cu = %q, want empty", got)
	}
}
