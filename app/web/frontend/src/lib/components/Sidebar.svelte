<script lang="ts">
  import { onMount } from 'svelte'
  import { daemon } from '../stores/daemon.svelte'
  import { halt } from '../stores/halt.svelte'
  import { account } from '../stores/account.svelte'
  import Avatar from './ui/Avatar.svelte'
  import Badge from './ui/Badge.svelte'
  import IconButton from './ui/IconButton.svelte'

  // Expand state: drawer stays open on click of the chevron OR while
  // the user is hovering it. Collapsed by default so the rail (56px)
  // is the resting state.
  let hovered = $state(false)
  let pinned = $state(false)
  let expanded = $derived(pinned || hovered)
  let currentHash = $state('')

  // Version is hardcoded — the settings store does not expose a
  // `version` field today. Update this when the next release ships.
  const VERSION = 'v0.1.0'

  type NavItem = {
    href: string
    label: string
    icon: string // inline svg body
    match: (hash: string) => boolean
  }

  // Inline SVG bodies are written for a 16x16 viewBox with stroke 1.75.
  // Using `currentColor` so the rail's color cascade applies.
  const items: NavItem[] = [
    {
      href: '#/chat',
      label: 'Chat',
      match: (h) => h === '' || h === '#' || h === '#/' || h.startsWith('#/chat'),
      icon: 'M3 3h10a2 2 0 0 1 2 2v6a2 2 0 0 1-2 2H6l-3 2V5a2 2 0 0 1 2-2z'
    },
    {
      href: '#/tasks',
      label: 'Tasks',
      match: (h) => h.startsWith('#/tasks'),
      icon: 'M3 4.5h10M3 8h10M3 11.5h7M3 14.5L5 12.5L3 14.5z'
    },
    {
      href: '#/files',
      label: 'Files',
      match: (h) => h.startsWith('#/files'),
      icon: 'M3 3h4l1.5 2H13a1 1 0 0 1 1 1V13a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1z'
    },
    {
      href: '#/audit',
      label: 'Audit',
      match: (h) => h.startsWith('#/audit'),
      icon: 'M4 3h8a1 1 0 0 1 1 1v9a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1zM4.5 6h7M4.5 8.5h7M4.5 11h5'
    },
    {
      href: '#/replay',
      label: 'Replay',
      match: (h) => h.startsWith('#/replay'),
      icon: 'M8 4.5a3.5 3.5 0 1 1-3.5 3.5M3 8a5 5 0 1 0 1.5-3.5M3 4.5V6.5H5'
    },
    {
      href: '#/hub',
      label: 'Hub',
      match: (h) => h.startsWith('#/hub'),
      icon: 'M8 2.5l5 2.5v3.5c0 3-2 4.8-5 5.5-3-0.7-5-2.5-5-5.5V5l5-2.5z'
    },
    {
      href: '#/sync',
      label: 'Sync',
      match: (h) => h.startsWith('#/sync'),
      icon: 'M4 8a4 4 0 0 1 7-2.7M12 8a4 4 0 0 1-7 2.7M11 4.5h2v2M5 11.5H3v-2'
    },
    {
      href: '#/skills',
      label: 'Skills',
      match: (h) => h.startsWith('#/skills'),
      icon: 'M3 3l5 1.5L13 3l-1.5 9L8 13.5 4.5 12 3 3zM8 4.5v9'
    },
    {
      href: '#/channels',
      label: 'Channels',
      match: (h) => h.startsWith('#/channels'),
      icon: 'M3 4h10v7H7l-3 2V11H3V4z'
    },
    {
      href: '#/delegation',
      label: 'Delegation',
      match: (h) => h.startsWith('#/delegation'),
      icon: 'M5 5.5a1.5 1.5 0 1 1-1.5-1.5M11 5.5A1.5 1.5 0 1 1 9.5 4M8 9.5A1.5 1.5 0 1 1 6.5 11M5.5 6.5L7 8.5M9.5 6.5L8 8.5'
    },
    {
      href: '#/settings',
      label: 'Settings',
      match: (h) => h.startsWith('#/settings'),
      icon: 'M8 5.5a2.5 2.5 0 1 0 0 5 2.5 2.5 0 0 0 0-5zM8 1.5v2M8 12.5v2M1.5 8h2M12.5 8h2M3.4 3.4l1.4 1.4M11.2 11.2l1.4 1.4M3.4 12.6l1.4-1.4M11.2 4.8l1.4-1.4'
    },
    {
      href: '#/about',
      label: 'About',
      match: (h) => h.startsWith('#/about'),
      icon: 'M8 14a6 6 0 1 0 0-12 6 6 0 0 0 0 12zM8 7v4M8 5h.01'
    }
  ]

  function navigate(href: string): void {
    window.location.hash = href
  }

  function isActive(item: NavItem, hash: string): boolean {
    return item.match(hash)
  }

  function togglePin(): void {
    pinned = !pinned
  }

  // Initials for the avatar fallback when no OAuth picture is present.
  const accountInitials = $derived.by(() => {
    const seed = account.displayName || account.email || '?'
    return seed.split(/\s+|@/).filter(Boolean).slice(0, 2).map((p) => p[0]?.toUpperCase() ?? '').join('') || '?'
  })

  onMount(() => {
    currentHash = window.location.hash || '#/'
    const onHash = () => {
      currentHash = window.location.hash || '#/'
    }
    window.addEventListener('hashchange', onHash)
    void account.checkStatus()
    return () => {
      window.removeEventListener('hashchange', onHash)
    }
  })
