# Tier 3 Workspace Analysis - 2026-06-19

## Executive Verdict

Condura has a substantial, testable Go foundation, but it is not yet an end-to-end operating agent. The daemon, IPC, LLM streaming, persistence, audit chain, consent engine, computer-use pipeline, delegation runner, backup, updater, and desktop shell all exist. Several launch-defining paths are nevertheless disconnected, internally contradictory, or verified only with fakes.

The current state is best described as **strong subsystem implementation with incomplete product integration**. It is suitable for focused backend completion work. It is not accurate to describe the current release as fully operational or public-launch verified.

## System Map

- `cmd/condurad`: standalone daemon entry point.
- `cmd/condura`: JSON-RPC CLI client.
- `cmd/condura-tui`: terminal client.
- `app/web`: Wails desktop host embedding the daemon.
- `app/web/frontend`: Svelte desktop UI.
- `web`: Next.js public site and two magic-link API routes.
- `internal/daemon`: composition root for long-lived subsystems and RPC registration.
- `internal/session`, `internal/stream`, `internal/llm`: chat lifecycle and provider streaming.
- `internal/gatekeeper`, `internal/audit`, `internal/halt`, `internal/anomaly`: safety controls.
- `internal/agent`, `internal/computeruse`: planning and physical action execution.
- `internal/delegation`: sub-agent CLI spawning.
- `internal/storage`, `internal/memory`, `internal/conversation`, `internal/adaptive`: local state.
- `internal/backup`, `internal/replay`, `internal/updater`, `internal/uninstall`: trust and lifecycle.
- `internal/account`, `internal/hub`, `internal/sync`, `internal/reach`: optional network integrations.

## Runtime Reality

### Working and wired

- `daemon.Run` validates configuration, acquires a single-instance lock, initializes subsystems, registers RPC methods, starts IPC listeners, and starts background backup/update services.
- `agent.ask` creates a real session, streams an LLM response, persists conversation turns, recalls memory, applies adaptive context, audits the request, and triggers post-session extraction.
- `cu.action` invokes a real planner -> resolver -> gated computer-use pipeline. The resolver captures replay screenshots and feeds the anomaly detector.
- Consent-required Gatekeeper decisions block on an SSE-backed consent ticket and fail closed on cancellation or timeout.
- Audit events are HMAC chained, halt state is persisted, and update manifests are Ed25519 verified before application.

### Implemented but not integrated end to end

1. The daemon's general `GatedAgentExecutor` wraps `noopAgentExecutor`, which always returns `agent executor not yet wired`. Normal chat sessions do not parse model tool calls or route them to `cu.action` or delegation. Computer use is a separate RPC, not a capability of the ordinary agent session.
2. `delegate.spawn` runs a subprocess and returns its output, but the daemon never calls `Delegation.ActionRequests` afterward and never gates or executes those structured requests. The package comment promises behavior that the runtime does not provide.
3. Default delegation CLI templates for Claude and Codex end in `--model`. When the request omits a model, the subprocess receives a dangling flag. Ollama has an empty command but is still passed to `exec.CommandContext`.
4. Default Gatekeeper policy allows delegation consent only for `claude` and `ollama`; all other advertised agents match the subsequent unconditional `delegation.spawn` deny rule.
5. The Wails app registers only the overlay hotkey. The configured kill-switch hotkey is not registered, so the documented emergency shortcut is not a live GUI path. Halt remains reachable through RPC.
6. The Wails presence orchestrator is created with `capture=nil`; the hotkey toggles the overlay but does not start microphone capture despite the lifecycle documentation.

## Safety and Security Boundaries

### Strengths

- Computer-use actions resolve through `computeruse.GatedExecutor` before backend execution.
- Gatekeeper defaults deny unknown actions by requiring presence and consent.
- Consent absence and timeout fail closed.
- IPC defaults to loopback and supports bearer-token authentication; WebSocket origins are constrained to local origins.
- Secrets and configured sensitive columns use local encryption, and API key listings strip secret values.
- Updates verify both an Ed25519 manifest signature and artifact SHA-256.

### Risks and contradictions

