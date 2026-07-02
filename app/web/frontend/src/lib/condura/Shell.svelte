<script lang="ts">
  /**
   * Shell.svelte — Condura Shell · v0.1.0
   *
   * The persistent desktop window that holds every post-Ritual surface in
   * Condura. Implements SCREEN_SHELL.md: Titlebar + NavRail + active route
   * (Main) + optional right rail + StatusBar + global overlays.
   *
   * Responsibilities (per spec):
   *   - Mount the Titlebar at the top, NavRail on the left, the active route
   *     in Main, the right rail, and StatusBar at the bottom.
   *   - 3-zone Titlebar: wordmark + route context (L) | TitlebarThread (C) |
   *     DynamicIsland + ⌘K hint + theme toggle (R).
   *   - Route transitions via `{#key route}` + `.route-enter` blur-in.
   *   - 9 reachable states (first-time, daemon-unreachable, no-API-key,
   *     onboarding-incomplete, kill-switch, consent-pending, streaming,
   *     static-error, signed-out) read from stores; Shell is the
   *     orchestrator.
   *   - 16 global keyboard chords (⌘K, ⌘⇧P, ⌘,, ⌘1-⌘0, ⌘[, ⌘], ⌘N, ⇧T, Esc,
   *     ?, Cmd+Shift+Escape hard kill) + Go-to two-key (`g` + letter).
   *   - Mount beat-by-beat: NavRail items stagger 60ms apart, main content
   *     fades + slides up 8px.
   *   - ThemePicker integration: clicking the theme toggle opens the
   *     ThemePicker sheet; the ThemePicker handles the clip-path reveal.
   *
   * The Shell composes from existing primitives. No third-party modal
   * libraries. Token discipline: no raw hex, no raw font, no raw spacing.
   */

  import { onMount } from 'svelte';
  import { initStores } from '../stores/init';
  import { ipc } from '../ipc/client';
  import { consent } from '../stores/consent.svelte';
  import { halt } from '../stores/halt.svelte';
  import { overlay } from '../stores/overlay.svelte';
  import { conversation } from '../stores/conversation.svelte';
  import { daemon } from '../stores/daemon.svelte';

  import NavRail, { type RouteId, ROUTE_HASH, hashToRoute } from './NavRail.svelte';
  import DynamicIsland from './DynamicIsland.svelte';
  import TitlebarThread from './TitlebarThread.svelte';
  import Chat from './Chat.svelte';
  import Ritual from './Ritual.svelte';
  import CommandPalette from './CommandPalette.svelte';
  import QuickPromptOverlay from './QuickPromptOverlay.svelte';
  import Replay from './Replay.svelte';
  import Audit from './Audit.svelte';
  import Sync from './Sync.svelte';
  import Settings from './Settings.svelte';
  import Hub from './Hub.svelte';
  import Skills from './Skills.svelte';
  import Channels from './Channels.svelte';
  import About from './About.svelte';
  import ConsentModal from './ConsentModal.svelte';
  import KillSwitchOverlay from './KillSwitchOverlay.svelte';
  import Glyph from './Glyph.svelte';
  import Delegation from './Delegation.svelte';
  import Thread from './Thread.svelte';
  import ThemePicker from './ThemePicker.svelte';

  // ── Local state ───────────────────────────────────────────────
  let showOnboarding = $state(false);
  let paletteOpen = $state(false);
  let quickOpen = $state(false);
  let themePickerOpen = $state(false);
  let currentHash = $state(
    typeof window !== 'undefined' ? window.location.hash || '#/' : '#/',
  );
  let route = $derived(hashToRoute(currentHash));
  let theme = $state<'light' | 'dark'>('light');

  // Mount beat-by-beat: NavRail items stagger 60ms apart, main content
  // fades + slides up 8px. The `mounted` flag flips true on the first
  // frame after the shell is laid out; classes hang off it.
  let mounted = $state(false);

  // TitlebarThread reinforcement: the Thread draws under the wordmark on
  // every route change. We key this off `route` to retrigger the draw.
  let wordmarkThreadKey = $state(0);

  // Agent phase drives the DynamicIsland. Same logic as the legacy shell —
  // streaming wins, then halted, then consent, then idle/error.
  let agentPhase = $derived(
    conversation.isStreaming
      ? 'thinking'
      : halt.state.halted
        ? 'error'
        : consent.ticket
          ? 'consent'
          : daemon.connected
            ? 'idle'
            : 'error',
  );

  // Route → display label (mono-uppercase whisper per spec §2.1).
  const ROUTE_LABEL: Record<RouteId, string> = {
    chat: 'Chat',
    hub: 'Hub',
    skills: 'Skills',
    sync: 'Sync',
    audit: 'Audit',
    channels: 'Channels',
    delegation: 'Delegation',
    account: 'Account',
    settings: 'Settings',
    about: 'About',
  };
  let routeLabel = $derived(ROUTE_LABEL[route] ?? 'Chat');

  // Whether the overlay mode (frameless, no NavRail) is active. The
  // spec §2.7 has overlay hide the rail.
  let isOverlay = $derived(overlay.active);

  // ── Lifecycle ───────────────────────────────────────────────
  onMount(() => {
    // theme is applied before paint by main.ts. Read it back so the
    // toggle reflects the active mode.
    theme = (document.documentElement.dataset.mode as 'light' | 'dark') ?? 'light';

    // Beat-by-beat mount: 60ms cadence. The flag flips on the next frame
    // so the entrance classes (and per-NavRail-item delay) can hang off it.
    const frame = requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        mounted = true;
      });
    });

    try {
      initStores();
    } catch (e) {
      console.warn('initStores failed (no daemon/wails?)', e);
    }
    try {
      halt.startPolling();
    } catch (e) {
      console.warn('halt.startPolling failed', e);
    }
    try {
      overlay.start();
    } catch (e) {
      console.warn('overlay.start failed', e);
    }

    void Promise.all([
      ipc.firstRunStatus().catch(() => ({ complete: false })),
      ipc.onboardingIsComplete().catch(() => true),
    ])
      .then(([fr, oc]) => {
        const daemonComplete = !!(fr.complete && oc);
        let seen = false;
        try {
          seen = !!localStorage.getItem('condura-ritual-seen');
        } catch {
          /* ignore */
        }
        showOnboarding = !daemonComplete || !seen;
        if (!seen) {
          try {
            localStorage.setItem('condura-ritual-seen', '1');
          } catch {
            /* ignore */
          }
        }
      })
      .catch(() => {});

    // Hash router — updates the route derived state on every change.
    const onHash = () => {
      currentHash = window.location.hash || '#/';
      // Re-trigger the TitlebarThread reinforcement per spec §4.7.
      wordmarkThreadKey++;
    };
    window.addEventListener('hashchange', onHash);

    // v0.1.0 dev affordance: Shift+O re-opens the Ritual (remove before ship).
    // Full keyboard chord map per SCREEN_SHELL.md §5.1.
    let gArmed = false;
    let gArmedAt = 0;
    const onKey = (e: KeyboardEvent) => {
      const mod = e.metaKey || e.ctrlKey;
      const k = e.key.toLowerCase();

      // Hard kill switch (Layer 1) — independent of the focused surface.
      // Per spec §5.1: this is also wired to internal/conductor/killswitch.go
      // at the Go layer; the GUI never blocks it.
      if (mod && e.shiftKey && k === 'escape') {
        e.preventDefault();
        try {
          void halt.halt('hard_hotkey');
        } catch {
          /* ignore */
        }
        return;
      }

      // Esc dismisses the topmost overlay (consistency with MOAT §2.10).
      if (e.key === 'Escape' && !e.shiftKey && !mod) {
        if (themePickerOpen) {
          themePickerOpen = false;
          e.preventDefault();
          return;
        }
        if (paletteOpen) {
          paletteOpen = false;
          e.preventDefault();
          return;
        }
        if (quickOpen) {
          quickOpen = false;
          e.preventDefault();
          return;
        }
      }

      // Chord with modifier — palette, quick prompt, settings, history,
      // route jumps, new conversation.
      if (mod) {
        switch (k) {
          case 'k':
            e.preventDefault();
            paletteOpen = true;
            return;
          case 'p':
            if (e.shiftKey) {
              e.preventDefault();
              quickOpen = true;
              return;
            }
            break;
          case ',':
            e.preventDefault();
            window.location.hash = ROUTE_HASH.settings;
            return;
          case '[':
            e.preventDefault();
            history.back();
            return;
          case ']':
            e.preventDefault();
            history.forward();
            return;
          case 'n':
            e.preventDefault();
            try {
              conversation.clear();
            } catch {
              /* ignore */
            }
            return;
          case '0':
            e.preventDefault();
            window.location.hash = ROUTE_HASH.account;
            return;
        }
        // ⌘1-⌘9 route jumps. Order matches the NavRail's 10-route order in
        // SCREEN_NAVRAIL §2.4: chat, hub, skills, sync, audit, channels,
        // delegation, [⌘8 reserved], settings. ⌘0 is wired above to account.
        if (k >= '1' && k <= '9') {
          const idx = Number(k) - 1;
          const order: (RouteId | null)[] = [
            'chat',
            'hub',
            'skills',
            'sync',
            'audit',
            'channels',
            'delegation',
            null, // ⌘8 reserved per spec
            'settings',
          ];
          const target = order[idx];
          if (target) {
            e.preventDefault();
            window.location.hash = ROUTE_HASH[target];
            return;
          }
        }
      }

      // Shift+T → toggle theme (the keyboard twin of the titlebar button).
      if (e.shiftKey && !mod && k === 't') {
        e.preventDefault();
        cycleTheme();
        return;
      }

      // Shift+O → reopen Ritual (dev affordance).
      if (e.shiftKey && !mod && k === 'o') {
        e.preventDefault();
        showOnboarding = true;
        return;
      }

      // ? → Shortcuts surface (v0.1.0: open the CommandPalette as the
      // catch-all shortcut surface; a dedicated Shortcuts.svelte is a
      // Phase-5 follow-up per SCREEN_SHELL.md §9.1).
      if (!mod && e.key === '?' && !e.shiftKey) {
        const target = e.target as HTMLElement | null;
        const tag = target?.tagName;
        if (tag === 'INPUT' || tag === 'TEXTAREA' || target?.isContentEditable) {
          return;
        }
        e.preventDefault();
        paletteOpen = true;
        return;
      }

      // `g` then a letter — Go-to surface (per spec §5.1).
      if (gArmed && Date.now() - gArmedAt <= 1200) {
        const map: Record<string, RouteId> = {
          s: 'settings',
          h: 'hub',
          a: 'about',
          c: 'channels',
          k: 'skills',
          r: 'replay',
          l: 'sync',
          d: 'delegation',
        };
        if (map[k]) {
          e.preventDefault();
          window.location.hash = ROUTE_HASH[map[k]];
          gArmed = false;
          return;
        }
        if (k === ',') {
          e.preventDefault();
          window.location.hash = ROUTE_HASH.about;
          gArmed = false;
          return;
        }
        if (k === '?') {
          e.preventDefault();
          paletteOpen = true;
          gArmed = false;
          return;
        }
        gArmed = false;
      }
      if (!mod && !e.shiftKey && k === 'g') {
        gArmed = true;
        gArmedAt = Date.now();
      }
    };
    window.addEventListener('keydown', onKey);

    return () => {
      cancelAnimationFrame(frame);
      window.removeEventListener('hashchange', onHash);
      window.removeEventListener('keydown', onKey);
      try {
        consent.stop();
      } catch {
        /* ignore */
      }
      try {
        halt.stopPolling();
      } catch {
        /* ignore */
      }
      try {
        overlay.stop();
      } catch {
        /* ignore */
      }
    };
  });

  // ── Theme ───────────────────────────────────────────────────
  function cycleTheme(): void {
    // Two-state fast toggle (light ↔ dark). The full ThemePicker sheet
    // gives the user a third "system" state.
    const next = theme === 'light' ? 'dark' : 'light';
    theme = next;
    document.documentElement.dataset.mode = next;
    try {
      localStorage.setItem('condura-theme', next);
    } catch {
      /* ignore */
    }
    try {
      window.dispatchEvent(
        new CustomEvent('condura-theme-change', {
          detail: { theme: next },
          bubbles: true,
        }),
      );
    } catch {
      /* ignore */
    }
  }

  function openThemePicker(): void {
    themePickerOpen = true;
  }

  function closeThemePicker(): void {
    themePickerOpen = false;
  }

  // ── Navigation ──────────────────────────────────────────────
  function navigate(r: RouteId): void {
    window.location.hash = ROUTE_HASH[r];
  }

  function completeOnboarding(routeHash?: string): void {
    showOnboarding = false;
    if (routeHash) window.location.hash = routeHash;
  }

  function handleResume(): void {
    try {
      void halt.resume();
    } catch (e) {
      console.warn(e);
    }
  }

  // ── Dynamic status (StatusBar) ──────────────────────────────
  // The StatusBar renders the per-turn stopwatch, tokens, cost, model
  // badge. The daemon contract owns the math; we read from
  // `conversation` for the runtime values and from `daemon` for the
  // connectivity heartbeat.
  let stopwatchMs = $derived(conversation.turnElapsedMs ?? 0);
  let tokenIn = $derived(conversation.tokensIn ?? 0);
  let tokenOut = $derived(conversation.tokensOut ?? 0);
  let costUsd = $derived(conversation.lastCostUsd ?? null);
  let modelLabel = $derived(conversation.modelLabel ?? '—');
  let isStreaming = $derived(conversation.isStreaming);

  // StatusBar visibility: collapsed into titlebar at <1024px (per spec §1.4
  // and §2.5). The statusbar still renders at all viewport sizes in the
  // current single-breakpoint surface, but its computed visibility respects
  // the gap via `--rightrail-w`.
  let statusbarVisible = $derived(!isOverlay);

  // Right-rail content is contextual per route (spec §2.4). v0.1.0:
  // Chat shows a conversations list (read-only placeholder), other routes
  // show route-specific content. The right rail is only rendered when
  // the route declares one and the breakpoint is wide enough.
  let hasRightRail = $derived(
    route === 'chat' ||
      route === 'hub' ||
      route === 'skills' ||
      route === 'sync' ||
      route === 'audit' ||
      route === 'channels' ||
      route === 'delegation',
  );
