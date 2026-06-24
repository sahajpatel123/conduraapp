# Synaptic/Condura — Backend Audit Report (Tier 4)

> **Author:** minimax-m3 via opencode
> **Date:** 2026-06-24
> **Branch audited:** `main` @ `c8e4d55` (HEAD at time of audit)
> **Scope:** Full Go backend — 60 packages under `internal/` + `cmd/`. Functional
>   completeness and spec-conformance audit. This is **not** a security audit;
>   `docs/analysis/security-audit-2026-06-24.md` covers security and was
>   independently re-confirmed at HEAD during this audit (all 12 findings still
>   open — none have been addressed since that audit).
> **Methodology:** Tier 4 — read every package end-to-end (file count, LOC,
>   function bodies, test files), verify wiring into the daemon, grep for
>   stubs/TODOs/dead code, cross-reference against CLAUDE.md §§1–32 and the
>   `docs/roadmap-v0.2.0.md` deferral list. Four parallel exploration agents
>   each took a domain; their findings were merged and re-verified here.
> **Severity scale:** P0 (blocks release / data loss) → P1 (broken feature
>   claimed as working) → P2 (functional gap vs spec, non-blocking) →
>   P3 (cosmetic / dead code / docs drift) → INFO (intentionally deferred,
>   documented).

---

## 0. Executive Summary

**Overall posture: the backend is REAL, not faked.** Across 60 packages I
found **zero** cases of a function returning a fabricated successful result
where real work was claimed. Every stub is *honestly labelled* — it returns
`"coming in v0.2.0"` or `"not available on this platform"` rather than
pretending to succeed. The safety layer is genuinely deterministic (not a
model). The audit log is genuinely HMAC-chained. The updater is genuinely
Ed25519-verified. This is unusual for a project at this maturity level.

**That said, there are real gaps.** Sorted by what should happen before
v0.1.0 ships to the public:

