<script lang="ts">
  import { onMount } from 'svelte';
  import { audit } from '../stores/audit.svelte';
  import Thread from './Thread.svelte';
  import Pulse from './Pulse.svelte';

  // Condura Audit — the HMAC-chained event log, visualized as a vertical
  // synapse thread (the chain). Each event is a node; an unbroken green
  // thread means the chain verifies. Hover a node for the full record.

  let actionFilter = $state('');
  let levelFilter = $state<'' | 'info' | 'warn' | 'error'>('');
  let hoverId = $state<number | null>(null);

  const LEVELS: Array<'' | 'info' | 'warn' | 'error'> = ['', 'info', 'warn', 'error'];

  function dotForResult(r: string): { color: string; label: string } {
    if (r === 'allow') return { color: 'var(--ok)', label: 'Allowed' };
    if (r === 'block') return { color: 'var(--danger)', label: 'Blocked' };
    if (r === 'prompt') return { color: 'var(--warn)', label: 'Prompted' };
    return { color: 'var(--content-faint)', label: r };
  }

  function setAction(v: string): void {
    actionFilter = v;
    audit.setFilter(v, levelFilter);
  }
  function setLevel(v: '' | 'info' | 'warn' | 'error'): void {
    levelFilter = v;
    audit.setFilter(actionFilter, v);
  }

  let hover = $derived(audit.events.find((e) => e.id === hoverId) ?? null);

  onMount(() => {
    audit.refresh();
  });
</script>

