package sync

import (
	"crypto/ed25519"
	"crypto/rand"
	"strings"
	"testing"
	"time"
)

// newTestPairedGate returns a PairedGate backed by a fresh CRDT store
// and a disk-backed PairedSet rooted in t.TempDir() so Add() can
// persist. identity is the local device whose ID gets the self-pass
// in Merge.
func newTestPairedGate(t *testing.T) (*PairedGate, *DeviceIdentity) {
	t.Helper()
	id, err := GenerateIdentity("test-device")
	if err != nil {
		t.Fatalf("GenerateIdentity: %v", err)
	}
	ps, err := LoadPairedSet(t.TempDir())
	if err != nil {
		t.Fatalf("LoadPairedSet: %v", err)
	}
	return &PairedGate{
		inner:    NewStore(),
		paired:   ps,
		identity: id,
	}, id
}

// makeEntry builds an Entry with a non-trivial vector clock so the
// underlying Store.Merge treats it as new (no local copy yet, so the
// happens-before branch never fires on the first apply).
func makeEntry(deviceID, key string, value []byte, counter int64) *Entry {
	vc := VectorClock{deviceID: counter}
	return &Entry{
		Key:       key,
		Value:     value,
		Version:   vc,
		DeviceID:  deviceID,
		Timestamp: time.Now().UTC(),
	}
}

// TestPairedGate_PutGetDelete_Entries pins the local-side pass-through
// contract: Put/Get/Delete/Entries must all reach the inner Store
// regardless of the paired-set state. This is the local path — no
// peer is involved — so the security gate must not interfere.
func TestPairedGate_PutGetDelete_Entries(t *testing.T) {
	g, _ := newTestPairedGate(t)
	deviceID := "local-device"

	g.Put(deviceID, "k1", []byte("v1"))
	if got := g.Get("k1"); got == nil || string(got.Value) != "v1" {
		t.Fatalf("Get(k1) = %+v, want value=v1", got)
	}

	g.Put(deviceID, "k2", []byte("v2"))
	entries := g.Entries()
	if len(entries) != 2 {
		t.Fatalf("Entries() len = %d, want 2", len(entries))
	}

	g.Delete(deviceID, "k1")
	if got := g.Get("k1"); got != nil {
		t.Fatalf("Get(k1) after Delete = %+v, want nil", got)
	}
	if got := g.Get("k2"); got == nil || string(got.Value) != "v2" {
		t.Fatalf("Get(k2) after Delete(k1) = %+v, want value=v2", got)
	}
}

// TestPairedGate_Merge_AllowsPairedDevice pins the happy path: an
// entry from a device that IS in the paired set must be applied to
// the inner store. This is the only path that lets legitimate peers
// land writes — without it, no sync would work.
func TestPairedGate_Merge_AllowsPairedDevice(t *testing.T) {
	g, local := newTestPairedGate(t)
	peerID := "peer-device-1"

	// Pair peer1.
	token, err := NewPairingToken()
	if err != nil {
		t.Fatalf("NewPairingToken: %v", err)
	}
	if _, err := g.paired.Add(peerID, "peer1", "fake-pubkey", token, local.DeviceID); err != nil {
		t.Fatalf("PairedSet.Add: %v", err)
	}

	// Peer sends an entry.
	entry := makeEntry(peerID, "shared-key", []byte("peer-value"), 1)
	if !g.Merge(peerID, entry) {
		t.Fatal("Merge from paired device returned false; want true")
	}
	got := g.Get("shared-key")
	if got == nil || string(got.Value) != "peer-value" {
		t.Fatalf("Get after Merge = %+v, want value=peer-value", got)
	}
}

// TestPairedGate_Merge_RejectsUnpairedDevice is the security-boundary
// test. An entry from a device NOT in the paired set must be
// rejected — Merge returns false, no entry is applied. Without this
// gate, any device that completes the encrypted handshake could
// inject entries into the CRDT.
func TestPairedGate_Merge_RejectsUnpairedDevice(t *testing.T) {
	g, _ := newTestPairedGate(t)
	attackerID := "unpaired-attacker"

	entry := makeEntry(attackerID, "injected", []byte("malicious"), 1)
	if g.Merge(attackerID, entry) {
		t.Fatal("Merge from unpaired device returned true; want false (gate must reject)")
	}
	if got := g.Get("injected"); got != nil {
		t.Fatalf("Get after rejected Merge = %+v, want nil (gate failed to block)", got)
	}
}

// TestPairedGate_Merge_AllowsSelfEntry pins the self-pass shortcut:
// the local device's own ID bypasses the gate even though the local
// device is not in its own paired set. This is correct behavior —
// local writes don't go through the peer-attestation path. A future
// refactor that removes the `remoteDeviceID != g.identity.DeviceID`
// guard would let unpaired peers inject as long as they spoof the
// local device ID, which is exactly what this shortcut defends.
func TestPairedGate_Merge_AllowsSelfEntry(t *testing.T) {
	g, local := newTestPairedGate(t)

	entry := makeEntry(local.DeviceID, "self-key", []byte("self-value"), 1)
	if !g.Merge(local.DeviceID, entry) {
		t.Fatal("Merge with self DeviceID returned false; want true (self-pass)")
	}
	if got := g.Get("self-key"); got == nil || string(got.Value) != "self-value" {
		t.Fatalf("Get after self Merge = %+v, want value=self-value", got)
	}
}

