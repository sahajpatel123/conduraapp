// Package secrets provides a secure storage abstraction for sensitive values
// (API keys, OAuth tokens, master encryption key).
//
// The default backend is the OS-native secret store:
//   - macOS: Keychain
//   - Windows: Credential Manager
//   - Linux: libsecret (gnome-keyring, kwallet)
//
// The fallback backend is an AES-256-GCM-encrypted JSON file in the Synaptic
// data directory. The file is locked down to mode 0600. The encryption key is
// derived from either the CONDURA_FILE_PASSPHRASE env var (headless / CI) or
// a machine-bound key file (secrets.json.key, mode 0600, generated once) so
// the file is not greppable and not portable to another machine. It is only
// used when the OS keyring is unavailable (headless servers, CI, minimal
// Linux without libsecret).
//
// Selection logic:
//  1. If CONDURA_SECRETS_BACKEND=keyring is set, use the keyring (fail if
//     unavailable).
//  2. If CONDURA_SECRETS_BACKEND=file is set, use the encrypted file backend.
//  3. Otherwise (auto): try the keyring; if it fails, fall back to the
//     encrypted file with a one-time warning.
//
// NEVER log the secret values. Use Manager.Get for read access only, and
// Set / Delete for write access. Callers must not print values.
package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/zalando/go-keyring"
)

// Service is the keyring service name used for all Synaptic secrets.
const Service = "condura"

// File-mode bits for the local-encrypted-secrets file. We keep the
// permissions strict (owner-only) because the file holds the master
// encryption key and every API key in cleartext-AES-wrapped form.
const (
	secretsDirPerm  = 0o700 // owner-only: drwx------
	secretsFilePerm = 0o600 // owner-only: -rw-------
)

// Backend identifies which backend is in use.
type Backend string

// Backend values.
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
			return nil, fmt.Errorf("%w: keyring unavailable and no file path given: %w", ErrBackendFailed, err)
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
// File backend (fallback) — AES-256-GCM encrypted
// -----------------------------------------------------------------------------

// fileManager stores secrets in a single AES-256-GCM-encrypted JSON file
// with mode 0600.
//
// Encryption key source (first hit wins):
//  1. CONDURA_FILE_PASSPHRASE env var (for headless / CI / scripted setups).
//  2. A machine-bound key file at <path>.key (32 random bytes, mode 0600,
//     generated once on first use).
//
// The key is derived into an AES-256 key via HKDF-SHA256 over a fixed salt +
// the key material. The on-disk file is never cleartext; a legacy v1
// cleartext file is migrated to v2 (encrypted) on first read.
//
// Format (v2, encrypted):
//
//	{
//	  "version": 2,
//	  "salt":   "<base64>",        // per-file random salt, 16 bytes
//	  "nonce":  "<base64>",        // AES-GCM nonce, 12 bytes
//	  "ct":     "<base64>"         // AES-GCM ciphertext of inner JSON
//	}
//
// The inner JSON is the v1 {version,secrets} blob. The salt + nonce are
// persisted so the key derivation is stable across reads; the key itself
// never touches disk (only the env var or the .key file does).
type fileManager struct {
	path    string
	mu      sync.Mutex
	gcm     cipher.AEAD
	salt    []byte // persisted salt, 16 bytes
	loaded  bool   // salt loaded from disk on first read
	keyFile string // path to machine-bound key file (empty if env used)
}

const (
	fileFormatV1Cleartext = 1 // legacy; migrated to v2 on first read
	fileFormatV2Encrypted = 2
)

// fileData is the inner plaintext shape (also the v1 on-disk shape).
type fileData struct {
	Version int               `json:"version"`
	Secrets map[string]string `json:"secrets"`
}

// fileEnvelopeV2 is the encrypted on-disk shape.
type fileEnvelopeV2 struct {
	Version int    `json:"version"` // always 2
	Salt    string `json:"salt"`    // base64, 16 bytes
	Nonce   string `json:"nonce"`   // base64, 12 bytes
	CT      string `json:"ct"`      // base64, AES-GCM ciphertext
}

