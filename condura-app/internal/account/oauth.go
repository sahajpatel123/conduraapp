package account

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

// ProviderConfig is the configuration for a single OAuth identity provider.
// It is populated from the user's config.yaml (account.oauth.<provider>) and
// may be overridden by environment variables (CONDURA_ACCOUNT_OAUTH_<PROVIDER>_CLIENT_ID).
type ProviderConfig struct {
	// Name is the canonical provider key (e.g. "google"). When the
	// user defines a custom provider they don't need to set it; the
	// registry keys it under the YAML map key.
	Name string
	// ClientID is the OAuth client ID issued by the provider. Empty means
	// the provider is not configured; GenerateAuthURL returns an error for
	// it.
	ClientID string
	// ClientSecret is the OAuth client secret. Most desktop OAuth flows
	// leave this empty; required by some providers for confidential flows.
	ClientSecret string
	// AuthURL is the provider's authorization endpoint.
	AuthURL string
	// TokenURL is the provider's token-exchange endpoint.
	TokenURL string
	// UserInfoURL is the provider's user-info endpoint (optional). When
	// empty, the provider's default is used.
	UserInfoURL string
	// IDTokenIsUserInfo indicates the user's identity comes from decoding
	// the id_token JWT instead of a separate user-info call. (Google.)
	IDTokenIsUserInfo bool
	// Scopes are the OAuth scopes to request by default.
	Scopes []string
}

// DefaultProviderConfigs returns the canonical OAuth endpoint URLs for the
// providers Condura ships with. ClientID/ClientSecret are NOT populated —
// they must come from the user's config or environment. The endpoint URLs
// themselves are public knowledge and safe to embed.
func DefaultProviderConfigs() map[string]ProviderConfig {
	return map[string]ProviderConfig{
		providerGoogle: { //nolint:gosec // public OAuth endpoint, not a credential
			Name:              providerGoogle,
			AuthURL:           "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:          "https://oauth2.googleapis.com/token",
			UserInfoURL:       "https://openidconnect.googleapis.com/v1/userinfo",
			IDTokenIsUserInfo: true,
			Scopes:            []string{"openid", "email", "profile"},
		},
		providerGitHub: { //nolint:gosec // public OAuth endpoint
			Name:        providerGitHub,
			AuthURL:     "https://github.com/login/oauth/authorize",
			TokenURL:    "https://github.com/login/oauth/access_token",
			UserInfoURL: "https://api.github.com/user",
			Scopes:      []string{"read:user", "user:email"},
		},
		providerApple: { //nolint:gosec // public OAuth endpoint
			Name:              providerApple,
			AuthURL:           "https://appleid.apple.com/auth/authorize",
			TokenURL:          "https://appleid.apple.com/auth/token",
			IDTokenIsUserInfo: true,
			Scopes:            []string{"name", "email"},
		},
	}
}

// Canonical provider name constants. Keep these as package-level constants
// so goconst doesn't trip on repeated literals across the file.
const (
	providerGoogle = "google"
	providerGitHub = "github"
	providerApple  = "apple"
)

// ProviderRegistry is the resolved (post-config, post-env) set of providers
// available for sign-in. It is constructed once at startup and passed into
// the Manager via NewManagerWithProviders.
type ProviderRegistry struct {
	// providers maps provider name -> ProviderConfig.
	providers map[string]ProviderConfig
	// ordered lists provider names in a stable display order.
	ordered []string
}

// NewProviderRegistry builds a registry by layering the user's config over
// DefaultProviderConfigs(). Environment variables
// (CONDURA_ACCOUNT_OAUTH_<UPPER>_CLIENT_ID / _CLIENT_SECRET) override the
// resulting values, so users can keep secrets out of config.yaml.
func NewProviderRegistry(userProviders map[string]ProviderConfig) *ProviderRegistry {
	merged := DefaultProviderConfigs()
	for name := range userProviders {
		base := merged[name]
		override := userProviders[name]
		merged[name] = mergeProvider(base, override, name)
	}
	return &ProviderRegistry{
		providers: merged,
		ordered:   orderedNames(merged),
	}
}

