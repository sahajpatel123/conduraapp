// Package perception implements the Selective Perception layer
// from CLAUDE.md §6. It is the unified system that delivers
// safety + battery efficiency + performance + reliability
// simultaneously: every screen perception has a purpose, a TTL,
// and a verification step; the cost of perception is amortized
// across decisions; no perception is wasted on a decision that
// gets aborted.
//
// The package provides:
//
//   - Strategy: the cheapest perception strategy that can answer
//     a given question (None / AXOnly / WindowRect / Differential /
//     FullScreen / VisionCUA). See the SmartCapturer cascade in §6.3.
//
//   - EnergyBudget: the per-session energy budget (Low / Balanced /
//     High / Auto) that gates which strategies are allowed. See §6.4.
//
//   - DirtyTracker: event-driven state-dirty tracking (CGEventTap on
//     macOS, EVENT_OBJECT_* on Windows, AT-SPI signals on Linux).
//     Avoids polling; the agent sleeps when the user is interacting.
//
//   - SmartCapturer: chooses the strategy for a (question, app, dirty
//     state) triple, records the energy used, and refuses if the
//     budget is exhausted.
//
//   - PIIRedactor: regex + known-pattern redaction of screen text
//     before it leaves the machine. Implements §6's "PII redaction
//     in perception pipeline" requirement.
//
// The package is self-contained: no GUI, no network, no LLM. It is
// the leaf that other subsystems (computeruse, agent) call to
// decide "what do I see?" and "is it safe to look?".
package perception

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrBudgetExhausted is returned by SmartCapturer.ChooseStrategy when
// the session energy budget is spent. Per CLAUDE.md §6.4 / decision
// #26, the caller MUST pause and ask the user rather than silently
// retrying or downgrading. Use errors.Is(err, perception.ErrBudgetExhausted)
// to detect it.
var ErrBudgetExhausted = errors.New("perception budget exhausted")

// Strategy is the cost-ordered tier of a screen capture.
// Cheaper strategies consume less energy and battery; more
// expensive ones yield richer information.
type Strategy int

const (
	// StrategyNone is the no-capture tier. The question can be
	// answered from agent state alone (e.g. "what time is it?").
	StrategyNone Strategy = iota
	// StrategyAXOnly reads the accessibility tree without any
	// pixel capture. ~10× cheaper than full screen. The cheapest
	// perception that yields element identity.
	StrategyAXOnly
	// StrategyWindowRect captures only the focused window's rect.
	// ~5× cheaper than full screen.
	StrategyWindowRect
	// StrategyDifferential captures only the rect that changed
	// since last capture. ~3× cheaper when the dirty flag is set.
	StrategyDifferential
	// StrategyFullScreen is the full-screen capture tier. 1×
	// baseline cost; no LLM involved.
	StrategyFullScreen
	// StrategyVisionCUA is the full-screen + vision-model tier.
	// ~50× baseline cost; only when cheaper strategies fail.
	StrategyVisionCUA
)

// Energy costs (relative to StrategyFullScreen baseline = 1.0).
// Numbers come from CLAUDE.md §6.3 (the SmartCapturer cascade).
const (
	costAXOnly       = 0.1
	costWindowRect   = 0.2
	costDifferential = 0.33
	costVisionCUA    = 50.0

	// Session budget fractions of fullScreen-equivalents.
	budgetLow      = 20.0
	budgetBalanced = 50.0
	budgetHigh     = 100.0
)

// String returns the canonical strategy name.
func (s Strategy) String() string {
	switch s {
	case StrategyNone:
		return "none"
	case StrategyAXOnly:
		return "ax_only"
	case StrategyWindowRect:
		return "window_rect"
	case StrategyDifferential:
		return "differential"
	case StrategyFullScreen:
		return "full_screen"
	case StrategyVisionCUA:
		return "vision_cua"
	default:
		return fmt.Sprintf("unknown(%d)", int(s))
	}
}

// EnergyCost returns the relative battery cost of running s once.
// 1.0 = full screen capture baseline.
func (s Strategy) EnergyCost() float64 {
	switch s {
	case StrategyNone:
		return 0.0
	case StrategyAXOnly:
		return costAXOnly
	case StrategyWindowRect:
		return costWindowRect
	case StrategyDifferential:
		return costDifferential
	case StrategyFullScreen:
		return 1.0
	case StrategyVisionCUA:
		return costVisionCUA
	default:
		return 1.0
	}
}

