<script lang="ts">
  // Sync — P2P device sync. Header status badge. Two columns:
  // discovered peers (left) and paired devices (right). Pairs use
  // the sync store; PairingModal opens when a pairing is initiated.
  import { ipc } from '../ipc/client'
  import { sync } from '../stores/sync.svelte'
  import { notifications } from '../stores/notifications.svelte'
  import { onMount, onDestroy } from 'svelte'
  import PairingModal from '../components/PairingModal.svelte'
  import ConfirmDialog from '../components/ConfirmDialog.svelte'
  import type { SyncStatus } from '../ipc/types'
  import { Avatar, Badge, Button, Card } from '../components/ui'

  let status = $state<SyncStatus | null>(null)
  let lastAction = $state('')
  let pairing = $state(false)
  let pollTimer: ReturnType<typeof setInterval> | null = null
  let confirmOpen = $state(false)
  let confirmAction = $state<(() => void) | null>(null)

  async function refreshStatus(): Promise<void> {
    try {
      status = await ipc.syncStatus()
    } catch {
      // status is best-effort
    }
  }

  async function refreshAll(): Promise<void> {
    await Promise.all([refreshStatus(), sync.refresh()])
  }

  async function startPair(peerId: string): Promise<void> {
    const res = await sync.pairWith(peerId)
    if (res) pairing = true
  }

  async function confirmPair(pin: string): Promise<void> {
    const res = await sync.confirmPairing(pin)
    if (res) {
      lastAction = `Paired with ${sync.pendingPeerId || ''}.`
      pairing = false
      notifications.push({
        kind: 'success',
        title: 'Paired',
        message: `New device paired.`,
      })
    }
  }

  function cancelPair(): void {
    sync.clearPending()
    pairing = false
  }

  async function revoke(deviceId: string): Promise<void> {
    confirmAction = async () => {
      if (await sync.revoke(deviceId)) {
        lastAction = `Revoked ${deviceId}.`
        notifications.push({ kind: 'info', title: 'Revoked', message: deviceId })
      }
    }
    confirmOpen = true
  }

  async function syncWith(deviceId: string): Promise<void> {
    try {
      const r = await ipc.syncWith(deviceId)
      lastAction = `Merged ${r.merged} entries with ${deviceId}.`
      notifications.push({ kind: 'success', title: 'Synced', message: `Merged ${r.merged} entries.` })
      await refreshAll()
    } catch (e) {
      lastAction = `Sync failed: ${String(e)}`
      notifications.push({ kind: 'error', title: 'Sync failed', message: String(e) })
    }
  }

  const pendingPeerName = $derived(
    sync.peerById(sync.pendingPeerId)?.name || sync.pendingPeerId
  )

  const statusLabel = $derived(
    !status?.running ? 'Paused' : sync.pairs.length === 0 ? 'No peers' : 'Synced'
  )
  const statusTone = $derived(
    !status?.running ? 'warn' : sync.pairs.length === 0 ? 'neutral' : 'success'
  ) as 'warn' | 'neutral' | 'success'

  function fmtLastSeen(iso?: string): string {
    if (!iso) return '—'
    try {
      const t = new Date(iso).getTime()
      const ms = Date.now() - t
      if (ms < 60_000) return 'just now'
      if (ms < 3_600_000) return `${Math.floor(ms / 60_000)}m ago`
      if (ms < 86_400_000) return `${Math.floor(ms / 3_600_000)}h ago`
      return `${Math.floor(ms / 86_400_000)}d ago`
    } catch {
      return iso
    }
  }

  onMount(() => {
    void refreshAll()
    pollTimer = setInterval(() => void refreshAll(), 5000)
  })
  onDestroy(() => {
    if (pollTimer) clearInterval(pollTimer)
  })
</script>

