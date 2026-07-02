<!--
  Sidebar — Condura v2 navigation.

  A book spine, not a strip. 72px wide when expanded, collapses to
  8px when not in use. Each route is a monogram + label; the label
  is rotated 90° when collapsed and on hover, slides into view from
  the right. Active state is a quiet accent dot — never a fill.

  The active state is purely visual; the parent route owns route
  state and passes the active id via `active` prop.

  Props:
    items:      {id, monogram, label, icon?}[]   the routes
    active:     id of the currently active route
    collapsed?: boolean (default false)
    onSelect:   (id) => void
    onToggle?:  () => void
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Ink, Stack, Glyph } from '$lib/v2'

  export interface SidebarItem {
    id: string
    monogram: string
    label: string
  }

  let {
    items = [] as SidebarItem[],
    active = '' as string,
    collapsed = false as boolean,
    onSelect = undefined as ((id: string) => void) | undefined,
    onToggle = undefined as (() => void) | undefined,
  } = $props()

  let hoveredId = $state<string | null>(null)
</script>

<nav
  data-v2
  data-collapsed={collapsed}
  aria-label="Primary navigation"
  style="
    width: {collapsed ? '8px' : '72px'};
    height: 100vh;
    background: var(--v2-paper-2);
    border-right: 1px solid color-mix(in srgb, var(--v2-rule) 60%, transparent);
    display: flex;
    flex-direction: column;
    align-items: {collapsed ? 'center' : 'center'};
    padding: {collapsed ? 'var(--v2-space-4) 0' : 'var(--v2-space-4) 0'};
    gap: var(--v2-space-2);
    transition: width var(--v2-dur-mid) var(--v2-ease-settle);
    position: relative;
    overflow: visible;
    flex-shrink: 0;
  "
>
  {#each items as item}
    {@const isActive = active === item.id}
    {@const isHovered = hoveredId === item.id}

    <button
      data-v2
      data-active={isActive}
      onclick={() => onSelect?.(item.id)}
      onmouseenter={() => { hoveredId = item.id }}
      onmouseleave={() => { hoveredId = null }}
      onfocus={() => { hoveredId = item.id }}
      onblur={() => { hoveredId = null }}
      aria-label={item.label}
      aria-current={isActive ? 'page' : undefined}
      style="
        position: relative;
        all: unset;
        cursor: pointer;
        width: {collapsed ? 'auto' : '40px'};
        height: 40px;
        display: grid;
        place-items: center;
        transition: transform var(--v2-dur-fast) var(--v2-ease-out-soft);
      "
    >
      <!-- Monogram disc — 32×32 centered in a 40×40 button. The
           extra 8px gives the active rail room to breathe. The disc
           is paper-spine-shaped (radius-1, 4px), italic display
           serif, so the column reads as a stack of book-spines. -->
      <span style="
        display: grid; place-items: center;
        width: 32px; height: 32px;
        border-radius: var(--v2-radius-1);
        background: {isActive ? 'var(--v2-accent)' : 'transparent'};
        color: {isActive ? 'var(--v2-paper)' : 'var(--v2-ink-3)'};
        font-family: var(--v2-font-display);
        font-size: var(--v2-text-14);
        font-style: italic;
        letter-spacing: 0.02em;
        transition:
          background-color var(--v2-dur-fast) var(--v2-ease-out-soft),
          color            var(--v2-dur-fast) var(--v2-ease-out-soft);
      ">
        {item.monogram}
      </span>

      <!-- Active rail — a quiet 4px whisper at the right edge. Never a fill. -->
      {#if isActive && !collapsed}
        <span style="
          position: absolute;
          right: -4px;
          top: 50%;
          transform: translateY(-50%);
          width: 3px; height: 18px;
          border-radius: var(--v2-radius-1);
          background: var(--v2-accent);
        "></span>
      {/if}

      <!-- Hover label — appears immediately to the right of the
           collapsed spine, slides in from the spine itself so it
           reads as attached to the book, not floating in the void. -->
      {#if collapsed && isHovered}
        <span
          role="tooltip"
          style="
            position: absolute;
            left: calc(100% + 8px);
            top: 50%;
            transform: translateY(-50%);
            white-space: nowrap;
            padding: var(--v2-space-2) var(--v2-space-3);
            background: var(--v2-ink);
            color: var(--v2-paper);
            border-radius: var(--v2-radius-1);
            font-family: var(--v2-font-sans);
            font-size: var(--v2-text-12);
            font-weight: 500;
            animation: v2-slide-right var(--v2-dur-fast) var(--v2-ease-out-soft) both;
            z-index: var(--v2-z-overlay);
            pointer-events: none;
            box-shadow: var(--v2-shadow-2);
          "
        >
          {item.label}
        </span>
      {/if}
    </button>
  {/each}

  <!-- Spacer that pushes the toggle to the bottom -->
  <div style="flex: 1;"></div>

  <!-- Collapse toggle at the bottom -->
  <button
    data-v2
    onclick={onToggle}
    aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
    style="
      position: relative;
      all: unset;
      cursor: pointer;
      width: {collapsed ? '8px' : '36px'};
      height: {collapsed ? '8px' : '24px'};
      display: grid;
      place-items: center;
      color: var(--v2-ink-3);
      transition: all var(--v2-dur-fast) var(--v2-ease-out-soft);
    "
  >
    {#if collapsed}
      <!-- Indicator dot when collapsed (the binding edge) -->
      <div style="
        width: 4px; height: 4px;
        border-radius: var(--v2-radius-pill);
        background: var(--v2-ink-3);
      "></div>
    {:else}
      <Glyph name="chevron-left" size={14} />
    {/if}
  </button>
</nav>
