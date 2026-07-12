# Condura — Project Understanding

> **Purpose.** The single document that any future agent (or human) reads to
> understand the Condura project deeply, quickly, and without ambiguity.
> Written by the 2026-07-12 deep-analysis session after a full sweep of
> `MISSION.md`, `FOOTHPATH.md`, `STYLE.md`, `LOGBOOK.md`, `CHANGELOG.md`,
> `AGENTS.md`, `docs/`, `condura-app/`, `condura-gui/`, `condura-ui/`,
> `condura-ops/`, `condura-brand/`, `condura-studio/`, `condura-hub/`,
> `condura-sdk/`, the working tree, the branch inventory, the worktree at
> `.worktrees/phase-15-ship-readiness/`, and the 3 stashes.
>
> **This document is the anchor.** It does not replace `MISSION.md` (the spec),
> `LOGBOOK.md` (the session log), or `FOOTHPATH.md` (the state ledger) — it
> summarizes, cross-references, and adds the *context* that no other doc
> carries in one place: the why behind the what, the drift between spec and
> reality, and the priorities a future agent must respect.
>
> **Previous version:** the 2026-06-24 anchor (same path) was structurally
> correct on the daemon and the brand but stale on paths, GUI layers,
> i18n completeness, and branch reality. This 2026-07-12 refresh supersedes
> it line-for-line; the old text is preserved in git history (commit before
> this file's rewrite).

---

## 1. The One-Paragraph Summary

**Condura is a free, OS-native AI agent that lives on a user's computer and
acts as the conductor of every other AI tool installed there.** It is summoned
by a custom global hotkey, opens an overlay chat/voice surface, clicks and
types through any app via a 4-tier computer-use stack, delegates long tasks to
8 sub-agent CLIs (Claude Code, Codex, Antigravity, OpenCode, Kilo, Hermes,
Gemini, Ollama), and learns the user's behavior over time — all gated by a
**deterministic Gatekeeper** (no model decides "is this safe?", only policy)
and an **HMAC-chained audit log** (every action is forensically recoverable).
It is the **missing conductor** between the half-dozen amazing but siloed AI
tools the user already has. **Free forever. No lock-in. No tracking. No
compromise on speed or safety.**

---

## 2. The Brand: Why "Condura" and Not "Synaptic"

The project was originally called **Synaptic** (the LOGBOOK H1 still has
historical `Synaptic` mentions — intentional drift). It was renamed to
**Condura** per the user's decision recorded in the LOGBOOK and applied across
the codebase in commit `b721855 fix: rebrand build/install pipeline from
Synaptic to Condura`.

- **Binary names:** `condurad` (daemon), `condura` (CLI), `condura-tui`
  (terminal UI), `condura-gui` (Wails-bundled desktop app).
- **Module path:** `github.com/sahajpatel123/conduraapp` (private GitHub).
- **Domain:** `condura.app` (marketing site). Hub: `hub.condura.app` (v0.2.0).
  Support: `support@condura.app`.
- **Wake word:** `"hey condura"` (MISSION §4 #35 + code defaults aligned
  since 2026-07-10 commit `02ec07f`).
- **OAuth scheme:** `condura://` (was `synaptic://`, renamed).
- **Backup directory:** `~/Documents/condura-backups/` (code + MISSION
  aligned; rename closed in P3-13 on 2026-07-10).
- **Config / data dir:** `~/.condura/` (was `~/.synaptic/`, renamed).
- **Config env-var prefix:** `CONDURA_<SEC>__<FIELD>` (was `SYNAPTIC_*`).
  Historical code comments may still mention `SYNAPTIC_` — not a bug.

**The agent must mentally substitute Condura → Synaptic when reading the
spec, and "Synaptic" still appearing in code is a backlog item, not a
project restart.** The README on the `phase-15-ship-readiness` worktree
still says `synaptic.app` (stale) and `./bin/synapticd` (stale) — worktree
has not been reorg'd into the new layout.

---

## 3. The Project Shape (Repo Map)

The repo is a **topic-sliced monorepo** since the 2026-07-04 reorg (commit
`9b893c1`). Every top-level directory is one product surface; you navigate
by topic without touching unrelated code.

| Layer | Language | Path | Purpose |
|---|---|---|---|
| **Daemon + libraries** | Go 1.25+ | `condura-app/cmd/`, `condura-app/internal/` | All Go code. Daemon, CLI, TUI, Wails shell, 60+ packages. |
| **Desktop GUI frontend** | Svelte 5 + TS + Vite | `condura-gui/frontend/` | The in-app user-facing surfaces. Meridian shell + v2 design system. |
| **Marketing site** | Next.js 16 + Tailwind 4 | `condura-ui/` | `condura.app`. Public, Vercel-deployed. |
| **Operations** | Bash + YAML + NSIS | `condura-ops/` | CI, install scripts, release tooling, deployment configs. |
| **Brand** | CSS + SVG | `condura-brand/` | Tokens, palette, motion, logos, fonts policy, hero assets. Honest starter-kit scope. |
| **Studio** | Remotion + Node | `condura-studio/` | Video / film creation. 3 active projects (demo / spotlight / thread) + 1 exploration (my-video). |
| **Constitution** | Markdown | `condura-mind/` | MISSION.md (spec), LOGBOOK.md, FOOTHPATH.md, STYLE.md, CHANGELOG.md, AGENTS.md, docs/, synapse/, CLAUDE.md (redirect stub). |
| **Reserved** | — | `condura-hub/`, `condura-sdk/` | Read-only scaffolding for v0.2.0 (Skills Hub + public Go SDK). |
| **Build artifacts** | — | `bin/`, `dist/` | gitignored. `bin/condurad` is rebuilt locally (21MB, dated session-day). |

**Source-of-truth hierarchy (per the human's convention):**
1. **User's direct instructions** — highest priority.
2. **`condura-mind/MISSION.md`** — the spec (32 sections, 1269 lines).
3. **`condura-mind/FOOTHPATH.md`** — live state ledger (548 lines, 3 entries).
4. **`condura-mind/LOGBOOK.md`** — append-only session history (1623 lines).
5. **`condura-mind/STYLE.md`** — the operating manual (1193 lines).
6. **`condura-mind/docs/`** — architecture, ADRs, runbooks, roadmap, analysis.
7. **Code** — the proof.

> ⚠️ The human told us on 2026-06-07: "MISSION.md and CLAUDE.md are the
> same file going forward. Treat MISSION.md as authoritative."
> On 2026-07-06 (commit `e70d552`) this was consolidated:
> `condura-mind/CLAUDE.md` is now a one-line redirect stub; the original
> content lives at `condura-mind/CLAUDE.md.legacy`. **Read MISSION.md;
> never read CLAUDE.md expecting substance.**

---

## 4. The Internal Go Packages — 60+ Subsystems

All Go code lives in `condura-app/internal/`. Each package is one
self-contained subsystem with at least one `_test.go`. Total ≈ 350 Go files
+ 130 test files. The `daemon` package alone is 66 files / 15,837 LOC.

### 4.1 Safety & Survival (the spine)

These packages implement `MISSION.md` §2 (the seven non-negotiable invariants)
and §5 (the seven technical non-negotiables). Without them, every non-READ
action is denied at runtime.

| Package | LOC | Files | Purpose |
|---|---|---|---|
| `blastradius` | 246 | 2 | READ / WRITE / NETWORK / DESTRUCTIVE classification. |
| `gatekeeper` | 1949 | 8 | **Deterministic rules engine.** Reads `~/.condura/policy.yaml`. Default-deny. |
| `halt` | 789 | 4 | Kill switch. Hard hotkey + watchdog + `NetworkGuard` interface + `InProcessGuard` impl. |
| `anomaly` | 802 | 6 | Behavioral anomaly detector (speed / loops / duration / new endpoints). |
| `audit` | 2044 | 4 | HMAC-chained append-only log. SHA-256 chain, 90-day retention, secret redaction. |
| `sanitize` | 2005 | 10 | Model isolation sanitizers (shell / Python AST / paths / URLs / bodies). |
| `sensitive` | 417 | 2 | Sensitive site detector (banking / health). |
| `failover` | 854 | 3 | Per-provider circuit breakers + daily spend monitor + chain runner. |
| `autonomy` | 140 | 2 | Autonomy matrix (per-task / per-app levels). |
| `watchdog` | 497 | 2 | Inactivity auto-halt + audit-before-halt ordering. |
| `permissions` | 1464 | 8 | TCC / UI Automation / Wayland portal probes per platform. **Rewritten 2026-07-06 (15-finding batch)** — cgo on darwin, PowerShell on windows, dbus/portal on linux, all behind an `execProbe` mock seam. 22 new platform tests. |
| `presence` | 617 | 4 | User activity heartbeat. |
| `trust` | 667 | 2 | Per-workspace trust grants. |

### 4.2 Perception (the eyes)
- `perception` (787 LOC) — Selective Perception data model (Strategy enum,
  EnergyMode, SmartCapturer, DirtyTracker, PIIRedactor). **Live event source
  (CGEventTap, AT-SPI) still not wired — v0.2.0.**

### 4.3 LLM & Routing
- `llm` (4189 LOC, 10 files) — `Provider` interface, 11 OpenAI-compat
  providers + dedicated Anthropic + Google impls + 4 local runtimes
  (Ollama, LM Studio, vLLM, LocalAI). Pricing registry + `EstimateCost`.
  **`modelinfo_json_test.go` (untracked, new)** — pricing JSON contract.
- `stream` (1262 LOC) — SSE streaming manager. `Loop.Ask` subscribes-before-
  start, accumulates deltas, persists assistant reply.
- `router` — **DOES NOT EXIST.** v0.2.0 work. v0.1.x uses a single configured
  `providerName` + `model`. `cfg.Router.Priorities` in `default.yaml`
  exists but is unused at runtime; priorities drift was fixed in 2026-07-06
  commit `992a8bd` to list only LLM providers (no more CLI names).

### 4.4 Memory & Skills
- `memory` (1168 LOC) — 3 layers (episodic, semantic, procedural). SQLite +
  FTS5 + sqlite-vec. Encrypted at rest.
- `skills` (1034 LOC) — `agentskills.io` compatible. Auto-create, self-improve,
  bundle, share via Hub, scan for promptware.

### 4.5 Computer Use
- `computeruse` (1616 LOC, 7 files) — 4-tier router (ORAX Eye → mac-cua →
  macOS-MCP → vision CUA). Backend dispatch. Has `ax/` subpackage for AX tree.
- `executor` (850 LOC) — `shell.exec` + `computeruse.*` dispatch with
  re-gate carve-out, timeouts, audit.
- `agent` (2331 LOC, 14 files) — agent loop + planner. `Loop.Ask` is the real
  stream-driven chat/voice loop (gatekeeper → audit → persist → SSE → TTS).
  `computer_use_executor.go`, `gated_executor.go`, `llm_planner.go`,
  `planner.go`, `culoop.go` are all hot files.
- `conductor` (551 LOC) — conductor-level orchestration.

### 4.6 Agent & Delegation
- `delegation` (1073 LOC, 6 files) — 8 sub-agent CLIs, `GatedRunner`,
  semaphores, file coordination, pending action queue.
- `pending` (888 LOC) — pending action queue (delegation requests awaiting
  approval).

### 4.7 Storage, IPC, Config
- `storage` (1550 LOC) — `modernc.org/sqlite` (pure Go) + AES-256-GCM
  column encryption. Schema v6.
- `ipc` (2324 LOC, 6 files) — JSON-RPC 2.0 server (HTTP + WebSocket),
  bearer-token auth, batch + notifications, typed Go client. **The
  `registerMethods` source-of-truth** for the daemon's public RPC surface.
- `config` (2878 LOC, 5 files) — YAML loader, env-override, `Validate()`,
  router-drift test (`router_drift_test.go`).
- `secrets` (1175 LOC) — OS keyring (`zalando/go-keyring`) with file fallback.
  **Known flake:** `TestNew_NoFilePath_Auto` fails on bare macOS dev boxes
  (skips on CI via `if os.Getenv("CI") != ""`).
- `api_key` (1672 LOC) — manager over storage + secrets; OAuth interface;
  Google PKCE. **OllamaLocalSentinel** for the no-API-key local case.
- `version`, `logger`, `health`, `lockfile`, `sse`, `status`, `crash`,
  `safego` (the `safego.Go` migrator — production `go` launches go through
  here now, **0 bare `go` left in prod**) — infrastructure.

### 4.8 User-Facing & Ecosystem
- `onboarding` (1461 LOC, 7 files) — 4-screen wizard (EULA → Permissions →
  Hotkey → Ready). **PermissionsScreen gate rewritten 2026-07-06**: now
  enforces `accessibility || screen_recording` before continuing (was
  `atLeastOneGranted || !busy` — effectively disabled).
- `account` (1794 LOC) — email magic link + OAuth (Google / GitHub / Apple).
- `channels` — Telegram (real), Signal/WhatsApp/iMessage (stubs).
- `reach` (1765 LOC, 9 files) — channel ecosystem plumbing.
- `hub` (984 LOC) — Hub client (publish + browse skills).
- `sync` (3004 LOC, 13 files) — P2P encrypted sync (Ed25519 identity, mDNS,
  Kademlia DHT, Noise XX, CRDT). Device pairing via QR + PIN + TTL.
- `replay` (932 LOC) — 24h action replay, scrubbable timeline.
- `session`, `conversation` — chat history, sessions.
- `voice` (2923 LOC, **25 files**) — STT (whisper.cpp local + OpenAI cloud) +
  TTS (OpenAI, ElevenLabs, native `say`).
- `hotkey` (816 LOC) — `github.com/atotto/carbon` (mac) +
  `golang.design/x/hotkey` (cross-platform).
- `overlay` (255 LOC) — overlay window state machine.
- `tray` (443 LOC) — system tray / menu bar.
- `window` (326 LOC) — window control (Wails owns actual windows).
- `backup` (2380 LOC, 10 files) — auto-backup scheduler, restore, rollback.
- `uninstall` (811 LOC) — auto-backup before uninstall.
- `updater` (1316 LOC, 13 files) — auto-update via `manifest.json` (Ed25519
  signed).
- `adaptive` (1400 LOC, 11 files) — user-adaptive engine (Honcho-style user
  model, dialectic proposer/critic/adjudicator, predictor, strength slider).
- `i18n` (495 LOC) — backend messages via `golang.org/x/text/message`.
- `mcp` (548 LOC) — MCP client/server (10k+ servers consumable).
- `telemetry` (421 LOC) — local-only (no cloud).
- `tui` (1499 LOC, 4 files) — terminal UI bindings (Bubble Tea).

### 4.9 The 4-Tier Computer-Use Router (`MISSION.md` §11.2)
1. **ORAX Eye** — structured AX tree (free, fast, ~50ms). Primary.
2. **mac-cua** — background-first, `CGEventPostToPid`. Apache 2.0.
3. **macOS-MCP** — comprehensive foreground interaction. MIT.
4. **Vision CUA** — Anthropic Computer Use or OpenAI CUA, ~$0.02-0.05/action.
   **Last resort, opt-in only** (currently disabled per Phase 17 Rec 2).

### 4.10 The 8 Default Sub-Agents (`MISSION.md` §13.2 + FOOTHPATH §8)

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

### 4.11 The 15 LLM Backends (current, `default.yaml` aligned)

**11 cloud APIs:** Anthropic, OpenAI, Google, xAI, Mistral, DeepSeek,
OpenRouter, Groq, Together, Fireworks, plus a **Custom OpenAI-compatible
slot**. **4 local runtimes:** Ollama, LM Studio, vLLM, LocalAI.

Default Ollama model is now `llama3.3` (was the nonexistent `llama4`,
fixed 2026-07-10 commit `807ad88`; pricing registry extended with
`llama3.3` and `llama3.1` in commit `b845730`).

### 4.12 The 4 User-Facing Binaries (`condura-app/cmd/`)
- `condurad/` — daemon. Owns storage / keyring / LLM registry.
- `condura/` — CLI client. Subcommands: `ping`, `version`, `config`, `llm`,
  `apikeys`, `delegate`, `sync`, `hub`, `skills`.
- `condura-tui/` — Bubble Tea terminal UI.
- `condura-gui/` — Wails desktop shell. Embeds the daemon via
  `internal/daemon.Run()`. The Wails build auto-generates
  `frontend/wailsjs/`; never commit those bindings (gitignored).
- `gen-update-manifest/` — release manifest generator.

A `build_all_test.go` meta-test compiles all four binaries in CI.

---

## 5. The Frontend — Three Generations Coexist (2026-07-12)

This is the **biggest drift** from the 2026-06-24 anchor. The GUI has been
rebuilt multiple times; everything coexists in the working tree.

### 5.1 The Live Mount

**`src/App.svelte` → mounts `MeridianShell.svelte`** (the only mount point;
the file's comment says verbatim *"MeridianShell — sole product GUI mount."*).
Wails loads `index.html` → `src/main.ts` → `App.svelte` → `MeridianShell`.

### 5.2 The Three Generations

| Generation | Path | State | Purpose |
|---|---|---|---|
| **`Meridian` (LIVE)** | `src/lib/shell/meridian/Meridian*.svelte` | **Production shell** | 15 surfaces: Shell, Arc, Dock, Chat (Ask), Hub, Skills, Sync, Audit, Replay, Channels, Delegation, Account, Settings, About, Palette, Consent, Halt, Toasts, Page. `meridian.css` + `onboarding-meridian.css`. |
| **`v2` (design system)** | `src/lib/v2/*.svelte` + `tokens.css` + `motion.css` + `reset.css` | **Newest primitives** | `Surface`, `Ink`, `Stack`, `Inline`, `Rule`, `Button`, `Avatar`, `Chip`, `Switch`, `Sidebar`, `Sidebar`, `SettingsDocument`, `ChatSurface`, `ConsentModal`, `StatusBar`, `Eyebrow`, `Glyph`. Also 15 demo routes under `routes/dev/V2*Demo.svelte`. |
| **`condura` (medium)** | `src/lib/condura/*.svelte` + `condura.css` | **Reference design system** | Has its own APPFLOW.md, DESIGNLANG.md, DIRECTION.md, MOAT.md, TEARDOWN.md. ~50 components. Specs dir. **Largely superseded by Meridian.** |
| **`v1` (legacy)** | `src/lib/components/v1/*.svelte` | **Phased out** | ~50 widgets (Button, Card, ChatSurface, Sidebar, etc.). Mostly dead but `SettingsPane.svelte` still M-staged (under active replacement). |
| **`components/` (current)** | `src/lib/components/*.svelte` | **Mixed utility widgets** | `ui/` (22 primitives — Button, Card, Dialog, Toast, Sheet, Select, etc.) + `onboarding/` (5 screens) + top-level (AccountMenu, ConsentModal, HotkeyRecorder, LiveTranscript, OverlayPrompt, PairingModal, PendingActions, PublishModal, Sidebar, SignInPanel, StatusRail, TitleBar, Toasts, VoiceOrb). |
| **`routes/` (legacy)** | `src/lib/routes/*.svelte` | **V1 pages** | About, Audit, Channels, Chat, ChatV1, Delegation, Hub, Replay, Settings, Skills, Sync. **In the working tree but being phased out.** |
| **`shell/` (mostly dead)** | `src/lib/shell/*.svelte` | **Mostly dead** | `index.ts` is `export {}` ("Old Living Paper barrel — unused by Meridian mount"). `LivingPaperShell.svelte`, `NavOrbit.svelte`, `StatusThread.svelte`, `TopBar.svelte` are all pre-Meridian shells. |
| **`living/` (mostly dead)** | `src/lib/components/living/*.svelte` | **Old** | BlurReveal, InkReveal, InkText, MagneticButton, PaperCard, PaperSurface, PollenNode, PollenSpark, PulseDot, QuillCursor, SynapseThread, ThreadLink, WordReveal. |

### 5.3 The 12 Meridian Routes (current `routes.ts`)

`#/` Chat (Ask) · `#/skills` Skills · `#/hub` Hub · `#/sync` Sync ·
`#/audit` Audit · `#/replay` Replay · `#/channels` Channels ·
`#/delegation` Delegation · `#/account` Account · `#/settings` Settings ·
`#/about` About · `#/palette` Palette (in-product theme picker).

### 5.4 Frontend Stack
- Svelte 5 + TypeScript + Vite 5 + vitest.
- `packageManager: npm@10`. **No pnpm-lock** — the lockfile policy is `npm`
  only, enforced by `make check-lockfiles`.
- Dependencies: `dompurify ^3.4.11`, `highlight.js ^11.10.0`, `marked ^14.1.3`,
  `qrcode ^1.5.4`. **Markdown safety:** `marked.parse()` is wrapped with
  DOMPurify in Chat + LiveTranscript (XSS fix, commit `6b05ecd`).
- Build scripts: `dev`, `build` (vite build), `preview`, `check`
  (svelte-check), `test` (vitest run).
- Vite config: `base: '/'` in serve, `'./'` in build (Wails asset-server
  quirk). Aliases: `$tokens`, `$components`, `$lib`. Proxy `/api` + `/events`
  to daemon on `127.0.0.1:7666` for browser previews (Cursor Simple
  Browser / Playwright).
- Fonts loaded from Google Fonts: Figtree (UI), IBM Plex Mono (code), Sora
  (display).

### 5.5 Frontend Store Architecture

24 Svelte 5 runes-based stores under `src/lib/stores/`:
`account`, `apikeys`, `audit`, `consent`, `conversation`, `daemon`, `halt`,
`hub`, `notifications`, `onboarding`, `overlay`, `pending`, `replay`,
`settings`, `spend`, `sync`, `trust`, `update`. Plus `init.ts` (store
orchestration). Each `.svelte.ts` is paired with `.test.ts` (recently added
for audit, hub, pending, settings, spend, sync).

### 5.6 i18n (the 6-Locale Reality)

All 6 locales shipped as static JSON files at `static/locales/{en,es,fr,de,ja,zh}.json`,
each ~574 lines / ~32 KB / **572 keys** at the top level. A new
`src/lib/i18n/locales/` mirror (untracked) was added so Vite bundles catalogs
at build time (the prior `public/` fetch fallback was silently failing when
Wails served SPA HTML for `/locales/*.json`).

`lib/i18n.ts` is the canonical loader. Bundled catalogs are always available;
daemon-provided catalogs (`ipc.i18nLocale(locale)`) override when present.
`locale-parity.test.ts` enforces key parity (non-EN must have every EN key).

**Non-EN locale content is mostly English placeholders.** The 2026-06-26
Kimi K2.7 convention was "add English placeholders to non-EN" rather than
omit keys. Real translations are v0.2.0 Crowdin work.

### 5.7 IPC Client Surface

`src/lib/ipc/client.ts` is the typed wrapper. Single `EventEmitter`-based
`ipc` instance with `start()`, `await ipc.call('method', {})`, `ipc.on(...)`,
`ipc.stop()`. WebSocket-style HTTP/1.1 keep-alive; reconnects with backoff.
SSE events on separate `/events` EventSource.

Daemon-side, the canonical RPC surface (from `internal/daemon/` + grep) is
**~80 methods** across:
- `daemon.{ping,info,uptime,pid,capabilities,halt,resume,resume_request}`
- `version`, `config.{get,update,yaml}`
- `apikeys.{set,list,delete}`, `account.{providers,status,oauth_url,oauth_callback,magic_link,logout}`
- `llm.{chat,stream,cancel}`, `agent.{ask,status}`
- `delegate.{spawn,list_agents,list_spawns,cancel}`, `delegation.spawn`
- `gatekeeper.{check,approve,deny,pending_consent}`, `gate.{allow,deny}`
- `audit.{list,export}`, `audit.replay.verify_integrity` (via replay pkg)
- `backup.{create,list,preview,restore,rollback,derive_key}`
- `channels.{list,connect,disconnect,status}`
- `computeruse.{click,type,scroll,read,launch}`, `cu.action`
- `adaptive.{profile,forget,reset}`, `adaptive.strength.{get,set}`
- `sync.{list,pair,approve,...}` (the actual method names vary; check
  `internal/sync/engine.go` for the current set)
- `hub.{search,get,install,publish}`
- `skills.{list,load,create,share}`
- `replay.{list,scrub,verify_integrity}`
- `mcp.{list_servers,install_server,call_tool}`
- `halt.{state,confirm_resume}`, `health.snapshot`
- `condurad.{addr,lock,sock}`, `condura.db`, `acct.db`
- `daemon.update`, `update.{check,apply}`, `uninstall`
- `onboarding.{eula,set_step,probe_power,finish,reset,permissions}`
- `permissions.{probe,list,request}`
- `i18n.{locale,list}`
- `spend.{summary,reset_daily}`, `notifications.{list,send}`
- `presence.{get,set}`, `trust.{list,grant,revoke}`
- `file.{read,write}` (declared but "not yet supported" per FOOTHPATH)

**New (untracked / M-staged) RPC methods** in the working tree: a few
audit/consent/sync extensions — verify by grepping
`internal/daemon/methods*.go` before declaring the surface.

### 5.8 The Web GUI Tests

Vitest is the runner. Test files:
- `src/lib/ipc/client.test.ts` (M-staged, new tests)
- `src/lib/markdown.test.ts` (untracked, new)
- `src/lib/i18n/locale-parity.test.ts` (M-staged, new)
- `src/lib/components/onboarding/FloatingOnboarding.test.ts` (M-staged)
- `src/lib/stores/{audit,conversation,hub,pending,settings,spend,sync}.test.ts`
  (untracked, **6 new store test files** in the working tree)

---

## 6. The Marketing Site (`condura-ui/`)

Next.js 16 + React 19.2 + Tailwind 4 + motion 12 + `isomorphic-dompurify`.
Lives at `condura.app` in production, `localhost:3000` in dev.

**This is the post-reorg successor to the 2026-06-10 "The Score" `web/` site
(which memory still references — that memory is stale on the path; the
content intent is unchanged).** README says `synaptic.app` (stale; domain
is `condura.app`).

App Router pages: `/` (home), `/about`, `/changelog`, `/download`,
`/ecosystem`, `/legal`, `/manifesto`, `/orchestration`, `/privacy`,
`/security`, plus `/api` route. Has `robots.ts`, `sitemap.ts`,
`opengraph-image.tsx`, `twitter-image.tsx`.

`_experiments/` holds archived marketing film HTMLs (kebab-case, not in
sitemap). `.next/` is built output (gitignored).

**Hard constraint (memory):** the website is kept strictly separate from
the technical side (Go daemon + `condura-gui/` Wails GUI) until the
technical side is finished. **Do not wire them together. Do not touch
`condura-app/` or `condura-gui/` Go code in website sessions.** The
website uses **isomorphic-dompurify** for any rendered markdown.

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
   isolation (separate process the agent cannot stop). **v0.1.x uses
   in-process guard** (`halt.NetworkGuard` interface, `InProcessGuard`
   implementation). Real `pf`/`netsh` companion binary is v0.2.0.
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

Key ones to internalize:
- **#1 Name: Condura** (was Synaptic, renamed).
- **#4 Foundation:** from scratch in Go + TypeScript (no Hermes fork).
- **#5 Computer use:** all 3 backends + vision CUA, 4-tier router.
- **#6 Routing:** hybrid with memory (cost-first cascade, bias toward what
  worked). v0.2.0; v0.1.x is single-provider-per-session.
- **#8 Hotkey:** user must set on first run (no default; suggestions:
  Option+Option, Cmd+Shift+Space, Ctrl+Space, Ctrl+Ctrl).
- **#11 Languages at v0.1.x:** English + Spanish + French + German +
  Japanese + Mandarin (6). Keys present in all 6 JSONs.
- **#13 Launch strategy:** Public v0.1.0 + v0.1.1, all in — Product Hunt +
  Hacker News + Reddit (r/singularity, r/LocalLLaMA, r/AI_Agents) on
  same day.
- **#15 Provider down:** auto-failover — Ollama local first, then any
  configured backup key.
- **#16 Multi-machine sync:** P2P encrypted sync, no central server.
- **#17 Uninstall behavior:** auto-backup before uninstall to
  `~/Documents/condura-backups/`.
- **#23 Concurrency:** default 2 parallel sub-agents, max 5, user-configurable.
- **#24 Autonomy:** default cautious (warn before any action).
- **#25 Uncertainty:** ask user immediately ("I'm 60% sure you want X. Proceed?").
- **#26 Energy budget:** refuse, force user decision.
- **#27 Daemon autostart:** auto-start on login.
- **#30 User account:** email + magic link (for hub, donations, support;
  P2P sync needs no account).
