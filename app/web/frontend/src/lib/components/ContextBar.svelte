<script lang="ts">
  import { spend } from '../stores/spend.svelte'
  import { daemon } from '../stores/daemon.svelte'
  import IrisOrb from './IrisOrb.svelte'
  import Kbd from './ui/Kbd.svelte'

  // The floating Context Bar — the agent's face at the top of the
  // content stage. Left: the route title in editorial Fraunces.
  // Center: a command field (⌘K). Right: a live presence capsule
  // (a tiny Iris + state word + spend) — the same organism as the
  // dock orb, in miniature.
  interface Props {
    title?: string
    eyebrow?: string
    orbState?: 'idle' | 'listening' | 'thinking' | 'acting' | 'consent' | 'offline'
  }
  let { title = 'Chat', eyebrow, orbState = 'idle' }: Props = $props()

  const stateWord = $derived(
    orbState === 'thinking' ? 'Thinking' :
    orbState === 'acting' ? 'Acting' :
    orbState === 'listening' ? 'Listening' :
    orbState === 'consent' ? 'Awaiting you' :
    orbState === 'offline' ? 'Offline' : 'Idle'
  )

  const spent = $derived(spend.summary ? `$${spend.summary.spent.toFixed(2)}` : '$0.00')

  function openPalette(): void {
    window.dispatchEvent(new CustomEvent('condura:open-palette'))
  }
</script>

<header class="bar">
  <div class="lead">
    {#if eyebrow}<span class="eyebrow">{eyebrow}</span>{/if}
    <h1 class="title">{title}</h1>
  </div>

  <button class="command" onclick={openPalette} aria-label="Open command palette">
    <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round">
      <circle cx="7" cy="7" r="4.2" /><path d="M10.2 10.2 13.5 13.5" />
    </svg>
    <span class="ph">Ask or run a command…</span>
    <span class="keys"><Kbd label="⌘" /><Kbd label="K" /></span>
  </button>

  <div class="presence" class:off={orbState === 'offline'} title="Agent {stateWord}">
    <IrisOrb state={orbState} size={14} title={`Agent ${stateWord}`} />
    <span class="state">{stateWord}</span>
    <span class="dot" aria-hidden="true"></span>
    <span class="spend" title="Spent today">{spent}</span>
  </div>
</header>

<style>
  .bar {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    height: var(--topbar-height);
    padding: 0 8px 0 18px;
    border-radius: var(--radius-lg);
    background: var(--glass-bg);
    backdrop-filter: var(--blur-base);
    -webkit-backdrop-filter: var(--blur-base);
    border: 1px solid var(--border);
    box-shadow: var(--shadow-md), var(--inset-hair);
    flex-shrink: 0;
  }

  .lead { display: flex; flex-direction: column; justify-content: center; min-width: 0; }
  .eyebrow {
    font-size: var(--size-2xs);
    font-weight: var(--weight-semibold);
    letter-spacing: var(--tracking-wider);
    text-transform: uppercase;
    color: var(--text-faint);
    line-height: 1;
  }
  .title {
    font-family: var(--font-display);
    font-weight: var(--weight-medium);
    font-size: var(--display-sm);
    letter-spacing: var(--tracking-tight);
    color: var(--text);
    line-height: 1.05;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .command {
    margin-left: auto;
    display: flex;
    align-items: center;
    gap: var(--space-2);
    height: 36px;
    width: min(380px, 38vw);
    padding: 0 8px 0 12px;
    border-radius: var(--radius-md);
    background: var(--surface-1);
    border: 1px solid var(--border);
    color: var(--text-faint);
    transition: border-color var(--transition-fast), background var(--transition-fast);
  }
  .command:hover { border-color: var(--border-strong); background: var(--surface-2); }
  .command svg { width: 15px; height: 15px; color: var(--text-muted); flex-shrink: 0; }
  .ph { flex: 1; text-align: left; font-size: var(--size-sm); color: var(--text-faint); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .keys { display: inline-flex; gap: 3px; flex-shrink: 0; }

  .presence {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    height: 36px;
    padding: 0 12px 0 10px;
    border-radius: var(--radius-pill);
    background: var(--surface-1);
    border: 1px solid var(--border);
  }
  .presence .state { font-size: var(--size-xs); font-weight: var(--weight-medium); color: var(--text); }
  .presence.off .state { color: var(--text-faint); }
  .presence .dot { width: 3px; height: 3px; border-radius: 50%; background: var(--text-faint); }
  .presence .spend { font-family: var(--font-mono); font-size: var(--size-xs); color: var(--text-muted); }

  @media (max-width: 720px) {
    .command { width: 44px; padding: 0; justify-content: center; }
    .ph, .keys { display: none; }
    .presence .spend, .presence .dot { display: none; }
  }
</style>
