<script lang="ts">
  import { onMount } from 'svelte'
  import { daemon } from '../stores/daemon.svelte'
  import { halt } from '../stores/halt.svelte'
  import { account } from '../stores/account.svelte'
  import IrisOrb from './IrisOrb.svelte'
  import Avatar from './ui/Avatar.svelte'

  // The floating Dock — IRIS navigation. A detached vertical glass
  // capsule that hovers over the Aurora Field (not flush to the
  // edge), so the living light wraps every side. The Iris orb sits
  // at the top as the agent's face; primary destinations follow;
  // Settings + account pin to the bottom.
  interface Props {
    orbState?: 'idle' | 'listening' | 'thinking' | 'acting' | 'consent' | 'offline'
  }
  let { orbState = 'idle' }: Props = $props()

  let currentHash = $state('#/')

  type NavItem = { href: string; label: string; icon: string; match: (h: string) => boolean }

  const primary: NavItem[] = [
    { href: '#/', label: 'Chat', match: (h) => h === '' || h === '#' || h === '#/' || h.startsWith('#/chat'),
      icon: 'M3 4.5a2 2 0 0 1 2-2h6a2 2 0 0 1 2 2V10a2 2 0 0 1-2 2H6.5L3.5 14.2a.4.4 0 0 1-.5-.4V4.5z' },
    { href: '#/audit', label: 'Audit', match: (h) => h.startsWith('#/audit'),
      icon: 'M4 2.5h6.2L13 5.3V13a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V3.5a1 1 0 0 1 1-1zM5.5 7.5h5M5.5 10h3.5' },
    { href: '#/replay', label: 'Replay', match: (h) => h.startsWith('#/replay'),
      icon: 'M8 3.2A4.8 4.8 0 1 1 3.4 6.6M3.2 3v3h3M8 5.6V8l1.8 1.1' },
    { href: '#/skills', label: 'Skills', match: (h) => h.startsWith('#/skills'),
      icon: 'M8 1.8l1.8 3.7 4.1.6-3 2.9.7 4-3.6-1.9-3.6 1.9.7-4-3-2.9 4.1-.6L8 1.8z' },
    { href: '#/delegation', label: 'Delegation', match: (h) => h.startsWith('#/delegation'),
      icon: 'M4.5 4.5a1.6 1.6 0 1 1-1.6-1.6M11.5 4.5A1.6 1.6 0 1 1 9.9 2.9M8 13.1a1.6 1.6 0 1 1-1.6-1.6M4.4 5.6 7 9.2M11.6 5.6 9 9.2' },
    { href: '#/channels', label: 'Channels', match: (h) => h.startsWith('#/channels'),
      icon: 'M3 4.2a1.5 1.5 0 0 1 1.5-1.5h7A1.5 1.5 0 0 1 13 4.2v5a1.5 1.5 0 0 1-1.5 1.5H7l-3 2.3V10.7a1.5 1.5 0 0 1-1-1.5v-5z' },
    { href: '#/sync', label: 'Sync', match: (h) => h.startsWith('#/sync'),
      icon: 'M4 8a4 4 0 0 1 6.9-2.8M12 8a4 4 0 0 1-6.9 2.8M11 4.2h1.8V6M5 11.8H3.2V10' },
    { href: '#/hub', label: 'Hub', match: (h) => h.startsWith('#/hub'),
      icon: 'M8 2.2l5 2.5v3.6c0 3-2.1 4.9-5 5.7-2.9-.8-5-2.7-5-5.7V4.7l5-2.5zM8 6v4M6 8h4' }
  ]

  const lower: NavItem[] = [
    { href: '#/settings', label: 'Settings', match: (h) => h.startsWith('#/settings'),
      icon: 'M8 5.6a2.4 2.4 0 1 0 0 4.8 2.4 2.4 0 0 0 0-4.8zM8 1.6v1.8M8 12.6v1.8M2.1 8h1.8M12.1 8h1.8M3.6 3.6l1.3 1.3M11.1 11.1l1.3 1.3M3.6 12.4l1.3-1.3M11.1 4.9l1.3-1.3' },
    { href: '#/about', label: 'About', match: (h) => h.startsWith('#/about'),
      icon: 'M8 14a6 6 0 1 0 0-12 6 6 0 0 0 0 12zM8 7.2v3.6M8 5.1h.01' }
  ]

  function navigate(href: string): void { window.location.hash = href }

  const initials = $derived.by(() => {
    const seed = account.displayName || account.email || 'C'
    return seed.split(/\s+|@/).filter(Boolean).slice(0, 2).map((p) => p[0]?.toUpperCase() ?? '').join('') || 'C'
  })

  onMount(() => {
    currentHash = window.location.hash || '#/'
    const onHash = () => { currentHash = window.location.hash || '#/' }
    window.addEventListener('hashchange', onHash)
    void account.checkStatus()
    return () => window.removeEventListener('hashchange', onHash)
  })
