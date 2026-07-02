# SCREEN_SHELL.md — Condura Shell · Screen Architecture

> **The contract.** This spec defines the layout, content slots, state matrix,
> motion, keyboard, component composition, data, and design decisions for the
> **Condura Shell** — the desktop window that holds `Titlebar` + `NavRail` +
> `Main` (the active route) + an optional contextual `RightRail` + a bottom
> `StatusBar`. Phase 4 will implement against this document; if a component
> disagrees, the component is wrong, not the spec.
>
> **Audience.** Phase-4 implementer. Designer for review.
>
> **Source-of-truth files** (read these alongside this spec):
> - `app/web/frontend/src/lib/condura/Shell.svelte` — current implementation.
> - `app/web/frontend/src/lib/condura/condura.css` — design tokens.
> - `app/web/frontend/src/lib/condura/APPFLOW.md` — full user-flow contract.
> - `app/web/frontend/src/lib/condura/MOAT.md` — premium-quality rules.
> - `app/web/frontend/src/lib/condura/TEARDOWN.md` — onboarding patterns.
>
> ### 0.0 Contract caveat (read first)
>
> The brief names five north-star docs: `DIRECTION.md`, `TEARDOWN.md`,
> `DESIGNLANG.md`, `APPFLOW.md`, `MOAT.md`. As of this writing, only **three
> exist on disk** in `app/web/frontend/src/lib/condura/`:
> `APPFLOW.md`, `MOAT.md`, `TEARDOWN.md`. `DIRECTION.md` and `DESIGNLANG.md`
> are referenced but absent. This spec is grounded in the three that exist
> plus the live `Shell.svelte`. If `DIRECTION.md` / `DESIGNLANG.md` are
> later produced, this spec should be re-read for conflicts; the material
> below already reflects most of what those docs would likely contain
> (thread as signature, anti-patterns, motion grammar) since they are
> encoded in the design tokens + `MOAT.md` + `TEARDOWN.md`.
>
> ### 0.1 What the Shell IS and IS NOT
>
> **IS:** the persistent desktop window that hosts every post-Ritual surface
> in the Condura product. Titlebar + NavRail + active route + (optionally)
> right rail + statusbar + global overlays (consent, kill switch, command
> palette, quick prompt). It mounts after `Ritual.svelte` completes (or if
> the Ritual decides the user can skip straight in), and it lives until the
> window closes.
>
> **IS NOT:** a router shell with internal tabs. Every route is a full
> surface in the Main region (we do not stack routes inside the shell).
> Not a wizard. Not a settings panel. Not a chat. Every nav item swaps the
> Main region via `{#key route}` re-mount + `.route-enter` blur-in.

---

## 1. LAYOUT GRID

### 1.1 Region map (annotated)

```
┌────────────────────────────────────────────────────────────────────────────┐
│ (1) Titlebar · 44px tall · spans all columns                              │
│     wordmark · TitlebarThread · DynamicIsland · ⌘K · ☼/☾ toggle           │
├──────────────┬────────────────────────────────────────┬────────────────────┤
│              │                                        │                    │
│ (2) NavRail  │   (4) Main · the active route           │ (5) RightRail     │
│  240 px      │        (1fr on ≥1440, 1fr + RR off    │   320 px           │
│  (3) below   │         on 1024–1439)                   │   (contextual ·    │
│              │                                        │    sheet on        │
│              │                                        │    small screen)   │
│              │                                        │                    │
├──────────────┴────────────────────────────────────────┴────────────────────┤
│ (3) StatusBar · 28 px tall · spans cols 2–4 (main + right rail)            │
│      per-turn stopwatch · git branch · token count · cost · model badge    │
└────────────────────────────────────────────────────────────────────────────┘
```

Overlay mode (when daemon reports an active `overlay` session — a
floating quick-prompt / interview / consent overlay is owned by the OS
above the shell): the NavRail hides and the grid collapses to a single
column (`grid-template-columns: 1fr`). The titlebar stays.

### 1.2 CSS Grid spec

```css
.shell {
  display: grid;
  grid-template-columns: var(--rail-w) minmax(0, 1fr) var(--rightrail-w, 0px);
  grid-template-rows: var(--titlebar-h) minmax(0, 1fr) var(--statusbar-h);
  grid-template-areas:
    "titlebar  titlebar   titlebar"
    "navrail   main       rightrail"
    "navrail   statusbar  statusbar";
  height: 100vh;
  height: 100dvh;                 /* dynamic vh on browsers that support it */
  width: 100vw;
  position: relative;
  overflow: hidden;
  background: var(--surface);
  isolation: isolate;             /* grain layer stays inside this stack */
}

.shell.overlay { grid-template-columns: 1fr; }   /* navrail hidden */
```

Token shortcuts (declared on `:root` next to the shell, or in `condura.css`):

```css
:root {
  --rail-w:       240px;
  --rightrail-w:  320px;          /* 0 when contextualRightRail === 'none' */
  --titlebar-h:   44px;
  --statusbar-h:  28px;           /* 0 when statusbar is collapsed by breakpoint */
}
```

The 320-px right rail grid track collapses to `0px` when there is no
contextual rail for the active route, so the Main region naturally fills
the space. This is **not** a `display: none` — the grid stays rigid,
eliminating layout jitter when routes with and without a right rail swap.

### 1.3 Per-region grid placement

| Region     | grid-area         | z-index              | Pointer events |
|------------|-------------------|----------------------|----------------|
| Titlebar   | `titlebar`        | `--z-sticky` (100)   | yes            |
| NavRail    | `navrail`         | `--z-sticky` (100)   | yes            |
| Main       | `main`            | `--z-base` (0)       | yes (route)    |
| RightRail  | `rightrail`       | `--z-base` (0)       | yes            |
| StatusBar  | `statusbar`       | `--z-sticky` (100)   | yes            |
| `.paper-grain` | n/a (absolutely positioned `inset: 0`) | `--z-base` | none |
| Global overlays (CommandPalette, QuickPromptOverlay, KillSwitchOverlay, ConsentModal) | `position: fixed` at top of tree | `--z-modal` (2000) / `--z-toast` (3000) | yes |

The titlebar lives at `z-sticky` (above Main) so the bending
`TitlebarThread` never has to fight route content for paint order. The
right rail sits at base because it lives *inside* the content region
contextually, never above it.

### 1.4 Breakpoints

The Shell is designed for ≥1024 px. Below 768 px it shifts to "phone"
ergonomics, gated behind `condura.css` media queries (no
component-local media queries — per MOAT §2.3, components never redeclare
reduced-motion, and the same applies to breakpoints).

| Breakpoint | Rail | Main | Right rail | Statusbar | Notes |
|------------|------|------|-----------|-----------|-------|
| **≥ 1440 px (default)** | 240 px (label) | 1fr | 320 px (contextual) | 28 px | full layout |
| **1024–1439 px**        | 240 px (label) | 1fr | **0 px** (re-homed to a `.c-sheet` from inside the route) | 28 px | right rail slides in only on demand |
| **768–1023 px**         | 56 px (icon-only, condensed) | 1fr | 0 px | **0 px** (compacted into titlebar) | wordmark scales to 18 px; ⌘K hint hidden |
| **< 768 px**            | full-screen drawer (44 px collapsed with hamburger) | 1fr | 0 px | 0 px (compacted) | NavRail only as a drawer; touch targets ≥ 44×44 |

The breakpoint contract: a route can *request* a right-rail by writing
`contextualRightRail = 'truth-and-detail'` (or `none`). The Shell reads
this off the route module (see §3.4). If the active breakpoint says
"no right rail", the rail **never appears** — its content slides in as
a sheet from the right, owned by the route.

