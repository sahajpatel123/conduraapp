package perception

import (
	"context"
	"strings"
	"testing"
)

func TestStrategy_String(t *testing.T) {
	cases := []struct {
		s    Strategy
		want string
	}{
		{StrategyNone, "none"},
		{StrategyAXOnly, "ax_only"},
		{StrategyWindowRect, "window_rect"},
		{StrategyDifferential, "differential"},
		{StrategyFullScreen, "full_screen"},
		{StrategyVisionCUA, "vision_cua"},
		{Strategy(99), "unknown(99)"},
	}
	for _, tc := range cases {
		if got := tc.s.String(); got != tc.want {
			t.Errorf("Strategy(%d).String() = %q, want %q", tc.s, got, tc.want)
		}
	}
}

func TestStrategy_EnergyCost(t *testing.T) {
	if StrategyNone.EnergyCost() >= 0.01 {
		t.Error("None should be near-zero cost")
	}
	if StrategyAXOnly.EnergyCost() >= 0.2 {
		t.Error("AXOnly should be ~10x cheaper than full")
	}
	if StrategyFullScreen.EnergyCost() != 1.0 {
		t.Errorf("FullScreen baseline = %v, want 1.0", StrategyFullScreen.EnergyCost())
	}
	if StrategyVisionCUA.EnergyCost() < 30.0 {
		t.Errorf("VisionCUA should be expensive, got %v", StrategyVisionCUA.EnergyCost())
	}
}

func TestEnergyMode_String(t *testing.T) {
	cases := []struct {
		m    EnergyMode
		want string
	}{
		{EnergyAuto, "auto"},
		{EnergyLow, "low"},
		{EnergyBalanced, "balanced"},
		{EnergyHigh, "high"},
	}
	for _, tc := range cases {
		if got := tc.m.String(); got != tc.want {
			t.Errorf("EnergyMode(%d).String() = %q, want %q", tc.m, got, tc.want)
		}
	}
}

func TestEnergyMode_SessionBudget(t *testing.T) {
	if EnergyLow.sessionBudget() != 20.0 {
		t.Errorf("Low budget = %v, want 20", EnergyLow.sessionBudget())
	}
	if EnergyBalanced.sessionBudget() != 50.0 {
		t.Errorf("Balanced budget = %v, want 50", EnergyBalanced.sessionBudget())
	}
	if EnergyHigh.sessionBudget() != 100.0 {
		t.Errorf("High budget = %v, want 100", EnergyHigh.sessionBudget())
	}
}

func TestEnergyMode_AllowedStrategies(t *testing.T) {
	low := EnergyLow.allowedStrategies()
	if low[StrategyVisionCUA] {
		t.Error("Low should not allow vision CUA")
	}
	if !low[StrategyAXOnly] {
		t.Error("Low must allow AX-only")
	}
	if low[StrategyFullScreen] {
		t.Error("Low should not allow full screen")
	}
	high := EnergyHigh.allowedStrategies()
	if !high[StrategyVisionCUA] {
		t.Error("High should allow vision CUA")
	}
	if !high[StrategyFullScreen] {
		t.Error("High should allow full screen")
	}
}

func TestSmartCapturer_ChooseStrategy_ElementIdentity(t *testing.T) {
	c := NewSmartCapturer(EnergyBalanced)
	got, err := c.ChooseStrategy(
		Question{NeedsElementIdentity: true},
		DirtyState{},
	)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// Element-identity without pixels: AX is the right pick.
	if got != StrategyAXOnly {
		t.Errorf("got %v, want AXOnly", got)
	}
}

func TestSmartCapturer_ChooseStrategy_RespectsMode(t *testing.T) {
	c := NewSmartCapturer(EnergyLow)
	// Exhaust the Low budget. Low = 20.0; AX = 0.1 per call;
	// 250 calls uses 25.0, which exceeds the budget.
	for i := 0; i < 250; i++ {
		c.Record(StrategyAXOnly)
	}
	// Budget is now exceeded; any new question should error.
	_, err := c.ChooseStrategy(
		Question{NeedsElementIdentity: true},
		DirtyState{},
	)
	if err == nil {
		t.Error("expected budget-exhausted error after Low budget is depleted")
	}
}

func TestSmartCapturer_RecordDebitsBudget(t *testing.T) {
	c := NewSmartCapturer(EnergyBalanced)
	initial := c.Budget()
	c.Record(StrategyAXOnly)
	if c.Used() < 0.05 {
		t.Errorf("expected Used > 0 after AXOnly, got %v", c.Used())
	}
	if c.Budget() != initial {
		t.Errorf("Budget changed: %v -> %v", initial, c.Budget())
	}
}

func TestSmartCapturer_ResolveAuto(t *testing.T) {
	c := NewSmartCapturer(EnergyAuto)
	c.ResolveAuto(true) // plugged in
	if c.Mode() != EnergyHigh {
		t.Errorf("auto+plugged = %v, want High", c.Mode())
	}
	c2 := NewSmartCapturer(EnergyAuto)
	c2.ResolveAuto(false) // battery
	if c2.Mode() != EnergyLow {
		t.Errorf("auto+battery = %v, want Low", c2.Mode())
	}
	// Non-Auto mode ignores ResolveAuto.
	c3 := NewSmartCapturer(EnergyBalanced)
	c3.ResolveAuto(true)
	if c3.Mode() != EnergyBalanced {
		t.Error("non-Auto mode should ignore ResolveAuto")
	}
}

func TestDirtyTracker_MarkAndSnapshot(t *testing.T) {
	d := NewDirtyTracker()
	if d.Snapshot().Dirty {
		t.Error("fresh tracker should not be dirty")
	}
	d.Mark("Safari")
	if !d.Snapshot().Dirty {
		t.Error("after Mark, should be dirty")
	}
	if d.Snapshot().LastApp != "Safari" {
		t.Errorf("lastApp=%q want Safari", d.Snapshot().LastApp)
	}
	d.Clear()
	if d.Snapshot().Dirty {
		t.Error("after Clear, should not be dirty")
	}
}

func TestPIIRedactor_Email(t *testing.T) {
	r := NewPIIRedactor()
	got, err := r.Redact(context.Background(), "contact me at jane.doe+test@example.co.uk please")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "jane.doe+test@example.co.uk") {
		t.Errorf("email not redacted: %q", got)
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Errorf("redaction marker missing: %q", got)
	}
}

func TestPIIRedactor_SSN(t *testing.T) {
	r := NewPIIRedactor()
	got, err := r.Redact(context.Background(), "SSN: 123-45-6789 (test)")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "123-45-6789") {
		t.Errorf("SSN not redacted: %q", got)
	}
}

func TestPIIRedactor_CreditCard(t *testing.T) {
	r := NewPIIRedactor()
	got, err := r.Redact(context.Background(), "card 4111 1111 1111 1111 ok")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "4111 1111 1111 1111") {
		t.Errorf("CC not redacted: %q", got)
	}
}

func TestPIIRedactor_NoMatch(t *testing.T) {
	r := NewPIIRedactor()
	in := "the quick brown fox jumps over the lazy dog"
	got, err := r.Redact(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if got != in {
		t.Errorf("non-PII changed: %q -> %q", in, got)
	}
}

func TestPIIRedactor_NilSafe(t *testing.T) {
	var r *PIIRedactor
	got, err := r.Redact(context.Background(), "anything goes")
	if err != nil {
		t.Fatal(err)
	}
	if got != "anything goes" {
		t.Errorf("nil redactor should pass through: %q", got)
	}
}
