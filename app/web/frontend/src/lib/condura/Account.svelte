<script lang="ts">
  // ─────────────────────────────────────────────────────────────────────
  // Condura · Account — the optional account route surface.
  // ─────────────────────────────────────────────────────────────────────
  // Implements SCREEN_ACCOUNT.md. The account is *not* required to use
  // Condura. The local agent runs without one; P2P sync works without one.
  // This route exists to make that honest. The hero copy says so.
  //
  // 5 states (S1–S5):
  //   S1  signed-out    — Hero + AuthPicker (Google / GitHub / Email) + footer
  //   S2  signing-in    — Chosen row's icon morphs to a Pulse; others dim 50%
  //   S3  signed-in     — AccountCard + list + threadlink + Sign out + footer
  //   S4  error         — Guide-error above AuthPicker / AccountCard
  //   S5  signing-out   — Confirm popover anchored to Sign out button
  //
  // Apple is intentionally absent (MOAT §1 restraint, defers to v0.2.0).
  // ─────────────────────────────────────────────────────────────────────

  import { onMount } from 'svelte';
  import { account } from '../stores/account.svelte';
  import type { AccountProvider } from '../ipc/types';
  import Glyph from './Glyph.svelte';
  import Pulse from './Pulse.svelte';
  import Thread from './Thread.svelte';
  import Button from './Button.svelte';

  // ── Constants ────────────────────────────────────────────────────────
  // Apple deferred per MOAT §1 restraint. Only OAuth providers the v0.1.0
  // route surfaces are Google and GitHub; Email is always available via
  // magic link (no third-party OAuth config required).
  const OAUTH_REDIRECT = 'condura://auth/callback';
  const MAGIC_REDIRECT = 'https://condura.app/auth/verify';

  type PickerRow = {
    id: 'google' | 'github' | 'email';
    label: string;
    icon: string;
  };

  const ROWS: PickerRow[] = [
    { id: 'google', label: 'Continue with Google', icon: 'google' },
    { id: 'github', label: 'Continue with GitHub', icon: 'github' },
    { id: 'email', label: 'Email me a magic link', icon: 'mail' },
  ];

  // What an account unlocks — three honest reasons (MOAT §4 #5/#6).
  // The wording never implies an account is required.
  const UNLOCKS = [
    { icon: 'hub', text: 'Publish to the public Skills Hub' },
    { icon: 'heart', text: 'Support the project on GitHub Sponsors, Open Collective, or Stripe' },
    { icon: 'lifebuoy', text: 'Open a support ticket via email or Discord' },
  ];

  // ── Local state ─────────────────────────────────────────────────────
  // We keep view-local UI state here and delegate auth state to the store.
  // chosen:  the AuthPicker row currently mid-flow (S2) — null when idle
  // mounted: route mount flag (drives entrance choreography)
  // confirmingSignOut: S5 popover open
  // email:   magic-link email input value
  // linkSent:S2.e.sent — link dispatched, awaiting click
  let chosen = $state<PickerRow['id'] | null>(null);
  let mounted = $state(false);
  let confirmingSignOut = $state(false);
  let email = $state('');
  let linkSent = $state(false);
  let expandError = $state<string | null>(null);

  // Derived: what state are we in? One name per spec §3.
  // Note: account.loading is true during any RPC; we don't use it to
  // drive the state directly — the chosen row + email flow are the truth.
  let view = $derived<'signed-out' | 'signing-in' | 'signed-in' | 'error' | 'signing-out'>(
    confirmingSignOut
      ? 'signing-out'
      : account.isSignedIn
        ? 'signed-in'
        : chosen
          ? 'signing-in'
          : 'signed-out',
  );

  // Derived: error source from store.error (the store sets this on RPC fail).
  // We surface it only when not actively signing-in (S2 owns its own UI).
  let errorText = $derived(account.error);

  // Eyebrow text — swaps between "ACCOUNT" and "ACCOUNT · SIGNED IN" per spec §3.3.
  let eyebrow = $derived(account.isSignedIn ? 'Account · signed in' : 'Account');

  // Hero copy — calm, honest (MOAT §4 #5/#6, no fake enthusiasm).
  const HERO_HEADLINE = 'A Condura account is optional.';
  const HERO_BODY =
    'The local agent works without one. An account unlocks the public Skills Hub, donations, and support.';

  // Provider label used in the "Via {provider}" account summary.
  function providerLabel(p: string): string {
    switch (p) {
      case 'google':
        return 'Google';
      case 'github':
        return 'GitHub';
      case 'magic':
        return 'Email';
      default:
        return p || 'Account';
    }
  }

  // Tier label — localized noun (free / pro / team / enterprise).
  function tierLabel(t: string): string {
    if (!t) return '';
    return t;
  }

  // Single-letter monogram for the avatar fallback.
  function monogram(name: string, emailStr: string): string {
    const src = (name || emailStr || '?').trim();
    return src.slice(0, 1).toUpperCase();
  }

  // Open URL in the system browser via the Wails runtime, falling back to
  // window.open. The OAuth flow lands back at condura://auth/callback.
  function openExternal(url: string): void {
    const w = window as unknown as { runtime?: { BrowserOpenURL?: (u: string) => void } };
    if (w.runtime?.BrowserOpenURL) {
      try {
        w.runtime.BrowserOpenURL(url);
        return;
      } catch {
        /* fall through */
      }
    }
    window.open(url, '_blank', 'noopener,noreferrer');
  }

  // ── Handlers ────────────────────────────────────────────────────────
  // Row click → starts the sign-in flow for that row. We set `chosen`
  // synchronously so the S2 choreography fires immediately; the async
  // RPC resolves into either S3 (callback completes) or S4 (RPC rejects).
  async function startFlow(row: PickerRow['id']): Promise<void> {
    if (chosen) return; // already mid-flow — ignore extra clicks
    expandError = null;
    linkSent = false;
    chosen = row;
    try {
      if (row === 'google') {
        const res = await account.signInWithGoogle(OAUTH_REDIRECT);
        if (res?.url) {
          openExternal(res.url);
        }
        // We don't clear `chosen` here — the OAuth round-trip is in flight
        // until the OS deep-links back to condura://auth/callback and the
        // Shell resolves it. The pulse continues.
        return;
      }
      if (row === 'github') {
        const res = await account.signInWithGitHub(OAUTH_REDIRECT);
        if (res?.url) {
          openExternal(res.url);
        }
        return;
      }
      if (row === 'email') {
        // For email, we don't actually send the link yet — the user must
        // type an address and click Send. We mark the row as chosen (the
        // row's icon morphs) and expand the inline form on the next paint.
        // (We don't auto-call signInWithEmail here.)
        return;
      }
    } catch (e) {
      // RPC rejected before URL opened → fall back to S4.
      expandError = String(e);
      chosen = null;
    }
  }

  // Send the magic link — only valid after the user types a valid address.
  async function sendLink(): Promise<void> {
    if (!emailValid) return;
    expandError = null;
    linkSent = false;
    try {
      const res = await account.signInWithEmail(email.trim(), 'en', MAGIC_REDIRECT);
      if (res?.sent) {
        linkSent = true;
      }
    } catch (e) {
      expandError = String(e);
    }
  }

  // Email format check — RFC 5322 light. The spec disables the button
  // when the format is invalid (test plan §11.1).
  let emailValid = $derived(/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim()));

  // Cancel the email row — collapses it, clears input + error + link-sent.
  function cancelEmail(): void {
    chosen = null;
    email = '';
    linkSent = false;
    expandError = null;
  }

  // Sign-out flow → opens the S5 confirm popover. Confirm calls
  // account.signOut(); on success the store clears status → S1.
  function openSignOutConfirm(): void {
    confirmingSignOut = true;
  }
  function closeSignOutConfirm(): void {
    confirmingSignOut = false;
  }
  async function confirmSignOut(): Promise<void> {
    await account.signOut();
    // closing popover + reset chosen happens because account.isSignedIn
    // flips false on the next render
    confirmingSignOut = false;
    chosen = null;
  }

  // ── Keyboard ────────────────────────────────────────────────────────
  // Tab order follows the visible interactive elements. Esc cancels the
  // popover / email row / pending signing-in.
  function onWindowKey(e: KeyboardEvent): void {
    if (e.key !== 'Escape') return;
    if (confirmingSignOut) {
      e.preventDefault();
      confirmingSignOut = false;
      return;
    }
    if (chosen === 'email' && !linkSent) {
      e.preventDefault();
      cancelEmail();
      return;
    }
    // S2 OAuth: we deliberately do NOT cancel a pending OAuth round-trip
    // (spec §5.2 — Esc during S2 closes nothing).
  }

  // Re-arm the chosen row to clear the S2 pulse after the user comes back
  // without having signed in (e.g., they closed the browser before OAuth
  // completed). For v0.1.0 we leave it; the next page enter will reset.
  // Note: this is intentionally a no-op — Esc on S2 does nothing (spec).

  // ── Lifecycle ───────────────────────────────────────────────────────
  onMount(async () => {
    // First hydration of account state — best-effort; checkStatus handles
    // its own errors and sets account.error.
    try {
      await account.checkStatus();
    } catch {
      /* store handles */
    }
    // Beat-by-beat mount: the route-enter animation is owned by Shell's
    // `.route-enter` (blur-in). We flip `mounted` on the next frame so
    // per-element staggers (list rows, picker rows) can hang off it.
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        mounted = true;
      });
    });
  });
