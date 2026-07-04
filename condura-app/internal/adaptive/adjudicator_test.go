package adaptive

import "testing"

func TestAdjudicator_AutoApply(t *testing.T) {
	a := NewAdjudicator([]string{"verbosity", "response_length"}, []string{"communication_style"}, 0.6)

	p := Proposal{Category: "verbosity", Value: "concise", Confidence: 0.8}
	if a.Evaluate(p) != DecisionAutoApply {
		t.Error("expected auto_apply for autoApply category")
	}
}

func TestAdjudicator_RequireConfirm(t *testing.T) {
	a := NewAdjudicator([]string{"verbosity"}, []string{"communication_style"}, 0.6)

	p := Proposal{Category: "communication_style", Value: "casual", Confidence: 0.9}
	if a.Evaluate(p) != DecisionRequireConfirm {
		t.Error("expected require_confirm for requireConfirm category")
	}
}

func TestAdjudicator_DiscardLowConfidence(t *testing.T) {
	a := NewAdjudicator([]string{"verbosity"}, nil, 0.6)

	p := Proposal{Category: "verbosity", Value: "concise", Confidence: 0.3}
	if a.Evaluate(p) != DecisionDiscard {
		t.Error("expected discard for low confidence")
	}
}

func TestAdjudicator_UnknownCategory(t *testing.T) {
	a := NewAdjudicator([]string{"verbosity"}, nil, 0.6)

	p := Proposal{Category: "unknown_thing", Value: "x", Confidence: 0.9}
	if a.Evaluate(p) != DecisionRequireConfirm {
		t.Error("expected require_confirm for unknown category (safe default)")
	}
}

func TestAdjudicator_Filter(t *testing.T) {
	a := NewAdjudicator([]string{"verbosity"}, []string{"communication_style"}, 0.6)

	proposals := []Proposal{
		{Category: "verbosity", Confidence: 0.8, Value: "concise"},
		{Category: "communication_style", Confidence: 0.9, Value: "casual"},
		{Category: "new_skill", Confidence: 0.3, Value: "skill-x"},
	}
	auto, confirm, discarded := a.Filter(proposals)
	if len(auto) != 1 {
		t.Errorf("auto = %d", len(auto))
	}
	if len(confirm) != 1 {
		t.Errorf("confirm = %d", len(confirm))
	}
	if discarded != 1 {
		t.Errorf("discarded = %d", discarded)
	}
}