<div class="sync-page">
  <header class="page-header">
    <div class="title-row">
      <div>
        <h2 class="display-h2">P2P sync</h2>
        <p class="lede">
          Device-to-device, E2E encrypted. No central server. Memory, skills, and config — never logs or API keys.
        </p>
      </div>
      <Badge tone={statusTone} size="md" dot pulse={statusTone === 'success'}>
        {statusLabel}
      </Badge>
    </div>

    {#if sync.error}
      <p class="error" role="alert">{sync.error}</p>
    {/if}
    {#if lastAction}
      <p class="success">{lastAction}</p>
    {/if}
  </header>

  <!-- ── This device card ───────────────────────── -->
  <Card elevation="glass" padding="md">
    <div class="this-device">
      <div class="td-info">
        <h3>This device</h3>
        {#if status}
          <dl class="kv-list">
            <div class="kv"><dt>Device ID</dt><dd class="mono">{status.device_id || '—'}</dd></div>
            <div class="kv"><dt>Name</dt><dd>{status.name || '—'}</dd></div>
            <div class="kv"><dt>Running</dt><dd>{status.running ? 'yes' : 'no'}</dd></div>
            <div class="kv"><dt>Entries</dt><dd>{status.entries ?? 0}</dd></div>
            <div class="kv"><dt>Paired</dt><dd>{sync.pairs.length}</dd></div>
          </dl>
        {:else}
          <p class="muted">Loading…</p>
        {/if}
      </div>
      <Button variant="ghost" size="sm" onclick={refreshAll} loading={sync.loading}>Refresh</Button>
    </div>
  </Card>

  <!-- ── Two columns: Peers | Paired ──────────────── -->
  <div class="columns">
    <!-- Peers -->
    <section class="col">
      <header class="col-head">
        <h3>Discovered peers <span class="count mono">{sync.peers.length}</span></h3>
        <p class="col-sub">Devices on your LAN via mDNS.</p>
      </header>

      {#if sync.peers.length === 0}
        <Card elevation={1} padding="lg">
          <div class="empty">
            <div class="empty-icon">
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <circle cx="12" cy="12" r="9" />
                <path d="M3.6 9h16.8M3.6 15h16.8M12 3a14 14 0 010 18M12 3a14 14 0 000 18" />
              </svg>
            </div>
            <h4>Looking for peers…</h4>
            <p>Make sure another Condura instance is on the same network and awake. Refresh in a few seconds.</p>
          </div>
        </Card>
      {:else}
        <ul class="peer-list">
          {#each sync.peers as p, i (p.device_id)}
            <li class="peer-row" style:--stagger-index={i}>
              <Avatar name={p.name} size="md" status="online" />
              <div class="peer-info">
                <span class="peer-name">{p.name}</span>
                <span class="peer-meta mono">{p.device_id}</span>
                <span class="peer-seen">last seen {fmtLastSeen(p.last_seen)}</span>
              </div>
              <div class="peer-actions">
                <Button variant="primary" size="sm" onclick={() => startPair(p.device_id)} disabled={sync.loading}>Pair</Button>
                <Button variant="ghost" size="sm" onclick={() => syncWith(p.device_id)}>Sync now</Button>
              </div>
            </li>
          {/each}
        </ul>
      {/if}
    </section>

    <!-- Paired -->
    <section class="col">
      <header class="col-head">
        <h3>Paired devices <span class="count mono">{sync.pairs.length}</span></h3>
        <p class="col-sub">Trusted. Memory and skills flow automatically.</p>
      </header>

      {#if sync.pairs.length === 0}
        <Card elevation={1} padding="lg">
          <div class="empty">
            <div class="empty-icon">
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M12 2l2 5h5l-4 3 1.5 5L12 12l-4.5 3L9 10 5 7h5z" />
              </svg>
            </div>
            <h4>No paired devices yet</h4>
            <p>Pair with a peer to start syncing memory, skills, and config across machines.</p>
          </div>
        </Card>
      {:else}
        <ul class="peer-list">
          {#each sync.pairs as p, i (p.device_id)}
            <li class="peer-row paired" style:--stagger-index={i}>
              <Avatar name={p.device_name} size="md" status="online" />
              <div class="peer-info">
                <span class="peer-name">{p.device_name}</span>
                <span class="peer-meta mono">{p.device_id}</span>
                <span class="peer-seen">paired {fmtLastSeen(p.paired_at)}</span>
              </div>
              <div class="peer-actions">
                <Button variant="primary" size="sm" onclick={() => syncWith(p.device_id)}>Sync now</Button>
                <Button variant="danger" size="sm" onclick={() => revoke(p.device_id)}>Revoke</Button>
              </div>
            </li>
          {/each}
        </ul>
      {/if}
    </section>
  </div>
</div>

{#if pairing && sync.pendingPin}
  <PairingModal
    deviceId={status?.device_id ?? ''}
    deviceName={status?.name ?? 'this device'}
    peerName={pendingPeerName}
    pin={sync.pendingPin}
    expiresAt={sync.pendingExpiresAt}
    busy={sync.loading}
    error={sync.error}
    onConfirm={confirmPair}
    onCancel={cancelPair}
  />
{/if}

<ConfirmDialog
  bind:open={confirmOpen}
  title="Revoke paired device"
  description="This device will be removed from the trusted set. The other side will lose access on its next sync."
  tone="danger"
  confirmLabel="Revoke"
  onconfirm={() => confirmAction?.()}
/>

<style>
  .sync-page {
    padding: var(--space-6) var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width-wide);
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  .page-header {
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }
  .title-row {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: var(--space-4);
    flex-wrap: wrap;
    margin-bottom: var(--space-3);
  }
  .display-h2 {
    font-family: var(--font-display);
    font-size: var(--size-2xl);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-tight);
    color: var(--text);
    margin: 0 0 var(--space-2) 0;
    line-height: var(--leading-tight);
  }
  .lede {
    font-size: var(--size-md);
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    max-width: 640px;
    margin: 0;
  }

  /* ── This device ───────────────────────────────── */
  .this-device {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: var(--space-4);
  }
  .td-info h3 {
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    color: var(--text);
    margin: 0 0 var(--space-3) 0;
  }
  .kv-list {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: var(--space-2) var(--space-5);
    margin: 0;
  }
  .kv {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .kv dt {
    font-size: var(--size-xs);
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
  }
  .kv dd {
    font-size: var(--size-sm);
    color: var(--text);
    margin: 0;
  }

  /* ── Two columns ───────────────────────────────── */
  .columns {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-5);
  }
  @media (max-width: 880px) {
    .columns { grid-template-columns: 1fr; }
  }

  .col-head {
    margin-bottom: var(--space-3);
  }
  .col-head h3 {
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    color: var(--text);
    margin: 0 0 var(--space-1) 0;
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .count {
    font-size: var(--size-xs);
    color: var(--text-muted);
    background: var(--surface-2);
    padding: 2px 6px;
    border-radius: var(--radius-pill);
  }
  .col-sub {
    font-size: var(--size-xs);
    color: var(--text-muted);
    margin: 0;
  }

  /* ── Peer lists ───────────────────────────────── */
  .peer-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .peer-row {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-3) var(--space-4);
    background: var(--surface-1);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    transition:
      background var(--transition-fast),
      border-color var(--transition-fast),
      box-shadow var(--transition-fast);
    animation: stagger-in var(--transition-base) var(--ease-out-expo) both;
    animation-delay: calc(var(--stagger-index, 0) * 60ms);
  }
  .peer-row:hover {
    background: var(--surface-2);
    border-color: var(--border-strong);
    box-shadow: var(--shadow-sm);
  }
  .peer-row.paired {
    border-color: var(--border-focus);
    background: var(--accent-faint);
  }

  .peer-info {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .peer-name {
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    color: var(--text);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .peer-meta {
    font-size: var(--size-xs);
    color: var(--text-faint);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .peer-seen {
    font-size: var(--size-xs);
    color: var(--text-muted);
  }

  .peer-actions {
    display: flex;
    gap: var(--space-2);
  }

  /* ── Empty state ──────────────────────────────── */
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: var(--space-2);
    padding: var(--space-5);
  }
  .empty-icon {
    width: 48px;
    height: 48px;
    border-radius: var(--radius-lg);
    background: var(--surface-2);
    border: 1px solid var(--border);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-faint);
    margin-bottom: var(--space-2);
  }
  .empty h4 {
    font-size: var(--size-md);
    font-weight: var(--weight-semibold);
    color: var(--text);
    margin: 0;
  }
  .empty p {
    font-size: var(--size-sm);
    color: var(--text-muted);
    line-height: var(--leading-relaxed);
    max-width: 360px;
    margin: 0;
  }

  .error {
    color: var(--error);
    font-size: var(--size-sm);
    padding: var(--space-2) var(--space-3);
    background: var(--error-soft);
    border: 1px solid var(--border-danger);
    border-radius: var(--radius-md);
    margin: var(--space-3) 0 0 0;
  }
  .success {
    color: var(--success);
    font-size: var(--size-sm);
    margin: var(--space-3) 0 0 0;
  }
  .muted {
    color: var(--text-muted);
    font-size: var(--size-sm);
    margin: 0;
  }
</style>
