<script lang="ts">
  /**
   * Audit — append-only ledger instrument.
   * Signature: chain seal · verdict filters · event stage plate.
   */
  import { onMount } from 'svelte'
  import MeridianPage from './MeridianPage.svelte'
  import { audit } from '../../stores/audit.svelte'
  import { daemon } from '../../stores/daemon.svelte'
  import { t } from '../../i18n'
  import type { AuditEvent } from '../../ipc/types'
  import type { VerdictFilter, WhenPreset } from '../../stores/audit.svelte'

  let selectedId = $state<number | null>(null)
  /** Tick so “last live” age stays honest without thrashing. */
  let nowTick = $state(Date.now())

  const FILTERS = ['all', 'allow', 'block', 'prompt'] as const
  const WHEN: { id: WhenPreset; label: string }[] = [
    { id: '1h', label: '1h' },
    { id: '24h', label: '24h' },
    { id: '7d', label: '7d' },
    { id: '30d', label: '30d' },
    { id: 'all', label: 'all time' },
  ]

  const whenLabel = $derived(
    WHEN.find((w) => w.id === audit.filters.whenPreset)?.label ?? audit.filters.whenPreset
  )

  onMount(() => {
    // Ensure 24h window + server-backed list (not client-only page slice).
    void audit.refresh()
    void audit.verifyIntegrity()
    audit.startLive()
    const tick = setInterval(() => {
      nowTick = Date.now()
    }, 5000)
    return () => {
      clearInterval(tick)
      audit.stopLive()
    }
  })

  const liveNote = $derived.by(() => {
    if (!daemon.connected) return 'Daemon offline — ledger updates when Condura reconnects'
    if (!audit.liveSubscribed) return 'Live feed not subscribed — use Refresh for new verdicts'
    if (audit.lastLiveAt > 0) {
      const sec = Math.max(0, Math.floor((nowTick - audit.lastLiveAt) / 1000))
      if (sec < 5) return 'Live · new verdict just arrived'
      if (sec < 60) return `Live · last event ${sec}s ago`
      return `Live · last event ${Math.floor(sec / 60)}m ago`
    }
    return 'Live stream listening — new verdicts append here without refresh'
  })

  /** Server already filters by verdict; list is the chain for this filter. */
  const rows = $derived(audit.events)
  const filter = $derived(audit.filters.verdict)
  const selected = $derived(rows.find((e) => e.id === selectedId) ?? rows[0] ?? null)

  /** Chip badges prefer server facetCounts over the loaded page. */
  const counts = $derived.by(() => {
    const f = audit.facetCounts
    if (f) {
      return {
        all: f.total,
        allow: f.verdicts.allow ?? 0,
        block: f.verdicts.block ?? 0,
        prompt: f.verdicts.prompt ?? 0,
      }
    }
    return {
      all: audit.events.length,
      allow: audit.events.filter((e) => e.result === 'allow').length,
      block: audit.events.filter((e) => e.result === 'block').length,
      prompt: audit.events.filter((e) => e.result === 'prompt').length,
    }
  })

  $effect(() => {
    if (!rows.length) {
      selectedId = null
      return
    }
    if (selectedId === null || !rows.some((e) => e.id === selectedId)) {
      selectedId = rows[0]!.id
    }
  })

  function setFilter(v: (typeof FILTERS)[number]): void {
    audit.setVerdict(v as VerdictFilter)
  }

  function setWhen(w: WhenPreset): void {
    audit.setWhenPreset(w)
  }

  function formatWhen(v: unknown): string {
    if (!v) return '—'
    try {
      const d = new Date(String(v))
      return Number.isNaN(d.getTime()) ? String(v) : d.toLocaleString()
    } catch {
      return String(v)
    }
  }

  function isOffline(err: string | null): boolean {
    return !!err && /IPC client not started|not connected|Failed to fetch|daemon/i.test(err)
  }

  const showError = $derived(!!audit.error && !isOffline(audit.error))

  function pick(ev: AuditEvent): void {
    selectedId = ev.id
  }

  function goReplay(): void {
    window.location.hash = '#/replay'
  }

  function goAsk(): void {
    window.location.hash = '#/'
  }

  function onKey(e: KeyboardEvent): void {
    if (!rows.length) return
    const t = e.target as HTMLElement | null
    if (
      t &&
      (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.tagName === 'SELECT' || t.isContentEditable)
    ) {
      return
    }
    const i = rows.findIndex((row) => row.id === selectedId)
    if (e.key === 'ArrowDown' || e.key === 'j' || e.key === 'J') {
      e.preventDefault()
      const next = rows[Math.min(rows.length - 1, Math.max(0, i) + 1)]
      if (next) selectedId = next.id
    }
    if (e.key === 'ArrowUp' || e.key === 'k' || e.key === 'K') {
      e.preventDefault()
      const prev = rows[Math.max(0, (i < 0 ? 0 : i) - 1)]
      if (prev) selectedId = prev.id
    }
  }
