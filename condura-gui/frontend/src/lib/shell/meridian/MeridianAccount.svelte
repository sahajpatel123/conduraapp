<script lang="ts">
  /**
   * Account — optional passport. Local work needs none.
   * Signature: local vs cloud atlas doors + OAuth passport strip.
   *
   * OAuth must use the desktop deep-link redirect (condura://auth/callback)
   * and open the returned URL — browser origin is not registered with the
   * providers and discarding the URL left buttons stuck on "Opening…".
   */
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { account } from '../../stores/account.svelte'
  import type { AccountProvider, OAuthURLResult } from '../../ipc/types'

  /** Must match processOAuthCallback in condura-app/cmd/condura-gui/main.go */
  const OAUTH_REDIRECT = 'condura://auth/callback'

  let signing = $state<string | null>(null)
  let signInNote = $state('')

  onMount(() => {
    void account.checkStatus()
  })

  const signedIn = $derived(!!account.status?.signed_in)
  const offlineError = $derived(
    !!account.error && /IPC client not started|not connected|Failed to fetch|daemon/i.test(account.error)
  )
  const liveNote = $derived(
    signedIn
      ? `Signed in as ${account.status?.display_name || account.status?.email || 'you'}`
      : offlineError
        ? 'Daemon offline — reconnect to sign in'
        : 'No account required for Ask, Audit, Halt, or consent'
  )

  type ProviderBtn = {
    id: AccountProvider
    label: string
    mark: string
  }

  const ALL_PROVIDERS: ProviderBtn[] = [
    { id: 'google', label: 'Google', mark: 'G' },
    { id: 'github', label: 'GitHub', mark: 'GH' },
    { id: 'apple', label: 'Apple', mark: '' },
  ]

  /** Only show providers the daemon reports as configured; fall back to all if unknown. */
  const providers = $derived.by(() => {
    const configured = account.configuredProviders?.filter((p) => p !== 'magic') ?? []
    if (configured.length === 0) return ALL_PROVIDERS
    return ALL_PROVIDERS.filter((p) => configured.includes(p.id))
  })

  function go(hash: string): void {
    window.location.hash = hash
  }

  function openDonate(): void {
    window.open('https://condura.app/donate', '_blank', 'noopener,noreferrer')
  }

  function openExternal(url: string): void {
    const w = window as unknown as { runtime?: { BrowserOpenURL?: (u: string) => void } }
    if (w.runtime?.BrowserOpenURL) {
      try {
        w.runtime.BrowserOpenURL(url)
        return
      } catch {
        /* fall through */
      }
    }
    window.open(url, '_blank', 'noopener,noreferrer')
  }

  async function signIn(id: AccountProvider): Promise<void> {
    if (signing) return
    signing = id
    signInNote = ''
    try {
      let res: OAuthURLResult | null = null
      if (id === 'google') res = await account.signInWithGoogle(OAUTH_REDIRECT)
      else if (id === 'github') res = await account.signInWithGitHub(OAUTH_REDIRECT)
      else if (id === 'apple') res = await account.signInWithApple(OAUTH_REDIRECT)
      if (res?.url) {
        openExternal(res.url)
        signInNote = 'Browser opened — finish sign-in there, then return here.'
      } else if (account.error) {
        signInNote = account.error
      } else {
        signInNote = 'No sign-in URL returned. Check daemon account config.'
      }
    } catch (e) {
      signInNote = String(e)
    } finally {
      signing = null
    }
  }
</script>

<MeridianPage
  kicker="Passport · optional"
  title="Account"
  lead="Local agent work never needs an account. Sign in only when you want Hub publishing, donations, or support."
