<script lang="ts">
  /**
   * Settings — clear tabs, plain language, advanced options tucked away.
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
  type SettingsTab = 'general' | 'permissions' | 'control' | 'models' | 'data'

  let theme = $state<ResolvedTheme>(getResolvedTheme())
  let saving = $state(false)
  let note = $state('')
  let modLabel = $state('⌘')
  let noteTimer: ReturnType<typeof setTimeout> | null = null
  let tab = $state<SettingsTab>('general')
  let showTaskRules = $state(false)
  let showSafetyDetail = $state(false)

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
  /** Draft base URLs keyed by provider name (local backends). */
  let baseUrlDraft = $state<Record<string, string>>({})
  let askDefault = $state('')

  const KEYLESS = new Set(['ollama', 'localai', 'lmstudio', 'vllm', 'custom'])
  const BASE_URL_PROVIDERS = new Set(['ollama', 'localai', 'lmstudio', 'vllm', 'custom'])
  const PROVIDER_LABELS: Record<string, string> = {
    anthropic: 'Anthropic',
    openai: 'OpenAI',
    google: 'Google',
    xai: 'xAI',
    mistral: 'Mistral',
    deepseek: 'DeepSeek',
    openrouter: 'OpenRouter',
    groq: 'Groq',
    together: 'Together',
    fireworks: 'Fireworks',
    ollama: 'Ollama',
    localai: 'LocalAI',
    lmstudio: 'LM Studio',
    vllm: 'vLLM',
    custom: 'Custom',
  }

  let expandedProvider = $state<string | null>(null)

  const askDefaultOptions = $derived(
    providers.flatMap((p) =>
      (p.models ?? []).map((m) => ({
        value: `${p.name}:${m.id}`,
        label: `${providerLabel(p.name)} · ${m.id}`,
      }))
    )
  )

  const localProviders = $derived(
    providers.filter((p) => !providerNeedsKey(p.name))
  )
  const cloudProviders = $derived(
    providers.filter((p) => providerNeedsKey(p.name))
  )

  const askModelReady = $derived.by(() => {
    if (!askDefault.includes(':')) return false
    const [name] = askDefault.split(':')
    if (!name) return false
    if (!providerEnabled(name)) return false
    return providerCredentialed(name)
  })

  const askModelStatus = $derived.by(() => {
    if (!settings.config) return { label: 'Offline', state: 'halt' as const }
    if (!askDefault.includes(':')) return { label: 'Not set', state: 'halt' as const }
    const [name] = askDefault.split(':')
    if (!name) return { label: 'Not set', state: 'halt' as const }
    if (!providerEnabled(name)) return { label: 'Provider off', state: 'warn' as const }
    if (!providerCredentialed(name)) {
      return { label: 'Needs API key', state: 'halt' as const }
    }
    return { label: 'Ready for Ask', state: 'ok' as const }
  })

  const modelsTabNeedsAttention = $derived(
    !!settings.config && providers.length > 0 && !askModelReady
  )

  function providerLabel(name: string): string {
    return PROVIDER_LABELS[name] ?? name
  }

  function providerKind(name: string): string {
    return providerNeedsKey(name) ? 'Cloud' : 'On this Mac'
  }

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
      body: 'Always wait for your OK before acting. Best for beginners.',
      tone: 'cobalt',
    },
    {
      id: 'warn',
      title: 'Suggest',
      body: 'Shows a plan, then asks when something needs consent.',
      tone: 'live',
    },
    {
      id: 'autonomous',
      title: 'More automatic',
      body: 'Can do routine work alone. Sensitive actions still need you.',
      tone: 'caution',
    },
  ]

  const TASK_TYPES: { key: string; label: string }[] = [
    { key: 'coding', label: 'Coding' },
    { key: 'file_operations', label: 'Files' },
    { key: 'web_browsing', label: 'Web' },
    { key: 'email', label: 'Email' },
    { key: 'calendar', label: 'Calendar' },
    { key: 'messaging', label: 'Messaging' },
    { key: 'shell_commands', label: 'Terminal' },
    { key: 'computer_use', label: 'Computer control' },
    { key: 'research', label: 'Research' },
    { key: 'image_generation', label: 'Images' },
    { key: 'code_review', label: 'Code review' },
  ]

  const STRENGTH_OPTS: { id: AdaptiveStrength; label: string }[] = [
    { id: 'off', label: 'Off' },
    { id: 'cautious', label: 'Careful' },
    { id: 'balanced', label: 'Balanced' },
    { id: 'aggressive', label: 'Strong' },
  ]

  const KEYS = [
    { keys: 'K', label: 'Open search' },
    { keys: 'Shift+T', label: 'Switch light / dark' },
    { keys: 'Esc', label: 'Close popups' },
    { keys: 'Halt', label: 'Stop everything (dock or search)' },
  ]

  const TABS: { id: SettingsTab; label: string; shortcut: string }[] = [
    { id: 'general', label: 'General', shortcut: '⌘1' },
    { id: 'permissions', label: 'Permissions', shortcut: '⌘2' },
    { id: 'control', label: 'Control', shortcut: '⌘3' },
    { id: 'models', label: 'Models', shortcut: '⌘4' },
    { id: 'data', label: 'Data', shortcut: '⌘5' },
  ]
  const tabIdx = $derived(Math.max(0, TABS.findIndex((t) => t.id === tab)))
  let tabNavEl = $state<HTMLElement | null>(null)

  /** ⌘1..5 → tab index. Wired from MeridianShell's global keydown so the
   *  shortcut works from anywhere (palette open, consent up, etc.) when
   *  the user is on the Settings route. */
  function selectTabByIndex(i: number): void {
    const target = TABS[i]
    if (!target) return
    selectTab(target.id)
    queueMicrotask(() => {
      const btn = tabNavEl?.querySelector<HTMLElement>(`[data-tab-id="${target.id}"]`)
      btn?.focus({ preventScroll: true })
    })
  }

  // Listen for the global ⌘1..5 dispatch from MeridianShell. Settings is
  // only mounted on the settings route, so the listener is naturally
  // scoped — but we still defensively ignore events that arrive after
  // a route change by checking the current route.
  $effect(() => {
    const onTab = (e: Event): void => {
      const idx = (e as CustomEvent<{ index: number }>).detail?.index
      if (typeof idx === 'number') selectTabByIndex(idx)
    }
    window.addEventListener('condura:settings-tab', onTab)
    return () => window.removeEventListener('condura:settings-tab', onTab)
  })

  /**
   * Roving tabindex across the settings tabs (mirrors MeridianDock).
   * ArrowLeft/Right move focus AND selection; Home/End jump to the
   * ends. Tabs are the page-level tab pattern (W3C APG), but the
   * Meridian shell uses hash deep-links (#/settings/models, etc.) and
   * each section is its own plate — so the cheaper "toolbar with
   * auto-activation" pattern fits.
   */
  function onTabKey(e: KeyboardEvent): void {
    let next = tabIdx
    if (e.key === 'ArrowRight') next = (tabIdx + 1) % TABS.length
    else if (e.key === 'ArrowLeft') next = (tabIdx - 1 + TABS.length) % TABS.length
    else if (e.key === 'Home') next = 0
    else if (e.key === 'End') next = TABS.length - 1
    else return
    e.preventDefault()
    const target = TABS[next]
    if (!target) return
    selectTab(target.id)
    queueMicrotask(() => {
      const btn = tabNavEl?.querySelector<HTMLElement>(`[data-tab-id="${target.id}"]`)
      btn?.focus({ preventScroll: true })
    })
  }

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
      ? 'Saving…'
      : note.includes('offline')
        ? note
        : note === 'Saved'
          ? 'Saved'
          : offline
            ? 'Offline — you can still change theme'
            : `Permissions ${grantedPerms}/5 · $${spentToday.toFixed(2)} of $${spendCap} today`
  )

  onMount(() => {
    theme = getResolvedTheme()
    modLabel = /Mac|iPhone|iPad/.test(navigator.platform) ? '⌘' : 'Ctrl'
    const off = onThemeChange((r) => {
      theme = r
    })
    applyHashTab()
    void bootstrap().then(() => scrollToHashSection())
    const onHash = () => {
      applyHashTab()
      scrollToHashSection()
    }
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

  /** Deep links: #/settings/models · #/settings/spend */
  function applyHashTab(): void {
    const h = window.location.hash || ''
    if (/spend|cap/i.test(h)) tab = 'control'
    else if (/models|providers|keys/i.test(h)) tab = 'models'
    else if (/perm/i.test(h)) tab = 'permissions'
  }

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

  function selectTab(next: SettingsTab): void {
    tab = next
    // Keep deep-links useful for Ask → Settings jumps.
    if (next === 'models') window.location.hash = '#/settings/models'
    else if (next === 'control' && /spend|cap/i.test(window.location.hash || '')) {
      /* keep spend fragment */
    } else if (/^#\/settings\//.test(window.location.hash || '')) {
      window.location.hash = '#/settings'
    }
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
        ? 'Offline — connect to save'
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
      providers = (p ?? []).map((pr) => ({
        name: pr.name,
        models: (pr.models ?? []).map((m) => ({ id: m?.id ?? '' })).filter((m) => !!m.id),
        available: pr.available,
      }))
      // Settings shows the full catalog — fill any empty model lists.
      await Promise.all(
        providers.map(async (pr, i) => {
          if (pr.models.length > 0) return
          try {
            const ms = await ipc.providersModels(pr.name)
            providers[i] = {
              ...pr,
              models: (ms ?? []).map((m) => ({ id: m.id })).filter((m) => !!m.id),
            }
          } catch {
            /* leave empty */
          }
        })
      )
      providers = [...providers]
      apiKeys = k ?? []
      if (!keyProvider && providers[0]) keyProvider = providers[0].name
      const urls: Record<string, string> = {}
      for (const pr of providers) {
        urls[pr.name] = settings.config?.llm?.providers?.[pr.name]?.base_url ?? ''
      }
      baseUrlDraft = urls
      // Prefer first enabled provider with a default_model.
      const cfg = settings.config?.llm?.providers
      if (cfg) {
        const hit = Object.entries(cfg).find(([, v]) => v.enabled && v.default_model)
        if (hit) askDefault = `${hit[0]}:${hit[1].default_model}`
      }
      if (!askDefault) {
        const first = providers.find((pr) => (pr.models?.length ?? 0) > 0)
        if (first?.models?.[0]) askDefault = `${first.name}:${first.models[0].id}`
      }
    } catch {
      providers = []
      apiKeys = []
    }
  }

  function providerNeedsKey(name: string): boolean {
    return !KEYLESS.has(name.toLowerCase())
  }

  function providerShowsBaseURL(name: string): boolean {
    return BASE_URL_PROVIDERS.has(name.toLowerCase())
  }

  function providerDefaultModel(name: string): string {
    return settings.config?.llm?.providers?.[name]?.default_model ?? ''
  }

  async function setProviderDefaultModel(name: string, model: string): Promise<void> {
    if (!settings.config || !model) return
    const providersCfg = { ...(settings.config.llm?.providers ?? {}) }
    const cur = providersCfg[name] ?? {
      enabled: false,
      api_key: '',
      base_url: '',
      default_model: '',
    }
    providersCfg[name] = { ...cur, default_model: model, enabled: true }
    await savePatch({
      llm: { ...settings.config.llm, providers: providersCfg },
    })
    askDefault = `${name}:${model}`
  }

  async function setAskDefaultModel(value: string): Promise<void> {
    askDefault = value
    if (!value.includes(':')) return
    const [provider, model] = value.split(':')
    if (!provider || !model) return
    await setProviderDefaultModel(provider, model)
  }

  async function saveBaseURL(name: string): Promise<void> {
    if (!settings.config) return
    const url = (baseUrlDraft[name] ?? '').trim()
    const providersCfg = { ...(settings.config.llm?.providers ?? {}) }
    const cur = providersCfg[name] ?? {
      enabled: false,
      api_key: '',
      base_url: '',
      default_model: '',
    }
    providersCfg[name] = { ...cur, base_url: url }
    await savePatch({
      llm: { ...settings.config.llm, providers: providersCfg },
    })
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
    if (apiKeys.some((k) => k.provider === name && k.has_token)) return true
    // Daemon may mark the provider available via env / OS keychain without vault meta.
    const info = providers.find((p) => p.name === name)
    if (info?.available) return true
    const cfgKey = settings.config?.llm?.providers?.[name]?.api_key?.trim()
    return !!cfgKey && cfgKey !== '***' && !cfgKey.startsWith('••••')
  }

  function providerCredentialed(name: string): boolean {
    return !providerNeedsKey(name) || providerHasKey(name)
  }

  function providerStatus(name: string): string {
    if (!providerNeedsKey(name)) {
      return providerEnabled(name) ? 'ready' : 'off'
    }
    if (providerHasKey(name)) return providerEnabled(name) ? 'key ready · on' : 'key ready · off'
    return 'needs API key'
  }

  async function toggleProvider(name: string, enabled: boolean): Promise<void> {
    if (!settings.config) return
    if (enabled && providerNeedsKey(name) && !providerHasKey(name)) {
      keyProvider = name
      keyNote = `Add an API key for ${providerLabel(name)} before turning it on.`
      expandedProvider = name
      return
    }
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
      backupNote = `Backup saved · ${path}`
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
      backupNote = 'Restore complete'
    } catch (e) {
      backupNote = String(e)
    } finally {
      restoreBusy = false
    }
  }
