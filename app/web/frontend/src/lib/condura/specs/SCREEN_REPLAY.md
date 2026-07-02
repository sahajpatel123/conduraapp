# SCREEN_REPLAY.md — Condura Replay · Screen Architecture

> **The contract.** This spec defines the layout, content slots, state matrix,
> motion, keyboard, component composition, data, and design decisions for the
> **Condura Replay** surface — the 24-hour scrubbable action replay at
> `#/replay`. Phase 4 implements against this document; if a component
> disagrees, the component is wrong, not the spec.
>
> **Audience.** Phase-4 implementer. Designer for review.
>
> **Source-of-truth files** (read these alongside this spec):
> - `app/web/frontend/src/lib/condura/Replay.svelte` — current implementation (the baseline this spec replaces / extends).
> - `app/web/frontend/src/lib/stores/replay.svelte.ts` — store contract.
> - `app/web/frontend/src/lib/ipc/types.ts` lines 534–560 — `ReplayFrame`, `ReplayIntegrityReport`, `ReplayExportResult`.
> - `app/web/frontend/src/lib/ipc/client.ts` lines 401–409 — IPC method names.
> - `app/web/frontend/src/lib/condura/condura.css` — design tokens.
> - `app/web/frontend/src/lib/condura/APPFLOW.md` §4.6 + §7 — current Replay behaviour + state inventory.
> - `app/web/frontend/src/lib/condura/MOAT.md` — premium-quality rules.
> - `app/web/frontend/src/lib/condura/TEARDOWN.md` — onboarding / interaction patterns.
> - `app/web/frontend/src/lib/condura/Thread.svelte` + `Pulse.svelte` + `Glyph.svelte` — the signature primitives.

### 0.0 Contract caveat (read first)

The brief names five north-star docs: `DIRECTION.md`, `TEARDOWN.md`,
`DESIGNLANG.md`, `APPFLOW.md`, `MOAT.md`. As of this writing, only **three
exist on disk** in `app/web/frontend/src/lib/condura/`:
`APPFLOW.md`, `MOAT.md`, `TEARDOWN.md`. `DIRECTION.md` and `DESIGNLANG.md`
are referenced but absent. This spec is grounded in the three that exist,
in the live `Replay.svelte`, in the `replay` store, and in the IPC types.
The material below already reflects most of what those docs would likely
contain (Thread as signature, anti-patterns, motion grammar) because they
are encoded in the design tokens + `MOAT.md` + the live component.

### 0.1 What Replay IS and IS NOT

**IS:** the **transparency surface** (CLAUDE.md §18). The last 24 hours as
a scrubbable synapse thread. Every action the agent took — with screenshot,
decision, outcome, and a verifiable HMAC chain — laid out so the user can
**prove what happened** without leaving the app. This is the page the user
opens when they suspect the agent did something they didn't ask for; it is
designed to be the answer.

**IS NOT:**
- A video editor. There is no timeline with multiple tracks, no
  compositing, no keyframe animation. It is one scrubber over a sequence
  of captured frames.
- A log viewer in the Audit sense. Audit (§4.5 of APPFLOW) shows every
  HMAC-chained *event* (including reads, errors, info-level); Replay shows
  the **physical-actionable subset** with a screenshot taken before and
  after. They overlap but the framing is different: Audit is the ledger;
  Replay is the eyewitness record.
- A compliance product. There is no "approve" / "deny" workflow; nothing
  in Replay mutates the past. It is read-only.
- A demo surface. The screenshot viewer has no nice typography overlay,
  no callouts, no "explainer" — what the agent saw is what you see.

### 0.2 Naming reconciliation (the brief vs. the IPC)

The brief refers to RPC method names `replay.list`, `replay.screenshot`,
and `replay.export`. The actual JSON-RPC 2.0 surface today is
`replay.timeline`, `replay.verify_integrity`, and `replay.export`. The
phase-4 implementation must **add `replay.screenshot_at(idx|t)`** (the
current implementation embeds `before_screenshot` / `after_screenshot`
base64 directly in each frame; for a long session this is 50MB/day and
must move to a per-frame fetch). Until that lands, this spec assumes the
lazy screenshot path is not present and shows a **no-screenshots** state
when a frame has no inline base64.

For the rest of this document the brief's vocabulary wins, with the actual
method name in parentheses on first use.

---

## Table of Contents

