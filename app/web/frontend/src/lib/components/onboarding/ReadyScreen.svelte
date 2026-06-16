<script lang="ts">
  import { onMount } from 'svelte'
  import { ipc } from '../../ipc/client'
  import { onboarding } from '../../stores/onboarding.svelte'
  import type { PowerProbeResult } from '../../ipc/types'

  interface Props {
    // Called after the wizard finishes. `route` is an optional
    // hash the parent should navigate to (e.g. '#/settings').
    onDone?: (route?: string) => void
  }
  let { onDone }: Props = $props()

  let probe = $state<PowerProbeResult | null>(null)
  let probing = $state(true)

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
  })

  const hotkey = $derived(onboarding.daemon?.steps?.hotkey?.data ?? '')
  const foundClis = $derived((probe?.clis ?? []).filter((c) => c.found).map((c) => c.name))

  function openExternal(url: string): void {
    const w = window as unknown as { runtime?: { BrowserOpenURL?: (u: string) => void } }
    if (w.runtime?.BrowserOpenURL) {
      w.runtime.BrowserOpenURL(url)
    } else {
      window.open(url, '_blank')
    }
  }

  async function finish(route?: string): Promise<void> {
    const res = await onboarding.finish({
      hotkey,
      eula_version: onboarding.eulaVersion || onboarding.daemon?.steps?.eula?.data || 'v1',
      permissions_skipped: onboarding.daemon?.steps?.permissions?.status === 'skipped'
    })
    if (res) {
      onDone?.(route)
    }
  }
</script>

<div class="wizard ready">
  <div class="check">✓</div>
  <h1>You're all set</h1>

  {#if probing}
    <p class="muted">Checking what's available on your machine…</p>
  {:else if probe?.ollama_reachable}
    <div class="primary-card good">
      <div class="card-icon">◍</div>
      <div class="card-text">
        <strong>Local model ready</strong>
        <p>
          Synaptic will use Ollama running on your machine — private and free.
          {#if probe.ollama_models.length}
            Detected: <code>{probe.ollama_models.slice(0, 3).join(', ')}</code>.
          {:else}
            No models pulled yet — run <code>ollama pull llama3.2</code> to get started.
          {/if}
        </p>
        {#if foundClis.length}
          <p class="clis">Also found: <code>{foundClis.join(', ')}</code></p>
        {/if}
      </div>
    </div>
  {:else}
    <div class="primary-card warn">
      <div class="card-icon">○</div>
      <div class="card-text">
        <strong>No local model detected</strong>
        <p>
          Install <button class="link" onclick={() => openExternal('https://ollama.com/download')}>Ollama</button> for a
          free local model, or add an API key for a cloud provider. You can do either after setup.
        </p>
      </div>
    </div>
  {/if}

  <div class="optional-cards">
    <button class="opt-card" onclick={() => finish('#/settings')} disabled={onboarding.busy}>
      <span class="opt-title">Add an API key</span>
      <span class="opt-sub">Use OpenAI, Anthropic, Google… → Settings</span>
    </button>
    <button class="opt-card" onclick={() => finish('#/settings')} disabled={onboarding.busy}>
      <span class="opt-title">Connect messaging</span>
      <span class="opt-sub">Telegram and more → Settings</span>
    </button>
  </div>

  {#if hotkey}
    <p class="hotkey-note">Press <kbd>{hotkey}</kbd> anytime to summon Synaptic.</p>
  {/if}

  {#if onboarding.error}
    <p class="error">{onboarding.error}</p>
  {/if}

  <div class="actions center">
    <button class="btn btn-primary big" onclick={() => finish()} disabled={onboarding.busy}>
      {onboarding.busy ? 'Starting…' : 'Start using Synaptic'}
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
    animation: popIn 0.6s var(--transition-spring);
  }
  @keyframes popIn {
    0% { transform: scale(0.8); opacity: 0; }
    50% { transform: scale(1.1); }
    100% { transform: scale(1); opacity: 1; }
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
  .muted {
    color: var(--color-text-muted);
  }
  .primary-card {
    display: flex;
    gap: var(--space-3);
    text-align: left;
    width: 100%;
    padding: var(--space-4);
    border-radius: var(--radius-lg);
    border: 1px solid var(--glass-border);
    background: var(--glass-bg);
    margin-bottom: var(--space-4);
  }
  .primary-card.good {
    border-color: var(--color-success);
    background: rgba(74, 222, 128, 0.06);
  }
  .card-icon {
    font-size: 28px;
    color: var(--color-accent);
  }
  .card-text strong {
    display: block;
    margin-bottom: 4px;
  }
  .card-text p {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
    margin: 0;
    line-height: 1.5;
  }
  code {
    font-family: var(--font-mono);
    font-size: 0.9em;
    background: rgba(0, 0, 0, 0.3);
    padding: 1px 5px;
    border-radius: 4px;
  }
  .link {
    background: none;
    border: none;
    color: var(--color-accent);
    cursor: pointer;
    text-decoration: underline;
    font: inherit;
    padding: 0;
  }
  .optional-cards {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3);
    width: 100%;
    margin-bottom: var(--space-4);
  }
  .opt-card {
    text-align: left;
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
  .opt-card:hover:not(:disabled) {
    border-color: var(--color-accent);
    transform: translateY(-2px);
  }
  .opt-card:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .opt-title {
    font-weight: 600;
    font-size: var(--size-md);
  }
  .opt-sub {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
  }
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
  .actions.center {
    display: flex;
    justify-content: center;
  }
  .btn {
    padding: 12px 24px;
    border-radius: var(--radius-pill);
    font-size: var(--size-md);
    font-weight: 500;
    cursor: pointer;
    border: none;
    transition: all var(--transition-spring);
  }
  .btn.big {
    padding: 14px 36px;
    font-size: var(--size-lg);
  }
  .btn-primary {
    background: var(--color-accent-gradient);
    color: white;
  }
  .btn-primary:hover:not(:disabled) {
    box-shadow: var(--shadow-glow);
    transform: translateY(-1px);
  }
  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .error {
    color: var(--color-error);
    font-size: var(--size-sm);
  }
</style>
