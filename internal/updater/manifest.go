package updater

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"strings"
)

// PlatformArtifact is a signed download target for one GOOS/GOARCH pair.
type PlatformArtifact struct {
	DownloadURL string `json:"download_url"`
	SHA256      string `json:"sha256"`
}

// ManifestPayload is the canonical unsigned body that gets Ed25519-signed.
type ManifestPayload struct {
	Version     string                      `json:"version"`
	Channel     string                      `json:"channel"`
	DownloadURL string                      `json:"download_url,omitempty"`
	SHA256      string                      `json:"sha256,omitempty"`
	Platforms   map[string]PlatformArtifact `json:"platforms,omitempty"`
	Mandatory   bool                        `json:"mandatory"`
	MinVersion  string                      `json:"min_version,omitempty"`
	Notes       string                      `json:"notes,omitempty"`
}

// PlatformKey returns the runtime platform identifier (e.g. "darwin/arm64").
func PlatformKey() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

// ResolveArtifact returns the download URL and SHA256 for this machine.
// When Platforms is set, the top-level download_url/sha256 fields are ignored.
func (sm *SignedManifest) ResolveArtifact() (url, sha string, err error) {
	if len(sm.Platforms) > 0 {
		key := PlatformKey()
		a, ok := sm.Platforms[key]
		if !ok {
			return "", "", fmt.Errorf("updater: no artifact for platform %s", key)
		}
		if a.DownloadURL == "" || a.SHA256 == "" {
			return "", "", fmt.Errorf("updater: incomplete artifact for %s", key)
		}
		return a.DownloadURL, a.SHA256, nil
	}
	if sm.DownloadURL == "" {
		return "", "", fmt.Errorf("updater: no download_url in manifest")
	}
	return sm.DownloadURL, sm.SHA256, nil
}

// Payload returns the canonical unsigned manifest body for signing or verification.
func (sm *SignedManifest) Payload() (ManifestPayload, error) {
	p := ManifestPayload{
		Version:     sm.Version,
		Channel:     sm.Channel,
		DownloadURL: sm.DownloadURL,
		SHA256:      sm.SHA256,
		Mandatory:   sm.Mandatory,
		MinVersion:  sm.MinVersion,
		Notes:       sm.Notes,
	}
	if len(sm.Platforms) > 0 {
		p.Platforms = make(map[string]PlatformArtifact, len(sm.Platforms))
		for k, v := range sm.Platforms {
			p.Platforms[k] = v
		}
	}
	return p, nil
}

// MarshalPayload JSON-encodes the canonical unsigned manifest (stable key order).
func MarshalPayload(p ManifestPayload) ([]byte, error) {
	// Normalize: when platforms are present, omit legacy single-target fields
	// from the signed bytes so verifiers don't disagree on empty strings.
	if len(p.Platforms) > 0 {
		p.DownloadURL = ""
		p.SHA256 = ""
	}
	return json.Marshal(p)
}

// SignPayload signs the manifest payload with priv and returns hex-encoded sig.
func SignPayload(p ManifestPayload, priv ed25519.PrivateKey) (string, error) {
	msg, err := MarshalPayload(p)
	if err != nil {
		return "", err
	}
	sig := ed25519.Sign(priv, msg)
	return hex.EncodeToString(sig), nil
}

// VerifyPayload checks sig against pub for the given payload.
func VerifyPayload(p ManifestPayload, pub ed25519.PublicKey, sigHex string) error {
	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}
	msg, err := MarshalPayload(p)
	if err != nil {
		return err
	}
	if !ed25519.Verify(pub, msg, sig) {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}

// ChecksumEntry is one line from GoReleaser checksums.txt.
type ChecksumEntry struct {
	Hash string
	Name string
}

// ParseChecksums parses GoReleaser checksums.txt content.
func ParseChecksums(data string) ([]ChecksumEntry, error) {
	var out []ChecksumEntry
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			return nil, fmt.Errorf("bad checksum line: %q", line)
		}
		out = append(out, ChecksumEntry{Hash: strings.ToLower(parts[0]), Name: parts[len(parts)-1]})
	}
	return out, nil
}

// PlatformFromArchiveName extracts "linux/amd64" from synaptic-1.0.0-linux-amd64.tar.gz.
func PlatformFromArchiveName(name, project string) (string, bool) {
	base := strings.TrimSuffix(name, ".tar.gz")
	base = strings.TrimSuffix(base, ".zip")
	prefix := project + "-"
	if !strings.HasPrefix(base, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(base, prefix)
	// rest = 1.0.0-linux-amd64 or 1.0.0-darwin-arm64
	idx := strings.Index(rest, "-")
	if idx < 0 {
		return "", false
	}
	rest = rest[idx+1:]
	// linux-amd64
	hyphen := strings.LastIndex(rest, "-")
	if hyphen < 0 {
		return "", false
	}
	goos := rest[:hyphen]
	goarch := rest[hyphen+1:]
	if goos == "" || goarch == "" {
		return "", false
	}
	return goos + "/" + goarch, true
}

// BuildManifestFromChecksums builds a multi-platform manifest from release artifacts.
// baseURL is the GitHub release asset prefix (no trailing slash).
func BuildManifestFromChecksums(version, channel, baseURL, notes string, entries []ChecksumEntry) (ManifestPayload, error) {
	if version == "" {
		return ManifestPayload{}, fmt.Errorf("version required")
	}
	if channel == "" {
		channel = "stable"
	}
	platforms := make(map[string]PlatformArtifact)
	for _, e := range entries {
		// Auto-update targets synapticd archives (daemon/GUI host binary name).
		if !strings.Contains(e.Name, "synapticd-") || strings.Contains(e.Name, "synaptic-cli-") {
			continue
		}
		plat, ok := PlatformFromArchiveName(e.Name, "synapticd")
		if !ok {
			continue
		}
		platforms[plat] = PlatformArtifact{
			DownloadURL: baseURL + "/" + e.Name,
			SHA256:      e.Hash,
		}
	}
	if len(platforms) == 0 {
		return ManifestPayload{}, fmt.Errorf("no synapticd archives found in checksums")
	}
	keys := make([]string, 0, len(platforms))
	for k := range platforms {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return ManifestPayload{
		Version:   strings.TrimPrefix(version, "v"),
		Channel:   channel,
		Platforms: platforms,
		Mandatory: false,
		Notes:     notes,
	}, nil
}
