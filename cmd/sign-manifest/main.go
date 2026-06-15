// sign-manifest signs Synaptic update manifests for auto-update distribution.
package main

// sign-manifest signs a Synaptic update manifest JSON for GitHub Releases.
// Usage: sign-manifest <unsigned.json> <output.json>
// Requires UPDATE_SIGNING_KEY (hex-encoded Ed25519 private seed or PEM).
import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: sign-manifest <unsigned.json> <output.json>\n")
		os.Exit(2)
	}
	raw, err := os.ReadFile(os.Args[1]) //nolint:gosec // CLI tool; paths are operator-supplied
	if err != nil {
		fatal(err)
	}
	var m struct {
		Version     string `json:"version"`
		Channel     string `json:"channel"`
		DownloadURL string `json:"download_url"`
		SHA256      string `json:"sha256"`
		Mandatory   bool   `json:"mandatory"`
		MinVersion  string `json:"min_version,omitempty"`
		Notes       string `json:"notes,omitempty"`
		Ed25519Sig  string `json:"ed25519_sig,omitempty"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		fatal(err)
	}
	priv, err := loadPrivateKey(os.Getenv("UPDATE_SIGNING_KEY"))
	if err != nil {
		fatal(err)
	}
	payload, err := json.Marshal(struct {
		Version     string `json:"version"`
		Channel     string `json:"channel"`
		DownloadURL string `json:"download_url"`
		SHA256      string `json:"sha256"`
		Mandatory   bool   `json:"mandatory"`
		MinVersion  string `json:"min_version,omitempty"`
		Notes       string `json:"notes,omitempty"`
	}{
		Version: m.Version, Channel: m.Channel, DownloadURL: m.DownloadURL,
		SHA256: m.SHA256, Mandatory: m.Mandatory, MinVersion: m.MinVersion, Notes: m.Notes,
	})
	if err != nil {
		fatal(err)
	}
	sig := ed25519.Sign(priv, payload)
	m.Ed25519Sig = hex.EncodeToString(sig)
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(os.Args[2], out, 0o600); err != nil { //nolint:gosec // CLI tool
		fatal(err)
	}
}

func loadPrivateKey(raw string) (ed25519.PrivateKey, error) {
	if raw == "" {
		return nil, fmt.Errorf("UPDATE_SIGNING_KEY not set")
	}
	b, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode UPDATE_SIGNING_KEY: %w", err)
	}
	if len(b) == ed25519.SeedSize {
		return ed25519.NewKeyFromSeed(b), nil
	}
	if len(b) == ed25519.PrivateKeySize {
		return ed25519.PrivateKey(b), nil
	}
	return nil, fmt.Errorf("UPDATE_SIGNING_KEY must be %d-byte seed or %d-byte private key", ed25519.SeedSize, ed25519.PrivateKeySize)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "sign-manifest: %v\n", err)
	os.Exit(1)
}
