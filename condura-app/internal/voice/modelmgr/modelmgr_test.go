package modelmgr

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureModel_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "ggml-base.bin")
	if err := os.WriteFile(modelPath, []byte("fake model data"), 0o644); err != nil {
		t.Fatal(err)
	}

	spec := ModelSpec{Name: "base", Filename: "ggml-base.bin"}
	got, err := EnsureModel(context.Background(), spec, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != modelPath {
		t.Errorf("expected %s, got %s", modelPath, got)
	}
}

func TestEnsureModel_DownloadsAndVerifies(t *testing.T) {
	content := []byte("fake whisper model content")
	hash := sha256.Sum256(content)
	expectedHash := hex.EncodeToString(hash[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(content)
	}))
	defer srv.Close()

	dir := t.TempDir()
	spec := ModelSpec{
		Name:     "base",
		URL:      srv.URL + "/ggml-base.bin",
		SHA256:   expectedHash,
		Filename: "ggml-base.bin",
	}

	got, err := EnsureModel(context.Background(), spec, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("read model: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Error("model content mismatch")
	}
}

func TestEnsureModel_ChecksumMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("wrong content"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	spec := ModelSpec{
		Name:     "base",
		URL:      srv.URL + "/ggml-base.bin",
		SHA256:   "0000000000000000000000000000000000000000000000000000000000000000",
		Filename: "ggml-base.bin",
	}

	_, err := EnsureModel(context.Background(), spec, dir)
	if err == nil {
		t.Fatal("expected checksum error")
	}
}

func TestEnsureModel_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dir := t.TempDir()
	spec := ModelSpec{
		Name:     "base",
		URL:      srv.URL + "/ggml-base.bin",
		Filename: "ggml-base.bin",
	}

	_, err := EnsureModel(context.Background(), spec, dir)
	if err == nil {
		t.Fatal("expected HTTP error")
	}
}

func TestModelForName(t *testing.T) {
	spec, err := ModelForName("base")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Name != "base" {
		t.Errorf("expected base, got %s", spec.Name)
	}

	_, err = ModelForName("unknown")
	if err == nil {
		t.Fatal("expected error for unknown model")
	}
}

func TestEnsureModel_NoChecksumSkipsVerification(t *testing.T) {
	content := []byte("model without checksum verification")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(content)
	}))
	defer srv.Close()

	dir := t.TempDir()
	spec := ModelSpec{
		Name:     "base",
		URL:      srv.URL + "/ggml-base.bin",
		SHA256:   "", // no checksum
		Filename: "ggml-base.bin",
	}

	got, err := EnsureModel(context.Background(), spec, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("read model: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Error("model content mismatch")
	}
}

// -----------------------------------------------------------------------------
// WakeModelForName — wake-word name dispatch
//
// Maps a user-supplied name string to its canonical WakeModelSpec.
// The "hey_synaptic" branch exists for backward compatibility with
// pre-rebrand installs; a regression that drops it would silently
// break upgrades for existing users. The "unknown name" branch
// returns an error message that lists the supported name so the
// caller can self-correct without grepping source.
// -----------------------------------------------------------------------------

func TestWakeModelForName_Canonical(t *testing.T) {
	got, err := WakeModelForName("hey_condura")
	if err != nil {
		t.Fatalf("WakeModelForName(\"hey_condura\"): unexpected error %v", err)
	}
	if got.Name != HeyConduraModel.Name {
		t.Errorf("WakeModelForName(\"hey_condura\").Name = %q, want %q", got.Name, HeyConduraModel.Name)
	}
	if got.URL != HeyConduraModel.URL {
		t.Errorf("WakeModelForName(\"hey_condura\").URL = %q, want %q", got.URL, HeyConduraModel.URL)
	}
}

func TestWakeModelForName_DeprecatedAliasStillSupported(t *testing.T) {
	// Pre-rebrand installs may have config files or saved state
	// referencing "hey_synaptic". The dispatch MUST accept this
	// and return the HeyConduraModel spec, matching the canonical
	// name. A regression that drops the alias would break upgrades
	// for users who had the app before the rebrand.
	got, err := WakeModelForName("hey_synaptic")
	if err != nil {
		t.Fatalf("WakeModelForName(\"hey_synaptic\"): unexpected error %v", err)
	}
	if got.Name != HeyConduraModel.Name {
		t.Errorf("deprecated alias returned Name %q, want %q (alias must map to canonical)", got.Name, HeyConduraModel.Name)
	}
}

func TestWakeModelForName_UnknownReturnsError(t *testing.T) {
	_, err := WakeModelForName("hey_nonexistent")
	if err == nil {
		t.Fatal("WakeModelForName on unknown name must return an error")
	}
	// The error message must list the supported name so the caller
	// can self-correct. A regression that dropped the hint would
	// leave users stuck without knowing what name to use.
	if !strings.Contains(err.Error(), "hey_condura") {
		t.Errorf("error message %q should mention supported name \"hey_condura\"", err.Error())
	}
}

// TestDownloadWakeModel_AlreadyExistsShortCircuits pins the
// "already downloaded" early-return branch in DownloadWakeModel.
// If the target file already exists with non-zero size, the
// function returns the path WITHOUT re-downloading or
// re-verifying — this is the "skip if cached" optimization.
//
// A regression that removed the size check would re-download
// every time the daemon starts; a regression that removed the
// existence check would overwrite a user-modified model.
// -----------------------------------------------------------------------------

func TestDownloadWakeModel_AlreadyExistsShortCircuits(t *testing.T) {
	dir := t.TempDir()
	// Pre-create the target file with non-zero content.
	target := filepath.Join(dir, HeyConduraModel.Filename)
	if err := os.WriteFile(target, []byte("preexisting-model-content"), 0o644); err != nil {
		t.Fatalf("seed model file: %v", err)
	}

	// Point URL at a deliberately-invalid address. If the function
	// actually attempts to download, it'll fail (no network in
	// tests). The pre-existing file MUST short-circuit the fetch.
	spec := HeyConduraModel
	spec.URL = "http://127.0.0.1:1/never-resolves"

	got, err := DownloadWakeModel(context.Background(), spec, dir)
	if err != nil {
		t.Fatalf("DownloadWakeModel with pre-existing file: unexpected error %v", err)
	}
	if got != target {
		t.Errorf("DownloadWakeModel returned %q, want %q (must short-circuit on existing file)", got, target)
	}
}
