<script lang="ts">
  /**
   * Settings — full Meridian instrument desk.
   * Lighting · OS permissions · hotkey · autonomy matrix · spend/privacy ·
   * providers · adaptive · backup · trust capabilities · keys.
   *
   * Autonomy levels match the daemon: supervised | warn | autonomous | block.
   */
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import MeridianSettingsPerms from './MeridianSettingsPerms.svelte'
  import { settings } from '../../stores/settings.svelte'
  import { trust } from '../../stores/trust.svelte'
  import { spend as spendStore } from '../../stores/spend.svelte'
  import { ipc } from '../../ipc/client'
  import type {
    AdaptiveStrength,
    AdaptiveUserModel,
    APIKeyMeta,
    DaemonCapabilities,
    InferredField,
    ProviderInfo,
  } from '../../ipc/types'
  import {
    getResolvedTheme,
    onThemeChange,
    setResolvedTheme,
    type ResolvedTheme,
  } from '../../theme/condura-theme'

  type AutonomyDefault = 'supervised' | 'warn' | 'autonomous'
  type AutonomyTask = 'autonomous' | 'warn' | 'block'

  let theme = $state<ResolvedTheme>(getResolvedTheme())
  let saving = $state(false)
  let note = $state('')
  let modLabel = $state('⌘')
  let noteTimer: ReturnType<typeof setTimeout> | null = null

  // Hotkey
  let hotkeyCombo = $state('')
  let recordingHotkey = $state(false)
  let wakeEnabled = $state(false)

  // Adaptive
  let strength = $state<AdaptiveStrength>('balanced')
  let adaptiveProfile = $state<AdaptiveUserModel | null>(null)
  let adaptiveLoading = $state(false)
  let adaptiveBusy = $state(false)
  let confirmResetAdaptive = $state(false)

  // Backup
  let backupBusy = $state(false)
  let restoreBusy = $state(false)
  let confirmRestorePath = $state<string | null>(null)
  let backupNote = $state('')

  // Trust
  let capabilities = $state<DaemonCapabilities | null>(null)
  let capsError = $state('')

  // Providers
  let providers = $state<ProviderInfo[]>([])
  let apiKeys = $state<APIKeyMeta[]>([])
  let keyDraft = $state('')
  let keyProvider = $state('')
  let keyBusy = $state(false)
  let keyNote = $state('')

  // Re-run first-run wizard (MeridianShell listens + onboarding.reset)
  let confirmRerunSetup = $state(false)
  let rerunBusy = $state(false)

  const AUTONOMY_DEFAULT: {
    id: AutonomyDefault
    title: string
    body: string
    tone: string
  }[] = [
    {
      id: 'supervised',
      title: 'Ask first',
      body: 'Every gated action waits for you. Safest default.',
      tone: 'cobalt',
    },
    {
      id: 'warn',
      title: 'Warn',
      body: 'Condura proposes; you still open the door.',
      tone: 'live',
    },
    {
      id: 'autonomous',
      title: 'Autonomous',
      body: 'Routine work may proceed — still Gatekeeper-gated.',
      tone: 'caution',
    },
  ]

  const TASK_TYPES: { key: string; label: string }[] = [
    { key: 'coding', label: 'Coding' },
    { key: 'file_operations', label: 'File operations' },
    { key: 'web_browsing', label: 'Web browsing' },
    { key: 'email', label: 'Email' },
    { key: 'calendar', label: 'Calendar' },
    { key: 'messaging', label: 'Messaging' },
    { key: 'shell_commands', label: 'Shell commands' },
    { key: 'computer_use', label: 'Computer use' },
    { key: 'research', label: 'Research' },
    { key: 'image_generation', label: 'Image generation' },
    { key: 'code_review', label: 'Code review' },
  ]

  const STRENGTH_OPTS: AdaptiveStrength[] = ['off', 'cautious', 'balanced', 'aggressive']

  const KEYS = [
    { keys: 'K', label: 'Jump anywhere' },
    { keys: 'Shift+T', label: 'Toggle light / dark (not while typing)' },
    { keys: 'Esc', label: 'Close overlays' },
    { keys: 'Halt', label: 'Cut the line in the dock · or ⌘K' },
  ]

  /** Cap from config; live spent from spend store (poll + spend_warning SSE). */
  const spendCap = $derived(
    Number(
      settings.config?.security?.spend_limit_usd_per_day ??
        spendStore.cap ??
        0
    )
  )
  const spentToday = $derived(spendStore.spent)
  const autonomy = $derived(
    normalizeDefault(settings.config?.autonomy?.default_level)
  )
  const spendPct = $derived(
    spendCap > 0 ? Math.min(100, Math.round((spentToday / spendCap) * 100)) : 0
  )
  const spendHot = $derived(spendCap > 0 && spendPct >= 80)
  const offline = $derived(!settings.config)
  const telemetryOn = $derived(!!settings.config?.telemetry?.enabled)
  const backupDir = $derived(
    settings.config?.storage?.backup?.dir || '~/Documents/condura-backups/'
  )
  const hotkeyDisplay = $derived(formatHotkeyDisplay(hotkeyCombo || settings.config?.hotkey?.overlay || ''))
  const grantedPerms = $derived(trust.permissions.filter((p) => p.status === 'granted').length)
  const keyedProviders = $derived(apiKeys.filter((k) => k.has_token).length)

  const autonomyTasks = $derived(
    TASK_TYPES.map((t) => ({
      key: t.key,
      label: t.label,
      level: taskLevel(t.key),
    }))
  )

  const adaptiveItems = $derived(rowsFromProfile(adaptiveProfile))

  const liveNote = $derived(
    saving
      ? 'Writing defaults to the daemon…'
      : note.includes('offline')
        ? note
        : note === 'Saved'
          ? 'Defaults saved — Gatekeeper still holds the door'
          : offline
            ? 'Daemon offline — lighting still works; other instruments need a connection'
            : `OS grants ${grantedPerms}/5 · autonomy ${autonomy} · spend $${spentToday.toFixed(2)}/$${spendCap} · ${keyedProviders} keyed provider${keyedProviders === 1 ? '' : 's'}`
  )

  onMount(() => {
    theme = getResolvedTheme()
    modLabel = /Mac|iPhone|iPad/.test(navigator.platform) ? '⌘' : 'Ctrl'
    const off = onThemeChange((r) => {
      theme = r
    })
    void bootstrap().then(() => scrollToHashSection())
    const onHash = () => scrollToHashSection()
    window.addEventListener('hashchange', onHash)
    return () => {
      off()
      window.removeEventListener('hashchange', onHash)
      if (noteTimer) clearTimeout(noteTimer)
    }
  })

  async function bootstrap(): Promise<void> {
    await Promise.allSettled([
      settings.refresh(),
      trust.refreshBackups(),
      loadAdaptive(),
      loadCapabilities(),
      loadProviders(),
      loadVoice(),
      spendStore.refresh(),
    ])
    hotkeyCombo = settings.config?.hotkey?.overlay ?? ''
    // Ensure SSE warnings are live even if init order skipped polling start.
    spendStore.startLive()
  }

  /** Ask "Configure model" lands on #/settings/models → scroll providers plate. */
  function scrollToHashSection(): void {
    const h = window.location.hash || ''
    queueMicrotask(() => {
      if (/spend|cap/i.test(h)) {
        const el = document.getElementById('md-spend')
        if (!el) return
        el.scrollIntoView({ behavior: 'smooth', block: 'start' })
        const input = el.querySelector<HTMLInputElement>('input[type="number"]')
        input?.focus({ preventScroll: true })
        return
      }
      if (!/models|providers|keys/i.test(h)) return
      const el = document.getElementById('md-models')
      if (!el) return
      el.scrollIntoView({ behavior: 'smooth', block: 'start' })
      const keyInput = el.querySelector<HTMLInputElement>('input[type="password"]')
      keyInput?.focus({ preventScroll: true })
    })
  }

  function normalizeDefault(raw: string | undefined): AutonomyDefault {
    if (raw === 'supervised' || raw === 'ask') return 'supervised'
    if (raw === 'autonomous' || raw === 'auto') return 'autonomous'
    if (raw === 'suggest') return 'warn'
    return 'warn'
  }

  function taskLevel(key: string): AutonomyTask {
    const level =
      settings.config?.autonomy?.per_task?.[key] ??
      settings.config?.autonomy?.default_level ??
      'warn'
    if (level === 'autonomous' || level === 'auto') return 'autonomous'
    if (level === 'block') return 'block'
    return 'warn'
  }

  function rowsFromProfile(model: AdaptiveUserModel | null) {
    if (!model) return [] as { field: string; value: string; evidence: string; confidence: number }[]
    const rows: { field: string; value: string; evidence: string; confidence: number }[] = []
    const push = (field: string, item: InferredField | undefined) => {
      if (!item?.value) return
      rows.push({
        field,
        value: item.value,
        evidence: (item.evidence ?? []).join(' · ') || '—',
        confidence: item.confidence ?? 0,
      })
    }
    const pushMany = (field: string, items: InferredField[] | undefined) => {
      for (const item of items ?? []) push(field, item)
    }
    pushMany('preferences', model.preferences)
    push('style', model.style)
    pushMany('expertise', model.expertise)
    pushMany('pet_peeves', model.pet_peeves)
    push('communication', model.communication)
    push('risk_tolerance', model.risk_tolerance)
    return rows
  }

  function formatHotkeyDisplay(spec: string): string {
    return spec
      .replace(/Cmd\+/gi, '⌘')
      .replace(/Ctrl\+/gi, '^')
      .replace(/Option\+/gi, '⌥')
      .replace(/Alt\+/gi, '⌥')
      .replace(/Shift\+/gi, '⇧')
      .replace(/\+/g, '')
  }

  function daemonHotkeyFromDisplay(display: string): string {
    let spec = display
    for (const [from, to] of [
      ['⌘', 'Cmd+'],
      ['^', 'Ctrl+'],
      ['⌥', 'Option+'],
      ['⇧', 'Shift+'],
    ] as const) {
      spec = spec.split(from).join(to)
    }
    if (spec.endsWith('+')) spec = spec.slice(0, -1)
    return spec
  }

  function formatBytes(n: number): string {
    if (n < 1024) return `${n} B`
    if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
    return `${(n / (1024 * 1024)).toFixed(1)} MB`
  }

  async function savePatch(patch: Record<string, unknown>): Promise<void> {
    saving = true
    note = ''
    if (noteTimer) clearTimeout(noteTimer)
    try {
      await settings.save(patch as never)
      note = 'Saved'
      noteTimer = setTimeout(() => {
        if (note === 'Saved') note = ''
      }, 2200)
    } catch (e) {
      const s = String(e)
      note = /IPC client not started|not connected|Failed to fetch/i.test(s)
        ? 'Daemon offline — connect to save'
        : s
    } finally {
      saving = false
    }
  }

  function setTheme(t: ResolvedTheme): void {
    theme = setResolvedTheme(t)
  }

  function setAutonomy(level: AutonomyDefault): void {
    if (!settings.config) return
    void savePatch({
      autonomy: {
        ...settings.config.autonomy,
        default_level: level,
      },
    })
  }

  async function setTaskAutonomy(key: string, level: AutonomyTask): Promise<void> {
    if (!settings.config) return
    const autonomyCfg = settings.config.autonomy ?? {
      default_level: 'warn',
      per_app: {},
      per_task: {},
    }
    await savePatch({
      autonomy: {
        ...autonomyCfg,
        per_task: { ...autonomyCfg.per_task, [key]: level },
      },
    })
  }

  function go(hash: string): void {
    window.location.hash = hash
  }

  function onHotkeyKey(e: KeyboardEvent): void {
    if (!recordingHotkey) return
    e.preventDefault()
    e.stopPropagation()
    if (e.key === 'Escape') {
      recordingHotkey = false
      return
    }
    const parts: string[] = []
    if (e.metaKey) parts.push('⌘')
    if (e.ctrlKey) parts.push('^')
    if (e.altKey) parts.push('⌥')
    if (e.shiftKey) parts.push('⇧')
    const key = e.key
    if (key === ' ') parts.push('Space')
    else if (['Meta', 'Control', 'Alt', 'Shift'].includes(key)) return
    else if (key.length === 1) parts.push(key.toUpperCase())
    else parts.push(key)
    const combo = parts.join('')
    if (combo.length > 1 && (combo.includes('⌘') || combo.includes('^') || combo.includes('⌥') || combo.includes('⇧'))) {
      recordingHotkey = false
      void saveHotkey(combo)
    }
  }

  async function saveHotkey(combo: string): Promise<void> {
    const spec = combo.includes('+') ? combo : daemonHotkeyFromDisplay(combo)
    if (!spec) return
    hotkeyCombo = spec
    await savePatch({
      hotkey: { ...(settings.config?.hotkey ?? { overlay: '' }), overlay: spec },
    })
  }

  async function setWakeEnabled(enabled: boolean): Promise<void> {
    const prev = wakeEnabled
    wakeEnabled = enabled
    try {
      // Deep-merge path via settings.save (not raw ipc) so local mirror
      // and daemon publicConfigView stay aligned.
      await settings.save({
        voice: {
          ...(settings.config?.voice ?? {}),
          wake: {
            ...(settings.config?.voice?.wake ?? {}),
            enabled,
          },
        },
      })
    } catch {
      wakeEnabled = prev
    }
  }

  async function setTelemetry(enabled: boolean): Promise<void> {
    if (!settings.config) return
    await savePatch({
      telemetry: { ...settings.config.telemetry, enabled },
    })
  }

  async function loadAdaptive(): Promise<void> {
    adaptiveLoading = true
    try {
      const [profile, strengthRes] = await Promise.all([
        ipc.adaptiveProfile(),
        ipc.adaptiveStrengthGet(),
      ])
      adaptiveProfile = profile
      strength = strengthRes.strength
    } catch {
      adaptiveProfile = null
    } finally {
      adaptiveLoading = false
    }
  }

  async function setStrength(next: AdaptiveStrength): Promise<void> {
    const prev = strength
    strength = next
    try {
      await ipc.adaptiveStrengthSet(next)
    } catch {
      strength = prev
    }
  }

  async function forgetAdaptive(field: string, value: string): Promise<void> {
    adaptiveBusy = true
    try {
      await ipc.adaptiveForget(field, value)
      await loadAdaptive()
    } finally {
      adaptiveBusy = false
    }
  }

  async function resetAdaptive(): Promise<void> {
    adaptiveBusy = true
    confirmResetAdaptive = false
    try {
      await ipc.adaptiveReset()
      await loadAdaptive()
    } finally {
      adaptiveBusy = false
    }
  }

  function exportAdaptive(): void {
    if (!adaptiveProfile) return
    const blob = new Blob([JSON.stringify(adaptiveProfile, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'condura-adaptive-profile.json'
    a.click()
    URL.revokeObjectURL(url)
  }

  async function loadCapabilities(): Promise<void> {
    capsError = ''
    try {
      capabilities = await ipc.daemonCapabilities()
    } catch (e) {
      capabilities = null
      capsError = String(e)
    }
  }

  async function loadProviders(): Promise<void> {
    try {
      const [p, k] = await Promise.all([ipc.providersList(), ipc.apiKeysList()])
      providers = p
      apiKeys = k
      if (!keyProvider && p[0]) keyProvider = p[0].name
    } catch {
      providers = []
      apiKeys = []
    }
  }

  async function loadVoice(): Promise<void> {
    try {
      const probe = await ipc.onboardingProbeVoice()
      wakeEnabled = probe.wake_word_enabled
    } catch {
      /* optional */
    }
  }

  /** Walk welcome flow again — data preserved; daemon step reset in shell. */
  function requestRerunSetup(): void {
    confirmRerunSetup = true
  }

  function cancelRerunSetup(): void {
    if (rerunBusy) return
    confirmRerunSetup = false
  }

  async function confirmAndRerunSetup(): Promise<void> {
    rerunBusy = true
    try {
      window.dispatchEvent(new CustomEvent('condura:show-onboarding'))
      confirmRerunSetup = false
    } finally {
      rerunBusy = false
    }
  }

  function providerEnabled(name: string): boolean {
    return !!settings.config?.llm?.providers?.[name]?.enabled
  }

  function providerHasKey(name: string): boolean {
    return apiKeys.some((k) => k.provider === name && k.has_token)
  }

  async function toggleProvider(name: string, enabled: boolean): Promise<void> {
    if (!settings.config) return
    const providersCfg = { ...(settings.config.llm?.providers ?? {}) }
    const cur = providersCfg[name] ?? {
      enabled: false,
      api_key: '',
      base_url: '',
      default_model: '',
    }
    providersCfg[name] = { ...cur, enabled }
    await savePatch({
      llm: { ...settings.config.llm, providers: providersCfg },
    })
  }

  async function saveApiKey(): Promise<void> {
    if (!keyProvider || !keyDraft.trim()) return
    keyBusy = true
    keyNote = ''
    try {
      await ipc.apiKeysSet({
        provider: keyProvider,
        label: 'default',
        secret: keyDraft.trim(),
      })
      keyDraft = ''
      keyNote = `Key saved for ${keyProvider}`
      await loadProviders()
    } catch (e) {
      keyNote = String(e)
    } finally {
      keyBusy = false
    }
  }

  async function deleteApiKey(id: number): Promise<void> {
    keyBusy = true
    try {
      await ipc.apiKeysDelete(id)
      await loadProviders()
      keyNote = 'Key removed'
    } catch (e) {
      keyNote = String(e)
    } finally {
      keyBusy = false
    }
  }

  async function createBackup(): Promise<void> {
    backupBusy = true
    backupNote = ''
    try {
      const path = await trust.createBackup()
      backupNote = `Backup written · ${path}`
      await trust.refreshBackups()
    } catch (e) {
      backupNote = String(e)
    } finally {
      backupBusy = false
    }
  }

  async function performRestore(path: string): Promise<void> {
    restoreBusy = true
    backupNote = ''
    confirmRestorePath = null
    try {
      await ipc.backupRestore({ path })
      await Promise.allSettled([settings.refresh(), loadAdaptive(), loadProviders()])
      backupNote = 'Restore complete — desk refreshed'
    } catch (e) {
      backupNote = String(e)
    } finally {
      restoreBusy = false
    }
  }
</script>

<svelte:window onkeydown={onHotkeyKey} />

<MeridianPage
  kicker="Desk · instruments"
  title="Settings"
  lead="Lighting, OS permissions, hotkey, autonomy, spend, providers, memory, backup, and trust — every instrument Condura needs on one desk."
>
  <div class="desk md-stagger">
    <p class="contract" class:hot={note === 'Saved'} class:off={offline || note.includes('offline')}>
      <span class="live-dot" aria-hidden="true"></span>
      {liveNote}.
    </p>

    <ol class="pipe" aria-label="Instrument map">
      <li><span class="n">01</span><span class="t">Lighting</span></li>
      <li><span class="n">02</span><span class="t">Permissions</span></li>
      <li><span class="n">03</span><span class="t">Hotkey</span></li>
      <li><span class="n">04</span><span class="t">Autonomy</span></li>
      <li><span class="n">05</span><span class="t">Spend</span></li>
      <li><span class="n">06</span><span class="t">Models</span></li>
      <li><span class="n">07</span><span class="t">Adaptive</span></li>
      <li><span class="n">08</span><span class="t">Backup</span></li>
      <li><span class="n">09</span><span class="t">Trust</span></li>
      <li><span class="n">10</span><span class="t">Keys</span></li>
    </ol>

    <!-- 01 Lighting -->
    <section class="plate lighting" data-mode={theme}>
      <header class="plate-head">
        <p class="cite">01 · appearance</p>
        <h2>Desk lighting</h2>
        <p class="hint">Shift+T toggles anytime. The mist follows — no daemon required.</p>
      </header>
      <div class="seg" role="group" aria-label="Theme">
        <button type="button" class:on={theme === 'light'} onclick={() => setTheme('light')}>Light</button>
        <button type="button" class:on={theme === 'dark'} onclick={() => setTheme('dark')}>Dark</button>
      </div>
      <div class="swatch" aria-hidden="true">
        <span class="sw mist"></span>
        <span class="sw stage"></span>
        <span class="sw cobalt"></span>
        <span class="sw live"></span>
      </div>
    </section>

    <!-- 02 Permissions -->
    <MeridianSettingsPerms />

    <!-- 03 Hotkey -->
    <section class="plate">
      <header class="plate-head">
        <p class="cite">03 · wake</p>
        <h2>Hotkey</h2>
        <p class="hint">Press a combo to record. Condura appears when you call it.</p>
      </header>
      <div class="hotkey-stage">
        <div class="hotkey-display">{hotkeyDisplay || 'Not set'}</div>
        <button
          type="button"
          class="md-btn md-btn-primary"
          class:recording={recordingHotkey}
          onclick={() => (recordingHotkey = !recordingHotkey)}
        >
          {recordingHotkey ? 'Listening… Esc to cancel' : 'Record new combo'}
        </button>
      </div>
      <label class="toggle-row">
        <input
          type="checkbox"
          checked={wakeEnabled}
          onchange={(e) => void setWakeEnabled((e.currentTarget as HTMLInputElement).checked)}
        />
        <span>
          <strong>Also say “hey condura”</strong>
          <em>Local wake word on this machine.</em>
        </span>
      </label>
    </section>

    <!-- 04 Autonomy -->
    <section class="plate gate">
      <header class="plate-head">
        <p class="cite">04 · gate defaults</p>
        <h2>How Condura asks</h2>
        <p class="hint">Autonomy never bypasses the Gatekeeper. Defaults and per-task dials use daemon levels: supervised, warn, autonomous, block.</p>
      </header>

      {#if !settings.config}
        <p class="muted">Config unavailable offline. Connect the daemon to edit defaults.</p>
      {:else}
        <div class="autonomy" role="radiogroup" aria-label="Default autonomy">
          {#each AUTONOMY_DEFAULT as a (a.id)}
            <button
              type="button"
              class="auto-card"
              class:on={autonomy === a.id}
              data-tone={a.tone}
              role="radio"
              aria-checked={autonomy === a.id}
              disabled={saving}
              onclick={() => setAutonomy(a.id)}
            >
              <span class="auto-dot" aria-hidden="true"></span>
              <strong>{a.title}</strong>
              <span>{a.body}</span>
              <span class="auto-id">{a.id}</span>
            </button>
          {/each}
        </div>

        <div class="matrix">
          <p class="matrix-k">Per-task matrix</p>
          {#each autonomyTasks as t (t.key)}
            <div class="matrix-row">
              <span class="matrix-label">{t.label}</span>
              <div class="dial" role="group" aria-label="{t.label} autonomy">
                <button
                  type="button"
                  class:on={t.level === 'autonomous'}
                  data-l="auto"
                  onclick={() => void setTaskAutonomy(t.key, 'autonomous')}
                >auto</button>
                <button
                  type="button"
                  class:on={t.level === 'warn'}
                  data-l="warn"
                  onclick={() => void setTaskAutonomy(t.key, 'warn')}
                >warn</button>
                <button
                  type="button"
                  class:on={t.level === 'block'}
                  data-l="block"
                  onclick={() => void setTaskAutonomy(t.key, 'block')}
                >block</button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </section>

    <!-- 05 Spend & privacy -->
    <section class="plate" id="md-spend">
      <header class="plate-head">
        <p class="cite">05 · spend · privacy</p>
        <h2>Ceiling and silence</h2>
        <p class="hint">Hard daily spend cap. Telemetry is off unless you turn it on.</p>
      </header>
      {#if !settings.config}
        <p class="muted">Connect the daemon to edit spend and telemetry.</p>
      {:else}
        <div class="spend" class:hot={spendHot} class:cap={spendPct >= 100}>
          <div class="spend-copy">
            <label for="spend">Daily spend cap</label>
            <p class="hint tight">
              Today: ${spentToday.toFixed(2)} of ${spendCap}
              {#if spendPct >= 100}
                · cap reached — Ask is blocked
              {:else if spendHot}
                · {spendPct}% used
              {/if}
              . Zero means no spend allowed.
              {#if spendStore.live}
                <span class="live-tag">live</span>
              {/if}
            </p>
          </div>
          <div class="spend-ctrl">
            <div class="gauge" aria-hidden="true" aria-valuenow={spendPct}>
              <div class="gauge-fill" style={`width: ${spendPct}%`}></div>
            </div>
            <div class="spend-row">
              <span class="currency">$</span>
              <input
                id="spend"
                type="number"
                min="0"
                step="1"
                value={spendCap}
                disabled={saving}
                onchange={(e) =>
                  void savePatch({
                    security: {
                      ...settings.config!.security,
                      spend_limit_usd_per_day: Number((e.currentTarget as HTMLInputElement).value),
                    },
                  })}
              />
              <span class="unit">/ day</span>
            </div>
          </div>
        </div>
        <label class="toggle-row">
          <input
            type="checkbox"
            checked={telemetryOn}
            onchange={(e) => void setTelemetry((e.currentTarget as HTMLInputElement).checked)}
          />
          <span>
            <strong>Telemetry</strong>
            <em>Anonymous usage signals. Off by default.</em>
          </span>
        </label>
      {/if}
      {#if note && note !== 'Saved'}
        <p class="note" class:warn={note.includes('offline') || note.includes('Error')}>{note}</p>
      {/if}
    </section>

    <!-- 06 Models -->
    <section class="plate" id="md-models">
      <header class="plate-head">
        <p class="cite">06 · providers</p>
        <h2>Models on this machine</h2>
        <p class="hint">Enable providers and store API keys locally. Deep routing stays in Ask.</p>
      </header>
      {#if providers.length === 0}
        <p class="muted">No providers reported. Connect the daemon to load the registry.</p>
      {:else}
        <ul class="prov-list">
          {#each providers as p (p.name)}
            <li class="prov-row" data-keyed={providerHasKey(p.name)}>
              <div>
                <strong>{p.name}</strong>
                <span class="prov-meta">
                  {providerHasKey(p.name) ? 'key on file' : 'no key'}
                  · {p.models?.length ?? 0} models
                </span>
              </div>
              <label class="mini-toggle">
                <input
                  type="checkbox"
                  checked={providerEnabled(p.name)}
                  disabled={!settings.config}
                  onchange={(e) =>
                    void toggleProvider(p.name, (e.currentTarget as HTMLInputElement).checked)}
                />
                <span>Enabled</span>
              </label>
            </li>
          {/each}
        </ul>
        <div class="key-form">
          <select bind:value={keyProvider} aria-label="Provider for key">
            {#each providers as p (p.name)}
              <option value={p.name}>{p.name}</option>
            {/each}
          </select>
          <input
            type="password"
            placeholder="API key"
            bind:value={keyDraft}
            autocomplete="off"
          />
          <button
            type="button"
            class="md-btn md-btn-primary"
            disabled={keyBusy || !keyDraft.trim()}
            onclick={() => void saveApiKey()}
          >
            {keyBusy ? 'Saving…' : 'Save key'}
          </button>
        </div>
        {#if apiKeys.length}
          <ul class="key-meta">
            {#each apiKeys as k (k.id)}
              <li>
                <span>{k.provider} · {k.label} · {k.auth_kind}</span>
                <button
                  type="button"
                  class="md-btn md-btn-danger"
                  disabled={keyBusy}
                  onclick={() => void deleteApiKey(k.id)}
                >Remove</button>
              </li>
            {/each}
          </ul>
        {/if}
        {#if keyNote}
          <p class="note">{keyNote}</p>
        {/if}
      {/if}
    </section>

    <!-- 07 Adaptive -->
    <section class="plate">
      <header class="plate-head">
        <p class="cite">07 · adaptive</p>
        <h2>What Condura has learned</h2>
        <p class="hint">Inferred from your behavior. Delete anything that does not fit.</p>
      </header>
      <div class="seg strength" role="group" aria-label="Adaptive strength">
        {#each STRENGTH_OPTS as opt (opt)}
          <button type="button" class:on={strength === opt} onclick={() => void setStrength(opt)}>
            {opt}
          </button>
        {/each}
      </div>
      {#if adaptiveLoading}
        <p class="muted">Loading learned profile…</p>
      {:else if adaptiveItems.length === 0}
        <p class="muted">Nothing learned yet. Use Condura — preferences will appear here.</p>
      {:else}
        <ul class="adapt-list">
          {#each adaptiveItems as item (item.field + item.value)}
            <li>
              <div>
                <strong>{item.value}</strong>
                <span class="adapt-ev">{item.evidence}</span>
                <span class="adapt-meta">{item.field} · {(item.confidence * 100).toFixed(0)}%</span>
              </div>
              <button
                type="button"
                class="md-btn md-btn-ghost"
                disabled={adaptiveBusy}
                onclick={() => void forgetAdaptive(item.field, item.value)}
              >Forget</button>
            </li>
          {/each}
        </ul>
      {/if}
      <div class="row-actions">
        <button type="button" class="md-btn" disabled={!adaptiveProfile} onclick={exportAdaptive}>
          Export profile
        </button>
        {#if confirmResetAdaptive}
          <button type="button" class="md-btn md-btn-danger" disabled={adaptiveBusy} onclick={() => void resetAdaptive()}>
            Confirm reset
          </button>
          <button type="button" class="md-btn md-btn-ghost" onclick={() => (confirmResetAdaptive = false)}>
            Cancel
          </button>
        {:else}
          <button type="button" class="md-btn md-btn-danger" disabled={adaptiveBusy} onclick={() => (confirmResetAdaptive = true)}>
            Reset everything
          </button>
        {/if}
      </div>
    </section>

    <!-- 08 Backup -->
    <section class="plate">
      <header class="plate-head">
        <p class="cite">08 · recovery</p>
        <h2>Backup & restore</h2>
        <p class="hint">Your data lives on this machine. Back it up before uninstalling. Dir: {backupDir}</p>
      </header>
      <div class="row-actions">
        <button type="button" class="md-btn md-btn-primary" disabled={backupBusy} onclick={() => void createBackup()}>
          {backupBusy ? 'Exporting…' : 'Export everything (.zip)'}
        </button>
        <button
          type="button"
          class="md-btn"
          disabled={restoreBusy || trust.backups.length === 0}
          onclick={() => (confirmRestorePath = trust.backups[0]?.path ?? null)}
        >
          Restore latest
        </button>
      </div>
      {#if confirmRestorePath}
        <div class="confirm">
          <p>Restore from <code>{confirmRestorePath}</code>? This replaces local data.</p>
          <div class="row-actions">
            <button type="button" class="md-btn md-btn-danger" disabled={restoreBusy} onclick={() => void performRestore(confirmRestorePath!)}>
              {restoreBusy ? 'Restoring…' : 'Confirm restore'}
            </button>
            <button type="button" class="md-btn md-btn-ghost" onclick={() => (confirmRestorePath = null)}>Cancel</button>
          </div>
        </div>
      {/if}
      {#if trust.backups.length}
        <ul class="backup-list">
          {#each trust.backups as b (b.path)}
            <li>
              <div>
                <strong>{b.name}</strong>
                <span class="adapt-meta">{formatBytes(b.size)} · {b.path}</span>
              </div>
              <button type="button" class="md-btn" disabled={restoreBusy} onclick={() => (confirmRestorePath = b.path)}>
                Restore
              </button>
            </li>
          {/each}
        </ul>
      {:else}
        <p class="muted">No backups on file yet.</p>
      {/if}
      {#if backupNote}
        <p class="note">{backupNote}</p>
      {/if}
    </section>

    <!-- 09 Trust -->
    <section class="plate">
      <header class="plate-head">
        <p class="cite">09 · trust · live</p>
        <h2>What this build can do</h2>
        <p class="hint">Runtime facts from the daemon — never aspirations.</p>
      </header>
      {#if capsError}
        <p class="muted">{capsError}</p>
      {:else if !capabilities}
        <p class="muted">Reading capabilities…</p>
      {:else}
        <div class="trust-grid">
          <div class="trust-card" data-on={capabilities.kill_switch.layer1_hotkey}>
            <span class="trust-k">layer 1</span>
            <strong>Hard hotkey</strong>
            <span>{capabilities.kill_switch.layer1_hotkey ? 'wired' : 'not wired'}</span>
          </div>
          <div class="trust-card" data-on={capabilities.kill_switch.layer2_watchdog}>
            <span class="trust-k">layer 2</span>
            <strong>Watchdog</strong>
            <span>{capabilities.kill_switch.layer2_watchdog ? 'armed' : 'off'}</span>
          </div>
          <div class="trust-card" data-on={capabilities.kill_switch.layer3_network_isolation.os_process}>
            <span class="trust-k">layer 3</span>
            <strong>Network isolation</strong>
            <span>
              {capabilities.kill_switch.layer3_network_isolation.os_process
                ? 'os process'
                : capabilities.kill_switch.layer3_network_isolation.in_process
                  ? 'in-process soft guard'
                  : `deferred · ${capabilities.kill_switch.layer3_network_isolation.deferred_to || '—'}`}
            </span>
          </div>
          <div class="trust-card" data-on={capabilities.audit.hmac_subkey}>
            <span class="trust-k">audit</span>
            <strong>Ledger seal</strong>
            <span>
              {capabilities.audit.hmac_subkey ? 'hmac' : 'unsealed'}
              {capabilities.audit.redaction ? ' · redaction' : ''}
            </span>
          </div>
        </div>
        <p class="trust-cu">
          Computer-use · orax {capabilities.computer_use.orax} · mac_cua {capabilities.computer_use.mac_cua} ·
          macos_mcp {capabilities.computer_use.macos_mcp} · vision {capabilities.computer_use.vision_cua}
        </p>
      {/if}
      <button type="button" class="md-btn" onclick={() => void loadCapabilities()}>Refresh capabilities</button>
    </section>

    <!-- 10 Keys -->
    <section class="plate keys">
      <header class="plate-head">
        <p class="cite">10 · atlas · keys</p>
        <h2>Always within reach</h2>
        <p class="hint">The same map About teaches — wired into the shell.</p>
      </header>
      <ul class="key-list">
        {#each KEYS as k (k.label)}
          <li>
            <kbd>{k.keys === 'K' ? `${modLabel}K` : k.keys}</kbd>
            <span>{k.label}</span>
          </li>
        {/each}
      </ul>
    </section>

    <!-- 11 Welcome again -->
    <section class="plate setup">
      <header class="plate-head">
        <p class="cite">11 · welcome</p>
        <h2>First-run, again</h2>
        <p class="hint">
          Re-walk EULA, OS permissions, and hotkey. Your chats, keys, and audit stay put.
        </p>
      </header>
      {#if confirmRerunSetup}
        <p class="muted">Walk the welcome flow again? Data is preserved.</p>
        <div class="actions-row">
          <button
            type="button"
            class="md-btn md-btn-primary"
            disabled={rerunBusy}
            onclick={() => void confirmAndRerunSetup()}
          >
            {rerunBusy ? 'Opening…' : 'Re-run setup'}
          </button>
          <button type="button" class="md-btn md-btn-ghost" disabled={rerunBusy} onclick={cancelRerunSetup}>
            Cancel
          </button>
        </div>
      {:else}
        <button type="button" class="md-btn md-btn-ghost" onclick={requestRerunSetup}>
          Re-run setup
        </button>
      {/if}
    </section>

    <div class="doors">
      <button type="button" class="door" onclick={() => go('#/account')}>
        <span class="door-k">passport</span>
        <strong>Account</strong>
        <span>Optional cloud doors</span>
      </button>
      <button type="button" class="door" onclick={() => go('#/sync')}>
        <span class="door-k">pair</span>
        <strong>Sync</strong>
        <span>Nearby devices</span>
      </button>
      <button type="button" class="door" onclick={() => go('#/channels')}>
        <span class="door-k">reach</span>
        <strong>Channels</strong>
        <span>Telegram & more</span>
      </button>
      <button type="button" class="door" onclick={() => go('#/audit')}>
        <span class="door-k">ledger</span>
        <strong>Audit</strong>
        <span>What Condura did</span>
      </button>
      <button type="button" class="door" onclick={() => go('#/replay')}>
        <span class="door-k">theatre</span>
        <strong>Replay</strong>
        <span>Day meridian frames</span>
      </button>
      <button type="button" class="door" onclick={() => go('#/about')}>
        <span class="door-k">colophon</span>
        <strong>About</strong>
        <span>Seven stations</span>
      </button>
    </div>
  </div>
</MeridianPage>

<style>
  .desk { display: grid; gap: 16px; }
  .contract {
    display: flex; align-items: flex-start; gap: 10px; margin: 0;
    padding: 12px 14px; border-radius: 14px; border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 80%, transparent);
    font-size: 13px; line-height: 1.45; color: var(--md-ink-mute);
  }
  .contract.hot {
    border-color: color-mix(in oklab, var(--md-live) 28%, transparent);
    background: color-mix(in oklab, var(--md-live) 6%, var(--md-surface));
  }
  .contract.off { border-color: color-mix(in oklab, var(--md-halt) 22%, var(--md-line)); }
  .live-dot {
    width: 8px; height: 8px; margin-top: 5px; flex: none; border-radius: 50%;
    background: var(--md-ink-faint);
  }
  .contract.hot .live-dot {
    background: var(--md-live);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-live) 16%, transparent);
  }
  .pipe {
    display: flex; flex-wrap: wrap; gap: 6px; margin: 0; padding: 0; list-style: none;
  }
  .pipe li {
    display: inline-flex; align-items: center; gap: 8px;
    padding: 7px 11px; border-radius: 999px; border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 70%, transparent);
  }
  .pipe .n { font-family: var(--md-font-mono); font-size: 10px; color: var(--md-cobalt); }
  .pipe .t { font-size: 12px; font-weight: 700; color: var(--md-ink-soft); }
  .plate {
    border-radius: 22px; border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
    padding: 22px 24px 24px;
    box-shadow: inset 0 1px 0 color-mix(in oklab, var(--md-surface) 55%, transparent);
  }
  #md-models,
  #md-spend {
    scroll-margin-top: 24px;
  }
  .plate-head { margin-bottom: 18px; }
  .cite {
    font-family: var(--md-font-mono); font-size: 10px; letter-spacing: 0.14em;
    text-transform: uppercase; color: var(--md-ink-faint); margin: 0 0 8px;
  }
  h2 {
    font-family: var(--md-font-display); font-size: 22px; letter-spacing: -0.04em; margin: 0 0 6px;
  }
  .hint { margin: 0; font-size: 13px; color: var(--md-ink-faint); line-height: 1.45; max-width: 52ch; }
  .hint.tight { margin-top: 4px; }
  .muted { color: var(--md-ink-mute); font-size: 14px; margin: 0 0 14px; line-height: 1.5; }
  .actions-row { display: flex; flex-wrap: wrap; gap: 10px; align-items: center; }
  .note {
    margin: 14px 0 0; font-family: var(--md-font-mono); font-size: 11px;
    letter-spacing: 0.04em; color: var(--md-live);
  }
  .note.warn { color: var(--md-halt); }

  .seg {
    display: inline-flex; padding: 4px; border-radius: 999px;
    background: var(--md-stage); border: 1px solid var(--md-line); flex-wrap: wrap;
  }
  .seg button {
    padding: 8px 14px; border-radius: 999px; font-weight: 700; font-size: 13px;
    color: var(--md-ink-mute); cursor: pointer;
  }
  .seg button.on {
    background: var(--md-cobalt); color: #fff;
    box-shadow: 0 8px 18px -10px color-mix(in oklab, var(--md-cobalt) 70%, transparent);
  }
  .seg button:focus-visible { outline: none; box-shadow: var(--md-focus); }
  .swatch { display: flex; gap: 8px; margin-top: 16px; }
  .sw { width: 28px; height: 28px; border-radius: 10px; border: 1px solid var(--md-line); }
  .sw.mist { background: var(--md-mist); }
  .sw.stage { background: var(--md-stage); }
  .sw.cobalt { background: var(--md-cobalt); }
  .sw.live { background: var(--md-live); }

  .hotkey-stage { display: flex; flex-wrap: wrap; align-items: center; gap: 14px; margin-bottom: 16px; }
  .hotkey-display {
    font-family: var(--md-font-display); font-size: 28px; letter-spacing: -0.04em;
    min-width: 140px; padding: 12px 16px; border-radius: 14px;
    border: 1px solid var(--md-line-strong); background: var(--md-stage);
  }
  .recording { outline: 2px solid var(--md-cobalt); }

  .toggle-row {
    display: flex; gap: 12px; align-items: flex-start; cursor: pointer;
    padding: 12px 0 0; border-top: 1px solid var(--md-line); margin-top: 4px;
  }
  .toggle-row input { margin-top: 4px; }
  .toggle-row strong { display: block; font-size: 14px; }
  .toggle-row em { display: block; font-style: normal; font-size: 12px; color: var(--md-ink-faint); margin-top: 2px; }

  .autonomy {
    display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; margin-bottom: 22px;
  }
  .auto-card {
    text-align: left; padding: 14px; border-radius: 16px; border: 1px solid var(--md-line-strong);
    background: var(--md-stage); cursor: pointer; display: grid; gap: 6px; color: inherit;
  }
  .auto-card strong { font-family: var(--md-font-display); font-size: 15px; letter-spacing: -0.03em; }
  .auto-card span:nth-child(3) { font-size: 12px; line-height: 1.4; color: var(--md-ink-mute); }
  .auto-id { font-family: var(--md-font-mono); font-size: 10px; color: var(--md-ink-faint); }
  .auto-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--md-ink-faint); }
  .auto-card[data-tone='cobalt'].on {
    border-color: color-mix(in oklab, var(--md-cobalt) 55%, transparent);
    background: color-mix(in oklab, var(--md-cobalt) 10%, var(--md-surface));
  }
  .auto-card[data-tone='cobalt'].on .auto-dot { background: var(--md-cobalt); }
  .auto-card[data-tone='live'].on {
    border-color: color-mix(in oklab, var(--md-live) 55%, transparent);
    background: color-mix(in oklab, var(--md-live) 10%, var(--md-surface));
  }
  .auto-card[data-tone='live'].on .auto-dot { background: var(--md-live); }
  .auto-card[data-tone='caution'].on {
    border-color: color-mix(in oklab, var(--md-halt) 40%, transparent);
    background: color-mix(in oklab, var(--md-halt) 8%, var(--md-surface));
  }
  .auto-card[data-tone='caution'].on .auto-dot { background: var(--md-halt); }
  .auto-card:hover:not(.on) { border-color: var(--md-cobalt); }
  .auto-card:focus-visible { outline: none; box-shadow: var(--md-focus); }

  .matrix { display: grid; gap: 8px; }
  .matrix-k {
    font-family: var(--md-font-mono); font-size: 10px; letter-spacing: 0.12em;
    text-transform: uppercase; color: var(--md-ink-faint); margin: 0 0 4px;
  }
  .matrix-row {
    display: grid; grid-template-columns: 1fr auto; gap: 12px; align-items: center;
    padding: 8px 0; border-bottom: 1px solid var(--md-line);
  }
  .matrix-label { font-size: 13px; font-weight: 600; }
  .dial { display: inline-flex; gap: 4px; padding: 3px; border-radius: 999px; background: var(--md-stage); border: 1px solid var(--md-line); }
  .dial button {
    padding: 5px 10px; border-radius: 999px; font-family: var(--md-font-mono);
    font-size: 10px; letter-spacing: 0.04em; color: var(--md-ink-faint); cursor: pointer;
  }
  .dial button.on[data-l='auto'] { background: var(--md-live); color: #fff; }
  .dial button.on[data-l='warn'] { background: var(--md-cobalt); color: #fff; }
  .dial button.on[data-l='block'] { background: var(--md-halt); color: #fff; }

  .spend {
    display: grid; grid-template-columns: 1fr 1fr; gap: 18px; align-items: end;
    padding-bottom: 14px; margin-bottom: 4px;
  }
  .spend.hot .gauge-fill {
    background: linear-gradient(90deg, #c4892a, var(--md-halt));
  }
  .spend.cap .gauge {
    border-color: color-mix(in oklab, var(--md-halt) 40%, var(--md-line));
  }
  .spend label { font-family: var(--md-font-display); font-size: 15px; font-weight: 700; }
  .live-tag {
    display: inline-block;
    margin-left: 6px;
    font-family: var(--md-font-mono);
    font-size: 9px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-live);
  }
  .gauge {
    height: 6px; border-radius: 999px; background: var(--md-stage);
    border: 1px solid var(--md-line); overflow: hidden; margin-bottom: 10px;
  }
  .gauge-fill {
    height: 100%; border-radius: inherit;
    background: linear-gradient(90deg, var(--md-live), var(--md-cobalt));
    transition: width 280ms var(--md-ease);
  }
  .spend-row { display: flex; align-items: center; gap: 8px; }
  .currency, .unit { font-family: var(--md-font-mono); font-size: 12px; color: var(--md-ink-faint); }
  .spend-row input {
    width: 96px; padding: 10px 12px; border-radius: 12px; border: 1px solid var(--md-line-strong);
    background: var(--md-surface); font-family: var(--md-font-mono); font-size: 16px; font-weight: 600;
  }
  .spend-row input:focus-visible { outline: none; border-color: var(--md-cobalt); box-shadow: var(--md-focus); }

  .prov-list, .adapt-list, .backup-list, .key-meta {
    list-style: none; margin: 0 0 14px; padding: 0; display: grid; gap: 8px;
  }
  .prov-row, .adapt-list li, .backup-list li, .key-meta li {
    display: flex; justify-content: space-between; align-items: center; gap: 12px;
    padding: 12px; border-radius: 14px; border: 1px solid var(--md-line); background: var(--md-stage);
  }
  .prov-row[data-keyed='true'] { border-left: 3px solid var(--md-live); }
  .prov-meta, .adapt-meta, .adapt-ev {
    display: block; font-size: 12px; color: var(--md-ink-faint); margin-top: 2px;
  }
  .adapt-ev { color: var(--md-ink-mute); }
  .mini-toggle { display: inline-flex; align-items: center; gap: 6px; font-size: 12px; font-weight: 600; }
  .key-form {
    display: grid; grid-template-columns: 140px 1fr auto; gap: 8px; margin-bottom: 12px;
  }
  .key-form select, .key-form input {
    padding: 10px 12px; border-radius: 12px; border: 1px solid var(--md-line-strong);
    background: var(--md-surface); font-size: 13px;
  }
  .row-actions { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 12px; }
  .confirm {
    margin-top: 12px; padding: 12px; border-radius: 14px;
    border: 1px solid color-mix(in oklab, var(--md-halt) 30%, var(--md-line));
    background: color-mix(in oklab, var(--md-halt) 6%, var(--md-surface));
  }
  .confirm code { font-family: var(--md-font-mono); font-size: 12px; word-break: break-all; }

  .trust-grid {
    display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; margin-bottom: 12px;
  }
  .trust-card {
    padding: 14px; border-radius: 16px; border: 1px solid var(--md-line); background: var(--md-stage);
    display: grid; gap: 4px;
  }
  .trust-card[data-on='true'] {
    border-color: color-mix(in oklab, var(--md-live) 35%, transparent);
  }
  .trust-k {
    font-family: var(--md-font-mono); font-size: 10px; letter-spacing: 0.12em;
    text-transform: uppercase; color: var(--md-ink-faint);
  }
  .trust-card strong { font-family: var(--md-font-display); font-size: 15px; }
  .trust-card span:last-child { font-size: 12px; color: var(--md-ink-mute); }
  .trust-cu {
    font-family: var(--md-font-mono); font-size: 11px; color: var(--md-ink-faint);
    line-height: 1.5; margin: 0 0 12px;
  }

  .key-list { display: grid; gap: 10px; margin: 0; padding: 0; }
  .key-list li {
    display: flex; align-items: center; gap: 14px; padding: 10px 0;
    border-bottom: 1px solid var(--md-line);
  }
  .key-list li:last-child { border-bottom: 0; }
  .key-list kbd {
    font-family: var(--md-font-mono); font-size: 11px; min-width: 88px; text-align: center;
    padding: 6px 10px; border-radius: 8px; background: var(--md-stage);
    border: 1px solid var(--md-line-strong); color: var(--md-ink-soft);
  }
  .key-list span { font-size: 14px; color: var(--md-ink-mute); }

  .doors {
    display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px;
  }
  .door {
    appearance: none; text-align: left; border: 1px solid var(--md-line-strong);
    background: var(--md-stage); border-radius: 18px; padding: 14px; cursor: pointer;
    display: grid; gap: 6px; color: inherit;
  }
  .door-k {
    font-family: var(--md-font-mono); font-size: 10px; letter-spacing: 0.14em;
    text-transform: uppercase; color: var(--md-ink-faint);
  }
  .door strong { font-family: var(--md-font-display); font-size: 16px; letter-spacing: -0.03em; }
  .door > span:not(.door-k) { font-size: 12px; line-height: 1.4; color: var(--md-ink-mute); }
  .door:hover { border-color: var(--md-cobalt); transform: translateY(-2px); box-shadow: var(--md-shadow); }
  .door:focus-visible { outline: none; box-shadow: var(--md-focus); border-color: var(--md-cobalt); }

  @media (max-width: 720px) {
    .autonomy, .doors, .trust-grid, .spend, .key-form { grid-template-columns: 1fr; }
    .matrix-row { grid-template-columns: 1fr; gap: 8px; }
  }
  @media (max-width: 420px) {
    .plate { padding: 18px 16px 20px; }
    .seg { width: 100%; }
    .seg button { flex: 1; }
  }
</style>
