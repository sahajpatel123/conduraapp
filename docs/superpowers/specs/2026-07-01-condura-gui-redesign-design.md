# Condura GUI Redesign — Design Spec

> **Date:** 2026-07-01
> **Status:** Draft for approval. Implementation follows explicit user approval of this document.
> **Scope:** Desktop application GUI (Wails + Svelte 5 frontend at `app/web/frontend/`). Does NOT touch the marketing site (`web/`).

---

## 0. One-Line Creative Brief

> A quiet companion that gets out of the way — premium, restrained, alive in the *details*, never in the chrome.

The user opened Condura for the first time and saw something feeling "vibe coded, no soul." This spec is the antidote.

---

## 1. Why Prior Work Felt "Vibe Coded" (and our refusal)

| Vibe-coded tell | Why it kills premium | What we do instead |
|---|---|---|
| Glassmorphism everywhere | Glass is "AI-era default"; reads as a template | Matte paper-and-ink surfaces, soft grain, no faked translucency |
| Neon-on-dark everywhere | Tries to look futuristic; ends up looking dated | Warm low-saturation palette; one accent, used sparingly |
| Every element animated | Animation noise = nervous = cheap | Motion is *acknowledgment* — fires on state changes only |
| Borderless cards on gradients | Generic HTML5 demo | Hairlines that are actually 1px at low alpha |
| Generic "AI sparkle" icons | Reads as stock | Single-stroke SVG icons drawn at 1.5px for our vocabulary |
| Floating everything | Layout becomes unintelligible | Strict type/grid ladder; surfaces earn elevation by hierarchy |
| Loading: shimmer skeletons | Lies about progress | Real percentages when known; honest "thinking…" copy when not |

---

## 2. The Five Design Principles

1. **Restraint over decoration.** If a pixel isn't *doing work*, remove it.
2. **Type as the visual.** Display sans for headings, true humanist for body, mono for status. The chrome fades; the words carry.
3. **Motion as acknowledgment.** Every transition tells the user "I see you" or "I did that" — not "look at me."
4. **Material honesty.** Surfaces feel like real things: paper, wood, soft fabric. Never faked material.
5. **One accent, used sparingly.** A single warm earth-amber anchors the palette. Everything else is grayscale + a quiet paper white.

---

## 3. The Palette (15 tokens, that is all)

```
paper       #F7F4EE   warm white canvas
paper-2     #EFEAE0   subtle recessed surface
ink         #1B1A17   true near-black for text
ink-2       #4A463E   secondary text
ink-3       #8A847A   tertiary, captions, hairlines at 100%
rule        #D9D2C2   hairline borders at low alpha (30-60%)
accent      #C18A4A   warm earth — the ONE color
accent-ink  #7A4F1E   accent text on paper
signal-go   #5C7F4A   success, quiet
signal-warn #B07A2E   cautious, warm
signal-stop #A84A3F   destructive, restrained
surface     #FFFFFF   elevated card
shadow-1    2 2 0     sharp, near-zero blur — for "pressed"
shadow-2    8 12 24   lift — for hovered surfaces
shadow-3    0 24 60   peak — for overlays and the consent card
focus       #C18A4A   at 40% alpha ring only (no jarring outline)
```

---

## 4. Typography — Type *is* the design

- **Display:** *Instrument Serif* (variable, true italic). Headings read like letters on a page, not UI labels.
- **Sans (UI):** *Inter* 400 / 500 / 600 with `font-feature-settings: 'ss01','cv11'` — typographic details users feel, not see.
- **Mono (status, code, timers):** *JetBrains Mono* 400 / 500 with slashed zero.
- **Scale:** 11 / 12 / 14 / 16 / 20 / 28 / 40 / 64 — a true modular scale, not 14 arbitrary sizes.
- **Body:** 16/1.55 — warm line-height for screens.
- **Numerals:** tabular figures everywhere in status, time, money — nothing jiggles.

---

## 5. Spacing, Grid, Radius

- **Base unit:** 4px. Scale: `4, 8, 12, 16, 24, 32, 48, 64, 96`.
- **Grid:** 12 columns, 24px gutter, 32px page margin (we are a desktop product; do not waste pixels).
- **Radius:** `0 / 4 / 8 / 16 / 999`. Most things are 4 or 8. Pills are 999. Nothing larger — ever (no "big rounded card" vibe-coded nonsense).