</script>

<!-- The Shell is "the home of the Thread." When the Ritual owns the
     window (first-run), the Shell is unmounted entirely; on completion
     it dissolves and the Shell develops underneath. See spec §3.4. -->
{#if showOnboarding}
  <Ritual onComplete={completeOnboarding} />
{:else}
  <div
    class="shell surface-paper"
    class:mounted
    class:overlay={isOverlay}
    class:halted={halt.state.halted}
    class:consent-pending={consent.ticket !== null}
    class:streaming={isStreaming}
    data-mode={theme}
    data-route={route}
  >
    <!-- Grain overlay (spec §2.6). Always mounted; visibility per
         condura.css :root[data-energy=low] + reduced-motion media query. -->
    <div class="paper-grain" aria-hidden="true"></div>

    <!-- ── Titlebar (grid-area: titlebar) ─────────────────────── -->
    <header class="titlebar" role="banner" aria-label="Condura">
      <!-- L · Identity: wordmark + route context whisper. -->
      <div class="tb-l">
        <div class="wordmark">
          <span class="wordmark-text">Condura</span>
          <span class="wordmark-dot" aria-hidden="true"></span>
          <!-- TitlebarThread reinforcement: a horizontal Thread draws
               under the wordmark on every route change (spec §4.7 #1). -->
          {#key wordmarkThreadKey}
            <span class="wordmark-thread" aria-hidden="true">
              <Thread orientation="h" draw={true} glow={false} />
            </span>
          {/key}
        </div>
        <span class="tb-route" aria-live="polite">· {routeLabel}</span>
      </div>

      <!-- C · Thread: the absolute-positioned signature (drawn by
           TitlebarThread). Lives in C, never re-mounts. -->
      <TitlebarThread />

      <!-- R · Status & Controls (spec §2.3). -->
      <div class="tb-r">
        <div class="tb-island-wrap">
          <DynamicIsland
            phase={agentPhase}
            task={conversation.isStreaming ? conversation.currentTitle : ''}
          />
        </div>
        <div class="tb-controls" role="group" aria-label="Titlebar controls">
          <span class="kbd-hint" title="Command palette" aria-hidden="true">
            <kbd>⌘</kbd><kbd>K</kbd>
          </span>
          <button
            type="button"
            class="theme-toggle tactile"
            onclick={openThemePicker}
            aria-label={`Theme: ${theme === 'dark' ? 'dark' : 'light'}. Open theme picker.`}
            aria-haspopup="dialog"
            aria-expanded={themePickerOpen}
          >
            <Glyph name={theme === 'light' ? 'moon' : 'sun'} size={14} />
          </button>
        </div>
      </div>
    </header>

    <!-- ── NavRail (grid-area: navrail) — only when not in overlay. -->
    {#if !isOverlay}
      <div class="navrail-slot">
        <NavRail {route} onnavigate={navigate} />
      </div>
    {/if}

    <!-- ── Main: the active route (spec §2.3). `{#key route}` re-mounts
         the route component cleanly; `.route-enter` is the blur-in. -->
    <main
      class="surface main"
      role="main"
      aria-label={routeLabel}
    >
      {#key route}
        <div class="route-container route-enter">
          {#if route === 'chat'}
            <Chat />
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
        </div>
      {/key}
    </main>

    <!-- ── RightRail (grid-area: rightrail, contextual per spec §2.4).
         Only renders when the route declares one and we're not in
         overlay mode. Width 320px on ≥1440, collapses to 0 below. -->
    {#if !isOverlay && hasRightRail}
      <aside class="rightrail" aria-label="{routeLabel} details">
        <!-- Per-route right-rail content. v0.1.0 keeps the layout
             consistent: a rail eyebrow + an empty-state whisper when
             the route hasn't yet shipped its specific rail surface.
             (Routes own their own right-rail content; this is the
             slot. The full surface comes in route-by-route.) -->
        <div class="rr-eyebrow">· {routeLabel}</div>
        <div class="rr-body">
          <p class="rr-whisper">
            Details for {routeLabel.toLowerCase()} will surface here as you use it.
          </p>
        </div>
      </aside>
    {/if}

    <!-- ── StatusBar (grid-area: statusbar, spec §2.5). 28px tall,
         mono 11px, uppercase. Reads from conversation store; stopwatch
         ticks at 60ms cadence during streaming. -->
    {#if statusbarVisible}
      <footer class="statusbar" role="contentinfo" aria-label="Status">
        <div class="sb-stopwatch" aria-label="Turn elapsed">
          {formatStopwatch(stopwatchMs)}
        </div>
        <div class="sb-git" aria-label="Branch">
          <Glyph name="chevron-right" size={10} />
          <span>main</span>
        </div>
        <div class="sb-tokens" aria-label="Tokens">
          <span class="sb-tok-in">↓ {formatTokens(tokenIn)}</span>
          <span class="sb-tok-out">↑ {formatTokens(tokenOut)}</span>
        </div>
        <div class="sb-cost" aria-label="Cost">
          {costUsd !== null ? `~$${costUsd.toFixed(3)}` : '—'}
        </div>
        <div class="sb-model" aria-label="Model">
          {modelLabel}
        </div>
      </footer>
    {/if}
  </div>
{/if}

<!-- ── Global overlays (mounted at the root, per spec §6) ─────── -->
<ConsentModal />
{#if halt.state.halted}
  <KillSwitchOverlay
    reason={halt.state.reason ?? 'user requested'}
    onresume={handleResume}
  />
{/if}

<CommandPalette
  open={paletteOpen}
  onclose={() => (paletteOpen = false)}
  onnavigate={navigate}
/>
<QuickPromptOverlay
  open={quickOpen}
  onclose={() => (quickOpen = false)}
/>

<!-- ThemePicker — the dedicated phase-2 picker owns the palette
     switch choreography (clip-path circle reveal via View Transitions
     API). The Shell hosts the toggle button only; clicking it opens
     the picker sheet. The picker dispatches `condura-theme-change` on
     commit; Shell listens via the standard event for live sync. -->
<ThemePicker
  open={themePickerOpen}
  onclose={closeThemePicker}
  onapply={(t) => {
    // The ThemePicker has applied the mode; reflect it locally.
    const eff = t === 'system'
      ? (matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
      : (t as 'light' | 'dark');
    theme = eff;
    document.documentElement.dataset.mode = eff;
    try {
      localStorage.setItem('condura-theme', t);
    } catch {
      /* ignore */
    }
    closeThemePicker();
  }}
/>

<!-- ── Format helpers ────────────────────────────────────────── -->
<script context="module" lang="ts">
  // Format helpers live in module-scope so they don't re-create closures
  // on every render. They are pure, no side effects.
  function formatStopwatch(ms: number): string {
    const total = Math.max(0, Math.floor(ms));
    const minutes = Math.floor(total / 60_000);
    const seconds = Math.floor((total % 60_000) / 1000);
    const millis = total % 1000;
    return `${pad2(minutes)}:${pad2(seconds)}.${pad3(millis)}`;
  }

  function pad2(n: number): string {
    return n < 10 ? `0${n}` : String(n);
  }

  function pad3(n: number): string {
    if (n < 10) return `00${n}`;
    if (n < 100) return `0${n}`;
    return String(n);
  }

  function formatTokens(n: number): string {
    if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
    return String(n);
  }
</script>

<style>
  /* ── Shell grid (spec §1.1) ──────────────────────────────────
     3 columns × 3 rows. Titlebar + statusbar span columns. The
     NavRail + Main + (optional) RightRail split the middle row.
     Overlay mode collapses to a single column (NavRail hidden). */
  .shell {
    display: grid;
    grid-template-columns: var(--rail-w, 64px) minmax(0, 1fr) var(--rightrail-w, 320px);
    grid-template-rows: var(--titlebar-h, 44px) minmax(0, 1fr) var(--statusbar-h, 28px);
    grid-template-areas:
      'titlebar  titlebar   titlebar'
      'navrail   main       rightrail'
      'navrail   statusbar  statusbar';
    height: 100vh;
    height: 100dvh;
    width: 100vw;
    position: relative;
    overflow: hidden;
    background: var(--surface);
    isolation: isolate; /* grain stays inside this stacking context */
    opacity: 0;
    transform: translateY(8px);
    transition:
      opacity var(--dur-slow) var(--ease),
      transform var(--dur-slow) var(--ease),
      grid-template-columns var(--dur) var(--ease);
  }

  .shell.mounted {
    opacity: 1;
    transform: translateY(0);
  }

  /* Overlay mode: rail hidden, main fills. */
  .shell.overlay {
    grid-template-columns: minmax(0, 1fr) 0px;
  }

  /* Titlebar geometry — three-zone grid, spec §1.1. */
  .titlebar {
    grid-area: titlebar;
    display: grid;
    grid-template-columns: minmax(220px, auto) 1fr minmax(220px, auto);
    align-items: center;
    gap: var(--space-4);
    padding: 0 var(--space-5);
    padding-left: max(
      env(safe-area-inset-left, 0px),
      var(--wails-traffic-light-w, 78px)
    );
    padding-right: max(env(safe-area-inset-right, 0px), var(--space-5));
    position: relative;
    border-bottom: 1px solid var(--hair);
    -webkit-app-region: drag;
    app-region: drag;
    z-index: var(--z-sticky);
    background: var(--surface);
  }

  /* The R-zone controls claim clicks; everything else passes the
     drag region to the OS (spec §1.3). */
  .titlebar button,
  .titlebar .tb-island-wrap,
  .titlebar .tb-controls,
  .titlebar .kbd-hint {
    -webkit-app-region: no-drag;
    app-region: no-drag;
  }

  .tb-l {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
    min-width: 0;
    z-index: 2;
  }

  .wordmark {
    position: relative;
    display: inline-flex;
    align-items: baseline;
    gap: 3px;
    padding-bottom: 2px;
  }

  .wordmark-text {
    font-family: var(--font-display);
    font-size: 22px;
    line-height: 1;
    letter-spacing: var(--ls-display);
    color: var(--content);
    /* First-paint wordmark draw: the clip-path eases to identity. */
    clip-path: inset(0 100% 0 0);
    animation: wordmark-reveal var(--dur-cine) var(--ease) forwards;
  }

  .wordmark-dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: var(--pollen);
    box-shadow: 0 0 8px color-mix(in oklab, var(--pollen) 60%, transparent);
    opacity: 0;
    animation: wordmark-dot-in 240ms var(--ease) calc(var(--dur-cine) - 60ms) forwards;
    transform: translateY(-1px);
  }

  .wordmark-thread {
    position: absolute;
    left: 0;
    right: 8px;
    bottom: 0;
    height: 1px;
    pointer-events: none;
  }

  @keyframes wordmark-reveal {
    to {
      clip-path: inset(0 0 0 0);
    }
  }

  @keyframes wordmark-dot-in {
    to {
      opacity: 1;
    }
  }

  .tb-route {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: var(--ls-mono);
    text-transform: uppercase;
    color: var(--content-faint);
    white-space: nowrap;
  }

  /* R-zone cluster (spec §2.3). */
  .tb-r {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    z-index: 2;
    justify-self: end;
  }

  .tb-controls {
    margin-left: auto;
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  .kbd-hint {
    display: flex;
    align-items: center;
    gap: 4px;
    opacity: 0;
    animation: kbd-hint-fade 200ms var(--ease) 600ms forwards;
  }

  @keyframes kbd-hint-fade {
    to {
      opacity: 1;
    }
  }

  .kbd-hint kbd {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    padding: 2px 6px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-xs);
    background: var(--surface-card);
    color: var(--content-soft);
  }

  /* Theme toggle — pill, 28×28, surface-card fill, hairline border
     (spec §2.1). Hover rotates -12deg; press inherits .tactile
     (already in condura.css). Focus uses synapse ring only — this
     is a pill, so the inset line is dropped per DESIGNLANG §8.6. */
  .theme-toggle {
    width: 28px;
    height: 28px;
    border-radius: var(--r-pill);
    border: 1px solid var(--hair);
    background: var(--surface-card);
    color: var(--content-soft);
    display: grid;
    place-items: center;
    transition:
      transform var(--dur-fast) var(--ease),
      border-color var(--dur) var(--ease),
      color var(--dur) var(--ease);
  }

  .theme-toggle:hover {
    transform: rotate(-12deg);
    border-color: var(--hair-strong);
    color: var(--content);
  }

  .theme-toggle:focus-visible {
    outline: none;
    box-shadow: 0 0 0 2px var(--synapse);
  }

  /* NavRail slot — host the existing NavRail, applied via :global.
     The grid-area is set on the slot; the rail itself takes the
     full slot. */
  .navrail-slot {
    grid-area: navrail;
    position: relative;
    z-index: var(--z-sticky);
    border-right: 1px solid var(--hair);
    background: var(--surface);
  }

  /* Main surface (spec §1.3) — base z, route-local pointer events. */
  .main {
    grid-area: main;
    position: relative;
    overflow: hidden;
    z-index: var(--z-base);
  }

  /* Route container — fills Main, scrolls. The `.route-enter` class
     is the `blur-in` keyframe declared globally in condura.css. */
  .route-container {
    height: 100%;
    width: 100%;
    overflow: auto;
  }

  /* RightRail — contextually rendered, 320px wide. No glass (per
     MOAT §4.3); surface-card fill, never backdrop-filter. */
  .rightrail {
    grid-area: rightrail;
    background: var(--surface-card);
    border-left: 1px solid var(--hair);
    padding: var(--space-6) var(--space-5);
    overflow-y: auto;
    overflow-x: hidden;
    min-width: 0;
    z-index: var(--z-base);
  }

  .rr-eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: var(--ls-mono);
    text-transform: uppercase;
    color: var(--content-faint);
    margin-bottom: var(--space-3);
  }

  .rr-body {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .rr-whisper {
    font-size: 13px;
    line-height: 1.55;
    color: var(--content-soft);
    max-width: 32ch;
  }

  /* StatusBar (spec §2.5) — 28px, mono 11px uppercase, hairline
     border-top, no shadow. The four data points + model badge
     arrange left → right. */
  .statusbar {
    grid-area: statusbar;
    display: flex;
    align-items: center;
    gap: var(--space-5);
    padding: 0 var(--space-5);
    padding-right: max(env(safe-area-inset-right, 0px), var(--space-3));
    padding-bottom: env(safe-area-inset-bottom, 0);
    border-top: 1px solid var(--hair);
    background: var(--surface);
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--content-faint);
    z-index: var(--z-sticky);
  }

  .sb-stopwatch,
  .sb-git,
  .sb-tokens,
  .sb-cost,
  .sb-model {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  .sb-stopwatch {
    color: var(--content-mute);
    font-variant-numeric: tabular-nums;
  }

  .sb-git {
    color: var(--content-mute);
  }

  .sb-tokens {
    gap: var(--space-3);
  }

  .sb-tok-in,
  .sb-tok-out,
  .sb-cost,
  .sb-model {
    color: var(--content-mute);
  }

  .sb-model {
    margin-left: auto;
  }

  /* When halted, the StatusBar draws a danger hairline beneath it
     (spec §3.2). One element, one signal. */
  .shell.halted .statusbar {
    box-shadow: inset 0 -1px 0 0 var(--danger);
  }

  /* Stagger the NavRail item entrance: 60ms each, top-down (spec §4.1). */
  .shell.mounted :global(.rail-row) {
    opacity: 0;
    transform: translateX(-6px);
    animation: nav-item-in 220ms var(--ease) forwards;
  }

  .shell.mounted :global(.rail-row:nth-child(1))  { animation-delay: 200ms; }
  .shell.mounted :global(.rail-row:nth-child(2))  { animation-delay: 260ms; }
  .shell.mounted :global(.rail-row:nth-child(3))  { animation-delay: 320ms; }
  .shell.mounted :global(.rail-row:nth-child(4))  { animation-delay: 380ms; }
  .shell.mounted :global(.rail-row:nth-child(5))  { animation-delay: 440ms; }
  .shell.mounted :global(.rail-row:nth-child(6))  { animation-delay: 500ms; }
  .shell.mounted :global(.rail-row:nth-child(7))  { animation-delay: 560ms; }
  .shell.mounted :global(.rail-row:nth-child(8))  { animation-delay: 620ms; }
  .shell.mounted :global(.rail-row:nth-child(9))  { animation-delay: 680ms; }
  .shell.mounted :global(.rail-row:nth-child(10)) { animation-delay: 740ms; }

  @keyframes nav-item-in {
    to {
      opacity: 1;
      transform: translateX(0);
    }
  }
</style>
