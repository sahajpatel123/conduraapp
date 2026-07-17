<script lang="ts">
  import type { RouteId } from './routes'
  import { halt } from '../../stores/halt.svelte'
  import { pendingCount } from '../../stores/pending.svelte'

  interface Props {
    route: RouteId
    onnavigate: (r: RouteId) => void
    /** Called before halt.halt() so the shell can capture the
     *  trigger element for focus restoration on close. */
    onbeforehalt?: () => void
  }
  let { route, onnavigate, onbeforehalt }: Props = $props()

  const PRIMARY: { id: RouteId; label: string; icon: string; shortcut: string }[] = [
    { id: 'chat', label: 'Ask', icon: 'ask', shortcut: '⌘1' },
    { id: 'hub', label: 'Hub', icon: 'hub', shortcut: '⌘2' },
    { id: 'skills', label: 'Skills', icon: 'skills', shortcut: '⌘3' },
    { id: 'sync', label: 'Sync', icon: 'sync', shortcut: '⌘4' },
    { id: 'audit', label: 'Audit', icon: 'audit', shortcut: '⌘5' },
    { id: 'replay', label: 'Replay', icon: 'replay', shortcut: '⌘6' },
    { id: 'channels', label: 'Channels', icon: 'channels', shortcut: '⌘7' },
    { id: 'delegation', label: 'Agents', icon: 'agents', shortcut: '⌘8' },
  ]
  const MORE: { id: RouteId; label: string }[] = [
    { id: 'account', label: 'Account' },
    { id: 'settings', label: 'Settings' },
    { id: 'about', label: 'About' },
  ]
  const pending = $derived($pendingCount)
  /** Flat navigation order — primary first, then more. Matches the visual order. */
  const NAV_IDS: RouteId[] = [...PRIMARY.map((p) => p.id), ...MORE.map((m) => m.id)]
  const activeIdx = $derived(Math.max(0, NAV_IDS.indexOf(route)))

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

  /**
   * Toolbar roving-tabindex keyboard nav (W3C APG pattern).
   * Arrow keys move both focus AND selection so a screen-reader user can
   * hear the new route as soon as they press the key — mirrors macOS Dock.
   * Halt button is a separate action and stays out of the roving order.
   */
  function onKey(e: KeyboardEvent): void {
    let next = activeIdx
    if (e.key === 'ArrowRight') next = (activeIdx + 1) % NAV_IDS.length
    else if (e.key === 'ArrowLeft') next = (activeIdx - 1 + NAV_IDS.length) % NAV_IDS.length
    else if (e.key === 'Home') next = 0
    else if (e.key === 'End') next = NAV_IDS.length - 1
    else return
    e.preventDefault()
    const target = NAV_IDS[next]
    if (!target) return
    onnavigate(target)
    queueMicrotask(() => {
      const btn = dockEl?.querySelector<HTMLElement>(`[data-tab-id="${target}"]`)
      btn?.focus({ preventScroll: true })
    })
  }
</script>

