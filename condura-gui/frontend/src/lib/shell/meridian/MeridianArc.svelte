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
    stroke-linecap="butt"
  />
  <!-- Solid halt red — no glow, no endpoint nodes -->
  <path
    class="beam"
    d="M 36 58 C 260 14, 940 14, 1164 58"
    fill="none"
    stroke="var(--md-halt)"
    stroke-width="2.75"
    stroke-linecap="butt"
  />
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
  .bad .beam {
    animation: none;
    opacity: 1;
    stroke-width: 3;
  }
  .bad .track {
    stroke: color-mix(in oklab, var(--md-halt) 28%, var(--md-line-strong));
  }
  :global(:root[data-mode='dark']) .track {
    stroke: color-mix(in oklab, var(--md-ink) 18%, transparent);
  }
  @media (prefers-reduced-motion: reduce) {
    .beam {
      animation: none !important;
    }
    /* Breathing animation is the primary cue for the 'live' state
       (streaming, awaiting consent, reconnecting). When motion is
       reduced we lose the speed difference, so swap to a static
       marker — thicker stroke + brighter tint — so users can still
       distinguish 'live' from 'ok' at a glance. */
    .live .beam {
      stroke-width: 3.25;
      opacity: 1;
    }
    /* Halt already uses stroke-width 3 + animation:none — make it pop
       a touch more without relying on motion. */
    .bad .beam {
      stroke-width: 3.5;
    }
  }
</style>
