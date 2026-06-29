// Package onboarding — power probe.
//
// ProbePower detects what AI providers are available on the
// user's machine without requiring any configuration. This runs
// during the Ready screen (step 4) so the user sees what's
// available before they finish onboarding.
package onboarding

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/delegation"
)

// PowerProbe is the result of scanning the local machine for
// available AI backends.
type PowerProbe struct {
	OllamaReachable bool       `json:"ollama_reachable"`
	OllamaModels    []string   `json:"ollama_models"`
	CLIs            []CLIProbe `json:"clis"`
	Recommended     string     `json:"recommended"`
}

// CLIProbe describes whether a specific CLI tool was found.
type CLIProbe struct {
	Name  string `json:"name"`
	Found bool   `json:"found"`
}

// ProbePower scans the local machine for available AI backends.
// It checks:
//  1. Ollama HTTP endpoint (2s timeout)
//  2. CLI tools via exec.LookPath (from delegation config)
//
// The caller is responsible for setting a reasonable context
// deadline (suggested: 3s).
func ProbePower(ctx context.Context) *PowerProbe {
	pp := &PowerProbe{Recommended: "none"}

	pp.probeOllama(ctx)
	pp.probeCLIs()

	if pp.OllamaReachable {
		pp.Recommended = "ollama"
	}
	return pp
}

func (pp *PowerProbe) probeOllama(ctx context.Context) {
	// Try once; if it fails, wait briefly and try again. Phase 15
	// Run #1 observed a race where the first call after daemon
	// startup returned "unreachable" but the second (sub-second
	// later) returned reachable — likely a TCP accept race on
	// Ollama's listener or a Go HTTP client first-use warm-up.
	// One retry with a 250ms back-off handles this without
	// making the common case (already-up) noticeably slower.
	if pp.tryOllamaOnce(ctx) {
		return
	}
	select {
	case <-ctx.Done():
		return
	case <-time.After(250 * time.Millisecond):
	}
	_ = pp.tryOllamaOnce(ctx)
}

// tryOllamaOnce attempts one Ollama HTTP probe. Returns true on
// success (OllamaReachable set, models populated).
func (pp *PowerProbe) tryOllamaOnce(ctx context.Context) bool {
	// Per-attempt timeout is 1s, not 2s, so the full retry
	// (1s + 250ms back-off + 1s = 2.25s) fits inside the 3s
	// parent context from ProbePowerWithTimeout.
	reqCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx,
		http.MethodGet, "http://127.0.0.1:11434/api/tags", nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false
	}
	pp.OllamaReachable = true

	var body struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return true // status was OK, count the probe as successful
	}
	for _, m := range body.Models {
		if m.Name != "" {
			pp.OllamaModels = append(pp.OllamaModels, m.Name)
		}
	}
	return true
}

func (pp *PowerProbe) probeCLIs() {
	for _, agent := range delegation.DefaultAgents() {
		probe := agent.BinaryProbe
		if probe == "" {
			probe = agent.Name
		}
		found := false
		if _, err := exec.LookPath(probe); err == nil {
			found = true
		}
		pp.CLIs = append(pp.CLIs, CLIProbe{
			Name:  agent.Name,
			Found: found,
		})
	}
}

// ProbePowerWithTimeout is a convenience wrapper that adds a
// 3s deadline. Returns nil if the context is already expired.
func ProbePowerWithTimeout(ctx context.Context) *PowerProbe {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return ProbePower(ctx)
}

// OllamaInstallURL returns the OS-appropriate Ollama download page.
func OllamaInstallURL() string {
	return "https://ollama.com/download"
}

// RecommendedOllamaModel is the fallback model name when Ollama is
// reachable but has no models pulled.
const RecommendedOllamaModel = "llama3.2"

// NoModels returns true if Ollama is reachable but has zero models.
func (pp *PowerProbe) NoModels() bool {
	return pp.OllamaReachable && len(pp.OllamaModels) == 0
}

// FirstModel returns the first model name, or the recommended
// fallback if no models are pulled.
func (pp *PowerProbe) FirstModel() string {
	if len(pp.OllamaModels) > 0 {
		return pp.OllamaModels[0]
	}
	return RecommendedOllamaModel
}

// ErrorJSON is a lightweight error envelope used by RPC handlers
// when PowerProbe fails in a way the client should display.
type PowerProbeError struct {
	Message string `json:"error"`
}

func (e PowerProbeError) Error() string {
	return e.Message
}

func newPowerProbeError(format string, args ...interface{}) *PowerProbeError {
	return &PowerProbeError{Message: fmt.Sprintf(format, args...)}
}
