# Phase 4 — The Living Presence (Design Spec)

> **Status:** Approved 2026-06-09. Implementation in progress.
> **Author session:** Claude Opus 4.8 (Claude Code), partner-implementer.
> **Read first:** `MISSION.md` §2 (Survival Rule), §6 (Selective Perception),
> §10 (Safety Layer), §11 (Computer-Use), §19 (Hotkey + Overlay + Voice),
> §21 (Interfaces), and `LOGBOOK.md` Session 7 (streaming pipeline).

---

## 1. Why This Phase Exists

Synaptic today (through Session 7) is **excellent plumbing with a kill
switch**: 22 internal packages — storage, IPC, LLM clients, SSE, a real
streaming pipeline, `audit`, and `halt` — plus a Wails + Svelte 5 chat
shell. What it is *not*, yet, is alive. There is no agent loop, no voice,
no overlay, no menu-bar presence that does anything, and — critically —
none of the deterministic safety core (gatekeeper, blast-radius
classifier, anomaly detector, sanitizers). Functionally it reads as
"Codex in an app," which is exactly the feeling the architect named.

The experience the architect wants — **press-and-hold a global hotkey, a
voice orb slides up, you speak, the agent answers in voice** — is not a
new direction. It is already MISSION §19 + §6 + §21. Phase 4 pulls those
forward and makes them real, while honoring the Survival Rule: the agent
gains *presence and voice* now, and gains *the ability to act on the OS*
only later, behind a safety seam that physically cannot be bypassed.

## 2. Goal and Non-Goals

**Goal.** A persistent menu-bar / tray agent that, on press-and-hold of
the global hotkey, shows a voice orb, transcribes speech **locally**,
answers via the existing streaming pipeline in streamed text **and**
spoken voice — on macOS, Windows, and Linux.

**Non-Goals (this phase).**
- No clicking, typing, shell execution, file writes, or any action *on
  the user's behalf*. Those are classified non-READ and **denied** by the
  Gatekeeper until Phase 5 builds the real rules engine.
- No wake word ("hey synaptic") — push-to-talk only this phase.
- No cloud STT/TTS — fully local, $0 runtime for voice.

## 3. Locked Decisions for This Phase

| # | Decision | Value | Source |
|---|---|---|---|
| 4A | Voice trigger | **Push-to-talk only** (hold hotkey → speak → release) | Architect, 2026-06-09 |
| 4B | Speech stack | **Fully local** — whisper.cpp (STT) + native OS TTS | Architect, 2026-06-09 |
| 4C | whisper integration | **Subprocess** to a `whisper` binary (binary + model download on first run) | Architect, 2026-06-09 |
| 4D | Platform scope | **Cross-platform from the start** (macOS + Windows + Linux) | Architect, 2026-06-09 |
| 4E | Sequencing | **Hybrid** — experience now, behind a deny-by-default Gatekeeper | Architect, 2026-06-09 |
| 4F | Binary budget | Daemon stays **< 20 MB**; whisper model downloads separately on first run | Architect, 2026-06-09 |
| 4G | Git workflow | Commit each green sub-phase to `main`; push at end after full verification | Architect, 2026-06-09 |

## 4. Guiding Principles

1. **Additive only.** New packages plus one new overlay window. The
   existing 24 green packages and the streaming pipeline are **reused, not
   modified**. The current working tree must never go red because of
   structural churn.
2. **Gatekeeper-gated from line one.** Every action a voice utterance
   could produce passes a `Gatekeeper` interface. Until the real rules
   engine exists, the implementation **denies everything beyond
   READ/chat**. The agent cannot act unsafely because that code path does
   not exist yet.
3. **Local-first, $0 voice runtime.** whisper.cpp + native TTS. No cloud
   calls for speech.
4. **Match the codebase.** Flat `internal/<pkg>` layout, package-doc
   headers, stdlib `testing` with `t.TempDir()`/`t.Helper()`, no testify,
   hand-rolled CSS (no Tailwind). >80% coverage per package, 0 lint, race
   clean — per `STYLE.md` §4 and §17.

## 5. New Components

| Package | Purpose |
|---|---|
| `internal/blastradius` | Deterministic action classifier: READ / WRITE / NETWORK / DESTRUCTIVE. Unknown actions classify as DESTRUCTIVE (most conservative). |
| `internal/gatekeeper` | The safety seam. `Gatekeeper` interface + v0 `DenyBeyondRead` impl: allow READ, deny all else with a clear reason. The real rules engine (MISSION §10.2) fills in behind the same interface in Phase 5. |
| `internal/voice` | `Recorder` (mic capture), `Transcriber` (whisper subprocess), `Speaker` (native TTS). All interfaces, swappable impls, per-platform files. |
| `internal/overlay` | Overlay-window lifecycle controller (show / hide / state machine), behind an interface so the Wails-window impl can be swapped for a native fallback. |
| `internal/agent` | The **thin** loop: utterance → Gatekeeper check → existing `stream.Manager` → text + TTS + audit. Not the full planner — just enough to converse. |

**Frontend (additive).** A new overlay window rendering the voice orb
(animated waveform, MISSION §19.4), hand-rolled CSS.

