<script lang="ts">
  /**
   * NavOrbit — Living Paper primary navigation.
   *
   * Synapse Spine: glyphs + FLIP thread + hover-expand labels, wired to
   * the same route contract as NavRail.svelte (SCREEN_NAVRAIL). Tokens use
   * the Living Paper palette (--lp-*).
   */
  import { onMount, tick } from 'svelte'
  import Glyph from '$lib/condura/Glyph.svelte'
  import Tooltip from '$lib/condura/Tooltip.svelte'
  import { createFLIP } from '$lib/condura/flip'
  import { halt } from '../stores/halt.svelte'
  import { sync } from '../stores/sync.svelte'
  import { audit } from '../stores/audit.svelte'
  import { replay } from '../stores/replay.svelte'
  import { account } from '../stores/account.svelte'
  import { pendingCount } from '../stores/pending.svelte'
  import { type RouteId } from '$lib/condura/NavRail.svelte'

  interface Props {
    route: RouteId
    onnavigate: (r: RouteId) => void
    /** When set, overrides which row shows active (e.g. replay has no rail entry). */
    activeRoute?: RouteId | null
  }

  let { route, onnavigate, activeRoute = undefined }: Props = $props()

  const highlightedRoute = $derived(
    activeRoute === undefined ? route : activeRoute
  )

  type BadgeTone = 'ok' | 'info' | 'warn' | 'danger' | 'pollen' | 'synapse'
  interface BadgeInfo {
    tone: BadgeTone
    label: string
  }

  type Item = {
    id: RouteId
    label: string
    icon: string
    chord: string
    hint: string
    badge: () => BadgeInfo | null
  }

  const ITEMS: Item[] = [
    {
      id: 'chat',
      label: 'Chat',
      icon: 'chat',
      chord: '⌘1',
      hint: 'Talk to Condura.',
      badge: () => null,
    },
    {
      id: 'hub',
      label: 'Hub',
      icon: 'hub',
      chord: '⌘2',
      hint: 'Browse the public Skills Hub.',
      badge: () => null,
    },
    {
      id: 'skills',
      label: 'Skills',
      icon: 'skills',
      chord: '⌘3',
      hint: 'Local installed procedures.',
      badge: () => null,
    },
    {
      id: 'sync',
      label: 'Sync',
      icon: 'sync',
      chord: '⌘4',
      hint: 'Pair a device.',
      badge: () => {
        const n = sync.pairs?.length ?? 0
        return n > 0 ? { tone: 'info', label: `${n} paired` } : null
      },
    },
    {
      id: 'audit',
      label: 'Audit',
      icon: 'audit',
      chord: '⌘5',
      hint: 'Every action, every model.',
      badge: () => {
        if (replay.integrity && replay.integrity.valid === false) {
          return { tone: 'danger', label: 'chain broken' }
        }
        const pending = (audit.events ?? []).filter(
          (e) => e.verdict === 'prompt' || e.result === 'prompt',
        ).length
        if (pending > 0) return { tone: 'warn', label: `${pending} unread` }
        return null
      },
    },
    {
      id: 'channels',
      label: 'Channels',
      icon: 'channels',
      chord: '⌘6',
      hint: 'Telegram, more soon.',
      badge: () => null,
    },
    {
      id: 'delegation',
      label: 'Delegation',
      icon: 'delegation',
      chord: '⌘7',
      hint: 'Sub-agents in flight.',
      badge: () => {
        const n = $pendingCount
        return n > 0 ? { tone: 'pollen', label: `${n} pending` } : null
      },
    },
    {
      id: 'account',
      label: 'Account',
      icon: 'account',
      chord: '⌘0',
      hint: 'Sign in for Hub, donations, support.',
      badge: () => {
        const s = account.status
        if (s?.signed_in) {
          return { tone: 'synapse', label: s.display_name || s.email || 'signed in' }
        }
        return null
      },
    },
    {
      id: 'settings',
      label: 'Settings',
      icon: 'settings',
      chord: '⌘9',
      hint: 'Power · autonomy · appearance · voice.',
      badge: () => null,
    },
    {
      id: 'about',
      label: 'About',
      icon: 'about',
      chord: '',
      hint: 'Colophon · the 7 invariants.',
      badge: () => null,
    },
  ]

  let railEl = $state<HTMLElement | null>(null)
  let threadEl = $state<HTMLElement | null>(null)
  let rowEls = $state<Record<RouteId, HTMLElement | null>>({
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
  })
  let haltEl = $state<HTMLElement | null>(null)

  let expanded = $state(false)
  let expandTimer: ReturnType<typeof setTimeout> | null = null
  let collapseTimer: ReturnType<typeof setTimeout> | null = null
  let focusedId = $state<RouteId | null>(null)
  let flip = $state<ReturnType<typeof createFLIP> | null>(null)

  function setExpanded(value: boolean, persist = false): void {
    if (expandTimer) {
      clearTimeout(expandTimer)
      expandTimer = null
    }
    if (collapseTimer) {
      clearTimeout(collapseTimer)
      collapseTimer = null
    }
    if (value || persist) {
      expanded = true
    } else {
      collapseTimer = setTimeout(() => {
        expanded = false
      }, 160)
    }
  }

  $effect(() => {
    if (typeof document === 'undefined') return
    document.documentElement.style.setProperty('--lp-nav-w', expanded ? '208px' : '56px')
  })

  $effect(() => {
    focusedId = route
  })

  $effect(() => {
    if (threadEl) {
      flip = createFLIP(threadEl, 320, { easing: 'var(--lp-ease-thread)' })
    }
  })

  async function moveThreadTo(id: RouteId | null, animate = true): Promise<void> {
    if (!threadEl) return
    if (!id) {
      threadEl.style.transition = animate ? 'opacity 200ms var(--lp-ease-thread)' : 'none'
      threadEl.style.opacity = '0'
      return
    }
    const rowEl = rowEls[id]
    if (!rowEl || !railEl) return
    await tick()
    const railRect = railEl.getBoundingClientRect()
    const rowRect = rowEl.getBoundingClientRect()
    const top = rowRect.top - railRect.top + (rowRect.height - 24) / 2
    const height = 24
    if (!flip || !animate) {
      threadEl.style.transition = 'none'
      threadEl.style.transform = ''
      threadEl.style.top = `${top}px`
      threadEl.style.height = `${height}px`
      threadEl.style.opacity = '1'
      return
    }
    flip.capture()
    threadEl.style.opacity = '1'
    flip.apply(top, height)
    flip.play()
  }

  $effect(() => {
    if (!threadEl) return
    const activeId = highlightedRoute
      ? (ITEMS.find((i) => i.id === highlightedRoute)?.id ?? null)
      : null
    void moveThreadTo(activeId, true)
  })

  function onRailEnter(): void {
    setExpanded(true)
  }

  function onRailLeave(): void {
    setExpanded(false, false)
  }

  function onRowClick(id: RouteId): void {
    onnavigate(id)
  }

  async function onHalt(): Promise<void> {
    try {
      await halt.halt('rail_button')
    } catch (e) {
      console.warn('halt.halt failed', e)
    }
  }

  function onRailKeydown(e: KeyboardEvent): void {
    const target = e.target as HTMLElement
    const which = target.dataset.route as RouteId | undefined
    if (!which && target.dataset.kind !== 'halt') return

    const isItem = target.dataset.kind === undefined
    const order: RouteId[] = ITEMS.map((i) => i.id)
    const idx = isItem ? order.indexOf(which as RouteId) : order.length

    const move = (next: RouteId | 'halt'): void => {
      e.preventDefault()
      const el = next === 'halt' ? haltEl : rowEls[next]
      if (el) {
        focusedId = next === 'halt' ? 'about' : next
        el.focus()
      }
    }

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        if (isItem && idx === order.length - 1) move(order[0])
        else if (isItem) move(order[idx + 1])
        else if (target.dataset.kind === 'halt') move(order[0])
        return
      case 'ArrowUp':
        e.preventDefault()
        if (isItem && idx === 0) move('halt')
        else if (isItem) move(order[idx - 1])
        else if (target.dataset.kind === 'halt') move(order[order.length - 1])
        return
      case 'Home':
        e.preventDefault()
        move(order[0])
        return
      case 'End':
        e.preventDefault()
        move(target.dataset.kind === 'halt' ? order[order.length - 1] : 'halt')
        return
      case 'Escape':
        e.preventDefault()
        setExpanded(false, false)
        target.blur()
        return
      case 'Enter':
      case ' ':
        e.preventDefault()
        if (isItem) onRowClick(which as RouteId)
        else void onHalt()
        return
    }
  }

  function onRailFocusOut(e: FocusEvent): void {
    const next = e.relatedTarget as HTMLElement | null
    if (next && railEl && railEl.contains(next)) return
    setExpanded(false, false)
  }

  function bindRow(node: HTMLElement, id: RouteId) {
    rowEls[id] = node
    return {
      destroy() {
        if (rowEls[id] === node) rowEls[id] = null
      },
    }
  }

  onMount(() => {
    void moveThreadTo(route, false)
  })
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<nav
  bind:this={railEl}
  class="lp lp-nav-orbit lp-grain"
  class:lp-nav-orbit--expanded={expanded}
  aria-label="Primary navigation"
  onmouseenter={onRailEnter}
  onmouseleave={onRailLeave}
  onfocusin={() => setExpanded(true)}
  onfocusout={onRailFocusOut}
  onkeydown={onRailKeydown}
  data-route={route}
