# Synaptic — Design v1 Migration Playbook

> **Status:** Playbook. Companion to `docs/design-v1-redesign.md` (the locked spec).
>
> The v1 design system is complete: 35 components, 8 token files, full spec §12 coverage. The redesign is visible at `/dev/v1`. **Nothing in the production app has been replaced yet.** This document is the playbook for doing that replacement safely.

---

## The Core Migration Principle

> **Same data layer, new visual presentation.**

Every existing route (`Chat`, `Settings`, `Skills`, `Hub`, `Audit`, `Replay`, `Channels`, `Delegation`, `Sync`, `About`) currently:
- Uses the same stores (`conversation.svelte`, `daemon.svelte`, `settings.svelte`, etc.)
- Uses the same IPC layer (`ipc.client`, `ipc.types`)
- Calls the same daemon RPCs

The migration preserves ALL of that. Only the visual layer changes — primitive components, CSS, and layout.

**Concretely:**
- `<button>` → `<Button>` (v1)
- `<input>` → `<Input>` (v1)
- Custom CSS → semantic token CSS (`var(--content-primary)`, `var(--space-3)`, etc.)
- Card layouts → `<Surface>` or `<Card>` (v1)
- Status badges → `<Pill>` (v1)
- Empty states → `<EmptyState>` (v1)

---

## Migration Order (11 routes, lowest risk first)

| # | Route | Lines (orig) | Migration risk | Why this order |
|---|---|---|---|---|
| 1 | `About.svelte` | smallest | Low | Static info page — no IPC, easy to verify |
| 2 | `Channels.svelte` | medium | Low | List view — uses Sidebar pattern, easy to verify |
| 3 | `Hub.svelte` | medium | Low | Mostly list — same pattern as Channels |
| 4 | `Skills.svelte` | medium | Low | Cards-on-grid — Card primitive is direct replacement |
| 5 | `Delegation.svelte` | medium | Medium | Spawns sub-agents — IPC-heavy, watch for state bugs |
| 6 | `Sync.svelte` | medium | Medium | Pairing flow has visual states — careful with PairingModal |
| 7 | `Replay.svelte` | medium | Medium | Timeline scrub — keep time precision |
| 8 | `Audit.svelte` | large | Medium | Dense table — AgentActionLog is the pattern |
| 9 | `Settings.svelte` | very large | High | Most complex; the SettingsPane is already built |
| 10 | `Chat.svelte` | very large | High | Most complex; ChatV1 is already built as template |
| 11 | `App.svelte` + `main.ts` | central | High | Switches the route imports — final step |

---

## Per-Route Migration Pattern (the recipe)

For each route, follow this 5-step recipe:

### Step 1 — Identify the data layer
```svelte
<script lang="ts">
  // KEEP — these imports don't change
  import { conversation } from '../stores/conversation.svelte';
  import { daemon } from '../stores/daemon.svelte';
  import { ipc } from '../ipc/client';
  import type { ProviderInfo } from '../ipc/types';

  // KEEP — all reactive state, derived values, event handlers
  let inputText = $state('');
  // ... etc
</script>
```

### Step 2 — Replace the primitive imports
```svelte
<script lang="ts">
  // REMOVE
  // import Button from '../components/ui/Button.svelte';
  // import Card from '../components/ui/Card.svelte';

  // ADD
  import Button from '$components/v1/Button.svelte';
  import Card from '$components/v1/Card.svelte';
  import Stack from '$components/v1/Stack.svelte';
  import Hairline from '$components/v1/Hairline.svelte';
  import Pill from '$components/v1/Pill.svelte';
</script>
```

