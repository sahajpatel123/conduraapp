<!--
  Button — Condura v2 primary action primitive.

  Three variants: primary (the warm-earth accent), ghost (no fill,
  just text + hairline on hover), and deny (the signal-stop tone,
  used only for destructive consent). Two sizes: default and small.
  Hardware-honest: a real-feeling press, not a flat rectangle.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLButtonAttributes } from 'svelte/elements'

  let {
    variant = 'primary' as 'primary' | 'ghost' | 'deny',
    size = 'default' as 'default' | 'small',
    disabled = false,
    type = 'button' as 'button' | 'submit' | 'reset',
    icon = undefined as Snippet | undefined,
    children,
    class: klass = '',
    ...rest
  }: HTMLButtonAttributes & {
    variant?: 'primary' | 'ghost' | 'deny'
    size?: 'default' | 'small'
    icon?: Snippet
    children?: Snippet
    class?: string
  } = $props()

  let pressed = $state(false)
</script>

<button
  data-v2-button
  data-variant={variant}
  data-size={size}
  disabled={disabled}
  {type}
  class={klass}
  class:pressed
  onmousedown={() => { pressed = true }}
  onmouseup={() => { pressed = false }}
  onmouseleave={() => { pressed = false }}
  {...rest}
>
  {#if icon}
    <span class="icon" aria-hidden="true">{@render icon()}</span>
  {/if}
  {@render children?.()}
</button>

<style>
  [data-v2-button] {
    display: inline-flex;
    align-items: center;
    gap: var(--v2-space-2);
    font-family: var(--v2-font-sans);
    font-size: var(--v2-text-14);
    line-height: 1;
    font-weight: 500;
    font-feature-settings: var(--v2-numeric-features);
    padding: 0 var(--v2-space-4);
    height: 36px;
    border-radius: var(--v2-radius-1);
    cursor: pointer;
    user-select: none;
    position: relative;
    transition:
      background-color var(--v2-dur-fast) var(--v2-ease-out-soft),
      color            var(--v2-dur-fast) var(--v2-ease-out-soft),
      border-color     var(--v2-dur-fast) var(--v2-ease-out-soft),
      transform        var(--v2-dur-fast) var(--v2-ease-out-soft),
      box-shadow       var(--v2-dur-fast) var(--v2-ease-settle);
  }
  [data-v2-button][data-size='small'] {
    height: 28px;
    padding: 0 var(--v2-space-3);
    font-size: var(--v2-text-12);
  }

  /* Primary — warm earth */
  [data-v2-button][data-variant='primary'] {
    background: var(--v2-accent);
    color: var(--v2-paper);
    border: 1px solid transparent;
    box-shadow: var(--v2-shadow-1);
  }
  [data-v2-button][data-variant='primary']:hover {
    background: color-mix(in srgb, var(--v2-accent) 92%, var(--v2-ink));
  }
  [data-v2-button][data-variant='primary'].pressed,
  [data-v2-button][data-variant='primary']:active {
    transform: translateY(1px);
    box-shadow: none;
  }

  /* Ghost — only shows itself on hover/focus */
  [data-v2-button][data-variant='ghost'] {
    background: transparent;
    color: var(--v2-ink);
    border: 1px solid transparent;
  }
  [data-v2-button][data-variant='ghost']:hover {
    background: color-mix(in srgb, var(--v2-ink) 4%, transparent);
  }
  [data-v2-button][data-variant='ghost'].pressed,
  [data-v2-button][data-variant='ghost']:active {
    transform: translateY(1px);
  }

  /* Deny — used only for destructive consent */
  [data-v2-button][data-variant='deny'] {
    background: transparent;
    color: var(--v2-signal-stop);
    border: 1px solid color-mix(in srgb, var(--v2-signal-stop) 30%, transparent);
  }
  [data-v2-button][data-variant='deny']:hover {
    background: color-mix(in srgb, var(--v2-signal-stop) 6%, transparent);
    border-color: var(--v2-signal-stop);
  }

  /* Focus — visible only via keyboard */
  [data-v2-button]:focus-visible {
    outline: none;
    box-shadow: var(--v2-focus-ring);
  }

  /* Disabled */
  [data-v2-button]:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
</style>
