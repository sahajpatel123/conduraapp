package hub

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Server is a small in-process Skills Hub. It serves skills from
// a local directory tree, with no network calls. This is the
// "local-first" fallback for users who don't want to depend on
// hub.synaptic.app.
//
// Directory layout:
//
//	<root>/
//	  skills.db          # index.json (managed by Server)
//	  skills/
//	    <skill-id>/
//	      meta.json      # SkillMeta
//	      archive.zip    # the skill archive
//
// Server reads meta.json + archive.zip at startup and serves
// them via the same /api/v1/* endpoints as the network hub.
// Safety-scanning still happens on the client side via the
// scan package.
type Server struct {
	root    string
	mu      sync.RWMutex
	index   map[string]SkillMeta // by id
	archive map[string]string    // by id -> archive path
	token   string               // optional bearer token; "" = open
}

// NewServer creates a local Hub rooted at the given directory.
// The directory is created if it doesn't exist.
func NewServer(root string, token string) (*Server, error) {
	if err := os.MkdirAll(filepath.Join(root, "skills"), 0o755); err != nil {
		return nil, fmt.Errorf("hub server: mkdir skills: %w", err)
	}
	s := &Server{
		root:    root,
		index:   make(map[string]SkillMeta),
		archive: make(map[string]string),
		token:   token,
	}
	if err := s.reindex(); err != nil {
		return nil, err
	}
	return s, nil
}

// reindex walks the skills directory and rebuilds the in-memory
// index from meta.json + archive.zip files. Called on startup
// and after any LocalAdd.
func (s *Server) reindex() error {
	skillsDir := filepath.Join(s.root, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return fmt.Errorf("hub server: read skills dir: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		// Reject any directory whose name doesn't match the safe
		// character set. This prevents pre-existing path-traversal
		// skill dirs (e.g. "../etc") from being loaded.
		if !validSkillID(id) {
			continue
		}
		dir := filepath.Join(skillsDir, id)
		metaPath := filepath.Join(dir, "meta.json")
		archivePath := filepath.Join(dir, "archive.zip")
		metaData, err := os.ReadFile(metaPath)
		if err != nil {
			continue // no meta = not a skill dir
		}
		var meta SkillMeta
		if err := json.Unmarshal(metaData, &meta); err != nil {
			continue
		}
		// Verify the archive exists and compute its checksum.
		archiveData, err := os.ReadFile(archivePath)
		if err != nil {
			continue
		}
		sum := sha256.Sum256(archiveData)
		meta.Checksum = hex.EncodeToString(sum[:])
		if meta.ID == "" {
			meta.ID = id
		}
		s.index[meta.ID] = meta
		s.archive[meta.ID] = archivePath
	}
	return nil
}

// Handler returns an http.Handler that serves the same /api/v1/*
// surface as the network hub.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/skills/search", s.handleSearch)
	mux.HandleFunc("/api/v1/skills/", s.handleSkillByID)
	mux.HandleFunc("/api/v1/health", s.handleHealth)
	return mux
}

// ListenAndServe binds to addr and serves until the program exits.
// Common usage: `go hub.NewServer("./hub-data", ""); go s.ListenAndServe("127.0.0.1:7777")`.
func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.Handler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	return srv.ListenAndServe()
}

func (s *Server) authenticate(r *http.Request) bool {
	if s.token == "" {
		return true // no token configured = open
	}
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return false
	}
	return auth[len(prefix):] == s.token
}

func (s *Server) writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func (s *Server) writeErr(w http.ResponseWriter, code int, msg string) {
	s.writeJSON(w, code, map[string]string{"error": msg})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	count := len(s.index)
	s.mu.RUnlock()
	s.writeJSON(w, http.StatusOK, map[string]any{
		"status":      "ok",
		"skill_count": count,
		"version":     "synaptic-hub-local/0.1.0",
	})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if !s.authenticate(r) {
		s.writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	q := strings.ToLower(r.URL.Query().Get("q"))
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 100 {
		limit = n
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var results []SkillMeta
	for _, m := range s.index {
		if q == "" ||
			strings.Contains(strings.ToLower(m.Name), q) ||
			strings.Contains(strings.ToLower(m.Description), q) ||
			strings.Contains(strings.ToLower(m.Author), q) {
			results = append(results, m)
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Name < results[j].Name })
	if len(results) > limit {
		results = results[:limit]
	}
	s.writeJSON(w, http.StatusOK, SearchResult{
		Skills: results,
		Total:  len(results),
		Query:  q,
	})
}

func (s *Server) handleSkillByID(w http.ResponseWriter, r *http.Request) {
	if !s.authenticate(r) {
		s.writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	// Path: /api/v1/skills/{id} or /api/v1/skills/{id}/download
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/skills/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		s.writeErr(w, http.StatusBadRequest, "skill id required")
		return
	}
	id := parts[0]
	download := len(parts) == 2 && parts[1] == "download"
	s.mu.RLock()
	meta, ok := s.index[id]
	archivePath, aok := s.archive[id]
	s.mu.RUnlock()
	if !ok || !aok {
		s.writeErr(w, http.StatusNotFound, "skill not found")
		return
	}
	if download {
		data, err := os.ReadFile(archivePath)
		if err != nil {
			s.writeErr(w, http.StatusInternalServerError, "read archive: "+err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		_, _ = w.Write(data)
		return
	}
	s.writeJSON(w, http.StatusOK, meta)
}

// LocalAdd adds a skill to the local server's index. Useful for
// bundling skills with a private Hub or seeding a fresh server
// with company-internal skills.
//
// Security: the skill ID is validated against [a-zA-Z0-9._-]+ to
// prevent path-traversal (e.g. "../etc/passwd"). IDs that don't
// match the safe character set are rejected.
func (s *Server) LocalAdd(meta SkillMeta, archive []byte) error {
	if meta.ID == "" {
		return fmt.Errorf("hub server: skill ID required")
	}
	if !validSkillID(meta.ID) {
		return fmt.Errorf("hub server: invalid skill ID %q (must match [a-zA-Z0-9._-]+)", meta.ID)
	}
	dir := filepath.Join(s.root, "skills", meta.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// Always recompute the checksum from the actual archive bytes
	// so it can't be tampered with via meta.json.
	sum := sha256.Sum256(archive)
	meta.Checksum = hex.EncodeToString(sum[:])
	if meta.PublishedAt == "" {
		meta.PublishedAt = time.Now().UTC().Format(time.RFC3339)
	}
	meta.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), metaData, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "archive.zip"), archive, 0o644); err != nil {
		return err
	}
	s.mu.Lock()
	s.index[meta.ID] = meta
	s.archive[meta.ID] = filepath.Join(dir, "archive.zip")
	s.mu.Unlock()
	return nil
}

// validSkillID checks that an ID only contains characters safe
// for use as a filesystem directory name. Beyond the character
// set, we also reject:
//   - "." and ".." (path traversal)
//   - leading dots (hidden directories)
//   - paths that contain only dots
func validSkillID(id string) bool {
	if id == "" || len(id) > 128 {
		return false
	}
	if id == "." || id == ".." {
		return false
	}
	allDots := true
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
			allDots = false
		case r >= 'A' && r <= 'Z':
			allDots = false
		case r >= '0' && r <= '9':
			allDots = false
		case r == '.' || r == '_' || r == '-':
		default:
			return false
		}
	}
	if allDots {
		return false
	}
	return true
}

// Count returns the number of skills currently indexed.
func (s *Server) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.index)
}
