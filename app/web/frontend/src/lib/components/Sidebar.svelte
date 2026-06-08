<script lang="ts">
  import { conversation } from '../stores/conversation.svelte'

  async function startNew(): Promise<void> {
    await conversation.createNew('New conversation')
  }

  async function openExisting(id: number): Promise<void> {
    await conversation.open(id)
  }

  async function deleteCurrent(): Promise<void> {
    if (confirm('Delete this conversation? This cannot be undone.')) {
      await conversation.deleteCurrent()
    }
  }
</script>

<aside class="sidebar">
  <div class="sidebar-header">
    <h3>Conversations</h3>
    <button class="btn btn-ghost" onclick={startNew}>+ New</button>
  </div>

  <div class="conversation-list">
    {#if conversation.conversations.length === 0}
      <p class="empty">No conversations yet.<br />Click <strong>+ New</strong> to start.</p>
    {/if}
    {#each conversation.conversations as c (c.id)}
      <button
        class="conversation-item"
        class:active={c.id === conversation.currentID}
        onclick={() => openExisting(c.id)}
      >
        <span class="title">{c.title}</span>
        <span class="meta">{c.message_count} msg · {new Date(c.updated_at).toLocaleDateString()}</span>
      </button>
    {/each}
  </div>

  {#if conversation.currentID}
    <div class="sidebar-footer">
      <button class="btn btn-ghost" onclick={deleteCurrent}>Delete current</button>
    </div>
  {/if}
</aside>

<style>
  .sidebar {
    width: 240px;
    background: var(--color-bg-elevated);
    border-right: 1px solid var(--color-border);
    display: flex;
    flex-direction: column;
    height: 100%;
  }
  .sidebar-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-4);
    border-bottom: 1px solid var(--color-border);
  }
  .sidebar-header h3 {
    font-size: var(--size-md);
    font-weight: 600;
  }
  .btn {
    padding: 4px 10px;
    border-radius: var(--radius-md);
    font-size: var(--size-sm);
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
  .conversation-list {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-2);
  }
  .conversation-item {
    display: flex;
    flex-direction: column;
    width: 100%;
    text-align: left;
    padding: var(--space-3);
    border-radius: var(--radius-md);
    background: transparent;
    color: var(--color-text);
    border: 1px solid transparent;
    margin-bottom: 2px;
    transition: background var(--transition-fast);
  }
  .conversation-item:hover {
    background: var(--color-bg-hover);
  }
  .conversation-item.active {
    background: var(--color-accent-soft);
    border-color: var(--color-accent);
  }
  .title {
    font-size: var(--size-md);
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .meta {
    font-size: var(--size-xs);
    color: var(--color-text-faint);
    margin-top: 2px;
  }
  .empty {
    color: var(--color-text-faint);
    font-size: var(--size-sm);
    text-align: center;
    padding: var(--space-5) var(--space-3);
  }
  .sidebar-footer {
    padding: var(--space-3);
    border-top: 1px solid var(--color-border);
  }
</style>
