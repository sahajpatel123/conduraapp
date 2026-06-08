<script lang="ts">
  import { onboarding } from '../stores/onboarding.svelte'
  import { apiKeys } from '../stores/apikeys.svelte'
  import { settings } from '../stores/settings.svelte'
  import { updateStore } from '../stores/update.svelte'
  import { ipc } from '../ipc/client'

  let testing = $state(false)
  let testResult = $state('')

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

  async function finish(): Promise<void> {
    if (settings.config) {
      // Persist the chosen hotkey + telemetry pref.
      await settings.save({
        hotkey: { ...settings.config.hotkey, overlay: onboarding.state.hotkey },
        telemetry: { ...settings.config.telemetry, enabled: onboarding.state.telemetryEnabled }
      })
    }
    updateStore.setEnabled(true) // auto-update still on by default
    await ipc.firstRunComplete()
    onboarding.nextStep()
  }
</script>

{#if onboarding.state.step === 'welcome'}
  <div class="wizard">
    <h1>Welcome to Synaptic</h1>
    <p class="lede">A free, on-device AI agent. Your keys, your data, your machine.</p>
    <p class="muted">This quick setup will take about 30 seconds.</p>
    <div class="actions">
      <button class="btn btn-primary" onclick={() => onboarding.nextStep()}>Get started →</button>
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
          {p}
        </button>
      {/each}
    </div>
    <div class="actions">
      <button class="btn btn-ghost" onclick={() => onboarding.prevStep()}>← Back</button>
      <button class="btn btn-primary" onclick={() => onboarding.nextStep()}>Next →</button>
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
      <button class="btn btn-ghost" onclick={() => onboarding.prevStep()}>← Back</button>
      <button
        class="btn btn-primary"
        onclick={() => onboarding.nextStep()}
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
      <p class="test-result">{testResult}</p>
    {/if}
    <div class="actions">
      <button class="btn btn-ghost" onclick={() => onboarding.prevStep()}>← Back</button>
      <button class="btn btn-primary" onclick={() => onboarding.nextStep()}>Next →</button>
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
      <button class="btn btn-ghost" onclick={() => onboarding.prevStep()}>← Back</button>
      <button class="btn btn-primary" onclick={() => onboarding.nextStep()}>Next →</button>
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
      <button class="btn btn-ghost" onclick={() => onboarding.prevStep()}>← Back</button>
      <button class="btn btn-primary" onclick={finish}>Finish setup</button>
    </div>
  </div>
{:else if onboarding.state.step === 'done'}
  <div class="wizard done">
    <h1>All set ✓</h1>
    <p class="lede">Press <code>{onboarding.state.hotkey}</code> anywhere to summon the overlay.</p>
    <div class="actions">
      <button class="btn btn-primary" onclick={() => onboarding.reset()}>Start chatting →</button>
    </div>
  </div>
{/if}

<style>
  .wizard {
    max-width: 540px;
    margin: 0 auto;
    padding: var(--space-8) var(--space-5);
    text-align: center;
  }
  .wizard h1 {
    font-size: var(--size-3xl);
    font-weight: 600;
    margin-bottom: var(--space-3);
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
    gap: var(--space-2);
    margin: var(--space-5) 0;
  }
  .provider-tile {
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
    color: var(--color-text);
    padding: var(--space-3);
    border-radius: var(--radius-md);
    font-size: var(--size-md);
  }
  .provider-tile.selected {
    border-color: var(--color-accent);
    background: var(--color-accent-soft);
  }
  .input {
    width: 100%;
    background: var(--color-bg);
    border: 1px solid var(--color-border);
    color: var(--color-text);
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    font-size: var(--size-md);
    margin-bottom: var(--space-4);
  }
  .input:focus {
    outline: none;
    border-color: var(--color-accent);
    box-shadow: var(--shadow-focus);
  }
  .actions {
    display: flex;
    justify-content: space-between;
    margin-top: var(--space-5);
  }
  .btn {
    padding: 10px 20px;
    border-radius: var(--radius-md);
    font-size: var(--size-md);
    font-weight: 500;
  }
  .btn-primary {
    background: var(--color-accent);
    color: white;
  }
  .btn-primary:hover:not(:disabled) {
    background: var(--color-accent-hover);
  }
  .btn-secondary {
    background: var(--color-bg-elevated);
    color: var(--color-text);
    border: 1px solid var(--color-border);
    margin: var(--space-3) auto;
  }
  .btn-ghost {
    background: transparent;
    color: var(--color-text-muted);
    border: 1px solid var(--color-border);
  }
  .checkbox {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    justify-content: center;
    margin: var(--space-4) 0;
  }
  .test-result {
    margin: var(--space-3) 0;
    font-family: var(--font-mono);
    font-size: var(--size-sm);
    color: var(--color-text-muted);
  }
</style>
