<script lang="ts">
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
  import { initStores } from './lib/stores/init'

  let showOnboarding = $state(false)
  let currentHash = $state('#/')

  let route = $derived(
    currentHash === '#/settings' ? 'settings' :
    currentHash === '#/audit' ? 'audit' :
    currentHash === '#/about' ? 'about' : 'chat'
  )

  onMount(() => {
    currentHash = window.location.hash || '#/'

    const onHashChange = (): void => {
      currentHash = window.location.hash
    }
    window.addEventListener('hashchange', onHashChange)

    // Initialize stores after the component tree is mounted.
    // This ensures Svelte's reactive context is fully set up
    // before any daemon communication starts.
    void initStores()

    // Check first-run status; show wizard if not yet complete.
    ipc.firstRunStatus().then((s) => {
      if (!s.complete) showOnboarding = true
    }).catch(() => {})

    return () => window.removeEventListener('hashchange', onHashChange)
  })
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
          <a href="#/" class:active={route === 'chat'}>Chat</a>
          <a href="#/settings" class:active={route === 'settings'}>Settings</a>
          <a href="#/audit" class:active={route === 'audit'}>Audit</a>
          <a href="#/about" class:active={route === 'about'}>About</a>
        </div>
        <div class="topbar-right">
          <span class="conn" class:connected={daemon.connected}>
            {daemon.connected ? '● connected' : '○ disconnected'}
          </span>
        </div>
      </nav>

      <div class="route-container">
        {#if route === 'settings'}
          <Settings />
        {:else if route === 'audit'}
          <Audit />
        {:else if route === 'about'}
          <About />
        {:else}
          <Chat />
        {/if}
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
