package api_key

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Google OAuth 2.0 endpoints. Google's refresh tokens are issued once; the
// access token expires in 1 hour. Scopes here are the minimal set needed
// for the Gemini Developer API (which is the consumer for our OAuth flow).
const (
	googleAuthEndpoint     = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenEndpoint    = "https://oauth2.googleapis.com/token"
	googleRevokeEndpoint   = "https://oauth2.googleapis.com/revoke"
	googleDefaultScope     = "https://www.googleapis.com/auth/generative-language"
	googleUserinfoEndpoint = "https://openidconnect.googleapis.com/v1/userinfo"
)

// GoogleProvider implements OAuthProvider for Google.
//
// Config: set ClientID and ClientSecret (from Google Cloud Console →
// Credentials → OAuth 2.0 Client IDs). For an installed/desktop app, the
// client secret is optional in newer flows; we still accept it for
// flexibility.
type GoogleProvider struct {
	ClientID     string
	ClientSecret string
	HTTPClient   httpDoer          // defaults to http.Client
	Transport    http.RoundTripper // optional, used to build HTTPClient
}

// NewGoogleProvider returns a GoogleProvider with default HTTP client.
func NewGoogleProvider(clientID, clientSecret string) *GoogleProvider {
	return &GoogleProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (g *GoogleProvider) client() httpDoer {
	// Transport takes priority — it's the test injection point.
	if g.Transport != nil {
		return roundTripperFor(g.Transport)
	}
	if g.HTTPClient != nil {
		return g.HTTPClient
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func (g *GoogleProvider) Name() string { return ProviderGoogle }

// AuthorizeURL builds the Google authorization URL with PKCE.
func (g *GoogleProvider) AuthorizeURL(state, codeChallenge, redirectURI string, scopes []string) (string, error) {
	if g.ClientID == "" {
		return "", fmt.Errorf("oauth/google: ClientID is empty")
	}
	if len(scopes) == 0 {
		scopes = []string{googleDefaultScope}
	}
	q := url.Values{}
	q.Set("client_id", g.ClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", strings.Join(scopes, " "))
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	q.Set("access_type", "offline") // request refresh token
	q.Set("prompt", "consent")      // force consent to get refresh token
	q.Set("include_granted_scopes", "true")
	return googleAuthEndpoint + "?" + q.Encode(), nil
}

// googleTokenResponse is Google's OAuth token response.
type googleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token,omitempty"`
	Error        string `json:"error,omitempty"`
	ErrorDesc    string `json:"error_description,omitempty"`
}

// ExchangeCode exchanges an authorization code for tokens.
func (g *GoogleProvider) ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI string) (Token, error) {
	values := url.Values{}
	values.Set("code", code)
	values.Set("client_id", g.ClientID)
	values.Set("code_verifier", codeVerifier)
	values.Set("grant_type", "authorization_code")
	values.Set("redirect_uri", redirectURI)
	if g.ClientSecret != "" {
		values.Set("client_secret", g.ClientSecret)
	}
	body, status, err := formPost(ctx, g.client(), googleTokenEndpoint, values, nil)
	if err != nil {
		return Token{}, err
	}
	if status >= 400 {
		return Token{}, fmt.Errorf("%w: %d: %s", ErrTokenExchange, status, string(body))
	}
	var r googleTokenResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return Token{}, fmt.Errorf("oauth/google: decode: %w", err)
	}
	if r.Error != "" {
		return Token{}, fmt.Errorf("%w: %s: %s", ErrTokenExchange, r.Error, r.ErrorDesc)
	}
	if r.AccessToken == "" {
		return Token{}, ErrInvalidResponse
	}
	return Token{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		TokenType:    r.TokenType,
		Scope:        r.Scope,
		ExpiresAt:    time.Now().Add(time.Duration(r.ExpiresIn) * time.Second),
	}, nil
}

// Refresh trades a refresh token for a new access token.
func (g *GoogleProvider) Refresh(ctx context.Context, refreshToken string) (Token, error) {
	if refreshToken == "" {
		return Token{}, fmt.Errorf("oauth/google: empty refresh token")
	}
	values := url.Values{}
	values.Set("refresh_token", refreshToken)
	values.Set("client_id", g.ClientID)
	values.Set("grant_type", "refresh_token")
	if g.ClientSecret != "" {
		values.Set("client_secret", g.ClientSecret)
	}
	body, status, err := formPost(ctx, g.client(), googleTokenEndpoint, values, nil)
	if err != nil {
		return Token{}, err
	}
	if status >= 400 {
		return Token{}, fmt.Errorf("%w: %d: %s", ErrTokenExchange, status, string(body))
	}
	var r googleTokenResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return Token{}, fmt.Errorf("oauth/google: decode: %w", err)
	}
	if r.Error != "" {
		return Token{}, fmt.Errorf("%w: %s: %s", ErrTokenExchange, r.Error, r.ErrorDesc)
	}
	if r.AccessToken == "" {
		return Token{}, ErrInvalidResponse
	}
	return Token{
		AccessToken: r.AccessToken,
		// Google does not rotate refresh tokens; preserve the existing one.
		RefreshToken: refreshToken,
		TokenType:    r.TokenType,
		Scope:        r.Scope,
		ExpiresAt:    time.Now().Add(time.Duration(r.ExpiresIn) * time.Second),
	}, nil
}

// Revoke invalidates an access (or refresh) token at Google.
func (g *GoogleProvider) Revoke(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleRevokeEndpoint,
		strings.NewReader(url.Values{"token": {token}}.Encode()))
	if err != nil {
		return fmt.Errorf("oauth/google: build revoke: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := g.client().Do(req)
	if err != nil {
		return fmt.Errorf("oauth/google: revoke: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("oauth/google: revoke: %d", resp.StatusCode)
	}
	return nil
}
