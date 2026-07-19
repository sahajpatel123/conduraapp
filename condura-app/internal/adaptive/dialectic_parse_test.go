package adaptive

import (
	"strings"
	"testing"
)

// TestParseProposals_EmptyInputReturnsNil pins the empty-input
// contract: parseProposals("") MUST return nil. A regression that
// returned an empty (non-nil) slice would cause downstream code
// to iterate over zero items — semantically equivalent to nil but
// could trigger subtle bugs in code that uses `len(proposals) > 0`
// vs `proposals != nil` checks differently.
func TestParseProposals_EmptyInputReturnsNil(t *testing.T) {
	got := parseProposals("")
	if got != nil {
		t.Errorf("parseProposals(\"\") = %v, want nil", got)
	}
}

// TestParseProposals_ValidJSONArrayReturnsProposals pins the
// happy-path contract: a well-formed JSON array MUST be
// unmarshaled into a []Proposal.
func TestParseProposals_ValidJSONArrayReturnsProposals(t *testing.T) {
	raw := `[{"category":"verbosity","field":"style","value":"casual","confidence":0.9,"reason":"user uses casual language"}]`
	got := parseProposals(raw)
	if len(got) != 1 {
		t.Fatalf("parseProposals returned %d proposals; want 1", len(got))
	}
	p := got[0]
	if p.Category != "verbosity" {
		t.Errorf("Category = %q, want \"verbosity\"", p.Category)
	}
	if p.Field != "style" {
		t.Errorf("Field = %q, want \"style\"", p.Field)
	}
	if p.Value != "casual" {
		t.Errorf("Value = %q, want \"casual\"", p.Value)
	}
	if p.Confidence != 0.9 {
		t.Errorf("Confidence = %v, want 0.9", p.Confidence)
	}
}

// TestParseProposals_InvalidJSONReturnsNil pins the defensive
// guard: a malformed JSON input MUST return nil (NOT a partial
// proposal list, NOT a panic). The LLM may return malformed JSON;
// the parser MUST fail safely.
func TestParseProposals_InvalidJSONReturnsNil(t *testing.T) {
	cases := []string{
		"not json at all",
		"[{ unclosed bracket",
		`[{"category": "x"`, // truncated
		`[{"category": }]`,  // missing value
	}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			got := parseProposals(raw)
			if got != nil {
				t.Errorf("parseProposals(%q) = %v, want nil", raw, got)
			}
		})
	}
}

// TestParseProposals_MissingConfidenceDefaultsToHalf pins the
// default-confidence contract: proposals with confidence <= 0
// MUST be defaulted to 0.5. A regression that left confidence
// at 0 would cause the adjudicator's threshold check
// (0.6 by default) to silently reject every unannotated
// proposal.
func TestParseProposals_MissingConfidenceDefaultsToHalf(t *testing.T) {
	raw := `[{"category":"verbosity","field":"style","value":"casual","confidence":0}]`
	got := parseProposals(raw)
	if len(got) != 1 {
		t.Fatalf("parseProposals returned %d proposals; want 1", len(got))
	}
	if got[0].Confidence != 0.5 {
		t.Errorf("Confidence = %v, want 0.5 (default for <= 0)", got[0].Confidence)
	}
}

// TestParseProposals_NegativeConfidenceDefaultsToHalf pins the
// negative-confidence edge case: proposals with confidence < 0
// MUST also be defaulted to 0.5 (the guard is `<= 0`, not
// `== 0`).
func TestParseProposals_NegativeConfidenceDefaultsToHalf(t *testing.T) {
	raw := `[{"category":"verbosity","field":"style","value":"casual","confidence":-1.5}]`
	got := parseProposals(raw)
	if len(got) != 1 {
		t.Fatalf("parseProposals returned %d proposals; want 1", len(got))
	}
	if got[0].Confidence != 0.5 {
		t.Errorf("Confidence = %v, want 0.5 (default for negative)", got[0].Confidence)
	}
}

