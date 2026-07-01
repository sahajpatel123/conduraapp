<!--
  Sidebar — compact left navigation rail (v1 design system).
-->
<script lang="ts">
  import Pulse from './Pulse.svelte';
  import Hairline from './Hairline.svelte';

  export type RouteId =
    | 'chat'
    | 'audit'
    | 'replay'
    | 'hub'
    | 'sync'
    | 'skills'
    | 'channels'
    | 'delegation'
    | 'settings'
    | 'about';

  interface Props {
    active?: RouteId;
    collapsed?: boolean;
    onnavigate?: (route: RouteId) => void;
    ontoggle?: () => void;
  }

  let { active = 'chat', collapsed = false, onnavigate, ontoggle }: Props = $props();

  const ROUTES: Array<{ id: RouteId; label: string; icon: string }> = [
    { id: 'chat',       label: 'Chat',       icon: '◌' },
    { id: 'audit',      label: 'Audit',      icon: '◉' },
    { id: 'replay',     label: 'Replay',     icon: '↺' },
    { id: 'hub',        label: 'Hub',        icon: '⬡' },
    { id: 'sync',       label: 'Sync',       icon: '⇄' },
    { id: 'skills',     label: 'Skills',     icon: '◈' },
    { id: 'channels',   label: 'Channels',   icon: '◎' },
    { id: 'delegation', label: 'Delegation', icon: '⬢' },
    { id: 'settings',   label: 'Settings',   icon: '⚙' },
    { id: 'about',      label: 'About',      icon: '◇' },
  ];
</script>

<aside class="sidebar" class:sidebar--collapsed={collapsed} aria-label="Primary navigation">
  <div class="sidebar__brand">
    <Pulse state="idle" size="sm" label="Synaptic" />
    {#if !collapsed}
      <span class="sidebar__brand-text">Synaptic</span>
    {/if}
  </div>

  <Hairline />

  <nav class="sidebar__nav">
    {#each ROUTES as route}
      <button
        class="route"
        class:route--active={active === route.id}
        class:route--collapsed={collapsed}
        type="button"
        onclick={() => onnavigate?.(route.id)}
        aria-label={route.label}
        aria-current={active === route.id ? 'page' : undefined}
        title={collapsed ? route.label : undefined}
      >
        <span class="route__icon" aria-hidden="true">{route.icon}</span>
        {#if !collapsed}
          <span class="route__label">{route.label}</span>
        {/if}
      </button>
    {/each}
  </nav>

  <div class="sidebar__footer">
    <Hairline />
    <button
      class="toggle"
      type="button"
      onclick={ontoggle}
      aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
      title={collapsed ? 'Expand (⌘+\\)' : 'Collapse (⌘+\\)'}
    >
      <span class="toggle__icon" aria-hidden="true">{collapsed ? '›' : '‹'}</span>
      {#if !collapsed}
        <span class="toggle__label">Collapse</span>
      {/if}
    </button>
  </div>
</aside>

<style>
  .sidebar {
    width: 240px;
    height: 100vh;
    display: flex;
    flex-direction: column;
    background-color: var(--surface-raised);
    border-right: 1px solid var(--border-default);
    transition: width var(--duration-base) var(--ease-standard);
    overflow: hidden;
    flex-shrink: 0;
  }

  .sidebar--collapsed {
    width: 60px;
  }

  .sidebar__brand {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-5) var(--space-4);
    height: 56px;
    flex-shrink: 0;
  }

  .sidebar__brand-text {
    font-family: var(--font-serif);
    font-size: var(--text-body-lg-size);
    font-weight: 500;
    color: var(--content-primary);
    letter-spacing: 0.01em;
    transition: opacity var(--duration-base) var(--ease-standard);
  }

  .sidebar--collapsed .sidebar__brand-text {
    opacity: 0;
  }

  .sidebar__nav {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: var(--space-3) var(--space-2);
    flex: 1;
    overflow-y: auto;
  }

  .route {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3);
    height: 36px;
    background-color: transparent;
    border: none;
    border-radius: var(--radius-sm);
    border-left: 2px solid transparent;
    color: var(--content-secondary);
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    cursor: pointer;
    text-align: left;
    transition:
      background-color var(--duration-fast) var(--ease-standard),
      border-color var(--duration-fast) var(--ease-standard),
      color var(--duration-fast) var(--ease-standard);
    white-space: nowrap;
    overflow: hidden;
  }

  .route:hover {
    background-color: var(--paper-warm-50);
    color: var(--content-primary);
  }

  .route:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: -2px;
  }

  .route--active {
    background-color: var(--paper-warm-100);
    border-left-color: var(--content-accent);
    color: var(--content-primary);
    font-weight: 500;
  }

  .route__icon {
    width: 20px;
    height: 20px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    font-family: var(--font-mono);
    font-size: 14px;
    color: inherit;
  }

  .route__label {
    opacity: 1;
    transition: opacity var(--duration-base) var(--ease-standard);
  }

  .sidebar--collapsed .route__label {
    opacity: 0;
    transition-delay: 140ms;
  }

  .sidebar__footer {
    flex-shrink: 0;
  }

  .toggle {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    padding: var(--space-3) var(--space-4);
    background-color: transparent;
    border: none;
    color: var(--content-tertiary);
    font-family: var(--font-sans);
    font-size: var(--text-caption-size);
    cursor: pointer;
    transition: color var(--duration-fast) var(--ease-standard);
    white-space: nowrap;
    overflow: hidden;
  }

  .toggle:hover {
    color: var(--content-primary);
  }

  .toggle:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: -2px;
  }

  .toggle__icon {
    width: 20px;
    text-align: center;
    font-family: var(--font-mono);
    font-size: 16px;
  }

  .toggle__label {
    opacity: 1;
    transition: opacity var(--duration-base) var(--ease-standard);
  }

  .sidebar--collapsed .toggle__label {
    opacity: 0;
    transition-delay: 140ms;
  }
</style>
