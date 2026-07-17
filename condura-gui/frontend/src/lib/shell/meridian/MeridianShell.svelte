<script lang="ts">
  /**
   * MeridianShell — sole product GUI mount.
   * Cool mist · cobalt · living arc · bottom dock.
   * No Inkboard / Living Paper / Lumen imports.
   */
  import { onMount } from 'svelte'
  import './meridian.css'
  import './onboarding-meridian.css'
  import { initStores } from '../../stores/init'
  import { consent } from '../../stores/consent.svelte'
  import { halt } from '../../stores/halt.svelte'
  import { overlay } from '../../stores/overlay.svelte'
  import { conversation } from '../../stores/conversation.svelte'
  import { daemon } from '../../stores/daemon.svelte'
  import { onboarding } from '../../stores/onboarding.svelte'
  import { account } from '../../stores/account.svelte'
  import { ipc } from '../../ipc/client'
  import OnboardingWizard from '../../components/OnboardingWizard.svelte'
  import {
    getResolvedTheme,
    onThemeChange,
    setResolvedTheme,
    toggleLightDark,
    type ResolvedTheme,
  } from '../../theme/condura-theme'
  import { ROUTE_HASH, hashToRoute, type RouteId } from './routes'

  import MeridianArc from './MeridianArc.svelte'
  import MeridianDock from './MeridianDock.svelte'
  import MeridianChat from './MeridianChat.svelte'
  import MeridianHub from './MeridianHub.svelte'
  import MeridianSkills from './MeridianSkills.svelte'
  import MeridianSync from './MeridianSync.svelte'
  import MeridianAudit from './MeridianAudit.svelte'
  import MeridianReplay from './MeridianReplay.svelte'
  import MeridianChannels from './MeridianChannels.svelte'
  import MeridianDelegation from './MeridianDelegation.svelte'
  import MeridianAccount from './MeridianAccount.svelte'
  import MeridianSettings from './MeridianSettings.svelte'
  import MeridianAbout from './MeridianAbout.svelte'
  import MeridianPalette from './MeridianPalette.svelte'
  import MeridianConsent from './MeridianConsent.svelte'
  import MeridianHalt from './MeridianHalt.svelte'
  import MeridianToasts from './MeridianToasts.svelte'
  import MeridianKeys from './MeridianKeys.svelte'

  /** First-run wizard — was only on dead App.svelte; Meridian is the sole mount. */
  let showOnboarding = $state(false)
  let onboardingChecked = $state(false)
  let paletteOpen = $state(false)
  /** Element that opened the palette — focus returns here on close. */
  let paletteTrigger = $state<HTMLElement | null>(null)
  /** Element that opened the cheatsheet — focus returns here on close. */
  let keysTrigger = $state<HTMLElement | null>(null)
  let currentHash = $state(
    typeof window !== 'undefined' ? window.location.hash || '#/' : '#/'
  )
  let route = $derived(hashToRoute(currentHash))
  let rootEl = $state<HTMLDivElement | null>(null)

  function pinShellFloor(): void {
    if (rootEl && rootEl.scrollTop !== 0) rootEl.scrollTop = 0
  }
  let theme = $state<ResolvedTheme>(getResolvedTheme())

  let statusLabel = $derived(
    halt.state.halted
      ? 'Halted'
      : conversation.isStreaming
        ? 'Thinking'
        : consent.ticket
          ? 'Awaiting consent'
          : daemon.connected
            ? 'Ready'
            : daemon.reconnectAttempt > 0
              ? `Reconnecting · ${daemon.reconnectAttempt}`
              : 'Offline'
  )

  let statusTone = $derived<'ok' | 'live' | 'bad'>(
    halt.state.halted || !daemon.connected
      ? 'bad'
      : conversation.isStreaming || consent.ticket || daemon.reconnectAttempt > 0
        ? 'live'
        : 'ok'
  )

  onMount(() => {
    theme = getResolvedTheme()
    try {
      const q = new URLSearchParams(window.location.search)
      const t = q.get('theme')
      if (t === 'light' || t === 'dark') theme = setResolvedTheme(t)
    } catch {
      /* ignore */
    }
    const offTheme = onThemeChange((resolved) => {
      theme = resolved
    })

    void initStores()
      .then(() => {
        try {
          consent.start()
        } catch {
          /* ignore */
        }
        return checkOnboarding()
      })
      .catch((e) => {
        console.warn('initStores failed', e)
        void checkOnboarding()
      })

    // halt/overlay polling start inside initStores after ipc.start —
    // do not race them here (empty same-origin baseURL used to skip start).

    // OAuth deep-link return: Go emits after ExchangeCode succeeds.
    let offOAuth: (() => void) | undefined
    try {
      const eventsOn = window.runtime?.EventsOn
      if (eventsOn) {
        offOAuth = eventsOn('condura:oauth-callback', () => {
          void account.checkStatus()
        })
      }
    } catch {
      /* not in Wails */
    }

    const onHash = () => {
      currentHash = window.location.hash || '#/'
      // Deep links (#/settings/models) must not scroll the shell — only .stage.
      queueMicrotask(pinShellFloor)
    }
    window.addEventListener('hashchange', onHash)
    pinShellFloor()

    const onShowOnboarding = (): void => {
      // Settings re-run opens the wizard. reset() needs the daemon and
      // swallows RPC failures into onboarding.error — clear transport
      // failures so offline Vite preview can show the bundled EULA.
      void onboarding.reset().finally(() => {
        const err = onboarding.error || ''
        if (/Load failed|Failed to fetch|NetworkError|ECONNREFUSED|not connected|IPC client not started/i.test(err)) {
          onboarding.error = null
        }
        showOnboarding = true
        window.location.hash = '#/'
      })
    }
    window.addEventListener('condura:show-onboarding', onShowOnboarding)

    // /help slash command (from the chat composer) opens the cheatsheet.
    // Same shortcut as the keyboard '?' handler — the shell handles both.
    const onShowKeys = (): void => {
      // Capture the trigger so we can return focus when the cheatsheet closes.
      // The '?' keystroke normally fires while the chat composer is focused.
      keysTrigger = (document.activeElement as HTMLElement) ?? null
      keysOpen = true
    }
    window.addEventListener('condura:show-keys', onShowKeys)

    const onKey = (e: KeyboardEvent) => {
      if (showOnboarding) return
      const mod = e.metaKey || e.ctrlKey
      const k = e.key.toLowerCase()
      // Never steal keys while the user is typing (Ask composer, search, token fields).
      const t = e.target as HTMLElement | null
      const typing =
        !!t &&
        (t.tagName === 'INPUT' ||
          t.tagName === 'TEXTAREA' ||
          t.tagName === 'SELECT' ||
          t.isContentEditable)
      if (mod && e.shiftKey && k === 'escape') {
        e.preventDefault()
        void halt.halt('hard_hotkey')
        return
      }
      if (e.key === 'Escape' && paletteOpen) {
        paletteOpen = false
        e.preventDefault()
        return
      }
      if (e.key === 'Escape' && keysOpen) {
        keysOpen = false
        e.preventDefault()
        return
      }
      if (mod && k === 'k') {
        e.preventDefault()
        paletteTrigger = (document.activeElement as HTMLElement) ?? null
        paletteOpen = true
        return
      }
      if (mod && k === ',') {
        e.preventDefault()
        navigate('settings')
        return
      }
      // Theme: Shift+T only outside editable fields so capital "T" works in Ask.
      if (!typing && e.shiftKey && !mod && k === 't') {
        e.preventDefault()
        theme = toggleLightDark()
      }
      // Copy last assistant response — Alt+C fires on the chat route.
      // Dispatched as an event so MeridianChat (only mounted on chat) owns
      // the actual clipboard + "Copied" feedback.
      if (e.altKey && !mod && k === 'c' && route === 'chat' && !typing) {
        e.preventDefault()
        window.dispatchEvent(new CustomEvent('condura:copy-last-response'))
        return
      }
      // New ask — ⌘⇧N mirrors the same event pattern: chat owns the
      // actual clearThread() so it can confirm before clobbering an
      // in-flight stream.
      if (mod && e.shiftKey && k === 'n' && route === 'chat' && !typing) {
        e.preventDefault()
        window.dispatchEvent(new CustomEvent('condura:new-ask'))
        return
      }
      // Settings tab jump — ⌘1..5 fires a custom event that
      // MeridianSettings listens for. Working from anywhere when the
      // user is on the Settings route.
      if (mod && !e.shiftKey && !e.altKey && route === 'settings' && !typing) {
        const idx = '12345'.indexOf(k)
        if (idx >= 0) {
          e.preventDefault()
          window.dispatchEvent(new CustomEvent('condura:settings-tab', { detail: { index: idx } }))
          return
        }
      }
      // Dock primary tabs — ⌘1..8 jump to the 8 primary dock tabs
      // (Ask · Hub · Skills · Sync · Audit · Replay · Channels · Agents).
      // Works from anywhere — including Settings. We use a different
      // chord (no Shift) so it doesn't collide with the ⌘1..5 Settings
      // shortcut (the Settings chord only fires when route === 'settings').
      if (mod && !e.shiftKey && !e.altKey && !typing) {
        const idx = '12345678'.indexOf(k)
        if (idx >= 0 && idx < 8) {
          e.preventDefault()
          const routeIds: RouteId[] = ['chat', 'hub', 'skills', 'sync', 'audit', 'replay', 'channels', 'delegation']
          const target = routeIds[idx]
          if (target) navigate(target)
          return
        }
      }
      // Export thread as Markdown — ⌘⇧E dispatches when on chat route.
      if (mod && e.shiftKey && k === 'e' && route === 'chat' && !typing) {
        e.preventDefault()
        window.dispatchEvent(new CustomEvent('condura:export-thread'))
        return
      }
      // Regenerate last assistant response — ⌘⇧R re-sends the last user
      // message. Disabled while streaming via the button state.
      if (mod && e.shiftKey && k === 'r' && route === 'chat' && !typing) {
        e.preventDefault()
        window.dispatchEvent(new CustomEvent('condura:regenerate-last'))
        return
      }
      // Open Audit — ⌘⇧A navigates to the audit ledger from anywhere.
      // Common path when debugging a refused/errored action.
      if (mod && e.shiftKey && k === 'a' && !typing) {
        e.preventDefault()
        navigate('audit')
        return
      }
      // Open Sync — ⌘⇧S navigates to the device-pairing surface.
      if (mod && e.shiftKey && k === 's' && !typing) {
        e.preventDefault()
        navigate('sync')
        return
      }
      // Open Hub — ⌘⇧H navigates to the community shelf.
      if (mod && e.shiftKey && k === 'h' && !typing) {
        e.preventDefault()
        navigate('hub')
        return
      }
      // Open Channels — ⌘⇧C navigates to the messaging-integrations
      // surface. ⌘⇧R is context-aware: Regenerate when in chat route,
      // navigate to Replay otherwise.
      if (mod && e.shiftKey && k === 'c' && !typing) {
        e.preventDefault()
        navigate('channels')
        return
      }
      // Open Replay — ⌘⇧R navigates to the day-meridian surface when
      // the user is NOT on the chat route. (On chat route it's the
      // Regenerate shortcut above — different surface, same chord.)
      if (mod && e.shiftKey && k === 'r' && route !== 'chat' && !typing) {
        e.preventDefault()
        navigate('replay')
        return
      }
      // ⌘. (Cmd-period) — macOS convention for "stop". On the chat
      // route, cancel any in-flight stream. Mirrors the Esc handler in
      // MeridianChat so power users don't have to leave the home row.
      if (mod && e.key === '.' && route === 'chat' && !typing) {
        e.preventDefault()
        window.dispatchEvent(new CustomEvent('condura:stop-stream'))
        return
      }
      // Help: ? opens the keyboard cheatsheet. Mirrors VS Code/Cursor — not
      // gated on `typing`, so "?" in Ask composer still opens help. `?`
      // requires Shift on most keyboards, so we only forbid Cmd/Ctrl.
      if (e.key === '?' && !mod) {
        e.preventDefault()
        keysOpen = !keysOpen
      }
      // /help slash command dispatches the same intent.
      // Listen at the shell level so it works regardless of chat mount.
      // Implementation lives in onMount via window.addEventListener below.
    }
    window.addEventListener('keydown', onKey)

    return () => {
      offTheme()
      try {
        offOAuth?.()
      } catch {
        /* ignore */
      }
      window.removeEventListener('hashchange', onHash)
      window.removeEventListener('keydown', onKey)
      window.removeEventListener('condura:show-onboarding', onShowOnboarding)
      window.removeEventListener('condura:show-keys', onShowKeys)
      try {
        consent.stop()
      } catch {
        /* ignore */
      }
      try {
        halt.stopPolling()
      } catch {
        /* ignore */
      }
      try {
        overlay.stop()
      } catch {
        /* ignore */
      }
    }
  })

  async function checkOnboarding(): Promise<void> {
    try {
      const [fr, onboardComplete] = await Promise.all([
        ipc.firstRunStatus().catch(() => ({ complete: false as boolean })),
        ipc.onboardingIsComplete().catch(() => true),
      ])
      // Daemon onboarding.is_complete is authoritative; first-run marker is
      // also written by onboarding.finish. Show wizard if either says first-run.
      const finished = !!onboardComplete || !!fr.complete
      showOnboarding = !finished
    } catch {
      // Fail open to main UI rather than trapping users offline.
      showOnboarding = false
    } finally {
      onboardingChecked = true
    }
  }

  function completeOnboarding(routeHash?: string): void {
    showOnboarding = false
    // finish() already writes first-run-complete; belt-and-suspenders marker.
    void ipc.firstRunComplete().catch(() => {})
    if (routeHash) window.location.hash = routeHash
  }

  $effect(() => {
    route
    currentHash
    queueMicrotask(pinShellFloor)
  })

  function navigate(r: RouteId): void {
    window.location.hash = ROUTE_HASH[r]
  }
