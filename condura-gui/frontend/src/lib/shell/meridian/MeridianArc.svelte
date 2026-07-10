<script lang="ts">
  interface Props {
    tone: 'ok' | 'live' | 'bad'
  }
  let { tone }: Props = $props()
</script>

<!--
  One continuous Meridian line:
  crown arc (same as before) → down both stage side margins (~25% page) fading out.
-->
<svg
  class="arc"
  class:live={tone === 'live'}
  class:bad={tone === 'bad'}
  viewBox="0 0 1200 420"
  preserveAspectRatio="none"
  aria-hidden="true"
>
  <defs>
    <linearGradient id="mdArcGrad" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" stop-color="var(--md-live)" stop-opacity="0.25" />
      <stop offset="45%" stop-color="var(--md-cobalt)" stop-opacity="1" />
      <stop offset="100%" stop-color="var(--md-live)" stop-opacity="0.35" />
    </linearGradient>
    <linearGradient id="mdArcSide" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="var(--md-cobalt)" stop-opacity="0.95" />
      <stop offset="18%" stop-color="var(--md-cobalt)" stop-opacity="0.55" />
      <stop offset="45%" stop-color="var(--md-cobalt)" stop-opacity="0.22" />
      <stop offset="72%" stop-color="var(--md-cobalt)" stop-opacity="0.06" />
      <stop offset="100%" stop-color="var(--md-cobalt)" stop-opacity="0" />
    </linearGradient>
    <linearGradient id="mdArcSideLive" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="var(--md-live)" stop-opacity="0.9" />
      <stop offset="20%" stop-color="var(--md-cobalt)" stop-opacity="0.5" />
      <stop offset="55%" stop-color="var(--md-cobalt)" stop-opacity="0.14" />
      <stop offset="100%" stop-color="var(--md-cobalt)" stop-opacity="0" />
    </linearGradient>
    <linearGradient id="mdArcSideBad" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="var(--md-halt)" stop-opacity="0.95" />
      <stop offset="25%" stop-color="var(--md-halt)" stop-opacity="0.4" />
      <stop offset="65%" stop-color="var(--md-halt)" stop-opacity="0.1" />
      <stop offset="100%" stop-color="var(--md-halt)" stop-opacity="0" />
    </linearGradient>
    <linearGradient id="mdArcTrackSide" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="var(--md-line-strong)" stop-opacity="0.9" />
      <stop offset="40%" stop-color="var(--md-line-strong)" stop-opacity="0.35" />
      <stop offset="100%" stop-color="var(--md-line-strong)" stop-opacity="0" />
    </linearGradient>
    <filter id="mdArcGlow" x="-20%" y="-40%" width="140%" height="180%">
      <feGaussianBlur stdDeviation="3.5" result="b" />
      <feMerge>
        <feMergeNode in="b" />
        <feMergeNode in="SourceGraphic" />
      </feMerge>
    </filter>
  </defs>

  <!-- Quiet track: crown + sides -->
  <path
    class="track track-crown"
    d="M 40 62 C 280 8, 920 8, 1160 62"
    fill="none"
    stroke="var(--md-line-strong)"
    stroke-width="2.5"
  />
  <path
    class="track track-side"
    d="M 40 62 L 40 400"
    fill="none"
    stroke="url(#mdArcTrackSide)"
    stroke-width="2.5"
    stroke-linecap="round"
  />
  <path
    class="track track-side"
    d="M 1160 62 L 1160 400"
    fill="none"
    stroke="url(#mdArcTrackSide)"
    stroke-width="2.5"
    stroke-linecap="round"
  />

  <!-- Living beam: same continuous stroke family -->
  <path
    class="beam beam-crown"
    d="M 40 62 C 280 8, 920 8, 1160 62"
    fill="none"
    stroke="url(#mdArcGrad)"
    stroke-width="3.25"
    stroke-linecap="round"
    filter="url(#mdArcGlow)"
  />
  <path
    class="beam beam-side"
    d="M 40 62 L 40 400"
    fill="none"
    stroke="url(#mdArcSide)"
    stroke-width="3.25"
    stroke-linecap="round"
  />
  <path
    class="beam beam-side"
    d="M 1160 62 L 1160 400"
    fill="none"
    stroke="url(#mdArcSide)"
    stroke-width="3.25"
    stroke-linecap="round"
  />
</svg>

<style>
  .arc {
    display: block;
    width: 100%;
    height: 100%;
    overflow: visible;
  }
  .beam-crown {
    animation: md-glow 3.2s var(--md-ease) infinite;
  }
  .live .beam-crown {
    animation: md-glow 1.1s var(--md-ease) infinite;
  }
  .live .beam-side {
    stroke: url(#mdArcSideLive);
  }
  .bad .beam-crown {
    stroke: var(--md-halt);
    filter: none;
    animation: none;
    opacity: 1;
    stroke-width: 3.5;
  }
  .bad .beam-side {
    stroke: url(#mdArcSideBad);
    filter: none;
    animation: none;
  }
  .bad .track-crown {
    stroke: color-mix(in oklab, var(--md-halt) 28%, var(--md-line-strong));
  }
  :global(:root[data-mode='dark']) .track-crown {
    stroke: color-mix(in oklab, var(--md-ink) 18%, transparent);
  }
  :global(:root[data-mode='dark']) .beam-crown {
    opacity: 1;
  }
  @media (prefers-reduced-motion: reduce) {
    .beam-crown {
      animation: none !important;
    }
  }
</style>
