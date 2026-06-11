package skills

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func testStore(t *testing.T) *SQLiteStore {
	t.Helper()
	path := t.TempDir() + "/skills.db"
	s, err := NewSQLiteStore(path)
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestSQLiteStore_CreateAndGet(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	sk := &Skill{
		ID: "test-1", Name: "organize-downloads", Description: "Clean up Downloads",
		Version: "1.0.0", Trust: TrustCommunity, TriggerPattern: "organize downloads",
		Steps:     []string{"open Finder", "navigate to ~/Downloads"},
		CreatedAt: time.Now(), UpdatedAt: time.Now(), LastUsed: time.Now(),
	}
	if err := s.Create(ctx, sk); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := s.Get(ctx, "test-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "organize-downloads" {
		t.Errorf("name = %q", got.Name)
	}
	if len(got.Steps) != 2 {
		t.Errorf("steps = %d", len(got.Steps))
	}
}

func TestSQLiteStore_List(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		sk := &Skill{ID: fmt.Sprintf("s%d", i), Name: "test", CreatedAt: time.Now(), UpdatedAt: time.Now(), LastUsed: time.Now()}
		if err := s.Create(ctx, sk); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.List(ctx, 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) < 3 {
		t.Errorf("got %d skills", len(got))
	}
}

func TestSQLiteStore_Search(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	if err := s.Create(ctx, &Skill{ID: "a", Name: "email", TriggerPattern: "send email", CreatedAt: time.Now(), UpdatedAt: time.Now(), LastUsed: time.Now()}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Search(ctx, "email", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d results", len(got))
	}
	if got[0].Name != "email" {
		t.Errorf("name = %q", got[0].Name)
	}
}

func TestSQLiteStore_Delete(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	if err := s.Create(ctx, &Skill{ID: "d", Name: "x", CreatedAt: time.Now(), UpdatedAt: time.Now(), LastUsed: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, "d"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get(ctx, "d"); err == nil {
		t.Fatal("expected error for deleted skill")
	}
}

func TestSQLiteStore_IncrementUsage(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	if err := s.Create(ctx, &Skill{ID: "u", Name: "test", CreatedAt: time.Now(), UpdatedAt: time.Now(), LastUsed: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if err := s.IncrementUsage(ctx, "u", true); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get(ctx, "u")
	if got.SuccessCount != 1 {
		t.Errorf("success = %d", got.SuccessCount)
	}
}
