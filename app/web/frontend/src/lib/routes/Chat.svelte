<script lang="ts">
  import { onMount } from 'svelte'
  import { marked } from 'marked'
  import DOMPurify from 'dompurify'

  import { conversation } from '../stores/conversation.svelte'
  import { daemon } from '../stores/daemon.svelte'
  import { settings } from '../stores/settings.svelte'
  import { ipc } from '../ipc/client'
  import type { ProviderInfo } from '../ipc/types'
  import { t } from '../i18n'

  import Button from '../components/ui/Button.svelte'
  import IconButton from '../components/ui/IconButton.svelte'
  import Card from '../components/ui/Card.svelte'
  import Badge from '../components/ui/Badge.svelte'
  import Select from '../components/ui/Select.svelte'
  import VoiceOrb from '../components/VoiceOrb.svelte'
  import EmptyState from '../components/ui/EmptyState.svelte'

  let inputText = $state('')
  let voiceOn = $state(false)
  let voiceListening = $state(false)
  let slashIndex = $state(0)
  let scrollerEl = $state<HTMLDivElement | null>(null)
  let providers = $state<ProviderInfo[]>([])

  const slashCommands = [
    { id: 'help',    label: '/help',    hint: t('composer.slash_help') },
    { id: 'model',   label: '/model',   hint: t('composer.slash_model') },
    { id: 'about',   label: '/about',   hint: t('composer.slash_about') },
    { id: 'clear',   label: '/clear',   hint: t('composer.slash_clear') },
    { id: 'compact', label: '/compact', hint: t('composer.slash_compact') },
  ]

  const modelOptions = $derived(
    providers.flatMap((p) =>
      p.models.map((m) => ({ value: `${p.name}:${m.id}`, label: `${p.name} · ${m.id}` }))
    )
  )

  function defaultProviderModel(): string {
    const cfg = settings.config
    if (!cfg) return modelOptions[0]?.value ?? ''
    const entries = Object.entries(cfg.llm.providers)
    const enabled = entries.find(([_, p]) => p.enabled)
    if (enabled) return `${enabled[0]}:${enabled[1].default_model}`
    return modelOptions[0]?.value ?? ''
  }

  let selectedModel = $state('')

  $effect(() => {
    if (!selectedModel && modelOptions.length > 0) {
      selectedModel = defaultProviderModel() || modelOptions[0].value
    }
  })

  const showSlashMenu = $derived(inputText.startsWith('/'))

  function onInput(e: Event): void {
    inputText = (e.currentTarget as HTMLTextAreaElement).value
  }

  function onKey(e: KeyboardEvent): void {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      void send()
    } else if (showSlashMenu && e.key === 'ArrowDown') {
      e.preventDefault()
      slashIndex = (slashIndex + 1) % slashCommands.length
    } else if (showSlashMenu && e.key === 'ArrowUp') {
      e.preventDefault()
      slashIndex = (slashIndex - 1 + slashCommands.length) % slashCommands.length
    }
  }

  async function send(): Promise<void> {
    const text = inputText.trim()
    if (!text || conversation.isStreaming) return
    inputText = ''

    if (!conversation.currentID) {
      await conversation.createNew(text.slice(0, 60))
    }

    const [provider, model] = selectedModel.split(':')
    if (!provider || !model) return
    await conversation.send(provider, model, text)
  }

  async function newChat(): Promise<void> {
    await conversation.createNew()
  }

  async function openConversation(id: number): Promise<void> {
    await conversation.open(id)
  }

  function toggleVoice(): void {
    voiceOn = !voiceOn
    voiceListening = voiceOn
  }

  function applySlash(): void {
    if (showSlashMenu && slashCommands[slashIndex]) {
      inputText = slashCommands[slashIndex].label + ' '
    }
  }

  $effect(() => {
    if (scrollerEl && (conversation.messages.length || conversation.isStreaming)) {
      requestAnimationFrame(() => {
        scrollerEl?.scrollTo({ top: scrollerEl.scrollHeight, behavior: 'smooth' })
      })
    }
  })

  onMount(() => {
    void conversation.refreshList()
    conversation.startListening()
    void settings.refresh()
    void (async () => {
      try { providers = await ipc.providersList() } catch { providers = [] }
    })()
    return () => conversation.stopListening()
  })

  function renderSafeMarkdown(text: string): string {
    // marked.parse() emits raw HTML; tool output, prompt-injected
    // model replies, and crafted STT transcripts all flow through
    // this path with full IPC access behind it. DOMPurify strips
    // <script>, event handlers, javascript: URLs, and any other
    // XSS vectors before the result reaches {@html}.
    const html = marked.parse(text, { async: false }) as string
    return DOMPurify.sanitize(html)
  }
