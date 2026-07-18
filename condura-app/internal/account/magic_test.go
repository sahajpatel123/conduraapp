package account

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// withMagicURLs redirects the package-level magic-link URL globals to the
// given test server's URLs and restores the defaults via t.Cleanup. The
// package-level state MUST be reset on cleanup — otherwise a URL set in one
// test silently leaks into the next.
func withMagicURLs(t *testing.T, issueURL, verifyURL string) {
	t.Helper()
	SetMagicLinkURL(issueURL, verifyURL)
	t.Cleanup(func() {
		SetMagicLinkURL("", "") // empty branch resets to defaults per package contract
	})
}

// TestSetMagicLinkURL_EmptyResetsToDefaults pins the empty-string reset
// contract: SetMagicLinkURL("", "") MUST restore the package-level URLs to
// the DefaultMagicLinkURL / DefaultMagicVerifyURL constants. Without this,
// a configuration change in one test would silently leak into the next
// (the URL globals are package-level, not per-Manager).
func TestSetMagicLinkURL_EmptyResetsToDefaults(t *testing.T) {
	// First apply a non-default so we can verify the reset actually moves
	// the globals back.
	SetMagicLinkURL("https://example.com/issue", "https://example.com/verify")
	if magicLinkURL == DefaultMagicLinkURL {
		t.Fatalf("setup: magicLinkURL did not change after non-empty override")
	}

	SetMagicLinkURL("", "")

	if magicLinkURL != DefaultMagicLinkURL {
		t.Errorf("magicLinkURL after empty reset = %q, want %q", magicLinkURL, DefaultMagicLinkURL)
	}
	if magicVerifyURL != DefaultMagicVerifyURL {
		t.Errorf("magicVerifyURL after empty reset = %q, want %q", magicVerifyURL, DefaultMagicVerifyURL)
	}
}

// TestSetMagicLinkURL_NonEmptyOverrides pins the override contract: passing
// non-empty strings MUST replace the package-level URLs. This is the path
// buildAccount uses to apply the user's account.magic_url config value.
func TestSetMagicLinkURL_NonEmptyOverrides(t *testing.T) {
	const (
		newIssue  = "https://override.example.com/issue"
		newVerify = "https://override.example.com/verify"
	)
	t.Cleanup(func() { SetMagicLinkURL("", "") })

	SetMagicLinkURL(newIssue, newVerify)

	if magicLinkURL != newIssue {
		t.Errorf("magicLinkURL after override = %q, want %q", magicLinkURL, newIssue)
	}
	if magicVerifyURL != newVerify {
		t.Errorf("magicVerifyURL after override = %q, want %q", magicVerifyURL, newVerify)
	}
}

// TestRequestMagicLink_RejectsInvalidEmail pins the email-format guard:
// invalid email format MUST return an error WITHOUT making any HTTP call.
// A regression that removed this guard would either (a) send a request to
// the server with a malformed address (server-side rejection + wasted
// call), or (b) propagate the email into a downstream system that trusts
// our pre-validation.
func TestRequestMagicLink_RejectsInvalidEmail(t *testing.T) {
	m, _ := newTestManager(t)

	err := m.RequestMagicLink(context.Background(), "not-an-email")
	if err == nil {
		t.Fatal("RequestMagicLink(invalid email) = nil; want error")
	}
	if !strings.Contains(err.Error(), "invalid email") {
		t.Errorf("error = %q, want it to contain 'invalid email'", err.Error())
	}
}

// TestRequestMagicLink_SuccessPath200 pins the happy path: a valid email
// MUST POST a JSON {"email":"..."} body with Content-Type: application/json
// to the configured magicLinkURL, and return nil on HTTP 200.
func TestRequestMagicLink_SuccessPath200(t *testing.T) {
	var (
		gotMethod      string
		gotContentType string
		gotBody        []byte
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	withMagicURLs(t, srv.URL, srv.URL)

	m, _ := newTestManager(t)

	if err := m.RequestMagicLink(context.Background(), "alice@example.com"); err != nil {
		t.Fatalf("RequestMagicLink: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("server received method %q, want POST", gotMethod)
	}
	if !strings.HasPrefix(gotContentType, "application/json") {
		t.Errorf("server received Content-Type %q, want application/json prefix", gotContentType)
	}
	var parsed struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(gotBody, &parsed); err != nil {
		t.Fatalf("server received body %q, want valid JSON: %v", gotBody, err)
	}
	if parsed.Email != "alice@example.com" {
		t.Errorf("server received email %q, want %q", parsed.Email, "alice@example.com")
	}
}

// TestRequestMagicLink_Non200ReturnsStatusError pins the non-200 error
// path: a non-200 response MUST produce an error that includes both the
// status code and the response body (so the user/operator can see WHY the
// server rejected — rate-limit, bad endpoint, misconfig). Without the body
// in the error, support has no signal.
func TestRequestMagicLink_Non200ReturnsStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("rate limited; try again in 60s"))
	}))
	t.Cleanup(srv.Close)
	withMagicURLs(t, srv.URL, srv.URL)

	m, _ := newTestManager(t)

	err := m.RequestMagicLink(context.Background(), "alice@example.com")
	if err == nil {
		t.Fatal("RequestMagicLink on 429 = nil; want error")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Errorf("error %q should mention status code 429", err.Error())
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("error %q should include response body for diagnostics", err.Error())
	}
}

