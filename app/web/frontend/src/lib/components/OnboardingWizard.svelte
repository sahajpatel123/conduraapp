<script lang="ts">
  import { onMount } from 'svelte'
  import { onboarding } from '../stores/onboarding.svelte'
  import EulaScreen from './onboarding/EulaScreen.svelte'
  import PermissionsScreen from './onboarding/PermissionsScreen.svelte'
  import HotkeyScreen from './onboarding/HotkeyScreen.svelte'
  import ReadyScreen from './onboarding/ReadyScreen.svelte'

  interface Props {
    // Invoked when the wizard completes. The parent (App) hides
    // the overlay; an optional route hash lets the Ready screen
    // deep-link into Settings.
    onComplete?: (route?: string) => void
  }
  let { onComplete }: Props = $props()

  const STEPS = ['eula', 'permissions', 'hotkey', 'complete'] as const

  onMount(() => {
    void onboarding.sync()
  })

  const activeIndex = $derived(STEPS.indexOf(onboarding.currentStep))

  function done(route?: string): void {
    onComplete?.(route)
  }
</script>

<div class="wizard-container">
  <div class="step-indicator">
    {#each STEPS as step, i}
      <div class="step-dot" class:active={i === activeIndex} class:past={i < activeIndex}></div>
    {/each}
  </div>

  {#if onboarding.loading && !onboarding.daemon}
    <div class="loading">Loading…</div>
  {:else if onboarding.currentStep === 'eula'}
    <EulaScreen />
  {:else if onboarding.currentStep === 'permissions'}
    <PermissionsScreen />
  {:else if onboarding.currentStep === 'hotkey'}
    <HotkeyScreen />
  {:else}
    <ReadyScreen onDone={done} />
  {/if}
</div>

<style>
  .wizard-container {
    position: relative;
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }
  .wizard-container::before {
    content: '';
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 800px;
    height: 800px;
    background: radial-gradient(circle, var(--color-accent) 0%, transparent 60%);
    opacity: 0.05;
    pointer-events: none;
    z-index: -1;
  }
  .step-indicator {
    position: absolute;
    top: var(--space-8);
    display: flex;
    gap: var(--space-2);
  }
  .step-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.1);
    transition: all var(--transition-base);
  }
  .step-dot.past {
    background: var(--color-accent);
    opacity: 0.5;
  }
  .step-dot.active {
    background: var(--color-accent);
    box-shadow: var(--shadow-glow);
    transform: scale(1.2);
  }
  .loading {
    color: var(--color-text-muted);
  }
</style>
