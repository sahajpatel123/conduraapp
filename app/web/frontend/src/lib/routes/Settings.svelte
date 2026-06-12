<script lang="ts">
  import { settings } from '../stores/settings.svelte'
  import { apiKeys } from '../stores/apikeys.svelte'
  import { updateStore } from '../stores/update.svelte'
  import { halt } from '../stores/halt.svelte'
  import { spend } from '../stores/spend.svelte'

  let hotkeyInput = $state('')
  let telemetryInput = $state(false)
  let newProvider = $state('openai')
  let newLabel = $state('default')
  let newSecret = $state('')
  let settingAPIKey = $state(false)

  $effect(() => {
    if (settings.config) {
      hotkeyInput = settings.config.hotkey.overlay
      telemetryInput = settings.config.telemetry.enabled
    }
  })

  async function saveHotkey(): Promise<void> {
    if (!settings.config) return
    await settings.save({ hotkey: { ...settings.config.hotkey, overlay: hotkeyInput } })
    alert('Hotkey saved. Restart the app to apply.')
  }

  async function saveTelemetry(): Promise<void> {
    if (!settings.config) return
    await settings.save({ telemetry: { ...settings.config.telemetry, enabled: telemetryInput } })
    updateStore.setEnabled(telemetryInput)
  }

  async function setAPIKey(): Promise<void> {
    if (!newSecret) return
    settingAPIKey = true
    try {
      await apiKeys.set(newProvider, newLabel, newSecret)
      newSecret = ''
      alert('API key saved.')
    } catch (err) {
      alert(`Failed: ${err}`)
    } finally {
      settingAPIKey = false
    }
  }

  async function deleteKey(id: number): Promise<void> {
    if (!confirm('Delete this API key?')) return
    await apiKeys.remove(id)
  }

  async function performHalt(): Promise<void> {
    if (!confirm('Halt all daemon activity? Streaming responses will be cancelled.')) return
    await halt.halt('user requested from settings')
  }

  async function performResume(): Promise<void> {
    await halt.resume()
  }
</script>

