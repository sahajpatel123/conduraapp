package computeruse

import (
	"context"
	"fmt"
)

// Router selects the cheapest available backend for each action.
// It implements the 4-tier strategy from MISSION §11.2:
//  1. ORAX Eye (free, fast, ~50ms)
//  2. mac-cua (background-first, Apache 2.0)
//  3. macOS-MCP (comprehensive, foreground, MIT)
//  4. Vision CUA (Anthropic/OpenAI, last resort, ~$0.02-0.05/action)
//
// Cascade policy (B-38): by default the router breaks on the first
// backend that returns an *action* error (the action was attempted
// but failed) and does NOT fall through to a costlier backend. This
// is the conservative choice: a click that missed its target should
// surface the real error to the planner, not silently retry on a
// costlier backend that may click a second time. Set
// CascadeOnFailure=true to make the router fall through on any error
// (matching the literal §11.2 "fall back to mac-cua" wording); this
// is appropriate when the caller knows the failure was a
// backend-capability issue rather than an action-semantics issue.
type Router struct {
	backends []Backend
	// CascadeOnFailure, when true, makes Execute fall through to
	// the next backend on any error (including action failures),
	// matching the literal §11.2 cascade wording. Default false
	// (break on first action error; only cascade on "unavailable").
	CascadeOnFailure bool
}

// NewRouter creates a router with the given backends in priority order.
func NewRouter(backends ...Backend) *Router {
	return &Router{
		backends: backends,
	}
}

// Execute calls the function on the first available backend. It
// distinguishes between "backend unavailable" (try next) and "action
// failed" (return immediately, preserving the real error) unless
// CascadeOnFailure is set, in which case any error falls through to
// the next available backend.
func Execute[T any](r *Router, ctx context.Context, fn func(Backend) (T, error)) (T, error) {
	var zero T
	var lastErr error
	for _, b := range r.backends {
		if !b.IsAvailable(ctx) {
			continue
		}
		result, err := fn(b)
		if err != nil {
			lastErr = err
			if r.CascadeOnFailure {
				// Fall through to the next available backend,
				// matching the literal §11.2 cascade wording.
				continue
			}
			// Default: the backend was available but the action
			// failed. Return the real error rather than silently
			// retrying on a costlier backend (which may click a
			// second time). See B-38 for the trade-off.
			break
		}
		return result, nil
	}
	if lastErr != nil {
		return zero, lastErr
	}
	return zero, fmt.Errorf("%w", ErrNoBackend)
}

// ExecuteAction executes a computer-use action using the best available backend.
func (r *Router) ExecuteAction(ctx context.Context, action *Action) (*ActionResult, error) {
	return Execute(r, ctx, func(b Backend) (*ActionResult, error) {
		return b.Execute(ctx, action)
	})
}

// FindBackend returns the first available backend that supports the given capability.
func (r *Router) FindBackend(ctx context.Context, capability Capability) Backend {
	for _, b := range r.backends {
		if !b.IsAvailable(ctx) {
			continue
		}
		for _, c := range b.Capabilities() {
			if c == capability {
				return b
			}
		}
	}
	return nil
}

// Backends returns the list of backends in priority order.
func (r *Router) Backends() []Backend {
	return r.backends
}

// AvailableBackends returns only the backends that are currently available.
func (r *Router) AvailableBackends(ctx context.Context) []Backend {
	var available []Backend
	for _, b := range r.backends {
		if b.IsAvailable(ctx) {
			available = append(available, b)
		}
	}
	return available
}
