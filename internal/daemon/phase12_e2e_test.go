package daemon

import (
	"encoding/json"
	"testing"
)

// TestPhase12E2E_I18nAndSyncRPCs verifies Phase 12 RPC wiring
// through the live daemon IPC layer.
func TestPhase12E2E_I18nAndSyncRPCs(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()

	res, err := trustCallRPC(t, addr, "i18n.locales", nil)
	if err != nil {
		t.Fatalf("i18n.locales: %v", err)
	}
	var locales []string
	if err := json.Unmarshal([]byte(res), &locales); err != nil {
		t.Fatalf("decode locales: %v: %s", err, res)
	}
	if len(locales) == 0 {
		t.Fatal("expected at least one locale")
	}

	res, err = trustCallRPC(t, addr, "i18n.locale", map[string]any{"locale": "en"})
	if err != nil {
		t.Fatalf("i18n.locale: %v", err)
	}
	var locResp struct {
		Locale       string            `json:"locale"`
		Translations map[string]string `json:"translations"`
	}
	if err := json.Unmarshal([]byte(res), &locResp); err != nil {
		t.Fatalf("decode locale: %v: %s", err, res)
	}
	if locResp.Locale != "en" {
		t.Fatalf("locale: %s", locResp.Locale)
	}

	res, err = trustCallRPC(t, addr, "sync.status", nil)
	if err != nil {
		t.Fatalf("sync.status: %v", err)
	}
	var syncSt map[string]any
	if err := json.Unmarshal([]byte(res), &syncSt); err != nil {
		t.Fatalf("decode sync.status: %v", err)
	}

	res, err = trustCallRPC(t, addr, "sync.peers", nil)
	if err != nil {
		t.Fatalf("sync.peers: %v", err)
	}
	var peersResp struct {
		Peers []any `json:"peers"`
	}
	if err := json.Unmarshal([]byte(res), &peersResp); err != nil {
		t.Fatalf("decode sync.peers: %v: %s", err, res)
	}
	if peersResp.Peers == nil {
		t.Fatal("peers field missing")
	}
}