// TestParseProposals_MultipleProposalsAllReturned pins the
// multi-proposal contract: an array with N entries MUST return
// N proposals. A regression that dropped entries would lose
// inferences.
func TestParseProposals_MultipleProposalsAllReturned(t *testing.T) {
	raw := `[
		{"category":"verbosity","field":"style","value":"casual","confidence":0.9},
		{"category":"response_length","field":"max","value":"short","confidence":0.7},
		{"category":"communication_style","field":"tone","value":"direct","confidence":0.8}
	]`
	got := parseProposals(raw)
	if len(got) != 3 {
		t.Errorf("parseProposals returned %d proposals; want 3", len(got))
	}
	categories := map[string]bool{}
	for _, p := range got {
		categories[p.Category] = true
	}
	for _, want := range []string{"verbosity", "response_length", "communication_style"} {
		if !categories[want] {
			t.Errorf("category %q missing from parsed proposals", want)
		}
	}
}

// TestParseProposals_ExtractsJSONBlock pins the
// "extract JSON from messy LLM output" contract: the input may
// have markdown code fences or surrounding text, and parseProposals
// MUST extract the JSON block before unmarshaling. A regression
// that required exact JSON would silently drop every proposal
// because the LLM adds ```json ... ``` fences.
func TestParseProposals_ExtractsJSONBlock(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want int // expected number of proposals
	}{
		{"markdown-fence-json", "```json\n[{\"category\":\"verbosity\",\"field\":\"style\",\"value\":\"casual\",\"confidence\":0.9}]\n```", 1},
		{"markdown-fence", "```\n[{\"category\":\"verbosity\",\"field\":\"style\",\"value\":\"casual\",\"confidence\":0.9}]\n```", 1},
		{"leading-text", "Here's my analysis:\n[{\"category\":\"verbosity\",\"field\":\"style\",\"value\":\"casual\",\"confidence\":0.9}]", 1},
		{"trailing-text", "[{\"category\":\"verbosity\",\"field\":\"style\",\"value\":\"casual\",\"confidence\":0.9}]\nThat's my analysis.", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseProposals(tc.raw)
			if len(got) != tc.want {
				t.Errorf("parseProposals with %s returned %d proposals; want %d (extraction failed?)",
					tc.name, len(got), tc.want)
			}
		})
	}
}

// TestExtractJSONBlock_PureJSONUnchanged pins the
// pass-through contract: when the input is already pure JSON
// (no fences, no surrounding text), extractJSONBlock MUST return
// it unchanged. This is the "happy path" of the extraction.
func TestExtractJSONBlock_PureJSONUnchanged(t *testing.T) {
	raw := `[{"category":"verbosity","field":"style","value":"casual","confidence":0.9}]`
	got := extractJSONBlock(raw)
	if got != raw {
		t.Errorf("extractJSONBlock(pure JSON) = %q, want unchanged", got)
	}
}

// TestExtractJSONBlock_StripsMarkdownFences pins the
// fence-stripping contract: ```json and ``` fences MUST be
// removed. Real LLM outputs (per the proposerSysPrompt:
// "Return ONLY valid JSON array") often include fences anyway.
func TestExtractJSONBlock_StripsMarkdownFences(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"json-fence", "```json\n[{\"x\":1}]\n```", "[{\"x\":1}]"},
		{"plain-fence", "```\n[{\"x\":1}]\n```", "[{\"x\":1}]"},
		{"fence-with-language-hint", "```JSON\n[{\"x\":1}]\n```", "[{\"x\":1}]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractJSONBlock(tc.raw)
			if got != tc.want {
				t.Errorf("extractJSONBlock(%s) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

// TestExtractJSONBlock_StripsSurroundingText pins the
// surrounding-text-stripping contract: text before/after the JSON
// block MUST be ignored. The extraction looks for the first `[` or
// `{` and the matching closing bracket.
func TestExtractJSONBlock_StripsSurroundingText(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"leading-text", "Some preamble\n[{\"x\":1}]", "[{\"x\":1}]"},
		{"trailing-text", "[{\"x\":1}]\nSome postamble", "[{\"x\":1}]"},
		{"both", "Preamble [{\"x\":1}] postamble", "[{\"x\":1}]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractJSONBlock(tc.raw)
			if !strings.Contains(got, tc.want) {
				t.Errorf("extractJSONBlock(%s) = %q, want it to contain %q",
					tc.name, got, tc.want)
			}
		})
	}
}
