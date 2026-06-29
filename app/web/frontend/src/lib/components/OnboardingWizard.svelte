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

  type StepId = 'eula' | 'permissions' | 'hotkey' | 'complete'
  const STEPS: StepId[] = ['eula', 'permissions', 'hotkey', 'complete']

  // The id of the previous step, used to drive outgoing direction.
  // We default to the current step so the first render is non-animated.
  let previousStep = $state<StepId>(onboarding.currentStep)
  let displayStep = $state<StepId>(onboarding.currentStep)

  onMount(() => {
    void onboarding.sync()
  })

  // The active index is the position of the *current* daemon step
  // in the wizard. It drives the dot fill and the track.
  const activeIndex = $derived(STEPS.indexOf(onboarding.currentStep))

  // When the current step changes, we animate out the previous step
  // (slide-left) and animate in the new one (slide-up). displayStep
  // lags the daemon by one tick so the outgoing frame has a chance
  // to render before we swap.
  $effect(() => {
    const next = onboarding.currentStep
    if (next === displayStep) return
    previousStep = displayStep
    // Use rAF so the previous class is applied before the new one
    // overrides it; this gives the slide-out a frame to paint.
    requestAnimationFrame(() => {
      displayStep = next
    })
  })

  // A step "exits to the left" when the new step has a higher
  // index; otherwise it exits to the right. (Pure cosmetic.)
  const direction = $derived(
    STEPS.indexOf(displayStep) >= STEPS.indexOf(previousStep) ? 'forward' : 'backward'
  )

  function done(route?: string): void {
    onComplete?.(route)
  }
</script>

<div class="wizard-container" data-direction={direction}>
  <div class="halo" aria-hidden="true"></div>

  <div class="step-indicator" role="list" aria-label={t('onboarding.loading')}>
    {#each STEPS as step, i (step)}
      <div
        class="step-node"
        class:past={i < activeIndex}
        class:active={i === activeIndex}
        role="listitem"
        aria-current={i === activeIndex ? 'step' : undefined}
      >
        <span class="step-dot"></span>
        <span class="step-num">{i + 1}</span>
      </div>
    {/each}
  </div>

  <div class="stage" data-direction={direction}>
    {#key displayStep}
      <div class="step-frame" data-step={displayStep}>
        {#if onboarding.loading && !onboarding.daemon}
          <div class="loading">{t('onboarding.loading')}</div>
        {:else if displayStep === 'eula'}
          <EulaScreen />
        {:else if displayStep === 'permissions'}
          <PermissionsScreen />
        {:else if displayStep === 'hotkey'}
          <HotkeyScreen />
        {:else}
          <ReadyScreen onDone={done} />
        {/if}
      </div>
    {/key}
  </div>
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
    padding: var(--space-7) var(--space-5);
  }

  /* Radial halo behind the active step. Sits behind the stage
     so it lights the content, not over it. */
  .halo {
    position: absolute;
    inset: 0;
    pointer-events: none;
    background:
      radial-gradient(
        620px circle at 50% 38%,
        var(--accent-glow) 0%,
        transparent 62%
      ),
      radial-gradient(
        460px circle at 22% 78%,
        var(--accent-faint) 0%,
        transparent 70%
      );
    animation: breathe-soft 9s var(--ease-in-out-quart) infinite;
    z-index: 0;
  }

  .step-indicator {
    position: relative;
    z-index: 2;
    display: flex;
    gap: var(--space-6);
    align-items: center;
    margin-bottom: var(--space-7);
  }

  .step-node {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: var(--surface-2);
    border: 1px solid var(--border-strong);
    color: var(--text-faint);
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    font-weight: var(--weight-semibold);
    letter-spacing: var(--tracking-wider);
    transition:
      background var(--transition-base),
      border-color var(--transition-base),
      color var(--transition-base),
      box-shadow var(--transition-base),
      transform var(--transition-base) var(--ease-spring);
  }

  /* The hairline ring around the active dot. */
  .step-node::after {
    content: '';
    position: absolute;
    inset: -5px;
    border-radius: 50%;
    border: 1px solid transparent;
    transition: border-color var(--transition-base);
  }

  .step-num {
    position: relative;
    z-index: 1;
  }

  .step-node.past {
    background: var(--accent-soft);
    border-color: var(--accent-soft);
    color: var(--accent);
  }

  .step-node.active {
    background: var(--accent-gradient);
    border-color: var(--accent);
    color: var(--text-inverse);
    transform: scale(1.05);
    box-shadow: var(--shadow-glow);
  }
  .step-node.active::after {
    border-color: var(--accent-glow);
    animation: breathe 2.4s var(--ease-in-out-quart) infinite;
  }

  .stage {
    position: relative;
    z-index: 1;
    width: 100%;
    display: flex;
    align-items: flex-start;
    justify-content: center;
  }

  .step-frame {
    width: 100%;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    animation: step-in var(--transition-slow) var(--ease-out-expo) both;
  }

  /* Outgoing frame, when the wizard moves forward: slide left. */
  .wizard-container[data-direction='forward'] .step-frame {
    animation: step-in-forward var(--transition-slow) var(--ease-out-expo) both;
  }
  /* Outgoing frame, when the wizard moves backward: slide right. */
  .wizard-container[data-direction='backward'] .step-frame {
    animation: step-in-backward var(--transition-slow) var(--ease-out-expo) both;
  }

  @keyframes step-in {
    from { opacity: 0; transform: translateY(18px); }
    to   { opacity: 1; transform: translateY(0); }
  }
  @keyframes step-in-forward {
    from { opacity: 0; transform: translateX(36px); }
    to   { opacity: 1; transform: translateX(0); }
  }
  @keyframes step-in-backward {
    from { opacity: 0; transform: translateX(-36px); }
    to   { opacity: 1; transform: translateX(0); }
  }

  .loading {
    color: var(--text-muted);
    font-size: var(--size-md);
  }
</style>
