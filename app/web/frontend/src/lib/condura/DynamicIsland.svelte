<script lang="ts">
  import Pulse from './Pulse.svelte';

  // The agent's face in the titlebar — a morphing status pill (SCREEN_TITLEBAR
  // §3). Six states, each with exact copy (§8.4 rule 4 — do not paraphrase).
  // The pill morphs by WIDTH + COLOR only (never scale, no layout shift); both
  // transitions run on the same motion token so the morph is synchronous
  // (§4.5). aria-live announces the agent's phase to screen readers.
  let {
    phase = 'idle',
    task = '',
    reason = '',
    onactivate,
  }: {
    phase?: 'idle' | 'thinking' | 'streaming' | 'consent' | 'halted' | 'offline';
    task?: string;
    reason?: string;
    onactivate?: () => void;
  } = $props();

  type PulsePhase = 'idle' | 'thinking' | 'awaiting' | 'acting' | 'consent' | 'error' | 'ok';

  // width: the only animated dimension. Kept ≤188px so the longest label
  // ("CONSENT REQUIRED") fits without truncation (§8.4 rule 4).
  const MAP: Record<
    string,
    { width: number; pulse: PulsePhase | null; live: 'polite' | 'assertive' }
  > = {
    idle: { width: 124, pulse: 'idle', live: 'polite' },
    thinking: { width: 156, pulse: 'thinking', live: 'polite' },
    streaming: { width: 188, pulse: 'acting', live: 'polite' },
    consent: { width: 188, pulse: 'consent', live: 'assertive' },
    halted: { width: 188, pulse: 'error', live: 'assertive' },
    offline: { width: 124, pulse: null, live: 'polite' },
  };

  let cfg = $derived(MAP[phase] ?? MAP.idle);

  // Exact copy per §3 / §8.4 rule 4. text-transform in CSS uppercases the
  // interpolated task/reason.
  let label = $derived.by(() => {
    switch (phase) {
      case 'thinking':
        return task ? `THINKING · ${truncate(task, 16)}` : 'THINKING';
      case 'streaming':
        return task ? `STREAMING · ${truncate(task, 18)}` : 'STREAMING';
      case 'consent':
        return 'CONSENT REQUIRED';
      case 'halted':
        return reason ? `HALTED · ${truncate(reason, 14)}` : 'HALTED';
      case 'offline':
        return 'OFFLINE';
      default:
        return 'IDLE · LISTENING';
    }
  });

  // Only the consent state is interactive — activating moves focus to the
  // always-mounted ConsentModal (§3.4). Everything else is a status readout.
  let interactive = $derived(phase === 'consent' && typeof onactivate === 'function');

  function truncate(s: string, n: number): string {
    return s.length > n ? s.slice(0, n - 1) + '…' : s;
  }

  function activate() {
    if (interactive) onactivate?.();
  }

  function onKey(e: KeyboardEvent) {
    if (!interactive) return;
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      activate();
    }
  }
</script>

<div
  class="island tb-island"
  class:interactive
  data-phase={phase}
  style="width:{cfg.width}px"
  role={interactive ? 'button' : 'status'}
  tabindex={interactive ? 0 : -1}
  aria-live={cfg.live}
  aria-atomic="true"
  onclick={activate}
  onkeydown={onKey}
>
  {#if phase === 'consent'}
    <span class="info" aria-hidden="true">&#x24D8;</span>
  {/if}
  {#if cfg.pulse}
    <Pulse phase={cfg.pulse} size={6} />
  {:else}
    <span class="static-dot" aria-hidden="true"></span>
  {/if}
  <span class="ilabel">{label}</span>
</div>

<style>
  .island {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    height: 28px;
    padding: 0 var(--space-3);
    background: var(--surface-card);
    border: 1px solid var(--synapse);
    border-radius: var(--r-pill);
    box-shadow: var(--shadow-paper);
    overflow: hidden;
    /* WIDTH + COLOR morph, one token, synchronous (§4.5). Never scale. */
    transition:
      width var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      background-color var(--dur) var(--ease),
      color var(--dur) var(--ease);
  }
  .island.interactive {
    cursor: pointer;
  }
  .island[data-phase='consent'] {
    border-color: var(--warn);
    background: color-mix(in oklab, var(--pollen-light) 35%, var(--surface-card));
  }
  .island[data-phase='halted'] {
    border-color: var(--danger);
    background: color-mix(in oklab, var(--danger) 8%, var(--surface-card));
  }
  .island[data-phase='offline'] {
    border-color: var(--ink-mute);
    background: var(--surface-card);
  }
  .ilabel {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--content-mute);
    white-space: nowrap;
  }
  .info {
    font-size: 13px;
    line-height: 1;
    color: var(--warn);
    flex: none;
  }
  .static-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--ink-mute);
    flex: none;
  }
</style>
