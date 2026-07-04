package api_key

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// PKCE / state
// -----------------------------------------------------------------------------

func TestGeneratePKCE(t *testing.T) {
	v, c, err := GeneratePKCE()
	require.NoError(t, err)
	assert.NotEmpty(t, v)
	assert.NotEmpty(t, c)
	assert.NotEqual(t, v, c)
}

func TestGeneratePKCE_Unique(t *testing.T) {
	v1, _, _ := GeneratePKCE()
	v2, _, _ := GeneratePKCE()
	assert.NotEqual(t, v1, v2)
}

func TestGenerateState(t *testing.T) {
	s, err := GenerateState()
	require.NoError(t, err)
	assert.NotEmpty(t, s)
	assert.Len(t, s, 32) // 16 bytes hex
}

func TestGenerateState_Unique(t *testing.T) {
	s1, _ := GenerateState()
	s2, _ := GenerateState()
	assert.NotEqual(t, s1, s2)
}

// -----------------------------------------------------------------------------
// Flow
// -----------------------------------------------------------------------------

func TestStartFlow(t *testing.T) {
	p := NewGoogleProvider("client-id", "")
	f, err := StartFlow(p, "http://localhost:9999/callback", []string{"openid"})
	require.NoError(t, err)
	assert.NotEmpty(t, f.State)
	assert.NotEmpty(t, f.CodeVerifier)
	assert.Equal(t, p, f.Provider)
	assert.Equal(t, "http://localhost:9999/callback", f.RedirectURI)
	assert.Equal(t, []string{"openid"}, f.Scopes)
}

func TestStartFlow_EmptyClientID(t *testing.T) {
	p := NewGoogleProvider("", "")
	_, err := StartFlow(p, "http://localhost/cb", nil)
	assert.Error(t, err)
}

func TestFlow_AuthURL(t *testing.T) {
	p := NewGoogleProvider("client-id", "")
	f, err := StartFlow(p, "http://localhost/cb", []string{"openid"})
	require.NoError(t, err)
	u, err := f.AuthURL()
	require.NoError(t, err)
	assert.Contains(t, u, "accounts.google.com")
	assert.Contains(t, u, "client_id=client-id")
	assert.Contains(t, u, "code_challenge_method=S256")
	assert.Contains(t, u, "state="+url.QueryEscape(f.State))
}

func TestToken_Expired(t *testing.T) {
	t1 := Token{ExpiresAt: time.Now().Add(time.Hour)}
	assert.False(t, t1.Expired())
	t2 := Token{ExpiresAt: time.Now().Add(-time.Hour)}
	assert.True(t, t2.Expired())
	t3 := Token{} // no expiry
	assert.False(t, t3.Expired())
}

// -----------------------------------------------------------------------------
// Google provider — AuthorizeURL
// -----------------------------------------------------------------------------

func TestGoogle_AuthorizeURL(t *testing.T) {
	g := NewGoogleProvider("cid", "")
	u, err := g.AuthorizeURL("state1", "challenge1", "http://localhost/cb", []string{"s1", "s2"})
	require.NoError(t, err)
	assert.Contains(t, u, "response_type=code")
	assert.Contains(t, u, "client_id=cid")
	assert.Contains(t, u, "state=state1")
	assert.Contains(t, u, "code_challenge=challenge1")
	assert.Contains(t, u, "redirect_uri=")
	assert.Contains(t, u, "scope=s1+s2")
	assert.Contains(t, u, "access_type=offline")
	assert.Contains(t, u, "prompt=consent")
}

func TestGoogle_AuthorizeURL_DefaultScope(t *testing.T) {
	g := NewGoogleProvider("cid", "")
	u, err := g.AuthorizeURL("s", "c", "http://localhost/cb", nil)
	require.NoError(t, err)
	assert.Contains(t, u, "scope=")
	// Default scope includes generative-language.
	assert.Contains(t, u, "generative-language")
}

