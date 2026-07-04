# Synaptic — Design System v1 (Redesign)

> **Status:** LOCKED. This is the synthesis of five parallel direction agents (brand, UX, motion, tokens, visual language) reconciled into a single coherent spec. Every implementation decision traces back here. If something needs to change, update this document first, then code.
>
> **Generated:** 2026-06-30, after one design-direction phase.
> **Source documents:** five agent outputs in `/private/tmp/claude-501/.../tasks/*.output` (creative, ux-engineer, animate-engineer, tokens, style-engineer). All preserved for reference.
>
> **Hard rule:** no new colors, no new fonts, no new durations, no new easing curves in implementation without updating this doc. The design system is a contract.

---

## 1. Brand Soul (one-paragraph anchor)

Synaptic is a **quiet, attentive presence** — the friend who notices your screen is stuck before you do. Its personality is *competent introvert*: deep capability, surface stillness, zero performance. The product feels like a perfectly organized workshop, not a fireworks display.

**The signature visual element:** a small, still, dim pulse — a single low-intensity dot, breathing at ~12 cycles per minute (5s period), that brightens imperceptibly when the agent is thinking and dims when it returns to listening. It is not a logo. It is a vital sign.

**The emotion on first open:** relief. Not awe. Not "wow." Relief — "it's already listening, and it didn't make me do anything yet."

---

## 2. Resolved Tensions (the hard decisions)

Five agents wrote from different altitudes; they disagreed in places. Every disagreement is resolved here.

| Tension | Resolution | Why |
|---|---|---|
| Style says "compact," Creative says "spacious" | **Dual-mode density.** Sidebar/lists/audit log = compact (Linear-grade). Chat reading surface / first-run / settings reading = spacious (Things-grade). The command overlay is medium — generous, but fits in 560px. | Both were right at their own altitude. The conflict was a failure to specify *where* each applied. |
| Style says "tinted neutrals + status colors," Creative says "one plum accent" | **Paper-warm cream + ink-cool near-black + electric plum as the SOLE brand accent.** Status semantics appear in *shape and motion*, not hue, in chrome. Plum appears only in: the pulse, the moment of permission is requested, and the trailing edge of user-caused animations. Destructive actions don't turn red — they slow down and stop. | The plum is the brand. Status colors compete with it. |
| Style says "variable sans + mono," Creative says "serif for agent voice + sans for chrome" | **Three families: humanist sans (UI), reading serif (agent voice), mono (data).** All three variable, all loaded `font-display: swap`. | The agent needs a different *voice* from the UI chrome. Serif-on-sans is the editorial convention for "this is a person talking, not a system printing." |
| Style says "hairline borders," Creative says "no border on agent surface" | **Hairline seams on all chrome. No border on the floating command overlay (which is glass).** A hairline seam *is* the border — it's just 1px instead of a frame. | Creative's intent was "no chrome frame around the floating surface." Honored by using glass + hairline = effectively borderless. |
| Style says "no shadows," motion uses scan-lines for kill switch | **No shadows for normal hierarchy. Single permitted exception:** the kill-switch red scan-line (per motion agent §3.10) is not a shadow but a state signal — it draws a 1px line across the screen as the agent freezes. This is exception, not pattern. | The kill switch is a safety contract; speed is part of the contract. |
| UX says "4 first-run screens," Creative says "5 including First Breath" | **UX's 4 screens + Creative's "First Breath" closing moment.** Power source is the real mandatory screen (UX). Hotkey is the only one with no skip (UX). "First Breath" is a 400ms dissolve into the floating surface with the pulse, not a separate screen. | Creative's First Breath is the transition *out* of the wizard, not a fifth wizard screen. |
| UX says "Power source is mandatory," Creative shows only 4 permission screens | **Both. The Power source screen is the real mandatory moment — without a brain, the agent can't act.** The two macOS permissions (Accessibility + Screen Recording) are presented as a single screen with two side-by-side panels (Creative's "Eyes"). | UX got the emotional priority right; Creative got the visual presentation right. |
| Motion says "instant / quick / base / emphasized / epic," Creative wants "breath" | **All five motion tiers stay. The "breath" is implemented as the ambient layer at 12 cycles per minute (5s period), expressed as a 0.3% luminance drift on the deepest background layer.** Breath is `motion-duration-slow` (320ms per cycle is too fast — use a 5s sine wave), not a separate timing system. | Motion's tier scale accommodates breath. Breath is one usage of the slowest tier, expressed as a continuous wave rather than a discrete transition. |

