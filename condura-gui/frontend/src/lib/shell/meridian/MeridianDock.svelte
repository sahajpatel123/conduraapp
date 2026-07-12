<script lang="ts">
  import type { RouteId } from './routes'
  import { halt } from '../../stores/halt.svelte'
  import { pendingCount } from '../../stores/pending.svelte'

  interface Props {
    route: RouteId
    onnavigate: (r: RouteId) => void
  }
  let { route, onnavigate }: Props = $props()

  const PRIMARY: { id: RouteId; label: string; icon: string }[] = [
    { id: 'chat', label: 'Ask', icon: 'ask' },
    { id: 'hub', label: 'Hub', icon: 'hub' },
    { id: 'skills', label: 'Skills', icon: 'skills' },
    { id: 'sync', label: 'Sync', icon: 'sync' },
    { id: 'audit', label: 'Audit', icon: 'audit' },
    { id: 'replay', label: 'Replay', icon: 'replay' },
    { id: 'channels', label: 'Channels', icon: 'channels' },
    { id: 'delegation', label: 'Agents', icon: 'agents' },
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
        <span class="ico" aria-hidden="true" data-i={item.icon}>
          {#if item.icon === 'ask'}
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M4 6.5h16M4 12h10M4 17.5h13" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
          {:else if item.icon === 'hub'}
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="2"/><path d="M12 3v3M12 18v3M3 12h3M18 12h3M5.6 5.6l2.1 2.1M16.3 16.3l2.1 2.1M18.4 5.6l-2.1 2.1M7.7 16.3l-2.1 2.1" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
          {:else if item.icon === 'skills'}
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M8 4h8l2 4-6 12L6 8l2-4z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/></svg>
          {:else if item.icon === 'sync'}
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M4 12a8 8 0 0114.5-4.5M20 7v4h-4M20 12a8 8 0 01-14.5 4.5M4 17v-4h4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
          {:else if item.icon === 'audit'}
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M8 4h8v16H8z" stroke="currentColor" stroke-width="1.8"/><path d="M10.5 9h5M10.5 13h5M10.5 17h3" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/></svg>
          {:else if item.icon === 'replay'}
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><rect x="3" y="6" width="18" height="12" rx="2" stroke="currentColor" stroke-width="1.8"/><path d="M10 9.5l5 2.5-5 2.5v-5z" fill="currentColor"/></svg>
          {:else if item.icon === 'channels'}
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><path d="M4 8l8-4 8 4v8l-8 4-8-4V8z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/><path d="M12 12v8M4 8l8 4 8-4" stroke="currentColor" stroke-width="1.6"/></svg>
          {:else}
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none"><circle cx="8" cy="10" r="2.5" stroke="currentColor" stroke-width="1.7"/><circle cx="16" cy="10" r="2.5" stroke="currentColor" stroke-width="1.7"/><path d="M4 18c.8-2.2 2.6-3.5 4-3.5s3.2 1.3 4 3.5M12 18c.8-2.2 2.6-3.5 4-3.5s3.2 1.3 4 3.5" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/></svg>
          {/if}
        </span>
        <span class="label">{item.label}</span>
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
    position: absolute;
    left: 28px;
    right: 28px;
    bottom: 16px;
    z-index: 20;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    gap: 8px;
    padding: 9px 11px;
    border-radius: 20px;
    background: color-mix(in oklab, var(--md-surface) 82%, transparent);
    border: 1px solid var(--md-line-strong);
    backdrop-filter: blur(22px) saturate(1.25);
    -webkit-backdrop-filter: blur(22px) saturate(1.25);
    box-shadow:
      0 1px 0 color-mix(in oklab, #fff 50%, transparent) inset,
      0 18px 32px -22px color-mix(in oklab, var(--md-ink) 32%, transparent);
    animation: md-dock-up 640ms var(--md-ease) 120ms both;
    overflow-x: auto;
    scrollbar-width: none;
    -webkit-overflow-scrolling: touch;
    scroll-padding-inline: 8px;
  }
  :root[data-mode='dark'] .dock {
    box-shadow:
      0 1px 0 color-mix(in oklab, #fff 8%, transparent) inset,
      0 18px 32px -20px rgba(0, 0, 0, 0.55);
  }
  .dock::-webkit-scrollbar { display: none; }
  .primary, .more { display: flex; align-items: center; gap: 3px; flex: none; }
  .divider {
    width: 1px; align-self: stretch; margin: 4px 4px;
    background: linear-gradient(180deg, transparent, var(--md-line-strong) 20%, var(--md-line-strong) 80%, transparent);
    flex: none;
  }
  .dock:hover {
    box-shadow:
      0 1px 0 color-mix(in oklab, #fff 50%, transparent) inset,
      0 22px 38px -22px color-mix(in oklab, var(--md-ink) 38%, transparent);
    border-color: color-mix(in oklab, var(--md-cobalt) 26%, var(--md-line-strong));
  }
  .tab {
    position: relative;
    appearance: none; border: 0; background: transparent; color: var(--md-ink-mute);
    font-family: var(--md-font-sans); font-size: 12.5px; font-weight: 600;
    padding: 9px 12px; border-radius: 999px; cursor: pointer; white-space: nowrap;
    display: inline-flex; align-items: center; gap: 6px;
    transition: color var(--md-dur) var(--md-ease), background var(--md-dur) var(--md-ease), transform 160ms var(--md-spring), box-shadow var(--md-dur) var(--md-ease);
  }
  .tab .ico {
    display: inline-flex; opacity: 0.72;
    transition: opacity var(--md-dur) var(--md-ease), transform 160ms var(--md-spring);
  }
  .tab.quiet { font-weight: 500; color: var(--md-ink-faint); }
  .tab:hover {
    color: var(--md-ink);
    background: color-mix(in oklab, var(--md-cobalt) 8%, transparent);
    transform: translateY(-1px);
  }
  .tab:hover .ico { opacity: 1; }
  .tab.active {
    color: #fff;
    background: linear-gradient(165deg, color-mix(in oklab, var(--md-cobalt) 90%, #fff), var(--md-cobalt) 55%, var(--md-cobalt-deep));
    box-shadow:
      0 1px 0 color-mix(in oklab, #fff 25%, transparent) inset,
      0 8px 20px -10px color-mix(in oklab, var(--md-cobalt) 72%, transparent);
    animation: md-tab-in 280ms var(--md-spring) both;
  }
  .tab.active .ico { opacity: 1; }
  .tab:active { transform: scale(0.96); }
  .tab:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
    color: var(--md-ink);
  }
  .tab.active:focus-visible {
    box-shadow:
      0 1px 0 color-mix(in oklab, #fff 25%, transparent) inset,
      0 8px 20px -10px color-mix(in oklab, var(--md-cobalt) 72%, transparent),
      var(--md-focus);
  }
  .halt:focus-visible {
    outline: none;
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-halt) 22%, transparent);
  }
  .badge {
    display: inline-flex; align-items: center; justify-content: center;
    min-width: 16px; height: 16px; margin-left: 2px; padding: 0 4px; border-radius: 999px;
    background: #fff; color: var(--md-cobalt); font-size: 10px; font-weight: 800;
    animation: md-rise 280ms var(--md-spring) both;
  }
  :root[data-mode='dark'] .badge {
    background: var(--md-surface); color: var(--md-cobalt);
  }
  .halt {
    appearance: none; margin-left: 6px;
    border: 1px solid color-mix(in oklab, var(--md-halt) 42%, transparent);
    background: color-mix(in oklab, var(--md-halt) 12%, transparent);
    color: var(--md-halt); font-size: 11px; font-weight: 800; letter-spacing: 0.06em;
    text-transform: uppercase; padding: 8px 14px; border-radius: 999px; cursor: pointer;
    transition: background var(--md-dur) var(--md-ease), box-shadow var(--md-dur) var(--md-ease), transform 160ms var(--md-spring);
  }
  .halt:hover {
    background: color-mix(in oklab, var(--md-halt) 20%, transparent);
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
      left: 14px; right: 14px; bottom: 10px;
      padding: 8px 9px; gap: 6px; border-radius: 18px;
      justify-content: flex-start;
    }
    .tab { padding: 8px 10px; font-size: 12px; gap: 5px; }
    .tab .ico { display: none; }
    .halt { padding: 7px 11px; font-size: 11px; margin-left: 4px; }
  }
  @media (max-width: 420px) {
    .dock {
      left: 10px; right: 10px; bottom: 8px;
      padding: 7px 8px;
    }
    .tab { padding: 8px 9px; font-size: 11.5px; }
    .divider { margin: 4px 2px; }
  }
  @media (prefers-reduced-motion: reduce) {
    .tab.active, .badge { animation: none !important; }
  }
</style>
