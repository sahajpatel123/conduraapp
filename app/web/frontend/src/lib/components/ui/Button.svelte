<script lang="ts">
  import type { Snippet } from 'svelte'

  type Variant = 'primary' | 'secondary' | 'ghost' | 'danger' | 'accent-ghost'
  type Size = 'xs' | 'sm' | 'md' | 'lg'

  interface Props {
    variant?: Variant
    size?: Size
    type?: 'button' | 'submit' | 'reset'
    disabled?: boolean
    loading?: boolean
    fullWidth?: boolean
    iconOnly?: boolean
    onclick?: (e: MouseEvent) => void
    children?: Snippet
    title?: string
    ariaLabel?: string
  }

  let {
    variant = 'secondary',
    size = 'md',
    type = 'button',
    disabled = false,
    loading = false,
    fullWidth = false,
    iconOnly = false,
    onclick,
    children,
    title,
    ariaLabel,
  }: Props = $props()
</script>

<button
  {type}
  {title}
  aria-label={ariaLabel}
  aria-busy={loading || undefined}
  disabled={disabled || loading}
  class="btn btn-{variant} btn-{size}"
  class:btn-full={fullWidth}
  class:btn-icon={iconOnly}
  class:is-loading={loading}
  {onclick}
>
  {#if loading}
    <span class="dot-loader" aria-hidden="true">
      <span></span><span></span><span></span>
    </span>
  {/if}
  {#if children}
    <span class="btn-label" class:btn-label-loading={loading}>{@render children()}</span>
  {/if}
</button>

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    border: 1px solid transparent;
    border-radius: var(--radius-md);
    font-family: var(--font-sans);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-normal);
    white-space: nowrap;
    user-select: none;
    -webkit-user-select: none;
    position: relative;
    cursor: pointer;
    transition:
      background-color var(--transition-fast) ease,
      border-color var(--transition-fast) ease,
      color var(--transition-fast) ease,
      box-shadow var(--transition-fast) ease,
      transform var(--transition-fast) var(--ease-spring);
    isolation: isolate;
  }

  .btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .btn:active:not(:disabled) {
    transform: scale(0.97);
    transition-duration: var(--transition-instant);
  }

  .btn-xs { padding: 4px 10px;  font-size: var(--size-xs); height: 24px; }
  .btn-sm { padding: 6px 14px;  font-size: var(--size-sm); height: 30px; }
  .btn-md { padding: 8px 18px;  font-size: var(--size-md); height: 36px; }
  .btn-lg { padding: 12px 28px; font-size: var(--size-lg); height: 44px; }

  .btn-icon { padding: 0; aspect-ratio: 1; }
  .btn-icon.btn-xs { width: 24px; }
  .btn-icon.btn-sm { width: 30px; }
  .btn-icon.btn-md { width: 36px; }
  .btn-icon.btn-lg { width: 44px; }

  .btn-full { width: 100%; }

  .btn-primary {
    background: var(--accent-gradient);
    color: var(--text-inverse);
    box-shadow: var(--shadow-inset), 0 1px 2px rgba(0, 0, 0, 0.25);
  }
  .btn-primary:hover:not(:disabled) {
    box-shadow: var(--shadow-inset), 0 0 24px var(--accent-glow);
    transform: translateY(-1px);
  }
  .btn-primary:active:not(:disabled) {
    transform: translateY(0) scale(0.97);
  }

  .btn-secondary {
    background: var(--surface-2);
    color: var(--text);
    border-color: var(--border-strong);
  }
  .btn-secondary:hover:not(:disabled) {
    background: var(--surface-3);
    border-color: var(--border-focus);
  }

  .btn-ghost {
    background: transparent;
    color: var(--text-muted);
    border-color: transparent;
  }
  .btn-ghost:hover:not(:disabled) {
    background: var(--surface-2);
    color: var(--text);
  }

  .btn-accent-ghost {
    background: transparent;
    color: var(--accent);
    border-color: var(--border);
  }
  .btn-accent-ghost:hover:not(:disabled) {
    background: var(--accent-soft);
    border-color: var(--accent-soft);
    color: var(--accent-hover);
  }

  .btn-danger {
    background: var(--danger-gradient);
    color: var(--text-inverse);
    font-weight: var(--weight-semibold);
    box-shadow: var(--shadow-inset), 0 1px 2px rgba(0, 0, 0, 0.25);
  }
  .btn-danger:hover:not(:disabled) {
    box-shadow: var(--shadow-inset), 0 0 24px var(--error-glow);
    transform: translateY(-1px);
  }

  .btn-label-loading { opacity: 0.6; }

  .btn-primary .dot-loader span,
  .btn-danger .dot-loader span {
    background: var(--text-inverse);
  }
  .dot-loader span {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: currentColor;
    animation: dot-bounce 1.4s infinite ease-in-out both;
  }
  .dot-loader span:nth-child(1) { animation-delay: -0.32s; }
  .dot-loader span:nth-child(2) { animation-delay: -0.16s; }
  .dot-loader span:nth-child(3) { animation-delay: 0s; }
</style>