package api_key

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sahajpatel123/synapticapp/internal/secrets"
	"github.com/sahajpatel123/synapticapp/internal/storage"
)

const testMK = "k6Qm1xJ4pYqZ8cV2nB3wD5rT7eH9uL0sA1bC2dE3fG4="

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	dir := t.TempDir()
	sm, err := secrets.New(filepath.Join(dir, "s.json"))
	require.NoError(t, err)
	db, err := storage.Open(context.Background(), storage.Config{
		Path:      filepath.Join(dir, "syn.db"),
		MasterKey: testMK,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return New(db, sm)
}

// -----------------------------------------------------------------------------
// Provider metadata
// -----------------------------------------------------------------------------

func TestAllProviders(t *testing.T) {
	assert.Len(t, AllProviders, 12)
	assert.Contains(t, AllProviders, ProviderAnthropic)
	assert.Contains(t, AllProviders, ProviderGoogle)
}

func TestIsValidProvider(t *testing.T) {
	assert.True(t, IsValidProvider("openai"))
	assert.False(t, IsValidProvider("not_a_provider"))
}

func TestNewID(t *testing.T) {
	id := NewID()
	assert.Len(t, id, 8)
	id2 := NewID()
	assert.NotEqual(t, id, id2, "IDs should be unique")
}

func TestProviderLabel(t *testing.T) {
	l := ProviderLabel("openai")
	assert.Contains(t, l, "openai-")
}

// -----------------------------------------------------------------------------
// Manager CRUD
// -----------------------------------------------------------------------------

func TestSet_Basic(t *testing.T) {
	m := newTestManager(t)
	id, err := m.Set(context.Background(), Key{
		Provider: ProviderOpenAI,
		Label:    "work",
		AuthKind: AuthAPIKey,
		Secret:   "sk-abc123",
	})
	require.NoError(t, err)
	assert.NotZero(t, id)
}

func TestSet_NoProvider(t *testing.T) {
	m := newTestManager(t)
	_, err := m.Set(context.Background(), Key{Secret: "x"})
	assert.ErrorIs(t, err, ErrNoProvider)
}

func TestSet_UnknownProvider(t *testing.T) {
	m := newTestManager(t)
	_, err := m.Set(context.Background(), Key{Provider: "fake", Secret: "x"})
	assert.Error(t, err)
}

func TestSet_EmptySecret(t *testing.T) {
	m := newTestManager(t)
	_, err := m.Set(context.Background(), Key{Provider: ProviderOpenAI, Secret: ""})
	assert.ErrorIs(t, err, ErrInvalidSecret)
}

func TestSet_InvalidAuthKind(t *testing.T) {
	m := newTestManager(t)
	_, err := m.Set(context.Background(), Key{Provider: ProviderOpenAI, AuthKind: "weird", Secret: "x"})
	assert.ErrorIs(t, err, ErrInvalidKind)
}

func TestSet_DefaultLabel(t *testing.T) {
	m := newTestManager(t)
	id, err := m.Set(context.Background(), Key{Provider: ProviderOpenAI, Secret: "x"})
	require.NoError(t, err)
	k, err := m.Get(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, "default", k.Label)
}

func TestSet_ReplacesExisting(t *testing.T) {
	m := newTestManager(t)
	id1, err := m.Set(context.Background(), Key{Provider: ProviderOpenAI, Label: "L", Secret: "v1"})
	require.NoError(t, err)
	id2, err := m.Set(context.Background(), Key{Provider: ProviderOpenAI, Label: "L", Secret: "v2"})
	require.NoError(t, err)
	assert.Equal(t, id1, id2, "same (provider, label) should upsert to same ID")
	k, err := m.Get(context.Background(), id1)
	require.NoError(t, err)
	assert.Equal(t, "v2", k.Secret)
}

func TestSet_OAuth(t *testing.T) {
	m := newTestManager(t)
	id, err := m.Set(context.Background(), Key{
		Provider:  ProviderGoogle,
		Label:     "personal",
		AuthKind:  AuthOAuth,
		Secret:    "ya29.access",
		Refresh:   "1//refresh",
		Scopes:    "openid email",
		ExpiresAt: time.Now().Add(time.Hour),
	})
	require.NoError(t, err)
	k, err := m.Get(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, AuthOAuth, k.AuthKind)
	assert.Equal(t, "ya29.access", k.Secret)
	assert.Equal(t, "1//refresh", k.Refresh)
	assert.Equal(t, "openid email", k.Scopes)
	assert.False(t, k.ExpiresAt.IsZero())
}

func TestGet_NotFound(t *testing.T) {
	m := newTestManager(t)
	_, err := m.Get(context.Background(), 9999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestGetByLabel(t *testing.T) {
	m := newTestManager(t)
	_, err := m.Set(context.Background(), Key{Provider: ProviderAnthropic, Label: "L", Secret: "sk-ant-x"})
	require.NoError(t, err)
	k, err := m.GetByLabel(context.Background(), ProviderAnthropic, "L")
	require.NoError(t, err)
	assert.Equal(t, "sk-ant-x", k.Secret)
}

func TestGetByLabel_NotFound(t *testing.T) {
	m := newTestManager(t)
	_, err := m.GetByLabel(context.Background(), ProviderAnthropic, "nope")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestList(t *testing.T) {
	m := newTestManager(t)
	_, err := m.Set(context.Background(), Key{Provider: ProviderOpenAI, Label: "A", Secret: "x"})
	require.NoError(t, err)
	_, err = m.Set(context.Background(), Key{Provider: ProviderGoogle, Label: "B", Secret: "y"})
	require.NoError(t, err)

	keys, err := m.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, keys, 2)
}

func TestListByProvider(t *testing.T) {
	m := newTestManager(t)
	_, err := m.Set(context.Background(), Key{Provider: ProviderOpenAI, Label: "A", Secret: "x"})
	require.NoError(t, err)
	_, err = m.Set(context.Background(), Key{Provider: ProviderGoogle, Label: "B", Secret: "y"})
	require.NoError(t, err)
	_, err = m.Set(context.Background(), Key{Provider: ProviderOpenAI, Label: "C", Secret: "z"})
	require.NoError(t, err)

	keys, err := m.ListByProvider(context.Background(), ProviderOpenAI)
	require.NoError(t, err)
	assert.Len(t, keys, 2)
}

func TestDelete(t *testing.T) {
	m := newTestManager(t)
	id, err := m.Set(context.Background(), Key{Provider: ProviderOpenAI, Secret: "x"})
	require.NoError(t, err)
	require.NoError(t, m.Delete(context.Background(), id))
	_, err = m.Get(context.Background(), id)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDelete_NotFound(t *testing.T) {
	m := newTestManager(t)
	err := m.Delete(context.Background(), 9999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestTouch(t *testing.T) {
	m := newTestManager(t)
	id, err := m.Set(context.Background(), Key{Provider: ProviderOpenAI, Secret: "x"})
	require.NoError(t, err)
	// Touch shouldn't error.
	require.NoError(t, m.Touch(context.Background(), id))
	k, err := m.Get(context.Background(), id)
	require.NoError(t, err)
	assert.False(t, k.LastUsedAt.IsZero())
}

// -----------------------------------------------------------------------------
// Authenticator interface
// -----------------------------------------------------------------------------

func TestAuthenticator_Interface(t *testing.T) {
	m := newTestManager(t)
	_, err := m.Set(context.Background(), Key{Provider: ProviderOpenAI, Label: "x", Secret: "x"})
	require.NoError(t, err)
	// All three Authenticator methods must work via the interface.
	var a Authenticator = m
	_, err = a.GetByLabel(context.Background(), ProviderOpenAI, "x")
	require.NoError(t, err)
	_, err = a.ListByProvider(context.Background(), ProviderOpenAI)
	require.NoError(t, err)
	require.NoError(t, a.Touch(context.Background(), 1))
}

// -----------------------------------------------------------------------------
// Validate
// -----------------------------------------------------------------------------

func TestValidate_OK(t *testing.T) {
	err := Validate(Key{Provider: ProviderOpenAI, Secret: "sk-x"})
	assert.NoError(t, err)
}

func TestValidate_NoProvider(t *testing.T) {
	assert.ErrorIs(t, Validate(Key{Secret: "x"}), ErrNoProvider)
}

func TestValidate_UnknownProvider(t *testing.T) {
	assert.Error(t, Validate(Key{Provider: "fake", Secret: "x"}))
}

func TestValidate_EmptySecret(t *testing.T) {
	assert.ErrorIs(t, Validate(Key{Provider: ProviderOpenAI}), ErrInvalidSecret)
}

func TestValidate_BadKind(t *testing.T) {
	assert.ErrorIs(t, Validate(Key{Provider: ProviderOpenAI, AuthKind: "x", Secret: "y"}), ErrInvalidKind)
}

func TestValidate_DefaultKind(t *testing.T) {
	// Empty kind is treated as api_key and accepted.
	assert.NoError(t, Validate(Key{Provider: ProviderOpenAI, Secret: "x"}))
}
