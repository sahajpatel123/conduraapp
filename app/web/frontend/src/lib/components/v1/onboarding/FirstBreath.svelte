<!--
  Onboarding · First Breath (closing moment)

  Per spec §10.6: the wizard dissolves over 400ms. In its place: the
  floating command surface, empty, with the pulse at center, breathing.
  A single serif line fades in over 1s: "I'm here. Type when you're
  ready." Then fades to 60% opacity after 4s.

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
  let fadeOut = $state(false);

  onMount(() => {
    // After 4 seconds, fade the line to 60% opacity
    setTimeout(() => {
      fadeOut = true;
    }, 4000);

    // After 8 seconds, complete
    setTimeout(() => {
      visible = false;
      oncomplete?.();
    }, 8000);
  });
</script>

{#if visible}
  <div class="breath" role="status" aria-label="Synaptic ready">
    <div class="breath__pulse">
      <Pulse state="idle" size="xl" label="Synaptic ready" />
    </div>
    <p class="breath__line" class:breath__line--faded={fadeOut}>
      I'm here. Type when you're ready.
    </p>
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
  }

  @keyframes breath-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  .breath__pulse {
    /* The pulse, large, at center. This is what remains. */
    opacity: 0.9;
  }

  .breath__line {
    font-family: var(--font-serif);
    font-size: var(--text-h3-size);
    font-weight: 400;
    color: var(--content-primary);
    margin: 0;
    opacity: 0;
    animation: line-fade-in 1s var(--ease-decelerate) 600ms forwards;
    transition: opacity 1.5s var(--ease-decelerate);
    text-align: center;
  }

  .breath__line--faded {
    opacity: 0.6 !important;
  }

  @keyframes line-fade-in {
    from { opacity: 0; transform: translateY(4px); }
    to { opacity: 1; transform: translateY(0); }
  }

  @media (prefers-reduced-motion: reduce) {
    .breath {
      animation: none;
    }
    .breath__line {
      animation: none;
      opacity: 1;
    }
  }
</style>