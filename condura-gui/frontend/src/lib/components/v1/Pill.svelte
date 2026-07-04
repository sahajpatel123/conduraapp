<!--
  Pill — status pill. Shape + icon + text, not color.

  Per spec §15: status colors only for STATE in chrome. A pill that says
  "Done", "Paused", "Error" with shape + text. Optional color is muted.

  Props:
    variant — 'success' | 'warning' | 'error' | 'info' | 'neutral' | 'accent'
    size    — 'xs' | 'sm' | 'md'
    label   — pill text
-->
<script lang="ts">
  interface Props {
    variant?: 'success' | 'warning' | 'error' | 'info' | 'neutral' | 'accent';
    size?: 'xs' | 'sm' | 'md';
    label?: string;
  }

  let { variant = 'neutral', size = 'sm', label }: Props = $props();
</script>

<span class="pill pill--{variant} pill--{size}" role="status">
  <span class="pill__dot" aria-hidden="true"></span>
  <span class="pill__label">{label}</span>
</span>

<style>
  .pill {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    border-radius: var(--radius-pill);
    font-family: var(--font-sans);
    font-weight: 500;
    white-space: nowrap;
  }
  .pill--xs {
    height: 18px;
    padding: 0 var(--space-2);
    font-size: 10px;
  }
  .pill--sm {
    height: 22px;
    padding: 0 var(--space-3);
    font-size: var(--text-caption-size);
  }
  .pill--md {
    height: 28px;
    padding: 0 var(--space-3);
    font-size: var(--text-body-sm-size);
  }

  /* Leading dot — the only color cue in the pill. */
  .pill__dot {
    width: 6px;
    height: 6px;
    border-radius: var(--radius-pill);
    background-color: currentColor;
    flex-shrink: 0;
  }

  /* Variant bg + fg, dot uses currentColor */
  .pill--success {
    background-color: var(--status-success-bg);
    color: var(--status-success-fg);
  }
  .pill--warning {
    background-color: var(--status-warning-bg);
    color: var(--status-warning-fg);
  }
  .pill--error {
    background-color: var(--status-error-bg);
    color: var(--status-error-fg);
  }
  .pill--info {
    background-color: var(--status-info-bg);
    color: var(--status-info-fg);
  }
  .pill--neutral {
    background-color: var(--status-neutral-bg);
    color: var(--status-neutral-fg);
  }
  .pill--accent {
    background-color: var(--plum-50);
    color: var(--plum-700);
  }
  .pill--accent .pill__dot {
    background-color: var(--plum-600);
    color: var(--plum-600);
  }
</style>