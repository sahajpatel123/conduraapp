# Condura — Project Understanding

> **Purpose.** The single document that any future agent (or human) reads to
> understand the Condura project deeply, quickly, and without ambiguity.
> Written by the 2026-06-24 deep-analysis session after a full sweep of
> `MISSION.md`, `FOOTHPATH.md`, `STYLE.md`, `LOGBOOK.md`, `docs/`,
> `internal/`, `app/`, `web/`, `cmd/`, and the git log.
>
> **This document is the anchor.** It does not replace `MISSION.md` (the spec),
> `LOGBOOK.md` (the session log), or `FOOTHPATH.md` (the state ledger) — it
> summarizes, cross-references, and adds the *context* that no other doc
> carries in one place: the why behind the what, the drift between spec and
> reality, and the priorities a future agent must respect.

---

## 1. The One-Paragraph Summary

**Condura is a free, OS-native AI agent that lives on a user's computer and
acts as the conductor of every other AI tool installed there.** It is summoned
by a custom global hotkey, opens an overlay chat/voice surface, clicks and
types through any app via a 4-tier computer-use stack, delegates long tasks to
8 sub-agent CLIs (Claude Code, Codex, Antigravity, OpenCode, Kilo, Hermes,
Gemini, Ollama), and learns the user's behavior over time — all gated by a
**deterministic Gatekeeper** (no model decides "is this safe?", only
policy) and an **HMAC-chained audit log** (every action is forensically
recoverable). It is the **missing conductor** between the half-dozen amazing
but siloed AI tools the user already has. **Free forever. No lock-in. No
tracking. No compromise on speed or safety.**

---

## 2. The Brand: Why "Condura" and Not "Synaptic"

The project was originally called **Synaptic** (see the `LOGBOOK.md` H1
header which still reads `# Synaptic — LOGBOOK.md`). It was renamed to
**Condura** per the user's decision recorded in the LOGBOOK and applied across
the codebase in commit `b721855 fix: rebrand build/install pipeline from
Synaptic to Condura`.

- **Binary names:** `condurad` (daemon), `condura` (CLI), `condura-tui`
  (terminal UI), `condura-gui` (Wails-bundled desktop app).
- **Makefile** `BINARY_NAME := condurad`.
- **Domain:** `condura.app`. Hub: `hub.condura.app`. Support: `support@condura.app`.
- **GitHub:** `github.com/sahajpatel123/conduraapp` (private).
- **Wake word still says "hey synaptic"** in `MISSION.md` §4 decision 35 and
  §19.3 — this is a **known drift**; the user wants it to be
  "hey condura" per the FOOTHPATH rename. Fix in a future session if
  asked.
- **OAuth scheme:** `condura://` (was `synaptic://`).
- **Backup directory:** `~/Documents/synaptic-backups/` — still uses the
  old name in code (known drift).

**The agent must mentally substitute Condura → Synaptic when reading the
spec, and "Synaptic" still appearing in code is a backlog item, not a
project restart.**

---

## 3. The Project Shape (Repo Map)

| Layer | Language | Path | Purpose |
|---|---|---|---|
| **Daemon** | Go 1.22+ | `cmd/condurad/`, `internal/daemon/` | Long-running daemon. Owns storage, LLM routing, safety, delegation, GUI IPC. |
| **CLI client** | Go | `cmd/condura/` | Single-binary CLI for the daemon. Subcommands: `ping`, `version`, `config`, `llm`, `apikeys`, `delegate`, `sync`, `hub`, `skills`. |
| **TUI** | Go | `cmd/condura-tui/` | Terminal UI for power users / SSH. |
| **GUI shell** | Go + Wails v2 | `app/web/` | Wails-bundled GUI binary embedding the daemon + Svelte frontend. |
| **GUI frontend** | Svelte 5 + TypeScript | `app/web/frontend/` | Reactive UI. 10 routes: Chat, Settings, Audit, Channels, Delegation, Hub, Replay, Sync, Skills, About. |
| **Marketing site** | Next.js 14 | `web/` | Public site at `condura.app` (Manifesto, Changelog, Legal, Download, etc.). |
| **Docs** | Markdown | `docs/` | ADRs, architecture, recipes, runbooks, on-device verification, roadmap. |
| **Operating manual** | Markdown | `MISSION.md` `STYLE.md` `LOGBOOK.md` `FOOTHPATH.md` | How the project thinks. |
| **Operating scripts** | Bash / NSIS | `scripts/` | build-gui.sh, install.sh, condura-gui.nsi (Windows NSIS), homebrew tap, package-gui-installers.sh, verify-release-artifacts.sh. |

