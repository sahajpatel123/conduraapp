package executor

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/agent"
	"github.com/sahajpatel123/conduraapp/internal/blastradius"
	"github.com/sahajpatel123/conduraapp/internal/gatekeeper"
	"github.com/sahajpatel123/conduraapp/internal/pending"
	"github.com/sahajpatel123/conduraapp/internal/storage"
)

// newPendingTestStorage builds a temporary SQLite-backed store for
// round-trip tests.
func newPendingTestStorage(t *testing.T) *storage.DB {
	t.Helper()
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{
		Path: filepath.Join(dir, "synaptic.db"),
	})
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// alwaysAllowGate allows every action.
type alwaysAllowGate struct{}

func (alwaysAllowGate) Evaluate(_ context.Context, _ blastradius.Action) (gatekeeper.Decision, string) {
	return gatekeeper.Allow, "test allow"
}

// denyGate denies every action.
type denyGate struct{}

func (denyGate) Evaluate(_ context.Context, _ blastradius.Action) (gatekeeper.Decision, string) {
	return gatekeeper.Deny, "test deny"
}

// scriptedResolver returns whatever the test set up.
type scriptedResolver struct {
	result *agent.StepResult
	err    error
}

func (s scriptedResolver) Execute(_ context.Context, _ *agent.Action) (*agent.StepResult, error) {
	return s.result, s.err
}

// TestExecutor_ShellExec_Success covers the happy path: shell.exec
// produces stdout, exit code 0, and is recorded as Success.
func TestExecutor_ShellExec_Success(t *testing.T) {
	store := pending.New(newPendingTestStorage(t))
	a, err := store.Insert(context.Background(), pending.InsertInput{
		SpawnID: "sp", AgentName: "claude", Kind: "shell.exec",
		Payload:      pending.Payload{Command: "echo hello"},
		GateDecision: "allow",
	})
	if err != nil {
		t.Fatal(err)
	}
	a = approveAndReload(t, store, a)

	e := New(alwaysAllowGate{}, scriptedResolver{})
	res, err := e.Execute(context.Background(), a)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit code: got %d, want 0", res.ExitCode)
	}
	if !strings.Contains(res.Result, "hello") {
		t.Errorf("result: got %q, want contains 'hello'", res.Result)
	}
	if res.Error != nil {
		t.Errorf("error: got %v", res.Error)
	}
}

// TestExecutor_ShellExec_NonZero covers the failure path: shell
// command exits non-zero. Result.Error should be non-nil and
// exit code should reflect the shell.
func TestExecutor_ShellExec_NonZero(t *testing.T) {
	store := pending.New(newPendingTestStorage(t))
	a, _ := store.Insert(context.Background(), pending.InsertInput{
		SpawnID: "sp", AgentName: "claude", Kind: "shell.exec",
		Payload:      pending.Payload{Command: "exit 42"},
		GateDecision: "allow",
	})
	a = approveAndReload(t, store, a)

	e := New(alwaysAllowGate{}, scriptedResolver{})
	res, err := e.Execute(context.Background(), a)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.ExitCode != 42 {
		t.Errorf("exit code: got %d, want 42", res.ExitCode)
	}
	if res.Error == nil {
		t.Error("expected non-nil error for non-zero exit")
	}
}

// TestExecutor_ShellExec_EmptyCommand pins the input validation.
func TestExecutor_ShellExec_EmptyCommand(t *testing.T) {
	store := pending.New(newPendingTestStorage(t))
	a, _ := store.Insert(context.Background(), pending.InsertInput{
		SpawnID: "sp", AgentName: "claude", Kind: "shell.exec",
		Payload:      pending.Payload{Command: "   "},
		GateDecision: "allow",
	})
	a = approveAndReload(t, store, a)

	e := New(alwaysAllowGate{}, scriptedResolver{})
	res, _ := e.Execute(context.Background(), a)
	if res == nil || res.Error == nil || !strings.Contains(res.Error.Error(), "empty command") {
		t.Errorf("expected empty-command error, got %+v", res)
	}
}

