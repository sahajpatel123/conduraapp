package tray

import (
	"strings"
	"testing"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/status"
)

// TestSetHalted_TrueStoresAndReadsTrue pins the setter/getter
// round-trip: SetHalted(true) followed by IsHalted() MUST return
// true. The atomic.Bool storage is the source of truth for the
// pause/resume state across goroutines.
func TestSetHalted_TrueStoresAndReadsTrue(t *testing.T) {
	m := New("title", "tooltip")
	m.SetHalted(true)
	if !m.IsHalted() {
		t.Error("SetHalted(true) then IsHalted() = false; want true")
	}
}

// TestSetHalted_FalseStoresAndReadsFalse pins the inverse
// round-trip. Together with the true-case test, this catches
// the regression where SetHalted(false) was a no-op (e.g., a
// stale `if halted` branch).
func TestSetHalted_FalseStoresAndReadsFalse(t *testing.T) {
	m := New("title", "tooltip")
	m.SetHalted(false)
	if m.IsHalted() {
		t.Error("SetHalted(false) then IsHalted() = true; want false")
	}
}

// TestSetHalted_OverwriteFalseToTrue pins the overwrite contract:
// a subsequent SetHalted(true) MUST overwrite a previous
// SetHalted(false). A regression that early-returned on `false`
// would silently leave the flag stuck at the old value.
func TestSetHalted_OverwriteFalseToTrue(t *testing.T) {
	m := New("title", "tooltip")
	m.SetHalted(false)
	m.SetHalted(true)
	if !m.IsHalted() {
		t.Error("SetHalted(false) then SetHalted(true) did not overwrite")
	}
}

// TestIsHalted_DefaultIsFalse pins the zero-value contract: a
// freshly-constructed Menu MUST have IsHalted() == false. The
// GUI starts in the un-halted state.
func TestIsHalted_DefaultIsFalse(t *testing.T) {
	m := New("title", "tooltip")
	if m.IsHalted() {
		t.Error("fresh Menu IsHalted() = true; want false")
	}
}

// TestSetSpendUSD_StoresCentsAsInteger pins the float→cents
// conversion contract: SetSpendUSD(1.50) MUST store 150 cents
// internally. The atomic.Uint64 storage means callers can do
// a lock-free read; the displayed value in the tray icon is
// formatted from the int.
//
// Note: float→cents has known precision issues for some values
// (e.g., 0.10 * 100 = 10.000000000000002 due to IEEE 754).
// We test values that are exactly representable in float64.
func TestSetSpendUSD_StoresCentsAsInteger(t *testing.T) {
	m := New("title", "tooltip")
	m.SetSpendUSD(1.50)
	// Internal state: m.spend.Load() should be 150 (cents).
	if got := m.spend.Load(); got != 150 {
		t.Errorf("m.spend after SetSpendUSD(1.50) = %d, want 150 (cents)", got)
	}
}

// TestSetSpendUSD_ZeroStoresZero pins the zero-input contract:
// SetSpendUSD(0) MUST store 0 cents. Defense against a regression
// that converted NaN or negative values.
func TestSetSpendUSD_ZeroStoresZero(t *testing.T) {
	m := New("title", "tooltip")
	m.SetSpendUSD(0)
	if got := m.spend.Load(); got != 0 {
		t.Errorf("m.spend after SetSpendUSD(0) = %d, want 0", got)
	}
}

// TestSetErrorMessage_StoresMessage pins the contract:
// SetErrorMessage stores the message that SetStatus(StatusError)
// will use in the tooltip. Read back via the (unexported)
// errMsg field — we verify via the SetStatus pipeline below.
func TestSetErrorMessage_StoresMessage(t *testing.T) {
	m := New("title", "tooltip")
	m.SetErrorMessage("permission denied")
	got := m.errMsg.Load()
	if got != "permission denied" {
		t.Errorf("errMsg after SetErrorMessage = %v, want \"permission denied\"", got)
	}
}

// TestSetStatus_StoresValueAndUpdatesTooltip pins the
// single-source-of-truth contract: SetStatus MUST (1) store the
// status value (readable via Status()), (2) update the tooltip
// to reflect the new status. A regression that did only (1)
// would leave the tray tooltip stale.
//
// Pre-Start, m.mVoice is nil so the voice menu update is skipped;
// the tooltip update is the observable side effect.
func TestSetStatus_StoresValueAndUpdatesTooltip(t *testing.T) {
	m := New("title", "tooltip")
	m.SetStatus(status.StatusListening)
	if m.Status() != status.StatusListening {
		t.Errorf("Status() after SetStatus(Listening) = %v, want Listening", m.Status())
	}
	// Tooltip is set via SetTooltip (which assigns to m.tooltip).
	// We can't observe the systray.SetTooltip call (m.mShow is nil
	// pre-Start), but the m.tooltip field is set.
	if !strings.Contains(m.tooltip, "Condura") {
		t.Errorf("m.tooltip after SetStatus(Listening) = %q; want it to be updated", m.tooltip)
	}
}

