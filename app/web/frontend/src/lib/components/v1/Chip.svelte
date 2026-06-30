<!--
  Chip — selectable suggestion chip, mono label.

  Used in the command surface's ranked interpretations and the empty-state
  suggestions row. Per spec §9.2, the highlighted chip gets a plum hairline
  on its left edge.

  Props:
    selected   — adds the plum hairline
    disabled   — disables selection
    children   — chip label
    onclick    — handler
-->
<script lang="ts">
  interface Props {
    selected?: boolean;
    disabled?: boolean;
    children?: import('svelte').Snippet;
    onclick?: (e: MouseEvent) => void;
  }

  let { selected = false, disabled = false, children, onclick }: Props = $props();
</script>

<button
  class="chip"
  class:chip--selected={selected}
  class:chip--disabled={disabled}
  {disabled}
  type="button"
  onclick={onclick}
>
  {@render children?.()}
</button>

<style>
  .chip {
    display: inline-flex;
    align-items: center;
    height: 28px;
    padding: 0 var(--space-3);
    border-radius: var(--radius-sm);
    background-color: var(--paper-warm-50);
    color: var(--content-secondary);
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.02em;
    cursor: pointer;
    border: 1px solid transparent;
    border-left: 2px solid transparent;
    transition:
      background-color var(--duration-fast) var(--ease-standard),
      border-color var(--duration-fast) var(--ease-standard),
      color var(--duration-fast) var(--ease-standard);
    white-space: nowrap;
  }
  .chip:hover:not(.chip--disabled) {
    background-color: var(--paper-warm-100);
    color: var(--content-primary);
  }
  .chip:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: 2px;
  }
  .chip--selected {
    background-color: var(--paper-warm-100);
    color: var(--content-primary);
    border-left-color: var(--content-accent);
    padding-left: calc(var(--space-3) - 1px); /* compensate so text doesn't jump */
  }
  .chip--disabled {
    color: var(--content-disabled);
    cursor: not-allowed;
  }
</style>