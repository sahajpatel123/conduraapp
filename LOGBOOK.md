# Synaptic — LOGBOOK.md

> **The Master Thinking log.**
> Every AI model that works on Synaptic MUST read this file before starting and MUST append an entry when finishing.
> This file is append-only. Never delete or rewrite past entries. If you need to correct something, add a new entry that references the old one.

---

## How to Use This Log

### Before You Start a Session
1. Read `CLAUDE.md` end-to-end.
2. **Read this file (LOGBOOK.md) from top to bottom** to see the full session history.
3. Note the most recent entry — that's your starting state.
4. Check the "Open questions for next session" sections — answer them if you can.

### When You Finish a Session
Append a new entry at the bottom of this file using the format below. Be specific. Be honest. Be detailed. The next AI (or human) will read this cold, with no other context.

### Entry Format

```markdown
## [<ISO timestamp UTC>] AI Model: <model name>
**Session ID:** <ULID>
**Branch:** <git branch, if any>
**Task:** <one-line summary of what you were asked to do>

### Files created
- `path/to/file.ext` — <purpose>

### Files modified
- `path/to/existing.ext` — <what changed and why>

### Decisions made
- <decision>: <rationale>

### Bugs / issues encountered
- <bug>: <description, stack trace, attempted fix>

### Open questions for next session
- <question>: <context>

### Next steps
- <what should be done next, in priority order>

---
```

---

## Session Log

<!-- ============================================================== -->
<!-- Append new entries BELOW this line. Never modify above.        -->
<!-- ============================================================== -->

## [2026-06-06 16:30 UTC] AI Model: Claude Sonnet 4.5 (opencode/minimax-m3-free, partner-architect)
**Session ID:** 01J9XYZABCDEF0000000000001
**Branch:** main
**Task:** Foundation phase: create the project documentation layer (CLAUDE.md, LOGBOOK.md, architecture docs, ADRs, EULA, README). No code yet.

### Files created
- `CLAUDE.md` — Master Thinking: the entire 36-decision spec, 7 non-negotiables, Selective Perception, Adaptive Engine, safety layer, computer use system, router, delegation bus, memory/skills/MCP, P2P sync, action replay, hotkey/overlay/voice, onboarding, distribution, i18n, build order, repo structure, AI workflow, partner commitment, glossary. ~1,800 lines. Every detail from the planning conversations is captured.
- `LOGBOOK.md` — This file. Append-only session log with format spec.

### Decisions made
- Foundation phase = documentation only. No code yet. User instruction was explicit: "only build the foundation of the project. I do not want you to make anything advanced right now."
- CLAUDE.md structured for AI consumption (tables, code blocks, anchor links) rather than human-only prose. AI agents pick up the project cold from this file.
- The original count "26 decisions" was actually 36 (additional small decisions made during finalization). All listed for completeness.
- The build order now has 37 steps, not 34. Added P2P Sync (step 26.5) and Action Replay (step 26.6) as separate phases.
- Directory structure created with all folders the eventual code will live in.

### Bugs / issues encountered
- None. Pure documentation creation.

### Open questions for next session
- **Visual brand palette** (Decision 12) is "decide later". Should be resolved before UI mockups in Phase 17/18.
- **Marketing copy** (tagline, hero text) needs to be written. Reserved for Phase 31.
- The exact wire format of the JSON-RPC 2.0 protocol methods/events needs to be locked into a TypeScript file (will be done in Phase 4).

### Next steps (priority order)
1. Create `EULA.md` — Synaptic Freeware License v1, with the exact clauses from Decision 36
2. Create `LICENSE` — short proprietary notice
3. Create `README.md` — public-facing, clean
4. Create `CONTRIBUTING.md` — conventions for AI + human contributors
5. Create `SECURITY.md` — vulnerability disclosure policy
6. Create `PRIVACY.md` — privacy policy
7. Create `docs/README.md` — docs index
8. Create `docs/architecture/00-overview.md` through `08-ipc.md` (9 files)
9. Create `docs/adr/0001` through `0005` (5 ADRs)
10. Create `docs/guides/ai-onboarding.md` — detailed onboarding for new AI agents
11. Create `docs/guides/code-style.md` — Go + TypeScript conventions
12. Create empty placeholder files for `docs/user-guide/` and `docs/recipes/` (to be filled in later phases)
13. **Stop. Wait for user to approve the foundation before starting any code.**

