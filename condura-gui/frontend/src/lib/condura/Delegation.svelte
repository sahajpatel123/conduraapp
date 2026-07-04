<script lang="ts">
  import { onMount } from 'svelte';
  import Glyph from './Glyph.svelte';
  import Pulse from './Pulse.svelte';
  import Thread from './Thread.svelte';
  import {
    pendingActions,
    pendingCount,
    refreshPendingActions,
    approvePending,
    denyPending,
    executePending,
    startPolling,
    stopPolling,
    type PendingAction,
    type PendingStatus,
  } from '../stores/pending.svelte';

  // Condura Delegation — the constellation. A breathing Condura core at the
  // center, sub-agent CLIs orbiting it, each connected by a self-drawing
  // thread. To the right (or below on narrow screens) is the "Live" panel:
  // every pending sub-agent action that has surfaced from the Gatekeeper.
  //
  // Honest empty-state: if no sub-agents are wired, the canvas shows a single
  // Glyph instead of faking empty cards. The Live panel is always live
  // (it's driven by `pendingActions`), but its initial population depends on
  // the daemon — we poll every 5s.
  //
  // a11y:
  // - the canvas itself is a role="img" with an aria-label describing what's
  //   drawn. Every interactive surface is a real <button> with aria-label.
  // - The Live panel uses <button> for every action and exposes the status
  //   text in the button's label so screen-readers can announce it.
  // - honors prefers-reduced-motion (no breath scale, no hover scale).

  type AgentKey = 'claude_code' | 'codex' | 'antigravity' | 'opencode' | 'kilo' | 'hermes' | 'ollama';

  interface SubAgent {
    key: AgentKey;
    name: string;
    blurb: string;
  }

  // The list is intentionally static — we surface what the system can
  // orchestrate and let the Live panel tell the user what's *actually* in
  // flight via the daemon's pending-actions list.
  const SUB_AGENTS: ReadonlyArray<SubAgent> = [
    { key: 'claude_code', name: 'Claude Code', blurb: 'Codegen, refactors, reviews.' },
    { key: 'codex', name: 'Codex', blurb: 'OpenAI coding agent.' },
    { key: 'antigravity', name: 'Antigravity', blurb: 'Background-first GUI control.' },
    { key: 'opencode', name: 'OpenCode', blurb: 'Open-source coding harness.' },
    { key: 'kilo', name: 'Kilo Code', blurb: 'Fast multi-file edits.' },
    { key: 'hermes', name: 'Hermes', blurb: 'Reasoning + tool use.' },
    { key: 'ollama', name: 'Ollama', blurb: 'Local models, no cloud.' },
  ];

  // Radius (px) of the orbit ring.
  const ORBIT_RADIUS = 170;
  // Size (px) of the canvas box (must be at least 2*R + node size for breathing room).
  const CANVAS_SIZE = 440;
  // Anchor for the very first node — top-of-circle.
  const ANGLE_START = -Math.PI / 2;

  let hovered = $state<AgentKey | null>(null);
  let pulseHover = $state(false);

  // Read pending actions reactively via $-store auto-subscription.
  let actions = $state<PendingAction[]>([]);
  let openCount = $state(0);
  pendingActions.subscribe((rows) => {
    actions = rows;
    openCount = rows.filter((r) => r.status === 'pending').length;
  });

  // Avoid an "unused" diagnostic on pendingCount (kept exported by the store so
  // other surfaces, like the tray, can use it directly).
  void pendingCount;

  function nodeAngle(i: number, n: number): number {
    if (n <= 0) return ANGLE_START;
    return ANGLE_START + (i / n) * Math.PI * 2;
  }

  function nodePosition(i: number, n: number): { x: number; y: number } {
    const a = nodeAngle(i, n);
    const cx = CANVAS_SIZE / 2;
    const cy = CANVAS_SIZE / 2;
    return { x: cx + Math.cos(a) * ORBIT_RADIUS, y: cy + Math.sin(a) * ORBIT_RADIUS };
  }

  function nodeLabel(a: SubAgent): string {
    return `Sub-agent: ${a.name} · Ready. Run: opens delegate stream. Cancel: stops the current delegation.`;
  }

  function setHover(key: AgentKey | null): void {
    hovered = key;
  }

  function statusToTone(status: PendingStatus): string {
    switch (status) {
      case 'pending':
        return 'pending';
      case 'approved':
        return 'approved';
      case 'denied':
        return 'denied';
      case 'executed':
        return 'executed';
      case 'failed':
      case 'expired':
      case 'superseded':
        return 'failed';
      default:
        return 'pending';
    }
  }

  function actionLabel(a: PendingAction): string {
    const kind = a.kind ? a.kind : 'action';
    const who = a.agent_name ? ` · ${a.agent_name}` : '';
    return `${kind}${who} — ${statusToTone(a.status)}`;
  }

  function canApprove(a: PendingAction): boolean {
    return a.status === 'pending';
  }
  function canDeny(a: PendingAction): boolean {
    return a.status === 'pending';
  }
  function canRun(a: PendingAction): boolean {
    return a.status === 'approved';
  }

  async function onApprove(a: PendingAction): Promise<void> {
    await approvePending(a.id, '', true);
  }
  async function onDeny(a: PendingAction): Promise<void> {
    await denyPending(a.id, '');
  }
  async function onRun(a: PendingAction): Promise<void> {
    await executePending(a.id);
  }

  async function refreshNow(): Promise<void> {
    await refreshPendingActions();
  }

  onMount(() => {
    // 5s cadence per the brief. The store is idempotent — multiple
    // surfaces polling won't multiply the network.
    startPolling(5000);
    return () => {
      stopPolling();
    };
  });

  // Build the aria-label for the canvas image.
  function buildCanvasLabel(): string {
    const names = SUB_AGENTS.map((s) => s.name).join(', ');
    return `Delegation constellation: Condura core surrounded by ${SUB_AGENTS.length} sub-agents — ${names}.`;
  }
