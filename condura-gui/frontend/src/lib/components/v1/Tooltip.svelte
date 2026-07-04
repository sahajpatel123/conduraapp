<!--
  Tooltip — small contextual help on hover.

  The native `title` attribute is ugly and slow. A proper tooltip:
    - appears after a 300ms delay (prevents flicker on mouse-over)
    - has refined positioning (top/bottom/left/right of trigger)
    - uses serif for the label, mono for the kbd shortcut
    - has a subtle entrance animation
    - is keyboard-accessible (focus + Esc)

  Usage:
    <Tooltip label="Save" kbd="⌘S">
      <button>...</button>
    </Tooltip>

  Props:
    label     — the tooltip text
    kbd       — optional keyboard shortcut to display
    position  — 'top' | 'bottom' | 'left' | 'right' (default: 'top')
    delay     — ms before showing (default: 300)
    children  — the trigger element(s)
-->
<script lang="ts">
  import KeyCombo from './KeyCombo.svelte';
  import type { Snippet } from 'svelte';

  type Position = 'top' | 'bottom' | 'left' | 'right';

  interface Props {
    label: string;
    kbd?: string;
    position?: Position;
    delay?: number;
    children: Snippet;
  }

  let { label, kbd, position = 'top', delay = 300, children }: Props = $props();

  let visible = $state(false);
  let triggerEl: HTMLElement | null = $state(null);
  let timer: ReturnType<typeof setTimeout> | null = null;

  function show() {
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => {
      visible = true;
    }, delay);
  }

  function hide() {
    if (timer) {
      clearTimeout(timer);
      timer = null;
    }
    visible = false;
  }

  function handleKey(e: KeyboardEvent) {
    if (e.key === 'Escape') hide();
  }
</script>

<svelte:window onkeydown={handleKey} />

<!-- svelte-ignore a11y_no_static_element_interactions -->
<span
  class="tooltip-host"
  bind:this={triggerEl}
  onmouseenter={show}
  onmouseleave={hide}
  onfocusin={show}
  onfocusout={hide}
>
  {@render children()}

  {#if visible}
    <span
      class="tooltip tooltip--{position}"
      role="tooltip"
      aria-hidden="false"
    >
      <span class="tooltip__label">{label}</span>
      {#if kbd}
        <span class="tooltip__kbd">
          <KeyCombo combo={kbd} size="sm" />
        </span>
      {/if}
    </span>
  {/if}
</span>

<style>
  .tooltip-host {
    position: relative;
    display: inline-flex;
  }

  .tooltip {
    position: absolute;
    z-index: var(--z-tooltip);
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    background-color: var(--ink-cool-900);
    color: var(--paper-warm-0);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-2);
    font-family: var(--font-serif);
    font-size: var(--text-body-sm-size);
    line-height: 1.2;
    white-space: nowrap;
    pointer-events: none;
    /* The entrance — quick fade + small scale-up */
    animation: tooltip-in var(--duration-fast) var(--ease-decelerate) both;
  }

  @keyframes tooltip-in {
    from {
      opacity: 0;
      transform: var(--tooltip-from, translate(-50%, -4px)) scale(0.96);
    }
    to {
      opacity: 1;
      transform: var(--tooltip-to, translate(-50%, 0)) scale(1);
    }
  }

  /* Position variants */
  .tooltip--top {
    bottom: calc(100% + var(--space-2));
    left: 50%;
    --tooltip-to: translate(-50%, 0);
    --tooltip-from: translate(-50%, 4px);
  }

  .tooltip--bottom {
    top: calc(100% + var(--space-2));
    left: 50%;
    --tooltip-to: translate(-50%, 0);
    --tooltip-from: translate(-50%, -4px);
  }

  .tooltip--left {
    right: calc(100% + var(--space-2));
    top: 50%;
    --tooltip-to: translate(0, -50%);
    --tooltip-from: translate(4px, -50%);
  }

  .tooltip--right {
    left: calc(100% + var(--space-2));
    top: 50%;
    --tooltip-to: translate(0, -50%);
    --tooltip-from: translate(-4px, -50%);
  }

  .tooltip__label {
    /* Serif voice — feels like a thought, not a tooltip */
  }

  .tooltip__kbd {
    display: inline-flex;
    align-items: center;
  }

  .tooltip__kbd :global(kbd) {
    background-color: rgba(255, 255, 255, 0.12);
    border-color: rgba(255, 255, 255, 0.2);
    color: var(--paper-warm-0);
  }

  @media (prefers-reduced-motion: reduce) {
    .tooltip {
      animation: none;
    }
  }
</style>