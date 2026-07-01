package audit

import (
	"context"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/sanitize"
)

// AppendForTest is the recommended way to append events in tests
// that involve user-derived content. It wraps Append with
// sanitize.RedactSecrets on the Message field before writing so
// test fixtures that hardcode a fake API key, a fake token, etc.
// do not end up in the audit chain verbatim.
//
// AppendForTest deliberately has the same return type and call
// shape as Append so existing test sites can swap to it with a
// one-token change. Production callers must continue to use
// Append directly so the redaction policy remains an explicit
// per-call decision (some test fixtures intentionally verify the
// unredacted path).
//
// Usage:
//
//	if err := audit.AppendForTest(t, log, audit.Event{
//	    Actor: "u", Action: "llm.chat",
//	    Message: "prompt=" + userPrompt,
//	}); err != nil { t.Fatal(err) }
//
// The t parameter is only used for t.Helper() — AppendForTest
// does not assert, so it works the same way in subtests and
// t.Run blocks.
func AppendForTest(t *testing.T, l *Log, e Event) error {
	t.Helper()
	if e.Message != "" {
		e.Message = sanitize.RedactSecrets(e.Message)
	}
	return l.Append(context.Background(), e)
}
