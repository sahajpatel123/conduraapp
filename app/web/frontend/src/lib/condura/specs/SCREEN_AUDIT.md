# SCREEN_AUDIT — Condura Audit · `Audit.svelte` · `#/audit`

> **Status:** Phase-2 architecture spec. Phase 4 implements against this.
> **Contract:** `MOAT.md` (premium bar), `APPFLOW.md §4.5` (current Audit
> route), `icons.ts` + `condura.css` (the actual tokens available). The
> original brief named five "north-star" documents; only `APPFLOW.md`,
> `MOAT.md`, and `TEARDOWN.md` are on disk in
> `app/web/frontend/src/lib/condura/`. `DIRECTION.md` and `DESIGNLANG.md`
> are absent at this commit, so this spec leans on `MOAT.md` for the
> design grammar and on the actual token table from `condura.css` for
> everything visual.
>
> The current `Audit.svelte` (599 lines) renders a single-column chain
> with a text input + 4 level-chips for filters, hover-only detail with
> a sticky right rail, and an inline `err-state` block copy-pasted into
> three sibling components. This spec keeps the visual chain metaphor
> (because the chain IS the brand, and the Audit is the one screen where
> it is load-bearing), but restructures the layout into a true
> three-region forensic surface — filter rail / timeline / detail —
> and re-states the empty/loading/error states under MOAT's three-line
> contract.

---

## Table of Contents

