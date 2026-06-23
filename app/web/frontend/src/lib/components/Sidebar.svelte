<script lang="ts">
  import { conversation } from '../stores/conversation.svelte'
  import { account } from '../stores/account.svelte'
  import { onMount } from 'svelte'
  import SignInPanel from './SignInPanel.svelte'
  import AccountMenu from './AccountMenu.svelte'
  import { t } from '../i18n'

  let currentHash: string = $state('')
  let showSignIn = $state(false)
  let showAccountMenu = $state(false)

  onMount(() => {
    currentHash = window.location.hash || '#/'
    const onHashChange = () => {
      currentHash = window.location.hash || '#/'
    }
    window.addEventListener('hashchange', onHashChange)
    void account.checkStatus()
    return () => window.removeEventListener('hashchange', onHashChange)
  })

  async function startNew(): Promise<void> {
    await conversation.createNew(t('sidebar.new_conversation'))
  }

  async function openExisting(id: number): Promise<void> {
    await conversation.open(id)
  }

  async function deleteCurrent(): Promise<void> {
    if (confirm(t('sidebar.delete_confirm'))) {
      await conversation.deleteCurrent()
    }
  }
</script>

<aside class="sidebar">
  <!-- Icon Rail -->
  <nav class="icon-rail">
    <div class="rail-glow"></div>
    <div class="rail-top">
      <a
        href="#/"
        class="rail-icon"
        class:active={currentHash === '#/' || currentHash === '#' || currentHash === ''}
        title={t('sidebar.nav.chat')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M4 4h12a2 2 0 012 2v7a2 2 0 01-2 2H7l-4 3V6a2 2 0 012-2z"/></svg>
      </a>
      <a
        href="#/audit"
        class="rail-icon"
        class:active={currentHash === '#/audit'}
        title={t('sidebar.nav.audit')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M10 2l7 3v5c0 4-3 6.5-7 8-4-1.5-7-4-7-8V5l7-3z"/></svg>
      </a>
      <a
        href="#/replay"
        class="rail-icon"
        class:active={currentHash === '#/replay'}
        title={t('sidebar.nav.replay')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="10" cy="10" r="7"/><path d="M8 7l5 3-5 3V7z"/></svg>
      </a>
      <a
        href="#/hub"
        class="rail-icon"
        class:active={currentHash === '#/hub'}
        title={t('sidebar.nav.skills_hub')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 7l7-4 7 4-7 4-7-4zM3 7v6l7 4M17 7v6l-7 4"/></svg>
      </a>
      <a
        href="#/skills"
        class="rail-icon"
        class:active={currentHash === '#/skills'}
        title={t('sidebar.nav.skills')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M4 4l6 2 6-2-2 12-4 2-4-2L4 4z"/><path d="M10 6v12"/></svg>
      </a>
      <a
        href="#/sync"
        class="rail-icon"
        class:active={currentHash === '#/sync'}
        title={t('sidebar.nav.sync')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M5 10a5 5 0 019-3l1 1m0-3v3h-3M15 10a5 5 0 01-9 3l-1-1m0 3v-3h3"/></svg>
      </a>
      <a
        href="#/channels"
        class="rail-icon"
        class:active={currentHash === '#/channels'}
        title={t('sidebar.nav.channels')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 5h14v9H8l-4 3v-3H3V5z"/></svg>
      </a>
      <a
        href="#/delegation"
        class="rail-icon"
        class:active={currentHash === '#/delegation'}
        title={t('sidebar.nav.delegation')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="6" cy="6" r="2.5"/><circle cx="14" cy="6" r="2.5"/><circle cx="10" cy="14" r="2.5"/><path d="M6 8.5v2M14 8.5v2M10 5v6.5"/></svg>
      </a>
      <a
        href="#/settings"
        class="rail-icon"
        class:active={currentHash === '#/settings'}
        title={t('sidebar.nav.settings')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="10" cy="10" r="3"/><path d="M10 1v2m0 14v2m-7-9h2m14 0h2m-3.5-5.5-1.4 1.4m-8.2 8.2-1.4 1.4m0-11-1.4 1.4m8.2 8.2 1.4 1.4"/></svg>
      </a>
      <a
        href="#/about"
        class="rail-icon"
        class:active={currentHash === '#/about'}
        title={t('sidebar.nav.about')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="10" cy="10" r="8"/><path d="M10 9v4m0-7h0"/></svg>
      </a>
    </div>

    <div class="rail-bottom">
      <button class="rail-new-btn" onclick={startNew} title={t('sidebar.new_conversation')}>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M10 4v12m-6-6h12"/></svg>
      </button>
    </div>
  </nav>

  <!-- Conversation Drawer -->
  <div class="drawer">
    <div class="drawer-header">
      <button class="new-conv-btn" onclick={startNew}>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M10 4v12m-6-6h12"/></svg>
        <span>{t('sidebar.new_conversation')}</span>
      </button>
      <span class="drawer-label">{t('sidebar.history')}</span>
    </div>

    <div class="conversation-list">
      {#if conversation.conversations.length === 0}
        <div class="empty">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M4 5h16a2 2 0 012 2v8a2 2 0 01-2 2H8l-5 4V7a2 2 0 012-2z"/><path d="M9 10h6M9 13h4"/></svg>
          <p>{t('sidebar.empty')}</p>
        </div>
      {/if}
      {#each conversation.conversations as c (c.id)}
        <button
          class="conversation-item"
          class:active={c.id === conversation.currentID}
          onclick={() => openExisting(c.id)}
        >
          <span class="title">{c.title}</span>
          <span class="meta">{t('sidebar.msg_count', c.message_count)} · {new Date(c.updated_at).toLocaleDateString()}</span>
        </button>
      {/each}
    </div>

    {#if conversation.currentID}
      <div class="drawer-footer">
        <button class="btn-delete" onclick={deleteCurrent}>
          <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" class="delete-icon"><path d="M4 6h12M8 6V4h4v2m-7 0v10a1 1 0 001 1h6a1 1 0 001-1V6"/></svg>
          {t('sidebar.delete_current')}
        </button>
      </div>
    {/if}

    <!-- Account footer (Phase 14B) -->
    <div class="account-footer">
      {#if account.isSignedIn}
        <button class="account-chip" onclick={() => (showAccountMenu = true)} title={t('sidebar.account')}>
          {#if account.avatarURL}
            <img class="chip-avatar" src={account.avatarURL} alt="" />
          {:else}
            <span class="chip-avatar fallback">{(account.displayName || '?').charAt(0).toUpperCase()}</span>
          {/if}
          <span class="chip-email">{account.email || account.displayName}</span>
        </button>
      {:else}
        <button class="signin-link" onclick={() => (showSignIn = true)}>{t('sidebar.signin')}</button>
      {/if}
    </div>
  </div>
</aside>

{#if showSignIn}
  <SignInPanel onClose={() => (showSignIn = false)} />
{/if}
{#if showAccountMenu}
  <AccountMenu onClose={() => (showAccountMenu = false)} />
{/if}

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
    width: var(--sidebar-rail-width);
    min-width: var(--sidebar-rail-width);
    background: var(--color-bg);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-4) 0;
    z-index: 2;
    position: relative;
  }

  /* Subtle gradient glow at the top of the rail */
  .rail-glow {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 120px;
    background: radial-gradient(ellipse at top, var(--color-accent-faint), transparent 70%);
    pointer-events: none;
    opacity: 0.6;
  }

  /* Gradient separator between rail and drawer */
  .icon-rail::after {
    content: '';
    position: absolute;
    top: 10%;
    bottom: 10%;
    right: 0;
    width: 1px;
    background: linear-gradient(180deg, transparent, var(--color-border-strong) 30%, var(--color-border-strong) 70%, transparent);
  }

  .rail-top {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-1);
    position: relative;
    z-index: 1;
  }

  .rail-icon {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    border-radius: var(--radius-md);
    color: var(--color-text-faint);
    text-decoration: none;
    transition: color var(--transition-fast), transform var(--transition-spring),
      background var(--transition-fast), box-shadow var(--transition-fast);
    cursor: pointer;
  }

  .rail-icon svg {
    width: 20px;
    height: 20px;
    position: relative;
    z-index: 1;
    transition: transform var(--transition-spring);
  }

  .rail-icon:hover {
    color: var(--color-text);
    transform: scale(1.08);
    background: var(--color-bg-hover);
    box-shadow: var(--shadow-glow);
  }

  .rail-icon:hover svg {
    transform: scale(1.05);
  }

  /* Active indicator — accent bar with spring animation */
  .active-indicator {
    position: absolute;
    left: -2px;
    top: 50%;
    transform: translateY(-50%) scaleY(0);
    width: 3px;
    height: 22px;
    border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
    background: var(--color-accent-gradient);
    box-shadow: 0 0 8px var(--color-glow);
    transition: transform var(--transition-spring);
  }

  .rail-icon.active .active-indicator {
    transform: translateY(-50%) scaleY(1);
  }

  .rail-icon.active {
    color: var(--color-accent);
    background: var(--color-accent-faint);
  }

  .rail-icon.active:hover {
    color: var(--color-accent-hover);
    transform: scale(1.08);
  }

  /* New conversation button in rail — gradient pill */
  .rail-bottom {
    display: flex;
    flex-direction: column;
    align-items: center;
    position: relative;
    z-index: 1;
  }

  .rail-new-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    border-radius: var(--radius-pill);
    background: var(--color-accent-gradient);
    border: none;
    color: #fff;
    cursor: pointer;
    transition: transform var(--transition-spring), box-shadow var(--transition-base);
    box-shadow: var(--shadow-sm);
  }

  .rail-new-btn svg {
    width: 18px;
    height: 18px;
  }

  .rail-new-btn:hover {
    transform: scale(1.1);
    box-shadow: var(--shadow-glow-strong);
  }

  .rail-new-btn:active {
    transform: scale(0.95);
  }

  /* ── Conversation Drawer ──────────────────────── */
  .drawer {
    width: var(--sidebar-drawer-width);
    min-width: var(--sidebar-drawer-width);
    background: var(--color-bg-elevated);
    backdrop-filter: var(--glass-blur);
    -webkit-backdrop-filter: var(--glass-blur);
    border-right: 1px solid var(--color-border);
    display: flex;
    flex-direction: column;
    height: 100%;
    z-index: 1;
  }

  .drawer-header {
    padding: var(--space-4) var(--space-3) var(--space-3);
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  /* Premium new conversation pill */
  .new-conv-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    width: 100%;
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-md);
    background: var(--color-accent-gradient);
    border: none;
    color: #fff;
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    font-family: var(--font-sans);
    cursor: pointer;
    transition: transform var(--transition-base), box-shadow var(--transition-base);
    box-shadow: var(--shadow-sm);
  }

  .new-conv-btn svg {
    width: 16px;
    height: 16px;
  }

  .new-conv-btn:hover {
    transform: translateY(-1px);
    box-shadow: var(--shadow-glow);
  }

  .new-conv-btn:active {
    transform: translateY(0);
    box-shadow: var(--shadow-sm);
  }

  .drawer-label {
    font-size: var(--size-2xs);
    font-weight: var(--weight-semibold);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
    color: var(--color-text-faint);
    padding-left: var(--space-1);
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
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-md);
    background: transparent;
    color: var(--color-text);
    border: 1px solid transparent;
    cursor: pointer;
    transition: background var(--transition-fast), border-color var(--transition-fast),
      transform var(--transition-fast), box-shadow var(--transition-fast);
  }

  .conversation-item:hover {
    background: var(--glass-bg-hover);
    border-color: var(--glass-border);
    transform: translateY(-1px);
    box-shadow: var(--shadow-xs);
  }

  .conversation-item.active {
    background: var(--color-accent-faint);
    border-color: var(--color-border-accent);
    box-shadow: inset 2px 0 0 var(--color-accent), var(--shadow-xs);
  }

  .title {
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    line-height: var(--leading-tight);
  }

  .meta {
    font-size: var(--size-2xs);
    color: var(--color-text-faint);
    margin-top: 2px;
    line-height: 1.3;
  }

  /* Empty state */
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: var(--space-8) var(--space-3);
    color: var(--color-text-faint);
  }

  .empty svg {
    width: 32px;
    height: 32px;
    opacity: 0.4;
    margin-bottom: var(--space-3);
  }

  .empty p {
    font-size: var(--size-xs);
    line-height: var(--leading-relaxed);
    max-width: 180px;
  }

  /* ── Drawer Footer ───────────────────────────── */
  .drawer-footer {
    padding: var(--space-2) var(--space-3);
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
    color: var(--color-error);
    background: var(--color-error-soft);
    border-color: rgba(248, 113, 113, 0.2);
  }

  .delete-icon {
    width: 14px;
    height: 14px;
    flex-shrink: 0;
  }

  /* ── Account Footer — subtle chip ─────────────── */
  .account-footer {
    padding: var(--space-2) var(--space-3);
    border-top: 1px solid var(--color-border);
  }

  .signin-link {
    width: 100%;
    text-align: left;
    background: transparent;
    border: none;
    color: var(--color-text-faint);
    font-size: var(--size-xs);
    padding: var(--space-2) var(--space-1);
    cursor: pointer;
    transition: color var(--transition-fast);
  }

  .signin-link:hover {
    color: var(--color-accent);
  }

  .account-chip {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    width: 100%;
    padding: var(--space-2);
    border-radius: var(--radius-pill);
    background: transparent;
    border: 1px solid transparent;
    color: var(--color-text);
    cursor: pointer;
    transition: background var(--transition-fast), border-color var(--transition-fast);
  }

  .account-chip:hover {
    background: var(--glass-bg-hover);
    border-color: var(--glass-border);
  }

  .chip-avatar {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    flex-shrink: 0;
    object-fit: cover;
  }

  .chip-avatar.fallback {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--color-accent-gradient);
    color: white;
    font-weight: var(--weight-semibold);
    font-size: var(--size-2xs);
  }

  .chip-email {
    font-size: var(--size-2xs);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--color-text-muted);
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

