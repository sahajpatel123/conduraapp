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

  /** First-run wizard — was only on dead App.svelte; Meridian is the sole mount. */
  let showOnboarding = $state(false)
  let onboardingChecked = $state(false)
  let paletteOpen = $state(false)
  let currentHash = $state(
    typeof window !== 'undefined' ? window.location.hash || '#/' : '#/'
  )
  let route = $derived(hashToRoute(currentHash))
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
      .then(() => checkOnboarding())
      .catch((e) => {
        console.warn('initStores failed', e)
        void checkOnboarding()
      })

    try {
      halt.startPolling()
    } catch {
      /* ignore */
    }
    try {
      overlay.start()
    } catch {
      /* ignore */
    }
    try {
      consent.start()
    } catch {
      /* ignore */
    }

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
    }
    window.addEventListener('hashchange', onHash)

    const onShowOnboarding = (): void => {
      // Settings re-run already calls onboarding.reset(); belt-and-suspenders
      // so any other dispatcher still walks EULA → perms → hotkey.
      void onboarding
        .reset()
        .catch(() => {})
        .finally(() => {
          showOnboarding = true
          window.location.hash = '#/'
        })
    }
    window.addEventListener('condura:show-onboarding', onShowOnboarding)

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
      if (mod && k === 'k') {
        e.preventDefault()
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
    <p class="boot-mark">Condura</p>
    <p class="boot-sub">Starting…</p>
  </div>
{:else}
  <div class="md root">
    <div class="wash" aria-hidden="true"></div>

    <header class="top">
      <div class="brand">
        <span class="word">Condura</span>
        <span class="edition">Meridian</span>
      </div>

      <button type="button" class="jump" onclick={() => (paletteOpen = true)} aria-label="Search (⌘K)">
        <span>Jump anywhere…</span>
        <kbd>⌘K</kbd>
      </button>

      <div class="right">
        <div class="status" data-tone={statusTone}>
          <span class="dot"></span>
          {statusLabel}
        </div>
        <button
          type="button"
          class="icon"
          onclick={() => setResolvedTheme(theme === 'light' ? 'dark' : 'light')}
          aria-label="Toggle theme"
        >
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
      onclose={() => (paletteOpen = false)}
      onnavigate={navigate}
    />
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
    gap: 8px;
    background: var(--md-mist);
    text-align: center;
  }
  .boot-mark {
    font-family: var(--md-font-display);
    font-size: 28px;
    letter-spacing: -0.04em;
    margin: 0;
    color: var(--md-ink);
  }
  .boot-sub {
    margin: 0;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .root {
    height: 100vh;
    height: 100dvh;
    display: flex;
    flex-direction: column;
    position: relative;
    overflow: hidden;
    background: var(--md-mist);
  }
  .wash {
    position: absolute;
    inset: -8% -4% -4%;
    background: var(--md-wash);
    filter: blur(28px);
    transform: translateZ(0);
    pointer-events: none;
    z-index: 0;
  }
  .root::before {
    /* Soft veil so mist never meets the stage as a hard band */
    content: '';
    position: absolute;
    inset: 0;
    z-index: 0;
    pointer-events: none;
    background:
      linear-gradient(
        180deg,
        color-mix(in oklab, var(--md-mist) 0%, transparent) 0%,
        color-mix(in oklab, var(--md-stage) 28%, transparent) 42%,
        color-mix(in oklab, var(--md-stage) 55%, transparent) 100%
      );
  }
  .top {
    position: relative;
    z-index: 2;
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    align-items: center;
    gap: 16px;
    padding: 14px 22px;
    min-height: 56px;
    animation: md-fade 400ms var(--md-ease) both;
  }
  .brand {
    display: flex;
    align-items: baseline;
    gap: 10px;
    flex: none;
    z-index: 1;
    justify-self: start;
    grid-column: 1;
  }
  .word {
    font-family: var(--md-font-display);
    font-size: 22px;
    font-weight: 700;
    letter-spacing: -0.05em;
  }
  .edition {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .jump {
    position: relative;
    grid-column: 2;
    justify-self: center;
    width: min(420px, 42vw);
    max-width: 420px;
    z-index: 1;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 10px 16px;
    border-radius: 999px;
    border: 1px solid var(--md-line-strong);
    background: color-mix(in oklab, var(--md-surface) 70%, transparent);
    color: var(--md-ink-mute);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    backdrop-filter: blur(10px);
    transition: border-color var(--md-dur) var(--md-ease), transform 200ms var(--md-spring), box-shadow 200ms var(--md-ease);
  }
  .jump:hover {
    border-color: color-mix(in oklab, var(--md-cobalt) 45%, transparent);
    color: var(--md-ink);
    transform: translateY(-1px);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--md-cobalt) 12%, transparent);
  }
  .jump kbd {
    font-family: var(--md-font-mono);
    font-size: 10px;
    padding: 3px 7px;
    border-radius: 6px;
    background: var(--md-stage);
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
    padding: 7px 11px;
    border-radius: 999px;
    background: color-mix(in oklab, var(--md-surface) 55%, transparent);
    border: 1px solid var(--md-line);
  }
  .dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--md-ink-faint);
  }
  .status[data-tone='ok'] .dot {
    background: var(--md-live);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-live) 22%, transparent);
  }
  .status[data-tone='live'] .dot {
    background: var(--md-cobalt);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-cobalt) 22%, transparent);
    animation: md-pulse 1.2s var(--md-ease) infinite;
  }
  .status[data-tone='bad'] .dot {
    background: var(--md-halt);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-halt) 20%, transparent);
  }
  .icon {
    width: 36px;
    height: 36px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 12px;
    border: 1px solid var(--md-line-strong);
    background: color-mix(in oklab, var(--md-surface) 65%, transparent);
    color: var(--md-ink-mute);
    cursor: pointer;
    transition: transform 180ms var(--md-spring), border-color var(--md-dur) var(--md-ease), color var(--md-dur) var(--md-ease), box-shadow var(--md-dur) var(--md-ease);
  }
  .icon:hover {
    transform: scale(1.04);
    border-color: color-mix(in oklab, var(--md-cobalt) 40%, transparent);
    color: var(--md-ink);
  }
  .icon:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
    border-color: var(--md-cobalt);
  }
  .jump:focus-visible {
    outline: none;
    border-color: var(--md-cobalt);
    box-shadow: var(--md-focus);
    color: var(--md-ink);
  }
  .arc-wrap {
    position: relative;
    z-index: 1;
    padding: 0 12px;
    margin-top: -4px;
  }
  .stage {
    position: relative;
    z-index: 1;
    flex: 1;
    min-height: 0;
    overflow: auto;
    background: color-mix(in oklab, var(--md-stage) 62%, transparent);
    border-radius: 28px 28px 0 0;
    margin: 0 12px;
    border: 1px solid var(--md-line);
    border-bottom: 0;
    box-shadow: inset 0 1px 0 color-mix(in oklab, var(--md-surface) 55%, transparent);
    /* Keep page content clear of the floating dock */
    padding-bottom: 8px;
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
    .stage { margin: 0 8px; border-radius: 22px 22px 0 0; }
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
    .top, .status .dot { animation: none !important; }
  }
</style>
