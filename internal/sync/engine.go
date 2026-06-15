package sync

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

// Engine orchestrates synchronization between local and remote stores.
// It uses the CRDT store for conflict-free merging and the Discovery
// service to find peers. Sync runs over encrypted, identity-verified
// TCP (AES-256-GCM under X25519-derived keys bound to long-term
// Ed25519 identities). Only **paired** devices are accepted — no
// auto-pair on LAN, no plaintext fallback.
type Engine struct {
	identity       *DeviceIdentity
	store          *Store
	discovery      *Discovery
	paired         *PairedSet
	logger         *slog.Logger
	discoveryPort  int
	mu             sync.Mutex
	running        bool
	stopCh         chan struct{}
	syncListener   net.Listener
}

// NewEngine creates a sync engine. If paired is nil, the engine
// accepts ANY authenticated device (insecure default; only used in
// tests). In production always pass the loaded PairedSet.
func NewEngine(identity *DeviceIdentity, store *Store, discovery *Discovery, paired *PairedSet, logger *slog.Logger) *Engine {
	port := 7667
	if discovery != nil {
		port = discovery.port
	}
	return &Engine{
		identity:      identity,
		store:         store,
		discovery:     discovery,
		paired:        paired,
		logger:        logger,
		discoveryPort: port,
		stopCh:        make(chan struct{}),
	}
}

// SetPairedSet updates the paired device store (e.g., after a
// revocation or new pairing). Safe to call while running.
func (e *Engine) SetPairedSet(ps *PairedSet) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.paired = ps
}

// PairedSet returns the current paired set (for inspection by RPCs).
func (e *Engine) PairedDevices() []*PairedDevice {
	if e.paired == nil {
		return nil
	}
	return e.paired.List()
}

// Start begins background sync operations: periodic announce, peer
// discovery, and the TCP sync listener.
func (e *Engine) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.running {
		return
	}
	e.running = true
	e.stopCh = make(chan struct{})
	go e.announceLoop()
	if e.discovery != nil {
		go func() {
			if err := e.discovery.Listen(); err != nil {
				e.logger.Debug("discovery listen stopped", "err", err)
			}
		}()
	}
	go e.serveSyncLoop()
	e.logger.Info("sync engine started",
		"device_id", e.identity.DeviceID,
		"name", e.identity.Name,
		"fingerprint", e.identity.Fingerprint(),
		"sync_port", SyncTCPPort(e.discoveryPort),
	)
}

// Stop halts background sync operations.
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.running {
		return
	}
	e.running = false
	close(e.stopCh)
	if e.syncListener != nil {
		_ = e.syncListener.Close()
		e.syncListener = nil
	}
	e.logger.Info("sync engine stopped")
}

// Status returns the current sync status.
func (e *Engine) Status() Status {
	var peerCount int
	if e.discovery != nil {
		peerCount = len(e.discovery.Peers())
	}
	var pairedCount int
	if e.paired != nil {
		pairedCount = len(e.paired.List())
	}
	return Status{
		DeviceID:      e.identity.DeviceID,
		Name:          e.identity.Name,
		Peers:         peerCount,
		Entries:       len(e.store.Entries()),
		Running:       e.running,
		SyncPort:      SyncTCPPort(e.discoveryPort),
		PairedDevices: pairedCount,
	}
}

// SyncWith performs a one-shot encrypted, identity-verified sync
// with a discovered peer. The peer must be in the paired set, or
// SyncWith returns an error.
func (e *Engine) SyncWith(peer *Peer) (int, error) {
	if peer == nil {
		return 0, fmt.Errorf("sync: peer is nil")
	}
	if e.paired != nil && !e.paired.Has(peer.DeviceID) {
		return 0, fmt.Errorf("sync: peer %s not paired — pair first via `synaptic sync pair`", peer.DeviceID)
	}
	addr, err := PeerSyncAddress(peer.Address, e.discoveryPort)
	if err != nil {
		return 0, err
	}
	e.logger.Info("sync with peer", "peer", peer.DeviceID, "addr", addr)
	n, err := DialAndSync(addr, e.identity, e.store)
	if err != nil {
		return 0, err
	}
	if e.paired != nil {
		e.paired.Touch(peer.DeviceID)
	}
	e.logger.Info("sync complete", "peer", peer.DeviceID, "merged", n)
	return n, nil
}

// PairWith creates a pairing request and returns the QR / PIN data
// the existing device needs to confirm. The new device's identity
// is not yet trusted — the existing device must add it via AddPaired.
//
// This is the FIRST half of the pairing flow. The second half is
// the user confirming the 6-digit PIN on the existing device, at
// which point the new device is added to the paired set.
func (e *Engine) PairWith(peer *Peer) (PairingToken, string, error) {
	if peer == nil {
		return "", "", fmt.Errorf("sync: peer is nil")
	}
	token, err := NewPairingToken()
	if err != nil {
		return "", "", err
	}
	pin := GeneratePairingPIN(token, peer.DeviceID, e.identity.DeviceID)
	e.logger.Info("pairing initiated",
		"new_peer", peer.DeviceID,
		"pin", pin,
	)
	return token, pin, nil
}

