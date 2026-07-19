package presence

import "testing"

// TestParseHIDIdleTime verifies the macOS ioreg HIDIdleTime parser
// (the real Darwin idle check). ioreg emits lines like
// `      "HIDIdleTime" = 314208851958` (nanoseconds).
func TestParseHIDIdleTime(t *testing.T) {
	tests := []struct {
		in   string
		want int64
		ok   bool
	}{
		{`      "HIDIdleTime" = 314208851958`, 314208851958, true},
		{`      "HIDIdleTime" = 0`, 0, true},
		{`  +-o IOHIDSystem  {...}`, 0, false},
		{``, 0, false},
		{`      "HIDIdleTime" = notanumber`, 0, false},
		{`      "OtherProperty" = 42`, 0, false},
	}
	for _, tt := range tests {
		got, ok := parseHIDIdleTime(tt.in)
		if ok != tt.ok {
			t.Errorf("parseHIDIdleTime(%q) ok=%v want %v", tt.in, ok, tt.ok)
			continue
		}
		if ok && got != tt.want {
			t.Errorf("parseHIDIdleTime(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

// TestCheckActiveOnLinux_FailClosed: the Linux idle probe is a
// placeholder (real X11/AT-SPI probe is v0.2.0). It MUST fail closed
// (return false) so a DESTRUCTIVE action on Linux is never auto-allowed
// by a probe that falsely claims the user is present.
func TestCheckActiveOnLinux_FailClosed(t *testing.T) {
	d := &Detector{}
	if d.checkActiveOnLinux() {
		t.Fatal("Linux checkActive must fail-closed (return false), not claim present")
	}
}

// TestParseHIDIdleTime_AmongOtherLines pins the search
// contract: parseHIDIdleTime MUST scan all lines for the
// "HIDIdleTime" key (not just the first line). Real ioreg
// output has many lines; the right line must be found.
func TestParseHIDIdleTime_AmongOtherLines(t *testing.T) {
	out := `Some other key = 42
Another random key = hello
      "HIDIdleTime" = 999999999
Yet another key = false`
	got, ok := parseHIDIdleTime(out)
	if !ok {
		t.Fatal("parseHIDIdleTime returned ok=false; want true")
	}
	if got != 999999999 {
		t.Errorf("parseHIDIdleTime = %d, want 999999999", got)
	}
}

// TestParseHIDIdleTime_NoEqualsReturnsFalse pins the
// malformed-line contract: a line containing "HIDIdleTime"
// but without "=" (corrupted ioreg output) MUST return
// ok=false. A regression that tried to slice past the
// non-existent "=" would panic.
func TestParseHIDIdleTime_NoEqualsReturnsFalse(t *testing.T) {
	out := `      "HIDIdleTime" 314208851958` // no =
	got, ok := parseHIDIdleTime(out)
	if ok {
		t.Errorf("parseHIDIdleTime = (%d, true); want (0, false)", got)
	}
	if got != 0 {
		t.Errorf("parseHIDIdleTime = %d; want 0 on malformed", got)
	}
}

// TestParseHIDIdleTime_HandlesLeadingAndTrailingWhitespace
// pins the whitespace-tolerance contract: the value after "="
// may have leading/trailing whitespace (ioreg is column-aligned).
// parseHIDIdleTime MUST trim before parsing.
func TestParseHIDIdleTime_HandlesLeadingAndTrailingWhitespace(t *testing.T) {
	out := `      "HIDIdleTime"   =   123456789   `
	got, ok := parseHIDIdleTime(out)
	if !ok {
		t.Fatal("parseHIDIdleTime returned ok=false; want true")
	}
	if got != 123456789 {
		t.Errorf("parseHIDIdleTime = %d, want 123456789", got)
	}
}

// TestParseHIDIdleTime_FirstHitWins pins the first-match-wins
// contract: if the input has multiple HIDIdleTime lines (which
// shouldn't happen in real ioreg but could in malformed input),
// parseHIDIdleTime MUST return the first one. A regression
// that returned the last would silently report wrong idle time.
func TestParseHIDIdleTime_FirstHitWins(t *testing.T) {
	out := `      "HIDIdleTime" = 100
      "HIDIdleTime" = 999`
	got, ok := parseHIDIdleTime(out)
	if !ok {
		t.Fatal("parseHIDIdleTime returned ok=false; want true")
	}
	if got != 100 {
		t.Errorf("parseHIDIdleTime = %d, want 100 (first match)", got)
	}
}

// TestParseHIDIdleTime_HIDIdleTimeAsSubstringNotFullKey pins
// the substring-match contract: any line containing
// "HIDIdleTime" as a substring (not just exact key match) is
// accepted. This is the production behavior — the source uses
// strings.Contains, not strings.HasPrefix. A regression to
// HasPrefix would miss lines like `"MyHIDIdleTime" = 42`.
// (A regression to a stricter match would be a defensive win,
// but the test pins the CURRENT contract so a refactor can
// change it explicitly.)
func TestParseHIDIdleTime_HIDIdleTimeAsSubstringNotFullKey(t *testing.T) {
	out := `      "MyHIDIdleTime" = 555`
	got, ok := parseHIDIdleTime(out)
	if !ok {
		t.Fatal("parseHIDIdleTime returned ok=false; want true (substring match)")
	}
	if got != 555 {
		t.Errorf("parseHIDIdleTime = %d, want 555", got)
	}
}
