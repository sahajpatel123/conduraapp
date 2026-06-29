package gatekeeper

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/blastradius"
)

// ConsentProvider drives the OS-level consent dialog. Implementations:
//   - rpcConsentProvider (internal/daemon/safety_wiring.go):
//     daemon → SSE → Wails modal → RPC. The only production
//     implementation shipped in v0.1.0.
//   - autoApproveConsentProvider (internal/daemon/safety_wiring.go):
//     test-only, gated behind the SYNAPTIC_TEST_AUTO_CONSENT env var.
//
// Not implemented in v0.1.0 (would ship in v0.2.x):
//   - osascriptProvider: AppleScript display dialog for headless
//     macOS daemon use cases where the GUI is unavailable.
//   - Windows / Linux native modal providers.
//   - noopProvider: deny-on-absence for headless CI runs.
//
// When no provider is wired, evaluateConsent fails closed (Deny)
// per MISSION §2 — the engine returns "consent required but no
// provider available" and the action is blocked.
type ConsentProvider interface {
	Show(ctx context.Context, ticket *ConsentTicket) (approved bool, err error)
	IsAvailable() bool
}

// HaltChecker is the subset of halt.Flag the Engine needs.
type HaltChecker interface {
	IsHalted() bool
}

// PresenceChecker reports whether the user is present at the machine.
// The Engine consults it before showing a consent modal for actions
// that require an active user (RequirePresenceAndConsent, or any rule
// with require_user_active:true). nil means presence-gating is not
// configured; the Engine then falls back to the consent modal's
// timeout-queue backstop (the modal blocks until clicked; on timeout
// the action queues and does not execute) — safe, just less
// immediate. N1: previously the presence detector was dead code and
// require_user_active was parsed but never read.
type PresenceChecker interface {
	IsPresent() bool
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
	// AnomalyHook receives the action AND the decision the engine
	// actually produced. The success flag was previously always false
	// at the only call site (safety_wiring.go), which turned the
	// §5.6 "5+ consecutive failures" trigger into a false-positive
	// machine that halted the agent after any 5 actions through the
	// gate. The hook now carries the verdict so the anomaly detector
	// can distinguish a successful Allow from a Deny.
	AnomalyHook  func(a blastradius.Action, d Decision, reason string)
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
	// PresenceChecker (N1): if non-nil, the Engine consults it before
	// showing the consent modal for RequirePresenceAndConsent or
	// require_user_active actions; an absent user is denied (action
	// held for safety). nil → fall back to the modal-timeout-queue
	// backstop (safe). Set via SetPresenceChecker.
	PresenceChecker PresenceChecker
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
			d, r := Deny, fmt.Sprintf("sanitizer blocked: %v", err)
			e.notifyAnomaly(a, d, r)
			return d, r
		}
	}

	// Sensitive site check: escalate any action on banking/health
	// sites to RequirePresenceAndConsent before evaluating policy.
	if e.SensitiveHook != nil {
		if e.SensitiveHook(a.TargetURL, a.Body) {
			d, r := RequirePresenceAndConsent, "sensitive site: escalated to presence-and-consent"
			e.notifyAnomaly(a, d, r)
			return d, r
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
			d, r := Deny, fmt.Sprintf("autonomy: blocked for %s/%s", a.Kind, a.TargetApp)
			e.notifyAnomaly(a, d, r)
			return d, r
		}
		// autonomous: check ApplyAutonomous
		// Note: lvl=0=Block, 1=Warn, 2=Ask, 3=Autonomous
		if lvl >= 3 { // Autonomous
			class := blastradius.Classify(a)
			if class != blastradius.DESTRUCTIVE &&
				(v.Decision == RequireConsent || v.Decision == RequirePresenceAndConsent) {
				d, r := Allow, "autonomous: auto-allowed (non-destructive consent-required)"
				e.notifyAnomaly(a, d, r)
				return d, r
			}
		}
	}

	if reason, ok := e.applyWorkspaceTrust(a); ok {
		d, r := Allow, reason
		e.notifyAnomaly(a, d, r)
		return d, r
	}

	// Direct decisions.
	if v.Decision == Allow {
		d, r := Allow, v.Reason
		e.notifyAnomaly(a, d, r)
		return d, r
	}
	if v.Decision == Deny {
		d, r := Deny, v.Reason
		e.notifyAnomaly(a, d, r)
		return d, r
	}

	// Consent-required: drive the consent provider.
	d, r := e.evaluateConsent(ctx, a, v)
	e.notifyAnomaly(a, d, r)
	return d, r
}

