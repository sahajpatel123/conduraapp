<!--
  ChatV1 — the migrated chat route using the v1 design system.

  This route is the migration template. It reuses the existing stores
  (`conversation.svelte`, `daemon.svelte`, `settings.svelte`) and IPC
  for full functionality, but applies the v1 visual presentation:
    - v1 ChatSurface for the conversation column
    - v1 EmptyState for the empty case
    - v1 Input for the composer
    - v1 Hairline separators
    - v1 Button for actions

  To migrate existing routes:
    1. Keep all data layer (stores, IPC, daemon calls) intact
    2. Replace visual primitives with v1 equivalents
    3. Use semantic tokens via CSS custom properties
    4. Match density (compact for lists, spacious for reading)
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import { conversation } from '../stores/conversation.svelte';
  import { settings } from '../stores/settings.svelte';
  import { ipc } from '../ipc/client';
  import type { ProviderInfo } from '../ipc/types';

  import ChatSurface from '$components/v1/ChatSurface.svelte';
  import EmptyState from '$components/v1/EmptyState.svelte';
  import Input from '$components/v1/Input.svelte';
  import Button from '$components/v1/Button.svelte';
  import Chip from '$components/v1/Chip.svelte';
  import Surface from '$components/v1/Surface.svelte';
  import Hairline from '$components/v1/Hairline.svelte';
  import Inline from '$components/v1/Inline.svelte';

  let inputText = $state('');
  let providers = $state<ProviderInfo[]>([]);
  let selectedModel = $state('');

  function providerDisplayName(p: string): string {
    return p.charAt(0).toUpperCase() + p.slice(1);
  }

  onMount(async () => {
    try {
      providers = await ipc.providersList();
    } catch {
      providers = [];
    }
  });

  const modelOptions = $derived(
    providers.flatMap((p) =>
      p.models.map((m) => ({ value: `${p.name}:${m.id}`, label: `${p.name} · ${m.id}` }))
    )
  );

  function defaultProviderModel(): string {
    const cfg = settings.config;
    if (!cfg) return modelOptions[0]?.value ?? '';
    const entries = Object.entries(cfg.llm.providers);
    const enabled = entries.find(([, p]) => p.enabled);
    if (enabled) return `${enabled[0]}:${enabled[1].default_model}`;
    return modelOptions[0]?.value ?? '';
  }

  $effect(() => {
    if (!selectedModel && modelOptions.length > 0) {
      selectedModel = defaultProviderModel() || modelOptions[0].value;
    }
  });

  import type { Message } from '../ipc/types';

  // Transform conversation messages to v1 ChatSurface turns
  type TurnStatus = 'streaming' | 'paused' | 'done' | 'error';
  type Turn = {
    id: string;
    role: 'user' | 'agent' | 'system';
    content: string;
    timestamp: string;
    status?: TurnStatus;
  };

  const turns = $derived.by((): Turn[] => {
    const base: Turn[] = conversation.messages.map((turn: Message, i: number) => ({
      id: String(i),
      role: turn.role === 'assistant' ? ('agent' as const)
          : turn.role === 'user'     ? ('user' as const)
          : ('system' as const),
      content: turn.content,
      timestamp: '',
      status: 'done' as const,
    }));

    if (conversation.isStreaming) {
      base.push({
        id: 'streaming',
        role: 'agent' as const,
        content: conversation.streamingDelta || '…',
        timestamp: '',
        status: 'streaming' as const,
      });
    }

    return base;
  });

  let selectedProvider = $derived(selectedModel.split(':')[0] || '');

  function onInput(e: Event): void {
    inputText = (e.target as HTMLInputElement).value;
  }

  function onKey(e: KeyboardEvent): void {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      void send();
    }
  }

  async function send(): Promise<void> {
    const text = inputText.trim();
    if (!text || conversation.isStreaming) return;
    inputText = '';

    const [providerName, modelId] = selectedModel.split(':');
    if (!providerName || !modelId) return;

    try {
      await conversation.send(providerName, modelId, text);
    } catch (e) {
      console.error('send failed', e);
    }
  }
</script>

<div class="chat-route">
  <header class="chat-route__topbar">
    <Inline gap="3" align="center">
      <label class="chat-route__model-label" for="chat-v1-model">
        <span class="caption">Model</span>
      </label>
      <select
        id="chat-v1-model"
        class="chat-route__model"
        bind:value={selectedModel}
        aria-label="Model"
      >
        {#each modelOptions as opt}
          <option value={opt.value}>{opt.label}</option>
        {/each}
      </select>
      {#if selectedProvider}
        <Chip>{providerDisplayName(selectedProvider)}</Chip>
      {/if}
    </Inline>
  </header>

  <Hairline />

  <div class="chat-route__surface">
    {#if turns.length === 0}
      <EmptyState
        primary="Awaiting task."
        secondary="The agent is listening. Type when you're ready."
        voice="mono"
      >
        {#snippet children()}
          <Inline gap="2">
            <Chip onclick={() => { inputText = 'Summarize my last 3 emails'; }}>Summarize emails</Chip>
            <Chip onclick={() => { inputText = 'Open Safari and search'; }}>Open Safari</Chip>
            <Chip onclick={() => { inputText = 'Rename a file on my Desktop'; }}>Rename file</Chip>
          </Inline>
        {/snippet}
      </EmptyState>
    {:else}
      <ChatSurface {turns} />
    {/if}
  </div>

  <footer class="chat-route__composer">
    <Surface variant="raised" padding="3" radius="lg">
      <form class="chat-route__form" onsubmit={(e) => { e.preventDefault(); void send(); }}>
        <Input
          variant="sans"
          size="md"
          placeholder="Ask anything, or press ⌘K for commands"
          ariaLabel="Message"
          value={inputText}
          oninput={onInput}
          onkeydown={onKey}
          disabled={conversation.isStreaming}
        />
        <Button
          variant="primary"
          size="md"
          type="submit"
          disabled={!inputText.trim() || conversation.isStreaming}
          loading={conversation.isStreaming}
        >
          {conversation.isStreaming ? 'Thinking…' : 'Send'}
        </Button>
      </form>
    </Surface>
  </footer>
</div>

<style>
  .chat-route {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    background-color: var(--surface-base);
  }

  .chat-route__topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-3) var(--space-5);
    height: 56px;
    flex-shrink: 0;
    background-color: var(--surface-raised);
  }

  .chat-route__model-label {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .caption {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--content-tertiary);
  }

  .chat-route__model {
    height: 32px;
    padding: 0 var(--space-3);
    background-color: var(--surface-base);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-sm);
    color: var(--content-primary);
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    cursor: pointer;
    transition: border-color var(--duration-fast) var(--ease-standard);
  }
  .chat-route__model:focus-visible {
    outline: var(--border-focus) solid var(--border-focus-width, 2px);
    outline-offset: 2px;
    border-color: var(--border-focus);
  }

  .chat-route__surface {
    flex: 1;
    overflow-y: auto;
    background-color: var(--surface-base);
  }

  .chat-route__composer {
    flex-shrink: 0;
    padding: var(--space-3) var(--space-5);
    background-color: var(--surface-raised);
    border-top: 1px solid var(--border-subtle);
  }

  .chat-route__form {
    display: flex;
    gap: var(--space-3);
    align-items: stretch;
  }

  .chat-route__form :global(.input) {
    flex: 1;
    border: 1px solid var(--border-default);
    border-radius: var(--radius-sm);
    padding: 0 var(--space-3);
  }

  .chat-route__form :global(.input:focus-visible) {
    border-color: var(--border-focus);
    box-shadow: 0 0 0 2px var(--border-focus);
  }
</style>