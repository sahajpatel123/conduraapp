<script lang="ts">
  import './living-paper.css'

  interface Props {
    /** 'thread' = full-width gradient hairline, 'skinny' = thin solid, 'spaced' = with vertical margin */
    variant?: 'thread' | 'skinny' | 'spaced'
    class?: string
    style?: string
  }

  let {
    variant = 'thread',
    class: className = '',
    style = '',
  }: Props = $props()
</script>

<hr
  class="lp {className}"
  class:lp-rule={variant === 'thread'}
  style="
    border: none;
    margin: 0;
    {variant === 'skinny' ? 'height: 1px; background: var(--lp-ink-ghost);' : ''}
    {variant === 'spaced' ? 'height: 1px; background: var(--lp-ink-ghost); margin: var(--lp-space-8) 0;' : ''}
    {style}
  "
/>
