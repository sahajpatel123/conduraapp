package audit

import (
	"context"
	"strings"
	"testing"
)

// TestAppendForTest_RedactsMessage confirms the test helper applies
// sanitize.RedactSecrets to the Message field before writing. This
// is the safety net that keeps future test fixtures from leaking
// fake API keys or other credential shapes into the audit chain.
func TestAppendForTest_RedactsMessage(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()

	const fakeKey = "ghp_abc123def456ghi789jkl012mno345pqr"
	in := Event{
		Actor:   "u",
		Action:  "llm.chat",
		Message: "prompt contains " + fakeKey,
	}
	if err := AppendForTest(t, l, in); err != nil {
		t.Fatal(err)
	}

	evs, err := l.List(ctx, Query{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 {
		t.Fatalf("got %d events, want 1", len(evs))
	}
	if strings.Contains(evs[0].Message, fakeKey) {
		t.Errorf("AppendForTest leaked the fake credential: %q", evs[0].Message)
	}
	if !strings.Contains(evs[0].Message, "<redacted>") {
		t.Errorf("AppendForTest did not insert redaction marker; got %q", evs[0].Message)
	}
}

// TestAppendForTest_LeavesEmptyMessageAlone documents that an empty
// Message passes through untouched (no spurious "<redacted>" for
// empty input).
func TestAppendForTest_LeavesEmptyMessageAlone(t *testing.T) {
	l := setupTestLog(t)
	if err := AppendForTest(t, l, Event{
		Actor: "u", Action: "no.message",
	}); err != nil {
		t.Fatal(err)
	}
	evs, err := l.List(context.Background(), Query{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 {
		t.Fatalf("got %d events, want 1", len(evs))
	}
	if evs[0].Message != "" {
		t.Errorf("empty Message should stay empty; got %q", evs[0].Message)
	}
}
