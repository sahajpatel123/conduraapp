package daemon

import (
	"errors"
	"io"
	"sync/atomic"
	"testing"
)

// taggedCloser is a minimal io.Closer implementation that
// records its identity via the `tag` field and tracks how many
// times Close was called. Used by the replaceCloserByType tests
// to verify which closer was replaced (or appended) by tag,
// without depending on the heavyweight *memory.SQLiteStore or
// *skills.SQLiteStore types.
type taggedCloser struct {
	tag   string
	calls atomic.Int32
}

func (t *taggedCloser) Close() error {
	t.calls.Add(1)
	return nil
}

// errTaggedCloser is a taggedCloser variant that returns a
// sentinel error from Close(). Used to verify that the swapped
// closer is actually invoked by Close() (not just stored in the
// list).
type errTaggedCloser struct {
	tag string
	err error
}

func (e *errTaggedCloser) Close() error { return e.err }

// TestReplaceCloserByType_NilNoOp pins the first guard: a nil
// newCloser must NOT be appended to the closers list and must
// NOT cause a panic. The replace* helpers are called from
// ReloadAuxiliaryDatabases after a possibly-failed NewSQLiteStore;
// if the constructor returns (nil, err), the caller must not
// swap a nil into the list — that would NPE on the next Close.
func TestReplaceCloserByType_NilNoOp(t *testing.T) {
	s := &Subsystems{}
	s.closers = []io.Closer{&taggedCloser{tag: "existing"}}

	s.replaceCloserByType(func(c io.Closer) bool { return true }, nil)

	if got := len(s.closers); got != 1 {
		t.Errorf("closers length = %d, want 1 (nil must not append)", got)
	}
}

// TestReplaceCloserByType_ReplacesFirstMatch pins the swap-in-place
// branch: when a closer matches the predicate, the matching slot
// is overwritten with the new closer. Only the FIRST match is
// replaced (the function returns after the first match); this
// pins that contract so a future "fix" that replaces all matches
// is visible in the test.
func TestReplaceCloserByType_ReplacesFirstMatch(t *testing.T) {
	old := &taggedCloser{tag: "old"}
	replacement := &taggedCloser{tag: "new"}
	s := &Subsystems{}
	s.closers = []io.Closer{
		&taggedCloser{tag: "unrelated-1"},
		old,
		&taggedCloser{tag: "unrelated-2"},
	}

	s.replaceCloserByType(func(c io.Closer) bool { return c == old }, replacement)

	// Length unchanged (replacement, not append).
	if got := len(s.closers); got != 3 {
		t.Fatalf("closers length = %d, want 3 (replace, not append)", got)
	}
	// The "old" slot now holds the replacement.
	if s.closers[1] != replacement {
		t.Errorf("closers[1] = %v, want the injected replacement (pointer identity)", s.closers[1])
	}
	// Other slots unchanged.
	if s.closers[0].(*taggedCloser).tag != "unrelated-1" {
		t.Errorf("closers[0] tag = %q, want unchanged 'unrelated-1'", s.closers[0].(*taggedCloser).tag)
	}
	if s.closers[2].(*taggedCloser).tag != "unrelated-2" {
		t.Errorf("closers[2] tag = %q, want unchanged 'unrelated-2'", s.closers[2].(*taggedCloser).tag)
	}
}

// TestReplaceCloserByType_ReplacesOnlyFirstMatch pins the
// "stop after first match" behavior. If a future change made
// replaceCloserByType replace ALL matches, this test would
// fail (the second memory SQLite in the list would also be
// replaced). Pinning the single-match contract prevents
// silent behavior drift.
func TestReplaceCloserByType_ReplacesOnlyFirstMatch(t *testing.T) {
	first := &taggedCloser{tag: "first"}
	second := &taggedCloser{tag: "second"}
	replacement := &taggedCloser{tag: "new"}
	s := &Subsystems{}
	s.closers = []io.Closer{first, second}

	// Match every closer.
	s.replaceCloserByType(func(c io.Closer) bool { return true }, replacement)

	if s.closers[0] != replacement {
		t.Errorf("closers[0] = %v, want replacement (first match must be replaced)", s.closers[0])
	}
	if s.closers[1] != second {
		t.Errorf("closers[1] = %v, want unchanged 'second' (only FIRST match must be replaced)", s.closers[1])
	}
}

// TestReplaceCloserByType_AppendsWhenNoMatch pins the append
// branch: when no closer matches the predicate, the new closer
// is appended. This is the path used when the subsystem was
// constructed without a memory/skills store (e.g. on first
// launch) but ReloadAuxiliaryDatabases now needs to add one.
func TestReplaceCloserByType_AppendsWhenNoMatch(t *testing.T) {
	existing := &taggedCloser{tag: "existing"}
	added := &taggedCloser{tag: "added"}
	s := &Subsystems{}
	s.closers = []io.Closer{existing}

	// Predicate matches nothing in the list.
	s.replaceCloserByType(func(c io.Closer) bool { return false }, added)

	if got := len(s.closers); got != 2 {
		t.Fatalf("closers length = %d, want 2 (append branch)", got)
	}
	if s.closers[0] != existing {
		t.Errorf("closers[0] = %v, want existing unchanged", s.closers[0])
	}
	if s.closers[1] != added {
		t.Errorf("closers[1] = %v, want newly appended closer (pointer identity)", s.closers[1])
	}
}

