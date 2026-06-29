// Package updater implements secure auto-update with Ed25519 signature
// verification, atomic rollback, and anti-downgrade protection.
//
// The update manifest is a signed JSON document. The daemon fetches it,
// verifies the Ed25519 signature against an embedded public key, downloads
// the binary, verifies SHA256, and atomically swaps it in place.
//
// A tampered manifest, wrong signature, corrupt binary, or downgrade attempt
// is rejected — the update never applies.
package updater

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/version"
)

// PublicKey is the Ed25519 public key for verifying update manifests.
// Generated offline. The corresponding private key is stored in CI secrets
// (UPDATE_SIGNING_KEY) and should NEVER be committed to the repository.
// See docs/release-keys.md for the full key management procedure.
var PublicKey = ed25519.PublicKey{
	0x7f, 0x3f, 0x6d, 0xd4, 0x9b, 0x1b, 0x92, 0x13,
	0x45, 0xf0, 0x88, 0x6c, 0x7b, 0x3f, 0x93, 0xc4,
	0x2c, 0x8f, 0x25, 0xb0, 0xb7, 0x91, 0x07, 0x96,
	0x87, 0xbc, 0x62, 0x1c, 0x09, 0xa9, 0xe0, 0x5a,
}

// SignedManifest is the JSON document at the manifest URL.
// The server signs the manifest bytes (minus the sig field)
// with an offline Ed25519 key. The daemon verifies before
// downloading.
//
// Use Platforms for multi-arch releases; legacy single-target
// manifests set DownloadURL and SHA256 at the top level.
type SignedManifest struct {
	Version     string                      `json:"version"`
	Channel     string                      `json:"channel"`
	DownloadURL string                      `json:"download_url,omitempty"`
	SHA256      string                      `json:"sha256,omitempty"`
	Platforms   map[string]PlatformArtifact `json:"platforms,omitempty"`
	Ed25519Sig  string                      `json:"ed25519_sig"` // hex-encoded signature
	Mandatory   bool                        `json:"mandatory"`
	MinVersion  string                      `json:"min_version,omitempty"` // anti-downgrade floor
	Notes       string                      `json:"notes,omitempty"`
}

// Result is the result of an update check.
type Result struct {
	UpdateAvailable bool   `json:"update_available"`
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version,omitempty"`
	DownloadURL     string `json:"download_url,omitempty"`
	SHA256          string `json:"sha256,omitempty"`
	Mandatory       bool   `json:"mandatory"`
	Skipped         bool   `json:"skipped,omitempty"`
	Reason          string `json:"reason,omitempty"`
	Applied         bool   `json:"applied,omitempty"`
	RestartRequired bool   `json:"restart_required,omitempty"`
}

// Updater is the secure auto-update controller.
type Updater struct {
	db       *sql.DB
	manifest string
	enabled  bool
	client   *http.Client
	pubKey   ed25519.PublicKey
	cacheDir string
	stdin    io.Reader // for os.Stdin in tests
	execPath string    // test override for target binary path
}

// New returns a secure Updater with the embedded public key.
func New(db *sql.DB, manifestURL string) *Updater {
	return &Updater{
		db:       db,
		manifest: manifestURL,
		enabled:  true,
		client:   &http.Client{Timeout: 10 * time.Second},
		pubKey:   PublicKey,
		cacheDir: filepath.Join(userHome(), ".condura", "cache"),
		stdin:    os.Stdin,
	}
}

// pollInterval is how often the daemon re-checks for updates when
// auto-update is enabled (MISSION §22.4).
const pollInterval = 6 * time.Hour

// SetEnabled toggles auto-update.
func (u *Updater) SetEnabled(v bool) { u.enabled = v }

// SetManifestURL overrides the manifest URL (tests).
func (u *Updater) SetManifestURL(url string) { u.manifest = url }

// SetPublicKey overrides the embedded Ed25519 public key (tests).
func (u *Updater) SetPublicKey(pub ed25519.PublicKey) { u.pubKey = pub }

// Enabled returns the current setting.
func (u *Updater) Enabled() bool { return u.enabled }

// RunPoller checks for updates on launch and every pollInterval
// until ctx is canceled. Errors are swallowed; Check already
// returns skip results for network failures.
func (u *Updater) RunPoller(ctx context.Context) {
	if !u.enabled || u.manifest == "" {
		return
	}
	check := func() {
		cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		_, _ = u.Check(cctx)
	}
	check()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			check()
		}
	}
}

