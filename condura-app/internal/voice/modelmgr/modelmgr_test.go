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