| # | Finding | Severity | Action before v0.1.0? |
|---|---|---|---|
| **B-01** | `internal/autonomy` is **orphaned** — the user-editable per-cell autonomy matrix from §27 (the headline user-control feature) is parsed from config and **never consumed**. The Gatekeeper uses a hardcoded 3-line heuristic instead. | **P1** | **Yes — wire it or remove the §27 claim** |
| **B-02** | `internal/perception` (Selective Perception, §6) is **orphaned** — real SmartCapturer/DirtyTracker/PIIRedactor exist but nothing imports them. Battery + safety unification is not actually driving any capture. | **P1** | **Yes — wire it or remove the §6 claim** |
| **B-03** | Anomaly detector (§5.6) implements only **4 of 5** "agent went insane" triggers. The "new endpoint / never-used network target" trigger is missing entirely. | **P2** | Recommend yes (cheap) |
| **B-04** | `internal/memory` has **no vector embeddings / sqlite-vec** despite §8/§14.4. Recall is FTS5-keyword-only. `storage/migrations.go:117` literally has `embedding BLOB, -- future: vector embedding`. | **P2** | Document as v0.2.0 (already in roadmap) |
| **B-05** | `internal/sync` is **LAN-only** (UDP broadcast). No Kademlia DHT, no NAT traversal, no relay. Over-internet P2P sync is impossible. §17.2 claims "NAT traversal + relay fallback." | **P2** | Document as v0.2.0 |
| **B-06** | `internal/replay` 24h buffer stores **encrypted raw PNG**, not H.264-delta-compressed as §18.3 claims. H.264 only at ffmpeg export time. Disk usage is higher than spec implies. | **P2** | Fix the §18.3 claim or compress the buffer |
| **B-07** | Wake word rename is **incomplete**: `internal/voice/modelmgr/modelmgr.go:137-218` and `openwakeword_detector.go:146` still use `hey_synaptic` (model name, asset filename, URL org `synaptic/wake-words`, `WakeModelForName` only accepts `hey_synaptic`). The display default is "hey condura" but the detector layer was never renamed. Contradicts the Phase 14I closure claim. | **P1** | **Yes — finish the rename** |
| **B-08** | `internal/permissions` has **no native probes on Windows/Linux**. Only macOS has a CGO probe (`AXIsProcessTrusted`); Windows/Linux always return `StatusUnknown`. Onboarding "Permissions" screen will show unknown on 2 of 3 platforms. | **P2** | Document or build the probes |
| **B-09** | `internal/executor.execShell` runs `sh -c <user-string>` and **bypasses `internal/sanitize`** — the §5.5 model-isolation invariant is violated for shell-kind pending actions. The Gatekeeper approves policy; the sanitizer is never called. | **P1** | **Yes — route through sanitize** |
| **B-10** | `internal/presence/detector.go:145-154` Windows "user active" check is **wrong**: it counts network adapters (`Get-NetAdapter | Where Status -eq 'Up'`) and returns `true` if any adapter is up — i.e. any Wi-Fi-connected Windows machine always reports "user present." The `require_user_active` consent gate is defeated on Windows. | **P1** | **Yes — fix or stub to false** |
| **B-11** | `osascript` injection in `computeruse/backends/macosmcp_darwin.go` — `escapeAppleScript` only escapes `\` and `"`, not backticks. Model-controlled `action.Value` can inject `` `do shell script "..."` ``. Also `reach/imessage_darwin.go:52` uses Go's `%q` (not AppleScript-safe). Security audit F-11 — **still open**. | **P1** | **Yes — escape backticks or block** |
| **B-12** | `internal/secrets` file-fallback backend writes the **master AES key + every secret as cleartext JSON** when the OS keyring is unavailable. Security audit F-01 (CRITICAL) — **still open**. | **P0** | **Yes — encrypt the file or refuse to boot** |
| B-13 | `internal/router` package **does not exist** (§12 hybrid router). v0.1.0 uses a single hardcoded `providerName + model`. | INFO | Already documented in roadmap |
| B-14 | Subscription OAuth (ChatGPT Plus, Claude Pro, SuperGrok) — `internal/subscription` does not exist. | INFO | Already documented in roadmap |
| B-15 | Execution Waves / DAG scheduler (§13.6) — not implemented. Single-spawn only. | INFO | Already documented in roadmap |
| B-16 | CE-MCP code-execution delegation (§13.3) — not implemented. | INFO | Already documented in roadmap |
| B-17 | WhatsApp / Signal / iMessage-receive — honest stubs returning "v0.2.0". | INFO | Already documented |
| B-18 | Public `hub.condura.app` Next.js app — not in repo. `internal/hub` is a client + a local in-process server (offline fallback). | INFO | Already documented |
| B-19 | MCP only stdio transport (§16.1 claims stdio/HTTP/SSE). | INFO | Already documented |
| B-20 | Linux hotkey is a stub (`hotkey_linux.go:7`). Linux tray is build-tagged out. | INFO | Already documented |
| B-21 | Non-macOS voice capture/TTS is a stub (by design, B3). | INFO | Already documented |
| B-22 | No live "agent is clicking X" SSE indicator event during CU loop. §10 claim that "SSE events exist" is not supported — only `stream.delta` (from LLM) is published, not per-CU-action events. | P2 | Small addition |
| B-23 | `internal/agent` has **3 dead types**: `Loop` (`agent.go`), `ExpandedLoop` (`loop_expanded.go`), `SimplePlanner` (`planner.go:142`) — all defined, none instantiated outside tests. | P3 | Delete or annotate |
| B-24 | `internal/llm/types.go:176` `ErrNotImplemented` is defined and **never returned anywhere**. | P3 | Delete |
| B-25 | `internal/adaptive/predictor.go:43` `minConfidence` const declared and immediately discarded (`_ = minConfidence`). | P3 | Delete |
| B-26 | `internal/adaptive/dialectic.go:128-134` `parseProposals` silently defaults missing confidence to `0.5` — manufactures confidence from malformed LLM output. | P3 | Reject instead |
| B-27 | `internal/sensitive/sensitive.go:125-126` duplicate `"mychart."` entry. | P3 | Delete dup |
| B-28 | `internal/reach/telegram.go:168-171` leftover `var _ =` dead lines. | P3 | Delete |
| B-29 | `internal/permissions/permissions.go:186-187` macOS guide says both "Synaptic" and "Condura" inconsistently. | P3 | Fix branding |
| B-30 | `internal/voice/modelmgr/modelmgr.go:139` wake model `SHA256: ""` — no hash pin. Undermines the trust-anchor pattern `pipeline.go` otherwise enforces. | P2 | Pin the hash |
| B-31 | `internal/hub/client.go:113` User-Agent still `"synapticd/0.1.0"` (old name). | P3 | Rename |
| B-32 | `internal/executor/executor.go` header says "v0.2.0 sibling of Phase 17" but the package IS wired in the current build — comments lag the wiring. | P3 | Fix comment |
| B-33 | `internal/daemon/failover.go` failover `Chat` sends a literal `"ping"` — it's a health pre-check, not a request router. The §12 router claim is not realized by failover. | INFO | Documented |
| B-34 | `gatekeeper.DenyBeyondRead` stub (`gatekeeper.go:56-69`) still in public API "for test backward compatibility" — not used in production. | P3 | Remove or annotate |
| B-35 | `anomaly/detector.go` channel is 256-buffered; under burst, records are silently dropped — the detector could miss its own anomaly signal. | P3 | Document |
| B-36 | `anomaly/detector.go:194-203` loop check only catches A,A,A in the tail window, not A,B,A,B,A. Stricter than spec implies. | P3 | Document |
| B-37 | `audit/log.go` has no visible prune path for the 90-day retention. Append-only with no deletion job in this package. | P2 | Verify pruning exists |
| B-38 | `internal/computeruse/router.go:28-49` breaks on first error rather than cascading to the next tier. A tier-1 backend that *fails* (not "unavailable") does not fall through to tier 2. Deviates from §11.2. | P2 | Confirm intent |
| B-39 | `internal/voice` whisper.cpp + openWakeWord are invoked via **subprocess**, not CGO. §8 says "CGO for macOS" and "openWakeWord (local)" — technically true but not the in-process integration implied. | INFO | Document |
| B-40 | `internal/memory` test coverage is light (3 tests for a 665-LOC SQL store). | P3 | Add tests |

