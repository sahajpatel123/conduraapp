# Channels Screen Spec — `#/channels` · `Channels.svelte`

> **Voice:** `DIRECTION.md` (paper notebook, never louder than the room).
> **Bar:** `MOAT.md` (restraint, detail, signature, anti-patterns).
> **Flow:** `APPFLOW.md` §4.7 · **Tokens:** `condura.css`.

---

## 1. LAYOUT & CONTENT

Single column, max-width 880 px, padding-top `--space-7`. Three regions: header → grid → footer note.

### 1.1 Page chrome

| Region | Type | Copy |
|---|---|---|
| Eyebrow | mono 11px, +0.22em, uppercase, `--content-faint` | `— Reach · on your terms` |
| h1 | Instrument Serif, `clamp(28px, 3vw, 40px)`, −0.03em, `--content` | `Threads outward.` |
| Sub | Inter 16/1.55, `--content-soft`, max 56ch | "Condura can reach you on Telegram today. WhatsApp, Slack, Discord, and iMessage arrive in v0.2.0 — we don't fake them. Each connection is a thread you tie, and you can revoke it any time." |

### 1.2 Row anatomy (3-col grid `1fr auto auto`, gap `--space-4`)

| Col | Slot | Type |
|---|---|---|
| **Cell** | Name + hint | Instrument Serif 18px (`--content`) + JetBrains Mono 10px, +0.1em, uppercase (`--content-faint`) |
| **Signal** | 5 cellular-bar dots, stepped heights | 4 px wide, heights **8 / 12 / 16 / 20 / 24 px**, gap 3 px, baseline-aligned, container 28 px |
| **Action** | Connect / pill | `<Button variant="primary" size="sm" magnetic>` or `.pill-soon` |

Row: 1 px `--hair` border, `--r-md`, bg `--surface-card`. Hover: `translateY(-1px)` + border `--hair-strong` + bg `--paper-2` over `--dur --ease`.

### 1.3 The five default rows

| id | name | default state | hint (mono micro) |
|---|---|---|---|
| `telegram` | Telegram | off | `Connect a BotFather token` |
| `whatsapp` | WhatsApp | soon | `coming in v0.2.0` |
| `slack` | Slack | soon | `coming in v0.2.0` |
| `discord` | Discord | soon | `coming in v0.2.0` |
| `imessage` | iMessage | soon | `coming in v0.2.0` |

### 1.4 Footer note (always present)

Italic Instrument Serif 14 px `--content-mute`. `Audit chain` is a `.threadlink` → `#/audit` via `ROUTE_HASH.audit`; underline draws in `scaleX(0→1)` on hover/focus. `<Pulse phase="idle" size=6 />` precedes. A `<Thread orientation="h" draw={noteIn} class="note-thread" />` finishes the block (signature Thread, 520 ms).

> Outbound messages always pass the consent gate. Inbound traffic is logged on the **Audit chain**.

### 1.5 Loading chip (conditional, when `hydrating`)

`<Pulse phase="thinking" size=8 />` + mono-uppercase label.

> PROBING REACH…

### 1.6 Error block (conditional, when `hydrateError`)

`<Pulse phase="error" size=8 />` + italic Instrument Serif 22px headline + italic 15px sub + `err-hair` (1 px gradient `scaleX(0→1)` over 600 ms `--ease`, 120 ms delay).

> Headline: *We couldn't reach the daemon.*
> Sub: `{error} The defaults below are honest — Telegram is connectable today, the rest are v0.2.0.`

The grid still renders (defaults — Telegram `off`, four `soon`). Per I4: every state is reachable; defaults are truthful degradation, not a dead wall.

---

## 2. STATE MATRIX

### 2.1 Row states

| State | Lit dots | Color | Animation | Action | a11y label |
|---|---|---|---|---|---|
| **connected** | 5 of 5 | `--synapse` | breathe cascade, 1.6 s `--ease` infinite, 120 ms per-dot stagger | `Manage` (Disconnect deferred) | `{name}, connected` |
| **degraded** | 3 of 5 (leftmost) | `--warn` | none — hold steady | `Continue →` | `{name}, degraded` |
| **off** | 0 of 5 | unlit (`--hair-strong`) | none | `Connect →` | `{name}, not connected` |
| **soon** | hidden (`display: none`) | — | — | `.pill-soon` `v0.2.0` | `{name}, unbuilt — coming in v0.2.0` |

### 2.2 Page states

