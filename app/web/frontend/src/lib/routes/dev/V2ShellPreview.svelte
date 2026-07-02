<!--
  V2ShellPreview — full app shell composed of v2 primitives.

  Mounts:
    - Sidebar (with hover labels, active state, collapse toggle)
    - StatusBar (with live stopwatch, heartbeat pulse, mono typography)
    - ChatSurface (the home route)

  This is the proof that the v2 chrome hangs together as a coherent
  app — Sidebar's monogram rail, StatusBar's vital-signs strip,
  ChatSurface's paper-scroll canvas. All three use the same tokens,
  same motion, same restraint.

  Routes are mocked locally. In production this would bind to
  conversation.svelte.ts and a router store.
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import {
    Sidebar, StatusBar, ChatSurface, Surface, Ink, Stack, Inline, Rule,
    type SidebarItem, type Turn,
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

  let active = $state('chat')
  let collapsed = $state(false)

  // Chat state
  let voiceMode = $state(false)
  let isStreaming = $state(false)
  let streamingDelta = $state('')
  let streamingTimer: ReturnType<typeof setInterval> | null = null
  const taskStartedAt = $state<Date | null>(null)

  let turns = $state<Turn[]>([
    {
      id: '1', role: 'user', ts: '09:14',
      content: 'Hey condura — can you take a look at my calendar tomorrow and tell me if I have any buffer for deep work?'
    },
    {
      id: '2', role: 'agent', ts: '09:14', status: 'done',
      content: `I checked your calendar. You have three meetings tomorrow: a standup at 9am, a design review at 11, and a 1:1 with Alex at 3. The block between 1pm and 3pm is open and the gap after 4pm is yours too — about three hours total.

If you want, I can move the 1:1 to Friday so you get a clean four-hour block in the afternoon. Want me to draft the reschedule note?`
    },
    {
      id: '3', role: 'user', ts: '09:15',
      content: 'Yeah, do that. And while you\'re at it, draft me a one-pager on the design review so I walk in prepared.'
    },
    {
      id: '4', role: 'agent', ts: '09:15', status: 'done',
      content: `Done. Reschedule sent to Alex; she has a window at Friday 2pm if that works for you.

For the design review one-pager — what's the project? I see two in your recent files: "Atlas onboarding v2" and "Onyx governance model." I'll start with whichever you'd find more useful.`
    },
  ])

  function onSend(text: string) {
    turns = [...turns, { id: String(turns.length + 1), role: 'user', content: text, ts: '' }]
    isStreaming = true
    streamingDelta = ''
    const target = "Atlas — the agent is making you a one-pager now. Give me a moment while I pull the relevant context."
    let i = 0
    if (streamingTimer) clearInterval(streamingTimer)
    streamingTimer = setInterval(() => {
      i += 2
      streamingDelta = target.slice(0, i)
      if (i >= target.length) {
        if (streamingTimer) clearInterval(streamingTimer)
        streamingTimer = null
        isStreaming = false
        turns = [...turns, { id: String(turns.length + 1), role: 'agent', content: target, status: 'done', ts: '' }]
        streamingDelta = ''
      }
    }, 22)
  }

  function onCancel() {
    if (streamingTimer) clearInterval(streamingTimer)
    streamingTimer = null
    isStreaming = false
    streamingDelta = ''
  }

  function onToggleVoice() { voiceMode = !voiceMode }

  // Computed status for the StatusBar
  const currentTask = $derived(
    isStreaming ? 'streaming response…' :
    null
  )

  // When user navigates away from chat, show a placeholder.
  // (In production this would be the corresponding route component.)
  const showChat = $derived(active === 'chat')

  // Hoist the stream-start timestamp so it doesn't reset on every
  // render. Without this, the StatusBar stopwatch reads 0s forever
  // because `new Date()` is called every time `isStreaming` re-evaluates.
  let streamStart = $state<Date | null>(null)
  $effect(() => {
    if (isStreaming && !streamStart) streamStart = new Date()
    if (!isStreaming) streamStart = null
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
    onSelect={(id) => { active = id; voiceMode = false; isStreaming = false; }}
    onToggle={() => { collapsed = !collapsed }}
  />

  <div style="flex: 1; display: flex; flex-direction: column; min-height: 100vh; min-width: 0;">
    <!-- The route body — Chat, or a "coming soon" placeholder for others -->
    <div style="flex: 1; min-height: 0; overflow: hidden;">
      {#if showChat}
        <ChatSurface
          {turns}
          {isStreaming}
          {streamingDelta}
          {voiceMode}
          onSend={onSend}
          onCancel={onCancel}
          onToggleVoice={onToggleVoice}
        />
      {:else}
        <!-- Placeholder for non-chat routes — proves the chrome wraps
             real content, not just chat. -->
        <div style="
          height: 100%;
          display: grid; place-items: center;
          padding: var(--v2-space-12);
          background: var(--v2-paper);
        ">
          <Surface elevation={0} padding="12" radius="3" tone="paper" style:max-width="540px">
            <Stack gap={4} align="center">
              <Ink kind="mono-cap" tone="accent">route</Ink>
              <Ink kind="display" style:font-size="var(--v2-text-40)" style:text-transform="capitalize">{active}</Ink>
              <Ink kind="body" tone="ink-2" style:text-align="center">
                This route isn't built on v2 yet. The chrome (sidebar + status bar)
                is real — proof that the surround works. The inner surface for
                {active} is the next thing to ship.
              </Ink>
              <Inline gap={3}>
                <button
                  data-v2
                  onclick={() => { active = 'chat' }}
                  style="
                    font-family: var(--v2-font-sans);
                    font-size: var(--v2-text-14);
                    font-weight: 500;
                    color: var(--v2-paper);
                    background: var(--v2-accent);
                    border: 1px solid transparent;
                    padding: var(--v2-space-3) var(--v2-space-4);
                    border-radius: var(--v2-radius-1);
                    cursor: pointer;
                  "
                >back to chat</button>
              </Inline>
            </Stack>
          </Surface>
        </div>
      {/if}
    </div>

    <StatusBar
      agentName="condura"
      currentTask={currentTask}
      taskStartedAt={streamStart}
      queueDepth={0}
      todaySpend="$0.0014"
      online={true}
      activeModel="ollama · qwen2.5-coder"
    />
  </div>
</div>
