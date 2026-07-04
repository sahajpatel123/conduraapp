<script lang="ts">
  /**
   * WordReveal — Per-word masked ink reveal animation.
   * 
   * Each word rises out of an overflow:hidden mask with stagger,
   * inspired by the website's WordReveal component.
   * The animation triggers once when the element enters the viewport.
   */
  import './living-paper.css'

  interface Props {
    /** The text to reveal. If empty, falls back to children. */
    text?: string
    /** Stagger delay between each word in ms */
    stagger?: number
    /** Tag to render */
    as?: 'h1' | 'h2' | 'h3' | 'h4' | 'p' | 'span' | 'div'
    class?: string
    children?: import('svelte').Snippet
    style?: string
    /** Threshold for IntersectionObserver (0-1) */
    threshold?: number
    /** Start delay before the first word in ms */
    delay?: number
  }

  let {
    text = '',
    stagger = 40,
    as: tag = 'span',
    class: className = '',
    children,
    style = '',
    threshold = 0.1,
    delay = 0,
  }: Props = $props()

  let rootEl = $state<HTMLElement | null>(null)
  let hasRevealed = $state(false)

  const words = $derived(text ? text.split(' ') : [])

  $effect(() => {
    if (!rootEl || hasRevealed || words.length === 0 && (!children)) return
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          hasRevealed = true
          observer.disconnect()
        }
      },
      { threshold }
    )
    observer.observe(rootEl)
    return () => observer.disconnect()
  })
</script>

<svelte:element
  this={tag}
  bind:this={rootEl}
  class="lp {className}"
  style="
    display: inline;
    {style}
  "
>
  {#if words.length > 0}
    {#each words as word, i}
      <span
        class="lp-word-reveal-mask"
        style="
          display: inline-block;
          overflow: hidden;
          vertical-align: top;
          padding-bottom: 0.08em;
        "
      >
        <span
          class="lp-word-reveal-word"
          style="
            display: inline-block;
            transform: translateY({hasRevealed ? '0%' : '110%'});
            transition: transform 0.7s var(--lp-ease-thread) {delay + i * stagger}ms;
          "
        >{word}{i < words.length - 1 ? '\u00A0' : ''}</span>
      </span>
    {/each}
  {:else if children}
    {@render children()}
  {/if}
</svelte:element>

<style>
  .lp-word-reveal-mask {
    vertical-align: top;
  }
</style>
