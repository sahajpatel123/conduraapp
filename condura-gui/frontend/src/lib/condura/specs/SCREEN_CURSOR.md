# SCREEN_CURSOR

> The pixel cursor + the trailing halo. One screen in the user's eye, three artifacts on disk: the SVG data-URI in `condura.css:341–349`, the hover-region detection in `Cursor.svelte`, and the `--pollen` ring toggle bound to `body[data-hover='1']`.

---

## 1. LAYOUT & CONTENT

| Region | What lives there | Why |
|---|---|---|
| **Pixel cursor (OS-level)** | 20×20 SVG quill nib, hot spot `(3, 2)`. Light: ink `#14110B` fill, pollen `#C97B2E` 1px stroke. Dark: paper `#F0EDE5` fill, pollen `#E0A862` stroke. Default fallback `, auto`. | The ink quill + warm paper metaphor extended; the cursor is the first thing the user touches, so it must feel like the same medium as the page. Hot spot `(3, 2)` puts the click point at the nib tip. |
| **Hover-region cursor** | 24×24 SVG. `4r` pollen fill at center, `9r` synapse ring around it. Hot spot `(12, 12)`. Fallback `, pointer`. | Activated only on `body[data-hover='1']`. Signals "this surface catches your mouse." Direct fulfillment of MOAT.md §5.1. |
| **Trailing halo (DOM)** | `position: fixed; 14×14` round div. Border `1px solid var(--synapse-glow)`. Background pollen 10% via `color-mix(in oklab, var(--pollen) 10%, transparent)`. `pointer-events: none; z-index: var(--z-max)`. | Decorative sibling to the pixel cursor, lag-target via rAF at 60fps. Brightens to 34×34 + pollen-tinted on hover. |
| **`data-hover` attribute on `<body>`** | `dataset.hover = '0'` or `'1'`, toggled by `pointerover` listener in `Cursor.svelte`. | The single source of truth the CSS reads to swap cursor variants. |

> The cursor is NOT a layout region — there is no DOM scroll viewport, no sidebar, no padding budget. It is a single raster pixel + a single fixed-position div layered on top of every Shell route.

---

## 2. STATE MATRIX

| State | Trigger | Pixel cursor | Halo | `body[data-hover]` |
|---|---|---|---|---|
| **default** | mouse over paper / content | ink quill (light) or paper quill (dark), `auto` fallback | hidden (`opacity: 0`) until first `pointermove`, then `opacity: 0.55`, synapse-glow border | `0` |
| **hover-interactive** | mouse over `button:not(.no-tactile)`, `.tactile`, `[role='button']:not(.no-tactile)`, `summary:not(.no-tactile)`, `.choice`, `.nav-item`, `.dock-item`, `.thread-link`, `input`, `textarea`, `a`, `[data-hoverable]` | pollen dot + synapse ring (24×24, `pointer` fallback) | `width/height: 34px`, `opacity: 0.7`, pollen border | `1` |
| **hover-text** | mouse over text input / textarea | native I-beam (OS owns it — no override) | trails as in default; ring does not apply | `1` only if input also matches a `[data-hoverable]` ancestor |
| **hover-drag** | mousedown on drag handle (CSS swaps) | native `grab` → `grabbing` | trails as in default | `1` |
| **hover-disabled** | mouse over `[disabled]` or `[aria-disabled='true']` | native `not-allowed` (CSS rule in `condura.css:328–331` — `cursor: not-allowed !important;`) | trails as in default | `0` (disabled elements excluded from selector list) |
| **hover-loading** | n/a — cursor holds current shape | unchanged | unchanged | unchanged. Per DIRECTION.md §6 rule 7 ("No spinner loaders"): loading is a Thread drawing in on the surface, never a cursor glyph. |
| **tab hidden / off-viewport** | `document.hidden === true` OR `IntersectionObserver` reports `isIntersecting: false` | unchanged | rAF loop suspended (`running = false`), no transform writes | unchanged |
| **`prefers-reduced-motion: reduce`** | `matchMedia` matches at mount | unchanged | component unmounts the rAF loop entirely (`if (reduce) return;`) — no halo rendered | unchanged |

The hover-region detection is exhaustive for the interactive set: every element that responds to a click owns the `pointer` cursor and the pollen ring. Disabled, native-text, and native-grab states are delegated to the OS so the user gets the affordance they expect.

---

## 3. MOTION CHOREOGRAPHY

The pixel cursor itself does not animate. A 20×20 SVG data-URI cannot crossfade on retina without a stack of layered cursors and a requestAnimationFrame swap-and-restore scheme that every platform-specific quirk would defeat. The cursor **swaps instantly** when `body[data-hover]` toggles. This is a feature, not a gap — the cursor is muscle memory, and muscle memory needs no transition.

