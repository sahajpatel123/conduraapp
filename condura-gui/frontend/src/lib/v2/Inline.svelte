<!--
  Inline — horizontal flow primitive.

  Use for rows of pills, chips, buttons, key-value pairs.
  When the row outgrows its container, items can wrap (default)
  or stay on one line (`nowrap`).
-->
<script lang="ts">
  import type { Snippet } from 'svelte'

  let {
    gap = 2 as 1 | 2 | 3 | 4 | 6 | 8 | 12,
    align = 'center' as 'start' | 'center' | 'end' | 'baseline' | 'stretch',
    justify = 'start' as 'start' | 'center' | 'end' | 'between' | 'around',
    nowrap = false,
    children,
    class: klass = '',
    ...rest
  } = $props()

  const justifyFor = (j: string): string => {
    switch (j) {
      case 'between': return 'space-between'
      case 'around':  return 'space-around'
      case 'start':   return 'flex-start'
      case 'end':     return 'flex-end'
      case 'center':  return 'center'
      default:        return 'flex-start'
    }
  }
</script>

<div
  data-v2-inline
  data-align={align}
  data-nowrap={nowrap}
  class={klass}
  style:gap={`var(--v2-space-${gap})`}
  style:align-items={align}
  style:justify-content={justifyFor(justify)}
  style:flex-wrap={nowrap ? 'nowrap' : 'wrap'}
  {...rest}
>
  {@render children()}
</div>

<style>
  [data-v2-inline] {
    display: flex;
    flex-direction: row;
  }
</style>
