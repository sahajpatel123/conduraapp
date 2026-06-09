package computeruse

import (
	"context"
	"fmt"
)

// Router selects the cheapest available backend for each action.
// It implements the 4-tier strategy from MISSION §11.2:
// 1. ORAX Eye (free, fast, ~50ms)
// 2. mac-cua (background-first, Apache 2.0)
// 3. macOS-MCP (comprehensive, foreground, MIT)
// 4. Vision CUA (Anthropic/OpenAI, last resort, ~$0.02-0.05/action)
type Router struct {
	backends []Backend
}

// NewRouter creates a router with the given backends in priority order.
func NewRouter(backends ...Backend) *Router {
	return &Router{
		backends: backends,
	}
}

// Execute calls the function on the first available backend. It
// distinguishes between "backend unavailable" (try next) and "action
// failed" (return immediately, preserving the real error).
func Execute[T any](r *Router, ctx context.Context, fn func(Backend) (T, error)) (T, error) {
	var zero T
	var lastErr error
	for _, b := range r.backends {
		if !b.IsAvailable(ctx) {
			continue
		}
		result, err := fn(b)
		if err != nil {
			// If the backend is available but the action failed,
			// return the real error rather than falling through
			// to a costlier backend.
			lastErr = err
			break
		}
		return result, nil
	}
	if lastErr != nil {
		return zero, lastErr
	}
	return zero, fmt.Errorf("no available backend for operation")
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