| Motion | Property | Duration | Easing | Trigger |
|---|---|---|---|---|
| **Pixel cursor swap (default → hover)** | `body[data-hover='1']` selector flips the `cursor: url(...)` | instantaneous (CSS layer) | n/a | `pointerover` enters an interactive region |
| **Pixel cursor swap (hover → default)** | `body[data-hover='0']` | instantaneous | n/a | `pointerover` leaves the region |
| **Halo position lerp** | `transform: translate(tx − 7, ty − 7)` | per rAF, factor `0.16` (60fps) | linear (lerp) | `pointermove` |
| **Halo reveal on first move** | `opacity: 0 → 0.55` | 0.25s | `--ease` (cubic-bezier(0.22, 1, 0.36, 1)) | first `pointermove` adds `.on` |
| **Halo hover-state grow** | `width/height: 14 → 34`, `border-color → --pollen`, `background 10% → 14% pollen`, `opacity 0.55 → 0.7` | 0.25s | `--ease` | `pointerover` matches interactive selector |
| **Halo hover-state shrink** | reverse of above | 0.25s | `--ease` | leave interactive region |
| **Halo rAF pause** | `running = false; raf canceled` | instantaneous | n/a | `visibilitychange` to `hidden` OR `IntersectionObserver` reports `isIntersecting: false` |
| **Halo rAF resume** | `running = true; raf = requestAnimationFrame(tick)` | instantaneous | n/a | `visibilitychange` to `visible` AND in-viewport |

The reduced-motion contract is held entirely by the JavaScript mount gate (`if (reduce) return;` at `Cursor.svelte:11`). Under `prefers-reduced-motion: reduce`, the halo is never mounted at all — there is nothing to animate, nothing to pause.

---

## 4. KEYBOARD

N/A. The cursor is a mouse-only artifact. The keyboard navigation surface is the `:focus-visible` halo (4px pollen halo + 1px synapse inset, defined at `condura.css:299–302` and shaped per surface by MOAT.md §6 rule 6). The two systems do not overlap: focus drives the keyboard, the cursor drives the mouse. When a surface is both keyboard-focusable *and* mouse-interactive, focus takes priority in the visual hierarchy (focus halo is 4px around the geometry; pollen ring is 24×24 around the cursor) and both render simultaneously without interference.

---

## 5. COMPONENTS USED

| Component | Role | File | How it produces the cursor state |
|---|---|---|---|
| `Cursor.svelte` | The trailing halo + the `data-hover` toggler | `condura/Cursor.svelte` | Mounts the fixed halo div, listens to `pointermove` + `pointerover` document-wide, lerps the halo, and toggles `document.body.dataset.hover` based on whether the `event.target`'s closest ancestor matches the interactive selector list. Skips entirely when `prefers-reduced-motion: reduce`. Suspends on tab-hide / off-viewport. |
| `condura.css` (lines 341–349) | The pixel cursor | `condura/condura.css` | Three CSS rules: (1) `body { cursor: url("...quill... ") 3 2, auto; }` for default, (2) `:root[data-mode='dark'] body` overrides for dark mode quill, (3) `body[data-hover='1']` swaps to the pollen-target cursor with `pointer` fallback. |
| `condura.css` (lines 328–331) | Disabled-state cursor | `condura/condura.css` | `[disabled], [aria-disabled='true'] { cursor: not-allowed !important; }` — the single override that beats the data-URI cursor, by `!important`. |
| `Button.svelte` (and every tactile surface) | The interactive set | `condura/Button.svelte` and the `.tactile` / `button` / `[role='button']` / `summary` selectors at `condura.css:309–325` | The `Cursor.svelte` selector list covers these via `button:not(.no-tactile)` / `.tactile` / `[role='button']:not(.no-tactile)`. When mouse enters one, `Cursor.svelte`'s `pointerover` handler matches and flips `data-hover='1'`. |
| `input` / `textarea` / `a` / `select` rows / `[data-hoverable]` | The interactive set, expanded | every surface in `condura/` | Same mechanism — the selector list in `Cursor.svelte:30–32` covers them all. |
| `--synapse-glow`, `--pollen`, `--z-max`, `--ease` design tokens | Halo border/background, layer z-index, transition easing | `condura/condura.css:151, 156, 129, 108` | Read directly in the `<style>` block of `Cursor.svelte`. The halo's colors are brand tokens, never raw RGB, with `color-mix(in oklab, ...)` for the translucent wash. |

The data flow is one-way: `pointerover` → Cursor.svelte → `body[data-hover]` → CSS rule → browser pixel cursor. There is no Svelte store, no IPC, no async.

---

## 6. DATA FETCHED

None. The cursor is a pure CSS + ~50 lines of JavaScript effect with no IPC, no Svelte stores, no async fetch, no LLM call, no DOM query that hits the network. The `pointermove` / `pointerover` listeners attach to `document` (and `addEventListener` on the window) and read `event.target` synchronously. The `IntersectionObserver` is local. There is no server round-trip at any point in the cursor lifecycle.

---

## 7. DESIGN DECISIONS