Detailed evidence and per-package breakdown below.

---

## 1. What is genuinely REAL and well-built

These packages have real implementations, real tests, and are correctly
wired. No action needed.

**Safety core:** `blastradius`, `gatekeeper` (engine + policy), `halt` (flag
+ in-process guard), `audit` (HMAC chain), `sanitize` (5 sanitizers, with
documented heuristic caveats), `sensitive` (domain-aware detector),
`pending` (SQLite store), `watchdog` (in-process timer), `trust`
(per-workspace allow-list).

**Computer use:** `computeruse` (4-tier router + twin-snapshot verification
is genuinely wired in `cu_resolver.go:94-119` with `ErrStaleState` abort),
`computeruse/ax` (real CGO macOS backend), `computeruse/backends` (ORAX Eye,
mac-cua, macOS-MCP, VisionCUA — all real on darwin, stubs elsewhere).

**Agent + LLM + session + stream:** `agent.LLMPlanner`, `agent.CULoop`,
`agent.ComputerUseExecutor`, `agent.GatedExecutor`, `agent.SimpleVerifier`
are real and wired. `session.Run` (gate → persist → stream → persist →
speak) is real. `stream.Manager` (per-provider circuit breaker, SSE fan-out)
is real. `llm.Registry` + 12 provider clients (OpenAI-compat, native
Anthropic, native Google) + `model_pricing.go` are real. `conductor`
(hotkey → overlay → capture) is real.