### Notes
- The user is the architect and product lead. The AI is the implementer and reviewer. This partnership matters.
- The user was very explicit: "Nothing should be left out or missed. Everything should be retained, not a single or small thing." → Every planning detail is now in CLAUDE.md.
- The user said "I am very serious right now" multiple times. Speed and thoroughness are both critical. Do not cut corners.
- The user wants the foundation to be solid enough that ANY AI model can pick up the project from CLAUDE.md + LOGBOOK.md alone and continue work without confusion.

---

## [2026-06-06 19:00 UTC] AI Model: opencode/minimax-m3-free (partner-architect)
**Session ID:** 01J9XYZABCDEF0000000000002
**Branch:** main
**Task:** Foundation phase continuation: complete all remaining documentation files (security, privacy, architecture deep-dives, ADRs, guides, placeholders). No code yet. Stop after for user approval.

### Files created
- `SECURITY.md` — Vulnerability disclosure policy. Supported versions, reporting flow, PGP key, threat model, specific risks mitigated (10-row table), bug bounty, out-of-scope list, contact email `security@synaptic.app`.
- `PRIVACY.md` — Privacy policy. TL;DR (data on-device, telemetry off by default), data we don't collect, opt-in telemetry, storage locations (encrypted at rest), network calls (only configured LLMs + update server), TCC permissions (microphone, screen recording, camera), children's privacy, GDPR/CCPA rights, third-party services, cookies, retention, contact `privacy@synaptic.app`.
- `docs/README.md` — Index of all documentation. Architecture table, ADR list, guides, user-guide, recipes, API ref, contributing.
- `docs/architecture/00-overview.md` — The conductor pattern. Three-layer mental model (interfaces / orchestration / execution), closed learning loop, "What's Different" comparison table vs Hermes/Antigravity/OpenClaw/Claude Code/mac-cua, performance targets, survival invariants recap.
- `docs/architecture/01-router.md` — Hybrid-with-memory router. TaskSpec schema, 13 sub-task types, 6-step routing algorithm, per-sub-task priority config, trust tiers, deterministic classifier, fallback logic, local-first/offline, streaming/cancellation, spend/rate-limit awareness, status UI.
- `docs/architecture/02-computer-use.md` — 4-tier computer use. Tier 1 (OS CLI/AppleScript), Tier 2 (Accessibility API), Tier 3 (cross-platform MCP), Tier 4 (Vision CUA, last resort). Tier Picker algorithm, the computer-use cycle (9 steps), 3 pinned backends (ORAAX / PyAutoGUI / nut.js+xdotool), integration with Selective Perception, failure modes & recovery, privacy hardpoints, action replay.
- `docs/architecture/03-perception.md` — Selective Perception. The insight (battery = safety = one problem), 6 capture strategies (None/AX-only/Window-rect/Differential/Full/Vision CUA), dirty tracking per-OS, energy budget (4 modes), per-app profiles, PII redaction, pause-on-privacy list (banking/1Password/Signal), the perception pipeline, transparency UI.
- `docs/architecture/04-safety.md` — The safety layer. 5 modules (Strategist/Gatekeeper/Blast-Radius/Anomaly/Audit), 4 blast-radius levels (READ/LOCAL/NETWORK/DESTRUCTIVE), presence tracker, 3-layer kill switch, 7 non-negotiables recap, threat-model-specific defenses (prompt injection, LLM rogue, user mistakes).
- `docs/architecture/05-adaptive.md` — The User-Adaptive Engine. Closed learning loop, the Observer (what it does/doesn't watch), the user model schema (identity/preferences/expertise/habits/social/consent/beliefs), the Honcho-style dialectic (proposer+critic+adjudicator), the predictor, 4 strength levels (off/suggest/anticipate/auto), visibility modes, on-device only, export/edit/delete, open questions.
- `docs/architecture/06-delegation.md` — Delegation Bus. The Delegate interface, 12 LLM providers (table), 8 sub-agent CLIs (table), the Bus itself (Go), task decomposition (DAG), spend monitor, cross-delegate handoff with model isolation, when each sub-agent is picked.
- `docs/architecture/07-memory.md` — 3-layer memory. Working (session, ephemeral), Episodic (cross-session, time-stamped events, 90-day default), Semantic (knowledge graph, confidence-scored, never expires), Audit (separate, append-only), embedding & recall, forgetting (GDPR + user control), privacy boundary (never remembered list), distillation pipeline (local small model), memory-aware prompting, memory schema (SQLite).
- `docs/architecture/08-sync.md` — P2P encrypted sync. Threat model, libp2p protocol stack, Ed25519 device identity, Noise XX handshake, mDNS LAN + DHT WAN discovery, relay options (user's own / Synaptic's / none), CRDTs for conflict-free merging, sync schedule, conflict resolution, revocation, storage on each device, server-side minimal role, trust hierarchy, zero-trust architecture.
- `docs/architecture/09-ipc.md` — JSON-RPC 2.0 IPC. Transports (Unix socket / named pipe / HTTPS / WebSocket), wire format (request/response/error + streaming extension), method namespace (11 namespaces), event namespace (16 event types), JSON Schema validation, authentication (local trusted / remote with Ed25519 + pairing token), the daemon's RPC server, cancellation, backpressure, latency targets, versioning, auto-generated SDKs.
- `docs/adr/0001-go-over-python.md` — ADR for Go over Python/Rust/Node for the core daemon. Rationale: single static binary, cross-compilation, concurrency primitives, ecosystem, performance, type safety, team velocity. Considered alternatives with pros/cons.
- `docs/adr/0002-typescript-for-ui.md` — ADR for TypeScript + React (overlay/web) and Ink (TUI). Why Wails over Electron/Tauri, why Zustand/TanStack Query, why Tailwind, why i18next, accessibility (WCAG 2.1 AA), testing stack.
- `docs/adr/0003-bridge-pattern.md` — ADR for Python subprocess bridges (3 bridges: orax, pyautogui, mcp). Why subprocess over cgo/in-process, why JSON-RPC over stdio, why not gRPC/HTTP, why not rewrite in Go, the bridge architecture, the wire protocol, 20+ bridge methods, lifecycle, security (no network, capability tokens, audit).
- `docs/adr/0004-ce-mcp.md` — ADR for Code-Execution MCP delegation. Anthropic's argument (context efficiency, deterministic control flow, privacy, composability, 70-90% token savings). When code-execution vs function calling. Sandbox per-OS (sandbox-exec / gVisor / AppContainer), tool library, sandboxed shell, the code-execution MCP server, the function-calling adapter, the router.
- `docs/adr/0005-p2p-sync.md` — ADR for P2P sync over central server. Threat model, why not central, why P2P, why libp2p (mature, cross-platform, crypto built-in, discovery, relay, MIT), why not Syncthing/Matrix/custom. The sync protocol stack, the server's role (minimal), user's choices, future (mobile push, cloud relay, snapshot restore).
- `docs/guides/ai-onboarding.md` — Step-by-step guide for AI agents picking up the project. 10 steps from reading source of truth to getting to work. AI workflow rules, hard rules, performance budgets, style quick reference, 37-step build order, partner commitment note.
- `docs/guides/code-style.md` — Full code style for Go, TypeScript/React, Python (bridges). General rules, file headers, imports, naming, errors, context, concurrency, testing, tools per language. Directory layout, commit messages (Conventional Commits), PR description template.
- `docs/user-guide/README.md` — Placeholder. Will be filled in Phase 5+.
- `docs/recipes/README.md` — Placeholder. Will be filled in Phase 6+.
- `docs/api/README.md` — Placeholder. Will be filled in Phase 4 (when protocol is locked).

### Files modified
- None (all new files).

### Decisions made
- **Foundation phase includes architecture deep-dives (10 files, 00-09) and ADRs (5 files)** as the "complete documentation" the user asked for. These complement CLAUDE.md's high-level decisions with implementation-ready detail.
- **Open questions explicitly noted in the docs** (e.g., visual brand palette in CLAUDE.md Decision 12, multi-user device partitioning deferred to v0.2, time-decay for beliefs, exact wire format deferred to Phase 4). No silent hand-waving.
- **The "What's Different" comparison table in 00-overview.md** explicitly lists Hermes Agent, Antigravity CLI, OpenClaw, Claude Code, mac-cua as competitors/comparables, with the unique combination of features Synaptic offers.
- **All 12 LLM providers + 8 CLI sub-agents documented in 06-delegation.md**, with auth options (API key vs OAuth for subscriptions) and the "user's existing subscriptions" use case the user emphasized.
- **Selective Perception and Safety treated as one unified system** in 03-perception.md, per the user's framing in Section 6 of CLAUDE.md.
- **P2P sync threat model and zero-trust architecture** documented in 08-sync.md and ADR-0005.
- **Code-Execution MCP vs function calling** is presented as "both, prefer CE-MCP for complex" per the user's research, with full sandbox design (sandbox-exec / gVisor / AppContainer).
- **The AI onboarding guide (ai-onboarding.md)** makes the AI workflow rules unmissable: "stay within locked decisions", "no half-measures", "test coverage >80% for safety/perception/agent/llm/ipc", "no TODOs", "no silent failures", "performance budgets are non-negotiable".
- **Code style (code-style.md)** is enforceable: gofmt/goimports/golangci-lint for Go, eslint/prettier/tsc for TS, ruff/black/mypy for Python. All test commands, file headers mandatory, public APIs documented.

### Bugs / issues encountered
- None. Pure documentation creation.

### Open questions for next session
- **User approval of the foundation** is the gate. Per the user's instruction: "only build the foundation of the project. I do not want you to make anything advanced right now." → Need explicit "go ahead" before starting Phase 1 (Repo Skeleton).
- **Visual brand palette** (CLAUDE.md Decision 12) still "decide later". Will be needed for any UI work in Phase 1 (Wails app shell) or Phase 4 (overlay). User should provide or delegate.
- **Marketing copy** (tagline, hero text on `synaptic.app`) deferred to Phase 31. Not blocking Phase 1.
- **Exact JSON-RPC wire format** (methods, params, return types in TypeScript) deferred to Phase 4. Architecture doc 09-ipc.md has the namespace and the high-level design; the lock-in happens when we write the SDK.
- **Skills Hub moderation policy** is mentioned in CLAUDE.md ("curated, safety-scanned") but the actual scanner is not specified. Defer to Phase 6 when we build the Hub.
- **Per-OS TCC/permission names** (macOS bundle IDs, Windows capabilities, Linux portal interfaces) are in code-style.md at a high level. The exhaustive per-app list is in `autonomy.yaml` and will be filled in incrementally as users install.
- **Wails v2 vs v3** — locked to v2 in ADR-0002. Re-evaluate if v3 ships stable during build.

### Next steps (priority order)
1. **STOP. Wait for user to approve the foundation before any code.**
2. If user approves: begin **Phase 1: Repo Skeleton** (build steps 1-7 of CLAUDE.md Section 28):
   - 1: Bootstrap (Makefile, go.work, CI, lint, .goreleaser)
   - 2: Storage (SQLite + FTS5 + vec, migrations, crypto)
   - 3: Config (YAML loader, schema, validation)
   - 4: IPC (JSON-RPC 2.0 + WebSocket + auth)
   - 5: LLM clients (12 providers + streaming + caching)
   - 6: API key manager (encrypted, validated, OAuth)
   - 7: Provider failover + spend monitor
3. If user wants changes to the foundation: apply them, then re-approve.
4. **Commit the foundation** (only if user explicitly asks). Not done in this session per AI Workflow rules.

### Notes
- The user said "only build the foundation of the project. I do not want you to make anything advanced right now" in the planning session. This session followed that instruction strictly: 30 markdown files, zero code.
- The user emphasized "Nothing should be left out or missed" — every architecture decision from CLAUDE.md now has a corresponding deep-dive doc, and every locked tech-stack choice has an ADR.
- The foundation is now solid enough that any AI model (or human) can pick up the project from CLAUDE.md + LOGBOOK.md + the architecture docs + the ADRs alone.
- 30 files created in this session:
  - Root: CLAUDE.md, LOGBOOK.md, EULA.md, LICENSE, README.md, CONTRIBUTING.md, SECURITY.md, PRIVACY.md (8 files)
  - docs/: README.md (1 file)
  - docs/architecture/: 00-09 (10 files)
  - docs/adr/: 0001-0005 (5 files)
  - docs/guides/: ai-onboarding.md, code-style.md (2 files)
  - docs/user-guide/: README.md (1 file)
  - docs/recipes/: README.md (1 file)
  - docs/api/: README.md (1 file)
  - Plus the empty directory tree pre-created in the prior session.
- The user is the architect. The AI is the implementer. The user reviews. This is the partnership.
- **No code has been written. No commits have been made. The repo does not yet exist as a git repo.** This is intentional and per the user's instruction.

---
