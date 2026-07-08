//go:build darwin

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

func TestDarwinProbe_Automation_Granted(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		if name == "osascript" {
			return []byte("System Events got version successfully\n"), nil
		}
		return nil, nil
	})
	defer restore()

	p := probeAutomation()
	if p.Kind != KindAutomation {
		t.Fatalf("kind: want %s, got %s", KindAutomation, p.Kind)
	}
	if p.Status != StatusGranted {
		t.Fatalf("status: want %s, got %s (note=%s)", StatusGranted, p.Status, p.Note)
	}
}

func TestDarwinProbe_Automation_UnknownWhenOsascriptFails(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		return nil, errStub // osascript not on path or denied
	})
	defer restore()

	p := probeAutomation()
	if p.Status != StatusUnknown {
		t.Fatalf("status: want %s, got %s", StatusUnknown, p.Status)
	}
}

// errStub is a sentinel error used by test stubs to simulate command failure.
var errStub = &stubError{"stub error"}

type stubError struct{ msg string }

func (e *stubError) Error() string { return e.msg }

func TestDarwinProbe_Automation_UnknownWhenStderrReturned(t *testing.T) {
	restore := withMockExec(t, func(name string, args ...string) ([]byte, error) {
		return []byte("Not authorized to send Apple events to System Events.\n"), errStub
	})
	defer restore()

	p := probeAutomation()
	if p.Status != StatusUnknown {
		t.Fatalf("status: want %s, got %s (note=%s)", StatusUnknown, p.Status, p.Note)
	}
}

// Cgo-backed probes (accessibility, screen_recording, microphone,
// notifications) cannot be fully exercised without a real TCC
// environment, but they MUST NOT panic or block when the cgo
// preamble is unreachable. These smoke tests verify the dispatch
// path works for each kind and that the returned Permission has a
// well-formed status.

func TestDarwinProbe_DispatchReturnsWellFormedStatus(t *testing.T) {
	for _, k := range []Kind{
		KindAccessibility, KindScreenRecording, KindMicrophone,
		KindAutomation, KindNotifications,
	} {
		p := darwinProbeOne(k)
		if p.Kind != k {
			t.Errorf("%s: kind echo wrong; want %s, got %s", k, k, p.Kind)
		}
		switch p.Status {
		case StatusGranted, StatusDenied, StatusUnknown:
			// ok
		default:
			t.Errorf("%s: invalid status %q", k, p.Status)
		}
		if p.Note == "" {
			t.Errorf("%s: note should be non-empty even for granted status", k)
		}
	}
}

func TestDarwinProbe_Microphone_NoteContainsAPIBaseline(t *testing.T) {
	// AVCaptureDevice.authorizationStatusForMediaType is a synchronous
	// cgo call. Either status is acceptable depending on the test
	// machine's TCC state, but the note should mention the API used
	// so users can debug.
	p := probeMicrophone()
	if p.Note == "" {
		t.Fatal("note should describe how the status was determined")
	}
	if !contains(p.Note, "AVCaptureDevice") {
		t.Fatalf("note should mention the underlying API; got %q", p.Note)
	}
}

func TestDarwinProbe_Notifications_NoteContainsAPIBaseline(t *testing.T) {
	p := probeNotifications()
	if p.Note == "" {
		t.Fatal("note should describe how the status was determined")
	}
	if !contains(p.Note, "UNUserNotificationCenter") {
		t.Fatalf("note should mention the underlying API; got %q", p.Note)
	}
}

// contains is a tiny stdlib wrapper kept here so this test file
// doesn't need an extra import.
func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
