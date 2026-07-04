<script lang="ts">
  import './living-paper.css'

  interface Props {
    /** Points the thread passes through: [{x: number, y: number}]. If empty, draws a simple horizontal/vertical line. */
    points?: { x: number; y: number }[]
    /** Orientation when no points provided */
    orientation?: 'horizontal' | 'vertical'
    /** Length of the line in px when no points */
    length?: number
    /** Glow layer beneath the thread */
    glow?: boolean
    /** Self-draw animation on mount */
    animate?: boolean
    /** Animation duration in ms */
    duration?: number
    /** Line color */
    color?: string
    /** Line width */
    width?: number
    class?: string
    style?: string
  }

  let {
    points = [],
    orientation = 'horizontal',
    length = 200,
    glow = false,
    animate = true,
    duration = 1200,
    color = 'var(--lp-synapse)',
    width = 1.25,
    class: className = '',
    style = '',
  }: Props = $props()

  let pathEl = $state<SVGPathElement | null>(null)
  let pathLength = $state(0)

  function buildPath(): string {
    if (points.length >= 2) {
      // Build a smooth curve through points
      let d = `M ${points[0].x} ${points[0].y}`
      for (let i = 1; i < points.length; i++) {
        const prev = points[i - 1]
        const curr = points[i]
        const cx1 = prev.x + (curr.x - prev.x) * 0.4
        const cy1 = prev.y
        const cx2 = curr.x - (curr.x - prev.x) * 0.4
        const cy2 = curr.y
        d += ` C ${cx1} ${cy1}, ${cx2} ${cy2}, ${curr.x} ${curr.y}`
      }
      return d
    }
    // Default straight line
    if (orientation === 'vertical') {
      return `M 0 0 L 0 ${length}`
    }
    return `M 0 0 L ${length} 0`
  }

  $effect(() => {
    if (pathEl && animate) {
      const len = pathEl.getTotalLength()
      pathLength = len
      pathEl.style.strokeDasharray = `${len}`
      pathEl.style.strokeDashoffset = `${len}`
      // Trigger animation on next frame
      requestAnimationFrame(() => {
        pathEl!.style.transition = `stroke-dashoffset ${duration}ms var(--lp-ease-thread)`
        pathEl!.style.strokeDashoffset = '0'
      })
    }
  })
</script>

<svg
  class="lp {className}"
  style="
    overflow: visible;
    pointer-events: none;
    {style}
  "
  width="100%"
  height="100%"
>
  <!-- Glow layer -->
  {#if glow}
    <path
      d={buildPath()}
      fill="none"
      stroke="var(--lp-synapse-glow)"
      stroke-width={width * 3}
      opacity="0.15"
      style="filter: blur(4px);"
    />
  {/if}
  <!-- Main thread -->
  <path
    bind:this={pathEl}
    d={buildPath()}
    fill="none"
    stroke={color}
    stroke-width={width}
    stroke-linecap="round"
    stroke-linejoin="round"
  />
</svg>
