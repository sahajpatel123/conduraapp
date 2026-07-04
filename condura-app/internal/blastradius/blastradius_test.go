package blastradius

import "testing"

func TestClassify_ReadActions(t *testing.T) {
	reads := []string{
		"chat",
		"llm.complete",
		"transcribe",
		"speak",
		"tts",
		"screenshot.read",
		"ax.read",
		"clipboard.read",
		"file.read",
	}
	for _, kind := range reads {
		if got := Classify(Action{Kind: kind}); got != READ {
			t.Errorf("Classify(%q) = %v, want READ", kind, got)
		}
	}
}

func TestClassify_WriteActions(t *testing.T) {
	writes := []string{
		"file.write",
		"type",
		"paste",
		"clipboard.write",
		"click",
	}
	for _, kind := range writes {
		if got := Classify(Action{Kind: kind}); got != WRITE {
			t.Errorf("Classify(%q) = %v, want WRITE", kind, got)
		}
	}
}

func TestClassify_NetworkActions(t *testing.T) {
	networks := []string{
		"http.request",
		"form.submit",
		"message.send",
		"email.send",
		"click.link",
	}
	for _, kind := range networks {
		if got := Classify(Action{Kind: kind}); got != NETWORK {
			t.Errorf("Classify(%q) = %v, want NETWORK", kind, got)
		}
	}
}

func TestClassify_DestructiveActions(t *testing.T) {
	destructives := []string{
		"file.delete",
		"shell.exec",
		"purchase",
		"transfer",
		"format",
		"key.send",
	}
	for _, kind := range destructives {
		if got := Classify(Action{Kind: kind}); got != DESTRUCTIVE {
			t.Errorf("Classify(%q) = %v, want DESTRUCTIVE", kind, got)
		}
	}
}

// Unknown action kinds must classify as DESTRUCTIVE — the most
// conservative class — so the Gatekeeper's default is maximal caution.
func TestClassify_UnknownIsDestructive(t *testing.T) {
	if got := Classify(Action{Kind: "wibble.frobnicate"}); got != DESTRUCTIVE {
		t.Errorf("Classify(unknown) = %v, want DESTRUCTIVE", got)
	}
}

func TestClassify_EmptyIsDestructive(t *testing.T) {
	if got := Classify(Action{Kind: ""}); got != DESTRUCTIVE {
		t.Errorf("Classify(empty) = %v, want DESTRUCTIVE", got)
	}
}

// Kind matching is normalized: surrounding whitespace and case do not
// change the classification.
func TestClassify_NormalizesKind(t *testing.T) {
	if got := Classify(Action{Kind: "  CHAT  "}); got != READ {
		t.Errorf("Classify(messy chat) = %v, want READ", got)
	}
}

func TestClass_String(t *testing.T) {
	cases := map[Class]string{
		READ:        "READ",
		WRITE:       "WRITE",
		NETWORK:     "NETWORK",
		DESTRUCTIVE: "DESTRUCTIVE",
	}
	for c, want := range cases {
		if got := c.String(); got != want {
			t.Errorf("Class(%d).String() = %q, want %q", c, got, want)
		}
	}
}

// An out-of-range Class must render as DESTRUCTIVE — the safe default —
// rather than an empty or numeric string in audit logs.
func TestClass_String_OutOfRangeIsDestructive(t *testing.T) {
	if got := Class(99).String(); got != "DESTRUCTIVE" {
		t.Errorf("Class(99).String() = %q, want DESTRUCTIVE", got)
	}
}
