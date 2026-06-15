package sync

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// SyncPortOffset is added to the UDP discovery port to reach the TCP sync port.
const SyncPortOffset = 1

// sessionProtocolID is mixed into every key derivation so a key derived
// here cannot be repurposed for another protocol.
const sessionProtocolID = "synaptic-sync-v1"

const (
	sessionKeySize = 32 // AES-256
	nonceSize      = 12 // AES-GCM nonce
)

// syncMsg is one JSON message on the wire.
//
// The protocol is:
//
//  1. <e1>   Initiator → Responder: hello_e (Ed25519 pub + X25519 ephemeral)
//  2. <e1>   Responder → Initiator: hello_e (Ed25519 pub + X25519 ephemeral)
//  3.        Both sides derive session_key = HKDF(X25519(localPriv, remoteEphemPub), edPubs)
//  4. <s>    Initiator → Responder: hello (Ed25519-signed identity, encrypted)
//  5. <s>    Responder → Initiator: hello (Ed25519-signed identity, encrypted)
//  6. <s>    Initiator → Responder: entries (CRDT, encrypted)
//  7. <s>    Responder → Initiator: entries (CRDT, encrypted)
//
// Steps 1-2 are plaintext but reveal only the long-term Ed25519 pub and
// an ephemeral X25519 pub. No CRDT data is sent in clear. The session
// key is unique to this connection (ephemeral X25519 provides forward
// secrecy within a single session; long-term Ed25519 keys are bound
// into the HKDF info so the channel is identity-locked).
type syncMsg struct {
	Type      string    `json:"type"`
	DeviceID  string    `json:"device_id,omitempty"`
	Name      string    `json:"name,omitempty"`
	PublicKey string    `json:"public_key,omitempty"`
	XPub      string    `json:"xpub,omitempty"`
	Signature string    `json:"signature,omitempty"`
	Entries   []*Entry  `json:"entries,omitempty"`
	// For pair_request / pair_response (out-of-band pairing flow)
	PairingToken    string `json:"pairing_token,omitempty"`
	PairingAccepted bool   `json:"pairing_accepted,omitempty"`
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

// signHello produces the Ed25519 signature over (DeviceID || PublicKey).
// Sent INSIDE the encrypted session — an eavesdropper cannot even see
// which device IDs are talking.
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

// ---------------------------------------------------------------------------
// Encryption (AES-256-GCM, X25519-derived key, identity-locked)
// ---------------------------------------------------------------------------

// session is an encrypted channel between two devices.
type session struct {
	gcm cipher.AEAD
}

func newSession(key []byte) (*session, error) {
	if len(key) != sessionKeySize {
		return nil, fmt.Errorf("sync: session key must be %d bytes, got %d", sessionKeySize, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("sync: aes: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("sync: gcm: %w", err)
	}
	return &session{gcm: gcm}, nil
}

func (s *session) encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("sync: nonce: %w", err)
	}
	sealed := s.gcm.Seal(nil, nonce, plaintext, nil)
	out := make([]byte, 0, nonceSize+len(sealed))
	out = append(out, nonce...)
	out = append(out, sealed...)
	return out, nil
}

func (s *session) decrypt(blob []byte) ([]byte, error) {
	if len(blob) < nonceSize {
		return nil, errors.New("sync: ciphertext too short")
	}
	nonce := blob[:nonceSize]
	sealed := blob[nonceSize:]
	pt, err := s.gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("sync: gcm open: %w", err)
	}
	return pt, nil
}

// generateX25519Ephemeral returns a fresh X25519 keypair for one session.
func generateX25519Ephemeral() (priv [32]byte, pub [32]byte, err error) {
	curve := ecdh.X25519()
	privKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return priv, pub, fmt.Errorf("sync: ephemeral x25519: %w", err)
	}
	copy(priv[:], privKey.Bytes())
	copy(pub[:], privKey.PublicKey().Bytes())
	return priv, pub, nil
}

// deriveX25519Shared runs the X25519 key agreement.
func deriveX25519Shared(localPriv [32]byte, remotePub [32]byte) ([]byte, error) {
	curve := ecdh.X25519()
	priv, err := curve.NewPrivateKey(localPriv[:])
	if err != nil {
		return nil, fmt.Errorf("sync: x25519 priv parse: %w", err)
	}
	pub, err := curve.NewPublicKey(remotePub[:])
	if err != nil {
		return nil, fmt.Errorf("sync: x25519 pub parse: %w", err)
	}
	shared, err := priv.ECDH(pub)
	if err != nil {
		return nil, fmt.Errorf("sync: x25519 ecdh: %w", err)
	}
	out := make([]byte, len(shared))
	copy(out, shared)
	return out, nil
}

