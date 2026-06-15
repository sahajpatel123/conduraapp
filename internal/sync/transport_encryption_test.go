package sync

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestSessionKey_Reproducible verifies that both sides of an
// X25519 ECDH derive the same session key.
func TestSessionKey_Reproducible(t *testing.T) {
	alicePriv, alicePub, err := generateX25519Ephemeral()
	if err != nil {
		t.Fatal(err)
	}
	bobPriv, bobPub, err := generateX25519Ephemeral()
	if err != nil {
		t.Fatal(err)
	}
	aID, _ := GenerateIdentity("alice")
	bID, _ := GenerateIdentity("bob")

	keyA, err := sessionKeyFrom(alicePriv, bobPub, aID.PublicKey, bID.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	keyB, err := sessionKeyFrom(bobPriv, alicePub, bID.PublicKey, aID.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(keyA, keyB) {
		t.Errorf("keys differ:\n  A=%x\n  B=%x", keyA, keyB)
	}
	if len(keyA) != sessionKeySize {
		t.Errorf("key length = %d, want %d", len(keyA), sessionKeySize)
	}
}

// TestSessionKey_DifferentRemotes verifies that different remotes
// produce different session keys (no key reuse).
func TestSessionKey_DifferentRemotes(t *testing.T) {
	alicePriv, _, err := generateX25519Ephemeral()
	if err != nil {
		t.Fatal(err)
	}
	_, bobPub, _ := generateX25519Ephemeral()
	_, evePub, _ := generateX25519Ephemeral()

	aID, _ := GenerateIdentity("alice")
	bID, _ := GenerateIdentity("bob")
	eID, _ := GenerateIdentity("eve")

	keyB, _ := sessionKeyFrom(alicePriv, bobPub, aID.PublicKey, bID.PublicKey)
	keyE, _ := sessionKeyFrom(alicePriv, evePub, aID.PublicKey, eID.PublicKey)

	if bytes.Equal(keyB, keyE) {
		t.Error("keys should differ for different remotes")
	}
}

// TestSessionEncryptDecrypt_RoundTrip verifies AES-256-GCM seal/open
// produces the original plaintext.
func TestSessionEncryptDecrypt_RoundTrip(t *testing.T) {
	key := make([]byte, sessionKeySize)
	for i := range key {
		key[i] = byte(i)
	}
	sess, err := newSession(key)
	if err != nil {
		t.Fatal(err)
	}
	plaintext := []byte("hello world \x00\x01\x02")
	ct, err := sess.encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}
	// Ciphertext must not equal plaintext.
	if bytes.Equal(ct, plaintext) {
		t.Error("ciphertext equals plaintext (encryption didn't happen)")
	}
	// Decrypt.
	pt, err := sess.decrypt(ct)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pt, plaintext) {
		t.Errorf("decrypted: %q, want %q", pt, plaintext)
	}
	// Tampered ciphertext must fail.
	ct[nonceSize+5] ^= 0x01
	if _, err := sess.decrypt(ct); err == nil {
		t.Error("decrypt should fail on tampered ciphertext")
	}
}