>
  <header class="lp-nav-brand" aria-hidden="true">
    <span class="lp-nav-brand-mark">C</span>
    <span class="lp-nav-brand-label">Condura</span>
  </header>

  <div class="lp-nav-routes" role="presentation">
    {#each ITEMS as item (item.id)}
      {@const isActive = highlightedRoute === item.id}
      {@const tabIndex = isActive || focusedId === item.id ? 0 : -1}
      {@const badgeInfo = item.badge()}
      <Tooltip label={item.label} chord={item.chord || undefined} placement="right">
        <button
          use:bindRow={item.id}
          type="button"
          role="link"
          class="lp-nav-row lp-focus"
          class:lp-nav-row--active={isActive}
          data-route={item.id}
          data-kind="item"
          tabindex={tabIndex}
          aria-current={isActive ? 'page' : undefined}
          aria-label={item.chord ? `${item.label}, command ${item.chord}` : item.label}
          onclick={() => onRowClick(item.id)}
        >
          <span class="lp-nav-row-icon" aria-hidden="true">
            <Glyph name={item.icon} size={20} stroke={1.5} />
          </span>
          {#if badgeInfo}
            <span
              class="lp-nav-row-badge"
              data-tone={badgeInfo.tone}
              aria-hidden="true"
              title={badgeInfo.label}
            ></span>
          {/if}
          <span class="lp-nav-row-label" aria-hidden="true">{item.label}</span>
          {#if item.chord}
            <kbd class="lp-nav-row-chord" aria-hidden="true">{item.chord}</kbd>
          {/if}
        </button>
      </Tooltip>
    {/each}
  </div>

  <div class="lp-nav-footer">
    <div class="lp-nav-divider" aria-hidden="true"></div>

    <Tooltip label="Halt the agent" placement="right">
      <button
        bind:this={haltEl}
        type="button"
        class="lp-nav-halt lp-focus"
        data-kind="halt"
        tabindex={focusedId === 'about' ? 0 : -1}
        aria-label="Halt the agent"
        onclick={onHalt}
        onkeydown={onRailKeydown}
      >
        <span class="lp-nav-row-icon" aria-hidden="true">
          <Glyph name="kill-switch" size={20} stroke={1.5} />
        </span>
        <span class="lp-nav-row-label" aria-hidden="true">Halt</span>
      </button>
    </Tooltip>
  </div>

  <div bind:this={threadEl} class="lp-nav-thread" aria-hidden="true">
    <span class="lp-nav-thread-line"></span>
    <span class="lp-nav-thread-glow"></span>
  </div>
</nav>

<style>
  .lp-nav-orbit {
    --lp-nav-collapsed: 60px;
    --lp-nav-expanded: 220px;
    width: var(--lp-nav-collapsed);
    min-width: var(--lp-nav-collapsed);
    height: 100%;
    transition: width var(--lp-dur-normal) var(--lp-ease-thread);
    display: flex;
    flex-direction: column;
    padding: var(--lp-space-2) 0 var(--lp-space-3);
    border-right: 1px solid color-mix(in srgb, var(--lp-ink-ghost) 35%, transparent);
    position: relative;
    background: linear-gradient(
      180deg,
      color-mix(in srgb, var(--lp-paper-warm) 70%, var(--lp-paper)) 0%,
      var(--lp-paper) 100%
    );
    overflow: visible;
    flex-shrink: 0;
    color: var(--lp-ink-mute);
    z-index: 2;
    -webkit-app-region: no-drag;
    app-region: no-drag;
    box-shadow: inset -1px 0 0 color-mix(in srgb, var(--lp-ink-ghost) 12%, transparent);
  }

  .lp-nav-brand {
    display: flex;
    align-items: center;
    gap: var(--lp-space-2);
    padding: var(--lp-space-2) var(--lp-space-3) var(--lp-space-3);
    min-height: 40px;
    overflow: hidden;
  }

  .lp-nav-brand-mark {
    width: 28px;
    height: 28px;
    border-radius: var(--lp-radius-pill);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-family: var(--lp-font-display);
    font-size: 15px;
    font-style: italic;
    color: var(--lp-paper);
    background: var(--lp-synapse);
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--lp-synapse-glow) 40%, transparent);
    flex-shrink: 0;
  }

  .lp-nav-brand-label {
    font-family: var(--lp-font-display);
    font-size: 17px;
    font-style: italic;
    color: var(--lp-ink);
    letter-spacing: -0.02em;
    opacity: 0;
    transform: translateX(-6px);
    transition:
      opacity 160ms var(--lp-ease-thread) 60ms,
      transform 160ms var(--lp-ease-thread) 60ms;
    white-space: nowrap;
  }

  .lp-nav-orbit--expanded .lp-nav-brand-label {
    opacity: 1;
    transform: translateX(0);
  }

  .lp-nav-footer {
    margin-top: auto;
    display: flex;
    flex-direction: column;
  }

  .lp-nav-orbit--expanded {
    width: var(--lp-nav-expanded);
    min-width: var(--lp-nav-expanded);
  }

  .lp-nav-routes {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .lp-nav-divider {
    height: 1px;
    margin: var(--lp-space-3) var(--lp-space-3);
    background: color-mix(in srgb, var(--lp-ink-ghost) 30%, transparent);
  }

  .lp-nav-row,
  .lp-nav-halt {
    appearance: none;
    background: transparent;
    border: 0;
    color: var(--lp-ink-mute);
    cursor: pointer;
    text-align: left;
    height: 44px;
    width: 100%;
    display: flex;
    align-items: center;
    gap: var(--lp-space-2);
    padding: 0 12px;
    border-radius: 0 var(--lp-radius-sm) var(--lp-radius-sm) 0;
    position: relative;
    font-family: var(--lp-font-sans);
    font-size: var(--lp-text-body-sm);
    letter-spacing: var(--lp-tracking-normal);
    transition:
      color var(--lp-dur-fast) var(--lp-ease-thread),
      background-color var(--lp-dur-fast) var(--lp-ease-thread);
  }

  .lp-nav-row:hover,
  .lp-nav-halt:hover {
    color: var(--lp-ink);
  }

  .lp-nav-row-icon {
    width: 28px;
    height: 28px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex: none;
    opacity: 0.72;
    border-radius: var(--lp-radius-pill);
    transition:
      opacity var(--lp-dur-fast) var(--lp-ease-thread),
      color var(--lp-dur-fast) var(--lp-ease-thread),
      background-color var(--lp-dur-fast) var(--lp-ease-thread),
      box-shadow var(--lp-dur-fast) var(--lp-ease-thread);
  }

  .lp-nav-row:hover .lp-nav-row-icon {
    opacity: 1;
    background: color-mix(in srgb, var(--lp-paper-warm) 80%, transparent);
  }

  .lp-nav-row--active {
    background: color-mix(in srgb, var(--lp-paper-warm) 85%, var(--lp-paper));
    color: var(--lp-ink);
  }

  .lp-nav-row--active .lp-nav-row-icon {
    opacity: 1;
    color: var(--lp-synapse);
    background: color-mix(in srgb, var(--lp-synapse) 10%, var(--lp-paper-warm));
    box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--lp-synapse) 22%, transparent);
  }

  .lp-nav-halt:hover .lp-nav-row-icon,
  .lp-nav-halt:focus-visible .lp-nav-row-icon {
    opacity: 1;
  }

  .lp-nav-row-label,
  .lp-nav-row-chord {
    opacity: 0;
    transform: translateX(-4px);
    transition:
      opacity 160ms var(--lp-ease-thread) 80ms,
      transform 160ms var(--lp-ease-thread) 80ms;
    pointer-events: none;
    white-space: nowrap;
  }

  .lp-nav-row-label {
    font-family: var(--lp-font-sans);
    font-size: var(--lp-text-body-sm);
    font-weight: 500;
    color: var(--lp-ink);
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .lp-nav-row-chord {
    margin-left: auto;
    font-family: var(--lp-font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--lp-ink-faint);
    background: var(--lp-paper-deep);
    border: 1px solid color-mix(in srgb, var(--lp-ink-ghost) 40%, transparent);
    border-radius: var(--lp-radius-xs);
    padding: 2px 6px;
    transition-delay: 120ms;
    flex: none;
  }

  .lp-nav-orbit--expanded .lp-nav-row-label,
  .lp-nav-orbit--expanded .lp-nav-row-chord,
  .lp-nav-halt:hover .lp-nav-row-label,
  .lp-nav-halt:focus-visible .lp-nav-row-label {
    opacity: 1;
    transform: translateX(0);
  }

  .lp-nav-row:focus-visible,
  .lp-nav-halt:focus-visible {
    outline: none;
    box-shadow:
      0 0 0 2px var(--lp-synapse),
      0 0 0 5px color-mix(in srgb, var(--lp-pollen) 28%, transparent);
  }

  .lp-nav-row:active:not([disabled]),
  .lp-nav-halt:active:not([disabled]) {
    transform: scale(0.97);
    filter: brightness(0.95) saturate(1.05);
    translate: 0 0.5px;
  }

  .lp-nav-halt {
    margin: 0 var(--lp-space-3);
    height: 44px;
  }

  .lp-nav-halt:hover,
  .lp-nav-halt:hover .lp-nav-row-icon,
  .lp-nav-halt:focus-visible,
  .lp-nav-halt:focus-visible .lp-nav-row-icon {
    color: var(--lp-danger);
  }

  .lp-nav-row-badge {
    position: absolute;
    top: 6px;
    left: 18px;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    z-index: 2;
  }

  .lp-nav-row-badge[data-tone='ok'] {
    background: var(--lp-ok);
    box-shadow: 0 0 0 1.5px color-mix(in srgb, var(--lp-ok) 25%, transparent);
    animation: lp-badge-breath 1.6s var(--lp-ease-thread) infinite;
  }

  .lp-nav-row-badge[data-tone='synapse'] {
    background: var(--lp-synapse);
    box-shadow: 0 0 0 1.5px color-mix(in srgb, var(--lp-synapse) 25%, transparent);
    animation: lp-badge-breath 1.6s var(--lp-ease-thread) infinite;
  }

  .lp-nav-row-badge[data-tone='warn'] {
    background: var(--lp-pollen);
    box-shadow: 0 0 0 1.5px color-mix(in srgb, var(--lp-pollen) 25%, transparent);
    animation: lp-badge-warn 1.4s var(--lp-ease-thread) infinite;
  }

  .lp-nav-row-badge[data-tone='danger'] {
    background: var(--lp-danger);
    box-shadow: 0 0 0 1.5px color-mix(in srgb, var(--lp-danger) 25%, transparent);
    animation: lp-badge-warn 1.4s var(--lp-ease-thread) infinite;
  }

  .lp-nav-row-badge[data-tone='info'] {
    background: var(--lp-sky-deep);
    box-shadow: 0 0 0 1.5px color-mix(in srgb, var(--lp-sky-deep) 25%, transparent);
  }

  .lp-nav-row-badge[data-tone='pollen'] {
    background: var(--lp-pollen);
    box-shadow: 0 0 0 1.5px color-mix(in srgb, var(--lp-pollen) 25%, transparent);
  }

  @keyframes lp-badge-breath {
    0%, 100% { transform: scale(1); }
    50% { transform: scale(1.15); }
  }

  @keyframes lp-badge-warn {
    0%, 100% { transform: scale(1); }
    50% {
      transform: scale(1.18);
      box-shadow: 0 0 0 2px color-mix(in srgb, var(--lp-pollen) 35%, transparent);
    }
  }

  .lp-nav-thread {
    position: absolute;
    left: 0;
    width: 2px;
    top: var(--lp-space-3);
    height: 44px;
    pointer-events: none;
    z-index: 3;
    opacity: 0;
    will-change: transform, top, height;
  }

  .lp-nav-thread-line {
    position: absolute;
    inset: 0;
    background: var(--lp-synapse);
    border-radius: 1px;
  }

  .lp-nav-thread-glow {
    position: absolute;
    inset: -3px;
    background: var(--lp-synapse-glow);
    border-radius: 1px;
    opacity: 0.35;
    filter: blur(3px);
    z-index: -1;
  }

  @media (prefers-reduced-motion: reduce) {
    .lp-nav-row-label,
    .lp-nav-row-chord {
      transition: opacity 80ms linear;
      transition-delay: 0ms;
    }

    .lp-nav-row-badge[data-tone='ok'],
    .lp-nav-row-badge[data-tone='synapse'],
    .lp-nav-row-badge[data-tone='warn'],
    .lp-nav-row-badge[data-tone='danger'] {
      animation: none;
    }
  }
</style>
