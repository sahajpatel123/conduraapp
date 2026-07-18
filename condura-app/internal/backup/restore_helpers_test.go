package backup

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

// TestValidateRestoreOptions_AllRequiredFields pins the
// happy-path contract: a fully-populated RestoreOptions with
// the documented field set passes validation. The fields
// ArchivePath + DataDir + 32-byte MasterKey are the minimum
// for any restore; other fields (Out, PreRestoreBackupPath,
// CurrentSchemaVersion, Now) have sensible defaults.
func TestValidateRestoreOptions_AllRequiredFields(t *testing.T) {
	opts := RestoreOptions{
		ArchivePath: "/var/lib/condura/backups/condura-backup.zip",
		DataDir:     "/var/lib/condura",
		MasterKey:   make([]byte, 32), // 32 zero bytes — valid length
	}
	if err := validateRestoreOptions(opts); err != nil {
		t.Errorf("validateRestoreOptions(complete) = %v, want nil", err)
	}
}

// TestValidateRestoreOptions_MissingArchivePath pins the
// first guard: empty ArchivePath must produce a clear error
// mentioning the field. Restore is called from CLI + GUI;
// a nil-error on empty ArchivePath would silently open a
// zip-slip-vulnerable path or NPE.
func TestValidateRestoreOptions_MissingArchivePath(t *testing.T) {
	opts := RestoreOptions{
		DataDir:   "/var/lib/condura",
		MasterKey: make([]byte, 32),
	}
	err := validateRestoreOptions(opts)
	if err == nil {
		t.Fatal("validateRestoreOptions(no ArchivePath) = nil; want error")
	}
	if !strings.Contains(err.Error(), "ArchivePath") {
		t.Errorf("error %q must mention 'ArchivePath'", err.Error())
	}
}

// TestValidateRestoreOptions_MissingDataDir pins the second
// guard: empty DataDir must produce a clear error mentioning
// the field. Without this, restore would target the current
// working directory and overwrite unrelated files.
func TestValidateRestoreOptions_MissingDataDir(t *testing.T) {
	opts := RestoreOptions{
		ArchivePath: "/var/lib/condura/backups/condura-backup.zip",
		MasterKey:   make([]byte, 32),
	}
	err := validateRestoreOptions(opts)
	if err == nil {
		t.Fatal("validateRestoreOptions(no DataDir) = nil; want error")
	}
	if !strings.Contains(err.Error(), "DataDir") {
		t.Errorf("error %q must mention 'DataDir'", err.Error())
	}
}

// TestValidateRestoreOptions_BadMasterKeyLength pins the
// third guard: MasterKey MUST be exactly 32 bytes (the
// AES-256 key size). Off-by-one (31 or 33) and the
// HKDF derive in DeriveKey will silently produce a wrong
// key, decrypting the archive to garbage instead of
// surfacing a clean auth failure.
//
// The error message must include both the expected length
// (32) and the actual length so the operator can debug
// from logs without re-reading source.
func TestValidateRestoreOptions_BadMasterKeyLength(t *testing.T) {
	cases := []struct {
		name string
		key  []byte
	}{
		{"empty", nil},
		{"too short (31)", make([]byte, 31)},
		{"too long (33)", make([]byte, 33)},
		{"way too long (64)", make([]byte, 64)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := RestoreOptions{
				ArchivePath: "/var/lib/condura/backups/condura-backup.zip",
				DataDir:     "/var/lib/condura",
				MasterKey:   tc.key,
			}
			err := validateRestoreOptions(opts)
			if err == nil {
				t.Fatal("validateRestoreOptions(bad MasterKey) = nil; want error")
			}
			// Error must mention "32" (the required length) and
			// the actual length, so logs are actionable.
			msg := err.Error()
			if !strings.Contains(msg, "32") {
				t.Errorf("error %q must mention '32' (required length)", msg)
			}
			if !strings.Contains(msg, "MasterKey") {
				t.Errorf("error %q must mention 'MasterKey'", msg)
			}
		})
	}
}

