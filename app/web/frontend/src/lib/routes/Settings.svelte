<script lang="ts">
  import { settings } from '../stores/settings.svelte'
  import { apiKeys } from '../stores/apikeys.svelte'
  import { updateStore } from '../stores/update.svelte'
  import { halt } from '../stores/halt.svelte'
  import { spend } from '../stores/spend.svelte'
  import { trust } from '../stores/trust.svelte'
  import { onboarding } from '../stores/onboarding.svelte'
  import { ipc } from '../ipc/client'
  import LocaleSelector from '../components/LocaleSelector.svelte'

  let hotkeyInput = $state('')
  let telemetryInput = $state(false)
  let newProvider = $state('openai')
  let newLabel = $state('default')
  let newSecret = $state('')
  let settingAPIKey = $state(false)
  let creatingBackup = $state(false)
  let permissionGuide = $state<{ kind: string; title: string; steps: string[] } | null>(null)
  let eulaText = $state('')
  let eulaTitle = $state('')
  let eulaVersion = $state('')
  let rerunning = $state(false)

  async function viewEula(): Promise<void> {
    try {
      const doc = await ipc.onboardingEula()
      eulaText = doc.text
      eulaTitle = 'Synaptic Freeware License'
      eulaVersion = doc.version
    } catch (err) {
      alert(`Could not load the EULA: ${err}`)
    }
  }

  async function rerunSetup(): Promise<void> {
    if (!confirm('Re-run the setup wizard? Your data and settings are not affected.')) return
    rerunning = true
    try {
      await onboarding.reset()
      window.dispatchEvent(new CustomEvent('synaptic:show-onboarding'))
    } catch (err) {
      alert(`Could not reset setup: ${err}`)
    } finally {
      rerunning = false
    }
  }

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

  async function createBackup(): Promise<void> {
    creatingBackup = true
    try {
      const path = await trust.createBackup()
      alert(`Backup created:\n${path}`)
    } catch (err) {
      alert(`Backup failed: ${err}`)
    } finally {
      creatingBackup = false
    }
  }

  async function showPermissionGuide(kind: string): Promise<void> {
    try {
      const g = await trust.loadGuide(kind)
      permissionGuide = { kind: g.kind, title: g.title, steps: g.steps }
    } catch (err) {
      alert(`Could not load guide: ${err}`)
    }
  }

  function formatBytes(n: number): string {
    if (n < 1024) return `${n} B`
    if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
    return `${(n / (1024 * 1024)).toFixed(1)} MB`
  }
</script>

<div class="settings-page">
  <header>
    <h2>Settings</h2>
    <p class="muted">Configuration is stored in <code>~/.synaptic/config.yaml</code>. Changes here write to that file via the daemon.</p>
  </header>

  <section class="card">
    <h3>Language</h3>
    <p class="muted">Select your preferred language. Changes apply immediately.</p>
    <div class="row">
      <LocaleSelector />
    </div>
  </section>

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

  <section class="card">
    <h3>Backups</h3>
    <p class="muted">Encrypted archives are stored in <code>~/Documents/synaptic-backups</code> (or <code>SYNAPTIC_BACKUP_DIR</code>).</p>
    <div class="row">
      <button class="btn btn-primary" onclick={createBackup} disabled={creatingBackup}>
        {creatingBackup ? 'Creating…' : 'Create backup now'}
      </button>
      <button class="btn btn-ghost" onclick={() => trust.refreshBackups()}>Refresh list</button>
    </div>
    {#if trust.loadingBackups}
      <p class="muted">Loading backups…</p>
    {:else if trust.backups.length === 0}
      <p class="muted">No backups yet.</p>
    {:else}
      <div class="backup-list">
        {#each trust.backups as b (b.path)}
          <div class="backup-row">
            <span class="backup-name">{b.name}</span>
            <span class="backup-size">{formatBytes(b.size)}</span>
          </div>
        {/each}
      </div>
    {/if}
  </section>

  <section class="card">
    <h3>OS permissions</h3>
    <p class="muted">Synaptic needs OS permissions for accessibility, screen recording, and microphone. Grant them in System Settings if status is not granted.</p>
    <button class="btn btn-ghost" onclick={() => trust.refreshPermissions()}>Refresh status</button>
    {#if trust.loadingPermissions}
      <p class="muted">Checking permissions…</p>
    {:else}
      <div class="perm-list">
        {#each trust.permissions as p (p.kind)}
          <div class="perm-row">
            <span class="perm-kind">{p.kind}</span>
            <span class="perm-status" class:granted={p.status === 'granted'} class:denied={p.status === 'denied'}>{p.status}</span>
            {#if p.status !== 'granted'}
              <button class="btn btn-ghost" onclick={() => showPermissionGuide(p.kind)}>How to grant</button>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
    {#if permissionGuide}
      <div class="guide-box">
        <h4>{permissionGuide.title}</h4>
        <ol>
          {#each permissionGuide.steps as step}
            <li>{step}</li>
          {/each}
        </ol>
        <button class="btn btn-ghost" onclick={() => { permissionGuide = null }}>Close</button>
      </div>
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

  <section class="card">
    <h3>Legal</h3>
    <p class="muted">Review the license you accepted during setup. The full text is also available online.</p>
    {#if eulaText}
      <div class="eula-view">
        <div class="eula-view-head">
          <strong>{eulaTitle}</strong>
          <span class="muted">{eulaVersion}</span>
        </div>
        <pre>{eulaText}</pre>
      </div>
      <button class="btn btn-ghost" onclick={() => { eulaText = '' }}>Hide</button>
    {:else}
      <button class="btn btn-ghost" onclick={viewEula}>View EULA</button>
    {/if}
  </section>

  <section class="card">
    <h3>Setup</h3>
    <p class="muted">Run the first-time setup wizard again — EULA, permissions, and hotkey. Your data is not affected.</p>
    <button class="btn btn-ghost" onclick={rerunSetup} disabled={rerunning}>
      {rerunning ? 'Resetting…' : 'Re-run setup'}
    </button>
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
  .eula-view {
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
    margin: var(--space-3) 0;
    max-height: 280px;
    overflow-y: auto;
  }
  .eula-view-head {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    margin-bottom: var(--space-2);
  }
  .eula-view pre {
    white-space: pre-wrap;
    word-wrap: break-word;
    font-family: var(--font-sans);
    font-size: var(--size-sm);
    line-height: 1.6;
    color: var(--color-text-muted);
    margin: 0;
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
  .input {
    background: rgba(0, 0, 0, 0.3);
    border: 1px solid var(--glass-border);
    color: var(--color-text);
    padding: 8px 12px;
    border-radius: var(--radius-md);
    font-size: var(--size-md);
    flex: 1;
    transition: all var(--transition-base);
  }
  .input:focus {
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
  .backup-list, .perm-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin-top: var(--space-3);
  }
  .backup-row, .perm-row {
    display: flex;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-2) var(--space-3);
    background: rgba(0,0,0,0.2);
    border-radius: var(--radius-md);
    font-size: var(--size-sm);
  }
  .backup-name, .perm-kind {
    flex: 1;
    font-family: var(--font-mono);
  }
  .backup-size {
    color: var(--color-text-muted);
  }
  .perm-status.granted { color: var(--color-success); }
  .perm-status.denied { color: #f87171; }
  .guide-box {
    margin-top: var(--space-4);
    padding: var(--space-4);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-md);
  }
  .guide-box ol {
    margin: var(--space-3) 0;
    padding-left: var(--space-5);
  }
</style>