// TestVerifyMagicToken_Non200ReturnsInvalidError pins the failure branch
// when the verify endpoint returns non-200: the function MUST return a
// "token invalid or expired" error so the GUI can show a clean
// "link expired / invalid" message instead of leaking the underlying HTTP
// error (which would expose server-side details to the user).
func TestVerifyMagicToken_Non200ReturnsInvalidError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("token"); !strings.HasPrefix(got, "tok-") {
			t.Errorf("server received token %q, want tok-*", got)
		}
		w.WriteHeader(http.StatusGone) // 410 — token expired
	}))
	t.Cleanup(srv.Close)
	withMagicURLs(t, srv.URL, srv.URL)

	m, _ := newTestManager(t)

	_, err := m.VerifyMagicToken(context.Background(), "tok-expired-123")
	if err == nil {
		t.Fatal("VerifyMagicToken on 410 = nil; want error")
	}
	if !strings.Contains(err.Error(), "invalid or expired") {
		t.Errorf("error %q should mention 'invalid or expired'", err.Error())
	}
}

// TestVerifyMagicToken_InvalidJSONReturnsParseError pins the JSON-decode
// failure path: a 200 response whose body is not valid JSON MUST return a
// wrapped parse error (preserving the underlying cause for logs), rather
// than panicking or returning the empty-email error (which would mask the
// actual failure).
func TestVerifyMagicToken_InvalidJSONReturnsParseError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("this is not json"))
	}))
	t.Cleanup(srv.Close)
	withMagicURLs(t, srv.URL, srv.URL)

	m, _ := newTestManager(t)

	_, err := m.VerifyMagicToken(context.Background(), "tok-bad-json")
	if err == nil {
		t.Fatal("VerifyMagicToken on bad JSON = nil; want error")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("error %q should mention 'parse'", err.Error())
	}
}

// TestVerifyMagicToken_EmptyEmailInResponse pins the empty-email guard:
// a 200 response with `{"email":""}` MUST be rejected. Empty email would
// otherwise create a session bound to no real user — defense against a
// malicious or misconfigured server returning a "success" payload without
// identifying the user.
func TestVerifyMagicToken_EmptyEmailInResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"email":""}`))
	}))
	t.Cleanup(srv.Close)
	withMagicURLs(t, srv.URL, srv.URL)

	m, _ := newTestManager(t)

	_, err := m.VerifyMagicToken(context.Background(), "tok-empty-email")
	if err == nil {
		t.Fatal("VerifyMagicToken on empty email = nil; want error")
	}
	if !strings.Contains(err.Error(), "empty email") {
		t.Errorf("error %q should mention 'empty email'", err.Error())
	}
}

// TestVerifyMagicToken_SuccessCreatesSession pins the happy path: a 200
// response with a valid JSON email MUST result in a session whose Email
// matches the response AND whose Provider is "magic_link" (so the GUI can
// distinguish magic-link sessions from OAuth/email-password sessions in
// the audit chain and UI history).
func TestVerifyMagicToken_SuccessCreatesSession(t *testing.T) {
	const wantEmail = "alice@example.com"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("token"); got != "tok-valid" {
			t.Errorf("server received token %q, want tok-valid", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(struct {
			Email string `json:"email"`
		}{Email: wantEmail})
	}))
	t.Cleanup(srv.Close)
	withMagicURLs(t, srv.URL, srv.URL)

	m, _ := newTestManager(t)

	sess, err := m.VerifyMagicToken(context.Background(), "tok-valid")
	if err != nil {
		t.Fatalf("VerifyMagicToken on success: %v", err)
	}
	if sess == nil {
		t.Fatal("VerifyMagicToken returned nil session on success")
	}
	if sess.Email != wantEmail {
		t.Errorf("session.Email = %q, want %q", sess.Email, wantEmail)
	}
	if sess.Provider != "magic_link" {
		t.Errorf("session.Provider = %q, want magic_link", sess.Provider)
	}
}
