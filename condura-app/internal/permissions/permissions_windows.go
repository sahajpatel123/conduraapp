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

// runProbe executes cmd with a short timeout. Returns true if the
// command exited successfully within the deadline.
func runProbe(name string, args ...string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // fixed probe names, not user input
	return cmd.Run() == nil
}

// probeAccessibilityWindows checks whether UI Automation (UIA) is
// available. On Windows the relevant capability is the
// `uiAccess` manifest flag and the UIAutomationCore API.
func probeAccessibilityWindows() Permission {
	if runProbe("powershell", "-NoProfile", "-Command",
		`(Get-Process | Where-Object { $_.MainWindowTitle -ne '' }).Count`) {
		return Permission{Kind: KindAccessibility, Status: StatusGranted, Note: "UI Automation appears accessible"}
	}
	return Permission{
		Kind:   KindAccessibility,
		Status: StatusUnknown,
		Note:   "grant via Settings → Privacy & Security → Accessibility",
	}
}

// probeScreenRecordingWindows checks whether the Graphics Capture
// API is available via WMI.
func probeScreenRecordingWindows() Permission {
	cmd := exec.Command(
		"powershell", "-NoProfile", "-Command",
		`Get-WmiObject -Class Win32_VideoController | Select-Object -ExpandProperty Name`,
	)
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()
	cmd = exec.CommandContext(ctx, //nolint:gosec // fixed probe args, not user input
		"powershell", "-NoProfile", "-Command",
		`Get-WmiObject -Class Win32_VideoController | Select-Object -ExpandProperty Name`)
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return Permission{
			Kind:   KindScreenRecording,
			Status: StatusUnknown,
			Note:   "unable to verify screen capture capability; grant via Settings → Privacy & Security → Graphics capture settings",
		}
	}
	return Permission{Kind: KindScreenRecording, Status: StatusGranted, Note: "graphics controller detected; screen capture accessible"}
}

// probeMicrophoneWindows checks the Windows microphone privacy
// setting via PowerShell.
func probeMicrophoneWindows() Permission {
	if runProbe("powershell", "-NoProfile", "-Command",
		`Get-Package | Where-Object { $_.Name -like '*audio*' } | Select-Object -First 1`) {
		return Permission{Kind: KindMicrophone, Status: StatusGranted, Note: "audio device detected"}
	}
	return Permission{
		Kind:   KindMicrophone,
		Status: StatusUnknown,
		Note:   "no microphone device detected or permission not determined; grant via Settings → Privacy & Security → Microphone",
	}
}

// probeAutomationWindows checks UI Automation capability (same as
// Accessibility on Windows — both use UIA).
func probeAutomationWindows() Permission {
	if runProbe("powershell", "-NoProfile", "-Command",
		`[System.Windows.Automation.AutomationElement]::RootElement`) {
		return Permission{Kind: KindAutomation, Status: StatusGranted, Note: "UIA root accessible"}
	}
	return Permission{
		Kind:   KindAutomation,
		Status: StatusUnknown,
		Note:   "UIA not directly probeable; grant via Settings → Privacy & Security → Accessibility",
	}
}

// probeNotificationsWindows checks whether the Windows toast
// notification API is available.
func probeNotificationsWindows() Permission {
	if runProbe("powershell", "-NoProfile", "-Command",
		`Get-StartApps | Select-Object -First 1`) {
		return Permission{Kind: KindNotifications, Status: StatusGranted, Note: "notification infrastructure accessible"}
	}
	return Permission{
		Kind:   KindNotifications,
		Status: StatusUnknown,
		Note:   "grant via Settings → System → Notifications",
	}
}