### Step 3 — Replace primitive usage
| v0 pattern | v1 replacement |
|---|---|
| `<button class="primary">` | `<Button variant="primary">` |
| `<button class="secondary">` | `<Button variant="secondary">` |
| `<input type="text">` | `<Input variant="sans" />` |
| `<input type="text" class="mono">` | `<Input variant="sans" monospace />` |
| `<div class="card">` | `<Card>` or `<Surface>` |
| `<span class="badge success">` | `<Pill variant="success" label="..." />` |
| `<div class="empty">` | `<EmptyState primary="..." voice="mono" />` |
| `<div class="divider">` | `<Hairline />` |
| Custom `display: flex; flex-direction: column; gap: 16px` | `<Stack gap="4">` |
| Custom `display: flex; flex-direction: row; gap: 12px` | `<Inline gap="3">` |

### Step 4 — Replace CSS color values with semantic tokens
| v0 value | v1 token |
|---|---|
| `#FBF8F2` or `var(--bg)` | `var(--surface-base)` |
| `#0E1014` or `var(--text)` | `var(--content-primary)` |
| `#6E3AFF` or `var(--accent)` | `var(--content-accent)` or `var(--action-primary-idle-bg)` |
| `16px` | `var(--space-4)` |
| `24px` | `var(--space-6)` |
| `rgba(255,255,255,0.06)` | `var(--shadow-1)` |
| `border-radius: 8px` | `var(--radius-md)` |
| `font-family: var(--font-mono)` | already correct — keep |

### Step 5 — Verify (this is where the type-check helps)
- `npm run check` should report 0 errors for the migrated route
- Run the route, click every interaction, watch for visual regressions
- Check dark mode by toggling `<html data-mode="dark">`

---

## Specific Route Migrations

### Chat.svelte → ChatV1.svelte (template already built)

`ChatV1.svelte` at `app/web/frontend/src/lib/routes/ChatV1.svelte` is the canonical example. It:
- Reuses `conversation.svelte`, `settings.svelte`, `ipc.providers.list`
- Uses v1 primitives: ChatSurface, EmptyState, Input, Button, Chip, Surface, Hairline, Inline
- Reduced from 703 lines to 268 lines (62% reduction)
- Same functionality, new visual presentation

To activate: change the route import in `App.svelte` from `./lib/routes/Chat.svelte` to `./lib/routes/ChatV1.svelte`.

### Settings.svelte → SettingsPane.svelte (template already built)

The full SettingsPane exists at `app/web/frontend/src/lib/components/v1/SettingsPane.svelte`. It has 7 sections wired (replay, adaptive, permissions, hotkey, autonomy, backup, account) per spec §11.3.

