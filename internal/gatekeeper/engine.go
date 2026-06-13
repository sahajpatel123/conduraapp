package gatekeeper

import (
	"context"
	"fmt"
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

	// Pure policy verdict.
	v := e.policy.Load().Evaluate(a)

	// Anomaly hook (9B, stubbed) — records action metadata.

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
	defer e.pendingMu.Unlock()
	for _, t := range e.pending {
		if t.Nonce == nonce {
			t.Approved = true
			if t.Result != nil {
				t.Result <- true
			}
			return true
		}
	}
	return false
}

// DenyTicket denies a pending consent by nonce.
func (e *Engine) DenyTicket(nonce string) bool {
	e.pendingMu.Lock()
	defer e.pendingMu.Unlock()
	for _, t := range e.pending {
		if t.Nonce == nonce {
			t.Approved = false
			if t.Result != nil {
				t.Result <- false
			}
			return true
		}
	}
	return false
}

// ReloadPolicy atomically swaps the policy.
func (e *Engine) ReloadPolicy(p *Policy) {
	e.policy.Store(p)
}

// ConsentTicket represents an in-flight consent request.
type ConsentTicket struct {
	ActionKind string    `json:"action_kind"`
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
