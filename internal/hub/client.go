package hub

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client talks to the Skills Hub (hub.synaptic.app). It fetches
// skill metadata, downloads skill archives, and publishes skills.
// All content is untrusted until verified; the scan package runs
// safety checks on downloaded artifacts.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient returns a hub client pointing at the given base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

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

// Search queries the hub for skills matching the query string.
func (c *Client) Search(query string, limit int) (*SearchResult, error) {
	u := fmt.Sprintf("%s/api/v1/skills/search?q=%s&limit=%d",
		c.baseURL, url.QueryEscape(query), limit)
	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("hub search: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
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
	u := fmt.Sprintf("%s/api/v1/skills/%s", c.baseURL, url.PathEscape(id))
	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("hub get: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("hub skill %q not found", id)
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

// Download fetches the skill archive and returns its bytes and SHA-256
// checksum. The caller should pass the result through scan.Verify
// before installing.
func (c *Client) Download(id string) ([]byte, string, error) {
	u := fmt.Sprintf("%s/api/v1/skills/%s/download", c.baseURL, url.PathEscape(id))
	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, "", fmt.Errorf("hub download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("hub download: status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("hub download read: %w", err)
	}
	sum := sha256.Sum256(data)
	return data, fmt.Sprintf("%x", sum), nil
}

// Publish uploads a skill archive to the hub. The archive must be
// signed by the author's Ed25519 key (verified server-side).
func (c *Client) Publish(archive []byte, meta SkillMeta) error {
	u := fmt.Sprintf("%s/api/v1/skills/publish", c.baseURL)
	req, err := http.NewRequest(http.MethodPost, u, nil)
	if err != nil {
		return fmt.Errorf("hub publish: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	body, err := json.Marshal(map[string]any{
		"archive": archive,
		"meta":    meta,
	})
	if err != nil {
		return fmt.Errorf("hub publish marshal: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("hub publish: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("hub publish: status %d", resp.StatusCode)
	}
	return nil
}
