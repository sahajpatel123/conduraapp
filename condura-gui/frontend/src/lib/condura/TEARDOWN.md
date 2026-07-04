# Onboarding & Settings Teardown — Mature 2026 Patterns

> **Goal:** steal the patterns that work, skip the ones that don't, and apply the
> synthesis to Condura's first-run. Source code is `Ritual.svelte` and
> `Settings.svelte` in this directory; the design system is `condura.css`.
> The user's brief: **OPTIONS** (not a forced wizard) and **smooth animations**.

---

## 1. Claude (claude.ai / claude.com)

### Onboarding observed
- **One-screen welcome.** New users land on the chat surface immediately. The
  input box is centered, large, auto-focused. Welcome copy is short: "Talk with
  Claude" or a single-line example prompt. No progress bar, no "step 1 of 4."
- **Sidebar appears on first message.** Threads are empty until you chat. The
  sidebar surfaces "Chats," "Projects," and "Recents" as separate collapsible
  groups, with monospaced timestamps in faint ink.
- **Settings is a modal sheet** that slides up from the right (on web) and a
  left-pushed page (on native). Theme picker is a dropdown with three words:
  *Light · Dark · System*. No icons, no preview swatches.
- **First-run animations are surgical.** A 320ms fade-up on the headline, a
  280ms slide-up on the composer on mount. No decorative shimmer, no looping
  pulse. The product is the calm — motion just reveals structure.
- **Onboarding voice:** quiet confidence. "Talk with Claude" + a one-line
  example ("Summarize this PDF") reads as "you're already using it."

### STEAL
- **One-screen welcome.** No wizard. The composer is the front door.
- **Sidebar appears empty** until the user has something to put in it. Don't
  show five empty states on first run — show the only thing that matters
  (the input).
- **Theme picker as three words in a row** (Light · Dark · System). No icons
  required to be understood. (Our current 3-button segmented text control
  already matches this — keep it.)
- **Surgical motion:** every animation explains a state change. Nothing loops
  on idle.

### SKIP
- The auto-focused input box isn't right for Condura — we have an EULA gate
  before any system access, which is correct for a system-level agent. We
  don't get to skip that.
- "Projects" / "Recents" sidebar groupings — overkill for a chat with the
  agent on a single machine. A simple thread list is enough.

### MAKE OUR OWN
- The EULA stamp as a **considered act**, not a checkbox. The wax-seal button
  with `sealBloom` keyframes is exactly the kind of detail mature products
  don't bother with — it earns the legal step's seriousness.

---

## 2. ChatGPT (chatgpt.com)

### Onboarding observed
- **One-screen welcome, prompt-centric.** Big input: "What can I help with?"
  Below it: suggested prompt chips (Create an image, Help me write, Make a
  plan, Code, etc.). The chips are tappable, immediately useful, and disappear
  once you've sent your first message.
- **Sidebar:** "New chat," "Library," "GPTs," "Projects" — flat vertical list,
  no pinned favorites, no status counts. Each row gets a one-line title that
  truncates.
- **Settings as a left-page sheet** (web) with a vertical sub-nav (General,
  Personalization, Data controls, Security, Account). Theme picker is in
  General → Appearance, dropdown with three words.
- **No first-run wizard. No forced account for browsing** (you can use it
  without an account, then get a soft "save your chats" prompt later).
- **Animations are minimal.** The composer has a 200ms focus-ring slide.
  Streaming tokens fade in character-by-character. No entry choreography.

### STEAL
- **Suggested-prompt chips** on the welcome screen. They are the
  lowest-friction way to teach a product: tap, see what happens, learn.
- **No forced account on first message.** Account prompt comes later as a
  gentle "save this thread?" — the user has felt value first.

### SKIP
- The 7+ vertical settings sub-nav. Condura's "flowing document" Settings
  (one long column, no tabs) is more deliberate; we don't need to copy the
  ChatGPT settings tree.
- Project/Recents grouping. We're a single-user, single-machine product.

### MAKE OUR OWN
- **Suggested prompt chips for the "First Breath" welcome.** Three chips, in
  monospace, faintly dimmed — *"Open Safari and find the docs I had open
  yesterday"*, *"Summarize the unread emails"*, *"Build me a quick script"*.
  They demonstrate breadth in two seconds without forcing a tutorial.

