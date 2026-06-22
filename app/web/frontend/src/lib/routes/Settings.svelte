<script lang="ts">
  import { settings } from '../stores/settings.svelte'
  import { apiKeys } from '../stores/apikeys.svelte'
  import { updateStore } from '../stores/update.svelte'
  import { halt } from '../stores/halt.svelte'
  import { spend } from '../stores/spend.svelte'
  import { trust } from '../stores/trust.svelte'
  import { onboarding } from '../stores/onboarding.svelte'
  import { account } from '../stores/account.svelte'
  import { ipc } from '../ipc/client'
  import { onMount } from 'svelte'
  import LocaleSelector from '../components/LocaleSelector.svelte'
  import SignInPanel from '../components/SignInPanel.svelte'

  let hotkeyInput = $state('')
  let telemetryInput = $state(false)
  let newProvider = $state('openai')
  let newLabel = $state('default')
  let newSecret = $state('')
  let settingAPIKey = $state(false)
  let creatingBackup = $state(false)
  let restoringBackup = $state<string | null>(null)
  let restoreTarget = $state<{ name: string; path: string; size: number } | null>(null)
  let permissionGuide = $state<{ kind: string; title: string; steps: string[] } | null>(null)
  let eulaText = $state('')
  let eulaTitle = $state('')
  let eulaVersion = $state('')
  // Adaptive engine (Phase 14I)
  let adaptiveStrength = $state<'off' | 'cautious' | 'balanced' | 'aggressive'>('balanced')
  let adaptiveProfile = $state<{
    identity?: Record<string, string>
    preferences?: Record<string, string>
    style?: Record<string, string>
    workflows?: string[]
    expertise?: Record<string, number>
    pet_peeves?: string[]
    time_patterns?: Record<string, string>
    tools_habits?: Record<string, number>
    model_prefs?: Record<string, string>
    risk_tolerance?: string
    communication?: Record<string, string>
    last_updated?: string
    version?: number
  } | null>(null)
  let adaptiveLoading = $state(false)
  let adaptiveError = $state<string | null>(null)
  let rerunning = $state(false)

  // Account (14B)
  let showSignIn = $state(false)

  // Voice (14H) — read/written via generic config RPCs because the
  // typed AppConfig doesn't model the voice subtree.
  interface WakeCfg { enabled: boolean; sensitivity: number; hotword: string }
  let wake = $state<WakeCfg>({ enabled: false, sensitivity: 0.5, hotword: 'hey condura' })
  let micTestResult = $state('')

  async function loadVoice(): Promise<void> {
    try {
      const cfg = await ipc.call<{ voice?: { wake?: Partial<WakeCfg> } }>('config.get', {})
      const w = cfg.voice?.wake
      if (w) {
        wake = {
          enabled: w.enabled ?? false,
          sensitivity: w.sensitivity ?? 0.5,
          hotword: w.hotword ?? 'hey condura',
        }
      }
    } catch {
      // keep defaults
    }
  }

  async function saveVoice(): Promise<void> {
    try {
      await ipc.call('config.update', { voice: { wake: { ...wake } } })
    } catch (err) {
      alert(`Could not save voice settings: ${err}`)
    }
  }

  async function micTest(): Promise<void> {
    micTestResult = 'Checking…'
    try {
      const perms = await ipc.permissionsStatus()
      const mic = perms.find((p) => p.kind === 'microphone')
      if (!mic) micTestResult = 'Microphone status unavailable.'
      else if (mic.status === 'granted') micTestResult = 'Microphone access granted ✓'
      else if (mic.status === 'denied') micTestResult = 'Microphone access denied — grant it in OS permissions above.'
      else micTestResult = 'Microphone access not yet granted.'
    } catch (err) {
      micTestResult = `Mic test failed: ${err}`
    }
  }

  function goToChannels(): void {
    window.location.hash = '#/channels'
  }

  onMount(() => {
    void account.checkStatus()
    void loadVoice()
    void loadAdaptive()
  })

  async function loadAdaptive(): Promise<void> {
    adaptiveLoading = true
    adaptiveError = null
    try {
      // Read the engine strength.
      const strengthResp = await ipc.call<{ strength: string }>('adaptive.strength.get', {})
      const s = (strengthResp?.strength ?? 'balanced') as typeof adaptiveStrength
      if (s === 'off' || s === 'cautious' || s === 'balanced' || s === 'aggressive') {
        adaptiveStrength = s
      }
      // Read the user model.
      const profile = await ipc.call<typeof adaptiveProfile>('adaptive.profile', {})
      adaptiveProfile = profile ?? null
    } catch (e) {
      adaptiveError = String(e)
    } finally {
      adaptiveLoading = false
    }
  }

  async function setAdaptiveStrength(s: typeof adaptiveStrength): Promise<void> {
    adaptiveStrength = s
    try {
      await ipc.call('adaptive.strength.set', { strength: s })
    } catch (e) {
      adaptiveError = String(e)
    }
  }

  async function forgetAdaptiveField(field: string, value: string): Promise<void> {
    if (!confirm(`Forget that you ${field === 'pet_peeves' ? 'dislike' : 'have'} "${value}"?`)) return
    try {
      await ipc.call('adaptive.forget', { field, value })
      void loadAdaptive()
    } catch (e) {
      adaptiveError = String(e)
    }
  }

  async function resetAdaptive(): Promise<void> {
    if (!confirm('Delete all learned inferences and start fresh? This cannot be undone.')) return
    try {
      await ipc.call('adaptive.reset', {})
      void loadAdaptive()
    } catch (e) {
      adaptiveError = String(e)
    }
  }

  function providerLabel(p: string): string {
    switch (p) {
      case 'google': return 'Google'
      case 'github': return 'GitHub'
      case 'apple': return 'Apple'
      case 'magic': return 'Email magic link'
      default: return p || 'Account'
    }
  }

  async function viewEula(): Promise<void> {
    try {
      const doc = await ipc.onboardingEula()
      eulaText = doc.text
      eulaTitle = 'Condura Freeware License'
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

  function askRestore(target: { name: string; path: string; size: number }): void {
    restoreTarget = target
  }

  function cancelRestore(): void {
    restoreTarget = null
  }

  async function confirmRestore(): Promise<void> {
    if (!restoreTarget) return
    const target = restoreTarget
    restoreTarget = null
    restoringBackup = target.path
    try {
      await ipc.backupRestore({ path: target.path })
      // The on-disk data is swapped. Refresh any views that
      // show restored data so the user sees the new state
      // immediately without a daemon restart.
      await trust.refreshBackups()
      alert(`Restored from ${target.name}.`)
    } catch (err) {
      alert(`Restore failed: ${err}`)
    } finally {
      restoringBackup = null
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
    <h3>Account</h3>
    {#if account.isSignedIn}
      <div class="account-row">
        {#if account.avatarURL}
          <img class="acc-avatar" src={account.avatarURL} alt="" />
        {:else}
          <span class="acc-avatar fallback">{(account.displayName || '?').charAt(0).toUpperCase()}</span>
        {/if}
        <div class="acc-info">
          <span class="acc-name">{account.displayName || account.email}</span>
          <span class="acc-meta">{account.email} · {providerLabel(account.provider)}{account.tier ? ` · ${account.tier}` : ''}</span>
        </div>
        <button class="btn btn-ghost" onclick={() => account.signOut()} disabled={account.loading}>
          {account.loading ? 'Signing out…' : 'Sign out'}
        </button>
      </div>
    {:else}
      <p class="muted">Condura works fully without an account. Sign in to:</p>
      <ul class="benefits">
        <li>Sync settings and skills across your devices</li>
        <li>Publish skills to the Hub under your identity</li>
        <li>Back up your encrypted data to the cloud</li>
      </ul>
      <div class="row">
        <button class="btn btn-primary" onclick={() => (showSignIn = true)}>Sign in</button>
      </div>
      {#if account.error}<p class="muted err">{account.error}</p>{/if}
    {/if}
  </section>

  <section class="card">
    <h3>Channels</h3>
    <p class="muted">Connect Telegram and other messaging channels to talk to Condura from anywhere.</p>
    <div class="row">
      <button class="btn btn-ghost" onclick={goToChannels}>Manage channels</button>
    </div>
  </section>

  <section class="card">
    <h3>Voice</h3>
    <p class="muted">Talk to Condura hands-free with a wake word. Voice runs entirely on this machine.</p>
    <label class="checkbox">
      <input
        type="checkbox"
        checked={wake.enabled}
        onchange={(e) => { wake.enabled = (e.target as HTMLInputElement).checked; void saveVoice(); }}
      />
      <span>Enable wake word</span>
    </label>
    <div class="row slider-row">
      <label for="wake-sens" class="slider-label">Sensitivity</label>
      <input
        id="wake-sens"
        type="range"
        min="0" max="1" step="0.05"
        bind:value={wake.sensitivity}
        onchange={saveVoice}
        disabled={!wake.enabled}
      />
      <span class="slider-val">{Math.round(wake.sensitivity * 100)}%</span>
    </div>
    <div class="row">
      <input
        type="text"
        class="input"
        bind:value={wake.hotword}
        placeholder="hey condura"
        disabled={!wake.enabled}
      />
      <button class="btn btn-ghost" onclick={saveVoice} disabled={!wake.enabled}>Save phrase</button>
      <button class="btn btn-ghost" onclick={micTest}>Test mic</button>
    </div>
    {#if micTestResult}<p class="muted">{micTestResult}</p>{/if}
  </section>

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
    <p class="muted">Condura auto-updates by default. Disable here to opt out — but you'll need to update manually going forward.</p>
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
    <p class="muted">Encrypted archives are stored in <code>~/Documents/condura-backups</code> (or set <code>CONDURA_BACKUP_DIR</code> to override).</p>
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
            <button
              class="btn btn-ghost btn-xs"
              type="button"
              onclick={() => askRestore(b)}
              disabled={restoringBackup !== null}
              aria-label={`Restore from ${b.name}`}
            >
              {restoringBackup === b.path ? 'Restoring…' : 'Restore'}
            </button>
          </div>
        {/each}
      </div>
    {/if}
  </section>

  <section class="card">
    <h3>OS permissions</h3>
    <p class="muted">Condura needs OS permissions for accessibility, screen recording, and microphone. Grant them in System Settings if status is not granted.</p>
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

  <section class="card">
    <h3>Adaptive engine</h3>
    <p class="muted">
      Condura learns your preferences over time — your writing style, your tool habits, the kinds of
      tasks you do at what time of day. It is local: nothing about you leaves this machine.
    </p>
    {#if adaptiveError}
      <p class="error">{adaptiveError}</p>
    {/if}
    <div class="strength">
      <span class="label">Strength</span>
      <select
        value={adaptiveStrength}
        onchange={(e) => setAdaptiveStrength((e.target as HTMLSelectElement).value as typeof adaptiveStrength)}
        disabled={adaptiveLoading}
      >
        <option value="off">Off — observe nothing, apply nothing</option>
        <option value="cautious">Cautious — observe, never apply automatically</option>
        <option value="balanced">Balanced — apply safe categories (verbosity, time patterns)</option>
        <option value="aggressive">Aggressive — apply everything learned</option>
      </select>
    </div>
    {#if adaptiveProfile}
      <h4 class="sub">What Condura has learned about you</h4>
      {#if adaptiveProfile.preferences && Object.keys(adaptiveProfile.preferences).length > 0}
        <div class="profile-group">
          <span class="profile-label">Preferences</span>
          {#each Object.entries(adaptiveProfile.preferences) as [k, v]}
            <div class="profile-row">
              <span class="k">{k}</span>
              <span class="v">{v}</span>
              <button class="btn btn-ghost btn-xs" onclick={() => forgetAdaptiveField('preferences', k)}>Forget</button>
            </div>
          {/each}
        </div>
      {/if}
      {#if adaptiveProfile.style && Object.keys(adaptiveProfile.style).length > 0}
        <div class="profile-group">
          <span class="profile-label">Style</span>
          {#each Object.entries(adaptiveProfile.style) as [k, v]}
            <div class="profile-row">
              <span class="k">{k}</span>
              <span class="v">{v}</span>
              <button class="btn btn-ghost btn-xs" onclick={() => forgetAdaptiveField('style', k)}>Forget</button>
            </div>
          {/each}
        </div>
      {/if}
      {#if adaptiveProfile.pet_peeves && adaptiveProfile.pet_peeves.length > 0}
        <div class="profile-group">
          <span class="profile-label">Pet peeves</span>
          {#each adaptiveProfile.pet_peeves as p}
            <div class="profile-row">
              <span class="v">{p}</span>
              <button class="btn btn-ghost btn-xs" onclick={() => forgetAdaptiveField('pet_peeves', p)}>Forget</button>
            </div>
          {/each}
        </div>
      {/if}
      {#if (!adaptiveProfile.preferences || Object.keys(adaptiveProfile.preferences).length === 0) &&
        (!adaptiveProfile.style || Object.keys(adaptiveProfile.style).length === 0) &&
        (!adaptiveProfile.pet_peeves || adaptiveProfile.pet_peeves.length === 0)}
        <p class="muted">
          Nothing learned yet. Use Condura for a while — preferences will appear here as the
          dialectic settles on confident inferences.
        </p>
      {/if}
      {#if adaptiveProfile.last_updated}
        <p class="muted small">Last updated: {new Date(adaptiveProfile.last_updated).toLocaleString()}</p>
      {/if}
      <button class="btn btn-ghost" onclick={resetAdaptive}>Reset everything</button>
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

{#if showSignIn}
  <SignInPanel onClose={() => (showSignIn = false)} />
{/if}

{#if restoreTarget}
  <div
    class="modal-backdrop"
    onclick={(e) => {
      if (e.target === e.currentTarget) cancelRestore()
    }}
    onkeydown={(e) => {
      if (e.key === 'Escape') cancelRestore()
    }}
    role="presentation"
  >
    <div class="modal danger" role="dialog" aria-modal="true" aria-labelledby="restore-title">
      <h3 id="restore-title">Restore from backup?</h3>
      <p class="muted">
        This will replace all current data — conversations, memory, skills, settings, audit log — with the contents of
        <code>{restoreTarget.name}</code>.
      </p>
      <p class="muted">
        A pre-restore safety snapshot is taken automatically, so you can recover if the restored data is wrong.
      </p>
      <div class="modal-actions">
        <button class="btn btn-ghost" type="button" onclick={cancelRestore}>Cancel</button>
        <button class="btn btn-danger" type="button" onclick={() => void confirmRestore()}>
          Replace all data and restart
        </button>
      </div>
    </div>
  </div>
{/if}

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
    background-clip: text;
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
  /* Account / Voice (Phase 14B/14H) */
  .account-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin-top: var(--space-2);
  }
  .acc-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    flex-shrink: 0;
    object-fit: cover;
  }
  .acc-avatar.fallback {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--color-accent-gradient);
    color: white;
    font-weight: 600;
  }
  .acc-info { display: flex; flex-direction: column; flex: 1; min-width: 0; }
  .acc-name { font-weight: 600; }
  .acc-meta { color: var(--color-text-muted); font-size: var(--size-xs); }
  .benefits {
    margin: var(--space-3) 0;
    padding-left: var(--space-5);
    color: var(--color-text-muted);
    font-size: var(--size-sm);
    line-height: 1.7;
  }
  .err { color: var(--color-error, #f87171); }
  .slider-row { align-items: center; }
  .slider-label { color: var(--color-text-muted); font-size: var(--size-sm); min-width: 80px; }
  .slider-row input[type='range'] { flex: 1; accent-color: var(--color-accent); }
  .slider-val { font-family: var(--font-mono); font-size: var(--size-sm); min-width: 44px; text-align: right; }
  .strength { display: flex; align-items: center; gap: var(--space-3); margin: var(--space-2) 0 var(--space-4); }
  .strength .label { color: var(--color-text-muted); font-size: var(--size-sm); min-width: 80px; }
  .strength select { flex: 1; padding: 6px 10px; background: var(--color-bg-elev, rgba(255,255,255,0.04)); color: var(--color-text); border: 1px solid var(--glass-border); border-radius: var(--radius-md, 6px); }
  .sub { margin-top: var(--space-3); font-size: var(--size-sm); font-weight: 600; }
  .profile-group { margin: var(--space-2) 0; }
  .profile-label { display: block; color: var(--color-text-muted); font-size: var(--size-xs); text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 4px; }
  .profile-row { display: flex; align-items: center; gap: var(--space-2); padding: 4px 0; font-size: var(--size-sm); }
  .profile-row .k { color: var(--color-text-muted); min-width: 100px; }
  .profile-row .v { flex: 1; }
  .btn-xs { padding: 2px 8px; font-size: var(--size-xs); }
  .small { font-size: var(--size-xs); }
  .error { color: var(--color-error, #f87171); margin: var(--space-2) 0; }
  /* Restore confirmation modal */
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
  }
  .modal {
    background: var(--color-bg);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-xl);
    padding: var(--space-5);
    max-width: 480px;
    width: calc(100% - 32px);
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.6);
  }
  .modal.danger {
    border-color: rgba(239, 68, 68, 0.4);
  }
  .modal h3 {
    font-size: var(--size-lg);
    font-weight: 600;
    margin-bottom: var(--space-3);
  }
  .modal code {
    font-family: var(--font-mono);
    background: rgba(0, 0, 0, 0.3);
    padding: 2px 6px;
    border-radius: var(--radius-sm, 4px);
    font-size: var(--size-sm);
  }
  .modal-actions {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
    margin-top: var(--space-4);
  }
</style>
