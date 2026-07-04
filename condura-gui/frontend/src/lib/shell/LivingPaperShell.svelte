<script lang="ts">
  /**
   * LivingPaperShell — The root orchestrator for the Living Paper GUI.
   * 
   * Layout: [TopBar] [NavOrbit | ContentCanvas] [StatusThread]
   * Overlays: Onboarding, CommandPalette, ConsentModal, KillSwitch
   * 
   * This file manages no business logic — it delegates to stores
   * and route components. The daemon contract is unchanged.
   */
  import { onMount } from 'svelte'
  import { initStores } from '../stores/init'
  import { ipc } from '../ipc/client'
  import { onboarding } from '../stores/onboarding.svelte'
  import { consent } from '../stores/consent.svelte'
  import { halt } from '../stores/halt.svelte'
  import { overlay } from '../stores/overlay.svelte'
  import { conversation } from '../stores/conversation.svelte'
  import { daemon } from '../stores/daemon.svelte'

  // Living Paper primitives
  import { PaperSurface, QuillCursor, PaperDivider } from '$lib/components/living'

  // Shell components
  import TopBar from './TopBar.svelte'
  import NavOrbit, { type RouteId } from './NavOrbit.svelte'
  import StatusThread from './StatusThread.svelte'

  // Route components — these still use condura/ components internally
  // but are wrapped in the Living Paper shell. Will be migrated to
  // Living Paper primitives route by route.
  import Chat from '$lib/condura/Chat.svelte'
  import Audit from '$lib/condura/Audit.svelte'
  import Replay from '$lib/condura/Replay.svelte'
  import Hub from '$lib/condura/Hub.svelte'
  import Sync from '$lib/condura/Sync.svelte'
  import Skills from '$lib/condura/Skills.svelte'
  import Channels from '$lib/condura/Channels.svelte'
  import Delegation from '$lib/condura/Delegation.svelte'
  import Settings from '$lib/condura/Settings.svelte'
  import About from '$lib/condura/About.svelte'
  import { FloatingOnboarding } from '$lib/components/onboarding'

  // Overlays
  import CommandPalette from '$lib/condura/CommandPalette.svelte'
  import QuickPromptOverlay from '$lib/condura/QuickPromptOverlay.svelte'
  import ConsentModal from '$lib/condura/ConsentModal.svelte'
  import KillSwitchOverlay from '$lib/condura/KillSwitchOverlay.svelte'

  import { ROUTE_HASH, hashToRoute } from '$lib/condura/NavRail.svelte'

  // ── State ────────────────────────────────────────────────
  let showOnboarding = $state(false)
  let paletteOpen = $state(false)
  let quickOpen = $state(false)
  let navCollapsed = $state(true)
  let currentHash = $state(
    typeof window !== 'undefined' ? window.location.hash || '#/' : '#/'
  )
  let route = $derived(hashToRoute(currentHash))
  let theme = $state<'light' | 'dark'>('light')

  // Agent phase derived from conversation + halt + consent stores
  let agentPhase = $derived(
    conversation.isStreaming
      ? 'thinking'
      : halt.state.halted
        ? 'error'
        : consent.ticket
          ? 'consent'
          : daemon.connected
            ? 'idle'
            : 'error'
  )

  const routeLabels: Record<RouteId, string> = {
    chat: 'Chat',
    audit: 'Audit',
    replay: 'Replay',
    hub: 'Hub',
    sync: 'Sync',
    skills: 'Skills',
    channels: 'Channels',
    delegation: 'Delegation',
    settings: 'Settings',
    about: 'About',
  }

  // ── Lifecycle ────────────────────────────────────────────
  onMount(() => {
    theme = (document.documentElement.dataset.mode as 'light' | 'dark') ?? 'light'

    try { initStores() } catch (e) { console.warn('initStores failed', e) }
    try { halt.startPolling() } catch (e) { console.warn('halt.startPolling failed', e) }
    try { overlay.start() } catch (e) { console.warn('overlay.start failed', e) }

    void Promise.all([
      ipc.firstRunStatus().catch(() => ({ complete: false })),
      ipc.onboardingIsComplete().catch(() => true),
    ]).then(([fr, oc]) => {
      const daemonComplete = !!(fr.complete && oc)
      let seen = false
      try { seen = !!localStorage.getItem('condura-ritual-seen') } catch { /* ignore */ }
      showOnboarding = !daemonComplete || !seen
      if (!seen) {
        try { localStorage.setItem('condura-ritual-seen', '1') } catch { /* ignore */ }
      }
    }).catch(() => {})

    const onHash = () => { currentHash = window.location.hash || '#/' }
    window.addEventListener('hashchange', onHash)

    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault()
        paletteOpen = true
        return
      }
      if (!e.shiftKey) return
      const k = e.key.toLowerCase()
      if (k === 'o') { e.preventDefault(); showOnboarding = true }
      else if (k === 'p') { e.preventDefault(); quickOpen = true }
    }
    window.addEventListener('keydown', onKey)

    return () => {
      window.removeEventListener('hashchange', onHash)
      window.removeEventListener('keydown', onKey)
      try { consent.stop() } catch { /* ignore */ }
      try { halt.stopPolling() } catch { /* ignore */ }
      try { overlay.stop() } catch { /* ignore */ }
    }
  })

  function setTheme(t: 'light' | 'dark'): void {
    theme = t
    document.documentElement.dataset.mode = t
    try { localStorage.setItem('condura-theme', t) } catch { /* ignore */ }
  }

  function navigate(r: RouteId): void {
    window.location.hash = ROUTE_HASH[r]
  }

  function toggleNav(): void {
    navCollapsed = !navCollapsed
  }

  // Toggle between collapsed/expanded nav via `Ctrl+\`
  // This is handled in the shell layout
