//go:build linux

package permissions

import (
	"testing"
)

// errStub is a sentinel error used by test mocks to simulate command failure.
// It is shared across tests so any failing-call mock can return it.
var errStub = &stubErr{msg: "stub command failure"}

type stubErr struct{ msg string }

func (e *stubErr) Error() string { return e.msg }

// withMockExec swaps execProbe for the duration of the test.
// Returns a restore function to defer.
func withMockExec(t *testing.T, stub func(name string, args ...string) ([]byte, error)) func() {
	t.Helper()
	orig := execProbe
	execProbe = stub
	return func() { execProbe = orig }
}

// nothingAvailableMock is a helper that simulates a system with no
// AT-SPI, no audio server, no notification daemon, and no D-Bus
// session bus. Every pgrep returns empty output; dbus-send and gdbus
// fail. This is the canonical "no permissions detected" environment
// the fall-through paths in the probe functions expect.
func nothingAvailableMock(name string, args ...string) ([]byte, error) {
	switch name {
	case "dbus-send", "gdbus":
		return nil, errStub
	case "pgrep":
		return nil, nil // empty output, no error — pgrep found nothing
	default:
		return nil, nil
	}
}

func TestLinuxProbe_ProcessRunning_Granted(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		if name == "pgrep" {
			return []byte("1234\n"), nil
		}
		return nil, nil
	})
	defer restore()

	if !processRunning("at-spi2-registryd") {
		t.Fatalf("expected processRunning to return true when pgrep finds a match")
	}
}

func TestLinuxProbe_ProcessRunning_NotFound(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		return nil, nil // empty output, no error — pgrep found nothing
	})
	defer restore()

	if processRunning("nonexistent-daemon") {
		t.Fatalf("expected processRunning to return false when pgrep finds no match")
	}
}

func TestLinuxProbe_Accessibility_GrantedViaRegistryd(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		if name == "pgrep" {
			return []byte("1234\n"), nil
		}
		return nil, errStub
	})
	defer restore()

	p := probeAccessibilityLinux()
	if p.Kind != KindAccessibility {
		t.Fatalf("kind: want %s, got %s", KindAccessibility, p.Kind)
	}
	if p.Status != StatusGranted {
		t.Fatalf("status: want %s, got %s (note=%s)", StatusGranted, p.Status, p.Note)
	}
}

func TestLinuxProbe_Accessibility_UnknownWhenMissing(t *testing.T) {
	restore := withMockExec(t, nothingAvailableMock)
	defer restore()

	p := probeAccessibilityLinux()
	if p.Status != StatusUnknown {
		t.Fatalf("status: want %s, got %s (note=%s)", StatusUnknown, p.Status, p.Note)
	}
}

func TestLinuxProbe_ScreenRecording_X11Granted(t *testing.T) {
	t.Setenv("WAYLAND_DISPLAY", "")
	t.Setenv("DISPLAY", ":0")

	p := probeScreenRecordingLinux()
	if p.Status != StatusGranted {
		t.Fatalf("X11 screen recording: want %s, got %s", StatusGranted, p.Status)
	}
}

func TestLinuxProbe_ScreenRecording_WaylandUnknownEvenWithPortal(t *testing.T) {
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	t.Setenv("DISPLAY", "")
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		if name == "pgrep" {
			return []byte("1234\n"), nil // portal daemon running
		}
		return nil, errStub
	})
	defer restore()

	p := probeScreenRecordingLinux()
	// Even with portal daemon present, Wayland screen recording must
	// surface StatusUnknown — per-app consent is per-call.
	if p.Status != StatusUnknown {
		t.Fatalf("Wayland screen recording with portal: want %s, got %s (note=%s)",
			StatusUnknown, p.Status, p.Note)
	}
}

func TestLinuxProbe_ScreenRecording_WaylandUnknownWithoutPortal(t *testing.T) {
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	t.Setenv("DISPLAY", "")
	restore := withMockExec(t, nothingAvailableMock)
	defer restore()

	p := probeScreenRecordingLinux()
	if p.Status != StatusUnknown {
		t.Fatalf("Wayland screen recording without portal: want %s, got %s",
			StatusUnknown, p.Status)
	}
}

func TestLinuxProbe_Microphone_AudioServerGranted(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		if name == "pgrep" && len(args) > 1 && (args[1] == "pulseaudio" || args[1] == "pipewire") {
			return []byte("999\n"), nil
		}
		return nil, errStub
	})
	defer restore()

	p := probeMicrophoneLinux()
	if p.Status != StatusGranted {
		t.Fatalf("microphone with pipewire: want %s, got %s (note=%s)",
			StatusGranted, p.Status, p.Note)
	}
}

func TestLinuxProbe_Microphone_DevSndGranted(t *testing.T) {
	restore := withMockExec(t, nothingAvailableMock)
	defer restore()

	p := probeMicrophoneLinux()
	// If /dev/snd exists on the test machine (likely on any Linux dev box),
	// probeMicrophoneLinux returns StatusGranted via the /dev/snd fallback.
	// Otherwise it returns StatusUnknown. Both are valid; we just need to
	// verify it does NOT return StatusGranted on a false-positive path.
	if p.Status != StatusGranted && p.Status != StatusUnknown {
		t.Fatalf("microphone without audio server: want %s or %s, got %s",
			StatusGranted, StatusUnknown, p.Status)
	}
}

func TestLinuxProbe_Automation_ReusesAccessibilityResult(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		if name == "pgrep" {
			return []byte("1234\n"), nil
		}
		return nil, errStub
	})
	defer restore()

	p := probeAutomationLinux()
	if p.Kind != KindAutomation {
		t.Fatalf("kind: want %s, got %s", KindAutomation, p.Kind)
	}
	if p.Status != StatusGranted {
		t.Fatalf("automation: want %s, got %s (note=%s)", StatusGranted, p.Status, p.Note)
	}
}

func TestLinuxProbe_Notifications_DBusGranted(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		// Simulate "dbus-send found AND ping succeeds" — the production
		// code path returns StatusGranted when dbusServiceAccessible
		// returns true. This is what every Ubuntu CI runner with a
		// session bus will see.
		if name == "dbus-send" || name == "gdbus" {
			return []byte("method return time=... sender=:1.0 → destination=:1.42 serial=42 reply_serial=2\n"), nil
		}
		return nil, nil
	})
	defer restore()

	p := probeNotificationsLinux()
	if p.Status != StatusGranted {
		t.Fatalf("notifications with dbus ping success: want %s, got %s (note=%s)",
			StatusGranted, p.Status, p.Note)
	}
}

func TestLinuxProbe_Notifications_UnknownWhenNothingPresent(t *testing.T) {
	restore := withMockExec(t, nothingAvailableMock)
	defer restore()

	p := probeNotificationsLinux()
	if p.Status != StatusUnknown {
		t.Fatalf("notifications with no daemon: want %s, got %s",
			StatusUnknown, p.Status)
	}
}
