<script lang="ts">
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { account } from '../../stores/account.svelte'

  onMount(() => {
    void account.checkStatus()
  })

  const signedIn = $derived(!!account.status?.signed_in)
  const offlineError = $derived(
    !!account.error && /IPC client not started|not connected|Failed to fetch|daemon/i.test(account.error)
  )

  const PROVIDERS = [
    { id: 'google', label: 'Continue with Google', mark: 'G', run: () => account.signInWithGoogle(window.location.origin + '/') },
    { id: 'github', label: 'Continue with GitHub', mark: 'GH', run: () => account.signInWithGitHub(window.location.origin + '/') },
    { id: 'apple', label: 'Continue with Apple', mark: '', run: () => account.signInWithApple(window.location.origin + '/') },
  ] as const
</script>

<MeridianPage
  kicker="Identity"
  title="Account"
  lead="Sign in for Hub publishing, donations, and support. Local agent work never requires an account."
>
  {#if account.loading && !account.status}
    <div class="md-empty">Checking session…</div>
  {:else if account.error && !offlineError}
    <div class="md-empty">{account.error}</div>
  {:else if signedIn}
    <div class="md-panel md-panel-static identity">
      <p class="label">Signed in</p>
      <h2>{account.status?.display_name || account.status?.email || 'You'}</h2>
      <p class="meta">{account.status?.email}</p>
      <button type="button" class="md-btn md-btn-danger" onclick={() => void account.signOut()}>Sign out</button>
    </div>
  {:else}
    {#if offlineError}
      <p class="offline">Daemon offline — sign-in opens in the browser once Condura is connected.</p>
    {/if}
    <div class="grid md-stagger">
      {#each PROVIDERS as p (p.id)}
        <button type="button" class="provider" data-id={p.id} onclick={() => void p.run()}>
          <span class="mark" aria-hidden="true">
            {#if p.id === 'apple'}
              <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M16.4 12.3c0-2.1 1.7-3.1 1.8-3.2-1-1.4-2.5-1.6-3-1.6-1.3-.1-2.5.8-3.1.8-.7 0-1.7-.7-2.8-.7-1.4 0-2.8.9-3.5 2.2-1.5 2.6-.4 6.5 1.1 8.6.7 1 1.6 2.2 2.7 2.1 1.1 0 1.5-.7 2.8-.7s1.7.7 2.8.7c1.2 0 1.9-1 2.6-2 .8-1.2 1.1-2.3 1.1-2.4-.1 0-2.2-.8-2.2-3.3zM14.5 5.8c.6-.7 1-1.7.9-2.7-.9 0-1.9.6-2.5 1.3-.6.7-1.1 1.7-.9 2.7 1 .1 1.9-.5 2.5-1.3z"/></svg>
            {:else}
              {p.mark}
            {/if}
          </span>
          <span class="label-text">{p.label}</span>
          <span class="chev" aria-hidden="true">→</span>
        </button>
      {/each}
    </div>
    <p class="note">OAuth opens in your browser. Condura never sees your password.</p>
  {/if}
</MeridianPage>

<style>
  .identity {
    max-width: 420px;
  }
  .label {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 8px;
  }
  h2 {
    font-family: var(--md-font-display);
    font-size: 28px;
    letter-spacing: -0.04em;
    margin: 0 0 6px;
  }
  .meta {
    color: var(--md-ink-mute);
    margin: 0 0 18px;
  }
  .offline {
    margin: 0 0 14px;
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
    max-width: 42ch;
  }
  .grid {
    display: grid;
    gap: 10px;
    max-width: 400px;
  }
  .provider {
    appearance: none;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    border-radius: 18px;
    padding: 14px 16px;
    font-size: 14px;
    font-weight: 700;
    color: var(--md-ink);
    cursor: pointer;
    text-align: left;
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: 12px;
    transition:
      border-color 180ms var(--md-ease),
      transform 180ms var(--md-spring),
      box-shadow 180ms var(--md-ease),
      background 180ms var(--md-ease);
  }
  .provider:hover {
    border-color: var(--md-cobalt);
    transform: translateY(-2px);
    box-shadow: var(--md-shadow);
  }
  .provider:focus-visible {
    outline: none;
    border-color: var(--md-cobalt);
    box-shadow: var(--md-focus);
  }
  .provider:active {
    transform: scale(0.99);
  }
  .mark {
    width: 32px;
    height: 32px;
    border-radius: 10px;
    display: grid;
    place-items: center;
    font-family: var(--md-font-display);
    font-size: 14px;
    font-weight: 700;
    color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 12%, var(--md-stage));
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 18%, var(--md-line));
  }
  .provider[data-id='github'] .mark {
    color: var(--md-ink);
    background: var(--md-stage);
  }
  .provider[data-id='apple'] .mark {
    color: var(--md-ink);
    background: var(--md-stage);
  }
  .label-text {
    min-width: 0;
  }
  .chev {
    color: var(--md-cobalt);
    font-weight: 600;
    opacity: 0.7;
    transition:
      transform 180ms var(--md-spring),
      opacity 180ms var(--md-ease);
  }
  .provider:hover .chev {
    transform: translateX(3px);
    opacity: 1;
  }
  .note {
    margin-top: 16px;
    font-size: 13px;
    color: var(--md-ink-faint);
    line-height: 1.45;
    max-width: 42ch;
  }
  @media (max-width: 420px) {
    .grid {
      max-width: none;
    }
  }
</style>
