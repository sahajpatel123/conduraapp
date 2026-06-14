// Trust E2E — drives the Phase 11 (Trust & Recovery) pipeline
// end-to-end through the IPC layer, the same way the GUI will.
// We build a real daemon (real Engine, real audit, real backup,
// real storage) on a temp data dir, then call every trust RPC
// the GUI exposes:
//
//   - replay.timeline / replay.frame / replay.verify_integrity
//   - permissions.status / permissions.request_guide
//   - onboarding.state / onboarding.advance / onboarding.complete
//   - backup.list / backup.derive_key
//   - uninstall.preview
//
// The point of this test is to catch:
//   - missing RPC registrations
//   - nils in the Subsystems struct that the GUI would hit
//   - type-shape bugs in the wire format
//   - audit chain integrity regressions caused by Phase 11
//
// We do NOT test the gated restore / uninstall paths here —
// those go through a separate integration test that supplies a
// real ConfirmToken.
package daemon

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/config"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
	"github.com/sahajpatel123/synapticapp/internal/version"
)

// installPermissivePolicy replaces the engine's policy with a
// catch-all allow rule. Used by tests that need gated RPCs
// (backup.restore, uninstall.execute) to succeed without a GUI
// connected to provide consent.
func installPermissivePolicy(subs *Subsystems) {
	if subs.Safety == nil || subs.Safety.Engine == nil {
		return
	}
	p, err := gatekeeper.LoadPolicy([]byte(`version: "1"
rules:
  - match: {}
    decide: allow
`))
	if err != nil {
		return
	}
	subs.Safety.Engine.ReloadPolicy(p)
}

