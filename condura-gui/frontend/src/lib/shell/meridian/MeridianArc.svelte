<script lang="ts">
  interface Props {
    tone: 'ok' | 'live' | 'bad'
  }
  let { tone }: Props = $props()
</script>

<svg class="arc" class:live={tone === 'live'} class:bad={tone === 'bad'} viewBox="0 0 1200 80" preserveAspectRatio="none" aria-hidden="true">
  <defs>
    <linearGradient id="mdArcGrad" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" stop-color="var(--md-halt)" stop-opacity="0.4" />
      <stop offset="42%" stop-color="var(--md-halt)" stop-opacity="1" />
      <stop offset="100%" stop-color="var(--md-ember, var(--md-halt))" stop-opacity="0.55" />
    </linearGradient>
    <!-- Glow bias downward (positive y) so blur does not climb into the header Jump bar. -->
    <filter id="mdArcGlow" x="-15%" y="-15%" width="130%" height="160%" color-interpolation-filters="sRGB">
      <feGaussianBlur stdDeviation="2.5" result="b" />
      <feOffset in="b" dy="2" result="bo" />
      <feMerge>
        <feMergeNode in="bo" />
        <feMergeNode in="SourceGraphic" />
      </feMerge>
    </filter>
  </defs>
  <path class="track" d="M 40 62 C 280 8, 920 8, 1160 62" fill="none" stroke="var(--md-line-strong)" stroke-width="2.5" />
  <path class="beam" d="M 40 62 C 280 8, 920 8, 1160 62" fill="none" stroke="url(#mdArcGrad)" stroke-width="3.4" stroke-linecap="round" filter="url(#mdArcGlow)" />
  <circle class="node left" cx="40" cy="62" r="4.5" />
  <circle class="node right" cx="1160" cy="62" r="4.5" />
</svg>

<style>
  .arc { width: 100%; height: 56px; display: block; overflow: hidden; }
  /* Soft opacity pulse only — no drop-shadow (that was bleeding red onto Jump). */
  .beam { animation: md-arc-pulse 3.2s var(--md-ease) infinite; }
  .live .beam { animation: md-arc-pulse 1.05s var(--md-ease) infinite; }
  @keyframes md-arc-pulse {
    0%, 100% { opacity: 0.78; }
    50% { opacity: 1; }
  }
  .node {
    fill: var(--md-surface);
    stroke: var(--md-halt);
    stroke-width: 2;
    opacity: 0.85;
  }
  .live .node {
    fill: var(--md-halt);
    stroke: color-mix(in oklab, var(--md-halt) 40%, #fff);
    animation: md-breathe 1.4s var(--md-ease) infinite;
  }
  .bad .beam {
    stroke: var(--md-halt);
    filter: none;
    animation: none;
    opacity: 1;
    stroke-width: 3.5;
  }
  .bad .track {
    stroke: color-mix(in oklab, var(--md-halt) 28%, var(--md-line-strong));
  }
  .bad .node {
    fill: var(--md-halt);
    stroke: color-mix(in oklab, var(--md-halt) 40%, #fff);
    animation: none;
  }
  :global(:root[data-mode='dark']) .track {
    stroke: color-mix(in oklab, var(--md-ink) 18%, transparent);
  }
  :global(:root[data-mode='dark']) .beam {
    opacity: 1;
  }
  @media (prefers-reduced-motion: reduce) {
    .beam, .node { animation: none !important; }
  }
</style>
