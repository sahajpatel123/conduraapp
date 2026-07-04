# Session Context — Condura v0.1.0 Ship-Readiness

> Generated 2026-06-22. Read this cold to understand the full state of the project, every decision made, every fix applied, and every remaining gap. This is the context you need to continue work without asking "what happened before."

---

## 1. Project Identity

**Condura** (formerly Synaptic) — a free desktop app that summons every AI tool on your computer with one hotkey. No account, no subscription, no data leaves your machine.

- **Repo**: `github.com/sahajpatel123/conduraapp` (private)
- **Binary**: `condurad` (Go daemon) + Wails GUI (Svelte 5 frontend)
- **Website**: `condura.app` (Next.js 14, `web/`)
- **License**: Proprietary source, free binary (Synaptic Freeware EULA v1)
- **Current version**: v0.1.0 (pre-release, not yet shipped)

---

## 2. Architecture (what you need to know)

```
condurad (Go daemon, ~21MB)
├── JSON-RPC 2.0 over HTTP + WebSocket + SSE
├── SQLite storage (migrations v1→v6)
├── 12 LLM providers (Anthropic, OpenAI, Google, xAI, Mistral, DeepSeek, etc.)
├── 8 sub-agent CLIs (Claude Code, Codex, Antigravity, OpenCode, Kilo, Hermes, Gemini, Ollama)
├── Safety layer: Gatekeeper (deterministic rules), Blast Radius, Audit (HMAC-chained), Anomaly Detector, Kill Switch (3 layers)
├── Circuit breaker + spend monitor per provider
├── Workspace trust (git-root-based, per-folder YAML)
├── Watchdog (inactivity timer → auto-halt)
├── Twin-snapshot verification (AX tree before/after, abort on diff)
├── Pending actions queue (sub-agent ActionRequests → gate → approve → execute)
├── SSE broker (events to GUI)
├── P2P sync (pairing, QR code, PIN)
├── Skills hub client
├── Voice pipeline (whisper.cpp local + wake word)
├── Adaptive engine (observer + dialectic + predictor)
└── Onboarding state machine (EULA → Permissions → Hotkey → Ready)

Wails GUI (Svelte 5, ~22MB)
├── 10 routes: Chat, Settings, Audit, Replay, Delegation, Channels, Sync, Hub, Skills, About
├── 17 components: Sidebar, OnboardingWizard, ConsentModal, PendingActions, VoiceOrb, etc.
├── 16 reactive stores, ~70 typed IPC methods
└── 6 i18n locale files (translations are stubs — all English)

Website (Next.js 14, `web/`)
├── Pages: Home, Download, Orchestration, Ecosystem, Security, Manifesto, Changelog, Legal, Privacy
├── Shell: GlobalNav (scroll-hide/reveal, mobile hamburger), SiteDock (5 quick-access), RouteProgress
└── Motion library: 30+ SVG icons, liquid-glass surfaces, mature-button styling
```

---

## 3. What Was Fixed in This Session

### Website fixes (6 commits, all CI green)

| # | Fix | Commit |
|---|---|---|
| 1 | Hero rewrite: "One hotkey. Every AI you own." replaces "intelligence layer" | `02cc6b1` |
| 2 | Ecosystem page: 12 real LLM providers + 8 agent CLIs replace fake web-dev grid | `2f93d6e` |
| 3 | Legal page: proper 10-section EULA replaces fake "Decrypting local document" | `9318c5b` |
| 4 | Mobile hamburger menu in GlobalNav | `28759c6` |
| 5 | Privacy policy page + 404 page + social proof row in hero | `b88ffd2` |
| 6 | Brand naming: `synaptic.dmg`→`condura.dmg`, `discord.gg/synaptic`→`discord.gg/condura` | `02cc6b1` |

### Backend fixes (4 commits, all CI green)

| # | Fix | Commit |
|---|---|---|
| 7 | `config.update` and `telemetry.setEnabled` now persist to disk via `Loader.Save()` | `e72e69f` |
| 8 | SSE auth via one-time ticket exchange (no more `?token=` in URL) | `917c916` |
| 9 | Circuit breaker + spend cap wired into `llm.chat` and `stream.Manager.Start` | `42a816c` |
| 10 | Lint cleanup: gofmt, errorlint, unused code, migration test (5→6 versions) | `6e5df7f` |

### Phase 17 fixes (user implemented, verified by me)

All 9 audit-driven ship-blockers closed: watchdog audit-before-halt (B3), ApproveTicket expiry check (A6), Terminal auto-allow removed (B2), apikeys.set auto-enables provider (B1), OnTimeout=queue honored (A7), twin-snapshot verify wired (B4), ActionRequests gated+audited (B5), magic-link host allowlist (R3), default config written on first start (R1+R2).

### Phase 18 (user implemented, verified by me)

Pending-actions executor + confirm-then-run UI for sub-agent ActionRequests. Backend executor dispatches approved actions (shell exec, computer use). Frontend `PendingActions.svelte` shows approve/deny/execute buttons. `f63b163`.

