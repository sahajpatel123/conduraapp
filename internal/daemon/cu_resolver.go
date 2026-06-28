package daemon

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/agent"
	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/computeruse"
	"github.com/sahajpatel123/synapticapp/internal/perception"
	"github.com/sahajpatel123/synapticapp/internal/replay"
)

// defaultActionTimeout is the fallback timeout for resolved actions.
const defaultActionTimeout = 5 * time.Second

// Well-known AX role strings used by the target resolver.
const (
	axButton      = "AXButton"
	axTextField   = "AXTextField"
	axTextArea    = "AXTextArea"
	axCheckBox    = "AXCheckBox"
	axRadioButton = "AXRadioButton"
	axMenu        = "AXMenu"
	axMenuItem    = "AXMenuItem"
	axWindow      = "AXWindow"
	axTab         = "AXTab"
	axList        = "AXList"
	axScrollBar   = "AXScrollBar"
	axSlider      = "AXSlider"
	axComboBox    = "AXComboBox"
	axPopUpButton = "AXPopUpButton"
	axLink        = "AXLink"
	axImage       = "AXImage"
	axTable       = "AXTable"
	axRow         = "AXRow"
	axStaticText  = "AXStaticText"
)

// CUResolver resolves agent.Action (high-level planner intent) into
// computeruse.Action (low-level executable command) and executes it
// through the GatedExecutor.
//
// The resolver is the bridge between the planner's abstract world
// ("click the Submit button") and the OS's concrete world (AX element
// "AXButton" titled "Submit" at process 1234 with bounds {100,200,80,30}).
//
// It wraps the GatedComputerUseExecutor so every resolved action
// still passes through the Gatekeeper before physical execution.
//
// When a perception.SmartCapturer is wired via SetCapturer, every
// AX-tree + screenshot capture is preceded by a strategy choice and
// followed by an energy debit. When the session energy budget is
// exhausted (per §6.4 / decision #26), Execute aborts with
// perception.ErrBudgetExhausted so the caller can pause and ask the
// user rather than silently draining the battery.
type CUResolver struct {
	cu       *computeruse.ComputerUse
	gate     *computeruse.GatedExecutor
	shots    *replay.ScreenshotStore
	onCUStep func(kind string, x, y float64, success bool)
	capturer *perception.SmartCapturer
}

// NewCUResolver creates an ActionResolver that bridges the agent
// and computer-use type systems.
func NewCUResolver(cu *computeruse.ComputerUse, gate *computeruse.GatedExecutor) *CUResolver {
	return &CUResolver{cu: cu, gate: gate}
}

// SetScreenshotStore wires the replay screenshot store so
// before/after screenshots are captured for every CU action.
func (r *CUResolver) SetScreenshotStore(shots *replay.ScreenshotStore) {
	r.shots = shots
}

// SetAnomalyHook wires the anomaly detector to fire on every CU step.
func (r *CUResolver) SetAnomalyHook(fn func(kind string, x, y float64, success bool)) {
	r.onCUStep = fn
}

// SetCapturer wires the Selective Perception energy-budget capturer.
// When set, every AX/screenshot capture consults the capturer for a
// strategy and debits the session energy budget. When the budget is
// exhausted, Execute aborts with perception.ErrBudgetExhausted.
func (r *CUResolver) SetCapturer(c *perception.SmartCapturer) {
	r.capturer = c
}

// chooseAndRecordPerception picks a perception strategy for the given
// action + dirty state, debits the budget, and returns the chosen
// strategy. A nil capturer short-circuits to StrategyNone (the
// v0.1.0 path before perception was wired).
func (r *CUResolver) chooseAndRecordPerception(a *agent.Action, dirty perception.DirtyState) (perception.Strategy, error) {
	if r.capturer == nil {
		return perception.StrategyNone, nil
	}
	q := perception.Question{
		Text:                 a.Type + " " + a.Target,
		NeedsElementIdentity: true, // CU always needs the AX tree to find the target
		NeedsPixels:          a.Type == "screenshot" || a.Type == "describe",
		NeedsOCR:             false,
		TargetApp:            "",
	}
	strategy, err := r.capturer.ChooseStrategy(q, dirty)
	if err != nil {
		return perception.StrategyNone, fmt.Errorf("perception: %w", err)
	}
	r.capturer.Record(strategy)
	return strategy, nil
}

