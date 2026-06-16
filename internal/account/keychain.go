package account

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// fileTokenManager stores tokens in an encrypted file instead of
// the OS keychain. Used as fallback on headless Linux / CI.
type fileTokenManager struct {
	path string
	key  []byte // AES-256 key derived from master key
}

// newFileTokenManager creates a file-backed token store.
func newFileTokenManager(dataDir string, masterKey []byte) *fileTokenManager {
	h := sha256.New()
	h.Write(masterKey)
	h.Write([]byte("synaptic-account-token-encryption-v1"))
	aesKey := h.Sum(nil)[:32] // AES-256
	return &fileTokenManager{
		path: filepath.Join(dataDir, "account-tokens.json.enc"),
		key:  aesKey,
	}
}

func (f *fileTokenManager) Get(key string) (string, error) {
	data, err := os.ReadFile(f.path) //nolint:gosec // trusted path
	if err != nil {
		return "", fmt.Errorf("account: read tokens: %w", err)
	}
	plain, err := decrypt(data, f.key)
	if err != nil {
		return "", fmt.Errorf("account: decrypt tokens: %w", err)
	}
	var tokens map[string]string
	if err := json.Unmarshal(plain, &tokens); err != nil {
		return "", fmt.Errorf("account: parse tokens: %w", err)
	}
	val, ok := tokens[key]
	if !ok {
		return "", fmt.Errorf("account: token %q not found", key)
	}
	return val, nil
}

func (f *fileTokenManager) Set(key, value string) error {
	tokens := make(map[string]string)
	// Load existing tokens.
	if data, err := os.ReadFile(f.path); err == nil { //nolint:gosec
		if plain, decErr := decrypt(data, f.key); decErr == nil {
			_ = json.Unmarshal(plain, &tokens)
		}
	}
	tokens[key] = value
	plain, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("account: marshal tokens: %w", err)
	}
	encrypted, err := encrypt(plain, f.key)
	if err != nil {
		return fmt.Errorf("account: encrypt tokens: %w", err)
	}
	_ = os.MkdirAll(filepath.Dir(f.path), 0o700)
	if err := os.WriteFile(f.path, encrypted, 0o600); err != nil {
		return fmt.Errorf("account: write tokens: %w", err)
	}
	return nil
}

func (f *fileTokenManager) Delete(key string) error {
	data, err := os.ReadFile(f.path) //nolint:gosec
	if err != nil {
		return nil //nolint:nilerr // file doesn't exist, nothing to delete
	}
	plain, err := decrypt(data, f.key)
	if err != nil {
		return nil //nolint:nilerr // can't decrypt, nothing meaningful to delete
	}
	var tokens map[string]string
	if err := json.Unmarshal(plain, &tokens); err != nil {
		return nil //nolint:nilerr // corrupted file, nothing to delete
	}
	delete(tokens, key)
	plain, _ = json.Marshal(tokens)
	encrypted, _ := encrypt(plain, f.key)
	_ = os.WriteFile(f.path, encrypted, 0o600)
	return nil
}

// keychainAdapter adapts the secrets.Manager-style interface to
// the TokenManager interface.
type keychainAdapter struct {
	getFn    func(key string) (string, error)
	setFn    func(key, value string) error
	deleteFn func(key string) error
}

func (k *keychainAdapter) Get(key string) (string, error) { return k.getFn(key) }
func (k *keychainAdapter) Set(key, value string) error    { return k.setFn(key, value) }
func (k *keychainAdapter) Delete(key string) error        { return k.deleteFn(key) }

// NewKeychainTokenManager wraps a secrets.Manager-like interface.
func NewKeychainTokenManager(getFn func(string) (string, error), setFn func(string, string) error, deleteFn func(string) error) TokenManager {
	return &keychainAdapter{getFn: getFn, setFn: setFn, deleteFn: deleteFn}
}

// NewFallbackTokenManager returns a file-backed token manager for
// environments without keychain access.
func NewFallbackTokenManager(dataDir string, masterKey []byte) TokenManager {
	return newFileTokenManager(dataDir, masterKey)
}

// encrypt a plaintext with AES-256-GCM. The nonce is prepended to
// the ciphertext (12 bytes nonce + tag + data).
func encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	// nonce + ciphertext (which includes the tag)
	return aesgcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("account: ciphertext too short")
	}
	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aesgcm.Open(nil, nonce, ct, nil)
}

var _ = hex.EncodeToString
