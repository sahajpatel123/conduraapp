<!--
  BrandWordmark — Synaptic's typographic identity.

  Per spec §1, the Pulse IS Synaptic's vital sign. The wordmark complements
  it: a single, considered wordmark in the serif, set with old-style
  figures and a deliberate tracking. Used in the sidebar, about screen,
  onboarding, and as the visual identity throughout the app.

  Two variants:
    - default: full wordmark with the Pulse mark to the left
    - text-only: just the wordmark, no Pulse (for compact contexts)

  The wordmark is set in Source Serif 4 with carefully tuned letter-spacing.
  Never replace it with a generic font — the wordmark IS the brand.

  Props:
    variant   — 'default' | 'text-only'
    size      — 'sm' | 'md' | 'lg' | 'xl' (font size scale)
    pulseSize — pulse size when variant === 'default' (sm/md/lg)
    color     — optional text color override
-->
<script lang="ts">
  import Pulse from './Pulse.svelte';
  import type { PulseState } from './Pulse.svelte';

  interface Props {
    variant?: 'default' | 'text-only';
    size?: 'sm' | 'md' | 'lg' | 'xl';
    pulseSize?: 'sm' | 'md' | 'lg' | 'xl';
    pulseState?: PulseState;
    color?: string;
  }

  let {
    variant = 'default',
    size = 'md',
    pulseSize = 'sm',
    pulseState = 'idle',
    color,
  }: Props = $props();

  const SIZE_PX: Record<NonNullable<Props['size']>, number> = {
    sm: 14,
    md: 18,
    lg: 24,
    xl: 32,
  };
</script>

<div class="wordmark wordmark--{size}" class:wordmark--text-only={variant === 'text-only'} style={color ? `color: ${color};` : ''}>
  {#if variant === 'default'}
    <Pulse state={pulseState} size={pulseSize} label="Synaptic" />
  {/if}
  <span class="wordmark__text">Synaptic</span>
</div>

<style>
  .wordmark {
    display: inline-flex;
    align-items: center;
    gap: var(--space-3);
    color: var(--content-primary);
    user-select: none;
  }

  .wordmark--text-only {
    /* No pulse — pure text identity */
  }

  /* The wordmark text — serif, deliberate tracking, optical alignment */
  .wordmark__text {
    font-family: var(--font-serif);
    font-weight: 500;
    line-height: 1;
    /* Slight positive tracking — tighter than typical serif, but enough
       for the wordmark to breathe. The letterforms themselves are tight
       so over-tracking would make them feel disconnected. */
    letter-spacing: 0.005em;
    /* Optical alignment for serifs at small sizes */
    padding-bottom: 0.04em;
  }

  .wordmark--sm .wordmark__text { font-size: 14px; }
  .wordmark--md .wordmark__text { font-size: 18px; }
  .wordmark--lg .wordmark__text { font-size: 24px; }
  .wordmark--xl .wordmark__text { font-size: 32px; }

  /* Dark mode: the wordmark stays the same color, but slightly lighter
     to maintain contrast against the darker background. The semantic
     layer handles this — we just use --content-primary. */

  /* Subtle motion: when the wordmark is alone (text-only), add the
     faintest drift on the text. When paired with the Pulse, the
     Pulse carries the motion. */
  .wordmark--text-only .wordmark__text {
    animation: wordmark-drift 8s ease-in-out infinite;
  }

  @keyframes wordmark-drift {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.92; }
  }

  @media (prefers-reduced-motion: reduce) {
    .wordmark--text-only .wordmark__text {
      animation: none;
    }
  }
</style>