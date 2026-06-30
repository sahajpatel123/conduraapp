<!--
  Button — primary / secondary / tertiary / destructive × idle / hover / active / disabled / loading.

  Per spec §15: never a red-button for destructive actions. Destructive is
  an outline that turns muted-red on hover. The action slows down, doesn't
  shout.

  Props:
    variant     — 'primary' | 'secondary' | 'tertiary' | 'destructive'
    size        — 'sm' | 'md' | 'lg'
    loading     — shows a subtle progress indicator, disables interaction
    disabled    — explicit disable
    icon        — optional icon snippet, rendered left of label
    children    — button label
    type        — 'button' | 'submit' | 'reset' (default 'button')
    onclick     — click handler
-->
<script lang="ts">
  interface Props {
    variant?: 'primary' | 'secondary' | 'tertiary' | 'destructive';
    size?: 'sm' | 'md' | 'lg';
    loading?: boolean;
    disabled?: boolean;
    icon?: import('svelte').Snippet;
    children?: import('svelte').Snippet;
    type?: 'button' | 'submit' | 'reset';
    onclick?: (e: MouseEvent) => void;
  }

  let {
    variant = 'secondary',
    size = 'md',
    loading = false,
    disabled = false,
    icon,
    children,
    type = 'button',
    onclick,
  }: Props = $props();

  let isDisabled = $derived(disabled || loading);
</script>

<button
  class="btn btn--{variant} btn--{size}"
  class:btn--loading={loading}
  class:btn--disabled={isDisabled}
  disabled={isDisabled}
  type={type}
  onclick={(e) => {
    if (!isDisabled && onclick) onclick(e);
  }}
>
  {#if loading}
    <span class="btn__spinner" aria-hidden="true"></span>
  {:else if icon}
    <span class="btn__icon">{@render icon()}</span>
  {/if}
  {#if children}
    <span class="btn__label">{@render children()}</span>
  {/if}
</button>

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    border-radius: var(--radius-md);
    font-family: var(--font-sans);
    font-weight: 500;
    cursor: pointer;
    user-select: none;
    white-space: nowrap;
    transition:
      background-color var(--duration-fast) var(--ease-standard),
      border-color var(--duration-fast) var(--ease-standard),
      color var(--duration-fast) var(--ease-standard),
      transform var(--duration-fast) var(--ease-standard);
    border: 1px solid transparent;
    /* Press feedback — 1px down on active */
  }
  .btn:active:not(.btn--disabled):not(.btn--loading) {
    transform: translateY(1px);
  }
  .btn--disabled,
  .btn--loading {
    cursor: not-allowed;
  }
  .btn:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: var(--focus-ring-offset);
  }

  /* ── Sizes ─────────────────────────────────────────────────────── */
  .btn--sm {
    height: 28px;
    padding: 0 var(--space-3);
    font-size: var(--text-body-sm-size);
  }
  .btn--md {
    height: 36px;
    padding: 0 var(--space-4);
    font-size: var(--text-body-size);
  }
  .btn--lg {
    height: 44px;
    padding: 0 var(--space-5);
    font-size: var(--text-body-lg-size);
  }

  /* ── Primary (the ONE plum button) ─────────────────────────────── */
  .btn--primary {
    background-color: var(--action-primary-idle-bg);
    color: var(--action-primary-idle-fg);
    border-color: var(--action-primary-idle-bg);
  }
  .btn--primary:hover:not(.btn--disabled):not(.btn--loading) {
    background-color: var(--action-primary-hover-bg);
    border-color: var(--action-primary-hover-bg);
  }
  .btn--primary:active:not(.btn--disabled):not(.btn--loading) {
    background-color: var(--action-primary-active-bg);
    border-color: var(--action-primary-active-bg);
  }
  .btn--primary.btn--disabled,
  .btn--primary.btn--loading {
    background-color: var(--action-primary-disabled-bg);
    border-color: var(--action-primary-disabled-bg);
    color: var(--action-primary-disabled-fg);
  }

  /* ── Secondary (outline / ghost) ──────────────────────────────── */
  .btn--secondary {
    background-color: var(--action-secondary-idle-bg);
    color: var(--action-secondary-idle-fg);
    border-color: var(--action-secondary-idle-border);
  }
  .btn--secondary:hover:not(.btn--disabled):not(.btn--loading) {
    background-color: var(--action-secondary-hover-bg);
    color: var(--action-secondary-hover-fg);
    border-color: var(--action-secondary-hover-border);
  }
  .btn--secondary:active:not(.btn--disabled):not(.btn--loading) {
    background-color: var(--action-secondary-active-bg);
  }
  .btn--secondary.btn--disabled {
    background-color: var(--action-secondary-disabled-bg);
    color: var(--action-secondary-disabled-fg);
    border-color: var(--action-secondary-disabled-border);
  }

  /* ── Tertiary (text-only) ─────────────────────────────────────── */
  .btn--tertiary {
    background-color: transparent;
    border-color: transparent;
    color: var(--action-tertiary-idle-fg);
    padding-left: var(--space-3);
    padding-right: var(--space-3);
  }
  .btn--tertiary:hover:not(.btn--disabled):not(.btn--loading) {
    background-color: var(--action-tertiary-hover-bg);
    color: var(--action-tertiary-hover-fg);
  }
  .btn--tertiary:active:not(.btn--disabled):not(.btn--loading) {
    background-color: var(--action-tertiary-active-bg);
    color: var(--action-tertiary-active-fg);
  }
  .btn--tertiary.btn--disabled {
    color: var(--action-tertiary-disabled-fg);
  }

  /* ── Destructive (outline, never red-button) ──────────────────── */
  .btn--destructive {
    background-color: var(--action-destructive-idle-bg);
    color: var(--action-destructive-idle-fg);
    border-color: var(--action-destructive-idle-border);
  }
  .btn--destructive:hover:not(.btn--disabled):not(.btn--loading) {
    background-color: var(--action-destructive-hover-bg);
    color: var(--action-destructive-hover-fg);
    border-color: var(--action-destructive-hover-border);
  }
  .btn--destructive:active:not(.btn--disabled):not(.btn--loading) {
    background-color: var(--action-destructive-active-bg);
    color: var(--action-destructive-active-fg);
  }

  /* ── Loading spinner (replaces icon) ─────────────────────────── */
  .btn__spinner {
    display: inline-block;
    width: 12px;
    height: 12px;
    border: 1.5px solid currentColor;
    border-top-color: transparent;
    border-radius: var(--radius-pill);
    animation: btn-spin 0.8s linear infinite;
  }

  @keyframes btn-spin {
    to { transform: rotate(360deg); }
  }

  @media (prefers-reduced-motion: reduce) {
    .btn__spinner {
      animation-duration: 1.6s; /* slow but stop motion shouldn't kill it entirely */
    }
  }
</style>