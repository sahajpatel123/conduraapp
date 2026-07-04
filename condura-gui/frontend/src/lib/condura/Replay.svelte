<script lang="ts">
  // Condura Replay — the last 24h as a synapse thread you scrub.
  // The thread at the bottom IS the scrubber: green from the left edge to
  // the playhead (time already lived), a hairline after (time ahead). A
  // pollen mote rides the playhead. Above it, the selected frame's
  // screenshot and a decision "Receipt". Arrow keys step frame by frame.
  // An integrity badge verifies the HMAC audit chain on mount.
  import { onMount } from 'svelte';
  import { replay } from '../stores/replay.svelte';
  import type { ReplayFrame } from '../ipc/types';
  import Thread from './Thread.svelte';
  import Pulse from './Pulse.svelte';
  import Glyph from './Glyph.svelte';

  // ── local UI state ──
  let exportPath = $state('');
  let scrubbing = $state(false);
  let scrubberEl = $state<HTMLDivElement | null>(null);

  // ── store-derived views ──
  let frames = $derived(replay.frames);
  let count = $derived(frames.length);
  let idx = $derived(replay.selectedIndex);
  let selected = $derived(replay.selected);
  let playheadPct = $derived(count > 1 ? (idx / (count - 1)) * 100 : 0);
  let integrity = $derived(replay.integrity);
  let loading = $derived(replay.loading);
  let exporting = $derived(replay.exporting);
  let lastError = $derived(replay.lastError);

  // ── helpers ──
  function fmtTs(ts: string): string {
    if (!ts) return '—';
    try {
      return new Date(ts).toLocaleString(undefined, {
        month: 'short',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
    } catch {
      return ts;
    }
  }

  function thumbMime(f: ReplayFrame): string {
    return f.before_screenshot
      ? f.before_screenshot_mime || 'image/png'
      : f.after_screenshot_mime || 'image/png';
  }

  // The serif-italic "I decided to…" line. The articulated decision lives in
  // `message`; fall back to a constructed clause from `action` when absent.
  function decisionLine(f: ReplayFrame): string {
    const m = (f.message || '').trim();
    return m ? m : `I decided to ${f.action || 'act'}.`;
  }

  function outcomeKind(o: string): 'good' | 'bad' | '' {
    if (o === 'allowed') return 'good';
    if (o === 'denied' || o === 'errored') return 'bad';
    return '';
  }

  // ── scrubbing (pointer drag + click-to-jump) ──
  function jumpToClientX(clientX: number): void {
    if (!scrubberEl || count < 1) return;
    const rect = scrubberEl.getBoundingClientRect();
    const pct = Math.min(1, Math.max(0, (clientX - rect.left) / rect.width));
    const i = count > 1 ? Math.round(pct * (count - 1)) : 0;
    replay.selectIndex(i);
  }

  function onPointerDown(e: PointerEvent): void {
    scrubbing = true;
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    jumpToClientX(e.clientX);
  }
  function onPointerMove(e: PointerEvent): void {
    if (!scrubbing) return;
    jumpToClientX(e.clientX);
  }
  function onPointerUp(e: PointerEvent): void {
    scrubbing = false;
    try {
      (e.currentTarget as HTMLElement).releasePointerCapture(e.pointerId);
    } catch {
      /* ignore */
    }
  }

  // ── keyboard: ←/→ step frame-by-frame, Home/End jump to ends ──
  function onKey(e: KeyboardEvent): void {
    if (e.key === 'ArrowLeft') {
      e.preventDefault();
      replay.selectIndex(idx - 1);
    } else if (e.key === 'ArrowRight') {
      e.preventDefault();
      replay.selectIndex(idx + 1);
    } else if (e.key === 'Home') {
      e.preventDefault();
      replay.selectIndex(0);
    } else if (e.key === 'End') {
      e.preventDefault();
      replay.selectIndex(count - 1);
    }
  }

  // ── export .mp4 (pollen-outline) ──
  async function onExport(): Promise<void> {
    exportPath = '';
    try {
      exportPath = await replay.exportMP4();
    } catch {
      /* replay.lastError surfaces in the status row */
    }
  }

  onMount(() => {
    void replay.refresh();
    void replay.verifyIntegrity();
  });
</script>

<section class="replay" aria-label="Action replay">
  <header class="r-head">
    <div class="r-head-text">
      <div class="eyebrow">Action Replay</div>
      <h1 class="headline">The last 24 hours, scrubbable.</h1>
      <p class="lead">
        Every action the agent took, in order. Drag the thread to inspect what it saw, and what it
        decided.
      </p>
    </div>

    <div class="r-head-actions">
      <button
        class="badge"
        class:ok={integrity?.valid === true}
        class:bad={integrity != null && integrity.valid === false}
        onclick={() => replay.verifyIntegrity()}
        title="Verify the HMAC audit chain"
      >
        {#if integrity?.valid}
          <Glyph name="check" size={13} stroke={2} />
          <span>Chain intact</span>
          <span class="badge-meta">· {integrity.rows_checked} rows</span>
        {:else if integrity}
          <Glyph name="shield" size={13} stroke={2} />
          <span>Chain broken</span>
        {:else}
          <Pulse phase="thinking" size={7} />
          <span>Verifying…</span>
        {/if}
      </button>

      <button class="export-btn" onclick={onExport} disabled={exporting || count === 0}>
        {#if exporting}
          <Pulse phase="acting" size={7} />
          <span>Exporting…</span>
        {:else}
          <Glyph name="replay" size={14} />
          <span>Export .mp4</span>
        {/if}
      </button>
    </div>
  </header>

  <div class="rule"><Thread orientation="h" /></div>

  {#if lastError}
    <div class="err-state" role="alert" aria-live="polite">
      <div class="err-row">
        <Pulse phase="error" size={8} />
        <span class="err-head">We couldn't read the timeline.</span>
      </div>
      <p class="err-sub">{lastError} Replay frames live on disk; the daemon may be offline or the file may be unreadable.</p>
      <div class="err-actions">
        <button class="retry" onclick={() => void replay.refresh()}>Try again</button>
      </div>
      <div class="err-hair"></div>
    </div>
  {/if}
  {#if integrity && !integrity.valid && integrity.first_break_reason}
    <p class="integrity-detail">Chain broken at row {integrity.first_break_id ?? '—'}: {integrity.first_break_reason}</p>
  {/if}
  {#if exportPath}
    <p class="export-result">Exported to {exportPath}</p>
  {/if}

  {#if loading && count === 0}
    <div class="state-empty">
      <Pulse phase="thinking" size={10} />
      <span class="state-text">LOADING FRAMES…</span>
    </div>
  {:else if count === 0 && !lastError}
    <div class="state-empty">
      <span class="state-head">Nothing to replay yet.</span>
      <span class="state-sub">Once the agent acts, every decision lands here — screenshot, decision, outcome. The last 24 hours, scrubbable.</span>
    </div>
  {:else if selected}
    {#key selected.id}
      <div class="frame-view">
        <div class="shots">
          {#if selected.before_screenshot}
            <figure class="shot-img">
              <img
                src={`data:${selected.before_screenshot_mime || 'image/png'};base64,${selected.before_screenshot}`}
                alt="Before action"
                loading="lazy"
              />
              <figcaption>Before</figcaption>
            </figure>
          {/if}
          {#if selected.after_screenshot}
            <figure class="shot-img">
              <img
                src={`data:${selected.after_screenshot_mime || 'image/png'};base64,${selected.after_screenshot}`}
                alt="After action"
                loading="lazy"
              />
              <figcaption>After</figcaption>
            </figure>
          {/if}
          {#if !selected.before_screenshot && !selected.after_screenshot}
            <div class="shot-empty">No screenshot for this frame</div>
          {/if}
        </div>

        <aside class="receipt">
          <div class="receipt-ts mono">{fmtTs(selected.timestamp)}</div>
          <p class="receipt-line">{decisionLine(selected)}</p>

          <dl class="receipt-rows">
            <div><dt>Action</dt><dd>{selected.action || '—'}</dd></div>
            <div><dt>App</dt><dd>{selected.app || '—'}</dd></div>
            <div><dt>Actor</dt><dd>{selected.actor || '—'}</dd></div>
            <div><dt>Result</dt><dd>{selected.result || '—'}</dd></div>
            <div>
              <dt>Outcome</dt>
              <dd class:good={outcomeKind(selected.outcome) === 'good'} class:bad={outcomeKind(selected.outcome) === 'bad'}>
                {selected.outcome || '—'}
              </dd>
            </div>
            <div><dt>Level</dt><dd>{selected.level || '—'}</dd></div>
          </dl>

          {#if selected.outcome_reason}
            <p class="receipt-reason">{selected.outcome_reason}</p>
          {/if}
        </aside>
      </div>
    {/key}

    <div class="scrubber-wrap">
      <div class="counter mono">
        <span>{idx + 1} / {count}</span>
        <span class="ts">{fmtTs(selected.timestamp)}</span>
        <span class="hint">← → to step · drag to scrub</span>
      </div>

      <div
        class="scrubber"
        bind:this={scrubberEl}
        class:scrubbing={scrubbing}
        role="slider"
        tabindex="0"
        aria-label="Replay timeline"
        aria-valuemin={0}
        aria-valuemax={count - 1}
        aria-valuenow={idx}
        aria-valuetext={fmtTs(selected.timestamp)}
        onpointerdown={onPointerDown}
        onpointermove={onPointerMove}
        onpointerup={onPointerUp}
        onpointercancel={onPointerUp}
        onkeydown={onKey}
      >
        <div class="track"></div>
        <div class="fill" style:width="{playheadPct}%">
          <Thread orientation="h" draw={true} glow={false} />
        </div>
        <div class="mote" style:left="{playheadPct}%">
          <span class="mote-halo"></span>
          <Pulse phase="acting" size={10} />
        </div>
      </div>
    </div>
  {/if}
</section>

<style>
  .replay {
    height: 100%;
    display: flex;
    flex-direction: column;
    max-width: 72rem;
    margin: 0 auto;
    padding: var(--space-7) var(--space-8) var(--space-6);
    animation: blur-in var(--dur-slow) var(--ease) both;
  }

  /* ── header ── */
  .r-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-6);
  }
  .r-head-text .headline {
    font-size: clamp(26px, 2.6vw, 34px);
    margin: var(--space-2) 0;
  }
  .r-head-text .lead {
    margin-top: var(--space-1);
    max-width: 46ch;
  }
  .r-head-actions {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    flex-shrink: 0;
  }

  .rule {
    width: 140px;
    margin: var(--space-3) 0 var(--space-6);
  }

  /* ── status lines ── */
  .integrity-detail,
  .export-result {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    margin-bottom: var(--space-3);
    word-break: break-word;
  }
  .integrity-detail {
    color: var(--danger);
  }
  .export-result {
    color: var(--ok);
  }

  /* ── empty / loading ── */
  .state-empty {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-3);
    text-align: center;
    color: var(--content-faint);
    padding: var(--space-7) 0;
  }
  .state-empty .state-text {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--content-mute);
  }
  .state-head {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 24px;
    line-height: 1.1;
    color: var(--content);
    letter-spacing: -0.015em;
  }
  .state-sub {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 15px;
    line-height: 1.55;
    color: var(--content-faint);
    max-width: 48ch;
  }

  /* ── error state ── */
  .err-state {
    max-width: 520px;
    padding: var(--space-4) 0 var(--space-3);
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

  /* ── frame viewer ── */
  .frame-view {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: minmax(0, 1fr) 340px;
    gap: var(--space-6);
    margin-bottom: var(--space-5);
    animation: blur-in var(--dur) var(--ease) both;
  }

  .shots {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: var(--space-4);
    align-content: start;
    margin: 0;
  }
  .shot-img {
    position: relative;
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    overflow: hidden;
    background: var(--paper-2);
    box-shadow: var(--shadow-paper);
  }
  .shot-img img {
    width: 100%;
    height: auto;
    max-height: 56vh;
    object-fit: contain;
    display: block;
  }
  .shot-img figcaption {
    position: absolute;
    top: var(--space-2);
    left: var(--space-2);
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--paper);
    background: color-mix(in oklab, var(--ink) 62%, transparent);
    padding: 3px 7px;
    border-radius: var(--r-pill);
  }
  :root[data-mode='dark'] .shot-img figcaption {
    color: var(--content-soft);
    background: color-mix(in oklab, var(--ink) 60%, transparent);
  }
  .shot-empty {
    grid-column: 1 / -1;
    border: 1px dashed var(--hair-strong);
    border-radius: var(--r-md);
    display: grid;
    place-items: center;
    min-height: 220px;
    color: var(--content-faint);
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
  }

  /* ── receipt ── */
  .receipt {
    background: var(--surface-card);
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    padding: var(--space-5);
    box-shadow: var(--shadow-paper);
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    align-self: start;
  }
  .receipt-ts {
    color: var(--content-faint);
  }
  .receipt-line {
    font-family: var(--font-display);
    font-style: italic;
    font-size: 22px;
    line-height: 1.22;
    letter-spacing: -0.02em;
    color: var(--content);
  }
  .receipt-rows {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: 0;
  }
  .receipt-rows > div {
    display: grid;
    grid-template-columns: 80px 1fr;
    gap: var(--space-3);
    align-items: baseline;
  }
  .receipt-rows dt {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .receipt-rows dd {
    font-size: 14px;
    color: var(--content-soft);
    margin: 0;
    word-break: break-word;
  }
  .receipt-rows dd.good {
    color: var(--ok);
  }
  .receipt-rows dd.bad {
    color: var(--danger);
  }
  .receipt-reason {
    font-size: 13px;
    line-height: 1.5;
    color: var(--content-mute);
    font-style: italic;
    border-top: 1px solid var(--hair);
    padding-top: var(--space-3);
    margin: 0;
  }

  /* ── scrubber ── */
  .scrubber-wrap {
    flex: none;
    padding-top: var(--space-4);
    border-top: 1px solid var(--hair);
  }
  .counter {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    font-size: 11px;
    color: var(--content-faint);
    margin-bottom: var(--space-4);
  }
  .counter .ts {
    color: var(--content-mute);
  }
  .counter .hint {
    margin-left: auto;
    opacity: 0.7;
  }

  .scrubber {
    position: relative;
    height: 30px;
    cursor: pointer;
    display: flex;
    align-items: center;
    touch-action: none;
  }
  .scrubber:focus-visible {
    outline: none;
  }
  .scrubber:focus-visible .track {
    background: var(--synapse);
    opacity: 1;
  }
  .scrubber:focus-visible .mote {
    transform: translate(-50%, -50%) scale(1.4);
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }

  .track {
    position: absolute;
    left: 0;
    right: 0;
    height: 1px;
    background: var(--hair-strong);
    transition: background var(--dur) var(--ease);
  }

  .fill {
    position: absolute;
    left: 0;
    top: 50%;
    height: 2px;
    transform: translateY(-50%);
    overflow: hidden;
    transition: width var(--dur) var(--ease);
  }
  .scrubbing .fill {
    transition: none;
  }

  .mote {
    position: absolute;
    top: 50%;
    left: 0;
    transform: translate(-50%, -50%);
    display: grid;
    place-items: center;
    pointer-events: none;
    transition: left var(--dur) var(--ease);
  }
  .scrubbing .mote {
    transition: none;
  }
  .mote-halo {
    position: absolute;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: radial-gradient(circle, var(--pollen-light) 0%, transparent 68%);
    opacity: 0.55;
  }
  :root[data-mode='dark'] .mote-halo {
    opacity: 0.4;
  }

  /* ── integrity badge ── */
  .badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    padding: 7px 12px;
    border-radius: var(--r-pill);
    border: 1px solid var(--hair-strong);
    color: var(--content-mute);
    background: transparent;
    transition:
      color var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      background var(--dur) var(--ease),
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .badge:hover {
    color: var(--content);
    border-color: var(--hair-strong);
    background: var(--surface-card);
    transform: translateY(-1px);
  }
  .badge:active {
    transform: scale(0.97);
  }
  .badge:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .badge.ok {
    color: var(--ok);
    border-color: var(--ok);
    background: color-mix(in oklab, var(--ok) 8%, transparent);
  }
  .badge.bad {
    color: var(--danger);
    border-color: var(--danger);
    background: color-mix(in oklab, var(--danger) 8%, transparent);
  }
  :root[data-mode='dark'] .badge.ok {
    background: color-mix(in oklab, var(--ok) 14%, transparent);
  }
  :root[data-mode='dark'] .badge.bad {
    background: color-mix(in oklab, var(--danger) 14%, transparent);
  }
  .badge-meta {
    opacity: 0.7;
  }

  /* ── pollen-outline export button ── */
  .export-btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-family: var(--font-sans);
    font-size: 13px;
    font-weight: 500;
    letter-spacing: -0.005em;
    padding: 9px 16px;
    border-radius: var(--r-pill);
    border: 1px solid var(--pollen);
    color: var(--pollen);
    background: transparent;
    transition:
      background var(--dur) var(--ease),
      color var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .export-btn:hover:not([disabled]) {
    background: var(--pollen);
    color: var(--paper);
    box-shadow: var(--pollen-halo);
    transform: translateY(-1px);
  }
  .export-btn:active:not([disabled]) {
    transform: scale(0.97);
  }
  .export-btn:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  :root[data-mode='dark'] .export-btn:hover:not([disabled]) {
    color: var(--ink);
  }
  .export-btn[disabled] {
    opacity: 0.42;
    pointer-events: none;
    cursor: not-allowed;
    filter: saturate(0.55);
  }

  /* ── responsive ── */
  @media (max-width: 880px) {
    .replay {
      padding: var(--space-6) var(--space-5) var(--space-5);
    }
    .r-head {
      flex-direction: column;
    }
    .frame-view {
      grid-template-columns: 1fr;
    }
    .counter .hint {
      display: none;
    }
  }
</style>