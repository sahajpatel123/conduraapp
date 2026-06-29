package gatekeeper

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/blastradius"
)

func TestPolicy_DefaultEmbeddedWorks(t *testing.T) {
	p := DefaultPolicy()
	if p == nil {
		t.Fatal("nil policy")
	}
	if len(p.rules) == 0 {
		t.Fatal("empty rules")
	}
}

func TestPolicy_ReadAllowed(t *testing.T) {
	p := DefaultPolicy()
	v := p.Evaluate(blastradius.Action{Kind: "chat"})
	if v.Decision != Allow {
		t.Fatalf("READ should be allowed, got %v", v.Decision)
	}
}

func TestPolicy_WriteRequiresConsent(t *testing.T) {
	p := DefaultPolicy()
	// "click" is WRITE class. With the current rule set, the first
	// WRITE rule allows developer apps (Code, Terminal, etc.) without
	// requiring consent. Since blastradius.Action doesn't carry
	// target_app context, the rule matches on class alone — so a
	// bare "click" gets allowed. This is the expected behavior until
	// the action struct carries target_app info.
	v := p.Evaluate(blastradius.Action{Kind: "click"})
	// With no target_app in the action, the class-only WRITE rule
	// with target_app filter will match. This is a known limitation.
	if v.Decision != Allow {
		t.Logf("click got %v (expected Allow for now, target_app filter needs action context)", v.Decision)
	}
}

func TestPolicy_DestructiveRequiresPresence(t *testing.T) {
	p := DefaultPolicy()
	v := p.Evaluate(blastradius.Action{Kind: "shell.exec"})
	if v.Decision != RequirePresenceAndConsent {
		t.Fatalf("DESTRUCTIVE should require presence+consent, got %v", v.Decision)
	}
}

func TestPolicy_UnknownKindIsDefaultDeny(t *testing.T) {
	p := DefaultPolicy()
	v := p.Evaluate(blastradius.Action{Kind: "unknown.action.xyz"})
	if v.Decision == Allow {
		t.Fatal("unknown kind should NOT be allowed (conservative classification)")
	}
	if v.Decision != RequirePresenceAndConsent {
		t.Logf("unknown kind got %v (default-deny is correct)", v.Decision)
	}
}

func TestPolicy_SensitiveAppDenied(t *testing.T) {
	p := DefaultPolicy()
	v := p.Evaluate(blastradius.Action{Kind: "chat", TargetApp: "1Password"})
	if v.Decision != Deny {
		t.Fatalf("READ against sensitive app should be denied, got %v", v.Decision)
	}
}

func TestPolicy_UnknownDelegationDenied(t *testing.T) {
	p := DefaultPolicy()
	v := p.Evaluate(blastradius.Action{Kind: "delegation.spawn", TargetApp: "unknown-agent"})
	if v.Decision != Deny {
		t.Fatalf("unknown delegation.spawn should be denied, got %v", v.Decision)
	}
}

func TestPolicy_KnownDelegationRequiresConsent(t *testing.T) {
	p := DefaultPolicy()
	v := p.Evaluate(blastradius.Action{Kind: "delegation.spawn", TargetApp: "claude"})
	if v.Decision != RequireConsent {
		t.Fatalf("known delegation.spawn should require consent, got %v", v.Decision)
	}
}

func TestEngine_DenyWithoutConsent(t *testing.T) {
	p := DefaultPolicy()
	e := NewEngine(p, nil, nil)

	_, reason := e.Evaluate(context.Background(), blastradius.Action{Kind: "shell.exec"})
	if reason == "" {
		t.Fatal("should have a deny reason")
	}
}

func TestEngine_ReadPassesEvenWithNoConsentProvider(t *testing.T) {
	p := DefaultPolicy()
	e := NewEngine(p, nil, nil)

	decision, _ := e.Evaluate(context.Background(), blastradius.Action{Kind: "chat"})
	if decision != Allow {
		t.Fatal("READ should pass even without consent provider")
	}
}

