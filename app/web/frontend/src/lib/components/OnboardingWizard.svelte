<script lang="ts">
  import { onMount } from 'svelte'
  import { onboarding } from '../stores/onboarding.svelte'
  import EulaScreen from './onboarding/EulaScreen.svelte'
  import PermissionsScreen from './onboarding/PermissionsScreen.svelte'
  import HotkeyScreen from './onboarding/HotkeyScreen.svelte'
  import ReadyScreen from './onboarding/ReadyScreen.svelte'
  import { t } from '../i18n'

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
  <div class="ambient" aria-hidden="true"></div>

  <div class="step-indicator">
    <div class="step-track" aria-hidden="true">
      <div
        class="step-fill"
        style="width: {STEPS.length > 1 ? (activeIndex / (STEPS.length - 1)) * 100 : 0}%"
      ></div>
    </div>
    {#each STEPS as step, i}
      <div class="step-dot" class:active={i === activeIndex} class:past={i < activeIndex}></div>
    {/each}
  </div>

  {#if onboarding.loading && !onboarding.daemon}
    <div class="loading">{t('onboarding.loading')}</div>
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
    overflow: hidden;
  }
  .ambient {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: -1;
    background:
      radial-gradient(560px circle at 50% 42%, var(--color-glow) 0%, transparent 65%),
      radial-gradient(420px circle at 30% 70%, rgba(139, 92, 246, 0.08) 0%, transparent 70%);
  }
  .step-indicator {
    position: absolute;
    top: var(--space-8);
    display: flex;
    gap: var(--space-3);
    align-items: center;
  }
  .step-track {
    position: absolute;
    inset: 0 0 auto 0;
    top: 50%;
    height: 1px;
    transform: translateY(-50%);
    background: var(--color-border);
  }
  .step-fill {
    height: 100%;
    background: var(--color-accent-gradient);
    box-shadow: 0 0 8px var(--color-glow-strong);
    transition: width var(--transition-spring-soft) var(--ease-out-expo);
  }
  .step-dot {
    position: relative;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--color-bg-active);
    border: 1px solid var(--color-border-strong);
    transition: all var(--transition-spring);
  }
  .step-dot.past {
    background: var(--color-accent);
    border-color: var(--color-accent);
    opacity: 0.55;
  }
  .step-dot.active {
    background: var(--color-accent);
    border-color: var(--color-accent-hover);
    box-shadow: var(--shadow-glow);
    transform: scale(1.35);
  }
  .loading {
    color: var(--color-text-muted);
    font-size: var(--size-md);
  }
</style>
