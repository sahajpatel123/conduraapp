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
  import DevComponents from './lib/routes/dev/Components.svelte'
  import Sidebar from './lib/components/Sidebar.svelte'
  import Toasts from './lib/components/Toasts.svelte'
  import LiveTranscript from './lib/components/LiveTranscript.svelte'
  import OverlayPrompt from './lib/components/OverlayPrompt.svelte'
  import OnboardingWizard from './lib/components/OnboardingWizard.svelte'
  import ConsentModal from './lib/components/ConsentModal.svelte'
  import StatusRail from './lib/components/StatusRail.svelte'
  import TitleBar from './lib/components/TitleBar.svelte'
  import CommandPalette from './lib/components/ui/CommandPalette.svelte'
  import { daemon } from './lib/stores/daemon.svelte'
  import { consent } from './lib/stores/consent.svelte'
  import { overlay } from './lib/stores/overlay.svelte'
  import { ipc } from './lib/ipc/client'
  import { initStores } from './lib/stores/init'
  import { t } from './lib/i18n'

  let showOnboarding = $state(false)
  let currentHash = $state('#/')
  let paletteOpen = $state(false)

  let route = $derived(
    currentHash.startsWith('#/settings') ? 'settings' :
    currentHash.startsWith('#/audit') ? 'audit' :
    currentHash.startsWith('#/replay') ? 'replay' :
    currentHash.startsWith('#/about') ? 'about' :
    currentHash.startsWith('#/hub') ? 'hub' :
    currentHash.startsWith('#/sync') ? 'sync' :
    currentHash.startsWith('#/skills') ? 'skills' :
    currentHash.startsWith('#/channels') ? 'channels' :
    currentHash.startsWith('#/delegation') ? 'delegation' :
    currentHash.startsWith('#/dev/components') ? 'dev-components' : 'chat'
  )

  let routeTitle = $derived(
    route === 'settings' ? t('nav.settings') :
    route === 'audit' ? t('nav.audit') :
    route === 'replay' ? t('nav.replay') :
    route === 'about' ? t('nav.about') :
    route === 'hub' ? t('nav.hub') :
    route === 'sync' ? t('nav.sync') :
    route === 'skills' ? t('nav.skills') :
    route === 'channels' ? t('nav.channels') :
    route === 'delegation' ? t('nav.delegation') :
    route === 'dev-components' ? 'Component smoke' :
    t('nav.chat')
  )

  onMount(() => {
    currentHash = window.location.hash || '#/'

    const onHashChange = (): void => {
      currentHash = window.location.hash
    }
    window.addEventListener('hashchange', onHashChange)

    void initStores()

    void Promise.all([
      ipc.firstRunStatus().catch(() => ({ complete: false })),
      ipc.onboardingIsComplete().catch(() => true)
    ]).then(([fr, onboardComplete]) => {
      showOnboarding = !fr.complete && !onboardComplete
    }).catch(() => {})

    const onShowOnboarding = (): void => {
      showOnboarding = true
      window.location.hash = '#/'
    }
    window.addEventListener('synaptic:show-onboarding', onShowOnboarding)

    consent.start()

    // Global ⌘K / Ctrl+K opens the command palette from anywhere
    const onKey = (e: KeyboardEvent): void => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault()
        paletteOpen = true
      }
    }
    window.addEventListener('keydown', onKey)

    const onOpenPalette = (): void => { paletteOpen = true }
    window.addEventListener('synaptic:open-palette', onOpenPalette)

    return () => {
      window.removeEventListener('hashchange', onHashChange)
      window.removeEventListener('synaptic:show-onboarding', onShowOnboarding)
      window.removeEventListener('synaptic:open-palette', onOpenPalette)
      window.removeEventListener('keydown', onKey)
      consent.stop()
    }
  })

  function completeOnboarding(route?: string): void {
    showOnboarding = false
    if (route) window.location.hash = route
  }

  const paletteItems = $derived([
    { id: 'chat', label: t('nav.chat'), group: t('common.search'), onselect: () => { window.location.hash = '#/' } },
    { id: 'settings', label: t('nav.settings'), group: t('common.search'), onselect: () => { window.location.hash = '#/settings' } },
    { id: 'audit', label: t('nav.audit'), group: t('common.search'), onselect: () => { window.location.hash = '#/audit' } },
    { id: 'replay', label: t('nav.replay'), group: t('common.search'), onselect: () => { window.location.hash = '#/replay' } },
    { id: 'hub', label: t('nav.hub'), group: t('common.search'), onselect: () => { window.location.hash = '#/hub' } },
    { id: 'sync', label: t('nav.sync'), group: t('common.search'), onselect: () => { window.location.hash = '#/sync' } },
    { id: 'skills', label: t('nav.skills'), group: t('common.search'), onselect: () => { window.location.hash = '#/skills' } },
    { id: 'channels', label: t('nav.channels'), group: t('common.search'), onselect: () => { window.location.hash = '#/channels' } },
    { id: 'delegation', label: t('nav.delegation'), group: t('common.search'), onselect: () => { window.location.hash = '#/delegation' } },
    { id: 'about', label: t('nav.about'), group: t('common.search'), onselect: () => { window.location.hash = '#/about' } },
    { id: 'dev', label: 'Component smoke', hint: 'Design system preview', group: 'Dev', onselect: () => { window.location.hash = '#/dev/components' } },
  ])
</script>

{#if showOnboarding}
  <div class="onboarding-shell">
    <div class="onboarding-glow"></div>
    <OnboardingWizard onComplete={completeOnboarding} />
  </div>
{:else}
  <div class="app-shell" class:overlay-mode={overlay.active}>
    {#if !overlay.active}
      <Sidebar />
    {/if}

    <div class="content" class:overlay-mode={overlay.active}>
      {#if !overlay.active}
        <TitleBar title={routeTitle} />
      {/if}

      <main class="main">
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
              {:else if route === 'dev-components'}
                <DevComponents />
              {:else}
                <Chat />
              {/if}
            </div>
          {/key}
        {/if}
      </main>

      {#if !overlay.active}
        <StatusRail />
      {/if}
    </div>

    <Toasts />
    <LiveTranscript />
  </div>
{/if}

<ConsentModal />
<CommandPalette
  bind:open={paletteOpen}
  items={paletteItems}
  placeholder={t('common.search')}
  emptyMessage={t('common.no_results') ?? 'No results'}
/>

<style>
  .app-shell {
    display: flex;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
    background: var(--bg);
  }

  .app-shell.overlay-mode { background: transparent; }

  .content {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
    position: relative;
  }

  .content.overlay-mode { background: transparent; }

  .main {
    flex: 1;
    overflow: hidden;
    display: flex;
    position: relative;
  }

  .route-container {
    flex: 1;
    overflow: hidden;
    display: flex;
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }
  .route-container > :global(*) {
    flex: 1;
    min-height: 0;
  }

  /* ── Onboarding shell — cinematic ─────────────────── */
  .onboarding-shell {
    position: fixed;
    inset: 0;
    background: var(--bg);
    z-index: var(--z-overlay);
    animation: fade-in var(--transition-slow) var(--ease-out-expo) both;
    overflow: hidden;
  }

  .onboarding-glow {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 720px;
    height: 720px;
    border-radius: 50%;
    background: radial-gradient(circle, var(--accent-soft) 0%, transparent 70%);
    pointer-events: none;
    animation: breathe-soft 6s ease-in-out infinite;
    opacity: 0.7;
  }
</style>