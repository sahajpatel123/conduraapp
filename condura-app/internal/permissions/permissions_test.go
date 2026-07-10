package permissions

import (
	"context"
	"runtime"
	"testing"
)

func TestProbe_ReturnsAllKinds(t *testing.T) {
	perms, err := Probe(context.Background())
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}
	if len(perms) != 5 {
		t.Fatalf("expected 5 kinds, got %d", len(perms))
	}
	wantKinds := map[Kind]bool{
		KindAccessibility:   false,
		KindScreenRecording: false,
		KindMicrophone:      false,
		KindAutomation:      false,
		KindNotifications:   false,
	}
	for _, p := range perms {
		if _, ok := wantKinds[p.Kind]; !ok {
			t.Fatalf("unexpected kind: %s", p.Kind)
		}
		wantKinds[p.Kind] = true
	}
	for k, found := range wantKinds {
		if !found {
			t.Fatalf("missing kind: %s", k)
		}
	}
}

func TestCheck_ReturnsKnownStatus(t *testing.T) {
	for _, k := range []Kind{
		KindAccessibility, KindScreenRecording, KindMicrophone,
		KindAutomation, KindNotifications,
	} {
		s := Check(k)
		switch s {
		case StatusGranted, StatusDenied, StatusUnknown:
			// ok
		default:
			t.Fatalf("Check(%s): invalid status %q", k, s)
		}
	}
}

func TestRequestGuide_ReturnsPerKind(t *testing.T) {
	for _, k := range []Kind{
		KindAccessibility, KindScreenRecording, KindMicrophone,
		KindAutomation, KindNotifications,
	} {
		g := RequestGuide(k)
		if g.Kind != k {
			t.Fatalf("kind: want %s, got %s", k, g.Kind)
		}
		if g.Platform != runtime.GOOS {
			t.Fatalf("platform: want %s, got %s", runtime.GOOS, g.Platform)
		}
		if g.Title == "" {
			t.Fatalf("title empty for %s", k)
		}
		if len(g.Steps) == 0 {
			t.Fatalf("steps empty for %s", k)
		}
	}
}

func TestPlatform_MatchesRuntime(t *testing.T) {
	if got := Platform(); got != runtime.GOOS {
		t.Fatalf("Platform: want %s, got %s", runtime.GOOS, got)
	}
}

func TestNewManager_AlwaysSucceeds(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatalf("NewManager returned nil")
	}
}

func TestOpenDeepLink_EmptyReturnsFalse(t *testing.T) {
	opened, err := openDeepLink(context.Background(), "", "", "darwin")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if opened {
		t.Fatal("expected opened=false for empty targets")
	}
}

func TestOpenSettings_ReturnsGuide(t *testing.T) {
	// We don't assert opened=true (CI may not have a display / open may fail).
	g, _, _ := OpenSettings(context.Background(), KindAccessibility)
	if g.Title == "" {
		t.Fatal("expected non-empty guide title")
	}
	if len(g.Steps) == 0 {
		t.Fatal("expected steps")
	}
}
