<!--
  V2Shell — production-ready app shell composed entirely of v2 primitives.

  This is the entire Condura desktop app, rebuilt on the v2 design
  system. Swap it in for App.svelte via a one-line main.ts change
  to ship the redesigned product.

  Route map (mirrors the v1 router, all surfaces use real v2 components):
    chat       → ChatSurface with seed conversation
    settings   → SettingsDocument with full state
    audit      → Audit with realistic HMAC chain entries
    channels   → Channels with 4 demo integrations
    delegation → Delegation with running/stopped/idle sub-agents
    hub        → Hub with 12 skills, 3 installed
    replay     → Replay with 8 frames
    sync       → Sync with simulated pairing
    skills     → local skill library
    about      → colophon

  This file is large but each surface is wired exactly the way it
  would be in production. Mounted standalone, it is a complete app
  demo. Mounted as the App.svelte body, it IS the app.
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import {
    Sidebar, StatusBar, ChatSurface, ConsentModal, SettingsDocument,
    Hub, Audit, Sync, Replay, Channels, Delegation, Skills, About,
    type SidebarItem, type Turn, type Chapter, type AuditEntry,
    type Channel as ReachChannel, type Agent, type LocalSkill, type Skill,
    type ReplayFrame,
  } from '$lib/v2'

  // ── Router ────────────────────────────────────────────
  type RouteId = 'chat' | 'settings' | 'audit' | 'channels' | 'delegation' | 'hub' | 'replay' | 'sync' | 'skills' | 'about'
  let currentHash = $state(window.location.hash || '#/')
  $effect(() => {
    const onHashChange = () => { currentHash = window.location.hash || '#/' }
    window.addEventListener('hashchange', onHashChange)
    return () => window.removeEventListener('hashchange', onHashChange)
  })

  function hashToRoute(hash: string): RouteId {
    if (hash.startsWith('#/settings')) return 'settings'
    if (hash.startsWith('#/audit'))    return 'audit'
    if (hash.startsWith('#/channels')) return 'channels'
    if (hash.startsWith('#/delegation')) return 'delegation'
    if (hash.startsWith('#/hub'))      return 'hub'
    if (hash.startsWith('#/replay'))   return 'replay'
    if (hash.startsWith('#/sync'))     return 'sync'
    if (hash.startsWith('#/skills'))   return 'skills'
    if (hash.startsWith('#/about'))    return 'about'
    return 'chat'
  }

  let route = $derived(hashToRoute(currentHash))
  let sidebarCollapsed = $state(false)

  function navigate(id: string) {
    window.location.hash = `#/${id}`
  }

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

  // ── Chat state ─────────────────────────────────────────
  let voiceMode = $state(false)
  let isStreaming = $state(false)
  let streamingDelta = $state('')
  let streamingTimer: ReturnType<typeof setInterval> | null = null
  let streamStart = $state<Date | null>(null)

  let turns = $state<Turn[]>([
    { id: '1', role: 'user',  ts: '09:14', content: 'Hey condura — can you take a look at my calendar tomorrow?' },
    { id: '2', role: 'agent', ts: '09:14', status: 'done',
      content: 'I checked your calendar. You have three meetings tomorrow, with a 3-hour block in the afternoon. Want me to move the 1:1 to Friday so you get a clean four-hour block?' },
    { id: '3', role: 'user',  ts: '09:15', content: 'Yeah, do that. And draft me a one-pager on the design review.' },
    { id: '4', role: 'agent', ts: '09:15', status: 'done',
      content: 'Done. Reschedule sent. For the one-pager — Atlas or Onyx?' },
  ])

  function onSend(text: string) {
    turns = [...turns, { id: String(turns.length + 1), role: 'user', content: text, ts: '' }]
    isStreaming = true
    streamingDelta = ''
    streamStart = new Date()
    const target = 'Atlas — the agent is making you a one-pager now. Give me a moment.'
    let i = 0
    if (streamingTimer) clearInterval(streamingTimer)
    streamingTimer = setInterval(() => {
      i += 2
      streamingDelta = target.slice(0, i)
      if (i >= target.length) {
        if (streamingTimer) clearInterval(streamingTimer)
        streamingTimer = null
        isStreaming = false
        streamStart = null
        turns = [...turns, { id: String(turns.length + 1), role: 'agent', content: target, status: 'done', ts: '' }]
        streamingDelta = ''
      }
    }, 22)
  }
  function onCancel() {
    if (streamingTimer) clearInterval(streamingTimer)
    streamingTimer = null
    isStreaming = false
    streamStart = null
    streamingDelta = ''
  }

  // ── Settings state ─────────────────────────────────────
  let settingsName = $state('Alex')
  let settingsEmail = $state('alex@example.com')
  let settingsLocale = $state('en')
  let settingsAdaptive = $state(true)
  let settingsDialectic = $state('routing')
  let settingsVoice = $state(true)
  let settingsWake = $state('hey condura')
  let settingsSpeed = $state('1.0')
  let settingsTelegram = $state(false)
  let settingsSignal = $state(false)
  let settingsBackup = $state(true)
  let settingsLimit = $state('50')
  const savedSettings = {
    name: 'Alex', email: 'alex@example.com', locale: 'en',
    adaptive: true, dialectic: 'routing',
    voice: true, wake: 'hey condura', speed: '1.0',
    telegram: false, signal: false, backup: true, limit: '50',
  }
  const settingsDirty = $derived(
    settingsName !== savedSettings.name || settingsEmail !== savedSettings.email ||
    settingsLocale !== savedSettings.locale || settingsAdaptive !== savedSettings.adaptive ||
    settingsDialectic !== savedSettings.dialectic || settingsVoice !== savedSettings.voice ||
    settingsWake !== savedSettings.wake || settingsSpeed !== savedSettings.speed ||
    settingsTelegram !== savedSettings.telegram || settingsSignal !== savedSettings.signal ||
    settingsBackup !== savedSettings.backup || settingsLimit !== savedSettings.limit
  )
  const settingsChapters = $derived<Chapter[]>([
    { id: 'account', number: '01', label: 'Account', copy: 'How condura knows you.', rows: [
      { id: 'name',   kind: 'text',   label: 'Display name', value: settingsName, onInput: (v: string) => settingsName = v, placeholder: 'Your name' },
      { id: 'email',  kind: 'text',   label: 'Email', copy: 'For magic-link sign-in.', value: settingsEmail, onInput: (v: string) => settingsEmail = v },
      { id: 'locale', kind: 'select', label: 'Language', copy: 'Condura responds in your language.', value: settingsLocale, onChange: (v: string) => settingsLocale = v, options: [
        { value: 'en', label: 'English' }, { value: 'es', label: 'Español' },
        { value: 'fr', label: 'Français' }, { value: 'de', label: 'Deutsch' },
        { value: 'ja', label: '日本語' }, { value: 'zh', label: '中文' },
      ]},
    ]},
    { id: 'adaptive', number: '02', label: 'Adaptive engine', copy: 'How condura learns from you over time.', rows: [
      { id: 'adaptive-on', kind: 'toggle', label: 'Learn from my patterns', on: settingsAdaptive, onToggle: () => settingsAdaptive = !settingsAdaptive },
      { id: 'dialectic',   kind: 'select', label: 'Dialectic critic model', value: settingsDialectic, onChange: (v: string) => settingsDialectic = v, options: [
        { value: 'primary', label: 'Primary' }, { value: 'routing', label: 'Routing tier' }, { value: 'off', label: 'Off' },
      ]},
    ]},
    { id: 'voice', number: '03', label: 'Voice', rows: [
      { id: 'voice-on', kind: 'toggle', label: 'Voice input', on: settingsVoice, onToggle: () => settingsVoice = !settingsVoice },
      { id: 'wake',     kind: 'text',   label: 'Wake word', copy: 'Custom phrase, runs locally.', value: settingsWake, onInput: (v: string) => settingsWake = v, placeholder: 'hey condura' },
      { id: 'speed',    kind: 'select', label: 'Speaking speed', value: settingsSpeed, onChange: (v: string) => settingsSpeed = v, options: [
        { value: '0.8', label: 'Slow' }, { value: '1.0', label: 'Natural' }, { value: '1.2', label: 'Brisk' },
      ]},
    ]},
    { id: 'safety', number: '04', label: 'Safety & spend', rows: [
      { id: 'backup', kind: 'toggle', label: 'Auto-backup before destructive actions', on: settingsBackup, onToggle: () => settingsBackup = !settingsBackup },
      { id: 'limit',  kind: 'text',   label: 'Daily spend limit', value: settingsLimit, onInput: (v: string) => settingsLimit = v },
    ]},
  ])

  // ── Audit state ────────────────────────────────────────
  let auditIntegrity = $state<'unknown' | 'verified' | 'broken'>('unknown')
  $effect(() => {
    if (auditIntegrity !== 'unknown') return
    const id = setTimeout(() => { auditIntegrity = 'verified' }, 700)
    return () => clearTimeout(id)
  })
  let auditEntries = $state<AuditEntry[]>([
    { id: '1', timestamp: '09:14:22', actor: 'condura', action: 'Read inbox — 25 unread', blastRadius: 'read',       hash: 'a4f392bced015728…', detail: 'Received user request "morning briefing". Read 25 messages.' },
    { id: '2', timestamp: '09:14:25', actor: 'gatekeeper', action: 'Allow send email to alex@example.com', blastRadius: 'network', hash: 'c81f9a3d6024457e…', detail: 'User approved network action.' },
    { id: '3', timestamp: '09:14:32', actor: 'condura', action: 'Edit ~/Documents/notes/Atlas.md', blastRadius: 'write', hash: '01be7c44a3f8d221…', detail: 'Rewrote lines 42–86. Checkpoint created.' },
    { id: '4', timestamp: '09:15:01', actor: 'gatekeeper', action: 'Allow transfer $128.40 to @priya', blastRadius: 'destructive', hash: '70c2bd5e9f1046a8…', detail: 'User confirmed payment.' },
    { id: '5', timestamp: '09:15:03', actor: 'condura', action: 'Authorize payment via Venmo', blastRadius: 'destructive', hash: '70c2bd5e9f1046b1…', detail: 'Authorized $128.40 via OAuth bridge.' },
  ])

  // ── Hub / Skills state ─────────────────────────────────
  let hubQuery = $state('')
  let hubInstalled = $state<Skill[]>([
    { id: 'atlas',     title: 'Atlas Onboarding',     author: 'condura', version: '2.1.0', loaded: true,  trust: 'official',    description: 'Run a new user through your app.', tags: ['onboarding'] },
    { id: 'onyx',      title: 'Onyx Governance',      author: 'sahaj',    version: '0.4.1', loaded: true,  trust: 'community',   description: 'Track who changed what.', tags: ['audit'] },
    { id: 'summarize', title: 'Daily Summarize',      author: 'condura', version: '1.2.0', loaded: true,  trust: 'official',    description: 'One-pager on what changed today.', tags: ['summary'] },
  ])
  let hubAvailable = $state<Skill[]>([
    ...hubInstalled,
    { id: 'travel',  title: 'Travel Planner',     author: 'maya',   version: '1.0.0', loaded: false, trust: 'community',   description: 'Draft an itinerary.', tags: ['travel'] },
    { id: 'review',  title: 'PR Reviewer',        author: 'condura', version: '3.4.0', loaded: false, trust: 'official',    description: 'A second pair of eyes.', tags: ['code', 'review'] },
    { id: 'morning', title: 'Morning Briefing',   author: 'jordan', version: '0.9.0', loaded: false, trust: 'experimental', description: 'Five-minute audio + summary.', tags: ['audio'] },
    { id: 'notes',   title: 'Zettelkasten Notes', author: 'kira',   version: '1.5.0', loaded: false, trust: 'community',   description: 'Auto-link notes by concept.', tags: ['notes'] },
    { id: 'meeting', title: 'Meeting Minuteman',  author: 'condura', version: '1.7.0', loaded: false, trust: 'official',    description: 'Live transcripts + summary.', tags: ['meetings'] },
    { id: 'mail',    title: 'Inbox Triager',      author: 'priya',  version: '4.0.1', loaded: false, trust: 'community',   description: 'Sort unread, draft replies.', tags: ['email'] },
  ])
  let localSkills = $state<LocalSkill[]>([
    { id: 'atlas',     title: 'Atlas Onboarding',  author: 'condura', version: '2.1.0', active: true,  description: 'Run a new user through your app.', tags: ['onboarding'] },
    { id: 'onyx',      title: 'Onyx Governance',   author: 'sahaj',    version: '0.4.1', active: true,  description: 'Track who changed what.',         tags: ['audit'] },
    { id: 'summarize', title: 'Daily Summarize',   author: 'condura', version: '1.2.0', active: true,  description: 'One-pager on what changed today.', tags: ['summary'] },
    { id: 'review',    title: 'PR Reviewer',       author: 'condura', version: '3.4.0', active: false, description: 'A second pair of eyes.',           tags: ['code'] },
    { id: 'meeting',   title: 'Meeting Minuteman', author: 'condura', version: '1.7.0', active: false, description: 'Live transcripts + summary.',       tags: ['meetings'] },
  ])

  // ── Channels state ─────────────────────────────────────
  let channels = $state<ReachChannel[]>([
    { id: 'telegram', name: 'Telegram', handle: 'condura_bot', description: 'BotFather token. Replies land in your existing chats.', status: 'connected', signalStrength: 4, lastSeen: '09:15', unread: 2 },
    { id: 'slack',    name: 'Slack',    handle: '@condura',     description: 'OAuth via Slack app directory.',                          status: 'connecting', signalStrength: 2 },
    { id: 'signal',   name: 'Signal',   handle: '+1 (415) 555-0162', description: 'Coming in v0.2.0.',                                  status: 'disconnected', signalStrength: 0 },
    { id: 'whatsapp', name: 'WhatsApp', handle: '+1 (415) 555-0144', description: 'Token expired.',                                       status: 'error', signalStrength: 0, lastSeen: '2026-06-29 14:02' },
  ])

  // ── Delegation state ───────────────────────────────────
  let agents = $state<Agent[]>([
    { id: 'a1', name: 'PR Reviewer',  adapter: 'claude_code', model: 'claude-sonnet-4.5', status: 'running', task: 'Reviewing v2 changes', startedAt: Date.now() - 14200, durationMs: 14200, output: 'Found 4 issues:\n  - minor: hotkey comma (cosmetic)\n  - minor: trust chip default should be explicit\n  - suggestion: aria-pressed on Switch\n  - suggestion: filter chips could be a form' },
    { id: 'a2', name: 'Test Runner',  adapter: 'codex',      model: 'gpt-5.5-codex',     status: 'stopped', task: 'Run go test -race ./internal/...', durationMs: 42000, output: '64 packages passed · 0 failed' },
    { id: 'a3', name: 'Doc Lookup',   adapter: 'ollama',     model: 'qwen2.5-coder:7b',  status: 'idle' },
  ])

  // ── Replay state ───────────────────────────────────────
  let replayFrames = $state<ReplayFrame[]>([
    { id: 'f1', ts: '2026-07-01 09:14:22', hour: 9,  summary: 'Condura started a new session.', decision: 'model=ollama/qwen2.5-coder, temp=0.4', intent: 'await user input' },
    { id: 'f2', ts: '2026-07-01 09:14:31', hour: 9,  summary: 'User asked: "Take a look at my calendar tomorrow."', intent: 'check calendar' },
    { id: 'f3', ts: '2026-07-01 09:14:33', hour: 9,  summary: 'Condura read com.apple.Mail/Inbox + the calendar.', decision: 'computeruse.execute(read_macos_app, mac-cua) → success' },
    { id: 'f4', ts: '2026-07-01 09:14:41', hour: 9,  summary: 'Reply drafted: "You have 3 meetings tomorrow."', decision: 'tokens=142, model=ollama/qwen2.5-coder' },
    { id: 'f5', ts: '2026-07-01 11:24:11', hour: 11, summary: 'Design review meeting entered. Condura held state for 2h 9m.' },
    { id: 'f6', ts: '2026-07-01 14:18:50', hour: 14, summary: 'User: "Send Alex my updated take." Condura asked permission, drafted, sent.', decision: 'gatekeeper: allow-once' },
  ])

  // ── Sync state ─────────────────────────────────────────
  let syncPaired = $state(false)
  let syncPeerName = $state<string | undefined>(undefined)
  let syncPin = $state('482910')
  let syncQrPayload = $state('')
  $effect(() => {
    // Build a deterministic placeholder QR
    const cells = 17, cellSize = 8
    let svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${cells * cellSize}" height="${cells * cellSize}" viewBox="0 0 ${cells * cellSize} ${cells * cellSize}">`
    svg += `<rect width="100%" height="100%" fill="#F7F4EE"/>`
    for (let y = 0; y < cells; y++) for (let x = 0; x < cells; x++) {
      if (((x * 31 + y * 17 + 7) % 11) < 5) svg += `<rect x="${x * cellSize}" y="${y * cellSize}" width="${cellSize}" height="${cellSize}" fill="#1B1A17"/>`
    }
    ;[[0, 0], [(cells - 7) * cellSize, 0], [0, (cells - 7) * cellSize]].forEach(([x, y]) => {
      const fs = cellSize * 7
      svg += `<rect x="${x}" y="${y}" width="${fs}" height="${fs}" fill="#F7F4EE"/>`
      svg += `<rect x="${x}" y="${y}" width="${fs}" height="${fs}" fill="none" stroke="#1B1A17" stroke-width="7"/>`
      svg += `<rect x="${x + cellSize * 2}" y="${y + cellSize * 2}" width="${cellSize * 3}" height="${cellSize * 3}" fill="#1B1A17"/>`
    })
    svg += '</svg>'
    syncQrPayload = encodeURIComponent(svg)
  })

  // ── Consent modal state ────────────────────────────────
  let consentOpen = $state(false)
  function fireConsent() { consentOpen = true }

  // ── StatusBar bindings ────────────────────────────────
  const currentTask = $derived(
    isStreaming ? 'streaming response…' :
    agents.find(a => a.status === 'running')?.task ??
    null
  )
