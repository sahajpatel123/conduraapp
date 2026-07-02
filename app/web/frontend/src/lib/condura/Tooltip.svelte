<script lang="ts">
  import type { Snippet } from 'svelte';

  // Condura Tooltip — primitive. Hover/focus after a delay; anchored to
  // its children via a wrapping span. Replaces native title="..." across
  // the shell (MOAT §2.9). Renders a label, an optional keyboard chord
  // (<kbd>), and a small SVG arrow that points at the anchor.
  //
  // Behavior:
  //   - Show after `delay` ms on mouseenter or focus (timer is cancelled
  //     on leave/blur so a late show never fires after a leave).
  //   - Hide immediately on mouseleave or blur.
  //   - prefers-reduced-motion skips the slide (the global block in
  //     condura.css flattens the transition; the `.reduce` class also
  //     zeros the offset transform as belt-and-suspenders).
  //
  // Accessibility:
  //   - The wrapper is role-less; the tooltip has role="tooltip".
  //   - aria-describedby on the wrapper is set only while visible, and
  //     clears when hidden so the description isn't announced forever.
  //   - aria-hidden on the tooltip mirrors visibility for AT.
  //
  // Style (all tokens from condura.css — never raw hex):
  //   background   --paper-2      (surface-card)
  //   text         --ink
  //   border       --hair-strong
  //   radius       --r-sm
  //   elevation    --shadow-card
  //   font         --text-caption (12px mono eyebrow / metadata scale)
  //   <kbd> chip   recessed chip: --surface-sunken + --hair-strong
  //   motion       fade + 4px slide, --dur-fast, --ease
  //   z-index      --z-tooltip (above sticky, below modal)

  type Placement = 'top' | 'bottom' | 'left' | 'right';

  let {
    label,
    chord,
    delay = 400,
    placement = 'top',
    children,
    class: cls = '',
  }: {
    label: string;
    chord?: string;
    delay?: number;
    placement?: Placement;
    children?: Snippet;
    class?: string;
  } = $props();

  let visible = $state(false);
  let reduceMotion = $state(false);
  let showTimer: ReturnType<typeof setTimeout> | null = null;

  // Stable per-instance id, wires the tooltip to its wrapper via aria-describedby.
  // Random suffix is fine here: the Wails GUI is CSR-only (no SSR hydration).
  const tipId = `condura-tip-${Math.random().toString(36).slice(2, 9)}`;

  $effect(() => {
    if (typeof window === 'undefined') return;
    reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  });

  function cancelTimer() {
    if (showTimer != null) {
      clearTimeout(showTimer);
      showTimer = null;
    }
  }

  function show() {
    cancelTimer();
    visible = true;
  }

  function hide() {
    cancelTimer();
    visible = false;
  }

  function scheduleShow() {
    cancelTimer();
    if (delay <= 0) {
      show();
      return;
    }
    showTimer = setTimeout(() => {
      showTimer = null;
      visible = true;
    }, delay);
  }

  function onMouseEnter() {
    scheduleShow();
  }
  function onMouseLeave() {
    hide();
  }
  function onFocus() {
    scheduleShow();
  }
  function onBlur() {
    hide();
  }
</script>

<!-- The anchor is role="presentation" — it's a transparent positional
     wrapper, not an interactive element. Real focus and activation
     live on whatever interactive child the consumer renders inside
     (a <button>, an <a>, an IconButton, etc.). aria-describedby is
     still read by AT when focus lands on that descendant, because
     aria-describedby is computed up the ancestor chain at lookup
     time on most modern screen readers. -->
<span
  class="tooltip-anchor {cls}"
  role="presentation"
  onmouseenter={onMouseEnter}
  onmouseleave={onMouseLeave}
  onfocus={onFocus}
  onblur={onBlur}
  aria-describedby={visible ? tipId : undefined}