// EnergyMode controls the maximum energy budget per session.
// Per CLAUDE.md §6.4.
type EnergyMode int

const (
	// EnergyAuto picks the right mode based on power state:
	// battery / no charger = Low; plugged in = High.
	EnergyAuto EnergyMode = iota
	// EnergyLow caps the session budget at 20% of the full baseline.
	// AX-only captures are allowed; vision and full-screen are denied.
	EnergyLow
	// EnergyBalanced caps the session budget at 50% of the full
	// baseline. Window-rect and differential captures are allowed.
	EnergyBalanced
	// EnergyHigh caps the session budget at the full 100% baseline.
	// All strategies including vision CUA are allowed.
	EnergyHigh
)

// String returns the canonical energy mode name.
func (e EnergyMode) String() string {
	switch e {
	case EnergyAuto:
		return "auto"
	case EnergyLow:
		return "low"
	case EnergyBalanced:
		return "balanced"
	case EnergyHigh:
		return "high"
	default:
		return fmt.Sprintf("unknown(%d)", int(e))
	}
}

// sessionBudget returns the total energy budget for a session, in
// StrategyFullScreen-equivalent units.
func (e EnergyMode) sessionBudget() float64 {
	switch e {
	case EnergyLow:
		return budgetLow
	case EnergyBalanced:
		return budgetBalanced
	case EnergyHigh:
		return budgetHigh
	case EnergyAuto:
		// Caller should resolve to a concrete mode based on
		// power state before querying budget.
		return budgetBalanced
	default:
		return budgetBalanced
	}
}

// allowedStrategies returns the strategy ceiling for the mode.
func (e EnergyMode) allowedStrategies() map[Strategy]bool {
	switch e {
	case EnergyLow:
		return map[Strategy]bool{
			StrategyNone: true, StrategyAXOnly: true,
		}
	case EnergyBalanced:
		return map[Strategy]bool{
			StrategyNone: true, StrategyAXOnly: true,
			StrategyWindowRect: true, StrategyDifferential: true,
		}
	case EnergyHigh:
		// All strategies allowed.
		return map[Strategy]bool{
			StrategyNone: true, StrategyAXOnly: true,
			StrategyWindowRect: true, StrategyDifferential: true,
			StrategyFullScreen: true, StrategyVisionCUA: true,
		}
	case EnergyAuto:
		// Default to Balanced ceiling until resolved.
		return EnergyBalanced.allowedStrategies()
	default:
		return EnergyBalanced.allowedStrategies()
	}
}

// Question describes what the agent is trying to learn from a
// screen perception. The SmartCapturer uses it to pick a strategy.
type Question struct {
	// Text is the natural-language question ("is the Submit button
	// visible?", "what is in the email body?").
	Text string
	// NeedsElementIdentity: question requires named element info
	// (ax tree, not just pixels).
	NeedsElementIdentity bool
	// NeedsPixels: question requires pixel-level information that
	// AX cannot answer.
	NeedsPixels bool
	// NeedsOCR: question requires reading arbitrary on-screen text.
	NeedsOCR bool
	// TargetApp, if set, scopes perception to a specific app's
	// window rect. Empty means "the whole screen".
	TargetApp string
}

// DirtyState describes whether the screen has changed since the
// last perception. Populated by the DirtyTracker.
type DirtyState struct {
	// Dirty is true if any visual change has been observed.
	Dirty bool
	// SinceLast: time since the last perception that found a change.
	SinceLast time.Duration
	// LastApp is the app that was active at the last dirty event.
	LastApp string
}

// SmartCapturer is the unified perception entry point. Callers
// ask "how should I see?" and the capturer returns a strategy +
// an energy-cost estimate. Callers then actually do the perception
// (via computeruse, vision, etc.) and report back via Record so
// the budget is debited.
type SmartCapturer struct {
	mu        sync.Mutex
	mode      EnergyMode
	used      float64
	budget    float64
	allowed   map[Strategy]bool
	resolved  bool
	lastUsage time.Time
}

