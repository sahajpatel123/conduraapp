<script lang="ts">
  import { ipc } from '../ipc/client'
  import { onMount } from 'svelte'

  let status = $state<any>(null)
  let peers = $state<any[]>([])
  let pairs = $state<any[]>([])
  let error = $state<string | null>(null)
  let loading = $state(false)
  let lastAction = $state<string>('')

  async function refresh() {
    loading = true
    error = null
    try {
      status = await ipc.syncStatus()
      const p = await ipc.syncPeers()
      peers = p.peers || []
      const ps = await ipc.syncListPairs()
      pairs = ps.devices || []
    } catch (e) {
      error = String(e)
    } finally {
      loading = false
    }
  }

  async function pair(deviceId: string) {
    try {
      const r = await ipc.syncPairBegin(deviceId)
      const pin = prompt(`Pairing initiated with ${deviceId}.\n\nEnter the 6-digit PIN shown on the other device:`, '')
      if (!pin) return
      await ipc.syncPairConfirm(deviceId, pin)
      lastAction = `paired ${deviceId}`
      await refresh()
    } catch (e) {
      error = `pair failed: ${e}`
    }
  }

  async function revoke(deviceId: string) {
    if (!confirm(`Revoke device ${deviceId}? Future sync attempts will be rejected.`)) return
    try {
      await ipc.syncRevoke(deviceId)
      lastAction = `revoked ${deviceId}`
      await refresh()
    } catch (e) {
      error = `revoke failed: ${e}`
    }
  }

  async function syncWith(deviceId: string) {
    try {
      const r = await ipc.syncWith(deviceId)
      lastAction = `merged ${r.merged} entries with ${deviceId}`
      await refresh()
    } catch (e) {
      error = `sync failed: ${e}`
    }
  }

  onMount(refresh)
</script>

<div class="sync-page">
  <header>
    <h2>P2P Sync</h2>
    <p class="muted">
      Encrypted device-to-device sync (AES-256-GCM, Ed25519 identity, no central server).
      Devices must be explicitly paired before they can exchange data.
    </p>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}
  {#if lastAction}
    <p class="success">{lastAction}</p>
  {/if}

  <section class="card">
    <h3>This device</h3>
    {#if status}
      <div class="kv"><span class="k">Device ID</span><span class="v mono">{status.device_id || '(unset)'}</span></div>
      <div class="kv"><span class="k">Name</span><span class="v">{status.name || '(unset)'}</span></div>
      <div class="kv"><span class="k">Enabled</span><span class="v">{status.enabled ? 'yes' : 'no'}</span></div>
      <div class="kv"><span class="k">Running</span><span class="v">{status.running ? 'yes' : 'no'}</span></div>
      <div class="kv"><span class="k">Entries</span><span class="v">{status.entries ?? 0}</span></div>
      <div class="kv"><span class="k">Paired</span><span class="v">{status.paired_devices ?? 0}</span></div>
    {:else}
      <p class="muted">Loading…</p>
    {/if}
    <button onclick={refresh} disabled={loading}>Refresh</button>
  </section>

  <section class="card">
    <h3>Discovered peers ({peers.length})</h3>
    {#if peers.length === 0}
      <p class="muted">No peers on the LAN. Devices announce themselves every 30 seconds via UDP broadcast on port 7667.</p>
    {:else}
      <ul>
        {#each peers as p}
          <li>
            <span class="mono id">{p.device_id}</span>
            <span class="name">{p.name}</span>
            <button onclick={() => pair(p.device_id)}>Pair</button>
            <button onclick={() => syncWith(p.device_id)}>Sync now</button>
          </li>
        {/each}
      </ul>
    {/if}
  </section>

  <section class="card">
    <h3>Paired devices ({pairs.length})</h3>
    {#if pairs.length === 0}
      <p class="muted">No paired devices. Pair a discovered peer to start syncing.</p>
    {:else}
      <ul>
        {#each pairs as p}
          <li>
            <span class="mono id">{p.device_id}</span>
            <span class="name">{p.device_name}</span>
            <span class="muted">paired {p.paired_at}</span>
            <button class="danger" onclick={() => revoke(p.device_id)}>Revoke</button>
          </li>
        {/each}
      </ul>
    {/if}
  </section>

  <footer>
    <p class="muted">
      Pairing flow: device A presses "Pair" → A's daemon generates a token + 6-digit PIN →
      B reads the PIN from its overlay → A's user types the PIN to confirm.
      After pairing, the devices exchange CRDT entries encrypted under a
      session key derived from a fresh X25519 ephemeral bound to both
      long-term Ed25519 identities.
    </p>
  </footer>
</div>

<style>
  .sync-page { padding: 16px; }
  .card { padding: 12px; margin: 12px 0; border: 1px solid var(--border, #ddd); border-radius: 6px; }
  .card h3 { margin-top: 0; }
  .kv { display: flex; justify-content: space-between; padding: 2px 0; }
  .k { color: var(--muted, #888); }
  .v { font-weight: 500; }
  ul { list-style: none; padding: 0; margin: 0; }
  li { display: flex; gap: 12px; align-items: center; padding: 6px 0; }
  .mono { font-family: ui-monospace, monospace; font-size: 12px; }
  .id { opacity: 0.7; }
  .name { flex: 1; }
  .error { color: var(--error, #c0392b); }
  .success { color: var(--success, #27ae60); }
  .muted { color: var(--muted, #888); }
  .danger { background: var(--danger, #c0392b); color: white; }
  button { padding: 4px 12px; }
</style>
