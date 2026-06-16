package account

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// OAuthProvider defines a single OAuth identity provider.
type OAuthProvider struct {
	Name     string
	AuthURL  string
	TokenURL string
	ClientID string
	Scopes   []string
}

// DefaultOAuthProviders returns the supported providers.
//
//nolint:gosec // public OAuth endpoints, no credentials hardcoded
func DefaultOAuthProviders() map[string]OAuthProvider {
	return map[string]OAuthProvider{
		"google": { //nolint:gosec // public endpoint
			Name:     "google",
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
			ClientID: "", //nolint:gosec
			Scopes:   []string{"openid", "email", "profile"},
		},
		"github": { //nolint:gosec // public endpoint
			Name:     "github",
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
			ClientID: "", //nolint:gosec
			Scopes:   []string{"user:email"},
		},
	}
}

// GenerateAuthURL builds the OAuth authorization URL with PKCE.
func (m *Manager) GenerateAuthURL(providerName, redirectURI string) (authURL, state string, err error) {
	prov, ok := DefaultOAuthProviders()[providerName]
	if !ok {
		return "", "", fmt.Errorf("account: unknown provider %q", providerName)
	}
	if prov.ClientID == "" {
		return "", "", fmt.Errorf("account: provider %q has no client_id configured", providerName)
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

	// Store state → verifier mapping (5 min TTL).
	m.oauthStates.Store(state, oauthStateEntry{
		verifier:  verifierStr,
		provider:  providerName,
		expiresAt: time.Now().Add(5 * time.Minute),
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
	u.RawQuery = q.Encode()

	return u.String(), state, nil
}

// ExchangeCode exchanges an OAuth authorization code for tokens.
// On success, OAuth tokens are stored via the token manager and
// a local session is created.
func (m *Manager) ExchangeCode(ctx context.Context, providerName, code, state string, redirectURI string) (*Session, error) {
	// Look up state.
	entryRaw, ok := m.oauthStates.Load(state)
	if !ok {
		return nil, fmt.Errorf("account: unknown or expired state")
	}
	entry := entryRaw.(oauthStateEntry)
	m.oauthStates.Delete(state)

	if time.Now().After(entry.expiresAt) {
		return nil, fmt.Errorf("account: state expired")
	}
	if entry.provider != providerName {
		return nil, fmt.Errorf("account: provider mismatch: got %q, expected %q", providerName, entry.provider)
	}

	prov, ok := DefaultOAuthProviders()[providerName]
	if !ok {
		return nil, fmt.Errorf("account: unknown provider %q", providerName)
	}
	if prov.ClientID == "" {
		return nil, fmt.Errorf("account: provider %q has no client_id configured", providerName)
	}

	// Exchange code for tokens.
	data := url.Values{
		"client_id":     {prov.ClientID},
		"code":          {code},
		"code_verifier": {entry.verifier},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, prov.TokenURL, strings.NewReader(data.Encode()))
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

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token,omitempty"`
		IDToken      string `json:"id_token,omitempty"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("account: parse token response: %w", err)
	}

	// Decode ID token (Google) or fetch user info (GitHub).
	email, avatar, err := m.fetchUserInfo(ctx, providerName, tokenResp.AccessToken, tokenResp.IDToken)
	if err != nil {
		return nil, err
	}

	// Store OAuth tokens.
	if err := m.tokenManager.Set("oauth-"+providerName, tokenResp.AccessToken); err != nil {
		return nil, fmt.Errorf("account: store tokens: %w", err)
	}
	if tokenResp.RefreshToken != "" {
		_ = m.tokenManager.Set("oauth-"+providerName+"-refresh", tokenResp.RefreshToken)
	}

	// Create local session.
	sess, err := m.NewSession(ctx, email, providerName)
	if err != nil {
		return nil, err
	}
	sess.AvatarURL = avatar
	return sess, nil
}

// fetchUserInfo retrieves the user's email and avatar from the
// provider. For Google, we decode the id_token JWT. For GitHub,
// we call the /user API.
func (m *Manager) fetchUserInfo(ctx context.Context, provider, accessToken, idToken string) (email, avatar string, err error) {
	switch provider {
	case "google":
		if idToken == "" {
			return "", "", fmt.Errorf("account: no id_token in Google response")
		}
		// Decode JWT body (second segment, base64url).
		email, avatar = decodeGoogleIDToken(idToken)
		return email, avatar, nil
	case "github":
		return fetchGitHubUser(ctx, accessToken)
	default:
		return "", "", fmt.Errorf("account: no user info for provider %q", provider)
	}
}

func decodeGoogleIDToken(idToken string) (email, avatar string) {
	parts := strings.Split(idToken, ".")
	if len(parts) < 2 {
		return "", ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", ""
	}
	var claims struct {
		Email   string `json:"email"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", ""
	}
	return claims.Email, claims.Picture
}

func fetchGitHubUser(ctx context.Context, accessToken string) (email, avatar string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("account: github user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	var user struct {
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", "", fmt.Errorf("account: parse github user: %w", err)
	}
	// GitHub may not return email if private. Fetch emails separately.
	if user.Email == "" {
		user.Email, _ = fetchGitHubEmails(ctx, accessToken)
	}
	return user.Email, user.AvatarURL, nil
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

// unused import guards
var _ = fmt.Sprintf
var _ = time.Now
var _ = io.ReadAll
var _ = http.DefaultClient
var _ = sync.Map{}
