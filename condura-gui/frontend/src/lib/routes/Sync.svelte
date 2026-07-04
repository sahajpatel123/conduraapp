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
  import Avatar from '$components/v1/Avatar.svelte'
  import Pill from '$components/v1/Pill.svelte'
  import Button from '$components/v1/Button.svelte'
  import Card from '$components/v1/Card.svelte'
  import EmptyState from '$components/v1/EmptyState.svelte'
  import Inline from '$components/v1/Inline.svelte'
  import Hairline from '$components/v1/Hairline.svelte'

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
        message: 'New device paired.',
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

  const statusVariant = $derived(
    !status?.running ? 'warning' as const : sync.pairs.length === 0 ? 'neutral' as const : 'success' as const
  )

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
    <Inline gap="4" align="end" justify="between" class="title-row">
      <div>
        <h2 class="page-title">P2P sync</h2>
        <p class="lede">
          Device-to-device, E2E encrypted. No central server. Memory, skills, and config — never logs or API keys.
        </p>
      </div>
      <Pill variant={statusVariant} size="md" label={statusLabel} />
    </Inline>

    {#if sync.error}
      <p class="error-banner" role="alert">{sync.error}</p>
    {/if}
    {#if lastAction}
      <p class="success-banner">{lastAction}</p>
    {/if}
  </header>

  <Card variant="raised" padding="4">
    {#snippet children()}
      <Inline gap="4" align="start" justify="between" class="this-device">
        <div class="td-info">
          <h3 class="section-title">This device</h3>
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
        <Button variant="tertiary" size="sm" onclick={refreshAll} loading={sync.loading}>Refresh</Button>
      </Inline>
    {/snippet}
  </Card>

  <div class="columns">
    <section class="col">
      <header class="col-head">
        <h3 class="section-title">
          Discovered peers <span class="count">{sync.peers.length}</span>
        </h3>
        <p class="col-sub">Devices on your LAN via mDNS.</p>
      </header>

      {#if sync.peers.length === 0}
        <Card variant="sunken" padding="5">
          {#snippet children()}
            <EmptyState
              primary="Looking for peers…"
              secondary="Make sure another Condura instance is on the same network and awake. Refresh in a few seconds."
              voice="mono"
            />
          {/snippet}
        </Card>
      {:else}
        <ul class="peer-list">
          {#each sync.peers as p, i (p.device_id)}
            <li class="peer-row" style:--stagger-index={i}>
              <Avatar name={p.name} variant="user" size="md" />
              <div class="peer-info">
                <span class="peer-name">{p.name}</span>
                <span class="peer-meta">{p.device_id}</span>
                <span class="peer-seen">last seen {fmtLastSeen(p.last_seen)}</span>
              </div>
              <Inline gap="2" class="peer-actions">
                <Button variant="primary" size="sm" onclick={() => startPair(p.device_id)} disabled={sync.loading}>Pair</Button>
                <Button variant="tertiary" size="sm" onclick={() => syncWith(p.device_id)}>Sync now</Button>
              </Inline>
            </li>
          {/each}
        </ul>
      {/if}
    </section>

    <section class="col">
      <header class="col-head">
        <h3 class="section-title">
          Paired devices <span class="count">{sync.pairs.length}</span>
        </h3>
        <p class="col-sub">Trusted. Memory and skills flow automatically.</p>
      </header>

      {#if sync.pairs.length === 0}
        <Card variant="sunken" padding="5">
          {#snippet children()}
            <EmptyState
              primary="No paired devices yet"
              secondary="Pair with a peer to start syncing memory, skills, and config across machines."
              voice="mono"
            />
          {/snippet}
        </Card>
      {:else}
        <ul class="peer-list">
          {#each sync.pairs as p, i (p.device_id)}
            <li class="peer-row paired" style:--stagger-index={i}>
              <Avatar name={p.device_name} variant="user" size="md" />
              <div class="peer-info">
                <span class="peer-name">{p.device_name}</span>
                <span class="peer-meta">{p.device_id}</span>
                <span class="peer-seen">paired {fmtLastSeen(p.paired_at)}</span>
              </div>
              <Inline gap="2" class="peer-actions">
                <Button variant="primary" size="sm" onclick={() => syncWith(p.device_id)}>Sync now</Button>
                <Button variant="destructive" size="sm" onclick={() => revoke(p.device_id)}>Revoke</Button>
              </Inline>
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
    max-width: 56rem;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
    background-color: var(--surface-base);
  }

  .page-header {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .title-row {
    width: 100%;
  }

  .page-title {
    font-family: var(--font-serif);
    font-size: var(--text-h2-size);
    font-weight: var(--text-h2-weight);
    letter-spacing: var(--text-h2-tracking);
    color: var(--content-primary);
    margin: 0 0 var(--space-2) 0;
    line-height: var(--text-h2-leading);
  }

  .lede {
    font-size: var(--text-body-size);
    color: var(--content-secondary);
    line-height: 1.55;
    max-width: 40rem;
    margin: 0;
  }

  .section-title {
    font-size: var(--text-body-size);
    font-weight: 500;
    color: var(--content-primary);
    margin: 0 0 var(--space-3) 0;
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .this-device {
    width: 100%;
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
    font-size: var(--text-caption-size);
    font-family: var(--font-mono);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--content-tertiary);
  }

  .kv dd {
    font-size: var(--text-body-sm-size);
    color: var(--content-primary);
    margin: 0;
  }

  .mono {
    font-family: var(--font-mono);
    font-variant-numeric: tabular-nums;
  }

  .columns {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-5);
  }

  @media (max-width: 880px) {
    .columns {
      grid-template-columns: 1fr;
    }
  }

  .col-head {
    margin-bottom: var(--space-3);
  }

  .col-head .section-title {
    margin-bottom: var(--space-1);
  }

  .count {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-secondary);
    background-color: var(--surface-sunken);
    padding: 2px 6px;
    border-radius: var(--radius-pill);
    font-weight: 400;
  }

  .col-sub {
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    margin: 0;
  }

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
    background-color: var(--surface-raised);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    transition:
      background-color var(--duration-fast) var(--ease-standard),
      border-color var(--duration-fast) var(--ease-standard);
    animation: sync-stagger var(--duration-base) var(--ease-standard) both;
    animation-delay: calc(var(--stagger-index, 0) * 60ms);
  }

  @keyframes sync-stagger {
    from { opacity: 0; transform: translateY(6px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .peer-row:hover {
    background-color: var(--surface-sunken);
    border-color: var(--border-strong);
  }

  .peer-row.paired {
    border-color: var(--border-focus);
    background-color: var(--plum-50);
  }

  .peer-info {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .peer-name {
    font-size: var(--text-body-size);
    font-weight: 500;
    color: var(--content-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .peer-meta {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .peer-seen {
    font-size: var(--text-caption-size);
    color: var(--content-secondary);
  }

  .peer-actions {
    flex-shrink: 0;
  }

  .error-banner {
    color: var(--status-error-fg);
    font-size: var(--text-body-sm-size);
    padding: var(--space-2) var(--space-3);
    background-color: var(--status-error-bg);
    border: 1px solid var(--status-error-border);
    border-radius: var(--radius-md);
    margin: 0;
  }

  .success-banner {
    color: var(--status-success-fg);
    font-size: var(--text-body-sm-size);
    margin: 0;
  }

  .muted {
    color: var(--content-tertiary);
    font-size: var(--text-body-sm-size);
    margin: 0;
  }
</style>
