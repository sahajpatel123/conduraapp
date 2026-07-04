package adaptive

import (
	"database/sql"
	"encoding/base64"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", t.TempDir()+"/test.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func passthroughEncrypt(s string, _ int64, _ string) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(s)), nil
}

func passthroughDecrypt(s string, _ int64, _ string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	return string(b), err
}

func TestEncryptedStore_SaveAndLoad(t *testing.T) {
	s, err := NewEncryptedStore(testDB(t), passthroughEncrypt, passthroughDecrypt)
	if err != nil {
		t.Fatalf("NewEncryptedStore: %v", err)
	}

	model := &UserModel{
		Identity:      InferredField{Value: "developer", Confidence: 0.9, LastSeen: time.Now(), Source: "observer"},
		RiskTolerance: InferredField{Value: "cautious", Confidence: 0.8, LastSeen: time.Now(), Source: "dialectic"},
		Version:       1,
	}
	if err := s.Save(model); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Identity.Value != "developer" {
		t.Errorf("identity = %q", loaded.Identity.Value)
	}
	if loaded.RiskTolerance.Value != "cautious" {
		t.Errorf("risk = %q", loaded.RiskTolerance.Value)
	}
}

func TestEncryptedStore_LoadEmpty(t *testing.T) {
	s, _ := NewEncryptedStore(testDB(t), passthroughEncrypt, passthroughDecrypt)
	model, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if model.Version != 1 {
		t.Errorf("version = %d", model.Version)
	}
}

func TestEncryptedStore_Reset(t *testing.T) {
	s, _ := NewEncryptedStore(testDB(t), passthroughEncrypt, passthroughDecrypt)
	model := &UserModel{Identity: InferredField{Value: "test", LastSeen: time.Now()}}
	_ = s.Save(model)
	if err := s.Reset(); err != nil {
		t.Fatal(err)
	}
	loaded, _ := s.Load()
	if loaded.Identity.Value != "" {
		t.Errorf("expected empty identity after reset, got %q", loaded.Identity.Value)
	}
}

func TestInferredField_Provenance(t *testing.T) {
	f := InferredField{
		Value: "prefers dark mode", Confidence: 0.85,
		Evidence: []string{"sess-1", "sess-3"},
		LastSeen: time.Now(), Source: "observer",
	}
	if len(f.Evidence) != 2 {
		t.Errorf("evidence = %d", len(f.Evidence))
	}
	if f.Confidence < 0.8 {
		t.Error("confidence too low")
	}
}