>
  {@render children?.()}
  <span
    class="tooltip tooltip-{placement}"
    class:shown={visible}
    class:reduce={reduceMotion}
    role="tooltip"
    id={tipId}
    aria-hidden={!visible}
  >
    <span class="tooltip-label">{label}</span>
    {#if chord}
      <kbd class="tooltip-kbd">{chord}</kbd>
    {/if}
    <svg class="tooltip-arrow" viewBox="0 0 8 8" aria-hidden="true">
      <path d="M0 8 L4 0 L8 8 Z" />
    </svg>
  </span>
</span>

<style>
  /* The anchor wraps the children and is the hover/focus target.
     `position: relative` anchors absolutely-positioned tooltip children;
     `display: inline-block` sizes the wrapper to its content (a button
     or <Glyph> sized 20×20 gets a tight wrapping box). */
  .tooltip-anchor {
    position: relative;
    display: inline-block;
  }

  /* The floating surface. The base class owns the visual language;
     the per-placement classes own position + initial slide offset. */
  .tooltip {
    position: absolute;
    z-index: var(--z-tooltip);
    pointer-events: none;
    white-space: nowrap;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 6px 10px;
    background: var(--paper-2);
    color: var(--ink);
    font-family: var(--font-mono);
    font-size: var(--text-caption);
    line-height: var(--lh-caption);
    letter-spacing: var(--ls-caption);
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-sm);
    box-shadow: var(--shadow-card);
    opacity: 0;
    transition:
      opacity   var(--dur-fast) var(--ease),
      transform var(--dur-fast) var(--ease);
    will-change: opacity, transform;
  }

  /* Initial offsets: each placement starts slightly offset toward the
     anchor (down for top, up for bottom, right for left, left for right)
     and animates into its resting position when `.shown` is added. */
  .tooltip-top {
    bottom: calc(100% + 8px);
    left: 50%;
    transform: translateX(-50%) translateY(4px);
  }
  .tooltip-bottom {
    top: calc(100% + 8px);
    left: 50%;
    transform: translateX(-50%) translateY(-4px);
  }
  .tooltip-left {
    right: calc(100% + 8px);
    top: 50%;
    transform: translateY(-50%) translateX(4px);
  }
  .tooltip-right {
    left: calc(100% + 8px);
    top: 50%;
    transform: translateY(-50%) translateX(-4px);
  }

  /* Final state: faded in and centered on the anchor. */
  .tooltip.shown {
    opacity: 1;
  }
  .tooltip-top.shown {
    transform: translateX(-50%) translateY(0);
  }
  .tooltip-bottom.shown {
    transform: translateX(-50%) translateY(0);
  }
  .tooltip-left.shown {
    transform: translateY(-50%) translateX(0);
  }
  .tooltip-right.shown {
    transform: translateY(-50%) translateX(0);
  }

  /* Reduced-motion: zero the offset transform so the tooltip sits at its
     final position from the very first frame. The global prefers-reduced-
     motion block in condura.css also flattens transitions to 0.01ms; this
     rule keeps the resting position correct even if the global rule is
     bypassed (e.g., during animation testing). */
  .tooltip-top.reduce {
    transform: translateX(-50%) translateY(0);
  }
  .tooltip-bottom.reduce {
    transform: translateX(-50%) translateY(0);
  }
  .tooltip-left.reduce {
    transform: translateY(-50%) translateX(0);
  }
  .tooltip-right.reduce {
    transform: translateY(-50%) translateX(0);
  }

  /* Label + chord. The label is plain tooltip text; the chord is a
     recessed chip in the kbd style from DESIGNLANG.md §1. */
  .tooltip-label {
    display: inline-block;
  }
  .tooltip-kbd {
    display: inline-block;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--content-soft);
    background: var(--surface-sunken);
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-xs);
    padding: 1px 5px;
    margin-left: 2px;
  }

  /* The arrow — a triangle in its base orientation (apex up). Filled
     with the tooltip background so the body of the arrow blends with
     the tooltip body; stroked with the tooltip border so the seam
     matches the frame. */
  .tooltip-arrow {
    position: absolute;
    width: 8px;
    height: 8px;
    fill: var(--paper-2);
    stroke: var(--hair-strong);
    stroke-width: 1;
    stroke-linejoin: round;
  }
  /* Top placement: arrow sits at bottom-center of the tooltip,
     rotated 180° to point down at the trigger below. */
  .tooltip-top .tooltip-arrow {
    bottom: -4px;
    left: 50%;
    transform: translateX(-50%) rotate(180deg);
  }
  /* Bottom placement: arrow sits at top-center, no rotation — apex
     points up at the trigger above. */
  .tooltip-bottom .tooltip-arrow {
    top: -4px;
    left: 50%;
    transform: translateX(-50%) rotate(0deg);
  }
  /* Left placement: arrow sits at the right edge, rotated 90° so
     the apex points right at the trigger. */
  .tooltip-left .tooltip-arrow {
    right: -4px;
    top: 50%;
    transform: translateY(-50%) rotate(90deg);
  }
  /* Right placement: arrow sits at the left edge, rotated -90° so
     the apex points left at the trigger. */
  .tooltip-right .tooltip-arrow {
    left: -4px;
    top: 50%;
    transform: translateY(-50%) rotate(-90deg);
  }
</style>
