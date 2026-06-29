package gatekeeper

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/blastradius"
	"github.com/sahajpatel123/conduraapp/internal/trust"
)

// TestE2E_TrustFlow_BypassesTerminalAutoAllow is a Tier-3
// reproduction of the highest-severity finding from the Phase 14I
// audit: defaults.yaml line 19 auto-allows WRITE for Terminal/etc.
// even for untrusted targets. Phase 16 added a trust hook that
// short-circuits the WRITE consent for trusted workspaces. This
// test pins the behavior in both directions:
//
//   - Untrusted workspace → consent required (denied by test
//     consent → Deny decision)
//   - Trusted workspace → Allow (consent bypassed)
//   - DESTRUCTIVE in trusted workspace → consent still required
//     (Survival Rule §2)
//
// Phase 17 Fix #3 (B2): removed the auto-allow rule that previously
// exempted WRITE for target_app=Code/VS Code/Cursor/Terminal/Finder.
// The rule was too broad — combined with `computeruse.type` being
// WRITE-classified, it let the agent type arbitrary shell commands
// into Terminal with no consent. The trust hook is the correct
// bypass for trusted workspaces; bundle-ID style matches (e.g.
// "com.apple.Terminal") never hit the old rule anyway, so the
// untrusted-workspace path is unchanged in behavior.
func TestE2E_TrustFlow_BypassesTerminalAutoAllow(t *testing.T) {
	dir := t.TempDir()
	store, err := trust.NewStore(filepath.Join(dir, "trust.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	p := DefaultPolicy()
	e := NewEngine(p, &testConsent{approved: false}, nil)
	e.TrustHook = func(wsID, app string) (any, bool) {
		entry := store.Lookup(wsID, app)
		if entry == nil {
			return nil, false
		}
		return entry, true
	}

	// 1. Untrusted workspace → consent required. Test consent
	//    denies → Deny.
	d, _ := e.Evaluate(context.Background(), blastradius.Action{
		Kind:      "computeruse.type",
		TargetApp: "com.apple.Terminal",
		Body:      "ls",
		Path:      filepath.Join(dir, "untrusted", "foo.go"),
	})
	if d != Deny {
		t.Fatalf("untrusted WRITE should require consent, got %v", d)
	}

	// 2. Trusted workspace → Allow (consent bypassed).
	repoDir := setupRepoWithGit(t)
	if _, err := store.Grant(repoDir, "TestRepo", trust.DefaultScope); err != nil {
		t.Fatal(err)
	}
	d, reason := e.Evaluate(context.Background(), blastradius.Action{
		Kind:      "file.write",
		TargetApp: "com.microsoft.VSCode",
		Path:      filepath.Join(repoDir, "src", "main.go"),
	})
	if d != Allow {
		t.Fatalf("trusted WRITE should bypass consent, got %v: %s", d, reason)
	}

	// 3. DESTRUCTIVE in trusted workspace → still requires consent.
	d, _ = e.Evaluate(context.Background(), blastradius.Action{
		Kind:      "shell.exec",
		TargetApp: "com.apple.Terminal",
		Body:      "rm -rf /",
		Path:      filepath.Join(repoDir, "whatever"),
	})
	if d != Deny {
		t.Fatalf("DESTRUCTIVE must not bypass consent, got %v", d)
	}
}
