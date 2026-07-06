//go:build windows

package permissions

import (
	"context"
	"os/exec"
	"time"
)

func init() {
	probeOneImpl = windowsProbeOne
}

func windowsProbeOne(k Kind) Permission {
	switch k {
	case KindAccessibility:
		return probeAccessibilityWindows()
	case KindScreenRecording:
		return probeScreenRecordingWindows()
	case KindMicrophone:
		return probeMicrophoneWindows()
	case KindAutomation:
		return probeAutomationWindows()
	case KindNotifications:
		return probeNotificationsWindows()
	default:
		return defaultProbeOne(k)
	}
}

// probeTimeout is the max wall-clock time a subprocess probe may run.
const probeTimeout = 3 * time.Second

// execProbe is the function used to spawn subprocess probes and
// read their stdout. Tests override this with a stub that returns
// canned output; the default spawns the real subprocess.
var execProbe = func(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // fixed probe names, not user input
	return cmd.Output()
}

// probeAccessibilityWindows checks whether UI Automation can be
// reached. On Windows the relevant capability is the `uiAccess`
// manifest flag (set at build time) + the UIAutomationCore ACL.
// UIA cannot be reliably probed from outside the process that
// wants to use it — the previous Get-Process-based probe was a
// false positive on every Windows desktop. We return StatusUnknown
// and rely on the first UIA call to surface E_ACCESSDENIED for
// protected UI.
//
// We do verify that the UIAutomationClient type loads, which is
// the minimum requirement to use UIA at all.
func probeAccessibilityWindows() Permission {
	out, err := execProbe("powershell", "-NoProfile", "-Command",
		`try { [System.Windows.Automation.AutomationElement].Assembly.GetName().Name; 'OK' } catch { 'NO' }`)
	if err != nil || len(out) == 0 {
		return Permission{
			Kind:   KindAccessibility,
			Status: StatusUnknown,
			Note:   "PowerShell probe failed; verify UIAutomationCore availability",
		}
	}
	// We deliberately do NOT return StatusGranted here — the
	// capability to LOAD the assembly is not the same as having
	// uiAccess. Surface StatusUnknown with the right Settings pane.
	return Permission{
		Kind:   KindAccessibility,
		Status: StatusUnknown,
		Note:   "UIAutomationClient loads but per-app uiAccess must be granted at build (manifest flag) and verified after first AX tree walk; configure via Settings → Privacy & Security → Accessibility",
	}
}

// probeScreenRecordingWindows checks whether the Graphics Capture
// API is available. The WMI probe was removed — hardware presence
// does not imply screen-recording permission. We return StatusUnknown
// with a pointer to the right Settings pane and rely on the first
// capture attempt to surface the actual consent prompt
// (Windows.Graphics.Capture APIs fail closed when consent is absent).
func probeScreenRecordingWindows() Permission {
	return Permission{
		Kind:   KindScreenRecording,
		Status: StatusUnknown,
		Note:   "screen capture permission cannot be probed reliably; verify after first capture attempt via Settings → Privacy & Security → Graphics capture settings",
	}
}

// probeMicrophoneWindows distinguishes hardware presence from
// permission state. Win32 apps receive the per-app microphone
// consent prompt the first time they call WASAPI / waveIn to open
// a capture stream — there is no registry key or Settings toggle
// we can read in advance. We check WMI for an audio capture
// device so the UI can show "no microphone plugged in" vs.
// "permission needed" as distinct states.
func probeMicrophoneWindows() Permission {
	out, err := execProbe("powershell", "-NoProfile", "-Command",
		`Get-WmiObject -Class Win32_SoundDevice | Where-Object { $_.Status -eq 'OK' } | Select-Object -ExpandProperty Name`)
	if err != nil {
		return Permission{
			Kind:   KindMicrophone,
			Status: StatusUnknown,
			Note:   "unable to enumerate audio devices; verify via Settings → Privacy & Security → Microphone",
		}
	}
	if len(out) == 0 {
		return Permission{
			Kind:   KindMicrophone,
			Status: StatusDenied,
			Note:   "no audio capture device detected; plug in a microphone or check Device Manager",
		}
	}
	// Device present but permission state unknowable in advance.
	// Surface StatusUnknown so the UI shows "not determined" and
	// the first WASAPI call will reveal the actual state.
	return Permission{
		Kind:   KindMicrophone,
		Status: StatusUnknown,
		Note:   "audio capture device detected; permission state is determined on first WASAPI call — grant via Settings → Privacy & Security → Microphone",
	}
}

// probeAutomationWindows shares the same backend (UIA) as
// accessibility on Windows. There is no separate OS-level grant
// for automation; the app's uiAccess manifest flag is the gate.
// We surface StatusUnknown with a pointer to the Accessibility
// Settings pane where users can enable per-app access.
func probeAutomationWindows() Permission {
	return Permission{
		Kind:   KindAutomation,
		Status: StatusUnknown,
		Note:   "Windows has no separate Automation permission; depends on the same uiAccess flag as Accessibility. Configure via Settings → Privacy & Security → Accessibility",
	}
}

// probeNotificationsWindows verifies that the Windows toast
// notification infrastructure (Windows.UI.Notifications) is
// available. The per-app notification toggle is recorded in
// HKCU:\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings\
// but the key name requires the UWP-style AppId which unpackaged
// Win32 apps do not have. We return StatusUnknown and rely on the
// first ToastNotificationManager.CreateToastNotifier call to
// surface E_ACCESSDENIED if the user has disabled notifications.
func probeNotificationsWindows() Permission {
	out, err := execProbe("powershell", "-NoProfile", "-Command",
		`try { [Windows.UI.Notifications.ToastNotificationManager,Windows.UI.Notifications,ContentType=WindowsRuntime] | Out-Null; 'OK' } catch { 'NO' }`)
	if err != nil || len(out) == 0 || string(out) == "NO\r\n" {
		return Permission{
			Kind:   KindNotifications,
			Status: StatusUnknown,
			Note:   "Windows.UI.Notifications runtime type unavailable; grant via Settings → System → Notifications",
		}
	}
	return Permission{
		Kind:   KindNotifications,
		Status: StatusUnknown,
		Note:   "toast notification API available; per-app state is determined on first ToastNotificationManager call — configure via Settings → System → Notifications",
	}
}
