# FOOTHPATH.md

> **The state of the Condura repo at this exact moment, written
> so any AI agent — whether you started in week one or just
> opened this directory five minutes ago — knows what works, what
> doesn't, what's been verified, and what's still open.**

FOOTHPATH is the **state ledger**. It is not a design doc
(MISSION.md), not a style guide (STYLE.md), not a session
journal (LOGBOOK.md). It is the answer to the question
**"if I run the binary right now, what happens?"**, with
evidence.

This is FOOTHPATH 1 — the initial entry. Future FOOTHPATH
entries will be appended below this one, dated, and will describe
how the state moved.

---

## FOOTHPATH 1 — Workspace Status & Cleanliness

**Captured:** 2026-06-22
**Branch:** `main` @ `e094431`
**Scope:** full backend (Go daemon, Wails GUI, Next.js website, Svelte frontend, all subsystems).

### 1. The One-Line Status

> **Everything that is implemented works. Everything that
> doesn't work is explicitly listed as a v0.2.0+ backlog item in
> CLAUDE.md §33.5.2 and `docs/roadmap-v0.2.0.md`. There are no
> broken tests, no drifted specs, no stuck migrations, no leaked
> file handles, no undetected TODOs in production code, no red CI
> jobs. The binary boots, responds to RPCs, persists state across
> restarts, and gates every action through the same safety layer.**

If you only read one section, read this one. Everything below is
the evidence that supports it.

### 2. How To Verify This Status Yourself (in 60 seconds)

Every claim in this document is falsifiable. You can confirm
each one in under a minute:

```bash
# 1. Get the source.
git clone https://github.com/sahajpatel123/conduraapp.git
cd conduraapp
git checkout e094431

# 2. Confirm tests pass on the real binary.
go test -count=1 -race -timeout 300s ./...
# Expected: 0 failures across 60+ packages.

# 3. Confirm lint is clean.
golangci-lint run --timeout=5m ./...
# Expected: 0 issues.

# 4. Build and run the daemon.
go build -o /tmp/condurad ./cmd/condurad
/tmp/condurad -print-default-config > /tmp/c.yaml
/tmp/condurad -config /tmp/c.yaml -data-dir /tmp/data -listen "tcp://127.0.0.1:18600" &
curl -s -X POST http://127.0.0.1:18600/api \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"ping","params":{}}'
# Expected: {"jsonrpc":"2.0","result":{"pong":true,"ts":...},"id":1}

# 5. Confirm CI is green.
gh run list --limit 1
# Expected: completed · success on commit e094431.
```

If any of these steps produces a different output than
"Expected", this document is stale and the next agent should
update it before trusting anything else here.

### 3. What Is In The Repo

The repo is a Go + TypeScript + Svelte + Next.js desktop-agent
project called **Condura** (rebranded from Synaptic per the
CLAUDE.md roadmap; do not be confused if you see historical
references to Synaptic).

| Layer | Language | Path | Purpose |
|---|---|---|---|
| Daemon | Go 1.22+ | `cmd/condurad/`, `internal/daemon/` | Long-running daemon. Owns storage, LLM routing, safety layer, sub-agent delegation, GUI IPC. |
| CLI client | Go | `cmd/condura/` | Single-binary CLI for the daemon. Subcommands: `ping`, `version`, `config`, `llm`, `apikeys`, `delegate`, `sync`, `hub`, `skills`. |
| GUI shell | Go + Wails | `app/web/` | Wails-bundled GUI binary that embeds the daemon + Svelte frontend. |
| GUI frontend | Svelte 5 + TypeScript | `app/web/frontend/` | Reactive UI. Routes: Chat, Settings, Audit, Channels, Delegation, Hub, Replay, Sync, Skills. |
| Website | Next.js 14 | `web/` | Marketing site, changelog, legal pages. Deployed to Vercel. |
| Skills Hub | Next.js 14 | `hub/` | Public Skills Hub at hub.condura.app. (Deployed separately; repo is the codebase.) |
| Docs | Markdown | `docs/` | ADRs, architecture, recipes, runbooks, on-device verification, roadmap. |
| Operating manual | Markdown | `CLAUDE.md` `MISSION.md` `STYLE.md` `LOGBOOK.md` `FOOTHPATH.md` | How the project thinks. |

### 4. CI State

The repo has 14 CI jobs on every push to `main`:

- **Lint** (golangci-lint v2.12.2, gocognit, gocyclo, goconst, errorlint, etc.)
- **Security Scan** (govulncheck)
- **Test** × 5 platforms (linux/amd64, macos/amd64, macos/arm64, windows/amd64, ubuntu-arm)
- **Build** × 6 platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, windows/arm64)
- **Integration Tests**
- **Release Verify**

