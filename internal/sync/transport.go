package sync

import (
	"bufio"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// SyncPortOffset is added to the UDP discovery port to reach the TCP sync port.
const SyncPortOffset = 1

// syncMsg is one newline-delimited JSON message on the sync wire.
type syncMsg struct {
	Type    string   `json:"type"`
	DeviceID string  `json:"device_id,omitempty"`
	Name    string   `json:"name,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
	Signature string `json:"signature,omitempty"`
	Entries []*Entry `json:"entries,omitempty"`
}

// SyncTCPPort returns the TCP port used for CRDT sync given a discovery UDP port.
func SyncTCPPort(discoveryPort int) int {
	if discoveryPort <= 0 {
		discoveryPort = 7667
	}
	return discoveryPort + SyncPortOffset
}

// PeerSyncAddress maps a discovery peer address (host:udpPort) to host:tcpSyncPort.
func PeerSyncAddress(peerAddr string, discoveryPort int) (string, error) {
	host, _, err := net.SplitHostPort(peerAddr)
	if err != nil {
		return "", fmt.Errorf("sync: peer address: %w", err)
	}
	return net.JoinHostPort(host, fmt.Sprintf("%d", SyncTCPPort(discoveryPort))), nil
}

func signHello(id *DeviceIdentity) (string, error) {
	payload := id.DeviceID + ":" + hex.EncodeToString(id.PublicKey)
	sig := id.Sign([]byte(payload))
	return hex.EncodeToString(sig), nil
}

func verifyHello(msg syncMsg) error {
	if msg.DeviceID == "" || msg.PublicKey == "" || msg.Signature == "" {
		return fmt.Errorf("sync: incomplete hello")
	}
	pubBytes, err := hex.DecodeString(msg.PublicKey)
	if err != nil || len(pubBytes) != ed25519.PublicKeySize {
		return fmt.Errorf("sync: public key invalid")
	}
	sigBytes, err := hex.DecodeString(msg.Signature)
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		return fmt.Errorf("sync: signature invalid")
	}
	payload := msg.DeviceID + ":" + msg.PublicKey
	if !ed25519.Verify(ed25519.PublicKey(pubBytes), []byte(payload), sigBytes) {
		return fmt.Errorf("sync: hello signature invalid")
	}
	return nil
}

func writeMsg(w io.Writer, msg syncMsg) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = w.Write(append(b, '\n'))
	return err
}

func readMsg(r *bufio.Reader) (syncMsg, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return syncMsg{}, err
	}
	line = strings.TrimSpace(line)
	var msg syncMsg
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return syncMsg{}, err
	}
	return msg, nil
}

// ExchangeEntries runs the sync protocol on an established connection.
// The initiator sends hello+entries first; the peer responds with hello+entries.
func ExchangeEntries(conn net.Conn, local *DeviceIdentity, store *Store, initiator bool) (int, error) {
	if local == nil || store == nil {
		return 0, fmt.Errorf("sync: identity and store required")
	}
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))
	br := bufio.NewReader(conn)
	bw := bufio.NewWriter(conn)

	sig, err := signHello(local)
	if err != nil {
		return 0, err
	}
	hello := syncMsg{
		Type:      "hello",
		DeviceID:  local.DeviceID,
		Name:      local.Name,
		PublicKey: hex.EncodeToString(local.PublicKey),
		Signature: sig,
		Entries:   store.Entries(),
	}

	if initiator {
		if err := writeMsg(bw, hello); err != nil {
			return 0, err
		}
		if err := bw.Flush(); err != nil {
			return 0, err
		}
		remote, err := readMsg(br)
		if err != nil {
			return 0, err
		}
		if remote.Type != "hello" {
			return 0, fmt.Errorf("sync: expected hello, got %s", remote.Type)
		}
		if err := verifyHello(remote); err != nil {
			return 0, err
		}
		return mergeEntries(store, remote.Entries), nil
	}

	remote, err := readMsg(br)
	if err != nil {
		return 0, err
	}
	if remote.Type != "hello" {
		return 0, fmt.Errorf("sync: expected hello, got %s", remote.Type)
	}
	if err := verifyHello(remote); err != nil {
		return 0, err
	}
	merged := mergeEntries(store, remote.Entries)
	if err := writeMsg(bw, hello); err != nil {
		return merged, err
	}
	if err := bw.Flush(); err != nil {
		return merged, err
	}
	return merged, nil
}

func mergeEntries(store *Store, remote []*Entry) int {
	n := 0
	for _, e := range remote {
		if e != nil && store.Merge(e) {
			n++
		}
	}
	return n
}

// DialAndSync connects to peerTCPAddr and exchanges CRDT entries.
func DialAndSync(peerTCPAddr string, local *DeviceIdentity, store *Store) (int, error) {
	conn, err := net.DialTimeout("tcp", peerTCPAddr, 10*time.Second)
	if err != nil {
		return 0, fmt.Errorf("sync: dial %s: %w", peerTCPAddr, err)
	}
	defer func() { _ = conn.Close() }()
	return ExchangeEntries(conn, local, store, true)
}

// ServeSync accepts one inbound sync connection and merges entries.
func ServeSync(conn net.Conn, local *DeviceIdentity, store *Store) (int, error) {
	defer func() { _ = conn.Close() }()
	return ExchangeEntries(conn, local, store, false)
}
