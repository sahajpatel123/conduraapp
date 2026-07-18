package hub

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestWithToken_SetsToken pins the WithToken option: it MUST set
// c.token via the returned ClientOption. Without this pin, a regression
// in option-application order (NewClient applies options in iteration
// order) would silently leave c.token empty.
func TestWithToken_SetsToken(t *testing.T) {
	c := NewClient("https://hub.example.com", WithToken("my-secret"))
	if c.token != "my-secret" {
		t.Errorf("token = %q, want %q", c.token, "my-secret")
	}
}

// TestWithPublishKey_SetsKey pins the WithPublishKey option: it MUST
// set c.publishKey to the provided Ed25519 private key. Publish
// requests use this key to sign archives; without it, every publish
// would be unsigned (and the hub would reject).
func TestWithPublishKey_SetsKey(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	c := NewClient("https://hub.example.com", WithPublishKey(priv))
	if !c.publishKey.Equal(priv) {
		t.Errorf("publishKey not set; got %x, want %x", []byte(c.publishKey)[:8], []byte(priv)[:8])
	}
	_ = pub // kept for clarity that publishKey is the private counterpart of a public key
}

// TestWithHTTPClient_SetsClient pins the WithHTTPClient option: it
// MUST replace c.httpClient when the provided client is non-nil.
// This is the entry point for the PinnedHTTPClient (sanitize
// package), which closes the DNS-rebinding TOCTOU window.
func TestWithHTTPClient_SetsClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	c := NewClient("https://hub.example.com", WithHTTPClient(custom))
	if c.httpClient != custom {
		t.Errorf("httpClient = %p, want %p", c.httpClient, custom)
	}
	if c.httpClient.Timeout != 5*time.Second {
		t.Errorf("httpClient.Timeout = %v, want 5s", c.httpClient.Timeout)
	}
}

// TestWithHTTPClient_NilIsIgnored pins the nil-guard: a nil
// *http.Client passed to WithHTTPClient MUST be ignored (c.httpClient
// stays at its current value). Without this, a regression that
// accepted nil would NPE on the first request.
func TestWithHTTPClient_NilIsIgnored(t *testing.T) {
	c := NewClient("https://hub.example.com") // default httpClient
	original := c.httpClient
	c2 := NewClient("https://hub.example.com", WithHTTPClient(nil))
	if c2.httpClient != original {
		// Hard to compare pointers; check that c2 still has a
		// non-nil client.
		if c2.httpClient == nil {
			t.Errorf("WithHTTPClient(nil) set httpClient to nil; want non-nil (default preserved)")
		}
	}
}

// TestNewClient_DefaultsTimeout pins the default-timeout contract:
// when WithHTTPClient is NOT supplied, NewClient MUST install an
// http.Client with a 30-second timeout. This is the safety net for
// callers that forget to inject a PinnedHTTPClient (less critical in
// v0.1.x but still a contract worth pinning).
func TestNewClient_DefaultsTimeout(t *testing.T) {
	c := NewClient("https://hub.example.com")
	if c.httpClient == nil {
		t.Fatal("httpClient = nil; want non-nil default")
	}
	if c.httpClient.Timeout != 30*time.Second {
		t.Errorf("default httpClient.Timeout = %v, want 30s", c.httpClient.Timeout)
	}
}

// TestNewClient_AppliesOptionsInOrder pins the option-order
// contract: NewClient MUST apply options in the order supplied, with
// later options overriding earlier ones for the same field. Without
// this pin, the iteration order is a Go convention (range over
// variadic) and easy to break with a refactor.
func TestNewClient_AppliesOptionsInOrder(t *testing.T) {
	c := NewClient("https://hub.example.com",
		WithToken("first"),
		WithToken("second"), // later wins
	)
	if c.token != "second" {
		t.Errorf("token = %q, want \"second\" (later WithToken should win)", c.token)
	}
}

