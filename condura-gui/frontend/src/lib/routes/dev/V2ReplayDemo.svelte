<!--
  V2ReplayDemo — Condura v2 Replay preview.

  8 frames across a 24-hour period. Drag the strip; the preview
  updates with the moment's summary, decision, and intent.
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Sidebar, StatusBar, Replay, type SidebarItem, type ReplayFrame } from '$lib/v2'

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

  let active = $state('replay')
  let collapsed = $state(false)

  let frames = $state<ReplayFrame[]>([
    {
      id: 'f1', ts: '2026-07-01 09:14:22', hour: 9,
      summary: 'Condura started a new session: "Good morning. How can I help today?"',
      decision: 'model=ollama/qwen2.5-coder, temp=0.4, max_tokens=512, system=default',
      intent:  'await user input',
    },
    {
      id: 'f2', ts: '2026-07-01 09:14:31', hour: 9,
      summary: 'User asked: "Take a look at my calendar tomorrow."',
      intent:  'check calendar',
    },
    {
      id: 'f3', ts: '2026-07-01 09:14:33', hour: 9,
      summary: 'Condura read com.apple.Mail/Inbox unread (read · blast radius 1) and the calendar.',
      decision: 'computeruse.execute(read_macos_app, mac-cua, target="com.apple.iCal", timeout=2s) → success',
    },
    {
      id: 'f4', ts: '2026-07-01 09:14:41', hour: 9,
      summary: 'Reply drafted: "You have 3 meetings tomorrow, with a 3-hour block in the afternoon."',
      decision: 'tokens=142, model=ollama/qwen2.5-coder',
    },
    {
      id: 'f5', ts: '2026-07-01 09:15:02', hour: 9,
      summary: 'User asked: "Draft the design-review one-pager too."',
      intent:  'draft document',
    },
    {
      id: 'f6', ts: '2026-07-01 09:15:09', hour: 9,
      summary: 'Condura edited ~/Documents/notes/Atlas.md, lines 42–86 (write · blast radius 2).',
      decision: 'checkpoint created Atlas.pre-09:15:09, change_drift=+12 lines',
    },
    {
      id: 'f7', ts: '2026-07-01 11:24:11', hour: 11,
      summary: 'Design review meeting entered. Condura held state for 2h 9m and waited.',
    },
    {
      id: 'f8', ts: '2026-07-01 14:18:50', hour: 14,
      summary: 'User: "Send Alex my updated take on the v2 metrics." Condura asked permission, drafted, sent.',
      decision: 'computeruse.execute(send_macos_app, mac-cua, blast=network, target=alex@example.com) → gatekeeper:allow-once',
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
  <Sidebar {items} {active} {collapsed} onSelect={(id) => active = id} onToggle={() => collapsed = !collapsed} />

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

    <Replay {frames} />
  </div>
</div>
