//go:build linux

package permissions

import (
	"context"
	"os"
	"os/exec"
	"time"
)

func init() {
	probeOneImpl = linuxProbeOne
}

func linuxProbeOne(k Kind) Permission {
	switch k {
	case KindAccessibility:
		return probeAccessibilityLinux()
	case KindScreenRecording:
		return probeScreenRecordingLinux()
	case KindMicrophone:
		return probeMicrophoneLinux()
	case KindAutomation:
		return probeAutomationLinux()
	case KindNotifications:
		return probeNotificationsLinux()
	default:
		return defaultProbeOne(k)
	}
}

// probeAccessibilityLinux checks for the AT-SPI2 accessibility bus,
// which most modern Linux desktops use (GNOME, KDE via plugin).
// The at-spi2-registryd process is a reliable indicator.
func probeAccessibilityLinux() Permission {
	if processRunning("at-spi2-registryd") || processRunning("at-spi-bus-launcher") {
		return Permission{Kind: KindAccessibility, Status: StatusGranted, Note: "AT-SPI2 accessibility bus detected"}
	}
	if dbusServiceAccessible("org.a11y.Bus") {
		return Permission{Kind: KindAccessibility, Status: StatusGranted, Note: "AT-SPI2 D-Bus service accessible"}
	}
	return Permission{
		Kind:   KindAccessibility,
		Status: StatusUnknown,
		Note:   "install AT-SPI2: sudo apt install at-spi2-core (Debian/Ubuntu) or sudo dnf install at-spi2-core (Fedora)",
	}
}

// probeScreenRecordingLinux checks for the screen-capture
// capability. On Wayland, xdg-desktop-portal mediates per-call,
// per-app consent via the org.freedesktop.portal.ScreenCapture
// interface — the portal daemon running does NOT mean consent is
// granted for this app, so we surface StatusUnknown and let the
// first capture call surface the portal dialog.
//
// On X11 there is no OS-level permission gate (any X client can
// capture any other X client's window), so we report StatusGranted
// as a capability signal.
func probeScreenRecordingLinux() Permission {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		hasPortal := processRunning("xdg-desktop-portal") ||
			processRunning("xdg-desktop-portal-gnome") ||
			processRunning("xdg-desktop-portal-kde") ||
			processRunning("xdg-desktop-portal-wlr")
		if hasPortal {
			return Permission{
				Kind:   KindScreenRecording,
				Status: StatusUnknown,
				Note:   "xdg-desktop-portal detected (Wayland); per-app consent is granted at each capture call via the portal dialog",
			}
		}
		return Permission{
			Kind:   KindScreenRecording,
			Status: StatusUnknown,
			Note:   "Wayland requires xdg-desktop-portal; install xdg-desktop-portal + a backend (gnome/kde/wlr)",
		}
	}
	if os.Getenv("DISPLAY") != "" {
		return Permission{Kind: KindScreenRecording, Status: StatusGranted, Note: "X11 session detected (no permission required)"}
	}
	return Permission{
		Kind:   KindScreenRecording,
		Status: StatusUnknown,
		Note:   "unable to detect display server; screen capture may not be available",
	}
}

// probeMicrophoneLinux checks PulseAudio or PipeWire for audio
// input devices and whether the user has sound device access.
func probeMicrophoneLinux() Permission {
	if processRunning("pulseaudio") || processRunning("pipewire") || processRunning("pipewire-pulse") {
		return Permission{Kind: KindMicrophone, Status: StatusGranted, Note: "PulseAudio / PipeWire audio server detected"}
	}
	if _, err := os.Stat("/dev/snd"); err == nil {
		return Permission{Kind: KindMicrophone, Status: StatusGranted, Note: "/dev/snd audio devices detected"}
	}
	return Permission{
		Kind:   KindMicrophone,
		Status: StatusUnknown,
		Note:   "add user to audio group (sudo usermod -aG audio $USER) and ensure PulseAudio/PipeWire is running",
	}
}

// probeAutomationLinux checks for AT-SPI2 which provides
// accessibility scripting; no separate OS-level grant is needed.
func probeAutomationLinux() Permission {
	result := probeAccessibilityLinux()
	result.Kind = KindAutomation
	if result.Status == StatusGranted {
		result.Note = "AT-SPI2 provides accessibility scripting; no separate OS grant required"
	}
	return result
}

// probeNotificationsLinux checks for the D-Bus notification
// service (org.freedesktop.Notifications) and libnotify.
func probeNotificationsLinux() Permission {
	if dbusServiceAccessible("org.freedesktop.Notifications") {
		return Permission{Kind: KindNotifications, Status: StatusGranted, Note: "D-Bus notification service detected"}
	}
	if processRunning("dunst") || processRunning("mako") ||
		processRunning("notification-daemon") || processRunning("xfce4-notifyd") {
		return Permission{Kind: KindNotifications, Status: StatusGranted, Note: "notification daemon detected"}
	}
	return Permission{
		Kind:   KindNotifications,
		Status: StatusUnknown,
		Note:   "install libnotify and a notification daemon (dunst, mako, or notification-daemon)",
	}
}

// commandTimeout is the max wall-clock time a subprocess probe may run.
const commandTimeout = 2 * time.Second

// execProbe is the function used to spawn subprocess probes and
// read their stdout. Tests override this with a stub that returns
// canned output; the default spawns the real subprocess.
var execProbe = func(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // fixed probe names, not user input
	return cmd.Output()
}

// processRunning checks whether a process with the given name
// is running on the system.
func processRunning(name string) bool {
	out, err := execProbe("pgrep", "-x", name)
	return err == nil && len(out) > 0
}

// dbusServiceAccessible checks whether a D-Bus service is
// reachable via dbus-send or gdbus.
func dbusServiceAccessible(service string) bool {
	if _, err := exec.LookPath("dbus-send"); err == nil {
		_, err := execProbe("dbus-send", "--session", "--print-reply",
			"--dest="+service, "/", "org.freedesktop.DBus.Peer.Ping")
		return err == nil
	}
	if _, err := exec.LookPath("gdbus"); err == nil {
		_, err := execProbe("gdbus", "call", "--session",
			"--dest", service, "--object-path", "/",
			"--method", "org.freedesktop.DBus.Peer.Ping")
		return err == nil
	}
	return false
}