</script>

<!-- Paper grain background for the entire app -->
<div class="lp lp-living-shell" data-mode={theme}>
  <PaperSurface variant="page" grain={true} padding="0" style="height: 100vh; display: flex; flex-direction: column; overflow: hidden;">
    <!-- Decorative cursor trail -->
    <QuillCursor />

    {#if showOnboarding}
      <!-- Floating onboarding card wizard -->
      <FloatingOnboarding
        oncomplete={() => { showOnboarding = false }}
      />
    {:else}
      <!-- ── Main App Shell ────────────────────────────── -->
      <div style="
        display: flex;
        flex-direction: column;
        height: 100%;
        position: relative;
        z-index: 1;
      ">
        <!-- Top bar -->
        <TopBar
          title={routeLabels[route]}
          agentPhase={agentPhase}
          theme={theme}
          onThemeToggle={() => setTheme(theme === 'light' ? 'dark' : 'light')}
          onPalette={() => (paletteOpen = true)}
        />

        <!-- Main content: NavOrbit + Route -->
        <div style="
          display: flex;
          flex: 1;
          overflow: hidden;
          position: relative;
        ">
          <!-- NavOrbit — toggle with Ctrl+\ or mouse hover edge -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div
            style="position: relative; display: flex;"
            onmouseenter={() => (navCollapsed = false)}
            onmouseleave={(e) => {
              // Only collapse if mouse leaves entirely (not to child)
              const rect = e.currentTarget.getBoundingClientRect()
              if (e.clientX >= rect.right || e.clientX <= rect.left) {
                navCollapsed = true
              }
            }}
          >
            <NavOrbit
              route={route}
              onnavigate={navigate}
              collapsed={navCollapsed}
            />
          </div>

          <!-- Content area — paper grain background -->
          <PaperSurface
            variant="page"
            grain={true}
            padding="0"
            style="flex: 1; overflow-y: auto; overflow-x: hidden; position: relative;"
          >
            <!-- Route content with key for re-render on navigation -->
            {#key route}
              {#if route === 'chat'}
                <Chat route={route} />
              {:else if route === 'audit'}
                <Audit />
              {:else if route === 'replay'}
                <Replay />
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
              {:else if route === 'settings'}
                <Settings />
              {:else if route === 'about'}
                <About />
              {/if}
            {/key}
          </PaperSurface>
        </div>

        <!-- Status thread -->
        <StatusThread
          agentPhase={agentPhase}
          agentLabel={daemon.connected ? 'Connected' : 'Disconnected'}
          halted={halt.state.halted}
          onKill={() => {
            if (halt.state.halted) {
              halt.resume()
            } else {
              halt.halt()
            }
          }}
        />
      </div>
    {/if}

    <!-- ── Overlays ────────────────────────────────────── -->
    {#if paletteOpen}
      <CommandPalette
        open={paletteOpen}
        onclose={() => (paletteOpen = false)}
        onnavigate={(r: RouteId) => { navigate(r); paletteOpen = false }}
      />
    {/if}

    {#if quickOpen}
      <QuickPromptOverlay
        open={quickOpen}
        onclose={() => (quickOpen = false)}
      />
    {/if}

    {#if consent.ticket}
      <ConsentModal
        ticket={consent.ticket}
        onresponse={() => { /* consent store handles it */ }}
      />
    {/if}

    {#if halt.state.halted}
      <KillSwitchOverlay
        onresume={() => halt.resume()}
      />
    {/if}
  </PaperSurface>
</div>
