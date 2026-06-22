<script lang="ts">
  // AccountMenu (Phase 14B). A dropdown shown when a signed-in user
  // clicks their avatar in the sidebar footer. Shows the email +
  // provider and offers "Sign out" with a confirmation step.
  import { account } from '../stores/account.svelte'
  import { t } from '../i18n'

  interface Props {
    onClose?: () => void
  }
  let { onClose }: Props = $props()

  let confirmingSignOut = $state(false)

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
</script>

<div class="menu-backdrop" role="presentation" onclick={() => onClose?.()}>
  <div
    class="account-menu"
    role="menu"
    tabindex="-1"
    aria-label={$t('account.menu.aria_label')}
    onclick={(e) => e.stopPropagation()}
    onkeydown={(e) => { if (e.key === 'Escape') onClose?.() }}
  >
    <div class="who">
      {#if account.avatarURL}
        <img class="avatar" src={account.avatarURL} alt="" />
      {:else}
        <div class="avatar fallback">{(account.displayName || '?').charAt(0).toUpperCase()}</div>
      {/if}
      <div class="who-text">
        <span class="name">{account.displayName || account.email}</span>
        <span class="email">{account.email}</span>
        <span class="provider">{$t('account.menu.via', providerLabel(account.provider))}{account.tier ? ` · ${account.tier}` : ''}</span>
      </div>
    </div>

    <div class="sep"></div>

    {#if confirmingSignOut}
      <p class="confirm-q">{$t('account.menu.signout_confirm')}</p>
      <div class="confirm-actions">
        <button class="ghost" onclick={() => (confirmingSignOut = false)}>{$t('account.menu.cancel')}</button>
        <button class="danger" onclick={doSignOut} disabled={account.loading}>
          {account.loading ? $t('account.menu.signing_out') : $t('account.menu.signout')}
        </button>
      </div>
    {:else}
      <button class="item" role="menuitem" onclick={() => (confirmingSignOut = true)}>
        {$t('account.menu.signout')}
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
    z-index: 150;
  }
  .account-menu {
    position: absolute;
    bottom: 64px;
    left: 12px;
    width: 248px;
    background: var(--color-bg-elevated, var(--color-bg));
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-lg);
    padding: var(--space-3);
    box-shadow: var(--shadow-lg, 0 16px 40px rgba(0, 0, 0, 0.4));
  }
  .who {
    display: flex;
    gap: var(--space-3);
    align-items: center;
  }
  .avatar {
    width: 38px;
    height: 38px;
    border-radius: 50%;
    flex-shrink: 0;
    object-fit: cover;
  }
  .avatar.fallback {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--color-accent-gradient);
    color: white;
    font-weight: 600;
  }
  .who-text {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .name {
    font-weight: 600;
    font-size: var(--size-sm);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .email {
    color: var(--color-text-muted);
    font-size: var(--size-xs);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .provider {
    color: var(--color-text-faint);
    font-size: var(--size-xs);
    margin-top: 2px;
  }
  .sep {
    height: 1px;
    background: var(--glass-border);
    margin: var(--space-3) 0;
  }
  .item {
    width: 100%;
    text-align: left;
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-md);
    background: transparent;
    border: none;
    color: var(--color-text);
    font-size: var(--size-sm);
    cursor: pointer;
  }
  .item:hover { background: var(--color-bg-hover); }
  .confirm-q {
    font-size: var(--size-sm);
    color: var(--color-text-muted);
    margin-bottom: var(--space-2);
  }
  .confirm-actions { display: flex; gap: var(--space-2); }
  .ghost, .danger {
    flex: 1;
    padding: var(--space-2);
    border-radius: var(--radius-md);
    font-size: var(--size-sm);
    cursor: pointer;
    border: 1px solid var(--glass-border);
  }
  .ghost { background: transparent; color: var(--color-text-muted); }
  .danger { background: linear-gradient(135deg, #ef4444, #dc2626); color: white; border: none; }
  .danger:disabled { opacity: 0.5; cursor: not-allowed; }
  .err { color: var(--color-error, #f87171); font-size: var(--size-xs); margin-top: var(--space-2); }
</style>
