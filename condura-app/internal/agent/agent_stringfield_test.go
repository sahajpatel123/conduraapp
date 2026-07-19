package agent

import (
	"testing"
)

// TestStringField_HappyPath pins the basic contract: when data
// is a JSON-marshalable object containing key="value", stringField
// returns ("value", true). This is the path the agent's tool-call
// pipeline uses to extract string arguments from incoming JSON.
func TestStringField_HappyPath(t *testing.T) {
	data := map[string]any{"name": "alice", "age": 30}
	v, ok := stringField(data, "name")
	if !ok {
		t.Fatal("stringField returned ok=false on existing key; want true")
	}
	if v != "alice" {
		t.Errorf("value = %q, want \"alice\"", v)
	}
}

// TestStringField_NilDataReturnsFalse pins the nil-input contract:
// stringField(nil, ...) MUST return ("", false). A regression that
// panicked on the json.Marshal(nil) call would crash the agent's
// argument-extraction pipeline.
func TestStringField_NilDataReturnsFalse(t *testing.T) {
	v, ok := stringField(nil, "key")
	if ok {
		t.Error("stringField(nil) returned ok=true; want false")
	}
	if v != "" {
		t.Errorf("value = %q, want \"\"", v)
	}
}

// TestStringField_WrongTypeReturnsFalse pins the type-discrimination
// contract: a key whose value is NOT a string MUST return ("", false).
// A regression that returned the non-string value (e.g., a number
// formatted as string) would silently corrupt downstream string
// operations (concatenation, length, etc.).
func TestStringField_WrongTypeReturnsFalse(t *testing.T) {
	data := map[string]any{"count": 42, "active": true, "items": []any{1, 2}}
	for _, key := range []string{"count", "active", "items"} {
		v, ok := stringField(data, key)
		if ok {
			t.Errorf("stringField(%s) returned ok=true for non-string value; want false", key)
		}
		if v != "" {
			t.Errorf("stringField(%s) returned non-empty value %q; want \"\"", key, v)
		}
	}
}

// TestStringField_MissingKeyReturnsFalse pins the missing-key contract:
// a key that doesn't exist in the data MUST return ("", false).
// A regression that returned an empty string with ok=true would
// mislead callers into thinking the empty string was an intentional
// value.
func TestStringField_MissingKeyReturnsFalse(t *testing.T) {
	data := map[string]any{"name": "alice"}
	v, ok := stringField(data, "missing")
	if ok {
		t.Error("stringField on missing key returned ok=true; want false")
	}
	if v != "" {
		t.Errorf("value = %q, want \"\"", v)
	}
}

// TestStringField_EmptyDataReturnsFalse pins the empty-map contract:
// a non-nil but empty data MUST return ("", false) for any key.
// A regression that returned empty + ok=true would silently pass
// empty strings to the agent.
func TestStringField_EmptyDataReturnsFalse(t *testing.T) {
	v, ok := stringField(map[string]any{}, "anykey")
	if ok {
		t.Error("stringField on empty map returned ok=true; want false")
	}
	if v != "" {
		t.Errorf("value = %q, want \"\"", v)
	}
}

// TestStringField_NonMarshalableDataReturnsFalse pins the defensive
// guard: data that json.Marshal cannot handle (e.g., a function
// value, or a chan) MUST return ("", false) rather than panicking.
// json.Marshal returns an error for these types; the function
// MUST catch that error and return ok=false.
func TestStringField_NonMarshalableDataReturnsFalse(t *testing.T) {
	// A chan is not JSON-serializable. json.Marshal returns an error.
	data := make(chan int)
	v, ok := stringField(data, "anykey")
	if ok {
		t.Error("stringField on chan returned ok=true; want false")
	}
	if v != "" {
		t.Errorf("value = %q, want \"\"", v)
	}
}

// TestStringField_StringValueExtractedCorrectly pins the string-
// extraction path: when fields[key] IS a string, it MUST be
// returned verbatim (no transformation). A regression that
// applied trim/lower/etc would silently corrupt the value.
func TestStringField_StringValueExtractedCorrectly(t *testing.T) {
	cases := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"simple", "hello"},
		{"with-spaces", "hello world"},
		{"with-unicode", "héllo wörld"},
		{"with-special-chars", "<script>alert(1)</script>"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]any{"key": tc.value}
			v, ok := stringField(data, "key")
			if !ok {
				t.Fatalf("stringField returned ok=false for %q", tc.value)
			}
			if v != tc.value {
				t.Errorf("value = %q, want %q (no transformation)", v, tc.value)
			}
		})
	}
}
