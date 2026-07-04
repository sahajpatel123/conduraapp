# SCREEN_PAIRINGMODAL · Condura

> **Screen architecture spec for `PairingModal.svelte`.** Phase 14F-style
> resurfacing: the v1 file at `app/web/frontend/src/lib/components/`
> `PairingModal.svelte` is a *right-side* `Sheet` (slide-from-right,
> 420px) with a vertical QR-then-PIN column. This spec replaces it with a
> **bottom sheet** — a redesign rooted in `MOAT.md §2.8`'s overlay
> taxonomy (`.c-sheet` slides from edge, doesn't block page scroll)
> and `MOAT.md §3`'s thread-as-signature. The brief asks for a sheet
> (not a modal) because pairing is a **task**, not a **confirmation**.
> A modal would say "wait"; the sheet says "go."
>
> **Reading order for the next agent.** §1 is drift; §2 is geometry;
> §3 is the state matrix; §4 is motion; §5–§9 round out. Skip §10–§13
> only if you already know the app.
>
> **Source-of-truth docs** (read before this one):
> - `MOAT.md` — quality bar. PairingModal must pass §1, §2.6, §3, §4.
> - `APPFLOW.md` §4.4 — Sync route and pending pairing card.
> - `condura.css` §3 — design tokens (`--dur`, `--dur-slow`, `--ease`,
>   `--synapse`, `--pollen`, `--danger`, paper / ink ramp).
> - `Sync.svelte` lines 1–424 — Sync garden + pending-card
>   (the synchronous surface this sheet stacks on).
> - `Glyph.svelte` + `icons.ts` — single-stroke, 1.5-weight, currentColor.
>
> **Existing implementation.** `app/web/frontend/src/lib/components/`
> `PairingModal.svelte` (the v1) is the baseline; this spec describes
> the v2 redesign that re-files it into the `condura/` tree (path
> target: `app/web/frontend/src/lib/condura/PairingModal.svelte`).
> The current `Sync.svelte` ships an embedded pending pairing card
> (lines 309–375 of the Sync route) — PairingModal is the
> **standalone variant** for ad-hoc invocation (e.g. from the floating
> interview, from a deep link, from the titlebar).

---

## Table of Contents

1. [Spec vs. Implementation Drift](#1-spec-vs-implementation-drift)
2. [Layout](#2-layout)
3. [State Matrix](#3-state-matrix)
4. [Motion Choreography](#4-motion-choreography)
5. [Keyboard](#5-keyboard)
6. [Components Used](#6-components-used)
7. [Data Fetched](#7-data-fetched)
8. [Design Decisions (MOAT Compliance)](#8-design-decisions-moat-compliance)
9. [Accessibility Contract](#9-accessibility-contract)
10. [Implementation Notes for the Next Agent](#10-implementation-notes-for-the-next-agent)
11. [Test Plan](#11-test-plan)
12. [i18n Surface](#12-i18n-surface)
13. [Open Questions](#13-open-questions)

---

## 1. Spec vs. Implementation Drift

What this spec changes in the current v1 `PairingModal.svelte`.
The next agent must remove the old, ship the new, in one atomic diff
— no half-states (per `MOAT.md §1` restraint).

| # | Today (v1 `components/PairingModal.svelte`) | Phase 14F v2 (this spec) | Why |
|---|---|---|---|
| 1 | Uses `<Sheet side="right" width="420px">` from `components/ui`. Side-drawer style. | **Bottom sheet** (`<Sheet side="bottom" maxHeight="80vh">` — implementing `.c-sheet` per `MOAT.md §2.8`). Slides from bottom edge. | MOAT §2.8 + brief: pairing is a task, not a confirmation. Bottom sheet occupies the user's peripheral vision (where their thumbs are) instead of crowding the sidebar where Sync node garden already lives. |
| 2 | QR + PIN stacked vertically in a single column. | **Two zones side by side on wide viewports (≥720px), stacked on narrow (<720px).** QR left, PIN right. | The two are simultaneous affordances, not sequential. Showing them next to each other lets the user act on the right (read PIN from phone) and the left (scan QR from phone) without losing their place. |
| 3 | QR is 200×200, white-bg, 8px padding. | **QR 240×240 with a 16px paper-raised border (the "paper card" metaphor).** White inner canvas. | Brief + MOAT §5 — premium quality reads as a paper card on a paper surface. The 16px gutter is the same as card padding across the app. |
| 4 | PIN block has a 84×84 SVG ring (32 radius), 4px stroke, 950ms linear transition. | **PIN ring 120×120 SVG (48 radius), 3px stroke, 500ms ease + `pathLength: 1` so it composes with the rest of the thread grammar.** Stroke color is `--pollen`; flips `--danger` at <30s (the `ttl-warning` state). | The ring is **the signature motion** of the modal (per brief). It should be the visual identity, not a utility. 84px is too small to be felt. |
| 5 | No QR-vs-PIN toggle. Both render at all times. | **"Show the PIN instead →" / "Show the QR instead →" toggle button** at the bottom of each zone. Default = both visible. Toggle hides the inactive zone (200ms cross-fade). | Accessibility — some users can't scan QR (camera broken, low-vision, screen reader). The toggle is honest (per MOAT §4 #5 copy rule). |
| 6 | `<input class="pin-input">` for the *local* user's PIN entry; the modal shows the *remote* PIN as read-only text. | **No local PIN input.** The whole point of P2P pairing is that the user reads this PIN on the *other* device. The modal shows this device's QR + the PIN to read; the *peer* enters the PIN on its own end. The CTA is **"I've entered the PIN on the other device →"**. | The v1 has the modal's local user typing the PIN — that's the wrong direction. Pairing is bidirectional: this side displays, peer side types. |
| 7 | "Sealing…" / "Seal link" CTA text. | **"I've entered the PIN on the other device →"** as primary CTA when peer has typed, **"Regenerate PIN"** as the secondary when TTL < 30s, **"Wait — I haven't done it yet"** as ghost when TTL is healthy. | The CTA's text encodes the actual moment. (MOAT §4 #5 — copy that tells the user what this surface does and what to do next.) |
| 8 | No "PIN copied" confirmation. | **"Copy PIN"** ghost button + ephemeral "PIN copied" confirmation chip (200ms fade, no modal). | Discoverable accessibility aid; some users prefer to paste. |
| 9 | Polls `sync.pair_confirm` every 5s while open. | **Subscribes to `sync.pair_status` SSE** (the live event). The modal updates the moment the peer types the PIN on its device, not on a 5s tick. | The modal is the moment. 5s of blindness after the peer confirmed is a `$50M-feel` loss. |
| 10 | Polls `sync.pair_status` to detect expiry. | **SSE-driven TTL** — the daemon's `expires_at` is the source of truth. The ring depletes toward `expires_at` with the rAF loop taking only the milliseconds-from-now calculation. | The v1 poll is two parallel timers. SSE is one. |
| 11 | Two `onDestroy` blocks (a Svelte lint hit). | **One `onDestroy`.** All timers / event listeners cleaned up in one place. | Mechanical correctness. |
| 12 | Sheet root uses the bundled v1 `components/ui/Sheet.svelte`. | **Implements the `.c-sheet` primitive from `condura.css`** directly in this file (per MOAT §2.8 — `c-sheet` is the primitive; the modal owns it). The animated draw from the bottom is the signature move. | Phase 2 already ships `.c-sheet` in tokens. The v1 imported a different Sheet from a different ui library, breaking the design-system contract. |
| 13 | Hard-coded radius / spacing (`var(--space-4)`). | **All spacing from `--space-*` tokens; all radii from `--r-*` tokens.** Audit against `condura.css` before commit. | MOAT §2.7 — one tactile vocabulary. |
| 14 | Two duplicate `onDestroy` handlers (lines 118 + 176). | **One handler, one cleanup, one truth.** | Mechanical correctness. |
| 15 | No announcement of pairing-complete in the modal itself. | **Success state morphs the content** (the sheet's children cross-fade to a "Paired." check + the new peer name), then **auto-dismisses after 1.5s** (per brief). | The user does the work; they see it land. |
| 16 | `<input>` receives Enter only via onkeydown. | **Full keyboard model** in §5; ⌘P to open, Esc to close, Tab between two zones + toggle + copy + regenerate + CTA, Enter on the CTA confirms. | MOAT §2.10 — the modal ships its keyboard story, not just Esc-to-close. |
| 17 | Error is rendered inline as `<p class="err">`. | **One instance of `ErrorState` consumed** — see MOAT §1.2 / §2.6: headline (italic Instrument Serif 22px) + 3 lines (what / cause / action) + `err-hair` that draws left→right. | The v1 violates the four-copies-too-many rule. Replace. |
| 18 | Tooltips missing. | **`<Tooltip>` on the QR toggle, the PIN toggle, the Regenerate button (400ms hover, exit 75ms).** | MOAT §2.9 — no `title=`. |

---

## 2. Layout

### 2.1 Geometry

The modal is a **bottom sheet** anchored to the viewport edge. It does
**not** block page scroll (per `MOAT.md §2.8` `.c-sheet` taxonomy) and
it does **not** focus-trap the page; the Sync garden underneath stays
alive and breathing (the sync-breathe 9s loop never stops).

```
Viewport (1440 × 900 example)
┌───────────────────────────────────────────────────────────────┐
│                                                               │
│   Sync garden (this device centered, peer nodes around it)    │
│   ─────────────────                                             │
│                                                                 │
│                                                                 │
│                                                                 │
│   ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒ soft scrim 12% paper-ink ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒   │
│   ▒                                                          ▒│
│   ▒   ┌─────────────────────────────────────────────────┐    ▒│
│   ▒   │ ───  drag handle (32 × 4 hairline pill)         │    ▒│
│   ▒   ├─────────────────────┬───────────────────────────┤    ▒│
│   ▒   │                     │                           │    ▒│
│   ▒   │   QR (240 × 240)    │    PIN (display-mono)     │    ▒│
│   ▒   │   on white card     │    + TTL ring (120 × 120) │    ▒│
│   ▒   │                     │                           │    ▒│
│   ▒   │   "Show PIN ↗"      │    "Copy PIN"             │    ▒│
│   ▒   ├─────────────────────┴───────────────────────────┤    ▒│
│   ▒   │  ✓  paired ↗  ·  ↻ regenerate  ·  ⌘P to summon   │    ▒│
│   ▒   └─────────────────────────────────────────────────┘    ▒│
│   ▒                                                          ▒│
│   └──────────────────────────────────────────────────────────┘│
└───────────────────────────────────────────────────────────────┘
                  ▲ max-width 560px; centered horizontally
                  ▲ max-height 80vh; sits flush against the bottom
```

### 2.2 Dimensions

| Property                       | Value                              | Why                                                            |
|---|---|---|
| `position`                     | `fixed; inset: auto 0 0 0;`        | Anchored to bottom edge. `.c-sheet` from MOAT §2.8.            |
| `max-width`                    | **560px**                          | Two zones side-by-side at this width and above.                |
| `width`                        | `min(560px, calc(100vw - 32px))`   | 16px gutter on narrow viewports.                               |
| `max-height`                   | **80vh**                           | Tall content (PIN + footer + scrim) fits in laptops & desktops.|
| `margin-inline`                | `auto`                             | Centered horizontally (the sheet is wider than the QR).        |
| `border-radius`                | `var(--r-lg) var(--r-lg) 0 0`      | Rounded only on top — bottom corners are flush to viewport.    |
| `background`                   | `var(--surface-raised)`            | One elevation above the Sync garden beneath.                   |
| `box-shadow`                   | `var(--shadow-float)`              | Exactly one shadow per MOAT §4 #9. **No layered box-shadow.**  |
| `border-top`                   | `1px solid var(--hair-strong)`     | Hairline separates the sheet from the scrim above.             |
| `padding`                      | `var(--space-5) var(--space-5)`    | Symmetric interior.                                            |
| `z-index`                      | `var(--z-sheet)` (300)             | Above Sync canvas, below consent modal (2000).                 |

### 2.3 The drag handle

A 32×4 hairline pill centered at the top of the sheet. The handle is the
**only signal that the sheet is dismissable by drag-down** — but in
practice the user dismisses with Esc. The handle is honest affordance,
not a feature.

| Property             | Value                              |
|---|---|
| Width × height       | 32 × 4                             |
| `background`         | `var(--ink-faint)`                 |
| `border-radius`      | `var(--r-pill, 999px)`             |
| `margin`             | `8px auto 12px`                    |
| Animation            | Opacity 0 → 1 over 200ms **with a 200ms delay** (after the slide-up completes — see §4.1). |

The handle has no `cursor: grab` (per MOAT §4 — non-affordance
decoration). It is the visual marker; the actual close is Esc / click
outside.

### 2.4 The two zones (wide layout, ≥720px viewport)

Inside the sheet, below the handle:

```
┌─────────────────────────┬───────────────────────────┐
│                         │                           │
│  ZONE A · QR            │  ZONE B · PIN             │
│  ──                     │  ──                       │
│                         │                           │
│  ┌─────────────────┐   │   ┌───────────────┐       │
│  │                 │   │   │               │       │
│  │  QR 240 × 240   │   │   │  PIN ring     │       │
│  │  (white card,   │   │   │  120 × 120    │       │
│  │   16px pad)     │   │   │  + PIN mono   │       │
│  │                 │   │   │  + TTL text   │       │
│  └─────────────────┘   │   └───────────────┘       │
│                         │                           │
│  Scan with phone        │   Or type the PIN         │
│                         │   on the other machine    │
│                         │                           │
│  ┌─────────────┐       │   ┌─────────────┐         │
│  │ Show PIN ↗ │       │   │ Copy PIN    │         │
│  └─────────────┘       │   └─────────────┘         │
│                         │                           │
└─────────────────────────┴───────────────────────────┘
```

**Zone A · QR (240 × 240)**

| Property      | Value                            |
|---|---|
| Card width    | `240px`                          |
| Card height   | `240px`                          |
| Background    | `#fff` (pure paper-white)        |
| Padding       | `var(--space-4)` (16px)          |
| Border-radius | `var(--r-md)` (12px)             |
| Box-shadow    | **Single** `var(--shadow-card)` (per MOAT §4 #9) |
| Image         | `<img>` of data-URL from `qrcode` package, with `margin: 1, width: 240` (qrcode options). |
| Caption       | Below the card, centered, italic Instrument Serif 14px in `--content-mute`: **"Scan with phone"** (one noun + verb; nothing else). |

**Zone B · PIN (display-mono + ring)**

| Property        | Value                                                |
|---|---|
| Ring dimensions | `120 × 120` SVG; `cx: 60, cy: 60, r: 48`              |
| Ring stroke     | `--pollen` (`#C97B2E`); flips `--danger` at TTL < 30s |
| Ring weight     | `3px`                                                |
| Ring track      | `var(--hair)` 1.5px, opacity 0.45                     |
| Ring transition | `stroke-dashoffset var(--dur) var(--ease)` (280ms)    |
| Ring rotation   | `transform: rotate(-90deg)` on the SVG center        |
| PIN text        | `--font-mono`, `26px`, `letter-spacing: 0.18em`, `--content` |
| PIN glow        | `text-shadow: 0 0 22px color-mix(in oklab, var(--pollen) 25%, transparent)` |
| Caption         | **"Or type the PIN on the other machine."** (italic Instrument Serif 14px) |

The ring is **the signature motion** of the modal. It must compose
with the rest of the thread grammar: `pathLength: 1`,
`stroke-dasharray: 1`, `stroke-dashoffset: <1 - ringPct/100>`,
`filter: drop-shadow(0 0 4px color-mix(in oklab, var(--pollen) 35%, transparent))`
(the same halo vocabulary as the Channels signal-bars).

### 2.5 The two zones (narrow layout, <720px)

The zones **stack vertically**. Zone A on top (QR + caption + toggle),
Zone B below (PIN ring + caption + toggle), separated by a 1px
hairline in `--hair` that draws left→right as the sheet enters. The
hairline is the **same** drawing motion used by Routes on enter —
consistency for free.

| Property    | Value                                                           |
|---|---|
| `flex-direction` | `column`                                                    |
| Stack order | QR (top) → hairline → PIN                                       |
| Each zone   | `width: 100%; max-width: 280px; margin-inline: auto;`           |

At very narrow viewports (< 420px) the QR shrinks to 200×200 to fit.

### 2.6 The footer

A single horizontal row below the two zones. Three slots:

```
┌──────────────────────────────────────────────────────────────┐
│ ✓ paired state   ·   ↻ regenerate   ·   ⌘P to summon again    │
└──────────────────────────────────────────────────────────────┘
```

| Slot         | Content                                                                        |
|---|---|
| Left         | **Live status pill** (see §3 for variants: `paired · sealed`, `awaiting peer`, `expired`, `error`). Pulse + mono 10px uppercase in `--content-mute`. |
| Center       | **"Regenerate PIN"** ghost button (visible only in TTL-warning / TTL-timeout states). |
| Right        | **"⌘P to summon again"** mono 10px hint (visible only when the sheet is dismissed and then re-invoked manually). |

Below the footer, a single-line 1px hairline in `--hair` (no shadow,
no border-radius). Above the hairline (still inside the sheet), the
"What gets synced / what doesn't" reminder.

### 2.7 What gets synced (the footer reminder)

A two-line italic Instrument Serif block (15px / 13px). Always
rendered except in the success state (when it cross-fades out). Wording
is exact:

> **Synced between paired devices:** memory, skills, config.
>
> **Never synced:** logs, audit, screenshots, API keys. Always local.

This is the trust-on-display. The user reads it once; the page
already aligns with CLAUDE.md §17.4 ("What's NOT synced: logs, audit,
API keys"). The reminder is a load-bearing promise, not decoration.

---

## 3. State Matrix

The modal has 8 visible states plus 1 closed (unrendered). Each is a
real state; every one is reachable.

### 3.1 closed (S0) — not rendered

**Visual signature:** `display: none`. No DOM, no animations, no
listeners. The Sync garden underneath is fully interactive.

**Trigger:** initial (`open={false}` from `props`), or Esc / ⌘. / click
outside / 1.5s after `paired` auto-dismiss.

### 3.2 open (S1) — sheet up, both zones visible (default state)

**Visual signature:** sheet sliding in from the bottom. Both QR and PIN
render with their `data-default-visible="true"`. The toggle pills
inside each zone read **"Hide ↗"** to indicate the action of hiding
them (the affordance label reveals the consequence).

**Loading:** while the QR data-URL is being generated, the QR card
renders a `Pulse phase="thinking" size={10}` centered on a `--hair`
rectangle. No spinner (per `MOAT.md §4 #7`). The Pulse fades when the
`<img>` `onload` fires.

**Trigger:** `open={true}` prop, regardless of TTL state.

### 3.3 qr-mode (S2) — toggle hidden PIN zone

**Visual signature:** Zone B is opacity 0 + scale 0.96 over 200ms.
Toggle pill in Zone A reads **"Show PIN ↗"**. The PIN toggle replaces
the "Hide" affordance with a verb that promises re-entry.

**Trigger:** user clicks "Hide" on Zone B, or "Show PIN ↗" on Zone A.
Both buttons toggle the same bit.

### 3.4 pin-mode (S3) — toggle hidden QR zone

**Visual signature:** Zone A is opacity 0 + scale 0.96 over 200ms.
Toggle pill in Zone B reads **"Show QR ↗"**.

**Trigger:** user clicks "Hide" on Zone A, or "Show QR ↗" on Zone B.

### 3.5 ttl-warning (S4) — TTL < 30s

**Visual signature:** the ring stroke flips from `--pollen` to
`--danger` over 200ms. The PIN text glow turns `--danger` (same
transition). The footer **center slot** animates in a **Regenerate PIN**
ghost button. The footer left slot reads: **"EXPIRING · 0:24"** (mono
10px, `--danger`).

**Trigger:** `secondsLeft > 0 && secondsLeft < 30`.

### 3.6 paired (S5) — successful handshake

**Visual signature:** the two zones **cross-fade out** as a single block
(opacity 1 → 0 over 200ms). In their place, a 360ms **Thread draws
across the sheet's width** in `--synapse`, then a pollen check mark
stamps itself (24×24 circle, scale 0 → 1 with a 320ms backOut easing).
Below the check, italic Instrument Serif 22px reads **"Paired — sealed."**
in `--synapse`. The footer reads **"1 device paired with this machine"**.

**Auto-dismiss:** 1500ms after the check stamps (`MOAT §4 #6` —
no celebration; one word of body copy "sealed" carries the moment).
The sheet slides back down (`--dur-slow`, same `cubic-bezier(0.22, 1, 0.36, 1)`).

**Trigger:** `sync.subscribe` SSE event `paired` arrives, OR
`sync.confirmPairing` resolves with `{ status: 'paired', device: ... }`.

### 3.7 error (S6) — daemon unreachable / RPC failure

**Visual signature:** an **ErrorState** instance (per `MOAT.md §2.6`)
renders below both zones. Italic Instrument Serif 22px headline +
3 lines:

1. **What failed** in one noun: **"Pairing handshake."**
2. **Likely cause** in one phrase (carried from raw error message):
   *"daemon was unreachable"* OR *"the PIN didn't match"*.
3. **Next action** as a button: **"Try again"** (calls
   `sync.generatePairingToken` again) primary, **"Cancel"** ghost.

Plus the standard 1px err-hair drawing left→right in `--danger` at
120ms delay, 600ms duration.

**Trigger:** `sync.generatePairingToken` rejects, OR
`sync.subscribe` SSE emits `error`.

### 3.8 timeout (S7) — PIN TTL hit 0

**Visual signature:** the ring is **filled with `--danger`** (full
circle, no depletion — the moment zeroed out). The PIN text blurs
(`filter: blur(2px)`) and fades to `opacity: 0.4` over 320ms. The
toggle pills hide. The footer center slot shows **"Regenerate PIN →"**
as a primary button. The footer left slot reads **"EXPIRED · regenerate?"**
in `--danger` mono 10px.

**Trigger:** `secondsLeft === 0 && pending`.

**Recovery:** clicking Regenerate calls `sync.generatePairingToken`. A
new PIN replaces the old one (with cross-fade). The CTA returns to the
ghost state; the toggle pills re-appear.

### 3.9 Cross-state reference table

| State → Trigger                | S0 | S1       | S2    | S3    | S4     | S5       | S6    | S7      |
|---|---|---|---|---|---|---|---|---|
| `open=true` prop               | →S1 | —       | —    | —    | —     | —       | —    | —      |
| Click outside / Esc            | —   | →S0     | →S0  | →S0  | →S0   | —(autodismiss|→S0| →S0   |
| Toggle QR zone                 | —   | S1 ⇄ S2 | —    | —    | —     | —       | —    | —      |
| Toggle PIN zone                | —   | S1 ⇄ S3 | —    | —    | —     | —       | —    | —      |
| `secondsLeft < 30 && > 0`      | —   | →S4     | →S4  | →S4  | —     | —       | →S4  | —      |
| SSE `paired` event             | —   | →S5     | →S5  | →S5  | →S5   | —       | —    | —      |
| SSE `error` or RPC reject      | —   | →S6     | →S6  | →S6  | →S6   | —       | —    | —      |
| `secondsLeft === 0`            | —   | →S7     | →S7  | →S7  | →S7   | —       | →S7  | —      |
| Regenerate PIN (success)       | —   | →S1     | →S1  | →S1  | →S1   | —       | —    | →S1    |
| 1.5s after `paired`            | —   | —       | —    | —    | —     | →S0     | —    | —      |

---

## 4. Motion Choreography

### 4.1 Open (`open=true` mount)

```
0ms         →  backdrop scrim opacity 0 → 1 over 280ms (var(--dur))
0ms         →  sheet transform: translateY(100%) → translateY(0)
              over 320ms var(--ease) (the entry is a slide, not a fade)
200ms       →  drag handle opacity 0 → 1 over 200ms (delayed reveal)
240ms       →  QR <img> opacity 0 → 1 over 200ms (post-slide fade-in)
300ms       →  PIN text scale 0.92 → 1 over 240ms (gentle settle)
280ms       →  footer hairline left→right draw over 520ms (--dur-slow)
```

**No fade-only.** A slide tells the user where the sheet came from
(MOAT §4 #10 — animation carries meaning). A fade would orphan the
modal in space. The slide is brief (320ms) so it doesn't feel slow.

### 4.2 Toggle QR ⇄ PIN (S1 ⇄ S2 ⇄ S3)

The hidden zone does `opacity: 1 → 0 + transform: scale(1) → scale(0.96)`
over 200ms. The revealed zone mirrors in reverse. **The two zones never
overlap** — the toggle is exclusive-but-could-be-reversed, never
simultaneous-hide-and-reveal.

### 4.3 TTL ring (S1 → S4 transition at < 30s)

| Phase           | Motion                                                                 |
|---|---|
| 60s → 30s       | Ring depletes stroke-dashoffset 0 → 0.5 in `--pollen`. 500ms ease on each rAF tick (the ring reads as a slow countdown). |
| 30s onward      | Ring stroke flips `--pollen` → `--danger` over 200ms. The PIN text glow flips in the same transition. |
| 0s              | Ring fills `--danger` (full circle, no depletion). PIN text blurs + fades (see §3.8). |

The ring uses the **same rAF loop** as the titlebar thread: a single
`requestAnimationFrame` callback that:
1. Computes `secondsLeft = max(0, expiresAt - Date.now())`.
2. Updates the ring's `stroke-dashoffset` to `1 - secondsLeft / 60`.
3. Calls `rAF` again **only if** `secondsLeft > 0 && document.visibilityState === 'visible'`.

This drops CPU to 0 in background tabs (per the TitlebarThread
`paused` contract — `intersectionObserver` + `visibilitychange`).
**The current v1 implementation spins a 1s `setInterval` regardless of
visibility; this spec inherits the rAF-with-pause pattern.**

### 4.4 Paired success (S1 → S5)

| t          | Motion                                                                             |
|---|---|
| 0ms        | Two zones: opacity 1 → 0 + scale 0.96 over 200ms (single block, not stagger). |
| 200ms      | Synapse Thread draws across the sheet's full width (top edge, `pathLength: 1`, dashoffset 1 → 0 over 360ms `--ease`). |
| 560ms      | Pollen check-mark stamps (24×24 circle, scale 0 → 1 with `backOut` 320ms). |
| 580ms      | Italic "Paired — sealed." fades in (opacity 0 → 1, 200ms). |
| 1500ms     | Sheet slides down: `translateY(0) → translateY(100%)` over `var(--dur-slow)` (520ms). |

This is the **callback to the titlebar**: the same synapse Thread that
greets the user at the top of the Shell now draws across the modal in
the moment of completion. The reused line is the brand.

### 4.5 Timeout (S1 → S7)

| t          | Motion                                                                                |
|---|---|
| 0ms        | Ring fills `--danger` (the depletion reverses — full circle appears as the timer hits zero; one beat). |
| 0ms        | PIN text: filter blur(0) → blur(2px) + opacity 1 → 0.4, over 320ms. |
| 200ms      | Toggle pills (Zone A & B) hide (opacity 1 → 0, 200ms). |
| 320ms      | Regenerate PIN ghost button fades in (opacity 0 → 1, 240ms). |

### 4.6 Error (S1 → S6)

ErrorState enters via the standard 6px-translate-up + 200ms blur-in.
The `err-hair` draws at 120ms delay, 600ms duration (per `MOAT.md §2.6`).

### 4.7 `prefers-reduced-motion: reduce`

Skips:
- the slide-up entrance (sheet renders at `translateY(0)`, instant)
- the ring animation (static full ring with the numeric countdown text replacing the visual one — readers get clarity, not spectacle)
- the paired success Thread draw (replaced by instant check-mark appearance)
- the timeout blur (PIN cross-fades opacity only, no blur)

Does NOT skip:
- the cross-fade between QR and PIN zones (it's instant anyway)
- the TTL color flip at <30s (semantic; users with vestibular issues still need to see danger)

This contract inherits the global `prefers-reduced-motion` block in
`condura.css:469–476`. Components NEVER re-declare the media query
(per `MOAT.md §2.3`).

---

## 5. Keyboard

The modal ships its own keyboard surface; Esc / Enter are not enough.

| Key             | Where active             | Action                                                       |
|---|---|---|
| `⌘P`            | Anywhere on the Sync route (`#/sync`) | **Open pairing modal** (only when `open=false` and at least one peer is discovered — or when user explicitly presses ⌘P from the Sync route even without discovered peers, which mints a fresh pairing token). |
| `Esc`           | Modal open               | **Close** the modal. Slides down (or instant under reduced motion). Same as click outside. |
| `Tab`           | Modal open               | Cycles through: Zone A toggle → Zone B toggle → Copy PIN → Regenerate PIN (when visible) → primary CTA. **Skip-link to the title** is the first focusable element. |
| `Shift+Tab`     | Modal open               | Reverse cycle. |
| `Enter`         | When CTA button is focused | **Confirms pairing** (calls `sync.confirmPairing`). Same as click. |
| `Space`         | On toggle / ghost / copy / regenerate | Activates the button. |

**Focus management:**
- The first focusable element (the drag handle's aria-label sibling or
  the sheet title) receives focus **240ms after the slide-up completes**
  (so the focus doesn't race the entry animation).
- Focus is **trapped inside the sheet** while open (per WAI-ARIA
  dialog). Tab from the last element wraps to the first.
- On close, focus returns to the element that opened the modal (the
  `Pair` button on a peer chip in `Sync.svelte`, or the global hotkey
  call-site). If no remembered opener, focus returns to the first
  peer-chip Pair button.

**Screen-reader announcements:**
- On open: the `<h1>` (the modal's title) is announced. Wording:
  **"Pair this device with {peer-name}."**
- On TTL warning: announcement fires once at the 30s threshold,
  aria-live="polite": **"Pairing code expires in 30 seconds."**
- On timeout: **"Pairing code expired. Regenerate to continue."**
- On paired: **"Paired with {peer-name}. Closing."**
- On error: announcement uses the standard ErrorState headline.

---

## 6. Components Used

Every component is from the `condura/` tree, **except** the existing
`<Sheet>` from `components/ui` — that one is **replaced** by the inline
`.c-sheet` primitive implementation (per MOAT §2.8 + the brief's
"peripheral vision" rationale).

| Component               | Role                                                                                  |
|---|---|
| `.c-sheet` (inline)     | The bottom-sheet primitive. Owns the scrim, the slide-up, the Esc handler, the focus trap. Implements MOAT §2.8's spec. |
| `Thread.svelte`         | The synapse Thread that draws across the sheet's top in the `paired` state (360ms `pathLength: 1` 1 → 0). **Reused** as a callback gesture. |
| `Pulse.svelte`          | Thinking breath when the QR is generating; awaiting breath when the modal is waiting for the peer to type the PIN on its end. |
| `Glyph.svelte`          | `sync` (modal icon, 14px), `copy` (12px on the copy button), `refresh` (12px on regenerate), `check` (24px on success). All single-stroke 1.5 weight, currentColor. |
| `Button.svelte`         | Primary CTA ("I've entered the PIN on the other device →"), ghost ("Regenerate PIN"), ghost ("Copy PIN"). |
| `Tooltip`               | Hover-400ms on: QR-zone toggle (label: "Switch to PIN code"), copy button (label: "Copy the 6-character code"). |
| `ErrorState`            | The error renderer. Per MOAT §1.2 / §2.6 — one component, used here once. |

**QR generation:** uses the `qrcode` package (already imported in
v1 — line 8 of `components/PairingModal.svelte`). Options:
`{ margin: 1, width: 240, color: { dark: '#0B3D2E', light: '#FFFFFF' } }`
— `--synapse` on pure paper-white. The dark color is the **only**
brand hex in this file (a permitted exception per MOAT §4 #1 — the QR
needs a single dark stroke to be readable by phone cameras).

**Imports disallowed in this file:**
- `Spinner` or any variant (per MOAT §4 #7).
- `<input>` for the PIN (per §1.6 drift fix).
- The v1 `<Sheet>` from `components/ui/Sheet.svelte`.

---

## 7. Data Fetched

The modal is a write surface, not a read surface. It owns three RPCs:

| RPC                          | Direction | Body                                                 | Returns                                                                  |
|---|---|---|
| `sync.generatePairingToken`  | → daemon  | `{ device_id: thisId }`                              | `{ token: string, pin: string, expires_at: ISO8601, peer_id: string }`   |
| `sync.subscribe` (SSE)       | ← daemon  | (open EventSource on mount)                          | Stream events: `pairing_progress`, `paired`, `error`, `peer_typing`     |
| `sync.confirmPairing`        | → daemon  | `{ token: string, pin: string }` (the PEER's pin)    | `{ status: 'paired', device: ... }` or `error: { code, message }`        |
| `sync.regeneratePin`         | → daemon  | `{ token: string }`                                  | `{ token: string, pin: string, expires_at: ISO8601 }`                  |
| `sync.list`                  | → daemon  | (no body)                                            | `{ pairs: [{ device_id, name, paired_at }] }`                          |

The `confirmPairing` call is the **primary action** of the modal. The
CTA copy — **"I've entered the PIN on the other device →"** —
encodes the relationship: this side displays, the peer side types. We
confirm only when the peer has done it. (No orphan auto-confirm.)

### 7.1 Event flow on mount

```
1. Modal opens (props.open=true, or user pressed ⌘P).
2. Modal calls sync.generatePairingToken → returns the {token, pin, expires_at}.
3. Modal subscribes to sync.subscribe({ token }) — the SSE channel.
4. rAF loop starts (paused on visibilitychange).
5. QR <img> renders the data-URL once qrcode.toDataURL() resolves.
```

### 7.2 Event flow on peer types the PIN

```
1. Peer enters PIN on its own device, presses "Submit" there.
2. The peer's device sends the typed PIN to our daemon (over LAN or relay).
3. Our daemon emits a `peer_confirmed` SSE event on our sync.subscribe channel.
4. Modal flips the primary CTA from "Wait — I haven't done it yet" to
   "I've entered the PIN on the other device →" (single enable transition).
5. User clicks the CTA. Modal calls sync.confirmPairing({ token, pin: <their pin> }).
6. Daemon verifies. Returns { status: 'paired' } OR { error }.
7. State transitions S1 → S5 (success) or → S6 (error).
```

### 7.3 Polling fallback

If `sync.subscribe` fails to connect (network / daemon unreachable),
the modal falls back to a **5s poll of `sync.pair_status`** — same
pattern as the current v1. The poll is the safety net; SSE is the
preferred path. If SSE reconnects, the poll stops.

---

## 8. Design Decisions (MOAT Compliance)

| MOAT Rule                                  | How this spec honors it                                                                 |
|---|---|
| §1 Restraint (no over-design)              | Two zones, not a tabbed panel. One line per body copy. No italic-green flourishes.       |
| §1.4 Cursor quill                          | The Cursor.svelte is opt-in; the modal does not opt in.                                  |
| §1.5 No `rotateX` for no reason            | No 3D tilts. The QR card is flat. The PIN ring is flat. `transform: scale(0.96)` at most. |
| §2.1 Focus rings track rounded shapes      | The QR card and PIN ring both have `border-radius ≥ 8px`. Focus halo uses `0 0 0 2px var(--synapse), 0 0 0 5px var(--pollen-halo)` per the fix. |
| §2.2 Press state has weight                | All `<Button>` instances inherit the global `.tactile` class — no per-component transition lists. |
| §2.3 `prefers-reduced-motion`              | Components NEVER re-declare the media query. The global block in `condura.css` handles it. |
| §2.4 Empty states teach (not decorate)     | The "Discovering peers..." empty state on the underlying Sync route tells the user what to do: open Condura on another machine on the LAN. The sheet itself has no empty state — it always has a token, a pin, or an error. |
| §2.5 Loading uses a thread                  | The QR-generation wait shows a `<Pulse>` + "Preparing code…" mono-uppercase. Pulse carries the *thread* signal of data moving. |
| §2.6 Errors guide (not poeticize)          | The ErrorState instance follows the three-line contract: what / cause / action. |
| §2.7 One tactile vocabulary                | All button presses inherit `.tactile`. No redeclaration.                                |
| §2.8 Overlay taxonomy                       | **This is the load-bearing decision.** PairingModal is `.c-sheet`, not `.c-modal`. Pairing is a task; a modal would say "wait"; the sheet says "go." The v1's `Sheet` (side-drawer) was wrong — wrong edge, wrong metaphor. |
| §2.9 Real tooltips, not `title=`           | Hover-tooltips on copy/regenerate/toggles. The v1 used `title="…"` nowhere — but we add explicit tooltips now (this spec is a regression-prevention). |
| §2.10 Keyboard story complete              | ⌘P open, Esc close, Tab order, Enter on CTA. Stated explicitly in §5.                  |
| §3 The signature — the Thread              | The `paired` success state draws the synapse Thread across the sheet — callback to the titlebar. **The sheet reuses the brand.** |
| §4 #1 No gradient text                      | The PIN glow is a `text-shadow` with a single drop-shadow color, not a gradient. The QR is single-stroke on white. |
| §4 #2 No emoji                              | All icons via `<Glyph>`. No `✓` literals (the success check is a `<Glyph name="check">` rendered as a 24×24 stamp). |
| §4 #3 No glassmorphism                      | The sheet uses `var(--shadow-float)`. No `backdrop-filter`. |
| §4 #4 No rainbow                            | Brand is `--synapse` (green) and `--pollen` (orange). Status is `--danger`. Nothing else. |
| §4 #5 Copy that teaches                     | "What gets synced / Never synced" footer reminder; "I've entered the PIN on the other device" CTA; "EXPIRING · 0:24" footer status. **Every label is a noun or a verb.** No "Welcome to pairing!" |
| §4 #6 No fake enthusiasm                    | Success state reads "Paired — sealed." (one word past the noun — "sealed" earns its presence; it's the partnership-and-done moment). No "Awesome!" No "You're all set!" celebration modal. |
| §4 #7 No spinner                            | Loading is `<Pulse>` + thread. The v1 had no spinner either; this spec is the regression-prevention. |
| §4 #8 No rectangular focus outlines         | Pollen halo + synapse ring. Per §2.1 fix. |
| §4 #9 No double shadows                      | Sheet shadow is single (`var(--shadow-float)`). The v1 had a layered `box-shadow` (`var(--shadow-glow)` + manual `padding`) — replaced by the token. |
| §4 #10 Animation carries meaning             | Slide-up = "I came from below" (the entry direction). Thread-draw = "the handshake completed." TTL color flip = "act now." Every motion says something. |
| §5 $50M feel — cursor on hoverable          | The hoverable surfaces (QR card, PIN block, CTAs) all set `document.body.dataset.hover='1'` via `use:hoverRegion`. The Cursor dot becomes a pollen ring. |
| §5.2 Composer focus draws a thread          | The Copy PIN input (when present) inherits the same draw-thread-on-focus animation. |
| §5.5 The Ritual on phone                    | n/a — this modal is desktop only (the Wails shell). |
| Signature decision: bottom sheet            | The brief is explicit and MOAT §2.8 is explicit. This isn't decoration; it's the taxonomy. |

---

## 9. Accessibility Contract

| Concern           | Implementation                                                                                |
|---|---|
| **Keyboard**      | See §5. Full support: ⌘P open, Esc close, Tab/Shift+Tab cycle, Enter confirm, Space on toggles. |
| **Focus trap**    | WAI-ARIA dialog pattern. Tab from last loops to first; focus returns to opener on close.        |
| **Screen reader** | `<h1>` announced on open. Live-region announcements for TTL, timeout, paired, error.           |
| **Color contrast** | PIN text `--content` on `--surface-raised` = 14:1 minimum (paper-raised is `#FBF7EC`, ink is `#14110B`). Ring on track: 4.5:1 minimum. |
| **Dark mode**     | Inherits `--content`, `--surface-raised`, `--paper-3` from `[data-mode='dark']` overrides in `condura.css`. Ring color flips to `--pollen` (light) / `--pollen-light` (dark). |
| **Reduced motion** | §4.7 — slide entry, ring, paired Thread, blur all skipped. Color flips preserved.        |
| **Cognitive accessibility** | The QR-vs-PIN toggle is the elevator for users who can't scan a QR. The "Or type the PIN on the other machine" caption is a sentence, not an icon. The "What gets synced / Never synced" footer is a reading exercise, not an interaction. |
| **No-required-color motion** | Status is encoded in BOTH color and text label (`EXPIRING`, `EXPIRED`, `paired`). Colorblind users get the text. |

---

## 10. Implementation Notes for the Next Agent

1. **Re-file the source.** Move `app/web/frontend/src/lib/components/`
   `PairingModal.svelte` → `app/web/frontend/src/lib/condura/`
   `PairingModal.svelte`. Update imports in `Sync.svelte` (which is the
   only consumer that might import it via the legacy path — confirm
   first).

2. **Replace the Sheet import.** Drop the import of `<Sheet>` from
   `components/ui/`. Implement `.c-sheet` directly: a `<dialog
   class="c-sheet">` with `position: fixed; inset: auto 0 0 0;` and
   the geometry in §2.2. Add the inline focus-trap helpers
   (`firstFocusable`, `lastFocusable`, `tab` event handler). Or, if a
   shared `.c-sheet` primitive already exists in `condura/`, import
   that (per MOAT §2.8 consolidation — single source for the
   taxonomy).

3. **AES-style use of the QR code.** Don't compute the QR from
   `JSON.stringify({ v, device_id, name })` directly — that was the
   v1. Define `payload = { v: 1, device_id, name, ttl: 60 }` so the
   scanner knows the format.

4. **The SSE channel** is opened via `new EventSource('/sync/' +
   token + '/events')` (the exact URL lives in the daemon contract —
   check `internal/daemon/methods_phase12.go` for the route). On
   unmount or close, close the EventSource explicitly.

5. **The rAF ring loop** must pause on `visibilitychange` and on
   `IntersectionObserver` of the sheet (when not visible). Reuse the
   pattern from `TitlebarThread.svelte:12` to avoid re-inventing the
   pause-handling.

6. **The cross-fade between QR and PIN** must respect `prefers-
   reduced-motion` — the global block handles it, but verify the
   toggle buttons themselves don't re-declare the media query.

7. **Glyph icons.** Use:
   - `name="sync"` for the modal title icon (14px, `--content-mute`)
   - `name="copy"` for Copy PIN (12px, currentColor)
   - `name="refresh"` for Regenerate (12px, currentColor)
   - `name="check"` for the success stamp (24px, `--synapse`)

8. **No `'use client'`, no prop drilling the IPC handlers.** The
   modal is presentation; it consumes the `sync` store (from
   `stores/sync.svelte.ts`) directly for `pair_with`,
   `confirm_pairing`, `clear_pending`, etc. The store owns the
   IPC.

9. **One onDestroy.** All timers / SSE channels cleaned up in one
   block (v1 had two — a Svelte lint hit).

10. **Run the existing tests.** `app/web/frontend/src/lib/condura/`
    has a Vitest setup (per CLAUDE.md §33.5.5 SB-09 fix). Add a
    `PairingModal.test.ts` covering the state transitions: S0 → S1
    on open, S1 → S4 at < 30s, S5 → S0 after 1.5s, S7 → S1 on
    Regenerate. The Token contract (`PinDisplay` only renders
    display-mono text), the TTL contract (ring's
    `stroke-dashoffset` matches `1 - secondsLeft/60` ± 0.01), and
    the SSE handler (subscribes on mount, closes on destroy).

---

## 11. Test Plan

A short Vitest suite (`condura/PairingModal.test.ts`) covering:

```ts
// token-contract tests (mount-level)
- Default open (S1) renders both zones.
- Both zones have role="region" and aria-labelledby.
// state-transition tests (component-driven)
- ⌘P synthesised keypress on Sync route triggers open().
- Esc on the modal triggers close() and slides down.
- TTL crosses 30s threshold → ring stroke flips to var(--danger).
- SSE emit("paired") → success state → auto-dismiss after 1500ms.
// lifecycle / no-leak tests
- onDestroy closes the SSE channel and clears the rAF loop.
// reduced-motion tests (matchesMedia mock)
- prefers-reduced-motion: reduce → entry has no slide (translateY(0)).
- Same flag → TTL ring is static (no stroke-dashoffset transition).
// keyboard story
- Tab cycles through the two zones, the toggles, copy, regenerate, CTA.
- Enter on CTA calls sync.confirmPairing().
```

---

## 12. i18n Surface

Strings the modal renders, in i18n keys (existing key namespace
`sync.pair.*` from `i18n/en.json`):

| Key                          | EN copy                                                          |
|---|---|
| `sync.pair.title`            | "Pair with {peerName}"                                            |
| `sync.pair.qr_cap`           | "Scan with phone. {deviceName}."                                  |
| `sync.pair.pin_label`        | "OR ENTER THIS PIN ON THE OTHER MACHINE"                          |
| `sync.pair.show_qr`          | "Show QR ↗"                                                       |
| `sync.pair.show_pin`         | "Show PIN ↗"                                                      |
| `sync.pair.hide`             | "Hide ↗"                                                          |
| `sync.pair.copy`             | "Copy PIN"                                                        |
| `sync.pair.copied`           | "PIN copied."                                                     |
| `sync.pair.cta`              | "I've entered the PIN on the other device →"                      |
| `sync.pair.cta_waiting`      | "Wait — I haven't done it yet"                                    |
| `sync.pair.regenerate`       | "Regenerate PIN"                                                  |
| `sync.pair.paired`           | "Paired — sealed."                                                |
| `sync.pair.paired_sub`       | "1 device paired with this machine."                              |
| `sync.pair.ttl_warning`      | "EXPIRING · {remaining}"                                          |
| `sync.pair.ttl_expired`      | "EXPIRED · regenerate?"                                           |
| `sync.pair.expires_in`       | "Expires in {remaining}"                                          |
| `sync.pair.expired`          | "This PIN expired"                                                |
| `sync.pair.error_headline`   | "Pairing handshake."                                              |
| `sync.pair.error_action`     | "Try again"                                                       |
| `sync.pair.cancel`           | "Cancel"                                                          |
| `sync.pair.hint`             | "Read the PIN on the peer device, then type it there."            |
| `sync.pair.synced_label`     | "Synced between paired devices: memory, skills, config."          |
| `sync.pair.never_synced`     | "Never synced: logs, audit, screenshots, API keys. Always local."  |
| `sync.pair.qr_alt`           | "QR code containing this device's identity"                       |

Keys `sync.pair.title`, `sync.pair.qr_cap`, `sync.pair.pin_label`,
`sync.pair.cta`, `sync.pair.expired`, `sync.pair.expires_in`,
`sync.pair.busy`, `sync.pair.confirm` already exist (v1 §8 used them).
The new keys (CTA text, paired state, regenerate, footer reminder)
are additions for v2 — add to `i18n/en.json` first, then
`i18n/{es,fr,de,ja,zh}.json` (translations out of scope per
`CLAUDE.md §23`; Crowdin in v0.2.0).

---

## 13. Open Questions

1. **Sheet from Sync route vs from titlebar global hotkey.** Sync
   ships the embedded pending card today (`Sync.svelte:309–375`).
   Should the standalone PairingModal also be summonable from the
   titlebar (⌘P), or only from a peer-chip Pair click? The brief
   says "⌘P to open (when on the Sync route)" — confirming the
   "from Sync only" path. **Recommend:** ⌘P is route-scoped (matches
   current brief). v0.2.0 may globalize.
2. **PIN length.** Current daemon returns a 6-character PIN. The
   input field on the v1 accepts `^\d{4,8}$`. v2 displays only
   (no local input) but the PIN copy button should preserve whatever
   length the daemon mints (already 6). Confirm with the daemon
   contract in `internal/sync/`.
3. **Timeout regeneration UX.** When `secondsLeft === 0`, the
   Regenerate button is the primary CTA. Should it auto-fire after
   a 5s grace period, or wait for the user? **Recommend:** wait.
   Auto-regeneration would defeat the user seeing the timeout — a
   `$50M-feel` loss (MOAT §5).
4. **Error: peer already paired elsewhere.** What happens if
   `sync.confirmPairing` returns `{ error: 'peer_paired_elsewhere' }`?
   The current error spec says "the PIN didn't match" — that's the
   simple fail case. The "paired elsewhere" case deserves its own
   state copy: "That device is paired with another machine." Plus a
   **"Forget the link"** ghost button. **Out of scope for v2; flagged.**
5. **SSE auth.** Pairing tokens are scoped to the device. The SSE
   channel authenticates via the token — confirm the daemon honors
   this without an additional cookie. (Likely yes, per
   `internal/sync/`.)

---

**End of spec.** The next agent reads §1 (drift) → §2 (geometry) →
§3 (states) → §4 (motion) → §8 (MOAT compliance). Implementation
follows §10 (notes). Tests follow §11. i18n keys land in §12.
