package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/account"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
)

func registerAccountMethods(srv *ipc.Server, subs *Subsystems) {
	if subs.Account == nil {
		registerAccountNotAvailable(srv)
		return
	}

	srv.Register("account.status", func(ctx context.Context, _ json.RawMessage) (any, error) {
		sess, err := subs.Account.Status(ctx)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		if sess == nil {
			return map[string]any{"signed_in": false, "email": nil, "provider": nil, "avatar_url": nil, "expires_at": nil}, nil
		}
		return map[string]any{
			"signed_in":  true,
			"email":      sess.Email,
			"provider":   sess.Provider,
			"avatar_url": sess.AvatarURL,
			"expires_at": sess.ExpiresAt.UTC().Format(time.RFC3339),
		}, nil
	})

	srv.Register("account.oauth_url", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Provider string `json:"provider"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		url, state, err := subs.Account.GenerateAuthURL(p.Provider, "synaptic://auth/callback")
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return map[string]any{"url": url, "state": state}, nil
	})

	srv.Register("account.oauth_callback", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Code  string `json:"code"`
			State string `json:"state"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		// Determine provider from stored state (done inside ExchangeCode).
		// We use a reasonable default based on what's available.
		providers := []string{providerGoogle, "github"}
		var sess *account.Session
		var lastErr error
		for _, prov := range providers {
			sess, lastErr = subs.Account.ExchangeCode(ctx, prov, p.Code, p.State, "synaptic://auth/callback")
			if lastErr == nil {
				break
			}
		}
		if sess == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: fmt.Sprintf("token exchange failed: %v", lastErr)}
		}
		_ = subs.Audit.Append(ctx, buildAuditEvent("account.oauth_callback", appSynapticd, auditResultAllow, "email="+sess.Email+" provider="+sess.Provider))
		return map[string]any{
			"signed_in":  true,
			"email":      sess.Email,
			"provider":   sess.Provider,
			"expires_at": sess.ExpiresAt.UTC().Format(time.RFC3339),
		}, nil
	})

	srv.Register("account.magic_link", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Email string `json:"email"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if err := subs.Account.RequestMagicLink(ctx, p.Email); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return map[string]any{"sent": true}, nil
	})

	srv.Register("account.logout", func(ctx context.Context, _ json.RawMessage) (any, error) {
		if err := subs.Account.SignOut(ctx); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		_ = subs.Audit.Append(ctx, buildAuditEvent("account.logout", appSynapticd, auditResultAllow, ""))
		return map[string]any{"ok": true}, nil
	})
}

func registerAccountNotAvailable(srv *ipc.Server) {
	na := func(_ context.Context, _ json.RawMessage) (any, error) {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "account subsystem not available"}
	}
	srv.Register("account.status", na)
	srv.Register("account.oauth_url", na)
	srv.Register("account.oauth_callback", na)
	srv.Register("account.magic_link", na)
	srv.Register("account.logout", na)
}
