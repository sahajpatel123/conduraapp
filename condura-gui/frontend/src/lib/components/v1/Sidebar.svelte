<!--
  Sidebar — the desktop app's table of contents.

  Not a nav rail. A curated index.

  Each route is presented as a chapter in a slow, considered book:
    01  Chat       Converse with the agent and tend its memory.
    02  Audit      A read-only ledger of every action taken.

  Numbers in mono, names in serif, descriptions in italic serif.
  Generous vertical rhythm. Hairlines, not fills. Plum dots, not pills.

  This file consumes ONLY existing semantic + primitive tokens.
-->
<script lang="ts">
  import Pulse from './Pulse.svelte';
  import BrandWordmark from './BrandWordmark.svelte';
  import type { PulseState } from '$tokens/motion';

  export type RouteId =
    | 'chat'
    | 'audit'
    | 'replay'
    | 'hub'
    | 'sync'
    | 'skills'
    | 'channels'
    | 'delegation'
    | 'settings'
    | 'about';

  interface RouteEntry {
    id: RouteId;
    /** The chapter title — set in Source Serif, calm weight. */
    name: string;
    /** The italic subtitle — one slow clause, never a sentence. */
    blurb: string;
  }

  interface Props {
    active?: RouteId;
    collapsed?: boolean;
    /** Mirrors the agent's vital sign. Defaults to idle. */
    pulseState?: PulseState;
    /** The italic line at the bottom. Defaults to a chapter footer. */
    footerLabel?: string;
    onnavigate?: (route: RouteId) => void;
    ontoggle?: () => void;
  }

  let {
    active = 'chat',
    collapsed = false,
    pulseState = 'idle',
    footerLabel = 'Listening, on the desktop.',
    onnavigate,
    ontoggle,
  }: Props = $props();

  /**
   * The index. Order is intentional — most-frequented routes first,
   * meta routes (settings, about) last, like a real book's TOC.
   */
  const ROUTES: ReadonlyArray<RouteEntry> = [
    { id: 'chat',       name: 'Chat',       blurb: 'Converse with the agent and tend its memory.' },
    { id: 'skills',     name: 'Skills',     blurb: 'Procedural knowledge, saved and recalled.' },
    { id: 'delegation', name: 'Delegation', blurb: 'Sub-agents, in parallel, under guard.' },
    { id: 'channels',   name: 'Channels',   blurb: 'Where the agent listens on your behalf.' },
    { id: 'sync',       name: 'Sync',       blurb: 'Device-to-device, end to end.' },
    { id: 'hub',        name: 'Hub',        blurb: 'Skills from the wider community.' },
    { id: 'audit',      name: 'Audit',      blurb: 'A read-only ledger of every action.' },
    { id: 'replay',     name: 'Replay',     blurb: 'The last twenty-four hours, scrubbable.' },
    { id: 'settings',   name: 'Settings',   blurb: 'Permissions, autonomy, the policy you write.' },
    { id: 'about',      name: 'About',      blurb: 'The mission, the manifest, the license.' },
  ];

  /** Two-digit numeral — 01, 02, … 10. Always two chars, so the column is tabular. */
  function chapterNumber(index: number): string {
    return String(index + 1).padStart(2, '0');
  }
</script>

