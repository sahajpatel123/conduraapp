package gatekeeper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/blastradius"
)

// ConsentProvider drives the OS-level consent dialog. Implementations:
//   - GUIProvider: daemon → SSE → Wails modal → RPC
//   - osascriptProvider: AppleScript display dialog (headless fallback)
//   - noopProvider: deny-on-absence (Windows/Linux headless stub)
type ConsentProvider interface {
	Show(ctx context.Context, ticket *ConsentTicket) (approved bool, err error)
	IsAvailable() bool
}

// HaltChecker is the subset of halt.Flag the Engine needs.
type HaltChecker interface {
	IsHalted() bool
}

// Engine implements gatekeeper.Gatekeeper with a real policy engine.
// Internally uses pure Policy.Evaluate for the verdict, then drives
// the ConsentProvider for consent-required actions.
type Engine struct {
	policy  *AtomicPolicy
	consent ConsentProvider
	halt    HaltChecker
	// pending holds in-flight consent tickets for GUI enumeration.
	pendingMu sync.Mutex
	pending   []*ConsentTicket
	// hooks for 9B-9E (stubbed until those sub-phases).
	AnomalyHook  func(a blastradius.Action)
	SanitizeHook func(a *blastradius.Action) error
	// AutonomyHook returns the current autonomy level. If nil, autonomy is disabled.
	AutonomyHook func(taskType, app string) (level int)
	// SensitiveHook checks whether a target URL or context is on a
	// sensitive site (banking, health, government). If true, the
	// decision is escalated to RequirePresenceAndConsent.
	SensitiveHook func(url, context string) bool
	// TrustHook (Phase 16, Rec 5): if the action targets a
	// workspace the user has marked as "always allow in this
	// folder", returns Allow. Only consulted for WRITE-class
	// actions in non-DESTRUCTIVE context — DESTRUCTIVE always
	// requires fresh consent per Survival Rule §2.
	//
	// Signature: (workspaceID, app) → trusted? The hook is also
	// given a "first-encounter" flag so the engine can build the
	// consent ticket that lets the user pick "Always allow".
	TrustHook func(workspaceID, app string) (entry any, ok bool)
}

// NewEngine creates the real Gatekeeper engine.
func NewEngine(policy *Policy, consent ConsentProvider, halt HaltChecker) *Engine {
	e := &Engine{consent: consent, halt: halt}
	e.policy = &AtomicPolicy{}
	e.policy.Store(policy)
	return e
}

// Evaluate implements gatekeeper.Gatekeeper. Returns Allow or Deny.
func (e *Engine) Evaluate(ctx context.Context, a blastradius.Action) (Decision, string) {
	// Sanitize (9C hook, stubbed).
	if e.SanitizeHook != nil {
		if err := e.SanitizeHook(&a); err != nil {
			return Deny, fmt.Sprintf("sanitizer blocked: %v", err)
		}
	}

	// Anomaly (9B hook, stubbed).
	if e.AnomalyHook != nil {
		e.AnomalyHook(a)
	}

	// Sensitive site check: escalate any action on banking/health
	// sites to RequirePresenceAndConsent before evaluating policy.
	if e.SensitiveHook != nil {
		if e.SensitiveHook(a.TargetURL, a.Body) {
			return RequirePresenceAndConsent, "sensitive site: escalated to presence-and-consent"
		}
	}

	// Pure policy verdict first. Autonomy can only bypass consent,
	// not explicit policy deny rules.
	v := e.policy.Load().Evaluate(a)

	// Autonomy pre-check (9E): if autonomous level, bypass consent
	// for allowed classes. DESTRUCTIVE always requires consent.
	if e.AutonomyHook != nil {
		lvl := e.AutonomyHook(a.Kind, a.TargetApp)
		if lvl == 0 { // Block
			return Deny, fmt.Sprintf("autonomy: blocked for %s/%s", a.Kind, a.TargetApp)
		}
		// autonomous: check ApplyAutonomous
		// Note: lvl=0=Block, 1=Warn, 2=Ask, 3=Autonomous
		if lvl >= 3 { // Autonomous
			class := blastradius.Classify(a)
			if class != blastradius.DESTRUCTIVE &&
				(v.Decision == RequireConsent || v.Decision == RequirePresenceAndConsent) {
				return Allow, "autonomous: auto-allowed (non-destructive consent-required)"
			}
		}
	}

	// Phase 16, Rec 5: per-workspace trust. If the user has marked
	// the target workspace as "always allow in this folder", we
	// skip the consent dialog for WRITE actions. DESTRUCTIVE
	// always requires fresh consent.
	if e.TrustHook != nil {
		class := blastradius.Classify(a)
		if class == blastradius.WRITE && a.Path != "" {
			wsID := workspaceIDFor(a.Path)
			if wsID != "" {
				if _, ok := e.TrustHook(wsID, a.TargetApp); ok {
					return Allow, "workspace trust: always-allow in this folder"
				}
			}
		}
	}

	// Direct decisions.
	if v.Decision == Allow {
		return Allow, v.Reason
	}
	if v.Decision == Deny {
		return Deny, v.Reason
	}

	// Consent-required: drive the consent provider.
	return e.evaluateConsent(ctx, a, v)
}

