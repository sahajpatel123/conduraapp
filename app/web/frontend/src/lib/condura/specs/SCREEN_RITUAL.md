# SCREEN_RITUAL.md — Constellation-as-Room · Screen Architecture

> **The contract.** The 9-step forced wizard in `Ritual.svelte` is being
> redesigned as a 2-screen first-run experience: **(1) Gate** (EULA + wax
> seal), **(2) Constellation** (the room with 6 live nodes). The wizard
> dissolves. The architecture is the skip.
>
> **Audience.** Next-session implementer. Designer for review.
>
> **Source-of-truth files** (read alongside this spec):
> - `app/web/frontend/src/lib/condura/Ritual.svelte` — current 1568-line implementation.
> - `app/web/frontend/src/lib/condura/DIRECTION.md` §2 — the call (Constellation-as-Room).
> - `app/web/frontend/src/lib/condura/APPFLOW.md` §2 — pre-ritual flow (now superseded).
> - `app/web/frontend/src/lib/condura/MOAT.md` — premium-quality rules.
> - `app/web/frontend/src/lib/condura/Shell.svelte:187–244` — mounts Ritual full-bleed.
> - `app/web/frontend/src/lib/condura/condura.css` — design tokens, motion grammar.

---

## 1. LAYOUT & CONTENT

### 1.1 Screen list

| # | Screen | When | Routes | Mount |
|---|--------|------|--------|-------|
| **1** | **Gate** (EULA + seal) | First run only | full-bleed pre-window | `firstRunStatus.complete === false` |
| **2** | **Constellation** (the room) | After seal stamps | full-bleed pre-window (still pre-shell) | `onComplete(dest)` after `onboarding.finish()` |

The previous 9 steps (`arrival → eula → permissions → power → hotkey → voice → channels → account → breath`) collapse to **two screens**. The Arrival cinematic is **not** part of the redesigned flow — the user's first visual moment is the Gate. EULA is the arrival signature.

### 1.2 Screen 1 · Gate

```
┌────────────────────────────────────────────────────────────────────────────┐
│ (centered, max-width 560px, vertically biased to 38% of viewport height)   │
│                                                                            │
│  — The terms                                                               │
│  First, the terms.                                                         │
│  Free for personal and commercial use, no tracking, no lock-in.            │
│  Read what that means — then stamp the seal.                               │
│                                                                            │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │ ▓ EULA scrollable well (paper-card, 280px max-height, 2px synapse    │  │
│  │ │ read-progress on left edge tracking scrollTop)                     │  │
│  │ │   [full Synaptic Freeware EULA v1 text, instrument serif 14px]    │  │
│  │ │                                                                   │  │
│  │ │   …                                                               │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                            │
│  ☐ I have read and accept the Condura Freeware EULA                        │
│                                                                            │
│       ┌──────┐  Stamp to accept / Scroll to the bottom first                │
│       │  C   │  a considered act — not a click                              │
│       │ ◯◯◯  │                                                            │
│       └──────┘                                                            │
│                                                                            │
│  not now · quit  (bottom-left italic skip-note, hard escape; #6.1 only)     │
└────────────────────────────────────────────────────────────────────────────┘
```

| Slot | Component | Content | Notes |
|------|-----------|---------|-------|
| Eyebrow | inline `text-caption` mono uppercase | `— The terms` | one per surface per MOAT §1.7 |
| Headline | `<RouteHero>` `text-h1` Instrument Serif | `First, the terms.` | serif 32px / --lh-h1 / --ls-h1 |
| Sub-copy | `<RouteHero>` `text-lead` Inter | `Free for personal and commercial use, no tracking, no lock-in. Read what that means — then stamp the seal.` | `max-width: 52ch` |
| EULA well | `.eula` recessed panel | inline license text + 2px synapse progress bar (left edge) | `max-height: 280px; overflow-y: auto;` |
| Checkbox + label | `.check` | `☐ I have read and accept the Condura Freeware EULA` | disabled until `eulaScrolled === true && eulaText` |
| Seal | `<button class="seal">` (existing, 64×64) | `C` glyph, radial gradient synapse | disabled until `canStamp === true` |
| Seal status | mono caption | `Stamp to accept` → `Accepted · thank you` | lingers 650ms post-stamp |
| Skip-note | `.skip-note` italic Instrument Serif (bottom-left, **positioned absolute, not a footer**) | `not now · quit →` | **single escape route**, dark-honest — calls `quitApp()` per Ritual.svelte:155–162 |
| Statusbar | (none on Gate; this is a focused screen) | — | — |
| Bottom-center pill | (none — only on Constellation) | — | — |

The seal is the only CTA on Gate. There is no "Continue →" button. Stamping it triggers `sealBloom` (radial 28px → 0px, 600ms `--ease`) → dissolves to Constellation. The user is not advancing — they are signing.

### 1.3 Screen 2 · Constellation

```
┌────────────────────────────────────────────────────────────────────────────┐
│  ─ First run (mono eyebrow, top-left, 24px from edge)                     │
│                                                                            │
│  Configure,                                                         ░░░░░░ │
│  don't comply.          ╱─[PERCEIVE]─╲               Channels voice        │
│                         │    ☉        │               threads,             │
│  A quiet, attentive    ◯─ Hotkey      ◯─ Power        summoned              │
│  presence.            │   Power       │   Voice        on your terms.      │
│                        ╲────Power─────╱                                    │
│                         Account  Hotkey  Voice                             │
│                                                                            │
│                       ───────────────────────                              │
│                                                                            │
│  ┌─ Hover preview strip (60px tall, fades in 220ms) ─────────────────────┐ │
│  │  "Press your combo to call Condura." + ⌘⇧Space keycap row            │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                            │
│             ●            ●            ●        ●       ●      ●           │
│          (6 live nodes sit on a soft dashed ring, 340px diameter)          │
│                                                                            │
│                          ┌─────────────────────────┐                      │
│                          │   Enter Condura →       │                      │
│                          └─────────────────────────┘                      │
│                                                                            │
└────────────────────────────────────────────────────────────────────────────┘
```