<aside class="index" class:index--collapsed={collapsed} aria-label="Primary navigation">
  <!-- ─── Half-title: brand mark + italic volume line ─────────────── -->
  <header class="index__masthead">
    <div class="index__wordmark">
      <BrandWordmark variant={collapsed ? 'text-only' : 'default'} size="md" pulseSize="sm" pulseState={pulseState} />
    </div>
    {#if !collapsed}
      <p class="index__volume">An index of routes, kept by hand.</p>
    {/if}
  </header>

  <!-- A single rule beneath the masthead, like the line under a book title -->
  <hr class="index__rule" aria-hidden="true" />

  <!-- ─── Table of contents ────────────────────────────────────────── -->
  <nav class="index__toc" aria-label="Routes">
    {#each ROUTES as route, i (route.id)}
      {@const isActive = active === route.id}
      <button
        class="entry"
        class:entry--active={isActive}
        class:entry--collapsed={collapsed}
        type="button"
        onclick={() => onnavigate?.(route.id)}
<<<<<<< Updated upstream
        aria-label={route.label}
        aria-current={active === route.id ? 'page' : undefined}
        title={collapsed ? route.label : undefined}
=======
        aria-current={isActive ? 'page' : undefined}
        title={collapsed ? `${chapterNumber(i)} ${route.name}` : undefined}
>>>>>>> Stashed changes
      >
        <!-- The 1px plum hairline on the active row. Hairlines, not pills. -->
        <span class="entry__rule" aria-hidden="true"></span>

        <!-- The numeral — mono, tabular, always two characters. -->
        <span class="entry__number" aria-hidden="true">{chapterNumber(i)}</span>

        {#if !collapsed}
          <span class="entry__body">
            <span class="entry__name">{route.name}</span>
            <span class="entry__blurb">{route.blurb}</span>
          </span>
        {/if}
      </button>
    {/each}
  </nav>

  <!-- ─── Footer: a chapter footer, italic, set in serif ───────────── -->
  <footer class="index__foot">
    <hr class="index__rule" aria-hidden="true" />
    <p class="index__status">
      <span class="index__pulse" aria-hidden="true">
        <Pulse state={pulseState} size="sm" label={footerLabel} />
      </span>
      {#if !collapsed}
        <span class="index__caption">{footerLabel}</span>
      {/if}
    </p>
    {#if !collapsed}
      <button
        class="index__collapse"
        type="button"
        onclick={ontoggle}
        aria-label="Collapse the index"
        title="Collapse (⌘+\\)"
      >
        <span class="index__collapse-mark" aria-hidden="true">←</span>
        <span>Fold the page</span>
      </button>
    {:else}
      <button
        class="index__expand"
        type="button"
        onclick={ontoggle}
        aria-label="Expand the index"
        title="Expand (⌘+\\)"
      >
        <span aria-hidden="true">→</span>
      </button>
    {/if}
  </footer>
</aside>

<style>
  /* ─── The page ─────────────────────────────────────────────────── */

  .index {
    width: 268px;
    height: 100vh;
    display: flex;
    flex-direction: column;
    /* Surface matches the page — no raised card around the index.
       The book sits on the paper, not above it. */
    background-color: var(--surface-base);
    /* A hairline rule on the right edge, like the gutter of an open book. */
    border-right: 1px solid var(--border-subtle);
    /* The width transition is the "turning a page" gesture. */
    transition:
      width 420ms var(--ease-standard, cubic-bezier(0.4, 0, 0.2, 1));
    flex-shrink: 0;
    overflow: hidden;
    /* No overflow on the rail itself; the toc scrolls inside. */
  }

  .index--collapsed {
    width: 64px;
  }

  /* ─── Masthead ─────────────────────────────────────────────────── */

  .index__masthead {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-7) var(--space-5) var(--space-5);
    flex-shrink: 0;
  }

  .index__wordmark {
    display: flex;
    align-items: center;
    min-height: 22px;
  }

  .index__volume {
    margin: 0;
    font-family: var(--font-serif);
    font-style: italic;
    font-size: 12px;
    line-height: 1.4;
    color: var(--content-tertiary);
    letter-spacing: 0.005em;
  }

  /* ─── The hairline rules — used twice, kept consistent ─────────── */

  .index__rule {
    margin: 0;
    border: 0;
    border-top: 1px solid var(--border-subtle);
    flex-shrink: 0;
  }

  /* ─── Table of contents ───────────────────────────────────────── */

  .index__toc {
    display: flex;
    flex-direction: column;
    /* Generous: this is a book, not a list of buttons. */
    gap: 2px;
    padding: var(--space-5) var(--space-2) var(--space-5);
    flex: 1;
    overflow-y: auto;
    /* Subtle mask — the kind printers leave at the edges of a page. */
    mask-image: linear-gradient(
      to bottom,
      transparent 0,
      black var(--space-3),
      black calc(100% - var(--space-6)),
      transparent 100%
    );
    -webkit-mask-image: linear-gradient(
      to bottom,
      transparent 0,
      black var(--space-3),
      black calc(100% - var(--space-6)),
      transparent 100%
    );
  }

  /* ─── An entry — the smallest unit of the index ───────────────── */

  .entry {
    position: relative;
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-3) var(--space-3) var(--space-4);
    background: transparent;
    border: none;
    cursor: pointer;
    text-align: left;
    color: var(--content-secondary);
    border-radius: var(--radius-xs);
    /* No background change on hover — book-like restraint. */
    transition:
      color var(--duration-fast) var(--ease-standard, cubic-bezier(0.4, 0, 0.2, 1));
  }

  /* The plum hairline that marks an active entry.
     Always rendered, just hidden unless active. */
  .entry__rule {
    position: absolute;
    left: 0;
    top: var(--space-3);
    bottom: var(--space-3);
    width: 2px;
    background-color: transparent;
    border-radius: var(--radius-pill);
    transition:
      background-color var(--duration-base) var(--ease-standard, cubic-bezier(0.4, 0, 0.2, 1));
  }

  .entry--active .entry__rule {
    background-color: var(--content-accent);
  }

  /* The numeral column — monospace, tabular, never italic. */
  .entry__number {
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 400;
    line-height: 1;
    color: var(--content-muted);
    letter-spacing: 0.04em;
    /* Sit on the same baseline as the first line of the title. */
    padding-top: 5px;
    width: 18px;
    flex-shrink: 0;
    text-align: left;
    font-variant-numeric: tabular-nums;
    transition: color var(--duration-fast) var(--ease-standard, cubic-bezier(0.4, 0, 0.2, 1));
  }

  .entry--active .entry__number {
    /* The numeral brightens with the title — they are the same row. */
    color: var(--content-accent);
  }

  .entry__body {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
  }

  .entry__name {
    font-family: var(--font-serif);
    font-size: 16px;
    font-weight: 500;
    line-height: 1.25;
    color: var(--content-secondary);
    letter-spacing: -0.005em;
    /* A small plum dot appears just before the active name. */
    position: relative;
    transition:
      color var(--duration-base) var(--ease-standard, cubic-bezier(0.4, 0, 0.2, 1));
  }

  /* The active state: name in accent + a 3px dot to the left of the title. */
  .entry--active .entry__name {
    color: var(--content-accent);
  }
  .entry--active .entry__name::before {
    content: '';
    position: absolute;
    left: calc(-1 * var(--space-3));
    top: 50%;
    transform: translateY(-50%);
    width: 4px;
    height: 4px;
    border-radius: var(--radius-pill);
    background-color: var(--content-accent);
  }

  /* Hover — only the name shifts. No background, no pill. A book stays calm. */
  .entry:hover .entry__name {
    color: var(--content-primary);
  }

  .entry__blurb {
    font-family: var(--font-serif);
    font-style: italic;
    font-size: 12px;
    font-weight: 400;
    line-height: 1.45;
    color: var(--content-tertiary);
    letter-spacing: 0.005em;
    transition: opacity var(--duration-base) var(--ease-standard, cubic-bezier(0.4, 0, 0.2, 1));
  }

  .entry--active .entry__blurb {
    color: var(--content-secondary);
  }

  /* Focus — visible, but quiet. A plum outline that hugs the row. */
  .entry:focus-visible {
    outline: 2px solid var(--border-focus);
    outline-offset: 2px;
    border-radius: var(--radius-xs);
  }

  /* Collapsed state — just the numeral column. Title and blurb fade after
     the rail has begun to shrink, so the fold reads as a single motion. */
  .entry--collapsed .entry__number {
    /* Sit on the column's horizontal center. */
    padding-top: 0;
    line-height: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 22px;
  }

  /* ─── Footer — the colophon ────────────────────────────────────── */

  .index__foot {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-4) var(--space-4) var(--space-5);
    flex-shrink: 0;
  }

  .index__status {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin: 0;
    min-height: 22px;
  }

  .index__pulse {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    flex-shrink: 0;
  }

  .index__caption {
    font-family: var(--font-serif);
    font-style: italic;
    font-size: 12px;
    line-height: 1.4;
    color: var(--content-tertiary);
    letter-spacing: 0.005em;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    /* A calm fade when the page folds. */
    transition: opacity 240ms var(--ease-standard, cubic-bezier(0.4, 0, 0.2, 1)) 120ms;
  }

  .index--collapsed .index__caption {
    opacity: 0;
    transition-delay: 0ms;
  }

  /* The collapse / expand control — set in italic serif, the way a
     typesetter might set a small instruction at the foot of a page. */
  .index__collapse,
  .index__expand {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    background: transparent;
    border: none;
    padding: var(--space-1) 0;
    color: var(--content-muted);
    font-family: var(--font-serif);
    font-style: italic;
    font-size: 11.5px;
    letter-spacing: 0.01em;
    cursor: pointer;
    transition: color var(--duration-fast) var(--ease-standard, cubic-bezier(0.4, 0, 0.2, 1));
  }

  .index__collapse:hover,
  .index__expand:hover {
    color: var(--content-primary);
  }

  .index__collapse:focus-visible,
  .index__expand:focus-visible {
    outline: 2px solid var(--border-focus);
    outline-offset: 2px;
    border-radius: var(--radius-xs);
  }

  .index__expand {
    justify-content: center;
    width: 100%;
  }

  .index__collapse-mark {
    font-family: var(--font-mono);
    font-style: normal;
    font-size: 12px;
    color: var(--content-tertiary);
    transition: transform var(--duration-base) var(--ease-standard, cubic-bezier(0.4, 0, 0.2, 1));
  }

  .index__collapse:hover .index__collapse-mark {
    transform: translateX(-2px);
    color: var(--content-accent);
  }

  .index__expand:hover {
    transform: translateX(1px);
  }

  /* ─── Reduced motion: turn the page turn into a step ──────────── */
  @media (prefers-reduced-motion: reduce) {
    .index {
      transition: none;
    }
    .index__caption,
    .entry__name,
    .entry__blurb,
    .entry__number,
    .entry__rule,
    .index__collapse-mark {
      transition: none;
    }
  }
</style>