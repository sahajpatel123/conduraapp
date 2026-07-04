# SCREEN_TITLEBAR.md — Condura Titlebar · Screen Architecture

> **The contract for `Titlebar.svelte`.** This is the top strip of the Condura Shell — the single most visible surface in the product, the home of the **Thread signature**, and the place where every "this is now finished" moment draws its hairline.
>
> **Source-of-truth files** (read alongside this spec):
> - `app/web/frontend/src/lib/condura/TitlebarThread.svelte` — the existing implementation; this spec codifies it.
> - `app/web/frontend/src/lib/condura/Shell.svelte` — the titlebar sits at `grid-area: titlebar`, row 1 of the shell grid.
> - `app/web/frontend/src/lib/condura/specs/SCREEN_SHELL.md` §1, §2.1 — the parent spec.
> - `app/web/frontend/src/lib/condura/MOAT.md` §3 — **THE SIGNATURE**.
> - `app/web/frontend/src/lib/condura/DIRECTION.md` §5 — the motion grammar, the Thread.
>
> **Reading order.** §1 (Layout) tells you the geometry. §3 (States) tells you what each phase reads as. §4 (Motion) is the heart of this document — the Thread is here. §8 (Decisions) is the MOAT binding.

---

## 1. LAYOUT

### 1.1 Geometry

The Titlebar is the **first row of the Shell grid** — a 44 px (≈48 px including the 2 px inset hairline below) full-bleed horizontal strip that spans columns 1–3 of the shell's `grid-template-areas: "titlebar titlebar titlebar"`. It is the only place in the product where the **drag region meets interactive chrome**: the entire strip forwards drag-to-move to the OS window manager (Wails / Tauri / browser-drag-region), except where the right-zone buttons claim a click surface.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ (L) WORDMARK                (C) THREAD       (R) ISLAND · ⌘K · ☼ · ◌▢✕       │
│   Condura · Chat              ──•──                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│   hairline · 1 px · var(--hair) · inset 0 across the bottom                 │
└─────────────────────────────────────────────────────────────────────────────┘
        64                160                flex                 240      px
