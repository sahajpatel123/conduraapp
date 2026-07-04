<script lang="ts">
  // Condura Sync — P2P device pairing as a garden of nodes threaded together.
  // THIS device is the green node at the centre (the agent's presence); each
  // paired device is a pollen node around it; the threads between them ARE the
  // sync links — they draw in on mount and breathe at a 9s cadence (a link is
  // alive). Pairing begins from a discovered peer; the pending PIN shows in a
  // centred card with a pollen TTL ring depleting toward expiry. Revocation
  // frays the thread and fades the node. No central server is visible — only
  // the threads between your machines.
  import { onMount, onDestroy } from 'svelte';
  import { sync } from '../stores/sync.svelte';
  import { ipc } from '../ipc/client';
  import type { SyncStatus, SyncPair } from '../ipc/types';
  import QRCode from 'qrcode';
  import Thread from './Thread.svelte';
  import Pulse from './Pulse.svelte';
  import Glyph from './Glyph.svelte';
  import Button from './Button.svelte';

  // ── this-device identity (the green node's QR face) ──
  let status = $state<SyncStatus | null>(null);
  let qrDataUrl = $state('');

  // ── local UI state ──
  let selectedPair = $state<SyncPair | null>(null);
  let pinInput = $state('');
  let frayedIds = $state<string[]>([]);
  let pollTimer: ReturnType<typeof setInterval> | null = null;

  // ── countdown ring (pollen, depleting toward expiry) ──
  let pendingTotalMs = $state(1);
  let remainingMs = $state(0);
  let ringPct = $derived(
    pendingTotalMs > 0 ? Math.min(100, Math.max(0, (remainingMs / pendingTotalMs) * 100)) : 0,
  );
  let ringLabel = $derived(fmtRemaining(remainingMs));
  let expired = $derived(remainingMs <= 0);

  // ── store-derived views ──
  let pending = $derived(!!sync.pendingPin && !!sync.pendingPeerId);
  let pendingPeerName = $derived(sync.peerById(sync.pendingPeerId)?.name || sync.pendingPeerId);
  let pinReady = $derived(/^\d{4,8}$/.test(pinInput.trim()));
  type Phase = 'idle' | 'thinking' | 'awaiting' | 'acting' | 'consent' | 'error' | 'ok';
  let badgePhase: Phase = $derived(
    !status?.running ? 'error' : sync.pairs.length === 0 ? 'thinking' : 'idle',
  );
  let statusLabel = $derived(
    !status?.running ? 'Sync paused' : sync.pairs.length === 0 ? 'Awaiting a peer' : `${sync.pairs.length} paired`,
  );

  // ── node positions (this device centred; pairs radial around it) ──
  const THIS: { x: number; y: number } = { x: 50, y: 50 };
  const RADIUS = 34; // % from centre — keeps labels inside the canvas
  let nodes = $derived(
    sync.pairs.map((pair, i) => {
      const n = sync.pairs.length;
      const angle = n === 1 ? -Math.PI / 2 : (i / n) * Math.PI * 2 - Math.PI / 2;
      return {
        pair,
        x: THIS.x + Math.cos(angle) * RADIUS,
        y: THIS.y + Math.sin(angle) * RADIUS,
        index: i,
      };
    }),
  );

  function nodeFor(pair: SyncPair | null): { x: number; y: number } | null {
    if (!pair) return null;
    return nodes.find((nd) => nd.pair.device_id === pair.device_id) ?? null;
  }

  // ── helpers ──
  function fmtRemaining(ms: number): string {
    if (ms <= 0) return 'expired';
    const s = Math.ceil(ms / 1000);
    return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;
  }
  function fmtSeen(iso: string): string {
    if (!iso) return '—';
    try {
      const t = new Date(iso).getTime();
      const ms = Date.now() - t;
      if (ms < 60_000) return 'just now';
      if (ms < 3_600_000) return `${Math.floor(ms / 60_000)}m ago`;
      if (ms < 86_400_000) return `${Math.floor(ms / 3_600_000)}h ago`;
      return `${Math.floor(ms / 86_400_000)}d ago`;
    } catch {
      return iso;
    }
  }

  // ── actions ──
  async function startPair(peerId: string): Promise<void> {
    pinInput = '';
    selectedPair = null;
    await sync.pairWith(peerId);
  }
  async function confirmPair(): Promise<void> {
    if (!pinReady) return;
    const v = pinInput.trim();
    pinInput = '';
    await sync.confirmPairing(v);
  }
  function cancelPending(): void {
    sync.clearPending();
    pinInput = '';
  }
  async function doRevoke(deviceId: string): Promise<void> {
    selectedPair = null;
    // fray the thread + fade the node while the RPC is in flight
    frayedIds = [...frayedIds, deviceId];
    const ok = await sync.revoke(deviceId);
    // on success the store refreshes and the pair leaves sync.pairs;
    // on failure, restore the thread so the user can retry.
    if (!ok) frayedIds = frayedIds.filter((id) => id !== deviceId);
  }

  // ── countdown ring ticker (re-armed per pending pairing) ──
  $effect(() => {
    const expiresAt = sync.pendingExpiresAt;
    const pin = sync.pendingPin;
    if (!expiresAt || !pin) {
      remainingMs = 0;
      pendingTotalMs = 1;
      return;
    }
    const total = Math.max(1, new Date(expiresAt).getTime() - Date.now());
    pendingTotalMs = total;
    remainingMs = Math.max(0, new Date(expiresAt).getTime() - Date.now());
    const timer = setInterval(() => {
      remainingMs = Math.max(0, new Date(expiresAt).getTime() - Date.now());
    }, 250);
    return () => clearInterval(timer);
  });

  // ── this-device QR (re-rendered when identity arrives / changes) ──
  $effect(() => {
    const id = status?.device_id ?? '';
    const name = status?.name ?? '';
    if (!id) {
      qrDataUrl = '';
      return;
    }
    let alive = true;
    const payload = JSON.stringify({ v: 1, device_id: id, name });
    QRCode.toDataURL(payload, { margin: 1, width: 160 })
      .then((url) => {
        if (alive) qrDataUrl = url;
      })
      .catch(() => {
        if (alive) qrDataUrl = '';
      });
    return () => {
      alive = false;
    };
  });

  // ── lifecycle ──
  async function refreshAll(): Promise<void> {
    try {
      const [s] = await Promise.all([ipc.syncStatus(), sync.refresh()]);
      status = s;
    } catch {
      // sync.error surfaces inline; status is best-effort
    }
  }

  onMount(() => {
    void refreshAll();
    pollTimer = setInterval(() => void refreshAll(), 5000);
  });
  onDestroy(() => {
    if (pollTimer) clearInterval(pollTimer);
  });
