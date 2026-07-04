<script lang="ts">
  import type { Snippet } from 'svelte';
  import { magnetic } from './magnetic';

  // Condura Button — primary is pollen + magnetic + halo; secondary/ghost
  // use hairlines; danger is the destructive accent. Magnetic pull only on
  // primary to avoid fighting the hover-lift transform on the others.
  let {
    variant = 'primary',
    size = 'md',
    magnetic: mag = false,
    disabled = false,
    type = 'button',
    onclick,
    children,
    class: cls = '',
  }: {
    variant?: 'primary' | 'secondary' | 'ghost' | 'danger';
    size?: 'md' | 'sm';
    magnetic?: boolean;
    disabled?: boolean;
    type?: 'button' | 'submit';
    onclick?: (e: MouseEvent) => void;
    children?: Snippet;
    class?: string;
  } = $props();
</script>

<button
  {type}
  class="btn btn-{variant} {cls}"
  class:sm={size === 'sm'}
  {disabled}
  {onclick}
  use:magnetic={{ enabled: mag, strength: 0.35 }}
>
  {@render children?.()}
</button>

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-family: var(--font-sans);
    font-size: 14px;
    font-weight: 500;
    letter-spacing: -0.005em;
    border-radius: var(--r-pill);
    padding: 11px 20px;
    border: 1px solid transparent;
    transition:
      transform var(--dur) var(--ease),
      background var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease),
      color var(--dur) var(--ease),
      border-color var(--dur) var(--ease);
    will-change: transform;
  }
  .btn.sm {
    padding: 7px 14px;
    font-size: 13px;
  }
  .btn[disabled] {
    opacity: 0.42;
    pointer-events: none;
    filter: saturate(0.55);
    cursor: not-allowed;
    color: var(--content-mute);
  }

  /* Tactile focus: pollen halo + hairline synapse inset. The base :focus-visible
     rule in condura.css is broad; we sharpen it here for buttons so the halo
     tracks the rounded pill rather than rendering as a rectangular glow. */
  .btn:focus-visible {
    outline: none;
    box-shadow:
      0 0 0 4px var(--pollen-halo),
      inset 0 0 0 1px color-mix(in oklab, var(--synapse) 30%, transparent);
  }

  .btn-primary {
    background: var(--pollen);
    color: var(--paper);
    box-shadow: var(--shadow-paper);
  }
  :global([data-mode='dark']) .btn-primary {
    color: var(--ink);
  }
  .btn-primary:hover:not([disabled]) {
    box-shadow:
      0 1px 0 color-mix(in oklab, var(--paper) 12%, transparent) inset,
      0 18px 40px -16px color-mix(in oklab, var(--ink) 60%, transparent),
      var(--pollen-halo);
    transform: translateY(-1px);
  }
  .btn-primary:active:not([disabled]) {
    transform: scale(0.97);
  }

  .btn-secondary {
    background: var(--surface-card);
    color: var(--content);
    border-color: var(--hair);
  }
  .btn-secondary:hover:not([disabled]) {
    border-color: var(--hair-strong);
    background: var(--paper-2);
    transform: translateY(-1px);
  }
  .btn-secondary:active:not([disabled]) {
    transform: scale(0.97);
  }

  .btn-ghost {
    background: transparent;
    color: var(--content);
    border-color: var(--hair-strong);
  }
  .btn-ghost:hover:not([disabled]) {
    border-color: var(--synapse);
    color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 6%, transparent);
    transform: translateY(-1px);
  }
  .btn-ghost:active:not([disabled]) {
    transform: scale(0.97);
  }

  .btn-danger {
    background: var(--danger);
    color: var(--paper);
  }
  .btn-danger:hover:not([disabled]) {
    box-shadow:
      0 0 0 4px color-mix(in oklab, var(--danger) 20%, transparent),
      0 18px 40px -16px color-mix(in oklab, var(--ink) 60%, transparent);
    transform: translateY(-1px);
  }
  .btn-danger:active:not([disabled]) {
    transform: scale(0.97);
  }
</style>