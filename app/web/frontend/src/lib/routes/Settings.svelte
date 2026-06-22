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
  import { t } from '../i18n'

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
      alert($t('settings.voice.save_error', err))
    }
  }

  async function micTest(): Promise<void> {
    micTestResult = $t('settings.voice.mic_checking')
    try {
      const perms = await ipc.permissionsStatus()
      const mic = perms.find((p) => p.kind === 'microphone')
      if (!mic) micTestResult = $t('settings.voice.mic_unavailable')
      else if (mic.status === 'granted') micTestResult = $t('settings.voice.mic_granted')
      else if (mic.status === 'denied') micTestResult = $t('settings.voice.mic_denied')
      else micTestResult = $t('settings.voice.mic_not_granted')
    } catch (err) {
      micTestResult = $t('settings.voice.mic_test_failed', err)
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
    const confirmMsg = field === 'pet_peeves'
      ? $t('settings.adaptive.forget_dislike', value)
      : $t('settings.adaptive.forget_have', value)
    if (!confirm(confirmMsg)) return
    try {
      await ipc.call('adaptive.forget', { field, value })
      void loadAdaptive()
    } catch (e) {
      adaptiveError = String(e)
    }
  }

  async function resetAdaptive(): Promise<void> {
    if (!confirm($t('settings.adaptive.reset_confirm'))) return
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
      eulaTitle = $t('onboarding.eula.title')
      eulaVersion = doc.version
    } catch (err) {
      alert($t('settings.legal.eula_error', err))
    }
  }

  async function rerunSetup(): Promise<void> {
    if (!confirm($t('settings.setup.rerun_confirm'))) return
    rerunning = true
    try {
      await onboarding.reset()
      window.dispatchEvent(new CustomEvent('synaptic:show-onboarding'))
    } catch (err) {
      alert($t('settings.setup.rerun_error', err))
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
    alert($t('settings.hotkey.saved_alert'))
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
      alert($t('settings.apikeys.saved_alert'))
    } catch (err) {
      alert($t('settings.apikeys.failed_alert', err))
    } finally {
      settingAPIKey = false
    }
  }

  async function deleteKey(id: number): Promise<void> {
    if (!confirm($t('settings.apikeys.delete_confirm'))) return
    await apiKeys.remove(id)
  }

  async function performHalt(): Promise<void> {
    if (!confirm($t('settings.killswitch.halt_confirm'))) return
    await halt.halt('user requested from settings')
  }

  async function performResume(): Promise<void> {
    await halt.resume()
  }

  async function createBackup(): Promise<void> {
    creatingBackup = true
    try {
      const path = await trust.createBackup()
      alert($t('settings.backup.created_alert', path))
    } catch (err) {
      alert($t('settings.backup.failed_alert', err))
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
      alert($t('settings.backup.restored_alert', target.name))
    } catch (err) {
      alert($t('settings.backup.restore_failed_alert', err))
    } finally {
      restoringBackup = null
    }
  }

  async function showPermissionGuide(kind: string): Promise<void> {
    try {
      const g = await trust.loadGuide(kind)
      permissionGuide = { kind: g.kind, title: g.title, steps: g.steps }
    } catch (err) {
      alert($t('settings.permissions.guide_error', err))
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
    <h2>{$t('settings.title')}</h2>
    <p class="muted">{$t('settings.subtitle')}</p>
  </header>

  <section class="card">
    <h3>{$t('settings.account.title')}</h3>
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
          {account.loading ? $t('settings.account.signing_out') : $t('settings.account.signout')}
        </button>
      </div>
    {:else}
      <p class="muted">{$t('settings.account.signed_out_intro')}</p>
      <ul class="benefits">
        <li>{$t('settings.account.benefit_1')}</li>
        <li>{$t('settings.account.benefit_2')}</li>
        <li>{$t('settings.account.benefit_3')}</li>
      </ul>
      <div class="row">
        <button class="btn btn-primary" onclick={() => (showSignIn = true)}>{$t('settings.account.signin')}</button>
      </div>
      {#if account.error}<p class="muted err">{account.error}</p>{/if}
    {/if}
  </section>

  <section class="card">
    <h3>{$t('settings.channels.title')}</h3>
    <p class="muted">{$t('settings.channels.intro')}</p>
    <div class="row">
      <button class="btn btn-ghost" onclick={goToChannels}>{$t('settings.channels.manage')}</button>
    </div>
  </section>

  <section class="card">
    <h3>{$t('settings.voice.title')}</h3>
    <p class="muted">{$t('settings.voice.intro')}</p>
    <label class="checkbox">
      <input
        type="checkbox"
        checked={wake.enabled}
        onchange={(e) => { wake.enabled = (e.target as HTMLInputElement).checked; void saveVoice(); }}
      />
      <span>{$t('settings.voice.enable_wake')}</span>
    </label>
    <div class="row slider-row">
      <label for="wake-sens" class="slider-label">{$t('settings.voice.sensitivity')}</label>
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
      <button class="btn btn-ghost" onclick={saveVoice} disabled={!wake.enabled}>{$t('settings.voice.save_phrase')}</button>
      <button class="btn btn-ghost" onclick={micTest}>{$t('settings.voice.test_mic')}</button>
    </div>
    {#if micTestResult}<p class="muted">{micTestResult}</p>{/if}
  </section>

  <section class="card">
    <h3>{$t('settings.language.title')}</h3>
    <p class="muted">{$t('settings.language.intro')}</p>
    <div class="row">
      <LocaleSelector />
    </div>
  </section>

  <section class="card">
    <h3>{$t('settings.spend.title')}</h3>
    {#if spend.summary}
      <div class="kv">
        <span class="k">{$t('settings.spend.spent_today')}</span><span class="v">${spend.summary.spent.toFixed(2)}</span>
      </div>
      <div class="kv">
        <span class="k">{$t('settings.spend.cap')}</span><span class="v">${spend.summary.cap.toFixed(2)}</span>
      </div>
      <div class="kv">
        <span class="k">{$t('settings.spend.remaining')}</span><span class="v">${spend.summary.remaining.toFixed(2)}</span>
      </div>
    {:else}
      <p class="muted">{$t('common.loading')}</p>
    {/if}
  </section>

  <section class="card">
    <h3>{$t('settings.hotkey.title')}</h3>
    <p class="muted">{$t('settings.hotkey.intro')}</p>
    <div class="row">
      <input
        type="text"
        bind:value={hotkeyInput}
        placeholder="Cmd+Shift+Space"
        class="input"
      />
      <button class="btn btn-primary" onclick={saveHotkey}>{$t('settings.hotkey.save')}</button>
    </div>
  </section>

  <section class="card">
    <h3>{$t('settings.update.title')}</h3>
    <p class="muted">{$t('settings.update.intro')}</p>
    <label class="checkbox">
      <input
        type="checkbox"
        checked={telemetryInput}
        onchange={(e) => { telemetryInput = (e.target as HTMLInputElement).checked; void saveTelemetry(); }}
      />
      <span>{$t('settings.update.enable')}</span>
    </label>
    {#if updateStore.lastCheck}
      <p class="muted">{$t('settings.update.last_checked', new Date(updateStore.lastCheck).toLocaleString())}</p>
    {/if}
  </section>

  <section class="card">
    <h3>{$t('settings.backup.title')}</h3>
    <p class="muted">{$t('settings.backup.intro')}</p>
    <div class="row">
      <button class="btn btn-primary" onclick={createBackup} disabled={creatingBackup}>
        {creatingBackup ? $t('settings.backup.creating') : $t('settings.backup.create')}
      </button>
      <button class="btn btn-ghost" onclick={() => trust.refreshBackups()}>{$t('settings.backup.refresh')}</button>
    </div>
    {#if trust.loadingBackups}
      <p class="muted">{$t('settings.backup.loading')}</p>
    {:else if trust.backups.length === 0}
      <p class="muted">{$t('settings.backup.empty')}</p>
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
              aria-label={$t('settings.backup.restore_aria', b.name)}
            >
              {restoringBackup === b.path ? $t('settings.backup.restoring') : $t('settings.backup.restore')}
            </button>
          </div>
        {/each}
      </div>
    {/if}
  </section>

  <section class="card">
    <h3>{$t('settings.permissions.title')}</h3>
    <p class="muted">{$t('settings.permissions.intro')}</p>
    <button class="btn btn-ghost" onclick={() => trust.refreshPermissions()}>{$t('settings.permissions.refresh')}</button>
    {#if trust.loadingPermissions}
      <p class="muted">{$t('settings.permissions.checking')}</p>
    {:else}
      <div class="perm-list">
        {#each trust.permissions as p (p.kind)}
          <div class="perm-row">
            <span class="perm-kind">{p.kind}</span>
            <span class="perm-status" class:granted={p.status === 'granted'} class:denied={p.status === 'denied'}>{p.status}</span>
            {#if p.status !== 'granted'}
              <button class="btn btn-ghost" onclick={() => showPermissionGuide(p.kind)}>{$t('settings.permissions.how_to_grant')}</button>
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
        <button class="btn btn-ghost" onclick={() => { permissionGuide = null }}>{$t('common.close')}</button>
      </div>
    {/if}
  </section>

  <section class="card">
    <h3>{$t('settings.adaptive.title')}</h3>
    <p class="muted">
      {$t('settings.adaptive.intro')}
    </p>
    {#if adaptiveError}
      <p class="error">{adaptiveError}</p>
    {/if}
    <div class="strength">
      <span class="label">{$t('settings.adaptive.strength')}</span>
      <select
        value={adaptiveStrength}
        onchange={(e) => setAdaptiveStrength((e.target as HTMLSelectElement).value as typeof adaptiveStrength)}
        disabled={adaptiveLoading}
      >
        <option value="off">{$t('settings.adaptive.off')}</option>
        <option value="cautious">{$t('settings.adaptive.cautious')}</option>
        <option value="balanced">{$t('settings.adaptive.balanced')}</option>
        <option value="aggressive">{$t('settings.adaptive.aggressive')}</option>
      </select>
    </div>
    {#if adaptiveProfile}
      <h4 class="sub">{$t('settings.adaptive.learned_title')}</h4>
      {#if adaptiveProfile.preferences && Object.keys(adaptiveProfile.preferences).length > 0}
        <div class="profile-group">
          <span class="profile-label">{$t('settings.adaptive.preferences')}</span>
          {#each Object.entries(adaptiveProfile.preferences) as [k, v]}
            <div class="profile-row">
              <span class="k">{k}</span>
              <span class="v">{v}</span>
              <button class="btn btn-ghost btn-xs" onclick={() => forgetAdaptiveField('preferences', k)}>{$t('settings.adaptive.forget')}</button>
            </div>
          {/each}
        </div>
      {/if}
      {#if adaptiveProfile.style && Object.keys(adaptiveProfile.style).length > 0}
        <div class="profile-group">
          <span class="profile-label">{$t('settings.adaptive.style')}</span>
          {#each Object.entries(adaptiveProfile.style) as [k, v]}
            <div class="profile-row">
              <span class="k">{k}</span>
              <span class="v">{v}</span>
              <button class="btn btn-ghost btn-xs" onclick={() => forgetAdaptiveField('style', k)}>{$t('settings.adaptive.forget')}</button>
            </div>
          {/each}
        </div>
      {/if}
      {#if adaptiveProfile.pet_peeves && adaptiveProfile.pet_peeves.length > 0}
        <div class="profile-group">
          <span class="profile-label">{$t('settings.adaptive.pet_peeves')}</span>
          {#each adaptiveProfile.pet_peeves as p}
            <div class="profile-row">
              <span class="v">{p}</span>
              <button class="btn btn-ghost btn-xs" onclick={() => forgetAdaptiveField('pet_peeves', p)}>{$t('settings.adaptive.forget')}</button>
            </div>
          {/each}
        </div>
      {/if}
      {#if (!adaptiveProfile.preferences || Object.keys(adaptiveProfile.preferences).length === 0) &&
        (!adaptiveProfile.style || Object.keys(adaptiveProfile.style).length === 0) &&
        (!adaptiveProfile.pet_peeves || adaptiveProfile.pet_peeves.length === 0)}
        <p class="muted">
          {$t('settings.adaptive.empty')}
        </p>
      {/if}
      {#if adaptiveProfile.last_updated}
        <p class="muted small">{$t('settings.adaptive.last_updated', new Date(adaptiveProfile.last_updated).toLocaleString())}</p>
      {/if}
      <button class="btn btn-ghost" onclick={resetAdaptive}>{$t('settings.adaptive.reset')}</button>
    {/if}
  </section>

  <section class="card danger">
    <h3>{$t('settings.killswitch.title')}</h3>
    <p class="muted">{$t('settings.killswitch.intro')}</p>
    {#if halt.state.halted}
      <p class="muted">{$t('settings.killswitch.halted_since', halt.state.since)}</p>
      <button class="btn btn-primary" onclick={performResume}>{$t('settings.killswitch.resume')}</button>
    {:else}
      <button class="btn btn-danger" onclick={performHalt}>{$t('settings.killswitch.halt')}</button>
    {/if}
  </section>

  <section class="card">
    <h3>{$t('settings.apikeys.title')}</h3>
    <p class="muted">{$t('settings.apikeys.intro')}</p>

    <div class="apikey-list">
      {#if apiKeys.list.length === 0}
        <p class="muted">{$t('settings.apikeys.empty')}</p>
      {/if}
      {#each apiKeys.list as k (k.id)}
        <div class="apikey-row">
          <span class="provider">{k.provider}</span>
          <span class="label">{k.label}</span>
          <span class="auth-kind">{k.auth_kind}</span>
          <span class="has-token">{k.has_token ? $t('settings.apikeys.has_token') : $t('settings.apikeys.no_token')}</span>
          <button class="btn btn-ghost" onclick={() => deleteKey(k.id)}>{$t('settings.apikeys.delete')}</button>
        </div>
      {/each}
    </div>

    <h4>{$t('settings.apikeys.add_title')}</h4>
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
        {settingAPIKey ? $t('settings.apikeys.saving') : $t('settings.apikeys.save')}
      </button>
    </div>
  </section>

  <section class="card">
    <h3>{$t('settings.legal.title')}</h3>
    <p class="muted">{$t('settings.legal.intro')}</p>
    {#if eulaText}
      <div class="eula-view">
        <div class="eula-view-head">
          <strong>{eulaTitle}</strong>
          <span class="muted">{eulaVersion}</span>
        </div>
        <pre>{eulaText}</pre>
      </div>
      <button class="btn btn-ghost" onclick={() => { eulaText = '' }}>{$t('settings.legal.hide')}</button>
    {:else}
      <button class="btn btn-ghost" onclick={viewEula}>{$t('settings.legal.view_eula')}</button>
    {/if}
  </section>

  <section class="card">
    <h3>{$t('settings.setup.title')}</h3>
    <p class="muted">{$t('settings.setup.intro')}</p>
    <button class="btn btn-ghost" onclick={rerunSetup} disabled={rerunning}>
      {rerunning ? $t('settings.setup.resetting') : $t('settings.setup.rerun')}
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
      <h3 id="restore-title">{$t('settings.backup.restore_title')}</h3>
      <p class="muted">
        {$t('settings.backup.restore_warning', restoreTarget.name)}
      </p>
      <p class="muted">
        {$t('settings.backup.restore_safety')}
      </p>
      <div class="modal-actions">
        <button class="btn btn-ghost" type="button" onclick={cancelRestore}>{$t('common.cancel')}</button>
        <button class="btn btn-danger" type="button" onclick={() => void confirmRestore()}>
          {$t('settings.backup.replace_all')}
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