A right-side panel (380px wide) slides in on node-click, hosting the same step-specific UI from the current Ritual (radio cards for Power, keycap surface for Hotkey, channel chips for Channels, email field for Account, permission rows for Perceive, voice-card for Voice). **Same components, relocated.** The Constellation is the index; the panel is the detail.

### 1.4 The 6 nodes

Each node is a clickable surface on the ring. Default radius 6px. Wired nodes flip to `var(--synapse-light)` fill + `var(--synapse)` stroke. Skipped nodes render at `var(--paper)` fill + `var(--content-faint)` stroke. Hovered (non-active) nodes lift `-2px` + `box-shadow: 0 0 0 5px var(--pollen-halo)`.

| Node | Label | Glyph | Probe RPC | Skippable? | Default if skipped |
|---|---|---|---|---|---|
| **Perceive** | Perceive | `bolt` | `permissions.status` (live, 2s poll) | yes (chat still works) | computer use gated |
| **Power** | Power | `powersource` | `onboarding.probePower()` | yes | local Ollama if reachable |
| **Summon** | The hotkey | `summon` | — | **no (single soft-lock)** | **unset — locked decision #8** |
| **Voice** | Voice | `mic` | `onboarding.probeVoice()` | yes | off |
| **Threads** | Channels | `channels` | — | yes | none |
| **Account** | Account | `account` | `account.status` | yes | signed-out |

### 1.5 Per-node options table

| Node | Options presented in side panel | Default | Notes |
|---|---|---|---|
| **Perceive** | Two permission rows: Accessibility (glyph `bolt`), Screen Recording (glyph `audit`). Each row: `data-status` badge + `Open System Settings →` deep link (Wails `runtime.BrowserOpenURL`, falls back to clipboard + `window.open`). | none granted | 2s poll per Ritual.svelte:170–178 |
| **Power** | Three radio cards: (1) Local — Ollama (Recommended tag if reachable, meta lists models), (2) Paste an API key, (3) Connect a subscription. | `local` (when reachable) | Subscription paths land in `Settings → Power` per CLAUDE.md locked decision #22 + #30 |
| **Summon** | Keycap surface (dashed border → solid pollen + halo on recording) + 3 presets (`Option+Option`, `Cmd+Shift+Space`, `Ctrl+Space`) + `Try it` button. | none (no silent default) | The **only** soft-lock — pill reads `Set a hotkey to enter →` with a faint synapse halo until set |
| **Voice** | Voice card with mic name + meta + enable toggle + sensitivity slider. | off (toggle off) | Wake word config persists via `config.update` |
| **Threads** | 5 channel chips (Telegram ready, others dim with `v0.2.0` pill). | none | selected chips turn synapse-green |
| **Account** | Email field + `Send magic link →` button + `skip — I'll do this later →` link. | signed-out | `account.signInWithEmail()` writes to daemon |

### 1.6 The bottom-center `Enter Condura →` pill

| Spec | Value |
|---|---|
| Position | `position: absolute; bottom: 28px; left: 50%; transform: translateX(-50%);` |
| Style | `Button` `kind="primary"` (existing) with pollen fill, italic arrow |
| Default state | `enabled`. **Always.** Even with zero nodes wired. |
| Soft-lock state | When `Summon` is unset: same pill, label swaps to `Set a hotkey to enter →`, faint synapse halo pulses (`breathe 1.6s infinite`). One click on the Summon node + one keypress releases it. |
| onClick | `enterCondura()` → `onboarding.finish({...})` → `.dissolving` 700ms opacity→0 + blur 8px → Shell renders underneath |

**Why always enabled:** the goal is *configure, don't comply*. The user can enter Condura with nothing wired. Permissions, voice, channels, account — all can be deferred to Settings. The hotkey is the single mandatory choice because there is **no default** (locked decision #8). One soft-lock per surface, by design.

### 1.7 Hero moment — the wax seal

The wax seal on Gate is the single most premium micro-interaction in the product (DIRECTION.md §2 "Hero moment"). Stamping is the arrival signature. The signature motion: `sealBloom` keyframe (Ritual.svelte:981–984) — radial 28px → 0px opacity, 600ms `--ease`. The seal scales `0.94` on stamp + `translateY(2px)` settles it. Reserved for legal consent only.

The seal stamps **the moment the user accepts** — not as a generic confirm. The wax seal is not reused as a "submit" affordance anywhere else.

### 1.8 Hover preview strip

A 60px-tall strip appears below the constellation on node-hover:

| Node | Preview copy |
|---|---|
| **Perceive** | `Accessibility · Screen Recording. Condura reads only what it must to act.` |
| **Power** | `Local Ollama · an API key · a subscription. Your model, your call.` |
| **Summon** | `Press your combo to call Condura.` + a one-line keycap row |
| **Voice** | `"hey condura" — say the wake word; it listens.` |
| **Threads** | `Telegram is ready. The rest when you wire them.` |
| **Account** | `Optional. Sign in for skills sync, donations, support.` |

Animation: `opacity: 0 → 1` + `translateY(4px) → 0` over `--dur` (280ms), `--ease`, with an 80ms delay. Dismissed on `mouseleave` + focus-out.

### 1.9 Idle invitation (when something is unwired)

