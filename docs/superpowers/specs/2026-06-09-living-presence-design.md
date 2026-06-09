# Phase 4 — The Living Presence (Complete Specification)

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
only later, behind a safety seam that physically cannot be bypassessed.

**In one sentence:** turn Synaptic from a chat window into a living
presence — a tray/menu-bar agent that you summon by holding a hotkey,
speak to with your voice, and hear answer back — entirely on-device,
with a deterministic safety seam that makes acting on your OS impossible
until Phase 5 earns it.

**The through-line:** every sub-phase is additive. The 26 existing
packages and the streaming pipeline are reused, never rewritten. The
working tree never goes red from structural churn.

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
- No mid-stream resume / SSE replay, no per-conversation SSE topic
  filtering (carried over from Session 7's deferred list).

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
| 4H | Whisper model | **base (multilingual ggml-base.bin)**, ~142MB. Small (~466MB) exposed as user upgrade in Settings | Architect, 2026-06-09 |
| 4I | Audio capture | **malgo** (miniaudio bindings), gated behind a build smoke test. CGO, covers CoreAudio/WASAPI/ALSA/PulseAudio | Architect, 2026-06-09 |

## 4. Guiding Principles

1. **Additive only.** New packages plus one new overlay window. The
   existing 26 green packages and the streaming pipeline are **reused, not
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
| `internal/voice` | `Recorder` (mic capture via malgo), `Transcriber` (whisper subprocess), `Speaker` (native TTS). All interfaces, swappable impls, per-platform files. |
| `internal/voice/modelmgr` | Model + binary lifecycle: `EnsureModel()`, download, SHA-256 verify, atomic rename. |
| `internal/overlay` | Overlay-window lifecycle controller (show / hide / state machine), behind an interface so the Wails-window impl can be swapped for a native fallback. |
| `internal/agent` | The **thin** loop: utterance → Gatekeeper check → existing `stream.Manager` → text + TTS + audit. Not the full planner — just enough to converse. |
| `internal/presence` | Orchestration glue: gesture → session lifecycle (hotkey → overlay → capture → transcribe → agent → TTS). |

**Frontend (additive).** A new overlay window rendering the voice orb
(animated waveform, MISSION §19.4), hand-rolled CSS.

**Daemon (additive).** New JSON-RPC methods `voice.*`, `overlay.*`,
`agent.*`, `presence.*`. New SSE events `voice.*`, `agent.*`.

## 6. The Press-and-Hold Flow

1. **Hold hotkey** → `internal/hotkey` fires `StartHold(onDown, onUp)` →
   `overlay.show(AtCursor)` + `voice.startCapture(request_id)` → state:
   Listening.
2. **Speaking** → whisper partials → SSE `voice.partial` → orb shows live
   transcription, waveform animates via `voice.level`.
3. **Release** (after `min_ms` debounce) → `voice.stopCapture` → final
   transcript → SSE `voice.final` → `agent.ask`.
4. **Answer** → Gatekeeper classifies (chat = READ → allow) → reuse
   `stream.Manager` → `stream.delta` renders in orb → on finish, native
   TTS speaks the answer → `agent.speaking`.
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

**There is exactly one path from intent to action —
`gatekeeper.Evaluate` — and in Phase 4 it denies everything but READ.
No sub-phase may add a second path.**

---

## 8. Sub-Phase Specifications

### 4.0 — Safety Seam ✅ COMPLETED

**Commit:** `061eb14`, pushed.
**Packages:** `internal/blastradius`, `internal/gatekeeper`.

- `blastradius`: `Action`, `Class` (READ/WRITE/NETWORK/DESTRUCTIVE),
  `Classify()`, `Class.String()`. Unknown → DESTRUCTIVE. 100% coverage.
- `gatekeeper`: `Gatekeeper` interface, `Decision` (Allow/Deny),
  `DenyBeyondRead`, `Evaluate()`. Allows READ, denies the rest with a
  class-named reason. 100% coverage.
- Verified: 26 pkgs green `-race`, lint 0, LOGBOOK Session 8.

---

### 4.1 — Local Speech Engine (the largest lift)

**Purpose:** capture the microphone, transcribe speech to text locally
via whisper, and speak text back via the OS-native voice — on all three
platforms, with no cloud calls.

#### Interfaces (the contract)

```go
package voice

// Recorder captures audio from the microphone.
type Recorder interface {
    // Start begins mic capture. Blocks until ctx is cancelled or Stop is called.
    Start(ctx context.Context) error
    // Stop halts capture and returns the recorded audio as WAV bytes.
    Stop() ([]byte, error)
    // Samples returns a channel of live PCM samples for waveform/VAD.
    Samples() <-chan []float32
}

// Transcriber converts audio bytes to text.
type Transcriber interface {
    // Transcribe performs a one-shot transcription of a complete audio clip.
    Transcribe(ctx context.Context, audio []byte) (Transcript, error)
    // TranscribeStream processes live audio samples and emits partial results.
    TranscribeStream(ctx context.Context, audio <-chan []float32) (<-chan Partial, error)
}

// Speaker converts text to speech using the OS-native voice.
type Speaker interface {
    // Speak blocks until the text is spoken or ctx is cancelled.
    Speak(ctx context.Context, text string) error
    // Stop halts any in-progress speech.
    Stop()
}
```

#### Value types

```go
type Transcript struct {
    Text       string
    Language   string
    Confidence float64
    Segments   []Segment
}

type Partial struct {
    Text    string
    IsFinal bool
}

type Segment struct {
    Start float64
    End   float64
    Text  string
}
```

#### Concrete implementations & files

| File | Purpose |
|---|---|
| `recorder_darwin.go` | Mic capture via malgo (CoreAudio) |
| `recorder_windows.go` | Mic capture via malgo (WASAPI) — may share code with darwin if malgo abstracts |
| `recorder_linux.go` | Mic capture via malgo (ALSA/PulseAudio) |
| `whisper_transcriber.go` | Subprocess wrapper: locates `whisper-cli`, spawns with model path + audio, parses `--output-json`. Streaming partials via incremental stdout. |
| `fake_transcriber_test.go` | Deterministic transcriber over recorded-WAV fixtures (test-only) |
| `speaker_darwin.go` | Native `say` command with voice/rate flags |
| `speaker_windows.go` | SAPI via PowerShell `System.Speech.Synthesis` or a small bridge |
| `speaker_linux.go` | `espeak-ng` / `festival` / `spd-say`, detected at runtime; graceful "no TTS engine found" fallback |
| `vad.go` | Voice-activity detection: energy-threshold based, trims leading/trailing silence. Pure Go. |

#### Model + binary distribution (first-run flow)

New sub-package: `internal/voice/modelmgr`.

```go
package modelmgr

// EnsureModel checks for the whisper model and binary at modelDir.
// If missing, downloads from pinned URLs over HTTPS, verifies SHA-256,
// and atomically renames into place. Returns the path on success.
func EnsureModel(ctx context.Context, spec ModelSpec, modelDir string) (string, error)

type ModelSpec struct {
    Name     string // "base", "small", etc.
    URL      string // pinned download URL
    SHA256   string // expected hex hash
    Filename string // e.g. "ggml-base.bin"
}
```

- Model directory: `~/.synaptic/models/`
- Download size: base ~142MB, small ~466MB
- Model is never bundled (daemon < 20MB, decision 4F)
- Whisper binary per-platform: downloaded or bundled if signed and small enough
- Atomic rename prevents partial files on crash
- Resume support for interrupted downloads

#### Config additions (`internal/config/config.go`)

New `VoiceConfig` under `Config`:

```yaml
voice:
  enabled: true
  stt:
    engine: whisper          # whisper | none
    model: base              # tiny | base | small | medium
    model_dir: ~/.synaptic/models
    language: auto           # auto | en | es | fr | de | ja | zh
    binary_path: ""          # "" = auto-detect/download
  tts:
    engine: native           # native | none
    voice: ""                # "" = OS default
    rate: 1.0
  push_to_talk:
    enabled: true
    min_ms: 200              # ignore accidental taps
  download:
    auto: true               # prompt vs auto-download model on first use
```

- Bump `ConfigSchemaVersion` 2 → 3
- Add migration in config loader
- Document defaults in `configs/default.yaml`

#### Daemon JSON-RPC methods (new `methods_voice.go`)

| Method | Params | Returns | Description |
|---|---|---|---|
| `voice.startCapture` | `{request_id}` | `{started: true}` | Begins mic capture. Publishes `voice.partial` as text arrives. |
| `voice.stopCapture` | `{request_id}` | `{text, language, confidence}` | Stops capture, runs final transcription, publishes `voice.final`. |
| `voice.transcribeFile` | `{path}` | `{text, language, confidence}` | Transcribe a WAV file (utility/testing). |
| `voice.speak` | `{text}` | `{started: true}` | TTS a string (gated: TTS is READ-class, Gatekeeper allows). |
| `voice.status` | `{}` | `{engine, model_present, download_progress}` | Engine readiness. |
| `voice.downloadModel` | `{model}` | `{started: true}` | Trigger/resume model download; publishes `voice.download.progress`. |

#### SSE events (new)

| Event | Payload | Trigger |
|---|---|---|
| `voice.partial` | `{request_id, text}` | Live transcription arrives |
| `voice.final` | `{request_id, text, language, confidence}` | Final transcription complete |
| `voice.level` | `{request_id, rms}` | Waveform amplitude (throttled ~30Hz) |
| `voice.download.progress` | `{bytes, total, done}` | Model download progress |
| `voice.error` | `{request_id, error}` | Capture/transcription error |

#### Permissions

- **macOS:** `NSMicrophoneUsageDescription` in `Info.plist`. TCC prompt on first capture.
- **Windows:** Microphone privacy setting; handle "access denied" gracefully.
- **Linux:** PulseAudio/PipeWire access; portal where sandboxed.

#### Dependencies (documented in MISSION §8 per Hard Rule #4)

- `github.com/gen2brain/malgo` — audio capture, CGO, permissive license
- External runtime: `whisper-cli` binary + `ggml-base.bin` model (downloaded, not vendored)

#### Tests

- Voice loop with `fake_transcriber` + recorded-WAV fixtures (deterministic text)
- `whisper_transcriber` against a stub whisper script (tiny Go fake on `$PATH` that emits known JSON) — no real model in CI
- `modelmgr`: download from `httptest.Server`, SHA-256 mismatch rejection, atomic-rename, resume
- Speaker: each platform impl behind a `Speaker` fake; real impls smoke-tested with `--version` probe (skipped on CI with no audio)
- VAD: silence vs speech fixtures
- Target >80% on voice; loop/parsing logic ~100%, platform mic/TTS impls covered by interface fakes

#### CI implications

- Add audio CGO deps (ALSA/PulseAudio headers on Linux) to lint/test/build jobs
- All real-audio paths `t.Skip` when no device (mirrors existing tray/keyring CI skips)

#### Definition of done (4.1)

Transcribe a fixture WAV to correct text via the subprocess path; speak a string on each OS (manually verified on macOS, smoke-probed elsewhere); model auto-downloads + verifies on first run; daemon < 20MB; race/lint green; LOGBOOK updated.

---

### 4.2 — Overlay Window + Voice Orb (highest technical risk)

**Purpose:** the floating, frameless, always-on-top voice orb that slides
up on summon (MISSION §19.2, §19.4).

#### New package: `internal/overlay`

```go
package overlay

// Controller manages the overlay window lifecycle.
type Controller interface {
    Show(ctx context.Context, opts ShowOpts) error
    Hide() error
    Toggle()
    State() OverlayState
    OnDismiss(func())
}

type OverlayState int
const (
    StateHidden OverlayState = iota
    StateListening
    StateThinking
    StateSpeaking
)

type ShowOpts struct {
    AtCursor bool
    X, Y     int
}
```

#### Implementations

| File | Description |
|---|---|
| `noop_controller.go` | **Real headless controller** — full state machine (Hidden→Listening→Thinking→Speaking), unit-tested, no window painting. Permanent fallback if multi-window is unstable. |
| `wails_controller.go` | Real second Wails window (frameless, transparent, always-on-top). Built only if the multi-window spike succeeds. |

**Key:** `noop_controller` is NOT a literal no-op. It implements the
full state machine, tracks transitions, fires `OnDismiss`, and is
unit-tested against the same spec as the real impl. 4.3/4.4 get genuine
logic coverage, not stubs.

#### Wails multi-window reality (the documented risk)

- Today `app/web/main.go` runs one window (1200×800). `app.go:89`
  already calls `WindowSetAlwaysOnTop`.
- Phase 4.2 must create a second frameless/transparent/always-on-top
  window. Wails v2's multi-window support is limited and uneven across
  platforms.
- **Strategy:** spike a 2nd-window prototype as a throwaway outside the
  main tree. Prove frameless + transparent + always-on-top on macOS
  before writing `wails_controller.go`. Confirm-then-build.
- If the spike fails, `noop_controller` is the permanent path, documented
  as an ADR.

#### Frontend (hand-rolled CSS, STYLE.md §17)

| File | Purpose |
|---|---|
| `Overlay.svelte` | Overlay window entry point (or separate Wails window entry) |
| `VoiceOrb.svelte` | Animated waveform, pulsing dots, color reflects confidence (§19.4); state-driven (listening/thinking/speaking) |
| `LiveTranscript.svelte` | Partial text as you speak |
| `voice.svelte.ts` | Runes store: capture state, partial/final text, level, speaking flag; subscribes to SSE `voice.*` + reused `stream.*` events |
| `overlay.css` | Vibrancy/acrylic backdrop, slide-up 200ms + fade entrance, transparent frame |

Keybindings in-overlay: `Esc` dismiss, `Cmd+Enter` submit, `Cmd+K` palette (§19.2).

#### Daemon methods (replace stubs in `methods_more.go`)

| Method | Description |
|---|---|
| `overlay.show` | Show overlay (real impl or noop_controller) + audit `window.event` |
| `overlay.hide` | Hide overlay |
| `overlay.state` | Return current `OverlayState` |

#### Performance budget (STYLE.md §17 — hard)

**Hotkey → overlay visible < 100ms.** Window must be pre-created hidden
at startup and shown, not constructed on demand.

#### Tests

- Controller state-machine tests (Hidden→Listening→Thinking→Speaking→Hidden) against noop_controller
- Frontend: orb renders each state; store transitions on mocked SSE events
- Manual: visual verification on real desktop session

#### Definition of done (4.2)

Overlay shows/hides via method + (later) hotkey; orb animates through all states; <100ms show; multi-window decision recorded as ADR; race/lint green; LOGBOOK + ADR.

---

### 4.3 — Push-to-Talk Wiring

**Purpose:** connect the physical gesture — hold the hotkey, speak,
release — to overlay + capture + transcribe.

#### The hotkey gap (must be fixed here)

`internal/hotkey.Manager` exposes `New(spec)`, `Start(handler func())`,
`Stop()`, `PressCount()`, `DefaultOverlay()` — a single press callback
only. Push-to-talk needs key-down and key-up as distinct events.

**4.3 extends hotkey:** add `StartHold(onDown func(), onUp func()) error`
using `golang.design/x/hotkey`'s `Keydown()`/`Keyup()` channels. Keep
`Start()` for tap-style hotkeys (kill-switch). Add a `min_ms` debounce
(config `push_to_talk.min_ms`) to ignore accidental taps.

**Platform note:** verify keyup delivery on each OS (x/hotkey behavior
differs; Carbon on mac, Win32 hooks, X11/evdev on Linux). Build-tagged
handling if needed.

#### Orchestration (`internal/presence`)

The glue that owns the gesture → session lifecycle:

1. Hotkey down → `overlay.show(AtCursor)` + `voice.startCapture(request_id)` → state: Listening
2. Whisper partials → `voice.partial` → orb live transcript
3. Hotkey up (after `min_ms`) → `voice.stopCapture` → `voice.final` → hand off to `agent.ask`
4. Esc / 5s inactivity → `overlay.hide` + cancel capture

#### Kill-switch integration (MISSION §5.3)

Before starting capture, check `halt.IsHalted()`; if halted, refuse +
audit. While capturing, a halt cancels capture, stream, and TTS
immediately. Wire halt into the presence orchestrator.

#### Daemon methods

| Method | Description |
|---|---|
| `presence.summon` | Programmatic equivalent of the gesture (for testing + GUI button) |
| `presence.dismiss` | Dismiss the overlay and cancel any active capture |

#### Tests

- Hotkey hold logic with a fake key-event source: down→up fires both callbacks; sub-min_ms tap is ignored
- Presence orchestrator with fake overlay + fake voice: full down→speak→up→final sequence; halt mid-capture aborts
- Race/lint green

#### Definition of done (4.3)

Holding the configured hotkey shows the orb and captures; releasing transcribes and emits `voice.final`; tap-debounce works; halt aborts cleanly; tests green; LOGBOOK.

---

### 4.4 — Thin Agent Loop

**Purpose:** the smallest real agent — take a spoken utterance, route it
through the Gatekeeper, answer via the existing streaming pipeline, speak
the answer, and audit the whole turn.

#### New package: `internal/agent`

```go
package agent

// Loop is the thin agent loop. Dependencies are injected.
type Loop struct {
    Gatekeeper   gatekeeper.Gatekeeper
    Stream       *stream.Manager
    Speaker      voice.Speaker
    Audit        *audit.Log
    Conversations *conversation.Store
}

type AskRequest struct {
    ConversationID int
    Text           string
    RequestID      string
    Spoken         bool // if true, speak the answer via TTS
}

type AskResult struct {
    RequestID string
    Finish    string // "stop", "blocked", etc.
}
```

#### Flow

a. Build `blastradius.Action{Kind:"chat"}` for the utterance →
   `gatekeeper.Evaluate` → (READ → Allow). Audit the decision (MISSION §5.4).
b. Append user message to `conversation.Store`.
c. Build `stream.Request{ConversationID, ProviderName, Model, Messages,
   RequestID}` and call `stream.Manager.Start` → tokens flow over existing
   `stream.delta` SSE.
d. On `stream.finished`, persist assistant message; if `Spoken`, call
   `Speaker.Speak` and emit `agent.speaking`.
e. Any non-chat intent the model might later produce (a tool call to
   click/type/exec) → classified non-READ → Gatekeeper denies → audited +
   surfaced to user as "blocked: safety layer pending." This is the proof
   the seam holds end-to-end.

#### Daemon methods (`methods_agent.go`)

| Method | Description |
|---|---|
| `agent.ask` | Drives `Loop.Ask`; returns `request_id` |
| `agent.cancel` | Cancels the underlying stream (`stream.Manager.Cancel`) + stops TTS |

#### SSE events (new)

| Event | Payload |
|---|---|
| `agent.speaking` | `{request_id, started\|stopped}` |
| `agent.blocked` | `{request_id, action_kind, class, reason}` |

#### Audit (MISSION §5.4 — mandatory)

Every turn writes: utterance received, gatekeeper decision + reason +
class, provider/model chosen, finish reason, TTS spoken. Reuse
`internal/audit` + `audit_consts.go` actor/app/level constants.

#### Tests

- Reuse Session 7's fake stream provider: `agent.ask` produces tokens end-to-end over real SSE
- **Gatekeeper-consulted test:** a fabricated non-READ action is refused, `agent.blocked` emitted, audit row written. *(This is the most important test in the phase — it proves the safety invariant.)*
- TTS via Speaker fake; assert `agent.speaking` start/stop
- Conversation persistence asserted
- Race/lint green; >80%

#### Definition of done (4.4)

Speak → text → answer streams into the orb → answer is spoken; every turn audited; a non-READ action is provably denied by the Gatekeeper end-to-end; tests green; LOGBOOK.

---

### 4.5 — Polish & Onboarding Integration

**Purpose:** make it feel finished and wire it into first-run.

#### Items (each small, all required)

1. **Waveform animation** tuned to real `voice.level` data; confidence-colored orb (§19.4)
2. **Auto-dismiss** after 5s inactivity; pin to keep open; smooth exit animation
3. **Tray/menu-bar presence** (`internal/tray`): reflect state — idle / listening / thinking. Add listening indicator + "Hold \<hotkey\> to talk" hint + quick "Talk now" item
4. **Onboarding voice test** (MISSION §20 screen 7): real "Say something" step in `OnboardingWizard.svelte` that records, transcribes, and shows the text — also the natural trigger for first-run model download
5. **Mic permission prompt** woven into onboarding (request before first capture; show grant instructions per OS)
6. **i18n** (MISSION §23): all new UI strings into 6 locale catalogs (en/es/fr/de/ja/zh); whisper language: auto already handles spoken language
7. **Settings panel for voice:** engine on/off, model size (with download), TTS voice/rate, push-to-talk hotkey + min_ms
8. **Error states:** no mic, no TTS engine, model download failed, whisper binary missing — each with clear message and recovery path
9. **Performance pass:** confirm hotkey→overlay <100ms, first token <1.5s budgets (STYLE.md §17) hold with voice in the loop

#### Tests

- Onboarding voice-test step (component + store with mocked transcribe)
- Settings voice panel round-trips config
- i18n: no missing keys across locales (catalog completeness test)
- Error-state rendering for each failure

#### Definition of done (4.5 = Phase 4 done)

The full press-and-hold voice loop works on macOS (primary), smoke-verified on Windows/Linux; onboarding includes a working voice test + model download; every non-READ action is provably refused by the Gatekeeper; all budgets met; race/lint green; LOGBOOK closes Phase 4.

---

## 9. Cross-Cutting Concerns

- **Safety invariant (MISSION §2):** exactly one path from intent to action — `gatekeeper.Evaluate` — and in Phase 4 it denies everything but READ. No sub-phase may add a second path.
- **Audit everything (§5.4):** capture start/stop, transcription, gatekeeper decisions, model calls, TTS — all into the HMAC-chained audit log.
- **Kill-switch (§5.3):** halt must abort capture, stream, and TTS at any instant; wired in 4.3/4.4.
- **Performance budgets (STYLE.md §17):** hotkey→overlay <100ms, first token <1.5s, idle mem <150MB, daemon binary <20MB (model excluded).
- **Local-first / privacy:** audio never leaves the machine; no telemetry of speech content; model/binary fetched over verified HTTPS only.
- **Quality bar (STYLE.md §4):** every sub-phase — tests written test-first and green with `-race`, lint 0, committed to main, pushed after verification, LOGBOOK entry.
- **New dependencies** (malgo, whisper binary/model, any TTS bridge) documented in MISSION §8 per Hard Rule #4. CI updated for audio CGO deps across OSes.
- **Docs:** an ADR for the multi-window decision; architecture doc for voice + overlay + agent under `docs/architecture/`.

## 10. Honest Risks

1. **Wails multi-window for the overlay (highest risk).** Cross-platform frameless + transparent + always-on-top is not Wails v2's strength. Mitigation: spike early, keep behind the `overlay` interface, native fallback if unstable.
2. **20 MB binary budget vs whisper model.** Resolved by 4F: daemon stays < 20 MB; whisper binary + model download on first run.
3. **Mic permission (TCC / Windows / Linux portal).** New permission surface. 4.1 adds the minimum needed, folds prompting into onboarding.
4. **malgo CGO cross-platform build.** Gated behind a smoke test. If a platform fails, evaluate PortAudio fallback.
5. **Timeline.** Local + cross-platform is the slowest combination. Realistic: a couple of weeks of focused sessions.

## 11. Definition of Done (Phase 4 as a whole)

Per `STYLE.md` §4: code written, tests written and green with `-race`,
lint 0 issues, committed to `main` with a conventional-commit message,
`LOGBOOK.md` updated. Phase 4 as a whole is done when the press-and-hold
voice loop works on all three platforms and every non-READ action is
provably refused by the Gatekeeper.

## 12. What Phase 4 Deliberately Does NOT Include

- No wake word ("hey synaptic") — Phase 4+ / later
- No cloud STT/TTS — local only
- No computer-use: no clicking, typing, shell, file writes, or network actions on your behalf — those are Phase 5
- No mid-stream resume / SSE replay, no per-conversation SSE topic filtering