**Source of truth hierarchy (per the human's convention):**
1. **User's direct instructions** — highest priority.
2. **`MISSION.md`** — the spec.
3. **`FOOTHPATH.md`** — the live state ledger.
4. **`LOGBOOK.md`** — the session history.
5. **`STYLE.md`** — the operating manual.
6. **`docs/`** — architecture, ADRs, runbooks.
7. **Code** — the proof.

> ⚠️ The human explicitly told me on 2026-06-07: "MISSION.md and CLAUDE.md
> are the same file going forward. Treat MISSION.md as authoritative."
> All other docs (LOGBOOK, CONTRIBUTING, docs/README) still reference
> "CLAUDE.md" — this is known and intentional drift; do not "fix" it
> unless asked.

---

## 4. The Internal Go Packages — 50+ Subsystems

The Go code lives in `internal/`. The 50+ packages form the daemon's nervous
system. Below is a categorized map, not an exhaustive list — see
`internal/` for the full enumeration.

### 4.1 Safety & Survival (the spine)
These packages implement `MISSION.md` §2 (the seven non-negotiable invariants)
and §5 (the seven technical non-negotiables). They are the **functional
unblock for the agent** — without them, every non-READ action is denied at
runtime.

- `internal/blastradius` — classifies every action as READ / WRITE / NETWORK
  / DESTRUCTIVE. ✅ Built.
- `internal/gatekeeper` — the **deterministic rules engine**. Reads
  `~/.synaptic/policy.yaml`. **Default-deny**. ✅ Built (policy engine real
  as of Phase 14I).
- `internal/halt` — kill switch. Three layers: hard hotkey, watchdog timer,
  network isolation (`halt.NetworkGuard` interface + `InProcessGuard`
  implementation). ✅ Built.
- `internal/anomaly` — behavioral anomaly detector (speed, loops, duration,
  new endpoints). ✅ Built.
- `internal/audit` — HMAC-chained append-only log. SHA-256 chain, 90-day
  retention, secret redaction. ✅ Built.
- `internal/sanitize` — model isolation sanitizers (shell commands, Python
  AST, file paths, URLs, message bodies). ✅ Built.
- `internal/sensitive` — sensitive site detector (banking/health). ✅ Built.
- `internal/failover` — circuit breakers + daily spend monitor + chain
  runner + failover orchestrator. ✅ Built.
- `internal/autonomy` — autonomy matrix (per-task-type and per-app autonomy
  levels). ✅ Built.
- `internal/watchdog` — inactivity auto-halt + audit-before-halt ordering.
  ✅ Built (Phase 16).

### 4.2 Perception (the eyes)
- `internal/perception` — **Selective Perception** data model (Strategy
  enum, EnergyMode, SmartCapturer, DirtyTracker, PIIRedactor). ✅ Built in
  Phase 14I. **The live event source (CGEventTap, AT-SPI) is not wired
  yet — that's v0.2.0.**

### 4.3 LLM & Routing
- `internal/llm` — `Provider` interface, OpenAI-compat for 9 providers,
  dedicated Anthropic + Google impls, pricing registry, `EstimateCost`.
  ✅ Built.
- `internal/stream` — SSE streaming manager. Loop.Ask subscribes-before-start,
  accumulates deltas, persists assistant reply. ✅ Built (Phase 18).
- `internal/router` — hybrid LLM router. ❌ **NOT IN REPO.** v0.2.0 work.
  v0.1.0 uses a single configured `providerName` + `model`.

### 4.4 Memory & Skills
- `internal/memory` — 3 layers (episodic, semantic, procedural). SQLite +
  FTS5 + sqlite-vec. Encrypted at rest.
- `internal/skills` — `agentskills.io` compatible. Auto-create, self-improve,
  bundle, share via Hub, scan for promptware.

### 4.5 Computer Use
- `internal/computeruse` — 4-tier router (ORAX Eye → mac-cua → macOS-MCP →
  vision CUA). Backend dispatch.
- `internal/executor` — `shell.exec` + `computeruse.*` dispatch with
  re-gate carve-out, timeouts, audit.
- `internal/blastradius` (also above) — feeds into the executor.

### 4.6 Agent & Delegation
- `internal/agent` — the agent loop + planner. `Loop.Ask` is the real
  stream-driven chat/voice loop (gatekeeper → audit → persist → SSE → TTS).
  ✅ Built.
- `internal/delegation` — 8 sub-agent CLIs, GatedRunner, semaphores, file
  coordination, pending action queue (`internal/pending`). ✅ Built.
- `internal/conductor` — conductor-level orchestration. ✅ Built.

### 4.7 Storage, IPC, Config
- `internal/storage` — `modernc.org/sqlite` (pure Go) + AES-256-GCM column
  encryption. Schema v6.
- `internal/ipc` — JSON-RPC 2.0 server (HTTP + WebSocket), bearer-token
  auth, batch + notifications, typed Go client.
- `internal/config` — YAML loader, env-override, `Validate()`.
- `internal/secrets` — OS keyring (`zalando/go-keyring`) with file fallback.
- `internal/api_key` — manager over storage + secrets; OAuth interface;
  Google PKCE. **OllamaLocalSentinel** for the no-API-key local case.
- `internal/version`, `internal/logger`, `internal/health`, `internal/lockfile`,
  `internal/sse`, `internal/status`, `internal/crash` — infrastructure.

### 4.8 User-Facing & Ecosystem
- `internal/onboarding` — first-run state machine (`eula → permissions →
  hotkey → complete`). 4 screens (Phase 14A converged flow). Drives
  `onboarding.*` RPCs.
- `internal/account` — email magic link + OAuth (Google / GitHub / Apple).
- `internal/channels` — Telegram (✅), Signal/WhatsApp/iMessage (stubs).
- `internal/reach` — channel ecosystem plumbing.
- `internal/hub` — Hub client (publish + browse skills).
- `internal/sync` — P2P encrypted sync (Ed25519 identity, mDNS, Kademlia
  DHT, Noise XX, CRDT). Device pairing via QR + PIN.
- `internal/replay` — 24h action replay, scrubbable timeline.
- `internal/session`, `internal/conversation` — chat history, sessions.
- `internal/voice` — STT (whisper.cpp local + OpenAI cloud) + TTS
  (OpenAI, ElevenLabs, native `say`).
- `internal/hotkey` — `github.com/atotto/carbon` (mac) +
  `golang.design/x/hotkey` (cross-platform).
- `internal/overlay` — overlay window state machine.
- `internal/tray` — system tray / menu bar.
- `internal/window` — window control (Wails owns actual windows).
- `internal/permissions` — TCC / UI Automation grants.
- `internal/presence` — user activity heartbeat.
- `internal/notify` — native notifications.
- `internal/backup`, `internal/uninstall`, `internal/updater` — install +
  update + backup flow.
- `internal/adaptive` — user-adaptive engine (Honcho-style user model,
  dialectic proposer/critic/adjudicator, predictor, strength slider).
- `internal/i18n` — backend messages via `golang.org/x/text/message`.
- `internal/mcp` — MCP client/server (10k+ servers consumable).
- `internal/telemetry` — local-only telemetry (no cloud).
- `internal/trust` — device trust grants.
- `internal/tui` — terminal UI bindings.

### 4.9 The 4-Tier Computer-Use Router
Per `MISSION.md` §11.2:
1. **ORAX Eye** — structured AX tree (free, fast, ~50ms). Primary.
2. **mac-cua** — background-first, `CGEventPostToPid`. Apache 2.0.
3. **macOS-MCP** — comprehensive foreground interaction. MIT.
4. **Vision CUA** — Anthropic Computer Use or OpenAI CUA, ~$0.02-0.05/action.
   **Last resort, opt-in only** (currently disabled per Phase 17 Rec 2).

### 4.10 The 8 Default Sub-Agents (per `MISSION.md` §13.2 + FOOTHPATH §8)
| Name | Binary | Adapter |
|---|---|---|
| `claude` | `claude` | stream-json, `--print --output-format stream-json --model` |
| `codex` | `codex` | json, `--json --model` |
| `antigravity` | `agy` | json, `--output-format json --model` |
| `opencode` | `opencode` | json, `--format json` |
| `kilo` | `kilo` | json, `--json` |
| `hermes` | `hermes` | json, `--format json` |
| `gemini` | `gemini` | json, `--output-format json` |
| `ollama` | (no subprocess) | direct HTTP to `localhost:11434` |

If a binary isn't installed, spawn returns `ErrAgentNotFound` — no
auto-install (product decision).

---

## 5. The Frontend (Svelte 5 GUI)

`app/web/frontend/` is the Svelte 5 + TypeScript UI. 10 routes:

| Route | Purpose |
|---|---|
| `#/` (Chat) | Primary UX. Streams LLM responses; renders tool calls as `<details>` blocks (persisted) and pills (streaming). |
| `#/settings` | All config: Account, Adaptive engine, Voice, Channels, Channels, Backup/Restore, Onboarding re-run, Legal, etc. |
| `#/audit` | HMAC-chained log viewer; integrity verification. |
| `#/channels` | Connect/disconnect Telegram (and the stubs). |
| `#/delegation` | Spawn sub-agents (`delegate.spawn` / `delegate.list_agents` / `delegate.cancel`). |
| `#/hub` | Skills Hub browser; `PublishModal` for publishing. |
| `#/replay` | 24h action replay timeline. |
| `#/sync` | P2P device pairing (`PairingModal` with QR + PIN + TTL). |
| `#/skills` | Local skill list; load bundles. |
| `#/about` | App info. |

Key components:
- `OnboardingWizard.svelte` + `lib/components/onboarding/*` — 4-screen
  wizard (EULA → Permissions → Hotkey → Ready).
- `OverlayPrompt.svelte` — **the primary UX surface**, hotkey-launched.
  Extracts from `App.svelte` in Phase 18.
- `ConsentModal.svelte` — polls `gatekeeper.pending_consent`; calls
  `gatekeeper.approve` / `gatekeeper.deny`.
- `VoiceOrb.svelte` — animated waveform during listening.
- `HotkeyRecorder.svelte` — combo capture.
- `PairingModal.svelte` — QR + PIN + TTL countdown.
- `PublishModal.svelte` — name + semver-validated version + archive upload.
- `AccountMenu.svelte` / `SignInPanel.svelte` — magic link + OAuth.
- `LocaleSelector.svelte` — UI language switcher (en/es/fr/de/ja/zh).
  **Locale JSON files only exist for `en` (438 lines); other 5 fall back
  to English content. v0.2.0 work via Crowdin.**
- `Sidebar.svelte`, `Toasts.svelte`, `LiveTranscript.svelte`,
  `PendingActions.svelte` — chrome.

**Frontend store architecture:** Svelte 5 runes (`$state`, `$derived`)
plus typed `lib/ipc/client.ts` wrappers around the JSON-RPC methods.
**No mock data in production paths** — every screen reads from real
`ipc.*` calls.

---

## 6. The Marketing Site (`web/`)

Next.js 16 + Tailwind v4 + motion v12. Creative direction: **"The Touch"**
(per memory). Pages: Home (the dark-bulb-then-light hero), Manifesto,
Changelog, Legal, Download, Ecosystem, Security, Orchestration, Bring-Your-Own-AI.

**Hard constraint (memory):** the website is **kept strictly separate from
the technical side** (Go daemon + `app/` Wails GUI) until the technical
side is finished. **Do not wire them together. Do not touch `app/` or Go
code in website sessions.** Sahaj's uncommitted `app/web/frontend` Svelte
changes are his own work in progress — never commit, modify, or review
them unless asked.

---

## 7. The Survival Invariants (the 7 + 7)

### 7.1 The Seven Non-Negotiable Invariants (`MISSION.md` §2.1)
1. **Strategist and Gatekeeper are separate systems.** Strategist = model.
   Gatekeeper = deterministic code. Never the same.
2. **Gatekeeper is the only path to physical action.** No model output
   flows to a click, type, or shell exec without passing the Gatekeeper.
3. **Destructive actions require a real human at the keyboard.** Native
   modal dialog. Blocks until clicked. No exceptions.
4. **The user can always stop the agent.** Hard hotkey + watchdog + network
   isolation + menu bar kill. Four independent mechanisms. The agent
   cannot disable any of them.
5. **Every action is auditable, in a tamper-resistant log.** HMAC-chained,
   append-only, never deleted.
6. **The agent is a guest, not an owner.** Requests permission to enter
   rooms; user grants or denies; never escalate.
7. **OS permissions are granted by the user, on their machine.** Onboarding
   makes this easy and clear.

### 7.2 Hard Constraints (`MISSION.md` §2.2)
1. User's API key is sacred — never log, never send, never include in
   telemetry.
2. User always in control — no action without consent or pre-approved rule.
3. Speed is the product — cold start < 500ms, hotkey→overlay < 100ms,
   first token < 1.5s.
4. Local-first — memory, skills, audit, embeddings on disk, encrypted.
5. Free forever — no feature gates, no premium tier, no nags.
6. Proprietary source, free binary — repo private, binaries signed +
   notarized + downloadable from `condura.app`.

### 7.3 The Seven Technical Non-Negotiables (`MISSION.md` §5)
1. **Action classification by blast radius** — READ / WRITE / NETWORK /
   DESTRUCTIVE, classified before execution.
2. **Mandatory pre-action verification (twin snapshots)** — capture AX
   tree, articulate target, compare to second snapshot, abort on diff.
3. **Kill switch (3 layers)** — hard hotkey + watchdog + network
   isolation (separate process the agent cannot stop).
4. **Audit log of everything** — HMAC-chained, append-only, never deleted.
5. **Model isolation, not just switching** — sanitizers between every
   handoff; never let Model A's output flow into Model B's context without
   a deterministic validation layer.
6. **Behavioral anomaly detector** — speed, loops, duration, new
   endpoints, consecutive errors. **Auto-recovery is the enemy.**
7. **Strategist vs Gatekeeper separation** — they must never be the same
   model or both probabilistic.

> **A feature that reaches a user without the armor under it is the wrong
> feature.** — `MISSION.md` §2.

---

## 8. The 36 Locked Decisions (per `MISSION.md` §4)

All decisions are listed verbatim in the spec. Key ones to internalize:

- **#1 Name: Condura** (was Synaptic, renamed).
- **#4 Foundation:** from scratch in Go + TypeScript (no Hermes fork).
- **#5 Computer use:** all 3 backends + vision CUA, 4-tier router.
- **#6 Routing:** hybrid with memory (cost-first cascade, bias toward what
  worked).
