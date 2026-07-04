<!--
  Inline — horizontal flex with token-driven gap.

  Wraps children onto multiple lines by default. Use `wrap={false}` for a
  single-line row that ellipsises.

  Props:
    gap       — space token key (1..13, or 0-5 for 2px)
    align     — 'start' | 'center' | 'end' | 'baseline' | 'stretch'
    justify   — 'start' | 'center' | 'end' | 'between' | 'around'
    wrap      — true (default) | false
    as        — element tag, defaults to 'div'
-->
<script lang="ts">
  interface Props {
    gap?: string;
    align?: 'start' | 'center' | 'end' | 'baseline' | 'stretch';
    justify?: 'start' | 'center' | 'end' | 'between' | 'around';
    wrap?: boolean;
    as?: 'div' | 'span' | 'nav' | 'header' | 'footer' | 'section' | 'ul' | 'ol';
    children?: import('svelte').Snippet;
    class?: string;
  }

  let {
    gap = '4',
    align = 'center',
    justify = 'start',
    wrap = true,
    as = 'div',
    children,
    class: className = '',
  }: Props = $props();
</script>

<svelte:element
  this={as}
  class="inline {className}"
  style="--inline-gap: var(--space-{gap}); --inline-align: {align === 'start' ? 'flex-start' : align === 'end' ? 'flex-end' : align}; --inline-justify: {justify === 'start' ? 'flex-start' : justify === 'end' ? 'flex-end' : justify};"
>
  {@render children?.()}
</svelte:element>

<style>
  .inline {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    gap: var(--inline-gap);
    align-items: var(--inline-align);
    justify-content: var(--inline-justify);
  }
  .inline[style*="wrap: false"],
  .inline:not([style*="wrap: true"]) {
    flex-wrap: wrap;
  }
  /* When wrap={false}, apply via attribute on parent */
  :global(.inline.no-wrap) {
    flex-wrap: nowrap;
  }
</style>