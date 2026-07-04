# SCREEN_SETTINGS — Condura

> The heart of configuration. One long column of italic Instrument Serif section
> titles separated by hairlines. Not a modal. Not a tabbed panel. **A flowing
> document.** The autonomy matrix is the hero; the sticky save bar is its
> callback to the titlebar Thread.
>
> Source: `condura/Settings.svelte` (Phase 15). Component contract:
> `app/web/frontend/src/lib/condura/specs/` family — sibling docs cover every
> other surface. This document is the spec the implementation is held to.

---

## 1. LAYOUT & CONTENT

### 1.1 Document shape

| Property | Value | Why |
|---|---|---|
| Type | Flowing document, NOT a dashboard / NOT a modal | MOAT §1 — premium products earn their flourishes, dashboards perform them |
| Max-width | `880px` (`65ch`), centered | Paper column. Readable on a 13" laptop without head-turn. |
| Top | `doc-head` — eyebrow "— Configuration", h1 "Settings", Thread, lead | Echoes every other route hero at 1/3 the weight |
| Body | 8 sections, separated by `.hair` (linear-gradient hairlines with 8% transparent tails) | No tabs. The hair is the divider. |
| Bottom | Sticky pollen save bar (when dirty or `savedFlash`) | The bar IS the titlebar Thread, called back into the document |
| Sticky nav | **None.** No left rail. No top tabs. Anchor links via `#/settings/<section>` hash | Section nav violates "flowing document" — discoverability lives in the CommandPalette (`⌘F → sections`) |

### 1.2 The 8 sections (in render order)

| # | Section | Sub-controls | Source line | Notes |
|---|---|---|---|---|
| 1 | **Appearance** | Theme (sun/auto/moon → delegates to `ThemePicker`), Motion strength (slider), Grain intensity (slider) | `Settings.svelte:498-560` | All local UI prefs — applied at once, persisted to localStorage, **never dirty** |
| 2 | **Power** | Energy budget (low/balanced/high/auto seg), Default model per provider (text fields, dirty) | `Settings.svelte:564-610` | Energy=low implies motion=0 (the two fight, kept consistent) |
| 3 | **Autonomy matrix** ★ HERO | 11 task rows × 3 state dots (block/warn/autonomous), live preview line | `Settings.svelte:614-672` | The single most important config in the app. CLAUDE.md §27. |
| 4 | **Adaptive engine** | Learning strength — 4 dots (off/cautious/balanced/aggressive) | `Settings.svelte:676-702` | Best-effort; if adaptive engine offline, dots render but don't persist |
| 5 | **Voice** | Wake word toggle, sensitivity slider, status line (`onboardingProbeVoice`) | `Settings.svelte:706-752` | Local prefs; sensitivity slider disables when wake off |
| 6 | **Account** | Signed-in chip (avatar + name + email + provider) OR signed-out inline link | `Settings.svelte:756-780` | `account.isSignedIn` is the gate; sign-in is a thread-link span |
| 7 | **Permissions** | Read-only list with `granted` / `denied` / `unknown` badges | `Settings.svelte:784-800` | Polled 2s via `ipc.permissionsStatus()`; deep-link row per permission |
| 8 | **Legal** | EULA inline expander — version chip + expanded `<pre>` | `Settings.svelte:804-823` | Loads on first open via `ipc.onboardingEula()` |

> The roadmap surface (Hotkey re-record, Kill Switch status row, Channels list,
> Sync pairing status, Advanced — data dir / reset / export) is **deferred to
> a future Phase 15B pass** and not in the shipped spec. When it lands, it
> arrives below Legal as section 9.

### 1.3 The Autonomy Matrix — the hero

