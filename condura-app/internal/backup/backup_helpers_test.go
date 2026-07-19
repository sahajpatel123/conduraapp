package backup

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

// TestRenameToFinal_NotTmpPathReturnsUnchanged pins the
// no-op contract: when tmpPath doesn't end with ".zip.tmp",
// renameToFinal MUST return it unchanged (no error). This
// protects against the case where renameToFinal is called on
// an already-renamed file or a non-tmp path.
func TestRenameToFinal_NotTmpPathReturnsUnchanged(t *testing.T) {
	b := &Manager{}
	cases := []string{
		"/tmp/backup.zip",
		"/tmp/backup",
		"/tmp/backup.tmp",  // .tmp but not .zip.tmp
		"/tmp/backup.zipx", // .zip but not .zip.tmp
	}
	for _, p := range cases {
		got, err := b.renameToFinal(p)
		if err != nil {
			t.Errorf("renameToFinal(%q) = error %v; want nil", p, err)
		}
		if got != p {
			t.Errorf("renameToFinal(%q) = %q; want unchanged", p, got)
		}
	}
}

// TestRenameToFinal_TmpPathSuffixRenamed pins the main contract:
// when tmpPath ends with ".zip.tmp", renameToFinal MUST strip
// the .tmp suffix and rename the file (so it ends with .zip).
// The atomic rename is critical — a partial-write backup must
// either complete fully or be invisible.
func TestRenameToFinal_TmpPathSuffixRenamed(t *testing.T) {
	dir := t.TempDir()
	tmpPath := filepath.Join(dir, "backup.zip.tmp")
	if err := os.WriteFile(tmpPath, []byte("archive data"), 0o600); err != nil {
		t.Fatal(err)
	}

	b := &Manager{}
	got, err := b.renameToFinal(tmpPath)
	if err != nil {
		t.Fatalf("renameToFinal: %v", err)
	}
	want := filepath.Join(dir, "backup.zip")
	if got != want {
		t.Errorf("renameToFinal returned %q; want %q (suffix .tmp stripped)", got, want)
	}
	// The new file MUST exist at the renamed path.
	if _, err := os.Stat(want); err != nil {
		t.Errorf("renamed file missing: %v", err)
	}
	// The old .tmp file MUST NOT exist (rename is atomic).
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf("old .tmp file still exists; rename was not atomic")
	}
}

// TestDeriveKeyBase64_Base64Encoded pins the base64 encoding
// contract: DeriveKeyBase64 MUST return the same bytes as
// DeriveKey, but base64-encoded. A regression that used a
// different encoding (hex, base32) would break the first-backup
// notice flow that relies on this specific format.
func TestDeriveKeyBase64_Base64Encoded(t *testing.T) {
	masterKey := []byte("test-master-key-0123456789abcdef")

	got, err := DeriveKeyBase64(masterKey)
	if err != nil {
		t.Fatalf("DeriveKeyBase64: %v", err)
	}

	// Must be valid base64.
	decoded, err := base64.StdEncoding.DecodeString(got)
	if err != nil {
		t.Errorf("DeriveKeyBase64 output %q is not valid base64: %v", got, err)
	}

	// Must match DeriveKey output (consistency contract).
	rawKey, err := DeriveKey(masterKey)
	if err != nil {
		t.Fatalf("DeriveKey: %v", err)
	}
	if !bytes.Equal(decoded, rawKey) {
		t.Errorf("DeriveKeyBase64 decodes to different bytes than DeriveKey; want symmetric encoding")
	}
}

// TestDeriveKeyBase64_EmptyKeyReturnsError pins the input
// validation contract: DeriveKey (and therefore DeriveKeyBase64)
// MUST reject an empty master key with an error. The HMAC key
// derivation would fail on empty input; the function catches
// this and returns an error rather than producing an empty
// derived key.
func TestDeriveKeyBase64_EmptyKeyReturnsError(t *testing.T) {
	_, err := DeriveKeyBase64(nil)
	if err == nil {
		t.Error("DeriveKeyBase64(nil) = nil error; want error")
	}
	_, err = DeriveKeyBase64([]byte{})
	if err == nil {
		t.Error("DeriveKeyBase64([]byte{}) = nil error; want error")
	}
}