</script>

<svelte:window onkeydown={onHotkeyKey} />

<MeridianPage
  kicker="Preferences"
  title="Settings"
  lead="Five places to steer Condura: look & feel, OS access, how it asks, which model powers Ask, and what it remembers."
>
  <div class="desk">
    <div class="status-bar" class:ok={note === 'Saved'} class:warn={offline || note.includes('offline')}>
      <span class="dot" aria-hidden="true"></span>
      <span>{liveNote}</span>
      {#if keyedProviders > 0}
        <span class="sep">·</span>
        <span>{keyedProviders} API key{keyedProviders === 1 ? '' : 's'}</span>
      {/if}
    </div>

    <!-- onkeydown is delegated from inner tab <button>s — the <nav> is
         the bubble target so arrow-key navigation between tabs has one
         listener, not N. Inner buttons are themselves keyboard-accessible. -->
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <nav class="tabs" aria-label="Settings sections" bind:this={tabNavEl} onkeydown={onTabKey}>
      {#each TABS as t (t.id)}
        <button
          type="button"
          class="tab"
          class:on={tab === t.id}
          class:need={t.id === 'models' && modelsTabNeedsAttention}
          tabindex={tab === t.id ? 0 : -1}
          data-tab-id={t.id}
          aria-current={tab === t.id ? 'page' : undefined}
          onclick={() => selectTab(t.id)}
        >
          {t.label}
          <span class="tab-kbd" aria-hidden="true">{t.shortcut}</span>
          {#if t.id === 'models' && modelsTabNeedsAttention}
            <span class="tab-dot" aria-hidden="true"></span>
          {/if}
        </button>
      {/each}
    </nav>

    {#if tab === 'general'}
      <section class="plate" aria-labelledby="gen-theme">
        <header class="plate-head">
          <h2 id="gen-theme">Appearance</h2>
          <p class="hint">Choose light or dark. Shortcut: Shift+T.</p>
        </header>
        <div class="seg" role="group" aria-label="Theme">
          <button type="button" class:on={theme === 'light'} onclick={() => setTheme('light')}>Light</button>
          <button type="button" class:on={theme === 'dark'} onclick={() => setTheme('dark')}>Dark</button>
        </div>
      </section>

      <section class="plate">
        <header class="plate-head">
          <h2>Wake hotkey</h2>
          <p class="hint">Press this combo anytime to open Condura quickly.</p>
        </header>
        <div class="hotkey-stage">
          <div class="hotkey-display">{hotkeyDisplay || 'Not set'}</div>
          <button
            type="button"
            class="md-btn md-btn-primary"
            class:recording={recordingHotkey}
            onclick={() => (recordingHotkey = !recordingHotkey)}
          >
            {recordingHotkey ? 'Press keys… Esc cancels' : 'Change hotkey'}
          </button>
        </div>
        <label class="toggle-row">
          <input
            type="checkbox"
            checked={wakeEnabled}
            onchange={(e) => void setWakeEnabled((e.currentTarget as HTMLInputElement).checked)}
          />
          <span>
            <strong>Also listen for “hey Condura”</strong>
            <em>Uses the microphone on this computer.</em>
          </span>
        </label>
      </section>

      <section class="plate">
        <header class="plate-head">
          <h2>Keyboard shortcuts</h2>
          <p class="hint">Handy keys while Condura is open.</p>
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
    {:else if tab === 'permissions'}
      <MeridianSettingsPerms />
    {:else if tab === 'control'}
      <section class="plate gate">
        <header class="plate-head">
          <h2>How Condura asks</h2>
          <p class="hint">Pick a default. You can still stop anything with Halt.</p>
        </header>

        {#if !settings.config}
          <p class="muted">Connect Condura to change these options.</p>
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
              </button>
            {/each}
          </div>

          <button
            type="button"
            class="adv-toggle"
            aria-expanded={showTaskRules}
            onclick={() => (showTaskRules = !showTaskRules)}
          >
            {showTaskRules ? 'Hide' : 'Show'} rules by task type
          </button>

          {#if showTaskRules}
            <div class="matrix">
              <p class="matrix-k">Allow = run · Ask = confirm · Block = never</p>
              {#each autonomyTasks as t (t.key)}
                <div class="matrix-row">
                  <span class="matrix-label">{t.label}</span>
                  <div class="dial" role="group" aria-label="{t.label} autonomy">
                    <button
                      type="button"
                      class:on={t.level === 'autonomous'}
                      data-l="auto"
                      onclick={() => void setTaskAutonomy(t.key, 'autonomous')}
                    >Allow</button>
                    <button
                      type="button"
                      class:on={t.level === 'warn'}
                      data-l="warn"
                      onclick={() => void setTaskAutonomy(t.key, 'warn')}
                    >Ask</button>
                    <button
                      type="button"
                      class:on={t.level === 'block'}
                      data-l="block"
                      onclick={() => void setTaskAutonomy(t.key, 'block')}
                    >Block</button>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        {/if}
      </section>

      <section class="plate" id="md-spend">
        <header class="plate-head">
          <h2>Daily spend limit</h2>
          <p class="hint">Stops cloud model use once today’s cost hits this amount. Set 0 to block paid models.</p>
        </header>
        {#if !settings.config}
          <p class="muted">Connect Condura to edit the limit.</p>
        {:else}
          <div class="spend" class:hot={spendHot} class:cap={spendPct >= 100}>
            <div class="spend-copy">
              <p class="hint tight">
                Used today: <strong>${spentToday.toFixed(2)}</strong> of ${spendCap}
                {#if spendPct >= 100}
                  — limit reached
                {:else if spendHot}
                  — {spendPct}%
                {/if}
              </p>
            </div>
            <div class="spend-ctrl">
              <div class="gauge" aria-hidden="true">
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
                  aria-label="Daily spend limit in dollars"
                  onchange={(e) =>
                    void savePatch({
                      security: {
                        ...settings.config!.security,
                        spend_limit_usd_per_day: Number((e.currentTarget as HTMLInputElement).value),
                      },
                    })}
                />
                <span class="unit">per day</span>
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
              <strong>Share anonymous usage</strong>
              <em>Helps improve Condura. Off by default. Never includes your chats’ content.</em>
            </span>
          </label>
        {/if}
        {#if note && note !== 'Saved'}
          <p class="note" class:warn={note.includes('offline') || note.includes('Error')}>{note}</p>
        {/if}
      </section>
    {:else if tab === 'models'}
      {#snippet providerCard(p: ProviderInfo)}
        <li
          class="prov-row prov-card"
          data-keyed={providerEnabled(p.name) && providerCredentialed(p.name)}
          data-missing={!providerCredentialed(p.name)}
          data-open={expandedProvider === p.name}
        >
          <div class="prov-top">
            <button
              type="button"
              class="prov-reveal"
              aria-expanded={expandedProvider === p.name}
              onclick={() =>
                (expandedProvider = expandedProvider === p.name ? null : p.name)}
            >
              <strong>{providerLabel(p.name)}</strong>
              <span class="prov-meta">
                {providerKind(p.name)} · {providerStatus(p.name)}
                · {p.models?.length ?? 0} models
              </span>
            </button>
            <label class="mini-toggle">
              <input
                type="checkbox"
                checked={providerEnabled(p.name)}
                disabled={!settings.config}
                onchange={(e) =>
                  void toggleProvider(p.name, (e.currentTarget as HTMLInputElement).checked)}
              />
              <span>On</span>
            </label>
          </div>
          {#if expandedProvider === p.name}
            {#if (p.models?.length ?? 0) > 0}
              <label class="prov-field">
                <span>Default model</span>
                <select
                  aria-label={`Default model for ${providerLabel(p.name)}`}
                  value={providerDefaultModel(p.name) || p.models![0]!.id}
                  disabled={!settings.config}
                  onchange={(e) =>
                    void setProviderDefaultModel(
                      p.name,
                      (e.currentTarget as HTMLSelectElement).value
                    )}
                >
                  {#each p.models ?? [] as m (m.id)}
                    <option value={m.id}>{m.id}</option>
                  {/each}
                </select>
              </label>
            {/if}
            {#if providerShowsBaseURL(p.name)}
              <label class="prov-field">
                <span>Base URL</span>
                <input
                  type="url"
                  placeholder={p.name === 'ollama'
                    ? 'http://127.0.0.1:11434'
                    : 'http://127.0.0.1:…'}
                  bind:value={baseUrlDraft[p.name]}
                  disabled={!settings.config}
                  onblur={() => void saveBaseURL(p.name)}
                />
              </label>
            {/if}
          {/if}
        </li>
      {/snippet}

      <section class="plate models-hero" id="md-models" data-state={askModelStatus.state}>
        <header class="plate-head">
          <div class="hero-top">
            <h2>Ask model</h2>
            <span class="status-chip" data-state={askModelStatus.state}>{askModelStatus.label}</span>
          </div>
          <p class="hint">
            This is what powers Ask. Local models need no key. Cloud models need a key and count
            toward your daily spend. Keys stay on this Mac.
          </p>
        </header>
        {#if !settings.config}
          <p class="muted">Connect Condura to choose a model.</p>
        {:else if askDefaultOptions.length === 0}
          <p class="muted">No model catalog yet — turn on a provider below, then pick a model here.</p>
        {:else}
          <label class="ask-default">
            <span class="ask-default-label">Used in Ask</span>
            <select
              aria-label="Model used in Ask"
              value={askDefault}
              onchange={(e) =>
                void setAskDefaultModel((e.currentTarget as HTMLSelectElement).value)}
            >
              {#each askDefaultOptions as opt (opt.value)}
                <option value={opt.value}>{opt.label}</option>
              {/each}
            </select>
          </label>
          {#if askModelStatus.state !== 'ok'}
            <p class="note warn">
              {#if askModelStatus.state === 'halt' && askModelStatus.label.includes('key')}
                Paste an API key in the Keys section, then this model will light up in Ask.
              {:else if askModelStatus.label === 'Provider off'}
                Turn the provider On below so Ask can use this model.
              {:else}
                Finish setup below — Ask will stay blocked until a model is ready.
              {/if}
            </p>
          {/if}
        {/if}
      </section>

      <section class="plate">
        <header class="plate-head">
          <h2>Providers</h2>
          <p class="hint">Turn one on. Expand a row to pick its default model or local address.</p>
        </header>
        {#if providers.length === 0}
          <p class="muted">No providers yet. Connect Condura to load them.</p>
        {:else}
          {#if localProviders.length}
            <p class="group-cite">On this Mac</p>
            <ul class="prov-list">
              {#each localProviders as p (p.name)}
                {@render providerCard(p)}
              {/each}
            </ul>
          {/if}
          {#if cloudProviders.length}
            <p class="group-cite">Cloud</p>
            <ul class="prov-list">
              {#each cloudProviders as p (p.name)}
                {@render providerCard(p)}
              {/each}
            </ul>
          {/if}
        {/if}
      </section>

      <section class="plate" id="md-keys">
        <header class="plate-head">
          <h2>API keys</h2>
          <p class="hint">Stored only on this Mac. Never sent to Condura’s servers.</p>
        </header>
        {#if providers.filter((p) => providerNeedsKey(p.name)).length === 0}
          <p class="muted">No cloud providers in the catalog.</p>
        {:else}
          <div class="key-form">
            <select bind:value={keyProvider} aria-label="Provider for key">
              {#each providers.filter((p) => providerNeedsKey(p.name)) as p (p.name)}
                <option value={p.name}>{providerLabel(p.name)}</option>
              {/each}
            </select>
            <input
              type="password"
              placeholder="Paste API key"
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
                  <span>{providerLabel(k.provider)} · {k.label}</span>
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
            <p class="note" class:warn={keyNote.toLowerCase().includes('before') || keyNote.toLowerCase().includes('error')}>{keyNote}</p>
          {/if}
        {/if}
      </section>
    {:else}
      <section class="plate">
        <header class="plate-head">
          <h2>What Condura remembers</h2>
          <p class="hint">Optional learning from how you work. You can forget anything.</p>
        </header>
        <div class="seg strength" role="group" aria-label="Learning strength">
          {#each STRENGTH_OPTS as opt (opt.id)}
            <button type="button" class:on={strength === opt.id} onclick={() => void setStrength(opt.id)}>
              {opt.label}
            </button>
          {/each}
        </div>
        {#if adaptiveLoading}
          <p class="muted">Loading…</p>
        {:else if adaptiveItems.length === 0}
          <p class="muted">Nothing learned yet. It will appear here as you use Condura.</p>
        {:else}
          <ul class="adapt-list">
            {#each adaptiveItems as item (item.field + item.value)}
              <li>
                <div>
                  <strong>{item.value}</strong>
                  <span class="adapt-ev">{item.evidence}</span>
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
          <button type="button" class="md-btn md-btn-ghost" disabled={!adaptiveProfile} onclick={exportAdaptive}>
            Export
          </button>
          {#if confirmResetAdaptive}
            <button type="button" class="md-btn md-btn-danger" disabled={adaptiveBusy} onclick={() => void resetAdaptive()}>
              Yes, reset all
            </button>
            <button type="button" class="md-btn md-btn-ghost" onclick={() => (confirmResetAdaptive = false)}>
              Cancel
            </button>
          {:else}
            <button type="button" class="md-btn md-btn-danger" disabled={adaptiveBusy} onclick={() => (confirmResetAdaptive = true)}>
              Reset learning
            </button>
          {/if}
        </div>
      </section>

      <section class="plate">
        <header class="plate-head">
          <h2>Backup</h2>
          <p class="hint">Save a copy of your Condura data on this computer. Folder: {backupDir}</p>
        </header>
        <div class="row-actions">
          <button type="button" class="md-btn md-btn-primary" disabled={backupBusy} onclick={() => void createBackup()}>
            {backupBusy ? 'Creating…' : 'Create backup'}
          </button>
          <button
            type="button"
            class="md-btn md-btn-ghost"
            disabled={restoreBusy || trust.backups.length === 0}
            onclick={() => (confirmRestorePath = trust.backups[0]?.path ?? null)}
          >
            Restore latest
          </button>
        </div>
        {#if confirmRestorePath}
          <div class="confirm">
            <p>Restore from this backup? Current local data will be replaced.</p>
            <p class="path">{confirmRestorePath}</p>
            <div class="row-actions">
              <button type="button" class="md-btn md-btn-danger" disabled={restoreBusy} onclick={() => void performRestore(confirmRestorePath!)}>
                {restoreBusy ? 'Restoring…' : 'Restore now'}
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
                  <span class="adapt-meta">{formatBytes(b.size)}</span>
                </div>
                <button type="button" class="md-btn md-btn-ghost" disabled={restoreBusy} onclick={() => (confirmRestorePath = b.path)}>
                  Restore
                </button>
              </li>
            {/each}
          </ul>
        {:else}
          <p class="muted">No backups yet.</p>
        {/if}
        {#if backupNote}
          <p class="note">{backupNote}</p>
        {/if}
      </section>

      <section class="plate">
        <button
          type="button"
          class="adv-toggle"
          aria-expanded={showSafetyDetail}
          onclick={() => (showSafetyDetail = !showSafetyDetail)}
        >
          {showSafetyDetail ? 'Hide' : 'Show'} safety system status
        </button>
        {#if showSafetyDetail}
          {#if capsError}
            <p class="muted">{capsError}</p>
          {:else if !capabilities}
            <p class="muted">Loading…</p>
          {:else}
            <div class="trust-grid">
              <div class="trust-card" data-on={capabilities.kill_switch.layer1_hotkey}>
                <strong>Emergency stop</strong>
                <span>{capabilities.kill_switch.layer1_hotkey ? 'Ready' : 'Not ready'}</span>
              </div>
              <div class="trust-card" data-on={capabilities.kill_switch.layer2_watchdog}>
                <strong>Watchdog</strong>
                <span>{capabilities.kill_switch.layer2_watchdog ? 'On' : 'Off'}</span>
              </div>
              <div class="trust-card" data-on={capabilities.audit.hmac_subkey}>
                <strong>Audit log</strong>
                <span>{capabilities.audit.hmac_subkey ? 'Sealed' : 'Basic'}</span>
              </div>
              <div class="trust-card" data-on={capabilities.kill_switch.layer3_network_isolation.os_process || capabilities.kill_switch.layer3_network_isolation.in_process}>
                <strong>Network guard</strong>
                <span>
                  {capabilities.kill_switch.layer3_network_isolation.os_process
                    ? 'Strong'
                    : capabilities.kill_switch.layer3_network_isolation.in_process
                      ? 'Soft'
                      : 'Not active'}
                </span>
              </div>
            </div>
          {/if}
          <button type="button" class="md-btn md-btn-ghost" onclick={() => void loadCapabilities()}>Refresh</button>
        {/if}
      </section>

      <section class="plate setup">
        <header class="plate-head">
          <h2>Welcome setup</h2>
          <p class="hint">Walk through the first-run screens again. Your chats and keys stay put.</p>
        </header>
        {#if confirmRerunSetup}
          <div class="row-actions">
            <button
              type="button"
              class="md-btn md-btn-primary"
              disabled={rerunBusy}
              onclick={() => void confirmAndRerunSetup()}
            >
              {rerunBusy ? 'Opening…' : 'Start setup'}
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

      <div class="links">
        <p class="links-k">Related</p>
        <div class="link-row">
          <button type="button" class="md-btn md-btn-ghost" onclick={() => go('#/account')}>Account</button>
          <button type="button" class="md-btn md-btn-ghost" onclick={() => go('#/sync')}>Sync</button>
          <button type="button" class="md-btn md-btn-ghost" onclick={() => go('#/audit')}>Audit log</button>
          <button type="button" class="md-btn md-btn-ghost" onclick={() => go('#/about')}>About</button>
        </div>
      </div>
    {/if}
  </div>
</MeridianPage>

<style>
  .desk { display: grid; gap: 14px; }
  .status-bar {
    display: flex; flex-wrap: wrap; align-items: center; gap: 8px;
    margin: 0; padding: 10px 14px; border-radius: 12px;
    border: 1px solid var(--md-line); background: color-mix(in oklab, var(--md-surface) 85%, transparent);
    font-size: 13px; color: var(--md-ink-mute);
  }
  .status-bar.ok {
    border-color: color-mix(in oklab, var(--md-live) 28%, transparent);
    background: color-mix(in oklab, var(--md-live) 6%, var(--md-surface));
    color: var(--md-live);
  }
  .status-bar.warn { border-color: color-mix(in oklab, var(--md-halt) 22%, var(--md-line)); }
  .status-bar .dot {
    width: 7px; height: 7px; border-radius: 50%; background: var(--md-ink-faint); flex: none;
  }
  .status-bar.ok .dot {
    background: var(--md-live);
    box-shadow: none;
  }
  .sep { opacity: 0.5; }

  .tabs {
    display: flex; flex-wrap: wrap; gap: 2px; padding: 3px;
    border-radius: 10px; background: var(--md-stage); border: 1px solid var(--md-line);
    position: sticky; top: 0; z-index: 5;
  }
  .tab {
    appearance: none; border: 0; background: transparent; cursor: pointer;
    display: inline-flex; align-items: center; gap: 6px;
    padding: 8px 12px; border-radius: 8px; font-size: 13px; font-weight: 550;
    color: var(--md-ink-mute); transition: background 140ms var(--md-ease), color 140ms var(--md-ease);
  }
  .tab:hover { color: var(--md-ink); background: color-mix(in oklab, var(--md-ink) 3%, transparent); }
  .tab.on {
    background: var(--md-surface); color: var(--md-ink);
    box-shadow: none;
    border: 1px solid var(--md-line);
  }
  .tab.need { color: var(--md-ink); }
  .tab:focus-visible { outline: none; box-shadow: var(--md-focus); }
  .tab-dot {
    width: 6px; height: 6px; border-radius: 50%; flex: none;
    background: var(--md-halt);
    box-shadow: none;
  }
  .tab-kbd {
    font-family: var(--md-font-mono);
    font-size: 9.5px;
    font-weight: 500;
    letter-spacing: 0.02em;
    padding: 1px 5px;
    border-radius: 5px;
    background: color-mix(in oklab, var(--md-stage) 80%, transparent);
    border: 1px solid var(--md-line);
    color: var(--md-ink-faint);
    margin-left: 4px;
    transition: color 140ms var(--md-ease), border-color 140ms var(--md-ease);
  }
  .tab:hover .tab-kbd,
  .tab.on .tab-kbd {
    color: var(--md-cobalt);
    border-color: color-mix(in oklab, var(--md-cobalt) 28%, var(--md-line));
  }
  @media (max-width: 720px) {
    /* On narrow screens the kbd hint crowds the label — drop it. */
    .tab-kbd { display: none; }
  }

  .plate {
    border-radius: 12px; border: 1px solid var(--md-line);
    background: var(--md-surface);
    padding: 18px 20px 20px;
  }
  #md-models, #md-spend { scroll-margin-top: 64px; }
  .plate-head { margin-bottom: 14px; }
  .hero-top {
    display: flex; flex-wrap: wrap; align-items: center; gap: 8px 12px; margin-bottom: 4px;
  }
  .hero-top h2 { margin: 0; }
  .models-hero {
    border-color: color-mix(in oklab, var(--md-cobalt) 22%, var(--md-line));
    background: var(--md-surface);
  }
  .models-hero[data-state='ok'] {
    border-color: color-mix(in oklab, var(--md-live) 32%, var(--md-line));
  }
  .models-hero[data-state='warn'],
  .models-hero[data-state='halt'] {
    border-color: color-mix(in oklab, var(--md-halt) 28%, var(--md-line));
  }
  .status-chip {
    display: inline-flex; align-items: center;
    padding: 3px 8px; border-radius: 6px;
    border: 1px solid var(--md-line); background: var(--md-stage);
    font-size: 10px; font-weight: 650; letter-spacing: 0.05em; text-transform: uppercase;
    color: var(--md-ink-mute);
  }
  .status-chip[data-state='ok'] {
    border-color: color-mix(in oklab, var(--md-live) 40%, transparent);
    background: color-mix(in oklab, var(--md-live) 10%, var(--md-surface));
    color: var(--md-live);
  }
  .status-chip[data-state='warn'],
  .status-chip[data-state='halt'] {
    border-color: color-mix(in oklab, var(--md-halt) 36%, transparent);
    background: color-mix(in oklab, var(--md-halt) 9%, var(--md-surface));
    color: var(--md-halt);
  }
  .group-cite {
    margin: 12px 0 6px; font-size: 11px; font-weight: 700;
    letter-spacing: 0.06em; text-transform: uppercase; color: var(--md-ink-faint);
  }
  .group-cite:first-of-type { margin-top: 0; }
  h2 {
    font-family: var(--md-font-display); font-size: 18px; letter-spacing: -0.03em; margin: 0 0 4px;
  }
  .hint { margin: 0; font-size: 13px; color: var(--md-ink-faint); line-height: 1.45; max-width: 52ch; }
  .hint.tight { margin-top: 0; }
  .hint.tight strong { color: var(--md-ink); font-weight: 700; }
  .muted { color: var(--md-ink-mute); font-size: 14px; margin: 0 0 12px; line-height: 1.5; }
  .note {
    margin: 12px 0 0; font-size: 12px; color: var(--md-live);
  }
  .note.warn { color: var(--md-halt); }

  .seg {
    display: inline-flex; padding: 3px; border-radius: 9px;
    background: var(--md-stage); border: 1px solid var(--md-line); flex-wrap: wrap; gap: 2px;
  }
  .seg button {
    padding: 7px 12px; border-radius: 7px; font-weight: 550; font-size: 13px;
    color: var(--md-ink-mute); cursor: pointer;
    border: 0; background: transparent;
    transition: background 140ms var(--md-ease), color 140ms var(--md-ease);
  }
  .seg button.on {
    background: var(--md-cobalt); color: #fff;
    box-shadow: none;
  }
  .seg button:focus-visible { outline: none; box-shadow: var(--md-focus); }

  .hotkey-stage { display: flex; flex-wrap: wrap; align-items: center; gap: 12px; margin-bottom: 14px; }
  .hotkey-display {
    font-family: var(--md-font-display); font-size: 24px; letter-spacing: -0.04em;
    min-width: 120px; padding: 10px 14px; border-radius: 12px;
    border: 1px solid var(--md-line-strong); background: var(--md-stage);
  }
  .recording { outline: 2px solid var(--md-cobalt); }

  .toggle-row {
    display: flex; gap: 12px; align-items: flex-start; cursor: pointer;
    padding: 12px 0 0; border-top: 1px solid var(--md-line); margin-top: 4px;
  }
  .toggle-row input { margin-top: 4px; accent-color: var(--md-cobalt); }
  .toggle-row strong { display: block; font-size: 14px; }
  .toggle-row em { display: block; font-style: normal; font-size: 12px; color: var(--md-ink-faint); margin-top: 2px; }

  .adv-toggle {
    appearance: none; border: 0; background: transparent; cursor: pointer;
    font-size: 13px; font-weight: 600; color: var(--md-cobalt); padding: 4px 0 10px;
    text-align: left;
  }
  .adv-toggle:hover { text-decoration: underline; }

  .autonomy {
    display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; margin-bottom: 8px;
  }
  .auto-card {
    text-align: left; padding: 12px 13px; border-radius: 10px; border: 1px solid var(--md-line-strong);
    background: var(--md-stage); cursor: pointer; display: grid; gap: 5px; color: inherit;
  }
  .auto-card strong { font-family: var(--md-font-display); font-size: 15px; letter-spacing: -0.03em; }
  .auto-card span:last-child { font-size: 12px; line-height: 1.4; color: var(--md-ink-mute); }
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

  .matrix { display: grid; gap: 6px; margin-top: 4px; }
  .matrix-k {
    font-size: 12px; color: var(--md-ink-faint); margin: 0 0 6px;
  }
  .matrix-row {
    display: grid; grid-template-columns: 1fr auto; gap: 12px; align-items: center;
    padding: 8px 0; border-bottom: 1px solid var(--md-line);
  }
  .matrix-label { font-size: 13px; font-weight: 600; }
  .dial { display: inline-flex; gap: 2px; padding: 2px; border-radius: 8px; background: var(--md-stage); border: 1px solid var(--md-line); }
  .dial button {
    padding: 5px 9px; border-radius: 6px; font-size: 11px; font-weight: 550;
    color: var(--md-ink-faint); cursor: pointer; border: 0; background: transparent;
  }
  .dial button.on[data-l='auto'] { background: var(--md-live); color: #fff; }
  .dial button.on[data-l='warn'] { background: var(--md-cobalt); color: #fff; }
  .dial button.on[data-l='block'] { background: var(--md-halt); color: #fff; }

  .spend {
    display: grid; grid-template-columns: 1fr 1fr; gap: 16px; align-items: end;
    padding-bottom: 12px; margin-bottom: 4px;
  }
  .spend.hot .gauge-fill { background: var(--md-halt); }
  .gauge {
    height: 4px; border-radius: 2px; background: var(--md-stage);
    border: 1px solid var(--md-line); overflow: hidden; margin-bottom: 10px;
  }
  .gauge-fill {
    height: 100%; border-radius: inherit;
    background: var(--md-cobalt);
    transition: width 280ms var(--md-ease);
  }
  .spend-row { display: flex; align-items: center; gap: 8px; }
  .currency, .unit { font-size: 12px; color: var(--md-ink-faint); }
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
    padding: 12px; border-radius: 12px; border: 1px solid var(--md-line); background: var(--md-stage);
  }
  .prov-card {
    flex-direction: column; align-items: stretch;
    transition: border-color 160ms var(--md-ease), background 160ms var(--md-ease);
  }
  .prov-row[data-missing='true'] {
    border-style: dashed; opacity: 0.94;
  }
  .prov-row[data-keyed='true'] { border-left: 3px solid var(--md-live); }
  .prov-row[data-open='true'] {
    border-color: color-mix(in oklab, var(--md-cobalt) 34%, var(--md-line));
    background: color-mix(in oklab, var(--md-surface) 70%, var(--md-stage));
  }
  .prov-top {
    display: flex; justify-content: space-between; align-items: center; gap: 12px; width: 100%;
  }
  .prov-reveal {
    flex: 1; min-width: 0; display: block; margin: 0; padding: 0; border: 0;
    background: transparent; color: inherit; text-align: left; font: inherit; cursor: pointer;
  }
  .prov-reveal:focus-visible {
    outline: none; border-radius: 8px; box-shadow: var(--md-focus);
  }
  .prov-reveal strong { display: block; font-size: 14px; }
  .prov-field {
    display: grid; gap: 4px; margin-top: 10px; font-size: 12px; font-weight: 600; color: var(--md-ink-mute);
  }
  .prov-field select, .prov-field input, .ask-default select {
    padding: 10px 12px; border-radius: 12px; border: 1px solid var(--md-line-strong);
    background: var(--md-surface); font-size: 13px; font-weight: 500; color: var(--md-ink);
  }
  .ask-default {
    display: grid; gap: 6px; margin-bottom: 8px;
  }
  .ask-default-label {
    font-size: 12px; font-weight: 700; letter-spacing: 0.02em;
    color: var(--md-ink-faint);
  }
  .prov-meta, .adapt-meta, .adapt-ev {
    display: block; font-size: 12px; color: var(--md-ink-faint); margin-top: 2px;
  }
  .adapt-ev { color: var(--md-ink-mute); }
  .mini-toggle { display: inline-flex; align-items: center; gap: 6px; font-size: 12px; font-weight: 600; }
  .mini-toggle input { accent-color: var(--md-cobalt); }
  .key-form {
    display: grid; grid-template-columns: 140px 1fr auto; gap: 8px; margin-bottom: 12px;
  }
  .key-form select, .key-form input {
    padding: 10px 12px; border-radius: 12px; border: 1px solid var(--md-line-strong);
    background: var(--md-surface); font-size: 13px;
  }
  .row-actions { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 4px; }
  .confirm {
    margin-top: 12px; padding: 12px; border-radius: 12px;
    border: 1px solid color-mix(in oklab, var(--md-halt) 30%, var(--md-line));
    background: color-mix(in oklab, var(--md-halt) 6%, var(--md-surface));
  }
  .confirm .path {
    font-family: var(--md-font-mono); font-size: 11px; color: var(--md-ink-faint);
    word-break: break-all; margin: 6px 0 10px;
  }

  .trust-grid {
    display: grid; grid-template-columns: repeat(2, 1fr); gap: 8px; margin: 10px 0 12px;
  }
  .trust-card {
    padding: 12px; border-radius: 12px; border: 1px solid var(--md-line); background: var(--md-stage);
    display: grid; gap: 4px;
  }
  .trust-card[data-on='true'] {
    border-color: color-mix(in oklab, var(--md-live) 35%, transparent);
  }
  .trust-card strong { font-size: 13px; }
  .trust-card span { font-size: 12px; color: var(--md-ink-mute); }

  .key-list { display: grid; gap: 8px; margin: 0; padding: 0; list-style: none; }
  .key-list li {
    display: flex; align-items: center; gap: 14px; padding: 8px 0;
    border-bottom: 1px solid var(--md-line);
  }
  .key-list li:last-child { border-bottom: 0; }
  .key-list kbd {
    font-family: var(--md-font-mono); font-size: 11px; min-width: 88px; text-align: center;
    padding: 6px 10px; border-radius: 8px; background: var(--md-stage);
    border: 1px solid var(--md-line-strong); color: var(--md-ink-soft);
  }
  .key-list span { font-size: 14px; color: var(--md-ink-mute); }

  .links { margin-top: 4px; }
  .links-k {
    font-size: 12px; font-weight: 600; color: var(--md-ink-faint); margin: 0 0 8px;
  }
  .link-row { display: flex; flex-wrap: wrap; gap: 8px; }

  @media (max-width: 720px) {
    .autonomy, .trust-grid, .spend, .key-form { grid-template-columns: 1fr; }
    .matrix-row { grid-template-columns: 1fr; gap: 8px; }
    .tabs { position: static; }
    .tab { flex: 1; text-align: center; padding: 9px 8px; font-size: 12px; }
  }
  @media (max-width: 420px) {
    .plate { padding: 16px 14px 18px; }
    .seg { width: 100%; }
    .seg button { flex: 1; }
  }
</style>