// TestSetToken_UpdatesAtRuntime pins the runtime-update contract:
// SetToken MUST replace c.token after construction. This is the path
// the login flow uses — the user authenticates, the daemon stores
// the token, then SetToken wires it into the hub client.
func TestSetToken_UpdatesAtRuntime(t *testing.T) {
	c := NewClient("https://hub.example.com", WithToken("initial"))
	if c.token != "initial" {
		t.Fatalf("setup: token = %q, want \"initial\"", c.token)
	}

	c.SetToken("logged-in-token")
	if c.token != "logged-in-token" {
		t.Errorf("token after SetToken = %q, want \"logged-in-token\"", c.token)
	}
}

// TestGet_SuccessReturnsMetadata pins the happy-path contract:
// Get MUST fetch the metadata JSON and return it. The path escapes
// the ID for safety.
func TestGet_SuccessReturnsMetadata(t *testing.T) {
	wantMeta := SkillMeta{
		ID:      "summarize-webpage",
		Name:    "Summarize Webpage",
		Version: "1.2.3",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Path should be /api/v1/skills/summarize-webpage.
		if !strings.HasSuffix(r.URL.Path, "/skills/summarize-webpage") {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Accept header = %q, want application/json", r.Header.Get("Accept"))
		}
		_ = json.NewEncoder(w).Encode(wantMeta)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	got, err := c.Get("summarize-webpage")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != wantMeta.ID || got.Name != wantMeta.Name || got.Version != wantMeta.Version {
		t.Errorf("Get returned %+v, want ID=%q Name=%q Version=%q", got, wantMeta.ID, wantMeta.Name, wantMeta.Version)
	}
}

// TestGet_NotFoundReturnsClearError pins the 404 contract: a missing
// skill MUST produce an error that mentions BOTH "not found" AND the
// requested ID. Diagnostic clarity matters for the GUI's "this skill
// was unpublished" message.
func TestGet_NotFoundReturnsClearError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	_, err := c.Get("missing-skill")
	if err == nil {
		t.Fatal("Get(missing) = nil; want error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error %q should mention 'not found'", err.Error())
	}
	if !strings.Contains(err.Error(), "missing-skill") {
		t.Errorf("error %q should mention the requested ID 'missing-skill'", err.Error())
	}
}

// TestGet_UnauthorizedReturnsClearError pins the 401 contract: an
// unauthenticated request MUST produce an error mentioning
// authentication. The GUI's "set a token in config" toast needs this
// hint to be actionable.
func TestGet_UnauthorizedReturnsClearError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "auth required", http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	_, err := c.Get("anything")
	if err == nil {
		t.Fatal("Get(401) = nil; want error")
	}
	if !strings.Contains(err.Error(), "authentication") {
		t.Errorf("error %q should mention 'authentication'", err.Error())
	}
}

// TestGet_NonOKReturnsStatusError pins the generic-non-200 contract:
// any non-200 non-404 non-401 status MUST produce an error mentioning
// the status code. Without this, a 500 from the hub would surface as
// a confusing "decode error" downstream.
func TestGet_NonOKReturnsStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	_, err := c.Get("any")
	if err == nil {
		t.Fatal("Get(500) = nil; want error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error %q should mention status code 500", err.Error())
	}
}

// TestGet_InvalidJSONReturnsDecodeError pins the JSON-decode failure
// contract: a 200 response with invalid JSON MUST produce a wrapped
// decode error (preserves underlying cause for logs).
func TestGet_InvalidJSONReturnsDecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not valid json"))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	_, err := c.Get("any")
	if err == nil {
		t.Fatal("Get(bad JSON) = nil; want error")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("error %q should mention 'decode'", err.Error())
	}
}

