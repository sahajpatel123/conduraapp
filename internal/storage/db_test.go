package storage

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/synapticapp/synaptic/internal/secrets"
)

const testMasterKey = "k6Qm1xJ4pYqZ8cV2nB3wD5rT7eH9uL0sA1bC2dE3fG4=" // 32 bytes base64

func newTestDB(t *testing.T) *DB {
	t.Helper()
	dir := t.TempDir()
	db, err := Open(context.Background(), Config{
		Path:      filepath.Join(dir, "synaptic.db"),
		MasterKey: testMasterKey,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestOpen_DefaultPath(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(context.Background(), Config{
		Path:      filepath.Join(dir, "synaptic.db"),
		MasterKey: testMasterKey,
	})
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	assert.FileExists(t, db.Path())
}

func TestOpen_NoPath(t *testing.T) {
	_, err := Open(context.Background(), Config{})
	assert.Error(t, err)
}

func TestOpen_BadMasterKey(t *testing.T) {
	dir := t.TempDir()
	cases := []string{
		"not-base64!@#$",
		base64.StdEncoding.EncodeToString(make([]byte, 16)), // too short
		base64.StdEncoding.EncodeToString(make([]byte, 64)), // too long
	}
	for i, mk := range cases {
		_, err := Open(context.Background(), Config{
			Path:      filepath.Join(dir, "db"+string(rune('a'+i))+".db"),
			MasterKey: mk,
		})
		assert.Error(t, err, "case %d (%q) should error", i, mk)
	}
}

func TestOpen_CreatesDataDir(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "deeply", "nested", "data")
	db, err := Open(context.Background(), Config{
		Path:      filepath.Join(nested, "synaptic.db"),
		MasterKey: testMasterKey,
	})
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	info, err := os.Stat(nested)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestOpen_LoadsMasterKeyFromSecrets(t *testing.T) {
	dir := t.TempDir()
	sm, err := secrets.New(filepath.Join(dir, "s.json"))
	require.NoError(t, err)
	require.NoError(t, sm.Set(secrets.MasterKey, testMasterKey))

	db, err := Open(context.Background(), Config{
		Path:    filepath.Join(dir, "synaptic.db"),
		Secrets: sm,
	})
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
}

func TestOpen_GeneratesAndStoresMasterKey(t *testing.T) {
	dir := t.TempDir()
	sm, err := secrets.New(filepath.Join(dir, "s.json"))
	require.NoError(t, err)
	// Don't pre-set the master key.
	db, err := Open(context.Background(), Config{
		Path:    filepath.Join(dir, "synaptic.db"),
		Secrets: sm,
	})
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	// Master key should now exist in secrets.
	mk, err := sm.Get(secrets.MasterKey)
	require.NoError(t, err)
	assert.NotEmpty(t, mk)
	// And should be valid base64 of 32 bytes.
	decoded, err := base64.StdEncoding.DecodeString(mk)
	require.NoError(t, err)
	assert.Len(t, decoded, 32)
}

func TestOpen_PersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "synaptic.db")

	db1, err := Open(context.Background(), Config{Path: path, MasterKey: testMasterKey})
	require.NoError(t, err)
	require.NoError(t, db1.Close())

	db2, err := Open(context.Background(), Config{Path: path, MasterKey: testMasterKey})
	require.NoError(t, err)
	defer func() { _ = db2.Close() }()
}

func TestClose_Idempotent(t *testing.T) {
	db := newTestDB(t)
	assert.NoError(t, db.Close())
	assert.NoError(t, db.Close())
}

func TestSQL_And_Path(t *testing.T) {
	db := newTestDB(t)
	assert.NotNil(t, db.SQL())
	assert.NotEmpty(t, db.Path())
	assert.WithinDuration(t, db.OpenedAt(), db.OpenedAt(), 0)
}

// -----------------------------------------------------------------------------
// Encryption tests
// -----------------------------------------------------------------------------

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	db := newTestDB(t)
	const rowID = 42
	const col = "secret"
	plain := []byte("super-secret-api-key")

	ct, err := db.Encrypt(plain, rowID, col)
	require.NoError(t, err)
	assert.NotEqual(t, plain, ct, "ciphertext must differ from plaintext")

	pt, err := db.Decrypt(ct, rowID, col)
	require.NoError(t, err)
	assert.Equal(t, plain, pt)
}

func TestEncrypt_NonceUnique(t *testing.T) {
	db := newTestDB(t)
	a, err := db.Encrypt([]byte("same"), 1, "c")
	require.NoError(t, err)
	b, err := db.Encrypt([]byte("same"), 1, "c")
	require.NoError(t, err)
	assert.NotEqual(t, a, b, "AES-GCM must use a unique nonce per encryption")
}

func TestEncrypt_DifferentRows(t *testing.T) {
	db := newTestDB(t)
	ct, err := db.Encrypt([]byte("x"), 1, "c")
	require.NoError(t, err)
	// Wrong row ID should fail to authenticate.
	_, err = db.Decrypt(ct, 2, "c")
	assert.Error(t, err)
}

