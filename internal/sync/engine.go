package sync

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Engine orchestrates synchronization between local and remote stores.
// It uses the CRDT store for conflict-free merging and the Discovery
// service to find peers. The actual transport (libp2p, TCP, etc.) is
// abstracted behind the Transport interface.
type Engine struct {
	identity  *DeviceIdentity
	store     *Store
	discovery *Discovery
	logger    *slog.Logger
	mu        sync.Mutex
	running   bool
	stopCh    chan struct{}
}

// NewEngine creates a sync engine.
func NewEngine(identity *DeviceIdentity, store *Store, discovery *Discovery, logger *slog.Logger) *Engine {
	return &Engine{
		identity:  identity,
		store:     store,
		discovery: discovery,
		logger:    logger,
		stopCh:    make(chan struct{}),
	}
}

// Start begins background sync operations: periodic announce and
// peer discovery. The actual sync protocol is a placeholder for
// Phase 12D (P2P transport not yet integrated).
func (e *Engine) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.running {
		return
	}
	e.running = true
	go e.announceLoop()
	e.logger.Info("sync engine started",
		"device_id", e.identity.DeviceID,
		"name", e.identity.Name,
		"fingerprint", e.identity.Fingerprint(),
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
	e.logger.Info("sync engine stopped")
}

// Status returns the current sync status.
func (e *Engine) Status() Status {
	peers := e.discovery.Peers()
	return Status{
		DeviceID: e.identity.DeviceID,
		Name:     e.identity.Name,
		Peers:    len(peers),
		Entries:  len(e.store.Entries()),
		Running:  e.running,
	}
}

// SyncWith performs a one-shot sync with a specific peer. This is
// a placeholder that will be replaced with real P2P transport.
func (e *Engine) SyncWith(peer *Peer) error {
	e.logger.Info("sync with peer (placeholder)", "peer", peer.DeviceID, "addr", peer.Address)
	// TODO: implement real sync over Noise XX + libp2p.
	return fmt.Errorf("sync: P2P transport not yet implemented")
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

func (e *Engine) announceLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			if err := e.discovery.Announce(); err != nil {
				e.logger.Debug("announce failed", "err", err)
			}
		}
	}
}

// Status is the current sync engine status.
type Status struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
	Peers    int    `json:"peers"`
	Entries  int    `json:"entries"`
	Running  bool   `json:"running"`
}