// TestValidateRestoreOptions_BadMasterKeyErrorIncludesActualLength
// pins the actionable-error contract: the operator-facing
// error message must include the ACTUAL length (not just the
// expected "32"), so a 31-byte key produces "got 31", not just
// "wrong length". The fmt.Errorf uses %d for this; this test
// guards against a future refactor that drops %d.
func TestValidateRestoreOptions_BadMasterKeyErrorIncludesActualLength(t *testing.T) {
	opts := RestoreOptions{
		ArchivePath: "/var/lib/condura/backups/condura-backup.zip",
		DataDir:     "/var/lib/condura",
		MasterKey:   make([]byte, 16), // 16 bytes — half the required
	}
	err := validateRestoreOptions(opts)
	if err == nil {
		t.Fatal("validateRestoreOptions(bad MasterKey) = nil; want error")
	}
	if !strings.Contains(err.Error(), "16") {
		t.Errorf("error %q must include actual length '16'", err.Error())
	}
}

// TestShortHash_LongStringReturnsPrefix pins the truncation
// contract: for a SHA-256 hex string (64 chars), shortHash
// returns the first shortHashPrefix (12) hex characters. Used
// in the InspectManifest human-readable summary and in
// fingerprint-comparison logs. A regression that returned the
// full string would blow up the line width.
func TestShortHash_LongStringReturnsPrefix(t *testing.T) {
	full := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	got := shortHash(full)
	if len(got) != shortHashPrefix {
		t.Errorf("shortHash(len=64) len = %d, want %d", len(got), shortHashPrefix)
	}
	if got != full[:shortHashPrefix] {
		t.Errorf("shortHash = %q, want %q (first %d chars)", got, full[:shortHashPrefix], shortHashPrefix)
	}
}

// TestShortHash_ExactPrefixBoundary pins the boundary: a
// string of EXACTLY shortHashPrefix length is returned in
// full (the condition is len(h) >= shortHashPrefix, not >).
// A regression to > would truncate the exact-boundary case.
func TestShortHash_ExactPrefixBoundary(t *testing.T) {
	exact := "0123456789ab" // 12 chars, equal to shortHashPrefix
	if len(exact) != shortHashPrefix {
		t.Fatalf("test setup error: exact length = %d, want %d", len(exact), shortHashPrefix)
	}
	got := shortHash(exact)
	if got != exact {
		t.Errorf("shortHash(exact-boundary) = %q, want %q (unmodified)", got, exact)
	}
}

// TestShortHash_ShortStringPassesThrough pins the defensive
// pass-through: a string SHORTER than shortHashPrefix is
// returned unchanged. This guards against a future "must be
// SHA-256 length" regression that would return "" or panic
// for non-standard hashes (hand-crafted test fixtures,
// truncated legacy data).
func TestShortHash_ShortStringPassesThrough(t *testing.T) {
	cases := []string{"", "a", "abc", "0123456789a"} // 0, 1, 3, 11 chars
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			if got := shortHash(c); got != c {
				t.Errorf("shortHash(%q) = %q, want pass-through %q", c, got, c)
			}
		})
	}
}

// TestMustDecodeBase64_ValidInput pins the happy-path
// contract: a valid base64 string decodes to the expected
// bytes. Used for decoding SHA-256 fingerprints and master
// key seeds in restore.
func TestMustDecodeBase64_ValidInput(t *testing.T) {
	original := []byte{0x00, 0x01, 0x02, 0x03, 0xff, 0xfe, 0xfd}
	encoded := base64.StdEncoding.EncodeToString(original)

	got := mustDecodeBase64(encoded)
	if !bytes.Equal(got, original) {
		t.Errorf("mustDecodeBase64 round-trip = %v, want %v", got, original)
	}
}

// TestMustDecodeBase64_InvalidInputReturnsEmpty pins the
// SILENT-FAILURE contract: 'mustDecodeBase64' discards the
// decode error and returns whatever DecodeString produced
// (which is nil/empty for invalid input). The name suggests
// panic-on-failure (Go convention for Must*), but the
// implementation does NOT panic — it silently returns empty.
//
// This is a documentation-test pin: it asserts the CURRENT
// (subtle) behavior so a future refactor either keeps it
// (and the test continues to pass) or deliberately changes
// it (and the test is updated to match the new contract).
//
// Use case in production: mustDecodeBase64 is only called on
// base64 strings that the daemon itself produced earlier
// (e.g. master key fingerprints from the secrets subsystem),
// so invalid input is not a real-world failure mode. But the
// silent-failure behavior MUST be intentional — the test name
// flags this for any future reviewer.
func TestMustDecodeBase64_InvalidInputReturnsEmpty(t *testing.T) {
	got := mustDecodeBase64("!!!not-valid-base64!!!")
	if len(got) != 0 {
		t.Errorf("mustDecodeBase64(invalid) = %v (len=%d), want empty (silent-failure contract)", got, len(got))
	}
}