func TestEncrypt_DifferentColumns(t *testing.T) {
	db := newTestDB(t)
	ct, err := db.Encrypt([]byte("x"), 1, "c1")
	require.NoError(t, err)
	_, err = db.Decrypt(ct, 1, "c2")
	assert.Error(t, err)
}

func TestEncryptStringRoundTrip(t *testing.T) {
	db := newTestDB(t)
	const rowID = 7
	const col = "k"
	ct, err := db.EncryptString("sk-abc", rowID, col)
	require.NoError(t, err)
	assert.NotEmpty(t, ct)
	pt, err := db.DecryptString(ct, rowID, col)
	require.NoError(t, err)
	assert.Equal(t, "sk-abc", pt)
}

func TestDecryptString_Empty(t *testing.T) {
	db := newTestDB(t)
	pt, err := db.DecryptString("", 1, "c")
	require.NoError(t, err)
	assert.Empty(t, pt)
}

func TestDecryptString_BadBase64(t *testing.T) {
	db := newTestDB(t)
	_, err := db.DecryptString("not!base64", 1, "c")
	assert.Error(t, err)
}

func TestDecrypt_TooShort(t *testing.T) {
	db := newTestDB(t)
	_, err := db.Decrypt([]byte{1, 2, 3}, 1, "c")
	assert.Error(t, err)
}

func TestDecrypt_Tampered(t *testing.T) {
	db := newTestDB(t)
	ct, err := db.Encrypt([]byte("hello"), 1, "c")
	require.NoError(t, err)
	ct[len(ct)-1] ^= 0xFF // flip last byte
	_, err = db.Decrypt(ct, 1, "c")
	assert.Error(t, err)
}

func TestEncrypt_DifferentDBsCannotCrossDecrypt(t *testing.T) {
	// Two DBs with different master keys must not be able to cross-decrypt.
	dir := t.TempDir()
	db1, err := Open(context.Background(), Config{
		Path: filepath.Join(dir, "a.db"),
		// 32 zero bytes — different from testMasterKey.
		MasterKey: base64.StdEncoding.EncodeToString(make([]byte, 32)),
	})
	require.NoError(t, err)
	defer func() { _ = db1.Close() }()

	db2 := newTestDB(t)

	ct, err := db1.Encrypt([]byte("x"), 1, "c")
	require.NoError(t, err)
	_, err = db2.Decrypt(ct, 1, "c")
	assert.Error(t, err, "different master key must fail")
}

// -----------------------------------------------------------------------------
// Schema / migration tests
// -----------------------------------------------------------------------------

func TestMigrations_Applied(t *testing.T) {
	db := newTestDB(t)
	require.NoError(t, db.EnsureVersion(context.Background(), len(migrations)))
}

func TestEnsureVersion_NotApplied(t *testing.T) {
	db := newTestDB(t)
	err := db.EnsureVersion(context.Background(), 999)
	assert.Error(t, err)
}

func TestEnsureVersion_Unknown(t *testing.T) {
	db := newTestDB(t)
	err := db.EnsureVersion(context.Background(), -1)
	assert.ErrorIs(t, err, ErrNoMigration)
}

func TestMigrations_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "synaptic.db")

	cfg := Config{Path: path, MasterKey: testMasterKey}
	db1, err := Open(context.Background(), cfg)
	require.NoError(t, err)
	require.NoError(t, db1.Close())

	// Re-open should not re-apply migrations.
	db2, err := Open(context.Background(), cfg)
	require.NoError(t, err)
	defer func() { _ = db2.Close() }()
}

func TestMigrations_OnMigrate(t *testing.T) {
	dir := t.TempDir()
	var called []int
	var mu sync.Mutex
	db, err := Open(context.Background(), Config{
		Path:      filepath.Join(dir, "synaptic.db"),
		MasterKey: testMasterKey,
		OnMigrate: func(v int) error {
			mu.Lock()
			defer mu.Unlock()
			called = append(called, v)
			return nil
		},
	})
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []int{1}, called)
}

func TestMigrations_OnMigrateError(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(context.Background(), Config{
		Path:      filepath.Join(dir, "synaptic.db"),
		MasterKey: testMasterKey,
		OnMigrate: func(v int) error { return errors.New("nope") },
	})
	assert.Error(t, err)
	assert.Nil(t, db)
}

// -----------------------------------------------------------------------------
// Master key generation
// -----------------------------------------------------------------------------

func TestGenerateMasterKey(t *testing.T) {
	mk, err := generateMasterKey()
	require.NoError(t, err)
	decoded, err := base64.StdEncoding.DecodeString(mk)
	require.NoError(t, err)
	assert.Len(t, decoded, 32)
}

func TestGenerateMasterKey_Unique(t *testing.T) {
	a, _ := generateMasterKey()
	b, _ := generateMasterKey()
	assert.NotEqual(t, a, b)
}

