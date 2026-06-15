<script lang="ts">
  import { onMount } from 'svelte'
  import { onboarding } from '../stores/onboarding.svelte'
  import { apiKeys } from '../stores/apikeys.svelte'
  import { settings } from '../stores/settings.svelte'
  import { updateStore } from '../stores/update.svelte'
  import { ipc } from '../ipc/client'

  let testing = $state(false)
  let testResult = $state('')
  let voiceTesting = $state(false)
  let voiceResult = $state('')

  onMount(() => {
    void ipc.onboardingState().then((s) => {
      if (s.completed_at) {
        onboarding.state.step = 'done'
      }
    }).catch(() => {})
  })

  async function advanceDaemon(): Promise<void> {
    try {
      await ipc.onboardingAdvance()
    } catch {
      // daemon onboarding is best-effort alongside local wizard
    }
  }

  async function testConnection(): Promise<void> {
    if (!onboarding.state.apiKey) {
      testResult = 'Enter an API key first.'
      return
    }
    testing = true
    testResult = ''
    try {
      // Save the key first, then send a minimal chat.
      await apiKeys.set(onboarding.state.provider, 'default', onboarding.state.apiKey)
      const r = await ipc.llmChat(onboarding.state.provider, '', {
        model: '',
        messages: [{ role: 'user', content: 'ping' }],
        max_tokens: 8
      })
      testResult = `✓ Connection OK · ${r.response.usage.total_tokens} tokens used`
    } catch (err) {
      testResult = `✗ ${err}`
    } finally {
      testing = false
    }
  }

  async function testVoice(): Promise<void> {
    voiceTesting = true
    voiceResult = ''
    try {
      const status = await ipc.call<{ status?: string }>('voice.status', {})
      if (status.status === 'idle') {
        voiceResult = '✓ Voice pipeline ready'
      } else if (status.status === 'disabled') {
        voiceResult = '⚠ Voice disabled — you can enable it later in Settings'
      } else {
        voiceResult = `✓ Voice status: ${status.status ?? 'unknown'}`
      }
    } catch (err) {
      voiceResult = `⚠ Voice not available: ${err}`
    } finally {
      voiceTesting = false
    }
  }

  async function finish(): Promise<void> {
    if (settings.config) {
      // Persist the chosen hotkey + telemetry pref.
      await settings.save({
        hotkey: { ...settings.config.hotkey, overlay: onboarding.state.hotkey },
        telemetry: { ...settings.config.telemetry, enabled: onboarding.state.telemetryEnabled }
      })
    }
    updateStore.setEnabled(true) // auto-update still on by default
    try {
      await ipc.onboardingComplete()
    } catch {
      // best-effort
    }
    await ipc.firstRunComplete()
    onboarding.state.step = 'done'
  }

  function goNext(): void {
    onboarding.nextStep()
    void advanceDaemon()
  }

  function goBack(): void {
    onboarding.prevStep()
  }
</script>