---

## 6. Motion Grammar — *alive but quiet*

All motion is a CSS custom-property runtime (`motion.css`) so it is orchestrated, not bolted on.

```
--ease-out-soft    cubic-bezier(.22, 1, .36, 1)   entry, arrival
--ease-in-honest   cubic-bezier(.55, .06, .68, .19) exit, dismiss
--ease-spring      cubic-bezier(.5, 1.4, .4, 1)     state-change, acknowledgment
--dur-fast         140ms  hover, focus, micro
--dur-mid          280ms  panel entry, list reorder
--dur-slow         520ms  overlay arrival, route transition
--dur-cinematic    900ms  first-paint, "The Touch"-style emergence
```

**The rule:** every motion must *mean one of these four*:

1. "I see the arrival" (entry)
2. "I see the departure" (exit)
3. "I changed state for you" (acknowledge)
4. "I want your attention" (rare; opt-in only via focus, no autoplay)

No decorative motion. No infinite loops in chrome. No glow pulses.

---

## 7. Sound — *opt-in but iconic*

A single, 16-instrument UI sound set, off by default:

- **"paper"** — soft tap for state commit
- **"settle"** — low thunk for panel arrival
- **"void"** — muted breath on dismiss
- **"stamp"** — gentle seal for consent approval (earned at Gatekeeper)

Master toggle in Settings; per-sound toggles; respects system mute. Sounds are <50ms and <−20dBFS. They are *signals*, not music. They exist so the user can *feel* the agent's state changes even when their eyes are elsewhere.

---

## 8. The Mandatory First-Time Floating Interview

**Not a wizard. Not a stepper. A floating interviewer anchored bottom-right.**

The user opens Condura for the first time. They see the **normal app shell** in the background (so they understand what they are configuring *for*), with **one floating panel** at bottom-right asking one question at a time:

1. "What should I call you?" — single field, ghost label animates on focus
2. "How do you want me to answer?" — 3 cards: *Local only* / *Cloud when needed* / *Always cloud*
3. "Anything you want me to *never* touch?" — chips: email, money, calendar, files, code (multi-select)
4. "What does a great day with me look like?" — freeform, optional
5. "Pick a hotkey" — live recording

Each answer *commits visibly* — the previous question becomes a quiet pill at the top of the next. After answer 5, the panel slides off (280ms ease-out), revealing the now-configured app underneath. **No ceremony. No confetti.** Just: the agent is there.

The same panel re-anchors (smaller, top-right corner dot) any time the user opens any longer-form settings door. Press `?` and it appears as a "what is this?" tour.

---

## 9. Surfaces — What Each Screen *Feels* Like

### 9.1 The Chat Surface (home)
A scroll, not a screen. Newest message at the bottom, paper edge texture at top of viewport, composer anchored bottom with a single line that grows. Voice mode: the whole canvas darkens to `paper-2` and a single orb breathes in the bottom-center — no chat surface, just presence. Toggle back: paper returns, conversation exactly where you left it.

### 9.2 The Overlay (Cmd+Shift+Space)
A sticky note placed on the user's screen, not a panel. Appears bottom-right of cursor with a 280ms slide-up + fade + a single 4px paper-shadow that *settles* after the motion (separate stage of the animation — that is the "alive" detail). Has its own micro-composer. ESC dismisses with the *void* sound.

### 9.3 The Sidebar
A book spine, not a strip. Vertical column, 72px wide when expanded, collapses to 8px when not in use (just an indicator dot). Each route is a 2-letter monogram + label (rotated 90° when collapsed). On hover, the label slides into view from the side; on click, the page content reflows with `--dur-mid` ease-spring.

### 9.4 Settings
A document, not a form. Sections are numbered chapter headings (`01 · Account`, `02 · Voice`, `03 · Channels`…). Toggles are hardware-honest — 28px switch with a 1px groove, not a flat rectangle. Save buttons are quiet text bottom-right ("Save changes" appears when there is something to save).

