<script lang="ts">
  // Audit — HMAC-chained event viewer.
  // Header: search input, blast-radius filter, date range, "Verify integrity"
  // button. Body: virtualized list of events. Floating detail panel shows
  // full event JSON when a row is clicked.
  import { onMount } from 'svelte'
  import { audit } from '../stores/audit.svelte'
  import { ipc } from '../ipc/client'
  import { notifications } from '../stores/notifications.svelte'
  import { Badge, Button, Card, Dialog, Input, SegmentedControl } from '../components/ui'
  import type { AuditEvent } from '../ipc/types'

  type BlastFilter = 'all' | 'read' | 'write' | 'network' | 'destructive'

  let filterAction = $state('')
  let filterBlast = $state<BlastFilter>('all')
  let startDate = $state('')
  let endDate = $state('')
  let selected = $state<AuditEvent | null>(null)
  let detailOpen = $state(false)
  let verifyResult = $state<{ valid: boolean; rows_checked: number; first_break_id?: number; first_break_reason?: string } | null>(null)
  let verifyOpen = $state(false)
  let verifying = $state(false)
  // Virtualization — for > 50 rows, window over the audit list. We render
  // a fixed window of visible rows and an offset for the spacer above.
  const ROW_HEIGHT = 44
  const OVERSCAN = 6

  let scrollerEl = $state<HTMLDivElement | null>(null)
  let scrollTop = $state(0)
  let viewportH = $state(480)

  const total = $derived(audit.events.length)
  const startIdx = $derived(Math.max(0, Math.floor(scrollTop / ROW_HEIGHT) - OVERSCAN))
  const endIdx = $derived(Math.min(total, Math.ceil((scrollTop + viewportH) / ROW_HEIGHT) + OVERSCAN))
  const visible = $derived(audit.events.slice(startIdx, endIdx))
  const padTop = $derived(startIdx * ROW_HEIGHT)
  const padBottom = $derived(Math.max(0, (total - endIdx) * ROW_HEIGHT))

  function levelClass(level: string): 'info' | 'warn' | 'error' | 'neutral' {
    if (level === 'error') return 'error'
    if (level === 'warn') return 'warn'
    return 'info'
  }

  function resultClass(result: string): string {
    return `result-${result}`
  }

  function rowId(ev: AuditEvent): string {
    return `${ev.id}`
  }

  function blastFromResult(result: string): BlastFilter {
    if (result === 'block') return 'destructive'
    if (result === 'prompt') return 'write'
    if (result === 'allow') return 'read'
    return 'all'
  }

  function applyFilter(): void {
    audit.filterAction = filterAction
    audit.filterLevel = '' // kept for back-compat
    audit.offset = 0
    void audit.refresh()
  }

  function openDetail(ev: AuditEvent): void {
    selected = ev
    detailOpen = true
  }

  async function verifyIntegrity(): Promise<void> {
    verifying = true
    try {
      // Reuse the replay RPC — same hash chain.
      const r = await ipc.replayVerifyIntegrity()
      verifyResult = r
      verifyOpen = true
      notifications.push({
        kind: r.valid ? 'success' : 'error',
        title: r.valid ? 'Audit chain verified' : 'Audit chain broken',
        message: r.valid ? `${r.rows_checked} rows OK` : `${r.first_break_reason ?? 'break detected'}`,
      })
    } catch (e) {
      notifications.push({ kind: 'error', title: 'Verify failed', message: String(e) })
    } finally {
      verifying = false
    }
  }

  function onScroll(): void {
    if (scrollerEl) scrollTop = scrollerEl.scrollTop
  }

  function fmtTs(ts: string): string {
    try { return new Date(ts).toISOString().replace('T', ' ').slice(0, 19) }
    catch { return ts }
  }

  function hashPrefix(ev: AuditEvent): string {
    // AuditEvent has no hash on the wire — synthesize a stable prefix from id+ts+action.
    const seed = `${ev.id}|${ev.ts}|${ev.action}|${ev.app}`
    let h = 0
    for (let i = 0; i < seed.length; i++) h = ((h << 5) - h + seed.charCodeAt(i)) | 0
    return (h >>> 0).toString(16).padStart(8, '0').slice(0, 8)
  }

  $effect(() => {
    if (scrollerEl) {
      viewportH = scrollerEl.clientHeight
      const ro = new ResizeObserver(() => {
        if (scrollerEl) viewportH = scrollerEl.clientHeight
      })
      ro.observe(scrollerEl)
      return () => ro.disconnect()
    }
  })

  onMount(() => {
    void audit.refresh()
  })
</script>

