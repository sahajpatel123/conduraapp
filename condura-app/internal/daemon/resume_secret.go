// Package daemon — resume secret manager.
//
// The resume secret closes P0-1: an in-process compromised conductor
// can read the IPC bearer token, but it cannot read the resume
// secret stored on disk (mode 0600) AND cannot synthesize a human
// at a terminal typing the secret into a CLI. The CLI subcommand
// `condura resume --confirm <ticket>` opens its own IPC client and
// calls halt.confirm_resume with the secret, so the human-confirmation
// path is OUT of the in-process trust boundary.
//
// Sources, in priority order:
//  1. CONDURA_RESUME_SECRET env var (headless / scripted / CI)
//  2. <DataDir>/resume.secret (auto-generated on first start)
//
// The secret is 32 random bytes hex-encoded (64 chars). Never logged.
// Never sent anywhere except halt.confirm_resume over the IPC channel,
// and only inside a constant-time compare.
package daemon

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// resumeSecretFileMode is the mode on disk for the resume.secret file.
const resumeSecretFileMode = 0o600

// ResumeSecretManager loads / auto-generates the human-confirmation
// secret used by the sticky resume flow.
type ResumeSecretManager struct {
	mu      sync.Mutex
	hex     string // hex-encoded secret; empty until first Load
	dataDir string
	envVar  string
}

// NewResumeSecretManager constructs a manager rooted at dataDir (e.g.
// ~/.condura). envOverride is the env var name to check first (e.g.
// "CONDURA_RESUME_SECRET"); pass "" to disable the env-override path.
func NewResumeSecretManager(dataDir, envOverride string) *ResumeSecretManager {
	return &ResumeSecretManager{dataDir: dataDir, envVar: envOverride}
}

// Load returns the secret. Order: env var > on-disk file > auto-generate
// and persist. Safe to call multiple times.
func (m *ResumeSecretManager) Load() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.hex != "" {
		return m.hex, nil
	}
	if m.envVar != "" {
		if v := os.Getenv(m.envVar); v != "" {
			m.hex = v
			return m.hex, nil
		}
	}
	if m.dataDir == "" {
		return "", errors.New("resume secret: no dataDir and no env var set")
	}
	// Try the file first.
	path := filepath.Join(m.dataDir, "resume.secret")
	if b, err := os.ReadFile(path); err == nil { //nolint:gosec // G304: path is filepath.Join(dataDir, "resume.secret"); dataDir is the same trusted root as the secrets file (initSubsystems).
		secret := string(b)
		if err := validateResumeSecret(secret); err != nil {
			return "", fmt.Errorf("resume secret file %s: %w", path, err)
		}
		m.hex = secret
		return m.hex, nil
	}
	// Auto-generate.
	raw := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return "", fmt.Errorf("resume secret: generate: %w", err)
	}
	secret := hex.EncodeToString(raw)
	if err := os.MkdirAll(m.dataDir, 0o700); err != nil {
		return "", fmt.Errorf("resume secret: mkdir: %w", err)
	}
	if err := os.WriteFile(path, []byte(secret), resumeSecretFileMode); err != nil {
		return "", fmt.Errorf("resume secret: write %s: %w", path, err)
	}
	m.hex = secret
	return m.hex, nil
}

// validateResumeSecret ensures the secret has the expected shape
// (32-byte hex, 64 chars). Returns an error otherwise.
func validateResumeSecret(s string) error {
	if len(s) != 64 {
		return fmt.Errorf("expected 64-char hex, got %d", len(s))
	}
	if _, err := hex.DecodeString(s); err != nil {
		return fmt.Errorf("not valid hex: %w", err)
	}
	return nil
}