// sessionKeyFrom derives the AES-256 key from a pair of X25519 keys
// and the two long-term Ed25519 identities. Bound to identities so
// the channel cannot be repurposed.
func sessionKeyFrom(
	localXPriv [32]byte, remoteXPub [32]byte,
	localEdPub ed25519.PublicKey, remoteEdPub ed25519.PublicKey,
) ([]byte, error) {
	shared, err := deriveX25519Shared(localXPriv, remoteXPub)
	if err != nil {
		return nil, err
	}
	// Sort the Ed pubkeys so the info is the same on both sides.
	var info []byte
	if string(localEdPub) < string(remoteEdPub) {
		info = append(info, localEdPub...)
		info = append(info, remoteEdPub...)
	} else {
		info = append(info, remoteEdPub...)
		info = append(info, localEdPub...)
	}
	reader := hkdfNew(shared, info)
	key := make([]byte, sessionKeySize)
	if _, err := io.ReadFull(reader, key); err != nil {
		return nil, fmt.Errorf("sync: hkdf: %w", err)
	}
	return key, nil
}

// prelimSessionKey derives a session key from the X25519 shared secret
// only, without identity binding. Used to encrypt the Ed25519 hello
// during the handshake, before we know the remote Ed25519 pub. After
// the hello exchange reveals it, we re-derive the identity-bound final
// key with `sessionKeyFrom`. Since both sides compute the same X25519
// shared secret, and the info string is a constant, both sides get the
// same prelim key. The prelim key is replaced with the identity-bound
// final key after step 5 of the protocol.
func prelimSessionKey(localXPriv [32]byte, remoteXPub [32]byte) ([]byte, error) {
	shared, err := deriveX25519Shared(localXPriv, remoteXPub)
	if err != nil {
		return nil, err
	}
	info := []byte("prelim-v1")
	reader := hkdfNew(shared, info)
	key := make([]byte, sessionKeySize)
	if _, err := io.ReadFull(reader, key); err != nil {
		return nil, fmt.Errorf("sync: prelim hkdf: %w", err)
	}
	return key, nil
}

// hkdfNew is an inline HKDF-SHA256 implementation (RFC 5869) so we
// don't need to pull in golang.org/x/crypto. We only need
// extract-and-expand with fixed-length outputs.
func hkdfNew(secret []byte, info []byte) io.Reader {
	return &hkdfReader{secret: secret, info: info, salt: []byte(sessionProtocolID)}
}

type hkdfReader struct {
	secret []byte
	info   []byte
	salt   []byte
	prk    []byte
	prev   []byte
	counter byte
	pos    int
	buf    []byte
	done   bool
}

func (h *hkdfReader) Read(out []byte) (int, error) {
	if h.done {
		return 0, io.EOF
	}
	if h.prk == nil {
		mac := hmac.New(sha256.New, h.salt)
		mac.Write(h.secret)
		h.prk = mac.Sum(nil)
	}
	written := 0
	for written < len(out) {
		if h.pos >= len(h.buf) {
			if h.counter >= 255 {
				h.done = true
				return written, nil
			}
			h.counter++
			mac := hmac.New(sha256.New, h.prk)
			mac.Write(h.prev)
			mac.Write(h.info)
			mac.Write([]byte{h.counter})
			h.buf = mac.Sum(nil)
			h.prev = h.buf
			h.pos = 0
		}
		n := copy(out[written:], h.buf[h.pos:])
		written += n
		h.pos += n
	}
	return written, nil
}

// writeEncryptedFrame: 4-byte BE length + (nonce || ciphertext)
func writeEncryptedFrame(w io.Writer, sess *session, plaintext []byte) error {
	var blob []byte
	if sess != nil {
		var err error
		blob, err = sess.encrypt(plaintext)
		if err != nil {
			return err
		}
	} else {
		blob = plaintext
	}
	if len(blob) > 0xFFFFFFFF {
		return fmt.Errorf("sync: frame too large: %d", len(blob))
	}
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(blob)))
	if _, err := w.Write(lenBuf[:]); err != nil {
		return err
	}
	_, err := w.Write(blob)
	return err
}

