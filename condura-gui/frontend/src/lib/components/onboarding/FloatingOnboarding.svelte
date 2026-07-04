<script lang="ts">
  /**
   * FloatingOnboarding — Card-based onboarding wizard.
   * 
   * Instead of a full-bleed wizard, this presents floating cards
   * over a blurred app background. Users select options by clicking
   * cards, creating a lightweight, choose-your-own-path feel.
   * 
   * Steps: Welcome → Permissions → Power Source → Hotkey → First Breath
   */
  import { PaperSurface, BlurReveal } from '$lib/components/living'
  import WelcomeCard from './WelcomeCard.svelte'
  import PermissionCards from './PermissionCards.svelte'
  import PowerCards from './PowerCards.svelte'
  import HotkeyCard from './HotkeyCard.svelte'
  import FirstBreath from './FirstBreath.svelte'

  interface Props {
    oncomplete: () => void
  }

  let { oncomplete }: Props = $props()

  type Step = 'welcome' | 'permissions' | 'power' | 'hotkey' | 'done'

  let step = $state<Step>('welcome')

  const stepLabels: Record<Step, string> = {
    welcome: 'Welcome',
    permissions: 'Permissions',
    power: 'Power Source',
    hotkey: 'Hotkey',
    done: 'Ready',
  }

  // Progress indicator
  const totalSteps = 5
  const stepOrder: Step[] = ['welcome', 'permissions', 'power', 'hotkey', 'done']
  const currentStepIndex = $derived(stepOrder.indexOf(step))

  function advance(to?: Step) {
    if (to) { step = to; return }
    const idx = stepOrder.indexOf(step)
    if (idx < stepOrder.length - 1) {
      step = stepOrder[idx + 1]
    }
  }

  function skip() {
    advance()
  }
</script>

<!-- Floating onboarding overlay — sits on top of the blurred app -->
<div
  class="lp"
  style="
    position: fixed;
    inset: 0;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(244, 239, 228, 0.75);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  "
>
  <!-- Floating card container -->
  <PaperSurface
    variant="raised"
    grain={true}
    padding="var(--lp-space-10) var(--lp-space-8)"
    radius="var(--lp-radius-lg)"
    style="
      max-width: 600px;
      width: 90vw;
      max-height: 85vh;
      overflow-y: auto;
      position: relative;
    "
  >
    <!-- Step indicator — thin synapse thread with dots -->
    <div
      style="
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 0;
        margin-bottom: var(--lp-space-8);
        position: relative;
        height: 16px;
      "
    >
      {#each stepOrder as s, i}
        <div style="display: flex; align-items: center;">
          <!-- Dot -->
          <div style="
            width: 8px; height: 8px;
            border-radius: 50%;
            background: {i <= currentStepIndex ? 'var(--lp-synapse)' : 'var(--lp-ink-ghost)'};
            transition: background var(--lp-dur-normal) var(--lp-ease-thread);
            box-shadow: {i === currentStepIndex ? '0 0 6px var(--lp-synapse-glow)' : 'none'};
            position: relative;
            z-index: 1;
          "></div>
          <!-- Connector line -->
          {#if i < stepOrder.length - 1}
            <div style="
              width: 48px; height: 1.5px;
              background: linear-gradient(90deg,
                {i < currentStepIndex ? 'var(--lp-synapse)' : 'var(--lp-ink-ghost)'},
                {i < currentStepIndex ? 'var(--lp-synapse)' : 'var(--lp-ink-ghost)'}
              );
              opacity: 0.5;
            "></div>
          {/if}
        </div>
      {/each}
    </div>

    <BlurReveal key={step} once={false} threshold={0}>
      <!-- Step content -->
      {#if step === 'welcome'}
        <WelcomeCard onnext={() => advance('permissions')} />
      {:else if step === 'permissions'}
        <PermissionCards onnext={() => advance('power')} onskip={() => advance('power')} />
      {:else if step === 'power'}
        <PowerCards onnext={(_choice: string) => advance('hotkey')} onskip={() => advance('hotkey')} />
      {:else if step === 'hotkey'}
        <HotkeyCard onnext={(_combo: string) => advance('done')} onskip={() => advance('done')} />
      {:else if step === 'done'}
        <FirstBreath oncomplete={oncomplete} />
      {/if}
    </BlurReveal>

    <!-- Step label footer -->
    <div style="
      text-align: center;
      margin-top: var(--lp-space-6);
    ">
      <span style="
        font-family: var(--lp-font-mono);
        font-size: var(--lp-text-micro);
        letter-spacing: 0.08em;
        text-transform: uppercase;
        color: var(--lp-ink-faint);
      ">
        {stepLabels[step]} · {currentStepIndex + 1} of {totalSteps}
      </span>
    </div>
  </PaperSurface>
</div>
