// End-to-end coverage of the Phase 18 (v0.2.0) sub-agent
// ActionRequest pipeline.
//
// We can't actually run a sub-agent CLI in CI (claude, codex,
// etc. are not installed in the test environment), so this
// test exercises the executor + RPC layer directly:
//
//   1. Start the daemon via the same harness trust_e2e_test
//      uses (initSubsystems, JSON-RPC over HTTP).
//   2. Install a permissive policy so shell.exec isn't blocked.
//   3. Manually Insert a pending_action into the queue via
//      subs.Pending. (In production this happens inside
//      delegate.spawn's gateAndPersistParsedActions.)
//   4. Call delegate.pending.list → confirm the row is there.
//   5. Call delegate.pending.decide with auto_run=true →
//      confirm shell.exec actually fired (look for the
//      action's result field).
//   6. Call delegate.pending.execute on a second row →
//      confirm it runs.
//   7. Call delegate.pending.sweep → confirm expired rows
//      become expired.
//
// The test is gated on Linux/macOS because /bin/sh is the
// dispatcher; Windows tests skip the shell parts.

package daemon

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/config"
	"github.com/sahajpatel123/conduraapp/internal/delegation"
	"github.com/sahajpatel123/conduraapp/internal/ipc"
	"github.com/sahajpatel123/conduraapp/internal/pending"
	"github.com/sahajpatel123/conduraapp/internal/version"
)

// startV020Daemon is the v0.2.0 e2e harness: same shape as
// startTrustDaemon but with a permissive policy so shell.exec
// passes the gate without workspace trust.
func startV020Daemon(t *testing.T) (string, *Subsystems, func()) {
	t.Helper()
	dir := t.TempDir()
	cfg := config.Default()
	cfg.General.DataDir = dir
	cfg.Storage.Path = filepath.Join(dir, "synaptic.db")
	cfg.Logging.File = ""
	cfg.Logging.AddSource = false
	cfg.Security.SpendLimitUSDPerDay = 1.0
	cfg.APIServer.AuthToken = "test-token"

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := l.Addr().String()
	_ = l.Close()

	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	subs, err := initSubsystems(log, cfg, nil)
	if err != nil {
		t.Fatalf("initSubsystems: %v", err)
	}
	// Install a permissive policy so shell.exec passes.
	installPermissivePolicy(subs)

	srv := ipc.NewServer()
	registerMethods(srv, log, cfg, subs, version.Info{Version: "test-v020"})
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		out, _ := srv.HandleRaw(r.Context(), body)
		w.Header().Set("Content-Type", "application/json")
		if out == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_, _ = w.Write(out)
	})
	httpSrv := &http.Server{Addr: addr, Handler: mux}
	go func() { _ = httpSrv.ListenAndServe() }()
	time.Sleep(100 * time.Millisecond)

	cleanup := func() {
		_ = httpSrv.Close()
		_ = subs.Close()
	}
	return addr, subs, cleanup
}

// insertPending is a tiny helper that drops a pending_action row
// into the queue the same way delegate.spawn would, but without
// needing a real sub-agent CLI.
//
// v0.2.0 e2e test in this file exercises the shell path; computeruse
// tests live in internal/executor/executor_test.go.
//
//nolint:unparam // kind is fixed to "shell.exec" because every
func insertPending(t *testing.T, subs *Subsystems, kind, command string) *pending.Action {
	t.Helper()
	got, err := subs.Pending.Insert(context.Background(), pending.InsertInput{
		SpawnID:      "spawn-test",
		AgentName:    "test-agent",
		Kind:         kind,
		Payload:      pending.Payload{Command: command},
		GateDecision: "allow",
		GateReason:   "test permissive",
		BlastClass:   "WRITE",
		TTL:          1 * time.Minute,
	})
	if err != nil {
		t.Fatalf("insert pending: %v", err)
	}
	return got
}