- **#34 Multi-install:** block second install.
- **#35 Wake word:** "hey condura" (code + MISSION aligned as of
  2026-07-10 commit `02ec07f` + P3-13 cleanup).
- **#36 EULA:** Freeware EULA — free personal + commercial, no
  redistribution, revocable for abuse.

---

## 9. Current Status (as of 2026-07-12)

**Per the working tree (`main` @ `d6c2e78`), the most recent LOGBOOK entry
(2026-07-12), the `phase-15-ship-readiness` worktree, and the 3 stashes.**

### 9.1 v0.1.0 + v0.1.1 are shipped
- **v0.1.0** (Phase 13, release/distribution) — published with signed
  auto-update (`manifest.json`), GoReleaser packages, GUI installers.
- **v0.1.1** — adds fail-closed destructive consent, watchdog-by-default,
  talk→act dispatch (N2 Path A), honest marketing/README pass. Dist
  artifacts in `dist/` are `condura_0.1.1-SNAPSHOT-08a803b_*`.
- **Tags:** `v0.1.1` and `v0.1.0` exist on remote.
- The 2026-07-06 Phase 14 cleanup session landed 9 commits on `main`,
  got CI green, and fixed the post-reorg drift cascade.

### 9.2 What's Working (Tier-3 verified end-to-end)
- **Daemon** — boots, migrates SQLite schema to v6, initializes ~60
  subsystems, listens on TCP/Unix, exposes ~80 JSON-RPC 2.0 methods,
  persists across restarts.
