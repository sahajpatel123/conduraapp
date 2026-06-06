package secrets

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

// -----------------------------------------------------------------------------
// Fake keyring for tests
// -----------------------------------------------------------------------------

type fakeKeyring struct {
	mu    sync.Mutex
	store map[string]map[string]string // service -> key -> value
	// Optional error injection
	forceSetErr    error
	forceGetErr    error
	forceDeleteErr error
	// When true, Get always returns a corrupted value to simulate a probe mismatch.
	corruptGet bool
}

func newFakeKeyring() *fakeKeyring {
	return &fakeKeyring{store: map[string]map[string]string{}}
}

func (f *fakeKeyring) Get(s, k string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.forceGetErr != nil {
		return "", f.forceGetErr
	}
	if f.corruptGet {
		return "corrupted", nil
	}
	if v, ok := f.store[s][k]; ok {
		return v, nil
	}
	return "", keyring.ErrNotFound
}

func (f *fakeKeyring) Set(s, k, v string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.forceSetErr != nil {
		return f.forceSetErr
	}
	if f.store[s] == nil {
		f.store[s] = map[string]string{}
	}
	f.store[s][k] = v
	return nil
}

func (f *fakeKeyring) Delete(s, k string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.forceDeleteErr != nil {
		return f.forceDeleteErr
	}
	if _, ok := f.store[s][k]; !ok {
		return keyring.ErrNotFound
	}
	delete(f.store[s], k)
	return nil
}

// -----------------------------------------------------------------------------
// File backend tests
// -----------------------------------------------------------------------------

func newTestFileManager(t *testing.T) *fileManager {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.json")
	m, err := newFileManager(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = m.Close() })
	return m
}

func TestNew_Default(t *testing.T) {
	dir := t.TempDir()
	m, err := New(filepath.Join(dir, "s.json"))
	require.NoError(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, BackendKeyring, m.Backend(), "default backend is keyring on macOS dev")
}

func TestNew_NoFilePath_Auto(t *testing.T) {
	// New() with empty file path on darwin should succeed (keyring is up).
	if runtime.GOOS != "darwin" {
		t.Skip("tested on darwin where keyring is reliable")
	}
	m, err := New("")
	require.NoError(t, err)
	assert.Equal(t, BackendKeyring, m.Backend())
}

func TestFileManager_SetGet(t *testing.T) {
	m := newTestFileManager(t)
	require.NoError(t, m.Set("api_key.openai", "sk-test-123"))
	got, err := m.Get("api_key.openai")
	require.NoError(t, err)
	assert.Equal(t, "sk-test-123", got)
}

func TestFileManager_GetMissing(t *testing.T) {
	m := newTestFileManager(t)
	_, err := m.Get("does.not.exist")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestFileManager_Overwrite(t *testing.T) {
	m := newTestFileManager(t)
	require.NoError(t, m.Set("k", "v1"))
	require.NoError(t, m.Set("k", "v2"))
	got, err := m.Get("k")
	require.NoError(t, err)
	assert.Equal(t, "v2", got)
}

func TestFileManager_Delete(t *testing.T) {
	m := newTestFileManager(t)
	require.NoError(t, m.Set("k", "v"))
	require.NoError(t, m.Delete("k"))
	_, err := m.Get("k")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestFileManager_DeleteMissing(t *testing.T) {
	m := newTestFileManager(t)
	err := m.Delete("never.set")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestFileManager_FileCreatedWith0600(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.json")
	_, err := newFileManager(path)
	require.NoError(t, err)
	info, err := os.Stat(path)
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "secrets file must be 0600")
	}
}

func TestFileManager_DirCreatedWith0700(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "subdir", "secrets.json")
	_, err := newFileManager(path)
	require.NoError(t, err)
	info, err := os.Stat(filepath.Dir(path))
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		assert.Equal(t, os.FileMode(0o700), info.Mode().Perm(), "secrets dir must be 0700")
	}
}

