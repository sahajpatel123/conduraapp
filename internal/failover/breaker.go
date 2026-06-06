// Package failover implements resilience layers for LLM calls:
//
//   - CircuitBreaker per provider (closed → open → half-open).
//   - Failover: try a chain of (provider, model) candidates in order.
//   - SpendMonitor: per-day spend tracking + cap.
//
// All three are safe for concurrent use.
package failover

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// CircuitState is the breaker state.
type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"    // calls flow normally
	CircuitOpen     CircuitState = "open"      // calls fail fast
	CircuitHalfOpen CircuitState = "half_open" // one trial call allowed
)

// CircuitBreaker tracks recent failures per provider and short-circuits
// calls when the failure rate is too high.
//
// State machine:
//   - closed:  N consecutive failures → open (for cool-down duration)
//   - open:    after cool-down, the next call is allowed (half-open)
//   - half_open: one trial call. success → closed, failure → open
type CircuitBreaker struct {
	mu sync.Mutex

	failureThreshold int           // consecutive failures to open
	coolDown         time.Duration // how long to stay open

	state         CircuitState
	failures      int
	openedAt      time.Time
	halfOpenInUse bool
}

// NewCircuitBreaker returns a breaker with sensible defaults: 3 consecutive
// failures, 30s cool-down.
func NewCircuitBreaker(failureThreshold int, coolDown time.Duration) *CircuitBreaker {
	if failureThreshold <= 0 {
		failureThreshold = 3
	}
	if coolDown <= 0 {
		coolDown = 30 * time.Second
	}
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		coolDown:         coolDown,
		state:            CircuitClosed,
	}
}

// State returns the current state, transitioning open → half-open if
// the cool-down has elapsed.
func (b *CircuitBreaker) State() CircuitState {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.maybeTransitionLocked()
	return b.state
}

// Allow reports whether a call may proceed. In half-open state, only one
// call is allowed at a time; subsequent calls fail fast until the trial
// completes.
func (b *CircuitBreaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.maybeTransitionLocked()
	switch b.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		return false
	case CircuitHalfOpen:
		if b.halfOpenInUse {
			return false
		}
		b.halfOpenInUse = true
		return true
	}
	return false
}

// RecordSuccess resets the breaker to closed.
func (b *CircuitBreaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.state = CircuitClosed
	b.failures = 0
	b.halfOpenInUse = false
}

// RecordFailure increments the failure count; opens the circuit if the
// threshold is reached.
func (b *CircuitBreaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	b.halfOpenInUse = false
	if b.state == CircuitHalfOpen {
		b.openLocked()
		return
	}
	if b.failures >= b.failureThreshold {
		b.openLocked()
	}
}

// Reset returns the breaker to the initial closed state.
func (b *CircuitBreaker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.state = CircuitClosed
	b.failures = 0
	b.halfOpenInUse = false
}

// maybeTransitionLocked moves open → half-open if the cool-down elapsed.
// Caller must hold b.mu.
func (b *CircuitBreaker) maybeTransitionLocked() {
	if b.state == CircuitOpen && time.Since(b.openedAt) >= b.coolDown {
		b.state = CircuitHalfOpen
		b.halfOpenInUse = false
	}
}

func (b *CircuitBreaker) openLocked() {
	b.state = CircuitOpen
	b.openedAt = time.Now()
	b.halfOpenInUse = false
}

// Common errors.
var (
	ErrCircuitOpen  = errors.New("failover: circuit open")
	ErrAllExhausted = errors.New("failover: all candidates exhausted")
	ErrSpendCap     = errors.New("failover: daily spend cap reached")
)

// -----------------------------------------------------------------------------
// Failover chain
// -----------------------------------------------------------------------------

// Candidate is a (provider, model) pair to try in order.
type Candidate struct {
	Provider string
	Model    string
}

// ChatFn is the per-candidate chat function. Implementations typically
// look up the provider in a registry and call Chat.
type ChatFn func(ctx context.Context, provider, model string) error

// Result is the outcome of a failover call.
type Result struct {
	Provider  string
	Model     string
	Attempts  int
	LastError error
	TotalTime time.Duration
}

