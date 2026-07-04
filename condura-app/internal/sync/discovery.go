package sync

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// Peer represents a discovered device on the local network.
type Peer struct {
	DeviceID    string    `json:"device_id"`
	Name        string    `json:"name"`
	PublicKey   string    `json:"public_key"`
	Address     string    `json:"address"`
	LastSeen    time.Time `json:"last_seen"`
	Fingerprint string    `json:"fingerprint"`
}

// Discovery announces and discovers peers on the local network using UDP broadcast.
type Discovery struct {
	device   *DeviceIdentity
	port     int
	mu       sync.RWMutex
	peers    map[string]*Peer
	onUpdate func(*Peer)
}

// NewDiscovery creates a LAN discovery service.
func NewDiscovery(device *DeviceIdentity, port int) *Discovery {
	return &Discovery{
		device: device,
		port:   port,
		peers:  make(map[string]*Peer),
	}
}

// SetOnUpdate sets the callback invoked when a new or updated peer is discovered.
func (d *Discovery) SetOnUpdate(fn func(*Peer)) {
	d.onUpdate = fn
}

// Announce broadcasts the device's presence on the local network.
func (d *Discovery) Announce() error {
	addr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:"+fmt.Sprintf("%d", d.port))
	if err != nil {
		return fmt.Errorf("sync: resolve broadcast: %w", err)
	}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return fmt.Errorf("sync: dial broadcast: %w", err)
	}
	defer func() { _ = conn.Close() }()

	msg := announceMsg{
		Type:        "announce",
		DeviceID:    d.device.DeviceID,
		Name:        d.device.Name,
		PublicKey:   fmt.Sprintf("%x", d.device.PublicKey),
		Fingerprint: d.device.Fingerprint(),
		Port:        d.port,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("sync: marshal announce: %w", err)
	}
	_, err = conn.Write(data)
	return err
}

// Listen listens for UDP broadcasts from other devices.
func (d *Discovery) Listen() error {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", d.port))
	if err != nil {
		return fmt.Errorf("sync: resolve listen: %w", err)
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return fmt.Errorf("sync: listen: %w", err)
	}
	defer func() { _ = conn.Close() }()

	buf := make([]byte, 4096)
	for {
		n, raddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}
		var msg announceMsg
		if err := json.Unmarshal(buf[:n], &msg); err != nil {
			continue
		}
		if msg.DeviceID == d.device.DeviceID {
			continue
		}
		peer := &Peer{
			DeviceID:    msg.DeviceID,
			Name:        msg.Name,
			PublicKey:   msg.PublicKey,
			Address:     fmt.Sprintf("%s:%d", raddr.IP, msg.Port),
			LastSeen:    time.Now().UTC(),
			Fingerprint: msg.Fingerprint,
		}
		d.mu.Lock()
		d.peers[peer.DeviceID] = peer
		d.mu.Unlock()
		if d.onUpdate != nil {
			d.onUpdate(peer)
		}
	}
}

// Peers returns the currently known peers.
func (d *Discovery) Peers() []*Peer {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]*Peer, 0, len(d.peers))
	for _, p := range d.peers {
		out = append(out, p)
	}
	return out
}

type announceMsg struct {
	Type        string `json:"type"`
	DeviceID    string `json:"device_id"`
	Name        string `json:"name"`
	PublicKey   string `json:"public_key"`
	Fingerprint string `json:"fingerprint"`
	Port        int    `json:"port"`
}
