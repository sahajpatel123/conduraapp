<!--
  Spacer — fixed-size empty box. Use to push elements apart in a flex
  container or to insert vertical breathing room.

  Props:
    size — space token key (1..13, or '0-5' for 2px). For horizontal
           spacers inside an Inline, use the `axis` prop.
    axis — 'vertical' | 'horizontal' (default: vertical)
-->
<script lang="ts">
  interface Props {
    size?: string;
    axis?: 'vertical' | 'horizontal';
  }

  let { size = '4', axis = 'vertical' }: Props = $props();
</script>

<div
  class="spacer spacer--{axis}"
  style="--spacer-size: var(--space-{size});"
  aria-hidden="true"
></div>

<style>
  .spacer--vertical {
    width: 100%;
    height: var(--spacer-size);
  }
  .spacer--horizontal {
    width: var(--spacer-size);
    height: 1em;
  }
</style>