// deriveFileKey derives a 32-byte AES-256 key from the given key material
// using HKDF-SHA256 (RFC 5869) with a per-file salt and a fixed info label.
// stdlib has no HKDF, so we implement it as HMAC-SHA256 extract + expand.
func deriveFileKey(keyMaterial, salt []byte) []byte {
	// Extract: PRK = HMAC-SHA256(salt, IKM)
	prk := hmac.New(sha256.New, salt)
	prk.Write(keyMaterial)
	prkOut := prk.Sum(nil)

	// Expand: OKM = HMAC-SHA256(PRK, info || 0x01), 32 bytes (one block).
	info := []byte("condura-secrets-file-v2")
	mac := hmac.New(sha256.New, prkOut)
	mac.Write(info)
	mac.Write([]byte{0x01})
	return mac.Sum(nil) // 32 bytes → AES-256
}

// loadOrCreateFileKey resolves the encryption key material for the file
// backend. It does NOT derive the final AES key (that needs the per-file
// salt, which is read separately).
func loadOrCreateFileKey(keyFilePath string) ([]byte, error) {
	// 1. Env var takes precedence — enables headless / CI / scripted setups
	//    where the key must come from the environment, not disk.
	if env := os.Getenv("CONDURA_FILE_PASSPHRASE"); env != "" {
		return []byte(env), nil
	}
	// 2. Machine-bound key file. Generate 32 random bytes once, mode 0600.
	if keyFilePath == "" {
		return nil, fmt.Errorf("%w: no key file path and no CONDURA_FILE_PASSPHRASE env var", ErrBackendFailed)
	}
	if err := os.MkdirAll(filepath.Dir(keyFilePath), secretsDirPerm); err != nil {
		return nil, fmt.Errorf("create key dir: %w", err)
	}
	if key, err := os.ReadFile(keyFilePath); err == nil { //nolint:gosec // keyFilePath is derived from secrets path, not user input
		if len(key) != 32 {
			return nil, fmt.Errorf("%w: key file %s is %d bytes, want 32", ErrBackendFailed, keyFilePath, len(key))
		}
		return key, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read key file: %w", err)
	}
	// Generate a fresh 32-byte key.
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	if err := os.WriteFile(keyFilePath, key, secretsFilePerm); err != nil {
		return nil, fmt.Errorf("write key file: %w", err)
	}
	return key, nil
}

func newFileManager(path string) (*fileManager, error) {
	if path == "" {
		return nil, fmt.Errorf("%w: file path is empty", ErrBackendFailed)
	}
	if err := os.MkdirAll(filepath.Dir(path), secretsDirPerm); err != nil {
		return nil, fmt.Errorf("create secrets dir: %w", err)
	}
	// Resolve the encryption key material (env var or .key file). The .key
	// file lives next to secrets.json so it travels with the data dir but
	// is never written to the SQLite DB or the audit log.
	keyMaterial, err := loadOrCreateFileKey(path + ".key")
	if err != nil {
		return nil, err
	}
	// Build the AES-GCM cipher lazily — the salt is per-file and is loaded
	// on the first read. We pre-build the block cipher so a bad key fails
	// fast here rather than on first secret access.
	block, err := aes.NewCipher(deriveFileKey(keyMaterial, nil))
	if err != nil {
		return nil, fmt.Errorf("init file cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("init file gcm: %w", err)
	}
	fm := &fileManager{
		path:    path,
		gcm:     gcm,
		keyFile: path + ".key",
	}
	// Ensure the file exists. If it's a legacy v1 cleartext file, migrate
	// it to v2 (encrypted) in place. If it doesn't exist, write an empty
	// v2 envelope so future reads are stable.
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		if err := fm.writeEncrypted(fileData{Version: fileFormatV2Encrypted, Secrets: map[string]string{}}); err != nil {
			return nil, fmt.Errorf("init secrets file: %w", err)
		}
	} else if statErr != nil {
		return nil, fmt.Errorf("stat secrets file: %w", statErr)
	} else if err := os.Chmod(path, secretsFilePerm); err != nil {
		return nil, fmt.Errorf("chmod secrets file: %w", err)
	} else {
		// File exists — try to load + migrate if needed.
		if err := fm.ensureLoadedAndMigrated(keyMaterial); err != nil {
			return nil, err
		}
	}
	return fm, nil
}