</script>

<svelte:window onkeydown={onKey} />

<MeridianPage
  kicker="Ledger · append-only"
  title="Audit"
  lead="Every gated action is written here. The chain is the truth — verify the seal, then read each link."
>
  {#snippet actions()}
    <button type="button" class="md-btn md-btn-ghost" onclick={() => void audit.refresh()}>Refresh</button>
    <button
      type="button"
      class="md-btn md-btn-ghost"
      disabled={audit.exportInFlight}
      onclick={() => void audit.exportChain()}
    >
      {audit.exportInFlight ? 'Exporting…' : 'Export JSONL'}
    </button>
    <button
      type="button"
      class="md-btn md-btn-primary"
      disabled={audit.integrityLoading}
      onclick={() => void audit.verifyIntegrity()}
    >
      {audit.integrityLoading ? 'Verifying…' : 'Verify seal'}
    </button>
  {/snippet}

  <div class="desk md-stagger">
    <p
      class="contract"
      class:hot={daemon.connected && audit.liveSubscribed}
      class:off={!daemon.connected}
    >
      <span class="live-dot" aria-hidden="true"></span>
      {liveNote}. The past cannot be quietly rewritten.
    </p>

    {#if audit.integrity}
      <div class="seal" class:bad={!audit.integrity.ok} class:ok={audit.integrity.ok}>
        <span class="seal-mark" aria-hidden="true"></span>
        <div>
          <p class="cite">chain seal</p>
          <strong>Chain {audit.integrity.ok ? 'intact' : 'broken'}</strong>
          {#if audit.integrity.reason}
            <p class="reason">{audit.integrity.reason}</p>
          {:else if audit.integrity.ok}
            <p class="reason">
              {audit.integrity.rows_verified} row{audit.integrity.rows_verified === 1 ? '' : 's'} verified
              {#if audit.integrity.duration_ms}
                · {audit.integrity.duration_ms}ms
              {/if}
            </p>
          {/if}
          {#if !audit.integrity.ok && audit.integrity.broken_at_id}
            <p class="reason">First break at id {audit.integrity.broken_at_id}</p>
          {/if}
        </div>
        <button type="button" class="md-btn md-btn-ghost tiny" onclick={goReplay}>Open Replay</button>
      </div>
    {:else if audit.integrityError}
      <div class="seal bad">
        <span class="seal-mark" aria-hidden="true"></span>
        <div>
          <p class="cite">chain seal</p>
          <strong>Verify failed</strong>
          <p class="reason">{audit.integrityError}</p>
        </div>
        <button
          type="button"
          class="md-btn md-btn-ghost tiny"
          disabled={audit.integrityLoading}
          onclick={() => void audit.verifyIntegrity()}
        >
          Retry
        </button>
      </div>
    {/if}

    {#if audit.exportResult}
      <p class="export-note ok">
        Exported {audit.exportResult.count} row{audit.exportResult.count === 1 ? '' : 's'} ·
        {audit.exportResult.path}
      </p>
    {:else if audit.exportError}
      <p class="export-note bad">{audit.exportError}</p>
    {/if}

    <div class="filter-desk">
      <div class="filters when" role="group" aria-label={t('audit.filter_time')}>
        {#each WHEN as w (w.id)}
          <button
            type="button"
            class:on={audit.filters.whenPreset === w.id}
            data-w={w.id}
            disabled={audit.loading}
            onclick={() => setWhen(w.id)}
          >
            {w.label}
          </button>
        {/each}
      </div>
      <div class="filters" role="group" aria-label={t('audit.filter_verdict')}>
        {#each FILTERS as f (f)}
          <button
            type="button"
            class:on={filter === f}
            data-f={f}
            disabled={audit.loading}
            onclick={() => setFilter(f)}
          >
            {f}
            <em>{counts[f]}</em>
          </button>
        {/each}
      </div>
    </div>

    {#if audit.loading && rows.length === 0}
      <div class="md-empty">Loading ledger…</div>
    {:else if showError}
      <div class="md-empty">{audit.error}</div>
    {:else if rows.length === 0 && filter === 'all' && audit.filters.whenPreset === 'all'}
      <div class="empty-atlas">
        <p class="cite">{isOffline(audit.error) ? 'ledger offline' : 'quiet chain'}</p>
        <h2>{isOffline(audit.error) ? 'Ledger unread' : 'No events yet'}</h2>
        <p class="empty-lead">
          {#if isOffline(audit.error)}
            Connect the daemon to load the action chain.
          {:else}
            When Condura acts, every verdict lands here as a sealed link. Ask something to start the chain.
          {/if}
        </p>
        <div class="empty-actions">
          <button type="button" class="md-btn md-btn-primary" onclick={goAsk}>Go to Ask</button>
          <button type="button" class="md-btn md-btn-ghost" onclick={() => void audit.refresh()}>Refresh</button>
        </div>
      </div>
    {:else if rows.length === 0}
      <div class="md-empty empty">
        <p class="empty-title">
          {#if filter !== 'all'}
            No {filter} events
          {:else}
            Nothing in this window
          {/if}
        </p>
        <p class="empty-lead">
          Server filter · try another verdict or widen the time window
          (now: {whenLabel}).
        </p>
        <div class="empty-actions">
          {#if filter !== 'all'}
            <button type="button" class="md-btn md-btn-ghost" onclick={() => setFilter('all')}>
              All verdicts
            </button>
          {/if}
          {#if audit.filters.whenPreset !== 'all'}
            <button type="button" class="md-btn md-btn-ghost" onclick={() => setWhen('all')}>
              All time
            </button>
          {/if}
          {#if audit.filters.whenPreset !== '7d' && audit.filters.whenPreset !== 'all'}
            <button type="button" class="md-btn md-btn-ghost" onclick={() => setWhen('7d')}>
              Last 7 days
            </button>
          {/if}
        </div>
      </div>
    {:else}
      <div class="layout">
        <div class="chain">
          <p class="cite">
            ↑↓ / J K · {rows.length} links · {whenLabel}
            {#if filter !== 'all'} · {filter}{/if}
          </p>
          {#each rows as ev (ev.id)}
            <button
              type="button"
              class="link"
              class:on={selected?.id === ev.id}
              onclick={() => pick(ev)}
            >
              <span class="tick" data-v={ev.result} aria-hidden="true"></span>
              <span class="link-copy">
                <strong>{ev.action || ev.message || 'event'}</strong>
                <span class="when">{formatWhen(ev.ts)}</span>
              </span>
              <span class="chip" data-v={ev.result}>{ev.result}</span>
            </button>
          {/each}
          <div class="pager">
            <button
              type="button"
              class="md-btn md-btn-ghost"
              disabled={!audit.hasMore || audit.loading}
              onclick={() => void audit.loadMore()}
            >
              {audit.loading ? 'Loading…' : audit.hasMore ? 'Load more' : 'End of chain'}
            </button>
          </div>
        </div>

        {#if selected}
          <article class="stage-plate">
            <p class="cite">event · {selected.id}</p>
            <h2>{selected.action || 'Event'}</h2>
            <p class="body">{selected.message || 'No message on this link.'}</p>
            <dl class="facts">
              <div>
                <dt>When</dt>
                <dd>{formatWhen(selected.ts)}</dd>
              </div>
              <div>
                <dt>Verdict</dt>
                <dd data-v={selected.result}>{selected.result}</dd>
              </div>
              <div>
                <dt>Actor</dt>
                <dd>{selected.actor || '—'}</dd>
              </div>
              <div>
                <dt>App</dt>
                <dd>{selected.app || selected.target_app || '—'}</dd>
              </div>
              {#if selected.command}
                <div class="wide">
                  <dt>Command</dt>
                  <dd class="mono">{selected.command}</dd>
                </div>
              {/if}
              {#if selected.path}
                <div class="wide">
                  <dt>Path</dt>
                  <dd class="mono">{selected.path}</dd>
                </div>
              {/if}
              {#if selected.this_hash}
                <div class="wide">
                  <dt>Hash</dt>
                  <dd class="mono">{selected.this_hash}</dd>
                </div>
              {/if}
              {#if selected.prev_hash}
                <div class="wide">
                  <dt>Prev</dt>
                  <dd class="mono">{selected.prev_hash}</dd>
                </div>
              {/if}
            </dl>
          </article>
        {/if}
      </div>
    {/if}
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
    border-color: color-mix(in oklab, var(--md-live) 22%, var(--md-line));
    background: color-mix(in oklab, var(--md-live) 6%, var(--md-surface));
  }
  .contract.off {
    border-color: color-mix(in oklab, var(--md-halt) 22%, var(--md-line));
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
    background: var(--md-live);
    box-shadow: none;
  }
  .cite {
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin: 0 0 6px;
  }
  .seal {
    display: flex;
    gap: 14px;
    align-items: center;
    padding: 14px 16px;
    border-radius: 12px;
    border: 1px solid var(--md-line);
    background: var(--md-surface);
  }
  .seal.ok {
    border-color: color-mix(in oklab, var(--md-live) 35%, transparent);
  }
  .seal.bad {
    border-color: color-mix(in oklab, var(--md-halt) 40%, transparent);
    color: var(--md-halt);
  }
  .seal-mark {
    width: 14px;
    height: 14px;
    border-radius: 50%;
    flex: none;
    background: var(--md-ink-faint);
    box-shadow: none;
  }
  .seal.ok .seal-mark {
    background: var(--md-live);
    box-shadow: none;
  }
  .seal.bad .seal-mark {
    background: var(--md-halt);
    box-shadow: none;
  }
  .seal strong {
    font-family: var(--md-font-display);
    font-size: 16px;
    letter-spacing: -0.03em;
  }
  .reason {
    margin: 4px 0 0;
    font-size: 13px;
    color: var(--md-ink-mute);
  }
  .seal :global(.md-btn.tiny) {
    margin-left: auto;
    padding: 7px 12px;
    font-size: 12px;
    flex: none;
  }
  .export-note {
    margin: 0;
    font-family: var(--md-font-mono);
    font-size: 11px;
    letter-spacing: 0.03em;
    line-height: 1.45;
    word-break: break-all;
  }
  .export-note.ok {
    color: var(--md-live);
  }
  .export-note.bad {
    color: var(--md-halt);
  }
  .filter-desk {
    display: grid;
    gap: 10px;
  }
  .filters {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .filters button {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 6px 10px;
    border-radius: 7px;
    border: 1px solid var(--md-line);
    background: var(--md-stage);
    font-family: var(--md-font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--md-ink-mute);
    cursor: pointer;
  }
  .filters button:disabled {
    opacity: 0.5;
    cursor: wait;
  }
  .filters button em {
    font-style: normal;
    opacity: 0.7;
  }
  .filters.when button.on {
    background: color-mix(in oklab, var(--md-cobalt) 14%, var(--md-surface));
    border-color: color-mix(in oklab, var(--md-cobalt) 40%, transparent);
    color: var(--md-cobalt);
  }
  .filters button.on[data-f='all'] {
    background: var(--md-cobalt);
    border-color: var(--md-cobalt);
    color: #fff;
  }
  .filters button.on[data-f='allow'] {
    color: var(--md-live);
    border-color: color-mix(in oklab, var(--md-live) 40%, transparent);
    background: color-mix(in oklab, var(--md-live) 12%, transparent);
  }
  .filters button.on[data-f='block'] {
    color: var(--md-halt);
    border-color: color-mix(in oklab, var(--md-halt) 40%, transparent);
    background: color-mix(in oklab, var(--md-halt) 12%, transparent);
  }
  .filters button.on[data-f='prompt'] {
    color: var(--md-cobalt);
    border-color: color-mix(in oklab, var(--md-cobalt) 40%, transparent);
    background: color-mix(in oklab, var(--md-cobalt) 12%, transparent);
  }
  .filters button:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .empty-atlas {
    border-radius: 12px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    padding: 22px 20px;
    box-shadow: none;
    max-width: 520px;
  }
  .empty-atlas h2 {
    font-family: var(--md-font-display);
    font-size: 22px;
    font-weight: 650;
    letter-spacing: -0.03em;
    margin: 0 0 8px;
  }
  .empty-lead {
    margin: 0 0 16px;
    font-size: 14px;
    line-height: 1.5;
    color: var(--md-ink-mute);
    max-width: 42ch;
  }
  .empty-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    justify-content: center;
  }
  .empty-title {
    margin: 0;
    font-family: var(--md-font-display);
    font-size: 16px;
  }
  .layout {
    display: grid;
    grid-template-columns: minmax(240px, 1fr) minmax(260px, 0.9fr);
    gap: 14px;
    align-items: start;
  }
  .chain {
    display: grid;
    gap: 0;
  }
  .link {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 12px;
    align-items: center;
    width: 100%;
    text-align: left;
    padding: 12px 12px 12px 8px;
    border: 0;
    border-bottom: 1px solid var(--md-line);
    background: transparent;
    cursor: pointer;
    color: inherit;
  }
  .link:hover {
    background: color-mix(in oklab, var(--md-stage) 70%, transparent);
  }
  .link.on {
    background: var(--md-surface);
    border-radius: 10px;
    border-bottom-color: transparent;
    box-shadow: none;
    border: 1px solid var(--md-line);
  }
  .link:focus-visible {
    outline: none;
    box-shadow: var(--md-focus);
  }
  .tick {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--md-ink-faint);
    box-shadow: none;
  }
  .tick[data-v='allow'] {
    background: var(--md-live);
  }
  .tick[data-v='block'] {
    background: var(--md-halt);
  }
  .tick[data-v='prompt'] {
    background: var(--md-cobalt);
  }
  .link-copy {
    min-width: 0;
    display: grid;
    gap: 3px;
  }
  .link-copy strong {
    font-size: 13px;
    font-weight: 700;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .when {
    font-family: var(--md-font-mono);
    font-size: 10px;
    color: var(--md-ink-faint);
  }
  .chip {
    font-family: var(--md-font-mono);
    font-size: 9px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    padding: 3px 7px;
    border-radius: 5px;
    border: 1px solid var(--md-line);
    color: var(--md-ink-mute);
  }
  .chip[data-v='allow'] {
    color: var(--md-live);
    border-color: color-mix(in oklab, var(--md-live) 30%, transparent);
  }
  .chip[data-v='block'] {
    color: var(--md-halt);
    border-color: color-mix(in oklab, var(--md-halt) 30%, transparent);
  }
  .chip[data-v='prompt'] {
    color: var(--md-cobalt);
    border-color: color-mix(in oklab, var(--md-cobalt) 30%, transparent);
  }
  .stage-plate {
    border-radius: 12px;
    border: 1px solid var(--md-line-strong);
    background: var(--md-surface);
    padding: 18px;
    box-shadow: none;
    position: sticky;
    top: 8px;
  }
  .stage-plate h2 {
    font-family: var(--md-font-display);
    font-size: 20px;
    font-weight: 650;
    letter-spacing: -0.03em;
    margin: 0 0 8px;
  }
  .body {
    margin: 0;
    font-size: 14px;
    line-height: 1.5;
    color: var(--md-ink-mute);
  }
  .facts {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
    margin: 18px 0 0;
    padding-top: 16px;
    border-top: 1px solid var(--md-line);
  }
  .facts .wide {
    grid-column: 1 / -1;
  }
  .facts dt {
    font-family: var(--md-font-mono);
    font-size: 9px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--md-ink-faint);
    margin-bottom: 4px;
  }
  .facts dd {
    margin: 0;
    font-size: 13px;
    font-weight: 600;
    color: var(--md-ink-soft);
    word-break: break-word;
  }
  .facts dd[data-v='allow'] {
    color: var(--md-live);
  }
  .facts dd[data-v='block'] {
    color: var(--md-halt);
  }
  .facts dd[data-v='prompt'] {
    color: var(--md-cobalt);
  }
  .facts dd.mono {
    font-family: var(--md-font-mono);
    font-size: 11px;
    font-weight: 500;
  }
  .pager {
    margin-top: 12px;
  }
  @media (max-width: 800px) {
    .layout {
      grid-template-columns: 1fr;
    }
    .stage-plate {
      position: static;
    }
    .seal {
      flex-wrap: wrap;
    }
    .seal :global(.md-btn.tiny) {
      margin-left: 0;
    }
  }
</style>
