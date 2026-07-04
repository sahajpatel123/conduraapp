package hub

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestServer_LocalAddSearch tests the full local server flow.
func TestServer_LocalAddSearch(t *testing.T) {
	root := t.TempDir()
	srv, err := NewServer(root, "")
	if err != nil {
		t.Fatal(err)
	}
	meta := SkillMeta{
		ID:          "weather-lookup",
		Name:        "Weather Lookup",
		Description: "Fetches weather from a public API",
		Version:     "1.0.0",
		Author:      "tester",
		License:     "MIT",
		Tags:        []string{"weather", "api"},
		Trust:       "official",
	}
	archive := []byte("PK\x03\x04fake zip body")
	if err := srv.LocalAdd(meta, archive); err != nil {
		t.Fatal(err)
	}
	if srv.Count() != 1 {
		t.Errorf("Count: got %d, want 1", srv.Count())
	}

	// Spin up an httptest server.
	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()

	// Search
	resp, err := http.Get(httpSrv.URL + "/api/v1/skills/search?q=weather")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("search status: %d", resp.StatusCode)
	}
	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 || len(result.Skills) != 1 {
		t.Errorf("search result: %+v", result)
	}
	if result.Skills[0].ID != "weather-lookup" {
		t.Errorf("skill id: %s", result.Skills[0].ID)
	}
	if result.Skills[0].Checksum == "" {
		t.Error("checksum not populated from archive")
	}

	// Get
	resp2, err := http.Get(httpSrv.URL + "/api/v1/skills/weather-lookup")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Fatalf("get status: %d", resp2.StatusCode)
	}

	// Download
	resp3, err := http.Get(httpSrv.URL + "/api/v1/skills/weather-lookup/download")
	if err != nil {
		t.Fatal(err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != 200 {
		t.Fatalf("download status: %d", resp3.StatusCode)
	}
	if ct := resp3.Header.Get("Content-Type"); ct != "application/zip" {
		t.Errorf("content-type: %s", ct)
	}
}

// TestServer_AuthToken verifies bearer token enforcement.
func TestServer_AuthToken(t *testing.T) {
	root := t.TempDir()
	srv, err := NewServer(root, "secret-token-123")
	if err != nil {
		t.Fatal(err)
	}
	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()

	// No token -> 401
	resp, err := http.Get(httpSrv.URL + "/api/v1/skills/search?q=foo")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 401 {
		t.Errorf("no-token request: got %d, want 401", resp.StatusCode)
	}

	// Wrong token -> 401
	req, _ := http.NewRequest("GET", httpSrv.URL+"/api/v1/skills/search?q=foo", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != 401 {
		t.Errorf("wrong-token request: got %d, want 401", resp2.StatusCode)
	}

	// Right token -> 200
	req2, _ := http.NewRequest("GET", httpSrv.URL+"/api/v1/skills/search?q=foo", nil)
	req2.Header.Set("Authorization", "Bearer secret-token-123")
	resp3, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	resp3.Body.Close()
	if resp3.StatusCode != 200 {
		t.Errorf("right-token request: got %d, want 200", resp3.StatusCode)
	}
}

// TestClient_AuthHeader ensures the client sends the bearer token.
func TestClient_AuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"skills":[],"total":0,"query":""}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, WithToken("my-secret"))
	_, err := c.Search("foo", 10)
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer my-secret" {
		t.Errorf("auth header: got %q, want %q", gotAuth, "Bearer my-secret")
	}
}

// TestClient_Publish_NoSignByDefault ensures Publish does NOT send
// a signature when no publish key is configured.
func TestClient_Publish_NoSignByDefault(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(201)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	meta := SkillMeta{ID: "x", Version: "1"}
	if err := c.Publish([]byte("archive"), meta); err != nil {
		t.Fatal(err)
	}
	if _, ok := gotBody["signature"]; ok {
		t.Error("signature should be absent when no publish key set")
	}
	if _, ok := gotBody["author_pubkey"]; ok {
		t.Error("author_pubkey should be absent when no publish key set")
	}
}

// TestServer_SearchEscape verifies URL escaping in search.
func TestServer_SearchEscape(t *testing.T) {
	root := t.TempDir()
	srv, err := NewServer(root, "")
	if err != nil {
		t.Fatal(err)
	}
	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()
	u := httpSrv.URL + "/api/v1/skills/search?" + url.Values{"q": {"a/b/c"}}.Encode()
	resp, err := http.Get(u)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("search with /: status %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Request.URL.RawQuery, "a%2Fb%2Fc") {
		t.Errorf("raw query: %s", resp.Request.URL.RawQuery)
	}
}

// TestServer_LocalAdd_PathTraversalRejected ensures skill IDs that
// could escape the skills/ directory are rejected.
func TestServer_LocalAdd_PathTraversalRejected(t *testing.T) {
	root := t.TempDir()
	srv, err := NewServer(root, "")
	if err != nil {
		t.Fatal(err)
	}
	bad := []string{
		"../etc/passwd",
		"..",
		"../foo",
		"foo/bar",
		"foo\\bar",
		"foo$bar",
		"",
		strings.Repeat("a", 200), // too long
	}
	for _, id := range bad {
		err := srv.LocalAdd(SkillMeta{ID: id, Name: id}, []byte("x"))
		if err == nil {
			t.Errorf("LocalAdd accepted bad ID %q", id)
		}
	}
}

// TestServer_LocalAdd_ValidIDsAccepted ensures common skill ID
// formats (kebab-case, dot-namespaced, versioned) are accepted.
func TestServer_LocalAdd_ValidIDsAccepted(t *testing.T) {
	root := t.TempDir()
	srv, err := NewServer(root, "")
	if err != nil {
		t.Fatal(err)
	}
	good := []string{
		"weather-lookup",
		"weather.lookup.v1",
		"my_skill",
		"Skill1",
	}
	for _, id := range good {
		if err := srv.LocalAdd(SkillMeta{ID: id, Name: id}, []byte("x")); err != nil {
			t.Errorf("LocalAdd rejected good ID %q: %v", id, err)
		}
	}
	if srv.Count() != len(good) {
		t.Errorf("Count: got %d, want %d", srv.Count(), len(good))
	}
}
