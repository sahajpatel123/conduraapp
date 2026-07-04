package sync

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"net"
)

// PairedGate wraps a CRDT Store and rejects entries from unpaired
// devices. The check is done AFTER the encrypted handshake (so the
// peer proves its identity) but BEFORE the CRDT exchange (so no
// data is leaked to unpaired peers).
type PairedGate struct {
	inner    *Store
	paired   *PairedSet
	identity *DeviceIdentity
}

// Put is a local put (no peer involved); always allowed.
func (g *PairedGate) Put(deviceID, key string, value []byte) {
	g.inner.Put(deviceID, key, value)
}

// Get reads the inner store.
func (g *PairedGate) Get(key string) *Entry {
	return g.inner.Get(key)
}

// Delete is local; always allowed.
func (g *PairedGate) Delete(deviceID, key string) {
	g.inner.Delete(deviceID, key)
}

// Merge is called on the responder side with the remote peer's
// entries. If the remote device is not in the paired set, the
// merge is rejected (returns false) and no entries are applied.
//
// We have to know the remote DeviceID before we can enforce this.
// Since the encrypted hello reveals it (after step 5 of the
// protocol), the gate check happens between the hello exchange
// and the CRDT exchange. The actual `ExchangeEntries` is called
// by ServeSyncWithGate (in transport.go) which extracts the
// remote device ID from the hello.
func (g *PairedGate) Merge(remoteDeviceID string, entry *Entry) bool {
	if g.paired != nil && remoteDeviceID != g.identity.DeviceID {
		if !g.paired.Has(remoteDeviceID) {
			return false
		}
	}
	return g.inner.Merge(entry)
}

// Entries returns all current entries (used to send to peers).
func (g *PairedGate) Entries() []*Entry {
	return g.inner.Entries()
}

// hexDecode is a thin wrapper used by pairing/revocation code.
func hexDecode(s string) ([]byte, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("sync: hex: %w", err)
	}
	return b, nil
}

// ServeSyncWithGate runs the inbound sync with a paired-set gate.
// The remote device ID is extracted from the encrypted hello; if
// it's not in the paired set, the connection is closed without
// any CRDT data being sent.
func ServeSyncWithGate(conn net.Conn, local *DeviceIdentity, gate *PairedGate) (int, error) {
	defer func() { _ = conn.Close() }()
	return ExchangeEntriesGated(conn, local, gate)
}

// AcceptRevocationSignature verifies that the revocation signature
// matches the given Ed25519 public key. Public for testing.
func AcceptRevocationSignature(pub ed25519.PublicKey, payload []byte, sig []byte) bool {
	return ed25519.Verify(pub, payload, sig)
}