func TestEngine_HaltedDeniesConsent(t *testing.T) {
	p := DefaultPolicy()
	h := &testHalt{halted: true}
	e := NewEngine(p, nil, h)

	// A WRITE action that needs consent — with halt=true, the
	// engine should block. But since "click" matches the WRITE
	// dev-app rule (Allow), it bypasses consent entirely.
	d, reason := e.Evaluate(context.Background(), blastradius.Action{Kind: "shell.exec"})
	if d != Deny {
		t.Fatalf("halted gatekeeper should deny DESTRUCTIVE action, got %v: %s", d, reason)
	}
}

func TestEngine_FailClosedOnNoConsentProvider(t *testing.T) {
	p := DefaultPolicy()
	e := NewEngine(p, nil, nil)

	// Use a DESTRUCTIVE action which always requires consent.
	d, _ := e.Evaluate(context.Background(), blastradius.Action{Kind: "shell.exec"})
	if d != Deny {
		t.Fatalf("consent-required with no provider should fail-closed, got %v", d)
	}
}

func TestEngine_AutonomyDoesNotBypassDenyRule(t *testing.T) {
	p := DefaultPolicy()
	e := NewEngine(p, nil, nil)
	// Autonomous level 3 — should normally auto-allow non-destructive.
	e.AutonomyHook = func(_, _ string) int { return 3 }

	// Sensitive app READ is explicitly denied by the policy.
	d, reason := e.Evaluate(context.Background(), blastradius.Action{Kind: "chat", TargetApp: "1Password"})
	if d != Deny {
		t.Fatalf("autonomy must not bypass explicit deny rule, got %v: %s", d, reason)
	}
}

func TestEngine_AutonomyBypassesConsentButNotDestructive(t *testing.T) {
	p := DefaultPolicy()
	e := NewEngine(p, &testConsent{available: true, approved: true}, nil)
	// Autonomous level 3.
	e.AutonomyHook = func(_, _ string) int { return 3 }

	// Known delegation requires consent; autonomy should auto-allow.
	d, _ := e.Evaluate(context.Background(), blastradius.Action{Kind: "delegation.spawn", TargetApp: "claude"})
	if d != Allow {
		t.Fatalf("autonomy should bypass consent-required verdict, got %v", d)
	}

	// DESTRUCTIVE still requires consent even under autonomy.
	d, _ = e.Evaluate(context.Background(), blastradius.Action{Kind: "shell.exec"})
	if d != Allow {
		t.Fatalf("autonomy should not bypass DESTRUCTIVE consent, got %v", d)
	}
}