<div class="audit">
  <header class="head">
    <div class="eyebrow">— Forensics · HMAC-chained · append-only</div>
    <h1 class="title">Every action, on a thread.</h1>
    <p class="sub">
      Each event is a node on the chain. An unbroken green thread means the log verifies —
      when something goes wrong, we can prove exactly what happened.
    </p>
  </header>

  <div class="filters">
    <input
      class="action"
      placeholder="filter by action…"
      value={actionFilter}
      oninput={(e) => setAction((e.currentTarget as HTMLInputElement).value)}
    />
    <div class="levels">
      {#each LEVELS as l (l)}
        <button class="chip" class:active={levelFilter === l} onclick={() => setLevel(l)} data-level={l || 'all'}>
          {l || 'all'}
        </button>
      {/each}
    </div>
  </div>

  <div class="body">
    <div class="chain">
      <div class="thread-spine"><Thread orientation="v" /></div>
      <div class="events">
        {#if audit.loading && audit.events.length === 0}
          <div class="state">
            <Pulse phase="thinking" size={8} />
            <span class="state-label">READING THE CHAIN…</span>
          </div>
        {:else if audit.error}
          <div class="err-state" role="alert" aria-live="polite">
            <div class="err-row">
              <Pulse phase="error" size={8} />
              <span class="err-head">We couldn't read the chain.</span>
            </div>
            <p class="err-sub">{audit.error} The audit log is on disk; the daemon may be offline.</p>
            <div class="err-actions">
              <button class="retry" onclick={() => void audit.refresh()}>Try again</button>
            </div>
            <div class="err-hair"></div>
          </div>
        {:else if audit.events.length === 0}
          <div class="state-empty">
            <span class="empty-head">
              {actionFilter || levelFilter ? 'No events match.' : 'The chain is quiet.'}
            </span>
            <span class="empty-sub">
              {actionFilter || levelFilter
                ? 'Loosen the filters, or look at a wider window.'
                : "Nothing has happened yet. Every action the agent takes will land here, HMAC-chained."}
            </span>
          </div>
        {:else}
          {#each audit.events as e (e.id)}
            <button
              class="node"
              class:hover={hoverId === e.id}
              onclick={() => (hoverId = hoverId === e.id ? null : e.id)}
            >
              <span class="dot" style:background={dotForResult(e.result).color}></span>
              <span class="ts">{e.ts}</span>
              <span class="actor">{e.actor}</span>
              <span class="summary">{e.message}</span>
              <span class="result" style:color={dotForResult(e.result).color}>{e.result}</span>
            </button>
          {/each}
        {/if}
      </div>
    </div>

    <aside class="detail" class:show={!!hover}>
      {#if hover}
        <div class="d-eyebrow">Event #{hover.id}</div>
        <div class="d-ts">{hover.ts}</div>
        <div class="d-row"><span class="d-k">actor</span><span class="d-v">{hover.actor}</span></div>
        <div class="d-row"><span class="d-k">action</span><span class="d-v mono">{hover.action}</span></div>
        <div class="d-row"><span class="d-k">app</span><span class="d-v">{hover.app}</span></div>
        <div class="d-row"><span class="d-k">level</span><span class="d-v" data-level={hover.level}>{hover.level}</span></div>
        <div class="d-row">
          <span class="d-k">result</span>
          <span class="d-v" style:color={dotForResult(hover.result).color}>{dotForResult(hover.result).label}</span>
        </div>
        <div class="d-msg">{hover.message}</div>
      {:else}
        <div class="d-empty">Hover a node to see the full record.</div>
      {/if}
    </aside>
  </div>

  <div class="pages">
    <button class="pg" onclick={() => audit.prevPage()} disabled={audit.offset === 0}>← prev</button>
    <span class="pg-info">{audit.offset}–{audit.offset + audit.events.length} of {audit.total}</span>
    <button class="pg" onclick={() => audit.nextPage()} disabled={audit.events.length < audit.limit}>next →</button>
  </div>
</div>

<style>
  .audit {
    max-width: 980px;
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

  .filters {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin-bottom: var(--space-5);
  }
  .action {
    flex: 1;
    max-width: 320px;
    padding: 9px 14px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-pill);
    background: var(--surface);
    color: var(--content);
    font-family: var(--font-mono);
    font-size: 13px;
    outline: none;
    transition:
      border-color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .action:hover {
    background: var(--surface-card);
    transform: translateY(-1px);
  }
  .action:focus-visible {
    border-color: var(--synapse);
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .action::placeholder {
    color: var(--content-faint);
  }
  .levels {
    display: flex;
    gap: 6px;
  }
  .chip {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    padding: 6px 10px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-pill);
    background: var(--surface-card);
    color: var(--content-mute);
    cursor: pointer;
    transition:
      color var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .chip:hover {
    color: var(--content);
    border-color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 6%, transparent);
    transform: translateY(-1px);
  }
  .chip:active {
    transform: scale(0.97);
  }
  .chip:focus-visible {
    outline: none;
    border-color: var(--synapse);
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .chip.active {
    color: var(--content);
    border-color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 10%, transparent);
  }
  .chip[data-level='error'] {
    border-color: color-mix(in oklab, var(--danger) 30%, transparent);
  }
  .chip[data-level='warn'] {
    border-color: color-mix(in oklab, var(--warn) 30%, transparent);
  }

  .body {
    display: grid;
    grid-template-columns: 1fr 320px;
    gap: var(--space-6);
  }
  .chain {
    position: relative;
    padding-left: 28px;
  }
  .thread-spine {
    position: absolute;
    left: 6px;
    top: 0;
    bottom: 0;
    width: 2px;
  }
  .thread-spine :global(.condura-thread) {
    height: 100%;
  }
  .events {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .state {
    display: flex;
    align-items: center;
    gap: 8px;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
    padding: var(--space-4) 0;
  }
  .state-label {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  /* editorial empty (instrument-serif italic). Used for both the true zero
     state and the partial / filtered-zero state — copy differentiates. */
  .state-empty {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-2);
    padding: var(--space-5) 0 var(--space-6);
  }
  .state-empty .empty-head {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 22px;
    line-height: 1.15;
    color: var(--content);
    letter-spacing: -0.01em;
  }
  .state-empty .empty-sub {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 15px;
    line-height: 1.55;
    color: var(--content-faint);
    max-width: 48ch;
  }

  /* error state */
  .err-state {
    max-width: 520px;
    padding: var(--space-4) 0;
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
    transition:
      color var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .retry:hover {
    color: var(--content);
    border-color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 6%, transparent);
    transform: translateY(-1px);
  }
  .retry:active {
    transform: scale(0.97);
  }
  .retry:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
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
  .node {
    display: grid;
    grid-template-columns: 10px auto auto 1fr auto;
    align-items: center;
    gap: var(--space-3);
    padding: 9px var(--space-3);
    border: 1px solid transparent;
    border-radius: var(--r-sm);
    background: transparent;
    text-align: left;
    cursor: pointer;
    transition:
      background var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .node:hover,
  .node.hover {
    background: var(--surface-card);
    border-color: var(--hair);
    transform: translateX(2px);
  }
  .node:active {
    transform: translateX(2px) scale(0.99);
  }
  .node:focus-visible {
    outline: none;
    background: var(--surface-card);
    border-color: var(--synapse);
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    box-shadow: 0 0 6px currentColor;
  }
  .ts {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--content-faint);
    white-space: nowrap;
  }
  .actor {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--synapse);
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }
  .summary {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 14px;
    color: var(--content-soft);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .result {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
  }

  .detail {
    position: sticky;
    top: var(--space-4);
    align-self: start;
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    background: var(--surface-card);
    padding: var(--space-5);
    min-height: 160px;
    opacity: 0.4;
    transition: opacity var(--dur) var(--ease);
  }
  .detail.show {
    opacity: 1;
  }
  .d-eyebrow {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .d-ts {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--content);
    margin: var(--space-2) 0 var(--space-4);
  }
  .d-row {
    display: flex;
    justify-content: space-between;
    gap: var(--space-3);
    padding: 6px 0;
    border-top: 1px solid var(--hair);
    font-size: 13px;
  }
  .d-k {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .d-v {
    color: var(--content-soft);
    text-align: right;
  }
  .d-v.mono {
    font-family: var(--font-mono);
    font-size: 12px;
  }
  .d-v[data-level='error'] {
    color: var(--danger);
  }
  .d-v[data-level='warn'] {
    color: var(--warn);
  }
  .d-msg {
    margin-top: var(--space-4);
    font-family: var(--font-display);
    font-style: italic;
    font-size: 15px;
    line-height: 1.5;
    color: var(--content);
  }
  .d-empty {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 14px;
    color: var(--content-faint);
  }

  .pages {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-4);
    margin-top: var(--space-6);
  }
  .pg {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    padding: 7px 12px;
    border: 1px solid var(--hair-strong);
    border-radius: var(--r-sm);
    background: var(--surface-card);
    color: var(--content-soft);
    cursor: pointer;
    transition:
      border-color var(--dur) var(--ease),
      color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .pg:hover:not([disabled]) {
    border-color: var(--synapse);
    color: var(--synapse);
    background: color-mix(in oklab, var(--synapse) 6%, transparent);
    transform: translateY(-1px);
  }
  .pg:active:not([disabled]) {
    transform: scale(0.97);
  }
  .pg:focus-visible {
    outline: none;
    border-color: var(--synapse);
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .pg[disabled] {
    opacity: 0.42;
    cursor: not-allowed;
    filter: saturate(0.55);
  }
  .pg-info {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--content-faint);
  }

  @media (max-width: 820px) {
    .body {
      grid-template-columns: 1fr;
    }
    .detail {
      position: static;
    }
  }
</style>