---

## 3. The Three Type Families (locked)

| Role | Family | Why | Fallback chain |
|---|---|---|---|
| **Sans (UI chrome)** | **IBM Plex Sans** | Geometric humanist, tool-feeling, NOT Inter/Geist/Manrope. Open source via Google Fonts. Variable font with weight axis. | `-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif` |
| **Serif (agent voice)** | **Source Serif 4** | Adobe's reading serif with old-style figures, real italics, variable weight. Set at body size, not display. | `Georgia, "Times New Roman", serif` |
| **Mono (data, code, IDs)** | **IBM Plex Mono** | Pairs with Plex Sans (same designer, same vertical metrics). Tabular figures throughout. | `"SF Mono", Menlo, Consolas, monospace` |

**Rules:**
- All three loaded via `<link>` with `font-display: swap` and `preload` (only for the first 80KB of each).
- Body text never goes below 13px effective size. Accessibility floor.
- Tabular numerals everywhere data appears (`font-variant-numeric: tabular-nums` on `.mono`, `.timestamp`, `.data`).
- Italics are real italics, never oblique-faked.
- Long-form reading measure: 64–72 characters (`max-width: 38em` on prose containers).

---

## 4. The Color System (locked, with hex)

### 4.1 Light mode (the hero, default on first launch)

**Base surfaces (paper-warm):**
- `--paper-warm-0`: `#FBF8F2` (page background, "paper")
- `--paper-warm-50`: `#F5F1E8` (sunken)
- `--paper-warm-100`: `#FAF6EE` (raised)

