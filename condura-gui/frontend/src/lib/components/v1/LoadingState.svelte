<!--
  LoadingState — per-state loading composition.

  Per motion agent §9 ("Loading State"): four distinct loading moments,
  none of them a generic spinner:
    - Cold app launch (< 500ms target): no splash, just the chrome
    - LLM thinking (1-15s): progress bar with model name + elapsed
    - Computer-use action in progress: flight-recorder chip
    - Long-running task (30s+): full panel with task name + pause

  This component abstracts the four moments. Each variant renders the
  appropriate composition.

  Props:
    kind        — 'cold' | 'thinking' | 'computer-use' | 'long-running'
    elapsedMs   — current elapsed time (for thinking/long)
    modelName   — current model name (for thinking)
    taskName    — task name (for long-running)
    stage       — current stage description (for computer-use/long-running)
    onpause     — pause handler (for long-running)
-->
<script lang="ts">
  import ProgressBar from './ProgressBar.svelte';
  import Dot from './Dot.svelte';
  import Button from './Button.svelte';

  type Kind = 'cold' | 'thinking' | 'computer-use' | 'long-running';

  interface Props {
    kind: Kind;
    elapsedMs?: number;
    modelName?: string;
    taskName?: string;
    stage?: string;
    onpause?: () => void;
  }

  let { kind, elapsedMs = 0, modelName, taskName, stage, onpause }: Props = $props();
</script>

<div class="loading loading--{kind}" role="status" aria-live="polite">
  {#if kind === 'cold'}
    <div class="loading__cold">
      <Dot variant="neutral" size="sm" pulse />
      <span class="loading__cold-text">starting</span>
    </div>
  {:else if kind === 'thinking'}
    <ProgressBar elapsedMs={elapsedMs} state="thinking" modelName={modelName} />
  {:else if kind === 'computer-use'}
    <div class="loading__cu">
      <Dot variant="accent" size="sm" pulse />
      <span class="loading__cu-stage">{stage ?? 'acting'}</span>
      <span class="loading__cu-elapsed">{(elapsedMs / 1000).toFixed(1)}s</span>
    </div>
  {:else if kind === 'long-running'}
    <div class="loading__long">
      <div class="loading__long-head">
        <span class="loading__long-name">{taskName ?? 'Working…'}</span>
        <span class="loading__long-elapsed">{(elapsedMs / 1000).toFixed(0)}s</span>
      </div>
      {#if stage}
        <div class="loading__long-stage">{stage}</div>
      {/if}
      <ProgressBar elapsedMs={elapsedMs} state="executing" />
      <div class="loading__long-actions">
        <Button variant="tertiary" size="sm" onclick={onpause}>⏸ Pause</Button>
      </div>
    </div>
  {/if}
</div>

<style>
  .loading {
    font-family: var(--font-sans);
  }

  /* Cold: barely-there indicator */
  .loading__cold {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  /* Computer-use: flight-recorder chip */
  .loading__cu {
    display: inline-flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-3);
    background-color: var(--paper-warm-50);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-sm);
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-secondary);
    letter-spacing: 0.02em;
  }
  .loading__cu-stage {
    color: var(--content-primary);
    text-transform: uppercase;
  }
  .loading__cu-elapsed {
    color: var(--content-tertiary);
    font-variant-numeric: tabular-nums;
  }

  /* Long-running: full panel */
  .loading__long {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-4);
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
  }
  .loading__long-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .loading__long-name {
    font-family: var(--font-serif);
    font-size: var(--text-body-size);
    color: var(--content-primary);
  }
  .loading__long-elapsed {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    font-variant-numeric: tabular-nums;
  }
  .loading__long-stage {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-secondary);
    letter-spacing: 0.02em;
  }
  .loading__long-actions {
    display: flex;
    justify-content: flex-end;
  }
</style>