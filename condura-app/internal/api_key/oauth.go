package api_key

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OAuthProvider implements the Authorization Code + PKCE flow for a
// single upstream OAuth 2.0 server.
//
// Implementations are expected to be stateless and safe for concurrent use
// across many flow invocations. State (verifier, code, expiry) lives in
// the returned *OAuthFlow handle and is owned by the caller.
type OAuthProvider interface {
	// Name returns the canonical provider name (e.g. "google").
	Name() string
	// AuthorizeURL builds the URL the user should be sent to. The state
	// and PKCE challenge are embedded as query parameters.
	AuthorizeURL(state, codeChallenge, redirectURI string, scopes []string) (string, error)
	// ExchangeCode exchanges an authorization code for an access token.
	// Returns Token with Refresh populated when the server supports it.
	ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI string) (Token, error)
	// Refresh exchanges a refresh token for a new access token.
	Refresh(ctx context.Context, refreshToken string) (Token, error)
	// Revoke invalidates a token at the provider. Best-effort: if the
	// provider doesn't support revocation, this returns nil.
	Revoke(ctx context.Context, token string) error
}

// Token is the result of an OAuth token exchange.
type Token struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Scope        string
	ExpiresAt    time.Time
}

// Expired reports whether the access token is past its expiry (with a 30s
// safety margin).
func (t Token) Expired() bool {
	if t.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().Add(30 * time.Second).After(t.ExpiresAt)
}

// -----------------------------------------------------------------------------
// PKCE helpers
// -----------------------------------------------------------------------------

// GeneratePKCE returns a fresh (verifier, challenge) pair.
func GeneratePKCE() (verifier, challenge string, err error) {
	// RFC 7636: verifier is 43-128 chars from [A-Z a-z 0-9 - . _ ~].
	// We use 32 random bytes base64url-encoded (no padding) = 43 chars.
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("oauth: rand: %w", err)
	}
	verifier = base64.RawURLEncoding.EncodeToString(buf)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

// GenerateState returns a random opaque state string for CSRF protection.
func GenerateState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// -----------------------------------------------------------------------------
// Flow
// -----------------------------------------------------------------------------

// Flow holds the state of an in-progress OAuth flow. Callers create one
// with StartFlow, send the user to AuthURL, then call Complete with the
// callback's code.
type Flow struct {
	Provider     OAuthProvider
	State        string
	CodeVerifier string
	RedirectURI  string
	Scopes       []string
	StartedAt    time.Time
}

// StartFlow prepares a new authorization flow.
func StartFlow(p OAuthProvider, redirectURI string, scopes []string) (*Flow, error) {
	state, err := GenerateState()
	if err != nil {
		return nil, err
	}
	verifier, _, err := GeneratePKCE()
	if err != nil {
		return nil, err
	}
	// We need the challenge for the auth URL but the verifier for exchange;
	// recompute the challenge here.
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	// Validate the auth URL builds before returning.
	if _, err := p.AuthorizeURL(state, challenge, redirectURI, scopes); err != nil {
		return nil, err
	}
	return &Flow{
		Provider:     p,
		State:        state,
		CodeVerifier: verifier,
		RedirectURI:  redirectURI,
		Scopes:       scopes,
		StartedAt:    time.Now(),
	}, nil
}

// AuthURL returns the URL the user should be sent to.
func (f *Flow) AuthURL() (string, error) {
	sum := sha256.Sum256([]byte(f.CodeVerifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return f.Provider.AuthorizeURL(f.State, challenge, f.RedirectURI, f.Scopes)
}

// Complete exchanges the authorization code for tokens.
func (f *Flow) Complete(ctx context.Context, code string) (Token, error) {
	return f.Provider.ExchangeCode(ctx, code, f.CodeVerifier, f.RedirectURI)
}

// Common errors.
var (
	ErrInvalidState     = errors.New("oauth: invalid state")
	ErrInvalidResponse  = errors.New("oauth: invalid response")
	ErrTokenExchange    = errors.New("oauth: token exchange failed")
	ErrUnsupportedScope = errors.New("oauth: unsupported scope")
)

// -----------------------------------------------------------------------------
// Low-level helpers shared by concrete providers
// -----------------------------------------------------------------------------

// httpDoer is the subset of *http.Client we need. Allows injection in tests.
type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// asDoer adapts an http.RoundTripper to httpDoer. This lets tests inject a
// transport (e.g. one that rewrites URLs) without standing up a full client.
type asDoer struct{ rt http.RoundTripper }

func (d asDoer) Do(r *http.Request) (*http.Response, error) { return d.rt.RoundTrip(r) }

// roundTripperFor returns an httpDoer that uses rt. If rt is nil, it returns
// the default http.Client.
func roundTripperFor(rt http.RoundTripper) httpDoer {
	if rt == nil {
		return &http.Client{Timeout: 30 * time.Second}
	}
	return asDoer{rt: rt}
}

// formPost POSTs application/x-www-form-urlencoded with the given values and
// returns the response body. Used by ExchangeCode and Refresh.
func formPost(ctx context.Context, client httpDoer, endpoint string, values url.Values, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, 0, fmt.Errorf("oauth: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("oauth: do: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body := make([]byte, 0, 1024)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	return body, resp.StatusCode, nil
}
