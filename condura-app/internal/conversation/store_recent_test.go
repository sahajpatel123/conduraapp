package conversation

import (
	"context"
	"errors"
	"testing"
)

// appendN is a small helper that appends n messages with
// monotonically-increasing content ("msg-0", "msg-1", ...) so
// the chronological-order assertions below are unambiguous.
// Using a numbered prefix also makes test failure messages
// self-describing.
func appendN(t *testing.T, s *Store, ctx context.Context, convID int64, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		if err := s.Append(ctx, convID, Message{
			Role:    "user",
			Content: "msg-" + itoaSmall(i),
		}); err != nil {
			t.Fatalf("Append(%d): %v", i, err)
		}
	}
}

// itoaSmall converts 0..999 to a string without importing
// strconv into this small test file. (The conversation
// package already imports encoding/json; adding strconv is
// fine but this keeps the helper self-contained.)
func itoaSmall(n int) string {
	if n == 0 {
		return "0"
	}
	var b [3]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// TestStore_GetRecentMessages_NotFound pins the first guard:
// GetRecentMessages MUST return ErrNotFound if the conversation
// does not exist. The session.Run path calls this for every
// user query; an unhandled error here would surface to the
// user as "query recent messages: <db error>" instead of a
// clean "conversation not found".
func TestStore_GetRecentMessages_NotFound(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	_, err := s.GetRecentMessages(ctx, 99999, 10)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetRecentMessages(missing conv) err = %v, want ErrNotFound", err)
	}
}

// TestStore_GetRecentMessages_AllMessagesWhenLimitZero pins
// the limit=0 contract: no LIMIT clause is appended, so ALL
// messages in the conversation are returned. This is the path
// used by sessions that want the full history (e.g. resume).
func TestStore_GetRecentMessages_AllMessagesWhenLimitZero(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	conv, err := s.Create(ctx, "full-history")
	if err != nil {
		t.Fatal(err)
	}
	appendN(t, s, ctx, conv.ID, 5)

	got, err := s.GetRecentMessages(ctx, conv.ID, 0)
	if err != nil {
		t.Fatalf("GetRecentMessages: %v", err)
	}
	if len(got) != 5 {
		t.Errorf("limit=0 returned %d messages, want 5 (all)", len(got))
	}
}

// TestStore_GetRecentMessages_LimitReturnsNMostRecent pins
// the limit>0 contract: returns the N most-recent messages
// (in chronological order — the function reverses the SQL's
// DESC result so the LLM sees old→new).
func TestStore_GetRecentMessages_LimitReturnsNMostRecent(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	conv, err := s.Create(ctx, "limited-history")
	if err != nil {
		t.Fatal(err)
	}
	appendN(t, s, ctx, conv.ID, 10)

	got, err := s.GetRecentMessages(ctx, conv.ID, 3)
	if err != nil {
		t.Fatalf("GetRecentMessages: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("limit=3 returned %d messages, want 3", len(got))
	}
	// Must be the LAST 3 messages in chronological order:
	// msg-7, msg-8, msg-9 (the function sorts ASC after SQL DESC).
	wantContents := []string{"msg-7", "msg-8", "msg-9"}
	for i, want := range wantContents {
		if got[i].Content != want {
			t.Errorf("got[%d].Content = %q, want %q (chronological order, most-recent-N)", i, got[i].Content, want)
		}
	}
}

// TestStore_GetRecentMessages_ChronologicalOrder pins the
// sort-direction contract: the SQL orders DESC (newest first)
// then the function reverses to ASC (oldest first) for the
// LLM prompt. A regression that removed the reverse would
// feed the LLM messages in wrong order.
//
// We verify by appending messages with distinguishable content
// in a known order, then asserting the returned slice matches
// the original append order.
func TestStore_GetRecentMessages_ChronologicalOrder(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	conv, err := s.Create(ctx, "order-test")
	if err != nil {
		t.Fatal(err)
	}
	for _, content := range []string{"first", "second", "third", "fourth"} {
		if err := s.Append(ctx, conv.ID, Message{Role: "user", Content: content}); err != nil {
			t.Fatal(err)
		}
	}

	got, err := s.GetRecentMessages(ctx, conv.ID, 0)
	if err != nil {
		t.Fatalf("GetRecentMessages: %v", err)
	}
	wantOrder := []string{"first", "second", "third", "fourth"}
	if len(got) != len(wantOrder) {
		t.Fatalf("got %d messages, want %d", len(got), len(wantOrder))
	}
	for i, want := range wantOrder {
		if got[i].Content != want {
			t.Errorf("got[%d].Content = %q, want %q (chronological ASC)", i, got[i].Content, want)
		}
	}
}

// TestStore_GetRecentMessages_LimitLargerThanHistory pins the
// boundary: limit > number of messages. Should return all
// messages (capped by history length, not by the LIMIT value).
// A regression here would silently drop the oldest messages.
func TestStore_GetRecentMessages_LimitLargerThanHistory(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	conv, err := s.Create(ctx, "short")
	if err != nil {
		t.Fatal(err)
	}
	appendN(t, s, ctx, conv.ID, 2)

	got, err := s.GetRecentMessages(ctx, conv.ID, 100)
	if err != nil {
		t.Fatalf("GetRecentMessages: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("limit=100 with 2 messages returned %d, want 2 (LIMIT is upper bound, not target)", len(got))
	}
}

// TestStore_GetRecentMessages_EmptyConversation pins the
// zero-message boundary: a conversation with no messages
// returns an empty slice (not nil-without-error, not
// ErrNotFound — the conversation EXISTS, it just has no
// messages yet).
func TestStore_GetRecentMessages_EmptyConversation(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()

	conv, err := s.Create(ctx, "empty")
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.GetRecentMessages(ctx, conv.ID, 10)
	if err != nil {
		t.Fatalf("GetRecentMessages on empty conv: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("empty conv returned %d messages, want 0", len(got))
	}
}