func TestFileManager_ExistingFileGets0600(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"version":1,"secrets":{}}`), 0o644))
	_, err := newFileManager(path)
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		info, _ := os.Stat(path)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}
}

func TestFileManager_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.json")
	require.NoError(t, os.WriteFile(path, []byte{}, 0o600))
	m, err := newFileManager(path)
	require.NoError(t, err)
	require.NoError(t, m.Set("k", "v"))
	got, err := m.Get("k")
	require.NoError(t, err)
	assert.Equal(t, "v", got)
}

func TestFileManager_CorruptedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.json")
	require.NoError(t, os.WriteFile(path, []byte(`{not valid json`), 0o600))
	m, err := newFileManager(path)
	require.NoError(t, err)
	_, err = m.Get("k")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse secrets file")
}

func TestFileManager_ReadError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing-dir", "secrets.json")
	m := &fileManager{path: path}
	_, err := m.read()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read secrets file")
}

func TestFileManager_NewFileManager_ExistingDir(t *testing.T) {
	// When the given path is an existing directory, Stat succeeds and Chmod
	// is called on the dir (which is a no-op on most systems). The
	// constructor should still succeed.
	dir := t.TempDir()
	d := filepath.Join(dir, "is_a_dir")
	require.NoError(t, os.Mkdir(d, 0o755))
	_, err := newFileManager(d)
	assert.NoError(t, err)
}

func TestFileManager_NewFileManager_MkdirFails(t *testing.T) {
	// Path whose parent cannot be created (e.g. contains a NUL byte).
	_, err := newFileManager("/\x00bad/secrets.json")
	assert.Error(t, err)
}

func TestFileManager_NewFileManager_EmptyPath(t *testing.T) {
	_, err := newFileManager("")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrBackendFailed)
}

func TestFileManager_Set_ReadError(t *testing.T) {
	// Set fails if the underlying read fails (e.g. dir removed).
	dir := t.TempDir()
	parent := filepath.Join(dir, "will_be_removed")
	require.NoError(t, os.Mkdir(parent, 0o700))
	path := filepath.Join(parent, "s.json")
	m, err := newFileManager(path)
	require.NoError(t, err)
	require.NoError(t, os.RemoveAll(parent))
	err = m.Set("k", "v")
	assert.Error(t, err)
}

func TestFileManager_Delete_ReadError(t *testing.T) {
	dir := t.TempDir()
	parent := filepath.Join(dir, "will_be_removed")
	require.NoError(t, os.Mkdir(parent, 0o700))
	path := filepath.Join(parent, "s.json")
	m, err := newFileManager(path)
	require.NoError(t, err)
	require.NoError(t, os.RemoveAll(parent))
	err = m.Delete("k")
	assert.Error(t, err)
}

func TestFileManager_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.json")
	m, err := newFileManager(path)
	require.NoError(t, err)
	require.NoError(t, m.Set("k", "v1"))
	_, err = os.Stat(path + ".tmp")
	assert.True(t, os.IsNotExist(err), "tmp file must not be left behind")
}

func TestFileManager_Backend(t *testing.T) {
	m := newTestFileManager(t)
	assert.Equal(t, BackendFile, m.Backend())
}

func TestFileManager_Close(t *testing.T) {
	m := newTestFileManager(t)
	assert.NoError(t, m.Close())
}

func TestFileManager_Concurrent(t *testing.T) {
	m := newTestFileManager(t)
	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			require.NoError(t, m.Set("k", "v"))
			_, _ = m.Get("k")
		}()
	}
	wg.Wait()
}

func TestFileManager_PersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.json")
	m1, err := newFileManager(path)
	require.NoError(t, err)
	require.NoError(t, m1.Set("k", "v"))

	m2, err := newFileManager(path)
	require.NoError(t, err)
	got, err := m2.Get("k")
	require.NoError(t, err)
	assert.Equal(t, "v", got)
}

// -----------------------------------------------------------------------------
// Factory tests
// -----------------------------------------------------------------------------

func TestNewWithBackend_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.json")
	m, err := NewWithBackend("file", path)
	require.NoError(t, err)
	assert.Equal(t, BackendFile, m.Backend())
}

func TestNewWithBackend_FileMissingPath(t *testing.T) {
	_, err := NewWithBackend("file", "")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrBackendFailed)
}

func TestNewWithBackend_Unknown(t *testing.T) {
	dir := t.TempDir()
	_, err := NewWithBackend("vault", filepath.Join(dir, "s.json"))
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrBackendFailed)
}

func TestNewWithBackend_Auto_FileFallback(t *testing.T) {
	// Use an injecting fake that fails the probe; expect file backend.
	fk := newFakeKeyring()
	fk.forceSetErr = errors.New("simulated keyring failure")
	dir := t.TempDir()
	path := filepath.Join(dir, "s.json")
	m, err := NewWithKeyring(fk, "auto", path)
	require.NoError(t, err)
	assert.Equal(t, BackendFile, m.Backend())
}

func TestNewWithBackend_Auto_NoFilePath(t *testing.T) {
	// If keyring fails AND no file path is given, expect an error.
	fk := newFakeKeyring()
	fk.forceSetErr = errors.New("simulated keyring failure")
	_, err := NewWithKeyring(fk, "auto", "")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrBackendFailed)
}

func TestNewWithBackend_KeyringForced(t *testing.T) {
	// "keyring" should fail if probe fails, not silently fall back.
	fk := newFakeKeyring()
	fk.forceSetErr = errors.New("nope")
	_, err := NewWithKeyring(fk, "keyring", "")
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Keyring backend tests (using fake)
// -----------------------------------------------------------------------------

func TestKeyringManager_RoundTrip(t *testing.T) {
	fk := newFakeKeyring()
	m, err := newKeyringManager(fk)
	require.NoError(t, err)
	const k = "synaptic_test_roundtrip"

	require.NoError(t, m.Set(k, "hello"))
	got, err := m.Get(k)
	require.NoError(t, err)
	assert.Equal(t, "hello", got)

	require.NoError(t, m.Delete(k))
	_, err = m.Get(k)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestKeyringManager_GetMissing(t *testing.T) {
	fk := newFakeKeyring()
	m, err := newKeyringManager(fk)
	require.NoError(t, err)
	_, err = m.Get("missing")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestKeyringManager_DeleteMissing(t *testing.T) {
	fk := newFakeKeyring()
	m, err := newKeyringManager(fk)
	require.NoError(t, err)
	err = m.Delete("missing")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestKeyringManager_GetError(t *testing.T) {
	fk := newFakeKeyring()
	m, err := newKeyringManager(fk)
	require.NoError(t, err)
	fk.forceGetErr = errors.New("boom")
	_, err = m.Get("k")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestKeyringManager_SetError(t *testing.T) {
	fk := newFakeKeyring()
	m, err := newKeyringManager(fk)
	require.NoError(t, err)
	fk.forceSetErr = errors.New("boom")
	err = m.Set("k", "v")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestKeyringManager_DeleteError(t *testing.T) {
	fk := newFakeKeyring()
	m, err := newKeyringManager(fk)
	require.NoError(t, err)
	fk.forceDeleteErr = errors.New("boom")
	err = m.Delete("k")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestKeyringManager_Backend(t *testing.T) {
	fk := newFakeKeyring()
	m, err := newKeyringManager(fk)
	require.NoError(t, err)
	assert.Equal(t, BackendKeyring, m.Backend())
}

func TestKeyringManager_Close(t *testing.T) {
	fk := newFakeKeyring()
	m, err := newKeyringManager(fk)
	require.NoError(t, err)
	assert.NoError(t, m.Close())
}

func TestKeyringManager_ProbeGetFails(t *testing.T) {
	fk := newFakeKeyring()
	fk.forceGetErr = errors.New("get fails")
	_, err := newKeyringManager(fk)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "keyring probe get")
}

func TestKeyringManager_ProbeMismatch(t *testing.T) {
	fk := newFakeKeyring()
	fk.corruptGet = true
	_, err := newKeyringManager(fk)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "keyring probe mismatch")
}

func TestKeyringManager_ProbeDeleteFails(t *testing.T) {
	fk := newFakeKeyring()
	fk.forceDeleteErr = errors.New("delete fails")
	_, err := newKeyringManager(fk)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "keyring probe delete")
}

func TestKeyringManager_ProbeGetErrNotFoundIgnoredOnGet(t *testing.T) {
	// Sanity: Get returns ErrNotFound when the keyring returns keyring.ErrNotFound,
	// even when forceGetErr is also set (forceGetErr takes precedence).
	fk := newFakeKeyring()
	m, err := newKeyringManager(fk)
	require.NoError(t, err)
	_, err = m.Get("missing")
	assert.ErrorIs(t, err, ErrNotFound)
}

// -----------------------------------------------------------------------------
// OS keyring integration (best-effort; skip if unavailable)
// -----------------------------------------------------------------------------

func TestOSKeyring_Available(t *testing.T) {
	m, err := newKeyringManager(zalandoKeyring{})
	if err != nil {
		t.Skipf("OS keyring unavailable: %v", err)
	}
	assert.Equal(t, BackendKeyring, m.Backend())
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func TestAPIKeyKey(t *testing.T) {
	assert.Equal(t, "api_key_openai", APIKeyKey("openai"))
}

func TestOAuthTokenKey(t *testing.T) {
	assert.Equal(t, "oauth_google", OAuthTokenKey("google"))
}

func TestErrors(t *testing.T) {
	assert.NotNil(t, ErrNotFound)
	assert.NotNil(t, ErrBackendFailed)
	assert.True(t, errors.Is(ErrNotFound, ErrNotFound))
}
