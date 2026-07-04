# SCREEN_ABOUT — Condura Colophon

> The `About` route. `app/web/frontend/src/lib/condura/About.svelte` mounted at `#/about`.
> The quietest surface in the product. A flowing document, not a modal, not a settings tab.
> This is where the seven non-negotiable invariants from `CLAUDE.md §2.1` are surfaced
> as the visible promise of the agent.

| Field | Value |
|---|---|
| Route | `#/about` |
| Component | `About.svelte` |
| NavRail label | "About" |
| Visible in NavRail | yes (item 10 of 10) |
| Max-width | `760px` (code; spec target was `~720px` — leave at 760, do not regress) |
| Pattern | flowing document |
| Container | `<article class="about">` |

---

## 1. LAYOUT & CONTENT

A single-column flowing document. Three vertical blocks: **Hero**, **Ledger**,
**Footer colophon**. Nothing else. No cards, no grid, no tabs.

### 1.1 Block anatomy

| Block | Component / element | Token role | Notes |
|---|---|---|---|
| **Hero** | `<header class="head">` | — | Always renders. |
| Hero eyebrow | `<div class="eyebrow">` | `font-mono · 11px · +0.22em · uppercase · content-faint` | `— The colophon` |
| Hero title | `<h1 class="title">` | `font-display · clamp(32,4vw,48) · -0.035em · content` | "Made by a human and an AI, in partnership." Italic lives on every invariant row, NOT here. (MOAT §1.7: `.alive` reserved once per surface; here it's the "alive" word *here* on First Breath instead.) |
| Hero sub | `<p class="sub">` | `15px · 1.6 · content-soft · max-width: 56ch` | Three sentences: what Condura is · what version · why the seven promises are non-negotiable. |
| Hero cred | `<div class="cred">` | `font-mono · 11px · +0.1em · uppercase · content-faint` | `<Pulse phase="thinking" size={6}/> + "thinking in public · <a>v0.1.0 · changelog</a>"` |
| **Ledger** | `<section class="ledger">` | top + bottom `1px hair` | One-thread container. |
| Ledger eyebrow | `<div class="ledger-eyebrow">` | same token as hero eyebrow | `— The seven invariants` |
| Invariant row (×7) | `<div class="row" data-line="0X" role="group">` | grid `56px / 1fr` | 01–07, numbered. |
| **Footer** | `<footer class="foot">` | text-align: center | Colophon + breath-thread. |
| Breath thread | `<Thread orientation="h" draw={colophonIn}/>` + `<Pulse phase="idle" size={6}/>` | 24px-tall row | Thread draws in on `requestAnimationFrame` after mount. |
| Colophon line | `<p class="colophon">` | `font-display · italic · 15px · 1.8 · content-soft` | Version · EULA · Privacy · Support Condura (donate). |

### 1.2 Invariant row — exact structure

| Sub-element | Style | Notes |
|---|---|---|
| `.row-n` | `font-display · 28px · 1.0 · synapse · opacity: 0.55` | On hover → `opacity: 1`. The number is the only piece of the row that brightens on hover. |
| `.row-title` | `font-display · italic · 18px · 1.25 · content` | The promise, one sentence. |
| `.row-text` | `14px · 1.55 · content-soft` | Plain-language explanation, 1–2 sentences. |
| `.hairline` | SVG `<line>` at `bottom: 0`, `height: 1px`, full row width | Draws in L→R via `stroke-dashoffset: 1 → 0` when row enters viewport. |
| `.armor` | SVG `<rect rx=14>` inset `1px`, fills row | Synapse stroke paints on hover (`stroke-dashoffset: 1 → 0`). |

### 1.3 The 7 invariants — exact text

| # | Title | Body |
|---|---|---|
| 01 | The Strategist and the Gatekeeper are separate. | The Strategist is any model. The Gatekeeper is deterministic code. They are never the same system. |
| 02 | The Gatekeeper is the only path to physical action. | No model output flows to a click, type, or shell exec without passing the Gatekeeper. |
| 03 | Destructive actions require a real human at the keyboard. | A native modal that halts execution until the human physically allows. No exceptions. |
| 04 | The user can always stop the agent. | A hard hotkey, a watchdog timer, network isolation, a menu-bar kill. Four independent mechanisms. |
| 05 | Every action is auditable. | HMAC-chained, append-only, tamper-resistant. If something goes wrong, we can prove exactly what happened. |
| 06 | The agent is a guest, not an owner. | It requests permission to enter rooms. The user grants or denies. We never escalate, never bypass. |
| 07 | OS permissions are granted by the user, on their machine. | We don't have access. We ask, they grant. The onboarding makes this easy and clear. |

**Provenance.** Source-of-truth: `CLAUDE.md §2.1` (the Seven Non-Negotiable
Invariants). Future revision of any row body must be a `CLAUDE.md` amendment;
do not edit About copy in isolation.

### 1.4 Monospace citation (per row)

Each row should carry a `font-mono · 12px · +0.12em` citation that names the
file in the daemon that enforces the invariant. Hover reveals it; it is
otherwise hairline-faint. This is a **spec target**, not yet wired on every
row — see §8 Drift.

| # | Citation (target) | Notes |
|---|---|---|
| 01 | `internal/gatekeeper/policy.go:42` | where `Strategist ≠ Gatekeeper` is asserted. |
| 02 | `internal/gatekeeper/gate.go:17` | the only path to `Click/Type/Exec`. |
| 03 | `internal/gatekeeper/consent.go:88` | native modal block. |
| 04 | `internal/halt/halt.go:31` | the three-layer halt. |
| 05 | `internal/audit/chain.go:55` | HMAC chain link. |
| 06 | `internal/safety/permission.go:14` | guest-not-owner. |
| 07 | `internal/permissions/permissions.go:23` | OS-level grants. |

### 1.5 Footer colophon

| Item | Destination | Style |
|---|---|---|
| `Condura · v0.1.0 · free for personal and commercial use.` | — | inline italic display |
| `EULA` | `#/about` (hash route — same page; preserved so the link stays "live" for future deep-links to the EULA section) | text link, hover = synapse + 8% wash + thread underline |
| `Privacy` | `#/privacy` | text link |
| `Support Condura` (donate) | `https://synaptic.app/donate` external — `target=_blank rel=noopener` | pollen-colored link (the only pollen-CTA in the document) |

---

## 2. STATE MATRIX

| State | What renders | Trigger | Degraded behavior |
|---|---|---|---|
| **Default** | Hero + 7 invariants + footer | mount with `IntersectionObserver` available | — |
| **Loading** | (none — page is static) | — | — |
| **Error** | Page renders normally; colophon row degrades to `— · — · version unavailable` if version manifest fails to load | `version.manifest` RPC rejects | Hairlines and armor rect still draw; only the version substring in the colophon is replaced with `version unavailable`. No dead wall. |
| **Empty (no JS)** | Hero + ledger + footer, all visible, all visible invariants render as static text. No motion. Hairlines render at full width (no draw-in). | `<noscript>` fallback | Document is still legible; the seven promises still read. |

There is **no loading state** for the seven invariants — they are a hard-coded
array in `About.svelte:19–55`. There is **no per-row error state** — the rows
do not fetch data.

---

## 3. MOTION CHOREOGRAPHY

Three motion systems own this surface. All three come from `DIRECTION.md §5`.

### 3.1 Choreography timeline

| t (ms) | Event | Property | Duration | Easing | Token |
|---|---|---|---|---|---|
| 0 | mount | — | — | — | — |
| 0 → ~`--dur-cine` (900ms) | (reserved for hero wordmark if a draw-in is added — currently the title is static; see §8 Drift) | `stroke-dashoffset` SVG | 900ms | `--ease` | `--dur-cine` |
| mount + `requestAnimationFrame` | colophon Thread + Pulse appear together | `draw=true` + Pulse mounts | 520ms (Thread) | `--ease` | `--dur-slow` |
| each row entering viewport (IO threshold 0.35) | row hairline draws L→R | `stroke-dashoffset: 1 → 0` | 520ms | `--ease` | `--dur-slow` |
| on hover of a row | row lifts `translateY(-1px)`, background → `surface-card`, armor rect paints in | `transform / background / stroke-dashoffset` | 280ms | `--ease` | `--dur` |
| on hover of a row | row-n `opacity: 0.55 → 1` | `opacity` | 280ms | `--ease` | `--dur` |
| on hover of colophon link | link gets synapse wash + thread underline `scaleX(0 → 1)` | `background / color / transform` | 280ms | `--ease` | `--dur` |
| on `:active` of any interactive | `transform: scale(0.97)` | `transform` | 140ms | `--ease` | `--dur-fast` |
| on `:focus-visible` | `box-shadow: 0 0 0 4px var(--pollen-halo)` (rounded — rows are `r-md`) | `box-shadow` | 140ms | `--ease` | `--dur-fast` |

### 3.2 Scroll-linked active state (spec target)

When a row is in the active scroll position (the row whose center is closest
to viewport center), it gets a left-border accent:

| Property | Value | Token |
|---|---|---|
| `border-left` | `2px solid var(--synapse)` | — |
| `padding-left` shift | `+ 2px` to compensate | — |
| `transition` | `border-color / padding / opacity` | `--dur` `--ease` |

This is **not yet wired** — current About.svelte only has the IO-driven hairline.
See §8 Drift.

### 3.3 Hover citation underline

Hovering a row reveals the monospace citation (currently absent; see §1.4).
The citation underline uses the Thread gesture:

```css
.citation::after {
  content: '';
  height: 1px;
  background: currentColor;
  transform: scaleX(0);
  transform-origin: left;
  transition: transform var(--dur) var(--ease);
}
.citation:hover::after { transform: scaleX(1); }
```

### 3.4 Reduced-motion

`prefers-reduced-motion: reduce` is owned by `condura.css` (DIRECTION.md §5
"reduced-motion contract"). About.svelte additionally overrides locally
(lines 409–418) to short-circuit:

| What | Reduced-motion behavior |
|---|---|
| `.row:hover` background | `background: transparent` (no lift tint) |
| `.armor rect` transition | `transition: none` |
| `.hairline line` transition | `transition: none` (hairlines render at full width immediately) |
| `.colophon a::after` transition | `transition: none` (underline shows immediately) |

**Spec rule:** do not introduce a wordmark draw-in to the title without also
adding a `prefers-reduced-motion` override for it. Stagger logic (60 ms per
row) must also collapse to `0ms` on reduced-motion.

---

## 4. KEYBOARD

| Key | Action | Where | Notes |
|---|---|---|---|
| `Tab` | moves focus through interactive elements (links, the donate link) | page-level | Each row's citation link (when wired) is focusable. |
| `Shift+Tab` | reverse tab | — | — |
| `Enter` / `Space` | activates the focused link | — | — |
| `Esc` | no-op (no overlay on this page) | — | — |
| `⌘D` / `Ctrl+D` | open `Support Condura` donate link in a new tab | page-level | Reserved chord. Currently the link is reachable by Tab. |

**No other chords.** This page is intentionally keyboard-light — the user
reads, not acts.

---

## 5. COMPONENTS USED

| Component | Import path | Where used | Spec |
|---|---|---|---|
| **`InvariantEntry`** | `./InvariantEntry.svelte` (target — see §8) | each of the 7 rows | Encapsulates `.row` + hairline + armor + body. Props: `{n, title, body, citation, visible}` (drives the IO-driven draw-in via `data-line` on the host). |
| **`Thread`** | `./Thread.svelte` | footer's `breath-thread` | `orientation="h"`, `draw={colophonIn}` (true after first rAF). Same language as the titlebar thread (`TitlebarThread.svelte`). |
| **`Pulse`** | `./Pulse.svelte` | hero `cred` (`phase="thinking"`) + footer `breath-thread` (`phase="idle"`) | size=6 in both slots. |
| **`Glyph`** | `./icons.ts` (target — see §8) | info glyph on the hero eyebrow row, donate glyph | Reserved for one or two slots. Not currently rendered — the titlebar already carries the wordmark; do not duplicate a Condura glyph here. |
| **`Button`** | `./Button.svelte` (target — see §8) | EULA link, donate link | Use `variant="ghost"` for inline text links, `variant="primary"` (pollen) for the donate CTA if it graduates from a link to a button. |
| **`Tooltip`** | `./Tooltip.svelte` (target — see §8) | hovering the monospace citation reveals the file path | hover-delay 400ms, exit 75ms (per MOAT §2.9). Stops using `title=`. |

**Imports in the current code:** `Pulse`, `Thread`, and `ROUTE_HASH` from
`NavRail.svelte`. The rest are spec targets — components the page should
consume to escape the MOAT §1.3 "hero/eyebrow/headline/sub hand-rolled in
every route" finding.

---

## 6. DATA FETCHED

| Source | When | What | Failure mode |
|---|---|---|---|
| `version.manifest` (daemon RPC) | once on mount | `{ build_hash, commit, date, eula_version }` | If RPC fails: colophon line shows `— · — · version unavailable` in place of `v0.1.0 · build_hash`. The page still renders. |

**Nothing else.** The seven invariants are a hard-coded array. The `DONATE_URL`
is a hard-coded constant. No telemetry, no per-row fetches, no usage tracking.

This page is the **proof point for `DIRECTION.md §1` "Local-first feels
local"** — there's no spinner because there's nothing to wait for.

---

## 7. DESIGN DECISIONS

| # | Decision | Rationale | Source |
|---|---|---|---|
| D1 | **No fake enthusiasm.** Hero sub is honest ("free, local-first, no telemetry, no lock-in"), not "Welcome to the future!" | MOAT §4 rule 5. | MOAT §4 |
| D2 | **The 7 invariants are non-negotiable.** The copy on the page is the verbatim content of `CLAUDE.md §2.1`. Any row edit must be a CLAUDE.md amendment first. | Survival Rule. | CLAUDE.md §2.1 |
| D3 | **No `.alive` italic-green on this surface.** The `.alive` accent (MOAT §1.7) is already used once — on First Breath in the Ritual ("Condura is *here*"). Using it here too would be a tic. The hero title is upright Instrument Serif, not italic-synapse. | MOAT §1.7 (one load-bearing phrase per surface, never more). | MOAT §1.7 |
| D4 | **Thread separators are the same language as the titlebar.** The footer breath-thread uses `<Thread>` (`DIRECTION.md §5 THE SIGNATURE`). The row hairlines use the same `pathLength=1 / stroke-dasharray=1 / stroke-dashoffset 1→0` gesture. The user learns one motion vocabulary. | DIRECTION.md §5 "where the Thread MUST appear" + MOAT §3. | DIRECTION §5, MOAT §3 |
| D5 | **Hover lifts the row `translateY(-1px)` (not `-4px`) and tints background to `surface-card`.** A 4px lift is what Skills cards do; About rows are text-led, not card-led, so the lift is subtle. The armor rect carries the "protection" gesture. | Per-surface restraint. About is a quiet surface; it shouldn't out-loud the cards. | MOAT §1.5 (no `rotateX(2deg)` flex) |
| D6 | **The donate link is pollen-colored; every other link is content-colored.** Pollen is reserved for action CTAs (DIRECTION §4 "Brand has two colors: synapse and pollen. Everything else is ink on paper."). One pollen CTA per document — the donate link earns it. | DIRECTION §4 + MOAT §4 rule 3. | DIRECTION §4 |
| D7 | **No spinner, no Pulse-as-loader.** The whole page renders synchronously. The Pulse appears in the `cred` slot as a heartbeat (always-on), not as a loading state. | MOAT §4 rule 7 + DIRECTION §6 rule 7. | DIRECTION §6 |
| D8 | **Reduced-motion is owned by `condura.css` AND by a local About block.** Per MOAT §2.3 the local block is a debt that should collapse into the global; until then, the local block stays because the row hover lift is a Surface-specific concern. | MOAT §2.3 acknowledged debt. | MOAT §2.3 |
| D9 | **Citation is monospace, hairline-faint, hover-revealed.** The user sees the promise on first read; the file path appears on attention. This is the "premium tell" — the code is verifiable, not just claimed. | $50M-feel principle: 50 micro-decisions that don't contradict each other. | MOAT §5 |
| D10 | **No emoji as UI glyphs.** Donate is a text link with `.donate` class; the info marker (if added) is a `<Glyph>`, not a 🛈. | MOAT §4 rule 2. | MOAT §4 |
| D11 | **No gradient text.** Title, sub, row text, colophon — all single-color. Background paper may carry grain/bloom; text never does. | MOAT §4 rule 1. | MOAT §4 |
| D12 | **No rectangular focus outlines.** Rows use the rounded-focus halo (`box-shadow: 0 0 0 4px var(--pollen-halo)` matching `r-md`). | MOAT §4 rule 8 + DIRECTION §6 rule 6. | DIRECTION §6 |
| D13 | **No double shadows.** The row hover uses one token (`--shadow-card`), not `--shadow-card` + a custom box-shadow. The armor rect is an SVG stroke, not a shadow. | MOAT §4 rule 9. | MOAT §4 |
| D14 | **Page is reachable from anywhere.** NavRail item 10 of 10; ⌘K palette command; the footer of any surface can link here. The 7 invariants are not hidden behind a Settings menu — they are the front door. | APPFLOW.md §4.9 + I5 ("The 7 invariants are visible"). | APPFLOW §1, §4.9 |
| D15 | **No live data, no telemetry.** The page does not call `track()` or send a beacon. If we wanted to know which invariant the user hovered longest, we'd add a local-only counter — but the Survival Rule forbids that without consent. | CLAUDE.md §2.1 invariant #5 (auditable) implies "don't add a side channel." | CLAUDE.md §2.1 |

---

## 8. DRIFT TABLE

This section tracks what changed (or should change) between the original
About surface (a hand-rolled hero + ledger) and the spec'd version.

| Status | Item | What | Why | Source |
|---|---|---|---|---|
| **REMOVED** | hand-rolled `.head` / `.title` / `.sub` / `.eyebrow` | Replace with shared `RouteHero.svelte` (size tokens, one canonical hero block). | MOAT §1.3 — hero is hand-rolled in 6 different ways. About is one of them. | MOAT §1.3 |
| **REMOVED** | hand-rolled `.ledger-eyebrow` | Replace with `<Eyebrow>` component (the small mono-uppercase element). Same reason. | MOAT §1.3. | MOAT §1.3 |
| **REMOVED** | local `prefers-reduced-motion` block | Collapse into `condura.css` global rule. | MOAT §2.3. | MOAT §2.3 |
| **REMOVED** | inline `<a onclick={(e) => { e.preventDefault(); window.location.hash = ...; }}>` for EULA + Privacy | Replace with `<a href="#/about">` native hash navigation; let the global hash listener handle the rest. | DRY — two bespoke click handlers do the same thing the global router does. | internal review |
| **REMOVED** | ad-hoc route constants | Replace with `ROUTE_HASH.about` (already imported from NavRail). | Same. | About.svelte:6 |
| **REMOVED** | comment block describing what the page does | Once `InvariantEntry` exists, the comment is redundant with the component name. | SPEC_ABOUT §1 supersedes inline narrative. | About.svelte:7–12 |
| **ADDED** | `InvariantEntry.svelte` | New component encapsulating the row + hairline + armor + body + IO-driven visibility. Props: `{n, title, body, citation, visible}`. | DRY the row across the 7 instances; make the citation a first-class slot. | SPEC_ABOUT §1.2, §5 |
| **ADDED** | monospace citation on each row (`internal/gatekeeper/policy.go:42` etc.) | First-class slot in `InvariantEntry`. Hover reveals; idle state is hairline-faint. | SPEC_ABOUT §1.4, §7 D9. | SPEC_ABOUT §1.4 |
| **ADDED** | scroll-linked active state (left-border accent on the row nearest viewport center) | New effect on `IntersectionObserver` (or scroll listener). Spec target; not yet wired. | SPEC_ABOUT §3.2. | SPEC_ABOUT §3.2 |
| **ADDED** | `Tooltip` for the citation hover | Replaces a future `title=` attribute. 400ms hover-delay, 75ms exit. | MOAT §2.9. | MOAT §2.9 |
| **ADDED** | `version.manifest` fetch + degraded colophon | Page must surface a real build hash + commit + date; today's page says only `v0.1.0`. | SPEC_ABOUT §6. | SPEC_ABOUT §6 |
| **ADDED** | `Glyph` import slot for one info marker (optional) | Reserved for a future "what's this?" marker beside the hero eyebrow. Do not introduce the Glyph until it carries meaning. | MOAT §1 (no decoration for its own sake). | MOAT §1 |
| **ADDED** | `⌘D` keyboard chord for donate | Page-level chord, currently reserved. | SPEC_ABOUT §4. | SPEC_ABOUT §4 |
| **ADDED** | `<noscript>` fallback | Hero + ledger + footer all render as static text. Hairlines at full width. | SPEC_ABOUT §2 (Empty state). | SPEC_ABOUT §2 |
| **DEFERRED** | wordmark draw-in animation on the title | Currently the title is upright and static. If added, must honor reduced-motion. Not needed; the title is small and the Thread below does the work. | Restraint. | SPEC_ABOUT §3.1, §8 D3 |
| **DEFERRED** | Replace text-link `EULA` with a `<Button variant="ghost">` | Cosmetic; the inline link works fine. | Defer. | — |
| **REJECTED** | Italic-synapse "alive" accent on the title | Already used on First Breath in the Ritual; using it here too would be a tic. | MOAT §1.7. | MOAT §1.7 |
| **REJECTED** | Emoji glyphs anywhere | Forbidden. Use `<Glyph>` from `icons.ts`. | MOAT §4 rule 2. | MOAT §4 |
| **REJECTED** | Spinner loader | There is no async work. A spinner would lie. | MOAT §4 rule 7. | MOAT §4 |
| **REJECTED** | Gradient text on any element | Forbidden. Text is one color per role. | MOAT §4 rule 1. | MOAT §4 |
| **REJECTED** | Adding a new brand color for the donate link | The donate link already uses pollen (the CTA color). Adding a third brand color requires a CLAUDE.md amendment. | MOAT §4 rule 3. | MOAT §4 |

---

## Provenance & cross-references

| Doc | What it says about About |
|---|---|
| `CLAUDE.md §2.1` | The seven non-negotiable invariants — verbatim source for the seven rows. |
| `CLAUDE.md §21` | About is one of the Wails main-window surfaces. |
| `DIRECTION.md §1` | Personality contract: "paper notebook that learned to listen" — this is the surface that most embodies it. |
| `DIRECTION.md §4` | Color tokens. The page must read `--paper`, `--ink`, `--synapse`, `--pollen` only. |
| `DIRECTION.md §5` | The Thread + the reduced-motion contract. |
| `DIRECTION.md §6` | The seven anti-patterns (no gradient text, no emoji UI, no rainbow, no fake enthusiasm, no spinners, no rectangular outlines, no double shadows, no ornament). |
| `MOAT.md §1.3` | Hero/eyebrow/headline/sub is hand-rolled in every route — About is one of them. Drift §8 marks the fix. |
| `MOAT.md §1.7` | The `.alive` accent is used 6+ times. About does NOT use it. |
| `MOAT.md §2.3` | Reduced-motion is re-declared in 4 components. About is one of them. Drift §8 marks the fix. |
| `MOAT.md §2.9` | No tooltip component exists; About's citation needs one. |
| `MOAT.md §3` | The Thread — every surface must ship at least one. About ships three (row hairlines, footer thread, link underlines). |
| `MOAT.md §4` | The 10 anti-patterns. About is exemplary: it triggers none. |
| `APPFLOW.md §4.9` | About's route reference. |
| `APPFLOW.md §6` | Edge cases — About has no edge cases except the version-manifest degradation (SPEC §2). |

---

**Spec lock.** Any change to the seven invariants' copy, the hero title, or
the colophon line is a CLAUDE.md amendment. Any change to the motion
choreography, layout, or component composition is an edit to this file in
the same commit as the code change.