//go:build !windows

package onboarding

import (
	"os"
	"testing"
)

// denyRead makes path unreadable by this process on POSIX systems
// by clearing every read bit. The file is restored to mode 0600 on
// cleanup so that t.TempDir can remove it again.
func denyRead(t *testing.T, path string) {
	t.Helper()
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatalf("chmod 000 %q: %v", path, err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
}
