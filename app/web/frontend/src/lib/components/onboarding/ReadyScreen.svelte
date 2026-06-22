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
  <div class="check">✓</div>
  <h1>{t('onboarding.ready.title')}</h1>

  {#if probing}
    <p class="muted">{t('onboarding.ready.probing')}</p>
  {/if}

  <div class="setup-cards">
    <button class="setup-card primary" onclick={() => finish('#/settings')} disabled={onboarding.busy}>
      <span class="card-label">{t('onboarding.ready.add_provider')}</span>
      <span class="card-desc">{t('onboarding.ready.add_provider_desc')}</span>
    </button>

    {#if probe?.ollama_reachable}
      <button class="setup-card" onclick={() => finish('#/settings')} disabled={onboarding.busy}>
        <span class="card-label">{t('onboarding.ready.ollama_detected')}</span>
        <span class="card-desc">
          {t('onboarding.ready.ollama_desc')}{#if probe.ollama_models.length} ({probe.ollama_models.slice(0, 2).join(', ')}){/if}
        </span>
      </button>
    {/if}

    <button class="setup-card" onclick={() => finish('#/settings')} disabled={onboarding.busy}>
      <span class="card-label">{t('onboarding.ready.connect_messaging')}</span>
      <span class="card-desc">{t('onboarding.ready.telegram_settings')}</span>
    </button>

    <button class="setup-card" onclick={() => finish('#/settings')} disabled={onboarding.busy}>
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
    <button class="btn btn-primary big" onclick={() => finish()} disabled={onboarding.busy}>
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
  }
  .check {
    font-size: 56px;
    color: var(--color-accent);
    margin-bottom: var(--space-2);
  }
  h1 {
    font-size: var(--size-3xl);
    font-weight: 600;
    margin-bottom: var(--space-4);
    background: var(--color-accent-gradient);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .muted { color: var(--color-text-muted); }
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
    border-radius: var(--radius-lg);
    border: 1px solid var(--glass-border);
    background: var(--glass-bg);
    color: var(--color-text);
    cursor: pointer;
    transition: all var(--transition-base);
  }
  .setup-card:hover:not(:disabled) {
    border-color: var(--color-accent);
  }
  .setup-card.primary {
    border-color: var(--color-accent);
    background: rgba(99, 102, 241, 0.06);
  }
  .setup-card:disabled { opacity: 0.5; cursor: not-allowed; }
  .card-label { font-weight: 600; font-size: var(--size-md); }
  .card-desc { color: var(--color-text-muted); font-size: var(--size-sm); }
  .hotkey-note {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
    margin-bottom: var(--space-4);
  }
  kbd {
    font-family: var(--font-mono);
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: 6px;
    padding: 2px 8px;
  }
  .actions.center { display: flex; justify-content: center; }
  .btn {
    padding: 12px 24px;
    border-radius: var(--radius-pill);
    font-size: var(--size-md);
    font-weight: 500;
    cursor: pointer;
    border: none;
    transition: all var(--transition-spring);
  }
  .btn.big { padding: 14px 36px; font-size: var(--size-lg); }
  .btn-primary {
    background: var(--color-accent-gradient);
    color: white;
  }
  .btn-primary:hover:not(:disabled) {
    box-shadow: var(--shadow-glow);
    transform: translateY(-1px);
  }
  .btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
