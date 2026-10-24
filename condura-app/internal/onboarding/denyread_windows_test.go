//go:build windows

package onboarding

import (
	"os"
	"runtime"
	"testing"

	"golang.org/x/sys/windows"
)

// denyRead makes path unreadable by this process on Windows.
//
// Windows has no per-file read bits for os.Chmod to clear: mode
// 0000 only sets the read-only attribute, so the file stays
// readable and this test used to be skipped there. The Windows
// equivalent of "this user has no read permission" is an explicit
// DENY ACE in the file's DACL, which the access check evaluates
// before any allow ACE and which applies to administrators too.
// Because this process created the file it remains its owner, and
// an owner always holds READ_CONTROL and WRITE_DAC implicitly, so
// cleanup can restore the DACL.
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
		// FILE_GENERIC_READ, not GENERIC_READ: an ACE must carry the
		// specific access bits. GENERIC_* is only expanded for the
		// access a caller *requests* at open time, never for the bits
		// stored in a DACL, so a GENERIC_READ ACE silently matches
		// nothing and denies nothing.
		AccessPermissions: windows.FILE_GENERIC_READ,
		AccessMode:        windows.DENY_ACCESS,
		Inheritance:       windows.NO_INHERITANCE,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_USER,
			TrusteeValue: windows.TrusteeValueFromSID(user.User.Sid),
		},
	}}, nil)
	runtime.KeepAlive(user)
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
	// Hand the file back with an unrestricted DACL before t.TempDir
	// removes it (this cleanup is registered later, so it runs first).
	t.Cleanup(func() {
		_ = windows.SetNamedSecurityInfo(
			path,
			windows.SE_FILE_OBJECT,
			windows.DACL_SECURITY_INFORMATION|windows.UNPROTECTED_DACL_SECURITY_INFORMATION,
			nil, nil, nil, nil,
		)
	})

	// Fail loudly here if the deny did not take effect, instead of
	// letting the caller report a misleading "no error" result. The
	// message re-reads the file's DACL so that a missing ACE and a
	// volume that does not enforce DACLs are distinguishable.
	if _, err := os.ReadFile(path); err == nil {
		dacl := "<unreadable>"
		if sd, serr := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION); serr == nil {
			dacl = sd.String()
		}
		t.Fatalf("deny-read ACE on %q did not prevent a read; DACL=%s user=%s; the permission path is not being exercised",
			path, dacl, user.User.Sid.String())
	}
}
