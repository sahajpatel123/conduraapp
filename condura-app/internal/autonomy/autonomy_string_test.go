package autonomy

import "testing"

// TestLevel_String_KnownLevels pins the Stringer contract for
// every defined Level value. String() is used wherever the
// autonomy level surfaces to logs, SSE events, the GUI tray,
// and audit chains — a typo (e.g. "autonomus") would silently
// mismatch in user-facing UI without any test failure.
func TestLevel_String_KnownLevels(t *testing.T) {
	cases := []struct {
		name string
		lvl  Level
		want string
	}{
		{"Block", Block, "block"},
		{"Warn", Warn, "warn"},
		{"Ask", Ask, "ask"},
		{"Autonomous", Autonomous, "autonomous"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.lvl.String(); got != tc.want {
				t.Errorf("Level(%d).String() = %q, want %q", int(tc.lvl), got, tc.want)
			}
		})
	}
}

// TestLevel_String_UnsetDefaultsToWarn pins the documented
// sentinel behavior: Unset (Level = -1) is the "not configured
// yet" value, and its String() falls through the switch to the
// default case, returning "warn".
//
// This is the behavior the GUI tray and audit log rely on: an
// unset autonomy level should never render as the empty string
// or the raw int (-1), both of which would confuse the user.
// "warn" is the conservative default — a level that always
// requires consent + presence if the matrix lookup missed.
func TestLevel_String_UnsetDefaultsToWarn(t *testing.T) {
	if got := Unset.String(); got != "warn" {
		t.Errorf("Unset.String() = %q, want %q (default case)", got, "warn")
	}
}

// TestLevel_String_UnknownDefaultsToWarn pins the negative
// contract: any Level value outside the four named constants
// (including -1 / Unset and any future enum values that get
// added without updating String()) falls through to the
// default case and renders as "warn".
//
// This is intentionally NOT a test for "the default should
// panic" or "should return '<unknown>'" — those would change
// the user-facing contract. The current contract is
// fail-open-to-warn: unknown levels render as a level that
// requires consent + presence, which is the safe default for
// the GUI tray.
func TestLevel_String_UnknownDefaultsToWarn(t *testing.T) {
	// Level(99) is far outside the named range (Block=0 …
	// Autonomous=3, Unset=-1).
	unknown := Level(99)
	if got := unknown.String(); got != "warn" {
		t.Errorf("Level(99).String() = %q, want %q (default case)", got, "warn")
	}
}

// TestLevel_ConstantsAreDistinct pins the enum integrity:
// the five named constants must have distinct numeric values,
// and Unset must be negative (separated from the named-level
// range, so it can serve as the "before NewMatrix" sentinel
// in matrix evaluators).
//
// We deliberately do NOT pin the absolute numeric values —
// the codebase uses each constant by name in switch statements,
// so the relative ordering and distinctness are what matter.
// Pinning absolute values would block a future cleanup of the
// const block (e.g. consolidating the four `iota` repetitions
// into idiomatic Go form), which is a stylistic-only change
// with no functional impact.
func TestLevel_ConstantsAreDistinct(t *testing.T) {
	values := map[Level]string{
		Unset:      "Unset",
		Block:      "Block",
		Warn:       "Warn",
		Ask:        "Ask",
		Autonomous: "Autonomous",
	}
	if got := len(values); got != 5 {
		t.Errorf("expected 5 distinct Level constants, got %d (collision in enum)", got)
	}

	// Unset must be negative — NewMatrix uses
	// `if defaultLevel == Unset { defaultLevel = Warn }` as
	// the sentinel check; if Unset were >= 0 it would collide
	// with one of the valid levels and the substitution would
	// silently fail for the matching default.
	if Unset >= 0 {
		t.Errorf("Unset = %d, must be < 0 (sentinel collides with a valid level)", int(Unset))
	}
}
