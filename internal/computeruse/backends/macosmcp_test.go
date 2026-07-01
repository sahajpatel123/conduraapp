package backends

import (
	"context"
	"strings"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/computeruse"
)

type fakeMCP struct {
	available  bool
	screenshot *computeruse.Screenshot
	axTree     *computeruse.AXTree
	execResult *computeruse.ActionResult
}

func (f *fakeMCP) name() string                                    { return "macos-mcp-test" }
func (f *fakeMCP) isAvailable() bool                               { return f.available }
func (f *fakeMCP) captureScreen() (*computeruse.Screenshot, error) { return f.screenshot, nil }
func (f *fakeMCP) getAXTree() (*computeruse.AXTree, error)         { return f.axTree, nil }
func (f *fakeMCP) execute(a *computeruse.Action) (*computeruse.ActionResult, error) {
	if f.execResult != nil {
		f.execResult.Action = a
		return f.execResult, nil
	}
	return &computeruse.ActionResult{Success: true}, nil
}

func TestMCP_ImplementsBackend(t *testing.T) {
	var _ computeruse.Backend = (*MacOSMCPBackend)(nil)
}

func TestMCP_Capabilities(t *testing.T) {
	b := &MacOSMCPBackend{impl: &fakeMCP{}}
	if len(b.Capabilities()) != 9 {
		t.Errorf("got %d capabilities, want 9", len(b.Capabilities()))
	}
}

func TestMCP_IsAvailable(t *testing.T) {
	b := &MacOSMCPBackend{impl: &fakeMCP{available: true}}
	if !b.IsAvailable(context.Background()) {
		t.Error("expected available")
	}
}

func TestMCP_Execute(t *testing.T) {
	b := &MacOSMCPBackend{impl: &fakeMCP{}}
	r, err := b.Execute(context.Background(), &computeruse.Action{Type: computeruse.ActionClick})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !r.Success {
		t.Error("expected success")
	}
}

func TestMCP_GetAXTree(t *testing.T) {
	want := &computeruse.AXTree{PID: 1}
	b := &MacOSMCPBackend{impl: &fakeMCP{axTree: want}}
	got, err := b.GetAXTree(context.Background())
	if err != nil {
		t.Fatalf("GetAXTree: %v", err)
	}
	if got.PID != 1 {
		t.Errorf("PID = %d", got.PID)
	}
}

// Audit 2026-07-01: AppleScript escaper hardening. The previous
// escaper handled backslash, double-quote, and backtick. It missed
// the ampersand (AppleScript string-concat operator — lets a model-
// controlled value splice in `& do shell script "..."`) and CR/LF/TAB
// (which AppleScript treats as expression terminators inside double-
// quoted strings, so a value with a newline could inject a sibling
// statement). These tests pin the new behavior.
func TestEscapeAppleScript_Ampersand(t *testing.T) {
	got := escapeAppleScript(`cmd & sudo bad`)
	want := `cmd \& sudo bad`
	if got != want {
		t.Errorf("escapeAppleScript(ampersand) = %q; want %q", got, want)
	}
	if !strings.Contains(got, `\&`) {
		t.Error("ampersand must be backslash-escaped to prevent AppleScript concat injection")
	}
}

func TestEscapeAppleScript_Newline(t *testing.T) {
	got := escapeAppleScript("cmd\nrm -rf /")
	want := `cmd\nrm -rf /`
	if got != want {
		t.Errorf("escapeAppleScript(newline) = %q; want %q", got, want)
	}
	// A literal newline in the output would let the model splice in a
	// second AppleScript statement.
	if strings.ContainsRune(got, '\n') {
		t.Error("output must not contain a raw newline — AppleScript treats it as a statement separator")
	}
}

func TestEscapeAppleScript_CarriageReturn(t *testing.T) {
	got := escapeAppleScript("cmd\rrm")
	if strings.ContainsRune(got, '\r') {
		t.Error("output must not contain a raw CR")
	}
	if !strings.Contains(got, `\r`) {
		t.Errorf("expected \\r escape sequence in output, got %q", got)
	}
}

func TestEscapeAppleScript_Tab(t *testing.T) {
	got := escapeAppleScript("col1\tcol2")
	if strings.ContainsRune(got, '\t') {
		t.Error("output must not contain a raw TAB")
	}
	if !strings.Contains(got, `\t`) {
		t.Errorf("expected \\t escape sequence in output, got %q", got)
	}
}

func TestEscapeAppleScript_AllKnownChars(t *testing.T) {
	// Combined: all the chars a model might inject. The security
	// property we want is: when this output is embedded inside an
	// AppleScript double-quoted string literal, NONE of the seven
	// characters below act as a syntactic boundary. Backslashes
	// appear inside the output as the first half of escape pairs
	// (\\, \", \`, \&, \n, \r, \t) — those are NOT unescaped
	// instances of `\`, they are escape characters in the literal.
	// So we check the raw control/quote characters only.
	in := "a\\b\"c`d&e\nf\rg\th"
	out := escapeAppleScript(in)

	// Each escape sequence below must be present.
	mustContain := []string{`\"`, `\&`, `\\`, "\\`", `\n`, `\r`, `\t`}
	for _, want := range mustContain {
		if !strings.Contains(out, want) {
			t.Errorf("output %q must contain escape %q", out, want)
		}
	}

	// The output must be byte-for-byte safe to drop inside an
	// AppleScript double-quoted string. The simplest end-to-end
	// check: an AppleScript-style expression with the output
	// interpolated must still parse as a single string literal.
	// Concretely, the number of unescaped backslashes must be
	// even (each "open" backslash is closed by the next char),
	// and the output must end with a state where the next " would
	// be inside the literal, not closing it. We approximate that
	// by counting backslashes not followed by an escape-pair char.
	unescaped := false
	for i := 0; i < len(out); i++ {
		if out[i] != '\\' {
			continue
		}
		// Is this the start of a known escape pair?
		if i+1 < len(out) {
			switch out[i+1] {
			case '\\', '"', '`', '&', 'n', 'r', 't':
				i++ // skip the escape char
				continue
			}
		}
		unescaped = true
		break
	}
	if unescaped {
		t.Errorf("output %q contains a backslash that is not the start of a known escape pair", out)
	}
}

func TestEscapeAppleScript_SafeStringUnchanged(t *testing.T) {
	in := "Hello, world"
	if got := escapeAppleScript(in); got != in {
		t.Errorf("safe string mutated: %q -> %q", in, got)
	}
}