// TestSetStatus_HaltedAlsoSetsHaltedFlag pins the sync contract:
// SetStatus(StatusHalted) MUST also set the halted flag (so
// SetHalted/IsHalted consumers see a consistent view). A
// regression that only set the status field would leave
// IsHalted() returning false even though Status() == Halted.
func TestSetStatus_HaltedAlsoSetsHaltedFlag(t *testing.T) {
	m := New("title", "tooltip")
	m.SetStatus(status.StatusHalted)
	if !m.IsHalted() {
		t.Error("IsHalted() after SetStatus(Halted) = false; want true")
	}
}

// TestSetStatus_NotHaltedClearsHaltedFlag pins the inverse
// sync: SetStatus(NotHalted) on a previously-halted Menu MUST
// clear the halted flag. A regression that didn't clear would
// leave the menu showing "Resume" even though the daemon is
// running again.
func TestSetStatus_NotHaltedClearsHaltedFlag(t *testing.T) {
	m := New("title", "tooltip")
	m.SetStatus(status.StatusHalted)
	if !m.IsHalted() {
		t.Fatal("setup: IsHalted should be true after SetStatus(Halted)")
	}
	m.SetStatus(status.StatusIdle)
	if m.IsHalted() {
		t.Error("IsHalted() after SetStatus(Idle) = true; want false (flag should clear)")
	}
}

// TestSetStatus_ErrorIncludesMessageInTooltip pins the
// error-message propagation contract: SetStatus(StatusError)
// with a previously-stored errMsg MUST include the message in
// the tooltip. A regression that ignored errMsg would show the
// user "see logs" with no hint of which log to check.
func TestSetStatus_ErrorIncludesMessageInTooltip(t *testing.T) {
	m := New("title", "tooltip")
	m.SetErrorMessage("oauth token expired")
	m.SetStatus(status.StatusError)
	if !strings.Contains(m.tooltip, "oauth token expired") {
		t.Errorf("m.tooltip after SetStatus(Error) with errMsg = %q; want it to include the errMsg",
			m.tooltip)
	}
}

// TestSetStatus_ErrorFallbackTooltip pins the fallback contract:
// SetStatus(StatusError) WITHOUT a previously-stored errMsg
// MUST still produce a sensible tooltip ("see logs"). A
// regression that produced an empty tooltip would leave the
// user with no signal.
func TestSetStatus_ErrorFallbackTooltip(t *testing.T) {
	m := New("title", "tooltip")
	// Note: SetErrorMessage NOT called.
	m.SetStatus(status.StatusError)
	if m.tooltip == "" {
		t.Error("m.tooltip after SetStatus(Error) without errMsg = empty; want a fallback message")
	}
	if !strings.Contains(m.tooltip, "Synaptic") {
		t.Errorf("m.tooltip = %q; want it to contain \"Synaptic\"", m.tooltip)
	}
}

// TestSetVoiceState_ListeningMapsToListening pins the
// backward-compatibility contract: SetVoiceState("listening")
// MUST map to SetStatus(StatusListening). A regression that
// hardcoded a status string would silently break the legacy
// callers that use the SetVoiceState API.
func TestSetVoiceState_ListeningMapsToListening(t *testing.T) {
	m := New("title", "tooltip")
	m.SetVoiceState("listening")
	if m.Status() != status.StatusListening {
		t.Errorf("Status() after SetVoiceState(\"listening\") = %v, want StatusListening",
			m.Status())
	}
}

// TestSetVoiceState_DefaultMapsToIdle pins the default
// dispatch contract: SetVoiceState(unknown) MUST map to
// SetStatus(StatusIdle) (the safe default). A regression that
// didn't recognize an unknown state would leave the tray stuck
// on the last status.
func TestSetVoiceState_DefaultMapsToIdle(t *testing.T) {
	m := New("title", "tooltip")
	m.SetVoiceState("totally-unknown-state")
	if m.Status() != status.StatusIdle {
		t.Errorf("Status() after SetVoiceState(\"totally-unknown-state\") = %v, want StatusIdle",
			m.Status())
	}
}

// TestEvents_ReturnsNonNilChannel pins the channel contract:
// Events() MUST return a non-nil channel (buffered, ready to
// receive). A regression that returned nil would cause
// "send on nil channel" panics at the first event.
func TestEvents_ReturnsNonNilChannel(t *testing.T) {
	m := New("title", "tooltip")
	if m.Events() == nil {
		t.Error("Events() = nil; want non-nil channel")
	}
	// Buffered channel should accept at least one send without
	// blocking (capacity 16 per New()).
	m.events <- Event(1)
}