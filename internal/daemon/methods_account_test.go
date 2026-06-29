package daemon

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/sahajpatel123/conduraapp/internal/account"
	"github.com/sahajpatel123/conduraapp/internal/ipc"
)

func TestDeriveMagicVerifyURL(t *testing.T) {
	cases := []struct {
		name     string
		issue    string
		explicit string
		want     string
	}{
		{
			name:  "default condura.app magic+verify pattern",
			issue: "https://condura.app/api/auth/magic",
			want:  "https://condura.app/api/auth/verify",
		},
		{
			name:     "explicit override wins",
			issue:    "https://condura.app/api/auth/magic",
			explicit: "https://custom.example/verify",
			want:     "https://custom.example/verify",
		},
		{
			name:  "trailing slash is trimmed before /verify",
			issue: "https://condura.app/api/auth/magic/",
			want:  "https://condura.app/api/auth/verify",
		},
		{
			name:  "non-magic path falls back to issue+/verify",
			issue: "https://example.com/custom/magic",
			want:  "https://example.com/custom/verify",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := deriveMagicVerifyURL(tc.issue, tc.explicit)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDisplayNameFor(t *testing.T) {
	cases := []struct {
		email string
		want  string
	}{
		{"alice@example.com", "alice"},
		{"bob", "bob"},
		{"", ""},
	}
	for _, tc := range cases {
		t.Run(tc.email, func(t *testing.T) {
			got := displayNameFor(&account.Session{Email: tc.email})
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestOAuthURLError_MapsProviderNotConfigured(t *testing.T) {
	e := oauthURLError("google", account.ErrProviderNotConfigured)
	if e.Code != ipc.CodeInvalidParams {
		t.Fatalf("code: got %d, want %d", e.Code, ipc.CodeInvalidParams)
	}
	if !strings.Contains(e.Message, "google") {
		t.Fatalf("message missing provider name: %s", e.Message)
	}
	if !strings.Contains(e.Message, "CONDURA_ACCOUNT_OAUTH_GOOGLE_CLIENT_ID") {
		t.Fatalf("message missing env-var hint: %s", e.Message)
	}
}

func TestOAuthURLError_OtherErrorsAreInternal(t *testing.T) {
	e := oauthURLError("google", errors.New("kaboom"))
	if e.Code != ipc.CodeInternalError {
		t.Fatalf("code: got %d, want %d", e.Code, ipc.CodeInternalError)
	}
	if e.Message != "kaboom" {
		t.Fatalf("message: got %q, want %q", e.Message, "kaboom")
	}
}

// --- end-to-end tests against the live IPC server (Tier 2) ---

func TestAccountStatusRPC_NoProvidersConfigured(t *testing.T) {
	mgr := mustManager(t, nil) // no providers configured
	srv := newTestAccountServer(t, mgr)

	resp, err := accountCallRPC(t, srv, "account.status", nil)
	if err != nil {
		t.Fatalf("account.status: %v", err)
	}
	got, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("response type: got %T", resp)
	}
	if signed, _ := got["signed_in"].(bool); signed {
		t.Fatalf("expected signed_in=false, got %v", got["signed_in"])
	}
	provs, _ := got["providers"].([]string)
	if len(provs) != 0 {
		t.Fatalf("expected empty providers, got %v", provs)
	}
}

func TestAccountStatusRPC_ProvidersConfigured(t *testing.T) {
	mgr := mustManager(t, map[string]account.ProviderConfig{
		"google": {ClientID: "test-cid"},
	})
	srv := newTestAccountServer(t, mgr)

	resp, err := accountCallRPC(t, srv, "account.status", nil)
	if err != nil {
		t.Fatalf("account.status: %v", err)
	}
	got := resp.(map[string]any)
	provs, _ := got["providers"].([]string)
	if len(provs) != 1 || provs[0] != "google" {
		t.Fatalf("providers: got %v, want [google]", provs)
	}
}

func TestAccountOAuthURL_UnconfiguredProvider(t *testing.T) {
	mgr := mustManager(t, map[string]account.ProviderConfig{
		"google": {ClientID: "test-cid"}, // github intentionally missing
	})
	srv := newTestAccountServer(t, mgr)

	_, err := accountCallRPC(t, srv, "account.oauth_url",
		json.RawMessage(`{"provider":"github"}`))
	if err == nil {
		t.Fatal("expected error for unconfigured provider")
	}
	ipcErr := &ipc.Error{}
	if !errors.As(err, &ipcErr) {
		t.Fatalf("expected *ipc.Error, got %T (%v)", err, err)
	}
	if ipcErr.Code != ipc.CodeInvalidParams {
		t.Fatalf("code: got %d, want %d", ipcErr.Code, ipc.CodeInvalidParams)
	}
	if !strings.Contains(ipcErr.Message, "github") {
		t.Fatalf("message missing provider name: %s", ipcErr.Message)
	}
}

func TestAccountOAuthURL_ConfiguredProvider(t *testing.T) {
	mgr := mustManager(t, map[string]account.ProviderConfig{
		"google": {ClientID: "test-cid"},
	})
	srv := newTestAccountServer(t, mgr)

	resp, err := accountCallRPC(t, srv, "account.oauth_url",
		json.RawMessage(`{"provider":"google"}`))
	if err != nil {
		t.Fatalf("account.oauth_url: %v", err)
	}
	m, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("response type: got %T", resp)
	}
	url, _ := m["url"].(string)
	state, _ := m["state"].(string)
	if !strings.Contains(url, "accounts.google.com") {
		t.Fatalf("url: got %q (missing google domain)", url)
	}
	if !strings.Contains(url, "client_id=test-cid") {
		t.Fatalf("url missing client_id: %q", url)
	}
	if !strings.Contains(url, "code_challenge_method=S256") {
		t.Fatalf("url missing PKCE: %q", url)
	}
	if state == "" {
		t.Fatal("state is empty")
	}
	// redirect_uri is URL-encoded in the query string.
	encoded := strings.ReplaceAll(strings.ReplaceAll(OAuthRedirectURI, ":", "%3A"), "/", "%2F")
	if !strings.Contains(url, "redirect_uri="+encoded) {
		t.Fatalf("url missing redirect_uri (encoded=%q): %q", encoded, url)
	}
	// Round-trip: state should be a 32-char hex string (16 random bytes).
	if len(state) != 32 {
		t.Fatalf("state length: got %d, want 32 (16 bytes hex)", len(state))
	}
}

func TestAccountProvidersRPC(t *testing.T) {
	mgr := mustManager(t, map[string]account.ProviderConfig{
		"google": {ClientID: "g"},
		"github": {}, // unconfigured
	})
	srv := newTestAccountServer(t, mgr)

	resp, err := accountCallRPC(t, srv, "account.providers", nil)
	if err != nil {
		t.Fatalf("account.providers: %v", err)
	}
	m := resp.(map[string]any)
	provs, _ := m["providers"].([]string)
	if len(provs) != 1 || provs[0] != "google" {
		t.Fatalf("providers: got %v, want [google]", provs)
	}
}

func TestAccountMagicLinkRPC(t *testing.T) {
	mgr := mustManager(t, nil)
	srv := newTestAccountServer(t, mgr)
	// Note: the actual HTTP call will fail because the magic-link URL
	// is unreachable in unit tests. We just want to make sure the
	// dispatcher accepts the request and returns a properly-shaped
	// error (not a panic or a 500).
	_, err := accountCallRPC(t, srv, "account.magic_link",
		json.RawMessage(`{"email":"not-an-email"}`))
	if err == nil {
		t.Fatal("expected validation error for invalid email")
	}
	ipcErr := &ipc.Error{}
	if !errors.As(err, &ipcErr) {
		t.Fatalf("expected *ipc.Error, got %T (%v)", err, err)
	}
	if ipcErr.Code != ipc.CodeInternalError {
		t.Fatalf("code: got %d, want %d", ipcErr.Code, ipc.CodeInternalError)
	}
}

func TestAccountLogoutRPC(t *testing.T) {
	mgr := mustManager(t, map[string]account.ProviderConfig{
		"google": {ClientID: "g"},
	})
	// Seed a session so logout has something to clear.
	_, _ = mgr.NewSession(context.Background(), "x@example.com", "google")

	srv := newTestAccountServer(t, mgr)
	resp, err := accountCallRPC(t, srv, "account.logout", nil)
	if err != nil {
		t.Fatalf("account.logout: %v", err)
	}
	m, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("response type: got %T", resp)
	}
	if ok, _ := m["ok"].(bool); !ok {
		t.Fatalf("ok: got %v, want true", m["ok"])
	}
	// Subsequent status should be signed out.
	resp, err = accountCallRPC(t, srv, "account.status", nil)
	if err != nil {
		t.Fatalf("account.status post-logout: %v", err)
	}
	got := resp.(map[string]any)
	if signed, _ := got["signed_in"].(bool); signed {
		t.Fatal("expected signed_in=false after logout")
	}
}

func TestAccountNotAvailableRPC_NoPanic(t *testing.T) {
	srv := newTestAccountServer(t, nil)
	for _, method := range []string{
		"account.status",
		"account.providers",
		"account.oauth_url",
		"account.oauth_callback",
		"account.magic_link",
		"account.logout",
	} {
		if _, err := accountCallRPC(t, srv, method, nil); err == nil {
			t.Fatalf("%s should return error when account is nil", method)
		}
	}
}

// --- helpers ---

func accountCallRPC(t *testing.T, srv *ipc.Server, method string, params json.RawMessage) (any, error) {
	t.Helper()
	resp, err := srv.Handle(context.Background(), &ipc.Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      json.RawMessage("1"),
	})
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Result, nil
}

// newTestAccountServer builds a server with an Account wired in. Use
// this when the test needs the actual account.* methods to work.
// Pass mgr=nil to exercise the "account subsystem not available" path.
func newTestAccountServer(t *testing.T, mgr *account.Manager) *ipc.Server {
	t.Helper()
	srv := ipc.NewServer()
	subs := &Subsystems{Account: mgr}
	registerAccountMethods(srv, subs)
	return srv
}

// mustManager builds a real account.Manager backed by an in-memory sqlite
// database. The manager is configured with the given providers.
func mustManager(t *testing.T, providers map[string]account.ProviderConfig) *account.Manager {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "acct.db"))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store, err := account.NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	tm := &mapTokenManager{data: map[string]string{}}
	registry := account.NewProviderRegistry(providers)
	mgr, err := account.NewManagerWithProviders(store, tm, []byte("test-master-key-0123456789abcdef"), 3600, registry)
	if err != nil {
		t.Fatalf("NewManagerWithProviders: %v", err)
	}
	return mgr
}

type mapTokenManager struct {
	data map[string]string
}

func (m *mapTokenManager) Get(k string) (string, error) {
	v, ok := m.data[k]
	if !ok {
		return "", errors.New("not found")
	}
	return v, nil
}
func (m *mapTokenManager) Set(k, v string) error { m.data[k] = v; return nil }
func (m *mapTokenManager) Delete(k string) error { delete(m.data, k); return nil }
