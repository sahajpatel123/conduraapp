<!--
  StatusBar — the ambient surface (menu bar / tray status).

  Per spec §8.5: a single 16×16 glyph in the OS chrome with a status dot
  overlay and a count badge when the agent is mid-task. Click reveals
  a small popover showing current task, queued tasks, and a pause toggle.

  Per spec §6 (The Survival Rule): the kill switch MUST be reachable from
  the status bar. It's the last item in the popover.

  Props:
    activeTask    — current task name (or null if idle)
    queuedCount   — number of queued tasks
    agentState    — 'idle' | 'thinking' | 'awaiting' | 'error'
    onpause       — handler for pause toggle
    onkill        — handler for kill switch
    onopen        — handler for opening the main window
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import Pulse from './Pulse.svelte';
  import Dot from './Dot.svelte';
  import Icon from './icons/Icon.svelte';

  interface Props {
    activeTask?: string | null;
    queuedCount?: number;
    agentState?: 'idle' | 'thinking' | 'awaiting' | 'error';
    onpause?: () => void;
    onkill?: () => void;
    onopen?: () => void;
  }

  let {
    activeTask = null,
    queuedCount = 0,
    agentState = 'idle',
    onpause,
    onkill,
    onopen,
  }: Props = $props();

  let paused = $state(false);
  let popoverOpen = $state(false);

  // Cursor awareness — the Pulse brightens when the cursor approaches
  // the top of the screen. The agent is "looking up" — aware of user
  // attention. Distance threshold: 80px from top edge.
  let cursorProximity = $state(0);  // 0..1, where 1 = at the top
  let hostEl: HTMLDivElement | undefined = $state();

  onMount(() => {
    if (typeof window === 'undefined') return;
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;

    let raf = 0;
    let targetProx = 0;
    let lastMove = 0;

    const onMove = (e: PointerEvent) => {
      lastMove = Date.now();
      // 80px from the top = full proximity (1). 200px+ = no proximity (0).
      const y = e.clientY;
      targetProx = Math.max(0, Math.min(1, (200 - y) / 120));
    };

    const onLeave = () => { targetProx = 0; };

    // Lerp the proximity over time so the brightening is smooth
    const tick = () => {
      cursorProximity += (targetProx - cursorProximity) * 0.06;
      // If the user hasn't moved in 60s, fade the proximity
      if (Date.now() - lastMove > 60000) {
        targetProx = 0;
      }
      raf = requestAnimationFrame(tick);
    };

    window.addEventListener('pointermove', onMove, { passive: true });
    window.addEventListener('pointerleave', onLeave);
    raf = requestAnimationFrame(tick);

    return () => {
      cancelAnimationFrame(raf);
      window.removeEventListener('pointermove', onMove);
      window.removeEventListener('pointerleave', onLeave);
    };
  });

  // The Pulse state reflects the agent's actual state, modulated by cursor proximity
  let effectivePulseState = $derived(
    agentState === 'thinking' ? 'thinking' :
    agentState === 'awaiting' ? 'awaiting' :
    agentState === 'error' ? 'error' :
    'idle'
  );

  function togglePopover() {
    popoverOpen = !popoverOpen;
  }

  function closePopover() {
    popoverOpen = false;
  }
</script>

