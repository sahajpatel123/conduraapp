<!--
  ChatSurface — Condura v2 chat surface.

  A scroll, not a screen. Newest message at the bottom, paper edge
  texture at top of viewport, composer anchored bottom with a single
  line that grows. Voice mode: the whole canvas darkens to paper-2
  and a single orb breathes in the bottom-center.

  The component is pure presentation. The parent route owns data
  flow (the conversation store) and onSend/onCancel handlers.

  Props:
    turns:           array of {id, role, content, status, ts?}
    streamingDelta?: string                 — appended to the active agent turn
    isStreaming?:    boolean                — controls the caret and footer state
    onSend?:         (text: string) => void — Enter to send
    onCancel?:       () => void             — footer "Stop" while streaming
    voiceMode?:      boolean                — when true, shows the voice orb instead of chat
    onToggleVoice?:  () => void
    emptyTitle?:     string                 — empty-state heading
    emptyHint?:      string                 — empty-state body
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Surface, Ink, Stack, Inline, Rule, Button, Avatar } from '$lib/v2'

  type Role = 'user' | 'agent' | 'system'
  type Status = 'streaming' | 'paused' | 'done' | 'error'

  export interface Turn {
    id: string
    role: Role
    content: string
    status?: Status
    ts?: string
  }

  let {
    turns = [] as Turn[],
    streamingDelta = '' as string,
    isStreaming = false as boolean,
    onSend = undefined as ((text: string) => void) | undefined,
    onCancel = undefined as (() => void) | undefined,
    voiceMode = false as boolean,
    onToggleVoice = undefined as (() => void) | undefined,
    emptyTitle = 'Say hello.' as string,
    emptyHint = "Type a question, ask for help, paste an error. I'm here." as string,
  } = $props()

  // Composer state — single-line that grows
  let draft = $state('')
  let taEl = $state<HTMLTextAreaElement | null>(null)

  function autoGrow(node: HTMLTextAreaElement) {
    const fit = () => {
      node.style.height = 'auto'
      node.style.height = `${Math.min(node.scrollHeight, 240)}px`
    }
    fit()
    node.addEventListener('input', fit)
    return { destroy: () => node.removeEventListener('input', fit) }
  }

  let scrollerEl = $state<HTMLDivElement | null>(null)

  function submit() {
    const text = draft.trim()
    if (!text) return
    onSend?.(text)
    draft = ''
    requestAnimationFrame(() => {
      taEl?.focus()
      // Scroll the message list to the bottom so the user sees the
      // new turn (and the agent's incoming response) without manual
      // scrolling. Smooth behavior respects prefers-reduced-motion
      // via the global rule in motion.css (scroll-behavior: auto).
      scrollerEl?.scrollTo({
        top: scrollerEl.scrollHeight,
        behavior: 'smooth',
      })
    })
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      submit()
    }
  }

  // Voice mode is its own canvas — the CSS keyframe handles the
  // single breath; no inline JS-driven shadow mutation. The spec
  // is uncompromising: one breath, not two.
</script>