// notifyAnomaly invokes the AnomalyHook with the action and the
// engine's actual verdict. This is the fix for the "false-positive
// machine" audit finding: the hook used to fire before the verdict
// with success always false, so any 5 actions through the gate
// tripped the §5.6 "5+ consecutive failures" trigger. The hook now
// sees the real outcome — Allow on a successful allowed action
// records success=true, Deny records success=false — which is the
// behavior the spec describes.
func (e *Engine) notifyAnomaly(a blastradius.Action, d Decision, reason string) {
	if e.AnomalyHook == nil {
		return
	}
	e.AnomalyHook(a, d, reason)
}

// applyWorkspaceTrust is Phase 16, Rec 5: per-workspace trust.
// If the user has marked the target workspace as "always allow in
// this folder", we skip the consent dialog for WRITE actions.
// DESTRUCTIVE always requires fresh consent. Extracted from
// Evaluate to keep the parent function under the
// cyclomatic-complexity cap.
//
// Returns (reason, true) if trust applies; the decision is
// always Allow in that case (encoded in the caller).
func (e *Engine) applyWorkspaceTrust(a blastradius.Action) (string, bool) {
	if e.TrustHook == nil {
		return "", false
	}
	class := blastradius.Classify(a)
	if class != blastradius.WRITE || a.Path == "" {
		return "", false
	}
	wsID := workspaceIDFor(a.Path)
	if wsID == "" {
		return "", false
	}
	if _, ok := e.TrustHook(wsID, a.TargetApp); !ok {
		return "", false
	}
	return "workspace trust: always-allow in this folder", true
}

// presenceDenied implements the N1 presence gate. It returns
// (true, reason) when the action must be denied because the user is
// absent; (false, "") to proceed to the consent modal.
//
// Presence is required when the decision is RequirePresenceAndConsent
// (DESTRUCTIVE) or the rule set require_user_active:true. If a
// PresenceChecker is wired and reports the user absent, the action is
// denied immediately (held for safety; true auto-re-prompt-on-return
// is v0.2.0). If no checker is wired, it returns false so the caller
// falls back to the consent modal's timeout-queue backstop — safe, and
// keeps existing tests that don't wire a checker on the unchanged modal
// path.
func (e *Engine) presenceDenied(v Verdict) (bool, string) {
	presenceRequired := v.Decision == RequirePresenceAndConsent || v.RequireActive
	if !presenceRequired || e.PresenceChecker == nil {
		return false, ""
	}
	if e.PresenceChecker.IsPresent() {
		return false, ""
	}
	if v.OnUserAbsent == onUserAbsentQueue {
		return true, "queued: user absent — action held; retry when present"
	}
	return true, "user absent: action held for safety"
}