func readEncryptedFrame(r io.Reader, sess *session) ([]byte, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf[:])
	if length == 0 || length > 64*1024*1024 {
		return nil, fmt.Errorf("sync: frame length invalid: %d", length)
	}
	blob := make([]byte, length)
	if _, err := io.ReadFull(r, blob); err != nil {
		return nil, err
	}
	if sess == nil {
		return blob, nil
	}
	return sess.decrypt(blob)
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

// exchangeHelloE does step 1-2 of the protocol. Returns the local
// ephemeral priv and the remote ephemeral pub.
func exchangeHelloE(conn net.Conn, local *DeviceIdentity, initiator bool) (localXPriv [32]byte, remoteXPub [32]byte, err error) {
	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))
	br := bufio.NewReader(conn)
	bw := bufio.NewWriter(conn)

	ephPriv, ephPub, err := generateX25519Ephemeral()
	if err != nil {
		return localXPriv, remoteXPub, err
	}
	hello := syncMsg{
		Type:      "hello_e",
		DeviceID:  local.DeviceID,
		Name:      local.Name,
		PublicKey: hex.EncodeToString(local.PublicKey),
		XPub:      hex.EncodeToString(ephPub[:]),
	}
	if initiator {
		if err := writeMsg(bw, hello); err != nil {
			return localXPriv, remoteXPub, err
		}
		if err := bw.Flush(); err != nil {
			return localXPriv, remoteXPub, err
		}
		remote, err := readMsg(br)
		if err != nil {
			return localXPriv, remoteXPub, err
		}
		if remote.Type != "hello_e" {
			return localXPriv, remoteXPub, fmt.Errorf("sync: expected hello_e, got %s", remote.Type)
		}
		remoteBytes, err := hex.DecodeString(remote.XPub)
		if err != nil || len(remoteBytes) != 32 {
			return localXPriv, remoteXPub, fmt.Errorf("sync: invalid ephemeral pub")
		}
		copy(remoteXPub[:], remoteBytes)
		return ephPriv, remoteXPub, nil
	}
	// Responder
	remote, err := readMsg(br)
	if err != nil {
		return localXPriv, remoteXPub, err
	}
	if remote.Type != "hello_e" {
		return localXPriv, remoteXPub, fmt.Errorf("sync: expected hello_e, got %s", remote.Type)
	}
	if err := writeMsg(bw, hello); err != nil {
		return localXPriv, remoteXPub, err
	}
	if err := bw.Flush(); err != nil {
		return localXPriv, remoteXPub, err
	}
	remoteBytes, err := hex.DecodeString(remote.XPub)
	if err != nil || len(remoteBytes) != 32 {
		return localXPriv, remoteXPub, fmt.Errorf("sync: invalid ephemeral pub")
	}
	copy(remoteXPub[:], remoteBytes)
	return ephPriv, remoteXPub, nil
}

