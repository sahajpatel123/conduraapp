package skills

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestParseArchive_FullSkill pins the happy path: a complete Skill JSON
// archive MUST unmarshal into a Skill with every field populated. This is
// the path the Skills Hub uses to ingest a downloaded archive; a regression
// here would silently drop fields after download.
func TestParseArchive_FullSkill(t *testing.T) {
	published := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	src := []byte(`{
		"id": "skill-001",
		"name": "Summarize Webpage",
		"description": "Fetch a URL and produce a 3-bullet summary",
		"version": "1.2.3",
		"trust": "official",
		"trigger_pattern": "summarize https://",
		"steps": ["fetch", "extract", "summarize"],
		"dependencies": ["requests", "beautifulsoup4"],
		"success_count": 42,
		"failure_count": 3,
		"created_at": "2026-06-01T10:00:00Z",
		"updated_at": "2026-07-01T11:30:00Z",
		"last_used": "2026-07-15T08:00:00Z",
		"author": "alice@example.com",
		"author_key": "abcd1234",
		"license": "Apache-2.0",
		"source": "hub",
		"hub_id": "summarize-webpage-v1",
		"checksum": "sha256:deadbeef",
		"published_at": "2026-07-01T12:00:00Z"
	}`)

	sk, err := ParseArchive(src)
	if err != nil {
		t.Fatalf("ParseArchive(full): %v", err)
	}
	if sk.ID != "skill-001" {
		t.Errorf("ID = %q, want skill-001", sk.ID)
	}
	if sk.Name != "Summarize Webpage" {
		t.Errorf("Name = %q, want %q", sk.Name, "Summarize Webpage")
	}
	if sk.Description == "" {
		t.Error("Description empty; want non-empty")
	}
	if sk.Version != "1.2.3" {
		t.Errorf("Version = %q, want 1.2.3", sk.Version)
	}
	if sk.Trust != TrustOfficial {
		t.Errorf("Trust = %q, want %q", sk.Trust, TrustOfficial)
	}
	if len(sk.Steps) != 3 || sk.Steps[0] != "fetch" {
		t.Errorf("Steps = %v, want [fetch extract summarize]", sk.Steps)
	}
	if len(sk.Dependencies) != 2 {
		t.Errorf("Dependencies len = %d, want 2", len(sk.Dependencies))
	}
	if sk.SuccessCount != 42 || sk.FailureCount != 3 {
		t.Errorf("success/failure counts = %d/%d, want 42/3", sk.SuccessCount, sk.FailureCount)
	}
	if sk.Author != "alice@example.com" {
		t.Errorf("Author = %q, want alice@example.com", sk.Author)
	}
	if sk.HubID != "summarize-webpage-v1" {
		t.Errorf("HubID = %q, want summarize-webpage-v1", sk.HubID)
	}
	if sk.PublishedAt == nil || !sk.PublishedAt.Equal(published) {
		t.Errorf("PublishedAt = %v, want %v", sk.PublishedAt, published)
	}
}

// TestParseArchive_MinimalRequired pins the minimum-viable archive: an
// archive containing only ID + Name MUST succeed (all other fields are
// optional). Pins the contract that provenance fields are NOT required
// for a valid skill — a freshly-authored skill has no Checksum, no
// HubID, no PublishedAt.
func TestParseArchive_MinimalRequired(t *testing.T) {
	src := []byte(`{"id":"skill-002","name":"Minimal Skill"}`)

	sk, err := ParseArchive(src)
	if err != nil {
		t.Fatalf("ParseArchive(minimal): %v", err)
	}
	if sk.ID != "skill-002" || sk.Name != "Minimal Skill" {
		t.Errorf("ID/Name = %q/%q, want skill-002/Minimal Skill", sk.ID, sk.Name)
	}
	if sk.Steps != nil {
		t.Errorf("Steps = %v, want nil (no steps in minimal archive)", sk.Steps)
	}
	if sk.PublishedAt != nil {
		t.Errorf("PublishedAt = %v, want nil for non-hub skill", sk.PublishedAt)
	}
}