func (e *Engine) evaluateConsent(ctx context.Context, a blastradius.Action, v Verdict) (Decision, string) {
	// Halt check.
	if e.halt != nil && e.halt.IsHalted() {
		return Deny, "halted: consent not available"
	}

	// N1: presence gate. presenceDenied denies immediately when a
	// checker is wired and the user is absent; when no checker is
	// wired it returns false so we fall back to the consent modal's
	// timeout-queue backstop (the modal blocks until a human clicks,
	// and on timeout the action queues and does not execute) — safe,
	// and keeps existing tests (which don't wire a checker) unchanged.
	if denied, reason := e.presenceDenied(v); denied {
		return Deny, reason
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
		Nonce:      generateNonce(),
		Result:     make(chan bool, 1),
	}

	// Track pending ticket for GUI enumeration.
	e.pendingMu.Lock()
	e.pending = append(e.pending, ticket)
	e.pendingMu.Unlock()

	// Phase 17, Fix #5 (A7): when the policy says OnTimeout=queue,
	// the engine-side timeout is suppressed so the consent dialog
	// has time to gather a real answer (especially useful for
	// DESTRUCTIVE actions where "user walked away" should NOT auto-
	// deny). The caller's context is the only thing that cancels
	// the wait. The ticket also stays in `pending` across the
	// wait so the GUI can surface it and let the user approve or
	// deny it manually if the dialog itself times out.
	queueOnTimeout := v.OnTimeout == "queue"
	if !queueOnTimeout {
		defer func() {
			e.pendingMu.Lock()
			e.pending = removeTicket(e.pending, ticket)
			e.pendingMu.Unlock()
		}()
	}

	// Block on consent. Run the provider in a goroutine so we
	// can enforce the engine-side timeout here (the provider may
	// return early on its own clock, but if it doesn't, we
	// honor OnTimeout here).
	type providerResult struct {
		approved bool
		err      error
	}
	done := make(chan providerResult, 1)
	go func() {
		ap, err := e.consent.Show(ctx, ticket)
		done <- providerResult{approved: ap, err: err}
	}()
	var (
		approved bool
		err      error
	)
	if v.OnTimeout == "queue" {
		// Block until provider returns, OR caller context is canceled.
		select {
		case res := <-done:
			approved, err = res.approved, res.err
		case <-ctx.Done():
			// Leave the ticket in pending so the GUI can resolve
			// it later via ApproveTicket/DenyTicket.
			return Deny, "queued: waiting for user response (ctx canceled)"
		}
	} else {
		// Default: race provider against engine-side timeout.
		t := time.NewTimer(time.Duration(v.TimeoutSecs) * time.Second)
		defer t.Stop()
		select {
		case res := <-done:
			approved, err = res.approved, res.err
		case <-t.C:
			return Deny, fmt.Sprintf("consent timeout after %ds", v.TimeoutSecs)
		case <-ctx.Done():
			return Deny, "consent canceled"
		}
	}
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

// ApproveTicket approves a pending consent by nonce. Returns
// false if the nonce is unknown OR if the ticket has already
// expired (ExpiresAt is in the past). Phase 17, Fix #2 (A6):
// expired nonces can no longer be replayed.
func (e *Engine) ApproveTicket(nonce string) bool {
	e.pendingMu.Lock()
	var ticket *ConsentTicket
	for _, t := range e.pending {
		if t.Nonce == nonce {
			if !t.ExpiresAt.IsZero() && time.Now().After(t.ExpiresAt) {
				e.pendingMu.Unlock()
				return false
			}
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

// DenyTicket denies a pending consent by nonce. Returns false if
// the nonce is unknown OR if the ticket has already expired
// (matching ApproveTicket's expiry check, Phase 17 Fix #2).
func (e *Engine) DenyTicket(nonce string) bool {
	e.pendingMu.Lock()
	var ticket *ConsentTicket
	for _, t := range e.pending {
		if t.Nonce == nonce {
			if !t.ExpiresAt.IsZero() && time.Now().After(t.ExpiresAt) {
				e.pendingMu.Unlock()
				return false
			}
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

// Policy returns the currently-active policy. Added 2026-06-29
// for the E2E test that verifies policy.reload is gated; the test
// compares the active policy pointer before and after a denied
// reload to confirm the gate actually blocked the swap.
func (e *Engine) Policy() *Policy {
	return e.policy.Load()
}

// SetConsentProvider swaps the consent provider (for testing).
func (e *Engine) SetConsentProvider(c ConsentProvider) {
	e.consent = c
}

// SetPresenceChecker wires the user-presence detector (N1). Pass nil
// to disable presence-gating (falls back to the consent modal's
// timeout-queue backstop).
func (e *Engine) SetPresenceChecker(p PresenceChecker) {
	e.PresenceChecker = p
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

// generateNonce returns a 16-byte cryptographically random hex string.
// Replaces the prior UnixNano-based nonce (P3 hygiene fix from the
// production-readiness audit). The nonce is a server-internal lookup
// key, not a security token — the real replay defense is ExpiresAt in
// ApproveTicket/DenyTicket. Using crypto/rand is the right primitive
// for identifiers that could be mistaken for tokens.
func generateNonce() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
