<script lang="ts">
  interface Props {
    tone: 'ok' | 'live' | 'bad'
  }
  let { tone }: Props = $props()
</script>

<svg class="arc" class:live={tone === 'live'} class:bad={tone === 'bad'} viewBox="0 0 1200 80" preserveAspectRatio="none" aria-hidden="true">
  <defs>
    <linearGradient id="mdArcGrad" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" stop-color="var(--md-live)" stop-opacity="0.25" />
      <stop offset="45%" stop-color="var(--md-cobalt)" stop-opacity="1" />
      <stop offset="100%" stop-color="var(--md-live)" stop-opacity="0.35" />
    </linearGradient>
    <filter id="mdArcGlow" x="-20%" y="-80%" width="140%" height="260%">
      <feGaussianBlur stdDeviation="3.5" result="b" />
      <feMerge>
        <feMergeNode in="b" />
        <feMergeNode in="SourceGraphic" />
      </feMerge>
    </filter>
  </defs>
  <path class="track" d="M 40 62 C 280 8, 920 8, 1160 62" fill="none" stroke="var(--md-line-strong)" stroke-width="2.5" />
  <path class="beam" d="M 40 62 C 280 8, 920 8, 1160 62" fill="none" stroke="url(#mdArcGrad)" stroke-width="3.25" stroke-linecap="round" filter="url(#mdArcGlow)" />
</svg>

<style>
  .arc { width: 100%; height: 56px; display: block; overflow: visible; }
  .beam { animation: md-glow 3.2s var(--md-ease) infinite; }
  .live .beam { animation: md-glow 1.1s var(--md-ease) infinite; }
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
  :global(:root[data-mode='dark']) .track {
    stroke: color-mix(in oklab, var(--md-ink) 18%, transparent);
  }
  :global(:root[data-mode='dark']) .beam {
    opacity: 1;
  }
  @media (prefers-reduced-motion: reduce) {
    .beam { animation: none !important; }
  }
</style>
