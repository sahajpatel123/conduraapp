<script lang="ts">
  import { onMount } from 'svelte'
  import { ipc } from '../../ipc/client'
  import { onboarding } from '../../stores/onboarding.svelte'
  import type { PowerProbeResult } from '../../ipc/types'
  import { t } from '../../i18n'

  interface Props {
    onDone?: (route?: string) => void
  }
  let { onDone }: Props = $props()

  interface VoiceProbe {
    mic_available: boolean
    voice_enabled: boolean
    wake_word_enabled: boolean
    wake_word: string
    ready: boolean
  }

  let probe = $state<PowerProbeResult | null>(null)
  let probing = $state(true)
  let voice = $state<VoiceProbe | null>(null)

  onMount(() => {
    void ipc.onboardingProbePower().then((p) => { probe = p }).catch(() => {
      probe = { ollama_reachable: false, ollama_models: [], clis: [], recommended: 'none' }
    }).finally(() => { probing = false })

    void ipc.call<VoiceProbe>('onboarding.probe_voice', {}).then((v) => { voice = v }).catch(() => { voice = null })
  })

  const hotkey = $derived(onboarding.daemon?.steps?.hotkey?.data ?? '')

  async function finish(route?: string): Promise<void> {
    const res = await onboarding.finish({
      hotkey,
      eula_version: onboarding.eulaVersion || onboarding.daemon?.steps?.eula?.data || 'v1',
      permissions_skipped: onboarding.daemon?.steps?.permissions?.status === 'skipped'
    })
    if (res) onDone?.(route)
  }

  function openExternal(url: string): void {
    const w = window as unknown as { runtime?: { BrowserOpenURL?: (u: string) => void } }
    if (w.runtime?.BrowserOpenURL) { w.runtime.BrowserOpenURL(url) } else { window.open(url, '_blank') }
  }
</script>

<div class="wizard ready">
  <div class="check" aria-hidden="true">
    <svg viewBox="0 0 52 52" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
      <circle class="check-ring" cx="26" cy="26" r="22" />
      <path class="check-mark" d="M16 27l7 7 13-14" />
    </svg>
  </div>
  <h1>{t('onboarding.ready.title')}</h1>

  {#if probing}
    <p class="muted">{t('onboarding.ready.probing')}</p>
  {/if}

  <div class="setup-cards">
    <button class="setup-card glass-card primary" onclick={() => finish('#/settings')} disabled={onboarding.busy}>
      <span class="card-label">{t('onboarding.ready.add_provider')}</span>
      <span class="card-desc">{t('onboarding.ready.add_provider_desc')}</span>
    </button>

    {#if probe?.ollama_reachable}
      <button class="setup-card glass-card" onclick={() => finish('#/settings')} disabled={onboarding.busy}>
        <span class="card-label">{t('onboarding.ready.ollama_detected')}</span>
        <span class="card-desc">
          {t('onboarding.ready.ollama_desc')}{#if probe.ollama_models.length} ({probe.ollama_models.slice(0, 2).join(', ')}){/if}
        </span>
      </button>
    {/if}

    <button class="setup-card glass-card" onclick={() => finish('#/settings')} disabled={onboarding.busy}>
      <span class="card-label">{t('onboarding.ready.connect_messaging')}</span>
      <span class="card-desc">{t('onboarding.ready.telegram_settings')}</span>
    </button>

    <button class="setup-card glass-card" onclick={() => finish('#/settings')} disabled={onboarding.busy}>
      <span class="card-label">{t('onboarding.ready.setup_voice')}</span>
      <span class="card-desc">
        {#if voice}
          {t('onboarding.ready.voice_status', voice.mic_available ? t('onboarding.ready.mic_ready') : t('onboarding.ready.mic_unavailable'), voice.wake_word_enabled ? t('onboarding.ready.wake_on') : t('onboarding.ready.wake_off'))}
        {:else}
          {t('onboarding.ready.talk_hands_free')}
        {/if}
      </span>
    </button>
  </div>

  {#if hotkey}
    <p class="hotkey-note">{t('onboarding.ready.hotkey_note', hotkey)}</p>
  {/if}

  <div class="actions center">
    <button class="btn btn-primary btn-lg cta" onclick={() => finish()} disabled={onboarding.busy}>
      {onboarding.busy ? t('onboarding.ready.starting') : t('onboarding.ready.start_button')}
    </button>
  </div>
</div>

<style>
  .wizard {
    width: 100%;
    max-width: 560px;
    padding: var(--space-6) var(--space-5);
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    animation: screen-in var(--transition-spring-soft) var(--ease-out-expo) both;
  }
  @keyframes screen-in {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: none; }
  }
  .check {
    width: 64px;
    height: 64px;
    margin-bottom: var(--space-3);
    color: var(--color-accent);
    filter: drop-shadow(0 0 16px var(--color-glow-strong));
  }
  .check svg { width: 100%; height: 100%; }
  .check-ring {
    stroke-dasharray: 140;
    stroke-dashoffset: 140;
    animation: draw 620ms var(--ease-out-expo) 80ms forwards;
  }
  .check-mark {
    stroke-dasharray: 40;
    stroke-dashoffset: 40;
    animation: draw 360ms var(--ease-out-expo) 540ms forwards;
  }
  @keyframes draw {
    to { stroke-dashoffset: 0; }
  }
  h1 {
    font-size: var(--size-3xl);
    font-weight: var(--weight-semibold);
    letter-spacing: var(--tracking-tight);
    margin-bottom: var(--space-4);
    background: var(--color-accent-gradient);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .wizard > .muted { color: var(--color-text-muted); }
  .setup-cards {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    width: 100%;
    margin-bottom: var(--space-4);
    text-align: left;
  }
  .setup-card {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: var(--space-3) var(--space-4);
    color: var(--color-text);
    cursor: pointer;
    transition: transform var(--transition-base), border-color var(--transition-base), box-shadow var(--transition-base);
  }
  .setup-card:hover:not(:disabled) {
    border-color: var(--color-border-accent);
    transform: translateY(-1px);
    box-shadow: var(--shadow-md);
  }
  .setup-card.primary {
    border-color: var(--color-border-accent);
    background: var(--color-accent-gradient-subtle), var(--glass-bg);
  }
  .setup-card:disabled { opacity: 0.5; cursor: not-allowed; }
  .card-label { font-weight: var(--weight-semibold); font-size: var(--size-md); }
  .card-desc { color: var(--color-text-muted); font-size: var(--size-sm); line-height: var(--leading-relaxed); }
  .hotkey-note {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
    margin-bottom: var(--space-4);
  }
  .actions.center { display: flex; justify-content: center; }
  .cta {
    border-radius: var(--radius-pill);
    box-shadow: var(--shadow-md);
  }
  .cta:hover:not(:disabled) {
    box-shadow: var(--shadow-glow-strong);
  }
</style>