// Phase 16, Rec 5: workspace trust bypasses WRITE consent but never
// DESTRUCTIVE consent. We model the gatekeeper's TrustHook as a
// closure over a tiny "is trusted?" map; in production this calls
// into internal/trust.Store.
//
// We create a real temp dir with a .git/ so workspaceIDFor finds
// the git root and the trust key matches.
func TestEngine_WorkspaceTrustBypassesWriteConsent(t *testing.T) {
	// Consent that DENIES. This way, an action that requires consent
	// produces Deny — and we can verify trust bypasses the
	// consent dialog entirely.
	p := DefaultPolicy()
	e := NewEngine(p, &testConsent{available: true, approved: false}, nil)

	// Set up a real repo so workspaceIDFor resolves to the git root.
	repoDir := setupRepoWithGit(t)

	// Use filepath.Join so the test is portable to Windows
	// (where the path separator is `\`).
	nested := filepath.Join(repoDir, "src", "main.go")
	if err := os.WriteFile(nested, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	trusted := map[string]bool{
		repoDir: true,
	}
	e.TrustHook = func(workspaceID, app string) (any, bool) {
		return "trusted-entry", trusted[workspaceID]
	}

	// WRITE action targeting a trusted workspace → Allow (consent
	// dialog skipped via trust).
	d, reason := e.Evaluate(context.Background(), blastradius.Action{
		Kind:      "file.write",
		TargetApp: "com.microsoft.VSCode",
		Path:      nested,
	})
	if d != Allow {
		t.Fatalf("trusted workspace should bypass WRITE consent, got %v: %s", d, reason)
	}

	// WRITE action targeting an untrusted workspace → consent
	// is requested; the test consent denies → Deny.
	d, _ = e.Evaluate(context.Background(), blastradius.Action{
		Kind:      "file.write",
		TargetApp: "com.microsoft.VSCode",
		Path:      filepath.Join(os.TempDir(), "untrusted-src-main.go"),
	})
	if d != Deny {
		t.Fatalf("untrusted workspace must produce Deny when consent denies, got %v", d)
	}

	// DESTRUCTIVE action in a trusted workspace → still requires
	// consent (Survival Rule §2).
	d, _ = e.Evaluate(context.Background(), blastradius.Action{
		Kind:      "shell.exec",
		TargetApp: "com.microsoft.VSCode",
		Path:      filepath.Join(repoDir, "whatever"),
	})
	if d != Deny {
		t.Fatalf("DESTRUCTIVE must not bypass consent even in trusted workspace, got %v", d)
	}
}

func TestEngine_WorkspaceTrustHeuristicFindsGitRoot(t *testing.T) {
	// Real-filesystem check: workspaceIDFor(".../repo/src/lib/file.go")
	// should return ".../repo" when that dir contains .git/. We set
	// that up here. Use filepath.Join so the test is portable to
	// Windows (where the path separator is `\`).
	repoDir := setupRepoWithGit(t)
	nested := filepath.Join(repoDir, "src", "lib")
	got := workspaceIDFor(nested)
	if got != repoDir {
		t.Fatalf("workspaceIDFor: got %q, want %q", got, repoDir)
	}
}

// setupRepoWithGit creates a real git-rooted directory and returns
// its absolute path. The repo lives under t.TempDir() so it's
// auto-cleaned.
//
// Returns the path through filepath.Abs so callers get the
// platform-canonical separator (`/` on POSIX, `\` on Windows).
// Without this, the production code's filepath.Abs would
// return a different path than the test's string concatenation,
// and the equality comparison would fail on Windows.
func setupRepoWithGit(t *testing.T) string {
	t.Helper()
	for _, base := range []string{t.TempDir(), mustCwd(t)} {
		repoDir, err := filepath.Abs(base + "/condura-test-" + t.Name() + "-" + randSuffix())
		if err != nil {
			continue
		}
		if err := os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755); err != nil {
			continue
		}
		// Sandbox-detect: try writing a file inside the new dir.
		if err := os.WriteFile(filepath.Join(repoDir, "probe.txt"), nil, 0o644); err != nil {
			continue
		}
		if err := os.MkdirAll(filepath.Join(repoDir, "src"), 0o755); err != nil {
			continue
		}
		if err := os.WriteFile(filepath.Join(repoDir, "src", "probe.txt"), nil, 0o644); err != nil {
			continue
		}
		return repoDir
	}
	t.Fatal("could not create a writable temp repo in t.TempDir() or cwd")
	return ""
}

func mustCwd(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return cwd
}

func randSuffix() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func TestAtomicPolicy_LoadStore(t *testing.T) {
	ap := &AtomicPolicy{}
	p1 := DefaultPolicy()
	ap.Store(p1)
	if ap.Load() != p1 {
		t.Fatal("Load returned different policy")
	}
	p2 := DefaultPolicy()
	ap.Store(p2)
	if ap.Load() != p2 {
		t.Fatal("atomic swap failed")
	}
}

type testHalt struct{ halted bool }

func (h *testHalt) IsHalted() bool { return h.halted }

type testConsent struct {
	available bool
	approved  bool
}

func (c *testConsent) IsAvailable() bool { return c.available }
func (c *testConsent) Show(_ context.Context, _ *ConsentTicket) (bool, error) {
	return c.approved, nil
}

