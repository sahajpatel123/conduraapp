// Package overlay manages the overlay window lifecycle.
//
// The overlay is a second window (frameless, transparent, always-on-top)
// that appears near the cursor when the user activates voice input or
// requests a compact chat interface.
//
// The Controller interface allows swapping implementations: the real
// Wails multi-window impl or a headless noop_controller for fallback.
package overlay

import "context"

// Controller manages the overlay window lifecycle.
type Controller interface {
	// Show displays the overlay window. opts controls positioning.
	Show(ctx context.Context, opts ShowOpts) error
	// Hide dismisses the overlay window.
	Hide() error
	// Toggle flips between shown and hidden states.
	Toggle()
	// State returns the current overlay state.
	State() State
	// OnDismiss registers a callback for when the overlay is dismissed.
	OnDismiss(func())
}

// State represents the visual state of the overlay.
type State int

const (
	// StateHidden means the overlay is not visible.
	StateHidden State = iota
	// StateListening means the overlay is capturing audio.
	StateListening
	// StateThinking means the overlay is processing a transcription.
	StateThinking
	// StateSpeaking means the overlay is playing back TTS audio.
	StateSpeaking
)

// String returns a human-readable label for the state.
func (s State) String() string {
	switch s {
	case StateHidden:
		return "hidden"
	case StateListening:
		return "listening"
	case StateThinking:
		return "thinking"
	case StateSpeaking:
		return "speaking"
	default:
		return "unknown"
	}
}

// ShowOpts configures how the overlay is displayed.
type ShowOpts struct {
	// AtCursor positions the overlay near the mouse cursor.
	AtCursor bool
	// X, Y override the cursor position (ignored if !AtCursor).
	X, Y int
}