- **Onboarding** — 4-screen wizard (EULA → Permissions → Hotkey → Ready)
  with proper gating: PermissionsScreen now enforces
  `accessibility || screen_recording` (commit `5708f51` related fix).
- **Chat** — `llm.chat` and `llm.stream` (SSE) working. Ollama local
  tested with `llama3.3` (the 2026-07-10 default fix).
- **Permissions** — all 3 OS probes rewritten 2026-07-06 with honest
  StatusUnknown contracts; cgo on darwin, PowerShell on windows, dbus
  on linux; `execProbe` mock seam + 22 new platform tests.
- **Audit** — HMAC-chained, `replay.verify_integrity` works.
- **Backup** — auto-backup scheduler creates
  `condura-backup-<ISO-date>.zip` on daemon startup.
- **GUI** — Meridian shell mounts via `App.svelte`. svelte-check clean.
  Tool calls render. DOMPurify wraps `marked.parse()` (XSS fix).
- **i18n** — all 6 locales present (572 keys, ~32 KB each). Parity test
  enforces coverage. Non-EN content is mostly English placeholders
  (translations v0.2.0).
- **Voice loop** — `Loop.Ask` is a real stream-driven loop
  (gatekeeper → audit → persist → SSE → TTS). N2 Path A "talk → act"
  dispatch shipped in v0.1.1.

