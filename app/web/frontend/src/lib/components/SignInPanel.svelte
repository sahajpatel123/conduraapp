<script lang="ts">
  // SignInPanel (Phase 14B). Shown when a signed-out user clicks
  // "Sign in" in the sidebar. Offers OAuth (Google / GitHub / Apple)
  // and email magic-link sign-in. The account store orchestrates the
  // flows; this component is presentational + wiring.
  //
  // Provider buttons are rendered dynamically based on
  // account.configuredProviders — the daemon returns only those
  // providers that have a ClientID configured. When the list is
  // empty, we show a setup hint instead of dead buttons.
  import { account } from '../stores/account.svelte'
  import type { AccountProvider } from '../ipc/types'
  import { t } from '../i18n'

  interface Props {
    onClose?: () => void
  }
  let { onClose }: Props = $props()

  // Desktop deep-link the daemon registers for the OAuth redirect.
  const OAUTH_REDIRECT = 'condura://auth/callback'
  // Where the magic link lands the user back.
  const MAGIC_REDIRECT = 'https://condura.app/auth/verify'

  let email = $state('')
  let magicSent = $state(false)

  function openExternal(url: string): void {
    const w = window as unknown as { runtime?: { BrowserOpenURL?: (u: string) => void } }
    if (w.runtime?.BrowserOpenURL) {
      w.runtime.BrowserOpenURL(url)
    } else {
      window.open(url, '_blank')
    }
  }

  function providerLabel(p: AccountProvider): string {
    switch (p) {
      case 'google': return 'Google'
      case 'github': return 'GitHub'
      case 'apple': return 'Apple'
      case 'magic': return 'Email'
      default: return p
    }
  }

  function providerClass(p: AccountProvider): string {
    return 'provider provider-' + p
  }

  async function signInWith(p: AccountProvider): Promise<void> {
    magicSent = false
    if (p === 'google') {
      const res = await account.signInWithGoogle(OAUTH_REDIRECT)
      if (res?.url) openExternal(res.url)
    } else if (p === 'github') {
      const res = await account.signInWithGitHub(OAUTH_REDIRECT)
      if (res?.url) openExternal(res.url)
    } else if (p === 'apple') {
      const res = await account.signInWithApple(OAUTH_REDIRECT)
      if (res?.url) openExternal(res.url)
    }
  }

  const emailValid = $derived(/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim()))
  const hasProviders = $derived((account.configuredProviders ?? []).length > 0)

  async function withEmail(): Promise<void> {
    if (!emailValid) return
    magicSent = false
    const res = await account.signInWithEmail(email.trim(), 'en', MAGIC_REDIRECT)
    if (res?.sent) magicSent = true
  }
</script>