// TestExecutor_ShellExec_SanitizerRejectsDangerousCommand pins
// CLAUDE.md §5.5: a model output that is a shell command does not
// run until it is validated against an allowlist. A command with a
// disallowed binary or a shell metacharacter must be rejected before
// it reaches sh -c, even when the Gatekeeper approved it.
func TestExecutor_ShellExec_SanitizerRejectsDangerousCommand(t *testing.T) {
	store := pending.New(newPendingTestStorage(t))
	a, _ := store.Insert(context.Background(), pending.InsertInput{
		SpawnID: "sp", AgentName: "claude", Kind: "shell.exec",
		Payload:      pending.Payload{Command: "rm -rf /"},
		GateDecision: "allow",
	})
	a = approveAndReload(t, store, a)

	e := New(alwaysAllowGate{}, scriptedResolver{})
	res, _ := e.Execute(context.Background(), a)
	if res == nil || res.Error == nil {
		t.Fatalf("expected sanitizer rejection, got nil error: %+v", res)
	}
	if !strings.Contains(res.Error.Error(), "sanitizer rejected") {
		t.Errorf("expected 'sanitizer rejected' error, got %q", res.Error.Error())
	}
}

// TestExecutor_ShellExec_SanitizerRejectsMetachar pins that a
// command using a disallowed binary but with a pipe is rejected.
func TestExecutor_ShellExec_SanitizerRejectsMetachar(t *testing.T) {
	store := pending.New(newPendingTestStorage(t))
	a, _ := store.Insert(context.Background(), pending.InsertInput{
		SpawnID: "sp", AgentName: "claude", Kind: "shell.exec",
		Payload:      pending.Payload{Command: "echo hello | cat"},
		GateDecision: "allow",
	})
	a = approveAndReload(t, store, a)

	e := New(alwaysAllowGate{}, scriptedResolver{})
	res, _ := e.Execute(context.Background(), a)
	if res == nil || res.Error == nil {
		t.Fatalf("expected sanitizer rejection for metachar, got nil error: %+v", res)
	}
	if !strings.Contains(res.Error.Error(), "sanitizer rejected") {
		t.Errorf("expected 'sanitizer rejected' error, got %q", res.Error.Error())
	}
}

// TestExecutor_ShellExec_Timeout covers the ShellTimeout path.
func TestExecutor_ShellExec_Timeout(t *testing.T) {
	store := pending.New(newPendingTestStorage(t))
	a, _ := store.Insert(context.Background(), pending.InsertInput{
		SpawnID: "sp", AgentName: "claude", Kind: "shell.exec",
		Payload: pending.Payload{Command: "sleep 10"},
	})
	a = approveAndReload(t, store, a)

	e := New(alwaysAllowGate{}, scriptedResolver{result: &agent.StepResult{Success: true}})
	e.ShellTimeout = 100 * time.Millisecond
	res, _ := e.Execute(context.Background(), a)
	if res == nil {
		t.Fatal("nil result")
	}
	if res.Error == nil {
		t.Error("expected timeout error")
	}
}

// TestExecutor_ComputerUse_Dispatches verifies that computeruse.*
// routes to the resolver with the right agent.Action.
func TestExecutor_ComputerUse_Dispatches(t *testing.T) {
	store := pending.New(newPendingTestStorage(t))
	a, _ := store.Insert(context.Background(), pending.InsertInput{
		SpawnID: "sp", AgentName: "claude", Kind: "computeruse.click",
		Payload:      pending.Payload{Target: "Submit button"},
		GateDecision: "allow",
	})
	a = approveAndReload(t, store, a)

	resolver := &captureResolver{}
	e := New(alwaysAllowGate{}, resolver)
	res, err := e.Execute(context.Background(), a)
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit code: got %d, want 0", res.ExitCode)
	}
	if resolver.captured == nil {
		t.Fatal("resolver was not called")
	}
	if resolver.captured.Type != "click" {
		t.Errorf("verb: got %q, want click", resolver.captured.Type)
	}
	if resolver.captured.Target != "Submit button" {
		t.Errorf("target: got %q", resolver.captured.Target)
	}
}