<div class="settings-page">
  <header>
    <h2>Settings</h2>
    <p class="muted">Configuration is stored in <code>~/.synaptic/config.yaml</code>. Changes here write to that file via the daemon.</p>
  </header>

  <section class="card">
    <h3>Spend</h3>
    {#if spend.summary}
      <div class="kv">
        <span class="k">Spent today</span><span class="v">${spend.summary.spent.toFixed(2)}</span>
      </div>
      <div class="kv">
        <span class="k">Cap</span><span class="v">${spend.summary.cap.toFixed(2)}</span>
      </div>
      <div class="kv">
        <span class="k">Remaining</span><span class="v">${spend.summary.remaining.toFixed(2)}</span>
      </div>
    {:else}
      <p class="muted">Loading…</p>
    {/if}
  </section>

  <section class="card">
    <h3>Hotkey</h3>
    <p class="muted">Press the key combination you want to use to summon the overlay. On macOS, <code>Cmd</code> is <code>Super</code> in the underlying API.</p>
    <div class="row">
      <input
        type="text"
        bind:value={hotkeyInput}
        placeholder="Cmd+Shift+Space"
        class="input"
      />
      <button class="btn btn-primary" onclick={saveHotkey}>Save</button>
    </div>
  </section>

  <section class="card">
    <h3>Auto-update</h3>
    <p class="muted">Synaptic auto-updates by default. Disable here to opt out — but you'll need to update manually going forward.</p>
    <label class="checkbox">
      <input
        type="checkbox"
        checked={telemetryInput}
        onchange={(e) => { telemetryInput = (e.target as HTMLInputElement).checked; void saveTelemetry(); }}
      />
      <span>Enable auto-updates</span>
    </label>
    {#if updateStore.lastCheck}
      <p class="muted">Last checked: {new Date(updateStore.lastCheck).toLocaleString()}</p>
    {/if}
  </section>

  <section class="card danger">
    <h3>Kill switch</h3>
    <p class="muted">Halt every active stream and pause the daemon. Use this if an agent is doing something you don't want.</p>
    {#if halt.state.halted}
      <p class="muted">⚠ Daemon is currently halted since {halt.state.since}.</p>
      <button class="btn btn-primary" onclick={performResume}>Resume daemon</button>
    {:else}
      <button class="btn btn-danger" onclick={performHalt}>HALT</button>
    {/if}
  </section>

  <section class="card">
    <h3>API keys</h3>
    <p class="muted">Stored encrypted in the OS keyring (or in <code>~/.synaptic/secrets.json</code> with 0600 perms if the keyring is unavailable).</p>

    <div class="apikey-list">
      {#if apiKeys.list.length === 0}
        <p class="muted">No API keys stored yet.</p>
      {/if}
      {#each apiKeys.list as k (k.id)}
        <div class="apikey-row">
          <span class="provider">{k.provider}</span>
          <span class="label">{k.label}</span>
          <span class="auth-kind">{k.auth_kind}</span>
          <span class="has-token">{k.has_token ? '✓ has token' : '✗ no token'}</span>
          <button class="btn btn-ghost" onclick={() => deleteKey(k.id)}>Delete</button>
        </div>
      {/each}
    </div>

    <h4>Add a key</h4>
    <div class="row">
      <select bind:value={newProvider} class="input">
        <option value="openai">openai</option>
        <option value="anthropic">anthropic</option>
        <option value="google">google</option>
        <option value="xai">xai</option>
        <option value="mistral">mistral</option>
        <option value="deepseek">deepseek</option>
        <option value="openrouter">openrouter</option>
        <option value="groq">groq</option>
        <option value="together">together</option>
        <option value="fireworks">fireworks</option>
      </select>
      <input type="text" bind:value={newLabel} placeholder="label" class="input" />
      <input
        type="password"
        bind:value={newSecret}
        placeholder="sk-…"
        class="input"
        autocomplete="off"
      />
      <button
        class="btn btn-primary"
        onclick={setAPIKey}
        disabled={!newSecret || settingAPIKey}
      >
        {settingAPIKey ? 'Saving…' : 'Save'}
      </button>
    </div>
  </section>
</div>

<style>
  .settings-page {
    padding: var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: 760px;
    margin: 0 auto;
  }
  .settings-page header h2 {
    font-size: var(--size-2xl);
    font-weight: 600;
    margin-bottom: var(--space-2);
    background: var(--color-accent-gradient);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .muted {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
  }
  .card {
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-xl);
    padding: var(--space-5);
    margin-top: var(--space-5);
    transition: border-color var(--transition-base);
  }
  .card:hover {
    border-color: rgba(255,255,255,0.12);
  }
  .card.danger {
    background: rgba(239, 68, 68, 0.04);
    border-color: rgba(239, 68, 68, 0.2);
  }
  .card h3 {
    font-size: var(--size-lg);
    font-weight: 600;
    margin-bottom: var(--space-3);
  }
  .card h4 {
    font-size: var(--size-md);
    font-weight: 600;
    margin: var(--space-4) 0 var(--space-2) 0;
  }
  .row {
    display: flex;
    gap: var(--space-2);
    align-items: center;
    margin-top: var(--space-3);
  }
  .input,
  .select {
    background: rgba(0, 0, 0, 0.3);
    border: 1px solid var(--glass-border);
    color: var(--color-text);
    padding: 8px 12px;
    border-radius: var(--radius-md);
    font-size: var(--size-md);
    flex: 1;
    transition: all var(--transition-base);
  }
  .input:focus,
  .select:focus {
    outline: none;
    border-color: var(--color-accent);
    box-shadow: var(--shadow-glow);
  }
  .btn {
    padding: 8px 16px;
    border-radius: var(--radius-md);
    font-size: var(--size-md);
    font-weight: 500;
    white-space: nowrap;
    cursor: pointer;
    transition: all var(--transition-base);
    border: none;
  }
  .btn-primary {
    background: var(--color-accent-gradient);
    color: white;
  }
  .btn-primary:hover:not(:disabled) {
    box-shadow: var(--shadow-glow);
  }
  .btn-ghost {
    background: transparent;
    color: var(--color-text-muted);
    border: 1px solid var(--glass-border);
  }
  .btn-ghost:hover {
    color: var(--color-text);
    border-color: rgba(255,255,255,0.15);
  }
  .btn-danger {
    background: linear-gradient(135deg, #ef4444, #dc2626);
    color: white;
    font-weight: 600;
  }
  .btn-danger:hover:not(:disabled) {
    box-shadow: 0 0 15px rgba(239, 68, 68, 0.3);
  }
  .checkbox {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-top: var(--space-3);
    cursor: pointer;
  }
  .kv {
    display: flex;
    justify-content: space-between;
    padding: var(--space-2) 0;
    font-size: var(--size-md);
    border-bottom: 1px dotted var(--glass-border);
  }
  .kv:last-child {
    border-bottom: none;
  }
  .kv .k {
    color: var(--color-text-muted);
  }
  .kv .v {
    color: var(--color-text);
    font-family: var(--font-mono);
  }
  .apikey-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin-top: var(--space-3);
  }
  .apikey-row {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr 1fr auto;
    gap: var(--space-2);
    padding: var(--space-3);
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-md);
    font-size: var(--size-sm);
    align-items: center;
  }
  .apikey-row .provider {
    font-weight: 600;
  }
  .apikey-row .has-token {
    color: var(--color-text-muted);
    font-family: var(--font-mono);
    font-size: var(--size-xs);
  }
</style>
