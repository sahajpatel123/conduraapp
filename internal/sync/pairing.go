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
	"sort"
	"sync"
	"time"
)

// Pairing is the user-facing flow for adding a new device to the
// trusted set. Synaptic P2P sync does NOT auto-pair: every new
// device must complete one of these flows before its CRDT entries
// will be accepted.
//
//   1. QR code (preferred, headless) — generates a one-time pairing
//      code that the existing device scans. The code contains the
//      new device's device ID, public key, and a 6-digit PIN.
//   2. 6-digit PIN (headless) — user reads the PIN from the new
//      device's overlay and types it into the existing device.
//   3. LAN auto-discovery (no user gesture) — *not* supported by
//      design. Auto-accepting LAN peers would re-introduce the
//      plaintext-sync vulnerability we just closed. The user
//      must explicitly pair each device.
//
// The pairing token is 32 bytes of random data, encoded as hex.
// The PIN is 6 decimal digits derived from the first 3 bytes of
// SHA-256(token + newDeviceID + existingDeviceID) mod 10^6.

// PairingToken is a one-time secret exchanged during pairing.
type PairingToken string

// NewPairingToken generates a fresh 32-byte pairing token.
func NewPairingToken() (PairingToken, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("sync: pairing token: %w", err)
	}
	return PairingToken(hex.EncodeToString(b)), nil
}

// GeneratePairingPIN derives a deterministic 6-digit PIN from
// (token, newDeviceID, existingDeviceID). The same devices always
// see the same PIN, so a user can re-read it on either side if the
// QR scan fails.
func GeneratePairingPIN(token PairingToken, newDeviceID, existingDeviceID string) string {
	h := sha256.New()
	h.Write([]byte(token))
	h.Write([]byte{0})
	h.Write([]byte(newDeviceID))
	h.Write([]byte{0})
	h.Write([]byte(existingDeviceID))
	digest := h.Sum(nil)
	// Take the first 4 bytes as a uint32 and mod by 10^6.
	n := uint32(digest[0])<<24 | uint32(digest[1])<<16 | uint32(digest[2])<<8 | uint32(digest[3])
	pin := n % 1000000
	return fmt.Sprintf("%06d", pin)
}

// PairedDevice is a record of a device that has been paired.
type PairedDevice struct {
	DeviceID   string    `json:"device_id"`
	Name       string    `json:"device_name"`
	PublicKey  string    `json:"public_key"`
	PairedAt   time.Time `json:"paired_at"`
	LastSeenAt time.Time `json:"last_seen_at,omitempty"`
}

// PairedSet is the local store of trusted device IDs and their
// public keys. Persisted to <dataDir>/paired.json with mode 0o600.
type PairedSet struct {
	mu      sync.RWMutex
	devices map[string]*PairedDevice
	path    string
}

// LoadPairedSet reads the paired set from disk, or returns an
// empty set if the file doesn't exist. The parent directory is
// created with mode 0o700 so subsequent writes don't fail.
func LoadPairedSet(dataDir string) (*PairedSet, error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("sync: mkdir paired-set dir: %w", err)
	}
	path := filepath.Join(dataDir, "paired.json")
	ps := &PairedSet{
		devices: make(map[string]*PairedDevice),
		path:    path,
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ps, nil
		}
		return nil, fmt.Errorf("sync: read paired set: %w", err)
	}
	if err := json.Unmarshal(data, &ps.devices); err != nil {
		return nil, fmt.Errorf("sync: parse paired set: %w", err)
	}
	return ps, nil
}

// NewEmptyPairedSet returns an in-memory PairedSet that does not
// touch the filesystem. Used as a fail-closed default when the
// on-disk paired set is missing or corrupt — the engine still
// starts, but no peers are trusted until the user pairs them.
func NewEmptyPairedSet() *PairedSet {
	return &PairedSet{
		devices: make(map[string]*PairedDevice),
		path:    "", // empty path = no persistence
	}
}

// save writes the paired set to disk atomically (write-temp + rename).
func (ps *PairedSet) save() error {
	if err := os.MkdirAll(filepath.Dir(ps.path), 0o700); err != nil {
		return err
	}
	tmp := ps.path + ".tmp"
	out, err := json.MarshalIndent(ps.devices, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, out, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, ps.path)
}