</script>

<aside
  class="sidebar"
  class:is-expanded={expanded}
  class:is-halted={halt.state.halted}
  onmouseenter={() => (hovered = true)}
  onmouseleave={() => (hovered = false)}
  role="navigation"
  aria-label="Primary navigation"
>
  <!-- Top: logo + wordmark -->
  <header class="brand">
    <span class="brand-dot" aria-hidden="true"></span>
    {#if expanded}
      <span class="brand-name anim-fade">Condura</span>
    {/if}
  </header>

  {#if halt.state.halted}
    <div class="halt-pill anim-fade-in-up" aria-live="polite">
      <Badge tone="error" size="xs" dot pulse>HALTED</Badge>
    </div>
  {/if}

  <!-- Nav items -->
  <nav class="nav" aria-label="Sections">
    {#each items as item (item.href)}
      {@const active = isActive(item, currentHash)}
      <button
        type="button"
        class="nav-item"
        class:is-active={active}
        title={item.label}
        aria-label={item.label}
        aria-current={active ? 'page' : undefined}
        onclick={() => navigate(item.href)}
      >
        <span class="nav-active-bar" aria-hidden="true"></span>
        <span class="nav-icon" aria-hidden="true">
          <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
            <path d={item.icon} />
          </svg>
        </span>
        {#if expanded}
          <span class="nav-label anim-fade">{item.label}</span>
        {/if}
      </button>
    {/each}
  </nav>

  <!-- Spacer pushes the footer to the bottom -->
  <div class="spacer"></div>

  <!-- Footer: dev surface, account chip, daemon + version -->
  <footer class="footer">
    <button
      type="button"
      class="nav-item dev-link"
      title="Dev components"
      aria-label="Dev components"
      onclick={() => navigate('#/dev/components')}
    >
      <span class="nav-active-bar" aria-hidden="true"></span>
      <span class="nav-icon" aria-hidden="true">
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
          <path d="M4 4l4-2 4 2-4 2-4-2zM2 9l4-2M14 9l-4-2M2 9v2l4 2M14 9v2l-4 2M8 6v7" />
        </svg>
      </span>
      {#if expanded}
        <span class="nav-label anim-fade">Dev surface</span>
        <span class="nav-badge anim-fade">
          <Badge tone="warn" size="xs">DEV</Badge>
        </span>
      {/if}
    </button>

    <!-- Daemon connection indicator -->
    <div class="daemon-row" aria-label="Daemon connection">
      <span
        class="daemon-dot"
        class:is-on={daemon.connected}
        class:anim-glow-pulse={daemon.connected}
        aria-hidden="true"
      ></span>
      {#if expanded}
        <span class="daemon-label anim-fade">
          {daemon.connected ? 'Connected' : 'Disconnected'}
        </span>
      {/if}
    </div>

    <!-- Account chip / sign-in -->
    {#if account.isSignedIn}
      <button
        type="button"
        class="account-chip"
        title={account.email || account.displayName}
        aria-label="Account menu"
        onclick={() => navigate('#/settings')}
      >
        <Avatar name={account.displayName || account.email || '?'} src={account.avatarURL} size="sm" status={daemon.connected ? 'online' : 'offline'} />
        {#if expanded}
          <span class="account-info anim-fade">
            <span class="account-name">{account.displayName || 'Account'}</span>
            <span class="account-email">{account.email}</span>
          </span>
        {/if}
      </button>
    {:else}
      <button
        type="button"
        class="signin-link"
        title="Sign in"
        aria-label="Sign in"
        onclick={() => navigate('#/settings')}
      >
        <Avatar name={accountInitials} size="sm" />
        {#if expanded}
          <span class="signin-text anim-fade">Sign in</span>
        {/if}
      </button>
    {/if}

    <!-- Version label + chevron toggle -->
    <div class="version-row">
      {#if expanded}
        <span class="version anim-fade" aria-label="Version">{VERSION}</span>
      {/if}
      <IconButton
        variant="ghost"
        size="sm"
        ariaLabel={pinned ? 'Collapse sidebar' : 'Pin sidebar open'}
        title={pinned ? 'Collapse sidebar' : 'Pin sidebar open'}
        onclick={togglePin}
      >
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" width="14" height="14">
          {#if pinned}
            <path d="M10 4l-4 4 4 4" />
          {:else}
            <path d="M6 4l4 4-4 4" />
          {/if}
        </svg>
      </IconButton>
    </div>
  </footer>
</aside>

<style>
  /* ── Shell ─────────────────────────────────────────────────── */
  .sidebar {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    width: var(--sidebar-rail-width);
    background: var(--surface-1);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    padding: var(--space-3) var(--space-2);
    gap: var(--space-2);
    overflow: hidden;
    z-index: var(--z-sticky);
    transition: width var(--transition-slow) var(--ease-out-quart);
  }
  .sidebar.is-expanded {
    width: var(--sidebar-drawer-width);
  }

  /* ── Brand ────────────────────────────────────────────────── */
  .brand {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    height: 28px;
    padding: 0 var(--space-2);
    flex-shrink: 0;
  }
  .brand-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--accent-gradient);
    box-shadow: 0 0 12px var(--accent-glow);
    flex-shrink: 0;
  }
  .brand-name {
    font-family: var(--font-sans);
    font-weight: var(--weight-semibold);
    font-size: var(--size-sm);
    letter-spacing: var(--tracking-tight);
    color: var(--text);
    white-space: nowrap;
  }

  /* ── Halt pill ─────────────────────────────────────────────── */
  .halt-pill {
    display: flex;
    justify-content: center;
    padding: 0 var(--space-1);
  }

  /* ── Nav list ──────────────────────────────────────────────── */
  .nav {
    display: flex;
    flex-direction: column;
    gap: 2px;
    margin-top: var(--space-1);
  }

  .nav-item {
    position: relative;
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    height: 36px;
    padding: 0 var(--space-2);
    background: transparent;
    border: 1px solid transparent;
    border-radius: var(--radius-md);
    color: var(--text-muted);
    font-family: var(--font-sans);
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    cursor: pointer;
    text-align: left;
    transition:
      background-color var(--transition-fast),
      border-color var(--transition-fast),
      color var(--transition-fast);
  }
  .nav-item:hover {
    background: var(--surface-2);
    color: var(--text);
  }

  .nav-active-bar {
    position: absolute;
    left: -2px;
    top: 50%;
    transform: translateY(-50%) scaleY(0);
    width: 2px;
    height: 18px;
    border-radius: 0 2px 2px 0;
    background: var(--accent);
    box-shadow: 0 0 8px var(--accent-glow);
    transition: transform var(--transition-base) var(--ease-spring);
  }
  .nav-item.is-active .nav-active-bar {
    transform: translateY(-50%) scaleY(1);
  }

  .nav-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    flex-shrink: 0;
  }
  .nav-icon svg {
    width: 16px;
    height: 16px;
  }

  .nav-label {
    flex: 1;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .nav-badge {
    flex-shrink: 0;
  }

  .nav-item.is-active {
    background: var(--accent-soft);
    color: var(--accent);
  }
  .nav-item.is-active:hover {
    background: var(--accent-soft);
    color: var(--accent-hover);
  }

  /* ── Spacer ────────────────────────────────────────────────── */
  .spacer {
    flex: 1;
    min-height: var(--space-2);
  }

  /* ── Footer stack ──────────────────────────────────────────── */
  .footer {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding-top: var(--space-2);
    border-top: 1px solid var(--border);
  }

  /* ── Daemon indicator ──────────────────────────────────────── */
  .daemon-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 0 var(--space-2);
    height: 22px;
  }
  .daemon-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--text-faint);
    flex-shrink: 0;
    transition: background-color var(--transition-base);
  }
  .daemon-dot.is-on {
    background: var(--success);
  }
  .daemon-label {
    font-family: var(--font-mono);
    font-size: var(--size-2xs);
    letter-spacing: var(--tracking-wider);
    text-transform: uppercase;
    color: var(--text-muted);
    white-space: nowrap;
  }

  /* ── Account chip / sign-in ────────────────────────────────── */
  .account-chip,
  .signin-link {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    width: 100%;
    height: 36px;
    padding: 0 var(--space-2);
    background: transparent;
    border: 1px solid transparent;
    border-radius: var(--radius-md);
    color: var(--text);
    font-family: var(--font-sans);
    font-size: var(--size-sm);
    cursor: pointer;
    text-align: left;
    transition:
      background-color var(--transition-fast),
      border-color var(--transition-fast);
  }
  .account-chip:hover,
  .signin-link:hover {
    background: var(--surface-2);
    border-color: var(--border-strong);
  }

  .account-info {
    display: flex;
    flex-direction: column;
    min-width: 0;
    flex: 1;
  }
  .account-name {
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    color: var(--text);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    line-height: var(--leading-tight);
  }
  .account-email {
    font-size: var(--size-2xs);
    font-family: var(--font-mono);
    color: var(--text-faint);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    line-height: var(--leading-tight);
  }

  .signin-text {
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    color: var(--text);
  }

  /* ── Version + toggle row ──────────────────────────────────── */
  .version-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 var(--space-2);
    height: 26px;
  }
  .version {
    font-family: var(--font-mono);
    font-size: var(--size-2xs);
    color: var(--text-faint);
    letter-spacing: var(--tracking-wide);
  }

  /* ── Narrow-window safety ──────────────────────────────────── */
  @media (max-width: 280px) {
    /* keep the rail at its 56px minimum so it never collapses the
       nav into nothing — but the surrounding app shell is responsible
       for hiding the rest of the layout below this width */
    .sidebar {
      min-width: var(--sidebar-rail-width);
    }
  }
</style>