### 9.3 The Honest Backlog (v0.2.0+)
Per `condura-mind/docs/roadmap-v0.2.0.md` and `MISSION.md` §33.5.2:
- **Subscription OAuth** (ChatGPT Plus, Claude Pro, SuperGrok) — 2-3
  weeks.
- **Hardened Layer 3** (`pf`/`netsh` companion binary, separate OS
  process the agent cannot stop).
- **CGEventTap / AT-SPI dirty tracking** wired to perception.
- **MCP UI** (`Mcp.svelte` route).
- **Crowdin i18n sync** + real translations for 5 non-English languages.
- **Public Hub + Dashboard deploy** (`hub.condura.app` and
  `condura.app/dashboard` as separate Next.js apps). The
  `condura-hub/` folder is currently a read-only stub.
- **Public SDK** (`condura-sdk/` is a read-only stub).
- **Vision CUA opt-in** (currently disabled).
- **Non-macOS voice** via cloud STT (GUI overlay for Windows/Linux is
  also v0.2.0 — only macOS has full GUI).
- **`file.*` executor dispatch** (currently "not yet supported").
- **Hybrid LLM router** (`internal/router/` package — does not exist).
- **Wave / DAG orchestration** across sub-agents (v0.1.x spawns individual
  sub-agents only).
