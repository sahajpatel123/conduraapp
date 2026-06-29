package daemon

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/blastradius"
	"github.com/sahajpatel123/conduraapp/internal/gatekeeper"
	"github.com/sahajpatel123/conduraapp/internal/halt"
	"github.com/sahajpatel123/conduraapp/internal/sse"

	_ "modernc.org/sqlite"
)

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

// buildTestEngine returns a production engine. Pass consent=nil for
// rpcConsentProvider (no GUI → returns false → fail-closed). Pass a
// consent provider to test specific consent behavior.
func buildTestEngine(t *testing.T, consent gatekeeper.ConsentProvider) *gatekeeper.Engine {
	t.Helper()
	hf := halt.New(testDB(t))
	policy := gatekeeper.DefaultPolicy()
	if consent == nil {
		consent = &rpcConsentProvider{} // no publish → returns false
	}
	return gatekeeper.NewEngine(policy, consent, hf)
}

func TestE2E_Safety_ChatPasses(t *testing.T) {
	e := buildTestEngine(t, nil)
	d, _ := e.Evaluate(context.Background(), blastradius.Action{Kind: "chat", Body: "hello"})
	if d != gatekeeper.Allow {
		t.Fatalf("chat (READ) must pass, got %v", d)
	}
}

func TestE2E_Safety_WriteBlocked(t *testing.T) {
	e := buildTestEngine(t, nil)
	d, _ := e.Evaluate(context.Background(), blastradius.Action{Kind: "click", Body: "test"})
	if d == gatekeeper.Allow {
		t.Fatal("WRITE must not be auto-allowed without consent")
	}
}

func TestE2E_Safety_ShellSanitizerCatchesPipe(t *testing.T) {
	// Use real buildSafetyLayer — that's where SanitizeHook is installed.
	// The Engine struct is public so we access the consent field directly.
	hf := halt.New(testDB(t))
	broker := sse.NewBroker()
	safety := buildSafetyLayer(hf, broker, nil, nil, nil)

	// Override consent to approve, so the only thing that can
	// produce Deny is the SanitizeHook.
	safety.Engine.SetConsentProvider(&approveConsent{})

	d, _ := safety.Engine.Evaluate(context.Background(), blastradius.Action{
		Kind:    "shell.exec",
		Command: "ls | rm -rf /",
	})
	if d != gatekeeper.Deny {
		t.Fatal("shell.exec with pipe must be Denied by SanitizeHook (not consent)")
	}
}

func TestE2E_Safety_ChatDoesNotHalt(t *testing.T) {
	// Use real buildSafetyLayer so AnomalyHook is installed.
	hf := halt.New(testDB(t))
	broker := sse.NewBroker()
	safety := buildSafetyLayer(hf, broker, nil, nil, nil)
	e := safety.Engine

	for i := 0; i < 3; i++ {
		d, _ := e.Evaluate(context.Background(), blastradius.Action{Kind: "chat"})
		if d != gatekeeper.Allow {
			t.Fatalf("chat %d must pass, got %v", i, d)
		}
	}
	// Anomaly detector processes async; wait so halt state is observable.
	time.Sleep(100 * time.Millisecond)
	if hf.IsHalted() {
		t.Fatal("3 chats must NOT halt the daemon")
	}
}

func TestE2E_Safety_HaltBlocks(t *testing.T) {
	hf := halt.New(testDB(t))
	broker := sse.NewBroker()
	safety := buildSafetyLayer(hf, broker, nil, nil, nil)
	e := safety.Engine

	_, _ = hf.Halt(context.Background(), "test halt")
	d, _ := e.Evaluate(context.Background(), blastradius.Action{Kind: "click", Body: "test"})
	if d != gatekeeper.Deny {
		t.Fatal("halted engine must deny")
	}
}

type approveConsent struct{}

func (a *approveConsent) Show(_ context.Context, _ *gatekeeper.ConsentTicket) (bool, error) {
	return true, nil
}
func (a *approveConsent) IsAvailable() bool { return true }
