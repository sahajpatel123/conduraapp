package status

import "testing"

func TestStatusString(t *testing.T) {
	tests := []struct {
		s    Status
		want string
	}{
		{StatusIdle, "idle"},
		{StatusListening, "listening"},
		{StatusThinking, "thinking"},
		{StatusSpeaking, "speaking"},
		{StatusHalted, "halted"},
		{StatusError, "error"},
		{Status(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.s.String(); got != tt.want {
			t.Errorf("Status(%d).String() = %q, want %q", int(tt.s), got, tt.want)
		}
	}
}

func TestStatusLabel(t *testing.T) {
	if got := StatusIdle.Label(); got != "Idle" {
		t.Errorf("StatusIdle.Label() = %q, want %q", got, "Idle")
	}
	if got := StatusListening.Label(); got != "Listening..." {
		t.Errorf("StatusListening.Label() = %q, want %q", got, "Listening...")
	}
}

func TestStatusIsActive(t *testing.T) {
	active := []Status{StatusListening, StatusThinking, StatusSpeaking}
	for _, s := range active {
		if !s.IsActive() {
			t.Errorf("Status(%v).IsActive() = false, want true", s)
		}
	}
	inactive := []Status{StatusIdle, StatusHalted, StatusError}
	for _, s := range inactive {
		if s.IsActive() {
			t.Errorf("Status(%v).IsActive() = true, want false", s)
		}
	}
}
