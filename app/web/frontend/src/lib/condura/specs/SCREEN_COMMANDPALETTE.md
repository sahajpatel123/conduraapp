# SCREEN_COMMANDPALETTE · Condura · The `⌘K` Power-User Surface

> **Screen architecture spec for `CommandPalette.svelte`.** Phase 4
> implements against this document; it is the contract between the
> design intent in `MOAT.md`/`APPFLOW.md`/`TEARDOWN.md` and the
> shipping code.
>
> **Reading order for the next agent.** Read §1 (drift) — it tells
> you what changes from the existing `CommandPalette.svelte`. Read
> §2 (layout) — the geometry. Read §3 (state matrix) and §4
> (motion) — the behavior. Skip §5–§8 only if you already know the
> app.
>
> **Source-of-truth docs** (read before this one):
> - `MOAT.md` — quality bar. The palette must pass §1 (restraint),
>   §2 (detail), §3 (signature), §4 (anti-patterns).
> - `APPFLOW.md` §3.4 — palette interaction in the global shell.
> - `TEARDOWN.md` §4 (Linear) — the meta-UI escape-valve pattern.
>
> **North-star docs cited by the brief but not on disk at this
> commit.** `DIRECTION.md` and `DESIGNLANG.md` are referenced in the
> Phase 1 brief but absent from `app/web/frontend/src/lib/condura/`.
> This spec leans on `MOAT.md` (the design grammar), `APPFLOW.md`
> (the surface boundaries), and on the live token table in
> `condura.css`. Where the brief and the codebase disagree (e.g. on
> max-width, on category groupings, on `⌘1`–`⌘9` chords), this spec
> honors the brief — the existing `CommandPalette.svelte` is the
> Phase 3 baseline, not the target.
>
> **Existing implementation.** `CommandPalette.svelte` (current
> file, 580 lines) is the Phase 3 baseline; this spec describes the
> Phase 4 redesign that replaces it.

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
10. [Implementation Notes for Phase 4](#10-implementation-notes-for-phase-4)
11. [Test Plan](#11-test-plan)

---

## 1. Spec vs. Implementation Drift

What this spec changes in the current `CommandPalette.svelte`. Phase 4
must remove the old, ship the new, in one atomic diff — no half-states.

| # | Today (current `CommandPalette.svelte`) | Phase 4 (this spec) | Why |
|---|---|---|---|
| 1 | Width `min(560px, 92vw)`. | **Width `min(640px, 92vw)`.** | Brief §1 — wider panel accommodates category groupings without row wrap. |
| 2 | Vertical offset `padding-top: 16vh`. | **Vertical offset `padding-top: 120px` from the viewport top.** | Brief §1 — fixed offset anchors the palette near the top, frees the view below. Below 720px viewport, `padding-top: 64px`. |
| 3 | `min-height: var(--space-9)` (36px) row, single list. | **Rows are category-grouped: Routes / Actions / Skills / Conversations / Docs.** Each group renders under its own `<CategoryGroup />` with the group label as a 11px mono caps eyebrow. | Brief §1 — the palette must feel instant *and* scannable; groups give the eye a heading to track even with no query. |
| 4 | Filter is `commands.filter(c => c.label.toLowerCase().includes(q))`. Two hardcoded kinds (`nav` + `action`). | **One search across 5 sources** (routes, actions, skills, conversations, docs) via `ipc.commandPalette.search(q)`, debounced 150ms. Client-side results are merged with server-side results into a single ranked list. | Brief §6 — fuzzy search is the whole job of the surface; one query hits everything. |
| 5 | `commands` is a static array (3 actions + 10 routes) — keyboard navigation uses the active index. | **The list is sourced from the IPC response, filtered by the query, and re-ranked.** A `RouteId` route still ends up as a nav command; a Skill ends up as an "Open in Chat" command; a Doc ends up as a "Search docs for query →" command; a Conversation ends up as a "Resume {title}…" command. | The palette is the meta-UI (Linear §4 in `TEARDOWN.md`); every routable destination becomes one of these 5 sources. |
| 6 | Three ColumnGrid + three, hardcoded actions (`theme`, `summon`, `stop`). | **Same three actions stay** (they are first-class shell verbs, not routes) — but rendered as a group ("Shell") beneath Actions, before Skills etc. The group label "Shell" replaces "Actions" to be honest about the distinction. | The mood here must stay precise. The Skill routes are in the Routes group (because Skill Hub is one of the 10 routes). The "Shell" group is for things that *aren't* routes. |
| 7 | Sliding highlight is one absolute-positioned `.highlight` whose `top` follows `rowEls[i].offsetTop` via CSS `transition: top 200ms var(--ease)`. | **The highlight uses the same FLIP technique but is a `<SlidingHighlight />` primitive** — a separate component with `lines: HTMLElement[]` (the row set) and an animated `transform: translateY(...)`. Reusable. | MOAT §3 — the signature flourish is one element, repeated everywhere. The NavRail uses it (active item); the Settings rail will; the palette will. |
| 8 | Match-flash on the active row's `↩` chip is `<span class="row-hint" class:flash={i === activeIndex}>` with `{#key flashKey}` to replay the keyframe when focus moves. | **Delete the match-flash.** The active row's right-aligned chord glyph stays, but does not animate per-active-row. The user's "press enter to run" affordance is communicated by the row's halo on focus + the footer "↵ run" hint, not by a per-row pulse. | MOAT §1 — restraint. The match-flash cycles on every arrow keypress and competes with the sliding highlight; it's noise. |
| 9 | `box-shadow: var(--shadow-float)` for elevation. | **Same `--shadow-float`.** No change. | MOAT §4 #3 — floating surfaces that earn elevation use the token. The palette is a dialog and earns elevation. |
| 10 | `backdrop-filter: blur(8px) saturate(0.9)` on the scrim. | **Same. The palette is one of the three surfaces where glassmorphism is permitted** (per MOAT §4 #3 — "a floating surface that needs to look elevated"). No change. | The palette is a meta-UI dialog; the scrim needs to dim the route behind it. |
| 11 | Panel entrance: `animation: palette-in var(--dur-slow) var(--ease) both;` with `from { opacity: 0; transform: translateY(10px) scale(0.97); filter: blur(6px); }` over 520ms. | **Entrance: `fade + scale(0.96 → 1)` over 200ms `--ease`.** No blur-in. No translateY. The brief's motion choreography is plain: the panel arrives, focus comes to the input, results are already there if cached. | Brief §3 — "the overlay fades in + scales from 0.96 → 1 over 200ms." A 520ms entrance is too long for a power-user shortcut. |
| 12 | Input focus uses an `input-draw` keyframe on `:focus` (240ms `box-shadow: inset 0 -1px 0 var(--synapse)` center-out). | **Focus-thread on the input: a `<Thread />` that draws across the top of the input** (not under it) over `--dur-slow` (520ms). This is the same gesture as the composer's focus state (`Chat.svelte:515–528`, MOAT §5.2). | Brief §1 — focus halo is the composer's focus-thread. The palette's input is a sibling of the composer; they should share the gesture. |
| 13 | Filter results re-render with no cross-fade. | **Results cross-fade 120ms** when the filtered set changes (entire `<CategoryGroup />` lists). Brief §3 — "filter in with a 120ms cross-fade (no row-by-row stagger)." | Per-row stagger feels slow for a power-user shortcut. One cross-fade, immediate. |
| 14 | No `⌘1`–`⌘9` quick-select chords. | **`⌘1`–`⌘9` quick-selects the first 9 visible results** (across groups, in order). Brief §4. | MOAT §2.10 — keyboard chords are part of the surface. Power users pick the third result with `⌘3`. |
| 15 | Empty state is `<div class="empty-head">Nothing matches this search.</div><div class="empty-sub">No command for "{query}". Try a route name, "theme", or "stop".</div>` | **Empty state rephrased to "No commands match — try a route or a skill."** plus an instruction copy line per the MOAT §2.4 three-line pattern. The italic display headline is unchanged. | Brief §2 (`open-no-results`). The three-line teach pattern is a project-wide rule. |
| 16 | Reduced-motion: backdrop-filter removed, `panel::before` hidden, search input no box-shadow, no match-flash animation. | **Reduced-motion: skip the entrance scale + skip the FLIP highlight slide.** Instant appearance. Top-edge synapse hairline still present (it's a static 1px line), but does not animate. The cross-fade on results still happens (it's a fade, not a movement; fades are accessible). | Brief §4 — `prefers-reduced-motion: reduce` is the only condition under which the highlight is instant. |
| 17 | `binding` — input is `bind:value={query}` and `bind:this={inputEl}`. | **Same bindings, plus `bind:this={listEl}`** (the scrollable `<div class="list">`) so the FLIP-Group parent can compute row offsets from a single source of truth. | The `SlidingHighlight` reads `listEl.children` for FLIP measurements; owning the ref here means the primitive stays dumb. |
| 18 | `class:flash` uses a keyblock `{#key flashKey}` to re-trigger the keyframe. | **Delete the match-flash entirely.** The new affordance is the focus halo on `:focus-visible` of the active row (2px synapse ring + 5px pollen halo per MOAT §2.1, rounded because rows have `border-radius: var(--r-sm)` ≥8px). | MOAT §1 — restraint. The match-flash competed with the sliding highlight. |
| 19 | Footer is `<kbd>↑</kbd><kbd>↓</kbd> navigate`, `<kbd>↵</kbd> run`, `<kbd>esc</kbd> close` — three hints, always visible. | **Footer adds `<kbd>⌘1-9</kbd> quick-select`** as a fourth hint. Same footer visual language. | Brief §1 — the chord hint is in the footer. The user needs to discover `⌘1`–`⌘9`. |
| 20 | Recent commands list is not implemented. | **`recent` placeholder rows render in the open-empty state** (when `query === ''` and the palette just opened). They are the last 5 commands the user ran. Cap at 5. | Brief §2 (`open-empty`) — the recent-commands list belongs here. |
| 21 | `CategoryGroup` is not a component; the surface has one flat `<div class="list">`. | **`<CategoryGroup />` is its own primitive** — used only here in v0.1.0, but reusable (Settings rail groups, Onboarding step lists). Props: `{ label, count?, children }`. Owns the mono-uppercase label and the gap to its rows. | MOAT §1.2 — extract repeated blocks into primitives. One eyebrow, one definition. |

---

## 2. Layout

### 2.1 Geometry

The palette is a **single-purpose overlay** over a blurred scrim,
anchored near the top of the viewport. It is rendered at `Shell.svelte`
level (`<CommandPalette open={paletteOpen} onclose={...} onnavigate={...} />`,
line 251), not at route level. It is `position: fixed`, `inset: 0`, with
`display: grid; align-items: start; justify-items: center; padding-top: 120px`.

```
Viewport (≥720px)                              Viewport (<720px)
┌──────────────────────────────────┐          ┌────────────────────────┐
│                                  │          │                        │
│  ┌────────────────────────────┐  │          │ ┌────────────────────┐  │
│  │  🔎  Type a command or…    │  │          │ │ 🔎 Type…           │  │
│  │  ─────────────────────────  │  │          │ │ ────────────────   │  │
│  │  ROUTES                      │  │          │ │ ROUTES             │  │
│  │   📍  Chat                    │  │          │ │  📍 Chat           │  │
│  │   📍  Hub                     │  │          │ │  📍 Hub            │  │
│  │   …                          │  │          │ │  …                 │  │
│  │  SHELL                        │  │          │ │ SHELL              │  │
│  │   ☀  Toggle Theme        ↵    │  │          │ │  ☀ Toggle Theme ↵  │  │
│  │  SKILLS                        │  │          │ │ SKILLS             │  │
│  │   ⬚  morning briefing    ↵    │  │          │ │  ⬚ morning brief   │  │
│  │   ⬚  summarize a PDF     ↵    │  │          │ │  ⬚ summarize PDF ↵ │  │
│  │  CONVERSATIONS                 │  │          │ │ CONVERSATIONS      │  │
│  │  DOCS                          │  │          │ │ DOCS               │  │
│  │  ─────────────────────────  │  │          │ │ ────────────────   │  │
│  │  ↑↓ navigate · ↵ run …       │  │          │ │ ↑↓ · ↵ · ⌘1–9 · esc │  │
│  └────────────────────────────┘  │          │ └────────────────────┘  │
│ 640px wide, 120px from top        │          │ 92vw wide, 64px from top │
│                                  │          │                        │
└──────────────────────────────────┘          └────────────────────────┘

Cumulative example shows the staggered focus-thread on top of the search
input + the pollen-highlight sliding between rows. The scrim behind the
panel is blurred 8px and tinted `--scrim` (the warm, radial gradient on top,
neutral scrim below).
```

The panel's `width` is **`min(640px, 92vw)`**. The panel's `max-height`
is **`72vh`** with `overflow: hidden` on the panel and `overflow-y: auto`
on the list. Empty state, single group, and 50-result list all sit within
this height; we never need to virtualize.

### 2.2 Vertical rhythm — from top of panel down

| y-offset | What |
|---|---|
| `0px` | Top edge — `border-radius: var(--r-lg)` (16px). The `border-top` is a 1px `var(--hair-strong)` hairline. |
| `0px` (the `::before`) | A 1px synapse thread that draws left→right over `--dur-slow` on entrance. The hairline is the spine of the surface (MOAT §3). |
| `0px → 60px` | **Search bar.** `<Glyph name="search" size={18} class="search-glyph" />` on the left. `<input>` filling the rest. `<button class="esc-chip">` showing `esc` on the right. `padding: var(--space-4) var(--space-5)` (16px vertical, 20px horizontal). The focus-thread draws across the top edge of the input on focus (the composer's gesture, see MOAT §5.2). |
| `60px` | **Hairline.** `border-top: 1px solid var(--hair)` separates search bar from list. |
| `60px → 60px + listH` | **List area.** `padding: var(--space-2)` (8px). The list contains up to 5 `<CategoryGroup />`s. Each row is **44px tall** (a touch-comfortable target per MOAT §2.5 — rows are not symbol-dense; they're a label and a glyph). |
| `60px + listH` | **Hairline.** `border-top: 1px solid var(--hair)`. |
| `60px + listH + 40px` | **Footer.** `padding: var(--space-3) var(--space-5)`. Four chord hints left-aligned, separated by `var(--space-5)` gap. |
| `60px + listH + 80px` | **Bottom edge.** `border-radius: var(--r-lg)`. |

### 2.3 Vertical rhythm — within the list

Each `<CategoryGroup />` is a `<section>` with:

| Block | Content | y-offset within group |
|---|---|---|
| Label | `<h3 class="cg-label">Routes</h3>` — display mono caps 11px, `--content-faint`, padding `var(--space-2) var(--space-3) 0`. | `0px → 24px` |
| Rows | Up to 8 commands per group in v0.1.0; collapse to `+N more` link beneath if overflow. | `24px → 24 + 8 * 44` |
| Pad | `padding-bottom: var(--space-3)` between groups. | — |

Each row is `<button class="row" data-kind={cmd.kind}>` with:

| Block | Content |
|---|---|
| Glyph | `<Glyph name={cmd.icon} size={18} class="row-glyph" />` — `var(--content-faint)` at rest, `var(--pollen)` when active. |
| Label | `<span class="row-label">{cmd.label}</span>` — flex 1, `var(--content-mute)` at rest, `var(--synapse)` when active. |
| Chord | `<span class="row-chord">{cmd.chord ?? '↵'}</span>` — mono 11px, `var(--content-faint)` at rest. The chord is `⌘1`–`⌘9` for the first 9 results and `↵` for everything else. |

Active row carries the sliding pollen highlight (its absolutely-
positioned `transform: translateY(...)` follows the active row). On a
single-row group (e.g. just one Conversation matches), the highlight is
centered on that row.

### 2.4 The sliding highlight — the signature

The highlight is a single `<SlidingHighlight />` instance owned by the
list. It is `position: absolute; left: var(--space-2); right: var(--space-2);
height: 44px; border-radius: var(--r-sm);` with these visual layers:

```css
.sliding-highlight {
  background: color-mix(in srgb, var(--pollen) 8%, transparent);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--pollen) 20%, transparent);
  border-left: 3px solid var(--pollen);
  pointer-events: none;
  z-index: 0; /* rows sit at z-index: 1 */
}
```

The highlight's `transform: translateY(...)` is **FLIP-animated**: when
the active index changes, the primitive measures the previous position
and the new position, computes the delta, applies the inverse transform
instantly (so the visual position is unchanged for one frame), then
transitions `transform` to `(0, newY)` over **240ms `--ease`**. Same
technique as the NavRail's active segment (SCREEN_NAVRAIL §5).

The FLIP technique is the project's signature motion (NavRail,
CommandPalette; the same primitive will live in `Settings.svelte`'s
inner rail). **If a future surface wants a sliding highlight, it
imports `<SlidingHighlight />`.** No re-implementations.

### 2.5 The focus-thread on the search input

A `<Thread />` instance is rendered **above** the input (`position:
absolute; top: 0; left: 0; right: 0; height: 1px;`). The thread's
`pathLength` is 1, `stroke-dasharray: 1`, `stroke-dashoffset: 1` while
idle, transitioning to `0` over `--dur-slow` (520ms) `--ease` on
`:focus-within`. This is the **same gesture** as the composer's
focus-thread (`Chat.svelte:515–528`, MOAT §5.2). The user learns once:
"a 1px synapse line draws across the top of the thing I'm writing in."
The palette inherits that gesture because the input is a sibling of the
composer.

### 2.6 Reduced motion

When `prefers-reduced-motion: reduce` is set:

- The entrance `scale(0.96 → 1)` is dropped. The palette appears at full
  size. The scrim still tints (the visual signal of "an overlay is
  here" must remain).
- The `SlidingHighlight` jumps instantly to the new row's offset. No
  FLIP. The `transition` on the highlight is `none`.
- The top-edge panel-thread draws instantly (or doesn't draw, it's a
  1px hairline — the visual is unchanged).
- The cross-fade on results is a single instant swap (no fade). No
  code path branches; the global `condura.css`
  `@media (prefers-reduced-motion: reduce) { *, *::before, *::after {
  animation-duration: 0.01ms !important; transition-duration: 0.01ms
  !important; } }` does the work. The component declares no
  media-query block.

---

## 3. State Matrix

Six states. The palette is **`open: false` for ~99% of the user's
session time** — closed is the dominant state.

| # | State | Trigger | What renders | Notes |
|---|---|---|---|---|
| **S1** | **Closed** | `open === false` | Nothing. The `{#if open}` block does not render. | The user's app is on screen. No background work, no listener. |
| **S2** | **Open-empty** | `open === true && query === ''` | The panel, the input with placeholder "Type a command…" (placeholder in display italic, `--content-faint`), and the **recent-commands** list (last 5 commands the user ran, capped, deduped by id, persisted via `localStorage` key `condura:palette-recent`). | Default state when the palette opens via `⌘K`. The 5 most recent items are shown — labelled "Recent" as the first group. They are all visible commands if they still exist. If the user has no recent items, the panel shows the empty hint below. |
| **S3** | **Open-with-results** | `open === true && query !== '' && results.length > 0` | The panel, the input, and the grouped results list. Each `<CategoryGroup />` for the sources that returned ≥1 result, in fixed order: **Routes → Shell → Skills → Conversations → Docs**. | Within each group, results are ranked: routes first by route order (`chat` before `hub`), shell actions by `displayOrder`, skills by `lastUsedAt` desc (if known), conversations by `updatedAt` desc, docs by `relevance` desc. |
| **S4** | **Open-no-results** | `open === true && query !== '' && results.length === 0 && !inFlight` | The panel, the input, and the empty state body. | Three-line teach pattern (MOAT §2.4). See §3.6. |
| **S5** | **Loading** | A search is in-flight (`ipc.commandPalette.search` pending). Triggered when `query !== ''` and a new debounced request has fired but not yet resolved. | The input shows a `<Pulse phase="thinking" size={8} />` *next to* a small mono "SEARCHING…" label in the input's right cluster (instead of the `esc` chip). The list still renders the last-good results underneath. The pulse is the only new visual element. The loading state is **never** a full-panel skeleton. | The user's prior results stay visible — there is no "blank screen between searches." When the new results resolve, the list cross-fades (§4.2). When `query === ''`, no pulse is shown. |
| **S6** | **Error** | The IPC rejected. | The list area renders an `<ErrorState />` (`head`, `cause`, `reason`, `onretry`). The footer hints stay visible; the `esc` chip is still present. | Per MOAT §2.6 — ErrorState owns all error rendering on every surface. The palette uses the same component. The retry calls `ipc.commandPalette.search` again with the current `query`. |

State precedence is `S1 < S2 < S5 < S3 < S4 < S6`. Two states can coexist
(e.g. S3 + S5 if a search is in flight), but only one of them owns the
*visual* focus. S3 results stay visible under the S5 pulse.

### 3.1 Closed state (S1) — exact copy

The palette does not render. The application behind it is interactive.
**There is no global keyboard listener while closed** — `⌘K` is a global
shortcut owned by `Shell.svelte:110`, not by `CommandPalette.svelte`. The
palette is a child component.

### 3.2 Open-empty state (S2) — exact copy

**Placeholder (in the input, italic display 14px, `--content-faint`):**

> `Type a command or search…` — display italic 14px, `--content-faint`,
> mono fallback `var(--font-mono)` 13px in light mode for the platform-correct
> feel.

**Recent group label** (the only thing in the panel):

> `RECENT` (mono caps 11px, `--content-faint`).

**Recent rows** (up to 5):

Each row is `<button class="row" data-kind={cmd.kind} data-recent="1">`
with the same anatomy as the active results. They render in `lastRunAt`
descending order. The most-recent row is the default active row (active
index = 0). A trailing row `Clear recent` (mono caps 11px, faint,
ghost-button style) appears at the bottom of the Recent group if any
items exist; clicking it clears `localStorage[condura:palette-recent]`
and re-renders the group empty.

If recent is empty (first ever `⌘K` press), S2 substitutes the follow
default copy (still in italic display, three-line per MOAT §2.4):

> **What's here.** A search bar over every routable surface in
> Condura — routes, shell verbs, skills, conversations, and docs.
>
> **Why it's empty.** You haven't run a command yet. Type a route
> name (`hub`, `audit`, `settings`), or `theme`, `stop`, or
> `summon` to start.
>
> **Next action.** Press `⌘1`–`⌘9` to quick-pick from any group, or
> type to search across everything.

(The empty hint is rare — most first-time users land in S3 immediately
because they type `chat` or `skills` or `stop`.)

### 3.3 Open-with-results state (S3) — exact copy

The list renders groups in this fixed order, even when some groups are
empty:

1. **Routes** — `chat`, `hub`, `skills`, `sync`, `audit`, `replay`,
   `channels`, `delegation`, `settings`, `about`. Same as NavRail.
2. **Shell** — three commands, in fixed order: `theme`, `summon`,
   `stop`. (These are first-class shell verbs that are NOT routes.)
3. **Skills** — installed skills from `ipc.skills.list`, plus Hub-
   browsable skills from `ipc.hub.search`. Each row's label is the
   skill name (`morning briefing`); the chord is `Open`. Pressing
   Enter on a skill sends a `chat/send` with the skill's command
   string (`/run morning-briefing`).
4. **Conversations** — recent conversation threads from
   `ipc.conversations.list`. Each row's label is the conversation
   title (or first 60 chars of the first user message); the chord is
   `Resume`. Pressing Enter navigates to `#/chat?conversation=<id>`.
5. **Docs** — search results from `ipc.docs.search` for the local
   `condura` corpus (CLAUDE.md, APPFLOW.md, MOAT.md, TEARDOWN.md,
   the local `app/web/frontend/src/lib/condura/` design-system page).
   Each row's label is the doc heading + page; the chord is `Open`.

A group whose source returned zero results for the current `query` is
**omitted** from the list entirely. Empty groups have no use in a
power-user tool — they create the impression that the search failed.
The group order is fixed regardless of which groups are present, so
the eye can scan top-to-bottom and find the category it's looking for.

Each row's right-hand chord is `⌘N` for the first 9 visible results
across all groups, or `↵` for everything else. The chord is mono 10px,
`--content-faint`, gutter 12px from the label.

### 3.4 Open-no-results state (S4) — exact copy

Three-line teach per MOAT §2.4. The input keeps the user's query and
the search is re-runnable.

> **No commands match.**
> Display italic 22px, `--content`, line-height 1.15.
>
> Why — your query for "asdf" matched no route, no shell verb, no
> installed skill, and no doc. The query is fuzzy across the
> categories named above.
> Display italic 13px, `--content-soft`, line-height 1.55,
> max-width: 48ch.
>
> Try — a route name like `audit`, a shell verb like `theme` or
> `stop`, or a substring from a skill you remember installing.
> Italic 13px, `--content-faint`.
>
> Below all of this, a single mono caps `SUGGESTED ROUTES` group (the
> first 4 routes from the Routes group) renders so the user has
> something to pick from without retyping. The Suggested group is
> labeled with the `RECENT` chrome pattern.

### 3.5 Loading state (S5) — exact copy

The input's right cluster swaps from the `esc` chip to:

```
[ ◌ SEARCHING · ⌘K to close ]
```

The `◌` is `<Pulse phase="thinking" size={8} />` (the same 8px
breathing glyph used elsewhere). The label is mono caps 11px,
`--content-faint`. The list renders the **last-good filtered results**
underneath — they do not blank out during the search. When the IPC
resolves:

- On success: the list cross-fades (§4.2) to the new results.
- On failure: the list keeps the last-good results but an inline
  mono caps 11px line appears under the footer: `Couldn't refresh ·
  showing last 6 commands`, in `--content-faint`. The user can keep
  navigating the stale list without retyping.

### 3.6 Error state (S6) — exact copy

The list area renders one `<ErrorState />` (per MOAT §2.6) with:

> **Couldn't reach the search index.** (italic display 22px)
>
> Cause: `ipc.commandPalette.search` returned no results in `n` ms
> or rejected with `{error}`.
>
> Likely reason: the daemon was restarted, or the index is rebuilding.
>
> Next action: `[Try again]` (mono-pollen pill) — retry the search.
> Above it, mono 11 `[Search docs only →]` if the failure persists.

The `<ErrorState />` is the same component Chat / Skills / Channels
use, after `MOAT.md §1.2` extraction. No inline `err-state` here.

---

## 4. Motion Choreography

Every animation answers the question (MOAT §4 #10): *What is this
communicating?* The palette has eight meaningful motions, all of which
are listed here. Decorative loops are forbidden.

### 4.1 Entrance — overlay fades in + scales, panel draws a thread on top

Trigger: `open` transitions `false → true`. The `{#if open}` block
mounts.

```
Scrim: opacity 0 → 1 over 200ms --ease
       + radial-gradient bloom at top (already in token)
       + backdrop-filter: blur(8px) saturate(0.9) (already on)

Panel: opacity 0 → 1 over 200ms --ease (faded with the scrim)
     + transform: scale(0.96) → scale(1) over 200ms --ease
       (transform-origin: top center)

Panel top edge: 1px synapse <Thread /> draws left→right
       over --dur-slow (520ms) --ease, starting 60ms after the panel
       mounts (so it doesn't compete with the entrance scale)

Focus-thread on input: starts drawing --dur-slow (520ms) --ease
       after the input receives focus (delayed 80ms by tick()
       so the focus-thread doesn't fire before the entrance).
```

The entrance is **fast** — 200ms total, with the thread as the
afterbeat. A power-user shortcut must not have a 520ms entrance
("the user is summoning a tool, not watching a logo reveal"). The
thread-draw is the afterbeat so the panel still has the signature
gesture, but it does not block the user's keyboard.

**Reduced motion:** the scale is dropped. The panel appears instantly
at full size. The thread-draw is unchanged (it's a quiet 1px line;
the visual end-state is identical).

### 4.2 Results cross-fade — the only filter-in animation

Trigger: the filtered list object changes (new `query`, new IPC
results, new dedup/group step).

```
Old list: opacity 1 → 0 over 120ms --ease
New list: opacity 0 → 1 over 120ms --ease (starts at the same time)
```

The lists are stack-rendered with `position: absolute; inset: 0;`
during the cross-fade, then the old list is unmounted. The user sees
one dissolve to the next. No per-row stagger. **A power-user palette
must not stagger** — staggered filters create the "the system can't
keep up" feeling, which is the opposite of what a palette is for.

The cross-fade is a CSS transition on the `<div class="list">`
wrapper; the contents re-render in place when the cross-fade ends.

**Reduced motion:** instant swap (no opacity transition).

### 4.3 Sliding highlight — the FLIP signature

Trigger: `activeIndex` changes (user pressed `↑`/`↓`, or `⌘1`–`⌘9`,
or the filtered list re-rendered).

The `<SlidingHighlight />` component reads `listEl.children` (the
row DOM nodes) and measures the active child's `offsetTop` and
`offsetHeight`. It compares to the previous offset and applies FLIP:

```ts
// On activeIndex change
const prev = { y: prevOffset, h: prevHeight };
const next = { y: nextOffset, h: nextHeight };
const dy = prev.y - next.y;

// Phase 1 — set inverse transform (no visual change)
highlight.style.transition = 'none';
highlight.style.transform = `translateY(${next.y + dy}px)`;
highlight.style.height = `${next.h}px`;

// Phase 2 — force layout, then transition to the real target
highlight.offsetHeight; // force reflow
highlight.style.transition = 'transform 240ms var(--ease), height 200ms var(--ease)';
highlight.style.transform = `translateY(${next.y}px)`;
highlight.style.height = `${next.h}px`;
```

The FLIP runs in a `tick()` after the active row's mount so the
DOM has settled. No per-property stagger. The same technique
NavRail uses (`SCREEN_NAVRAIL.md §5`). The primitive is shared
because the technique is shared.

**Reduced motion:** no FLIP. The highlight's `transition` is set
to `none` by the global media query. The highlight snaps to the
new row.

### 4.4 Input focus-thread — the composer's gesture

Trigger: input `:focus`. A `<Thread />` is rendered above the input
with `stroke-dashoffset: 1 → 0` over `--dur-slow --ease`. Same
gesture as the composer focus state.

```css
.palette-search { position: relative; }
.palette-search .focus-thread {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  pointer-events: none;
}
.palette-search input:focus + .focus-thread line {
  stroke-dashoffset: 0;
  transition: stroke-dashoffset var(--dur-slow) var(--ease);
}
```

The thread persists while the input holds focus. On blur, it
reverses (1 → 0). The transition is the same gesture; the user
never sees two different focus animations.

**Reduced motion:** the global rule sets `transition-duration:
0.01ms`. The thread still appears (it's a 1px line, not an
animation), but doesn't draw.

### 4.5 Close — fade + scale-down

Trigger: `open` transitions `true → false` (Esc, click scrim,
`⌘K` again).

```css
.panel.leaving {
  animation: palette-out 160ms var(--ease) forwards;
}
@keyframes palette-out {
  to {
    opacity: 0;
    transform: scale(0.96);
  }
}
.scrim.leaving {
  opacity: 0;
  transition: opacity 160ms var(--ease);
}
```

After 160ms, the `{#if open}` block unmounts. The leaving animation
uses a Svelte `transition:` directive (e.g. `out:fade|local={...}`)
to capture this in one place.

**Reduced motion:** no scale; just an instant disappearance.

### 4.6 The Panel's hairline (top edge) — the afterbeat thread

The `::before` on the panel is a 1px synapse line with
`transform-origin: left; transform: scaleX(0) → 1` over 520ms
`--dur-slow --ease`, delayed by 60ms after the panel mounts. This
is the same gesture `CommandPalette.svelte:321-337` uses today,
preserved. The hairline *is* the MOAT §3 signature; it doesn't
need to be loud, it just needs to draw.

### 4.7 The `esc`-chip hover — the only per-element motion

The `esc-chip` button has:

```css
.esc-chip {
  transition:
    color var(--dur) var(--ease),
    border-color var(--dur) var(--ease),
    background var(--dur) var(--ease),
    transform var(--dur) var(--ease);
}
.esc-chip:hover {
  color: var(--content);
  border-color: var(--hair-strong);
  background: color-mix(in oklab, var(--content) 6%, transparent);
  transform: translateY(-1px);
}
.esc-chip:active {
  transform: scale(0.96);
}
```

These are the global tactile rules (MOAT §2.7), re-declared here
because the chip is a `<button>` outside `.tactile`. No new
transitions; the same vocabulary lives on every chip in the app.

**Reduced motion:** the global rule suppresses all transition
timings. The chip still changes color on hover (the color is
information, not motion).

### 4.8 What does NOT animate

- The list rows do not enter with a per-row stagger.
- The category labels do not animate.
- The footer hints do not animate.
- The right-hand `↵`/`⌘N` chord does not flash when active (the
  match-flash is deleted per §1 row 8).
- The empty state body has no entrance animation (S4 inherits
  the panel entrance; the body fades in with it).
- The Pulse in the loading state (S5) breathes at 1s — this is
  the only permitted idle loop (a Pulse communicates "alive,
  listening," MOAT §2.5).

---

## 5. Keyboard

The palette is a keyboard-first surface. Every interaction has a chord.

### 5.1 Global chords (palette-driven)

| Chord | Action | Notes |
|---|---|---|
| `⌘K` / `Ctrl+K` | **Open the palette.** | Owned by `Shell.svelte:110` — fires on `keydown` regardless of focus. Pressing again while open closes the palette. |
| `⌘P` | **Open the palette in "search docs" mode.** | Same as `⌘K` but the input is pre-populated with `/docs ` and the docs group is auto-focused. The user types their query after the prefix. (Out of scope for v0.1.0 — placeholder rows for the docs-first mode are visible in S2 but the route is `Chat`'s `/docs` command, not yet wired.) |
| `Esc` | **Close the palette.** | Fires when palette is open and focus is anywhere inside the panel. |
| `Tab` | **Move focus to the chord-hints area in the footer.** | For screen readers — the hints are non-interactive text but they receive focus so they are announced. |
| `Shift+Tab` | **Move focus back to the input.** | Cyclic between input and footer; nothing else inside the panel is focusable while the palette is open (the rows are owned by `aria-activedescendant`). |
| `↑` / `↓` | **Previous / next result.** | Bounded by `[0, results.length - 1]`. Stays inside the list (does not wrap). |
| `Home` / `End` | **First / last result.** | Skips directly. |
| `PageUp` / `PageDown` | **Scroll the list by `listEl.clientHeight - 80px` (one row short of a full page).** | Doesn't move the active row unless the visible window stops being the active row. |
| `Enter` | **Run the active command.** | Same as clicking the active row. If the row is a navigation command, navigates. If it's a skill, sends `/run <skill>` to chat. If it's a conversation, navigates to `#/chat?c=<id>`. If it's a doc, opens `/docs/<slug>`. If it's an action, runs the action. |
| `⌘1`–`⌘9` | **Quick-select the Nth visible result.** | Index is across all groups, in their fixed order. `⌘1` is the first result regardless of group. The first 9 results animate to `⌘1`–`⌘9` chords in their right-hand cluster; results 10+ show `↵`. |
| `⌘Enter` | **Run the active command without closing the palette** (rare; reserved for the Stop action so the user can stay in place after halting). | A single escape hatch — kept narrow. |
| `?` | **Open the **Shortcuts** sheet.** | While the palette is closed, `?` opens the Shortcuts sheet (MOAT §2.10) with the `⌘K` palette listed. While the palette is open, `?` types a literal `?` into the input. |

### 5.2 Input focus discipline

When the palette opens, the input auto-focuses after an 80ms delay
(`void tick().then(() => inputEl?.focus())`). The 80ms matches the
entrance — the focus-thread begins drawing just as the panel settles.
The user can start typing within ~280ms total from the `⌘K` press.

### 5.3 `aria-activedescendant` pattern

The palette uses `role="combobox"` on the input and `role="listbox"`
on the list. Rows are `role="option"`. The input carries
`aria-activedescendant={activeRowId}` pointing at the active row's `id`.
This is the canonical pattern for combobox-with-listbox in WAI-ARIA 1.2.

**Focus stays in the input.** Arrow keys move `aria-activedescendant`
without moving focus. Screen readers announce the active row's label
on every arrow press. The user's type position in the input is
preserved across arrow presses (unlike moving focus into rows).

### 5.4 Reduced motion

All keybindings are unchanged. (MOAT §2.3 — components never branch
on motion prefs; the global CSS handles it.)

---

## 6. Components Used

The palette is composed of these primitives. The full prop contracts
below belong in each component's source file — this spec is the
short-form.

### 6.1 `<SearchInput />` — to be extracted per MOAT §5.2

```ts
let {
  value: string,
  placeholder: string,
  glyph: string,                 // 'search' | 'command' | 'route'
  loading: boolean,
  loadingGlyph: boolean,         // when true, render <Pulse> alongside value
  footerSlot: Snippet | null,    // optional inline right-cluster
  oninput: (v: string) => void,
  onkeydown: (e: KeyboardEvent) => void,
  class: string = '',
} = $props();
```

Owns:

- The 14px mono input field with italic display placeholder.
- The left-side `<Glyph>` (search/command/route icon).
- The right-side slot — the `esc` chip OR the loading Pulse OR
  nothing — depending on `loading` and the caller's `footerSlot`.
- The `<Thread />` focus-thread above the input.

Does **not** own:

- The filter/ranking logic (owned by the palette).
- The result list (owned by the palette).

### 6.2 `<SlidingHighlight />` — new primitive

```ts
let {
  container: HTMLElement,    // the scrollable parent (listEl)
  activeIndex: number,
  children: HTMLElement[],   // the row elements in render order
  reducedMotion: boolean,    // passed in from the global media query
  class: string = '',
} = $props();
```

Owns:

- The absolute-positioned highlight div.
- The FLIP measurement and inverse-transform dance on
  `activeIndex` change.
- The transition toggle for `prefers-reduced-motion`.

Does **not** own:

- The row children (owned by the caller).
- The scroll-into-view (caller may call `activeChild.scrollIntoView`
  separately; this primitive does NOT auto-scroll because the
  palette's `scrollIntoView({block: 'nearest'})` may differ from
  the NavRail's).

### 6.3 `<CategoryGroup />` — new primitive

```ts
let {
  label: string,             // 'Routes' | 'Shell' | ...
  count: number | null,      // optional mono count after the label
  children: Snippet,
  class: string = '',
} = $props();
```

Owns:

- The mono caps 11px label (`<h3 class="cg-label">`).
- The 8px vertical gap to its rows.
- The 12px bottom gap to the next group.

Does **not** own:

- The row children (passed via `children` snippet).
- The highlight (owned by `<SlidingHighlight />`).

### 6.4 `<Thread />` — already exists, reused

Props unchanged: `{ orientation: 'h' | 'v', draw: boolean, glow:
boolean, class?: string }`. Used in:

- The panel's top-edge hairline (entrance afterbeat).
- The focus-thread on the input.
- The category-group dividers (optional — see §6.7).

### 6.5 `<Pulse />` — already exists, reused

Props unchanged: `{ phase: 'thinking' | 'acting' | 'ok' | ..., size:
number, class?: string }`. Used in the loading state (S5) only —
8px `thinking` next to the SEARCHING label. The pulse is the only
idle loop permitted on this surface.

### 6.6 `<Glyph />` — already exists, reused

The palette consumes these icons from `icons.ts`:

| Where | Glyph name | Notes |
|---|---|---|
| Left of input | `search` | 18px |
| Row glyphs — Routes | `route`, `hub`, `skills`, `sync`, `audit`, `replay`, `channels`, `delegation`, `settings`, `about` | 18px |
| Row glyphs — Shell | `sun` (theme when going light), `moon` (theme when going dark), `bolt` (summon), `stop` (stop) | 18px, theme icon flips with `data-mode` (current `CommandPalette.svelte:46-54` logic). |
| Row glyphs — Skills | `bolt` (installed) / `arrow-down-right` (Hub browsable) | 18px |
| Row glyphs — Conversations | `message` | 18px |
| Row glyphs — Docs | `book` | 18px |
| Empty S4 body | `eye` (decorative pollen ring, locked) | 32px, 30% opacity, behind the headline |
| Right of input — loading | `<Pulse phase="thinking">` | 8px |

If any row's icon doesn't exist in `icons.ts`, add it (per MOAT §4 #2),
do not fall back to emoji or unicode.

### 6.7 `<Button />` — already exists, reused

Props unchanged: `{ variant, size, icon, loading, children, onclick }`.
Used in:

- The `esc` chip (already a `<button>` — variant `ghost`, size `sm`,
  label `esc`).
- The S4 "Try again" pill in the empty state (variant `primary`,
  pollen-outline, label `Try again`).
- The `Clear recent` ghost button at the bottom of the Recent group.
- The S6 `[Try again]` in `<ErrorState />`.

### 6.8 `<Tooltip />` — to be created per MOAT §2.9

Props (per MOAT §2.9): `{ label, placement, delay, exit, children }`.
Used only on the `esc` chip — `label="Close (esc)"`. (Other rows have
their labels right next to them; tooltips would be redundant.)

### 6.9 `<ErrorState />` — to be extracted per MOAT §1.2 / §2.6

Props (per Skills spec §7.9): `{ head, cause, reason, onretry,
retryLabel?, onsettings?, class? }`. Used in S6.

### 6.10 `<EmptyState />` — to be extracted per MOAT §2.4

Props (per Skills spec §7.10): `{ what, why, action, sample, class? }`.
Used in S4. The `sample` snippet is the SUGGESTED ROUTES block under
the teach copy.

---

## 7. Data Fetched

The palette has exactly **one** IPC call. Everything else is local
filter/dedup/rank.

### 7.1 `ipc.commandPalette.search(query, limit = 20)`

RPC: `commandPalette.search`. Returns `CommandPaletteSearchResult
{ routes: NavResult[]; shell: NavResult[]; skills: SkillResult[];
conversations: ConvResult[]; docs: DocResult[]; taken_ms: number; }`.
Each `NavResult` is `{ id, label, icon, route?, chord?, rank: number }`,
where `rank` is the source's confidence score (0–1).

```ts
commandPaletteSearch(
  query: string,
  limit = 20
): Promise<CommandPaletteSearchResult>
```

**Cache policy:**

- Debounce input changes by **150ms**. After the user stops typing,
  one IPC fires for the final query.
- In-flight cancellation: if the user types again before the
  previous request resolves, the previous `AbortController` is
  cancelled. (Implementation: a per-component AbortController
  whose signal is passed to `ipc.call`.) Stale results are
  discarded.
- On result: results are normalized into a flat ordered list of
  commands. The `commands` array structure is rebuilt via
  `$derived.by(...)` to trigger the cross-fade (§4.2).

**Error policy:**

- On rejection, an `<ErrorState />` renders (S6). The user can
  click `[Try again]` which re-invokes the IPC with the same
  query.
- On resolve-with-empty, an `<EmptyState />` renders (S4).

**When `query === ''`:**

- No IPC is called. The S2 recent-commands list is used directly.
- Recent commands are read from `localStorage[condura:palette-recent]`
  on mount, parsed from a JSON array of `{ id, kind, runAt }` rows.
  Cap is 5; FIFO eviction on overflow.

**What is NOT called:**

- `ipc.skills.list`, `ipc.skills.get`, `ipc.hub.search`,
  `ipc.conversations.list`, `ipc.docs.search` — these are the
  sources the backend merges from inside `commandPalette.search`.
  The palette does not call them individually. The daemonside
  aggregation is the source of truth.

### 7.2 Local writes

The palette writes **only** to `localStorage`:

- `condura:palette-recent` (JSON array of recent command ids + runAt).
  Written when a command runs (on Enter or click), pushed to the
  front of the list, deduped, capped at 5.

The palette does NOT persist the user's query (cleared on close)
or the active index (reset on open).

### 7.3 Offline / daemon-down behavior

`ipc.commandPalette.search` rejects with the same envelope as other
RPCs. The palette catches the rejection and renders S6. The user
can still navigate to routes from the Routes group IF those rows
came from a prior successful search cached in `sessionStorage`
(cap 64KB, key `condura:palette-cache`, expires on close).

The fallback cache is best-effort — the user is told (via the
"Showing last 6 commands" note in S5) when they're seeing cached
results. No silent data show.

---

## 8. Design Decisions (MOAT Compliance)

These are the load-bearing calls — where this spec disagrees with the
current `CommandPalette.svelte`, and where it earns the MOAT bar.

### 8.1 Five sources, one query

**The problem (per brief §6).** The current `CommandPalette.svelte`
filters a hardcoded 13-command list. It cannot search the user's
skills, conversations, or docs.

**What this spec does.** `commandPalette.search` aggregates five
sources on the daemon side. Routes and Shell (the local verbs)
return instantly from the daemon's in-memory model. Skills return
from `skills.list` and `hub.search`. Conversations return from
`conversations.list` indexed by FTS5 + embeddings. Docs return from
the local docs corpus indexed by FTS5. The frontend receives a flat
array of ranked candidates. The 150ms debounce matches Raycast's
characteristic "no spinning, instant as you type" feel — users
perceive the search as instantaneous.

### 8.2 Restraint — delete the match-flash

**The problem (per MOAT §1, restraint test).** The current
`CommandPalette.svelte:494-502` replays a 180ms match-flash on
the active row's `↩` glyph every time `activeIndex` changes. Pressing
arrow keys 10 times fires 10 flashes. The flash competes with the
sliding highlight (also firing per-arrow-press) and reads as
**the system trying too hard**.

**What this spec does.** The match-flash is gone. The affordance is
communicated by (a) the rounded focus halo on the active row
(per MOAT §2.1), (b) the chord glyph `⌘N` for the first 9 results
(this makes "Enter to run" obvious — the chord is right there), (c)
the footer hint `↵ run`.

### 8.3 The signature is shared — extract `<SlidingHighlight />`

**The problem (per MOAT §3).** The NavRail and the CommandPalette
each implement their own sliding-highlight mechanism. Both use
FLIP, both are 200–240ms, both are `--ease`. Two implementations
of the same gesture. Future surfaces (Settings inner rail, Onboarding
"you are here" cursor) will want the same thing.

**What this spec does.** `<SlidingHighlight />` is extracted. Its
props are container, active index, children, and reduced-motion
flag. The NavRail uses it. The palette uses it. Future surfaces
import it. The signature is one element, one component, one
implementation.

### 8.4 The input is the composer — share the focus-thread

**The problem (per MOAT §5.2).** The composer focus state draws a
synapse thread under the textarea on focus. It's one of the great
micro-moments. The palette's input is a sibling (it's a text input
in the same app), but currently uses a different gesture
(`box-shadow: inset 0 -1px 0 var(--synapse)` center-out keyframe).

**What this spec does.** The palette's input uses the same `<Thread
/>` focus-thread. The user learns one gesture for "I am writing
here" and sees it on the composer, on the palette input, on the
Settings search field, on the Hub search field. **One gesture, five
inputs.** The thread-draw is the project-wide "writing here" signal.

### 8.5 Window disclosure — never widen the panel

**The problem.** The brief names a max-width of 640px. The current
panel is 560px. Both are within "is the user on a 13-inch
laptop at 1280px wide?" — a comfortable single-column width.
A wider panel would invite row-wrap and become a different beast.

**What this spec does.** Width is 640px on desktop. Below 720px the
panel is 92vw (the same min-width as the current implementation).
The panel never widens beyond 640px because widening would break
the "this is a dialog, not a route" promise (a route can be wider;
a dialog can't).

### 8.6 Visibility of state — RECENT is a real group, not a guess

**The problem (per MOAT §1, restraint test).** A power-user
palette must be predictable. If the panel shows **different**
results on first open than on second open, the user thinks the
palette is broken. The current `CommandPalette.svelte` always shows
the same 13 commands. There's no surprise — there's also no
personalization.

**What this spec does.** The S2 state shows the **5 most-recently-run
commands** in a `RECENT` group. The Recent list is persisted (so
returning users see their own patterns). The Recent group sits
above the source groups. If recent is empty (first-ever `⌘K`),
the S2 fallback copy (three-line teach per MOAT §2.4) renders
instead. The user is never surprised — recent items are
**labeled** and **above** the source groups, so the user knows
what they're looking at.

### 8.7 The error is honest — never "no results" without reason

**The problem (per MOAT §2.6).** The current empty-state copy
is "Nothing matches this search." This is what shows for both
zero-results AND for daemon-down. They are different conditions.

**What this spec does.** Three distinct states: S3 (results),
S4 (zero-results — three-line teach), S6 (daemon error —
`<ErrorState />` with cause/reason). The user knows whether the
search "found nothing" or "couldn't run." S6 even distinguishes
between "daemon restarted" and "index rebuilding" via the cause
line.

### 8.8 The scrim earns glassmorphism — cards never do

**The problem (per MOAT §4 #3).** Glassmorphism is forbidden on
cards, lists, and inline surfaces. It is permitted on dialog
scrims (because the dialog scrim must dim the route behind it).

**What this spec does.** `backdrop-filter: blur(8px) saturate(0.9)`
on the scrim is unchanged. The panel itself has no backdrop-filter
(it's a solid `--paper` background — no glass on glass). The
rows inside the panel have no backdrop-filter. The scrim is the
only surface with the blur. This passes MOAT §4 #3.

### 8.9 No new thread uses — the moment of "command selected"

**The problem (per MOAT §3 — commit to one element).** The
command-palette-running-a-command moment could draw a thread
across the row. It would be cute. It would also be a fifth
thread use case in the project, which dilutes the signature.

**What this spec does.** The palette does **not** draw a thread
on a selected command. The visual feedback is: (a) the panel
closes (S1 transition), (b) the route enters (existing route-enter
animation from `Shell.svelte` / `Transition.svelte`), (c) the
underlying surface (chat, skills, etc.) shows its own arrival
moment. The thread is the **completion** gesture, not the
**selection** gesture. Selecting a route completes when the new
route finishes entering. Selecting an action completes when the
action's surface (e.g. a quick-prompt overlay, a theme flip)
finishes its own animation. The palette's role is to dispatch;
the destination's role is to celebrate.

### 8.10 Keyboard chords match the global table

**The problem (per MOAT §2.10).** The current implementation has
no `⌘1`–`⌘9`, no `⌘P`, no `⌘,` (Settings, owned by Shell), no `?`
(open Shortcuts). The keyboard surface is incomplete.

**What this spec does.** `⌘1`–`⌘9` for quick-select. `⌘P` opens
docs-mode. `⌘K` opens. `Esc` closes. `?` opens the Shortcuts sheet
when palette is closed (when palette is open, `?` types a literal
`?`). Every palette chord is listed in the global `Shortcuts.svelte`
overlay. There are no one-off chords.

### 8.11 The chord glyph is what makes power-users fast

**The problem.** A power user picks `Summon Quick Prompt` on
average 12ms faster if they can `⌘1` instead of two arrows + Enter.
Without the chord glyph, the only way to know which chord belongs
to which row is to read the footer.

**What this spec does.** The first 9 visible rows render `⌘N` in
their right-hand chord slot. The chord is **assigned to the row**
when the filtered list settles (it's a function of the rendered
list, not stored state). The chord re-distributes when the user
types — `⌘3` is always the third result regardless of which group
it lives in.

### 8.12 The palette works during the ritual (TEARDOWN §4 synthesis)

**The problem (per TEARDOWN §4, Linear).** Linear's `⌘K` opens
**during** onboarding. There's no situation where the user is
trapped.

**What this spec does.** `Shell.svelte:251` always mounts the
palette. When the ritual is active, the palette overlays the
ritual surface. The user can `⌘K` to "jump to settings" mid-ritual
without finishing the ritual. This matches Linear's escape-valve
pattern. The ritual still has its own `Begin →` etc.; the palette
is a parallel front door.

---

## 9. Accessibility Contract

| Surface | What | ARIA |
|---|---|---|
| Scrim | Click target for "click outside to close." | `role="presentation"` (the scrim is decorative; the panel is the dialog). |
| Panel | The dialog itself. | `role="dialog" aria-modal="true" aria-label="Command palette" tabindex="-1"` |
| Input | The combobox. | `role="combobox" aria-expanded="true" aria-controls="cmd-list" aria-autocomplete="list" aria-activedescendant={activeRowId} aria-label="Search commands"` |
| List | The listbox. | `role="listbox" aria-label="Commands"` |
| Row | An option. | `role="option" aria-selected={isActive} id="cmd-{i}"` |
| Category group | A labelled group within the list. | `role="group" aria-labelledby={cgLabelId}` where `cgLabelId` references the `cg-label` heading. |
| Highlight | Decorative. | `aria-hidden="true"` |
| Empty / error body | An alert region when a new condition applies. | `role="region" aria-live="polite"` for S4 and S6 transitions. S2 and S3 don't announce (they're not state changes). |

Focus management:

- On open: input auto-focuses after 80ms.
- On close: focus returns to the element that triggered the
  palette (saved at open time — typically the last-focused
  surface in the shell). The palette does NOT keep focus on the
  input after close.
- Tab: cycles between input and footer chord-hints area only.
  Rows are selected via `aria-activedescendant` — not focus.
- Esc: closes; the focus return path runs.

Screen reader announcements:

- The listbox's `aria-activedescendant` updates on every
  `activeIndex` change so the screen reader announces the new
  active row's label + group.
- The footer hints are reachable via Tab; they announce as a
  group ("↑↓ navigate, ↵ run, ⌘1–9 quick-select, esc close").
- State changes to S4 / S6 are announced via `aria-live`.

Keyboard-only flow:

- A user can open the palette, type a query, press `⌘1`–`⌘9` to
  pick a result, and run it without ever using the mouse or
  Tab. The whole surface is keyboard-first.

Reduced motion:

- All transitions are suppressed. The panel appears instantly.
  The highlight jumps to the new row. The cross-fade on results
  is an instant swap. Focus-thread draws instantly. The
  per-component `prefers-reduced-motion` block is replaced by the
  global rule per MOAT §2.3.

---

## 10. Implementation Notes for Phase 4

These are the practical hooks the implementation needs to land
this spec.

### 10.1 The order of work

1. Extract `<SlidingHighlight />` first (shared with NavRail).
   Tests pin the FLIP measurement + the `reducedMotion` flag.
2. Add `ipc.commandPalette.search` to the daemon RPCs.
   Backend: aggregate from `routes.list`, `skills.list`,
   `hub.search`, `conversations.list`, `docs.search`. Return
   flat ranked candidates.
3. Replace `CommandPalette.svelte` line-by-line from the top of
   the file: state → input → list → footer → styles. The diff
   is one atomic commit.
4. Add the recent-commands `localStorage` read/write.
5. Wire `⌘1`–`⌘9` and `⌘P`. Update `Shortcuts.svelte` to list
   the palette's chords.
6. Add `<Thread />` focus-thread on the input. Delete the
   `box-shadow` keyframe.

### 10.2 Files this spec creates or modifies

| File | What |
|---|---|
| `app/web/frontend/src/lib/condura/CommandPalette.svelte` | The main rewrite (this spec). |
| `app/web/frontend/src/lib/condura/SlidingHighlight.svelte` | **New** — extracted FLIP primitive. |
| `app/web/frontend/src/lib/condura/CategoryGroup.svelte` | **New** — labelled group with mono caps heading. |
| `app/web/frontend/src/lib/condura/SearchInput.svelte` | **New** — extracted per MOAT §5.2. |
| `app/web/frontend/src/lib/condura/NavRail.svelte` | Migrate its inline highlight to `<SlidingHighlight />` (already-extracted). |
| `app/web/frontend/src/lib/ipc/types.ts` | Add `CommandPaletteSearchResult`, `NavResult`, `SkillResult`, `ConvResult`, `DocResult`. |
| `app/web/frontend/src/lib/ipc/client.ts` | Add `commandPaletteSearch(query, limit)`. |
| `app/web/frontend/src/lib/condura/Shortcuts.svelte` | List the palette's chords: `⌘K`, `⌘P`, `⌘1`–`⌘9`, `Esc`, `?`. |
| `app/web/frontend/src/lib/condura/Shell.svelte` | Wire `commandPalette.search` to the global `⌘K` listener (line 110). Verify the palette is mounted during the ritual. |
| `internal/daemon/*.go` | Add `commandPalette.search` handler. |

### 10.3 Hot paths to test

- Open with `query === ''`: S2 → RECENT group.
- Open with `query === 'audit'`: S3 → 1 group (Routes) with 1
  active row.
- Open with `query === 'summ'`: S3 → 2 groups (Routes `audit`?
  No — `summon` Shell verb only).
- Open with `query === 'asdf'`: S4 → empty state with the
  SUGGESTED ROUTES sample.
- Open with daemon down: S6 → `<ErrorState />` with `[Try
  again]`.
- `⌘K` to open, type `theme`, `⌘1` to quick-pick, Enter: theme
  flips, palette closes.
- `⌘K` to open, type `audit`, ↓ once, Enter: navigates to
  `#/audit`.
- `prefers-reduced-motion: reduce`: entrance is instant;
  highlight jumps; cross-fade is instant.

---

## 11. Test Plan

The palette has dense state and motion logic; the tests below are
the floor for "this spec is implemented."

### 11.1 Unit — `SlidingHighlight.svelte`

- `activeIndex=0` + three children → highlight at `children[0].offsetTop`.
- `activeIndex` changes from 0 → 2 → highlight animates to
  `children[2].offsetTop` over 240ms `--ease`.
- `reducedMotion=true` → no transition.
- The container scroll position is preserved across FLIP (the
  highlight moves; the list doesn't jump).

### 11.2 Unit — `CategoryGroup.svelte`

- Renders the label as `<h3>` with mono caps 11px, faint.
- The `count` renders after the label, separated by an inline gap.
- Children pass through.

### 11.3 Unit — `SearchInput.svelte`

- Focus auto-applies the focus-thread (visible via the
  `.focus-thread` SVG having `stroke-dashoffset: 0`).
- Loading state shows `<Pulse>` in the right cluster.
- Footer slot renders without disrupting the input.

### 11.4 Integration — `CommandPalette.svelte`

- `open=false` renders nothing (DOM empty).
- `open=true` renders the panel.
- `query=''` → RECENT group (S2).
- `query='audit'` → S3 with 1 group (Routes).
- `query='asdf'` → S4 with SUGGESTED ROUTES.
- IPC error → S6 with `<ErrorState />`.
- `Esc` closes the panel.
- `⌘1`–`⌘9` selects the Nth result.
- `Enter` runs the active row.
- Click outside (scrim) closes the panel.
- Click on a row runs that command and closes the panel.

### 11.5 a11y

- `aria-activedescendant` updates on every `activeIndex` change.
- `Esc` returns focus to the element that triggered open.
- `Tab` cycles between input and footer.
- Screen reader announces S4 and S6 transitions.

### 11.6 Reduced motion

- Entrance: no scale, no transition.
- FLIP: highlight jumps instantly to the new row.
- Cross-fade: instant swap.
- Focus-thread: still rendered, draws instantly (the visual end-
  state is unchanged).

---

## Closing note for Phase 4

The palette is the **power-user shortcut**, not the meta-meta-UI. It
must do one thing — answer the question "what do I want to do?" in
under a second. The five sources (Routes, Shell, Skills,
Conversations, Docs) come together because they are the five
routable destinations in Condura. The sliding highlight is the
project's signature because the user sees it everywhere (NavRail,
palette, future Settings inner rail). The focus-thread on the input
binds the palette to the composer — same gesture, two surfaces.

The current `CommandPalette.svelte` is competent — it works, it
ships, it doesn't crash. But it has nothing the user can't get
elsewhere. This spec gives it:

1. **Five sources** instead of two (the user's whole command
   surface, one query away).
2. **A signature FLIP shared with the NavRail** instead of a
   one-off imperative CSS transition.
3. **A focus-thread shared with the composer** instead of an
   ad-hoc input keyframe.
4. **Chord glyphs that make power users faster** (`⌘1`–`⌘9`).
5. **Three distinct states** (results, empty, error) instead of
   one overloaded "Nothing matches."
6. **Restraint** — the match-flash is gone; the per-row stagger
   is gone; the row-rotation `scale(1.04)` is gone.

Ship the `<SlidingHighlight />` primitive early; both the NavRail
and the palette use it. Ship the `<SearchInput />` extraction
because Settings, Hub, and Channels will want the same focus-thread.
Ship `<CategoryGroup />` because the same mono-caps eyebrow will
show up on the Settings inner rail before this spec ships.

If during Phase 4 you find a place where the spec is silent and the
default would be wrong, the answer is to ask before inventing. The
MOAT bar is "premium-quality." Anything that fails it is competent-
but-generic, and competent-but-generic is how a year gets lost.