<!--
  Outer container. Voice mode swaps the inner content via {#if}.
  The paper-2 background means the voice canvas reads as one
  continuous space, not a separate panel.
-->
<div data-v2 style="
  height: 100%;
  display: flex; flex-direction: column;
  background: {voiceMode ? 'var(--v2-paper-2)' : 'var(--v2-paper)'};
  transition: background var(--v2-dur-slow) var(--v2-ease-settle);
">
  {#if !voiceMode}
    <!-- ── Chat mode ─────────────────────────────────── -->
    <div
      bind:this={scrollerEl}
      data-v2-scroll
      role="log"
      aria-live="polite"
      aria-label="Conversation"
      style="flex: 1; overflow-y: auto; padding: var(--v2-space-8) var(--v2-space-12);"
    >
      <!-- Paper edge texture: a subtle hairline at the top, fading -->
      <div style="
        height: 1px;
        background: linear-gradient(to right, transparent, var(--v2-rule), transparent);
        margin: 0 auto var(--v2-space-6);
        max-width: 720px;
      "></div>

      <div style="max-width: 720px; margin: 0 auto;">
        {#if turns.length === 0 && !isStreaming && !streamingDelta}
          <!-- Empty state -->
          <div style="
            padding: var(--v2-space-16) 0;
            text-align: center;
          ">
            <Stack gap={4} align="center">
              <Ink kind="display" as="h1" style:max-width="480px">{emptyTitle}</Ink>
              <Ink kind="body-2" tone="ink-3" style:max-width="480px">{emptyHint}</Ink>
              <div style="height: var(--v2-space-6)"></div>
              <Inline gap={2}>
                {#each ['Summarize a doc', 'Draft an email', 'Debug a stack trace', 'Plan my week'] as prompt}
                  <button
                    data-v2
                    onclick={() => { draft = prompt; taEl?.focus() }}
                    style="
                      font-family: var(--v2-font-sans);
                      font-size: var(--v2-text-12);
                      padding: var(--v2-space-2) var(--v2-space-3);
                      border-radius: var(--v2-radius-pill);
                      border: 1px solid var(--v2-rule);
                      color: var(--v2-ink-2);
                      background: var(--v2-paper);
                      cursor: pointer;
                      transition: all var(--v2-dur-fast) var(--v2-ease-out-soft);
                    "
                  >{prompt}</button>
                {/each}
              </Inline>
            </Stack>
          </div>

        {:else}
          <!-- Message list -->
          <Stack gap={8}>
            {#each turns as t (t.id)}
              <div style="
                display: flex;
                flex-direction: {t.role === 'user' ? 'row-reverse' : 'row'};
                gap: var(--v2-space-4);
                align-items: flex-start;
              ">
                <!-- Avatar / role indicator -->
                <Avatar role={t.role} size={28} class="chat-avatar" />

                <!-- Bubble or running text -->
                <div style="flex: 1; min-width: 0;">
                  <Stack gap={2}>
                    {#if t.role === 'agent' || t.role === 'system'}
                      <Ink kind="ui-small" tone="ink-3" weight="medium">
                        {t.role === 'agent' ? 'condura' : 'system'}
                        {#if t.status === 'streaming'}
                          <span style="
                            display: inline-block;
                            width: 6px; height: 6px;
                            border-radius: var(--v2-radius-pill);
                            background: var(--v2-accent);
                            margin-left: var(--v2-space-1);
                            animation: v2-heartbeat 1s var(--v2-ease-linear) infinite;
                          "></span>
                        {/if}
                      </Ink>
                      <Surface
                        elevation={0}
                        padding="6"
                        radius="2"
                        tone="paper"
                        style:display="inline-block"
                        style:max-width="560px"
                      >
                        <Ink kind="body">
                          {t.content}{#if t.status === 'streaming' && streamingDelta}<span style="color: var(--v2-accent);">{streamingDelta}</span>{/if}
                        </Ink>
                      </Surface>
                    {:else}
                      <Surface
                        elevation={0}
                        padding="6"
                        radius="2"
                        tone="paper-2"
                        style:display="inline-block"
                        style:max-width="560px"
                      >
                        <Ink kind="body">{t.content}</Ink>
                      </Surface>
                    {/if}
                  </Stack>
                </div>
              </div>
            {/each}

            {#if isStreaming && (!turns.length || turns[turns.length - 1].role !== 'agent')}
              <div style="display: flex; gap: var(--v2-space-4); align-items: flex-start;">
                <Avatar role="agent" size={28} />
                <Surface elevation={0} padding="6" radius="2" tone="paper">
                  <Ink kind="ui-small" tone="ink-3" italic>{streamingDelta || 'thinking…'}</Ink>
                </Surface>
              </div>
            {/if}
          </Stack>
        {/if}
      </div>
    </div>

    <!-- Composer — bottom-anchored, single-line growing -->
    <div style="
      border-top: 1px solid color-mix(in srgb, var(--v2-rule) 60%, transparent);
      padding: var(--v2-space-4) var(--v2-space-12) var(--v2-space-6);
      background: var(--v2-paper);
    ">
      <div style="max-width: 720px; margin: 0 auto;">
        <Surface
          elevation={isStreaming ? 0 : 2}
          padding="4"
          radius="2"
          tone="paper"
        >
          <Stack gap={3}>
            <textarea
              bind:this={taEl}
              use:autoGrow
              bind:value={draft}
              onkeydown={onKeydown}
              placeholder={isStreaming ? 'condura is thinking…' : 'ask anything · ⏎ to send · ⇧⏎ for newline'}
              rows="1"
              disabled={isStreaming}
              style="
                width: 100%;
                font-family: var(--v2-font-sans);
                font-size: var(--v2-text-16);
                line-height: var(--v2-leading-default);
                color: var(--v2-ink);
                background: transparent;
                resize: none;
                outline: none;
                border: none;
                padding: var(--v2-space-2) 0;
              "
            ></textarea>
            <Inline gap={2} justify="between" align="center">
              <Inline gap={2}>
                <button
                  data-v2
                  onclick={onToggleVoice}
                  style="
                    all: unset;
                    cursor: pointer;
                    font-family: var(--v2-font-sans);
                    font-size: var(--v2-text-12);
                    color: var(--v2-ink-3);
                    padding: var(--v2-space-1) var(--v2-space-2);
                    border-radius: var(--v2-radius-1);
                    transition: color var(--v2-dur-fast) var(--v2-ease-out-soft);
                  "
                >voice mode →</button>
              </Inline>
              <Inline gap={2} align="center">
                <Ink kind="caption" tone="ink-3">{draft.length} chars</Ink>
                {#if isStreaming}
                  <Button variant="deny" size="small" onclick={onCancel ?? (() => {})}>Stop</Button>
                {:else}
                  <Button
                    variant="primary"
                    size="small"
                    disabled={draft.trim().length === 0}
                    onclick={submit}
                  >Send</Button>
                {/if}
              </Inline>
            </Inline>
          </Stack>
        </Surface>
      </div>
    </div>

  {:else}
    <!-- ── Voice mode ─────────────────────────────────── -->
    <div style="
      flex: 1;
      display: grid; place-items: center;
      padding: var(--v2-space-12);
    ">
      <Stack gap={8} align="center">
        <!-- The orb is a single 96px disc resting on its own shadow.
             Only the inner 24px disc breathes — slow, 4s, like a
             held breath, not a screen-saver. -->
        <div style="
          width: 96px; height: 96px;
          border-radius: var(--v2-radius-pill);
          background: var(--v2-accent);
          display: grid; place-items: center;
          box-shadow: var(--v2-shadow-2);
        ">
          <div style="
            width: 24px; height: 24px;
            border-radius: var(--v2-radius-pill);
            background: var(--v2-paper);
            animation: v2-heartbeat 4s var(--v2-ease-linear) infinite;
          "></div>
        </div>

        <Stack gap={2} align="center">
          <Ink kind="mono-cap" tone="accent">listening</Ink>
          <Ink kind="display" as="h2" style:font-size="var(--v2-text-28)">
            say it. condura hears.
          </Ink>
        </Stack>

        <button
          data-v2
          onclick={onToggleVoice}
          style="
            all: unset;
            cursor: pointer;
            font-family: var(--v2-font-sans);
            font-size: var(--v2-text-12);
            color: var(--v2-ink-3);
            padding: var(--v2-space-2) var(--v2-space-3);
          "
        >← back to text</button>
      </Stack>
    </div>

  {/if}
</div>