// Execute resolves an agent-level action into a computer-use
// executable command, runs it through the GatedExecutor, and
// maps the result back to agent.StepResult.
func (r *CUResolver) Execute(ctx context.Context, a *agent.Action) (*agent.StepResult, error) {
	cuAction, err := r.resolve(ctx, a)
	if err != nil {
		return &agent.StepResult{Success: false, Error: err}, err
	}

	// Phase 17, Fix #6 (B4): twin-snapshot verification. The
	// pre/post AX tree comparison is the anti-staleness mechanism
	// from CLAUDE.md §5.2 / Survival Rule §2.1.2. Without it,
	// the agent plays darts with the OS — by the time the click
	// fires, the screen has changed and we hit the wrong target.
	//
	// We take the pre-snapshot, run the gated action, take the
	// post-snapshot, compare; if the diff is critical (window
	// focus changed, target node removed) we ABORT the action
	// and return ErrStaleState so the planner can retry with a
	// fresh AX tree.
	//
	// Selective Perception (§6): before capturing, consult the
	// energy-budget capturer for a strategy. When the budget is
	// exhausted, abort with a clear error so the caller can pause
	// and ask the user (decision #26: refuse, force user decision).
	dirty := perception.DirtyState{}
	if _, perr := r.chooseAndRecordPerception(a, dirty); perr != nil {
		return &agent.StepResult{
			Success: false,
			Error:   perr,
			Output:  "selective perception budget exhausted — pause and ask the user before retrying",
		}, perr
	}

	pre := r.captureAXSnapshot(ctx)

	// Fail closed for non-READ actions if the AX tree couldn't be
	// captured (accessibility permission missing, no backends).
	if err := r.requireAXForNonRead(pre, cuAction); err != nil {
		return &agent.StepResult{
			Success: false,
			Error:   err,
			Output:  err.Error(),
		}, err
	}

	ssBeforeRef := r.captureScreenshot(ctx, "before")
	result, err := r.gate.Execute(ctx, cuAction)
	ssAfterRef := r.captureScreenshot(ctx, "after")

	// Post-action twin-snapshot verification.
	if aborted := r.verifyPostAction(ctx, pre, cuAction, result, ssBeforeRef, ssAfterRef); aborted != nil {
		return aborted, computeruse.ErrStaleState
	}

	// Anomaly recording: real coordinates from CU action.
	if r.onCUStep != nil && result != nil {
		x, y := 0.0, 0.0
		if cuAction.Bounds != nil {
			x, y = cuAction.Bounds.X, cuAction.Bounds.Y
		}
		r.onCUStep(string(cuAction.Type), x, y, result.Success)
	}

	if err != nil {
		return &agent.StepResult{
			Success:     result != nil && result.Success,
			Error:       err,
			Output:      errorText(result),
			Duration:    durationSeconds(result),
			SSBeforeRef: ssBeforeRef,
			SSAfterRef:  ssAfterRef,
		}, err
	}

	return &agent.StepResult{
		Success:     result.Success,
		Output:      describeResult(result),
		Duration:    durationSeconds(result),
		SSBeforeRef: ssBeforeRef,
		SSAfterRef:  ssAfterRef,
	}, nil
}

// captureAXSnapshot takes a pre/post AX tree for twin-snapshot
// verification. Best-effort: returns nil if the AX backend is
// unavailable or the tree is empty, in which case the verifier
// is skipped (the action still runs).
func (r *CUResolver) captureAXSnapshot(ctx context.Context) *computeruse.Snapshot {
	if r.cu == nil {
		return nil
	}
	tree, err := r.cu.GetAXTree(ctx)
	if err != nil || tree == nil {
		return nil
	}
	return computeruse.NewSnapshot(tree, 0)
}

// captureScreenshot takes a screenshot via the CU backend and stores
// it in the replay screenshot store. Returns the ref ID, or "" on
// any failure (best-effort; screenshots must never block execution).
func (r *CUResolver) captureScreenshot(ctx context.Context, position string) string {
	if r.shots == nil || r.cu == nil {
		return ""
	}
	ss, serr := r.cu.CaptureScreen(ctx)
	if serr != nil || ss == nil || len(ss.Image) == 0 {
		return ""
	}
	ref, perr := r.shots.Put(ctx, position, ss.Width, ss.Height, ss.Image)
	if perr != nil {
		return ""
	}
	return ref
}

