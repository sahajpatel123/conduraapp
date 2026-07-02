<script lang="ts">
  /**
   * ThemePicker.svelte — Condura theme picker (Phase 4)
   *
   * ──────────────────────────────────────────────────────────────────────
   * API contract
   * ──────────────────────────────────────────────────────────────────────
   * Props:   none (self-contained; reads its initial value from localStorage
   *          on mount and falls back to 'light' if absent).
   * Events:  dispatches `CustomEvent('condura-theme-change', {
   *            detail: { theme }
   *          })` on `window` (bubbles: true).  Ancestors (Shell, Settings)
   *          listen and re-apply as needed.
   * Storage: writes `localStorage.setItem('condura-theme', value)` where
   *          value is one of `'light' | 'dark' | 'system'`.  Initial value
   *          on mount is `localStorage.getItem('condura-theme') ?? 'light'`.
   *
   * ARIA:    `role="radiogroup"` on the container, `role="radio"` on each
   *          button, `aria-checked` reflecting selection.  Roving tabindex;
   *          focus follows the active option.
   * Keyboard: ←/↑ previous, →/↓ next, Home first, End last, Enter/Space
   *           select.  Arrow navigation moves focus but does NOT commit a
   *           theme change — only Enter/Space/click commits.
   *
   * ──────────────────────────────────────────────────────────────────────
   * Palette-switch choreography (the $50M moment — Approach C)
   * ──────────────────────────────────────────────────────────────────────
   * When the user clicks an option, a circle of the *destination* palette
   * grows from the click point until it covers the viewport; the data-mode
   * attribute then commits underneath and the overlay is removed.  The
   * transition reads as "the new theme is being painted in from where you
   * reached for it."
   *
   *   1. Capture click origin (clientX, clientY).
   *   2. Write `--ox` / `--oy` custom properties on :root so CSS knows
   *      where to grow from.
   *   3. If `document.startViewTransition` is available (Chrome 111+, Edge,
   *      Safari 18+), use it and override ::view-transition-new(root) with
   *      a custom clip-path morph.  This is the cleanest path — the
   *      browser handles the snapshot diff and the geometry math.
   *   4. Otherwise, fall back to a manual overlay: a full-viewport element
   *      with the destination paper as backdrop, transitioning its
   *      clip-path from `circle(0 at ox oy)` to `circle(150% at ox oy)`,
   *      commit the data-mode change at the transition's peak, then
   *      unmount the overlay.
   *   5. `prefers-reduced-motion: reduce` → instant switch, no animation.
   *
   * The selected option's icon flips instantly so the user always sees the
   * pressed state register; only the *palette* animates.
   *
   * ──────────────────────────────────────────────────────────────────────
   * Mature Rules enforced (DESIGNLANG.md + MOAT.md)
   * ──────────────────────────────────────────────────────────────────────
   *   • Focus ring tracks the rounded pill (pollen halo, no rectangular
   *     outline).  `box-shadow` rather than `outline` so the halo curves.
   *   • Press state scales 0.97 via the global `.tactile` rule (we don't
   *     declare it locally — one source for the gesture).
   *   • All colors via condura.css tokens; the two destination paper hex
   *     values (#F4EFE4 / #16140F) are the ONLY raw hex in the file, used
   *     solely for the reveal overlay's destination backdrop.  They mirror
   *     condura.css :root[data-mode] exactly — required because the
   *     :root[data-mode] token selectors don't cascade through a foreign
   *     data-mode attribute on a descendant element.
   *   • 1.25px thread-weight icons via <Glyph stroke={1.25}/> (matches
   *     the "thread weight" pattern from condura.css §C / DESIGNLANG.md).
   *   • prefers-reduced-motion respected across both approaches.
   *   • Sun / auto / moon are inline SVG paths from ./icons.ts — not
   *     lucide, not unicode glyphs.
   * ──────────────────────────────────────────────────────────────────────
   */
  import { onMount } from 'svelte';
  import Glyph from './Glyph.svelte';

  type Theme = 'light' | 'dark' | 'system';
  type EffectiveMode = 'light' | 'dark';

  const STORAGE_KEY = 'condura-theme';

  const OPTIONS: ReadonlyArray<{
    value: Theme;
    label: string;
    glyph: 'sun' | 'auto' | 'moon';
  }> = [
    { value: 'light',  label: 'Light', glyph: 'sun'  },
    { value: 'system', label: 'Auto',  glyph: 'auto' },
    { value: 'dark',   label: 'Dark',  glyph: 'moon' },
  ];

  // Destination palette papers, mirrored from condura.css :root[data-mode=*].
  // These two values are the ONLY raw hex in the file.  They are required
  // for the reveal overlay because :root[data-mode] tokens don't cascade
  // through a foreign data-mode attribute on a descendant element.
  const DESTINATION_PAPER: Record<EffectiveMode, string> = {
    light: '#F4EFE4',
    dark:  '#16140F',
  };

  let current = $state<Theme>('light');
  let activeIdx = $state(0);
  let buttonEls = $state<(HTMLButtonElement | null)[]>([]);

  // Manual reveal overlay state — only mounted during the animation.
  let revealEl: HTMLDivElement | null = $state(null);
  let revealState = $state<{ active: boolean; mode: EffectiveMode }>({
    active: false,
    mode: 'light',
  });

  function getEffective(t: Theme): EffectiveMode {
    if (t === 'system') {
      return matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }
    return t;
  }

  function readStored(): Theme {
    try {
      const v = localStorage.getItem(STORAGE_KEY);
      if (v === 'light' || v === 'dark' || v === 'system') return v;
    } catch {
      /* ignore (private mode, etc.) */
    }
    return 'light';
  }

  function persist(t: Theme): void {
    try {
      localStorage.setItem(STORAGE_KEY, t);
    } catch {
      /* ignore */
    }
  }

  function announce(): void {
    try {
      window.dispatchEvent(
        new CustomEvent('condura-theme-change', {
          detail: { theme: current },
          bubbles: true,
        }),
      );
    } catch {
      /* ignore */
    }
  }

  function applyMode(eff: EffectiveMode): void {
    document.documentElement.setAttribute('data-mode', eff);
  }

  function setOrigin(x: number, y: number): void {
    const root = document.documentElement;
    root.style.setProperty('--ox', `${x}px`);
    root.style.setProperty('--oy', `${y}px`);
  }

  function setTheme(
    t: Theme,
    origin: { x: number; y: number },
    animate: boolean,
  ): void {
    if (t === current) return;
    current = t;
    activeIdx = OPTIONS.findIndex((o) => o.value === t);
    persist(t);

    const eff = getEffective(t);
    const reduced = matchMedia('(prefers-reduced-motion: reduce)').matches;

    setOrigin(origin.x, origin.y);

    if (!animate || reduced) {
      // Instant switch — no animation.  Honors prefers-reduced-motion.
      applyMode(eff);
      announce();
      return;
    }

    // Approach A — View Transitions API, custom-morphed into Approach C's
    // clip-path circle reveal (see ::view-transition-new(root) below).
    type VTDocument = Document & {
      startViewTransition?: (cb: () => void) => unknown;
    };
    const startVT = (document as VTDocument).startViewTransition;
    if (typeof startVT === 'function') {
      startVT.call(document, () => {
        applyMode(eff);
      });
      announce();
      return;
    }

    // Approach C — manual clip-path reveal (fallback for browsers without
    // View Transitions, or when JS animation is the safer choice).
    beginManualReveal(eff);
    announce();
  }

  function beginManualReveal(eff: EffectiveMode): void {
    revealState = { active: true, mode: eff };
    // Two RAFs so the initial state (clip-path: circle(0)) paints before
    // we add the .expanding class.  Without this, the browser would never
    // see the starting geometry and the transition would skip to the end.
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        revealEl?.classList.add('expanding');
        let committed = false;
        const commit = (): void => {
          if (committed) return;
          committed = true;
          applyMode(eff);
          // Brief pause so the eye registers the fully-painted new
          // palette before the overlay is removed.
          setTimeout(() => {
            revealEl?.classList.remove('expanding');
            revealState = { active: false, mode: 'light' };
          }, 80);
        };
        revealEl?.addEventListener('transitionend', commit, { once: true });
        // Safety net: if transitionend never fires (e.g., element was
        // detached), commit anyway after the duration + a small buffer.
        setTimeout(commit, 640);
      });
    });
  }

  function clickOrigin(ev: MouseEvent): { x: number; y: number } {
    return { x: ev.clientX, y: ev.clientY };
  }

  function keyboardOrigin(idx: number): { x: number; y: number } {
    const btn = buttonEls[idx];
    if (btn) {
      const r = btn.getBoundingClientRect();
      return { x: r.left + r.width / 2, y: r.top + r.height / 2 };
    }
    return { x: window.innerWidth / 2, y: window.innerHeight / 2 };
  }

  function onClick(t: Theme, ev: MouseEvent): void {
    setTheme(t, clickOrigin(ev), true);
  }

  function onKey(ev: KeyboardEvent, idx: number): void {
    switch (ev.key) {
      case 'ArrowLeft':
      case 'ArrowUp': {
        ev.preventDefault();
        const next = (idx - 1 + OPTIONS.length) % OPTIONS.length;
        buttonEls[next]?.focus();
        return;
      }
      case 'ArrowRight':
      case 'ArrowDown': {
        ev.preventDefault();
        const next = (idx + 1) % OPTIONS.length;
        buttonEls[next]?.focus();
        return;
      }
      case 'Home': {
        ev.preventDefault();
        buttonEls[0]?.focus();
        return;
      }
      case 'End': {
        ev.preventDefault();
        buttonEls[OPTIONS.length - 1]?.focus();
        return;
      }
      case 'Enter':
      case ' ': {
        ev.preventDefault();
        setTheme(OPTIONS[idx].value, keyboardOrigin(idx), true);
        return;
      }
      default:
        return;
    }
  }

  onMount(() => {
    current = readStored();
    activeIdx = OPTIONS.findIndex((o) => o.value === current);
  });