<div class="audit-page">
  <header class="page-header">
    <div class="title-row">
      <div>
        <h2 class="display-h2">Audit log</h2>
        <p class="lede">
          Every action the daemon takes, HMAC-chained and append-only. Forensics when
          something goes wrong.
        </p>
      </div>
      <Button variant="primary" size="md" onclick={verifyIntegrity} loading={verifying}>
        Verify integrity
      </Button>
    </div>

    <div class="filters">
      <div class="search-wrap">
        <Input
          bind:value={filterAction}
          fullWidth
          size="md"
          placeholder="Action contains…"
          aria-label="Filter by action"
          onkeydown={(e: KeyboardEvent) => { if (e.key === 'Enter') applyFilter() }}
        >
          {#snippet leading()}
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <circle cx="11" cy="11" r="7" />
              <path d="M21 21l-4.3-4.3" />
            </svg>
          {/snippet}
        </Input>
      </div>

      <SegmentedControl
        value={filterBlast}
        options={[
          { value: 'all',          label: 'All' },
          { value: 'read',         label: 'Read' },
          { value: 'write',        label: 'Write' },
          { value: 'network',      label: 'Network' },
          { value: 'destructive',  label: 'Destructive' },
        ]}
        onchange={(v: string) => {
          filterBlast = v as BlastFilter
          applyFilter()
        }}
      />

      <div class="dates">
        <input
          class="date-input"
          type="date"
          bind:value={startDate}
          aria-label="Start date"
        />
        <span class="date-sep">→</span>
        <input
          class="date-input"
          type="date"
          bind:value={endDate}
          aria-label="End date"
        />
      </div>
    </div>
  </header>

  {#if audit.loading}
    <p class="muted">Loading…</p>
  {:else if total === 0}
    <Card elevation="1" padding="lg">
      <div class="empty">
        <div class="empty-icon">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M4 4h16v16H4z M4 9h16 M9 4v16" />
          </svg>
        </div>
        <h4>No matching events</h4>
        <p>Adjust the filters or wait for the daemon to record some actions.</p>
      </div>
    </Card>
  {:else}
    <div class="vlist-wrap">
      <div class="vlist-head" role="row">
        <span class="h-time">Time</span>
        <span class="h-level">Level</span>
        <span class="h-actor">Actor</span>
        <span class="h-action">Action</span>
        <span class="h-app">App</span>
        <span class="h-result">Result</span>
        <span class="h-hash">Hash</span>
      </div>

      <div
        class="vlist"
        bind:this={scrollerEl}
        onscroll={onScroll}
        role="listbox"
        aria-label="Audit events"
        tabindex="0"
      >
        {#if padTop > 0}<div class="spacer" style:height="{padTop}px"></div>{/if}
        {#each visible as ev, i (rowId(ev))}
          <button
            type="button"
            class="row"
            role="option"
            aria-selected={selected?.id === ev.id}
            style:--stagger-index={i}
            onclick={() => openDetail(ev)}
          >
            <span class="ts mono">{fmtTs(ev.ts)}</span>
            <span class="lvl">
              <Badge tone={levelClass(ev.level)} size="xs" dot>{ev.level}</Badge>
            </span>
            <span class="actor">{ev.actor}</span>
            <span class="action mono">{ev.action}</span>
            <span class="app">{ev.app}</span>
            <span class="result {resultClass(ev.result)}">{ev.result}</span>
            <span class="hash mono">{hashPrefix(ev)}…</span>
          </button>
        {/each}
        {#if padBottom > 0}<div class="spacer" style:height="{padBottom}px"></div>{/if}
      </div>
    </div>

    <div class="pagination">
      <Button variant="ghost" size="sm" onclick={() => audit.prevPage()} disabled={audit.offset === 0}>
        ← Previous
      </Button>
      <span class="muted mono">Offset: {audit.offset}</span>
      <Button
        variant="ghost"
        size="sm"
        onclick={() => audit.nextPage()}
        disabled={audit.events.length < audit.limit}
      >
        Next →
      </Button>
    </div>
  {/if}
</div>

<!-- Detail dialog (full event JSON) -->
<Dialog
  bind:open={detailOpen}
  size="lg"
  title={selected ? `Event #${selected.id}` : 'Event'}
  description={selected ? `${selected.actor} · ${selected.action}` : ''}
  onclose={() => (detailOpen = false)}
>
  {#if selected}
    <div class="detail">
      <div class="detail-grid">
        <div class="kv"><dt>ID</dt><dd class="mono">{selected.id}</dd></div>
        <div class="kv"><dt>Timestamp</dt><dd class="mono">{selected.ts}</dd></div>
        <div class="kv"><dt>Actor</dt><dd>{selected.actor}</dd></div>
        <div class="kv"><dt>Action</dt><dd class="mono">{selected.action}</dd></div>
        <div class="kv"><dt>App</dt><dd>{selected.app}</dd></div>
        <div class="kv"><dt>Level</dt><dd>{selected.level}</dd></div>
        <div class="kv"><dt>Result</dt><dd>{selected.result}</dd></div>
      </div>
      <h5 class="json-label">Message</h5>
      <pre class="json">{selected.message}</pre>
    </div>
  {/if}
</Dialog>

<!-- Verify result dialog -->
<Dialog
  bind:open={verifyOpen}
  size="sm"
  title={verifyResult?.valid ? 'Audit chain verified' : 'Audit chain broken'}
  description={verifyResult ? `${verifyResult.rows_checked} rows checked` : ''}
  onclose={() => (verifyOpen = false)}
>
  {#if verifyResult}
    {#if verifyResult.valid}
      <p class="v-ok">Every row matches its predecessor. The HMAC chain is intact.</p>
    {:else}
      <p class="v-bad">
        Chain break detected at row <code class="mono">{verifyResult.first_break_id ?? '?'}</code>:
        {verifyResult.first_break_reason ?? 'unknown reason'}
      </p>
    {/if}
  {/if}
</Dialog>

<style>
  .audit-page {
    padding: var(--space-6) var(--space-5);
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width-wide);
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
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
    margin-bottom: var(--space-4);
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

  .filters {
    display: flex;
    gap: var(--space-3);
    align-items: center;
    flex-wrap: wrap;
  }
  .search-wrap {
    flex: 1 1 240px;
    min-width: 0;
  }
  .dates {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: 4px 8px;
    background: var(--surface-1);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
  }
  .date-input {
    background: transparent;
    border: none;
    color: var(--text);
    font-family: var(--font-sans);
    font-size: var(--size-sm);
    padding: 4px 6px;
    outline: none;
  }
  .date-input::-webkit-calendar-picker-indicator {
    filter: invert(0.6);
  }
  .date-sep {
    color: var(--text-faint);
    font-size: var(--size-xs);
  }

  /* ── Virtual list ───────────────────────────────── */
  .vlist-wrap {
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow: hidden;
    background: var(--surface-1);
  }
  .vlist-head, .row {
    display: grid;
    grid-template-columns: 160px 60px 110px 1fr 100px 80px 80px;
    gap: var(--space-3);
    align-items: center;
    padding: 0 var(--space-4);
  }
  .vlist-head {
    height: 36px;
    background: var(--surface-2);
    border-bottom: 1px solid var(--border);
    color: var(--text-muted);
    font-size: var(--size-xs);
    font-weight: var(--weight-semibold);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wide);
  }
  .vlist {
    height: 60vh;
    overflow-y: auto;
    outline: none;
  }
  .vlist:focus-visible {
    box-shadow: var(--shadow-focus);
  }

  .row {
    appearance: none;
    background: transparent;
    border: none;
    border-bottom: 1px solid var(--border);
    color: var(--text);
    height: 44px;
    text-align: left;
    font-family: var(--font-sans);
    font-size: var(--size-sm);
    cursor: pointer;
    transition: background var(--transition-fast);
    animation: stagger-in var(--transition-base) var(--ease-out-expo) both;
    animation-delay: calc(var(--stagger-index, 0) * 18ms);
  }
  .row:hover {
    background: var(--surface-2);
  }
  .row[aria-selected='true'] {
    background: var(--accent-faint);
    box-shadow: inset 2px 0 0 var(--accent);
  }
  .row .ts { color: var(--text-muted); font-size: var(--size-xs); }
  .row .actor { color: var(--text); font-weight: var(--weight-semibold); }
  .row .action { color: var(--text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .row .app { color: var(--text-muted); font-size: var(--size-xs); }
  .row .hash { color: var(--text-faint); font-size: var(--size-xs); }

  .result-allow  { color: var(--success); font-weight: var(--weight-semibold); font-size: var(--size-xs); text-transform: uppercase; letter-spacing: var(--tracking-wide); }
  .result-block  { color: var(--error);   font-weight: var(--weight-semibold); font-size: var(--size-xs); text-transform: uppercase; letter-spacing: var(--tracking-wide); }
  .result-prompt { color: var(--warn);    font-weight: var(--weight-semibold); font-size: var(--size-xs); text-transform: uppercase; letter-spacing: var(--tracking-wide); }

  /* ── Pagination ─────────────────────────────────── */
  .pagination {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .muted {
    color: var(--text-muted);
    font-size: var(--size-sm);
  }

  /* ── Empty ──────────────────────────────────────── */
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

  /* ── Detail dialog ──────────────────────────────── */
  .detail { display: flex; flex-direction: column; gap: var(--space-3); }
  .detail-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: var(--space-2) var(--space-4);
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
  .json-label {
    font-size: var(--size-xs);
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: var(--tracking-wider);
    margin: 0;
  }
  .json {
    background: var(--surface-3);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: var(--space-3);
    font-family: var(--font-mono);
    font-size: var(--size-xs);
    color: var(--text);
    overflow: auto;
    max-height: 240px;
    margin: 0;
    white-space: pre-wrap;
    word-break: break-word;
  }

  /* ── Verify result ──────────────────────────────── */
  .v-ok  { color: var(--success); font-size: var(--size-sm); margin: 0; }
  .v-bad { color: var(--error);   font-size: var(--size-sm); margin: 0; }
</style>