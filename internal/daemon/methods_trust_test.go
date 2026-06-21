package daemon

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/ipc"
	"github.com/sahajpatel123/synapticapp/internal/trust"
)

// trustListLen is a small helper because the JSON decoder can
// return either []any or []map[string]any depending on context.
// We just need the length.
func trustListLen(t *testing.T, v any) int {
	t.Helper()
	switch x := v.(type) {
	case []map[string]any:
		return len(x)
	case []any:
		return len(x)
	default:
		t.Fatalf("unexpected list type: %T", v)
		return -1
	}
}

func trustListAt(t *testing.T, v any, i int) map[string]any {
	t.Helper()
	switch x := v.(type) {
	case []map[string]any:
		return x[i]
	case []any:
		return x[i].(map[string]any)
	default:
		t.Fatalf("unexpected list type: %T", v)
		return nil
	}
}

// Phase 16, Rec 5: end-to-end test of the trust.* RPC family.
// Drives a real ipc.Server through every method.
func TestTrustMethods_EndToEnd(t *testing.T) {
	dir := t.TempDir()
	store, err := trust.NewStore(filepath.Join(dir, "trust.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	srv := ipc.NewServer()
	subs := &Subsystems{
		Safety: &SafetyComponents{Trust: store},
	}
	registerTrustMethods(srv, subs)

	// 1. list (empty)
	resp, err := trustRPCCall(t, srv, "trust.list", nil)
	if err != nil {
		t.Fatalf("trust.list (empty): %v", err)
	}
	if trustListLen(t, resp) != 0 {
		t.Fatalf("expected 0 entries, got %d", trustListLen(t, resp))
	}

	// 2. grant
	resp, err = trustRPCCall(t, srv, "trust.grant", json.RawMessage(
		`{"workspace_id":"/path/to/repo","label":"My Repo"}`,
	))
	if err != nil {
		t.Fatalf("trust.grant: %v", err)
	}
	got, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("trust.grant response type: got %T", resp)
	}
	if got["workspace_id"] != "/path/to/repo" {
		t.Errorf("workspace_id: got %v", got["workspace_id"])
	}

	// 3. list (one entry)
	resp, _ = trustRPCCall(t, srv, "trust.list", nil)
	if trustListLen(t, resp) != 1 {
		t.Fatalf("after grant, list should have 1 entry, got %d", trustListLen(t, resp))
	}
	entry := trustListAt(t, resp, 0)
	if entry["label"] != "My Repo" {
		t.Errorf("entry.label: got %v", entry["label"])
	}

	// 4. workspace_id_for (real git repo)
	repoDir := setupTrustRepo(t)
	pathBytes, err := json.Marshal(map[string]any{
		"path": filepath.ToSlash(filepath.Join(repoDir, "src", "lib", "foo.go")),
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err = trustRPCCall(t, srv, "trust.workspace_id_for", pathBytes)
	if err != nil {
		t.Fatalf("trust.workspace_id_for: %v", err)
	}
	if got := resp.(map[string]any)["workspace_id"]; got != repoDir {
		t.Errorf("workspace_id_for: got %v, want %v", got, repoDir)
	}

	// 5. revoke
	resp, _ = trustRPCCall(t, srv, "trust.revoke", json.RawMessage(
		`{"workspace_id":"/path/to/repo"}`,
	))
	if ok, _ := resp.(map[string]any)["ok"].(bool); !ok {
		t.Errorf("revoke should return ok=true, got %v", resp)
	}
	resp, _ = trustRPCCall(t, srv, "trust.list", nil)
	if trustListLen(t, resp) != 0 {
		t.Errorf("after revoke, list should be empty, got %d entries", trustListLen(t, resp))
	}

	// 6. validation: empty workspace_id
	_, err = trustRPCCall(t, srv, "trust.grant", json.RawMessage(
		`{"workspace_id":""}`,
	))
	if err == nil {
		t.Fatal("trust.grant with empty workspace_id should fail")
	}
}

func TestTrustMethods_NotAvailable(t *testing.T) {
	srv := ipc.NewServer()
	subs := &Subsystems{Safety: nil}
	registerTrustMethods(srv, subs)
	for _, method := range []string{
		"trust.list",
		"trust.grant",
		"trust.revoke",
		"trust.workspace_id_for",
	} {
		if _, err := trustRPCCall(t, srv, method, nil); err == nil {
			t.Errorf("%s should return error when trust is nil", method)
		}
	}
}

// trustRPCCall invokes a JSON-RPC method on the server. Mirrors
// the pattern in methods_account_test.go so the trust tests use
// the same harness as the account tests.
func trustRPCCall(t *testing.T, srv *ipc.Server, method string, params json.RawMessage) (any, error) {
	t.Helper()
	resp, err := srv.Handle(context.Background(), &ipc.Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      json.RawMessage("1"),
	})
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Result, nil
}

// setupTrustRepo creates a real git-rooted directory at
// repoDir/.git and returns its absolute path (filepath.Abs so the
// path uses the platform-canonical separator). Mirrors the sandbox
// workaround in gatekeeper/e2e_test.go.
func setupTrustRepo(t *testing.T) string {
	t.Helper()
	for _, base := range []string{t.TempDir(), trustMustGetCwd(t)} {
		repoDir, err := filepath.Abs(filepath.Join(base, "trust-test-"+trustRandSuffix()+"-repo"))
		if err != nil {
			continue
		}
		if err := os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755); err != nil {
			continue
		}
		if err := os.WriteFile(filepath.Join(repoDir, "probe.txt"), nil, 0o644); err != nil {
			continue
		}
		if err := os.MkdirAll(filepath.Join(repoDir, "src", "lib"), 0o755); err != nil {
			continue
		}
		if err := os.WriteFile(filepath.Join(repoDir, "src", "lib", "probe.txt"), nil, 0o644); err != nil {
			continue
		}
		return repoDir
	}
	t.Fatal("could not create writable temp repo")
	return ""
}

func trustMustGetCwd(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return cwd
}

func trustRandSuffix() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "fallback"
	}
	return hex.EncodeToString(b[:])
}
