# Condura · DIRECTION

> **The north-star document for the Condura shell.** Read this before opening
> any surface. Every screen, token, and animation in
> `app/web/frontend/src/lib/condura/` is downstream of the six sections
> below. If a surface contradicts one of them, the surface is wrong; fix
> the surface, not the direction.
>
> **Sister documents.** `TEARDOWN.md` is the *why* (what we stole from
> Claude/ChatGPT/Manus/Linear/Arc/Notion, what we skipped, what we made
> our own). `MOAT.md` is the *bar* (the five tests a premium product must
> pass). `APPFLOW.md` is the *spec* (every screen the user can land on,
> the data each writes, the animations each owns). `DIRECTION.md` is the
> **voice** — what Condura sounds, looks, and feels like when every screen
> in those docs is built right.

---

## 1. THE PERSONALITY

**Condura is a paper notebook that learned to listen — warm, awake, and never louder than the room it's in.**

That sentence is the only test that matters. Every screen, every animation,
every word of copy answers one question: does this read like *that*?

- **Paper notebook, not app.** A surface should feel like ink on a page,
  not pixels on a screen. The grain texture (`.paper-grain`), the warm
  hex `#F4EFE4` by default, the hairline rules instead of cards-of-cards,
  the serif headlines — these are not decoration; they are the surface's
  identity. A surface without paper texture is not a Condura surface.
- **Learned to listen, not engineered to compute.** The agent is attentive.
  It perceives what it must, asks before it acts, draws a thread when it
  finishes. It does not flash, pulse, or celebrate. It does not say
  "Welcome!" or "Great choice!" It says, in the quietest voice it can,
  exactly what is happening and why.
- **Warm, not cool.** No blue gradients. No teal accents. No neon anywhere.
  The brand has two colors: synapse (deep forest green) and pollen
  (warm amber). Everything else is ink on paper. Warmth is the
  differentiator; coldness is the failure mode.
- **Awake, not loud.** The agent is alive — it breathes (the Pulse), it
  listens (the garden motes), it draws (the Thread). But it never
  performs these things. They appear when the room is quiet and disappear
  when the user is moving. Premium products *are* alive; cheap products
  *show* they are alive.
- **Never louder than the room it's in.** This is the hardest constraint.
  When the user is mid-task, the agent is silent. When the user pauses,
  the agent surfaces a single, useful detail. When the user is idle,
  the agent may breathe, but never speak first. The product earns the
  right to exist on the screen by being useful, never by being visible.

---

## 2. THE PRE-RITUAL CONCEPT

### The call: **Constellation-as-Room.**

The current `Ritual.svelte` is a 9-step forced wizard — arrival, EULA,
permissions, power, hotkey, voice, channels, account, breath. It works,
but it violates §I1 of `APPFLOW.md` ("Configure, not comply") in spirit:
the user is *complying* with a sequence rather than *configuring* their
agent. The TEARDOWN proposes replacing it with a single constellation
screen (six live nodes, side panels, hover previews, an "Enter Condura"
pill that is always enabled). **We adopt that proposal in full.** This
is the call: the constellation becomes the room, the wizard dissolves.

### The exact screen list (in render order)

1. **Cinematic Arrival** (full-bleed, 1.8 s, then dissolves).
   - 0–400 ms: paper void fades in (`.a-void`, no `voidHold` linger).
   - 400 ms: the wordmark "Condura" wipes in via `clip-path: inset(0 100% 0 0)`
     → `inset(0 0 0 0)` over 700 ms `--ease` (`.wordReveal`).
   - 900 ms: the 1.5 px synapse underline draws in left→right over 700 ms
     (the **Thread** already in the design system).
   - 1500 ms: a 10 px pollen mote breathes at the underline's right edge
     (single ambient `breathe`, no `firstBeat`, no `moteDrift`).
   - Total: **1.8 seconds, no six-keyframe cascade**. The user sees a name
     arrive. That is enough.
   - Click anywhere (or `↵`) → advance to EULA.

