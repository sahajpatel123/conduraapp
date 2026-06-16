// Package onboarding — voice readiness probe (Phase 14H).
//
// ProbeVoice reports what the daemon knows about voice readiness so
// the Ready screen (step 4) can show a "Set up voice" card. It is
// intentionally dependency-free and side-effect-free: microphone
// hardware is assumed present on desktop platforms (the actual OS
// permission grant is surfaced separately by the permissions
// subsystem and the Permissions onboarding step), and wake-word
// detection defaults to off until the user enables it in Settings.
package onboarding

import "runtime"

// VoiceProbe is the result of inspecting voice readiness.
type VoiceProbe struct {
	// MicAvailable is true when the platform is expected to expose a
	// microphone. This is a capability hint, not a permission grant.
	MicAvailable bool `json:"mic_available"`
	// VoiceEnabled mirrors whether the voice pipeline is turned on.
	// The static probe reports false; voice is opt-in via Settings.
	VoiceEnabled bool `json:"voice_enabled"`
	// WakeWordEnabled mirrors wake-word detection state (opt-in).
	WakeWordEnabled bool `json:"wake_word_enabled"`
	// WakeWord is the default hotword shown in the UI.
	WakeWord string `json:"wake_word"`
	// Ready is true when voice can plausibly be used on this machine
	// (a microphone is expected to be present).
	Ready bool `json:"ready"`
}

// DefaultWakeWord is shown when no hotword is configured yet.
const DefaultWakeWord = "hey synaptic"

// ProbeVoice returns the current voice readiness snapshot.
func ProbeVoice() *VoiceProbe {
	mic := micExpected()
	return &VoiceProbe{
		MicAvailable:    mic,
		VoiceEnabled:    false,
		WakeWordEnabled: false,
		WakeWord:        DefaultWakeWord,
		Ready:           mic,
	}
}

// micExpected returns true on platforms where a microphone is
// typically available. We never claim a mic on unknown platforms.
func micExpected() bool {
	switch runtime.GOOS {
	case "darwin", "windows", "linux":
		return true
	default:
		return false
	}
}