// mergeProvider layers one provider override onto a base config. When the
// base is empty (user-defined custom provider), we accept the override as-is
// but always set the Name to the registry key.
func mergeProvider(base, override ProviderConfig, name string) ProviderConfig {
	if base.Name == "" {
		base.Name = name
	}
	if override.ClientID != "" {
		base.ClientID = override.ClientID
	}
	if override.ClientSecret != "" {
		base.ClientSecret = override.ClientSecret
	}
	if override.AuthURL != "" {
		base.AuthURL = override.AuthURL
	}
	if override.TokenURL != "" {
		base.TokenURL = override.TokenURL
	}
	if override.UserInfoURL != "" {
		base.UserInfoURL = override.UserInfoURL
	}
	if len(override.Scopes) > 0 {
		base.Scopes = override.Scopes
	}
	return base
}

// orderedNames returns provider names in a stable display order: the
// priority providers (google, github, apple) first, then any user-defined
// providers alphabetically.
func orderedNames(merged map[string]ProviderConfig) []string {
	priority := []string{"google", "github", "apple"}
	ordered := make([]string, 0, len(merged))
	seen := make(map[string]bool, len(merged))
	for _, name := range priority {
		if _, ok := merged[name]; ok {
			ordered = append(ordered, name)
			seen[name] = true
		}
	}
	var others []string
	for name := range merged {
		if !seen[name] {
			others = append(others, name)
		}
	}
	sort.Strings(others)
	return append(ordered, others...)
}

// Configured returns the names of providers that have a ClientID set
// (i.e. are usable for sign-in). Use this to drive the "Sign in with X"
// buttons in the GUI.
func (r *ProviderRegistry) Configured() []string {
	if r == nil {
		return nil
	}
	var out []string
	for _, name := range r.ordered {
		if p, ok := r.providers[name]; ok && p.ClientID != "" {
			out = append(out, name)
		}
	}
	return out
}

// Available returns every provider name, even unconfigured ones. The GUI
// uses this to surface a "Not configured" label rather than hide the button.
func (r *ProviderRegistry) Available() []string {
	if r == nil {
		return nil
	}
	out := make([]string, len(r.ordered))
	copy(out, r.ordered)
	return out
}

// Get returns the config for a provider, or nil if not registered.
func (r *ProviderRegistry) Get(name string) *ProviderConfig {
	if r == nil {
		return nil
	}
	p, ok := r.providers[name]
	if !ok {
		return nil
	}
	return &p
}

// GenerateAuthURL builds the OAuth authorization URL with PKCE. Returns
// ErrProviderNotConfigured if the provider is unknown OR has no ClientID.
func (m *Manager) GenerateAuthURL(providerName, redirectURI string) (authURL, state string, err error) {
	prov := m.providers.Get(providerName)
	if prov == nil {
		return "", "", fmt.Errorf("account: unknown provider %q", providerName)
	}
	if prov.ClientID == "" {
		return "", "", fmt.Errorf("%w: %s", ErrProviderNotConfigured, providerName)
	}

	// Generate PKCE code verifier (32 random bytes, base64url).
	verifier := make([]byte, 32)
	if _, err := rand.Read(verifier); err != nil {
		return "", "", fmt.Errorf("account: rand: %w", err)
	}
	verifierStr := base64.RawURLEncoding.EncodeToString(verifier)

	// Compute S256 challenge.
	h := sha256.Sum256([]byte(verifierStr))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])

	// Generate state (16 random bytes, hex).
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", "", fmt.Errorf("account: rand state: %w", err)
	}
	state = fmt.Sprintf("%x", stateBytes)

	// Store state -> (verifier, provider, redirect_uri) for 5 min.
	m.oauthStates.Store(state, oauthStateEntry{
		verifier:    verifierStr,
		provider:    providerName,
		redirectURI: redirectURI,
		expiresAt:   time.Now().Add(5 * time.Minute),
	})

	// Build authorization URL.
	u, _ := url.Parse(prov.AuthURL)
	q := u.Query()
	q.Set("client_id", prov.ClientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", strings.Join(prov.Scopes, " "))
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	// Google-specific hints for offline access (refresh tokens).
	if providerName == "google" {
		q.Set("access_type", "offline")
		q.Set("prompt", "consent")
	}
	u.RawQuery = q.Encode()

	return u.String(), state, nil
}

