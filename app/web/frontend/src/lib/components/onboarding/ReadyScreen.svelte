<script lang="ts">
  import { onMount } from 'svelte'
  import { ipc } from '../../ipc/client'
  import { onboarding } from '../../stores/onboarding.svelte'
  import type { PowerProbeResult } from '../../ipc/types'
  import Button from '../ui/Button.svelte'
  import Card from '../ui/Card.svelte'
  import Badge from '../ui/Badge.svelte'
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
    void ipc
      .onboardingProbePower()
      .then((p) => {
        probe = p
      })
      .catch(() => {
        probe = { ollama_reachable: false, ollama_models: [], clis: [], recommended: 'none' }
      })
      .finally(() => {
        probing = false
      })

    void ipc
      .call<VoiceProbe>('onboarding.probe_voice', {})
      .then((v) => {
        voice = v
      })
      .catch(() => {
        voice = null
      })
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
    if (w.runtime?.BrowserOpenURL) {
      w.runtime.BrowserOpenURL(url)
    } else {
      window.open(url, '_blank')
    }
  }

  // Short label for a detected CLI, e.g. "Claude Code" instead of
  // the raw binary name. Falls back to the raw string if we don't
  // have a friendly mapping.
  const CLI_LABELS: Record<string, string> = {
    'claude': 'Claude Code',
    'codex': 'Codex',
    'agy': 'Antigravity',
    'opencode': 'OpenCode',
    'kilo': 'Kilo',
    'hermes': 'Hermes',
    'gemini': 'Gemini CLI'
  }
  function labelFor(cli: import('../../ipc/types').PowerProbeCLI): string {
    const name = cli.name ?? 'unknown'
    const base = name.split(/[\\/]/).pop() ?? name
    return CLI_LABELS[base] ?? base
  }
</script>