**Ink (text, dark UI elements):**
- `--ink-cool-900`: `#0E1014` (primary text, near-black with cool undertone — never pure #000)
- `--ink-cool-700`: `#2A2D36` (secondary text)
- `--ink-cool-500`: `#5A5E6B` (tertiary text, placeholders)
- `--ink-cool-300`: `#9DA1AD` (muted, disabled)
- `--ink-cool-100`: `#D4D7DD` (borders-subtle)
- `--ink-cool-50`: `#E8EAEF` (borders-default)

**Brand accent — Electric Plum (the ONE accent):**
- `--plum-50`: `#F2EDFF`
- `--plum-100`: `#E0D5FF`
- `--plum-300`: `#B89AFF`
- `--plum-500`: `#8A66FF`
- `--plum-600`: `#6E3AFF` ← **canonical brand accent**
- `--plum-700`: `#5A2BD9`
- `--plum-900`: `#321B7A`

**Status semantics (used sparingly, mostly as shape/icons):**
- `--success-500`: `#2E7D5B` (muted green, desaturated — never bright)
- `--warning-500`: `#A06A1F` (muted amber)
- `--error-500`: `#A83232` (muted red, used SO rarely that when it appears, it matters)
- `--info-500`: `#3A5A8C` (muted slate-blue)

### 4.2 Dark mode (sibling, not default)

Dark is **not** inverted light. It's a separately-tuned scale optimized for low-light perception:

- `--paper-warm-0-dark`: `#0E1014` (page bg, ink-cool near-black)
- `--paper-warm-50-dark`: `#161922` (sunken)
- `--paper-warm-100-dark`: `#1C1F2A` (raised)
- `--ink-cool-900-dark`: `#F0EDE5` (primary text — warm off-white, not pure #FFF)
- Brand plum scale: **identical** across modes (plum-600 stays plum-600). Only the semantic mapping shifts.

### 4.3 High-contrast mode (a11y)

A separate scale with WCAG 2.2 AAA compliance (7:1 body text):
- Background: `#FFFFFF` / `#000000`
- Text: pure opposites
- Plum: `#4D1FCC` (deeper, more saturated for visibility)
- All borders: 2px instead of 1px
- All focus rings: 3px

### 4.4 Color rules

- **Plum appears in at most 5% of any given screen.** Reserved for: the pulse, the moment a permission is requested, the trailing edge of user-caused animations, and one moment of emphasis per screen (max).
- **Status colors only for state, never for emphasis.** A bold span in body copy does not turn blue.
- **No background gradients except** the kill-switch scan-line overlay and the floating command overlay's glass (both specified exceptions).
- **No glow, no neon, no outer shadow with color.**

---

## 5. Spacing, Radius, Shadow, Z-Index (locked)

### 5.1 Spacing scale (base 4px)

```
--space-0:    0
--space-1:    4px
--space-2:    8px
--space-3:    12px
--space-4:    16px
--space-5:    20px
--space-6:    24px
--space-7:    32px
--space-8:    40px
--space-9:    48px
--space-10:   64px
--space-11:   80px
--space-12:   96px
--space-13:   128px
```

### 5.2 Radius

```
--radius-xs:    2px   (icon strokes, micro)
--radius-sm:    4px   (chips, small buttons)
--radius-md:    8px   (buttons, inputs, cards)
--radius-lg:    12px  (large cards, popovers)
--radius-xl:    20px  (modals, command overlay)
--radius-pill:  9999px (tags, status dots)
```

### 5.3 Shadow (used minimally — most hierarchy is hairline + tone)

```
--shadow-0:    none
--shadow-1:    0 1px 2px rgba(14, 16, 20, 0.04)   (cards at rest — often replaced by hairline)
--shadow-2:    0 2px 8px rgba(14, 16, 20, 0.06)   (hover state)
--shadow-3:    0 12px 32px rgba(14, 16, 20, 0.10) (popovers, command overlay)
--shadow-4:    0 24px 64px rgba(14, 16, 20, 0.14) (modals)
--shadow-inner: inset 0 1px 2px rgba(14, 16, 20, 0.06) (input wells)
```

Dark mode: shadow opacity halved, hairline emphasis increased.

### 5.4 Blur (only on the command overlay)

```
--blur-sm:   4px
--blur-md:   8px
--blur-lg:   16px
--blur-xl:   24px
--blur-2xl:  40px   (command overlay backdrop)
```

### 5.5 Border widths

```
--border-hairline: 1px
--border-default:  1px
--border-strong:   2px
--border-focus:    2px   (always 2px for visibility)
```

### 5.6 Z-index ladder

```
--z-base:      0
--z-raised:    10     (sticky headers)
--z-sticky:    100    (persistent bars)
--z-overlay:   1000   (floating command surface)
--z-modal:     2000   (dialogs)
--z-toast:     3000   (notifications)
--z-tooltip:   4000   (ephemeral hints)
--z-max:       9999   (debug-only escape hatch)
```

**No arbitrary z-index values anywhere in component CSS.** If a component needs a new layer, the token ladder is wrong — update it here.

---

## 6. Motion Tokens (locked, from animate-engineer)

### 6.1 Durations

```
--duration-instant:    50ms    (opacity-only state flips, cursor color)
--duration-fast:       120ms   (hovers, button press, tooltip)
--duration-base:       220ms   (standard transitions, panel open/close, message arrival)
--duration-slow:       320ms   (panel expansions, ambient breath per cycle)
--duration-emphasized: 420ms   (overlay appear, consent modal)
--duration-epic:       800ms   (app first paint, daemon handshake — once per session)
```

### 6.2 Easings

```
--ease-standard:     cubic-bezier(0.2, 0, 0, 1)        (default)
--ease-decelerate:   cubic-bezier(0, 0, 0.2, 1)        (elements entering)
--ease-accelerate:   cubic-bezier(0.4, 0, 1, 1)        (elements leaving)
--ease-emphasized:   cubic-bezier(0.3, 0, 0, 1)        (consent moments)
--spring-soft:       stiffness 180, damping 24         (message arrival)
--spring-medium:     stiffness 240, damping 28         (panel slide)
--spring-snappy:     stiffness 320, damping 32         (cursor follow)
--spring-gentle:     stiffness 140, damping 20         (ambient breath)
```

### 6.3 Distances

```
--distance-micro:   4px    (hover lift, focus ring)
--distance-near:    12px   (tooltip shift, indent)
--distance-base:    8px    (default translate)
--distance-far:     32px   (panel slide, modal rise)
--distance-viewport: vh/vw (full-screen transitions)
```

### 6.4 Stagger

```
--stagger-fast:      30ms    (list items, identical)
--stagger-base:      60ms    (entering groups, mixed)
--stagger-slow:      120ms   (settings sections)
--stagger-deliberate: 180ms  (consent moments, kill switch)
```

### 6.5 The Pulse (signature vital sign)

- **Period:** 5000ms (12 cycles/min)
- **Opacity:** 0.85 → 1.0 → 0.85
- **Scale:** 0.98 → 1.02 → 0.98
- **Color:** `--plum-600` in light mode, `--plum-500` in dark
- **State shifts:**
  - Idle: as above
  - Thinking: 7500ms period (8 cycles/min), opacity 0.7 → 1.0
  - Awaiting user: 3000ms period (20 cycles/min), opacity 1.0
  - Error: one-shot flash to `--error-500`, return to idle

### 6.6 Performance budget

Three energy modes that preserve identity but trim flourishes:
- **High** (plugged in): full motion grammar
- **Balanced** (default): ambient breath at 50% amplitude, luminance drift paused, stagger halved
- **Low** (battery): no breath, no drift, easings replaced with linear in non-critical transitions

**Never reduced, even in Low mode:** kill switch speed, consent moment duration, streaming text reveal, agent presence pulse.

---

## 7. Iconography (locked)

- **Style:** 1.25px line, perfectly geometric, slightly rounded line joins (1.5–2px). Never filled. Never duotone.
- **Stroke width uniform across the entire set** — visual rhythm stays even.
- **Sizes:** 16px in chrome, 20px in command overlay, 24px in empty states. Optical sizing: 16px icons next to 14px text usually render at 16px to match perceived weight.
- **Padding:** 1px breathing room on every side, no exceptions.
- **State changes are subtle:** chevron rotates 90°, circle fills 20% when active. No shape replacement.
- **The "summarize this document" icon:** rectangle (page) + 3 horizontal lines inside + small downward arrow on right edge pointing to a shorter 2-line stack below. No AI sparkle. No robot.

---

## 8. The Five Surfaces (locked architecture)

| # | Surface | Where | Density | Glass? | Critical element |
|---|---|---|---|---|---|
| 1 | **Chat surface** (center of gravity) | Main window, full height | Spacious (64–72ch measure, 1.55 leading) | No | Editorial column, mono timestamps on left margin, serif for agent voice, sans for UI chrome, hairline separators |
| 2 | **Command surface** (floating, primary interaction) | Cursor-anchored, 560px wide | Medium (generous but fits 560px) | **YES — only place glass is used** | Contextual strip + omni-bar + pulse |
| 3 | **Control surface** (settings, audit, accounts) | Full window pane | Compact lists + comfortable reading | No | Section nav (mono labels) + content (sentence per setting) |
| 4 | **Agent surface** (live action log, replay) | Reachable from chat/settings | Dense table (Linear-grade) | No | Mono timestamp + action chip + target + decision + verification |
| 5 | **Ambient surface** (menu bar / tray / status) | OS chrome | N/A | N/A | Single glyph + status dot + count badge |

**Material mapping:**
- Chat surface → paper-warm-0 bg, sans chrome, serif agent voice, mono timestamps
- Command surface → blur-2xl backdrop, paper-warm-100 panel, plum accent for pulse
- Control surface → paper-warm-0 bg, mono labels in section nav, comfortable reading width
- Agent surface → paper-warm-0 bg, mono timestamps, dense rows
- Ambient surface → OS-native menu bar icon (16×16), status dot overlay

---

## 9. The Floating Command Surface (the heart)

### 9.1 Architecture

**Layered omni-bar (Option C from creative agent).** Three components stacked:
1. **Contextual strip** (44px tall): shows what the agent noticed on screen (a focused window, a selected text, a file). Empty state = visible placeholder row with pulse on left edge.
2. **Omni-bar input** (64px tall): single serif text field, cursor blinking, placeholder "What would you like me to do?"
3. **Hint row** (16px tall, very small): keyboard shortcuts (⌘↵ send · esc dismiss · ⌘K everything).

### 9.2 Four states

**Idle / Empty (560×~140px):**
- Contextual strip: empty placeholder with pulse
- Input field: serif, placeholder visible
- Hint row visible

**Active / Typing (560×~280px):**
- Serif grows slightly (16px → 18px) after 8+ characters typed (subtle cue the agent is reading carefully)
- Below input: ranked interpretation cards (up to 5, 48px tall each, serif interpretation + sans steps preview)
- Highlighted card has a single plum hairline on left edge
- User presses ↓ to scroll, ↵ to confirm

**Processing (560×~180px):**
- Input collapses to single line
- Below: progress strip with three animating dots + live token stream from model
- "About to:" preview line: "About to: click 'Send' in Gmail" — user can press Esc to abort
- ⏸ Pause button at right

**Result (560×~240px, stays 8s or until user moves away):**
- "✓ Done" pill (top right)
- One-line summary
- "Show details" expander
- Three buttons: ↻ Undo · 📌 Pin · ⌘↩ Send to chat

### 9.3 Position & anchoring

- Anchored to cursor (within 80px), with 12px gap and a small upward bias
- Fallback: if cursor in top-left 100×100px quadrant, anchor to bottom-center
- Multi-monitor: appears on display where cursor is
- Never covers user's active selection or current caret (avoids 60×24px zone around caret)

### 9.4 Animation in/out

**In:** scales 96% → 100%, fades 0 → 1, 180ms ease-out. Settles, doesn't slide. Pulse already breathing on arrival. **Target latency: <100ms from hotkey to visible.**

**Out (dismiss):** scales 100% → 97%, fades to 0, 140ms ease-in.

**Out (submit):** collapses downward 200ms, pulse briefly accelerates (one beat at 18 cycles/min) before returning to rest.

### 9.5 Triggers (priority order)

1. **Hotkey** (user-set from first-run) — appears within 100ms
2. **Menu bar click** — always works
3. **Drag-drop** file/URL/text onto menu bar icon — pre-fills "What should I do with this?"
4. **Voice wake** "hey synaptic" (if enabled) — voice feedback "Yes?" then appears

---

## 10. First-Run Wizard (4 screens + 1 closing moment)

### 10.1 Screen 1 — Invitation

- **Mood:** A door, opening inward.
- **Content:** A single sentence in serif, centered: *"I'd like to help you use your computer. Before I do, I need to ask your permission three times."* The word "three" is the only plum accent.
- **Button:** "Begin." (no logo, no illustration, no marketing copy)
- **Skip:** No skip. The first screen is the relationship.

### 10.2 Screen 2 — EULA

- **Mood:** A contract between adults.
- **Content:** The license scrolls inside the agent surface itself, serif at body size.
- **Accept button:** Disabled until scrolled to bottom (standard pattern, presented as a small inline moment).
- **On accept:** 1-second message: *"Thank you. Now I can read these words aloud if you'd like."* (moment of personality)

### 10.3 Screen 3 — Eyes

- **Mood:** Asking to see.
- **Layout:** Two side-by-side panels in the agent surface.
  - **Left:** Accessibility — diagram showing named buttons, window titles, text fields the agent will perceive. Button: "Grant on this Mac" (opens System Settings to exact pane). Live status dot.
  - **Right:** Screen Recording — diagram showing screen as whole, sampled rarely, never recorded. Button: "Grant on this Mac". Live status dot.
- **Footer (smallest type):** *"You can revoke either of these at any time. I will stop the moment you do."*
- **Skip:** "Limited mode" — agent works for chat, file reading, web search. Chip shows "Limited mode" indicator.

### 10.4 Screen 4 — Power source (the *real* mandatory moment)

- **Mood:** A workshop, picking your tools.
- **Content:** Four cards in a 2×2 grid:
  - **Card A:** "Use my Claude Pro" — OAuth, one click. Cost note: "Uses your existing subscription."
  - **Card B:** "Use my ChatGPT Plus" — OAuth, one click.
  - **Card C:** "Paste an API key" — text field with paste-detect.
  - **Card D:** "Use a local model" — auto-detects Ollama at localhost:11434.
- **Skip:** "I'll set this up later in Settings" link at bottom. If skipped: auto-enables Ollama if present, otherwise runs in read-only demo mode.
- **This is the real mandatory moment.** Without a brain, the agent can't act.

### 10.5 Screen 5 — Key (hotkey recording)

- **Mood:** Teaching a friend your secret handshake.
- **Layout:** Large central recordable field. Three suggested combos as ghosted text: ⌥⌥, ⌘⇧Space, ^Space.
- **On valid combo recorded:** Surface shifts (no "next" button), user is in the app.
- **Below:** Voice test toggle: "Want to also say 'hey synaptic' to summon me?" on/off.
- **Skip:** None. Without a hotkey, the agent is unreachable.

### 10.6 Closing moment — First Breath

- **Timing:** Immediately after hotkey is recorded.
- **Motion:** Onboarding dissolves over 400ms. In its place: the floating command surface, empty, with the pulse at center, breathing.
- **First text:** Single serif line fades in over 1s: *"I'm here. Type when you're ready."* Then fades to 60% opacity after 4s.

---

## 11. The Three Day-1 Screens

### 11.1 Screen 1 — The Surface (home)

The floating command surface, alone on the desktop. No sidebar in this view. The user opens, types/selects, agent acts, surface dismisses or shows result.

### 11.2 Screen 2 — The Conversation Drawer (history)

- **Access:** Two-finger trackpad swipe right, or ⌘K → "Show history", or surface expand button.
- **Layout:** Left-edge drawer, 320px wide.
- **Rows:** Date in sans, first sentence of user's request in serif, single small plum dot if agent acted (completion indicator).
- **Search:** Top-of-drawer field in serif with faint plum underline. Real-time filter, 40ms stagger on results.
- **Motion:** Slides in from left, 220ms ease-out, pushes command surface aside (not overlays). Slides out 180ms ease-in.

### 11.3 Screen 3 — Settings Pane ("What Synaptic Knows")

- **Access:** ⌘K → Settings, or single key from surface.
- **Layout:** Full-window pane. First row always: **"What I've done in the last 24 hours"** (the audit/replay — trust is built by visibility).
- **Sections in order:**
  1. Action replay (24h scrubbable timeline, expandable per action)
  2. Adaptive engine profile (editable, deletable, exportable)
  3. Permission grants (one-click revoke)
  4. Hotkey configuration
  5. Autonomy matrix (per-app + per-task-type dials)
  6. Backup controls
  7. Account, sync, integrations (lower priority, but visible)
- **Motion:** Surface expands to fill window from center, 260ms ease-out. Pulse moves to top-left corner. Settings content fades in after surface settles.

---

## 12. Component Primitive List (implementation order)

Primitives are built in dependency order. Each consumes only Layer 2 (semantic) tokens.

### 12.1 Tier 1 — Atomic (no dependencies)

1. `Hairline` — 1px line in `--border-subtle`, supports horizontal/vertical
2. `Pulse` — the vital sign. Props: state (idle/thinking/awaiting/error), size
3. `Dot` — status indicator. Props: variant (success/warning/error/info/neutral), size
4. `Stack` — vertical flex with token-driven gap
5. `Inline` — horizontal flex with token-driven gap
6. `Icon` — wraps SVG with locked stroke width, optical sizing
7. `Spacer` — fixed-size empty box using space tokens

### 12.2 Tier 2 — Inputs & controls

8. `Button` — primary/secondary/tertiary/destructive × idle/hover/active/disabled/loading
9. `Input` — text field, serif for command surface, sans for settings
10. `Textarea` — multiline
11. `Chip` — selectable suggestion chip, mono label
12. `Pill` — status pill (Done, Paused, Error), shape + icon + text
13. `Switch` — boolean toggle
14. `Slider` — value selector (sensitivity, strength)
15. `KeyCombo` — renders a key combo (`⌘⇧Space`) in mono
16. `HotkeyRecorder` — captures key combo, shows as `KeyCombo`

### 12.3 Tier 3 — Display

17. `Surface` — base container, supports token-driven bg/border/radius/padding
18. `Card` — Surface with optional title + actions
19. `Receipt` — one-line action result: timestamp mono + verb + target + check
20. `ProgressBar` — thin mono progress (thinking… 1.2s · claude-sonnet-4-6)
21. `EmptyState` — Equipment-at-rest composition (one muted line + optional affordances)
22. `LoadingState` — Per-state (cold/thinking/computer-use/long-running)
23. `Suggestion` — interpretation card (serif + sans steps preview, plum hairline when highlighted)
24. `ContextChip` — strip item showing detected screen element
25. `Avatar` — NOT a face. A pulse or initials, depending on context.

### 12.4 Tier 4 — Composite surfaces

26. `CommandSurface` — the layered omni-bar, all 4 states
27. `OnboardingWizard` — the 4 screens + First Breath
28. `ChatSurface` — editorial column
29. `ConversationDrawer` — history
30. `SettingsPane` — sectioned control surface
31. `AgentActionLog` — dense replay table
32. `Sidebar` — compact nav
33. `StatusBar` — ambient surface (menu bar status, kill switch status)
34. `ConsentModal` — native macOS dialog wrapper
35. `KillSwitchOverlay` — full-viewport black with red scan-line

---

## 13. Implementation Order (locked)

The build must proceed in this exact order. Skipping ahead risks building on undeclared foundations.

### Step 1: Tokens (foundational)
- Create `app/web/frontend/src/lib/tokens/` with: `primitives.css`, `semantic.css`, `motion.css`, `motion.ts`, `themes/{light,dark,hc,system}.css`, `themes.ts`, `index.ts`, `tokens.types.ts`
- Wire into `app/web/frontend/src/lib/style.css` as the first `@import`
- Verify: dark mode toggle via `<html data-mode="dark">` flips everything

### Step 2: Tier 1 primitives
- Build the 7 atomic primitives above
- Each gets a Storybook-style preview file in `app/web/frontend/src/lib/components/ui/__preview__/`
- Verify: each renders correctly in light, dark, hc modes; reduced-motion respected

### Step 3: Tier 2 controls
- Build the 9 input/control primitives
- Verify: keyboard navigation, focus rings, screen-reader announcements

### Step 4: Tier 3 display primitives
- Build the 9 display primitives
- Verify: each composes Tier 1 + Tier 2 correctly

### Step 5: CommandSurface (the heart)
- Build `CommandSurface.svelte` with all 4 states
- Wire to global hotkey
- Verify: <100ms hotkey-to-visible, glass backdrop, all 4 states, accessibility

### Step 6: OnboardingWizard
- Build the 4 screens + First Breath
- Wire to daemon `onboarding.*` RPCs
- Verify: each screen reachable, hotkey step is the only mandatory-no-skip

### Step 7: ChatSurface + ConversationDrawer
- Build the editorial column chat
- Build the drawer
- Wire streaming text motion (line-growing reveal)
- Verify: streaming looks alive, mono timestamps align, serif-vs-sans distinction clear

### Step 8: SettingsPane
- Build the 7-section pane
- Audit-first row (action replay) prominent
- Verify: each setting has a sentence explanation, all 7 sections reachable

### Step 9: Sidebar + StatusBar + App.svelte
- Wire the layout
- Compact sidebar, ambient status bar
- Verify: full app layout renders, navigation works, kill switch always reachable

### Step 10: Migrate existing routes
- Update existing routes (About, Audit, Channels, Delegation, Hub, Replay, Skills, Sync) to use new primitives
- Preserve functionality; only visual changes
- Verify: no regressions, all existing features still work

### Step 11: Polish & verification
- Run `npm run check` — 0 errors, 0 warnings
- Visual review at all 4 breakpoints
- Accessibility audit (axe-core or equivalent)
- Battery/perf check: motion respects energy mode
- Dark mode + HC mode visual review
- Reduced-motion check
- Update LOGBOOK.md with the redesign session

---

## 14. Accessibility (non-negotiable, locked)

The single rule from the UX agent that shapes everything:

> **If the user cannot see the screen, cannot use a mouse, and cannot speak — they must still be able to summon the agent, give it a task, and stop it. Everything else is a refinement.**

This means:
- All interactions have a keyboard-only path (Tab/Arrow/Enter/Esc).
- All status conveyed visually is also conveyed via text or ARIA live regions.
- All animations have a `prefers-reduced-motion` fallback that preserves the *intent*, not just removes the motion.
- All color contrast meets WCAG 2.2 AA (4.5:1 body, 3:1 large); HC mode meets AAA (7:1).
- All interactive elements have a visible focus ring (2px solid `--border-focus`, 3:1 contrast vs background).

**Voice as first-class input:** every interactive element in the command surface is reachable by voice. Voice transcript is editable before submission. STT confidence shown as subtle underline on uncertain words.

---

## 15. Anti-Patterns Guard (locked, do not violate)

The agent must NEVER:
- Default to dark mode on first launch (light is hero)
- Use a neon glow on dark glass
- Use Inter, SF Pro unmodified, Geist, or Manrope as primary font
- Use rainbow gradients or "AI = colorful blob" aesthetics
- Use a centered logo or wordmark in the title bar
- Use emoji in primary surfaces
- Use avatars with faces, robots, glowing orbs with eyes
- Use generic SaaS dashboard patterns (sidebar of icons + list of Workspaces/Projects/Members/Billing)
- Use spinners for "loading" — use the pause-level heartbeat instead
- Use color to convey status in chrome — use shape and motion
- Use shadows for normal hierarchy — use hairlines
- Use glass anywhere except the floating command overlay
- Use drop shadows with color
- Use any of the locked-out Inter/Geist/Manrope fonts

---

## 16. Verification Checklist (locked, before declaring done)

- [ ] `npm run check` — 0 errors, 0 warnings
- [ ] Light mode renders correctly across all routes
- [ ] Dark mode renders correctly across all routes (toggle `<html data-mode="dark">`)
- [ ] High-contrast mode renders correctly (toggle `<html data-mode="hc">`)
- [ ] `prefers-reduced-motion` falls back correctly (no motion, intent preserved)
- [ ] Keyboard-only navigation works on every screen
- [ ] Screen-reader announces correctly on every state change (NVDA + VoiceOver)
- [ ] Hotkey → CommandSurface appears in <100ms
- [ ] Kill switch always works (even when CommandSurface has crashed)
- [ ] First-run wizard completes and transitions to "First Breath"
- [ ] Streaming text reveals via line-growing, not character-fade
- [ ] Plum appears in ≤5% of any screen
- [ ] No shadows used for hierarchy (except kill switch scan-line)
- [ ] No glass anywhere except the CommandSurface
- [ ] All motion respects energy mode (motion ambient breath scales with battery)
- [ ] LOGBOOK.md updated with this redesign session

---

**END OF SPEC.** Implementation begins at Step 1 (Tokens) and proceeds in the order in §13. No new tokens, no new colors, no new fonts, no new durations without updating this document first.