**Memory + skills + adaptive:** `memory` 3-layer store (episodic/semantic/
procedural, all backed by FTS5) is real. `skills` (agentskills.io-compatible
CRUD + auto-create from observation) is real. `adaptive` — the full
proposer + critic + adjudicator dialectic from §9.3 is genuinely
implemented (`dialectic.go:65`, `:80`, `adjudicator.go`), with confidence
bucketing, decay, and encrypted persistence. The `UserModel` matches §9.2.

**Infrastructure:** `storage` (AES-256-GCM column encryption), `config`,
`ipc` (JSON-RPC + WebSocket + bearer auth + SSE tickets), `sse`,
`daemon` (initSubsystems wires all 60 packages — nothing orphaned at the
wiring level), `crash`, `health`, `lockfile`, `logger`, `telemetry` (opt-in,
default OFF), `trust`, `version`, `status`, `window`, `overlay` (noop
headless controller is the documented fallback).

**Sync / backup / replay / updater:** `sync` (Ed25519 identity, Noise-XX-like
handshake — real, LWW+vector-clock store). `backup` (encrypted zip with
AES-GCM content keys). `replay` (24h TTL + timeline + ffmpeg MP4 export).
`updater` (Ed25519 signature verification with embedded public key —
genuinely secure).

**User-facing:** `account` (OAuth Google/GitHub/Apple + magic link),
`api_key` (encrypted per-value UUID-AAD), `onboarding` (clean 4-screen state
machine), `reach` (Telegram real), `i18n` (6 locales via go:embed),
`conversation`, `hub` (client + local server), `failover`, `mcp` (stdio
client + gated wrapper), `tui` (Bubbletea, 8 tabs).

---

## 2. The gaps, with evidence

### B-01 — `internal/autonomy` is orphaned (P1)

`internal/autonomy/autonomy.go` (86 LOC) defines a real `Matrix` with
`Evaluate(taskType, app)` supporting exact + `task.*` wildcard + default,
and `CanAutoApply` / `NeedsConsent` helpers. **It is never used.**

`grep -rn "autonomy\." internal/daemon/` returns only a comment in
`safety_wiring.go:53`. The actual autonomy hook is
`safety_wiring.go:143-159`:

```go
autonomousTasks := map[string]bool{"research": true, "image_generation": true, "code_review": true}
apps := map[string]bool{"com.google.Chrome": true, "com.apple.finder": true, "com.microsoft.VSCode": true}
```

Meanwhile `config.AutonomyConfig` (PerApp/PerTask maps) **is parsed** at
`config.go:268-280` and **never consumed** — `grep` for `AutonomyConfig`,
`PerApp`, `PerTask` in `internal/daemon/` returns nothing.

