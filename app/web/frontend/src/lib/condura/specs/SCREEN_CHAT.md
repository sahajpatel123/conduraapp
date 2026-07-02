# SCREEN_CHAT.md — Condura Chat · Screen Architecture

> **The contract.** This spec defines the layout, content slots, state
> matrix, motion choreography, keyboard story, component composition,
> data fetch contract, and design decisions for the **Condura Chat**
> surface — the hero route of the desktop shell. Phase-4/Phase-5
> implementations of `Chat.svelte` must conform to this document;
> if a component disagrees, the component is wrong, not the spec.
>
> **Audience.** Phase implementer. Designer for review.
>
> **Sibling specs:** `SCREEN_SHELL.md` (the frame that hosts this route),
> `SCREEN_NAVRAIL.md`, `SCREEN_SKILLS.md`, `SCREEN_ACCOUNT.md`,
> `SCREEN_QUICKPROMPTOVERLAY.md`, `SCREEN_REPLAY.md`.
>
> **Source-of-truth files:**
> - `app/web/frontend/src/lib/condura/Chat.svelte` — current implementation
>   (pre-spec; some drift listed in §8).
> - `app/web/frontend/src/lib/condura/condura.css` — design tokens.
> - `app/web/frontend/src/lib/condura/DIRECTION.md` — voice + personality.
> - `app/web/frontend/src/lib/condura/APPFLOW.md` — full user flow.
> - `app/web/frontend/src/lib/condura/MOAT.md` — premium-quality bar.

---

## 0. The contract in one sentence

**Condura Chat is where the user spends ~90% of their time.** It is a
quiet, attentive room — alive, but never loud. Its dominant feeling
should be "I can think here." Stream only when the user is streaming.
Listen when the user is quiet. The Thread is the only flourish; the
composer is the only anchor.

> **Voice test (DIRECTION.md §1):** *does this feel like a paper
> notebook that learned to listen — warm, awake, and never louder than
> the room it's in?*

---

## 1. LAYOUT & CONTENT

### 1.1 Region map (annotated)

Within the `Main` region of the Shell (`grid-area: main`, see
`SCREEN_SHELL.md §2.3`), Chat renders a **three-zone vertical stack**:

```
┌──────────────────────────────────────────────────────────────┐
│ (1) ConversationHeader · 56 px · sticky top (above scroll)    │
│     title • model badge • status chip • kebab                 │
├──────┬───────────────────────────────────────────┬──────────┤
│      │                                           │          │
│ (2)  │   (3) MessageFeed · 1fr                    │  (4)     │
│ Cl   │        virtualized list (760 px col,     │  Right   │
│  ~   │        binary-searched window)           │  Rail    │
│ 280  │        — or empty-state hero              │  (opt.)  │
│  px  │        — or streaming caret + bar         │          │
│      │                                           │          │
│      │                                           │          │
├──────┴───────────────────────────────────────────┴──────────┤
│ (5) Composer · ~120 px · hairline above                       │
│     textarea + voice orb + model select + send/stop          │
│     └─ hint line (mono, 10 px)                                │
└──────────────────────────────────────────────────────────────┘
```

On `<1440 px` the `(2) ConversationList` collapses to a small **rail
handle button** in the header (drawer-style); the `(4) RightRail`
collapses entirely — its content re-homes to a `.c-sheet` (per
`SCREEN_SHELL.md §1.4`). On `<768 px` the header compresses to a
single row and the composer grows to fill the bottom inset.

### 1.2 Region-by-region contract

#### (1) ConversationHeader — 56 px tall, sticky

