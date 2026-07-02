<!--
  V2ChannelsDemo — Condura v2 Channels preview.

  Four channels across the reach surface: Telegram connected, Slack
  connecting, Signal disconnected, WhatsApp in error state. Each
  card shows the appropriate connection affordance.
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Sidebar, StatusBar, Channels, type SidebarItem, type Channel } from '$lib/v2'

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

  let active = $state('channels')
  let collapsed = $state(false)

  let channels = $state<Channel[]>([
    {
      id: 'telegram', name: 'Telegram', handle: 'condura_bot',
      description: 'BotFather token. Replies land in your existing chats.',
      status: 'connected', signalStrength: 4, lastSeen: '09:15', unread: 2,
    },
    {
      id: 'slack', name: 'Slack', handle: '@condura',
      description: 'OAuth via Slack app directory. Per-channel scoping supported.',
      status: 'connecting', signalStrength: 2,
    },
    {
      id: 'signal', name: 'Signal', handle: '+1 (415) 555-0162',
      description: 'Coming in v0.2.0 — needs linked-device pairing.',
      status: 'disconnected', signalStrength: 0,
    },
    {
      id: 'whatsapp', name: 'WhatsApp', handle: '+1 (415) 555-0144',
      description: 'Token expired. Re-link required.',
      status: 'error', signalStrength: 0, lastSeen: '2026-06-29 14:02',
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

    <Channels
      {channels}
      onConnect={(id) => {
        const c = channels.find(x => x.id === id)
        if (c) {
          c.status = 'connecting'
          setTimeout(() => { c.status = 'connected'; c.signalStrength = 3; c.lastSeen = 'just now' }, 1500)
        }
      }}
      onDisconnect={(id) => {
        const c = channels.find(x => x.id === id)
        if (c) { c.status = 'disconnected'; c.signalStrength = 0 }
      }}
    />
  </div>
</div>
