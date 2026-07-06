<script lang="ts">
  /**
   * LivingPaperShell — The root orchestrator for the Living Paper GUI.
   *
   * Layout: [TopBar] [NavOrbit | ContentCanvas] [StatusThread]
   * Overlays: Onboarding, CommandPalette, ConsentModal, KillSwitch
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

  import { PaperSurface, QuillCursor } from '$lib/components/living'

  import TopBar from './TopBar.svelte'
  import NavOrbit from './NavOrbit.svelte'
  import StatusThread from './StatusThread.svelte'

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
  import Account from '$lib/condura/Account.svelte'
  import { FloatingOnboarding } from '$lib/components/onboarding'
  import {
    getResolvedTheme,
    onThemeChange,
    setResolvedTheme,
    toggleLightDark,
    type ResolvedTheme,
  } from '../theme/condura-theme'

  import CommandPalette from '$lib/condura/CommandPalette.svelte'
  import QuickPromptOverlay from '$lib/condura/QuickPromptOverlay.svelte'
  import ConsentModal from '$lib/condura/ConsentModal.svelte'
  import KillSwitchOverlay from '$lib/condura/KillSwitchOverlay.svelte'

  import { ROUTE_HASH, hashToRoute, type RouteId } from '$lib/condura/NavRail.svelte'

  type ShellRoute = RouteId | 'replay'

  function shellHashToRoute(hash: string): ShellRoute {
    if (hash.startsWith('#/replay')) return 'replay'
    return hashToRoute(hash)
  }

  let showOnboarding = $state(false)
  let paletteOpen = $state(false)
  let quickOpen = $state(false)
  let currentHash = $state(
    typeof window !== 'undefined' ? window.location.hash || '#/' : '#/'
  )
  let route = $derived(shellHashToRoute(currentHash))
  let theme = $state<ResolvedTheme>(getResolvedTheme())

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

  const routeLabels: Record<ShellRoute, string> = {
    chat: 'Chat',
    hub: 'Hub',
    skills: 'Skills',
    sync: 'Sync',
    audit: 'Audit',
    replay: 'Replay',
    channels: 'Channels',
    delegation: 'Delegation',
    account: 'Account',
    settings: 'Settings',
    about: 'About',
  }

  onMount(() => {
    theme = getResolvedTheme()
    const offTheme = onThemeChange((resolved) => {
      theme = resolved
    })

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

    let gArmed = false
    let gArmedAt = 0

    const onKey = (e: KeyboardEvent) => {
      const mod = e.metaKey || e.ctrlKey
      const k = e.key.toLowerCase()

      if (mod && e.shiftKey && k === 'escape') {
        e.preventDefault()
        try { void halt.halt('hard_hotkey') } catch { /* ignore */ }
        return
      }

      if (e.key === 'Escape' && !e.shiftKey && !mod) {
        if (paletteOpen) {
          paletteOpen = false
          e.preventDefault()
          return
        }
        if (quickOpen) {
          quickOpen = false
          e.preventDefault()
          return
        }
      }

      if (mod) {
        switch (k) {
          case 'k':
            e.preventDefault()
            paletteOpen = true
            return
          case 'p':
            if (e.shiftKey) {
              e.preventDefault()
              quickOpen = true
              return
            }
            break
          case ',':
            e.preventDefault()
            navigate('settings')
            return
          case '0':
            e.preventDefault()
            navigate('account')
            return
        }

        if (k >= '1' && k <= '9') {
          const idx = Number(k) - 1
          const order: (RouteId | null)[] = [
            'chat',
            'hub',
            'skills',
            'sync',
            'audit',
            'channels',
            'delegation',
            null,
            'settings',
          ]
          const target = order[idx]
          if (target) {
            e.preventDefault()
            navigate(target)
            return
          }
        }
      }

      if (e.shiftKey && !mod && k === 't') {
        e.preventDefault()
        theme = toggleLightDark()
        return
      }

      if (e.shiftKey && !mod && k === 'o') {
        e.preventDefault()
        showOnboarding = true
        return
      }

      if (!mod && e.key === '?' && !e.shiftKey) {
        const target = e.target as HTMLElement | null
        const tag = target?.tagName
        if (tag === 'INPUT' || tag === 'TEXTAREA' || target?.isContentEditable) return
        e.preventDefault()
        paletteOpen = true
        return
      }

      if (gArmed && Date.now() - gArmedAt <= 1200) {
        const map: Record<string, ShellRoute> = {
          s: 'settings',
          h: 'hub',
          a: 'about',
          c: 'channels',
          k: 'skills',
          r: 'replay',
          l: 'sync',
          d: 'delegation',
        }
        if (map[k]) {
          e.preventDefault()
          navigateShell(map[k])
          gArmed = false
          return
        }
        gArmed = false
      }
      if (!mod && !e.shiftKey && k === 'g') {
        gArmed = true
        gArmedAt = Date.now()
      }
    }
    window.addEventListener('keydown', onKey)

    return () => {
      offTheme()
      window.removeEventListener('hashchange', onHash)
      window.removeEventListener('keydown', onKey)
      try { consent.stop() } catch { /* ignore */ }
      try { halt.stopPolling() } catch { /* ignore */ }
      try { overlay.stop() } catch { /* ignore */ }
    }
  })

  function setTheme(t: ResolvedTheme): void {
    theme = setResolvedTheme(t)
  }

  function navigate(r: RouteId): void {
    window.location.hash = ROUTE_HASH[r]
  }

  function navigateShell(r: ShellRoute): void {
    if (r === 'replay') {
      window.location.hash = '#/replay'
      return
    }
    navigate(r)
  }