- **Vector embeddings / semantic recall.**

### 9.4 Drift Between Spec and Implementation (Known)
- **`internal/router/`** — spec describes a hybrid-with-memory router;
  package does not exist. v0.2.0 work.
- **`cfg.Router.Priorities`** — `default.yaml` field exists, was drifted
  against `knownProviders()`; fixed 2026-07-06 (commit `992a8bd`) and
  guarded by `internal/config/router_drift_test.go`.
- **OAuth subscription flows** — marketing copy mentions ChatGPT Plus /
  Claude Pro / SuperGrok; backend stubs return "coming in v0.2.0".
- **i18n locale JSONs** — all 6 exist; non-EN content is English
  placeholders, not real translations.
- **Marketing copy** still has 10k+ MCP servers claim, real Signal/WhatsApp
  /iMessage, etc. The website is a separate track; align with backend in
  v0.2.0 coordinated pass.
- **`condura-receipt/`** was an orphan studio stub never merged to main;
  dropped 2026-07-10 (P3-10). Don't recreate it.

### 9.5 The Single Next Human Action
**On-device verification on clean macOS, Windows, Linux machines** per
`condura-mind/docs/on-device-verification.md`,
`condura-mind/docs/macos-verification-runbook.md`, and the Phase 15
checklist (`docs/phase15-verification.md`). This is the gate before public
launch. The human must drive the physical keyboard. An agent cannot do this.

