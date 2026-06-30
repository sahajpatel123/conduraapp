<script lang="ts">
  // Audit — HMAC-chained event viewer.
  import { onMount } from 'svelte'
  import { audit } from '../stores/audit.svelte'
  import { ipc } from '../ipc/client'
  import { notifications } from '../stores/notifications.svelte'
  import type { AuditEvent } from '../ipc/types'
  import Button from '$components/v1/Button.svelte'
  import Input from '$components/v1/Input.svelte'
  import Chip from '$components/v1/Chip.svelte'
  import Stack from '$components/v1/Stack.svelte'
  import Inline from '$components/v1/Inline.svelte'
  import Surface from '$components/v1/Surface.svelte'
  import EmptyState from '$components/v1/EmptyState.svelte'
  import LoadingState from '$components/v1/LoadingState.svelte'
  import AgentActionLog from '$components/v1/AgentActionLog.svelte'
  import Pill from '$components/v1/Pill.svelte'

  type BlastFilter = 'all' | 'read' | 'write' | 'network' | 'destructive'
  type ActionType = 'read' | 'write' | 'network' | 'destructive'

  const blastOptions: { value: BlastFilter; label: string }[] = [
    { value: 'all', label: 'All' },
    { value: 'read', label: 'Read' },
    { value: 'write', label: 'Write' },
    { value: 'network', label: 'Network' },
    { value: 'destructive', label: 'Destructive' },
  ]

  let filterAction = $state('')
  let filterBlast = $state<BlastFilter>('all')
  let startDate = $state('')
  let endDate = $state('')
  let selected = $state<AuditEvent | null>(null)
  let detailOpen = $state(false)
  let verifyResult = $state<{
    valid: boolean
    rows_checked: number
    first_break_id?: number
    first_break_reason?: string
  } | null>(null)
  let verifyOpen = $state(false)
  let verifying = $state(false)

  const total = $derived(audit.events.length)

  const filteredEvents = $derived(
    filterBlast === 'all'
      ? audit.events
      : audit.events.filter((ev) => blastType(ev) === filterBlast)
  )

  const actionLogItems = $derived(
    filteredEvents.map((ev) => ({
      id: `${ev.id}`,
      time: fmtTs(ev.ts),
      type: blastType(ev),
      target: ev.app ? `${ev.app} · ${ev.action}` : ev.action,
      decision: `${ev.result} — ${ev.actor}`,
      verified: ev.result === 'allow',
      model: ev.actor,
    }))
  )

  function blastType(ev: AuditEvent): ActionType {
    if (ev.result === 'block') return 'destructive'
    if (ev.result === 'prompt') return 'network'
    return 'read'
  }

  function levelPill(level: string): 'success' | 'warning' | 'error' | 'info' | 'neutral' {
    if (level === 'error') return 'error'
    if (level === 'warn') return 'warning'
    return 'info'
  }

  function applyFilter(): void {
    audit.filterAction = filterAction
    audit.filterLevel = ''
    audit.offset = 0
    void audit.refresh()
  }

  function openDetailById(id: string): void {
    const ev = audit.events.find((e) => `${e.id}` === id)
    if (!ev) return
    selected = ev
    detailOpen = true
  }

  async function verifyIntegrity(): Promise<void> {
    verifying = true
    try {
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

  function fmtTs(ts: string): string {
    try {
      return new Date(ts).toISOString().replace('T', ' ').slice(0, 19)
    } catch {
      return ts
    }
  }

  onMount(() => {
    void audit.refresh()
  })
</script>

<Stack class="audit-page" gap="5" padding="7">
  <header class="page-header">
    <Inline gap="4" align="end" justify="between" class="title-row">
      <div class="title-block">
        <h2 class="page-title">Audit log</h2>
        <p class="lede">
          Every action the daemon takes, HMAC-chained and append-only. Forensics when
          something goes wrong.
        </p>
      </div>
      <Button variant="primary" size="md" loading={verifying} onclick={verifyIntegrity}>
        Verify integrity
      </Button>
    </Inline>

    <Stack gap="3" class="filters">
      <div class="search-wrap">
        <Input
          bind:value={filterAction}
          type="search"
          placeholder="Action contains…"
          ariaLabel="Filter by action"
          onkeydown={(e: KeyboardEvent) => {
            if (e.key === 'Enter') applyFilter()
          }}
        />
      </div>

      <Inline gap="2" align="center" wrap={true} class="blast-filters">
        {#each blastOptions as opt (opt.value)}
          <Chip
            selected={filterBlast === opt.value}
            onclick={() => {
              filterBlast = opt.value
              applyFilter()
            }}
          >
            {opt.label}
          </Chip>
        {/each}
      </Inline>

      <Inline gap="2" align="center" class="dates">
        <input
          class="date-input"
          type="date"
          bind:value={startDate}
          aria-label="Start date"
        />
        <span class="date-sep" aria-hidden="true">→</span>
        <input
          class="date-input"
          type="date"
          bind:value={endDate}
          aria-label="End date"
        />
      </Inline>
    </Stack>
  </header>

  {#if audit.loading}
    <LoadingState kind="cold" />
  {:else if total === 0}
    <Surface variant="raised" padding="4">
      <EmptyState
        primary="No matching events"
        voice="mono"
        secondary="Adjust the filters or wait for the daemon to record some actions."
      />
    </Surface>
  {:else}
    <AgentActionLog
      actions={actionLogItems}
      selectedId={selected ? `${selected.id}` : undefined}
      onrowclick={openDetailById}
    />

    <Inline gap="3" align="center" justify="between" class="pagination">
      <Button
        variant="tertiary"
        size="sm"
        disabled={audit.offset === 0}
        onclick={() => audit.prevPage()}
      >
        ← Previous
      </Button>
      <span class="offset mono">Offset: {audit.offset}</span>
      <Button
        variant="tertiary"
        size="sm"
        disabled={audit.events.length < audit.limit}
        onclick={() => audit.nextPage()}
      >
        Next →
      </Button>
    </Inline>
  {/if}
</Stack>

{#if detailOpen && selected}
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
  <div class="modal-scrim" onclick={() => (detailOpen = false)} role="presentation"></div>
  <aside
    class="modal-panel"
    role="dialog"
    aria-labelledby="audit-detail-title"
    aria-modal="true"
  >
    <Surface variant="overlay" padding="5" class="modal-surface">
      <Stack gap="4">
        <header>
          <h3 id="audit-detail-title" class="modal-title">Event #{selected.id}</h3>
          <p class="modal-sub">{selected.actor} · {selected.action}</p>
        </header>

        <div class="detail-grid">
          <div class="kv">
            <dt>ID</dt>
            <dd class="mono">{selected.id}</dd>
          </div>
          <div class="kv">
            <dt>Timestamp</dt>
            <dd class="mono">{selected.ts}</dd>
          </div>
          <div class="kv">
            <dt>Actor</dt>
            <dd>{selected.actor}</dd>
          </div>
          <div class="kv">
            <dt>Action</dt>
            <dd class="mono">{selected.action}</dd>
          </div>
          <div class="kv">
            <dt>App</dt>
            <dd>{selected.app}</dd>
          </div>
          <div class="kv">
            <dt>Level</dt>
            <dd><Pill variant={levelPill(selected.level)} size="xs" label={selected.level} /></dd>
          </div>
          <div class="kv">
            <dt>Result</dt>
            <dd class="mono">{selected.result}</dd>
          </div>
        </div>

        <div class="message-block">
          <h4 class="message-label">Message</h4>
          <pre class="message-body">{selected.message}</pre>
        </div>

        <footer class="modal-foot">
          <Button variant="secondary" size="md" onclick={() => (detailOpen = false)}>
            Close
          </Button>
        </footer>
      </Stack>
    </Surface>
  </aside>
{/if}

{#if verifyOpen && verifyResult}
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
  <div class="modal-scrim" onclick={() => (verifyOpen = false)} role="presentation"></div>
  <aside
    class="modal-panel modal-panel--sm"
    role="alertdialog"
    aria-labelledby="verify-title"
    aria-modal="true"
  >
    <Surface variant="overlay" padding="5" class="modal-surface">
      <Stack gap="4">
        <header>
          <h3 id="verify-title" class="modal-title">
            {verifyResult.valid ? 'Audit chain verified' : 'Audit chain broken'}
          </h3>
          <p class="modal-sub">{verifyResult.rows_checked} rows checked</p>
        </header>

        {#if verifyResult.valid}
          <p class="verify-ok">
            Every row matches its predecessor. The HMAC chain is intact.
          </p>
        {:else}
          <p class="verify-bad">
            Chain break at row <code class="mono">{verifyResult.first_break_id ?? '?'}</code>:
            {verifyResult.first_break_reason ?? 'unknown reason'}
          </p>
        {/if}

        <footer class="modal-foot">
          <Button variant="secondary" size="md" onclick={() => (verifyOpen = false)}>
            Close
          </Button>
        </footer>
      </Stack>
    </Surface>
  </aside>
{/if}

<style>
  .audit-page {
    overflow-y: auto;
    height: 100%;
    max-width: var(--content-max-width-wide, 72rem);
    margin: 0 auto;
  }

  .page-header {
    animation: audit-enter var(--duration-slow) var(--ease-standard) both;
  }

  @keyframes audit-enter {
    from {
      opacity: 0;
      transform: translateY(6px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .title-row {
    width: 100%;
  }

  .title-block {
    flex: 1;
    min-width: 0;
  }

  .page-title {
    font-family: var(--font-serif);
    font-size: var(--text-h2-size);
    font-weight: 500;
    letter-spacing: -0.02em;
    color: var(--content-primary);
    margin: 0 0 var(--space-2);
    line-height: var(--text-h2-leading);
  }

  .lede {
    font-size: var(--text-body-size);
    color: var(--content-secondary);
    line-height: 1.55;
    max-width: 40rem;
    margin: 0;
  }

  .search-wrap {
    max-width: 28rem;
  }

  .dates {
    padding: var(--space-2) var(--space-3);
    background-color: var(--surface-sunken);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    width: fit-content;
  }

  .date-input {
    background: transparent;
    border: none;
    color: var(--content-primary);
    font-family: var(--font-sans);
    font-size: var(--text-body-sm-size);
    padding: var(--space-1) var(--space-2);
    outline: none;
  }

  .date-input::-webkit-calendar-picker-indicator {
    filter: invert(0.55);
  }

  .date-sep {
    color: var(--content-tertiary);
    font-size: var(--text-caption-size);
  }

  .offset {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
  }

  .mono {
    font-family: var(--font-mono);
  }

  /* Modal (Channels sheet pattern) */
  .modal-scrim {
    position: fixed;
    inset: 0;
    background-color: var(--scrim-default, rgba(0, 0, 0, 0.4));
    z-index: 200;
  }

  .modal-panel {
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: min(36rem, calc(100vw - var(--space-6)));
    max-height: calc(100vh - var(--space-8));
    overflow-y: auto;
    z-index: 201;
  }

  .modal-panel--sm {
    width: min(28rem, calc(100vw - var(--space-6)));
  }

  :global(.modal-surface) {
    box-shadow: var(--shadow-overlay, 0 8px 32px rgba(0, 0, 0, 0.12));
  }

  .modal-title {
    font-family: var(--font-serif);
    font-size: var(--text-h3-size);
    font-weight: 500;
    color: var(--content-primary);
    margin: 0;
  }

  .modal-sub {
    font-size: var(--text-body-sm-size);
    color: var(--content-tertiary);
    margin: var(--space-1) 0 0;
  }

  .detail-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: var(--space-3) var(--space-4);
  }

  .kv {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .kv dt {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }

  .kv dd {
    font-size: var(--text-body-sm-size);
    color: var(--content-primary);
    margin: 0;
  }

  .message-label {
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-tertiary);
    text-transform: uppercase;
    letter-spacing: 0.06em;
    margin: 0;
  }

  .message-body {
    background-color: var(--surface-sunken);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    padding: var(--space-3);
    font-family: var(--font-mono);
    font-size: var(--text-caption-size);
    color: var(--content-primary);
    overflow: auto;
    max-height: 240px;
    margin: 0;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .modal-foot {
    display: flex;
    justify-content: flex-end;
  }

  .verify-ok {
    color: var(--status-success-fg);
    font-size: var(--text-body-sm-size);
    margin: 0;
  }

  .verify-bad {
    color: var(--status-error-fg);
    font-size: var(--text-body-sm-size);
    margin: 0;
  }
</style>
