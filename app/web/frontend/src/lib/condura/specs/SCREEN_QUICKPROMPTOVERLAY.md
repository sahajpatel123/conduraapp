# SCREEN_QUICKPROMPTOVERLAY.md — Condura Quick Prompt Overlay · Screen Architecture

> **The contract.** This spec defines the layout, content slots, state matrix,
> motion choreography, keyboard model, component composition, data contracts,
> and design decisions for the **Condura Quick Prompt Overlay** — the global
> hotkey overlay that floats above whatever the user is doing and turns a
> one-line thought into a delegated agent run without ever leaving the active
> app. Phase 2 will implement against this document; if a component disagrees
> with the spec, the component is wrong, not the spec.
>
> **Audience.** Phase-2 implementer. Designer for review.
>
> **Source-of-truth files** (read these alongside this spec):
> - `app/web/frontend/src/lib/condura/QuickPromptOverlay.svelte` — current implementation (the seed).
> - `app/web/frontend/src/lib/condura/condura.css` — design tokens.
> - `app/web/frontend/src/lib/condura/APPFLOW.md` §3.5 — quick-prompt summary.
> - `app/web/frontend/src/lib/condura/MOAT.md` — premium-quality rules.
> - `app/web/frontend/src/lib/condura/CLAUDE.md` §19.2 — overlay contract.

---

## 0.0 Contract caveat (read first)

The brief names five north-star docs: `DIRECTION.md`, `TEARDOWN.md`,
`DESIGNLANG.md`, `APPFLOW.md`, `MOAT.md`. As of this writing, only **three
exist on disk** in `app/web/frontend/src/lib/condura/`:
`APPFLOW.md`, `MOAT.md`, `TEARDOWN.md`. `DIRECTION.md` and `DESIGNLANG.md`
are referenced but absent. This spec is grounded in the three that exist
plus the live `QuickPromptOverlay.svelte` (which already implements a
working top-of-card version of this surface — the seed we refine).

If `DIRECTION.md` / `DESIGNLANG.md` are later produced, re-read this spec
for conflicts. The material below already reflects what those docs would
likely contain (thread as signature, motion grammar, anti-patterns)
because those rules are encoded in the design tokens + `MOAT.md` +
`TEARDOWN.md`.

---

## 0.1 What the Quick Prompt Overlay IS and IS NOT

**IS:** the **glance surface** of Condura. A floating paper card, ~480px
wide, summoned by the user's chosen global hotkey (default Option+Option,
no silent default per locked decision #8). It appears at the cursor or
centered on the active window, holds a **single-line composer** (one
focused question at a time, no scroll-back), and routes the thought to
the agent. The DynamicIsland status pill appears below the input while
the agent is responding. A "Open full Chat →" threadlink at the bottom
is the explicit bridge to deep work.

**IS NOT:** a full Chat surface. No sidebar, no message history, no
conversation list. The overlay is a **scratchpad, not a room**. After 5
seconds of inactivity it dismisses (it is for quick asks, not deep
work). It is not a modal — it is a sheet that lives above the active
app but does not block page scroll, does not trap focus away from the
host app on Esc, and does not own a route.

---

## 1. LAYOUT GRID

### 1.1 Region map (annotated)

```
            ┌───────────────────────────────────────────────┐
            │  (1) TopEdge Thread · 1px hairline · spans   │
            │      the card's top edge between --space-4   │
            │      gutters · draws in left→right (520ms)   │
            │      on summon                              │
            ├───────────────────────────────────────────────┤
            │  (2) Header row · horizontal · single row    │
            │  ┌────────┬──────────────┬────────┬─────────┐ │
            │  │ Pulse  │ "⌥⌥ Wake"   │ model  │  ✕      │ │
            │  │ 8px    │ mono chip    │ select │  close  │ │
            │  └────────┴──────────────┴────────┴─────────┘ │
            ├───────────────────────────────────────────────┤
            │  (3) Composer body · the writing surface     │
            │  • idle: <textarea> auto-grows (1→6 lines)  │
            │  • streaming: <YieldBar> replaces the        │
            │    textarea in the same slot (height         │
            │    collapse, 240ms ease)                      │
            │  • error: <ErrBlock> replaces the textarea   │
            ├───────────────────────────────────────────────┤
            │  (4) DynamicIsland mini-pill · ≤52px tall   │
            │      renders only in submitting/responding/ │
            │      done/error states; slides in below the │
            │      body in 240ms                            │
            ├───────────────────────────────────────────────┤
            │  (5) Footer · mic (left) · Send/Stop (right) │
            │  ┌──────┐                              ┌─────┐│
            │  │ mic  │                              │Send ││
            │  │ 32px │                              │ ↗   ││
            │  └──────┘                              └─────┘│
            ├───────────────────────────────────────────────┤
            │  (6) Hints row · "⌘↵ to send · Esc to       │
            │      close · idle dismiss in 5s" · mono 10px │
            │      hidden mid-stream                        │
            ├───────────────────────────────────────────────┤
            │  (7) "Open full Chat →" threadlink ·         │
            │      centered, mono, only renders in the    │
            │      done/error states (the bridge to deep   │
            │      work)                                    │
            └───────────────────────────────────────────────┘

            ┌───────────────────────────────────────────────┐
            │              ::before pollen halo              │
            │        (24px warm shadow under the card)       │
            └───────────────────────────────────────────────┘

   Anchor:   cursor position (preferred) → centered on active window
             (fallback). Anchor rule: if cursor is within 96px of any
             screen edge OR more than 720px from the top, recenter.
```