---

## 10. The Working Style — How Every AI Must Operate

Per `STYLE.md` and the human's convention. This is non-negotiable.

### 10.1 Identity & Honesty
- **Byline whatever actually ran.** Don't impersonate other models. The
  harness varies by session; truth over roleplay. The user values
  honesty over roleplay. (Per memory: declined on 2026-06-08, the
  `dynamic-workflow-emulator` skill's fake swarm is also declined for the
  same reason — scope to maturity of work / one goal per session.)
- LOGBOOK byline format: `AI Model: <model> (<harness>)` or
  `AI Model: <model> — <one-line task>` for terse entries.
- This session's byline: **GLM 5.2 by Z.ai via Claude Code.**

### 10.2 The Three-Tier Verification Ladder
**A green test is not proof the feature works.** Every shipped feature
passes all three:
- **Tier 1 — Unit tests.** Single package, controlled fixture.
- **Tier 2 — Integration / E2E test in Go.** Real `initSubsystems`,
  real `ipc.Server`, real SQLite.
- **Tier 3 — Runtime smoke test.** `go build`, run the actual
  `condurad` binary, drive it with `curl` over its real RPC surface,
  inspect the real on-disk state with `ls`, `sqlite3`, `unzip`.

The 2026-07-06 permissions-fix session adds a 4th de-facto tier: **CI
matrix on all 3 platforms × 2 archs** (darwin / linux / windows ×
amd64 / arm64). A green local run is not enough; the matrix must be green.

**A mediocre AI ships a passing test suite. A partner AI ships a passing
test suite AND runs the binary to confirm AND watches the CI matrix.**
— `STYLE.md` §0.

### 10.3 Commit Policy
**The human commits manually by default.** AI must not commit on his
behalf. **Exception (added 2026-06-09):** for green sub-phases, AI may
commit directly to `main` with conventional-commit messages and push at
end of session. Use `Co-Authored-By: Claude <noreply@anthropic.com>`
trailer (or whatever the harness dictates). One commit = one logical
change per `STYLE.md` §10.

