<script lang="ts">
  interface Props {
    tone: 'ok' | 'live' | 'bad'
  }
  let { tone }: Props = $props()
</script>

<svg
  class="arc"
  class:live={tone === 'live'}
  class:bad={tone === 'bad'}
  viewBox="0 0 1200 72"
  preserveAspectRatio="none"
  aria-hidden="true"
>
  <!-- Quiet track under the live beam — original bend restored -->
  <path
    class="track"
    d="M 36 58 C 260 14, 940 14, 1164 58"
    fill="none"
    stroke="var(--md-line-strong)"
    stroke-width="2"
  />
  <!-- Solid halt red — no glow, no faded gradient -->
  <path
    class="beam"
    d="M 36 58 C 260 14, 940 14, 1164 58"
    fill="none"
    stroke="var(--md-halt)"
    stroke-width="2.75"
    stroke-linecap="round"
  />
  <circle class="node left" cx="36" cy="58" r="4.25" />
  <circle class="node right" cx="1164" cy="58" r="4.25" />
</svg>

<style>
  .arc {
    width: 100%;
    height: 52px;
    display: block;
    overflow: visible;
  }
  .beam {
    animation: md-arc-breathe 3.4s var(--md-ease) infinite;
  }
  .live .beam {
    animation: md-arc-breathe 1.1s var(--md-ease) infinite;
  }
  @keyframes md-arc-breathe {
    0%,
    100% {
      opacity: 0.92;
    }
    50% {
      opacity: 1;
    }
  }
  .node {
    fill: var(--md-surface);
    stroke: var(--md-halt);
    stroke-width: 2;
  }
  .live .node {
    fill: var(--md-halt);
    stroke: color-mix(in oklab, var(--md-halt) 55%, #fff);
    animation: md-breathe 1.4s var(--md-ease) infinite;
  }
  .bad .beam {
    animation: none;
    opacity: 1;
    stroke-width: 3;
  }
  .bad .track {
    stroke: color-mix(in oklab, var(--md-halt) 28%, var(--md-line-strong));
  }
  .bad .node {
    fill: var(--md-halt);
    stroke: color-mix(in oklab, var(--md-halt) 55%, #fff);
    animation: none;
  }
  :global(:root[data-mode='dark']) .track {
    stroke: color-mix(in oklab, var(--md-ink) 18%, transparent);
  }
  @media (prefers-reduced-motion: reduce) {
    .beam,
    .node {
      animation: none !important;
    }
  }
</style>
