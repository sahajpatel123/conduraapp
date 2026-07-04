<!--
  V2SyncDemo — Condura v2 Sync preview.

  Mounts the Sync surface with a fake paired state. The connecting
  line draws on mount (cinematic) and settles into a paired state
  with a small check. Toggle "Simulate pairing" to see the line
  re-draw and settle.
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Sidebar, StatusBar, Sync, type SidebarItem } from '$lib/v2'

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

  let active = $state('sync')
  let collapsed = $state(false)
  let paired = $state(false)
  let pin = $state('482910')
  let peerName = $state<string | undefined>(undefined)
  let qrPayload = $state('')

  // Generate a minimal placeholder QR (data URL of a tiny SVG).
  // In production this would come from the qrcode lib via a `sync.qr`
  // IPC. For the demo we draw an SVG pattern that *looks* like a QR
  // grid but is not scannable — clear demo affordance.
  $effect(() => {
    const cells = 21
    const cellSize = 7
    let svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${cells * cellSize}" height="${cells * cellSize}" viewBox="0 0 ${cells * cellSize} ${cells * cellSize}">`
    svg += `<rect width="100%" height="100%" fill="#F7F4EE"/>`
    for (let y = 0; y < cells; y++) {
      for (let x = 0; x < cells; x++) {
        // Pseudo-random pattern, deterministic from x,y.
        if (((x * 31 + y * 17 + 7) % 11) < 5) {
          svg += `<rect x="${x * cellSize}" y="${y * cellSize}" width="${cellSize}" height="${cellSize}" fill="#1B1A17"/>`
        }
      }
    }
    // Three finder squares (corners)
    const finderSize = cellSize * 7
    ;[
      [0, 0], [(cells - 7) * cellSize, 0], [0, (cells - 7) * cellSize]
    ].forEach(([x, y]) => {
      svg += `<rect x="${x}" y="${y}" width="${finderSize}" height="${finderSize}" fill="#F7F4EE"/>`
      svg += `<rect x="${x}" y="${y}" width="${finderSize}" height="${finderSize}" fill="none" stroke="#1B1A17" stroke-width="7"/>`
      svg += `<rect x="${x + cellSize * 2}" y="${y + cellSize * 2}" width="${cellSize * 3}" height="${cellSize * 3}" fill="#1B1A17"/>`
    })
    svg += '</svg>'
    qrPayload = encodeURIComponent(svg)
  })

  function simulate() {
    paired = false
    peerName = undefined
    pin = '482910'
    setTimeout(() => {
      paired = true
      peerName = 'alex-iphone'
    }, 1400)
  }
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

    <Sync
      myId="alex-mbp"
      myName="MacBook Pro"
      {peerName}
      {qrPayload}
      {pin}
      ttlSeconds={90}
      {paired}
      onPinChange={() => pin = String(Math.floor(100000 + Math.random() * 900000))}
    />

    <div style="padding: 0 var(--v2-space-12) var(--v2-space-12);">
      <Surface elevation={0} padding="4" radius="2" tone="paper-2">
        <Inline gap={3} justify="between" align="center">
          <Ink kind="ui-small" tone="ink-3">demo control</Ink>
          <Button variant="ghost" size="small" onclick={simulate}>
            Simulate pairing
          </Button>
        </Inline>
      </Surface>
    </div>
  </div>
</div>
