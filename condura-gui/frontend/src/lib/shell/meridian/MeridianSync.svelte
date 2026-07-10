<script lang="ts">
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
</script>

<MeridianPage
  kicker="Devices"
  title="Sync"
  lead="Pair another device. Condura stays local — sync is opt-in and revocable."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void sync.refresh()}>Refresh</button>
  {/snippet}

  {#if showError}
    <div class="md-empty">{sync.error}</div>
  {/if}

  {#if sync.pendingPin}
    <div class="md-panel md-panel-static pin">
      <p class="label">Confirm pairing PIN</p>
      <p class="code">{sync.pendingPin}</p>
      <div class="row">
        <input bind:value={pin} placeholder="Enter PIN on peer" maxlength="8" aria-label="Pairing PIN" />
        <button type="button" class="md-btn md-btn-primary" onclick={() => void sync.confirmPairing(pin)}>
          Confirm
        </button>
        <button type="button" class="md-btn md-btn-ghost" onclick={() => sync.clearPending()}>Cancel</button>
      </div>
    </div>
  {/if}

  <h2 class="sec">Nearby</h2>
  {#if sync.loading && sync.peers.length === 0}
    <div class="md-empty">Looking for peers…</div>
  {:else if sync.peers.length === 0}
    <div class="md-empty empty">
      <p class="empty-title">No peers nearby</p>
      <p class="empty-lead">
        {#if isOffline(sync.error)}
          Connect the daemon to discover devices on your local network.
        {:else}
          Make sure the other device is online and Sync is running.
        {/if}
      </p>
    </div>
  {:else}
    <div class="list md-stagger">
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

  <h2 class="sec">Paired</h2>
  {#if sync.pairs.length === 0}
    <div class="md-empty empty">
      <p class="empty-title">Nothing paired yet</p>
      <p class="empty-lead">Paired devices appear here. You can revoke access anytime.</p>
    </div>
  {:else}
    <div class="list md-stagger">
      {#each sync.pairs as pair (pair.device_id)}
        <div class="item">
          <div class="mono live" aria-hidden="true">{initial(pair.device_name || pair.device_id)}</div>
          <div class="copy">
            <strong>{pair.device_name || pair.device_id}</strong>
            <span class="meta">{pair.device_id}</span>
          </div>
          <button type="button" class="md-btn md-btn-danger" onclick={() => void sync.revoke(pair.device_id)}>
            Revoke
          </button>
        </div>
      {/each}
    </div>
  {/if}
</MeridianPage>

<style>
  .sec {
    font-family: var(--md-font-display);
    font-size: 22px;
    letter-spacing: -0.03em;
    margin: 28px 0 12px;
  }
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
  }
  .empty-title {
    margin: 0;
    font-family: var(--md-font-display);
    font-size: 16px;
    letter-spacing: -0.03em;
    color: var(--md-ink);
  }
  .empty-lead {
    margin: 0;
    max-width: 40ch;
    font-size: 13px;
    line-height: 1.5;
    color: var(--md-ink-mute);
  }
  .list {
    background: var(--md-surface);
    border: 1px solid var(--md-line-strong);
    border-radius: 22px;
    padding: 8px;
    box-shadow: var(--md-shadow);
  }
  .item {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 14px;
    align-items: center;
    margin: 0;
    padding: 14px 12px;
    border-radius: 16px;
    transition: background 160ms var(--md-ease);
  }
  .item:hover {
    background: var(--md-stage);
  }
  .item:focus-within {
    background: var(--md-stage);
    box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--md-cobalt) 35%, transparent);
  }
  .mono {
    width: 42px;
    height: 42px;
    border-radius: 14px;
    display: grid;
    place-items: center;
    font-family: var(--md-font-display);
    font-size: 18px;
    font-weight: 700;
    letter-spacing: -0.04em;
    color: var(--md-cobalt);
    background: color-mix(in oklab, var(--md-cobalt) 12%, var(--md-stage));
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 18%, var(--md-line));
  }
  .mono.live {
    color: var(--md-live);
    background: color-mix(in oklab, var(--md-live) 12%, var(--md-stage));
    border-color: color-mix(in oklab, var(--md-live) 18%, var(--md-line));
  }
  .copy {
    min-width: 0;
  }
  strong {
    display: block;
    font-family: var(--md-font-display);
    font-size: 16px;
    font-weight: 700;
    letter-spacing: -0.03em;
  }
  .meta {
    display: block;
    margin-top: 3px;
    font-family: var(--md-font-mono);
    font-size: 10px;
    color: var(--md-ink-faint);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .pin {
    margin-bottom: 20px;
  }
  .label {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 8px;
  }
  .code {
    font-family: var(--md-font-display);
    font-size: 40px;
    letter-spacing: 0.14em;
    margin: 0 0 16px;
    color: var(--md-cobalt);
    font-variant-numeric: tabular-nums;
  }
  .row {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    align-items: center;
  }
  .row input {
    flex: 1;
    min-width: 140px;
    border: 1px solid var(--md-line-strong);
    border-radius: 999px;
    padding: 10px 14px;
    background: var(--md-stage);
    color: var(--md-ink);
    outline: none;
    transition:
      border-color var(--md-dur) var(--md-ease),
      box-shadow var(--md-dur) var(--md-ease);
  }
  .row input:hover {
    border-color: color-mix(in oklab, var(--md-cobalt) 28%, var(--md-line-strong));
  }
  .row input:focus {
    border-color: var(--md-cobalt);
    box-shadow: var(--md-focus);
  }
  @media (max-width: 560px) {
    .sec {
      font-size: 18px;
      margin: 22px 0 10px;
    }
    .empty {
      padding: 28px 14px;
      gap: 4px;
    }
    .empty-title {
      font-size: 15px;
    }
    .empty-lead {
      font-size: 12.5px;
    }
    .item {
      grid-template-columns: auto 1fr;
      gap: 10px;
      padding: 12px 10px;
    }
    .item .md-btn,
    .item button {
      grid-column: 1 / -1;
      justify-self: stretch;
    }
    .code {
      font-size: 32px;
      letter-spacing: 0.1em;
    }
    .row {
      flex-direction: column;
      align-items: stretch;
    }
    .row input {
      width: 100%;
    }
  }
</style>