// TestExchangeEntries_WireIsEncrypted verifies that an attacker who
// taps the TCP connection cannot read the CRDT entries in clear.
// This is the regression test for the "plaintext sync" deal-breaker.
func TestExchangeEntries_WireIsEncrypted(t *testing.T) {
	aID, _ := GenerateIdentity("a")
	bID, _ := GenerateIdentity("b")
	storeA := NewStore()
	// Use distinctive, unlikely-to-collide plaintext values.
	storeA.Put(aID.DeviceID, "memory.color", []byte("PURPLE_TEST_VALUE_X9K2"))
	storeA.Put(aID.DeviceID, "skills.weather", []byte("WEATHER_LOOKUP_KEY_R7M3"))
	storeB := NewStore()

	// Set up a proxy listener that tees bytes through a recording
	// buffer. The real client dials this proxy, and the proxy
	// transparently relays bytes to the real server.
	proxyLn, _ := net.Listen("tcp", "127.0.0.1:0")
	defer proxyLn.Close()
	realLn, _ := net.Listen("tcp", "127.0.0.1:0")
	defer realLn.Close()
	proxyAddr := proxyLn.Addr().String()
	realAddr := realLn.Addr().String()

	var captured bytes.Buffer
	var captureMu sync.Mutex
	srvDone := make(chan error, 1)
	go func() {
		// Real server side
		realConn, err := realLn.Accept()
		if err != nil {
			srvDone <- err
			return
		}
		realConn.SetDeadline(time.Now().Add(5 * time.Second))
		_, err = ServeSync(realConn, bID, storeB)
		srvDone <- err
	}()

	go func() {
		// Proxy side: relay proxyAddr <-> realAddr, capturing all bytes.
		for {
			clientConn, err := proxyLn.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				serverConn, err := net.Dial("tcp", realAddr)
				if err != nil {
					return
				}
				defer serverConn.Close()
				// Two goroutines: client->server and server->client.
				// We capture bytes in BOTH directions.
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					buf := make([]byte, 4096)
					for {
						n, err := c.Read(buf)
						if n > 0 {
							captureMu.Lock()
							captured.Write(buf[:n])
							captureMu.Unlock()
							serverConn.Write(buf[:n])
						}
						if err != nil {
							return
						}
					}
				}()
				go func() {
					defer wg.Done()
					buf := make([]byte, 4096)
					for {
						n, err := serverConn.Read(buf)
						if n > 0 {
							captureMu.Lock()
							captured.Write(buf[:n])
							captureMu.Unlock()
							c.Write(buf[:n])
						}
						if err != nil {
							return
						}
					}
				}()
				wg.Wait()
			}(clientConn)
		}
	}()

	clientConn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer clientConn.Close()
	clientConn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = ExchangeEntries(clientConn, aID, storeA, true)
	if err != nil {
		t.Fatalf("ExchangeEntries: %v", err)
	}
	if e := <-srvDone; e != nil {
		t.Fatalf("server: %v", e)
	}

	captureMu.Lock()
	wireBytes := captured.Bytes()
	captureMu.Unlock()
	if len(wireBytes) == 0 {
		t.Fatal("no bytes captured (proxy not in path?)")
	}

	// The plaintext should NOT appear in the captured wire bytes.
	if bytes.Contains(wireBytes, []byte("PURPLE_TEST_VALUE_X9K2")) {
		t.Error("CRDT value 'PURPLE_TEST_VALUE_X9K2' appears in clear on the wire")
	}
	if bytes.Contains(wireBytes, []byte("WEATHER_LOOKUP_KEY_R7M3")) {
		t.Error("CRDT key 'WEATHER_LOOKUP_KEY_R7M3' appears in clear on the wire")
	}
	if bytes.Contains(wireBytes, []byte("memory.color")) {
		t.Error("CRDT key 'memory.color' appears in clear on the wire")
	}
}

// TestHelloE_RejectsWrongType ensures the handshake's first step is
// type-checked (defense against a misbehaving or malicious peer).
func TestHelloE_RejectsWrongType(t *testing.T) {
	aID, _ := GenerateIdentity("a")

	// Use a real TCP listener so the writer and reader don't deadlock
	// on a synchronous net.Pipe().
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	addr := ln.Addr().String()

	srvDone := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			srvDone <- err
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(2 * time.Second))
		// Pretend to be a peer that sends a non-hello_e message.
		bw := bufio.NewWriter(conn)
		_ = writeMsg(bw, syncMsg{Type: "garbage", DeviceID: "x"})
		_ = bw.Flush()
		// Hold the connection open until the client times out.
		time.Sleep(2 * time.Second)
		srvDone <- nil
	}()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(2 * time.Second))

	_, _, err = exchangeHelloE(conn, aID, true)
	if err == nil {
		t.Fatal("expected error on wrong message type")
	}
	if !strings.Contains(err.Error(), "expected hello_e") {
		t.Errorf("error %q should mention hello_e", err)
	}
}

// TestVerifyHello_BadSignatureFormat ensures signature validation is
// strict — wrong-length sigs and wrong-length pubkeys fail.
func TestVerifyHello_BadSignatureFormat(t *testing.T) {
	id, _ := GenerateIdentity("x")
	// Valid msg
	ok := syncMsg{
		Type:      "hello",
		DeviceID:  id.DeviceID,
		PublicKey: hex.EncodeToString(id.PublicKey),
		Signature: "00", // 1 byte - wrong length
	}
	if err := verifyHello(ok); err == nil {
		t.Error("expected verify failure for short sig")
	}
	// Empty fields
	if err := verifyHello(syncMsg{}); err == nil {
		t.Error("expected verify failure for empty msg")
	}
}
