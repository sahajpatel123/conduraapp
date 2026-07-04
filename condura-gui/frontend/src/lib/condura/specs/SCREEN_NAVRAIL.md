# SCREEN_NAVRAIL · Condura

> **Screen architecture spec for `NavRail.svelte`.** Phase 4 implements from this
> document; it is the contract between design intent and shipping code. The
> structural skeleton is laid out here, in full, before any CSS lands.
>
> **Reading order for the next agent.** Read §1 (Drift) — it tells you what
> changes. Read §2 (Layout) — the geometry. Read §3 (State Matrix) and §4
> (Motion) — the behavior. Skip §5–§8 only if you already know the app.
>
> **Source-of-truth docs** (read before this one):
> - `MOAT.md` — quality bar. The NavRail must pass §1, §3, §4.
> - `APPFLOW.md` §3, §5 — route inventory and hash mapping.
> - `TEARDOWN.md` §1–§8 — mature 2026 sidebar patterns.
>
> **Existing implementation.** `NavRail.svelte` (current file) is the baseline;
> this spec describes the Phase 4 redesign that replaces it.

---

## Table of Contents

1. [Spec vs. Implementation Drift](#1-spec-vs-implementation-drift)
2. [Layout](#2-layout)
3. [Content Slots](#3-content-slots)
4. [State Matrix](#4-state-matrix)
5. [Motion Choreography](#5-motion-choreography)
6. [Keyboard](#6-keyboard)
7. [Components Used](#7-components-used)
8. [Data Fetched](#8-data-fetched)
9. [Design Decisions (MOAT Compliance)](#9-design-decisions-moat-compliance)
10. [Accessibility Contract](#10-accessibility-contract)
11. [Implementation Notes for Phase 4](#11-implementation-notes-for-phase-4)
12. [Test Plan](#12-test-plan)

---

## 1. Spec vs. Implementation Drift

What this spec changes in the current `NavRail.svelte`. Phase 4 must remove the
old, ship the new, in one atomic diff — no half-states.

| # | Today (current `NavRail.svelte`) | Phase 4 (this spec) | Why |
|---|---|---|---|
| 1 | Variable width (flex with labels always rendered). | **Fixed 64px collapsed / 200px expanded.** | MOAT §1 — restraint. Icons-only by default. |
| 2 | Labels always visible (`<span class="ni-label">`). | **Labels appear only on hover-expand or active.** | MOAT §3 — only the active state announces. |
| 3 | `.rail-brand` "Condura · v0.1" header. | **No header.** The wordmark lives in the titlebar. | APPFLOW §3.3 — titlebar owns brand identity. |
| 4 | `.rail-foot` with Pulse + "Local · Ollama". | **Kill-switch trigger only.** Status dot moves to titlebar DynamicIsland. | APPFLOW §3.6 — agent phase lives in DynamicIsland. |
| 5 | Per-item hover `transform: scale(1.04)`. | **No per-item scale.** A tint + label reveal. | MOAT §1 — `rotateX` and gratuitous scale shows off, not signal. |
| 6 | `.active::before` uses a CSS keyframe (`drawv`) that runs on each render — the segment re-animates every time the class toggles. | **Single persistent `<Thread>` FLIP-animated to the active item.** | Animation between items is the signature; per-item re-animating is noise. |
| 7 | `.active::after` shows a pollen dot at the right edge. | **Removed.** A single synapse thread is sufficient. One signature, one signal. | MOAT §3 — commit to one element. Don't dilute it. |
| 8 | 10 items include `replay`. | **10 items: `account` replaces `replay`.** `account` is the AccountMenu entry. | APPFLOW §14B — account is its own rail entry, not buried in Settings. |
| 9 | No badge dots. | **Badge dots for `channels` (connected), `sync` (paired peers count), `audit` (unread events).** | MOAT §2.4 — empty states teach; live state teaches too. |
| 10 | No tooltips. | **400ms-hover tooltip:** route name + keyboard chord. | MOAT §2.9 — no `title=`. The tooltip is a real primitive. |
| 11 | No keyboard chord handlers on the rail. | **`⌘1`–`⌘9` / `⌘0` jump to specific routes.** | MOAT §2.10 — the rail ships keyboard chords, not just Tab. |
| 12 | `:focus-visible` uses 4px pollen halo as a rectangular box-shadow. | **Same 4px halo for square-radius; drops inset for elements with `border-radius ≥ 8px`.** | MOAT §2.1 — focus rings track rounded shapes. |
| 13 | Items register no `prefers-reduced-motion` override. | **All animation suppressed; thread snaps instantly; tooltips appear on focus instead of hover.** | MOAT §2.3 — single global override, components never repeat the rule. |
| 14 | Per-item transition list re-declares 4 properties. | **Inherits the global `.tactile` class.** Components own meaning, not timing. | MOAT §2.7 — one tactile vocabulary. |
| 15 | `.nav-item:hover` lifts with a `var(--surface-card)` background tint. | **No background tint on hover.** Tint only on active (ink wash at `--surface-card`). | MOAT §1 — restraint. Hover reveals the label; active alone is filled. |
| 16 | `ondblclick`-able per item (no behavior), but `aria-current` is only set on active. | **`aria-current="page"` on active, `aria-label` on each item includes the chord (`"Chat, command 1"`), `aria-describedby` on focusable surfaces links to tooltip.** | A11y — screen reader must hear the chord. |

---

## 2. Layout

### 2.1 Geometry

The NavRail is the **leftmost vertical strip of the Shell**, fixed width with a
single dimension toggle (collapsed → expanded). It is positioned `position: fixed`
relative to the Shell viewport; the main surface begins at `left: 64px` (collapsed)
or `left: 200px` (expanded). The Shell's grid changes shape to accommodate.

```
Shell · collapsed (default)            Shell · expanded
┌─────┬──────────────────────┐         ┌────────┬──────────────────────┐
│  ⚏  │                      │         │  ⚏  Chat│                      │
│  ✦  │                      │         │  ✦  Hub│                      │
│  ◯  │   Main surface        │         │  ◯  Skl│                      │
│  ⌬  │   (current route)    │         │  ⌬  Syn│   Main surface        │
│  □  │                      │         │  □  Aud│   (current route)    │
│  ☼  │                      │         │  ☼  Chn│                      │
│  ⤳  │                      │         │  ⤳  Dlg│                      │
│  ⌥  │                      │         │  ⌥  Act│                      │
│  ◇  │                      │         │  ◇  Set│                      │
│  ⓘ  │                      │         │  ⓘ  Abt│                      │
│  ─  │                      │         │  ─────│                      │
│  ⛨  │                      │         │  ⛨ Halt│                      │
│     │                      │         │        │                      │
│ 64px│  flexible            │         │ 200px  │   flexible           │
└─────┴──────────────────────┘         └────────┴──────────────────────┘
```

### 2.2 Dimensions (collapsed default)

| Property | Value | Why |
|---|---|---|
| `width` | **64px** (fixed) | Houses a 32×32 hit target with 16px gutters each side. |
| `top` | `0` (full viewport height) | No vertical scroll. The Shell titlebar overlays above. |
| `padding-block` | `12px` top / `12px` bottom | First item aligns to the titlebar's content baseline. |
| `padding-inline` | `0` | The icons are visually centered via flexbox in the 64px-wide row. |
| `border-right` | `1px solid var(--hair)` | Right edge is a hairline. The thread lives *inside* this hairline. |
| `position` | `relative` (in Shell grid) | The Shell places it as grid-column 1. |
| `background` | `var(--paper)` (transparent over paper) | Inherits shell surface. No elevation. |
| `z-index` | `2` | Sits above the main surface, below modal overlays. |

### 2.3 Dimensions (expanded on hover or focus-within)

| Property | Value | Why |
|---|---|---|
| `width` | **200px** | Houses a 32px icon + 8px gutter + ~140px label + 16px right padding. |
| Transition | `width var(--dur) var(--ease)` (200ms) | Linear's reveal cadence; one width animate, no per-property stagger. |

The expanded width is the **outer** width of the rail. The main surface animates
its `margin-left` to match over the same 200ms (the rail and the main surface
move together — a coordinated FLIP-Group, not a CSS artifice).

### 2.4 Vertical rhythm — the 10 routes

The 10 route icons are arranged top-to-bottom in this order. The order is the
**single navigational contract** of the surface and is a locked decision:

```
┌───────────────────────────────────────────────────┐
│                                                   │
│  01  Chat         (⌘1)        · always available  │  ← route 1
│  02  Hub          (⌘2)        · always available  │
│  03  Skills       (⌘3)        · always available  │
│  04  Sync         (⌘4)        · status dot · peers│
│  05  Audit        (⌘5)        · status dot · pending│
│  06  Channels     (⌘6)        · status dot · connected│
│  07  Delegation   (⌘7)        · always available  │
│  08  Account      (⌘0)        · signed-in pill    │
│  09  Settings     (⌘9 / ⌘,)   · always available  │
│  10  About        (no chord)  · always available  │
│                                                   │
│  ─────────  (1px hairline divider)  ─────────     │  ← separators: none between routes
│                                                   │
│  ⛨  Halt         (no chord, danger on hover)      │  ← kill switch (separate group)
│                                                   │
└───────────────────────────────────────────────────┘
```

- **There are no visible separators between the routes** — the geometric rhythm
  and the typing-distance are the structure. (`MOAT.md §1` — restraint.)
- **One 1px hairline divider sits between routes and the kill switch** in
  `--hair`. The rail has exactly two visual regions: routes, halt.
- **The rail does not scroll.** All 10 routes fit in a 700px viewport
  (64px/wide × 10 rows × 44px/row + 12px+12px padding = 464px). At ≤700px
  viewport height the kill switch sits flush against the bottom of the rail
  with `margin-top: auto`. Above 700px the centered row group has equal
  top/bottom padding.

### 2.5 Route row geometry

Each route is a **44px tall, 100% width row** (44 × 64 collapsed / 44 × 200
expanded). Centered content; the icon is the focal element.

| Property | Value | Why |
|---|---|---|
| `height` | **44px** | At least 44×44 touch target (a11y). |
| `padding-inline` | `0` (collapsed) / `12px` (expanded) | Centers icon when 64px wide; left-aligns icon+label when 200px wide. |
| `border-radius` | `var(--r-sm)` (10px) on the active state | Only the active row gets a fill. |
| `gap` between icon and label | **8px** (expanded only) | Visual rhythm: icon label-icon label. |

### 2.6 The active row (filled)

When a row is the active route:
1. The icon turns `var(--content)` (full ink) — was `var(--content-mute)` on idle.
2. The icon's stroke color flips to `var(--synapse)` (green) — full presence.
3. The row background fills with `var(--surface-card)` (subtle ink wash).
4. A 2px-wide vertical **synapse thread** (`<Thread orientation="v" />`)
   sits flush against the row's left edge.
5. The thread color is `--synapse` (not `--content`); it is the only thing
   that announces the row's activeness from a distance.

The active state is **filled with one synapse thread**. Nothing else.

### 2.7 The synapse thread (the signature)

The thread is one persistent `<Thread orientation="v">` element, positioned
absolutely inside the rail at `left: 0; width: 2px`. Its `height` and `top`
position are derived via FLIP on every route change — see §5.1. The thread is
**always present**; opacity is `0` when no route is active (an unlikely state
but covered) and `1` otherwise.

| Property | Value |
|---|---|
| Element | `<Thread orientation="v" />` (existing component, vertical mode) |
| Width | `2px` |
| Color | `stroke: var(--synapse)` (`--synapse` from existing token) |
| Glow | `filter: blur(3px)` with `stroke: var(--synapse-glow); opacity: 0.4` |
| `pathLength` | `1` (existing Thread contract) |
| Transition | `transform var(--dur) var(--ease)` for FLIP; `opacity var(--dur) var(--ease)` for show/hide |
| Z-index | `3` (above row background tint, below row content) |

---

## 3. Content Slots

### 3.1 Per-route slot contract

Each route row has **four slots** in collapsed state, **five** in expanded.
All slots are computed properties; nothing is hard-coded per route.

```svelte
<button
  class="rail-row"
  class:active={route === item.id}
  aria-current={route === item.id ? 'page' : undefined}
  aria-label="{item.label}, command {item.chord}"
  aria-describedby="navrail-tip-{item.id}"
  data-route={item.id}
  tabindex="-1"
>
  <Glyph name={item.icon} size={20} stroke={1.5} />
  {#if item.badge}
    <span class="badge" data-tone={item.badge.tone} aria-hidden="true"></span>
  {/if}
  <span class="rail-row-label">{item.label}</span>
  <kbd class="rail-row-chord">{item.chord}</kbd>
  <Tooltip id="navrail-tip-{item.id}" delay={400}>
    <strong>{item.label}</strong>
    {#if item.chord}
      <kbd>{item.chord}</kbd>
    {/if}
    {#if item.tooltipHint}
      <span>{item.tooltipHint}</span>
    {/if}
  </Tooltip>
</button>
```

### 3.2 The 10 routes (locked content)

| # | Route ID | Label | Icon | Chord | Badge | Tooltip hint |
|---|---|---|---|---|---|---|
| 01 | `chat` | **Chat** | `chat` (speech-bubble) | `⌘1` | none | "Talk to Condura." |
| 02 | `hub` | **Hub** | `hub` (concentric arcs) | `⌘2` | none | "Browse the public Skills Hub." |
| 03 | `skills` | **Skills** | `skills` (sparkle) | `⌘3` | none | "Local installed procedures." |
| 04 | `sync` | **Sync** | `sync` (refresh arcs) | `⌘4` | dot · tone=`info` | "Pair a device. {N} paired." |
| 05 | `audit` | **Audit** | `audit` (doc with check) | `⌘5` | dot · tone=`warn` if any pending consent / `danger` if chain broken | "Every action, every model. {N} unread." |
| 06 | `channels` | **Channels** | `channels` (broadcast) | `⌘6` | dot · tone=`ok` if ≥1 connected | "Telegram, more soon. {N} connected." |
| 07 | `delegation` | **Delegation** | `delegation` (rays) | `⌘7` | dot · tone=`pollen` if any pending | "Sub-agents in flight. {N} pending." |
| 08 | `account` | **Account** | `account` (head+shoulders) | `⌘0` | dot · tone=`synapse` if signed in | "Sign in for Hub, donations, support." OR "Signed in as {email}." |
| 09 | `settings` | **Settings** | `settings` (gear) | `⌘9` (also `⌘,`) | none | "Power · autonomy · appearance · voice." |
| 10 | `about` | **About** | `about` (info-circle) | none | none | "Colophon · the 7 invariants." |

**Lock order.** The order above is fixed. The kill-switch is **not** one of the
ten; it is the eleventh element, separated by a hairline.

### 3.3 Status dots (badges)

A status dot is **always a 6px circle** anchored top-right of the row, with a
1.5px outer halo at the same color (suggests presence without screaming).
Each route has **at most one** dot.

| Tone | Color | Outer halo |
|---|---|---|
| `ok` | `var(--ok)` | `--ok` 25% via `box-shadow: 0 0 0 1.5px var(--ok)` |
| `info` | `var(--info)` | `--info` 25% |
| `warn` | `var(--warn)` | `--warn` 25% |
| `danger` | `var(--danger)` | `--danger` 25% |
| `pollen` | `var(--pollen)` | `--pollen` 25% |
| `synapse` | `var(--synapse)` | `--synapse` 25% (the "you're logged in" dot) |

The dot **does not animate** except:
- A 1.6s `breath` (subtle scale 1 → 1.15 → 1) on `tone="ok"` and
  `tone="synapse"` to indicate aliveness;
- A 1.4s `breathe` on `tone="warn"` and `tone="danger"` for urgency.

Both animations respect `prefers-reduced-motion`.

### 3.4 Tooltip primitive contract

A real component, not `title=`. Shared across the chrome.

```svelte
<Root
  target={buttonOrLink}    // The trigger
  label={string}           // Required (the visible line)
  chord?={string}          // Optional keyboard chord shown right-aligned
  hint?={string}           // Optional second line — sub-copied smaller
  delay?={number}          // Hover delay ms; default 400
  exit?={number}           // Close delay ms; default 75
  placement="right"        // Only "right" is supported in v0.1.0
/>
```

| Property | Value |
|---|---|
| Hover delay | **400ms** (matches MOAT §2.9 + appflow §3.4 hovered-palette cadence). |
| Exit delay | 75ms. |
| Animation | `opacity 0 → 1` + `translate(-4px, 0) → 0` over `160ms var(--ease)`. |
| Background | `var(--ink)` (ink surface, not paper) — contrasts to the paper rail. |
| Foreground | `var(--paper)` for text; `var(--pollen)` for chord keys. |
| Border-radius | `var(--r-sm)` (10px). |
| Box-shadow | `var(--shadow-float)` (the only allowed elevation token). |
| `aria-describedby` | Wired to the trigger's `id` (Rails trigger forwards to Tooltip). |
| Keyboard | `Esc` dismisses; `Tab` moves past without dismissing; persists while focused. |

The tooltip arrow is **omitted** — the placement is always 8px to the right
of the trigger, with the trigger's right edge visually anchored by a 2px
hairline. The hairline is the alignment cue. (MOAT §1 — restraint.)

---

## 4. State Matrix

Six states per row. The states compose; a row can be `hover` AND `focus`
simultaneously; both intent sources contribute.

### 4.1 States

| State | Trigger | Visual signature |
|---|---|---|
| **A · Idle** | default | Icon in `var(--content-mute)` at 0.65 opacity. No tint. No label. Row is `<button tabindex="-1">`. |
| **B · Hover-expand** | `mouseenter` on the row OR the row's child slot | Rail width animates 64 → 200 (200ms). Label fades in + slides 4px right (160ms `var(--ease)`, 80ms delay). Icon opacity → 1, icon color → `var(--content)`. Tooltip primes (400ms delay starts on hover). |
| **C · Active** (always coexists with A or B) | `route === item.id` | Row background fills `var(--surface-card)`; icon stroke `var(--synapse)`; the `<Thread>` vertical segment sits at the row's left edge. No label required in collapsed (the synapse thread is enough). |
| **D · Focus** | `:focus-visible` on the row OR Tab lands on it | 4px `var(--pollen-halo)` ring around the row at 4px inset (NOT a square box-shadow — see MOAT §2.1). `outline: none`. The halo is a `box-shadow: 0 0 0 2px var(--synapse), 0 0 0 5px var(--pollen-halo)`. The synapse ring + pollen halo track the rounded shape. |
| **E · Keyboard-focused** | row has focus AND no chord-mode | Tooltip **shows immediately on focus** (no 400ms delay). Tooltip persists while focused. The chord `kbd` in the tooltip lights up (`kbd { background: var(--pollen); color: var(--paper); }`). |
| **F · Press** | `:active` mousedown | `transform: scale(0.97)` + `filter: brightness(0.95) saturate(1.1)` + `translateY(0.5px)`. Released to idle (A/B) on `mouseup`. Follows `.tactile` class. |

### 4.2 Composability table

Each cell = resulting visual.

| | Idle | Hover | Focus | Press | Active (coexists) |
|---|---|---|---|---|---|
| Idle | A | A → B | A + D | A + F | A + C |
| Hover | — | B | B + D | B + F | B + C |
| Focus | — | — | D + E | D + E + F | (D or E) + C |
| Active | — | — | — | (C) + F | C (the persistent state) |

### 4.3 Tab-order inventory (the focused-ring order)

Tab traverses the rail in this exact order (the **third state, focus**, is
the visible surface for keyboard nav):

```
1.  Chat       (route · 1)
2.  Hub        (route · 2)
3.  Skills     (route · 3)
4.  Sync       (route · 4)
5.  Audit      (route · 5)
6.  Channels   (route · 6)
7.  Delegation (route · 7)
8.  Account    (route · 0)
9.  Settings   (route · 9)
10. About      (route · 10)
11. Halt       (kill switch — separated region)
```

`tabindex="-1"` is the default for all rows. The currently-active row gets
`tabindex="0"` (Roving tabindex pattern). The chord handlers (`⌘1`–`⌘9`/`⌘0`)
move focus + activate. Tab never lands on a non-active row until one of them
becomes active, which avoids scanning 10 rows with Tab alone.

### 4.4 The temporary "expand to chord" affordance

When the user presses `⌘K` to summon the palette, the **rail may auto-expand**
briefly (160ms `var(--ease)` on width) to telegraph its chord vocabulary. If
the user cancels the palette, the rail collapses back to 64px.

This is optional and behind a `prefers-reduced-motion` check. Implementation
note for Phase 4: keep the auto-expand gesture; do not show it as a marketing
animation. (MOAT §2.10 — the chord vocabulary is visible when asked for.)

---

## 5. Motion Choreography

### 5.1 The synapse-thread FLIP

The most important animation in this component: when the active route
changes, the persistent `<Thread>` element moves to the new row's left edge.
The move is **FLIP, not a layout animation**:

```
Phase 1 — Measure OLD:
  threadEl.getBoundingClientRect() → OLD.top, OLD.height

Phase 2 — Layout (synchronous, no transition):
  threadEl.style.top = NEW.top + 'px'
  threadEl.style.height = NEW.height + 'px'

Phase 3 — Measure NEW (also synchronous, frame 0):
  threadEl.getBoundingClientRect() → NEW.top, NEW.height

Phase 4 — Invert (apply OLD via transform, no transition):
  threadEl.style.transform = `translateY(${OLD.top - NEW.top}px)`
                                   + `scaleY(${OLD.height / NEW.height})`

Phase 5 — Play (next frame, transition enabled):
  threadEl.style.transition = 'transform 320ms var(--ease)'
  threadEl.style.transform = ''
  // → animates back to identity over 320ms
```

| Property | Value |
|---|---|
| Duration | **320ms** |
| Easing | `var(--ease)` (the global ease — `cubic-bezier(0.22, 1, 0.36, 1)`) |
| `transform-origin` | `top center` (so scaleY animates from the top edge, not the middle) |
| Path-draw | The thread's `pathLength` is `1`; `stroke-dashoffset` does **not** transition during FLIP (only on first mount). The thread is a continuous stroke that moves. |

**Why not a CSS keyframe re-run.** A `transition`-based re-run of
`scaleY(0→1)` (`drawv` in the current code) fires on every route change. It
implies "this thing just arrived" each time — which is true on the **first**
active state but **false** on subsequent moves. FLIP says "this thing is
*there*, it's just at a different address." Motion fidelity > visual flourish.

### 5.2 Hover-expand (width transition)

| Property | Value |
|---|---|
| Width | `64px` → `200px` |
| Duration | **200ms** |
| Easing | `var(--ease)` |
| `prefers-reduced-motion` | Width snaps; label fades-without-translate over 80ms. |

The rail's `width` and the Shell's `grid-template-columns: 64px 1fr` animate
in lockstep. Both listen to a shared CSS variable
(`--navrail-w: 64px`) so the layout reflows once.

### 5.3 Label appear (per slot)

| Property | Value |
|---|---|
| `opacity` | `0 → 1` |
| `transform` | `translateX(-4px) → translateX(0)` |
| Duration | **160ms** |
| Delay | **80ms** (allows the width to start opening first; the label arrives into the slot, not before) |
| Easing | `var(--ease)` |

The `<kbd>` chord follows the label by +40ms (3-line stagger).

### 5.4 Tooltip appear

| Property | Value |
|---|---|
| Trigger | `mouseenter` on row → after **400ms hover delay** |
| `opacity` | `0 → 1` |
| `transform` | `translateX(-4px) → translateX(0)` |
| Duration | **160ms** |
| Easing | `var(--ease)` |
| `prefers-reduced-motion` | Appears instantly on hover; no slide. |
| Keyboard-focus path | Appears on `focus` (no delay); persists while focused; dismisses on `blur`. |

### 5.5 Press (`.tactile`)

Inherits the global `.tactile` class. The class is the single source of
timing; this component does not redeclare transitions.

```css
.tactile {
  transition:
    transform var(--dur) var(--ease),
    background var(--dur) var(--ease),
    color var(--dur) var(--ease),
    box-shadow var(--dur) var(--ease),
    filter var(--dur-fast) var(--ease);
}
.tactile:active {
  transform: scale(0.97);
  filter: brightness(0.95) saturate(1.1);
  translate: 0 0.5px;  /* modern shorthand, single frame */
}
```

### 5.6 Focus halo

Inherits the global `.focus-ring` class (per MOAT §2.1). Tracks rounded shapes:

```css
.focus-ring:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--synapse),
    0 0 0 5px var(--pollen-halo);
}
```

### 5.7 Status dot breath

| Tone | Animation | Duration |
|---|---|---|
| `ok`, `synapse` | `breath`: `transform: scale(1) → scale(1.15) → scale(1)` | **1.6s** `ease-in-out` infinite |
| `warn`, `danger` | `pulse-warn`: same scale + a `--warn` halo `box-shadow` cycle | **1.4s** `ease-in-out` infinite |
| `info`, `pollen` | none — static presence | — |

All suppressed by `prefers-reduced-motion: reduce` (one global override in
`condura.css`, components do not redeclare).

### 5.8 `prefers-reduced-motion` summary

| Feature | Default | Reduced |
|---|---|---|
| Thread FLIP | animates 320ms | snaps |
| Width hover-expand | animates 200ms | snaps |
| Label appear | animates 160ms / 80ms delay | fades-only, 80ms |
| Tooltip appear | animates 160ms / 400ms delay | appears instantly on hover OR focus |
| Press | tactile scale + filter | scale only (no filter, no translate) |
| Status dots | breathe / pulse-warn | static |

All five rows above are owned by a single `@media (prefers-reduced-motion: reduce)`
block in `condura.css`. NavRail contributes **zero** prefers-reduced-motion
rules of its own (MOAT §2.3 — components do not redeclare the media query).

---

## 6. Keyboard

### 6.1 Chords (the locked chord map)

The chord map is a contract; it does not change per user.

| Chord | Action | Notes |
|---|---|---|
| `⌘1` | Open `chat` | The default route. Chord is active even if user is already here. |
| `⌘2` | Open `hub` | |
| `⌘3` | Open `skills` | |
| `⌘4` | Open `sync` | |
| `⌘5` | Open `audit` | |
| `⌘6` | Open `channels` | |
| `⌘7` | Open `delegation` | |
| `⌘8` | (reserved — currently inactive; Halt is not a route) | Future: maybe a quick switcher. |
| `⌘9` | Open `settings` | Mirrors `⌘,` (a separate binding on `Shell.svelte`). |
| `⌘0` | Open `account` | |
| no chord | `about` | Intentionally chord-less — the slowest, least-used route. `⌘?` may be added in v0.2.0. |

Chords work **regardless of focus** while the Webview is foreground (the
Shell owns the keydown listener — `window.addEventListener('keydown', …)`
in `Shell.svelte`, filtered by `e.metaKey || e.ctrlKey` to match the existing
`⌘K`/`Shift+P` pattern). The NavRail does not own the listener.

### 6.2 Intra-rail arrow nav

When focus is **on the rail** (any row has `:focus-visible`):
- `ArrowDown` / `↓` → focus the next route row (wraps from `about` →
  `chat`).
- `ArrowUp` / `↑` → focus the previous route row (wraps from `chat` →
  `about`).
- `Home` → focus the first route row (`chat`).
- `End` → focus the last row before the kill switch (`about`).
- `Enter` / `Space` → activate the focused row (navigate to the route).
- `Escape` → collapse the rail (if expanded) AND blur focus.
- `Tab` (forward) → leave the rail to the next focusable element
  (titlebar's ⌘K chip).
- `Shift+Tab` (backward) → leave the rail to the previous focusable
  element.

When focus is on the **kill-switch row**:
- `ArrowDown` → wraps to `chat`.
- `ArrowUp` → focuses `about`.
- `Enter` / `Space` → activates kill switch (calls
  `halt.halt(ctx, 'rail_button')`).

### 6.3 Activation behavior

When a chord or arrow-nav-and-Enter activates a route:
1. The route's hash is set: `window.location.hash = ROUTE_HASH[id]`
   (matches the existing handler).
2. The route's row receives `class="active"` synchronously (no transition).
3. The synapse thread FLIPs to the new row (320ms).
4. Focus moves **to** the new route's main surface (the first focusable element
   inside the route's component). This is the "you are here" handoff — focus
   follows navigation.
5. The rail does **not** retain focus (the handoff is to the new route).

The only exception is `account`, which opens an **in-place popover**
(`AccountMenu`), not a route change. Focus stays on the rail row; the menu
opens to the right with focus-trap. The synapse thread stays put
(`account` is the active row).

### 6.4 Tab order (final)

```
1. Titlebar (44px)     :  wordmark · theme toggle · ⌘K chip
2. NavRail — Chat      :  (active row has tabindex=0; the others have tabindex=-1)
3. NavRail — Hub
4. NavRail — Skills
5. NavRail — Sync
6. NavRail — Audit
7. NavRail — Channels
8. NavRail — Delegation
9. NavRail — Account
10. NavRail — Settings
11. NavRail — About
12. NavRail — Halt     :  kill switch
13. Main surface       :  the route's first focusable element
```

The kill switch is the **last** rail stop because halting is the least
common action. (Per MOAT §2.10: "The Esc dismisses the topmost overlay" and
the rail is the last interactive region before the route content.)

---

## 7. Components Used

### 7.1 Existing primitives (consumed)

| Component | Source | Role in NavRail |
|---|---|---|
| `Glyph` | `./Glyph.svelte` | The icon per row. Always `size=20 stroke=1.5`. |
| `Thread` | `./Thread.svelte` | The persistent vertical active-state segment (`orientation="v"`). |
| `Pulse` (optional) | `./Pulse.svelte` | The status dot lives `Pulse`-adjacent — a `<Pulse size=6 phase={tone}/>` for live tones (`ok`, `warn`, `danger`). This avoids hand-rolling breath keyframes. |
| `Tooltip` | `./Tooltip.svelte` (new in Phase 4) | The hover/focus tooltip. New primitive; this spec requires its creation if not present. |

### 7.2 Internal sub-components (defined in `NavRail.svelte`)

- `RailRow` — a single route row; takes `{id, label, icon, chord, badge, hint, isActive, onActivate}`. Owns the tooltip + label slot + icon slot + badge slot.
- `RailHalt` — the kill-switch row; takes `{onHalt}`. Has different hover colors (danger).

### 7.3 Consumed events / actions

| Source | What | When |
|---|---|---|
| `Shell.svelte` `route` prop | the active `RouteId` | drives `<Thread>` FLIP + `aria-current` |
| `Shell.svelte` `onnavigate` callback | `(id) => window.location.hash = ROUTE_HASH[id]` | wired into the `RouteId → hash` mapping |
| `halt` store | a `halt` method that the rail calls on kill-switch activate | routes to the existing `halt.halt(ctx, reason)` RPC |
| Channels / Sync / Audit / Account stores (svelte) | status + count | drives `badge.tone` and tooltip hint text |

### 7.4 Kill-switch overlay trigger

The rail's kill-switch row **only triggers the halt**. It does not render the
halted overlay — that is `KillSwitchOverlay.svelte` in the Shell. The chain:

```
row click / row Enter / row Space
  → halt.halt(ctx, 'rail_button')
  → daemon: emit halt event
  → halt store: state.halted = true, state.reason = 'rail_button'
  → Shell renders <KillSwitchOverlay> (already in place via {#if halt.state.halted})
```

The rail's kill-switch row is the **input**, not the display. (Per
APPFLOW.md §6.4, the overlay is always rendered by the Shell.)

---

## 8. Data Fetched

### 8.1 Reactivity contract

The rail does **not** fetch; it subscribes. The Shell is the data fetcher; the
rail renders from props and stores.

| Data | Source | When read | When written |
|---|---|---|---|
| Current active `RouteId` | `Shell.svelte` (`route` prop) | every render | on hash change |
| Channels status | `channels` svelte store | on mount + 10s poll | updates `channels[*].state` |
| Sync peers | `sync` svelte store | on mount + 5s poll | updates `sync.pairs.length` |
| Audit pending | `audit` svelte store | on mount + after each `audit.refresh()` | updates `audit.events.filter(needsAttention).length` |
| Account sign-in | `account` svelte store | on mount | on `account.signInWithEmail` success |
| Delegation pending | `pendingActions` svelte store | polled at 5s (per APPFLOW §4.8) | updates `pendingActions.length` |
| Replay chain integrity | `replay` svelte store | on mount + every 60s | updates `replay.integrity.valid` |

Each subscription is **at most one per route row**. No double-fetching,
no shared cache (each store owns its own fetch).

### 8.2 The badge pipeline

For each route, the badge is computed from the relevant store(s):

| Route | Tone | Source |
|---|---|---|
| `chat` | none | — |
| `hub` | none | — |
| `skills` | none | — |
| `sync` | `info` if `sync.pairs.length > 0`, else none | `sync.pairs.length` |
| `audit` | `warn` if `audit.events.filter(e => e.requires_attention).length > 0`; `danger` if `replay.integrity.valid === false`; else none | combined from `audit` and `replay` stores |
| `channels` | `ok` if any `channels[i].state === 'connected'`; `warn` if any `degraded`; else none | `channels.state` |
| `delegation` | `pollen` if `pendingActions.length > 0`; else none | `pendingActions.length` |
| `account` | `synapse` if `account.signedIn === true`; else none | `account.signedIn` |
| `settings` | none | — |
| `about` | none | — |

Each badge is computed via a `$derived` rune per row in the component. The
tooltip's `hint` prop also pulls from the same source — the dot and the
tooltip agree.

### 8.3 Error / loading posture

The rail does **not** render skeletons. If a store has not resolved (first
mount), the row renders with **no badge** — silent absence. The tooltip
hint falls back to the route's static hint. This is the MOAT §1 restraint:
the rail's job is to be quiet until something needs to be said.

If a store errors (e.g., daemon unreachable), the row still renders. The
error is owned by the **main surface**, not the rail. The rail never
cascades an error.

---

## 9. Design Decisions (MOAT Compliance)

Which mature rules from `MOAT.md` this NavRail passes, and how.

### 9.1 §1 The Restraint Test

| Rule | What we did | Pass? |
|---|---|---|
| §1.1 The Ritual's 6 keyframes | N/A (Ritual, not NavRail) | n/a |
| §1.2 The err-state copy-paste | N/A | n/a |
| §1.3 The hero/eyebrow headlined block | NavRail has no hero. The active state is one thread + one icon. | **Pass.** |
| §1.4 The `<Cursor />` unconditionally | NavRail adds `data-hover="1"` on rows on `pointerenter` (a `hover-region` directive). The cursor ring (§5.1) is enabled for the rail. | **Pass — enabled.** |
| §1.5 The Skills card hover `rotateX(2deg)` | NavRail rows do not transform on hover. Width changes are width-only (a horizontal reveal, not a 3D tilt). | **Pass.** |
| §1.6 The Ritual constellation SVG | N/A | n/a |
| §1.7 The `.alive` italic-green span | NavRail uses it once: the **active row's icon stroke color**, in `--synapse`. Not italic, not green text — a single-stroke icon recolor. The single load-bearing place. | **Pass.** |

### 9.2 §2 The Detail Test

| Rule | Pass? |
|---|---|
| §2.1 Focus rings track rounded shapes | Halo via `box-shadow: 0 0 0 2px var(--synapse), 0 0 0 5px var(--pollen-halo)` (the synapse ring + pollen halo combo). The 10px row border-radius gets a rounded halo. **Pass.** |
| §2.2 Press states have weight | The `.tactile` class adds `filter: brightness(0.95) saturate(1.1)` + 0.5px translateY. **Pass.** |
| §2.3 `prefers-reduced-motion` consistent | One global override in `condura.css` (per the spec). NavRail contributes zero local rules. **Pass.** |
| §2.4 Empty states must teach | The rail *itself* never empties; the **tooltips teach the chord vocabulary** (`"Chat, command 1"`). When a route's main surface is empty, its tooltip hint says so indirectly ("Browse the public Skills Hub"). **Pass.** |
| §2.5 Loading states must feel alive | The rail does not load (it never shows a spinner — see §4.3 of the rule). The active state has the thread; idle has nothing. **Pass.** |
| §2.6 Error states must guide | Errors belong to the main surface, not the rail. See §8.3 above. **Pass.** |
| §2.7 Tactile vocabulary one thing | The `.tactile` class is global; NavRail inherits it. **Pass.** |
| §2.8 Three overlays — modal/sheet/popover | `account` opens a popover (`AccountMenu`, treated as `.c-popover`). `<Tooltip>` is *not* an overlay (it's anchored, dismiss-on-leave-only); it lives separately. **Pass.** |
| §2.9 Tooltip vs popover vs sheet | A real `Tooltip` primitive is required (see §3.4). No `title=` attributes anywhere. **Pass.** |
| §2.10 Keyboard story | §6 of this spec: chords, arrow-nav, focus persistence, `?` for Shortcuts (handled by Shell), `Esc` to dismiss overlays. **Pass.** |

### 9.3 §3 The Signature (the Thread)

The Thread appears in NavRail:
1. As the **active-row segment** — vertical, 2px wide, synapse-green. The
   *anchor*. The one thing the user sees at a glance.
2. As the kill-switch row's left-edge **divider hairline** (same Thread
   component, but `glow={false}`, `orientation="h"`, sitting horizontally
   between routes and the kill switch). The user sees the Thread *as the
   structure*, not just the active state.

Two manifestations, one component, one metaphor. The brand.

### 9.4 §4 What We Will Not Do

| Anti-pattern | NavRail posture |
|---|---|
| Gradient text | None. Icons are single-stroke. |
| Emoji as UI icons | None. All icons come from `icons.ts`. |
| Glassmorphism | No `backdrop-filter`. The rail is paper-on-paper. |
| Rainbow accents | Status dots are token-driven (`ok/warn/danger/info/pollen/synapse`). No purple, cyan, teal. |
| "Welcome to the future" copy | No marketing copy on the rail. Tooltips are descriptive nouns. |
| Fake enthusiasm | None. |
| Spinner loaders | The rail never loads. |
| Rectangular focus outlines | The rounded halo (§9.2 §2.1). |
| Double shadows | The Tooltip is the only elevated surface (`--shadow-float`). The rail is paper-on-paper; rows are flat. |
| Animation that doesn't carry meaning | Every animation has a verb: FLIP (route changed), width-expand (hovering), label-slide (label arrived), tooltip-appear (400ms passed), press (down event), breath (alive). |

### 9.5 §5 The $50M Feel — what the rail contributes

- **The thread is moving with intent.** Other sidebars change what is
  highlighted with a class toggle. This rail's thread physically *moves*
  to the new row over 320ms. Two pixels of `synapse-green` and the user
  feels the connection being remade.
- **The hover-expand reveals the vocabulary.** The rail is quiet by
  default — but on hover, the user sees all the labels and chords. The
  rail teaches itself.
- **The kill switch is rare and recentered.** Most sidebars bury a
  destructive action. Putting it on a hairline below the routes, in its
  own region, with hover-to-danger — this says "this exists, but you
  shouldn't reach for it."

---

## 10. Accessibility Contract

### 10.1 Roles, names, and relationships

- The rail container is `<nav aria-label="Primary">`.
- Each row is `<button type="button" role="link" aria-label="{label}, command {chord}">`.
- The active row is `aria-current="page"`.
- The kill-switch row is `<button type="button" aria-label="Halt the agent">`.
- The Tooltip is referenced via `aria-describedby` on the trigger.

### 10.2 Focus

- The active row has `tabindex="0"`.
- All others have `tabindex="-1"` (roving tabindex).
- `:focus-visible` is the only visible ring trigger — never `:focus`.
- The 4px pollen halo + 2px synapse ring combination is the focus state
  (per MOAT §2.1).

### 10.3 Screen reader

- The icon is `aria-hidden="true"` (decorative).
- The chord `<kbd>` is read as "command one" via the chord's
  `aria-label` (the chord field carries both visual + screen-reader text).
- The Tooltip uses `role="tooltip"` + `aria-live="polite"` so its content
  is announced when first appearing, but does not interrupt.
- Status dots are `aria-hidden="true"` (the tooltip communicates the
  state with numbers — "3 paired", "5 unread").

### 10.4 Reduced motion

- `prefers-reduced-motion: reduce` users: thread snaps, tooltips appear
  instantly on focus, no width animation (snap to 200px), no label
  slide, no press filter.
- Focus behavior is identical — the rail is a keyboard surface
  independent of motion preference.

### 10.5 High contrast / forced colors

- In `forced-colors: active`, the synapse thread becomes
  `ButtonText` (system); the icon stroke becomes `ButtonText`. The
  rail reads in OS-driven contrast.
- Tooltip background uses `Canvas` (system); text uses `CanvasText`.

---

## 11. Implementation Notes for Phase 4

### 11.1 Files to create or replace

| File | Action |
|---|---|
| `/condura/NavRail.svelte` | **Replace** with the redesigned component. |
| `/condura/Tooltip.svelte` | **Create** (new primitive; consumed in §3.4). |
| `/condura/specs/SCREEN_NAVRAIL.md` | This file. |
| `/condura/specs/SCREEN_NAVRAIL.test.ts` | Test file (see §12). |

### 11.2 Files to read before editing

- `/condura/Shell.svelte` — to confirm the `route` prop contract and
  the row → hash mapping export point.
- `/condura/icons.ts` — to confirm `kill-switch` glyph exists (yes; line 104).
- `/condura/condura.css` — to confirm the global tokens (`--ease`,
  `--dur`, `--dur-slow`, `--pollen-halo`, `--synapse`, `--shadow-float`).
- `/condura/Thread.svelte` — to confirm the `orientation="v"` API (yes; lines 6–15).
- `/condura/Glyph.svelte` — to confirm `size` and `stroke` props (yes; line 10).

### 11.3 micro-decisions deferred to the next session

1. **Hover-expand persistence.** Does the rail stay expanded after the
   user clicks away from it? Current direction: **yes, for 1.4s**, then
   collapses. Implementation: a `lastInteraction` timer in NavRail's
   state. (Cosmetic; Phase 4 may decide differently.)
2. **Sticky `:focus-visible` after click.** When the user clicks a row,
   should the row retain focus briefly so the tooltip shows? Current
   direction: **no** — focus transfers to the main surface on activation.
3. **The Halt chord.** Currently the rail-halt row has no chord. Is there
   a chord that makes sense (e.g., `⌘.` to halt)? Current direction:
   **leave it chord-less** in v0.1.0. The button is discoverable.
4. **Account popover placement.** When user activates `account` (⌘0 or
   click), does the AccountMenu open to the right of the rail or below?
   Current direction: **to the right** (consistent with Tooltip placement).
   Implementation note: the popover is `.c-popover` per MOAT §2.8.
5. **The `⌘8` reserved slot.** Currently unused. Should the rail render
   `⌘8` as a future placeholder (greyed)? Current direction: **no** —
   silence is better than a ghost slot.
6. **`prefers-reduced-motion` focus halo.** When motion is reduced, does
   the focus halo still pulse? Current direction: **static** (the global
   media query kills all keyframes; halo is a static `box-shadow`).
7. **The `data-hover` directive.** MOAT §5.1 requires a `use:hoverRegion`
   action. If not yet implemented in `condura/`, create it.

### 11.4 Implementation order (recommended)

1. Add `Tooltip.svelte` (new primitive) — needed for the rail but also
   valuable for the titlebar's theme toggle (MOAT §2.9).
2. Add a `hover-region` action (`/condura/hoverRegion.ts`) — single line,
   but used by every interactive surface.
3. Add the FLIP helper utility (`/condura/flip.ts`) — used by the rail's
   thread animation; reusable for the Settings sub-nav (MOAT §5.4).
4. Replace `NavRail.svelte` with the new component, keeping the exported
   `RouteId`, `ROUTE_HASH`, `hashToRoute` symbols.
5. Add tests (see §12).
6. Smoke-test by hand: load the dev webview, hit each chord, watch the
   thread move. Verify the kill-switch triggers the overlay.

### 11.5 Composer-driven hints (do not write code from this)

The `RouteRow` slot contract in §3.1 is a Svelte 5 runes-era snippet (a
`<button>` with `<Glyph>` and `<Tooltip>` children). The exact import is
`import { Tooltip } from './Tooltip.svelte'` (the new file). The
`navrail-tip-{id}` ID pattern is shared across all 11 rows (10 routes + 1
halt); make sure the IDs are unique per row.

### 11.6 What I deliberately did NOT decide here

- **The exact polish color values in the dark mode.** Inherits the global
  tokens. The `prefers-color-scheme: dark` flip happens at `:root` and
  re-derives all tokens.
- **The exact Tooltip's tail/arrow.** Per §3.4, no arrow. Hairline alignment.
- **The `AccountMenu` popover content.** That component exists independently
  (Phase 14B, per APPFLOW §6.9). The rail only opens it.

---

## 12. Test Plan

Vitest + @testing-library/svelte 5 + jsdom (per `CLAUDE.md` §8). All tests
live in `/condura/specs/SCREEN_NAVRAIL.test.ts`.

### 12.1 Required test cases

**Render & semantics**
1. `renders 10 nav rows + 1 halt row when mounted.`
2. `sets aria-current="page" on the active row only.`
3. `aria-label includes the chord for rows that have one.`
4. `aria-label omits the chord for the about row.`

**Active state**
5. `only one row has class="active" at any time.`
6. `applying class="active" FLIPs the Thread to the row's left edge.`
7. `the Thread is positioned 2px wide at left:0 of the rail.`
8. `the Thread uses stroke var(--synapse).`

**Hover-expand**
9. `hovering a row animates the rail width from 64 to 200 over 200ms.`
10. `mouse leave collapses the rail back to 64.`
11. `mouse leave within 75ms does not flicker the tooltip.`

**Tooltip**
12. `tooltip appears after 400ms hover delay.`
13. `tooltip content includes the route's label and chord.`
14. `tooltip has aria-describedby wired to the trigger.`
15. `focus on a row shows tooltip immediately (no 400ms wait).`
16. `blur removes the tooltip after 75ms.`
17. `keyboard user sees chord kbd lit when tooltip is open.`

**Keyboard**
18. `ArrowDown from active row focuses the next row.`
19. `ArrowUp from the first row wraps to the last.`
20. `Home focuses the first row.`
21. `End focuses the last row.`
22. `Enter on a focused row calls onnavigate with that route id.`
23. `Space does the same.`
24. `Esc collapses the rail and blurs.`
25. `Tab leaves the rail to the main surface focusable.`

**Kill switch**
26. `clicking the halt row calls halt.halt(ctx, 'rail_button') with the right args.`
27. `the halt row has aria-label="Halt the agent".`
28. `the halt row tooltip says "Stop every stream." (no chord).`
29. `the halt row hover turns the icon color var(--danger).`

**Status dots**
30. `channels row shows an ok-tone dot when any channel is connected.`
31. `audit row shows a warn-tone dot when pending count > 0.`
32. `audit row shows danger-tone when replay.integrity.valid === false.`
33. `dots inherit the global prefers-reduced-motion suppress rule.`

**Reduced motion**
34. `with prefers-reduced-motion: reduce, FLIP is instant.`
35. `with prefers-reduced-motion: reduce, tooltip appears on hover with no slide.`
36. `with prefers-reduced-motion: reduce, width snaps.`

**Stores / data fetching (mocked)**
37. `subscribes to channels, sync, audit, account, replay, pending stores on mount.`
38. `unsubscribes from stores on destroy.`
39. `does not call ipc.X directly — stores are the contract.`

**Edge cases**
40. `renders with no route active (empty hash) — no row has class="active".`
41. `renders with a malformed hash — fallback to chat row.`
42. `nav to a non-existent route — captured silently (no console error).`

### 12.2 Visual regression (Loki or pixelmatch — pick one)

A single snapshot per state × per route combination, run via Playwright:

1. `idle · all 10 routes` (collapsed screenshot)
2. `hover · chat row` (expanded)
3. `active · hub row` (collapsed, thread at row 2)
4. `focus · skills row` (visible halo)
5. `keyboard · tab in to row 1` (focus halo)
6. `reduced-motion · same as #3 but instant thread`

The 6 frames are reference visuals; Phase 4 owns whether to wire Loki.

### 12.3 A11y assertions

For every test file: `expect(axe(container)).toHaveNoViolations()`. (`axe-core`
is already in the dev-dependencies per the existing harness setup.)

---

**End of spec. The next agent implements Phase 4 from this document.**
