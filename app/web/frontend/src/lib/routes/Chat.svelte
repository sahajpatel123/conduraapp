<script lang="ts">
  import { conversation } from '../stores/conversation.svelte'
  import { daemon } from '../stores/daemon.svelte'
  import { settings } from '../stores/settings.svelte'
  import { t } from '../i18n'

  // The chat input. v5 runes: $state is implicit, no `let` needed.
  let inputText = $state('')
  let selectedProvider = $state('openai')
  let selectedModel = $state('')

  // Auto-pick default model when the provider changes.
  $effect(() => {
    if (daemon.connected && settings.config && !selectedModel) {
      const p = settings.config.llm.providers[selectedProvider]
      if (p?.default_model) {
        selectedModel = p.default_model
      }
    }
  })

  async function send(): Promise<void> {
    if (!inputText.trim() || conversation.isStreaming) {
      return
    }
    const text = inputText.trim()
    inputText = ''
    await conversation.send(selectedProvider, selectedModel, text)
  }

  function onKeydown(ev: KeyboardEvent): void {
    if (ev.key === 'Enter' && !ev.shiftKey) {
      ev.preventDefault()
      void send()
    }
  }

  async function cancel(): Promise<void> {
    await conversation.cancel()
  }

  // Auto-scroll to the bottom as new messages / tokens arrive.
  let scrollAnchor: HTMLDivElement | null = $state(null)
  $effect(() => {
    // touch the reactive values so this effect re-runs
    conversation.messages.length
    conversation.streamingDelta.length
    if (scrollAnchor) {
      scrollAnchor.scrollIntoView({ block: 'end', behavior: 'smooth' })
    }
  })
</script>

