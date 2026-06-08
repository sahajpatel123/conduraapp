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
  <header class="chat-header">
    <h2>{conversation.currentTitle}</h2>
    <div class="chat-status">
      {#if conversation.isStreaming}
        <span class="streaming-pill">streaming…</span>
        <button class="btn btn-ghost" onclick={cancel}>Stop</button>
      {/if}
    </div>
  </header>

  <div class="chat-thread">
    {#each conversation.messages as msg, i (i)}
      <div class="message message-{msg.role}">
        <div class="message-role">{msg.role}</div>
        <div class="message-content">{msg.content}</div>
      </div>
    {/each}

    {#if conversation.isStreaming && conversation.streamingDelta}
      <div class="message message-assistant streaming">
        <div class="message-role">assistant</div>
        <div class="message-content">{conversation.streamingDelta}<span class="cursor">▍</span></div>
      </div>
    {/if}

    {#if conversation.streamingError}
      <div class="message message-error">
        <div class="message-role">error</div>
        <div class="message-content">{conversation.streamingError}</div>
      </div>
    {/if}

    <div bind:this={scrollAnchor}></div>
  </div>

  <footer class="chat-input">
    <div class="input-row">
      <select bind:value={selectedProvider} class="select">
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
      <input
        type="text"
        bind:value={selectedModel}
        placeholder="model"
        class="input input-model"
      />
      <input
        type="text"
        bind:value={inputText}
        onkeydown={onKeydown}
        placeholder="Ask anything… (Enter to send, Shift+Enter for newline)"
        class="input input-message"
        disabled={conversation.isStreaming}
      />
      <button
        class="btn btn-primary"
        onclick={send}
        disabled={!inputText.trim() || conversation.isStreaming}
      >
        Send
      </button>
    </div>
    <p class="hint">
      {#if !daemon.connected}
        <span class="warn">⚠ Not connected to the daemon.</span>
      {:else}
        Connected · streaming via SSE
      {/if}
    </p>
  </footer>
</div>

<!-- The model field is bound to selectedModel directly. -->

<style>
  .chat-page {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--color-bg);
  }

  .chat-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-4) var(--space-5);
    border-bottom: 1px solid var(--color-border);
    background: var(--color-bg-elevated);
  }
  .chat-header h2 {
    font-size: var(--size-lg);
    font-weight: 600;
  }
  .chat-status {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .streaming-pill {
    background: var(--color-accent-soft);
    color: var(--color-accent);
    padding: 2px 8px;
    border-radius: var(--radius-pill);
    font-size: var(--size-xs);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .chat-thread {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-5);
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  .message {
    max-width: 80%;
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    background: var(--color-bg-elevated);
    border: 1px solid var(--color-border);
  }
  .message-user {
    align-self: flex-end;
    background: var(--color-accent-soft);
    border-color: var(--color-accent);
  }
  .message-assistant {
    align-self: flex-start;
  }
  .message-error {
    align-self: stretch;
    max-width: 100%;
    background: rgba(248, 113, 113, 0.1);
    border-color: var(--color-error);
  }
  .message-role {
    font-size: var(--size-xs);
    color: var(--color-text-faint);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: var(--space-1);
  }
  .message-content {
    font-size: var(--size-md);
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-word;
  }
  .streaming .cursor {
    color: var(--color-accent);
    animation: blink 1s steps(2) infinite;
  }
  @keyframes blink {
    50% { opacity: 0; }
  }

  .chat-input {
    padding: var(--space-4) var(--space-5);
    border-top: 1px solid var(--color-border);
    background: var(--color-bg-elevated);
  }
  .input-row {
    display: flex;
    gap: var(--space-2);
    align-items: center;
  }
  .select,
  .input {
    background: var(--color-bg);
    border: 1px solid var(--color-border);
    color: var(--color-text);
    padding: 10px 12px;
    border-radius: var(--radius-md);
    font-size: var(--size-md);
  }
  .input:focus,
  .select:focus {
    outline: none;
    border-color: var(--color-accent);
    box-shadow: var(--shadow-focus);
  }
  .input-model {
    width: 180px;
  }
  .input-message {
    flex: 1;
  }
  .btn {
    padding: 10px 18px;
    border-radius: var(--radius-md);
    font-size: var(--size-md);
    font-weight: 500;
    border: 1px solid transparent;
  }
  .btn-primary {
    background: var(--color-accent);
    color: white;
  }
  .btn-primary:hover:not(:disabled) {
    background: var(--color-accent-hover);
  }
  .btn-ghost {
    background: transparent;
    color: var(--color-text-muted);
    border: 1px solid var(--color-border);
  }
  .btn-ghost:hover {
    color: var(--color-text);
    border-color: var(--color-border-strong);
  }
  .hint {
    font-size: var(--size-xs);
    color: var(--color-text-faint);
    margin-top: var(--space-2);
  }
  .hint .warn {
    color: var(--color-warn);
  }
</style>
