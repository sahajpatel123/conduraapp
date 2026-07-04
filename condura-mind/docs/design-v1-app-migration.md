# Synaptic — App.svelte v1 Migration Recipe

> **Status:** Companion to `docs/design-v1-migration.md`. App.svelte is the central routing file; this recipe documents the minimal changes to expose the v1 design without breaking the existing app.

---

## Current App.svelte Structure

```
┌─────────────────────────────────────────────────────────────┐
│  App.svelte (v0)                                             │
│  ─────────────                                                │
│  Hash-based routing (/#/, #/chat, #/settings, etc.)          │
│  ├── Sidebar (v0)              ← replace with v1            │
│  ├── TitleBar (v0)             ← can stay (small)            │
│  ├── <main>                    ← routes live here            │
│  │   ├── Chat (v0)             ← switch to ChatV1            │
│  │   ├── Settings (v0)         ← replace with SettingsPane   │
│  │   ├── Audit, Replay, etc.   ← migrate per playbook        │
│  │   └── DevComponents         ← ADD /dev/v1 route          │
│  ├── StatusRail (v0)           ← replace with v1 StatusBar   │
│  ├── Toasts                    ← keep, migrate later         │
│  ├── LiveTranscript             ← keep, migrate later         │
│  ├── OverlayPrompt             ← keep, migrate later         │
│  ├── ConsentModal (v0)         ← replace with v1 ConsentModal│
│  ├── CommandPalette (v0)       ← replace with v1 CommandSurface│
│  └── OnboardingWizard (v0)    ← replace with v1 OnboardingWizard│
└─────────────────────────────────────────────────────────────┘
```

## Two Migration Strategies

### Strategy A: Additive (zero risk, recommended first step)

Add a NEW route for `/dev/v1` alongside existing routes. The v0 app continues working; the v1 design becomes reachable at `#/dev/v1`. Use this for review.

**Changes:**

1. Add the import at the top:
   ```svelte
   import V1 from './lib/routes/dev/V1.svelte'
   ```

2. Extend the route $derived to include the new path:
   ```ts
   let route = $derived(
     currentHash.startsWith('#/dev/v1') ? 'v1-preview' :  // ← NEW
     currentHash.startsWith('#/settings') ? 'settings' :
     // ... existing routes unchanged
   )
   ```

3. Extend routeTitle similarly (or use a static label for v1-preview).

4. Add the route to paletteItems:
   ```ts
   { id: 'v1', label: 'v1 design preview', hint: 'See the new design system', group: 'Dev', onselect: () => { window.location.hash = '#/dev/v1' } },
   ```

5. Add the route render in the main switch:
   ```svelte
   {#if route === 'v1-preview'}
     <V1 />
   {:else if route === 'dev-components'}
     <DevComponents />
   {:else if route === 'chat'}
     <Chat />
   {/if}
   ```

**Total diff: ~5 lines.** Zero impact on existing functionality. The v1 design is reachable via the command palette (⌘K → "v1 design preview") or by navigating to `#/dev/v1`.

### Strategy B: Full replacement (production migration)

Replace the v0 shell with the v1 shell. This is the "ship it" state.

**Changes:**

1. **Replace `import Sidebar from './lib/components/Sidebar.svelte'`** with `import Sidebar from '$components/v1/Sidebar.svelte'`. The v1 Sidebar accepts `active`, `collapsed`, `onnavigate`, `ontoggle` props.

2. **Replace the inline route rendering** with a state-driven activeRoute + v1 Sidebar's `onnavigate` handler that sets a state variable (in addition to setting `window.location.hash`).

3. **Replace `StatusRail` import** with `import StatusBar from '$components/v1/StatusBar.svelte'`. Render in the same fixed position.

4. **Replace `CommandPalette` import** with `import CommandSurface from '$components/v1/CommandSurface.svelte'`. Render as an overlay when `paletteOpen` is true.

5. **Replace `ConsentModal` (v0) import** with `import ConsentModal from '$components/v1/ConsentModal.svelte'`. Same component name, different API; check the new API surface.

