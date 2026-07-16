<script lang="ts">
  /**
   * Sync — opt-in device meridian.
   * Signature: PIN ceremony hero · nearby / paired arcs · cut-the-line revoke.
   *
   * Live discovery: poll peers every 5s while mounted (legacy Sync).
   * Ceremony: honest TTL from pair_begin expires_in + local countdown.
   */
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { sync } from '../../stores/sync.svelte'

  let pin = $state('')
  let revoking = $state<string | null>(null)
  let pairing = $state<string | null>(null)
  let syncingId = $state<string | null>(null)
  let lastSyncNote = $state('')
  let confirmRevokeId = $state<string | null>(null)
  let remainingSec = $state(0)
  let pinEl = $state<HTMLInputElement | null>(null)

  onMount(() => {
    void sync.refresh()
    const poll = setInterval(() => {
      void sync.refresh({ quiet: true })
    }, 5000)
    return () => clearInterval(poll)
  })

  // Ceremony countdown — clear pending when PIN window closes.
  $effect(() => {
    if (!sync.pendingPin || !sync.pendingExpiresAt) {
      remainingSec = 0
      return
    }
    const tick = () => {
      const ms = new Date(sync.pendingExpiresAt).getTime() - Date.now()
      if (!Number.isFinite(ms)) {
        remainingSec = 0
        return
      }
      // Leave pendingPin set so the expired plate can offer "Back to peers".
      remainingSec = Math.max(0, Math.ceil(ms / 1000))
    }
    tick()
    const id = setInterval(tick, 1000)
    return () => clearInterval(id)
  })

  /** Auto-focus the PIN input the moment a pairing ceremony appears.
   *  The user just walked over to this device and read the PIN off it —
   *  making them click before typing defeats the purpose. */
  $effect(() => {
    if (!sync.pendingPin || pinExpired) return
    queueMicrotask(() => {
      pinEl?.focus({ preventScroll: true })
      pinEl?.select()
    })
  })

  function initial(name: string): string {
    const t = name.trim()
    return t ? t[0]!.toUpperCase() : '?'
  }

  function isOffline(err: string | null): boolean {
    return !!err && /IPC client not started|not connected|Failed to fetch|daemon/i.test(err)
  }

  function fmtRemaining(sec: number): string {
    if (sec <= 0) return 'expired'
    const m = Math.floor(sec / 60)
    const s = sec % 60
    return `${m}:${String(s).padStart(2, '0')}`
  }

  const showError = $derived(!!sync.error && !isOffline(sync.error))
  const pinDigits = $derived((sync.pendingPin || '').split(''))
  const pinExpired = $derived(!!sync.pendingPin && remainingSec <= 0)
  const pinUrgent = $derived(remainingSec > 0 && remainingSec <= 30)
  const liveNote = $derived(
    sync.pendingPin
      ? pinExpired
        ? 'PIN expired — start pair again'
        : `Ceremony open · ${fmtRemaining(remainingSec)} left`
      : isOffline(sync.error)
        ? 'Daemon offline — discovery paused'
        : `${sync.peers.length} nearby · ${sync.pairs.length} paired`
  )
  const peerLabel = $derived(
    sync.peerById(sync.pendingPeerId)?.name || sync.pendingPeerId || 'peer'
  )

  async function pair(id: string): Promise<void> {
    pairing = id
    pin = ''
    try {
      await sync.pairWith(id)
    } finally {
      pairing = null
    }
  }

  async function confirm(): Promise<void> {
    if (!pin.trim() || pinExpired) return
    await sync.confirmPairing(pin)
    pin = ''
  }

  function requestRevoke(id: string): void {
    confirmRevokeId = id
  }

  function cancelRevoke(): void {
    if (revoking) return
    confirmRevokeId = null
  }

  async function confirmRevoke(): Promise<void> {
    const id = confirmRevokeId
    if (!id || revoking) return
    revoking = id
    try {
      await sync.revoke(id)
      confirmRevokeId = null
    } finally {
      revoking = null
    }
  }

  async function revoke(id: string): Promise<void> {
    requestRevoke(id)
  }

  async function syncNow(id: string): Promise<void> {
    if (syncingId || revoking) return
    syncingId = id
    lastSyncNote = ''
    try {
      const r = await sync.syncWith(id)
      if (r) {
        lastSyncNote =
          r.merged === 0
            ? 'Sync complete · nothing new to merge'
            : `Sync complete · merged ${r.merged} entr${r.merged === 1 ? 'y' : 'ies'}`
      }
    } finally {
      syncingId = null
    }
  }
</script>

