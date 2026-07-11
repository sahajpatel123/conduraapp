<script lang="ts">
  /**
   * Ask — the conductor’s desk.
   * Signature: ask atlas (empty) · plan plates with gated steps (thread) · tone-aware composer.
   */
  import { onMount } from 'svelte'
  import { conversation } from '../../stores/conversation.svelte'
  import { settings } from '../../stores/settings.svelte'
  import { halt } from '../../stores/halt.svelte'
  import { daemon } from '../../stores/daemon.svelte'
  import { ipc } from '../../ipc/client'
  import type { Message, ProviderInfo, ToolCall } from '../../ipc/types'

  const STARTERS: { id: string; kicker: string; label: string; body: string }[] = [
    {
      id: 'sum',
      kicker: 'See',
      label: 'Summarize what’s on my screen',
      body: 'Read the visible surface. No clicks until you allow them.',
    },
    {
      id: 'fix',
      kicker: 'Repair',
      label: 'Find and fix the last error I hit',
      body: 'Diagnose first. Propose a fix. Wait for the gate.',
    },
    {
      id: 'plan',
      kicker: 'Chart',
      label: 'Plan the next hour of work',
      body: 'A plan you can edit — not a silent sprint.',
    },
    {
      id: 'safe',
      kicker: 'Contract',
      label: 'What can you do without asking me?',
      body: 'Name the safe band. Everything else stays locked.',
    },
  ]

  const PIPE = [
    { n: '01', t: 'Ask' },
    { n: '02', t: 'Plan' },
    { n: '03', t: 'Consent' },
    { n: '04', t: 'Act' },
    { n: '05', t: 'Audit' },
  ]

  let draft = $state('')
  let focused = $state(false)
  let ta = $state<HTMLTextAreaElement | null>(null)
  let scrollEl = $state<HTMLDivElement | null>(null)
  let selectedModel = $state('')
  let providers = $state<ProviderInfo[]>([])
  let atlasFocus = $state<string | null>(null)
  let copied = $state(false)
  let reduceMotion = $state(false)

  const modelOptions = $derived(
    providers.flatMap((p) =>
      (p.models ?? []).map((m) => ({
        value: `${p.name}:${m.id}`,
        label: `${p.name} · ${m.id}`,
      }))
    )
  )

  const hasMessages = $derived(conversation.messages.length > 0 || conversation.isStreaming)
  const halted = $derived(halt.state.halted)
  const connected = $derived(daemon.connected)
  const canSend = $derived(
    !!draft.trim() && !conversation.isStreaming && !halted && selectedModel.includes(':')
  )
  const tone = $derived<'halted' | 'thinking' | 'ready' | 'offline' | 'idle'>(
    halted
      ? 'halted'
      : conversation.isStreaming
        ? 'thinking'
        : !connected && !selectedModel
          ? 'offline'
          : canSend
            ? 'ready'
            : 'idle'
  )
  const toneLabel = $derived(
    tone === 'halted'
      ? 'Halted — line cut'
      : tone === 'thinking'
        ? 'Thinking · watch the plan'
        : tone === 'ready'
          ? 'Ready to send'
          : tone === 'offline'
            ? 'Daemon offline · model optional in preview'
            : 'Waiting for a line'
  )
  const liveNote = $derived(
    halted
      ? 'Halt is armed — nothing leaves this desk until you resume.'
      : connected
        ? selectedModel
          ? `Live · ${selectedModel.replace(':', ' · ')}`
          : 'Connected · choose a model below'
        : 'Offline · connect the daemon to stream'
  )
  const recent = $derived(conversation.conversations.slice(0, 5))

  onMount(() => {
    reduceMotion =
      typeof matchMedia !== 'undefined' && matchMedia('(prefers-reduced-motion: reduce)').matches
    void (async () => {
      await settings.refresh().catch(() => {})
      await conversation.refreshList().catch(() => {})
      try {
        const list = await ipc.providersList()
        providers = list ?? []
        if (settings.config?.llm?.providers) {
          const enabled = Object.entries(settings.config.llm.providers).find(([, p]) => p.enabled)
          if (enabled?.[1]?.default_model) selectedModel = `${enabled[0]}:${enabled[1].default_model}`
        }
        if (!selectedModel && list?.[0]) {
          const p = list[0]
          const mid = p.models?.[0]?.id || ''
          if (mid) selectedModel = `${p.name}:${mid}`
        }
      } catch {
        /* preview */
      }
      try {
        const seeded = sessionStorage.getItem('md-ask-starter')
        if (seeded) {
          sessionStorage.removeItem('md-ask-starter')
          draft = seeded
        }
      } catch {
        /* ignore */
      }
      ta?.focus()
      resize()
    })()
  })

  $effect(() => {
    conversation.messages.length
    conversation.streamingDelta
    conversation.streamingToolCalls.length
    if (!scrollEl) return
    scrollEl.scrollTop = scrollEl.scrollHeight
  })

  function resize(): void {
    if (!ta) return
    ta.style.height = 'auto'
    ta.style.height = `${Math.min(ta.scrollHeight, 160)}px`
  }

  async function send(text = draft): Promise<void> {
    const trimmed = text.trim()
    if (!trimmed || conversation.isStreaming || halted) return
    const [provider, model] = selectedModel.split(':')
    if (!provider || !model) {
      draft = trimmed
      return
    }
    draft = ''
    queueMicrotask(resize)
    await conversation.send(provider, model, trimmed)
  }

  function onKey(e: KeyboardEvent): void {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      void send()
    }
    if (e.key === 'Escape' && conversation.isStreaming) void conversation.cancel()
  }

  function onInput(): void {
    resize()
  }

  function pickStarter(label: string): void {
    draft = label
    void send(label)
  }

  function toolLabel(tc: ToolCall): string {
    return tc.function?.name || 'gated action'
  }

  function toolArgs(tc: ToolCall): string {
    const raw = tc.function?.arguments?.trim()
    if (!raw) return ''
    try {
      const parsed = JSON.parse(raw) as Record<string, unknown>
      const keys = Object.keys(parsed)
      if (!keys.length) return ''
      return keys
        .slice(0, 3)
        .map((k) => `${k}: ${String(parsed[k]).slice(0, 48)}`)
        .join(' · ')
    } catch {
      return raw.slice(0, 80)
    }
  }

  function stepsFor(msg: Message): ToolCall[] {
    return msg.tool_calls ?? []
  }

  async function clearThread(): Promise<void> {
    if (conversation.isStreaming) await conversation.cancel()
    await conversation.createNew().catch(() => {
      conversation.messages = []
    })
  }

  async function openThread(id: number): Promise<void> {
    await conversation.open(id).catch(() => {})
  }

  function goAudit(): void {
    window.location.hash = '#/audit'
  }

  async function copyLast(): Promise<void> {
    const last = [...conversation.messages].reverse().find((m) => m.role === 'assistant' && m.content)
    if (!last?.content) return
    try {
      await navigator.clipboard.writeText(last.content)
      copied = true
      setTimeout(() => (copied = false), 1600)
    } catch {
      /* ignore */
    }
  }
