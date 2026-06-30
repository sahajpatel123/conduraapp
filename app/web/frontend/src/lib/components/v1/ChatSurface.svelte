<!--
  ChatSurface — the editorial column.

  Per spec §8.1: full-height editorial column with mono timestamps on the
  left margin and prose flowing in proportional type. Hairline separators
  between turns; nothing else. No avatars, no bubbles, no reaction buttons.

  The agent's voice is serif ("as if written by hand"). The user's voice
  is sans. Mono timestamps always. Surface tint distinguishes them subtly.

  Props:
    turns    — array of {id, role, content, timestamp, status, pausedMs?}
    children — optional slot for empty-state
-->
<script lang="ts">
  import Hairline from './Hairline.svelte';
  import StreamingText from './StreamingText.svelte';

  type Role = 'user' | 'agent' | 'system';
  type Status = 'streaming' | 'paused' | 'done' | 'error';

  interface Turn {
    id: string;
    role: Role;
    content: string;
    timestamp: string;
    status?: Status;
    pausedMs?: number;
  }

  interface Props {
    turns?: Turn[];
    children?: import('svelte').Snippet;
  }

  let { turns = [], children }: Props = $props();
</script>

<article class="chat" role="log" aria-label="Conversation">
  {#if turns.length === 0}
    {#if children}
      {@render children()}
    {/if}
  {:else}
    {#each turns as turn, i}
      <div class="chat__turn chat__turn--{turn.role}" data-status={turn.status ?? 'done'}>
        <div class="chat__timestamp" aria-hidden="true">{turn.timestamp}</div>
        <div class="chat__body">
          {#if turn.role === 'agent'}
            <StreamingText
              text={turn.content}
              voice="serif"
              state={turn.status ?? 'done'}
              pausedMs={turn.pausedMs ?? 0}
            />
          {:else if turn.role === 'system'}
            <p class="chat__system">{turn.content}</p>
          {:else}
            <p class="chat__user">{turn.content}</p>
          {/if}
        </div>
      </div>
      {#if i < turns.length - 1}
        <Hairline variant="subtle" inset={true} />
      {/if}
    {/each}
  {/if}
</article>

<style>
  .chat {
    /* The editorial column — narrow, centered, generous */
    max-width: var(--container-prose);
    margin: 0 auto;
    padding: var(--space-9) var(--space-6);
    display: flex;
    flex-direction: column;
  }

  /* Each turn — timestamp on the left margin (mono), body in the column */
  .chat__turn {
    display: grid;
    grid-template-columns: 96px 1fr;
    gap: var(--space-5);
    padding: var(--space-5) 0;
    transition: background-color var(--duration-base) var(--ease-standard);
  }

  /* Subtle surface tint distinguishes user from agent.
     Per spec §11.1: agent on base, user on a slightly different tint. */
  .chat__turn--user {
    background-color: var(--paper-warm-50);
    border-radius: var(--radius-md);
    padding-left: var(--space-3);
    padding-right: var(--space-3);
    margin-left: calc(-1 * var(--space-3));
    margin-right: calc(-1 * var(--space-3));
  }

  .chat__turn--system {
    color: var(--content-tertiary);
    font-style: italic;
  }

  /* Timestamp — mono, tabular, muted */
  .chat__timestamp {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.02em;
    padding-top: var(--space-1);
    text-align: right;
  }

  /* Body — serif for agent, sans for user */
  .chat__body {
    min-width: 0;
  }

  .chat__user {
    font-family: var(--font-sans);
    font-size: var(--text-body-size);
    line-height: 1.6;
    color: var(--content-primary);
    margin: 0;
    /* User typing isn't on display — the message just is. */
  }

  .chat__system {
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    margin: 0;
  }

  /* Mobile — stack the timestamp above the body */
  @media (max-width: 720px) {
    .chat__turn {
      grid-template-columns: 1fr;
      gap: var(--space-2);
    }
    .chat__timestamp {
      text-align: left;
    }
  }
</style>