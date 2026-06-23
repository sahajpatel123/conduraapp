<script lang="ts">
  import { onMount } from 'svelte'

  import Chat from './lib/routes/Chat.svelte'
  import Settings from './lib/routes/Settings.svelte'
  import Audit from './lib/routes/Audit.svelte'
  import Replay from './lib/routes/Replay.svelte'
  import About from './lib/routes/About.svelte'
  import Hub from './lib/routes/Hub.svelte'
  import Sync from './lib/routes/Sync.svelte'
  import Skills from './lib/routes/Skills.svelte'
  import Channels from './lib/routes/Channels.svelte'
  import Delegation from './lib/routes/Delegation.svelte'
  import Sidebar from './lib/components/Sidebar.svelte'
  import Toasts from './lib/components/Toasts.svelte'
  import LiveTranscript from './lib/components/LiveTranscript.svelte'
  import OverlayPrompt from './lib/components/OverlayPrompt.svelte'
  import OnboardingWizard from './lib/components/OnboardingWizard.svelte'
  import ConsentModal from './lib/components/ConsentModal.svelte'
  import { daemon } from './lib/stores/daemon.svelte'
  import { consent } from './lib/stores/consent.svelte'
  import { overlay } from './lib/stores/overlay.svelte'
  import { ipc } from './lib/ipc/client'
  import { initStores } from './lib/stores/init'
  import { t } from './lib/i18n'

  let showOnboarding = $state(false)
  let currentHash = $state('#/')

  let route = $derived(
    currentHash === '#/settings' ? 'settings' :
    currentHash === '#/audit' ? 'audit' :
    currentHash === '#/replay' ? 'replay' :
    currentHash === '#/about' ? 'about' :
    currentHash === '#/hub' ? 'hub' :
    currentHash === '#/sync' ? 'sync' :
    currentHash === '#/skills' ? 'skills' :
    currentHash === '#/channels' ? 'channels' :
    currentHash === '#/delegation' ? 'delegation' : 'chat'
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

    // Show the wizard only when BOTH gates say setup is unfinished:
    // the legacy first-run marker AND the onboarding state machine.
    // This keeps upgrades smooth — a user who finished onboarding in
    // an older build (marker set) is never re-wizarded.
    void Promise.all([
      ipc.firstRunStatus().catch(() => ({ complete: false })),
      ipc.onboardingIsComplete().catch(() => true)
    ]).then(([fr, onboardComplete]) => {
      showOnboarding = !fr.complete && !onboardComplete
    }).catch(() => {})

    // Settings "Re-run setup" asks us to re-show the wizard even
    // though the first-run marker is already set.
    const onShowOnboarding = (): void => {
      showOnboarding = true
      window.location.hash = '#/'
    }
    window.addEventListener('synaptic:show-onboarding', onShowOnboarding)

    // Start polling for Gatekeeper consent tickets once the daemon
    // connection is up.
    consent.start()

    return () => {
      window.removeEventListener('hashchange', onHashChange)
      window.removeEventListener('synaptic:show-onboarding', onShowOnboarding)
      consent.stop()
    }
  })

  function completeOnboarding(route?: string): void {
    showOnboarding = false
    if (route) window.location.hash = route
  }
</script>

{#if showOnboarding}
  <div class="onboarding-overlay">
    <OnboardingWizard onComplete={completeOnboarding} />
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
          {daemon.connected ? t('app.status.connected') : t('app.status.disconnected')}
        </span>
      </div>

      {#if overlay.active}
        <OverlayPrompt />
      {:else}
        {#key route}
        <div class="route-container">
          {#if route === 'settings'}
            <Settings />
          {:else if route === 'audit'}
            <Audit />
          {:else if route === 'replay'}
            <Replay />
          {:else if route === 'about'}
            <About />
          {:else if route === 'hub'}
            <Hub />
          {:else if route === 'sync'}
            <Sync />
          {:else if route === 'skills'}
            <Skills />
          {:else if route === 'channels'}
            <Channels />
          {:else if route === 'delegation'}
            <Delegation />
          {:else}
            <Chat />
          {/if}
        </div>
        {/key}
      {/if}
    </main>

    <Toasts />
    <LiveTranscript />
  </div>
{/if}

<ConsentModal />

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

  /* ── Status bar — refined glass pill ─────────────── */
  .status-bar {
    position: absolute;
    top: var(--space-3);
    right: var(--space-4);
    display: flex;
    align-items: center;
    gap: var(--space-2);
    z-index: var(--z-elevated);
    padding: var(--space-1) var(--space-3);
    border-radius: var(--radius-pill);
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    box-shadow: var(--shadow-xs);
    transition: border-color var(--transition-base);
  }

  .status-bar:hover {
    border-color: var(--glass-border-hover);
  }

  .conn {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--color-text-faint);
    transition: background var(--transition-base), box-shadow var(--transition-base);
    flex-shrink: 0;
  }

  .conn.connected {
    background: var(--color-success);
    box-shadow: 0 0 0 3px var(--color-success-soft), 0 0 10px rgba(74, 222, 128, 0.5);
    animation: breathe 2.4s var(--ease-in-out-quart) infinite;
  }

  .conn-label {
    font-family: var(--font-mono);
    font-size: var(--size-2xs);
    color: var(--color-text-faint);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wide);
    font-weight: var(--weight-medium);
    transition: color var(--transition-base);
  }

  .conn-label.connected {
    color: var(--color-success);
  }

  @keyframes breathe {
    0%, 100% { opacity: 1; transform: scale(1); }
    50% { opacity: 0.4; transform: scale(0.85); }
  }

  /* ── Route container — fade-in on route change ──── */
  .route-container {
    flex: 1;
    overflow: hidden;
    display: flex;
    animation: routeFadeIn var(--transition-slow) var(--ease-out-expo);
  }

  @keyframes routeFadeIn {
    from {
      opacity: 0;
      transform: translateY(6px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .route-container :global(div) {
    display: flex;
    flex-direction: column;
    width: 100%;
  }

  /* ── Onboarding overlay — premium entrance ──────── */
  .onboarding-overlay {
    position: fixed;
    inset: 0;
    background: var(--color-bg);
    z-index: var(--z-overlay);
    animation: onboardingIn var(--transition-slow) var(--ease-out-expo);
  }

  @keyframes onboardingIn {
    from {
      opacity: 0;
      transform: scale(1.02);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }
</style>
