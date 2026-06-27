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
