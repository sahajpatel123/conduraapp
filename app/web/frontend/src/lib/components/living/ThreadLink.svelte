<script lang="ts">
  /**
   * ThreadLink — Link with animated thread underline.
   * 
   * An ink-colored link with a synapse-green thread that draws
   * itself in from the left on hover. Inspired by the website's
   * .thread-link class.
   */
  import './living-paper.css'

  interface Props {
    href?: string
    /** Color variant */
    variant?: 'ink' | 'synapse' | 'pollen'
    class?: string
    children?: import('svelte').Snippet
    onclick?: () => void
    style?: string
  }

  let {
    href,
    variant = 'ink',
    class: className = '',
    children,
    onclick,
    style = '',
  }: Props = $props()

  const colorMap = {
    ink: 'var(--lp-ink)',
    synapse: 'var(--lp-synapse)',
    pollen: 'var(--lp-pollen)',
  }
</script>

{#if href}
  <a
    href={href}
    class="lp lp-thread-link {className}"
    style="
      color: {colorMap[variant]};
      text-decoration: none;
      position: relative;
      cursor: pointer;
      font-family: inherit;
      font-size: inherit;
      line-height: inherit;
      transition: color var(--lp-dur-normal) var(--lp-ease-thread);
      {style}
    "
    onclick={onclick}
  >
    {#if children}{@render children()}{/if}
  </a>
{:else}
  <button
    type="button"
    class="lp lp-thread-link {className}"
    style="
      color: {colorMap[variant]};
      text-decoration: none;
      position: relative;
      cursor: pointer;
      background: none;
      border: none;
      padding: 0;
      font-family: inherit;
      font-size: inherit;
      line-height: inherit;
      transition: color var(--lp-dur-normal) var(--lp-ease-thread);
      {style}
    "
    onclick={onclick}
  >
    {#if children}{@render children()}{/if}
  </button>
{/if}

<style>
  .lp-thread-link::after {
    content: '';
    position: absolute;
    bottom: -2px;
    left: 0;
    width: 100%;
    height: 1.5px;
    background: var(--lp-synapse);
    transform: scaleX(0);
    transform-origin: left center;
    transition: transform 0.4s var(--lp-ease-thread);
  }

  .lp-thread-link:hover::after {
    transform: scaleX(1);
  }

  .lp-thread-link:hover {
    color: var(--lp-synapse) !important;
  }
</style>