| Property | Value |
|---|---|
| Source of truth | `CLAUDE.md §27` (Autonomy Matrix, locked) |
| Grid | `display: grid; grid-template-columns: minmax(120px,1fr) 28px 28px 28px minmax(110px,1.2fr)` |
| Row count | **11 canonical task types** in stable order: `coding`, `file_operations`, `web_browsing`, `email`, `calendar`, `messaging`, `shell_commands`, `computer_use`, `research`, `image_generation`, `code_review` |
| Columns | Task name · Block · Warn · Auto · "Now" (live verb phrase) |
| Cell | `<button class="auto-dot">` — 18×18 px hollow ring with hairline in the **state color** (inactive) or filled dot inside a pollen halo (active) |
| Click inactive dot | Set that state |
| Click active dot | Cycle forward (block → warn → auto → block) |
| State colors | block=`--synapse`, warn=`--warn`, autonomous=`--ok` (status palette, not brand — see DIRECTION §4) |
| Pop animation | `dot-pop` keyframe — scale `1 → 1.18 → 1` over **180ms** `--ease`. Cleared on 200ms timeout so repeated clicks re-trigger. |
| Loading | `matrix-loading` — `—` in `--content-faint` + 360px pollen `<Thread>` drawing in (`configEmpty === !settings.config && !settings.loaded`) |
| Live preview line | Right below the matrix, in a `.preview` paper-card well. Renders *"Right now, for **coding**, Condura will **block** before acting."* — the verb takes the state color, the rest stays in `--content-soft`. Cross-fades 200ms when the dot for `coding` is clicked. |
| Hairlines between rows | `:before` on each `.matrix-task-col` (except first) — draws a 1px `--hair` line across the full grid row |

### 1.4 Right-rail removed