// ConfirmPairing is the SECOND half of the pairing flow: the user
// reads the 6-digit PIN from the new device and types it into the
// existing device's overlay. If the PIN matches, the new device is
// added to the paired set.
func (e *Engine) ConfirmPairing(peer *Peer, token PairingToken, userPIN string) error {
	if peer == nil {
		return fmt.Errorf("sync: peer is nil")
	}
	if e.paired == nil {
		return fmt.Errorf("sync: no paired set loaded")
	}
	if !VerifyPairingPIN(token, peer.DeviceID, e.identity.DeviceID, userPIN) {
		return fmt.Errorf("sync: PIN mismatch")
	}
	_, err := e.paired.Add(peer.DeviceID, peer.Name, peer.PublicKey, token, e.identity.DeviceID)
	if err != nil {
		return err
	}
	e.logger.Info("device paired", "device_id", peer.DeviceID)
	return nil
}

// RevokeDevice removes a device from the paired set and signs a
// revocation message. The revocation is returned so the caller can
// broadcast it to other paired devices.
func (e *Engine) RevokeDevice(deviceID string) (*Revocation, error) {
	if e.paired == nil {
		return nil, fmt.Errorf("sync: no paired set loaded")
	}
	if !e.paired.Has(deviceID) {
		return nil, fmt.Errorf("sync: device %s not paired", deviceID)
	}
	if err := e.paired.Remove(deviceID); err != nil {
		return nil, err
	}
	rev, err := NewRevocation(e.identity, deviceID)
	if err != nil {
		return nil, err
	}
	e.logger.Info("device revoked", "device_id", deviceID)
	return rev, nil
}

// AcceptRevocation validates a revocation signed by another paired
// device and applies it locally. Replay protection: only fresh
// revocations (within MaxRevocationAge) are honored.
func (e *Engine) AcceptRevocation(rev *Revocation) error {
	if e.paired == nil {
		return fmt.Errorf("sync: no paired set loaded")
	}
	if rev == nil {
		return fmt.Errorf("sync: nil revocation")
	}
	if !rev.IsFresh() {
		return fmt.Errorf("sync: revocation too old (revoked_at=%s)", rev.RevokedAt)
	}
	revoker := e.paired.Get(rev.RevokerDeviceID)
	if revoker == nil {
		return fmt.Errorf("sync: revoker %s is not paired", rev.RevokerDeviceID)
	}
	pubBytes, err := hexDecode(revoker.PublicKey)
	if err != nil {
		return fmt.Errorf("sync: revoker pubkey: %w", err)
	}
	if !rev.Verify(pubBytes) {
		return fmt.Errorf("sync: revocation signature invalid")
	}
	if !e.paired.Has(rev.TargetDeviceID) {
		return nil // already revoked; idempotent
	}
	if err := e.paired.Remove(rev.TargetDeviceID); err != nil {
		return err
	}
	e.logger.Info("device revoked (by peer)", "device_id", rev.TargetDeviceID, "by", rev.RevokerDeviceID)
	return nil
}

// Put stores a key-value pair in the local CRDT store.
func (e *Engine) Put(key string, value []byte) {
	e.store.Put(e.identity.DeviceID, key, value)
}

// Get retrieves a value from the local CRDT store.
func (e *Engine) Get(key string) []byte {
	entry := e.store.Get(key)
	if entry == nil {
		return nil
	}
	return entry.Value
}

// Delete removes a key from the local CRDT store.
func (e *Engine) Delete(key string) {
	e.store.Delete(e.identity.DeviceID, key)
}

// DiscoveredPeers returns the current peer list from discovery.
func (e *Engine) DiscoveredPeers() []*Peer {
	if e.discovery == nil {
		return nil
	}
	return e.discovery.Peers()
}

func (e *Engine) announceLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			if e.discovery != nil {
				if err := e.discovery.Announce(); err != nil {
					e.logger.Debug("announce failed", "err", err)
				}
			}
		}
	}
}

// serveSyncLoop accepts inbound sync connections. Each connection
// runs the encrypted handshake; if the peer is not in the paired
// set, the handshake is aborted and the connection is dropped.
func (e *Engine) serveSyncLoop() {
	addr := fmt.Sprintf(":%d", SyncTCPPort(e.discoveryPort))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		e.logger.Warn("sync listener failed", "addr", addr, "err", err)
		return
	}
	e.mu.Lock()
	e.syncListener = ln
	e.mu.Unlock()
	defer func() { _ = ln.Close() }()

	for {
		select {
		case <-e.stopCh:
			return
		default:
		}
		if tl, ok := ln.(*net.TCPListener); ok {
			_ = tl.SetDeadline(time.Now().Add(1 * time.Second))
		}
		conn, err := ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			if e.running {
				e.logger.Debug("sync accept stopped", "err", err)
			}
			return
		}
		go e.handleInbound(conn)
	}
}

// handleInbound is the per-connection sync handler. It enforces the
// paired-set policy: if the peer is not paired, the connection is
// dropped after the encrypted hello (so the peer never gets
// CRDT data). The paired check happens BEFORE the CRDT exchange.
func (e *Engine) handleInbound(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	// Run the full encrypted handshake, but plug a guard in the middle
	// to enforce the paired-set policy. We do this by wrapping the
	// store with a PairedGate that refuses to apply CRDT entries
	// from a non-paired peer.
	gatedStore := &PairedGate{inner: e.store, paired: e.paired, identity: e.identity}
	n, err := ServeSyncWithGate(conn, e.identity, gatedStore)
	if err != nil {
		e.logger.Debug("inbound sync failed", "err", err)
		return
	}
	e.logger.Info("inbound sync merged", "merged", n)
}

// Status is the current sync engine status.
type Status struct {
	DeviceID      string `json:"device_id"`
	Name          string `json:"name"`
	Peers         int    `json:"peers"`
	Entries       int    `json:"entries"`
	Running       bool   `json:"running"`
	SyncPort      int    `json:"sync_port,omitempty"`
	PairedDevices int    `json:"paired_devices"`
}