---

## 3. Manus (manus.im)

### Onboarding observed
- **Tagline-first hero.** "Less structure, more intelligence." A single text
  input centered on the page with five chips below: Create slides, Build
  website, Develop desktop apps, Design, More. The chips are the entire
  suggested-prompt system — no welcome video, no wizard.
- **No account wall before value.** You can submit a prompt immediately; the
  auth step appears as a side-effect of wanting to *save* the output.
- **Voice:** restrained, technical, agent-first. The product assumes you know
  what an AI agent does. No "Welcome to Manus!" energy.
- **No first-run theme picker.** Inherits the system, surfaces the toggle in
  Settings.

### STEAL
- **Tagline as the front door.** One sentence that orients the product.
- **Restrained copy tone.** "It perceives only what it must, and acts only
  after it shows you what it's about to do" — this is exactly the Manus
  register. Keep it.
- **Five chips, not fifteen.** Five is the threshold where the user reads
  every chip. Fifteen is the threshold where they read none.

### SKIP
- The flat "Build website / Develop desktop apps" framing. Condura is more
  nuanced — it observes, decides, and asks permission. Manus's chip language
  is too transactional.

### MAKE OUR OWN
- **Probe-first onboarding.** Manus doesn't probe for tools installed on your
  machine. We do (Ollama reachable, screen recording granted, mic
  available). Our constellation already reflects the probe state — Manus's
  single input box doesn't show that depth. The constellation is *the*
  differentiating pattern; lean into it.

---

## 4. Linear (linear.app)

### Onboarding observed
- **No wizard at all.** You sign up and you're in the workspace. The empty
  states are *opinionated populated states*: "Backlog 8 · Todo 71 · In
  Progress 3 · Done 53" — already populated with demo data so the structure
  is legible before you've created anything.
- **Cmd+K (command palette) is the meta UI.** Every action, every nav, every
  settings page is reachable through the palette. Theme is `Cmd+K → "Theme"`
  → select. Never a forced modal.
- **Sidebar:** vertical, collapsible groups (Inbox, My Issues, Views,
  Workspace, Initiatives). Pinned favorites at the bottom. Issue counts in
  monospaced `02/145` style.
- **Visual polish:** Inter for UI text, monospace for metadata (timestamps,
  issue IDs, status). 12px monospace for `ENG-2703` and `4 min ago` style
  timestamps. Code blocks use a real code mono font with proper gutters.
- **Animations:** opacity-only. No bouncy springs, no scale-pop. The product
  moves by changing what's there, not by animating how it gets there.

### STEAL
- **Opinionated populated empty states.** Even on first-run, show the user
  what their tree *will* look like after they fill it in. The constellation
  already does this — nodes are visually present before they're wired.
- **Cmd+K as the escape valve.** The user should be able to escape the
  first-run ritual via `Cmd+K` and reach any setting. (Our existing
  `CommandPalette` already does this once the ritual completes — extend it
  to work *during* the ritual too.)
- **Monospaced metadata:** issue counts, status, timestamps. We already use
  `--font-mono` for these in the constellation ("CONDURA" label) — extend to
  node counts: "1/6 wired."

### SKIP
- Linear's no-onboarding is **the wrong model for us.** They can get away
  with it because Linear is a self-serve SaaS where the empty workspace IS
  the value. Condura is a system-level agent that needs permissions and an
  EULA gate before it can touch the OS. We can't skip the legal step.
- "No animations" is also too far for us. Our design system is paper-
  textured, breathing-pulse, life-on-screen. We're not Linear. **We need
  motion that says "I am alive, I am listening"** — Linear's product
  doesn't need that voice.

### MAKE OUR OWN
- **Live status on each constellation node.** Linear's sidebar shows
  `02/145` issue counts in monospace. Our constellation nodes should show
  live status: "Access · granted", "Hotkey · ⌘⇧Space", "Voice · ready",
  "Threads · 0/5 connected", "Account · signed in". The constellation
  stops being a decorative SVG and becomes a status sidebar.

---

## 5. Arc (arc.net)