### 10.4 Context Loading per Session
- **Deep on code** — read every file you might touch.
- **Light on docs** — re-read `MISSION.md`, `LOGBOOK.md`, the specific
  architecture doc relevant to the task. Don't re-read every doc.
- Always re-read the memory pointers (`synaptic-identity`,
  `synaptic-canon-files`, `synaptic-conventions`) from memory.
- **For project-wide orientation, read this file (`synapse/understanding.md`)
  first.** It is faster and more complete than the MISSION + LOGBOOK +
  FOOTHPATH dance.

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
- Match `MISSION.md`. Section headers like `## 4. The 36 Locked Decisions`
  (numbered, Title Case, period). 5-line file headers in code. Long,
  didactic, opinionated, exhaustive.
- LOGBOOK entries are denser — see `MISSION.md` §30.3 for the format.

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

1. **`condura-mind/CLAUDE.md` is a redirect stub** (since 2026-07-06).
   Historical content preserved at `condura-mind/CLAUDE.md.legacy`.
   `LOGBOOK.md`, `CONTRIBUTING.md`, `docs/README.md`, `.goreleaser.yml`
   release notes still say "read CLAUDE.md" — intentional drift per
   append-only rule. `MISSION.md` H1 still reads `# Condura — CLAUDE.md`
   — preserves link targets.

2. **Wake word is "hey condura"** in MISSION + code defaults. Deprecated
   `hey_synaptic` alias remains in `internal/voice/modelmgr` for old
   configs only. ONNX asset URL remains on the pre-rebrand HuggingFace
   path (model detects the condura phrase).

3. **OAuth scheme `condura://`** — was `synaptic://`, renamed.

4. **Backup directory `~/Documents/condura-backups/`** — aligned with
   MISSION §24.1 / decisions #17 and #28. Rename closed 2026-07-10
   (P3-13 brand-honesty pass).

5. **Two consent namespaces:** `gatekeeper.*` (canonical) and
   `safety.consent.*` (DEPRECATED alias). GUI uses `gatekeeper.*`.

6. **`internal/router/`** does not exist. `cfg.Router.Priorities` in
   `default.yaml` is parsed but not used at runtime. v0.2.0 work.

7. **`subs.Executor` is nil** when `cuComps` is nil (no LLM configured),
   blocking shell-only sub-agents. Open question for next session.

8. **`internal/secrets.TestNew_NoFilePath_Auto`** — passes 3/3 in CI but
   historically fails 1/3 on bare macOS. Tracked, not blocking.

9. **Wails build under Go 1.26+** has duplicate `_OBJC_*_AppDelegate`
   symbols — Wails v2.12.0 upstream issue, not a project bug. Local Go
   1.26+ devs should pin to 1.25.x via `go.work` toolchain directive.
   CI on Go 1.25.11 is green. Root `go.mod` is on `go 1.25.12`.

10. **`internal/daemon/methods_phase12.go` is 681 lines.** `subsystems.go`
    is 1786 lines. These are intentionally large for daemon startup
    orchestration and `registerMethods`; do not split without a clear
    reason — the size is a feature, not a bug, for this layer.

11. **GUI Build smoke check** is marked `continue-on-error: true` in CI
    (commit `3535692`). Known flake on darwin/arm64; tracked, not
    blocking.

12. **`default.yaml` `cfg.Router.Priorities` lock was relaxed** for the
    2026-07-06 router-drift fix (commit `992a8bd`). The new lock is:
    *must match `knownProviders()` in `internal/daemon/providers.go`*,
    enforced by `internal/config/router_drift_test.go`.

