<!--
  Rule — hairline primitive.

  The only way to draw a divider in v2. Uses color-mix to sit at
  exactly the right alpha against whatever surface it lives on.

  orientation: 'horizontal' is the default.
  weight: '1' = single hairline (default), '2' = slightly stronger.
  tone: 'rule' (default) / 'ink-3' / 'accent' / 'ink'.
-->
<script lang="ts">
  let {
    orientation = 'horizontal' as 'horizontal' | 'vertical',
    weight = '1' as '1' | '2',
    tone = 'rule' as 'rule' | 'ink-3' | 'accent' | 'ink',
    inset = '0',
    class: klass = '',
    ...rest
  } = $props()

  const colorFor = (t: string): string => {
    switch (t) {
      case 'ink-3':  return 'var(--v2-ink-3)'
      case 'accent': return 'var(--v2-accent)'
      case 'ink':    return 'var(--v2-ink)'
      default:       return 'var(--v2-rule)'
    }
  }

  const alphaFor = (t: string): string => {
    // rule gets a low alpha; the others get full opacity
    return t === 'rule' ? 'color-mix(in srgb, var(--v2-rule) 60%, transparent)' : colorFor(t)
  }
</script>

<div
  data-v2-rule
  data-orientation={orientation}
  class={klass}
  style:background={alphaFor(tone)}
  style:height={orientation === 'horizontal' ? `${weight}px` : 'auto'}
  style:width={orientation === 'vertical' ? `${weight}px` : '100%'}
  style:margin={orientation === 'horizontal'
    ? `${inset} 0`
    : `0 ${inset}`}
  {...rest}
></div>

<style>
  [data-v2-rule] {
    display: block;
    flex-shrink: 0;
  }
</style>
