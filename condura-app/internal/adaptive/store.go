package adaptive

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"
)

// EncryptedStore implements Store per hard-invariant #1: every byte of
// the user model is encrypted at rest via storage.DB encryption.
type EncryptedStore struct {
	db      *sql.DB
	encrypt EncryptFunc
	decrypt DecryptFunc
	mu      sync.Mutex
}

// EncryptFunc matches storage.DB.EncryptString signature.
type EncryptFunc func(plaintext string, rowID int64, column string) (string, error)

// DecryptFunc matches storage.DB.DecryptString signature.
type DecryptFunc func(ciphertext string, rowID int64, column string) (string, error)

// NewEncryptedStore creates a user model store backed by encrypted storage.
func NewEncryptedStore(db *sql.DB, encrypt EncryptFunc, decrypt DecryptFunc) (*EncryptedStore, error) {
	s := &EncryptedStore{db: db, encrypt: encrypt, decrypt: decrypt}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

const modelRowID = 1
const modelColumn = "user_model_data"

func (s *EncryptedStore) migrate() error {
	ctx := context.Background()
	_, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS adaptive_user_model (
		id INTEGER PRIMARY KEY DEFAULT 1,
		user_model_data TEXT NOT NULL DEFAULT '',
		meta_data TEXT NOT NULL DEFAULT '{}',
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("adaptive: migrate: %w", err)
	}
	_, _ = s.db.ExecContext(ctx, `INSERT OR IGNORE INTO adaptive_user_model (id) VALUES (1)`)
	return nil
}

// Load retrieves and decrypts the stored user model.
func (s *EncryptedStore) Load() (*UserModel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()
	var encrypted string
	err := s.db.QueryRowContext(ctx, `SELECT user_model_data FROM adaptive_user_model WHERE id = ?`, modelRowID).Scan(&encrypted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &UserModel{LastUpdated: time.Now(), Version: 1}, nil
		}
		return nil, fmt.Errorf("adaptive: load: %w", err)
	}
	if encrypted == "" {
		return &UserModel{LastUpdated: time.Now(), Version: 1}, nil
	}
	plain, err := s.decrypt(encrypted, modelRowID, modelColumn)
	if err != nil {
		return nil, fmt.Errorf("adaptive: decrypt: %w", err)
	}
	return UnmarshalUserModel([]byte(plain))
}

// Save encrypts and persists the user model.
func (s *EncryptedStore) Save(model *UserModel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()
	model.LastUpdated = time.Now()
	data, err := model.Marshal()
	if err != nil {
		return fmt.Errorf("adaptive: marshal: %w", err)
	}
	encrypted, err := s.encrypt(string(data), modelRowID, modelColumn)
	if err != nil {
		return fmt.Errorf("adaptive: encrypt: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `UPDATE adaptive_user_model SET user_model_data=?, meta_data=?, updated_at=? WHERE id=?`,
		encrypted, "{}", model.LastUpdated, modelRowID)
	return err
}

// Reset clears all user model data.
func (s *EncryptedStore) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := context.Background()
	_, err := s.db.ExecContext(ctx, `UPDATE adaptive_user_model SET user_model_data='', meta_data='{}', updated_at=? WHERE id=?`,
		time.Now(), modelRowID)
	return err
}

// Close releases resources held by the store.
func (s *EncryptedStore) Close() error { return nil }
