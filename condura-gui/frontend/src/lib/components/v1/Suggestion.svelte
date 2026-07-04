<!--
  Suggestion — interpretation card.

  Per spec §9.2: ranked interpretations appear below the omni-bar as the
  user types. Each is 48px tall, serif interpretation + sans steps preview,
  plum hairline on left edge when highlighted.

  Props:
    interpretation — the user-facing sentence (serif)
    steps          — preview of steps the agent will take (sans, smaller)
    highlighted    — adds the plum hairline
    children       — slot for trailing actions or icons
    onclick        — handler when selected
-->
<script lang="ts">
  interface Props {
    interpretation: string;
    steps?: string;
    highlighted?: boolean;
    children?: import('svelte').Snippet;
    onclick?: (e: MouseEvent) => void;
  }

  let { interpretation, steps, highlighted = false, children, onclick }: Props = $props();
</script>

<button
  class="suggestion"
  class:suggestion--highlighted={highlighted}
  type="button"
  onclick={onclick}
>
  <div class="suggestion__text">
    <span class="suggestion__interpretation">{interpretation}</span>
    {#if steps}
      <span class="suggestion__steps">{steps}</span>
    {/if}
  </div>
  {#if children}
    <div class="suggestion__trailing">{@render children()}</div>
  {/if}
</button>

<style>
  .suggestion {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    width: 100%;
    min-height: 48px;
    padding: var(--space-3) var(--space-4);
    padding-left: calc(var(--space-4) - 2px); /* compensate for hairline */
    background-color: transparent;
    border: 1px solid transparent;
    border-left: 2px solid transparent;
    border-radius: var(--radius-sm);
    cursor: pointer;
    text-align: left;
    transition: background-color var(--duration-fast) var(--ease-standard);
    font-family: var(--font-sans);
  }
  .suggestion:hover {
    background-color: var(--paper-warm-50);
  }
  .suggestion:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: -2px;
  }
  .suggestion--highlighted {
    background-color: var(--paper-warm-100);
    border-left-color: var(--content-accent);
  }

  .suggestion__text {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }
  .suggestion__interpretation {
    font-family: var(--font-serif);
    font-size: var(--text-body-size);
    color: var(--content-primary);
    line-height: 1.4;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .suggestion__steps {
    font-family: var(--font-sans);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    line-height: 1.4;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .suggestion__trailing {
    flex-shrink: 0;
    color: var(--content-tertiary);
  }
</style>