- **#8 Hotkey:** user must set on first run (no default; suggestions:
  Option+Option, Cmd+Shift+Space, Ctrl+Space, Ctrl+Ctrl).
- **#11 Languages at v0.1.0:** English + Spanish + French + German +
  Japanese + Mandarin (6).
- **#13 Launch strategy:** Public v0.1.0, all in — Product Hunt + Hacker
  News + Reddit (r/singularity, r/LocalLLaMA, r/AI_Agents) on same day.
- **#15 Provider down:** auto-failover — Ollama local first, then any
  configured backup key.
- **#16 Multi-machine sync:** P2P encrypted sync, no central server.
- **#17 Uninstall behavior:** auto-backup before uninstall to
  `~/Documents/synaptic-backups/`.
- **#23 Concurrency:** default 2 parallel sub-agents, max 5, user-configurable.
- **#24 Autonomy:** default cautious (warn before any action).
- **#25 Uncertainty:** ask user immediately ("I'm 60% sure you want X. Proceed?").
- **#26 Energy budget:** refuse, force user decision.
- **#27 Daemon autostart:** auto-start on login.
- **#30 User account:** email + magic link (for hub, donations, support;
  P2P sync needs no account).
- **#34 Multi-install:** block second install.
- **#35 Wake word:** "hey synaptic" — **STILL DRIFT in code**, the user
  wants "hey condura". Backlog item.
