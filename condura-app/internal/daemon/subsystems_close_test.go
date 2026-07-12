package daemon

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"
)

// errAlwaysFail is a synthetic closer error used to verify
// error propagation through sync.Once (the first call records
// the error; subsequent calls must NOT re-run the closure and
// must NOT re-collect or re-return the error).
var errAlwaysFail = errors.New("synthetic close error")

// failCloser is a closer that always returns errAlwaysFail. It
// also counts how many times Close was invoked so the test can
// assert idempotency at the closer level (not just at the
// Subsystems.Close level).
type failCloser struct {
	calls atomic.Int32
}

func (f *failCloser) Close() error {
	f.calls.Add(1)
	return errAlwaysFail
}

// TestSubsystems_Close_Idempotent verifies that calling Close()
// twice does not re-iterate the closers list, re-run each
// closer's Close(), or re-collect their errors. Pre-§1.6, the
// second call would re-iterate and (a) potentially re-close an
// already-closed *sql.DB (panic on closed handle) and (b) lose
// the first call's error in the second call's error slice.
func TestSubsystems_Close_Idempotent(t *testing.T) {
	fc := &failCloser{}
	s := &Subsystems{closers: []io.Closer{fc}}

	// First Close: should call failCloser.Close() exactly once
	// and return errAlwaysFail wrapped in "subsystems close".
	err1 := s.Close()
	if err1 == nil {
		t.Fatal("first Close should return an error (synthetic failCloser)")
	}
	// The production Close() wraps the error slice with
	// fmt.Errorf("subsystems close: %v", errs) — %v on a
	// []error flattens to the joined Error() strings. We use
	// errors.Is as the canonical check, but it requires the
	// wrapped error to be a single error (not a slice). To be
	// resilient to either wrapping shape, check both forms:
	// (a) errors.Is on the wrapped value, (b) string contains.
	if !errors.Is(err1, errAlwaysFail) && !containsErr(err1.Error(), errAlwaysFail.Error()) {
		t.Fatalf("first Close error %q should contain errAlwaysFail %q", err1.Error(), errAlwaysFail.Error())
	}
	if fc.calls.Load() != 1 {
		t.Fatalf("after first Close: failCloser.calls = %d, want 1", fc.calls.Load())
	}

	// Second Close: should be a no-op. failCloser must NOT be
	// invoked again, and the return value must be nil (no error
	// re-collected from a second iteration).
	err2 := s.Close()
	if err2 != nil {
		t.Fatalf("second Close should return nil, got: %v", err2)
	}
	if fc.calls.Load() != 1 {
		t.Fatalf("after second Close: failCloser.calls = %d, want still 1 (idempotent)", fc.calls.Load())
	}

	// Third Close for good measure.
	if err3 := s.Close(); err3 != nil {
		t.Fatalf("third Close should return nil, got: %v", err3)
	}
	if fc.calls.Load() != 1 {
		t.Fatalf("after third Close: failCloser.calls = %d, want still 1", fc.calls.Load())
	}
}

// TestSubsystems_Close_ConcurrentRace fires N concurrent Close()
// calls and asserts that the underlying closer runs exactly
// once. Pre-§1.6 the race detector would have caught the
// double-close path; now we assert the functional contract.
func TestSubsystems_Close_ConcurrentRace(t *testing.T) {
	fc := &failCloser{}
	s := &Subsystems{closers: []io.Closer{fc}}

	const N = 64
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			_ = s.Close()
		}()
	}
	wg.Wait()

	if fc.calls.Load() != 1 {
		t.Fatalf("after %d concurrent Close calls: failCloser.calls = %d, want 1",
			N, fc.calls.Load())
	}
}

// TestSubsystems_CloseDatabases_Idempotent verifies that
// CloseDatabases() is independently idempotent — a second
// call must not re-close already-closed handles. This matters
// because backup.restore calls CloseDatabases() before a
// Storage.Reload(); if the SIGINT shutdown goroutine also
// calls Close() during the same window, the relationship
// between the two must not corrupt the closers list.
func TestSubsystems_CloseDatabases_Idempotent(t *testing.T) {
	fc := &failCloser{}
	s := &Subsystems{closers: []io.Closer{fc}}

	// CloseDatabases must not panic, and must invoke the closer
	// exactly once.
	s.CloseDatabases()
	if fc.calls.Load() != 1 {
		t.Fatalf("after first CloseDatabases: failCloser.calls = %d, want 1", fc.calls.Load())
	}

	// Second CloseDatabases: no-op.
	s.CloseDatabases()
	if fc.calls.Load() != 1 {
		t.Fatalf("after second CloseDatabases: failCloser.calls = %d, want 1", fc.calls.Load())
	}
}

// TestSubsystems_CloseAndCloseDatabases_Independent verifies
// that calling one method does not lock out the other. The
// two guards (closeOnce, closeDatabasesOnce) are independent
// because backup.restore relies on this — it calls
// CloseDatabases, then Storage.Reload, then later the daemon
// shutdown path calls Close. If they shared a Once, the second
// call would silently skip the close work.
func TestSubsystems_CloseAndCloseDatabases_Independent(t *testing.T) {
	var dbCalls, fullCalls atomic.Int32
	dbCloser := closerFunc(func() error { dbCalls.Add(1); return nil })
	fullCloser := closerFunc(func() error { fullCalls.Add(1); return nil })

	s := &Subsystems{closers: []io.Closer{fullCloser}}

	// CloseDatabases must run the full closers list (pre-existing
	// behavior — CloseDatabases iterates all closers, not just
	// DBs). This is verified by fullCloser.calls incrementing.
	s.CloseDatabases()
	if fullCalls.Load() != 1 {
		t.Fatalf("CloseDatabases must run closers once; got fullCalls = %d", fullCalls.Load())
	}
	_ = dbCloser

	// After CloseDatabases, Close must STILL run (independent
	// guard). This is the regression we're protecting — pre-§1.6
	// if both methods shared a guard, this would be a no-op.
	if err := s.Close(); err != nil {
		t.Fatalf("Close after CloseDatabases returned error: %v", err)
	}
	if fullCalls.Load() != 2 {
		t.Fatalf("Close must run closers even after CloseDatabases; got fullCalls = %d, want 2",
			fullCalls.Load())
	}

	// Both methods are idempotent — third calls of either
	// should be no-ops.
	s.CloseDatabases()
	if err := s.Close(); err != nil {
		t.Fatalf("third Close returned error: %v", err)
	}
	if fullCalls.Load() != 2 {
		t.Fatalf("third calls must not re-run closers; got fullCalls = %d, want still 2",
			fullCalls.Load())
	}
}

// closerFunc adapts a function to the io.Closer interface so
// tests can use closures without defining a one-off type.
type closerFunc func() error

func (f closerFunc) Close() error { return f() }

// containsErr is a thin string-match fallback for assertions
// against errors that the production code wraps via
// fmt.Errorf("... %v", []error{...}). The %v rendering of a
// []error is the joined Error() output, so string contains is
// the most reliable check across wrapping shapes.
func containsErr(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle ||
		(len(needle) > 0 && stringIndex(haystack, needle) >= 0))
}

func stringIndex(s, sub string) int {
	// Tiny substring search — avoids importing strings just for
	// this one test helper.
	if len(sub) == 0 {
		return 0
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
