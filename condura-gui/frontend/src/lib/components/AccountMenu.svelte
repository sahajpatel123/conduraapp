<script lang="ts">
  // AccountMenu (Phase 14B). A dropdown shown when a signed-in user
  // clicks their avatar in the sidebar footer. Shows the email +
  // provider, a "Manage account" link, and a sign-out button with
  // a two-step confirmation.
  import { account } from '../stores/account.svelte'
  import Avatar from './ui/Avatar.svelte'
  import Button from './ui/Button.svelte'
  import { t } from '../i18n'

  interface Props {
    onClose?: () => void
    onManage?: () => void
  }
  let { onClose, onManage }: Props = $props()

  let confirmingSignOut = $state(false)

  $effect(() => {
    confirmingSignOut // reset focus index when dialog state toggles
    focusedIndex = 0
  })

  function providerLabel(p: string): string {
    switch (p) {
      case 'google': return 'Google'
      case 'github': return 'GitHub'
      case 'apple': return 'Apple'
      case 'magic': return 'Email'
      default: return p || 'Account'
    }
  }

  async function doSignOut(): Promise<void> {
    await account.signOut()
    onClose?.()
  }

  function handleBackdropClick(): void {
    onClose?.()
  }

  let menuEl = $state<HTMLDivElement | null>(null)
  let focusedIndex = $state(0)

  function navigateMenu(e: KeyboardEvent): void {
    if (!menuEl) return
    const items = menuEl.querySelectorAll<HTMLElement>('[role="menuitem"]')
    if (items.length === 0) return

    if (e.key === 'ArrowDown') {
      e.preventDefault()
      focusedIndex = (focusedIndex + 1) % items.length
      items[focusedIndex]?.focus()
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      focusedIndex = (focusedIndex - 1 + items.length) % items.length
      items[focusedIndex]?.focus()
    } else if (e.key === 'Home') {
      e.preventDefault()
      focusedIndex = 0
      items[0]?.focus()
    } else if (e.key === 'End') {
      e.preventDefault()
      focusedIndex = items.length - 1
      items[items.length - 1]?.focus()
    } else if (e.key === 'Escape') {
      onClose?.()
    }
  }

  function handleBackdropKey(e: KeyboardEvent): void {
    if (e.key === 'Escape') onClose?.()
  }
</script>

<div
  class="menu-backdrop anim-fade"
  role="presentation"
  onclick={handleBackdropClick}
  onkeydown={handleBackdropKey}
>
  <div
    bind:this={menuEl}
    class="account-menu anim-pop"
    role="menu"
    tabindex="-1"
    aria-label={t('account.menu.aria_label')}
    onclick={(e) => e.stopPropagation()}
    onkeydown={navigateMenu}
  >
    <div class="who">
      <Avatar
        name={account.displayName || account.email || 'Account'}
        src={account.avatarURL}
        size="md"
      />
      <div class="who-text">
        <span class="name">{account.displayName || account.email}</span>
        <span class="email">{account.email}</span>
        <span class="provider">
          {t('account.menu.via', providerLabel(account.provider))}{account.tier ? ` · ${account.tier}` : ''}
        </span>
      </div>
    </div>

    <div class="sep"></div>

    {#if confirmingSignOut}
      <p class="confirm-q">{t('account.menu.signout_confirm')}</p>
      <div class="confirm-actions">
        <Button variant="ghost" size="sm" fullWidth onclick={() => (confirmingSignOut = false)}>
          {t('account.menu.cancel')}
        </Button>
        <Button
          variant="danger"
          size="sm"
          fullWidth
          loading={account.loading}
          onclick={doSignOut}
        >
          {account.loading ? t('account.menu.signing_out') : t('account.menu.signout')}
        </Button>
      </div>
    {:else}
      <button class="item press" role="menuitem" onclick={onManage}>
        {t('account.menu.manage')}
      </button>
      <button class="item press" role="menuitem" onclick={() => (confirmingSignOut = true)}>
        {t('account.menu.signout')}
      </button>
    {/if}

    {#if account.error}
      <p class="err">{account.error}</p>
    {/if}
  </div>
</div>

<style>
  .menu-backdrop {
    position: fixed;
    inset: 0;
    z-index: var(--z-elevated);
    background: transparent;
  }
  .account-menu {
    position: absolute;
    bottom: 64px;
    left: 12px;
    width: 248px;
    padding: var(--space-3);
    background: var(--glass-bg-solid);
    backdrop-filter: var(--glass-blur-heavy);
    -webkit-backdrop-filter: var(--glass-blur-heavy);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
  }

  .who {
    display: flex;
    gap: var(--space-3);
    align-items: center;
  }
  .who-text {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .name {
    font-weight: var(--weight-semibold);
    font-size: var(--size-sm);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--text);
  }
  .email {
    color: var(--text-muted);
    font-size: var(--size-xs);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .provider {
    color: var(--text-faint);
    font-size: var(--size-xs);
    margin-top: 2px;
  }

  .sep {
    height: 1px;
    background: var(--border);
    margin: var(--space-3) 0;
  }

  .item {
    width: 100%;
    text-align: left;
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-md);
    background: transparent;
    border: none;
    color: var(--text);
    font-size: var(--size-sm);
    font-family: var(--font-sans);
    cursor: pointer;
    transition: background-color var(--transition-fast) ease, color var(--transition-fast) ease;
  }
  .item:hover {
    background: var(--surface-3);
  }
  .item:focus-visible {
    outline: 2px solid var(--border-focus);
    outline-offset: 2px;
  }

  .confirm-q {
    font-size: var(--size-sm);
    color: var(--text-muted);
    margin: 0 0 var(--space-2);
    line-height: var(--leading-relaxed);
  }
  .confirm-actions {
    display: flex;
    gap: var(--space-2);
  }
  .err {
    color: var(--error);
    font-size: var(--size-xs);
    margin-top: var(--space-2);
  }
</style>