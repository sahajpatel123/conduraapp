package daemon

import (
	"context"
	"database/sql"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/halt"

	_ "modernc.org/sqlite"
)

// testConsent always returns false — simulates no GUI present.
type testConsent struct{}

func (t *testConsent) Show(ctx context.Context, ticket *gatekeeper.ConsentTicket) (bool, error) {
	return false, nil
}
func (t *testConsent) IsAvailable() bool { return true }

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", t.TempDir()+"/test.db")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS halt_state (id INTEGER PRIMARY KEY DEFAULT 1, halted INTEGER DEFAULT 0, since TEXT, reason TEXT)`)
	_, _ = db.Exec(`INSERT OR IGNORE INTO halt_state (id) VALUES (1)`)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func testEngine(t *testing.T) *gatekeeper.Engine {
	t.Helper()
	policy := gatekeeper.DefaultPolicy()
	hf := halt.New(testDB(t))
	return gatekeeper.NewEngine(policy, &testConsent{}, hf)
}

func TestE2E_Safety_ChatPassesGate(t *testing.T) {
	e := testEngine(t)
	d, _ := e.Evaluate(context.Background(), blastradius.Action{Kind: "chat", Body: "hello"})
	if d != gatekeeper.Allow {
		t.Fatalf("chat (READ) must pass, got %v", d)
	}
}

func TestE2E_Safety_WriteBlockedWithoutConsent(t *testing.T) {
	e := testEngine(t)
	d, _ := e.Evaluate(context.Background(), blastradius.Action{Kind: "click", Body: "test"})
	if d == gatekeeper.Allow {
		t.Fatal("WRITE should not be auto-allowed without consent")
	}
}

func TestE2E_Safety_ChatDoesNotHalt(t *testing.T) {
	e := testEngine(t)
	for i := 0; i < 3; i++ {
		d, _ := e.Evaluate(context.Background(), blastradius.Action{Kind: "chat"})
		if d != gatekeeper.Allow {
			t.Fatalf("chat %d should pass, got %v", i, d)
		}
	}
}

func TestE2E_Safety_SanitizerRejectsShell(t *testing.T) {
	e := testEngine(t)
	d, _ := e.Evaluate(context.Background(), blastradius.Action{
		Kind:    "shell.exec",
		Command: "ls | rm -rf /",
	})
	if d != gatekeeper.Deny {
		t.Fatal("shell.exec with pipe should be denied by sanitizer")
	}
}
