package permissions

import (
	"context"
	"strings"
	"testing"
)

// TestStepsFor_DispatchByPlatform pins the cross-platform
// dispatch contract: stepsFor MUST route to the platform-
// specific helper (darwinSteps, windowsSteps, linuxSteps)
// based on the platform argument. The runner (probe, RequestGuide)
// depends on this dispatch being correct — a regression
// would route Windows users to Mac instructions or vice
// versa.
func TestStepsFor_DispatchByPlatform(t *testing.T) {
	cases := []struct {
		name     string
		platform string
		kind     Kind
	}{
		{"darwin-accessibility", "darwin", KindAccessibility},
		{"darwin-microphone", "darwin", KindMicrophone},
		{"windows-accessibility", "windows", KindAccessibility},
		{"linux-microphone", "linux", KindMicrophone},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			steps, _, _ := stepsFor(tc.kind, tc.platform)
			if len(steps) == 0 {
				t.Errorf("stepsFor(%q, %q) returned 0 steps; want at least 1",
					tc.platform, tc.kind)
			}
			// Each platform-specific step MUST have at least
			// 2 lines of actionable content (defense: a
			// regression that returns a single "Open settings"
			// line would not give the user enough to act on).
			if len(steps) < 2 {
				t.Errorf("stepsFor(%q, %q) returned %d steps; want >= 2 (actionable content)",
					tc.platform, tc.kind, len(steps))
			}
		})
	}
}

// TestStepsFor_UnknownPlatformReturnsFallback pins the
// unknown-platform contract: when stepsFor receives an unknown
// platform string, it MUST return a single-step fallback
// mentioning the platform name. A regression that returned
// an empty slice would silently leave the user without
// instructions.
func TestStepsFor_UnknownPlatformReturnsFallback(t *testing.T) {
	steps, _, _ := stepsFor(KindAccessibility, "plan9")
	if len(steps) == 0 {
		t.Fatal("stepsFor(unknown) returned 0 steps; want fallback message")
	}
	if !strings.Contains(steps[0], "plan9") {
		t.Errorf("fallback step %q does not mention the unknown platform", steps[0])
	}
}

// TestDefaultProbeOne_ReturnsUnknownWithPlatformNote pins the
// fallback-probe contract: defaultProbeOne MUST return
// Status=StatusUnknown with a Note mentioning both the platform
// and the Kind. This is the safety net for platforms that
// haven't implemented a real probe — the GUI shows the user
// "we don't know how to check this on your platform".
func TestDefaultProbeOne_ReturnsUnknownWithPlatformNote(t *testing.T) {
	cases := []Kind{
		KindAccessibility,
		KindMicrophone,
		KindScreenRecording,
		KindAutomation,
		KindNotifications,
	}
	for _, k := range cases {
		t.Run(string(k), func(t *testing.T) {
			p := defaultProbeOne(k)
			if p.Status != StatusUnknown {
				t.Errorf("Status = %v, want StatusUnknown", p.Status)
			}
			if p.Kind != k {
				t.Errorf("Kind = %v, want %v", p.Kind, k)
			}
			if !strings.Contains(p.Note, string(k)) {
				t.Errorf("Note %q does not mention Kind %q", p.Note, k)
			}
		})
	}
}

// TestCheck_ReturnsValidStatus pins the Check contract: Check(k)
// MUST return one of the defined Status values (not an arbitrary
// int). The GUI uses Check's return value in a switch statement;
// an invalid int would fall through to a default branch (or
// crash the switch). We pin the contract that Check returns
// a value from the Status enum.
func TestCheck_ReturnsValidStatus(t *testing.T) {
	// All defined Status values.
	valid := map[Status]bool{
		StatusUnknown: true,
		StatusGranted: true,
		StatusDenied:  true,
	}

	got := Check(KindAccessibility)
	if !valid[got] {
		t.Errorf("Check = %q; want a defined Status value", got)
	}

	// Also verify all the other Kinds return valid statuses.
	for _, k := range []Kind{KindAccessibility, KindMicrophone, KindScreenRecording, KindAutomation, KindNotifications} {
		if !valid[Check(k)] {
			t.Errorf("Check(%v) returned invalid Status %v", k, Check(k))
		}
	}
}

// TestOpenSettings_NilContextFallback pins the nil-context
// guard: OpenSettings MUST accept a nil context and fall back
// to context.Background(). A regression that removed this
// guard would pass a nil ctx to exec.CommandContext, which
// would panic on the first call from any caller that
// forgot to thread the context through.
//
// This is the most-tested surface for permissions and the
// nil-context path is easy to miss in caller code (the RPC
// handler chains contexts through several layers).
func TestOpenSettings_NilContextFallback(t *testing.T) {
	//nolint:staticcheck // SA1012: testing the nil-context fallback on purpose
	g, opened, err := OpenSettings(nil, KindAccessibility)
	if err != nil {
		t.Fatalf("OpenSettings(nil, KindAccessibility) returned error: %v", err)
	}
	// The guide is returned even when the OS open fails
	// (caller can still show the steps).
	if len(g.Steps) == 0 {
		t.Error("Guide has no steps; caller would show empty state")
	}
	// On darwin (where this runs), 'open' should succeed and
	// 'opened' should be true. On other platforms it may fail
	// but should never panic.
	_ = opened
}

// TestOpenSettings_ValidContextWorks pins the happy path:
// OpenSettings with a real context returns the guide and
// attempts the OS open. On darwin this should succeed.
func TestOpenSettings_ValidContextWorks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, opened, err := OpenSettings(ctx, KindMicrophone)
	if err != nil {
		t.Fatalf("OpenSettings(ctx, KindMicrophone) returned error: %v", err)
	}
	if len(g.Steps) == 0 {
		t.Error("Guide has no steps")
	}
	// 'opened' is platform-dependent; on darwin it should be true
	// (the 'open' command exists), on Linux it depends on xdg-open.
	_ = opened
}

// TestRequestGuide_ReturnsNonEmptyStepsForEachKind pins the
// per-kind content contract: RequestGuide MUST return a non-empty
// Steps slice for every Kind on the current platform. A regression
// that returned an empty Steps for one Kind would leave the GUI
// showing "no instructions" for that permission type.
func TestRequestGuide_ReturnsNonEmptyStepsForEachKind(t *testing.T) {
	for _, k := range []Kind{
		KindAccessibility,
		KindMicrophone,
		KindScreenRecording,
		KindAutomation,
		KindNotifications,
	} {
		t.Run(string(k), func(t *testing.T) {
			g := RequestGuide(k)
			if len(g.Steps) == 0 {
				t.Errorf("RequestGuide(%v).Steps is empty; want >= 1 actionable step", k)
			}
		})
	}
}
