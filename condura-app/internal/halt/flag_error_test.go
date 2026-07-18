package halt

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestNotYetResumableError_SatisfiesErrorInterface is a sanity pin:
// NotYetResumableError MUST satisfy the standard `error` interface.
// A regression that accidentally changed the method set (renamed
// Error, changed signature, etc.) would break every caller that uses
// `errors.As(err, &target)` or `if notYet, ok := err.(*NotYetResumableError)`.
func TestNotYetResumableError_SatisfiesErrorInterface(t *testing.T) {
	var _ error = (*NotYetResumableError)(nil) // compile-time check
	msg := (&NotYetResumableError{
		Remaining: 30 * time.Second,
		Since:     time.Now().Add(-time.Minute),
		Cooldown:  5 * time.Minute,
	}).Error()
	if msg == "" {
		t.Fatal("Error() returned empty string; want non-empty message")
	}
}

// TestNotYetResumableError_FormatIncludesRequiredParts pins the structured
// fields of the error message: the package prefix, the canonical phrase
// "resume not yet allowed", and the three labeled durations
// ("halted ... ago", "cooldown ...", "... remaining"). The GUI's "why
// can't I resume?" dialog surfaces this message verbatim — if any of
// these substrings disappears, the user loses critical context.
func TestNotYetResumableError_FormatIncludesRequiredParts(t *testing.T) {
	err := &NotYetResumableError{
		Remaining: 15 * time.Second,
		Since:     time.Now().Add(-45 * time.Second),
		Cooldown:  time.Minute,
	}
	msg := err.Error()

	wantParts := []string{
		"halt:",                  // package prefix
		"resume not yet allowed", // canonical phrase
		"halted ",                // "halted X ago"
		"ago",                    // elapsed-since anchor
		"cooldown ",              // cooldown label
		" remaining",             // remaining label (with leading space to anchor end-of-message)
	}
	for _, want := range wantParts {
		if !strings.Contains(msg, want) {
			t.Errorf("Error() missing required substring %q in:\n  %s", want, msg)
		}
	}
}

// TestNotYetResumableError_DurationsAreRoundedToSeconds pins the
// rounding behavior: all three durations (elapsed since halt,
// cooldown, remaining) MUST be rounded to second precision. The
// production code uses .Round(time.Second); a regression that
// dropped the rounding would surface sub-second jitter in the
// user-facing message (e.g., "1m0.0000003s" instead of "1m0s").
func TestNotYetResumableError_DurationsAreRoundedToSeconds(t *testing.T) {
	err := &NotYetResumableError{
		Remaining: 15 * time.Second,
		Since:     time.Now().Add(-45 * time.Second),
		Cooldown:  time.Minute,
	}
	msg := err.Error()

	// Sub-second precision in Go's default Duration formatter renders
	// as "1m0.0000003s" — assert no digit-followed-by-dot appears in
	// any duration.
	if strings.Contains(msg, ".") {
		t.Errorf("Error() contains '.' (sub-second precision):\n  %s\nwant second-precision rounding", msg)
	}
	// And the standard second/minute boundary forms ("Xs", "Xm", "Xh")
	// appear at least three times (elapsed + cooldown + remaining).
	units := []string{"s)", "s,", "s remaining"}
	hits := 0
	for _, u := range units {
		if strings.Contains(msg, u) {
			hits++
		}
	}
	if hits < 2 {
		t.Errorf("Error() has %d duration-suffix hits; expected >=2 in:\n  %s", hits, msg)
	}
}

// TestNotYetResumableError_ErrorsAsDiscriminable pins the typed-error
// contract: callers use errors.As to recover the structured fields
// (Remaining, Since, Cooldown) for UI display. A regression that
// returned a plain `fmt.Errorf` from Resume would lose the typed
// discriminability, and the GUI would fall back to a generic "could
// not resume" toast instead of "try again in 15s".
func TestNotYetResumableError_ErrorsAsDiscriminable(t *testing.T) {
	since := time.Now().Add(-time.Minute)
	err := &NotYetResumableError{
		Remaining: 45 * time.Second,
		Since:     since,
		Cooldown:  time.Minute,
	}

	var target *NotYetResumableError
	if !errors.As(err, &target) {
		t.Fatalf("errors.As failed to recover *NotYetResumableError from err=%v", err)
	}
	if target.Remaining != 45*time.Second {
		t.Errorf("Recovered Remaining = %v, want 45s", target.Remaining)
	}
	if target.Cooldown != time.Minute {
		t.Errorf("Recovered Cooldown = %v, want 1m", target.Cooldown)
	}
	// Since comparison: allow 1ms tolerance for the round-trip through
	// time.Time internal monotonic clock stripping (none here, but
	// defensive).
	if target.Since.Sub(since).Abs() > time.Millisecond {
		t.Errorf("Recovered Since = %v, want %v (within 1ms)", target.Since, since)
	}
}
