// Package secrets provides a secure storage abstraction for sensitive values
// (API keys, OAuth tokens, master encryption key).
//
// The default backend is the OS-native secret store:
//   - macOS: Keychain
//   - Windows: Credential Manager
//   - Linux: libsecret (gnome-keyring, kwallet)
//
// The fallback backend is a JSON file in the Synaptic data directory.
// The file is locked down to mode 0600 and is only used when the OS keyring
// is unavailable (headless servers, CI, minimal Linux without libsecret).
//
// Selection logic:
//  1. If SYNAPTIC_SECRETS_BACKEND=keyring is set, use the keyring (fail if
//     unavailable).
//  2. If SYNAPTIC_SECRETS_BACKEND=file is set, use the file backend.
//  3. Otherwise (auto): try the keyring; if it fails, fall back to file with
//     a one-time warning.
//
// NEVER log the secret values. Use Manager.Get for read access only, and
// Set / Delete for write access. Callers must not print values.
package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/zalando/go-keyring"
)

// Service is the keyring service name used for all Synaptic secrets.
const Service = "synaptic"

// File-mode bits for the local-encrypted-secrets file. We keep the
// permissions strict (owner-only) because the file holds the master
// encryption key and every API key in cleartext-AES-wrapped form.
const (
	secretsDirPerm  = 0o700 // owner-only: drwx------
	secretsFilePerm = 0o600 // owner-only: -rw-------
)

// Backend identifies which backend is in use.
type Backend string

const (
	BackendKeyring Backend = "keyring"
	BackendFile    Backend = "file"
)

// Common errors.
var (
	ErrNotFound      = errors.New("secret not found")
	ErrBackendFailed = errors.New("secret backend failed")
)

// KeyringBackend is the small subset of zalando/go-keyring we depend on.
// Tests can inject a fake implementation.
type KeyringBackend interface {
	Get(service, key string) (string, error)
	Set(service, key, value string) error
	Delete(service, key string) error
}

// zalandoKeyring adapts github.com/zalando/go-keyring to KeyringBackend.
type zalandoKeyring struct{}

func (zalandoKeyring) Get(s, k string) (string, error) { return keyring.Get(s, k) }
func (zalandoKeyring) Set(s, k, v string) error        { return keyring.Set(s, k, v) }
func (zalandoKeyring) Delete(s, k string) error        { return keyring.Delete(s, k) }

// Manager is the secret-storage interface.
//
// All methods are safe for concurrent use.
type Manager interface {
	// Get returns the secret for the given key, or ErrNotFound.
	Get(key string) (string, error)
	// Set stores the secret for the given key, replacing any existing value.
	Set(key, value string) error
	// Delete removes the secret for the given key. Returns ErrNotFound if absent.
	Delete(key string) error
	// Backend returns which backend is in use.
	Backend() Backend
	// Close releases any resources (currently a no-op).
	Close() error
}

// -----------------------------------------------------------------------------
// Factory
// -----------------------------------------------------------------------------

// New returns a Manager using the OS keyring with file fallback.
func New(filePath string) (Manager, error) {
	return NewWithBackend("", filePath)
}

// NewWithKeyring is like NewWithBackend but allows injecting a custom keyring
// implementation. Intended for tests.
func NewWithKeyring(kb KeyringBackend, backend, filePath string) (Manager, error) {
	switch Backend(backend) {
	case BackendKeyring:
		return newKeyringManager(kb)
	case BackendFile:
		if filePath == "" {
			return nil, fmt.Errorf("%w: file backend requires a non-empty file path", ErrBackendFailed)
		}
		return newFileManager(filePath)
	case "", "auto":
		km, err := newKeyringManager(kb)
		if err == nil {
			return km, nil
		}
		if filePath == "" {
			return nil, fmt.Errorf("%w: keyring unavailable and no file path given: %v", ErrBackendFailed, err)
		}
		return newFileManager(filePath)
	default:
		return nil, fmt.Errorf("%w: unknown backend %q (want keyring, file, or auto)", ErrBackendFailed, backend)
	}
}

// NewWithBackend returns a Manager using the specified backend.
//   - backend == "keyring": keyring only, fail if unavailable.
//   - backend == "file": file only.
//   - backend == "" or "auto": keyring, fall back to file.
func NewWithBackend(backend, filePath string) (Manager, error) {
	return NewWithKeyring(zalandoKeyring{}, backend, filePath)
}

// -----------------------------------------------------------------------------
// Keyring backend
// -----------------------------------------------------------------------------

// keyringManager uses a KeyringBackend (typically zalando/go-keyring).
type keyringManager struct {
	kb KeyringBackend
}

