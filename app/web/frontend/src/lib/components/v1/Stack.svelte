<!--
  Stack — vertical flex with token-driven gap.

  Per spec §5.1, spacing is a 4px base unit. Always use this rather than
  arbitrary margins so density mode and responsive layouts stay consistent.

  Props:
    gap       — space token key (1..13, or 0-5 for 2px)
    align     — 'start' | 'center' | 'end' | 'stretch'
    justify   — 'start' | 'center' | 'end' | 'between' | 'around'
    padding   — optional padding (space token key)
    as        — element tag, defaults to 'div'
-->
<script lang="ts">
  interface Props {
    gap?: string;
    align?: 'start' | 'center' | 'end' | 'stretch';
    justify?: 'start' | 'center' | 'end' | 'between' | 'around';
    padding?: string;
    as?: 'div' | 'section' | 'article' | 'main' | 'aside' | 'header' | 'footer' | 'nav' | 'ul' | 'ol';
    children?: import('svelte').Snippet;
    class?: string;
  }

  let {
    gap = '4',
    align = 'stretch',
    justify = 'start',
    padding,
    as = 'div',
    children,
    class: className = '',
  }: Props = $props();
</script>

<svelte:element this={as} class="stack {className}" style="--stack-gap: var(--space-{gap}); --stack-padding: {padding ? `var(--space-${padding})` : '0'};">
  {@render children?.()}
</svelte:element>

<style>
  .stack {
    display: flex;
    flex-direction: column;
    gap: var(--stack-gap);
    padding: var(--stack-padding);
  }
  .stack[style*="start"] { align-items: flex-start; }
  .stack[style*="center"] { align-items: center; }
  .stack[style*="end"] { align-items: flex-end; }
  .stack[style*="stretch"] { align-items: stretch; }
</style>