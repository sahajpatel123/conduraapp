// Package watchdog is Layer 2 of the kill switch (CLAUDE.md §5.3).
//
// The watchdog watches the agent for "stuck" states where no human
// has verified the agent's actions for too long. Every user-visible
// action (or explicit GUI ping) calls Watchdog.Touch() to reset the
// inactivity timer. If `timeout` elapses without a Touch, the
// watchdog auto-halt's the daemon via a HaltFlag.
//
// Every halt MUST be forensically recorded per CLAUDE.md §5.4
// ("HMAC-chained, append-only, never deleted"). The watchdog's
// Halt() call goes straight to the HaltFlag — bypassing any RPC
// handler that would normally write the audit row. So we accept
// an optional Auditor and write the audit row **before** the
// Halt(), so even a slow Halt() still leaves a trace.
//
// Why a separate package:
//   - Layer 2 must be implemented in code the agent cannot disable.
//     Putting it in its own package (with no imports of the agent
//     loop, the LLM registry, or the stream manager) makes it harder
//     for a future contributor to accidentally bypass the check.
//   - Layer 2 must survive the agent crashing. The watchdog lives
//     in the daemon's main process and runs on its own goroutine
//     keyed off ctx; the agent's goroutines do not gate it.
//
// Phase 16, Rec 2: ship a simple in-process timer now. A separate
// watcher process that the agent can't kill is v0.2.0 hardening.
package watchdog

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/halt"
)

// DefaultTimeout is the inactivity threshold before the watchdog
// fires. The user can override via config (autonomy.watchdog_timeout).
const DefaultTimeout = 30 * time.Minute

// DefaultCheckInterval is how often the watchdog polls the clock.
// Should be << timeout (typically 1 minute) so the auto-halt is
// within a minute of expiry.
const DefaultCheckInterval = 1 * time.Minute

// HaltFlag is the minimal interface the Watchdog needs. We mirror
// *halt.Flag's signatures exactly so the production wiring is a
// single upcast (no adapter needed). The watchdog package imports
// halt to match the real signature.
type HaltFlag interface {
	Halt(ctx context.Context, reason string) (halt.State, error)
	IsHalted() bool
}

// AuditEvent mirrors audit.Event's subset we need. Declared as
// an interface here so the watchdog package doesn't import
// internal/audit (avoiding a possible future cycle if the audit
// package ever depends on halt). The production wiring passes a
// thin closure that fills the field from the real audit.Event.
type AuditEvent struct {
	Actor  string
	Action string
	Level  string
	Result string
	Detail string
}

// Auditor is the optional audit hook. nil = skip audit (NOT
// recommended — every halt should be forensically traceable).
type Auditor interface {
	RecordHalt(ctx context.Context, e AuditEvent)
}

// Watchdog auto-halt's the daemon after `timeout` of inactivity.
// "Inactivity" is the time since the last Touch().
type Watchdog struct {
	mu        sync.Mutex
	lastTouch time.Time
	timeout   time.Duration
	interval  time.Duration
	halt      HaltFlag
	auditor   Auditor
	log       *slog.Logger

	// onTrip fires when the watchdog fires. nil = use the default
	// (call halt.Halt). Tests set this to a spy.
	onTrip func(reason string)
}

// New creates a Watchdog. The first Touch() is implicitly called
// at construction time so the daemon doesn't immediately halt
// itself on startup.
//
// log may be nil; a nil logger is replaced with slog.Default() so
// callers don't have to nil-check on every path.
// auditor may be nil, but that means halts won't be forensically
// recorded — the production wiring always passes a real auditor.
func New(timeout time.Duration, interval time.Duration, haltFlag HaltFlag, auditor Auditor, log *slog.Logger) *Watchdog {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	if interval <= 0 {
		interval = DefaultCheckInterval
	}
	if log == nil {
		log = slog.Default()
	}
	return &Watchdog{
		lastTouch: time.Now(),
		timeout:   timeout,
		interval:  interval,
		halt:      haltFlag,
		auditor:   auditor,
		log:       log,
	}
}

// Touch records that the user verified the agent's actions
// (e.g. typed into the chat, approved a consent, clicked a UI
// element). Safe for concurrent use.
func (w *Watchdog) Touch() {
	w.mu.Lock()
	w.lastTouch = time.Now()
	w.mu.Unlock()
}

// LastTouch returns the time of the most recent Touch. Used by
// tests and by the GUI's "session idle for X minutes" display.
func (w *Watchdog) LastTouch() time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lastTouch
}

// IdleDuration returns time since the last Touch. Zero if the
// watchdog has never been touched (which shouldn't happen — New
// calls lastTouch = time.Now()).
func (w *Watchdog) IdleDuration() time.Duration {
	w.mu.Lock()
	defer w.mu.Unlock()
	return time.Since(w.lastTouch)
}

// Run blocks until ctx is canceled, ticking at `interval` and
// auto-halting if IdleDuration() exceeds timeout. Designed to be
// called in a goroutine:
//
//	go watchdog.Run(ctx)
//
// The watchdog does NOT call os.Exit or panic on trip; it delegates
// to haltFlag.Halt() so the rest of the daemon can shut down
// cleanly (audit log write, broker close, file flush).
//
// On trip we:
//  1. Write an audit row BEFORE Halt() — even a slow Halt()
//     still leaves a trace. This satisfies CLAUDE.md §5.4's
//     "every halt must be forensically recorded."
//  2. Call Halt() with the same reason.
//  3. Log the trip at WARN level.
func (w *Watchdog) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if w.IdleDuration() < w.timeout {
				continue
			}
			// Already halted — no-op. Otherwise we spam the
			// audit log if the tick fires again before the
			// daemon actually stops.
			if w.halt != nil && w.halt.IsHalted() {
				return
			}
			reason := "watchdog: no user verification for " + w.timeout.String()
			if w.onTrip != nil {
				w.onTrip(reason)
				return
			}
			// Phase 17, Fix #1 (B3): write the audit row BEFORE
			// calling Halt() so a slow halt doesn't lose the
			// trace. The auditor may be nil in tests; production
			// always wires one.
			if w.auditor != nil {
				w.auditor.RecordHalt(ctx, AuditEvent{
					Actor:  "watchdog",
					Action: "daemon.halt",
					Level:  "warn",
					Result: "watchdog_timeout",
					Detail: reason,
				})
			}
			if w.halt != nil {
				if _, err := w.halt.Halt(ctx, reason); err != nil {
					w.log.Warn("watchdog: halt failed", "err", err)
				}
			}
			w.log.Warn("watchdog fired — daemon auto-halted", "idle", w.IdleDuration().String())
			return
		}
	}
}