func TestGoogle_AuthorizeURL_NoClientID(t *testing.T) {
	g := NewGoogleProvider("", "")
	_, err := g.AuthorizeURL("s", "c", "http://localhost/cb", nil)
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Google provider — ExchangeCode (mocked server)
// -----------------------------------------------------------------------------

func TestGoogle_ExchangeCode_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.NotFound(w, r)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		assert.Equal(t, "authorization_code", r.Form.Get("grant_type"))
		assert.Equal(t, "the-code", r.Form.Get("code"))
		assert.Equal(t, "the-verifier", r.Form.Get("code_verifier"))
		assert.Equal(t, "http://localhost/cb", r.Form.Get("redirect_uri"))
		assert.Equal(t, "cid", r.Form.Get("client_id"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"access_token": "ya29.AAAA",
			"refresh_token": "1//BBBB",
			"token_type": "Bearer",
			"expires_in": 3600,
			"scope": "openid"
		}`))
	}))
	defer srv.Close()

	// Point the provider at the test server.
	g := NewGoogleProvider("cid", "")
	// We can't easily override the endpoint, so test the lower-level
	// behavior via the AuthorizeURL/ExchangeCode methods using a custom
	// HTTP client. For the exchange to hit the test server, we patch the
	// token endpoint through a custom HTTPClient that rewrites the URL.
	g.Transport = rewriteTransport{target: srv.URL, base: http.DefaultTransport}

	token, err := g.ExchangeCode(context.Background(), "the-code", "the-verifier", "http://localhost/cb")
	require.NoError(t, err)
	assert.Equal(t, "ya29.AAAA", token.AccessToken)
	assert.Equal(t, "1//BBBB", token.RefreshToken)
	assert.Equal(t, "Bearer", token.TokenType)
	assert.Equal(t, "openid", token.Scope)
	assert.False(t, token.ExpiresAt.IsZero())
}

func TestGoogle_ExchangeCode_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"bad code"}`))
	}))
	defer srv.Close()
	g := NewGoogleProvider("cid", "")
	g.Transport = rewriteTransport{target: srv.URL, base: http.DefaultTransport}
	_, err := g.ExchangeCode(context.Background(), "bad", "v", "http://localhost/cb")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenExchange)
}

func TestGoogle_ExchangeCode_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()
	g := NewGoogleProvider("cid", "")
	g.Transport = rewriteTransport{target: srv.URL, base: http.DefaultTransport}
	_, err := g.ExchangeCode(context.Background(), "c", "v", "http://localhost/cb")
	assert.Error(t, err)
}

func TestGoogle_ExchangeCode_EmptyAccessToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"token_type":"Bearer"}`))
	}))
	defer srv.Close()
	g := NewGoogleProvider("cid", "")
	g.Transport = rewriteTransport{target: srv.URL, base: http.DefaultTransport}
	_, err := g.ExchangeCode(context.Background(), "c", "v", "http://localhost/cb")
	assert.ErrorIs(t, err, ErrInvalidResponse)
}

func TestGoogle_ExchangeCode_InlineError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"error":"server_error","error_description":"x"}`))
	}))
	defer srv.Close()
	g := NewGoogleProvider("cid", "")
	g.Transport = rewriteTransport{target: srv.URL, base: http.DefaultTransport}
	_, err := g.ExchangeCode(context.Background(), "c", "v", "http://localhost/cb")
	assert.ErrorIs(t, err, ErrTokenExchange)
}

// -----------------------------------------------------------------------------
// Google provider — Refresh
// -----------------------------------------------------------------------------

func TestGoogle_Refresh_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		assert.Equal(t, "refresh_token", r.Form.Get("grant_type"))
		assert.Equal(t, "1//OLD", r.Form.Get("refresh_token"))
		_, _ = w.Write([]byte(`{
			"access_token": "ya29.NEW",
			"token_type": "Bearer",
			"expires_in": 3600,
			"scope": "openid"
		}`))
	}))
	defer srv.Close()
	g := NewGoogleProvider("cid", "")
	g.Transport = rewriteTransport{target: srv.URL, base: http.DefaultTransport}
	tok, err := g.Refresh(context.Background(), "1//OLD")
	require.NoError(t, err)
	assert.Equal(t, "ya29.NEW", tok.AccessToken)
	assert.Equal(t, "1//OLD", tok.RefreshToken, "refresh token should be preserved")
}

func TestGoogle_Refresh_Empty(t *testing.T) {
	g := NewGoogleProvider("cid", "")
	_, err := g.Refresh(context.Background(), "")
	assert.Error(t, err)
}

