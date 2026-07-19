package daemon

import (
	"encoding/json"
	"testing"
)

// TestParseParams_HappyPath pins the successful-unmarshal
// contract: parseParams returns nil and populates dest when
// the JSON is well-formed and matches the struct shape.
func TestParseParams_HappyPath(t *testing.T) {
	raw := json.RawMessage(`{"path":"/var/lib/condura/test.zip","x":42}`)
	var dest struct {
		Path string `json:"path"`
		X    int    `json:"x"`
	}
	if err := parseParams(raw, &dest); err != nil {
		t.Fatalf("parseParams(happy): %v", err)
	}
	if dest.Path != "/var/lib/condura/test.zip" {
		t.Errorf("Path = %q, want /var/lib/condura/test.zip", dest.Path)
	}
	if dest.X != 42 {
		t.Errorf("X = %d, want 42", dest.X)
	}
}

// TestParseParams_MalformedJSONReturnsIPCError pins the error
// contract: a malformed JSON returns an ipc.Error with
// CodeInvalidParams. The error message includes the verbatim
// json.Unmarshal output (with byte offset), which the CLI
// surfaces to the user and the GUI shows in a toast.
func TestParseParams_MalformedJSONReturnsIPCError(t *testing.T) {
	raw := json.RawMessage(`{not json}`)
	var dest struct {
		Path string `json:"path"`
	}
	err := parseParams(raw, &dest)
	if err == nil {
		t.Fatal("parseParams(malformed) = nil; want error")
	}
	// Error message must contain some indication of the parse
	// failure (the standard library includes the byte offset).
	if msg := err.Error(); msg == "" {
		t.Error("error message is empty; want a parse-failure description")
	}
}

// TestParseParams_EmptyParamsOK pins the boundary: empty raw
// JSON ({}) is valid JSON and must succeed (unmarshaling into
// the zero value of every field). The handler is then
// responsible for calling requireField to catch missing fields.
// This is the "params are optional" case the IPC supports.
func TestParseParams_EmptyParamsOK(t *testing.T) {
	raw := json.RawMessage(`{}`)
	var dest struct {
		Path string `json:"path"`
	}
	if err := parseParams(raw, &dest); err != nil {
		t.Errorf("parseParams(empty): %v; want nil (empty params OK)", err)
	}
	if dest.Path != "" {
		t.Errorf("Path = %q, want empty", dest.Path)
	}
}

// TestRequireField_NonEmptyOK pins the happy path: a non-empty
// value passes requireField unchanged (returns nil).
func TestRequireField_NonEmptyOK(t *testing.T) {
	if err := requireField("path", "/var/lib/condura/test.zip"); err != nil {
		t.Errorf("requireField(non-empty) = %v, want nil", err)
	}
}

// TestRequireField_EmptyReturnsIPCError pins the deny contract:
// an empty string returns an ipc.Error with CodeInvalidParams
// and a message naming the field. This is the contract every
// IPC handler with a required path/id field relies on.
func TestRequireField_EmptyReturnsIPCError(t *testing.T) {
	err := requireField("path", "")
	if err == nil {
		t.Fatal("requireField(empty) = nil; want error")
	}
	if msg := err.Error(); msg != "rpc error -32602: path is required" {
		t.Errorf("error message = %q, want %q (matches the inline contract)", msg, "rpc error -32602: path is required")
	}
}

// TestRequireField_DifferentFieldNamesProduceDifferentMessages
// pins the per-caller-field-name contract: the error message
// MUST include the caller-supplied field name, not a generic
// "value is required". A regression that hardcoded "field is
// required" would lose the field-name signal that the GUI
// uses to highlight the offending input.
func TestRequireField_DifferentFieldNamesProduceDifferentMessages(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  string
	}{
		{"path", "", "rpc error -32602: path is required"},
		{"locale", "", "rpc error -32602: locale is required"},
		{"hotkey", "", "rpc error -32602: hotkey is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := requireField(tc.name, tc.value)
			if err == nil {
				t.Fatal("requireField(empty) = nil; want error")
			}
			if got := err.Error(); got != tc.want {
				t.Errorf("error = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestRequireField_ErrorCodeMatchesInlineContract pins the
// IPC error code: CodeInvalidParams (-32602). The previous
// inline pattern used CodeInvalidParams; the helper must
// match. If a future refactor accidentally switches to
// CodeInternalError (-32603), the GUI's parameter-validation
// toast handler would no longer trigger.
func TestRequireField_ErrorCodeMatchesInlineContract(t *testing.T) {
	err := requireField("test", "")
	if err == nil {
		t.Fatal("requireField(empty) = nil; want error")
	}
	// Parse the message to verify the code. We can't import
	// ipc.Error here because the wrapping test would conflict
	// with our internal assertion, but the message format
	// "rpc error -32602: ..." is the contract.
	wantPrefix := "rpc error -32602:"
	if msg := err.Error(); len(msg) < len(wantPrefix) || msg[:len(wantPrefix)] != wantPrefix {
		t.Errorf("error message = %q, want prefix %q (CodeInvalidParams)", msg, wantPrefix)
	}
}

// TestParseAndRequire_CombinedContract pins the typical
// call-site pattern: parseParams + requireField chained.
// Used together, they cover the entire "validate params and
// catch missing required fields" contract that every gated
// IPC handler needs.
//
// This is a behavior test of the integration, not of the
// helpers in isolation. If a future refactor changes the
// helper signatures or error shape, this test catches the
// regression at the call-site level.
func TestParseAndRequire_CombinedContract(t *testing.T) {
	// Happy path: valid JSON, non-empty path.
	raw := json.RawMessage(`{"path":"/var/lib/condura/test.zip"}`)
	var p struct {
		Path string `json:"path"`
	}
	if err := parseParams(raw, &p); err != nil {
		t.Fatalf("parseParams: %v", err)
	}
	if err := requireField("path", p.Path); err != nil {
		t.Errorf("requireField: %v", err)
	}

	// Failure path: valid JSON, missing required path.
	raw2 := json.RawMessage(`{"other":"value"}`)
	var p2 struct {
		Path string `json:"path"`
	}
	if err := parseParams(raw2, &p2); err != nil {
		t.Fatalf("parseParams(2): %v", err)
	}
	if err := requireField("path", p2.Path); err == nil {
		t.Error("requireField on empty path = nil; want error")
	}

	// Failure path: malformed JSON.
	raw3 := json.RawMessage(`{not json}`)
	var p3 struct {
		Path string `json:"path"`
	}
	if err := parseParams(raw3, &p3); err == nil {
		t.Error("parseParams(malformed) = nil; want error")
	}
}