One slow pollen mote drifts from the center toward an unwired node every 14s, 1.4s duration, then dissolves. Suggests "tap me" without nagging. **Stops** once the user has wired at least one node.

---

## 2. STATE MATRIX

### 2.1 Gate screen states

| State | Trigger | Visual |
|---|---|---|
| `gate-unread` | mount, `eulaText` not loaded | placeholder: `Loading the license…` |
| `gate-read` | `eulaText` loaded + checkbox armed | seal pulses faintly (`breathe 1.6s`); checkbox `enabled` if scrolled |
| `gate-half-read` | `eulaText` loaded, not scrolled to 85% | seal `disabled`; sub-text reads `Scroll to the bottom first` |
| `gate-fallback` | `eulaIsFallback === true` | italic note above seal: `Read offline (daemon unreachable) — your acceptance will be replayed to Condura on next boot.` |
| `gate-stamped` | user clicked seal (`stamped === true`) | seal scales `0.94`, sealBloom fires, status lingers `Accepted · thank you` for 650ms, then dissolves to Constellation |
| `gate-error` | EULA load threw | italic + retry |

### 2.2 Per-node states (Constellation)

| State | Pulse color | Indicator label | Side panel content |
|---|---|---|---|
| `node-probe-loading` | `pollen` (drift) | `…` mono | `<Pulse phase="thinking" size={8} />` + mono label |
| `node-done` | `synapse` (steady) | `wired` mono | form completed, no further action |
| `node-skipped` | `faint` | `later` mono | panel shows empty state: `You can set this up in Settings → {section}.` |
| `node-error` | `danger` | `!` mono | `<Pulse phase="error" size={8} />` + retry button |
| `node-pending-action` | `warn` (breathing) | `awaiting` mono | (e.g., magic-link sent state) |

### 2.3 Constellation-level states

| State | Trigger | Behavior |
|---|---|---|
| `constellation-probe-in-flight` | `permissions.status` or `onboarding.probePower` / `probeVoice` still resolving | non-active nodes render at `data-status="probing"`; the active nodes (already probed) render fully |
| `constellation-all-done` | every probe resolved, regardless of skip/done | `Enter Condura →` enabled (subject to hotkey soft-lock) |
| `constellation-ready` | Summon set + at least one other node done | `Enter Condura →` enabled with no soft-lock |
| `constellation-error` | any probe failed | `.err-state` row on the failing node; pill stays enabled |

The Constellation renders usable even if **all** probes fail. The error appears per-node, not globally — the room is alive even with no furniture.

---

## 3. MOTION CHOREOGRAPHY

Per MOAT §2.3 + DIRECTION.md §5: one `@media (prefers-reduced-motion: reduce)` block in `condura.css` does the whole work. **Components never redeclare.** This section describes the gestures; the durations live in the global tokens.

### 3.1 Gate choreography

| Beat | What | Duration | Easing | Trigger |
|---|---|---|---|---|
| 0–140ms | Paper void (`:root`) fades in | `--dur-slow` (520ms) | `--ease` | mount |
| 140–340ms | Eyebrow fades + `translateY(4px) → 0` | `--dur` (280ms) | `--ease` | mount |
| 340–660ms | Headline `clip-path: inset(0 100% 0 0) → inset(0 0 0 0)` (wordReveal) | `--dur-cine` (900ms) | `--ease` | mount |
| 660–880ms | Sub-copy fade + lift | `--dur-slow` | `--ease` | mount |
| 880–1100ms | EULA well fade + 2px synapse progress bar `transform: scaleY(0) → 1` from `transform-origin: top` | `--dur-slow` | `--ease` | mount |
| 1100–1380ms | Checkbox + label fade | `--dur` | `--ease` | mount |
| 1380–1800ms | Seal `opacity: 0 → 1` + `transform: scale(0.7) → 1` (ease-pop) | `--dur-cine` | `--ease-pop` | mount |
| On scroll | Left-edge synapse progress bar `height` updates | 100ms linear | linear | `onscroll` (eulaEl) |
| On scroll-bottom | Checkbox armed | — | — | scrollTop reaches 85% |
| On checkbox-tick | Seal halo softens (synapse glow → pollen glow at 30% alpha) | `--dur-fast` | `--ease` | click |
| **On stamp (the hero)** | **Seal scales `0.94` + `translateY(2px)` settles + `sealBloom` radial 28px → 0px opacity, then `.dissolving` 700ms fade + 8px blur on the Gate wrapper** | 600ms (bloom) + 700ms (dissolve) | `--ease` | seal click |
| Post-stamp | Status lingers `Accepted · thank you` for 650ms, then Constellation mounts | — | — | t=600ms |
| Pre-stamp (hover) | Seal halo brightens (3px → 6px pollen) | `--dur` | `--ease` | hover |

### 3.2 Constellation choreography (mount)

| Beat | What | Duration | Easing |
|---|---|---|---|
| 0–140ms | Paper void fades in (carried from Gate dissolve) | `--dur-slow` | `--ease` |
| 140–660ms | Headline (left) + lead copy fade + lift | `--dur-cine` | `--ease` |
| 660–1140ms | Constellation ring renders: 6 nodes stagger in (`opacity: 0 → 1` + `scale: 0.7 → 1`) | per-node 320ms | `--ease-pop` |
|                                       | **80ms stagger between nodes** (counterclockwise from top) | — | — |
| 1140–1660ms | Thread-draw between wired nodes (left → right → each thread) | `--dur-slow` | `--ease` |
| 1660ms onward | `Enter Condura →` pill fades + rises 8px | `--dur-slow` | `--ease-pop` (once) |
| Idle | Slow pollen mote drifts every 14s toward unwired node | 1.4s | `linear` |

### 3.3 On node-hover (gesture-by-gesture)