- **#36 EULA:** Freeware EULA — free personal + commercial, no
  redistribution, revocable for abuse.

---

## 9. Current Status (as of 2026-06-24)

**Per `FOOTHPATH 3` (captured 2026-06-22, main @ `1c41506`) and the most
recent LOGBOOK entries:**

### 9.1 v0.1.0 is shipped
- **Phase 13 (release/distribution) is complete.** v0.1.0 is published
  with signed auto-update (`manifest.json`), GoReleaser packages, and GUI
  installers (DMG / portable exe / Linux binary).
- `release-verify` CI runs on every `main` push.
- **15/15 green on main CI + 3/3 green on Release Verify** (commit `b254108`).

### 9.2 What's Working (Tier-3 verified end-to-end)
- **Daemon** — boots, migrates SQLite schema to v6, initializes 18+
  subsystems, listens on TCP/Unix, exposes ~80 JSON-RPC 2.0 methods,
  persists across restarts.
- **Onboarding** — 4-screen wizard (EULA → Permissions → Hotkey → Ready)
  with `onboarding.eula`, `onboarding.set_step`, `onboarding.probe_power`
  (Ollama + 8 CLIs), `onboarding.finish`, `onboarding.reset`. Live
  e2e verified.
- **Chat** — `llm.chat` and `llm.stream` (SSE) working. Ollama local
  tested: "What is 2+2? One word." → "Four" in 128 output tokens, $0
  cost. **Ollama no-key** auto-fills `OllamaLocalSentinel =
  "ollama-local-no-key"**.
- **Audit** — HMAC-chained, `replay.verify_integrity` returns
  `{"valid":true,"rows_checked":N}`. 4+ events captured in Phase 15 Run #1.
- **Backup** — auto-backup scheduler creates
  `condura-backup-<ISO-date>.zip` on daemon startup.
- **GUI** — overlay sends messages, restore works, svelte-check 0/0,
  tool calls render, premium design system overhaul landed.
- **i18n** — `t(key)` function works; `en.json` has 438 lines; es/fr/de/
  ja/zh fall back to English content.
- **Voice loop** — `Loop.Ask` is a real stream-driven loop
  (gatekeeper → audit → persist → SSE → TTS).

### 9.3 The Honest Backlog (v0.2.0+)
Per `docs/roadmap-v0.2.0.md` and `MISSION.md` §33.5.2:
- **Subscription OAuth** (ChatGPT Plus, Claude Pro, SuperGrok) — 2-3
  weeks of work.
- **Hardened Layer 3** (`pf`/`netsh` companion binary).
- **CGEventTap / AT-SPI dirty tracking** wired to perception.
- **MCP UI** (`Mcp.svelte` route).
- **Crowdin i18n sync** + real translations for 5 non-English languages.
- **Public Hub + Dashboard deploy** (`hub.condura.app` and
  `condura.app/dashboard` as separate Next.js apps).
- **Vision CUA opt-in** (currently disabled).
- **Non-macOS voice** via cloud STT.
- **`file.*` executor dispatch** (currently "not yet supported").
- **Hybrid LLM router** (`internal/router/` package) — v0.1.0 uses single
  configured provider+model.

### 9.4 Drift Between Spec and Implementation (Known)
- **Wake word "hey synaptic"** in `MISSION.md` §4 #35 and §19.3 — user
  wants "hey condura". Backlog.
- **Backup directory `~/Documents/synaptic-backups/`** — still uses old
  name. Backlog.
- **`internal/router/`** — spec describes a hybrid-with-memory router;
  package does not exist. v0.2.0 work.
- **OAuth subscription flows** — marketing copy mentions ChatGPT Plus /
  Claude Pro / SuperGrok; backend stubs return "coming in v0.2.0".
- **i18n locale JSONs** — only `en.json` exists; the other 5 fall back
  to English.
- **Marketing copy** still has 10k+ MCP servers claim, real Signal/WhatsApp
  /iMessage, etc. The website is a separate track; align with backend in
  v0.2.0 coordinated pass.

### 9.5 The Single Next Human Action
**On-device verification on a clean macOS machine, per
`docs/on-device-verification.md` and the Phase 15 checklist.** This is
the gate before public launch. The human must drive the physical
keyboard. An agent cannot do this.

---

## 10. The Working Style — How Every AI Must Operate

Per `STYLE.md` and the human's convention. This is non-negotiable.

### 10.1 Identity & Honesty
- **Byline whatever actually ran.** Don't impersonate other models. The
  harness varies by session; truth over roleplay. The user values
  honesty over roleplay.
- The user runs **local Ollama**, not Claude Code. Byline as
  `minimax-m3 via ollama` (or whatever the live harness is).
- **Do not run fake "agent swarm" simulators** (the
  `dynamic-workflow-emulator` skill is declined; it conflicts with the
  "scope to the maturity of the work / one goal per session" rule).

### 10.2 The Three-Tier Verification Ladder
**A green test is not proof the feature works.** Every shipped feature
passes all three:
- **Tier 1 — Unit tests.** Single package, controlled fixture.
- **Tier 2 — Integration / E2E test in Go.** Real `initSubsystems`,
  real `ipc.Server`, real SQLite.
- **Tier 3 — Runtime smoke test.** `go build`, run the actual
  `condurad` binary, drive it with `curl` over its real RPC surface,
  inspect the real on-disk state with `ls`, `sqlite3`, `unzip`.

**A mediocre AI ships a passing test suite. A partner AI ships a passing
test suite AND runs the binary to confirm.** — `STYLE.md` §0.

### 10.3 Commit Policy
**The human commits manually.** AI must not commit on his behalf. **Exception
(added 2026-06-09):** for green sub-phases, AI may commit directly to `main`
with conventional-commit messages and push at end of session. Use
`Co-Authored-By` trailer.

### 10.4 Context Loading per Session
- **Deep on code** — read every file you might touch.
- **Light on docs** — re-read `MISSION.md`, `LOGBOOK.md`, the specific
  architecture doc relevant to the task. Don't re-read every doc.
- Always re-read `synaptic-identity`, `synaptic-canon-files`,
  `synaptic-conventions` from memory.

### 10.5 Session Shape
- **One goal per session** — the goal determines the size.
- "Do whatever it takes to accomplish one goal" within the session.

### 10.6 Review Posture
- **When the user asks for a "review" or "how is this looking"** — read-only
  review with honest critique. Do NOT edit, fix, or augment code. Call out:
  what the change is actually doing, behavior changes vs cleanups, things
  to verify before commit, concerns even when the direction is right.
- **Foundation-level reviews** when the user is laying down new code:
  structural soundness for what it is, not full architectural critique.

### 10.7 Prose Style
- Match `MISSION.md`. Section headers like `## 4. The 26 Locked Decisions`
  (numbered, Title Case, period). 5-line file headers in code. Long,
  didactic, opinionated, exhaustive.