</script>

<div class="lp lp-living-shell">
  <PaperSurface variant="page" grain={true} padding="0" style="height: 100vh; display: flex; flex-direction: column; overflow: hidden;">
    <QuillCursor />

    {#if showOnboarding}
      <FloatingOnboarding oncomplete={() => { showOnboarding = false }} />
    {:else}
      <div class="lp-shell-body">
        <TopBar
          title={routeLabels[route]}
          agentPhase={agentPhase}
          theme={theme}
          onThemeToggle={() => setTheme(theme === 'light' ? 'dark' : 'light')}
          onPalette={() => (paletteOpen = true)}
        />

        <div class="lp-shell-main">
          <NavOrbit
            route={route === 'replay' ? 'chat' : route}
            activeRoute={route === 'replay' ? null : route}
            onnavigate={navigate}
          />

          <PaperSurface
            variant="page"
            grain={true}
            padding="0"
            style="flex: 1; overflow-y: auto; overflow-x: hidden; position: relative; min-width: 0;"
          >
            {#key route}
              {#if route === 'chat'}
                <Chat route="chat" />
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
              {:else if route === 'account'}
                <Account />
              {:else if route === 'settings'}
                <Settings />
              {:else if route === 'about'}
                <About />
              {/if}
            {/key}
          </PaperSurface>
        </div>

        <StatusThread
          agentPhase={agentPhase}
          agentLabel={daemon.connected ? 'Connected' : 'Disconnected'}
          halted={halt.state.halted}
          onKill={() => {
            if (halt.state.halted) halt.resume()
            else halt.halt()
          }}
        />
      </div>
    {/if}

    {#if paletteOpen}
      <CommandPalette
        open={paletteOpen}
        onclose={() => (paletteOpen = false)}
        onnavigate={(r: RouteId) => {
          if (String(r) === 'replay') navigateShell('replay')
          else navigate(r)
          paletteOpen = false
        }}
      />
    {/if}

    {#if quickOpen}
      <QuickPromptOverlay open={quickOpen} onclose={() => (quickOpen = false)} />
    {/if}

    {#if consent.ticket}
      <ConsentModal ticket={consent.ticket} onresponse={() => { /* store handles */ }} />
    {/if}

    {#if halt.state.halted}
      <KillSwitchOverlay onresume={() => halt.resume()} />
    {/if}
  </PaperSurface>
</div>

<style>
  .lp-shell-body {
    display: flex;
    flex-direction: column;
    height: 100%;
    position: relative;
    z-index: 1;
  }

  .lp-shell-main {
    display: flex;
    flex: 1;
    overflow: hidden;
    position: relative;
    min-height: 0;
  }
</style>
