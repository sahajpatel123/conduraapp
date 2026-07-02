# SCREEN — Hub · `Hub.svelte` · `#/hub`

> **Status:** Phase-2 architecture spec. Phase 4 implements against this.
> **Contract:** `MOAT.md` (premium bar), `APPFLOW.md` §4.3 (current Hub),
> `DIRECTION.md` (the voice). The current `Hub.svelte` (650 lines) is a
> 3D bookshelf with `perspective: 800px`, `transform: rotateY(-22deg)`,
> `rotateX`-style hover, and inline overlay/sheet plumbing. `MOAT.md §1.5`
> names the rotateX pattern by file/line; this spec deletes the entire
> 3D vocabulary and replaces it with a flat featured-shelf + grid.

---

## 1. Inheritance

The spec assumes the following are already in place; this spec does
**not** redefine them.

| From | What this spec uses as-is |
|---|---|
| `MOAT.md §2.3` | `prefers-reduced-motion` is respected via one global rule in `condura.css`. The component declares no media-query blocks. |
| `MOAT.md §2.4` | Empty states are three lines: **what** / **why empty** / **next action.** The empty state in this spec follows that shape exactly. |
| `MOAT.md §2.6` | One `ErrorState` component owns all error rendering. The route uses it. No inline err blocks. |
| `MOAT.md §2.7` | `.tactile` global class owns press/hover transition timing. The component declares no per-card `transition:` lists. |
| `MOAT.md §2.8` | Publish flow is a `.c-sheet` (slides from bottom — a **task**, not a confirmation). Detail panel is a `.c-sheet`. Filter rail popover is a `.c-popover`. Loading state is a `Thread` draw. |
| `MOAT.md §2.9` | Icon buttons get `<Tooltip label>` — no `title=` attributes. |
| `MOAT.md §3` | "Finished" states draw a `Thread` left-to-right (subscribe settled, publish settled, search resolved). |
| `MOAT.md §4 #1` | No gradient text. The featured shelf uses solid `synapse` / `pollen` / trust-color dots. |
| `MOAT.md §4 #2` | All icons go through `<Glyph name="…" />`. No emoji. No Unicode-as-icon. |
| `MOAT.md §4 #3` | No `backdrop-filter: blur()` on this route. Detail sheet is paper-on-paper. Publish sheet is paper-on-paper. |
| `MOAT.md §4 #4` | Status colors are `--ok` / `--warn` / `--danger` / `--info`. Trust badges use synapse-filled / synapse-outlined / pollen-filled. No purple, no teal. |
| `MOAT.md §4 #5` | No "Welcome to the future" copy. The eyebrow + title read like a museum label. |
| `MOAT.md §4 #6` | No "Subscribed!" celebration. The card settles, a thread draws, a synapse dot lights. |
| `MOAT.md §4 #7` | No spinner. Loading is a `Thread` drawing left-to-right (`drawthread`, `--dur-slow`). |
| `MOAT.md §4 #8` | Focus halos are rounded (pollen halo tracks the radius). Pill elements use the 2px synapse ring + 5px halo pattern. |
| `MOAT.md §4 #9` | One elevation token per surface. Cards use `--shadow-paper` at rest and `--shadow-card` at hover. The publish sheet uses `--shadow-float`. |
| `MOAT.md §4 #10` | Every animation answers: *what is this communicating?* Entrance = data ready; hover = takes focus; subscribe flip = state change; thread-draw = completed. |
| `DIRECTION.md §5` | Motion grammar uses `--dur` / `--dur-slow`, `--ease` only. The Hero `synapse-pulse` is `pollen-breath` (1.6s). The publish sheet slides from bottom (`translateY(24px) → 0`). |
| `DIRECTION.md §3` | Type scale is read from tokens (`--text-display`, `--text-h1`, `--text-lead`, `--text-caption`, `--text-micro`). |
| `DIRECTION.md §4` | Color roles via tokens (`--content*`, `--synapse*`, `--pollen*`, `--ok`, `--warn`). The brand has synapse + pollen only. |
| `APPFLOW.md §4.3` | The existing Hub IPC contract is followed (`hub.search` debounced 250ms). The route IDs in §7 below are an *addition* — the route shipped a broad `'skill'` priming query; this spec adds a featured-shelf read on mount so a curated highlight lives above the grid from the first paint. |
| `SCREEN_SKILLS.md` | Card grammar (the 5-row anatomy, trust badge, action row, install flip, settled thread) is reused. The Hub surface is the **public side** of the same shape — no second card grammar. |

---

## 2. Layout & Content

### 2.1 Page-level structure

The Hub route renders inside the shell's main surface, right of the
`NavRail` (which is 232px wide per `APPFLOW.md §3.1`).

```
┌────────────────── Hub route (max-width 1280px, padded) ──────────────────┐
│                                                                        │
│  ── The library · curated · safety-scanned                              │  eyebrow
│                                                                        │
│  Skills for the things you ask.                                        │  title (display, clamp 28→40)
│  A shelf of procedures — auto-created from your complex tasks, or     │  sub (sans 16, max 56ch)
│  curated by the community. Install draws a thread into your machine.   │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ HeroShelf — featured (max 3)            ┌──────────────────────┐  │  │  Featured row (12 viewport, synapse-pulse)
│  │ ┌────────┐ ┌────────┐ ┌────────┐       │  Publish a skill →     │  │  Publish CTA (pollen pill, top-right)
│  │ │ feat 1 │ │ feat 2 │ │ feat 3 │       │  PublishSheet trigger  │  │
│  │ └────────┘ └────────┘ └────────┘       └──────────────────────┘  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                        │
│  ┌────────────────┐  ┌──────────────────────────────────────────────┐ │
│  │                │  │                                              │ │
│  │  Filter rail   │  │   Grid (auto-fill, 1–3 cols, 280px min)    │ │
│  │  (240px,       │  │                                              │ │
│  │   sticky       │  │   SkillCard x N (the SCREEN_SKILLS grammar)  │ │
│  │   on scroll)   │  │                                              │ │
│  │                │  │                                              │ │
│  │  Categories    │  │                                              │ │
│  │  Trending      │  │                                              │ │
│  │  Trust levels  │  │                                              │ │
│  │  My subs       │  │                                              │ │
│  │                │  │                                              │ │
│  └────────────────┘  └──────────────────────────────────────────────┘ │
│                                                  ┌────────────────────┐│
│                                                  │   Detail panel     ││
│                                                  │   (.c-sheet, 360px)││
│                                                  │   on card click    ││
│                                                  │                    ││
│                                                  └────────────────────┘│
│                                                                        │
│  ── Footer band ──  hairline + small mono link row                      │
└────────────────────────────────────────────────────────────────────────┘
```

The page total max-width is 1280px. Above 1280px viewport, content
stays centered. Below 960px viewport, the filter rail collapses to a
`[⛁]` button that opens a `.c-popover` sheet. Below 720px viewport,
the detail panel becomes a full-screen `.c-sheet`. Below 480px, the
grid goes single-column and `min-width: 0` lets descriptions truncate.

### 2.2 Region A — HeroShelf (featured, max 3)