</script>

<div class="chat">
  <!-- ── Conversation rail ────────────────────────────── -->
  <aside class="chat-rail">
    <div class="chat-rail-header">
      <Button variant="primary" size="sm" fullWidth onclick={newChat}>
        + {t('chat.new_chat')}
      </Button>
    </div>
    <nav class="chat-rail-list" aria-label="Conversations">
      {#if conversation.conversations.length === 0}
        <div class="chat-rail-empty">{t('chat.no_conversations') ?? 'No conversations yet.'}</div>
      {:else}
        {#each conversation.conversations as c (c.id)}
          <button
            type="button"
            class="chat-rail-item"
            class:active={conversation.currentID === c.id}
            onclick={() => openConversation(c.id)}
          >
            <span class="chat-rail-title">{c.title || t('chat.new_chat')}</span>
            <span class="chat-rail-meta">{new Date(c.updated_at).toLocaleString()}</span>
          </button>
        {/each}
      {/if}
    </nav>
  </aside>

  <!-- ── Conversation pane ───────────────────────────── -->
  <section class="chat-pane">
    <header class="chat-header">
      <div class="chat-header-title">
        <h2 class="chat-title">{conversation.currentTitle || t('chat.new_chat')}</h2>
        <span class="chat-meta">
          {conversation.messages.length} {t('chat.messages') ?? 'messages'}
        </span>
      </div>
      <div class="chat-header-actions">
        <Select
          bind:value={selectedModel}
          options={modelOptions}
          placeholder={t('composer.model_picker_label')}
        />
        {#if conversation.currentID}
          <IconButton variant="ghost" ariaLabel="Delete conversation" onclick={() => conversation.deleteCurrent()}>
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
              <path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
            </svg>
          </IconButton>
        {/if}
      </div>
    </header>

    <div class="chat-scroll" bind:this={scrollerEl}>
      {#if conversation.messages.length === 0 && !conversation.isStreaming}
        <EmptyState
          title={t('chat.welcome_title') ?? 'Ask Condura anything.'}
          description={t('chat.welcome_lede') ?? 'A free, local AI conductor for everything on your computer.'}
        >
          {#snippet icon()}
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M12 2L3 7l9 5 9-5-9-5zM3 17l9 5 9-5M3 12l9 5 9-5" />
            </svg>
          {/snippet}
          {#snippet action()}
            <div class="welcome-actions">
              <Button variant="accent-ghost" size="sm" onclick={() => { inputText = 'Review my recent file changes' }}>
                Review my recent file changes
              </Button>
              <Button variant="accent-ghost" size="sm" onclick={() => { inputText = 'Draft a release note for v0.1.0' }}>
                Draft a release note for v0.1.0
              </Button>
              <Button variant="accent-ghost" size="sm" onclick={() => { inputText = 'Summarize the audit log' }}>
                Summarize the audit log
              </Button>
            </div>
          {/snippet}
        </EmptyState>
      {:else}
        <div class="chat-messages">
          {#each conversation.messages as msg, i (i)}
            <article class="msg msg-{msg.role}">
              <div class="msg-meta">
                <span class="msg-role">{msg.role}</span>
                {#if msg.tool_calls && msg.tool_calls.length > 0}
                  <Badge tone="accent" size="xs">{msg.tool_calls.length} tool calls</Badge>
                {/if}
              </div>
              <div class="msg-body">
                {#if msg.role === 'assistant'}
                  <div class="msg-markdown">{@html renderSafeMarkdown(msg.content)}</div>
                {:else}
                  <div class="msg-text">{msg.content}</div>
                {/if}
                {#if msg.tool_calls}
                  <div class="msg-tools">
                    {#each msg.tool_calls as tc (tc.id)}
                      <Card elevation={2} padding="sm">
                        <div class="msg-tool-row">
                          <span class="msg-tool-name">{tc.function.name}</span>
                          <code class="msg-tool-args">{tc.function.arguments}</code>
                        </div>
                      </Card>
                    {/each}
                  </div>
                {/if}
              </div>
            </article>
          {/each}

          {#if conversation.isStreaming}
            <article class="msg msg-assistant msg-streaming">
              <div class="msg-meta">
                <span class="msg-role">{t('chat.assistant')}</span>
                <Badge tone="accent" dot pulse size="xs">{t('chat.streaming')}</Badge>
              </div>
              <div class="msg-body">
                {#if conversation.streamingDelta}
                  <div class="msg-markdown">{@html renderSafeMarkdown(conversation.streamingDelta)}</div>
                {:else}
                  <div class="msg-thinking">
                    <span class="dot-loader"><span></span><span></span><span></span></span>
                    {t('chat.thinking') ?? 'Thinking…'}
                  </div>
                {/if}
                {#if conversation.streamingToolCalls.length > 0}
                  <div class="msg-tools">
                    {#each conversation.streamingToolCalls as tc (tc.id)}
                      <Card elevation={2} padding="sm">
                        <div class="msg-tool-row">
                          <span class="msg-tool-name">{tc.function.name}</span>
                          <code class="msg-tool-args">{tc.function.arguments}</code>
                        </div>
                      </Card>
                    {/each}
                  </div>
                {/if}
              </div>
            </article>
          {/if}

          {#if conversation.streamingError}
            <div class="chat-error">{conversation.streamingError}</div>
          {/if}
        </div>
      {/if}
    </div>

    <!-- ── Composer ───────────────────────────────────── -->
    <footer class="composer">
      {#if voiceOn}
        <div class="composer-voice">
          <VoiceOrb size="md" status={voiceListening ? 'listening' : 'off'} />
        </div>
      {/if}

      <div class="composer-shell">
        {#if showSlashMenu}
          <div class="slash-menu" role="listbox">
            {#each slashCommands as cmd, i (cmd.id)}
              <button
                type="button"
                class="slash-item"
                class:active={i === slashIndex}
                onclick={() => { slashIndex = i; applySlash() }}
                onmouseenter={() => slashIndex = i}
              >
                <span class="slash-label">{cmd.label}</span>
                <span class="slash-hint">{cmd.hint}</span>
              </button>
            {/each}
          </div>
        {/if}

        <div class="composer-input-row">
          <textarea
            class="composer-input"
            placeholder={voiceOn ? t('composer.voice_idle') : t('chat.placeholder')}
            value={inputText}
            oninput={onInput}
            onkeydown={onKey}
            rows="1"
          ></textarea>
          <div class="composer-actions">
            <IconButton
              variant="ghost"
              ariaLabel={voiceOn ? 'Turn voice off' : 'Turn voice on'}
              active={voiceOn}
              onclick={toggleVoice}
            >
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z" />
                <path d="M19 10v2a7 7 0 0 1-14 0v-2M12 19v4M8 23h8" />
              </svg>
            </IconButton>
            {#if conversation.isStreaming}
              <Button variant="danger" size="md" onclick={() => conversation.cancel()}>
                {t('chat.stop')}
              </Button>
            {:else}
              <Button variant="primary" size="md" onclick={send} disabled={!inputText.trim()}>
                {t('chat.send')}
              </Button>
            {/if}
          </div>
        </div>

        <div class="composer-hint">
          <span>{daemon.connected ? '● ' + t('app.status.connected') : '○ ' + t('app.status.disconnected')}</span>
          <span class="dot">·</span>
          <span>{t('chat.composer_hint') ?? 'Shift+Enter for newline. / for commands.'}</span>
        </div>
      </div>
    </footer>
  </section>
</div>

<style>
  .chat {
    display: grid;
    grid-template-columns: 280px 1fr;
    height: 100%;
    min-height: 0;
  }

  .chat-rail {
    display: flex;
    flex-direction: column;
    background: var(--surface-1);
    border-right: 1px solid var(--border);
    min-height: 0;
  }

  .chat-rail-header {
    padding: var(--space-3);
    border-bottom: 1px solid var(--border);
  }

  .chat-rail-list {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-2);
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .chat-rail-empty {
    padding: var(--space-5) var(--space-3);
    color: var(--text-faint);
    font-size: var(--size-sm);
    text-align: center;
  }

  .chat-rail-item {
    appearance: none;
    background: transparent;
    border: 1px solid transparent;
    border-radius: var(--radius-md);
    padding: var(--space-3);
    color: var(--text-muted);
    cursor: pointer;
    text-align: left;
    display: flex;
    flex-direction: column;
    gap: 2px;
    transition: background-color var(--transition-fast) ease, color var(--transition-fast) ease, border-color var(--transition-fast) ease;
  }
  .chat-rail-item:hover { background: var(--surface-2); color: var(--text); }
  .chat-rail-item.active {
    background: var(--surface-2);
    color: var(--text);
    border-color: var(--border-focus);
  }

  .chat-rail-title {
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .chat-rail-meta {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-faint);
  }

  .chat-pane {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
    background: var(--bg);
  }

  .chat-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    padding: var(--space-3) var(--space-5);
    border-bottom: 1px solid var(--border);
    background: var(--surface-1);
  }
  .chat-header-title {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }
  .chat-title {
    font-family: var(--font-display);
    font-size: var(--size-lg);
    font-weight: var(--weight-medium);
    color: var(--text);
    letter-spacing: var(--tracking-tight);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .chat-meta {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-faint);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
  }

  .chat-header-actions {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .chat-header-actions :global(.select-wrap) { min-width: 220px; }

  .chat-scroll {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-6) var(--space-7);
    min-height: 0;
  }

  .chat-messages {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
    max-width: 820px;
    margin: 0 auto;
  }

  .msg {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    animation: fade-in-up var(--transition-base) var(--ease-out-expo) both;
  }

  .msg-meta {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .msg-role {
    font-family: var(--font-mono);
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: var(--tracking-widest);
    color: var(--text-faint);
    font-weight: var(--weight-semibold);
  }
  .msg-user .msg-role  { color: var(--text-muted); }
  .msg-assistant .msg-role { color: var(--accent); }

  .msg-body { color: var(--text); line-height: var(--leading-normal); font-size: var(--size-md); }
  .msg-text { white-space: pre-wrap; }

  .msg-markdown { line-height: var(--leading-relaxed); }
  .msg-markdown :global(p) { margin: 0.5em 0; }
  .msg-markdown :global(p:first-child) { margin-top: 0; }
  .msg-markdown :global(p:last-child) { margin-bottom: 0; }
  .msg-markdown :global(h1),
  .msg-markdown :global(h2),
  .msg-markdown :global(h3) {
    font-family: var(--font-display);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-tight);
    margin: 0.8em 0 0.4em;
  }
  .msg-markdown :global(h1) { font-size: var(--size-xl); }
  .msg-markdown :global(h2) { font-size: var(--size-lg); }
  .msg-markdown :global(h3) { font-size: var(--size-md); }
  .msg-markdown :global(ul),
  .msg-markdown :global(ol) { margin: 0.5em 0; padding-left: 1.4em; }
  .msg-markdown :global(li) { margin: 0.25em 0; }
  .msg-markdown :global(code) {
    font-family: var(--font-mono);
    font-size: 0.88em;
    background: var(--surface-2);
    border: 1px solid var(--border);
    padding: 1px 6px;
    border-radius: var(--radius-xs);
  }
  .msg-markdown :global(pre) {
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: var(--space-3);
    overflow-x: auto;
    margin: 0.7em 0;
  }
  .msg-markdown :global(pre code) { background: none; border: none; padding: 0; }
  .msg-markdown :global(blockquote) {
    border-left: 2px solid var(--accent);
    padding-left: var(--space-3);
    margin: 0.7em 0;
    color: var(--text-muted);
    font-style: italic;
  }

  .msg-tools { display: flex; flex-direction: column; gap: var(--space-2); margin-top: var(--space-2); }
  .msg-tool-row { display: flex; flex-direction: column; gap: 4px; }
  .msg-tool-name {
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    color: var(--accent);
    font-weight: var(--weight-semibold);
  }
  .msg-tool-args {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-muted);
    white-space: pre-wrap;
    word-break: break-all;
  }

  .msg-thinking {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    color: var(--text-muted);
    font-size: var(--size-sm);
  }
  .msg-thinking .dot-loader span { background: var(--accent); }

  .msg-streaming .msg-body::after {
    content: '▍';
    display: inline-block;
    color: var(--accent);
    margin-left: 2px;
    animation: cursor-blink 1s var(--ease-in-out-quart) infinite;
  }

  .chat-error {
    background: var(--error-soft);
    border: 1px solid var(--border-danger);
    color: var(--error);
    border-radius: var(--radius-md);
    padding: var(--space-3);
    font-size: var(--size-sm);
  }

  .composer {
    padding: var(--space-3) var(--space-5) var(--space-4);
    background: var(--surface-1);
    border-top: 1px solid var(--border);
  }
  .composer-voice {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-3) 0;
  }

  .composer-shell {
    position: relative;
    max-width: 820px;
    margin: 0 auto;
  }

  .slash-menu {
    position: absolute;
    bottom: 100%;
    left: 0;
    right: 0;
    margin-bottom: var(--space-2);
    background: var(--surface-2);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    padding: var(--space-1);
    display: flex;
    flex-direction: column;
    gap: 1px;
    animation: fade-in-scale var(--transition-fast) var(--ease-out-expo) both;
  }
  .slash-item {
    appearance: none;
    background: transparent;
    border: none;
    color: var(--text);
    text-align: left;
    padding: var(--space-3);
    border-radius: var(--radius-sm);
    cursor: pointer;
    display: flex;
    flex-direction: column;
    gap: 2px;
    transition: background-color var(--transition-fast) ease;
  }
  .slash-item.active { background: var(--surface-3); }
  .slash-label { font-family: var(--font-mono); font-size: var(--size-sm); color: var(--text); }
  .slash-hint { font-size: var(--size-xs); color: var(--text-faint); }

  .composer-input-row {
    display: flex;
    align-items: end;
    gap: var(--space-2);
    background: var(--surface-2);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-lg);
    padding: var(--space-3);
    transition: border-color var(--transition-fast) ease, box-shadow var(--transition-fast) ease;
  }
  .composer-input-row:focus-within {
    border-color: var(--border-focus);
    box-shadow: 0 0 0 3px var(--accent-soft);
  }

  .composer-input {
    flex: 1;
    background: transparent;
    border: none;
    outline: none;
    color: var(--text);
    font-family: var(--font-sans);
    font-size: var(--size-md);
    line-height: var(--leading-normal);
    resize: none;
    min-height: 24px;
    max-height: 240px;
    padding: 6px 4px;
  }
  .composer-input::placeholder { color: var(--text-faint); }

  .composer-actions {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .composer-hint {
    margin-top: var(--space-2);
    text-align: center;
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--text-faint);
    letter-spacing: var(--tracking-wide);
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
  }
  .composer-hint .dot { opacity: 0.5; }

  .welcome-actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    justify-content: center;
    margin-top: var(--space-3);
  }
</style>