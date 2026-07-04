# SCREEN — Consent Modal · `ConsentModal.svelte` (global, mounts at `Shell.svelte:246`)

> **Status:** Phase-2 architecture spec. The current `ConsentModal.svelte`
> (314 lines) already ships the **armor rect** entrance, **wax-seal stamp** on
> approval, ink-variant surface for destructive actions, and the countdown bar.
> What it does **not** ship: a **"Trust this app" toggle** (adds to the autonomy
> matrix), the **consent-provider explanation footer** (the Gatekeeper is
> deterministic, not the model), the **timeout policy disclosure**, the tooltips,
> and the `prefers-reduced-motion` discipline that MOAT §2.3 demands. This spec
> keeps the existing signature gestures and ships what is missing.
>
> **Contract:** `MOAT.md` (premium bar), `APPFLOW.md §6.7` (current Consent row),
> `CLAUDE.md §2` (the Survival Rule — destructive actions require a real human
> at the keyboard), `CLAUDE.md §10.2` (the Gatekeeper is deterministic, not an
> LLM), `icons.ts` + `condura.css` (the actual tokens available). `DIRECTION.md`
> and `DESIGNLANG.md` were named in the brief but are not on disk at this
> commit, so this spec leans on `MOAT.md` for the design grammar and on the
> actual token table from `condura.css` for everything visual.

---

## Table of Contents

