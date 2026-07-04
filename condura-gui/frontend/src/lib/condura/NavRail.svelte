<!--
  Condura · NavRail — the left-edge vertical rail.

  Implemented to SCREEN_NAVRAIL.md (Phase 4). The contract:
    - Fixed 64px collapsed / 200px expanded (width animates var(--dur))
    - 10 routes in locked order:
        01 Chat (⌘1) · 02 Hub (⌘2) · 03 Skills (⌘3) · 04 Sync (⌘4) ·
        05 Audit (⌘5) · 06 Channels (⌘6) · 07 Delegation (⌘7) ·
        08 Account (⌘0) · 09 Settings (⌘9) · 10 About (no chord)
    - Persistent FLIP-animated synapse Thread segment 2px wide that
      moves to the active row's left edge over 320ms via `--ease`
    - Tooltips (one Tooltip wrapper per row) show route name + chord;
      400ms default hover delay
    - Roving tabindex: only the focused/active row has tabindex=0
    - Hover-expand reveals labels (slide-in 4px with 80ms delay)
    - Kill-switch at the bottom on its own hairline region (danger)
    - Status dots from `sync` / `audit` / `replay` / `pending` /
      `account` stores (no direct IPC)
    - Press feedback via scale(0.97) + brightness(0.95) (MOAT §2.2)
    - Rounded focus halo (MOAT §2.1): 2px synapse ring + 5px pollen halo

  Quiet by default. The only animation that announces active state is
  the Thread FLIP. Labels stay hidden until the user hovers or focuses.

  Shell.svelte owns the ⌘1-⌘9 / ⌘0 chord handler; this NavRail only
  contributes the Tab / Arrow / Home / End / Esc intra-rail nav.
-->
<script module lang="ts">
  export type RouteId =
    | 'chat'
    | 'hub'
    | 'skills'
    | 'sync'
    | 'audit'
    | 'channels'
    | 'delegation'
    | 'account'
    | 'settings'
    | 'about';

  export const ROUTE_HASH: Record<RouteId, string> = {
    chat: '#/',
    hub: '#/hub',
    skills: '#/skills',
    sync: '#/sync',
    audit: '#/audit',
    channels: '#/channels',
    delegation: '#/delegation',
    account: '#/account',
    settings: '#/settings',
    about: '#/about',
  };

  // startsWith matching so #/settings/legal still maps to settings (matches v1).
  export function hashToRoute(hash: string): RouteId {
    if (hash.startsWith('#/settings')) return 'settings';
    if (hash.startsWith('#/audit')) return 'audit';
    if (hash.startsWith('#/about')) return 'about';
    if (hash.startsWith('#/hub')) return 'hub';
    if (hash.startsWith('#/sync')) return 'sync';
    if (hash.startsWith('#/skills')) return 'skills';
    if (hash.startsWith('#/channels')) return 'channels';
    if (hash.startsWith('#/delegation')) return 'delegation';
    if (hash.startsWith('#/account')) return 'account';
    return 'chat';
  }
</script>

