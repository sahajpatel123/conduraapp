package window

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/storage"
)

func setup(t *testing.T) (*Manager, *storage.DB) {
	t.Helper()
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{Path: filepath.Join(dir, "test.db")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return New(db.SQL()), db
}

func TestDefault(t *testing.T) {
	d := Default()
	if d.Width != 1200 || d.Height != 800 {
		t.Fatalf("Default = %+v", d)
	}
}

func TestNew_LoadsDefault(t *testing.T) {
	m, _ := setup(t)
	s := m.Snapshot()
	if s.Width != 1200 || s.Height != 800 {
		t.Fatalf("fresh load = %+v", s)
	}
}

func TestSetSize(t *testing.T) {
	m, _ := setup(t)
	if err := m.SetSize(context.Background(), 1400, 900); err != nil {
		t.Fatal(err)
	}
	s := m.Snapshot()
	if s.Width != 1400 || s.Height != 900 {
		t.Fatalf("after SetSize = %+v", s)
	}
}

func TestSetSize_RejectsZero(t *testing.T) {
	m, _ := setup(t)
	if err := m.SetSize(context.Background(), 0, 100); err == nil {
		t.Fatal("SetSize(0, 100) should error")
	}
	if err := m.SetSize(context.Background(), 100, 0); err == nil {
		t.Fatal("SetSize(100, 0) should error")
	}
}

func TestSetPosition(t *testing.T) {
	m, _ := setup(t)
	x, y := 100, 50
	if err := m.SetPosition(context.Background(), &x, &y); err != nil {
		t.Fatal(err)
	}
	s := m.Snapshot()
	if s.X == nil || *s.X != 100 {
		t.Fatalf("X = %v", s.X)
	}
	if s.Y == nil || *s.Y != 50 {
		t.Fatalf("Y = %v", s.Y)
	}
}

func TestSetPosition_NilSafe(t *testing.T) {
	m, _ := setup(t)
	if err := m.SetPosition(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestSetPosition_Partial(t *testing.T) {
	m, _ := setup(t)
	x := 200
	if err := m.SetPosition(context.Background(), &x, nil); err != nil {
		t.Fatal(err)
	}
	s := m.Snapshot()
	if s.X == nil || *s.X != 200 {
		t.Fatalf("X = %v", s.X)
	}
	if s.Y != nil {
		t.Fatalf("Y = %v, want nil", s.Y)
	}
}

func TestSetLastConversation(t *testing.T) {
	m, _ := setup(t)
	if err := m.SetLastConversation(context.Background(), 42); err != nil {
		t.Fatal(err)
	}
	if got := m.Snapshot().LastConversationID; got != 42 {
		t.Fatalf("LastConversationID = %d, want 42", got)
	}
}

// TestPersistence_ReloadOnNewManager verifies that the state survives
// a daemon restart (i.e. a fresh Manager pointing at the same DB).
func TestPersistence_ReloadOnNewManager(t *testing.T) {
	m1, db := setup(t)
	if err := m1.SetSize(context.Background(), 1500, 950); err != nil {
		t.Fatal(err)
	}
	m2 := New(db.SQL())
	s := m2.Snapshot()
	if s.Width != 1500 || s.Height != 950 {
		t.Fatalf("after reload = %+v", s)
	}
}