### 1.2 Width / position spec

```css
.qp-card {
  position: fixed;
  /* anchor: default cursor position via JS (state.cursorX/cursorY);
     if anchor would clip the viewport, snap to a 24px safe-area inset. */
  top:  var(--qp-top, 96px);   /* computed anchor, falls back to 96px */
  left: var(--qp-left, 50%);   /* computed anchor, falls back to center */
  transform: translate(-50%, 0); /* center horizontally when left=50% */
  width: min(30rem, 92vw);       /* 480px target, never exceed 92vw */
  max-height: min(80vh, 720px);
  z-index: var(--z-overlay);     /* above sheets, below modal consent */
  background: var(--surface);
  border: 1px solid var(--hair);
  border-radius: var(--r-lg);
  box-shadow: var(--shadow-float);
  overflow: hidden;              /* the top-edge thread clips cleanly */
  color: var(--content);
}

/* the warm pollen halo underneath — a quiet, ambient warmth */
.qp-card::before {
  content: '';
  position: absolute;
  inset: -8px;
  z-index: -1;
  border-radius: calc(var(--r-lg) + 4px);
  box-shadow: 0 0 24px color-mix(in oklab, var(--pollen) 18%, transparent);
  pointer-events: none;
}
```

**Width rationale.** 30rem (~480px) holds a single full-width textarea +
the 32px mic + a Send button that comfortably contains "Send ↗" without
wrapping. Below 480px (i.e., narrow windows or split-screen), `min(30rem,
92vw)` guarantees the card fits with 4vw left/right gutters. The card
never grows past `min(80vh, 720px)` tall — past that, content scrolls
internally (relevant only in done/error states with a long pill).

**Z-index rationale.** Above the Shell (`--z-shell`), above any sheets
(`--z-sheet`), below the consent modal (`--z-modal`), below the kill
switch (`--z-halt`). The cursor stays usable in whatever app the user
was in.

### 1.3 Padding / spacing

The card's interior rhythm matches `condura.css` tokens:

```css
.qp-body {
  padding: var(--space-5) var(--space-5) var(--space-4);  /* 20 / 20 / 16 */
  display: flex;
  flex-direction: column;
  gap: var(--space-3);                                     /* 12 between sections */
}
```

The single gap is fluid: when the `DynamicIsland` pill is absent
(idle/typing states), the body's `gap` collapses cleanly with the
streaming class — see §3.2.

---

## 2. CONTENT SLOTS

### 2.1 Top-edge thread (the "overlay arriving")

- **What:** a 1px hairline in `--synapse` that spans the inside of the
  card's top edge between `--space-4` gutters.
- **Component:** `<Thread orientation="h" draw={true}>` (the same component
  used in `Shell.svelte` and `Chat.svelte`).
- **Motion:** draws in left→right over `--dur-slow` (520ms, cubic-bezier
  ease) **the moment the overlay is summoned**, not when content paints.
  Honors `prefers-reduced-motion` (no draw, just appears).
