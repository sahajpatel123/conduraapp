//go:build windows

package onboarding

import (
	"os"
	"testing"

	"golang.org/x/sys/windows"
)

// denyRead makes path unreadable by this process on Windows, by
// holding a mandatory whole-file lock on it.
//
// Windows has no per-file read bits for os.Chmod to clear: mode 0000
// only sets the read-only attribute, so the file stays readable. A
// deny ACE is not enough either, and that was measured on the CI
// runner: with the DACL "D:PAI(D;;FR;;;LA)" in place - a protected
// DACL denying FILE_GENERIC_READ to the process's own account - the
// read still succeeded, because Go opens read-only files with
// FILE_FLAG_BACKUP_SEMANTICS and that token carries the backup
// privilege, which overrides access checks. The same DACL did deny a
// write, so it is enforced; it just cannot stop this process from
// reading.
//
// Byte-range locks are enforced by the I/O manager rather than by the
// access check, so no privilege bypasses them: another handle that
// reads the locked range fails with ERROR_LOCK_VIOLATION. That is a
// genuine "cannot read this path", which is exactly what the caller
// asserts (a failed read wrapped with "read EULA").
func denyRead(t *testing.T, path string) {
	t.Helper()

	// Read-only access only: a read handle is not denied by the file's
	// DACL for this token, and asking for write access would be.
	h, err := windows.CreateFile(
		windows.StringToUTF16Ptr(path),
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		t.Fatalf("open %q to lock it: %v", path, err)
	}
	t.Cleanup(func() { _ = windows.CloseHandle(h) })

	if err := windows.LockFileEx(h, windows.LOCKFILE_EXCLUSIVE_LOCK, 0, ^uint32(0), ^uint32(0), new(windows.Overlapped)); err != nil {
		t.Fatalf("lock %q: %v", path, err)
	}

	// Fail loudly here if the denial did not take effect, instead of
	// letting the caller report a misleading "no error" result.
	if _, err := os.ReadFile(path); err == nil {
		t.Fatalf("whole-file lock on %q did not prevent a read; the read-failure path is not being exercised", path)
	}
}
