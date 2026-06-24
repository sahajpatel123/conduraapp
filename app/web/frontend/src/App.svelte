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
    <div class="onboarding-glow"></div>
    <OnboardingWizard onComplete={completeOnboarding} />
  </div>
{:else}
  <div class="app-shell" class:overlay-mode={overlay.active}>
    {#if !overlay.active}
      <Sidebar />
    {/if}

    <main class="main">
      <div class="status-bar" class:connected={daemon.connected}>
        <span class="conn-dot"></span>
        <span class="conn-ring"></span>
        <span class="conn-label">
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

  /* ── Status bar — living breathing pill ──────────── */
  .status-bar {
    position: absolute;
    top: var(--space-3);
    right: var(--space-4);
    display: flex;
    align-items: center;
    gap: var(--space-2);
    z-index: var(--z-elevated);
    padding: 6px 14px;
    border-radius: var(--radius-pill);
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    box-shadow: var(--shadow-xs), var(--shadow-inset);
    transition: all var(--transition-base);
  }

  .status-bar:hover {
    border-color: var(--glass-border-hover);
    box-shadow: var(--shadow-sm), var(--shadow-inset);
  }

  .status-bar.connected {
    border-color: rgba(16, 185, 129, 0.15);
  }

  .conn-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--color-text-faint);
    flex-shrink: 0;
    position: relative;
    transition: background var(--transition-base);
  }

  .status-bar.connected .conn-dot {
    background: var(--color-success);
    box-shadow: 0 0 8px var(--color-success-glow);
    animation: breathe 2.4s var(--ease-in-out-quart) infinite;
  }

  .conn-ring {
    position: absolute;
    left: 10px;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    border: 2px solid var(--color-success);
    opacity: 0;
    pointer-events: none;
  }

  .status-bar.connected .conn-ring {
    animation: ring-expand 2.4s var(--ease-out-quart) infinite;
  }

  .conn-label {
    font-family: var(--font-mono);
    font-size: var(--size-2xs);
    color: var(--color-text-faint);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
    font-weight: var(--weight-semibold);
    transition: color var(--transition-base);
  }

  .status-bar.connected .conn-label {
    color: var(--color-success);
  }

  /* ── Route container — premium entrance ──────────── */
  .route-container {
    flex: 1;
    overflow: hidden;
    display: flex;
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo);
  }

  .route-container :global(div) {
    display: flex;
    flex-direction: column;
    width: 100%;
  }

  /* ── Onboarding overlay — cinematic entrance ─────── */
  .onboarding-overlay {
    position: fixed;
    inset: 0;
    background: var(--color-bg);
    z-index: var(--z-overlay);
    animation: fade-in var(--transition-slow) var(--ease-out-expo);
    overflow: hidden;
  }

  .onboarding-glow {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 600px;
    height: 600px;
    border-radius: 50%;
    background: radial-gradient(circle, var(--color-accent-soft) 0%, transparent 70%);
    pointer-events: none;
    animation: breathe-soft 6s ease-in-out infinite;
    opacity: 0.6;
  }
</style>
