<script lang="ts">
  import { conversation } from '../stores/conversation.svelte'
  import { account } from '../stores/account.svelte'
  import { overlay } from '../stores/overlay.svelte'
  import { notifications } from '../stores/notifications.svelte'
  import { onMount } from 'svelte'
  import SignInPanel from './SignInPanel.svelte'
  import AccountMenu from './AccountMenu.svelte'
  import { t } from '../i18n'

  let currentHash: string = $state('')
  let showSignIn = $state(false)
  let showAccountMenu = $state(false)
  let pendingDeleteId: number | null = $state(null)
  let deleteTimer: ReturnType<typeof setTimeout> | null = $state(null)

  onMount(() => {
    currentHash = window.location.hash || '#/'
    const onHashChange = () => {
      currentHash = window.location.hash || '#/'
    }
    window.addEventListener('hashchange', onHashChange)
    void account.checkStatus()
    return () => {
      window.removeEventListener('hashchange', onHashChange)
      if (deleteTimer) clearTimeout(deleteTimer)
    }
  })

  async function startNew(): Promise<void> {
    await conversation.createNew(t('sidebar.new_conversation'))
  }

  async function openExisting(id: number): Promise<void> {
    await conversation.open(id)
  }

  function deleteCurrent(): void {
    const id = conversation.currentID
    if (id === null) return
    // Clear any pending timer from a prior delete click so we don't
    // leak stale setTimeouts on rapid double-clicks (each click used
    // to start a fresh timer without clearing the previous).
    if (deleteTimer) clearTimeout(deleteTimer)
    pendingDeleteId = id
    notifications.push({
      kind: 'warn',
      title: t('sidebar.delete_current'),
      message: t('sidebar.delete_confirm'),
      sticky: false
    })
    deleteTimer = setTimeout(async () => {
      // Use the pending id (NOT conversation.currentID): if the user
      // opened a different conversation during the undo window, the
      // store's currentID now points to that other conversation, and
      // calling conversation.deleteCurrent() would delete the wrong
      // one. deleteById targets the conversation the user actually
      // clicked on. See internal audit fix: undo-delete wrong-target.
      if (pendingDeleteId !== null) {
        const target = pendingDeleteId
        pendingDeleteId = null
        await conversation.deleteById(target)
      }
    }, 5000)
  }

  function undoDelete(): void {
    if (deleteTimer) clearTimeout(deleteTimer)
    pendingDeleteId = null
    notifications.push({
      kind: 'info',
      title: t('sidebar.delete_current'),
      message: t('sidebar.delete_cancelled')
    })
  }
</script>

