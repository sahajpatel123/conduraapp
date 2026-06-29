<script lang="ts">
  import type { Snippet } from 'svelte'

  type Variant = 'secondary' | 'ghost' | 'accent-ghost'
  type Size = 'sm' | 'md' | 'lg'

  interface Props {
    variant?: Variant
    size?: Size
    disabled?: boolean
    active?: boolean
    onclick?: (e: MouseEvent) => void
    ariaLabel: string
    title?: string
    children: Snippet
  }

  let { variant = 'ghost', size = 'md', disabled = false, active = false,
        onclick, ariaLabel, title, children }: Props = $props()
</script>

<button
  type="button"
  {title}
  aria-label={ariaLabel}
  aria-pressed={active || undefined}
  {disabled}
  class="icon-btn icon-btn-{variant} icon-btn-{size}"
  class:is-active={active}
  {onclick}
>
  {@render children()}
</button>

<style>
  .icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 1px solid transparent;
    background: transparent;
    border-radius: var(--radius-md);
    color: var(--text-muted);
    cursor: pointer;
    transition:
      background-color var(--transition-fast) ease,
      border-color var(--transition-fast) ease,
      color var(--transition-fast) ease,
      transform var(--transition-fast) var(--ease-spring);
  }
  .icon-btn:disabled { opacity: 0.4; cursor: not-allowed; }
  .icon-btn:active:not(:disabled) { transform: scale(0.92); }

  .icon-btn-sm { width: 26px; height: 26px; }
  .icon-btn-md { width: 34px; height: 34px; }
  .icon-btn-lg { width: 42px; height: 42px; }

  .icon-btn-ghost:hover:not(:disabled) {
    background: var(--surface-2);
    color: var(--text);
  }
  .icon-btn-secondary {
    background: var(--surface-2);
    border-color: var(--border-strong);
  }
  .icon-btn-secondary:hover:not(:disabled) {
    background: var(--surface-3);
    border-color: var(--border-focus);
    color: var(--text);
  }
  .icon-btn-accent-ghost:hover:not(:disabled) {
    background: var(--accent-soft);
    color: var(--accent);
  }
  .icon-btn.is-active {
    background: var(--accent-soft);
    color: var(--accent);
  }
</style>