// TestParseArchive_MissingID pins the ID-required guard: an archive
// without an ID MUST be rejected. An empty-ID skill would silently
// collide with every other empty-ID skill in the local Store.Create path.
func TestParseArchive_MissingID(t *testing.T) {
	src := []byte(`{"name":"No-ID Skill"}`)

	_, err := ParseArchive(src)
	if err == nil {
		t.Fatal("ParseArchive(missing id) = nil; want error")
	}
	if !strings.Contains(err.Error(), "missing id or name") {
		t.Errorf("error %q should mention 'missing id or name'", err.Error())
	}
}

// TestParseArchive_MissingName pins the Name-required guard: same
// reasoning as TestParseArchive_MissingID — empty Name would collide
// in the GUI list view.
func TestParseArchive_MissingName(t *testing.T) {
	src := []byte(`{"id":"skill-no-name"}`)

	_, err := ParseArchive(src)
	if err == nil {
		t.Fatal("ParseArchive(missing name) = nil; want error")
	}
	if !strings.Contains(err.Error(), "missing id or name") {
		t.Errorf("error %q should mention 'missing id or name'", err.Error())
	}
}

// TestParseArchive_EmptyJSONObject pins the `{}` boundary: a syntactically
// valid JSON object that has neither ID nor Name MUST be rejected by the
// same missing-id-or-name guard (not by a JSON parse error).
func TestParseArchive_EmptyJSONObject(t *testing.T) {
	src := []byte(`{}`)

	_, err := ParseArchive(src)
	if err == nil {
		t.Fatal("ParseArchive({}) = nil; want error")
	}
	if !strings.Contains(err.Error(), "missing id or name") {
		t.Errorf("error %q should mention 'missing id or name' (not a JSON parse error)", err.Error())
	}
}

// TestParseArchive_MalformedJSON pins the JSON-decode-failure path:
// invalid JSON syntax MUST return a wrapped parse error (preserving the
// underlying cause for logs), not a panic.
func TestParseArchive_MalformedJSON(t *testing.T) {
	src := []byte(`{"id":"foo","name":"bar"`) // truncated, missing closing braces

	_, err := ParseArchive(src)
	if err == nil {
		t.Fatal("ParseArchive(malformed) = nil; want error")
	}
	if !strings.Contains(err.Error(), "parse archive") {
		t.Errorf("error %q should mention 'parse archive'", err.Error())
	}
}

// TestParseArchive_EmptyBytes pins the empty-input boundary: zero bytes
// MUST return a wrapped parse error. A regression that treated empty as
// "valid empty skill" would let a truncated download create a phantom
// skill in the Store.
func TestParseArchive_EmptyBytes(t *testing.T) {
	_, err := ParseArchive([]byte(""))
	if err == nil {
		t.Fatal("ParseArchive(empty) = nil; want error")
	}
	if !strings.Contains(err.Error(), "parse archive") {
		t.Errorf("error %q should mention 'parse archive'", err.Error())
	}
}

// TestMarshalArchive_ValidSkillProducesJSON pins the happy path: a
// non-nil Skill MUST marshal to non-empty JSON bytes with no error.
// Pins that the JSON tag structure (id, name, description, ...) is
// applied and that a non-trivial skill serializes to non-empty bytes.
func TestMarshalArchive_ValidSkillProducesJSON(t *testing.T) {
	sk := &Skill{
		ID:      "skill-003",
		Name:    "Marshal Test",
		Version: "0.1.0",
		Trust:   TrustCommunity,
	}

	raw, err := MarshalArchive(sk)
	if err != nil {
		t.Fatalf("MarshalArchive: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("MarshalArchive returned empty bytes; want non-empty JSON")
	}

	// Round-trip parse just to confirm the output is valid JSON
	// (would otherwise be a tautology test).
	var back Skill
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("MarshalArchive output is not valid JSON: %v", err)
	}
	if back.ID != sk.ID || back.Name != sk.Name {
		t.Errorf("marshal output lost fields: ID=%q Name=%q", back.ID, back.Name)
	}
}

