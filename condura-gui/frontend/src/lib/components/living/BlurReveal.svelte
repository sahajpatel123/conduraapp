<script lang="ts">
  /**
   * BlurReveal — Scroll-triggered blur-in reveal.
   * 
   * Starts with blur + translateY + opacity 0, transitions to
   * clear when the element enters the viewport. Uses
   * IntersectionObserver for efficient scroll detection.
   * 
   * Inspired by the website's Reveal component.
   */
  import './living-paper.css'

  interface Props {
    /** Delay in ms before the reveal starts */
    delay?: number
    /** Duration in ms for the reveal animation */
    duration?: number
    /** Distance to translate up in px */
    distance?: number
    /** IntersectionObserver threshold (0-1) */
    threshold?: number
    class?: string
    children?: import('svelte').Snippet
    style?: string
    /** Only animate once (default: true) */
    once?: boolean
  }

  let {
    delay = 0,
    duration = 700,
    distance = 24,
    threshold = 0.1,
    class: className = '',
    children,
    style = '',
    once = true,
  }: Props = $props()

  let el = $state<HTMLDivElement | null>(null)
  let isVisible = $state(false)

  $effect(() => {
    if (!el) return
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          isVisible = true
          if (once) observer.disconnect()
        } else if (!once) {
          isVisible = false
        }
      },
      { threshold }
    )
    observer.observe(el)
    return () => observer.disconnect()
  })
</script>

<div
  bind:this={el}
  class="lp {className}"
  style="
    opacity: {isVisible ? 1 : 0};
    filter: blur({isVisible ? 0 : 6}px);
    transform: translateY({isVisible ? 0 : distance}px);
    transition: opacity {duration}ms var(--lp-ease-thread) {delay}ms,
                filter {duration}ms var(--lp-ease-thread) {delay}ms,
                transform {duration}ms var(--lp-ease-thread) {delay}ms;
    will-change: transform, opacity, filter;
    {style}
  "
>
  {#if children}{@render children()}{/if}
</div>
