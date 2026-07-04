package delegation

import (
	"context"
	"math"
	"sync"
)

// Limiter enforces recursion depth and per-agent budget bounds.
type Limiter struct {
	cfg          Config
	spendMon     BudgetChecker
	agentBudgets map[string]agentBudget
	mu           sync.Mutex
}

type agentBudget struct {
	spent float64
	cap   float64
}

// BudgetChecker is the subset of SpendMonitor we delegate to.
type BudgetChecker interface {
	Allow(amount float64) bool
}

// NewLimiter creates a spawn limiter. sp may be nil (skip global check).
func NewLimiter(cfg Config, sp BudgetChecker) *Limiter {
	return &Limiter{
		cfg:          cfg,
		spendMon:     sp,
		agentBudgets: make(map[string]agentBudget),
	}
}

// CheckSpawn atomically checks recursion depth and budget, reserving
// the budget on success. Call ReleaseBudget on error to roll back.
func (l *Limiter) CheckSpawn(ctx context.Context, agentName string, depth int, amount float64) error {
	_ = ctx
	agentCfg, ok := l.cfg.FindAgent(agentName)
	if !ok {
		return ErrAgentNotFound
	}

	// Recursion depth check.
	if depth > agentCfg.MaxDepth {
		return ErrRecursionLimit
	}

	// Validate budget amount: negative values corrupt accounting; NaN or Inf
	// poison the ledger permanently.
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount < 0 {
		return ErrBudgetExceeded
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Per-agent budget check.
	ab, exists := l.agentBudgets[agentName]
	if !exists {
		ab = agentBudget{cap: agentCfg.BudgetCap}
	}
	if ab.cap > 0 && ab.spent+amount > ab.cap {
		return ErrBudgetExceeded
	}

	// Global budget check.
	if l.spendMon != nil && !l.spendMon.Allow(amount) {
		return ErrBudgetExceeded
	}

	ab.spent += amount
	l.agentBudgets[agentName] = ab
	return nil
}

// ReleaseBudget rolls back a reserved budget amount.
func (l *Limiter) ReleaseBudget(agentName string, amount float64) {
	if math.IsNaN(amount) || amount <= 0 {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	ab, exists := l.agentBudgets[agentName]
	if exists && ab.spent >= amount {
		ab.spent -= amount
		l.agentBudgets[agentName] = ab
	}
}

// Spend returns the amount spent per agent.
func (l *Limiter) Spend(agentName string) float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.agentBudgets[agentName].spent
}
