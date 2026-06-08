package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/storage"
	"github.com/sahajpatel123/synapticapp/internal/version"
)

func setupUpdater(t *testing.T, manifestURL string) *Updater {
	t.Helper()
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{
		Path: filepath.Join(dir, "test.db"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return New(db.SQL(), manifestURL)
}

func TestUpdater_DefaultEnabled(t *testing.T) {
	u := setupUpdater(t, "")
	if !u.Enabled() {
		t.Fatal("updater should be enabled by default")
	}
}

func TestUpdater_PlatformKey(t *testing.T) {
	k := PlatformKey()
	if !strings.Contains(k, "-") {
		t.Fatalf("PlatformKey = %q, want os-arch", k)
	}
}

func TestUpdater_Check_NoManifest(t *testing.T) {
	u := setupUpdater(t, "")
	r, err := u.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if r.UpdateAvailable {
		t.Fatal("with no manifest URL, should not report update")
	}
	if !r.Skipped {
		t.Fatal("should report skipped=true")
	}
}

func TestUpdater_Check_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	u := setupUpdater(t, srv.URL)
	r, err := u.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !r.Skipped {
		t.Fatal("server error should be a skip, not a hard failure")
	}
}

func TestUpdater_Check_NoUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Same version as the running binary.
		_ = json.NewEncoder(w).Encode(Manifest{Version: version.Get().Version})
	}))
	defer srv.Close()
	u := setupUpdater(t, srv.URL)
	r, err := u.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if r.UpdateAvailable {
		t.Fatal("no update should be available for a matching version")
	}
}

func TestUpdater_Check_UpdateAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Manifest{
			Version:     "99.99.99",
			DownloadURL: "https://example.com/synaptic-99.99.99.dmg",
			SHA256:      "deadbeef",
			Mandatory:   true,
			Notes:       "Critical security fix",
		})
	}))
	defer srv.Close()
	u := setupUpdater(t, srv.URL)
	r, err := u.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !r.UpdateAvailable {
		t.Fatal("update should be available")
	}
	if !r.Mandatory {
		t.Fatal("manifest said mandatory, result should reflect that")
	}
	if r.LatestVersion != "99.99.99" {
		t.Fatalf("latest = %q, want 99.99.99", r.LatestVersion)
	}
	if !r.Forced {
		t.Fatal("forced should be true when enabled")
	}
}

func TestUpdater_Cached_AfterCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Manifest{Version: "99.99.99", DownloadURL: "https://x.test"})
	}))
	defer srv.Close()
	u := setupUpdater(t, srv.URL)
	if _, err := u.Check(context.Background()); err != nil {
		t.Fatal(err)
	}
	c, err := u.Cached()
	if err != nil {
		t.Fatal(err)
	}
	if c.LatestVersion != "99.99.99" {
		t.Fatalf("cached = %q, want 99.99.99", c.LatestVersion)
	}
}

func TestUpdater_Disabled_SkipsCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called when updater is disabled")
	}))
	defer srv.Close()
	u := setupUpdater(t, srv.URL)
	u.SetEnabled(false)
	r, err := u.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if r.Forced {
		t.Fatal("Forced should be false when updater is disabled")
	}
}

func TestHashFile(t *testing.T) {
	h := HashFile([]byte("hello"))
	if len(h) != 64 { // sha256 hex is 64 chars
		t.Fatalf("hash length = %d, want 64", len(h))
	}
}

func TestUpdater_Apply_NoUpdate(t *testing.T) {
	u := setupUpdater(t, "")
	_, err := u.Apply(context.Background(), Result{})
	if err == nil {
		t.Fatal("apply with no update_available should fail")
	}
}

func TestUpdater_Apply_MalformedManifest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()
	u := setupUpdater(t, srv.URL)
	_, err := u.Check(context.Background())
	if err == nil {
		t.Fatal("malformed manifest JSON should produce an error from Check")
	}
}

func TestUpdater_Check_NotModified(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotModified)
	}))
	defer srv.Close()
	u := setupUpdater(t, srv.URL)
	r, err := u.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !r.Skipped {
		t.Fatal("304 not modified should be a skip")
	}
}

func TestUpdater_Cached_NoDB(t *testing.T) {
	// Updater with a nil db: Cached should return a zero Result, not
	// panic or error.
	u := New(nil, "")
	r, err := u.Cached()
	if err != nil {
		t.Fatalf("Cached with nil db: %v", err)
	}
	if r.LatestVersion != "" {
		t.Fatalf("cached with nil db should be empty, got %q", r.LatestVersion)
	}
}
