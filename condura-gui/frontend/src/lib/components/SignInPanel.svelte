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
  import { Dialog } from './ui'
  import Card from './ui/Card.svelte'
  import Button from './ui/Button.svelte'
  import Input from './ui/Input.svelte'
  import { t } from '../i18n'

  interface Props {
    onClose?: () => void
  }
  let { onClose }: Props = $props()

  // Desktop deep-link the daemon registers for the OAuth redirect.
  const OAUTH_REDIRECT = 'condura://auth/callback'
  // Where the magic link lands the user back.
  const MAGIC_REDIRECT = 'https://condura.app/auth/verify'

  let open = $state(true)
  let email = $state('')
  let magicSent = $state(false)
  let mode = $state<'signin' | 'signup' | 'magic'>('signin')

  function close(): void {
    open = false
    onClose?.()
  }

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

<Dialog
  bind:open
  title={t('account.signin.title')}
  size="sm"
  onclose={close}
>
  {#snippet children()}
    <div class="signin-body">
      <p class="lead">{t('account.signin.lead')}</p>

      <div class="mode-toggle" role="tablist">
        <button
          type="button"
          class="mode-tab"
          class:active={mode === 'signin'}
          role="tab"
          aria-selected={mode === 'signin'}
          onclick={() => (mode = 'signin')}
        >
          {t('account.signin.tab_signin')}
        </button>
        <button
          type="button"
          class="mode-tab"
          class:active={mode === 'signup'}
          role="tab"
          aria-selected={mode === 'signup'}
          onclick={() => (mode = 'signup')}
        >
          {t('account.signin.tab_signup')}
        </button>
        <button
          type="button"
          class="mode-tab"
          class:active={mode === 'magic'}
          role="tab"
          aria-selected={mode === 'magic'}
          onclick={() => (mode = 'magic')}
        >
          {t('account.signin.tab_magic')}
        </button>
      </div>

      {#if hasProviders}
        <Card elevation="glass" padding="sm">
          <div class="providers">
            {#each account.configuredProviders as p (p)}
              <button
                class="provider-row press"
                onclick={() => signInWith(p)}
                disabled={account.loading}
              >
                <span class="p-icon" aria-hidden="true">
                  {providerLabel(p).charAt(0)}
                </span>
                <span class="p-label">
                  {t('account.signin.continue_with', providerLabel(p))}
                </span>
              </button>
            {/each}
          </div>
        </Card>

        <div class="or-divider"><span>{t('account.signin.or')}</span></div>
      {:else}
        <Card elevation="glass" padding="md">
          <p class="setup-hint">{t('account.signin.setup_hint')}</p>
        </Card>
      {/if}

      <div class="magic">
        <Input
          id="signin-email"
          fullWidth
          type="email"
          label={t('account.signin.magic_label')}
          bind:value={email}
          placeholder="you@example.com"
          autocomplete="email"
        />
        <Button
          variant="primary"
          fullWidth
          disabled={!emailValid || account.loading}
          loading={account.loading}
          onclick={withEmail}
        >
          {account.loading
            ? t('account.signin.sending')
            : t('account.signin.send_link')}
        </Button>
      </div>

      {#if magicSent}
        <p class="ok">{t('account.signin.check_inbox', email)}</p>
      {/if}
      {#if account.error}
        <p class="err">{account.error}</p>
      {/if}

      <p class="fineprint">{t('account.signin.fineprint')}</p>

      <div class="toggle-row">
        <span class="toggle-q">
          {mode === 'signup'
            ? t('account.signin.have_account_q')
            : t('account.signin.no_account_q')}
        </span>
        <button
          type="button"
          class="toggle-link"
          onclick={() => (mode = mode === 'signup' ? 'signin' : 'signup')}
        >
          {mode === 'signup'
            ? t('account.signin.signin_link')
            : t('account.signin.signup_link')}
        </button>
      </div>
    </div>
  {/snippet}
</Dialog>

<style>
  .signin-body {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .lead {
    color: var(--text-muted);
    font-size: var(--size-sm);
    line-height: var(--leading-relaxed);
    margin: 0;
  }

  .mode-toggle {
    display: flex;
    gap: var(--space-1);
    padding: var(--space-1);
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
  }
  .mode-tab {
    flex: 1;
    appearance: none;
    background: transparent;
    border: none;
    padding: var(--space-2) var(--space-3);
    color: var(--text-muted);
    font-family: var(--font-sans);
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    border-radius: var(--radius-sm);
    cursor: pointer;
    transition:
      background-color var(--transition-fast) ease,
      color var(--transition-fast) ease;
  }
  .mode-tab:hover {
    color: var(--text);
  }
  .mode-tab.active {
    background: var(--surface-3);
    color: var(--text);
    box-shadow: var(--shadow-xs);
  }

  .providers {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .provider-row {
    appearance: none;
    background: var(--surface-2);
    border: 1px solid var(--border-strong);
    color: var(--text);
    font-family: var(--font-sans);
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: var(--space-3);
    transition:
      background-color var(--transition-fast) ease,
      border-color var(--transition-fast) ease,
      transform var(--transition-fast) var(--ease-spring);
  }
  .provider-row:hover:not(:disabled) {
    background: var(--surface-3);
    border-color: var(--border-focus);
    transform: translateY(-1px);
  }
  .provider-row:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .p-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    font-weight: var(--weight-bold);
    background: var(--surface-3);
    border: 1px solid var(--border-strong);
    font-size: var(--size-sm);
    color: var(--text);
  }
  .p-label {
    flex: 1;
    text-align: left;
  }

  .setup-hint {
    color: var(--text-muted);
    font-size: var(--size-sm);
    line-height: var(--leading-relaxed);
    margin: 0;
  }

  .or-divider {
    display: flex;
    align-items: center;
    text-align: center;
    color: var(--text-faint);
    font-size: var(--size-xs);
  }
  .or-divider::before,
  .or-divider::after {
    content: '';
    flex: 1;
    height: 1px;
    background: var(--border);
  }
  .or-divider span {
    padding: 0 var(--space-3);
  }

  .magic {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .ok {
    color: var(--success);
    font-size: var(--size-sm);
    margin: 0;
  }
  .err {
    color: var(--error);
    font-size: var(--size-sm);
    margin: 0;
  }

  .fineprint {
    color: var(--text-faint);
    font-size: var(--size-xs);
    line-height: var(--leading-relaxed);
    margin: 0;
  }

  .toggle-row {
    display: flex;
    gap: var(--space-2);
    align-items: baseline;
    justify-content: center;
    padding-top: var(--space-2);
    border-top: 1px solid var(--border);
  }
  .toggle-q {
    color: var(--text-muted);
    font-size: var(--size-sm);
  }
  .toggle-link {
    appearance: none;
    background: transparent;
    border: none;
    color: var(--accent);
    font-family: var(--font-sans);
    font-size: var(--size-sm);
    font-weight: var(--weight-medium);
    cursor: pointer;
    padding: 0;
  }
  .toggle-link:hover {
    color: var(--accent-hover);
    text-decoration: underline;
  }
</style>