1. [Spec vs. Implementation Drift](#1-spec-vs-implementation-drift)
2. [Layout & Content](#2-layout--content)
3. [State Matrix](#3-state-matrix)
4. [Motion Choreography](#4-motion-choreography)
5. [Keyboard](#5-keyboard)
6. [Components Used](#6-components-used)
7. [Data Fetched](#7-data-fetched)
8. [Design Decisions — MOAT compliance](#8-design-decisions--moat-compliance)
9. [Accessibility Contract](#9-accessibility-contract)
10. [What this spec deletes from the current `Replay.svelte`](#10-what-this-spec-deletes-from-the-current-replaysvelte)
11. [Test Plan](#11-test-plan)

---

## 1. Spec vs. Implementation Drift

What this spec changes in the current `Replay.svelte` (778 lines). Phase 4
must apply these in one atomic diff — no half-states.

| # | Today (current `Replay.svelte`) | Phase 4 (this spec) | Why |
|---|---|---|---|
| 1 | Playhead position is driven only by pointer drag (no play / pause). | **Add play / pause / 0.5× / 1× / 2× / 4× speed control**, wired to a real `requestAnimationFrame` loop over `selectedIndex`. | A 24-hour timeline without play is a slide-rule; with play it is the user's "show me what happened" answer. |
| 2 | `before_screenshot` + `after_screenshot` are rendered **side-by-side** (`.shots` grid). | **Single primary screenshot at the playhead** (the current frame's `before_screenshot` if present, else `after_screenshot`); zoom in/out + reset. The "Before → After" comparison moves to a small dropdown or `B / A` toggle. | Two side-by-side screenshots force the viewport to split — the user can't see the actual screen in detail. A single primary image, with comparison on demand, is honest UX. |
| 3 | Scrubber `track` is a `1px hairline + a 2px Thread fill + a pollen mote`. | **Track is the Thread.** The whole scrubber **IS** the brand-spine Thread (a horizontal `<Thread orientation="h" />`), with tick marks every 5 minutes drawn into the same thread band; the playhead is a pollen mote riding the leading edge. | The brief says "the signature Thread as the time axis." The current implementation only uses Thread as the post-event fill; this spec makes the whole timeline a Thread. |
| 4 | Integrity badge uses `glyph:check` / `glyph:shield` with `var(--ok)` / `var(--danger)` text fill. | **Same badge**, but the `ok` state has a subtle pollen halo on the pulse dot (the chain being "alive and intact" deserves a heartbeat). `bad` state gets a `var(--danger)` hairline drawing in below the badge (the same `err-hair-draw` recipe used elsewhere). | MOAT §3 — "thread for completion." A broken chain is the textbook case of "this is now (brokenly) finished." |
| 5 | Export is a "Export .mp4" pill that hard-codes the call `replay.exportMP4()`. | **Same surface, but the disabled state is explicit and the result row goes under the hero as a `var(--ok)` mono note;** for v0.1.0 an explicit note reads `Exports to .mp4 in v0.2.0. (Saves screenshots + decisions to ~/Documents/synaptic-backups/ today.)` | Honesty per CLAUDE.md §30; per the brief: "No fake polish on the export (if v0.2.0, say so honestly)." |
| 6 | The decision receipt (`<aside class="receipt">`) is right of the screenshots. | **The decision receipt moves to a sticky bottom-left strip** (per the brief), with the screenshots above it and the scrubber at the very bottom. | The brief positions the receipt below the screenshot; this matches the user's mental model of "what happened" → "where was it in the film." |
| 7 | Bottom right has no action; the only call-to-action is the top-right `Export .mp4` pill. | **Bottom right gains a "Jump to Audit →" `<a class="threadlink">`**, navigating to `#/audit` and pre-selecting the same frame's event via `audit.select_by_id(replay.selected.id)`. | Cross-surface threadlink. Both Audit and Replay index the same HMAC chain (different lens); the jump should land the user on the same row. |
| 8 | Reload calls `replay.refresh()` on mount only. | **Polling cadence: 5 s while mount-open and `playhead === 'live'`** (i.e., the user is at the latest frame). Stops polling when the user scrubs back. Resumes when the user clicks "Live →" (re-jump to the latest). | The agent can act right now; a stuck replay is the worst answer to "what just happened." |
| 9 | `binds` ARIA `aria-valuemin/max/now` on the scrubber but no `aria-valuetext` exception for empty / single-frame. | **Scrubber is `role="slider"` with full aria-value triplet, but the meta-row above ("0 / 0 · —") is a separate live region** so screen readers don't get spammed when the count grows by one. | A11y. |
| 10 | `prefers-reduced-motion` is not respected at the scrubber level (the transition lives on the .fill width + .mote left). | **Disabled — the `.fill` width snaps and the `.mote` snaps** (the global `condura.css` `prefers-reduced-motion` block already zeroes both; this spec explicitly does not re-declare the rule). | MOAT §2.3 — one block, no per-component override. |
| 11 | No zoom controls on the screenshot. | **Zoom in / out / reset** (clickable glyph trio, or scroll-wheel + `⌘+` / `⌘−` / `⌘0` keyboard). Magnification 1× to 4× in 0.25× steps. `prefers-reduced-motion` snaps. | The frames are full-resolution screen captures; without zoom you cannot read what the agent saw. |
| 12 | Receipt's `decisionLine` falls back to `I decided to ${action}.` when `message` is empty, in a non-italic mono context. | **Same fallback, but the line is set in Instrument Serif italic 22px** (matching `.receipt-line` today; just lifted from the receipt header for clarity). | MOAT §1 — restraint; the receipt is the one place a serif italic earns its space. |

---

## 2. LAYOUT & CONTENT

### 2.1 Page-level geometry

The Replay route renders inside the Shell's main surface, right of the
NavRail (240 px on ≥1440 px; 56 px on 768–1023 px), per APPFLOW.md §3.1.

```
┌────────────────────── Replay route (max-width 1280px, padded) ──────────────────────┐
│                                                                                     │
│  ── Action Replay                                                [Chain intact] [↥]│  hero (eyebrow + headline + lead + integrity badge + export)
│                                                                                     │
│  The last 24 hours, scrubbable.                                                    │
│  Every action the agent took, in order. Drag the thread to inspect                 │
│  what it saw, and what it decided.                                                 │
│                                                                                     │
│  ───────────────────  Thread (hairline rule)  ───────────────────                  │  divider
│                                                                                     │
│  ┌─────────────────────────────────────────┐  ┌──────────────┐                      │
│  │                                         │  │ RECEIPT      │  decision receipt    │
│  │   Screenshot viewer (single primary,    │  │ ──────       │  (action, app,       │
│  │   hairline border, paper bg, zoom)      │  │              │   actor, result,     │
│  │                                         │  │ "I decided   │   outcome, reason,  │
│  │   [B] [A]   ·   100%  [+] [-] [⊙]        │  │   to ..."     │   HMAC row id)      │
│  │                                         │  │              │                      │
│  │                                         │  │ App ·   Mail │                      │
│  │                                         │  │ Actor ·      │                      │
│  │                                         │  │  haiku       │                      │
│  │                                         │  │ Result · ok  │                      │
│  │                                         │  │ Outcome ·    │                      │
│  │                                         │  │  allowed     │                      │
│  └─────────────────────────────────────────┘  └──────────────┘                      │
│                                                                                     │
│  ◀──────  Thread (scrubber, full-width, 5-min ticks)  ──────────────────   ●  ─────▶│  scrubber
│  │  │  │  │  │  │  │  │  │  │  │  │  │  │  │  │  │  │  │  │  │                       │
│  24h ago                                                                            now
│                                                                                     │
│  08:14:22 · frame 12 / 142  · · · · · · · · · · · · · · · · · ← → step · drag scrub  │  meta row
│                                                                                     │
│                  [↥ Jump to Audit →]                          0.5×  1×  2×  4×  [⏵]   │  control row
│                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 Region map

| Region | Where | Component | Visible when |
|---|---|---|---|
| **(A) Hero** | top, full-width | inline `<header>` | always |
| **(B) Divider** | below hero | `<Thread orientation="h" />` | always |
| **(C) Screenshot viewer** | top-left, primary | `<ScreenshotViewer />` | `state ∈ {playing, paused, seeking, no-screenshots}` and `selected != null` and `selected.screenshot != null` |
| **(D) Decision receipt** | top-right, 320–360 px | `<DecisionList />` | same as C; **also visible in the no-screenshots state** so the user always knows what the frame is even if the pixels are absent |
| **(E) Scrubber** | full-width, bottom | `<TimelineScrubber />` | `count > 1` |
| **(F) Meta row** | below scrubber | inline `<div>` mono | `count > 0` |
| **(G) Control row** | bottom of viewport | inline `<div>` | `count > 0` |
| **(H) Error row** | above the divider, when `lastError` present | inline `<div>` (the standard `err-state` block) | `lastError != ''` |
| **(I) Empty state** | centered | inline `<div>` | `count === 0 && !loading && !lastError` |
| **(J) Loading state** | centered | inline `<div>` | `loading && count === 0` |

### 2.3 Per-region slot detail (exact copy + components)

#### (A) Hero

| # | Slot | Content |
|---|------|---------|
| 1 | Eyebrow | `-- Action Replay` (mono 11 px / `--ls-mono` / `--content-faint`). |
| 2 | Headline | `The last 24 hours, scrubbable.` (Instrument Serif `--text-h1` / `--ls-h1` / `--content`). |
| 3 | Lead | `Every action the agent took, in order. Drag the thread to inspect what it saw, and what it decided.` (Inter 15 px / `--content-soft` / max-width 56ch). |
| 4 | Integrity badge | The pill described in (D) of the existing implementation. Colors: `var(--ok)` border + 8 % fill when `valid`; `var(--danger)` when broken (with the `err-hair-draw` hairline below); `Pulse phase="thinking"` + label `Verifying…` while the verify RPC is in flight. Click → `replay.verifyIntegrity()`. |
| 5 | Export pill | `Export .mp4` (the existing pill). When `disabled`, the inline note in §1 #5 is shown under the hero as a faint mono line. When `exporting`, the pill becomes `Exporting…` + `Pulse phase="acting"` + a 50 %-opacity grain over the screenshot. When complete, an `--ok` mono line reads `Exported to /Users/.../replay-2026-07-02T14.32.mp4`. |

**Layout note.** Slots 4 and 5 are right-clustered (the user reads the
lead first; the actions live in the top-right corner of the hero strip,
right of the eyebrow/headline/lead block). On `<880 px` the hero stacks
vertically — the badges drop to their own row under the lead.

#### (B) Divider

A 140 px wide `<Thread orientation="h" draw={true} glow={false} />`
(matches the existing `.rule` style). Left-aligned, 12 px below the hero.

#### (C) Screenshot viewer

| # | Slot | Content |
|---|------|---------|
| 1 | Image frame | The current frame's **primary** screenshot. `selected.before_screenshot` if present, else `selected.after_screenshot`. Decoded from base64 → `Blob` → `URL.createObjectURL()` (the existing component inlines the data-URL; the spec keeps that path so we don't ship a new IPC just yet — see §0.2). |
| 2 | Image border | `1 px var(--hair)` on the default `<figure>`; rounds to `var(--r-md)`. The spec adds a `2 px var(--synapse)` border on focus (`.shot-img:focus-visible`) so the user can see what is focused after `Tab`-stepping in. |
| 3 | Zoom controls (top-right of the frame) | Three glyph buttons in a `<Tooltip label>` row: `[+] zoom in` (`glyph:zoom-in`), `[−] zoom out` (`glyph:zoom-out`), `[⊙] reset` (`glyph:refresh`). Toggles 0.25× increments from 1× to 4×. Default 1×. `prefers-reduced-motion` snaps with no transition. |
| 4 | Before / After toggle (bottom-left of the frame) | Two glyph-only chips: `B` and `A`. The active one is filled (`var(--surface-card)`). Default is `B` if `before_screenshot` is present, `A` otherwise. Keyboard: `B` / `A` keys when the image has focus (NOT when the scrubber is focused — this is image-local). |
| 5 | Image not available | A 220 px-tall dotted-hairline frame with the mono eyebrow `NO SCREENSHOT FOR THIS FRAME` and a sub line `Captures were skipped or expired. The decision record is below.` The receipt still renders to the right. |

**Why one primary, not side-by-side.** A 24h timeline can include a
click in Mail, a Terminal command, and a browser navigation in three
consecutive frames. The user wants to read **what the agent actually saw**
when it clicked "Send" — that demands the full viewport, not a 50/50
split. The Before/After toggle is the comparison tool (and the brief
explicitly calls out the two-shot view).

#### (D) Decision receipt

| # | Slot | Content |
|---|------|---------|
| 1 | Timestamp | mono 11 px / `--content-faint` / `Mmm DD · HH:MM:SS` |
| 2 | Decision line | Instrument Serif italic 22 px / `--ls-h2` / `--content` — `I decided to ${selected.message ?? "act"}.` (falls back to a constructed phrase when `message` is empty). |
| 3 | Property grid | A 6-row `<dl>` (Action, App, Actor, Result, Outcome, Level) in the same pattern as the existing receipt. Outcome cell uses `var(--ok)` for `allowed`, `var(--danger)` for `denied` / `errored`, plain otherwise. |
| 4 | Outcome reason | An italic small line under the grid (when `selected.outcome_reason` is set). Hairline-top border separates it from the grid; mono small; `--content-mute`. |
| 5 | HMAC row id | mono 11 px / `--content-faint` / `id: 12 · audit_id: 47a3…`. The `47a3…` is a threadlink: clicking jumps to `#/audit` and pre-selects that row. |

#### (E) Scrubber — the Thread-as-time-axis

This is the **signature element** of the surface, per the brief. It IS a
`Thread`, not a track with a Thread fill on it.

| # | Slot | Content |
|---|------|---------|
| 1 | Track | A 2 px tall horizontal `<Thread orientation="h" />` spanning the full content width. (The existing component uses a `1 px hairline + 2 px fill`; this spec replaces it with the spine-of-the-app Thread.) |
| 2 | Tick marks | Every **5 minutes**, a 4 px tall vertical hairline is drawn into the Thread band (`color: var(--content-faint)`, 1 px wide, transform-origin top). The first tick after 00:00 (when `count > 0` and frame 0 is at `t₀`) is the **frame origin** — a 6 px tall, `--synapse` tinted tick. |
| 3 | Playout region | The portion of the track **left of the playhead** is `var(--synapse)` (the brand green); **right of the playhead** it is `var(--hair-strong)` (the "ahead" hairline). The transition is at the playhead itself, not a separate fill — the Thread's gradient does this. |
| 4 | Playhead | A pollen mote (`24 px halo + 8 px solid pollen circle + `<Pulse phase="acting" size={6} />` inside the halo`) at `left = ${playheadPct}%`, vertical-centered on the track. The mote's halo opacity **rises** while `playing === true` (the synapse-glow pulse on the current playhead position) and sits still when `paused`. |
| 5 | Step dots | At every captured frame, a 2 px `--content-faint` dot is rendered on the Thread band (only those whose timestamp does not collide with a 5-min tick; otherwise the tick wins). |

**Implementation note.** The current implementation's `<Pulse
phase="acting" />` inside the `.mote` is the right call (it's a
breathing dot, not a static circle). The halo gradient already matches
MOAT §3 — a single hairline with one pollen accent. The spec adds the
rising-halo opacity so the playhead feels alive during playback.

#### (F) Meta row

A mono 11 px strip below the scrubber. Three segments, left-aligned with
margin-right separators:

- `08:14:22` (the current frame's `timestamp` formatted).
- `· frame 12 / 142` (1-indexed counter, `--content-faint`).
- right-aligned: `← → step   Shift+←/→ ±10s   Home/End jump   ⌘E export` (faint, 0.7 opacity).

#### (G) Control row

Bottom of the viewport. Left cluster: a `<a class="threadlink">Jump to Audit →</a>`
that calls `audit.selectById(replay.selected.id)` and navigates to `#/audit`.
Right cluster: a segmented `<div role="group">` of speed pills (`0.5× / 1×
/ 2× / 4×`; active one filled `var(--surface-card)` with `var(--synapse)`
text) + a single Play/Pause button (`glyph:play` / `glyph:pause`,
`var(--r-pill)`, `var(--pollen)` border, `var(--pollen)` text + halo on
hover). The Play button is the *only* affordance that uses pollen border
on the surface (it earns it — it is the "show me what happened" gesture).

#### (H) Error row

Standard `ErrorState` block (the extraction per MOAT §1.2). Italic
headline `We couldn't read the timeline.`; sub line `{lastError}`; retry
button (mono upper-case `--synapse` border pill) that calls
`replay.refresh()`.

#### (I) Empty state

The exact copy currently in production: a vertically-centered pair of
serif italic lines:

- Headline (Instrument Serif italic 24 px / `--ls-h2` / `--content`):
  `Nothing to replay yet.`
- Sub (Instrument Serif italic 15 px / `--content-faint` / max-width 48ch):
  `Once the agent acts, every decision lands here — screenshot, decision, outcome. The last 24 hours, scrubbable.`

Below them, **just the scrubber band** — a Thread hairline running the
full width with a **single quiet pulse** at the start (0 → 1 → 0, 4 s
loop, 1.4 s ease-in-out, opacity 0.4 → 0.7 → 0.4). This is the
"transparency surface waiting for the agent to act" gesture — not a
spinner, not a Pulse.

#### (J) Loading state

Same shape as Empty but with mono uppercase text `LOADING FRAMES…` (11 px
/ `--ls-mono` / `--content-mute`) replacing the headline, and a
`<Pulse phase="thinking" size={10} />` above the text. NO Thread draw on
the scrubber (the scrubber doesn't exist yet — `count === 0`). Per
MOAT §2.5 + §4 #7, no spinner.

### 2.4 What this surface does NOT show

- **No multiple tracks.** One scrubber, one screenshot, one decision.
- **No audio.** Replay is silent — there's no sound the agent makes.
- **No "agent was watching a video" marker.** If the user wants to know
  *why* the agent was in a window, the receipt's `Action` + `App` is the
  answer; we don't add a parenthetical note.
- **No annotations / redaction.** v0.1.0 shows the literal captured
  pixels. A future v0.2.0 PII-redaction toggle may live in Settings →
  Audit.

---

## 3. STATE MATRIX

This is the spec for what Replay shows in each reachable state. The
default for unknown errors is **honest degradation** — never a dead wall
(APPFLOW I4: "every state is reachable").

State predicates reference:

- `count = replay.frames.length`
- `idx = replay.selectedIndex`
- `selected = replay.frames[idx]`
- `loading = replay.loading`
- `exporting = replay.exporting`
- `integrity = replay.integrity`
- `lastError = replay.lastError`
- `play = replay.playState`  (`'idle' | 'playing' | 'paused'` — added by this spec)

### 3.1 Empty (cold start, no agent actions yet)

**Trigger:** `count === 0 && !loading && !lastError`

**Visual:**
- Hero + divider + empty block per §2.3 (I).
- Sub-line pulse: Thread at the bottom of the viewport, 0.4 → 0.7 →
  0.4 opacity, 4 s loop.
- No control row, no meta row, no scrubber (no frames to play).

**Copy:**
- Headline: `Nothing to replay yet.`
- Sub: `Once the agent acts, every decision lands here — screenshot, decision, outcome. The last 24 hours, scrubbable.`

### 3.2 Loading

**Trigger:** `loading && count === 0`

**Visual:**
- Hero + divider + loading block per §2.3 (J).
- Mono uppercase `LOADING FRAMES…` (this is the only place this exact
  string is used; per MOAT §4 #5 it reads as honest work, not "loading
  the experience…").
- Below the text, a Thread draws in from left (`stroke-dashoffset 1 → 0`,
  `--dur-cine`).

**Copy:** as above.

### 3.3 Playing

**Trigger:** `play === 'playing'`

**Visual:**
- Hero + receipt + screenshot + scrubber + meta + control row, all as in
  §2.3.
- The playhead's halo opacity ramps to **0.95** (from a paused 0.55)
  with a single `transition: opacity var(--dur) var(--ease)` over 280 ms.
- The Play/Pause button shows `glyph:pause` and reads `Pause`.
- The control row's speed pills update to reflect the active speed.
- `selectedIndex` advances by `speed × dt × frameRate` per rAF tick. At
  the end of the timeline (`idx === count - 1`), playback auto-pauses
  and the button flips back to `glyph:play` reading `Play`.

**Speed semantics:**
- `0.5×` — one frame per 2 s (slow; reads as "show me the decision
  step-by-step").
- `1×` — one frame per 1 s (default).
- `2×` — one frame per 500 ms.
- `4×` — one frame per 250 ms (the "scan the day" speed; at 4× a
  24 h × 142 frames timeline scrubs in ~35 s).

### 3.4 Paused

**Trigger:** `play === 'paused' || (play === 'idle' && count > 0)`

**Visual:**
- Identical to playing in layout, but the playhead halo rests at 0.55
  opacity (the `paused` state of `<Pulse phase="thinking" size={6} />`).
- The Play/Pause button shows `glyph:play` and reads `Play`.

### 3.5 Seeking

**Trigger:** `scrubbing === true` (the user is holding the playhead with
the pointer).

**Visual:**
- The playhead **snaps** to the cursor — `transition: left 120ms
  var(--ease)` for *all* fills (matches the brief), `var(--ease)` curves
  both ends; the mote is `transition: none` only while dragging
  (matching the existing `.scrubbing .mote { transition: none; }` rule).
- The screenshot cross-fades to the new frame over **160 ms** (the
  brief's "the screenshot cross-fades to the new frame over 160ms");
  fade out → fade in on the same image element, no remount.
- Playback state: if was `playing`, it auto-pauses (the user is manually
  scrubbing; do not fight them) and resumes when they release.
- Selection updates in real-time — `selectedIndex = closest-frame(idx
  from cursor)` on every `pointermove` event (throttled to 30 Hz).

### 3.6 Error

**Trigger:** `lastError != ''`

**Visual:**
- The hero is still rendered. Above the divider, the standard
  `ErrorState` block mounts:
  - Headline: `We couldn't read the timeline.` (Instrument Serif italic
    22 px).
  - Sub: `${lastError} Replay frames live on disk; the daemon may be
    offline or the file may be unreadable.`
  - Retry button (ghost pill): `Try again` → `replay.refresh()`.
  - A `--danger` hairline drawing in below the error block (the
    `err-hair-draw` 600 ms animation; matches the existing component).
- Below the error block, the **other regions still render** if `count
  > 0`. The error is local: the user can still scrub whatever frames
  loaded before the error.

### 3.7 Integrity broken (overlay on top of any other state)

**Trigger:** `integrity && integrity.valid === false`

**Visual:**
- The integrity badge flips to `Chain broken` + `glyph:shield` + red
  border (the existing pattern), and a `--danger` `err-hair-draw`
  hairline draws in below the badge.
- A single mono small line below the export pill reads exactly
  `Chain broken at row ${integrity.first_break_id ?? '—'}: ${integrity.first_break_reason}` (`--danger` text).
- This state **does not block the rest of the surface** — the user can
  still scrub, still export (which will write a "chain broken at row
  X" note in the export's first caption). The bad data is shown; the
  user decides what to do.

### 3.8 No-screenshots (the screenshot is missing for this frame)

**Trigger:** `selected != null && (selected.before_screenshot == null &&
selected.after_screenshot == null)`

**Visual:**
- The screenshot viewer renders the dotted-hairline empty figure from
  §2.3 (C) slot 5: 220 px-tall, mono eyebrow `NO SCREENSHOT FOR THIS FRAME`,
  italic sub `Captures were skipped or expired. The decision record is below.`
- The decision receipt still renders normally (the user can read the
  decision even without the pixels).
- The scrubber still functions; the user can move on to the next frame.

### 3.9 Static-error (any other state plus a non-fatal daemon issue)

**Trigger:** outside the four above (per APPFLOW §6.1)

**Visual:**
- Shell-level: `DynamicIsland phase="error"`.
- Route-level: any per-region error renders inline. The Thread in the
  titlebar keeps drawing — the surface is alive even when a frame load
  fails.

### 3.10 What this state matrix does NOT cover

- **Voice / consent / kill-switch.** Those are owned by the Shell and
  covered in `SCREEN_SHELL.md`. Replay doesn't have its own consent
  flow — exporting a `.mp4` is local-only.
- **Multi-frame selection.** The user scrubs one frame at a time. There
  is no "lasso three frames and export just those"; v0.2.0 may add.

---

## 4. MOTION CHOREOGRAPHY

The motion vocabulary uses the tokens in `condura.css`. Per MOAT §2.3
and §2.10, components never redeclare reduced-motion overrides; one
block in the global stylesheet (lines 469–476) does the whole work.

### 4.1 Mount sequence (route enter from `#/replay`)

| Beat | What | Duration | Easing |
|---|---|---|---|
| 0–320 ms | `.replay` opacity 0 → 1 + `translateY(8px) → 0` (the existing `blur-in` recipe) | 320 ms | `--ease` |
| 60–380 ms | Hero headline opacity 0 → 1 (no transform) | 320 ms | `--ease` |
| 100–420 ms | Divider `<Thread>` draws in left → right | 320 ms | `--ease` (this is one of the signatures — Thread drawing in is "data ready") |
| 180–800 ms | Screenshot + receipt fade in (`blur-in`) | 620 ms | `--ease` |
| 380–900 ms | Scrubber draws in (Thread stroke-dashoffset 1 → 0) | 520 ms | `--ease` |
| 540–820 ms | Meta + control rows fade in | 280 ms | `--ease` |

`prefers-reduced-motion`: all of the above collapse to opacity-only
transitions at `--dur-fast` (the global override already does this; this
spec does not re-declare the media query).

### 4.2 Scrubbing (the playhead)

| Beat | What | Duration | Easing |
|---|---|---|---|
| on `pointerdown` | `.mote` halo scale 1 → 1.4 (the focus-visible style lifted); `cursor: grabbing` | 120 ms | `--ease` |
| on `pointermove` | `.mote` left snaps to cursor; `.fill` width matches; selection updates in real-time; screenshot cross-fades | 160 ms | `--ease` |
| on `pointerup` | `.mote` halo returns to 1; `cursor: pointer`; playback resumes if it was playing | 120 ms | `--ease` |

Per the brief: "the playhead snaps to the cursor position with a 120 ms
ease, the screenshot cross-fades to the new frame over 160 ms." Both
numbers match.

### 4.3 Playing (the playhead moves over time)

| Beat | What | Duration | Easing |
|---|---|---|---|
| on `play` event | `.mote` halo opacity 0.55 → 0.95 | 280 ms | `--ease` |
| on every rAF tick | `.mote` left updates to `playheadPct` (no transition — the rAF is the animation); `selectedIndex` increments by `speed × dt` | continuous | — |
| on `pause` | `.mote` halo opacity 0.95 → 0.55 | 280 ms | `--ease` |
| on speed change | The 5-min ticks do NOT redraw (they're anchored to absolute timestamps) — only the mote moves | 120 ms | `--ease` |

The playhead's pulse (the synapse-glow halo) uses the existing
`<Pulse phase="acting" />` element, which already runs `breathe` at
`1s var(--ease) infinite` (the "quick" cadence). The spec changes
**nothing** about the pulse itself; what the spec adds is the halo's
opacity transition in/out of the playing state.

### 4.4 Export (a Thread draws across the bottom of the screenshot)

The brief says: "Export: a Thread draws across the bottom of the
screenshot viewer as the export progresses." This is the signature.

| Beat | What | Duration | Easing |
|---|---|---|---|
| on `click` of `Export .mp4` | A 2 px tall Thread mounts **inside** the screenshot viewer's bottom edge (below the zoom controls, above the Before/After toggle). `width: 0 → 100%` linearly over the export RPC duration. Color: `var(--synapse)`. | continuous | linear |
| on `exporter progress events` | The Thread grows with the export's reported byte-progress fraction (`0 → 1`). If the RPC doesn't expose progress, the Thread grows on a deterministic timer (estimate: 6 s for a 1-minute timeline, +2 s per additional minute). | tracks RPC | linear |
| on `done` | The Thread grows to 100 % and fades over 400 ms. The result line `Exported to /…/replay-…mp4` appears beneath the export pill (`--ok` color, mono). | 400 ms | `--ease` |

**v0.1.0 honest note.** Per the brief: "No fake polish on the export
(if v0.2.0, say so honestly)." When the `replay.export` RPC returns an
error (which it does for v0.1.0 if the underlying `.mp4` producer
isn't shipped), the Thread draws to ~80 % then **stops, fades to a
`--warn` color, and reverses** (the `--danger` `err-hair-draw` pattern
plays in reverse). The export pill flips to `Export .mp4` (cleared);
the inline note under the hero reads exactly:

> Exports to .mp4 in v0.2.0. Today: saves screenshots + decisions to `~/Documents/synaptic-backups/replay-<ISO-date>.jsonl`.

This is the honest answer, in the same `mono 11 px / 0.14em / uppercase` style the rest of the meta uses.

### 4.5 Integrity-bad transition

| Beat | What | Duration | Easing |
|---|---|---|---|
| on `valid === false` | A `--danger` 1 px hairline draws in below the integrity badge (the `err-hair-draw` keyframe, 600 ms) | 600 ms | `--ease`, delay 120 ms |
| on subsequent `verifyIntegrity()` re-run that succeeds | The hairline flips to `var(--ok)` and the badge updates; same keyframe plays | 600 ms | `--ease` |

### 4.6 Reduced-motion exception list

Under `prefers-reduced-motion: reduce`:

| Beat | What changes |
|---|---|
| Scrubbing | The `.fill` width and `.mote` left snap to the cursor position with no transition. The screenshot is replaced instantly (no cross-fade). |
| Playing | The pulse halo's opacity still tracks playing/paused, but the **breathing animation on the pulse inside the mote stops** (it's a global rule — `.mote` is just a wrapper around `<Pulse />`). The playhead still moves at the same rate (the user asked to play; don't take that away). |
| Pause / Play | The opacity transition is replaced with an instant flip. |
| Export | The Thread draw snaps directly to ~95 % (still tracks), then to 100 % on completion. No fade. |
| Integrity transition | Instant flip. |
| Mount | The Thread draws-in collapses to opacity-only at `--dur-fast`. |

These are **changes to the existing CSS transitions already in the file**;
the component does not register a per-component `prefers-reduced-motion`
block — the global block at `condura.css:469–476` covers the values and
the manual punctuation above is only for the user-visible ones.

---

## 5. KEYBOARD

Per MOAT §2.10 the keyboard story is a first-class surface. Routes own
route-local chords; the Shell owns global ones (per `SCREEN_SHELL.md`
§5.1, `g r` jumps to Replay; `⌘6` does the same with no go-to prefix).

### 5.1 Replay route-local chords (focus on scrubber or anywhere on the surface)

| Chord | Action | Notes |
|---|---|---|
| `Space` | Toggle play / pause. If at the end and playing, pause. If at the start (idx 0) and paused, start play. | Active when the scrubber has focus OR no focused input. |
| `←` | Step backward one frame. If at idx 0, hold. If playing, pause. | The existing component already handles this (line 94–108); this spec keeps the existing behavior. |
| `→` | Step forward one frame. If at idx `count - 1` and playing, pause (auto-stop at end). | Same. |
| `Shift+←` | Step backward 10 frames. If underflow, jump to idx 0. | New. |
| `Shift+→` | Step forward 10 frames. If overflow, jump to idx `count - 1`. | New. |
| `Home` | Jump to idx 0. | Existing. |
| `End` | Jump to idx `count - 1`. | Existing. |
| `1` / `2` / `3` / `4` | Set playback speed to 0.5× / 1× / 2× / 4×. | New. The number keys double as speed presets. |
| `0` | Jump to "live" (idx = count - 1; resume polling if at the end). | New. |
| `⌘E` / `Ctrl+E` | Trigger `replay.exportMP4()`. | Brief calls this out; matches the export pill's own click handler. |
| `B` | Show `before_screenshot` (if present). Toggle to `A` if already showing. | Image-local — only fires when the screenshot viewer has focus (and the Before/After toggle is visible). |
| `A` | Show `after_screenshot` (if present). | Same. |
| `Tab` / `Shift+Tab` | Walks the tab order (see §5.3). | Standard. |

### 5.2 When the screenshot has focus

In addition to the route-level chords above (which still fire):

| Chord | Action |
|---|---|
| `=` / `+` | Zoom in 0.25× (capped at 4×). |
| `−` | Zoom out 0.25× (floor 1× — zoom never goes below 100 %). |
| `0` | Reset zoom to 1×. (NB: this conflicts with the route-level `0` jump-to-live. The route-level handler wins when the scrubber has focus; this handler wins when the image has focus. The image's focus ring is a 2 px `var(--synapse)` border on focus, which makes the focus target clear.) |
| `↑` / `↓` / `←` / `→` | Pan the zoomed image by 8 % of the viewport per press. |

### 5.3 Tab order

Top-down, route-local:

1. Integrity badge (top-right).
2. Export `.mp4` pill.
3. Screenshot viewer (the image; focus shows the 2 px synapse border).
4. Zoom + / − / reset trio.
5. Before (`B`) / After (`A`) toggle.
6. Decision receipt's HMAC threadlink `id: 12 · audit_id: 47a3…`.
7. Scrubber (the `role="slider"`).
8. Speed seg (0.5× / 1× / 2× / 4×).
9. Play / Pause button.
10. "Jump to Audit →" threadlink.

`Shift+Tab` wraps to (1).

### 5.4 What this route does NOT handle

- **Anything the Shell owns.** Global chords (`⌘K`, `⌘,`, `g r`,
  `⌘6`, `Esc` for top-most overlay, etc.) are the Shell's. This
  surface does not re-listen for them.
- **Voice.** Replay has no STT hook.
- **Window-level shortcuts the OS owns** (Cmd+Q, Cmd+W on macOS). The
  Wails / Tauri substrate owns those.

---

## 6. COMPONENTS USED

Every Condura component this surface composes. Component names match
files under `app/web/frontend/src/lib/condura/`.

| Component | Where used | Role in Replay |
|-----------|------------|----------------|
| `Thread.svelte` | the divider (B); the scrubber's track (E); the empty-state thread (I); the export progress (G.4); the integrity hairline (D) | The brand-spine Thread — used here five times, every one of which is load-bearing (the divider = "data ready"; the scrubber = "this is the timeline itself"; the empty thread = "the agent will land here"; the export progress = "this is now being saved"; the broken-chain hairline = "this is now (brokenly) finished"). |
| `Pulse.svelte` | the integrity badge's `Verifying…` state; the playhead mote (`phase="thinking"` when paused, `phase="acting"` when playing) | The single dot of "alive." |
| `Glyph.svelte` | the integrity badge (`check` / `shield`), the export pill (`replay`), the speed seg (no glyphs; text only), the play/pause button (`play` / `pause`), the zoom trio (`+`, `−`, `⊙`), the Before/After (`B` / `A`), the Jump to Audit arrow | The icon set. |
| `Button.svelte` | the export pill; the integrity badge; the retry button on error; the speed seg; the play/pause button (variant `secondary`) | The tactile primitives. |
| `Tooltip.svelte` (planned per MOAT §2.9 — for v0.1.0 this is the inline `<Tooltip label>` primitive in `condura.css`) | every icon-only glyph (zoom, before/after, play/pause) | The keyboard- and mouse-readable label. No `title=` attributes anywhere on this surface. |
| `ErrorState.svelte` (planned per MOAT §1.2) | the error row (H) | The single source of error rendering on the surface. The current component re-declares this; the spec deletes the copy-paste and imports the component. |

### Components this surface does NOT use

- `Cursor.svelte` — owned by the Shell; Replay does not mount it. The
  whole-shell cursor is enough.
- `KillSwitchOverlay.svelte`, `ConsentModal.svelte` —
  Shell-level overlays. Replay is read-only.
- `CommandPalette.svelte`, `QuickPromptOverlay.svelte` — Shell-level.
- `Placeholder.svelte` — resigned (no route is unimplemented; per
  SCREEN_SHELL §6).

### New components this spec adds (Phase 4)

| Component | Purpose | Lines (estimate) |
|-----------|---------|------------------|
| `ScreenshotViewer.svelte` | Wraps the current `.shot-img` block. Props: `{ frame, zoom, onZoomChange, beforeOrAfter, onToggle }`. Owns the B/A toggle and the zoom UI. | 80–100. |
| `DecisionList.svelte` | Wraps the current `.receipt` block. Props: `{ frame }`. Owns the timestamp, the italic decision line, the property grid, and the HMAC threadlink. | 90–120. |
| `TimelineScrubber.svelte` | Wraps the current `.scrubber-wrap` + `.scrubber` blocks. Props: `{ frames, selectedIndex, play, onSeek, onPlayPause }`. Owns the Thread-as-time-axis, the 5-min ticks, the playhead mote, and the keyboard handlers in §5.1. | 140–180. |

The route component (`Replay.svelte`) becomes a slim coordinator that
wires the three new components to the `replay` store and owns the route
header + export pill + control row.

---

## 7. DATA FETCHED

The surface reads from the `replay` store and fires a small number of
mount-time calls. The IPC contract is JSON-RPC 2.0 over Unix socket (per
`SCREEN_SHELL.md` §7).

### 7.1 IPC calls (JSON-RPC 2.0)

| Brief name | Actual method | Returns | Source |
|---|---|---|---|
| `replay.list` | `replay.timeline` | `ReplayFrame[]` | `client.ts:401–403` |
| `replay.screenshot` | not yet implemented; would return `{ mime, base64 }` for a given `idx` | (planned v0.2.0) | — |
| `replay.export` | `replay.export` (`{destination?}`) | `{ path: string }` (v0.2.0) OR `{ error: string }` (v0.1.0) | `client.ts:407–409` |
| `replay.verify_integrity` | `replay.verify_integrity` | `ReplayIntegrityReport` | `client.ts:404–406` |

### 7.2 Initial calls on mount

| Call | When | Purpose |
|---|---|---|
| `replay.refresh()` | on mount | Fills `replay.frames` from `replay.timeline`. |
| `replay.verifyIntegrity()` | on mount, on user click of the integrity badge | Fills `replay.integrity` (HMAC chain walk). |
| `replay.exportMP4()` | on click of the export pill or `⌘E` | Returns `{path}` (v0.2.0) or throws (v0.1.0). |

### 7.3 Polling cadence

| Polling | Cadence | When |
|---|---|---|
| `replay.refresh()` | every **5 s** | WHILE `selectedIndex === count - 1` (the user is at "live"). |
| `replay.refresh()` | on user `0` (Jump to Live) | user-driven. |
| `replay.verifyIntegrity()` | on mount + on user click + on `0` jump | user-driven. |

When the user is scrubbing back through the timeline, no polling fires
(would override their position).

### 7.4 Cross-store reads / writes

| Store | Direction | What |
|---|---|---|
| `audit` | READ from Replay's "Jump to Audit →" click | The audit events have an `id`; Replay's `selected` has an `id`; on click we call `audit.selectById(replay.selected.id)` and `navigate('audit')`. |
| `replay` ↔ `audit` | snapshot at mount | These are two lenses on the same HMAC chain. They do not stay in sync during a session — if the user reloads `audit` while sitting on a frame in Replay, the Jump to Audit handler still finds the row by `id`. |

### 7.5 What the surface does NOT fetch

- **No provider / model list.** Replay does not concern itself with
  which LLM ran the action; the receipt's `actor` is opaque.
- **No memory / skills data.** Not relevant to "what the agent did."
- **No settings.** Theme, autonomy matrix, etc. — route-level concerns,
  not Replay.

---

## 8. DESIGN DECISIONS — MOAT compliance

Every surface in Condura must earn the MOAT bar. This section enumerates
the rules the design passes.

### 8.1 Premium tests passed

| Test | How Replay passes it |
|---|---|
| **The Restraint Test** (MOAT §1) | The scrubber IS one element, used 5 times — all five load-bearing (the spec lists them in §6). No decorative pulse loops. The empty-state pulse is a quiet breath, not an animation that "looks alive." No `rotateX` 3D flex on any receipt or screenshot. |
| **The Detail Test · focus rings** (MOAT §2.1) | The existing button focus rules in `Button.svelte:75–80` apply. The screenshot viewer adds a 2 px `--synapse` border on `:focus-visible` (track-rounded). The integrity badge's `glyph:check` uses a hex colored dot — no rectangular outline. |
| **The Detail Test · press states** (MOAT §2.2) | The export pill + Play/Pause + speed seg + before/after toggle all inherit the global `.tactile` class via `Button.svelte`. No per-component scale override. |
| **The Detail Test · reduced motion** (MOAT §2.3) | The global `condura.css:469–476` block does the whole work. Replay declares no media query; the spec explicitly does NOT add one (§4.6). |
| **The Detail Test · empty states** (MOAT §2.4) | The Empty state teaches (the sub line explains what will land here). The Loading state uses uppercase mono + the Thread-draw recipe (per §4.1). The No-screenshots state explains why the pixels are absent. The Error state explains the noun ("timeline") + the cause ("daemon offline / file unreadable") + the action (Retry). |
| **The Detail Test · loading states** (MOAT §2.5) | The Loading state draws a Thread (per the brief). No `<Spinner />` import. |
| **The Detail Test · error states** (MOAT §2.6) | The Error row follows the three-line pattern: noun, cause, action. The retry button is a ghost pill. The `--danger` `err-hair-draw` hairline plays once. |
| **The Detail Test · tactile vocabulary** (MOAT §2.7) | One transition list (the global `.tactile` rule) covers every press on the surface. Per-component overrides are forbidden by this spec. |
| **The Detail Test · overlay taxonomy** (MOAT §2.8) | Replay has no overlays — the screenshots live inline. The Jump to Audit threadlink is just a navigation, not a popover. |
| **The Detail Test · tooltip vs popover** (MOAT §2.9) | Icon-only buttons use the `<Tooltip label>` wrapper; no `title=` attributes anywhere. The Scrubber uses `title=` only as a fallback when the tooltip wrapper can't mount (browser without JS). |
| **The Detail Test · keyboard story** (MOAT §2.10) | §5 of this spec lists every chord. The 1/2/3/4 speed chords are a first-class part of the surface. |
| **The Signature — the Thread** (MOAT §3) | Replay is **the home of "Thread as time axis."** The scrubber is the longest, most readable Thread on the surface. Thread is used 5×, every one load-bearing. |
| **The $50M Feel · cursor catches hoverable** (MOAT §5.1) | The image border, all buttons, the scrubber, the toggles, and the threadlinks are all `data-hover="1"` aware via the global `use:hover-region` action (when wired). The Cursor is owned by the Shell and works here. |
| **The $50M Feel · composer thread** (MOAT §5.2) | N/A — Replay has no text input. |
| **The $50M Feel · Stop telegraphs yielding** (MOAT §5.3) | Replay's play→pause "yielding" is the playhead mote's halo opacity dropping 0.95 → 0.55 (with the `<Pulse>` reverting from `phase="acting"` to `phase="thinking"`). |
| **The $50M Feel · Settings nav updates without route enter** (MOAT §5.4) | N/A — Settings is a different route. |
| **The $50M Feel · mobile ritual** (MOAT §5.5) | Below 880 px the regions stack: hero, divider, screenshot (full-width), receipt (under it), scrubber, meta + control row. The receipt's italic line is still readable; the controls hit 44×44. |

### 8.2 Anti-patterns avoided (per MOAT §4)

| # | Anti-pattern | How Replay avoids it |
|---|--------------|------------------------|
| 1 | No gradient text | The headline and decision line are solid `--content`. |
| 2 | No emoji as UI icons | Every glyph is `<Glyph name="…" />`. The before/after "B / A" toggle is letter buttons (mono caps 12 px), not emoji. |
| 3 | No glassmorphism unless earned | The screenshot viewer is paper-on-paper (no `backdrop-filter`). The receipt card is `var(--surface-card)` with one `--shadow-paper`. The hover on zoom is `--shadow-card`, not stacked. |
| 4 | No rainbow accents | Status colors are `--ok` / `--warn` / `--danger` / `--info`. The playhead is `--pollen` (the brand); the play button border is `--pollen`. The Thread is `--synapse`. No purple/cyan/teal. |
| 5 | No "Welcome to the future" copy | Hero copy reads `The last 24 hours, scrubbable.` Empty state reads `Nothing to replay yet.` No exclamation points. |
| 6 | No fake enthusiasm | No "Got it!" / "Saved!" toasts on export. The export result is a `--ok` mono line that reads `Exported to /…`. On failure (v0.1.0), the mono line is the honest `Exports to .mp4 in v0.2.0…` note. |
| 7 | No spinner loaders | No import of any spinner. The loading state is the Thread-draw recipe per §4.1. |
| 8 | No rectangular focus outlines | Buttons and toggles inherit the global `--shadow-focus` (rounded halo). The image is `2 px var(--synapse)` on focus (rounded `var(--r-md)`). |
| 9 | No double shadows | Cards / receipt / viewer each have one shadow token. The hover on zoom lifts `--shadow-card`, not a layered custom shadow. |
| 10 | No animation that doesn't carry meaning | Every animation in §4 carries a meaning: divider = "data ready"; scrubber draw = "timeline arrived"; halo opacity = "playback state"; export Thread = "now saving"; broken-chain hairline = "this is now (brokenly) finished"; mount = "you just landed here." No decorative loops. |

### 8.3 What Replay is uniquely responsible for

Per CLAUDE.md §18 + §10.5 (Audit) and per the design:

- **Being the transparency surface.** Every action the agent took, with
  its screenshot, decision, outcome, and the HMAC row id. The user can
  prove what happened. If a user emails the team saying "your agent
  clicked Send on an email I didn't authorize", the user can open
  Replay, scrub to that moment, see the screenshot of the mail, and
  see the receipt's `actor` + `action` + `outcome`. This is the page
  we hand them.
- **Making the Thread the time axis.** The Thread is the brand spine
  (`SCREEN_SHELL.md` §8.3 "Being the home of the Thread"). Replay is
  where the Thread-as-time-axis earns its highest visible real estate.
- **Verifying the HMAC chain.** The integrity badge is the only
  surface-level action that walks the HMAC chain (apart from Settings
  → Legal). If the chain is broken, the surface shows it without
  hiding it.
- **The cross-route threadlink.** Replay ↔ Audit is the most important
  navigation in the app. Both surfaces index the same chain with
  different lenses; jumping between them on the same row keeps the user
  oriented.

### 8.4 What Replay explicitly does NOT do

- **Auto-export.** The user clicks; we don't push.
- **Edit the past.** Replay is read-only. There is no "mark this frame
  as reviewed" or "delete this screenshot".
- **Annotate.** No drawing tools, no callouts, no PII redaction
  (deferred to v0.2.0 per Settings scope).
- **Drive consent.** There is no destructive action here. Export is a
  local file write — the user already has filesystem rights; no
  Gatekeeper ticket is minted.
- **Cancel the agent.** Replay never touches `conversation.cancel()`.

### 8.5 The transparency contract — three lines

When the user is on this surface, they must be able to answer **three
questions** without being told:

1. **What did the agent do?** (receipt: action + app + outcome).
2. **What did the agent see?** (screenshot viewer at this moment).
3. **Can I trust this record?** (integrity badge + HMAC row id threadlink to Audit).

Replay earns the MOAT bar when, viewing any frame of any state listed
in §3, a user can answer those three.

---

## 9. ACCESSIBILITY CONTRACT

| Concern | Approach |
|---|---|
| **Screen reader structure** | Hero is `<header>`; screenshot + receipt are wrapped in `<main>`; the scrubber is `<div role="slider" aria-valuemin={0} aria-valuemax={count-1} aria-valuenow={idx} aria-valuetext={fmtTs(selected.timestamp)}>`; the meta row is a `<div aria-live="polite">`. The integrity badge is a `<button>`. The export pill is a `<button>`. The speed seg + before/after + play/pause are `<button role="button" aria-pressed={…}>` or `<input type="radio">` (radio group is the right primitive for the speed seg). |
| **Live regions** | The meta row announces "frame N of total, timestamp." ONLY when the count grows by 1 (a new frame landed). The verifier result is NOT a live region — it has its own banner when it breaks. |
| **Keyboard reachability** | Every interactive element is in the tab order (§5.3). The scrubber is reachable; the image is reachable; the zoom trio is reachable. |
| **Color-blind safety** | The `outcome` text color (`--ok` / `--danger`) is paired with the eyebrow text and the inline reason — color is never the only signal. |
| **Reduced motion** | §4.6 enumerates every change. The user's Motion Strength slider in Settings also gates the playhead pulse (`data-energy="low"` hides the halo per `SCREEN_SHELL.md` §4.7). |
| **Touch targets** | ≥ 44×44 on every button on mobile (the existing component already enforces this). |

---

## 10. What this spec deletes from the current `Replay.svelte`

These are the changes the Phase-4 implementer must apply (the inverse
of §1's drift table). The file goes from 778 lines to roughly 280 lines
plus 80–120 line new components.

| # | Deletion | Why |
|---|---|---|
| 1 | The `.shots` grid with `figure > img + figcaption` (lines 205–228). | Replaced by `ScreenshotViewer.svelte` (single primary + B/A toggle). |
| 2 | The inline `.shot-empty` placeholder. | Replaced by the dotted-hairline `NO SCREENSHOT FOR THIS FRAME` in `ScreenshotViewer.svelte`. |
| 3 | The `<aside class="receipt">` block (lines 231–252) in the route component. | Moved to `DecisionList.svelte`. |
| 4 | The `.scrubber-wrap` + `.scrubber` blocks (lines 256–289). | Moved to `TimelineScrubber.svelte`. |
| 5 | The `binding "this"` on the scrubber (`.scrubberEl`). | Moved into `TimelineScrubber.svelte`. |
| 6 | The `.state-empty` block (lines 348–381) being split between Loading and Empty states. | Replaced by the dedicated §2.3 (I) and (J) blocks; the spec keeps them in the route header but the bottom Thread is in `TimelineScrubber.svelte`. |
| 7 | The hard-coded "Exported to /path" line (line 188). | Replaced by the bottom Thread + the inline note under the hero. |
| 8 | The `.export-btn` pollen-outline style (lines 721–761). | Replaced by `Button.svelte variant="secondary"` + a pollen border accent (one CSS line). |

### 10.1 What stays in `Replay.svelte`

The hero block (header + eyebrow + headline + lead + integrity badge +
export pill), the divider, the error row, the empty/loading block, the
control row. Roughly 280 lines.

---

## 11. Test Plan

### 11.1 Unit (vitest)

The spec implies the following unit tests (each test file is a Svelte
file or TS file; `@testing-library/svelte` for components, plain `vitest`
for the store):

| Test file | Tests |
|---|---|
| `TimelineScrubber.test.ts` | `5_min_ticks_draw_at_correct_pct`, `playhead_snaps_to_cursor_on_pointerdown`, `screenshot_cross_fades_during_scrub`, `home_jumps_to_idx_0`, `end_jumps_to_count_minus_1`, `arrow_keys_step`, `shift_arrow_keys_step_10`, `space_toggles_play`, `speed_pills_update_selected_index_per_rAF`. |
| `ScreenshotViewer.test.ts` | `shows_before_screenshot_by_default`, `B_toggle_switches_to_after_screenshot`, `no_screenshot_renders_dotted_placeholder`, `zoom_in_clamps_at_4x`, `zoom_out_floors_at_1x`, `reset_returns_to_1x`. |
| `DecisionList.test.ts` | `renders_italic_decision_line_from_message`, `falls_back_to_action_when_message_empty`, `outcome_allowed_uses_ok_color`, `outcome_denied_or_errored_uses_danger_color`, `hmac_threadlink_contains_audit_id`. |
| `replay.store.test.ts` | `refresh_sets_frames`, `refresh_sets_lastError_on_failure`, `selectIndex_clamps`, `verifyIntegrity_does_not_clear_frames`, `exportMP4_returns_path_or_throws`. |

### 11.2 E2E

Two flows:

| Flow | Steps | Expected outcome |
|---|---|---|
| **Empty → one-frame → playback** | `replay.refresh()` returns `[]` → mount with empty state → mock a frame → refresh → assert scrubber renders → click play → wait 1s → assert selectedIndex advanced by ~1. | Scrubber visible; playhead moves; meta row updates. |
| **Error path** | `replay.refresh()` rejects → assert Error row appears → click retry → assert frames load (success path). | Error block visible; retry works. |

### 11.3 A11y smoke

| Test | Tool |
|---|---|
| `role="slider"` has valid `aria-value*` triplet. | `axe-core` via `vitest-axe`. |
| Tab order matches §5.3. | `@testing-library/svelte` keyboard simulation. |
| `prefers-reduced-motion: reduce` collapses every transition to instant. | `matchMedia` mock + `@testing-library/svelte`. |

---

**This document is the architecture. The code is the implementation. They
agree. When they diverge, the divergence is the spec-bug — fix the doc,
then fix the code, in one commit.** (APPFLOW.md closing note, applies
here too.)
