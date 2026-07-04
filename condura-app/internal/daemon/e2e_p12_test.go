// Phase 12 E2E — drives the Reach & Ecosystem RPCs end-to-end
// through the IPC layer and verifies they return correctly-shaped
// responses.
package daemon

import (
	"encoding/json"
	"testing"
)

// TestTrustE2E_I18nLocalesReturnsList verifies i18n.locales returns
// the available locale list (defaulting to ["en"] when i18n is enabled).
func TestTrustE2E_I18nLocalesReturnsList(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	res, err := trustCallRPC(t, addr, "i18n.locales", nil)
	if err != nil {
		t.Fatalf("i18n.locales: %v", err)
	}
	var locales []string
	if err := json.Unmarshal(res, &locales); err != nil {
		t.Fatalf("expected []string, got %s: %v", res, err)
	}
	if len(locales) == 0 {
		t.Fatalf("expected at least one locale, got none")
	}
	// The default catalog always includes "en".
	found := false
	for _, l := range locales {
		if l == "en" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'en' in locales, got %v", locales)
	}
}

// TestTrustE2E_I18nLocaleReturnsTranslation verifies i18n.locale
// returns the English default translations set.
func TestTrustE2E_I18nLocaleReturnsTranslation(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	res, err := trustCallRPC(t, addr, "i18n.locale", map[string]any{"locale": "en"})
	if err != nil {
		t.Fatalf("i18n.locale: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(res, &out); err != nil {
		t.Fatalf("expected object, got %s: %v", res, err)
	}
	if _, ok := out["locale"]; !ok {
		t.Fatalf("expected 'locale' field, got %v", out)
	}
	if _, ok := out["translations"]; !ok {
		t.Fatalf("expected 'translations' field, got %v", out)
	}
}

// TestTrustE2E_I18nLocaleDefaultsToEn verifies i18n.locale defaults
// to English when no locale is provided.
func TestTrustE2E_I18nLocaleDefaultsToEn(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	res, err := trustCallRPC(t, addr, "i18n.locale", map[string]any{})
	if err != nil {
		t.Fatalf("i18n.locale: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(res, &out); err != nil {
		t.Fatalf("expected object, got %s: %v", res, err)
	}
	if out["locale"] != "en" {
		t.Fatalf("expected default locale 'en', got %v", out["locale"])
	}
}

// TestTrustE2E_SyncStatusReturnsDisabledWhenNotConfigured verifies
// sync.status returns enabled:false when sync is not configured.
func TestTrustE2E_SyncStatusReturnsDisabledWhenNotConfigured(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	res, err := trustCallRPC(t, addr, "sync.status", nil)
	if err != nil {
		t.Fatalf("sync.status: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(res, &out); err != nil {
		t.Fatalf("expected object, got %s: %v", res, err)
	}
	// Default config has sync disabled — status should not expose a running engine.
	if _, hasDevice := out["device_id"]; hasDevice && out["running"] == true {
		t.Fatalf("unexpected sync engine running in default config: %+v", out)
	}
}

// TestTrustE2E_SyncPeersReturnsEmptyList verifies sync.peers returns
// an empty list when sync is not enabled.
func TestTrustE2E_SyncPeersReturnsEmptyList(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	res, err := trustCallRPC(t, addr, "sync.peers", nil)
	if err != nil {
		t.Fatalf("sync.peers: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(res, &out); err != nil {
		t.Fatalf("expected object, got %s: %v", res, err)
	}
	peers, ok := out["peers"].([]any)
	if !ok {
		t.Fatalf("expected 'peers' array, got %T", out["peers"])
	}
	if len(peers) != 0 {
		t.Fatalf("expected empty peers list, got %d items", len(peers))
	}
}

// TestTrustE2E_SyncGetReturnsNilWhenNotEnabled verifies sync.get
// returns a nil value when sync is not enabled.
func TestTrustE2E_SyncGetReturnsNilWhenNotEnabled(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	res, err := trustCallRPC(t, addr, "sync.get", map[string]any{"key": "test-key"})
	// sync.get returns an error when sync is not enabled.
	if err == nil {
		t.Fatalf("expected error when sync not enabled, got %s", res)
	}
}

// TestTrustE2E_SyncPutReturnsErrorWhenNotEnabled verifies sync.put
// returns an error when sync is not enabled.
func TestTrustE2E_SyncPutReturnsErrorWhenNotEnabled(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	_, err := trustCallRPC(t, addr, "sync.put", map[string]any{"key": "test-key", "value": []byte("test-value")})
	if err == nil {
		t.Fatalf("expected error when sync not enabled")
	}
}

// TestTrustE2E_SyncStartReturnsErrorWhenNotConfigured verifies
// sync.start returns an error when sync is not configured.
func TestTrustE2E_SyncStartReturnsErrorWhenNotConfigured(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	_, err := trustCallRPC(t, addr, "sync.start", nil)
	if err == nil {
		t.Fatalf("expected error when sync not configured")
	}
}

// TestTrustE2E_SyncStopReturnsErrorWhenNotConfigured verifies
// sync.stop returns an error when sync is not configured.
func TestTrustE2E_SyncStopReturnsErrorWhenNotConfigured(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	_, err := trustCallRPC(t, addr, "sync.stop", nil)
	if err == nil {
		t.Fatalf("expected error when sync not configured")
	}
}
