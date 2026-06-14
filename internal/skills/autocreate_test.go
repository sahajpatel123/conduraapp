package skills

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

// helper: in-memory store for tests. Wraps SQLiteStore with a
// temp file path so each test is isolated.
func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	dir := t.TempDir()
	s, err := NewSQLiteStore(filepath.Join(dir, "skills.db"))
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestAutoCreate_NilStoreReturnsSentinel(t *testing.T) {
	ac := NewAutoCreate(nil)
	err := ac.Observe(context.Background(), "s1", "open chrome", []string{"step"})
	if !errors.Is(err, ErrStoreMissing) {
		t.Fatalf("expected ErrStoreMissing, got %v", err)
	}
}

func TestAutoCreate_EmptyQueryReturnsSentinel(t *testing.T) {
	ac := NewAutoCreate(newTestStore(t))
	for _, q := range []string{"", "   ", "\t\n"} {
		err := ac.Observe(context.Background(), "s1", q, nil)
		if !errors.Is(err, ErrEmptyQuery) {
			t.Fatalf("expected ErrEmptyQuery for %q, got %v", q, err)
		}
	}
}

func TestAutoCreate_BelowThresholdReturnsSentinel(t *testing.T) {
	store := newTestStore(t)
	ac := NewAutoCreate(store)
	for i := 0; i < MinSamples-1; i++ {
		err := ac.Observe(context.Background(), "s1", "open chrome", []string{"click"})
		if !errors.Is(err, ErrNoSkillCreated) {
			t.Fatalf("observation %d: expected ErrNoSkillCreated, got %v", i, err)
		}
	}
	if got := ac.PendingCount(); got != 1 {
		t.Fatalf("pending count: want 1, got %d", got)
	}
	// No skill persisted yet.
	hits, err := store.Search(context.Background(), "open chrome", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) != 0 {
		t.Fatalf("expected 0 skills, got %d", len(hits))
	}
}

func TestAutoCreate_AtThresholdCreatesCommunitySkill(t *testing.T) {
	store := newTestStore(t)
	ac := NewAutoCreate(store)
	ctx := context.Background()
	// First two: below threshold.
	for i := 0; i < MinSamples-1; i++ {
		if err := ac.Observe(ctx, "s1", "open chrome", []string{"click icon"}); !errors.Is(err, ErrNoSkillCreated) {
			t.Fatalf("obs %d: %v", i, err)
		}
	}
	// Third: threshold.
	if err := ac.Observe(ctx, "s1", "open chrome", []string{"click icon", "wait 200ms"}); err != nil {
		t.Fatalf("third obs: %v", err)
	}
	// Pending should be empty now.
	if got := ac.PendingCount(); got != 0 {
		t.Fatalf("pending count after create: want 0, got %d", got)
	}
	// Skill should exist with TrustCommunity, NEVER official.
	hits, err := store.Search(ctx, "open chrome", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(hits))
	}
	if hits[0].Trust != TrustCommunity {
		t.Fatalf("trust: want community, got %s", hits[0].Trust)
	}
	if hits[0].Source != "auto" {
		t.Fatalf("source: want auto, got %s", hits[0].Source)
	}
}

func TestAutoCreate_AlreadyCreatedIncrementsUsage(t *testing.T) {
	store := newTestStore(t)
	ac := NewAutoCreate(store)
	ctx := context.Background()
	// Reach threshold.
	for i := 0; i < MinSamples; i++ {
		_ = ac.Observe(ctx, "s1", "open chrome", []string{"x"})
	}
	hits, _ := store.Search(ctx, "open chrome", 10)
	if len(hits) != 1 {
		t.Fatalf("setup: want 1, got %d", len(hits))
	}
	id := hits[0].ID
	// Observe again — should NOT create a second skill, and
	// should bump the success counter.
	if err := ac.Observe(ctx, "s1", "open chrome", []string{"x"}); err != nil {
		t.Fatalf("re-observe: %v", err)
	}
	hits2, _ := store.Search(ctx, "open chrome", 10)
	if len(hits2) != 1 {
		t.Fatalf("expected still 1 skill, got %d", len(hits2))
	}
	got, _ := store.Get(ctx, id)
	if got.SuccessCount < MinSamples+1 {
		t.Fatalf("success count: want >= %d, got %d", MinSamples+1, got.SuccessCount)
	}
}

func TestAutoCreate_DifferentTriggerPhrasesBucketSeparately(t *testing.T) {
	store := newTestStore(t)
	ac := NewAutoCreate(store)
	ctx := context.Background()
	// Two distinct triggers, both reach threshold.
	for i := 0; i < MinSamples; i++ {
		_ = ac.Observe(ctx, "s1", "open chrome", []string{"a"})
	}
	for i := 0; i < MinSamples; i++ {
		_ = ac.Observe(ctx, "s1", "open firefox", []string{"b"})
	}
	// 2 skills now.
	hits, _ := store.List(ctx, 50)
	if len(hits) != 2 {
		t.Fatalf("want 2 skills, got %d", len(hits))
	}
}

