<!--
  Hairline — 1px line in --border-subtle. The atomic unit of separation.

  Per spec §2.3: hierarchy is expressed through hairline borders + tone,
  never drop shadows. This component is the seam.

  Props:
    orientation — 'horizontal' | 'vertical'
    variant    — 'subtle' | 'default' | 'strong' | 'accent'
    inset      — adds horizontal/vertical margin (uses --space-* tokens)
-->
<script lang="ts">
  interface Props {
    orientation?: 'horizontal' | 'vertical';
    variant?: 'subtle' | 'default' | 'strong' | 'accent';
    inset?: boolean;
  }

  let { orientation = 'horizontal', variant = 'subtle', inset = false }: Props = $props();
</script>

<div
  class="hairline hairline--{orientation} hairline--{variant}"
  class:hairline--inset={inset}
  role="separator"
  aria-orientation={orientation}
></div>

<style>
  .hairline {
    background-color: var(--border-subtle);
    flex-shrink: 0;
  }

  .hairline--horizontal {
    width: 100%;
    height: 1px;
  }

  .hairline--vertical {
    width: 1px;
    height: 100%;
  }

  .hairline--default {
    background-color: var(--border-default);
  }

  .hairline--strong {
    background-color: var(--border-strong);
  }

  .hairline--accent {
    background-color: var(--content-accent);
  }

  /* Inset variants — pull away from edges by the standard space-6 inset */
  .hairline--inset.hairline--horizontal {
    margin-left: var(--space-6);
    margin-right: var(--space-6);
  }

  .hairline--inset.hairline--vertical {
    margin-top: var(--space-6);
    margin-bottom: var(--space-6);
  }
</style>