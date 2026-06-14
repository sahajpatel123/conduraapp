// Package permissions probes the OS-level permissions Synaptic
// needs to operate (Phase 11, sub-phase 11E).
//
// Honest constraint: we CANNOT grant OS permissions programmatically
// on most platforms. The TCC/UI Automation/AT-SPI stacks require
// user interaction (System Settings → Privacy & Security). All we
// can do is:
//  1. Probe whether each permission is currently granted.
//  2. Return per-platform guide steps for the user to grant them.
//
// macOS can technically be probed via the real APIs (AXIsProcessTrusted
// etc); we wrap those. Windows and Linux have a mix of runtime
// detection (UI Automation, AT-SPI bridge) and heuristics
// (config presence, service running).
//
// The hard rule: if a permission probe returns "denied", the
// caller is responsible for telling the user what to do. We
// never silently mark something as "granted".
package permissions

import (
	"context"
	"runtime"
)

// Kind is a Synaptic permission requirement.
type Kind string

// Permission kinds Synaptic needs on every supported platform.
const (
	KindAccessibility   Kind = "accessibility"    // macOS AX / Windows UIA / Linux AT-SPI
	KindScreenRecording Kind = "screen_recording" // macOS ScreenCapture / Windows GraphicsCapture
	KindMicrophone      Kind = "microphone"       // all platforms
	KindAutomation      Kind = "automation"       // macOS AppleEvents (other apps)
	KindNotifications   Kind = "notifications"    // all platforms
)

// Status is the current grant state.
type Status string

const (
	// StatusGranted means the OS confirms the permission is
	// currently held by this process.
	StatusGranted Status = "granted"
	// StatusDenied means the OS confirms the permission is NOT
	// held. The user must grant it manually.
	StatusDenied Status = "denied"
	// StatusUnknown means the platform or build cannot determine
	// the state (no native probe, CI environment, etc).
	StatusUnknown Status = "unknown"
)

// Permission is the result of probing a single Kind.
type Permission struct {
	Kind   Kind   `json:"kind"`
	Status Status `json:"status"`
	Note   string `json:"note,omitempty"`
}

// Probe checks every permission Synaptic needs. The ctx is
// honored by per-platform probes that may make a syscall or
// shell out (not currently used, but reserved for future
// platform-specific implementations).
func Probe(ctx context.Context) ([]Permission, error) {
	_ = ctx
	return probeAll(), nil
}

// probeAll is the platform-agnostic dispatcher. Platform-specific
// overrides are in permissions_darwin.go etc.
func probeAll() []Permission {
	all := []Kind{
		KindAccessibility,
		KindScreenRecording,
		KindMicrophone,
		KindAutomation,
		KindNotifications,
	}
	out := make([]Permission, 0, len(all))
	for _, k := range all {
		out = append(out, probeOne(k))
	}
	return out
}

// probeOne returns the Status for a single Kind on the current
// platform. Platform files override the function for kinds
// they can probe natively.
func probeOne(k Kind) Permission {
	// Default: unknown. Platform files override.
	return Permission{Kind: k, Status: StatusUnknown, Note: "platform " + runtime.GOOS + " has no native probe for " + string(k)}
}

// Check returns the status of a single kind. Convenience for
// the GUI ("is accessibility granted right now?").
func Check(k Kind) Status {
	return probeOne(k).Status
}

// Manager is a thin sentinel that lets the daemon hold the
// permissions package as a Subsystem. The actual work is done
// by the package-level Probe / Check / RequestGuide functions.
type Manager struct{}

// NewManager returns a Manager. Always succeeds; the package
// has no construction state.
func NewManager() *Manager { return &Manager{} }

// Platform returns the current OS identifier (e.g. "darwin",
// "windows", "linux"). Wraps runtime.GOOS for use in RPC
// responses and logs.
func Platform() string { return runtime.GOOS }

// Guide is a per-platform, per-kind set of steps the user
// follows to grant the permission. macOS gets specific
// System Settings paths; Windows gets Settings + capabilities;
// Linux gets the portal URL or the relevant package name.
//
// Honest constraint: we CANNOT link directly to the right
// System Settings pane on most OS versions. We provide the
// canonical path text and the user clicks.
type Guide struct {
	Kind     Kind     `json:"kind"`
	Platform string   `json:"platform"`
	Title    string   `json:"title"`
	Steps    []string `json:"steps"`
	DeepLink string   `json:"deep_link,omitempty"` // best-effort, may be empty
	HelpURL  string   `json:"help_url,omitempty"`
}

// RequestGuide returns the per-platform guide for granting k.
// The GUI surfaces this when Status is Denied or Unknown.
func RequestGuide(k Kind) Guide {
	return buildGuide(k, runtime.GOOS)
}

// buildGuide is the platform-agnostic dispatcher.
func buildGuide(k Kind, platform string) Guide {
	g := Guide{Kind: k, Platform: platform, Title: humanTitle(k)}
	g.Steps, g.DeepLink, g.HelpURL = stepsFor(k, platform)
	return g
}