- **Semantic:** the thread is the signature gesture ("a connection being
  made") — its draw in = "the overlay arrived". One thread per surface,
  always. This is the `MOAT.md §3` commitment.

### 2.2 Header row

A 32px-tall row containing four items in this order, left-aligned:

| Slot     | Content                                                  | Element                  |
|----------|----------------------------------------------------------|--------------------------|
| Heartbeat | `<Pulse phase={idle/acting/error} size={8}>`            | existing `Pulse.svelte`  |
| Wake chip | `"{hotkey} Wake"` in mono-uppercase, `--synapse` color  | `.qp-chip.qp-wake`       |
| Model chip | `<select>` with `provider · model` rows OR a "no model" `qp-mute` chip when no providers configured | `.qp-chip.qp-model` |
| Close button | `<Glyph name="close" size={14}>` 32×32 pill, right-aligned | `<button class="qp-close">` |

The chip trio (Pulse + Wake + Model) tells the user **what they will be
talking to** before they write a single character. It is the only
contextual chrome; everything below it is the prompt itself.

### 2.3 Composer body

A single slot. Exactly one of three things renders in it at any moment:

```
state.phase          → element rendered
─────────────────────────────────────────────
idle/typing          → <textarea>  (auto-grows 1 → 6 lines)
submitting/responding→ <YieldBar>  (40px height, "Condura is writing…" + Stop)
error                → <ErrBlock>  (italic, 1px danger hairline that draws in)
```

**Textarea (idle/typing).** Plain `font-sans 16px`, transparent
background, no border, no padding. The focus state draws a **1px
synapse hairline from the center outward** (`Chat.svelte:515–528`
reference implementation) — this is `MOAT.md §5.2`'s "I am writing
here" gesture applied to every input. Placeholder: **"Say something…"**
(instrument-serif italic, `--content-faint`), not "Ask anything…" — the
brief asked for "Ask anything…" but the existing seed uses "Say
something…" and the seed copy is the right one for an overlay that
hovers above whatever the user is doing. **Confirm the exact word in
review.**

The textarea auto-grows up to 6 lines (`TA_MAX_LINES`); past that, the
inner vertical scroll engages. **The composer is single-line for the
common case.** Multi-line is an opt-in via Shift+Enter, never the default.

**YieldBar (submitting/responding).** Replaces the textarea in the same
slot (height collapse, 240ms ease). Composition:

- A `Pulse phase="thinking" size={10}` on the left (the agent heartbeat).
- Centered italic Instrument Serif 14px text in `--pollen`: **"Condura is
  writing…"** while streaming, **"Done."** on completion.
- A 24px-tall **Stop** chip on the right (mono-uppercase 10px,
  `--pollen` text, `--pollen` border, `--pollen-halo` on hover). Clicking
  it calls `conversation.cancel()` and returns the overlay to the
  error/typing state.

**ErrBlock (error).** Italic Instrument Serif title 14px + sub 13px
(max 48ch), `--danger` border (32% tint), `--danger` background (6%
tint), 1px hairline at the bottom that draws in left→right over 600ms
with a 120ms delay (`err-hair-draw`, ref. `Chat.svelte:388`). Three
required lines per `MOAT.md §2.6`:

1. **What failed** in one noun ("The daemon dropped the thread.")
2. **Likely cause** in one phrase (carried from
   `conversation.streamingError` raw text)
3. **Next action** as a button — **"Try again"** primary or
   **"Open full Chat →"** threadlink as the bridge.

### 2.4 DynamicIsland mini-pill

A 36px-tall pill rendering only in states `submitting`, `responding`,
`done`, and `error`. Slides in below the body in 240ms (the "yield
gesture" shared with `Chat.svelte`'s stream yield slot). Composition
varies by state — see §3.2.

The pill is the **glance-OK**: a user pressing Esc while the agent is
responding sees the pill as the last thing; if they were quick, they
will read the summary inline without re-opening the overlay.

### 2.5 Footer (mic + Send)

A 32px-tall row.

- **Mic** (left, 32×32): a circular pill with `--synapse` border and a
  centered 12px `--pollen` dot that breathes at 1.6s. Hover fills to
  `--synapse` tint (8%) with a `--synapse-halo`. The mic is a **first-
  class affordance**, never hidden behind a menu (brief §7, `MOAT.md`
  restraint test "no hidden affordances"). The dot is **purely visual**
  in v0.1.0 (Push-to-talk integrates in v0.2.0; the affordance ships now
  so the user discovers it).
- **Send** (right, 32px-tall magnetic CTA): `<Button variant="primary"
  magnetic>`. Label **"Send"** + a thin arrow `↗` that translates (+2,
  -2)px on hover. The button morphs through states:
  - **idle**: solid `--pollen` primary. Label: "Send ↗".
  - **press** (mousedown/touchstart): `scale(0.97) + brightness(0.95)
    + saturate(1.1)` (`MOAT.md §2.2` press vocabulary).
  - **submitting**: the icon glyph morphs into a **synapse-pulse**
    (8px pollen dot growing/contracting — `magnetic.ts` reuse).
  - **streaming**: replaced by **"Stop"** button (polled, no magnetic).
- **Magnetic directive** (`magnetic.ts`): on pointermove within ~96px of
  the Send button, the button's transform shifts toward the cursor by up
  to 4px. Already a signature of the seed implementation. Costs ~10
  lines; looks premium.

### 2.6 Hints row (idle/typing only)

A centered mono-uppercase 10px line in `--content-faint`, hidden during
streaming: **"⌘↵ to send · Esc to close · idle dismiss in 5s"**.

This is the only place in the Condura shell where the keyboard story is
spelled out on-screen. It is honest (`prefers-reduced-motion` does not
remove it; the idle line stays true). The hint disappears once the
stream starts — the user is no longer typing, they are reading.

### 2.7 "Open full Chat →" threadlink (done/error only)

A centered mono 11px line with the arrow `→`, **only rendered** in the
`done` and `error` states (the two moments a user might want to dive
deeper). The text inherits the synapse hairline underline on hover (the
same `Thread` color treatment from `Channels.svelte:423–438`). Clicking
calls `onclose()` then `window.location.hash = '#/'` — the bridge to
the full Chat surface.

The threadlink is **deliberately absent in idle/typing**. Hide the
exit during arrival; show the exit once the answer has landed.

---

## 3. STATE MATRIX

### 3.1 State definitions

The overlay has **seven reachable states** + `closed` (the default).

| State         | When                                    | Visual signature                          |
|---------------|------------------------------------------|--------------------------------------------|
| **closed**    | `open === false`                         | not rendered (`{#if open}` short-circuits) |
| **open-empty**| just summoned, `inputText.length === 0`  | header chips + placeholder + 3 recent-prompts list |
| **open-typing** | user typing, `inputText.length > 0`, not streaming | input shows user text + char count (mono) if >200 |
| **submitting**| Enter pressed, `conversation.send` invoked, before first token | Send icon → synapse-pulse, DynamicIsland pill slides in |
| **responding**| `conversation.isStreaming === true` and at least one token has arrived | pill shows streaming text, input shows faint echo of prompt |
| **done**      | stream completed, `conversation.isStreaming === false` after a send | pill shows single-line summary, "Open full Chat →" link appears |
| **error**     | `conversation.streamingError !== null`   | err-block replaces textarea, pill shows the same error |
| **idle-dismiss**| 5s elapsed without activity (and not in submitting/responding/error, and no unsent draft) | fade out + slide down 4px over 160ms |

### 3.2 Per-state content slots

```
                  header  textarea   pill    hints  threadlink  err-block  send-label
closed             —        —         —       —      —           —          —
open-empty         ✓        placeholder✓       ✓      —           —          "Send ↗"
open-typing        ✓        user-text ✓        ✓      —           —          "Send ↗"
submitting         ✓        yield-bar ✓        —      —           —          synapse-pulse
responding        ✓        yield-bar ✓        —      —           —          "Stop"
done              ✓        yield-bar ✓        —      ✓           —          "Stop" (held)
error             ✓        err-block  ✓        —      ✓           ✓          "Try again"
idle-dismiss      slides down + fades (the entire card transitions to closed)
```

### 3.3 Copy (exact strings)

Every string below is the literal copy that renders. **No paraphrasing.**

**Placeholder (open-empty):**
```
Say something…
```
(instrument-serif italic, `--content-faint`)

**Recent-prompts list (open-empty, ≤3 items):**
Each row is the prompt's opening line, mono 13px, `--content-mute`,
truncated to one line with ellipsis. A `→` on hover implies "ask this
again". Click fills the textarea with the full original prompt.

**Char count (open-typing, only if input length >200):**
```
{chars}/2000
```
in `--font-mono`, 10px, `--content-faint`, right-aligned inside the
input. Soft cap at 2000 (the channel limit before IPC rate-limits
kick in).

**YieldBar (submitting/responding/done):**
- submitting: `Condura is writing…`
- responding (early): `Condura is writing…` (same — the pill is
  visual; streaming text isn't here in the seed, but we may pipe a
  60-character echo of the latest token in v0.1.1)
- done: `Done — read the full answer in Chat.`

**ErrBlock title (error):** `The daemon dropped the thread.`
**ErrBlock sub (error):** the raw `conversation.streamingError` text +
` " Press your hotkey again, or check the daemon."`

**DynamicIsland pill (mini — used by the done state):**
A single-line summary, truncated to 60ch + ellipsis, in
`--font-display` italic 13px, `--content-mute`. Example: `Opened
Safari to the docs page you had open yesterday.`

**Hints row (idle/typing):**
```
⌘↵ to send · Esc to close · idle dismiss in 5s
```

**"Open full Chat →" threadlink (done/error):**
```
Open full Chat →
```

### 3.4 State transitions

```
[closed]
   hotkey tap, focus trap armed
        ↓
[open-empty]      ←──┐
   user types         │
        ↓             │
[open-typing]        │
   Enter / Send click │
        ↓             │
[submitting]         │
   first token arrives│
        ↓             │
[responding]         │
   stream ends cleanly│
        ↓             │
[done]               │
   stream errors      │
        ↓             │
[error]              │  ↑ Esc / 5s idle dismiss
        └────────────┘
```

### 3.5 Closed state

When `open === false`, the overlay is **not rendered at all** — Svelte's
`{#if open}` short-circuits every DOM node, every timer, every global
event listener. This is the guarantee `MOAT.md §4.7` requires: "no
spinners that imply backend waits when the answer is already on disk."
A closed overlay is no overlay.

When the user presses the hotkey while closed:
1. `open` flips to `true` synchronously.
2. `draw = false` → `requestAnimationFrame(() => draw = true)` (the
   thread draws on the next frame).
3. `textareaEl?.focus()` is called **after an 80ms delay** (`MOAT.md`
   intent: "the overlay arrives first, then the user writes"). Focus
   during the entrance animation feels jumpy.
4. `idleTimer = setTimeout(onIdleDismiss, 5000)` armed.
5. Global `keydown` listener registered for Esc.
6. `drawTimer`/`idleTimer` cleared on close.

---

## 4. MOTION CHOREOGRAPHY

Every motion has a meaning. **No decorative loops.** (`MOAT.md §4.10`)

### 4.1 Open (summon)

| Element                | Motion                                              | Duration | Easing               |
|------------------------|------------------------------------------------------|----------|----------------------|
| Card                   | `opacity: 0 → 1` + `translateY(8px) → 0`            | 200ms    | `--ease` (cubic-bezier(0.22, 1, 0.36, 1)) |
| Pollen halo (::before)| fade in with the card (ambient, no separate motion) | 200ms    | `--ease`             |
| Top-edge thread        | `scaleX(0→1)` from left (`--synapse`)                | 520ms    | `--ease`             |
| Input focus ring       | drawn inward from center as input gains focus (80ms after summon) | 240ms | `--ease` |
| Wake/Model chips       | fade in, no stagger (they're context, not content)   | 160ms    | `--ease`             |

**Why 200ms.** The seed animation is already 200ms (`qp-enter`); 200ms
is the right number for "appears, doesn't intrude." Faster feels
broken; slower feels slow. (Reference: `Chat.svelte` route-enter is
280ms — slower because the user actively navigates; the overlay is
passive, so 200ms is correct.)

**Focus delay rationale.** 80ms gives the user's eye a frame to register
"the card is here" before the input gains its hairline. Pulling focus
during the slide-up makes the input ring land in motion — visually
noisy.

### 4.2 Submit (Enter / Send click)

| Element                | Motion                                              | Duration | Easing |
|------------------------|------------------------------------------------------|----------|--------|
| Send button            | `scale(1) → scale(0.97)` on mousedown                | 80ms     | `--ease` |
| Send button            | `scale(0.97) → scale(1)` on mouseup                  | 120ms    | `--ease` |
| Send button            | `brightness(1) → 0.95` + `saturate(1) → 1.1`        | 80ms     | `--ease` (concurrent) |
| Send icon              | morphs into a synapse-pulse (8px dot, breathing)     | 200ms    | `--ease` |
| Textarea               | collapses to 0 height                                | 240ms    | `--ease` (replaced by YieldBar in same slot) |
| YieldBar               | fades in + `translateY(-2px) → 0`                    | 240ms    | `--ease` |
| DynamicIsland pill     | fades in + slides in (`translateY(4px) → 0`)         | 240ms    | `--ease` |

The press state uses the `MOAT.md §2.2` press vocabulary: shrinkage
plus weight (brightness+saturation) plus a `translateY(0.5px)` settle.
**Not** just `scale(0.97)` — that has been the failure mode the brief
calls out.

### 4.3 Responding (streaming)

The user is reading. **No spinners** (`MOAT.md §4.7`). Activity is
expressed by:

- **Pulse** (`size=10`, `thinking`) in the YieldBar, breathing at 2s
  (the agent's heartbeat).
- **Thread draw across the pill's bottom edge** as activity — the
  existing `Thread` component's `draw={isStreaming}` flag animates a
  `--synapse` line at the pill's base. 16ms cadence per token arrival is
  what the brief specified; we translate that to a single thread-draw
  that resets on each new token (the line re-draws, `0 → 1`, on every
  delta). **Not** sixteen-millisecond separate animations per token —
  that would thrash the GPU and pull focus from the response text.

**Pill text echo (v0.1.1, post-launch).** The first 60 characters of
the latest assistant token appear in the pill in `--font-display`
italic, fading in for 80ms on each update. **Not in v0.1.0** — keeps
the surface lean for ship. The seed already does not pipe text into
the pill; we honor that.

### 4.4 Done

| Element      | Motion                                                              | Duration |
|--------------|---------------------------------------------------------------------|----------|
| Pulse         | `thinking` → `ok` (synapse, stable, no breathing)                   | 200ms    |
| Pollen check | a 12px pollen check-mark draws itself at the right of the YieldBar   | 360ms    |
| YieldBar text | cross-fades from "Condura is writing…" to "Done."                    | 200ms    |
| Thread link  | fades in (`opacity 0 → 1`) after the pill resolves                  | 240ms    |

The pollen check-mark is the **commit gesture** — it answers
`MOAT.md §3`'s "what says 'this is now finished'": in the overlay, the
check-mark IS the thread. It draws in 360ms (longer than the send
press, because the user now has nothing to do but watch).

### 4.5 Dismiss (Esc / 5s idle / close button)

| Element | Motion                                              | Duration | Easing |
|---------|------------------------------------------------------|----------|--------|
| Card    | `opacity: 1 → 0` + `translateY(0) → translateY(4px)` | 160ms    | `--ease` |
| Pollen halo (::before) | fade out with the card                    | 160ms    | `--ease` |

**Dismiss is faster than summon (160ms vs 200ms).** The user has
already chosen to leave; honoring that choice quickly is more premium
than a slow reverse-animation. (The reverse of the slide-up would be
`translateY(-8px) → 0` per `MOAT.md §4.10` "every animation must answer
'what is this communicating?'" — but the 4px downward feel reads as
"the card is being let go," which is what dismissal should feel like.)

### 4.6 Pin (user pressed Esc but wants to stay)

If `inputText.trim().length > 0` OR `conversation.isStreaming` OR
`conversation.streamingError`, the idle-dismiss timer **does NOT fire**
— instead, `markActivity()` re-arms the timer. The Close button changes
its label from `✕` (glyph) to `Pinned` (mono 10px text) when the timer
is suspended by an active condition. **The visual tells the user why
the overlay is still here.** This is a low-cost `MOAT.md §5` move:
"remove this and the product would just be confusing."

### 4.7 Reduced motion

`@media (prefers-reduced-motion: reduce)` collapses all durations to
0.01ms globally (already in `condura.css`). Per-component additions:

- The 200ms slide-up becomes instant (`enter` class omitted → card
  paints at rest).
- The pollen halo (`::before`) is **hidden** entirely (it has no
  functional purpose).
- The mic dot stops breathing.
- The textarea focus hairline is suppressed (the synapse halo remains;
  see `MOAT.md §2.1` for focus rings).
- The yield slide-in is instant.
- **The thread still draws.** It is the signature; we don't gut it for
  reduced-motion users. The reduced-motion respect is "less
  decorative motion," not "no motion."

---

## 5. KEYBOARD MODEL

### 5.1 Global hotkey (open)

The user's chosen overlay hotkey (default `Option+Option`, locked
decision #8). Triggers `open = true` from any state, even when the
Shell is hidden or the focus is in another app. Implementation:
`Shell.svelte:121–124` listens for the global hotkey and forwards to
the overlay. The brief explicitly permits the cursor-anchored summon.

If the overlay is already open and the user presses the hotkey again,
**nothing happens** (the overlay is already up; pressing again is a
no-op, not a re-summon). If the user is mid-stream when they press the
hotkey, the existing overlay stays (we don't dismiss a live
conversation). If the user presses the hotkey while a prompt is in
their textarea, the overlay re-focuses the input (does not clear the
draft).

### 5.2 Esc

Dismiss the overlay from anywhere — global `keydown` listener. If the
overlay is **pinned** (active condition — draft, streaming, error),
Esc still dismisses. **Pinning changes the dismiss affordance, not the
possibility.**

### 5.3 Enter

Submit. Plain Enter submits. **`isComposing`** (IME) blocks Enter so
Japanese / Mandarin / Korean users don't fire mid-composition. **No
"Enter to confirm and close"** — that's a different gesture (`⌘↵`).

### 5.4 Shift+Enter

Insert a newline in the textarea. The textarea is multi-line capable
(it auto-grows up to 6 lines), so this is meaningful.

### 5.5 ⌘↵ (Cmd+Enter / Ctrl+Enter)

**Submit AND open the full Chat in a new hash route.** This is the
bridge gesture: the user knows they're starting deep work, so the
overlay dismisses, sends the prompt, and lands them on `#/` already
mid-stream. They can read the full answer there. Implementation: same
as Enter, but on success, also `window.location.hash = '#/'` before
calling `onclose()`. The full Chat surface's composer should pick up
the in-flight stream from `conversation` store (it already does).

### 5.6 ↑ / ↓ (recall)

- `↑` while the textarea is empty and **not composing**: fills the
  textarea with the most recent prompt from the user's local history
  (stored in `localStorage` under `condura.recentPrompts`, capped at
  10). The cursor stays at the end of the inserted text.
- Each subsequent `↑` moves to the prior prompt.
- `↓` moves forward through the same list.
- The list resets to "newest" when the textarea regains focus or any
  character is typed.

### 5.7 Tab / Shift+Tab

Focus-trap the overlay while open. Tab from the textarea → mic →
Send → close button → textarea (cycle). Shift+Tab reverses. Escape
from the trap (Tab off the last element) is **not allowed** — the user
must use Esc to dismiss.

### 5.8 All other keys

Pass through to the host app. The overlay is **not modal** (`aria-modal="false"`
in the seed); it sits above the host without taking focus from the
host app's text fields when the user's intent is to write there. **The
textarea field is the only key-consuming element in the overlay.**

---

## 6. COMPONENTS USED

The Quick Prompt Overlay composes (does not reinvent):

| Component             | Path (relative to `condura/`) | Purpose                                      |
|-----------------------|-------------------------------|----------------------------------------------|
| `FloatingOverlay`     | primitive (future)            | the card chrome (border, shadow, halo)       |
| `Composer` (minimal)  | `<textarea>` inline (`qp-input`) | single-line input with hairline focus         |
| `SendButton`          | `Button.svelte` variant=primary, magnetic | The human-spark CTA                       |
| `MicButton`           | inline (`.qp-mic`)            | 32×32 circular mic with breathing pollen dot  |
| `DynamicIsland` (mini)| inline or a derived mini-prop variant of `Shell`'s DynamicIsland | The status pill below the body               |
| `Thread`              | `Thread.svelte`               | Top-edge thread + bottom-of-pill activity    |
| `Pulse`               | `Pulse.svelte`                | Heartbeat / thinking / error indicator       |
| `Glyph`               | `Glyph.svelte`                | send / mic / close / stop icons              |
| `Button`              | `Button.svelte`               | Primary magnetic Send, Stop, Try again       |
| `Tooltip`             | new (`MOAT.md §2.9`)          | Hover-revealed hint on model select, mic     |

**Composition rule:** the overlay imports existing primitives. It does
not introduce new components unless one is truly missing — the brief
flags `Tooltip` because `MOAT.md §2.9` calls it out as a gap to close.
The overlay is one of the surfaces that benefits from tooltips on the
mic (hover: "Voice input (v0.2.0)") and the model select (hover: the
provider's full display name and tier).

**Reuse obligations:**
- The pollen halo (`.qp-card::before`) is duplicated in `Chat.svelte`
  and elsewhere — it should collapse to a single `--shadow-halo` token.
  Acknowledge as a follow-up if not already covered.
- The `qp-wake` mono chip should reuse the same chip class as the
  existing `.qp-chip` (already in the seed).
- The mic-with-breathing-dot is bespoke in the seed; consider lifting
  to `MicButton.svelte` for reuse in `CommandPalette` and `FloatingInterview`.

---

## 7. DATA FETCHED

### 7.1 IPC calls (existing in the seed)

| Call                     | When                              | Returned                                         | Failure handling            |
|--------------------------|-----------------------------------|--------------------------------------------------|-----------------------------|
| `ipc.providersList()`    | once on mount                     | `ProviderInfo[]` (drives `selectedModel`)        | `providers = []` (mute chip) |
| `settings.config?.llm?.providers` | on settings store hydration | first enabled provider + default model            | fall through to providerList |
| `settings.config?.hotkey?.overlay` | on settings store hydration | the user's chosen combo (e.g. "⌥⌥")             | fall through to "⌥⌥"        |
| `conversation.send(provider, model, text)` | on Enter / Send click / ⌘↵  | streams the response into `conversation` store   | sets `streamingError`       |
| `conversation.cancel()`  | on Stop click                     | stops the in-flight stream                       | logs, no UI error           |

### 7.2 IPC needs (new for this spec)

| Call                     | When                              | Returned                                         |
|--------------------------|-----------------------------------|--------------------------------------------------|
| `presence.state()`       | **on mount** + **on idle rearm**  | `{active: boolean}` — drives the dismiss rule (`MOAT.md §5`) |

The dismiss rule is currently local: arm a 5s timer, reset on activity.
`MOAT.md §1` restraint test would prefer the dismiss to honor the user's
**actual presence** (away from keyboard → sooner dismiss; actively
typing elsewhere on the desktop → longer grace). Implementation:
`presence.state()` is polled on mount only in v0.1.0 (the seed's
behavior); v0.1.1 may switch to a polling cadence every idle rearm.

We do **not** call `conversation.fetchHistory()` — the overlay has no
history view. The recalled recent-prompts list is local (`localStorage`).

---

## 8. DESIGN DECISIONS

### 8.1 The overlay is a **glance surface**, not a full Chat

`MOAT.md` is the test: QuickPromptOverlay must feel like a **quick
scratchpad**, not a full Chat. Concretely:

- **Width 480px, not 720px.** A full Chat is 720px+; the overlay is
  half. The narrower width says "this is for one thought."
- **No scroll-back, no conversation list.** The textarea grows but the
  history is read-only. Five seconds is the engagement expectation.
- **Single-line default, multi-line on opt-in.** Most quick asks are
  one line. Long thoughts go to Chat.
- **No sidebar.** Sidebar is `MOAT.md §1.5`'s lesson — don't smuggle
  full-app chrome into a glance surface.

### 8.2 The 5s idle auto-dismiss is deliberate

It says **"this is for quick asks, not deep work."** A user who is
typing something longer than 5 seconds should Cmd+Enter to Chat, not
wait for the overlay to dismiss.

We re-arm the timer on **any activity over the card**:
`pointermove`, `pointerdown`, `input`, `keydown`. The user is engaged;
the overlay stays. They are not engaged; it dissolves.

The dismiss **never fires mid-stream** and **never with an unsent
draft** — this is invariant. The pinned state shows the user the
affordance change ("Pinned" instead of `✕`).

### 8.3 The "Open full Chat →" threadlink is the bridge

`MOAT.md §4.7` ("no spinners") and §3 (the thread as the signature) are
both satisfied here: the only way out of "the answer is here" is
**through** the thread. Visually underlined (synapse hairline), placed
centered below the YieldBar. The threadlink only renders in `done` and
`error` states — these are the two states where the user might want
more, and it is dishonest to suggest "Open full Chat" when the answer
hasn't arrived.

### 8.4 The voice mic is a first-class affordance

`MOAT.md §1` restraint test ("no hidden affordances") and §5.1
("hoverable surfaces catch the cursor"). The mic is **always visible**,
**always 32×32**, **always in the footer left slot**. The breathing
pollen dot is the **only** moment-to-moment animation in the seed that
loops without state change — we accept this for v0.1.0 because the dot
advertises that voice exists. v0.2.0 wires it; the affordance ships now.

### 8.5 Anchoring the overlay

The brief allows **cursor position or active-window-center**. Decision:
**default to cursor position with a 96px edge-snapping rule**, falling
back to active-window-center if the cursor anchor would clip. Reasoning:

- Cursor position is the user's attention point — the overlay
  appearing **there** is the cheapest visual scan cost (`MOAT.md §5.1`
  intent).
- Active-window-center is the centered-on-window mode the seed uses; we
  retain it as the fallback for safety (e.g., a screen-recording state
  where knowing the cursor position is sketchy).
- Multi-monitor: anchor to the monitor containing the cursor.

### 8.6 Sent drafts vs unsent drafts on Esc

If the user has typed text but presses Esc, **we lose the draft**. This
is wrong. The decision log:

- **Option A** (this spec): dismiss unsent drafts on Esc, retain via
  `⌘↵` only. Simple, matches the seed.
- **Option B**: Esc triggers "Save draft to localStorage" + "Open
  overlay again to recall." Better UX, more complexity.
- **Decision**: ship **A** in v0.1.0; track **B** as a v0.1.1 follow-up
  (the recall list from `↑` already provides partial mitigation — but
  it only stores completed prompts, not drafts).

### 8.7 The `Open full Chat →` threadlink vs `⌘↵`

They are **two exits, not one**.

- **⌘↵** submits the prompt and immediately routes to `#/`. The user
  is watching the stream progress in the full Chat. This is for users
  who know they want deep work.
- **"Open full Chat →"** routes to `#/` **after** the answer has
  landed in the pill. The user gets a teaser here; deep work happens
  there. This is for users who wanted a quick answer.

A user who wants both: hit ⌘↵. The brief specifies both gestures; we
honor them as two affordances.

### 8.8 The YieldBar replaces the textarea, it does not stack

A common anti-pattern is to **stack** the yield bar below the
textarea (giving the user a frozen textarea with progress below it).
The seed correctly **replaces** the textarea in the same slot — and we
preserve this. The reasoning: a frozen textarea is **a question the
user can't act on** while the agent is thinking. Replacing it (with a
YieldBar that contains a Stop button) gives the user **agency** while
the agent works. This is `MOAT.md §5.3`'s lesson — the Stop button on
streaming is a place where the model yields to the user.

### 8.9 The pulse phase vocabulary

The `Pulse` component takes a `phase` prop. The overlay uses:

- `idle` (synapse, breathing, 4s) — when open and not streaming.
- `thinking` (pollen, animated, 2s) — when streaming.
- `error` (danger, single pulse) — when an error has surfaced.
- `ok` (synapse, stable, no breathing) — when the stream completed.

This matches `APPFLOW.md §3.6` `DynamicIsland` semantics exactly. The
**header Pulse** carries the heartbeat; **the YieldBar Pulse** carries
the same heartbeat louder (`size=10` instead of `size=8`). Two voices,
one rhythm. They agree.

### 8.10 Pinned vs unpinned dismiss affordance

When the idle timer is **suspended** by an active condition (draft,
streaming, error), the close button changes from `✕` (glyph) to
**"Pinned"** (mono 10px text). A **synapse dot** (4px, breathing at
2s) appears next to it. Hovering the pinned close shows a tooltip:
"This overlay is pinned because you're [typing | streaming | showing
an error]. Esc or click to dismiss."

This is one of the **cheapest animations to ship** (`MOAT.md §5.3`)
and one of the highest signal-to-noise improvements in the seed.

---

## 9. SPEC SUMMARY (for the implementer)

The Quick Prompt Overlay is **a thin floating paper card, anchored to
the cursor, 480px wide, that turns one focused thought into a delegated
agent run.** It is the **glance surface**, not the deep-work surface.
The 5-second idle dismiss is the design — not a limitation. The
"Open full Chat →" threadlink is the bridge — explicit, not accidental.
The voice mic is the affordance — first-class, not hidden. The
YieldBar replaces the textarea mid-stream — agency, not a frozen UI.
The pinned state tells the user why the overlay is still here. The
thread is the signature — every entrance draws it; every reduction
keeps it.

The implementer should:

1. Read `MOAT.md` §1 (restraint), §3 (the thread), §4 (anti-patterns),
   §5.2 (composer focus hairline), §5.3 (Stop break-the-model gesture).
2. Preserve the seed's existing motion vocabulary (`qp-enter`,
   `qp-thread`, `qp-mic` breathing dot, `qp-stop` halo) — they are the
   foundation.
3. Add: `recent-prompts` list (idle state, ≤3 rows), `Open full Chat
   →` threadlink (done/error states), `Pinned` close affordance,
   cursor-anchored positioning with edge-snap, `↑/↓` prompt recall.
4. Remove: the model select becomes hover-tooltip-on-Wake (it is
   redundant — the user picked a model in Chat; don't re-prompt in the
   overlay). **Verify the intent** before deleting — `MOAT.md §1.5`
   says "if removing this make the product worse, it's earned;
   otherwise, delete it." Default model selection is **earned** by
   retaining; if absent, the overlay is useless without a model.
5. Refactor: the model select moves to a `<Tooltip>` on the Wake chip —
   the chip becomes "↧ model-name" with hover revealing the picker.

### 9.1 Acceptance criteria

| # | Criterion | Test                                                                                                                                  |
|---|-----------|---------------------------------------------------------------------------------------------------------------------------------------|
| 1 | Cursor anchor | Move cursor to top-right of screen; press hotkey; overlay appears in top-right area (within 96px edge snap).                       |
| 2 | Idle dismiss | Type one character, leave overlay alone for 5s; overlay fades + slides down; idle timer resets on character input.                  |
| 3 | Streaming collapse | Press Enter with text typed; textarea collapses; YieldBar slides in within 240ms; Pollen Pulse breathes in YieldBar.                 |
| 4 | Stop button | Click Stop while streaming; the stream cancels; overlay returns to typing state; the Send button reappears.                        |
| 5 | Done state | After stream completes; YieldBar shows "Done."; threadlink "Open full Chat →" fades in below; pollen check-mark draws at right.    |
| 6 | Recalled prompts | Press hotkey, press ↑; the most recent prompt fills the textarea; press ↑ again; the prior prompt fills; press Esc; draft lost.    |
| 7 | ⌘↵ bridge | With text typed, press ⌘↵; overlay dismisses; `window.location.hash === '#/'`; conversation stream is already in progress in Chat. |
| 8 | Pinned affordance | Type text, leave idle (idle timer suspended); close button shows "Pinned" + synapse dot.                                              |
| 9 | Reduced motion | `prefers-reduced-motion: reduce`; summon is instant (no 200ms slide); thread draw is instant; pollen halo hidden.                   |
| 10 | Error recover | Force a `streamingError`; ErrBlock renders with title "The daemon dropped the thread.", hairline draws in 600ms; "Try again" + threadlink appear. |

### 9.2 Definition of done

- [ ] All ten acceptance criteria pass on macOS, Windows, Linux dev
      machines.
- [ ] `make verify` green: 0 lint issues, all Go + TS + Svelte tests
      pass with `-race`, vitest suite green with `KillSwitchOverlay`
      + `Pulse` + a new `QuickPromptOverlay.test.ts` covering
      criteria 2, 3, 5, 8, 10.
- [ ] No new dependencies added unless documented in `CLAUDE.md §8`.
- [ ] A `prefers-reduced-motion: reduce` test case exists in
      `QuickPromptOverlay.test.ts` (asserts no `qp-enter` class is
      applied).
- [ ] Live verification on a clean macOS machine via the Phase 15
      checklist (`docs/phase15-verification.md`).

---

**This document is the architecture. The code is the implementation.
They agree. When they diverge, the divergence is the spec-bug — fix
the doc, then fix the code, in one commit.**
