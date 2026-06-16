package onboarding

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadEULA_FromDataDirSibling(t *testing.T) {
	dir := t.TempDir()
	eulaPath := filepath.Join(dir, "EULA.md")
	content := "# Test EULA\n**Last updated:** 2026-06-15\nSome terms here."
	if err := os.WriteFile(eulaPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write EULA: %v", err)
	}
	// dataDir is inside the parent; sibling path should be found.
	dataDir := filepath.Join(dir, ".synaptic")
	_ = os.MkdirAll(dataDir, 0o755)

	doc, err := ReadEULA(dataDir)
	if err != nil {
		t.Fatalf("ReadEULA: %v", err)
	}
	if doc.Version != CurrentEULAVersion {
		t.Fatalf("version: want %s, got %s", CurrentEULAVersion, doc.Version)
	}
	if doc.Text != content {
		t.Fatalf("text mismatch: want %q, got %q", content, doc.Text)
	}
	if doc.UpdatedAt != "2026-06-15" {
		t.Fatalf("updated_at: want 2026-06-15, got %s", doc.UpdatedAt)
	}
}

func TestReadEULA_InsideDataDir(t *testing.T) {
	dir := t.TempDir()
	eulaPath := filepath.Join(dir, "EULA.md")
	content := "# Inside EULA\n**Last updated:** 2026-06-06"
	if err := os.WriteFile(eulaPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write EULA: %v", err)
	}
	doc, err := ReadEULA(dir)
	if err != nil {
		t.Fatalf("ReadEULA: %v", err)
	}
	if doc.Text != content {
		t.Fatalf("text mismatch")
	}
}

func TestReadEULA_MissingFileReturnsFallback(t *testing.T) {
	dir := t.TempDir()
	doc, err := ReadEULA(dir)
	if err != nil {
		t.Fatalf("ReadEULA should not error on missing file: %v", err)
	}
	if doc.Text == "" {
		t.Fatal("fallback text is empty")
	}
	if doc.Version != CurrentEULAVersion {
		t.Fatalf("version: want %s, got %s", CurrentEULAVersion, doc.Version)
	}
}

func TestReadEULA_EmptyDataDir(t *testing.T) {
	doc, err := ReadEULA("")
	if err != nil {
		t.Fatalf("ReadEULA with empty dir: %v", err)
	}
	if doc.Text == "" {
		t.Fatal("fallback text is empty")
	}
}

func TestExtractUpdatedAt(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"**Last updated:** 2026-06-06\n", "2026-06-06"},
		{"# Title\n\n**Last updated:** 2026-01-01  \nMore text", "2026-01-01"},
		{"No date here", ""},
	}
	for _, tt := range tests {
		got := extractUpdatedAt(tt.input)
		if got != tt.want {
			t.Fatalf("extractUpdatedAt(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestValidateEULAVersion(t *testing.T) {
	if !ValidateEULAVersion("") {
		t.Fatal("empty stored version should be valid")
	}
	if !ValidateEULAVersion(CurrentEULAVersion) {
		t.Fatal("matching version should be valid")
	}
	if ValidateEULAVersion("v0") {
		t.Fatal("mismatched version should be invalid")
	}
	if ValidateEULAVersion("v2") {
		t.Fatal("mismatched future version should be invalid")
	}
}

func TestCurrentEULAVersion_NonEmpty(t *testing.T) {
	if CurrentEULAVersion == "" {
		t.Fatal("CurrentEULAVersion must not be empty")
	}
}