// ExchangeEntries runs the full sync protocol on an established
// connection: hello_e handshake → encrypted hello exchange →
// encrypted CRDT exchange. All CRDT data is encrypted under
// AES-256-GCM with a session key derived from X25519 ECDH bound
// to both long-term Ed25519 identities.
func ExchangeEntries(conn net.Conn, local *DeviceIdentity, store *Store, initiator bool) (int, error) {
	if local == nil || store == nil {
		return 0, fmt.Errorf("sync: identity and store required")
	}

	localXPriv, remoteXPub, err := exchangeHelloE(conn, local, initiator)
	if err != nil {
		return 0, err
	}

	// Reset deadline for the encrypted phase.
	_ = conn.SetDeadline(time.Now().Add(60 * time.Second))

	br := bufio.NewReader(conn)
	bw := bufio.NewWriter(conn)

	// Step 4-5: encrypted hello (Ed25519-signed identity) under prelim session.
	// After the hello exchange reveals the remote Ed25519 pub, we re-derive
	// the identity-bound session key.
	prelimKey, err := prelimSessionKey(localXPriv, remoteXPub)
	if err != nil {
		return 0, err
	}
	prelimSess, err := newSession(prelimKey)
	if err != nil {
		return 0, err
	}

	sig, err := signHello(local)
	if err != nil {
		return 0, err
	}
	localHello := syncMsg{
		Type:      "hello",
		DeviceID:  local.DeviceID,
		Name:      local.Name,
		PublicKey: hex.EncodeToString(local.PublicKey),
		Signature: sig,
	}
	helloBytes, err := json.Marshal(localHello)
	if err != nil {
		return 0, err
	}

	if initiator {
		if err := writeEncryptedFrame(bw, prelimSess, helloBytes); err != nil {
			return 0, err
		}
		if err := bw.Flush(); err != nil {
			return 0, err
		}
		remoteBytes, err := readEncryptedFrame(br, prelimSess)
		if err != nil {
			return 0, err
		}
		var remote syncMsg
		if err := json.Unmarshal(remoteBytes, &remote); err != nil {
			return 0, err
		}
		if err := verifyHello(remote); err != nil {
			return 0, err
		}
		remoteEdPub, err := hex.DecodeString(remote.PublicKey)
		if err != nil || len(remoteEdPub) != ed25519.PublicKeySize {
			return 0, fmt.Errorf("sync: remote ed pub invalid")
		}
		finalKey, err := sessionKeyFrom(localXPriv, remoteXPub, local.PublicKey, remoteEdPub)
		if err != nil {
			return 0, err
		}
		finalSess, err := newSession(finalKey)
		if err != nil {
			return 0, err
		}
		return exchangeCRDT(br, bw, store, finalSess, initiator)
	}
	// Responder
	remoteBytes, err := readEncryptedFrame(br, prelimSess)
	if err != nil {
		return 0, err
	}
	var remote syncMsg
	if err := json.Unmarshal(remoteBytes, &remote); err != nil {
		return 0, err
	}
	if err := verifyHello(remote); err != nil {
		return 0, err
	}
	remoteEdPub, err := hex.DecodeString(remote.PublicKey)
	if err != nil || len(remoteEdPub) != ed25519.PublicKeySize {
		return 0, fmt.Errorf("sync: remote ed pub invalid")
	}
	if err := writeEncryptedFrame(bw, prelimSess, helloBytes); err != nil {
		return 0, err
	}
	if err := bw.Flush(); err != nil {
		return 0, err
	}
	finalKey, err := sessionKeyFrom(localXPriv, remoteXPub, local.PublicKey, remoteEdPub)
	if err != nil {
		return 0, err
	}
	finalSess, err := newSession(finalKey)
	if err != nil {
		return 0, err
	}
	return exchangeCRDT(br, bw, store, finalSess, initiator)
}