1. [What this surface is and isn't](#1-what-this-surface-is-and-isnt)
2. [Inheritance — what the spec inherits from MOAT/APPFLOW/CLAUDE](#2-inheritance)
3. [Layout & Content](#3-layout--content)
4. [State Matrix — six states, full copy](#4-state-matrix)
5. [Motion Choreography](#5-motion-choreography)
6. [Keyboard](#6-keyboard)
7. [Components Used — boundaries and props](#7-components-used)
8. [Data Fetched — IPC contract](#8-data-fetched)
9. [Design Decisions — which MOAT rules this passes](#9-design-decisions)
10. [What this spec adds to / deletes from the current `ConsentModal.svelte`](#10-what-this-spec-adds-to--deletes-from-the-current-consentmodalsvelte)

---

## 1. What this surface is and isn't

**Is.** The single safety surface every gated WRITE / NETWORK / DESTRUCTIVE
action passes through before the agent acts. Per `CLAUDE.md §2.1` invariants
#2 ("The Gatekeeper is the only path to physical action") and #3 ("Destructive
actions require a real human at the keyboard"), this modal — together with the
daemon-side Gatekeeper — is the only software chokepoint between a model's
output and a real click, keystroke, or shell exec on the user's machine. The
modal **blocks until clicked**; it is not an in-feed warning, not a toast, not
an explainer — it is a *halt*. The component mounts globally at
`Shell.svelte:246`, polls `gatekeeper.pending_consent` every 1.2s, and only
renders when a ticket is non-null (`consent.ticket !== null`).

**Is not.** Not a confirmation dialog styled to look pretty (`MOAT §4 #6` —
no celebration). Not a permission system on its own (the Gatekeeper's
deterministic policy engine is; this is its face). Not an LLM-judged
risk prompt (`CLAUDE.md §5.7` — the Strategist and Gatekeeper must be
separate systems; this modal is the GUI face of the **Gatekeeper**, not the
Strategist). Not an optional flow (every non-READ action gates here; the
component does not render at all when there is no ticket, so there is no
"how do I make this modal re-appear" question).

**Mental model the user carries away:** *This is the one screen the agent
cannot dismiss, defer, or "auto-confirm." When this is on screen, the
computer is waiting for you.*

**One sentence for the eyebrow:** *— {blast-label} action · requires your
consent.* (`Read` / `Write` / `Network` / `Destructive`.)

**One sentence for the title:** *Condura wants to act.*

**One sentence for the body:** *Review what will happen before you allow.*

**Why this is the signature surface, not just a dialog:** per `APPFLOW.md §1`
intent **I5** ("the 7 invariants are visible") and `APPFLOW.md §6.7`, this
modal *is* the visible survival rule. The armor rect drawing in around the
action summary, the wax seal on approval, and the paper-then-ink surface
escalation for destructive actions are the only gestures in Condura that
**carry weight without utility** — they exist specifically to make Allow /
Deny feel *consequential*. The user must come away from this modal
remembering that it is a real gate.

---

## 2. Inheritance

The spec assumes the following are already in place; this spec does
**not** redefine them.

| From | What this spec uses as-is |
|---|---|
| `MOAT.md §2.1` | Focus halos are rounded — `--shadow-focus` (the 4px pollen halo) or, for ≥8px radii elements, a 5px halo + 2px synapse ring. The Allow / Deny / Trust toggle halos all follow this rule. |
| `MOAT.md §2.2` | Press states carry weight (`translateY(0.5px) brightness(0.95) saturate(1.05)`), not just shrinkage. Allow and Deny inherit via `<Button>`'s existing `.btn-primary` / `.btn-ghost` press rules. |
| `MOAT.md §2.3` | `prefers-reduced-motion: reduce` is respected via one global rule in `condura.css`. The component declares no per-keyframe media-query blocks. Armor-rect draw, stamp gesture, and shake gesture are suppressed; the rest (countdown, modal fade) are unchanged. |
| `MOAT.md §2.9` | Tooltips via `<Tooltip label>` — no `title=` attributes that misbehave. The "What does Gatekeeper mean?" affordance uses `<Tooltip>`. |
| `MOAT.md §3` | The Thread is the visual grammar for "moment of completion." The wax seal IS the thread-on-allow (drawn in 320ms, equivalent semantic). The armor rect IS the protection gesture. Both are variants of the Thread idiom. |
| `MOAT.md §4 #1` | No gradient text. The modal is plain paper (or ink for destructive). |
| `MOAT.md §4 #2` | All icons go through `<Glyph name="…" />`. The shield (gate icon), check (allow), close (deny) all come from `icons.ts`. |
| `MOAT.md §4 #3` | No glassmorphism unless elevation is earned. The scrim uses `backdrop-filter: blur(6px)` — this is the rare exception (`APPFLOW.md §6.7` confirms: "32% ink + 6px blur"), because the modal halts the page and the blur is the "the world behind you has gone soft" cue. The card itself does not blur. |
| `MOAT.md §4 #4` | Blast-radius colors: READ `--ok` (green), WRITE `--pollen`, NETWORK `--danger` (red), DESTRUCTIVE `var(--surface-ink)` (the deep-ink surface per `APPFLOW.md §6.7`). No purple, no teal. |
| `MOAT.md §4 #5` | No "Welcome to the future" copy. Title: "Condura wants to act." Sub: "Review what will happen before you allow." Restraint is the message. |
| `MOAT.md §4 #6` | No "Awesome!", no "Allowed!" celebration. The wax seal is a stamp, not a confetti pop. The "Allowed" wordmark inside the seal is austere. |
| `MOAT.md §4 #7` | No spinner. The countdown bar is the time indicator (a linear thing), not a `<Pulse>` or `<Spinner>`. |
| `MOAT.md §4 #8` | Focus halos are rounded, per §2.1. The Allow pill (radius `--r-pill`) gets the bare synapse ring + 4px pollen halo (the global shadow already does this for ≥999px). |
| `MOAT.md §4 #9` | One elevation token per surface. The card carries `--shadow-float` at rest — no second shadow on hover. |
| `MOAT.md §4 #10` | Every animation answers: *what is this communicating?* Armor-rect draw = "this action is being protected." Stamp gesture = "your decision is now committed." Shake gesture = "this didn't go through, here's the form again." Modal fade = "the world is back." |
| `APPFLOW.md §6.7` | Current implementation contract — armor rect, wax seal, ink surface for destructive, countdown bar, Esc-to-deny. The spec keeps all of these. |
| `APPFLOW.md §7` | Consent row in the state inventory. The spec adds three rows: timeout-countdown (timer < 60s), trust-toggled, sealed. |
| `CLAUDE.md §2` invariant #3 | "Destructive actions require a real human at the keyboard. Native modal dialog. Blocks until clicked." This spec is the GUI implementation of that invariant — the modal **does not auto-dismiss**, and Esc maps to Deny, not to close-and-defer. |
| `CLAUDE.md §10.2` | "The Gatekeeper is deterministic. Not an LLM. Cannot be prompt-injected, cannot hallucinate." The footer copy in §3.1 below restates this in user-facing language. |
| `CLAUDE.md §27` | The Autonomy Matrix. The "Trust this app" toggle is wired to write the appropriate `apps.<bundle-id>: warn → autonomous` cell of the user's `~/.synaptic/autonomy.yaml`. The matrix already has 12 example app entries. |

---

## 3. Layout & Content

### 3.1 Component-level structure

The modal mounts at `Shell.svelte:246` (a sibling of `<KillSwitchOverlay>`,
`<DynamicIsland>`, `<CommandPalette>`, and `<QuickPromptOverlay>`). It is
positioned `position: fixed; inset: 0` so it covers the entire window on top
of every route, modal, and overlay. The semantic structure inside is:

```
┌──────────────── Scrim (full viewport, ink + 6px blur) ───────────────┐
│                                                                      │
│            ┌──── Consent card (520px max, paper) ─────┐             │
│            │                                            │            │
│            │  ─ WRITE ACTION · REQUIRES YOUR CONSENT    │  eyebrow    │
│            │                                            │            │
│            │  Condura wants to act.                     │  title      │
│            │                                            │            │
│            │  Review what will happen before you allow. │  sub        │
│            │                                            │            │
│            │  ┌──── Action summary card ──────────────┐ │            │
│            │  │  — Action summary                     │ │ eyebrow    │
│            │  │                                       │ │            │
│            │  │  Send an email in Gmail.              │ │  body      │
│            │  │                                       │ │  display   │
│            │  │  Requested by claude · nonce 1f4a7b…  │ │  meta      │
│            │  │  ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒ armor rect   │ │  stroke    │
│            │  └───────────────────────────────────────┘ │            │
│            │                                            │            │
│            │  ☐ Trust this app                          │  trust row  │
│            │    Don't ask again for {bundle-id}. Add to  │  line 2    │
│            │    autonomy matrix.                         │            │
│            │                                            │            │
│            │  ┌─ footer hairline ────────────────────┐  │            │
│            │  │  ⓘ The Gatekeeper is a local rule     │  │            │
│            │  │    engine. Your model did not decide  │  │            │
│            │  │    to ask.                            │  │            │
│            │  └──────────────────────────────────────┘  │            │
│            │                                            │            │
│            │           Esc to deny      [Deny]  [Allow] │  foot      │
│            │                                            │            │
│            │  ▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░ countdown bar       │  timer      │
│            │                                            │            │
│            └────────────────────────────────────────────┘            │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

The card is `max-width: 520px` (per the brief). The current implementation
uses `width: min(540px, 92vw)`; this spec tightens to `min(520px, 92vw)` to
match the brief. The card sits inside `padding: var(--space-8)` (= 32px),
border-radius `var(--r-lg)` (= 24px), `border: 1px solid var(--hair-strong)`,
`background: var(--surface)`. Destructive swaps the card to `var(--surface-ink)`
with `color: var(--paper)` per `APPFLOW.md §6.7`.

### 3.2 Region A — Header (eyebrow + title + sub)

| Element | What it is | Copy exactly |
|---|---|---|
| **Eyebrow** | `font-mono 11px uppercase letter-spacing 0.16em`, `color: var(--content-mute)`. For destructive, color shifts to `color-mix(in oklab, var(--paper) 55%, transparent)`. | Format: `{blastLabel} action · requires your consent`. Examples: <br/>- `Read action · requires your consent` <br/>- `Write action · requires your consent` <br/>- `Network action · requires your consent` <br/>- `Destructive action · requires your consent` |
| **Title** | `font-display 30px line-height 1.1 letter-spacing -0.03em`, `color: var(--content)`. For destructive, `color: var(--paper)`. | `Condura wants to act.` |
| **Sub** | `font-sans 14px line-height 1.5`, `color: var(--content-mute)`. For destructive, `color-mix(in oklab, var(--paper) 55%, transparent)`. | For non-destructive: `Review what will happen before you allow.` <br/>For destructive: `This cannot be undone. Review exactly what will happen before you allow.` |

A `Thread` (1px, `--hair-strong`, horizontal) sits between sub and the
action-summary card — the "we are entering the protected zone" divider.

### 3.3 Region B — Action summary card (the payload)

This is the card the armor rect arms. It contains, top-to-bottom:

| Row | Component | What |
|---|---|---|
| 1 | Eyebrow | `font-mono 10px uppercase letter-spacing 0.14em color: var(--synapse)`, all caps `Action summary`. For destructive, color shifts to `color-mix(in oklab, var(--synapse-light) 85%, transparent)`. |
| 2 | Body | `font-display 22px line-height 1.2 letter-spacing -0.02em`. The **articulated decision** — what the agent thinks it's about to do, in plain English. Sourced from `ticket.detail`. Example: `Send an email in Gmail.` |
| 3 | Meta | `font-mono 11px color: var(--content-faint)`. Format: `Requested by {actor} · nonce {first-8-chars}…`. `actor` is `ticket.actor` (e.g. `claude`, `ollama`, or the user-facing model name). |
| 4 | Armor rect | A `<svg>` overlay (not a child) sitting on top of the action-summary card via `position: absolute; inset: -4px;`. A 1.5px `--synapse-glow` rectangle with rounded corners (`rx=12`, 96% width/height), `pathLength="1"`, `stroke-dashoffset: 1 → 0` over `1.4s var(--ease) 0.15s` (the signature **armor gesture**). Draws on every new ticket arrival. |

For destructive, the action-summary card itself swaps to:
- `border-color: color-mix(in oklab, var(--paper) 12%, transparent);`
- `background: color-mix(in oklab, var(--paper) 4%, transparent);`

The armor rect's stroke color is **always `--synapse-glow`** regardless of
blast radius — the protection gesture is one color, not four. Different
blast-radius classes are conveyed by the eyebrow color and the trust row's
wording, not by a colored armor.

**Why the armor rect, not a thicker border or a separate "PROTECTED" label:**
the rect draws in left-to-right over 1.4s, the same cadence as every Thread
in the app. The user has already learned that "left-to-right drawing line =
something arriving" — the armor rect is the same idiom applied to a closed
shape, communicating "this surface is being wrapped in protection." A static
border would not convey the moment of arming.

### 3.4 Region C — Trust-this-app row

A new addition over the current `ConsentModal.svelte`. Per `CLAUDE.md §27`
(the Autonomy Matrix), users can pre-approve apps they've already decided are
safe. The Trust toggle records that decision at the appropriate cell of
`~/.synaptic/autonomy.yaml`, with the user's finger on the Allow button.

| Row | Element | What |
|---|---|---|
| 1 | Toggle row | A `<Switch />` (off by default) prefaced by `Trust this app`. Sub line (one sentence below, `font-sans 13 --content-mute`): |
| 2 | Sub | For WRITE / NETWORK: `Don't ask again for {ticket.actor or app bundle-id}. Add to autonomy matrix.` <br/>For READ: toggle **hidden entirely** — READ actions auto-allow per `CLAUDE.md §10.1` (`match: { class: READ } decide: allow`), so the Trust toggle is meaningless for READ. <br/>For DESTRUCTIVE: `Trust this app for these actions in the future. (Destructive actions still require physical presence.)` The destructive caveat is required — invariant #3 still holds even with the app trusted. |

When the user flips the Trust toggle on, the Allow button's label does
**not** change — the decision is composed and resolved in a single IPC call
(`gatekeeper.approve({nonce, trustApp: true})`). If the IPC call fails,
the local toggle resets to off and an inline error appears in the sub line:
`Could not record trust preference. {error message}`.

**Why the Trust row sits between the action summary and the footer (not in
the footer):** the trust decision is *about the action, not about the
modal*. Placing it adjacent to the articulated action lets the user reason
about both at once. Placing it in the footer would conflate it with the
Allow/Deny decision.

### 3.5 Region D — Consent-provider footer (the hairline + explanation)

A new addition. A 1px hairline (`var(--hair)`) above a one-line italic
explanation. This is the line that answers "is the LLM asking me, or is
something else asking me?"

| Element | What |
|---|---|
| Hairline | `<Thread orientation="h" />`, 1px, `var(--hair)`, full card width. |
| Body | `font-display italic 14px line-height 1.5 color: var(--content-soft)`. |
| Info icon | `<Glyph name="info" size={14} />` in `var(--content-faint)`, hover triggers `<Tooltip label="What does Gatekeeper mean?">`. |
| Copy (exact) | `ⓘ The Gatekeeper is a local rules engine. Your model did not decide to ask — the policy did.` (one sentence, per `MOAT §4 #5`). |
| Tooltip body (on icon hover) | `The Gatekeeper is a deterministic policy engine on your machine. It classifies every action by blast radius (Read / Write / Network / Destructive) and runs the rules in your autonomy matrix. Models cannot override it.` (~ 240 chars — the tooltip is a one-breath explanation, not an essay.) |

For destructive: italic copy color shifts to
`color-mix(in oklab, var(--paper) 65%, transparent)`; info icon shifts to
`color-mix(in oklab, var(--paper) 40%, transparent)`; tooltip body uses the
ink-paper surface.

**Why this lives here and not on the Settings → Autonomy matrix page:** the
user is in the modal *right now*, about to click Allow. Showing them why
they're being asked is the moment. The Settings page is one click away;
they can audit the matrix there. The footer is *just enough* explanation
to settle "is this safe?" in five seconds.

### 3.6 Region E — Footer (buttons)

The current `ConsentModal.svelte:114-120` already implements this section.
The spec keeps the implementation, naming the responsibilities:

| Element | What |
|---|---|
| Hint (left, mono uppercase 11, faint) | `Esc to deny` — always present, never changes. Reinforces the safe-default mapping (Esc = deny). |
| **Deny** button | `<Button variant="ghost" class="deny">` with the destructive variant override (`variant={destructive ? 'danger' : 'ghost'}` in the current implementation). Width: auto; rightmost thumb of the row. |
| **Allow** button | `<Button variant="primary" magnetic class="allow">` — the primary pollen pill. Label: `Allow` (changes to `Allowed` while the wax seal stamps, then back to `Allow` once the modal closes). Magnetic is on — the existing `magnetic` directive pulls the cursor slightly toward the button. **Default focus on mount** (the Allow button is the initial `focus()` target after the 200ms entrance delay — see §6.1). |

The buttons together live in `.consent-foot` — a flex row with `gap: var(--space-3)`
and `margin-right: auto` on the hint so it floats left.

### 3.7 Region F — Countdown bar

A 2px horizontal bar, full card width, beneath the buttons. The bar is
filled with `--warn` (amber) and drains from full to empty over 5 minutes
(`CONSENT_TIMEOUT_MS = 300_000`). When the bar drops below 20% the warn
color shifts subtly toward danger (the bar reads as "the queue is
imminent" — not "the system is broken"). Below 5% the hair-faint
underneath pulses briefly to draw the eye (the "decision time" cue,
honest).

**Why a bar at all, not a number:** the user does not need to know "you
have 3:47 left." They need to know "the system is waiting on you, and
will stop waiting if you walk away." The bar's *shape* — a draining
line — reads as that. The exact number is one click away in `audit`.
Per `MOAT §4 #5`: no countdown numerals.

**What happens when the timer hits zero:** the daemon's on-timeout policy
takes over. The spec's contract for the GUI:

- When `consent.timer` is in [0, 60_000ms]: the bar pulses (a quiet
  600ms amber→faint→amber loop, per `MOAT §4 #10` — *what is this
  communicating?* "the queue is here soon, decide before you walk
  away").
- When the daemon resolves the ticket (which it does on timeout — queue
  or deny, per policy in `~/.synaptic/policy.yaml`): the modal's `consent.
  ticket` becomes `null`, the modal fades, and a notification is pushed
  (`Action queued — Condura will resume when you're back.` or
  `Action denied — see autonomy matrix for the policy.`). The notification
  is shown via the existing `notifications.push(...)` API.
- The countdown never *blocks* clicking Allow / Deny. The bar is a soft
  indicator, not a hard timeout from the modal's perspective.

### 3.8 Region G — Wax seal (post-approval, transient)

A transient `<div class="seal">` that fades in over the card after the
user clicks Allow. The seal is a 108×108 radial-gradient circle centered on
the card, paper-on-synapse, with two lines of text inside:

- **Word (display 20):** `Allowed`
- **Sub (mono 8, uppercase 0.18em letter-spacing):** `by you · now`

The seal animates via the **`stamp` keyframe**: `transform: translate(-50%,
-50%) scale(0)` → `scale(1.04)` → `scale(0.96)` over `var(--dur-slow)` (520ms)
ease-out. The first half of the animation is the "coming down" (0 → 1.04);
the second half is the "settling" (1.04 → 0.96). The seal holds at
`scale(0.96)` for 320ms, then the whole modal fades out (opacity 1 → 0 over
240ms `--ease`). Total elapsed: ~1080ms from click to modal-gone.

The seal is `position: absolute; left: 50%; top: 50%;` so it scales from
the center of the card. It's `pointer-events: none` so the user can't
click "through" it. It's `z-index: var(--z-tooltip)` (above everything in
the card but below any future nested overlay).

### 3.9 Ink variant (destructive)

When `ticket.action_kind === 'destructive'`, the entire card swaps to the
deep-ink surface per `APPFLOW.md §6.7`:

- `background: var(--surface-ink)` (the rare ink surface — only used for
  destructive consent + the kill-switch overlay)
- `border-color: color-mix(in oklab, var(--paper) 10%, transparent)`
- `color: var(--paper)` everywhere

The Deny button changes variant to `danger` (per the current
implementation); Allow stays `primary` (pollen). The wax seal's gradient
flips to `radial-gradient(circle at 35% 30%, var(--synapse-glow),
var(--synapse-deep) 70%)` — the existing palette; reads correctly on the
ink background.

The **scrim itself** changes: from
`color-mix(in oklab, var(--ink) 32%, transparent)` to
`color-mix(in oklab, var(--surface-ink) 50%, transparent)`. The world
behind the modal goes deeper-ink; the user feels the weight of the
destructive class. See `APPFLOW.md §6.7`.

---

## 4. State Matrix — six states, full copy

The modal has six reachable states. **Mutually exclusive at any instant.**
The same component file owns the rendering for all six; only one block is
visible at a time.

| # | State | What you see | Trigger / data condition |
|---|---|---|---|
| **S1** | **Closed** | Not rendered. The DOM contains no `.scrim` element. The body remains interactive; the user can navigate, type, scroll. The polling loop continues (every 1.2s). | `consent.ticket === null`. |
| **S2** | **Open (waiting on user)** | The full modal per §3. The Allow button has focus (after the 200ms entrance delay — see §6.1). The armor rect has finished drawing. The countdown bar shows full. | `consent.ticket !== null && consent.timer > 60_000`. |
| **S3** | **Open (imminent timeout)** | Same modal as S2, but the countdown bar pulses at amber→faint cadence (`pulse-urgent`, 600ms, infinite). The user's eye is drawn to the bar. | `consent.ticket !== null && consent.timer <= 60_000 && consent.timer > 0`. |
| **S4** | **Approved** | The wax seal stamps on (per §3.8) over the card. The Allow button's label briefly changes to `Allowed` (matching the seal wordmark — the seal IS the wordmark of approval). After the seal settles, the modal fades out (240ms). Then `consent.ticket` becomes null and the modal unmounts to S1. | User clicks Allow (or Enter on focused Allow). `consent.approve(nonce)` resolves. |
| **S5** | **Denied** | The shake gesture runs (§5.4) over 200ms. After the shake settles, the modal fades out (240ms). Then `consent.deny(nonce)` resolves; the ticket is gone. | User clicks Deny, presses Esc, or Shift+Tab + Enter on Deny (focus trap ensures this is reachable). |
| **S6** | **Timeout-resolved** | The modal disappears (fade 240ms), `consent.ticket === null`. A toast notification is pushed via `notifications.push(...)` informing the user what the daemon decided (queue or deny, per the active autonomy policy). | `consent.timer === 0 && ticket !== null`. The daemon's policy decides queue/deny; the GUI just records the outcome. |

### 4.1 The Closed state (S1) — exact behavior

The component is mounted (it lives in `Shell.svelte:246`) but renders nothing.
Polling continues in the background:
- `setInterval(() => void poll(), 1_200)` is started in `onMount`.
- `poll()` calls `ipc.gatekeeperPendingConsent()` which returns `{tickets: ConsentTicket[]}`.
- If `tickets.length === 0`, `consent.ticket = null`.
- If `tickets.length > 0`, the first ticket is treated as the current ticket;
  if its `nonce !== consent.ticket?.nonce`, the local state is updated and
  the countdown reset.

### 4.2 The Open state (S2) — exact behavior

Per §3 in full. The Allow button has `focus()` applied after a 200ms
`setTimeout` from `onMount` of the scrim, so the focus halo doesn't compete
with the entrance fade. The armor rect's `stroke-dashoffset` animates from
1 → 0 over 1.4s `--ease` with a 0.15s delay (so it begins after the scrim
has settled). The countdown bar shows `width: 100%` and stays there.

### 4.3 The Imminent-timeout state (S3) — exact behavior

Exactly S2 + the bar's `pulse-urgent` keyframe. The pulse communicates
"decide soon" without saying "you're being timed out." When the timer
hits zero, `consent.resetCountdown()`'s internal interval clears and
`consent.ticket = null` — the modal handles the transition to S6
automatically (no extra logic in this component).

### 4.4 The Approved state (S4) — exact behavior

The seal handle. When the user clicks Allow:

1. **`sealed = true`** is set immediately (Svelte 5 `$state`), which
   triggers the `.seal` element to mount with `class="seal show"`.
2. The seal runs the `stamp` keyframe over 520ms — scale 0 → 1.04 →
   0.96, paused at 0.96.
3. Simultaneously, `consent.approve(nonce)` is called (the IPC roundtrip
   happens in parallel with the seal animation — the user perceives
   approval as instantaneous).
4. After the seal holds (520ms seal + 320ms hold = 840ms), the modal
   begins fade-out (`scrim` opacity 1 → 0 over 240ms `--ease`).
5. When the fade completes, `consent.ticket = null` (already done by
   `consent.approve`'s `finally` block), and Svelte 5 reactivity
   removes the `.scrim` from the DOM.

If the IPC rejects, `notifications.push({kind: 'error', title: 'Could not
allow action', message: '...'})` is called and the seal does NOT stamp —
the modal stays in S2 and an inline error appears in the trust row's sub
line (the user can retry). The countdown bar continues to tick.

### 4.5 The Denied state (S5) — exact behavior

The shake handle. When the user clicks Deny (or presses Esc):

1. The shake gesture runs (`scale 1 → 0.96 → 1.02 → 1` over 200ms
   `--ease`).
2. Simultaneously, `consent.deny(nonce)` is called.
3. After the shake settles (200ms), the modal begins fade-out (240ms).
4. `consent.ticket = null` from `consent.deny`'s `finally`.

If the IPC rejects, the same inline error appears in the trust row's sub
line; the modal stays in S2.

### 4.6 The Timeout-resolved state (S6) — exact behavior

When `consent.timer` hits zero, the store internally:
- clears its countdown interval,
- sets `consent.ticket = null` (so the modal unmounts),
- the daemon's on-timeout policy decides whether to queue (NETWORK
  default) or deny (DESTRUCTIVE default).

The GUI pushes one of two notifications (the store handles this; the
modal does not see it):
- `notifications.push({kind: 'info', title: 'Action queued', message: 'Condura will resume when you return.'})`
- `notifications.push({kind: 'info', title: 'Action denied', message: 'See autonomy matrix for the policy.'})`

The user can read the queued/denied action in `#/audit` (the audit log
captures the timeout event with reason `timeout` and the policy
branch).

### 4.7 Reduced-motion note for all six states

`prefers-reduced-motion: reduce` (handled by the global rule per `MOAT §2.3`):
- The scrim fade-in becomes instant.
- The armor rect does not draw — it appears already-drawn at full opacity.
- The wax seal does not stamp — it appears already at scale 1, holds for
  320ms, then the modal fades.
- The shake gesture on deny is replaced by a fade.
- The countdown bar's `pulse-urgent` becomes static (the bar's color shift
  to danger is preserved as a flat change).

Per `MOAT §4 #10`, the *meaning* is preserved (the user knows approval
happened, denial happened, the queue is imminent); the choreography is
abbreviated.

---

## 5. Motion Choreography

The modal's motion answers **what is this communicating?** for each
gesture. None loop on idle.

### 5.1 Entrance — backdrop fades in, card scales in, armor draws

Trigger: a new ticket arrives (`consent.ticket.nonce` changes), the scrim
mounts.

**Stage A — backdrop fade (200ms).**
```css
.scrim {
  animation: blur-in var(--dur) var(--ease);  /* var(--dur) = 200ms */
}
@keyframes blur-in {
  from { opacity: 0; }
  to   { opacity: 1; }
}
```
The backdrop-filter blur (6px) is applied immediately at the start of
the keyframe; only the opacity animates. (Animating `backdrop-filter`
itself is a perf hazard on low-end machines.)

**Stage B — card scale-in (280ms, starts at 80ms).**
```css
.consent-card {
  animation: consent-in 280ms var(--ease) 80ms backwards;
}
@keyframes consent-in {
  from { opacity: 0; transform: scale(0.96); }
  to   { opacity: 1; transform: scale(1); }
}
```
The card scales from 0.96 → 1 (a slight zoom-in feel — the inverse of
the stamp's "settle down"). `backwards` fill mode means the card is
already at scale 0.96 before the animation begins; this prevents a
"pop" at t=0.

**Stage C — armor-rect draw (1400ms, starts at 150ms).**
```css
.armor-rect rect {
  stroke-dashoffset: 1;
  transition: stroke-dashoffset 1.4s var(--ease) 0.15s;
}
```
The 0.15s delay puts the draw just after the card has settled, so the
armor arrives *after* the surface, not at the same moment. The user
reads: "the card arrived" → "and now it's being protected."

**Stage D — focus on Allow (200ms delay).**
```ts
onMount(() => {
  setTimeout(() => {
    allowButton?.focus();
  }, 200);
});
```
The focus halo appears *after* the entrance choreography so it doesn't
fight for attention with the scale-in. Per `MOAT §2.7` (tactile
consistency), Allow inherits the global pill-press treatment.

**Reduced-motion:** Stage A instant; Stage B instant at scale 1 with no
opacity change; Stage C skipped (rect appears already-drawn at full
opacity); Stage D still happens but the focus halo is itself a static
visual.

### 5.2 Approve — wax-seal stamp

Trigger: user clicks Allow or presses Enter on focused Allow.

The current `.seal` element already implements this; the spec owns the
**exact motion vocabulary**:

```css
.seal {
  transform: translate(-50%, -50%) scale(0);
  opacity: 0;
}
.seal.show {
  animation: stamp var(--dur-slow) var(--ease) forwards;  /* 520ms */
}
@keyframes stamp {
  0%   { transform: translate(-50%, -50%) scale(0);    opacity: 0; }
  60%  { transform: translate(-50%, -50%) scale(1.04); opacity: 1; }
  80%  { transform: translate(-50%, -50%) scale(0.96); }
  100% { transform: translate(-50%, -50%) scale(0.96); opacity: 1; }
}
```

- **Frames 0→60%: the seal arrives** (0 → 1.04 = the "pressing in"
  gesture, 312ms).
- **Frames 60→80%: the seal settles** (1.04 → 0.96, 104ms — the
  "bounce-back" that makes it feel physical).
- **Frames 80→100%: the seal holds** (0.96, 104ms).

After the stamp keyframe completes (520ms), the seal holds for an
additional 320ms (`setTimeout`), then the modal begins fade-out
(`.scrim` opacity 1 → 0, 240ms `--ease`). Total elapsed time from
click to modal-gone: 520 + 320 + 240 = **1080ms.**

**Reduced-motion:** the seal appears at scale 1 (no animation), holds for
320ms, then the modal fades. The "this was approved" communication is
preserved by the wordmark and color alone.

### 5.3 Deny — shake gesture

Trigger: user clicks Deny, presses Esc, or Shift+Tab + Enter on Deny.

```css
.consent-card.deny-shake {
  animation: deny-shake 200ms var(--ease);
}
@keyframes deny-shake {
  0%   { transform: scale(1); }
  35%  { transform: scale(0.96); }
  65%  { transform: scale(1.02); }
  100% { transform: scale(1); }
}
```

The shake is a 200ms three-keyframe wobble (1 → 0.96 → 1.02 → 1) —
short, not jittery. The new parameter on the card is the
`.deny-shake` class, applied for one tick of the animation. After the
shake settles, the modal fades out over 240ms (same as approve, but
without the seal). Total elapsed time from click to modal-gone: 200 +
240 = **440ms.**

**Reduced-motion:** the shake is replaced by an instant fade.

### 5.4 Imminent timeout — pulse-urgent

Trigger: `consent.timer <= 60_000ms && consent.timer > 0`.

```css
.countdown-fill.urgent {
  animation: pulse-urgent 600ms var(--ease) infinite alternate;
}
@keyframes pulse-urgent {
  from { opacity: 0.85; }
  to   { opacity: 0.55; }
}
```

The bar's alpha oscillates 0.85 ↔ 0.55, alternating, over 1.2s per cycle.
This communicates "decide soon" without resorting to numerals. The
underlying `--warn` color is unchanged — only the alpha oscillates.

**Reduced-motion:** the bar's color shifts to flat `--danger` once
the timer hits 60s (no pulse).

### 5.5 Reduced-motion: one global rule

`condura.css` has the global `@media (prefers-reduced-motion: reduce)`
block (per `MOAT §2.3`). This component declares no media-query blocks
of its own. The global rule covers:
- Armor-rect draw → instant appearance at full opacity.
- Stamp gesture → seal appears at scale 1.
- Shake gesture → fade.
- Pulse-urgent → static color.
- Card scale-in → instant at scale 1.
- Scrim backdrop fade → instant at opacity 1.

What is **not** reduced (because it carries meaning, not motion):
- The countdown bar's *linear* width change (this is information, not
  flourish).
- The wax seal's wordmark and color (the user needs to read "Allowed").
- The Allow/Deny button's standard `.tactile` press (this is global
  tactile feedback per `MOAT §2.2`).

---

## 6. Keyboard

The modal owns a complete focus trap when open. The trap implements:

### 6.1 Initial focus and focus trap

On the moment the scrim mounts (the `$effect` that watches `ticket?.
nonce`), the component schedules `setTimeout(() => allowButton?.focus(),
200)`. The 200ms delay is to let the card-scale-in finish before the
focus halo joins the choreography.

After focus is on Allow, the modal installs a window-level `keydown`
listener that:

1. **Traps Tab** — focus cycles between Allow / Deny / Trust toggle (in
   that order: Allow → Trust toggle → Deny → Allow). Shift+Tab reverses.
2. **Traps Shift+Tab** — symmetric.
3. **Esc** → calls `deny()` (mapped via §6.2).
4. **Enter** → activates the focused button:
   - If Allow has focus → calls `approve()`.
   - If Deny has focus → calls `deny()`.
   - If Trust toggle has focus → toggles the switch (does NOT call
     approve/deny — Enter on the toggle is the toggle action).
5. **Space** → if Trust toggle has focus, toggles the switch.

The trap does NOT prevent the user from clicking outside the card (the
scrim is a sibling of `<Shell>`; nothing else is interactive behind
it). It DOES prevent Tab from leaving the card.

### 6.2 Esc maps to Deny (the safe default)

`Esc` calls `deny()`. This is intentional and is repeated as the
hint text in the footer (`Esc to deny`). Per `CLAUDE.md §2.1` invariant
#3 (destructive actions require a real human at the keyboard), the
"safe exit" from the modal is **deny**, not close-and-defer. Esc cannot
just dismiss the modal — it must route through the Gatekeeper's deny
path so the audit log captures an explicit refusal.

**Why Esc maps to Deny, not Close:** if Esc just hid the modal, the
ticket would remain pending on the daemon and the agent would eventually
time out, with no human decision recorded. Mapping Esc to Deny makes
the refusal as explicit and auditable as a Deny click.

### 6.3 Enter on Allow → Approve

`Enter` while Allow has focus calls `approve()`. The 200ms entrance
delay (§6.1) means the user has to wait briefly before they can
keyboard-confirm, but this is correct — the entrance is part of the
"this is consequential" message.

### 6.4 Space on Trust toggle → toggle the switch

Per `MOAT §2.7` (tactile consistency), the toggle uses Space and
Enter the same way form checkboxes do. The `<Switch>` component is
expected to follow platform conventions (Space toggles, Enter
submits).

### 6.5 Tab order (the explicit chain)

| Tab stop | Element | Visible affordance |
|---|---|---|
| 1 | **Allow** button | `Allow` (or `Allowed` while sealing). Primary pollen pill. |
| 2 | **Trust toggle** (if present) | `<Switch />`. Mono-pollen `Trust this app` label + one-line sub. |
| 3 | **Deny** button | `Deny` (or `Denied` while shaking — but the shake usually finishes the modal exit). Ghost or danger (destructive). |

**Trust toggle is skipped if `action_kind === 'read'`** (READ has no
trust decision).

### 6.6 Outside-tab prevention

The keydown handler at the window level cancels the default Tab
behavior when `event.key === 'Tab'` is pressed and focus would
otherwise leave the card. Implementation:

```ts
function trapTab(e: KeyboardEvent) {
  if (e.key !== 'Tab') return;
  const focusables = [allowEl, trustEl, denyEl].filter(Boolean);
  if (focusables.length === 0) return;
  const first = focusables[0];
  const last = focusables[focusables.length - 1];
  if (e.shiftKey && document.activeElement === first) {
    e.preventDefault();
    last.focus();
  } else if (!e.shiftKey && document.activeElement === last) {
    e.preventDefault();
    first.focus();
  }
}
```

This is the standard modal focus-trap pattern. Implementation in
`onMount`, removal in the cleanup function.

### 6.7 No global shortcuts honored while modal is open

`Cmd+K`, `Shift+P`, `Shift+O`, `?`, ⌘[ / ⌘] — all of these are
blocked while a ticket is pending. The modal's `keydown` handler calls
`stopPropagation()` on every key except `Tab`, `Shift`, `Enter`, `Esc`,
and `Space`. The user cannot summon another modal while a consent
prompt is up — per `CLAUDE.md §2.1` invariant #3, "blocks until
clicked."

The exception is the kill switch (Cmd+Shift+Escape): that hard
hotkey is intercepted by the OS-level watcher, not the modal, and
its handler still runs. The user can ALWAYS halt the agent.

---

## 7. Components Used

The modal is composed of these components. The full prop contracts
below belong in each component's source file — this spec is the
short-form.

### 7.1 `<ConsentModal />` — current file (314 lines)

The route-level component. Mounts globally at `Shell.svelte:246`.

**Owns:**
- The scrim + card layout per §3.
- The armor-rect draw animation.
- The wax-seal stamp animation.
- The shake-on-deny animation.
- The focus trap (Tab / Esc / Enter / Space handling).
- The countdown bar (reading from `consent.timer`).
- All IPC calls to `gatekeeper.approve` / `gatekeeper.deny`.

**Does not own:**
- The polling loop (owned by `consent` store in `consent.svelte.ts`).
- The policy decisions (owned by the daemon's Gatekeeper).
- The Switch component's internals (just composes it).
- The Glyph rendering (just composes `<Glyph />`).
- The thread footer hairline (just composes `<Thread />`).

### 7.2 `<Button />` — `condura/Button.svelte`

Props used: `variant: 'primary' | 'ghost' | 'danger'`, `magnetic:
boolean`, `class: string`, `onclick: (e) => void`, children. The
Allow button uses `variant="primary" magnetic={true}`; the Deny
button uses `variant="ghost"` for non-destructive or
`variant="danger"` for destructive. The pill radii follow
`--r-pill`, the focus halo follows `Button.svelte:75-80` (rounded).

### 7.3 `<Glyph />` — `condura/Glyph.svelte`

Icons used:

| Where | Glyph name | Size | Default stroke |
|---|---|---|---|
| Trust toggle (when shown) | `shield` | 14 | 1.5 |
| Allow button label (after sealing) | `check` | 14 | 1.5 |
| Deny button label (after shaking) | `close` | 14 | 1.5 |
| Consent-provider footer info icon | `info` | 14 | 1.5 |
| Tooltip on the footer info icon | `info` (glyph) + `<Tooltip label>` (text) | n/a | n/a |

If any icon doesn't exist in `icons.ts` (per the grep against `icons.ts`:
`shield`, `check`, `close`, `info` all exist), the spec says **do not
invent a new one inline**; add to `icons.ts` per `MOAT §4 #2`.

### 7.4 `<Switch />` — to be created

Trust-toggle primitive. Lives at `condura/Switch.svelte`.

```ts
let {
  checked: boolean,
  onchange: (next: boolean) => void,
  label?: string,        // aria-label
  disabled?: boolean,
  class?: string,
} = $props();
```

The Switch is the trust toggle for §3.4. It renders a horizontal pill
with a circle that slides. Pressed state: the circle is at the right,
the pill's fill is `--synapse`. Unpressed: the circle is at the left,
the pill's fill is `var(--hair-strong)`. The component uses the global
`--r-pill` shape and the global tactile (`MOAT §2.7`) press. No
component-internal `prefers-reduced-motion` (per `MOAT §2.3`).

### 7.5 `<Thread />` — `condura/Thread.svelte` (already exists)

```ts
let { orientation: 'h' | 'v' = 'h', draw = true, glow = true, class?: string } = $props();
```

Used as the hairline between Region B's sub and the action-summary
card. One `<Thread orientation="h" draw={false} glow={false} />`
renders the static 1px divider. (`draw={false}` because the hairline
is structural, not a moment of completion.)

The armor rect in Region B is **not** a `<Thread />` — it's a custom
`<svg>` with rounded corners. `<Thread />` only renders straight lines.
But the armor's behavior (1.5px stroke, `pathLength="1"`,
`stroke-dashoffset 1 → 0`, 1.4s `--ease`) is exactly the Thread idiom
applied to a closed shape.

### 7.6 `<Tooltip />` — to be created per `MOAT §2.9`

```ts
let {
  label: string,
  placement: 'top' | 'right' | 'bottom' | 'left' = 'top',
  delay: number = 400,
  exit: number = 75,
  children: Snippet,
} = $props();
```

Used once on the consent-provider footer info icon. Hover-delay 400ms,
exit 75ms (per the global Tooltip contract).

### 7.7 `<GlyphShield />`, `<GlyphCheck />`, `<GlyphClose />` — aliases

These can be wrappers that delegate to `<Glyph name="shield|check|close"
/>`. They're not separate components; `<Glyph />` covers all of them.
The spec mentions them only as a naming convenience for the impl.

---

## 8. Data Fetched

The modal's IPC contract is **four** methods. Three are existing
(`gatekeeperPendingConsent`, `gatekeeperApprove`, `gatekeeperDeny`);
one is **new** in this spec: `gatekeeperTrustApp`.

### 8.1 `ipc.gatekeeperPendingConsent()`

RPC: `gatekeeper.pending_consent`. Returns
`ConsentPendingResult { tickets: ConsentTicket[] }`. The modal polls
this every 1.2s via `consent.start()` (in `consent.svelte.ts:22-28`).
The first poll runs on mount (`void this.poll()` synchronously after
the interval is set).

```ts
// Existing: app/web/frontend/src/lib/ipc/client.ts
gatekeeperPendingConsent(): Promise<ConsentPendingResult>
```

If `ipc.isConnected()` returns false (daemon unreachable), the poll
silently returns; the modal stays closed. The store's `error` field
is set to the error string but never surfaced to the UI (the modal is
the only render path for tickets; with no tickets there's no modal).

If `tickets.length > 0`, the **first** ticket is shown. The first is
the daemon's choice — it's not arbitrary. (The daemon enforces ticket
ordering per its queue.) When the daemon resolves or expires the
ticket, the next poll returns `tickets.length === 0` and the modal
unmounts.

### 8.2 `ipc.gatekeeperApprove(nonce, opts?)`

RPC: `gatekeeper.approve`. Resolves an existing ticket with the user's
consent. The `opts` parameter is **new** in this spec:

```ts
gatekeeperApprove(
  nonce: string,
  opts?: {
    trustApp?: {
      appBundleId: string,  // e.g. "com.apple.Mail"
      taskType: 'write' | 'network' | 'destructive',
      // matrix cell to write: apps.<bundle-id>: warn → autonomous
    }
  }
): Promise<{ ok: boolean }>
```

When `trustApp` is omitted (the Trust toggle is off, or the ticket
class is READ), the call passes `{nonce}` only and the Gatekeeper
records approval without writing the autonomy matrix.

When `trustApp` is set (the Trust toggle was on), the daemon:
1. Records the approval in the audit log with `reason: user_trust`.
2. Updates `~/.synaptic/autonomy.yaml` cell `apps.<bundle-id>` from its
   current state (warn or autonomous, never block) to `autonomous`.
3. Returns `{ok: true}` after both writes succeed.

If the autonomy matrix write fails (e.g. file permission), the call
returns `{ok: false, error: 'matrix_write_failed'}` and the modal
shows an inline error in the Trust toggle's sub line:
`Could not record trust preference. Matrix write failed.`

### 8.3 `ipc.gatekeeperDeny(nonce)`

RPC: `gatekeeper.deny`. Resolves the ticket with an explicit denial.
The audit log captures the denial with `reason: user_deny`. The agent
receives a "denied" response and the action does not proceed.

```ts
gatekeeperDeny(nonce: string): Promise<{ ok: boolean }>
```

No opts — denial is always just denial. The audit chain captures
which model was asking, what it wanted to do, when the user denied,
and the nonce of the refusal (for cross-referencing with later
audit queries).

### 8.4 `ipc.gatekeeperTrustApp()` — new in this spec

If the gatekeeper prefers to handle trust separately from approval,
this RPC writes the matrix cell directly:

```ts
gatekeeperTrustApp(args: {
  appBundleId: string,
  // The task class for the autonomy matrix cell.
  taskType: 'write' | 'network' | 'destructive',
  // Always 'autonomous' from a Trust toggle click.
  state: 'autonomous',
}): Promise<{ ok: boolean }>
```

This is an alternative implementation to §8.2 — the daemon may prefer
to keep the two concerns separated so an approval can succeed even if
the trust write fails (or vice versa). The spec leaves the
implementation choice to the daemon; the GUI calls whichever IPC the
daemon exposes. Either way, the user-facing affordance is identical.

The spec recommendation: **prefer §8.2** (pass `trustApp` as an opt
to `gatekeeperApprove`). This makes the approval and the trust a
single human decision, recorded atomically. Splitting them risks a
"approved but not trusted" race. (See `CLAUDE.md §2` invariant #1,
#5: "Every action is auditable, in a tamper-resistant log." A split
transaction creates two log entries; an atomic transaction creates
one.)

### 8.5 No retry, no backoff

The IPC calls do not retry on failure. If `gatekeeperApprove` rejects,
the modal stays open, the inline error appears, and the user can
click again. There is no automatic retry — the user is always in
control of the consent decision.

The polling loop (§8.1) silently swallows errors. The daemon may be
restarting, the network may be flaky, or the user may be on a slow
machine; none of these should auto-dismiss the modal if it's already
up. The 1.2s poll cadence continues; if a previously-displayed ticket
vanishes from the next poll (rare — daemon deletes resolved tickets),
the modal unmounts cleanly.

---

## 9. Design Decisions

These are the load-bearing calls — the place where this spec
disagrees with the current `ConsentModal.svelte` and where it earns
the MOAT bar.

### 9.1 The modal IS the survival rule — it cannot be dismissed

**The problem (per `CLAUDE.md §2.1` invariant #3).** "Destructive
actions require a real human at the keyboard. Native modal dialog.
Blocks until clicked." The current `ConsentModal.svelte` mostly
honors this — it has no dismiss button, only Deny. But the modal
also doesn't make the *why* visible: the user has to *trust* that
the Gatekeeper is the obstacle, not the model.

**What this spec does.** Three signals, layered:
1. **The armor-rect draws on every new ticket arrival.** This is the
   visible "this is being protected" gesture — the action summary
   card is, by visual convention, *under guard*.
2. **The footer hairline carries the consent-provider explanation.**
   "The Gatekeeper is a local rules engine. Your model did not decide
   to ask — the policy did." This is the only place in the app where
   the Gatekeeper / model distinction is shown; it's shown *exactly
   when it matters*, not in a Settings page the user never visits.
3. **Esc maps to Deny, not Close.** This is in the footer hint copy.
   The user reads `Esc to deny` and learns the safe-exit keybinding.

All three signals together communicate: "this is enforced by the
local rule engine, not by the chat model; your safe exit is to
deny." Per `MOAT §4 #5` and `APPFLOW.md §1 I5`: the 7 invariants
are visible.

### 9.2 The blast-radius colors are four, not five

**The problem.** `MOAT §4 #4`: no rainbow accents. The brief lists
four colors (READ green, WRITE pollen, NETWORK danger, DESTRUCTIVE
deep-danger). The risk is over-coloring — if every blast class gets
its own color anywhere on the modal, the modal becomes a status grid.

**What this spec does.**
- **Eyebrow color** carries the blast-radius color (`--ok` / `--pollen`
  / `--danger` / `var(--surface-ink)` for the destructive eyebrow
  which uses ink, not danger-red — destructive is *not the same as
  danger-network*; it gets its own surface).
- **Armor rect is always `--synapse-glow`** regardless of class. The
  protection is one color.
- **The "Trust this app" toggle's wording** is the only place the
  blast class is referenced in user-facing copy. The user reads
  "Trust this app for these actions in the future" — they don't
  need the color to be told the class; they were just shown the
  articulated action.

This keeps the modal to two colors of state (paper / ink-paper for
destructive) plus one armor color, plus the existing accent set in
the rest of the app.

### 9.3 The Trust toggle is honest — it edits the matrix, not a magic rule

**The problem.** A naive "Don't ask again for this app" button would
feel magical — the user would have no idea where the decision goes.
Today's apps solve this with hidden "exception lists" the user can't
audit (e.g. macOS Accessibility exceptions, browser notification
permissions). This is exactly the "we pretend to give the user
control" pattern that `MOAT §4 #6` and `CLAUDE.md §2.1` invariant #6
forbid.

**What this spec does.** The Trust toggle's sub line says
`Add to autonomy matrix.` The user can open Settings → Autonomy
matrix and see the row they just created. The toggle's IPC §8.2
writes the cell the same way the Settings matrix UI does. The user
has one source of truth for "apps I've trusted" — the matrix — not a
hidden list inside the consent code.

**Why the destructive caveat is in the sub line, not the IPC:** per
`CLAUDE.md §2.1` invariant #3, destructive actions still require
real human presence even with the app trusted. The toggle writes
`apps.<bundle-id>: autonomous` only for the blast class. Destructive
actions on the trusted app would still queue if the user is away
(because the Gatekeeper's `require_user_active: true` rule still
applies — that's a separate policy, not the app-trust setting).

### 9.4 The wax seal is a stamp, not a celebration

**The problem.** `MOAT §4 #6` forbids fake enthusiasm. "Awesome!" /
"Great choice!" / "You're all set!" toasts are exactly the energy
this surface must not have. The user is *allowing an action that may
be irreversible*; the appropriate response is gravitas, not confetti.

**What this spec does.** The wax seal is a 108px circle with the
radial-gradient synapse palette and two lines: `Allowed` (display 20)
above `by you · now` (mono 8, uppercase, +18% letter-spacing). No
emoji, no checkmark, no "🎉 Action complete!" The seal is a *stamp*,
not a *fanfare*. The icon next to it is the same `check` glyph that
the rest of the app uses for "this is done" — no new icon invented.

The seal's shadow is `box-shadow: 0 20px 50px -16px color-mix(in
oklab, var(--synapse) 70%, transparent)` — a heavy drop shadow at the
underside only, the same shadow that a real wax seal would cast.
This is one of the rare double-shadow surfaces in the app (per
`MOAT §4 #9`, double shadows are normally forbidden). The exception
is documented here: the seal is a load-bearing metaphor, not
decoration, and the shadow is part of the metaphor.

### 9.5 The shake on deny communicates refusal without animation noise

**The problem.** A naive shake animation is fast and jittery — it
communicates "error" rather than "you said no." The user just made a
considered decision; the modal's response should match the
deliberateness, not the speed.

**What this spec does.** The shake is 200ms total (1 → 0.96 → 1.02 →
1), not 50ms. Three keyframes, smooth easing. The card *settles* at
the end — the user's decision has been recorded; the modal catches
its breath and fades. This is "the form returned" rather than "the
form shook."

The shake is also one of the rare places a `transform: scale` is
applied to a Card, not a button. The press feedback in `MOAT §2.2`
applies to buttons; the modal's response is different by design —
it's the surface responding, not a control being pressed.

### 9.6 The countdown bar communicates time without numerals

**The problem.** A countdown timer with numerals would say
"3 minutes 47 seconds left" — but the user doesn't need to make a
decision at 3:47. They need to know "the system is waiting on me
and will stop waiting." A *bar* communicates that better than a
*number*.

**What this spec does.** A 2px bar across the card width fills as the
timer drains. There are no numerals. There is one state change: below
60 seconds, the bar pulses (`pulse-urgent`, 600ms alternate) to
indicate the queue is imminent. The pulse is alpha-only, not color —
the same warn tone, just breathing.

Per `MOAT §4 #10`, every animation answers "what is this
communicating?" The bar communicates "the system has a deadline";
the pulse communicates "the deadline is close." Neither needs to
say "you have X seconds."

### 9.7 The modal is a global mount, not per-route

**The problem.** If the modal lived in `#/chat` or `#/skills`, the
agent could trigger a consent-required action from `#/delegation`
and the user wouldn't see it. Per `CLAUDE.md §5`, the Gatekeeper is
the only path to physical action; its GUI face must be available
regardless of which route the user is on when the ticketed action
arrives.

**What this spec does.** The component mounts at `Shell.svelte:246`
alongside `<KillSwitchOverlay>`, `<DynamicIsland>`,
`<CommandPalette>`, and `<QuickPromptOverlay>`. It is a sibling of
the entire app, not a child of any route. The `consent` store starts
its poll on mount and stops on unmount; the only way the modal
disappears from the screen is for `consent.ticket` to be null.

### 9.8 The countdown never blocks the user from clicking

**The problem.** A timer that *blocks* clicks would be a dark
pattern — the user must be able to make the decision whenever they
make it, even if the bar is fully drained. The daemon's on-timeout
policy (`~/.synaptic/policy.yaml`) decides queue vs. deny; the modal
is not the enforcer of that policy.

**What this spec does.** Allow and Deny are always enabled. Clicking
Allow at 0s submits `gatekeeperApprove(nonce)`, which records the
approval and clears the ticket. The daemon resolves as though the
user had approved before the timer expired (because they did — the
user clicked Allow, not the timeout). The countdown is an
*indicator*, not a *gate*. `MOAT §1.1` and `APPFLOW §1 I5`: the
invariants are visible — and the user's control is the most visible
of all.

---

## 10. What this spec adds to / deletes from the current `ConsentModal.svelte`

The following table is the explicit delta the implementation must do
when this spec replaces the existing file. Each row cites the rule
that motivates the change.

| Current `ConsentModal.svelte` | What it does now | What this spec says |
|---|---|---|
| `width: min(540px, 92vw)` (line 153) | Card is 540px wide max. | Tighten to `min(520px, 92vw)` to match the brief. |
| **No Trust toggle.** The component only has Allow / Deny + countdown. | The user has no way to record "don't ask again for this app." | Add Region C: a `<Switch />` row prefaced by `Trust this app` + a one-line sub describing the matrix write. Skipped for READ tickets (auto-allow). |
| **No consent-provider explanation.** | The user is told to trust the modal is doing the right thing. | Add Region D: a hairline + a one-line italic explanation (`The Gatekeeper is a local rules engine...`) + an info `<Glyph />` with a `<Tooltip label="What does Gatekeeper mean?">` showing the longer explanation. |
| **No Tooltip primitive in component imports.** | The footer info icon would have to use `title=` to provide its explanation. | Add `<Tooltip>` to the import list (per `MOAT §2.9`). |
| No `Switch` primitive exists in condura/. | The Trust toggle is unimplemented; if added naively it would be a hand-rolled `<button>` with two CSS states. | Add `condura/Switch.svelte` per §7.4. Reused in Settings → Voice (the wake-word toggle) and the future "Trust this channel" UI in Channels. |
| `consent.approve(nonce)` (line 68) | Single-arg approve. | Extend to `consent.approve(nonce, {trustApp: ...}?)`. Default unchanged; the second arg is optional and gated by the Trust toggle state. |
| **No `gatekeeper.trustApp` IPC method** in the type contract (line 836-855 of `ipc/types.ts` — just `approve` / `deny` are there). | The Trust toggle would have to call a new IPC that doesn't exist. | Add the IPC per §8.4 (or, recommended, fold into `gatekeeper.approve`'s opts per §8.2). |
| Inline `@keyframes stamp` exists (line 299) | The wax-seal animation runs but the keyframe sequence is `scale(0) → scale(1)` with no settle. | Refine to `scale(0) → scale(1.04) → scale(0.96)` per §5.2 — adds the "settling" gesture that makes the seal feel physical. |
| No shake animation exists. | Denying the modal would just close it — no visible response to the user's decision. | Add the `.deny-shake` class + `deny-shake` keyframe per §5.3 — 200ms shake + fade. |
| `transition: stroke-dashoffset 1.4s var(--ease) 0.15s;` (line 109) | Armor-rect draws on mount. | Keep exactly. The spec's Stage C in §5.1 is identical; this becomes the documented behavior. |
| `role="alertdialog" aria-modal="true"` (line 85) | Standard modal aria semantics. | Keep; add `aria-labelledby` pointing to the title (`Condura wants to act.`) and `aria-describedby` pointing to the body sub. The screen-reader experience is: announce title, then body, then buttons. |
| No `prefers-reduced-motion` blocks in the file. | The component has no per-element motion handling. | Keep none. Per `MOAT §2.3`, the global rule in `condura.css` handles it. The component is **not allowed** to declare its own media-query blocks — doing so is explicitly forbidden. |
| **No focus trap.** Tab walks through every focusable on the page (Composer, settings link, etc.) while the modal is open. | The user can Tab to "Settings" and dismiss the modal that way. | Add the window-level keydown listener per §6.6: trap Tab, route Esc to deny, route Enter to focused-button. |
| Initial focus: `<Button variant="primary" ... class="allow">` is the first focusable (by tab order) but no `focus()` is called. | The user must Tab once to focus Allow before pressing Enter. | Add `setTimeout(() => allowButton?.focus(), 200)` in `onMount`. The 200ms is per §5.1 Stage D. |
| The countdown bar's color is `var(--warn)` (line 274) and never changes. | Below 60s, the bar looks identical to 4 minutes remaining. | Add a `.countdown-fill.urgent` class that applies `pulse-urgent` (alpha 0.85 ↔ 0.55, 600ms alternate) when `consent.timer <= 60_000`. |
| **No tooltip on the countdown bar.** | The user might wonder "what happens when this hits zero?" | Optional: add a `<Tooltip label="Queue by default — see autonomy matrix for the policy">` on hover, with delay 400ms. Lower priority than the consent-provider footer tooltip. Documented here for the eventual cleanup pass. |
| The wax seal holds `scale(1)` (line 299 `transform: ... scale(0)` then `animation: stamp ... forwards`). | The seal arrives and stays at scale 1; no settle gesture. | Refine per §5.2 to arrive at 1.04 then settle to 0.96, with the seal held at 0.96 for an additional 320ms before fade-out. |

After these changes, the file grows from 314 lines to roughly 440 lines
— the **consent-modal** is a load-bearing safety surface, and the
expectation is that the code is dense with comments, not sparse. The
component file becomes the canonical documentation of "how the
Gatekeeper shows itself to the user."

---

## Closing note for Phase 4

The ConsentModal is one of three load-bearing safety surfaces —
alongside the **Autonomy Matrix** in Settings and the **Audit chain**
viewable in `#/audit`. Each one shows the same survival rule from a
different angle: the matrix lets the user pre-authorize, the modal
enforces at the moment of action, the audit records what happened.
A user who uses all three will never have to guess "what did the
agent do, why did it do it, and can I stop it?" — the answer is on
screen.

The eleven things you'll find when you implement from this spec:

1. **Yes** to the armor-rect drawing in left-to-right on every new
   ticket — already implemented; documented.
2. **Yes** to the wax seal stamp with the 1.04 → 0.96 settle — refine
   the existing keyframe.
3. **Yes** to the shake gesture on deny — new, 200ms, three-keyframe.
4. **Yes** to the Trust toggle (Switch primitive) for WRITE / NETWORK
   tickets — new, with the destructive caveat.
5. **Yes** to the consent-provider footer — new, with the info
   Tooltip explaining the Gatekeeper.
6. **Yes** to the focus trap — new, with Esc → Deny and Enter on
   focused Allow.
7. **Yes** to the focus-delay of 200ms — new, so the focus halo
   doesn't fight the entrance choreography.
8. **Yes** to `pulse-urgent` below 60s — new, alpha-only, no numerals.
9. **Yes** to the `gatekeeperApprove(nonce, {trustApp?})` signature —
   new, atomic record of approval + trust in the audit log.
10. **Yes** to the global `prefers-reduced-motion` discipline — no
    per-component media-query blocks.
11. **Yes** to the unchanged thread grammar — the armor-rect draw and
    the seal stamp are both Thread-idiom applied to closed and
    radial shapes, respectively. They carry the moment's weight the
    same way the titlebar thread does at 520ms.

If during Phase 4 you find a place where the spec is silent and the
default would be wrong, the answer is to ask before inventing. The
MOAT bar is "premium-quality." Anything that fails it is competent-
but-generic, and competent-but-generic is how a year gets lost.
This surface in particular — the place the user sees when the
agent is about to do something irreversible on their machine — is
the *least* appropriate place to be competent-but-generic.
