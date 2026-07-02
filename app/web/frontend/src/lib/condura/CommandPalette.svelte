<script lang="ts">
  // Condura CommandPalette — the ⌘K power-user + accessibility surface
  // (spec §19.2). Raycast-style: a paper panel over a blurred scrim, a mono
  // search input, and a list of commands. The signature interaction is a
  // sliding pollen highlight that springs between rows as focus travels.
  // Every motion is a verb: the sliding highlight = focus traveling; the
  // scrim blur = graduated attention. No decorative motion.

  import { tick } from 'svelte';
  import Glyph from './Glyph.svelte';
  import { type RouteId, ROUTE_HASH } from './NavRail.svelte';
  import { halt } from '../stores/halt.svelte';
  import { overlay } from '../stores/overlay.svelte';

  let {
    open,
    onclose,
    onnavigate,
  }: {
    open: boolean;
    onclose: () => void;
    onnavigate?: (r: RouteId) => void;
  } = $props();

  // ── command model ──
  type Command =
    | { kind: 'nav'; id: string; route: RouteId; label: string; icon: string }
    | { kind: 'action'; id: string; label: string; icon: string; run: () => void };

  // The 10 RouteIds become navigation commands (open skills / pair a device /
  // open audit / etc. are all just routes). Plus three actions.
  const NAV: { route: RouteId; label: string; icon: string }[] = [
    { route: 'chat', label: 'Go to Chat', icon: 'chat' },
    { route: 'hub', label: 'Open Skills Hub', icon: 'hub' },
    { route: 'skills', label: 'Open Skills', icon: 'skills' },
    { route: 'sync', label: 'Pair a Device', icon: 'sync' },
    { route: 'audit', label: 'Open Audit Log', icon: 'audit' },
    { route: 'replay', label: 'Open Action Replay', icon: 'replay' },
    { route: 'channels', label: 'Open Channels', icon: 'channels' },
    { route: 'delegation', label: 'Open Delegation', icon: 'delegation' },
    { route: 'settings', label: 'Open Settings', icon: 'settings' },
    { route: 'about', label: 'About Condura', icon: 'about' },
  ];

  // Theme is read off :root[data-mode]; toggling flips it. The icon shown is
  // the *destination* (moon → going to night, sun → going to day).
  let mode = $state<'light' | 'dark'>(
    (document.documentElement.getAttribute('data-mode') as 'light' | 'dark') ?? 'light',
  );
  function toggleTheme(): void {
    const next = mode === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-mode', next);
    mode = next;
  }
  function summonQuickPrompt(): void {
    overlay.show();
  }
  function stopEverything(): void {
    halt.halt('user from command palette').catch(() => {
      // daemon may be unreachable; the palette still closes
    });
  }

  // Full command list — re-derives when mode flips so the theme icon updates.
  let commands = $derived<Command[]>([
    ...NAV.map((n) => ({
      kind: 'nav' as const,
      id: `nav:${n.route}`,
      route: n.route,
      label: n.label,
      icon: n.icon,
    })),
    {
      kind: 'action' as const,
      id: 'action:theme',
      label: 'Toggle Theme',
      icon: mode === 'dark' ? 'sun' : 'moon',
      run: toggleTheme,
    },
    {
      kind: 'action' as const,
      id: 'action:summon',
      label: 'Summon Quick Prompt',
      icon: 'bolt',
      run: summonQuickPrompt,
    },
    {
      kind: 'action' as const,
      id: 'action:stop',
      label: 'Stop Everything',
      icon: 'stop',
      run: stopEverything,
    },
  ]);

  // ── query / filter / active row ──
  let query = $state('');
  let activeIndex = $state(0);
  let inputEl = $state<HTMLInputElement | null>(null);
  let rowEls = $state<(HTMLElement | null)[]>([]);

  // match-flash: when the active row changes, the right-side "↩" briefly
  // lights pollen — a key-press affordance, "press enter to run".
  let flashKey = $state(0);
  let prevActive = $state(-1);
  $effect(() => {
    if (activeIndex !== prevActive) {
      prevActive = activeIndex;
      flashKey++;
    }
  });

  // Fuzzy match on label / id / route. Simple includes is enough (spec).
  let filtered = $derived.by<Command[]>(() => {
    const q = query.trim().toLowerCase();
    if (!q) return commands;
    return commands.filter((c) =>
      c.label.toLowerCase().includes(q) ||
      c.id.toLowerCase().includes(q) ||
      (c.kind === 'nav' && c.route.toLowerCase().includes(q)),
    );
  });

  // The sliding highlight: a single absolutely-positioned element whose `top`
  // follows the active row's offsetTop. CSS transition (var(--ease)) glides it
  // between rows — that glide is focus traveling.
  let highlightTop = $derived.by<number>(() => {
    const el = rowEls[activeIndex];
    return el ? el.offsetTop : 0;
  });

  let activeRowId = $derived<string | undefined>(
    filtered.length ? `cmd-${activeIndex}` : undefined,
  );

  function runCommand(cmd: Command): void {
    if (cmd.kind === 'nav') {
      onnavigate?.(cmd.route);
      window.location.hash = ROUTE_HASH[cmd.route];
    } else {
      cmd.run();
    }
    onclose();
  }

  function onKeydown(e: KeyboardEvent): void {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      activeIndex = Math.min(activeIndex + 1, Math.max(0, filtered.length - 1));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      activeIndex = Math.max(activeIndex - 1, 0);
    } else if (e.key === 'Enter') {
      e.preventDefault();
      const cmd = filtered[activeIndex];
      if (cmd) runCommand(cmd);
    } else if (e.key === 'Escape') {
      e.preventDefault();
      onclose();
    }
  }

  // Reset + focus when the palette opens. Writing $state in an effect is safe
  // here because the effect only depends on `open` (no loop risk).
  $effect(() => {
    if (!open) return;
    query = '';
    activeIndex = 0;
    void tick().then(() => inputEl?.focus());
  });

  // Keep the active row within the scroll viewport as focus travels.
  $effect(() => {
    rowEls[activeIndex]?.scrollIntoView({ block: 'nearest' });
  });

  // Click on the scrim (not the panel) closes.
  function onScrimClick(e: MouseEvent): void {
    if (e.target === e.currentTarget) onclose();
  }
