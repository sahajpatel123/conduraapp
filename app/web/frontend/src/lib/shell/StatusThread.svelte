<script lang="ts">
  /**
   * StatusThread — Bottom living status indicator.
   * 
   * A thin horizontal bar with a living pulse dot showing agent
   * state. Replaces the fixed StatusBar popover. Minimal, elegant,
   * always visible but never intrusive.
   */
  import { PulseDot, InkText } from '$lib/components/living'
  import type { PulsePhase } from '$lib/components/living'

  interface Props {
    agentPhase?: PulsePhase
    agentLabel?: string
    onKill?: () => void
    onPause?: () => void
    /** If true, show a halted/stopped state */
    halted?: boolean
  }

  let {
    agentPhase = 'idle',
    agentLabel = 'Ready',
    onKill,
    onPause,
    halted = false,
  }: Props = $props()
</script>

<footer
  class="lp lp-status-thread"
  style="
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--lp-space-2) var(--lp-space-6);
    border-top: 1px solid var(--lp-ink-ghost);
    background: var(--lp-paper);
    min-height: 28px;
    position: relative;
    z-index: 10;
  "
>
  <!-- Left: Agent state -->
  <div style="display: flex; align-items: center; gap: var(--lp-space-2);">
    <PulseDot phase={halted ? 'error' : agentPhase} size={5} />
    <InkText kind="caption" tone="ink-mute" style="letter-spacing: 0.02em;">
      {halted ? 'Halted' : agentLabel}
    </InkText>
  </div>

  <!-- Right: Actions -->
  <div style="display: flex; align-items: center; gap: var(--lp-space-2);">
    {#if onPause && !halted}
      <button
        type="button"
        class="lp-status-btn lp-focus"
        onclick={onPause}
        aria-label="Pause agent"
        title="Pause agent"
        style="
          display: inline-flex;
          align-items: center;
          gap: 4px;
          padding: 2px 8px;
          border: none;
          border-radius: var(--lp-radius-xs);
          background: transparent;
          color: var(--lp-ink-mute);
          font-family: var(--lp-font-mono);
          font-size: var(--lp-text-micro);
          cursor: pointer;
          transition: background var(--lp-dur-fast) ease,
                      color var(--lp-dur-fast) ease;
        "
      >
        <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
          <rect x="6" y="4" width="4" height="16" rx="1" />
          <rect x="14" y="4" width="4" height="16" rx="1" />
        </svg>
        Pause
      </button>
    {/if}

    {#if onKill}
      <button
        type="button"
        class="lp-status-btn lp-focus"
        onclick={onKill}
        aria-label={halted ? 'Resume agent' : 'Stop agent'}
        title={halted ? 'Resume agent' : 'Stop agent'}
        style="
          display: inline-flex;
          align-items: center;
          gap: 4px;
          padding: 2px 8px;
          border: none;
          border-radius: var(--lp-radius-xs);
          background: transparent;
          color: {halted ? 'var(--lp-ok)' : 'var(--lp-danger)'};
          font-family: var(--lp-font-mono);
          font-size: var(--lp-text-micro);
          cursor: pointer;
          transition: background var(--lp-dur-fast) ease,
                      color var(--lp-dur-fast) ease;
        "
      >
        {#if halted}
          <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
            <path d="M8 5v14l11-7z" />
          </svg>
          Resume
        {:else}
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
            <rect x="6" y="4" width="4" height="16" rx="1" />
            <rect x="14" y="4" width="4" height="16" rx="1" />
          </svg>
          Stop
        {/if}
      </button>
    {/if}
  </div>
</footer>

<style>
  .lp-status-btn:hover {
    background: var(--lp-paper-warm);
  }
  .lp-status-btn:active {
    transform: scale(0.97);
  }
</style>