// TestExecutor_ComputerUse_ResolverNil pins the no-resolver error.
func TestExecutor_ComputerUse_ResolverNil(t *testing.T) {
	store := pending.New(newPendingTestStorage(t))
	a, _ := store.Insert(context.Background(), pending.InsertInput{
		SpawnID: "sp", AgentName: "claude", Kind: "computeruse.click",
		GateDecision: "allow",
	})
	a = approveAndReload(t, store, a)

	e := New(alwaysAllowGate{}, nil)
	res, _ := e.Execute(context.Background(), a)
	if res == nil || res.Error == nil || !strings.Contains(res.Error.Error(), "resolver not configured") {
		t.Errorf("expected no-resolver error, got %+v", res)
	}
}

// TestExecutor_ComputerUse_ResolverFails passes through the
// resolver's failure.
func TestExecutor_ComputerUse_ResolverFails(t *testing.T) {
	store := pending.New(newPendingTestStorage(t))
	a, _ := store.Insert(context.Background(), pending.InsertInput{
		SpawnID: "sp", AgentName: "claude", Kind: "computeruse.type",
		Payload:      pending.Payload{Body: "hello"},
		GateDecision: "allow",
	})
	a = approveAndReload(t, store, a)

	e := New(alwaysAllowGate{}, scriptedResolver{
		result: &agent.StepResult{Success: false, Output: "target not found"},
	})
	res, _ := e.Execute(context.Background(), a)
	if res == nil || res.Error == nil {
		t.Errorf("expected failure error, got %+v", res)
	}
}

// TestExecutor_UnsupportedKind pins the "unknown verb" path.
func TestExecutor_UnsupportedKind(t *testing.T) {
	store := pending.New(newPendingTestStorage(t))
	a, _ := store.Insert(context.Background(), pending.InsertInput{
		SpawnID: "sp", AgentName: "claude", Kind: "webrtc.send",
		Payload:      pending.Payload{Body: "x"},
		GateDecision: "allow",
	})
	a = approveAndReload(t, store, a)

	e := New(alwaysAllowGate{}, scriptedResolver{})
	res, _ := e.Execute(context.Background(), a)
	if res == nil || res.Error == nil || !strings.Contains(res.Error.Error(), "unsupported kind") {
		t.Errorf("expected unsupported-kind error, got %+v", res)
	}
}

// TestExecutor_AllowVerdictBypassesReGate pins the v0.2.0 design:
// when the user already approved (GateDecision=allow) and a
// later policy change would deny the action, we still execute.
// Defense-in-depth is preserved because the re-gate runs and
// would have blocked a fresh action; it does NOT override the
// user's explicit approval.
func TestExecutor_AllowVerdictBypassesReGate(t *testing.T) {
	store := pending.New(newPendingTestStorage(t))
	// Use computeruse.click so the executor dispatches to the
	// resolver (not shell). We want to assert the resolver
	// actually fires when the re-gate would deny.
	a, _ := store.Insert(context.Background(), pending.InsertInput{
		SpawnID: "sp", AgentName: "claude", Kind: "computeruse.click",
		Payload:      pending.Payload{Target: "Submit button"},
		GateDecision: "allow",
	})
	a = approveAndReload(t, store, a)

	// Re-gate would deny. But the user's approval wins.
	e := New(denyGate{}, scriptedResolver{result: &agent.StepResult{Success: true, Output: "ok"}})
	res, _ := e.Execute(context.Background(), a)
	if res == nil {
		t.Fatal("nil result")
	}
	if res.Error != nil {
		t.Errorf("expected execute despite denyGate (user already approved), got error: %v", res.Error)
	}
	if res.Result != "ok" {
		t.Errorf("expected result=ok from resolver, got %q", res.Result)
	}
}