<div class="signin-backdrop" role="presentation" onclick={() => onClose?.()}>
  <div
    class="signin-panel"
    role="dialog"
    aria-modal="true"
    aria-label={$t('account.signin.aria_label')}
    tabindex="-1"
    onclick={(e) => e.stopPropagation()}
    onkeydown={(e) => { if (e.key === 'Escape') onClose?.() }}
  >
    <header>
      <h2>{$t('account.signin.title')}</h2>
      <button class="close" aria-label={$t('account.signin.close')} onclick={() => onClose?.()}>&times;</button>
    </header>

    <p class="lead">
      {$t('account.signin.lead')}
    </p>

    {#if hasProviders}
      <div class="providers">
        {#each account.configuredProviders as p (p)}
          <button
            class={providerClass(p)}
            onclick={() => signInWith(p)}
            disabled={account.loading}
          >
            <span class="p-icon" aria-hidden="true">{providerLabel(p).charAt(0)}</span>
            {$t('account.signin.continue_with', providerLabel(p))}
          </button>
        {/each}
      </div>
      <div class="divider"><span>{$t('account.signin.or')}</span></div>
    {:else}
      <p class="setup-hint">
        {$t('account.signin.setup_hint')}
      </p>
    {/if}

    <div class="magic">
      <label for="signin-email">{$t('account.signin.magic_label')}</label>
      <div class="magic-row">
        <input
          id="signin-email"
          type="email"
          bind:value={email}
          placeholder="you@example.com"
          autocomplete="email"
          onkeydown={(e) => { if (e.key === 'Enter') withEmail() }}
        />
        <button class="send" onclick={withEmail} disabled={!emailValid || account.loading}>
          {account.loading ? $t('account.signin.sending') : $t('account.signin.send_link')}
        </button>
      </div>
    </div>

    {#if magicSent}
      <p class="ok">{$t('account.signin.check_inbox', email)}</p>
    {/if}
    {#if account.error}
      <p class="err">{account.error}</p>
    {/if}

    <p class="fineprint">
      {$t('account.signin.fineprint')}
    </p>
  </div>
</div>

<style>
  .signin-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
    padding: var(--space-4);
  }
  .signin-panel {
    width: 100%;
    max-width: 420px;
    background: var(--color-bg-elevated, var(--color-bg));
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-xl);
    padding: var(--space-5);
    box-shadow: var(--shadow-lg, 0 20px 60px rgba(0, 0, 0, 0.4));
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-3);
  }
  h2 {
    font-size: var(--size-xl);
    font-weight: 600;
  }
  .close {
    background: none;
    border: none;
    color: var(--color-text-faint);
    font-size: 24px;
    cursor: pointer;
    line-height: 1;
  }
  .close:hover { color: var(--color-text); }
  .lead {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
    line-height: 1.5;
    margin-bottom: var(--space-4);
  }
  .providers {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .provider {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    padding: 11px 16px;
    border-radius: var(--radius-md);
    border: 1px solid var(--glass-border);
    background: var(--glass-bg);
    color: var(--color-text);
    font-size: var(--size-md);
    font-weight: 500;
    cursor: pointer;
    transition: all var(--transition-base);
  }
  .provider:hover:not(:disabled) {
    border-color: var(--color-accent);
    transform: translateY(-1px);
  }
  .provider:disabled { opacity: 0.5; cursor: not-allowed; }
  .p-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    font-weight: 700;
    background: var(--color-bg);
    border: 1px solid var(--glass-border);
    font-size: 13px;
  }
  .setup-hint {
    color: var(--color-text-muted);
    font-size: var(--size-sm);
    line-height: 1.5;
    background: var(--glass-bg);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-md);
    padding: var(--space-3);
    margin-bottom: var(--space-4);
  }
  .setup-hint code {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 12px;
    background: rgba(0, 0, 0, 0.25);
    padding: 1px 5px;
    border-radius: 4px;
  }
  .divider {
    display: flex;
    align-items: center;
    text-align: center;
    color: var(--color-text-faint);
    font-size: var(--size-xs);
    margin: var(--space-4) 0;
  }
  .divider::before, .divider::after {
    content: '';
    flex: 1;
    height: 1px;
    background: var(--glass-border);
  }
  .divider span { padding: 0 var(--space-3); }
  .magic label {
    display: block;
    font-size: var(--size-sm);
    color: var(--color-text-muted);
    margin-bottom: var(--space-2);
  }
  .magic-row { display: flex; gap: var(--space-2); }
  .magic-row input {
    flex: 1;
    padding: 10px 12px;
    border-radius: var(--radius-md);
    border: 1px solid var(--glass-border);
    background: rgba(0, 0, 0, 0.3);
    color: var(--color-text);
    font-size: var(--size-md);
  }
  .magic-row input:focus {
    outline: none;
    border-color: var(--color-accent);
  }
  .send {
    padding: 10px 16px;
    border-radius: var(--radius-md);
    border: none;
    background: var(--color-accent-gradient);
    color: white;
    font-weight: 500;
    cursor: pointer;
    white-space: nowrap;
  }
  .send:disabled { opacity: 0.5; cursor: not-allowed; }
  .ok { color: var(--color-success); font-size: var(--size-sm); margin-top: var(--space-3); }
  .err { color: var(--color-error, #f87171); font-size: var(--size-sm); margin-top: var(--space-3); }
  .fineprint {
    color: var(--color-text-faint);
    font-size: var(--size-xs);
    line-height: 1.5;
    margin-top: var(--space-4);
  }
</style>
