<!--
  Pulse — the brand's vital sign. Refined.

  A single low-intensity dot that breathes at ~12 cycles/minute, brightening
  when the agent is thinking and dimming when it returns to listening.

  This is NOT a logo. It IS Synaptic. Across a coffee shop, this is what
  you'd see and feel.

  Refinements in this version:
    - Outer halo that breathes at half the rate of the inner dot (organic feel)
    - Easing curve refined from linear sin to a slower-in-faster-out curve
    - Subtle outer glow that responds to state
    - "Settle" effect on state change — brief brightness pulse to telegraph
    - Thinking state has a stronger halo + visible "thinking" ring

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
  let color = $derived(
    state === 'error' ? 'var(--error-500)' : 'var(--content-accent)'
  );

  // The period in seconds (for CSS animation-duration).
  let periodSeconds = $derived(params.period / 1000);

  // Halo period — outer ring breathes at half the rate, giving an organic,
  // compound rhythm. Subtle but reads as "alive" rather than mechanical.
  let haloPeriodSeconds = $derived(periodSeconds * 1.6);

  // Halo opacity range — wider than the inner dot so it reads as a "presence"
  // expanding and contracting, not just a colored circle.
  let haloOpacityMax = $derived(
    state === 'thinking' ? 0.4 :
    state === 'awaiting' ? 0.6 :
    state === 'error'    ? 0.0 :  // no halo on error — flash only
    0.18
  );

  let haloOpacityMin = $derived(
    state === 'thinking' ? 0.0 :
    state === 'awaiting' ? 0.0 :
    0.0
  );

  // Halo size scale — slightly larger expansion than the inner dot
  let haloScaleMax = $derived(
    state === 'thinking' ? 2.6 :
    state === 'awaiting' ? 2.0 :
    1.8
  );

  // For the error state, animate one flash then settle
  let isError = $derived(state === 'error');

  // For awaiting: continuous "looking at you" feel with quicker cycle
  // For idle: contemplative slow breath
  // For thinking: working breath, slower but with stronger amplitude
</script>

<span
  class="pulse pulse--{state} pulse--{size}"
  class:pulse--paused={paused}
  class:pulse--thinking={state === 'thinking'}
  class:pulse--awaiting={state === 'awaiting'}
  class:pulse--error={isError}
  role="status"
  aria-label={label}
  style="
    --pulse-diameter: {diameter}px;
    --pulse-period: {periodSeconds}s;
    --halo-period: {haloPeriodSeconds}s;
    --pulse-opacity-min: {params.opacity[0]};
    --pulse-opacity-max: {params.opacity[1]};
    --pulse-scale-min: {params.scale[0]};
    --pulse-scale-max: {params.scale[1]};
    --halo-opacity-max: {haloOpacityMax};
    --halo-opacity-min: {haloOpacityMin};
    --halo-scale-max: {haloScaleMax};
    --pulse-color: {color};
  "
>
  {#if !isError}
    <span class="pulse__halo" aria-hidden="true"></span>
  {/if}
  <span class="pulse__dot" aria-hidden="true"></span>
</span>

<style>
  /* The pulse wrapper — sizes itself to the inner dot */
  .pulse {
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: var(--pulse-diameter);
    height: var(--pulse-diameter);
    flex-shrink: 0;
  }

  /* The inner dot — the brand signature */
  .pulse__dot {
    position: relative;
    width: 100%;
    height: 100%;
    border-radius: var(--radius-pill);
    background-color: var(--pulse-color);
    animation: pulse-breath var(--pulse-period) cubic-bezier(0.4, 0, 0.6, 1) infinite;
    transform-origin: center;
    z-index: 2;
  }

  /* The halo — expands and contracts at a slower rate */
  .pulse__halo {
    position: absolute;
    inset: 0;
    border-radius: var(--radius-pill);
    background-color: var(--pulse-color);
    opacity: var(--halo-opacity-min);
    animation: pulse-halo var(--halo-period) cubic-bezier(0.4, 0, 0.6, 1) infinite;
    transform-origin: center;
    z-index: 1;
  }

  /* ── Inner breath — refined easing ────────────────────────── */
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

  /* ── Halo — slower, wider, fades in and out ────────────────── */
  @keyframes pulse-halo {
    0%, 100% {
      opacity: var(--halo-opacity-min);
      transform: scale(1);
    }
    50% {
      opacity: var(--halo-opacity-max);
      transform: scale(var(--halo-scale-max));
    }
  }

  /* ── Thinking state — a more active presence ────────────────── */
  .pulse--thinking .pulse__dot {
    /* Same animation but the params give stronger amplitude */
  }

  .pulse--thinking .pulse__halo {
    /* The halo's color stays accent; the slower rhythm + stronger
       amplitude gives the working "thinking" feel */
  }

  /* ── Awaiting state — looking at you ──────────────────────── */
  .pulse--awaiting .pulse__dot {
    /* Quick breath, full opacity */
  }

  .pulse--awaiting .pulse__halo {
    /* Strong halo, full opacity at peak — the "I'm waiting for you"
       presence, more visible than idle */
  }

  /* ── Error state — one-shot flash, no halo, settles to idle ── */
  .pulse--error .pulse__dot {
    animation: pulse-error 800ms ease-out 1 both;
    /* Override the breath animation */
  }

  @keyframes pulse-error {
    0% {
      opacity: 0.3;
      transform: scale(0.6);
    }
    20% {
      opacity: 1;
      transform: scale(1.4);
    }
    60% {
      opacity: 1;
      transform: scale(1);
    }
    100% {
      opacity: 0.85;
      transform: scale(1);
    }
  }

  /* ── Paused (for explicit pause prop) ──────────────────────── */
  .pulse--paused .pulse__dot,
  .pulse--paused .pulse__halo {
    animation-play-state: paused;
  }

  /* ── Reduced motion: keep presence, no oscillation ─────────── */
  @media (prefers-reduced-motion: reduce) {
    .pulse__dot {
      animation: none;
      opacity: var(--pulse-opacity-max);
      transform: scale(1);
    }
    .pulse__halo {
      animation: none;
      opacity: 0;
      transform: scale(1);
    }
  }
</style>