<div class="wizard ready">
  <header class="hero">
    <h1>{t('onboarding.ready.title')}</h1>
    <p class="subtitle">{t('onboarding.ready.subtitle')}</p>
  </header>

  {#if probing}
    <p class="muted">{t('onboarding.ready.probing')}</p>
  {/if}

  {#if probe}
    <section class="power">
      <h3 class="section-title">{t('onboarding.ready.power_title')}</h3>
      <div class="power-grid">
        <Card elevation="glass" padding="md" class="power-card">
          <div class="power-row">
            <div class="power-icon" class:on={probe.ollama_reachable} aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="3" />
                <path d="M12 1v6" />
                <path d="M12 17v6" />
                <path d="M4.22 4.22l4.24 4.24" />
                <path d="M15.54 15.54l4.24 4.24" />
                <path d="M1 12h6" />
                <path d="M17 12h6" />
                <path d="M4.22 19.78l4.24-4.24" />
                <path d="M15.54 8.46l4.24-4.24" />
              </svg>
            </div>
            <div class="power-body">
              <div class="power-head">
                <span class="power-name">Ollama</span>
                <Badge tone={probe.ollama_reachable ? 'success' : 'neutral'} dot pulse={probe.ollama_reachable}>
                  {probe.ollama_reachable
                    ? t('onboarding.ready.ollama_reachable')
                    : t('onboarding.ready.ollama_offline')}
                </Badge>
              </div>
              <p class="power-desc">
                {#if probe.ollama_reachable && probe.ollama_models.length}
                  {t('onboarding.ready.ollama_models', probe.ollama_models.slice(0, 3).join(', '))}
                {:else}
                  {t('onboarding.ready.ollama_desc')}
                {/if}
              </p>
            </div>
          </div>
        </Card>

        {#if probe.clis && probe.clis.length}
          <Card elevation="glass" padding="md" class="power-card">
            <div class="power-row">
              <div class="power-icon" aria-hidden="true">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="4 17 10 11 4 5" />
                  <line x1="12" y1="19" x2="20" y2="19" />
                </svg>
              </div>
              <div class="power-body">
                <div class="power-head">
                  <span class="power-name">{t('onboarding.ready.clis_title')}</span>
                  <Badge tone="accent" size="sm">
                    {probe.clis.length}
                  </Badge>
                </div>
                <div class="cli-chips">
                  {#each probe.clis as cli, i (cli.name + i)}
                    <span class="cli-chip">{labelFor(cli)}</span>
                  {/each}
                </div>
              </div>
            </div>
          </Card>
        {/if}
      </div>
    </section>
  {/if}

  <section class="optional">
    <h3 class="section-title">{t('onboarding.ready.optional_title')}</h3>
    <div class="optional-grid">
      <Card
        elevation="glass"
        padding="md"
        interactive
        onclick={() => finish('#/settings')}
      >
        <span class="opt-eyebrow">{t('onboarding.ready.opt_api_eyebrow')}</span>
        <span class="opt-title">{t('onboarding.ready.add_provider')}</span>
        <span class="opt-desc">{t('onboarding.ready.add_provider_desc')}</span>
      </Card>

      <Card
        elevation="glass"
        padding="md"
        interactive
        onclick={() => finish('#/channels')}
      >
        <span class="opt-eyebrow">{t('onboarding.ready.opt_msg_eyebrow')}</span>
        <span class="opt-title">{t('onboarding.ready.connect_messaging')}</span>
        <span class="opt-desc">{t('onboarding.ready.telegram_settings')}</span>
      </Card>

      <Card
        elevation="glass"
        padding="md"
        interactive
        onclick={() => finish('#/settings')}
      >
        <span class="opt-eyebrow">{t('onboarding.ready.opt_account_eyebrow')}</span>
        <span class="opt-title">{t('onboarding.ready.opt_account_title')}</span>
        <span class="opt-desc">{t('onboarding.ready.opt_account_desc')}</span>
      </Card>
    </div>
  </section>

  {#if hotkey}
    <p class="hotkey-note">{t('onboarding.ready.hotkey_note', hotkey)}</p>
  {/if}

  <div class="actions center">
    <Button
      variant="primary"
      size="lg"
      onclick={() => finish()}
      disabled={onboarding.busy}
      loading={onboarding.busy}
    >
      {onboarding.busy ? t('onboarding.ready.starting') : t('onboarding.ready.start_button')}
    </Button>
  </div>
</div>

<style>
  .wizard {
    width: 100%;
    max-width: 720px;
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: var(--space-5);
  }

  .hero {
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-2);
  }

  h1 {
    font-family: var(--font-display);
    font-size: var(--size-4xl);
    font-weight: var(--weight-light);
    letter-spacing: var(--tracking-tighter);
    line-height: var(--leading-tight);
    margin: 0;
    background: var(--accent-gradient);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .subtitle {
    color: var(--text-muted);
    font-size: var(--size-md);
    line-height: var(--leading-relaxed);
    margin: 0;
    max-width: 52ch;
  }

  .section-title {
    font-size: var(--size-xs);
    font-family: var(--font-mono);
    text-transform: uppercase;
    letter-spacing: var(--tracking-widest);
    color: var(--text-faint);
    font-weight: var(--weight-semibold);
    margin: 0 0 var(--space-3) 0;
  }

  .power-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: var(--space-3);
  }
  @media (max-width: 640px) {
    .power-grid { grid-template-columns: 1fr; }
  }

  .power-row {
    display: flex;
    gap: var(--space-3);
    align-items: flex-start;
  }
  .power-icon {
    width: 36px;
    height: 36px;
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--surface-2);
    color: var(--text-muted);
    flex-shrink: 0;
  }
  .power-icon.on {
    background: var(--accent-soft);
    color: var(--accent);
  }
  .power-icon svg { width: 20px; height: 20px; }

  .power-body { flex: 1; min-width: 0; }
  .power-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
    margin-bottom: var(--space-1);
  }
  .power-name {
    font-weight: var(--weight-semibold);
    font-size: var(--size-md);
  }
  .power-desc {
    color: var(--text-muted);
    font-size: var(--size-sm);
    line-height: var(--leading-relaxed);
    margin: 0;
  }

  .cli-chips {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .cli-chip {
    display: inline-flex;
    align-items: center;
    padding: 4px 10px;
    border-radius: var(--radius-pill);
    background: var(--surface-3);
    border: 1px solid var(--border);
    color: var(--text);
    font-size: var(--size-xs);
    font-family: var(--font-mono);
  }

  .optional-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: var(--space-3);
  }
  @media (max-width: 640px) {
    .optional-grid { grid-template-columns: 1fr; }
  }

  .opt-eyebrow {
    font-size: var(--size-xs);
    font-family: var(--font-mono);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
    color: var(--text-faint);
    margin-bottom: var(--space-1);
  }
  .opt-title {
    display: block;
    font-weight: var(--weight-semibold);
    font-size: var(--size-md);
    color: var(--text);
    margin-bottom: var(--space-1);
  }
  .opt-desc {
    display: block;
    color: var(--text-muted);
    font-size: var(--size-sm);
    line-height: var(--leading-relaxed);
  }

  .hotkey-note {
    text-align: center;
    color: var(--text-muted);
    font-size: var(--size-sm);
    margin: 0;
  }

  .actions.center {
    display: flex;
    justify-content: center;
    margin-top: var(--space-3);
  }

  .muted {
    color: var(--text-muted);
    font-size: var(--size-md);
    text-align: center;
  }
</style>