| State | Trigger | Visible |
|---|---|---|
| **hydrating** | mount, before `channelsList` resolves | grid + loading chip (1.5) |
| **error** | `channelsList` rejected | grid + error block (1.6) |
| **idle** | hydrate resolved | grid + footer note |

### 2.3 Dot counts

```
connected → 5    degraded → 3    off, soon → 0 (soon hidden)
```

### 2.4 Exact copy

| Where | Exact text |
|---|---|
| Eyebrow | `— Reach · on your terms` |
| Title | `Threads outward.` |
| Connect (idle) | `Connect →` |
| Connect (busy) | `opening…` |
| Pill | `v0.2.0` |
| Hint, telegram off | `Connect a BotFather token` |
| Hint, telegram degraded | `token entry → open Channels` |
| Hint, soon | `coming in v0.2.0` |
| Loading | `PROBING REACH…` |
| Error headline | `We couldn't reach the daemon.` |
| Footer | `Outbound messages always pass the consent gate. Inbound traffic is logged on the Audit chain.` |

---

## 3. MOTION CHOREOGRAPHY

| Element | Property | Duration | Easing | Trigger |
|---|---|---|---|---|
| Row hover | `transform: translateY(-1px)` + `border: --hair-strong` + `bg: --paper-2` | `--dur` | `--ease` | mouseenter |
| Row focus-visible | `border: --synapse` + `0 0 0 4px var(--pollen-halo)` | `--dur-fast` | `--ease` | focus-visible |
| **Cellular-bar cascade (signature)** | dot 0..4 lit, `breathe` keyframe | 1.6 s infinite, 120 ms per-dot stagger | `--ease` | row mounts `connected` |
| **Cellular-bar degraded** | 3 leftmost hold `--warn`, no animation | — | — | `state === 'degraded'` |
| Cellular-bar off / soon | unlit or hidden | — | — | `state ∈ {off, soon}` |
| Footer Thread | `stroke-dashoffset 1 → 0` | `--dur-slow` | `--ease` | mount, next rAF |
| Threadlink underline | `transform: scaleX(0→1)`, `transform-origin: left` | `--dur` | `--ease` | hover / focus-visible |
| Soon row hover | `transform: none`, bg unchanged | — | — | mouseenter |
| `err-hair` | `transform: scaleX(0→1)` | 600 ms | `--ease`, 120 ms delay | mount of error block |
| Pulse (loading + footer idle + error) | `scale(1 → 1.18)` + `opacity 0.85 → 1` | 4 s loop | `--ease` | always-on |

### 3.1 Connect click → cascade kickoff

`busy = r.id` (button → `opening…`, disabled) → `openBotFather()` opens `https://t.me/BotFather` (Wails `runtime.BrowserOpenURL` → `window.open` fallback) → `ipc.call('channels.telegram.start', {})` → local state flips to `degraded` (3 warn dots hold steady). The full cascade fires only when `channels.status` later reports the token healthy.

### 3.2 Reduced-motion contract

| Override | What it does |
|---|---|
| `:global(:root[data-energy='low']) .row.connected .dot.on { animation: none }` | Cascade stops when battery is low (DIRECTION.md §5 / CLAUDE.md §6.4). Dots hold steady. |
| `@media (prefers-reduced-motion: reduce)` (component-local at `Channels.svelte:444–451`) | Cascade off; `err-hair` renders at `scaleX(1)`; threadlink `::after` `transition: none`. |

The component-local reduced-motion block is a known MOAT §2.3 duplication; the global block in `condura.css` is the canonical owner. Documented as drift in §8.

---

## 4. KEYBOARD

| Key | Action |
|---|---|
| **Tab** / **Shift+Tab** | Focus next / previous row, then `Audit chain` threadlink |
| **Enter** on connectable row | `connect(id)` — opens BotFather, flips local state to `degraded` |
| **Enter** on `soon` row | noop (cursor `not-allowed`) |
| **Enter** on `Audit chain` threadlink | navigate `#/audit` |
| **⌘T** | jump focus to Telegram row's `<Button>` (route-local chord, scoped to `#/channels`, only fires when Telegram is `off` or `degraded`) |
| **Esc** | blur focused element (no modal exists on this route) |

### 4.1 Focus halo (per DIRECTION.md §6 Rule 6 / MOAT.md §2.1)

| Element | Halo |
|---|---|
| Row (rounded `--r-md`) | `box-shadow: 0 0 0 4px var(--pollen-halo)` (no inset) |
| `<Button>` (pill `--r-pill`) | `0 0 0 2px var(--synapse), 0 0 0 5px var(--pollen-halo)` |
| Threadlink | `0 0 0 4px var(--pollen-halo)` |