| # | Slot | Content |
|---|------|---------|
| 1a | Eyebrow | `— A conversation` (mono 10 px / +0.14em / uppercase / `var(--content-faint)`). Empty-conversation eyebrow reads `— A fresh page`. |
| 1b | Title | Auto-set on first message (truncated user's first sentence, 60 chars max, Instrument Serif 20 px / −0.025em). Editable on click → inline `<input>` with the same serif typography (no chrome change). Pencil glyph on hover → `Tooltip label="Rename"` (`MOAT §2.9`). |
| 1c | Model badge | `claude-sonnet-4.5 · anthropic` mono 11 px in `--paper-sunken` pill. Click → `QuickPromptOverlay`-style model picker sheet (or, if no overlay, inline `<select>` collapse). |
| 1d | Status chip | `<Pulse phase={...} size={6}>` + 2-3 word label (`Idle`, `Thinking…`, `Listening…`, `Offline`). Border ties to `phase` color (synapse / pollen / warn / danger). |
| 1e | Kebab | `<Glyph name="more" size={16}>` opens `<Menu>` with: *Rename · Pin · Delete · Export transcript (.md)*. Destructive items render in `--danger`; Esc/outside-click closes. |

**States:**

- **Idle:** `Pulse phase="idle"` + `Idle`. Title in `--content`.
- **Streaming:** eyebrow flips to `— Thinking`, status chip → `Pulse
  phase="acting" size={8}` + `Thinking…` (in `--warn` border). The
  title is **locked** during streaming (no inline edit affordance).
- **Consent pending:** status chip → `Pulse phase="awaiting"` + `Waiting on you` (in `--warn`). The kebab adds a top row *"Resolve consent"* linking to the audit/replay surface.
- **Listening (voice session):** status chip → `Pulse phase="listening"` + `Listening…` (pollen border). The VoiceOrb in the composer is the visual emphasis (see 1.5).

A 1 px `var(--hair)` hairline sits beneath the header; on `route-enter`
the hairline is the **first** element to draw in (`stroke-dashoffset 1→0`,
`--dur-fast`).

#### (2) ConversationList — 280 px wide rail (≥1440 px)

A scrollable column of conversation rows, sorted by `last_touch DESC`
top → bottom. Each row is a `<button class="cl-row">`:

- **At-rest:** `14 px` serif title (1-line clamp via `text-overflow:
  ellipsis`), mono `10 px / +0.12em / uppercase` `LAST · 14:32`
  timestamp, `glyph:pin` if pinned (right edge, `--pollen`).
- **Hover:** `var(--surface-card)` fill; title color brightens to
  `--content`. `<Pulse phase="idle" size={4}>` appears in the row's
  right margin when the conversation has unread activity.
- **Active (current):** left edge gets a 2 px `--synapse` segment that
  draws in via `scaleY(0 → 1)` over `--dur-slow` (the Thread
  signature — same gesture as `NavRail.svelte:146–158`).
- **Focus-visible:** `--shadow-focus` (the synapse + pollen halo per
  MOAT §2.1). The halo tracks the 8 px radius of the row.
- **Right-click / kebab on row:** menu with *Rename · Pin/Unpin ·
  Delete · Open in new window*. Delete is `--danger`.

A sticky **"+ New conversation →"** row sits at the top of the rail
(under the brand line, separated by a hairline). Click → fires
`conversation.clear()` + focuses the composer (`MOAT §5.3` flow).

Below the rows, a **footer whisper** in mono 10 px faint: `N conversations` (the count from the conversation store, polled at 60 s cadence when visible).

#### (3) MessageFeed — the heart

A vertically-scrolled column with `max-width: 760 px` (the brand
reading-line) centered in the available width. Three sub-states:

**3a. Empty (no messages, no stream).** The empty garden. Per
`MOAT §2.4` (empty states teach), three lines:

1. **What this area is** — eyebrow: `— A quiet place to write`.
2. **What the room is for** — Instrument Serif display headline
   "Your / computer, / *alive.*" (`clamp(40px, 6vw, 68px)`,
   40/68 clamp — the `.alive` class is the *one* allowed
   use-per-surface, per `MOAT §1.7`).
3. **The one action that fills it** — Inter 17 px / 52ch max lead:
   *"A quiet, attentive presence on your machine. It perceives only
   what it must, and it acts only after it shows you what it's about
   to do. Press your hotkey, or ask below."* — followed by **3–5
   suggested-prompt chips**.

**The 3–5 chips.** Each is `<button class="chip">` (pill, paper-sunken
fill, mono 11 px label). Slots:

1. `↳ Find a file from last week.`
2. `↳ Draft a reply to Maya.`
3. `↳ Summarize the open PRs in /code/synaptic.`
4. `↳ Watch my screen and tell me what I'm doing.`
5. `↳ Walk me through this PDF.` *(visible only when `permission:documents-granted === true`)*

**Chip hover:** `transform: translateY(-1px)` + border-color
`--hair-strong`; press: scale(0.97) + brightness(0.95). Click fills
the composer + sends (1-tap action).

Behind the empty headline, **14 pollen motes** drift in a 9 s linear
loop (`pollen-float` keyframe, randomized left 0–100 %, bottom 0–40 %,
dx ±60 px, dy −120 to −300 px). The garden is the only decorative
ambient on this surface; it is **hidden** under `prefers-reduced-motion`
and under `data-energy="low"` (per `condura.css` global rule — no
component-local override, per `MOAT §2.3`).

**3b. Mid-conversation (turns + threads).** Each `Turn` is a
`<div class="msg {role}">` block:

- **Label.** Mono 10 px / +0.14em / uppercase. `You` (pollen) /
  `Condura` (synapse) / `System` (faint). The label sits **above** the
  bubble, not beside it — paper-notebook convention.
- **Bubble.** `max-width: 88%`, `padding: 16px 20px`, `border-radius:
  14px`. User bubble: pollen tint (5 % wash). Condura bubble: 1 px
  synapse hairline. System bubble: hair-only, no fill.
- **Turn-thread.** Between every pair of turns, a 2 px × 28 px
  vertical `<Thread orientation="v">` (the `<Thread>` component from
  `Thread.svelte`). It draws in left → bottom → right (the H+V leg
  pattern, the same gestural vocabulary as the constellation) over
  `--dur-slow` when a new turn lands — this is the **live Thread**
  for the surface. Hover on a completed turn draws a faint pollen halo
  around the bubble (the universal "you're here now" cue).
- **Streaming caret + bar.** The active Condura bubble carries a
  blinking 8 × 1.1 em caret in `--synapse` at the tail of
  `content`. Below it sits the `.stream-bar` (a 1 px hairline with a
  40 × 3 px pollen dash that travels left → right on a 2.6 s ease
  infinite loop). The dash halts on `conversation.cancel()`.

**Turn virtualization.** DOM holds only the visible window plus a
600 px `BUFFER_PX` (binary search over `itemOffsets`). Cost is constant
in message count — a 2,000-message thread renders the same DOM nodes
as a 20-message one. See `Chat.svelte:60–98` for the existing logic.

**Tool-call chips.** When the streaming agent invokes a tool, a
`<span class="tool-chip">→ tool_name</span>` chip renders below the
active bubble (5–7px mono, synapse border, 22 px tall pill — `<Glyph
name="bolt">` prefix). Up to 6 chips visible; overflow collapses
behind a `+3 more` chip on hover. Tool chips are not clickable in
this view — they belong to the audit thread (right rail or
`#/audit`).

**3c. Error mid-stream.** A `<div class="err-state">` renders between
the last user message and the partial agent bubble. Per `MOAT §2.6`,
three lines:

1. **What failed** — italic Instrument Serif headline 22 px:
   *"Connection to daemon"*.
2. **Likely cause** — italic 15 px:
   *"The thread stopped mid-sentence. The daemon may have been
   restarted, or the network to your provider was lost."*
3. **Next action** — a pill `<Button variant="secondary">` reading
   *"Try again →"* calling `conversation.retry()` + a
   `<Button variant="ghost">` reading *"Open Settings → Connection"*.

Below: the standard `.err-hair` drawing left → right over 600 ms
(`--ease`, 120 ms delay).

#### (4) RightRail (optional, contextual) — 320 px

Currently exposed for routes that have a sheet, but Chat keeps the
rail closed in v0.1.0. The variant for `≥1440 px` future use is the
**Context panel**: per-message detail (selected turn's actor, blast
radius, audit receipt — read-only mirror of the audit entry). When
the rail is open, a `close` (×) glyph appears at the top-right of the
rail. On `<1440 px` the content re-homes to a `.c-sheet` slide-in
from the right (`SCREEN_SHELL.md §1.4`).

For v0.1.0 the rail is conditionally rendered only when a message is
right-clicked (the message-side details sheet). The grid track stays
zero when not needed (no layout jitter).

#### (5) Composer — pinned to bottom

The only anchor. Always visible, always reachable. 1 px `--hair` rules
above; rounded `--r-lg` card with `var(--shadow-card)` lift.

**Anatomy:**

| # | Slot | Content |
|---|------|---------|
| 5a | Textarea | `<textarea>` 16 px Inter, `min-height: 48 px`, auto-grow to `240 px`. Placeholder: `→ Ask Condura to do something…`. Single `Ghost button`-style flush-right `↗` send-arrow inside (only visible when `inputText.trim().length > 0`). |
| 5b | VoiceOrb | 28 × 28 round, `--surface-card` fill, `--hair` border. Holds the `<Glyph name="mic">`. Hover: cursor changes to the pollen ring (via `use:hover-region`, `MOAT §5.1`). Click → toggles `voice.active`; while listening, the orb cycles the `breathe` keyframe and shows a 3-dot waveform (the `<Glyph name="waveform">`). Disabled when `voice.micAvailable === false`; tooltip *"Mic not granted"* per `MOAT §2.9`. |
| 5c | Model select | `<select>` mono 11 px / +0.08em, paper-sunken pill. Only renders when `modelOptions.length > 0` (per existing `Chat.svelte:267–275`); when empty, collapses to a whisper hint ("No model — add one in Settings → Power"). The select's `:focus-visible` is the synapse ring + 5 px pollen halo (per `MOAT §2.1` — pill-radius). |
| 5d | Send / Stop | Primary button (pollen fill, ink text). Stop variant (synapse border) only when `conversation.isStreaming`. Per `MOAT §5.3`: on Stop click, the label fades `Send` → `Stopping…` over 280 ms, then to `Stopped` + a 1× pollen pulse that dissolves the streaming turn (the thread above it draws out **right → left** — the reverse gesture — at `--dur-slow`). |

**Hint line.** Below the card, mono 10 px / +0.12em / uppercase /
`var(--content-faint)`, centered: `⌘↵ to send · Esc to stop · your
hotkey to summon`. Drawn in `--content-mute` on hover (the hint
brightens — it is decorative, not the affordance).

**Composer focus state.** The signature focus gesture — already
implemented (`Chat.svelte:515–528`): a 1 px `--synapse` line `::before`
on the composer card scales from `scaleX(0)` → `scaleX(1)` from the
left over `--dur-slow`, `--ease`. Per `MOAT §5.2`, this gesture is
the **universal "I am writing here" cue** — it must read identically
across every text input in the product (already the pattern in
`CommandPalette`, the `email` field in Account, the hotkey capture,
the Hub search).

**Composer disabled state (no provider).** The whole card greys to
`opacity: 0.55`, cursor `not-allowed`, textarea carries
`aria-disabled="true"`. Below: a non-shaming whisper card (see §2.7).

---

## 2. STATE MATRIX

This is what Chat renders in every reachable state. Defaults for
unhandled errors: **honest degradation** (a route-level `ErrorState`
per `MOAT §2.6`) — never a dead wall.

### 2.1 Empty (no messages, no stream, provider available)

- Empty hero + 14-mote garden.
- Composer enabled.
- Status chip: `Idle`.
- Suggested-prompt chips rendered (the 5 from §1.3a — chips 2–4 hidden
  on `permission:documents-granted === false`).
- Eyebrow in header: `— A fresh page`.
- ConversationList rail: empty state whisper *"No conversations yet"*
  + `+ New conversation` button.

### 2.2 Empty, **no provider configured**

- Same empty hero + garden.
- Composer **disabled** (whole card greys to 0.55 opacity).
- A non-shaming whisper card sits above the composer (mono 11 px
  paper-sunken fill, hairline border):
  - Headline (Instrument Serif italic 17 px / −0.025em):
    *"No model wired yet."*
  - Sub (Inter 14 px):
    *"Condura works with Ollama on this machine, or with any API key
    you paste in Settings → Power."*
  - Two CTAs:
    - `Add an API key →` (pollen primary)
    - `Detect Ollama →` (paper-sunken secondary)
- StatusBar model badge reads `no model`.
- Send button is disabled (textarea has `aria-disabled="true"`).

### 2.3 Loading (conversation list hydrating)

- ConversationList rail renders a vertical stack of **7** skeleton
  rows (each: 1 row × 24 px with a 1.6 s `drawthread` keyframe left →
  right — the loading state per `MOAT §2.5`).
- A mono-uppercase label above the skeletons:
  `INDEXING CONVERSATIONS…`.
- A 6 px `<Pulse phase="thinking">` dot to the right of the label.
- **No spinner.** Per `MOAT §4 rule 7`.

### 2.4 Streaming

- Active turn bubble inserts at the bottom of `MessageFeed` (after the
  last user message).
- Caret + stream-bar pulse in the active bubble (the pollen dash
  travels on a 2.6 s loop).
- DynamicIsland phase → `thinking`; header status chip → `Pulse
  phase="acting"> + Thinking…` (warn border).
- StatusBar stopwatch + tokens + cost columns begin ticking (per
  `SCREEN_SHELL.md §2.5`).
- Tool-call chips render below the active bubble as tools fire.
- Turn-thread between the previous turn and the new bubble draws in
  over `--dur-slow` (the signature).
- Auto-scroll: if `scrollEl.scrollHeight - scrollTop - clientHeight <
  200 px`, the scroll **re-snaps** to the bottom via a `requestAnimationFrame`
  loop. Otherwise the user can scroll freely (a `data-pinned="false"`
  attribute appears on the container so the user knows they've
  detached).

### 2.5 Interrupted (user clicked Stop, or daemon timed out)

- The streaming bubble freezes on its last received character (caret
  stops blinking; stream-bar pollen dash halts).
- A **"Continue →"** pill renders below the frozen bubble (paper-sunken
  fill, mono 11 px label, single chevron). Click fills the composer
  with a "Continue: " prefix and focuses the textarea.
- Status chip → `Pulse phase="idle"> + Stopped`.
- The thread above the bubble draws out **right → left** in `--danger`
  (`scaleX(1 → 0)` from the right, `--dur`) to telegraph *"the stream
  ended."*

### 2.6 Error (mid-stream daemon / network / provider failure)

Per `MOAT §2.6`. The `ErrorState` is rendered (between the last user
message and the partial agent bubble) with:

1. **What failed** — italic Instrument Serif 22 px (one noun):
   "Connection to daemon" / "Provider rejected request" / "Network
   failure" — surfaced from `conversation.streamingError.kind`.
2. **Likely cause** — italic 15 px (one phrase): drawn from
   `kind` → cause map (`daemon` → "daemon was restarted"; `provider`
   → "the API key may have rotated"; `network` → "your network is
   unreachable").
3. **Next action** — pill button:
   - daemon → "Restart daemon →" (`ipc.daemonRestart()`) +
     "Open Settings → Connection" ghost.
   - provider → "Try a different model →" (opens the model select) +
     "Open Settings → API keys".
   - network → "Try again →" + "Open Network Settings".

Below: the standard `.err-hair` draws in 600 ms (`--ease`,
120 ms delay).

### 2.7 No-provider state (composer disabled)

See §2.2 — the whisper card with "No model wired yet." +
`Add an API key →` / `Detect Ollama →`. The send button is also
disabled. The status chip in the header reads `Idle · no model`.

### 2.8 Kill-switch-active (`halt.state.halted === true`)

- The whole Chat surface dims to `opacity: 0.4` (the page is still
  recognizable).
- `<KillSwitchOverlay>` mounts at `--z-modal` (per
  `SCREEN_SHELL.md §3.5`).
- **No Chat-local state change beyond the dim**. The overlay owns
  the messaging. The composer is read-only during the halt (esc
  cannot bypass).
- Resume: the user clicks "Mint resume ticket →" on the overlay,
  which calls `halt.resume()`. Chat un-dims when `halt.state.halted`
  flips back to false.

### 2.9 Consent-pending (the Gatekeeper blocked a non-READ action)

- `<ConsentModal>` mounts (per `SCREEN_SHELL.md §3.6`).
- DynamicIsland flips to `consent` phase.
- Chat composer's textarea is **disabled** (a non-READ action needs
  consent, but consent is modal-driven, not composer-driven). The
  existing draft in the textarea persists.
- When the user clicks Allow or Deny in the consent modal, the
  composer's disabled state lifts, and a pulse ↗ shows on the Send
  arrow to invite the user to retry.

### 2.10 Voice session active (mic hot)

- VoiceOrb in the composer cycles the `breathe` keyframe in
  `--pollen` (the breathe is the universal "we're listening" cue).
- A **3-dot waveform** replaces the mic glyph (3 stacked mono
  rectangles at stepped heights 8/16/24 px, color `--pollen`,
  animated amplitude on `waveform-tick` 1.4 s ease infinite).
- Transcript lines render **above the composer** (a hairline-
  bordered paper-sunken paper-card with mono 11 px text), max 6 lines,
  collapsed above to last 4.
- On submit (silence 1.5 s, or the user clicks Stop), the transcript
  populates `inputText` and the orb fades to the at-rest state.

### 2.11 Message hover (universal)

- A 4 px pollen halo around the bubble (`box-shadow: 0 0 0 4px var(--pollen-halo)`),
  drawn in over `--dur-fast`. On hover leave, dissolves.
- The "you" indicator (pollen dot, 5 px) appears at the right edge of
  the bubble.

---

## 3. MOTION CHOREOGRAPHY

The motion grammar lives entirely in `DIRECTION.md §5` — components
do not declare their own timings. The four durations (`--dur-fast
140ms` / `--dur 280ms` / `--dur-slow 520ms` / `--dur-cine 900ms`) and
the three eases (`--ease` / `--ease-in` / `--ease-pop`) are the only
vocabulary.

### 3.1 Enter (Chat mounts as the active route)

| Beat | What | Duration | Easing |
|------|------|----------|--------|
| 0–140 ms | Surface `opacity 0→1` | `--dur-fast` | `--ease` |
| 60–260 ms | (2) ConversationList rows stagger in: opacity 0→1 + translateX(-6px)→0 | 220 ms each, **stagger 40 ms** | `--ease` |
| 140–420 ms | (1) Header `.route-enter` opacity 0→1 + translateY(-8px)→0, then hairline draws in (`stroke-dashoffset 1→0`) | `--dur` | `--ease` |
| 200–720 ms | (3) MessageFeed `blur-in`: opacity 0→1 + `filter: blur(8px)→blur(0)` + translateY(12px)→0 | `--dur-slow` | `--ease` |
| 280–800 ms | (5) Composer card pops: opacity 0→1 + translateY(8px)→0 + shadow-card→shadow-float ramp | `--dur-slow` (parallel) | `--ease` |
| 800 ms+ | Garden motes begin drifting (continuous, 9 s linear loop) | — | — |

`prefers-reduced-motion: reduce` collapses the entire mount to opacity
0→1 over `--dur-fast` (no translate, no blur). Garden motes are hidden
via the global `condura.css` rule — Chat never re-declares this.

### 3.2 Streaming

- **Token cadence**: 16 ms per token (`@media (max-width:
  prefers-reduced-motion: reduce)` collapses to whole-message
  reveals; the message fade-in uses `--dur-fast`).
- **Stream-bar pollen dash** travels left → right on a 2.6 s
  `--ease` infinite loop. Halts instantly on `conversation.cancel()`.
- **Caret** blinks (`steps(2)` 1 s infinite).
- **Turn-thread draw**: `stroke-dashoffset 1 → 0` over `--dur-slow`
  as soon as a new turn is appended (between the previous completed
  turn and the active one).
- **Tool-chip append**: fade-in (`opacity 0 → 1`, 160 ms,
  `--ease`).

### 3.3 Route transitions (out → in when Chat is re-mounted)

- **Outgoing** `.route-container`: `opacity 1→0` +
  `translateX(0 → -12px)` + `filter: blur(0 → 4px)` over 200 ms
  (`--ease`).
- **`{#key route}`** triggers unmount + remount.
- **Incoming** `.route-container.route-enter`: `blur-in` over 240 ms
  (`--ease`).
- Total transition: ~440 ms.

`prefers-reduced-motion: reduce` collapses both to opacity-only.

### 3.4 Composer focus (the signature gesture)

The composer's `::before` 1 px `--synapse` line scales from
`scaleX(0)` → `scaleX(1)` over `--dur-slow` (`--ease`) when the
textarea gains focus-visible. On blur, reverses over `--dur`. The
hairline spans the inner padding (`left: var(--space-4); right:
var(--space-4);`) and sits at `bottom: 0` of the card, padding
above for the keyboard stroke. Per `MOAT §5.2`, this gesture must
read identically across every text input in the product.

### 3.5 Hover on message

- The pollen halo around the bubble fades in over `--dur-fast`
  (140 ms).
- The "you" pollen dot (5 px, `--pollen`) fades in at the right edge
  over `--dur`.
- Hover-leave reverses both.

### 3.6 Re-measure on window resize (virtualization correctness)

- A `ResizeObserver` on the scroll container rebuilds `viewportH` and
  re-runs the binary-search windowing. Existing `Chat.svelte:100–120`
  does this on `scroll`. `ResizeObserver` is added so window-resize
  doesn't desync visible offsets.

### 3.7 Stop / Yield (the `$50M` gesture)

Per `MOAT §5.3`:

1. Click Stop. Send button text fades `Send` → `Stopping…` over
   280 ms (label cross-fade).
2. Bubble's caret freezes; stream-bar pollen dash halts; the
   preceding turn-thread draws **out** right → left at `--dur`
   (the "yield" gesture).
3. The bubble itself dissolves in 200 ms (`opacity 1 → 0`). A
   1× pollen pulse beneath the Send button signals *"the agent
   yielded to you"*.
4. After 280 ms, the button reads `Stopped` (paper-sunken fill,
   faint text). It returns to `Send` pollen on the next user input.

### 3.8 Empty hero wordrise

The three-word headline uses a **clip-path wipe per word** (`inset(0
100% 0 0)` → `inset(0 0 0 0)` over 600 ms, `--ease`, **stagger
120 ms**). Under `prefers-reduced-motion: reduce`, the words render
as a single block, no animation.

### 3.9 Reduced-motion contract

`condura.css` owns the single media query (per `MOAT §2.3`). Chat
never re-declares `@media (prefers-reduced-motion: reduce)`. Any
animation it names uses the four duration tokens so the global
override zeroes them.

### 3.10 Energy budget

Under `:root[data-energy="low"]` (battery-aware mode, per
`DIRECTION.md §5`):

- Garden motes hidden (via the global rule).
- Stream-bar pollen dash cadence drops to 1 × per 6 s (1/4 the
  normal tempo). The dash still travels, but slowly.
- Token cadence collapses from 16 ms to 64 ms (the typewriter slows
  to a tick).
- Caret no longer blinks (steps are paused).

The product remains usable; it just stops performing.

---

## 4. KEYBOARD

Per `MOAT §2.10`, the keyboard story on Chat is a first-class surface.
Local chords (Chat owns them) layered on top of the global chords in
`SCREEN_SHELL.md §5.1`.

### 4.1 Chat-local chords

| Chord | Action | Notes |
|-------|--------|-------|
| `Enter` | Send (only when `inputText.trim() !== '' && !conversation.isStreaming`) | Plain Enter sends. With Shift+Enter inserts a newline (default `<textarea>` behavior). |
| `⌘↵` / `Ctrl↵` | Send (when IME `isComposing === true` it does **not** fire) | Already wired in `Chat.svelte:159`. |
| `Shift+Enter` | Newline (default behavior) | — |
| `Esc` (while streaming) | Cancel the active stream | Calls `conversation.cancel()`. The yield gesture fires. |
| `Esc` (no stream, textarea empty) | No-op (the global Esc dismisses any top overlay — see `SCREEN_SHELL.md §5.1`) | — |
| `Esc` (composer focused, with text) | Blur the composer | — |
| `↑` (composer empty, not focused) | Edit the **last** user message | Re-fills the textarea with the last user turn's content (mono 11 px "editing" eyebrow appears above the textarea). Enter sends. |
| `⌘.` | Stop (alias for the global kill switch L1 hotkey) | Halt intent — the same as clicking Stop while streaming; or, if no stream, sets `halt.requestHalt('user_request')`. |
| `⌘N` / `Ctrl+N` | New conversation | `conversation.clear()` + focus composer. Per `SCREEN_SHELL.md §5.1`. |
| `⌘K` / `Ctrl+K` | Open palette | Global (Shell). |
| `⌘⇧P` / `Ctrl+Shift+P` | Summon QuickPromptOverlay | Global. The current composer text drafts into the overlay if non-empty. |
| `⌘[` / `Ctrl+[` | History back | Global. |
| `⌘]` / `Ctrl+]` | History forward | Global. |
| `⌘R` / `Ctrl+R` | Retry last turn (when mid-stream error present) | Visible iff `conversation.streamingError !== null`. |
| `⌘D` / `Ctrl+D` | Delete the active conversation | With confirm modal (paper-sunken card, "Type the conversation title to confirm"). |

### 4.2 Modal focus traps (Chat-local)

- **Error retry sheet** (`:has(.err-state)`): focus is trapped inside
  the retry buttons. Esc → dismiss (revert to error block, leave
  state unchanged).
- **Continue-pill tooltip** (when hovering a stopped turn's
  "Continue →" pill): a `<Tooltip label="Re-fill your composer with 'Continue: ' and send">` per `MOAT §2.9`.

### 4.3 Tab order (no overlay open)

1. ConversationList rail: `+ New conversation` → conversation rows
   top → bottom → `Pin`/`Delete` kebab of the last row.
2. Conversation header: kebab → model select → rename pencil.
3. MessageFeed: any focused bubble (e.g., user message with selected
   text or right-click affordance).
4. Composer: VoiceOrb → model select → send (or stop) button.

### 4.4 What Chat does NOT do via keyboard

- **Voice activation** — that's the wake word (`hey condura`)
  via the daemon. Chat only consumes `voice.transcript` from the
  store; it never starts a mic session via keypress.
- **Conversation-list reordering** — defer to drag in v0.2.0.
- **Citation hops** — Chat is the surface for streaming; citations
  live on the message bubble (future).

---

## 5. COMPONENTS USED

Every Condura component the Chat surface composes, with where each
lives and its role here. **No third-party modal/tooltip/spinner libs.**

| Component | Path | Role in Chat |
|-----------|------|--------------|
| `ConversationList.svelte` | `app/web/frontend/src/lib/condura/ConversationList.svelte` | The (2) 280 px conversation rail. Rows sorted by last touch; pin / delete affordances; `+ New` row at top. |
| `Message.svelte` *(new)* | extracted from `Chat.svelte` | The `<div class="msg {role}">` block — turn body. Props: `role`, `content`, `isStreaming`, `streamingToolCalls`, `onRetry?`. Variants: **user** (right-aligned, pollen tint), **assistant** (left-aligned, synapse hairline), **system** (centered, hair-only). |
| `MessageFeed.svelte` *(new)* | extracted from `Chat.svelte` | The (3) virtualized list. Owns binary search + windowing + spacer divs. **Inputs:** a list of `Turn`s and the conversation-scroll state. **Outputs:** nothing (read-only). |
| `Composer.svelte` *(new)* | extracted from `Chat.svelte` | The (5) composer card. Owns the textarea, focus-thread `::before`, model select, send/stop, and `magnetic` Send affordance (`MOAT §5.3`). |
| `VoiceOrb.svelte` *(new)* | new component | The 28 × 28 round voice button. States: `at-rest`, `listening` (breathe pollen + waveform glyph), `disabled` (mic not granted). |
| `ConversationHeader.svelte` *(new)* | extracted from `Chat.svelte` (header) | The (1) sticky top. Owns the title input, model badge, status chip, kebab menu. |
| `ErrorState.svelte` | `app/web/frontend/src/lib/condura/ErrorState.svelte` | The §1.3c / §2.6 error block. Three lines + err-hair. Already collapsed per `MOAT §1.2`. |
| `EmptyState.svelte` *(new)* | extracted from `Chat.svelte:185–212` | The §1.3a empty garden + headline + chips. Composes `Garden.svelte` (the 14 pollen motes). |
| `Garden.svelte` *(new)* | new | The deterministic pollen motes layer. <span class="mote"> × 14, randomized drift; respects global `prefers-reduced-motion` and energy-mode rules — no local override. |
| `Thread.svelte` | `app/web/frontend/src/lib/condura/Thread.svelte` | The 2 px × 28 px vertical `Thread orientation="v"` between turns (and the top-edge hairline for the header). |
| `Pulse.svelte` | `app/web/frontend/src/lib/condura/Pulse.svelte` | The status-chip pulse + the inline loaders' pulse (Per `MOAT §4 rule 7`, the **only** allowed loading indicator). |
| `Glyph.svelte` | `app/web/frontend/src/lib/condura/Glyph.svelte` | Icons: `send`, `mic`, `waveform`, `stop`, `more`, `pin`, `delete`, `pencil`, `close`. |
| `Button.svelte` | `app/web/frontend/src/lib/condura/Button.svelte` | Primary (pollen), secondary (paper-sunken), ghost (text), danger (rust). All tactile. |
| `Tooltip.svelte` *(new)* | per `MOAT §2.9` | Hover-delay 400 ms / exit 75 ms / `aria-describedby` / one-line max. Used on VoiceOrb, kebab, model badge, pencil. |
| `Menu.svelte` *(new)* | new primitive | Right-click / kebab surface. Anchored, focus-trapped, dismiss-on-outside-click + Esc. |
| `Sheet.svelte` *(new)* | `.c-sheet` wrapper (per `MOAT §2.8` & `SCREEN_SHELL.md §1.4`) | When the right rail re-homes to a slide-from-edge sheet on `<1440 px`. |
| `DynamicIsland.svelte` | `app/web/frontend/src/lib/condura/DynamicIsland.svelte` | Titlebar pill (not owned by Chat, but Chat supplies the `phase` value via `agentPhase` derived from `conversation`). |
| `TitlebarThread.svelte` | owned by the Shell, not directly imported | The signature — but every Chat Thread *echoes* it (same `--dur-slow`, same `--ease`). |
| `magnetic` action | `app/web/frontend/src/lib/condura/magnetic.ts` | The Send button pollen-magnetic effect (`MOAT §5.3`). 8 px radius cursor-pull. |

---

## 6. DATA FETCHED

Chat reads from stores; the stores wrap the JSON-RPC 2.0 daemon IPC.
No direct fetches inside Chat.

### 6.1 Initial IPC calls on mount

| Call | Store | Cadence | Purpose |
|------|-------|---------|---------|
| `ipc.providersList()` | direct (not in a store) | once, on mount | Populates the model `<select>` in the composer (`Chat.svelte:23`). Failures swallowed — `providers = []`. |
| `settings.config?.llm?.providers` (read) | `settings` store | once | Picks the first enabled `providerName:modelId` for `selectedModel`. |
| `conversation.hydrate()` | `conversation` store | once | Loads the active conversation's messages + last-touched timestamp. |
| `presence.state` | `presence` store | every 5 s on `idle`; event-driven on activity changes | Powers the `Last seen` indicator on the header (`Active / 2 min ago`). |

### 6.2 Conversation list polling

| Call | Cadence | Source |
|------|---------|--------|
| `db.thread.list` (via store `conversationList`) | 60 s on tab focus | `ConversationList.svelte` mounts its own subscription. Inactive tabs (`visibilitychange`) zero out the timer. |

### 6.3 SSE subscriptions

| Stream | Owner | Purpose |
|--------|-------|---------|
| `stream:delta` (SSE) | `conversation.svelte.ts` | Token-by-token stream into `conversation.streamingDelta` (renders the live caret + bubble). |
| `stream:tool` (SSE) | `conversation.svelte.ts` | `conversation.streamingToolCalls` (renders the synapse tool chips). |
| `stream:done` (SSE) | `conversation.svelte.ts` | `conversation.isStreaming = false`; the final turn is appended to `conversation.messages`. |
| `stream:error` (SSE) | `conversation.svelte.ts` | Sets `conversation.streamingError` (renders §2.6). |

Chat does **not** subscribe to SSE itself — the conversation store
owns the listener (per `Chat.svelte:16` comment).

### 6.4 Writes from Chat

| Action | IPC call | Effect |
|--------|----------|--------|
| Send | `llm.chat(providerName, modelId, text)` | Spawns the SSE stream. |
| Stop | `conversation.cancel()` | Closes the stream + transitions state to `interrupted`. |
| Retry | `conversation.retry(lastTurnId)` | Re-sends the last user turn. |
| Rename title | `db.thread.update(id, { title })` | Persists. |
| Pin / unpin | `db.thread.update(id, { pinned })` | Persists. |
| Delete | `db.thread.delete(id)` + `conversation.clear()` | Removes + loads the next conversation (or empty state). |
| New conversation | `conversation.clear()` + `db.thread.create()` | Fires on `⌘N`. |

### 6.5 Store reads (Svelte runes)

| Store | What Chat reads |
|-------|-----------------|
| `conversation` | `messages`, `streamingDelta`, `streamingToolCalls`, `isStreaming`, `streamingError`, `currentTitle`, `lastCost`. |
| `conversationList` | `threads[]`, `loading`, `error`. |
| `settings` | `config.llm.providers[*].enabled` (picks default model). |
| `presence` | `state` (Active / Away / Idle). |
| `voice` | `active`, `transcript`, `micAvailable`. |
| `consent` | `ticket` (disables composer on pending, per §2.9). |
| `halt` | `state.halted` (dims surface per §2.8). |
| `daemon` | `connected` (used to render the header's status chip in `Offline`). |
| `adaptive` | `communication.verbosity` (chips collapse differently). |

### 6.6 What Chat does NOT fetch

- **User API keys** (Settings owns that).
- **Skills** (Skills route owns that).
- **Audit events** (Audit route + Replay own those — Chat links to
  them on click).
- **Sync status** (Sync route owns that).
- **Channel state** (Channels route owns that).
- **Replays** (Replay route owns that — Chat only offers a
  "View in Replay →" inline button when `streamingToolCalls` is
  non-empty).

---

## 7. DESIGN DECISIONS — MOAT compliance

### 7.1 The five tests

| Test (MOAT §) | Chat's pass |
|---------------|------------|
| **1. Restraint** | The empty garden uses one `pollen-float` keyframe (not six). The Thread draw is signature, not decorative. There is no `.alive` misuse — the only one is "alive" in the empty headline (one place, full stop). |
| **2. Detail (×10)** | Press states use the global `.tactile` (no per-component scale overrides). Focus rings track the geometry (the composer is 16 px radius → `0 0 0 2px var(--synapse), 0 0 0 5px var(--pollen-halo)`). Reduced motion has no per-component override (single block in `condura.css`). Empty states teach (3 lines: what / why / next). Loading states draw the Thread (no spinners). Errors guide (3 lines: failure / cause / next). Tactile vocabulary is global. Overlays collapse to the three primitives (`.c-modal`, `.c-sheet`, `.c-popover`). Tooltips replace all `title=` on read-critical affordances. Keyboard story is complete (see §4). |
| **3. Signature (Thread)** | (a) The header hairline draws in on route-enter. (b) The composer's focus-thread (already shipped per `MOAT §5.2`). (c) Between-message vertical Threads on every new turn. (d) The yield gesture on Stop draws the thread out right → left. (e) The error state's `err-hair` (the failure variant). (f) The CTA on the streaming bubble (when the agent finishes a tool, the trailing hairline beneath the bubble draws in). **Six Threads per Chat.** |
| **4. Anti-patterns (×10)** | No gradient text. No emoji as UI icons (every icon is a `<Glyph>`). No glassmorphism on cards (the composer is `--shadow-card`, not `backdrop-filter`). No rainbow accents. No "Welcome to the future" copy. No fake enthusiasm (no "Awesome!" toasts). No spinner loaders (the Thread draws). No rectangular outlines. No double shadows. Every animation carries meaning (stream-bar = "moving"; thread draw = "complete"; pollen halo = "you are here"). |
| **5. $50M feel (×5)** | Cursor catches hoverable surfaces (VoiceOrb / Send / kebab all use `use:hover-region`). Composer focus → inward thread (per `MOAT §5.2`). Stop yields the agent (per `MOAT §5.3` — the Stop label cross-fade + thread reverse-out). NavRail active state has its own animation (decoupled from route mount). Mobile: Chat at `<768 px` is a single column, 44 × 44 touch targets, the composer grows. |

### 7.2 Why each major decision

- **Why is the empty headline three words** instead of a paragraph?
  Because the hero wordmark carries the emotional load (per
  `DIRECTION §1` "Warm, awake"). Three words is the same gesture
  the `Ritual` uses for the wordmark reveal (one name, one gesture).
  More words would make the empty state feel like a tutorial.

- **Why no spinner** for the conversation list? Because the list is
  already populated for most users (the prior reads are cached). A
  spinner would promise "wait, fetching data" when the data is
  already on disk (APPFLOW §I3). A Thread-drawn skeleton is the
  honest teaching gesture — "data is moving."

- **Why pin the (5) Composer at the bottom** rather than let it
  scroll with the conversation? Because the composer is the user's
  anchor. Scrolling it away would force a memory tax ("how do I
  type again?") every few turns. The composer is the *one* place the
  user returns to.

- **Why are the suggested-prompt chips surfaced on empty** instead of
  hidden? Because per `MOAT §2.4` an empty state must teach. Three
  to five concrete phrasings ("Find a file from last week.") reduce
  the cognitive load of a blank page more than a paragraph does.
  The chips also model the *register* the user should write in —
  short, agent-shaped, conditional.

- **Why is the Stop gesture so explicit** (label cross-fade + thread
  reverse-out + pollen pulse)? Because per `MOAT §5.3` the user
  must see the agent *yielding* — not a button becoming unfocused.
  The yield is the cheapest premium animation to ship and the
  loudest silence in the product.

- **Why is the VoiceOrb in the composer** rather than in the
  titlebar? Because the composer is the user's writing surface; the
  voice is a writing input. Pairing them physically invites the user
  to swap modalities mid-thought without an extra hop.

### 7.3 The single success criterion

> **The user can have a single, sustained, focused conversation
> with Condura — from welcome garden to tool-call chips to a
> finished answer — without once reaching for a UI outside of Chat.**

If the user has to leave Chat to recover context, the surface is
wrong. If the user feels heard by Chat, the surface is right.

---

## 8. DRIFT TABLE

The current `Chat.svelte` predates this spec. The table below is the
gap. The Phase implementer is expected to close it section by
section; closing rows are deleted, not edited.

### 8.1 To be REMOVED from current `Chat.svelte`

| Row | What | Why (per the spec / MOAT) |
|-----|------|---------------------------|
| R1 | `<select>` for **model pick** rendered inside the textarea footer (lines 268–274). | A `<select>` is the wrong affordance for "12 providers × N models." It should be a sheet / picker triggered from the header's model badge. The composer keeps the *current* model as a faint mono label only. |
| R2 | Inline `errorHead` "We couldn't reach the daemon." rendered in `<div class="err-state">` (lines 213–221). | Collapses to the global `<ErrorState>` component (MOAT §1.2). |
| R3 | The `err-hair-draw` keyframe declared inline in `Chat.svelte:388` (the `err-hair` ruleset). | `err-hair-draw` is owned by `<ErrorState>` (and by `condura.css` as the canonical recipe). Chat never declares it. |
| R4 | The reduced-motion `@media` block in `Chat.svelte:391–396`. | Owned exclusively by `condura.css` (MOAT §2.3). Chat never re-declares. |
| R5 | Inline `onMount` `motes` array + `pollen-float` keyframe (lines 165–176, 328–337). | Collapses to `<Garden>` (new component). The keyframe is shared with `About.svelte`'s motes — both read it from `condura.css` (added as a global if not already present). |
| R6 | The Stream-bar `travel` keyframe (lines 459–470) — the pollen dash that travels left → right. | This keyframe is fine; rename it to `stream-dash-tick` in `condura.css` so other surfaces can use it (Replay, Audit). |
| R7 | The inline `send-arrow` SVG/HTML arrow `↗` raw character in the Send button (line 282–283). | Use `<Glyph name="send">` from `icons.ts`. The arrow character is a system-font fallback; the Glyph is the brand vocabulary. |
| R8 | The `<Glyph name="stop">` glyph in the Stop button (line 278) — review only; this one may already be correct. | Confirm in `icons.ts`; if missing, add `stop` (a square −4° rounded). |
| R9 | The `last()` module-context helper (lines 291–296) — exported, unused at the moment. | Delete. The conversation list virtualization already handles "last item" via `turns.length`. |
| R10 | The two `magnetic` import (line 9) + use (line 282). | Keep — this is correct per `MOAT §5.3`. Just clarify the import path is from `magnetic.ts`, not bundled. |

### 8.2 To be ADDED to current `Chat.svelte` (and refactored into named components)

| Row | What | New location |
|-----|------|--------------|
| A1 | `(1)` `<ConversationHeader>` — extracts the current header into a sticky top region. | new: `ConversationHeader.svelte` |
| A2 | `(2)` `<ConversationList>` — the 280 px rail (with `+ New`, rows, footer whisper, polling). | new: `ConversationList.svelte` |
| A3 | `(3a)` `<EmptyState>` — empty garden + hero + 3–5 chips. | new: `EmptyState.svelte` |
| A4 | `(3a-garden)` `<Garden>` — the 14 pollen motes; randomized once per mount. | new: `Garden.svelte` |
| A5 | `(3b-message)` `<Message>` — one turn; user / assistant / system / tool variants. | new: `Message.svelte` |
| A6 | `(3b-feed)` `<MessageFeed>` — virtualized list, binary search, spacer divs, auto-resize. | new: `MessageFeed.svelte` |
| A7 | `(3c-error)` Mount `<ErrorState>` from `MOAT §1.2`. | already exists |
| A8 | `(5)` `<Composer>` — pinned card with focus-thread, model label, VoiceOrb, send/stop. | new: `Composer.svelte` |
| A9 | `VoiceOrb` — the 28 × 28 round with breathe + waveform glyph states. | new: `VoiceOrb.svelte` |
| A10 | `<Tooltip>` on VoiceOrb (mic not granted), kebab (more), pencil (rename). | new: `Tooltip.svelte` |
| A11 | `<Menu>` for the kebab + per-row actions (pin, rename, delete, export). | new: `Menu.svelte` |
| A12 | State-driven logic: `2.1`-`2.11` — empty, loading, streaming, interrupted, error, no-provider, halted, consent, voice-active, hover. | `Chat.svelte` orchestrator + child components |
| A13 | Composer focus-thread `::before` (already in `Chat.svelte:515–528`) — verify it spans the inner padding when the textarea grows. | `Composer.svelte` |
| A14 | Reduced-motion + energy-mode compliance: no per-component `@media` blocks. The garden uses CSS `display: none` on `prefers-reduced-motion` via the global `condura.css`. | global only |
| A15 | Keyboard handler for `↑` (edit last), `⌘.`, `⌘R`, `⌘D` per §4.1. | `Composer.svelte` + `Chat.svelte` orchestrator |
| A16 | `aria-live` polite on the error-state and on the streaming bubble. | `<Message>` |

### 8.3 To be carried over AS-IS

| Row | What | Location in current `Chat.svelte` |
|-----|------|------------------------------------|
| K1 | The virtualized-list logic (binary search + `BUFFER_PX` + `topPad`/`botPad`) | lines 60–98 |
| K2 | The auto-scroll pin-to-bottom (200 px threshold + double-rAF) | lines 122–137 |
| K3 | The re-measure after windowed change | lines 139–143 |
| K4 | The `magnetic` send affordance (`use:magnetic` directive) | line 10 / line 282 |
| K5 | The composer focus-thread `::before` (`scaleX(0 → 1)` on `:focus-within`) | lines 515–528 |
| K6 | The streaming caret + stream-bar dot | lines 444–470 |
| K7 | The tool-chip pattern | lines 471–484 |
| K8 | The Mono-uppercase eyebrow (`m-label`) | lines 414–424 |

### 8.4 Sign-off criteria

`Chat.svelte` is spec-compliant when:

1. All rows in 8.1 are removed and the corresponding rows in 8.2 are
   in place.
2. The empty state renders the three-line copy from §1.3a verbatim.
3. Every state in §2 is reachable via the documented trigger and
   renders the documented content.
4. The Thread draws in for every completion moment listed in §3.
5. Every keyboard chord in §4 is bound (the wildcard: `↑` edit-last
   must work without breaking IME composition).
6. All animations use the four duration tokens (no `320 ms`, no
   `400 ms`, no `600 ms` — these are spec bugs if they appear).
7. `make verify` is green.

When all seven hold, the spec and the implementation agree.
Append-only in spirit: any future divergence goes in a new row
of §8.1 / §8.2, never silently edited.

---

## 9. Versioning + change control

- **Spec version:** 0.1.0 — pinned to the v0.1.0 Chat surface.
- **Update policy:** append-only in spirit, per `CLAUDE.md §30.5`.
  Corrections are not edits — add a row to §8.1/8.2 and to the
  LOGBOOK.
- **Implementation divergences:** when the implementation disagrees,
  fix the implementation. When the implementation has a reason the
  spec doesn't cover (e.g., a new keyboard chord the spec omitted),
  add a row to §8.2 explaining the cause, in the same commit.
- **Phase hand-off:** §1–§7 are implementable as-is. §8 is a
  per-phase ticket list. Closing §8 in order closes the drift.

---

**This document is the architecture. The code is the implementation.
They agree. When they diverge, the divergence is the spec-bug — fix
the doc, then fix the code, in one commit.**