func TestAutoCreate_NormalizeCollapsesCaseAndWhitespace(t *testing.T) {
	cases := []struct{ in, want string }{
		{"open chrome", "open chrome"},
		{"Open Chrome", "open chrome"},
		{"  open   chrome  ", "open chrome"},
		{"open\tchrome", "open chrome"},
		{"OPEN CHROME", "open chrome"},
	}
	for _, c := range cases {
		if got := normalize(c.in); got != c.want {
			t.Errorf("normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestAutoCreate_LongQueryTruncatedTo80Chars(t *testing.T) {
	long := ""
	for i := 0; i < 200; i++ {
		long += "x"
	}
	got := normalize(long)
	if len(got) != 80 {
		t.Fatalf("len: want 80, got %d", len(got))
	}
}

func TestAutoCreate_LRUEvictionBoundsMemory(t *testing.T) {
	store := newTestStore(t)
	ac := NewAutoCreate(store)
	ctx := context.Background()
	// Stuff MaxPending+10 distinct triggers into the map.
	// We won't reach the threshold for any of them (we only
	// observe each once) so the skill store stays empty and
	// we can test the LRU behavior in isolation.
	for i := 0; i < MaxPending+10; i++ {
		q := "trigger " + itoa(i)
		_ = ac.Observe(ctx, "s1", q, nil)
		// ErrNoSkillCreated is expected.
	}
	if got := ac.PendingCount(); got != MaxPending {
		t.Fatalf("pending: want %d (bounded), got %d", MaxPending, got)
	}
}

func TestAutoCreate_ResetClearsPending(t *testing.T) {
	store := newTestStore(t)
	ac := NewAutoCreate(store)
	ctx := context.Background()
	_ = ac.Observe(ctx, "s1", "open chrome", nil)
	_ = ac.Observe(ctx, "s1", "open firefox", nil)
	if ac.PendingCount() != 2 {
		t.Fatalf("setup: want 2")
	}
	ac.Reset()
	if ac.PendingCount() != 0 {
		t.Fatalf("after reset: want 0")
	}
}

func TestAutoCreate_StepsAreDeduplicatedAndCapped(t *testing.T) {
	store := newTestStore(t)
	ac := NewAutoCreate(store)
	ctx := context.Background()
	// Reach threshold with duplicate-adjacent steps.
	for i := 0; i < MinSamples; i++ {
		steps := []string{"click", "click", "wait", "wait", "click"}
		_ = ac.Observe(ctx, "s1", "open chrome", steps)
	}
	hits, _ := store.Search(ctx, "open chrome", 10)
	if len(hits) != 1 {
		t.Fatalf("want 1, got %d", len(hits))
	}
	// Steps should be [click, wait, click] (deduped).
	want := []string{"click", "wait", "click"}
	got := hits[0].Steps
	if len(got) != len(want) {
		t.Fatalf("steps: want %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("step[%d]: want %s, got %s", i, want[i], got[i])
		}
	}
}

func TestAutoCreate_HumanizeProducesTitleCase(t *testing.T) {
	cases := []struct{ in, want string }{
		{"open chrome", "Open Chrome"},
		{"send email", "Send Email"},
		{"", ""},
	}
	for _, c := range cases {
		if got := humanize(c.in); got != c.want {
			t.Errorf("humanize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestAutoCreate_StoreFailureRollsBackEvidence(t *testing.T) {
	// Simulate a store failure by closing it before the third
	// observation. AutoCreate should surface the error and the
	// pending map should be at MinSamples-1, not 0 (so the
	// next observe can retry).
	store := newTestStore(t)
	ac := NewAutoCreate(store)
	ctx := context.Background()
	// Two observations (below threshold).
	for i := 0; i < MinSamples-1; i++ {
		_ = ac.Observe(ctx, "s1", "open chrome", []string{"a"})
	}
	// Now close the store; the third Observe will fail on Create.
	_ = store.Close()
	err := ac.Observe(ctx, "s1", "open chrome", []string{"a"})
	if err == nil {
		t.Fatalf("expected error from closed store")
	}
	// The pending map should still have an entry (we rolled
	// back the last observation so the user can retry without
	// waiting 3 more sessions).
	if got := ac.PendingCount(); got != 1 {
		t.Fatalf("pending: want 1, got %d", got)
	}
}

// itoa is a tiny int-to-string helper so we don't import strconv
// just for tests.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
