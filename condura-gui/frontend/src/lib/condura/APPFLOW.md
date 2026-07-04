# APPFLOW.md — Condura

> **The complete user flow specification for the Condura desktop shell.**
> This document is the source of truth for every screen the user can land in,
> the transitions between them, and the data each screen writes to the daemon.
>
> Implementation lives in `app/web/frontend/src/lib/condura/`. The daemon
> contract is JSON-RPC 2.0 over Unix socket (see `app/web/frontend/src/lib/ipc/`).
>
> ---
>
> **Reading order.** Read this top-to-bottom: pre-ritual → main shell →
> edge cases. Each screen is named by the route or component that renders it,
> so engineers can search the repo.

---

## Table of Contents

1. [Design Intents (read first)](#1-design-intents-read-first)
2. [Pre-Ritual Flow — First Run](#2-pre-ritual-flow--first-run)
3. [Main App Flow — Post-Ritual](#3-main-app-flow--post-ritual)
4. [Route-by-Route Reference](#4-route-by-route-reference)
5. [Settings — The Heart of Configuration](#5-settings--the-heart-of-configuration)
6. [Edge Cases & Failure States](#6-edge-cases--failure-states)
7. [State Inventory](#7-state-inventory)

---

## 1. Design Intents (read first)

Five principles drive every screen below. If a screen contradicts one, the
screen is wrong.

| # | Intent | What it forbids |
|---|---|---|
| **I1** | **Configure, not comply.** Every first-run question asks the user to *choose*, never to *accept*. | "I agree" walls, forced opt-ins, dark-pattern defaults. |
| **I2** | **Smooth is honest.** Animation means a verb: thread-drawing = the thing arrived; pulse = heartbeat; pollen halo = human CTA; route enter = a real state change. | Decorative loops, ornamental easing, motion that misleads. |
| **I3** | **Local-first feels local.** No spinners that imply backend waits when the answer is already on disk. | Loading states that lie about where the data is. |
| **I4** | **Every state is reachable.** Daemon down, no Ollama, no API key, kill switch on, permissions denied — all are real screens a user can land on. | Dead walls, "service unavailable" with no fallback. |
| **I5** | **The 7 invariants are visible.** The About page renders the survival rule. The gatekeeper consent modal precedes every physical action. The audit chain is a thread, not a log. | Hiding safety, "convenience" that skips the armor. |

These are derived from `CLAUDE.md §2` (the Survival Rule), `§10` (Safety
Layer), and `§20` (Onboarding Flow).

---

## 2. Pre-Ritual Flow — First Run

### 2.1 What the user sees

The Ritual is not a wizard. It is a **full-bleed, frameless pre-window**: it
*is* the whole window on first run, with no shell behind it. When the user
finishes, `completeOnboarding(routeHash)` is called from `Ritual.svelte`,
which dissolves the ritual (700ms fade + blur) and the Shell renders in the
same window.

**Implementation:** `Shell.svelte:187–244`, `Ritual.svelte:1–51`.

### 2.2 The 9-Step Sequence (current Ritual.svelte)

These 9 steps execute in order. The first 3 (Arrival → EULA → Permissions)
are required; the middle 4 (Power → Hotkey → Voice → Channels) are
recommended but skippable; Account is optional; Breath is the handoff.

| # | Step | Component | Required? | IPC Writes |
|---|---|---|---|---|
| 1 | Arrival | Ritual (arrival block, line 488) | yes — must click Begin | — |
| 2 | EULA | Ritual (eula block, line 493) | **yes** — record acceptance | `onboarding.acceptEula(version)` |
| 3 | Permissions | Ritual (permissions block, line 516) | yes (skip allowed) | `onboarding.completePermissions()` or `onboarding.skipStep('permissions')` |
| 4 | Power | Ritual (power block, line 539) | skip allowed | `onboarding.probePower()` (read) — choice is local; actual key paste happens in Settings |
| 5 | Hotkey | Ritual (hotkey block, line 559) | required (no silent default) | `onboarding.saveHotkey(combo)` |
| 6 | Voice | Ritual (voice block, line 577) | skip allowed | `onboarding.probeVoice()` (read); wake word config persists via `config.update` |
| 7 | Channels | Ritual (channels block, line 596) | skip allowed (Telegram is the only wired channel) | — (selection saved locally) |
| 8 | Account | Ritual (account block, line 610) | skip allowed | `account.signInWithEmail(email, locale, origin)` |
| 9 | Breath | Ritual (breath block, line 631) | yes — must click Enter Condura | `onboarding.finish({hotkey, eula_version, permissions_skipped})` |

### 2.3 Per-Screen Specification

#### 2.3.1 Arrival — `Ritual.svelte:488–492`

- **Purpose:** open the door, no commitment.
- **Options presented:** one (Begin →).
- **Animation:** enters with the awakening sequence (lines 354–360, 1480–1567).
  - 0–1400ms: paper void fades in (`.a-void`, 600ms hold).
  - 1100–3400ms: a pollen mote drifts from 38% to 50% of the viewport.
  - 1400ms: the wordmark "Condura" reveals via a horizontal clip-path wipe (`.wordReveal`, 900ms).
  - 1500ms: an underline draws in below the wordmark (path-stroke offset 1→0).
  - 2600ms: a first pollen beat fires (`.firstBeat`, 900ms), then becomes a continuous breathing pulse (5s).
- **State (loading/error/done):** none — pure visual. Honors
  `prefers-reduced-motion` by skipping the awakening overlay entirely.
- **Writes to daemon:** none.

#### 2.3.2 EULA — `Ritual.svelte:493–515`

- **Purpose:** legal consent happens before any system access. Live `onboarding.eula`
  via `ipc.onboardingEula()` is preferred; if the daemon is unreachable, falls
  back to `fallbackEula.ts` (a local copy of the Synaptic Freeware EULA v1).
- **Options presented:**
  1. **Stamp to accept** (radial seal button, disabled until the user has
     ticked "I have read and accept the Condura Freeware EULA"). Self-resolves
     when the text fits on one screen (no scroll required).
  2. **Not now · quit** (italic skip-note in the bottom-left).
- **Animation:**
  - Stamping: the seal scales 0.94, the radial sealBloom ring expands
    (28px → 0px, 600ms ease-out).
  - A left-side `eula-read` bar tracks scroll progress (height %= scrollTop/max).
- **State:**
  - Loading: "Loading the license…" placeholder text.
  - Fallback: italic "Read offline (daemon unreachable) — your acceptance will
    be replayed to Condura on next boot."
  - Stamped: "Accepted · thank you / a considered act — not a click" lingers 650ms.
- **Writes to daemon:** `onboarding.acceptEula(eulaVersion)`.

#### 2.3.3 Permissions — `Ritual.svelte:516–538`

- **Purpose:** OS grants for Accessibility + Screen Recording (the two permissions
  computer use actually needs).
- **Options presented:**
  - Two permission rows (Accessibility, Screen Recording) with a status badge
    (`granted` / `denied` / `unknown`) and an **Open System Settings →** deep link
    (uses Wails `runtime.BrowserOpenURL`, falls back to clipboard + `window.open`).
  - **Continue →** (always enabled — sync marks step complete, advances).
  - **not now** (italic skip-note).
- **Polling:** every 2s via `setInterval(refreshPerms, 2000)` for the duration of
  the step.
- **State:**
  - Normal: live status badges; borders turn synapse-green when `granted`.
  - Error: `.rit-err` block — "We couldn't read the permission status." italic,
    with the raw error inline.
- **Animation:** the per-row border transition (200ms) when status flips.
- **Writes to daemon:** `onboarding.completePermissions()` or `onboarding.skipStep('permissions')`.

#### 2.3.4 Power — `Ritual.svelte:539–558`

- **Purpose:** pick a power source — Ollama local, paste an API key, or connect a subscription.
- **Options presented (radio cards with title + meta + optional tag):**
  1. **Local — Ollama.** Recommended tag appears if `onboarding.probePower()`
     reports `ollama_reachable: true`. Meta line lists detected models.
  2. **Paste an API key.** Meta: "Anthropic · OpenAI · Google · xAI · more".
  3. **Connect a subscription.** Meta: "Claude Pro · ChatGPT Plus · Gemini · SuperGrok". (Not implemented in v0.1.0 — the user is directed to Settings.)
- **State (loading):** until the probe resolves, the cards render with a generic
  meta line.
- **Writes to daemon:** none directly — the user's choice is recorded locally
  (`powerChoice`), then `onboarding.finish` records the EULA version + hotkey.
  When the choice is `apikey` or `sub`, the Shell routes to `#/settings` after
  completion so the user can paste the key (line 348).

#### 2.3.5 Hotkey — `Ritual.svelte:559–576`

- **Purpose:** the global combo to summon the quick-prompt overlay. Per locked
  decision #8, no silent default.
- **Options presented:**
  - **Live combo.** Click the keycap surface, the surface enters
    `recording` state (dashed border → solid pollen + halo), the user's next
    keyboard combo populates `combo`. `Escape` cancels recording.
  - **Three presets:** Option+Option, Cmd+Shift+Space, Ctrl+Space.
  - **Try it** — once a combo is set, the button pulses; pressing it sets
    `tried = true` for 700ms (a visual confirmation).
  - **set up later** (italic skip-note).
- **State:** the Continue button is disabled until a non-empty combo + daemon is
  not busy.
- **Writes to daemon:** `onboarding.saveHotkey(combo)`.

#### 2.3.6 Voice — `Ritual.svelte:577–595`

- **Purpose:** opt into the wake word "hey condura". Optional. All platforms.
- **Options presented:**
  - **Enable voice →** (records in `wired` set, marks step as enabled).
  - Implicit skip via the bottom-left "leave this for later" link.
- **State:**
  - Normal: `voiceProbe.mic_available` resolves to a Pulse + name + meta line.
  - Error: italic "We couldn't probe the microphone." with raw error inline —
    Continue is still enabled.
- **Writes to daemon:** the wake config persists via `config.update` from Settings.
  No write during the ritual itself.

#### 2.3.7 Channels — `Ritual.svelte:596–609`

- **Purpose:** opt into messaging channels (Telegram is the only wired channel in v0.1.0).
- **Options presented:** five channel cards.
  - **Telegram** (ready).
  - **WhatsApp** / **Slack** / **Discord** / **iMessage** — greyed out (`.dim`,
    cursor not-allowed) with state "v0.2.0".
- **Animation:** hover lifts non-dim cards; selected cards turn synapse-green.
- **State:** Continue is always enabled.
- **Writes to daemon:** none — selection saved to local `channelPick` Set.

#### 2.3.8 Account — `Ritual.svelte:610–630`

- **Purpose:** optional magic-link sign-in for hub/donations/support.
- **Options presented:**
  - Email field + **Send magic link →** button (disabled until a non-empty email).
  - **skip — I'll do this later →**.
- **State:**
  - Sending: `accountBusy = true`, button disabled.
  - Sent: green-bordered "Check your email" card with a Pulse `ok`.
  - Error: italic "The magic link didn't go through" + raw error.
- **Writes to daemon:** `account.signInWithEmail(email, locale, origin)`.

#### 2.3.9 Breath — `Ritual.svelte:631–637`

- **Purpose:** the closing ceremonial handoff.
- **Aesthetic:** center-aligned hero, a 48px breathing pollen sphere, the line
  "Condura is here." in synapse italic.
- **Options:** **Enter Condura →** (calls `onboarding.finish`, then dissolves the ritual 900ms later).
- **Animation:** the ritual wrapper gets `.dissolving` class → 700ms opacity→0 + blur(8px). The Shell renders underneath.
- **Side effect:** navigates the user to `#/settings` if `powerChoice === 'apikey' || 'sub'`, otherwise to the default chat route (`undefined`).

### 2.4 Animation Choreography Summary

| Step | Enter | Exit |
|---|---|---|
| Arrival | awakening (1400ms reveal) | Begin click → step 2 fade-in |
| EULA | scroll-to-bottom arm, seal glow | sealBloom (600ms) |
| Permissions | rows fade in with `permPoll` data on 2s cadence | row checkmark slide |
| Power | radio dots fade | radio dot ring |
| Hotkey | keycap surface `.recording` | keycap settles |
| Voice | voice card lights up | — |
| Channels | channel chips lift on hover | — |
| Account | email focus ring | sent state cross-fade |
| Breath | breath-pulse | `.dissolving` 700ms fade + blur |

All animations respect `prefers-reduced-motion: reduce` — durations drop to 0
and decorative overlays are removed.

### 2.5 Re-Ritual on Boot

`Shell.svelte:81–102` polls `firstRunStatus()` + `onboardingIsComplete()` on mount.
If either returns false, or the local `condura-ritual-seen` flag is absent, the
Ritual shows. The dev affordance **Shift+O** re-opens the Ritual at any time
(marked for removal before ship per the inline comment).

---

## 3. Main App Flow — Post-Ritual

### 3.1 First-Time-in-Shell

After the Ritual dissolves, `completeOnboarding()` in `Shell.svelte:164–167`
hides the ritual and either sets `window.location.hash` to the requested
destination or leaves it at `#/` (the chat route). The Shell renders with the
**NavRail hidden only when overlay mode is active** (`Shell.svelte:212`), so the
first thing the user sees on the chat route is:

1. **The titlebar** (`Shell.svelte:196–210`):
   - Wordmark "Condura" with the pollen dot accent (left).
   - A live `<TitlebarThread>` (the titlebar hairline that bends toward the
     cursor, paused on visibilitychange + IntersectionObserver — 0% CPU in
     background tabs).
   - A `<DynamicIsland>` showing the current agent phase (idle / thinking / consent / error).
   - `<kbd>⌘</kbd><kbd>K</kbd>` hint + a moon/sun theme toggle (single-button
     override; the proper segmented picker lives in Settings — see §5).
2. **The 232px-wide NavRail** on the left (`NavRail.svelte:50–82`):
   - 10 nav items: Chat, Hub, Skills, Sync, Audit, Replay, Channels, Delegation, Settings, About.
   - Each shows a glyph + label, scales 1.04 on hover, draws a vertical
     synapse thread segment on activation (left side, `.active::before`).
   - A bottom footer with `Local · Ollama` indicator + a pollen "you" dot.
3. **The main surface** — `Chat.svelte` by default.
   - Empty state: garden motes drift, "<wordrise>Your</wordrise> <wordrise>computer,</wordrise> <wordrise>alive.</wordrise>" hero + composer pinned at the bottom.
   - A model selector dropdown (`<select>` filled from `ipc.providersList()`).
   - Send / Stop button toggles based on `conversation.isStreaming`.
   - Bottom mono hint: `⌘↵ to send · Esc to stop · your hotkey to summon`.

### 3.2 Navigation

Three ways to navigate, all routed through the same hash-based router.

| Method | Mechanism | Source |
|---|---|---|
| **NavRail click** | `<button>.nav-item` → `onnavigate(route)` → `window.location.hash = ROUTE_HASH[route]` | `NavRail.svelte:67–76` |
| **⌘K palette** | `Cmd+K` / `Ctrl+K` anywhere on the page → `paletteOpen = true` → fuzzy search → run command (route or action) | `Shell.svelte:110–115`, `CommandPalette.svelte:1–260` |
| **Shift+P** | Quick Prompt overlay (summons the floating text input) | `Shell.svelte:121–124` |

**The 10 routes** (hashes from `NavRail.svelte:16–27`, `hashToRoute` from
`NavRail.svelte:30–41`):

```
#/             chat          (default)
#/hub          hub           public Skills Hub
#/skills       skills        local installed skills
#/sync         sync          P2P device pairing
#/audit        audit         HMAC-chained event log
#/replay       replay        action replay (last 24h)
#/channels     channels      Telegram etc.
#/delegation   delegation    sub-agent constellation
#/settings     settings      flowing settings document
#/about        about         colophon + 7 invariants
```

Route transitions: `Shell.svelte:217–241` wraps the current route in
`{#key route}` + `.route-enter` class so the new route re-mounts on every
navigation (a clean unmount/remount animation per route).

### 3.3 The Shell Itself

```
┌────────────────────────────────────────────────────────────────────┐
│ Titlebar (44px tall)                                               │
│ Wordmark · TitlebarThread · DynamicIsland · ⌘K · ☼/☾ toggle        │
├──────────────┬─────────────────────────────────────────────────────┤
│              │                                                     │
│  NavRail     │   Main surface                                      │
│  (232px)     │   (the current route)                               │
│              │                                                     │
│  Chat        │                                                     │
│  Hub         │                                                     │
│  Skills      │                                                     │
│  Sync        │                                                     │
│  Audit       │                                                     │
│  Replay      │                                                     │
│  Channels    │                                                     │
│  Delegation  │                                                     │
│  Settings    │                                                     │
│  About       │                                                     │
│              │                                                     │
│  Local·Ollama│                                                     │
└──────────────┴─────────────────────────────────────────────────────┘
```

Overlay mode (when the daemon reports an active overlay session): the NavRail
hides via `Shell.svelte:212` and the layout collapses to a single column.

### 3.4 ⌘K Command Palette

`CommandPalette.svelte`. Opens over a blurred scrim (8px backdrop-filter). 13
commands: 10 nav routes + 3 actions (Toggle Theme, Summon Quick Prompt, Stop
Everything).

**Signature interactions:**
- A **sliding pollen highlight** (`top: highlightTop`, 200ms ease) follows the
  active row as focus travels (lines 426–441). This *is* "focus traveling"
  given visual form.
- A **synapse hairline** across the top edge of the panel draws left→right
  over 320ms when the palette opens (`.panel::before`, lines 320–337).
- A **match-flash** (180ms) on the `↩` hint of the active row on every active
  change (the "press enter" affordance, lines 235–241 + 496–502).
- An **input hairline** that draws from the center outward on focus
  (`input-draw`, 240ms, lines 378–381).

Keyboard: ↑/↓ navigate, ↵ run, Esc close.

### 3.5 Quick Prompt Overlay

`QuickPromptOverlay.svelte`. The hero interaction. Floating paper card at
top-center, summoned by the configured overlay hotkey (default `Option+Option`,
configurable via Settings → Hotkey).

- **Top-edge thread-draw** on summon (the overlay arriving).
- **Breathing Pulse** in idle state; `acting` (synapse) when streaming.
- **Magnetic pollen Send button** — supports the `use:magnetic` directive
  (slight cursor-pull within a small radius).
- **Auto-dismiss** after 5s idle (never mid-stream, never with unsent draft;
  the timer re-arms).
- **Esc** closes globally (window-keydown listener).
- Enter sends; Shift+Enter newlines; `isComposing` blocks Enter while IME.

### 3.6 Dynamic Island & Thread

`DynamicIsland.svelte`: a small pill in the titlebar that morphs width and color
based on the agent phase:
- `idle` (synapse) — agent waiting
- `thinking` (pollen, animated) — streaming in progress
- `awaiting` (warn) — terminal user-action required
- `acting` (synapse filled) — actively executing
- `ok` (synapse-glow, stable) — last action succeeded
- `listening` (pollen, breathing) — wake word / mic
- `error` (danger) — last action failed
- `consent` (warn) — gatekeeper consent modal is open
- `paused` (warn) — paused by watchdog

`TitlebarThread.svelte` draws a thin synapse line across the titlebar with a
pollen node that subtly bends toward the pointer position
(`pointermove → rAF`, 0.06 lerp). Pauses on `visibilitychange` and
IntersectionObserver (zero CPU in background tabs).

---

## 4. Route-by-Route Reference

### 4.1 Chat — `Chat.svelte` — `#/`

The hero. The composer is pinned; the message list virtualizes for
performance (the DOM holds only the visible window + a 600px buffer; `firstIdx`
computed by binary search over `itemOffsets`).

**Primary action:** Ask Condura to do something. ⌘↵ to send. Esc to stop.

**Empty state:** 14 garden motes drift (CSS keyframe `pollen-float`),
wordrise hero "Your / computer, / alive." appears with a clip-path wipe on
each line.

**Mid-stream error:** italic Instrument Serif error block with a hairline that
draws left→right (`err-hair-draw`, 600ms).

**Streaming:** a `.caret` blinks in the active bubble; a `.stream-bar` carries
a pollen dash (`travel` 2.6s ease infinite); tool-call chips render below.

**Data fetched:**
- `ipc.providersList()` — once on mount, populates the model `<select>`.
- `settings.config?.llm?.providers` — picks the first enabled provider/model for the default.
- `conversation.messages`, `conversation.streamingDelta`, `conversation.streamingToolCalls` — driven by the SSE stream listener owned by the conversation store.

**Auto-scroll:** keeps the scroll pinned to the bottom as the stream grows
(the user can scroll up; if they were already within 200px of the bottom, it
re-snaps).

### 4.2 Skills — `Skills.svelte` — `#/skills`

Local installed skills as a card index. Each card is a procedure. Auto-created
skills carry a **green thread** (agent-authored); user-authored skills carry a
**pollen thread** (`.thread[data-author]`).

**Primary action:** click a card → right-side detail sheet slides in
(`slide-in`, 520ms ease). Run / Improve buttons in the sheet.

**Data fetched:** `ipc.skillsList(100)` once on mount; reload on user click.

**Loading:** "INDEXING…" with a Pulse `thinking`.
**Empty:** "No skills yet. Run a complex task — Condura will save the procedure
as a skill automatically."
**Error:** italic + retry button + err-hair.

### 4.3 Hub — `Hub.svelte` — `#/hub`

Public Skills Hub as a **3D bookshelf**. Each skill is a slim vertical spine
(36px wide, 200px tall) with `transform: rotateY(-22deg)` for the angled book
face. Hover tilts further to `-30deg` and lifts 6px. Installed spines turn
synapse-green. Tag chips filter (up to 6 unique tags from current results).

**Detail sheet:** slides in from the right (`slide-in-right`, 520ms). Install
button → `hub.install(id)`. After install, the spine highlights synapse; the
button reads `Installed ✓`.

**Data fetched:** `hub.search(query)` (debounced 250ms). On mount, a broad
'skill' query primes the shelf.

### 4.4 Sync — `Sync.svelte` — `#/sync`

P2P device pairing as a **garden of nodes threaded together**. This device is
the green node at the centre; each paired device is a pollen node. The
**threads between them are the sync links** — they draw in on mount
(`sync-draw`, 1.1s) and breathe at 9s. Revocation **frays** the thread
(`sync-fray`, 0.9s).

**Pending pairing card** (centred overlay): pollen TTL ring depleting,
pressable `Seal link` button. PIN input is 4–8 digits. ESC closes.

**Peers rail:** discovered peers on the LAN (`mDNS`); each has a Pair button
that starts a pairing flow. The store handles the mDNS / handshake.

**Data fetched:** `sync.status()`, `ipc.syncStatus()`, polled every 5s. On
identity change, regenerates the **QR** of `{v, device_id, name}` via `qrcode`.

**Integrations:**
- `Sync.svelte` exports nothing; the entrypoint is the `#/sync` route.
- `PairingModal.svelte` lives in `app/web/frontend/src/lib/components/`
  (older variant); the modern surface is the embedded pending card in `Sync.svelte`.
- For pairing ad-hoc (e.g. floating interview), the components pair modal
  takes the same RPC contract (`sync.pairWith`, `sync.confirmPairing`).

**Integrity:** every device in the canvas is clickable; clicking opens a
revoke popover with explicit "Keep" / "Revoke" actions (revoke is destructive).

### 4.5 Audit — `Audit.svelte` — `#/audit`

The HMAC-chained event log visualized as a **vertical synapse thread
(`.thread-spine`)** with one node per event. Click a node → right detail pane
fills with the full record.

**Filters:** action text + 4-level chip (all / info / warn / error).

**Empty:** "The chain is quiet." when zero events; "Nothing on this shelf matches." when filtered.

**Data fetched:** `audit.refresh()` on mount; `audit.prevPage()` /
`audit.nextPage()` paginate via store RPCs.

**Detail panel:** sticky right rail. Result dot color: ok (allow), danger
(block), warn (prompt), faint (other).

### 4.6 Replay — `Replay.svelte` — `#/replay`

Last 24 hours as a **scrubbable synapse thread**. The thread at the bottom IS
the scrubber: green from the left edge to the playhead, hairline after.
A **pollen mote** rides the playhead.

**Frame view:** Before / After screenshots side-by-side + a **decision
receipt** (italic "I decided to…", outcome reasoning). Arrow keys step
frame-by-frame; Home/End jump.

**Integrity badge:** a `Chain intact` / `Chain broken` pill (top-right).
Click → `replay.verifyIntegrity()` runs the HMAC re-verification.

**Export:** `.mp4` export via `replay.exportMP4()`; result printed as a mono
status line ("Exported to /path/...").

**Data fetched:** `replay.refresh()` + `replay.verifyIntegrity()` on mount.
Pointer drag on the scrubber → `replay.selectIndex(i)`.

### 4.7 Channels — `Channels.svelte` — `#/channels`

Five channel rows, **signal-bars metaphor** (5 dots at stepped heights 8/12/16/20/24px):

- **connected** → all 5 dots on, breathing in cascade (1.6s, 0.12s per dot delay).
- **degraded** → 3 dots on (warn color), static.
- **off** → 0 dots on.
- **soon** → row dim (opacity 0.55), `v0.2.0` pill instead of Connect.

**Telegram Connect:** opens BotFather in the system browser
(`runtime.BrowserOpenURL` with `window.open` fallback), then calls
`channels.telegram.start` to nudge the daemon. Local row state flips to
degraded with hint "token entry → open Channels".

**Footer note:** "Outbound messages always pass the consent gate. Inbound traffic is logged on the Audit chain." (a `.threadlink` button that
navigates to `#/audit`).

**Data fetched:** `ipc.channelsList()` on mount (best-effort hydration); if it
fails, the rows fall back to the local defaults (truthful).

### 4.8 Delegation — `Delegation.svelte` — `#/delegation`

The **constellation**: Condura at the centre, 7 sub-agent CLIs orbiting on a
340px dashed ring (Claude Code, Codex, Antigravity, OpenCode, Kilo, Hermes,
Ollama). Each node has a 6px pollen pulse-dot (2.4s breathing). Hover pops
out a tooltip card with name + blurb + footer.

**Thread legs:** diagonal connections split into H + V segments so the SVG
threads reach the centre cleanly. Class `.leg.hover` turns the synapse line
pollen.

**Live panel** (right rail): every pending sub-agent action surfaced from the
Gatekeeper. Each row has **Approve / Deny / Run** buttons. States via
`badge-{tone}` — pending (warn), approved (ok), denied / failed (danger),
executed (synapse).

**Empty canvas:** if zero sub-agents are wired (v0.2.0 case), a Glyph + "No sub-agents wired" caption renders. Doesn't fake it.

**Data fetched:** subscribes to `pendingActions` store; `startPolling(5000)`
on mount (5s cadence per the brief), `stopPolling()` on destroy.

### 4.9 About — `About.svelte` — `#/about`

The colophon. The quietest surface in the product.

**Header:** "Made by a human and an AI, in partnership."

**Ledger:** the seven non-negotiable invariants, rendered as
**01..07** numbered rows in italic Instrument Serif. Each row carries:
  - A 1px hairline that **draws in left→right when the row scrolls into
    view** (IntersectionObserver threshold 0.35).
  - A protective synapse armor `<rect>` that **paints on hover** with a
    `pathLength=1` `stroke-dashoffset` from 1→0 (the live "armoring"
    gesture for each promise).
  - The italic title + body of the invariant.

**Footer:** a horizontal Thread + Pulse "breath", then the colophon line
"EULA · Privacy · Support Condura".

### 4.10 Settings — `Settings.svelte` — `#/settings`

See §5 below. It's the second-most-complex surface (only Chat is more so).

---

## 5. Settings — The Heart of Configuration

Settings is intentionally **not** a modal and **not** a tabbed panel. It's a
flowing document (`.settings`) — italic Instrument Serif section titles
separated by hairlines, no sidebar, no top tabs. Sections are bookmark-anchorable
via the startWith-hash trick (`#/settings/legal` still maps to settings
per `NavRail.svelte:31`).

### 5.1 Sections in Order

| # | Section | Component / Source |
|---|---|---|
| 1 | **Appearance** | Theme (3-way seg), Motion strength slider, Grain intensity slider |
| 2 | **Power** | Energy budget seg (low / balanced / high / auto), Default model per provider |
| 3 | **Autonomy matrix** | hero — 11 task types × 3 states (block / warn / autonomous), live preview line |
| 4 | **Adaptive engine** | Learning strength — 4 dots (off / cautious / balanced / aggressive) |
| 5 | **Voice** | Wake word toggle + sensitivity slider |
| 6 | **Account** | Signed-in chip OR "sign in to sync skills and hub bookmarks" inline link |
| 7 | **Permissions** | Read-only list (granted/denied/unknown badges) |
| 8 | **Legal** | EULA inline expander (load `ipc.onboardingEula()` on first open) |

### 5.2 The Autonomy Matrix — `Settings.svelte:614–672`

The hero of Settings. 11 canonical task types from CLAUDE.md §27:
`coding`, `file_operations`, `web_browsing`, `email`, `calendar`, `messaging`,
`shell_commands`, `computer_use`, `research`, `image_generation`, `code_review`.

Each type is a row of 3 dots (block / warn / autonomous). Click an inactive dot
→ set state; click the active dot → cycle forward. Active dot pops with a
180ms scale animation (`dot-pop`, lines 1193–1197).

The **preview line** below renders "Right now, for **coding**, Condura will
**{verb}** {tail}" with the verb in the state color (synapse for block, warn,
ok for autonomous).

**State colors:**
- block → `var(--synapse)` (synapse green)
- warn → `var(--warn)` (amber)
- autonomous → `var(--ok)` (green)

### 5.3 The Save Bar

Daemon-backed changes (autonomy matrix + per-provider default model) set a
`dirty` flag. When dirty, a **sticky pollen save bar** springs in from the
bottom (`fly` transition, 380ms backOut easing, `.save-bar` line 1513).

- Status: idle shows "Unsaved changes" with a Pulse `awaiting`.
- Saving: "SAVING…" + Pulse `acting`.
- Saved: green Pulse `ok` + "Saved", then the bar fades over 400ms (the
  `.save-bar--saved` opacity transition).
- Failed: Pulse `error` + "Save failed" + `settings.lastSaveError` shown.

Local UI prefs (theme, motion, grain, energy, wake word, wake sensitivity) are
applied at once via CSS variables + persisted to localStorage. They **never**
trigger the save bar.

### 5.4 The Theme Picker — `ThemePicker.svelte`

The dedicated segmented picker (sun / auto / moon) is wired in `Settings.svelte`
(`Auto / Light / Dark` seg, lines 502–520). **It must be prominent** in
Settings — first row under "Appearance".

The picker uses the **palette-switch choreography** described in the inline
JSDoc:

1. Capture click origin (clientX, clientY).
2. Write `--ox` / `--oy` custom props on `:root`.
3. If `document.startViewTransition` is available (Chrome 111+, Edge, Safari 18+),
   use it + override `::view-transition-new(root)` with a custom clip-path morph.
4. Otherwise, fall back to a manual overlay element with `clip-path:
   circle(0 at ox oy) → circle(150% at ox oy)`, commit `data-mode` at the
   transition peak, then unmount.
5. `prefers-reduced-motion` → instant switch.

The destination paper hex values (`#F4EFE4` / `#16140F`) are the **only** raw
hex in the file — required because `:root[data-mode]` tokens don't cascade through
a foreign `data-mode` attribute on a descendant element.

### 5.5 Hotkey

Per CLAUDE.md §8 / locked decision #8, the hotkey is set on first run (in the
Ritual), not from Settings. The Settings surface exposes a "Hotkey" row with the
current combo + **Re-record** action (uses the same recording flow as the
Ritual — `keycaps` surface + presets).

### 5.6 Kill Switch

A **read-only status row** showing the kill-switch hotkey (default
`Cmd+Shift+Escape`). Cannot be disabled (CLAUDE.md §2.1 invariant #4).

A **Test kill switch** button fires a confirmation: "Press
Cmd+Shift+Escape to halt the agent. Test wires a tap to the same flow."

---

## 6. Edge Cases & Failure States

Every state below has a defined presentation. The default for unknown errors is
**honest degradation** — never a dead wall.

### 6.1 Daemon Unreachable

**Detected by:** any `ipc.X()` rejecting or `daemon.connected === false`.

**Per-screen treatment:**

| Screen | Behavior |
|---|---|
| Chat | Card `<div class="err-state" role="alert" aria-live="polite">` with italic Instrument Serif "We couldn't reach the daemon." + a hairline that draws left→right. The composer stays usable — `conversation.send` will retry. |
| Skills | Same italic err state + a "Try again" pill button. Cards disappear. |
| Hub | Same italic err + hairline. The 3D bookshelf renders empty. |
| Sync | Same italic err + "Try again" pill. The canvas shows "Discovering peers…" whisper. |
| Audit | Same italic err + "Try again" pill. The chain renders empty. |
| Channels | Telegram row stays connectable; the daemon call fails silently. Hydration error inline. |
| Delegation | The canvas renders (it's a static sub-agent list). Live panel polls gracefully — empty state shows "Nothing in flight." |
| Settings | The matrix shows a loading rule (`matrix-loading`) — a 360px thread that fills with pollen. The "No providers configured" empty state shows under Power. |

**Shell-level:** the `<DynamicIsland>` flips to `error` phase (red Pulse).
The titlebar stays alive (still navigable). The user can switch routes; the
error appears per-route, not globally.

### 6.2 No API Key, No Ollama

**Chat (the most-affected screen):**

- Empty state stays unchanged (hero + composer pinned).
- The model `<select>` shows whatever `ipc.providersList()` returned. If empty,
  the model selector is hidden entirely (`.model-select` only renders when
  `modelOptions.length > 0`).
- Send button is enabled; clicking it triggers
  `conversation.send(providerName, modelId, text)`. If the daemon rejects
  with "no provider", an italic `streamingError` block renders mid-list.

**Settings → Power:** the "Default model per provider" section renders the
italic empty state "No providers configured. Add an API key to begin."

The Power step in the Ritual shows the Local Ollama card with meta "not
detected — install Ollama, or pick another".

### 6.3 Permissions Denied (Accessibility / Screen Recording)

**First-run path:** the Ritual Permissions step shows a badge with
`data-status="denied"` (red border + danger tint). The user can:
1. Click **Open System Settings →** (deep link via `runtime.BrowserOpenURL`)
   to grant manually.
2. Continue anyway — `onboarding.completePermissions()` is called, allowing
   the user to skip. Computer use will be partial until granted.

**Settings → Permissions (read-only):** the `.perm` list shows the current
state. A pill button per row links to the System Settings deep link.

**Ritual-replay:** if permissions were skipped but later granted in System
Settings, the live badges flip on the next 2s poll cycle (no re-Ritual
needed). The user must `onboarding.completePermissions()` to proceed.

### 6.4 Kill Switch Pressed

`Shell.svelte:247–249`:

```svelte
{#if halt.state.halted}
  <KillSwitchOverlay reason={halt.state.reason ?? 'user requested'} onresume={handleResume} />
{/if}
```

`KillSwitchOverlay.svelte`:

- **Full-screen scrim** at 86% surface-ink with 8px backdrop-blur.
- A single **kill-card** centered: red Pulse `error` (12px), italic eyebrow
  `— Halted · kill switch engaged`, headline "Condura has stopped.".
- Body explains: "Every active stream was canceled. The agent is not running."
- Italic note: "Resuming mints a ticket you confirm from the CLI — the GUI
  never auto-restarts a halted agent. **Auto-recovery is the enemy.**"
- Single button: **Mint resume ticket**. Calls `halt.resume()` which
  generates a CLI-confirmed ticket.

**No auto-dismiss.** The modal stays until the user explicitly resumes.
This is invariant §2.1 #4.

### 6.5 Onboarding Incomplete (Ritual Resurfaces)

`Shell.svelte:81–102` polls:
1. `ipc.firstRunStatus()` → returns `{complete: boolean}`.
2. `ipc.onboardingIsComplete()` → returns `boolean`.
3. `localStorage.getItem('condura-ritual-seen')` → user-affordance flag.

If the first two pass and the third is unset, the Ritual still shows (so a
user who wiped localStorage but had finished server-side gets the cinematic
once more). If either server-side check fails, the Ritual mounts.

### 6.6 Streaming Interruption

When `conversation.streamingError` is set (mid-stream disconnect), the Chat
surface renders the italic err-state between the user message and the partial
agent bubble: "We couldn't reach the daemon. The thread stopped mid-sentence."

The user can:
- Click **Try again** to retry the last turn.
- Click **New chat** to clear and start over (the conversation store has a
  `clear()` action).
- The stream-bar stops animating; the **caret** stays frozen on the last
  character.

### 6.7 Consent Modal (Gatekeeper)

`ConsentModal.svelte` mounts globally at `Shell.svelte:246`. It only renders
when `consent.ticket !== null` (the store polls every 1.2s).

- **Scrim:** 32% ink + 6px blur; for **destructive** actions, the scrim
  shifts to `var(--surface-ink)` 50% (the rare ink surface).
- **Card:** central, 540px max. Eyebrow: "{blast_label} action · requires
  your consent". Headline: "Condura wants to act." Italic body: "Review what
  will happen before you allow." (Or for destructive: "This cannot be undone.
  Review exactly what will happen before you allow.")
- **Action summary** block: the articulated decision, actor line, nonce.
- **Armor rect:** a 1.5px synapse border that **draws in left→right** over
  1.4s (`armor` gesture), identifying the action as protected.
- **Allow / Deny** buttons. Esc → deny.
- **Countdown bar** at the bottom (timer / 300000 ms × width %). When the
  timer hits zero, the ticket expires; the daemon queues or denies based on
  policy.
- On Allow: a **wax seal** stamps (radial gradient, 108px circle, scale 0→
  1 with a `stamp` keyframe).

### 6.8 Voice Disabled / Mic Missing

In `Settings → Voice`:
- If `voiceProbe.mic_available === false`, a status line reads "Mic not
  granted · wake word 'hey condura' off per the daemon."
- The Wake Word toggle is independent and still works — it persists to
  localStorage. The sensitivity slider disables.
- `prefers-reduced-motion` users see voice as off by default (no celebration
  micro-animations).

### 6.9 Account / Hub Offline

`account.isSignedIn` is the gate. Signed-out users see:

- The Audit / Hub / etc. surfaces all function locally (no auth required for
  local data).
- Settings → Account shows "Not signed in. The agent works without an
  account — sign in to sync skills and hub bookmarks." with an inline sign-in
  link (opens a magic-link flow → `account.signInWithEmail`).
- The CommandPalette has no Account/Sign-in command in v0.1.0 (deferred to
  v0.2.0).

### 6.10 Empty States — Summary Table

| Screen | Empty Truth | Copy |
|---|---|---|
| Chat | no messages yet | "Your computer, alive." hero |
| Skills | no skills installed | "No skills yet. Run a complex task — Condura will save the procedure as a skill automatically." |
| Hub | curated library is empty (rare) | "The shelf is quiet. The Hub is empty. New skills land here as the community publishes them." |
| Hub | search no-match | "Nothing on this shelf matches." |
| Audit | chain is quiet | "Nothing has happened yet. Every action the agent takes will land here, HMAC-chained." |
| Audit | filtered to zero | "No events match. Loosen the filters, or look at a wider window." |
| Replay | no frames | "Once the agent acts, every decision lands here — screenshot, decision, outcome." |
| Sync | no peers | "Discovering peers on your LAN…" whisper |
| Sync | no pairs | (the canvas shows the green node alone, threads empty) |
| Channels | bot token not yet entered | defaults render with `state: 'off'` and Connect buttons |
| Delegation | no pending actions | "Nothing in flight. When a sub-agent asks the Gatekeeper to act, the ticket will land here." |
| Delegation | no sub-agents wired (rare) | Glyph + "No sub-agents wired. Install Claude Code, Codex, or Ollama to populate the constellation." |

---

## 7. State Inventory

This is the state the user can land in, mapped to the CSS classes / routes /
stores that render them. **If a state is in this table, it is reachable.**

| State | Visual signature | Trigger |
|---|---|---|
| **Booting** | white paper, no UI yet | `main.ts` theme script runs before paint |
| **Ritual (full-bleed)** | arrival awakening → 9-step sequence | `firstRunStatus.complete === false` |
| **Empty Shell** | NavRail + Chat empty state | First time the shell renders |
| **Chat · loading** | streaming caret + stream-bar pollen dash | `conversation.isStreaming` |
| **Chat · error** | italic err-state mid-list with hairline | `conversation.streamingError` |
| **Skills · loading** | INDEXING pulse | `skills.loading === true && skills.length === 0` |
| **Skills · empty** | "No skills yet" serif | 0 skills |
| **Skills · error** | "We couldn't reach the daemon." + retry | `loadError` |
| **Skills · detail** | right-side sheet with thread + Run/Improve | card click |
| **Hub · loading** | INDEXING THE SHELF pulse | `hub.loading && hub.results.length === 0` |
| **Hub · empty (curated)** | "The shelf is quiet." | 0 results, no query |
| **Hub · empty (search)** | "Nothing on this shelf matches." | 0 results, query non-empty |
| **Hub · error** | italic err + hairline | `hub.error` |
| **Hub · detail** | right-side sheet + Install button | spine click |
| **Sync · pending pairing** | centred card with pollen TTL ring + PIN input | `sync.pendingPin && sync.pendingPeerId` |
| **Sync · revoke** | centred popover | paired-node click |
| **Sync · empty** | "Discovering peers on your LAN…" whisper | 0 pairs, 0 pending |
| **Sync · error** | italic err + retry | `sync.error` |
| **Audit · loading** | READING THE CHAIN pulse | `audit.loading && audit.events.length === 0` |
| **Audit · empty (truly)** | "The chain is quiet." | 0 events, no filter |
| **Audit · empty (filtered)** | "No events match." | 0 events, filter applied |
| **Audit · detail** | right-side detail pane | node click |
| **Replay · loading** | LOADING FRAMES pulse | `replay.loading && count === 0` |
| **Replay · empty** | "Nothing to replay yet." | 0 frames |
| **Replay · integrity broken** | "Chain broken at row X: reason" line | `!integrity.valid` |
| **Replay · frame** | before/after screenshots + receipt + scrubber | frame selected |
| **Channels · hydrating** | PROBING REACH pulse | `hydrating` |
| **Channels · error** | italic err (defaults still render) | `hydrateError` |
| **Channel · connected** | 5-dot breathing signal | `state === 'connected'` |
| **Channel · degraded** | 3-dot static signal in warn | `state === 'degraded'` |
| **Channel · soon** | dim row + `v0.2.0` pill | `state === 'soon'` |
| **Delegation · live** | canvas with pulses + Live panel rows | canvas always live; live panel polls every 5s |
| **Delegation · empty live** | "Nothing in flight." | 0 pending actions |
| **Delegation · empty canvas** | Glyph + "No sub-agents wired." | 0 sub-agents wired |
| **Consent · read** | scrim + card with armor rect drawing | `consent.ticket !== null` |
| **Consent · destructive** | ink scrim + ink card | `consent.ticket.action_kind === 'destructive'` |
| **Consent · countdown low** | amber countdown bar | `consent.timer / 300000 < 0.2` |
| **Consent · stamped** | wax seal rendered over the card | user clicked Allow |
| **Kill switch** | full-screen overlay + resume button | `halt.state.halted === true` |
| **Command palette · open** | blurred scrim + paper panel + sliding pollen highlight | `paletteOpen === true` |
| **Command palette · empty** | "Nothing matches this search." | `filtered.length === 0` |
| **Quick prompt · open** | top-centre floating card with thread-draw top edge | `quickOpen === true` |
| **Quick prompt · streaming** | Pulse `acting`, sync-line stream-bar | `conversation.isStreaming && open === true` |
| **Settings · dirty** | sticky pollen save bar in | `dirty === true` |
| **Settings · saving** | "SAVING…" + Pulse | `settings.saving === true` |
| **Settings · saved** | "Saved" + green Pulse, bar fades 400ms | `savedFlash === true` |
| **Settings · save failed** | "Save failed" + raw error | `settings.lastSaveError` |
| **Settings · matrix loading** | "—" + 360px pollen thread rule | `configEmpty === true` |

---

**This document is the architecture. The code is the implementation. They
agree.** When they diverge, the divergence is the spec-bug — fix the doc, then
fix the code, in one commit.