### Onboarding observed
- **Cinematic first-run.** Arc was the product that brought the
  choreographed-onboarding moment into mainstream — when you first launch,
  the sidebar animates in from the left edge, the day's wallpaper blooms,
  and your first pinned tab gets a tooltip "this is where things go." The
  whole thing takes 8 seconds and is paced to feel cinematic without
  feeling slow.
- **Sidebar with Spaces + Pinned Tabs + Profiles.** Three orthogonal axes
  for organizing browser state. Spaces are top-level tabs in the sidebar;
  Pinned tabs are the second-tier icons that don't have a tab; Profiles
  swap entire sets of Spaces.
- **Themes are per-space.** Arc ships 7+ color presets (Day, Night,
  Midnight, Sunrise, etc.) plus a custom-CSS escape hatch. Theme is a
  space-level setting, not a global one.
- **The "Favorites" preview strip** — when you hover a pinned tab, Arc
  shows a 240px-tall preview pane of that tab's content. This is the
  **single highest-leverage UI pattern Arc invented.** It tells you what
  each pinned thing *is* without forcing a click.
- **Animations:** 320-560ms ease-out, scale + opacity. Tasteful. Used to
  indicate hierarchy (a pinned tab selected grows from 64px → 80px).

### STEAL
- **Live preview strip.** When the user hovers a constellation node
  ("Summon"), a 200px-tall preview strip appears to the right showing
  *what that node currently is* — e.g., hover "Hotkey" → preview shows
  the actual keycap row "⌘ ⇧ Space". Hover "Channels" → preview shows
  "Telegram · ready". This is the single highest-leverage addition to
  the existing constellation.
- **Cinematic first beat.** Arc's 8-second arrival sequence is exactly
  the right amount of theater for a first launch. The current
  Ritual.svelte already does this (3.4s `moteDrift` + 2.6s `firstBeat`).
  Keep it, but make it shorter (1.8s) so the user is engaged, not
  waiting.
