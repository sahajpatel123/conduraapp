<!--
  OnboardingWizard — orchestrator for the 5 screens + First Breath.

  Per spec §10 + §11: a 5-screen wizard (Invitation → EULA → Eyes →
  Power source → Hotkey) with a First Breath closing moment. State machine
  is local; each screen receives props and emits events.

  This is a STANDALONE version. The daemon-backed wizard with real
  onboarding.* RPCs lives at the original path. This one is for design
  review and demo.

  Props:
    oncomplete — fired when the user finishes all 5 screens + First Breath
-->
<script lang="ts">
  import Invitation from './Invitation.svelte';
  import Eula from './Eula.svelte';
  import Eyes from './Eyes.svelte';
  import PowerSource from './PowerSource.svelte';
  import Hotkey from './Hotkey.svelte';
  import FirstBreath from './FirstBreath.svelte';

  type Screen = 'invitation' | 'eula' | 'eyes' | 'power' | 'hotkey' | 'breath';

  interface Props {
    oncomplete?: () => void;
  }

  let { oncomplete }: Props = $props();

  // State machine
  let current = $state<Screen>('invitation');
  let previous = $state<Screen | null>(null);

  // Persisted state across screens
  let eulaAccepted = $state(false);
  let accessibilityGranted = $state(false);
  let screenRecordingGranted = $state(false);
  let powerChoice = $state<'claude-pro' | 'chatgpt-plus' | 'api-key' | 'ollama' | null>(null);
  let apiKey = $state('');
  let hotkey = $state('');
  let voiceWakeEnabled = $state(false);
  let ollamaDetected = $state(false);

  // Detect Ollama (would normally come from daemon)
  $effect(() => {
    // Stub for demo — would call daemon RPC in production
    ollamaDetected = false;
  });

  // Navigation helpers
  function goTo(screen: Screen) {
    previous = current;
    current = screen;
  }

  function goBack() {
    if (current === 'eula') goTo('invitation');
    else if (current === 'eyes') goTo('eula');
    else if (current === 'power') goTo('eyes');
    else if (current === 'hotkey') goTo('power');
  }

  // Screen transitions
  function handleBegin() { goTo('eula'); }
  function handleEulaAccept() { eulaAccepted = true; goTo('eyes'); }
  function handleEyesContinue() { goTo('power'); }
  function handleEyesSkip() { goTo('power'); }
  function handlePowerSkip() { goTo('hotkey'); }
  function handleHotkeyContinue() { goTo('breath'); }
  function handleBreathComplete() { oncomplete?.(); }

  // Track transition direction for animation
  let direction = $derived<'forward' | 'back'>(
    previous === null
      ? 'forward'
      : ['invitation', 'eula', 'eyes', 'power', 'hotkey', 'breath'].indexOf(current) >
        ['invitation', 'eula', 'eyes', 'power', 'hotkey', 'breath'].indexOf(previous)
        ? 'forward'
        : 'back'
  );
</script>

<div class="wizard" data-direction={direction} data-screen={current}>
  {#if current === 'invitation'}
    <div class="wizard__screen">
      <Invitation onbegin={handleBegin} />
    </div>
  {:else if current === 'eula'}
    <div class="wizard__screen">
      <Eula onaccept={handleEulaAccept} onback={goBack} />
    </div>
  {:else if current === 'eyes'}
    <div class="wizard__screen">
      <Eyes
        accessibilityGranted={accessibilityGranted}
        screenRecordingGranted={screenRecordingGranted}
        ongrantAccessibility={() => { accessibilityGranted = true; }}
        ongrantScreenRecording={() => { screenRecordingGranted = true; }}
        oncontinue={handleEyesContinue}
        onskip={handleEyesSkip}
        onback={goBack}
      />
    </div>
  {:else if current === 'power'}
    <div class="wizard__screen">
      <PowerSource
        selected={powerChoice}
        apiKey={apiKey}
        ollamaDetected={ollamaDetected}
        onselect={(c) => powerChoice = c}
        onapiKeyChange={(k) => apiKey = k}
        onskip={handlePowerSkip}
        onback={goBack}
      />
    </div>
  {:else if current === 'hotkey'}
    <div class="wizard__screen">
      <Hotkey
        hotkey={hotkey}
        voiceWakeEnabled={voiceWakeEnabled}
        onrecord={(h) => hotkey = h}
        onvoicetoggle={(v) => voiceWakeEnabled = v}
        oncontinue={handleHotkeyContinue}
        onback={goBack}
      />
    </div>
  {:else if current === 'breath'}
    <FirstBreath oncomplete={handleBreathComplete} />
  {/if}
</div>

<style>
  .wizard {
    position: relative;
    width: 100%;
    height: 100vh;
    overflow: hidden;
  }

  .wizard__screen {
    position: absolute;
    inset: 0;
    animation: screen-in var(--duration-base) var(--ease-decelerate) both;
  }

  /* Forward transition */
  [data-direction="forward"] .wizard__screen {
    animation: screen-in-forward var(--duration-base) var(--ease-decelerate) both;
  }

  [data-direction="back"] .wizard__screen {
    animation: screen-in-back var(--duration-base) var(--ease-decelerate) both;
  }

  @keyframes screen-in-forward {
    from {
      opacity: 0;
      transform: translateX(20px);
    }
    to {
      opacity: 1;
      transform: translateX(0);
    }
  }

  @keyframes screen-in-back {
    from {
      opacity: 0;
      transform: translateX(-20px);
    }
    to {
      opacity: 1;
      transform: translateX(0);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .wizard__screen {
      animation: none;
    }
  }
</style>