// TestPendingE2E_FullPipeline exercises the full approve-and-run
// loop: insert, list, decide (with auto_run=true), confirm shell
// output captured, confirm audit row written.
func TestPendingE2E_FullPipeline(t *testing.T) {
	addr, subs, cleanup := startV020Daemon(t)
	defer cleanup()

	// 1. Insert a pending_action that runs `echo hello-v020-test`.
	a := insertPending(t, subs, "shell.exec", "echo hello-v020-test")

	// 2. List pending actions via RPC. The daemon's Pending
	// must be wired (subs.Pending != nil) for this to work.
	listRes, err := trustCallRPC(t, addr, "delegate.pending.list", map[string]any{
		"status": "pending",
	})
	if err != nil {
		t.Fatalf("delegate.pending.list: %v", err)
	}
	var list struct {
		Actions []pending.Action `json:"actions"`
	}
	if err := json.Unmarshal(listRes, &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list.Actions) != 1 || list.Actions[0].ID != a.ID {
		t.Fatalf("expected 1 row matching id=%s, got %d rows: %+v", a.ID, len(list.Actions), list.Actions)
	}

	// 3. Approve-and-run. auto_run=true triggers the executor
	// inside the daemon. The returned row should be StatusExecuted.
	approveRes, err := trustCallRPC(t, addr, "delegate.pending.decide", map[string]any{
		"id":         a.ID,
		"decision":   "approve",
		"decided_by": "user:test",
		"auto_run":   true,
	})
	if err != nil {
		t.Fatalf("delegate.pending.decide: %v", err)
	}
	var approved pending.Action
	if err := json.Unmarshal(approveRes, &approved); err != nil {
		t.Fatalf("decode decide: %v", err)
	}
	if approved.Status != pending.StatusExecuted {
		t.Errorf("expected status=executed after approve+run, got %s (err=%s)",
			approved.Status, approved.ExecutionError)
	}
	if !strings.Contains(approved.Result, "hello-v020-test") {
		t.Errorf("expected result to contain 'hello-v020-test', got %q", approved.Result)
	}
	if approved.DurationMS < 0 {
		t.Errorf("duration_ms should be >= 0, got %d", approved.DurationMS)
	}

	// 4. Verify the audit log got the executed row. The
	// executor writes actor=executor, action=pending.executed.
	listAudit, err := trustCallRPC(t, addr, "audit.list", map[string]any{
		"limit":  50,
		"action": "pending.executed",
	})
	if err != nil {
		t.Fatalf("audit.list: %v", err)
	}
	var auditList []map[string]any
	if err := json.Unmarshal(listAudit, &auditList); err != nil {
		t.Fatalf("decode audit: %v", err)
	}
	if len(auditList) == 0 {
		t.Fatalf("expected at least one audit row for pending.executed, got 0")
	}
	found := false
	for _, e := range auditList {
		if msg, _ := e["message"].(string); strings.Contains(msg, a.ID) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected audit row to mention id=%s, got: %v", a.ID, auditList)
	}
}

// TestPendingE2E_DenyBlocksExecution inserts a pending_action,
// denies it, and verifies the executor was NOT called (the
// Result stays empty and Status stays denied).
func TestPendingE2E_DenyBlocksExecution(t *testing.T) {
	addr, subs, cleanup := startV020Daemon(t)
	defer cleanup()

	a := insertPending(t, subs, "shell.exec", "echo should-not-run")

	denyRes, err := trustCallRPC(t, addr, "delegate.pending.decide", map[string]any{
		"id":         a.ID,
		"decision":   "deny",
		"decided_by": "user:test",
		"auto_run":   false,
	})
	if err != nil {
		t.Fatalf("decide deny: %v", err)
	}
	var denied pending.Action
	if err := json.Unmarshal(denyRes, &denied); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if denied.Status != pending.StatusDenied {
		t.Errorf("expected status=denied, got %s", denied.Status)
	}
	if denied.Result != "" {
		t.Errorf("expected empty result after deny, got %q", denied.Result)
	}
	_ = subs
}