2. **EULA Gate** (single screen, legal-first).
   - Headline: "First, the terms." (Instrument Serif, 32 px).
   - Scrollable license in a paper-card well with a 2 px synapse
     progress bar on the left edge (height tracks scroll percentage).
   - Checkbox + the existing wax seal stamp button (`<button class="seal">`).
   - The seal is the only CTA — no "Continue" button. Stamping it
     triggers `sealBloom` (the radial pulse, 600 ms) → dissolves to the
     Constellation.
   - Bottom-left margin: a single italic skip-note, "not now · quit."
     This is the only escape from the legal step, by design.

3. **The Constellation** (the room).
   - Full-bleed. Constellation SVG centered, **but the six nodes are
     now live clickable surfaces** with monospace status labels under
     each one (10 px, `--content-faint`):
     - **Perceive** — `Accessibility: granted · Screen: pending`
     - **Power** — `Ollama: 3 models detected` or `Ollama: not detected`
     - **Summon** — `Hotkey: ⌘⇧Space` or `Hotkey: not set`
     - **Voice** — `Mic: available · wake: off` or `Mic: not detected`
     - **Threads** — `0 of 5 ready` or `1 of 5 ready · Telegram`
     - **Account** — `Not signed in` or `Signed in: email`
   - **Click a node → right-side panel slides in** (380 px wide,
     `translateX(24px) → 0`, `opacity 0 → 1`, 320 ms `--ease`). The
     panel contains the same step-specific UI we have today (radio
     cards for power, keycap surface for hotkey, channel chips for
     threads, email field for account, permission rows for perceive,
     voice-card for voice) — relocated, not redesigned.
   - **Hover a node → 60 px-tall preview strip** appears below the
     constellation, fading in over 220 ms (`--ease`) with an 80 ms delay.
     Hover Summon → "Press your combo to call Condura" + a one-line
     keycap row. Hover Power → "Your model. Your key. Local or remote."
     This is Arc's hover preview, condensed.
   - **Bottom-center: "Enter Condura →" pill.** Always enabled. The
     only soft-lock is when Summon (hotkey) is unset — the pill pulses
     a faint synapse halo and reads "Set a hotkey to enter →." One
     click on the Summon node, one keypress, and the lock releases.
   - **Idle invitation (when something is unwired):** one slow pollen
     mote drifts from the center toward one unwired node every 14 s,
     1.4 s duration, then dissolves. Suggests "tap me" without nagging.
     Stops once the user has wired at least one node.

4. **First Breath** (the closing handoff, shown when "Enter Condura"
   is clicked).
   - Centered hero: "Condura is here." in Instrument Serif italic with
     the "alive" synapse-green accent on the word *here*.
   - A 48 px breathing pollen sphere (the `breathe` keyframe, 4 s loop).
   - One CTA: "Press your hotkey to summon →." (No enter-button —
     the actual hotkey press *is* the activation.)
   - When the user presses their hotkey anywhere on the desktop, the
     ritual gets `.dissolving` (700 ms fade + 8 px blur) and the Shell
     fades in beneath. The moment of *invocation* is the moment the
     user knows the agent is alive.

### What the user CHOOSES (six nodes, independent, any order)

| Node | What the user picks | Default if skipped |
|---|---|---|
| **Perceive** | Accessibility + Screen Recording grants | Neither granted (computer use is gated) |
| **Power** | Ollama local / paste API key / connect subscription | Local Ollama if reachable; otherwise "no provider" |
| **Summon** | The hotkey combo | **No default — locked decision #8** |
| **Voice** | Wake word on/off + sensitivity | Off |
| **Threads** | Connect Telegram (the only wired channel in v0.1.0) | None |
| **Account** | Magic-link sign-in | Signed-out (everything works locally) |

### What the user SKIPS

- The hotkey prompt (only soft-lock — set or no entry).
- Voice, channels, account (all skippable by simply not tapping the
  node — the **architecture is the skip**).