### 10.8 Hard Rules for AI Agents (from `MISSION.md` §30.5)
1. **Never** delete or rewrite `MISSION.md` content silently. Append or
   annotate.
2. **Never** commit secrets, API keys, OAuth tokens, or `.env` files.
3. **Never** bypass the safety layer to "make something work faster".
4. **Never** introduce a new dependency without documenting it in
   `MISSION.md`.
5. **Never** skip tests for the safety or perception modules.
6. **Always** update the LOGBOOK before you finish.
7. **Always** read the latest LOGBOOK before you start.

---

## 11. The Known Drifts, Quirks, and Gotchas

1. **`MISSION.md` H1 reads `# Condura — CLAUDE.md`** but `CLAUDE.md` does
   not exist. Intentional drift. The H1 preserves the link targets other
   docs use.

2. **Wake word is "hey synaptic"** in `MISSION.md` §4 #35 and §19.3, but
   the user wants "hey condura". Backlog. Files affected: 11.

3. **OAuth scheme `condura://`** — was `synaptic://`, renamed in 6 files
   (Go + Svelte + tests).

4. **Backup directory `~/Documents/synaptic-backups/`** — still uses old
   name. Backlog.

5. **Two consent namespaces:** `gatekeeper.*` (canonical) and
   `safety.consent.*` (DEPRECATED alias). GUI uses `gatekeeper.*`.