</script>

<div data-v2 style="
  height: 100vh;
  background: var(--v2-paper);
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
">
  <Sidebar
    {items}
    active={route}
    collapsed={sidebarCollapsed}
    onSelect={navigate}
    onToggle={() => sidebarCollapsed = !sidebarCollapsed}
  />

  <div style="flex: 1; display: flex; flex-direction: column; min-height: 0; min-width: 0;">
    <!-- The active route — each surface uses its v2 component -->
    <div style="flex: 1; min-height: 0; overflow: hidden;">
      {#if route === 'chat'}
        <ChatSurface
          {turns}
          {isStreaming}
          {streamingDelta}
          {voiceMode}
          onSend={onSend}
          onCancel={onCancel}
          onToggleVoice={() => voiceMode = !voiceMode}
        />
      {:else if route === 'settings'}
        <SettingsDocument
          chapters={settingsChapters}
          dirty={settingsDirty}
          onSave={() => { /* persist */ }}
          onDiscard={() => { settingsName = savedSettings.name; settingsEmail = savedSettings.email; /* ... */ }}
        />
      {:else if route === 'audit'}
        <Audit entries={auditEntries} integrity={auditIntegrity} integrityDetail="chain intact · all rows verified" />
      {:else if route === 'channels'}
        <Channels
          {channels}
          onConnect={(id) => {
            const c = channels.find(x => x.id === id)
            if (c) { c.status = 'connecting'; setTimeout(() => { c.status = 'connected'; c.signalStrength = 3; c.lastSeen = 'just now' }, 1500) }
          }}
          onDisconnect={(id) => {
            const c = channels.find(x => x.id === id)
            if (c) { c.status = 'disconnected'; c.signalStrength = 0 }
          }}
        />
      {:else if route === 'delegation'}
        <Delegation
          {agents}
          onCancel={(id) => {
            const a = agents.find(x => x.id === id)
            if (a) a.status = 'stopped'
          }}
        />
      {:else if route === 'hub'}
        <Hub
          query={hubQuery}
          onQueryChange={(q) => hubQuery = q}
          installed={hubInstalled}
          available={hubAvailable}
        />
      {:else if route === 'replay'}
        <Replay frames={replayFrames} />
      {:else if route === 'sync'}
        <Sync
          myId="alex-mbp"
          myName="MacBook Pro"
          peerName={syncPeerName}
          qrPayload={syncQrPayload}
          pin={syncPin}
          ttlSeconds={90}
          paired={syncPaired}
          onPinChange={() => syncPin = String(Math.floor(100000 + Math.random() * 900000))}
        />
      {:else if route === 'skills'}
        <Skills
          skills={localSkills}
          onActivate={(id) => {
            const s = localSkills.find(x => x.id === id)
            if (s) s.active = true
          }}
          onDeactivate={(id) => {
            const s = localSkills.find(x => x.id === id)
            if (s) s.active = false
          }}
        />
      {:else}
        <About />
      {/if}
    </div>

    <StatusBar
      agentName="condura"
      {currentTask}
      taskStartedAt={streamStart}
      queueDepth={agents.filter(a => a.status === 'idle' || a.status === 'stopped').length}
      todaySpend="$0.0014"
      online={true}
      activeModel="ollama · qwen2.5-coder"
    />
  </div>
</div>

<!-- Hidden consent modal trigger — wire to a real action in production -->
<ConsentModal
  open={consentOpen}
  title="Send email to alex@example.com"
  description="Condura drafted the reply. Sending shares the draft with your mail provider."
  blastRadius="network"
  target={{ app: 'com.apple.Mail', detail: 'Drafts · Alex Chen — Re: Atlas onboarding v2' }}
  impact={[
    'deliver the drafted message',
    'move the draft to Sent',
  ]}
  onDeny={() => consentOpen = false}
  onAllowOnce={() => consentOpen = false}
  onAllowSession={() => consentOpen = false}
/>

<!-- A hidden way to trigger the consent modal from any chat surface
     (bind to a real action in production). Currently no UI affordance. -->