The existing `Settings.svelte` should be REPLACED by routing to a route that mounts SettingsPane directly (it's already a full-window component).

### Audit.svelte → AgentActionLog.svelte (template already built)

`AgentActionLog.svelte` at `app/web/frontend/src/lib/components/v1/AgentActionLog.svelte` is the dense replay table per spec §8.4.

### Sidebar.svelte

The existing `Sidebar.svelte` (in `app/web/frontend/src/lib/components/`) should be REPLACED by the v1 Sidebar at `$components/v1/Sidebar.svelte`. Either:
- Rename the v1 file to `Sidebar.svelte` and update all imports
- Update each route's import path

### Style.css + tokens.css

The old `app/web/frontend/src/lib/styles/tokens.css` and the existing color/dark-mode CSS in `style.css` should be REPLACED by:
1. Remove `import './lib/styles/tokens.css'` from style.css
2. Add `import './lib/tokens/primitives.css'` (Layer 1)
3. Add `import './lib/tokens/semantic.css'` (Layer 2)
4. Add `import './lib/tokens/motion.css'` (Layer 3)
5. Add `import './lib/tokens/themes/system.css'` (Layer 4)
6. Initialize theme on app boot: `import { initTheme } from '$tokens/themes'; initTheme();`

The v0 variables (`--bg`, `--text`, `--accent`, etc.) will be unmapped. Components that consume them will fall back to defaults. After migration, the v0 tokens.css can be deleted.

---

## Density Rules (per spec §2.1)

The v1 design uses **dual-mode density**. Pick the right mode per surface:

| Surface type | Density | Examples |
|---|---|---|
| List / table / dense data | **Compact** (Linear-grade) | Audit, AgentActionLog, history, settings lists |
| Reading room | **Spacious** (Things-grade) | Chat, Onboarding, empty states, settings reading |
| Command surface | **Medium** (560px wide) | CommandSurface overlay |
| Sidebar nav | **Compact** (always) | Sidebar, nav |

If a route currently has a single density throughout, audit whether it should split.

---

## Anti-Pattern Checks (per spec §15)

After each migration, scan the route for these anti-patterns:
- ❌ Drop shadows on cards (use hairlines instead)
- ❌ Glass anywhere except the CommandSurface
- ❌ Spinners (use heartbeat that scales with pause duration)
- ❌ Avatars with faces (use Pulse + initials)
- ❌ Emojis in primary surfaces (decorative only)
- ❌ Inter / Geist / Manrope fonts
- ❌ Rainbow gradients
- ❌ Plum backgrounds (plum is accent only, ≤5% of screen)

---

## Token Migration Cheat-Sheet

| v0 token | v1 token | Where it appears |
|---|---|---|
| `--bg` | `--surface-base` | Page background |
| `--bg-elevated` | `--surface-raised` | Cards, panels |
| `--bg-sunken` | `--surface-sunken` | Recessed sections |
| `--text` | `--content-primary` | Primary text |
| `--text-muted` | `--content-tertiary` | Tertiary text, placeholders |
| `--text-faint` | `--content-muted` | Very subtle text |
| `--accent` | `--content-accent` | The plum accent (text) |
| `--accent-bg` | `--action-primary-idle-bg` | Primary button bg |
| `--border` | `--border-default` | Default borders |
| `--border-strong` | `--border-strong` | Emphasized borders |
| `--shadow-sm` | `--shadow-1` | Card shadow (often replaced by hairline) |
| `--shadow-md` | `--shadow-3` | Popover shadow |

For values not in this table, search `primitives.css` for the closest match, or add a new semantic token to `semantic.css` (and update the synthesis spec to lock it).

---

## Pre-Flight Checklist (before declaring a route migrated)

- [ ] No raw hex values remain in component CSS (except where documented as `// primitive-escape`)
- [ ] All shadows are either `--shadow-1` (hairline-equivalent) or `--shadow-3` (popovers). No `--shadow-2` or `--shadow-4` in normal layout.
- [ ] Spacing values use `--space-*` tokens (not raw px)
- [ ] Radius values use `--radius-*` tokens
- [ ] All text colors use `--content-*` tokens (not raw colors)
- [ ] All interactive elements have visible focus rings
- [ ] `prefers-reduced-motion` fallbacks work (test in browser dev tools)
- [ ] Dark mode renders correctly (toggle `<html data-mode="dark">`)
- [ ] High-contrast mode doesn't break (where applicable)
- [ ] No new colors, fonts, durations, or easing curves introduced (else update spec first)

---

## Final Step: App.svelte Wiring

After all 11 routes migrated, the final step is to update `App.svelte` (or `main.ts`) to:
1. Initialize the theme on boot (`initTheme()`)
2. Render the v1 Sidebar instead of the v0 one
3. Render the v1 StatusBar as an overlay
4. Render the v1 CommandSurface as a Cmd+K overlay
5. Keep the existing route paths so deep-links work

The dev preview at `/dev/v1` already demonstrates this full shell. The production `App.svelte` should mirror it.

---

## When to Stop and Ask

If during migration you discover:
- A new color is needed (don't add it; update the spec first)
- A new motion timing is needed (don't add it; update the spec first)
- A new font pairing is needed (don't add it; update the spec first)
- An existing v0 component's functionality can't be expressed with v1 primitives (that's a v1 gap; flag it)

**The spec is the contract. Any deviation is a violation of §1–§15.**

---

**This playbook + `docs/design-v1-redesign.md` = the complete migration story.**

The hardest migration is the first one (Chat, which we've already done as ChatV1.svelte). Once that pattern is internalized, the rest are mechanical.

Estimated migration effort per route: 30–90 minutes depending on complexity. Total estimated: ~10-15 hours of focused work.