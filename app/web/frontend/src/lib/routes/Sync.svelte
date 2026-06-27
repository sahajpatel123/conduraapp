<script lang="ts">
  // P2P Sync route (Phase 14F). Uses the sync store for peers,
  // pairs, and the pairing flow; replaces the old window.prompt()
  // with PairingModal (+ QR of this device's identity). Auto-
  // refreshes discovered peers every 5 seconds.
  import { ipc } from '../ipc/client'
  import { sync } from '../stores/sync.svelte'
  import { onMount, onDestroy } from 'svelte'
  import PairingModal from '../components/PairingModal.svelte'
  import ConfirmDialog from '../components/ConfirmDialog.svelte'
  import type { SyncStatus } from '../ipc/types'
  import { t } from '../i18n'

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
      lastAction = t('sync.last_action.paired', sync.pendingPeerId || '')
      pairing = false
    }
  }

  function cancelPair(): void {
    sync.clearPending()
    pairing = false
  }

  async function revoke(deviceId: string): Promise<void> {
    confirmAction = async () => {
      if (await sync.revoke(deviceId)) lastAction = t('sync.last_action.revoked', deviceId)
    }
    confirmOpen = true
  }

  async function syncWith(deviceId: string): Promise<void> {
    try {
      const r = await ipc.syncWith(deviceId)
      lastAction = t('sync.last_action.merged', r.merged, deviceId)
      await refreshAll()
    } catch (e) {
      // surface via store-independent path
      lastAction = t('sync.last_action.failed', e)
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
  <header class="page-header">
    <h2>{t('sync.title')}</h2>
    <p class="muted">{t('sync.intro')}</p>
  </header>

  {#if sync.error}
    <p class="error">{sync.error}</p>
  {/if}
  {#if lastAction}
    <p class="success">{lastAction}</p>
  {/if}

  <section class="glass-card card">
    <h3>{t('sync.this_device')}</h3>
    {#if status}
      <div class="kv"><span class="k">{t('sync.device_id')}</span><span class="v mono">{status.device_id || t('sync.unset')}</span></div>
      <div class="kv"><span class="k">{t('sync.name')}</span><span class="v">{status.name || t('sync.unset')}</span></div>
      <div class="kv"><span class="k">{t('sync.running')}</span><span class="v">{status.running ? t('sync.yes') : t('sync.no')}</span></div>
      <div class="kv"><span class="k">{t('sync.entries')}</span><span class="v">{status.entries ?? 0}</span></div>
      <div class="kv"><span class="k">{t('sync.paired')}</span><span class="v">{sync.pairs.length}</span></div>
    {:else}
      <p class="muted">{t('common.loading')}</p>
    {/if}
    <button class="btn btn-ghost btn-sm refresh" onclick={refreshAll} disabled={sync.loading}>{t('sync.refresh')}</button>
  </section>

  <section class="glass-card card">
    <h3>{t('sync.discovered_peers', sync.peers.length)}</h3>
    {#if sync.peers.length === 0}
      <p class="muted">{t('sync.no_peers')}</p>
    {:else}
      <ul class="peer-list">
        {#each sync.peers as p, i (p.device_id)}
          <li class="stagger-item" style="--stagger-index: {i}">
            <span class="mono id">{p.device_id}</span>
            <span class="name">{p.name}</span>
            <span class="row-actions">
              <button class="btn btn-ghost btn-xs" onclick={() => startPair(p.device_id)}>{t('sync.pair')}</button>
              <button class="btn btn-ghost btn-xs" onclick={() => syncWith(p.device_id)}>{t('sync.sync_now')}</button>
            </span>
          </li>
        {/each}
      </ul>
    {/if}
  </section>

  <section class="glass-card card">
    <h3>{t('sync.paired_devices', sync.pairs.length)}</h3>
    {#if sync.pairs.length === 0}
      <p class="muted">{t('sync.no_paired')}</p>
    {:else}
      <ul class="peer-list">
        {#each sync.pairs as p, i (p.device_id)}
          <li class="stagger-item" style="--stagger-index: {i}">
            <span class="mono id">{p.device_id}</span>
            <span class="name">{p.device_name}</span>
            <span class="muted paired-at">{t('sync.paired_at', p.paired_at)}</span>
            <span class="row-actions">
              <button class="btn btn-danger btn-xs" onclick={() => revoke(p.device_id)}>{t('sync.revoke')}</button>
            </span>
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

<ConfirmDialog
  bind:open={confirmOpen}
  title={t('sync.revoke')}
  message={t('sync.revoke_confirm', '')}
  danger={true}
  onconfirm={() => confirmAction?.()}
/>

<style>
  .sync-page {
    padding: var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width);
    margin: 0 auto;
  }
  .page-header {
    animation: fade-in-up var(--transition-slow) var(--ease-out-expo) both;
  }
  .card {
    padding: var(--space-5);
    margin-top: var(--space-5);
  }
  .card h3 {
    font-size: var(--size-lg);
    font-weight: var(--weight-semibold);
    margin-bottom: var(--space-3);
  }
  .refresh {
    margin-top: var(--space-3);
  }
  .peer-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
  .peer-list li {
    display: flex;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-md);
    background: var(--color-bg-hover);
    border: 1px solid transparent;
    transition: background var(--transition-base), border-color var(--transition-base), box-shadow var(--transition-base);
  }
  .peer-list li:hover {
    background: var(--glass-bg);
    border-color: var(--glass-border-hover);
    box-shadow: var(--shadow-glow-accent);
  }
  .id {
    color: var(--color-text-faint);
  }
  .name {
    flex: 1;
    color: var(--color-text);
    font-size: var(--size-sm);
  }
  .paired-at {
    font-size: var(--size-xs);
  }
  .row-actions {
    display: flex;
    gap: var(--space-2);
  }
  .error {
    color: var(--color-error);
    font-size: var(--size-sm);
  }
  .success {
    color: var(--color-success);
    font-size: var(--size-sm);
  }
</style>
