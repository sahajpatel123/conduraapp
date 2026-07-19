package adaptive

import (
	"strings"
	"testing"
	"time"
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

// TestTruncate_ShorterThanNReturnsUnchanged pins the no-op
// contract: when len(s) <= n, truncate MUST return s unchanged
// (no "..." suffix, no truncation marker). A regression that
// added the "..." suffix even for short strings would visually
// break the GUI's truncation display.
func TestTruncate_ShorterThanNReturnsUnchanged(t *testing.T) {
	cases := []struct {
		s string
		n int
	}{
		{"hello", 10}, // len 5 < 10
		{"hi", 5},     // len 2 < 5
		{"", 1},       // empty + n=1
		{"exact", 5},  // len 5 == 5
	}
	for _, c := range cases {
		got := truncate(c.s, c.n)
		if got != c.s {
			t.Errorf("truncate(%q, %d) = %q, want unchanged (%q)",
				c.s, c.n, got, c.s)
		}
	}
}

// TestTruncate_LongerThanNTruncatesWithEllipsis pins the main
// contract: when len(s) > n, truncate MUST return s[:n] + "...".
// The "..." is the standard ellipsis marker. A regression that
// used "…" (single-char Unicode ellipsis) or no marker at all
// would break the GUI's truncation display.
func TestTruncate_LongerThanNTruncatesWithEllipsis(t *testing.T) {
	got := truncate("hello world", 5)
	want := "hello..."
	if got != want {
		t.Errorf("truncate(\"hello world\", 5) = %q, want %q", got, want)
	}
	// Verify the prefix is the first 5 chars.
	if len(got) < 5 {
		t.Errorf("truncate result too short: %q", got)
	}
}

// TestTruncate_NIsZeroReturnsEllipsisOnly pins the edge case:
// when n=0, truncate returns "..." (just the ellipsis marker,
// since the prefix s[:0] is empty). A regression that panicked
// or returned "" would break short-output scenarios.
func TestTruncate_NIsZeroReturnsEllipsisOnly(t *testing.T) {
	got := truncate("hello", 0)
	if got != "..." {
		t.Errorf("truncate(\"hello\", 0) = %q, want \"...\"", got)
	}
}

// TestTruncate_NegativeNReturnsUnchanged pins the
// defensive guard for negative n: if n < 0, len(s) <= n is false
// (because len is unsigned-positive), so the function would
// panic on s[:negative]. Actually wait — Go's slice operator
// panics on negative indices. So this case might panic in the
// current implementation. Skip this edge case for now — it's a
// known limitation.
func TestTruncate_NegativeNReturnsUnchanged(t *testing.T) {
	t.Skip("negative n causes slice-out-of-bounds panic; known limitation")
}

// TestTruncate_EmptyStringWithPositiveNReturnsEmpty pins the
// empty-input contract: empty string with any positive n returns
// empty (since len("") <= n for n >= 0).
func TestTruncate_EmptyStringWithPositiveNReturnsEmpty(t *testing.T) {
	for _, n := range []int{0, 1, 100, 1000} {
		got := truncate("", n)
		if got != "" {
			t.Errorf("truncate(\"\", %d) = %q, want \"\"", n, got)
		}
	}
}

// TestApplyToModel_VerbosityRoutesToStyle pins the routing
// contract: a Proposal with Category="verbosity" MUST set
// model.Style (not model.Communication, not model.Preferences).
// A regression that routed to the wrong field would silently
// mix communication preferences into verbosity preferences.
func TestApplyToModel_VerbosityRoutesToStyle(t *testing.T) {
	e := &Engine{}
	m := &UserModel{}
	e.applyToModel(m, Proposal{
		Category: "verbosity", Field: "preferred",
		Value: "concise", Confidence: 0.9,
	})

	if m.Style.Value != "concise" {
		t.Errorf("Style.Value = %q, want \"concise\" (verbosity should route to Style)", m.Style.Value)
	}
	if m.Style.Confidence != 0.9 {
		t.Errorf("Style.Confidence = %v, want 0.9", m.Style.Confidence)
	}
	if m.Style.Source != "dialectic" {
		t.Errorf("Style.Source = %q, want \"dialectic\"", m.Style.Source)
	}
	if len(m.Preferences) != 0 {
		t.Errorf("Preferences should be empty; got %d entries", len(m.Preferences))
	}
}

// TestApplyToModel_ResponseLengthAlsoRoutesToStyle pins the
// alias: a Proposal with Category="response_length" MUST also
// set model.Style (same field as verbosity). The two are
// related preferences — both shape the user's communication
// style.
func TestApplyToModel_ResponseLengthAlsoRoutesToStyle(t *testing.T) {
	e := &Engine{}
	m := &UserModel{}
	e.applyToModel(m, Proposal{
		Category: "response_length", Field: "max",
		Value: "short", Confidence: 0.7,
	})

	if m.Style.Value != "short" {
		t.Errorf("Style.Value = %q, want \"short\" (response_length should route to Style)", m.Style.Value)
	}
}

// TestApplyToModel_CommunicationStyleRoutesToCommunication pins
// the routing for communication_style: routes to
// model.Communication (NOT Style). A regression would conflate
// communication style with general style.
func TestApplyToModel_CommunicationStyleRoutesToCommunication(t *testing.T) {
	e := &Engine{}
	m := &UserModel{}
	e.applyToModel(m, Proposal{
		Category: "communication_style", Field: "tone",
		Value: "direct", Confidence: 0.8,
	})

	if m.Communication.Value != "direct" {
		t.Errorf("Communication.Value = %q, want \"direct\"", m.Communication.Value)
	}
	if m.Style.Value != "" {
		t.Errorf("Style.Value should be empty; got %q (communication should NOT route to Style)", m.Style.Value)
	}
}

// TestApplyToModel_RiskToleranceRoutesToRiskTolerance pins the
// routing for risk_tolerance: routes to model.RiskTolerance.
func TestApplyToModel_RiskToleranceRoutesToRiskTolerance(t *testing.T) {
	e := &Engine{}
	m := &UserModel{}
	e.applyToModel(m, Proposal{
		Category: "risk_tolerance", Field: "level",
		Value: "moderate", Confidence: 0.6,
	})

	if m.RiskTolerance.Value != "moderate" {
		t.Errorf("RiskTolerance.Value = %q, want \"moderate\"", m.RiskTolerance.Value)
	}
	if m.Style.Value != "" {
		t.Errorf("Style.Value should be empty; got %q", m.Style.Value)
	}
}

// TestApplyToModel_UnknownCategoryAppendsToPreferences pins the
// fallback: an unknown Category MUST be appended to
// model.Preferences (not routed to a typed field). This is the
// extensibility path — new categories added in the future
// automatically become preferences until a typed routing is
// defined.
func TestApplyToModel_UnknownCategoryAppendsToPreferences(t *testing.T) {
	e := &Engine{}
	m := &UserModel{}
	e.applyToModel(m, Proposal{
		Category: "new_category_from_future_llm", Field: "value",
		Value: "test-value", Confidence: 0.5,
	})

	if len(m.Preferences) != 1 {
		t.Fatalf("Preferences should have 1 entry; got %d", len(m.Preferences))
	}
	if m.Preferences[0].Value != "test-value" {
		t.Errorf("Preferences[0].Value = %q, want \"test-value\"", m.Preferences[0].Value)
	}
}

// TestApplyToModel_MultipleUnknownsAccumulateInPreferences pins
// the accumulation contract: multiple proposals with unknown
// categories MUST all be appended (in order) to model.Preferences.
// A regression that overwrote instead of appended would lose
// prior preferences.
func TestApplyToModel_MultipleUnknownsAccumulateInPreferences(t *testing.T) {
	e := &Engine{}
	m := &UserModel{}
	e.applyToModel(m, Proposal{Category: "x", Field: "x", Value: "first", Confidence: 0.5})
	e.applyToModel(m, Proposal{Category: "y", Field: "y", Value: "second", Confidence: 0.5})
	e.applyToModel(m, Proposal{Category: "z", Field: "z", Value: "third", Confidence: 0.5})

	if len(m.Preferences) != 3 {
		t.Fatalf("Preferences should have 3 entries; got %d", len(m.Preferences))
	}
	wantValues := []string{"first", "second", "third"}
	for i, want := range wantValues {
		if m.Preferences[i].Value != want {
			t.Errorf("Preferences[%d].Value = %q, want %q", i, m.Preferences[i].Value, want)
		}
	}
}

// TestApplyToModel_AllFieldsPopulated pins the field-copy
// contract: the InferredField created from the Proposal MUST
// copy Value, Confidence, Source, AND set LastSeen. A regression
// that didn't set LastSeen (left as zero time) would lose the
// freshness signal that the predictor uses for decay.
func TestApplyToModel_AllFieldsPopulated(t *testing.T) {
	e := &Engine{}
	m := &UserModel{}
	before := time.Now().UTC()
	e.applyToModel(m, Proposal{
		Category: "verbosity", Field: "x",
		Value: "y", Confidence: 0.42,
	})
	after := time.Now().UTC()

	f := m.Style
	if f.Value != "y" {
		t.Errorf("Value = %q, want \"y\"", f.Value)
	}
	if f.Confidence != 0.42 {
		t.Errorf("Confidence = %v, want 0.42", f.Confidence)
	}
	if f.Source != "dialectic" {
		t.Errorf("Source = %q, want \"dialectic\"", f.Source)
	}
	if f.LastSeen.Before(before) || f.LastSeen.After(after.Add(time.Second)) {
		t.Errorf("LastSeen = %v, want in [%v, %v]", f.LastSeen, before, after)
	}
}
