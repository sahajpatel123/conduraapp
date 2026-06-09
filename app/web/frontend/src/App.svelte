<script lang="ts">
  import Router from 'svelte-spa-router'
  import { onMount } from 'svelte'

  import Chat from './lib/routes/Chat.svelte'
  import Settings from './lib/routes/Settings.svelte'
  import Audit from './lib/routes/Audit.svelte'
  import About from './lib/routes/About.svelte'
  import Sidebar from './lib/components/Sidebar.svelte'
  import Toasts from './lib/components/Toasts.svelte'
  import OnboardingWizard from './lib/components/OnboardingWizard.svelte'
  import { daemon } from './lib/stores/daemon.svelte'
  import { ipc } from './lib/ipc/client'

  const routes = {
    '/': Chat,
    '/settings': Settings,
    '/audit': Audit,
    '/about': About
  }

  let showOnboarding = $state(false)
  let currentHash = $state(typeof window !== 'undefined' ? window.location.hash : '')

  // Listen for hash changes (svelte-spa-router updates window.location.hash).
  if (typeof window !== 'undefined') {
    window.addEventListener('hashchange', () => {
      currentHash = window.location.hash
    })
  }

  onMount(async () => {
    // Check first-run status; show wizard if not yet complete.
    try {
      const s = await ipc.firstRunStatus()
      if (!s.complete) {
        showOnboarding = true
      }
    } catch {
      // ignore
    }
  })

  function closeOnboarding(): void {
    showOnboarding = false
  }
</script>

{#if showOnboarding}
  <div class="onboarding-overlay">
    <OnboardingWizard />
  </div>
{:else}
  <div class="app-shell">
    <Sidebar />

    <main class="main">
      <nav class="topbar">
        <div class="topbar-left">
          <a href="#/" class:active={currentHash === '#/' || currentHash === '' || currentHash === '#'}>Chat</a>
          <a href="#/settings" class:active={currentHash === '#/settings'}>Settings</a>
          <a href="#/audit" class:active={currentHash === '#/audit'}>Audit</a>
          <a href="#/about" class:active={currentHash === '#/about'}>About</a>
        </div>
        <div class="topbar-right">
          <span class="conn" class:connected={daemon.connected}>
            {daemon.connected ? '● connected' : '○ disconnected'}
          </span>
        </div>
      </nav>

      <div class="route-container">
        <Router {routes} />
      </div>
    </main>

    <Toasts />
  </div>
{/if}

<style>
  .app-shell {
    display: flex;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
  }
  .main {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 var(--space-4);
    height: 48px;
    background: var(--color-bg-elevated);
    border-bottom: 1px solid var(--color-border);
  }
  .topbar-left {
    display: flex;
    gap: var(--space-4);
  }
  .topbar-left a {
    color: var(--color-text-muted);
    font-size: var(--size-md);
    padding: var(--space-2) 0;
    border-bottom: 2px solid transparent;
    text-decoration: none;
  }
  .topbar-left a:hover {
    color: var(--color-text);
  }
  .topbar-left a.active {
    color: var(--color-text);
    border-bottom-color: var(--color-accent);
  }
  .topbar-right .conn {
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    color: var(--color-text-faint);
  }
  .topbar-right .conn.connected {
    color: var(--color-success);
  }
  .route-container {
    flex: 1;
    overflow: hidden;
    display: flex;
  }
  .route-container :global(div) {
    display: flex;
    flex-direction: column;
    width: 100%;
  }
  .onboarding-overlay {
    position: fixed;
    inset: 0;
    background: var(--color-bg);
    z-index: 100;
  }
</style>