{#if onboarding.state.step !== undefined}
<div class="wizard-container">
  <div class="step-indicator">
    {#each ['welcome', 'provider', 'apikey', 'test', 'voice', 'hotkey', 'privacy', 'done'] as step}
      <div 
        class="step-dot" 
        class:active={onboarding.state.step === step} 
        class:past={['welcome', 'provider', 'apikey', 'test', 'voice', 'hotkey', 'privacy', 'done'].indexOf(step) < ['welcome', 'provider', 'apikey', 'test', 'voice', 'hotkey', 'privacy', 'done'].indexOf(onboarding.state.step)}>
      </div>
    {/each}
  </div>

  {#if onboarding.state.step === 'welcome'}
    <div class="wizard">
      <h1>Welcome to Synaptic</h1>
      <p class="lede">A free, on-device AI agent. Your keys, your data, your machine.</p>
      <p class="muted">This quick setup will take about 30 seconds.</p>
      <div class="actions center">
        <button class="btn btn-primary" onclick={() => goNext()}>Get started →</button>
      </div>
    </div>
  {:else if onboarding.state.step === 'provider'}
    <div class="wizard">
      <h2>Pick a provider</h2>
      <p class="muted">You can add more providers later in Settings → API keys.</p>
      <div class="provider-grid">
        {#each ['openai', 'anthropic', 'google', 'xai', 'mistral', 'deepseek', 'openrouter', 'groq'] as p}
          <button
            class="provider-tile"
            class:selected={onboarding.state.provider === p}
            onclick={() => onboarding.state.provider = p}
          >
            {p.charAt(0).toUpperCase() + p.slice(1)}
          </button>
        {/each}
      </div>
      <div class="actions">
        <button class="btn btn-ghost" onclick={() => goBack()}>← Back</button>
        <button class="btn btn-primary" onclick={() => goNext()}>Next →</button>
      </div>
    </div>
  {:else if onboarding.state.step === 'apikey'}
    <div class="wizard">
      <h2>Enter your <code>{onboarding.state.provider}</code> API key</h2>
      <p class="muted">Stored encrypted in the OS keyring. We never see it.</p>
      <input
        type="password"
        bind:value={onboarding.state.apiKey}
        placeholder="sk-…"
        class="input"
        autocomplete="off"
      />
      <div class="actions">
        <button class="btn btn-ghost" onclick={() => goBack()}>← Back</button>
        <button
          class="btn btn-primary"
          onclick={() => goNext()}
          disabled={!onboarding.state.apiKey}
        >
          Next →
        </button>
      </div>
    </div>
  {:else if onboarding.state.step === 'test'}
    <div class="wizard">
      <h2>Test the connection</h2>
      <p class="muted">We'll send a tiny "ping" message to verify the key works.</p>
      <button class="btn btn-secondary" onclick={testConnection} disabled={testing}>
        {testing ? 'Testing…' : 'Test connection'}
      </button>
      {#if testResult}
        <p class="test-result" class:success={testResult.startsWith('✓')} class:error={testResult.startsWith('✗')}>{testResult}</p>
      {/if}
      <div class="actions">
        <button class="btn btn-ghost" onclick={() => goBack()}>← Back</button>
        <button class="btn btn-primary" onclick={() => goNext()}>Next →</button>
      </div>
    </div>
  {:else if onboarding.state.step === 'voice'}
    <div class="wizard">
      <h2>Test your voice</h2>
      <p class="muted">Synaptic can listen and speak. Test your microphone and speaker now, or skip this step.</p>
      <button class="btn btn-secondary" onclick={testVoice} disabled={voiceTesting}>
        {voiceTesting ? 'Checking…' : 'Check voice status'}
      </button>
      {#if voiceResult}
        <p class="test-result" class:success={voiceResult.startsWith('✓')} class:error={voiceResult.startsWith('✗')}>{voiceResult}</p>
      {/if}
      <div class="actions">
        <button class="btn btn-ghost" onclick={() => goBack()}>← Back</button>
        <button class="btn btn-primary" onclick={() => goNext()}>Next →</button>
      </div>
    </div>
  {:else if onboarding.state.step === 'hotkey'}
    <div class="wizard">
      <h2>Pick your hotkey</h2>
      <p class="muted">This summons the Synaptic overlay. Default: <code>Cmd+Shift+Space</code> (macOS) or <code>Ctrl+Shift+Space</code> (Win/Linux). You can change it later in Settings.</p>
      <input
        type="text"
        bind:value={onboarding.state.hotkey}
        placeholder="Cmd+Shift+Space"
        class="input"
      />
      <div class="actions">
        <button class="btn btn-ghost" onclick={() => goBack()}>← Back</button>
        <button class="btn btn-primary" onclick={() => goNext()}>Next →</button>
      </div>
    </div>
  {:else if onboarding.state.step === 'privacy'}
    <div class="wizard">
      <h2>Privacy</h2>
      <p class="muted">By default, Synaptic does <strong>not</strong> collect any telemetry. If you opt in below, we'll only ever send:</p>
      <ul>
        <li>Your app version + OS (e.g. <code>synaptic 0.1.0 / darwin/arm64</code>)</li>
        <li>Anonymous command counters (e.g. "user sent 12 messages today")</li>
        <li>Crash signatures (stack-trace hashes, never source)</li>
      </ul>
      <p class="muted">Never your prompts, files, or anything that identifies you.</p>
      <label class="checkbox">
        <input
          type="checkbox"
          checked={onboarding.state.telemetryEnabled}
          onchange={(e) => onboarding.state.telemetryEnabled = (e.target as HTMLInputElement).checked}
        />
        <span>Send anonymous usage stats (off by default)</span>
      </label>
      <div class="actions">
        <button class="btn btn-ghost" onclick={() => goBack()}>← Back</button>
        <button class="btn btn-primary" onclick={finish}>Finish setup</button>
      </div>
    </div>
  {:else if onboarding.state.step === 'done'}
    <div class="wizard done">
      <div class="done-check">✓</div>
      <h1 class="gradient-text">All set</h1>
      <p class="lede">Press <code>{onboarding.state.hotkey}</code> anywhere to summon the overlay.</p>
      <div class="actions center">
        <button class="btn btn-primary" onclick={() => onboarding.reset()}>Start chatting →</button>
      </div>
    </div>
  {/if}
</div>
{/if}

<style>
  .wizard-container {
    position: relative;
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }
  
  .wizard-container::before {
    content: '';
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 800px;
    height: 800px;
    background: radial-gradient(circle, var(--color-accent) 0%, transparent 60%);
    opacity: 0.05;
    pointer-events: none;
    z-index: -1;
  }

  .step-indicator {
    position: absolute;
    top: var(--space-8);
    display: flex;
    gap: var(--space-2);
  }
  .step-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.1);
    transition: all var(--transition-base);
  }
  .step-dot.past {
    background: var(--color-accent);
    opacity: 0.5;
  }
  .step-dot.active {
    background: var(--color-accent);
    box-shadow: var(--shadow-glow);
    transform: scale(1.2);
  }

  .wizard {
    width: 100%;
    max-width: 540px;
    padding: var(--space-6) var(--space-5);
    text-align: center;
  }
  .wizard h1 {
    font-size: var(--size-3xl);
    font-weight: 600;
    margin-bottom: var(--space-3);
    background: var(--color-accent-gradient);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .wizard h2 {
    font-size: var(--size-2xl);
    font-weight: 600;
    margin-bottom: var(--space-3);
  }
  .wizard .lede {
    font-size: var(--size-lg);
    color: var(--color-text-muted);
    margin-bottom: var(--space-5);
  }
  .wizard .muted {
    color: var(--color-text-muted);
    font-size: var(--size-md);
    margin-bottom: var(--space-5);
  }
  .wizard ul {
    text-align: left;
    color: var(--color-text-muted);
    margin: 0 auto var(--space-5) auto;
    max-width: 360px;
    list-style: disc inside;
  }
  
  .provider-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: var(--space-3);
    margin: var(--space-5) 0;
  }
  .provider-tile {
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    color: var(--color-text);
    padding: var(--space-4);
    border-radius: var(--radius-lg);
    font-size: var(--size-md);
    font-weight: 500;
    cursor: pointer;
    transition: all var(--transition-base);
  }
  .provider-tile:hover {
    background: rgba(255, 255, 255, 0.06);
    transform: translateY(-2px);
  }
  .provider-tile.selected {
    border-color: var(--color-accent);
    background: var(--color-accent-soft);
    box-shadow: var(--shadow-glow);
  }
  
  .input {
    width: 100%;
    background: rgba(0, 0, 0, 0.3);
    border: 1px solid var(--glass-border);
    color: var(--color-text);
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-lg);
    font-size: var(--size-lg);
    margin-bottom: var(--space-4);
    transition: all var(--transition-base);
  }
  .input:focus {
    outline: none;
    border-color: var(--color-accent);
    box-shadow: var(--shadow-glow);
  }
  
  .actions {
    display: flex;
    justify-content: space-between;
    margin-top: var(--space-5);
  }
  .actions.center {
    justify-content: center;
  }
  
  .btn {
    padding: 12px 24px;
    border-radius: var(--radius-pill);
    font-size: var(--size-md);
    font-weight: 500;
    cursor: pointer;
    transition: all var(--transition-spring);
    border: none;
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
  .btn-secondary {
    background: var(--glass-bg);
    color: var(--color-text);
    border: 1px solid var(--glass-border);
    margin: var(--space-3) auto;
  }
  .btn-secondary:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.08);
  }
  .btn-ghost {
    background: transparent;
    color: var(--color-text-muted);
    border: 1px solid var(--glass-border);
  }
  .btn-ghost:hover {
    color: var(--color-text);
    border-color: rgba(255, 255, 255, 0.15);
  }
  
  .checkbox {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    justify-content: center;
    margin: var(--space-4) 0;
    cursor: pointer;
  }
  
  .test-result {
    margin: var(--space-3) 0;
    font-family: var(--font-mono);
    font-size: var(--size-sm);
  }
  .test-result.success { color: var(--color-success); }
  .test-result.error { color: var(--color-error); }
  
  @keyframes popIn {
    0% { transform: scale(0.8); opacity: 0; }
    50% { transform: scale(1.1); }
    100% { transform: scale(1); opacity: 1; }
  }
  .done-check {
    font-size: 64px;
    color: var(--color-accent);
    margin-bottom: var(--space-4);
    animation: popIn 0.6s var(--transition-spring);
  }
  .gradient-text {
    background: var(--color-accent-gradient);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }
</style>
