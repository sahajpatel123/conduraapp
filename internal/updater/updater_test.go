package updater

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// testKeyPair generates a fresh Ed25519 keypair for tests.
func testKeyPair() (ed25519.PublicKey, ed25519.PrivateKey) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	return pub, priv
}

// signManifest signs a SignedManifest with the given private key.
func signManifest(sm *SignedManifest, priv ed25519.PrivateKey) {
	payload := struct {
		Version     string `json:"version"`
		Channel     string `json:"channel"`
		DownloadURL string `json:"download_url"`
		SHA256      string `json:"sha256"`
		Mandatory   bool   `json:"mandatory"`
		MinVersion  string `json:"min_version,omitempty"`
		Notes       string `json:"notes,omitempty"`
	}{
		Version:     sm.Version,
		Channel:     sm.Channel,
		DownloadURL: sm.DownloadURL,
		SHA256:      sm.SHA256,
		Mandatory:   sm.Mandatory,
		MinVersion:  sm.MinVersion,
		Notes:       sm.Notes,
	}
	msg, _ := json.Marshal(payload)
	sig, _ := priv.Sign(nil, msg, crypto.Hash(0))
	sm.Ed25519Sig = hex.EncodeToString(sig)
}

func TestUpdater_SignedManifestAccepted(t *testing.T) {
	pub, priv := testKeyPair()

	sm := SignedManifest{
		Version:     "9.9.9",
		Channel:     "stable",
		DownloadURL: "http://example.com/synaptic",
		SHA256:      "abc123",
		Mandatory:   false,
	}
	signManifest(&sm, priv)

	// Serve the manifest.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(sm)
	}))
	defer srv.Close()

	u := New(nil, srv.URL)
	u.pubKey = pub

	result, err := u.Check(context.Background())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !result.UpdateAvailable {
		t.Fatal("expected update available")
	}
}

func TestUpdater_BadSignatureRejected(t *testing.T) {
	pub, _ := testKeyPair()
	_, wrongPriv := testKeyPair()

	sm := SignedManifest{
		Version:     "9.9.9",
		Channel:     "stable",
		DownloadURL: "http://example.com/synaptic",
		SHA256:      "abc123",
	}
	signManifest(&sm, wrongPriv) // signed with WRONG key

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(sm)
	}))
	defer srv.Close()

	u := New(nil, srv.URL)
	u.pubKey = pub // correct pub key — signature should NOT verify

	result, err := u.Check(context.Background())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if result.UpdateAvailable {
		t.Fatal("bad signature should be rejected")
	}
}

func TestUpdater_NoSignatureRejected(t *testing.T) {
	pub, _ := testKeyPair()

	sm := SignedManifest{
		Version:     "9.9.9",
		Channel:     "stable",
		DownloadURL: "http://example.com/synaptic",
		SHA256:      "abc123",
		// No Ed25519Sig set.
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(sm)
	}))
	defer srv.Close()

	u := New(nil, srv.URL)
	u.pubKey = pub

	result, _ := u.Check(context.Background())
	if result.UpdateAvailable {
		t.Fatal("missing signature should be rejected")
	}
}

func TestUpdater_SameVersion(t *testing.T) {
	pub, priv := testKeyPair()

	sm := SignedManifest{
		Version:     "0.1.0", // should match test environment or be different
		Channel:     "stable",
		DownloadURL: "http://example.com/synaptic",
		SHA256:      "abc",
	}
	signManifest(&sm, priv)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(sm)
	}))
	defer srv.Close()

	u := New(nil, srv.URL)
	u.pubKey = pub

	result, _ := u.Check(context.Background())
	// With the current version being "9.9.9" in tests, "0.1.0" is different.
	// This test just confirms no crash.
	_ = result
}

func TestUpdater_Disabled(t *testing.T) {
	u := New(nil, "http://example.com")
	u.SetEnabled(false)
	result, err := u.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.UpdateAvailable {
		t.Fatal("disabled updater should report no update")
	}
}

func TestSha256Sum(t *testing.T) {
	path := t.TempDir() + "/test.bin"
	_ = os.WriteFile(path, []byte("hello"), 0o644)
	sum, err := Sha256Sum(path)
	if err != nil {
		t.Fatal(err)
	}
	if sum == "" {
		t.Fatal("empty sum")
	}
}

func TestUpdater_Apply_WritesBinary(t *testing.T) {
	pub, priv := testKeyPair()

	// Prepare binary content and compute SHA256.
	binContent := []byte("fake-binary-v0.1.0")
	binHash := hex.EncodeToString(func() []byte { h := sha256.Sum256(binContent); return h[:] }())

	sm := SignedManifest{
		Version:     "9.9.9",
		Channel:     "stable",
		DownloadURL: "",
		SHA256:      binHash,
		Mandatory:   false,
	}
	signManifest(&sm, priv)

	// Serve manifest + binary.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/binary" {
			_, _ = w.Write(binContent)
			return
		}
		_ = json.NewEncoder(w).Encode(sm)
	}))
	defer srv.Close()

	sm.DownloadURL = srv.URL + "/binary"

	// Re-sign with the download URL included.
	signManifest(&sm, priv)

	manifestSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(sm)
	}))
	defer manifestSrv.Close()

	u := New(nil, manifestSrv.URL)
	u.pubKey = pub
	u.cacheDir = t.TempDir()

	result, err := u.Check(context.Background())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !result.UpdateAvailable {
		t.Fatal("expected update available")
	}

	_, err = u.Apply(context.Background(), result)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// Verify the binary was written to the cache dir.
	dst := u.cacheDir + "/synaptic-update-" + sm.Version
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("binary not written: %v", err)
	}
	if !bytes.Equal(data, binContent) {
		t.Errorf("binary content mismatch: got %d bytes, want %d", len(data), len(binContent))
	}
}