### 9.5 Status Bar (bottom strip, always present)
A typographic vital-signs display. Mono typeface, single line. Left: agent name + `·`. Center: current task with stopwatch running (real seconds, not fake animation). Right: queue depth, today's spend, network status. The whole bar pulses subtly (1px of background alpha) once per second while the agent is doing work — that is the agent's heartbeat.

### 9.6 Audit
An evidence locker. Each entry a paper card with a left rule in `ink-2`, top-right a hash, body in mono. Click a row → expands to reveal the full reasoning trace (collapsible, like a court transcript). HMAC verification at the top with a green checkmark or red exclamation — no confetti either way.

### 9.7 Consent Modal (the Gatekeeper surface)
A wax seal on a letter. Appears as a centered card (max 480px), with a top "ribbon" of `accent` as the "this needs a human" signal. Body explains the action in plain language ("You are about to send an email to *alex@example.com*"). Two buttons: **Deny** (quiet) and **Allow once / Allow for this session** (primary). On Allow: the *stamp* sound fires. This is the moment the wax seal *visibly stamps* — a CSS animation of a circle drawing in `accent` over 280ms.

### 9.8 Channels / Delegation / Hub / Replay / Sync / Skills / About

| Route | Mental model | Live-in-the-chrome detail |
|---|---|---|
| **Channels** | A control panel of radios. Each channel is a "tuner" row with signal-quality dots. | Connecting a channel plays a brief "carrier-lock" tone (sound). |
| **Delegation** | A control room. Each sub-agent is a "station" with a real-time waveform of its activity. | Sub-agent output streams directly into a "monitor" frame that gently pulses when active. |
| **Hub** | A library with a card catalog. Skills are real book spines; browse = pull one out. | Hover: spine tilts 4°. |
| **Replay** | A film strip you can scrub. Scrubber is a real timeline with 24h of frames. | Frame thumbnails crossfade as you scrub. |
| **Sync** | Two devices meeting in a quiet room. Pairing is a "handshake" — your device is on the left, theirs on the right, and a curved line animates between them. | The connecting line draws in `--dur-cinematic`, then *settles*. |
| **Skills** | Your local library — same card-catalog model as Hub, but with a "loaded" dot on each. | Loaded skills have a faint red-warm ribbon at the top. |
| **About** | A colophon. Like the credits of a thoughtful book. Subtle, not boastful. | A single word cycles in the footer: "With care." → "For everyone." → "Forever free." |

---

## 10. The Implementation Roadmap

| Phase | What ships | What it proves |
|---|---|---|
| **1. Foundation** | `app/web/frontend/src/lib/v2/` — `tokens.css`, `motion.css`, `surface` primitives, `sound.ts`. Wired into `App.svelte` as the new theme. | The design language can be expressed in code. |
| **2. Showcase** | The 3 hero moments: Chat surface, Overlay arrival, Mandatory first-time floating interviewer. | They feel alive and premium. |
| **3. Surfaces** | Each remaining route rebuilt on the v2 system. Sidebar, Settings, StatusBar, Audit, Consent, plus Channels/Delegation/Hub/Replay/Sync/Skills/About. | Every screen looks like it belongs. |
| **4. Polish** | Sound implementation, focus rings tuned, motion prefers-reduced fallback, microtypography tweaks. | The $50M finish. |

**Hard rule:** Never touch the existing `v1/` components. Build `v2/` as a parallel sibling and reconcile at the end.

---

## 11. Open Questions for User Review

1. **Warm earth as the single accent** — keep, swap (forest / slate / ink-blue), or add a second accent for warnings?
2. **Sound** — keep all four on opt-in, or ship silent with sound as a v0.2.0 toggle?
3. **Floating Interview anchoring** — bottom-right is the default; alternative is bottom-center. Preference?
4. **Composer on Overlay** — full chat composer or single-line only? (Single-line = faster; full = matches Chat.)
5. **Status bar position** — always-bottom is current; alternative is always-top for a "transcript header" feel.

---

## 12. The Crucial Constraint

`web/` (the marketing site) is KIMI K2.6's territory and not in scope.

`app/web/frontend/src/lib/components/v1/` contains the user's uncommitted WIP. Do not modify, commit, or review it. We build `v2/` parallel and reconcile later.

---

**End of spec.** Awaiting user review and approval before implementation.
