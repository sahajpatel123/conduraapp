<script lang="ts">
  import { onMount } from 'svelte'
  import { conversation } from '../../stores/conversation.svelte'
  import { settings } from '../../stores/settings.svelte'
  import { halt } from '../../stores/halt.svelte'
  import { ipc } from '../../ipc/client'
  import type { ProviderInfo } from '../../ipc/types'

  const STARTERS = [
    { id: 'sum', label: 'Summarize what’s on my screen' },
    { id: 'fix', label: 'Find and fix the last error I hit' },
    { id: 'plan', label: 'Plan the next hour of work' },
    { id: 'safe', label: 'What can you do without asking me?' },
  ]

  let draft = $state('')
  let focused = $state(false)
  let ta = $state<HTMLTextAreaElement | null>(null)
  let scrollEl = $state<HTMLDivElement | null>(null)
  let selectedModel = $state('')
  let providers = $state<ProviderInfo[]>([])

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
  const canSend = $derived(
    !!draft.trim() && !conversation.isStreaming && !halted && selectedModel.includes(':')
  )

  onMount(() => {
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
</script>

<section class="chat">
  <div class="feed" bind:this={scrollEl}>
    {#if !hasMessages}
      <div class="hero">
        <div class="orb" aria-hidden="true"></div>
        <p class="kicker">Local conductor · on this machine</p>
        <h1>Ask once.<br /><span>Watch the plan.</span></h1>
        <p class="sub">
          Condura shows what it will do before it does it. Start a line below, or choose a starter.
        </p>
        <div class="starters md-stagger">
          {#each STARTERS as s (s.id)}
            <button type="button" class="starter" onclick={() => pickStarter(s.label)}>
              <span class="starter-dot" aria-hidden="true"></span>
              <span>{s.label}</span>
            </button>
          {/each}
        </div>
      </div>
    {:else}
      <div class="messages">
        {#each conversation.messages as msg, i (i)}
          <article class="msg" data-role={msg.role}>
            <header>
              <span>{msg.role === 'user' ? 'You' : 'Condura'}</span>
              {#if msg.role === 'assistant'}
                <span class="cite">plan · gated</span>
              {/if}
            </header>
            <div class="bubble">{msg.content}</div>
          </article>
        {/each}
        {#if conversation.isStreaming}
          <article class="msg" data-role="assistant">
            <header>
              <span>Condura</span>
              <span class="cite live">thinking</span>
            </header>
            <div class="bubble streaming">{conversation.streamingDelta || '…'}</div>
          </article>
        {/if}
        {#if conversation.streamingError}
          <p class="err">{conversation.streamingError}</p>
        {/if}
      </div>
    {/if}
  </div>

  <div class="composer" class:focused class:ready={canSend}>
    <textarea
      bind:this={ta}
      bind:value={draft}
      class="input"
      rows="1"
      placeholder={halted ? 'Halted — resume to ask' : 'Ask Condura…'}
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
          <span class="offline" class:warn={!selectedModel} title={selectedModel || 'No model · connect daemon'}>
            <span class="full">{selectedModel || 'No model · connect daemon'}</span>
            <span class="short">{selectedModel ? selectedModel.split(':').pop() : 'No model'}</span>
          </span>
        {/if}
        <span class="hint">Enter to send · Shift+Enter for line</span>
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
  .hero {
    max-width: 640px;
    margin: min(10vh, 72px) auto 0;
    position: relative;
  }
  .orb {
    width: 64px;
    height: 64px;
    margin-bottom: 20px;
    border-radius: 50%;
    background: conic-gradient(from 200deg, var(--md-cobalt), var(--md-live), transparent 70%, var(--md-cobalt));
    mask: radial-gradient(circle at 50% 50%, transparent 42%, #000 44%);
    -webkit-mask: radial-gradient(circle at 50% 50%, transparent 42%, #000 44%);
    animation: md-rise 520ms var(--md-spring) both, md-orb-spin 10s linear 520ms infinite;
    filter: drop-shadow(0 8px 24px color-mix(in oklab, var(--md-cobalt) 28%, transparent));
  }
  .kicker {
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 16px;
    animation: md-rise 500ms var(--md-ease) both;
  }
  h1 {
    font-family: var(--md-font-display);
    font-size: clamp(40px, 7vw, 64px);
    font-weight: 700;
    letter-spacing: -0.055em;
    line-height: 0.98;
    margin: 0 0 16px;
    animation: md-rise 560ms var(--md-ease) 60ms both;
  }
  h1 span {
    background: linear-gradient(105deg, var(--md-cobalt), var(--md-live) 55%, var(--md-cobalt));
    background-size: 200% 100%;
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
    animation: md-shimmer 5s linear infinite;
  }
  .sub {
    font-size: 16px;
    line-height: 1.55;
    color: var(--md-ink-mute);
    max-width: 42ch;
    margin: 0 0 28px;
    animation: md-rise 560ms var(--md-ease) 120ms both;
  }
  .starters {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }
  @media (max-width: 640px) {
    .starters {
      grid-template-columns: 1fr;
    }
    .feed {
      padding: 20px 16px 12px;
    }
  }
  .starter {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    text-align: left;
    padding: 15px 16px;
    border-radius: 18px;
    border: 1px solid var(--md-line-strong);
    background: color-mix(in oklab, var(--md-surface) 82%, transparent);
    color: var(--md-ink-soft);
    font-size: 13px;
    font-weight: 600;
    line-height: 1.4;
    cursor: pointer;
    backdrop-filter: blur(8px);
    transition:
      transform 220ms var(--md-spring),
      border-color var(--md-dur) var(--md-ease),
      box-shadow var(--md-dur) var(--md-ease),
      background var(--md-dur) var(--md-ease);
  }
  .starter-dot {
    width: 8px;
    height: 8px;
    margin-top: 4px;
    border-radius: 50%;
    flex: none;
    background: var(--md-cobalt);
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-cobalt) 18%, transparent);
    transition: transform 220ms var(--md-spring);
  }
  .starter:hover {
    transform: translateY(-3px);
    border-color: color-mix(in oklab, var(--md-cobalt) 45%, var(--md-line-strong));
    box-shadow: var(--md-shadow);
    background: var(--md-surface);
  }
  .starter:hover .starter-dot {
    transform: scale(1.25);
  }
  .starter:active {
    transform: scale(0.98);
  }
  .starter:focus-visible {
    outline: none;
    border-color: var(--md-cobalt);
    box-shadow: var(--md-focus);
  }
  .messages {
    max-width: 720px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 18px;
    padding-bottom: 8px;
  }
  .msg {
    display: flex;
    flex-direction: column;
    gap: 6px;
    max-width: min(100%, 560px);
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
    color: var(--md-ink);
  }
  .msg[data-role='assistant'] .bubble {
    background: var(--md-surface);
    border: 1px solid var(--md-line-strong);
    border-bottom-left-radius: 6px;
    box-shadow: var(--md-shadow);
    color: var(--md-ink);
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
  .err {
    color: var(--md-halt);
    font-size: 13px;
    align-self: flex-start;
  }
  .composer {
    margin: 0 28px 88px;
    max-width: 760px;
    width: calc(100% - 56px);
    align-self: center;
    background: var(--md-surface);
    border: 1px solid var(--md-line-strong);
    border-radius: 24px;
    padding: 14px 16px 12px;
    box-shadow: var(--md-shadow-lift);
    transition:
      box-shadow 280ms var(--md-ease),
      transform 240ms var(--md-spring),
      border-color 240ms var(--md-ease);
    animation: md-dock-up 620ms var(--md-ease) 160ms both;
  }
  .composer.focused,
  .composer.ready.focused {
    transform: translateY(-2px);
    border-color: color-mix(in oklab, var(--md-cobalt) 50%, var(--md-line-strong));
    box-shadow: var(--md-focus), var(--md-shadow-lift);
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
      no-repeat right 8px center;
    color: var(--md-ink-soft);
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.02em;
    padding: 5px 26px 5px 10px;
    border-radius: 999px;
    outline: none;
    cursor: pointer;
    transition:
      border-color var(--md-dur) var(--md-ease),
      box-shadow var(--md-dur) var(--md-ease);
  }
  .model select:hover {
    border-color: color-mix(in oklab, var(--md-cobalt) 35%, var(--md-line));
  }
  .model select:focus-visible {
    border-color: var(--md-cobalt);
    box-shadow: var(--md-focus);
  }
  .model select:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .offline {
    display: inline-flex;
    align-items: center;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    padding: 4px 10px;
    border-radius: 999px;
    border: 1px dashed var(--md-line-strong);
    background: color-mix(in oklab, var(--md-stage) 80%, transparent);
  }
  .offline .short {
    display: none;
  }
  .offline.warn {
    border-color: color-mix(in oklab, var(--md-halt) 28%, var(--md-line-strong));
    color: var(--md-ink-mute);
  }
  .hint {
    font-family: var(--md-font-sans);
    font-size: 11px;
    letter-spacing: 0;
    color: var(--md-ink-faint);
    opacity: 0.85;
  }
  .actions {
    display: flex;
    gap: 8px;
    flex: none;
  }
  @keyframes md-orb-spin {
    to {
      transform: rotate(360deg);
    }
  }
  @keyframes md-shimmer {
    0% {
      background-position: 100% 0;
    }
    100% {
      background-position: -100% 0;
    }
  }
  @keyframes md-caret {
    50% {
      opacity: 0;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .orb,
    h1 span,
    .msg,
    .composer,
    .bubble.streaming::after {
      animation: none !important;
    }
    h1 span {
      color: var(--md-cobalt);
      background: none;
      -webkit-background-clip: unset;
      background-clip: unset;
    }
    .feed {
      scroll-behavior: auto;
    }
  }
  @media (max-width: 720px) {
    .composer {
      margin: 0 12px 78px;
      width: calc(100% - 24px);
      border-radius: 20px;
      padding: 12px 14px 10px;
    }
    .composer.focused,
    .composer.ready.focused {
      transform: none;
    }
    .bar {
      margin-top: 8px;
      padding-top: 8px;
      gap: 8px;
    }
    .hint {
      display: none;
    }
    .offline {
      max-width: min(220px, 48vw);
      font-size: 9px;
      padding: 3px 8px;
    }
    .model select {
      max-width: min(180px, 42vw);
      font-size: 9px;
      padding: 4px 22px 4px 8px;
    }
    .hero {
      margin: min(4vh, 28px) auto 0;
    }
    h1 {
      font-size: clamp(32px, 9vw, 44px);
    }
    .sub {
      font-size: 14px;
      margin-bottom: 20px;
    }
    .orb {
      width: 52px;
      height: 52px;
      margin-bottom: 14px;
    }
  }
  @media (max-width: 420px) {
    .composer {
      margin: 0 8px 72px;
      width: calc(100% - 16px);
      border-radius: 18px;
      padding: 10px 12px 9px;
    }
    .feed {
      padding: 16px 12px 8px;
    }
    .actions :global(.md-btn) {
      padding: 8px 14px;
      font-size: 12px;
    }
    .offline .full {
      display: none;
    }
    .offline .short {
      display: inline;
    }
  }
</style>