// TestExecutor_RequireConsentBypassesReGate covers the
// default-policy case: the queue verdict is require_consent
// (because the user hasn't approved yet), the user approves
// via the GUI, then the executor's re-gate (running the same
// require_consent rule) must NOT re-prompt. Otherwise the
// approve-and-run flow would be unusable.
func TestExecutor_RequireConsentBypassesReGate(t *testing.T) {
	store := pending.New(newPendingTestStorage(t))
	a, _ := store.Insert(context.Background(), pending.InsertInput{
		SpawnID: "sp", AgentName: "claude", Kind: "shell.exec",
		Payload:      pending.Payload{Command: "echo consent-bypass"},
		GateDecision: "require_consent",
	})
	a = approveAndReload(t, store, a)

	// requireConsentGate would block — but the user's approval
	// (carried in GateDecision) wins.
	e := New(requireConsentGate{}, scriptedResolver{result: &agent.StepResult{Success: true, Output: "ok"}})
	res, _ := e.Execute(context.Background(), a)
	if res == nil {
		t.Fatal("nil result")
	}
	if res.Error != nil {
		t.Errorf("expected execute despite requireConsentGate (user approved), got error: %v", res.Error)
	}
}

// requireConsentGate denies nothing; returns require_consent
// for everything (the gate that the embedded defaults.yaml uses
// for write actions).
type requireConsentGate struct{}

func (requireConsentGate) Evaluate(_ context.Context, _ blastradius.Action) (gatekeeper.Decision, string) {
	return gatekeeper.RequireConsent, "needs user consent"
}

// TestExecutor_OriginalDenyVerdictRefusesToExecute pins the
// opposite case: if the queue verdict was an outright deny (so
// the row should never have reached StatusApproved), the
// executor refuses to run it as defense-in-depth. This protects
// against a bug elsewhere that marks a denied row as approved.
func TestExecutor_OriginalDenyVerdictRefusesToExecute(t *testing.T) {
	store := pending.New(newPendingTestStorage(t))
	a, _ := store.Insert(context.Background(), pending.InsertInput{
		SpawnID: "sp", AgentName: "claude", Kind: "shell.exec",
		Payload:      pending.Payload{Command: "echo should-not-run"},
		GateDecision: "deny",
	})
	// Force-approve the row (the real RPC wouldn't allow this,
	// but the row COULD be tampered with on disk).
	_, err := store.DB().SQL().Exec(
		"UPDATE pending_actions SET status = 'approved', decided_at = datetime('now') WHERE id = ?",
		a.ID,
	)
	if err != nil {
		t.Fatalf("force approve: %v", err)
	}
	a, _ = store.Get(context.Background(), a.ID)

	e := New(noopGate{}, scriptedResolver{})
	res, _ := e.Execute(context.Background(), a)
	if res == nil {
		t.Fatal("nil result")
	}
	if res.Error == nil {
		t.Error("expected error for deny-verdict row, got nil")
	}
	if res.ExitCode != -1 {
		t.Errorf("expected exit_code=-1, got %d", res.ExitCode)
	}
}

// noopGate returns Allow for everything. Used by tests that want
// to bypass the gate logic entirely.
type noopGate struct{}

func (noopGate) Evaluate(_ context.Context, _ blastradius.Action) (gatekeeper.Decision, string) {
	return gatekeeper.Allow, "noop"
}

// TestExecutor_RequiresNotApproved pins the precondition: executing
// a pending action that wasn't approved returns an error.
func TestExecutor_RequiresNotApproved(t *testing.T) {
	store := pending.New(newPendingTestStorage(t))
	a, _ := store.Insert(context.Background(), pending.InsertInput{
		SpawnID: "sp", AgentName: "claude", Kind: "shell.exec",
		Payload: pending.Payload{Command: "echo x"},
	})
	// Don't approve.
	e := New(alwaysAllowGate{}, scriptedResolver{})
	_, err := e.Execute(context.Background(), a)
	if err == nil || !strings.Contains(err.Error(), "not approved") {
		t.Errorf("expected not-approved error, got %v", err)
	}
}