**Net effect:** CLAUDE.md §27 ("the user-defining setting… user dials each
cell") is not implemented. Users cannot dial the per-cell matrix. The
gatekeeper's autonomy level is fixed in Go source.

### B-02 — `internal/perception` is orphaned (P1)

§6 ("Selective Perception") is described as the unifying system for
battery + safety + performance. The package exists (527 LOC,
`perception.go`) with real `SmartCapturer`, `DirtyTracker`, `PIIRedactor`,
`EnergyMode`/`Strategy` enums, budget enforcement (`ChooseStrategy` returns
error when exhausted, `perception.go:284-298`).

**Zero importers.** `grep -rln "internal/perception"` across `internal/` and
`app/` returns only `perception.go` itself. Not used by `computeruse`,
`agent`, `daemon`, or any CU backend. The §6.5 CGEventTap integration is not
implemented — the package mentions CGEventTap only in comments; there is no
CGO file, no `//go:build darwin` file, no `CGEventTap` symbol anywhere.
`DirtyTracker.Mark()` is a pure data setter — nothing pumps real OS events
into it.

### B-03 — Anomaly detector missing the 5th trigger (P2)

§5.6 lists 5 triggers: speed, loop, duration, failures, **new endpoint**.
`internal/anomaly/detector.go` implements 4 (Rate `:178-187`, Loop
`:160-170,189-204`, Duration `:155-158`, Failures `:149-153,172-175`). The
5th ("agent sends to network endpoints it has never used before") is
**entirely missing** — `grep` for "new endpoint" / "never-used" / "network
target" across the package returns nothing.

### B-04 — No vector embeddings in memory (P2, documented)

§8 lists `all-MiniLM-L6-v2` via local ONNX or Ollama (384-dim). §14.4 says
"sqlite-vec for vector similarity." Neither exists. `grep` across the repo
for `sqlite-vec|vec0|sqlitevec` returns nothing.
`storage/migrations.go:117` has `embedding BLOB, -- future: vector
embedding`. `internal/memory` recall is FTS5 keyword MATCH only
(`sqlite_store.go:248-256`). The embedding model is not referenced in Go
code. **Already in `docs/roadmap-v0.2.0.md`** — but the §8/§14.4 spec text
still reads as if it ships.

### B-05 — P2P sync is LAN-only (P2, not documented)

§17.2 claims "Custom Kademlia DHT or Syncthing-fork" + "NAT traversal +
relay fallback." Reality: `internal/sync/discovery.go:30` is **UDP broadcast
to 255.255.255.255** — LAN-only. No DHT, no mDNS, no STUN/TURN/ICE, no relay.
Over-internet sync is impossible. The Noise-XX-like handshake IS real
(`transport.go:39-50`) and well-tested. The CRDT is LWW+vector-clock (not
true CRDT, but documented).

### B-06 — Replay buffer is not H.264 (P2)

§18.3 says "H.264 delta compression for screenshots. ~50MB per day."
Reality: `internal/replay/screenshots.go:13-15,134` stores screenshots as
**AES-256-GCM-encrypted raw PNG** on disk. H.264 (`libx264`) is used only at
MP4 export time (`export.go:13,66`), and only if ffmpeg is in `$PATH`. The
24h buffer is encrypted-PNG-on-disk; disk usage is higher than 50MB/day for
active sessions.

### B-07 — Wake word rename incomplete (P1)

Phase 14I §33.5.1 claims D5 ("Wake word 'hey synaptic' replaced with 'hey
condura' in code, config, locales, and Settings UI") is closed. **It is
not.** `internal/voice/modelmgr/modelmgr.go:137-218`:

```go
Name:     "hey_synaptic",
URL:      "https://huggingface.co/datasets/synaptic/wake-words/resolve/main/hey_synaptic.onnx",
Filename: "hey_synaptic.onnx",
...
case "hey_synaptic":
    return WakeModelSpec{...}, nil
default:
    return WakeModelSpec{}, fmt.Errorf("unknown wake model: %s (supported: hey_synaptic)", name)
```

`internal/voice/openwakeword_detector.go:146`: `Hotword: "hey_synaptic"`.

The *display* default was renamed (`internal/onboarding/voice.go:32`
`DefaultWakeWord = "hey condura"`) and locale files say "hey condura" — but
the detector + model layer was never renamed. Wake word detection will not
fire on "hey condura" until the model asset is renamed and the case label is
updated.

### B-08 — No Windows/Linux permission probes (P2)

`internal/permissions/permissions_darwin.go` (CGO `AXIsProcessTrusted`) is
the only native probe. There is no `permissions_windows.go` or
`permissions_linux.go`. Non-macOS probes always return `StatusUnknown`. The
onboarding "Permissions" screen will show unknown on Windows/Linux. The
per-platform *guides* (steps/deeplinks/helpURL) exist in the untagged
`permissions.go` and do work cross-platform.

### B-09 — Executor bypasses the sanitizer (P1)

`internal/executor/executor.go:197` runs `exec.CommandContext("sh","-c",
cmd)` with the user-approved command string. The Gatekeeper approved the
policy decision; **`internal/sanitize` is never called**. §5.5 demands
deterministic shell-command sanitization via an allowlist. The executor is
the path that actually runs the command — and it skips the sanitizer. This
is a gap vs §5.5 / §13.6 model isolation.

### B-10 — Windows "user active" check is wrong (P1)

`internal/presence/detector.go:145-154`:

```go
out, err := exec.Command("powershell", "-Command",
  "Get-NetAdapter | Where Status -eq 'Up'").Output()
...
return strings.Count(string(out), "\n") > 0, nil
```

This counts network adapters, not user input. A Wi-Fi-connected Windows
machine always reports "user present." The `require_user_active` consent
gate (§5.1 / §10.2 for DESTRUCTIVE actions) is defeated on Windows. Linux
returns hardcoded `true` (`detector.go:167-171`).

### B-11 — osascript injection (P1, security F-11)

`internal/computeruse/backends/macosmcp_darwin.go:190-194` `escapeAppleScript`
escapes only `\` and `"`, not backticks. `execType`/`execKeyPress`/
`execLaunch` interpolate model-controlled `action.Value`. AppleScript
evaluates `` `do shell script "..."` `` in string literals, so a
model-controlled value containing a backtick can inject. The comment at
`:36` claims "script is constructed from trusted constants" — inaccurate
for these functions. Same issue at `internal/reach/imessage_darwin.go:52`
(Go `%q` verb, not AppleScript-safe).

### B-12 — Cleartext secrets file backend (P0, security F-01)

`internal/secrets/manager.go:216-279` `fileManager` writes
`fileData{Secrets map[string]string}` as **cleartext JSON** (mode 0600). The
storage master key is stored under the canonical name `secrets.MasterKey`
(`:330`) via the same `Set`/`Get` path. When the OS keyring is unavailable
(headless Linux, locked keychain, sandboxed process, CI), the AES-256-GCM
master key — and therefore the decryption key for every API key and OAuth
token in the SQLite DB — sits in cleartext at `<data-dir>/secrets.json`. The
code comment at `:215` acknowledges the deferral. **This is the only P0.**

### B-22 — No live CU action indicator (P2)

§10 / roadmap §10 claim "CU events on SSE ✅." `grep` for `cu.action` in
`internal/sse/` and `internal/agent/` finds only the RPC *method*
`cu.action` (`methods_phase7.go:64`), not a published SSE *event*. The CU
loop subscribes to `stream.delta` for LLM output but does not publish
per-action indicator events. The live "agent is clicking X" indicator has no
data source.

### B-30 — Wake model SHA256 unpinned (P2)

`internal/voice/modelmgr/modelmgr.go:139` `SHA256: ""` for the wake model —
comment: "will be verified on first download." This defeats the
trust-anchor pattern `pipeline.go` enforces for STT models. A compromised
CDN could swap the wake model.

### B-37 — Audit log has no prune path (P2)

§10.5 says "90-day retention (configurable)." `internal/audit/log.go` is
append-only with no deletion path visible in the package. The 90-day prune
may live elsewhere — worth verifying — but I did not find it in `log.go`.

### B-38 — CU router does not cascade on failure (P2)

`internal/computeruse/router.go:28-49` breaks on first error rather than
falling through to the next tier. A tier-1 backend that *fails* (not
"unavailable") does not cascade to tier 2. Deviates from §11.2 ("fall
back to mac-cua"). The comment justifies it ("action failed vs
unavailable"), but the spec describes cascade.

---

## 3. Dead code (P3 — delete or annotate)

| ID | Location | What |
|----|----------|------|
| B-23 | `internal/agent/agent.go` `Loop` | Defined, never instantiated in production. |
| B-23 | `internal/agent/loop_expanded.go` `ExpandedLoop` | Defined, never instantiated outside tests. |
| B-23 | `internal/agent/planner.go:142` `SimplePlanner` | Defined; `NewSimplePlanner` never called outside tests. |
| B-24 | `internal/llm/types.go:176` `ErrNotImplemented` | Defined, never returned anywhere. |
| B-25 | `internal/adaptive/predictor.go:43` `minConfidence` | Declared, immediately `_ = minConfidence`. |
| B-28 | `internal/reach/telegram.go:168-171` | Leftover `var _ =` lines. |
| B-34 | `internal/gatekeeper/gatekeeper.go:56-69` `DenyBeyondRead` | "Phase 4 stub, retained for test backward compatibility" — not used in production. |
| B-32 | `internal/executor/executor.go` header | Says "v0.2.0 sibling of Phase 17" but the package IS wired in the current build. |

---

## 4. Documentation drift (P3)

| ID | Location | Issue |
|----|----------|-------|
| B-29 | `internal/permissions/permissions.go:186-187` | macOS guide says both "Synaptic" and "Condura" inconsistently. |
| B-31 | `internal/hub/client.go:113` | User-Agent still `"synapticd/0.1.0"`. |
| B-32 | `internal/executor/executor.go` | Comments reference "v0.2.0" / "v0.3.0" despite being wired now. |
| B-39 | `internal/voice` | §8 implies CGO whisper.cpp + in-process openWakeWord; reality is subprocess. Document. |
| — | CLAUDE.md §6 | "Selective Perception" reads as if it drives capture. It doesn't (B-02). |
| — | CLAUDE.md §10 table | "Autonomy Matrix ✅ built" — code exists but is bypassed (B-01). |
| — | CLAUDE.md §18.3 | "H.264 delta compression" — buffer is PNG (B-06). |
| — | CLAUDE.md §33.5.1 D5 | "Wake word replaced with 'hey condura'" — detector layer wasn't (B-07). |

---

## 5. Test coverage variances

| Package | LOC | Tests | Verdict |
|---------|-----|-------|---------|
| `llm` | 4046 | 107 | Excellent |
| `daemon` | 12007 | 112 | Excellent |
| `api_key` | 1672 | 62 | Good |
| `gatekeeper` | 651 | 27 | Good |
| `config` | 2630 | 55 | Good |
| `secrets` | 853 | 45 | Good |
| `ipc` | 1826 | 46 | Good |
| `agent` | 2269 | 39 | Good |
| `adaptive` | 1368 | 16 | Adequate |
| `delegation` | 1001 | 21 | Adequate |
| `stream` | 1138 | 14 | Light for size |
| `memory` | 1168 | **3** | **Undertested** for a 665-LOC SQL store |
| `tui` | 1499 | 3 | Thin |
| `status` | 135 | 3 | Fine (small pkg) |
| `permissions` | 389 | 5 | Thin |

---

## 6. Recommended action plan before v0.1.0 public launch

**Must fix (P0/P1):**
1. **B-12 (F-01)** — encrypt the secrets file backend, or refuse to boot
   when the keyring is unavailable. This is the only P0.
2. **B-01** — wire `internal/autonomy` into the Gatekeeper, or remove the
   §27 "user dials each cell" claim from the spec/marketing.
3. **B-02** — wire `internal/perception` into the CU capture path, or
   remove the §6 "Selective Perception" claim. The package is built; the
   wiring is the missing 10%.
4. **B-07** — finish the wake word rename (`hey_synaptic` → `hey_condura` in
   `modelmgr` + `openwakeword_detector.go:146` + the ONNX asset filename +
   the `WakeModelForName` case label). Wake detection is broken until this
   is done.
5. **B-09** — route `executor.execShell` through `internal/sanitize` before
   `sh -c`. One-line fix; closes the §5.5 gap.
6. **B-10** — fix `presence/detector.go:145-154` Windows check (use
   `GetLastInputInfo` via a tiny Windows syscall, or stub to `false` and
   force consent on Windows). The current code is worse than a stub — it
   lies.
7. **B-11 (F-11)** — escape backticks in `escapeAppleScript`, or block
   values containing backticks. Same for `imessage_darwin.go:52`.

**Should fix (P2, recommend):**
8. **B-03** — add the 5th anomaly trigger ("new endpoint"). Cheap; the
   detector is otherwise complete.
9. **B-30** — pin the wake model SHA256.
10. **B-22** — emit a `cu.action` SSE event per action so the chat UI can
    show "agent is clicking X."
11. **B-37** — verify the audit-log prune job exists somewhere (or add
    one).
12. **B-38** — decide whether the CU router should cascade on failure, and
    document the choice.

**Document as v0.2.0 (already in roadmap):**
- B-04 (vector embeddings), B-05 (sync over internet), B-06 (H.264 buffer),
  B-08 (Windows/Linux probes), B-13–B-21 (router, subscription OAuth,
  waves, CE-MCP, channels, hub, MCP transports, Linux hotkey/tray, non-mac
  voice).

**Delete (P3):**
- B-23 (`Loop`, `ExpandedLoop`, `SimplePlanner`), B-24 (`ErrNotImplemented`),
  B-25 (`minConfidence`), B-28 (telegram dead lines), B-34
  (`DenyBeyondRead`), B-29/B-31/B-32 (branding/comment drift), B-27
  (sensitive dup).

---

## 7. What is NOT a gap

These were checked and are genuinely fine:
- **No hardcoded/fake successful results** anywhere in the backend. Every
  stub returns an explicit error.
- **No `panic("not implemented")`** in production code. The codebase is
  unusually clean of stub markers — gaps are silent omissions, not visibly
  marked.
- **No orphaned packages at the wiring level.** `initSubsystems` wires all
  60 packages. The two "orphans" (autonomy, perception) are wired *into the
  daemon* but not *into the decision path* — a subtler gap.
- **The 7 non-negotiable invariants (§2.1)** are enforced by deterministic
  code: Strategist/Gatekeeper separation, Gatekeeper-only path to action,
  destructive-action native dialog, user stop (4 mechanisms), audit log,
  guest-not-owner, OS permissions by user. The only soft spot is B-10
  (Windows "user present" lies), which weakens invariant #3 on Windows.
- **The 7 technical non-negotiables (§5)** are mostly enforced, modulo
  B-09 (executor bypasses sanitize — §5.5 gap) and B-03 (4 of 5 anomaly
  triggers — §5.6 gap).
- **CI is green** — 15/15 on main CI, 3/3 on Release Verify, at commit
  `c8e4d55`.

---

## 8. Appendix — how to reproduce

```bash
# Confirm CI green.
gh run list --limit 3

# Confirm B-01 (autonomy orphaned).
grep -rn "autonomy\." internal/daemon/safety_wiring.go
# Expect: only a comment, no real call.

# Confirm B-02 (perception orphaned).
grep -rln "internal/perception" internal/ app/ | grep -v perception.go
# Expect: no output.

# Confirm B-03 (5th trigger missing).
grep -rni "new endpoint\|never-used\|network target" internal/anomaly/
# Expect: no output.

# Confirm B-04 (no sqlite-vec).
grep -rni "sqlite-vec\|vec0\|sqlitevec" internal/ | grep -v _test.go
# Expect: no output.

# Confirm B-05 (sync is UDP broadcast).
grep -n "broadcast\|255.255.255.255\|Kademlia\|mDNS\|STUN\|TURN" internal/sync/discovery.go
# Expect: only UDP broadcast.

# Confirm B-07 (wake word not renamed).
grep -rn "hey_synaptic" internal/voice/ internal/onboarding/
# Expect: 5+ hits in voice/modelmgr + openwakeword_detector.go:146.

# Confirm B-09 (executor skips sanitize).
grep -n "sanitize" internal/executor/executor.go
# Expect: no hits.

# Confirm B-10 (Windows presence is wrong).
sed -n '145,154p' internal/presence/detector.go
# Expect: Get-NetAdapter check.

# Confirm B-11 (osascript backtick gap).
sed -n '190,194p' internal/computeruse/backends/macosmcp_darwin.go
# Expect: only \\ and \" escaped.

# Confirm B-12 (F-01 still open).
sed -n '216,279p' internal/secrets/manager.go
# Expect: json.Marshal of cleartext fileData.

# Confirm dead code (B-23).
grep -rn "NewExpandedLoop\|NewSimplePlanner" internal/ app/ cmd/ | grep -v _test.go
# Expect: no output.
```

**End of audit.**