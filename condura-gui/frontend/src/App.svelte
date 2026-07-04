<script lang="ts">
  import { onMount } from 'svelte'

  import ChatV1 from './lib/routes/ChatV1.svelte'
  import SettingsPane from '$components/v1/SettingsPane.svelte'
  import Audit from './lib/routes/Audit.svelte'
  import Replay from './lib/routes/Replay.svelte'
  import About from './lib/routes/About.svelte'
  import Hub from './lib/routes/Hub.svelte'
  import Sync from './lib/routes/Sync.svelte'
  import Skills from './lib/routes/Skills.svelte'
  import Channels from './lib/routes/Channels.svelte'
  import Delegation from './lib/routes/Delegation.svelte'
  import DevComponents from './lib/routes/dev/Components.svelte'
  import V1 from './lib/routes/dev/V1.svelte'

  import Sidebar, { type RouteId } from '$components/v1/Sidebar.svelte'
  import StatusBar from '$components/v1/StatusBar.svelte'
  import NavPalette from '$components/v1/NavPalette.svelte'
  import ConversationDrawer from '$components/v1/ConversationDrawer.svelte'
  import ConsentModalHost from '$components/v1/ConsentModalHost.svelte'
  import KillSwitchOverlay from '$components/v1/KillSwitchOverlay.svelte'
  import Button from '$components/v1/Button.svelte'
  import Pulse from '$components/v1/Pulse.svelte'
  import Hairline from '$components/v1/Hairline.svelte'
  import Inline from '$components/v1/Inline.svelte'

  import Toasts from './lib/components/Toasts.svelte'
  import LiveTranscript from './lib/components/LiveTranscript.svelte'
  import OverlayPrompt from './lib/components/OverlayPrompt.svelte'
  import OnboardingWizard from './lib/components/OnboardingWizard.svelte'

  import { daemon } from './lib/stores/daemon.svelte'
  import { consent } from './lib/stores/consent.svelte'
  import { overlay } from './lib/stores/overlay.svelte'
  import { conversation } from './lib/stores/conversation.svelte'
  import { halt } from './lib/stores/halt.svelte'
  import { ipc } from './lib/ipc/client'
  import { initStores } from './lib/stores/init'
  import { t } from './lib/i18n'

  let showOnboarding = $state(false)
  let currentHash = $state('#/')
  let paletteOpen = $state(false)
  let sidebarCollapsed = $state(false)
  let drawerOpen = $state(false)

  const ROUTE_HASH: Record<RouteId, string> = {
    chat: '#/',
    audit: '#/audit',
    replay: '#/replay',
    hub: '#/hub',
    sync: '#/sync',
    skills: '#/skills',
    channels: '#/channels',
    delegation: '#/delegation',
    settings: '#/settings',
    about: '#/about',
  }

  function hashToRoute(hash: string): RouteId | 'dev-components' | 'v1-preview' {
    if (hash.startsWith('#/settings')) return 'settings'
    if (hash.startsWith('#/audit')) return 'audit'
    if (hash.startsWith('#/replay')) return 'replay'
    if (hash.startsWith('#/about')) return 'about'
    if (hash.startsWith('#/hub')) return 'hub'
    if (hash.startsWith('#/sync')) return 'sync'
    if (hash.startsWith('#/skills')) return 'skills'
    if (hash.startsWith('#/channels')) return 'channels'
    if (hash.startsWith('#/delegation')) return 'delegation'
    if (hash.startsWith('#/dev/components')) return 'dev-components'
    if (hash.startsWith('#/dev/v1')) return 'v1-preview'
    return 'chat'
  }

  let route = $derived(hashToRoute(currentHash))

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
    route === 'v1-preview' ? 'v1 Design Preview' :
    t('nav.chat')
  )

  let agentState = $derived(
    conversation.isStreaming ? ('thinking' as const) :
    halt.state.halted ? ('error' as const) :
    ('idle' as const)
  )

  let activeTask = $derived(
    conversation.isStreaming ? conversation.currentTitle : null
  )

  const drawerConversations = $derived(
    conversation.conversations.map((c) => ({
      id: String(c.id),
      date: formatDrawerDate(c.updated_at),
      firstSentence: c.title,
      agentActed: c.message_count > 1,
      active: c.id === conversation.currentID,
    }))
  )

  function formatDrawerDate(iso: string): string {
    const d = new Date(iso)
    const now = new Date()
    if (d.toDateString() === now.toDateString()) {
      return `Today ${d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
    }
    const yesterday = new Date(now)
    yesterday.setDate(now.getDate() - 1)
    if (d.toDateString() === yesterday.toDateString()) return 'Yesterday'
    return d.toLocaleDateString()
  }

  function navigate(routeId: RouteId): void {
    window.location.hash = ROUTE_HASH[routeId]
  }

  onMount(() => {
    currentHash = window.location.hash || '#/'

    const onHashChange = (): void => {
      currentHash = window.location.hash
    }
    window.addEventListener('hashchange', onHashChange)

    void initStores()

    void Promise.all([
      ipc.firstRunStatus().catch(() => ({ complete: false })),
      ipc.onboardingIsComplete().catch(() => true),
    ]).then(([fr, onboardComplete]) => {
      showOnboarding = !fr.complete && !onboardComplete
    }).catch(() => {})

    const onShowOnboarding = (): void => {
      showOnboarding = true
      window.location.hash = '#/'
    }
    window.addEventListener('synaptic:show-onboarding', onShowOnboarding)

    consent.start()

    const onKey = (e: KeyboardEvent): void => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault()
        paletteOpen = true
      } else if ((e.metaKey || e.ctrlKey) && e.key === '\\') {
        e.preventDefault()
        sidebarCollapsed = !sidebarCollapsed
      } else if (e.key === 'Escape') {
        paletteOpen = false
        drawerOpen = false
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

  function completeOnboarding(routeHash?: string): void {
    showOnboarding = false
    if (routeHash) window.location.hash = routeHash
  }

  async function handleDrawerSelect(id: string): Promise<void> {
    const num = Number(id)
    if (!Number.isNaN(num) && num > 0) {
      await conversation.open(num)
      window.location.hash = '#/'
    }
    drawerOpen = false
  }

  async function handlePause(): Promise<void> {
    if (conversation.isStreaming) {
      await conversation.cancel()
    }
  }

  async function handleKill(): Promise<void> {
    await halt.halt('user requested from status bar')
  }

  async function handleResume(): Promise<void> {
    await halt.resume()
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
    { id: 'history', label: 'Conversation history', hint: 'Open drawer', group: 'Actions', onselect: () => { drawerOpen = true } },
    { id: 'dev', label: 'Component smoke', hint: 'Design system preview', group: 'Dev', onselect: () => { window.location.hash = '#/dev/components' } },
    { id: 'v1', label: 'v1 shell demo', hint: 'Isolated preview', group: 'Dev', onselect: () => { window.location.hash = '#/dev/v1' } },
  ])
</script>

{#if showOnboarding}
  <div class="onboarding-shell">
    <OnboardingWizard onComplete={completeOnboarding} />
  </div>
{:else}
  <div class="shell" class:shell--overlay={overlay.active}>
    {#if !overlay.active}
      <Sidebar
        active={route === 'dev-components' || route === 'v1-preview' ? 'chat' : route}
        collapsed={sidebarCollapsed}
        onnavigate={navigate}
        ontoggle={() => { sidebarCollapsed = !sidebarCollapsed }}
      />
    {/if}

    <div class="main" class:main--overlay={overlay.active}>
      {#if !overlay.active}
        <header class="main__topbar">
          <div class="main__topbar-left">
            <Pulse state={agentState === 'thinking' ? 'thinking' : 'idle'} size="sm" label="Synaptic" />
            <span class="main__title">{routeTitle}</span>
            {#if !daemon.connected}
              <span class="main__offline">offline</span>
            {/if}
          </div>
          <Inline gap="2">
            {#if route === 'chat'}
              <Button size="sm" variant="tertiary" onclick={() => { drawerOpen = !drawerOpen }}>
                History
              </Button>
            {/if}
            <Button size="sm" variant="secondary" onclick={() => { paletteOpen = true }}>
              ⌘K
            </Button>
          </Inline>
        </header>
        <Hairline />
      {/if}

      <div class="main__content">
        {#if overlay.active}
          <OverlayPrompt />
        {:else}
          {#key route}
            <div class="route-container">
              {#if route === 'settings'}
                <SettingsPane />
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
              {:else if route === 'v1-preview'}
                <V1 />
              {:else}
                <ChatV1 />
              {/if}
            </div>
          {/key}
        {/if}
      </div>
    </div>

    {#if !overlay.active}
      <StatusBar
        {activeTask}
        queuedCount={0}
        {agentState}
        onopen={() => { paletteOpen = true }}
        onpause={handlePause}
        onkill={handleKill}
      />
    {/if}

    <ConversationDrawer
      conversations={drawerConversations}
      open={drawerOpen}
      onselect={handleDrawerSelect}
      onclose={() => { drawerOpen = false }}
    />

    <Toasts />
    <LiveTranscript />
  </div>
{/if}

<ConsentModalHost />

{#if halt.state.halted}
  <KillSwitchOverlay reason="user" onresume={handleResume} />
{/if}

<NavPalette
  bind:open={paletteOpen}
  items={paletteItems}
  placeholder={t('common.search')}
  emptyMessage={t('common.no_results') ?? 'No results'}
/>

<style>
  .shell {
    display: flex;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
    background: var(--surface-base);
    color: var(--content-primary);
    position: relative;
  }

  .shell--overlay {
    background: transparent;
  }

  .main {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
    overflow: hidden;
  }

  .main--overlay {
    background: transparent;
  }

  .main__topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-3) var(--space-5);
    height: 56px;
    flex-shrink: 0;
    background-color: var(--surface-raised);
  }

  .main__topbar-left {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    min-width: 0;
  }

  .main__title {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--content-tertiary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .main__offline {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--status-warning-fg);
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }

  .main__content {
    flex: 1;
    min-height: 0;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .route-container {
    flex: 1;
    min-height: 0;
    overflow: hidden;
    display: flex;
    animation: fade-in-up var(--duration-slow) var(--ease-decelerate) both;
  }

  .route-container > :global(*) {
    flex: 1;
    min-height: 0;
  }

  .onboarding-shell {
    position: fixed;
    inset: 0;
    background: var(--surface-base);
    z-index: var(--z-modal);
    overflow: hidden;
  }

  @media (prefers-reduced-motion: reduce) {
    .route-container {
      animation: none;
    }
  }
</style>