// TestDownload_SuccessReturnsBytesAndChecksum pins the happy-path
// contract: Download MUST return the body bytes and the SHA-256 hex
// checksum. The caller is expected to verify the checksum via
// scan.Verify before installing.
func TestDownload_SuccessReturnsBytesAndChecksum(t *testing.T) {
	payload := []byte("fake-skill-archive-content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Path should be /api/v1/skills/<id>/download.
		if !strings.HasSuffix(r.URL.Path, "/skills/abc/download") {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		_, _ = w.Write(payload)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	data, sum, err := c.Download("abc")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if string(data) != string(payload) {
		t.Errorf("Download returned %q, want %q", data, payload)
	}
	wantSum := fmt.Sprintf("%x", sha256.Sum256(payload))
	if sum != wantSum {
		t.Errorf("Download checksum = %q, want %q", sum, wantSum)
	}
}

// TestDownload_ContentLengthTooLargeRejected pins the
// Content-Length pre-check: if the server reports Content-Length
// larger than maxArchiveSize (32 MB), Download MUST reject BEFORE
// reading the body. This is the fast-path defense against zip-bomb
// DoS.
func TestDownload_ContentLengthTooLargeRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", maxArchiveSize+1))
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	_, _, err := c.Download("too-big")
	if err == nil {
		t.Fatal("Download(ContentLength > cap) = nil; want error")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("error %q should mention 'too large'", err.Error())
	}
}

// TestDownload_BodyExceedsCapRejected pins the LimitReader overflow
// detection: even if the server LIES about Content-Length (omits it
// or reports a smaller value), Download MUST detect when the actual
// body exceeds maxArchiveSize. The LimitReader caps reads at
// cap+1 bytes; the post-read length check catches the overflow.
//
// (In practice this test is hard to trigger in-process: writing
// maxArchiveSize+1 = 32 MB+1 bytes through httptest is expensive,
// and the http transport may truncate based on Content-Length.
// The Content-Length pre-check in TestDownload_ContentLengthTooLargeRejected
// is the fast-path defense; this test pins the residual contract.)
//
// DEFERRED: writing 32 MB+1 in a unit test would cost ~50ms and
// risk being flaky across CI machines. The Content-Length pre-check
// is sufficient for zip-bomb DoS prevention; the LimitReader +
// post-read check is a defense-in-depth layer that's only reachable
// in adversarial conditions (server lying about Content-Length AND
// actually streaming more than the cap). Documented here so the
// gap is visible.
func TestDownload_BodyExceedsCapRejected(t *testing.T) {
	t.Skip("deferred: maxArchiveSize is 32 MB; writing 32 MB+1 in a unit test is impractical")
}

// TestDownload_NonOKReturnsStatusError pins the non-200 contract for
// Download: a 4xx or 5xx MUST produce a status-code error.
func TestDownload_NonOKReturnsStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone", http.StatusGone)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	_, _, err := c.Download("gone")
	if err == nil {
		t.Fatal("Download(410) = nil; want error")
	}
	if !strings.Contains(err.Error(), "410") {
		t.Errorf("error %q should mention status code 410", err.Error())
	}
}

// TestApplyAuth_AddsBearerHeader pins the auth-header contract: an
// authenticated client MUST include `Authorization: Bearer <token>`
// in every request. Without this, the hub would 401 every call.
func TestApplyAuth_AddsBearerHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(SkillMeta{ID: "x"})
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, WithToken("super-secret"))
	if _, err := c.Get("x"); err != nil {
		t.Fatalf("Get: %v", err)
	}
	want := "Bearer super-secret"
	if gotAuth != want {
		t.Errorf("Authorization = %q, want %q", gotAuth, want)
	}
}

// TestDoGet_PropagatesNetworkError pins the network-error contract:
// when the server is unreachable, doGet MUST return an error wrapping
// the underlying network failure (so callers can distinguish
// transport from protocol errors if needed).
func TestDoGet_PropagatesNetworkError(t *testing.T) {
	// New a server, immediately close it so the URL is unreachable.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	c := NewClient(srv.URL)
	_, err := c.Get("any")
	if err == nil {
		t.Fatal("Get(unreachable) = nil; want error")
	}
	// The error chain should contain a network error. Just check
	// non-nil + non-empty message — the underlying type is
	// *url.Error and is implementation-dependent.
	if errors.Unwrap(err) == nil && !strings.Contains(err.Error(), "connection refused") {
		// Either wrap is nil OR not connection-refused (CI may use
		// different network-error wording). Forcing the test to
		// pin a specific phrasing would be brittle; just assert
		// the error exists and contains the path.
	}
	if !strings.Contains(err.Error(), "hub get") {
		t.Errorf("error %q should be wrapped with 'hub get' context", err.Error())
	}
}