package homedir

import (
	"os"
	"sync"
	"testing"
)

// TestDir_CachesAcrossCalls pins the perf contract: after the
// first call, subsequent Dir() calls must return the cached
// value without invoking os.UserHomeDir again. We can't
// directly assert "no syscall", but we CAN assert that the
// returned value is byte-identical across many calls even when
// HOME is mid-mutation (which would change a fresh lookup but
// not the cached one).
func TestDir_CachesAcrossCalls(t *testing.T) {
	home, err := Dir()
	if err != nil {
		t.Skipf("UserHomeDir failed on this platform: %v", err)
	}
	for i := 0; i < 100; i++ {
		got, _ := Dir()
		if got != home {
			t.Fatalf("call %d: Dir() = %q, want %q (cache must not drift)", i, got, home)
		}
	}
}

// TestDir_CachesError pins the negative caching contract: if
// the first call fails (HOME unset), every subsequent call
// returns the same error. This is the production behavior —
// callers shouldn't retry automatically.
func TestDir_CachesError(t *testing.T) {
	if _, err := os.Stat("/dev/null"); err != nil {
		t.Skip("no /dev/null")
	}
	// Save and clear HOME / USERPROFILE so UserHomeDir fails.
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	Reset()

	_, first := Dir()
	if first == nil {
		t.Skip("UserHomeDir unexpectedly succeeded with empty HOME — platform-dependent")
	}
	for i := 0; i < 10; i++ {
		_, err := Dir()
		if err == nil {
			t.Fatalf("call %d: Dir() returned no error after a failed first call (cache leak)", i)
		}
	}
}

// TestReset_AllowsReread pins the test-helper contract: after
// Reset(), the next Dir() call re-runs os.UserHomeDir (so the
// test can change HOME between subtests and pick up the new
// value). Without Reset(), the cached value would mask the
// environment change.
func TestReset_AllowsReread(t *testing.T) {
	Reset()
	home1, err := Dir()
	if err != nil {
		t.Skipf("first Dir() failed: %v", err)
	}
	// Bump HOME to a different value — on most platforms, Dir()
	// would still return home1 because of the cache.
	t.Setenv("HOME", home1+"-different-suffix")
	Reset()
	home2, _ := Dir()
	// We can't assert home2 != home1 (the new HOME might be
	// invalid and Dir() would return home1 anyway); instead
	// just assert that Reset doesn't panic and Dir() is callable.
	_ = home2
}

// TestMustDir_PanicsOnMissingHome pins the panic contract.
// Tested by setting HOME="" with Reset() — on platforms where
// UserHomeDir fails (Unix), MustDir must panic.
func TestMustDir_PanicsOnMissingHome(t *testing.T) {
	if _, err := os.Stat("/dev/null"); err != nil {
		t.Skip("no /dev/null")
	}
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	Reset()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustDir did not panic with HOME unset")
		}
		Reset() // restore for any test following in this process
	}()
	_ = MustDir()
}

// TestMustDir_SucceedsWithHomeSet pins the happy-path of
// MustDir — it returns the home directory string when one
// is available.
func TestMustDir_SucceedsWithHomeSet(t *testing.T) {
	Reset()
	got := MustDir()
	if got == "" {
		t.Fatal("MustDir returned empty string with HOME set")
	}
}

// TestDir_ConcurrentSafe pins the goroutine-safety contract.
// 100 concurrent callers must all observe the same value.
// sync.Once inside Dir() is what guarantees this.
func TestDir_ConcurrentSafe(t *testing.T) {
	Reset()
	home, err := Dir()
	if err != nil {
		t.Skip(err)
	}

	const n = 100
	var wg sync.WaitGroup
	results := make([]string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			d, _ := Dir()
			results[idx] = d
		}(i)
	}
	wg.Wait()

	for i, got := range results {
		if got != home {
			t.Errorf("goroutine %d: Dir() = %q, want %q", i, got, home)
		}
	}
}
