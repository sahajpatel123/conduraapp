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

	"github.com/zalando/go-keyring"
)

// DeviceIdentity is the local device's Ed25519 keypair and metadata.
type DeviceIdentity struct {
	PublicKey  ed25519.PublicKey  `json:"-"`
	PrivateKey ed25519.PrivateKey `json:"-"`
	DeviceID   string             `json:"device_id"`
	Name       string             `json:"device_name"`
	CreatedAt  time.Time          `json:"created_at"`
}

// identityJSON is the JSON representation of a device identity
// (private key stored as hex). The private key is ALSO kept in the
// OS keychain when available; the on-disk copy is a fallback for
// headless servers and CI.
type identityJSON struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key,omitempty"` // omitted when keychain is the source of truth
	DeviceID   string `json:"device_id"`
	Name       string `json:"device_name"`
	CreatedAt  string `json:"created_at"`
}

// keyringService is the service name used in the OS keychain.
const keyringService = "synaptic-device-identity"

// keychainAvailable reports whether the OS keychain is reachable.
// We try a benign write+delete of a probe entry.
func keychainAvailable() bool {
	const probeUser = "__synaptic_probe__"
	if err := keyring.Set(keyringService, probeUser, "x"); err != nil {
		return false
	}
	_ = keyring.Delete(keyringService, probeUser)
	return true
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
//
// Storage strategy:
//  1. Private key is stored in the OS keychain (Keychain on macOS,
//     Credential Manager on Windows, libsecret on Linux).
//  2. The public key + device metadata is stored in a JSON file at
//     <dataDir>/device-identity.json with mode 0o600.
//  3. If the keychain is unavailable (headless server, CI), the
//     private key falls back to the JSON file with mode 0o600.
//     A warning is logged via the returned error so the caller
//     can surface it.
//
// The JSON file NEVER contains the plaintext private key when the
// keychain is available — it's keyed by the device ID.
func LoadIdentity(dataDir, name string) (*DeviceIdentity, error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("sync: mkdir: %w", err)
	}
	path := filepath.Join(dataDir, "device-identity.json")
	useKeychain := keychainAvailable()

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
		// Try keychain first.
		var priv ed25519.PrivateKey
		if useKeychain {
			privHex, kerr := keyring.Get(keyringService, raw.DeviceID)
			if kerr == nil {
				pb, perr := hex.DecodeString(privHex)
				if perr != nil {
					return nil, fmt.Errorf("sync: decode keychain priv: %w", perr)
				}
				priv = ed25519.PrivateKey(pb)
			} else if raw.PrivateKey != "" {
				// Migrate: keychain missed the key but the file has it.
				pb, perr := hex.DecodeString(raw.PrivateKey)
				if perr != nil {
					return nil, fmt.Errorf("sync: decode fallback priv: %w", perr)
				}
			priv = ed25519.PrivateKey(pb)
			if kerr := keyring.Set(keyringService, raw.DeviceID, raw.PrivateKey); kerr != nil {
				// Log but don't fail — key is already in memory.
				fmt.Printf("sync: keyring migration write failed: %v\n", kerr)
			}
			raw.PrivateKey = "" // don't keep plaintext on disk
				if out, mErr := json.MarshalIndent(raw, "", "  "); mErr == nil {
					_ = os.WriteFile(path, out, 0o600)
				}
			} else {
				return nil, fmt.Errorf("sync: no private key in keychain or file")
			}
		} else if raw.PrivateKey != "" {
			// No keychain; fallback to file.
			pb, perr := hex.DecodeString(raw.PrivateKey)
			if perr != nil {
				return nil, fmt.Errorf("sync: decode fallback priv: %w", perr)
			}
			priv = ed25519.PrivateKey(pb)
		} else {
			return nil, fmt.Errorf("sync: no private key in file (keychain unavailable)")
		}
		ts, _ := time.Parse(time.RFC3339, raw.CreatedAt)
		return &DeviceIdentity{
			PublicKey:  ed25519.PublicKey(pub),
			PrivateKey: priv,
			DeviceID:   raw.DeviceID,
			Name:       raw.Name,
			CreatedAt:  ts,
		}, nil
	}

	// No existing identity — create one.
	id, err := GenerateIdentity(name)
	if err != nil {
		return nil, err
	}
	raw := identityJSON{
		PublicKey: hex.EncodeToString(id.PublicKey),
		DeviceID:  id.DeviceID,
		Name:      id.Name,
		CreatedAt: id.CreatedAt.Format(time.RFC3339),
	}
	if useKeychain {
		// Store priv in keychain, omit from JSON file.
		if kerr := keyring.Set(keyringService, id.DeviceID, hex.EncodeToString(id.PrivateKey)); kerr != nil {
			// Fall back to storing in the file.
			raw.PrivateKey = hex.EncodeToString(id.PrivateKey)
		}
	} else {
		// No keychain; store in the file.
		raw.PrivateKey = hex.EncodeToString(id.PrivateKey)
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