6. **`noopAgentExecutor` was a no-op** until Phase 14I replaced it with
   `agent.NewComputerUseExecutor` that routes through the real ComputerUse
   pipeline (gated, audited, with smart fallback for unknown action types).

7. **`internal/perception` package was missing** until Phase 14I. Now
   built: Strategy enum, EnergyMode, SmartCapturer, DirtyTracker,
   PIIRedactor. **14 unit tests.** Live event source still not wired.

8. **Network isolation Layer 3 had no in-process implementation** until
   Phase 14I. Now `halt.NetworkGuard` interface + `InProcessGuard`
   implementation. v0.2.0 replaces with real `pf`/`netsh` companion.

9. **"What Synaptic learned about you" UI + strength slider** — added to
   Settings.svelte in Phase 14I (Adaptive engine section). Reads
   `adaptive.profile`, `adaptive.strength.get/set`, `adaptive.forget`,
   `adaptive.reset`.

10. **No delegation UI** — until Phase 14I added `Delegation.svelte`
    route + sidebar entry. Spawn via `delegate.spawn` / `delegate.list_agents`
    / `delegate.cancel`.

11. **`internal/secrets.TestNew_NoFilePath_Auto`** — passes 3/3 in CI but
    historically fails 1/3 on bare macOS. Tracked, not blocking.