6. **Replace `OnboardingWizard` (v0) import** with `import OnboardingWizard from '$components/v1/onboarding/OnboardingWizard.svelte'`. The v1 version takes `oncomplete` instead of `onComplete`.

7. **Update style.css imports** to swap v0 tokens for v1 tokens:
   ```css
   /* REMOVE */
   /* @import './lib/styles/tokens.css'; */
   /* ADD */
   @import './lib/tokens/primitives.css';
   @import './lib/tokens/semantic.css';
   @import './lib/tokens/motion.css';
   @import './lib/tokens/themes/system.css';
   ```

8. **Initialize theme** at the top of the script:
   ```ts
   import { initTheme } from '$tokens/themes';
   initTheme();
   ```

9. **Swap route components** per the playbook:
   - `Chat` → `ChatV1`
   - `Settings` → `SettingsPane` (mounted as a route directly — it's a full-window component)
   - `Audit` → AgentActionLog + a wrapper route
   - Other routes follow the playbook per-route

10. **Delete or archive v0 files** (after verification):
    - `lib/styles/tokens.css` → delete (replaced by v1 tokens)
    - `lib/components/Sidebar.svelte` → delete (replaced by v1)
    - `lib/components/ui/Button.svelte`, etc. → delete (replaced by v1)
    - `lib/components/StatusRail.svelte` → delete
    - `lib/components/ui/CommandPalette.svelte` → delete
    - `lib/components/ConsentModal.svelte` → delete
    - `lib/components/OnboardingWizard.svelte` → delete

---

## API Differences to Watch

Some v0 components have the same name as v1 components but different APIs:

| v0 API | v1 API | Where |
|---|---|---|
| `OnboardingWizard onComplete={fn}` | `OnboardingWizard oncomplete={fn}` | App.svelte |
| `Sidebar` (no props) | `Sidebar active={route} collapsed={...} onnavigate={fn}` | App.svelte |
| `CommandPalette bind:open={...}` | `CommandSurface mode={...} contextChips={...}` | App.svelte |
| `StatusRail` (no props) | `StatusBar activeTask={...} queuedCount={...} agentState={...}` | App.svelte |
| `ConsentModal` (internal state) | `ConsentModal verb={...} target={...} onapprove ondeny` | App.svelte |

---

## Order of Operations

The full replacement is high-risk. Do it in this order:

1. **Strategy A first** (additive `/dev/v1` route, ~5 lines). Verify the new design works.
2. **Strategy B in micro-steps**, each verified:
   - Step B.1: Swap Sidebar (5 lines)
   - Step B.2: Swap StatusRail → StatusBar (3 lines)
   - Step B.3: Swap CommandPalette → CommandSurface (10 lines)
   - Step B.4: Swap OnboardingWizard (3 lines)
   - Step B.5: Swap ConsentModal (5 lines)
   - Step B.6: Update style.css imports (4 lines)
   - Step B.7: Initialize theme (1 line)
   - Step B.8: Per-route migrations (Chat → ChatV1 first)
   - Step B.9: Delete v0 files (last, after all references gone)

After each step: `npm run check`, run the app, smoke-test the affected surface.

---

## Why the Additive Strategy First

Strategy A is intentionally additive. It lets reviewers see the new design WITHOUT disturbing the production app. If the redesign is rejected, no rollback is needed — just delete the route addition. If the redesign is approved, Strategy B becomes a sequence of small, verifiable migrations.

This is the same "expand-contract" pattern used in database schema migrations: add the new, verify it works, then remove the old.

---

## Estimated Effort

| Strategy | Lines changed | Risk | Time |
|---|---|---|---|
| A (additive) | ~5 | None | 15 min |
| B (full replacement) | ~50 lines across 7 files | High but incremental | 2-3 hours |

The full B migration should take less than half a day, including per-route work. The playbook covers each route; App.svelte is the central wiring.

---

**Once Strategy A is live, the v1 design is reachable at `#/dev/v1` via the command palette (⌘K → "v1 design preview") or by typing the hash directly.**