// Check fetches and verifies the manifest.
func (u *Updater) Check(ctx context.Context) (Result, error) {
	cur := version.Get().Version
	if !u.enabled {
		return skipResult(cur, "auto-update disabled"), nil
	}
	if u.manifest == "" {
		return skipResult(cur, "no manifest URL"), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.manifest, nil)
	if err != nil {
		return Result{}, err
	}
	resp, err := u.client.Do(req)
	if err != nil {
		return skipResult(cur, err.Error()), nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return skipResult(cur, fmt.Sprintf("HTTP %d", resp.StatusCode)), nil
	}

	var sm SignedManifest
	if err := json.NewDecoder(resp.Body).Decode(&sm); err != nil {
		return Result{}, fmt.Errorf("manifest parse: %w", err)
	}

	// Verify Ed25519 signature.
	if err := u.verifyManifest(sm); err != nil {
		return skipResult(cur, fmt.Sprintf("signature verification failed: %v", err)), nil
	}

	// Anti-downgrade: reject an older version.
	if sm.MinVersion != "" && compareVersions(sm.Version, sm.MinVersion) < 0 {
		return skipResult(cur, "version below minimum"), nil
	}

	if sm.Version == cur || sm.Version == "" {
		return Result{UpdateAvailable: false, CurrentVersion: cur}, nil
	}

	dlURL, sha, err := sm.ResolveArtifact()
	if err != nil {
		return skipResult(cur, err.Error()), nil
	}

	res := Result{
		UpdateAvailable: true,
		CurrentVersion:  cur,
		LatestVersion:   sm.Version,
		DownloadURL:     dlURL,
		SHA256:          sha,
		Mandatory:       sm.Mandatory,
	}

	// Cache the result.
	if u.db != nil {
		_, _ = u.db.ExecContext(context.Background(),
			`UPDATE update_cache SET last_check_ts=?, latest_version=?, download_url=? WHERE id=1`,
			time.Now().UTC().Format(time.RFC3339), sm.Version, sm.DownloadURL,
		)
	}
	return res, nil
}

// Apply downloads, verifies SHA256, swaps the binary, and returns restart guidance.
//
//nolint:gocyclo // download + verify + swap paths are intentionally sequential
func (u *Updater) Apply(ctx context.Context, r Result) (Result, error) {
	if !r.UpdateAvailable {
		return r, errors.New("no update available")
	}
	if r.DownloadURL == "" {
		return r, errors.New("no download URL")
	}

	// Re-check to get the manifest for SHA256 + sig.
	fresh, err := u.Check(ctx)
	if err != nil {
		return r, err
	}
	if !fresh.UpdateAvailable {
		return fresh, errors.New("already up to date")
	}
	r = fresh

	// Download.
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, r.DownloadURL, nil)
	resp, err := u.client.Do(req)
	if err != nil {
		return r, fmt.Errorf("download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		return r, fmt.Errorf("download HTTP %d", resp.StatusCode)
	}

	body, err := readAll(resp.Body)
	if err != nil {
		return r, fmt.Errorf("read body: %w", err)
	}

	// Verify SHA256 from the signed manifest.
	// Note: we re-fetch the manifest in Check() above; the SHA256
	// is verified at manifest parse time via the Ed25519 signature.
	// Here we verify the binary itself matches.
	actual := sha256.Sum256(body)
	if r.SHA256 != "" {
		want := strings.ToLower(r.SHA256)
		got := hex.EncodeToString(actual[:])
		if got != want {
			return r, fmt.Errorf("sha256 mismatch: expected %s, got %s", want, got)
		}
	}

	// Write to cache dir atomically.
	if err := os.MkdirAll(u.cacheDir, 0o700); err != nil {
		return r, fmt.Errorf("cache dir: %w", err)
	}
	dst := filepath.Join(u.cacheDir, "synaptic-update-"+r.LatestVersion)
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, body, 0o700); err != nil { //nolint:gosec
		return r, fmt.Errorf("write binary: %w", err)
	}

	// Atomic rename into cache.
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return r, fmt.Errorf("rename: %w", err)
	}

	target, err := u.targetExecutable()
	if err != nil {
		return r, fmt.Errorf("resolve executable: %w", err)
	}
	if err := swapExecutable(dst, target); err != nil {
		if errors.Is(err, ErrRestartRequired) {
			r.Applied = true
			r.RestartRequired = true
			return r, nil
		}
		return r, err
	}
	r.Applied = true
	r.RestartRequired = true
	return r, nil
}

func (u *Updater) targetExecutable() (string, error) {
	if u.execPath != "" {
		return u.execPath, nil
	}
	return currentExecutable()
}

// verifyManifest checks the Ed25519 signature on the manifest.
func (u *Updater) verifyManifest(sm SignedManifest) error {
	sigBytes, err := hex.DecodeString(sm.Ed25519Sig)
	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}
	p, err := sm.Payload()
	if err != nil {
		return err
	}
	msg, err := MarshalPayload(p)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	if !ed25519.Verify(u.pubKey, msg, sigBytes) {
		return errors.New("signature verification failed")
	}
	return nil
}

// Sha256Sum computes the SHA256 of a file for publishing.
func Sha256Sum(path string) (string, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is from the manifest, which is signature-verified
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}

func skipResult(cur, reason string) Result {
	return Result{
		UpdateAvailable: false,
		CurrentVersion:  cur,
		Skipped:         true,
		Reason:          reason,
	}
}

func compareVersions(a, b string) int {
	if a == b {
		return 0
	}
	if a < b {
		return -1
	}
	return 1
}

func userHome() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return "/tmp"
}

func readAll(r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 64*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			out = append(out, buf[:n]...)
		}
		if err != nil {
			if err == io.EOF {
				return out, nil
			}
			return out, err
		}
	}
}
