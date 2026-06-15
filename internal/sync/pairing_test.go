package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPairedSet_AddRemove verifies basic add/list/remove.
func TestPairedSet_AddRemove(t *testing.T) {
	dir := t.TempDir()
	ps, err := LoadPairedSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	token, _ := NewPairingToken()
	pin, err := ps.Add("dev-1", "Laptop", "pubkey-hex", token, "self-id")
	if err != nil {
		t.Fatal(err)
	}
	if len(pin) != 6 {
		t.Errorf("PIN: %q (len %d)", pin, len(pin))
	}
	if !ps.Has("dev-1") {
		t.Error("Has(dev-1) should be true")
	}
	if ps.Has("dev-2") {
		t.Error("Has(dev-2) should be false")
	}
	if got := len(ps.List()); got != 1 {
		t.Errorf("list: got %d, want 1", got)
	}
	if err := ps.Remove("dev-1"); err != nil {
		t.Fatal(err)
	}
	if ps.Has("dev-1") {
		t.Error("Has(dev-1) should be false after remove")
	}
	// Idempotent remove
	if err := ps.Remove("dev-1"); err == nil {
		t.Error("Remove of non-existent should fail")
	}
}

// TestPairedSet_Persistence verifies the set survives a reload.
func TestPairedSet_Persistence(t *testing.T) {
	dir := t.TempDir()
	ps, _ := LoadPairedSet(dir)
	token, _ := NewPairingToken()
	if _, err := ps.Add("dev-A", "A", "pubA", token, "self"); err != nil {
		t.Fatal(err)
	}
	if _, err := ps.Add("dev-B", "B", "pubB", token, "self"); err != nil {
		t.Fatal(err)
	}
	// Reload from disk.
	ps2, err := LoadPairedSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !ps2.Has("dev-A") || !ps2.Has("dev-B") {
		t.Error("paired set did not persist")
	}
}

// TestGeneratePairingPIN_Deterministic verifies the PIN is
// reproducible for the same triple.
func TestGeneratePairingPIN_Deterministic(t *testing.T) {
	token, _ := NewPairingToken()
	a := GeneratePairingPIN(token, "new-dev", "self")
	b := GeneratePairingPIN(token, "new-dev", "self")
	if a != b {
		t.Errorf("PIN not deterministic: %q vs %q", a, b)
	}
	// Different token -> different PIN
	token2, _ := NewPairingToken()
	c := GeneratePairingPIN(token2, "new-dev", "self")
	if a == c {
		t.Error("different token should produce different PIN")
	}
	// Different order -> different PIN
	d := GeneratePairingPIN(token, "self", "new-dev")
	if a == d {
		t.Error("swapped args should produce different PIN")
	}
}

// TestVerifyPairingPIN confirms the verify helper matches the generator.
func TestVerifyPairingPIN(t *testing.T) {
	token, _ := NewPairingToken()
	pin := GeneratePairingPIN(token, "new-dev", "self")
	if !VerifyPairingPIN(token, "new-dev", "self", pin) {
		t.Error("correct PIN rejected")
	}
	if VerifyPairingPIN(token, "new-dev", "self", "000000") {
		// Possible to be lucky; try a few wrong pins to be safe
		for _, bad := range []string{"000001", "999999", "000010", "100000", "111111"} {
			if bad != pin && VerifyPairingPIN(token, "new-dev", "self", bad) {
				t.Errorf("wrong PIN %q accepted", bad)
			}
		}
	}
}

// TestRevocation_SignVerify ensures the revocation signature
// is verified correctly.
func TestRevocation_SignVerify(t *testing.T) {
	revoker, _ := GenerateIdentity("revoker")
	rev, err := NewRevocation(revoker, "target-id")
	if err != nil {
		t.Fatal(err)
	}
	if !rev.IsFresh() {
		t.Error("fresh revocation should pass IsFresh")
	}
	if !rev.Verify(revoker.PublicKey) {
		t.Error("valid signature rejected")
	}
	// Wrong key
	other, _ := GenerateIdentity("other")
	if rev.Verify(other.PublicKey) {
		t.Error("invalid signature accepted with wrong key")
	}
	// Tampered target
	tampered := *rev
	tampered.TargetDeviceID = "different-id"
	if tampered.Verify(revoker.PublicKey) {
		t.Error("tampered target accepted")
	}
}

// TestPairedSet_AtomicSave checks the save is atomic (file replaced
// in one shot, no half-written state).
func TestPairedSet_AtomicSave(t *testing.T) {
	dir := t.TempDir()
	ps, _ := LoadPairedSet(dir)
	token, _ := NewPairingToken()
	if _, err := ps.Add("dev-1", "Laptop", "pubkey", token, "self"); err != nil {
		t.Fatal(err)
	}
	// File should exist and be parseable.
	path := filepath.Join(dir, "paired.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "dev-1") {
		t.Error("paired.json does not contain device ID")
	}
	// No temp file should be left behind.
	if _, err := os.Stat(path + ".tmp"); err == nil {
		t.Error("temp file left behind after save")
	}
}