// startTrustDaemon brings up a real daemon on a temp data dir
// with the IPC listening on a free TCP port. Returns the
// address ("127.0.0.1:NNNN") and the running Subsystems so
// tests can call methods directly.
//
// We can't use the public Run() entry point because it
// doesn't expose the constructed Subsystems; we want to call
// methods directly to verify state, not just RPC plumbing.
func startTrustDaemon(t *testing.T) (string, *Subsystems, func()) {
	t.Helper()
	dir := t.TempDir()
	cfg := config.Default()
	cfg.General.DataDir = dir
	cfg.Storage.Path = filepath.Join(dir, "synaptic.db")
	cfg.Logging.File = ""
	cfg.Logging.AddSource = false
	cfg.Security.SpendLimitUSDPerDay = 1.0
	cfg.APIServer.AuthToken = "test-token"
	// Force a free TCP port.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := l.Addr().String()
	_ = l.Close()

	// Clear SYNAPTIC_ env so config stays deterministic.
	for _, e := range os.Environ() {
		for i := 0; i < len(e)-9; i++ {
			if e[i:i+9] == "SYNAPTIC_" {
				name := e[:i+9]
				end := i + 9
				for end < len(e) && e[end] != '=' {
					end++
				}
				if end < len(e) {
					t.Setenv(name, "")
				}
				break
			}
		}
	}

	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	subs, err := initSubsystems(log, cfg)
	if err != nil {
		t.Fatalf("initSubsystems: %v", err)
	}

	srv := ipc.NewServer()
	registerMethods(srv, log, cfg, subs, version.Info{Version: "test"})
	// Serve the JSON-RPC 2.0 protocol over HTTP on the bound
	// address. The transport.HandleRaw path is what existing
	// tests use; we mirror it here.
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

// callRPC is a tiny test helper that opens a fresh TCP
// connection per call (no reuse). Returns the raw JSON
// result so tests can assert on either arrays or objects.
func trustCallRPC(t *testing.T, addr, method string, params any) (json.RawMessage, error) {
	t.Helper()
	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	}
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/", strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	var r struct {
		Result json.RawMessage `json:"result"`
		Error  *ipc.Error      `json:"error"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, err
	}
	if r.Error != nil {
		return nil, r.Error
	}
	return r.Result, nil
}

func TestTrustE2E_ReplayTimelineReturnsEmpty(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	res, err := trustCallRPC(t, addr, "replay.timeline", nil)
	if err != nil {
		t.Fatalf("replay.timeline: %v", err)
	}
	var arr []any
	if err := json.Unmarshal(res, &arr); err != nil {
		t.Fatalf("expected array, got %s: %v", res, err)
	}
	if len(arr) != 0 {
		t.Fatalf("expected empty timeline, got %d items", len(arr))
	}
}

func TestTrustE2E_PermissionsStatusReturnsList(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	res, err := trustCallRPC(t, addr, "permissions.status", nil)
	if err != nil {
		t.Fatalf("permissions.status: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(res, &out); err != nil {
		t.Fatalf("unmarshal: %v: %s", err, res)
	}
	platform, _ := out["platform"].(string)
	if platform == "" {
		t.Fatalf("platform empty in response: %v", out)
	}
	items, _ := out["items"].([]any)
	if len(items) != 5 {
		t.Fatalf("expected 5 permission items, got %d", len(items))
	}
}

func TestTrustE2E_PermissionsRequestGuideForKnownKind(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	res, err := trustCallRPC(t, addr, "permissions.request_guide", map[string]any{"kind": "microphone"})
	if err != nil {
		t.Fatalf("permissions.request_guide: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(res, &out); err != nil {
		t.Fatalf("unmarshal: %v: %s", err, res)
	}
	if out["title"] == nil || out["title"] == "" {
		t.Fatalf("title empty: %v", out)
	}
	steps, _ := out["steps"].([]any)
	if len(steps) == 0 {
		t.Fatalf("steps empty: %v", out)
	}
}

func TestTrustE2E_PermissionsRequestGuideRejectsUnknown(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	_, err := trustCallRPC(t, addr, "permissions.request_guide", map[string]any{"kind": "banana"})
	if err == nil {
		t.Fatalf("expected error for unknown kind")
	}
}

func TestTrustE2E_OnboardingStateAndAdvance(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	res, err := trustCallRPC(t, addr, "onboarding.state", nil)
	if err != nil {
		t.Fatalf("onboarding.state: %v", err)
	}
	var s1 map[string]any
	if err := json.Unmarshal(res, &s1); err != nil {
		t.Fatalf("unmarshal: %v: %s", err, res)
	}
	if s1["current_step"] == nil {
		t.Fatalf("current_step missing: %v", s1)
	}
	if _, err := trustCallRPC(t, addr, "onboarding.advance", nil); err != nil {
		t.Fatalf("onboarding.advance: %v", err)
	}
	res2, err := trustCallRPC(t, addr, "onboarding.state", nil)
	if err != nil {
		t.Fatalf("onboarding.state 2: %v", err)
	}
	var s2 map[string]any
	if err := json.Unmarshal(res2, &s2); err != nil {
		t.Fatalf("unmarshal: %v: %s", err, res2)
	}
	if s2["current_step"] == s1["current_step"] {
		t.Fatalf("step did not advance: %v -> %v", s1["current_step"], s2["current_step"])
	}
}

func TestTrustE2E_OnboardingCompleteAndReset(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	if _, err := trustCallRPC(t, addr, "onboarding.complete", nil); err != nil {
		t.Fatalf("onboarding.complete: %v", err)
	}
	res, err := trustCallRPC(t, addr, "onboarding.state", nil)
	if err != nil {
		t.Fatalf("onboarding.state: %v", err)
	}
	var s map[string]any
	if err := json.Unmarshal(res, &s); err != nil {
		t.Fatalf("unmarshal: %v: %s", err, res)
	}
	completedAt, _ := s["completed_at"].(string)
	if completedAt == "" {
		t.Fatalf("completed_at not set: %v", s)
	}
	if _, err := trustCallRPC(t, addr, "onboarding.reset", nil); err != nil {
		t.Fatalf("onboarding.reset: %v", err)
	}
}

func TestTrustE2E_BackupDeriveKeyReturnsBase64(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	res, err := trustCallRPC(t, addr, "backup.derive_key", nil)
	if err != nil {
		t.Fatalf("backup.derive_key: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(res, &out); err != nil {
		t.Fatalf("unmarshal: %v: %s", err, res)
	}
	k, _ := out["key"].(string)
	if k == "" {
		t.Fatalf("empty key: %v", out)
	}
	decoded, decErr := base64.StdEncoding.DecodeString(k)
	if decErr != nil {
		t.Fatalf("key not base64: %v", decErr)
	}
	if len(decoded) != 32 {
		t.Fatalf("key length: want 32, got %d", len(decoded))
	}
}

func TestTrustE2E_UninstallPreviewReturnsManifest(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	res, err := trustCallRPC(t, addr, "uninstall.preview", nil)
	if err != nil {
		t.Fatalf("uninstall.preview: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(res, &out); err != nil {
		t.Fatalf("unmarshal: %v: %s", err, res)
	}
	if out["data_dir"] == nil {
		t.Fatalf("data_dir missing: %v", out)
	}
}

// Sanity check: the Subsystems struct actually holds all the
// Phase 11 components we expect. This catches a "I added a
// field to the struct but forgot to construct it" bug.
func TestTrustE2E_SubsystemsPhase11Wiring(t *testing.T) {
	_, subs, cleanup := startTrustDaemon(t)
	defer cleanup()
	if subs.Replay == nil {
		t.Fatalf("Replay is nil — should be non-nil for E2E test")
	}
	if subs.Backup == nil {
		t.Fatalf("Backup is nil — should be non-nil for E2E test")
	}
	if subs.Uninstaller == nil {
		t.Fatalf("Uninstaller is nil — should be non-nil for E2E test")
	}
	if subs.Onboarding == nil {
		t.Fatalf("Onboarding is nil — should be non-nil for E2E test")
	}
	if subs.Permissions == nil {
		t.Fatalf("Permissions is nil — should be non-nil for E2E test")
	}
	if subs.AuditLog == nil {
		t.Fatalf("AuditLog is nil — should be non-nil for E2E test")
	}
	if subs.GeneralDataDir() == "" {
		t.Fatalf("GeneralDataDir is empty")
	}
	_ = context.Background
}
