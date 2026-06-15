<script lang="ts">
  import { onMount } from 'svelte'

  import Chat from './lib/routes/Chat.svelte'
  import Settings from './lib/routes/Settings.svelte'
  import Audit from './lib/routes/Audit.svelte'
  import Replay from './lib/routes/Replay.svelte'
  import About from './lib/routes/About.svelte'
  import Sidebar from './lib/components/Sidebar.svelte'
  import Toasts from './lib/components/Toasts.svelte'
  import LiveTranscript from './lib/components/LiveTranscript.svelte'
  import VoiceOrb from './lib/components/VoiceOrb.svelte'
  import OnboardingWizard from './lib/components/OnboardingWizard.svelte'
  import { daemon } from './lib/stores/daemon.svelte'
  import { overlay } from './lib/stores/overlay.svelte'
  import { ipc } from './lib/ipc/client'
  import { initStores } from './lib/stores/init'

  let showOnboarding = $state(false)
  let currentHash = $state('#/')

  let route = $derived(
    currentHash === '#/settings' ? 'settings' :
    currentHash === '#/audit' ? 'audit' :
    currentHash === '#/replay' ? 'replay' :
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
  <div class="app-shell" class:overlay-mode={overlay.active}>
    {#if !overlay.active}
      <Sidebar />
    {/if}

    <main class="main">
      <div class="status-bar">
        <span class="conn" class:connected={daemon.connected}>
          {daemon.connected ? '' : ''}
        </span>
        <span class="conn-label" class:connected={daemon.connected}>
          {daemon.connected ? 'connected' : 'disconnected'}
        </span>
      </div>

      {#if overlay.active}
        <div class="overlay-prompt">
          <VoiceOrb />
          <div class="overlay-input-row">
            <input
              type="text"
              class="overlay-input"
              placeholder="Ask Synaptic..."
              onkeydown={(e) => {
                if (e.key === 'Escape') overlay.hide()
              }}
            />
            <button class="overlay-close" onclick={() => overlay.hide()}>
              &times;
            </button>
          </div>
        </div>
      {:else}
        <div class="route-container">
          {#if route === 'settings'}
            <Settings />
          {:else if route === 'audit'}
            <Audit />
          {:else if route === 'replay'}
            <Replay />
          {:else if route === 'about'}
            <About />
          {:else}
            <Chat />
          {/if}
        </div>
      {/if}
    </main>

    <Toasts />
    <LiveTranscript />
  </div>
{/if}

<style>
  .app-shell {
    display: flex;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
    background: var(--color-bg);
  }

  .app-shell.overlay-mode {
    background: transparent;
  }

  .main {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
    position: relative;
  }

  .status-bar {
    position: absolute;
    top: 12px;
    right: 16px;
    display: flex;
    align-items: center;
    gap: 6px;
    z-index: 10;
    padding: 4px 12px;
    border-radius: var(--radius-pill);
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    backdrop-filter: var(--glass-blur);
  }

  .conn {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--color-text-faint);
    transition: background var(--transition-base);
  }

  .conn.connected {
    background: var(--color-success);
    box-shadow: 0 0 8px rgba(74, 222, 128, 0.4);
    animation: breathe 2s ease-in-out infinite;
  }

  .conn-label {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--color-text-faint);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .conn-label.connected {
    color: var(--color-success);
  }

  @keyframes breathe {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
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

  .overlay-prompt {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
  }

  .overlay-input-row {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;
    max-width: 600px;
  }

  .overlay-input {
    flex: 1;
    padding: 16px 24px;
    font-size: 18px;
    font-family: var(--font-sans);
    color: var(--color-text);
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-lg);
    backdrop-filter: var(--glass-blur);
    outline: none;
  }

  .overlay-input:focus {
    border-color: var(--color-accent);
    box-shadow: 0 0 0 2px rgba(var(--color-accent-rgb), 0.2);
  }

  .overlay-close {
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 20px;
    color: var(--color-text-faint);
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: 50%;
    cursor: pointer;
    transition: color var(--transition-base), border-color var(--transition-base);
  }

  .overlay-close:hover {
    color: var(--color-text);
    border-color: var(--color-text-faint);
  }
</style>