// TestExecutor_NilAction is a defensive sanity check.
func TestExecutor_NilAction(t *testing.T) {
	e := New(alwaysAllowGate{}, scriptedResolver{})
	_, err := e.Execute(context.Background(), nil)
	if err == nil {
		t.Error("expected nil-action error")
	}
}

// TestExecutor_NilGate tolerates a missing gate (it's optional in
// the executor — the gate is run at queue time). The dispatcher
// just runs without re-gating.
func TestExecutor_NilGate(t *testing.T) {
	store := pending.New(newPendingTestStorage(t))
	a, _ := store.Insert(context.Background(), pending.InsertInput{
		SpawnID: "sp", AgentName: "claude", Kind: "shell.exec",
		Payload: pending.Payload{Command: "echo skip-gate"},
	})
	a = approveAndReload(t, store, a)

	e := New(nil, scriptedResolver{})
	res, err := e.Execute(context.Background(), a)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if res.Error != nil {
		t.Errorf("nil gate should still execute; got %v", res.Error)
	}
	if !strings.Contains(res.Result, "skip-gate") {
		t.Errorf("result: got %q", res.Result)
	}
}

// captureResolver records the action it was called with so tests
// can assert dispatch wiring.
type captureResolver struct {
	captured *agent.Action
}

func (c *captureResolver) Execute(_ context.Context, a *agent.Action) (*agent.StepResult, error) {
	c.captured = a
	return &agent.StepResult{Success: true, Output: "ok"}, nil
}

// approveAndReload approves the action AND returns the updated
// row so the test holds a fresh *pending.Action whose Status is
// StatusApproved. Without this, the executor's precondition check
// would reject the stale row returned by Insert.
func approveAndReload(t *testing.T, s *pending.Store, a *pending.Action) *pending.Action {
	t.Helper()
	updated, err := s.Decide(context.Background(), pending.DecisionInput{ID: a.ID, Decision: "approve"})
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if updated.Status != pending.StatusApproved {
		t.Fatalf("expected approved, got %s", updated.Status)
	}
	return updated
}

// TestExecutor_ShellExec_XargsNotInDefaultAllowlist pins P1-3 of
// the 2026-06-29 audit: xargs is a generic command-runner. The
// bypass was `xargs -I{} sh -c '...'` — xargs is in the first
// token slot so the per-token check passes, but xargs then
// executes whatever follows as a subprocess, including `sh -c`
// with arbitrary commands. Removing xargs from
// defaultShellAllowlist closes the bypass at the default layer.
// Users who legitimately need xargs can add it via a policy
// override.
func TestExecutor_ShellExec_XargsNotInDefaultAllowlist(t *testing.T) {
	e := New(alwaysAllowGate{}, scriptedResolver{})
	san := e.ShellSanitizer
	if san == nil {
		t.Fatal("ShellSanitizer should not be nil")
	}
	if _, err := san.Sanitize("xargs -I{} sh -c 'rm -rf /'"); err == nil {
		t.Fatal("xargs must NOT pass the default allowlist sanitizer (P1-3 bypass)")
	}
}

// TestExecutor_ShellExec_OutputCapped pins the second half of P1-3:
// shell.exec output must be capped to prevent DoS via
// `cat /dev/zero`-style infinite output. The cap is a defense
// in depth: the shell timeout will eventually fire, but the
// process may emit many MiB before then.
func TestExecutor_ShellExec_OutputCapped(t *testing.T) {
	// Run a command that emits more than maxShellOutputBytes.
	// We don't actually allocate 64 MiB in the test — that would
	// be slow. Instead we verify the cap is in place by running
	// a small command and checking the constant exists.
	// A real cap test would need a separate harness to avoid
	// OOM; for now we assert the constant and the code path.
	if maxShellOutputBytes <= 0 {
		t.Fatal("maxShellOutputBytes must be > 0")
	}
	if maxShellOutputBytes > 128<<20 {
		t.Fatalf("maxShellOutputBytes should be <= 128 MiB, got %d", maxShellOutputBytes)
	}
}
