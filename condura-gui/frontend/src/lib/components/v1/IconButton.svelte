<!--
  IconButton — square button containing only an icon.

  Used for toolbar actions, popover triggers, and inline controls where
  text isn't needed. Always has a visible focus ring and an aria-label.

  Props:
    name      — IconName (from the icons module)
    label     — required aria-label
    size      — button size in px (32/36/40)
    variant   — 'ghost' | 'secondary' | 'primary'
    disabled  — disables the button
    onclick   — click handler
-->
<script lang="ts">
  import Icon, { type IconName } from './icons/Icon.svelte';

  interface Props {
    name: IconName;
    label: string;
    size?: number;
    variant?: 'ghost' | 'secondary' | 'primary';
    disabled?: boolean;
    onclick?: (e: MouseEvent) => void;
  }

  let {
    name,
    label,
    size = 32,
    variant = 'ghost',
    disabled = false,
    onclick,
  }: Props = $props();

  // Icon size inside the button (leave 8px padding for sm, 10px for md/lg)
  let iconPx = $derived(size <= 32 ? 16 : size <= 36 ? 18 : 20);
</script>

<button
  class="icon-btn icon-btn--{variant}"
  class:icon-btn--disabled={disabled}
  {disabled}
  type="button"
  aria-label={label}
  title={label}
  style="--icon-btn-size: {size}px; --icon-px: {iconPx}px;"
  onclick={onclick}
>
  <Icon name={name} size="md" />
</button>

<style>
  .icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: var(--icon-btn-size);
    height: var(--icon-btn-size);
    padding: 0;
    background-color: transparent;
    border: 1px solid transparent;
    border-radius: var(--radius-md);
    color: var(--content-secondary);
    cursor: pointer;
    flex-shrink: 0;
    transition:
      background-color var(--duration-fast) var(--ease-standard),
      border-color var(--duration-fast) var(--ease-standard),
      color var(--duration-fast) var(--ease-standard),
      transform var(--duration-fast) var(--ease-standard);
  }

  .icon-btn :global(.icon) {
    width: var(--icon-px);
    height: var(--icon-px);
  }

  /* Ghost (default) — neutral until hover */
  .icon-btn--ghost:hover:not(.icon-btn--disabled) {
    background-color: var(--paper-warm-50);
    color: var(--content-primary);
  }

  /* Secondary — outlined, looks like a control */
  .icon-btn--secondary {
    border-color: var(--border-default);
    background-color: var(--surface-raised);
  }
  .icon-btn--secondary:hover:not(.icon-btn--disabled) {
    border-color: var(--border-strong);
    background-color: var(--paper-warm-50);
    color: var(--content-primary);
  }

  /* Primary — filled with plum */
  .icon-btn--primary {
    background-color: var(--action-primary-idle-bg);
    border-color: var(--action-primary-idle-bg);
    color: var(--action-primary-idle-fg);
  }
  .icon-btn--primary:hover:not(.icon-btn--disabled) {
    background-color: var(--action-primary-hover-bg);
    border-color: var(--action-primary-hover-bg);
  }

  .icon-btn:active:not(.icon-btn--disabled) {
    transform: translateY(1px);
  }

  .icon-btn:focus-visible {
    outline: var(--border-focus) solid 2px;
    outline-offset: 2px;
  }

  .icon-btn--disabled {
    color: var(--content-disabled);
    cursor: not-allowed;
  }
</style>