**Daemon (additive).** New JSON-RPC methods `voice.startCapture`,
`voice.stopCapture`, real `overlay.show` / `overlay.hide` (currently
audit-only stubs in `methods_more.go`), and `agent.ask`. New SSE events
`voice.partial`, `voice.final`, `agent.speaking` (in addition to the
existing `stream.*` events).

## 6. The Press-and-Hold Flow

1. **Hold hotkey** → `internal/hotkey` (already built) fires →
   `overlay.show` + `voice.startCapture` (mic on).
2. **Speaking** → whisper subprocess emits partial text → SSE
   `voice.partial` → orb shows live transcription, waveform animates.
3. **Release** → `voice.stopCapture` → final transcript →
   SSE `voice.final` → `agent.ask`.
4. **Answer** → Gatekeeper classifies (chat = READ → allow) → reuse
   `stream.Manager` → `stream.delta` renders in orb → on finish, native
   TTS speaks the answer (`agent.speaking`).
5. **Dismiss** → `Esc` or 5 s inactivity → `overlay.hide`. The existing
   `halt` kill-switch hard-stops capture, stream, and TTS at any point.

## 7. The Safety Seam (the heart of the hybrid)

`blastradius.Classify(Action) Class` maps a proposed action to its blast
radius. v0 needs only to distinguish READ (transcribe, speak, chat
completion, screenshot-for-reading) from everything else; unknown action
kinds classify as DESTRUCTIVE so the default is maximal caution.

`gatekeeper.Gatekeeper` is `Evaluate(ctx, Action) (Decision, reason)`.
The v0 `DenyBeyondRead` implementation returns `Allow` for READ and
`Deny` for all else, with reason *"safety layer not yet implemented
(Phase 5); only READ/chat actions are permitted."* The real
rules-engine implementation (policy.yaml, consent dialogs, queueing —
MISSION §10.2) is a drop-in replacement behind this interface.

This is what makes the hybrid honest: the moment any future code tries to
make the agent click, type, or exec, it must route through `Evaluate`,
and v0 denies it. There is no second path to physical action.

## 8. Sub-Phases (sequenced; safety first per MISSION §10)

- **4.0 — Safety seam.** `internal/blastradius` + `internal/gatekeeper`
  (deny-beyond-read). Pure Go, fully unit-tested. *No native risk. This is
  the keystone and is built first.*
- **4.1 — Local speech.** `internal/voice`: `Recorder`, whisper
  `Transcriber` (subprocess), native `Speaker`, per platform. First-run
  model+binary download flow. *Largest lift.*
- **4.2 — Overlay window + orb.** New Wails window + voice-orb UI.
  Multi-window spike + fallback decision (see Risks).
- **4.3 — Push-to-talk wiring.** Hotkey hold → overlay → capture →
  transcribe; SSE `voice.partial` / `voice.final`.
- **4.4 — Thin agent loop.** `agent.ask` → Gatekeeper → `stream.Manager`
  → TTS; audit every turn (MISSION §5.4).
- **4.5 — Polish.** Waveform animation, auto-dismiss, kill-switch
  interplay, audit of every voice session, onboarding voice test (§20).

## 9. Testing Strategy

- **4.0:** table tests for every `Class` mapping incl. the
  unknown→DESTRUCTIVE default; gatekeeper allow-READ / deny-rest with
  reason assertions. Target ~100% (pure logic).
- **4.1:** `Transcriber` and `Speaker` are interfaces; unit-test the loop
  with a fake transcriber over recorded-WAV fixtures and a fake speaker;
  test the subprocess wrapper with a stub `whisper` script.
- **4.4:** reuse Session 7's fake stream provider; assert Gatekeeper is
  consulted and a non-READ action is refused end-to-end.
- All sub-phases: `go test -race -count=1 -timeout=120s ./...` green,
  `golangci-lint run ./...` 0 issues, Wails `.app` < 20 MB.

## 10. Honest Risks

1. **Wails multi-window for the overlay (highest risk).** Cross-platform
   frameless + transparent + always-on-top is not Wails v2's strength.
   Mitigation: spike early in 4.2, keep behind the `overlay` interface,
   native fallback if unstable. *Does not affect 4.0–4.1.*
2. **20 MB binary budget vs whisper model.** Resolved by 4F: the daemon
   stays < 20 MB; whisper binary + model download on first run, wired into
   the onboarding voice test.
3. **Mic permission (TCC / Windows / Linux portal).** New permission
   surface; no `permissions` package exists yet (only referenced in
   MISSION §29). 4.1 adds the minimum needed, or folds prompting into
   onboarding.
4. **Timeline.** Local + cross-platform is the slowest of the offered
   combinations. Realistic: a couple of weeks of focused sessions, not a
   weekend. Accepted by the architect.

## 11. Definition of Done (per sub-phase)

Per `STYLE.md` §4: code written, tests written and green with `-race`,
lint 0 issues, committed to `main` with a conventional-commit message,
`LOGBOOK.md` updated. Phase 4 as a whole is done when the press-and-hold
voice loop works on all three platforms and every non-READ action is
provably refused by the Gatekeeper.
