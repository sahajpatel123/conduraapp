package homedir

import (
	"os"
	"sync"
)

// homeOnce guards the lazy initialization of homeVal. After the
// first successful Dir() call, subsequent calls return the
// cached value without an OS call.
var (
	homeOnce sync.Once
	homeVal  string
	homeErr  error
)

// Dir returns the current user's home directory. The result is
// cached on first call; concurrent callers block on homeOnce and
// see the same value.
//
// The error path is also cached: if the first call fails (e.g.
// $HOME is unset and os.UserHomeDir returns ""), every
// subsequent call returns the same error. Callers that want to
// retry after a config change must call Reset().
func Dir() (string, error) {
	homeOnce.Do(func() {
		homeVal, homeErr = os.UserHomeDir()
	})
	return homeVal, homeErr
}

// MustDir is the panic-on-error variant. Use in init() and
// startup paths where a missing home directory is fatal and
// recovering would mask a real configuration error.
//
// MustDir still respects HOME / USERPROFILE overrides — it's
// "panic if the OS can't tell us where home is", not
// "hardcode a path".
func MustDir() string {
	d, err := Dir()
	if err != nil {
		panic("homedir: " + err.Error())
	}
	return d
}

// Reset clears the cached value so the next Dir() call re-runs
// the OS lookup. Test-only — production code should never
// need to call this, since the home directory does not change
// at runtime.
func Reset() {
	homeOnce = sync.Once{}
	homeVal = ""
	homeErr = nil
}
