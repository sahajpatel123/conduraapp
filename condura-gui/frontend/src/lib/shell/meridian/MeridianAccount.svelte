<script lang="ts">
  /**
   * Account — optional passport. Local work needs none.
   * Signature: local vs cloud atlas doors + OAuth passport strip.
   */
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { account } from '../../stores/account.svelte'

  let signing = $state<string | null>(null)

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
        ? 'Daemon offline — OAuth still opens in the browser'
        : 'No account required for Ask, Audit, Halt, or consent'
  )

  const PROVIDERS = [
    {
      id: 'google',
      label: 'Google',
      mark: 'G',
      run: () => account.signInWithGoogle(window.location.origin + '/'),
    },
    {
      id: 'github',
      label: 'GitHub',
      mark: 'GH',
      run: () => account.signInWithGitHub(window.location.origin + '/'),
    },
    {
      id: 'apple',
      label: 'Apple',
      mark: '',
      run: () => account.signInWithApple(window.location.origin + '/'),
    },
  ] as const

  function go(hash: string): void {
    window.location.hash = hash
  }

  function openDonate(): void {
    window.open('https://condura.app/donate', '_blank', 'noopener,noreferrer')
  }

  async function signIn(id: string, run: () => Promise<unknown> | unknown): Promise<void> {
    if (signing) return
    signing = id
    try {
      await run()
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
            {#each PROVIDERS as p (p.id)}
              <button
                type="button"
                class="provider"
                data-id={p.id}
                disabled={signing === p.id}
                onclick={() => void signIn(p.id, p.run)}
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
          <p class="note">OAuth · browser · optional</p>
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
    box-shadow: 0 0 0 3px color-mix(in oklab, var(--md-live) 16%, transparent);
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
    padding: 7px 11px;
    border-radius: 999px;
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
    font-weight: 700;
    color: var(--md-ink-soft);
  }
  .plate {
    border-radius: 22px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 88%, transparent);
    padding: 24px;
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
    width: 48px;
    height: 48px;
    border-radius: 16px;
    display: grid;
    place-items: center;
    font-family: var(--md-font-display);
    font-size: 20px;
    font-weight: 700;
    color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 12%, var(--md-stage));
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 22%, var(--md-line));
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
    border: 1px solid var(--md-line-strong);
    background: var(--md-stage);
    border-radius: 18px;
    padding: 16px 16px 18px;
    cursor: pointer;
    display: grid;
    gap: 6px;
    color: inherit;
    transition:
      border-color 180ms var(--md-ease),
      transform 180ms var(--md-spring),
      box-shadow 180ms var(--md-ease);
  }
  .door strong {
    font-family: var(--md-font-display);
    font-size: 17px;
    letter-spacing: -0.03em;
  }
  .door > span:not(.door-k):not(.cta) {
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
  }
  .door:hover {
    border-color: var(--md-cobalt);
    transform: translateY(-2px);
    box-shadow: var(--md-shadow);
  }
  .door:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
    border-color: var(--md-cobalt);
  }
  .door.primary {
    background: color-mix(in oklab, var(--md-cobalt) 8%, var(--md-surface));
    border-color: color-mix(in oklab, var(--md-cobalt) 28%, var(--md-line));
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
    transform: none;
    box-shadow: none;
    border-color: var(--md-line-strong);
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
    padding: 9px 12px;
    border-radius: 999px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    font-size: 13px;
    font-weight: 700;
    cursor: pointer;
    color: var(--md-ink);
    transition:
      border-color 180ms var(--md-ease),
      transform 160ms var(--md-spring),
      box-shadow 180ms var(--md-ease);
  }
  .provider:hover:not(:disabled) {
    border-color: var(--md-cobalt);
    transform: translateY(-1px);
    box-shadow: var(--md-shadow);
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

  @media (max-width: 720px) {
    .atlas,
    .doors {
      grid-template-columns: 1fr;
    }
  }
</style>