---

## 4. Audit Verifications Performed

### Website audit (44 findings)
- **38/44 accurate**, 3 partially accurate, 2 false, 1 minor naming error
- The 5 "BLOCKERS" in a later Phase 17 audit were **3/5 already fixed** (Phase 17 remediation commits predated the audit)
- 2 genuine blockers remained: B4 (twin-snapshot never called), B5 (ActionRequests never executed)

### Phase 17 verification
- All 9 fixes verified in source code with exact line numbers
- CI had transient lint failure on intermediate commit; fixed by subsequent lint cleanup commit
- Final state: all 14 CI jobs green

### Backend final audit verification
- `go build ./...` clean, `go vet` clean, `golangci-lint` 0 issues
- 64+ packages pass `go test ./...`, zero failures
- CI 14/14 green on `e094431`
- Working tree clean (only pre-existing untracked artifacts)
- **Verdict: backend is in clean, shippable state**

---

## 5. Key Design Decisions Made

| # | Decision | Rationale |
|---|---|---|
| 1 | No CAPTCHA on website | No forms, no writable endpoints, no attack surface |
| 2 | No cookie consent banner | No analytics, no cookies, no tracking — nothing to consent to |
| 3 | `hub.condura.app` (not `hub.synaptic.app`) | Rename is complete everywhere else; don't leave a holdout |
| 4 | Watchdog timer in v0.1.0, separate process in v0.2.0 | Ship simple now, harden later |
| 5 | UUID-based AAD for API keys | Cleaner identity per key, easier to audit/rotate. Pay the refactoring cost now (zero production users) |
| 6 | Document LWW sync, ship CRDT in v0.2.0 | LWW works for common case; CRDT is right long-term but complex |
| 7 | Per-workspace "trust this app" — git root matching | Matches editor mental model. Fall back to cwd for non-git dirs |
| 8 | Anomaly detector reset: both conversation start AND 30-min idle | Two triggers cover explicit and implicit context boundaries. False positives from stale state are worse than brief re-learning |
| 9 | Pre-launch gate (Phase 15 checklist) BEFORE v0.2.0 work | Tier 3 verification on real hardware is mandatory per STYLE.md |

---

## 6. Current State — What's Green

### Backend (Go)
- Build: clean
- Lint: 0 issues
- Tests: all 64+ packages pass
- CI: 14/14 green
- Daemon boots, all subsystems initialize, schema migrates to v6
- End-to-end pipeline: spawn → pending → approve → run → audit (Tier-3 verified)
- SSE broker reachable, HMAC audit chain intact
- No production TODOs, no drifted specs

### Website (Next.js)
- Build: clean (`npx next build`)
- Typecheck: clean (`npx tsc --noEmit`)
- All 9 pages render, all routes work
- Privacy policy, 404 page, mobile hamburger menu all present
- Brand naming consistent (condura, not synaptic)

---

## 7. Remaining Gaps

### App UI (Wails frontend) — ~80% complete

| Gap | Severity | Detail |
|---|---|---|
| **Overlay window is a visual placeholder** | HIGH | Renders VoiceOrb + text input but input is NOT wired to send messages. This is the primary user interaction surface (the hotkey overlay). Non-functional. |
| **Backup restore missing** | MEDIUM | Can create/list backups, no restore button or flow |
| **i18n translations are stubs** | MEDIUM | All 6 locale JSON files exist but contain identical English text. No actual translations for es/fr/de/ja/zh |
| **Tool calls not rendered in Chat** | MEDIUM | Data types support ToolCall/ToolSpec, but UI doesn't show tool invocations or results |
| **Build not clean** | MEDIUM | `svelte-check` reports 1 error + 11 warnings |

### Missing (v0.2.0 or native Wails)
- MCP UI (backend exists, no frontend)
- Menu bar app / system tray / status icon (native Wails concern)
- Uninstall flow (no UI, no RPCs in client)
- TUI (separate Ink/React project)
- Hardened Layer 3 pf/netsh (currently InProcessGuard skeleton)
- CGEventTap/AT-SPI dirty tracking (perception is data model only)
- Audit log encryption at rest (plaintext SQLite)
- Subscription OAuth UI
- Public Skills Hub + web dashboard (separate Next.js apps)
- Crowdin i18n sync
- Demo video

### Website (minor)
- No OG image, no Twitter cards, no robots.txt, no sitemap.xml, no JSON-LD
- 7/9 pages are "use client" — can't export per-page metadata (SEO gap)
- No `next/image` usage, no dynamic imports
- Color contrast fails WCAG AA on some pages (text-white/40 = 3.66:1)
- Some dead code files and unused image assets
- `config.update` on the web frontend doesn't persist (backend fix applied, but web frontend's config.update may have same issue — not verified)

---

## 8. Next Steps (in priority order)