// TestPairedGate_Merge_NilPairedSet pins the no-gate fallback: when
// the PairedSet pointer is nil, Merge must allow every entry. This
// matches the construction-time contract — a nil paired set means
// "no gate configured" rather than "block everything". Blocking
// everything would break the engine before the first pairing is set
// up, which is the exact path the empty-paired-set docs warn about.
func TestPairedGate_Merge_NilPairedSet(t *testing.T) {
	g, local := newTestPairedGate(t)
	g.paired = nil // explicitly disable gate

	for _, remoteID := range []string{"any-peer", "another-peer", local.DeviceID} {
		entry := makeEntry(remoteID, "k-"+remoteID, []byte("v"), 1)
		if !g.Merge(remoteID, entry) {
			t.Fatalf("Merge(%q) with nil PairedSet returned false; want true", remoteID)
		}
	}
}

// TestHexDecode_Valid pins the happy path: a valid hex string
// decodes to its raw bytes. Without coverage, a regression that
// swapped hex.DecodeString for base64 would only surface deep in
// the pairing flow.
func TestHexDecode_Valid(t *testing.T) {
	got, err := hexDecode("48656c6c6f") // "Hello"
	if err != nil {
		t.Fatalf("hexDecode(valid): %v", err)
	}
	if string(got) != "Hello" {
		t.Fatalf("hexDecode = %q, want %q", got, "Hello")
	}
}

// TestHexDecode_Invalid pins the error contract: a non-hex input
// returns a wrapped error mentioning "hex" so callers can diagnose
// the failure mode without exposing the underlying encoding error.
func TestHexDecode_Invalid(t *testing.T) {
	_, err := hexDecode("not-hex!!")
	if err == nil {
		t.Fatal("hexDecode(invalid): expected error, got nil")
	}
	if !strings.Contains(err.Error(), "hex") {
		t.Fatalf("error %q should mention 'hex'", err.Error())
	}
}

// TestAcceptRevocationSignature_Valid pins the happy path: a payload
// signed with the matching private key must verify. Without this,
// a future change that flips a sign-check negation would silently
// accept every revocation (the worse failure mode — an attacker
// gets to forge pair-termination events).
func TestAcceptRevocationSignature_Valid(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	payload := []byte(`{"device":"peer-1","reason":"user_revoked"}`)
	sig := ed25519.Sign(priv, payload)
	if !AcceptRevocationSignature(pub, payload, sig) {
		t.Fatal("AcceptRevocationSignature returned false on valid signature; want true")
	}
}

// TestAcceptRevocationSignature_TamperedPayload pins the integrity
// check: flipping any byte in the payload must invalidate the
// signature. Ed25519 is malleable against chosen-plaintext attacks
// only if the signature scheme itself is broken, so a regression
// here would point at a real bug in the verification path.
func TestAcceptRevocationSignature_TamperedPayload(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	original := []byte(`{"device":"peer-1","reason":"user_revoked"}`)
	sig := ed25519.Sign(priv, original)

	tampered := []byte(`{"device":"peer-1","reason":"user_paired"}`) // differs in middle word
	if string(tampered) == string(original) {
		t.Fatal("test fixture error: tampered payload equals original")
	}
	if AcceptRevocationSignature(pub, tampered, sig) {
		t.Fatal("AcceptRevocationSignature returned true on tampered payload; want false")
	}
}

// TestAcceptRevocationSignature_WrongPublicKey pins the binding
// between signature and public key. A signature produced by key A
// must NOT verify against key B, even if both are valid ed25519
// keys. A regression here would let any device accept any other
// device's revocation — the worst possible outcome for a pair-
// termination event.
func TestAcceptRevocationSignature_WrongPublicKey(t *testing.T) {
	pubA, privA, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey A: %v", err)
	}
	pubB, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey B: %v", err)
	}
	payload := []byte("revocation-event")
	sig := ed25519.Sign(privA, payload)

	if AcceptRevocationSignature(pubB, payload, sig) {
		t.Fatal("AcceptRevocationSignature returned true with mismatched public key; want false")
	}
	// Sanity: it must still verify against the original key.
	if !AcceptRevocationSignature(pubA, payload, sig) {
		t.Fatal("sanity: AcceptRevocationSignature returned false on the correct key")
	}
}

// TestAcceptRevocationSignature_EmptySignature pins the all-zeros
// signature rejection. A revocation with a zero-length or all-zero
// signature is always invalid; if it were ever accepted, an attacker
// could broadcast empty revocations to trigger pair termination.
func TestAcceptRevocationSignature_EmptySignature(t *testing.T) {
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	payload := []byte("revocation-event")
	if AcceptRevocationSignature(pub, payload, nil) {
		t.Fatal("AcceptRevocationSignature returned true on nil signature; want false")
	}
	if AcceptRevocationSignature(pub, payload, make([]byte, ed25519.SignatureSize)) {
		t.Fatal("AcceptRevocationSignature returned true on zero-filled signature; want false")
	}
}
