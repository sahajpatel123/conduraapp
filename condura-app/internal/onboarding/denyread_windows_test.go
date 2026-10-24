//go:build windows

package onboarding

import (
	"os"
	"testing"

	"golang.org/x/sys/windows"
)

// denyRead makes path unreadable by this process on Windows.
//
// Windows has no per-file read bits for os.Chmod to clear: mode 0000
// only sets the read-only attribute, so the file stays readable. Two
// stronger mechanisms are applied here because neither is sufficient
// on its own on every host:
//
//   - An explicit DENY FILE_GENERIC_READ ACE in the file's DACL, the
//     Windows equivalent of "this account has no read permission".
//     Note FILE_GENERIC_READ, not GENERIC_READ: generic rights are
//     only expanded for the access a caller requests at open time,
//     never for the bits stored in an ACE.
//   - A mandatory byte-range lock over the whole file. Windows file
//     locks are enforced by the I/O manager, not by the security
//     reference monitor, so they deny reads even to a token that is
//     not subject to the DACL: the CI runner runs the tests as the
//     built-in Administrator, which keeps reading a file whose DACL
//     denies it (observed: DACL "D:PAI(D;;FR;;;LA)", read allowed).
//
// Either mechanism alone yields a genuine "cannot read this path", and
// the caller only asserts that the read fails and that the error is
// wrapped. The file's owner keeps READ_CONTROL and WRITE_DAC
// implicitly, so cleanup can restore the DACL and drop the lock.
func denyRead(t *testing.T, path string) {
	t.Helper()

	tok, err := windows.OpenCurrentProcessToken()
	if err != nil {
		t.Fatalf("open process token: %v", err)
	}
	defer tok.Close()
	user, err := tok.GetTokenUser()
	if err != nil {
		t.Fatalf("read process token user: %v", err)
	}

	acl, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.FILE_GENERIC_READ,
		AccessMode:        windows.DENY_ACCESS,
		Inheritance:       windows.NO_INHERITANCE,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_USER,
			TrusteeValue: windows.TrusteeValueFromSID(user.User.Sid),
		},
	}}, nil)
	if err != nil {
		t.Fatalf("build deny-read ACL for %q: %v", path, err)
	}

	// PROTECTED drops the inherited ACEs so the deny cannot be
	// outweighed by an inherited allow.
	if err := windows.SetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil, nil, acl, nil,
	); err != nil {
		t.Fatalf("apply deny-read ACL to %q: %v", path, err)
	}
	t.Cleanup(func() {
		_ = windows.SetNamedSecurityInfo(
			path,
			windows.SE_FILE_OBJECT,
			windows.DACL_SECURITY_INFORMATION|windows.UNPROTECTED_DACL_SECURITY_INFORMATION,
			nil, nil, nil, nil,
		)
	})

	h, err := windows.CreateFile(
		windows.StringToUTF16Ptr(path),
		windows.GENERIC_READ|windows.GENERIC_WRITE,
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

	// Whole-file exclusive lock: any other handle reading the range
	// fails with ERROR_LOCK_VIOLATION.
	if err := windows.LockFileEx(h, windows.LOCKFILE_EXCLUSIVE_LOCK, 0, ^uint32(0), ^uint32(0), new(windows.Overlapped)); err != nil {
		t.Fatalf("lock %q: %v", path, err)
	}

	// Fail loudly here if the denial did not take effect, instead of
	// letting the caller report a misleading "no error" result. The
	// message re-reads the file's DACL so that a missing ACE and a
	// host that ignores it stay distinguishable.
	if _, err := os.ReadFile(path); err == nil {
		dacl := "<unreadable>"
		if sd, serr := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION); serr == nil {
			dacl = sd.String()
		}
		t.Fatalf("deny-read ACE plus whole-file lock on %q still allowed a read; DACL=%s user=%s; the read-failure path is not being exercised",
			path, dacl, user.User.Sid.String())
	}
}
