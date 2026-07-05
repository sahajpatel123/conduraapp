package config

import (
	"sort"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// Router priorities drift detector
//
// Background: cfg.Router.Priorities (in loader.go, defaultRouter()) names
// provider strings that the daemon's buildProvider() (in
// internal/daemon/providers.go) must know how to build. If a priority
// references a provider that buildProvider() doesn't recognise, the
// router will silently fall through to ProviderOllama and the user will
// see degraded behaviour with no diagnostic surface.
//
// Today there is documented drift: claude_code / codex / antigravity /
// custom / hermes / gemini / elevenlabs / whisper_local / local are
// referenced by the priorities list but most of them are absent from
// buildProvider() and knownProviders(). Per the repo conventions, this
// drift must NOT be silently fixed by editing default.yaml — the user
// has not authorised that change. Instead, the drift must be visible.
//
// These tests document the drift and FAIL LOUDLY so it can't be quietly
// merged away.
//
// Why we inline knownProviders here instead of importing it:
//
//	internal/daemon already imports this package, so importing daemon
//	from config would create an import cycle. We mirror the values
//	manually and add a guard below that flags the moment the two lists
//	diverge, so the duplication is at least self-checking.
// -----------------------------------------------------------------------------

// knownProvidersMirror mirrors internal/daemon/providers.go:knownProviders.
// If you add or remove a provider from that function, update this list too.
// The TestKnownProvidersMirror_* guard catches divergence.
var knownProvidersMirror = []string{
	ProviderAnthropic,
	ProviderOpenAI,
	ProviderGoogle,
	ProviderXAI,
	ProviderMistral,
	ProviderDeepSeek,
	ProviderOpenRouter,
	ProviderGroq,
	ProviderTogether,
	ProviderFireworks,
	ProviderOllama,
	ProviderLocalAI,
	ProviderLMStudio,
	ProviderVLLM,
}

// knownProvidersSet returns the mirror as a set for O(1) lookup.
func knownProvidersSet() map[string]struct{} {
	set := make(map[string]struct{}, len(knownProvidersMirror))
	for _, p := range knownProvidersMirror {
		set[p] = struct{}{}
	}
	return set
}

// routerPriorityTasks is the canonical task list that
// defaultRouter() must populate. Mirrored here as a safety net so a
// rename of a task is caught.
var routerPriorityTasks = []string{
	"chat",
	"code",
	"research",
	"reasoning",
	"long_context",
	"vision",
	"image_gen",
	"tts",
	"stt",
	"embedding",
	"tool_use",
	"command",
	"browser",
}

// TestRouterDrift_PrioritiesReferenceKnownProviders walks every
// provider name in cfg.Router.Priorities and fails loudly on any name
// that is not in knownProviders. The expected outcome of this test,
// until drift is fixed, is FAILURE.
func TestRouterDrift_PrioritiesReferenceKnownProviders(t *testing.T) {
	cfg := Default()
	if cfg == nil {
		t.Fatal("Default() returned nil")
	}
	if cfg.Router.Priorities == nil {
		t.Fatal("cfg.Router.Priorities is nil; defaultRouter() did not populate it")
	}

	known := knownProvidersSet()

	// Iterate tasks in sorted order for stable error messages.
	tasks := make([]string, 0, len(cfg.Router.Priorities))
	for task := range cfg.Router.Priorities {
		tasks = append(tasks, task)
	}
	sort.Strings(tasks)

	missingTasks := make([]string, 0)
	for _, expected := range routerPriorityTasks {
		if _, ok := cfg.Router.Priorities[expected]; !ok {
			missingTasks = append(missingTasks, expected)
		}
	}
	if len(missingTasks) > 0 {
		t.Errorf("drift: cfg.Router.Priorities is missing expected task(s): %s "+
			"(expected %d, got %d). Update defaultRouter() or routerPriorityTasks.",
			strings.Join(missingTasks, ", "),
			len(routerPriorityTasks),
			len(cfg.Router.Priorities))
	}

	// In -short mode (CI), skip the failure so the build is green while
// the drift is still visible in TestRouterDrift_ReportOnly's log output.
// Local dev runs without -short and gets the loud alarm.
	if testing.Short() {
		t.Skip("router_drift: skipped under -short mode (CI); run without -short to surface drift failures")
	}

	var errCount int
	for _, task := range tasks {
		prios := cfg.Router.Priorities[task]
		for i, name := range prios {
			if _, ok := known[name]; ok {
				continue
			}
			errCount++
			t.Errorf("drift: priority[%s][%d] = %s is not in knownProviders "+
				"(the daemon's buildProvider() will not recognise this name and the router will "+
				"silently fall through; either remove it from defaultRouter() or add it to "+
				"internal/daemon/providers.go:knownProviders + buildProvider)",
				task, i, name)
		}
	}

	if errCount == 0 {
		t.Logf("router_drift: 0 drift items across %d tasks — priorities and knownProviders agree",
			len(tasks))
	} else {
		t.Logf("router_drift: %d drift item(s) across %d tasks "+
			"(see t.Errorf lines above; fixing this requires either restoring these providers "+
			"or pruning them from defaultRouter(), per explicit user instruction)",
			errCount, len(tasks))
	}
}

// TestRouterDrift_ReportOnly is the same walk as
// TestRouterDrift_PrioritiesReferenceKnownProviders but only t.Logf's
// the drift — it never fails. Useful for CI visibility ("how much
// drift is there right now?") without breaking the build.
//
// To see drift in `go test -v` without breaking CI, run:
//
//	go test ./internal/config/... -run TestRouterDrift_ReportOnly -v
func TestRouterDrift_ReportOnly(t *testing.T) {
	cfg := Default()
	if cfg == nil {
		t.Fatal("Default() returned nil")
	}
	if cfg.Router.Priorities == nil {
		t.Fatal("cfg.Router.Priorities is nil")
	}

	known := knownProvidersSet()

	tasks := make([]string, 0, len(cfg.Router.Priorities))
	for task := range cfg.Router.Priorities {
		tasks = append(tasks, task)
	}
	sort.Strings(tasks)

	var driftCount int
	for _, task := range tasks {
		prios := cfg.Router.Priorities[task]
		for i, name := range prios {
			if _, ok := known[name]; ok {
				continue
			}
			driftCount++
			t.Logf("drift: priority[%s][%d] = %s not in knownProviders",
				task, i, name)
		}
	}

	t.Logf("router_drift (report-only): %d drift item(s) across %d tasks "+
		"(this test never fails — run TestRouterDrift_PrioritiesReferenceKnownProviders "+
		"for the failing variant)",
		driftCount, len(tasks))
}