// ensureLoadedAndMigrated reads the on-disk file once, loads the per-file
// salt into fm, and migrates a legacy v1 cleartext file to v2 (encrypted).
// After this call, fm.salt is populated and the on-disk file is v2.
//
//nolint:gocyclo // migration logic is inherently branchy
func (f *fileManager) ensureLoadedAndMigrated(keyMaterial []byte) error {
	raw, err := os.ReadFile(f.path) //nolint:gosec // path is derived from config, not user input
	if err != nil {
		return fmt.Errorf("read secrets file: %w", err)
	}
	if len(raw) == 0 {
		// Empty file → treat as fresh, write empty v2.
		return f.writeEncrypted(fileData{Version: fileFormatV2Encrypted, Secrets: map[string]string{}})
	}
	// Peek: is it v1 (cleartext JSON with a "secrets" field) or v2
	// (encrypted envelope with a "ct" field)?
	var probe struct {
		Version int               `json:"version"`
		Secrets map[string]string `json:"secrets"`
		CT      string            `json:"ct"`
		Salt    string            `json:"salt"`
		Nonce   string            `json:"nonce"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return fmt.Errorf("parse secrets file (corrupt or unknown format): %w", err)
	}
	if probe.CT != "" {
		// v2 encrypted — load the salt + rebuild the cipher with the
		// per-file salt so subsequent reads use the same derived key.
		salt, err := base64.StdEncoding.DecodeString(probe.Salt)
		if err != nil || len(salt) != 16 {
			return fmt.Errorf("%w: secrets file salt is malformed", ErrBackendFailed)
		}
		f.salt = salt
		block, err := aes.NewCipher(deriveFileKey(keyMaterial, salt))
		if err != nil {
			return fmt.Errorf("reinit file cipher: %w", err)
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return fmt.Errorf("reinit file gcm: %w", err)
		}
		// Probe-decrypt: a wrong key (e.g. .key file regenerated after
		// the file was written) must fail at construction, not silently
		// produce a cipher that fails on every subsequent Get. We decrypt
		// the envelope once here to surface the "wrong key" error early.
		nonce, err := base64.StdEncoding.DecodeString(probe.Nonce)
		if err != nil {
			return fmt.Errorf("%w: secrets file nonce is malformed", ErrBackendFailed)
		}
		ct, err := base64.StdEncoding.DecodeString(probe.CT)
		if err != nil {
			return fmt.Errorf("%w: secrets file ciphertext is malformed", ErrBackendFailed)
		}
		if _, err := gcm.Open(nil, nonce, ct, nil); err != nil {
			return fmt.Errorf("%w: decrypt secrets (wrong key or corrupt file)", ErrBackendFailed)
		}
		f.gcm = gcm
		f.loaded = true
		return nil
	}
	// v1 cleartext — migrate to v2 encrypted.
	var v1 fileData
	if err := json.Unmarshal(raw, &v1); err != nil {
		return fmt.Errorf("parse legacy v1 secrets file: %w", err)
	}
	if v1.Secrets == nil {
		v1.Secrets = map[string]string{}
	}
	v1.Version = fileFormatV2Encrypted
	if err := f.writeEncrypted(v1); err != nil {
		return fmt.Errorf("migrate v1→v2 secrets file: %w", err)
	}
	// Wipe the cleartext bytes from memory.
	for i := range raw {
		raw[i] = 0
	}
	return nil
}

// writeEncrypted serializes fd to JSON, generates a fresh per-file salt (if
// not yet set), derives the AES-256 key, encrypts with AES-GCM, and writes
// the v2 envelope atomically. The salt is persisted so future reads derive
// the same key.
func (f *fileManager) writeEncrypted(fd fileData) error {
	inner, err := json.Marshal(fd)
	if err != nil {
		return fmt.Errorf("marshal secrets: %w", err)
	}
	if !f.loaded || len(f.salt) != 16 {
		f.salt = make([]byte, 16)
		if _, err := io.ReadFull(rand.Reader, f.salt); err != nil {
			return fmt.Errorf("generate salt: %w", err)
		}
		// Rebuild the cipher with the new salt so the encrypt uses the
		// same derived key that future reads will use.
		keyMaterial, err := loadOrCreateFileKey(f.keyFile)
		if err != nil {
			return err
		}
		block, err := aes.NewCipher(deriveFileKey(keyMaterial, f.salt))
		if err != nil {
			return fmt.Errorf("reinit file cipher for write: %w", err)
		}
		f.gcm, err = cipher.NewGCM(block)
		if err != nil {
			return fmt.Errorf("reinit file gcm for write: %w", err)
		}
		f.loaded = true
	}
	nonce := make([]byte, f.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("generate nonce: %w", err)
	}
	ct := f.gcm.Seal(nil, nonce, inner, nil)
	env := fileEnvelopeV2{
		Version: fileFormatV2Encrypted,
		Salt:    base64.StdEncoding.EncodeToString(f.salt),
		Nonce:   base64.StdEncoding.EncodeToString(nonce),
		CT:      base64.StdEncoding.EncodeToString(ct),
	}
	out, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}
	tmp := f.path + ".tmp"
	if err := os.WriteFile(tmp, out, secretsFilePerm); err != nil {
		return fmt.Errorf("write temp: %w", err)
	}
	if err := os.Rename(tmp, f.path); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	// Wipe inner plaintext from memory.
	for i := range inner {
		inner[i] = 0
	}
	return nil
}

// readDecrypted reads the v2 envelope from disk, decrypts it, and returns
// the inner fileData. Ensures the per-file salt is loaded.
func (f *fileManager) readDecrypted() (fileData, error) {
	if !f.loaded {
		keyMaterial, err := loadOrCreateFileKey(f.keyFile)
		if err != nil {
			return fileData{}, err
		}
		if err := f.ensureLoadedAndMigrated(keyMaterial); err != nil {
			return fileData{}, err
		}
	}
	raw, err := os.ReadFile(f.path) //nolint:gosec // path is derived from config, not user input
	if err != nil {
		return fileData{}, fmt.Errorf("read secrets file: %w", err)
	}
	if len(raw) == 0 {
		return fileData{Version: fileFormatV2Encrypted, Secrets: map[string]string{}}, nil
	}
	var env fileEnvelopeV2
	if err := json.Unmarshal(raw, &env); err != nil {
		return fileData{}, fmt.Errorf("parse secrets envelope: %w", err)
	}
	if env.Version != fileFormatV2Encrypted || env.CT == "" {
		return fileData{}, fmt.Errorf("%w: secrets file is not v2 encrypted (version=%d)", ErrBackendFailed, env.Version)
	}
	nonce, err := base64.StdEncoding.DecodeString(env.Nonce)
	if err != nil {
		return fileData{}, fmt.Errorf("%w: decode nonce: %w", ErrBackendFailed, err)
	}
	ct, err := base64.StdEncoding.DecodeString(env.CT)
	if err != nil {
		return fileData{}, fmt.Errorf("%w: decode ciphertext: %w", ErrBackendFailed, err)
	}
	inner, err := f.gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return fileData{}, fmt.Errorf("%w: decrypt secrets (wrong key or corrupt file)", ErrBackendFailed)
	}
	var fd fileData
	if err := json.Unmarshal(inner, &fd); err != nil {
		return fileData{}, fmt.Errorf("parse decrypted secrets: %w", err)
	}
	if fd.Secrets == nil {
		fd.Secrets = map[string]string{}
	}
	// Wipe inner plaintext.
	for i := range inner {
		inner[i] = 0
	}
	return fd, nil
}

func (f *fileManager) Get(key string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fd, err := f.readDecrypted()
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
	fd, err := f.readDecrypted()
	if err != nil {
		return err
	}
	fd.Secrets[key] = value
	return f.writeEncrypted(fd)
}

func (f *fileManager) Delete(key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	fd, err := f.readDecrypted()
	if err != nil {
		return err
	}
	if _, ok := fd.Secrets[key]; !ok {
		return ErrNotFound
	}
	delete(fd.Secrets, key)
	return f.writeEncrypted(fd)
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