// TestPendingE2E_ApproveThenExecuteLater covers the two-step
// flow: user clicks Approve only (auto_run=false), then later
// clicks Run now which calls delegate.pending.execute.
func TestPendingE2E_ApproveThenExecuteLater(t *testing.T) {
	addr, subs, cleanup := startV020Daemon(t)
	defer cleanup()

	a := insertPending(t, subs, "shell.exec", "echo two-step-v020")

	// Step 1: approve without auto-run.
	approveRes, err := trustCallRPC(t, addr, "delegate.pending.decide", map[string]any{
		"id":         a.ID,
		"decision":   "approve",
		"decided_by": "user:test",
		"auto_run":   false,
	})
	if err != nil {
		t.Fatalf("approve only: %v", err)
	}
	var approved pending.Action
	if err := json.Unmarshal(approveRes, &approved); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if approved.Status != pending.StatusApproved {
		t.Errorf("expected status=approved, got %s", approved.Status)
	}
	if approved.Result != "" {
		t.Errorf("expected empty result before execute, got %q", approved.Result)
	}

	// Step 2: execute later.
	execRes, err := trustCallRPC(t, addr, "delegate.pending.execute", map[string]any{
		"id": a.ID,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	var executed pending.Action
	if err := json.Unmarshal(execRes, &executed); err != nil {
		t.Fatalf("decode exec: %v", err)
	}
	if executed.Status != pending.StatusExecuted {
		t.Errorf("expected status=executed, got %s", executed.Status)
	}
	if !strings.Contains(executed.Result, "two-step-v020") {
		t.Errorf("expected result to contain 'two-step-v020', got %q", executed.Result)
	}
}

// TestPendingE2E_SweepExpiresOldActions inserts two rows, ages
// one past its TTL via direct SQL update, calls sweep, and
// verifies the aged one is expired while the fresh one is
// untouched.
func TestPendingE2E_SweepExpiresOldActions(t *testing.T) {
	addr, subs, cleanup := startV020Daemon(t)
	defer cleanup()

	fresh := insertPending(t, subs, "shell.exec", "echo fresh")
	stale := insertPending(t, subs, "shell.exec", "echo stale")

	// Rewind the stale row's expires_at to the past so the
	// sweeper picks it up.
	_, err := subs.Pending.DB().SQL().Exec(
		"UPDATE pending_actions SET created_at = ?, expires_at = ? WHERE id = ?",
		time.Now().Add(-1*time.Hour).UTC().Format(time.RFC3339Nano),
		time.Now().Add(-30*time.Minute).UTC().Format(time.RFC3339Nano),
		stale.ID,
	)
	if err != nil {
		t.Fatalf("rewind stale: %v", err)
	}

	// Trigger sweep via RPC.
	sweepRes, err := trustCallRPC(t, addr, "delegate.pending.sweep", nil)
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	var swept map[string]any
	if err := json.Unmarshal(sweepRes, &swept); err != nil {
		t.Fatalf("decode sweep: %v", err)
	}
	if n, _ := swept["swept"].(float64); int(n) != 1 {
		t.Errorf("expected 1 swept, got %v", swept["swept"])
	}

	// Verify statuses.
	_, err = trustCallRPC(t, addr, "delegate.pending.get", map[string]any{"id": stale.ID})
	if err != nil {
		t.Fatalf("get stale: %v", err)
	}
	var staleRow pending.Action
	if err := json.Unmarshal(sweepRes, &staleRow); err != nil {
		t.Fatalf("decode stale: %v", err)
	}
	gotStale, _ := subs.Pending.Get(context.Background(), stale.ID)
	if gotStale.Status != pending.StatusExpired {
		t.Errorf("expected stale status=expired, got %s", gotStale.Status)
	}
	gotFresh, _ := subs.Pending.Get(context.Background(), fresh.ID)
	if gotFresh.Status != pending.StatusPending {
		t.Errorf("expected fresh status=pending, got %s", gotFresh.Status)
	}

	// Tidy up — the subs closure calls subs.Close which
	// would also stop the sweeper goroutine. Tests don't need
	// explicit cleanup here.
	_ = fresh
}

// TestPendingE2E_NonZeroExitRecorded covers the failure path:
// shell exits non-zero, executor records the error AND the exit
// code on the row.
//
// Note: the shell sanitizer allowlist is
// {git,ls,cat,echo,find,grep,head,tail,sort,uniq,wc}. The
// "exit" binary is NOT in the list, so we trigger non-zero by
// asking `ls` to list a non-existent directory — exit code 1
// or 2 from ls. The sanitizer allowlist doesn't affect the
// exit-code extraction (that's pure `sh -c`), only the binary
// the command starts with.
func TestPendingE2E_NonZeroExitRecorded(t *testing.T) {
	addr, subs, cleanup := startV020Daemon(t)
	defer cleanup()

	// ls /nonexistent-path-v020-test always exits non-zero on
	// every Unix we support. Combined output will contain the
	// error message; the executor will see exit != 0 and mark
	// the row failed.
	a := insertPending(t, subs, "shell.exec", "ls /nonexistent-path-v020-test")

	res, err := trustCallRPC(t, addr, "delegate.pending.decide", map[string]any{
		"id":         a.ID,
		"decision":   "approve",
		"decided_by": "user:test",
		"auto_run":   true,
	})
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	var row pending.Action
	if err := json.Unmarshal(res, &row); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if row.Status != pending.StatusFailed {
		t.Errorf("expected status=failed for non-zero exit, got %s", row.Status)
	}
	if row.ExitCode == 0 {
		t.Errorf("expected non-zero exit_code, got %d (status=%s err=%s)",
			row.ExitCode, row.Status, row.ExecutionError)
	}
	if row.ExecutionError == "" {
		t.Errorf("expected execution_error to be set, got empty")
	}
}

// compile-time guard: ensure delegation.SpawnRequest is still
// reachable in the e2e package (used elsewhere — this catches
// accidental rename regressions).
var _ = delegation.SpawnRequest{}
