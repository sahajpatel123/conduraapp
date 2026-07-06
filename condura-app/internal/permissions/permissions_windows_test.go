//go:build windows

package permissions

import (
	"testing"
)

// withMockExec swaps execProbe for the duration of the test.
func withMockExec(t *testing.T, stub func(name string, args ...string) ([]byte, error)) func() {
	t.Helper()
	orig := execProbe
	execProbe = stub
	return func() { execProbe = orig }
}

func TestWindowsProbe_ScreenRecording_AlwaysUnknown(t *testing.T) {
	// The Windows probe intentionally does not call execProbe — it
	// surfaces StatusUnknown because the WMI-based probe was a
	// false positive on every Windows desktop (hardware presence
	// != permission). Verify the constant contract.
	p := probeScreenRecordingWindows()
	if p.Status != StatusUnknown {
		t.Fatalf("status: want %s, got %s (note=%s)", StatusUnknown, p.Status, p.Note)
	}
	if p.Note == "" {
		t.Fatal("note must point users to the Settings pane")
	}
}

func TestWindowsProbe_Accessibility_Unknown(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		// Simulate UIAutomationClient.Assembly loading successfully.
		return []byte("UIAutomationClient\nOK\n"), nil
	})
	defer restore()

	p := probeAccessibilityWindows()
	if p.Status != StatusUnknown {
		t.Fatalf("status: want %s, got %s (note=%s)", StatusUnknown, p.Status, p.Note)
	}
}

func TestWindowsProbe_Accessibility_UnknownWhenProbeFails(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		return nil, &stubErr{msg: "powershell not found"}
	})
	defer restore()

	p := probeAccessibilityWindows()
	if p.Status != StatusUnknown {
		t.Fatalf("status: want %s, got %s", StatusUnknown, p.Status)
	}
}

func TestWindowsProbe_Microphone_DevicePresent_Unknown(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		return []byte("Realtek High Definition Audio\n"), nil
	})
	defer restore()

	p := probeMicrophoneWindows()
	// Device present, permission unknowable → StatusUnknown.
	if p.Status != StatusUnknown {
		t.Fatalf("status: want %s, got %s (note=%s)", StatusUnknown, p.Status, p.Note)
	}
}

func TestWindowsProbe_Microphone_NoDevice_Denied(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		return nil, nil // no audio device enumerated
	})
	defer restore()

	p := probeMicrophoneWindows()
	if p.Status != StatusDenied {
		t.Fatalf("status: want %s, got %s (note=%s)", StatusDenied, p.Status, p.Note)
	}
}

func TestWindowsProbe_Microphone_ProbeError_Unknown(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		return nil, &stubErr{msg: "powershell crashed"}
	})
	defer restore()

	p := probeMicrophoneWindows()
	if p.Status != StatusUnknown {
		t.Fatalf("status: want %s, got %s", StatusUnknown, p.Status)
	}
}

func TestWindowsProbe_Automation_AlwaysUnknown(t *testing.T) {
	// Automation on Windows shares UIA backend with accessibility;
	// there is no separate OS-level grant. Must always return Unknown.
	p := probeAutomationWindows()
	if p.Status != StatusUnknown {
		t.Fatalf("status: want %s, got %s (note=%s)", StatusUnknown, p.Status, p.Note)
	}
}

func TestWindowsProbe_Notifications_ToastAvailable_Unknown(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		return []byte("OK\n"), nil
	})
	defer restore()

	p := probeNotificationsWindows()
	if p.Status != StatusUnknown {
		t.Fatalf("status: want %s, got %s (note=%s)", StatusUnknown, p.Status, p.Note)
	}
}

func TestWindowsProbe_Notifications_NoToast_Unknown(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		return []byte("NO\n"), nil
	})
	defer restore()

	p := probeNotificationsWindows()
	if p.Status != StatusUnknown {
		t.Fatalf("status: want %s, got %s (note=%s)", StatusUnknown, p.Status, p.Note)
	}
}

func TestWindowsProbe_DispatchReturnsWellFormedStatus(t *testing.T) {
	for _, k := range []Kind{
		KindAccessibility, KindScreenRecording, KindMicrophone,
		KindAutomation, KindNotifications,
	} {
		p := windowsProbeOne(k)
		if p.Kind != k {
			t.Errorf("%s: kind echo wrong; want %s, got %s", k, k, p.Kind)
		}
		switch p.Status {
		case StatusGranted, StatusDenied, StatusUnknown:
			// ok
		default:
			t.Errorf("%s: invalid status %q", k, p.Status)
		}
	}
}

// stubErr is a sentinel error used by test stubs to simulate command failure.
type stubErr struct{ msg string }

func (e *stubErr) Error() string { return e.msg }