// TestMarshalArchive_NilSkillReturnsError pins the nil-receiver guard:
// MarshalArchive MUST reject a nil Skill with a clear error rather than
// panicking on the nil-pointer dereference inside json.Marshal.
func TestMarshalArchive_NilSkillReturnsError(t *testing.T) {
	_, err := MarshalArchive(nil)
	if err == nil {
		t.Fatal("MarshalArchive(nil) = nil; want error")
	}
	if !strings.Contains(err.Error(), "nil skill") {
		t.Errorf("error %q should mention 'nil skill'", err.Error())
	}
}

// TestMarshalArchive_RoundTripPreservesFields pins the marshal/parse
// round-trip: every important field on the Skill struct MUST survive
// MarshalArchive -> ParseArchive. This is the contract the Skills Hub
// publish + download flow relies on (publish = Marshal, download =
// Parse). A regression in either direction would silently corrupt
// skills at the hub boundary.
//
// Time fields are compared via Equal() at second precision because JSON
// RFC3339Nano round-trips at nanosecond but JSON-decoded time.Time is
// not always Equal() to the original due to monotonic-clock stripping.
func TestMarshalArchive_RoundTripPreservesFields(t *testing.T) {
	original := &Skill{
		ID:             "skill-roundtrip",
		Name:           "Round-trip Test",
		Description:    "Verify marshal -> parse preserves fields",
		Version:        "2.0.1",
		Trust:          TrustExperimental,
		TriggerPattern: "rt ",
		Steps:          []string{"step-a", "step-b"},
		Dependencies:   []string{"dep-x"},
		SuccessCount:   7,
		FailureCount:   1,
		CreatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:      time.Date(2026, 7, 19, 0, 0, 0, 0, time.UTC),
		Author:         "bob@example.com",
		License:        "MIT",
		Source:         "local",
	}

	raw, err := MarshalArchive(original)
	if err != nil {
		t.Fatalf("MarshalArchive: %v", err)
	}
	back, err := ParseArchive(raw)
	if err != nil {
		t.Fatalf("ParseArchive after Marshal: %v", err)
	}

	if back.ID != original.ID {
		t.Errorf("ID: got %q, want %q", back.ID, original.ID)
	}
	if back.Name != original.Name {
		t.Errorf("Name: got %q, want %q", back.Name, original.Name)
	}
	if back.Description != original.Description {
		t.Errorf("Description: got %q, want %q", back.Description, original.Description)
	}
	if back.Version != original.Version {
		t.Errorf("Version: got %q, want %q", back.Version, original.Version)
	}
	if back.Trust != original.Trust {
		t.Errorf("Trust: got %q, want %q", back.Trust, original.Trust)
	}
	if back.TriggerPattern != original.TriggerPattern {
		t.Errorf("TriggerPattern: got %q, want %q", back.TriggerPattern, original.TriggerPattern)
	}
	if len(back.Steps) != len(original.Steps) {
		t.Errorf("Steps len: got %d, want %d", len(back.Steps), len(original.Steps))
	}
	if len(back.Dependencies) != len(original.Dependencies) {
		t.Errorf("Dependencies len: got %d, want %d", len(back.Dependencies), len(original.Dependencies))
	}
	if back.SuccessCount != original.SuccessCount {
		t.Errorf("SuccessCount: got %d, want %d", back.SuccessCount, original.SuccessCount)
	}
	if back.FailureCount != original.FailureCount {
		t.Errorf("FailureCount: got %d, want %d", back.FailureCount, original.FailureCount)
	}
	if back.Author != original.Author {
		t.Errorf("Author: got %q, want %q", back.Author, original.Author)
	}
	if back.License != original.License {
		t.Errorf("License: got %q, want %q", back.License, original.License)
	}
	if back.Source != original.Source {
		t.Errorf("Source: got %q, want %q", back.Source, original.Source)
	}
	// Time fields: compare via Equal() — RFC3339Nano preserves enough
	// precision for second-level timestamps to round-trip exactly.
	if !back.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("CreatedAt: got %v, want %v", back.CreatedAt, original.CreatedAt)
	}
	if !back.UpdatedAt.Equal(original.UpdatedAt) {
		t.Errorf("UpdatedAt: got %v, want %v", back.UpdatedAt, original.UpdatedAt)
	}
}