12. **Wails build under Go 1.26+** has duplicate `_OBJC_*_AppDelegate`
    symbols — Wails v2.12.0 upstream issue, not a project bug. Local Go
    1.26+ devs should pin to 1.25.x via `go.work` toolchain directive.
    CI on Go 1.25.11 is green.

13. **`subs.Executor` is nil** when `cuComps` is nil (no LLM configured),
    blocking shell-only sub-agents. Open question for next session.

14. **`default.yaml` config drift:** `cfg.Router.Priorities["chat"]`
    references `claude_code` etc., which aren't wired. **Do not change
    without explicit instruction** (per memory).

15. **`svelte-check` warnings** — 7 pre-existing unused-CSS-selector
    noise (`.modal code`, `.kbd`, etc. dead style blocks left behind by
    i18n sweep). Trivial 10-min cleanup. Not blocking.

16. **The website (`web/`) is KIMI K2.6's territory.** Don't touch `app/`
    or Go code in website sessions. Don't modify his uncommitted
    `app/web/frontend` Svelte changes.

---

## 12. The Cross-Reference Index

### 12.1 If you need to know…
- **The project spec** → `MISSION.md` (canonical), `STYLE.md` (operating
  manual), `FOOTHPATH.md` (state ledger), `LOGBOOK.md` (session log).