The shipped Settings has **no left nav and no right rail.** Section navigation
is by scroll, by `⌘F` palette search (filters to sections), and by the
`#/settings/<anchor>` hash. This is a deliberate rejection of the tabbed-panel
settings pattern — see MOAT §1 ("competent but generic is how solo founders
lose a year"). The page is read like a document.

### 1.5 The Save Bar (the callback to the titlebar Thread)

| Property | Value |
|---|---|
| Mount | `position: sticky; bottom: 0; z-index: var(--z-sticky)` |
| Mount trigger | `dirty === true \|\| savedFlash === true` (state flag, not route enter) |
| Enter | `transition:fly={y: 64, duration: 380, easing: backOut}` — springs up from below |
| Paper | `--pollen` background, `--paper` text, `--r-pill`, `--pollen-halo` + `--shadow-float` |
| Hairline below | `.save-bar-thread` — a `<Thread orientation="h">` at `opacity: 0.55`, in `--pollen` |
| Status | mono-uppercase `SAVING…` (acting) / `Unsaved changes` (awaiting) / `Saved` (ok) / `Save failed` (error) + `Pulse` per state |
| Actions | `[Revert]` (ghost) + `[Save]` (primary, paper-on-pollen) |
| Exit on saved | After 400ms the `.save-bar--saved` class fades opacity 0 + translateY(8px), then unmounts via `transition:fly` reverse |
| Exit on revert | `dirty = false; syncFromConfig()` — bar unmounts |
| Failed exit | Bar **stays** (still dirty) — error message inline |
| Backdrop hover | `box-shadow` widens the pollen halo 8px (`pollen × 14% transparent`) |

---

## 2. STATE MATRIX

| State | Visual signature | Trigger |
|---|---|---|
| **default** | Document renders; autonomy matrix shows current `settings.config.autonomy`; local prefs read from localStorage on mount; permissions/voice probed best-effort | `onMount` |
| **loading** (config) | Autonomy matrix shows `—` in `--content-faint` with a 360px pollen Thread drawing in. Power section shows "—" in provider rows. EULA shows nothing until opened. | `!settings.config && !settings.loaded` |
| **loading** (EULA) | `.eula-loading` block — `Pulse phase="thinking" size=8` + `READING THE LICENSE…` mono-uppercase label (not a spinner — MOAT §2.5) | `openEula()` invoked, awaiting `ipc.onboardingEula()` |
| **dirty** | Save bar springs in from bottom (backOut 380ms). Revert + Save buttons enabled. | User clicks a matrix dot, edits a provider model, changes adaptive strength |
| **saving** | Save bar status → `SAVING…` + `Pulse acting`. Save button disabled. Revert disabled. | `settings.saving === true` |
| **saved** | Save bar status → `Saved` + `Pulse ok`. After 400ms the bar fades out (`opacity 0`, `translateY(8px)`) and unmounts. | Daemon acks `config.update` |
| **error (save)** | Save bar status → `Save failed` + `Pulse error` + raw error message inline. Bar stays (still dirty). | `settings.lastSaveError` set, daemon rejected the patch |
| **permission-denied** | Permissions list: each `.perm-badge[data-status='denied']` in `--danger`. Per-row deep-link "Open System Settings" button. | `ipc.permissionsStatus()` returns `denied` |
| **theme-switching** | **Delegated to ThemePicker.** Settings calls `applyTheme(t)` which sets `data-mode` and localStorage. The picker mounts separately and owns the palette-switch choreography (clip-path circle from `--ox/--oy`). | User clicks a theme seg |
| **adaptive-offline** | Adaptive dots render, click writes local-only state, `ipc.adaptiveStrengthSet` rejects silently. Best-effort. | `ipc.adaptiveStrengthGet` rejects |
| **voice-probe-failed** | Voice section renders without a status line. Toggle + slider still work (localStorage). | `ipc.onboardingProbeVoice` rejects |
| **daemon-offline (general)** | Local prefs continue to work. Daemon-backed changes (matrix, models) don't persist; save bar shows "Save failed" indefinitely. | All daemon calls reject |

---

## 3. MOTION CHOREOGRAPHY

### 3.1 Enter

| Surface | Motion | Duration | Easing |
|---|---|---|---|
| Route enter (whole document) | `opacity 0→1` + `blur(8px)→blur(0)` + `translateY(12px)→0` | `--dur-slow` (520ms) | `--ease` |
| Section headers | `sect-title` italic Instrument Serif renders statically (no per-section enter) — the `.hair` rule draws in below it | 280ms (thread draw) | `--ease` |
| Autonomy matrix rows | **None on mount** — the dots are present immediately. The `popping` animation only fires on click. | — | — |

### 3.2 Interactions

| Gesture | Motion | Duration | Easing |
|---|---|---|---|
| Dot click (autonomy matrix) | `dot-pop` — `scale 1 → 1.18 → 1`. Cleared on 200ms timeout so repeated clicks re-trigger. | 180ms | `--ease` |
| Dot hover | `scale(1.08)` + halo `pollen × 18%` | `--dur` (280ms) | `--ease` |
| Live preview line update | Verb color cross-fades (200ms transition on `color`). Sentence text fades 200ms in/out. | 200ms | `--ease` |
| Save bar enter | `transition:fly` — `y: 64 → 0` + `opacity 0 → 1` | 380ms | `backOut` (settling overshoot) |
| Save bar saved exit | `.save-bar--saved` class — `opacity 1 → 0` + `translateY(0) → translateY(8px)`, then unmount | 400ms | `--ease` |
| Save bar hover | Box-shadow halo widens from 4px pollen to 8px pollen × 14% | `--dur` | `--ease` |
| EULA inline expand | `<pre>` fades in below the row, max-height 320px overflow auto | `--dur-slow` | `--ease` |
| Theme switch | **DELEGATED** — see `ThemePicker.svelte:151-225`. Clip-path circle from click origin, OR View Transitions API custom-morphed. Reduced-motion → instant switch. | 560ms (or instant) | `--ease` |

### 3.3 Reduced-motion contract

Owned by `condura.css` (one block, MOAT §2.3). Settings reads it:

- `dot-pop` animation → `none`
- Save bar `fly` enter → duration 0
- Section thread draw → duration 0
- `prefers-reduced-motion: reduce` + `data-energy='low'` (set by `applyEnergy('low')`) → both snap durations

---

## 4. KEYBOARD

| Combo | Action | Notes |
|---|---|---|
| `Tab` / `Shift+Tab` | Cycle through sections in DOM order | All interactive elements have a focus-visible halo (pollen halo, MOAT §2.1) |
| `⌘S` | Save (calls `save()`) | Bound at the document level when route is `#/settings` |
| `⌘Z` | Undo last in-flight edit | Reverts to the last saved config snapshot |
| `Esc` | Discard pending changes | Reverts all working-copy state, hides save bar |
| `⌘F` | Focus the section filter input (via CommandPalette) | Filters the section list by query |
| Autonomy matrix: `Tab` into row, then `←` / `→` | Move between Block / Warn / Auto dots | Roving tabindex |
| Autonomy matrix: `Enter` / `Space` on a focused dot | Cycle forward (block → warn → auto) | Same as clicking the lit dot |
| ThemePicker | `←` / `→` move focus between sun/auto/moon; `Enter` / `Space` commit | Per `ThemePicker.svelte:244-279` |
| EULA row | `Enter` / `Space` toggles inline expander | `aria-expanded` reflects state |

---

## 5. COMPONENTS USED

| Component | Role | Source |
|---|---|---|
| `ThemePicker` | Phase 2 shipped. Segmented control + palette-switch choreography. **Replaces** the inline `seg-btn` theme row in the old Settings. | `condura/ThemePicker.svelte` |
| `AutonomyMatrix` | The hero. 11×3 dot grid + preview line. (Currently inlined in `Settings.svelte:614-672` — refactor to dedicated component if Phase 16 splits.) | `condura/Settings.svelte` (inlined) |
| `ApiKeyInput` | Text field for provider default model (`<input class="field">` underline-only). Mono font, hairline border-bottom. | `condura/Settings.svelte:597-606` |
| `PermissionBadge` | `.perm-badge[data-status=...]` — 10px mono-uppercase, color follows `--ok` / `--danger` / `--content-faint`. | `condura/Settings.svelte:1425-1439` |
| `StickySaveBar` | The pollen save bar (`.save-bar` + `.save-bar-inner` + `.save-bar-thread`). | `condura/Settings.svelte:1512-1608` |
| `Thread` | Hairline draw used in 4 places: doc-head rule, section `.hair` dividers (via gradient), EULA loading state, save-bar-thread. | `condura/Thread.svelte` |
| `Pulse` | Save-bar status indicator (acting/ok/error/awaiting phases), EULA loading indicator. 8px. | `condura/Pulse.svelte` |
| `Glyph` | Icons for the EULA expand chevron (when extracted), optional inline icons in permission rows. | `condura/Glyph.svelte` |
| `Button` | Save bar actions (Revert ghost, Save primary). Uses paper-on-pollen contrast overrides. | `condura/Button.svelte` |
| `Tooltip` | Hover hints on save-bar error message (truncated to 30ch); not yet active. | not yet built (MOAT §2.9) |
| `Switch` | Wake word toggle (`.toggle` + `.toggle-knob` — 44×24px pill). | `condura/Settings.svelte:1286-1337` |
| `Slider` | Motion strength, grain intensity, wake sensitivity — pollen thumb, hairline track with `--slider-fill` CSS var. | `condura/Settings.svelte:1006-1075` |

---

## 6. DATA FETCHED

### 6.1 Daemon RPCs (JSON-RPC 2.0 over Unix socket)

| RPC | Direction | Trigger | Used by |
|---|---|---|---|
| `config.get` | request | mount (via `settings.refresh()`) | populates `settings.config`, drives the working-copy `autonomyPerTask` + `providerModels` |
| `config.update` | request | save bar Save | writes the autonomy patch + provider model patch back to daemon |
| `permissions.status` | request | mount + polled every 2s | populates `.perm-badge` rows |
| `onboarding.eula` | request | first time EULA inline expander opens | populates `eulaText`, `eulaVersion`, `eulaUpdated` |
| `onboarding.probe_voice` | request | mount | populates `voiceProbe` (mic_available, wake_word_enabled) |
| `adaptive.strength.get` | request | mount | populates `adaptiveStrength` |
| `adaptive.strength.set` | request | user clicks a strength dot | persists the chosen strength |

### 6.2 Local stores

| Store | Direction | Notes |
|---|---|---|
| `settings.store` | read | `settings.config`, `settings.loaded`, `settings.saving`, `settings.lastSaveError` |
| `account.store` | read | `account.isSignedIn`, `account.avatarURL`, `account.displayName`, `account.email`, `account.provider`; `account.checkStatus()` on mount |

### 6.3 LocalStorage (local UI prefs only — never hit the save bar)

| Key | Value |
|---|---|
| `condura-theme` | `'light' \| 'dark' \| 'system' \| absent` (absent = auto) |
| `condura-energy` | `'low' \| 'balanced' \| 'high' \| 'auto' \| absent` |
| `condura-motion` | `0-100` integer (defaults to `100`) |
| `condura-grain` | `0-100` integer (defaults to `100`) |
| `condura-wake-enabled` | `'1' \| '0'` |
| `condura-wake-sensitivity` | `0-100` integer (defaults to `60`) |

### 6.4 Applied at once (CSS variables)

| Pref | Effect |
|---|---|
| theme | `document.documentElement.dataset.mode = 'light' \| 'dark'` |
| energy low | `data-energy='low'` attr + sets `condura-motion` to 0 |
| motion | `--dur-cine` = `900 × frac` ms; `--dur-slow` = `520 × frac` ms |
| grain | `--grain-opacity` = `(v/100) × 0.6` |

---

## 7. DESIGN DECISIONS

| # | Decision | Rationale | MOAT clause |
|---|---|---|---|
| D1 | **Flowing document, not dashboard.** No left nav rail, no tabs, no cards-of-cards. | A settings screen that reads like a document invites reading. A dashboard invites skimming. CLAUDE.md §20 (Converged onboarding) and the Configurator rhythm of the Ritual both favor choice-without-pressure. | §1 restraint |
| D2 | **The autonomy matrix IS the hero.** 11×3 grid gets the most vertical real estate, the only dedicated Thread + paper-card preview line, and the only `popping` keyframe. | CLAUDE.md §27 names this as "the user-defining setting." A premium product elevates the highest-leverage control to a hero, not a checkbox in a list. | §3 signature |
| D3 | **Save bar is the callback to the titlebar Thread.** When the user dirties the document, the pollen Thread on the titlebar re-appears as the pollen save bar. Same color, same motion grammar, different elevation. | MOAT §3 commits to one visual grammar used everywhere. The save bar uses the Thread as its seam + Thread as its underline. The user learns "a thread = something changed" once. | §3 thread |
| D4 | **Honest defaults.** All 11 tasks default to `warn` (CLAUDE.md §27 global default). All local prefs default to "look like the system the user already has." Theme = auto. Energy = auto. Motion = 100. Grain = 100. Wake = off. Account = signed-out. | MOAT §1 (anti-patterns): no fake enthusiasm, no celebration. Defaults that respect the user read as designed; defaults that lean on the brand read as a sales pitch. | §4 no fake enthusiasm |
| D5 | **Local UI prefs never dirty the bar.** Theme, motion, grain, energy, wake word, sensitivity — applied at once to CSS vars, persisted to localStorage. The bar is **only** for daemon-backed state. | The save bar is a promise: "I have unsaved changes that affect the agent's behavior." Appearance prefs do not affect behavior. Firing the bar for cosmetic changes dilutes the signal. | §I3 honest |
| D6 | **Two kinds of state, two visualizations.** Daemon state → save bar (commit semantics). Local prefs → instant apply (no commit semantics). | The user shouldn't have to think "do I need to save this?" — they should always know. | §I1 configure |
| D7 | **No gradient text. No emoji icons. No rainbow accents.** | DIRECTION §6 rules 1, 2, 3. All non-negotiable. | §4 anti-patterns |
| D8 | **Theme picker delegated to `ThemePicker.svelte`.** Settings owns only the row that mounts it. | The picker owns its own palette-switch choreography (clip-path circle from click origin, View Transitions API fallback). Settings should not redeclare this — `Settings.svelte` calls `applyTheme()` which simply sets `data-mode` + localStorage; the picker, if mounted in the titlebar, listens for the custom event and animates. | §1 restraint |
| D9 | **The Thread draw is the only loading affordance.** Matrix loading = pollen Thread filling across. EULA loading = Pulse + label. No spinners, no three-dots. | MOAT §2.5 (loading states must feel alive). Pulse breathes; the label teaches. | §2.5 |
| D10 | **Error states teach, not poeticize.** "Save failed" with raw `settings.lastSaveError` inline. | MOAT §2.6 — error states guide, not poeticize. | §2.6 |

---

## 8. DRIFT TABLE

| # | What was removed | Why | What replaced it |
|---|---|---|---|
| DRIFT-1 | **Inline `.seg-btn` theme row** (the original `auto / light / dark` seg with text labels) | Lacked the palette-switch choreography. Theme change was an instant snap — no `data-mode` migration story, no `--ox/--oy` origin, no view-transition morph. | `ThemePicker.svelte` (Phase 2). The row in Settings mounts the picker; Settings calls `applyTheme()` to persist. |
| DRIFT-2 | **Inline `.auto-dot` adaptive strength grid** that mixed dots + segmented buttons | Two metaphors (dot + seg) argued. | Single metaphor: 4-dot row in the new Adaptive engine section. |
| DRIFT-3 | **Save bar as a permanent footer** with "All settings auto-saved" | Misled the user — most prefs WERE auto-saved (local), but matrix changes were NOT. The bar's permanence hid the difference. | Sticky save bar that springs in ONLY when daemon state is dirty. Local prefs never trigger it. |
| DRIFT-4 | **Tabbed layout (General / Privacy / Advanced)** in the v0 sketch | Tabs read as a settings page, not a document. A document invites reading. | One flowing column with `.hair` dividers. |
| DRIFT-5 | **Modal "Reset all" affordance** | Reset is destructive and irreversible (loses daemon config). A modal hides the action behind a button. | Deferred to Phase 15B under "Advanced" — will be a row with inline confirm, not a modal. |

| # | What was added | Why | Where |
|---|---|---|---|
| ADD-1 | **Live preview line** under the autonomy matrix | A matrix of 33 dots is abstract. The preview line anchors the choice: *"Right now, for coding, Condura will **block** before acting."* The user reads a sentence, not a configuration. | `Settings.svelte:660-670` |
| ADD-2 | **Pop animation on the matrix dots** (`dot-pop`, 180ms scale) | Clicking a dot must feel like setting a switch — not like ticking a checkbox. The pop says "I heard you." | `Settings.svelte:1193-1197` |
| ADD-3 | **Hairline dividers between matrix rows** | The matrix is the only table on the page. Without hairlines, it reads as a wrap of buttons. | `Settings.svelte:1127-1138` (`:before` on `.matrix-task-col`) |
| ADD-4 | **Pollen save bar with Thread underline** | The save bar is the visual callback to the titlebar Thread. Same color, same draw, different elevation. The user learns the gesture once. | `Settings.svelte:1512-1608` |
| ADD-5 | **State-color halos on the dots (inactive ring color = state color)** | Even before the active fill lands, the three columns read as three categories. Blue/amber/green = block/warn/auto. | `Settings.svelte:1153-1197` |
| ADD-6 | **EULA inline expander** | The Legal section used to be a deep-link to a modal. The inline expander keeps the document flow and the user's scroll position. | `Settings.svelte:804-823` |
| ADD-7 | **Energy = low implies motion = 0** | The two fight. `applyEnergy('low')` calls `applyMotion(0)` to keep the contract honest. | `Settings.svelte:297-298` |
| ADD-8 | **`prefers-reduced-motion` is honored at the source** | The Motion strength slider disables entirely when the OS prefers reduced motion; the dot-pop keyframe gets a media-query override in the same file. | `Settings.svelte:118-119, 526, 1198-1200` |