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
	"path/filepath"
	"runtime"
	"strings"
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

	u := New(nil, srv.URL).SetURLSanitizeForTest(true)
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

	u := New(nil, srv.URL).SetURLSanitizeForTest(true)
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

	u := New(nil, srv.URL).SetURLSanitizeForTest(true)
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

	u := New(nil, srv.URL).SetURLSanitizeForTest(true)
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

	u := New(nil, manifestSrv.URL).SetURLSanitizeForTest(true)
	u.pubKey = pub
	u.cacheDir = t.TempDir()
	target := filepath.Join(t.TempDir(), "condurad")
	if err := os.WriteFile(target, []byte("old-binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	u.execPath = target

	result, err := u.Check(context.Background())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !result.UpdateAvailable {
		t.Fatal("expected update available")
	}

	applied, err := u.Apply(context.Background(), result)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !applied.Applied {
		t.Fatal("expected applied=true")
	}

	if runtime.GOOS == "windows" {
		if !applied.RestartRequired {
			t.Fatal("expected restart_required on windows")
		}
		dst := filepath.Join(u.cacheDir, "synaptic-update-"+sm.Version)
		got, err := os.ReadFile(dst)
		if err != nil {
			t.Fatalf("staged binary: %v", err)
		}
		if !bytes.Equal(got, binContent) {
			t.Errorf("staged binary content mismatch")
		}
		return
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("target binary: %v", err)
	}
	if !bytes.Equal(got, binContent) {
		t.Errorf("binary content mismatch after swap")
	}
}

// Audit 2026-07-01: updater manifest + download URLs now run
// through the strict URL sanitizer before any HTTP call. A
// tampered config (or downgrade default) that points the updater
// at a private/metadata IP must be refused at the sanitizer, not
// silently fetched.
func TestUpdater_RejectsMetadataIPManifest(t *testing.T) {
	pub, _ := testKeyPair()

	u := New(nil, "http://169.254.169.254/latest.json")
	u.pubKey = pub

	// The sanitizer is host-blocklist aware (cloud metadata IP).
	// Check must fail before any HTTP request.
	_, err := u.Check(context.Background())
	if err == nil {
		t.Fatal("Check accepted http://169.254.169.254/latest.json; expected SSRF refusal")
	}
	if !strings.Contains(err.Error(), "rejected") && !strings.Contains(err.Error(), "denied") {
		t.Errorf("error must explain rejection, got: %v", err)
	}
}

func TestUpdater_RejectsLoopbackManifest(t *testing.T) {
	pub, _ := testKeyPair()

	u := New(nil, "http://localhost/manifest.json")
	u.pubKey = pub

	_, err := u.Check(context.Background())
	if err == nil {
		t.Fatal("Check accepted http://localhost/manifest.json; expected SSRF refusal")
	}
}

func TestUpdater_RejectsPlainHTTPManifest(t *testing.T) {
	pub, _ := testKeyPair()

	u := New(nil, "http://example.com/manifest.json")
	u.pubKey = pub

	_, err := u.Check(context.Background())
	if err == nil {
		t.Fatal("Check accepted http:// (plain HTTP) manifest; expected protocol refusal")
	}
}

func TestSanitizeUpdaterURL_AcceptsHTTPSPublic(t *testing.T) {
	// Positive case for the helper. The httptest server URL is
	// plain HTTP, which the sanitizer rejects, so we test the
	// HTTPS shape directly to pin the "public URL passes" path.
	if _, err := sanitizeUpdaterURL("https://example.com/manifest.json"); err != nil {
		t.Errorf("HTTPS public URL must pass sanitizer, got: %v", err)
	}
}

func TestSanitizeUpdaterURL_RejectsBlocklist(t *testing.T) {
	cases := []string{
		"https://metadata.azure.com/",
		"https://metadata.aliyun.com/",
		"https://metadata.tencentyun.com/",
		"https://100.100.100.200/",
		"https://192.0.0.192/",
	}
	for _, u := range cases {
		if _, err := sanitizeUpdaterURL(u); err == nil {
			t.Errorf("blocklisted URL %q must be denied by sanitizeUpdaterURL", u)
		}
	}
}

func TestSanitizeUpdaterURL_EmptyPasses(t *testing.T) {
	// Empty string is the "no manifest configured" case — must
	// not error so the updater falls back to the default skip.
	if _, err := sanitizeUpdaterURL(""); err != nil {
		t.Errorf("empty URL must pass sanitizeUpdaterURL, got: %v", err)
	}
}