| Decision | Rationale |
|---|---|
| **Pixel quill + pollen tip as the default cursor.** | Direct fulfillment of DIRECTION.md §1 ("paper notebook, not app"). The cursor is the first surface the user touches, so it must feel like the same medium as the page — warm ink on warm paper, with a single warm-amber accent at the tip. |
| **Hover-region swaps to pollen dot + synapse ring.** | MOAT.md §5.1 ("The cursor on hoverable surfaces changes to a pollen ring"). The plumbing existed in `condura.css` but no `Cursor.svelte` was setting `data-hover`. The decision is to ship it as a *brand* signal: catching the mouse is the surface admitting it can be caught. |
| **Never a spinner.** | DIRECTION.md §6 rule 7 + MOAT.md §4 anti-pattern 7. Loading shows a Thread drawing in on the surface (`drawthread` keyframe) or a hash-route pulse — never a cursor glyph. The cursor holds its current shape through every loading state. |
| **Never a system arrow.** | The default cursor IS the brand. Replacing it with a system arrow during scroll, drag, or `wait` would break the metaphor. Drag handles use the native `grab`/`grabbing` cursors only because the OS affordance is correct; everything else stays brand. |
| **Hot spot at `(3, 2)`.** | That puts the click point at the geometric tip of the quill nib — where the ink would meet the paper. A 20×20 cursor with hot spot `(10, 10)` would feel like clicking with the nib's middle, which feels wrong. |
| **Halo, not larger cursor.** | The pixel cursor can't grow or fade; retina cursors don't crossfade. The halo is the lagging, brightening *companion* — it doesn't replace the cursor, it trails it. The two are intentionally different layers with different behaviors. |
| **Brand colors used directly in data-URI.** | The SVG data-URI is the only place in `condura/` that uses raw hex (`%2314110B`, `%23C97B2E`, `%23F0EDE5`, `%23E0A862`) — flagged in the CSS comment at `condura.css:339–340`. CSS variables inside `url(...)` data-URIs are still poorly supported across browsers, so the light/dark cursor variants must be hand-inlined. |
| **rAF paused on `document.hidden` AND off-viewport.** | An always-rAF cursor halo in a backgrounded tab is a known battery sink (per MOAT.md §1.4 the cursor itself is opt-in by design). The pause-then-resume rAF pattern keeps total CPU cost at zero during idle, matches the energy budget from `app/web/frontend/src/lib/tokens/motion.css`, and never drops a frame in interactive use. |
| **`prefers-reduced-motion: reduce` short-circuits at mount.** | The halo is decorative. A reduced-motion user gets the pixel cursor only — no lag, no grow, no border-color transition. The contract is held by a single line in `Cursor.svelte:11`, not by a per-component media query (per DIRECTION.md §5 reduced-motion contract). |
| **Halo z-index is `--z-max` (9999).** | The halo must always be the topmost pixel; if a modal, sheet, or toast slides over it, the user loses the cursor-affordance signal. `--z-max` puts it above the highest z-indexed token in the design system. |

---

## 8. DRIFT TABLE

| Item | Removed | Added |
|---|---|---|
| v1 (pre-Wave-2) | The 9-keyframe `voidHold` ritual arrival choreography was firing the quill as a transient decoration. | The quill is now a permanent default; the cursor IS the ritual's signature. |
| v1 | Cursor was rendered unconditionally on every OS without checking `prefers-reduced-motion`. | Reduced-motion users get no halo — `if (reduce) return;` at mount. |
| v1 | Cursor was a single hardcoded SVG without dark-mode variant. | Two light/dark cursor data-URIs with paper-on-paper swap; one pollen ring variant driven by `data-hover`. |
| v1 | No hover-region cursor existed (MOAT.md §5.1 said "plumbing exists, directive doesn't"). | `body[data-hover]` toggle wired through `Cursor.svelte`'s `pointerover` listener with an exhaustive interactive selector list. |
| v1 | `Cursor.svelte` ran rAF unconditionally — a backgrounded tab wasted ~108k style writes per 30 minutes. | `document.hidden` + `IntersectionObserver` pause the rAF loop; resume on tab focus / re-entry. |
| v1 | Default cursor hot spot was `(0, 0)` — clicks registered in the upper-left of the cursor, off the nib. | Hot spot moved to `(3, 2)` so the click point reads as the nib tip on paper. |
| v1 | The `[disabled]` / `[aria-disabled='true']` rule used `cursor: default` as the override — wrong affordance. | `cursor: not-allowed !important;` is the canonical "intentionally not now" affordance per `condura.css:328–331`. |
| v1 | The cursor was a 9-12px generic pointer, identical to every Electron app. | A 20×20 ink quill with a 1px pollen stroke is brand-distinct against any background — even dark mode. |
| v1 | `prefers-reduced-motion` was a per-component concern (Skills, Channels, About, Chat each redeclared it; see MOAT.md §2.3). | Cursor respects the global `@media (prefers-reduced-motion)` contract via the single mount-time `matchMedia` check. |

No existing functionality was demoted or removed. Additions only — every row either fixes an existing gap (hover-region, dark variant, reduced-motion, hot spot, disabled affordance) or tightens the energy budget (rAF pause/resume).
