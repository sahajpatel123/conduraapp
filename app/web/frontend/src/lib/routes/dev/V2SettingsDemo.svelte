<!--
  V2SettingsDemo — Condura v2 settings surface preview.

  Composes SettingsDocument over the Sidebar + StatusBar chrome.
  Demonstrates the document-style chapter layout, hardware-honest
  Switch, dirty-state "Save changes" affordance, and the chapter
  numbering / horizontal-rule rhythm.
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import {
    Sidebar, StatusBar, SettingsDocument,
    type SidebarItem, type Chapter,
  } from '$lib/v2'

  const items: SidebarItem[] = [
    { id: 'chat',       monogram: 'Ch', label: 'Chat' },
    { id: 'settings',   monogram: 'St', label: 'Settings' },
    { id: 'audit',      monogram: 'Au', label: 'Audit' },
    { id: 'channels',   monogram: 'Co', label: 'Channels' },
    { id: 'delegation', monogram: 'De', label: 'Delegation' },
    { id: 'hub',        monogram: 'Hu', label: 'Hub' },
    { id: 'replay',     monogram: 'Re', label: 'Replay' },
    { id: 'sync',       monogram: 'Sy', label: 'Sync' },
    { id: 'skills',     monogram: 'Sk', label: 'Skills' },
    { id: 'about',      monogram: 'Ab', label: 'About' },
  ]

  let active = $state('settings')
  let collapsed = $state(false)

  // State — every row's state lives here
  let name = $state('Alex')
  let email = $state('alex@example.com')
  let locale = $state('en')
  let adaptiveEnabled = $state(true)
  let dialecticCritic = $state('routing')  // 'routing' | 'primary' | 'off'
  let voiceEnabled = $state(true)
  let wakeWord = $state('hey condura')
  let voiceSpeed = $state('1.0')
  let telegramEnabled = $state(false)
  let signalEnabled = $state(false)
  let autoBackup = $state(true)
  let spendLimit = $state('50')

  // Track the "saved" snapshot so we can detect dirty state.
  let saved = $state({
    name: 'Alex', email: 'alex@example.com', locale: 'en',
    adaptiveEnabled: true, dialecticCritic: 'routing',
    voiceEnabled: true, wakeWord: 'hey condura', voiceSpeed: '1.0',
    telegramEnabled: false, signalEnabled: false,
    autoBackup: true, spendLimit: '50',
  })

  const dirty = $derived(
    name !== saved.name ||
    email !== saved.email ||
    locale !== saved.locale ||
    adaptiveEnabled !== saved.adaptiveEnabled ||
    dialecticCritic !== saved.dialecticCritic ||
    voiceEnabled !== saved.voiceEnabled ||
    wakeWord !== saved.wakeWord ||
    voiceSpeed !== saved.voiceSpeed ||
    telegramEnabled !== saved.telegramEnabled ||
    signalEnabled !== saved.signalEnabled ||
    autoBackup !== saved.autoBackup ||
    spendLimit !== saved.spendLimit
  )

  function save() {
    saved = {
      name, email, locale,
      adaptiveEnabled, dialecticCritic,
      voiceEnabled, wakeWord, voiceSpeed,
      telegramEnabled, signalEnabled,
      autoBackup, spendLimit,
    }
  }
  function discard() {
    name = saved.name
    email = saved.email
    locale = saved.locale
    adaptiveEnabled = saved.adaptiveEnabled
    dialecticCritic = saved.dialecticCritic
    voiceEnabled = saved.voiceEnabled
    wakeWord = saved.wakeWord
    voiceSpeed = saved.voiceSpeed
    telegramEnabled = saved.telegramEnabled
    signalEnabled = saved.signalEnabled
    autoBackup = saved.autoBackup
    spendLimit = saved.spendLimit
  }

  const chapters = $derived<Chapter[]>([
    {
      id: 'account', number: '01', label: 'Account',
      copy: 'How condura knows you.',
      rows: [
        { id: 'name',     kind: 'text', label: 'Display name', value: name, onInput: (v) => name = v, placeholder: 'Your name' },
        { id: 'email',    kind: 'text', label: 'Email', copy: 'For magic-link sign-in and digest emails.', value: email, onInput: (v) => email = v },
        { id: 'locale',   kind: 'select', label: 'Language', copy: 'Condura responds in your language regardless of UI locale.', value: locale, onChange: (v) => locale = v, options: [
          { value: 'en', label: 'English' },
          { value: 'es', label: 'Español' },
          { value: 'fr', label: 'Français' },
          { value: 'de', label: 'Deutsch' },
          { value: 'ja', label: '日本語' },
          { value: 'zh', label: '中文' },
        ]},
      ],
    },
    {
      id: 'adaptive', number: '02', label: 'Adaptive engine',
      copy: 'How condura learns from you over time.',
      rows: [
        { id: 'adaptive-on', kind: 'toggle', label: 'Learn from my patterns', copy: 'Condura observes your work and proposes suggestions. Off = pure stateless.', on: adaptiveEnabled, onToggle: () => adaptiveEnabled = !adaptiveEnabled },
        { id: 'dialectic-critic', kind: 'select', label: 'Dialectic critic model', copy: 'The smaller model that argues against each inferred preference.', value: dialecticCritic, onChange: (v) => dialecticCritic = v, options: [
          { value: 'primary', label: 'Primary (best quality)' },
          { value: 'routing', label: 'Routing tier (balanced)' },
          { value: 'off', label: 'Off (no critic)' },
        ]},
      ],
    },
    {
      id: 'voice', number: '03', label: 'Voice',
      copy: 'Listening + speaking.',
      rows: [
        { id: 'voice-on', kind: 'toggle', label: 'Voice input', copy: 'Press-and-hold anywhere. Whisper.cpp local by default.', on: voiceEnabled, onToggle: () => voiceEnabled = !voiceEnabled },
        { id: 'wake', kind: 'text', label: 'Wake word', copy: 'Custom phrase, runs locally via openWakeWord.', value: wakeWord, onInput: (v) => wakeWord = v, placeholder: 'hey condura' },
        { id: 'voice-speed', kind: 'select', label: 'Speaking speed', value: voiceSpeed, onChange: (v) => voiceSpeed = v, options: [
          { value: '0.8', label: 'Slow · 0.8×' },
          { value: '1.0', label: 'Natural · 1.0×' },
          { value: '1.2', label: 'Brisk · 1.2×' },
        ]},
      ],
    },
    {
      id: 'channels', number: '04', label: 'Channels',
      copy: 'Reach condura from anywhere you already talk.',
      rows: [
        { id: 'telegram', kind: 'toggle', label: 'Telegram', copy: 'Connect your BotFather token in the channels page.', on: telegramEnabled, onToggle: () => telegramEnabled = !telegramEnabled },
        { id: 'signal',   kind: 'toggle', label: 'Signal',  copy: 'Coming in v0.2.0', on: signalEnabled, onToggle: () => signalEnabled = !signalEnabled, disabled: true },
      ],
    },
    {
      id: 'safety', number: '05', label: 'Safety & spend',
      copy: 'Where the armor comes from.',
      rows: [
        { id: 'auto-backup', kind: 'toggle', label: 'Auto-backup before destructive actions', copy: 'Creates a recovery checkpoint before any destructive work.', on: autoBackup, onToggle: () => autoBackup = !autoBackup },
        { id: 'spend-limit', kind: 'text', label: 'Daily spend limit', copy: 'When reached, condura pauses and asks before any cloud call.', value: spendLimit, onInput: (v) => spendLimit = v, placeholder: '50' },
      ],
    },
  ])
</script>

<div data-v2 style="
  min-height: 100vh;
  background: var(--v2-paper);
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
">
  <Sidebar
    {items}
    {active}
    {collapsed}
    onSelect={(id) => active = id}
    onToggle={() => collapsed = !collapsed}
  />

  <div style="flex: 1; display: flex; flex-direction: column; min-height: 100vh; min-width: 0;">
    <StatusBar
      agentName="condura"
      currentTask={null}
      taskStartedAt={null}
      queueDepth={0}
      todaySpend="$0.0014"
      online={true}
      activeModel="ollama · qwen2.5-coder"
    />

    <SettingsDocument {chapters} {dirty} onSave={save} onDiscard={discard} />
  </div>
</div>