1. [What this route is and isn't](#1-what-this-route-is-and-isnt)
2. [Inheritance — what the spec inherits from MOAT/APPFLOW/TEARDOWN](#2-inheritance)
3. [Layout & Content](#3-layout--content)
4. [State Matrix — five states, full copy](#4-state-matrix)
5. [Motion Choreography](#5-motion-choreography)
6. [Keyboard](#6-keyboard)
7. [Components Used — boundaries and props](#7-components-used)
8. [Data Fetched — IPC contract](#8-data-fetched)
9. [Design Decisions — which MOAT rules this passes](#9-design-decisions)
10. [What this spec deletes from the current `Audit.svelte`](#10-what-this-spec-deletes-from-the-current-auditsvelte)

---

## 1. What this route is and isn't

**Is.** The forensic surface for Condura — the human-readable window
onto the HMAC-chained, append-only event log that records every agent
action, every IPC, every gatekeeper decision, and every consent ticket.
Per `CLAUDE.md §10.5` and `§18`, this log is for **forensics**: "When
the agent sends the email the user didn't authorize, the user can prove
exactly which model and which prompt caused it." The surface must be
**honest** (every entry, no redaction except secrets), **readable**
(mono timestamps, color-coded blast-class, expandable details), and
**scrubbable** (the last 24 hours, navigable by timeline).

**Is not.** Not a debugger (there is no step-through, no inspection of
in-flight state). Not a dashboard (no charts, no aggregates, no
"today's count" tile — those belong to `Replay`). Not a privacy export
(export of the full chain as JSONL is one short row, not the spine of
the surface). Not a chat, not a settings surface, not a wizard.

**Mental model the user carries away:** *Audit is a ledger. Each row
is a fact. The chain is unbroken unless a row says otherwise. If the
user wants to know "what did the agent do yesterday," they find the
timestamp and read down.*

**One sentence for the eyebrow:** *— Forensics · HMAC-chained · append-only.*

**One sentence for the title:** *Every action, on a thread.*

**Why "on a thread":** the route's defining graphic is the titlebar's
`TitlebarThread` — the spine of the brand — reused here at the bottom
of the screen as the timeline scrubber. The user has been trained by
the titlebar to read this line as "what is the agent doing." Putting
the same line in the Audit tells them: *here is what it did.*

---

## 2. Inheritance

The spec assumes the following are already in place; this spec does
**not** redefine them.

| From | What this spec uses as-is |
|---|---|
| `MOAT.md §2.3` | `prefers-reduced-motion` is respected via one global rule in `condura.css`. The component declares no media-query blocks. |
| `MOAT.md §2.4` | Empty states are three lines: **what** / **why empty** / **next action.** The Audit empty states follow that shape exactly. |
| `MOAT.md §2.6` | One `ErrorState` component owns all error rendering. The route uses it. No inline `err-state` blocks. The current `Audit.svelte:75-86` inline block is deleted; the route uses the shared component. |
| `MOAT.md §2.7` | `.tactile` global class owns the row's transition timing. The component declares no per-row `transition:` list. |
| `MOAT.md §2.8` | The Detail panel is a `.c-sheet` (not the sticky-right-rail hover-card the current Audit uses). The filter rail is a left `.c-popover` on narrow viewports. |
| `MOAT.md §2.9` | Icon-only buttons in the row menu / scrubber controls / right-rail use `<Tooltip label>` — no `title=`. |
| `MOAT.md §3` | "Finished" states are communicated by a `Thread` drawing in (left-to-right). The scrubber uses the thread as its time axis. The detail panel's "expand row" draws a thread under the row when the user expands. |
| `MOAT.md §4 #1` | No gradient text. No gradient. |
| `MOAT.md §4 #2` | All icons go through `<Glyph name="…" />`. No emoji. No Unicode-as-icon. |
| `MOAT.md §4 #3` | No `backdrop-filter: blur()` on this route. The detail panel is a `.c-sheet` — paper-on-paper, hairline left edge, `--shadow-float` if it needs elevation. |
| `MOAT.md §4 #4` | Status colors are `--ok` (READ), `--synapse` (READ per blast class), `--pollen` (WRITE), `--warn` (prompt verdict), `--danger` (NETWORK / DESTRUCTIVE / block verdict). No purple, no teal. The blast-class badge is the *only* place these colors are used on the row, and they map 1:1 to the four blast classes per `internal/blastradius/blastradius.go`. |
| `MOAT.md §4 #5` | No "Welcome to the future" copy. The empty state and the title read like a museum label, not a landing page. |
| `MOAT.md §4 #6` | No "Awesome!", no "Audit passing!" celebration. The chain badge reads `Chain intact` or `Chain broken at row N`; that is the entire signal. |
| `MOAT.md §4 #7` | No spinner. Loading is a `Thread` drawing left-to-right over the scrubber (`drawthread`, `var(--dur-slow)`, `var(--ease)`), exactly once per surface. |
| `MOAT.md §4 #8` | Focus halos are rounded — the global `--shadow-focus` (the 4px pollen halo) or, for ≥8px radii elements, a 5px halo + 2px synapse ring per MOAT §2.1. Square halos are forbidden. |
| `MOAT.md §4 #9` | One elevation token per surface. The detail sheet uses `--shadow-float` at rest and at hover. No stacked shadows. The current `Audit.svelte`'s per-card `transform: translateX(2px)` on hover remains the only motion; no scale effects, no rotate. |
| `MOAT.md §4 #10` | Every animation in this spec answers *what is this communicating?* Entrance = "data ready"; row fade-in stagger = "rows are landing in order"; live-entry slide-in = "a new fact has been recorded"; scrubber thread draw = "this is the time axis"; detail panel slide = "you have selected one fact"; expand-inline = "you have asked for the full record"; chain-broken hair draws left→right-only-halfway = "this is broken, here is where." |
| `CLAUDE.md §10.5` | Audit log is HMAC-chained, append-only. The UI never edits a row. The "verify chain" affordance is read-only — if it returns broken, the UI surfaces the broken-row index but does not offer "fix." |
| `CLAUDE.md §6` | No new permission prompts are triggered from this surface. The Audit is read-only forensics, with one consent-gated action: **export the filtered chain as JSONL.** Export triggers the existing ConsentModal flow (a NETWORK WRITE to disk) — no new modal in this route. |
| `APPFLOW.md §7` | The state inventory table is augmented with five new rows for the redesigned Audit. Old `Audit · loading / empty / error / detail` rows are replaced with the regions from §3. |
| `TEARDOWN.md §7` | Loading state uses a thread draw (the loom pattern). Empty states are 3-line. Filters use chip toggles, not dropdowns. |

---

## 3. Layout & Content

### 3.1 Page-level structure

The Audit route renders inside the Shell's Main region, right of the
`NavRail` (which is 64px collapsed / 200px expanded per
`SCREEN_NAVRAIL.md §2.2`).

```
┌────────────────── Audit route (max-width 1480px, padded) ──────────────────┐
│                                                                            │
│  — Forensics · HMAC-chained · append-only                                  │  eyebrow (mono, 11px, caps, --content-faint)
│                                                                            │
│  Every action, on a thread.                                                │  title (display, clamp 28→40px, --content)
│  Each row is a fact. An unbroken green thread means the chain              │  sub (sans 16, --content-soft, max 56ch)
│  verifies — when something goes wrong, you can prove exactly               │
│  what happened.                                                             │
│                                                                            │
│  ┌────────────────── Toolbar row (sticky, --hair bottom) ────────────────┐ │
│  │  [⌕ search action…]   [class: ▾READ WRITE NETWORK DESTRUCTIVE ]        │ │  left cluster: search, class chips
│  │  [app: ▾]           [model: ▾]            [↻ refresh]    [⤓ export]  │ │  middle/right cluster: app, model, refresh, export
│  │  1,284 entries · 24h window · Chain intact                              │ │  meta + chain badge
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                                                            │
│  ┌──────────────┐  ┌──────────────────────────────┐  ┌─────────────────┐ │
│  │              │  │                              │  │                 │ │
│  │  Filter      │  │   Timeline                   │  │   Detail        │ │
│  │  rail        │  │   (rows of HMAC entries)     │  │   panel         │ │
│  │              │  │                              │  │                 │ │
│  │  240px       │  │   flexes                     │  │   360px         │ │
│  │  sticky      │  │   with viewport              │  │   appears       │ │
│  │              │  │                              │  │   when a        │ │
│  │  (collapses  │  │   infinite scroll,           │  │   row is        │ │
│  │   to a       │  │   paginated 100/page         │  │   selected      │ │
│  │   ⛁ button  │  │                              │  │                 │ │
│  │   below      │  │                              │  │                 │ │
│  │   960px)     │  │                              │  │                 │ │
│  └──────────────┘  └──────────────────────────────┘  └─────────────────┘ │
│                                                                            │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │ [─── ◍─────────────────────────────] scrubber · 24h                 │  │  TimelineScrubber
│  │  00:00                       12:00                       now          │  │  60–80px tall, sticky to bottom
│  └──────────────────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────────────────┘
```

The page total max-width is **1480px** (wider than the current 980px —
the timeline earns the room, and three columns need the air). Above
1480px, content stays centered. Below 1280px the detail panel becomes a
`.c-sheet` (slides over the timeline, does not shrink it). Below 960px,
the filter rail collapses to a `[⛁]` button that opens a `.c-popover`
sheet from the left. Below 720px, the detail sheet becomes a full-screen
`.c-sheet` (covers the timeline; Esc returns). Below 480px, the toolbar's
right cluster collapses to icon-only.

### 3.2 Region A — Filter rail (240px, sticky on scroll)

A `.c-paper` vertical column, 16px from the timeline. The rail is its
own landmark (`<nav aria-label="Audit filters">`). It contains,
top-to-bottom:

#### A.1 "When" group — date range

- A pair of `<input type="date">` fields (Start, End). Defaults to
  **last 24 hours** (now − 24h, now).
- Three preset chips below the inputs, mono 11 caps: `Last 1h`,
  `Last 24h` (default, active), `Last 7d`, `Last 30d`, `All`.
- A `<input type="search">` free-text **search box** above the chips
  that filters by `actor`, `action`, `message`, `app`, `path`, `url`
  — matches anywhere in those fields. The search box has a 1px
  hairline focus halo per MOAT §5.2.

#### A.2 "What" group — blast class + verdict

A stacked column of 4 toggle chips (one per blast class from
`internal/blastradius/`): `READ`, `WRITE`, `NETWORK`, `DESTRUCTIVE`.
Each chip:

- Has the blast-class color as its left border (1.5px, the
  blast-class-color, hairline otherwise).
- Mono 11 caps label.
- A subtle bg tint when active (6% mix of the class color, see §4.1).
- Below the 4 blast-class chips, a single 5-way segmented toggle for
  **verdict**: `all` / `allow` / `block` / `prompt` / `error`. The
  verdicts come from `result` in the audit row: `allow` (synapse), `block`
  (danger), `prompt` (warn), `error` (danger).

#### A.3 "App" group — checkbox column

The distinct `app` values from the current result set, with mono
counts. Default: all checked. Five most frequent pinned, `+N more`
overflow at the bottom that opens an inline expand. If 12+ apps are
in the result, the rail collapses the list and shows a `+ N more`
link.

#### A.4 "Model" group — checkbox column

The distinct `model` values from the current result set (per the new
`model` field on the event), with mono counts. Behaves like App: 5
pinned, `+N more` overflow. The Model column is what answers "whose
prompt was this?" — the user comes here when a specific agent output
felt wrong and they need to see if it's a model issue.

#### A.5 Footer of the rail

A single line, monospaced, 11px, faint:
`Filters update as you check.`

A `.btn-ghost` mono-pollen `Reset filters` link appears below it
whenever any filter is non-default.

### 3.3 Region B — Timeline

#### B.1 Container

The timeline sits in a 1fr-wide column with `padding: 0 var(--space-4)`.
A subtle vertical 1px `<Thread />` runs down the left edge as the
**chain spine** (in `--synapse`); each row's blast-class dot sits
on the spine.

Above the rows, a one-line meta strip: `124 rows · sorted newest first ·
window: last 24h`. Mono 11 faint.

Below the rows (still inside the region, not at the page bottom): a
**[ Load older ]** mono-pollen ghost button that loads the next
pagination window (100 rows). No "next/prev page" pagination — the
timeline is infinite scroll backwards (older rows), newest first.

#### B.2 Row anatomy (each row is a `.row` button)

```
┌─ row · 1 row = 1 fact ────────────────────────────────────────────────┐
│  ⬤  READ   12:04:18.412Z   llm.chat                             ▾     │  ← chain dot, blast-class badge, mono ts, action mono, chevron
│                  "…draft the apology email…"  — sonnet-4.5  · 1.2k in  │  ← intent summary (italic), model + tokens meta
└──────────────────────────────────────────────────────────────────────┘
```

Five columns, top to bottom:

| Row | Field | Style |
|---|---|---|
| 1 | **Chain dot** | 8px filled dot in blast-class color, sitting on the chain spine. `box-shadow: 0 0 6px currentColor` to glow the chain. |
| 2 | **Blast-class badge** | `READ` / `WRITE` / `NETWORK` / `DESTRUCTIVE`. Mono 10 caps, letter-spacing 0.12em. 1.5px left border in blast-class color; bg tint when active (6% mix of color). Same color mapping as §A.2. |
| 3 | **Timestamp** | `HH:MM:SS.mmm` in mono 12, `--content-soft` (when ≤ 24h ago) or `--content-faint` (older than 7d, mono 11). The milliseconds catch read failures; they are honest. |
| 4 | **Action** | The `action` field, mono 11. When the action is `chat`, the column carries the action glyph (`<Glyph name="chat" />`) inline at 10px. |
| 5 | **Intent summary** | Italic display 14, `--content-soft`. Truncated at one line (`text-overflow: ellipsis; white-space: nowrap`). 200-char ceiling before truncation. The summary is the agent's articulated intent ("…draft the apology email…"); not the literal message body. |
| 6 | **Model + tokens** | Mono 10 faint right cluster: `{model} · {tokens_in}in / {tokens_out}out`. Hidden when the row has neither field. |
| 7 | **Chevron** | `<Glyph name="chevron-down" />` 14px, faint. Rotates 180° on expand. |

The row's left edge carries a **2px synapse strip** that scales from
`scaleY(0)` to `scaleY(1)` over `var(--dur)` when the row is selected
— the same gesture as the NavRail's active segment (SCREEN_NAVRAIL
§9). When unselected, the strip is invisible (`opacity: 0`).

#### B.3 Blast-class color spec

This is the most visible color system in the route. The colors are
**locked** at the values below; they match `blastradius.Class` in
`internal/blastradius/blastradius.go:15-31` and the existing
`MOAT.md §4 #4` palette rules. The chain spine's color follows the
**highest-risk** row in the current viewport (so a WRITE-with-
DESTRUCTIVE sequence reads as a danger gradient climbing up).

| Blast class | Color (light) | Color (dark) | Token |
|---|---|---|---|
| READ | `--ok` (#2E7D4F) | `--ok` (`#5BAE7F` after the dark-map) | Verified observation. |
| WRITE | `--pollen` (#C97B2E) | `--pollen` (`#E89B4F`) | Local mutation. |
| NETWORK | `--danger` (#A3312A) | `--danger` (`#C76A60`) | Outbound call, click link, send message. |
| DESTRUCTIVE | `--danger`, **filled badge** with `--paper` text | same, `--paper` text | Hard-to-undo, requires consent. |

NETWORK and DESTRUCTIVE share `--danger` because their consequence is
similar (irreversible or external); the row's badge text is what tells
them apart at a glance. The chain-dot coloring uses the same mapping.

The blast-class colors appear in **three places** and **only** three:

- Row §3.2-1 — chain dot (always colored).
- Row §3.2-2 — badge border + bg tint on active.
- Toolbar meta strip, line 2 — `Chain: 84% READ · 12% WRITE · 4% DESTRUCTIVE`, the ratio histogram in mono 10 caps with each segment in the corresponding color.

No other surface in the route uses these colors for decoration.

### 3.4 Region C — Detail panel (360px, slides in from right as a `.c-sheet`)

The detail panel is a `.c-sheet` (per MOAT §2.8) that slides in when a
row is clicked. It is owned by the page, not floating with a scrim —
the rest of the route stays live. Esc closes it. It contains:

| Block | What |
|---|---|
| **Sheet header** | Close `×` button (top-right, 32×32 hit-area, hairline focus halo per MOAT §2.1); row `#id` in mono 11 caps; blast-class badge in the same color as the row; timestamp `2026-07-02T12:04:18.412Z` in mono 12; verdict pill (`Allow` synapse / `Block` danger / `Prompt` warn / `Error` danger). |
| **Intent** block | Italic display 17, `--content`. The full intent string (no truncation). One single paragraph; this is what the agent said it was going to do. |
| **Action** block | A code-block-surface (`.codeblock` `--surface-sunken`, mono 12) showing the JSON of the action payload (verbatim — `target_app`, `target_url`, `path`, `command`, `body`). The user can `⌘C` to copy (the panel owns the keyboard handler). For shell actions, the second line is the parsed+validated AST from `internal/sanitize/shell/` ("Sanitizer allowed: bash, no `..`"); for destructive actions it is followed by the ConsentResult from the Gatekeeper (`consent_required: true → user_approved` etc.). |
| **Model + prompt** block | Two sub-blocks, top to bottom. **(a) Model meta:** `model · provider · backend · session_id`, mono 12. **(b) Prompt excerpt:** the `user_prompt` (first 600 chars) and the `assistant_intent` (first 600 chars) the model output before the action was issued. Both mono 12 in a scrollable `--surface-sunken` well (max-height 280px). The prompt is `⌘C`-copyable as one block. |
| **Hash chain** block | Mono 11: `prev_hash · sha256(this_hash) · HMAC verified: ✓` for the row. The HMAC is **read from the chain, not recomputed client-side** — the panel fetches `audit.verifyIntegrity` for the row's range and displays the row's own verification status. A red label `"Chain broken at row #N: …"` replaces the green check when broken. (Per MOAT §4 #6, no celebration. A `⚠` glyph + `Chain broken` mono caps is the entire signal.) |
| **Audit trail** block | A 3-line `mono 11` summary: `app: com.tinyspeck.chatly · path: /Users/.../inbox.json · url: https://…`. One key per line. Empty fields are omitted entirely (no `—` placeholder; the omission is the signal). |
| **Footer actions** | Two ghost buttons, mono 11 caps: `[ Copy intent ]` on the left, `[ Export this row ]` on the right. Both `⌘C` / `⌘E` keyboard-accessible; the panel owns both chords while it has focus. |

The detail panel renders inside a `{#key row?.id}` block so that
switching between rows re-mounts and replays the slide-in.

If the chain is broken (the `verifyIntegrity` RPC returns
`{ ok: false, broken_at_id: N }`), the **chain-broken badge** in
the toolbar swaps from `Chain intact` to `Chain broken at row #N`,
and the first row of the detail panel shows a one-line italic
warning: *"The chain is broken at this row. The HMAC does not match
the previous hash. The rows below this point should not be trusted."*
No "panic" copy; no celebration copy. The italic hair draws halfway
left→right (the `err-hair` pattern, but visually truncated at 50%).

### 3.5 Region D — TimelineScrubber (60–80px, sticky to bottom of route)

A `.c-paper` strip pinned to the bottom of the route (not the page,
because the page is the Shell). It contains, left to right:

- A **timestamp label** "00:00 ──── now" pinned to the left edge,
  mono 11 faint. Updates as the scrubber moves.
- A 1fr-wide **track**: a 1px hairline (the time axis).
- **Tick marks** at minute granularity for the last hour, hour
  granularity for the last 24h, day granularity for the last 30d.
  Each tick is a 4px vertical, `--hair-strong`. Hovering a tick
  reveals a small mono label above it (`12:34`).
- A **thread draw across the track** as the `<Thread />` component:
  the brand spine, used here as the time axis. 1.25px `--synapse` at
  18% glow, drawn in left→right over `--dur-slow` when the route
  mounts (MOAT §3).
- A **playhead** — a 6px pollen dot at the current scrubber position,
  with a 14px `--synapse-glow` halo behind it. The dot is the
  "selected timestamp" — it tells the user *which row would land on
  this point if they clicked*.
- A faint mono `↹` hint to the right of the playhead:
  `←  →  to step · drag to scrub · click to jump`.

The scrubber collapses to a single 1px hairline + the playhead if the
user has not interacted with it for > 20 seconds (the breathing
gesture goes to sleep). Re-entering the route or pressing any key
restores the full chrome.

The scrubber does **not** snap to events — it is a continuous time
axis. Each event from the result set is drawn as a **2px dot** at its
timestamp along the track (in blast-class color, see §3.3). Up to 200
dots render (further events bin into a faint histogram color).

### 3.6 The toolbar's chain badge

A `<button class="chain-badge">` in the toolbar's right cluster (just
left of the export button). Two states:

- **`Chain intact`**, with a 6×6 synapse filled dot + `Chain intact`
  mono 11 caps. Hover → tooltip `The HMAC chain verifies. Every row
  has been signed, in order, since the first write.`
- **`Chain broken at row #N`**, with a 6×6 danger filled dot + the
  label. Hover → tooltip `The HMAC chain is broken at row #N. The rows
  beyond this point should not be trusted. [Verify integrity →]`
  (clicking calls `audit.verifyIntegrity`).

The badge is **always visible** when the route is mounted; the page
is not honest without it.

---

## 4. State Matrix — five states, full copy

These are the five visual states the Audit route can be in. They are
**not** mutually exclusive — the Timeline has a state, the
TimelineScrubber has none, the Detail Panel has its own
selected-with-row state. The toolbar and the row layout are shared.

| # | State | What you see | Trigger / data condition |
|---|---|---|---|
| **S1** | **Empty — "Awaiting first action"** | Quiet empty state (see §4.1) | `audit.list` returns zero rows AND no in-flight filter applied AND the daemon reports uptime > 60s (so we know the boot is genuinely through). |
| **S2** | **Loading** | The TimelineScrubber's `<Thread />` draws left-to-right over the page width (1px `--synapse`, `--dur-slow`, `--ease`). The toolbar meta strip reads `Reading the chain · /audit`. Once data resolves, rows fade in beneath the scrubber. | `audit.list` first call pending; the request takes > 180ms. Below 180ms the user never sees a loading state. |
| **S3** | **Filter-no-results** | Filtered empty state (see §4.2) | `audit.list` returns zero rows AND any non-default filter is applied. |
| **S4** | **Error** | A route-level `<ErrorState />` (per MOAT §2.6) renders once, anchored above the toolbar. The TimelineScrubber still draws (the page is alive even when the chain is unreachable). | `audit.list`, `audit.export`, or the live SSE channel rejects. |
| **S5** | **Kill-switch-armed** | Full-screen overlay. The Audit route is hidden behind `<KillSwitchOverlay />` (per `APPFLOW.md §6.4`). | `halt.state.halted === true`. |

### 4.1 The Empty state (S1) — exact copy

Three lines, exactly per MOAT §2.4:

> **What's here.** The HMAC-chained event log. Every action the agent
> takes — chat turns, file edits, clicks, network requests, consent
> tickets — lands here as one row, in order, signed.
>
> **Why it's empty.** Nothing has happened yet. Every row is one fact;
> the first row will appear the first time the agent acts.
>
> **Next action.** *(no button — the empty state is honest about
> absence.)*

Beneath the three lines: a single `<Pulse phase="thinking" size={10} />`
next to a mono caps line, faintly pulsing every 1.8s:

> `AWAITING FIRST ACTION · the chain will write its own first row.`

The pulse is **not** a spinner. Per MOAT §4 #10, it is the heartbeat
gesture: the chain is alive, just quiet. The pulse rate (1.8s) is
slower than the live SSE entry rate, so when entries do arrive, the
transition from "awaiting" to "live" reads as a tempo acceleration.
The pulse rate does **not** accelerate; new entries insert at the top
of the timeline (their slide-in handles the visual arrival).

When the SSE channel receives its first event, the pulse fades to 0
opacity over `--dur-slow` and the row inserts.

**Explicit non-copy (intentional absence):** there is **no** suggestion
"try sending a message to see an event appear." That is `MOAT §4 #5`—
"Configure, not comply" doesn't apply here, but `MOAT §4 #6` does: no
fake enthusiasm, no "Let's make a row!" CTA. The audit is a ledger;
it logs what happens, and instructs by being quiet.

### 4.2 The Filter-no-results state (S3) — exact copy

The same three-line shape, but with different copy:

> **What's here.** The HMAC-chained event log.
>
> **Why it's empty.** No entries match this filter in the current
> window.
>
> **Next action.** *[ Widen the date range ]* or *[ Clear filters ]*
> — two mono-pollen ghost buttons at the bottom of the filter rail.
> Clicking either resets that group to its default state.

No pulse. The empty filter result is **truly empty**, not "awaiting"
— the page has rows, just not rows the filter accepts. The two ghost
buttons are the only affordance the user needs; per MOAT §4 #6, no
"Hey, looks quiet!" celebration.

### 4.3 The Loading state (S2) — exact copy

> **A thread draws across the scrubber** (left-to-right, 1px,
> `--synapse`, `--dur-slow`, `--ease`). The eyebrow under it reads, in
> mono caps 11, `--content-faint`:
>
> `READING THE CHAIN · /audit`

Once `audit.list` resolves, the thread arrives 100% of the way across
the page width and the rows fade in beneath it, staggered (see §5.1).
The route does **not** show partial rows while loading. If `audit.list`
returns in <180ms, the user never sees the loading state at all.

The toolbar's "↻ refresh" button is **disabled** while the loading
state is in progress (preventing double-fetches). The "⤓ export" button
is similarly disabled.

### 4.4 The Error state (S4) — exact copy

The route uses one `<ErrorState />` component (per MOAT §2.6, extracted
across `Chat.svelte` / `Channels.svelte` / `Skills.svelte` / `Hub.svelte`
/ `Audit.svelte`). It renders once per route, anchored above the toolbar.

The `ErrorState` shape (always three lines, exactly):

> **We couldn't read the chain.** (italic display 22, `--content`.)
>
> Cause: *`{noun}`*, e.g.
> - "Audit list returned no rows from `ipc.audit.list`."
> - "SSE channel to /events disconnected."
> - "Export failed: `permission denied writing to ~/Downloads/audit.jsonl`."
>
> Likely reason: *`{phrase}`*, e.g.
> - "The daemon was restarted."
> - "Your disk is full."
> - "The user denied the export consent modal."
>
> Next action: `[Try again]` pill — pollen-outline button. Above it on
> the right, a mono 11 link `[Open Settings →]`.
>
> *(Below all of this, an `err-hair` rule, 1px, left→right draw over 600ms `--ease`.)*

The TimelineScrubber continues to render (the page is alive even when
the chain is unreachable). The toolbar's chain badge swaps to `Chain
unknown · last verified {ago}` (mono 11 faint). The filter rail stays
interactive (filtering a broken chain is still meaningful — the user
might be probing whether the chain ever had rows).

### 4.5 The Kill-switch-armed state (S5) — exact copy

The full-screen `<KillSwitchOverlay />` is mounted per
`APPFLOW.md §6.4`:

> Headline: `Condura has stopped.` (italic display 36, `--content`.)
>
> Body: `Every active stream was canceled. The agent is not running.`
>
> Note: `Resuming mints a ticket you confirm from the CLI — the GUI
> never auto-restarts a halted agent. **Auto-recovery is the enemy.**`
>
> Button: `[ Mint resume ticket ]` (mono-pollen ghost).

While the overlay is mounted, the Audit route is hidden behind it.
The chain badge stays at its last-known state (e.g., `Chain intact` or
`Chain broken at row #N`); the route does not re-poll during kill-switch.

When the user resumes (via the CLI), the overlay dissolves and the
Audit route re-renders at its last view (filters preserved).

---

## 5. Motion Choreography

The route's motion follows one rule: **every animation answers "what
is this communicating?"** (MOAT §4 #10). Decorative loops are
forbidden; the only idle loop on the route is the S1 empty-state pulse
and the post-row-arrival `pollen-float` (both are explained below).

### 5.1 Entrance — TimelineScrubber draws, rows stagger in, then pulse

Trigger: `audit.list` first resolves AND the page has been mounted for
at least 60ms (let the route-enter animation land first).

```
Sequence:
  0ms     TimelineScrubber <Thread> starts drawing left→right (520ms, var(--ease))
  60ms    Toolbar animates in (route-enter, blur-in, 240ms)
  80ms    Filter rail animates in (translateX(-16px)→0, opacity 0→1, 280ms)
  120ms   Rows begin to stagger in (40ms per row, fade + translateY(8px)→0)
  120+30·N ms   Last row landed (N is row count, capped at 30)
```

Per row:

```
opacity: 0 → 1
transform: translateY(8px) → 0
duration: 320ms
easing: var(--ease)
delay: i * 40ms, where i is the row's index (top to bottom)
```

The stagger is computed in Svelte via `style:animation-delay={
${index * 40}ms}`. The stagger is capped at 30 rows
(`max(0, min(29, i)) * 40ms`) so the last visible row doesn't appear
1.2s after the first.

**Reduced-motion:** the entire 40ms-stagger sequence collapses to a
single 0ms stagger — all rows appear at once. The scrubber thread
still draws (the scrubber draw is *meaningful* — it's the time axis,
and per MOAT §2.3 reduced-motion only suppresses decorative motion).

### 5.2 Hover — row tints, chain dot brightens

```css
.row:hover {
  background: var(--surface-card);
  border-color: var(--hair);
  transform: translateX(2px);  /* the only row-level transform */
}
.row:hover .chain-dot {
  filter: brightness(1.15);   /* the dot brightens, never scales */
  box-shadow: 0 0 10px currentColor;
}
```

`:active` adds the global tactile rule (per MOAT §2.7): `transform:
translateX(0.5px) scale(0.985) filter: brightness(0.96) saturate(1.05)`.
The row visibly settles into the page.

`:focus-visible` adds the rounded halo (per MOAT §2.1): for rows with
`border-radius: var(--r-sm)` (10px), the halo is
`box-shadow: 0 0 0 5px var(--pollen-halo-color), 0 0 0 2px var(--synapse)`.
No rectangular outlines anywhere.

**Hard delete:** the current `Audit.svelte:421-433` per-card
transition list that redeclares `transform var(--dur) var(--ease),
background var(--dur) var(--ease), border-color var(--dur) var(--ease),
box-shadow var(--dur) var(--ease)`. The row uses the global
`.tactile` class; the CSS owns the timing.

**Reduced-motion:** the row's `transform: translateX(2px)` collapses to
`translateX(0)`. The chain-dot brightening does not run. The focus
halo still draws.

### 5.3 Select — row gets a left strip; detail panel slides in

When a row is `aria-pressed="true"` (clicked or Enter/space pressed
on a focused row):

1. **Row left strip.** The row's `::before` (a 2px-wide vertical
   element at `left: -8px`) animates from `scaleY(0)` to `scaleY(1)`
   over `var(--dur)` (280ms). The strip is `--synapse`. The strip's
   color matches the row's blast-class color (so a DESTRUCTIVE row
   gets a danger strip; the synapse strip is reserved for READ-only
   rows). The strip is the same gesture as NavRail's active segment
   (SCREEN_NAVRAIL §9).

2. **Detail panel slides in.** `transform: translateX(360px)` →
   `translateX(0)` over 320ms `--ease`. Background changes to
   `--surface-card` simultaneously. Shadow (`--shadow-float`) fades
   from 0 to 1 over the same duration.

The detail panel always mounts on top of the timeline; it never
shrinks the timeline below the panel width. The user can still scroll
the timeline beneath the panel (the panel has no scrim). The detail
panel itself does not have a backdrop blur (per MOAT §4 #3, no glass).

The panel renders inside a `{#key row?.id}` block so that switching
between rows re-mounts and re-runs the slide-in.

If the user deselects (Esc, second click on the same row, close ×),
the strip scales back to `scaleY(0)` and the panel slides back over
320ms `--ease` simultaneously.

**Reduced-motion:** both move instantly. No transform animation. The
shadow still fades.

### 5.4 Expand — inline details grow from the row

When a row is in `expanded` state (clicked chevron, or `Enter` while
focused with `expanded=true`):

- The row's `transform: translateX(2px)` reverts to 0; the row
  **gains 280px of vertical height inline** (between the row's
  meta-strip and the next row) over 240ms `--ease`. The expansion
  pushes the rows below it downward with no re-mount (`grid-row` +
  CSS transition on `height`, not on the DOM).
- Inside the expanded area, two sub-blocks fade in:
  - **Action payload** — the JSON of the action (verbatim), in a
    `--surface-sunken` mono 12 well. `opacity: 0 → 1` over 240ms.
  - **Hash chain** — the prev_hash + this_hash + verify status, mono
    11. `opacity: 0 → 1` over 240ms with a 80ms delay.
- The row's chevron rotates 180° over 280ms `--ease`.

When collapsed, the row's `max-height` collapses from `280px+` to its
rest height (40px) over 240ms `--ease`. The two sub-blocks fade out
180ms first (so they don't pop during the collapse).

**Reduced-motion:** the row's height collapses instantly. The two
sub-blocks fade in/out without translate. The chevron snaps.

### 5.5 Live entries (SSE) — slide in from top with a pollen glow

When the SSE channel (`audit.subscribe`) receives a new event, the
route inserts it at the **top** of the timeline (newest first). The
insert animation:

```
0ms       Row opacity: 0 → 0.4
          Row translateY: -16px → 0
          Synapse glow under row: opacity 0 → 1
120ms     Row opacity: 0.4 → 1
          Synapse glow fades out (600ms total, ease-out)
240ms     Glow fully gone
```

The "synapse glow" is a 2px vertical strip on the row's left edge,
in `--synapse-glow` at 18% opacity, that **fades** over 600ms (not
flashes). The user's eye reads it as "this row is new." Once the glow
is gone, the row is just a row.

The blast-class badge on the live row uses the same animation as
S5.1's stagger (`opacity 0 → 1`, `translateY(8px) → 0` over 320ms).

**Rate-limit:** the route coalesces bursts of live entries. If 5+
entries arrive within 400ms, the route inserts the **first** one with
the glow animation, and the remaining 4+ without the glow (just
opacity). The toolbar's `<Thread />` re-draws from the right edge for
each coalesced batch, signaling "more arrived."

**Reduced-motion:** the slide-in is replaced with an instant insert
(no translate, no glow). The row just appears.

### 5.6 The Thread (MOAT §3) — when it draws

The thread draws across the Audit route in **seven** moments:

1. **Route mount** — across the TimelineScrubber (S2).
2. **Filter rail re-render after a successful filter change** — a
   short hairline under the rail's footer (1px × 240px, `--synapse`,
   `--dur`, `--ease`).
3. **Live entry coalesced batch** — the scrubber re-draws from the
   right edge (a "reverse" thread draw, signaling arrival).
4. **Chain badge swaps to `Chain broken`** — the `err-hair` variant
   draws left→right only halfway (50% of the way), then stops. The
   remaining 50% stays undrawn. This is the visual signature of
   "broken" — the chain cannot complete.
5. **Error state S4** — the `err-hair` below the `ErrorState`
   component (the existing pattern, untouched).
6. **Detail panel opens** — the chain spine (left edge of the
   panel) draws top→bottom over 320ms (a vertical thread).
7. **Export success** — a single 1px hairline at the very bottom of
   the toolbar, drawing left→right over `--dur-slow`. The export
   modal's confirmation plays this as "the chain has been copied
   faithfully" (not a celebration — a completion gesture, per MOAT §3).

The thread is the spine of this route the same way it is the spine of
the titlebar (`TitlebarThread.svelte`) and the chat-turn dividers. It
is the only allowed flourish for "this is now finished." (MOAT §3.)

### 5.7 Pulse animations on this route

Two pulse animations, both with explicit meaning:

1. **S1 empty-state pulse.** `<Pulse phase="thinking" size={10} />`
   next to the `AWAITING FIRST ACTION` mono caps line. Breathes at
   1.8s (slower than the SSE arrival rate — when entries start
   arriving, the gap between pulses and arrivals is clear). When the
   first live entry arrives, the pulse fades to 0 over `--dur-slow`.

2. **Live-entry pulse during coalesce** — the chain-dot briefly
   pulses (`pollen-breath` 1.6s) on the most recent live entry. Only
   when more than 1 entry arrives per second. Answers: *this row is
   the newest; the chain is moving fast.*

No other idle loop exists on this route. The scrubber's playhead is
static when idle (it only moves on user interaction). The chain dot is
static when no hover/active. The reduced-motion profile freezes both.

---

## 6. Keyboard

The route's keyboard surface is added to the global shortcuts
documented in MOAT §2.10. Keys are bound at the route level
(`onMount` registers them, `onDestroy` removes them).

### 6.1 Global keybindings (already in scope of MOAT §2.10)

| Key | Action |
|---|---|
| `/` | Focus the date-range search input at the top of the filter rail. If the user is already inside a text input, `/` types a literal `/`. |
| `Esc` | If a row is expanded → collapse. If a row is selected → deselect + close detail panel. If a `c-popover` is open (the search-by-class popover at narrow viewports) → close it. If the search input is focused → clear + blur. If everything is closed, Esc is a no-op. |
| `⌘K` | Opens the command palette (already global; this route does not re-handle). |
| `⌘,` | Opens Settings (already global per MOAT §2.10). |

### 6.2 Timeline (focus inside region B)

The timeline is one virtual focus-trap when active. Tab inside the
timeline travels rows top-to-bottom.

| Key | Action |
|---|---|
| `Tab` | Next row. |
| `Shift+Tab` | Previous row. |
| `Enter` | Open the focused row's detail panel AND expand inline. |
| `Space` | Open the focused row's detail panel (without expanding inline). |
| `↑` | Move focus to the row directly above (one row, regardless of blast class). |
| `↓` | Symmetric — row directly below. |
| `Home` | Move focus to the topmost row. |
| `End` | Move focus to the bottommost row. |
| `PageDown` / `PageUp` | Scroll the timeline container (when the timeline extends past the viewport). |
| `⌘E` | Export the **filtered** chain as JSONL. Triggers the existing ConsentModal flow (per CLAUDE.md §6). Modal must approve before disk write. Disabled if there are zero filtered rows. |
| `⌘F` | Focus the date-range search input (same as `/`). |
| `⌘R` | Refresh (same as the toolbar's `↻ refresh` button). |
| `V` | Verify integrity (calls `audit.verifyIntegrity`). The chain-badge updates. |

### 6.3 Row-level keybindings (when a row is focused)

| Key | Action |
|---|---|
| `→` | Expand the focused row's inline details (same as clicking the chevron, or `Enter`). |
| `←` | Collapse inline details. |
| `E` | Expand inline details. (Alias for `→`.) |
| `C` | Collapse inline details. (Alias for `←`.) |
| `Return` | Move focus to the detail panel if it is open (focuses the close button at top, then `Tab` traverses the panel). |

### 6.4 Detail panel (open) keybindings

When the detail panel is mounted and focus is inside it:

| Key | Action |
|---|---|
| `Tab` | Cycles through detail panel interactive elements (top to bottom: close ×, copy-intent, expand-prompt, copy-prompt, export-row). |
| `Esc` | Closes the detail panel and returns focus to the originating row. |
| `⌘C` | With focus on a copyable surface (Intent / Prompt / Action payload), copies that surface's content to clipboard. |
| `⌘E` | Exports the **single row** (not the filtered chain) as a one-line JSON. |

### 6.5 Filter rail keybindings

| Key | Action |
|---|---|
| `Tab` | Cycles through rail groups (When → What → App → Model → Footer), then the interactive elements within each group, top to bottom. |
| `Space` / `Enter` | Toggle a chip / checkbox in the rail. (Both keys, per mature forms convention.) |
| `⌘F` | Opens the rail as a `.c-popover` sheet — useful at narrow viewports where the rail is collapsed. |
| `Esc` | When focus is inside a `.c-popover`, closes it and returns focus to the trigger. |

### 6.6 TimelineScrubber keybindings (when focused)

| Key | Action |
|---|---|
| `←` / `→` | Step the playhead by 1 minute forward / backward. Hold for repeat-step. |
| `Shift+←` / `Shift+→` | Step the playhead by 1 hour. |
| `Home` | Move playhead to the leftmost (oldest) timestamp in the current window. |
| `End` | Move playhead to the rightmost (now). |
| `Click / Space / Enter` | Move playhead to the position under the cursor (mouse) or to `End` (keyboard). |
| `Drag` | Pointer-down + drag the playhead. While dragging, the playhead's halo brightens to `--synapse-glow` 60%. |
| `S` | Scrub to the next event blast-class-WARNING row (a WRITE or above). Same gesture as `j/k` in vim for "next thing that matters." |

### 6.7 Reduced-motion note

All keybindings above are unchanged when the user has `prefers-
reduced-motion: reduce` set. Animations change; the keys don't. (Per
MOAT §2.3, components never branch on motion prefs; the global CSS
handles it.)

### 6.8 Focus visibility

Every row, every chip, every checkbox, the toolbar's buttons, the
detail panel's close button, and the scrubber carry a rounded halo on
`:focus-visible` (per MOAT §2.1). Rows with `border-radius: var(--r-sm)`
(10px) get a 5px pollen halo + 2px synapse ring stacked. Buttons and
chips inherit the global `--shadow-focus`. The scrubber's `role="slider"`
gets a thicker 6px halo (it carries continuous state).

No `outline: 1px solid var(--content)`. No rectangular halos. (MOAT
§4 #8.)

---

## 7. Components Used

The route is composed of these components. The full prop contracts
below belong in each component's source file — this spec is the
short-form.

### 7.1 `<FilterRail />`

```ts
let {
  filters: {
    when: { start: Date; end: Date; presets: Set<'1h'|'24h'|'7d'|'30d'|'all'> };
    what: { classes: Set<'READ'|'WRITE'|'NETWORK'|'DESTRUCTIVE'>; verdict: 'all'|'allow'|'block'|'prompt'|'error' };
    apps: Set<string>;
    models: Set<string>;
    search: string;
  },
  facetCounts: {
    apps: Array<{ name: string; count: number }>;
    models: Array<{ name: string; count: number }>;
    blastClasses: Record<'READ'|'WRITE'|'NETWORK'|'DESTRUCTIVE', number>;
    verdicts: Record<'all'|'allow'|'block'|'prompt'|'error', number>;
    total: number;
  },
  onfilters: (next: Filters) => void,
  collapsed = false,  // true below 960px
  class: cls = '',
} = $props();
```

Owns:

- The four filter groups (When → What → App → Model) with the
  blast-class + verdict segmented toggles.
- The sticky positioning when not collapsed.
- The `[⛁]` collapse trigger below 960px viewport (opens a
  `.c-popover`).
- The "Reset filters" link.
- The filter-level keyboard handlers (Space/Enter on chips).
- The facet counts (live from the result set — when no rows match,
  the rail stays interactive).

Does **not** own:

- The Timeline.
- The Detail Panel.
- The scrubber.
- The IPC — filters are a client-side concern; the IPC is called by
  the parent.

### 7.2 `<TimelineRow />`

```ts
let {
  row: AuditEvent & { blast_class: 'READ'|'WRITE'|'NETWORK'|'DESTRUCTIVE'; model?: string; tokens_in?: number; tokens_out?: number },
  selected = false,
  expanded = false,
  isLive = false,         // SSE-arrived within the last 2s
  onselect: (id: number) => void,
  onexpand: (id: number) => void,
  oncopy: (id: number) => void,
  class: cls = '',
} = $props();
```

Owns:

- The 7-row layout per §3.2.
- The blast-class color mapping (READ → `--ok`, WRITE → `--pollen`,
  NETWORK → `--danger`, DESTRUCTIVE → `--danger` filled with `--paper`).
- The hover/active/selected visual states.
- The expand-inline animation (§5.4).
- The live-entry glow (§5.5).
- The row-level keyboard handlers (`E`, `C`, `→`, `←`, `Return`).

Does **not** own:

- The timeline container (owned by `Audit.svelte`).
- The detail panel (owned by `<AuditDetailPanel />`).
- The scrubber dot rendering (the scrubber owns its own event dot
  overlay).

### 7.3 `<AuditDetailPanel />`

```ts
let {
  row: AuditEvent & { blast_class: ...; verdict: ...; prev_hash: string; this_hash: string; chain_verified: boolean; payload: AuditPayload },
  chainStatus: { ok: boolean; brokenAtId: number | null; reason: string | null },
  onclose: () => void,
  oncopyIntent: () => void,
  oncopyPrompt: () => void,
  onexportRow: () => void,
  class: cls = '',
} = $props();
```

Owns:

- The 360px `.c-sheet` slide-in / out (transform translateX).
- The seven blocks per §3.4 (sheet header, intent, action, model+prompt,
  hash chain, audit trail, footer actions).
- The chain-broken italic warning row when `chainStatus.ok === false`
  AND the broken row is at or before `row.id`.
- The panel-level keyboard handlers (`Esc` closes, `⌘C` on
  copyable surfaces, `⌘E` on export).

Does **not** own:

- The row selection state — the page owns `selectedRowId` and passes
  the resolved row here.

### 7.4 `<TimelineScrubber />`

```ts
let {
  windowStart: Date,
  windowEnd: Date,
  events: Array<{ ts: string; blast_class: 'READ'|'WRITE'|'NETWORK'|'DESTRUCTIVE' }>,
  onscrub: (timestamp: Date) => void,
  onscrubCommit: (timestamp: Date) => void,   // fires on pointerup
  class: cls = '',
} = $props();
```

Owns:

- The 1fr-wide `.c-paper` strip pinned to the bottom of the route.
- The 1px `<Thread />` time axis (left-to-right on mount).
- The 2px event dots overlaid on the track (in blast-class colors).
- The 6px pollen playhead + 14px `--synapse-glow` halo.
- The breathing-gesture collapse to a 1px hairline after 20s idle.
- The keyboard handler for `←` / `→` / `Shift+←` / `Shift+→` /
  `Home` / `End` / `S` (jump-to-warning).
- The `role="slider"` ARIA semantics — value, valuemin, valuemax,
  valuenow, valuetext (timestamp).

Does **not** own:

- The events themselves — the parent passes the array.
- The Timeline rows (owned by `Audit.svelte`).

### 7.5 `<Thread />` — already exists, reused

```ts
let { orientation: 'h' | 'v' = 'h', draw = true, glow = true, class?: string } = $props();
```

Used in (this route):

- The TimelineScrubber's mount-draw (§3.5, §5.6 #1).
- The `ErrorState`-rendered `err-hair` (the existing pattern).
- The detail panel's left edge (vertical orientation, draw on mount).
- The toolbar's `Chain broken` swap (#4 in §5.6) — only halfway.
- The export-success completion hairline (#7 in §5.6).

### 7.6 `<Pulse />` — already exists, reused

```ts
let { phase: 'idle'|'thinking'|'awaiting'|'acting'|'consent'|'error'|'ok', size?: number, class?: string } = $props();
```

Used in:

- The S1 empty-state pulse (`thinking`, 10px, slow 1.8s breath).
- The toolbar's "↻ refresh" button while a refresh is in flight
  (`acting`, 8px).
- The S4 ErrorState header (`error`, 8px, with the
  existing pattern).
- Live-entry coalesce (#2 in §5.7) — `pollen-breath` 1.6s on the
  most recent dot when SSE bursts hit.

### 7.7 `<Glyph />` — already exists, reused

Icons used on this route, with their canonical `name` from
`app/web/frontend/src/lib/condura/icons.ts`:

| Where | Glyph name | Default size |
|---|---|---|
| Toolbar, search-input leading icon | `search` | 14 |
| Toolbar, refresh button | `sync` | 14 |
| Toolbar, export button | `chevron-down` (rotated 90° for "save-as") OR a custom `download` — see note below | 14 |
| Filter rail, class chip blast-class dots | `dot-active` | 8 (colored) |
| Filter rail, "Reset filters" | `chevron-right` (mirrored to `back`) | 12 |
| Row, action `chat` glyph | `chat` | 10 |
| Row, chevron (expand) | `chevron-down` | 14 |
| Row, blast-class badge (no glyph — text only) | — | — |
| Detail panel, close × | `close` | 14 |
| Detail panel, copy-intent | `chevron-right` mirrored? **No.** Copy uses no icon — the label is "Copy intent" mono-uppercase 11. (MOAT §4 #2.) | — |
| Detail panel, copy-prompt | (same — text only) | — |
| Detail panel, export-row | `chevron-right` | 12 |
| Detail panel, hash chain ✓ | `check` | 12 (synapse) |
| Detail panel, hash chain ⚠ | `warning` | 12 (danger) |
| Scrubber, playhead halo | (CSS-rendered, no glyph — a 14px glow circle) | — |
| Toolbar, chain-badge dot | (CSS-rendered, no glyph) | 6×6 dot |
| NavRail `audit` glyph | `audit` | 16 (already exists — document-with-check-inside) |

> **Note on `download` glyph:** `icons.ts` does not currently ship a
> `download` icon; the closest fit is `chevron-down` rotated or the
> `sync` arrow rotated 180°. Per MOAT §4 #2, the spec does **not**
> invent a new icon — Phase 4 must add `download` to `icons.ts` if the
> toolbar's export button warrants its own metaphor, or reuse
> `chevron-down` rotated.

### 7.8 `<Button />` — already exists, reused

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

- **primary** — Never. The Audit is a read-only forensic surface; no
  primary action exists. (The toolbar's export triggers the
  global ConsentModal flow, not a primary button.)
- **secondary** — Filter chips (paper-2 fill, hairline).
- **ghost** — "Reset filters", "Load older", copy/export buttons in
  the detail panel footer (mono-uppercase link with a 1px underline
  that grows from the center on hover per MOAT §5.2).
- **danger** — Never. There is no destructive action on this route
  except the kill-switch (which lives in the titlebar / NavRail).

### 7.9 `<Tooltip />` — to be created per MOAT §2.9

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

- Toolbar's `↻ refresh` button ("Refresh chain").
- Toolbar's `⤓ export` button ("Export filtered chain as JSONL —
  triggers consent").
- Toolbar's `Chain intact` / `Chain broken` badge ("The HMAC chain
  verifies." / "The HMAC chain is broken at row #N. The rows beyond
  this point should not be trusted.").
- Filter rail's blast-class chips ("READ = observation; WRITE =
  local mutation; NETWORK = outbound call; DESTRUCTIVE = hard-to-undo,
  requires consent").
- Filter rail's verdict segmented control ("Allow = gatekeeper
  allowed; Block = gatekeeper blocked; Prompt = user was asked
  first; Error = system error").
- Row chevron ("Expand details").
- Scrubber playhead ("{timestamp} · {N events since this point}").

The tooltip is **never** used as a celebration; MOAT §4 #6 forbids
that.

### 7.10 `<ErrorState />` — to be extracted per MOAT §2.6

```ts
let {
  head: string,         // "We couldn't read the chain."
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
route, not inline. The current `Audit.svelte:76-86` inline error
block is deleted; the route uses the shared component.

### 7.11 `<EmptyState />` — to be extracted per MOAT §2.4

```ts
let {
  what: string,
  why: string,
  action: Snippet | null,
  class?: string,
} = $props();
```

The Audit uses `<EmptyState />` for S1 (no action snippet — the empty
state is quiet) and S3 (two ghost buttons as the `action` snippet).
After this route ships, `Chat.svelte`, `Hub.svelte`, `Skills.svelte`,
`Settings.svelte`, `Replay.svelte` empty states should also migrate to
it (out of scope for Phase 4 — they can continue using inline patterns
until a cleanup pass).

### 7.12 `<ConsentModal />` — already exists, reused

The Audit does **not** mount its own consent modal. Export triggers
the existing global `ConsentModal` (per `Shell.svelte:246` and
`APPFLOW.md §6.7`), which gates all destructive actions through the
gatekeeper. The Audit requests the export, the gatekeeper decides
(usually: `require_consent` for a disk write), and the modal opens
once on the global layer.

---

## 8. Data Fetched

The route's IPC contract is **five** methods (one existing, four to
add). Signatures come from
`/Users/sahajpatel/synaptic/app/web/frontend/src/lib/ipc/client.ts`
and `…/ipc/types.ts` (existing).

### 8.1 `ipc.auditList(params)` — existing

RPC: `audit.list`. Returns `AuditEvent[]`. Used to populate the
Timeline.

```ts
// Existing: client.ts:216
auditList(p: AuditListParams = {}): Promise<AuditEvent[]>

interface AuditListParams {
  limit?: number          // default 100
  offset?: number         // default 0
  since?: string          // ISO 8601, optional
  action?: string         // exact-match `action`, optional
  level?: 'info' | 'warn' | 'error'  // exact-match `level`, optional
}
```

Cache policy: fetched on mount and on user-triggered refresh (toolbar
`↻` button or `⌘R`). The route **does not** poll — the live entries
arrive via SSE channel (see §8.3). The route does **not** show a
loading state during a manual refresh if data is already on screen;
the new rows fade in as they resolve.

Error policy: failed calls land in `<ErrorState />` (S4). The route
does **not** fall back to cached data on error — the user gets an
honest "we couldn't read it" message.

### 8.2 `ipc.auditExport(params)` — **to add**

RPC: `audit.export`. Returns `{ path: string; count: number }`.
Triggered by the toolbar's `⤓ export` button (and `⌘E` shortcut when
focus is in the timeline).

```ts
// To add: client.ts
auditExport(p: AuditExportParams): Promise<{ path: string; count: number }>

interface AuditExportParams {
  // The filtered view (taken from the current Filters snapshot)
  since?: string
  until?: string
  blast_classes?: Array<'READ' | 'WRITE' | 'NETWORK' | 'DESTRUCTIVE'>
  verdict?: 'all' | 'allow' | 'block' | 'prompt' | 'error'
  apps?: string[]
  models?: string[]
  search?: string
  // Where to write the JSONL file
  destination?: string   // defaults to ~/Downloads/condura-audit-<date>.jsonl
}
```

This export is gated by the global `ConsentModal` (gatekeeper policy
classifies disk-write of personal forensic data as
`{ kind: 'file.write', blast_radius: 'WRITE' }` which requires
consent). The route does **not** bypass the gatekeeper.

On the daemon side, this method is a thin wrapper over
`audit.Log.Export(ctx, query, destination)` — read rows matching the
filter, write them as JSONL (one JSON object per line), return the
path.

### 8.3 `ipc.auditSubscribe()` — **to add** (SSE channel)

The current IPC client has an SSE channel already (`client.ts:522-560`),
emitting named events (`halt`, `spend_warning`, `audit`, `stream`).
The current `audit` named event is emitted but does not carry a
typed payload — Phase 4 must extend it to carry the typed
`AuditEvent` payload so the Audit store can directly push new
entries.

```ts
// Existing SSE channel: client.ts:636
// To add: typed payload for the 'audit' event
ipc.on('audit', (event: AuditEvent) => {
  audit.appendLiveEvent(event)
})
```

The `audit.subscribe` RPC may not exist as a named method — the SSE
channel's per-event subscription is owned by the client. The route
**subscribes** to the channel on mount and **unsubscribes** on
destroy (no leaks when the route is unmounted by hash navigation).

Cache policy: the live entries are appended to the in-memory
`audit.events` array; they do **not** trigger a full re-fetch. The
toolbar's meta strip updates from `events.length` (e.g., `124 rows` →
`125 rows`).

Error policy: if the SSE channel disconnects (transient network drop),
the route polls `audit.list` once on `onreconnected` to refill any
missed entries, then resumes streaming. The user sees no UI change
during a transient drop (the `Reading the chain` loading state from
S2 was meant for first-load only; transient re-fetches do not surface
the loader — the existing rows stay visible).

### 8.4 `ipc.auditVerifyIntegrity(params)` — **to add**

RPC: `audit.verifyIntegrity`. Returns `AuditIntegrityReport`. Used by
the toolbar's `Chain badge` (`V` keyboard chord) and by the detail
panel when it loads.

```ts
// To add: client.ts
auditVerifyIntegrity(p: AuditVerifyParams = {}): Promise<AuditIntegrityReport>

interface AuditVerifyParams {
  since?: string          // ISO 8601, optional
  blast_classes?: Array<...>  // optional
}

interface AuditIntegrityReport {
  ok: boolean
  broken_at_id: number | null
  reason: string | null   // e.g. "prev_hash mismatch at row 42"
  rows_verified: number
  rows_skipped: number
  duration_ms: number
}
```

On the daemon side, this method walks the chain in order, recomputing
the HMAC of each row and comparing against its stored HMAC. The first
row that fails is `broken_at_id`. This is **read-only** — it does
not write to the log, it does not modify the log, and it has no
consent requirement.

### 8.5 `ipc.auditFacetCounts(params)` — **to add** (optional)

RPC: `audit.facetCounts`. Returns the count of distinct `app`,
`model`, `blast_class`, and `verdict` values in the current filtered
set, so the Filter rail's right-side counts (the "App: 23" mono
numbers) are accurate.

```ts
// To add: client.ts (optional — can be derived client-side)
auditFacetCounts(p: AuditListParams = {}): Promise<AuditFacetCounts>

interface AuditFacetCounts {
  apps: Array<{ name: string; count: number }>
  models: Array<{ name: string; count: number }>
  blast_classes: Record<'READ'|'WRITE'|'NETWORK'|'DESTRUCTIVE', number>
  verdicts: Record<'allow'|'block'|'prompt'|'error', number>
  total: number
}
```

If this RPC is not implemented in Phase 4, the route falls back to
deriving counts client-side from the in-memory event array. The
filter rail UI does not block on this RPC.

### 8.6 Data shape enrichment (Phase 4 must add to `AuditEvent`)

The current `AuditEvent` interface in
`/Users/sahajpatel/synaptic/app/web/frontend/src/lib/ipc/types.ts:226-235`
is:

```ts
interface AuditEvent {
  id: number
  ts: string
  actor: string
  action: string
  app: string
  level: 'info' | 'warn' | 'error'
  result: 'allow' | 'block' | 'prompt'
  message: string
}
```

To support this spec, the interface MUST be enriched with the
following fields (all already populated server-side in
`internal/audit/log.go:54-77`, just not yet forwarded over the
JSON-RPC layer):

```ts
interface AuditEvent {
  id: number
  ts: string                       // ISO 8601 with millisecond precision
  actor: string
  action: string                   // e.g. "shell.exec", "llm.chat"
  app: string                      // e.g. "com.apple.Terminal"
  level: 'info' | 'warn' | 'error'
  result: 'allow' | 'block' | 'prompt'
  message: string                  // the agent's articulated intent

  // Enrichments (Phase 4):
  blast_class: 'READ' | 'WRITE' | 'NETWORK' | 'DESTRUCTIVE'
  verdict: 'allow' | 'block' | 'prompt' | 'error'
  target_app?: string
  target_url?: string
  path?: string
  command?: string
  body?: string
  consent_result?: string
  model?: string                   // e.g. "claude-sonnet-4-5"
  provider?: string                // e.g. "anthropic"
  backend?: string                 // e.g. "claude_code"
  session_id?: string
  tokens_in?: number
  tokens_out?: number
  prev_hash?: string               // for the detail panel's hash chain
  this_hash?: string
}
```

The blast_class field is **not** the same as the existing
`result` field. `result` carries the gatekeeper verdict; `blast_class`
carries the class assigned by `internal/blastradius.Classify`. A row
with `blast_class: 'DESTRUCTIVE'` and `result: 'block'` means "the
agent tried to do a destructive action, the gatekeeper blocked it."

---

## 9. Design Decisions

These are the load-bearing calls — the place where this spec
disagrees with the current `Audit.svelte` and where it earns the
MOAT bar.

### 9.1 The Audit is a forensic surface — no decoration, ever

**The problem (per MOAT §1, the Restraint test).** The current
`Audit.svelte` has three forms of decoration that have nothing to do
with forensics:

- The eyebrow serif italic font (the audit "head" block).
- The "READING THE CHAIN" mono caps label, duplicated from
  `Skills.svelte` and `Channels.svelte`.
- The visual padding around the chain-spine (the spine is offset
  28px from the left edge — 28px of decorative gutter).

**What this spec does.** The eyebrow becomes mono caps 11 (the design
language convention), the title is display serif 28-40px (a true
hero), the toolbar is `display:flex; gap:var(--space-3)` and the
spine is offset 24px from the left edge (the design system's default
column gutter). Every glyph is a `<Glyph />`. Every animation has a
load-bearing purpose (see §5).

### 9.2 The blast-class badge, not the level chip

**The problem (per `internal/blastradius/blastradius.go` and
`MOAT §4 #4`).** The current `Audit.svelte` filters by `level`
(info / warn / error). This conflates two distinct concepts:

- **Blast class** — the *category* of the action (READ / WRITE /
  NETWORK / DESTRUCTIVE). This comes from
  `internal/blastradius.Classify`. It does not change based on the
  outcome; it tells you what *kind* of action this was.
- **Level** — the *severity* the agent's intent analyzer assigned
  (info / warn / error). This conflates "the action succeeded" with
  "the agent flagged this for attention."

Filtering only by `level` makes the user unable to find "every
DESTRUCTIVE action the agent ever considered." Filtering only by
`blast_class` makes the user unable to find "every action that
failed" or "every action that triggered consent."

**What this spec does.** Two filter dimensions, clearly named:

- **What (blast class)**: READ, WRITE, NETWORK, DESTRUCTIVE. Chips.
- **What (verdict)**: allow / block / prompt / error. Segmented.

The user can filter on both, independently. The toolbar's blast-
class histogram meta strip shows the ratio at-a-glance. The route
stops pretending `level` is the only interesting axis.

### 9.3 The detail panel is a `.c-sheet`, not a sticky-right-rail hover card

**The problem (per MOAT §2.8, the overlay taxonomy).** The current
`Audit.svelte:116-132` renders a `<aside class="detail">` as a
**sticky** element on the right side of the audit body. It is always
visible (`opacity: 0.4` at rest). Hover a row, the aside fills in
(`opacity: 1`). No animation, no slide. No scrim, no Esc-to-close
behavior. The aside has "Hover a node to see the full record" placeholder
copy when nothing is hovered.

This conflates two distinct interactions:

- **Theoread the live timeline** — events scroll by, the user is
  watching.
- **The user has selected one event** — they want the full payload,
  the hash chain, the model prompt that caused it, and the choice
  to copy / export that one event.

A sticky hover card can't do the second; it can't have its own
keyboard surface, can't be expanded to show more, can't have its own
actions.

**What this spec does.** The detail panel is a `.c-sheet` that
**only** mounts when a row is selected (Region C from §3.4). It
slides in from the right (`transform: translateX(360px)` →
`translateX(0)` over 320ms `--ease`). It has its own focus-trap.
`Esc` returns to the timeline. The previous sticky-right-rail hover
card is **deleted** (no fallback).

This is consistent with the 2026 mature pattern: details on demand,
in their own context, owned by the user (Esc/click-to-close).

### 9.4 The scrubber is the time axis, not a UI

**The problem (per `Replay.svelte:256-289`).** The current `Replay`
route has a timeline scrubber at the bottom of the route that is
**the** time axis — a 1px `<Thread />` with event dots and a playhead.
The Audit spec wants to reuse this metaphor, but with two key
differences:

- The Audit scrubber covers **all** events visible on screen, not
  just one selected event.
- The Audit scrubber is a **continuous** timeline (24h spans), not
  a discrete timeline (24 frames).

The user has been trained by `Replay` to read the `<Thread />` as
"time flows here." Reusing the same metaphor in Audit tells them:
*events are points in time, not pauses in time.*

**What this spec does.** The TimelineScrubber (§3.5) is a `.c-paper`
strip at the bottom of the route, sticky. It carries:

- A 1px `<Thread />` drawn left→right on mount (MOAT §3 — brand).
- Tick marks at minute / hour / day granularity based on the window.
- 2px event dots along the track in blast-class color.
- A 6px pollen playhead + 14px `--synapse-glow` halo.
- A faint mono `↹` hint on the right.

The scrubber collapses to a 1px hairline after 20s of idle
interaction (the breathing gesture goes to sleep). Re-entering the
route or pressing any key restores the full chrome. Per MOAT §4 #10,
the collapse communicates "this is dormant"; the expansion on key-
press communicates "you are here again."

### 9.5 The chain badge is always visible

**The problem (per MOAT §4 #6 and `internal/audit/log.go`).** The
chain is the spine of the product. If it's broken, the product cannot
honestly claim to be working. The current `Audit.svelte` has no
visual indication of chain integrity at all — the user has to
read the audit to discover the chain is broken.

**What this spec does.** A persistent `Chain intact` / `Chain
broken at row #N` badge in the toolbar's right cluster. Always
rendered. Three states:

- **`Chain intact`** (synapse filled dot + `Chain intact` mono
  caps): tooltip "The HMAC chain verifies. Every row has been signed,
  in order, since the first write."
- **`Chain unknown · last verified {ago}`** (faint): tooltip
  "Chain integrity has not been verified yet. [Verify integrity →]."
- **`Chain broken at row #N`** (danger filled dot + label):
  tooltip "The HMAC chain is broken at row #N. The rows beyond this
  point should not be trusted. [Verify integrity →]."

The detail panel re-queries `auditVerifyIntegrity` whenever it
mounts. The toolbar queries it on mount + on `V` keyboard chord +
every 10 minutes (whichever comes first) while the route is mounted.

This is consistent with `internal/audit/log.go` which already
implements `Replay.VerifyIntegrity` (per `internal/daemon/methods_phase11.go:60`).
The same logic is exposed via `audit.verifyIntegrity` (to add) so the
Audit route can call it directly.

### 9.6 Live entries arrive via SSE — the route never polls

**The problem (per `internal/audit/log.go` and IPC SSE channel).**
The current `audit` store's `refresh()` polls `audit.list` on user
demand only. Live entries (rows the agent just appended) are not
visible to the user until they refresh. This makes the Audit useless
during a long-running agent session — the user has to refresh to see
"what just happened."

**What this spec does.** The IPC SSE channel already emits a named
event `audit` (per `client.ts:587,636`). Phase 4 must:

1. Type the SSE event payload as `AuditEvent` (currently the
   payload is `unknown` — see `client.ts:636`).
2. Add `audit.appendLiveEvent(row)` to the store.
3. Wire the store to push new rows to the **top** of the timeline.
4. Trigger the slide-in glow animation per §5.5.

The toolbar's `<Thread />` re-draws from the right edge for each
coalesced batch. Rate-limit per §5.5. The user perceives the Audit
as live during agent activity.

### 9.7 Export is consent-gated, but success feels like completion

**The problem (per `MOAT §4 #6`).** The Audit has one **mutating**
user action: `Export the filtered chain as JSONL`. This is a
disk-write (file write at `~/Downloads/condura-audit-<date>.jsonl`).
Per `internal/blastradius/`, file.write is `blast_radius: WRITE`.
Per the gatekeeper policy in `~/.synaptic/policy.yaml` (the default
v0.1.0 policy), `WRITE` actions trigger
`decide: require_consent`. The user must click Allow in the global
ConsentModal for the export to land.

**What this spec does.** The toolbar's export button triggers the
global ConsentModal — same modal as the consent flow in `Chat.svelte`
and `Delegation.svelte`. The Audit does not own a modal of its own.

On success, the toolbar's `<Thread />` draws the export-success
hairline (`#7` in §5.6) — the brand's "completed" gesture, not a
celebration. The user's toast region (per `MOAT §2.9`) shows a
mono-pollen status `Exported {N} rows to ~/Downloads/condura-audit-<date>.jsonl`
for 4 seconds, then fades.

Per MOAT §4 #6, no "Exported!" celebration. Per MOAT §4 #5, no
"Hope you find what you're looking for!" copy. The status line is
the entire signal.

### 9.8 The empty state is honest — no "let's make a row!" CTA

**The problem (per `MOAT §2.4`, `MOAT §4 #5`, `MOAT §4 #6`).** The
current `Audit.svelte:88-97` empty state copies:

> "The chain is quiet." / "Nothing has happened yet. Every action the agent
> takes will land here, HMAC-chained."

This is **almost** right but reads as "more events are coming, wait
for them." That's how mature forensics tools work — they wait. But
the copy could be tighter; it could be honest about *why* the chain
is quiet (the user hasn't done anything yet).

**What this spec does.** Three lines (per MOAT §2.4):

> **What's here.** The HMAC-chained event log. Every action the agent
> takes — chat turns, file edits, clicks, network requests, consent
> tickets — lands here as one row, in order, signed.
>
> **Why it's empty.** Nothing has happened yet. Every row is one fact;
> the first row will appear the first time the agent acts.
>
> **Next action.** *(no button — the empty state is honest about
> absence.)*

The "no button" is **deliberate** per MOAT §4 #5. There is no "Try
something!" CTA in a ledger. There is no "Write a test event." The
ledger logs what happens. The empty state is the structure of
absence, not a prompt for action.

### 9.9 The chain spine's color follows the highest-risk row in view

**The design language move.** The vertical chain spine (the 1px
thread on the left edge of the Timeline) is normally `--synapse`
(green). But **when** the viewport contains a row of class
`DESTRUCTIVE`, the spine from the top of the viewport to that row's
vertical position is rendered in `--danger`. From that row to the
bottom of the viewport (and if subsequent rows are DESTRUCTIVE, all
the way down), the spine stays `--danger`. When the user scrolls
past the last DESTRUCTIVE row, the spine returns to `--synapse`.

This is a **visual ratio histogram** in the spine itself. The user
sees the danger-zone at a glance; their eye is drawn there first.
The implementation is a single CSS variable update on the spine
container, computed once per viewport intersection.

This is **the only color-on-state animation** on the route's spine,
and it answers "what is this communicating?" in two words: *danger
happened here.*

---

## 10. What this spec deletes from the current `Audit.svelte`

The current `Audit.svelte` (599 lines) has the following per-file
content that Phase 4 must remove in the atomic diff:

| # | Current `Audit.svelte` | Replaced with | Why |
|---|---|---|---|
| 1 | Eyebrow italic "— Forensics · HMAC-chained · append-only" rendered in italic display. | Mono caps 11 eyebrow (per the design language). | Italic display is reserved for the title + empty/error state. The eyebrow is the chrome convention. |
| 2 | Display title "Every action, on a thread." (line 43). | **Kept** — this title is correct. | It's the brand voice. Per MOAT §1, restraint: only the load-bearing flourish survives. |
| 3 | Free-text `<input>` for "filter by action…" + 4 level-chip toggles (lines 51-64). | Full filter rail (Region A). | The current filter is `action` (exact-match text) + `level` only. The new filter has 4 groups (When / What / App / Model). |
| 4 | `.chain` block — vertical `<Thread />` spine with event `<button>` rows (lines 67-114). | New `<FilterRail />` + new `<TimelineRow />` (per §7). | The chain metaphor stays; the layout is restructured into three columns. |
| 5 | Sticky `<aside class="detail">` right rail (lines 116-133). | New `<AuditDetailPanel />` as a `.c-sheet` (per §7.3). | Per MOAT §2.8, sticky hover-cards and sheets are different things. |
| 6 | Inline `<div class="err-state">` block (lines 75-86), copy-pasted from Chat/Skills/Channels. | Shared `<ErrorState />` component (per MOAT §2.6). | Per MOAT §1.2, "Extract ErrorState. Delete the four copies." Audit is the fourth. |
| 7 | Per-card `transition:` list (lines 413-417), 5 declarations. | Global `.tactile` class (per MOAT §2.7). | Components own meaning, not timing. |
| 8 | Pagination buttons ("← prev", "next →") at the bottom (lines 135-139). | Infinite scroll backwards (newest first, `[ Load older ]` ghost button). | Pagination is the wrong model for "what just happened" — the user wants newest, then older on demand. |
| 9 | The chain-spine `padding-left: 28px;` offset (line 261). | 24px (`var(--space-6)`), the design language's default column gutter. | The audit has been using a decorative 28px gutter that the rest of the design system does not match. |
| 10 | `.dot` (line 434-439) with `box-shadow: 0 0 6px currentColor` glow. | **Kept** — the dot glow IS the audit's signature. | Per MOAT §1, the chain-dot glow is the one allowed flourish on this screen. |
| 11 | `.node` hover with `transform: translateX(2px)` (line 423). | **Kept** — the row-level motion is honest. | Per MOAT §1.5, scale and 3D rotation are deleted. translate is not. |
| 12 | `.err-hair-draw` keyframe (lines 393-395). | **Kept** — used by the global ErrorState. | The existing pattern is correct. |

---

**This document is the contract.** Phase 4 implementation in
`Audit.svelte` + `stores/audit.svelte.ts` + the four new IPC
methods + the SSE typed-payload extension all flow from this spec.
When code disagrees with this spec, the spec is correct and the code
needs adjustment.