<aside class="sidebar">
  <!-- Icon Rail -->
  <nav class="icon-rail" aria-label="Primary navigation">
    <div class="rail-glow"></div>
    <div class="rail-top">
      <a
        href="#/"
        class="rail-icon"
        class:active={currentHash === '#/' || currentHash === '#' || currentHash === ''}
        title={t('sidebar.nav.chat')}
        aria-label={t('sidebar.nav.chat')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M4 4h12a2 2 0 012 2v7a2 2 0 01-2 2H7l-4 3V6a2 2 0 012-2z"/></svg>
      </a>
      <a
        href="#/audit"
        class="rail-icon"
        class:active={currentHash === '#/audit'}
        title={t('sidebar.nav.audit')}
        aria-label={t('sidebar.nav.audit')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M10 2l7 3v5c0 4-3 6.5-7 8-4-1.5-7-4-7-8V5l7-3z"/></svg>
      </a>
      <a
        href="#/replay"
        class="rail-icon"
        class:active={currentHash === '#/replay'}
        title={t('sidebar.nav.replay')}
        aria-label={t('sidebar.nav.replay')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="10" cy="10" r="7"/><path d="M8 7l5 3-5 3V7z"/></svg>
      </a>
      <a
        href="#/hub"
        class="rail-icon"
        class:active={currentHash === '#/hub'}
        title={t('sidebar.nav.skills_hub')}
        aria-label={t('sidebar.nav.skills_hub')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 7l7-4 7 4-7 4-7-4zM3 7v6l7 4M17 7v6l-7 4"/></svg>
      </a>
      <a
        href="#/skills"
        class="rail-icon"
        class:active={currentHash === '#/skills'}
        title={t('sidebar.nav.skills')}
        aria-label={t('sidebar.nav.skills')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M4 4l6 2 6-2-2 12-4 2-4-2L4 4z"/><path d="M10 6v12"/></svg>
      </a>
      <a
        href="#/sync"
        class="rail-icon"
        class:active={currentHash === '#/sync'}
        title={t('sidebar.nav.sync')}
        aria-label={t('sidebar.nav.sync')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M5 10a5 5 0 019-3l1 1m0-3v3h-3M15 10a5 5 0 01-9 3l-1-1m0 3v-3h3"/></svg>
      </a>
      <a
        href="#/channels"
        class="rail-icon"
        class:active={currentHash === '#/channels'}
        title={t('sidebar.nav.channels')}
        aria-label={t('sidebar.nav.channels')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 5h14v9H8l-4 3v-3H3V5z"/></svg>
      </a>
      <a
        href="#/delegation"
        class="rail-icon"
        class:active={currentHash === '#/delegation'}
        title={t('sidebar.nav.delegation')}
        aria-label={t('sidebar.nav.delegation')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="6" cy="6" r="2.5"/><circle cx="14" cy="6" r="2.5"/><circle cx="10" cy="14" r="2.5"/><path d="M6 8.5v2M14 8.5v2M10 5v6.5"/></svg>
      </a>
      <button
        type="button"
        class="rail-icon"
        title={t('sidebar.nav.quick_prompt')}
        aria-label={t('sidebar.nav.quick_prompt')}
        onclick={() => overlay.toggle()}
      >
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M4 14l4-8 4 4 4-6"/><path d="M3 17h14"/></svg>
      </button>
      <div class="rail-spacer"></div>
      <a
        href="#/settings"
        class="rail-icon"
        class:active={currentHash === '#/settings'}
        title={t('sidebar.nav.settings')}
        aria-label={t('sidebar.nav.settings')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="10" cy="10" r="3"/><path d="M10 1v2m0 14v2m-7-9h2m14 0h2m-3.5-5.5-1.4 1.4m-8.2 8.2-1.4 1.4m0-11-1.4 1.4m8.2 8.2 1.4 1.4"/></svg>
      </a>
      <a
        href="#/about"
        class="rail-icon"
        class:active={currentHash === '#/about'}
        title={t('sidebar.nav.about')}
        aria-label={t('sidebar.nav.about')}
      >
        <span class="active-indicator"></span>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="10" cy="10" r="8"/><path d="M10 9v4m0-7h0"/></svg>
      </a>
    </div>
  </nav>

  <!-- Conversation Drawer -->
  <div class="drawer" role="complementary" aria-label="Conversations">
    <div class="drawer-header">
      <button class="new-conv-btn" onclick={startNew}>
        <div class="new-conv-bg"></div>
        <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 4v12m-6-6h12"/></svg>
        <span>{t('sidebar.new_conversation')}</span>
      </button>
      <div class="drawer-label-wrap">
        <span class="drawer-label">{t('sidebar.history')}</span>
        <div class="drawer-label-line"></div>
      </div>
    </div>

    <div class="conversation-list">
      {#if conversation.conversations.length === 0}
        <div class="empty">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1"><path d="M4 5h16a2 2 0 012 2v8a2 2 0 01-2 2H8l-5 4V7a2 2 0 012-2z"/><path d="M9 10h6M9 13h4"/></svg>
          <p>{t('sidebar.empty')}</p>
        </div>
      {/if}
      {#each conversation.conversations as c (c.id)}
        <button
          class="conversation-item"
          class:active={c.id === conversation.currentID}
          onclick={() => openExisting(c.id)}
        >
          <div class="item-content">
            <span class="title">{c.title}</span>
            <span class="meta">{t('sidebar.msg_count', c.message_count)} · {new Date(c.updated_at).toLocaleDateString()}</span>
          </div>
          <div class="active-glow"></div>
        </button>
      {/each}
    </div>

    {#if conversation.currentID}
      <div class="drawer-footer">
        {#if pendingDeleteId !== null}
          <button class="btn-undo" onclick={undoDelete} aria-label="Undo delete conversation">
            <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" class="delete-icon"><path d="M4 10h12M4 10l4-4M4 10l4 4"/></svg>
            <span>{t('sidebar.undo_delete')}</span>
          </button>
        {:else}
          <button class="btn-delete" onclick={deleteCurrent} aria-label={t('sidebar.delete_current')}>
            <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" class="delete-icon"><path d="M4 6h12M8 6V4h4v2m-7 0v10a1 1 0 001 1h6a1 1 0 001-1V6"/></svg>
            <span>{t('sidebar.delete_current')}</span>
          </button>
        {/if}
      </div>
    {/if}

    <!-- Account footer -->
    <div class="account-footer">
      {#if account.isSignedIn}
        <button class="account-chip" onclick={() => (showAccountMenu = true)} title={t('sidebar.account')}>
          {#if account.avatarURL}
            <img class="chip-avatar" src={account.avatarURL} alt="" />
          {:else}
            <span class="chip-avatar fallback">{(account.displayName || '?').charAt(0).toUpperCase()}</span>
          {/if}
          <div class="chip-info">
            <span class="chip-name">{account.displayName || 'User'}</span>
            <span class="chip-email">{account.email || account.displayName}</span>
          </div>
        </button>
      {:else}
        <button class="signin-link" onclick={() => (showSignIn = true)}>
          <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M15 10l-4-4m4 4l-4 4m4-4H5"/></svg>
          {t('sidebar.signin')}
        </button>
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
    position: relative;
    z-index: var(--z-elevated);
  }

  /* ── Icon Rail — the spine of the app ──────────── */
  .icon-rail {
    width: var(--sidebar-rail-width);
    min-width: var(--sidebar-rail-width);
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur-heavy);
    -webkit-backdrop-filter: var(--glass-blur-heavy);
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: var(--space-4) 0;
    z-index: 2;
    position: relative;
    border-right: 1px solid var(--color-border);
  }

  /* Ambient glow at the top of the rail */
  .rail-glow {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 200px;
    background: radial-gradient(ellipse at top, var(--color-accent-faint), transparent 70%);
    pointer-events: none;
    animation: breathe-soft 8s ease-in-out infinite;
  }

  .rail-top {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    position: relative;
    z-index: 1;
    width: 100%;
    height: 100%;
  }

  .rail-spacer {
    flex-grow: 1;
    min-height: var(--space-4);
  }

  /* ── Rail icons — tactile, magnetic ─────────────── */
  .rail-icon {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 42px;
    height: 42px;
    border-radius: var(--radius-lg);
    color: var(--color-text-faint);
    text-decoration: none;
    transition: all var(--transition-spring);
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
    background: var(--glass-bg-hover);
    box-shadow: var(--shadow-sm);
  }

  .rail-icon:hover svg {
    transform: scale(1.15);
  }

  .rail-icon:active {
    transform: scale(0.92);
    transition-duration: var(--transition-instant);
  }

  /* Active indicator — glowing accent bar */
  .active-indicator {
    position: absolute;
    left: -2px;
    top: 50%;
    transform: translateY(-50%) scaleY(0);
    width: 3px;
    height: 22px;
    border-radius: 0 4px 4px 0;
    background: var(--color-accent);
    box-shadow: 0 0 12px var(--color-accent-glow), 0 0 24px var(--color-glow);
    transition: transform var(--transition-spring);
  }

  .rail-icon.active .active-indicator {
    transform: translateY(-50%) scaleY(1);
  }

  .rail-icon.active {
    color: var(--color-text);
    background: linear-gradient(90deg, var(--color-accent-soft) 0%, transparent 100%);
  }

  .rail-icon.active svg {
    filter: drop-shadow(0 0 6px var(--color-glow));
  }

  .rail-icon.active:hover {
    color: var(--color-accent-hover);
    transform: translateX(2px);
  }

  /* ── Conversation Drawer ──────────────────────── */
  .drawer {
    width: var(--sidebar-drawer-width);
    min-width: var(--sidebar-drawer-width);
    background: var(--glass-bg-solid);
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
    gap: var(--space-4);
  }

  /* ── New conversation button — the hero CTA ────── */
  .new-conv-btn {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    width: 100%;
    padding: 12px var(--space-3);
    border-radius: var(--radius-lg);
    background: var(--glass-bg);
    border: 1px solid var(--color-border-strong);
    color: var(--color-text);
    font-size: var(--size-sm);
    font-weight: var(--weight-semibold);
    font-family: var(--font-sans);
    cursor: pointer;
    overflow: hidden;
    transition: all var(--transition-base);
    box-shadow: var(--shadow-sm), var(--shadow-inset);
    letter-spacing: var(--tracking-normal);
  }

  .new-conv-bg {
    position: absolute;
    inset: 0;
    background: var(--color-accent-gradient);
    opacity: 0;
    transition: opacity var(--transition-base);
    z-index: 0;
  }

  .new-conv-btn svg, .new-conv-btn span {
    position: relative;
    z-index: 1;
    transition: transform var(--transition-spring);
  }

  .new-conv-btn svg {
    width: 16px;
    height: 16px;
  }

  .new-conv-btn:hover {
    transform: translateY(-2px);
    border-color: var(--color-accent);
    box-shadow: var(--shadow-md), var(--shadow-glow), var(--shadow-inset);
  }

  .new-conv-btn:hover .new-conv-bg {
    opacity: 0.2;
  }

  .new-conv-btn:hover svg {
    transform: rotate(90deg) scale(1.1);
  }

  .new-conv-btn:active {
    transform: translateY(0) scale(0.98);
    transition-duration: var(--transition-instant);
  }

  .drawer-label-wrap {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 0 var(--space-1);
  }

  .drawer-label {
    font-size: var(--size-2xs);
    font-weight: var(--weight-bold);
    text-transform: uppercase;
    letter-spacing: var(--tracking-widest);
    color: var(--color-text-dim);
  }

  .drawer-label-line {
    flex-grow: 1;
    height: 1px;
    background: linear-gradient(90deg, var(--color-border) 0%, transparent 100%);
  }

  /* ── Conversation List — staggered entrance ────── */
  .conversation-list {
    flex: 1;
    overflow-y: auto;
    padding: 0 var(--space-2) var(--space-2);
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .conversation-item {
    position: relative;
    display: flex;
    align-items: center;
    width: 100%;
    text-align: left;
    padding: 10px 14px;
    border-radius: var(--radius-md);
    background: transparent;
    color: var(--color-text-muted);
    border: 1px solid transparent;
    cursor: pointer;
    overflow: hidden;
    transition: all var(--transition-fast);
    animation: stagger-in var(--transition-slow) var(--ease-out-expo) both;
  }

  .item-content {
    display: flex;
    flex-direction: column;
    gap: 3px;
    position: relative;
    z-index: 1;
    width: 100%;
  }

  .conversation-item:hover {
    background: var(--glass-bg-hover);
    color: var(--color-text);
    transform: translateX(2px);
  }

  .conversation-item.active {
    background: var(--glass-bg-active);
    color: var(--color-text);
    border-color: var(--glass-border);
  }

  .active-glow {
    position: absolute;
    left: 0;
    top: 50%;
    transform: translateY(-50%) scaleY(0);
    width: 3px;
    height: 70%;
    background: var(--color-accent);
    border-radius: 0 4px 4px 0;
    box-shadow: 0 0 12px var(--color-accent-glow);
    opacity: 0;
    transition: all var(--transition-spring);
  }

  .conversation-item.active .active-glow {
    opacity: 1;
    transform: translateY(-50%) scaleY(1);
  }

  .conversation-item:active {
    transform: scale(0.98);
    transition-duration: var(--transition-instant);
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
    color: var(--color-text-dim);
    line-height: 1.3;
    font-family: var(--font-mono);
  }

  /* Empty state — with personality */
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: var(--space-8) var(--space-3);
    color: var(--color-text-dim);
    animation: fade-in-scale var(--transition-slow) var(--ease-out-expo) both;
  }

  .empty svg {
    width: 32px;
    height: 32px;
    opacity: 0.25;
    margin-bottom: var(--space-3);
    animation: breathe-soft 4s ease-in-out infinite;
  }

  .empty p {
    font-size: var(--size-xs);
    line-height: var(--leading-relaxed);
  }

  /* ── Drawer Footer ───────────────────────────── */
  .drawer-footer {
    padding: var(--space-2) var(--space-3);
    border-top: 1px solid var(--color-border);
  }

  .btn-delete {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    width: 100%;
    padding: 10px;
    border-radius: var(--radius-md);
    font-size: var(--size-xs);
    font-weight: var(--weight-medium);
    background: transparent;
    color: var(--color-text-faint);
    border: 1px solid transparent;
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .btn-delete:hover {
    color: var(--color-error);
    background: var(--color-error-soft);
    border-color: rgba(163, 49, 42, 0.25);
  }

  .btn-delete:active {
    transform: scale(0.97);
    transition-duration: var(--transition-instant);
  }

  .btn-undo {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    width: 100%;
    padding: 10px;
    border-radius: var(--radius-md);
    font-size: var(--size-xs);
    font-weight: var(--weight-medium);
    background: var(--color-accent-soft);
    color: var(--color-accent);
    border: 1px solid rgba(11, 61, 46, 0.25);
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .btn-undo:hover {
    background: var(--color-accent);
    color: var(--color-paper);
    border-color: var(--color-accent);
  }

  .btn-undo:active {
    transform: scale(0.97);
    transition-duration: var(--transition-instant);
  }

  .delete-icon {
    width: 14px;
    height: 14px;
    flex-shrink: 0;
  }

  /* ── Account Footer ──────────────────────────── */
  .account-footer {
    padding: var(--space-3);
    border-top: 1px solid var(--color-border);
    background: var(--glass-bg);
  }

  .signin-link {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    width: 100%;
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-md);
    color: var(--color-text);
    font-size: var(--size-xs);
    font-weight: var(--weight-medium);
    padding: 10px var(--space-2);
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .signin-link:hover {
    background: var(--glass-bg-hover);
    border-color: var(--color-accent);
    color: var(--color-accent);
    box-shadow: 0 0 16px var(--color-accent-faint);
  }

  .signin-link svg {
    width: 16px;
    height: 16px;
    transition: transform var(--transition-spring);
  }

  .signin-link:hover svg {
    transform: translateX(2px);
  }

  .account-chip {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    padding: 8px;
    border-radius: var(--radius-lg);
    background: transparent;
    border: 1px solid transparent;
    color: var(--color-text);
    cursor: pointer;
    text-align: left;
    transition: all var(--transition-fast);
  }

  .account-chip:hover {
    background: var(--glass-bg-hover);
    border-color: var(--glass-border);
    transform: translateY(-1px);
  }

  .chip-avatar {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    flex-shrink: 0;
    object-fit: cover;
    box-shadow: var(--shadow-sm);
  }

  .chip-avatar.fallback {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--color-accent-gradient);
    color: var(--color-paper);
    font-weight: var(--weight-bold);
    font-size: var(--size-sm);
    box-shadow: var(--shadow-sm), 0 0 16px var(--color-glow);
  }

  .chip-info {
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .chip-name {
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .chip-email {
    font-size: var(--size-2xs);
    color: var(--color-text-faint);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-family: var(--font-mono);
  }
</style>
