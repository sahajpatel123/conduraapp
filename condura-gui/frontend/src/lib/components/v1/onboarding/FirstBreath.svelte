<!--
  Onboarding · First Breath (closing moment) — refined

  Per spec §10.6: the wizard dissolves. In its place: the floating command
  surface idea, with the pulse at center. A single serif line fades in.

  Refinements in this version:
    - Pulse ARRIVES from 0.6x scale with brightness bloom, then settles
    - Ambient plum tint that fades in behind the pulse, then fades back
    - Two-line moment: first "I'm here" → fades to a quieter hint
    - Subtle paper-warm-tint grain in the background
    - Esc to skip (user can dismiss this moment if they want)

  This is the first thing the user remembers about Synaptic. It didn't
  ask for an email.
-->
<script lang="ts">
  import Pulse from '../Pulse.svelte';
  import { onMount } from 'svelte';

  interface Props {
    oncomplete?: () => void;
  }

  let { oncomplete }: Props = $props();

  let visible = $state(true);
  let phase = $state<'arriving' | 'settled' | 'fading'>('arriving');
  let pulseReveal = $state(false);
  let bloomIntensity = $state(0);

  onMount(() => {
    // Phase 1: Pulse arrives (0-800ms)
    pulseReveal = true;

    // Phase 2: Bloom peak then settle (800-1500ms)
    setTimeout(() => {
      bloomIntensity = 1;
    }, 600);
    setTimeout(() => {
      bloomIntensity = 0.4; // settle to a faint glow
    }, 1500);

    // Phase 3: Line fades in (1500-2500ms)
    setTimeout(() => {
      phase = 'settled';
    }, 1200);

    // Phase 4: After 5 seconds, fade the line to 60% opacity and show the hint
    setTimeout(() => {
      phase = 'fading';
    }, 5000);

    // Phase 5: Complete after 8 seconds
    setTimeout(() => {
      visible = false;
      oncomplete?.();
    }, 8000);
  });

  function handleKey(e: KeyboardEvent) {
    if (e.key === 'Escape' || e.key === 'Enter' || e.key === ' ') {
      visible = false;
      oncomplete?.();
    }
  }
</script>

<svelte:window onkeydown={handleKey} />

{#if visible}
  <div class="breath" role="status" aria-label="Synaptic ready">
    <!-- The ambient plum bloom — concentrated behind the pulse, then fades -->
    <div class="breath__bloom" style="opacity: {bloomIntensity};" aria-hidden="true"></div>

    <!-- Subtle grain texture — barely visible -->
    <div class="breath__grain" aria-hidden="true"></div>

    <!-- The pulse, large, at center. ARRIVES, then settles to idle. -->
    <div class="breath__pulse-wrap" class:breath__pulse-wrap--revealed={pulseReveal}>
      <Pulse state="idle" size="xl" label="Synaptic ready" />
    </div>

    <!-- The first line: "I'm here." -->
    <p class="breath__line" data-phase={phase}>
      <span class="breath__line-primary">I'm here.</span>
      {#if phase === 'fading'}
        <span class="breath__line-secondary">Type when you're ready.</span>
      {/if}
    </p>

    <!-- Bottom hint — keyboard shortcut, subtle -->
    <div class="breath__hint" aria-hidden="true">
      <kbd>⌘K</kbd> <span>anytime</span>
    </div>
  </div>
{/if}

<style>
  .breath {
    position: fixed;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-7);
    background-color: var(--surface-base);
    color: var(--content-primary);
    z-index: var(--z-modal);
    animation: breath-in var(--duration-slow) var(--ease-decelerate) both;
    overflow: hidden;
  }

  @keyframes breath-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  /* The plum bloom — a soft radial glow that concentrates behind the pulse,
     peaks at 600-1500ms, then settles to a faint constant presence. */
  .breath__bloom {
    position: absolute;
    top: 45%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 600px;
    height: 600px;
    background: radial-gradient(
      circle,
      var(--content-accent) 0%,
      transparent 65%
    );
    transition: opacity 800ms var(--ease-decelerate);
    pointer-events: none;
    mix-blend-mode: multiply;
    opacity: 0;
  }

  /* Subtle grain — barely visible texture, adds depth */
  .breath__grain {
    position: absolute;
    inset: 0;
    background-image: radial-gradient(circle at 20% 30%, rgba(14, 16, 20, 0.012) 0%, transparent 60%),
                      radial-gradient(circle at 80% 70%, rgba(14, 16, 20, 0.012) 0%, transparent 60%);
    pointer-events: none;
  }

  /* The pulse — arrives from 0.6x scale, settles */
  .breath__pulse-wrap {
    transform: scale(0.6);
    opacity: 0;
    transition:
      transform 600ms var(--ease-decelerate),
      opacity 400ms var(--ease-decelerate);
  }

  .breath__pulse-wrap--revealed {
    transform: scale(1);
    opacity: 1;
  }

  /* The text — composed of two lines, transitions between them */
  .breath__line {
    margin: 0;
    text-align: center;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .breath__line-primary {
    font-family: var(--font-serif);
    font-size: var(--text-h3-size);
    font-weight: 400;
    color: var(--content-primary);
    opacity: 0;
    animation: line-fade-in 1.2s var(--ease-decelerate) 600ms forwards;
    transition: opacity 1.5s var(--ease-decelerate);
  }

  /* When phase = 'fading', the primary line dims to 60% */
  .breath__line[data-phase="fading"] .breath__line-primary {
    opacity: 0.5 !important;
  }

  /* The secondary line — appears with a subtle motion */
  .breath__line-secondary {
    font-family: var(--font-serif);
    font-size: var(--text-body-lg-size);
    font-style: italic;
    color: var(--content-secondary);
    opacity: 0;
    animation: line-fade-in 1.2s var(--ease-decelerate) forwards;
  }

  @keyframes line-fade-in {
    from {
      opacity: 0;
      transform: translateY(6px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  /* The keyboard hint at the bottom */
  .breath__hint {
    position: absolute;
    bottom: var(--space-9);
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    letter-spacing: 0.04em;
    opacity: 0;
    animation: line-fade-in 1s var(--ease-decelerate) 1.8s forwards;
  }

  .breath__hint kbd {
    font-family: var(--font-mono);
    background-color: var(--paper-warm-50);
    border: 1px solid var(--border-default);
    padding: 2px 6px;
    border-radius: var(--radius-xs);
    color: var(--content-primary);
  }

  @media (prefers-reduced-motion: reduce) {
    .breath {
      animation: none;
    }
    .breath__pulse-wrap {
      transform: scale(1);
      opacity: 1;
      transition: none;
    }
    .breath__line-primary,
    .breath__line-secondary,
    .breath__hint {
      animation: none;
      opacity: 1;
      transform: none;
    }
  }
</style>