</script>

<svelte:window
  onkeydown={(e) => {
    if (e.key === 'Escape') {
      if (selectedPair) selectedPair = null;
      else if (pending) cancelPending();
    }
  }}
/>

<section class="sync" aria-label="P2P device sync">
  <header class="s-head">
    <div class="s-head-text">
      <div class="eyebrow">P2P Sync</div>
      <h1 class="headline">Your machines, <span class="alive">threaded.</span></h1>
      <p class="lead">
        Memory, skills, and config flow directly between your devices — encrypted, no central
        server. Logs and API keys never leave a machine.
      </p>
    </div>
    <div class="s-head-badge">
      <Pulse phase={badgePhase} size={8} />
      <span class="badge-text">{statusLabel}</span>
    </div>
  </header>

  <div class="rule"><Thread orientation="h" /></div>

  {#if sync.error}
    <div class="err-state" role="alert" aria-live="polite">
      <div class="err-row">
        <Pulse phase="error" size={8} />
        <span class="err-head">We couldn't reach the daemon.</span>
      </div>
      <p class="err-sub">{sync.error} Try again — sync state lives in the daemon.</p>
      <div class="err-actions">
        <button class="retry" onclick={() => void sync.refresh()}>Try again</button>
      </div>
      <div class="err-hair"></div>
    </div>
  {/if}

  <div class="canvas" role="presentation">
    <!-- thread layer: the sync links, breathing at 9s -->
    <svg class="thread-layer" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
      {#each nodes as nd (nd.pair.device_id)}
        <path
          class="sync-glow"
          class:frayed={frayedIds.includes(nd.pair.device_id)}
          d={`M ${THIS.x} ${THIS.y} L ${nd.x} ${nd.y}`}
          pathLength="1"
          vector-effect="non-scaling-stroke"
        />
        <path
          class="sync-link"
          class:frayed={frayedIds.includes(nd.pair.device_id)}
          d={`M ${THIS.x} ${THIS.y} L ${nd.x} ${nd.y}`}
          pathLength="1"
          vector-effect="non-scaling-stroke"
          style={`--breathe-delay:${(1.1 + nd.index * 0.6).toFixed(2)}s`}
        />
      {/each}
    </svg>

    <!-- this device: the green node, the agent's presence -->
    <div class="this-node" style={`left:${THIS.x}%;top:${THIS.y}%`}>
      <div class="node-dot this-dot">
        <div class="this-face">
          {#if qrDataUrl}
            <img class="this-qr" src={qrDataUrl} alt="This device's identity code" />
          {:else}
            <Glyph name="sync" size={26} stroke={1.5} />
          {/if}
        </div>
        <span class="this-halo"></span>
      </div>
      <div class="node-label">
        <span class="node-name">{status?.name || 'this device'}</span>
        <span class="node-sub mono">{status?.device_id ? status.device_id.slice(0, 12) + '…' : '—'}</span>
      </div>
    </div>

    <!-- paired devices: pollen nodes -->
    {#each nodes as nd (nd.pair.device_id)}
      <button
        class="pair-node"
        style={`left:${nd.x}%;top:${nd.y}%;--spring-delay:${(nd.index * 0.08).toFixed(2)}s`}
        class:revoking={frayedIds.includes(nd.pair.device_id)}
        class:selected={selectedPair?.device_id === nd.pair.device_id}
        onclick={() =>
          (selectedPair = selectedPair?.device_id === nd.pair.device_id ? null : nd.pair)}
        aria-label={`Paired device ${nd.pair.device_name}. Click to review or revoke.`}
      >
        <span class="node-dot pollen-dot"><span class="pollen-halo"></span></span>
        <span class="node-label">
          <span class="node-name">{nd.pair.device_name}</span>
          <span class="node-sub mono">paired {fmtSeen(nd.pair.paired_at)}</span>
        </span>
      </button>
    {/each}

    <!-- revoke popover (centred over the canvas, with a soft backdrop) -->
    {#if selectedPair}
      <div
        class="pop-backdrop"
        onclick={() => (selectedPair = null)}
        onkeydown={(e) => e.key === 'Escape' && (selectedPair = null)}
        role="presentation"
      ></div>
      <div class="revoke-pop" role="dialog" aria-label="Revoke paired device">
        <div class="pop-eyebrow">Paired device</div>
        <div class="pop-name">{selectedPair.device_name}</div>
        <div class="pop-id mono">{selectedPair.device_id}</div>
        <p class="pop-warn">
          Revoking severs the thread. The other side loses access on its next sync.
        </p>
        <div class="pop-actions">
          <Button variant="ghost" size="sm" onclick={() => (selectedPair = null)}>Keep</Button>
          <Button
            variant="danger"
            size="sm"
            disabled={sync.loading}
            onclick={() => {
              const p = selectedPair;
              if (p) doRevoke(p.device_id);
            }}
          >
            <Glyph name="trash" size={13} /> Revoke
          </Button>
        </div>
      </div>
    {/if}

    <!-- pending pairing card (centred) -->
    {#if pending}
      <div class="pending-card" role="dialog" aria-label="Pairing pending">
        <div class="p-eyebrow">
          Pairing with <span class="p-peer">{pendingPeerName}</span>
        </div>

        <div class="p-ring-row">
          <div class="ring-wrap">
            <svg class="ring" viewBox="0 0 36 36" aria-hidden="true">
              <circle class="ring-track" cx="18" cy="18" r="16" />
              <circle
                class="ring-fill"
                cx="18"
                cy="18"
                r="16"
                pathLength="1"
                stroke-dasharray="1"
                stroke-dashoffset={1 - ringPct / 100}
              />
            </svg>
            <div class="ring-centre">
              <div class="ring-pin">{sync.pendingPin}</div>
              <div class="ring-ttl" class:expired>{ringLabel}</div>
            </div>
          </div>

          <div class="p-pulse">
            <Pulse phase="awaiting" size={12} />
            <span class="p-waiting">{expired ? 'token expired' : 'waiting for the peer…'}</span>
          </div>
        </div>

        <p class="p-hint">
          Read the PIN on the peer device, then type it here to seal the link.
        </p>

        <div class="p-confirm">
          <input
            class="pin-input"
            type="text"
            inputmode="numeric"
            bind:value={pinInput}
            placeholder="000000"
            maxlength="8"
            disabled={expired}
            onkeydown={(e) => {
              if (e.key === 'Enter' && pinReady) confirmPair();
            }}
            aria-label="Pairing PIN"
          />
          <Button
            variant="primary"
            disabled={!pinReady || sync.loading || expired}
            onclick={confirmPair}
          >
            {sync.loading ? 'Sealing…' : 'Seal link'}
          </Button>
        </div>

        {#if sync.error}
          <p class="p-err">{sync.error}</p>
        {/if}

        <button class="p-cancel" onclick={cancelPending}>Cancel pairing</button>
      </div>
    {/if}

    <!-- empty-state whisper -->
    {#if sync.pairs.length === 0 && !pending && !sync.loading}
      <div class="canvas-whisper">
        <Pulse phase="thinking" size={10} />
        <span>Discovering peers on your LAN…</span>
      </div>
    {/if}
  </div>

  <!-- peers rail: discoverable on the LAN, each with a Pair affordance -->
  <section class="peers" aria-label="Discovered peers">
    <header class="peers-head">
      <div class="eyebrow">Discovered on your LAN</div>
      <button class="refresh-btn" onclick={() => void refreshAll()} disabled={sync.loading}>
        <Glyph name="sync" size={13} />
        <span>{sync.loading ? 'Refreshing…' : 'Refresh'}</span>
      </button>
    </header>

    {#if sync.peers.length === 0}
      <div class="peers-empty">
        <span class="peers-head">No peers visible yet.</span>
        <span class="peers-sub">Open Condura on another machine on the same network — both devices must be on the LAN for mDNS discovery to find each other.</span>
      </div>
    {:else}
      <ul class="peer-list">
        {#each sync.peers as p, i (p.device_id)}
          <li class="peer-chip" style={`--stagger:${(i * 0.04).toFixed(2)}s`}>
            <div class="chip-id">
              <span class="chip-name">{p.name}</span>
              <span class="chip-meta mono">
                {p.fingerprint ? p.fingerprint.slice(0, 12) : p.address}
              </span>
            </div>
            <Button
              variant="primary"
              size="sm"
              disabled={sync.loading || pending}
              onclick={() => startPair(p.device_id)}
            >
              Pair
            </Button>
          </li>
        {/each}
      </ul>
    {/if}
  </section>

  <p class="foot">No central server — just threads between your machines.</p>
</section>

<style>
  .sync {
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    max-width: 64rem;
    margin: 0 auto;
    padding: var(--space-6) var(--space-5) var(--space-5);
    overflow-y: auto;
  }

  /* ── header ── */
  .s-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-4);
  }
  .s-head-text {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .s-head-badge {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: 7px 12px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-pill);
    background: var(--surface-card);
    flex: none;
  }
  .badge-text {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.08em;
    color: var(--content-mute);
    white-space: nowrap;
  }

  .rule {
    height: 2px;
  }

  /* error state */
  .err-state {
    max-width: 560px;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .err-row {
    display: inline-flex;
    align-items: center;
    gap: 10px;
  }
  .err-head {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 22px;
    line-height: 1.15;
    color: var(--content);
    letter-spacing: -0.01em;
  }
  .err-sub {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 15px;
    line-height: 1.55;
    color: var(--content-faint);
    max-width: 48ch;
  }
  .err-actions {
    margin-top: var(--space-2);
  }
  .retry {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--synapse);
    background: none;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-pill);
    padding: 6px 14px;
    cursor: pointer;
    transition: color var(--dur) var(--ease), border-color var(--dur) var(--ease);
  }
  .retry:hover {
    color: var(--content);
    border-color: var(--synapse);
  }
  .err-hair {
    height: 1px;
    width: 100%;
    background: linear-gradient(90deg, var(--hair-strong) 0%, var(--hair-strong) 60%, transparent 100%);
    transform: scaleX(0);
    transform-origin: left;
    animation: err-hair-draw 600ms var(--ease) 120ms forwards;
  }
  @keyframes err-hair-draw {
    to { transform: scaleX(1); }
  }
  @media (prefers-reduced-motion: reduce) {
    .err-hair {
      transform: scaleX(1);
      animation: none;
    }
  }

  /* ── canvas (the node garden) ── */
  .canvas {
    position: relative;
    flex: 1 1 auto;
    min-height: 420px;
    border-radius: var(--r-lg);
    border: 1px solid var(--hair);
    overflow: visible;
    isolation: isolate;
  }
  .canvas::before {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: inherit;
    background-color: var(--surface);
    background-image:
      radial-gradient(ellipse at 18% -5%, var(--bloom-1), transparent 50%),
      radial-gradient(ellipse at 92% 8%, var(--bloom-2), transparent 45%),
      radial-gradient(ellipse at 50% 105%, var(--bloom-3), transparent 55%);
    z-index: -1;
  }

  .thread-layer {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    z-index: 0;
  }
  .sync-link {
    fill: none;
    stroke: var(--synapse);
    stroke-width: var(--thread-w);
    stroke-linecap: round;
    stroke-dasharray: 1;
    stroke-dashoffset: 1;
    animation:
      sync-draw 1.1s var(--ease) forwards,
      sync-breathe 9s ease-in-out var(--breathe-delay, 1.1s) infinite;
  }
  .sync-glow {
    fill: none;
    stroke: var(--synapse-glow);
    stroke-width: 3;
    opacity: 0.16;
    filter: blur(3px);
    stroke-linecap: round;
    stroke-dasharray: 1;
    stroke-dashoffset: 1;
    animation:
      sync-draw 1.1s var(--ease) forwards,
      sync-glow-breathe 9s ease-in-out 1.1s infinite;
  }
  .sync-link.frayed,
  .sync-glow.frayed {
    animation: sync-fray 0.9s var(--ease) forwards;
    stroke-dasharray: 0.18 0.16;
  }
  @keyframes sync-draw {
    to {
      stroke-dashoffset: 0;
    }
  }
  @keyframes sync-breathe {
    0%,
    100% {
      opacity: 0.55;
    }
    50% {
      opacity: 1;
    }
  }
  @keyframes sync-glow-breathe {
    0%,
    100% {
      opacity: 0.12;
    }
    50% {
      opacity: 0.34;
    }
  }
  @keyframes sync-fray {
    to {
      opacity: 0;
      stroke-dashoffset: 1;
    }
  }

  /* ── nodes ── */
  .this-node,
  .pair-node {
    position: absolute;
    transform: translate(-50%, -50%);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-2);
    z-index: 2;
    background: none;
    border: none;
    padding: 0;
    color: inherit;
    text-align: center;
  }
  .pair-node {
    cursor: pointer;
    animation: node-spring 0.7s var(--ease) both;
    animation-delay: var(--spring-delay, 0s);
  }
  .pair-node:focus-visible {
    outline: none;
  }
  .pair-node:focus-visible .node-dot {
    outline: 2px solid var(--synapse);
    outline-offset: 4px;
    border-radius: 50%;
  }
  .pair-node.revoking {
    animation: node-fade 0.9s var(--ease) forwards;
    pointer-events: none;
  }
  @keyframes node-spring {
    from {
      opacity: 0;
      transform: translate(-50%, -50%) scale(0.3);
    }
    60% {
      opacity: 1;
      transform: translate(-50%, -50%) scale(1.12);
    }
    to {
      opacity: 1;
      transform: translate(-50%, -50%) scale(1);
    }
  }
  @keyframes node-fade {
    to {
      opacity: 0;
      transform: translate(-50%, -50%) scale(0.5);
    }
  }

  .node-dot {
    position: relative;
    display: grid;
    place-items: center;
    border-radius: 50%;
    flex: none;
  }
  .this-dot {
    width: 96px;
    height: 96px;
    background: radial-gradient(circle at 35% 30%, var(--synapse-glow), var(--synapse-deep) 72%);
    box-shadow:
      0 18px 44px -16px color-mix(in oklab, var(--synapse) 55%, transparent),
      inset 0 0 0 3px color-mix(in oklab, var(--paper) 22%, transparent);
  }
  .this-face {
    width: 72px;
    height: 72px;
    border-radius: 14px;
    background: var(--paper);
    padding: 5px;
    overflow: hidden;
    display: grid;
    place-items: center;
    color: var(--synapse-deep);
  }
  .this-qr {
    width: 100%;
    height: 100%;
    object-fit: contain;
    display: block;
  }
  .this-halo {
    position: absolute;
    inset: -10px;
    border-radius: 50%;
    border: 1px solid var(--synapse-glow);
    opacity: 0.4;
    animation: sync-glow-breathe 9s ease-in-out infinite;
    pointer-events: none;
  }

  .pollen-dot {
    width: 26px;
    height: 26px;
    background: radial-gradient(circle at 35% 30%, var(--pollen-light), var(--pollen) 70%);
    box-shadow: var(--pollen-halo);
    transition: transform var(--dur) var(--ease);
  }
  .pair-node:hover .pollen-dot {
    transform: scale(1.18);
  }
  .pair-node.selected .pollen-dot {
    box-shadow:
      0 0 0 4px color-mix(in oklab, var(--pollen) 28%, transparent),
      var(--pollen-halo);
  }
  .pollen-halo {
    position: absolute;
    inset: -8px;
    border-radius: 50%;
    border: 1px solid var(--pollen);
    opacity: 0.3;
    pointer-events: none;
  }

  .node-label {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1px;
    max-width: 140px;
  }
  .node-name {
    font-size: 13px;
    font-weight: 500;
    color: var(--content);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 100%;
  }
  .node-sub {
    font-size: 10px;
    color: var(--content-mute);
    letter-spacing: 0.04em;
  }

  /* ── revoke popover ── */
  .pop-backdrop {
    position: absolute;
    inset: 0;
    z-index: 4;
    cursor: default;
  }
  .revoke-pop {
    position: absolute;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    z-index: 5;
    width: min(320px, 86%);
    padding: var(--space-5);
    background: var(--surface);
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-md);
    box-shadow: var(--shadow-float);
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    animation: blur-in var(--dur) var(--ease);
  }
  .pop-eyebrow {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--danger);
  }
  .pop-name {
    font-family: var(--font-display);
    font-size: 24px;
    line-height: 1.1;
    letter-spacing: -0.02em;
    color: var(--content);
  }
  .pop-id {
    font-size: 11px;
    color: var(--content-mute);
    word-break: break-all;
  }
  .pop-warn {
    font-size: 13px;
    color: var(--content-mute);
    margin: var(--space-1) 0 var(--space-2);
  }
  .pop-actions {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
  }

  /* ── pending pairing card ── */
  .pending-card {
    position: absolute;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    z-index: 6;
    width: min(420px, 92%);
    padding: var(--space-6);
    background: var(--surface);
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-lg);
    box-shadow: var(--shadow-float);
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    animation: blur-in var(--dur-slow) var(--ease);
  }
  .p-eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--content-mute);
  }
  .p-peer {
    color: var(--pollen);
    text-transform: none;
    letter-spacing: 0;
  }
  .p-ring-row {
    display: flex;
    align-items: center;
    gap: var(--space-5);
  }
  .ring-wrap {
    position: relative;
    width: 132px;
    height: 132px;
    flex: none;
  }
  .ring {
    width: 100%;
    height: 100%;
    transform: rotate(-90deg);
  }
  .ring-track {
    fill: none;
    stroke: var(--hair);
    stroke-width: 2;
  }
  .ring-fill {
    fill: none;
    stroke: var(--pollen);
    stroke-width: 2.5;
    stroke-linecap: round;
    transition: stroke-dashoffset 0.25s linear;
    filter: drop-shadow(0 0 4px color-mix(in oklab, var(--pollen) 35%, transparent));
  }
  .ring-centre {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 2px;
  }
  .ring-pin {
    font-family: var(--font-mono);
    font-size: 26px;
    font-weight: 600;
    letter-spacing: 0.18em;
    color: var(--content);
    text-shadow: 0 0 18px color-mix(in oklab, var(--pollen) 25%, transparent);
  }
  .ring-ttl {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    color: var(--content-mute);
  }
  .ring-ttl.expired {
    color: var(--danger);
  }
  .p-pulse {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-2);
  }
  .p-waiting {
    font-size: 13px;
    color: var(--content-soft);
  }
  .p-hint {
    font-size: 13px;
    color: var(--content-mute);
    line-height: 1.5;
  }
  .p-confirm {
    display: flex;
    gap: var(--space-2);
    align-items: stretch;
  }
  .pin-input {
    flex: 1;
    min-width: 0;
    background: var(--surface-card);
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-sm);
    color: var(--content);
    font-family: var(--font-mono);
    font-size: 18px;
    text-align: center;
    letter-spacing: 0.18em;
    padding: 0 var(--space-3);
    height: 40px;
    transition:
      border-color var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease),
      background var(--dur) var(--ease);
  }
  .pin-input:hover:not(:disabled) {
    background: var(--paper-2);
    border-color: var(--pollen-deep);
  }
  .pin-input::placeholder {
    color: var(--content-faint);
  }
  .pin-input:focus {
    outline: none;
    border-color: var(--pollen);
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .pin-input:disabled {
    opacity: 0.42;
    cursor: not-allowed;
    filter: saturate(0.55);
  }
  .p-err {
    font-size: 12px;
    color: var(--danger);
  }
  .p-cancel {
    align-self: center;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-mute);
    padding: var(--space-1) var(--space-3);
    border-radius: var(--r-sm);
    transition:
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease);
  }
  .p-cancel:hover {
    color: var(--content);
    background: color-mix(in oklab, var(--pollen) 8%, transparent);
  }
  .p-cancel:active {
    transform: scale(0.97);
  }
  .p-cancel:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }

  /* ── empty whisper ── */
  .canvas-whisper {
    position: absolute;
    left: 50%;
    bottom: var(--space-4);
    transform: translateX(-50%);
    z-index: 1;
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: 6px 12px;
    border-radius: var(--r-pill);
    background: var(--surface-card);
    border: 1px solid var(--hair);
    font-size: 12px;
    color: var(--content-mute);
  }

  /* ── peers rail ── */
  .peers {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  .peers-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .refresh-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.08em;
    color: var(--content-mute);
    padding: 5px 10px;
    border-radius: var(--r-pill);
    border: 1px solid var(--hair);
    transition:
      color var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .refresh-btn:hover:not(:disabled) {
    color: var(--content);
    border-color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 6%, transparent);
    transform: translateY(-1px);
  }
  .refresh-btn:active:not(:disabled) {
    transform: scale(0.97);
  }
  .refresh-btn:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .refresh-btn:disabled {
    opacity: 0.42;
    cursor: not-allowed;
    filter: saturate(0.55);
  }
  .peers-empty {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-5);
    border: 1px dashed var(--hair-strong);
    border-radius: var(--r-md);
  }
  .peers-head {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 20px;
    line-height: 1.15;
    color: var(--content);
    letter-spacing: -0.01em;
  }
  .peers-sub {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 14px;
    line-height: 1.55;
    color: var(--content-faint);
    max-width: 52ch;
  }
  .peer-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .peer-chip {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    background: var(--surface-card);
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    animation: fade-in-up var(--dur) var(--ease) both;
    animation-delay: var(--stagger, 0s);
    transition: border-color var(--dur) var(--ease);
  }
  .peer-chip:hover {
    border-color: var(--hair-strong);
  }
  .chip-id {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }
  .chip-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--content);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .chip-meta {
    font-size: 10px;
    color: var(--content-mute);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .foot {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 15px;
    color: var(--content-mute);
    text-align: center;
    margin-top: var(--space-2);
  }

  /* ── responsive ── */
  @media (max-width: 640px) {
    .sync {
      padding: var(--space-5) var(--space-4);
    }
    .s-head {
      flex-direction: column;
      align-items: flex-start;
    }
    .canvas {
      min-height: 340px;
    }
    .this-dot {
      width: 78px;
      height: 78px;
    }
    .this-face {
      width: 58px;
      height: 58px;
    }
    .ring-wrap {
      width: 116px;
      height: 116px;
    }
    .p-ring-row {
      flex-direction: column;
      align-items: flex-start;
      gap: var(--space-3);
    }
  }

  /* reduced motion: the global rule neutralises durations; we just ensure
     the resting states are legible (threads visible, nodes opaque). */
  @media (prefers-reduced-motion: reduce) {
    .sync-link,
    .sync-glow {
      stroke-dashoffset: 0;
      opacity: 0.7;
      animation: none;
    }
    .pair-node,
    .this-halo {
      animation: none;
    }
  }
</style>