// exchangeCRDT sends and receives CRDT entries over the fully
// identity-bound encrypted channel.
func exchangeCRDT(br *bufio.Reader, bw *bufio.Writer, store *Store, sess *session, initiator bool) (int, error) {
	entries := store.Entries()
	entriesMsg := syncMsg{Type: "entries", Entries: entries}
	entriesBytes, err := json.Marshal(entriesMsg)
	if err != nil {
		return 0, err
	}
	if initiator {
		if err := writeEncryptedFrame(bw, sess, entriesBytes); err != nil {
			return 0, err
		}
		if err := bw.Flush(); err != nil {
			return 0, err
		}
		remoteBytes, err := readEncryptedFrame(br, sess)
		if err != nil {
			return 0, err
		}
		var remoteEntriesMsg syncMsg
		if err := json.Unmarshal(remoteBytes, &remoteEntriesMsg); err != nil {
			return 0, err
		}
		return mergeEntries(store, remoteEntriesMsg.Entries), nil
	}
	// Responder
	remoteBytes, err := readEncryptedFrame(br, sess)
	if err != nil {
		return 0, err
	}
	var remoteEntriesMsg syncMsg
	if err := json.Unmarshal(remoteBytes, &remoteEntriesMsg); err != nil {
		return 0, err
	}
	merged := mergeEntries(store, remoteEntriesMsg.Entries)
	if err := writeEncryptedFrame(bw, sess, entriesBytes); err != nil {
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

// DialAndSync connects to peerTCPAddr and exchanges CRDT entries over
// an encrypted, identity-verified channel.
func DialAndSync(peerTCPAddr string, local *DeviceIdentity, store *Store) (int, error) {
	conn, err := net.DialTimeout("tcp", peerTCPAddr, 10*time.Second)
	if err != nil {
		return 0, fmt.Errorf("sync: dial %s: %w", peerTCPAddr, err)
	}
	defer func() { _ = conn.Close() }()
	return ExchangeEntries(conn, local, store, true)
}

// ServeSync accepts one inbound sync connection, runs the encrypted
// handshake, verifies the peer against the paired set, and merges
// CRDT entries.
func ServeSync(conn net.Conn, local *DeviceIdentity, store *Store) (int, error) {
	defer func() { _ = conn.Close() }()
	return ExchangeEntries(conn, local, store, false)
}

// ExchangeEntriesGated is the paired-set-aware variant of
// ExchangeEntries used for inbound connections. It runs the same
// encrypted handshake, but after the remote device ID is known
// (from the encrypted hello), it consults the gate (which knows
// the paired set) before applying CRDT entries.
func ExchangeEntriesGated(conn net.Conn, local *DeviceIdentity, gate *PairedGate) (int, error) {
	if local == nil || gate == nil {
		return 0, fmt.Errorf("sync: identity and gate required")
	}

	localXPriv, remoteXPub, err := exchangeHelloE(conn, local, false)
	if err != nil {
		return 0, err
	}
	_ = conn.SetDeadline(time.Now().Add(60 * time.Second))

	br := bufio.NewReader(conn)
	bw := bufio.NewWriter(conn)

	prelimKey, err := prelimSessionKey(localXPriv, remoteXPub)
	if err != nil {
		return 0, err
	}
	prelimSess, err := newSession(prelimKey)
	if err != nil {
		return 0, err
	}

	sig, err := signHello(local)
	if err != nil {
		return 0, err
	}
	localHello := syncMsg{
		Type:      "hello",
		DeviceID:  local.DeviceID,
		Name:      local.Name,
		PublicKey: hex.EncodeToString(local.PublicKey),
		Signature: sig,
	}
	helloBytes, err := json.Marshal(localHello)
	if err != nil {
		return 0, err
	}

	// Read the remote's encrypted hello.
	remoteBytes, err := readEncryptedFrame(br, prelimSess)
	if err != nil {
		return 0, err
	}
	var remote syncMsg
	if err := json.Unmarshal(remoteBytes, &remote); err != nil {
		return 0, err
	}
	if err := verifyHello(remote); err != nil {
		return 0, err
	}
	remoteEdPub, err := hex.DecodeString(remote.PublicKey)
	if err != nil || len(remoteEdPub) != ed25519.PublicKeySize {
		return 0, fmt.Errorf("sync: remote ed pub invalid")
	}
	// Send our encrypted hello.
	if err := writeEncryptedFrame(bw, prelimSess, helloBytes); err != nil {
		return 0, err
	}
	if err := bw.Flush(); err != nil {
		return 0, err
	}
	finalKey, err := sessionKeyFrom(localXPriv, remoteXPub, local.PublicKey, remoteEdPub)
	if err != nil {
		return 0, err
	}
	finalSess, err := newSession(finalKey)
	if err != nil {
		return 0, err
	}
	return exchangeCRDTGated(br, bw, gate, remote.DeviceID, finalSess, false)
}

// exchangeCRDTGated is the gated variant of exchangeCRDT. It reads
// the remote's entries and applies them through the gate, which
// enforces the paired-set policy.
func exchangeCRDTGated(
	br *bufio.Reader, bw *bufio.Writer,
	gate *PairedGate, remoteDeviceID string, sess *session, _ bool,
) (int, error) {
	// Responder reads first.
	remoteBytes, err := readEncryptedFrame(br, sess)
	if err != nil {
		return 0, err
	}
	var remoteEntriesMsg syncMsg
	if err := json.Unmarshal(remoteBytes, &remoteEntriesMsg); err != nil {
		return 0, err
	}
	// Apply through gate. The gate's Merge checks the paired set.
	merged := 0
	for _, e := range remoteEntriesMsg.Entries {
		if e == nil {
			continue
		}
		if gate.Merge(remoteDeviceID, e) {
			merged++
		}
	}
	// Send our own entries back.
	entries := gate.Entries()
	entriesMsg := syncMsg{Type: "entries", Entries: entries}
	entriesBytes, err := json.Marshal(entriesMsg)
	if err != nil {
		return merged, err
	}
	if err := writeEncryptedFrame(bw, sess, entriesBytes); err != nil {
		return merged, err
	}
	if err := bw.Flush(); err != nil {
		return merged, err
	}
	return merged, nil
}
