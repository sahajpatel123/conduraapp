// Package health aggregates the health of internal subsystems (database,
// secrets, providers, spend monitor, etc.) and exposes the result via
// the IPC layer and a future /healthz HTTP endpoint.
//
// Each subsystem registers a Check function. Snapshot() runs all checks
// concurrently and returns an aggregate Status.
package health

import (
	"context"
	"sort"
	"sync"
	"time"
)

// State is the health state of a single check or the overall system.
type State string

const (
	StateOK       State = "ok"
	StateDegraded State = "degraded"
	StateDown     State = "down"
)

// Check is a single health check. It must be safe for concurrent use.
type Check struct {
	Name     string
	Timeout  time.Duration // per-check timeout; default 2s
	Check    func(ctx context.Context) error
	Required bool // if true, a failure makes overall State=down
}

// Result is the outcome of one check.
type Result struct {
	Name   string `json:"name"`
	State  State  `json:"state"`
	Error  string `json:"error,omitempty"`
	TookMs int64  `json:"took_ms"`
}

// Snapshot is the aggregate result of all checks.
type Snapshot struct {
	Time    time.Time `json:"time"`
	State   State     `json:"state"`
	Results []Result  `json:"results"`
}

// Register aggregates one or more checks.
type Register struct {
	mu     sync.RWMutex
	checks map[string]Check
}

// New returns a new Register.
func New() *Register {
	return &Register{checks: map[string]Check{}}
}

// Add registers a check. If a check with the same name exists, it is replaced.
func (r *Register) Add(c Check) {
	if c.Timeout <= 0 {
		c.Timeout = 2 * time.Second
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checks[c.Name] = c
}

// Remove removes a check.
func (r *Register) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.checks, name)
}

// Names returns the registered check names, sorted.
func (r *Register) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.checks))
	for n := range r.checks {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Snapshot runs all checks in parallel and returns the aggregate.
func (r *Register) Snapshot(ctx context.Context) Snapshot {
	r.mu.RLock()
	checks := make([]Check, 0, len(r.checks))
	for _, c := range r.checks {
		checks = append(checks, c)
	}
	r.mu.RUnlock()

	results := make([]Result, len(checks))
	var wg sync.WaitGroup
	for i, c := range checks {
		wg.Add(1)
		go func(i int, c Check) {
			defer wg.Done()
			results[i] = runCheck(ctx, c)
		}(i, c)
	}
	wg.Wait()

	overall := StateOK
	for i, c := range checks {
		switch results[i].State {
		case StateDown:
			if c.Required {
				overall = StateDown
				break
			}
			if overall == StateOK {
				overall = StateDegraded
			}
		case StateDegraded:
			if overall == StateOK {
				overall = StateDegraded
			}
		}
		if overall == StateDown {
			break
		}
	}
	return Snapshot{
		Time:    time.Now(),
		State:   overall,
		Results: results,
	}
}

func runCheck(parent context.Context, c Check) Result {
	start := time.Now()
	ctx, cancel := context.WithTimeout(parent, c.Timeout)
	defer cancel()
	res := Result{Name: c.Name}
	if err := c.Check(ctx); err != nil {
		res.State = StateDown
		res.Error = err.Error()
	} else {
		res.State = StateOK
	}
	res.TookMs = time.Since(start).Milliseconds()
	return res
}
