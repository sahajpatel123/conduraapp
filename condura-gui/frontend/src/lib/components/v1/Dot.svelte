<!--
  Dot — status indicator.

  Per spec §4.4 + §15: status colors only for STATE, never for emphasis.
  A Dot conveys "this is a state, not a decoration." Used in:
    - menu bar / status bar
    - audit log rows (the plum completion dot)
    - permission grant status (green = granted, amber = pending, red = denied)

  Props:
    variant — 'success' | 'warning' | 'error' | 'info' | 'neutral' | 'accent'
    size    — 'xs' (4px) | 'sm' (6px) | 'md' (8px) | 'lg' (10px)
    pulse   — animate the dot (low-amplitude, indicates live/active)
    label   — aria-label override; defaults to variant
-->
<script lang="ts">
  interface Props {
    variant?: 'success' | 'warning' | 'error' | 'info' | 'neutral' | 'accent';
    size?: 'xs' | 'sm' | 'md' | 'lg';
    pulse?: boolean;
    label?: string;
  }

  let {
    variant = 'neutral',
    size = 'sm',
    pulse = false,
    label,
  }: Props = $props();

  const SIZE_PX: Record<NonNullable<Props['size']>, number> = {
    xs: 4,
    sm: 6,
    md: 8,
    lg: 10,
  };
</script>

<span
  class="dot dot--{variant} dot--{size}"
  class:dot--pulse={pulse}
  role="status"
  aria-label={label ?? variant}
  style="--dot-size: {SIZE_PX[size]}px;"
></span>

<style>
  .dot {
    display: inline-block;
    width: var(--dot-size);
    height: var(--dot-size);
    border-radius: var(--radius-pill);
    flex-shrink: 0;
  }

  .dot--success { background-color: var(--success-500); }
  .dot--warning { background-color: var(--warning-500); }
  .dot--error   { background-color: var(--error-500); }
  .dot--info    { background-color: var(--info-500); }
  .dot--neutral { background-color: var(--ink-cool-300); }
  .dot--accent  { background-color: var(--plum-600); }

  /* Low-amplitude pulse — used sparingly, indicates "live" status.
     Much subtler than the Pulse component's vital-sign breath. */
  .dot--pulse {
    animation: dot-pulse 2s ease-in-out infinite;
  }

  @keyframes dot-pulse {
    0%, 100% { opacity: 0.7; }
    50% { opacity: 1.0; }
  }

  @media (prefers-reduced-motion: reduce) {
    .dot--pulse {
      animation: none;
      opacity: 1;
    }
  }
</style>