### 1.5 Safe areas

The Shell respects OS safe-area insets on every platform:

```css
.titlebar  { padding-left:  max(env(safe-area-inset-left,  0px), var(--space-5)); }
.navrail   { padding-top:   env(safe-area-inset-top,     0); padding-left: env(safe-area-inset-left, 0); }
.statusbar { padding-right: max(env(safe-area-inset-right, 0px), var(--space-3));
             padding-bottom: env(safe-area-inset-bottom, 0); }
```

macOS titlebar traffic-light buttons (close / min / max) sit on the very
left of the titlebar; the wordmark starts to their right with
`padding-left: 78px` (or `var(--wails-traffic-light-w, 0px)` if Wails is
driving the window chrome).

---

## 2. CONTENT SLOTS

Each region has explicit contents with exact copy and component names.
Nothing in the shell is rendered "by default" — every slot answers the
question "what is meant to live here?"

### 2.1 Titlebar (44 px tall, grid-area `titlebar`)

**Anatomy, left → right:**

| # | Slot | Component | Content |
|---|------|-----------|---------|
| 1 | Wordmark | inline (h-Condura) | `Condura` + a 5 px `--pollen` dot accent. `var(--font-display)` 22 px / `--ls-display -0.03em` / `var(--content)`. The dot is the only flourish. |
| 2 | Bending thread | `<TitlebarThread />` | The signature. 1.25-px synapse stroke at `var(--synapse)`, 6-8 dash pattern, 3-px `var(--synapse-glow)` blur underlay at 18 % opacity. A 3-px `--pollen` node rides it, `pointermove → rAF` bending at 0.06 lerp toward (x,y). `visibilitychange` + `IntersectionObserver` zero out the rAF loop in background tabs. Under reduced-motion, the loop never starts (`Shell.svelte:60`; `TitlebarThread.svelte:12`). |
| 3 | Agent status island | `<DynamicIsland phase={agentPhase} task={...} />` | Morphing pill. Width 124–188 px depending on phase (`idle` / `listening` / `thinking` / `acting` / `consent` / `error`). Border color: `var(--synapse)` for default, `var(--warn)` on `consent`, `var(--danger)` on `error`. Mono label uppercase (`letter-spacing: 0.12em`). The single point of agent truth in the chrome. |
| 4 | ⌘K hint | `.kbd-hint` (inline) | Two `<kbd>` tags reading `⌘` and `K`, monospace 11 px uppercase `0.12em`, color `var(--content-faint)`. Hover state brightens to `var(--content-soft)`. **This is decorative** — never the affordance; the chord does the work. (See MOAT §2.10 — we still display it because mature shells expose their meta-UI.) |
| 5 | Theme toggle | `<button class="theme-toggle">` with `<Glyph name={moon↔sun} size={14} />` | 28×28 round (pill), `var(--surface-card)` fill, hairline border. Hover: rotate(-12deg), border-color `var(--hair-strong)`. Press: scale(0.97) (inherits `.tactile`). The chord `⇧T` is bound to the same action (see §6). |

`.tb-controls` is `margin-left: auto` so the right cluster floats; the
thread spans the gap between the wordmark and the controls.

**States baked into the titlebar:**

- **No daemon** → `phase = 'error'` → island border `var(--danger)`,
  label reads `OFFLINE`. Thread keeps drawing (the page is alive even
  when the daemon is not).
- **Streaming** → `phase = 'thinking'` → island widens to fit
  `truncate(currentTitle, 18) + cfg.label`. The dot inside the thread
  tracks the cursor faster (no rAF budget difference; the visual reads
  as "the agent is moving").
- **Consent pending** → `phase = 'consent'`. The DynamicIsland expands
  to 188 px and a `var(--warn)` border. The titlebar does **not**
  trigger the modal — `ConsentModal.svelte` owns the modal scrim.
- **Voice listening** → `phase = 'listening'`. The thread node
  amplifies its halo to 6 px from 3 px (a single CSS transition).

### 2.2 NavRail (240 px, grid-area `navrail`)

**Anatomy, top → bottom:**

| # | Slot | Component | Content |
|---|------|-----------|---------|
| A | Brand label | `.rail-brand` | `Condura · v0.1`, monospace 10 px / 0.22em / uppercase / `var(--content-faint)`. Padding `0 var(--space-3) var(--space-4)`. |
| B | Nav list | `{#each ITEMS}` | 10 routes in this fixed order: **Chat · Hub · Skills · Sync · Audit · Replay · Channels · Delegation · Settings · About**. Each is a `<button class="nav-item">` with `<Glyph name=… size=18 />` + `<span class="ni-label">`. |
| C | Footer | `.rail-foot` | `<Pulse phase="idle" size={8} />` + `Local · Ollama` (mono 10 px / 0.14em / uppercase / `var(--content-faint)`) + a 6-px `--pollen` "you" dot `margin-left: auto`. Border-top 1 px `var(--hair)`. |

**Nav item states:**

- **Default**: `color: var(--content-mute)`, opacity 0.7 on the glyph.
- **Hover**: `color: var(--content); background: var(--surface-card); transform: scale(1.04)`. Glyph opacity 1. (Transitions at `--dur` for color/bg, `--dur-fast` for transform.)
- **Active**: `color: var(--content);` glyph `color: var(--synapse); opacity: 1`. A 2-px synapse `::before` scales from `scaleY(0)` to `scaleY(1)` (`drawv` keyframe, `--dur-slow`) — this **is** the Thread signature: "you are here" is a drawn thread, not a static bar (see MOAT §3).
- **Active "you" dot**: a 5-px `--pollen` `::after` with a 4-px box-shadow halo. Reserved for the active row.
- **Focus-visible**: `box-shadow: var(--pollen-halo)` (per `.nav-item:focus-visible` line 134–137). The halo tracks the 10-px radius — pollen round, synapse inset.

**Z-order within the rail:** brand → items → footer. The active item's
`::before` lives in the rail's own stacking context, drawing on top of
the row's background but below its label.

**Badging (future, not v0.1.0):** nav items may carry a `<Pulse>` orb
(e.g., a green dot when Channels has a degraded channel). v0.1.0 leaves
the foot Pulse as the only status indicator on the rail.

### 2.3 Main (grid-area `main`, variable)

The active route's component renders here. Routes:

| Route | Component | What it owns |
|-------|-----------|--------------|
| `#/` (default) | `<Chat />` | composer, message list, model select, empty-state hero |
| `#/hub` | `<Hub />` | 3-D bookshelf, detail sheet (right rail when ≥1440) |
| `#/skills` | `<Skills />` | local skills index, detail sheet |
| `#/sync` | `<Sync />` | garden canvas, peers rail, pending pairing card |
| `#/audit` | `<Audit />` | HMAC-chained thread, detail pane (right rail) |
| `#/replay` | `<Replay />` | scrubber, frame view, integrity badge |
| `#/channels` | `<Channels />` | 5 channel rows, signal-bars metaphor |
| `#/delegation` | `<Delegation />` | constellation canvas + live panel (right rail always on) |
| `#/settings` | `<Settings />` | flowing document |
| `#/about` | `<About />` | colophon, the 7 invariants ledger |

Each route is mounted via `{#key route}` so a navigation re-mounts the
component cleanly (no stale state) and the `.route-enter` blur-in fires
(see §5.2). Routes that need a right rail declare it through a
`contextualRightRail` prop or via reading the active route in
`Shell.svelte`.

