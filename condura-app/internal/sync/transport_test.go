package sync

import (
	"encoding/hex"
	"net"
	"testing"
	"time"
)

func TestExchangeEntries_RoundTrip(t *testing.T) {
	aID, err := GenerateIdentity("device-a")
	if err != nil {
		t.Fatal(err)
	}
	bID, err := GenerateIdentity("device-b")
	if err != nil {
		t.Fatal(err)
	}
	storeA := NewStore()
	storeB := NewStore()
	storeA.Put(aID.DeviceID, "key1", []byte("from-a"))
	storeB.Put(bID.DeviceID, "key2", []byte("from-b"))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	done := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			done <- err
			return
		}
		_, err = ServeSync(conn, bID, storeB)
		done <- err
	}()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	n, err := ExchangeEntries(conn, aID, storeA, true)
	_ = conn.Close()
	_ = ln.Close()
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	if err := <-done; err != nil {
		t.Fatalf("server: %v", err)
	}
	if n == 0 {
		t.Fatal("expected merged entries on client")
	}
	if got := string(storeA.Get("key2").Value); got != "from-b" {
		t.Fatalf("storeA key2: %q", got)
	}
	if got := string(storeB.Get("key1").Value); got != "from-a" {
		t.Fatalf("storeB key1: %q", got)
	}
}

func TestVerifyHello_RejectsBadSig(t *testing.T) {
	id, _ := GenerateIdentity("x")
	msg := syncMsg{
		Type:      "hello",
		DeviceID:  id.DeviceID,
		PublicKey: hex.EncodeToString(id.PublicKey),
		Signature: "00",
	}
	if err := verifyHello(msg); err == nil {
		t.Fatal("expected verify failure")
	}
}

func TestPeerSyncAddress(t *testing.T) {
	addr, err := PeerSyncAddress("192.168.1.5:7667", 7667)
	if err != nil {
		t.Fatal(err)
	}
	if addr != "192.168.1.5:7668" {
		t.Fatalf("got %s", addr)
	}
}

func TestDialAndSync_Timeout(t *testing.T) {
	id, _ := GenerateIdentity("lonely")
	store := NewStore()
	_, err := DialAndSync("127.0.0.1:1", id, store)
	if err == nil {
		t.Fatal("expected dial error")
	}
	_ = time.Second // keep time import used on older toolchains
}