1. **Pre-launch gate**: Run Phase 15 checklist on clean macOS — fresh install → onboarding → chat → computer use → delegation → safety → voice → backup/restore/uninstall. This is hardware work, not code. Must be done before v0.1.0 ships.

2. **App UI gaps** (before v0.1.0):
   - Wire overlay input to actually send messages (the hotkey overlay must work)
   - Add backup restore flow
   - Fix `svelte-check` errors/warnings
   - (i18n translations and tool call rendering can ship as-is with a "coming soon" note)

3. **v0.2.0 backlog** (after v0.1.0 ships):
   - Hardened Layer 3 pf/netsh
   - CGEventTap/AT-SPI dirty tracking
   - MCP UI
   - Audit log encryption at rest
   - Public Hub + dashboard deploy
   - Crowdin sync
   - Vision CUA opt-in
   - Non-macOS voice
   - Subscription OAuth

---

## 9. Critical Files Reference

| File | Purpose |
|---|---|
| `CLAUDE.md` | Single source of truth — spec, architecture, locked decisions |
| `STYLE.md` | Working style — Tier 3 verification, commit hygiene, anti-patterns |
| `LOGBOOK.md` | Append-only session log — read latest entry for current state |
| `MISSION.md` | Project mission and invariants |
| `docs/roadmap-v0.2.0.md` | v0.2.0 deferred items |
| `docs/phase15-verification.md` | Pre-launch checklist |
| `internal/daemon/subsystems.go` | All subsystem construction and wiring |
| `internal/gatekeeper/defaults.yaml` | Live gatekeeper policy (authoritative, not CLAUDE.md examples) |
| `internal/daemon/methods.go` | Core RPC handlers (llm.chat, apikeys, providers) |
| `internal/daemon/methods_more.go` | Config, telemetry, window, first-run RPC handlers |
| `internal/stream/manager.go` | Streaming LLM lifecycle |
| `internal/ipc/transport.go` | HTTP/WS/SSE transport + auth + ticket exchange |
| `internal/failover/breaker.go` | Circuit breaker per provider |
| `internal/watchdog/watchdog.go` | Inactivity timer → auto-halt |
| `internal/computeruse/verify.go` | Twin-snapshot verification (VerifySnapshots) |
| `internal/delegation/gated_runner.go` | Sub-agent spawn + ActionRequests parsing |
| `internal/daemon/delegation_wiring.go` | ActionRequests gating + audit (Phase 17 B5) |
| `internal/daemon/cu_resolver.go` | Computer-use execution + twin-snapshot verify (Phase 17 B4) |
| `internal/pending/store.go` | Pending actions queue for sub-agent approval |
| `internal/executor/executor.go` | Action execution dispatcher (shell + computer-use) |
| `web/lib/site.ts` | Website constants (SITE, NAV_LINKS, PLATFORMS, TOOL_ROSTER, INVARIANTS) |
| `web/components/shell/GlobalNav.tsx` | Top nav pill with scroll-hide/reveal + mobile hamburger |
| `web/components/shell/SiteDock.tsx` | Bottom dock — 5 quick-access actions |
| `web/components/home/HeroSection.tsx` | Landing hero — "One hotkey. Every AI you own." |
| `web/app/api/auth/magic/route.ts` | Magic link auth with host allowlist |
| `app/web/frontend/src/lib/ipc/client.ts` | IPC client — ~70 typed RPC methods |
| `app/web/frontend/src/App.svelte` | Root component — hash router, onboarding gate, overlay mode |
| `app/web/frontend/src/lib/routes/Chat.svelte` | Main chat UI |
| `app/web/frontend/src/lib/routes/Settings.svelte` | Settings — 14 sections |
| `app/web/frontend/src/lib/components/OnboardingWizard.svelte` | 4-screen onboarding |
| `app/web/frontend/src/lib/components/ConsentModal.svelte` | Gatekeeper Allow/Deny dialog |
| `app/web/frontend/src/lib/components/PendingActions.svelte` | Sub-agent action approval panel |

---

## 10. Working Style (from STYLE.md)

- **Tier 3 verification**: Run the real binary, inspect on-disk state. Don't trust tests alone.
- **Commit hygiene**: One logical change per commit. Message says WHAT + WHY.
- **Lint + test before commit**: `golangci-lint` 0 issues, `go test ./...` all pass.
- **LOGBOOK after every session**: Append entry that next agent can read cold.
- **Never claim done without verifying**: "A green test is not proof the feature works."
- **Surface confusion, don't swallow it**: State assumptions explicitly.
- **Default to action when reversible, ask when irreversible**.

---

## 11. The Single Most Important Thing

From STYLE.md §0:

> **A green test is not proof the feature works.** It is proof the test passed. The only test that proves a feature works in production is one that drives the real production binary, on a real data dir, with a real RPC call, and inspects the on-disk result.

The pre-launch gate (Phase 15 checklist on clean macOS) is the Tier 3 verification that proves v0.1.0 actually works. Everything else — unit tests, integration tests, CI green — is Tier 1+2. Do not ship without Tier 3.
