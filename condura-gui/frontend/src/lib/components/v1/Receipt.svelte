<!--
  Receipt — one-line action result.

  Per spec §9.2 (result state): each action produces a one-line receipt
  in the result state of the command surface. Format:
    timestamp mono | verb (sans) | target (sans) | check (icon)

  Props:
    timestamp — e.g., "2 seconds ago", "14:22:07"
    verb      — "renamed", "clicked", "sent", "deleted"
    target    — the file / element / recipient
    state     — 'done' | 'paused' | 'error' | 'pending'
-->
<script lang="ts">
  interface Props {
    timestamp: string;
    verb: string;
    target: string;
    state?: 'done' | 'paused' | 'error' | 'pending';
  }

  let { timestamp, verb, target, state = 'done' }: Props = $props();
</script>

<div class="receipt receipt--{state}" role="status">
  <span class="receipt__time">{timestamp}</span>
  <span class="receipt__verb">{verb}</span>
  <span class="receipt__target">{target}</span>
  <span class="receipt__icon" aria-hidden="true">
    {#if state === 'done'}✓
    {:else if state === 'paused'}⏸
    {:else if state === 'error'}✕
    {:else}○
    {/if}
  </span>
</div>

<style>
  .receipt {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) 0;
    color: var(--content-secondary);
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    line-height: 1.5;
  }

  .receipt__time {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.02em;
    flex-shrink: 0;
    min-width: 80px;
  }

  .receipt__verb {
    color: var(--content-primary);
    font-weight: 500;
  }

  .receipt__target {
    color: var(--content-secondary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
    flex: 1;
  }

  .receipt__icon {
    flex-shrink: 0;
    font-family: var(--font-mono);
    font-size: var(--text-body-size);
    width: 16px;
    text-align: center;
  }

  .receipt--done .receipt__icon { color: var(--success-500); }
  .receipt--paused .receipt__icon { color: var(--warning-500); }
  .receipt--error .receipt__icon { color: var(--error-500); }
  .receipt--pending .receipt__icon { color: var(--content-tertiary); }
</style>