</script>

{#if open}
  <div class="scrim" onclick={onScrimClick} role="presentation">
    <div
      class="panel"
      role="dialog"
      aria-modal="true"
      aria-label="Command palette"
      tabindex="-1"
    >
      <div class="search-bar">
        <Glyph name="search" size={18} class="search-glyph" />
        <input
          class="search-input"
          type="text"
          placeholder="Type a command or search…"
          bind:this={inputEl}
          bind:value={query}
          oninput={() => (activeIndex = 0)}
          onkeydown={onKeydown}
          autocomplete="off"
          spellcheck="false"
          role="combobox"
          aria-expanded="true"
          aria-controls="cmd-list"
          aria-autocomplete="list"
          aria-activedescendant={activeRowId}
          aria-label="Search commands"
        />
        <button class="esc-chip" onclick={onclose} aria-label="Close command palette">
          <span class="esc-key">esc</span>
        </button>
      </div>

      <div class="rule-ink"></div>

      <div class="list" id="cmd-list" role="listbox" aria-label="Commands">
        {#if filtered.length}
          <div class="highlight" style="top: {highlightTop}px" aria-hidden="true"></div>
        {/if}
        {#each filtered as cmd, i (cmd.id)}
          <button
            class="row"
            class:active={i === activeIndex}
            id="cmd-{i}"
            role="option"
            aria-selected={i === activeIndex}
            bind:this={rowEls[i]}
            onclick={() => runCommand(cmd)}
            onmouseenter={() => (activeIndex = i)}
          >
            <Glyph name={cmd.icon} size={18} class="row-glyph" />
            <span class="row-label">{cmd.label}</span>
            {#key flashKey}
              <span
                class="row-hint"
                class:flash={i === activeIndex}
                aria-hidden="true"
              >↵</span>
            {/key}
          </button>
        {:else}
          <div class="empty">
            <span class="empty-head">Nothing matches this search.</span>
            <span class="empty-sub">No command for "{query.trim() || '—'}". Try a route name, "theme", or "stop".</span>
          </div>
        {/each}
      </div>

      <div class="rule-ink"></div>

      <div class="footer">
        <span class="foot-hint"><kbd>↑</kbd><kbd>↓</kbd> navigate</span>
        <span class="foot-hint"><kbd>↵</kbd> run</span>
        <span class="foot-hint"><kbd>esc</kbd> close</span>
      </div>
    </div>
  </div>
{/if}

<style>
  /* ── scrim: graduated attention + warm pollen bloom at the top ── */
  .scrim {
    position: fixed;
    inset: 0;
    z-index: var(--z-modal);
    display: grid;
    align-items: start;
    justify-items: center;
    padding-top: 16vh;
    background:
      radial-gradient(ellipse at 50% 0%, color-mix(in oklab, var(--pollen) 6%, transparent), transparent 60%),
      var(--scrim, color-mix(in oklab, var(--surface-ink) 42%, transparent));
    backdrop-filter: blur(8px) saturate(0.9);
    -webkit-backdrop-filter: blur(8px) saturate(0.9);
    animation: scrim-in var(--dur) var(--ease) both;
  }
  :global(:root[data-mode='dark']) .scrim {
    --scrim: color-mix(in oklab, var(--ink) 55%, transparent);
  }
  @keyframes scrim-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  /* ── paper panel: fade + scale + blur-in, with a synapse hairline that
     draws left→right across the top edge. The hairline uses a 1px tall
     ::before with scaleX(0→1) + transform-origin: left. ── */
  .panel {
    position: relative;
    width: min(560px, 92vw);
    max-height: 72vh;
    display: flex;
    flex-direction: column;
    background: var(--paper);
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-lg);
    box-shadow: var(--shadow-float);
    overflow: hidden;
    transform-origin: top center;
    animation: palette-in var(--dur-slow) var(--ease) both;
  }
  /* paper texture — 4px rounded svg grain (reuses the condura palette) */
  .panel::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: inherit;
    pointer-events: none;
    z-index: 0;
    opacity: var(--grain-opacity);
    mix-blend-mode: multiply;
    background-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240'><filter id='n'><feTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='2' stitchTiles='stitch'/><feColorMatrix values='0 0 0 0 0.08  0 0 0 0 0.07  0 0 0 0 0.04  0 0 0 0.06 0'/></filter><rect width='100%25' height='100%25' filter='url(%23n)'/></svg>");
    background-size: 240px 240px;
  }
  :global(:root[data-mode='dark']) .panel::after {
    mix-blend-mode: screen;
  }
  /* top-edge synapse hairline — draws left→right over 320ms (--dur-slow) */
  .panel::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 1px;
    background: var(--synapse);
    transform: scaleX(0);
    transform-origin: left;
    z-index: 4;
    pointer-events: none;
    animation: panel-draw var(--dur-slow) var(--ease) 60ms forwards;
  }
  @keyframes panel-draw {
    to { transform: scaleX(1); }
  }
  @keyframes palette-in {
    from { opacity: 0; transform: translateY(10px) scale(0.97); filter: blur(6px); }
    to { opacity: 1; transform: translateY(0) scale(1); filter: blur(0); }
  }

  /* ── search bar ── */
  .search-bar {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-4) var(--space-5);
  }
  .search-bar :global(.search-glyph) {
    color: var(--content-faint);
    flex: none;
  }
  .search-input {
    flex: 1;
    min-width: 0;
    font-family: var(--font-mono);
    font-size: 14px;
    letter-spacing: 0.02em;
    color: var(--content);
    background: transparent;
    border: none;
    outline: none;
    position: relative;
    /* center-out synapse hairline: a 1px ink under the input that draws
       from the middle outward over 240ms on focus. */
  }
  .search-input::placeholder {
    font-family: var(--font-display);
    font-style: italic;
    letter-spacing: -0.01em;
    color: var(--content-faint);
  }
  .search-input:focus {
    box-shadow: inset 0 -1px 0 var(--synapse);
    animation: input-draw 240ms var(--ease) both;
  }
  @keyframes input-draw {
    from { box-shadow: inset 0 -1px 0 transparent; }
    to { box-shadow: inset 0 -1px 0 var(--synapse); }
  }
  .esc-chip {
    display: inline-flex;
    align-items: center;
    padding: 2px 8px;
    border-radius: var(--r-xs);
    border: 1px solid var(--hair);
    color: var(--content-faint);
    font-family: var(--font-mono);
    transition:
      color var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .esc-chip:hover {
    color: var(--content);
    border-color: var(--hair-strong);
    background: color-mix(in oklab, var(--content) 6%, transparent);
    transform: translateY(-1px);
  }
  .esc-chip:active {
    transform: scale(0.96);
  }
  .esc-chip:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .esc-key {
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
  }

  /* ── list + sliding pollen highlight ── */
  .list {
    position: relative;
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    overscroll-behavior: contain;
    padding: var(--space-2);
  }

  /* ── sliding pollen highlight: tactile. 8% pollen background, 1px
     pollen hairline, 3px pollen left-border "you-are-here" bar. The
     spring-feel 200ms transition (--ease) glides between rows. ── */
  .highlight {
    position: absolute;
    left: var(--space-2);
    right: var(--space-2);
    height: var(--space-9);
    border-radius: var(--r-sm);
    background: color-mix(in srgb, var(--pollen) 8%, transparent);
    box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--pollen) 20%, transparent);
    border-left: 3px solid var(--pollen);
    transition: top 200ms var(--ease);
    pointer-events: none;
    z-index: 0;
  }

  .row {
    position: relative;
    z-index: 1;
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    height: var(--space-9);
    padding: 0 var(--space-4);
    border-radius: var(--r-sm);
    color: var(--content-mute);
    text-align: left;
    transition:
      color var(--dur) var(--ease),
      transform var(--dur) var(--ease);
  }
  .row:active {
    transform: scale(0.99);
  }
  .row:focus-visible {
    outline: none;
  }
  .row:focus-visible::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: inherit;
    box-shadow: 0 0 0 2px color-mix(in oklab, var(--pollen) 40%, transparent);
    pointer-events: none;
  }
  .row :global(.row-glyph) {
    color: var(--content-faint);
    flex: none;
    transition: color var(--dur) var(--ease);
  }
  .row-label {
    flex: 1;
    font-size: 14px;
    letter-spacing: -0.005em;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    transition: color var(--dur) var(--ease);
  }
  .row-hint {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--content-faint);
    flex: none;
    transition: color var(--dur) var(--ease);
  }
  /* match-flash: when the active row changes, the focused row's "↩" flashes
     pollen for 180ms — a key-press affordance, "press enter to run". */
  .row-hint.flash {
    animation: match-flash 180ms var(--ease);
  }
  @keyframes match-flash {
    0% { color: var(--pollen); }
    100% { color: var(--content-faint); }
  }
  /* active row: label shifts to synapse, glyph to pollen — focus has landed */
  .row.active {
    color: var(--content);
  }
  .row.active :global(.row-glyph) {
    color: var(--pollen);
  }
  .row.active .row-label {
    color: var(--synapse);
  }
  .row.active .row-hint {
    color: var(--content-mute);
  }

  .empty {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-1);
    padding: var(--space-5);
    color: var(--content-faint);
  }
  .empty-head {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 18px;
    line-height: 1.15;
    color: var(--content);
    letter-spacing: -0.01em;
  }
  .empty-sub {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 13px;
    line-height: 1.55;
    color: var(--content-faint);
    max-width: 48ch;
  }

  /* ── footer ── */
  .footer {
    display: flex;
    align-items: center;
    gap: var(--space-5);
    padding: var(--space-3) var(--space-5);
    color: var(--content-faint);
  }
  .foot-hint {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
  }
  .foot-hint kbd {
    font-family: var(--font-mono);
    font-size: 10px;
    padding: 1px 5px;
    border-radius: var(--r-xs);
    border: 1px solid var(--hair);
    color: var(--content-mute);
    background: var(--surface-card);
  }

  /* ── reduced motion: highlight jumps instantly, no blur-in, no scrim blur,
     no blooms, no hairline draw, no match flash, no panel thread-draw. ── */
  @media (prefers-reduced-motion: reduce) {
    .scrim {
      backdrop-filter: none;
      -webkit-backdrop-filter: none;
    }
    .panel::before { display: none; }
    .search-input:focus { box-shadow: none; }
    .row-hint.flash { animation: none; color: var(--content-faint); }
  }
</style>