</script>

<section class="chat">
  <div class="feed" bind:this={scrollEl}>
    {#if !hasMessages}
      <div class="hero" class:calm={reduceMotion}>
        <header class="thesis">
          <p class="brand-line">
            <span class="word">Condura</span>
            <span class="slash" aria-hidden="true">/</span>
            <span class="edition">Ask</span>
          </p>
          <h1>Ask once.<br /><span>Watch the plan.</span></h1>
          <p class="sub">
            Free. Local. Consent before action. Condura shows what it will do — then waits for you.
          </p>
          <p class="live" class:hot={connected && !halted} class:bad={halted}>
            <span class="live-dot" aria-hidden="true"></span>
            {liveNote}
          </p>
        </header>

        <ol class="pipe" aria-label="How an ask becomes an act">
          {#each PIPE as p (p.n)}
            <li>
              <span class="n">{p.n}</span>
              <span class="t">{p.t}</span>
            </li>
          {/each}
        </ol>

        <div class="atlas">
          <div class="atlas-head">
            <p class="cite">Ways to begin</p>
            <h2>Four doors into the desk</h2>
            <p class="atlas-note">Each starter is a real ask. Pick one, or write your own below.</p>
          </div>
          <div class="atlas-grid md-stagger">
            {#each STARTERS as s (s.id)}
              <button
                type="button"
                class="door"
                class:focus={atlasFocus === s.id}
                disabled={halted}
                onmouseenter={() => (atlasFocus = s.id)}
                onfocus={() => (atlasFocus = s.id)}
                onmouseleave={() => (atlasFocus = null)}
                onblur={() => (atlasFocus = null)}
                onclick={() => pickStarter(s.label)}
              >
                <span class="door-k">{s.kicker}</span>
                <span class="door-t">{s.label}</span>
                <span class="door-b">{s.body}</span>
                <span class="door-a">
                  Begin
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                    <path
                      d="M5 12h14M13 6l6 6-6 6"
                      stroke="currentColor"
                      stroke-width="1.8"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                  </svg>
                </span>
              </button>
            {/each}
          </div>
        </div>

        {#if recent.length}
          <div class="recent">
            <p class="cite">Recent threads</p>
            <div class="recent-row">
              {#each recent as c (c.id)}
                <button type="button" class="recent-chip" onclick={() => void openThread(c.id)}>
                  {c.title || `Thread ${c.id}`}
                </button>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {:else}
      <div class="thread">
        <div class="thread-bar">
          <div>
            <p class="cite">thread · gated</p>
            <h2 class="thread-title">{conversation.currentTitle || 'Ask'}</h2>
          </div>
          <div class="thread-actions">
            <button type="button" class="md-btn md-btn-ghost tiny" onclick={() => void copyLast()} disabled={!conversation.messages.some((m) => m.role === 'assistant' && m.content)}>
              {copied ? 'Copied' : 'Copy last'}
            </button>
            <button type="button" class="md-btn md-btn-ghost tiny" onclick={goAudit}>Open audit</button>
            <button type="button" class="md-btn md-btn-ghost tiny" onclick={() => void clearThread()}>
              New ask
            </button>
          </div>
        </div>

        {#if recent.length > 1}
          <div class="rail" aria-label="Recent threads">
            {#each recent as c (c.id)}
              <button
                type="button"
                class="rail-item"
                class:on={c.id === conversation.currentID}
                onclick={() => void openThread(c.id)}
              >
                {c.title || `Thread ${c.id}`}
              </button>
            {/each}
          </div>
        {/if}

        <div class="messages">
          {#each conversation.messages as msg, i (i)}
            <article class="msg" data-role={msg.role}>
              <header>
                <span>{msg.role === 'user' ? 'You' : 'Condura'}</span>
                {#if msg.role === 'assistant'}
                  <span class="cite">
                    {stepsFor(msg).length
                      ? `${stepsFor(msg).length} gated step${stepsFor(msg).length === 1 ? '' : 's'}`
                      : 'plan · gated'}
                  </span>
                {/if}
              </header>
              {#if msg.content}
                <div class="bubble">{msg.content}</div>
              {/if}
              {#if stepsFor(msg).length}
                <ol class="steps">
                  {#each stepsFor(msg) as tc, si (tc.id)}
                    <li>
                      <span class="step-n" aria-hidden="true">{String(si + 1).padStart(2, '0')}</span>
                      <div>
                        <strong>{toolLabel(tc)}</strong>
                        {#if toolArgs(tc)}
                          <span class="args">{toolArgs(tc)}</span>
                        {/if}
                        <span class="gate">awaits consent before action</span>
                      </div>
                    </li>
                  {/each}
                </ol>
              {/if}
            </article>
          {/each}
          {#if conversation.isStreaming}
            <article class="msg" data-role="assistant">
              <header>
                <span>Condura</span>
                <span class="cite live">thinking</span>
              </header>
              {#if conversation.streamingDelta}
                <div class="bubble streaming">{conversation.streamingDelta}</div>
              {:else}
                <div class="bubble streaming muted">Reading the room…</div>
              {/if}
              {#if conversation.streamingToolCalls.length}
                <ol class="steps live">
                  {#each conversation.streamingToolCalls as tc, si (tc.id)}
                    <li>
                      <span class="step-n" aria-hidden="true">{String(si + 1).padStart(2, '0')}</span>
                      <div>
                        <strong>{toolLabel(tc)}</strong>
                        {#if toolArgs(tc)}
                          <span class="args">{toolArgs(tc)}</span>
                        {/if}
                        <span class="gate">proposed · not yet allowed</span>
                      </div>
                    </li>
                  {/each}
                </ol>
              {/if}
            </article>
          {/if}
          {#if conversation.streamingError}
            <p class="err">{conversation.streamingError}</p>
          {/if}
        </div>
      </div>
    {/if}
  </div>

  <div class="composer" class:focused class:ready={canSend} data-tone={tone}>
    <div class="tone" aria-live="polite">
      <span class="tone-dot" aria-hidden="true"></span>
      {toneLabel}
    </div>
    <textarea
      bind:this={ta}
      bind:value={draft}
      class="input"
      rows="1"
      placeholder={halted
        ? 'Halted — resume to ask'
        : conversation.isStreaming
          ? 'Thinking…'
          : 'Ask Condura…'}
      disabled={halted || conversation.isStreaming}
      onfocus={() => (focused = true)}
      onblur={() => (focused = false)}
      onkeydown={onKey}
      oninput={onInput}
    ></textarea>
    <div class="bar">
      <div class="meta">
        {#if modelOptions.length > 0}
          <label class="model">
            <span class="sr">Model</span>
            <select
              bind:value={selectedModel}
              aria-label="Model"
              disabled={halted || conversation.isStreaming}
            >
              {#each modelOptions as opt (opt.value)}
                <option value={opt.value}>{opt.label}</option>
              {/each}
            </select>
          </label>
        {:else}
          <span
            class="offline"
            class:warn={!selectedModel}
            title={selectedModel || 'No model · connect daemon'}
          >
            <span class="full">{selectedModel || 'No model · connect daemon'}</span>
            <span class="short">{selectedModel ? selectedModel.split(':').pop() : 'No model'}</span>
          </span>
        {/if}
        <span class="hint">Enter to send · Esc to stop · Shift+Enter for line</span>
      </div>
      <div class="actions">
        {#if conversation.isStreaming}
          <button type="button" class="md-btn md-btn-ghost" onclick={() => void conversation.cancel()}>
            Stop
          </button>
        {/if}
        <button
          type="button"
          class="md-btn md-btn-primary"
          disabled={!canSend}
          onclick={() => void send()}
        >
          Send
        </button>
      </div>
    </div>
  </div>
</section>

<style>
  .chat {
    height: 100%;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }
  .feed {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: 28px 28px 16px;
    scroll-behavior: smooth;
  }

  /* —— Empty desk —— */
  .hero {
    max-width: 760px;
    margin: min(6vh, 48px) auto 0;
  }
  .brand-line {
    display: flex;
    align-items: baseline;
    gap: 8px;
    margin: 0 0 18px;
    animation: md-rise 420ms var(--md-ease) both;
  }
  .word {
    font-family: var(--md-font-display);
    font-size: 15px;
    font-weight: 700;
    letter-spacing: -0.03em;
    color: var(--md-ink);
  }
  .slash {
    color: var(--md-ink-faint);
    font-weight: 400;
  }
  .edition {
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  h1 {
    font-family: var(--md-font-display);
    font-size: clamp(40px, 7vw, 64px);
    font-weight: 700;
    letter-spacing: -0.055em;
    line-height: 0.98;
    margin: 0 0 16px;
    animation: md-rise 520ms var(--md-ease) 40ms both;
  }
  h1 span {
    color: var(--md-cobalt);
    background: linear-gradient(105deg, var(--md-cobalt), var(--md-live));
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
  }
  .hero.calm h1 span {
    color: var(--md-cobalt);
    background: none;
    -webkit-background-clip: unset;
    background-clip: unset;
  }
  .sub {
    margin: 0 0 14px;
    max-width: 42ch;
    font-size: 16px;
    line-height: 1.55;
    color: var(--md-ink-mute);
    animation: md-rise 520ms var(--md-ease) 80ms both;
  }
  .live {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    margin: 0 0 28px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.06em;
    color: var(--md-ink-faint);
    animation: md-rise 520ms var(--md-ease) 100ms both;
  }
  .live-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--md-ink-faint);
  }
  .live.hot {
    color: var(--md-live);
  }
  .live.hot .live-dot {
    background: var(--md-live);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-live) 18%, transparent);
  }
  .live.bad {
    color: var(--md-halt);
  }
  .live.bad .live-dot {
    background: var(--md-halt);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-halt) 18%, transparent);
  }

  .pipe {
    display: flex;
    flex-wrap: wrap;
    gap: 6px 4px;
    margin: 0 0 32px;
    padding: 0;
    list-style: none;
    animation: md-rise 520ms var(--md-ease) 120ms both;
  }
  .pipe li {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    border-radius: 999px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 70%, transparent);
  }
  .pipe .n {
    font-family: var(--md-font-mono);
    font-size: 10px;
    color: var(--md-cobalt);
    letter-spacing: 0.08em;
  }
  .pipe .t {
    font-size: 12px;
    font-weight: 700;
    color: var(--md-ink-soft);
  }

  .atlas-head {
    margin-bottom: 14px;
  }
  .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 8px;
  }
  .atlas h2 {
    font-family: var(--md-font-display);
    font-size: 22px;
    letter-spacing: -0.04em;
    margin: 0 0 6px;
  }
  .atlas-note {
    margin: 0;
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
    max-width: 46ch;
  }
  .atlas-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }
  .door {
    text-align: left;
    display: grid;
    gap: 6px;
    padding: 16px 16px 14px;
    border-radius: 18px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    cursor: pointer;
    color: inherit;
    transition:
      border-color 180ms var(--md-ease),
      transform 180ms var(--md-spring),
      box-shadow 180ms var(--md-ease);
  }
  .door:hover,
  .door.focus {
    border-color: color-mix(in oklab, var(--md-cobalt) 45%, transparent);
    transform: translateY(-2px);
    box-shadow: var(--md-shadow);
  }
  .door:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .door:disabled {
    opacity: 0.45;
    cursor: not-allowed;
    transform: none;
  }
  .door-k {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-cobalt);
  }
  .door-t {
    font-family: var(--md-font-display);
    font-size: 16px;
    font-weight: 700;
    letter-spacing: -0.03em;
    line-height: 1.2;
  }
  .door-b {
    font-size: 12.5px;
    line-height: 1.45;
    color: var(--md-ink-mute);
  }
  .door-a {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    margin-top: 6px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.06em;
    color: var(--md-cobalt);
    font-weight: 600;
  }

  .recent {
    margin-top: 28px;
    padding-top: 20px;
    border-top: 1px solid var(--md-line);
  }
  .recent-row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .recent-chip {
    padding: 8px 12px;
    border-radius: 999px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-stage);
    font-size: 12px;
    font-weight: 600;
    color: var(--md-ink-mute);
    cursor: pointer;
    max-width: 220px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .recent-chip:hover {
    border-color: var(--md-cobalt);
    color: var(--md-ink);
  }
  .recent-chip:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }

  /* —— Thread —— */
  .thread {
    max-width: 760px;
    margin: 0 auto;
  }
  .thread-bar {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 14px;
    padding-bottom: 14px;
    border-bottom: 1px solid var(--md-line);
  }
  .thread-title {
    font-family: var(--md-font-display);
    font-size: 22px;
    letter-spacing: -0.04em;
    margin: 0;
  }
  .thread-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  :global(.md-btn.tiny) {
    padding: 7px 12px;
    font-size: 12px;
  }
  .rail {
    display: flex;
    gap: 6px;
    overflow-x: auto;
    margin-bottom: 16px;
    padding-bottom: 2px;
  }
  .rail-item {
    flex: none;
    padding: 7px 12px;
    border-radius: 999px;
    border: 1px solid var(--md-line);
    background: transparent;
    font-size: 12px;
    font-weight: 600;
    color: var(--md-ink-faint);
    cursor: pointer;
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .rail-item.on {
    color: #fff;
    background: var(--md-cobalt);
    border-color: var(--md-cobalt);
  }
  .rail-item:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }

  .messages {
    display: flex;
    flex-direction: column;
    gap: 18px;
    padding-bottom: 8px;
  }
  .msg {
    display: flex;
    flex-direction: column;
    gap: 6px;
    max-width: min(100%, 580px);
    animation: md-rise 360ms var(--md-ease) both;
  }
  .msg[data-role='user'] {
    align-self: flex-end;
    align-items: flex-end;
  }
  .msg[data-role='assistant'] {
    align-self: flex-start;
    align-items: flex-start;
  }
  .msg header {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    padding: 0 4px;
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .msg header .cite {
    margin: 0;
    letter-spacing: 0.1em;
    color: var(--md-cobalt);
    opacity: 0.85;
  }
  .msg header .cite.live {
    color: var(--md-live);
  }
  .msg[data-role='user'] header {
    color: var(--md-cobalt);
  }
  .msg[data-role='assistant'] header {
    color: var(--md-live);
  }
  .bubble {
    font-size: 15.5px;
    line-height: 1.55;
    white-space: pre-wrap;
    word-break: break-word;
    padding: 12px 16px;
    border-radius: 18px;
  }
  .msg[data-role='user'] .bubble {
    background: color-mix(in oklab, var(--md-cobalt) 12%, var(--md-surface));
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 22%, var(--md-line-strong));
    border-bottom-right-radius: 6px;
  }
  .msg[data-role='assistant'] .bubble {
    background: var(--md-surface);
    border: 1px solid var(--md-line-strong);
    border-bottom-left-radius: 6px;
    box-shadow: var(--md-shadow);
  }
  .bubble.muted {
    color: var(--md-ink-mute);
    font-style: italic;
  }
  .bubble.streaming::after {
    content: '';
    display: inline-block;
    width: 8px;
    height: 1.1em;
    margin-left: 2px;
    vertical-align: text-bottom;
    background: var(--md-live);
    border-radius: 2px;
    animation: md-caret 1s steps(1) infinite;
  }
  .steps {
    margin: 0;
    padding: 0;
    list-style: none;
    display: grid;
    gap: 8px;
    width: 100%;
  }
  .steps li {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 12px;
    align-items: start;
    padding: 12px 14px;
    border-radius: 16px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-stage) 65%, var(--md-surface));
  }
  .steps.live li {
    border-color: color-mix(in oklab, var(--md-live) 28%, transparent);
  }
  .step-n {
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.08em;
    color: var(--md-cobalt);
    margin-top: 2px;
  }
  .steps.live .step-n {
    color: var(--md-live);
  }
  .steps strong {
    display: block;
    font-size: 13px;
    font-weight: 700;
    letter-spacing: -0.02em;
  }
  .steps .args {
    display: block;
    margin-top: 3px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    color: var(--md-ink-faint);
    word-break: break-word;
  }
  .gate {
    display: block;
    margin-top: 6px;
    font-family: var(--md-font-mono);
    font-size: 9px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .err {
    color: var(--md-halt);
    font-size: 13px;
    align-self: flex-start;
  }

  /* —— Composer —— */
  .composer {
    margin: 0 auto 88px;
    max-width: 760px;
    width: calc(100% - 56px);
    border-radius: 22px;
    border: 1px solid var(--md-line-strong);
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
    box-shadow: var(--md-shadow);
    padding: 10px 12px 12px;
    transition:
      border-color var(--md-dur) var(--md-ease),
      box-shadow var(--md-dur) var(--md-ease),
      transform 240ms var(--md-spring);
  }
  .composer .tone {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    margin: 2px 6px 8px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .tone-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--md-ink-faint);
  }
  .composer[data-tone='ready'] {
    border-color: color-mix(in oklab, var(--md-cobalt) 40%, transparent);
  }
  .composer[data-tone='ready'] .tone {
    color: var(--md-cobalt);
  }
  .composer[data-tone='ready'] .tone-dot {
    background: var(--md-cobalt);
  }
  .composer[data-tone='thinking'] {
    border-color: color-mix(in oklab, var(--md-live) 40%, transparent);
  }
  .composer[data-tone='thinking'] .tone {
    color: var(--md-live);
  }
  .composer[data-tone='thinking'] .tone-dot {
    background: var(--md-live);
    animation: md-tone-pulse 1.2s var(--md-ease) infinite;
  }
  .composer[data-tone='halted'] {
    border-color: color-mix(in oklab, var(--md-halt) 45%, transparent);
  }
  .composer[data-tone='halted'] .tone {
    color: var(--md-halt);
  }
  .composer[data-tone='halted'] .tone-dot {
    background: var(--md-halt);
  }
  .composer[data-tone='offline'] .tone {
    color: var(--md-ink-mute);
  }
  .composer.focused,
  .composer.ready.focused {
    transform: translateY(-2px);
    border-color: color-mix(in oklab, var(--md-cobalt) 50%, var(--md-line-strong));
    box-shadow: var(--md-focus), var(--md-shadow-lift);
    background: var(--md-surface);
  }
  .input {
    width: 100%;
    border: 0;
    background: transparent;
    color: var(--md-ink);
    font-size: 16px;
    line-height: 1.5;
    resize: none;
    min-height: 28px;
    max-height: 160px;
    outline: none;
    overflow-y: auto;
  }
  .input::placeholder {
    color: var(--md-ink-faint);
  }
  .bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-top: 10px;
    padding-top: 10px;
    border-top: 1px solid var(--md-line);
  }
  .meta {
    display: flex;
    flex-direction: column;
    gap: 3px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.04em;
    color: var(--md-ink-faint);
    overflow: hidden;
    min-width: 0;
  }
  .sr {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
  .model {
    display: inline-flex;
    max-width: 100%;
  }
  .model select {
    appearance: none;
    max-width: min(280px, 52vw);
    border: 1px solid var(--md-line);
    background: var(--md-stage)
      url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%236B7A90' stroke-width='2'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E")
      no-repeat right 10px center;
    padding: 6px 28px 6px 10px;
    border-radius: 999px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    color: var(--md-ink-soft);
    outline: none;
  }
  .model select:focus-visible {
    border-color: var(--md-cobalt);
    box-shadow: var(--md-focus);
  }
  .offline.warn {
    color: var(--md-halt);
  }
  .offline .short {
    display: none;
  }
  .hint {
    opacity: 0.85;
  }
  .actions {
    display: flex;
    gap: 8px;
    flex: none;
  }

  @keyframes md-rise {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  @keyframes md-caret {
    50% {
      opacity: 0;
    }
  }
  @keyframes md-tone-pulse {
    0%,
    100% {
      transform: scale(1);
      opacity: 1;
    }
    50% {
      transform: scale(1.35);
      opacity: 0.65;
    }
  }

  @media (max-width: 720px) {
    .feed {
      padding: 20px 14px 12px;
    }
    .atlas-grid {
      grid-template-columns: 1fr;
    }
    .composer {
      width: calc(100% - 28px);
      margin-bottom: 96px;
    }
    .thread-bar {
      flex-direction: column;
      align-items: stretch;
    }
    .pipe li {
      padding: 6px 10px;
    }
  }
  @media (max-width: 420px) {
    h1 {
      font-size: clamp(32px, 11vw, 40px);
    }
    .offline .full {
      display: none;
    }
    .offline .short {
      display: inline;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .hero *,
    .msg,
    .composer,
    .bubble.streaming::after,
    .tone-dot {
      animation: none !important;
    }
    .feed {
      scroll-behavior: auto;
    }
  }
</style>