No `outline: 1px solid var(--content)` anywhere on this surface.

---

## 5. COMPONENTS USED

| Component | Where | Purpose |
|---|---|---|
| `Pulse.svelte` | loading (`phase="thinking"`, size 8) + footer (`phase="idle"`, size 6) + error (`phase="error"`, size 8) | heartbeat / phase indicator — no spinners (MOAT §4.7) |
| `Thread.svelte` | footer note | signature Thread drawing in on mount |
| `Button.svelte` | row action slot | variant `primary`, size `sm`, `magnetic={true}`, disabled while `busy === r.id` |
| `.threadlink` (inline) | footer note `Audit chain` | synapse link with hover-draw underline + focus halo |
| `Glyph.svelte` | (reserved v0.2.0) per-channel icon next to name | single-stroke 1.5 px currentColor (DIRECTION.md §6 Rule 2) |
| `Tooltip.svelte` | (reserved v0.2.0) hover over `.pill-soon` | hover-delay 400 ms / exit 75 ms (MOAT §2.9) |
| `ROUTE_HASH.audit` | from `NavRail.svelte` | threadlink target |

**Not used here (and why):** `Spinner` (MOAT §4.7 — doesn't exist); `<ErrorState>` (MOAT §1.2 — extraction pending; Channels has its own inline err block); `<RouteHero>` (MOAT §1.3 — extraction pending; Channels hand-rolls eyebrow / h1 / sub); detail sheet / modal / popover (APPFLOW §4.7 — flat list, no right rail).

---

## 6. DATA FETCHED

| RPC | When | Purpose |
|---|---|---|
| `ipc.channelsList()` | onMount, best-effort | hydrate rows; defaults stand if reject |
| `ipc.call('channels.telegram.start', {})` | on Connect click (Telegram only) | nudge daemon — untyped wrapper today |

| Local store | Polling |
|---|---|
| none | none |

The shipped contract is onMount-only. APPFLOW §4.7 mentions a 10 s `channels.status` poll; the current code does not poll. Tracked as drift in §8.

**Not fetched (intentional):** user content / messages (audit chain handles that), provider list (Chat route), skill / sync data (separate routes), BotFather URL (hard-coded `https://t.me/BotFather`).

---

## 7. DESIGN DECISIONS

| Decision | Why | Anchored in |
|---|---|---|
| One wired channel in v0.1.0; four honestly marked `soon` | No fake enthusiasm | MOAT §4.6 · CLAUDE.md Decision #14 |
| `v0.2.0` pill, not a "Coming Soon" button | `not-allowed` cursor + dim row (opacity 0.55); mono uppercase `--content-faint` on `--hair` border | MOAT §4.1 + §4.6 |
| Cellular-bar cascade is the signature motion for this surface | 5 stepped dots breathing left-to-right — "signal is live" lives nowhere else | MOAT §3 (one unmistakable thing) |
| Footer always shows consent + audit thread | I5 (7 invariants are visible); safety posture legible without scolding | CLAUDE.md §2.1 #3 · APPFLOW §1 I5 |
| Defaults stay rendered on daemon-down | I4 (every state reachable); honest degradation, not dead wall | APPFLOW §6.1 |
| Connect opens BotFather → flip to `degraded` | Token paste happens in bot's chat; "degraded" is honest: "you've started, token isn't in yet" | APPFLOW §4.7 |
| Hint `token entry → open Channels` | Three words, no exclamation, no emoji | DIRECTION.md §6 Rule 2 + §1 |
| No glassmorphism on row | Paper-on-paper; a list of paper rows doesn't earn elevation | MOAT §4.3 |
| Focus ring tracks rounded shape | `--shadow-focus` / pill ring; never rectangular outline | DIRECTION.md §6 Rule 6 · MOAT §2.1 |
| Soon rows cannot lift on hover | Affordance honesty: non-interactive things shouldn't pretend | DIRECTION.md §1 |
| `Pulse` + mono label, never a spinner | Label teaches, Pulse breathes; loading as verb | DIRECTION.md §6 Rule 7 · MOAT §2.5 |
| Error block uses `err-hair` | The Thread is the design system's completion gesture; even errors finish | MOAT §3 |
| No 3D tilt, no rotateX, no decorative SVG | One metaphor per component; cellular-bar carries the load | DIRECTION.md §6 Rule 4 · MOAT §1.5 / §1.6 |

---

## 8. DRIFT TABLE

### 8.1 Removed (vs. earlier draft or natural temptation)

| Item | Why removed |
|---|---|
| Spinner in loading block | DIRECTION §6 R7 + MOAT §4.7 — no spinner; replaced with `Pulse` + mono label |
| "Welcome to Channels" eyebrow | DIRECTION §6 R5 — no "welcome to the future"; replaced with `— Reach · on your terms` (noun phrase that teaches) |
| "Slack is coming soon!" with emoji | MOAT §4.6 + DIRECTION §6 R2 — replaced with dimmed row + mono `v0.2.0` pill + hint `coming in v0.2.0` |
| Rotating-glow on connected rows | MOAT §4.10 — animation must carry meaning; replaced with cellular-bar breathe cascade |
| Backdrop-blur scrim over the page | MOAT §4.3 — a list of paper rows doesn't earn elevation |
| Custom `box-shadow` stacked on `--shadow-card` for row hover | MOAT §4.9 — no double shadows; the `-1px` lift alone reads as elevation |
| `cursor: pointer` on `.row` | A row is not a button; only the Connect affordance is |
| "Last sync 2 minutes ago" status bar | Not exposed by `reach` subsystem; would lie. Honest surface ships only what daemon returns |

### 8.2 Added

| Item | Why |
|---|---|
| Footer signature Thread | MOAT §3 — every surface ships ≥1 thread |
| Italic err-block with `err-hair` | Consistent with Chat / Skills (pending `ErrorState` extraction) |
| Per-row focus halo on `border: --synapse` | Focus tracks the rounded shape (DIRECTION §6 R6) |
| `noteIn` rAF toggle for footer Thread | Draws in after first paint — "data has arrived" gesture |
| `:global(:root[data-energy='low'])` cascade-disable | Battery-aware (CLAUDE.md §6.4 · DIRECTION §5) |
| `--warn` for degraded (distinct from `--pollen`) | Status-vs-brand rule (DIRECTION §4) |
| Reduced-motion override for cascade | Accessibility — `prefers-reduced-motion: reduce` honored |
| `cursor: not-allowed` on `.row.soon` | Affordance honesty |
| `aria-label` per row | Screen-reader contract: `${name}, ${stateLabel}` |
| `aria-live="polite"` on error block | Announce daemon failure once, not every re-render |
| `aria-hidden="true"` on cellular-bar | State lives in `aria-label`; dot array is decorative |

### 8.3 Known drift (current `Channels.svelte` vs. spec)

| ID | Item | Source | Resolution |
|---|---|---|---|
| D-CH-01 | Component-local `prefers-reduced-motion` block (lines 444–451) duplicates the global block in `condura.css:469–476` | MOAT §2.3 | Move into global block; remove local override (v0.2.0 cleanup) |
| D-CH-02 | `err-hair-draw` keyframe re-declared locally (lines 369–371) | MOAT §1.2 | Extract to `condura.css` (v0.2.0 cleanup) |
| D-CH-03 | Inline `.err-state` block — same shape in Chat / Skills | MOAT §1.2 | Extract `<ErrorState>` (v0.2.0 cleanup) |
| D-CH-04 | `.row` re-declares `transition: ...` list (lines 224–229) | MOAT §2.7 | Use global `.tactile` (v0.2.0 cleanup) |
| D-CH-05 | No 10 s `channels.status` poll | APPFLOW §4.7 | Add `setInterval(channelsStatus, 10000)` in `onMount`, clear in `onDestroy` (v0.2.0) |
| D-CH-06 | No typed wrapper for `channels.telegram.start` (raw `ipc.call`) | IPC refactor | Add `ipc.channelsTelegramStart({})` typed wrapper (v0.2.0) |
| D-CH-07 | `Tooltip.svelte` + `Glyph.svelte` not yet integrated | MOAT §2.9 · DIRECTION §6 R2 | Wire per-channel Glyphs next to name; Tooltip on `.pill-soon` (v0.2.0) |
| D-CH-08 | `Manage` button (post-`connected`) not implemented in code | APPFLOW §4.7 | Wire to Settings sub-flow (v0.2.0) |
| D-CH-09 | `⌘T` route-local chord not implemented | spec §4 | Add `keydown` listener scoped to Channels route (v0.2.0) |
| D-CH-10 | Footer Thread hidden during hydration | spec §3.5 / current code | Aligned (intentional; honest degradation) |

---

**This document is the screen. The code is the implementation. When they diverge, the divergence is the bug — fix both in one commit.**