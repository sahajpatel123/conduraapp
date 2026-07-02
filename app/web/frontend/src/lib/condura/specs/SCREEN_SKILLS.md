# SCREEN — Skills · `Skills.svelte` · `#/skills`

> **Status:** Phase-2 architecture spec. Phase 4 implements against this.
> **Contract:** `MOAT.md` (premium bar), `APPFLOW.md` §4.2 (current Skills),
> `TEARDOWN.md` §7 (synthesis), `icons.ts` + `condura.css` (the actual
> tokens available). Two of the original five "north-star" documents
> (`DIRECTION.md`, `DESIGNLANG.md`) were named in the brief but are not on
> disk at this commit, so this spec leans on `MOAT.md` for the design
> grammar and on the actual token table from `condura.css` for everything
> visual.
>
> The current `Skills.svelte` (376 lines) is a single-column grid with a
> sheet detail panel, `rotateX(2deg)` hover, an inline copy-pasted
> `err-state`, and an overlay with `backdrop-filter: blur(4px)`. Three
> things in that file are called out by name in `MOAT.md §1.2`,
> `§1.5`, and `§4.3` as things to delete. This spec deletes all three and
> replaces them with the primitives the design system has now grown.

---

## Table of Contents

1. [What this route is and isn't](#1-what-this-route-is-and-isnt)
2. [Inheritance — what the spec inherits from MOAT/APPFLOW/TEARDOWN](#2-inheritance)
3. [Layout & Content](#3-layout--content)
4. [State Matrix — six states, full copy](#4-state-matrix)
5. [Motion Choreography](#5-motion-choreography)
6. [Keyboard](#6-keyboard)
7. [Components Used — boundaries and props](#7-components-used)
8. [Data Fetched — IPC contract](#8-data-fetched)
9. [Design Decisions — which MOAT rules this passes](#9-design-decisions)
10. [What this spec deletes from the current `Skills.svelte`](#10-what-this-spec-deletes)

---

## 1. What this route is and isn't

**Is.** A unified index of every skill the user can see — installed
locally, available in the Hub, user-authored, agent-authored — in one
filterable grid. A skill is **a procedure the agent can re-run.** It is a
file on disk in `~/.synaptic/skills/<id>/SKILL.md` (or `.toml`,
agentskills.io format). The Hub is just a downloadable source for the
same shape.

**Is not.** Not a "library" (no borrow/return mental model). Not a
package manager (no dependency-resolution screen). Not a marketplace
(no reviews, no ratings, no comments — `MOAT.md §4 #1–#9` forbid the
visual language of one). Not a tutorial.

**Mental model the user carries away:** *Skills is the shelf behind the
glass. Some are already behind the glass; some are on the other side
waiting to be brought in.*

**One sentence for the eyebrow:** *— Procedures the agent can re-run.*

**One sentence for the title:** *Skills, behind the glass.*

**Why "behind the glass":** the route unifies "what I have" with "what
I can install" in the same visual plane, separated by a single filter
(Source → Behind the glass / Hub / All). No metaphor beat harder than
"the shelf is right here, one switch tells you which side." MOAT §1.5
kills the 3D bookshelf; this is the cheapest way to keep the browse-
both-sides idea without the perspective tax.

---

## 2. Inheritance

The spec assumes the following are already in place; this spec does
**not** redefine them.

| From | What this spec uses as-is |
|---|---|
| `MOAT.md §2.3` | `prefers-reduced-motion` is respected via one global rule in `condura.css`. The component declares no media-query blocks. |
| `MOAT.md §2.4` | Empty states are three lines: **what** / **why empty** / **next action.** The empty state in this spec follows that shape exactly. |
| `MOAT.md §2.6` | One `ErrorState` component owns all error rendering. The route uses it. No inline err blocks. |
| `MOAT.md §2.7` | `.tactile` global class owns press/hover transition timing. The component declares no per-card `transition:` lists. |
| `MOAT.md §2.8` | Detail panel is a `.c-sheet`. The filter rail popover is a `.c-popover`. Loading state is a `Thread` draw, not a `Pulse`. |
| `MOAT.md §2.9` | Icon buttons in the card header get `<Tooltip label>` — no `title=` attributes that misbehave. |
| `MOAT.md §3` | "Finished" states are communicated by a `Thread` drawing in (left-to-right). The install-settled state, the update-installed state, and the post-error-recovery all draw a thread. |
| `MOAT.md §4 #1` | No gradient text. No gradient anywhere on this route. |
| `MOAT.md §4 #2` | All icons go through `<Glyph name="…" />`. No emoji. No Unicode-as-icon. |
| `MOAT.md §4 #3` | No `backdrop-filter: blur()` on this route. The detail panel is a `.c-sheet` — paper-on-paper, hairline left edge, `--shadow-float` if it needs elevation. The modal-overlay blur used today is deleted. |
| `MOAT.md §4 #4` | Status colors are `--ok` / `--warn` / `--danger` / `--info`. The trust badge uses these — no purple, no teal, no new accent. |
| `MOAT.md §4 #5` | No "Welcome to the future" copy. The empty state and the title read like a museum label, not a landing page. |
| `MOAT.md §4 #6` | No "Awesome!", no "Installed!" celebration. The card settles to a hairline check-mark and the body copy reads "Installed". |
| `MOAT.md §4 #7` | No spinner. Loading is a `Thread` drawing left-to-right (`drawthread`, `var(--dur-slow)`, `var(--ease)`), exactly once per surface. |
| `MOAT.md §4 #8` | Focus halos are rounded — `--shadow-focus` (the 4px pollen halo) or, for ≥8px radii elements, a 5px halo + 2px synapse ring. Square halos are forbidden. |
| `MOAT.md §4 #9` | One elevation token per surface. Cards use `--shadow-paper` at rest and `--shadow-card` at hover. No stacked shadows. |
| `MOAT.md §4 #10` | Every animation in this spec answers: *what is this communicating?* Entrance communicates "data ready"; hover communicates "this takes focus"; flip communicates "state change to installing"; thread-draw communicates "completed." |
| `APPFLOW.md §7` | The state inventory table is augmented with six new rows for the redesigned Skills. Old `Skills · loading / empty / error / detail` rows are **replaced**, not deleted. |
| `TEARDOWN.md §7` | Empty state is opinionated-populated — it shows the user what the shelf *will* look like with three sample cards. The empty copy teaches before asking. |

---

## 3. Layout & Content

### 3.1 Page-level structure

The Skills route renders inside the shell's main surface, right of the
`NavRail` (which is 232px wide per `APPFLOW.md §3.1`).

```
┌─────────────── Skills route (max-width 1280px, padded) ────────────────┐
│                                                                       │
│  ── Procedures the agent can re-run.                                  │  eyebrow (mono, 11px, caps, --content-faint)
│                                                                       │
│  Skills, behind the glass.                                            │  title (display, clamp 28→40px, --content)
│  One shelf — what's already yours, and what's on the other            │  sub (sans 16, --content-soft, max 56ch)
│  side waiting to come in.                                             │
│                                                                       │
│  ┌───────────────── Toolbar row (sticky, --surface, --hair bottom) ─┐ │
│  │ Search ⌕ ⌕  (1px halo on focus; / to focus, Esc to clear+blur) │ │ search input fills width-280
│  │                                          Sort · 6 cards ·   [⛁]  │ │ right cluster: sort label, count, refresh
│  └───────────────────────────────────────────────────────────────────┘ │
│                                                                       │
│  ┌────────────┐  ┌──────────────────────────────┐  ┌──────────────┐  │
│  │            │  │                              │  │              │  │
│  │ Filter     │  │   Grid                       │  │   Detail     │  │
│  │ rail       │  │   (1-3 cols, auto-fill)      │  │   panel      │  │
│  │            │  │                              │  │              │  │
│  │ 240px      │  │   flexes with viewport       │  │   360px      │  │
│  │  sticky    │  │                              │  │   appears    │  │
│  │  on        │  │                              │  │   when a     │  │
│  │  scroll    │  │                              │  │   card is    │  │
│  │            │  │                              │  │   selected   │  │
│  │            │  │                              │  │              │  │
│  └────────────┘  └──────────────────────────────┘  └──────────────┘  │
│                                                                       │
│  ── Hub chrome / footer band ──                                       │  hairline + small mono link row
└───────────────────────────────────────────────────────────────────────┘
```

The page total max-width is 1280px (wider than today's 980px — the new
grid earns the room). Above 1280px viewport, content stays centered.
Below 960px viewport, the filter rail collapses to a `[⛁]` button that
opens a `.c-popover` sheet. Below 720px viewport, the detail panel
becomes a full-screen `.c-sheet`. Below 480px, the grid goes single-
column and `min-width: 0` lets descriptions truncate.

### 3.2 Region A — Filter rail (240px, sticky on scroll)

A `.c-paper` vertical column, 16px from the grid. The rail is its own
landmark (`<nav aria-label="Skill filters">`). It contains, top-to-bottom:

#### A.1 "View" group

- **— Behind the glass · Installed only.** The count in mono: `23`.
  Default selection. This is what's local-first.
- **— Other side · Hub browsable.** The count in mono: `184`. Things the
  Hub ships.
- **— All.** `207`.

Two rules (`<Thread />` hairlines) separate this group from the next.

#### A.2 "Trust" group

A stacked checkbox column. Each row is `[☐] Label (count)`. The three
trust levels are **honest**, never collapsed:

| Level | What it means | Color |
|---|---|---|
| **Official** | Curated by Condura maintainers; signed against the org key. | `var(--synapse)` filled |
| **Community** | Community-published with checksum; published_at > 30d; ≥10 downloads | `var(--synapse)` outlined (hairline + dot) |
| **Experimental** | New, unreviewed, or published by an author without a community history. | `var(--pollen)` filled |

The legend never blurs these. A card is never "Verified." It is exactly
one of the three above. **No card is hidden behind "Verified ✓" if its
trust level is `experimental`** — the dot is right there on the card,
same position for every card. (MOAT §4 #5 — no fake enthusiasm.)

#### A.3 "Source" group

A second stacked column of checkboxes. Each row is `[☐] Source (count)`.

- **Local file** — installed from a path on this machine (you `import`
  it).
- **Hub curated** — published on hub.condura.app; safety-scanned.
- **Hub community** — published by a community author.
- **Auto-created** — the agent wrote this from a past task.

#### A.4 "Category" group

Tags from the union of current results, with `+N more` overflow.
Each is a chip. Click a chip → same as a checkbox.

#### A.5 "Author" group

A small monospaced list (`sans` 13, `--content-soft`). "Condura" pinned
first if present. Then alphabetical authors with mono counts.

At the bottom of the rail: a single line, monospaced, 11px, faint:
**Filters update as you check.** A `Reset filters` link in mono-pollen
appears below it whenever any filter is active.

### 3.3 Region B — Grid

#### B.1 Container

`display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));`
**No `perspective`. No `transform-style: preserve-3d`. No `rotateX`.**
This is the explicit `MOAT.md §1.5` deletion; the grid is a clean flat
plane.

`gap: var(--space-4)` (16px). `align-items: stretch`. The grid sits
inside a `.grid-wrap` that has `padding: 0 var(--space-4) var(--space-9)`.

The grid renders inside a `{#key gridLoadId}` block so that a manual
refresh re-mounts the cards and the entrance stagger fires again.

#### B.2 Card anatomy (per card)

```
┌────────────────── 280px ──────────────────┐
│                                            │
│  ⌁                                  [⋯]   │  row 1: trust badge · menu · 20px tall
│  Title goes here in two lines max          │  row 2: name (display 18)
│                                            │
│  Description here in three short lines.    │  row 3: desc (italic display 13, 3-line clamp)
│  It describes what the procedure does     │
│  in plain language — no jargon — and       │
│  the third line cuts at "…"                 │
│                                            │
│  ─ v0.2.1 · Condura · 184 dl · community   │  row 4: meta hairline row (mono 10, faint)
│                                            │
│  [ Install  →  ]                  Updated · 2d │  row 5: action row
│                                            │
└───────────────────────────────────────────┘
                                              ↕ 184px (min)
```

Each row is exactly:

| Row | Component | Notes |
|---|---|---|
| 1 | Trust badge (left) + card menu (right) | Trust is a `Glyph` + label: `official` → dot-active filled synapse + "Official"; `community` → dot-active hairline outline synapse + "Community"; `experimental` → dot-active filled pollen + "Experimental". Card menu is a `<button aria-label="Card menu">` containing `<Glyph name="menu" />`; opens a `.c-popover` with "Open in Chat", "Copy skill ID", "View author on Hub". |
| 2 | Skill name (`<h3 class="c-name">`) | `display, 18px, line-height 1.15, --content`. Two-line clamp via `-webkit-line-clamp: 2`. Not truncated at one line — users have grown to expect two. |
| 3 | Short description | `display italic, 13px, line-height 1.5, --content-soft`. Three-line clamp with a fade gradient on the third line (the gradient is `--content-soft → transparent`, not the MOAT-forbidden gradient text). |
| 4 | Meta hairline | Single hairline above, then mono 10px inline meta: `v` + `version` · `author` · downloads (if any) · trust level collapsed to a dot. The trust dot here is identical to row 1's badge — same color, same size, so the user reads the same information twice only if they look for it. No visual conflict. |
| 5 | Action row | Left: pill primary action that changes per state — `Install →` (not installed), `Open →` (installed, not running), `Updating…` (installing), `Installed` (installed + pulse-check visible). Right: a small mono label for `Updated 2d` or `+12 versions` (when an update is available without being installed yet). The full pill is a `<Button variant="primary">` for the install moment, `.btn-ghost` (mono uppercase 11) when idle. |

#### B.3 Per-state card appearances

The same card shape adapts to one of these states:

1. **Not installed (Hub browsable).** Primary = `Install →` (synapse-filled
   pill). Trust badge sits in row 1 in the trust color. No "behind the glass"
   thumbprint.
2. **Installed.** Primary = `Open →` (pollen-outline pill, clears on
   hover). A 6×6 **filled synapse dot** sits in the top-right corner
   adjacent to the menu — the "behind the glass" mark. The menu
   popover changes to include "Uninstall". Trust badge stays.
3. **Installing (in-flight).** The whole card flips on its Y axis over
   480ms to a "progress face" (see §5.4). The action button on the back
   face reads `Installing…` and shows a `Pulse phase="acting"`.
4. **Installed-settled.** Front face flips back. The action button
   reads `Open →`, top-right has the installed dot. A `Thread` draws
   across the bottom edge (the "completion" gesture per MOAT §3).
5. **Update available (installed, but newer on Hub).** Same as Installed
   but: (a) the trust dot is replaced with a **breathing pollen dot**
   (`pollen-breath` 1.6s loop, synapse-glow when focused), (b) the meta
   row's right-hand label reads `Update · v0.3.0 →` instead of "Updated
   2d", (c) the primary action becomes `Update →` (synapse-filled
   pill), pressed → goes into the Installing state with version diff in
   the back-face body.
6. **Error (install failed).** Same card shape; the trust badge swaps
   to `<Glyph name="warning" />` filled `--danger` for ~600ms, then
   returns to the trust color when retried. The action button reads
   `Try again` (pollen-outline). A `Thread` under the meta row stays
   broken (left→right only halfway) — the "err-hair" gesture.

### 3.4 Region C — Detail panel (360px, slides in from right)

The detail panel is a `.c-sheet` (per MOAT §2.8) that slides in when a
card is selected. It is owned by the page, not floating with a scrim —
the rest of the route stays live (Esc closes it). It contains:

| Block | What |
|---|---|
| **Sheet header** | Close `×` button (top-right, 32×32 hit-area, hairline focus halo per MOAT §2.1); skill title (display 28, `--content`); eyebrow mono caps: "— Trust · `official`". |
| **Description** | Full un-clamped description, display italic 15, `--content-soft`. |
| **Usage example** | A code block in `--surface-sunken` with mono 13. The example is **literally the prompt the user would type** to invoke this skill: e.g. `/run morning-briefing` or `morning briefing`. Copy-to-clipboard `<button>` glyph-only with a `<Tooltip label="Copy example">`. |
| **Dependencies** | Mono 12 list, each line `· name v0.1.2`. If empty, the line "No dependencies. This skill runs on its own." in italic display. |
| **Version history** | A vertical list of versions, mono 11, each line `[v0.2.1 · 2026-06-30 · 184 dl · community]`. Up to 8 visible; older collapse behind a `[+ 5 older] ` mono link. |
| **Audit trail** | Mono 11 line `Last run · 2 days ago by you. 7 runs in the last 30 days.` Empty case "Never run. The agent hasn't used this procedure yet." |
| **Primary action row** | Sticky to the bottom of the sheet (24px from bottom edge): one `<Button variant="primary">Open in Chat →</Button>` (always — `Open` if installed, `Install` if not). To the right of it, a `.btn-ghost` `Uninstall` (only when installed). |

The detail panel renders inside a `{#key detail?.id}` block so that
switching between cards re-mounts and replays the slide-in.

### 3.5 The footer band

A single hairline + a mono row at the bottom of the route body, max-width
1280px:

> `Skills are stored locally. The Hub is a source. Auto-created skills are
> the agent's procedures. Last synced · 12 min ago.`

The "Last synced" label re-renders when `skills.list` last succeeded; on
error it goes red and reads `Last sync failed at 14:02 · check connection`.

---

## 4. State Matrix — six states, full copy

These are the six visual states the route can be in. They are not
mutually exclusive — the Grid has a state, the Filter Rail has none,
the Detail Panel has its own selected-with-card state.

| # | State | What you see | Trigger / data condition |
|---|---|---|---|
| **S1** | **Empty (no skills in the shelf at all)** | Empty state — see §4.1 below. | `local.installed.length === 0` AND (no Hub results OR user filtered to "Behind the glass" + nothing there). |
| **S2** | **Loading** | The **entire** route body shows a `Thread` drawing left-to-right (`drawthread`, `--dur-slow`, `--ease`) across the page height — never a row of skeleton cards. The page header remains interactive. | `ipc.skillsList` or `ipc.hubSearch` pending; the request takes > 180ms. Below 180ms the route does not show a loading state. |
| **S3** | **Installing (one card mid-flip)** | The picked card flips Y over 480ms. The rest of the route stays interactive. The Action row of the back face reads `Installing…` with a `Pulse phase="acting"` (8px). The card cannot be re-clicked during the flip. | User pressed Install on a not-installed card OR pressed Update on an update-available card. State holds until `ipc.hubInstall({id})` resolves or rejects. |
| **S4** | **Installed (settled)** | The card's flip reverses. The action button reads `Open →`. The top-right "behind the glass" dot is a 6×6 filled synapse dot. A `Thread` draws along the bottom edge of the card over 520ms (`thread-draw-left-right`). The `Update · +12 versions` label disappears if it was there. | `ipc.hubInstall` returns `{ok: true, id}`. The local card appears in the next `skills.list` poll (or the route invalidates and refetches immediately). |
| **S5** | **Error** | An `ErrorState` component renders once at the route level (NOT inline in the grid). This means if one card failed, the user gets a clean explanation, not a polluted grid. | `ipc.skillsList`, `ipc.hubSearch`, `ipc.hubInstall`, `ipc.skillsDelete` rejects. |
| **S6** | **Update available** | An installed card swaps its top-right synapse dot for a **breathing pollen dot** (`pollen-breath` 1.6s, `--pollen`). Meta row's right-most label reads `Update · v0.3.0 →`. Action row reads `Update →`. | The card exists locally (installed) AND a Hub search with the same `hub_id` returns a `version` greater than the local `version`. |

### 4.1 The Empty state (S1) — exact copy

Three lines, exactly per MOAT §2.4:

> **What's here.** A shelf of procedures the agent can re-run — recipes
> for "summarize this PDF," "morning briefing," "diff two PRs." Some
> were written by you. Some were written by Condura from past tasks.
>
> **Why it's empty.** Nothing on this shelf yet. Run a complex task and
> Condura will save the procedure here automatically. Or bring one over
> from the other side.
>
> **Next action.** *[Browse the Hub →]* — the only button. Mono-pollen
> outline, opens `#/hub`. The route does not embed a Hub preview; it
> hands off cleanly.

Beneath the three lines (under the next action), three **example card
placeholders** render at 25% opacity with a small mono caption "What a
filled shelf looks like":

```
[ "morning briefing"               ]    [ "summarize a PDF"            ]
[ v0.2.0 · Condura · community     ]    [ v0.4.1 · you · official      ]

[ "diff two PRs"                  ]
[ v0.1.3 · you · experimental     ]
```

The example cards are not clickable. They're decorative populated
state per TEARDOWN §7 — they teach the shape before the user has
filled it. They breathe (synapse-glow pulse 4s loop, three dots
staggered) so the eye knows they're sample data.

### 4.2 The Loading state (S2) — exact copy

> **A thread draws across the page** (left-to-right, 1px, `--synapse`,
> `--dur-slow`, `--ease`). The eyebrow under it reads, in mono caps 11,
> `--content-faint`: `INDEXING THE SHELF · /skills`. Once the data
> resolves, the thread arrives 100% of the way across the page width
> and the cards fade in beneath it, staggered.

The route does not show partial cards while loading. If `skills.list`
returns in <180ms the user never sees the loading state at all. The
card grid is invisible until the first `skills.list` resolves.

### 4.3 The Installing state (S3) — exact copy

The chosen card flips on Y. Front-face and back-face both render; CSS
hides the wrong one with `backface-visibility: hidden`.

Front face (visible at rest, fades to 0°→90° over 480ms `--ease`):
the regular card (rows 1–5 above).

Back face (visible at 90°→360°, same 480ms `--ease`, mirrored):
```
┌────── mirrored card ──────┐
│                            │
│   Installing morning       │  display 18
│   briefing…                │
│                            │
│   ⬤                       │  Pulse phase=acting, 12px
│                            │
│   v0.2.1 · 184 dl          │  mono 11 faint
│                            │
└────────────────────────────┘
```

The pulse breathes at 1s. The card is non-interactive until the flip
back; clicks on it during flip are swallowed (no double-install).
On resolve: back face rotates past 90° → 360° (which means visual flip
back to front), the pollen check-mark draws itself onto the front-face
button over 320ms (the button morphs from `Installing…` to
`Installed`, the check is drawn via `<Glyph name="check" />` with a
stroke-dashoffset transition from 1→0, parallel to the thread-draw).

On reject: the card returns to its starting state, the trust badge
swaps to `<Glyph name="warning" />` filled `--danger` for 600ms, the
action button reads `Try again`. An ErrorState panel at route level
explains the cause. (See S5.)

### 4.4 The Installed-settled state (S4) — exact copy

After S3's success path, the card resolves to:

- Trust badge: unchanged.
- Top-right dot: 6×6 filled `--synapse` (the "behind the glass" mark).
- Action button: `Open →` (filled synapse pill, but `Open` swaps to pollen
  outline on hover per MOAT §2.7 tactile).
- Thread: a 1px synapse line draws across the card's bottom edge left-to-
  right over 520ms (`drawthread` keyframe, `--dur-slow`, `--ease`). This
  IS the moment of "completed" per MOAT §3.

No toast. No "Installed!" celebration. No animation other than the
thread-draw. Per MOAT §4 #6.

### 4.5 The Error state (S5) — exact copy

The route uses one `ErrorState` component (per MOAT §2.6, it should be
extracted from `Chat.svelte` / `Channels.svelte` / `Skills.svelte` /
`Hub.svelte` anyway — this spec is the first place that uses the
extracted form). It renders once per route, anchored above the grid.

The `ErrorState` shape (always three lines, exactly):

> **We couldn't reach the daemon.** (italic display 22, `--content`.)
>
> Cause: *`{noun}`*, e.g.
> - "Skills list returned no rows from `ipc.skills.list`."
> - "Hub search timed out after 2.4s."
> - "Install failed: `404 — skill not found on hub.condura.app`."
>
> Likely reason: *`{phrase}`*, e.g.
> - "The daemon was restarted."
> - "Your network is offline."
> - "This skill was unpublished since you pinned it."
>
> Next action: `[Try again]` pill — pollen-outline button. Above it on the
> right, a mono 11 link "[Open Settings →]".
>
> *(Below all of this, an `err-hair` rule, 1px, left→right draw over 600ms `--ease`.)*

### 4.6 The Update-available state (S6) — exact copy

The card stays in its Installed state, with two differences:

- Top-right dot: a **pollen dot that breathes** (`pollen-breath`,
  1.6s `--pollen`, 60%→100% opacity, `--ease`). On focus, the dot
  brightens to `--synapse-glow` (the pollen-dot breathes even when not
  focused, because the event is rare and the user will want to spot
  it).
- Meta row's right-most mono label: `Update · v0.3.0 →`.
- Action row: `Update →` (filled synapse pill, primary style).

Pressing `Update` → goes into S3 (Installing) with the same UI but the
back face shows:

```
│   Updating morning           │
│   briefing                   │
│   v0.2.1 → v0.3.0            │  mono 11 — the version diff
│   ⬤                          │  Pulse phase=acting, 12px
```

The post-update settled state (`v0.3.0 · community`) draws a thread
across the card bottom edge (same as S4).

---

## 5. Motion Choreography

The route's motion follows one rule: **every animation answers "what
is this communicating?"** (MOAT §4 #10). Decorative loops are
forbidden.

### 5.1 Entrance — cards fade in with a 40ms stagger

Trigger: `ipc.skillsList` and `ipc.hubSearch` first resolve. The
`{#key gridLoadId}` re-mounts the grid.

Per card:

```
opacity: 0 → 1
transform: translateY(8px) → 0
duration: 320ms
easing: var(--ease)
delay: i * 40ms, where i is the card's grid index (left-to-right,
top-to-bottom)
```

The stagger is computed in Svelte via `style:animation-delay={
${index * 40}ms}` (no IntersectionObserver needed — the cards render
together at the grid mount). The stagger is capped at 30 cards
(`max(0, min(29, i)) * 40ms`) so the last card doesn't appear 8s
after the first.

**Reduced-motion:** the entire 40ms-stagger sequence is replaced with
a single 0ms stagger — all cards appear at once. No fade. (Per
MOAT §2.3, one global rule in `condura.css` does this.)

### 5.2 Hover — card lifts 2px, hairline becomes synapse-strong

```css
.card:hover {
  transform: translateY(-2px);
  border-color: var(--hair-strong);
  box-shadow: var(--shadow-card);
}
```

`:active` adds the global tactile rule (per MOAT §2.7): `transform:
translateY(0) scale(0.985) filter: brightness(0.96) saturate(1.05)`.

`:focus-visible` adds a **rounded halo** (per MOAT §2.1): for cards
with `border-radius: var(--r-md)` (16px), the halo is `box-shadow:
var(--shadow-card), 0 0 0 5px var(--pollen-halo-color)` — not a
rectangular outline.

**Hard delete:** the current `Skills.svelte:283` rule
`transform: translateY(-4px) rotateX(2deg);` and the `:active`
companion `translateY(-2px) rotateX(1deg) scale(0.98);`. These are
called out by `MOAT.md §1.5` by name. The `perspective: 1000px;` on
the grid and `transform-style: preserve-3d;` on each card are
deleted with them.

**Reduced-motion:** `transform: none`; the card glows its focus halo
and goes no further.

### 5.3 Select — card slides right, detail panel slides in

When a card is `aria-pressed="true"` or clicked:

- The card's left edge becomes a 2px synapse strip (a vertically-
  oriented `<Thread />`).
- The grid's `padding-right` increases by 360px (`grid-wrap` style
  with CSS transition) over 320ms `--ease`. This pushes the grid
  content leftward visually.
- The detail panel (Region C) slides in from `right: -360px →
  right: 0` over 320ms `--ease`. It uses `transform: translateX(...)`,
  not `left:` (transforms are GPU-compositable, smoother on lower-end
  devices).

If the user deselects (Esc, second click on the same card, close ×),
both move back over 320ms.

**Reduced-motion:** the grid padding and panel both move instantly
with no slide.

### 5.4 Install — card flips on Y, back face progress, settled thread

Trigger: user pressed Install on a not-installed card OR Update on an
update-available card.

#### Stage A — front face fades to 90° over 480ms `--ease`

The selected card adds `data-state="installing"`. CSS:

```css
.card[data-state="installing"] {
  animation: card-flip-out 480ms var(--ease) forwards;
}
.card .face { backface-visibility: hidden; transform: rotateY(0); }
.card .face--back { transform: rotateY(180deg); }

@keyframes card-flip-out {
  0%   { transform: rotateY(0); }
  100% { transform: rotateY(180deg); }
}
```

#### Stage B — back face visible from 90° → 360° (same 480ms `--ease`)

```css
.card[data-state="installing"] .face--back {
  animation: card-flip-in 480ms var(--ease) forwards;
}
@keyframes card-flip-in {
  0%   { transform: rotateY(180deg); }
  100% { transform: rotateY(360deg); }
}
```

The Pulse on the back face (`phase="acting"`, 12px) starts at the
beginning of stage B and breathes at 1s for the duration of the install.

#### Stage C — settled on success

`data-state` is removed from the card. The card returns to its
starting rotation. The action button swaps from `Installing…` to a
`<Glyph name="check" />` drawing itself (stroke-dashoffset 1 → 0 over
320ms `--ease`). Then 200ms later, the button label morphs to
`Installed` (per MOAT §4 #6 — no celebration, just the state). A
`<Thread />` draws across the card's bottom edge (a hairline 1px,
`--synapse`, left→right, 520ms `--ease`) — this is the "completion"
gesture (MOAT §3).

#### Stage C — failed install

`data-state` returns to its previous value. The card flips back to
rotation 0 over 320ms `--ease` (faster than the install flip — the
reversal is "things didn't go as planned, here's the card you came
in with"). The trust badge swaps to `<Glyph name="warning" />` filled
`--danger` for 600ms, then back. The action button reads `Try again`.
A route-level `ErrorState` renders above the grid (S5).

#### Reduced-motion

If `prefers-reduced-motion: reduce` is set, **Stage A and B are
skipped entirely.** The card does not flip. Instead, the card's
action button changes in place from `Install →` to `Installing…`
with the acting-pulse next to it. On success, the button morphs in
place to `Installed` and the thread-draw still happens. The
`<Glyph name="check" />` drawing itself is also skipped (the check
just appears).

### 5.5 Update — pollen dot breathes

The pollen dot is `animation: pollen-breath 1.6s var(--ease)
infinite`.

```css
@keyframes pollen-breath {
  0%, 100% { opacity: 0.6; transform: scale(1); }
  50%      { opacity: 1.0; transform: scale(1.15); }
}
```

On `:focus-within` (parent card is focused), the dot brightens to
`--synapse-glow`:

```css
.card:focus-within [data-update-dot] {
  background: var(--synapse-glow);
  box-shadow: 0 0 8px 2px var(--synapse-glow);
}
```

This is the ONLY idle loop on the route, and it answers the question
"What is this communicating? *An update is here.*" (MOAT §4 #10.)

**Reduced-motion:** the breath becomes a static 100% opacity dot in
pollen. No scale change.

### 5.6 The Thread (MOAT §3) — when it draws

The thread draws across the card in **four** moments only:

1. **Settled install (S4 success)** — bottom edge of the card.
2. **Settled update (S6 success)** — bottom edge of the card.
3. **Post-retry success** — same as S4.
4. **Search query resolves** — across the top of the grid (a 1px
   segment, not a full-width line; a quiet "the shelf has been
   re-indexed" signal).

The thread is the spine of this route the same way it is the spine of
the titlebar (`TitlebarThread.svelte`) and the chat-turn dividers.
It is the only allowed flourish for "this is now finished." (MOAT §3.)

---

## 6. Keyboard

The route's keyboard surface is added to the global shortcuts
documented in MOAT §2.10. Keys are bound at the route level
(`onMount` registers them, `onDestroy` removes them).

### 6.1 Global keybindings (already in scope of MOAT §2.10)

| Key | Action |
|---|---|
| `/` | Focus the search input. If the user is already inside a text input, `/` types a literal `/`. |
| `Esc` | Closes the topmost overlay/panel — first the detail panel, then any open `.c-popover`, then clears the search input. If all three are already closed, Esc is a no-op (does not navigate). |
| `⌘K` | Opens the command palette (already global). |
| `⌘,` | Opens Settings (already global per MOAT §2.10). |

### 6.2 Grid navigation

The grid is one virtual focus-trap when active. Tab inside the grid
travels cards left-to-right, top-to-bottom, skipping focused cards
(arriving back at the first card after the last). Shift+Tab reverses.

| Key | Action |
|---|---|
| `Tab` | Next card. |
| `Shift+Tab` | Previous card. |
| `Enter` | Open the focused card's detail panel. (Same as clicking the card body. The trust badge and card menu keep their own Tab stops; they do not `Enter` into the detail panel.) |
| `Space` | Toggle the focused card's menu popover (same as clicking `[⋯]`). |
| `→` | Move focus to the next card in row-major order. If focus is on the rightmost card of a row, jump to the first card of the next row. |
| `←` | Symmetric — row-major reverse. |
| `↓` | Move focus to the card directly below the focused card (same column). If no card exists directly below, jump to the first card in the next row that exists in that column or the nearest column to the right. |
| `↑` | Symmetric. |
| `Home` | Move focus to the first card (top-left). |
| `End` | Move focus to the last card (bottom-right). |
| `PageDown` / `PageUp` | Scroll the grid container (when the grid extends past the viewport). |

### 6.3 Card-level keybindings (when a card is focused)

| Key | Action |
|---|---|
| `⌘I` | Install the focused card (only enabled when card state is "not installed"). Mirrors the "Install →" primary action. |
| `⌘U` | Uninstall the focused card (only enabled when card state is "installed"). Mirrors the "Uninstall" ghost action. |
| `⌘Enter` | Open in Chat — equivalent to clicking the primary action button when installed; install-then-run when not installed. (This is the discovery shortcut: install the skill and immediately send a `/run <skill>` chat message.) |

### 6.4 Detail panel (open) keybindings

When the detail panel is mounted and focused is inside it:

| Key | Action |
|---|---|
| `Tab` | Cycles through detail panel interactive elements (top to bottom). |
| `Esc` | Closes the detail panel and returns focus to the originating card. |
| `⌘C` | With the focus on the usage-example code block, copies the example to clipboard. (Same as the Copy button on that block.) |
| `⌘I` / `⌘U` | Same as card-level keybindings; they trigger the primary action row in the detail panel. |
| `⌘Enter` | "Open in Chat" primary action (the sticky-bottom button). |

### 6.5 Filter rail keybindings

| Key | Action |
|---|---|
| `Tab` | Cycles through rail groups, then the rail interactive elements within each group, top to bottom. |
| `Space` | Toggles a checkbox / chip in the rail. |
| `Enter` | Toggles a checkbox / chip in the rail. (Both Space and Enter, per mature forms convention.) |
| `⌘F` | Opens the rail as a `.c-popover` sheet — useful at narrow viewports where the rail is collapsed. |
| `Esc` | When focus is inside a `.c-popover`, closes it and returns focus to the trigger. |

### 6.6 Reduced-motion note

All keybindings above are unchanged when the user has `prefers-
reduced-motion: reduce` set. Animations change; the keys don't. (Per
MOAT §2.3, components never branch on motion prefs; the global CSS
handles it.)

### 6.7 Focus visibility

Every card, every pill, every chip carries a rounded halo on `:focus-
visible` (per MOAT §2.1). Cards have a 5px pollen halo + 2px synapse
ring stacked. Buttons and chips inherit the global `--shadow-focus`.

No `outline: 1px solid var(--content)`. No rectangular halos. (MOAT
§4 #8.)

---

## 7. Components Used

The route is composed of these components. The full prop contracts
below belong in each component's source file — this spec is the
short-form.

### 7.1 `<SkillCard />`

Sits in `app/web/frontend/src/lib/condura/SkillCard.svelte`.

```ts
type CardState =
  | { kind: 'not-installed' }
  | { kind: 'installed'; ranAt?: string }
  | { kind: 'installing'; version?: string }
  | { kind: 'error'; reason: string };

let {
  skill: InstalledSkill | HubSkillMeta,
  state: CardState,
  selected = false,
  onselect: (id: string) => void,
  oninstall: (id: string) => void,
  onuninstall: (id: string) => void,
  onopen: (id: string) => void,
  onmenu: (id: string, anchor: HTMLElement) => void,
  class: cls = '',
} = $props();
```

Owns:

- The 5-row card layout per §3.2.
- The per-state visual transformation (front face vs back face; the flip
  animation; the trust badge swap on error).
- The hover / focus / selected visual states.
- The `.card-menu` button + popover content (built inline; the popover
  uses `<CardMenu.svelte>` shared with `Hub.svelte`).
- The card-level keyboard handlers (`⌘I`, `⌘U`, `Enter`).

Does **not** own:

- The grid container (owned by `Skills.svelte`).
- The detail panel (owned by `<SkillDetailPanel />`).
- The data fetching (owned by the page-level script).

### 7.2 `<FilterRail />`

```ts
let {
  counts: {
    view: { installed: number; hub: number; all: number };
    trust: { official: number; community: number; experimental: number };
    sources: Record<SourceKey, number>;
    tags: Array<{ name: string; count: number }>;
    authors: Array<{ name: string; count: number }>;
  },
  filters: Filters,
  onfilters: (next: Filters) => void,
  collapsed = false,  // true below 960px
  class: cls = '',
} = $props();
```

Owns:

- The four filter groups (View / Trust / Source / Category / Author).
- The sticky positioning (when not collapsed).
- The `[⛁]` collapse trigger (when below 960px viewport).
- The "Reset filters" link (only when any filter is active).
- The filter-level keyboard handlers (Space/Enter on chips).

Does **not** own:

- The grid.
- The detail panel.
- The IPC — filters are an entirely client-side concern.

### 7.3 `<SkillDetailPanel />`

```ts
let {
  skill: InstalledSkill | HubSkillMeta,
  deps: Array<{ name: string; version: string }>,
  versions: Array<{ version: string; published_at: string; downloads: number; trust: string }>,
  stats: { lastRunAt: string | null; runsLast30d: number },
  state: CardState,
  onclose: () => void,
  oninstall: (id: string) => void,
  onuninstall: (id: string) => void,
  onopen: (id: string) => void,
  class: cls = '',
} = $props();
```

Owns:

- The 360px `.c-sheet` slide-in / out (transform translateX).
- The seven blocks per §3.4 (header, description, example, deps,
  versions, audit, primary action).
- The panel-level keyboard handlers (Esc closes; `⌘C` on usage block;
  `⌘Enter` on the primary action).

Does **not** own:

- The card selection state — the page owns selectedSkillId and passes
  the resolved skill here.

### 7.4 `<Thread />` — already exists, reused

```ts
let { orientation: 'h' | 'v' = 'h', draw = true, glow = true, class?: string } = $props();
```

Used in:

- The grid's bottom-card-settled thread (S4 / S6 success).
- The search-results thread (per §5.6 — searches resolve with a thread).
- The detail panel's header hairline (left edge of the sheet, a 1px
  synapse vertical).
- The `ErrorState`-rendered `err-hair` (the existing pattern, untouched).

### 7.5 `<Pulse />` — already exists, reused

```ts
let { phase: 'idle'|'thinking'|'awaiting'|'acting'|'consent'|'error'|'ok', size?: number, class?: string } = $props();
```

Used in:

- The loading state S2 (single 8px `idle` pulse next to the eyebrow —
  paired with the thread-draw).
- The installing back face (12px `acting` pulse, 1s breath).
- The settled-installed button (6px `ok` pulse, 4s breath — this is
  the rare "permission to be a loop" exception because it
  communicates "alive"; MOAT §4 #10's narrow exception).

### 7.6 `<Glyph />` — already exists, reused

```ts
let { name: string, size?: number, stroke?: number, class?: string } = $props();
```

Icons used on this route, with their canonical `name` from `icons.ts`:

| Where | Glyph name | Default size |
|---|---|---|
| Card row 1, trust — official | `dot-active` (filled synapse) | 8 |
| Card row 1, trust — community | `dot` (hairline synapse) | 8 |
| Card row 1, trust — experimental | `dot-active` (filled pollen) | 8 |
| Card row 1, menu trigger | `menu` | 16 |
| Card row 1, close on hover | `close` | 16 |
| Action row, install | `chevron-right` | 12 inline with label |
| Action row, open | `chevron-right` | 12 inline with label |
| Action row, update | `spark` | 12 inline with label |
| Action row, error | `warning` | 12 inline with label |
| Top-right installed dot | (CSS-rendered, no glyph) | 6×6 dot |
| Top-right update dot | (CSS-rendered, no glyph) | 6×6 dot |
| Detail panel close button | `close` | 14 |
| Detail panel copy example | `chevron-right` mirror? **No.** Copy uses no icon — the label is "Copy" mono-uppercase 11. (MOAT §4 #2 — no new glyphs invented; the chip is text-only.) |
| Detail panel audit trail info | `info` | 12 |
| Detail panel install primary | `chevron-right` | 12 |
| Detail panel uninstall | `trash` | 12 |
| Footer band sync status | `sync` | 12 |

If any icon on this table doesn't exist in `icons.ts`, the spec says
to **add it to `icons.ts`** (per MOAT §4 #2), not to use Unicode or
emoji as a shortcut.

### 7.7 `<Button />` — already exists, reused

```ts
let {
  variant: 'primary' | 'secondary' | 'ghost' | 'danger',
  size: 'sm' | 'md' | 'lg',
  icon: string | null,
  loading: boolean,
  children: Snippet,
  onclick: (e: MouseEvent) => void,
  ...
} = $props();
```

Variants used:

- **primary** — Install, Update, Open in Chat (synapse-filled pill;
  pollen outline on hover per MOAT §2.7).
- **secondary** — Filter chips, "Reset filters" (paper-2 fill, hairline).
- **ghost** — Try again, Uninstall, "Open" when Installed (mono-uppercase
  link with a 1px underline that grows from the center on hover per
  MOAT §5.2).
- **danger** — Never; uninstall is a ghost because the action is
  reversible (per the safety-layer spec, uninstall is a DESTRUCTIVE
  WRITE, but the GUI affordance is still "Uninstall" — the gatekeeper
  consent modal handles the friction; MOAT §10).

### 7.8 `<Tooltip />` — to be created per MOAT §2.9

```ts
let {
  label: string,
  placement: 'top' | 'right' | 'bottom' | 'left' = 'top',
  delay: number = 400,  // ms hover before tooltip appears
  exit: number = 75,    // ms exit
  children: Snippet,
} = $props();
```

Used on:

- Card menu `[⋯]` button ("Card menu").
- Detail panel close `×` ("Close").
- Detail panel "Copy example" button.
- All icon-only buttons in the toolbar's right cluster.
- All non-`<label>`-paired checkboxes in the filter rail (the
  checkbox gets the tooltip on hover-failure of label-discovery).

The tooltip is **never** used as a celebration; MOAT §4 #6 forbids
that.

### 7.9 `<ErrorState />` — to be extracted per MOAT §2.6

```ts
let {
  head: string,         // "We couldn't reach the daemon."
  cause: string,        // one noun
  reason: string,       // one phrase
  onretry: () => void,
  retryLabel?: string,  // defaults to "Try again"
  onsettings?: () => void,  // optional "Open Settings" link
  class?: string,
} = $props();
```

Renders the italic display headline + mono cause/reason + retry pill +
`err-hair` (per the existing pattern, unchanged). Used **once** per
route, not inline.

### 7.10 `<EmptyState />` — to be extracted per MOAT §2.4

```ts
let {
  what: string,         // "A shelf of procedures the agent can re-run."
  why: string,          // "Nothing on this shelf yet."
  action: Snippet,      // the Browse the Hub button (rendered via snippet)
  sample: Snippet | null,// the three example populated cards
  class?: string,
} = $props();
```

The Skills route is the first place this component is used. After it
ships, `Chat.svelte`, `Hub.svelte`, `Settings.svelte` empty states
should also migrate to it (out of scope for Phase 4 — they can
continue using inline pattern until a cleanup pass).

---

## 8. Data Fetched

The route's IPC contract is six methods. Signatures come from
`/Users/sahajpatel/synaptic/app/web/frontend/src/lib/ipc/client.ts`
and `…/ipc/types.ts` (existing).

### 8.1 `ipc.skillsList(limit = 100)`

RPC: `skills.list`. Returns `InstalledSkill[]`. Used to populate the
"behind the glass" half of the grid.

```ts
// Existing: app/web/frontend/src/lib/ipc/client.ts:368
skillsList(limit = 100): Promise<InstalledSkill[]>
```

Cache policy: fetched on mount and on user-triggered refresh. Stale-
while-revalidate: re-fetched in the background every 60 seconds while
the route is mounted, never while the page is hidden (`document.
visibilityState !== 'visible'` skips the poll). The user is never
shown a loading state during a stale refresh — the cards keep showing
the last good data.

Error policy: failed calls land in `ErrorState` (S5). The route does
**not** fall back to `skills.list` cached data; instead, the cached
data stays visible with an inline mono note "Showing cached data ·
last refresh at {HH:MM}" in the toolbar's right cluster.

### 8.2 `ipc.hubSearch(query, limit = 20)`

RPC: `hub.search`. Returns `HubSearchResult { skills: HubSkillMeta[];
total: number; query: string }`. Used to populate the "other side"
half of the grid and to drive the update-available detection (S6).

```ts
// Existing: app/web/frontend/src/lib/ipc/client.ts:354
hubSearch(query: string, limit = 20): Promise<HubSearchResult>
```

Cache policy: debounced 250ms after the user types. The route uses the
`query` "skill" on mount (the broad priming pass, per `Hub.svelte:46`).
When the user clears the query, the route returns to the primed set.

When `Filters.view === 'installed'`, this IPC is **not called** (the
route fetches only `skills.list`). When `Filters.view === 'hub'`, this
IPC is called and `skills.list` is **not** (the route treats the Hub
as the authoritative list of "not-yet-installed" skills). When
`Filters.view === 'all'`, both are called; the route merges by
`hub_id` (so an installed card with a `hub_id` only renders once,
with its installed-side metadata taking precedence).

Update detection: when a card exists locally (from `skills.list`) AND
`hub.search` returns a matching `hub_id` with a higher `version`,
the card adopts the S6 state. The route stores the last-seen Hub
version per `hub_id` in `sessionStorage` so the pollen dot doesn't
flicker across re-mounts.

Error policy: same as 8.1.

### 8.3 `ipc.skillsGet(id)`

RPC: `skills.get`. Returns `InstalledSkill`. Used by the detail panel
when the user clicks a card whose state requires fresh server-authoritative
metadata (e.g. the audit-trail `runsLast30d` field comes from a deeper
RPC; in v0.1.0, only the version history and last-run timestamp are
returned; the rest is computed client-side from the data already in
memory). May be deferred if `skills.list` returns enough — flagged here
so Phase 4 doesn't make extra calls it doesn't need.

```ts
// Existing: client.ts:371
skillsGet(id: string): Promise<InstalledSkill>
```

### 8.4 `ipc.skillsDelete(id)`

RPC: `skills.delete`. Returns `{ ok: boolean }`. Triggered by the
"Uninstall" button in the detail panel (and `⌘U` shortcut).

```ts
// Existing: client.ts:374
skillsDelete(id: string): Promise<{ ok: boolean }>
```

UX: the card stays in the grid for 400ms after the delete resolves
(because the optimistic state is set immediately), then it animates
out (opacity 1 → 0, 240ms `--ease`) and the grid reflows. The detail
panel closes if it was open. If the user re-renders the page, the
card is gone for good.

Errors: route-level `ErrorState`; the optimistic state reverts.

### 8.5 `ipc.hubInstall(id)`

RPC: `hub.install`. Returns `HubInstallResult { ok: boolean; id: string }`.
Triggered by the "Install →" pill (and `⌘I` shortcut) when the card is
not installed.

```ts
// Existing: client.ts:360
hubInstall(id: string): Promise<HubInstallResult>
```

UX: the card flips to S3 immediately. The grid keeps the old card
visible until the install resolves (the optimistic state is the
back-face installation animation). On `ok: true`, the card
transitions to S4. On failure, it transitions to the failed-install
sub-state of S3, then to the route-level `ErrorState` (S5).

**Gatekeeper consent:** if `hub.install` would be classified as a
WRITE-by-network action, the existing `ConsentModal` opens first
(per the global pattern; nothing custom here). The Skills route
renders a disabled card action with a `<Tooltip label="Requires
consent">` until the user has approved at least one install in
the current session.

### 8.6 `ipc.hubPublish(id, path)` — out of immediate scope

The Skills route is described in the brief as including a "publisher."
Phase 4 keeps the publisher UI **as a primary action in the
detail panel** for skills that are user-authored (i.e. have no
`hub_id` and `author === 'you'`):

- The user-authored detail panel gets a third action row, "Publish to
  Hub →" (`.btn-secondary`, `pollen` outline). Clicking it routes to
  `#/hub/publish/<skill-id>`, which is the existing publish flow
  (per CLAUDE.md §14G).
- This spec does **not** render the publish modal in-line — it
  delegates to the existing flow.

The `ipc.hubPublish` IPC is not called by Skills in Phase 4; it is
called by the existing publish route. Confirming that avoids creating
a second publisher surface, which would violate MOAT §1.2
("one source per role").

### 8.7 Replays — audit / install / sync events

The detail panel's audit trail uses no new IPCs. It composes from:
- `lastRunAt` and `runsLast30d` from `skills.get` (when
  available) or computed client-side from a sliding window over
  `ipc.replayTimeline()` already in memory (the route does **not**
  fetch `replay.timeline` itself).
- The "Last synced" footer reads from the `skills.list` last-success
  timestamp persisted in `sessionStorage`; on error, from the
  `ErrorState` resolution.

---

## 9. Design Decisions

These are the load-bearing calls — the place where this spec
disagrees with the current `Skills.svelte` and where it earns the
MOAT bar.

### 9.1 Flatten the shelf — kill the 3D bookshelf

**The problem (per MOAT §1.5).** The current `Skills.svelte:283`
applies `transform: translateY(-4px) rotateX(2deg);` on hover, with
`perspective: 1000px;` on the parent grid. There is no 3D surface
to read against; cards have no varied Z. This is the "vibe-coded
flex" in its purest form. MOAT §1.5 names this code by line number
and instructs to delete `perspective`, `transform-style`, and both
`rotateX` calls.

**What this spec does.** The grid loses `perspective` entirely.
Cards lose `transform-style: preserve-3d`. Hover is `translateY(-2px)
+ box-shadow: var(--shadow-card)`. The `-2px lift alone reads as
elevation`; the 3D tilt reads as the designer showing off.

### 9.2 One installation source — kill the "Install from Skills" tab

**The problem.** The original `Skills.svelte` was a local-skills
index; the `#/hub` route was the install-from-curated flow. Two
routes, two ecosystems, two metaphors. The user's request for
"Skills (the local skills browser + installer + publisher)"
implies the unification.

**What this spec does.** Skills is one route, one grid, one filter
(View → Behind the glass / Other side / All). Hub is the
deep-search-and-publish surface for when the user wants to leave the
Skills shelf and dive into the curated collection. The user does
not have to learn two routes to install one skill.

This is consistent with the 2026 mature pattern per TEARDOWN §7:
"the input box is the front door" — and here the install affordance
is the front door of every skill card.

### 9.3 Honest trust badges — kill the "Verified ✓"

**The problem.** The current `Skills.svelte` does not render trust
badges at all; the meta row just reads `author`. The Hub bibliography
in the design intent (per MOAT §4 #5 — "no fake enthusiasm") forbids
inventing a single-tier "verified" badge that hides the
sources/review process from the user.

**What this spec does.** Three trust levels, sourced exactly from
the data. No new tier, no green check for "everything that comes
from Condura is verified."

| Level | Honest definition |
|---|---|
| Official | Curated by Condura maintainers; signed against the org key. |
| Community | Published by a community author with ≥30 days on the Hub and ≥10 downloads. |
| Experimental | Everything else — new, unreviewed, or from an author without a community history. |

Each card carries the badge as a 6px dot at row 1 + a one-word
mono label. The user sees the level on every card, every time. Per
MOAT §4 #4 (no rainbow accents), the colors are `--synapse` filled
(official), `--synapse` outlined (community), `--pollen` filled
(experimental). No purple, no teal.

### 9.4 The empty state teaches — not "No skills yet"

**The problem.** The current `Skills.svelte:82-86` reads:

> "No skills yet."
>
> "Run a complex task — Condura will save the procedure as a skill
> automatically."

This is *close* to right, but missing the third line. MOAT §2.4
demands three lines always: **what** / **why empty** / **next
action.**

**What this spec does.** The empty state (S1 in §4.1) follows the
three-line shape exactly, adds three populated example cards under
40% opacity to teach the shape, and gives a single mono-pollen
"Next action" link to the Hub. This is the TEARDOWN §7
"opinionated-populated empty states" pattern. The user never sees
an empty page; they see an empty shelf with three sample items
already arranged on it, and one button pointing at the curated
library.

### 9.5 The hover is honest, the focus is a halo, the press is mass

**The problem.** MOAT §2.2 calls out that `scale(0.97)` on press
makes a button look smaller but does not feel pressed. The
current `Skills.svelte:288` uses `translateY(-2px) rotateX(1deg)
scale(0.98)` — three transforms stacked, none of which is
"weight."

**What this spec does.** The press state is owned by the global
`.tactile` class (MOAT §2.7): `transform: translateY(0.5px)
filter: brightness(0.95) saturate(1.05)`. The card visibly settles
into the page. No stacked transforms. No scale-pop.

### 9.6 Rounded halos, not rectangular outlines

**The problem.** MOAT §2.1 calls out that the universal
`:focus-visible` rule uses a square `box-shadow` even for rounded
elements.

**What this spec does.** Cards have `border-radius: var(--r-md)`
(16px) → focus halo is `box-shadow: var(--shadow-card), 0 0 0 5px
var(--pollen-halo-color), 0 0 0 2px var(--synapse)`. The pollen
halo is rounded because `border-radius` clips the
`box-shadow`. Buttons inherit the global rounded halo from MOAT
§2.1's update to `condura.css:299-302`. No `outline: 1px solid
var(--content)` anywhere.

### 9.7 The overlay is a sheet, not a modal

**The problem.** The current `Skills.svelte:108-127` builds a
modal-style overlay (`<div class="overlay">` + `backdrop-filter:
blur(4px)` + click-to-close on the scrim) for what is actually a
detail panel. MOAT §2.8 names this: "No taxonomy. Modal vs sheet
vs popover are not defined anywhere in codebase." MOAT §4 #3
also names this: the `backdrop-filter: blur(4px)` on the overlay
is exactly the glassmorphism the rule forbids.

**What this spec does.** The detail panel is a `.c-sheet`
(slide-from-right, no scrim blocking the rest of the page, Esc
closes). The card menu `[⋯]` is a `.c-popover`. The empty state
is a `.c-popover`-free inline card. No `backdrop-filter: blur`
on this route. No rect overlays. The aesthetic of "the rest of
the route stays live while you read a skill's details" matches
the user's request and matches MOAT §2.8.

### 9.8 Errors live at the route level, not in the grid

**The problem.** The current `Skills.svelte:71-81` inlines an
err-state block (`<div class="err-state">`) inside the component,
re-declared across `Chat.svelte`, `Channels.svelte`, and
`Hub.svelte`. MOAT §1.2 calls this out as four copies of the same
block.

**What this spec does.** One `<ErrorState />` component is used
**once** per route, anchored above the grid. The 240px rail
stays live so the user can filter the error away if they want.

### 9.9 The thread is the moment — limit it to four uses

**The problem.** A "thread draws to celebrate" pattern can become
a tic (MOAT §1.7 names the same pattern for `.alive`). The route
could draw threads everywhere and lose the spine.

**What this spec does.** The thread draws across the card on
exactly **four** moments:

1. Install settled (S4 success).
2. Update settled (S6 success).
3. Post-retry success.
4. Search resolved (a single hairline at the top of the grid,
   not a card-level celebration).

`err-hair` is unchanged: when an error resolves, the existing
hair-line draws left→right. No new thread uses.

### 9.10 Update notifications use a pollen dot, not a badge

**The problem.** The user could imagine "Installed · Update
available" rendered as a yellow dot, a colored badge, or a
notification toast. Each of these eats space on a small card.

**What this spec does.** The pollen dot replaces the synapse dot
in the same position. Same size, same shape, different color, and
it breathes (1.6s `pollen-breath`). The user's eye is trained to
read the dot's color: synapse = installed; pollen = update; pulse
red = error. The pollen dot is also reduced-motion-aware: a static
pollen dot is shown.

### 9.11 The keyboard is exhaustive, the focus is always visible

**The problem.** The current Skills route has no keyboard handlers
beyond what the browser does for `<button>` elements. The grid is
not a single Tab stop, so `Tab` walks through every card's
internal controls; arrow keys do nothing.

**What this spec does.** The grid is one virtual trap that uses
its own arrow / Home / End / PageUp / PageDown handling, mirroring
what every data grid does in 2026. The card-level shortcuts
(`⌘I`, `⌘U`, `⌘Enter`) shortcut the primary actions. The
filter rail is keyboard-navigable with full Space/Enter handling.
The detail panel has its own Trapped-when-open behavior with Esc
that returns focus to the originating card. Every focused element
carries the rounded halo (MOAT §2.1).

### 9.12 Don't ship a new publisher — route to the existing flow

**The problem.** The brief names "publisher" as a Skills concern.
A naively-built Skills page would re-implement the publish flow
that already exists at `#/hub/publish/<id>` (per CLAUDE.md §14G).

**What this spec does.** The detail panel has a "Publish to Hub →"
link for user-authored skills, and that link routes to the
existing publish flow. No second `PublishModal.svelte`. No
duplication. Per MOAT §1.2 ("three named primitives"), we use
the existing surface.

---

## 10. What this spec deletes from the current `Skills.svelte`

The following are the **explicit deletions** the implementation must
do when this spec replaces the existing file. Each row cites the
MOAT/TEARDOWN finding that motivates the deletion.

| Current `Skills.svelte` line(s) | What it does | Why it goes |
|---|---|---|
| Lines 65–69 | Inline `<Pulse phase="thinking" size=8 />` + "INDEXING…" mono label as the loading state. | MOAT §2.5 — loading states use Thread draws, not a single breathing dot. The mono label is preserved at the eyebrow; the pulse is moved to the toolbar next to the search field, paired with the thread. |
| Lines 71–81 | Inline `<div class="err-state" role="alert">` with err-row / err-head / err-sub / err-actions / err-hair all declared inline. | MOAT §1.2 — this same block is also in `Chat.svelte:213-221` and `Channels.svelte:150-158`. Extract to `<ErrorState />`. |
| Lines 82–86 | Inline empty state "No skills yet." + "Run a complex task…" | MOAT §2.4 — three-line teach pattern. Use `<EmptyState />`. |
| Lines 88–105 | The card loop. The card has trust `data-author` only, no trust badge. | Spec §3.2 swaps author for the three-tier trust badge and adds the action row. |
| Lines 108–127 | The overlay + sheet built inline. `<div class="overlay">` with `backdrop-filter: blur(4px)`, click-to-close on the scrim, slide-in sheet on the right. | MOAT §4 #3 + §2.8 — deletes glassmorphism on cards. Uses `.c-sheet` taxonomy. Loses the scrim entirely (`.c-sheet` does not block the rest of the route). |
| Lines 254–262 | `@keyframes err-hair-draw { to { transform: scaleX(1); } }` declared inline. | MOAT §2.6 — the `err-hair` token is the only error flourish; it lives in `condura.css` (or wherever `<ErrorState />` declares it). One copy, not four. |
| Lines 268–289 | `perspective: 1000px` on the grid + `transform-style: preserve-3d` on each card + `rotateX(2deg)` hover + `rotateX(1deg) scale(0.98)` press. | MOAT §1.5 — named by line number. Deleted in full. |
| Lines 295–308 | The colored vertical `<Thread />` strip on each card (`data-author="you"` → pollen, `data-author="agent"` → synapse). | Spec replaces this with the trust badge as the row-1 leftmost element. The colored vertical thread stays **only** inside the detail panel header — for the selected skill, signaling "this one is currently on stage." |
| Lines 309–355 | The card body / name / author / desc / foot ("Run →" + "Improve" mono links). | Spec replaces with the 5-row anatomy per §3.2. Author is collapsed into the meta row; a new trust badge is added at row 1; the action row is a primary pill. |
| Lines 357–380 | `<div class="overlay">` + `<aside class="sheet">` slide-in from the right. | Spec replaces with `<SkillDetailPanel />` (`.c-sheet`). No scrim. Esc closes (managed by the panel). |
| Lines 382–413 | `<button class="s-close">` × overlay close. | Spec replaces with the global `<Button variant="ghost" icon="close">` from MOAT §2.7 / 2.9. |
| Lines 415–511 | `.s-eyebrow` / `.s-title` / `.s-desc` / `.s-steps` / `.s-actions` / `.s-run` / `.s-improve` rules declared inline. | Spec replaces with the panel blocks per §3.4. The eyebrow mono 11px is the same convention; the title is the same display rule; the rest is fresh. |
| Lines 513–521 | `@media (prefers-reduced-motion: reduce) { .card, .card:hover { transform: none; } .sheet { animation: none; } }` declared inline. | MOAT §2.3 — single global rule. Delete. |

After these deletions, the file shrinks from 521 lines to roughly
180 lines — `<Skills />` becomes an orchestrator, and the work moves
into `SkillCard.svelte`, `FilterRail.svelte`, `SkillDetailPanel.svelte`,
`EmptyState.svelte`, `ErrorState.svelte`. MOAT §1.2's
"shared component earns the same flourish everywhere" pays off the
moment any of the four routes (Chat / Channels / Hub / Skills) need
a new error state or empty state — the answer is reuse, not
re-declaration.

---

## Closing note for Phase 4

This spec is exhaustive enough that a single screen of type design
and animation can be built without follow-up. Every state has its
copy; every motion has its duration; every key has its handler; every
icon has its name; every IPC has its signature; every deletion has
its reason.

The fourteen things you'll find when you implement from this spec:

1. **No** 3D perspective on the grid, ever.
2. **No** glassmorphism on the detail panel — it's a paper sheet on
   paper.
3. **No** "Verified ✓" badge — three honest trust levels, all visible.
4. **No** "Welcome aboard!" energy — three-line teach in the empty
   state, sample cards showing the shape.
5. **No** double shadows — one `--shadow-card` on hover, no second.
6. **No** rectangular focus halos — rounded pollen + synapse rings.
7. **No** press-scale-only — filter brightness + 0.5px settle, per
   MOAT §2.2.
8. **No** inlined err or empty or loading blocks — three components
   own the three states.
9. **No** second publisher — route to the existing publish flow.
10. **No** "Install!" celebration — the card settles, a thread draws,
    and the dot lights. That's it.
11. **Yes** to the keyboard grid — Tab cycles, arrows move, ⌘I / ⌘U /
    ⌘Enter shortcut the action, Esc closes, Home/End jump.
12. **Yes** to the unified install view — one grid, one filter, one
    install pill.
13. **Yes** to the update dot — pollen, breathing, slot-swapped with
    the synapse installed dot.
14. **Yes** to reduced-motion global — the route declares no
    `prefers-reduced-motion` blocks.

If during Phase 4 you find a place where the spec is silent and the
default would be wrong, the answer is to ask before inventing. The
MOAT bar is "premium-quality." Anything that fails it is competent-
but-generic, and competent-but-generic is how a year gets lost.