// resolve converts an agent.Action (planner intent) into a
// computeruse.Action (executable command). It:
//  1. Parses the Type string into a typed enum (error on unknown verb)
//  2. Resolves the Target string into an AX *Target (error if unresolvable)
//  3. Fills execution context (timeout default, value passthrough)
//
//nolint:unparam // ctx is reserved for future AX-resolution queries
func (r *CUResolver) resolve(_ context.Context, a *agent.Action) (*computeruse.Action, error) {
	// Step 1: parse the verb.
	actType, err := parseActionType(a.Type)
	if err != nil {
		return nil, err
	}

	// Step 2: resolve the target descriptor into an AX element query.
	var target *computeruse.Target
	if a.Target != "" {
		t, err := resolveTarget(a.Target)
		if err != nil {
			return nil, err
		}
		target = t
	}

	// Step 3: build the executable action.
	return &computeruse.Action{
		Type:    actType,
		Target:  target,
		Value:   a.Value,
		Timeout: defaultActionTimeout,
	}, nil
}

// parseActionType converts a planner-emitted verb string into the
// typed computeruse.ActionType enum. Unknown verbs return an error
// so the planner knows the action can't be executed.
func parseActionType(verb string) (computeruse.ActionType, error) {
	canonical := strings.ToLower(strings.TrimSpace(verb))
	switch canonical {
	case "click":
		return computeruse.ActionClick, nil
	case "type":
		return computeruse.ActionTypeText, nil
	case "scroll":
		return computeruse.ActionScroll, nil
	case "key_press", "keypress":
		return computeruse.ActionKeyPress, nil
	case "drag":
		return computeruse.ActionDrag, nil
	case "launch":
		return computeruse.ActionLaunch, nil
	case "focus":
		return computeruse.ActionFocus, nil
	case "wait":
		return computeruse.ActionWait, nil
	default:
		return "", fmt.Errorf("cu_resolver: unknown action type %q", verb)
	}
}

// resolveTarget parses a planner-emitted target description string
// into an AX element query. The caller is expected to check for
// empty strings before calling this function.
//
//nolint:unparam // error return is reserved for future AX-validation
func resolveTarget(desc string) (*computeruse.Target, error) {
	// Parse the description into role + title components.
	// Examples:
	//   "Submit button"           → role=button, title=Submit
	//   "password field"          → role=textfield, title=password
	//   "the OK button in dialog" → role=button, title=OK
	role, title := parseTargetDescriptor(desc)

	target := &computeruse.Target{
		Role:  role,
		Title: title,
	}
	return target, nil
}