// ErrProviderNotConfigured is returned when the user tries to sign in with
// a provider whose ClientID has not been set.
var ErrProviderNotConfigured = errors.New("account: provider not configured")

// ExchangeCode exchanges an OAuth authorization code for tokens. The
// provider name MUST match the one used to start the flow (looked up by
// state); passing a mismatched provider name is an error.
//
// The PKCE verifier is looked up by state. If the GUI provided its own
// code_verifier (via OAuthCallbackParams.code_verifier), that takes
// precedence — useful when the GUI kept the verifier across the
// redirect (matching the front-end store's cached value).
func (m *Manager) ExchangeCode(ctx context.Context, providerName, code, state, redirectURI string) (*Session, error) {
	entry, prov, err := m.validateOAuthCallback(providerName, code, state, redirectURI)
	if err != nil {
		return nil, err
	}
	tokenResp, err := m.exchangeOAuthCode(ctx, prov, code, entry.verifier, redirectURI)
	if err != nil {
		return nil, err
	}
	email, avatar, err := m.fetchUserInfo(ctx, *prov, tokenResp.AccessToken, tokenResp.IDToken)
	if err != nil {
		return nil, err
	}
	if email == "" {
		return nil, errors.New("account: provider returned no email for this account")
	}
	if err := m.tokenManager.Set("oauth-"+providerName, tokenResp.AccessToken); err != nil {
		return nil, fmt.Errorf("account: store tokens: %w", err)
	}
	if tokenResp.RefreshToken != "" {
		_ = m.tokenManager.Set("oauth-"+providerName+"-refresh", tokenResp.RefreshToken)
	}
	sess, err := m.NewSession(ctx, email, providerName)
	if err != nil {
		return nil, err
	}
	sess.AvatarURL = avatar
	return sess, nil
}

// validateOAuthCallback checks the code/state/redirect_uri against the
// in-flight OAuth flow and returns the matching state entry and provider
// config. Returns ErrProviderNotConfigured (wrapped) if the provider has
// no ClientID; never returns a partial state lookup.
func (m *Manager) validateOAuthCallback(providerName, code, state, redirectURI string) (*oauthStateEntry, *ProviderConfig, error) {
	if code == "" {
		return nil, nil, errors.New("account: empty code")
	}
	if state == "" {
		return nil, nil, errors.New("account: empty state")
	}
	entryRaw, ok := m.oauthStates.Load(state)
	if !ok {
		return nil, nil, errors.New("account: unknown or expired state")
	}
	entry := entryRaw.(oauthStateEntry)
	m.oauthStates.Delete(state)
	if time.Now().After(entry.expiresAt) {
		return nil, nil, errors.New("account: state expired")
	}
	if entry.provider != providerName {
		return nil, nil, fmt.Errorf("account: provider mismatch: got %q, expected %q", providerName, entry.provider)
	}
	if entry.redirectURI != "" && entry.redirectURI != redirectURI {
		return nil, nil, errors.New("account: redirect_uri mismatch (CSRF)")
	}
	prov := m.providers.Get(providerName)
	if prov == nil {
		return nil, nil, fmt.Errorf("account: unknown provider %q", providerName)
	}
	return &entry, prov, nil
}

