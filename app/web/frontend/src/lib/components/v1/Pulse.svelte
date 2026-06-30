<!--
  Pulse — the brand's vital sign.

  A single low-intensity dot that breathes at ~12 cycles/minute, brightening
  when the agent is thinking and dimming when it returns to listening.

  This is NOT a logo. It IS Synaptic. Across a coffee shop, this is what
  you'd see and feel.

  Locked: see docs/design-v1-redesign.md §1, §6.5.

  Props:
    state    — 'idle' | 'thinking' | 'awaiting' | 'error'
    size     — 'sm' (8px) | 'md' (12px) | 'lg' (16px) | 'xl' (24px)
    label    — optional aria-label, defaults to "Agent status"
    paused   — freeze the animation (for reduced-motion or user pref)
-->
<script lang="ts">
  import { PULSE_PARAMS, type PulseState } from '$tokens/motion';

  interface Props {
    state?: PulseState;
    size?: 'sm' | 'md' | 'lg' | 'xl';
    label?: string;
    paused?: boolean;
  }

  let { state = 'idle', size = 'md', label = 'Agent status', paused = false }: Props = $props();

  // Resolve the current animation params. Reactive to state.
  let params = $derived(PULSE_PARAMS[state]);

  // Size to px.
  const SIZE_PX: Record<NonNullable<Props['size']>, number> = {
    sm: 8,
    md: 12,
    lg: 16,
    xl: 24,
  };
  let diameter = $derived(SIZE_PX[size]);

  // Error state has a one-shot flash, then returns to idle visual.
  // We let CSS handle the flash via animation; the steady state is opacity 1.
  let color = $derived(
    state === 'error'
      ? 'var(--error-500)'
      : 'var(--content-accent)'
  );

  // The period in seconds (for CSS animation-duration).
  let periodSeconds = $derived(params.period / 1000);
</script>

<span
  class="pulse pulse--{state} pulse--{size}"
  class:pulse--paused={paused}
  role="status"
  aria-label={label}
  style="
    --pulse-diameter: {diameter}px;
    --pulse-period: {periodSeconds}s;
    --pulse-opacity-min: {params.opacity[0]};
    --pulse-opacity-max: {params.opacity[1]};
    --pulse-scale-min: {params.scale[0]};
    --pulse-scale-max: {params.scale[1]};
    --pulse-color: {color};
  "
></span>

<style>
  .pulse {
    display: inline-block;
    width: var(--pulse-diameter);
    height: var(--pulse-diameter);
    border-radius: var(--radius-pill);
    background-color: var(--pulse-color);
    /* The breath: gentle opacity + scale oscillation. */
    animation: pulse-breath var(--pulse-period) ease-in-out infinite;
    transform-origin: center;
    /* Accessibility: don't trap the focus ring inside this tiny dot. */
    flex-shrink: 0;
  }

  /* State-specific selectors exist for potential future
     per-state CSS overrides (e.g., a "thinking" glow). The actual
     animation parameters are driven by inline style + PULSE_PARAMS. */

  .pulse--error {
    /* One-shot flash then settle. The animation declaration on .pulse runs
       continuously; we use animation-iteration-count to limit to 1. */
    animation-iteration-count: 1;
  }

  .pulse--paused {
    animation-play-state: paused;
  }

  @keyframes pulse-breath {
    0%, 100% {
      opacity: var(--pulse-opacity-min);
      transform: scale(var(--pulse-scale-min));
    }
    50% {
      opacity: var(--pulse-opacity-max);
      transform: scale(var(--pulse-scale-max));
    }
  }

  /* Reduced motion: still convey presence, but no oscillation.
     Keep it visible at a steady opacity — the agent is here, just not breathing. */
  @media (prefers-reduced-motion: reduce) {
    .pulse {
      animation: none;
      opacity: var(--pulse-opacity-max);
      transform: scale(1);
    }
  }
</style>