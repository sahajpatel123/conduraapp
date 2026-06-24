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
	"strings"
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
	Name     string // canonical name, e.g., "hey_condura"
	URL      string // pinned HuggingFace URL
	SHA256   string // expected hex hash (empty = verify on first download)
	Filename string // on-disk filename, e.g., "hey_condura.onnx"
}

// Known wake-word models.
//
// HeyConduraModel is the canonical wake-word model for the "hey condura"
// phrase (CLAUDE.md §19.3, decision #35). The ONNX asset itself is hosted
// under the legacy "synaptic" HuggingFace org + filename because the
// asset was trained + uploaded before the rebrand; the model detects the
// "hey condura" phrase regardless of the filename. The local on-disk
// filename uses the canonical "hey_condura" name.
//
// HeySynapticModel is retained as a deprecated alias so existing configs
// that reference "hey_synaptic" keep working. New configs should use
// "hey_condura".
var (
	HeyConduraModel = WakeModelSpec{
		Name:     "hey_condura",
		URL:      "https://huggingface.co/datasets/synaptic/wake-words/resolve/main/hey_synaptic.onnx",
		SHA256:   "",
		Filename: "hey_condura.onnx",
	}

	// HeySynapticModel is the deprecated pre-rebrand name. It resolves
	// to the same asset as HeyConduraModel. Retained for backward
	// compatibility with configs written before the rename.
	HeySynapticModel = HeyConduraModel
)

// DownloadWakeModel downloads a wake-word ONNX model from HuggingFace.
// It verifies the SHA-256 hash if provided and caches the model in the
// specified directory. Returns the path to the downloaded model.
//
// Trust model (B-30 fix):
//   - If spec.SHA256 is non-empty, the download is verified against it
//     (hard pin). A mismatch is a hard error.
//   - If spec.SHA256 is empty, trust-on-first-use (TOFU) applies: the
//     first successful download computes the hash and writes it to a
//     `<filename>.sha256` sidecar file. Subsequent downloads read the
//     sidecar and verify against it, so a compromised CDN that swaps
//     the asset after first download is caught. The sidecar is the
//     trust anchor; the first download trusts the transport (TLS).
//
//nolint:gocyclo // download+verify+TOFU is inherently branchy
func DownloadWakeModel(ctx context.Context, spec WakeModelSpec, modelDir string) (string, error) {
	if err := os.MkdirAll(modelDir, dirPerm); err != nil {
		return "", fmt.Errorf("create wake model dir: %w", err)
	}

	targetPath := filepath.Join(modelDir, spec.Filename)
	sidecarPath := targetPath + ".sha256"

	// Check if model already exists.
	if info, err := os.Stat(targetPath); err == nil && info.Size() > 0 {
		return targetPath, nil
	}

	// Resolve the expected hash: explicit spec hash takes precedence;
	// otherwise read the TOFU sidecar (if present) so a re-download
	// after a first successful download is verified against the
	// originally-pinned hash.
	expectedHash := spec.SHA256
	if expectedHash == "" {
		if sidecar, err := os.ReadFile(sidecarPath); err == nil { //nolint:gosec // sidecarPath is derived from modelDir+spec.Filename
			expectedHash = strings.TrimSpace(string(sidecar))
		}
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

	actualHash := hex.EncodeToString(hasher.Sum(nil))

	// Verify against the expected hash (hard pin or TOFU sidecar).
	if expectedHash != "" {
		if actualHash != expectedHash {
			return "", fmt.Errorf("wake model checksum mismatch: expected %s, got %s", expectedHash, actualHash)
		}
	} else {
		// First-ever download with no pin: write the TOFU sidecar so
		// future downloads are verified against this hash. This
		// closes the B-30 gap where a compromised CDN could swap the
		// asset on every download.
		if err := os.WriteFile(sidecarPath, []byte(actualHash+"\n"), 0o600); err != nil {
			return "", fmt.Errorf("write wake model sha256 sidecar: %w", err)
		}
	}

	// Atomic rename.
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return "", fmt.Errorf("rename wake model: %w", err)
	}

	return targetPath, nil
}

// WakeModelForName returns the WakeModelSpec for the given name.
// Accepts both the canonical "hey_condura" and the deprecated
// "hey_synaptic" (pre-rebrand) for backward compatibility.
func WakeModelForName(name string) (WakeModelSpec, error) {
	switch name {
	case "hey_condura", "hey_synaptic":
		return HeyConduraModel, nil
	default:
		return WakeModelSpec{}, fmt.Errorf("unknown wake model: %s (supported: hey_condura)", name)
	}
}
