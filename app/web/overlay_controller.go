package main

import (
	"context"
	"sync"

	"github.com/sahajpatel123/synapticapp/internal/overlay"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// wailsController drives the Wails window's frameless/always-on-top/
// cursor-positioned overlay mode. It mirrors noopController's state
// machine (so tests for presence/dismissal logic still apply) but
// every state transition also calls the Wails runtime to update the
// actual OS window.
//
// The Wails window is the *only* window in this app. We don't open
// a second window — we re-purpose the existing one. In normal mode
// the window shows the full chat UI at 1200x800. In overlay mode
// the window is frameless, ~600x80, positioned at the cursor, and
// always-on-top. The Svelte component reacts to the state change
// via the existing SSE channel and switches its own layout to the
// compact prompt bar (OverlayPrompt.svelte).
//
// The controller is constructed once in main.go after wails.Run
// sets up the context. It implements overlay.Controller so the
// daemon's conductor and presence orchestrator can drive it
// without knowing about Wails.
type wailsController struct {
	mu        sync.RWMutex
	state     overlay.State
	dismissFn func()

	// ctx is the Wails runtime context. nil until the Wails app
	// has called OnStartup. All Wails runtime methods panic on a
	// nil context, so Show/Hide guard on it.
	ctx context.Context

	// overlayWidth / overlayHeight are the compact overlay-mode
	// dimensions. The normal-mode dimensions are the ones passed
	// to wails.Run's options.App (1200x800).
	overlayWidth  int
	overlayHeight int
}

// newWailsController returns a controller configured to use the
// given Wails runtime context. The ctx is captured by reference —
// call SetContext() from App.startup() to wire it once Wails is
// ready (the context isn't available before wails.Run).
func newWailsController() *wailsController {
	return &wailsController{
		state:         overlay.StateHidden,
		overlayWidth:  620,
		overlayHeight: 88,
	}
}

// SetContext wires the Wails runtime context. Safe to call once
// during App.startup; calling it again is a no-op so the controller
// survives a redundant startup callback.
func (c *wailsController) SetContext(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ctx == nil {
		c.ctx = ctx
	}
}

// Show switches the window into overlay mode and (when AtCursor is
// true) centers it on the current screen.
//
// Wails-side effects:
//   - WindowSetAlwaysOnTop(ctx, true)
//   - WindowSetSize(ctx, overlayWidth, overlayHeight)
//   - WindowCenter(ctx) — when AtCursor is true
//   - WindowSetBackgroundColour(ctx, 18, 18, 22, 230)
//
// Wails has no portable "get cursor position" API. v0.1.0 uses
// WindowCenter() as a Spotlight/Alfred-style compromise — the
// overlay appears at a fixed location on the active screen
// rather than chasing the cursor. v0.2.0 can add a platform-
// specific cursor-pos lookup (CGEventCreate on macOS,
// GetCursorPos on Windows, AT-SPI on Linux) and switch to
// WindowSetPosition(x, y).
//
// When AtCursor is false the window stays at its previous
// position. This is the right default for repeated hotkey
// presses: the user wants the overlay near where it last
// appeared, not jumping around.
func (c *wailsController) Show(ctx context.Context, opts overlay.ShowOpts) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ctx == nil {
		// Wails not ready yet — track state but skip the
		// window ops. The conductor will retry on the next
		// hotkey press.
		c.state = overlay.StateListening
		return nil
	}

	wailsruntime.WindowSetAlwaysOnTop(c.ctx, true)
	wailsruntime.WindowSetSize(c.ctx, c.overlayWidth, c.overlayHeight)
	// 230/255 = ~10% transparency. macOS composites this so the
	// user can see the desktop behind the overlay. On Windows
	// and Linux the effect varies by compositor.
	wailsruntime.WindowSetBackgroundColour(c.ctx, 18, 18, 22, 230)

	if opts.AtCursor {
		// Spotlight/Alfred-style center-on-screen. See the
		// comment above for why we don't track the cursor.
		wailsruntime.WindowCenter(c.ctx)
	}

	c.state = overlay.StateListening
	return nil
}

// Hide switches the window back to normal mode. The window size
// and position are NOT restored — the user might have moved the
// window between overlay invocations and we want to remember that.
// In v0.2.0 we can save/restore on dismiss.
func (c *wailsController) Hide() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != overlay.StateHidden && c.dismissFn != nil {
		c.dismissFn()
	}

	if c.ctx != nil {
		wailsruntime.WindowSetAlwaysOnTop(c.ctx, false)
		// Restore normal-mode dimensions. The chat UI was
		// last seen at 1200x800 — re-asserting those prevents
		// the user from being stuck in the 620x88 size after
		// the first overlay.
		wailsruntime.WindowSetSize(c.ctx, 1200, 800)
		wailsruntime.WindowSetBackgroundColour(c.ctx, 18, 18, 22, 255)
	}

	c.state = overlay.StateHidden
	return nil
}

// Toggle mirrors the noop controller's semantics: hidden → listening,
// anything else → hidden. The daemon's presence orchestrator calls
// this on hotkey press, not Show/Hide directly.
func (c *wailsController) Toggle() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == overlay.StateHidden {
		// Use a zero-value ShowOpts so callers using Toggle()
		// (which doesn't accept opts) keep the window at its
		// last position rather than jumping to the cursor.
		// Cursor-positioning is reserved for explicit Show(ctx, opts).
		if c.ctx != nil {
			wailsruntime.WindowSetAlwaysOnTop(c.ctx, true)
			wailsruntime.WindowSetSize(c.ctx, c.overlayWidth, c.overlayHeight)
			wailsruntime.WindowSetBackgroundColour(c.ctx, 18, 18, 22, 230)
		}
		c.state = overlay.StateListening
	} else {
		if c.dismissFn != nil {
			c.dismissFn()
		}
		if c.ctx != nil {
			wailsruntime.WindowSetAlwaysOnTop(c.ctx, false)
			wailsruntime.WindowSetSize(c.ctx, 1200, 800)
			wailsruntime.WindowSetBackgroundColour(c.ctx, 18, 18, 22, 255)
		}
		c.state = overlay.StateHidden
	}
}

// State returns the current overlay state.
func (c *wailsController) State() overlay.State {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// OnDismiss registers a callback for when the overlay is dismissed.
// The callback fires from inside Hide() or the dismiss branch of
// Toggle() — same semantics as the noop controller.
func (c *wailsController) OnDismiss(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dismissFn = fn
}

// SetState transitions the overlay to a new state. Exposed for
// the presence orchestrator. The Wails controller does NOT push
// the state to the OS in SetState — that's Show/Hide's job — but
// it does update the in-memory state so State() reports truthfully.
func (c *wailsController) SetState(state overlay.State) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = state
}
