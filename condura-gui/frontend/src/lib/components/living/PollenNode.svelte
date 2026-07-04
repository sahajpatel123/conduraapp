<script lang="ts">
  /**
   * PollenNode — Glowing pollen dot with an expanding ring.
   * 
   * A pollen-amber dot with a subtle ring pulse. Used as nav
   * indicators, connection points, and decorative nodes.
   * Simulates the website's synapse-node + synapse-node-ring.
   */
  import './living-paper.css'

  interface Props {
    size?: number
    /** Color variant */
    variant?: 'pollen' | 'synapse' | 'ink' | 'danger' | 'ok'
    /** Whether to show the pulsing ring around the dot */
    ring?: boolean
    /** Active state — brighter glow */
    active?: boolean
    class?: string
    style?: string
    onclick?: () => void
  }

  let {
    size = 8,
    variant = 'pollen',
    ring = false,
    active = false,
    class: className = '',
    style = '',
    onclick,
  }: Props = $props()

  const colorMap = {
    pollen:  { fill: 'var(--lp-pollen)',  glow: 'var(--lp-pollen-glow)' },
    synapse: { fill: 'var(--lp-synapse)',  glow: 'var(--lp-synapse-glow)' },
    ink:     { fill: 'var(--lp-ink-mute)', glow: 'var(--lp-ink-ghost)' },
    danger:  { fill: 'var(--lp-danger)',   glow: '#D45A4A' },
    ok:      { fill: 'var(--lp-ok)',       glow: 'var(--lp-synapse-glow)' },
  }

  const c = $derived(colorMap[variant])
</script>

{#if onclick}
  <button
    type="button"
    class="lp {className}"
    onclick={onclick}
    style="
      display: inline-flex;
      align-items: center;
      justify-content: center;
      position: relative;
      cursor: pointer;
      border: none;
      background: transparent;
      padding: 0;
      width: {size * (ring ? 3 : 1)}px;
      height: {size * (ring ? 3 : 1)}px;
      {style}
    "
  >
  {#if ring}
    <span
      class="lp-pollen-node-ring"
      style="
        position: absolute;
        width: {size * 2.5}px;
        height: {size * 2.5}px;
        border-radius: 50%;
        border: 1.5px solid {c.fill};
        opacity: 0.35;
        animation: lp-node-ring-pulse 3s var(--lp-ease-in-out) infinite;
      "
    ></span>
  {/if}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <span
    style="
      width: {size}px;
      height: {size}px;
      border-radius: 50%;
      background: {c.fill};
      box-shadow: 0 0 {active ? size * 2 : size}px {c.glow};
      transition: box-shadow var(--lp-dur-normal) var(--lp-ease-thread),
                  transform var(--lp-dur-normal) var(--lp-ease-spring);
    "
    onmouseenter={(e) => { if (onclick) (e.target as HTMLElement).style.transform = 'scale(1.25)' }}
    onmouseleave={(e) => { (e.target as HTMLElement).style.transform = 'scale(1)' }}
  ></span>
  </button>
{:else}
  <span
    class="lp {className}"
    style="
      display: inline-flex;
      align-items: center;
      justify-content: center;
      position: relative;
      width: {size * (ring ? 3 : 1)}px;
      height: {size * (ring ? 3 : 1)}px;
      {style}
    "
  >
    {#if ring}
      <span
        class="lp-pollen-node-ring"
        style="
          position: absolute;
          width: {size * 2.5}px;
          height: {size * 2.5}px;
          border-radius: 50%;
          border: 1.5px solid {c.fill};
          opacity: 0.35;
          animation: lp-node-ring-pulse 3s var(--lp-ease-in-out) infinite;
        "
      ></span>
    {/if}
    <span
      style="
        width: {size}px;
        height: {size}px;
        border-radius: 50%;
        background: {c.fill};
        box-shadow: 0 0 {active ? size * 2 : size}px {c.glow};
        transition: box-shadow var(--lp-dur-normal) var(--lp-ease-thread),
                    transform var(--lp-dur-normal) var(--lp-ease-spring);
      "
    ></span>
  </span>
{/if}

<style>
  @keyframes lp-node-ring-pulse {
    0%, 100% { transform: scale(1); opacity: 0.35; }
    50% { transform: scale(1.1); opacity: 0.15; }
  }
</style>
