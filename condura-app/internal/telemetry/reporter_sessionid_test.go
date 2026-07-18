package telemetry

import (
	"strings"
	"testing"
)

// TestSessionIDPrefix_EmptyReturnsZero pins the empty-input
// contract: sessionIDPrefix("") MUST return 0. A regression that
// panicked on empty input would break the counter-grouping path
// at startup (before any session ID is set).
func TestSessionIDPrefix_EmptyReturnsZero(t *testing.T) {
	if got := sessionIDPrefix(""); got != 0 {
		t.Errorf("sessionIDPrefix(\"\") = %d, want 0", got)
	}
}

// TestSessionIDPrefix_TooShortReturnsZero pins the
// short-input contract: sessionIDPrefix with id length < 4
// MUST return 0 (early return). The function only consumes
// the first 4 hex chars (= 4 bytes = 8 hex chars), so any
// shorter input can't be processed.
func TestSessionIDPrefix_TooShortReturnsZero(t *testing.T) {
	cases := []string{"a", "ab", "abc"} // lengths 1, 2, 3
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			if got := sessionIDPrefix(c); got != 0 {
				t.Errorf("sessionIDPrefix(%q) = %d, want 0 (early return)", c, got)
			}
		})
	}
}

// TestSessionIDPrefix_ValidHexReturnsInt pins the happy-path
// contract: sessionIDPrefix with a valid 8+ char hex string MUST
// return the int formed by the first 4 bytes (8 hex chars).
// This is the privacy-preserving grouping key used in counters.
//
// NOTE: int width is platform-dependent (32-bit on some systems,
// 64-bit on most modern systems including darwin dev). The
// production code does `int(v[0])<<24 | int(v[1])<<16 | int(v[2])<<8 | int(v[3])`
// which produces a value that fits in 32 bits but is stored in
// an int. We assert on the lower 32 bits via uint32 comparison
// to avoid platform-dependent signedness.
func TestSessionIDPrefix_ValidHexReturnsInt(t *testing.T) {
	cases := []struct {
		in      string
		wantU32 uint32
	}{
		{"00000000", 0x00000000},
		{"ffffffff", 0xFFFFFFFF},
		{"deadbeef", 0xDEADBEEF},
		{"00000001", 0x00000001},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := sessionIDPrefix(c.in)
			if uint32(got) != c.wantU32 {
				t.Errorf("sessionIDPrefix(%q) = %d (0x%X), want 0x%X",
					c.in, got, uint32(got), c.wantU32)
			}
		})
	}
}

// TestSessionIDPrefix_InvalidHexReturnsZero pins the
// input-validation contract: sessionIDPrefix with a non-hex
// string (8 chars but not valid hex) MUST return 0 (NOT panic).
// A regression that propagated the decode error would let
// invalid session IDs corrupt the counter-grouping buckets.
func TestSessionIDPrefix_InvalidHexReturnsZero(t *testing.T) {
	cases := []string{
		"zzzzzzzz",   // all 'z' is not hex
		"1234567g",   // 8 chars but 'g' is not hex (g > f)
		"xxxxxxxx",   // 8 chars but 'x' is not hex
		"          ", // spaces
		"!@#$%^&*",   // 8 punctuation chars
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			if got := sessionIDPrefix(c); got != 0 {
				t.Errorf("sessionIDPrefix(%q) = %d, want 0 (decode error -> early return)", c, got)
			}
		})
	}
}

// TestSessionIDPrefix_LongerThanEightUsesPrefix pins the
// "use only the first 8 chars" contract: sessionIDPrefix with
// a longer hex string MUST use only the first 8 chars (4 bytes).
// The privacy contract says "first 4 bytes", not "all of it".
// A regression that used the whole string would leak more
// entropy than intended.
func TestSessionIDPrefix_LongerThanEightUsesPrefix(t *testing.T) {
	// First 8 chars = "deadbeef" -> 0xDEADBEEF.
	// Trailing chars "12345678" should be IGNORED.
	got := sessionIDPrefix("deadbeef12345678")
	want := uint32(0xDEADBEEF)
	if uint32(got) != want {
		t.Errorf("sessionIDPrefix(\"deadbeef12345678\") = 0x%X, want 0x%X (first 8 chars)",
			uint32(got), want)
	}
}

// TestSessionIDPrefix_PrivacyContract pins the non-empty /
// non-trivial guarantee: a valid hex id MUST produce a non-zero
// prefix (otherwise the privacy-preserving grouping collapses
// all sessions into bucket 0). The 8-byte prefix space is
// 4 billion buckets; "deadbeef..." should NOT collapse.
func TestSessionIDPrefix_PrivacyContract(t *testing.T) {
	// Two random-looking hex IDs MUST produce different prefixes.
	a := sessionIDPrefix("deadbeef01234567")
	b := sessionIDPrefix("12345678deadbeef")
	if a == 0 || b == 0 {
		t.Errorf("privacy contract: random hex IDs collapsed to zero (a=%d, b=%d)", a, b)
	}
	if a == b {
		t.Errorf("privacy contract: two distinct hex IDs produced same prefix (a=%d, b=%d)", a, b)
	}
}

// TestNewSessionID_FormatHex pins the session-ID format contract:
// newSessionID() MUST return a 16-character hex string (8 random
// bytes encoded as hex). The privacy-preserving prefix logic
// depends on this format — anything else (base64, raw bytes)
// would break sessionIDPrefix.
func TestNewSessionID_FormatHex(t *testing.T) {
	id := newSessionID()
	if len(id) != 16 {
		t.Errorf("newSessionID length = %d, want 16", len(id))
	}
	// Every character MUST be a valid hex digit.
	for i, c := range id {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Errorf("newSessionID[%d] = %q, not a hex char", i, c)
		}
	}
	// Two successive calls MUST produce different IDs
	// (collision would mean the RNG is broken or seeded constant).
	id2 := newSessionID()
	if id == id2 {
		t.Error("two newSessionID() calls returned the same value; RNG is broken or seeded")
	}
}