// Add pairs a new device. Returns the PIN for the user to confirm
// on the other device.
func (ps *PairedSet) Add(deviceID, name, publicKey string, token PairingToken, existingDeviceID string) (string, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if _, exists := ps.devices[deviceID]; exists {
		return "", fmt.Errorf("sync: device %s already paired", deviceID)
	}
	ps.devices[deviceID] = &PairedDevice{
		DeviceID:  deviceID,
		Name:      name,
		PublicKey: publicKey,
		PairedAt:  time.Now().UTC(),
	}
	if err := ps.save(); err != nil {
		delete(ps.devices, deviceID)
		return "", err
	}
	return GeneratePairingPIN(token, deviceID, existingDeviceID), nil
}

// Has reports whether deviceID is in the paired set.
func (ps *PairedSet) Has(deviceID string) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	_, ok := ps.devices[deviceID]
	return ok
}

// Get returns the paired device record, or nil if not paired.
func (ps *PairedSet) Get(deviceID string) *PairedDevice {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.devices[deviceID]
}

// Touch updates the LastSeenAt of a paired device (best-effort).
func (ps *PairedSet) Touch(deviceID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if d, ok := ps.devices[deviceID]; ok {
		d.LastSeenAt = time.Now().UTC()
		_ = ps.save()
	}
}

// Remove revokes a paired device. After revocation, the device's
// CRDT entries will be rejected on the next sync.
func (ps *PairedSet) Remove(deviceID string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if _, ok := ps.devices[deviceID]; !ok {
		return fmt.Errorf("sync: device %s not paired", deviceID)
	}
	delete(ps.devices, deviceID)
	return ps.save()
}

// List returns all paired devices, sorted by DeviceID for stability.
func (ps *PairedSet) List() []*PairedDevice {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	out := make([]*PairedDevice, 0, len(ps.devices))
	for _, d := range ps.devices {
		c := *d
		out = append(out, &c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DeviceID < out[j].DeviceID })
	return out
}

// VerifyPairingPIN checks that a user-typed PIN matches the
// expected one for the (token, newDeviceID, existingDeviceID)
// triple. Used by the existing device to confirm the new device
// scanned the QR / entered the correct PIN.
func VerifyPairingPIN(token PairingToken, newDeviceID, existingDeviceID, userPIN string) bool {
	expected := GeneratePairingPIN(token, newDeviceID, existingDeviceID)
	return expected == userPIN
}

// generateRevocationSignature produces a signed revocation message
// for a paired device. Any active paired device can revoke any
// other (per spec §17.4). The signature is on (revokerID || targetID || timestamp)
// so a replay of the revocation is bounded by the timestamp.
type Revocation struct {
	RevokerDeviceID string    `json:"revoker_device_id"`
	TargetDeviceID  string    `json:"target_device_id"`
	RevokedAt       time.Time `json:"revoked_at"`
	Signature       string    `json:"signature"`
}

// NewRevocation creates a signed revocation. The revoker's private
// key signs (revoker || target || timestamp) so a future replay
// requires a re-signature.
func NewRevocation(revoker *DeviceIdentity, targetDeviceID string) (*Revocation, error) {
	now := time.Now().UTC()
	payload := revoker.DeviceID + ":" + targetDeviceID + ":" + now.Format(time.RFC3339Nano)
	sig := hex.EncodeToString(revoker.Sign([]byte(payload)))
	return &Revocation{
		RevokerDeviceID: revoker.DeviceID,
		TargetDeviceID:  targetDeviceID,
		RevokedAt:       now,
		Signature:       sig,
	}, nil
}

// Verify checks the revocation's signature against the revoker's
// public key (looked up in the paired set).
func (r *Revocation) Verify(revokerPub ed25519.PublicKey) bool {
	payload := r.RevokerDeviceID + ":" + r.TargetDeviceID + ":" + r.RevokedAt.Format(time.RFC3339Nano)
	sig, err := hex.DecodeString(r.Signature)
	if err != nil {
		return false
	}
	return ed25519.Verify(revokerPub, []byte(payload), sig)
}

// MaxRevocationAge is how long a revocation message is considered
// valid. After this, the receiving device ignores it (defense
// against replay of old revocations).
const MaxRevocationAge = 24 * time.Hour

// IsFresh returns true if the revocation is recent enough to be
// honored.
func (r *Revocation) IsFresh() bool {
	return time.Since(r.RevokedAt) <= MaxRevocationAge
}