<svelte:window onkeydown={(e) => { if (e.key === 'Escape') closePopover(); }} />

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="status-host" onclick={closePopover} onkeydown={(e) => { if (e.key === 'Escape') closePopover(); }}>
  <button
    class="trigger"
    type="button"
    onclick={(e) => { e.stopPropagation(); togglePopover(); }}
    aria-label="Synaptic status"
    aria-expanded={popoverOpen}
  >
    <Pulse state={effectivePulseState} size="sm" label="Synaptic" />
    {#if queuedCount > 0}
      <span class="badge" aria-label="{queuedCount} queued">{queuedCount}</span>
    {/if}
  </button>

  {#if popoverOpen}
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="popover"
      role="dialog"
      aria-label="Synaptic menu"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
    >
      <div class="popover__head">
        <div class="popover__status">
          <Dot variant={agentState === 'error' ? 'error' : agentState === 'thinking' ? 'accent' : 'neutral'} size="sm" pulse={agentState !== 'idle'} />
          <span class="popover__status-text">
            {agentState === 'thinking' ? 'Working' : agentState === 'error' ? 'Error' : paused ? 'Paused' : 'Idle'}
          </span>
        </div>
      </div>

      {#if activeTask}
        <div class="popover__section">
          <div class="popover__label">Current task</div>
          <div class="popover__task">{activeTask}</div>
        </div>
      {/if}

      {#if queuedCount > 0}
        <div class="popover__section">
          <div class="popover__label">Queued</div>
          <div class="popover__queue-count">{queuedCount}</div>
        </div>
      {/if}

      <div class="popover__divider"></div>

      <div class="popover__actions">
        <button class="action" type="button" onclick={onopen}>
          <span class="action__icon"><Icon name="home" size="sm" /></span>
          <span>Open Synaptic</span>
          <span class="action__hint">⌘⇧Space</span>
        </button>
        <button
          class="action"
          type="button"
          onclick={() => { paused = !paused; onpause?.(); }}
        >
          <span class="action__icon">
            <Icon name={paused ? 'play' : 'pause'} size="sm" />
          </span>
          <span>{paused ? 'Resume agent' : 'Pause agent'}</span>
        </button>
        <button class="action action--danger" type="button" onclick={onkill}>
          <span class="action__icon"><Icon name="power" size="sm" /></span>
          <span>Stop everything</span>
          <span class="action__hint">⌘⇧⎋</span>
        </button>
      </div>

      <div class="popover__footer">
        <span class="popover__footer-text">Synaptic</span>
      </div>
    </div>
  {/if}
</div>

<style>
  .status-host {
    position: fixed;
    top: var(--space-3);
    right: var(--space-3);
    z-index: var(--z-sticky);
  }

  .trigger {
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-pill);
    cursor: pointer;
    transition:
      border-color var(--duration-fast) var(--ease-standard),
      background-color var(--duration-fast) var(--ease-standard);
    padding: 0;
  }
  .trigger:hover {
    border-color: var(--border-strong);
    background-color: var(--paper-warm-50);
  }
  .trigger:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: 2px;
  }

  .badge {
    position: absolute;
    top: -4px;
    right: -4px;
    min-width: 16px;
    height: 16px;
    padding: 0 4px;
    background-color: var(--plum-600);
    color: var(--paper-warm-0);
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    line-height: 16px;
    border-radius: var(--radius-pill);
    text-align: center;
    border: 1.5px solid var(--surface-base);
  }

  /* Popover */
  .popover {
    position: absolute;
    top: calc(100% + var(--space-2));
    right: 0;
    width: 280px;
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-3);
    padding: var(--space-3);
    z-index: var(--z-tooltip);
    animation: popover-in var(--duration-base) var(--ease-decelerate) both;
  }

  @keyframes popover-in {
    from {
      opacity: 0;
      transform: translateY(-4px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .popover__head {
    padding: var(--space-2) var(--space-3) var(--space-3);
  }

  .popover__status {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .popover__status-text {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--content-secondary);
  }

  .popover__section {
    padding: var(--space-2) var(--space-3) var(--space-3);
  }

  .popover__label {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--content-tertiary);
    margin-bottom: var(--space-1);
  }

  .popover__task {
    font-family: var(--font-serif);
    font-size: var(--text-body-sm-size);
    color: var(--content-primary);
    line-height: 1.4;
  }

  .popover__queue-count {
    font-family: var(--font-mono);
    font-size: var(--text-body-size);
    color: var(--content-secondary);
    font-variant-numeric: tabular-nums;
  }

  .popover__divider {
    height: 1px;
    background-color: var(--border-subtle);
    margin: var(--space-2) 0;
  }

  .popover__actions {
    display: flex;
    flex-direction: column;
    gap: 1px;
  }

  .action {
    display: grid;
    grid-template-columns: 20px 1fr auto;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-2) var(--space-3);
    background-color: transparent;
    border: none;
    border-radius: var(--radius-sm);
    color: var(--content-secondary);
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    cursor: pointer;
    text-align: left;
    transition: background-color var(--duration-fast) var(--ease-standard);
  }
  .action:hover {
    background-color: var(--paper-warm-50);
    color: var(--content-primary);
  }
  .action:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: -2px;
  }

  .action__icon {
    text-align: center;
    font-family: var(--font-mono);
    font-size: 14px;
    color: var(--content-tertiary);
  }
  .action:hover .action__icon {
    color: var(--content-primary);
  }

  .action__hint {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--content-muted);
    letter-spacing: 0.02em;
  }

  .action--danger {
    color: var(--error-500);
  }
  .action--danger .action__icon {
    color: var(--error-500);
  }
  .action--danger:hover {
    background-color: var(--error-100);
    color: var(--error-700);
  }
  .action--danger:hover .action__icon {
    color: var(--error-700);
  }

  .popover__footer {
    padding: var(--space-3) var(--space-3) var(--space-2);
    border-top: 1px solid var(--border-subtle);
    margin-top: var(--space-2);
  }
  .popover__footer-text {
    font-family: var(--font-serif);
    font-style: italic;
    font-size: var(--text-caption-size);
    color: var(--content-muted);
  }

  @media (prefers-reduced-motion: reduce) {
    .popover {
      animation: none;
    }
  }
</style>