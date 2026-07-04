<script lang="ts">
  import { onMount } from 'svelte';
  import { fly, fade } from 'svelte/transition';
  import { onboarding } from '../stores/onboarding.svelte';
  import { ipc } from '../ipc/client';
  import Button from './Button.svelte';
  import Glyph from './Glyph.svelte';
  import Pulse from './Pulse.svelte';

  // The floating interview — first-run setup as a conversation over the
  // living shell, one question at a time. Renders the real daemon onboarding
  // state machine (eula → permissions → hotkey → complete); the power-source
  // choice lives in the Ready screen. Nothing here invents a parallel flow.

  let { onComplete }: { onComplete?: (route?: string) => void } = $props();

  type StepId = 'eula' | 'permissions' | 'hotkey' | 'complete';
  const STEPS: StepId[] = ['eula', 'permissions', 'hotkey', 'complete'];
  const STEP_LABEL: Record<StepId, string> = {
    eula: 'The rules',
    permissions: 'Permissions',
    hotkey: 'The key',
    complete: 'Ready',
  };

  let displayStep = $state<StepId>('eula');
  let previousStep = $state<StepId>('eula');

  $effect(() => {
    const next = (onboarding.currentStep as StepId) ?? 'eula';
    if (next === displayStep) return;
    previousStep = displayStep;
    requestAnimationFrame(() => (displayStep = next));
  });

  let stepIndex = $derived(STEPS.indexOf(displayStep));

  // ── EULA ──
  let eulaScrolled = $state(false);
  let eulaAccepted = $state(false);
  let eulaEl = $state<HTMLDivElement | undefined>(undefined);
  function checkEulaScroll() {
    if (!eulaEl) return;
    if (eulaEl.scrollHeight - (eulaEl.scrollTop + eulaEl.clientHeight) <= 8) eulaScrolled = true;
  }
  let canAcceptEula = $derived(eulaScrolled && eulaAccepted && !onboarding.busy);
  async function acceptEula() {
    try {
      await onboarding.acceptEula(onboarding.eula?.version ?? 'v1');
    } catch (e) {
      console.error('acceptEula failed', e);
    }
  }

  // ── Permissions ──
  type PermRow = { kind: string; status: 'granted' | 'denied' | 'unknown'; note?: string };
  let permRows = $state<PermRow[]>([]);
  const REQUIRED_PERMS = ['accessibility', 'screen_recording'];
  let permPoll = 0;
  async function refreshPerms() {
    try {
      permRows = (await ipc.permissionsStatus()) as PermRow[];
    } catch {
      permRows = REQUIRED_PERMS.map((k) => ({ kind: k, status: 'unknown' as const }));
    }
  }
  async function openPermSettings(kind: string) {
    try {
      const g = await ipc.permissionsGuide(kind);
      const url = (g as { deep_link?: string; help_url?: string }).deep_link ?? (g as { help_url?: string }).help_url;
      if (url) window.open(url, '_blank');
    } catch (e) {
      console.error('permissionsGuide failed', e);
    }
  }
  let atLeastOneGranted = $derived(permRows.some((r) => r.status === 'granted'));
  async function completePermissions() {
    try {
      await onboarding.completePermissions();
    } catch (e) {
      console.error(e);
    }
  }
  async function skipPermissions() {
    try {
      await onboarding.skipStep('permissions');
    } catch (e) {
      console.error(e);
    }
  }
  async function back() {
    try {
      await onboarding.back();
    } catch (e) {
      console.error(e);
    }
  }

  // ── Hotkey ──
  let combo = $state(onboarding.daemon?.steps?.hotkey?.data ?? '');
  const PRESETS = ['Option+Option', 'Cmd+Shift+Space', 'Ctrl+Space'];
  let recording = $state(false);
  function startRecording() {
    recording = true;
  }
  function onRecordKey(e: KeyboardEvent) {
    if (!recording) return;
    e.preventDefault();
    e.stopPropagation();
    if (e.key === 'Escape') {
      recording = false;
      return;
    }
    const parts: string[] = [];
    if (e.metaKey) parts.push('Cmd');
    if (e.ctrlKey) parts.push('Ctrl');
    if (e.altKey) parts.push('Option');
    if (e.shiftKey) parts.push('Shift');
    let k = e.key;
    if (k === ' ') k = 'Space';
    if (k.length === 1) k = k.toUpperCase();
    if (!['Meta', 'Control', 'Alt', 'Shift'].includes(e.key)) parts.push(k);
    if (parts.length > 0) {
      combo = parts.join('+');
      recording = false;
    }
  }
  let canSaveHotkey = $derived(!!combo && !onboarding.busy);
  async function saveHotkey() {
    onboarding.setHotkey(combo);
    try {
      await onboarding.saveHotkey();
    } catch (e) {
      console.error(e);
    }
  }
  async function skipHotkey() {
    onboarding.setHotkey('');
    try {
      await onboarding.skipStep('hotkey');
    } catch (e) {
      console.error(e);
    }
  }

  // ── Ready ──
  let probe = $state<{ ollama_reachable: boolean; ollama_models: string[]; clis: { name: string; found: boolean }[]; recommended: string } | null>(null);
  let voice = $state<{ mic_available: boolean; ready: boolean; wake_word: string } | null>(null);
  let probing = $state(true);
  let powerChoice = $state<'local' | 'apikey' | 'sub'>('local');
  async function finish(route?: string) {
    try {
      const res = await onboarding.finish({
        hotkey: onboarding.daemon?.steps?.hotkey?.data ?? onboarding.hotkeyValue,
        eula_version: onboarding.eulaVersion ?? onboarding.daemon?.steps?.eula?.data ?? 'v1',
        permissions_skipped: onboarding.daemon?.steps?.permissions?.status === 'skipped',
      });
      if (res) {
        const dest = route ?? (powerChoice === 'local' ? undefined : '#/settings');
        onComplete?.(dest);
      }
    } catch (e) {
      console.error('finish failed', e);
    }
  }

  onMount(() => {
    try {
      void onboarding.sync();
      void onboarding.loadEula();
    } catch (e) {
      console.warn('onboarding sync failed', e);
    }
    void refreshPerms();
    permPoll = window.setInterval(refreshPerms, 2000);
    void ipc.onboardingProbePower().then((p) => (probe = p as typeof probe)).catch(() => {
      probe = { ollama_reachable: false, ollama_models: [], clis: [], recommended: 'none' };
    }).finally(() => (probing = false));
    void ipc.onboardingProbeVoice().then((v) => (voice = v as typeof voice)).catch(() => (voice = null));

    window.addEventListener('keydown', onRecordKey, true);
    return () => {
      window.clearInterval(permPoll);
      window.removeEventListener('keydown', onRecordKey, true);
    };
  });

  let progressPct = $derived((stepIndex / (STEPS.length - 1)) * 100);