</script>

{#if showOnboarding}
  <div class="md onboarding-layer" role="dialog" aria-modal="true" aria-label="Welcome to Condura">
    <div class="onboarding-wash" aria-hidden="true"></div>
    <OnboardingWizard onComplete={completeOnboarding} />
  </div>
{:else if !onboardingChecked}
  <div class="md boot" aria-busy="true" aria-label="Starting Condura">
    <div class="boot-gem" aria-hidden="true"></div>
    <p class="boot-mark">Condura</p>
    <p class="boot-sub">Starting…</p>
  </div>
{:else}
  <div class="md root" bind:this={rootEl}>
    <div class="wash" aria-hidden="true"></div>
    <div class="grain" aria-hidden="true"></div>

    <header class="top">
      <div class="brand">
        <span class="gem" aria-hidden="true"></span>
        <span class="word">Condura</span>
        <span class="edition">Meridian</span>
      </div>

      <button type="button" class="jump" onclick={() => { paletteTrigger = (document.activeElement as HTMLElement) ?? null; paletteOpen = true }} aria-label="Search (⌘K)">
        <span>Jump anywhere…</span>
        <kbd>⌘K</kbd>
      </button>

      <div class="right">
        <div class="status" data-tone={statusTone} aria-live="polite">
          <span class="dot"></span>
          {statusLabel}
        </div>
        <button
          type="button"
          class="icon keys-toggle"
          onclick={() => {
            keysTrigger = (document.activeElement as HTMLElement) ?? null
            keysOpen = true
          }}
          aria-label="Keyboard shortcuts (?)"
          title="Keyboard shortcuts (?)"
        >
          <span class="keys-label" aria-hidden="true">Keys</span>
        </button>
        <button
          type="button"
          class="icon theme-toggle"
          onclick={() => setResolvedTheme(theme === 'light' ? 'dark' : 'light')}
          aria-label="Toggle theme ({theme})"
          title="Switch light / dark (⇧T)"
        >
          <span class="theme-label" aria-hidden="true">{theme === 'light' ? 'Light' : 'Dark'}</span>
          {#if theme === 'light'}
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z" />
            </svg>
          {:else}
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <circle cx="12" cy="12" r="5" />
              <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" />
            </svg>
          {/if}
        </button>
      </div>
    </header>

    <div class="arc-wrap">
      <MeridianArc tone={statusTone} />
    </div>

    <main class="stage">
      {#key route}
        {#if route === 'chat'}
          <MeridianChat />
        {:else if route === 'hub'}
          <MeridianHub />
        {:else if route === 'skills'}
          <MeridianSkills />
        {:else if route === 'sync'}
          <MeridianSync />
        {:else if route === 'audit'}
          <MeridianAudit />
        {:else if route === 'replay'}
          <MeridianReplay />
        {:else if route === 'channels'}
          <MeridianChannels />
        {:else if route === 'delegation'}
          <MeridianDelegation />
        {:else if route === 'account'}
          <MeridianAccount />
        {:else if route === 'settings'}
          <MeridianSettings />
        {:else if route === 'about'}
          <MeridianAbout />
        {/if}
      {/key}
    </main>

    <MeridianDock route={route} onnavigate={navigate} />
    <MeridianPalette
      open={paletteOpen}
      route={route}
      onclose={() => { paletteOpen = false; queueMicrotask(() => paletteTrigger?.focus({ preventScroll: true })); paletteTrigger = null }}
      onnavigate={navigate}
    />
    <MeridianKeys open={keysOpen} onclose={() => { keysOpen = false; queueMicrotask(() => keysTrigger?.focus({ preventScroll: true })); keysTrigger = null }} />
    <MeridianConsent />
    <MeridianToasts />
    {#if halt.state.halted}
      <MeridianHalt />
    {/if}
  </div>
{/if}

<style>
  .onboarding-layer {
    position: fixed;
    inset: 0;
    z-index: 100;
    overflow: auto;
    background: var(--md-mist);
  }
  .onboarding-wash {
    position: fixed;
    inset: -10% -6%;
    background: var(--md-wash);
    filter: blur(32px);
    pointer-events: none;
    z-index: 0;
  }
  .onboarding-layer :global(.wizard-container) {
    position: relative;
    z-index: 1;
    min-height: 100dvh;
  }
  .boot {
    height: 100vh;
    height: 100dvh;
    display: grid;
    place-content: center;
    justify-items: center;
    gap: 10px;
    background: var(--md-mist);
    background-image: var(--md-wash);
    text-align: center;
  }
  .boot-gem {
    width: 10px;
    height: 10px;
    border-radius: 2px;
    background: var(--md-cobalt);
    box-shadow: none;
    animation: md-breathe 2.8s var(--md-ease) infinite;
    margin-bottom: 4px;
  }
  .boot-mark {
    font-family: var(--md-font-display);
    font-size: 26px;
    font-weight: 650;
    letter-spacing: -0.04em;
    margin: 0;
    color: var(--md-ink);
  }
  .boot-sub {
    margin: 0;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .root {
    height: 100vh;
    height: 100dvh;
    display: grid;
    grid-template-rows: auto auto minmax(0, 1fr);
    position: relative;
    /* clip — not hidden — so deep-link scrollIntoView cannot shift the shell
       and leave a mist shelf under the stage. Only .stage scrolls. */
    overflow: clip;
    overscroll-behavior: none;
    background: var(--md-mist);
  }
  .wash {
    position: absolute;
    inset: -8% -4% -4%;
    background: var(--md-wash);
    filter: blur(32px);
    transform: translateZ(0);
    pointer-events: none;
    z-index: 0;
  }
  .grain {
    position: absolute;
    inset: 0;
    z-index: 0;
    pointer-events: none;
    opacity: 0.045;
    background-image: var(--md-grain);
    background-repeat: repeat;
    background-size: 180px 180px;
    mix-blend-mode: multiply;
  }
  :root[data-mode='dark'] .grain {
    opacity: 0.07;
    mix-blend-mode: soft-light;
  }
  .root::before {
    content: '';
    position: absolute;
    inset: 0;
    z-index: 0;
    pointer-events: none;
    background:
      linear-gradient(
        180deg,
        color-mix(in oklab, var(--md-mist) 0%, transparent) 0%,
        color-mix(in oklab, var(--md-stage) 18%, transparent) 45%,
        color-mix(in oklab, var(--md-stage) 28%, transparent) 100%
      );
  }
  .top {
    position: relative;
    z-index: 3;
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    align-items: center;
    gap: 16px;
    padding: 14px 22px 6px;
    min-height: 58px;
    animation: md-fade 400ms var(--md-ease) both;
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
    flex: none;
    z-index: 1;
    justify-self: start;
    grid-column: 1;
  }
  .gem {
    width: 10px;
    height: 10px;
    border-radius: 2px;
    flex: none;
    background: var(--md-cobalt);
    box-shadow: none;
  }
  .word {
    font-family: var(--md-font-display);
    font-size: 20px;
    font-weight: 650;
    letter-spacing: -0.04em;
  }
  .edition {
    font-family: var(--md-font-mono);
    font-size: 9px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    padding: 2px 6px;
    border-radius: 4px;
    border: 1px solid var(--md-line);
    background: transparent;
  }
  .jump {
    position: relative;
    grid-column: 2;
    justify-self: center;
    width: min(400px, 42vw);
    max-width: 400px;
    z-index: 4;
    isolation: isolate;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 9px 14px;
    border-radius: 10px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 90%, transparent);
    color: var(--md-ink-mute);
    font-size: 13px;
    font-weight: 500;
    letter-spacing: -0.01em;
    cursor: pointer;
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    box-shadow: none;
    transition:
      border-color 140ms var(--md-ease),
      color 140ms var(--md-ease),
      background 140ms var(--md-ease);
  }
  .jump:hover {
    border-color: var(--md-line-strong);
    color: var(--md-ink);
    background: var(--md-surface);
  }
  .jump kbd {
    font-family: var(--md-font-mono);
    font-size: 10px;
    padding: 3px 7px;
    border-radius: 6px;
    background: color-mix(in oklab, var(--md-stage) 70%, var(--md-surface));
    border: 1px solid var(--md-line);
    color: var(--md-ink-faint);
  }
  .right {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 10px;
    flex: none;
    z-index: 1;
    justify-self: end;
    grid-column: 3;
  }
  .status {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--md-ink-mute);
    padding: 6px 10px;
    border-radius: 9px;
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
    border: 1px solid var(--md-line);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    box-shadow: none;
  }
  .dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--md-ink-faint);
  }
  .status[data-tone='ok'] .dot {
    background: var(--md-live);
    box-shadow: none;
  }
  .status[data-tone='live'] .dot {
    background: var(--md-cobalt);
    box-shadow: none;
    animation: md-pulse 1.2s var(--md-ease) infinite;
  }
  .status[data-tone='bad'] .dot {
    background: var(--md-halt);
    box-shadow: none;
  }
  .icon {
    width: 34px;
    height: 34px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 9px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
    color: var(--md-ink-mute);
    cursor: pointer;
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    box-shadow: none;
    transition: border-color 140ms var(--md-ease), color 140ms var(--md-ease), background 140ms var(--md-ease);
  }
  .icon:hover {
    border-color: var(--md-line-strong);
    color: var(--md-ink);
    background: var(--md-surface);
  }
  /* Theme toggle: hide the SVG behind a "Light"/"Dark" label by
     default — the icon is recognizable but the state name is even
     more so. Hover reveals both. */
  .theme-toggle .theme-label {
    font-family: var(--md-font-mono);
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }
  .theme-toggle svg {
    transition: opacity 140ms var(--md-ease);
  }
  .theme-toggle:hover svg,
  .theme-toggle:focus-visible svg {
    opacity: 0.35;
  }
  .icon:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
    border-color: var(--md-cobalt);
  }
  .kbd-q {
    font-family: var(--md-font-mono);
    font-size: 13px;
    font-weight: 600;
    letter-spacing: 0;
    line-height: 1;
  }
  /* Cheatsheet toggle: dual-channel label like the theme toggle so
     the button's purpose is obvious without hovering for the title. */
  .keys-toggle .keys-label {
    font-family: var(--md-font-mono);
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }
  .jump:focus-visible {
    outline: none;
    border-color: var(--md-cobalt);
    color: var(--md-ink);
    box-shadow: var(--md-focus);
  }
  .arc-wrap {
    position: relative;
    z-index: 1;
    padding: 10px 16px 4px;
    /* Clear breathing room under Jump / header — never kiss the chrome */
    margin-top: 8px;
    margin-bottom: 0;
    pointer-events: none;
    overflow: visible;
  }
  .stage {
    position: relative;
    z-index: 1;
    min-height: 0;
    overflow: auto;
    background: var(--md-stage);
    border-radius: 12px 12px 0 0;
    margin: 0 12px;
    border: 1px solid color-mix(in oklab, var(--md-line-strong) 80%, transparent);
    border-bottom: 0;
    box-shadow: none;
    /* Dock floats over the stage — keep scroll content clear of it. */
    padding-bottom: calc(88px + env(safe-area-inset-bottom, 0px));
  }
  .stage :global(> *) {
    animation: md-rise 420ms var(--md-ease) both;
  }
  @keyframes md-pulse {
    0%, 100% { transform: scale(1); opacity: 1; }
    50% { transform: scale(1.35); opacity: 0.7; }
  }
  @media (max-width: 720px) {
    .edition,
    .jump span {
      display: none;
    }
    .top {
      padding: 12px 14px;
      gap: 10px;
      min-height: 48px;
    }
    .jump {
      width: 44px;
      max-width: 44px;
      justify-content: center;
      padding: 10px;
    }
    .jump kbd { display: none; }
    .status {
      padding: 6px 9px;
      font-size: 9px;
    }
    .stage {
      margin: 0 8px;
      border-radius: 12px 12px 0 0;
      padding-bottom: calc(96px + env(safe-area-inset-bottom, 0px));
    }
  }
  @media (max-width: 420px) {
    .word { font-size: 18px; }
    .right { gap: 6px; }
    .icon {
      width: 34px;
      height: 34px;
      border-radius: 10px;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .top, .status .dot, .boot-gem, .stage :global(> *) { animation: none !important; }
  }
</style>
