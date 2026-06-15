<script lang="ts">
  import { conversation } from '../stores/conversation.svelte'
  import { onMount } from 'svelte'

  let currentHash: string = $state('')

  onMount(() => {
    currentHash = window.location.hash || '#/'
    const onHashChange = () => {
      currentHash = window.location.hash || '#/'
    }
    window.addEventListener('hashchange', onHashChange)
    return () => window.removeEventListener('hashchange', onHashChange)
  })

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
  <!-- Icon Rail -->
  <nav class="icon-rail">
    <div class="rail-top">
      <a
        href="#/"
        class="rail-icon"
        class:active={currentHash === '#/' || currentHash === '#' || currentHash === ''}
        title="Chat"
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M4 4h12a2 2 0 012 2v7a2 2 0 01-2 2H7l-4 3V6a2 2 0 012-2z"/></svg>
      </a>
      <a
        href="#/audit"
        class="rail-icon"
        class:active={currentHash === '#/audit'}
        title="Audit"
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M10 2l7 3v5c0 4-3 6.5-7 8-4-1.5-7-4-7-8V5l7-3z"/></svg>
      </a>
      <a
        href="#/replay"
        class="rail-icon"
        class:active={currentHash === '#/replay'}
        title="Replay"
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="10" cy="10" r="7"/><path d="M8 7l5 3-5 3V7z"/></svg>
      </a>
      <a
        href="#/hub"
        class="rail-icon"
        class:active={currentHash === '#/hub'}
        title="Skills Hub"
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 7l7-4 7 4-7 4-7-4zM3 7v6l7 4M17 7v6l-7 4"/></svg>
      </a>
      <a
        href="#/skills"
        class="rail-icon"
        class:active={currentHash === '#/skills'}
        title="Skills"
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M4 4l6 2 6-2-2 12-4 2-4-2L4 4z"/><path d="M10 6v12"/></svg>
      </a>
      <a
        href="#/sync"
        class="rail-icon"
        class:active={currentHash === '#/sync'}
        title="P2P Sync"
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M5 10a5 5 0 019-3l1 1m0-3v3h-3M15 10a5 5 0 01-9 3l-1-1m0 3v-3h3"/></svg>
      </a>
      <a
        href="#/settings"
        class="rail-icon"
        class:active={currentHash === '#/settings'}
        title="Settings"
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="10" cy="10" r="3"/><path d="M10 1v2m0 14v2m-7-9h2m14 0h2m-3.5-5.5-1.4 1.4m-8.2 8.2-1.4 1.4m0-11-1.4 1.4m8.2 8.2 1.4 1.4"/></svg>
      </a>
      <a
        href="#/about"
        class="rail-icon"
        class:active={currentHash === '#/about'}
        title="About"
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="10" cy="10" r="8"/><path d="M10 9v4m0-7h0"/></svg>
      </a>
    </div>

    <div class="rail-bottom">
      <button class="rail-new-btn" onclick={startNew} title="New conversation">
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M10 4v12m-6-6h12"/></svg>
      </button>
    </div>
  </nav>

  <!-- Conversation Drawer -->
  <div class="drawer">
    <div class="drawer-header">
      <span class="drawer-label">History</span>
    </div>

    <div class="conversation-list">
      {#if conversation.conversations.length === 0}
        <p class="empty">No conversations yet.<br />Click <strong>+</strong> to start.</p>
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
      <div class="drawer-footer">
        <button class="btn-delete" onclick={deleteCurrent}>
          <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" class="delete-icon"><path d="M4 6h12M8 6V4h4v2m-7 0v10a1 1 0 001 1h6a1 1 0 001-1V6"/></svg>
          Delete current
        </button>
      </div>
    {/if}
  </div>
</aside>

<style>
  /* ── Layout Shell ─────────────────────────────── */
  .sidebar {
    display: flex;
    flex-direction: row;
    height: 100%;
    flex-shrink: 0;
  }

  /* ── Icon Rail ────────────────────────────────── */
  .icon-rail {
    width: 64px;
    min-width: 64px;
    background: var(--color-bg);
    border-right: 1px solid var(--color-border);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-4) 0;
    z-index: 2;
  }

  .rail-top {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-2);
  }

  .rail-icon {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 42px;
    height: 42px;
    border-radius: var(--radius-md);
    color: var(--color-text-faint);
    text-decoration: none;
    transition: color var(--transition-fast), transform var(--transition-fast),
      background var(--transition-fast);
    cursor: pointer;
  }

  .rail-icon svg {
    width: 20px;
    height: 20px;
    position: relative;
    z-index: 1;
  }

  .rail-icon:hover {
    color: var(--color-text-muted);
    transform: scale(1.1);
    background: var(--color-bg-hover);
  }

  /* Active indicator — thin accent bar on the left */
  .active-indicator {
    position: absolute;
    left: 0;
    top: 50%;
    transform: translateY(-50%) scaleY(0);
    width: 3px;
    height: 20px;
    border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
    background: var(--color-accent);
    transition: transform var(--transition-base);
  }

  .rail-icon.active .active-indicator {
    transform: translateY(-50%) scaleY(1);
  }

  .rail-icon.active {
    color: var(--color-accent);
  }

  .rail-icon.active:hover {
    color: var(--color-accent);
    transform: none;
  }

  /* New conversation button */
  .rail-bottom {
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .rail-new-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 38px;
    height: 38px;
    border-radius: var(--radius-pill);
    background: transparent;
    border: 1px solid var(--color-border);
    color: var(--color-text-faint);
    cursor: pointer;
    transition: color var(--transition-fast), transform var(--transition-fast),
      border-color var(--transition-fast), background var(--transition-fast);
  }

  .rail-new-btn svg {
    width: 18px;
    height: 18px;
  }

  .rail-new-btn:hover {
    color: var(--color-accent);
    border-color: var(--color-accent);
    transform: scale(1.1);
    background: var(--color-accent-soft);
  }

  /* ── Conversation Drawer ──────────────────────── */
  .drawer {
    width: 220px;
    min-width: 220px;
    background: var(--color-bg-elevated);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border-right: 1px solid var(--color-border);
    display: flex;
    flex-direction: column;
    height: 100%;
    z-index: 1;
  }

  .drawer-header {
    padding: var(--space-4) var(--space-4) var(--space-3);
  }

  .drawer-label {
    font-size: var(--size-xs);
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--color-text-faint);
  }

  /* ── Conversation List ────────────────────────── */
  .conversation-list {
    flex: 1;
    overflow-y: auto;
    padding: 0 var(--space-2) var(--space-2);
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
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
    border-left: 3px solid transparent;
    cursor: pointer;
    transition: background var(--transition-fast), border-color var(--transition-fast),
      transform var(--transition-fast), box-shadow var(--transition-fast);
  }

  .conversation-item:hover {
    background: var(--color-bg-hover);
    transform: translateY(-1px);
    box-shadow: var(--shadow-sm);
  }

  .conversation-item.active {
    background: var(--color-accent-soft);
    border-left-color: var(--color-accent);
    box-shadow: inset 0 0 12px var(--color-accent-soft), var(--shadow-sm);
  }

  .title {
    font-size: var(--size-sm);
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    line-height: 1.4;
  }

  .meta {
    font-size: var(--size-xs);
    color: var(--color-text-faint);
    margin-top: 2px;
    line-height: 1.3;
  }

  .empty {
    color: var(--color-text-faint);
    font-size: var(--size-sm);
    text-align: center;
    padding: var(--space-5) var(--space-3);
    line-height: 1.6;
  }

  /* ── Drawer Footer ───────────────────────────── */
  .drawer-footer {
    padding: var(--space-3);
    border-top: 1px solid var(--color-border);
  }

  .btn-delete {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    width: 100%;
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-md);
    font-size: var(--size-sm);
    background: transparent;
    color: var(--color-text-faint);
    border: 1px solid transparent;
    cursor: pointer;
    transition: color var(--transition-fast), background var(--transition-fast),
      border-color var(--transition-fast);
  }

  .btn-delete:hover {
    color: var(--color-danger);
    background: rgba(239, 68, 68, 0.08);
    border-color: rgba(239, 68, 68, 0.2);
  }

  .delete-icon {
    width: 14px;
    height: 14px;
    flex-shrink: 0;
  }

  /* ── Scrollbar ────────────────────────────────── */
  .conversation-list::-webkit-scrollbar {
    width: 4px;
  }

  .conversation-list::-webkit-scrollbar-track {
    background: transparent;
  }

  .conversation-list::-webkit-scrollbar-thumb {
    background: var(--color-border);
    border-radius: var(--radius-pill);
  }

  .conversation-list::-webkit-scrollbar-thumb:hover {
    background: var(--color-border-strong);
  }
</style>

