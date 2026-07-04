<!--
  V2AuditDemo — Condura v2 Audit preview.

  Renders a small but realistic audit-log story: a calendar buffer
  check, a Venmo payment (destructive, gatekeeper-approved), a file
  edit, and an email send. The integrity badge starts in 'unknown'
  state with a spinner; ~600ms later it flips to 'verified'. The
  destructive entry has a 4px left rule; the network a 3px; the
  write a 2px; the read a 1px.
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Sidebar, StatusBar, Audit, type SidebarItem, type AuditEntry } from '$lib/v2'

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

  let active = $state('audit')
  let collapsed = $state(false)
  let integrity = $state<'unknown' | 'verified' | 'broken'>('unknown')
  let integrityDetail = $state<string | undefined>('Walking chain from row 1…')

  let entries = $state<AuditEntry[]>([
    {
      id: '1', timestamp: '09:14:22', actor: 'condura',
      action: 'Read inbox — first 25 unread',
      blastRadius: 'read',
      hash: 'a4f392bced015728…',
      detail: 'Received user request "morning briefing". Read 25 messages from com.apple.Mail/Inbox. No outbound action.',
    },
    {
      id: '2', timestamp: '09:14:25', actor: 'gatekeeper',
      action: 'Allow send email to alex@example.com',
      blastRadius: 'network',
      hash: 'c81f9a3d6024457e…',
      detail: 'User approved network action. Blast radius: network. Policy: require_consent (consumed). Re-requested for next similar action.',
    },
    {
      id: '3', timestamp: '09:14:32', actor: 'condura',
      action: 'Edit ~/Documents/notes/Atlas.md (lines 42–86)',
      blastRadius: 'write',
      hash: '01be7c44a3f8d221…',
      detail: 'Replaced 44 lines rewriting the activation-metrics section. Created checkpoint "Atlas.pre-14:14:32" before saving.',
    },
    {
      id: '4', timestamp: '09:14:38', actor: 'condura',
      action: 'Read vault item "github-api-token"',
      blastRadius: 'read',
      hash: '6f829ab13c000a40…',
      detail: 'OnePassword read via macOS-MCP bridge. Used to query user\'s starred repos for the Onyx skill.',
    },
    {
      id: '5', timestamp: '09:15:01', actor: 'gatekeeper',
      action: 'Allow transfer $128.40 to @priya',
      blastRadius: 'destructive',
      hash: '70c2bd5e9f1046a8…',
      detail: 'User confirmed payment. Blast radius: destructive. Policy: require_presence_and_consent (consumed). User was active.',
    },
    {
      id: '6', timestamp: '09:15:03', actor: 'condura',
      action: 'Authorize payment via Venmo',
      blastRadius: 'destructive',
      hash: '70c2bd5e9f1046b1…',
      detail: 'Authorized payment of $128.40 USD to venmo.com / @priya via OAuth bridge. Notification sent to venmo://receipts.',
    },
    {
      id: '7', timestamp: '09:16:44', actor: 'system',
      action: 'Snapshot at /tmp/condura-snapshot-2026-07-01-091644.zip',
      blastRadius: 'write',
      hash: '119ad04d70c2bd5e…',
      detail: 'Scheduled auto-backup. 14.2 MB. Includes memory, skills, recent audit (last 24h).',
    },
  ])

  // Simulate the integrity check resolving after ~600ms.
  $effect(() => {
    if (integrity !== 'unknown') return
    const id = setTimeout(() => {
      integrity = 'verified'
      integrityDetail = `${entries.length} rows · chain intact · last verified ${new Date().toLocaleTimeString()}`
    }, 600)
    return () => clearTimeout(id)
  })
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

    <Audit {entries} {integrity} {integrityDetail} />
  </div>
</div>