// parseTargetDescriptor heuristically pulls an AX role and element
// title from a natural-language target description. This is a
// lightweight parser; the full vision-based resolution lives in
// the Vision CUA backend (7E).
func parseTargetDescriptor(desc string) (role, title string) {
	// Normalize: lowercase, strip articles.
	normalized := strings.ToLower(strings.TrimSpace(desc))
	normalized = strings.TrimPrefix(normalized, "the ")
	normalized = strings.TrimPrefix(normalized, "a ")
	normalized = strings.TrimPrefix(normalized, "an ")

	// Known role keywords and their AX role mappings.
	// Sorted longest-first so "menu item" matches before "menu".
	type kwEntry struct {
		keyword string
		axRole  string
	}
	roleKeywords := []kwEntry{
		{"radio button", axRadioButton},
		{"scroll bar", axScrollBar},
		{"combo box", axComboBox},
		{"menu item", axMenuItem},
		{"text field", axTextField},
		{"text area", axTextArea},
		{"textfield", axTextField},
		{"button", axButton},
		{"checkbox", axCheckBox},
		{"dropdown", axPopUpButton},
		{"dialog", axWindow},
		{"window", axWindow},
		{"toggle", axCheckBox},
		{"slider", axSlider},
		{"field", axTextField},
		{"input", axTextField},
		{"radio", axRadioButton},
		{"label", axStaticText},
		{"table", axTable},
		{"image", axImage},
		{"scroll", axScrollBar},
		{"combo", axComboBox},
		{"menu", axMenu},
		{"link", axLink},
		{"list", axList},
		{"tab", axTab},
		{"row", axRow},
	}

	for _, rk := range roleKeywords {
		if !strings.Contains(normalized, rk.keyword) {
			continue
		}
		role = rk.axRole

		// Everything before the role keyword is the title.
		// "submit button" → title = "submit"
		idx := strings.Index(normalized, rk.keyword)
		prefix := strings.TrimSpace(normalized[:idx])
		if prefix != "" {
			title = prefix
			return role, title
		}

		// No prefix: the role keyword starts the string.
		// Single-word keywords (e.g. "input", "checkbox") carry
		// that word as the title since it's the descriptor.
		// Compound keywords (e.g. "menu item", "text field") with
		// no extra text → role-only match (title empty).
		if normalized == rk.keyword {
			if !strings.Contains(rk.keyword, " ") {
				title = rk.keyword
			}
			return role, title
		}

		// Partial prefix: the role keyword starts at 0 but the
		// string has more text (e.g. "button OK" → the keyword
		// is embedded, not standalone). Use the remainder as title.
		title = strings.TrimSpace(normalized[len(rk.keyword):])
		return role, title
	}

	// No role keyword found. Use the full descriptor as title
	// and leave role empty for generic AX matching.
	return "", normalized
}

// durationSeconds converts a time.Duration to float64 seconds.
func durationSeconds(result *computeruse.ActionResult) float64 {
	if result == nil {
		return 0
	}
	return result.Duration.Seconds()
}

// errorText returns the error message from a result, or empty string.
func errorText(result *computeruse.ActionResult) string {
	if result == nil || result.Error == nil {
		return ""
	}
	return result.Error.Error()
}

// verifyPostAction runs twin-snapshot verification after the action
// has executed. Returns nil if verification passed or was skipped
// (pre is nil = no AX tree available). Returns an abort result if
// the post-action snapshot shows a critical state change (window
// focus moved, target node removed).
func (r *CUResolver) verifyPostAction(
	ctx context.Context,
	pre *computeruse.Snapshot,
	cuAction *computeruse.Action,
	result *computeruse.ActionResult,
	ssBeforeRef, ssAfterRef string,
) *agent.StepResult {
	if pre == nil || r.cu == nil {
		return nil
	}
	postTree, err := r.cu.GetAXTree(ctx)
	if err != nil || postTree == nil {
		return nil
	}
	post := computeruse.NewSnapshot(postTree, 0)
	vres := computeruse.VerifySnapshots(pre, post, cuAction)
	if vres != nil && !vres.Valid && vres.Aborted {
		return &agent.StepResult{
			Success:     false,
			Error:       computeruse.ErrStaleState,
			Output:      "twin-snapshot verification aborted: " + vres.Reason,
			Duration:    durationSeconds(result),
			SSBeforeRef: ssBeforeRef,
			SSAfterRef:  ssAfterRef,
		}
	}
	return nil
}

// requireAXForNonRead checks whether the given pre-action AX snapshot is nil
// (capture failed — accessibility permission missing, no backends) for a
// non-READ action. If so, it returns an error because we cannot safely verify
// what the agent is about to click.
func (r *CUResolver) requireAXForNonRead(pre *computeruse.Snapshot, cuAction *computeruse.Action) error {
	if pre != nil || cuAction == nil || cuAction.Type == "" {
		return nil
	}
	ba := cuAction.ToBlastRadius()
	if blastradius.Classify(ba) <= blastradius.READ {
		return nil
	}
	return fmt.Errorf("%w: AX tree capture failed — refusing to execute %s action without verification; grant Accessibility permission and retry", computeruse.ErrNoBackend, cuAction.Type)
}

// describeResult produces a human-readable summary of the result.
func describeResult(result *computeruse.ActionResult) string {
	if result == nil {
		return ""
	}
	if result.Success {
		dur := result.Duration.Round(time.Millisecond)
		return fmt.Sprintf("action completed in %v", dur)
	}
	if result.Error != nil {
		return fmt.Sprintf("action failed: %v", result.Error)
	}
	return "action completed"
}
