# Condura · DESIGNLANG

> **The design language reference for the Condura shell.** One coherent
> vocabulary, locked. Read `DIRECTION.md` first for the *voice*, then this
> file for the *tokens* that voice reads through. Read `MOAT.md` for the
> *premium bar* and `TEARDOWN.md` for the *why* behind the picks. Read
> `APPFLOW.md` for the *surfaces* the tokens render. When this doc and the
> code disagree, the code is wrong.
>
> **Sister documents.** `DIRECTION.md` is the personality · `MOAT.md` is the
> bar · `TEARDOWN.md` is the source-pattern breakdown · `APPFLOW.md` is the
> screen-by-screen spec · `DESIGNLANG.md` (this) is the token + grammar
> reference.
>
> **Source of truth.** All values here are declared in
> `app/web/frontend/src/lib/condura/condura.css`. If you find yourself
> needing a value that isn't here, the surface is wrong — extend the doc,
> the CSS, and the consuming component in the same commit.

---

## Table of Contents

1. [Typography Scale](#1-typography-scale)
2. [Color System](#2-color-system)
3. [Spacing](#3-spacing)
4. [Radii](#4-radii)
5. [Shadows](#5-shadows)
6. [Motion Grammar](#6-motion-grammar)
7. [Iconography](#7-iconography)
8. [The Mature Rules](#8-the-mature-rules)

---

## 1. Typography Scale

**Locked pairing.** Three families, no fourth.

| Role | Family | Token | Size | Weight | Line-height | Letter-spacing | Use |
|---|---|---|---:|---:|---:|---:|---|
| **display** | Instrument Serif | `--font-display` | `clamp(40px, 5vw, 72px)` (`--text-display`) | 400 | 0.95 (`--lh-display`) | −0.040em (`--ls-display`) | Hero only — Chat empty state, About colophon, Ritual Arrival wordmark. One per surface, never more. |
| **h1** | Instrument Serif | `--font-display` | 32 px (`--text-h1`) | 400 | 1.10 (`--lh-h1`) | −0.030em (`--ls-h1`) | Route titles — Settings sections, About header, Channels h1. |
| **h2** | Instrument Serif | `--font-display` | 24 px (`--text-h2`) | 400 | 1.20 (`--lh-h2`) | −0.025em (`--ls-h2`) | Sub-section titles within a route. |
| **h3** | Inter | `--font-sans` | 20 px (`--text-h3`) | 600 | 1.30 (`--lh-h3`) | −0.020em (`--ls-h3`) | Card titles, sheet headers, dialog headlines. |
| **h4** | Inter | `--font-sans` | 17 px (`--text-h4`) | 600 | 1.40 (`--lh-h4`) | −0.015em (`--ls-h4`) | Row labels, form labels, list-item titles. |
| **body** | Inter | `--font-sans` | 15 px (`--text-body`) | 400 | 1.60 (`--lh-body`) | −0.008em (`--ls-body`) | Default body, table cells, lists. The unit. |
| **caption** | JetBrains Mono | `--font-mono` | 12 px (`--text-caption`) | 500 | 1.40 (`--lh-caption`) | +0.040em (`--ls-caption`) | Eyebrows, status pills, kbd hints. |
| **mono** | JetBrains Mono | `--font-mono` | 12 px (`--text-mono`) | 500 | 1.50 (`--lh-mono`) | +0.120em (`--ls-mono`) | Code, paths, IDs, timestamps, hex references. |

**Family declarations:**

```css
--font-display: 'Instrument Serif', Georgia, 'Times New Roman', serif;
--font-sans:    'Inter', -apple-system, BlinkMacSystemFont, system-ui, sans-serif;
--font-mono:    'JetBrains Mono', ui-monospace, 'SF Mono', Menlo, monospace;
```

**Italic accent (`.alive`).** Reserved for the **single load-bearing word** on a
surface. Rendered in Instrument Serif italic, color `--synapse`. One per
surface, never two. Five uses of `.alive` is a tic; one is a wink. Surface
candidates: the word *alive* in the Chat empty hero, *here* in the Ritual
breath, *Allowed* on the consent stamp. Everywhere else, write the headline
well enough that italic-green isn't needed.

**`<kbd>` style.** Inter is a poor fit for keyboard hints. `<kbd>` elements:
JetBrains Mono, 10 px, +0.08em tracking, uppercase, paper-card background,
1 px `--hair-strong` border, `--r-xs` (6 px). Never plain text — keyboard
shortcuts must read as keyboard shortcuts, not as prose.

**Tracking direction.** Display & headings **tighten** (−em) because large
serifs visually bloom. UI & body **breathe** (−em slightly) because dense
sans-serif at small sizes needs the opposite. Captions & mono **open**
(+em) because uppercase metadata and code at small sizes collapse without
extra air.

**Anti-patterns.**

- No fourth font family. If a surface needs something not here, redesign the
  surface.
- No mixing weights within a role. h3 is always 600, body is always 400.
- No `font-size` numbers outside the scale. Reach for `--text-*` first.
- No `font-family` declarations in component CSS. Reuse `--font-*` tokens.

---

## 2. Color System

### 2.1 The semantics, not the hex

Components read **roles**, never raw hex. The token is the contract. If a
surface picks a color that isn't in the table below, the surface is wrong.

### 2.2 Light mode · the default · `:root[data-mode='light']`

| Role | Token | Light hex | Use |
|---|---|---|---|
| **paper (page)** | `--paper` | `#F4EFE4` | The whole shell. Warm cream, never pure white. |
| **paper-2** | `--paper-2` | `#ECE5D4` | Cards sitting on paper — one step warmer. |
| **paper-raised** | `--paper-raised` | `#FBF7EC` | Floating surfaces — one step lighter than paper. |
| **paper-sunken** | `--paper-3` | `#E2DAC6` | Recessed wells (inputs, sunken panels). |
| **paper-shadow** | `--paper-shadow` | `#D8CFB8` | Deepest paper tint; the dark mode paper base. |
| **surface (alias)** | `--surface` | `var(--paper)` | The page. |
| **surface-card** | `--surface-card` | `var(--paper-2)` | A card. |
| **surface-raised** | `--surface-raised` | `var(--paper-raised)` | Popover, tooltip, modal. |
| **surface-sunken** | `--surface-sunken` | `var(--paper-3)` | A recessed well. |
| **surface-ink** | `--surface-ink` | `#16140F` | Rare: dark pill on a light page (the inverse). |
| **content (ink)** | `--content` | `var(--ink)` = `#14110B` | Primary content — warm near-black, never `#000000`. |
| **content-soft** | `--content-soft` | `var(--ink-soft)` = `#2A2519` | Secondary content (bodies, descriptions). |
| **content-mute** | `--content-mute` | `var(--ink-mute)` = `#5C5443` | Tertiary (timestamps, hints, metadata). |
| **content-faint** | `--content-faint` | `var(--ink-faint)` = `#8A8169` | Quaternary (eyebrows at rest). |
| **content-ghost** | `--ink-ghost` | `#B8AF98` | Decorative only — never for content the user must read. |
| **hair** | `--hair` | `rgba(20, 17, 11, 0.10)` | The only sanctioned 1 px divider. |
| **hair-strong** | `--hair-strong` | `rgba(20, 17, 11, 0.18)` | The only sanctioned emphasized divider. |
| **accent (synapse)** | `--synapse` / `--accent` | `#0B3D2E` | Brand link, "alive" accent. Deep forest green. |
| **accent-deep** | `--synapse-deep` | `#06241B` | Pressed / inset synapse. |
| **accent-glow** | `--synapse-glow` / `--accent-glow` | `#1A8A6A` | Focus highlight (brighter than synapse). |
| **accent-light** | `--synapse-light` | `#9CE8C8` | Tinted surfaces only — never for content. |
| **action (pollen)** | `--pollen` / `--action` | `#C97B2E` | Primary CTA color, mote color, cursor-tip. Warm amber. |
| **action-deep** | `--pollen-deep` | `#9A5A1A` | Pressed pollen. |
| **action-light** | `--pollen-light` | `#F0C082` | Tinted wash backgrounds only. |
| **focus** | `--focus` | `var(--synapse-glow)` | The color at the center of every focus halo. |
| **danger** | `--danger` | `#A3312A` | Rust red — not stop-sign red. |
| **success (ok)** | `--ok` | `#2E7D4F` | Forest green, distinct from `--synapse` so status ≠ brand. |
| **warn** | `--warn` | `#B07A2E` | Amber, distinct from `--pollen` so status ≠ action. |
| **info** | `--info` | `#3A5A8C` | Muted indigo — rare; About / Audit meta only. |

### 2.3 Dark mode · `:root[data-mode='dark']`

| Role | Token | Dark hex |
|---|---|---|
| **paper** | `--paper` | `#16140F` |
| **paper-2** | `--paper-2` | `#1F1C16` |
| **paper-raised** | `--paper-raised` | `#221E17` |
| **paper-sunken** | `--paper-3` | `#2A2519` |
| **paper-shadow** | `--paper-shadow` | `#100E0A` |
| **surface-ink** | `--surface-ink` | `#0B0A07` |
| **content** | `--content` | `#F0EDE5` |
| **content-soft** | `--content-soft` | `#D8D2C4` |
| **content-mute** | `--content-mute` | `#8A847A` |
| **content-faint** | `--content-faint` | `#5C564C` |
| **content-ghost** | `--ink-ghost` | `#3A352D` |
| **hair** | `--hair` | `rgba(240, 237, 229, 0.16)` |
| **hair-strong** | `--hair-strong` | `rgba(240, 237, 229, 0.28)` |
| **accent (synapse)** | `--synapse` | `#1A8A6A` (brighter so contrast holds against `#16140F`) |
| **accent-deep** | `--synapse-deep` | `#0B3D2E` |
| **accent-glow** | `--synapse-glow` | `#2FBF95` |
| **accent-light** | `--synapse-light` | `#0B3D2E` |
| **action (pollen)** | `--pollen` | `#E0A862` |
| **action-deep** | `--pollen-deep` | `#C97B2E` |
| **action-light** | `--pollen-light` | `#9A5A1A` |
| **danger** | `--danger` | `#C75449` |
| **success (ok)** | `--ok` | `#4FAE74` |
| **warn** | `--warn` | `#D49A4E` |
| **info** | `--info` | `#6B8FB8` |

### 2.4 Shadows (the warm ones)

| Token | Light composition | Dark composition |
|---|---|---|
| `--shadow-paper` | `0 1px 0 rgba(20,17,11,.04), 0 8px 24px -12px rgba(20,17,11,.12)` | `0 1px 0 rgba(240,237,229,.04), 0 8px 24px -12px rgba(0,0,0,.55)` |
| `--shadow-card` | `0 1px 0 rgba(20,17,11,.06), 0 24px 60px -28px rgba(20,17,11,.22), 0 2px 8px -4px rgba(20,17,11,.08)` | `0 1px 0 rgba(240,237,229,.05), 0 24px 60px -28px rgba(0,0,0,.70), 0 2px 8px -4px rgba(0,0,0,.45)` |
| `--shadow-float` | `0 2px 0 rgba(20,17,11,.05), 0 40px 90px -32px rgba(20,17,11,.30), 0 4px 12px -6px rgba(20,17,11,.10)` | `0 2px 0 rgba(240,237,229,.05), 0 40px 90px -32px rgba(0,0,0,.75), 0 4px 12px -6px rgba(0,0,0,.50)` |
| `--pollen-halo-color` | `rgba(201,123,46,.18)` | `rgba(224,168,98,.16)` |
| `--shadow-focus` | `0 0 0 4px var(--pollen-halo-color), inset 0 0 0 1px color-mix(in oklab, var(--focus) 32%, transparent)` | same composition, dark colors |

Warm, ink-tinted, never neutral gray. The shadow tokens are the elevation
language — see §5 for the grammar.

### 2.5 Blooms (used by `.surface-paper` only)

```css
--bloom-1: rgba(180, 216, 232, 0.18);   /* cool wash, top-left */
--bloom-2: rgba(232, 216, 180, 0.22);   /* warm wash, top-right */
--bloom-3: rgba(11,  61,  46,  0.05);   /* synapse echo, bottom */
```

Dark mode values shift intensity; same composition. Blooms live behind a
sibling `.paper-grain` overlay. **Never** use the bloom colors as a fill on
an interactive surface — they're ambient decoration, not content.

### 2.6 The rule for status vs. brand

**Brand colors (`--synapse`, `--pollen`) and status colors (`--ok`, `--warn`,
`--danger`, `--info`) are never reused across categories.** `--synapse` is
the link color; `--ok` is the success badge color. They look related by
design — both warm, both green-leaning — but they are not interchangeable.
A green border on a card reads "link"; a green border on an autonomy-matrix
dot reads "autonomous." The user's eye needs to read both without ambiguity.

### 2.7 No new colors without an amendment

If a surface needs a hue that isn't in the four brand colors (synapse,
synapse-deep, synapse-glow, synapse-light + the same for pollen), the four
status colors, or the ink/paper scales, **the surface is wrong.** Adding a
new color requires a CLAUDE.md amendment, per MOAT.md §4 rule 4. The 2026
rule: a constrained palette reads as designed; an open palette reads as
improvised.

---

## 3. Spacing

```css
--space-1:   4px;
--space-2:   8px;
--space-3:  12px;
--space-4:  16px;
--space-5:  20px;
--space-6:  24px;
--space-7:  32px;
--space-8:  40px;
--space-9:  48px;
--space-10: 64px;
--space-11: 80px;
```

**The scale is one.** Eleven steps. Components do not declare
`margin: 13px` or `padding: 18px`. They reach for the closest step that
serves the gestalt. `4, 8, 12, 16, 20, 24, 32, 40, 48, 64, 80` — every
number that lands on a surface lives here.

### 3.1 Collapse decisions

The scale collapses four standard Tailwind-style steps (`14, 22, 28`) on
purpose. Where the team might naturally write `padding: 14px`, **collapse
to `12px` (`--space-3`) or `16px` (`--space-4`)** — the difference between
12 and 14 reads as a typo, not as choice. Where the team might reach for
`20–28`, the scale offers `--space-5` (20) and `--space-7` (32). The space
between (`24, 28`) is intentionally absent to keep rhythm intent: if a
surface needs more than `24` it deserves `32`.

| Intent | Use | Not |
|---|---|---|
| Inline icon to text | `--space-2` (8) | `6, 10` |
| Card padding | `--space-6` (24) | `--space-7` (only the inside of an inset control) |
| Section gap (within a route) | `--space-7` (32) | `--space-6` (too tight) |
| Route-to-route gap | `--space-9` (48) | `--space-8` (40) |
| Hero padding (Chat empty / About colophon) | `--space-10`/`--space-11` | custom `96–128px` |

### 3.2 Density rules

| Surface | Token | Notes |
|---|---|---|
| Card inner padding | `--space-5` (20) | `--space-6` (24) for generous, `--space-4` (16) for tight. |
| Card-to-card gap | `--space-4` (16) | Never `--space-3`. |
| Form label to input | `--space-2` (8) | The exception — tighter because label is small. |
| List row vertical padding | `--space-3` (12) | `--space-4` (16) for touch surfaces. |
| Section heading to body | `--space-5` (20) | |
| Page top margin (first surface) | `--space-7` (32) | Resets to `--space-9` on the empty hero only. |

### 3.3 What spacing is not

- **Not a motion ease.** If you want softness, use `--ease`, not larger padding.
- **Not a layout grid.** The shell uses flex and grid directly. Spacing is the
  between, not the structure.
- **Not a stacking context manager.** Use `z-*` tokens for that.

---

## 4. Radii

```css
--r-xs:      6px;     /* chips, eyebrow tags, kbd */
--r-sm:     10px;     /* secondary buttons */
--r-control: 10px;     /* inputs, buttons — same as sm by design */
--r-md:     16px;     /* cards */
--r-lg:     24px;     /* panels, side sheets */
--r-xl:     36px;     /* modals, hero surfaces */
--r-pill:   999px;    /* toggles, status pills */
```

### 4.1 The role-to-radius map (one role, one radius — never stack)

| Element | Token | Why |
|---|---|---|
| Chip / eyebrow / tag / `<kbd>` | `--r-xs` (6) | Small, dense, decorative. |
| Secondary button (e.g., "Try it" preset) | `--r-sm` (10) | Rounded, not pill — readable as a button. |
| Primary button / input / dropdown / toggle surface | `--r-control` (10) | Canonical control radius. Alias of `--r-sm`. |
| Card (paper-card, skill card, channel row) | `--r-md` (16) | The room. |
| Side sheet, panel (Settings sub-section, sync pairing card) | `--r-lg` (24) | |
| Modal (consent, kill switch) | `--r-xl` (36) | |
| Toggle (pill), status pill, autonomy-matrix dot | `--r-pill` (999) | Renders as fully rounded regardless of size. |

**`--r-control` is the canonical name.** Components must read `--r-control`
for inputs and buttons. The duplicate `--r-sm` exists so a surface can say
"small" semantically (a chip vs. a control) without inventing a third value.
Never both. Never an average. **The radius is the role.**

### 4.2 Container clipping

Any container with non-zero radius must set `overflow: hidden` if it
contains a child that paints to the bleed (a paper-grain overlay, a
subtle bloom, a child card-image). A card with `border-radius: 16px`
that hosts an oversized image without clipping is a layout bug.

### 4.3 Anti-patterns

- No percentage-based `border-radius` on rectangular controls. `50%` belongs
  to circles only.
- No `border-radius: 0` on cards. A card with `0` reads as a wireframe, not a
  product. The minimum card radius is `--r-md` (16).
- No mixing `--r-sm` and `--r-control` on adjacent interactive surfaces.
  The visual seam is a lie.

---

## 5. Shadows

```css
--shadow-paper:  0 1px 0 rgba(...), 0 8px 24px -12px rgba(...);    /* resting hairline + soft outer */
--shadow-card:   0 1px 0 rgba(...), 0 24px 60px -28px rgba(...), 0 2px 8px -4px rgba(...); /* resting + lift + crisp */
--shadow-float:  0 2px 0 rgba(...), 0 40px 90px -32px rgba(...), 0 4px 12px -6px rgba(...); /* pressed hairline + float + crisp */
--shadow-focus:  0 0 0 4px var(--pollen-halo-color), inset 0 0 0 1px color-mix(...);    /* the universal hover */
```

Four tokens. Each one owns one elevation tier. Colors are warm, ink-tinted
in both light and dark mode (never neutral gray). Components never stack
two shadows, never average two shadow tokens, never declare a custom
`box-shadow` outside the four. **If a surface needs more elevation than
`--shadow-float`, the answer is `--shadow-float`.**

### 5.1 The elevation tier map

| Tier | Token | Surface examples | When |
|---|---|---|---|
| **Paper** | `--shadow-paper` | Inline notice, hover-state of an embedded surface | resting, almost-on-the-page |
| **Card** | `--shadow-card` | Cards on the main surface, hover lift of a card hover | the default elevation |
| **Float** | `--shadow-float` | Floating picker (command palette), side sheet, modal scrim edge | above the page, hovering |
| **Focus** | `--shadow-focus` | Universally — applied via `:focus-visible` | keyboard focus only |

### 5.2 Transitions between tiers

Hover from rest → lift is `--shadow-paper → --shadow-card`, animated by the
universal transition list (see §6.2). Hover from a card → pop is
`--shadow-card → --shadow-float`. Going *back* to rest returns through the
same chain. **Never** stack `--shadow-card + --shadow-float` simultaneously —
that's `rule 4` of the Mature Rules (see §8).

### 5.3 The ink-tint warmth

Shadows tint `--ink` (light mode) or `--content` (dark mode). The combined
`rgba` in the shadow composition is what carries the brand voice down into
the depth. Neutral-gray shadows read as Material Design defaults; warm
shadows read as paper casting a shadow on paper. The opacity depths:

| Tier | Y-off (px) | Blur (px) | Color | Opacity (light) |
|---|---|---|---|---:|
| paper | 8 | 24 | `--ink` | 0.12 |
| card | 24 (with 60 blur) + 2 (with 8 blur) | -- | `--ink` | 0.22 outer, 0.08 inner |
| float | 40 (with 90 blur) + 4 (with 12 blur) | -- | `--ink` | 0.30 outer |

The deeper the tier, the higher the outer-blur opacity. Dark mode follows
the same shape but substitutes `rgba(0,0,0)` for the outer shadow and a
small `rgba(--content, .04–.05)` for the resting hairline.

### 5.4 What shadows are not

- Not a press state. Press is `scale(0.97)` + `filter: brightness(0.95)
  saturate(1.1)`, not a darker shadow.
- Not a hover affordance on rectangular CTAs (pollen halo does that — see §6.3).
- Not a divider. A 1 px hairline with `--hair` is the divider.
- Not four stacked shadowns on a hero card. One is a card; four is a
  texture demo.

---

## 6. Motion Grammar

### 6.1 Eases & durations (locked — the entire vocabulary)

```css
--ease:     cubic-bezier(0.22, 1, 0.36, 1);   /* out — the default */
--ease-in:  cubic-bezier(0.65, 0, 0.35, 1);   /* in  — entrances */
--ease-pop: cubic-bezier(0.34, 1.56, 0.64, 1); /* press, single-element scale */

--dur-fast: 140ms;   /* state changes — hover, press, focus halo, color shifts */
--dur:      280ms;   /* the everyday transition */
--dur-slow: 520ms;   /* route enters, panel slides, threads draw, signature gestures */
--dur-cine: 900ms;   /* Ritual blooms, modal reveals, the wordmark arrival */
```

The four durations are the *entire* vocabulary. Components do not declare
`transition: ... 320ms`. They use `--dur`. If a surface needs a duration
that isn't in the four, the surface is wrong — find a way to express the
gesture in 140 / 280 / 520 / 900 ms.

### 6.2 The transition list (declared in `.tactile`, never redeclared)

```css
.tactile,
button:not(.no-tactile),
[role='button']:not(.no-tactile),
summary:not(.no-tactile) {
  transition:
    transform         var(--dur-fast) var(--ease),
    background-color  var(--dur) var(--ease),
    border-color      var(--dur) var(--ease),
    color             var(--dur) var(--ease),
    box-shadow        var(--dur) var(--ease);
}

.tactile:active:not([disabled]):not([aria-disabled='true']),
button:not(.no-tactile):active:not([disabled]):not([aria-disabled='true']),
[role='button']:not(.no-tactile):active:not([disabled]):not([aria-disabled='true']) {
  transform: scale(0.97);
}
```

Components that want tactile behavior use `class="tactile"`. The CSS owns
the transition. Components are free to add *meaning* — a tinted shadow on a
primary press, a thread stroke that brightens on a card lift — but they do
not duplicate the timing. Press is a `scale(0.97)` + the brightness shift
below, animated at `--dur-fast` so it lands before the user's eye leaves the
press.

### 6.3 The four motion primitives

These are the only gestures that earn the privilege of repeating across
surfaces. Everything else is composition of these four.

#### Primitive 1 · **Tactile Press**

```css
/* on .tactile + role='button' */
:active { transform: scale(0.97); }
/* on :active, applied alongside scale: */
filter: brightness(0.95) saturate(1.1);
transform: translateY(0.5px);
```

- **When:** any clickable surface is pressed.
- **Duration:** `--dur-fast` (140 ms).
- **Easing:** `--ease`.
- **Why:** a smaller element reads as "smaller," not "pressed." The
  brightness drop + the 0.5 px settle *are* the press. Premium touch
  screens do this for free; the web has to choose to.

#### Primitive 2 · **Focus Halo**

The universal focus treatment follows the geometry the user is touching.
**No element in the codebase may use `outline: 1px solid var(--content)` or
any other rectangular outline.** Focus is the user's pointer; it must
follow the shape they reach for.

```css
/* Universal (default for any focusable surface) */
:focus-visible {
  outline: none;
  box-shadow: var(--shadow-focus);   /* 4 px pollen halo + 1 px synapse inset */
}

/* Pill radius (>= 999px) — drop the inset, carry the focus on the ring */
.pill:focus-visible {
  box-shadow:
    0 0 0 2px var(--synapse),
    0 0 0 5px var(--pollen-halo);
}

/* Rectangular inputs (the EULA well, the keycap surface) — halo only, no inset */
.rect:focus-visible {
  box-shadow: 0 0 0 4px var(--pollen-halo);
}
```

- **When:** `:focus-visible` (keyboard only, never mouse).
- **Duration:** `--dur-fast` (140 ms).
- **Easing:** `--ease`.
- **Why:** a focus ring is a promise to the keyboard user that the
  location is current. The promise reads as a halo, not as a rectangle.

#### Primitive 3 · **Thread Draw**

**The signature animation.** A 1.25 px synapse hairline that *draws* in
from left to right, signaling that a connection was made — a message
landed, a permission was granted, an action completed.

```css
/* SVG path variant */
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

/* CSS-only variant — the canonical class */
.thread {
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--synapse) 20%, var(--synapse) 80%, transparent);
  transform: scaleX(0);
  transform-origin: left;
  transition: transform var(--dur-slow) var(--ease);
}
.thread.draw { transform: scaleX(1); }
```

- **When:** a moment needs weight — completion, success, attention, error
  resolution. **Where the Thread MUST appear.** Chat empty (under hero
  copy), Chat mid-conversation (between every pair of turns), Settings
  sections (under each h1), Audit log (vertical spine + per-node hairline
  on click), Replay (scrubber fill left of playhead), Sync (inter-device
  threads on pair complete), Hub (implied under "Installed ✓"), Channels
  (under the audit-threadlink footer), Constellation (bezier paths
  connecting wired nodes to center), Consent modal (the armor rect).
- **Duration:** `--dur-slow` (520 ms).
- **Easing:** `--ease`.
- **Where the Thread MUST NOT appear.** As a decorative flourish on a
  button hover. As a divider between unrelated surfaces (use `--hair`).
  On the titlebar wordmark (the wordmark dot is its own gesture).

#### Primitive 4 · **Synapse Pulse**

The idle heartbeat. The agent breathes — it is alive without performing it.

```css
@keyframes breathe {
  0%, 100% { transform: scale(1);    opacity: 0.85; }
  50%      { transform: scale(1.18); opacity: 1; }
}
/* applied to the Pulse component, the Dot constellation, the garden motes */

.Pulse {
  animation: breathe 4s var(--ease) infinite;
}
```

- **When:** only for nodes in `awaiting` / `listening` / `idle` phase.
  Never on a fresh action; never inside a streaming chip.
- **Duration:** 4 s loop.
- **Easing:** `--ease`.
- **Why:** the breathing pulse is the visible heartbeat of an agent that
  is alive but resting. A Pulsing element next to a streaming chip reads
  as competing for attention; a Pulsing element next to silence reads as
  presence.

### 6.4 The reduced-motion contract

Owned **exclusively** by `condura.css`. A single block:

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

Components never re-declare their own reduced-motion media queries.
`IntersectionObserver` and `Page Visibility` listeners short-circuit on
`matchMedia` once at mount, never per-component. **Battery is the user,
not the design** (MOAT.md §2.3).

### 6.5 Energy budget

`data-energy="low"` (set by the platform via `internal/perception` energy
mode) and `@media (prefers-reduced-motion)` both disable ambient pollen
motes, paper grain, and the slow pulse — and snap `--dur-cine` /
`--dur-slow` to `0ms`. The product is still usable; it just stops
performing.

### 6.6 Gestures that aren't the four

These are derived motions — components compose the four primitives into
specific gestures. They aren't in the primitives list because they don't
recur by themselves; they recur as *that surface's* gesture.

| Gesture | Composition | Used by |
|---|---|---|
| **Route enter** | `opacity 0→1` + `filter: blur(8px)→0` + `translateY(12px)→0`, `--dur-slow`, `--ease` | Every route (Shell wraps the surface) |
| **Panel slide-in (right)** | `translateX(24px)→0` + `opacity 0→1`, `--dur-slow`, `--ease` | Settings sub-panels, Hub / Skills detail sheet, Pairing |
| **Panel slide-out (right)** | reverse, `--dur`, `--ease-in` | Same panels on close / outside-click |
| **Save bar rise** | `translateY(40px)→0` + `opacity 0→1`, `--dur-slow`, `--ease-pop` | Settings sticky save bar (`dirty = true`) |
| **Palette highlight slide** | `top: highlightTop`, `transform: translateY(...)`, `--dur`, `--ease` | Command palette active row |
| **Pollen float (ambient)** | `translate(dx,dy) rotate(0→360)` + `opacity 0→0.9→0`, 7–13 s loop, `linear` | Empty-state decoration only — Chat empty garden, Constellation motes |
| **Word rise (per-word mask rise)** | `translateY(110%) → 0`, `--dur-cine`, `--ease` | Ritual arrival text, error headlines |
| **Stop (streaming)** | "Send" → "Stopping…" 280 ms → "Stopped" + 1× pollen pulse, `--dur` + `--dur-slow`, `--ease` | Chat send/stop toggle |
| **Seal bloom** | `transform: scale(0.94)` on press, radial sealBloom ring 28 px → 0 px, 600 ms, `--ease` | EULA stamp only (per DIRECTION.md — never reuse) |
| **Stamp** | `scale(0)→1.18→0.96→1` + `rotate(-18deg)→6deg→-2deg→0deg`, `--dur-cine` | Wax seal on the consent modal Allow |

### 6.7 Anti-patterns

- No custom cubic-beziers. The three eases are the vocabulary.
- No motion that doesn't carry meaning. A spinned loader carries nothing;
  the Thread drawing in carries "this is now finished." If a gesture is
  decorative with no meaning, delete it.
- No `transition: all` declarations. They hide what is changing and tend
  to animate properties that shouldn't move (e.g., `width`).
- No animations triggered by mount without an exit counterpart — a 280 ms
  fade-in on a 100 % opacity target wastes 280 ms of the user's time.

---

## 7. Iconography

### 7.1 The contract (single-source, locked)

Declared in `app/web/frontend/src/lib/condura/icons.ts`. Rendered via
`<Glyph name="..." stroke={n} />` (default `stroke={1.5}`).

| Property | Value |
|---|---|
| **Grid** | 24u (viewBox `0 0 24 24`) |
| **Stroke** | 1.5 px uniform, currentColor (overridable per `<Glyph stroke={n}>`) |
| **Caps / joins** | round (uniform) |
| **Metaphors** | one per icon — no doubles, no flourishes, no "and also…" |
| **Path shape** | one path (or a small set of paths) per icon. No filled shapes unless the metaphor demands it (dot, dot-active, menu kebab, stop, kill-switch). |
| **Naming** | kebab-case (`chevron-right`, `theme-sun`, `kill-switch`) |

### 7.2 Categories

| Category | Purpose | Examples |
|---|---|---|
| **nav** | Sidebar / nav rail / command palette routes | `chat`, `skills`, `hub`, `channels`, `audit`, `about`, `settings`, `sync`, `delegation`, `account`, `kill-switch`, `replay` |
| **action** | Generic UI controls (send, close, check, etc.) | `send`, `close`, `back`, `check`, `plus`, `search`, `command`, `menu`, `trash`, `shield`, `bolt`, `spark`, `chevron-right`, `chevron-down`, `chevron-left`, `info`, `warning`, `stop`, `book`, `circle` |
| **state** | Status indicators (the dot family) | `dot`, `dot-active` |
| **theme** | Theme picker set (sun / auto / moon) | `theme-sun`, `theme-auto`, `theme-moon` |
| **media** | Input/output affordances | `mic`, `key`, `power` |

### 7.3 The required glyph inventory

#### nav (12)

`chat` · `skills` · `hub` · `channels` · `audit` · `about` · `settings` · `sync` · `delegation` · `account` · `kill-switch` · `replay`

#### action (19)

`send` · `close` · `back` · `check` · `plus` · `search` · `command` · `menu` · `trash` · `shield` · `bolt` · `spark` · `chevron-right` · `chevron-down` · `chevron-left` · `info` · `warning` · `stop` · `book` · `circle`

#### state (2)

`dot` · `dot-active`

#### theme (3)

`theme-sun` · `theme-auto` · `theme-moon`

#### media (3)

`mic` · `key` · `power`

**Total: 39 canonical icons + 7 aliases** (defined in `ALIASES` in
`icons.ts`: `chevronRight`, `chevronDown`, `chevronLeft`, `killSwitch`,
`themeSun`, `themeAuto`, `themeMoon`, `gear`, `sparkle`, `sun`, `auto`,
`moon`, `x`).

### 7.4 How to size

| Context | Size | Stroke |
|---|---:|---:|
| NavRail glyph (default) | 20 px | 1.5 |
| Titlebar glyph (kill switch, theme) | 18 px | 1.5 |
| Button glyph | 16 px | inherits (1.5) |
| Inline meta glyph (eyebrow tag, kbd) | 12 px | 1.25 |
| Large display glyph (illustration-of-an-icon, empty state) | 48–64 px | 1.25 |
| Audit spine, channel signal bars | n/a (geometry) | inherit |

### 7.5 How to color

`<Glyph>` draws in `currentColor`. The icon takes the parent's text color.
Set color via class or inline `style="color: var(--accent)"` — never hard-hex.

- Default: `--content`.
- Hover state: `--content-soft` → `--content` (transition `--dur`).
- Active/pressed: `--content`.
- Disabled: `--content-faint`.
- Brand intent (kill switch glyph when armed, theme toggle on hover): `--accent`.

### 7.6 Anti-patterns

- **No emoji as UI icons.** Use `<Glyph name=...>`. The only emoji allowed
  in the codebase are in user-facing copy *about* emoji (e.g., a tooltip
  explaining how to type an emoji into voice transcription). They are never
  the icon for a button, a nav item, a status badge, or a state indicator.
- **No icon fonts.** A glyph is a stroke, not a font character.
- **No `fontawesome` / `lucide` / `material-icons`.** The icon set is
  internal — every visible icon ships from `icons.ts`. Three new icons per
  quarter is healthy; thirty new emoji in a route is a smell.
- **No rotated icons as state.** State is communicated via color, dot size,
  and `--dot-active` — never via 45° / 90° / 180° rotations of unrelated
  icons.
- **No mixed metaphors.** A `trash` with a `shield` overlay is two icons
  arguing. Pick the one that conveys the meaning and remove the other.
- **No 16-line SVG illustrations in place of icons.** An icon is one path;
  a hero illustration is a separate surface with its own canvas.

---

## 8. The Mature Rules

These are the ten rules that guarantee Condura does not look like a weekend
hack. Any surface that breaks one is wrong; any PR that lands one is
rejected. No exceptions, no "but it's just this once." They are the
operational form of MOAT.md §4 ("WHAT WE WILL NOT DO") and DIRECTION.md §6
("THE ANTI-VIBE-CODED RULES"), made enforceable through this token system.

### 1 · **No gradient text.**

The brand voice is paper and ink. Gradient text is a 2017-portfolio-site
tic that screams "I didn't have a brand so I made one mid-build." Text is
one color per role. Headlines are `--content` (or `--synapse` when the word
is the `.alive` accent). Sub-copy is `--content-soft`. Eyebrows are
`--content-faint`. **Never a gradient between two of them.** A background
can carry a gradient (the `--bloom-*` radials on `.surface-paper`); text
never does.

### 2 · **No emoji as UI.**

Use `<Glyph name=...>` from `icons.ts`. The Glyph set is single-stroke,
1.5 px weight, currentColor, drawn in the brand line vocabulary. If a
surface needs an icon that isn't there, add it to `icons.ts` — three new
icons per quarter is healthy; thirty new emoji in a single route is a
smell. The only emoji allowed in the codebase are in user-facing copy
*about* emoji (e.g., a tooltip explaining how to type an emoji into voice
transcription). They are never the icon for a button, a nav item, a
status badge, or a state indicator.

### 3 · **No glassmorphism unless earned.**

A floating surface that needs to look elevated (CommandPalette, Settings
sheet, Consent modal, Kill-switch overlay) uses `--shadow-float`. Period.
`backdrop-filter: blur` is **prohibited** on cards, lists, and inline
surfaces — `Channels.svelte`'s `.overlay` `backdrop-filter: blur(4px)` is
the kind of thing we delete, not add. A scrim under a modal is the only
allowed use of `backdrop-filter`, and only when it earns the elevation.

### 4 · **No double shadows.**

Components stack `--shadow-card` and a one-off
`box-shadow: 0 8px 20px …`. **This is the deepest failure.** One elevation
token per surface; that's it. The four tokens (`--shadow-paper`,
`--shadow-card`, `--shadow-float`, `--shadow-focus`) are the entire
vocabulary. Hover-lifts use `--shadow-card → --shadow-float` via the
existing transition; they never add a third shadow on top. Press states
are NOT shadows (see rule 5); press is `filter: brightness(0.95)
saturate(1.1)` + `translateY(0.5px)`.

### 5 · **One metaphor per component.**

A component picks one metaphor and commits to it. The signal-bar
component in `Channels.svelte` uses cellular-bar dots (5 stepped heights)
and only cellular-bar dots. The constellation in `Ritual.svelte` uses
nodes and bezier-thread paths and only nodes and bezier-thread paths.
The skill cards in `Skills.svelte` use card lift + thread-stroke and only
those two. A card with a gradient header and a 3D tilt and a ribbon and
a chip and a thread is four metaphors arguing. Pick the one that the
user reads in 200 ms and remove the other three.

### 6 · **Focus rings track rounded shape via `box-shadow`.**

Pill elements (`border-radius >= 999px`) get a 2 px synapse ring + 5 px
pollen halo. Rounded surfaces (8–16 px radius) get `--shadow-focus`
(4 px pollen halo + 1 px synapse inset). Rectangular inputs (the EULA
scrollable well, the keycap on hotkey capture) get the 4 px halo with no
inset line. **No element in the codebase may use `outline: 1px solid
var(--content)` or any other rectangular outline.** Focus is the user's
pointer; it must follow the geometry they are touching, not flatten it
into a rectangle.

### 7 · **Loading states teach, not spin.**

A loading state is a verb; it tells the user what is happening. The
Thread drawing in says **"data is moving."** A hash-route pulse says
**"this is a quiet surface."** A spinning circle says **"I don't know
what's happening, so I'm just spinning."** No `<Spinner />` component
exists in the codebase. No three-dot variant. No `↻` glyph in a loading
slot. The `Pulse` component is the only allowed loading indicator, and
even it must be paired with a mono-uppercase label that names the
operation: `INDEXING…`, `PROBING REACH…`, `READING THE CHAIN…`. The Pulse
breathes; the label teaches.

### 8 · **Transitions respect `prefers-reduced-motion`.**

Owned exclusively by `condura.css`. A single `@media
(prefers-reduced-motion: reduce)` block reads `*, *::before, *::after
{ animation-duration: 0.01ms !important; transition-duration: 0.01ms
!important; }` and hides the paper grain + motes + ambient threads.
Components never re-declare their own reduced-motion media queries.
`IntersectionObserver` and `Page Visibility` listeners short-circuit on
`matchMedia` once at mount, never per-component. **Battery is the user,
not the design.**

### 9 · **Color is semantic.**

Components read roles, never raw hex. The token is the contract:

- Use `--content`, `--content-soft`, `--content-mute`, `--content-faint`
  for content hierarchy — never reach for `--ink-*` directly (those
  aliases exist only for legacy v1 components still on disk).
- Use `--surface`, `--surface-card`, `--surface-raised`, `--surface-sunken`
  for surface hierarchy — never reach for `--paper-*` directly.
- Use `--hair`, `--hair-strong` for dividers — never `rgba(...)` inline.
- Use `--accent`, `--action`, `--danger`, `--warn`, `--ok`, `--info` for
  intent — never reach for color literals.
- Use `--shadow-paper`, `--shadow-card`, `--shadow-float`, `--shadow-focus`
  for elevation — never a custom `box-shadow` outside the four tokens.
- Use `--space-1` through `--space-11` for spacing — never `padding: 18px`.
- Use `--r-xs`, `--r-sm`/`--r-control`, `--r-md`, `--r-lg`, `--r-xl`,
  `--r-pill` for radii — never a custom `border-radius`.

When a surface needs to express something the tokens don't cover, that is a
signal to extend `condura.css` and this doc — in the same commit.

### 10 · **Typography leads; color follows.**

The first thing the eye reads on a Condura surface is the type — the
serif headlines, the mono eyebrows, the colon-aligned body. Type
establishes the hierarchy; color reinforces it; elevation catches the eye
on hover. The reverse order fails every time: a button that catches the
eye through color before the user has read the label is noise; a card
that lifts on hover before the user has parsed the headline is motion
that doesn't teach.

The test for every new surface:

> *Does the typography lead? Does the color reinforce? Does the elevation
> reward? Does the motion carry meaning?*

If all four answer yes, ship it. If any answer is no, the surface is
incomplete.

---

## Closing note

This document is the grammar. The five sister docs are the grammar *applied:*

- `DIRECTION.md` — the voice · what Condura sounds, looks, and feels like
- `MOAT.md` — the bar · the five tests a premium product must pass
- `TEARDOWN.md` — the why · what we stole, what we skipped, what we made our own
- `APPFLOW.md` — the spec · every screen the user can land on
- `DESIGNLANG.md` (this) — the tokens, the grammar, the rules

When five disagree, the disagreement is the bug. Fix the code or fix the
doc — but do it in the same commit, and never by silently rewriting one to
match the other.

**The single test for every new surface:**

> *Does this read like a paper notebook that learned to listen — warm,
> awake, and never louder than the room it's in?*

If yes, ship it. If no, it isn't finished yet.