</script>

<div class="scrim" transition:fade={{ duration: 320 }}>
  <div class="float-card" transition:fly={{ y: 16, duration: 520, delay: 80 }}>
    <svg class="fc-thread" preserveAspectRatio="none" aria-hidden="true">
      <path d="M 0 12 L 9999 12" stroke="var(--synapse)" stroke-width="1.25" fill="none" stroke-linecap="round" pathLength="1" vector-effect="non-scaling-stroke" stroke-dasharray="1" stroke-dashoffset="0" />
      <circle cx="0" cy="12" r="3" fill="var(--pollen)" />
    </svg>

    <div class="wordmark">Condura<span class="dot"></span></div>
    <div class="eyebrow">Step {stepIndex + 1} of {STEPS.length} · {STEP_LABEL[displayStep]}</div>

    {#key displayStep}
      <div class="step-body" in:fly={{ y: 12, duration: 520, delay: 120 }}>
        {#if displayStep === 'eula'}
          <h2 class="fc-title">Before anything else, the rules.</h2>
          <p class="fc-sub">The Condura Freeware EULA. Scroll to the bottom, then accept.</p>
          <div class="eula" bind:this={eulaEl} onscroll={checkEulaScroll}>
            {onboarding.eula?.text ?? 'Loading the license…'}
          </div>
          <label class="check"><input type="checkbox" bind:checked={eulaAccepted} /> I accept the Condura Freeware EULA</label>
          <div class="fc-foot">
            <Button variant="primary" magnetic disabled={!canAcceptEula} onclick={acceptEula}>Accept & continue →</Button>
          </div>
        {:else if displayStep === 'permissions'}
          <h2 class="fc-title">Two things, to see and touch apps.</h2>
          <p class="fc-sub">Accessibility and Screen Recording. Microphone and others come later, when you need them.</p>
          <div class="perms">
            {#each REQUIRED_PERMS as kind (kind)}
              {@const row = permRows.find((r) => r.kind === kind)}
              <div class="perm">
                <div class="perm-body">
                  <div class="perm-name">{kind === 'accessibility' ? 'Accessibility' : 'Screen Recording'}</div>
                  <div class="perm-status" data-status={row?.status ?? 'unknown'}>{row?.status ?? 'unknown'}</div>
                </div>
                <button class="perm-open" onclick={() => openPermSettings(kind)}>Open System Settings →</button>
              </div>
            {/each}
          </div>
          <div class="fc-foot">
            <button class="link-back" onclick={back}>← Back</button>
            <button class="link-skip" onclick={skipPermissions}>Skip for now</button>
            <Button variant="primary" magnetic onclick={completePermissions}>Continue →</Button>
          </div>
        {:else if displayStep === 'hotkey'}
          <h2 class="fc-title">Choose the key that summons Condura.</h2>
          <p class="fc-sub">No silent default — you pick. Try a suggestion, or record your own.</p>
          <div class="presets">
            {#each PRESETS as p (p)}
              <button class="preset" class:active={combo === p} onclick={() => (combo = p)}>{p}</button>
            {/each}
          </div>
          <button class="recorder" class:recording onclick={startRecording}>
            {recording ? 'Press a key combo…' : combo ? `Recorded: ${combo}` : 'Record your own →'}
          </button>
          <div class="fc-foot">
            <button class="link-back" onclick={back}>← Back</button>
            <button class="link-skip" onclick={skipHotkey}>Skip</button>
            <Button variant="primary" magnetic disabled={!canSaveHotkey} onclick={saveHotkey}>Continue →</Button>
          </div>
        {:else}
          <h2 class="fc-title">Your computer, <span class="alive">alive.</span></h2>
          <p class="fc-sub">Condura is ready. How should it think? Pick one now — change it any time in Settings.</p>
          {#if probing}
            <div class="probe-loading"><Pulse phase="thinking" size={8} /> <span>Detecting what's on your machine…</span></div>
          {/if}
          <div class="choices">
            <button class="choice" class:selected={powerChoice === 'local'} onclick={() => (powerChoice = 'local')}>
              <span class="c-radio"></span>
              <span class="c-body"><span class="c-title">Local model · Ollama</span>
                <span class="c-meta">{probe?.ollama_reachable ? `${probe.ollama_models.length} models detected · no data leaves your machine` : 'install Ollama to use local models'}</span></span>
              {#if probe?.ollama_reachable}<span class="c-tag">Recommended</span>{/if}
            </button>
            <button class="choice" class:selected={powerChoice === 'apikey'} onclick={() => (powerChoice = 'apikey')}>
              <span class="c-radio"></span>
              <span class="c-body"><span class="c-title">Paste an API key</span><span class="c-meta">Anthropic · OpenAI · Google · xAI · more</span></span>
            </button>
            <button class="choice" class:selected={powerChoice === 'sub'} onclick={() => (powerChoice = 'sub')}>
              <span class="c-radio"></span>
              <span class="c-body"><span class="c-title">Connect a subscription</span><span class="c-meta">Claude Pro · ChatGPT Plus · Gemini · SuperGrok</span></span>
            </button>
          </div>
          <div class="fc-foot">
            <button class="link-back" onclick={back}>← Back</button>
            <Button variant="primary" magnetic onclick={() => finish()}>
              {powerChoice === 'local' ? 'Enter Condura →' : 'Enter Condura · add key →'}
            </Button>
          </div>
        {/if}
      </div>
    {/key}

    <div class="fc-prog">
      <svg viewBox="0 0 100 8" preserveAspectRatio="none">
        <line x1="0" y1="4" x2="100" y2="4" stroke="var(--hair-strong)" stroke-width="1.5" stroke-linecap="round" />
        <line x1="0" y1="4" x2={progressPct} y2="4" stroke="var(--synapse)" stroke-width="1.5" stroke-linecap="round" style="transition: x2 var(--dur-slow) var(--ease)" />
      </svg>
      <span class="step-count">{stepIndex + 1} / {STEPS.length}</span>
    </div>
  </div>
</div>

<style>
  .scrim {
    position: fixed;
    inset: 0;
    z-index: var(--z-modal);
    background: color-mix(in oklab, var(--ink) 40%, transparent);
    backdrop-filter: blur(6px);
    display: grid;
    place-items: center;
    overflow: auto;
    padding: var(--space-6);
  }
  .float-card {
    position: relative;
    width: min(560px, 92vw);
    background: var(--surface);
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-xl);
    box-shadow: var(--shadow-float);
    padding: var(--space-8) var(--space-8) var(--space-6);
  }
  .fc-thread {
    position: absolute;
    top: 0;
    left: var(--space-8);
    right: var(--space-8);
    height: 24px;
    width: auto;
    overflow: visible;
  }
  .wordmark {
    font-family: var(--font-display);
    font-size: 22px;
    letter-spacing: -0.03em;
    display: flex;
    align-items: baseline;
    gap: 3px;
    margin-bottom: var(--space-3);
  }
  .wordmark .dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--pollen);
    transform: translateY(-1px);
    box-shadow: 0 0 8px color-mix(in oklab, var(--pollen) 60%, transparent);
  }
  .fc-title {
    font-family: var(--font-display);
    font-size: clamp(26px, 3vw, 34px);
    line-height: 1.08;
    letter-spacing: -0.03em;
    margin: var(--space-3) 0;
  }
  .fc-title .alive {
    font-style: italic;
    color: var(--synapse);
  }
  .fc-sub {
    color: var(--content-mute);
    font-size: 14px;
    margin-bottom: var(--space-6);
  }
  .step-body {
    min-height: 220px;
  }

  .eula {
    max-height: 180px;
    overflow-y: auto;
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface-card);
    padding: var(--space-4);
    font-size: 13px;
    line-height: 1.6;
    color: var(--content-soft);
    white-space: pre-wrap;
    margin-bottom: var(--space-4);
  }
  .check {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    color: var(--content);
    margin-bottom: var(--space-5);
    cursor: pointer;
  }
  .check input {
    width: 16px;
    height: 16px;
    accent-color: var(--pollen);
  }

  .perms {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin-bottom: var(--space-5);
  }
  .perm {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface-card);
  }
  .perm-name {
    font-size: 15px;
    color: var(--content);
  }
  .perm-status {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    margin-top: 2px;
  }
  .perm-status[data-status='granted'] {
    color: var(--ok);
  }
  .perm-status[data-status='denied'] {
    color: var(--danger);
  }
  .perm-status[data-status='unknown'] {
    color: var(--content-faint);
  }
  .perm-open {
    margin-left: auto;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.08em;
    color: var(--synapse);
    text-transform: uppercase;
  }
  .perm-open:hover {
    text-decoration: underline;
  }

  .presets {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
    margin-bottom: var(--space-3);
  }
  .preset {
    font-family: var(--font-mono);
    font-size: 12px;
    padding: 8px 12px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-sm);
    background: var(--surface-card);
    color: var(--content-soft);
    transition: border-color var(--dur) var(--ease), color var(--dur) var(--ease);
  }
  .preset.active {
    border-color: var(--pollen);
    color: var(--content);
    box-shadow: var(--pollen-halo);
  }
  .recorder {
    width: 100%;
    text-align: center;
    padding: var(--space-4);
    border: 1px dashed var(--hair-strong);
    border-radius: var(--r-md);
    background: transparent;
    color: var(--content-mute);
    font-size: 14px;
    margin-bottom: var(--space-5);
    transition: border-color var(--dur) var(--ease), color var(--dur) var(--ease);
  }
  .recorder.recording {
    border-color: var(--pollen);
    color: var(--pollen);
  }

  .probe-loading {
    display: flex;
    align-items: center;
    gap: 8px;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
    margin-bottom: var(--space-4);
  }
  .choices {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin-bottom: var(--space-5);
  }
  .choice {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-4) var(--space-5);
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface-card);
    text-align: left;
    transition: border-color var(--dur) var(--ease), box-shadow var(--dur) var(--ease), transform var(--dur) var(--ease);
  }
  .choice:hover {
    transform: translateY(-1px);
    border-color: var(--hair-strong);
  }
  .choice.selected {
    border-color: var(--pollen);
    box-shadow: var(--pollen-halo);
  }
  .c-radio {
    width: 18px;
    height: 18px;
    border-radius: 50%;
    border: 2px solid var(--hair-strong);
    flex: none;
    display: grid;
    place-items: center;
    transition: border-color var(--dur) var(--ease);
  }
  .choice.selected .c-radio {
    border-color: var(--pollen);
  }
  .choice.selected .c-radio::after {
    content: '';
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--pollen);
  }
  .c-body {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .c-title {
    font-size: 15px;
    color: var(--content);
  }
  .c-meta {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .c-tag {
    margin-left: auto;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--synapse);
    border: 1px solid color-mix(in oklab, var(--synapse) 30%, transparent);
    padding: 3px 8px;
    border-radius: var(--r-pill);
    background: color-mix(in oklab, var(--synapse) 6%, transparent);
  }

  .fc-foot {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    margin-top: var(--space-2);
  }
  .link-back,
  .link-skip {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .link-back:hover,
  .link-skip:hover {
    color: var(--content);
  }
  .fc-foot :global(.btn-primary) {
    margin-left: auto;
  }

  .fc-prog {
    margin-top: var(--space-7);
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }
  .fc-prog svg {
    flex: 1;
    height: 8px;
    overflow: visible;
  }
  .step-count {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
</style>