</script>

<svelte:window onkeydown={onWindowKey} />

<main class="account" aria-labelledby="account-route-hero">
  <div class="column" data-view={view}>
    <!-- ── Hero ────────────────────────────────────────────────── -->
    <header class="hero">
      <div class="eyebrow" class:signed-in-eyebrow={account.isSignedIn}>{eyebrow}</div>
      <h1 id="account-route-hero" class="headline">
        {HERO_HEADLINE}
      </h1>
      <p class="lead">
        {HERO_BODY}
      </p>
    </header>

    <div class="hair hero-hair" aria-hidden="true">
      <Thread orientation="h" draw={mounted} glow={false} />
    </div>

    <!-- ── Guide-error (S4) — renders above the AuthPicker or
         AccountCard. Only mounts when account.error is non-null
         AND we're not actively signing-in (S2 owns its own UI). -->
    {#if errorText && view !== 'signing-in'}
      <section class="err-block" role="alert" aria-live="assertive">
        <div class="err-row">
          <Pulse phase="error" size={8} />
          <span class="err-head">{errorHead(errorText)}</span>
        </div>
        <p class="err-sub">{errorBody(errorText)}</p>
        <div class="err-hair" aria-hidden="true"></div>
      </section>
    {/if}

    <!-- ── AuthPicker (S1 + S2) ──────────────────────────────── -->
    {#if !account.isSignedIn}
      <section class="block" aria-labelledby="account-section-unlocks">
        <div class="section-eyebrow" id="account-section-unlocks">
          What an account unlocks
        </div>
        <ul class="unlock-list">
          {#each UNLOCKS as u, i (u.icon)}
            <li
              class="unlock-row"
              class:mounted
              style:--stagger-index={i}
            >
              <Glyph name={u.icon} size={16} class="unlock-glyph" />
              <span class="unlock-text">{u.text}</span>
            </li>
          {/each}
        </ul>
      </section>

      <div class="hair" aria-hidden="true">
        <Thread orientation="h" draw={mounted} glow={false} />
      </div>

      <section class="block" aria-labelledby="account-section-picker">
        <div class="section-eyebrow" id="account-section-picker">Choose a method</div>
        <div class="picker" role="radiogroup" aria-labelledby="account-section-picker">
          {#each ROWS as row, i (row.id)}
            {@const isChosen = chosen === row.id}
            {@const isDimmed = chosen !== null && !isChosen}
            <div
              class="row-wrap"
              class:mounted
              style:--stagger-index={i + 3}
            >
              <!-- The chosen row morphs its icon to a Pulse mid-flow.
                   We render the row as a button; the icon slot is
                   either Glyph or Pulse depending on `chosen`. -->
              <button
                type="button"
                class="auth-row tactile"
                class:chosen={isChosen}
                class:dimmed={isDimmed}
                role="radio"
                aria-checked={isChosen}
                aria-busy={isChosen && account.loading}
                disabled={chosen !== null && !isChosen}
                onclick={() => void startFlow(row.id)}
              >
                <span class="row-icon" aria-hidden="true">
                  {#if isChosen && (row.id === 'google' || row.id === 'github')}
                    <!-- Chosen OAuth row: icon morphs to a thinking Pulse. -->
                    <span class="icon-morph">
                      <Pulse phase="thinking" size={10} />
                    </span>
                  {:else if isChosen && row.id === 'email'}
                    <!-- Email row stays as the mail glyph while expanded —
                         the Pulse appears in the expanded form below. -->
                    <Glyph name={row.icon} size={18} />
                  {:else}
                    <Glyph name={row.icon} size={18} />
                  {/if}
                </span>
                <span class="row-label">{row.label}</span>
                <span class="row-right" aria-hidden="true">
                  {#if !isChosen && !chosen}
                    <!-- Idle pulse on the right edge — each row breathes
                         at its own cadence (the Pulse is per-row, not a
                         shared global). -->
                    <Pulse phase="idle" size={6} />
                  {/if}
                </span>
              </button>

              <!-- Sub-thread under the chosen row (the "this is the one
                   being acted on" gesture — MOAT §3 / spec §4.2). -->
              {#if isChosen}
                <div class="row-thread" aria-hidden="true">
                  <Thread orientation="h" draw={mounted} glow={true} />
                </div>
              {/if}

              <!-- Email expansion — only when the Email row is chosen.
                   The spec calls for the row to grow from 44px → 96px.
                   For v0.1.0 we use an inline reveal block below the
                   row, keeping the row itself at touch-target height. -->
              {#if isChosen && row.id === 'email'}
                <div class="email-expand">
                  <label for="account-email" class="email-label">
                    Email address
                  </label>
                  <div class="email-row">
                    <input
                      id="account-email"
                      type="email"
                      class="email-input"
                      autocomplete="email"
                      placeholder="you@example.com"
                      bind:value={email}
                      aria-invalid={!emailValid && email.length > 0}
                      disabled={linkSent}
                    />
                    <Button
                      variant="primary"
                      size="sm"
                      disabled={!emailValid || linkSent || account.loading}
                      onclick={() => void sendLink()}
                    >
                      {linkSent ? 'Sent ✓' : account.loading ? 'Sending…' : 'Send link'}
                    </Button>
                  </div>
                  {#if linkSent}
                    <p class="email-sent">
                      <Pulse phase="ok" size={6} />
                      <span>
                        Link sent to <span class="mono-inline">{email.trim()}</span> ·
                        check spam, or use another method →
                      </span>
                    </p>
                  {/if}
                  {#if expandError}
                    <p class="email-err">{expandError}</p>
                  {/if}
                  <button
                    type="button"
                    class="email-cancel tactile"
                    onclick={cancelEmail}
                  >
                    Use another method →
                  </button>
                </div>
              {/if}
            </div>
          {/each}
        </div>
      </section>

      <div class="hair" aria-hidden="true">
        <Thread orientation="h" draw={mounted} glow={false} />
      </div>
    {:else}
      <!-- ── AccountCard (S3) ──────────────────────────────── -->
      <section class="block" aria-labelledby="account-card-name">
        <div class="card" data-testid="account-card">
          <div class="card-avatar" aria-hidden="true">
            {#if account.avatarURL}
              <img class="card-avatar-img" src={account.avatarURL} alt="" />
            {:else}
              <span class="card-avatar-fallback">
                {monogram(account.displayName, account.email)}
              </span>
            {/if}
          </div>
          <div class="card-meta">
            <div id="account-card-name" class="card-name">
              {account.displayName || 'Signed in'}
            </div>
            <div class="card-email">{account.email}</div>
            <div class="card-provider">
              <span>Via {providerLabel(account.provider)}</span>
              {#if account.tier}
                <span class="card-sep" aria-hidden="true">·</span>
                <span class="tier-badge" aria-label={`Tier: ${tierLabel(account.tier)}`}>
                  {tierLabel(account.tier)}
                </span>
              {/if}
            </div>
          </div>
        </div>
      </section>

      <div class="hair" aria-hidden="true">
        <Thread orientation="h" draw={mounted} glow={false} />
      </div>

      <section class="block" aria-labelledby="account-section-unlocks-s3">
        <div class="section-eyebrow" id="account-section-unlocks-s3">
          What this account unlocks
        </div>
        <ul class="unlock-list">
          {#each UNLOCKS as u, i (u.icon)}
            <li
              class="unlock-row"
              class:mounted
              style:--stagger-index={i}
            >
              <Glyph name={u.icon} size={16} class="unlock-glyph" />
              <span class="unlock-text">{u.text}</span>
            </li>
          {/each}
        </ul>
      </section>

      <div class="hair" aria-hidden="true">
        <Thread orientation="h" draw={mounted} glow={false} />
      </div>

      <section class="block threadlink-block">
        <a
          class="thread-link"
          href="https://synaptic.app/account"
          target="_blank"
          rel="noopener noreferrer"
          aria-label="Manage your account on condura.app — opens in browser"
        >
          Manage on condura.app →
        </a>
      </section>

      <section class="block signout-block">
        <div class="signout-anchor" bind:this={undefined}>
          <Button
            variant="ghost"
            size="sm"
            onclick={openSignOutConfirm}
            aria-haspopup="dialog"
          >
            Sign out
          </Button>
        </div>
      </section>
    {/if}

    <!-- ── Footer note — ALWAYS rendered (spec §1.4 / §8.4). ──
         Load-bearing honesty copy. The account is genuinely optional;
         the user must see that here, every time, no matter the state. -->
    <footer class="footer-note" aria-label="Account scope note">
      <Pulse phase="idle" size={6} />
      <p>
        Your account is only for Hub + donations + support. Sync is P2P — no
        account needed.
      </p>
    </footer>
  </div>
</main>

<!-- ── Sign-out confirm popover (S5) ────────────────────────────────
     Anchored to the Sign out button via .popover-anchor (the button
     stays in flow; the popover is `position: absolute` relative to the
     `.signout-block`). Per MOAT §2.8: .c-popover is dismiss-on-Esc +
     outside-click. Focus moves to Cancel on open; returns to the
     Sign out button on close. -->
{#if confirmingSignOut}
  <div
    class="popover-scrim"
    role="presentation"
    onclick={closeSignOutConfirm}
    onkeydown={(e) => {
      if (e.key === 'Escape') closeSignOutConfirm();
    }}
  >
    <div
      class="popover"
      role="dialog"
      aria-modal="false"
      aria-labelledby="signout-confirm-q"
    >
      <p id="signout-confirm-q" class="popover-q">
        Sign out of <span class="mono-inline">{account.email}</span>?
      </p>
      <div class="popover-actions">
        <Button
          variant="ghost"
          size="sm"
          onclick={closeSignOutConfirm}
        >
          Cancel
        </Button>
        <Button
          variant="danger"
          size="sm"
          disabled={account.loading}
          onclick={() => void confirmSignOut()}
        >
          {account.loading ? 'Signing out…' : 'Sign out'}
        </Button>
      </div>
      {#if account.error && confirmingSignOut}
        <p class="popover-err">{account.error}</p>
      {/if}
    </div>
  </div>
{/if}

<script context="module" lang="ts">
  // Error copy per spec §3.4 (MOAT §2.6 — guide-error, not poeticize).
  // The store hands us a raw error string; we map by substring to a
  // (head, body) pair. Unknown errors get a generic fallback.
  function errorHead(err: string): string {
    const e = err.toLowerCase();
    if (e.includes('daemon') || e.includes('network') || e.includes('econn')) {
      return "We couldn't reach the daemon.";
    }
    if (e.includes('state') || e.includes('csrf')) {
      return 'Sign-in was interrupted.';
    }
    if (e.includes('email') && e.includes('rate')) {
      return 'Too many tries.';
    }
    if (e.includes('email')) {
      return "Email didn't go through.";
    }
    if (e.includes('token') || e.includes('provider') || e.includes('exchange')) {
      return 'Provider rejected the sign-in.';
    }
    return "Sign-in didn't complete.";
  }

  function errorBody(err: string): string {
    const e = err.toLowerCase();
    if (e.includes('daemon') || e.includes('network') || e.includes('econn')) {
      return 'Condura is running locally. If this keeps happening, restart Condura from the menu bar.';
    }
    if (e.includes('state') || e.includes('csrf')) {
      return "We didn't recognize the return request. Try again — if it persists, clear your browser cookies for condura.app.";
    }
    if (e.includes('email') && e.includes('rate')) {
      return 'Wait a minute, then try again. Or use Google or GitHub.';
    }
    if (e.includes('email')) {
      return 'Check the address — or try Google or GitHub instead.';
    }
    if (e.includes('token') || e.includes('provider') || e.includes('exchange')) {
      return "The token exchange didn't complete. Try another method, or check Discord for known outages.";
    }
    return 'Try another method, or restart Condura.';
  }
</script>

<style>
  /* ── Page geometry ────────────────────────────────────────────────
     Single flowing column. 560px cap keeps the hero copy at a
     comfortable line-length (~66 chars at 15px). No card chrome —
     the surface is document-style with hairline dividers between
     sections (spec §2.2 / §2.5). */
  .account {
    width: 100%;
    height: 100%;
    overflow-y: auto;
    padding: var(--space-9) var(--space-6) var(--space-10);
  }

  .column {
    max-width: 560px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  /* ── Hero ────────────────────────────────────────────────────────── */
  .hero {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .eyebrow {
    font-family: var(--font-mono);
    font-size: var(--text-caption);
    letter-spacing: var(--ls-mono);
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .signed-in-eyebrow {
    color: var(--synapse);
  }
  .headline {
    font-family: var(--font-display);
    font-weight: 400;
    font-style: italic;
    font-size: 28px;
    line-height: 1.15;
    letter-spacing: -0.02em;
    color: var(--content);
    margin: 0;
  }
  .lead {
    font-family: var(--font-sans);
    font-size: var(--text-body);
    line-height: var(--lh-body);
    color: var(--content-soft);
    margin: 0;
    max-width: 52ch;
  }

  /* ── Hairline ────────────────────────────────────────────────────── */
  .hair {
    height: 1px;
    width: 100%;
    overflow: hidden;
    margin: var(--space-2) 0;
  }

  /* ── Section eyebrows + blocks ───────────────────────────────────── */
  .block {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  .section-eyebrow {
    font-family: var(--font-mono);
    font-size: var(--text-caption);
    letter-spacing: var(--ls-mono);
    text-transform: uppercase;
    color: var(--content-faint);
  }

  /* ── Unlocks list ────────────────────────────────────────────────── */
  .unlock-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: 0;
    margin: 0;
    list-style: none;
  }
  .unlock-row {
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
    font-size: 14px;
    line-height: 1.55;
    color: var(--content);
    opacity: 0;
    transform: translateY(4px);
    transition:
      opacity var(--dur) var(--ease) calc(120ms + var(--stagger-index, 0) * 80ms),
      transform var(--dur) var(--ease) calc(120ms + var(--stagger-index, 0) * 80ms);
  }
  .unlock-row.mounted {
    opacity: 1;
    transform: translateY(0);
  }
  .unlock-row :global(.unlock-glyph) {
    color: var(--content-soft);
    flex: none;
    margin-top: 2px;
  }

  /* ── AuthPicker ──────────────────────────────────────────────────── */
  .picker {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .row-wrap {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    opacity: 0;
    transform: translateY(4px);
    transition:
      opacity var(--dur) var(--ease) calc(120ms + var(--stagger-index, 0) * 80ms),
      transform var(--dur) var(--ease) calc(120ms + var(--stagger-index, 0) * 80ms);
  }
  .row-wrap.mounted {
    opacity: 1;
    transform: translateY(0);
  }

  .auth-row {
    display: grid;
    grid-template-columns: 28px 1fr auto;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    height: 44px;
    padding: 0 var(--space-4);
    background: var(--surface-card);
    border: 1px solid var(--hair);
    border-radius: var(--r-control);
    color: var(--content);
    font-family: var(--font-sans);
    font-size: 14px;
    font-weight: 500;
    letter-spacing: -0.005em;
    cursor: pointer;
    transition:
      border-color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      color var(--dur) var(--ease),
      transform var(--dur-fast) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .auth-row:hover:not(:disabled) {
    border-color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 6%, var(--surface-card));
    transform: translateY(-1px);
  }
  .auth-row:active:not(:disabled) {
    transform: translateY(0.5px) scale(0.985);
    filter: brightness(0.96) saturate(1.05);
  }
  .auth-row:focus-visible {
    outline: none;
    box-shadow:
      0 0 0 2px var(--synapse),
      0 0 0 5px var(--pollen-halo);
  }
  .auth-row:disabled {
    cursor: not-allowed;
  }
  .auth-row.chosen {
    border-color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 10%, var(--surface-card));
  }
  .auth-row.dimmed {
    opacity: 0.5;
  }
  .row-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--content);
    width: 18px;
    height: 18px;
  }
  .row-icon :global(svg) {
    transition: opacity var(--dur) var(--ease), transform var(--dur) var(--ease);
  }
  .icon-morph {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    animation: icon-pop 240ms var(--ease) 40ms backwards;
  }
  @keyframes icon-pop {
    from {
      transform: scale(0.4);
      opacity: 0;
    }
    to {
      transform: scale(1);
      opacity: 1;
    }
  }
  .row-label {
    text-align: left;
  }
  .row-right {
    display: inline-flex;
    align-items: center;
    justify-content: flex-end;
    min-width: 14px;
  }

  /* Sub-thread under the chosen row — the signature gesture
     (MOAT §3, spec §4.2). */
  .row-thread {
    height: 2px;
    margin-top: 2px;
    opacity: 0.85;
  }

  /* ── Email expansion ─────────────────────────────────────────────── */
  .email-expand {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-3) var(--space-4) var(--space-4);
    background: var(--surface-sunken);
    border: 1px solid var(--hair);
    border-top: none;
    border-radius: 0 0 var(--r-control) var(--r-control);
    animation: email-expand-in 240ms var(--ease) both;
  }
  @keyframes email-expand-in {
    from {
      opacity: 0;
      transform: translateY(-4px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  .email-label {
    font-family: var(--font-mono);
    font-size: var(--text-caption);
    letter-spacing: var(--ls-mono);
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .email-row {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: var(--space-2);
    align-items: center;
  }
  .email-input {
    appearance: none;
    background: var(--surface-raised);
    border: 1px solid var(--hair);
    border-radius: var(--r-control);
    padding: 10px var(--space-3);
    color: var(--content);
    font-family: var(--font-sans);
    font-size: 14px;
    letter-spacing: -0.005em;
    transition:
      border-color var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .email-input::placeholder {
    color: var(--content-faint);
  }
  .email-input:hover:not(:disabled) {
    border-color: var(--hair-strong);
  }
  .email-input:focus-visible {
    outline: none;
    border-color: var(--synapse);
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .email-input[aria-invalid='true'] {
    border-color: var(--danger);
  }
  .email-input:disabled {
    cursor: not-allowed;
    color: var(--content-mute);
  }
  .email-sent {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    font-family: var(--font-mono);
    font-size: 12px;
    letter-spacing: 0.04em;
    color: var(--content-soft);
    margin: 0;
  }
  .email-err {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    color: var(--danger);
    margin: 0;
  }
  .email-cancel {
    align-self: flex-start;
    background: transparent;
    border: 0;
    padding: 0;
    color: var(--content-faint);
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    cursor: pointer;
    transition: color var(--dur) var(--ease);
  }
  .email-cancel:hover {
    color: var(--synapse);
  }
  .email-cancel:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
    border-radius: var(--r-xs);
  }

  .mono-inline {
    font-family: var(--font-mono);
    font-size: 12px;
    letter-spacing: 0.04em;
    color: var(--content);
  }

  /* ── AccountCard (S3) ────────────────────────────────────────────── */
  .card {
    display: grid;
    grid-template-columns: auto 1fr;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-5);
    background: var(--surface-card);
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
  }
  .card-avatar {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    overflow: hidden;
    background: color-mix(in oklab, var(--synapse) 12%, var(--surface-sunken));
    border: 1px solid var(--hair-strong);
    display: grid;
    place-items: center;
  }
  .card-avatar-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .card-avatar-fallback {
    font-family: var(--font-sans);
    font-size: 18px;
    font-weight: 600;
    color: var(--synapse-glow);
  }
  .card-meta {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }
  .card-name {
    font-family: var(--font-display);
    font-weight: 400;
    font-size: 18px;
    letter-spacing: -0.01em;
    color: var(--content);
    line-height: 1.2;
  }
  .card-email {
    font-family: var(--font-mono);
    font-size: 12px;
    letter-spacing: 0.04em;
    color: var(--content-soft);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .card-provider {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--content-faint);
    margin-top: 4px;
  }
  .card-sep {
    color: var(--content-ghost);
  }
  .tier-badge {
    display: inline-flex;
    align-items: center;
    padding: 2px 8px;
    border-radius: var(--r-pill);
    border: 1px solid var(--hair-strong);
    background: color-mix(in oklab, var(--synapse) 8%, var(--surface-raised));
    color: var(--synapse);
    font-size: 10px;
    letter-spacing: 0.14em;
  }

  /* ── Threadlink (Manage on condura.app) ──────────────────────────── */
  .threadlink-block {
    padding-top: var(--space-2);
  }
  .thread-link {
    position: relative;
    font-family: var(--font-mono);
    font-size: 12px;
    letter-spacing: 0.04em;
    color: var(--content);
    text-decoration: none;
    padding-bottom: 2px;
    cursor: pointer;
    transition: color var(--dur) var(--ease);
  }
  .thread-link::after {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    height: 1px;
    background: var(--synapse);
    transform: scaleX(0);
    transform-origin: left;
    transition: transform var(--dur-slow) var(--ease);
  }
  .thread-link:hover {
    color: var(--synapse);
  }
  .thread-link:hover::after,
  .thread-link:focus-visible::after {
    transform: scaleX(1);
  }
  .thread-link:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
    border-radius: var(--r-xs);
  }

  /* ── Sign out ────────────────────────────────────────────────────── */
  .signout-block {
    padding-top: var(--space-2);
  }
  .signout-anchor {
    display: flex;
    justify-content: flex-end;
  }

  /* ── Guide-error (S4) ────────────────────────────────────────────── */
  .err-block {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-4);
    background: color-mix(in oklab, var(--danger) 6%, var(--surface-card));
    border: 1px solid color-mix(in oklab, var(--danger) 24%, transparent);
    border-radius: var(--r-md);
  }
  .err-row {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
  }
  .err-head {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 18px;
    line-height: 1.2;
    color: var(--content);
  }
  .err-sub {
    font-family: var(--font-mono);
    font-size: 12px;
    letter-spacing: 0.02em;
    line-height: 1.55;
    color: var(--content-soft);
    margin: 0;
  }
  .err-hair {
    height: 1px;
    width: 100%;
    background: linear-gradient(
      90deg,
      color-mix(in oklab, var(--danger) 35%, transparent) 0%,
      color-mix(in oklab, var(--danger) 35%, transparent) 60%,
      transparent 100%
    );
    transform: scaleX(0);
    transform-origin: left;
    animation: err-hair-draw 600ms var(--ease) 120ms forwards;
  }
  @keyframes err-hair-draw {
    to {
      transform: scaleX(1);
    }
  }

  /* ── Footer note (always rendered, spec §1.4 / §8.4) ───────────── */
  .footer-note {
    display: inline-flex;
    align-items: flex-start;
    gap: var(--space-2);
    margin-top: var(--space-3);
    padding-top: var(--space-4);
    border-top: 1px solid var(--hair);
  }
  .footer-note p {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    line-height: 1.55;
    color: var(--content-faint);
    margin: 0;
  }

  /* ── Sign-out confirm popover (S5) ──────────────────────────────── */
  .popover-scrim {
    position: fixed;
    inset: 0;
    z-index: var(--z-modal);
    background: color-mix(in oklab, var(--ink) 12%, transparent);
    display: flex;
    align-items: center;
    justify-content: center;
    animation: popover-fade-in 200ms var(--ease) both;
  }
  @keyframes popover-fade-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  .popover {
    width: min(360px, 92vw);
    padding: var(--space-5);
    background: var(--surface-raised);
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-md);
    box-shadow: var(--shadow-float);
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    animation: popover-pop-in 200ms var(--ease) both;
  }
  @keyframes popover-pop-in {
    from {
      opacity: 0;
      transform: scale(0.96);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }
  .popover-q {
    font-size: 14px;
    line-height: 1.55;
    color: var(--content);
    margin: 0;
  }
  .popover-actions {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
  }
  .popover-err {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    color: var(--danger);
    margin: 0;
  }

  /* ── Reduced-motion: collapse all transitions + animations ─────── */
  @media (prefers-reduced-motion: reduce) {
    .row-wrap,
    .unlock-row {
      opacity: 1;
      transform: none;
      transition: none;
    }
    .email-expand,
    .err-hair,
    .popover,
    .popover-scrim,
    .icon-morph {
      animation: none;
    }
    .thread-link::after {
      transition: none;
    }
  }
</style>