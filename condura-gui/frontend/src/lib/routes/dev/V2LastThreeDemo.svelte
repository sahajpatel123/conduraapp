<!--
  V2LastThreeDemo — Condura v2 final-batch preview route.

  Mounts Delegation + Skills + About side-by-side via a tab strip,
  since each is a short surface and they don't need a full app shell.
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Sidebar, StatusBar, Delegation, Skills, About, type SidebarItem, type Agent, type LocalSkill } from '$lib/v2'

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

  let active = $state<'delegation' | 'skills' | 'about'>('delegation')
  let collapsed = $state(false)

  let agents = $state<Agent[]>([
    {
      id: 'a1', name: 'PR Reviewer',
      adapter: 'claude_code', model: 'claude-sonnet-4.5',
      status: 'running', task: 'Reviewing changes in app/web/frontend/src/lib/v2/',
      startedAt: Date.now() - 14200, durationMs: 14200,
      output: `Found 4 issues in v2/HUb.svelte:\n  - minor: hotkey comma after last skill (cosmetic)\n  - minor: trust chip defaults to 'community' should be explicit\n  - suggestion: aria-pressed on Switch for screen readers\n  - suggestion: filter chips could be a form`}
  },
    {
      id: 'a2', name: 'Test Runner',
      adapter: 'codex', model: 'gpt-5.5-codex',
      status: 'stopped', task: 'Run go test -race ./internal/...',
      durationMs: 42000,
      output: `64 packages passed · 0 failed · 3.2s real · 12.4s wall`,
    },
    {
      id: 'a3', name: 'Doc Lookup',
      adapter: 'ollama', model: 'qwen2.5-coder:7b',
      status: 'idle',
    },
  ])

  let skills = $state<LocalSkill[]>([
    {
      id: 'atlas', title: 'Atlas Onboarding', author: 'condura',
      description: 'Run a new user through your app in five steps.',
      version: '2.1.0', active: true,
      tags: ['onboarding', 'tutorial'],
    },
    {
      id: 'onyx', title: 'Onyx Governance', author: 'sahaj',
      description: 'Track who changed what in your repo, with line-level attribution.',
      version: '0.4.1', active: true,
      tags: ['audit', 'git'],
    },
    {
      id: 'summarize', title: 'Daily Summarize', author: 'condura',
      description: 'One-pager on what changed in your world today.',
      version: '1.2.0', active: true,
      tags: ['summary', 'daily'],
    },
    {
      id: 'review', title: 'PR Reviewer', author: 'condura',
      description: 'A second pair of eyes that knows your coding style.',
      version: '3.4.0', active: false,
      tags: ['code', 'review'],
    },
    {
      id: 'meeting', title: 'Meeting Minuteman', author: 'condura',
      description: 'Live transcripts + a one-page summary at the end.',
      version: '1.7.0', active: false,
      tags: ['meetings', 'audio'],
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
      currentTask={agents.find(a => a.status === 'running')?.task ?? null}
      taskStartedAt={agents.find(a => a.status === 'running')?.startedAt ?? null}
      queueDepth={agents.filter(a => a.status === 'idle' || a.status === 'stopped').length}
      todaySpend="$0.0014"
      online={true}
      activeModel="ollama · qwen2.5-coder"
    />

    {#if active === 'delegation'}
      <Delegation {agents} onCancel={(id) => {
        const a = agents.find(x => x.id === id)
        if (a) { a.status = 'stopped' }
      }} />
    {:else if active === 'skills'}
      <Skills {skills} onActivate={(id) => {
        const s = skills.find(x => x.id === id)
        if (s) s.active = true
      }} onDeactivate={(id) => {
        const s = skills.find(x => x.id === id)
        if (s) s.active = false
      }} />
    {:else}
      <About />
    {/if}
  </div>
</div>
