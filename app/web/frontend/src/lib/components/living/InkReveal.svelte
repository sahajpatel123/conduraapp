<script lang="ts">
  /**
   * InkReveal — Clip-path ink reveal animation.
   * 
   * Content is revealed via a clip-path that animates from
   * inset(0 100% 0 0) to inset(0 0 0 0), like ink spreading
   * across paper. Used for section titles and emphasis moments.
   * 
   * Inspired by the website's @keyframes ink-reveal.
   */
  import './living-paper.css'

  interface Props {
    /** Direction of the reveal */
    direction?: 'left' | 'right' | 'up' | 'down'
    /** Duration in ms */
    duration?: number
    /** Delay in ms */
    delay?: number
    class?: string
    children?: import('svelte').Snippet
    style?: string
    /** Trigger the animation on viewport entry */
    viewport?: boolean
    /** Threshold for IntersectionObserver */
    threshold?: number
  }

  let {
    direction = 'left',
    duration = 800,
    delay = 0,
    class: className = '',
    children,
    style = '',
    viewport = true,
    threshold = 0.1,
  }: Props = $props()

  let el = $state<HTMLDivElement | null>(null)
  let isVisible = $state(false)

  // Determine the clip-path based on direction
  const startClip = $derived({
    left: 'inset(0 100% 0 0)',
    right: 'inset(0 0 0 100%)',
    up: 'inset(100% 0 0 0)',
    down: 'inset(0 0 100% 0)',
  }[direction])

  $effect(() => {
    if (!viewport || !el) return
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          isVisible = true
          observer.disconnect()
        }
      },
      { threshold }
    )
    observer.observe(el)
    return () => observer.disconnect()
  })

  // Auto-trigger if not viewport-based
  $effect(() => {
    if (!viewport) isVisible = true
  })
</script>

<div
  bind:this={el}
  class="lp {className}"
  style="
    clip-path: {isVisible ? 'inset(0 0 0 0)' : startClip};
    transition: clip-path {duration}ms var(--lp-ease-thread) {delay}ms;
    {style}
  "
>
  {#if children}{@render children()}{/if}
</div>