- The claim that sub-agents have no direct filesystem/network/terminal access is not enforceable: the runner launches user-installed coding CLIs directly with the daemon user's privileges. Gatekeeping the spawn is not capability sandboxing.
- `APIServerConfig.AllowedOrigins` is declared but not wired into `ServerTransport`; runtime origin policy is hard-coded.
- Release CI uploads an unsigned manifest under the signed filename when `UPDATE_SIGNING_KEY` is absent. Clients correctly reject it, but the workflow can publish a broken update channel while appearing successful.
- The default autonomy hook is hard-coded rather than reading the configured autonomy matrix. Finder, Chrome, and VS Code are treated as autonomous for non-destructive consent-required actions.
- The public site's magic-link backend depends on optional packages not declared in `web/package.json`. Production without KV fails closed, but the build only emits unresolved-module warnings and no API tests exist.

## Verification Results

Executed on macOS in the current workspace:

- `go test ./...`: pass across 62 listed packages.
- `go test -race -count=1 -timeout=300s ./...`: pass; deprecated macOS Process Manager API warnings in ORAX/mac-cua.
- `go vet ./...`: pass.
- `golangci-lint run --timeout=5m`: pass, 0 issues.
- `go build ./cmd/condurad ./cmd/condura ./cmd/condura-tui`: pass.
- `go test ./...` in `app/web`: pass, but no tests exist.
- Wails Svelte `npm run check`: 0 errors, 9 warnings.
- Wails Svelte `npm test`: fail because no test files exist.
- Wails Svelte `npm run build`: pass with accessibility/state warnings.
- Next.js `npm run build`: pass with unresolved optional-module warnings for `@vercel/kv` and `resend`.
- Next.js `npm run lint`: fail with 9 errors and 5 warnings.

Critical package statement coverage:

| Package | Coverage |
|---|---:|
| `internal/gatekeeper` | 64.3% |
| `internal/agent` | 75.1% |
| `internal/computeruse` | 76.7% |
| `internal/delegation` | 47.3% |
| `internal/audit` | 78.8% |
| `internal/anomaly` | 76.9% |
| `internal/ipc` | 78.4% |
| `internal/daemon` | 45.0% |

No measured critical package meets the mission's 80% target. CI's coverage step uses `set +e` and only prints coverage, so it does not enforce its stated threshold. The integration job skips because `test/integration` does not exist. The Phase 15 and on-device verification checklists remain unsigned.

## Documentation Drift

- `docs/guides/ai-onboarding.md` still describes a Phase 0 repository with no code.
- README and release claims are ahead of runtime integration and on-device evidence.
- Synaptic naming remains in protocol URLs, paths, diagnostics, package comments, Hub endpoints, backup names, and wake-word text.
- The repository includes stale root binaries named `synaptic`, `synapticd`, and `synaptic-tui`.

## Recommended Backend Order

1. **Unify the agent runtime:** make one session/tool loop route chat tool calls into gated computer use and delegation. Remove the no-op executor from production composition.
2. **Make delegation truthful:** fix command construction, decide supported agents, process structured action requests, add cancellation/streaming semantics, and explicitly define the security model. Do not claim sandboxing without an actual sandbox.
3. **Wire emergency controls:** register the kill-switch hotkey, cancel active streams/CU/delegations, and prove the behavior in a desktop integration test.
4. **Build vertical integration tests:** daemon boot -> authenticated IPC -> chat -> consent -> action -> audit -> halt, using real process fixtures and OS-specific tests where required.
5. **Turn release gates into gates:** enforce coverage, stop skipping absent integration tests, test Wails/frontend builds in normal CI, and fail release when the signing key is absent.
6. **Complete clean-machine verification:** macOS first, then Windows and Linux. Record evidence instead of marking roadmap phases complete from unit tests.
7. **Consolidate docs and branding:** update onboarding, runtime claims, protocol names, default endpoints, and artifact names after behavior is stable.

## Suggested First Technical Milestone

Build a single macOS vertical slice:

`agent.ask` -> model emits a typed `computer_use` tool call -> Gatekeeper consent -> ORAX/mac-cua execution -> twin-snapshot verification -> audit/replay event -> response returned to the same conversation.

This milestone tests the product's central promise and forces the session, planner, Gatekeeper, computer-use, consent, audit, replay, cancellation, and UI event contracts to agree before more backend breadth is added.