**No-loading illusion.** Per APPFLOW §6.1 + §2.1.6 ("configure, not
comply" + "smooth is honest"), Main never renders a centered spinner.
It renders the route's empty/skeleton state (already populated with
its shape), then hydrates when the IPC resolves. If the IPC fails, the
route renders its `ErrorState` (per MOAT §2.6: italic serif headline
+ "what failed" / "likely cause" / "next action").

### 2.4 RightRail (320 px, grid-area `rightrail`, contextual)

The rail is **conditionally rendered per route**, not always present.
Width: 320 px on ≥1440 px. On 1024–1439 px the grid track collapses to
0 (the content re-homes to a `.c-sheet`). On <1440 px the rail
**never shows** — the content uses an inline `.c-sheet` (slide from
right edge, 520 ms ease, focus-trapped, Esc closes).

| Route | Right-rail content |
|-------|---------------------|
| Chat | **Conversations list** — recent threads (10 most recent, sorted by `last_touch DESC`); each thread is a row with a 14-px serif title (1-line clamp), mono 10-px timestamp (`LAST · 14:32`), and a tiny `glyph:star` if bookmarked. A `+ New conversation` button at the top opens the QuickPromptOverlay. |
| Hub | **Skill detail sheet.** Spine metadata + description + Install button. (Sheet also exists in-route at smaller breakpoints.) |
| Skills | **Skill detail panel.** Procedure preview, `Run` + `Improve` affordances, versioned metadata. |
| Sync | **This device's metadata.** Name, public key fingerprint, paired devices count, "Pair new device" CTA. |
| Audit | **Event detail pane.** Selected event's full record: actor, blast radius, Gatekeeper decision, free-text decision reasoning. |
| Replay | **Frame index + decision receipt.** Frame number + a 1-line "I decided to …" italic sentence. |
| Channels | **Channel detail panel.** Signal-bars legend, last 10 messages, message-composer placeholder. |
| Delegation | **Live panel.** Pending Gatekeeper tickets, status pills (pending / approved / denied / executed), Approve / Deny / Run buttons. |
| Settings | **Section nav.** Anchor-list of every section in the Settings flowing document. (Already a vertical sub-list inside Settings when on `<1440 px`.) |
| About | **Empty.** Column rendered with `min-width: 0` and no content; the grid track is 0 anyway. (No right rail.) |

