package autonomy

import "testing"

func TestMatrix_Evaluate(t *testing.T) {
	m := NewMatrix(Warn, map[string]Level{
		"chat.Code": Autonomous,
		"chat.*":    Ask,
	})
	if m.Evaluate("chat", "Terminal") != Ask {
		t.Error("expected Ask for chat.* wildcard")
	}
	if m.Evaluate("chat", "Code") != Autonomous {
		t.Error("expected Autonomous for specific chat.Code")
	}
	if m.Evaluate("browse", "Safari") != Warn {
		t.Error("expected default Warn")
	}
}

func TestMatrix_DefaultBlockIsHonored(t *testing.T) {
	m := NewMatrix(Block, nil)
	if m.Evaluate("chat", "Safari") != Block {
		t.Error("explicit Block default should be honored")
	}
}

func TestCanAutoApply_DestructiveCarveOut(t *testing.T) {
	if CanAutoApply(Autonomous, false) != true {
		t.Error("autonomous should apply to non-destructive")
	}
	if CanAutoApply(Autonomous, true) != false {
		t.Error("autonomous should NOT apply to destructive")
	}
}

func TestNeedsConsent(t *testing.T) {
	need, needPresence := NeedsConsent(Block, false)
	if !need || !needPresence {
		t.Error("block should need consent + presence")
	}
	need, needPresence = NeedsConsent(Ask, false)
	if !need || !needPresence {
		t.Error("ask should need consent + presence")
	}
	need, needPresence = NeedsConsent(Autonomous, false)
	if need || needPresence {
		t.Error("autonomous should not need consent for non-destructive")
	}
	need, needPresence = NeedsConsent(Autonomous, true)
	if !need || !needPresence {
		t.Error("autonomous+destructive should still need consent + presence")
	}
}
