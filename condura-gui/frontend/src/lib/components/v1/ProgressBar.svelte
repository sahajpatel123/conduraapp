<!--
  ProgressBar — thin mono progress indicator.

  Per spec §9 (Loading State): a thin progress bar in mono text.
  "▌▌▌▌▌░░░░░  thinking…  1.2s · claude-sonnet-4-6"

  Per motion agent §6: NO spinners for "loading". Use a heartbeat that
  scales with pause duration. This component shows the live progress.

  Props:
    elapsedMs    — elapsed milliseconds
    state        — current agent state
    modelName    — current model name (optional)
-->
<script lang="ts">
  interface Props {
    elapsedMs: number;
    state?: 'thinking' | 'tool-call' | 'verifying' | 'executing';
    modelName?: string;
  }

  let { elapsedMs, state = 'thinking', modelName }: Props = $props();

  // Format elapsed time. Spec uses 0.1s precision.
  function fmtElapsed(ms: number): string {
    if (ms < 1000) return `${(ms / 1000).toFixed(1)}s`;
    return `${(ms / 1000).toFixed(1)}s`;
  }

  const STATE_LABEL: Record<NonNullable<Props['state']>, string> = {
    'thinking':   'thinking',
    'tool-call':  'calling tool',
    'verifying':  'verifying',
    'executing':  'acting',
  };
</script>

<div class="progress" role="status" aria-live="polite">
  <span class="progress__bar" aria-hidden="true">
    <span class="progress__fill" style="--progress-pct: {Math.min(95, (elapsedMs / 15000) * 100)}%"></span>
  </span>
  <span class="progress__text">
    <span class="progress__state">{STATE_LABEL[state]}</span>
    <span class="progress__sep">·</span>
    <span class="progress__elapsed">{fmtElapsed(elapsedMs)}</span>
    {#if modelName}
      <span class="progress__sep">·</span>
      <span class="progress__model">{modelName}</span>
    {/if}
  </span>
</div>

<style>
  .progress {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-2) 0;
  }

  /* The bar — character-driven, per spec */
  .progress__bar {
    display: flex;
    height: 4px;
    background-color: var(--paper-warm-50);
    border-radius: var(--radius-xs);
    overflow: hidden;
  }

  .progress__fill {
    width: var(--progress-pct);
    background-color: var(--content-accent);
    transition: width var(--duration-base) var(--ease-standard);
  }

  .progress__text {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    letter-spacing: 0.02em;
  }

  .progress__state {
    color: var(--content-secondary);
    font-weight: 500;
  }
  .progress__sep {
    color: var(--content-muted);
  }
  .progress__elapsed,
  .progress__model {
    font-variant-numeric: tabular-nums;
  }
</style>