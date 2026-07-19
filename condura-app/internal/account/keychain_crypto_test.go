package account

import (
	"bytes"
	"testing"
)

// TestEncryptDecryptRoundtrip pins the basic correctness:
// encrypt(plaintext, key) followed by decrypt(ciphertext, key)
// MUST return the original plaintext. This is the security
// foundation — every token stored in the keychain flows through
// this round-trip.
func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := make([]byte, 32) // AES-256
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := []byte("oauth-token-secret-12345")

	ct, err := encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if bytes.Equal(ct, plaintext) {
		t.Fatal("ciphertext equals plaintext; encryption is a no-op")
	}

	got, err := decrypt(ct, key)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("roundtrip = %q, want %q", got, plaintext)
	}
}

// TestEncryptUsesRandomNonce pins the nonce-randomness contract:
// encrypt(plaintext, key) called twice with the SAME plaintext and
// key MUST produce different ciphertexts (because the nonce is
// random per call). A regression that used a fixed nonce would
// leak information about repeated encryptions of the same data.
func TestEncryptUsesRandomNonce(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte("same-plaintext")

	ct1, _ := encrypt(plaintext, key)
	ct2, _ := encrypt(plaintext, key)

	if bytes.Equal(ct1, ct2) {
		t.Errorf("two encryptions of the same plaintext produced identical ciphertexts; nonce is not random")
	}
	// The first 12 bytes are the nonce — they MUST differ.
	if bytes.Equal(ct1[:12], ct2[:12]) {
		t.Errorf("two encryptions produced identical nonces; nonce is not random")
	}
}

// TestEncryptEmptyPlaintext pins the empty-input contract:
// encrypt([]byte{}, key) MUST succeed (return a valid ciphertext)
// and decrypt the result MUST return the empty plaintext. A
// regression that errored on empty would break the empty-token
// case.
func TestEncryptEmptyPlaintext(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte{}

	ct, err := encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt(empty): %v", err)
	}
	if len(ct) <= 12 {
		t.Errorf("ciphertext for empty plaintext = %d bytes; want > 12 (12 nonce + GCM tag)", len(ct))
	}

	got, err := decrypt(ct, key)
	if err != nil {
		t.Fatalf("decrypt(empty ct): %v", err)
	}
	if len(got) != 0 {
		t.Errorf("decrypted empty plaintext = %q; want empty", got)
	}
}

// TestDecryptTamperedCiphertextFails pins the GCM auth-tag
// verification contract: modifying even one byte of the
// ciphertext (or nonce) MUST cause decrypt to fail. This is the
// AES-GCM integrity guarantee — without it, an attacker could
// silently modify tokens on disk.
func TestDecryptTamperedCiphertextFails(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte("oauth-token")

	ct, err := encrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}

	// Flip one bit in the ciphertext (after the nonce).
	ct[15] ^= 0x01

	if _, err := decrypt(ct, key); err == nil {
		t.Error("decrypt of tampered ciphertext returned nil error; GCM tag verification broken")
	}
}

// TestDecryptTooShortCiphertextReturnsError pins the input
// validation contract: decrypt MUST reject ciphertexts shorter
// than the nonce size (12 bytes for AES-GCM). A regression that
// panicked on `ct[:nonceSize]` would crash the daemon on a
// corrupted-token read.
func TestDecryptTooShortCiphertextReturnsError(t *testing.T) {
	key := make([]byte, 32)

	for _, ct := range [][]byte{
		nil,
		{},
		[]byte{0x01, 0x02, 0x03}, // 3 bytes < 12
		make([]byte, 11),         // 11 bytes < 12
	} {
		_, err := decrypt(ct, key)
		if err == nil {
			t.Errorf("decrypt(%d bytes) returned nil error; want error for too-short ciphertext",
				len(ct))
		}
	}
}

// TestEncryptWrongKeySizeReturnsError pins the key-size
// validation contract: encrypt MUST return an error for keys
// that aren't 16, 24, or 32 bytes (the AES-128/192/256 sizes).
// A regression that silently truncated/zero-padded the key would
// silently weaken encryption.
func TestEncryptWrongKeySizeReturnsError(t *testing.T) {
	plaintext := []byte("data")

	// 8 bytes — too short
	if _, err := encrypt(plaintext, make([]byte, 8)); err == nil {
		t.Error("encrypt with 8-byte key returned nil error; want 'invalid key size'")
	}
	// 20 bytes — not a valid AES key size
	if _, err := encrypt(plaintext, make([]byte, 20)); err == nil {
		t.Error("encrypt with 20-byte key returned nil error; want 'invalid key size'")
	}
	// 40 bytes — too long
	if _, err := encrypt(plaintext, make([]byte, 40)); err == nil {
		t.Error("encrypt with 40-byte key returned nil error; want 'invalid key size'")
	}
}