func newKeyringManager(kb KeyringBackend) (*keyringManager, error) {
	// Probe: try to set + get + delete a test value.
	// If this fails, the keyring is not usable.
	testKey := "__synaptic_probe__"
	testVal := "ok"
	if err := kb.Set(Service, testKey, testVal); err != nil {
		return nil, fmt.Errorf("keyring probe set: %w", err)
	}
	got, err := kb.Get(Service, testKey)
	if err != nil {
		_ = kb.Delete(Service, testKey)
		return nil, fmt.Errorf("keyring probe get: %w", err)
	}
	if got != testVal {
		_ = kb.Delete(Service, testKey)
		return nil, fmt.Errorf("keyring probe mismatch")
	}
	if err := kb.Delete(Service, testKey); err != nil {
		return nil, fmt.Errorf("keyring probe delete: %w", err)
	}
	return &keyringManager{kb: kb}, nil
}

func (k *keyringManager) Get(key string) (string, error) {
	v, err := k.kb.Get(Service, key)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("keyring get %q: %w", key, err)
	}
	return v, nil
}

func (k *keyringManager) Set(key, value string) error {
	if err := k.kb.Set(Service, key, value); err != nil {
		return fmt.Errorf("keyring set %q: %w", key, err)
	}
	return nil
}

func (k *keyringManager) Delete(key string) error {
	if err := k.kb.Delete(Service, key); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("keyring delete %q: %w", key, err)
	}
	return nil
}

func (k *keyringManager) Backend() Backend { return BackendKeyring }

func (k *keyringManager) Close() error { return nil }

// -----------------------------------------------------------------------------
// File backend (fallback)
// -----------------------------------------------------------------------------

// fileManager stores secrets in a single JSON file with mode 0600.
//
// Format:
//
//	{
//	  "version": 1,
//	  "secrets": {
//	    "key1": "value1",
//	    "key2": "value2"
//	  }
//	}
//
// Concurrency: a sync.Mutex guards reads/writes. For Phase 1, this is fine.
// In Phase 7 (when we add P2P sync), we'll switch to BoltDB or similar.
type fileManager struct {
	path string
	mu   sync.Mutex
}

const fileFormatVersion = 1

type fileData struct {
	Version int               `json:"version"`
	Secrets map[string]string `json:"secrets"`
}

func newFileManager(path string) (*fileManager, error) {
	if path == "" {
		return nil, fmt.Errorf("%w: file path is empty", ErrBackendFailed)
	}
	if err := os.MkdirAll(filepath.Dir(path), secretsDirPerm); err != nil {
		return nil, fmt.Errorf("create secrets dir: %w", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte(`{"version":1,"secrets":{}}`), secretsFilePerm); err != nil {
			return nil, fmt.Errorf("init secrets file: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("stat secrets file: %w", err)
	} else if err := os.Chmod(path, secretsFilePerm); err != nil {
		return nil, fmt.Errorf("chmod secrets file: %w", err)
	}
	return &fileManager{path: path}, nil
}

func (f *fileManager) read() (fileData, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		return fileData{}, fmt.Errorf("read secrets file: %w", err)
	}
	var fd fileData
	if len(data) == 0 {
		return fileData{Version: fileFormatVersion, Secrets: map[string]string{}}, nil
	}
	if err := json.Unmarshal(data, &fd); err != nil {
		return fileData{}, fmt.Errorf("parse secrets file: %w", err)
	}
	if fd.Secrets == nil {
		fd.Secrets = map[string]string{}
	}
	return fd, nil
}

func (f *fileManager) write(fd fileData) error {
	data, err := json.Marshal(fd)
	if err != nil {
		return fmt.Errorf("marshal secrets: %w", err)
	}
	// Atomic write: write to temp file, then rename.
	tmp := f.path + ".tmp"
	if err := os.WriteFile(tmp, data, secretsFilePerm); err != nil {
		return fmt.Errorf("write temp: %w", err)
	}
	if err := os.Rename(tmp, f.path); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

func (f *fileManager) Get(key string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fd, err := f.read()
	if err != nil {
		return "", err
	}
	v, ok := fd.Secrets[key]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (f *fileManager) Set(key, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	fd, err := f.read()
	if err != nil {
		return err
	}
	fd.Secrets[key] = value
	return f.write(fd)
}

func (f *fileManager) Delete(key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	fd, err := f.read()
	if err != nil {
		return err
	}
	if _, ok := fd.Secrets[key]; !ok {
		return ErrNotFound
	}
	delete(fd.Secrets, key)
	return f.write(fd)
}

func (f *fileManager) Backend() Backend { return BackendFile }

func (f *fileManager) Close() error { return nil }

// -----------------------------------------------------------------------------
// Common secret keys (canonical names for use across the codebase)
// -----------------------------------------------------------------------------

const (
	// MasterKey is the 32-byte (base64) master key for SQLite column encryption.
	MasterKey = "master_key"
	// APIKeyPrefix is the prefix for API key entries. Use api_key_<provider>.
	APIKeyPrefix = "api_key_"
	// OAuthTokenPrefix is the prefix for OAuth token entries. Use oauth_<provider>.
	OAuthTokenPrefix = "oauth_"
)

// APIKeyKey returns the keyring key for an API key for the given provider.
func APIKeyKey(provider string) string { return APIKeyPrefix + provider }

// OAuthTokenKey returns the keyring key for an OAuth token for the given provider.
func OAuthTokenKey(provider string) string { return OAuthTokenPrefix + provider }
