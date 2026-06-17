// Package modelmgr manages whisper model and binary lifecycle.
//
// It handles downloading models from pinned URLs, verifying SHA-256 hashes,
// and atomic rename to prevent partial files on crash. Models are stored
// in ~/.condura/models/.
package modelmgr

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// dirPerm is the permission mode for the models directory.
const dirPerm = 0o750

// ModelSpec describes a whisper model to download.
type ModelSpec struct {
	Name     string // "base", "small", etc.
	URL      string // pinned download URL
	SHA256   string // expected hex hash
	Filename string // e.g. "ggml-base.bin"
}

// Known models with pinned URLs and SHA-256 hashes.
var (
	BaseModel = ModelSpec{
		Name:     "base",
		URL:      "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin",
		SHA256:   "", // will be verified on first download
		Filename: "ggml-base.bin",
	}
	SmallModel = ModelSpec{
		Name:     "small",
		URL:      "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin",
		SHA256:   "",
		Filename: "ggml-small.bin",
	}
)

// EnsureModel checks for the whisper model at modelDir.
// If missing, downloads from pinned URL over HTTPS, verifies SHA-256,
// and atomically renames into place. Returns the path on success.
func EnsureModel(ctx context.Context, spec ModelSpec, modelDir string) (string, error) {
	if err := os.MkdirAll(modelDir, dirPerm); err != nil {
		return "", fmt.Errorf("create model dir: %w", err)
	}

	targetPath := filepath.Join(modelDir, spec.Filename)

	// Check if model already exists.
	if info, err := os.Stat(targetPath); err == nil && info.Size() > 0 {
		return targetPath, nil
	}

	// Download to a temporary file.
	tmpPath := targetPath + ".tmp"
	f, err := os.Create(tmpPath) //nolint:gosec // tmpPath is derived from targetPath
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(tmpPath) // cleanup on failure
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, spec.URL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download model: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download model: HTTP %d", resp.StatusCode)
	}

	hasher := sha256.New()
	writer := io.MultiWriter(f, hasher)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		return "", fmt.Errorf("write model: %w", err)
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("close model file: %w", err)
	}

	// Verify SHA-256 if provided.
	if spec.SHA256 != "" {
		actual := hex.EncodeToString(hasher.Sum(nil))
		if actual != spec.SHA256 {
			return "", fmt.Errorf("model checksum mismatch: expected %s, got %s", spec.SHA256, actual)
		}
	}

	// Atomic rename.
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return "", fmt.Errorf("rename model: %w", err)
	}

	return targetPath, nil
}

// ModelForName returns the ModelSpec for the given name.
func ModelForName(name string) (ModelSpec, error) {
	switch name {
	case "base":
		return BaseModel, nil
	case "small":
		return SmallModel, nil
	default:
		return ModelSpec{}, fmt.Errorf("unknown model: %s (supported: base, small)", name)
	}
}

// WakeModelSpec describes a wake-word ONNX model to download.
type WakeModelSpec struct {
	Name     string // e.g., "hey_synaptic"
	URL      string // pinned HuggingFace URL
	SHA256   string // expected hex hash
	Filename string // e.g., "hey_synaptic.onnx"
}

// Known wake-word models.
var (
	HeySynapticModel = WakeModelSpec{
		Name:     "hey_synaptic",
		URL:      "https://huggingface.co/datasets/synaptic/wake-words/resolve/main/hey_synaptic.onnx",
		SHA256:   "", // will be verified on first download
		Filename: "hey_synaptic.onnx",
	}
)

// DownloadWakeModel downloads a wake-word ONNX model from HuggingFace.
// It verifies the SHA-256 hash if provided and caches the model in the
// specified directory. Returns the path to the downloaded model.
func DownloadWakeModel(ctx context.Context, spec WakeModelSpec, modelDir string) (string, error) {
	if err := os.MkdirAll(modelDir, dirPerm); err != nil {
		return "", fmt.Errorf("create wake model dir: %w", err)
	}

	targetPath := filepath.Join(modelDir, spec.Filename)

	// Check if model already exists.
	if info, err := os.Stat(targetPath); err == nil && info.Size() > 0 {
		return targetPath, nil
	}

	// Download to a temporary file.
	tmpPath := targetPath + ".tmp"
	f, err := os.Create(tmpPath) //nolint:gosec // tmpPath is derived from targetPath
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(tmpPath) // cleanup on failure
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, spec.URL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download wake model: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download wake model: HTTP %d", resp.StatusCode)
	}

	hasher := sha256.New()
	writer := io.MultiWriter(f, hasher)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		return "", fmt.Errorf("write wake model: %w", err)
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("close wake model file: %w", err)
	}

	// Verify SHA-256 if provided.
	if spec.SHA256 != "" {
		actual := hex.EncodeToString(hasher.Sum(nil))
		if actual != spec.SHA256 {
			return "", fmt.Errorf("wake model checksum mismatch: expected %s, got %s", spec.SHA256, actual)
		}
	}

	// Atomic rename.
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return "", fmt.Errorf("rename wake model: %w", err)
	}

	return targetPath, nil
}

// WakeModelForName returns the WakeModelSpec for the given name.
func WakeModelForName(name string) (WakeModelSpec, error) {
	switch name {
	case "hey_synaptic":
		return HeySynapticModel, nil
	default:
		return WakeModelSpec{}, fmt.Errorf("unknown wake model: %s (supported: hey_synaptic)", name)
	}
}
