package sync

import (
	"io"
	"log/slog"
	"testing"
)

func TestEngine_PutGetStatus(t *testing.T) {
	id, err := GenerateIdentity("test-device")
	if err != nil {
		t.Fatalf("identity: %v", err)
	}
	store := NewStore()
	disc := NewDiscovery(id, 0)
	eng := NewEngine(id, store, disc, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))

	eng.Put("hello", []byte("world"))
	if got := string(eng.Get("hello")); got != "world" {
		t.Fatalf("Get: got %q", got)
	}
	st := eng.Status()
	if st.DeviceID != id.DeviceID {
		t.Fatalf("status device_id: %s", st.DeviceID)
	}
	if st.Entries != 1 {
		t.Fatalf("entries: %d", st.Entries)
	}
}

func TestEngine_DiscoveredPeersEmpty(t *testing.T) {
	id, _ := GenerateIdentity("peer-test")
	store := NewStore()
	disc := NewDiscovery(id, 0)
	eng := NewEngine(id, store, disc, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if peers := eng.DiscoveredPeers(); len(peers) != 0 {
		t.Fatalf("expected 0 peers, got %d", len(peers))
	}
}

func TestEngine_StartStop(t *testing.T) {
	id, _ := GenerateIdentity("lifecycle")
	store := NewStore()
	disc := NewDiscovery(id, 47667)
	eng := NewEngine(id, store, disc, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	eng.Start()
	st := eng.Status()
	if !st.Running {
		t.Fatal("expected running after Start")
	}
	eng.Stop()
	st = eng.Status()
	if st.Running {
		t.Fatal("expected stopped after Stop")
	}
}
