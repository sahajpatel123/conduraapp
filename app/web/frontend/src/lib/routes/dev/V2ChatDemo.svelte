<!--
  V2ChatDemo — Condura v2 chat surface preview.

  Mounts ChatSurface over a fake sidebar/canvas mock so the user can
  preview the redesigned home route. Includes a small seeded
  conversation to show how messages render, plus a working composer
  (Enter sends, ⇧Enter newline). Voice mode toggle swaps the
  canvas to dark + orb.
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import {
    ChatSurface, Surface, Ink, Stack, Inline, Rule,
    type Turn,
  } from '$lib/v2'

  let voiceMode = $state(false)
  let isStreaming = $state(false)
  let streamingDelta = $state('')
  let streamingTimer: ReturnType<typeof setInterval> | null = null

  // A seeded conversation that proves the layout reads.
  let turns = $state<Turn[]>([
    {
      id: '1',
      role: 'user',
      content: 'Hey condura — can you take a look at my calendar tomorrow and tell me if I have any buffer for deep work?',
      ts: '09:14',
    },
    {
      id: '2',
      role: 'agent',
      content: `I checked your calendar. You have three meetings tomorrow: a standup at 9am, a design review at 11, and a 1:1 with Alex at 3. The block between 1pm and 3pm is open and the gap after 4pm is yours too — about three hours total.

If you want, I can move the 1:1 to Friday so you get a clean four-hour block in the afternoon. Want me to draft the reschedule note?`,
      status: 'done',
      ts: '09:14',
    },
    {
      id: '3',
      role: 'user',
      content: 'Yeah, do that. And while you\'re at it, draft me a one-pager on the design review so I walk in prepared.',
      ts: '09:15',
    },
    {
      id: '4',
      role: 'agent',
      content: `Done. Reschedule sent to Alex; she has a window at Friday 2pm if that works for you.

For the design review one-pager — what's the project? I see two in your recent files: "Atlas onboarding v2" and "Onyx governance model." I'll start with whichever you'd find more useful.`,
      status: 'done',
      ts: '09:15',
    },
  ])

  function onSend(text: string) {
    const userTurn: Turn = { id: String(turns.length + 1), role: 'user', content: text, ts: '' }
    turns = [...turns, userTurn]
    isStreaming = true
    streamingDelta = ''

    const reply = "Atlas — the agent will be making you a one-pager now. Give me a moment."
    const target = reply
    let i = 0
    if (streamingTimer) clearInterval(streamingTimer)
    streamingTimer = setInterval(() => {
      i += 2
      streamingDelta = target.slice(0, i)
      if (i >= target.length) {
        if (streamingTimer) clearInterval(streamingTimer)
        streamingTimer = null
        isStreaming = false
        turns = [...turns, {
          id: String(turns.length + 1),
          role: 'agent',
          content: target,
          status: 'done',
          ts: '',
        }]
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

  function onToggleVoice() {
    voiceMode = !voiceMode
  }
</script>

<div data-v2 style="
  min-height: 100vh;
  background: var(--v2-paper);
  display: flex;
  box-sizing: border-box;
">
  <!-- Sidebar mock -->
  <div style="
    width: 72px;
    background: var(--v2-paper-2);
    border-right: 1px solid color-mix(in srgb, var(--v2-rule) 60%, transparent);
    display: flex; flex-direction: column;
    align-items: center;
    padding: var(--v2-space-4) 0;
    gap: var(--v2-space-3);
    flex-shrink: 0;
  ">
    {#each ['Ch', 'St', 'Au', 'Co', 'De', 'Hu', 'Re', 'Sy', 'Sk', 'Ab'] as monogram, i}
      <div style="
        width: 36px; height: 36px;
        border-radius: var(--v2-radius-1);
        background: {i === 0 ? 'var(--v2-accent)' : 'transparent'};
        color: {i === 0 ? 'var(--v2-paper)' : 'var(--v2-ink-3)'};
        display: grid; place-items: center;
        font-family: var(--v2-font-display);
        font-size: var(--v2-text-12);
        letter-spacing: 0.04em;
      ">{monogram}</div>
    {/each}
  </div>

  <!-- Main: ChatSurface takes everything; the wrapper just sets height -->
  <div style="flex: 1; display: flex; flex-direction: column; min-height: 100vh;">
    <!-- Status bar mock -->
    <div style="
      height: 32px;
      background: var(--v2-paper);
      border-bottom: 1px solid color-mix(in srgb, var(--v2-rule) 60%, transparent);
      display: flex; align-items: center;
      padding: 0 var(--v2-space-8);
      gap: var(--v2-space-3);
      font-family: var(--v2-font-mono);
      font-size: var(--v2-text-12);
      color: var(--v2-ink-3);
      font-feature-settings: var(--v2-numeric-features);
      flex-shrink: 0;
    ">
      <span style="color: var(--v2-accent);">●</span>
      <span style="color: var(--v2-ink-2);">condura</span>
      <span>·</span>
      <span>{isStreaming ? 'thinking…' : 'idle'}</span>
      <span style="flex: 1"></span>
      <span>9.42s</span>
      <span>·</span>
      <span>$0.0014</span>
      <span>·</span>
      <span>3 queued</span>
      <span>·</span>
      <span style="color: var(--v2-signal-go);">●</span>
      <span>online</span>
    </div>

    <div style="flex: 1; min-height: 0;">
      <ChatSurface
        {turns}
        {isStreaming}
        {streamingDelta}
        {voiceMode}
        onSend={onSend}
        onCancel={onCancel}
        onToggleVoice={onToggleVoice}
      />
    </div>
  </div>
</div>