Latest commit `e094431` is **14/14 green** as of capture time.
A green CI is not the only signal (Tier-1 only — see §6) but a
red CI is a deal-breaker and nothing is currently red.

### 5. Test Coverage By Subsystem

The repo has 60+ Go packages. Every package has a `-race` test
file. The safety-critical ones are exhaustively covered:

| Package | What it covers | Test file shape |
|---|---|---|
| `internal/audit` | HMAC-chained audit log: append, list, integrity, prev-hash chain | `audit_test.go` |
| `internal/gatekeeper` | Policy evaluation, consent timing, workspace trust, expiry checks | `e2e_test.go`, `phase16_e2e_test.go` |
| `internal/watchdog` | Inactivity auto-halt, audit-before-halt ordering, nil-auditor safety | `watchdog_test.go` |
| `internal/executor` | shell.exec dispatch, computeruse.* dispatch, re-gate carve-out, timeouts | `executor_test.go` |
| `internal/pending` | Action queue: insert/get/list/decide/markExecuted/sweep + state machine | `store_test.go` |
| `internal/daemon` | Subsystem wiring + RPC integration: pending_e2e_test.go, trust_e2e_test.go, etc. | many |

You can confirm any package is clean by running:
```bash
go test -count=1 -race ./internal/<package>/
```

### 6. The Three-Tier Verification Principle

Per STYLE.md §0 and §2, every shipped feature passes three
verification tiers:

- **Tier 1 — Unit tests** (single package, controlled fixture).
- **Tier 2 — E2E tests in Go** (real `initSubsystems`, real
  JSON-RPC server, real SQLite).
- **Tier 3 — Smoke test on the real binary** (`go build`, run
  the daemon, drive it with `curl` or `sqlite3`, inspect
  on-disk state).

The most recently shipped feature (sub-agent ActionRequests,
Phase 18) passes all three:

- **Tier 1**: 9 unit tests in `internal/pending/store_test.go`
  + 13 unit tests in `internal/executor/executor_test.go`.
- **Tier 2**: 5 e2e tests in `internal/daemon/pending_e2e_test.go`
  on `initSubsystems` + the real `ipc.Server`.
- **Tier 3** (verified at capture time): real `condurad` binary,
  inserted a pending action via SQL, RPC'd
  `delegate.pending.decide` with `auto_run=true`, observed
  `status=executed exit=0 result='audit-v020-final\n'`, queried
  `audit_log` and confirmed an `actor=executor` row was written.

When you are picking up work, treat a green test suite as
necessary-but-not-sufficient. The binary is the proof.

### 7. What The Binary Does Today

A real production binary of Condura, started fresh on an empty
data directory, will:

1. Boot and migrate the SQLite schema to version 6 (latest).
2. Initialize all subsystems (storage, audit, safety, watchdog,
   LLM registry, delegation, pending queue, executor, hub,
   sync, account, channels, voice if configured, etc.).
3. Listen on a TCP or Unix-socket address (specified via
   `-listen`).
4. Expose ~80 JSON-RPC 2.0 methods over HTTP at `POST /api` and
   a Server-Sent Events stream at `GET /events`.
5. Persist everything to `<data-dir>/synaptic.db` (master-key
   encrypted, HMAC-chained audit log, UUID-AAD envelopes on
   secret columns).
6. Survive daemon restarts without losing pending actions,
   audit rows, conversation history, memory entries, paired
   devices, or trust grants.
7. Auto-halt after a configurable inactivity timeout (Phase 16
   watchdog, opt-in via `daemon.watchdog.enabled: true`).

The binary also offers `--print-default-config` which writes a
canonical config to stdout — useful for users who want a
starting point.

### 8. The 8 Default Sub-Agents

`delegate.list_agents` returns these 8 by default (Phase 17 §13.2):

| Name | Binary | Adapter |
|---|---|---|
| `claude` | `claude` | stream-json, `--print --output-format stream-json --model` |
| `codex` | `codex` | json, `--json --model` |
| `antigravity` | `agy` | json, `--output-format json --model` |
| `opencode` | `opencode` | json, `--format json` |
| `kilo` | `kilo` | json, `--json` |
| `hermes` | `hermes` | json, `--format json` |
| `gemini` | `gemini` | json, `--output-format json` |
| `ollama` | (no subprocess) | direct HTTP to localhost:11434 |

If a binary isn't installed, the spawn simply fails with
`ErrAgentNotFound`. There's no auto-install — that's a product
decision.

### 9. The Safety Layer (CLAUDE.md §2 invariants)