<!-- onkeydown is delegated: keys fire on inner <button> tabs (which are
     the real interactive elements). The <nav> is the bubble target so
     arrow-key navigation between tabs has one listener, not N. -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<nav class="dock" aria-label="Primary" bind:this={dockEl} onkeydown={onKey}>
  <div class="primary">
    {#each PRIMARY as item (item.id)}
      <button
        type="button"
        class="tab"
        class:active={route === item.id}
        tabindex={route === item.id ? 0 : -1}
        data-tab-id={item.id}
        aria-current={route === item.id ? 'page' : undefined}
        onclick={() => onnavigate(item.id)}
      >
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
        <span class="dock-kbd" aria-hidden="true">{item.shortcut}</span>
        {#if item.id === 'delegation' && pending > 0}<span class="badge">{pending}</span>{/if}
      </button>
    {/each}
  </div>
  <div class="divider" aria-hidden="true"></div>
  <div class="more">
    {#each MORE as item (item.id)}
      <button
        type="button"
        class="tab quiet"
        class:active={route === item.id}
        tabindex={route === item.id ? 0 : -1}
        data-tab-id={item.id}
        aria-current={route === item.id ? 'page' : undefined}
        onclick={() => onnavigate(item.id)}
      >
        {item.label}
      </button>
    {/each}
    <button type="button" class="halt" tabindex="-1" onclick={() => { onbeforehalt?.(); void halt.halt() }} aria-label="Halt agent">Halt</button>
  </div>
</nav>

<style>
  .dock {
    position: absolute;
    left: 20px;
    right: 20px;
    bottom: 12px;
    z-index: 20;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    gap: 4px;
    padding: 5px 6px;
    border-radius: 12px;
    background: color-mix(in oklab, var(--md-surface) 94%, transparent);
    border: 1px solid var(--md-line);
    backdrop-filter: blur(12px) saturate(1.05);
    -webkit-backdrop-filter: blur(12px) saturate(1.05);
    box-shadow: none;
    animation: md-dock-up 420ms var(--md-ease) 60ms both;
    overflow-x: auto;
    scrollbar-width: none;
    -webkit-overflow-scrolling: touch;
    scroll-padding-inline: 8px;
  }
  :root[data-mode='dark'] .dock {
    box-shadow: none;
  }
  .dock::-webkit-scrollbar { display: none; }
  .primary, .more { display: flex; align-items: center; gap: 1px; flex: none; }
  .divider {
    width: 1px; align-self: stretch; margin: 7px 5px;
    background: var(--md-line);
    flex: none;
  }
  .dock:hover {
    border-color: var(--md-line-strong);
  }
  .tab {
    position: relative;
    appearance: none; border: 0; background: transparent; color: var(--md-ink-mute);
    font-family: var(--md-font-sans); font-size: 12px; font-weight: 500;
    padding: 7px 10px; border-radius: 8px; cursor: pointer; white-space: nowrap;
    display: inline-flex; align-items: center; gap: 5px;
    transition: color 120ms var(--md-ease), background 120ms var(--md-ease);
  }
  .tab .ico {
    display: inline-flex; opacity: 0.65;
    transition: opacity 120ms var(--md-ease);
  }
  .tab.quiet { font-weight: 450; color: var(--md-ink-faint); }
  .tab:hover {
    color: var(--md-ink);
    background: color-mix(in oklab, var(--md-ink) 3.5%, transparent);
  }
  .tab:hover .ico { opacity: 0.95; }
  .tab.active {
    color: #fff;
    background: var(--md-cobalt);
    box-shadow: none;
    animation: none;
    font-weight: 550;
  }
  .tab.active .ico { opacity: 1; }
  .tab:active { opacity: 0.88; }
  .tab:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
    color: var(--md-ink);
  }
  .tab.active:focus-visible {
    box-shadow: var(--md-focus);
  }
  .halt:focus-visible {
    outline: none;
    box-shadow: 0 0 0 2px color-mix(in oklab, var(--md-halt) 26%, transparent);
  }
  .badge {
    display: inline-flex; align-items: center; justify-content: center;
    min-width: 14px; height: 14px; margin-left: 1px; padding: 0 4px; border-radius: 999px;
    background: color-mix(in oklab, var(--md-cobalt) 10%, #fff); color: var(--md-cobalt);
    font-size: 9px; font-weight: 700;
  }
  :root[data-mode='dark'] .badge {
    background: color-mix(in oklab, var(--md-cobalt) 16%, var(--md-surface)); color: var(--md-cobalt);
  }
  .dock-kbd {
    font-family: var(--md-font-mono);
    font-size: 9px;
    font-weight: 500;
    letter-spacing: 0.02em;
    padding: 0 4px;
    border-radius: 4px;
    background: color-mix(in oklab, var(--md-stage) 60%, transparent);
    border: 1px solid color-mix(in oklab, var(--md-line) 70%, transparent);
    color: var(--md-ink-faint);
    margin-left: 2px;
    transition: color 140ms var(--md-ease), border-color 140ms var(--md-ease);
    /* Always hidden by default — only shows on hover/active to keep the
       dock calm. Power users can also see it persistently if their
       platform strips hover (touch). */
    opacity: 0;
  }
  .tab:hover .dock-kbd,
  .tab.active .dock-kbd {
    opacity: 1;
  }
  .tab.active .dock-kbd {
    color: rgba(255, 255, 255, 0.85);
    border-color: rgba(255, 255, 255, 0.32);
    background: transparent;
  }
  /* Touch / no-hover: keep the badge visible so the discoverability
     promise holds without needing hover. */
  @media (hover: none) {
    .dock-kbd { opacity: 1; }
  }
  @media (max-width: 720px) {
    /* Tight dock on small screens — drop the badge to save space. */
    .dock-kbd { display: none; }
  }
  .halt {
    appearance: none; margin-left: 2px;
    border: 1px solid color-mix(in oklab, var(--md-halt) 28%, var(--md-line-strong));
    background: transparent;
    color: var(--md-halt); font-size: 10.5px; font-weight: 650; letter-spacing: 0.06em;
    text-transform: uppercase; padding: 6px 10px; border-radius: 8px; cursor: pointer;
    transition: background 120ms var(--md-ease), border-color 120ms var(--md-ease);
  }
  .halt:hover {
    background: color-mix(in oklab, var(--md-halt) 6%, transparent);
    border-color: color-mix(in oklab, var(--md-halt) 44%, var(--md-line-strong));
  }
  .halt:active { opacity: 0.88; }
  @keyframes md-tab-in {
    from { transform: scale(0.92); }
    to { transform: scale(1); }
  }
  @media (max-width: 720px) {
    .dock {
      left: 14px; right: 14px; bottom: 10px;
      padding: 5px 6px; gap: 4px; border-radius: 12px;
      justify-content: flex-start;
    }
    .tab { padding: 7px 9px; font-size: 12px; gap: 5px; }
    .tab .ico { display: none; }
    .halt { padding: 6px 10px; font-size: 10.5px; margin-left: 2px; }
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