The HeroShelf is **the public face of the route.** It is one row of
**at most three** featured skill cards. "At most three" is the
non-negotiable MOAT rule (see §10 — "Featured must stay curated, not
crowded"). On mount, the route reads `hub.featured(max=3)`; the
shelf renders the cards in row-major order. If fewer than three exist
(empty or sparse week), the shelf shows what it has and the grid fills
the rest of the page honestly.

#### A.1 Card anatomy (HeroShelf variant of `SkillCard`)

Same 5-row grammar as `SCREEN_SKILLS §3.2`, with three additions:

- **Synapse-pulse dot** in the top-right corner — a 6×6 dot in
  `--synapse` with `pollen-breath` 1.6s loop (the same breath used in
  the Skills "update available" state; here it signals "the curator's
  pick this week").
- **Eyebrow override:** the card eyebrow reads `— this week's pick`
  (mono caps 10, `--synapse`), replacing the trust-level eyebrow.
- **Card weight:** the featured card uses `--shadow-card` at rest
  (the regular grid card uses `--shadow-paper` at rest). Featured
  reads as elevated, not loud.

#### A.2 HeroShell layout

```
┌─ featured ─ 12 col grid, 24px gap ───────────────────┐
│                                                       │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐          │
│  │  feat 1   │  │  feat 2   │  │  feat 3   │          │
│  │  weekly   │  │  weekly   │  │  weekly   │          │
│  │  synapse  │  │  synapse  │  │  synapse  │          │
│  │  pulse    │  │  pulse    │  │  pulse    │          │
│  └───────────┘  └───────────┘  └───────────┘          │
│                                                       │
└───────────────────────────────────────────────────────┘
```

Each card is `minmax(0, 1fr)`. Below 960px the shelf collapses to a
horizontal scroll (`overflow-x: auto`, snap to each card) — the
featured row never wraps. Below 480px the shelf collapses to one
column, one featured card per scroll-snap.

### 2.3 Region B — Publish CTA (top-right of the HeroShelf row)

A single `<Button variant="primary">Publish a skill →</Button>` sits
at the right end of the HeroShelf row, vertically centered. The
button is always enabled — pressing it opens the **PublishSheet**
(see §3.4 below), which is a `.c-sheet` sliding up from the bottom
(MOAT §2.8: a sheet, not a modal — publishing is a *task*, not a
*confirmation*).

The CTA changes appearance based on `account.isSignedIn`:

- **Signed-in:** primary pollen pill, full label "Publish a skill →".
- **Signed-out:** ghost pill, label "Sign in to publish a skill →".
  Clicking routes the user to the existing magic-link sign-in
  (`account.signInWithEmail`) and then re-opens the sheet on success.

### 2.4 Region C — Filter rail (240px, sticky on scroll)

A `.c-paper` vertical column, 16px from the grid. Same landmark and
keyboard surface pattern as `SCREEN_SKILLS §3.2`. It contains,
top-to-bottom:

#### C.1 "Categories" group

A stacked list of category chips. Categories are the union of
`tags` from the current featured + grid results, deduped, top-8 by
frequency. Click a chip → filters the grid (just like the Skills
filter rail; the Hub and Skills share filter vocabulary).

A separator hairline (`<Thread />`).

#### C.2 "Trending" group

A second stacked list — currently top-5 most-downloaded this week
(same data source as `hub.featured`, sorted by `downloads_7d desc`).
Clicking a trending tag pins it as a chip filter.

#### C.3 "Trust" group

Same three honest trust levels as `SCREEN_SKILLS §3.2 A.2`. Each row
is a stacked checkbox `[☐] Label (count)` with the same color and
geometry conventions (synapse filled / outlined / pollen filled).

#### C.4 "My subscriptions" group

The fifth and final group — a list of skills the user has already
subscribed to. The list is client-side (read from the
`hub.subscribed` Set, populated from `hub.subscribedList()` on
mount). Click a subscribed skill → opens the detail panel with an
"Uninstall" action in the primary slot.

If empty, the row reads `— nothing subscribed yet` in mono 11
`--content-faint`. The rail is never empty in a way that confuses
the user.

At the bottom of the rail: a single line, monospaced, 11px, faint:
**Filters update as you check.** A `Reset filters` link in mono-pollen
appears below it whenever any filter is active.

### 2.5 Region D — Grid

The grid uses the **same flat `SkillCard` grammar as `SCREEN_SKILLS`**
— one source of truth, no second card type. Reuse the
`SCREEN_SKILLS §3.2 B.x` rows 1–5 unchanged.

`display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));`
**No `perspective`. No `transform-style: preserve-3d`. No `rotateX`.
No `rotateY`.** This is the explicit `MOAT.md §1.5` / `APPFLOW.md
§4.3` deletion; the grid is a clean flat plane (the same deletion
already applied to the Skills page).

`gap: var(--space-4)` (16px). `align-items: stretch`. The grid sits
inside a `.grid-wrap` that has `padding: 0 var(--space-4) var(--space-9)`.

The grid renders inside a `{#key gridLoadId}` block so that a manual
refresh re-mounts the cards and the entrance stagger fires again.

### 2.6 Region E — Detail panel

A `.c-sheet` that slides in from the right (per MOAT §2.8 — no scrim
blocking the rest of the route). **Same block list as
`SCREEN_SKILLS §3.4`** with one substitution:

- The primary action row swaps `Install →` for `Subscribe →`
  (signed-in, not subscribed) / `Subscribed ✓` (signed-in, subscribed)
  / `Sign in to subscribe →` (signed-out). On subscribe settled, a
  `Thread` draws across the bottom edge of the card in the grid (the
  same `thread-draw-left-right` gesture as Skills install settled).
  The detail panel's own bottom-button label morphs in place
  (no celebration, per MOAT §4 #6).

The detail panel renders inside a `{#key detail?.id}` block so that
switching between cards re-mounts and replays the slide-in.

### 2.7 The footer band

A single hairline + a mono row at the bottom of the route body,
max-width 1280px:

> `The Hub is curated. Skills are safety-scanned before they land
> here. Subscribed skills land in your local shelf under
> "Skills" (the same thing, your side). Last index
> refresh · 12 min ago.`

The "Last index refresh" label re-renders when `hub.search` /
`hub.featured` last succeeded; on error it goes red and reads
`Last refresh failed at {HH:MM} · check connection`.

---

## 3. State Matrix

These are the visual states the route can be in. They are not
mutually exclusive — the HeroShelf has its own state, the Grid has
its own, the Detail Panel has its own selected-with-card state, and
the PublishSheet has a mounted-with-task state.

| # | State | What you see | Trigger / data condition |
|---|---|---|---|
| **S1** | **Empty (Hub just waking up)** | Empty state — see §3.1 below. | `hub.loading && hub.results.length === 0 && hub.featured.length === 0` AND no `hub.error`. This is the "waking up" empty state, distinct from "shelf is quiet" (which means the Hub is indexed but quiet right now). |
| **S2** | **Loading (featured + grid)** | A `Thread` draws left-to-right across the **entire** route body (`drawthread`, `--dur-slow`, `--ease`) — never a row of skeleton cards. The hero and the toolbar remain interactive. | `hub.loading && hub.featured.length === 0 && hub.results.length === 0` AND no `hub.error`. The HeroShelf shows a single full-width skeleton row; the grid hides entirely until the first result. |
| **S3** | **Subscribed (settled)** | The card stays in the grid. A filled `--synapse` 6×6 dot sits in the top-right corner (the "subscribed" mark — same slot as Skills' "behind the glass" mark). The detail panel's primary action reads `Subscribed ✓`; clicking opens an "Unsubscribe" confirmation popover. A `Thread` draws across the card's bottom edge (`thread-draw-left-right`, 520ms `--ease`). | `hub.subscribe(id)` returns `{ok: true}`. The Hub subscribes by installing the skill locally — the user's `#/skills` view shows it as `installed` on the next poll. |
| **S4** | **Publishing (publish sheet mounted)** | The publish sheet slides up from the bottom (MOAT §2.8: a task, not a confirmation). The route body stays live behind it. The sheet's submit row shows a `Pulse phase="acting"` next to the form. | User clicked the "Publish a skill →" CTA in the HeroShelf right slot. State holds until `hub.publish(...)` resolves or rejects. |
| **S5** | **Error** | An `ErrorState` component renders once at the route level (NOT inline in the grid). | `hub.error` (search, featured, subscribe, unsubscribe, publish). The route stays navigable; the search/featured/grid/show-error pattern is unchanged from the other routes. |
| **S6** | **Not signed in** | The Publish CTA label is "Sign in to publish a skill →" (ghost pill). Clicking routes to the magic-link sign-in flow. The rest of the route is fully usable without an account — local subscribed skills still work, search still works, detail still works. | `account.isSignedIn === false`. |
| **S7** | **Subscribing (one card mid-flip)** | The picked card flips Y over 480ms to a "subscribing face" (see §5.4 — same flip grammar as `SCREEN_SKILLS §5.4`). The back face shows `Subscribing…` with a `Pulse phase="acting"`. The card is non-interactive during the flip. | User pressed Subscribe on a not-subscribed card OR pressed Unsubscribe on a subscribed card (the flip reverses, no progress face). |

### 3.1 The Empty state (S1) — exact copy

Three lines, exactly per MOAT §2.4:

> **What's here.** A shelf of procedures the agent can re-run —
> recipes for "summarize this PDF," "morning briefing," "diff two
> PRs." The Hub is the curated, safety-scanned collection; what
> you find here is what you can subscribe to.
>
> **Why it's empty.** The Hub is waking up. We're fetching the
> shelf's curated set, indexed by community, safety-scanned.
>
> **Next action.** Wait a moment. The featured shelf will land as
> soon as the index returns. (No button — the empty state owns
> patience, not action.)

Beneath the three lines, a single full-bleed `Thread` drawing
left-to-right signals that data is on its way (the same gesture as
the loading state — they're the same visual moment, just labeled
differently depending on whether the index has returned at least
once).

### 3.2 The Loading state (S2) — exact copy

> **A thread draws across the page** (left-to-right, 1px, `--synapse`,
> `--dur-slow`, `--ease`). The eyebrow under it reads, in mono caps
> 11, `--content-faint`: `INDEXING THE SHELF · /hub`. Once the
> featured read resolves, the thread arrives 100% of the way across
> the page width and the HeroShelf fades in beneath it (single
> 0→1 opacity over 480ms `--ease`, no stagger — three cards max, no
> need to stagger). Once the search read resolves, the grid fades
> in beneath, staggered (per §5.1).

The route does not show partial cards while loading. If both calls
return in <180ms the user never sees the loading state at all. The
grid is invisible until the first `hub.search` resolves.

### 3.3 The Subscribed state (S3) — exact copy

After S7's success path, the card resolves to:

- Trust badge: unchanged.
- Top-right dot: 6×6 filled `--synapse` (the "subscribed" mark —
  same slot and color as the Skills "behind the glass" mark).
- Action button: `Subscribed ✓` (synapse-filled pill, clears to
  pollen outline on hover per MOAT §2.7). Clicking opens the
  detail panel's "Unsubscribe" popover (not a DELETE primary —
  unsubscribe is a network WRITE, not a destructive, so the action
  is a small ghost confirmation).
- Thread: a 1px synapse line draws across the card's bottom edge
  left-to-right over 520ms (`drawthread` keyframe, `--dur-slow`,
  `--ease`). This IS the moment of "completed" per MOAT §3.

No toast. No "Subscribed!" celebration. No animation other than the
thread-draw. Per MOAT §4 #6.

### 3.4 The Publishing state (S4) — exact copy

The publish sheet slides up from the bottom edge of the route (`translateY(24px) → 0`,
`opacity 0 → 1`, `--dur-slow`, `--ease`). The sheet is paper-on-paper
(no scrim blur; MOAT §4 #3). The route body stays live.

The sheet contains (no per-row contract change from the existing
`PublishModal.svelte` — this spec migrates it from a centered modal
to a `.c-sheet` per MOAT §2.8):

| Block | What |
|---|---|
| **Sheet header** | Close `×` button (top-right, 32×32 hit area); title (display 28, `--content`); eyebrow mono caps: "— Publish to the public Hub". |
| **Name field** | One `<input type="text">` with placeholder "Skill name" + the same `::before` thread-draw on focus that `SCREEN_SKILLS §5.2` uses (the universal "I am writing here" gesture). |
| **Version field** | One `<input type="text">` with placeholder "v0.1.0" + a mono-11 helper line "semver — major.minor.patch". |
| **Description** | A 4-line `<textarea>` with placeholder "What does this skill do?" in italic display. |
| **Author field** | Pre-filled with the user's account email (read-only display + "Change in Settings →"). |
| **License chips** | Three radio chips: "MIT", "Apache-2.0", "Freeware". |
| **Tag chips** | Multi-select chip input — typed tags, comma or `↵` to commit, backspace to remove. |
| **Archive picker** | Single `<input type="file" accept=".zip">` styled as a ghost pill. Below, a mono-11 helper: "≤32 MB · .zip · one skill per archive". |
| **Submit row** | Sticky to the bottom of the sheet: one `<Button variant="primary">Publish →</Button>`. Disabled until all required fields are filled. To the right of it, a `.btn-ghost` `Cancel` (Esc) and a mono 11 link `[What gets scanned?]`. |

The sheet's submit is `hub.publish(payload)` → server-side
safety-scans the archive against the promptware rules → returns
`{ok: true, url: 'hub.condura.app/skills/<id>'}` (the canonical URL
on the Hub).

On success: the sheet slides out to the bottom (reverse `--dur`),
a `Thread` draws across the route's footer hairline (per MOAT §3
— the moment of completion), the new skill appears in the grid on
the next poll, and the user sees a mono 11 link at the top of the
HeroShelf row: `Your skill is live → <url>`.

On failure: the sheet stays open, the submit button enables back,
a route-level `ErrorState` renders beneath the sheet (never inside —
the user needs to see the rest of the route too), and the mono 11
helper above the file picker reads `Safety scan did not pass: {reason}`
with `reason` populated from the daemon's response.

### 3.5 The Error state (S5) — exact copy

The route uses one `ErrorState` component (per MOAT §2.6). It
renders once per route, anchored above the HeroShelf.

> **We couldn't reach the Hub.** (italic display 22, `--content`.)
>
> Cause: *`{noun}`*, e.g.
> - "Featured list returned no rows from `ipc.hub.featured`."
> - "Search timed out after 2.4s."
> - "Subscribe failed: `404 — skill not found on hub.condura.app`."
>
> Likely reason: *`{phrase}`*, e.g.
> - "Hub is being re-indexed."
> - "Your network is offline."
> - "This skill was unpublished since you last saw it."
>
> Next action: `[Try again]` pill — pollen-outline button. Above it
> on the right, a mono 11 link `[Open Settings → Connection]`.
>
> *(Below all of this, an `err-hair` rule, 1px, left→right draw over
> 600ms `--ease`.)*

### 3.6 The Not-signed-in state (S6) — exact copy

The Publish CTA label is "Sign in to publish a skill →" (ghost pill).
Clicking routes the user through `account.signInWithEmail(email)` →
on success, `account.isSignedIn === true` flips, the CTA label
morphs to "Publish a skill →" (primary pill), and the PublishSheet
auto-opens.

If the user signs in via Settings → Account while the Hub is
mounted, the CTA label morphs in place — no route reload.

### 3.7 The Subscribing state (S7) — exact copy

The chosen card flips on Y. Front face and back face both render;
CSS hides the wrong one with `backface-visibility: hidden`.

Front face (visible at rest, fades to 0°→90° over 480ms `--ease`):
the regular card (rows 1–5).

Back face (visible at 90°→360°, same 480ms `--ease`, mirrored):
the card progress face shows `Subscribing…` (display 18), a
`Pulse phase="acting"` (12px), and `v0.2.1 · {author}` (mono 11
faint).

The pulse breathes at 1s. The card is non-interactive until the
flip back; clicks on it during flip are swallowed (no
double-subscribe).

**On success:** the front face rotates past 90° → 360°, the
`Subscribed ✓` synapse check-mark draws itself onto the front-face
button (stroke-dashoffset 1 → 0 over 320ms `--ease`), and the
bottom-edge thread draws over 520ms `--ease`.

**On failure:** the card returns to its starting state, the trust
badge swaps to `<Glyph name="warning" />` filled `--danger` for
600ms, the action button reads `Try again`. An `ErrorState` panel
at the route level explains the cause.

---

## 4. Motion Choreography

The route's motion follows one rule: **every animation answers
"what is this communicating?"** (MOAT §4 #10). Decorative loops
are forbidden.

### 4.1 Entrance — HeroShelf 0→1, cards fade in with a 40ms stagger

Trigger: `hub.featured` first resolves → HeroShelf fades in
(opacity 0 → 1, no transform, 480ms `--ease`). Once `hub.search`
first resolves → the grid's `{#key gridLoadId}` re-mounts and the
cards fade in with the standard 40ms stagger (per `SCREEN_SKILLS §5.1`).

**Reduced-motion:** the entire 40ms-stagger sequence is replaced
with a single 0ms stagger — all cards appear at once. (Per MOAT §2.3,
one global rule in `condura.css` does this.)

### 4.2 Hover — card lifts 2px, hairline becomes synapse-strong

Identical to `SCREEN_SKILLS §5.2`.

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
var(--shadow-card), 0 0 0 5px var(--pollen-halo-color), 0 0 0 2px
var(--synapse)`.

**Hard delete:** the current `Hub.svelte:372-379`
`transform-style: preserve-3d` + `translateY(-6px)` on hover +
`translateY(-3px) scale(0.97)` on active. The `Hub.svelte:388-407`
`transform: rotateY(-22deg)` on the spine face +
`Hub.svelte:345-347` `perspective: 800px`. The
`Hub.svelte:413-415` `transform: rotateY(-30deg)` on `:hover`.
Deleted in full per MOAT §1.5 and APPFLOW §4.3.

**Reduced-motion:** `transform: none`; the card glows its focus halo
and goes no further.

### 4.3 Detail panel slide-in (right)

Identical to `SCREEN_SKILLS §5.3`.

The grid's `padding-right` increases by 360px (`grid-wrap` style
with CSS transition) over 320ms `--ease`. The detail panel slides
in from `right: -360px → right: 0` over 320ms `--ease`. (Uses
`transform: translateX(...)`, not `left:` — transforms are
GPU-compositable, smoother on lower-end devices.)

**Reduced-motion:** the grid padding and panel both move instantly
with no slide.

### 4.4 Hero synapse-pulse on featured cards

The synapse pulse is `animation: pollen-breath 1.6s var(--ease) infinite`,
applied to the 6×6 synapse dot in the top-right corner of each
featured card.

```css
@keyframes pollen-breath {
  0%, 100% { opacity: 0.7; transform: scale(1); }
  50%      { opacity: 1.0; transform: scale(1.15); }
}
```

The breath is the only idle loop on the route, and it answers "What
is this communicating? *The curator picked this one.*" (MOAT §4 #10.)

**Reduced-motion:** the breath becomes a static 100% opacity dot in
synapse. No scale change.

### 4.5 Search cross-fade 200ms

When the search input debounces (`hub.search(q)` fires), the grid
fades the old results out over 120ms (`opacity 1 → 0`, `--ease`),
the new results fade in over 120ms (`opacity 0 → 1`, `--ease`)
after the IPC resolves. Total cross-fade: 240ms. The Thread does
not draw here — search-as-cross-fade is a quieter gesture than the
install-settled thread.

**Reduced-motion:** instant transition. No fade.

### 4.6 Install — card flips on Y (S7)

Trigger: user pressed Subscribe on a not-subscribed card.

Identical to `SCREEN_SKILLS §5.4`:

- **Stage A** — front face fades to 90° over 480ms `--ease`
  (`card-flip-out`).
- **Stage B** — back face visible from 90° → 360° (same 480ms
  `--ease`, mirrored, `card-flip-in`). The Pulse on the back face
  starts at the beginning of stage B and breathes at 1s.
- **Stage C (success)** — `data-state` is removed; the card returns
  to its starting rotation; the action button swaps from
  `Subscribing…` to a `<Glyph name="check" />` drawing itself
  (stroke-dashoffset 1 → 0 over 320ms `--ease`), then morphs to
  `Subscribed ✓`; the bottom-edge thread draws over 520ms `--ease`
  (the "completion" gesture per MOAT §3).
- **Stage C (failure)** — `data-state` returns to its previous
  value; the card flips back to rotation 0 over 320ms `--ease`;
  the trust badge swaps to `<Glyph name="warning" />` filled
  `--danger` for 600ms; the action button reads `Try again`. A
  route-level `ErrorState` renders above the HeroShelf (S5).

**Reduced-motion:** Stages A and B are skipped entirely. The card's
action button changes in place. The thread-draw still happens. The
check glyph drawing is also skipped (the check just appears).

### 4.7 PublishModal slides up from bottom (S4)

Trigger: user clicked Publish a skill.

The publish sheet is `.c-sheet`. It slides up from the bottom edge:

```
transform: translateY(24px) → 0
opacity: 0 → 1
duration: --dur-slow (520ms)
easing: --ease
```

The route body stays live behind it (no scrim blur; MOAT §4 #3).
Esc closes the sheet.

**Reduced-motion:** instant appearance. No slide.

### 4.8 The Thread (MOAT §3) — when it draws

The thread draws on **four** moments only:

1. **Subscribed settled (S3 success)** — bottom edge of the card.
2. **Search resolved** — across the top of the grid (a 1px segment,
   not a full-width line; a quiet "the shelf has been re-indexed"
   signal).
3. **Publish settled (S4 success)** — across the route's footer
   hairline.
4. **Featured first-arrives** — across the HeroShelf row's top edge
   (a single hairline, left→right over 520ms `--ease`).

`err-hair` is unchanged: when an error resolves, the existing
hair-line draws left→right. No new thread uses.

---

## 5. Keyboard

The route's keyboard surface is added to the global shortcuts
documented in MOAT §2.10. Keys are bound at the route level
(`onMount` registers them, `onDestroy` removes them).

### 5.1 Global keybindings

| Key | Action |
|---|---|
| `/` | Focus the search input. If the user is already inside a text input, `/` types a literal `/`. |
| `Esc` | Closes the topmost overlay/panel — first the publish sheet, then the detail panel, then any open `.c-popover`, then clears the search input. If all four are already closed, Esc is a no-op (does not navigate). |
| `⌘K` | Opens the command palette (already global). |
| `⌘P` | Opens the PublishSheet. (Conflicts with Print? No — the global handler from MOAT §2.10 already routes `⌘P` to the publish flow.) |
| `⌘,` | Opens Settings (already global per MOAT §2.10). |

### 5.2 Grid navigation

Same as `SCREEN_SKILLS §6.2` (Tab cycles cards, arrow keys move
row-major, Home/End jump, PageUp/PageDown scroll).

### 5.3 Card-level keybindings (when a card is focused)

| Key | Action |
|---|---|
| `⌘S` | Subscribe to the focused card (only enabled when card state is "not subscribed"). Mirrors the "Subscribe →" primary action. |
| `⌘U` | Unsubscribe the focused card (only enabled when card state is "subscribed"). Mirrors the "Unsubscribe" ghost action. |
| `⌘Enter` | Open in Skills — equivalent to clicking the primary action button when subscribed; subscribe-then-open when not subscribed. (This is the discovery shortcut: subscribe to the skill and immediately send a `/run <skill>` chat message.) |

### 5.4 Detail panel keybindings

| Key | Action |
|---|---|
| `Tab` | Cycles through detail panel interactive elements (top to bottom). |
| `Esc` | Closes the detail panel and returns focus to the originating card. |
| `⌘C` | With the focus on the usage-example code block, copies the example to clipboard. |
| `⌘S` / `⌘U` | Same as card-level keybindings; they trigger the primary action row in the detail panel. |
| `⌘Enter` | "Open in Skills" primary action (the sticky-bottom button). |

### 5.5 PublishSheet keybindings (open)

| Key | Action |
|---|---|
| `Tab` / `Shift+Tab` | Cycles through the sheet's fields, top to bottom. |
| `Esc` | Closes the sheet and returns focus to the Publish CTA. |
| `⌘Enter` | Submits the publish payload. Equivalent to clicking `Publish →`. |

### 5.6 Filter rail keybindings

Identical to `SCREEN_SKILLS §6.5` — Space/Enter toggle chips,
`⌘F` opens the rail as a `.c-popover` sheet at narrow viewports.

### 5.7 Reduced-motion note

All keybindings above are unchanged when the user has `prefers-
reduced-motion: reduce` set. Animations change; the keys don't.

### 5.8 Focus visibility

Every card, every pill, every chip, every field, every publish-sheet
input carries a rounded halo on `:focus-visible` (per MOAT §2.1). No
`outline: 1px solid var(--content)`. No rectangular halos.

---

## 6. Components Used

The route is composed of these components. The full prop contracts
below belong in each component's source file — this spec is the
short-form.

### 6.1 `<HeroShelf />`

Sits in `app/web/frontend/src/lib/condura/HeroShelf.svelte`.

```ts
let {
  skills: HubSkillMeta[],       // at most 3
  onselect: (id: string) => void,
  onpublish: () => void,
  isSignedIn: boolean,
  class: cls = '',
} = $props();
```

Owns:
- The featured row layout (max 3, vertical-centered to Publish CTA).
- The hero synapse-pulse on each featured card (4s breath, synapse
  color, 6×6 dot in top-right).
- The Publish CTA's two-state appearance (primary if signed-in,
  ghost if signed-out).
- Its own loading skeleton (single full-width placeholder, never
  skeleton cards).

Does **not** own:
- The card body itself — that's `<SkillCard />` (shared with the
  Skills page).
- The data fetching — `Hub.svelte` owns `hub.featured(max=3)` and
  passes the array down.

### 6.2 `<SkillCard />` (public variant)

Reuses the same component as `SCREEN_SKILLS §7.1` with one extra
state (`subscribed`) and one extra prop (`featured: boolean`):

```ts
type CardState =
  | { kind: 'not-subscribed' }
  | { kind: 'subscribed' }
  | { kind: 'subscribing' }
  | { kind: 'error'; reason: string };

let {
  skill: HubSkillMeta,
  state: CardState,
  featured = false,        // hero variant — heavier shadow + weekly eyebrow
  selected = false,
  onselect: (id: string) => void,
  onsubscribe: (id: string) => void,
  onunsubscribe: (id: string) => void,
  onopen: (id: string) => void,
  onmenu: (id: string, anchor: HTMLElement) => void,
  class: cls = '',
} = $props();
```

Owns the per-state visual transformation (front face vs back face;
the flip animation; the trust badge swap on error).

### 6.3 `<FilterRail />`

Reuses the same component as `SCREEN_SKILLS §7.2` (the same four
filter groups — View / Trust / Source / Category / Author — plus
the Hub's two unique groups: **Categories** and **Trending**, both
of which the same chip-input primitive owns). The component takes
a `groups: FilterGroupSpec[]` array and renders each one in order.

### 6.4 `<DetailPanel />`

Reuses the same component as `SCREEN_SKILLS §7.3`. The primary
action row swaps `Install` → `Subscribe` / `Sign in to subscribe`
based on `account.isSignedIn` and the card's `state.kind`.

### 6.5 `<PublishSheet />` (replaces existing `PublishModal.svelte`)

```ts
let {
  open: boolean,
  isSignedIn: boolean,
  onsubmit: (payload: PublishPayload) => Promise<void>,
  onclose: () => void,
  class: cls = '',
} = $props();

type PublishPayload = {
  name: string;
  version: string;          // semver
  description: string;
  author: string;           // pre-filled, read-only display
  license: 'mit' | 'apache-2.0' | 'freeware';
  tags: string[];
  archive: File;            // .zip, ≤32 MB
};
```

Owns:
- The `.c-sheet` slide-up from bottom (no scrim, MOAT §4 #3).
- All eight form fields per §3.4.
- The submit row (Publish / Cancel / "What gets scanned?").
- The sheet-level keyboard handlers (Esc closes; `⌘Enter` submits).

Does **not** own:
- The publish IPC call itself — `onSubmit` is a callback that
  `Hub.svelte` wires to `hub.publish(payload)`.

**Migration note:** the existing `PublishModal.svelte` (centered
modal) becomes a deprecated aliased re-export of this component,
for the duration of one release, to avoid breaking any other
caller. After one release, `PublishModal.svelte` is deleted.

### 6.6 `<Thread />` — already exists, reused

```ts
let { orientation: 'h' | 'v' = 'h', draw = true, glow = true, class?: string } = $props();
```

Used in:
- The grid's bottom-card-settled thread (S3 success).
- The search-resolved hairline at the top of the grid (S2 → S3).
- The publish-settled footer hairline (S4 success).
- The featured-first-arrives HeroShelf top-edge hairline.
- The `ErrorState`-rendered `err-hair` (the existing pattern).

### 6.7 `<Pulse />` — already exists, reused

```ts
let { phase: 'idle'|'thinking'|'awaiting'|'acting'|'consent'|'error'|'ok', size?: number, class?: string } = $props();
```

Used in:
- The loading state S2 (single 8px `idle` pulse next to the eyebrow).
- The subscribing back face (12px `acting` pulse, 1s breath).
- The settled-subscribed button (6px `ok` pulse, 4s breath — the rare
  "permission to be a loop" exception per MOAT §4 #10's narrow
  exception).

### 6.8 `<Glyph />` — already exists, reused

Same `name` table as `SCREEN_SKILLS §7.6` (`dot-active` for filled
trust badges, `dot` for outlined community, `chevron-right` for
forward-arrow CTAs, `close` for the sheet close, `warning` for
error state, `info` for the audit-trail info row in the detail
panel, `sync` for the footer). No new glyphs needed for this route.

If any icon on this table doesn't exist in `icons.ts`, the spec says
to **add it to `icons.ts`** (per MOAT §4 #2), not to use Unicode
or emoji as a shortcut.

### 6.9 `<Button />` — already exists, reused

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

- **primary** — Publish CTA (synapse-filled pill; clears to pollen
  outline on hover per MOAT §2.7).
- **secondary** — License chips in the publish sheet (paper-2 fill,
  hairline).
- **ghost** — "Try again", "Unsubscribe", "Sign in to publish a skill"
  (mono-uppercase link with a 1px underline that grows from the
  center on hover per MOAT §5.2).
- **danger** — Never; unsubscribe is a ghost because the action is
  reversible (the `hub.unsubscribe` RPC is a network WRITE, not
  destructive).

### 6.10 `<Tooltip />` — to be created per MOAT §2.9

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

- The "Publish a skill →" CTA (the tooltip explains "What gets
  scanned?", gated by the visible `[What gets scanned?]` link being
  off-screen on narrow viewports).
- The card menu `[⋯]` button (already covered by Skills).
- The detail panel close `×` ("Close").
- The detail panel "Copy example" button.
- All icon-only buttons in the toolbar's right cluster.

The tooltip is **never** used as a celebration; MOAT §4 #6 forbids
that.

### 6.11 `<ErrorState />` — to be extracted per MOAT §2.6

```ts
let {
  head: string,         // "We couldn't reach the Hub."
  cause: string,        // one noun
  reason: string,       // one phrase
  onretry: () => void,
  retryLabel?: string,  // defaults to "Try again"
  onsettings?: () => void,
  class?: string,
} = $props();
```

Used **once** per route, anchored above the HeroShelf.

### 6.12 `<EmptyState />` — to be extracted per MOAT §2.4

```ts
let {
  what: string,
  why: string,
  action: Snippet | null,
  sample: Snippet | null,
  class?: string,
} = $props();
```

The Hub uses the simpler three-line empty-state (per §3.1) — no
sample cards below, the empty state owns patience.

---

## 7. Data Fetched

The route's IPC contract is **eight** methods (the existing six
plus two new ones). All signatures come from
`/Users/sahajpatel/synaptic/app/web/frontend/src/lib/ipc/client.ts`
and `…/ipc/types.ts`.

### 7.1 `ipc.hubFeatured(max = 3)`

RPC: `hub.featured`. Returns `HubSkillMeta[]` (at most 3).

```ts
// New: client.ts (Hub namespace)
hubFeatured(max = 3): Promise<HubSkillMeta[]>
```

Cache policy: fetched on mount and on user-triggered refresh.
Stale-while-revalidate: re-fetched in the background every 300
seconds (5 minutes) while the route is mounted, never while the
page is hidden (`document.visibilityState !== 'visible'` skips
the poll). The user is never shown a loading state during a
stale refresh — the HeroShelf keeps showing the last good
featured set.

Error policy: failed calls fall through to the route-level
`ErrorState` (S5). The HeroShelf renders empty until the next
successful refresh; the grid is unaffected.

### 7.2 `ipc.hubSearch(query, limit = 20)`

RPC: `hub.search`. Returns `HubSearchResult { skills:
HubSkillMeta[]; total: number; query: string }`. Used to populate
the grid.

```ts
// Existing: client.ts:354
hubSearch(query: string, limit = 20): Promise<HubSearchResult>
```

Cache policy: debounced 250ms after the user types. The route uses
the query "skill" on mount (the broad priming pass, per
`Hub.svelte:46`).

Update detection: when a card exists locally (from `skills.list`,
read across the route bridge) AND `hub.search` returns a matching
`hub_id`, the card adopts the `subscribed` state — the Hub
treats subscribed-by-install as authoritative.

Error policy: failed calls land in `ErrorState` (S5).

### 7.3 `ipc.hubDetail(id)`

RPC: `hub.detail`. Returns `HubSkillDetail` (the full un-clamped
description, deps, version history, audit trail).

```ts
// New: client.ts (Hub namespace)
hubDetail(id: string): Promise<HubSkillDetail>
```

Used by the detail panel when the user clicks a card. May be
deferred if `hub.featured` / `hub.search` returns enough — flagged
here so Phase 4 doesn't make extra calls it doesn't need.

### 7.4 `ipc.hubSubscribe(id)`

RPC: `hub.subscribe`. Returns `{ ok: boolean; id: string }`.
Triggered by the "Subscribe →" pill (and `⌘S` shortcut) when the
card is not subscribed.

```ts
// New: client.ts (Hub namespace)
hubSubscribe(id: string): Promise<{ ok: boolean; id: string }>
```

Subscribing installs the skill locally — the same daemon flow as
the Skills route's `hub.install`. The user's `#/skills` view shows
the skill as `installed` on the next poll (no separate IPC for
cross-route notification).

**Gatekeeper consent:** if `hub.subscribe` would be classified as
a WRITE-by-network action, the existing `ConsentModal` opens
first (per the global pattern; nothing custom here).

### 7.5 `ipc.hubUnsubscribe(id)`

RPC: `hub.unsubscribe`. Returns `{ ok: boolean; id: string }`.
Triggered by the "Unsubscribe" ghost button in the detail panel
(and `⌘U` shortcut) when the card is subscribed.

```ts
// New: client.ts (Hub namespace)
hubUnsubscribe(id: string): Promise<{ ok: boolean; id: string }>
```

UX: the card stays in the grid for 400ms after the unsubscribe
resolves (because the optimistic state is set immediately), then
the bottom-edge synapse dot dims out and the action button reads
`Subscribe →` again.

### 7.6 `ipc.hubPublish(payload)`

RPC: `hub.publish`. Returns `HubPublishResult { ok: boolean;
url: string }`. Triggered by the PublishSheet's submit.

```ts
// Existing: client.ts (Hub namespace) — payload shape fixed by §6.5
hubPublish(payload: PublishPayload): Promise<HubPublishResult>
```

The daemon performs server-side safety-scan against the promptware
rules before returning `ok: true`. The `url` is the canonical Hub
URL (`hub.condura.app/skills/<id>`).

### 7.7 `ipc.hubSubscribedList()`

RPC: `hub.subscribed_list`. Returns `HubSkillMeta[]` (the user's
currently-subscribed skills, fetched fresh).

```ts
// New: client.ts (Hub namespace)
hubSubscribedList(): Promise<HubSkillMeta[]>
```

Cache policy: fetched on mount. Stale-while-revalidate on a 60-second
cadence while the route is mounted, never while the page is hidden.

### 7.8 `ipc.accountStatus()` — reused

RPC: `account.status`. Returns `{ isSignedIn: boolean; email?: string }`.
Used to drive the Publish CTA's two-state appearance (primary vs
ghost pill) and the gate on `hub.publish` (which requires
`isSignedIn === true`).

### 7.9 Replays — subscribe / publish events

The detail panel's audit trail uses no new IPCs. It composes from
the `audit.refresh()` events already in memory (publish and
subscribe events land on the HMAC chain via the existing
`audit.append()` calls in the daemon).

---

## 8. Design Decisions

These are the load-bearing calls — the place where this spec
disagrees with the current `Hub.svelte` and where it earns the
MOAT bar.

### 8.1 Public face — curated, not crowded

**The problem (per MOAT §1 — "RESTRAINT TEST").** A public Skills
Hub is the obvious place to brag: show all the categories, list
all 184 published skills, animate the whole thing with rotating
breathing dots. That's the "Welcome to the future" failure mode
the rules forbid.

**What this spec does.** The HeroShelf shows **at most three**
featured skills — a curated highlight, not a catalog dump. The
rest of the page is a flat filterable grid with the same card
grammar as the local Skills page. There is no carousel, no
autoplay, no "today's pick" rotating out, no badges on the
featured cards other than the synapse pulse (the breath itself
IS the "curator's pick" signal — three cards is enough; more than
three is the moment the curated shelf becomes a row).

The eyebrow reads "— The library · curated · safety-scanned" —
not "Trending this week!" — so the reader knows this is a
considered set, not a popularity metric.

### 8.2 Featured shelf max 3 — kill the popularity waterfall

**The problem.** A "trending" or "most downloaded" feed is a
popularity contest that punishes new skills and rewards the
already-popular. The Hub is a curated library, not a leaderboard.

**What this spec does.** The HeroShelf is `hub.featured(max=3)`,
served by a maintainer-curated list (not sorted by `downloads_7d`
desc). The "Trending" filter in the rail sorts the grid by
`downloads_7d` if the user wants to see one — but the hero is not
the trending feed.

### 8.3 Trust badges honest — kill the "Verified ✓"

**The problem.** The current `Hub.svelte:108` colors the trust
dot by `trust_level ?? trust`: `official` → `--synapse`,
`experimental` → `--hair-strong`, everything else → `--pollen`.
The `--hair-strong` color for experimental reads as "pending"
rather than "experimental" — a popular skill that hasn't been
classified yet looks the same as one that's flagged but not yet
removed. The badge conflates "new" with "unchecked."

**What this spec does.** Three trust levels, sourced exactly from
the data, no fourth tier:

| Level | Honest definition | Color |
|---|---|---|
| **Official** | Curated by Condura maintainers; signed against the org key. | `var(--synapse)` filled |
| **Community** | Published by a community author with ≥30 days on the Hub and ≥10 downloads. | `var(--synapse)` outlined (hairline + dot) |
| **Experimental** | New, unreviewed, or from an author without a community history. | `var(--pollen)` filled |

Each card carries the badge as a 6px dot at row 1 + a one-word
mono label. The user sees the level on every card, every time.
Per MOAT §4 #4 (no rainbow accents), the colors are `--synapse`
filled (official), `--synapse` outlined (community), `--pollen`
filled (experimental). No purple, no teal, no `--hair-strong`.

### 8.4 Publish flow is a sheet — not a modal

**The problem (per MOAT §1.5 / APPFLOW §4.3).** The current
`PublishModal.svelte` is a centered modal — a `role="dialog"
aria-modal="true"` with `backdrop-filter: blur(4px)` and a click-
to-close on the scrim. That's the wrong shape for publishing:
publishing is a *task* with multiple fields, a file picker, and a
submit — not a confirmation dialog. The modal shape is also the
exact `MOAT §4 #3` glassmorphism the rules forbid.

**What this spec does.** A `<PublishSheet />` — a `.c-sheet` that
slides up from the bottom (MOAT §2.8). The route body stays live
behind it (no scrim blur). Esc closes. The route never installs
glassmorphism on a publish flow. This migration is unconditional
— the existing `PublishModal.svelte` is deprecated and deleted
after one release.

### 8.5 Flatten the shelf — kill the 3D bookshelf and rotateX

**The problem (per MOAT §1.5).** The current `Hub.svelte` is a
3D bookshelf. Each skill is a slim vertical spine (36px wide,
200px tall) with `transform: rotateY(-22deg)` for the angled book
face. Hover tilts to `rotateY(-30deg)` and lifts 6px. MOAT §1.5
names `Skills.svelte:283`'s `rotateX(2deg)` — the same MOAT bar
applies to Hub's `rotateY` bookshelf (the rule's principle is
"no 3D tilt that doesn't read against a 3D surface"). There is no
3D surface in the Hub either; the spines have no varied Z.

**What this spec does.** All `transform-style: preserve-3d` and
`perspective` declarations are deleted. The `rotateY(-22deg)` on
the spine face is deleted. The `rotateY(-30deg)` on `:hover` is
deleted. The 6px lift on hover alone reads as elevation, exactly
as MOAT §1.5 prescribes for the Skills page. The grid is a clean
flat plane.

### 8.6 One card grammar — kill the second-card-type

**The problem.** The current `Hub.svelte` defines a `spine` card
type (36px wide, rotated spine, vertical-rl text) that has no
shared DNA with the Skills page's `card` type. Two card grammars,
two hover patterns, two per-state transformations. The user has
to learn two surfaces to do one thing.

**What this spec does.** The Hub and Skills pages share
**one** `<SkillCard />` component (and one `<FilterRail />` and
one `<DetailPanel />`). The Hub's "subscribed" state is
`<SkillCard />`'s state with `state.kind === 'subscribed'` — no
second card vocabulary. The two surfaces are one surface viewed
from two sides (the public side, the local side).

### 8.7 Not-signed-in is honest, not blocked

**The problem.** A public Skills Hub could trivially block all
action behind a sign-in wall. That's a hostile pattern (per
DIRECTION §1 "Configure, not comply") and it fails the "the
agent works without an account" promise from CLAUDE.md §3.

**What this spec does.** Signed-out users get:
- Full read access to the featured shelf and grid.
- The detail panel works (read-only).
- The Publish CTA reads "Sign in to publish a skill →" and routes
  through the magic-link flow — never blocks, always opens.
- The Subscribe action works (subscribe is local-first — the
  install RPC is signed-out-safe, the daemon just records the
  installation under the local user profile without sending the
  user identity to the Hub).

### 8.8 Loading states teach, not spin

**The problem (per MOAT §2.5).** The current `Hub.svelte:75-79`
loading state is `<Pulse phase="thinking" size=8 /> + "INDEXING
THE SHELF…"` — the single-breathing-dot pattern that says
"data might be moving" instead of "data is moving." MOAT §2.5
specifically calls this out and prescribes a thread-draw.

**What this spec does.** A single `Thread` drawing left-to-right
across the page (`drawthread`, `--dur-slow`, `--ease`) is the
loading state. The eyebrow mono label `INDEXING THE SHELF · /hub`
sits beneath the thread. The `Pulse` is moved to the toolbar next
to the search field, paired with the thread, never alone. Below
180ms the user never sees the loading state at all.

### 8.9 Errors live at the route level, not in the grid

**The problem (per MOAT §1.2).** The current `Hub.svelte:81-88`
inlines an `<div class="err-state" role="alert">` block —
declared verbatim across `Chat.svelte:213-221`, `Skills.svelte:71-81`,
`Channels.svelte:150-158`, and (per MOAT §1.2) is the **fifth
copy** of the same block.

**What this spec does.** One `<ErrorState />` component is used
**once** per route, anchored above the HeroShelf. The 240px rail
stays live so the user can filter the error away if they want.

### 8.10 The eyebrow and title read like a museum label

**The problem (per MOAT §4 #5).** "Discover amazing skills!" is
the failure mode the rules forbid.

**What this spec does.** The eyebrow reads `— The library ·
curated · safety-scanned` — three monosyllable fragments,
exactly the museum-label rhythm. The title reads `Skills for the
things you ask.` — a complete sentence in display italic. The
sub-copy reads `A shelf of procedures — auto-created from your
complex tasks, or curated by the community. Install draws a
thread into your machine.` — three sentences, each grounded in
fact, no "Welcome" energy.

---

## 9. Drift Table — what this spec changes vs. `Hub.svelte`

The implementation must apply every row below in one commit. Each
row cites the MOAT/APPFLOW finding that motivates the change.

### 9.1 Removed (per MOAT §1.5, §1.2, §4 #3)

| Current `Hub.svelte` line(s) | What it does | Why it goes |
|---|---|---|
| `Hub.svelte:342-347` | `.shelf { perspective: 800px; perspective-origin: 50% 60%; }` | MOAT §1.5 — named by analogy; no 3D surface reads against this perspective. Deleted. |
| `Hub.svelte:357-406` | `.spine { transform-style: preserve-3d; }` + `.spine-face { transform: rotateY(-22deg); transform-origin: left center; }` + the `:hover` companion `transform: rotateY(-30deg);` | MOAT §1.5 — the entire 3D bookshelf vocabulary. Deleted. The grid is flat. |
| `Hub.svelte:413-415` | `.spine:hover { transform: translateY(-6px); }` (the lift) is **kept** — only the `rotateY` is removed. The lift alone reads as elevation. | MOAT §1.5 — `translateY(-2px)` per the rule. The current value of `-6px` is overstated; spec says `-2px` per SCREEN_SKILLS §5.2. |
| `Hub.svelte:374-379` | `.spine:active { transform: translateY(-3px) scale(0.97); }` — three transforms, no weight. | MOAT §2.2 — replaced with the global `.tactile` rule (`translateY(0.5px) scale(0.985) filter: brightness(0.96) saturate(1.05)`). |
| `Hub.svelte:339-340` | `.shelf-stage { min-height: 320px; }` — leftover staging space for the spines after deletion. | The grid is the only thing left; min-height becomes irrelevant. |
| `Hub.svelte:67-71` | Tag chips derived from current results (`Array.from(new Set(hub.results.flatMap((s) => tagsFor(s)))).slice(0, 6)`). | Replaced by the `<FilterRail />` Categories group, which is a real filter, not a derived display. |
| `Hub.svelte:81-88` | Inline `<div class="err-state" role="alert">` with err-row / err-head / err-sub / err-hair all declared inline. | MOAT §1.2 — fifth copy of the same block. Extracted to `<ErrorState />`. |
| `Hub.svelte:439-501` | The overlay + sheet built inline. `<div class="detail-overlay">` with `backdrop-filter: blur(4px)` + click-to-close on the scrim + slide-in sheet on the right. | MOAT §4 #3 + §2.8 — deletes glassmorphism on the detail panel. Uses `.c-sheet` taxonomy. No scrim (`.c-sheet` does not block the rest of the route). |
| `Hub.svelte:294-334` | `@keyframes err-hair-draw` declared inline. | MOAT §2.6 — the `err-hair` token lives in `condura.css` (or wherever `<ErrorState />` declares it). One copy. |
| `Hub.svelte:486-501` | `@keyframes slide-in-right` + `@keyframes fade-in` declared inline. | Scoped to the deleted overlay. The `.c-sheet` primitive owns its own slide-in keyframe (in `condura.css`). |
| `Hub.svelte:439-498` | The detail sheet is owned by `Hub.svelte` itself. | `<DetailPanel />` is the single component — shared with `Skills.svelte`. |
| `Hub.svelte:489-491` | `.d-close { position: absolute; top: var(--space-4); right: var(--space-4); width: 32px; height: 32px; border-radius: 50%; … }` — the close-button styling is hand-rolled in the route. | MOAT §2.9 — the global `<Button variant="ghost" icon="close">` owns this. |
| `Hub.svelte:496-500` | `backdrop-filter: blur(4px)` on the detail overlay. | MOAT §4 #3 — explicit glassmorphism ban. |
| `Hub.svelte:636-649` | `@media (prefers-reduced-motion: reduce)` block declared inline (overrides for the shelf, the spine, the sheet, the overlay). | MOAT §2.3 — single global rule. The local block is deleted. |

### 9.2 Added (per MOAT §2.5, §2.6, §2.7, §2.8, §2.9, §3)

| What is added | Where | Why |
|---|---|---|
| `<HeroShelf />` with at most 3 featured cards. | New file `HeroShelf.svelte`. | §8.1 — public face, curated not crowded. |
| `<FilterRail />` with Categories / Trending / Trust / My subscriptions groups. | Reused from `SCREEN_SKILLS` (shared component); the route passes the four group specs via `groups` prop. | §8.6 — one filter vocabulary across Hub and Skills. |
| `<PublishSheet />` (`.c-sheet` sliding up from bottom). | New file `PublishSheet.svelte`, replacing the existing `PublishModal.svelte`. | §8.4 — publish is a task, not a confirmation. |
| `hub.featured(max=3)` IPC + `hub.subscribe(id)` + `hub.unsubscribe(id)` + `hub.detail(id)` + `hub.subscribed_list()`. | `app/web/frontend/src/lib/ipc/client.ts` + `…/ipc/types.ts`. | §7 — the route's IPC contract is eight methods, not six. |
| `Thread` under the hero eyebrow + above the grid on first paint + across the card on subscribe settled + across the footer on publish settled. | Reused `<Thread />` component. | MOAT §3 — the thread is the signature gesture; the Hub owns four thread-draw moments, the same count as Skills. |
| `Tooltip` on the Publish CTA, the card menu `[⋯]`, the detail panel close. | New `<Tooltip />` component per MOAT §2.9. | MOAT §2.9 — no `title=` attributes that misbehave. |
| Subscribe settled = `Subscribed ✓` with synapse dot + bottom-edge thread + no toast. | `<SkillCard />` `state.kind === 'subscribed'` rendering. | §3.3 + MOAT §4 #6 — no celebration, just state. |
| `ErrorState` (once, route-level). | New `<ErrorState />` per MOAT §2.6. | MOAT §1.2 — extracted from the four inline copies. |
| `EmptyState` for the waking-up shelf (the "shelf is waking up" three-line state). | New `<EmptyState />` per MOAT §2.4. | MOAT §2.4 — three-line teach, not "Loading…". |
| Search cross-fade 200ms (`opacity 1 → 0` for old results, then `opacity 0 → 1` for new results). | New CSS in the route. | §4.5 — search resolves quietly, no thread, no shimmer. |
| Auto-sign-in-then-auto-open the PublishSheet, gated by `account.isSignedIn`. | In the PublishSheet's `onMount`. | §3.6 — Sign-in-then-publish is one motion, not two. |

### 9.3 Unchanged

| What stays | Why |
|---|---|
| The route path `#/hub` and the NavRail entry. | No external behavior changes. |
| The `hub.search(query)` debounced 250ms priming query. | APPFLOW §4.3 — keeps current behavior. The featured set is *added*, not a replacement. |
| The empty-state copy "The shelf is quiet." (kept for the S1 waking-up variant). | The user already read this phrase; the same words still work for the same condition. |
| The eyebrow `— The library · curated · safety-scanned`. | MOAT §4 #5 — the museum-label rhythm is correct; only the title's display italic needs to read like a complete sentence (which it already does). |
| The detail panel slide-in-right timing (320ms `--ease`). | MOAT §2.8 + DIRECTION §5 — the .c-sheet primitive owns this; the route reads the primitive's default. |
| The 240px sticky filter rail. | Matches Skills; one rail vocabulary. |

---

## 10. Closing note for Phase 4

This spec is exhaustive enough that a single screen of type design
and animation can be built without follow-up. Every state has its
copy; every motion has its duration; every key has its handler;
every icon has its name; every IPC has its signature; every
deletion has its reason.

The eleven things you'll find when you implement from this spec:

1. **No** 3D perspective on the shelf, the spines, or the grid —
   ever.
2. **No** `rotateY(-22deg)`, `rotateY(-30deg)`, or any
   `transform-style: preserve-3d`.
3. **No** `backdrop-filter: blur(4px)` anywhere — the detail panel
   and the publish sheet are paper-on-paper.
4. **No** "Verified ✓" badge — three honest trust levels, all
   visible.
5. **No** "Welcome to the future" copy — eyebrow + title read like
   a museum label.
6. **No** double shadows — one `--shadow-card` on hover, no second.
7. **No** rectangular focus halos — rounded pollen + synapse rings.
8. **No** press-scale-only — filter brightness + 0.5px settle, per
   MOAT §2.2.
9. **No** inline err / empty / loading / overlay blocks — five
   primitives (`<ErrorState />`, `<EmptyState />`, `<Thread />`,
   `<Tooltip />`, the `.c-sheet` taxonomy) own the five patterns.
10. **No** "Subscribed!" celebration — the card settles, a thread
    draws, a synapse dot lights. That's it.
11. **No** publish modal — the publish flow is a `.c-sheet` sliding
    up from the bottom (a task, not a confirmation).

And the eight things you will ship:

1. **HeroShelf — at most three featured cards, weekly synapse
   pulse, vertical-centered to the Publish CTA.**
2. **Filter rail — Categories / Trending / Trust / My subscriptions.**
3. **One card grammar — shared with `Skills.svelte`.**
4. **Detail panel — same `.c-sheet`, swaps `Install` → `Subscribe`
   / `Sign in to subscribe`.**
5. **PublishSheet — slides up from bottom; no scrim; eight fields.**
6. **Not-signed-in is honest — full read access, the Subscribe
   action is local-first, the Publish CTA routes through
   sign-in-then-publish as one motion.**
7. **The keyboard is exhaustive — `/` focuses search, Tab cycles
   cards, arrow keys move row-major, ⌘S subscribes, ⌘U unsubscribes,
   ⌘Enter opens, ⌘P publishes, Esc closes the topmost overlay.**
8. **Reduced-motion is global — the route declares no `prefers-
   reduced-motion` blocks.**

If during Phase 4 you find a place where the spec is silent and the
default would be wrong, the answer is to ask before inventing. The
MOAT bar is "premium-quality." Anything that fails it is competent-
but-generic, and competent-but-generic is how a year gets lost.
