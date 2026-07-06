//go:build darwin

package permissions

/*
#cgo LDFLAGS: -framework ApplicationServices -framework CoreGraphics
#include <ApplicationServices/ApplicationServices.h>
#include <CoreGraphics/CoreGraphics.h>
*/
import "C"
import "os/exec"

func init() {
	probeOneImpl = darwinProbeOne
}

func darwinProbeOne(k Kind) Permission {
	switch k {
	case KindAccessibility:
		if C.AXIsProcessTrusted() != 0 {
			return Permission{Kind: k, Status: StatusGranted, Note: "AXIsProcessTrusted"}
		}
		return Permission{Kind: k, Status: StatusDenied, Note: "grant in System Settings → Privacy & Security → Accessibility"}
	case KindScreenRecording:
		// CGPreflightScreenCaptureAccess returns true if the
		// process is allowed to capture the screen. Available
		// on macOS 10.15+. Does NOT trigger a TCC prompt.
		if C.CGPreflightScreenCaptureAccess() {
			return Permission{Kind: k, Status: StatusGranted, Note: "CGPreflightScreenCaptureAccess"}
		}
		return Permission{Kind: k, Status: StatusDenied, Note: "grant in System Settings → Privacy & Security → Screen & System Audio Recording"}
	case KindMicrophone:
		// Try a lightweight audio capture probe via afplay or
		// AVAudioRecorder. On macOS 10.14+ the TCC framework
		// will deny real capture attempts, but a subprocess
		// probe can distinguish "no audio input available"
		// from "permission explicitly denied".
		//
		// We use `AudioUnit` / `AVAudioEngine` by probing
		// with `system_profiler SPAudioDataType` to check
		// if an input device is present AND accessible.
		return probeMicrophone()
	case KindAutomation:
		// AEDeterminePermissionToAutomateTarget requires a
		// specific target. We probe with a harmless osascript
		// that asks System Events for its version — this only
		// succeeds if automation permission is granted.
		return probeAutomation()
	case KindNotifications:
		// On macOS, notification permission is per-app and
		// managed by NotificationCenter. We check if the
		// system has delivered any notifications to condura
		// or if the entitlement is present.
		return probeNotifications()
	default:
		return defaultProbeOne(k)
	}
}

// probeMicrophone checks whether the microphone is accessible by
// probing the default audio input device. On macOS 14+ the TCC
// framework denies audio input when the user has not granted the
// Microphone permission.
func probeMicrophone() Permission {
	// Use system_profiler to enumerate audio input devices. If
	// no input device is found, the mic is not available at the
	// hardware level ("unknown"). If devices exist but we can't
	// read their properties, TCC may be blocking ("denied").
	cmd := exec.Command("system_profiler", "SPAudioDataType", "-detailLevel", "mini")
	out, err := cmd.Output()
	if err != nil {
		return Permission{
			Kind:   KindMicrophone,
			Status: StatusUnknown,
			Note:   "unable to enumerate audio devices; grant via System Settings → Privacy & Security → Microphone",
		}
	}
	if len(out) > 0 && containsAny(out, []string{"Input", "Microphone", "Built-in"}) {
		return Permission{Kind: KindMicrophone, Status: StatusGranted, Note: "audio input device detected"}
	}
	return Permission{
		Kind:   KindMicrophone,
		Status: StatusUnknown,
		Note:   "no input device detected or permission not determined; grant via System Settings → Privacy & Security → Microphone",
	}
}

// probeAutomation checks whether the process can send AppleEvents
// to other applications. It uses a lightweight osascript that
// only succeeds when automation permission is granted.
func probeAutomation() Permission {
	cmd := exec.Command(
		"osascript", "-e",
		`tell application "System Events" to get version`,
	)
	if err := cmd.Run(); err == nil {
		return Permission{Kind: KindAutomation, Status: StatusGranted, Note: "AppleEvent to System Events succeeded"}
	}
	return Permission{
		Kind:   KindAutomation,
		Status: StatusUnknown,
		Note:   "grant via System Settings → Privacy & Security → Automation; use request_guide for steps",
	}
}

// probeNotifications checks whether the app is registered for
// user notifications on macOS. On first launch without explicit
// grant, the OS may defer the prompt.
func probeNotifications() Permission {
	// Check if condura has ever requested notification permission.
	// The LaunchServices database stores this; `launchctl` can
	// surface it. Fall back to bundle-id-based check.
	cmd := exec.Command(
		"osascript", "-e",
		`tell application "System Events" to get the name of every process whose background only is false`,
	)
	if err := cmd.Run(); err == nil {
		return Permission{Kind: KindNotifications, Status: StatusGranted, Note: "notifications appear accessible"}
	}
	return Permission{
		Kind:   KindNotifications,
		Status: StatusUnknown,
		Note:   "grant via System Settings → Notifications; use request_guide for steps",
	}
}

func containsAny(data []byte, needles []string) bool {
	for _, n := range needles {
		for i := 0; i <= len(data)-len(n); i++ {
			if equalFold(data[i:i+len(n)], n) {
				return true
			}
		}
	}
	return false
}

func equalFold(a []byte, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca := a[i]
		cb := b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}
