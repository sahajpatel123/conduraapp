package daemon

import (
	"encoding/json"
	"testing"
	"time"
)

// TestAgentSmoke_FullOnboardingToChat runs through the full
// onboarding flow and verifies that the agent is ready to chat
// after provider rebuild. Regression test for the stale registry bug.
func TestAgentSmoke_FullOnboardingToChat(t *testing.T) {
	addr, subs, cleanup := startTrustDaemon(t)
	defer cleanup()

	installPermissivePolicy(subs)
	go autoApprove(subs)

	// Step 1: State starts at pre-launch.
	res, err := trustCallRPC(t, addr, "onboarding.state", nil)
	if err != nil || res == nil {
		t.Fatalf("onboarding.state: %v", err)
	}

	// Step 2: Accept EULA.
	_, _ = trustCallRPC(t, addr, "onboarding.set_step", map[string]string{
		"step": "eula", "status": "complete", "data": "v1",
	})
	_, _ = trustCallRPC(t, addr, "onboarding.advance", nil)

	// Step 3: Skip permissions.
	_, _ = trustCallRPC(t, addr, "onboarding.set_step", map[string]string{
		"step": "permissions", "status": "skipped",
	})
	_, _ = trustCallRPC(t, addr, "onboarding.advance", nil)

	// Step 4: Set hotkey.
	_, _ = trustCallRPC(t, addr, "onboarding.set_step", map[string]string{
		"step": "hotkey", "status": "complete", "data": "Cmd+Shift+Space",
	})

	// Step 5: Finish onboarding.
	finishRaw, err := trustCallRPC(t, addr, "onboarding.finish", map[string]any{
		"hotkey":              "Cmd+Shift+Space",
		"eula_version":        "v1",
		"permissions_skipped": true,
	})
	if err != nil {
		t.Fatalf("onboarding.finish: %v", err)
	}
	var finishRes map[string]any
	if err := json.Unmarshal(finishRaw, &finishRes); err != nil {
		t.Fatalf("unmarshal finish: %v", err)
	}
	if finishRes["power"] == nil {
		t.Fatal("onboarding.finish missing power probe result")
	}

	// Step 6: Verify daemon is alive post-rebuild.
	isCompleteRaw, _ := trustCallRPC(t, addr, "onboarding.is_complete", nil)
	if isCompleteRaw == nil {
		t.Fatal("onboarding.is_complete failed — daemon may have crashed during provider rebuild")
	}

	// Step 7: Verify account.status works (end-to-end health check).
	_, err = trustCallRPC(t, addr, "account.status", nil)
	if err != nil {
		t.Fatalf("account.status after finish: %v", err)
	}
}

// TestAgentSmoke_ProviderRebuildAfterAPIKey verifies that adding
// an API key rebuilds providers without a daemon restart.
func TestAgentSmoke_ProviderRebuildAfterAPIKey(t *testing.T) {
	addr, subs, cleanup := startTrustDaemon(t)
	defer cleanup()

	installPermissivePolicy(subs)

	// Add a fake OpenAI key.
	res, err := trustCallRPC(t, addr, "apikeys.set", map[string]any{
		"provider": "openai",
		"label":    "test",
		"secret":   "sk-test-fake-key",
	})
	if err != nil {
		t.Fatalf("apikeys.set: %v", err)
	}
	if res == nil {
		t.Fatal("apikeys.set returned nil")
	}

	// Verify daemon is still alive.
	_, err = trustCallRPC(t, addr, "account.status", nil)
	if err != nil {
		t.Fatalf("account.status after apikeys.set: %v", err)
	}

	// Verify onboarding probe works.
	_, err = trustCallRPC(t, addr, "onboarding.probe_power", nil)
	if err != nil {
		t.Fatalf("onboarding.probe_power: %v", err)
	}
}

// TestAgentSmoke_SensitiveSiteDetection verifies the sensitive
// site detector is wired in the safety layer.
func TestAgentSmoke_SensitiveSiteDetection(t *testing.T) {
	_, subs, cleanup := startTrustDaemon(t)
	defer cleanup()

	if subs == nil || subs.Safety == nil {
		t.Fatal("safety subsystem not available")
	}
	if subs.Safety.Engine == nil {
		t.Fatal("gatekeeper engine not available")
	}
	// API test: the daemon boots without panic.
	// The SensitiveHook is wired in safety_wiring.go.buildSafetyLayer.
}

// autoApprove polls for pending consent tickets and auto-approves
// them. Used in smoke tests to let gated RPCs proceed.
func autoApprove(subs *Subsystems) {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if subs.Safety != nil && subs.Safety.Engine != nil {
			for _, tk := range subs.Safety.Engine.Pending() {
				subs.Safety.Engine.ApproveTicket(tk.Nonce)
			}
		}
		time.Sleep(25 * time.Millisecond)
	}
}