The right rail's scroll behavior is `overflow-y: auto`; it never
horizontally scrolls. Its background is `var(--surface-card)` (one
level lifted from the page) — see MOAT §4.3 ("no glassmorphism
unless elevation is earned") — so it reads as a slightly recessed
panel, never a glass overlay.

### 2.5 StatusBar (28 px, grid-area `statusbar`, bottom row)

The bottom strip — present on ≥1024 px. On 768–1023 it collapses into
the titlebar (the right cluster of the titlebar gains a status segment);
below 768 it is gone (system-level presence is delegated to the menu
bar / OS).

**Anatomy, left → right, mono 11 px uppercase:**

| # | Slot | Content |
|---|------|---------|
| 1 | Stopwatch | `00:42.118` — a per-turn stopwatch. Resets on `conversation.start`. Freezes on `conversation.idle` with a trailing `.idle` glyph. The user sees agent thinking time without checking a stats panel. |
| 2 | Git branch (optional) | `⎇ main` when `cwd` is inside a git repo. Press to enter a `<BranchPopover>` — but this only renders on `hover`, not as a clickable chip when no repo is detected. Hides entirely outside git repos. |
| 3 | Token count | `↓ 1.2k ↑ 380`. In + out tokens for the active conversation. Mono numeric, faint ink. |
| 4 | Cost | `~$0.014`. Computed from `conversation.lastCost`. Format: `~$X.XXX` (always three decimals so the column doesn't jitter as the run progresses). |
| 5 | Model badge | The active model + provider (e.g., `anthropic · claude-sonnet-4.5`). Mono, faint. Click → opens Settings → Power. |

**State semantics in the StatusBar:**

- **Idle**: stopwatch frozen at last value; token + cost show `—`
  when no conversation is mounted; otherwise show last-known values.
- **Streaming**: stopwatch ticks at 60 ms cadence; tokens animate
  upward as bytes arrive; cost grows linearly. The StatusBar is one of
  two surfaces that shows real-time progress (the other is the
  `stream-bar` in Chat).
- **Error**: stopwatch freezes; cost stops growing; a `--danger`
  hairline draws in beneath the bar (the Thread signature — the
  "this is now finished (and finished badly)" gesture).
- **Energy: low** (`data-energy="low"` on `:root`): stopwatch ticks
  once per second, not per frame (the cadence drops so we conserve
  battery).

**Why a bottom strip at all?** TUI precedent (Claude Code / Codex /
OpenCode status bars). The four data points (time · git · tokens ·
cost) are the meta-answers a developer wants *while* the agent is
working, without leaving the route. The Mono-strip keeps the
visual weight negligible — it's a footnote, not a HUD.

### 2.6 Grain overlay (`.paper-grain`)

Positioned `inset: 0` on top of the shell, `pointer-events: none`,
`opacity: var(--grain-opacity)` (0.50 light / 0.42 dark), `z-index:
var(--z-base)`. The texture comes from a `data:image/svg+xml` SVG
feTurbulence. The whole grain layer is hidden under:

- `prefers-reduced-motion: reduce` (via the single block in
  `condura.css:469–476`).
- `:root[data-energy='low']` (via the same).
- `display: none` when the daemon reports `kernel.disable_grain` (the
  user-level toggle in Settings → Appearance → Grain intensity).

### 2.7 Overlay mode

When `overlay.active === true` (a floating overlay session — quick
prompt, consent, kill switch — is shown), the rail hides via the
`.shell.overlay` modifier (`grid-template-columns: 1fr`). The
titlebar stays at `--z-sticky` so the user can always find home. The
Main region fills the full width.

The Rail never re-collapses mid-overlay; the collapse happens at
mount + at every `overlay.active` flip. When the overlay dismisses
the rail returns with a 320 ms slide-in from `translateX(-8px)` →
`0` + opacity 0 → 1 (the tactile return).

---

## 3. STATE MATRIX

This is the spec for what the **shell as a whole** shows in each
reachable state. Per-route states (loading, empty, error in
`Skills.svelte`) are owned by APPFLOW.md §6 and §7 — the table
below covers the cross-cutting states.

The default for unknown errors is **honest degradation** — never a
dead wall (APPFLOW I4: "every state is reachable").

### 3.1 First-time-in-shell

The very first paint after `Ritual.svelte` dissolves.

**Visual reading:**

- Titlebar: wordmark + thread + `DynamicIsland phase="idle"` reading
  `IDLE · LISTENING`. ⌘K hint visible, theme toggle visible.
- NavRail: `Chat` is the active row (it has the synapse `::before`
  thread + the pollen "you" dot from the start — no entrance
  flash; just the resting state).
- Main: `<Chat />` mounted. Empty state is the canonical hero
  (`"Your computer, alive."` in `var(--font-display)` 48 px, plus
  the 14-mote garden in the soft `$paper-3` corner gradient).
  Composer pinned at the bottom. Model `<select>` is rendered
  with whatever `ipc.providersList()` returned (zero or more; if
  zero, the select is hidden — see §3.3).
- Right rail: Chat's **Conversations list**, which on a first-time
  install contains 0 threads. The rail renders a one-line whisper
  in the `mono eyebrow` style: "No conversations yet." with a
  `+ New conversation` button below. Polling for `db.thread.list`
  kicks off once per minute (see §7).
- StatusBar: stopwatch at `00:00.000`; tokens/cost both `—`; model
  badge empty until a provider is picked. Git branch hidden (we
  don't know cwd yet).

**Animation** (per MOAT §3, §5): Shell fades in over 320 ms (opacity
0→1 + 4 px translateY → 0), NavRail items stagger 60 ms each (left
to right in a top-down cascade), Main blurs in via `blur-in`
(--dur-slow). The thread's `IntersectionObserver` starts on mount.

### 3.2 Daemon-unreachable

Triggered when `daemon.connected === false` (any `ipc.X()` rejects).

**Visual reading (shell-level):**

- `DynamicIsland phase="error"`. Border `var(--danger)`; label
  `OFFLINE`. The Thread keeps drawing (the page is alive even when
  the daemon is not — this is the MOAT §3 commitment to the
  signature gesture regardless of state).
- NavRail: still navigable. Items route normally; the Shell does
  not freeze.
- StatusBar: stopwatch frozen at last value; tokens/cost freeze at
  last-known. A `--danger` 1-px hairline draws in beneath the
  StatusBar (the Thread-drawn "this is now (brokenly) finished"
  gesture — see §5.3).
- Main: the route renders its route-level error state (per APPFLOW
  §6.1): italic serif headline "We couldn't reach the daemon." + a
  hairline that draws left→right (the existing `err-hair-draw`
  pattern), plus a `Try again` button that calls
  `ipc.daemonPing()` and the three MOsat §2.6 lines ("Connection to
  daemon" / "daemon was restarted" / "Restart daemon" / "Open
  Settings → Connection").
- Right rail: Chat's list shows the previous threads (if any) with
  `data-state="stale"` and a top-of-list whisper "Showing cached
  conversations."

**The Titlebar stays fully interactive**, including the ⌘K palette
(which works without the daemon — it routes the user, not the
agent).

### 3.3 No-API-key, No-Ollama (and both)

Trigger: zero providers returned from `ipc.providersList()` OR no
Ollama on `$PATH`.

**Visual reading (shell-level):**

- `DynamicIsland phase="idle"`. (No error — this is a configuration
  state, not an outage.)
- NavRail + StatusBar render normally.
- Main: `<Chat />` empty state, **but** the composer shows a
  beneath-the-input whisper in mono:
  `No providers configured. Open Settings → Power to add an API key or install Ollama.`
  Send is enabled (clicking it returns the daemon's "no provider"
  rejection, which renders the route-level streaming-error block).
- Right rail (Chat): unchanged (conversations list, possibly
  empty).
- StatusBar: model badge reads `no model`. Click → Settings → Power.

**Settings route** (if user navigates there): Power section's
"Default model per provider" list is empty; the empty state is the
italic serif "No providers configured. Add an API key to begin." per
APPFLOW §6.2.

### 3.4 Onboarding-incomplete (Ritual active)

Trigger: `firstRunStatus.complete === false` OR `!seen`
(`localStorage['condura-ritual-seen']` missing) OR the
`!daemonComplete` flag from `Shell.svelte:81–102`.

**Visual reading (shell-level):**

- **The Shell does not render.** `<Ritual />` is mounted **full-bleed,
  frameless, pre-window** (it IS the window). On
  `completeOnboarding(routeHash)`, the ritual gets
  `.dissolving` (700 ms opacity → 0 + blur 8 px) and the Shell
  renders underneath. `Shell.svelte:187–192` shows this is the
  intended pattern.
- All IPC store subscriptions, polling, and the titlebar thread
  pause while the ritual owns the window. `initStores()` and the
  three startPolling calls are still invoked (`Shell.svelte:60–80`),
  but they fail-soft through the `try/catch` boundary.
- The dev affordance `Shift+O` re-opens the Ritual at any time
  (marked for removal before ship per the inline comment).

### 3.5 Kill-switch-active

Trigger: `halt.state.halted === true` (the user pressed
`Cmd+Shift+Escape` or the watchdog tripped).

**Visual reading (shell-level):**

- `<KillSwitchOverlay reason={halt.state.reason ?? 'user requested'} onresume={handleResume} />`
  mounts at top of tree (z `--z-modal`). 86 % `var(--surface-ink)`
  scrim + 8 px backdrop-blur. See APPFLOW §6.4 for the card.
- **The Shell stays paintable behind the scrim** but every input is
  dead (the scrim catches pointer events). The thread keeps
  drawing.
- This state has **no auto-dismiss**. The modal stays until the
  user explicitly mints the resume ticket (which requires CLI-side
  confirmation — "auto-recovery is the enemy" per CLAUDE.md §2.1
  invariant #4).
- `DynamicIsland phase="error"` (`HALTED` label, danger border)
  for the duration.

### 3.6 Consent-pending (the Gatekeeper blocked a non-READ action)

Trigger: `consent.ticket !== null` (store polls every 1.2 s).

**Visual reading (shell-level):**

- `<ConsentModal />` mounts at top of tree (always-mounted at
  `Shell.svelte:246`; visible iff `consent.ticket`). Scrim 32 %
  ink + 6 px blur; for **destructive** actions scrim shifts to
  `var(--surface-ink)` at 50 % (rare ink surface).
- The card has a **1.5 px synapse armor rect** that draws in
  left→right over 1.4 s (the `armor` gesture — this is the
  Thread used at higher weight to mean "this action is protected").
- Esc → deny (per APPFLOW §6.7).
- `DynamicIsland phase="consent"` (`CONSENT REQUIRED`, warn border).
  The thread keeps drawing.

### 3.7 Streaming-thinking

Trigger: `conversation.isStreaming === true`.

**Visual reading (shell-level):**

- `DynamicIsland phase="thinking"` widens to fit
  `truncate(currentTitle, 18) + ' · Thinking'`. Color `var(--synapse)`.
  The thread node amplifies its halo (3 → 6 px) and the dot
  itself pulses with the existing `breathe` keyframe.
- StatusBar stopwatch ticks at 60 ms cadence (frames); tokens +
  cost grow in real time. The model badge gets a tiny inline
  `<Pulse phase="acting" />` glyph prefix.
- Right rail (Chat): if a thread is selected, its row gets a
  `<Pulse phase="thinking" />` 4-px dot on the right edge. Other
  rows are dim (`opacity: 0.6`) to focus the user's eye.
- Main: route is `<Chat />`; the stream-bar (`travel` 2.6 s ease
  infinite) pulses pollen dashes in the active bubble.

### 3.8 Static-error (any route)

Trigger: route-level error outside the four above.

**Visual reading (shell-level):**

- `DynamicIsland phase="error"`.
- The route renders its `ErrorState` (per MOAT §2.6 — three lines:
  "what failed" / "likely cause" / "next action" + a retry
  button).
- Right rail (if present) renders its last-known state (no extra
  error copy — the route owns the messaging).
- The Thread in the titlebar keeps drawing. This is intentional:
  the page is structurally alive; a single route is broken.

### 3.9 Account-signed-out (default for v0.1.0)

Trigger: `account.isSignedIn === false`. This is the **default** for
new users.

**Visual reading (shell-level):**

- Shell renders **completely normally**. Sign-in is optional
  (locked decision #30). The DynamicIsland never reads sign-in
  state.
- The right-rail footer (a "Signed in" cell in `AccountMenu`) shows
  `Not signed in` with a `Sign in` link. Not present in v0.1.0 —
  deferred to v0.2.0 per APPFLOW §6.9.
- Settings → Account section is reachable via the rail. It shows
  the same `Not signed in` copy.

---

## 4. MOTION CHOREOGRAPHY

The motion vocabulary uses the tokens in `condura.css`. Per MOAT §2.3
and §2.10, components never redeclare reduced-motion overrides; one
block in the global stylesheet does the whole work.

### 4.1 Mount sequence

| Beat | What | Duration | Easing |
|------|------|----------|--------|
| 0–60 ms | `.shell` opacity 0 → 1 | 320 ms | `--ease` |
| 60–260 ms | `.shell` translateY(8px) → 0 | 320 ms (parallel) | `--ease` |
| 200–540 ms | NavRail item N opacity 0 → 1 + translateX(-6px) → 0 | 220 ms each | `--ease`; **stagger 60 ms** from the top |
| 200–260 ms | Titlebar `.tb-controls` group opacity 0 → 1 | 240 ms | `--ease` |
| 280–800 ms | Main `.route-container.route-enter` (blur-in) | 520 ms | `--ease` |
| 320–840 ms | Right rail (when present) opacity 0 → 1 + translateX(12px) → 0 | 360 ms | `--ease` |
| 360–680 ms | StatusBar opacity 0 → 1 + translateY(4px) → 0 | 320 ms | `--ease` |
| 800 ms onward | Titlebar `IntersectionObserver` fires; thread bend loop starts (if not reduced-motion) | continuous | rAF |

`prefers-reduced-motion: reduce` collapses all of the above to
opacity 0 → 1 (no translate, no blur) over `--dur-fast` (140 ms);
the thread bend loop never starts.

### 4.2 Route transitions

When the user navigates (rail click, ⌘K palette, ⌘1-9 chord, hash
change):

1. The outgoing route's `.route-container` gets
   `.route-exit` class: `opacity 1 → 0` + `translateX(0 → -12px)` +
   `blur(0 → 4px)` over 200 ms, `--ease`.
2. The `{#key route}` block triggers the unmount + remount. The new
   route's `.route-container.route-enter` fires `blur-in` over 240 ms.
3. Right rail: a cross-fade — outgoing rail slides 12 px left and
   fades over 200 ms; incoming rail slides 12 px right, fades in
   over 240 ms. Total transition: 440 ms (slightly over the 320 ms
   `--dur-slow` token, deliberate — the eye registers the rail
   swap as a separate event).

`prefers-reduced-motion`: both reduce to opacity-only transitions
(200 ms in / 200 ms out). No translate, no blur.

### 4.3 Titlebar Thread signature — the one element

The Thread is the canonical "this is now finished" gesture (MOAT §3).
For the Shell:

- **Always running**: once `TitlebarThread` mounts (after the IO
  triggers), it runs at 60 fps via rAF when visible. When the
  document is hidden, `visibilitychange` zeros out the rAF. When
  scrolled off-screen, `IntersectionObserver` zeros out the rAF.
  Both reduce CPU to 0 (verified in `TitlebarThread.svelte:21–67`).
- **Paused under reduced-motion**: the `matchMedia('(prefers-reduced-motion: reduce)')`
  early-returns on mount; the thread renders as a static horizontal
  line.
- **Paused under low-energy**: when `:root[data-energy="low"]`,
  the rAF cadence drops to 4 fps (one rAF tick every
  `requestIdleCallback` window). The grain layer also hides.
- **Reactive to agent state**: when `conversation.isStreaming`, the
  thread node (`<circle>`) switches its `r` attribute from 3 px to 6 px
  and back when idle. This is the only place an animation state changes
  in response to agent state.
- **Thread-drawn completion gestures**: any action inside the Shell
  that completes a **shell-level** state change (route entered,
  conversation started, conversation ended, error resolved, kill
  switch cleared) draws a **horizontal TitlebarThread reinforcement**
  on the next paint: a 1 px synapse stroke `pathLength: 1;
  stroke-dasharray: 1; stroke-dashoffset: 1 → 0` over `--dur-slow`
  (520 ms). This is one CSS line repeated, identical to the
  `err-hair-draw` recipe. (App-route-level completions use their own
  routes' Threads — the Shell's TitlebarThread reinforcement is for
  shell-level events only: nav, stream-start, stream-end, halt,
  consent-granted.)

### 4.4 Tactile press (per MOAT §2.2)

Two independent motions on every press:

1. `transform: scale(0.97)` — already in `condura.css:321–325`.
2. `filter: brightness(0.95) saturate(1.1)` — **added** to the
   global `.tactile` rule, per MOAT §2.2. This is the micro-change
   the spec calls out as missing.
3. `translateY(0.5px)` — visibly settle. Same addition.

All three apply to every `button:not(.no-tactile)`,
`[role='button']:not(.no-tactile)`, and `.tactile` element. One
vocabulary, no per-component overrides. (Per MOAT §2.7 we already
own this globally; the Shell enforces it by never importing any
component-local `transform: scale(...)` declarations.)

### 4.5 Focus halo (per MOAT §2.1)

The default `:focus-visible` is the polygon halo + 1 px synapse
inset (`var(--shadow-focus)`). For elements with rounded shapes:

| Border radius | Focus treatment |
|---------------|-----------------|
| < 8 px (chips, kbd keys) | default `--shadow-focus` (polygon halo) |
| ≥ 8 px (buttons, cards, inputs) | `box-shadow: 0 0 0 2px var(--synapse), 0 0 0 5px var(--pollen-halo-color)` — a synapse outline that tracks the radius, with a wider pollen halo. |
| `var(--r-pill)` (toggles, the theme button) | drop the inset 1 px line; `box-shadow: 0 0 0 2px var(--synapse)`. The synapse ring alone carries the state. |

The focus halo never uses `outline`. The Shell enforces this in
`Shell.svelte:212–214` by setting `outline: none` on every focusable
in the titlebar (the global `condura.css:299–302` does this for the
rest).

### 4.6 Reduced motion (the one place)

`condura.css:469–476`:

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
  .wordrise > span { transform: none; }
  .paper-grain, .mote, .ambient-thread { display: none; }
}
```

The Shell's contract: every animation declared inside `Shell.svelte`
routes through either `--dur-fast`, `--dur`, `--dur-slow`, or
`--dur-cine`. The global override zeroes all four. There are no
per-component reduced-motion blocks; this is the only block in the
codebase.

### 4.7 Energy budget (per MOAT §3 + §5.4)

`:root[data-energy='low']` (set when battery is low and no charger is
present):

- `--dur-cine` and `--dur-slow` drop to `0ms` (no slow entrances;
  no blur-ins).
- `.paper-grain`, `.mote`, `.ambient-thread` hide.
- The TitlebarThread `rAF` cadence drops to 4 fps via the existing
  visible-loop branch.
- The StatusBar stopwatch ticks at 1 Hz instead of 60 Hz.

The Shell queries `navigator.getBattery()` once at mount and on
`batterylevelchange` if available; otherwise falls back to the
`Settings → Power → Energy budget` user-configurable value.

---

## 5. KEYBOARD CHORDS

Per MOAT §2.10 the keyboard story is a first-class surface of the
shell. The Shell owns global chords; routes own route-local chords
(but never `Esc` to dismiss — that's Shell).

### 5.1 Global chords (registered at window level)

| Chord | Action | Source component |
|-------|--------|------------------|
| `⌘K` / `Ctrl+K` | Open Command Palette | `CommandPalette` |
| `⌘⇧P` / `Ctrl+Shift+P` | Open Quick Prompt overlay | `QuickPromptOverlay` |
| `⌘,` | Open Settings (`#/settings`) | `navigate('settings')` |
| `⌘1` / `Ctrl+1` | Chat | `navigate('chat')` |
| `⌘2` / `Ctrl+2` | Hub | `navigate('hub')` |
| `⌘3` / `Ctrl+3` | Skills | `navigate('skills')` |
| `⌘4` / `Ctrl+4` | Sync | `navigate('sync')` |
| `⌘5` / `Ctrl+5` | Audit | `navigate('audit')` |
| `⌘6` / `Ctrl+6` | Replay | `navigate('replay')` |
| `⌘7` / `Ctrl+7` | Channels | `navigate('channels')` |
| `⌘8` / `Ctrl+8` | Delegation | `navigate('delegation')` |
| `⌘9` / `Ctrl+9` | Settings | `navigate('settings')` |
| `⌘0` / `Ctrl+0` | About | `navigate('about')` |
| `⌘[` / `Ctrl+[` | History back (within hash history) | popstate handler |
| `⌘]` / `Ctrl+]` | History forward | pushState handler |
| `⌘N` / `Ctrl+N` | New conversation | `conversation.clear()` + focus composer |
| `⇧T` / `Ctrl+T` | Toggle light/dark theme | `setTheme(...)` |
| `Esc` (no focused input) | Dismiss topmost overlay (consent, kill switch, palette, quick prompt) | overlay store |
| `?` (focused shell only) | Open the `Shortcuts` overlay (lists all chords) | new `<Shortcuts />` overlay (v0.1.0; per MOAT §2.10) |
| `Cmd+Shift+Escape` | **Hard kill switch** (Layer 1, per CLAUDE.md §2.1) — wired to `haltFlag.Halt(ctx, "hard_hotkey")` via `internal/conductor/killswitch.go`. Independent of this Shell's keyboard handler. |

The `g` + `[s | h | a | c | k | r | l | d | , | ?]` two-key Go-to
gesture (per MOAT §2.10) is also active, registered on `keydown`
with a 1.2 s timeout:

| Chord | Action |
|-------|--------|
| `g s` | Settings (`#/settings`) |
| `g h` | Hub (`#/hub`) |
| `g a` | About (`#/about`) |
| `g c` | Channels (`#/channels`) |
| `g k` | Skills (`#/skills`) — `k` for **s**k**k**ills |
| `g r` | Replay (`#/replay`) |
| `g l` | Sync — `l` for **s**y**n**c (or `g n`, two letters) |
| `g d` | Delegation |
| `g ,` | About — mnemonic for "?" |
| `g ?` | Shortcuts overlay |

### 5.2 Modal focus traps

Three global overlays mount at the Shell level:

| Overlay | Focus trap? | Tab order |
|---------|-------------|-----------|
| `CommandPalette` | yes — `dialog` with `aria-modal="true"`, focus trapped in the input + listbox; first focusable is the search input. | input → listbox rows (Esc closes). |
| `QuickPromptOverlay` | yes — focus trapped in the composer input + Send button + 2-3 secondary buttons (model select, attach, settings). | composer input → Send → secondary buttons. Esc closes. |
| `ConsentModal` | yes — focus trapped in the modal card's buttons (Allow / Deny). | Allow button has initial focus (it is the rarer action). Esc → Deny. |
| `KillSwitchOverlay` | yes — focus trapped on the **Mint resume ticket** button. Esc does nothing (the user must explicitly resume). | only Mint button. |

Each trap is implemented with the standard pattern: a
`keydown`-on-Tab handler that wraps focus from last → first when
Shift+Tab is pressed on the first element, and from first → last on
plain Tab on the last element. The Shell does NOT implement this for
routes — routes own their own focus traps; the Shell only owns the
four global overlays.

### 5.3 Tab order (normal, no modal open)

The DOM order is the visible order (CSS-grid order = DOM order here):

1. **Titlebar** (left → right):
   1. wordmark (no focus — it's a `<div>`, not a button)
   2. theme toggle
   3. ⌘K hint (no focus)
   4. DynamicIsland (no focus — it's a `<div>`)
2. **NavRail** (top → bottom):
   1. Chat → 2. Hub → 3. Skills → 4. Sync → 5. Audit → 6. Replay →
      7. Channels → 8. Delegation → 9. Settings → 10. About
3. **Main** route (route-local tab order)
4. **Right rail** (top → bottom):
   1. New conversation button (Chat rail) / section anchor
      (Settings rail)
   2. items
5. **StatusBar** (left → right):
   1. stopwatch (no focus)
   2. git branch popover trigger (only when in a git repo)
   3. model badge button (only when a model is configured)

The visible focus halo (per §4.5) is always present.

`Shift+Tab` wraps to the last element (StatusBar's model badge,
when present, else the theme toggle).

### 5.4 What the Shell does NOT handle

- **Per-route keybindings** (e.g., Chat's `⌘↵` to send, Audit's
  `f` to filter). The Shell owns the global chords; routes own the
  local ones.
- **Window-level shortcuts the OS owns** (Cmd+Q, Cmd+W, Cmd+Tab on
  macOS; Alt+Tab on Windows; Super on Linux). Wails / Tauri / the
  browser handles these.
- **Mouse / pointer interactions** (no keyboard chord for clicking
  a glyph; that's just `Enter` or `Space` on a focused button via
  the tab order).

---

## 6. COMPONENTS USED

Every Condura component the Shell composes. Each entry shows what
the component does in the Shell and where it lives. **No third-party
modal / tooltip / popover libraries.** Per MOAT §2.8, three Svelte
primitives (`.c-modal`, `.c-sheet`, `.c-popover`) in `condura.css`
plus their component wrappers (planned for Phase 3.5) are the only
overlays.

| Component | Where used | Role in Shell |
|-----------|------------|----------------|
| `Cursor.svelte` | mounted once at the top of the Shell's template (above Ritual too, gated) | The pixel quill cursor + pollen hover ring. **Per MOAT §1.4 this should be opt-in** (Settings → Developer toggle). The brief mentions the `Cursor`; in v0.1.0 we keep it on by default for cross-modal review, then default-off when shipped. |
| `Ritual.svelte` | mounted when `showOnboarding === true` | The pre-window. Owns the window until `completeOnboarding(routeHash)` is called. |
| `NavRail.svelte` | grid-area `navrail` | The 10-route list + brand + footer. |
| `DynamicIsland.svelte` | inside `.tb-controls` | The status pill. |
| `TitlebarThread.svelte` | absolutely positioned inside `.titlebar` | The signature. |
| `Glyph.svelte` | inside `theme-toggle`, inside each `nav-item` | The icon set. |
| `Pulse.svelte` | inside `DynamicIsland`, inside rail-foot | The pulse indicator. |
| `Chat.svelte` | Main region when route is `chat` | The composer + conversation. |
| `Hub.svelte` | Main when `hub` | Public Skills Hub. |
| `Skills.svelte` | Main when `skills` | Local skills index. |
| `Sync.svelte` | Main when `sync` | P2P pairing. |
| `Audit.svelte` | Main when `audit` | HMAC-chained log. |
| `Replay.svelte` | Main when `replay` | 24-h scrubber. |
| `Channels.svelte` | Main when `channels` | Channel rows. |
| `Delegation.svelte` | Main when `delegation` | Sub-agent constellation. |
| `Settings.svelte` | Main when `settings` | The flowing settings document. |
| `About.svelte` | Main when `about` | Colophon + invariants. |
| `Placeholder.svelte` | n/a in v0.1.0 (was used for unimplemented routes) | Resigned. All 10 routes have real components. |
| `Delegation.svelte` (duplicate import) | imported but used only for the `delegation` route | OK — single component, single import line. |
| `CommandPalette.svelte` | mounted globally; `open` prop bound to `paletteOpen` state | The ⌘K meta-UI. |
| `QuickPromptOverlay.svelte` | mounted globally; `open` bound to `quickOpen` | The hero overlay. |
| `ConsentModal.svelte` | mounted globally at `Shell.svelte:246` | Always-mounted; visible iff `consent.ticket`. |
| `KillSwitchOverlay.svelte` | mounted globally; visible iff `halt.state.halted` | The kill-card. |
| `PairingModal.svelte` (legacy) | **not mounted** in v0.1.0 Shell — `Sync.svelte` owns its own pending card | Legacy variant; the modern Sync surface is the embedded card. |

**Note on `Cursor.svelte`:** per MOAT §1.4 this is an opt-in
component on `Settings → Developer`. The Shell imports + mounts it
always in v0.1.0 to keep the dev surface cohesive, but the gate
controlled by `developer.show_pixel_cursor` (a localStorage flag) is
read on mount; if `false`, `Cursor` renders nothing.

---

## 7. DATA FETCHED

The Shell reads state from stores (which wrap the daemon IPC) and
fires a small number of mount-time calls.

### 7.1 Initial IPC calls on mount

| Call | Source | Purpose |
|------|--------|---------|
| `initStores()` (`Shell.svelte:65–69`) | `lib/stores/init.ts` | Boots every store with its initial fetch (audit page 1, replay frames, channels, providers, etc.). Failures swallowed per-store; the Shell still renders. |
| `halt.startPolling()` (`Shell.svelte:70–73`) | `lib/stores/halt.svelte.ts` | Subscribes the store to `halt.status` (every 1 s). |
| `overlay.start()` (`Shell.svelte:75–79`) | `lib/stores/overlay.svelte.ts` | Subscribes to `overlay.status` (event-driven). |
| `ipc.firstRunStatus()` (`Shell.svelte:81–101`) | `lib/ipc/client.ts` | `{ complete: boolean }` used to decide whether to mount `Ritual`. |
| `ipc.onboardingIsComplete()` (same call) | same | Boolean; same gating. |
| `localStorage.getItem('condura-ritual-seen')` (`Shell.svelte:88–92`) | local | User-affordance flag. |

The route components handle their own IPC. The Shell never reaches
into the route.

### 7.2 SSE subscriptions

The Shell does not own SSE; `conversation.svelte.ts` owns the agent
stream subscription, mounted when `<Chat />` is in the tree.
However, the Shell does wire up **window-level event listeners**:

| Event | Listener | Source | Purpose |
|-------|----------|--------|---------|
| `hashchange` | `onHash` (`Shell.svelte:104–106`) | window | Updates `currentHash` → `route`. |
| `keydown` | `onKey` (`Shell.svelte:110–126`) | window | Global chords (⌘K, Shift+O, Shift+P). |
| `pointermove` | inside `<TitlebarThread />` | window | Bends the thread. |
| `visibilitychange` | inside `<TitlebarThread />` | document | Pauses the rAF loop in background tabs. |
| `matchMedia('(prefers-reduced-motion: reduce)')` | inside `<TitlebarThread />` | window | Disables the bend loop. |

### 7.3 Store reads

The Shell reads these stores reactively (Svelte runes):

| Store | Source | What the Shell reads |
|-------|--------|----------------------|
| `onboarding` | `lib/stores/onboarding.svelte.ts` | not directly — only the IPC call reads `onboarding.complete`. |
| `consent` | `lib/stores/consent.svelte.ts` | `consent.ticket` (to render `phase='consent'` on the island). |
| `halt` | `lib/stores/halt.svelte.ts` | `halt.state.halted`, `halt.state.reason` (to mount KillSwitchOverlay + `phase='error'` on the island). |
| `overlay` | `lib/stores/overlay.svelte.ts` | `overlay.active` (to toggle `.shell.overlay`). |
| `conversation` | `lib/stores/conversation.svelte.ts` | `isStreaming`, `currentTitle` (for the island task label). |
| `daemon` | `lib/stores/daemon.svelte.ts` | `daemon.connected` (for the island phase fallback). |

### 7.4 Polling cadences

| Polling | Cadence | Owner |
|---------|---------|-------|
| halt status | 1 s | `halt.startPolling()` |
| overlay status | event-driven (no fixed cadence) | `overlay.start()` |
| conversation stream | event-driven (SSE) | `conversation.svelte.ts` |
| consent ticket | 1.2 s | `consent` store |
| permissions (Ritual only) | 2 s | `Ritual.svelte` (not Shell) |
| conversations list (Chat rail) | 60 s | route-local inside `<Chat />` |
| ledger / onboarding probes | on mount + on `resume` (no fixed cadence) | `init.ts` |

The Shell itself runs no polling; it subscribes to stores that do.

### 7.5 What the Shell does NOT fetch

- **User content (messages, skills, audit events, replay frames)**: the
  route owns it. The Shell can never assume what route is mounted, so
  it cannot pre-fetch.
- **Provider / model list**: route-local (Chat).
- **Skill / sync / channel data**: route-local.

---

## 8. DESIGN DECISIONS — MOAT compliance + anti-patterns

### 8.1 Premium tests passed

| Test | Where it shows up in the Shell |
|------|--------------------------------|
| **The Restraint Test** (MOAT §1) | The Shell is a frame, not a stage. No six-keyframe awakening. No "err-state" block. The only reused flourish is `.alive` (the pollen "you" dot in the wordmark and the active-row pollen dot on the rail) — each `.alive` instance is load-bearing, not decorative. |
| **The Detail Test · focus rings** (MOAT §2.1) | The Shell uses `--shadow-focus` everywhere except pill radii (the theme toggle + DynamicIsland), where it uses the synapse ring only. Pairs with §4.5 of this spec. |
| **The Detail Test · press states** (MOAT §2.2) | The Shell adds the `filter: brightness(0.95) saturate(1.1) + translateY(0.5px)` to the global `.tactile` rule per MOAT §2.2. The Shell's component imports never import a per-component scale override. |
| **The Detail Test · reduced motion** (MOAT §2.3) | The Shell never re-declares reduced-motion. The single block in `condura.css:469–476` is the only place the override lives. |
| **The Detail Test · empty states** (MOAT §2.4) | The right-rail "No conversations yet." + `+ New conversation` button follows the three-line (what / why / action) empty-state pattern. Settings → Power's "No providers configured" follows the same pattern. |
| **The Detail Test · loading states** (MOAT §2.5) | The Shell has no loading state of its own — it just renders. Routes that fetch render skeleton states with their own Thread-drawn loaders (e.g., Audit renders a `READING THE CHAIN` Pulse + 360-px thread). |
| **The Detail Test · error states** (MOAT §2.6) | The Shell's `DynamicIsland phase="error"` is a single dot, single label (e.g., `OFFLINE`). The "what failed / likely cause / next action" pattern is enforced at the route level; the Shell just signals. |
| **The Detail Test · tactile vocabulary** (MOAT §2.7) | All buttons in the Shell use `class="tactile"` (or inherit via `button:not(.no-tactile)`). No per-component transition re-declarations. |
| **The Detail Test · overlay taxonomy** (MOAT §2.8) | The four global overlays (CommandPalette, QuickPromptOverlay, ConsentModal, KillSwitchOverlay) collapse to the three planned primitives (`.c-modal`, `.c-sheet`, `.c-popover`). PairingModal is resigned. |
| **The Detail Test · tooltip vs popover** (MOAT §2.9) | The Shell uses `title=` only for the kbd hint and the "you" dot, where the placement isn't read-critical. The theme toggle gets an opt-in `<Tooltip label>` wrapper (Phase 3.5). |
| **The Detail Test · keyboard story** (MOAT §2.10) | §5 of this spec. |
| **The Signature — the Thread** (MOAT §3) | The Shell is the **home** of the Thread. The `<TitlebarThread />` always renders when the Shell is mounted. Route-completion Threads draw from inside the route. Shell-level completions (route-entry, stream-start, stream-end, halt-cleared) draw the TitlebarThread reinforcement. |
| **The $50M Feel — cursor catches hoverable surfaces** (MOAT §5.1) | The Shell mounts the `use:hover-region` action via `<Cursor />`; `body[data-hover='1']` flips the cursor to the pollen-ring data-URI. |
| **The $50M Feel — composer focus → inward thread** (MOAT §5.2) | The Shell is not the composer (Chat owns it), but every text input the Shell exposes (e.g., ⌘K input on the palette via the CommandPalette; the QuickPrompt composer via the QuickPromptOverlay) reuses the same `::before` thread-draw under the input. |
| **The $50M Feel — Stop button telegraphs yielding** (MOAT §5.3) | Not directly — Chat owns Stop. But the Shell's DynamicIsland responds: on `stream.cancel()`, the island transitions from `thinking` (warn, breathing) → `acting` for 280 ms → `idle` (no badge), and the thread node's halo collapses. The shell-level stop signal. |
| **The $50M Feel — Settings nav has its own active state animation** (MOAT §5.4) | The right rail's active-row indicator (`:active::before`, the synapse segment) draws in via `scaleY(0 → 1)` on route change with no remount — same gesture, same `--dur-slow` ease. This is implemented in `NavRail.svelte:146–158` already. |
| **The $50M Feel — mobile ritual** (MOAT §5.5) | Not directly the Shell's concern (Ritual owns the breakpoint). The Shell defers to the same breakpoint contract (768 px collapse). |

### 8.2 Anti-patterns avoided (MOAT §4 — "What We Will Not Do")

| # | Anti-pattern | How the Shell avoids it |
|---|--------------|-------------------------|
| 1 | No gradient text | Titlebar uses `var(--content)` for the wordmark; no `background-clip: text`. |
| 2 | No emoji as UI icons | Every icon is a `<Glyph>` from `icons.ts`. The pollen dot in the wordmark is a 5-px `border-radius: 50%`, not an emoji. |
| 3 | No glassmorphism unless earned | Titlebar's glass is just `var(--surface)`; right rail is `var(--surface-card)`; no `backdrop-filter` anywhere. The 8 px blur on the CommandPalette + KillSwitchOverlay is the *only* allowed exception — and that's elevation-token-driven, not card-driven. |
| 4 | No rainbow accents | Status colors are `--ok` / `--warn` / `--danger` / `--info`. Brand is `--synapse` + `--pollen`. There is no purple/cyan/teal in the Shell. |
| 5 | No "Welcome to the future" copy | The titlebar shows literal ambient truth: `IDLE · LISTENING`. Empty states follow the MOAT §2.4 three-line pattern (see §3 of this spec). |
| 6 | No fake enthusiasm | No "Awesome!" / "Great choice!" / "You're all set!" toasts. The Thread is the only flourish for a completed moment — and it's earned, not performed. |
| 7 | No spinner loaders | No `<Spinner />` import anywhere in the Shell. Routes render their own Thread-drawn loaders (`INDEXING…` Pulse + thread). |
| 8 | No rectangular focus outlines | Default focus is `--shadow-focus` (rounded halo). Pill-radius elements use the synapse ring only. Topic covered in §4.5. |
| 9 | No double shadows | Titlebar has no shadow (it's a flat strip with a hairline). NavRail has no shadow. Right rail has `--shadow-card` only — no layered `box-shadow` overrides. StatusBar is hairline-bordered, no shadow. |
| 10 | No animation that doesn't carry meaning | Every animation in the Shell carries meaning (MOAT §4.10): titlebar thread = "alive / pointer-tracked"; focus halo = "selected"; pollen "you" dot = "you are here"; route-enter blur-in = "new surface arrived"; nav-item drawv = "you are here"; kill switch scrim = "halted". The Thread is the only flourish for completion. No decorative loops. |

### 8.3 What the Shell is uniquely responsible for

- **Being the home of the Thread.** Per MOAT §3, "A premium product
  earns one element. This is ours." Every "this is now finished"
  moment in the Shell draws the Thread. This is the one element the
  Shell owns that nothing else does. Every other surface in Condura
  *uses* the Thread, but the Shell is where the Thread *lives*.
- **Routing without routing.** The Shell swaps routes but never
  owns route state. The active route's component is mounted
  wholesale; the Shell never reaches inside. The hash-based router
  (`NavRail.svelte:30–41`) is the entire contract.
- **Hiding the OS from the user.** The Shell is the only surface
  aware of the Wails / Tauri / pure-Vite substrate. Routes don't
  know which they run in. The Shell does the
  `Cursor / Wails Drag region / safe-area` reconciliation.
- **Surviving every state.** §3 of this spec lists each reachable
  state; the Shell has a rendering decision for each. None of them
  fail silently.

### 8.4 What the Shell explicitly does NOT do

- **Open OAuth flows.** That's `account.signInWithEmail` /
  `account.oauthUrl` — owned by `Settings.svelte` (Account section).
- **Manage providers.** That's `Settings.svelte` (Power section).
- **Edit autonomy.** That's `Settings.svelte` (Autonomy matrix).
- **Emit audit events.** Only `Safety` modules emit. The Shell
  reads them, doesn't write them.
- **Cancel a stream.** That's `conversation.cancel()`. The Shell
  shows the result via the island phase; it never invokes the
  cancel itself.
- **Approve / deny a consent.** That's the Gatekeeper. The Shell
  renders `ConsentModal.svelte`; the modal calls `gatekeeper.*`.

### 8.5 Success criterion

The Shell earns the MOAT bar when, viewing any frame of any state
listed in §3, a new user can answer three questions without being
told:

1. **Where am I?** (titlebar wordmark + active nav row's "you"
   indicator).
2. **What is the agent doing right now?** (DynamicIsland phase +
   pulse color).
3. **What can I press to do the next useful thing?** (⌘K hint +
   affordances one tab away).

That's the shell's job. The routes do everything else.

---

## 9. Versioning + change-control

- **Spec version:** 0.1.0 — pinned to the v0.1.0 Shell component
  (`Shell.svelte:1`) and to the v0.1.0 design tokens
  (`condura.css:1`).
- **Update policy:** this document is **append-only in spirit**,
  per CLAUDE.md §30.5. Corrections are not edits — add an entry
  below.
- **Implementation divergences:** when the implementation disagrees
  with the spec, fix the implementation. When the implementation
  has a reason the spec doesn't cover, add a row to §8.6 explaining
  the cause.
- **Phase 4 hand-off:** §1–§8 are implementable as-is. If a Phase-4
  implementer needs to deviate (e.g., the right rail needs an
  additional `320 px → 360 px` transition for some screen), update
  this document in the same commit.

### 9.1 Open questions for the next session

- **StatusBar visibility** — is it always on ≥1024 px or only when
  there's a conversation in flight? Current spec assumes always on.
- **Voice indicator** — the DynamicIsland reads `phase='listening'`
  but there's no `<Glyph>` for "voice active" in `icons.ts` (the
  `<Pulse>` does the work). Confirm or replace.
- **`<Cursor />` default** — MOAT §1.4 says opt-in, Settings →
  Developer. The Shell mounts it always but reads the flag. Confirm
  the on-by-default behavior survives (or ship it off; the spec
  picks).
- **Right-rail "Conversations list" polling cadence** — current
  spec: 60 s. Could be on-mount + on-route-entry + on-write only.
- **`Shortcuts` overlay** — not yet a component in v0.1.0; current
  Shell has no `?` chord handler. Add as a Phase-5 deferred item
  per MOAT §2.10, or ship now.

---

**This document is the architecture. The code is the implementation. They
agree. When they diverge, the divergence is the spec-bug — fix the doc,
then fix the code, in one commit.** (APPFLOW.md closing note, applies here too.)
