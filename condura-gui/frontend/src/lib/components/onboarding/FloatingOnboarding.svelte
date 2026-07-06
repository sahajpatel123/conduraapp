<script lang="ts">
  /**
   * FloatingOnboarding — Living Paper card wizard wired to the daemon.
   *
   * Steps: EULA → Permissions → Power (probe) → Hotkey → First Breath
   * All legal/setup state is persisted via onboarding.svelte.ts RPCs.
   */
  import { onMount } from 'svelte'
  import { PaperSurface, BlurReveal } from '$lib/components/living'
  import EulaScreen from './EulaScreen.svelte'
  import PermissionCards from './PermissionCards.svelte'
  import PowerCards from './PowerCards.svelte'
  import HotkeyCard from './HotkeyCard.svelte'
  import FirstBreath from './FirstBreath.svelte'
  import { onboarding } from '../../stores/onboarding.svelte'
  import { ipc } from '../../ipc/client'

  interface Props {
    oncomplete: () => void
  }

  let { oncomplete }: Props = $props()

  type Step = 'eula' | 'permissions' | 'power' | 'hotkey' | 'done'

  let step = $state<Step>('eula')
  let booting = $state(true)
  let finishing = $state(false)

  const stepLabels: Record<Step, string> = {
    eula: 'Agreement',
    permissions: 'Permissions',
    power: 'Power Source',
    hotkey: 'Hotkey',
    done: 'Ready',
  }

  const stepOrder: Step[] = ['eula', 'permissions', 'power', 'hotkey', 'done']
  const currentStepIndex = $derived(stepOrder.indexOf(step))

  function daemonStepToUi(cs: string | undefined): Step {
    switch (cs) {
      case 'permissions':
        return 'permissions'
      case 'hotkey':
        return 'hotkey'
      case 'complete':
        return 'done'
      default:
        return 'eula'
    }
  }

  function advance(to?: Step) {
    if (to) {
      step = to
      return
    }
    const idx = stepOrder.indexOf(step)
    if (idx < stepOrder.length - 1) {
      step = stepOrder[idx + 1]
    }
  }

  async function onEulaAccepted() {
    advance('permissions')
  }

  async function onPermissionsContinue() {
    await onboarding.completePermissions()
    if (!onboarding.error) advance('power')
  }

  async function onPermissionsSkip() {
    await onboarding.skipStep('permissions')
    if (!onboarding.error) advance('power')
  }

  function onPowerNext() {
    advance('hotkey')
  }

  async function onHotkeyNext(combo: string) {
    onboarding.setHotkey(combo)
    await onboarding.saveHotkey()
    if (!onboarding.error) advance('done')
  }

  async function onHotkeySkip() {
    await onboarding.skipStep('hotkey')
    if (!onboarding.error) advance('done')
  }

  async function handleFinish() {
    if (finishing) return
    finishing = true
    try {
      const permissionsSkipped =
        onboarding.daemon?.steps?.permissions?.status === 'skipped'
      const hotkey =
        onboarding.daemon?.steps?.hotkey?.data ?? onboarding.hotkeyValue ?? ''
      const result = await onboarding.finish({
        hotkey,
        eula_version:
          onboarding.eulaVersion ||
          onboarding.daemon?.steps?.eula?.data ||
          'v1',
        permissions_skipped: permissionsSkipped,
      })
      if (!result || onboarding.error) return
      await ipc.firstRunComplete()
      oncomplete()
    } finally {
      finishing = false
    }
  }

  $effect(() => {
    if (step === 'power') {
      void onboarding.probePower()
    }
  })

  onMount(() => {
    void (async () => {
      await onboarding.sync()
      if (onboarding.isComplete) {
        oncomplete()
        return
      }
      step = daemonStepToUi(onboarding.daemon?.current_step)
      const savedHotkey = onboarding.daemon?.steps?.hotkey?.data
      if (savedHotkey) onboarding.setHotkey(savedHotkey)
      booting = false
    })()
  })
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
  <PaperSurface
    variant="raised"
    grain={true}
    padding="var(--lp-space-10) var(--lp-space-8)"
    radius="var(--lp-radius-lg)"
    style="
      max-width: 640px;
      width: 90vw;
      max-height: 85vh;
      overflow-y: auto;
      position: relative;
    "
  >
    {#if booting}
      <p style="
        text-align: center;
        font-family: var(--lp-font-sans);
        color: var(--lp-ink-mute);
        padding: var(--lp-space-8);
      ">Loading setup…</p>
    {:else}
      <!-- Step indicator -->
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
            <div
              style="
                width: 8px;
                height: 8px;
                border-radius: 50%;
                background: {i <= currentStepIndex ? 'var(--lp-synapse)' : 'var(--lp-ink-ghost)'};
                transition: background var(--lp-dur-normal) var(--lp-ease-thread);
                box-shadow: {i === currentStepIndex ? '0 0 6px var(--lp-synapse-glow)' : 'none'};
                position: relative;
                z-index: 1;
              "
            ></div>
            {#if i < stepOrder.length - 1}
              <div
                style="
                  width: 48px;
                  height: 1.5px;
                  background: linear-gradient(90deg,
                    {i < currentStepIndex ? 'var(--lp-synapse)' : 'var(--lp-ink-ghost)'},
                    {i < currentStepIndex ? 'var(--lp-synapse)' : 'var(--lp-ink-ghost)'}
                  );
                  opacity: 0.5;
                "
              ></div>
            {/if}
          </div>
        {/each}
      </div>

      <BlurReveal key={step} once={false} threshold={0}>
        {#if step === 'eula'}
          <EulaScreen onaccepted={onEulaAccepted} />
        {:else if step === 'permissions'}
          <PermissionCards
            onnext={() => void onPermissionsContinue()}
            onskip={() => void onPermissionsSkip()}
            busy={onboarding.busy}
          />
        {:else if step === 'power'}
          <PowerCards onnext={onPowerNext} onskip={onPowerNext} />
        {:else if step === 'hotkey'}
          <HotkeyCard
            onnext={(combo) => void onHotkeyNext(combo)}
            onskip={() => void onHotkeySkip()}
            busy={onboarding.busy}
            initialCombo={onboarding.hotkeyValue}
          />
        {:else if step === 'done'}
          <FirstBreath
            oncomplete={() => void handleFinish()}
            busy={finishing || onboarding.busy}
          />
        {/if}
      </BlurReveal>

      {#if onboarding.error}
        <p
          style="
            margin-top: var(--lp-space-4);
            text-align: center;
            color: var(--lp-danger);
            font-family: var(--lp-font-sans);
            font-size: var(--lp-text-body-sm);
          "
        >
          {onboarding.error}
        </p>
      {/if}

      <div style="text-align: center; margin-top: var(--lp-space-6);">
        <span
          style="
            font-family: var(--lp-font-mono);
            font-size: var(--lp-text-micro);
            letter-spacing: 0.08em;
            text-transform: uppercase;
            color: var(--lp-ink-faint);
          "
        >
          {stepLabels[step]} · {currentStepIndex + 1} of {stepOrder.length}
        </span>
      </div>
    {/if}
  </PaperSurface>
</div>