<div class="chat-page">
  <!-- Floating header -->
  <header class="chat-header">
    <h2 class="chat-title">{conversation.currentTitle}</h2>
    {#if conversation.isStreaming}
      <div class="chat-status">
        <span class="streaming-pill">{t('chat.streaming')}</span>
        <button class="btn-stop" onclick={cancel}>
          <svg viewBox="0 0 16 16" fill="currentColor" width="12" height="12"><rect x="3" y="3" width="10" height="10" rx="2" /></svg>
          {t('chat.stop')}
        </button>
      </div>
    {/if}
  </header>

  <!-- Message thread -->
  <div class="chat-thread">
    <div class="thread-inner">
      {#if conversation.messages.length === 0}
        <div class="empty-state">
          <div class="empty-icon"><svg viewBox="0 0 48 48" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 14h24a3 3 0 013 3v14a3 3 0 01-3 3H18l-9 7V17a3 3 0 013-3z"/><circle cx="18" cy="24" r="1.5" fill="currentColor"/><circle cx="24" cy="24" r="1.5" fill="currentColor"/><circle cx="30" cy="24" r="1.5" fill="currentColor"/></svg></div>
          <h3>{t('chat.empty.title')}</h3>
          <p>
            {#if !settings.config}
              {t('chat.empty.checking')}
            {:else if !daemon.connected}
              {t('chat.empty.waiting')}
            {:else}
              {#if Object.values(settings.config.llm?.providers ?? {}).every((p: any) => !p?.enabled)}
                {t('chat.empty.no_provider')} <a href="#/settings">{t('chat.empty.settings_link')}</a> {t('chat.empty.no_provider_after')}
              {:else}
                {t('chat.empty.get_started', selectedProvider)}
              {/if}
            {/if}
          </p>
        </div>
      {/if}
      {#each conversation.messages as msg, i (i)}
        <div class="message message-{msg.role}">
          <div class="message-content">{msg.content}</div>
          {#if msg.tool_calls && msg.tool_calls.length > 0}
            <div class="tool-calls" aria-label="Tool calls">
              {#each msg.tool_calls as tc (tc.id)}
                <details class="tool-call">
                  <summary>
                    <span class="tool-icon" aria-hidden="true"><svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="8" cy="8" r="2"/><path d="M8 1v2m0 10v2m-7-7h2m10 0h2M3.5 3.5l1.4 1.4m6.2 6.2 1.4 1.4m0-11-1.4 1.4M5.1 10.9l-1.4 1.4"/></svg></span>
                    <span class="tool-name">{tc.function.name}</span>
                  </summary>
                  <pre class="tool-args">{tc.function.arguments}</pre>
                </details>
              {/each}
            </div>
          {/if}
        </div>
      {/each}

      {#if conversation.isStreaming && conversation.streamingDelta}
        <div class="message message-assistant streaming">
          <div class="message-content">{conversation.streamingDelta}<span class="cursor"></span></div>
          {#if conversation.streamingToolCalls.length > 0}
            <div class="tool-calls" aria-label="Tool calls in flight">
              {#each conversation.streamingToolCalls as tc (tc.id)}
                <div class="tool-call streaming">
                  <span class="tool-icon" aria-hidden="true"><svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="8" cy="8" r="2"/><path d="M8 1v2m0 10v2m-7-7h2m10 0h2M3.5 3.5l1.4 1.4m6.2 6.2 1.4 1.4m0-11-1.4 1.4M5.1 10.9l-1.4 1.4"/></svg></span>
                  <span class="tool-name">{tc.function.name}</span>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}

      {#if conversation.streamingError}
        <div class="message message-error">
          <div class="message-content">{conversation.streamingError}</div>
        </div>
      {/if}

      <div bind:this={scrollAnchor}></div>
    </div>
  </div>

  <!-- Floating input pill -->
  <footer class="chat-input-wrap" class:is-streaming={conversation.isStreaming}>
    <div class="chat-input-pill">
      <div class="input-meta-row">
        <select bind:value={selectedProvider} class="meta-select">
          <option value="openai">openai</option>
          <option value="anthropic">anthropic</option>
          <option value="google">google</option>
          <option value="xai">xai</option>
          <option value="mistral">mistral</option>
          <option value="deepseek">deepseek</option>
          <option value="openrouter">openrouter</option>
          <option value="groq">groq</option>
          <option value="together">together</option>
          <option value="fireworks">fireworks</option>
          <option value="ollama">ollama</option>
        </select>
        <span class="meta-sep">·</span>
        <input
          type="text"
          bind:value={selectedModel}
          placeholder="model"
          class="meta-model"
        />
      </div>
      <div class="input-main-row">
        <input
          type="text"
          bind:value={inputText}
          onkeydown={onKeydown}
          placeholder={t('chat.placeholder')}
          class="input-message"
          disabled={conversation.isStreaming}
        />
        <button
          class="btn-send"
          onclick={send}
          disabled={!inputText.trim() || conversation.isStreaming}
          aria-label={t('chat.send')}
          title={t('chat.send')}
        >
          <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18" aria-hidden="true"><path d="M4 10h12m-5-5 5 5-5 5"/></svg>
        </button>
      </div>
      <p class="hint">
        {#if !daemon.connected}
          <span class="warn">{t('chat.not_connected')}</span>
        {:else}
          {t('chat.keyhint')}
        {/if}
      </p>
    </div>
  </footer>
</div>

<style>
  /* ── Layout ─────────────────────────────────────────── */
  .chat-page {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--color-bg);
    position: relative;
  }

  /* ── Header ─────────────────────────────────────────── */
  .chat-header {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-4) var(--space-5);
    flex-shrink: 0;
  }

  .chat-title {
    font-size: var(--size-xs);
    font-weight: var(--weight-medium);
    color: var(--color-text-faint);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
    text-align: center;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 60%;
  }

  .chat-status {
    position: absolute;
    right: var(--space-5);
    top: 50%;
    transform: translateY(-50%);
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .streaming-pill {
    background: var(--color-accent-gradient);
    color: #fff;
    padding: 3px 10px;
    border-radius: var(--radius-pill);
    font-size: var(--size-2xs);
    font-weight: var(--weight-semibold);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wide);
    box-shadow: 0 0 12px var(--color-glow);
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  .streaming-pill::before {
    content: '';
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: #fff;
    animation: pulse 1.2s ease-in-out infinite;
  }

  .btn-stop {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    background: transparent;
    color: var(--color-text-muted);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-pill);
    padding: 4px 12px;
    font-size: var(--size-2xs);
    font-weight: var(--weight-medium);
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .btn-stop:hover {
    color: var(--color-error);
    border-color: var(--color-error);
    background: var(--color-error-soft);
  }

  /* ── Thread ─────────────────────────────────────────── */
  .chat-thread {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-4) var(--space-5);
  }

  .thread-inner {
    max-width: var(--content-max-width);
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    padding-bottom: var(--space-4);
  }

  /* ── Messages ───────────────────────────────────────── */
  @keyframes messageIn {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .message {
    max-width: 85%;
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-lg);
    animation: messageIn var(--transition-slow) var(--ease-out-expo);
  }

  .message-user {
    align-self: flex-end;
    background: var(--color-accent-gradient-subtle);
    border: 1px solid var(--color-border-accent);
    border-left: 3px solid var(--color-accent);
  }

  .message-assistant {
    align-self: flex-start;
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    box-shadow: var(--shadow-xs), var(--shadow-inset);
  }

  .message-error {
    align-self: stretch;
    max-width: 100%;
    background: var(--color-error-soft);
    border-left: 3px solid var(--color-error);
    border-radius: var(--radius-md);
  }

  .message-content {
    font-size: var(--size-md);
    line-height: var(--leading-relaxed);
    white-space: pre-wrap;
    word-break: break-word;
    color: var(--color-text);
  }

  /* ── Streaming cursor — blinking bar ────────────────── */
  .streaming .cursor {
    display: inline-block;
    width: 2px;
    height: 1.1em;
    background: var(--color-accent);
    border-radius: 1px;
    margin-left: 3px;
    vertical-align: text-bottom;
    animation: cursorBlink 1s steps(2) infinite;
  }

  @keyframes cursorBlink {
    0%, 50% { opacity: 1; }
    50.01%, 100% { opacity: 0; }
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.3; }
  }

  /* ── Tool calls — collapsible glass cards ───────────── */
  .tool-calls {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin-top: var(--space-3);
    padding-top: var(--space-3);
    border-top: 1px solid var(--glass-border);
  }

  .tool-call {
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-md);
    font-size: var(--size-sm);
    overflow: hidden;
    transition: border-color var(--transition-fast);
  }

  .tool-call:hover {
    border-color: var(--glass-border-hover);
  }

  .tool-call.streaming {
    padding: var(--space-2) var(--space-3);
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .tool-call summary {
    cursor: pointer;
    padding: var(--space-2) var(--space-3);
    display: flex;
    align-items: center;
    gap: var(--space-2);
    list-style: none;
    color: var(--color-text-muted);
    user-select: none;
    -webkit-user-select: none;
    transition: color var(--transition-fast);
  }

  .tool-call summary:hover {
    color: var(--color-text);
  }

  .tool-call summary::-webkit-details-marker {
    display: none;
  }

  .tool-call summary::before {
    content: '';
    width: 0;
    height: 0;
    border-left: 4px solid var(--color-text-faint);
    border-top: 4px solid transparent;
    border-bottom: 4px solid transparent;
    transition: transform var(--transition-fast);
    flex-shrink: 0;
  }

  .tool-call[open] summary::before {
    transform: rotate(90deg);
  }

  .tool-icon {
    display: flex;
    align-items: center;
    color: var(--color-accent);
    flex-shrink: 0;
  }

  .tool-icon svg {
    width: 14px;
    height: 14px;
  }

  .tool-name {
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    color: var(--color-text);
    font-weight: var(--weight-medium);
  }

  .tool-args {
    padding: var(--space-2) var(--space-3);
    margin: 0;
    background: rgba(0, 0, 0, 0.25);
    border-top: 1px solid var(--glass-border);
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    color: var(--color-text-muted);
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 200px;
    overflow-y: auto;
    line-height: var(--leading-relaxed);
  }

  /* ── Input Pill ─────────────────────────────────────── */
  .chat-input-wrap {
    flex-shrink: 0;
    padding: 0 var(--space-5) var(--space-4);
  }

  .chat-input-pill {
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-2xl);
    padding: var(--space-3) var(--space-4);
    transition: border-color var(--transition-base), box-shadow var(--transition-base);
    box-shadow: var(--shadow-sm);
  }

  .chat-input-pill:focus-within {
    border-color: var(--color-accent);
    box-shadow: var(--shadow-glow), var(--shadow-inset);
  }

  /* Streaming glow pulse on the pill */
  @keyframes glowPulse {
    0%, 100% { box-shadow: var(--shadow-sm), 0 0 12px var(--color-glow); }
    50% { box-shadow: var(--shadow-sm), 0 0 24px var(--color-glow-strong); }
  }

  .is-streaming .chat-input-pill {
    border-color: var(--color-accent);
    animation: glowPulse 2s ease-in-out infinite;
  }

  /* Meta row: provider + model */
  .input-meta-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding-bottom: var(--space-2);
    border-bottom: 1px solid var(--color-border);
    margin-bottom: var(--space-2);
  }

  .meta-select {
    appearance: none;
    -webkit-appearance: none;
    background: transparent;
    border: none;
    color: var(--color-text-faint);
    font-size: var(--size-xs);
    font-family: var(--font-sans);
    cursor: pointer;
    padding: 2px 0;
    outline: none;
    transition: color var(--transition-fast);
  }

  .meta-select:hover,
  .meta-select:focus {
    color: var(--color-text-muted);
  }

  .meta-sep {
    color: var(--color-text-faint);
    font-size: var(--size-xs);
    opacity: 0.4;
    user-select: none;
  }

  .meta-model {
    background: transparent;
    border: none;
    color: var(--color-text-faint);
    font-size: var(--size-xs);
    font-family: var(--font-mono);
    outline: none;
    width: 140px;
    padding: 2px 0;
    transition: color var(--transition-fast);
  }

  .meta-model::placeholder {
    color: var(--color-text-faint);
    opacity: 0.5;
  }

  .meta-model:focus {
    color: var(--color-text-muted);
  }

  /* Main input row */
  .input-main-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .input-message {
    flex: 1;
    background: transparent;
    border: none;
    color: var(--color-text);
    font-size: var(--size-lg);
    font-family: var(--font-sans);
    padding: var(--space-2) 0;
    outline: none;
  }

  .input-message::placeholder {
    color: var(--color-text-faint);
  }

  .input-message:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Send button */
  .btn-send {
    width: 36px;
    height: 36px;
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    border: none;
    background: var(--color-accent-gradient);
    color: #fff;
    cursor: pointer;
    transition: opacity var(--transition-fast), transform var(--transition-spring),
      box-shadow var(--transition-base);
  }

  .btn-send:hover:not(:disabled) {
    transform: scale(1.08);
    box-shadow: var(--shadow-glow);
  }

  .btn-send:active:not(:disabled) {
    transform: scale(0.92);
  }

  .btn-send:disabled {
    opacity: 0.25;
    cursor: not-allowed;
  }

  /* Hint */
  .hint {
    font-size: var(--size-2xs);
    color: var(--color-text-faint);
    margin-top: var(--space-2);
    padding-left: var(--space-1);
    opacity: 0.7;
  }

  .hint .warn {
    color: var(--color-warn);
    opacity: 1;
  }

  /* ── Empty state ────────────────────────────────────── */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-10) var(--space-4);
    text-align: center;
  }

  .empty-icon {
    margin-bottom: var(--space-5);
    color: var(--color-accent);
    opacity: 0.5;
  }

  .empty-icon svg {
    width: 48px;
    height: 48px;
  }

  .empty-state h3 {
    font-size: var(--size-xl);
    font-weight: var(--weight-medium);
    color: var(--color-text);
    margin-bottom: var(--space-2);
    letter-spacing: var(--tracking-tight);
  }

  .empty-state p {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
    max-width: 360px;
    line-height: var(--leading-relaxed);
  }

  .empty-state a {
    color: var(--color-accent);
    text-decoration: none;
    border-bottom: 1px solid var(--color-accent-soft);
    transition: border-color var(--transition-fast);
  }

  .empty-state a:hover {
    border-color: var(--color-accent);
  }
</style>
