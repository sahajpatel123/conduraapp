<script lang="ts">
  // Condura Thread — the self-drawing synapse line. The brand's spine motif.
  // pathLength="1" + non-scaling-stroke means one component works as a 2px
  // divider or a tall progress line; stroke-dashoffset 1→0 draws it in.
  let {
    orientation = 'h',
    draw = true,
    glow = true,
    class: cls = '',
  }: {
    orientation?: 'h' | 'v';
    draw?: boolean;
    glow?: boolean;
    class?: string;
  } = $props();

  let d = $derived(orientation === 'v' ? 'M 0.5 0 L 0.5 1' : 'M 0 0.5 L 1 0.5');
  let dashoffset = $derived(draw ? 0 : 1);
</script>

<svg
  class="condura-thread {orientation} {cls}"
  preserveAspectRatio="none"
  viewBox="0 0 1 1"
  aria-hidden="true"
>
  {#if glow}
    <path class="glow" {d} pathLength="1" vector-effect="non-scaling-stroke" />
  {/if}
  <path
    class="line"
    {d}
    pathLength="1"
    vector-effect="non-scaling-stroke"
    stroke-dasharray="1"
    stroke-dashoffset={dashoffset}
  />
</svg>

<style>
  .condura-thread {
    display: block;
    overflow: visible;
  }
  .condura-thread.h {
    width: 100%;
    height: 2px;
  }
  .condura-thread.v {
    width: 2px;
    height: 100%;
  }
  .condura-thread .glow {
    fill: none;
    stroke: var(--synapse-glow);
    stroke-width: 3;
    opacity: 0.18;
    filter: blur(3px);
  }
  .condura-thread .line {
    fill: none;
    stroke: var(--synapse);
    stroke-width: var(--thread-w);
    stroke-linecap: round;
    transition: stroke-dashoffset var(--dur-slow) var(--ease);
  }
</style>