<!--
  Surface — base container with token-driven bg/border/radius/padding.

  The atomic level above Stack/Inline/Hairline. Use this when you need
  a panel or card surface, not just a layout primitive.

  Per spec §5: most hierarchy is hairline + tone, NEVER drop shadows.
  Use this instead of stacking divs.

  Props:
    variant   — 'base' | 'sunken' | 'raised' | 'overlay' | 'inverted'
    bordered  — adds a hairline border
    padding   — space token (e.g., '4', '6')
    radius    — 'none' | 'sm' | 'md' | 'lg' | 'xl' | 'pill'
    as        — element tag, defaults to 'div'
    children  — slot content
-->
<script lang="ts">
  interface Props {
    variant?: 'base' | 'sunken' | 'raised' | 'overlay' | 'inverted';
    bordered?: boolean;
    padding?: string;
    radius?: 'none' | 'sm' | 'md' | 'lg' | 'xl' | 'pill';
    as?: 'div' | 'section' | 'article' | 'aside' | 'header' | 'footer' | 'main' | 'nav';
    children?: import('svelte').Snippet;
    class?: string;
  }

  let {
    variant = 'raised',
    bordered = true,
    padding,
    radius = 'md',
    as = 'div',
    children,
    class: className = '',
  }: Props = $props();
</script>

<svelte:element
  this={as}
  class="surface surface--{variant} surface--r-{radius} {className}"
  class:surface--bordered={bordered}
  style={padding ? `--surface-padding: var(--space-${padding});` : ''}
>
  {@render children?.()}
</svelte:element>

<style>
  .surface {
    background-color: var(--surface-raised);
    padding: var(--surface-padding, 0);
  }

  .surface--base { background-color: var(--surface-base); }
  .surface--sunken { background-color: var(--surface-sunken); }
  .surface--raised { background-color: var(--surface-raised); }
  .surface--overlay { background-color: var(--surface-overlay); }
  .surface--inverted { background-color: var(--surface-inverted); color: var(--content-inverse); }

  .surface--bordered {
    border: 1px solid var(--border-default);
  }

  /* Radius */
  .surface--r-none { border-radius: 0; }
  .surface--r-sm { border-radius: var(--radius-sm); }
  .surface--r-md { border-radius: var(--radius-md); }
  .surface--r-lg { border-radius: var(--radius-lg); }
  .surface--r-xl { border-radius: var(--radius-xl); }
  .surface--r-pill { border-radius: var(--radius-pill); }
</style>