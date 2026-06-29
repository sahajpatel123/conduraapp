// Phase 13 E2E — drives update.check / update.apply through IPC.
package daemon

import (
	"crypto/ed25519"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/updater"
)

func TestE2E_UpdateCheck_NoManifest(t *testing.T) {
	addr, subs, cleanup := startTrustDaemon(t)
	defer cleanup()

	// Avoid hitting the production manifest URL (slow / flaky in CI).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	subs.Updater.SetManifestURL(srv.URL)

	raw, err := trustCallRPC(t, addr, "update.check", map[string]any{})
	if err != nil {
		t.Fatalf("update.check: %v", err)
	}
	var result updater.Result
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, raw)
	}
	if result.CurrentVersion == "" {
		t.Error("expected current_version in result")
	}
	_ = subs
}

func TestE2E_UpdateCheck_SignedManifest(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	platform := updater.PlatformKey()
	binHash := "abc123"
	payload := updater.ManifestPayload{
		Version: "99.99.99",
		Channel: "stable",
		Platforms: map[string]updater.PlatformArtifact{
			platform: {
				DownloadURL: "http://example.invalid/binary",
				SHA256:      binHash,
			},
		},
	}
	sig, err := updater.SignPayload(payload, priv)
	if err != nil {
		t.Fatal(err)
	}
	sm := updater.SignedManifest{
		Version:    payload.Version,
		Channel:    payload.Channel,
		Platforms:  payload.Platforms,
		Ed25519Sig: sig,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(sm)
	}))
	defer srv.Close()

	addr, subs, cleanup := startTrustDaemon(t)
	defer cleanup()
	if subs.Updater == nil {
		t.Fatal("Updater is nil")
	}
	// Point updater at test manifest with test key.
	subs.Updater.SetManifestURL(srv.URL)
	subs.Updater.SetPublicKey(pub)

	raw, err := trustCallRPC(t, addr, "update.check", map[string]any{})
	if err != nil {
		t.Fatalf("update.check: %v", err)
	}
	var result updater.Result
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !result.UpdateAvailable {
		t.Fatalf("expected update available, got %+v", result)
	}
	if result.LatestVersion != "99.99.99" {
		t.Errorf("latest=%q", result.LatestVersion)
	}
	if result.SHA256 != binHash {
		t.Errorf("sha256=%q", result.SHA256)
	}
}

func TestE2E_UpdateCheck_BadSignatureRejected(t *testing.T) {
	pub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	_, wrongPriv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	payload := updater.ManifestPayload{
		Version: "99.99.99",
		Channel: "stable",
		Platforms: map[string]updater.PlatformArtifact{
			updater.PlatformKey(): {DownloadURL: "http://x", SHA256: "aa"},
		},
	}
	sig, _ := updater.SignPayload(payload, wrongPriv)
	sm := updater.SignedManifest{
		Version: payload.Version, Channel: payload.Channel,
		Platforms: payload.Platforms, Ed25519Sig: sig,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(sm)
	}))
	defer srv.Close()

	addr, subs, cleanup := startTrustDaemon(t)
	defer cleanup()
	subs.Updater.SetManifestURL(srv.URL)
	subs.Updater.SetPublicKey(pub)

	raw, err := trustCallRPC(t, addr, "update.check", map[string]any{})
	if err != nil {
		t.Fatalf("update.check: %v", err)
	}
	var result updater.Result
	_ = json.Unmarshal(raw, &result)
	if result.UpdateAvailable {
		t.Fatal("bad signature must not yield update_available")
	}
}
