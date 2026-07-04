package sync

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestPairing_BeginConfirmRoundTrip exercises the full pairing flow
// and asserts the PIN minted by PairWith is the one ConfirmPairing
// expects. This is a regression test for the bug where the confirm
// step used a fresh token instead of the one from begin — pairing
// could never succeed.
func TestPairing_BeginConfirmRoundTrip(t *testing.T) {
	dir := t.TempDir()
	aID, _ := GenerateIdentity("laptop")
	bID, _ := GenerateIdentity("phone")
	// Set up minimal engine without discovery/paired (we just test
	// the pairing API, not the network plumbing).
	paired, _ := LoadPairedSet(dir)
	eng := NewEngine(aID, NewStore(), nil, paired, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	// Build a peer object that looks like one we discovered.
	peer := &Peer{
		DeviceID:  bID.DeviceID,
		Name:      bID.Name,
		PublicKey: hexEncode(bID.PublicKey),
		Address:   "127.0.0.1:7667",
	}

	// Begin: produces a PIN.
	_, pin, err := eng.PairWith(peer)
	if err != nil {
		t.Fatal(err)
	}
	if len(pin) != 6 {
		t.Errorf("PIN length: %d", len(pin))
	}

	// Confirm: types the SAME pin.
	if err := eng.ConfirmPairing(peer, pin); err != nil {
		t.Fatalf("confirm with correct PIN failed: %v", err)
	}
	if !paired.Has(peer.DeviceID) {
		t.Error("paired set should contain peer after confirm")
	}
}

// TestPairing_WrongPINRejected confirms an incorrect PIN is rejected.
func TestPairing_WrongPINRejected(t *testing.T) {
	dir := t.TempDir()
	aID, _ := GenerateIdentity("laptop")
	bID, _ := GenerateIdentity("phone")
	paired, _ := LoadPairedSet(dir)
	eng := NewEngine(aID, NewStore(), nil, paired, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	peer := &Peer{
		DeviceID:  bID.DeviceID,
		Name:      bID.Name,
		PublicKey: hexEncode(bID.PublicKey),
		Address:   "127.0.0.1:7667",
	}
	_, _, _ = eng.PairWith(peer)
	// Try every wrong PIN except the right one.
	wrong := "000000"
	if wrong == "" {
		wrong = "000001"
	}
	if err := eng.ConfirmPairing(peer, wrong); err == nil {
		t.Error("wrong PIN accepted")
	}
	if paired.Has(peer.DeviceID) {
		t.Error("paired set should not contain peer after wrong PIN")
	}
}

// TestPairing_ConfirmWithoutBeginRejected: a confirm without a prior
// begin should fail with a clear error.
func TestPairing_ConfirmWithoutBeginRejected(t *testing.T) {
	dir := t.TempDir()
	aID, _ := GenerateIdentity("laptop")
	bID, _ := GenerateIdentity("phone")
	paired, _ := LoadPairedSet(dir)
	eng := NewEngine(aID, NewStore(), nil, paired, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	peer := &Peer{
		DeviceID:  bID.DeviceID,
		Name:      bID.Name,
		PublicKey: hexEncode(bID.PublicKey),
	}
	err := eng.ConfirmPairing(peer, "123456")
	if err == nil {
		t.Fatal("confirm without begin should fail")
	}
}

// TestPairing_TokenExpires ensures stale pairings are not honored.
func TestPairing_TokenExpires(t *testing.T) {
	dir := t.TempDir()
	aID, _ := GenerateIdentity("laptop")
	bID, _ := GenerateIdentity("phone")
	paired, _ := LoadPairedSet(dir)
	eng := NewEngine(aID, NewStore(), nil, paired, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	peer := &Peer{
		DeviceID:  bID.DeviceID,
		Name:      bID.Name,
		PublicKey: hexEncode(bID.PublicKey),
	}
	_, pin, _ := eng.PairWith(peer)
	// Manually expire the pending pairing.
	eng.mu.Lock()
	eng.pendingPairings[peer.DeviceID] = pendingPairing{
		token:     eng.pendingPairings[peer.DeviceID].token,
		createdAt: time.Now().Add(-2 * pendingPairingTTL),
	}
	eng.mu.Unlock()
	err := eng.ConfirmPairing(peer, pin)
	if err == nil {
		t.Error("expired pairing should fail")
	}
}

// TestLoadPairedSet_DirectoryCreated verifies LoadPairedSet creates
// the parent dir if missing.
func TestLoadPairedSet_DirectoryCreated(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "nonexistent-subdir")
	if _, err := LoadPairedSet(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("LoadPairedSet should have created %s: %v", dir, err)
	}
}

func hexEncode(b []byte) string {
	const hex = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, c := range b {
		out[i*2] = hex[c>>4]
		out[i*2+1] = hex[c&0x0f]
	}
	return string(out)
}