- Permissions can be granted later in Settings — computer use is gated
  but chat still works without them.

### Hero moment

The wax seal stamp on the EULA screen. `sealBloom` is the single most
premium micro-interaction in the product. Nothing else earns the
0.5-second dwell time. Reserve the seal for legal consent only — never
re-use it as a generic "confirm" button.

### Cinematic intro

The 1.8 s arrival sequence above. Six keyframes collapsed to four
(`voidHold` and `moteDrift` deleted per MOAT.md §1.1). The user's
attention for a wordmark reveal is shorter than 1.8 s; we use all of it
without exceeding it.

---

## 3. TYPOGRAPHY

**Pairing locked.** Instrument Serif (display, italic accents) + Inter
(UI) + JetBrains Mono (metadata, eyebrows, timestamps, kbd hints). No
fourth family. Every text node in the codebase reads from this scale.

### Type scale (exact px / weight / leading / tracking)

| Role | Family | Size | Weight | Line-height | Letter-spacing | Use |
|---|---|---|---|---|---|---|
| **display** | Instrument Serif | `clamp(40px, 5vw, 72px)` | 400 | 0.95 | −0.040em | Hero only — Chat empty, About colophon, Ritual Arrival wordmark |
| **h1** | Instrument Serif | 32 px | 400 | 1.10 | −0.030em | Route titles (Settings sections, About header, Channels h1) |
| **h2** | Instrument Serif | 24 px | 400 | 1.20 | −0.025em | Sub-section titles within a route |
| **h3** | Inter | 20 px | 600 | 1.30 | −0.020em | Card titles, sheet headers, dialog headlines |
| **h4** | Inter | 17 px | 600 | 1.40 | −0.015em | Row labels, form labels, list-item titles |
| **body** | Inter | 15 px | 400 | 1.60 | −0.008em | Default body, table cells, lists |
| **lead** | Inter | 17 px | 400 | 1.55 | −0.010em | Hero sub-copy (one per surface max — `max-width: 52ch`) |
| **caption** | JetBrains Mono | 12 px | 500 | 1.40 | +0.040em | Eyebrows, status pills, kbd hints |
| **mono** | JetBrains Mono | 12 px | 500 | 1.50 | +0.120em | Code, paths, IDs, timestamps, `ENG-2703`-style references |
| **micro** | JetBrains Mono | 10 px | 500 | 1.40 | +0.100em | Node-counts (`02/145`), status badges, signal-bar hints |

### Tokens (replace existing `--text-*` in `condura.css`)

```css
--text-display: clamp(40px, 5vw, 72px);
--text-h1:      32px;
--text-h2:      24px;
--text-h3:      20px;
--text-h4:      17px;
--text-lead:    17px;
--text-body:    15px;
--text-caption: 12px;
--text-mono:    12px;
--text-micro:   10px;

--lh-display: 0.95;
--lh-h1:      1.10;
--lh-h2:      1.20;
--lh-h3:      1.30;
--lh-h4:      1.40;
--lh-lead:    1.55;
--lh-body:    1.60;
--lh-caption: 1.40;
--lh-mono:    1.50;
--lh-micro:   1.40;

--ls-display: -0.040em;
--ls-h1:      -0.030em;
--ls-h2:      -0.025em;
--ls-h3:      -0.020em;
--ls-h4:      -0.015em;
--ls-lead:    -0.010em;
--ls-body:    -0.008em;
--ls-caption: +0.040em;
--ls-mono:    +0.120em;
--ls-micro:   +0.100em;
```

### Italic accent

The single load-bearing phrase per surface is rendered in **Instrument
Serif italic, `--synapse`** color. One per surface, never more. This is
the `.alive` class — reserved for the headline word that earns the
gesture. Five uses of `.alive` is a tic; one is a wink.

### kbd hint style

