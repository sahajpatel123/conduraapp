<script lang="ts">
  /**
   * Sync — pairing ceremony as the hero; nearby vs paired as two arcs.
   */
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { sync } from '../../stores/sync.svelte'

  let pin = $state('')

  onMount(() => {
    void sync.refresh()
  })

  function initial(name: string): string {
    const t = name.trim()
    return t ? t[0]!.toUpperCase() : '?'
  }

  function isOffline(err: string | null): boolean {
    return !!err && /IPC client not started|not connected|Failed to fetch|daemon/i.test(err)
  }

  const showError = $derived(!!sync.error && !isOffline(sync.error))
  const pinDigits = $derived((sync.pendingPin || '').split(''))
</script>

<MeridianPage
  kicker="Meridian · devices"
  title="Sync"
  lead="Pair another device on your terms. Condura stays local — sync is opt-in and always revocable."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void sync.refresh()}>Refresh</button>
  {/snippet}

  {#if showError}
    <div class="md-empty">{sync.error}</div>
  {/if}

  {#if sync.pendingPin}
    <section class="ceremony" aria-live="polite">
      <p class="cite">pairing ceremony</p>
      <h2>Confirm the PIN</h2>
      <p class="hint">Read these digits on this machine. Enter the same PIN on the peer to seal the pair.</p>
      <div class="digits" aria-label={`PIN ${sync.pendingPin}`}>
        {#each pinDigits as d, i (i)}
          <span>{d}</span>
        {/each}
      </div>
      <div class="row">
        <input bind:value={pin} placeholder="Enter PIN on peer" maxlength="8" aria-label="Pairing PIN" />
        <button type="button" class="md-btn md-btn-primary" onclick={() => void sync.confirmPairing(pin)}>
          Confirm
        </button>
        <button type="button" class="md-btn md-btn-ghost" onclick={() => sync.clearPending()}>Cancel</button>
      </div>
    </section>
  {/if}

  <div class="arcs md-stagger">
    <section class="arc nearby">
      <header>
        <p class="cite">nearby</p>
        <h2>Discover</h2>
        <p class="hint">Devices on your local network offering a pair.</p>
      </header>
      {#if sync.loading && sync.peers.length === 0}
        <div class="md-empty soft">Looking for peers…</div>
      {:else if sync.peers.length === 0}
        <div class="md-empty soft">
          <p class="empty-title">No peers nearby</p>
          <p class="empty-lead">
            {#if isOffline(sync.error)}
              Connect the daemon to discover devices.
            {:else}
              Make sure the other device is online with Sync running.
            {/if}
          </p>
        </div>
      {:else}
        <div class="list">
          {#each sync.peers as peer (peer.device_id)}
            <div class="item">
              <div class="mono" aria-hidden="true">{initial(peer.name || peer.device_id)}</div>
              <div class="copy">
                <strong>{peer.name || peer.device_id}</strong>
                <span class="meta">{peer.device_id}</span>
              </div>
              <button type="button" class="md-btn md-btn-primary" onclick={() => void sync.pairWith(peer.device_id)}>
                Pair
              </button>
            </div>
          {/each}
        </div>
      {/if}
    </section>

    <section class="arc paired">
      <header>
        <p class="cite">paired</p>
        <h2>Trusted line</h2>
        <p class="hint">Revoke anytime — cutting the line is immediate.</p>
      </header>
      {#if sync.pairs.length === 0}
        <div class="md-empty soft">
          <p class="empty-title">Nothing paired yet</p>
          <p class="empty-lead">Confirmed devices appear here as live nodes.</p>
        </div>
      {:else}
        <div class="list">
          {#each sync.pairs as pair (pair.device_id)}
            <div class="item">
              <div class="mono live" aria-hidden="true">{initial(pair.device_name || pair.device_id)}</div>
              <div class="copy">
                <strong>{pair.device_name || pair.device_id}</strong>
                <span class="meta">{pair.device_id}</span>
              </div>
              <button type="button" class="md-btn md-btn-danger" onclick={() => void sync.revoke(pair.device_id)}>
                Cut line
              </button>
            </div>
          {/each}
        </div>
      {/if}
    </section>
  </div>
</MeridianPage>

<style>
  .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 8px;
  }
  .ceremony {
    margin-bottom: 20px;
    padding: 28px 24px;
    border-radius: 24px;
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 35%, transparent);
    background:
      radial-gradient(120% 80% at 50% 0%, color-mix(in oklab, var(--md-cobalt) 14%, transparent), transparent 60%),
      var(--md-surface);
    box-shadow: var(--md-shadow-lift);
    text-align: center;
  }
  .ceremony h2 {
    font-family: var(--md-font-display);
    font-size: 28px;
    letter-spacing: -0.04em;
    margin: 0 0 8px;
  }
  .hint {
    margin: 0 auto 18px;
    max-width: 42ch;
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
  }
  .digits {
    display: inline-flex;
    gap: 8px;
    margin-bottom: 20px;
  }
  .digits span {
    width: 48px;
    height: 58px;
    display: grid;
    place-items: center;
    border-radius: 14px;
    font-family: var(--md-font-display);
    font-size: 28px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    color: var(--md-cobalt);
    background: var(--md-stage);
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 22%, var(--md-line));
  }
  .row {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    justify-content: center;
    align-items: center;
  }
  .row input {
    min-width: 160px;
    border: 1px solid var(--md-line-strong);
    border-radius: 999px;
    padding: 10px 14px;
    background: var(--md-stage);
    outline: none;
  }
  .row input:focus {
    border-color: var(--md-cobalt);
    box-shadow: var(--md-focus);
  }

  .arcs {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 14px;
  }
  .arc {
    border-radius: 22px;
    border: 1px solid var(--md-line);
    background: color-mix(in oklab, var(--md-surface) 85%, transparent);
    padding: 18px;
    min-height: 220px;
  }
  .arc header {
    margin-bottom: 14px;
  }
  .arc h2 {
    font-family: var(--md-font-display);
    font-size: 22px;
    letter-spacing: -0.03em;
    margin: 0 0 6px;
  }
  .arc.paired {
    border-color: color-mix(in oklab, var(--md-live) 22%, var(--md-line));
  }
  .md-empty.soft {
    background: transparent;
    border: 0;
    box-shadow: none;
    padding: 24px 8px;
  }
  .empty-title {
    margin: 0;
    font-family: var(--md-font-display);
    font-size: 15px;
    letter-spacing: -0.03em;
  }
  .empty-lead {
    margin: 6px 0 0;
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
    max-width: 32ch;
  }
  .list {
    display: grid;
    gap: 8px;
  }
  .item {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 12px;
    align-items: center;
    padding: 10px;
    border-radius: 14px;
    background: var(--md-stage);
    border: 1px solid var(--md-line);
  }
  .mono {
    width: 40px;
    height: 40px;
    border-radius: 13px;
    display: grid;
    place-items: center;
    font-family: var(--md-font-display);
    font-size: 16px;
    font-weight: 700;
    color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 12%, var(--md-surface));
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 18%, var(--md-line));
  }
  .mono.live {
    color: var(--md-live);
    background: color-mix(in oklab, var(--md-live) 12%, var(--md-surface));
    border-color: color-mix(in oklab, var(--md-live) 18%, var(--md-line));
  }
  .copy {
    min-width: 0;
  }
  strong {
    display: block;
    font-family: var(--md-font-display);
    font-size: 15px;
    letter-spacing: -0.03em;
  }
  .meta {
    display: block;
    margin-top: 2px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    color: var(--md-ink-faint);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  @media (max-width: 720px) {
    .arcs {
      grid-template-columns: 1fr;
    }
    .digits span {
      width: 40px;
      height: 50px;
      font-size: 22px;
    }
  }
</style>