</script>

<div class="delegation">
  <header class="head">
    <div class="eyebrow">— The conductor</div>
    <h1 class="title">Delegation.</h1>
    <p class="sub">
      Condura at the center, sub-agents on the ring. Each thread is a delegation
      channel; the Live panel shows what's actually in flight.
    </p>
  </header>

  <div class="surface">
    <!-- Canvas (the constellation) -->
    <section
      class="canvas"
      role="img"
      aria-label={buildCanvasLabel()}
    >
      <div class="orbit-ring" aria-hidden="true"></div>

      <!-- Threads first so the node cards sit on top. -->
      {#each SUB_AGENTS as agent, i (agent.key)}
        {@const p = nodePosition(i, SUB_AGENTS.length)}
        {@const cx = CANVAS_SIZE / 2}
        {@const cy = CANVAS_SIZE / 2}
        {@const left = Math.min(p.x, cx)}
        {@const top = Math.min(p.y, cy)}
        {@const width = Math.abs(p.x - cx)}
        {@const height = Math.abs(p.y - cy)}
        {@const vx = p.x >= cx ? 'h' : 'h'}
        {@const vy = p.y >= cy ? 'v' : 'v'}
        <!-- Diagonal thread: draw two perpendiculars (one H, one V) through (cx,cy)→(px,py). -->
        <span
          class="leg leg-h"
          class:hover={hovered === agent.key}
          style:left="{left}px"
          style:top="{cy - 1}px"
          style:width="{width}px"
        >
          <Thread orientation={vx} draw={true} />
        </span>
        <span
          class="leg leg-v"
          class:hover={hovered === agent.key}
          style:left="{cx - 1}px"
          style:top="{top}px"
          style:height="{height}px"
        >
          <Thread orientation={vy} draw={true} />
        </span>
      {/each}

      <!-- Core (the breathing Condura heart) -->
      <button
        class="core"
        type="button"
        aria-label="Condura core. Pause to focus the constellation."
        onmouseenter={() => (pulseHover = true)}
        onmouseleave={() => (pulseHover = false)}
        onfocus={() => (pulseHover = true)}
        onblur={() => (pulseHover = false)}
      >
        <span class="halo" aria-hidden="true">
          <Pulse phase="acting" size={20} />
        </span>
      </button>

      <!-- Nodes -->
      {#each SUB_AGENTS as agent, i (agent.key)}
        {@const p = nodePosition(i, SUB_AGENTS.length)}
        <button
          class="node"
          type="button"
          aria-label={nodeLabel(agent)}
          onmouseenter={() => setHover(agent.key)}
          onmouseleave={() => setHover(null)}
          onfocus={() => setHover(agent.key)}
          onblur={() => setHover(null)}
          style:left="{p.x}px"
          style:top="{p.y}px"
          style:transform="translate(-50%, -50%) scale({hovered === agent.key ? 1.06 : 1})"
          data-hovered={hovered === agent.key ? 'true' : 'false'}
        >
          <span class="node-name">{agent.name}</span>
          <span class="node-status">
            <span class="node-dot" aria-hidden="true"></span>
            ready
          </span>
        </button>

        {#if hovered === agent.key}
          {@const p2 = nodePosition(i, SUB_AGENTS.length)}
          <span
            class="popover"
            role="tooltip"
            aria-label={nodeLabel(agent)}
            style:left="{p2.x}px"
            style:top="{p2.y + 38}px"
          >
            <span class="pop-arrow" aria-hidden="true"></span>
            <span class="pop-name">{agent.name}</span>
            <span class="pop-blurb">{agent.blurb}</span>
            <span class="pop-foot">Run opens delegate stream · Cancel stops the current delegation</span>
          </span>
        {/if}
      {/each}

      {#if SUB_AGENTS.length === 0}
        <div class="empty">
          <Glyph name="delegation" size={56} />
          <div class="empty-eyebrow">— No sub-agents wired</div>
          <div class="empty-body">
            Install Claude Code, Codex, or Ollama to populate the constellation.
          </div>
        </div>
      {/if}
    </section>

    <!-- Live panel -->
    <aside class="live" aria-label="Live sub-agent activity">
      <div class="live-head">
        <div class="live-eyebrow">— Live</div>
        <div class="live-meta">
          <span class="live-count" aria-live="polite">
            {openCount} pending
          </span>
          <button
            type="button"
            class="refresh"
            aria-label="Refresh pending actions"
            onclick={() => void refreshNow()}
          >
            Refresh
          </button>
        </div>
      </div>

      {#if actions.length === 0}
        <div class="live-empty">
          <span class="live-empty-h">Nothing in flight.</span>
          <span class="live-empty-s">
            When a sub-agent asks the Gatekeeper to act, the ticket will land here.
          </span>
        </div>
      {:else}
        <ul class="rows" role="list">
          {#each actions as a (a.id)}
            <li class="row" data-status={a.status}>
              <div class="row-body">
                <div class="row-label">{actionLabel(a)}</div>
                <div class="row-meta mono">
                  {a.agent_name ?? 'sub-agent'} · {a.kind ?? ''}{a.gate_decision ? ` · ${a.gate_decision}` : ''}
                </div>
              </div>
              <div class="row-actions">
                <span class="badge badge-{statusToTone(a.status)}" aria-label="Gate status {a.status}">
                  {a.status}
                </span>
                <button
                  type="button"
                  class="row-btn"
                  aria-label="Approve {a.kind ?? 'pending action'} from {a.agent_name ?? 'sub-agent'}"
                  disabled={!canApprove(a)}
                  onclick={() => void onApprove(a)}
                >
                  Approve
                </button>
                <button
                  type="button"
                  class="row-btn"
                  aria-label="Deny {a.kind ?? 'pending action'} from {a.agent_name ?? 'sub-agent'}"
                  disabled={!canDeny(a)}
                  onclick={() => void onDeny(a)}
                >
                  Deny
                </button>
                <button
                  type="button"
                  class="row-btn primary"
                  aria-label="Run approved action from {a.agent_name ?? 'sub-agent'}"
                  disabled={!canRun(a)}
                  onclick={() => void onRun(a)}
                >
                  Run
                </button>
              </div>
            </li>
          {/each}
        </ul>
      {/if}
    </aside>
  </div>
</div>

<style>
  .delegation {
    max-width: 1100px;
    padding-top: var(--space-7);
  }

  .head {
    margin-bottom: var(--space-6);
  }
  .eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.22em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .title {
    font-family: var(--font-display);
    font-size: clamp(28px, 3vw, 40px);
    line-height: 1.08;
    letter-spacing: -0.03em;
    color: var(--content);
    margin: var(--space-3) 0 var(--space-2);
  }
  .sub {
    font-size: 16px;
    line-height: 1.55;
    color: var(--content-soft);
    max-width: 56ch;
  }

  .surface {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(280px, 360px);
    gap: var(--space-6);
    align-items: start;
  }

  /* ── Canvas ───────────────────────────────────────────────────── */
  .canvas {
    position: relative;
    width: 440px;
    max-width: 100%;
    height: 440px;
    margin: 0 auto;
    border-radius: var(--r-lg);
    background:
      radial-gradient(ellipse at center, color-mix(in oklab, var(--synapse) 4%, transparent), transparent 60%),
      var(--surface-card);
    border: 1px solid var(--hair);
    box-shadow: var(--shadow-paper);
    overflow: visible;
  }
  :global([data-mode='dark']) .canvas {
    background:
      radial-gradient(ellipse at center, color-mix(in oklab, var(--synapse-glow) 8%, transparent), transparent 60%),
      var(--surface-card);
  }

  .orbit-ring {
    position: absolute;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    width: 340px;
    height: 340px;
    border-radius: 50%;
    border: 1px dashed var(--hair-strong);
    pointer-events: none;
  }

  /* Thread legs (diagonal split into H + V through the core center) */
  .leg {
    position: absolute;
    pointer-events: none;
    color: var(--synapse);
    transition: filter var(--dur) var(--ease), opacity var(--dur) var(--ease);
    opacity: 0.85;
  }
  .leg-h {
    height: 2px;
  }
  .leg-v {
    width: 2px;
  }
  .leg :global(.condura-thread .line) {
    stroke: var(--synapse);
    transition: stroke var(--dur) var(--ease);
  }
  .leg :global(.condura-thread .glow) {
    stroke: var(--synapse-glow);
  }
  .leg.hover :global(.condura-thread .line) {
    stroke: var(--pollen);
  }
  .leg.hover :global(.condura-thread .glow) {
    stroke: var(--pollen-light);
  }

  /* Core */
  .core {
    position: absolute;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    width: 96px;
    height: 96px;
    border-radius: 50%;
    border: none;
    background: transparent;
    display: grid;
    place-items: center;
    cursor: pointer;
    z-index: 2;
    padding: 0;
  }
  .halo {
    width: 96px;
    height: 96px;
    border-radius: 50%;
    background: radial-gradient(
      circle,
      color-mix(in oklab, var(--synapse) 18%, transparent) 0%,
      color-mix(in oklab, var(--synapse) 8%, transparent) 55%,
      transparent 80%
    );
    box-shadow: 0 0 48px color-mix(in oklab, var(--synapse) 40%, transparent);
    display: grid;
    place-items: center;
    color: var(--synapse);
  }
  :global([data-mode='dark']) .halo {
    background: radial-gradient(
      circle,
      color-mix(in oklab, var(--synapse-glow) 30%, transparent) 0%,
      color-mix(in oklab, var(--synapse-glow) 12%, transparent) 55%,
      transparent 80%
    );
    box-shadow: 0 0 48px color-mix(in oklab, var(--synapse-glow) 45%, transparent);
  }

  /* Sub-agent node cards */
  .node {
    position: absolute;
    display: inline-flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    min-width: 96px;
    padding: 10px 12px;
    border-radius: var(--r-md);
    background: var(--surface-card);
    border: 1px solid var(--hair);
    color: var(--content);
    cursor: pointer;
    z-index: 3;
    box-shadow: var(--shadow-paper);
    transition:
      transform var(--dur-slow) var(--ease),
      border-color var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease),
      background var(--dur) var(--ease);
  }
  .node:hover,
  .node[data-hovered='true'] {
    border-color: var(--synapse);
    box-shadow:
      0 0 0 4px color-mix(in oklab, var(--synapse-glow) 18%, transparent),
      var(--shadow-card);
  }
  .node-name {
    font-family: var(--font-display);
    font-size: 16px;
    line-height: 1.1;
    letter-spacing: -0.01em;
    color: var(--content);
  }
  .node-status {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--content-mute);
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .node-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--synapse-glow);
    box-shadow: 0 0 4px var(--synapse-glow);
    animation: pulse-dot 2.4s var(--ease) infinite;
    flex: none;
  }
  @keyframes pulse-dot {
    0%, 100% { opacity: 0.65; }
    50% { opacity: 1; }
  }

  /* Popover on hover */
  .popover {
    position: absolute;
    z-index: 5;
    transform: translateX(-50%);
    width: 220px;
    padding: 10px 12px;
    border-radius: var(--r-md);
    background: var(--surface);
    border: 1px solid var(--hair-strong);
    box-shadow: var(--shadow-float);
    color: var(--content);
    font-size: 12px;
    line-height: 1.45;
    pointer-events: none;
    animation: pop-in var(--dur-fast) var(--ease) both;
  }
  .pop-arrow {
    position: absolute;
    top: -6px;
    left: 50%;
    transform: translateX(-50%) rotate(45deg);
    width: 12px;
    height: 12px;
    background: var(--surface);
    border-left: 1px solid var(--hair-strong);
    border-top: 1px solid var(--hair-strong);
  }
  .pop-name {
    font-family: var(--font-display);
    font-size: 14px;
    display: block;
    margin-bottom: 2px;
  }
  .pop-blurb {
    color: var(--content-soft);
    display: block;
    margin-bottom: 6px;
  }
  .pop-foot {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--content-mute);
    display: block;
  }
  @keyframes pop-in {
    from { opacity: 0; transform: translate(-50%, -4px); }
    to { opacity: 1; transform: translate(-50%, 0); }
  }

  /* Empty state for the canvas (no sub-agents wired) */
  .empty {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-3);
    text-align: center;
    color: var(--content-mute);
    padding: var(--space-5);
  }
  .empty-eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.22em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .empty-body {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 15px;
    color: var(--content-soft);
    max-width: 28ch;
  }

  /* ── Live panel ───────────────────────────────────────────────── */
  .live {
    border-radius: var(--r-md);
    border: 1px solid var(--hair);
    background: var(--surface-card);
    padding: var(--space-4);
    box-shadow: var(--shadow-paper);
    min-height: 360px;
  }
  .live-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-bottom: var(--space-3);
  }
  .live-eyebrow {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.22em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .live-meta {
    display: inline-flex;
    align-items: center;
    gap: var(--space-3);
  }
  .live-count {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--synapse);
    letter-spacing: 0.08em;
  }
  .refresh {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--content-mute);
    padding: 4px 10px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-pill);
    background: transparent;
    cursor: pointer;
    transition:
      color var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .refresh:hover:not(:disabled) {
    color: var(--content);
    border-color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 6%, transparent);
    transform: translateY(-1px);
  }
  .refresh:active:not(:disabled) {
    transform: scale(0.97);
  }
  .refresh:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }

  .live-empty {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: var(--space-4) 0;
  }
  .live-empty-h {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 18px;
    color: var(--content);
  }
  .live-empty-s {
    color: var(--content-mute);
    font-size: 13px;
  }

  .rows {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
  .row {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: var(--space-3);
    border-radius: var(--r-sm);
    background: var(--surface);
    border: 1px solid var(--hair);
  }
  .row-body {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .row-label {
    font-family: var(--font-display);
    font-size: 15px;
    line-height: 1.2;
    color: var(--content);
  }
  .row-meta {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--content-mute);
  }
  .row-actions {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
  }
  .badge {
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    padding: 3px 8px;
    border-radius: var(--r-pill);
    border: 1px solid var(--hair);
    color: var(--content-mute);
    background: var(--surface-card);
    margin-right: auto;
  }
  .badge-pending { color: var(--warn); border-color: color-mix(in oklab, var(--warn) 40%, transparent); }
  .badge-approved { color: var(--ok); border-color: color-mix(in oklab, var(--ok) 40%, transparent); }
  .badge-denied { color: var(--danger); border-color: color-mix(in oklab, var(--danger) 40%, transparent); }
  .badge-executed { color: var(--synapse); border-color: color-mix(in oklab, var(--synapse) 40%, transparent); }
  .badge-failed { color: var(--danger); border-color: color-mix(in oklab, var(--danger) 40%, transparent); }

  .row-btn {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    padding: 5px 10px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-pill);
    background: var(--surface-card);
    color: var(--content);
    cursor: pointer;
    transition:
      border-color var(--dur) var(--ease),
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .row-btn:hover:not([disabled]) {
    border-color: var(--synapse);
    color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 6%, transparent);
    transform: translateY(-1px);
  }
  .row-btn:active:not([disabled]) {
    transform: scale(0.97);
  }
  .row-btn:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .row-btn[disabled] {
    opacity: 0.42;
    pointer-events: none;
    cursor: not-allowed;
    filter: saturate(0.55);
  }
  .row-btn.primary {
    background: var(--synapse);
    color: var(--paper);
    border-color: var(--synapse);
  }
  :global([data-mode='dark']) .row-btn.primary {
    color: var(--ink);
  }
  .row-btn.primary:hover:not([disabled]) {
    transform: translateY(-1px);
    box-shadow: 0 0 0 4px color-mix(in oklab, var(--synapse-glow) 18%, transparent);
  }
  .row-btn.primary:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }

  /* ── Responsive ───────────────────────────────────────────────── */
  @media (max-width: 820px) {
    .surface {
      grid-template-columns: 1fr;
    }
    .canvas {
      width: 100%;
      height: 380px;
    }
    .orbit-ring {
      width: 300px;
      height: 300px;
    }
    .live {
      min-height: 0;
    }
  }

  /* ── Reduced motion ───────────────────────────────────────────── */
  @media (prefers-reduced-motion: reduce) {
    .node,
    .leg,
    .core,
    .halo,
    .row-btn,
    .node-dot {
      transition: none !important;
      animation: none !important;
    }
    .node:hover {
      transform: translate(-50%, -50%) scale(1) !important;
    }
    .popover {
      animation: none !important;
    }
  }
</style>
