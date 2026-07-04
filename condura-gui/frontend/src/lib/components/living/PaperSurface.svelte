<script lang="ts">
  import './living-paper.css'

  type Variant = 'page' | 'card' | 'card-deep' | 'ink' | 'raised'
  type Tone = 'paper' | 'warm' | 'deep' | 'ink'

  interface Props {
    variant?: Variant
    tone?: Tone
    grain?: boolean
    padding?: string
    radius?: string
    class?: string
    children?: import('svelte').Snippet
    style?: string
  }

  let {
    variant = 'page',
    tone = 'paper',
    grain = true,
    padding = '0',
    radius = '0',
    class: className = '',
    children,
    style = '',
  }: Props = $props()

  const toneMap: Record<Tone, string> = {
    paper: 'var(--lp-paper)',
    warm: 'var(--lp-paper-warm)',
    deep: 'var(--lp-paper-deep)',
    ink: 'var(--lp-ink)',
  }

  const shadowMap: Record<Variant, string> = {
    page: 'none',
    card: 'var(--lp-shadow-card)',
    'card-deep': 'var(--lp-shadow-float)',
    ink: 'none',
    raised: 'var(--lp-shadow-float)',
  }
</script>

<div
  class="lp lp-surface-{variant} {className}"
  class:lp-grain={grain && variant !== 'ink'}
  style="
    background: {toneMap[tone]};
    box-shadow: {shadowMap[variant]};
    padding: {padding};
    border-radius: {radius};
    position: relative;
    overflow: hidden;
    {style}
  "
>
  {#if children}
    <div style="position: relative; z-index: 1;">
      {@render children()}
    </div>
  {/if}
</div>