- **The architecture** → `docs/architecture/00-overview.md` through
  `09-ipc.md`.
- **The 5 ADRs** → `docs/adr/0001-go-over-python.md` through
  `0005-p2p-sync.md`.
- **The current phase state** → `docs/phase14-completion.md` and
  `docs/phase15-verification.md`.
- **What's deferred to v0.2.0+** → `docs/roadmap-v0.2.0.md` and
  `MISSION.md` §33.5.2 (the spec-debt ledger).
- **How to build / install / release** → `Makefile`, `scripts/build-gui.sh`,
  `docs/release-runbook.md`, `docs/release-keys.md`.
- **The release verification** → `docs/phase15-verification.md` and
  `docs/on-device-verification.md` and `docs/macos-verification-runbook.md`.
- **The threat model** → `docs/threat-model-reach.md`.
- **User-facing guides** → `docs/user-guide/`, `docs/guides/`.
- **The most recent work** → last 5 entries of `LOGBOOK.md`.

### 12.2 If you need to do…
- **Build the daemon:** `make build` → `bin/condurad`, `bin/condura`,
  `bin/condura-tui`.
- **Build the GUI:** `scripts/build-gui.sh` or
  `wails build` from `app/web/`.
- **Run tests:** `go test -count=1 -race -timeout 300s ./...`.
- **Run lint:** `golangci-lint run --timeout=5m ./...`.
- **Boot the daemon:** `condurad -config /tmp/c.yaml -data-dir /tmp/data
  -listen "tcp://127.0.0.1:18600"`.
- **Ping it:** `curl -X POST http://127.0.0.1:18600/api -H "Content-Type:
  application/json" -d '{"jsonrpc":"2.0","id":1,"method":"ping","params":{}}'`.
- **Append a LOGBOOK entry:** follow the format in `MISSION.md` §30.3.

---

## 13. The Partnership Framing

Per `MISSION.md` §31 and the human's convention:

> This project is being built by a human + AI partnership. The human is
> the architect and product lead. The AI is the implementer and reviewer.
> We move fast. We do not ship broken code. We do not cut corners on
> security. We ship the best version of what we imagined, then we ship a
> better one. **Laziness is not in our vocabulary. Everything is
> possible.**

The user is **Sahaj**. He is the product lead. The AI is the implementer
and reviewer. He reviews and commits.

**The mission, in one line:**

> Make AI useful to every ordinary person, on every computer, for free.
> No lock-in. No tracking. No compromise on speed or safety.

---

## 14. The Checklist for the Next Agent

When you pick up this project:

1. **Read this file top to bottom** (you just did).
2. **Read `LOGBOOK.md` end to end** to see the most recent sessions,
   in order, including the open questions left for you.
3. **Read `MISSION.md` end to end** for the spec.
4. **Read `FOOTHPATH.md` end to end** for the live state.
5. **Read `STYLE.md` end to end** for the operating manual.
6. **Read `docs/architecture/00-overview.md`** for the mental model.
7. **Then run the binary.** Boot it, ping it, install an API key via
   `apikeys.set`, spawn a sub-agent, approve a pending action. **The
   binary is the source of truth, not this doc.**
8. **Check `git status` and `git log --oneline -10`** — the human may
   have uncommitted work in progress.
9. **Ask if anything is ambiguous.** The user has explicitly invited
   unlimited questions.
10. **Append to LOGBOOK.md before you finish.** Format per `MISSION.md` §30.3.

---

**Last updated:** 2026-06-24 by the deep-analysis session.
**Byline:** minimax-m3 via ollama.
**Status:** Anchor document. Cross-references all primary sources. Use
this to orient, then drill into the specific source for the detail.