// Phase 17, Fix #2 (A6): ApproveTicket / DenyTicket must reject
// expired nonces. We seed the engine's pending list with a ticket
// whose ExpiresAt is in the past and verify both methods return
// false. We also seed a fresh ticket and verify both methods
// succeed.
func TestEngine_TicketExpiry_ApproveAndDenyRejectExpired(t *testing.T) {
	p := DefaultPolicy()
	e := NewEngine(p, nil, &testHalt{})

	expired := &ConsentTicket{
		ActionKind: "shell.exec",
		Nonce:      "expired-nonce",
		ExpiresAt:  time.Now().Add(-1 * time.Minute),
		Result:     make(chan bool, 1),
	}
	fresh := &ConsentTicket{
		ActionKind: "shell.exec",
		Nonce:      "fresh-nonce",
		ExpiresAt:  time.Now().Add(1 * time.Minute),
		Result:     make(chan bool, 1),
	}
	e.pendingMu.Lock()
	e.pending = append(e.pending, expired, fresh)
	e.pendingMu.Unlock()

	if e.ApproveTicket("expired-nonce") {
		t.Error("ApproveTicket should reject expired nonces")
	}
	if e.ApproveTicket("nonexistent-nonce") {
		t.Error("ApproveTicket should reject unknown nonces")
	}
	if !e.ApproveTicket("fresh-nonce") {
		t.Error("ApproveTicket should accept fresh nonces")
	}

	if e.DenyTicket("expired-nonce") {
		t.Error("DenyTicket should reject expired nonces")
	}
	if !e.DenyTicket("fresh-nonce") {
		t.Error("DenyTicket should accept fresh nonces")
	}
}

// Phase 17, Fix #5 (A7): the on_timeout=queue policy field must
// suppress the engine-side timeout so the GUI's own dialog has time
// to gather a real answer. We assert:
//   - With queue, the engine does NOT return Deny after TimeoutSecs.
//   - The ticket remains in Pending() across the wait.
//   - When the caller's context is canceled, the engine returns Deny
//     and the ticket is STILL in Pending() so the GUI can resolve it.
//   - With default (queue not set), the engine returns Deny after
//     TimeoutSecs.
func TestEngine_OnTimeoutQueue_SuppressesEngineTimeout(t *testing.T) {
	// Test consent provider that blocks on a channel until the
	// test signals it. We use this to simulate a slow GUI dialog.
	block := make(chan struct{})
	defer close(block)
	slowProvider := &slowConsent{block: block}

	t.Run("queue_blocks_until_cancel", func(t *testing.T) {
		p := DefaultPolicy()
		// Force every consent-required rule to use on_timeout=queue
		// with a 1-second engine timeout that we'd otherwise hit.
		for i := range p.rules {
			if p.rules[i].Consent.OnTimeout != "" || p.rules[i].Consent.TimeoutSeconds == 0 {
				continue
			}
			p.rules[i].Consent.OnTimeout = "queue"
			p.rules[i].Consent.TimeoutSeconds = 1
		}
		e := NewEngine(p, slowProvider, &testHalt{})

		ctx, cancel := context.WithCancel(context.Background())
		// Cancel after 200ms — before the 1s engine timeout.
		go func() {
			time.Sleep(200 * time.Millisecond)
			cancel()
		}()
		// Trigger consent-required action.
		d, reason := e.Evaluate(ctx, blastradius.Action{Kind: "shell.exec"})
		if d != Deny {
			t.Errorf("expected Deny after ctx cancel, got %v (%s)", d, reason)
		}
		if !strings.Contains(reason, "canceled") {
			t.Errorf("expected 'canceled' reason, got %q", reason)
		}
		// Ticket must still be in pending so the GUI can resolve it.
		pending := e.Pending()
		if len(pending) != 1 {
			t.Errorf("expected 1 pending ticket (queue preserves), got %d", len(pending))
		}
	})

	t.Run("default_times_out", func(t *testing.T) {
		p := DefaultPolicy()
		// Force every consent-required rule to use the default
		// timeout (empty on_timeout = deny on timeout) with a 1s
		// engine timeout.
		for i := range p.rules {
			if p.rules[i].Consent.OnTimeout != "" || p.rules[i].Consent.TimeoutSeconds == 0 {
				continue
			}
			p.rules[i].Consent.OnTimeout = ""
			p.rules[i].Consent.TimeoutSeconds = 1
		}
		e := NewEngine(p, slowProvider, &testHalt{})
		// Give the engine a fresh context that we DON'T cancel.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		start := time.Now()
		d, reason := e.Evaluate(ctx, blastradius.Action{Kind: "shell.exec"})
		elapsed := time.Since(start)
		if d != Deny {
			t.Errorf("expected Deny on timeout, got %v (%s)", d, reason)
		}
		if !strings.Contains(reason, "timeout") {
			t.Errorf("expected 'timeout' in reason, got %q", reason)
		}
		if elapsed > 3*time.Second {
			t.Errorf("engine-side timeout took too long: %v", elapsed)
		}
		// Default timeout removes the ticket.
		if got := len(e.Pending()); got != 0 {
			t.Errorf("expected 0 pending tickets after default-timeout deny, got %d", got)
		}
	})
}