</script>

<div class="picker" role="radiogroup" aria-label="Theme">
  {#each OPTIONS as opt, i (opt.value)}
    <button
      bind:this={buttonEls[i]}
      type="button"
      role="radio"
      class="seg"
      class:active={current === opt.value}
      aria-checked={current === opt.value}
      aria-label={opt.label}
      tabindex={i === activeIdx ? 0 : -1}
      onclick={(e) => onClick(opt.value, e)}
      onkeydown={(e) => onKey(e, i)}
    >
      <span class="glyph" aria-hidden="true">
        <Glyph name={opt.glyph} size={16} stroke={1.25} />
      </span>
      <span class="label">{opt.label}</span>
    </button>
  {/each}
</div>

{#if revealState.active}
  <div
    bind:this={revealEl}
    class="reveal"
    style="background: {DESTINATION_PAPER[revealState.mode]};"
    aria-hidden="true"
  ></div>
{/if}

<style>
  /* ── the segmented control ─────────────────────────────────────── */

  .picker {
    display: inline-flex;
    gap: 2px;
    padding: 3px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-pill);
    background: var(--surface-card);
    box-shadow: inset 0 1px 0 color-mix(in oklab, var(--synapse) 4%, transparent);
  }

  .seg {
    position: relative;
    display: inline-flex;
    align-items: center;
    gap: 7px;
    padding: 7px 14px;
    font-family: var(--font-sans);
    font-size: 13px;
    font-weight: 500;
    line-height: 1;
    letter-spacing: var(--ls-body);
    color: var(--content-mute);
    background: transparent;
    border: 1px solid transparent;
    border-radius: var(--r-pill);
    cursor: pointer;
    transition:
      color           var(--dur) var(--ease),
      background-color var(--dur) var(--ease),
      border-color    var(--dur) var(--ease),
      box-shadow      var(--dur) var(--ease);
  }

  /* Hover (inactive only) — hair-strong border to suggest interactivity. */
  .seg:hover:not(.active) {
    color: var(--content);
    border-color: var(--hair-strong);
  }

  /* Focus ring — pollen halo, tracks the pill's rounded geometry.
     We override the global :focus-visible here because the picker sits
     on a hairline-stroked surface where the synapse-glow focus would
     compete with the segmented-control's frame. */
  .seg:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo-color);
  }

  /* Active — pollen fill, paper text.  Press scale comes from the
     global `.tactile` rule in condura.css; we don't redeclare it. */
  .seg.active {
    background: var(--pollen);
    color: var(--paper);
    border-color: color-mix(in oklab, var(--pollen-deep) 50%, transparent);
    box-shadow: var(--shadow-paper);
  }

  /* Dark mode lightens the pollen to a more readable yellow-orange —
     dark text reads better on it than cream text would. */
  :global([data-mode='dark']) .seg.active {
    color: var(--ink);
  }

  .glyph {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    color: currentColor;
  }

  .label {
    font-size: 13px;
  }

  /* ── Approach C — manual clip-path reveal ────────────────────────
     A full-viewport overlay whose clip-path grows from circle(0) at the
     click origin to circle(150%).  At the peak, the underlying data-mode
     commits and the overlay is removed (no visual change to the user). */

  .reveal {
    position: fixed;
    inset: 0;
    z-index: var(--z-max);
    pointer-events: none;
    clip-path: circle(0 at var(--ox, 50%) var(--oy, 50%));
    transition: clip-path 560ms var(--ease);
    will-change: clip-path;
  }

  /* `.expanding` is toggled via JS (classList.add) after a double-RAF so
     the initial geometry paints first.  `:global()` opts out of Svelte's
     unused-selector check — the class is set at runtime, not via `class:`. */
  :global(.reveal.expanding) {
    clip-path: circle(150% at var(--ox, 50%) var(--oy, 50%));
  }

  /* ── Approach A — View Transitions API, custom-morphed ───────────
     When startViewTransition is available, the browser snapshots the
     old and new :root states and morphs between them.  We override
     the default cross-fade with a clip-path circle that grows from
     the click origin (old stays solid underneath; new expands on top). */

  :global(::view-transition-old(root)),
  :global(::view-transition-new(root)) {
    animation: none;
    mix-blend-mode: normal;
  }
  :global(::view-transition-new(root)) {
    clip-path: circle(0 at var(--ox, 50%) var(--oy, 50%));
    animation: theme-reveal 560ms var(--ease) forwards;
  }

  @keyframes theme-reveal {
    to {
      clip-path: circle(150% at var(--ox, 50%) var(--oy, 50%));
    }
  }

  /* ── prefers-reduced-motion ───────────────────────────────────── */

  @media (prefers-reduced-motion: reduce) {
    .reveal {
      transition: none;
      clip-path: none;
    }
    :global(::view-transition-old(root)),
    :global(::view-transition-new(root)) {
      animation: none !important;
    }
  }
</style>