</script>

<nav class="dock" class:halted={halt.state.halted} aria-label="Primary navigation">
  <button class="orb-btn" title="Agent status" aria-label="Agent status" onclick={() => navigate('#/')}>
    <IrisOrb state={orbState} size={30} />
  </button>

  <span class="rule" aria-hidden="true"></span>

  <div class="group">
    {#each primary as item (item.href)}
      {@const active = item.match(currentHash)}
      <button
        class="item"
        class:active
        aria-label={item.label}
        aria-current={active ? 'page' : undefined}
        onclick={() => navigate(item.href)}
      >
        <span class="bar" aria-hidden="true"></span>
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d={item.icon} />
        </svg>
        <span class="tip">{item.label}</span>
      </button>
    {/each}
  </div>

  <span class="spacer"></span>

  <div class="group">
    {#each lower as item (item.href)}
      {@const active = item.match(currentHash)}
      <button
        class="item"
        class:active
        aria-label={item.label}
        aria-current={active ? 'page' : undefined}
        onclick={() => navigate(item.href)}
      >
        <span class="bar" aria-hidden="true"></span>
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d={item.icon} />
        </svg>
        <span class="tip">{item.label}</span>
      </button>
    {/each}

    <button class="account" title={account.email || 'Account'} aria-label="Account" onclick={() => navigate('#/settings')}>
      <Avatar name={account.displayName || account.email || initials} src={account.avatarURL} size="sm" status={daemon.connected ? 'online' : 'offline'} />
      <span class="tip">{account.isSignedIn ? (account.displayName || 'Account') : 'Sign in'}</span>
    </button>
  </div>
</nav>

<style>
  .dock {
    position: fixed;
    left: var(--app-pad);
    top: 50%;
    transform: translateY(-50%);
    z-index: var(--z-sticky);
    width: var(--dock-width);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    padding: 12px 0;
    border-radius: var(--dock-radius);
    background: var(--glass-bg);
    backdrop-filter: var(--blur-base);
    -webkit-backdrop-filter: var(--blur-base);
    border: 1px solid var(--border);
    box-shadow: var(--shadow-float);
    max-height: calc(100vh - var(--app-pad) * 2);
  }
  .dock.halted { box-shadow: var(--shadow-float), 0 0 0 1px var(--border-warm), var(--glow-coral); }

  .orb-btn {
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    border-radius: var(--radius-lg);
    background: transparent;
  }

  .rule {
    width: 22px;
    height: 1px;
    background: var(--border-strong);
    margin: 2px 0 6px;
  }

  .group { display: flex; flex-direction: column; align-items: center; gap: 6px; }
  .spacer { flex: 1; min-height: 8px; }

  .item {
    position: relative;
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    border-radius: var(--radius-md);
    color: var(--text-muted);
    background: transparent;
    transition: transform var(--transition-base) var(--ease-spring),
                background var(--transition-fast), color var(--transition-fast);
  }
  .item svg { width: 18px; height: 18px; }
  .item:hover { background: var(--surface-2); color: var(--text); transform: scale(1.12); }
  .item:active { transform: scale(0.96); }
  .item.active { background: var(--accent-soft); color: var(--accent); }

  .bar {
    position: absolute;
    left: -7px;
    top: 50%;
    transform: translateY(-50%) scaleY(0);
    width: 3px;
    height: 18px;
    border-radius: 0 3px 3px 0;
    background: var(--accent);
    box-shadow: var(--glow-iris);
    transition: transform var(--transition-base) var(--ease-spring);
  }
  .item.active .bar { transform: translateY(-50%) scaleY(1); }

  .account {
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    border-radius: var(--radius-pill);
    background: transparent;
    position: relative;
    margin-top: 2px;
    transition: transform var(--transition-base) var(--ease-spring);
  }
  .account:hover { transform: scale(1.1); }

  /* right-side tooltip */
  .tip {
    position: absolute;
    left: calc(100% + 12px);
    top: 50%;
    transform: translateY(-50%) translateX(-4px);
    white-space: nowrap;
    padding: 6px 10px;
    border-radius: var(--radius-sm);
    background: var(--glass-bg-solid);
    backdrop-filter: var(--blur-thin);
    -webkit-backdrop-filter: var(--blur-thin);
    border: 1px solid var(--border);
    box-shadow: var(--shadow-md);
    color: var(--text);
    font-size: var(--size-xs);
    font-weight: var(--weight-medium);
    opacity: 0;
    pointer-events: none;
    transition: opacity var(--transition-fast), transform var(--transition-fast);
    z-index: var(--z-overlay);
  }
  .item:hover .tip,
  .account:hover .tip { opacity: 1; transform: translateY(-50%) translateX(0); }
</style>
