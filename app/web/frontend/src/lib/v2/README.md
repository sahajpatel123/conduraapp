# Condura v2 Design System

> **A quiet companion that gets out of the way.**
> Built in parallel to v1. Never touches v1. Same IPC, same routes, same daemon — just a *premium* surface.

---

## Why v2 Exists

The v1 design system is thoughtful, but the user opened Condura and felt something was "vibe coded, no soul." This is the antidote.

| Vibe-coded tell | What v2 does instead |
|---|---|
| Glassmorphism everywhere | Matte paper-and-ink surfaces |
| Neon-on-dark | Warm low-saturation, one accent |
| Every element animated | Motion is *acknowledgment* |
| Borderless cards on gradients | Hairlines at low alpha |
| Floating everything | Strict type/grid ladder |
| Shimmer skeletons | Real percentages |

See [`../../../../docs/superpowers/specs/2026-07-01-condura-gui-redesign-design.md`](../../../../docs/superpowers/specs/2026-07-01-condura-gui-redesign-design.md) for the full creative direction.

---

## Scoping

**Every v2 token is prefixed `--v2-`.** Every v2 component sets `data-v2` on its root and scopes all selectors under `[data-v2]`. **No v2 rule leaks onto the v1 DOM.**

To activate v2 on a node, set `data-v2` on a containing element:

```svelte
<div data-v2>
  <Surface elevation={1} padding="6" radius="2">
    ...
  </Surface>
</div>
```

The CSS custom properties cascade into the subtree automatically.

---

## The Components

| Component | Role | When to reach for it |
|---|---|---|
| **Surface**   | Paper card / panel / sheet | Wrapping any grouped content |
| **Ink**       | Text (display / title / body / ui / mono) | Every text element |
| **Stack**     | Vertical rhythm | Anything vertically stacked |
| **Inline**    | Horizontal flow | Buttons, chips, key-value rows |
| **Rule**      | Hairline divider | Dividing two related surfaces |
| **Button**    | Primary / ghost / deny | Every actionable surface |
| **Switch**    | Hardware-honest toggle | Settings, anywhere a binary choice lives |

A new primitive needs a strong reason. If you can't defend it in one sentence, it goes in the consuming route, not here.

---

## CSS Loading

The three CSS files must be imported at the route that activates v2:

```svelte
<script>
  import '$lib/v2/tokens.css'   // palette, type, spacing, radii, shadows, z
  import '$lib/v2/motion.css'   // easings, durations, keyframes
  import '$lib/v2/reset.css'    // scoped element resets
</script>
```

Not imported globally — only on routes that use v2. Keeps the v1 bundle clean.

---

## Wiring `V2Preview.svelte` Into `App.svelte`

The full integration is a 6-line additive change in
`app/web/frontend/src/App.svelte`. It does NOT touch any v1 route,
v1 component, or any IPC/store code — only registers a new
`#/dev/v2-preview` route alongside `#/dev/components` and `#/dev/v1`.

1. **Import the preview route.** Find the existing `import V1` line
   in `App.svelte` and add this right below it:

   ```ts
   import V2Preview from './lib/routes/dev/V2Preview.svelte'
   ```

2. **Add the route branch in `hashToRoute`.** Find the existing
   `if (hash.startsWith('#/dev/v1')) return 'v1-preview'` line and
   add this right below it:

   ```ts
   if (hash.startsWith('#/dev/v2-preview')) return 'v2-preview'
   ```

3. **Add the case to `routeTitle`.** Find `'v1-preview' => 'v1 Design Preview'`
   and add this after it:

   ```ts
   route === 'v2-preview' ? 'v2 Design Preview' :
   ```

4. **Add the `<svelte:component>` branch in the template.** Find
   the existing `{:else if route === 'v1-preview'} <V1 />` block
   and add this right below it:

   ```svelte
   {:else if route === 'v2-preview'}
     <V2Preview />
   ```

5. **Update the type union.** Find
   `function hashToRoute(hash: string): RouteId | 'dev-components' | 'v1-preview'`
   and replace with:

   ```ts
   function hashToRoute(hash: string): RouteId | 'dev-components' | 'v1-preview' | 'v2-preview'
   ```

For the **floating interview preview** (`V2InterviewDemo.svelte`), repeat the same pattern with `V2InterviewDemo` and `#/dev/v2-interview`.

For the **chat surface preview** (`V2ChatDemo.svelte`), repeat with `V2ChatDemo` and `#/dev/v2-chat`.

For the **full app shell preview** (`V2ShellPreview.svelte`), repeat with `V2ShellPreview` and `#/dev/v2-shell`. This mounts Sidebar + StatusBar + ChatSurface together as a coherent app.

For the **gatekeeper consent preview** (`V2ConsentDemo.svelte`), repeat with `V2ConsentDemo` and `#/dev/v2-consent`. This shows the wax-seal-on-a-letter modal across four scenarios (read / write / network / destructive).

After saving, navigate to any of `#/dev/v2-preview`, `#/dev/v2-interview`, `#/dev/v2-chat`, `#/dev/v2-shell`, or `#/dev/v2-consent` in the dev server (`bun run dev` or `wails dev`). Each renders the respective piece of the v2 system.

**Critical:** none of these changes touch a single line of v1 code
or any IPC/store wiring. They are additive only.

---

## The Colors (Quick Reference)

| Token | Value | Role |
|---|---|---|
| `--v2-paper`       | `#F7F4EE` | canvas |
| `--v2-paper-2`     | `#EFEAE0` | recessed |
| `--v2-surface`     | `#FFFFFF` | elevated card |
| `--v2-ink`         | `#1B1A17` | primary text |
| `--v2-ink-2`       | `#4A463E` | secondary text |
| `--v2-ink-3`       | `#8A847A` | tertiary |
| `--v2-rule`        | `#D9D2C2` | hairlines |
| `--v2-accent`      | `#C18A4A` | **the one color** |
| `--v2-accent-ink`  | `#7A4F1E` | accent text |
| `--v2-signal-go`   | `#5C7F4A` | success |
| `--v2-signal-warn` | `#B07A2E` | cautious |
| `--v2-signal-stop` | `#A84A3F` | destructive |

---

## The Motion (Quick Reference)

| Token | Value | Use |
|---|---|---|
| `--v2-ease-out-soft` | `cubic-bezier(.22, 1, .36, 1)` | arrival |
| `--v2-ease-in-honest` | `cubic-bezier(.55, .06, .68, .19)` | exit |
| `--v2-ease-spring` | `cubic-bezier(.5, 1.4, .4, 1)` | state ack |
| `--v2-ease-settle` | `cubic-bezier(.16, 1, .3, 1)` | shadow stage |
| `--v2-dur-fast` | `140ms` | hover, focus |
| `--v2-dur-mid` | `280ms` | panel entry |
| `--v2-dur-slow` | `520ms` | overlay arrival |
| `--v2-dur-cinematic` | `900ms` | first-paint |

**Rule:** every motion must mean one of {arrival, departure, ack, attention}.

---

## What's Next

After the user wires `V2Preview` into `App.svelte` and confirms the
foundation feels premium, the next iterations build:

1. **Showcase** — chat surface, overlay arrival, the actual first-time
   floating interview wired to `onboarding.*` RPCs.
2. **Surfaces** — sidebar, settings, status bar, audit, consent modal,
   then channels/delegation/hub/replay/sync/skills/about.
3. **Polish** — sound set (opt-in), focus ring tuning, micro-typography.

See the spec for the full roadmap.

---

**The mission, restated:** make AI useful to every ordinary person, on
every computer, for free. The GUI is the handshake. Make it count.