// NewSmartCapturer returns a capturer in the given energy mode.
// In EnergyAuto, ResolveAuto should be called once power state
// is known before recording real costs.
func NewSmartCapturer(mode EnergyMode) *SmartCapturer {
	c := &SmartCapturer{
		mode:    mode,
		budget:  mode.sessionBudget(),
		allowed: mode.allowedStrategies(),
	}
	return c
}

// ResolveAuto picks a concrete energy mode based on power state.
// pluggedIn=true → High; pluggedIn=false → Low.
func (c *SmartCapturer) ResolveAuto(pluggedIn bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.mode != EnergyAuto {
		return
	}
	if pluggedIn {
		c.mode = EnergyHigh
	} else {
		c.mode = EnergyLow
	}
	c.budget = c.mode.sessionBudget()
	c.allowed = c.mode.allowedStrategies()
	c.resolved = true
}

// ChooseStrategy returns the cheapest strategy that can answer q,
// respecting the energy budget and dirty state. Returns an error
// if the budget is exhausted; the caller should pause and ask the
// user (per CLAUDE.md §6.4 and decision #26).
func (c *SmartCapturer) ChooseStrategy(q Question, dirty DirtyState) (Strategy, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	preference := strategyPreferenceFor(q, dirty)
	for _, s := range preference {
		if !c.allowed[s] {
			continue
		}
		if c.used+s.EnergyCost() > c.budget {
			continue
		}
		return s, nil
	}
	return StrategyNone, fmt.Errorf("%w (used=%.2f, budget=%.2f, mode=%s)", ErrBudgetExhausted, c.used, c.budget, c.mode)
}

// Record debits the energy used for an executed strategy. Call
// after the actual perception completes (success or failure).
func (c *SmartCapturer) Record(s Strategy) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.used += s.EnergyCost()
	c.lastUsage = time.Now()
}

// Used returns the current energy used this session.
func (c *SmartCapturer) Used() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.used
}

// Budget returns the session budget.
func (c *SmartCapturer) Budget() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.budget
}

// Mode returns the resolved energy mode.
func (c *SmartCapturer) Mode() EnergyMode {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.mode
}

// strategyPreferenceFor returns the candidate strategies in
// cheapest-first order for a given question + dirty state.
// Implements the SmartCapturer cascade from CLAUDE.md §6.3.
//
// StrategyNone is only a candidate when the question can plausibly
// be answered from agent state alone (no flag set). For questions
// about visible UI state, at least AX is required.
func strategyPreferenceFor(q Question, dirty DirtyState) []Strategy {
	// State-only question: agent's own state is enough.
	stateOnly := !q.NeedsElementIdentity && !q.NeedsPixels && !q.NeedsOCR
	// Element identity without pixels: AX is enough.
	if q.NeedsElementIdentity && !q.NeedsPixels && !q.NeedsOCR {
		return []Strategy{
			StrategyAXOnly,
		}
	}
	// Dirty + OCR/visual → differential is the cheapest.
	if dirty.Dirty && (q.NeedsOCR || q.NeedsPixels) {
		out := []Strategy{}
		if stateOnly {
			out = append(out, StrategyNone)
		}
		out = append(out,
			StrategyAXOnly,
			StrategyDifferential,
			StrategyWindowRect,
			StrategyFullScreen,
			StrategyVisionCUA,
		)
		return out
	}
	// Scoped to one app → window rect is the right size.
	if q.TargetApp != "" {
		out := []Strategy{}
		if stateOnly {
			out = append(out, StrategyNone)
		}
		out = append(out,
			StrategyAXOnly,
			StrategyWindowRect,
			StrategyFullScreen,
			StrategyVisionCUA,
		)
		return out
	}
	// Default cascade.
	if stateOnly {
		return []Strategy{
			StrategyNone,
		}
	}
	return []Strategy{
		StrategyAXOnly,
		StrategyWindowRect,
		StrategyFullScreen,
		StrategyVisionCUA,
	}
}