func TestGoogle_Refresh_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer srv.Close()
	g := NewGoogleProvider("cid", "")
	g.Transport = rewriteTransport{target: srv.URL, base: http.DefaultTransport}
	_, err := g.Refresh(context.Background(), "1//x")
	assert.Error(t, err)
}

func TestGoogle_Refresh_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()
	g := NewGoogleProvider("cid", "")
	g.Transport = rewriteTransport{target: srv.URL, base: http.DefaultTransport}
	_, err := g.Refresh(context.Background(), "1//x")
	assert.Error(t, err)
}

func TestGoogle_Refresh_EmptyAccessToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"token_type":"Bearer"}`))
	}))
	defer srv.Close()
	g := NewGoogleProvider("cid", "")
	g.Transport = rewriteTransport{target: srv.URL, base: http.DefaultTransport}
	_, err := g.Refresh(context.Background(), "1//x")
	assert.ErrorIs(t, err, ErrInvalidResponse)
}

// -----------------------------------------------------------------------------
// Google provider — Revoke
// -----------------------------------------------------------------------------

func TestGoogle_Revoke_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		assert.Equal(t, "the-token", r.Form.Get("token"))
		w.WriteHeader(200)
	}))
	defer srv.Close()
	g := NewGoogleProvider("cid", "")
	g.Transport = rewriteTransport{target: srv.URL, base: http.DefaultTransport}
	assert.NoError(t, g.Revoke(context.Background(), "the-token"))
}

func TestGoogle_Revoke_Empty(t *testing.T) {
	g := NewGoogleProvider("cid", "")
	assert.NoError(t, g.Revoke(context.Background(), ""))
}

func TestGoogle_Revoke_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	}))
	defer srv.Close()
	g := NewGoogleProvider("cid", "")
	g.Transport = rewriteTransport{target: srv.URL, base: http.DefaultTransport}
	assert.Error(t, g.Revoke(context.Background(), "x"))
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// rewriteTransport routes all requests to the test server URL. Used to
// short-circuit Google's hard-coded endpoints. The original request's
// path is preserved (so /token → test server's /token, /revoke → /revoke).
type rewriteTransport struct {
	target string
	base   http.RoundTripper
}

func (t rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	u, err := url.Parse(t.target)
	if err != nil {
		return nil, err
	}
	rewritten := req.Clone(req.Context())
	rewritten.URL.Scheme = u.Scheme
	rewritten.URL.Host = u.Host
	// Preserve the original path/query.
	rewritten.RequestURI = ""
	return t.base.RoundTrip(rewritten)
}

// Sanity: ensure the formPost helper handles headers correctly.
func TestFormPost_Headers(t *testing.T) {
	// Just exercise the path through Google_ExchangeCode (already does it).
	// The header assertion is in TestGoogle_ExchangeCode_OK.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		_, _ = w.Write([]byte(`{"access_token":"a","expires_in":60}`))
	}))
	defer srv.Close()
	g := NewGoogleProvider("cid", "")
	g.Transport = rewriteTransport{target: srv.URL, base: http.DefaultTransport}
	_, _ = g.ExchangeCode(context.Background(), "c", "v", "cb")
}

// TestFlow_Complete — happy path through Flow.
func TestFlow_Complete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"flow-tok","expires_in":60}`))
	}))
	defer srv.Close()
	g := NewGoogleProvider("cid", "")
	g.Transport = rewriteTransport{target: srv.URL, base: http.DefaultTransport}
	f, err := StartFlow(g, "http://localhost/cb", nil)
	require.NoError(t, err)
	tok, err := f.Complete(context.Background(), "code")
	require.NoError(t, err)
	assert.Equal(t, "flow-tok", tok.AccessToken)
}

// Sanity: the provider name is "google".
func TestGoogle_Name(t *testing.T) {
	g := NewGoogleProvider("cid", "")
	assert.Equal(t, "google", g.Name())
}

// Sanity: NewGoogleProvider sets timeout.
func TestNewGoogleProvider_Defaults(t *testing.T) {
	g := NewGoogleProvider("cid", "")
	assert.Equal(t, "cid", g.ClientID)
	assert.Equal(t, "", g.ClientSecret)
	assert.NotNil(t, g.HTTPClient)
}

// Ensure all imports are used.
var _ = strings.NewReader
