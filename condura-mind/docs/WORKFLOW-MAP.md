# Condura — End-to-end workflow map

> Source of truth for how the product boots, talks to itself, and ships.
> Verified against `main` as of the workflow audit. Canonical main DB: **`condura.db`**.

---

## 1. Surfaces (who runs what)

| Binary / surface | Path | Role |
|---|---|---|
| `condurad` | `condura-app/cmd/condurad` | Standalone daemon CLI |
| `condura-gui` | `condura-app/cmd/condura-gui` | Wails shell + **embedded** `daemon.Run()` |
| `condura` | `condura-app/cmd/condura` | User CLI (RPC client) |
| `condura-tui` | `condura-app/cmd/condura-tui` | Terminal UI (RPC client) |
| Svelte UI | `condura-gui/frontend` | In-app UI (talks HTTP JSON-RPC + SSE) |
| Marketing | `condura-ui` | Next.js site (download honesty) |
| Hub / SDK | `condura-hub`, `condura-sdk` | Reserved stubs (v0.2.0) |

---

## 2. Boot sequence

### 2a. Standalone daemon (`condurad`)

```
parseFlags → buildConfig/loader
  → signal.NotifyContext(SIGINT/SIGTERM)
  → daemon.Run(ctx, Options)
```

### 2b. Desktop GUI (`condura-gui`)

```
resolveConfig
  → safego.Go { daemon.Run(...); tray; conductor hotkey; kill-switch hotkey }
  → wails.Run(App + embedded frontend assets)
```

### 2c. Inside `daemon.Run` (single source of truth)

```
1. Validate config
2. mkdir DataDir (~/.condura by default)
3. migrateLegacyDataDir: ~/.synaptic → ~/.condura + synaptic.db → condura.db
4. CompletePendingUpdate (Windows staged swap)
5. lockfile.TryAcquire(<data-dir>/condurad.lock)  // single instance
6. Logger (stderr + optional size-rotated file)
7. MarkDaemonStart
8. initSubsystems (storage=condura.db, secrets, audit, gatekeeper, llm, …)
9. registerMethods (JSON-RPC surface)
10. Listen TCP 127.0.0.1:0 (+ Unix <data-dir>/condurad.sock on mac/linux)
11. writeAddrFile (condurad.addr sidecar)
12. startBackgroundServices (backup, updater, watchdog, audit pruner, …)
13. block until ctx cancel → shutdown
```

**Canonical paths**

| Artifact | Path |
|---|---|
| Data dir | `~/.condura/` |
| Main DB | `~/.condura/condura.db` |
| Lock | `~/.condura/condurad.lock` |
| Addr | `~/.condura/condurad.addr` |
| Unix socket | `~/.condura/condurad.sock` |
| User backups | `~/Documents/condura-backups/` (or `CONDURA_BACKUP_DIR`) |
| Wake word | `hey condura` |

Legacy: `~/.synaptic/` and `synaptic.db` are migrated/accepted, not preferred.

---

## 3. Client ↔ daemon protocol

```
GUI/CLI
  │  HTTP POST /api   JSON-RPC 2.0  (+ Bearer if auth_token set)
  │  EventSource /events  SSE (stream.*, halt, audit, …)
  ▼
ipc.ServerTransport
  → method handlers in internal/daemon/methods*.go
  → stream.Manager → llm.Registry → provider
  → gatekeeper / audit / halt as required
```

**GUI init** (`stores/init.ts`):

1. Wails `DaemonStatus()` → baseURL  
2. Fallback `http://127.0.0.1:7666`  
3. Optional `config.get` for auth token  
4. Start IPC + poll spend/halt/update + load settings/conversations  

---

## 4. First-run / onboarding flow

```
App mount
  → firstRunStatus + onboardingIsComplete
  → if incomplete: OnboardingWizard / FloatingInterview
       EULA → permissions (TCC) → hotkey → ready (Ollama/API probe)
  → onboarding.finish → first_run=false
  → main shell (chat)
```

Settings can re-open onboarding via custom event `condura:show-onboarding`.

---

## 5. Chat / stream flow

```
User send
  → conversation.send(provider, model, text)
  → conversations.append (user message)
  → llm.stream RPC → { request_id }
  → SSE stream.delta / stream.finished / stream.error
  → store filters by conversation_id + request_id
  → idle watchdog 45s + disconnect clears isStreaming
  → on done: conversations.append (assistant)
```

Safety on physical actions (computer use / shell / restore / uninstall):  
**Gatekeeper** (deterministic) → consent UI → audit HMAC chain.  
Models never bypass the gatekeeper.

---

## 6. Safety / kill path

| Layer | Mechanism |
|---|---|
| 1 | Hard kill-switch hotkey → `halt.Flag` |
| 2 | Watchdog (idle / verification timeout) |
| 3 | Network guard (**in-process** in v0.1.x; hard OS process = v0.2.0) |
| Always | User can halt; resume needs human confirm ticket |

Consent modals + kill overlay use focus traps (a11y).

---

## 7. Backup / restore / uninstall

```
backup.create  → packs condura.db (or legacy synaptic.db) as condura.db
                 + memory/skills/secrets/config → encrypted zip
backup.restore → Gatekeeper → atomic swap → integrity check
uninstall.*    → Gatekeeper → optional pre-backup → remove manifest
                 (both condura.db and legacy synaptic.db listed)
```

---

## 8. Release / ship flow

See `docs/SHIP-CHECKLIST.md` and `docs/release-runbook.md`.

```
make release-dry-run-local
  → secrets (UPDATE_SIGNING_KEY + 6 Apple)
  → tag v0.0.0-test dry-run
  → Phase 15 human Mac
  → public v0.1.x tag (fail-closed if secrets missing)
```

---

## 9. What is intentionally NOT in v0.1.x flow

Hybrid multi-provider router, hard Layer-3 OS guard, subscription OAuth,  
Wave/DAG parallel agents, public Skills Hub, public SDK, full channel bots.

---

## 10. Custom events (frontend)

| Event | Direction |
|---|---|
| `condura:show-onboarding` | Settings → App |
| `condura:open-palette` | TitleBar → App |
| Theme key | `localStorage['condura:mode']` |

---

## Verification commands

```bash
make check-lockfiles
go test -count=1 -short ./condura-app/internal/{config,backup,uninstall,safego,stream,session,gatekeeper,onboarding}/
cd condura-gui/frontend && npm test
# CI green on main before any public tag
```