13. **Three Svelte 5 GUI generations coexist** — see §5.2. Meridian is
    live; v2 is the design-system primitives; condura/ is reference;
    v1 and shell/ and routes/ and living/ are legacy/dead. **Don't add
    new screens in `components/v1/` or `routes/` — they are being
    phased out.** New screens go in `shell/meridian/` (or extend
    `v2/` if it's a primitive).

14. **Three stashes** exist (see §11.16). `stash@{2}` on main is labeled
    *"user WIP + v2 design system, preserve through merge to main"* —
    **never drop this one**. It's the user's live WIP.

15. **`bin/condurad` is rebuilt locally** every session-day. ~21 MB.
    Don't commit binaries (gitignored); use `make build` or
    `scripts/release-dry-run-local.sh`.

16. **The phase-15-ship-readiness worktree is on the pre-reorg layout.**
    It has `app/`, `web/`, `scripts/` at root — not `condura-app/`,
    `condura-ui/`, `condura-ops/scripts/`. Its README still says
    `synaptic.app` and `./bin/synapticd`. **Don't be confused by its
    contents — when reading files in that worktree, mentally remap
    paths to the new layout.** Its `git diff main..HEAD` is 1085 files
    / 6314 insertions / 118,488 deletions — that delta is the reorg,
    not substantive work.

17. **The website (`condura-ui/`) is a separate track.** Don't touch
    `condura-app/` or `condura-gui/` Go code in website sessions. The
    website uses `isomorphic-dompurify` for any rendered markdown.

18. **`condura-hub/` and `condura-sdk/`** are read-only stubs (just
    READMEs). Don't scaffold anything there until v0.2.0 work begins;
    use `condura-ui/_experiments/` for experimentation.

19. **`condura-studio/condura-receipt/`** was dropped (orphan stub never
    on main). Don't recreate it. Receipt visuals now live in
    `condura-studio/condura-thread/` and the desktop Replay surface.

20. **`condura-brand/` is honestly scoped as a starter kit, not a full
    brand package** (P3-11 honesty pass, 2026-07-10). Tokens are the
    canonical artifact; palette / motion / logos are reference; fonts
    are policy-only (no binaries); assets is media (MP4s gitignored).
    The `make brand` sync target is aspirational — until it lands,
    edit both `condura-brand/tokens/` and `condura-gui/frontend/src/lib/tokens/`
    by hand.

21. **Cross-package import lock** between `internal/config` and
    `internal/daemon` (they already import each other transitively).
    The router-drift test had to use a duplicate `knownProvidersMirror`
    instead of importing the canonical list. v0.2.0 SDK migration
    (internal/ → pkg/) should break this cycle.

22. **CI workflow discovery** has been silently broken since 2026-07-04:
    `condura-ops/ci/workflows/ci.yml` and `condura-ops/ci/workflows/release-verify.yml`
    exist but GitHub Actions only reads `.github/workflows/`. The orphan
    copies are drifting; the actual workflows at `.github/workflows/`
    are what runs. Same for CODEOWNERS / dependabot.yml. Fix in a
    v0.2.0 sweep — not blocking, but breaks security governance.

23. **Synaptic→Condura migration residual:** ~149 Synaptic mentions
    remain in `LOGBOOK.md`, `CLAUDE.md.legacy`, and inline code comments.
    User ratified "Hold — append-only, deferred" (2026-07-06). Don't
    bulk-rewrite.

---

## 12. The Cross-Reference Index

### 12.1 If you need to know…
- **The project spec** → `condura-mind/MISSION.md` (canonical, 1269 lines,
  32 sections, 36 locked decisions, 7+7 invariants, 4-tier CUA, 8
  sub-agents, 6 locales, hybrid router spec).
- **The state right now** → `condura-mind/FOOTHPATH.md` (548 lines, 3
  entries: workspace status → UI ship-gaps → audit-driven ship-readiness).
  **Last entry is from before Meridian — the Meridian + permissions batch
  work is not yet captured in a new FOOTHPATH. Append a FOOTHPATH 4
  when shipping the next ship-readiness batch.**
- **The session history** → `condura-mind/LOGBOOK.md` (1623 lines,
  append-only, last entry 2026-07-12).
- **The operating manual** → `condura-mind/STYLE.md` (1193 lines).
- **The agent registry** → `condura-mind/AGENTS.md` (5 custom agents:
  Production-Ready Analysis, System Architect, Implementation Engineer,
  AI Systems Specialist, Security & Privacy Guardian).
- **The constitution + canvas** → `condura-mind/CHANGELOG.md` and
  `condura-mind/docs/`.
- **The architecture** → `condura-mind/docs/architecture/00-overview.md`
  through `09-ipc.md` (10 files).
- **The 5 ADRs** → `condura-mind/docs/adr/0001-go-over-python.md`
  through `0005-p2p-sync.md`.
- **The current phase state** → `condura-mind/docs/phase14-completion.md`
  and `condura-mind/docs/phase15-verification.md`.
- **What's deferred to v0.2.0+** → `condura-mind/docs/roadmap-v0.2.0.md`
  and `MISSION.md` §33.5.2 (the spec-debt ledger).
- **How to build / install / release** → `Makefile`,
  `condura-ops/scripts/install.sh`,
  `condura-ops/scripts/build-gui.sh`,
  `condura-ops/scripts/release-dry-run-local.sh`,
  `condura-ops/release/goreleaser.yml`,
  `condura-mind/docs/release-runbook.md`,
  `condura-mind/docs/release-keys.md`.
- **The release verification** → `condura-mind/docs/phase15-verification.md`,
  `condura-mind/docs/on-device-verification.md`,
  `condura-mind/docs/macos-verification-runbook.md`.
- **The threat model** → `condura-mind/docs/threat-model-reach.md`.
- **User-facing guides** → `condura-mind/docs/user-guide/`,
  `condura-mind/docs/guides/`.
- **The most recent work** → last 5 entries of `LOGBOOK.md`.
- **Pre-July 2026 history** → `condura-mind/docs/logbook-archive/LOGBOOK-2026-06.md`
  (relocated, not deleted — append-only contract).

### 12.2 If you need to do…
- **Build the daemon:** `make build` → `bin/condurad`, `bin/condura`,
  `bin/condura-tui`.
- **Build the GUI:** `make build-gui` (uses `condura-ops/scripts/build-gui.sh`).
- **Run tests:** `go test -count=1 -race -timeout 300s ./...`.
- **Run lint:** `golangci-lint run --timeout=5m ./...`.
- **Boot the daemon:** `condurad -config /tmp/c.yaml -data-dir /tmp/data
  -listen "tcp://127.0.0.1:18600"`.
- **Ping it:** `curl -X POST http://127.0.0.1:18600/api -H "Content-Type:
  application/json" -d '{"jsonrpc":"2.0","id":1,"method":"ping","params":{}}'`.
- **Append a LOGBOOK entry:** follow the format in `MISSION.md` §30.3.
- **Append a FOOTHPATH entry:** follow the template at the end of
  `FOOTHPATH.md`.
- **Add a new LLM provider:** edit `internal/llm/` (add to `providers.go`
  + pricing registry in `modelinfo.json`), and add the name to
  `internal/daemon/providers.go:knownProviders()` + `default.yaml`'s
  `router.priorities` (the drift test will fail if you don't).
- **Add a new Meridian screen:** edit `condura-gui/frontend/src/lib/shell/meridian/`,
  add to `routes.ts`, mount in `MeridianShell.svelte`. Reuse v2 primitives
  from `condura-gui/frontend/src/lib/v2/` for design-system consistency.

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
6. **Read `condura-mind/docs/architecture/00-overview.md`** for the
   mental model.
7. **Then run the binary.** Boot it, ping it, install an API key via
   `apikeys.set`, spawn a sub-agent, approve a pending action. **The
   binary is the source of truth, not this doc.**
8. **Check `git status` and `git log --oneline -10`** — the human may
   have uncommitted work in progress. **Check the stash list** —
   `stash@{2}` on main is the user's live WIP, do not drop.
9. **Check `git worktree list`** — `.worktrees/phase-15-ship-readiness/`
   is on the pre-reorg layout; don't confuse its paths.
10. **Ask if anything is ambiguous.** The user has explicitly invited
    unlimited questions.
11. **Append to LOGBOOK.md before you finish.** Format per `MISSION.md`
    §30.3.

---

**Last updated:** 2026-07-12 by the deep-analysis session.
**Byline:** GLM 5.2 by Z.ai via Claude Code.
**Status:** Anchor document. Cross-references all primary sources. Use
this to orient, then drill into the specific source for the detail.
