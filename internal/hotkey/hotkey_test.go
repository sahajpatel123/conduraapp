package hotkey

import (
	"strings"
	"testing"
)

func TestParseSpec_DefaultOverlay(t *testing.T) {
	spec := DefaultOverlay()
	mods, key, err := ParseSpec(spec)
	if err != nil {
		t.Fatalf("ParseSpec(%q): %v", spec, err)
	}
	if len(mods) < 1 {
		t.Fatalf("expected at least 1 modifier, got 0")
	}
	if key == 0 {
		t.Fatal("key is zero")
	}
}

func TestParseSpec_Empty(t *testing.T) {
	_, _, err := ParseSpec("")
	if err == nil {
		t.Fatal("empty spec should error")
	}
}

func TestParseSpec_Whitespace(t *testing.T) {
	_, _, err := ParseSpec("   ")
	if err == nil {
		t.Fatal("whitespace-only spec should error")
	}
}

func TestParseSpec_NoModifier(t *testing.T) {
	_, _, err := ParseSpec("Space")
	if err == nil {
		t.Fatal("spec without modifier should error")
	}
}

func TestParseSpec_UnknownModifier(t *testing.T) {
	_, _, err := ParseSpec("Hyper+K")
	if err == nil {
		t.Fatal("unknown modifier should error")
	}
	if !strings.Contains(err.Error(), "Hyper") {
		t.Fatalf("err = %v, want mention of Hyper", err)
	}
}

func TestParseSpec_UnknownKey(t *testing.T) {
	_, _, err := ParseSpec("Cmd+Mystery")
	if err == nil {
		t.Fatal("unknown key should error")
	}
}

func TestParseSpec_NamedKeys(t *testing.T) {
	for _, name := range []string{"Space", "Escape", "Tab", "Return", "Delete", "F1", "F12"} {
		_, k, err := ParseSpec("Cmd+" + name)
		if err != nil {
			t.Fatalf("Cmd+%s: %v", name, err)
		}
		if k == 0 {
			t.Fatalf("Cmd+%s: key is zero", name)
		}
	}
}

func TestParseSpec_Aliases(t *testing.T) {
	// "control" is an alias for "ctrl"; "option"/"opt" for "alt".
	_, k1, err := ParseSpec("Control+K")
	if err != nil {
		t.Fatal(err)
	}
	_, k2, err := ParseSpec("Ctrl+K")
	if err != nil {
		t.Fatal(err)
	}
	if k1 != k2 {
		t.Fatalf("Control vs Ctrl produced different keys: %d vs %d", k1, k2)
	}
}

func TestParseSpec_SinglePrintable(t *testing.T) {
	for _, c := range []string{"K", "k", "0", "9", "\\", "=", "-"} {
		_, _, err := ParseSpec("Cmd+Shift+" + c)
		if err != nil {
			t.Fatalf("Cmd+Shift+%s: %v", c, err)
		}
	}
}

func TestParseSpec_RejectsControlChar(t *testing.T) {
	_, _, err := ParseSpec("Cmd+\x01")
	if err == nil {
		t.Fatal("control-character key should error")
	}
}

func TestParseSpec_MultipleModifiers(t *testing.T) {
	mods, _, err := ParseSpec("Cmd+Shift+Alt+K")
	if err != nil {
		t.Fatal(err)
	}
	if len(mods) != 3 {
		t.Fatalf("expected 3 modifiers, got %d", len(mods))
	}
}

func TestNew_Defaults(t *testing.T) {
	m := New("Cmd+K")
	if m == nil {
		t.Fatal("New returned nil")
	}
	if m.spec != "Cmd+K" {
		t.Fatalf("spec = %q", m.spec)
	}
	if m.Started() {
		t.Fatal("should not be started before Start()")
	}
}

func TestStart_NilHandler(t *testing.T) {
	m := New("Cmd+K")
	if err := m.Start(nil); err == nil {
		t.Fatal("Start(nil) should error")
	}
}

func TestStart_BadSpec(t *testing.T) {
	m := New("not a real spec")
	if err := m.Start(func() {}); err == nil {
		t.Fatal("Start with bad spec should error")
	}
}

func TestStop_Idempotent(t *testing.T) {
	m := New("Cmd+K")
	m.Stop() // no-op
	m.Stop() // still no-op
}