| Gesture | Property | Duration | Easing | Trigger |
|---|---|---|---|---|
| **Hover lift (card)** | `transform: translateY(-2px)` + `--shadow-card` → `--shadow-float` + 5px pollen halo | `--dur` | `--ease` | mouseenter |
| **Hover preview strip** | `opacity: 0 → 1` + `translateY(4px) → 0`, 80ms delay | `--dur` | `--ease` | mouseenter |
| **Hover (active/probing node)** | breath-pulse: 1.6s `transform: scale(1 → 1.08)` + opacity `0.85 → 1` | 1.6s loop | `--ease` | always-on when `awaiting`/`listening` |
| **Hover (wired-done node)** | single pollen mote beats at node center, dissolves | 1.4s | `linear` | mouseenter, fires once |

### 3.4 On node-click → side panel slide-in

| Beat | What | Duration | Easing |
|---|---|---|---|
| 0ms | Side panel slides in from right (`translateX(24px) → 0` + `opacity: 0 → 1`) | `--dur-slow` | `--ease` |
| 80ms | Panel content staggers (eyebrow, headline, fields) — 60ms per row | `--dur` | `--ease` |
| On completion | Thread-draws at the panel's bottom edge over `--dur-slow` | `--dur-slow` | `--ease` |
| On panel-close | Reverse (`translateX(0 → 24px)` + `opacity: 1 → 0`) | `--dur` | `--ease-in` |
| Outside-click | Same as close | `--dur` | `--ease-in` |
| Esc | Same as close | `--dur` | `--ease-in` |

### 3.5 On Enter Condura →

| Beat | What | Duration | Easing |
|---|---|---|---|
| 0ms | Shell mounts underneath; Ritual wrapper gets `.dissolving` | — | — |
| 0–700ms | Dissolve: `opacity 1 → 0` + `filter: blur(8px)` | `--dur-cine` | `--ease` |
| 700ms onward | Shell `.route-container.route-enter` fires (per SCREEN_SHELL §4.1) | `--dur-slow` | `--ease` |
| → | Titlebar thread `rAF` resumes | — | — |

### 3.6 Reduced-motion table

All gestures reduce to `opacity: 0 → 1` over `--dur-fast` (140ms) or instant on user-triggered motions. The global block in `condura.css:469–476` owns the contract.

| Gesture | Reduced-motion behavior |
|---|---|
| Gate wordReveal + headline lift | `transform: none` (per the global block) |
| Constellation node stagger | all 6 nodes mount simultaneously; no scale, no opacity ramp |
| Seal `sealBloom` | skipped; stamp is instant; status text reads `Accepted` immediately |
| Hover-lift on cards | no transform; halo only |
| Side panel slide-in | opacity-only (no translate) |
| Dissolve to Shell | `opacity 1 → 0` only (no blur per the global) |
| Idle pollen mote | hidden via `.ambient-thread { display: none }` |
| Titlebar thread | static line, no bend |

---

## 4. KEYBOARD

### 4.1 Tab order — Gate

`Tab` from the document root (when Gate is the only mounted surface):

1. EULA well (receives focus, the well is `tabindex="0"` and `role="region"`).
2. Checkbox.
3. Seal.
4. Skip-note (`not now · quit`).
5. Cycle.

`Shift+Tab` reverses. `Esc` does nothing on Gate — the only escape is the skip-note which `quitApp()`s. This is by design: the legal step has no implicit dismissal.

### 4.2 Tab order — Constellation

Visual layout = DOM order = tab order. Left → right, top → bottom:

1. **Headline area** — no interactive controls.
2. **6 nodes in render order** (counterclockwise from top, top-to-bottom in the SVG's reading order):
   1. Perceive
   2. Power
   3. Summon
   4. Voice
   5. Threads
   6. Account
3. **Side panel content** (when a node is active) — focus trap inside the panel:
   - Perceive: row 1 (Accessibility deep link) → row 2 (Screen Recording deep link) → Continue.
   - Power: radio card 1 → radio card 2 → radio card 3 → Continue.
   - Summon: keycap surface (focus-trap-on-recording) → presets (3) → Try it → Continue.
   - Voice: enable toggle → sensitivity slider → Continue.
   - Threads: 5 chips → Continue.
   - Account: email input → Send magic link → skip-link.
4. **`Enter Condura →` pill**.

### 4.3 Focus traps

| Surface | Focus trapped? | Initial focus | Esc behavior |
|---|---|---|---|
| Gate (EULA well) | no | EULA well | nothing |
| Gate seal | no | seal on stamp | nothing |
| Side panel (any node) | **yes** | first focusable in panel content | close panel |
| Hotkey recording (inside Summon panel) | **yes** | keycap surface | cancel recording, surface returns to resting state (`recording = false`) |
| Constellation pill | no | pill on Enter-pressed + soft-lock cleared | nothing |

The hotkey recording focus-trap is the most important: while `recording === true`, all keys flow into the trap. The trap listens for `Escape` (cancel), `Enter`/`Space` (accept current combo), and any modifier+key combo (finalize). The trap releases when `recording === false`.

### 4.4 Arrow-key navigation between nodes

`←` / `→` move focus between adjacent nodes on the ring (counterclockwise / clockwise).
`↑` / `↓` move focus to the same node in the next / previous arc if the constellation is split into two arcs on smaller viewports.
`Home` / `End` jump to first / last.
`Space` / `Enter` activate the focused node (opens side panel).
`Esc` with a panel open closes the panel (returns focus to the focused node).

### 4.5 Constellation-pill soft-lock release

While the soft-lock is engaged (`Summon === false`):

1. `Tab` order skips the pill (or lands on it but it's `aria-disabled="true"`).
2. The pill announces via `aria-live="polite"`: `Set a hotkey to enter Condura.`
3. The Summon node has a visible pollen halo + `aria-describedby="soft-lock-help"`.

Release path: `Tab` to Summon → `Enter` (opens panel) → `Tab` to keycap → `Enter` (recording begins) → press a combo → `recording = false` → combo persists → focus returns to pill (now `aria-disabled="false"`).

---

## 5. COMPONENTS USED

Per MOAT §2.8: **three named primitives** (`.c-modal`, `.c-sheet`, `.c-popover`) own modal/sheet/popover. The redesigned Ritual uses only the `.c-sheet` primitive for the Constellation panels. Existing `ConsentModal.svelte` remains the gatekeeper modal — not relevant here.

| Component | Where used | Role in redesigned Ritual |
|-----------|------------|--------------------------|
| **`Gate.svelte`** | new — first-run legal screen | scrollable EULA well + wax seal stamp + checkbox + skip-note |
| **`WaxSeal.svelte`** | inside Gate (extracted from the inline `button.seal` in Ritual.svelte:513) | the 64×64 radial-gradient seal with `sealBloom` keyframe |
| **`Constellation.svelte`** | new — replaces the entire 9-step ritual after Gate | 6-node ring + threading + Enter pill + idle invitation |
| **`ConstellationNode.svelte`** | inside Constellation × 6 | clickable ring node; hover preview trigger; accepts probe state |
| **`HoverPreview.svelte`** | new — strip below the ring | 60px-tall copy strip; fades in 220ms with 80ms delay |
| **`SidePanel.svelte`** (.c-sheet) | inside Constellation × 1 | 380px slide-in panel hosting per-node detail UI |
| **`ThemePicker.svelte`** | (Phase 2 shipped) | mentioned for capability but **not** inside Constellation — moved to Settings per Phase-2 |
| **`KeyCapture.svelte`** | inside SidePanel: Summon | keycap surface + recording trap + 3 presets + Try-it |
| **`Thread.svelte`** | between wired nodes in Constellation; bottom edge of side panel; bottom spine | signature gesture, used in 3 places here |
| **`Pulse.svelte`** | every node (probe-loading + active-state heartbeat); pill welcome | load-time + heartbeat |
| **`Glyph.svelte`** | on each node; icons (bolt, audit, summon, mic, channels, account) | the icon set |
| **`Button.svelte`** | pill (`Enter Condura →`); Continue buttons inside side panels | primary CTA |
| **`Tooltip.svelte`** | soft-lock helper; skip-note (`title="Quit the app"`) | hover-delayed 400ms |

> Note: **no** `Spinner`, no `LoadingSpinner`, no `↻`. Loading renders a Pulse + mono label + a thread-draw (per MOAT §2.5, DIRECTION.md §5 "The Thread").

### 5.1 Components that are RESIGNED (from current Ritual)

| Old component / inline block | Status | Notes |
|---|---|---|
| `Ritual.svelte:354–360, 1480–1567` awakening keyframes (`voidHold`, `moteDrift`, `wordReveal`, `fadeUp`, `firstBeat`, `breathe`) | **collapsed to 2** per MOAT §1.1 + DIRECTION.md §2.1 (4-keyframe arrival `voidFade` + `wordReveal`) | Cinematic Arrival is **removed** — Gate IS the arrival. |
| `Ritual.svelte:1480–1567` awakening overlay (`.a-void`, `.a-mote`, `.a-wordmark`, `.a-underline`, `.a-pulse`) | **deleted** | No arrival cinematic. Gate's headline + seal serves the arrival signature. |
| `Ritual.svelte:413–442` (decorative constellation SVG with draw-on-wired / draw-on-skipped bezier paths) | **promoted** to live `<Constellation>` component per MOAT §1.6 option (a) | Now reflects actual system state — what is wired, what is skipped |
| `Ritual.svelte:155–162` `quitApp()` skip-note for `not now · quit` | **promoted** to Gate's skip-note | Still calls `window.close?.()`. The only escape. |
| `Ritual.svelte:631–637` `breath` step | **deleted** | The "Condura is here." hero + Enter button is now the Constellation's `Enter Condura →` pill. The dissolving fade lives at Constellation's exit, not its own step. |

---

## 6. DATA FETCHED

Per-existing Ritual.svelte:362–387 mount-time IPC; per node the probes are the same. **What changes is the order and the parallelism.**

### 6.1 Initial mount (Gate)

| Call | Source | Purpose | Failure |
|---|---|---|---|
| `initStores()` | `lib/stores/init.ts` | Boot every store (best-effort) | swallow per-store |
| `ipc.firstRunStatus()` | `lib/ipc/client.ts` | `{ complete: boolean }` — gate Ritual mount | if false, do nothing (we are Ritual) |
| `onboarding.sync()` | `lib/stores/onboarding.svelte.ts` | hydrate `eula`, `daemon.steps`, etc. | swallow (fallback) |
| `onboarding.loadEula()` | same | refresh canonical EULA text from daemon | fall back to `FALLBACK_EULA_TEXT` |
| `permissionsStatus()` poll starts | `lib/ipc/client.ts` | 2s polling for the duration of the Constellation | if unreachable, badge reads `unknown` |

### 6.2 Constellation probes (parallel)

| Probe | Source | When | Failure |
|---|---|---|---|
| `onboarding.probePower()` → `{ollama_reachable, ollama_models, recommended}` | `lib/ipc/client.ts` | mount | null → `Local` card shows `not detected — install Ollama, or pick another` |
| `onboarding.probeVoice()` → `{mic_available, ready, wake_word}` | same | mount | null → voice card shows `not detected`; toggle still toggleable |
| `permissionsStatus()` → `[{kind:'accessibility',status},...]` | same | mount + 2s poll | `unknown` badge + retry copy |
| `account.status` (signed-in?) | `lib/stores/account.svelte.ts` | mount | `Not signed in` copy |
| `channelsList()` (best-effort) | `channels` route store | lazy (on Threads panel open) | defaults render |

### 6.3 Per-node writes (on side-panel completion)

| Node | Write | Trigger |
|---|---|---|
| Perceive | `onboarding.completePermissions()` (when both `granted`) OR `onboarding.skipStep('permissions')` | Continue button |
| Power | local `powerChoice` ('local' \| 'apikey' \| 'sub') — no daemon write; user redirects to Settings on enter | radio card select |
| Summon | `onboarding.setHotkey(combo)` + `onboarding.saveHotkey()` | Continue button (combo non-empty) |
| Voice | local `voiceEnabled`; wake config persists via `config.update({wake:{enabled, sensitivity}})` | Enable toggle |
| Threads | local `channelPick` Set; **no daemon write at this stage** | Connect chip |
| Account | `account.signInWithEmail(email, locale, origin)` | Send magic link |

### 6.4 On Enter Condura →

```ts
onboarding.finish({
  hotkey: onboarding.daemon?.steps?.hotkey?.data ?? onboarding.hotkeyValue,
  eula_version: onboarding.eulaVersion ?? 'v1',
  permissions_skipped: onboarding.daemon?.steps?.permissions?.status === 'skipped',
});
```

The existing `enterCondura()` (Ritual.svelte:338–351) is reused; the `dest` arg is `'#/settings'` if `powerChoice === 'apikey' || 'sub'`, otherwise default (`undefined`).

The local `condura-ritual-seen` flag is set after `finish()` resolves (Shell polls for it on next mount).

### 6.5 What does NOT change in the IPC contract

- `onboarding.acceptEula(version)` — still called on stamp.
- `onboarding.completePermissions()` — still called on Permissions panel Continue.
- `onboarding.skipStep('hotkey')` — still called on skip (though skipping hotkey is the soft-lock — see §1.6).
- `onboarding.saveHotkey()`, `onboarding.setHotkey(combo)` — same.
- `onboarding.probePower()`, `onboarding.probeVoice()` — same.
- `onboarding.finish()` — same payload.

The redesign is a UI refactor. The IPC contract is preserved.

---

## 7. DESIGN DECISIONS

### 7.1 MOAT compliance

| MOAT Test | How this spec passes |
|---|---|
| **§I1 — Configure, not comply** | Constellation presents six independent choices (any order, any subset). No "I agree" wall beyond Gate (which is legal, not config). The hotkey soft-lock is the *only* gate; everything else can be deferred. |
| **§I2 — Smooth is honest** | Every animation carries meaning: seal-stamp = consent given; node stagger = room becoming legible; thread-draw between wired nodes = "this is connected"; pill pulse = "the door is open." No decorative loops except the idle pollen mote (which is restrained and bounded — 1.4s every 14s). |
| **§I3 — Local-first feels local** | No spinners. Every probe renders its existing state (or "not detected" fallback) instantly. The thread-draw loader is the only wait indicator. |
| **§I4 — Every state is reachable** | §2 lists gate states (6), node states (5), constellation states (4). Every state has defined visual. No dead walls. |
| **§I5 — The 7 invariants are visible** | Gate: legal consent happens first. About already renders the invariants (post-Ritual). ConsentModal (separate) precedes every physical action. The wax seal IS the consent stamp, fore-fronted. |

### 7.2 DIRECTION alignment

| DIRECTION.md rule | How this spec follows |
|---|---|
| **Constellation-as-Room** (§2) | The redesign is literal: 6 nodes, ring, side panels, hover preview, Enter pill. The wizard dissolves. |
| **Wax seal is the arrival signature** | The seal lives on Gate and ONLY on Gate. No other surface uses it (MOAT §4 anti-pattern #6 — "no fake enthusiasm"). |
| **Light is default** (DIRECTION §4) | All Gate + Constellation surfaces default to `--paper #F4EFE4`. Dark via `:root[data-mode='dark']` cascade. |
| **Motion grammar** (DIRECTION §5) | All durations use `--dur-fast`, `--dur`, `--dur-slow`, `--dur-cine`. All easings use `--ease`, `--ease-in`, `--ease-pop`. The seal uses `--ease-pop` once. The thread-draw uses `--ease` once per surface. |
| **Reduced-motion contract** (DIRECTION §5 + MOAT §2.3) | One block in `condura.css`. Components never redeclare. |

### 7.3 Configure-not-comply

**Principle.** Every node the user can wire, they can skip. Skipping a node must feel like *deferring*, not *failing*. The architecture is the skip.

- **No shame copy.** Default skip label: `later`. Never `skipped`. Never `incomplete`.
- **No red badges on skip.** Skipped nodes render at `--content-faint`, not `--danger`.
- **No progress bar counting.** The spine is decorative. The user sees the room, not the to-do list.
- **No `X of 6 complete` line.** The pill says `Enter Condura →` even when zero nodes are wired.

### 7.4 Wax seal commitment

The seal stamps **once per install**. It is not used anywhere else in the product. Not in ConsentModal (which uses a different consent gesture — `armor` rect + allow/deny), not in Settings, not as a general confirm button.

The seal is the *arrival* signature. The ConsentModal is the *action* signature. Two different gestures, two different surfaces, zero overlap.

### 7.5 What this spec does NOT do

| Anti-pattern | Where we avoid it |
|---|---|
| Gradient text on headlines (MOAT §4.1) | All headlines are `--content` or `--synapse` for the `.alive` accent word — flat color, no clip-text gradient. |
| Emoji as UI icons (MOAT §4.2) | Every node icon is `<Glyph>` from `icons.ts`. The seal letter `C` is text, not an emoji. |
| Rainbow accents (MOAT §4.3) | Brand has synapse + pollen; status has ok/warn/danger/info. No new colors. |
| `Welcome to the future` copy (MOAT §4.5) | Headlines: `First, the terms.` / `Configure, don't comply.` — declarative, not exclamatory. |
| Fake enthusiasm (MOAT §4.6) | No `Awesome!`, no `Great choice!`, no `You're all set!` toasts. The wax seal stamp IS the celebration; it doesn't need a toast. |
| Spinner loaders (MOAT §4.7) | Thread-draw + `<Pulse>` instead. |
| Rectangular focus outlines (MOAT §4.8) | Pollen halo + synapse inset on every focusable; pill-radius elements get synapse ring only. |
| Double shadows (MOAT §4.9) | One elevation token per surface (`--shadow-card` for nodes, `--shadow-float` for hover). |
| Decorative animation (MOAT §4.10) | Every animation in this spec answers "what does this communicate?" |
| 5+ keyframes for a wordmark (MOAT §1.1) | Gate uses 6 keyframes **across the whole screen** (one per element: paper fade, eyebrow, headline, EULA well, checkbox, seal). Not stacked on one word. |
| Multi-version `alive` spans (MOAT §1.7) | Headline's `.alive` accent (single italic synapse word on `Configure, don't comply.`) is the only `.alive` use on this surface. |

### 7.6 Success criterion (the MOAT test)

The redesigned Ritual passes when:

1. The wizard dissolves; the room arrives. (No 9-step forced sequence.)
2. The user can read the seal as the arrival signature. (Unique, never reused.)
3. The user can wire 0, 1, 2, or all 6 nodes — in any order — and enter Condura with the rest deferred to Settings.
4. The Constellation surfaces real system state (what is granted, what is wired) — not a decorative timeline.
5. The hotkey is the only soft-lock; everything else deferrable.
6. A blind user using only `Tab` + `Space` + `Esc` + arrow keys can complete the first-run.
7. The reduced-motion user has the same flow with no animation drama.

If any of those is false, the spec is wrong; fix the spec, then fix the code, in the same commit.

---

## 8. DRIFT TABLE

What is **removed** from the current 9-step Ritual (Ritual.svelte:1–1568) and what is **added** by this redesign.

### 8.1 Removed

| Old | Where (line range) | New home | Reason |
|---|---|---|---|
| `StepId.arrival` + awakening keyframes (6) | `Ritual.svelte:488–492, 1480–1567` | **deleted** | Gate IS the arrival (DIRECTION §2). 4 keyframes collapsed to 2. |
| `a-void`, `a-mote`, `a-wordmark`, `a-underline`, `a-pulse` overlays | `Ritual.svelte:644–654` | **deleted** | no awakening cinematic; Gate opens with paper background already warm |
| `voidHold`, `moteDrift`, `wordReveal`-arrival, `fadeUp`, `firstBeat` (5 of the 6 awakening keyframes) | `Ritual.svelte:1550–1564` | **collapsed** | `wordReveal` reused for Gate headline clip-path; `breathe` reused for soft-lock halo |
| `StepId.permissions` as a discrete step | `Ritual.svelte:516–538` | **promoted** to Constellation's `Perceive` node side panel | configure-don't-comply |
| `StepId.power` as a discrete step | `Ritual.svelte:539–558` | **promoted** to Constellation's `Power` node side panel | same |
| `StepId.hotkey` as a discrete step | `Ritual.svelte:559–576` | **promoted** to Constellation's `Summon` node side panel (with soft-lock) | same |
| `StepId.voice` as a discrete step | `Ritual.svelte:577–595` | **promoted** to Constellation's `Voice` node side panel | same |
| `StepId.channels` as a discrete step | `Ritual.svelte:596–609` | **promoted** to Constellation's `Threads` node side panel | same |
| `StepId.account` as a discrete step | `Ritual.svelte:610–630` | **promoted** to Constellation's `Account` node side panel | same |
| `StepId.breath` (`Condura is here.` + `Enter Condura` button) | `Ritual.svelte:631–637` | **collapsed** to Constellation's `Enter Condura →` pill + dissolve-to-Shell | the closing IS the constellation exit |
| `.spine` (the bottom progress bar tracking `stepIndex`) | `Ritual.svelte:445–471, 739–751` | **deleted** | no step sequence; the spine is decorative | configure-don't-comply |
| `.island` (top-center step-label pill) | `Ritual.svelte:407–411, 692–717` | **deleted** | no steps to label | same |
| Decorative constellation SVG (bezier paths + circles, draw-on-wired/skipped) | `Ritual.svelte:413–442` | **promoted** to live `<Constellation>` component | MOAT §1.6(a) — make it teach the actual system state |
| Per-step bottom-left skip-notes (5 distinct: `quit`, `not now`, `set up later`, `leave this for later`) | `Ritual.svelte:474–482` | **collapsed** to **1** (the Gate's `not now · quit`) | all other nodes skip-by-architecture — they don't need a skip-note |
| Per-step `Continue →` CTAs (5: permissions, power, hotkey, voice, channels) | scattered | **collapsed** to the Constellation's `Enter Condura →` pill | one door, not five |
| `enterCondura()` → `dest` routing logic (Shell routes to `#/settings` when `powerChoice === 'apikey' \|\| 'sub'`) | `Ritual.svelte:347–349` | **kept** | re-use inside Constellation |

### 8.2 Added

| New | Lives where | Why |
|---|---|---|
| **`Gate.svelte`** | new file `app/web/frontend/src/lib/condura/Gate.svelte` | the redesigned first-run legal screen — replaces the inline Ritual.svelte:493–515 |
| **`WaxSeal.svelte`** | new — extracted from inline `<button class="seal">` | the 64×64 radial-gradient seal — promoted because it's a first-class component now |
| **`Constellation.svelte`** | new — replaces the entire 9-step ritual | the room — 6 nodes + ring + Enter pill |
| **`ConstellationNode.svelte`** | new × 6 instances | per-node clickable surface + status indicator + side-panel trigger |
| **`HoverPreview.svelte`** | new — strip below the ring | Arc-style hover preview, condensed (DIRECTION §2.3) |
| **`SidePanel.svelte`** (`.c-sheet` primitive wrapper) | new × 1 instance | 380px slide-in panel hosting per-node detail |
| **Idle invitation** (pollen mote drift every 14s) | inside Constellation | suggests "tap me" without nagging — DIRECTION §2.3 |
| **Soft-lock halo on Enter pill** | conditional styling | the only "you must do this" signal; one soft-lock, by design |
| **`HoverPreview.svelte`** + `KeyCapture.svelte` upgrade | the keycap surface uses the existing dashed-border recording UI | componentized for reuse |

### 8.3 Verbatim from current Ritual (kept as-is)

| Block | Line | Why |
|---|---|---|
| `permissionsStatus()` polling + 2s `setInterval` | `Ritual.svelte:170–178, 380–386` | working well; lift into Constellation's Perceive panel |
| `onRecordKey` (keycap recording) | `Ritual.svelte:234–255` | working perfectly; move into `KeyCapture.svelte` |
| 3 presets (`Option+Option`, `Cmd+Shift+Space`, `Ctrl+Space`) | `Ritual.svelte:228` | no silent default; these are the three honest choices |
| EULA `recomputeEulaScroll()` (long-text 85% threshold + short-text auto-resolve) | `Ritual.svelte:101–125` | do not break the user with a one-screen license that won't scroll |
| `sealBloom` keyframe (radial 28px → 0px opacity) | `Ritual.svelte:981–984` | the hero signature |
| Fallback EULA handling (`fallbackEula.ts`) | `Ritual.svelte:79–99, 136–147` | daemon-unreachable must not block the seal |
| `.err-state` italic Instrument Serif per-node error pattern | `Ritual.svelte:1441–1470` | the channel-error / voice-error / permission-error blocks; promote into `ErrorState.svelte` per MOAT §2.6 |
| `onboarding.completePermissions()` + `skipStep()` calls | `Ritual.svelte:206–215` | same |
| `account.signInWithEmail(email, locale, origin)` call | `Ritual.svelte:317–331` | same |
| `wired: Set<string>` + `skipped: Set<StepId>` state | `Ritual.svelte:56–57` | state model is fine — only the surface changes |

### 8.4 Net line-count delta (estimate)

| Component | Before (lines) | After (lines) | Delta |
|---|---|---|---|
| `Ritual.svelte` | ~1568 | **~150** (just `completeOnboarding()` + dissolve wrapper) | **−1418** |
| `Gate.svelte` (new) | — | ~280 | +280 |
| `Constellation.svelte` (new) | — | ~520 | +520 |
| `ConstellationNode.svelte` (new ×6) | — | ~80 each → ~480 (or 1 file with 6 instances) | +480 (or +80) |
| `SidePanel.svelte` (new) | — | ~360 | +360 |
| `HoverPreview.svelte` (new) | — | ~80 | +80 |
| `WaxSeal.svelte` (extracted, new) | — | ~120 | +120 |
| `KeyCapture.svelte` (extracted, new) | — | ~140 | +140 |
| **Total** | 1568 | ~2100 (or ~1700 if nodes reuse one file) | **+532 to +2060** |

The redesign **trades a monolithic 1568-line file for 8 focused files.** Each new file is <600 lines. Each is independently testable and editable. The win is not fewer lines; it's cohesion — **one component, one concern.**

### 8.5 Sequencing for the next session

If implementing in one pass:

1. **Extract `WaxSeal.svelte` from Ritual.svelte:931–984.** (~120 lines; no behavior change.)
2. **Build `Gate.svelte`.** Reuse the EULA well verbatim (Ritual.svelte:497–515). Implement 6-keyframe Gate entrance. Add `sealBloom` at end.
3. **Build `Constellation.svelte` + `ConstellationNode.svelte` (×6 instances from 1 file).** Thread-draw between wired nodes. Idle invitation. Enter pill.
4. **Build `HoverPreview.svelte`.** One component, content slot per node ID.
5. **Build `SidePanel.svelte` (`.c-sheet` wrapper).** Per-node content slotted.
6. **Extract `KeyCapture.svelte` from Ritual.svelte:559–576, 1159–1262.**
7. **Mount Gate + Constellation from `Shell.svelte:187–244`** (the existing Ritual mount point).
8. **Delete the obsolete blocks** from `Ritual.svelte` (per §8.1).
9. **Test** — `prefers-reduced-motion` path, hotkey soft-lock, all 6 nodes, reduced-still-usable when all probes fail.
10. **Update `APPFLOW.md` §2** to reflect the new pre-Ritual flow (Gate → Constellation → Shell). The current 9-step table is replaced by §1 of this spec.

If implementing in two passes: pass 1 is (1) + (2) (Gate ships independently); pass 2 is (3) → (10) (Constellation + cleanup).

---

**This document is the architecture. The code is the implementation. They agree. When they diverge, the divergence is the spec-bug — fix the doc, then fix the code, in one commit.**