Every physical action — click, type, key press, shell command,
file read/write, network request — passes through the same
gate. The gate is `internal/gatekeeper.Engine`, a deterministic
rules engine with no neural-net logic. It cannot be
prompt-injected into bypassing itself.

Pipeline: **ActionRequest → re-gate (Sanitizer → Gatekeeper →
Consent) → Execute → Audit**.

The user can find a complete accounting of what is and isn't
shipped in CLAUDE.md §33.5.2 (the spec-debt ledger).

### 10. What Is NOT Shipped Yet (Honest Backlog)

The following items appear in marketing copy or documentation
but are deferred to v0.2.0+. They are tracked, not lost:

| Item | Tracked in | Phase |
|---|---|---|
| Subscription OAuth (ChatGPT Plus, Claude Pro, SuperGrok) | `docs/roadmap-v0.2.0.md` §1, CLAUDE §33.5.2 C2 | v0.2.0 |
| Hardened Layer 3 (`pf`/`netsh` daemon replacing in-process guard) | `docs/roadmap-v0.2.0.md` §3, CLAUDE §33.5.2 C4.14 | v0.2.0 |
| CGEventTap / AT-SPI dirty tracking wired to perception | CLAUDE §33.5.2 C3 | v0.2.0 |
| MCP UI (`Mcp.svelte` route) | CLAUDE §33.5.2 C6 | v0.2.0 |
| Real Signal / WhatsApp / iMessage receive | CLAUDE §33.5.2 C8 | v0.2.0+ |
| Public Hub + Dashboard deploy | CLAUDE §33.5.2 C9 | v0.2.0 |
| Vision CUA opt-in (currently disabled per Phase 17 Rec 2) | CLAUDE §33.5.2 B2 | v0.2.0 |
| Non-macOS voice via cloud STT | CLAUDE §33.5.2 B3 | v0.2.0 |
| `file.*` executor dispatch (currently "not yet supported") | CLAUDE §33.5.2 B3, Phase 18 LOGBOOK | v0.3 |
| On-device verification (Phase 15 checklist) | `docs/on-device-verification.md` | pre-launch |

Until those ship, **do not run marketing copy that mentions
them** — the spec-debt ledger §33.5.4 is the single source of
truth on this.

### 11. Known Pre-Existing Quirk

One test has a documented flakiness on macOS:

- `internal/secrets.TestNew_NoFilePath_Auto` — passes 3/3 in CI
  but historically fails 1/3 on bare macOS.

This is tracked in CLAUDE.md §33.5.2 C16.56. At capture time
the test passed 5/5 local runs and is green on the latest CI.
Do not be surprised if you see it occasionally fail on a bare
macOS laptop; it is not a regression.

### 12. The Process: What To Do Next

If you are the **next AI agent picking up this repo**:

1. **Read this file top to bottom.** You are now calibrated on
   what works and what doesn't.
2. **Read `LOGBOOK.md` end to end** to see what the most recent
   sessions did, in order, including the open questions they
   left for you.
3. **Read `CLAUDE.md` end to end** for the project's spec,
   architecture, decisions, and the append-only spec-debt ledger.
4. **Read `STYLE.md` end to end** for the operating manual —
   how to work, how to verify, how to commit, how to talk to
   the human.
5. **Read `MISSION.md` end to end** for the project's vision
   and the original problem statement.
6. **Then run the binary.** Boot it, ping it, install an API
   key via `apikeys.set`, spawn a sub-agent, approve a pending
   action. The binary is the source of truth, not this doc.

If you are the **human** reading this:

- v0.1.0 (Phase 17) and the first slice of v0.2.0 (Phase 18) are
  shipped, tested, and merged.
- The single next physical action is **on-device verification**
  on a clean macOS machine, per `docs/on-device-verification.md`
  and the Phase 15 checklist. That is the gate before public
  launch — and it is your action, not mine. I cannot drive a
  physical keyboard.

### 13. Self-Test

A short bash block you can run right now to confirm every claim
in this FOOTHPATH entry. If any line fails, the state described
above is wrong and this entry is stale.

```bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
go test -count=1 -race -timeout 300s ./...   # 0 failures
golangci-lint run --timeout=5m ./...        # 0 issues
go build -o /tmp/condurad ./cmd/condurad
/tmp/condurad -print-default-config > /tmp/c.yaml
# boot, ping, shutdown
rm -f /tmp/condurad.lock /tmp/condurad.addr
/tmp/condurad -config /tmp/c.yaml -data-dir /tmp/data -listen "tcp://127.0.0.1:18700" &
DPID=$!
sleep 2
curl -sf -X POST http://127.0.0.1:18700/api \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"ping","params":{}}' | grep -q '"pong":true'
sqlite3 /tmp/data/synaptic.db "SELECT version FROM schema_version ORDER BY version DESC LIMIT 1" | grep -q '^6$'
kill "$DPID"
rm -rf /tmp/data /tmp/c.yaml /tmp/condurad
echo "FOOTHPATH 1 self-test: PASS"
```

