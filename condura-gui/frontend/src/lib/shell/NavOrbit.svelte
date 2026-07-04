<script lang="ts">
  /**
   * NavOrbit — Floating dot-based navigation.
   * 
   * A column of node-dots connected by a living SVG synapse thread.
   * Each dot represents a route. The active dot has synapse-green glow
   * + pollen core. On hover, nearby labels fade in. Clicking navigates.
   * 
   * The whole thing feels organic — dots breathe, the thread pulses,
   * labels ink-reveal on expand.
   */
  import { PollenNode, InkText, SynapseThread } from '$lib/components/living'
  import { ROUTE_HASH, type RouteId } from '$lib/condura/NavRail.svelte'

  export type { RouteId }

  interface NavItem {
    id: RouteId
    label: string
    blurb: string
    icon: string
  }

  interface Props {
    route: RouteId
    onnavigate: (r: RouteId) => void
    /** Collapsed (dots only, ~40px) or expanded (dots + labels, ~220px) */
    collapsed?: boolean
  }

  let {
    route: activeRoute,
    onnavigate,
    collapsed = true,
  }: Props = $props()

  const ITEMS: NavItem[] = [
    { id: 'chat', label: 'Chat', blurb: 'Converse with the agent', icon: 'chat' },
    { id: 'hub', label: 'Hub', blurb: 'Discover skills', icon: 'hub' },
    { id: 'skills', label: 'Skills', blurb: 'Installed procedures', icon: 'skills' },
    { id: 'sync', label: 'Sync', blurb: 'Pair devices', icon: 'sync' },
    { id: 'audit', label: 'Audit', blurb: 'Event log', icon: 'audit' },
    { id: 'replay', label: 'Replay', blurb: 'Action timeline', icon: 'replay' },
    { id: 'channels', label: 'Channels', blurb: 'Messaging integrations', icon: 'channels' },
    { id: 'delegation', label: 'Delegation', blurb: 'Sub-agent constellation', icon: 'delegation' },
    { id: 'settings', label: 'Settings', blurb: 'Configuration', icon: 'settings' },
    { id: 'about', label: 'About', blurb: 'Colophon', icon: 'about' },
  ]

  // Calculate dot positions for the SVG thread
  // Each dot: 32px apart, starting at 28px from top, 20px from left
  const DOT_GAP = 36
  const DOT_LEFT = 20
  const DOT_TOP = 24
  const DOT_SIZE = 8

  const threadPoints = $derived(
    ITEMS.map((_, i) => ({
      x: DOT_LEFT + DOT_SIZE / 2,
      y: DOT_TOP + i * DOT_GAP + DOT_SIZE / 2,
    }))
  )

  // Track hover state for magnetic proximity
  let hoveredId = $state<string | null>(null)

  // Total height for the SVG + container
  const totalHeight = $derived(DOT_TOP + ITEMS.length * DOT_GAP + 20)
</script>

<nav
  class="lp lp-nav-orbit"
  class:lp-nav-orbit--collapsed={collapsed}
  style="
    position: relative;
    width: {collapsed ? '52px' : '240px'};
    min-height: {totalHeight}px;
    transition: width var(--lp-dur-slow) var(--lp-ease-thread);
    flex-shrink: 0;
    padding-top: var(--lp-space-2);
    overflow: hidden;
  "
  aria-label="Main navigation"
>
  <!-- SVG thread connecting all dots -->
  <div style="position: absolute; inset: 0; pointer-events: none; z-index: 0;">
    <SynapseThread
      points={threadPoints}
      animate={true}
      glow={true}
      color="var(--lp-synapse)"
      width={1}
      duration={1500}
    />
  </div>

  <!-- Route dots -->
  <div style="position: relative; z-index: 1; display: flex; flex-direction: column;">
    {#each ITEMS as item, i (item.id)}
      <button
        type="button"
        class="lp-nav-item lp-focus"
        class:lp-nav-item--active={activeRoute === item.id}
        class:lp-nav-item--hovered={hoveredId === item.id}
        onclick={() => onnavigate(item.id)}
        onmouseenter={() => (hoveredId = item.id)}
        onmouseleave={() => (hoveredId === item.id && (hoveredId = null))}
        aria-current={activeRoute === item.id ? 'page' : undefined}
        aria-label={item.label}
        style="
          display: flex;
          align-items: center;
          gap: var(--lp-space-3);
          padding: 6px var(--lp-space-3);
          border: none;
          background: transparent;
          cursor: pointer;
          text-align: left;
          font-family: var(--lp-font-sans);
          min-height: {DOT_GAP}px;
          transition: background var(--lp-dur-fast) var(--lp-ease-thread);
          border-radius: 0 var(--lp-radius-sm) var(--lp-radius-sm) 0;
          position: relative;
        "
      >
        <!-- Dot indicator -->
        <div style="flex-shrink: 0; width: {DOT_LEFT}px; display: flex; justify-content: center;">
          <PollenNode
            size={activeRoute === item.id ? 8 : 5}
            variant={activeRoute === item.id ? 'synapse' : 'ink'}
            ring={activeRoute === item.id}
            active={activeRoute === item.id}
          />
        </div>

        <!-- Label + blurb — only visible when expanded -->
        <div
          class="lp-nav-label"
          style="
            opacity: {collapsed ? 0 : 1};
            transform: translateX({collapsed ? -12 : 0}px);
            transition: opacity var(--lp-dur-normal) var(--lp-ease-thread),
                        transform var(--lp-dur-normal) var(--lp-ease-thread);
            pointer-events: {collapsed ? 'none' : 'auto'};
          "
        >
          <div style="
            font-family: var(--lp-font-display);
            font-size: var(--lp-text-body);
            color: var(--lp-ink);
            font-weight: 500;
            line-height: 1.2;
          ">{item.label}</div>
          <div style="
            font-family: var(--lp-font-sans);
            font-size: var(--lp-text-caption);
            color: var(--lp-ink-mute);
            line-height: 1.3;
            margin-top: 1px;
          ">{item.blurb}</div>
        </div>

        <!-- Active indicator — synapse left border -->
        {#if activeRoute === item.id}
          <div style="
            position: absolute;
            left: 0;
            top: 6px;
            bottom: 6px;
            width: 2px;
            background: var(--lp-synapse);
            border-radius: 1px;
            transform-origin: top;
          "></div>
        {/if}
      </button>
    {/each}
  </div>
</nav>

<style>
  .lp-nav-item:hover {
    background: var(--lp-paper-warm);
  }

  .lp-nav-item--active {
    background: var(--lp-paper-warm);
  }

  .lp-nav-item:active {
    transform: scale(0.98);
  }

  .lp-nav-item:focus-visible {
    outline: 2px solid var(--lp-synapse);
    outline-offset: -2px;
  }
</style>
