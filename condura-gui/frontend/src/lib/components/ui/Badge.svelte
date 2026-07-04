<script lang="ts">
  import type { Snippet } from 'svelte'

  type Tone = 'neutral' | 'accent' | 'success' | 'warn' | 'error' | 'info'
  type Size = 'xs' | 'sm' | 'md'

  interface Props {
    tone?: Tone
    size?: Size
    dot?: boolean
    pulse?: boolean
    children?: Snippet
  }

  let { tone = 'neutral', size = 'sm', dot = false, pulse = false, children }: Props = $props()
</script>

<span class="badge badge-{tone} badge-{size}" class:has-dot={dot}>
  {#if dot}<span class="badge-dot" class:anim-glow-pulse={pulse}></span>{/if}
  {#if children}{@render children()}{/if}
</span>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-pill);
    background: var(--surface-2);
    color: var(--text-muted);
    font-family: var(--font-mono);
    font-weight: var(--weight-medium);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
    line-height: 1;
  }

  .badge-xs { padding: 3px 8px;  font-size: 9px;  }
  .badge-sm { padding: 4px 10px; font-size: 10px; }
  .badge-md { padding: 6px 14px; font-size: 11px; }

  .badge-neutral { color: var(--text-muted); }

  .badge-accent {
    color: var(--accent);
    border-color: var(--accent-soft);
    background: var(--accent-faint);
  }
  .badge-success {
    color: var(--success);
    border-color: var(--success-soft);
    background: var(--success-soft);
  }
  .badge-warn {
    color: var(--warn);
    border-color: var(--warn-soft);
    background: var(--warn-soft);
  }
  .badge-error {
    color: var(--error);
    border-color: var(--border-danger);
    background: var(--error-soft);
  }
  .badge-info {
    color: var(--info);
    border-color: var(--info-soft);
    background: var(--info-soft);
  }

  .badge-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
    flex-shrink: 0;
  }
</style>