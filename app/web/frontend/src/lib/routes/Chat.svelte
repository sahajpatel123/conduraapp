<script lang="ts">
  import { conversation } from '../stores/conversation.svelte'
  import { daemon } from '../stores/daemon.svelte'
  import { settings } from '../stores/settings.svelte'

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
        <span class="streaming-pill">streaming…</span>
        <button class="btn-stop" onclick={cancel}>
          <svg viewBox="0 0 16 16" fill="currentColor" width="12" height="12"><rect x="3" y="3" width="10" height="10" rx="2" /></svg>
          Stop
        </button>
      </div>
    {/if}
  </header>

  <!-- Message thread -->
  <div class="chat-thread">
    <div class="thread-inner">
      {#each conversation.messages as msg, i (i)}
        <div class="message message-{msg.role}">
          <div class="message-content">{msg.content}</div>
        </div>
      {/each}

      {#if conversation.isStreaming && conversation.streamingDelta}
        <div class="message message-assistant streaming">
          <div class="message-content">{conversation.streamingDelta}<span class="cursor">▍</span></div>
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
          placeholder="Ask anything…"
          class="input-message"
          disabled={conversation.isStreaming}
        />
        <button
          class="btn-send"
          onclick={send}
          disabled={!inputText.trim() || conversation.isStreaming}
        >
          <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><path d="M4 10h12m-5-5 5 5-5 5"/></svg>
        </button>
      </div>
      <p class="hint">
        {#if !daemon.connected}
          <span class="warn">⚠ Not connected to the daemon.</span>
        {:else}
          Enter to send · Shift+Enter for newline
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
    font-size: var(--size-sm);
    font-weight: 500;
    color: var(--color-text-faint);
    text-transform: uppercase;
    letter-spacing: 0.1em;
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
    font-size: var(--size-xs);
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    box-shadow: 0 0 12px var(--color-glow);
  }
  .btn-stop {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    background: transparent;
    color: var(--color-text-muted);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-pill);
    padding: 4px 12px;
    font-size: var(--size-xs);
    font-weight: 500;
    cursor: pointer;
    transition: all var(--transition-fast);
  }
  .btn-stop:hover {
    color: var(--color-error);
    border-color: var(--color-error);
    background: rgba(248, 113, 113, 0.06);
  }

  /* ── Thread ─────────────────────────────────────────── */
  .chat-thread {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-4) var(--space-5);
  }
  .thread-inner {
    max-width: 720px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  /* ── Messages ───────────────────────────────────────── */
  @keyframes messageIn {
    from {
      opacity: 0;
      transform: translateY(8px);
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
    animation: messageIn 0.3s var(--ease-out-expo);
  }
  .message-user {
    align-self: flex-end;
    background: linear-gradient(135deg, rgba(99, 102, 241, 0.08), rgba(139, 92, 246, 0.08));
    border-left: 2px solid var(--color-accent);
    border-image: var(--color-accent-gradient) 1;
    border-image-slice: 1;
    border-width: 0 0 0 2px;
    border-style: solid;
  }
  .message-assistant {
    align-self: flex-start;
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
  }
  .message-error {
    align-self: stretch;
    max-width: 100%;
    background: rgba(248, 113, 113, 0.06);
    border-left: 2px solid var(--color-error);
  }
  .message-content {
    font-size: var(--size-md);
    line-height: 1.7;
    white-space: pre-wrap;
    word-break: break-word;
    color: var(--color-text);
  }

  /* ── Streaming cursor ───────────────────────────────── */
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.3; }
  }
  .streaming .cursor {
    color: var(--color-accent);
    animation: pulse 1.2s ease-in-out infinite;
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
  }
  .chat-input-pill:focus-within {
    border-color: var(--color-accent);
    box-shadow: var(--shadow-glow);
  }

  /* Streaming glow pulse on the pill */
  @keyframes glowPulse {
    0%, 100% { box-shadow: 0 0 12px var(--color-glow); }
    50% { box-shadow: 0 0 24px var(--color-glow-strong); }
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
  }
  .meta-select:hover,
  .meta-select:focus {
    color: var(--color-text-muted);
  }
  .meta-sep {
    color: var(--color-text-faint);
    font-size: var(--size-xs);
    opacity: 0.5;
    user-select: none;
  }
  .meta-model {
    background: transparent;
    border: none;
    color: var(--color-text-faint);
    font-size: var(--size-xs);
    font-family: var(--font-sans);
    outline: none;
    width: 140px;
    padding: 2px 0;
  }
  .meta-model::placeholder {
    color: var(--color-text-faint);
    opacity: 0.6;
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
    transition: opacity var(--transition-fast), transform var(--transition-fast);
  }
  .btn-send:hover:not(:disabled) {
    transform: scale(1.05);
    box-shadow: var(--shadow-glow);
  }
  .btn-send:active:not(:disabled) {
    transform: scale(0.95);
  }
  .btn-send:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  /* Hint */
  .hint {
    font-size: var(--size-xs);
    color: var(--color-text-faint);
    margin-top: var(--space-2);
    padding-left: var(--space-1);
    opacity: 0.7;
  }
  .hint .warn {
    color: var(--color-warn);
    opacity: 1;
  }
</style>