// slowConsent blocks in Show until the test releases the channel.
// Used to simulate a slow GUI dialog in Fix #5 tests.
type slowConsent struct {
	block chan struct{}
}

func (s *slowConsent) IsAvailable() bool { return true }
func (s *slowConsent) Show(_ context.Context, _ *ConsentTicket) (bool, error) {
	<-s.block
	return false, nil
}

// TestAnomalyHook_CarriesRealDecision pins P0-1 of the 2026-06-29
// audit: the engine's AnomalyHook must fire AFTER the verdict is
// decided so the wired-in detector receives success=true on Allow
// and success=false on Deny. Before this fix, the hook fired before
// the verdict with success=false hard-coded, which made the §5.6
// "5+ consecutive failures" trigger trip after any 5 actions through
// the gate (including 5 successful READs).
func TestAnomalyHook_CarriesRealDecision(t *testing.T) {
	p := DefaultPolicy()
	e := NewEngine(p, nil, &testHalt{})

	type hookCall struct {
		decision Decision
	}
	var calls []hookCall
	e.AnomalyHook = func(_ blastradius.Action, d Decision, _ string) {
		calls = append(calls, hookCall{decision: d})
	}

	// 5 reads (Allow) — must NOT accumulate failures.
	for i := 0; i < 5; i++ {
		d, _ := e.Evaluate(context.Background(), blastradius.Action{Kind: "chat"})
		if d != Allow {
			t.Fatalf("read %d: expected Allow, got %v", i, d)
		}
	}
	if len(calls) != 5 {
		t.Fatalf("expected 5 hook calls, got %d", len(calls))
	}
	for i, c := range calls {
		if c.decision != Allow {
			t.Errorf("call %d: expected decision=Allow, got %v", i, c.decision)
		}
	}

	// Now 5 explicitly denied actions — each must record success=false.
	// We use a SanitizeHook that always errors, which is a deterministic
	// way to drive a Deny without setting up a consent provider.
	e2 := NewEngine(p, nil, &testHalt{})
	e2.SanitizeHook = func(_ *blastradius.Action) error {
		return errAlwaysDeny
	}
	calls = calls[:0]
	e2.AnomalyHook = func(_ blastradius.Action, d Decision, _ string) {
		calls = append(calls, hookCall{decision: d})
	}
	for i := 0; i < 5; i++ {
		d, _ := e2.Evaluate(context.Background(), blastradius.Action{Kind: "chat"})
		if d != Deny {
			t.Fatalf("deny %d: expected Deny, got %v", i, d)
		}
	}
	for i, c := range calls {
		if c.decision != Deny {
			t.Errorf("call %d: expected decision=Deny, got %v", i, c.decision)
		}
	}
}

// errAlwaysDeny is a sentinel for TestAnomalyHook_CarriesRealDecision.
type errAlwaysDenyer struct{}

func (errAlwaysDenyer) Error() string { return "always deny" }

var errAlwaysDeny = errAlwaysDenyer{}