func (e *Engine) evaluateConsent(ctx context.Context, a blastradius.Action, v Verdict) (Decision, string) {
	// Halt check.
	if e.halt != nil && e.halt.IsHalted() {
		return Deny, "halted: consent not available"
	}

	// No consent provider → deny (fail-closed).
	if e.consent == nil || !e.consent.IsAvailable() {
		return Deny, "consent required but no provider available"
	}

	// Create consent ticket.
	ticket := &ConsentTicket{
		ActionKind: a.Kind,
		Verdict:    v,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(time.Duration(v.TimeoutSecs) * time.Second),
		Nonce:      fmt.Sprintf("%d", time.Now().UnixNano()),
		Result:     make(chan bool, 1),
	}

	// Track pending ticket for GUI enumeration.
	e.pendingMu.Lock()
	e.pending = append(e.pending, ticket)
	e.pendingMu.Unlock()
	defer func() {
		e.pendingMu.Lock()
		e.pending = removeTicket(e.pending, ticket)
		e.pendingMu.Unlock()
	}()

	// Block on consent.
	approved, err := e.consent.Show(ctx, ticket)
	if err != nil {
		return Deny, fmt.Sprintf("consent error: %v", err)
	}
	ticket.Approved = approved

	if approved {
		return Allow, "consent granted"
	}
	return Deny, "consent denied"
}

// Pending returns in-flight consent tickets for GUI enumeration.
func (e *Engine) Pending() []*ConsentTicket {
	e.pendingMu.Lock()
	defer e.pendingMu.Unlock()
	out := make([]*ConsentTicket, len(e.pending))
	copy(out, e.pending)
	return out
}

// ApproveTicket approves a pending consent by nonce.
func (e *Engine) ApproveTicket(nonce string) bool {
	e.pendingMu.Lock()
	var ticket *ConsentTicket
	for _, t := range e.pending {
		if t.Nonce == nonce {
			t.Approved = true
			ticket = t
			break
		}
	}
	e.pendingMu.Unlock()
	if ticket == nil {
		return false
	}
	if ticket.Result != nil {
		select {
		case ticket.Result <- true:
		default:
		}
	}
	return true
}

// DenyTicket denies a pending consent by nonce.
func (e *Engine) DenyTicket(nonce string) bool {
	e.pendingMu.Lock()
	var ticket *ConsentTicket
	for _, t := range e.pending {
		if t.Nonce == nonce {
			t.Approved = false
			ticket = t
			break
		}
	}
	e.pendingMu.Unlock()
	if ticket == nil {
		return false
	}
	if ticket.Result != nil {
		select {
		case ticket.Result <- false:
		default:
		}
	}
	return true
}

// ReloadPolicy atomically swaps the policy.
func (e *Engine) ReloadPolicy(p *Policy) {
	e.policy.Store(p)
}

// SetConsentProvider swaps the consent provider (for testing).
func (e *Engine) SetConsentProvider(c ConsentProvider) {
	e.consent = c
}

// workspaceIDFor returns the canonical workspace ID for a path.
// Walks up from path looking for a .git/ directory; if found,
// returns its absolute path. Otherwise returns the absolute path
// of the input (most conservative).
//
// Phase 16, Rec 5: matches the heuristic in internal/trust so
// the gatekeeper's lookup key is consistent with what the trust
// store was populated with. Inlined here to avoid an import
// cycle (trust → safety → gatekeeper; gatekeeper cannot import
// trust without going through safety, which we want to keep
// optional).
func workspaceIDFor(path string) string {
	if path == "" {
		return ""
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	cur := abs
	for {
		gitPath := filepath.Join(cur, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return cur
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return abs
}

// ConsentTicket represents an in-flight consent request.
type ConsentTicket struct {
	ActionKind string    `json:"action_kind"`
	Actor      string    `json:"actor"`
	Detail     string    `json:"detail"`
	Verdict    Verdict   `json:"verdict"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	Nonce      string    `json:"nonce"`
	Approved   bool      `json:"approved"`
	Result     chan bool `json:"-"`
}

func removeTicket(tickets []*ConsentTicket, target *ConsentTicket) []*ConsentTicket {
	for i, t := range tickets {
		if t == target {
			return append(tickets[:i], tickets[i+1:]...)
		}
	}
	return tickets
}
