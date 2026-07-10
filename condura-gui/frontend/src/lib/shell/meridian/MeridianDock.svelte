<script lang="ts">
  import type { RouteId } from './routes'
  import { halt } from '../../stores/halt.svelte'
  import { pendingCount } from '../../stores/pending.svelte'

  interface Props {
    route: RouteId
    onnavigate: (r: RouteId) => void
  }
  let { route, onnavigate }: Props = $props()

  const PRIMARY: { id: RouteId; label: string }[] = [
    { id: 'chat', label: 'Ask' },
    { id: 'hub', label: 'Hub' },
    { id: 'skills', label: 'Skills' },
    { id: 'sync', label: 'Sync' },
    { id: 'audit', label: 'Audit' },
    { id: 'channels', label: 'Channels' },
    { id: 'delegation', label: 'Agents' },
  ]
  const MORE: { id: RouteId; label: string }[] = [
    { id: 'account', label: 'Account' },
    { id: 'settings', label: 'Settings' },
    { id: 'about', label: 'About' },
  ]
  const pending = $derived($pendingCount)

  let dockEl = $state<HTMLElement | null>(null)

  $effect(() => {
    route
    if (!dockEl) return
    const active = dockEl.querySelector<HTMLElement>('.tab.active')
    if (!active) return
    const reduce =
      typeof matchMedia !== 'undefined' && matchMedia('(prefers-reduced-motion: reduce)').matches
    queueMicrotask(() => {
      active.scrollIntoView({
        inline: 'center',
        block: 'nearest',
        behavior: reduce ? 'auto' : 'smooth',
      })
    })
  })
</script>
<nav class="dock" aria-label="Primary" bind:this={dockEl}>
  <div class="primary">
    {#each PRIMARY as item (item.id)}
      <button type="button" class="tab" class:active={route === item.id} onclick={() => onnavigate(item.id)}>
        {item.label}
        {#if item.id === 'delegation' && pending > 0}<span class="badge">{pending}</span>{/if}
      </button>
    {/each}
  </div>
  <div class="divider" aria-hidden="true"></div>
  <div class="more">
    {#each MORE as item (item.id)}
      <button type="button" class="tab quiet" class:active={route === item.id} onclick={() => onnavigate(item.id)}>
        {item.label}
      </button>
    {/each}
    <button type="button" class="halt" onclick={() => void halt.halt()} aria-label="Halt agent">Halt</button>
  </div>
</nav>

<style>
  .dock {
    position: absolute; left: 16px; right: 16px; bottom: 16px; z-index: 20;
    display: flex; align-items: center; justify-content: flex-start; gap: 8px;
    padding: 10px 12px; border-radius: 22px;
    background: color-mix(in oklab, var(--md-surface) 78%, transparent);
    border: 1px solid var(--md-line-strong);
    backdrop-filter: blur(18px) saturate(1.2);
    -webkit-backdrop-filter: blur(18px) saturate(1.2);
    box-shadow: var(--md-shadow);
    animation: md-dock-up 640ms var(--md-ease) 120ms both;
    overflow-x: auto; scrollbar-width: none;
    -webkit-overflow-scrolling: touch;
    scroll-padding-inline: 8px;
    mask-image: linear-gradient(90deg, transparent 0, #000 12px, #000 calc(100% - 12px), transparent);
    -webkit-mask-image: linear-gradient(90deg, transparent 0, #000 12px, #000 calc(100% - 12px), transparent);
  }
  .dock::-webkit-scrollbar { display: none; }
  .primary, .more { display: flex; align-items: center; gap: 4px; flex: none; }
  .divider {
    width: 1px; align-self: stretch; margin: 4px 2px;
    background: var(--md-line-strong); flex: none;
  }
  .dock:hover {
    box-shadow: var(--md-shadow-lift);
    border-color: color-mix(in oklab, var(--md-cobalt) 22%, var(--md-line-strong));
  }
  .tab {
    position: relative;
    appearance: none; border: 0; background: transparent; color: var(--md-ink-mute);
    font-family: var(--md-font-sans); font-size: 13px; font-weight: 600;
    padding: 10px 14px; border-radius: 999px; cursor: pointer; white-space: nowrap;
    transition: color var(--md-dur) var(--md-ease), background var(--md-dur) var(--md-ease), transform 160ms var(--md-spring), box-shadow var(--md-dur) var(--md-ease);
  }
  .tab.quiet { font-weight: 500; color: var(--md-ink-faint); }
  .tab:hover { color: var(--md-ink); background: color-mix(in oklab, var(--md-cobalt) 8%, transparent); transform: translateY(-1px); }
  .tab.active {
    color: #fff; background: var(--md-cobalt);
    box-shadow: 0 8px 20px -10px color-mix(in oklab, var(--md-cobalt) 70%, transparent);
    animation: md-tab-in 280ms var(--md-spring) both;
  }
  .tab:active { transform: scale(0.96); }
  .tab:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
    color: var(--md-ink);
  }
  .tab.active:focus-visible {
    box-shadow: 0 8px 20px -10px color-mix(in oklab, var(--md-cobalt) 70%, transparent), var(--md-focus);
  }
  .halt:focus-visible {
    outline: none;
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-halt) 22%, transparent);
  }
  .badge {
    display: inline-flex; align-items: center; justify-content: center;
    min-width: 16px; height: 16px; margin-left: 4px; padding: 0 4px; border-radius: 999px;
    background: #fff; color: var(--md-cobalt); font-size: 10px; font-weight: 800;
    animation: md-rise 280ms var(--md-spring) both;
  }
  :root[data-mode='dark'] .badge {
    background: var(--md-surface); color: var(--md-cobalt);
  }
  .halt {
    appearance: none; margin-left: 6px;
    border: 1px solid color-mix(in oklab, var(--md-halt) 40%, transparent);
    background: color-mix(in oklab, var(--md-halt) 10%, transparent);
    color: var(--md-halt); font-size: 12px; font-weight: 800; letter-spacing: 0.04em;
    text-transform: uppercase; padding: 8px 14px; border-radius: 999px; cursor: pointer;
    transition: background var(--md-dur) var(--md-ease), box-shadow var(--md-dur) var(--md-ease), transform 160ms var(--md-spring);
  }
  .halt:hover {
    background: color-mix(in oklab, var(--md-halt) 18%, transparent);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-halt) 16%, transparent);
    transform: translateY(-1px);
  }
  .halt:active { transform: scale(0.96); }
  @keyframes md-tab-in {
    from { transform: scale(0.92); }
    to { transform: scale(1); }
  }
  @media (max-width: 720px) {
    .dock {
      left: 8px; right: 8px; bottom: 10px;
      padding: 8px 10px; gap: 6px; border-radius: 18px;
      justify-content: flex-start;
    }
    .tab { padding: 8px 11px; font-size: 12px; }
    .halt { padding: 7px 11px; font-size: 11px; margin-left: 4px; }
  }
  @media (max-width: 420px) {
    .dock {
      left: 6px; right: 6px; bottom: 8px;
      padding: 7px 8px;
    }
    .tab { padding: 8px 9px; font-size: 11.5px; }
    .divider { margin: 4px 0; }
  }
  @media (prefers-reduced-motion: reduce) {
    .tab.active, .badge { animation: none !important; }
  }
</style>