<MeridianPage
  kicker="Meridian · devices"
  title="Sync"
  lead="Pair another device on your terms. Condura stays local — sync is opt-in, PIN-sealed, and always revocable."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void sync.refresh()}>Refresh</button>
  {/snippet}

  <div class="desk md-stagger">
    <p class="contract" class:hot={!!sync.pendingPin && !pinExpired} class:off={isOffline(sync.error) || pinExpired}>
      <span class="live-dot" aria-hidden="true"></span>
      {liveNote}. Cutting a line is immediate — no silent re-pair.
    </p>

    {#if showError}
      <div class="md-empty">{sync.error}</div>
    {/if}

    {#if lastSyncNote && !showError}
      <p class="sync-note">{lastSyncNote}</p>
    {/if}

    {#if sync.pendingPin}
      <section class="ceremony" class:expired={pinExpired} aria-live="polite">
        <p class="cite">pairing ceremony</p>
        <h2>{pinExpired ? 'PIN expired' : 'Confirm the PIN'}</h2>
        <p class="hint">
          {#if pinExpired}
            The daemon closed this window. Pair again from the peer list to mint a fresh PIN.
          {:else}
            Read these digits on this machine. Enter the same PIN on
            <strong>{peerLabel}</strong>
            before the timer ends to seal the pair.
          {/if}
        </p>
        {#if !pinExpired}
          <div class="digits" aria-label={`PIN ${sync.pendingPin}`}>
            {#each pinDigits as d, i (i)}
              <span style={`animation-delay: ${i * 40}ms`}>{d || '·'}</span>
            {/each}
          </div>
          <p class="ttl" class:urgent={pinUrgent} aria-live="polite">
            Expires in {fmtRemaining(remainingSec)}
          </p>
          <div class="row">
            <input
              bind:this={pinEl}
              bind:value={pin}
              placeholder="Enter PIN to seal"
              maxlength="8"
              inputmode="numeric"
              autocomplete="one-time-code"
              aria-label="Pairing PIN"
            />
            <button
              type="button"
              class="md-btn md-btn-primary"
              disabled={!pin.trim() || pinExpired}
              onclick={() => void confirm()}
            >
              Seal pair
            </button>
            <button
              type="button"
              class="md-btn md-btn-ghost"
              onclick={() => {
                pin = ''
                sync.clearPending()
              }}
            >
              Cancel
            </button>
          </div>
        {:else}
          <div class="row">
            <button
              type="button"
              class="md-btn md-btn-primary"
              onclick={() => {
                pin = ''
                sync.clearPending()
              }}
            >
              Back to peers
            </button>
          </div>
        {/if}
      </section>
    {/if}

    {#if confirmRevokeId}
      <section class="revoke-plate" aria-live="polite">
        <p class="cite">cut the line</p>
        <h2>Revoke this device?</h2>
        <p class="hint">
          Immediate and local. The peer must pair again with a new PIN to reconnect.
        </p>
        <div class="row">
          <button
            type="button"
            class="md-btn md-btn-danger"
            disabled={!!revoking}
            onclick={() => void confirmRevoke()}
          >
            {revoking ? 'Cutting…' : 'Cut line'}
          </button>
          <button type="button" class="md-btn md-btn-ghost" disabled={!!revoking} onclick={cancelRevoke}>
            Keep paired
          </button>
        </div>
      </section>
    {/if}

    <div class="arcs">
      <section class="arc nearby">
        <header>
          <p class="cite">nearby · discover</p>
          <h2>On the LAN</h2>
          <p class="hint">Devices offering a pair. Refreshes every few seconds while this page is open.</p>
        </header>
        {#if sync.loading && sync.peers.length === 0}
          <div class="md-empty soft">Looking for peers…</div>
        {:else if sync.peers.length === 0}
          <div class="md-empty soft">
            <p class="empty-title">No peers nearby</p>
            <p class="empty-lead">
              {#if isOffline(sync.error)}
                Connect the daemon to discover devices on your local network.
              {:else}
                Make sure the other device is online with Sync running — discovery updates live.
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
                <button
                  type="button"
                  class="md-btn md-btn-primary"
                  disabled={!!sync.pendingPin || pairing === peer.device_id}
                  onclick={() => void pair(peer.device_id)}
                >
                  {pairing === peer.device_id ? 'Starting…' : 'Pair'}
                </button>
              </div>
            {/each}
          </div>
        {/if}
      </section>

      <section class="arc paired">
        <header>
          <p class="cite">paired · trusted line</p>
          <h2>Sealed devices</h2>
          <p class="hint">Sync when online on the LAN. Revoke anytime — the cut is immediate and local.</p>
        </header>
        {#if sync.pairs.length === 0}
          <div class="md-empty soft">
            <p class="empty-title">Nothing paired yet</p>
            <p class="empty-lead">Confirmed devices appear here as live nodes on the meridian.</p>
          </div>
        {:else}
          <div class="list">
            {#each sync.pairs as pairRow (pairRow.device_id)}
              {@const online = sync.isDiscoverable(pairRow.device_id)}
              <div class="item paired-item">
                <div class="mono live" aria-hidden="true">{initial(pairRow.device_name || pairRow.device_id)}</div>
                <div class="copy">
                  <strong>{pairRow.device_name || pairRow.device_id}</strong>
                  <span class="meta">
                    {online ? 'on LAN' : 'not discoverable'}
                    · {pairRow.device_id}
                  </span>
                </div>
                <div class="item-actions">
                  <button
                    type="button"
                    class="md-btn md-btn-primary mini"
                    disabled={
                      !!syncingId ||
                      !!revoking ||
                      confirmRevokeId === pairRow.device_id ||
                      !online
                    }
                    title={online
                      ? 'Pull and merge with this device'
                      : 'Device must be online and announcing on the LAN'}
                    onclick={() => void syncNow(pairRow.device_id)}
                  >
                    {syncingId === pairRow.device_id ? 'Syncing…' : 'Sync now'}
                  </button>
                  <button
                    type="button"
                    class="md-btn md-btn-danger mini"
                    disabled={
                      revoking === pairRow.device_id ||
                      confirmRevokeId === pairRow.device_id ||
                      !!syncingId
                    }
                    onclick={() => void revoke(pairRow.device_id)}
                  >
                    {revoking === pairRow.device_id ? 'Cutting…' : 'Cut line'}
                  </button>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </section>
    </div>
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
    border-color: color-mix(in oklab, var(--md-cobalt) 35%, transparent);
    background: color-mix(in oklab, var(--md-cobalt) 6%, var(--md-surface));
  }
  .contract.off {
    border-color: color-mix(in oklab, var(--md-halt) 25%, var(--md-line));
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
    background: var(--md-cobalt);
    box-shadow: none;
  }
  .contract:not(.hot):not(.off) .live-dot {
    background: var(--md-live);
    box-shadow: none;
  }
  .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 8px;
  }
  .ceremony,
  .revoke-plate {
    padding: 22px 20px;
    border-radius: 12px;
    border: 1px solid color-mix(in oklab, var(--md-cobalt) 28%, var(--md-line));
    background: var(--md-surface);
    box-shadow: none;
    text-align: center;
  }
  .ceremony.expired,
  .revoke-plate {
    border-color: color-mix(in oklab, var(--md-halt) 28%, var(--md-line));
    background: var(--md-surface);
  }
  .ceremony h2,
  .revoke-plate h2 {
    font-family: var(--md-font-display);
    font-size: 28px;
    letter-spacing: -0.04em;
    margin: 0 0 8px;
  }
  .hint {
    margin: 0 auto 18px;
    max-width: 44ch;
    font-size: 13px;
    line-height: 1.45;
    color: var(--md-ink-mute);
  }
  .hint strong {
    display: inline;
    font-family: inherit;
    font-size: inherit;
    letter-spacing: inherit;
    color: var(--md-ink);
  }
  .digits {
    display: inline-flex;
    gap: 8px;
    margin-bottom: 12px;
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
    animation: md-rise 320ms var(--md-ease) both;
  }
  .ttl {
    margin: 0 0 16px;
    font-family: var(--md-font-mono);
    font-size: 12px;
    letter-spacing: 0.08em;
    color: var(--md-ink-mute);
  }
  .ttl.urgent {
    color: var(--md-halt);
    font-weight: 700;
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
    border: 1px solid var(--md-line);
    border-radius: 8px;
    padding: 9px 12px;
    background: var(--md-stage);
    outline: none;
    font-family: var(--md-font-mono);
    letter-spacing: 0.12em;
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
    border-radius: 12px;
    border: 1px solid var(--md-line);
    background: var(--md-surface);
    padding: 16px;
    min-height: 240px;
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
  .sync-note {
    margin: 0;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    color: var(--md-live);
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
  .item-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    justify-content: flex-end;
  }
  .mini {
    padding: 8px 12px;
    font-size: 12px;
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
  @keyframes md-rise {
    from {
      opacity: 0;
      transform: translateY(6px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
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
  @media (prefers-reduced-motion: reduce) {
    .digits span {
      animation: none !important;
    }
  }
</style>
