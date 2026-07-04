<!--
  Stack — vertical spacing primitive.

  Every vertical rhythm in v2 routes through Stack. It guarantees
  that no v2 surface accidentally stacks elements with magic margins.

  `gap` is a named spacing token (1..8, 12, 16, 24 from tokens.css).
  `align` controls horizontal alignment of children.
  `tight` removes default paragraph margins when children are text.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'

  let {
    gap = 4 as 1 | 2 | 3 | 4 | 6 | 8 | 12 | 16 | 24,
    align = 'stretch' as 'start' | 'center' | 'end' | 'stretch',
    tight = false,
    children,
    class: klass = '',
    ...rest
  } = $props()
</script>

<div
  data-v2-stack
  data-align={align}
  data-tight={tight}
  class={klass}
  style:gap={`var(--v2-space-${gap})`}
  style:align-items={align}
  {...rest}
>
  {@render children()}
</div>

<style>
  [data-v2-stack] {
    display: flex;
    flex-direction: column;
  }
  [data-v2-stack][data-tight='true'] > :global(p),
  [data-v2-stack][data-tight='true'] > :global(.v2-prose) {
    margin-block: 0;
  }
</style>