// Run tries each candidate in order, calling chat for each. Stops on the
// first success. Returns ErrAllExhausted if every candidate failed.
func Run(ctx context.Context, candidates []Candidate, chat ChatFn) (Result, error) {
	start := time.Now()
	res := Result{Attempts: 0}
	for _, c := range candidates {
		res.Attempts++
		res.Provider = c.Provider
		res.Model = c.Model
		err := chat(ctx, c.Provider, c.Model)
		if err == nil {
			res.TotalTime = time.Since(start)
			return res, nil
		}
		res.LastError = err
		// Honor context cancellation.
		if ctx.Err() != nil {
			res.TotalTime = time.Since(start)
			return res, ctx.Err()
		}
	}
	res.TotalTime = time.Since(start)
	return res, fmt.Errorf("%w: %d candidates", ErrAllExhausted, len(candidates))
}

// -----------------------------------------------------------------------------
// BreakerRegistry — one breaker per provider
// -----------------------------------------------------------------------------

// BreakerRegistry manages a CircuitBreaker per provider name.
type BreakerRegistry struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker

	failureThreshold int
	coolDown         time.Duration
}

// NewBreakerRegistry returns a registry with the given settings for all
// breakers it creates.
func NewBreakerRegistry(failureThreshold int, coolDown time.Duration) *BreakerRegistry {
	return &BreakerRegistry{
		breakers:         map[string]*CircuitBreaker{},
		failureThreshold: failureThreshold,
		coolDown:         coolDown,
	}
}

// For returns (creating if needed) the breaker for the given provider.
func (r *BreakerRegistry) For(provider string) *CircuitBreaker {
	r.mu.RLock()
	b, ok := r.breakers[provider]
	r.mu.RUnlock()
	if ok {
		return b
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.breakers[provider]; ok {
		return b
	}
	b = NewCircuitBreaker(r.failureThreshold, r.coolDown)
	r.breakers[provider] = b
	return b
}

// ResetAll returns every breaker to closed.
func (r *BreakerRegistry) ResetAll() {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.breakers {
		b.Reset()
	}
}

// States returns a snapshot of every breaker's state.
func (r *BreakerRegistry) States() map[string]CircuitState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]CircuitState, len(r.breakers))
	for n, b := range r.breakers {
		out[n] = b.State()
	}
	return out
}

// -----------------------------------------------------------------------------
// SpendMonitor
// -----------------------------------------------------------------------------

// SpendCap is the daily USD limit.
type SpendCap struct {
	USDPerDay float64
}

// SpendMonitor tracks per-day spend in memory. Persistence (to llm_calls
// and spend_daily) is the caller's responsibility; this monitor is the
// in-process gate that decides whether a call may proceed.
type SpendMonitor struct {
	mu    sync.Mutex
	cap   SpendCap
	day   string // YYYY-MM-DD in local time
	spent float64
	nowFn func() time.Time
}

// NewSpendMonitor returns a monitor with the given cap.
func NewSpendMonitor(cap SpendCap) *SpendMonitor {
	return &SpendMonitor{cap: cap, nowFn: time.Now}
}

// SetCap updates the cap at runtime.
func (m *SpendMonitor) SetCap(cap SpendCap) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cap = cap
}

// Cap returns the current cap.
func (m *SpendMonitor) Cap() SpendCap {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cap
}

// Spent returns the spend recorded for the current day.
func (m *SpendMonitor) Spent() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rollIfNewDayLocked()
	return m.spent
}

// Remaining returns the cap minus today's spend (never negative).
func (m *SpendMonitor) Remaining() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rollIfNewDayLocked()
	r := m.cap.USDPerDay - m.spent
	if r < 0 {
		return 0
	}
	return r
}

// Record adds amount to today's spend.
func (m *SpendMonitor) Record(amount float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rollIfNewDayLocked()
	m.spent += amount
}

// Allow reports whether amount can be spent without exceeding the cap.
func (m *SpendMonitor) Allow(amount float64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rollIfNewDayLocked()
	return m.spent+amount <= m.cap.USDPerDay
}

func (m *SpendMonitor) rollIfNewDayLocked() {
	today := m.nowFn().Format("2006-01-02")
	if today != m.day {
		m.day = today
		m.spent = 0
	}
}