// DirtyTracker is the event-driven state-dirty hook. Other
// subsystems push events into Mark() whenever they observe a
// state change (mouse click, window move, text input, etc.).
// The SmartCapturer reads the latest state via Snapshot().
//
// On macOS the production source is CGEventTap + NSWindowDidUpdate
// notifications. On Windows, EVENT_OBJECT_LOCATIONCHANGE +
// EVENT_OBJECT_NAMECHANGE. On Linux, AT-SPI object:state-changed.
// This implementation is platform-agnostic; the platform-specific
// event source pumps events into Mark().
type DirtyTracker struct {
	mu        sync.Mutex
	dirty     bool
	sinceLast time.Duration
	lastApp   string
	lastEvent time.Time
}

// NewDirtyTracker returns a fresh tracker.
func NewDirtyTracker() *DirtyTracker {
	return &DirtyTracker{
		lastEvent: time.Now(),
	}
}

// Mark records a state-change event. Pass app="" if the
// originating app is unknown.
func (d *DirtyTracker) Mark(app string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	if d.dirty {
		d.sinceLast = now.Sub(d.lastEvent)
	} else {
		d.sinceLast = 0
	}
	d.dirty = true
	d.lastApp = app
	d.lastEvent = now
}

// Clear resets the dirty flag (called by SmartCapturer after a
// successful perception that consumed the dirty event).
func (d *DirtyTracker) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.dirty = false
	d.sinceLast = 0
}

// Snapshot returns the current dirty state.
func (d *DirtyTracker) Snapshot() DirtyState {
	d.mu.Lock()
	defer d.mu.Unlock()
	since := d.sinceLast
	if d.dirty {
		since = time.Since(d.lastEvent)
	}
	return DirtyState{
		Dirty:     d.dirty,
		SinceLast: since,
		LastApp:   d.lastApp,
	}
}

// PerceptionHook is the interface that platform-specific event
// sources implement. internal/computeruse and internal/perception
// callers construct a hook; the GUI wires the platform source to
// the hook at startup.
//
//nolint:revive // stutter is intentional: callers see perception.Hook
type PerceptionHook interface {
	// Observe is called by the platform event source whenever a
	// screen state change is observed. app is the bundle ID /
	// window class / AT-SPI app name; "" if unknown.
	Observe(app string)
}

// PIIRedactor scrubs PII patterns from screen text before it
// leaves the machine. Implements the §6 PII redaction
// requirement at a minimal level: emails, phone numbers, SSNs,
// and credit-card-shaped digit runs.
type PIIRedactor struct {
	// Patterns is the set of regex matchers; tested in order. The
	// first match wins. DefaultPatterns() returns the baseline set.
	Patterns []string
	// Replacement is the string substituted for any matched span.
	// Defaults to "[REDACTED]".
	Replacement string
}

// NewPIIRedactor returns a redactor with the default pattern set.
func NewPIIRedactor() *PIIRedactor {
	return &PIIRedactor{
		Patterns:    DefaultPatterns(),
		Replacement: "[REDACTED]",
	}
}

// DefaultPatterns returns the built-in PII regex set. The exact
// regexes are deliberately conservative; the redactor is a
// first line of defense, not a substitute for the safety layer.
func DefaultPatterns() []string {
	return []string{
		// Email
		`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`,
		// US SSN: 3-2-4 with optional dashes
		`\b\d{3}-\d{2}-\d{4}\b`,
		// Credit card: 13-19 digit run with optional dashes/spaces.
		// Conservative — will over-match on long numeric IDs.
		`\b(?:\d[ -]?){13,19}\b`,
		// International phone: +CC followed by 7-15 digits
		`\+\d{1,3}[ -]?\d{3,4}[ -]?\d{3,4}[ -]?\d{0,4}`,
	}
}

// Redact returns a copy of input with PII patterns replaced.
// This is a string-level operation; call sites that operate on
// screen text MUST redact before crossing any IPC boundary.
func (r *PIIRedactor) Redact(ctx context.Context, input string) (string, error) {
	if r == nil {
		return input, nil
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	out := input
	repl := r.Replacement
	if repl == "" {
		repl = "[REDACTED]"
	}
	for _, pat := range r.Patterns {
		// Pre-compile once per call; the patterns are short and
		// stable so this is acceptable. (For high-volume paths,
		// compile-once would be a small follow-up.)
		out = compileAndReplace(pat, out, repl, out)
	}
	return out, nil
}