`<kbd>` elements: JetBrains Mono, 10 px, +0.08em tracking, uppercase,
paper-card background, 1 px hair-strong border, 6 px radius. Never
plain text — keyboard shortcuts must read as keyboard shortcuts.

---

## 4. COLOR

### The decision: **light is the default.**

Dark mode exists; light is the front door. A new user opens Condura
and sees `#F4EFE4` warm paper, `#14110B` ink, `#0B3D2E` synapse, and
`#C97B2E` pollen. Every other color is downstream of those four.

### Role → hex table (light mode — the default)

| Role | Hex | Notes |
|---|---|---|
| `--paper` (the page) | `#F4EFE4` | Warm cream. Never pure white. |
| `--paper-2` (cards sitting on paper) | `#ECE5D4` | One step warmer. |
| `--paper-raised` (floating surfaces) | `#FBF7EC` | One step lighter than paper. |
| `--paper-sunken` (recessed panels, input wells) | `#E2DAC6` | One step cooler. |
| `--paper-shadow` (deepest paper tint, dark mode `--paper` base) | `#D8CFB8` | One step cooler than sunken. |
| `--surface` (alias of paper) | `#F4EFE4` | The page. |
| `--surface-card` | `#ECE5D4` | A card. |
| `--surface-raised` | `#FBF7EC` | A popover, tooltip, modal. |
| `--surface-sunken` | `#E2DAC6` | A recessed well. |
| `--surface-ink` (rare inverse surface) | `#16140F` | The dark pill on a light page. |
| `--ink` (primary content) | `#14110B` | Warm near-black. Never `#000000`. |
| `--ink-soft` | `#2A2519` | Secondary content (bodies, descriptions). |
| `--ink-mute` | `#5C5443` | Tertiary (timestamps, hints, meta). |
| `--ink-faint` | `#8A8169` | Quaternary (labels, eyebrows in resting state). |
| `--ink-ghost` | `#B8AF98` | Decorative only (never for content the user must read). |
| `--content` | `var(--ink)` | The contract — components read this, never `--ink-*` directly. |
| `--content-soft` | `var(--ink-soft)` | |
| `--content-mute` | `var(--ink-mute)` | |
| `--content-faint` | `var(--ink-faint)` | |
| `--hair` | `rgba(20, 17, 11, 0.10)` | The only sanctioned way to draw a 1 px divider. |
| `--hair-strong` | `rgba(20, 17, 11, 0.18)` | The only sanctioned way to draw an emphasized divider. |
| `--synapse` (brand accent, links, "alive") | `#0B3D2E` | Deep forest green. Never teal, never mint. |
| `--synapse-deep` (pressed / inset) | `#06241B` | One shade darker than synapse. |
| `--synapse-glow` (focus, highlight) | `#1A8A6A` | One shade brighter. |
| `--synapse-light` (wash backgrounds) | `#9CE8C8` | For tinted surfaces only, never for content. |
| `--pollen` (action, brand accent #2) | `#C97B2E` | Warm amber. The CTA color, the mote color, the cursor-tip color. |
| `--pollen-deep` (pressed pollen) | `#9A5A1A` | One shade darker. |
| `--pollen-light` (pollen wash backgrounds) | `#F0C082` | For tinted surfaces only. |
| `--accent` (alias of synapse for component code) | `var(--synapse)` | The single brand accent. |
| `--accent-glow` | `var(--synapse-glow)` | |
| `--action` (primary CTA) | `var(--pollen)` | The button color, not the link color. |
| `--focus` (focus ring base) | `var(--synapse-glow)` | |
| `--danger` | `#A3312A` | Rust red, not stop-sign red. |
| `--ok` (success, autonomous) | `#2E7D4F` | Forest green, distinct from `--synapse` so status and brand don't blur. |
| `--warn` (warn state, awaiting action) | `#B07A2E` | Amber, distinct from `--pollen` so status and action don't blur. |
| `--info` | `#3A5A8C` | Muted indigo. Rare; used only in About or Audit meta. |

### Role → hex table (dark mode — `data-mode="dark"`)

| Role | Hex |
|---|---|
| `--paper` | `#16140F` |
| `--paper-2` | `#1F1C16` |
| `--paper-raised` | `#221E17` |
| `--paper-sunken` | `#2A2519` |
| `--paper-shadow` | `#100E0A` |
| `--surface` | `var(--paper)` |
| `--surface-card` | `var(--paper-2)` |
| `--surface-raised` | `var(--paper-raised)` |
| `--surface-sunken` | `var(--paper-shadow)` |
| `--surface-ink` | `#0B0A07` |
| `--ink` | `#F0EDE5` |
| `--ink-soft` | `#D8D2C4` |
| `--ink-mute` | `#8A847A` |
| `--ink-faint` | `#5C564C` |
| `--ink-ghost` | `#3A352D` |
| `--content` / soft / mute / faint | aliases of the `--ink-*` scale |
| `--hair` | `rgba(240, 237, 229, 0.16)` |
| `--hair-strong` | `rgba(240, 237, 229, 0.28)` |
| `--synapse` | `#1A8A6A` (brighter in dark so contrast holds against `#16140F`) |
| `--synapse-deep` | `#0B3D2E` |
| `--synapse-glow` | `#2FBF95` |
| `--synapse-light` | `#0B3D2E` |
| `--pollen` | `#E0A862` |
| `--pollen-deep` | `#C97B2E` |
| `--pollen-light` | `#9A5A1A` |
| `--danger` | `#C75449` |
| `--ok` | `#4FAE74` |
| `--warn` | `#D49A4E` |
| `--info` | `#6B8FB8` |

### The rule for status vs. brand

Brand colors (synapse, pollen) and status colors (ok, warn, danger, info)
are **never** reused across categories. `--synapse` is the brand link
color; `--ok` is the success badge color. They look related by design —
both greens, both warm — but they are not interchangeable. This is so
the user can read a green border on a card and know "that's a link,"
and a green border on an autonomy-matrix dot and know "that's
autonomous," without ambiguity.

### No new colors without an amendment

If a surface needs a hue that isn't in the four brand colors, the four
status colors, or the ink scale, **the surface is wrong.** Adding a new
color requires a CLAUDE.md amendment (per MOAT.md §4 rule 4). The 2026
rule is: a constrained palette reads as designed; an open palette reads
as improvised.

---

## 5. MOTION GRAMMAR

### Eases (locked)

```css
--ease:    cubic-bezier(0.22, 1, 0.36, 1);   /* out — the default */
--ease-in: cubic-bezier(0.65, 0, 0.35, 1);   /* in  — entrances */
--ease-pop: cubic-bezier(0.34, 1.56, 0.64, 1); /* press, single-element scale — used at most once per surface */
```

The default for every transition is `--ease`. Components do not declare
their own cubic-beziers. `--ease-in` is reserved for entrances
(panel-slide-in, route-enter, modal-reveal). `--ease-pop` is reserved
for the dot-pop on the autonomy matrix and the seal bloom on the EULA
stamp — two places, no more.

### Durations (locked)

```css
--dur-fast: 140ms;  /* state changes (hover, press, focus halo, color shifts) */
--dur:      280ms;  /* the everyday transition (border-color, background, transform) */
--dur-slow: 520ms;  /* route enters, panel slides, threads draw, signature gestures */
--dur-cine: 900ms;  /* Ritual blooms, modal reveals, the wordmark arrival */
```

The four tokens are the entire vocabulary. Components do not declare
`transition: ... 320ms`. They use `--dur`. If a surface needs a duration
that isn't in the four, the surface is wrong — find a way to express the
gesture in 140 / 280 / 520 / 900 ms.

### What moves (and how)

| Gesture | Property | Duration | Easing | Trigger |
|---|---|---|---|---|
| **Hover (button)** | `transform: translateY(-1px)` + `border-color: --hair-strong` | `--dur` | `--ease` | mouseenter |
| **Hover (card lift)** | `transform: translateY(-2px)` + `--shadow-card` → `--shadow-float` | `--dur` | `--ease` | mouseenter |
| **Press (tactile)** | `transform: scale(0.97)` + `filter: brightness(0.95) saturate(1.1)` + `translateY(0.5px)` | `--dur-fast` | `--ease` | mousedown / keydown |
| **Focus (pill, ≥999px)** | `box-shadow: 0 0 0 2px --synapse, 0 0 0 5px --pollen-halo` | `--dur-fast` | `--ease` | focus-visible |
| **Focus (rounded, 8–16px)** | `box-shadow: --shadow-focus` (4 px pollen halo + 1 px synapse inset) | `--dur-fast` | `--ease` | focus-visible |
| **Focus (rectangular)** | `box-shadow: 0 0 0 4px --pollen-halo` (no inset) | `--dur-fast` | `--ease` | focus-visible |
| **Route enter** | `opacity 0 → 1` + `filter: blur(8px) → blur(0)` + `translateY(12px) → 0` | `--dur-slow` | `--ease` | hashchange |
| **Panel slide-in (right)** | `transform: translateX(24px) → 0` + `opacity 0 → 1` | `--dur-slow` | `--ease` | node click |
| **Panel slide-out (right)** | reverse | `--dur` | `--ease-in` | close / outside-click |
| **Thread draw (signature)** | `stroke-dashoffset 1 → 0` (SVG) or `transform: scaleX(0) → scaleX(1)` (CSS, `transform-origin: left`) | `--dur-slow` | `--ease` | completion, success, attention |
| **Pulse (idle heartbeat)** | `transform: scale(1 → 1.18)` + `opacity 0.85 → 1` | 4 s loop | `--ease` | always-on for nodes in `awaiting` / `listening` phase |
| **Pollen float (ambient)** | `translate(dx, dy) rotate(0 → 360deg)` + opacity 0 → 0.9 → 0 | 7–13 s loop | `linear` | empty-state decoration only |
| **Stop (streaming)** | fade "Send" → "Stopping…" 280 ms → fade to "Stopped" + 1× pollen pulse | `--dur` + `--dur-slow` | `--ease` | Stop click |
| **Save bar (Settings)** | `translateY(40px) → 0` + `opacity 0 → 1` | `--dur-slow` | `--ease-pop` | dirty = true |
| **Palette highlight slide** | `top: highlightTop` (transform: translateY) | `--dur` | `--ease` | ↑/↓ nav |

### The reduced-motion contract

`@media (prefers-reduced-motion: reduce)` is owned **exclusively** by
`condura.css`. A single block:

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
  .paper-grain, .mote, .ambient-thread { display: none; }
  .wordrise > span { transform: none; }
}
```

Components never re-declare their own reduced-motion media queries
(MOAT.md §2.3). The single block in `condura.css` owns the contract;
`IntersectionObserver` and `Page Visibility` listeners short-circuit on
`matchMedia` once at mount, never per-component.

### Energy budget (battery-aware)

`@media (prefers-reduced-motion)` and `data-energy="low"` (set by the
platform via the `internal/perception` energy mode) both disable
ambient pollen motes, paper grain, and the slow pulse — and snap
`--dur-cine` / `--dur-slow` to `0ms`. The product is still usable; it
just stops performing. Battery is the user, not the design.

---

### THE SIGNATURE ANIMATION: **The Thread**

A 1.25 px synapse hairline that **draws in from left to right** over
`--dur-slow` (520 ms), `cubic-bezier(0.22, 1, 0.36, 1)`. That is all.

```
SVG path:
  <path
    d="M 0 1 L 340 1"
    stroke="var(--synapse)"
    stroke-width="1.25"
    pathLength="1"
    vector-effect="non-scaling-stroke"
    stroke-dasharray="1"
    stroke-dashoffset="1"
    style="transition: stroke-dashoffset var(--dur-slow) var(--ease)"
  />
  /* set stroke-dashoffset → 0 to draw */

