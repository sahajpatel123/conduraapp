package account

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DefaultMagicLinkURL is the Condura server endpoint that issues magic-link emails.
// Resolved by SetMagicLinkURL on startup; this constant is the fallback for
// development and tests.
const DefaultMagicLinkURL = "https://condura.app/api/auth/magic"

// DefaultMagicVerifyURL is the Condura server endpoint that verifies one-time
// magic-link tokens. Resolved by SetMagicLinkURL on startup; this constant is
// the fallback for development and tests.
const DefaultMagicVerifyURL = "https://condura.app/api/auth/verify"

var (
	magicLinkURL   = DefaultMagicLinkURL
	magicVerifyURL = DefaultMagicVerifyURL
)

// SetMagicLinkURL overrides the magic-link endpoint URLs at runtime. Called
// from buildAccount using the user's account.magic_url config value when set.
// Resets to defaults when called with empty strings.
func SetMagicLinkURL(issue, verify string) {
	if issue == "" {
		magicLinkURL = DefaultMagicLinkURL
	} else {
		magicLinkURL = issue
	}
	if verify == "" {
		magicVerifyURL = DefaultMagicVerifyURL
	} else {
		magicVerifyURL = verify
	}
}

// RequestMagicLink sends a one-time sign-in link to the user's email.
func (m *Manager) RequestMagicLink(ctx context.Context, email string) error {
	if !validEmail(email) {
		return fmt.Errorf("account: invalid email %q", email)
	}
	body := fmt.Sprintf(`{"email":%q}`, email)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, magicLinkURL,
		strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("account: magic link request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("account: magic link server: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("account: magic link server returned %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// VerifyMagicToken verifies a one-time token from the magic link.
// On success, creates a local session.
func (m *Manager) VerifyMagicToken(ctx context.Context, token string) (*Session, error) {
	if token == "" {
		return nil, fmt.Errorf("account: empty token")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		magicVerifyURL+"?token="+token, nil)
	if err != nil {
		return nil, fmt.Errorf("account: verify request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("account: verify server: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("account: token invalid or expired")
	}
	var result struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("account: parse verify response: %w", err)
	}
	if result.Email == "" {
		return nil, fmt.Errorf("account: empty email in verify response")
	}
	return m.NewSession(ctx, result.Email, "magic_link")
}
