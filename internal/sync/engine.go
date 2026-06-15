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
// service to find peers. Sync runs over signed TCP (discovery port + 1).
type Engine struct {
	identity       *DeviceIdentity
	store          *Store
	discovery      *Discovery
	logger         *slog.Logger
	discoveryPort  int
	mu             sync.Mutex
	running        bool
	stopCh         chan struct{}
	syncListener   net.Listener
}

// NewEngine creates a sync engine.
func NewEngine(identity *DeviceIdentity, store *Store, discovery *Discovery, logger *slog.Logger) *Engine {
	port := 7667
	if discovery != nil {
		port = discovery.port
	}
	return &Engine{
		identity:      identity,
		store:         store,
		discovery:     discovery,
		logger:        logger,
		discoveryPort: port,
		stopCh:        make(chan struct{}),
	}
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
	peers := e.discovery.Peers()
	return Status{
		DeviceID: e.identity.DeviceID,
		Name:     e.identity.Name,
		Peers:    len(peers),
		Entries:  len(e.store.Entries()),
		Running:  e.running,
		SyncPort: SyncTCPPort(e.discoveryPort),
	}
}

// SyncWith performs a one-shot signed TCP sync with a discovered peer.
func (e *Engine) SyncWith(peer *Peer) (int, error) {
	if peer == nil {
		return 0, fmt.Errorf("sync: peer is nil")
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
	e.logger.Info("sync complete", "peer", peer.DeviceID, "merged", n)
	return n, nil
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
			if err := e.discovery.Announce(); err != nil {
				e.logger.Debug("announce failed", "err", err)
			}
		}
	}
}

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
		go func(c net.Conn) {
			if _, err := ServeSync(c, e.identity, e.store); err != nil {
				e.logger.Debug("inbound sync failed", "err", err)
			}
		}(conn)
	}
}

// Status is the current sync engine status.
type Status struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
	Peers    int    `json:"peers"`
	Entries  int    `json:"entries"`
	Running  bool   `json:"running"`
	SyncPort int    `json:"sync_port,omitempty"`
}