- **Per-context customization.** Arc's per-space themes are over-engineered
  for us, but the *idea* of context-aware settings (e.g., "voice is
  enabled for this device profile") is a useful future direction.

### SKIP
- Three orthogonal axes (Spaces/Pinned/Profiles). We don't need that
  organizational depth. The constellation IS our sidebar.
- Custom CSS escape hatch. Power-user feature; not for v1.
- The 8-second cinematic sequence. Arc got away with it because the
  browser itself has a high "time to value" (you have to load a page).
  Condura's first-run is faster — the user is on a small overlay,
  not a full window. **1.8 seconds of arrival is enough.**

### MAKE OUR OWN
- **The constellation IS the spaces/Pinned/Profiles collapsed into one
  metaphor.** Six nodes, each is a "space" you can wire up or skip. The
  hover-preview gives the same at-a-glance scanability Arc invented,
  but condensed into a single radial layout. This is a uniquely Condura
  pattern that no reference product does.

---

## 6. Notion (notion.so / notion.com)

### Onboarding observed
- **Marketing-site-as-onboarding.** Notion's onboarding is **the marketing
  site itself.** They give you templates, example workspaces, a "What is
  Notion?" video right on the homepage. When you sign up, the empty
  workspace shows you the templates you just saw on the site — already
  familiar.
- **Block editor as onboarding.** The single deepest innovation: the
  *type* is the verb. You type `/` and a menu appears with blocks (image,
  table, code, etc.). You don't navigate a toolbar — the cursor IS the
  navigation.
- **Sidebar is collapsible + nested.** Workspace → Pages → Sub-pages.
  Favorites pinned at top. Templates in a "Templates" section.
- **Theme picker is in Settings → Appearance, dropdown with three words**
  (Light · Dark · System). Same pattern as everyone else. No icons.
- **Animations:** opacity-only transitions on page navigation. No
  decorative motion. The product feels like a calm writing app.

### STEAL
- **The cursor-as-verb.** Condura's Quick Prompt overlay already does
  this — type to talk, `Cmd+K` to command. The whole product should feel
  like "you type, things happen" without navigating. This is the
  *fundamental mature pattern* of 2026 AI products: the input is the
  UI.
- **Templates as marketing-as-onboarding.** The suggested-prompt chips
  on the First Breath welcome are our templates. Three to five. Already
  proposed in §2.

### SKIP
- Workspace → Pages → Sub-pages nesting. We're not a doc product.
- The `?` help-menu tour overlay. Condura's `Cmd+K` palette is a more
  advanced affordance.

### MAKE OUR OWN
- **The "press your hotkey" final beat.** Notion has no equivalent of
  "the product only exists when you press a key." Our First Breath →
  Enter Condura → user presses their hotkey → overlay appears → product
  is alive is a uniquely Condura moment. Make it last 600ms longer
  than feels necessary. The payoff is the user *invoking* the agent.

---

## 7. Synthesis

### What's the 2026 mature onboarding pattern?

Common threads across the six teardowns:

1. **One screen, not a wizard.** Five of six teardowns have *no wizard at
   all.* The sixth (Claude's gated thread creation) has a single optional
   question. **The forced 5+ step wizard is dead.** It survived from
   2014-era SaaS (MailChimp, Intercom v1, early Stripe) and was killed by
   Linear (2019) and Arc (2021) showing the world you can just *be in
   the product.*
2. **The input box is the front door.** ChatGPT, Claude, Manus, Notion —
   the user types immediately. The product is the input, not the
   prelude to the input.
3. **Suggested-prompt chips for guidance.** Five chips is the threshold
   for "I read every chip." Three is the minimum. Fifteen is too many.
4. **Probe-and-default.** Mature products detect what's available
   (camera, mic, screen recording, system theme, locale) and only ask
   for what's missing.
5. **Theme picker as 3-way segmented control or dropdown with three
   words.** Sun/auto/moon *icons* (Slack) or just text. **No forced
   modal on first run.** Inherit system, expose toggle in Settings.
6. **Skip is a first-class affordance, not a tiny link.** Every optional
   step has a real "skip, do this in Settings later" entry point. The
   user is never trapped.
7. **Motion explains state.** Loading→loaded fades. Empty→filled
   springs. Decorative loops are dead.
8. **Opinionated populated empty states.** Show the user what their
   tree looks like *before* they've filled it.

### The anti-pattern we're avoiding

The **vibe-coded onboarding**:

- Forced 5+ step wizard before any value.
- "Step 3 of 7" progress that signals "you're in a form, not a product."
- "Let's set up your account!" energy. Emojis. Exclamation points.
  "🎉 Welcome aboard!" is a smell.
- **TWO buttons per step** (Continue + Skip) → user picks one → next →
  repeat → fatigue → dropoff. **NEVER both as buttons.**
- Mandatory account creation before any exploration.
- "Choose your preferences!" form with 12 checkboxes.
- Forced theme picker modal on first run.
- "Loading the experience…" spinners over a blank screen.

**The current Condura Ritual.svelte has several of these smells:**
- 9-step forced wizard (`arrival → eula → permissions → power → hotkey →
  voice → channels → account → breath`).
- 6 of the 9 steps are optional (voice, channels, account, hotkey is
  skippable, power can be deferred, permissions can be skipped) but the
  user must click through each one. That's not "options" — that's
  **forced choice architecture.**
- "Leave this for later" appears as a faint italic text-link in the
  bottom-left margin. That's a skip-link as afterthought, not as
  primary affordance.

### Apply to Condura: the new first-run

**Make a call: ONE screen with constellation cards + EULA gate.** Not
one screen with a wizard (which is what we have). One screen with
**six self-contained cards arranged as a radial constellation** (the
SVG we already have, but the circles become clickable cards).

#### The flow

**Screen 1 — EULA gate** (kept as-is, slightly tightened)
- Single screen, no wizard context.
- "First, the terms." + scrollable EULA + checkbox + wax seal button.
- The seal `C` button (already in `Ritual.svelte`) is the only
  highlight. No "Continue" button — just the stamp.
- On stamp: 600ms `sealBloom` → dissolve → Screen 2.

**Screen 2 — The Constellation** (the new room)
- The window's full bleed is one screen.
- The constellation SVG is centered, **but the six nodes are now live
  clickable surfaces**, not passive decoration.
- Each node shows its **live status** in monospaced 10px label below:
  - **Perceive** · "Accessibility: granted · Screen Recording: pending"
  - **Power** · "Ollama: 3 models detected"
  - **Summon** · "Hotkey: not set" or "Hotkey: ⌘⇧Space"
  - **Voice** · "Mic: available · wake: off"
  - **Threads** · "0 of 5 ready"
  - **Account** · "Not signed in" or "Signed in: email@example.com"
- **Click a node → right-side panel slides in (380px wide, slide 24px
  from right, 320ms ease-out).** The panel contains that step's full
  options — the same UI elements we have today (`completePower`,
  `saveHotkey`, etc.), just relocated into a side panel rather than
  stepping through them as wizards.
- **Hover a node → preview strip** (Arc's pattern). A 60px-tall strip
  appears below the constellation showing a one-line preview: hover
  "Summon" → "Press your combo to call Condura" + the actual keycap row
  if set. Hover "Power" → "Your model. Your key. Local or remote."
- **Bottom-center pill: "Enter Condura →".** Always enabled, always
  visible. The user can wire 0 nodes or all 6 — the door is always
  open.
- **The constellation is the legend.** Wired nodes are solid
  synapse-green. Skipped nodes are dotted faint. Unvisited nodes are
  empty circles with a faint hairline. **Each node tells you its
  state at a glance.** This is what Linear does with `02/145` issue
  counts in the sidebar, but for your personal setup.

#### The animation grammar

Consistent across all node interactions:

- **Node entry (mount):** `scale 0.92 → 1`, `opacity 0 → 1`, 280ms
  ease-out. Stagger 60ms across nodes for a constellation fade-in.
- **Node hover:** `scale 1 → 1.05`, 180ms ease-out. Border shifts to
  `--pollen`. Preview strip fades in below, 220ms ease-out, 80ms
  delay.
- **Node click:** `scale 1 → 0.97 → 1` (the same `dot-pop` keyframe
  used by `Settings.svelte` autonomy dots, 180ms). Side panel slides
  in simultaneously.
- **Side panel entry:** `translateX(24px) → 0`, `opacity 0 → 1`,
  320ms ease-out. Background masked by the panel, constellation
  stays visible.
- **Wired transition:** node fill animates from `var(--paper)` →
  `var(--synapse-light)` over 600ms. SVG path from
  `dashedOffset 1 → 0` (existing `drawthread` keyframe) over 700ms.
- **Skip transition:** reverse — solid node fades to dotted outline
  over 400ms.
- **Constellation idle (only when something is unwired):** one slow
  pollen mote drifts from the center to one unwired node every 14s,
  1.4s duration, then disappears. Suggests "tap me" without nagging.
  Honors `prefers-reduced-motion`.
- **"Enter Condura →" pill:** sits on a faint synapse halo that
  pulses 3s loop when no nodes are wired (inviting), and stops
  pulsing once at least one node is wired (the user is engaged, the
  mote is enough).

#### The "skippable" rule

Everything on Screen 2 is skippable from a single click — the user
just doesn't tap the node. They press "Enter Condura →" and the
constellation moves to its unwired state (dotted lines, empty
circles). **No "skip" link needed.** The architecture *is* the skip.
The product works fully without any of the six nodes wired; the
EULA gate is the only non-skippable step.

#### One exception: the hotkey

If the user reaches "Enter Condura →" with no hotkey set, the pill
hovers to "Set a hotkey to enter →" and pulses. This is the only
moment of forced choice — and it's the **only** mandatory setup item
beyond the EULA, because (per `CLAUDE.md` locked decision #8) the
user must pick a hotkey; there is no default. One click on "Summon"
node, one keypress, done.

#### Why this is right for Condura specifically

- **It uses the metaphor we already have.** The constellation SVG is
  already in `Ritual.svelte:414-442`. We're not inventing a new
  pattern; we're activating the one that exists but is decorative.
- **It respects the EULA gate.** Screen 1 stays. Legal happens
  before any system access.
- **It gives OPTIONS.** Six independent choices, no forced order.
  Linear-style.
- **It has smooth animations.** Each interaction has a defined
  motion. None are decorative. All explain state. The user *sees*
  what they just wired light up.
- **It's symmetric with `Settings.svelte`.** Both screens are
  documents the user can enter and leave. The constellation appears
  in both places (collapsed on Settings, expanded in the ritual).
- **It's not a wizard.** One screen with side panels is the 2026
  mature pattern (System Settings on macOS, Slack admin, Linear
  workspace settings). The wizard pattern would be 2014.

### The settings theme picker (teardown)

| Product | Control | Labels | Icons |
|---|---|---|---|
| Linear | Dropdown | "Light / Dark / System" | None |
| Notion | Dropdown | "Light / Dark / System" | None |
| ChatGPT | Dropdown | "Light / Dark / System" | None |
| Claude | Dropdown | "Light / Dark / System" | None |
| Arc | Per-space picker | 7+ named presets + custom CSS | Theme swatches |
| Slack | 3-way segmented | "Light / Auto / Dark" | Sun, half-circle, moon |
| macOS System Settings | 3-way segmented | "Light / Dark / Auto" | Sun, moon, half-circle |
| iOS | 3-way segmented | "Light / Dark / Automatic" | Sun, moon, half-circle |
| **Current Condura** | 3-way segmented | "auto / light / dark" | None |

**The mature pattern is the 3-way segmented control with sun/half-moon/moon icons** — macOS, iOS, Slack. The dropdown-with-three-words pattern is the *web* pattern (Linear, Notion, ChatGPT, Claude). Both work.

**STEAL:** Replace our text-only buttons in `Settings.svelte:508-519` with the icon-prefixed 3-way segmented control (sun / half-circle / moon). 24×24 icons inline, labels to the right.

**SKIP:** Making the theme picker the centerpiece of first-run. Don't ask. Inherit the system. The user can change it in Settings once they're using the product.

**MAKE OUR OWN:** The segmented control's active button gets a `--synapse` fill with `--paper` icon — same treatment as today. But the three icons themselves should be drawn in **Condura's line vocabulary** — 1.25px thread-weight, ink-faint when inactive, paper-on-synapse when active. Three custom inline SVGs, not lucide.

### Concrete next steps (for the design pass)

1. **Replace the 9-step wizard state machine in `Ritual.svelte` with a 2-state machine** (`'gate' | 'constellation'`). The `gate` state is the EULA screen. The `constellation` state is the new room.
2. **Promote the existing constellation SVG from passive decoration to active UI.** Nodes become buttons. Add live status labels under each node (mono, 10px).
3. **Add a slide-in side panel component** (`NodePanel.svelte`) — 380px wide, right-aligned, 320ms slide-in. Renders the step-specific content from the existing handlers (`completePower`, `saveHotkey`, etc.).
4. **Add the hover preview strip** — 60px tall, appears below the constellation on node hover. Single line of copy + a small visual (keycap row, status dot, etc.).
5. **Replace "Continue →" per step with one "Enter Condura →" pill** at the bottom-center. Always enabled. Only soft-locked when hotkey is unset.
6. **Strip the bottom-margin skip-note affordance.** Skipping is now the *default* — the user simply doesn't tap a node.
7. **Add three sun/half-moon/moon inline SVGs** to `Glyph.svelte` for the theme picker. Replace text-only buttons in `Settings.svelte:508-519` with icon + label.
8. **Keep the arrival sequence, tighten it.** Arc got away with 8s because it's a browser with high time-to-value. Condura's first-run overlay should bloom in 1.8s, not 3.4s. Cut the `moteDrift` duration in half.

---

## Closing note

The constellation metaphor is already the right answer. We're not
redesigning from scratch — we're activating an existing artifact. The
six nodes were always meant to be live. The wizard shell around them
was the temporary scaffolding we built while the daemon wasn't ready
to defer steps; now that it is, the wizard can dissolve and the
constellation can become the room.

The single biggest UX win: **replace the 9-step forced wizard with a
1-screen constellation + 1-screen EULA gate.** That alone collapses
the perceived time-to-value from "9 clicks + 9 animations" to "1
considered act + 0-6 voluntary actions."

The second biggest: **add the Arc-style hover preview strip on each
constellation node.** This is the highest-leverage micro-pattern in
modern UI, and it costs ~30 lines of Svelte.

The third: **inherit system theme silently; expose sun/half-moon/moon
segmented control in Settings.** Never force theme on first run.