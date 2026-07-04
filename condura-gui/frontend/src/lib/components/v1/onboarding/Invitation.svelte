<!--
  Onboarding · Screen 1 · Invitation

  Mood: a door, opening inward.
  Per spec §10.1: a single sentence in serif, centered. The word "three"
  is the only plum accent. No logo, no illustration, no marketing copy.

  The first screen IS the relationship. There's no skip.
-->
<script lang="ts">
  import Pulse from '../Pulse.svelte';
  import Button from '../Button.svelte';

  interface Props {
    onbegin?: () => void;
  }

  let { onbegin }: Props = $props();
</script>

<div class="screen">
  <div class="screen__pulse">
    <Pulse state="idle" size="xl" label="Synaptic waiting" />
  </div>

  <h1 class="screen__invitation">
    I'd like to help you use your computer. <span class="accent">Before I do, I need to ask your permission three times.</span>
  </h1>

  <div class="screen__actions">
    <Button variant="primary" size="lg" onclick={onbegin}>Begin.</Button>
  </div>
</div>

<style>
  .screen {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    padding: var(--space-9);
    background-color: var(--surface-base);
    color: var(--content-primary);
    gap: var(--space-9);
  }

  .screen__pulse {
    /* The pulse, large, breathing alone. This is the brand. */
    opacity: 0.85;
  }

  /* The sentence. Serif. Centered. Generous measure. */
  .screen__invitation {
    font-family: var(--font-serif);
    font-size: var(--text-h2-size);
    line-height: 1.4;
    font-weight: 400;
    text-align: center;
    max-width: 32em;
    color: var(--content-secondary);
    margin: 0;
    /* Subtle entrance */
    animation: fade-up var(--duration-slow) var(--ease-decelerate) both;
    animation-delay: 200ms;
  }

  /* The ONLY plum accent on the screen. */
  .accent {
    color: var(--content-accent);
  }

  .screen__actions {
    animation: fade-up var(--duration-slow) var(--ease-decelerate) both;
    animation-delay: 400ms;
  }

  @keyframes fade-up {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .screen__invitation,
    .screen__actions {
      animation: none;
    }
  }
</style>