>
  <div class="desk md-stagger">
    <p class="contract" class:hot={signedIn} class:off={offlineError && !signedIn}>
      <span class="live-dot" aria-hidden="true"></span>
      {liveNote}.
    </p>

    {#if account.loading && !account.status}
      <div class="md-empty">Checking session…</div>
    {:else if account.error && !offlineError}
      <div class="md-empty">{account.error}</div>
    {:else if signedIn}
      <section class="session plate">
        <p class="cite">signed in · cloud doors open</p>
        <div class="who">
          <div class="avatar" aria-hidden="true">
            {(account.status?.display_name || account.status?.email || '?').slice(0, 1).toUpperCase()}
          </div>
          <div>
            <h2>{account.status?.display_name || account.status?.email || 'You'}</h2>
            <p class="meta">{account.status?.email}</p>
          </div>
        </div>
        <p class="unlock">
          This identity unlocks optional cloud doors. Local Ask, Audit, and Halt still work the same.
        </p>
        <div class="doors">
          <button type="button" class="door" onclick={() => go('#/hub')}>
            <span class="door-k">shelf</span>
            <strong>Open Hub</strong>
            <span>Browse and install skills locally</span>
          </button>
          <button type="button" class="door" onclick={openDonate}>
            <span class="door-k">freedom</span>
            <strong>Donate</strong>
            <span>Keep Condura unbound</span>
          </button>
          <button type="button" class="door" onclick={() => go('#/settings')}>
            <span class="door-k">desk</span>
            <strong>Settings</strong>
            <span>Lighting, spend, autonomy</span>
          </button>
          <button type="button" class="door" onclick={() => go('#/')}>
            <span class="door-k">ask</span>
            <strong>Back to Ask</strong>
            <span>Your machine stays the source of truth</span>
          </button>
        </div>
        <button type="button" class="md-btn md-btn-danger signout" onclick={() => void account.signOut()}>
          Sign out
        </button>
      </section>
    {:else}
      <ol class="pipe" aria-label="What needs an account">
        <li><span class="n">01</span><span class="t">Local · no account</span></li>
        <li><span class="n">02</span><span class="t">Hub publish · optional</span></li>
        <li><span class="n">03</span><span class="t">Donate · optional</span></li>
      </ol>

      <div class="atlas">
        <button type="button" class="door primary" onclick={() => go('#/')}>
          <span class="door-k">local</span>
          <strong>Stay local</strong>
          <span>Chat, audit, halt, and consent work with no account. Your machine stays the source of truth.</span>
          <span class="cta">Go to Ask →</span>
        </button>
        <div class="door secondary">
          <span class="door-k">cloud · optional</span>
          <strong>Carry a passport</strong>
          <span>OAuth opens in your browser. Condura never sees your password.</span>
          <div class="passport" role="group" aria-label="Sign in providers">
            {#each providers as p (p.id)}
              <button
                type="button"
                class="provider"
                data-id={p.id}
                disabled={signing === p.id}
                onclick={() => void signIn(p.id)}
              >
                <span class="mark" aria-hidden="true">
                  {#if p.id === 'apple'}
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor"
                      ><path
                        d="M16.4 12.3c0-2.1 1.7-3.1 1.8-3.2-1-1.4-2.5-1.6-3-1.6-1.3-.1-2.5.8-3.1.8-.7 0-1.7-.7-2.8-.7-1.4 0-2.8.9-3.5 2.2-1.5 2.6-.4 6.5 1.1 8.6.7 1 1.6 2.2 2.7 2.1 1.1 0 1.5-.7 2.8-.7s1.7.7 2.8.7c1.2 0 1.9-1 2.6-2 .8-1.2 1.1-2.3 1.1-2.4-.1 0-2.2-.8-2.2-3.3zM14.5 5.8c.6-.7 1-1.7.9-2.7-.9 0-1.9.6-2.5 1.3-.6.7-1.1 1.7-.9 2.7 1 .1 1.9-.5 2.5-1.3z"
                      /></svg
                    >
                  {:else}
                    {p.mark}
                  {/if}
                </span>
                {signing === p.id ? 'Opening…' : p.label}
              </button>
            {/each}
          </div>
          {#if signInNote}
            <p class="signin-note" class:bad={!!account.error && !signInNote.startsWith('Browser')}>
              {signInNote}
            </p>
          {/if}
          <p class="note">OAuth · browser · optional · deep-link return</p>
        </div>
      </div>
    {/if}
  </div>
</MeridianPage>

<style>
  .desk {
    display: grid;
    gap: 16px;
  }
  .contract {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    margin: 0;
    padding: 12px 14px;
    border-radius: 14px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 80%, transparent);
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
  }
  .contract.hot {
    border-color: color-mix(in oklab, var(--md-live) 28%, transparent);
    background: color-mix(in oklab, var(--md-live) 6%, var(--md-surface));
  }
  .contract.off {
    border-color: color-mix(in oklab, var(--md-halt) 22%, var(--md-line));
  }
  .live-dot {
    width: 8px;
    height: 8px;
    margin-top: 5px;
    flex: none;
    border-radius: 50%;
    background: var(--md-ink-faint);
  }
  .contract.hot .live-dot {
    background: var(--md-live);
    box-shadow: none;
  }
  .pipe {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin: 0;
    padding: 0;
    list-style: none;
  }
  .pipe li {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 6px 10px;
    border-radius: 7px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 70%, transparent);
  }
  .pipe .n {
    font-family: var(--md-font-mono);
    font-size: 10px;
    color: var(--md-cobalt);
  }
  .pipe .t {
    font-size: 12px;
    font-weight: 650;
    color: var(--md-ink-soft);
  }
  .plate {
    border-radius: 12px;
    border: 1px solid var(--md-line);
    background: var(--md-surface);
    padding: 20px;
    max-width: 640px;
  }
  .cite,
  .door-k {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 10px;
    display: block;
  }
  .who {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-bottom: 14px;
  }
  .avatar {
    width: 44px;
    height: 44px;
    border-radius: 10px;
    display: grid;
    place-items: center;
    font-family: var(--md-font-display);
    font-size: 18px;
    font-weight: 700;
    color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 10%, var(--md-stage));
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 18%, var(--md-line));
  }
  h2 {
    font-family: var(--md-font-display);
    font-size: 26px;
    letter-spacing: -0.04em;
    margin: 0 0 4px;
  }
  .meta {
    margin: 0;
    color: var(--md-ink-mute);
    font-size: 14px;
  }
  .unlock {
    margin: 0 0 18px;
    font-size: 14px;
    line-height: 1.5;
    color: var(--md-ink-mute);
    max-width: 48ch;
  }
  .doors {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
    margin-bottom: 18px;
  }
  .door {
    appearance: none;
    text-align: left;
    border: 1px solid var(--md-line);
    background: var(--md-stage);
    border-radius: 10px;
    padding: 14px 14px 16px;
    cursor: pointer;
    display: grid;
    gap: 6px;
    color: inherit;
    transition: border-color 140ms var(--md-ease), background 140ms var(--md-ease);
  }
  .door strong {
    font-family: var(--md-font-display);
    font-size: 16px;
    letter-spacing: -0.03em;
  }
  .door > span:not(.door-k):not(.cta) {
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
  }
  .door:hover {
    border-color: var(--md-line-strong);
    background: color-mix(in oklab, var(--md-surface) 70%, var(--md-stage));
  }
  .door:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
    border-color: var(--md-cobalt);
  }
  .door.primary {
    background: color-mix(in oklab, var(--md-cobalt) 6%, var(--md-surface));
    border-color: color-mix(in oklab, var(--md-cobalt) 22%, var(--md-line));
  }
  .cta {
    margin-top: 8px;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.06em;
    color: var(--md-cobalt);
    font-weight: 600;
  }
  .signout {
    width: fit-content;
  }
  .atlas {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 14px;
    max-width: 820px;
  }
  .door.secondary {
    cursor: default;
  }
  .door.secondary:hover {
    border-color: var(--md-line);
  }
  .passport {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 10px;
  }
  .provider {
    appearance: none;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 11px;
    border-radius: 8px;
    border: 1px solid var(--md-line);
    background: var(--md-surface);
    font-size: 13px;
    font-weight: 650;
    cursor: pointer;
    color: var(--md-ink);
    transition: border-color 140ms var(--md-ease), background 140ms var(--md-ease);
  }
  .provider:hover:not(:disabled) {
    border-color: var(--md-line-strong);
    background: color-mix(in oklab, var(--md-stage) 50%, var(--md-surface));
  }
  .provider:disabled {
    opacity: 0.6;
    cursor: wait;
  }
  .provider:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .mark {
    width: 22px;
    height: 22px;
    border-radius: 7px;
    display: grid;
    place-items: center;
    font-family: var(--md-font-display);
    font-size: 11px;
    font-weight: 700;
    color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 12%, var(--md-stage));
  }
  .provider[data-id='github'] .mark,
  .provider[data-id='apple'] .mark {
    color: var(--md-ink);
    background: var(--md-stage);
  }
  .note {
    margin: 12px 0 0;
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
  }
  .signin-note {
    margin: 12px 0 0;
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-live);
  }
  .signin-note.bad {
    color: var(--md-halt);
  }

  @media (max-width: 720px) {
    .atlas,
    .doors {
      grid-template-columns: 1fr;
    }
  }
</style>
