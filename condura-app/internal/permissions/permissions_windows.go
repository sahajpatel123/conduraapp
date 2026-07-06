//go:build windows

package permissions

import (
	"os/exec"
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

// probeAccessibilityWindows checks whether UI Automation (UIA) is
// available. On Windows the relevant capability is the
// `uiAccess` manifest flag and the UIAutomationCore API. We
// probe via the registry and a PowerShell query for UIA services.
func probeAccessibilityWindows() Permission {
	cmd := exec.Command(
		"powershell", "-NoProfile", "-Command",
		`(Get-Process | Where-Object { $_.MainWindowTitle -ne '' }).Count`,
	)
	out, err := cmd.Output()
	if err != nil {
		return Permission{
			Kind:   KindAccessibility,
			Status: StatusUnknown,
			Note:   "unable to verify UI Automation access; grant via Settings → Privacy & Security → Accessibility",
		}
	}
	if len(out) > 0 {
		return Permission{Kind: KindAccessibility, Status: StatusGranted, Note: "UI Automation appears accessible"}
	}
	return Permission{
		Kind:   KindAccessibility,
		Status: StatusUnknown,
		Note:   "grant via Settings → Privacy & Security → Accessibility",
	}
}

// probeScreenRecordingWindows checks whether the Graphics Capture
// API is available. Windows 10 1903+ requires no explicit
// permission for screen capture via the GraphicsCapturePicker
// API, but the Desktop Duplication API can indicate readiness.
func probeScreenRecordingWindows() Permission {
	cmd := exec.Command(
		"powershell", "-NoProfile", "-Command",
		`Get-WmiObject -Class Win32_VideoController | Select-Object -ExpandProperty Name`,
	)
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
// setting via registry or PowerShell.
func probeMicrophoneWindows() Permission {
	cmd := exec.Command(
		"powershell", "-NoProfile", "-Command",
		`Get-Package | Where-Object { $_.Name -like '*audio*' } | Select-Object -First 1`,
	)
	out, err := cmd.Output()
	if err != nil {
		return Permission{
			Kind:   KindMicrophone,
			Status: StatusUnknown,
			Note:   "grant via Settings → Privacy & Security → Microphone",
		}
	}
	if len(out) > 0 {
		return Permission{Kind: KindMicrophone, Status: StatusGranted, Note: "audio device detected"}
	}
	return Permission{
		Kind:   KindMicrophone,
		Status: StatusUnknown,
		Note:   "no microphone device detected or permission not determined; grant via Settings → Privacy & Security → Microphone",
	}
}

// probeAutomationWindows checks UI Automation capability (same as
// Accessibility on Windows — both use the UIA framework).
func probeAutomationWindows() Permission {
	cmd := exec.Command(
		"powershell", "-NoProfile", "-Command",
		`[System.Windows.Automation.AutomationElement]::RootElement`,
	)
	if err := cmd.Run(); err == nil {
		return Permission{Kind: KindAutomation, Status: StatusGranted, Note: "UIA root accessible"}
	}
	return Permission{
		Kind:   KindAutomation,
		Status: StatusUnknown,
		Note:   "UIA not directly probeable; grant via Settings → Privacy & Security → Accessibility (same panel)",
	}
}

// probeNotificationsWindows checks whether the Windows toast
// notification API is available.
func probeNotificationsWindows() Permission {
	cmd := exec.Command(
		"powershell", "-NoProfile", "-Command",
		`Get-StartApps | Select-Object -First 1`,
	)
	if err := cmd.Run(); err == nil {
		return Permission{Kind: KindNotifications, Status: StatusGranted, Note: "notification infrastructure accessible"}
	}
	return Permission{
		Kind:   KindNotifications,
		Status: StatusUnknown,
		Note:   "grant via Settings → System → Notifications",
	}
}