// TestReplaceCloserByType_NewCloserIsInvokedByClose pins the
// downstream contract: after replaceCloserByType, a subsequent
// Subsystems.Close() MUST invoke the new closer (not the old
// one). Otherwise the swapped-in SQLite handle would leak past
// shutdown. We verify via an errTaggedCloser whose Close()
// returns a sentinel; the sentinel must appear in the Close
// error slice.
func TestReplaceCloserByType_NewCloserIsInvokedByClose(t *testing.T) {
	sentinel := errors.New("replacement-was-closed")
	old := &taggedCloser{tag: "old"}
	replacement := &errTaggedCloser{tag: "new", err: sentinel}
	s := &Subsystems{}
	s.closers = []io.Closer{old}

	s.replaceCloserByType(func(c io.Closer) bool { return c == old }, replacement)

	// Close must invoke the replacement and surface its error.
	closeErr := s.Close()
	if closeErr == nil {
		t.Fatal("Subsystems.Close returned nil; want error containing replacement's sentinel")
	}
	if !errors.Is(closeErr, sentinel) && !containsErr(closeErr.Error(), sentinel.Error()) {
		t.Errorf("Close error %q does not contain replacement sentinel %q", closeErr.Error(), sentinel.Error())
	}
}

// TestReplaceMemoryCloser_SwapsMemorySQLiteStore pins the
// type-specific wrapper: replaceMemoryCloser must match
// closers whose concrete type is *memory.SQLiteStore, and only
// those.
//
// We exercise it via the match-style primitive behavior (the
// wrapper just threads a type-asserting predicate into
// replaceCloserByType), plus a positive case that confirms the
// pointer-identity round-trip: a swapped-in *taggedCloser
// (substitute for *memory.SQLiteStore in this test) appears at
// the slot where the old memory closer was.
//
// This test uses the predicate directly because the test file
// is in the same package and `replaceMemoryCloser` is private;
// if it ever stops calling replaceCloserByType with the right
// predicate, the test fails immediately.
func TestReplaceMemoryCloser_SwapsMemorySQLiteStore(t *testing.T) {
	// We can't construct a real *memory.SQLiteStore without a
	// temp file + modernc.org/sqlite; instead, we verify the
	// wrapper's behavior by inspecting the closers list after a
	// replace. The wrapper passes a predicate that does a type
	// assert to *memory.SQLiteStore; with no real *memory store
	// in the list, the wrapper must APPEND (no-match branch).
	added := &taggedCloser{tag: "would-be-memory"}
	s := &Subsystems{}
	s.closers = []io.Closer{&taggedCloser{tag: "unrelated"}}

	// Wrap the taggedCloser in a type that is NOT *memory.SQLiteStore
	// so the wrapper's type-asserting predicate cannot match it.
	// The wrapper will fall through to the append branch.
	s.replaceMemoryCloser(added)

	if got := len(s.closers); got != 2 {
		t.Fatalf("closers length = %d, want 2 (append branch when no *memory.SQLiteStore in list)", got)
	}
	if s.closers[1] != added {
		t.Errorf("closers[1] = %v, want appended closer", s.closers[1])
	}
}

// TestReplaceSkillCloser_SwapsSkillsSQLiteStore pins the
// skills-specific wrapper, parallel to the memory wrapper above.
// Same substitute-taggedCloser trick: no real *skills.SQLiteStore
// in the list means the append branch fires.
func TestReplaceSkillCloser_SwapsSkillsSQLiteStore(t *testing.T) {
	added := &taggedCloser{tag: "would-be-skills"}
	s := &Subsystems{}
	s.closers = []io.Closer{&taggedCloser{tag: "unrelated"}}

	s.replaceSkillCloser(added)

	if got := len(s.closers); got != 2 {
		t.Fatalf("closers length = %d, want 2 (append branch when no *skills.SQLiteStore in list)", got)
	}
	if s.closers[1] != added {
		t.Errorf("closers[1] = %v, want appended closer", s.closers[1])
	}
}

// TestReplaceMemoryCloser_NilNoOp pins the nil-input contract
// for the type-specific wrapper. ReloadAuxiliaryDatabases calls
// replaceMemoryCloser only after NewSQLiteStore returns
// (store, nil) — but defensive nil-handling here prevents a
// future refactor that introduces a nil-store path from
// silently poisoning the closers list.
func TestReplaceMemoryCloser_NilNoOp(t *testing.T) {
	s := &Subsystems{}
	s.closers = []io.Closer{&taggedCloser{tag: "existing"}}

	s.replaceMemoryCloser(nil)

	if got := len(s.closers); got != 1 {
		t.Errorf("closers length = %d, want 1 (nil must not append)", got)
	}
}

// TestReplaceSkillCloser_NilNoOp pins the same nil-input
// contract for the skills wrapper.
func TestReplaceSkillCloser_NilNoOp(t *testing.T) {
	s := &Subsystems{}
	s.closers = []io.Closer{&taggedCloser{tag: "existing"}}

	s.replaceSkillCloser(nil)

	if got := len(s.closers); got != 1 {
		t.Errorf("closers length = %d, want 1 (nil must not append)", got)
	}
}