<script lang="ts">
  import { onMount, tick } from 'svelte';
  import Glyph from './Glyph.svelte';
  import Tooltip from './Tooltip.svelte';
  import { createFLIP } from './flip';
  import { halt } from '../stores/halt.svelte';
  import { sync } from '../stores/sync.svelte';
  import { audit } from '../stores/audit.svelte';
  import { replay } from '../stores/replay.svelte';
  import { account } from '../stores/account.svelte';
  import { pendingCount } from '../stores/pending.svelte';

  let {
    route,
    onnavigate,
  }: {
    route: RouteId;
    onnavigate: (r: RouteId) => void;
  } = $props();

  type BadgeTone =
    | 'ok'
    | 'info'
    | 'warn'
    | 'danger'
    | 'pollen'
    | 'synapse';
  interface BadgeInfo {
    tone: BadgeTone;
    label: string;
  }
  type Item = {
    id: RouteId;
    label: string;
    icon: string;
    chord: string;
    badge: () => BadgeInfo | null;
  };

  // Locked order per SCREEN_NAVRAIL §2.4 / §3.2.
  const ITEMS: Item[] = [
    {
      id: 'chat',
      label: 'Chat',
      icon: 'chat',
      chord: '⌘1',
      badge: () => null,
    },
    {
      id: 'hub',
      label: 'Hub',
      icon: 'hub',
      chord: '⌘2',
      badge: () => null,
    },
    {
      id: 'skills',
      label: 'Skills',
      icon: 'skills',
      chord: '⌘3',
      badge: () => null,
    },
    {
      id: 'sync',
      label: 'Sync',
      icon: 'sync',
      chord: '⌘4',
      badge: () => {
        const n = sync.pairs?.length ?? 0;
        return n > 0 ? { tone: 'info', label: `${n} paired` } : null;
      },
    },
    {
      id: 'audit',
      label: 'Audit',
      icon: 'audit',
      chord: '⌘5',
      badge: () => {
        if (replay.integrity && replay.integrity.valid === false) {
          return { tone: 'danger', label: 'chain broken' };
        }
        const pending = (audit.events ?? []).filter(
          (e) => e.verdict === 'prompt' || e.result === 'prompt',
        ).length;
        if (pending > 0) return { tone: 'warn', label: `${pending} unread` };
        return null;
      },
    },
    {
      id: 'channels',
      label: 'Channels',
      icon: 'channels',
      chord: '⌘6',
      badge: () => null,
    },
    {
      id: 'delegation',
      label: 'Delegation',
      icon: 'delegation',
      chord: '⌘7',
      badge: () => {
        const n = $pendingCount;
        return n > 0 ? { tone: 'pollen', label: `${n} pending` } : null;
      },
    },
    {
      id: 'account',
      label: 'Account',
      icon: 'account',
      chord: '⌘0',
      badge: () => {
        const s = account.status;
        if (s?.signed_in) {
          return { tone: 'synapse', label: s.display_name || s.email || 'signed in' };
        }
        return null;
      },
    },
    {
      id: 'settings',
      label: 'Settings',
      icon: 'settings',
      chord: '⌘9',
      badge: () => null,
    },
    {
      id: 'about',
      label: 'About',
      icon: 'about',
      chord: '',
      badge: () => null,
    },
  ];

  // ── DOM refs ─────────────────────────────────────────────
  let railEl = $state(null) as HTMLElement | null;
  let railRoutesEl = $state(null) as HTMLElement | null;
  let threadEl = $state(null) as HTMLElement | null;
  let rowEls: Record<RouteId, HTMLElement | null> = $state({
    chat: null,
    hub: null,
    skills: null,
    sync: null,
    audit: null,
    channels: null,
    delegation: null,
    account: null,
    settings: null,
    about: null,
  });
  let haltEl = $state(null) as HTMLElement | null;

  // ── Hover-expand state (200ms width + 80ms label delay) ──
  let expanded = $state(false);
  let expandTimer: ReturnType<typeof setTimeout> | null = null;
  let collapseTimer: ReturnType<typeof setTimeout> | null = null;

  function setExpanded(value: boolean, persist = false): void {
    if (expandTimer) {
      clearTimeout(expandTimer);
      expandTimer = null;
    }
    if (collapseTimer) {
      clearTimeout(collapseTimer);
      collapseTimer = null;
    }
    if (value || persist) {
      expanded = true;
    } else {
      // Grace period so the user can move between rows without the
      // width flickering.
      collapseTimer = setTimeout(() => {
        expanded = false;
      }, 160);
    }
  }

  // Drive the Shell's grid-template-columns via the shared --rail-w
  // CSS variable (per SCREEN_NAVRAIL §2.3 + §5.2 — the rail and the
  // main surface move together). The CSS variable lives on :root so
  // it cascades into the Shell's grid template.
  $effect(() => {
    if (typeof document === 'undefined') return;
    const railWidth = expanded ? '200px' : '64px';
    document.documentElement.style.setProperty('--rail-w', railWidth);
  });

  // ── Roving-tabindex (focused row gets tabindex=0) ─────────
  let focusedId = $state<RouteId | null>(null);
  $effect(() => {
    // When the active route changes (e.g., via chord handler in Shell),
    // bring focus with it.
    focusedId = route;
  });

  // ── FLIP — animate the persistent Thread segment ────────
  let flip = $state<ReturnType<typeof createFLIP> | null>(null);
  $effect(() => {
    if (threadEl) {
      flip = createFLIP(threadEl, 320);
    }
  });

  async function moveThreadTo(id: RouteId | null, animate = true): Promise<void> {
    if (!threadEl) return;
    if (!id) {
      threadEl.style.transition = animate
        ? 'opacity 200ms var(--ease)'
        : 'none';
      threadEl.style.opacity = '0';
      return;
    }
    const rowEl = rowEls[id];
    if (!rowEl) return;
    await tick();
    const railRect = railEl?.getBoundingClientRect();
    const rowRect = rowEl.getBoundingClientRect();
    if (!railRect) return;
    const top = rowRect.top - railRect.top + (rowRect.height - 24) / 2;
    const height = 24;
    if (!flip || !animate) {
      threadEl.style.transition = 'none';
      threadEl.style.transform = '';
      threadEl.style.top = `${top}px`;
      threadEl.style.height = `${height}px`;
      threadEl.style.opacity = '1';
      return;
    }
    flip.capture();
    threadEl.style.opacity = '1';
    flip.apply(top, height);
    flip.play();
  }

  // Move thread when route changes (and on first mount).
  $effect(() => {
    if (!threadEl) return;
    const activeId: RouteId | null =
      (ITEMS.find((i) => i.id === route)?.id as RouteId) ?? null;
    void moveThreadTo(activeId, true);
  });

  // ── Mouse handlers for hover-expand ──────────────────────
  function onRailEnter(): void {
    setExpanded(true);
  }
  function onRailLeave(): void {
    setExpanded(false, false);
  }

  // ── Row click — Shell handles navigation via `onnavigate` ──
  function onRowClick(id: RouteId): void {
    onnavigate(id);
  }

  // ── Kill-switch ──────────────────────────────────────────
  async function onHalt(): Promise<void> {
    try {
      await halt.halt('rail_button');
    } catch (e) {
      console.warn('halt.halt failed', e);
    }
  }

  // ── Keyboard nav (on the rail) — SCREEN_NAVRAIL §6.2 ─────
  function onRailKeydown(e: KeyboardEvent): void {
    const target = e.target as HTMLElement;
    const which = target.dataset.route as RouteId | undefined;
    if (!which && target.dataset.kind !== 'halt') return;

    const isItem = target.dataset.kind === undefined;
    const order: RouteId[] = ITEMS.map((i) => i.id);
    const idx = isItem ? order.indexOf(which as RouteId) : order.length;

    const move = (next: RouteId | 'halt'): void => {
      e.preventDefault();
      const el = next === 'halt' ? haltEl : rowEls[next];
      if (el) {
        focusedId = next;
        el.focus();
      }
    };

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        if (isItem && idx === order.length - 1) move(order[0]);
        else if (isItem) move(order[idx + 1]);
        else if (target.dataset.kind === 'halt') move(order[0]);
        return;
      case 'ArrowUp':
        e.preventDefault();
        if (isItem && idx === 0) move('halt');
        else if (isItem) move(order[idx - 1]);
        else if (target.dataset.kind === 'halt') move(order[order.length - 1]);
        return;
      case 'Home':
        e.preventDefault();
        move(order[0]);
        return;
      case 'End':
        e.preventDefault();
        move(target.dataset.kind === 'halt' ? order[order.length - 1] : 'halt');
        return;
      case 'Escape':
        e.preventDefault();
        setExpanded(false, false);
        (target as HTMLElement).blur();
        return;
      case 'Enter':
      case ' ':
        e.preventDefault();
        if (isItem) onRowClick(which as RouteId);
        else void onHalt();
        return;
    }
  }

  // ── Doc focus tracking: collapse + hide tooltips when focus
  //    leaves the rail entirely. Tooltips internally clear on blur. ──
  function onRailFocusOut(e: FocusEvent): void {
    const next = e.relatedTarget as HTMLElement | null;
    if (next && railEl && railEl.contains(next)) return;
    setExpanded(false, false);
  }

  // ── Bind each row's <button> by id (Svelte 5 callback ref) ──
  function bindRow(id: RouteId) {
    return (node: HTMLElement) => {
      rowEls[id] = node;
    };
  }

  // ── Mount: prime the thread at the active row ────────────
  onMount(() => {
    void moveThreadTo(route, false);
  });
