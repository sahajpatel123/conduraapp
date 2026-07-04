// Package hub talks to the Skills Hub (hub.condura.app).
//
// Two modes:
//
//  1. **Network client** — `Client` fetches metadata, downloads
//     skill archives, and publishes skills to the central Hub.
//     Authentication is a bearer token. Publish is client-side
//     Ed25519 signed so a server compromise can't forge skills.
//
//  2. **Local server** — `Server` is a small in-process hub that
//     serves skills from a local directory tree. This is the
//     "offline" mode for users who don't want to depend on a
//     remote hub. Skills are still safety-scanned on install.
//
// The Hub is opt-in. The default is `enabled: false` in config
// (see internal/config/config.go). Setting `enabled: true` and
// pointing `base_url` at a remote hub is the user's choice.
package hub

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client talks to the Skills Hub (hub.condura.app by default).
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string // optional bearer token (from config or login)
	// publishKey is the author's Ed25519 keypair used to sign
	// Publish requests. Optional — when nil, Publish sends the
	// archive unsigned (the server may reject it, depending on
	// the deployment's policy).
	publishKey ed25519.PrivateKey
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithToken sets a bearer token for authenticated requests.
func WithToken(token string) ClientOption {
	return func(c *Client) { c.token = token }
}

// WithPublishKey sets the Ed25519 private key used to sign Publish
// requests. The corresponding public key is sent in the publish
// payload so the server can verify the signature.
func WithPublishKey(priv ed25519.PrivateKey) ClientOption {
	return func(c *Client) { c.publishKey = priv }
}

// NewClient returns a hub client pointing at the given base URL.
func NewClient(baseURL string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SetToken updates the bearer token at runtime (e.g., after login).
func (c *Client) SetToken(token string) { c.token = token }

// SkillMeta is the hub's representation of a skill. It carries
// provenance that the local store doesn't have.
type SkillMeta struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	License     string   `json:"license"`
	Tags        []string `json:"tags"`
	Trust       string   `json:"trust"`
	Checksum    string   `json:"checksum"` // SHA-256 of the archive
	Downloads   int      `json:"downloads"`
	PublishedAt string   `json:"published_at"` // RFC 3339
	UpdatedAt   string   `json:"updated_at"`   // RFC 3339
}

// SearchResult is a single search result from the hub.
type SearchResult struct {
	Skills []SkillMeta `json:"skills"`
	Total  int         `json:"total"`
	Query  string      `json:"query"`
}

// doGet performs an authenticated GET request.
func (c *Client) doGet(path string, params url.Values) (*http.Response, error) {
	u := c.baseURL + path
	if params != nil {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	c.applyAuth(req)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "condurad/0.1.0")
	return c.httpClient.Do(req)
}

// doPost performs an authenticated POST request with a JSON body.
func (c *Client) doPost(path string, body any) (*http.Response, error) {
	u := c.baseURL + path
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	c.applyAuth(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "condurad/0.1.0")
	return c.httpClient.Do(req)
}

func (c *Client) applyAuth(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

// Search queries the hub for skills matching the query string.
func (c *Client) Search(query string, limit int) (*SearchResult, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	params := url.Values{"q": {query}, "limit": {fmt.Sprintf("%d", limit)}}
	resp, err := c.doGet("/api/v1/skills/search", params)
	if err != nil {
		return nil, fmt.Errorf("hub search: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("hub search: authentication required (set a token in config)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hub search: status %d", resp.StatusCode)
	}
	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("hub search decode: %w", err)
	}
	return &result, nil
}

// Get fetches skill metadata by ID.
func (c *Client) Get(id string) (*SkillMeta, error) {
	resp, err := c.doGet("/api/v1/skills/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, fmt.Errorf("hub get: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("hub skill %q not found", id)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("hub get: authentication required")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hub get: status %d", resp.StatusCode)
	}
	var meta SkillMeta
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, fmt.Errorf("hub get decode: %w", err)
	}
	return &meta, nil
}

// maxArchiveSize is the largest skill archive Download will
// accept. Synaptic skills are typically 1-50 KB; a 32 MB cap
// is a generous safety margin that prevents a malicious or
// buggy hub from filling the daemon's memory with a 4 GB zip
// bomb. 32 MB also matches the encrypted-frame size cap used
// by the P2P sync engine.
const maxArchiveSize = 32 * 1024 * 1024

// Download fetches the skill archive and returns its bytes and SHA-256
// checksum. The caller should pass the result through scan.Verify
// before installing. The download is capped at maxArchiveSize
// (32 MB) to defend against zip-bomb DoS.
func (c *Client) Download(id string) ([]byte, string, error) {
	resp, err := c.doGet("/api/v1/skills/"+url.PathEscape(id)+"/download", nil)
	if err != nil {
		return nil, "", fmt.Errorf("hub download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("hub download: status %d", resp.StatusCode)
	}
	// Pre-check Content-Length when the server reports it. This
	// gives a clean error before we allocate a large buffer.
	if resp.ContentLength > maxArchiveSize {
		return nil, "", fmt.Errorf("hub download: too large (%d bytes, cap %d)", resp.ContentLength, maxArchiveSize)
	}
	// Use a LimitReader so a server that lies about Content-Length
	// still can't OOM us. Read exactly one byte past the cap to
	// detect the overflow.
	limited := io.LimitReader(resp.Body, maxArchiveSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, "", fmt.Errorf("hub download read: %w", err)
	}
	if len(data) > maxArchiveSize {
		return nil, "", fmt.Errorf("hub download: too large (exceeded %d bytes)", maxArchiveSize)
	}
	sum := sha256.Sum256(data)
	return data, fmt.Sprintf("%x", sum), nil
}

// Publish uploads a skill archive to the hub. When a publish key
// is configured (via WithPublishKey), the request is signed:
//
//	payload = hex(sha256(archive)) || meta.id || meta.version
//	signature = ed25519.sign(publishKey, payload)
//
// The server is expected to verify the signature against the
// author's known public key. This prevents a hub-compromise
// attacker from uploading skills under someone else's name.
func (c *Client) Publish(archive []byte, meta SkillMeta) error {
	body := map[string]any{
		"archive": archive,
		"meta":    meta,
	}
	if c.publishKey != nil {
		sum := sha256.Sum256(archive)
		payload := hex.EncodeToString(sum[:]) + "|" + meta.ID + "|" + meta.Version
		sig := ed25519.Sign(c.publishKey, []byte(payload))
		body["author_pubkey"] = hex.EncodeToString(c.publishKey.Public().(ed25519.PublicKey))
		body["signature"] = hex.EncodeToString(sig)
	}
	resp, err := c.doPost("/api/v1/skills/publish", body)
	if err != nil {
		return fmt.Errorf("hub publish: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("hub publish: authentication required")
	}
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("hub publish: signature invalid or author not registered")
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("hub publish: status %d", resp.StatusCode)
	}
	return nil
}