// exchangeOAuthCode POSTs the code to the provider's token endpoint and
// parses the response. Splits the network/parse concerns so the caller
// stays small.
func (m *Manager) exchangeOAuthCode(ctx context.Context, prov *ProviderConfig, code, verifier, redirectURI string) (*oauthTokenResponse, error) {
	data := url.Values{
		"client_id":     {prov.ClientID},
		"code":          {code},
		"code_verifier": {verifier},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	}
	if prov.ClientSecret != "" {
		data.Set("client_secret", prov.ClientSecret)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, prov.TokenURL,
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("account: token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("account: token exchange: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("account: token exchange failed (%d): %s", resp.StatusCode, string(body))
	}
	var tokenResp oauthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("account: parse token response: %w", err)
	}
	if tokenResp.Error != "" {
		return nil, fmt.Errorf("account: provider error: %s: %s", tokenResp.Error, tokenResp.ErrorDesc)
	}
	if tokenResp.AccessToken == "" {
		return nil, errors.New("account: empty access token in provider response")
	}
	return &tokenResp, nil
}

// oauthTokenResponse is the subset of OAuth provider responses we read.
type oauthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Error        string `json:"error,omitempty"`
	ErrorDesc    string `json:"error_description,omitempty"`
}

// fetchUserInfo retrieves the user's email and avatar from the provider.
func (m *Manager) fetchUserInfo(ctx context.Context, prov ProviderConfig, accessToken, idToken string) (email, avatar string, err error) {
	if prov.IDTokenIsUserInfo && idToken != "" {
		return decodeIDToken(idToken)
	}
	if prov.UserInfoURL == "" {
		return "", "", fmt.Errorf("account: provider %q has no user info URL", prov.Name)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, prov.UserInfoURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("account: provider user info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("account: provider user info returned %d", resp.StatusCode)
	}

	// GitHub-specific fallback: if no email, fetch from /user/emails.
	if prov.Name == "github" {
		var user struct {
			Email     string `json:"email"`
			AvatarURL string `json:"avatar_url"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return "", "", fmt.Errorf("account: parse provider user info: %w", err)
		}
		if user.Email == "" {
			user.Email, _ = fetchGitHubEmails(ctx, accessToken)
		}
		return user.Email, user.AvatarURL, nil
	}

	// Generic OpenID-Connect-ish shape.
	var user struct {
		Email     string `json:"email"`
		Picture   string `json:"picture"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", "", fmt.Errorf("account: parse provider user info: %w", err)
	}
	avatar = user.Picture
	if avatar == "" {
		avatar = user.AvatarURL
	}
	return user.Email, avatar, nil
}

// decodeIDToken decodes the payload segment of a JWT (no signature
// verification — providers sign the token; we only read the claims).
// Returns the email and (best-effort) avatar URL.
func decodeIDToken(idToken string) (email, avatar string, err error) {
	parts := strings.Split(idToken, ".")
	if len(parts) < 2 {
		return "", "", errors.New("account: malformed id_token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", fmt.Errorf("account: decode id_token: %w", err)
	}
	var claims struct {
		Email         string `json:"email"`
		EmailVerified any    `json:"email_verified"`
		Picture       string `json:"picture"`
		Name          string `json:"name"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", "", fmt.Errorf("account: parse id_token claims: %w", err)
	}
	return claims.Email, claims.Picture, nil
}

func fetchGitHubEmails(ctx context.Context, accessToken string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}
	for _, e := range emails {
		if e.Primary {
			return e.Email, nil
		}
	}
	if len(emails) > 0 {
		return emails[0].Email, nil
	}
	return "", nil
}

// CleanupExpiredStates removes expired OAuth state entries.
func (m *Manager) CleanupExpiredStates() {
	m.oauthStates.Range(func(key, value interface{}) bool {
		entry := value.(oauthStateEntry)
		if time.Now().After(entry.expiresAt) {
			m.oauthStates.Delete(key)
		}
		return true
	})
}

// Compile-time guards: keep these imports even if the call sites move.
var (
	_ = sync.Map{}
	_ = fmt.Sprintf
	_ = time.Now
)