</script>

<!--
  The <nav> has hover/focus listeners — these drive the visual expand and
  collapse of the rail (per SCREEN_NAVRAIL §2.3). Svelte's a11y rule
  discourages mouse/keyboard listeners on a non-interactive landmark,
  but the expansion is a presentational affordance; the actual
  interaction happens on each row's <button> child (which IS
  interactive). Suppress the rule here.
-->
<!-- svelte-ignore a11y_no_noninteractive_element_to_interactive_role -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<nav
  bind:this={railEl}
  class="rail"
  class:expanded
  aria-label="Primary"
  onmouseenter={onRailEnter}
  onmouseleave={onRailLeave}
  onfocusin={() => setExpanded(true)}
  onfocusout={onRailFocusOut}
  onkeydown={onRailKeydown}
  data-route={route}
>
  <div class="rail-routes" bind:this={railRoutesEl} role="presentation">
    {#each ITEMS as item (item.id)}
      {@const isActive = route === item.id}
      {@const tabIndex = isActive || focusedId === item.id ? 0 : -1}
      {@const badgeInfo = item.badge()}
      <Tooltip
        label={item.label}
        chord={item.chord}
        placement="right"
      >
        <button
          use:bindRow={item.id}
          type="button"
          role="link"
          class="rail-row"
          class:active={isActive}
          data-route={item.id}
          data-kind="item"
          tabindex={tabIndex}
          aria-current={isActive ? 'page' : undefined}
          aria-label={item.chord
            ? `${item.label}, command ${item.chord}`
            : item.label}
          onclick={() => onRowClick(item.id)}
        >
          <span class="rail-row-icon" aria-hidden="true">
            <Glyph name={item.icon} size={20} stroke={1.5} />
          </span>
          {#if badgeInfo}
            <span
              class="rail-row-badge"
              data-tone={badgeInfo.tone}
              aria-hidden="true"
            ></span>
          {/if}
          <span class="rail-row-label" aria-hidden="true">{item.label}</span>
          {#if item.chord}
            <kbd class="rail-row-chord" aria-hidden="true">{item.chord}</kbd>
          {/if}
        </button>
      </Tooltip>
    {/each}
  </div>

  <!-- one 1px hairline divider between routes and the kill switch -->
  <div class="rail-divider" aria-hidden="true"></div>

  <Tooltip label="Halt the agent" placement="right">
    <button
      bind:this={haltEl}
      type="button"
      class="rail-halt"
      data-kind="halt"
      tabindex={focusedId === 'about' ? 0 : -1}
      aria-label="Halt the agent"
      onclick={onHalt}
      onkeydown={onRailKeydown}
    >
      <span class="rail-row-icon" aria-hidden="true">
        <Glyph name="kill-switch" size={20} stroke={1.5} />
      </span>
      <span class="rail-row-label" aria-hidden="true">Halt</span>
    </button>
  </Tooltip>

  <!-- The persistent synapse Thread segment, FLIP-animated -->
  <div
    bind:this={threadEl}
    class="rail-thread"
    aria-hidden="true"
  >
    <span class="rail-thread-line"></span>
    <span class="rail-thread-glow" aria-hidden="true"></span>
  </div>
</nav>

<style>
  /* ── Geometry (SCREEN_NAVRAIL §2.1–§2.5) ──────────────── */
  .rail {
    width: 64px;
    transition: width var(--dur) var(--ease);
    display: flex;
    flex-direction: column;
    padding: var(--space-3) 0 var(--space-3);
    border-right: 1px solid var(--hair);
    position: relative;
    background: var(--surface);
    overflow: visible;
    color: var(--content);
  }
  .rail.expanded {
    width: 200px;
  }

  .rail-routes {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .rail-divider {
    height: 1px;
    margin: var(--space-3) var(--space-3);
    background: var(--hair);
  }

  /* ── Each row (44px tall, fills rail width) ─────────── */
  .rail-row,
  .rail-halt {
    appearance: none;
    background: transparent;
    border: 0;
    color: var(--content-mute);
    cursor: pointer;
    text-align: left;
    height: 44px;
    width: 100%;
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 0 12px;
    border-radius: var(--r-sm);
    position: relative;
    font-family: var(--font-sans);
    font-size: 13px;
    letter-spacing: -0.005em;
    transition:
      color var(--dur) var(--ease),
      background-color var(--dur) var(--ease);
    -webkit-app-region: no-drag;
    app-region: no-drag;
  }

  .rail-row:hover,
  .rail-halt:hover {
    color: var(--content);
  }

  /* The icon container — fixed 20px wide so the geometry holds while
     the row expands. */
  .rail-row-icon {
    width: 20px;
    height: 20px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex: none;
    opacity: 0.65;
    transition: opacity var(--dur) var(--ease), color var(--dur) var(--ease);
  }
  .rail-row:hover .rail-row-icon,
  .rail-row.active .rail-row-icon,
  .rail-halt:hover .rail-row-icon,
  .rail-halt:focus-visible .rail-row-icon {
    opacity: 1;
  }
  .rail-row.active .rail-row-icon {
    color: var(--synapse);
  }

  /* ── Active state fill — subtle ink wash, no scale ── */
  .rail-row.active {
    background: var(--surface-card);
    color: var(--content);
  }

  /* ── Label + chord: hidden collapsed, slide in on hover-expand ── */
  .rail-row-label,
  .rail-row-chord {
    opacity: 0;
    transform: translateX(-4px);
    transition:
      opacity 160ms var(--ease) 80ms,
      transform 160ms var(--ease) 80ms;
    pointer-events: none;
  }
  .rail-row-chord {
    margin-left: auto;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--content-faint);
    background: var(--surface-card);
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-xs);
    padding: 2px 6px;
    transition-delay: 120ms;
  }
  .rail.expanded .rail-row-label,
  .rail.expanded .rail-row-chord,
  .rail-halt:hover .rail-row-label,
  .rail-halt:focus-visible .rail-row-label {
    opacity: 1;
    transform: translateX(0);
  }

  /* ── Focus halo (rounded, MOAT §2.1) ────────────────── */
  .rail-row:focus-visible,
  .rail-halt:focus-visible {
    outline: none;
    box-shadow: 0 0 0 2px var(--synapse), 0 0 0 5px var(--pollen-halo-color);
  }

  /* ── Press (MOAT §2.2 — brightness + 0.5px settle) ──── */
  .rail-row:active:not([disabled]),
  .rail-halt:active:not([disabled]) {
    transform: scale(0.97);
    filter: brightness(0.95) saturate(1.1);
    translate: 0 0.5px;
  }

  /* ── Halt (kill switch) — separate region, danger on hover/focus ── */
  .rail-halt {
    margin: 0 var(--space-3);
    height: 44px;
  }
  .rail-halt .rail-row-icon {
    color: var(--content-mute);
  }
  .rail-halt:hover,
  .rail-halt:hover .rail-row-icon,
  .rail-halt:focus-visible,
  .rail-halt:focus-visible .rail-row-icon {
    color: var(--danger);
  }

  /* ── Status badge dots (6px circle with 1.5px outer halo) ── */
  .rail-row-badge {
    position: absolute;
    top: 6px;
    left: 18px;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    z-index: 2;
  }
  .rail-row-badge[data-tone='ok'] {
    background: var(--ok);
    box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--ok) 25%, transparent);
    animation: badge-breath 1.6s var(--ease) infinite;
  }
  .rail-row-badge[data-tone='synapse'] {
    background: var(--synapse);
    box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--synapse) 25%, transparent);
    animation: badge-breath 1.6s var(--ease) infinite;
  }
  .rail-row-badge[data-tone='warn'] {
    background: var(--warn);
    box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--warn) 25%, transparent);
    animation: badge-warn 1.4s var(--ease) infinite;
  }
  .rail-row-badge[data-tone='danger'] {
    background: var(--danger);
    box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--danger) 25%, transparent);
    animation: badge-warn 1.4s var(--ease) infinite;
  }
  .rail-row-badge[data-tone='info'] {
    background: var(--info);
    box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--info) 25%, transparent);
  }
  .rail-row-badge[data-tone='pollen'] {
    background: var(--pollen);
    box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--pollen) 25%, transparent);
  }

  @keyframes badge-breath {
    0%, 100% { transform: scale(1); }
    50%      { transform: scale(1.15); }
  }
  @keyframes badge-warn {
    0%, 100% { transform: scale(1); }
    50%      { transform: scale(1.18); box-shadow: 0 0 0 2px color-mix(in oklab, var(--warn, var(--danger)) 35%, transparent); }
  }

  /* ── Thread (the persistent active-state segment) ─── */
  .rail-thread {
    position: absolute;
    left: 0;
    width: 2px;
    top: var(--space-3);
    height: 44px;
    pointer-events: none;
    z-index: 3;
    opacity: 0;
    will-change: transform, top, height;
  }
  .rail-thread-line {
    position: absolute;
    inset: 0;
    background: var(--synapse);
    border-radius: 1px;
  }
  .rail-thread-glow {
    position: absolute;
    inset: -3px;
    background: var(--synapse-glow);
    border-radius: 1px;
    opacity: 0.4;
    filter: blur(3px);
    z-index: -1;
  }

  /* ── Reduced-motion contract (MOAT §2.3) ──────────── */
  /* The global condura.css rule already kills animation/transition.
     Locally we ensure the label reveal falls back to a fast fade, and
     the badge breath stops spinning. */
  @media (prefers-reduced-motion: reduce) {
    .rail-row-label,
    .rail-row-chord {
      transition: opacity 80ms linear;
      transition-delay: 0ms;
    }
    .rail-row-badge[data-tone='ok'],
    .rail-row-badge[data-tone='synapse'],
    .rail-row-badge[data-tone='warn'],
    .rail-row-badge[data-tone='danger'] {
      animation: none;
    }
  }
</style>
