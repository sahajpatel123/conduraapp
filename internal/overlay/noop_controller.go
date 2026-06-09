package overlay

import (
	"context"
	"sync"
)

// noopController is a headless overlay controller with a full state machine.
// It tracks state transitions, fires OnDismiss, and is unit-tested against
// the same spec as the real implementation. This is NOT a literal no-op —
// it provides genuine logic coverage for 4.3/4.4 even without a visible window.
type noopController struct {
	mu        sync.RWMutex
	state     State
	dismissFn func()
}

// NewNoopController creates a headless overlay controller.
func NewNoopController() Controller {
	return &noopController{
		state: StateHidden,
	}
}

func (c *noopController) Show(_ context.Context, _ ShowOpts) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.state = StateListening
	return nil
}

func (c *noopController) Hide() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != StateHidden && c.dismissFn != nil {
		c.dismissFn()
	}

	c.state = StateHidden
	return nil
}

func (c *noopController) Toggle() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == StateHidden {
		c.state = StateListening
	} else {
		if c.dismissFn != nil {
			c.dismissFn()
		}
		c.state = StateHidden
	}
}

func (c *noopController) State() State {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.state
}

func (c *noopController) OnDismiss(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.dismissFn = fn
}

// SetState transitions the overlay to a new state. This is exposed for
// use by the presence orchestrator (4.3) which drives state changes.
func (c *noopController) SetState(state State) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.state = state
}
