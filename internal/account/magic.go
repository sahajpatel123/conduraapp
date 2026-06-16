package account

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// MagicLinkURL is the Synaptic server endpoint for magic-link auth.
const MagicLinkURL = "https://synaptic.app/api/auth/magic"

// MagicVerifyURL is the Synaptic server endpoint for verifying
// one-time magic-link tokens.
const MagicVerifyURL = "https://synaptic.app/api/auth/verify"

// RequestMagicLink sends a one-time sign-in link to the user's email.
func (m *Manager) RequestMagicLink(ctx context.Context, email string) error {
	if !validEmail(email) {
		return fmt.Errorf("account: invalid email %q", email)
	}
	body := fmt.Sprintf(`{"email":%q}`, email)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, MagicLinkURL,
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
		MagicVerifyURL+"?token="+token, nil)
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