```

The horizontal geometry is a **three-zone grid** (`grid-template-columns: auto 1fr auto`), left-aligned flush at `padding-inline: var(--space-5)` on screens ≥1024 px and `--space-3` on the 768–1023 rail-collapsed breakpoint. On macOS the first 78 px are reserved for the traffic-light buttons (close / min / max) supplied by the OS; the Wordmark sits to their right with `padding-left: var(--wails-traffic-light-w, 78px)`.

### 1.2 Three-zone placement

| Zone | grid area | Width | Contents |
|------|-----------|-------|----------|
| **L · Identity** | column 1 (`auto`, ~280 px) | The Condura wordmark + the active **route context** (e.g., `· Chat`, `· Settings`). Inline, never re-mounts across navigation. |
| **C · Thread** | column 2 (`1fr`, fills) | The `<TitlebarThread />` SVG, positioned absolute, `left: 160px right: 200px top: 0 height: 100%`, `pointer-events: none`. The thread is the only content here. |
| **R · Status & Controls** | column 3 (`auto`, ~240 px) | DynamicIsland (status pill) · ⌘K hint chip · theme toggle · (macOS window controls in the OS-supplied area outside our grid). `margin-left: auto`. |

The `<TitlebarThread />` is `position: absolute` *inside* the C zone rather than being a grid child so it can extend across the eye-line between L and R without participating in their flex layout. `pointer-events: none` keeps the drag region intact: clicks anywhere in C pass through to the parent drag region.

### 1.3 Drag region

The entire Titlebar except the R-zone buttons has `pointer-events: none` on the gesture elements (the wordmark, the route-context text, the SVG). The parent `.titlebar` has `pointer-events: auto` so the OS can read the drag gesture from the underlying area. This is the inverse of the Shell's rest: the titlebar is not interactive in its middle by design.

```css
.titlebar { -webkit-app-region: drag; app-region: drag; }
.titlebar button, .titlebar [role='button'], .titlebar .tb-island { -webkit-app-region: no-drag; app-region: no-drag; }
```

### 1.4 Z-stack

| z-index | Layer |
|--------|-------|
| `var(--z-sticky)` (100) | the titlebar itself |
| `(inside titlebar)` | wordmark (bottom) → SVG thread + glow + node (middle) → right-zone buttons (top) |
| below titlebar | the 1 px `--hair` bottom border (drawn by `.titlebar::after` so the thread's bend can pass *over* the hairline on extension) |

### 1.5 Breakpoints

| Breakpoint | Titlebar | Wordmark | ⌘K hint | Window controls |
|---|---|---|---|---|
| **≥ 1440 px** | 44 px, three-zone | 22 px Instrument Serif | visible | OS-rendered |
| **1024–1439 px** | 44 px, three-zone | 22 px | visible | OS-rendered |
| **768–1023 px** | 44 px, three-zone | 18 px | hidden (moved to palette tooltip) | OS-rendered |
| **< 768 px** | collapsed into a top touch-bar (32 px) + status pill into a Drawer card | hidden (in hamburger) | hidden | hidden (system handles) |

---

## 2. CONTENT SLOTS

### 2.1 L · Identity (`grid-column: 1`)

| Slot | Component | Content | Token |
|------|-----------|---------|-------|
| Wordmark | `<div class="tb-wordmark">` | `Condura` followed by a 5 px `--pollen` dot accent (the `.alive` punctuation — the only allowed use on this surface). | `var(--font-display)`, 22 px, `--ls-display`, `var(--content)`. |
| Route context | `<span class="tb-route">` | `· Chat` · `· Hub` · `· Skills` · `· Sync` · `· Audit` · `· Replay` · `· Channels` · `· Delegation` · `· Settings` · `· About`. Reads from the active route. | JetBrains Mono, 11 px, `0.12em` tracking, uppercase, `var(--content-faint)`. |

The route context is **never bolded, never colored**. It is a whisper that tells the user which surface they have landed on; the synapse pillar in the NavRail is what says *you are here*. The wordmark carries the brand; the route context carries the location.

The pollen dot is a `border-radius: 50%` 5 px square with `box-shadow: 0 0 8px color-mix(in oklab, var(--pollen) 60%, transparent)`. It is not an emoji, not an `<Glyph>`, not a CSS pseudo-element — it is the only allowed decorative flourish on the wordmark, MOAT §4.2.

### 2.2 C · Thread (`grid-column: 2`)

| Slot | Component | Content |
|------|-----------|---------|
| Synapse stroke | `<path bind:this={line} d="M 0 22 L 9999 22">` | 1.25 px stroke, `var(--synapse)`, dasharray `6 8`, `stroke-linecap: round`. Anchored to `y = 22` (the vertical midline of the 44-tall active strip). |
| Synapse glow | `<path bind:this={glow}>` | 3 px `var(--synapse-glow)` underlay at `opacity: 0.18`, `filter: blur(3px)`. Gives the thread its breath when the cursor is far from the bend anchor. |
| Pollen node | `<circle bind:this={node} r=3>` | `fill: var(--pollen)`, `filter: drop-shadow(0 0 6px color-mix(in oklab, var(--pollen) 70%, transparent))`. Halves to `r=1.5` and lifts `opacity: 0` on `prefers-reduced-motion`. |

The SVG sits `position: absolute; left: 160px; right: 200px; top: 0; height: 100%; width: auto; overflow: visible`. The bend anchor is `(cx * width, mid + (cy * H - mid) * 0.5)` clamped to `[6, H-6]`. The bend is **subtle**: at rest the path is straight; on `pointermove` only the local curvature eases toward the cursor at a 0.06 lerp per rAF tick.

`prefers-reduced-motion: reduce` short-circuits the rAF loop on mount — the path renders as a flat dashed line, the dot disappears, the SVG is still aria-hidden.

### 2.3 R · Status & Controls (`grid-column: 3`)

| # | Slot | Component | Content |
|---|------|-----------|---------|
| 1 | Status pill | `<DynamicIsland phase={agentPhase} task={currentTitle} aria-live="polite">` | A morphing capsule. Width 124–188 px depending on phase. Mono-uppercase label (`letter-spacing: 0.12em`). See §3 for the 6 state labels verbatim. |
| 2 | ⌘K hint | `<span class="kbd-hint">` with two `<kbd>` chips | `⌘ K` (mac) / `Ctrl K` (win/linux). Mono 11 px uppercase 0.12em, `var(--content-faint)`. **Decorative.** The chord does the work. |
| 3 | Theme toggle | `<Button kind="icon" aria-label="Toggle theme">` wrapping `<Glyph name={theme === 'dark' ? 'sun' : 'moon'} />` | 28×28 round (pill, `border-radius: 9999px`), `var(--surface-card)` fill, 1 px hairline border. Hover: rotate(-12deg), border-color `var(--hair-strong)`. Press: scale(0.97) + `filter: brightness(0.95) saturate(1.1)` per MOAT §2.2. |
| 4 | Window controls | OS-supplied (mac traffic-light area) | Out of our DOM on macOS. Empty flex-box on win/linux reserved for future in-app controls. |

The hint chip and theme toggle share a flex cluster (`.tb-controls`) with `gap: var(--space-3)` and `margin-left: auto`. Tooltip-wrapped per MOAT §2.9 (since these are icon buttons and native `title=` is forbidden for read-critical affordances).

---

## 3. STATE MATRIX

The Titlebar has two layers of state — the **Thread's own state** (governed entirely by pointer position + visibility + reduced-motion) and the **DynamicIsland's phase state** (governed by the agent store, the consent store, and the halt store). The Shell's `phase` is the single source of truth for the island.

### 3.1 Default (agent idle, thread idle)

- **Thread**: dashes alone, the bezier rests flat at `(cx=0.5, cy=0.5)`; the pollen node sits at the geometric center.
- **DynamicIsland**: width 124 px, label `IDLE · LISTENING`. Color `var(--synapse)`. One small inline `<Pulse size=6>` to the left of the label (or right of it — design lands on the left, so the eye reads pulse → label).
- **⌘K hint**: visible, `var(--content-faint)`.
- **Theme toggle**: visible, default rotation.
- **Drag region**: enabled.

### 3.2 Thinking (the model is reasoning)

Trigger: `conversation.isThinking === true && !conversation.isStreaming`.

- **Thread**: the pollen node's `r` attribute transitions from 3 → 6 px (visibly larger cursor footprint in the bend path). Halo lifts to 6 px drop-shadow. The thread itself does not change color.
- **DynamicIsland**: widens to 156 px. Label `THINKING · <truncated task, 16 chars>`. Border `var(--synapse)`, label `var(--content)`. An inline `<Pulse phase="thinking">` sits to the left of the label, breathing at the 4 s cadence.
- **aria-live**: `aria-live="polite"` announces `Now thinking about <task>` on transition into thinking.

### 3.3 Streaming (tokens arriving)

Trigger: `conversation.isStreaming === true`.

- **Thread**: a **second hairline draws** under the primary thread — the `progress-hairline`. 1 px `var(--synapse-glow)` at `opacity: 0.4`, `stroke-dasharray: 12 8`, animated `stroke-dashoffset` left-to-right over 1.6 s linear infinite. This is the only place a continuous loop is allowed in the titlebar, and only while tokens stream. The hairline disappears when streaming ends.
- **DynamicIsland**: widens to 188 px. Label `STREAMING · <truncated task, 18 chars>`. Border `var(--synapse)`. The `<Pulse>` moves to `phase="streaming"` (faster 1.6 s breathe). A token counter — `↓ 1.2k ↑ 380` — sits to the right of the label, mono 10 px `var(--content-mute)`.
- **aria-live**: `polite`, update interval clamped to 2 s so screen readers do not chat through the response.

### 3.4 Consent pending (Gatekeeper blocked a non-READ action)

Trigger: `consent.ticket !== null` (store polls every 1.2 s).

- **Thread**: bends harder — the **bend anchor weight** jumps from 0.06 to 0.12 lerp. The cursor's pull becomes visible.
- **DynamicIsland**: widens to 188 px, **border `var(--warn)`**, **fill `var(--pollen-light)` at 35 %**, label `CONSENT REQUIRED`. The Pulse moves to `phase="consent"` (pollen-tinted, 1.0 s breathe). A small ⓘ glyph prefix (Unicode `U+24D8`, our only allowed Unicode glyph per §8 — see decisions §8.4) tells the user the island is openable.
- **Click behavior**: `DynamicIsland` becomes a button that calls `consent.focus()` — moves focus to the (always-mounted) `<ConsentModal />` without opening a new modal.
- **aria-live**: `assertive`. Screen readers hear `Consent required` immediately on transition.

### 3.5 Kill switch armed (`Cmd+Shift+Escape` or watchdog tripped)

Trigger: `halt.state.halted === true`.

- **Thread**: the synapse stroke **hardens to `var(--danger)`** (the only time it is not synapse). The pollen node drops to `r=2` and stops moving — the gesture is **"the organism stopped."**
- **DynamicIsland**: `phase="error"`. Border `var(--danger)`, fill `var(--danger)` at 8 %, label `HALTED · <reason>`. Full-pulse breathe at 2.0 s. No glyph prefix.
- **The Thread is not paused on this state.** The titlebar remains paintable behind the (always-mounted) `<KillSwitchOverlay />` scrim; the island is the persistent affordance.
- **aria-live**: `assertive`. Announces once: `Agent halted: <reason>. Mint a resume ticket to continue.`

### 3.6 Offline (daemon unreachable)

Trigger: `daemon.connected === false`.

- **Thread**: stays synapse. The pollen node keeps drifting at its slow natural cadence (the page is alive even when the daemon is not — per MOAT §3, the signature never goes dead).
- **DynamicIsland**: `phase="error"` but **without** the danger fill — island border `var(--ink-mute)`, fill `var(--surface-card)`, label `OFFLINE`. Static, no pulse. (The island is muted, not alarmed; offline is recoverable in 1 click of a re-launch.)
- **aria-live**: `polite`. Announces on transition `Daemon offline. Reconnecting…` once.

### 3.7 Reduced motion (system override)

When `matchMedia('(prefers-reduced-motion: reduce)').matches` is `true`:

- The Thread's rAF loop never starts; the path renders as a flat dashed horizontal at `y=22`.
- The pollen node renders at `opacity: 0`.
- The progress hairline (during streaming) renders as a static pill instead of a moving dasharray.
- The DynamicIsland phases still apply (color + width + label) but the breathing pulses are disabled via the single block in `condura.css`.

---

## 4. MOTION CHOREOGRAPHY

The Titlebar owns the **single most recognizable gesture in the product**: the bending Thread. This section is the contract for that gesture, plus the four smaller motions that surround it.

### 4.1 The Thread (the signature)

The Thread is the visual grammar the user learns once and sees everywhere (DIRECTION.md §5). It is a 1.25 px synapse hairline that **draws in from left to right** under `--dur-slow` (520 ms) and **bends toward the cursor** in real time, with a low-pass-filter (0.06 lerp per rAF tick) that smooths the gesture into something organic rather than twitchy.

The rAF loop touches **three SVG attributes per frame**: `line.setAttribute('d', …)`, `glow.setAttribute('d', …)`, `node.setAttribute('cx', …)` + `node.setAttribute('cy', …)`. That is the entire CPU budget — verified in `TitlebarThread.svelte:32–46`.

```ts
// snippet from TitlebarThread.svelte
const bend = () => {
  cx += (px - cx) * 0.06;            // low-pass filter
  cy += (py - cy) * 0.06;
  const x = cx * W;
  const y = Math.max(6, Math.min(H - 6, cy * H));
  const d = `M 0 ${mid} C ${W*0.25} ${mid - (mid-y)*0.7}, ${W*0.42} ${y}, ${W*0.5} ${y} S ${W*0.78} ${mid + (y-mid)*0.4}, ${W} ${mid}`;
  line.setAttribute('d', d);
  glow.setAttribute('d', d);
  node.setAttribute('cx', String(x));
  node.setAttribute('cy', String(y));
  raf = running ? requestAnimationFrame(bend) : 0;
};
```

The loop is **disabled** in three conditions:

1. `prefers-reduced-motion: reduce` — early return on mount; the line renders as a static dashed line and the dot disappears (see §3.7).
2. `document.hidden === true` — `visibilitychange` flips `running` to `false` and zeroes the rAF. A background tab spends 0 CPU on the thread.
3. The host element is scrolled out of view — `IntersectionObserver` (threshold 0) flips `inView` to `false` and zeroes the rAF. Scrolling to Chat's stream or to a long Settings page keeps the loop dormant.

On resume from any condition the loop re-arms inside the next rAF tick; no flash, no rebuild.

### 4.2 The progress hairline (streaming-only)

Trigger: `conversation.isStreaming`.

A second SVG path (sibling to the primary thread) renders under it: 1 px `var(--synapse-glow)`, opacity 0.4, dasharray `12 8`, animated `stroke-dashoffset` linearly left-to-right over 1.6 s. This is the only allowed **continuous loop** in the titlebar, and it is bounded — it disables when `isStreaming` flips to false.

`prefers-reduced-motion`: the hairline renders as a **static fill** at full opacity rather than a moving dasharray. The activity still reads; the motion is removed.

### 4.3 The wordmark draw (first paint)

On first mount of the titlebar after Ritual completes, the wordmark `Condura` **draws itself in** over `--dur-cine` (900 ms):

1. The text is masked by `clip-path: inset(0 100% 0 0)` at t=0.
2. The clip-path eases to `inset(0 0 0 0)` over 900 ms `--ease`.
3. Simultaneously, the **synapse underline** (a 1.25 px `var(--synapse)` rectangle under the wordmark, height 2 px) draws in via the same stroke-dasharray recipe as the Thread (`pathLength: 1; stroke-dasharray: 1; stroke-dashoffset: 1 → 0`).
4. At t=900 ms, the pollen dot fades in (`opacity 0 → 1` over 240 ms `--ease`).

On every subsequent navigation (route change), only the route context text fades (`opacity 0 → 1` over 200 ms) — the wordmark is settled.

`prefers-reduced-motion`: the wordmark renders fully opaque immediately. The underline is a static 1 px hair. The pollen dot is present from t=0.

### 4.4 ⌘K hint fade-in

The `⌘K` hint chip fades in 600 ms after the titlebar mount completes (`opacity 0 → 1` over 200 ms `--ease`). The 600 ms delay exists so the user's eye is not pulled away from the wordmark in the first second; once they have registered the brand, the meta-UI announces itself.

### 4.5 DynamicIsland morphing

Each phase transition uses **two** morphs:

1. **Width** — a `transition: width 320ms var(--ease)` on the island shell. 124 → 156 → 188 px depending on phase.
2. **Color** — a `transition: border-color var(--dur) var(--ease), background-color var(--dur) var(--ease), color var(--dur) var(--ease)` triple on the label, border, and fill.

The morph is **synchronous** — both 320 ms transitions start at the same paint and complete at the same paint. No interleaved fade-fade-fade. The island **never** scales (no `transform: scale()`); width is the only dimension that animates.

### 4.6 Theme toggle rotation

Hover: `transform: rotate(-12deg)` over `--dur-fast` (140 ms) `--ease`. The rotation tells the user the icon is a switch (sun ↔ moon) without animating the swap — when they click, the icon crossfades over `--dur` (280 ms) to the *opposite* glyph, and the rotation settles back to 0 over `--dur-fast`.

Press: inherits `.tactile` — `scale(0.97) + filter: brightness(0.95) saturate(1.1) + translateY(0.5px)` per MOAT §2.2. No component-local override.

### 4.7 The Thread-drawn "this is now finished" reinforcement

The Titlebar is **the home of the Thread**, and therefore the place where every **shell-level completion** draws its reinforcement. Five events:

1. Route entered (`hashchange` to a known route) — the underline beneath the wordmark redraws (`stroke-dashoffset: 1 → 0` over `--dur-slow`).
2. Stream start (`isStreaming` flips true) — the synapse stroke brightens (`opacity 0.55 → 0.85`) over `--dur`.
3. Stream end (`isStreaming` flips false) — the progress hairline draws **right-to-left** (`stroke-dashoffset: 0 → 1` reverse) over `--dur-slow`. The gesture of *yielding*.
4. Consent granted (`consent.ticket === null`) — the synapse glow lifts to `opacity 0.32` for 900 ms then settles back.
5. Halt cleared (`!halt.state.halted` after being true) — the synapse stroke redraws left-to-right over `--dur-slow` (the breath returns).

These are the only five shell-level completions the Titlebar owns. Route-internal completions draw their own Threads.

---

## 5. KEYBOARD

The Titlebar's keyboard story is minimal by design: **no element inside the center zone is interactive**, and the right-zone elements participate in the Shell's normal Tab order (DOC order = visual order here).

### 5.1 Tab order (no modal open, no focused overlay)

The Titlebar contributes **zero** items to the first tab stop. The first interactive element below the document is the **theme toggle** (zone R slot 3), followed by the **⌘K hint** — but the ⌘K hint is not a button, it is a decorative chip; `Tab` skips it.

Actual Tab order in the right zone, left → right:

1. **DynamicIsland** (when `phase === 'consent'` only — it is `tabindex="0"` then; otherwise `tabindex="-1"`).
2. **Theme toggle** (always `tabindex="0"`).

The macOS window controls (close / min / max) are OS-owned and are reached by Tab as the OS exposes them — out of our DOM.

### 5.2 No drag-region focus

The center zone and wordmark area are not focusable. The drag region forwards pointer events but does not claim keyboard focus; there is **no** "focus the thread" affordance, ever. (A focus ring drawn around the thread would betray what the thread is — a gesture, not an element.)

### 5.3 Global chords that pass through the Titlebar

The Titlebar does not own any global chords. The chords it visually advertises (the ⌘K hint, the sun/moon chord `⇧T`) are bound to **window-level** keydown handlers registered in `Shell.svelte`. The Titlebar's job is to show the affordance, not to handle it.

The `?` chord (open Shortcuts overlay) does not originate in the Titlebar; it is bound at the Shell level. The Titlebar contributes the hint visually through the ⌘K chip's tooltip.

### 5.4 DynamicIsland as the shell-level `aria-live` surface

When the agent transitions between phases, screen readers are notified through the island:

- `aria-live="polite"` for idle / thinking / streaming / offline transitions (debounced to 2 s while streaming).
- `aria-live="assertive"` for consent-pending and kill-switch-armed transitions (immediate).

The label text inside the island is the announcement text. `aria-atomic="true"` so the full label (not just the changed portion) is announced.

---

## 6. COMPONENTS USED

| Component | Where used | Role in Titlebar |
|-----------|------------|------------------|
| `TitlebarThread.svelte` | center zone (C) | The Thread signature — 1.25 px synapse stroke + 3 px synapse-glow underlay + pollen node, bending via rAF toward the cursor with a 0.06 lerp low-pass filter. |
| `DynamicIsland.svelte` | right zone slot 1 | The morphing status pill. Reads `phase` (idle / thinking / streaming / consent / error) + `task` (current title for streaming). aria-live region. |
| `Glyph.svelte` | inside the theme toggle | The `sun` and `moon` icons. 1.5 weight, currentColor, single-stroke. |
| `Button.svelte` (`kind="icon"`) | theme toggle | Wraps the Glyph, provides the tactile press vocabulary, tooltip support. |
| `Tooltip.svelte` | wrapping theme toggle + ⌘K hint | Hover-delay 400 ms, exit 75 ms, `aria-describedby`. Replaces the `title=` attribute that MOAT §2.9 forbids. |
| `Pulse.svelte` | inside DynamicIsland, for thinking / streaming / consent / kill phases | The breathing indicator. Phase prop maps to the four tinted states. |
| `<kbd>` (inline) | inside the ⌘K hint | JetBrains Mono 11 px, 6 px radius, paper-card background, 1 px hair-strong border. DIRECTION.md §3 "kbd hint style." |
| `.tb-route` (inline) | left zone (after the wordmark) | Reads `currentRoute` from the Shell's hash router; renders `· <route-cap>` in mono uppercase. |
| (OS-rendered) Window controls | far right | macOS traffic lights. Out of our DOM. |

No third-party dependencies. No new components created *for* the Titlebar — every part composes from primitives that already exist in `lib/condura/`.

---

## 7. DATA FETCHED

The Titlebar is **read-only**. It pulls from three sources; it writes to one.

| Source | What is read | When |
|--------|--------------|------|
| `agent` / `conversation` store (`lib/stores/conversation.svelte.ts`) | `agentPhase` (`idle` / `thinking` / `streaming`), `currentTitle` (the active task name for streaming label) | Reactive subscription — every store flip re-evaluates the DynamicIsland's `phase` + label. |
| `consent` store | `consent.ticket !== null` | Polled every 1.2 s; flips the island into `phase='consent'`. |
| `halt` store | `halt.state.halted`, `halt.state.reason` | Polled every 1 s; flips the island into `phase='error'` with `HALTED · <reason>`. |
| `daemon` store | `daemon.connected` | Polled alongside the halt cadence; flips the island into the **muted error** state (§3.6). |
| `theme` (localStorage `condura.theme`) | `'light'` / `'dark'` | Read on mount + on every `storage` event (cross-tab sync). Drives the sun/moon glyph choice. |

Writes:

| What is written | When | Where |
|-----------------|------|-------|
| `theme` to localStorage | On theme-toggle click | `localStorage.setItem('condura.theme', next)` + a `storage` event so other open Condura tabs flip in lockstep. |

The Titlebar **does not** read the audit log, the channels store, the skills store, the delegation store, or the routes store. Those are owned by Main region routes. The Titlebar's job is to broadcast the agent's *current phase* to the chrome and nothing else.

---

## 8. DESIGN DECISIONS — MOAT compliance

### 8.1 The Titlebar hosts THE SIGNATURE

The Thread is the **one unmistakable thing** in Condura (MOAT §3). The Titlebar is its home. The contract:

1. **The Thread is always running when the Titlebar is mounted.** Even when the agent is offline, even when the user is mid-task, even when the page is hidden. `visibilitychange` + `IntersectionObserver` zero the CPU but the visual reading is still *"the organism is alive."*
2. **The Thread bends toward the cursor.** The bend is a low-pass-filtered 0.06 lerp — slow enough to read as organic, fast enough to react. Three SVG attributes per frame; budget verified.
3. **The Thread re-draws on every "this is now finished" moment.** Route entry, stream start, stream end (yielding), consent granted, halt cleared — five shell-level completions (per §4.7). The CSS is one rule repeated everywhere the Thread earns its appearance.

```css
/* the one rule, used in five places on this surface */
.tb-thread-stroke,
.tb-thread-reinforce,
.tb-progress-hairline {
  stroke: var(--synapse);
  pathLength: 1;
  stroke-dasharray: 1;
  stroke-dashoffset: 1;
  transition: stroke-dashoffset var(--dur-slow) var(--ease);
}
/* set stroke-dashoffset → 0 to draw; → 1 to retract */
```

### 8.2 The $50M feel (MOAT §5)

| Move | Where in the Titlebar |
|------|-----------------------|
| §5.2 — Composer-focus inward-thread (re-used) | The Titlebar is not a composer, but the same `::before` recipe is used as the **wordmark underline** (drawn under `Condura` on every route entry). The user learns one gesture for "this surface is the current surface" — it draws in. |
| §5.3 — Stop yields the page (not just the button) | When `conversation.isStreaming` flips false, the progress hairline **draws right-to-left** (`stroke-dashoffset: 0 → 1` reverse) over `--dur-slow`. The user sees the agent yielding from the titlebar itself, not from a button click in Chat. |
| §5.5 — Mobile ritual (responsive) | < 768 px collapses the titlebar to 32 px and moves the Status pill into a Drawer card. The wordmark drops out of the way so the status pill has the full width. The thread dims to 0.2 opacity at narrow widths (still draws, but quiet). |

### 8.3 Anti-patterns avoided (MOAT §4 — "What We Will Not Do")

| # | Anti-pattern | How the Titlebar avoids it |
|---|--------------|----------------------------|
| 1 | No gradient text | Wordmark is `var(--content)`, no `background-clip: text`. Route context is `var(--content-faint)`. |
| 2 | No emoji as UI icons | The pollen dot in the wordmark is a 5 px `border-radius: 50%`, not an emoji. The theme toggle is a `<Glyph>`. The ⓘ prefix in §3.4 is Unicode `U+24D8`, not emoji. |
| 3 | No glassmorphism unless earned | The titlebar surface is `var(--surface)` flat. No `backdrop-filter`. Elevation is supplied by the 1 px `--hair` bottom border only. |
| 4 | No rainbow accents | Status colors are `--ok` / `--warn` / `--danger` / `--info`. Brand is `--synapse` + `--pollen`. The kill-switch state is the only `--danger`-on-titlebar moment, and it is **earned** by the halt. |
| 5 | No "Welcome to the future" copy | The island reads literal ambient truth: `IDLE · LISTENING` / `THINKING · <task>` / `STREAMING · <task>` / `CONSENT REQUIRED` / `HALTED · <reason>` / `OFFLINE`. No celebration, no "Ready!" toasts. |
| 6 | No fake enthusiasm | The Thread is the only flourish, and it is earned (cursor-position + completion-events), not performed. No decoy animations. |
| 7 | No spinner loaders | No `<Spinner />` import anywhere in the titlebar. The streaming state uses a Thread hairline, not a circle. The thinking state uses the `<Pulse>` breathing (with a label), not a spin. |
| 8 | No rectangular focus outlines | Default focus is `--shadow-focus` (rounded 4 px halo + 1 px synapse inset). The theme toggle (pill) uses `border-radius: 9999px` so the synapse-ring-only focus rule applies. ⌘K hint is `tabindex="-1"`, so it never gets a ring. |
| 9 | No double shadows | The 1 px bottom hairline is `--hair` (one shadow token), the pollen dot's halo is `box-shadow: 0 0 8px …`, and the focus rings are tokens — three separated layers, none stacked. No `box-shadow` layered on the titlebar surface itself. |
| 10 | No animation that doesn't carry meaning | Every animation in the titlebar answers "what does this communicate?": thread bend = *alive, pointer-tracked*; progress hairline = *data is moving*; DynamicIsland width morph = *phase changed*; theme rotate = *I am a switch*; wordmark draw = *you arrived here*. No decorative loops (streaming's progress hairline is the only continuous loop, and it carries the meaning). |

### 8.4 Constraints locked

These decisions are **not re-opened without a CLAUDE.md amendment**, per MOAT §4 rule 4:

1. Titlebar height is **44 px** (with a 1 px `--hair` border below; total visual mass 45 px). Matched to `Shell.svelte` `--titlebar-h` and to `TitlebarThread.svelte:35` (`const H = 44`).
2. The pollen dot in the wordmark is the **only** allowed flourish on the wordmark (MOAT §1.7 — one `.alive` per surface, here applied to the dot).
3. The `ⓘ` Unicode glyph used in §3.4 (consent state) is the **only** Unicode-glyph-as-UI allowed on the titlebar. Every other icon is a `<Glyph>` from `icons.ts`. (The ⓘ is allowed because the consent state needs a single static indicator that the island is *interactive* in that phase, and the Glyph set has no `info` icon — adding one is a separate decision.)
4. The DynamicIsland's six phase labels (`IDLE · LISTENING`, `THINKING · <task>`, `STREAMING · <task>`, `CONSENT REQUIRED`, `HALTED · <reason>`, `OFFLINE`) are **exact copy** — do not paraphrase in the implementation. They were chosen for length (under 24 chars for the longest at 188 px) and tone (mono-uppercase, declarative, no exclamation).
5. The five shell-level completion draws (§4.7) are the **only** events that re-trigger the Thread reinforcement. Do not add a sixth.
6. The bend-anchor lerp coefficient is `0.06` (low-pass filter). Anything stronger reads as twitchy; anything weaker reads as sluggish. Tuned in `TitlebarThread.svelte:33`; do not tune per-component.

### 8.5 What the Titlebar is uniquely responsible for

- **Being the home of the Thread.** Per MOAT §3, *"a premium product earns one element. This is ours."* The Thread lives here. Every other surface in Condura *uses* the Thread (under chat messages, between Section h1s in Settings, along the audit chain), but the Titlebar is where the Thread *is*.
- **Hosting the shell's `aria-live` truth.** The island is the screen reader's single point of contact for *"what is the agent doing right now."* Every other surface either listens to it or ignores it. No other surface *is* it.
- **Carrying the drag region.** The titlebar doubles as the OS's "drag here" affordance. The drag region in C and L must not have any clickable element inside it; only the R zone can claim clicks.

### 8.6 What the Titlebar explicitly does NOT do

- **Cancel a stream.** Stop is Chat's affordance. The Titlebar shows the result via `isStreaming` flipping false.
- **Open Settings.** That's the shell-level `⌘,` chord.
- **Approve / deny consent.** That's `<ConsentModal />`'s job. The island *announces* the pending consent; the modal handles it.
- **Manage providers, edit autonomy, configure power, or any settings write.** Settings is the route for that.
- **Render an emoji.** See §8.4 rule 3.

---

## 9. Versioning + change-control

- **Spec version:** 0.1.0 — pinned to the v0.1.0 `TitlebarThread.svelte` (the SVG component) and to v0.1.0 `DynamicIsland.svelte`.
- **Update policy:** this document is **append-only in spirit** per CLAUDE.md §30.5. Corrections are not edits; add an entry below.
- **Implementation divergences:** when the implementation disagrees with the spec, fix the implementation. When the implementation has a reason the spec doesn't cover, add a row to §9.1.
- **Phase 4 hand-off:** §1–§8 are implementable as-is. The existing `TitlebarThread.svelte` already satisfies §4.1 (the bend loop) and §4.2 (the glow + node). Phase 4 needs to add: the wordmark draw (§4.3), the ⌘K hint fade-in (§4.4), the DynamicIsland morphing (§4.5), and the five shell-level completion draws (§4.7).

### 9.1 Open questions for the next session

- **Voice indicator** — when voice listening is active, should the DynamicIsland surface `LISTENING · <input>` or should the pollen node's halo amplify (per SHELL.md §2.1)? Currently both are drafted; pick one.
- **Tab order on the ⌘K hint** — currently `tabindex="-1"`. Should it become a real button that opens the CommandPalette on `Enter`? The brief leaves this ambiguous.
- **The progress hairline's reverse yield** — under `prefers-reduced-motion`, the hairline is a static fill at full opacity. Under streaming, it should also dim when the user is far from the titlebar (energy budget). Confirm or simplify.
- **`ⓘ` in the consent state** — keep Unicode `U+24D8` (per §8.4), or add an `info` glyph to `icons.ts` and replace it? Adding the glyph is a two-line change; keeping the Unicode is one less icon to maintain.

---

**This document is the architecture. The code is the implementation. They agree. When they diverge, the divergence is the spec-bug — fix the doc, then fix the code, in one commit.** (APPFLOW.md closing note, applies here too.)
