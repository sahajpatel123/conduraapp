// Package status defines the unified agent status enum used by the
// tray, the overlay, and the voice pipeline.
//
// MISSION §19.4 calls out four states the agent can be in: idle,
// listening, thinking, speaking. Phase 6 adds two more for safety
// and resilience: halted (kill switch is active) and error
// (something went wrong and the agent is in a degraded mode).
//
// Every status visible to the user (tray icon, overlay, voice orb)
// flows through this enum. The tray's SetStatus method derives its
// label and icon from the value, ensuring a single source of truth.
package status

// Status is the agent's current operational state.
type Status int

const (
	// StatusIdle means the agent is on and listening for the
	// hotkey, but no session is active.
	StatusIdle Status = iota
	// StatusListening means the user is recording a voice query.
	StatusListening
	// StatusThinking means the agent is processing a request.
	StatusThinking
	// StatusSpeaking means the agent is playing back TTS audio.
	StatusSpeaking
	// StatusHalted means the kill switch is active and all
	// actions are blocked.
	StatusHalted
	// StatusError means something went wrong and the agent is
	// in a degraded mode. The user should check the tray
	// tooltip for details.
	StatusError
)

// String returns a lowercase, human-readable label for the status.
// The label is safe for use in UI strings and for the audit log.
func (s Status) String() string {
	switch s {
	case StatusIdle:
		return "idle"
	case StatusListening:
		return "listening"
	case StatusThinking:
		return "thinking"
	case StatusSpeaking:
		return "speaking"
	case StatusHalted:
		return "halted"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// Label returns a title-cased, human-readable label suitable for
// tray menu items and overlay headers.
func (s Status) Label() string {
	switch s {
	case StatusIdle:
		return "Idle"
	case StatusListening:
		return "Listening..."
	case StatusThinking:
		return "Thinking..."
	case StatusSpeaking:
		return "Speaking..."
	case StatusHalted:
		return "Halted"
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// IsActive reports whether the status represents an in-progress
// session (listening, thinking, speaking). Used by the tray to
// determine whether the icon should pulse.
func (s Status) IsActive() bool {
	switch s {
	case StatusListening, StatusThinking, StatusSpeaking:
		return true
	default:
		return false
	}
}