OR (CSS-only, no SVG):

  .thread {
    height: 1px;
    background: linear-gradient(90deg, transparent, var(--synapse) 20%, var(--synapse) 80%, transparent);
    transform: scaleX(0);
    transform-origin: left;
    transition: transform var(--dur-slow) var(--ease);
  }
  .thread.draw { transform: scaleX(1); }
```

**The Thread is the visual grammar the user learns once and sees
everywhere.** It already lives in:

- `TitlebarThread.svelte` — the titlebar hairline that bends toward
  the cursor.
- `Chat.svelte` — between message turns (`Thread` component), under the
  composer's focus state (`composer-card::before`, scaleX 0→1).
- `Ritual.svelte` — the breathing spine at the bottom of the wizard.
- `Channels.svelte` — under the audit-threadlink footer note.
- `About.svelte` — the section hairlines that draw in on scroll.
- `Sync.svelte` — the inter-device threads (sync-draw, 1.1 s).
- `Replay.svelte` — the scrubber's played-segment fill.
- `Hub.svelte` — the implied thread under "Installed ✓."

#### Where the Thread MUST appear (per surface)

| Surface | Thread placement | Trigger |
|---|---|---|
| Chat empty | under the hero copy, before the composer | mount |
| Chat mid-conversation | between every pair of turns | new message |
| Settings sections | under each section's h1 | mount + on save |
| Audit log | vertical spine + per-node hairline on click | mount + node click |
| Replay | scrubber fill (left of playhead) | scrub |
| Sync | inter-device threads (this-device ↔ peer) | pair complete |
| Hub spine | implied thread under the 3D bookshelf | install success |
| Channels | under the audit-threadlink footer | mount |
| Ritual Constellation | the bezier paths connecting wired nodes to center | node wired |
| Consent modal | the "armor" rect that draws in on the action summary | mount |

#### Where the Thread MUST NOT appear

- As a decorative flourish on a button hover. The thread is *meaning*,
  not *ornament*.
- On the titlebar wordmark area (that has the wordmark dot, which is
  its own gesture).
- As a divider between unrelated surfaces (use `--hair` instead).

#### Why the Thread, not a spinner or a check

The Thread says **a connection was made.** A spinner says **wait.**
A checkmark says **done.** The agent is not waiting; the agent is
*tying a thread to the user's request.* That is what makes it feel
alive: the moment of completion is the moment of connection.

This is the single most recognizable gesture in the product. A user
who sees the Thread once recognizes it everywhere. A new surface that
ships without a Thread is a surface that doesn't *finish* — it just
ends. Every surface must ship at least one Thread. Every error state
uses `err-hair` (the Thread variant that draws in after the headline).
Every completed action draws the Thread.

**Commit to it everywhere.**

---

## 6. THE ANTI-VIBE-CODED RULES

These are the seven rules that guarantee Condura does not look like a
weekend hack. Any surface that breaks one is wrong; any PR that lands
one is rejected. No exceptions, no "but it's just this once."

### Rule 1 · No gradient text.

The brand voice is paper and ink. Gradient text is a 2017-portfolio-site
tic that screams "I didn't have a brand so I made one mid-build." Text
is one color per role. Headlines are `--content` (or `--synapse` when
the word is the `.alive` accent). Sub-copy is `--content-soft`. Eyebrows
are `--content-faint`. Never a gradient between two of them. A background
can carry a gradient (the `--bloom-1` / `--bloom-2` / `--bloom-3`
radials on `.surface-paper`); text never does.

### Rule 2 · No emoji as UI.

Use `<Glyph name=...>` from `icons.ts`. The Glyph set is single-stroke,
1.5 px weight, currentColor, drawn in the brand line vocabulary. If a
surface needs an icon that isn't there, add it to `icons.ts` — three
new icons per quarter is healthy; thirty new emoji in a single route is
a smell. The only emoji allowed in the codebase are in user-facing copy
*about* emoji (e.g., a tooltip explaining how to type an emoji into
voice transcription). They are never the icon for a button, a nav item,
a status badge, or a state indicator.

### Rule 3 · No rainbow accents.

Brand has synapse (green) and pollen (orange). Status has `--ok`,
`--warn`, `--danger`, `--info`. There is no purple, no cyan, no teal,
no pink, no yellow outside of pollen. Adding a new color requires a
CLAUDE.md amendment. The constraint is the design. An open palette
reads as improvised; a constrained palette reads as designed.

### Rule 4 · One metaphor per component.

A component picks one metaphor and commits to it. The signal-bar
component in `Channels.svelte` uses cellular-bar dots (5 stepped
heights, 8/12/16/20/24 px) and only cellular-bar dots. The
constellation in `Ritual.svelte` uses nodes and bezier-thread paths and
only nodes and bezier-thread paths. The skill cards in `Skills.svelte`
use card lift + thread-stroke and only those two. A card with a
gradient header and a 3D tilt and a ribbon and a chip and a thread is
four metaphors arguing. Pick the one that the user reads in 200 ms and
remove the other three.

### Rule 5 · Shadows earn their weight.

There are exactly four shadow tokens: `--shadow-paper`, `--shadow-card`,
`--shadow-float`, `--shadow-focus`. Each is for one elevation tier. A
component never stacks two shadows. A component never adds a custom
`box-shadow` outside the four tokens. If a surface needs elevation that
none of the four provides, the answer is `--shadow-float`, not a fifth
ad-hoc shadow. The pressed-tactile `filter: brightness(0.95)` is not a
shadow — it's a *pressed-state* signal — and it is the only allowed
state-based filter. Hover-lifts use `--shadow-card → --shadow-float`
via the existing transition; they never add a third shadow on top.

### Rule 6 · Focus rings track the shape.

Pill elements (`border-radius ≥ 999px`) get a 2 px synapse ring + 5 px
pollen halo. Rounded surfaces (8–16 px radius) get the standard
`--shadow-focus` (4 px pollen halo + 1 px synapse inset). Rectangular
inputs (the EULA scrollable well, the keycap on hotkey capture) get
the 4 px halo with no inset line. **No element in the codebase may use
`outline: 1px solid var(--content)` or any other rectangular outline.**
Focus is the user's pointer; it must follow the geometry they are
touching, not flatten it into a rectangle.

### Rule 7 · Loading states teach, not spin.

A loading state is a verb; it tells the user what is happening. The
Thread drawing in says **"data is moving."** A hash-route pulse says
**"this is a quiet surface."** A spinning circle says **"I don't know
what's happening, so I'm just spinning."** No `<Spinner />` component
exists in the codebase. No three-dot variant. No `↻` glyph in a
loading slot. The Pulse component (`Pulse.svelte`) is the only allowed
loading indicator, and even it must be paired with a mono-uppercase
label that names the operation: `INDEXING…`, `PROBING REACH…`,
`READING THE CHAIN…`. The Pulse breathes; the label teaches.

---

## Closing note

This document is the voice. The five sister docs are the voice *applied:*

- `TEARDOWN.md` — what we stole, what we skipped, what we made our own.
- `MOAT.md` — the bar a premium product must pass.
- `APPFLOW.md` — every screen the user can land on, the data each
  writes, the animations each owns.
- `DIRECTION.md` (this) — the personality, the type, the color, the
  motion, the rules.

When the four disagree, the disagreement is the bug. Fix the code or
fix the doc — but do it in the same commit, and never by silently
rewriting one to match the other.

**The single test for every new surface:**

> *Does this read like a paper notebook that learned to listen — warm,
> awake, and never louder than the room it's in?*

If yes, ship it. If no, it isn't finished yet.