<!--
  StatusBar — Condura v2 typographic vital-signs strip.

  Bottom strip, 32px tall. Mono typeface, single line. The agent's
  heartbeat lives here — a subtle 1Hz pulse on the whole strip
  while the agent is doing work.

  Props:
    agentName?:    string  — defaults to "condura"
    currentTask?:  string | null  — null = "idle"
    taskStartedAt?: Date | null   — when set, runs a stopwatch
    queueDepth?:   number
    todaySpend?:   string  — pre-formatted like "$0.0014"
    online?:       boolean
    activeModel?:  string  — e.g. "ollama · qwen2.5"
    onClick?:      () => void  — open detailed status
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Ink, Glyph } from '$lib/v2'

  let {
    agentName = 'condura',
    currentTask = null as string | null,
    taskStartedAt = null as Date | null,
    queueDepth = 0,
    todaySpend = '$0.0000',
    online = true,
    activeModel = '',
    onClick = undefined as (() => void) | undefined,
  } = $props()

  // Live ticker — fires once per second so the stopwatch stays real
  let now = $state(new Date())
  $effect(() => {
    const id = setInterval(() => { now = new Date() }, 1000)
    return () => clearInterval(id)
  })

  // Stopwatch: real elapsed time, not a fake animation
  const elapsedMs = $derived(taskStartedAt
    ? Math.max(0, now.getTime() - taskStartedAt.getTime())
    : 0
  )
  function fmtElapsed(ms: number): string {
    if (ms < 1000) return `${ms}ms`
    const total = Math.floor(ms / 1000)
    const h = Math.floor(total / 3600)
    const m = Math.floor((total % 3600) / 60)
    const s = total % 60
    if (h > 0) return `${h}h${m.toString().padStart(2, '0')}m`
    if (m > 0) return `${m}m${s.toString().padStart(2, '0')}s`
    return `${s}.${Math.floor((ms % 1000) / 100)}s`
  }
  const elapsed = $derived(fmtElapsed(elapsedMs))

  // The whole strip pulses subtly while the agent is working.
  // 1px of background alpha on/off, alternating once per second.
  const isWorking = $derived(currentTask !== null)
</script>

<footer
  data-v2
  data-working={isWorking}
  onclick={onClick}
  role={onClick ? 'button' : undefined}
  tabindex={onClick ? 0 : undefined}
  aria-label="Agent status"
  style="
    height: 32px;
    background: var(--v2-paper);
    border-top: 1px solid color-mix(in srgb, var(--v2-rule) 60%, transparent);
    padding: 0 var(--v2-space-6);
    display: flex;
    align-items: center;
    gap: var(--v2-space-3);
    font-family: var(--v2-font-mono);
    font-size: var(--v2-text-12);
    color: var(--v2-ink-3);
    font-feature-settings: var(--v2-numeric-features);
    user-select: none;
    cursor: {onClick ? 'pointer' : 'default'};
    position: relative;
    overflow: hidden;
    flex-shrink: 0;
  "
>
  <!-- The heartbeat pulse. Painted behind the content as a full-strip
       alpha overlay that toggles once per second. CSS handles it. -->
  {#if isWorking}
    <div style="
      position: absolute;
      inset: 0;
      background: rgba(193, 138, 74, 0.04);
      animation: v2-heartbeat 1s var(--v2-ease-linear) infinite;
      pointer-events: none;
    "></div>
  {/if}

  <span style="display: flex; align-items: center; gap: var(--v2-space-2); position: relative;">
    <Glyph name={isWorking ? 'dot-active' : 'dot'} size={8} />
    <span style="color: var(--v2-ink-2);">{agentName}</span>
  </span>

  <span style="color: var(--v2-rule); position: relative;">·</span>

  <span style="
    color: {isWorking ? 'var(--v2-ink)' : 'var(--v2-ink-3)'};
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    position: relative;
  ">
    {#if isWorking}
      {currentTask}<span style="color: var(--v2-accent);"> {elapsed}</span>
    {:else}
      idle
    {/if}
  </span>

  {#if activeModel}
    <span style="color: var(--v2-ink-3); position: relative;">{activeModel}</span>
    <span style="color: var(--v2-rule); position: relative;">·</span>
  {/if}

  <span style="color: var(--v2-ink-3); position: relative;">{queueDepth} queued</span>
  <span style="color: var(--v2-rule); position: relative;">·</span>

  <span style="color: var(--v2-ink-3); position: relative;">{todaySpend}</span>
  <span style="color: var(--v2-rule); position: relative;">·</span>

  <span style="
    color: {online ? 'var(--v2-signal-go)' : 'var(--v2-signal-warn)'};
    display: flex; align-items: center; gap: var(--v2-space-2);
    position: relative;
  ">
    <Glyph name="dot-active" size={8} />
    <span>{online ? 'online' : 'local only'}</span>
  </span>
</footer>
