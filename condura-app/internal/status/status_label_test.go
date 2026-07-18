package status

import "testing"

// TestStatusLabel_NamedStates pins the title-cased Label contract for
// every named constant. The existing TestStatusLabel in status_test.go
// covers only StatusIdle + StatusListening; the remaining 4 (Thinking,
// Speaking, Halted, Error) had no coverage. These labels flow through
// the tray menu and overlay header — a regression in any one would
// produce a wrong word in the GUI.
func TestStatusLabel_NamedStates(t *testing.T) {
	tests := []struct {
		s    Status
		want string
	}{
		{StatusIdle, "Idle"},
		{StatusListening, "Listening..."},
		{StatusThinking, "Thinking..."},
		{StatusSpeaking, "Speaking..."},
		{StatusHalted, "Halted"},
		{StatusError, "Error"},
	}
	for _, tt := range tests {
		if got := tt.s.Label(); got != tt.want {
			t.Errorf("Status(%d).Label() = %q, want %q", int(tt.s), got, tt.want)
		}
	}
}

// TestStatusLabel_UnknownDefaultsToUnknown pins the default-branch
// contract: a Status outside the named-constant range MUST return
// "Unknown" (Title-case, matching the label style). A regression that
// returned the empty string or the lowercase "unknown" would surface
// to the GUI as a blank menu item or an inconsistent-style label.
func TestStatusLabel_UnknownDefaultsToUnknown(t *testing.T) {
	cases := []Status{
		Status(-1),
		Status(99),
		Status(1 << 20),
	}
	for _, s := range cases {
		if got := s.Label(); got != "Unknown" {
			t.Errorf("Status(%d).Label() = %q, want %q", int(s), got, "Unknown")
		}
	}
}

// TestStatusLabel_ErrorHasNoEllipsis pins the "Error" vs "Error..."
// convention: StatusError is a PERMANENT state (the agent is in
// degraded mode until the user intervenes), so its label carries NO
// trailing ellipsis. The three in-progress states (Listening, Thinking,
// Speaking) DO carry "..." because the agent is actively doing
// something. The Halted state is also permanent and matches Error's
// no-ellipsis style.
//
// A regression where every Label got "..." appended (or where the
// source-copier missed the exception) would mislead the user about
// whether the agent is actively doing work.
func TestStatusLabel_ErrorHasNoEllipsis(t *testing.T) {
	// States that are PERMANENT (not in-progress): no ellipsis.
	permanent := []Status{StatusIdle, StatusHalted, StatusError}
	for _, s := range permanent {
		got := s.Label()
		if len(got) > 0 && got[len(got)-3:] == "..." {
			t.Errorf("permanent status %d Label() = %q; permanent states must not end with '...'", int(s), got)
		}
	}
	// States that are IN-PROGRESS (active work): with ellipsis.
	inProgress := []Status{StatusListening, StatusThinking, StatusSpeaking}
	for _, s := range inProgress {
		got := s.Label()
		if len(got) < 3 || got[len(got)-3:] != "..." {
			t.Errorf("in-progress status %d Label() = %q; in-progress states must end with '...'", int(s), got)
		}
	}
}

// TestStatus_StringVsLabel_CasingDivergence pins the convention that
// String() returns lowercase (audit-safe, log-safe, file-name-safe)
// while Label() returns Title-case (UI-display-safe). A regression that
// unified the two (or accidentally introduced title-case into String,
// or lowercase into Label) would silently change every log line and
// every tray menu item.
func TestStatus_StringVsLabel_CasingDivergence(t *testing.T) {
	// For the named states: String must be the lowercase form of
	// Label-without-the-trailing-ellipsis. (Listening/Thinking/
	// Speaking all have '...' in Label; the String drops it.)
	all := []Status{
		StatusIdle, StatusListening, StatusThinking,
		StatusSpeaking, StatusHalted, StatusError,
	}
	for _, s := range all {
		str := s.String()
		// String must be all-lowercase ASCII.
		for i, r := range str {
			if r >= 'A' && r <= 'Z' {
				t.Errorf("Status(%d).String() = %q contains uppercase at byte %d (%q); String must be lowercase", int(s), str, i, r)
			}
		}
		// Label must NOT be all-lowercase (must have at least one Title-case letter).
		lab := s.Label()
		hasUpper := false
		for _, r := range lab {
			if r >= 'A' && r <= 'Z' {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			t.Errorf("Status(%d).Label() = %q contains no uppercase; Label must be Title-case", int(s), lab)
		}
	}
}

// TestStatus_EnumIntegrity pins three structural invariants of the
// Status enum that callers depend on:
//
//  1. All six named constants have DISTINCT numeric values (otherwise
//     switch cases in IsActive/String/Label could silently alias).
//  2. The values are SEQUENTIAL 0..5 starting at StatusIdle (because
//     the const block uses bare iota). This matters for any code
//     that iterates Status by int (e.g., metrics dashboards) or
//     serializes a Status as an int.
//  3. StatusIdle is the ZERO value (Status{} == StatusIdle). This
//     matters because every uninitialized Status field — in the
//     tray, the overlay, the voice pipeline — MUST default to
//     "idle" (the safe default) rather than "error" or some other
//     non-default state. A regression that reorders the const block
//     or inserts a new zero-value constant would break this invariant
//     silently.
func TestStatus_EnumIntegrity(t *testing.T) {
	// 1. All distinct.
	seen := map[Status]string{}
	all := []Status{
		StatusIdle, StatusListening, StatusThinking,
		StatusSpeaking, StatusHalted, StatusError,
	}
	for _, s := range all {
		if _, dup := seen[s]; dup {
			t.Errorf("duplicate Status value %d for %q (already seen as %q)", int(s), nameOf(s), seen[s])
		}
		seen[s] = nameOf(s)
	}

	// 2. Sequential 0..5 starting at StatusIdle.
	for want, s := range all {
		if int(s) != want {
			t.Errorf("Status(%s) = %d, want %d (sequential iota)", nameOf(s), int(s), want)
		}
	}

	// 3. StatusIdle is the zero value of Status.
	var zero Status
	if zero != StatusIdle {
		t.Errorf("Status{} zero value = %d, want %d (StatusIdle)", int(zero), int(StatusIdle))
	}
}

// nameOf returns the canonical name of a Status for error messages.
// Used only by TestStatus_EnumIntegrity to give readable failure
// messages ("duplicate Status value 3 for StatusSpeaking" rather
// than "duplicate Status value 3").
func nameOf(s Status) string {
	switch s {
	case StatusIdle:
		return "StatusIdle"
	case StatusListening:
		return "StatusListening"
	case StatusThinking:
		return "StatusThinking"
	case StatusSpeaking:
		return "StatusSpeaking"
	case StatusHalted:
		return "StatusHalted"
	case StatusError:
		return "StatusError"
	default:
		return "unknown"
	}
}