func humanTitle(k Kind) string {
	switch k {
	case KindAccessibility:
		return "Grant Accessibility access"
	case KindScreenRecording:
		return "Grant Screen Recording access"
	case KindMicrophone:
		return "Grant Microphone access"
	case KindAutomation:
		return "Grant Automation access (AppleEvents)"
	case KindNotifications:
		return "Grant Notifications access"
	}
	return string(k)
}

func stepsFor(k Kind, platform string) (steps []string, deep, help string) {
	switch platform {
	case "darwin":
		return darwinSteps(k)
	case "windows":
		return windowsSteps(k)
	case "linux":
		return linuxSteps(k)
	}
	return []string{"No per-platform guide available for " + platform}, "", ""
}

func darwinSteps(k Kind) ([]string, string, string) {
	switch k {
	case KindAccessibility:
		return []string{
			"Open System Settings → Privacy & Security → Accessibility",
			"Click the lock icon and authenticate",
			"Find Synaptic in the list and toggle it ON",
			"If Synaptic is not in the list, click + and add it from /Applications",
		}, "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility", "https://support.apple.com/guide/mac-help/mh43185/mac"
	case KindScreenRecording:
		return []string{
			"Open System Settings → Privacy & Security → Screen & System Audio Recording",
			"Click the lock icon and authenticate",
			"Find Synaptic in the list and toggle it ON",
		}, "x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture", "https://support.apple.com/guide/mac-help/mh43185/mac"
	case KindMicrophone:
		return []string{
			"Open System Settings → Privacy & Security → Microphone",
			"Click the lock icon and authenticate",
			"Find Synaptic in the list and toggle it ON",
		}, "x-apple.systempreferences:com.apple.preference.security?Privacy_Microphone", "https://support.apple.com/guide/mac-help/mh43185/mac"
	case KindAutomation:
		return []string{
			"Open System Settings → Privacy & Security → Automation",
			"Find Synaptic in the list and toggle it ON for each app Synaptic should control",
		}, "x-apple.systempreferences:com.apple.preference.security?Privacy_Automation", ""
	case KindNotifications:
		return []string{
			"Open System Settings → Notifications",
			"Find Synaptic in the list and set Allow Notifications to ON",
		}, "x-apple.systempreferences:com.apple.preference.notifications", ""
	}
	return nil, "", ""
}

func windowsSteps(k Kind) ([]string, string, string) {
	switch k {
	case KindAccessibility:
		return []string{
			"Open Settings → Privacy & Security → Accessibility",
			"Click Synaptic and toggle ON",
		}, "ms-settings:privacy-accessibility", ""
	case KindScreenRecording:
		return []string{
			"Open Settings → Privacy & Security → Graphics capture settings",
			"Click Synaptic and toggle ON",
		}, "ms-settings:privacy-graphicsCapture", ""
	case KindMicrophone:
		return []string{
			"Open Settings → Privacy & Security → Microphone",
			"Click Synaptic and toggle ON",
		}, "ms-settings:privacy-microphone", ""
	case KindAutomation:
		return []string{
			"No OS-level Automation permission on Windows; Synaptic uses UI Automation (UIA) which is enabled per-app via the same Privacy panel as Accessibility.",
		}, "ms-settings:privacy-accessibility", ""
	case KindNotifications:
		return []string{
			"Open Settings → System → Notifications",
			"Find Synaptic and configure your preferences",
		}, "ms-settings:notifications", ""
	}
	return nil, "", ""
}

func linuxSteps(k Kind) ([]string, string, string) {
	switch k {
	case KindAccessibility:
		return []string{
			"Install at-spi2-core: sudo apt install at-spi2-core (Debian/Ubuntu) or sudo dnf install at-spi2-core (Fedora)",
			"GNOME: Settings → Accessibility → Always Show Universal Access Menu",
			"KDE: System Settings → Accessibility",
		}, "", "https://www.freedesktop.org/wiki/Accessibility/AT-SPI2/"
	case KindScreenRecording:
		return []string{
			"Wayland: install xdg-desktop-portal and a backend (e.g. xdg-desktop-portal-gnome)",
			"X11: no permission required (any X client can capture the screen)",
		}, "", ""
	case KindMicrophone:
		return []string{
			"Add your user to the 'audio' group: sudo usermod -aG audio $USER",
			"GNOME: Settings → Privacy → Microphone",
			"PipeWire / PulseAudio: ensure the Synaptic process can access the mic source",
		}, "", ""
	case KindAutomation:
		return []string{
			"AT-SPI2 provides accessibility scripting; no separate OS-level grant needed",
		}, "", ""
	case KindNotifications:
		return []string{
			"GNOME: Settings → Notifications",
			"Install libnotify (usually preinstalled)",
		}, "", ""
	}
	return nil, "", ""
}
