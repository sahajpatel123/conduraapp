<script lang="ts">
  import { conversation } from '../stores/conversation.svelte'
  import { daemon } from '../stores/daemon.svelte'
  import { settings } from '../stores/settings.svelte'
  import { t } from '../i18n'

  let inputText = $state('')
  let selectedProvider = $state('openai')
  let selectedModel = $state('')

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

  let scrollAnchor: HTMLDivElement | null = $state(null)
  $effect(() => {
    conversation.messages.length
    conversation.streamingDelta.length
    if (scrollAnchor) {
      scrollAnchor.scrollIntoView({ block: 'end', behavior: 'smooth' })
    }
  })
</script>

<div class="chat-page">
  <!-- Top glow effect for depth -->
  <div class="ambient-glow-top"></div>

  <!-- Floating header -->
  <header class="chat-header">
    <h2 class="chat-title">{conversation.currentTitle || t('chat.empty.title')}</h2>
    {#if conversation.isStreaming}
      <div class="chat-status">
        <span class="streaming-pill">{t('chat.streaming')}</span>
        <button class="btn-stop" onclick={cancel}>
          <svg viewBox="0 0 16 16" fill="currentColor"><rect x="3" y="3" width="10" height="10" rx="2" /></svg>
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
          <div class="empty-logo-glow"></div>
          <div class="empty-icon">
            <svg viewBox="0 0 48 48" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 14h24a3 3 0 013 3v14a3 3 0 01-3 3H18l-9 7V17a3 3 0 013-3z"/><circle cx="18" cy="24" r="1.5" fill="currentColor"/><circle cx="24" cy="24" r="1.5" fill="currentColor"/><circle cx="30" cy="24" r="1.5" fill="currentColor"/></svg>
          </div>
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
        <div class="message-row" class:is-user={msg.role === 'user'}>
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
        </div>
      {/each}

      {#if conversation.isStreaming && conversation.streamingDelta}
        <div class="message-row">
          <div class="message message-assistant streaming">
            <div class="message-content">{conversation.streamingDelta}<span class="cursor"></span></div>
            {#if conversation.streamingToolCalls.length > 0}
              <div class="tool-calls" aria-label="Tool calls in flight">
                {#each conversation.streamingToolCalls as tc (tc.id)}
                  <div class="tool-call streaming">
                    <span class="tool-icon" aria-hidden="true"><svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="8" cy="8" r="2"/><path d="M8 1v2m0 10v2m-7-7h2m10 0h2M3.5 3.5l1.4 1.4m6.2 6.2 1.4 1.4m0-11-1.4 1.4M5.1 10.9l-1.4 1.4"/></svg></span>
                    <span class="tool-name">{tc.function.name}</span>
                    <div class="tool-pulse"></div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        </div>
      {/if}

      {#if conversation.streamingError}
        <div class="message-row">
          <div class="message message-error">
            <div class="message-content">{conversation.streamingError}</div>
          </div>
        </div>
      {/if}

      <div bind:this={scrollAnchor} class="scroll-anchor"></div>
    </div>
  </div>

  <!-- Floating input pill -->
  <footer class="chat-input-wrap" class:is-streaming={conversation.isStreaming}>
    <div class="chat-input-glow"></div>
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
        <span class="meta-sep">/</span>
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
          <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><path d="M5 10h10M11 6l4 4-4 4"/></svg>
        </button>
      </div>
    </div>
    <div class="input-footer">
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
    background: transparent;
    position: relative;
    overflow: hidden;
  }

  /* Ambient glow at the top — breathes softly */
  .ambient-glow-top {
    position: absolute;
    top: -120px;
    left: 50%;
    transform: translateX(-50%);
    width: 900px;
    height: 250px;
    background: radial-gradient(ellipse at top, var(--color-accent-soft), transparent 60%);
    pointer-events: none;
    z-index: 0;
    opacity: 0.4;
    animation: breathe-soft 8s ease-in-out infinite;
  }

  /* ── Header ─────────────────────────────────────────── */
  .chat-header {
    position: relative;
    z-index: 10;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-4) var(--space-5);
    flex-shrink: 0;
    background: linear-gradient(180deg, var(--color-bg) 0%, transparent 100%);
  }

  .chat-title {
    font-size: var(--size-xs);
    font-weight: var(--weight-bold);
    color: var(--color-text-dim);
    text-transform: uppercase;
    letter-spacing: var(--tracking-widest);
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
    gap: var(--space-3);
    animation: fade-in-scale var(--transition-base) var(--ease-out-expo) both;
  }

  /* Streaming pill — living, pulsing */
  .streaming-pill {
    background: var(--color-accent-soft);
    color: var(--color-accent);
    border: 1px solid var(--color-border-accent);
    padding: 5px 14px;
    border-radius: var(--radius-pill);
    font-size: 10px;
    font-weight: var(--weight-bold);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wide);
    display: inline-flex;
    align-items: center;
    gap: 8px;
    animation: streaming-pulse 2s ease-in-out infinite;
  }

  .streaming-pill::before {
    content: '';
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--color-accent);
    box-shadow: 0 0 10px var(--color-accent);
    animation: breathe 1s ease-in-out infinite;
  }

  .btn-stop {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    background: var(--glass-bg);
    color: var(--color-text-muted);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-pill);
    padding: 6px 14px;
    font-size: 11px;
    font-weight: var(--weight-semibold);
    cursor: pointer;
    transition: all var(--transition-fast);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wide);
  }

  .btn-stop svg {
    width: 10px;
    height: 10px;
  }

  .btn-stop:hover {
    color: var(--color-error);
    border-color: rgba(239, 68, 68, 0.4);
    background: var(--color-error-soft);
    box-shadow: 0 0 16px rgba(239, 68, 68, 0.2);
  }

  .btn-stop:active {
    transform: scale(0.95);
    transition-duration: var(--transition-instant);
  }

  /* ── Thread ─────────────────────────────────────────── */
  .chat-thread {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-4) var(--space-5) 180px;
    z-index: 1;
    position: relative;
    scroll-behavior: smooth;
  }

  .thread-inner {
    max-width: var(--content-max-width);
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  /* ── Messages — staggered entrance with blur ──────── */
  .message-row {
    display: flex;
    width: 100%;
    animation: message-in var(--transition-spring) var(--ease-out-expo) both;
  }

  .message-row.is-user {
    justify-content: flex-end;
  }

  .message {
    max-width: 85%;
    padding: var(--space-4) var(--space-5);
    border-radius: var(--radius-xl);
    position: relative;
  }

  /* User messages — glass with accent border glow */
  .message-user {
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    border: 1px solid var(--glass-border);
    border-bottom-right-radius: var(--radius-sm);
    color: var(--color-text);
    box-shadow: var(--shadow-md), var(--shadow-inset);
  }

  .message-user::before {
    content: '';
    position: absolute;
    inset: -1px;
    border-radius: inherit;
    background: var(--color-accent-gradient-subtle);
    z-index: -1;
    opacity: 0.5;
    mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
    -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
    -webkit-mask-composite: xor;
    mask-composite: exclude;
  }

  /* Assistant messages — clean, no bubble */
  .message-assistant {
    background: transparent;
    padding-left: 0;
    max-width: 90%;
  }

  .message-error {
    background: var(--color-error-soft);
    border-left: 3px solid var(--color-error);
    border-radius: var(--radius-md);
    padding: var(--space-3) var(--space-4);
    box-shadow: 0 0 20px rgba(239, 68, 68, 0.1);
  }

  .message-content {
    font-size: var(--size-md);
    line-height: var(--leading-relaxed);
    white-space: pre-wrap;
    word-break: break-word;
    color: var(--color-text);
  }

  /* ── Streaming cursor — blinking bar ──────────────── */
  .streaming .cursor {
    display: inline-block;
    width: 3px;
    height: 1.2em;
    background: var(--color-accent);
    border-radius: 2px;
    margin-left: 4px;
    vertical-align: middle;
    box-shadow: 0 0 10px var(--color-accent-glow);
    animation: cursor-blink 0.8s steps(2) infinite;
  }

  /* ── Tool calls — premium glass cards ─────────────── */
  .tool-calls {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin-top: var(--space-4);
  }

  .tool-call {
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-md);
    overflow: hidden;
    transition: all var(--transition-fast);
  }

  .tool-call:hover {
    border-color: var(--glass-border-hover);
    background: rgba(0, 0, 0, 0.35);
  }

  .tool-call.streaming {
    padding: 10px 14px;
    display: flex;
    align-items: center;
    gap: var(--space-3);
    border-color: var(--color-border-accent);
    box-shadow: 0 0 20px var(--color-accent-soft);
    position: relative;
    animation: streaming-pulse 2s ease-in-out infinite;
  }

  .tool-pulse {
    position: absolute;
    right: 14px;
    top: 50%;
    transform: translateY(-50%);
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--color-accent);
    box-shadow: 0 0 10px var(--color-accent);
    animation: breathe 1s ease-in-out infinite;
  }

  .tool-call summary {
    cursor: pointer;
    padding: 10px 14px;
    display: flex;
    align-items: center;
    gap: var(--space-3);
    list-style: none;
    color: var(--color-text-muted);
    user-select: none;
    transition: color var(--transition-fast);
  }

  .tool-call summary:hover {
    color: var(--color-text);
  }

  .tool-call summary::-webkit-details-marker { display: none; }

  .tool-icon {
    color: var(--color-accent);
    display: flex;
    align-items: center;
  }
  .tool-icon svg { width: 16px; height: 16px; }

  .tool-name {
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    font-weight: var(--weight-medium);
  }

  .tool-args {
    padding: var(--space-3);
    margin: 0;
    background: rgba(0, 0, 0, 0.4);
    border-top: 1px solid var(--glass-border);
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--color-text-dim);
    white-space: pre-wrap;
    max-height: 200px;
    overflow-y: auto;
  }

  /* ── Input Pill — the hero of the chat ────────────── */
  .chat-input-wrap {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    padding: var(--space-6) var(--space-5) var(--space-4);
    background: linear-gradient(0deg, var(--color-bg) 40%, transparent 100%);
    z-index: 100;
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  /* Glow behind the input — intensifies on focus */
  .chat-input-glow {
    position: absolute;
    top: 20px;
    width: 60%;
    height: 80px;
    background: var(--color-accent);
    filter: blur(80px);
    opacity: 0.08;
    border-radius: 50%;
    pointer-events: none;
    z-index: 0;
    transition: all var(--transition-slow);
  }

  .chat-input-pill {
    width: 100%;
    max-width: var(--content-max-width);
    background: var(--glass-bg-solid);
    backdrop-filter: var(--glass-blur-heavy);
    -webkit-backdrop-filter: var(--glass-blur-heavy);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-2xl);
    padding: 14px 20px;
    position: relative;
    z-index: 1;
    transition: all var(--transition-base);
    box-shadow: var(--shadow-lg), var(--shadow-inset);
  }

  .chat-input-pill:focus-within {
    border-color: var(--color-border-accent);
    box-shadow: var(--shadow-xl), 0 0 40px var(--color-accent-faint), var(--shadow-inset);
  }

  .chat-input-wrap:focus-within .chat-input-glow {
    opacity: 0.2;
    transform: scale(1.2);
  }

  /* Streaming state — the pill breathes */
  .is-streaming .chat-input-pill {
    border-color: var(--color-accent-soft);
    animation: streaming-pulse 3s ease-in-out infinite;
  }

  .input-meta-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: 8px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--color-border);
  }

  .meta-select {
    appearance: none;
    background: transparent;
    border: none;
    color: var(--color-text-dim);
    font-size: 11px;
    font-weight: var(--weight-bold);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wide);
    cursor: pointer;
    outline: none;
    transition: color var(--transition-fast);
    font-family: var(--font-mono);
  }

  .meta-select:hover, .meta-select:focus {
    color: var(--color-accent);
  }

  .meta-sep {
    color: var(--color-text-dim);
    font-size: 10px;
    opacity: 0.4;
  }

  .meta-model {
    background: transparent;
    border: none;
    color: var(--color-text-dim);
    font-size: 11px;
    font-family: var(--font-mono);
    outline: none;
    width: 140px;
    transition: color var(--transition-fast);
  }

  .meta-model::placeholder { color: var(--color-text-dim); }
  .meta-model:focus { color: var(--color-text); }

  .input-main-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  .input-message {
    flex: 1;
    background: transparent;
    border: none;
    color: var(--color-text);
    font-size: var(--size-lg);
    font-weight: var(--weight-light);
    outline: none;
    letter-spacing: var(--tracking-normal);
  }

  .input-message::placeholder { color: var(--color-text-faint); }
  .input-message:disabled { opacity: 0.5; cursor: not-allowed; }

  /* Send button — tactile, alive */
  .btn-send {
    width: 40px;
    height: 40px;
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    border: none;
    background: var(--color-text);
    color: var(--color-bg);
    cursor: pointer;
    transition: all var(--transition-spring);
    box-shadow: 0 4px 16px rgba(255, 255, 255, 0.08);
    position: relative;
    overflow: hidden;
  }

  .btn-send::before {
    content: '';
    position: absolute;
    inset: 0;
    background: var(--color-accent-gradient);
    opacity: 0;
    transition: opacity var(--transition-base);
    border-radius: inherit;
  }

  .btn-send svg {
    position: relative;
    z-index: 1;
    transition: transform var(--transition-spring);
  }

  .btn-send:hover:not(:disabled) {
    transform: scale(1.1) rotate(-5deg);
    box-shadow: 0 8px 28px var(--color-glow-strong);
  }

  .btn-send:hover:not(:disabled)::before {
    opacity: 1;
  }

  .btn-send:hover:not(:disabled) svg {
    transform: translateX(2px);
  }

  .btn-send:active:not(:disabled) {
    transform: scale(0.92);
    transition-duration: var(--transition-instant);
  }

  .btn-send:disabled {
    background: rgba(255, 255, 255, 0.08);
    color: rgba(255, 255, 255, 0.25);
    cursor: not-allowed;
    box-shadow: none;
  }

  .input-footer {
    width: 100%;
    max-width: var(--content-max-width);
    text-align: center;
  }

  .hint {
    font-size: 11px;
    color: var(--color-text-dim);
    margin-top: 12px;
    letter-spacing: var(--tracking-wide);
    font-family: var(--font-mono);
  }

  .hint .warn {
    color: var(--color-warn);
  }

  /* ── Empty state — welcoming, alive ───────────────── */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-12) var(--space-4);
    text-align: center;
    position: relative;
    animation: fade-in-scale var(--transition-slow) var(--ease-out-expo) both;
  }

  .empty-logo-glow {
    position: absolute;
    top: 45%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 200px;
    height: 200px;
    background: var(--color-accent);
    filter: blur(120px);
    opacity: 0.12;
    z-index: 0;
    animation: breathe-soft 6s ease-in-out infinite;
  }

  .empty-icon {
    margin-bottom: var(--space-5);
    color: var(--color-text);
    position: relative;
    z-index: 1;
  }

  .empty-icon svg {
    width: 56px;
    height: 56px;
    filter: drop-shadow(0 10px 30px rgba(0, 0, 0, 0.5));
    animation: breathe-soft 4s ease-in-out infinite;
  }

  .empty-state h3 {
    font-size: var(--size-2xl);
    font-weight: var(--weight-medium);
    color: var(--color-text);
    margin-bottom: var(--space-3);
    letter-spacing: var(--tracking-tight);
    position: relative;
    z-index: 1;
  }

  .empty-state p {
    color: var(--color-text-muted);
    font-size: var(--size-md);
    max-width: 400px;
    line-height: var(--leading-relaxed);
    position: relative;
    z-index: 1;
  }

  .empty-state a {
    color: var(--color-accent);
    text-decoration: none;
    transition: all var(--transition-fast);
  }

  .empty-state a:hover {
    color: var(--color-accent-hover);
    text-shadow: 0 0 12px var(--color-glow);
  }
</style>
