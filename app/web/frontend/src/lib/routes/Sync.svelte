<script lang="ts">
  // P2P Sync route (Phase 14F). Uses the sync store for peers,
  // pairs, and the pairing flow; replaces the old window.prompt()
  // with PairingModal (+ QR of this device's identity). Auto-
  // refreshes discovered peers every 5 seconds.
  import { ipc } from '../ipc/client'
  import { sync } from '../stores/sync.svelte'
  import { onMount, onDestroy } from 'svelte'
  import PairingModal from '../components/PairingModal.svelte'
  import type { SyncStatus } from '../ipc/types'
  import { t } from '../i18n'

  let status = $state<SyncStatus | null>(null)
  let lastAction = $state('')
  let pairing = $state(false)
  let pollTimer: ReturnType<typeof setInterval> | null = null

  async function refreshStatus(): Promise<void> {
    try {
      status = await ipc.syncStatus()
    } catch {
      // status is best-effort; peer/pair errors surface via the store
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
      lastAction = $t('sync.last_action.paired', sync.pendingPeerId || '')
      pairing = false
    }
  }

  function cancelPair(): void {
    sync.clearPending()
    pairing = false
  }

  async function revoke(deviceId: string): Promise<void> {
    if (!confirm($t('sync.revoke_confirm', deviceId))) return
    if (await sync.revoke(deviceId)) lastAction = $t('sync.last_action.revoked', deviceId)
  }

  async function syncWith(deviceId: string): Promise<void> {
    try {
      const r = await ipc.syncWith(deviceId)
      lastAction = $t('sync.last_action.merged', r.merged, deviceId)
      await refreshAll()
    } catch (e) {
      // surface via store-independent path
      lastAction = $t('sync.last_action.failed', e)
    }
  }

  const pendingPeerName = $derived(
    sync.peerById(sync.pendingPeerId)?.name || sync.pendingPeerId
  )

  onMount(() => {
    void refreshAll()
    pollTimer = setInterval(() => void refreshAll(), 5000)
  })
  onDestroy(() => {
    if (pollTimer) clearInterval(pollTimer)
  })
</script>

<div class="sync-page">
  <header>
    <h2>{$t('sync.title')}</h2>
    <p class="muted">
      {$t('sync.intro')}
    </p>
  </header>

  {#if sync.error}
    <p class="error">{sync.error}</p>
  {/if}
  {#if lastAction}
    <p class="success">{lastAction}</p>
  {/if}

  <section class="card">
    <h3>{$t('sync.this_device')}</h3>
    {#if status}
      <div class="kv"><span class="k">{$t('sync.device_id')}</span><span class="v mono">{status.device_id || $t('sync.unset')}</span></div>
      <div class="kv"><span class="k">{$t('sync.name')}</span><span class="v">{status.name || $t('sync.unset')}</span></div>
      <div class="kv"><span class="k">{$t('sync.running')}</span><span class="v">{status.running ? $t('sync.yes') : $t('sync.no')}</span></div>
      <div class="kv"><span class="k">{$t('sync.entries')}</span><span class="v">{status.entries ?? 0}</span></div>
      <div class="kv"><span class="k">{$t('sync.paired')}</span><span class="v">{sync.pairs.length}</span></div>
    {:else}
      <p class="muted">{$t('common.loading')}</p>
    {/if}
    <button class="btn-ghost" onclick={refreshAll} disabled={sync.loading}>{$t('sync.refresh')}</button>
  </section>

  <section class="card">
    <h3>{$t('sync.discovered_peers', sync.peers.length)}</h3>
    {#if sync.peers.length === 0}
      <p class="muted">{$t('sync.no_peers')}</p>
    {:else}
      <ul>
        {#each sync.peers as p (p.device_id)}
          <li>
            <span class="mono id">{p.device_id}</span>
            <span class="name">{p.name}</span>
            <button class="btn-ghost" onclick={() => startPair(p.device_id)}>{$t('sync.pair')}</button>
            <button class="btn-ghost" onclick={() => syncWith(p.device_id)}>{$t('sync.sync_now')}</button>
          </li>
        {/each}
      </ul>
    {/if}
  </section>

  <section class="card">
    <h3>{$t('sync.paired_devices', sync.pairs.length)}</h3>
    {#if sync.pairs.length === 0}
      <p class="muted">{$t('sync.no_paired')}</p>
    {:else}
      <ul>
        {#each sync.pairs as p (p.device_id)}
          <li>
            <span class="mono id">{p.device_id}</span>
            <span class="name">{p.device_name}</span>
            <span class="muted">{$t('sync.paired_at', p.paired_at)}</span>
            <button class="danger" onclick={() => revoke(p.device_id)}>{$t('sync.revoke')}</button>
          </li>
        {/each}
      </ul>
    {/if}
  </section>
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

<style>
  .sync-page { padding: var(--space-5); overflow-y: auto; height: 100%; max-width: 760px; margin: 0 auto; }
  header h2 {
    font-size: var(--size-2xl);
    font-weight: 600;
    margin-bottom: var(--space-2);
    background: var(--color-accent-gradient);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }
  .card {
    background: var(--glass-bg);
    backdrop-filter: var(--glass-blur);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-xl);
    padding: var(--space-5);
    margin-top: var(--space-5);
  }
  .card h3 { font-size: var(--size-lg); font-weight: 600; margin-bottom: var(--space-3); }
  .kv { display: flex; justify-content: space-between; padding: var(--space-2) 0; border-bottom: 1px dotted var(--glass-border); }
  .kv:last-of-type { border-bottom: none; }
  .k { color: var(--color-text-muted); }
  .v { font-weight: 500; }
  ul { list-style: none; padding: 0; margin: 0; }
  li { display: flex; gap: var(--space-3); align-items: center; padding: var(--space-2) 0; }
  .mono { font-family: var(--font-mono); font-size: var(--size-xs); }
  .id { opacity: 0.7; }
  .name { flex: 1; }
  .error { color: var(--color-error, #f87171); }
  .success { color: var(--color-success); }
  .muted { color: var(--color-text-muted); font-size: var(--size-sm); line-height: 1.5; }
  .btn-ghost {
    padding: 6px 12px;
    border-radius: var(--radius-md);
    border: 1px solid var(--glass-border);
    background: transparent;
    color: var(--color-text-muted);
    cursor: pointer;
    font-size: var(--size-sm);
  }
  .btn-ghost:hover:not(:disabled) { color: var(--color-text); border-color: var(--color-accent); }
  .danger {
    padding: 6px 12px;
    border-radius: var(--radius-md);
    border: none;
    background: linear-gradient(135deg, #ef4444, #dc2626);
    color: white;
    cursor: pointer;
    font-size: var(--size-sm);
  }
</style>
