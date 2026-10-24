//go:build windows

package onboarding

import (
	"os"
	"syscall"
	"testing"
)

// denyRead makes path unreadable by this process on Windows.
//
// Windows exposes no "deny read to the owner" bit that os.Chmod
// can clear: mode 0000 only sets the read-only attribute, so the
// file stays readable and this test failed on the windows CI leg.
// The Windows equivalent of "this process cannot read this path"
// is an open handle with dwShareMode = 0: every later open,
// including our own os.ReadFile, is then refused with a sharing
// violation. The handle is kept open until cleanup, which runs
// before t.TempDir's removal (cleanup is last-in, first-out), so
// the temp directory can still be deleted.
func denyRead(t *testing.T, path string) {
	t.Helper()

	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatalf("UTF16PtrFromString(%q): %v", path, err)
	}
	h, err := syscall.CreateFile(
		p,
		syscall.GENERIC_READ,
		0, // dwShareMode 0: no other open of this file may succeed
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		t.Fatalf("open exclusive handle on %q: %v", path, err)
	}
	t.Cleanup(func() { _ = syscall.CloseHandle(h) })

	// Fail loudly here if the lock did not take effect, instead of
	// letting the caller report a misleading "no error" result.
	if _, err := os.ReadFile(path); err == nil {
		t.Fatalf("exclusive handle on %q did not deny reads; the permission path is not being exercised", path)
	}
}
