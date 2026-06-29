package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/account"
	"github.com/sahajpatel123/conduraapp/internal/ipc"
)

// OAuthRedirectURI is the desktop deep-link the OAuth providers redirect to.
// Must match the value registered in each OAuth app's allowed-redirects.
const OAuthRedirectURI = "condura://auth/callback"

func registerAccountMethods(srv *ipc.Server, subs *Subsystems) {
	if subs.Account == nil {
		registerAccountNotAvailable(srv)
		return
	}
	// Each handler is registered via a small factory so the dispatcher's
	// cyclomatic/cognitive complexity stays under the lint ceiling. The
	// factories close over `subs` and return an ipc handler function.
	srv.Register("account.status", accountStatusHandler(subs))
	srv.Register("account.providers", accountProvidersHandler(subs))
	srv.Register("account.oauth_url", accountOAuthURLHandler(subs))
	srv.Register("account.oauth_callback", accountOAuthCallbackHandler(subs))
	srv.Register("account.magic_link", accountMagicLinkHandler(subs))
	srv.Register("account.logout", accountLogoutHandler(subs))
}

func accountStatusHandler(subs *Subsystems) ipc.HandlerFunc {
	return func(ctx context.Context, _ json.RawMessage) (any, error) {
		sess, err := subs.Account.Status(ctx)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		if sess == nil {
			return signedOutStatus(configuredProviders(subs)), nil
		}
		return signedInStatus(sess, configuredProviders(subs)), nil
	}
}

func accountProvidersHandler(subs *Subsystems) ipc.HandlerFunc {
	return func(_ context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{"providers": configuredProviders(subs)}, nil
	}
}

func accountOAuthURLHandler(subs *Subsystems) ipc.HandlerFunc {
	return func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Provider string `json:"provider"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		url, state, err := subs.Account.GenerateAuthURL(p.Provider, OAuthRedirectURI)
		if err != nil {
			return nil, oauthURLError(p.Provider, err)
		}
		return map[string]any{"url": url, "state": state}, nil
	}
}

func accountOAuthCallbackHandler(subs *Subsystems) ipc.HandlerFunc {
	return func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Provider     string `json:"provider"`
			Code         string `json:"code"`
			State        string `json:"state"`
			CodeVerifier string `json:"code_verifier"`
			RedirectURI  string `json:"redirect_uri"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if p.Provider == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "missing provider"}
		}
		if p.Code == "" || p.State == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "missing code or state"}
		}
		if p.RedirectURI == "" {
			p.RedirectURI = OAuthRedirectURI
		}
		// The daemon already stores state -> (verifier, provider) at
		// GenerateAuthURL time, so we look up the verifier from the
		// state. We do NOT loop over providers — that would let a
		// Google code "successfully" redeem as a GitHub session
		// because the GitHub token endpoint would reject it but the
		// loop's lastErr logic would still produce a result.
		sess, err := subs.Account.ExchangeCode(ctx, p.Provider, p.Code, p.State, p.RedirectURI)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		_ = subs.Audit.Append(ctx, buildAuditEvent(
			"account.oauth_callback",
			appCondurad,
			auditResultAllow,
			"email="+sess.Email+" provider="+sess.Provider,
		))
		return signedInStatus(sess, configuredProviders(subs)), nil
	}
}

func accountMagicLinkHandler(subs *Subsystems) ipc.HandlerFunc {
	return func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Email       string `json:"email"`
			Locale      string `json:"locale"`
			RedirectURL string `json:"redirect_url"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if p.Locale == "" {
			p.Locale = "en"
		}
		// The redirect_url is used by the web-side /api/auth/magic
		// route to construct the link emailed to the user. For
		// desktop sign-in the magic-link callback is handled
		// separately (the user opens the link in a browser which
		// redirects to the web verify page, then sets a cookie that
		// the desktop polls). Pass-through is fine.
		if err := subs.Account.RequestMagicLink(ctx, p.Email); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return map[string]any{"sent": true, "expires_in": 300, "dev_token": ""}, nil
	}
}

func accountLogoutHandler(subs *Subsystems) ipc.HandlerFunc {
	return func(ctx context.Context, _ json.RawMessage) (any, error) {
		if err := subs.Account.SignOut(ctx); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		_ = subs.Audit.Append(ctx, buildAuditEvent("account.logout", appCondurad, auditResultAllow, ""))
		return map[string]any{"ok": true}, nil
	}
}

// oauthURLError maps an error from GenerateAuthURL into the right IPC
// code + message. ErrProviderNotConfigured is a user-actionable configuration
// problem (CodeInvalidParams) — anything else is internal.
func oauthURLError(provider string, err error) *ipc.Error {
	if errors.Is(err, account.ErrProviderNotConfigured) {
		return &ipc.Error{
			Code: ipc.CodeInvalidParams,
			Message: fmt.Sprintf(
				"OAuth provider %q is not configured (set CONDURA_ACCOUNT_OAUTH_%s_CLIENT_ID or account.oauth.%s.client_id in config.yaml)",
				provider, strings.ToUpper(provider), provider,
			),
		}
	}
	return &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
}

// signedInStatus / signedOutStatus build the response shapes for
// account.status. Centralized so the two arms can't drift.
func signedInStatus(sess *account.Session, providers []string) map[string]any {
	return map[string]any{
		"signed_in":    true,
		"email":        sess.Email,
		"provider":     sess.Provider,
		"avatar_url":   sess.AvatarURL,
		"display_name": displayNameFor(sess),
		// tier lives on the server-side hub; not yet integrated.
		"tier":       "",
		"expires_at": sess.ExpiresAt.UTC().Format(time.RFC3339),
		"providers":  providers,
	}
}

func signedOutStatus(providers []string) map[string]any {
	return map[string]any{
		"signed_in":    false,
		"email":        nil,
		"provider":     nil,
		"avatar_url":   nil,
		"display_name": nil,
		"tier":         nil,
		"expires_at":   nil,
		"providers":    providers,
	}
}

func registerAccountNotAvailable(srv *ipc.Server) {
	na := func(_ context.Context, _ json.RawMessage) (any, error) {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "account subsystem not available"}
	}
	srv.Register("account.status", na)
	srv.Register("account.providers", na)
	srv.Register("account.oauth_url", na)
	srv.Register("account.oauth_callback", na)
	srv.Register("account.magic_link", na)
	srv.Register("account.logout", na)
}

// configuredProviders returns the list of OAuth providers that have a
// ClientID configured. The GUI uses this to decide which "Sign in with X"
// buttons to show. When none are configured, the GUI shows an empty list
// rather than failing at button-click time.
func configuredProviders(subs *Subsystems) []string {
	if subs == nil || subs.Account == nil {
		return nil
	}
	r := subs.Account.Providers()
	if r == nil {
		return nil
	}
	return r.Configured()
}

// displayNameFor returns a human-readable display name. The Session struct
// doesn't carry a display name; fall back to the email's local part.
func displayNameFor(sess *account.Session) string {
	if sess == nil {
		return ""
	}
	at := strings.IndexByte(sess.Email, '@')
	if at < 0 {
		return sess.Email
	}
	return sess.Email[:at]
}
