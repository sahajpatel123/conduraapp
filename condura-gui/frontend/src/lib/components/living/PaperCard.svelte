<script lang="ts">
  import './living-paper.css'

  interface Props {
    border?: 'synapse' | 'pollen' | 'none'
    padding?: string
    class?: string
    onclick?: () => void
    children?: import('svelte').Snippet
    style?: string
  }

  let {
    border = 'none',
    padding = 'var(--lp-space-5)',
    class: className = '',
    onclick,
    children,
    style = '',
  }: Props = $props()

  const borderColors: Record<string, string> = {
    synapse: 'var(--lp-synapse)',
    pollen: 'var(--lp-pollen)',
    none: 'transparent',
  }
</script>

<button
  class="lp lp-paper-card {className} lp-focus"
  class:lp-paper-card--clickable={!!onclick}
  type="button"
  onclick={onclick}
  style="
    background: var(--lp-paper-warm);
    box-shadow: var(--lp-shadow-card);
    padding: {padding};
    border-radius: var(--lp-radius-md);
    border: none;
    border-left: 2px solid {borderColors[border]};
    position: relative;
    overflow: hidden;
    cursor: {onclick ? 'pointer' : 'default'};
    text-align: left;
    font-family: inherit;
    width: 100%;
    transition: box-shadow var(--lp-dur-normal) var(--lp-ease-thread),
                transform var(--lp-dur-normal) var(--lp-ease-spring);
    {style}
  "
>
  <!-- paper grain -->
  <div
    class="lp-grain"
    style="position: absolute; inset: 0; border-radius: inherit;"
  ></div>
  <!-- content -->
  <div style="position: relative; z-index: 1;">
    {#if children}{@render children()}{/if}
  </div>
</button>

<style>
  .lp-paper-card--clickable:hover {
    box-shadow: var(--lp-shadow-float);
    transform: translateY(-1px);
  }
  .lp-paper-card--clickable:active {
    transform: translateY(0);
    box-shadow: var(--lp-shadow-card);
  }
</style>