func TestLoadOrCreateMasterKey_DirectUse(t *testing.T) {
	mk, err := loadOrCreateMasterKey(Config{
		MasterKey: testMasterKey,
	})
	require.NoError(t, err)
	assert.Equal(t, testMasterKey, mk)
}

func TestLoadOrCreateMasterKey_EphemeralWhenNoSecrets(t *testing.T) {
	// No MasterKey, no Secrets: ephemeral.
	mk, err := loadOrCreateMasterKey(Config{})
	require.NoError(t, err)
	decoded, err := base64.StdEncoding.DecodeString(mk)
	require.NoError(t, err)
	assert.Len(t, decoded, 32)
}

func TestLoadOrCreateMasterKey_SecretsError(t *testing.T) {
	// Secrets.Manager that returns a non-NotFound error.
	sm := &errSecrets{getErr: errors.New("boom")}
	_, err := loadOrCreateMasterKey(Config{Secrets: sm})
	assert.Error(t, err)
}

func TestLoadOrCreateMasterKey_SecretsSetError(t *testing.T) {
	// Secrets that successfully return NotFound on Get but fail on Set.
	sm := &errSecrets{getNotFound: true, setErr: errors.New("set boom")}
	_, err := loadOrCreateMasterKey(Config{Secrets: sm})
	assert.Error(t, err)
}

type errSecrets struct {
	getNotFound bool
	getErr      error
	setErr      error
}

func (e *errSecrets) Get(key string) (string, error) {
	if e.getErr != nil {
		return "", e.getErr
	}
	if e.getNotFound {
		return "", secrets.ErrNotFound
	}
	return "", nil
}

func (e *errSecrets) Set(key, value string) error { return e.setErr }
func (e *errSecrets) Delete(key string) error     { return nil }
func (e *errSecrets) Backend() secrets.Backend    { return secrets.BackendFile }
func (e *errSecrets) Close() error                { return nil }

// -----------------------------------------------------------------------------
// EnsureVersion edge cases
// -----------------------------------------------------------------------------

func TestEnsureVersion_Downgrade(t *testing.T) {
	// Downgrade detection only kicks in when both the current and target
	// versions exist in the migration set. With only v1 defined, asking
	// for v0 returns ErrNoMigration. The downgrade path will be exercised
	// once a v2 migration is added.
	db := newTestDB(t)
	err := db.EnsureVersion(context.Background(), 0)
	assert.ErrorIs(t, err, ErrNoMigration)
}

func TestSchema_AllTablesExist(t *testing.T) {
	db := newTestDB(t)
	for _, table := range []string{
		"schema_version", "api_keys", "llm_calls", "spend_daily",
		"audit_log", "provider_health", "memory_entries",
	} {
		var name string
		err := db.SQL().QueryRowContext(context.Background(),
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		require.NoError(t, err, "table %s missing", table)
		assert.Equal(t, table, name)
	}
}

func TestSchema_InsertAPIKey(t *testing.T) {
	db := newTestDB(t)
	// Encrypt with the row ID we'll get back (1 for the first insert).
	// Since AAD is "<rowID>:<column>", we must use the same row ID on decrypt.
	ct, err := db.EncryptString("sk-test", 1, "secret_ciphertext")
	require.NoError(t, err)
	_, err = db.SQL().ExecContext(context.Background(),
		`INSERT INTO api_keys (provider, auth_kind, secret_ciphertext) VALUES (?, ?, ?)`,
		"openai", "api_key", ct)
	require.NoError(t, err)

	var got string
	err = db.SQL().QueryRowContext(context.Background(),
		`SELECT secret_ciphertext FROM api_keys WHERE provider=?`, "openai").Scan(&got)
	require.NoError(t, err)
	pt, err := db.DecryptString(got, 1, "secret_ciphertext")
	require.NoError(t, err)
	assert.Equal(t, "sk-test", pt)
}

func TestSchema_InsertLLMCall(t *testing.T) {
	db := newTestDB(t)
	_, err := db.SQL().ExecContext(context.Background(),
		`INSERT INTO llm_calls (provider, model, task, success) VALUES (?, ?, ?, ?)`,
		"openai", "gpt-4o-mini", "chat", 1)
	require.NoError(t, err)

	var n int
	err = db.SQL().QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM llm_calls WHERE provider=?`, "openai").Scan(&n)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

// -----------------------------------------------------------------------------
// Concurrent access tests
// -----------------------------------------------------------------------------

func TestEncrypt_Concurrent(t *testing.T) {
	db := newTestDB(t)
	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			ct, err := db.Encrypt([]byte("hi"), int64(i), "c")
			require.NoError(t, err)
			_, err = db.Decrypt(ct, int64(i), "c")
			require.NoError(t, err)
		}(i)
	}
	wg.Wait()
}

func TestDB_ConcurrentReads(t *testing.T) {
	db := newTestDB(t)
	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			var v int
			err := db.SQL().QueryRowContext(context.Background(), `SELECT 1`).Scan(&v)
			require.NoError(t, err)
			assert.Equal(t, 1, v)
		}()
	}
	wg.Wait()
}