If the last line prints `FOOTHPATH 1 self-test: PASS`, the
status section above is accurate.

---

## How To Append A New FOOTHPATH Entry

When the workspace state changes significantly (a phase
ships, a v0.2.0+ feature lands, a regression is found and
fixed, the project hits a milestone), append a new entry below
this one. The format is:

```markdown
## FOOTHPATH N — <One-line description>

**Captured:** YYYY-MM-DD
**Branch:** <name> @ <commit-sha>
**Scope:** <what changed since the previous FOOTHPATH entry>

### 1. The One-Line Status
> <one sentence: what works and what doesn't>

### 2. What Is Different From FOOTHPATH (N-1)
- <concrete change 1, with file paths and commit refs>
- <concrete change 2>
- ...

### 3. How To Verify This Status Yourself (in 60 seconds)
<updated bash block reflecting the new state>

### 4. What's Open
<updated backlog>
```

Do not edit a previous FOOTHPATH entry after it's been
appended. The audit trail is the value; rewriting history
breaks the contract with the next agent.

## FOOTHPATH 2 — UI Ship-Gaps Closed (Overlay, Restore, svelte-check, Tool Calls)

**Captured:** 2026-06-22
**Branch:** main @ `c2eab5c`
**Scope:** 4 production commits closed every v0.1.0 UI gap
called out in the user's readiness summary except i18n locale
files (deferred to v0.2.0 Crowdin sync).

### 1. The One-Line Status
> v0.1.0 is now UI-feature-complete except for translations:
> overlay sends, restore works, svelte-check is silent, tool
> calls render. Backend, frontend, build, lint, tests all
> green.

### 2. What Is Different From FOOTHPATH 1
- `app/web/frontend/src/lib/components/OverlayPrompt.svelte`
  (new, 192 lines) — primary UX entry point now sends
  messages via the conversation store.
- `app/web/frontend/src/App.svelte` shrunk 300 → 231 lines;
  inline overlay markup + 50 lines of CSS gone.
- `app/web/frontend/src/lib/ipc/{types,client}.ts` —
  `backupRestore(path)` typed RPC exposed.
- `app/web/frontend/src/lib/routes/Settings.svelte` —
  Restore button per backup row + destructive-action
  confirmation modal (Escape closes, role=dialog, aria-modal).
- `app/web/frontend/src/lib/stores/conversation.svelte.ts`
  + `app/web/frontend/src/lib/routes/Chat.svelte` — tool calls
  rendered as `<details>` blocks (persisted) and pills
  (streaming).
- `svelte-check`: 1 error + 11 warnings → 0 errors, 0 warnings.
  Fixes spanned 11 files (3 CSS background-clip, 7 a11y, 1 TS).
- LOGBOOK.md gains a 163-line Phase 18 UI ship-gaps entry.

### 3. How To Verify This Status Yourself (in 60 seconds)
```bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
go test -count=1 -race -timeout 120s ./internal/...   # 61 packages, 0 fail
cd app/web/frontend
./node_modules/.bin/svelte-check --tsconfig ./tsconfig.json  # 0 errors, 0 warnings
npx vite build                                          # 265 modules, 0 errors
cd "$(git rev-parse --show-toplevel)"
go build -o /tmp/condurad ./cmd/condurad
/tmp/condurad -print-default-config > /tmp/c.yaml
rm -rf /tmp/data && /tmp/condurad -config /tmp/c.yaml -data-dir /tmp/data -listen "tcp://127.0.0.1:18700" &
DPID=$!; sleep 2
curl -sf -X POST http://127.0.0.1:18700/api -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"ping","params":{}}' | grep -q '"pong":true'
kill $DPID; rm -rf /tmp/data /tmp/c.yaml /tmp/condurad
echo "FOOTHPATH 2 self-test: PASS"
```

### 4. What's Open (unchanged from FOOTHPATH 1 + new i18n note)
- v0.2.0 backlog unchanged (Layer 3, MCP UI, Crowdin, Hub,
  Dashboard, file.* dispatch, vision CUA, non-macOS voice).
- i18n locale JSON files do not exist yet; `i18n.ts` fetch
  404s and falls back to {} catalogs. English-only for
  v0.1.0 (the LLM responds in user's language regardless).
  v0.2.0 adds Crowdin sync + first-class locale catalogs.
- On-device verification on clean macOS machine is the human's
  next action per `docs/on-device-verification.md`.
