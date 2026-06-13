package sync

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DeviceIdentity is the local device's Ed25519 keypair and metadata.
type DeviceIdentity struct {
	PublicKey  ed25519.PublicKey  `json:"-"`
	PrivateKey ed25519.PrivateKey `json:"-"`
	DeviceID   string            `json:"device_id"`
	Name       string            `json:"name"`
	CreatedAt  time.Time         `json:"created_at"`
}

// identityJSON is the JSON representation of a device identity
// (private key stored as hex).
type identityJSON struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	DeviceID   string `json:"device_id"`
	Name       string `json:"name"`
	CreatedAt  string `json:"created_at"`
}

// GenerateIdentity creates a new Ed25519 keypair.
func GenerateIdentity(name string) (*DeviceIdentity, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("sync: generate key: %w", err)
	}
	return &DeviceIdentity{
		PublicKey:  pub,
		PrivateKey: priv,
		DeviceID:   hex.EncodeToString(pub),
		Name:       name,
		CreatedAt:  time.Now().UTC(),
	}, nil
}

// Sign signs data with the device's private key.
func (d *DeviceIdentity) Sign(data []byte) []byte {
	return ed25519.Sign(d.PrivateKey, data)
}

// VerifySignature verifies a signature against a public key.
func VerifySignature(pub ed25519.PublicKey, data, sig []byte) bool {
	return ed25519.Verify(pub, data, sig)
}

// Fingerprint returns a short hex fingerprint of the public key.
func (d *DeviceIdentity) Fingerprint() string {
	sum := sha256.Sum256(d.PublicKey)
	return hex.EncodeToString(sum[:8])
}

// LoadIdentity loads or creates a device identity in the data directory.
func LoadIdentity(dataDir, name string) (*DeviceIdentity, error) {
	path := filepath.Join(dataDir, "device-identity.json")
	data, err := os.ReadFile(path)
	if err == nil {
		var raw identityJSON
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("sync: parse identity: %w", err)
		}
		pub, err := hex.DecodeString(raw.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("sync: decode public key: %w", err)
		}
		priv, err := hex.DecodeString(raw.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("sync: decode private key: %w", err)
		}
		ts, _ := time.Parse(time.RFC3339, raw.CreatedAt)
		return &DeviceIdentity{
			PublicKey:  ed25519.PublicKey(pub),
			PrivateKey: ed25519.PrivateKey(priv),
			DeviceID:   raw.DeviceID,
			Name:       raw.Name,
			CreatedAt:  ts,
		}, nil
	}
	id, err := GenerateIdentity(name)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("sync: mkdir: %w", err)
	}
	raw := identityJSON{
		PublicKey:  hex.EncodeToString(id.PublicKey),
		PrivateKey: hex.EncodeToString(id.PrivateKey),
		DeviceID:   id.DeviceID,
		Name:       id.Name,
		CreatedAt:  id.CreatedAt.Format(time.RFC3339),
	}
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		return nil, fmt.Errorf("sync: write identity: %w", err)
	}
	return id, nil
}
